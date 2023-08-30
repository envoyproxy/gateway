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
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/status"
)

// Provider is the scaffolding for the Kubernetes provider. It sets up dependencies
// and defines the topology of the provider and its managed components, wiring
// them together.
type Provider struct {
	client  client.Client
	manager manager.Manager
}

// New creates a new Provider from the provided EnvoyGateway.
func New(cfg *rest.Config, svr *config.Server, resources *message.ProviderResources, eStatuses *message.EnvoyPatchPolicyStatuses) (*Provider, error) {
	// TODO: Decide which mgr opts should be exposed through envoygateway.provider.kubernetes API.
	mgrOpts := manager.Options{
		Scheme:                 envoygateway.GetScheme(),
		Logger:                 svr.Logger.Logger,
		LeaderElection:         false,
		HealthProbeBindAddress: ":8081",
		LeaderElectionID:       "5b9825d2.gateway.envoyproxy.io",
		Metrics: metricsserver.Options{
			BindAddress: ":8080",
		},
	}

	if svr.EnvoyGateway.Provider != nil &&
		svr.EnvoyGateway.Provider.Kubernetes != nil &&
		(svr.EnvoyGateway.Provider.Kubernetes.Watch != nil) &&
		(len(svr.EnvoyGateway.Provider.Kubernetes.Watch.Namespaces) > 0) {
		for _, watchNS := range svr.EnvoyGateway.Provider.Kubernetes.Watch.Namespaces {
			mgrOpts.Cache.DefaultNamespaces[watchNS] = cache.Config{}
		}
	}

	mgr, err := ctrl.NewManager(cfg, mgrOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create manager: %w", err)
	}

	updateHandler := status.NewUpdateHandler(mgr.GetLogger(), mgr.GetClient())
	if err := mgr.Add(updateHandler); err != nil {
		return nil, fmt.Errorf("failed to add status update handler %v", err)
	}

	// Create and register the controllers with the manager.
	if err := newGatewayAPIController(mgr, svr, updateHandler.Writer(), resources, eStatuses); err != nil {
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

	return &Provider{
		manager: mgr,
		client:  mgr.GetClient(),
	}, nil
}

// Start starts the Provider synchronously until a message is received from ctx.
func (p *Provider) Start(ctx context.Context) error {
	errChan := make(chan error)
	go func() {
		errChan <- p.manager.Start(ctx)
	}()

	// Wait for the manager to exit or an explicit stop.
	select {
	case <-ctx.Done():
		return nil
	case err := <-errChan:
		return err
	}
}
