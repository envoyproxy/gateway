package kubernetes

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/message"
)

// ResourceTable is a listing of all of the Kubernetes resources being
// watched.
type ResourceTable struct {
	// Initialized.Wait() will return once each of the maps in the
	// table have been initialized at startup.
	Initialized sync.WaitGroup

	GatewayClasses watchable.Map[string, *gwapiv1b1.GatewayClass]
	Gateways       watchable.Map[types.NamespacedName, *gwapiv1b1.Gateway]
	HTTPRoutes     watchable.Map[types.NamespacedName, *gwapiv1b1.HTTPRoute]
}

// Provider is the scaffolding for the Kubernetes provider. It sets up dependencies
// and defines the topology of the provider and its managed components, wiring
// them together.
type Provider struct {
	client  client.Client
	manager manager.Manager
}

// New creates a new Provider from the provided EnvoyGateway.
func New(controllerName string, logger logr.Logger, resources *message.ProviderResources) (*Provider, error) {
	cfg, err := ctrl.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get kubeconfig: %w", err)
	}
	// TODO: Decide which mgr opts should be exposed through envoygateway.provider.kubernetes API.
	mgrOpts := manager.Options{
		Scheme:             envoygateway.GetScheme(),
		Logger:             logger,
		LeaderElection:     false,
		LeaderElectionID:   "5b9825d2.gateway.envoyproxy.io",
		MetricsBindAddress: ":8080",
	}
	mgr, err := ctrl.NewManager(cfg, mgrOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create manager: %w", err)
	}

	// Create and register the controllers with the manager.
	if err := newGatewayClassController(controllerName, mgr, logger, resources); err != nil {
		return nil, fmt.Errorf("failed to create gatewayclass controller: %w", err)
	}
	if err := newGatewayController(controllerName, mgr, logger, resources); err != nil {
		return nil, fmt.Errorf("failed to create gateway controller: %w", err)
	}
	if err := newHTTPRouteController(mgr, svr, resourceTable); err != nil {
		return nil, fmt.Errorf("failed to create httproute controller: %w", err)
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
