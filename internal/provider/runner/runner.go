// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"context"
	"fmt"

	"k8s.io/client-go/rest"
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

func (r *Runner) Name() string {
	return string(egv1a1.LogComponentProviderRunner)
}

// Start the provider runner
func (r *Runner) Start(ctx context.Context) (err error) {
	r.Logger = r.Logger.WithName(r.Name()).WithValues("runner", r.Name())

	var p provider.Provider
	switch r.EnvoyGateway.Provider.Type {
	case egv1a1.ProviderTypeKubernetes:
		var cfg *rest.Config
		cfg, err = ctrl.GetConfig()
		if err != nil {
			return fmt.Errorf("failed to get kubeconfig: %w", err)
		}
		p, err = kubernetes.New(cfg, &r.Config.Server, r.ProviderResources)
		if err != nil {
			return fmt.Errorf("failed to create provider %s: %w", egv1a1.ProviderTypeKubernetes, err)
		}

	case egv1a1.ProviderTypeFile:
		p, err = file.New(&r.Config.Server, r.ProviderResources)
		if err != nil {
			return fmt.Errorf("failed to create provider %s: %w", egv1a1.ProviderTypeFile, err)
		}

	default:
		// Unsupported provider.
		return fmt.Errorf("unsupported provider type %v", r.EnvoyGateway.Provider.Type)
	}

	r.Logger.Info("Running provider", "type", p.Type())
	go func() {
		if err = p.Start(ctx); err != nil {
			r.Logger.Error(err, "unable to start provider")
		}
	}()

	return nil
}
