package runner

import (
	"context"

	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/message"
)

type Config struct {
	config.Server
	ProviderResources *message.ProviderResources
	XdsIR             *message.XdsIR
	InfraIR           *message.InfraIR
}

type Runner struct {
	Config
}

func New(cfg *Config) *Runner {
	return &Runner{Config: *cfg}
}

func (r *Runner) Name() string {
	return "gateway-api"
}

// Start starts the gateway-api translator runner
func (r *Runner) Start(ctx context.Context) error {
	r.Logger = r.Logger.WithValues("runner", r.Name())
	// Wait until provider resources have been initialized during startup
	r.ProviderResources.Initialized.Wait()
	go r.subscribeAndTranslate(ctx)

	return nil
}

func (r *Runner) subscribeAndTranslate(ctx context.Context) {
	// Subscribe to resources
	gatewayClassesCh := r.ProviderResources.GatewayClasses.Subscribe(ctx)
	gatewaysCh := r.ProviderResources.Gateways.Subscribe(ctx)
	httpRoutesCh := r.ProviderResources.HTTPRoutes.Subscribe(ctx)
	for ctx.Err() == nil {
		var in gatewayapi.Resources
		// Receive subscribed resource notifications
		select {
		case <-gatewayClassesCh:
		case <-gatewaysCh:
		case <-httpRoutesCh:
		}
		// Load all resources required for translation
		in.Gateways = r.ProviderResources.GetGateways()
		in.HTTPRoutes = r.ProviderResources.GetHTTPRoutes()
		gatewayClasses := r.ProviderResources.GetGatewayClasses()
		// Fetch the first gateway class since there should be only 1
		// gateway class linked to this controller
		t := &gatewayapi.Translator{
			GatewayClassName: v1beta1.ObjectName(gatewayClasses[0].GetName()),
		}
		// Translate to IR
		result := t.Translate(&in)
		// Publish the IR
		// Use the service name as the key to ensure there is always
		// one element in the map
		r.XdsIR.Store(r.Name(), result.XdsIR)
	}
	r.Logger.Info("shutting down")
}
