// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"fmt"
	"time"

	"k8s.io/client-go/rest"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/config"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/envoyproxy/gateway/internal/envoygateway"
	ec "github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/message"
)

// Provider is the scaffolding for the Kubernetes provider. It sets up dependencies
// and defines the topology of the provider and its managed components, wiring
// them together.
type Provider struct {
	client  client.Client
	manager manager.Manager
}

// New creates a new Provider from the provided EnvoyGateway.
func New(cfg *rest.Config, svr *ec.Server, resources *message.ProviderResources) (*Provider, error) {
	// TODO: Decide which mgr opts should be exposed through envoygateway.provider.kubernetes API.

	mgrOpts := manager.Options{
		Scheme:                  envoygateway.GetScheme(),
		Logger:                  svr.Logger.Logger,
		HealthProbeBindAddress:  ":8081",
		LeaderElectionID:        "5b9825d2.gateway.envoyproxy.io",
		LeaderElectionNamespace: svr.Namespace,
	}

	if !ptr.Deref(svr.EnvoyGateway.Provider.Kubernetes.LeaderElection.Disable, false) {
		mgrOpts.LeaderElection = true
		if svr.EnvoyGateway.Provider.Kubernetes.LeaderElection.LeaseDuration != nil {
			ld, err := time.ParseDuration(string(*svr.EnvoyGateway.Provider.Kubernetes.LeaderElection.LeaseDuration))
			if err != nil {
				return nil, err
			}
			mgrOpts.LeaseDuration = ptr.To(ld)
		}

		if svr.EnvoyGateway.Provider.Kubernetes.LeaderElection.RetryPeriod != nil {
			rp, err := time.ParseDuration(string(*svr.EnvoyGateway.Provider.Kubernetes.LeaderElection.RetryPeriod))
			if err != nil {
				return nil, err
			}
			mgrOpts.RetryPeriod = ptr.To(rp)
		}

		if svr.EnvoyGateway.Provider.Kubernetes.LeaderElection.RenewDeadline != nil {
			rd, err := time.ParseDuration(string(*svr.EnvoyGateway.Provider.Kubernetes.LeaderElection.RenewDeadline))
			if err != nil {
				return nil, err
			}
			mgrOpts.RenewDeadline = ptr.To(rd)
		}
		mgrOpts.Controller = config.Controller{NeedLeaderElection: ptr.To(false)}
	}

	if svr.EnvoyGateway.NamespaceMode() {
		mgrOpts.Cache.DefaultNamespaces = make(map[string]cache.Config)
		for _, watchNS := range svr.EnvoyGateway.Provider.Kubernetes.Watch.Namespaces {
			mgrOpts.Cache.DefaultNamespaces[watchNS] = cache.Config{}
		}
	}
	mgr, err := ctrl.NewManager(cfg, mgrOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create manager: %w", err)
	}

	updateHandler := NewUpdateHandler(mgr.GetLogger(), mgr.GetClient())
	if err := mgr.Add(updateHandler); err != nil {
		return nil, fmt.Errorf("failed to add status update handler %w", err)
	}

	// Create and register the controllers with the manager.
	if err := newGatewayAPIController(mgr, svr, updateHandler.Writer(), resources); err != nil {
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

	// Emit elected & continue with deployment of infra resources
	go func() {
		<-mgr.Elected()
		close(svr.Elected)
	}()

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
