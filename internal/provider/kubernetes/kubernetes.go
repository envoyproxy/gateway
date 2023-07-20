// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"fmt"

	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/runner"
	"github.com/envoyproxy/gateway/internal/status"
)

const Runner = "kubernetes-provider"

func Register(resources Resources, globalConfig config.Server) {
	r, err := New(resources, globalConfig)
	if err == nil {
		runner.Manager().Register(r, string(v1alpha1.LogComponentProviderRunner))
	}
}

// kubernetesRunner is the scaffolding for the Kubernetes provider. It sets up dependencies
// and defines the topology of the provider and its managed components, wiring
// them together.
type kubernetesRunner struct {
	*runner.GenericRunner[Resources]

	client  client.Client
	manager manager.Manager
}

type Resources struct {
	ProviderResources *message.ProviderResources
	Cfg               *rest.Config
}

// New creates a new Provider from the provided EnvoyGateway.
func New(resources Resources, svr config.Server) (*kubernetesRunner, error) {
	// TODO: Decide which mgr opts should be exposed through envoygateway.provider.kubernetes API.
	mgrOpts := manager.Options{
		Scheme:                 envoygateway.GetScheme(),
		Logger:                 svr.Logger.Logger,
		LeaderElection:         false,
		HealthProbeBindAddress: ":8081",
		LeaderElectionID:       "5b9825d2.gateway.envoyproxy.io",
		MetricsBindAddress:     ":8080",
	}

	if svr.EnvoyGateway.Provider != nil &&
		svr.EnvoyGateway.Provider.Kubernetes != nil &&
		(svr.EnvoyGateway.Provider.Kubernetes.Watch != nil) &&
		(len(svr.EnvoyGateway.Provider.Kubernetes.Watch.Namespaces) > 0) {
		mgrOpts.Cache.Namespaces = svr.EnvoyGateway.Provider.Kubernetes.Watch.Namespaces
	}

	mgr, err := ctrl.NewManager(resources.Cfg, mgrOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create manager: %w", err)
	}

	updateHandler := status.NewUpdateHandler(mgr.GetLogger(), mgr.GetClient())
	if err := mgr.Add(updateHandler); err != nil {
		return nil, fmt.Errorf("failed to add status update handler %v", err)
	}

	// Create and register the controllers with the manager.
	if err := newGatewayAPIController(mgr, &svr, updateHandler.Writer(), resources.ProviderResources); err != nil {
		return nil, fmt.Errorf("failted to create gatewayapi controller: %w", err)
	}

	// Add health check health probes.
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		return nil, fmt.Errorf("unable to set up health check: %w", err)
	}

	// Add ready check health probes.
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		return nil, fmt.Errorf("unable to set up ready check: %w", err)
	}

	return &kubernetesRunner{
		manager:       mgr,
		client:        mgr.GetClient(),
		GenericRunner: runner.New(Runner, resources, svr),
	}, nil
}

// Start starts the Provider synchronously until a message is received from ctx.
func (r *kubernetesRunner) Start(ctx context.Context) error {
	r.Init(ctx)
	r.Logger.Info("started")

	errChan := make(chan error)
	go func() {
		errChan <- r.manager.Start(ctx)
	}()

	// Wait for the manager to exit or an explicit stop.
	select {
	case <-ctx.Done():
		return nil
	case err := <-errChan:
		return err
	}
}

func (r *kubernetesRunner) ShutDown(ctx context.Context) {
	r.Resources.ProviderResources.Close()
}

func (r *kubernetesRunner) SubscribeAndTranslate(ctx context.Context) {}
