// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package host

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	func_e_api "github.com/tetratelabs/func-e/api"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/infrastructure/common"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils"
	"github.com/envoyproxy/gateway/internal/xds/bootstrap"
)

// proxyContext corresponds to the context of the Envoy process.
type proxyContext struct {
	// cancel is the function to cancel the context passed to the Envoy process.
	cancel context.CancelFunc
	// exit will receive an item when the Envoy process completely stopped, via funcE.ExitChannel.
	exit chan struct{}
}

// Close implements the Manager interface.
func (i *Infra) Close() error {
	var wg sync.WaitGroup

	// Stop any Envoy subprocesses in parallel
	i.proxyContextMap.Range(func(key, value any) bool {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			i.stopEnvoy(name)
		}(key.(string))
		return true
	})

	wg.Wait()
	return nil
}

// CreateOrUpdateProxyInfra creates the managed host process, if it doesn't exist.
func (i *Infra) CreateOrUpdateProxyInfra(ctx context.Context, infra *ir.Infra) error {
	if infra == nil {
		return errors.New("infra ir is nil")
	}

	if infra.Proxy == nil {
		return errors.New("infra proxy ir is nil")
	}

	proxyInfra := infra.GetProxyInfra()
	proxyName := utils.GetHashedName(proxyInfra.Name, 64)
	// Return directly if the proxy is running.
	if _, loaded := i.proxyContextMap.Load(proxyName); loaded {
		return nil
	}

	proxyConfig := proxyInfra.GetProxyConfig()
	// Build proxy metrics with Prometheus disabled for host mode,
	// but preserve any user-configured sinks (e.g., OpenTelemetry).
	proxyMetrics := &egv1a1.ProxyMetrics{
		Prometheus: &egv1a1.ProxyPrometheusProvider{
			Disable: true,
		},
	}
	if proxyConfig.Spec.Telemetry != nil && proxyConfig.Spec.Telemetry.Metrics != nil {
		proxyMetrics.Sinks = proxyConfig.Spec.Telemetry.Metrics.Sinks
		proxyMetrics.Matches = proxyConfig.Spec.Telemetry.Metrics.Matches
	}

	resolvedMetricSinks := convertResolvedMetricSinks(proxyInfra.ResolvedMetricSinks)

	bootstrapConfigOptions := &bootstrap.RenderBootstrapConfigOptions{
		ProxyMetrics:        proxyMetrics,
		ResolvedMetricSinks: resolvedMetricSinks,
		SdsConfig: bootstrap.SdsConfigPath{
			Certificate: filepath.Join(i.sdsConfigPath, common.SdsCertFilename),
			TrustedCA:   filepath.Join(i.sdsConfigPath, common.SdsCAFilename),
		},
		XdsServerHost:   ptr.To("0.0.0.0"),
		AdminServerPort: ptr.To(int32(0)),
		StatsServerPort: ptr.To(int32(0)),
		// Always disable the topology injector in standalone mode. The topology
		// injector adds an EDS local_cluster to the bootstrap config for
		// zone-aware routing, which is both irrelevant outside K8s, but also
		// causes a 15-30s startup delay on the admin /ready endpoint.
		TopologyInjectorDisabled: true,
	}
	args, err := common.BuildProxyArgs(proxyInfra, proxyConfig.Spec.Shutdown, bootstrapConfigOptions, proxyName, false)
	if err != nil {
		return err
	}
	i.runEnvoy(ctx, i.getEnvoyVersion(proxyConfig), proxyName, args)
	return nil
}

// runEnvoy runs the Envoy process with the given arguments and name in a separate goroutine.
func (i *Infra) runEnvoy(ctx context.Context, envoyVersion, name string, args []string) {
	pCtx, cancel := context.WithCancel(ctx)
	exit := make(chan struct{}, 1)
	i.proxyContextMap.Store(name, &proxyContext{cancel: cancel, exit: exit})
	go func() {
		// Run blocks until pCtx is done or the process exits where the latter doesn't happen when
		// Envoy successfully starts up. So, this will not return until pCtx is done in practice.
		defer func() {
			exit <- struct{}{}
		}()
		err := i.envoyRunner(pCtx, args,
			func_e_api.ConfigHome(i.Paths.ConfigHome),
			func_e_api.DataHome(i.Paths.DataHome),
			func_e_api.StateHome(i.Paths.StateHome),
			func_e_api.RuntimeDir(i.Paths.RuntimeDir),
			func_e_api.Out(i.Stdout),
			func_e_api.EnvoyOut(i.Stdout),
			func_e_api.EnvoyErr(i.Stderr),
			func_e_api.EnvoyVersion(envoyVersion))
		if err != nil {
			i.Logger.Error(err, "failed to run envoy")
			// If the Envoy process fails to start, notify an unrecoverable error so that the main control
			// loop can properly handle it.
			i.errors.Store(err)
		}
	}()
}

// DeleteProxyInfra removes the managed host process, if it doesn't exist.
func (i *Infra) DeleteProxyInfra(_ context.Context, infra *ir.Infra) error {
	if infra == nil {
		return errors.New("infra ir is nil")
	}

	proxyInfra := infra.GetProxyInfra()
	proxyName := utils.GetHashedName(proxyInfra.Name, 64)
	i.stopEnvoy(proxyName)
	return nil
}

// stopEnvoy stops the Envoy process by its name. It will block until the process completely stopped.
func (i *Infra) stopEnvoy(proxyName string) {
	value, ok := i.proxyContextMap.LoadAndDelete(proxyName)
	if ok {
		pCtx := value.(*proxyContext)
		pCtx.cancel()    // Cancel causes the Envoy process to exit.
		<-pCtx.exit      // Wait for the Envoy process to completely exit.
		close(pCtx.exit) // Close the channel to avoid leaking.
	}
}

// getEnvoyVersion returns the version of Envoy to use.
func (i *Infra) getEnvoyVersion(proxyConfig *egv1a1.EnvoyProxy) string {
	// Note these helper functions gracefully handle nil pointer dereferencing, so it's safe to
	// chain method calls.
	version := proxyConfig.GetEnvoyProxyProvider().GetEnvoyProxyHostProvider().GetEnvoyVersion()
	if version == "" {
		// If the version is not explicitly set, use the default version EG is built with.
		// This is only populated to a concrete version in release branches.
		// For `main` it may fail. In that case, we return an empty version and let the func-e library
		// decide what version to use.
		// This keeps the old behaviour for backwards compatibility.
		version, _ = extractSemver(i.defaultEnvoyImage)
	}

	if version == "" {
		i.Logger.Info("no explicit Envoy version is set and " +
			"could not extract a default version from the default Envoy image")
	}

	return version
}

// extractSemver takes an image reference like "docker.io/envoyproxy/envoy:distroless-v1.35.0"
// and returns the semver string, e.g. "1.35.0".
func extractSemver(image string) (string, error) {
	// Split to isolate the tag part after the colon
	parts := strings.Split(image, ":")
	if len(parts) < 2 {
		return "", fmt.Errorf("no tag found in default Envoy image reference: %s", image)
	}
	tag := parts[len(parts)-1]

	re := regexp.MustCompile(`\d+\.\d+\.\d+`)
	semver := re.FindString(tag)
	if semver == "" {
		return "", fmt.Errorf("no semver found in tag: %s", tag)
	}
	return semver, nil
}

// convertResolvedMetricSinks converts IR metric sinks to bootstrap format.
func convertResolvedMetricSinks(irSinks []ir.ResolvedMetricSink) []bootstrap.MetricSink {
	result := make([]bootstrap.MetricSink, 0, len(irSinks))
	for _, sink := range irSinks {
		if len(sink.Destination.Settings) == 0 || len(sink.Destination.Settings[0].Endpoints) == 0 {
			continue
		}
		// Metrics are aggregated locally in Envoy and exported to one collector.
		ep := sink.Destination.Settings[0].Endpoints[0]
		ms := bootstrap.MetricSink{
			Address:                  ep.Host,
			Port:                     ep.Port,
			Authority:                sink.Authority,
			ReportCountersAsDeltas:   sink.ReportCountersAsDeltas,
			ReportHistogramsAsDeltas: sink.ReportHistogramsAsDeltas,
			Headers:                  sink.Headers,
		}
		if tls := sink.Destination.Settings[0].TLS; tls != nil {
			ms.TLS = &bootstrap.MetricSinkTLS{
				UseSystemTrustStore: tls.UseSystemTrustStore,
			}
			if tls.SNI != nil {
				ms.TLS.SNI = *tls.SNI
			}
			if tls.CACertificate != nil {
				ms.TLS.CACertificate = tls.CACertificate.Certificate
			}
		}
		result = append(result, ms)
	}
	return result
}
