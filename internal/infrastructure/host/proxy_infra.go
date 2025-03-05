// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package host

import (
	"context"
	"errors"
	"os/signal"
	"path/filepath"
	"syscall"

	funcE "github.com/tetratelabs/func-e/api"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/infrastructure/common"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils"
	"github.com/envoyproxy/gateway/internal/xds/bootstrap"
)

type proxyContext struct {
	ctx    context.Context
	cancel context.CancelFunc
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
		XdsServerHost:   ptr.To("0.0.0.0"),
		WasmServerPort:  ptr.To(int32(0)),
		AdminServerPort: ptr.To(int32(0)),
		StatsServerPort: ptr.To(int32(0)),
	}

	args, err := common.BuildProxyArgs(proxyInfra, proxyConfig.Spec.Shutdown, bootstrapConfigOptions, proxyName)
	if err != nil {
		return err
	}

	// Create a new context for up-running proxy.
	pCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	i.proxyContextMap[proxyName] = &proxyContext{ctx: pCtx, cancel: stop}
	return funcE.Run(pCtx, args, funcE.HomeDir(i.HomeDir))
}

// DeleteProxyInfra removes the managed host process, if it doesn't exist.
func (i *Infra) DeleteProxyInfra(ctx context.Context, infra *ir.Infra) error {
	if infra == nil {
		return errors.New("infra ir is nil")
	}

	proxyInfra := infra.GetProxyInfra()
	proxyName := utils.GetHashedName(proxyInfra.Name, 64)
	if pCtx, ok := i.proxyContextMap[proxyName]; ok {
		pCtx.cancel()
	}

	// Return directly if the proxy is already stopped.
	return nil
}
