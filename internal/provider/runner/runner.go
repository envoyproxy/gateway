// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/provider"
	"github.com/envoyproxy/gateway/internal/provider/file"
	"github.com/envoyproxy/gateway/internal/provider/kubernetes"
)

type Config struct {
	config.Server
	ProviderResources *message.ProviderResources
}

type Runner struct {
	Config
}

func New(cfg *Config) *Runner {
	return &Runner{Config: *cfg}
}

// Close implements Runner interface.
func (r *Runner) Close() error { return nil }

// Name implements Runner interface.
func (r *Runner) Name() string {
	return string(egv1a1.LogComponentProviderRunner)
}

// Start implements Runner interface.
func (r *Runner) Start(ctx context.Context) (err error) {
	r.Logger = r.Logger.WithName(r.Name()).WithValues("runner", r.Name())

	var p provider.Provider
	switch r.EnvoyGateway.Provider.Type {
	case egv1a1.ProviderTypeKubernetes:
		p, err = r.createKubernetesProvider(ctx)
		if err != nil {
			return fmt.Errorf("failed to create kubernetes provider: %w", err)
		}

	case egv1a1.ProviderTypeCustom:
		p, err = r.createCustomResourceProvider()
		if err != nil {
			return fmt.Errorf("failed to create custom provider: %w", err)
		}

	default:
		// Unsupported provider.
		return fmt.Errorf("unsupported provider type %v", r.EnvoyGateway.Provider.Type)
	}

	r.Logger.Info("Running provider", "type", p.Type())
	go func() {
		if err := p.Start(ctx); err != nil {
			r.Logger.Error(err, "unable to start provider")
		}
	}()

	return nil
}

func (r *Runner) createKubernetesProvider(ctx context.Context) (*kubernetes.Provider, error) {
	cfg, err := ctrl.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get kubeconfig: %w", err)
	}

	p, err := kubernetes.New(ctx, cfg, &r.Server, r.ProviderResources)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider %s: %w", egv1a1.ProviderTypeKubernetes, err)
	}

	return p, err
}

func (r *Runner) createCustomResourceProvider() (p provider.Provider, err error) {
	switch r.EnvoyGateway.Provider.Custom.Resource.Type {
	case egv1a1.ResourceProviderTypeFile:
		p, err = file.New(&r.Server, r.ProviderResources)
		if err != nil {
			return nil, fmt.Errorf("failed to create provider %s: %w", egv1a1.ProviderTypeCustom, err)
		}

	default:
		return nil, fmt.Errorf("unsupported resource provider type")
	}

	return
}
