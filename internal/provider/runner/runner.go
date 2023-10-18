// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/message"
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

func (r *Runner) Name() string {
	return string(v1alpha1.LogComponentProviderRunner)
}

// Start the provider runner
func (r *Runner) Start(ctx context.Context) (err error) {
	r.Logger = r.Logger.WithName(r.Name()).WithValues("runner", r.Name())
	if r.EnvoyGateway.Provider.Type == v1alpha1.ProviderTypeKubernetes {
		r.Logger.Info("Using provider", "type", v1alpha1.ProviderTypeKubernetes)
		cfg, err := ctrl.GetConfig()
		if err != nil {
			return fmt.Errorf("failed to get kubeconfig: %w", err)
		}
		p, err := kubernetes.New(cfg, &r.Config.Server, r.ProviderResources)
		if err != nil {
			return fmt.Errorf("failed to create provider %s: %w", v1alpha1.ProviderTypeKubernetes, err)
		}
		go func() {
			err := p.Start(ctx)
			if err != nil {
				r.Logger.Error(err, "unable to start provider")
			}
		}()
		return nil
	}
	// Unsupported provider.
	return fmt.Errorf("unsupported provider type %v", r.EnvoyGateway.Provider.Type)
}
