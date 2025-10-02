// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package host

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	func_e "github.com/tetratelabs/func-e"
	"github.com/tetratelabs/func-e/api"
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
	for name := range i.proxyContextMap {
		i.stopEnvoy(name)
	}
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
	if _, ok := i.proxyContextMap[proxyName]; ok {
		return nil
	}

	proxyConfig := proxyInfra.GetProxyConfig()
	// Disable Prometheus to make envoy running as a host process successfully.
	// TODO: Add Prometheus support to host infra.
	bootstrapConfigOptions := &bootstrap.RenderBootstrapConfigOptions{
		ProxyMetrics: &egv1a1.ProxyMetrics{
			Prometheus: &egv1a1.ProxyPrometheusProvider{
				Disable: true,
			},
		},
		SdsConfig: bootstrap.SdsConfigPath{
			Certificate: filepath.Join(i.sdsConfigPath, common.SdsCertFilename),
			TrustedCA:   filepath.Join(i.sdsConfigPath, common.SdsCAFilename),
		},
		ServiceName:     i.ServiceName,
		XdsServerHost:   ptr.To("0.0.0.0"),
		AdminServerPort: ptr.To(int32(0)),
		StatsServerPort: ptr.To(int32(0)),
	}
	if i.EnvoyGateway != nil {
		bootstrapConfigOptions.TopologyInjectorDisabled = i.EnvoyGateway.TopologyInjectorDisabled()
	}
	args, err := common.BuildProxyArgs(proxyInfra, proxyConfig.Spec.Shutdown, bootstrapConfigOptions, proxyName, false)
	if err != nil {
		return err
	}
	i.runEnvoy(ctx, os.Stdout, i.getEnvoyVersion(proxyConfig), proxyName, args)
	return nil
}

// runEnvoy runs the Envoy process with the given arguments and name in a separate goroutine.
func (i *Infra) runEnvoy(ctx context.Context, out io.Writer, envoyVersion, name string, args []string) {
	pCtx, cancel := context.WithCancel(ctx)
	exit := make(chan struct{}, 1)
	i.proxyContextMap[name] = &proxyContext{cancel: cancel, exit: exit}
	go func() {
		// Run blocks until pCtx is done or the process exits where the latter doesn't happen when
		// Envoy successfully starts up. So, this will not return until pCtx is done in practice.
		defer func() {
			exit <- struct{}{}
		}()
		err := func_e.Run(pCtx, args, api.HomeDir(i.HomeDir), api.Out(out), api.EnvoyVersion(envoyVersion))
		if err != nil {
			i.Logger.Error(err, "failed to run envoy")
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
	if pCtx, ok := i.proxyContextMap[proxyName]; ok {
		pCtx.cancel()    // Cancel causes the Envoy process to exit.
		<-pCtx.exit      // Wait for the Envoy process to completely exit.
		close(pCtx.exit) // Close the channel to avoid leaking.
		delete(i.proxyContextMap, proxyName)
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
