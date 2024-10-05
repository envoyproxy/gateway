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
	"github.com/envoyproxy/gateway/internal/common"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/provider"
	"github.com/envoyproxy/gateway/internal/provider/file"
	"github.com/envoyproxy/gateway/internal/provider/kubernetes"
)

var _ common.Runner = &Runner{}

type Config struct {
	ServerCfg         *config.Server
	ProviderResources *message.ProviderResources
	Logger            logging.Logger
}

type Runner struct {
	Config
	p   provider.Provider
	ctx context.Context
}

func New(cfg *Config) *Runner {
	return &Runner{Config: *cfg}
}

func (r *Runner) Name() string {
	return string(egv1a1.LogComponentProviderRunner)
}

// Start the provider runner
func (r *Runner) Start(ctx context.Context) (err error) {
	r.Logger = r.ServerCfg.Logger.WithName(r.Name())
	r.ctx = ctx

	switch r.ServerCfg.EnvoyGateway.Provider.Type {
	case egv1a1.ProviderTypeKubernetes:
		r.p, err = r.createKubernetesProvider()
		if err != nil {
			return fmt.Errorf("failed to create kubernetes provider: %w", err)
		}
	case egv1a1.ProviderTypeCustom:
		r.p, err = r.createCustomResourceProvider()
		if err != nil {
			return fmt.Errorf("failed to create custom provider: %w", err)
		}
	default:
		// Unsupported provider.
		return fmt.Errorf("unsupported provider type %v", r.ServerCfg.EnvoyGateway.Provider.Type)
	}

	r.Logger.Info("Running provider", "type", r.p.Type())
	go func() {
		if err = r.p.Start(ctx); err != nil {
			r.Logger.Error(err, "unable to start provider")
		}
	}()

	return nil
}

func (r *Runner) Reload(serverCfg *config.Server) error {
	r.Logger = serverCfg.Logger.WithName(r.Name())
	r.ServerCfg = serverCfg

	r.Logger.Info("reloaded")
	r.p.Stop()
	return r.Start(r.ctx)
}

func (r *Runner) createKubernetesProvider() (*kubernetes.Provider, error) {
	rest, err := ctrl.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get kubeconfig: %w", err)
	}

	p, err := kubernetes.New(rest, r.ServerCfg, r.ProviderResources)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider %s: %w", egv1a1.ProviderTypeKubernetes, err)
	}

	return p, err
}

func (r *Runner) createCustomResourceProvider() (p provider.Provider, err error) {
	switch r.ServerCfg.EnvoyGateway.Provider.Custom.Resource.Type {
	case egv1a1.ResourceProviderTypeFile:
		p, err = file.New(r.ServerCfg, r.ProviderResources)
		if err != nil {
			return nil, fmt.Errorf("failed to create provider %s: %w", egv1a1.ProviderTypeCustom, err)
		}
	default:
		return nil, fmt.Errorf("unsupported resource provider type")
	}

	return
}
