// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package host

import (
	"context"
	"errors"
	"fmt"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"io"
	"os"
	"path/filepath"

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
		XdsServerHost:    ptr.To("0.0.0.0"),
		AdminServerPort:  ptr.To(int32(0)),
		StatsServerPort:  ptr.To(int32(0)),
		LocalClusterName: ptr.To(fmt.Sprintf("%s-%s", config.EnvoyPrefix, proxyName)),
	}

	args, err := common.BuildProxyArgs(proxyInfra, proxyConfig.Spec.Shutdown, bootstrapConfigOptions, proxyName, false)
	if err != nil {
		return err
	}
	i.runEnvoy(ctx, os.Stdout, proxyName, args)
	return nil
}

// runEnvoy runs the Envoy process with the given arguments and name in a separate goroutine.
func (i *Infra) runEnvoy(ctx context.Context, out io.Writer, name string, args []string) {
	pCtx, cancel := context.WithCancel(ctx)
	exit := make(chan struct{}, 1)
	i.proxyContextMap[name] = &proxyContext{cancel: cancel, exit: exit}
	go func() {
		// Run blocks until pCtx is done or the process exits where the latter doesn't happen when
		// Envoy successfully starts up. So, this will not return until pCtx is done in practice.
		defer func() {
			exit <- struct{}{}
		}()
		err := func_e.Run(pCtx, args, api.HomeDir(i.HomeDir), api.Out(out))
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
