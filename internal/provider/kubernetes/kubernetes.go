package kubernetes

import (
	"context"
	"fmt"

	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
)

// Provider is the scaffolding for the Kubernetes provider. It sets up dependencies
// and defines the topology of the provider and its managed components, wiring
// them together.
type Provider struct {
	client  client.Client
	manager manager.Manager
}

// New creates a new Provider from the provided EnvoyGateway.
func New(cfg *rest.Config, svr *config.Server) (*Provider, error) {
	// TODO: Decide which mgr opts should be exposed through envoygateway.provider.kubernetes API.
	mgrOpts := manager.Options{
		Scheme:             envoygateway.GetScheme(),
		Logger:             svr.Logger,
		LeaderElection:     false,
		LeaderElectionID:   "5b9825d2.gateway.envoyproxy.io",
		MetricsBindAddress: ":8080",
	}
	mgr, err := ctrl.NewManager(cfg, mgrOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create manager: %w", err)
	}

	// Create and register the controllers with the manager.
	if err := newGatewayClassController(mgr, svr); err != nil {
		return nil, fmt.Errorf("failed to create gatewayclass controller: %w", err)
	}
	// TODO: Add gateway, httproute, etc. controllers.

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
