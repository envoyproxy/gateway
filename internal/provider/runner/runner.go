// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/provider/kubernetes"
	"github.com/envoyproxy/gateway/internal/runner"
)

func Register(resources Resources, globalConfig config.Server) {
	runner.Manager().Register(New(resources, globalConfig), runner.RootParentRunner)
}

type providerRunner struct {
	*runner.GenericRunner[Resources]
}

type Resources struct {
	ProviderResources *message.ProviderResources
}

func New(resources Resources, globalConfig config.Server) *providerRunner {
	return &providerRunner{GenericRunner: runner.New(string(v1alpha1.LogComponentProviderRunner), resources, globalConfig)}
}

// Start starts the provider runner.
func (r *providerRunner) Start(ctx context.Context) error {
	r.Init(ctx)

	if r.Config().EnvoyGateway.Provider.Type == v1alpha1.ProviderTypeKubernetes {
		r.Logger.Info("Using provider", "type", v1alpha1.ProviderTypeKubernetes)
		cfg, err := ctrl.GetConfig()
		if err != nil {
			return fmt.Errorf("failed to get kubeconfig: %w", err)
		}
		kubernetes.Register(kubernetes.Resources{ProviderResources: r.Resources.ProviderResources, Cfg: cfg}, r.Server)
	} else {
		// Unsupported provider.
		return fmt.Errorf("unsupported provider type %v", r.EnvoyGateway.Provider.Type)
	}

	go r.SubscribeAndTranslate(ctx)

	r.Logger.Info("started")
	return nil
}

// SubscribeAndTranslate implements the generic runner logic
func (r *providerRunner) SubscribeAndTranslate(ctx context.Context) {
	if err := runner.Manager().Start(ctx, kubernetes.Runner); err != nil {
		r.Logger.Error(err, "unable to start provider")
	}
}

func (r *providerRunner) ShutDown(ctx context.Context) {
	r.Resources.ProviderResources.Close()
}
