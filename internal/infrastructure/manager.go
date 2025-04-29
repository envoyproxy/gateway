// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package infrastructure

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"
	clicfg "sigs.k8s.io/controller-runtime/pkg/client/config"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/infrastructure/host"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/logging"
)

var (
	_ Manager = (*kubernetes.Infra)(nil)
	_ Manager = (*host.Infra)(nil)
)

// Manager provides the scaffolding for managing infrastructure.
type Manager interface {
	// Close is called when Envoy Gateway is shutting down, it can be used to block until all resources are cleaned up.
	Close() error
	// CreateOrUpdateProxyInfra creates or updates infra.
	CreateOrUpdateProxyInfra(ctx context.Context, infra *ir.Infra) error
	// DeleteProxyInfra deletes infra.
	DeleteProxyInfra(ctx context.Context, infra *ir.Infra) error
	// CreateOrUpdateRateLimitInfra creates or updates rate limit infra.
	CreateOrUpdateRateLimitInfra(ctx context.Context) error
	// DeleteRateLimitInfra deletes rate limit infra.
	DeleteRateLimitInfra(ctx context.Context) error
}

// NewManager returns a new infrastructure Manager.
func NewManager(ctx context.Context, cfg *config.Server, logger logging.Logger) (mgr Manager, err error) {
	switch cfg.EnvoyGateway.Provider.Type {
	case egv1a1.ProviderTypeKubernetes:
		mgr, err = newManagerForKubernetes(cfg)
	case egv1a1.ProviderTypeCustom:
		mgr, err = newManagerForCustom(ctx, cfg, logger)
	}

	if err != nil {
		return nil, err
	}
	return mgr, nil
}

func newManagerForKubernetes(cfg *config.Server) (Manager, error) {
	clientConfig := clicfg.GetConfigOrDie()
	clientConfig.QPS, clientConfig.Burst = cfg.EnvoyGateway.Provider.Kubernetes.Client.RateLimit.GetQPSAndBurst()
	cli, err := client.New(clientConfig, client.Options{Scheme: envoygateway.GetScheme()})
	if err != nil {
		return nil, err
	}
	return kubernetes.NewInfra(cli, cfg), nil
}

func newManagerForCustom(ctx context.Context, cfg *config.Server, logger logging.Logger) (Manager, error) {
	infra := cfg.EnvoyGateway.Provider.Custom.Infrastructure
	switch infra.Type {
	case egv1a1.InfrastructureProviderTypeHost:
		return host.NewInfra(ctx, cfg, logger)
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", infra.Type)
	}
}
