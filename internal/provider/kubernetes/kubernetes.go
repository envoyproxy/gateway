package kubernetes

import (
	"context"
	"fmt"
	"sync"

	"github.com/telepresenceio/watchable"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
)

// ResourceTable is a listing of all of the Kubernetes resources bing
// watched.
type ResourceTable struct {
	// Initialized.Wait() will return once each of the maps in the
	// table have been initialized at startup.
	Initialized sync.WaitGroup

	GatewayClasses watchable.Map[string, *gwapiv1b1.GatewayClass]
	Gateways       watchable.Map[types.NamespacedName, *gwapiv1b1.Gateway]
}

// Provider is the scaffolding for the Kubernetes provider. It sets up dependencies
// and defines the topology of the provider and its managed components, wiring
// them together.
type Provider struct {
	client        client.Client
	manager       manager.Manager
	resourceTable *ResourceTable
}

// New creates a new Provider from the provided EnvoyGateway.
func New(cfg *rest.Config, svr *config.Server, resourceTable *ResourceTable) (*Provider, error) {
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
	if err := newGatewayClassController(mgr, svr, resourceTable); err != nil {
		return nil, fmt.Errorf("failed to create gatewayclass controller: %w", err)
	}
	if err := newGatewayController(mgr, svr, resourceTable); err != nil {
		return nil, fmt.Errorf("failed to create gateway controller: %w", err)
	}
	// TODO: Add httproute controllers.
	// xref: https://github.com/envoyproxy/gateway/issues/163

	return &Provider{
		manager:       mgr,
		client:        mgr.GetClient(),
		resourceTable: resourceTable,
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

// Resources returns an updating table of all of the resources being
// watched.
func (p *Provider) Resources() *ResourceTable {
	return p.resourceTable
}
