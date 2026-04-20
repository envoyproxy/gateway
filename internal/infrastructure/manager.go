// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package infrastructure

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/infrastructure/host"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/message"
)

var (
	_ Manager = (*kubernetes.Infra)(nil)
	_ Manager = (*host.Infra)(nil)
)

type kubernetesClientContextKey struct{}

type kubernetesClientHolder struct {
	client client.Client
}

func WithKubernetesClientHolder(ctx context.Context) context.Context {
	return context.WithValue(ctx, kubernetesClientContextKey{}, &kubernetesClientHolder{})
}

func SetKubernetesClient(ctx context.Context, cli client.Client) {
	holder, _ := ctx.Value(kubernetesClientContextKey{}).(*kubernetesClientHolder)
	if holder != nil {
		holder.client = cli
	}
}

func KubernetesClientFromContext(ctx context.Context) client.Client {
	holder, _ := ctx.Value(kubernetesClientContextKey{}).(*kubernetesClientHolder)
	if holder == nil {
		return nil
	}
	return holder.client
}

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
func NewManager(ctx context.Context, cfg *config.Server, logger logging.Logger, errors message.RunnerErrorNotifier) (mgr Manager, err error) {
	switch cfg.EnvoyGateway.Provider.Type {
	case egv1a1.ProviderTypeKubernetes:
		cli := KubernetesClientFromContext(ctx)
		if cli == nil {
			return nil, fmt.Errorf("kubernetes client not found in context")
		}
		mgr = kubernetes.NewInfra(cli, cfg, errors)
	case egv1a1.ProviderTypeCustom:
		mgr, err = newManagerForCustom(ctx, cfg, logger, errors)
	}

	if err != nil {
		return nil, err
	}
	return mgr, nil
}

func newManagerForCustom(ctx context.Context, cfg *config.Server, logger logging.Logger, errors message.RunnerErrorNotifier) (Manager, error) {
	infra := cfg.EnvoyGateway.Provider.Custom.Infrastructure
	switch infra.Type {
	case egv1a1.InfrastructureProviderTypeHost:
		return host.NewInfra(ctx, cfg, logger, errors)
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", infra.Type)
	}
}
