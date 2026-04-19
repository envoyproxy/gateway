// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

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
	RunnerErrors      *message.RunnerErrors
}

type Runner struct {
	Config
	provider provider.Provider
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
	if r.provider == nil {
		return fmt.Errorf("provider must be prepared before start")
	}

	r.Logger.Info("Running provider", "type", r.provider.Type())
	go func() {
		if err := r.provider.Start(ctx); err != nil {
			r.Logger.Error(err, "unable to start provider")
		}
	}()

	return nil
}

// PrepareProvider constructs the provider before starting the runner.
// It is used to prepare the kubernetes client so it can be shared with the infrastructure manager.
func (r *Runner) PrepareProvider(ctx context.Context) (provider.Provider, error) {
	if r.provider != nil {
		return r.provider, nil
	}

	errNotifier := message.RunnerErrorNotifier{RunnerName: r.Name(), RunnerErrors: r.RunnerErrors}

	switch r.EnvoyGateway.Provider.Type {
	case egv1a1.ProviderTypeKubernetes:
		p, err := r.createKubernetesProvider(ctx, errNotifier)
		if err != nil {
			return nil, fmt.Errorf("failed to create kubernetes provider: %w", err)
		}
		r.provider = p
	case egv1a1.ProviderTypeCustom:
		p, err := r.createCustomResourceProvider(ctx, errNotifier)
		if err != nil {
			return nil, fmt.Errorf("failed to create custom provider: %w", err)
		}
		r.provider = p
	default:
		return nil, fmt.Errorf("unsupported provider type %v", r.EnvoyGateway.Provider.Type)
	}

	return r.provider, nil
}

func (r *Runner) KubernetesClient() client.Client {
	p, ok := r.provider.(*kubernetes.Provider)
	if !ok {
		return nil
	}

	return p.GetClient()
}

func (r *Runner) createKubernetesProvider(ctx context.Context, errors message.RunnerErrorNotifier) (*kubernetes.Provider, error) {
	cfg, err := ctrl.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get kubeconfig: %w", err)
	}

	p, err := kubernetes.New(ctx, cfg, &r.Server, r.ProviderResources, errors)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider %s: %w", egv1a1.ProviderTypeKubernetes, err)
	}

	return p, err
}

func (r *Runner) createCustomResourceProvider(ctx context.Context, errors message.RunnerErrorNotifier) (provider.Provider, error) {
	switch r.EnvoyGateway.Provider.Custom.Resource.Type {
	case egv1a1.ResourceProviderTypeFile:
		p, err := file.New(ctx, &r.Server, r.ProviderResources, errors)
		if err != nil {
			return nil, fmt.Errorf("failed to create provider %s: %w", egv1a1.ProviderTypeCustom, err)
		}
		return p, err

	default:
		return nil, fmt.Errorf("unsupported resource provider type")
	}
}
