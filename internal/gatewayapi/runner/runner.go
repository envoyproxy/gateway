package runner

import (
	"context"

	"sigs.k8s.io/gateway-api/apis/v1beta1"
	"sigs.k8s.io/yaml"

	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/ir"
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
	go r.subscribeAndTranslateInfra(ctx)
	go r.subscribeAndTranslateXds(ctx)
	r.Logger.Info("started")
	return nil
}

func (r *Runner) subscribeAndTranslateInfra(ctx context.Context) {
	r.Logger.Info("done initializing provider infra resources")
	// Subscribe to infra resources
	gatewayClassesCh := r.ProviderResources.GatewayClasses.Subscribe(ctx)
	gatewaysCh := r.ProviderResources.Gateways.Subscribe(ctx)

	// Wait until provider infra resources have been initialized during startup.
	r.ProviderResources.Initialized.Wait()
	for ctx.Err() == nil {
		var in gatewayapi.Resources
		// Receive subscribed resource notifications.
		select {
		case <-gatewayClassesCh:
		case <-gatewaysCh:
		}
		r.Logger.Info("received infra resource notifications")
		// Load all resources required for infra IR translation.
		in.Gateways = r.ProviderResources.GetGateways()
		gatewayClasses := r.ProviderResources.GetGatewayClasses()
		// Fetch the first gatewayclass since there should only be 1
		// linked to this controller.
		switch {
		case gatewayClasses == nil:
			// Envoy Gateway startup.
			continue
		case gatewayClasses[0] == nil:
			// A nil ProxyInfra tells the Infra Manager to delete the managed proxy infra.
			r.InfraIR.Store(r.Name(), &ir.Infra{Proxy: nil})
		default:
			r.Logger.Info("translating infra ir")
			// Translate and publish IRs.
			t := &gatewayapi.Translator{
				GatewayClassName: v1beta1.ObjectName(gatewayClasses[0].GetName()),
			}
			// Translate to infra IR.
			result := t.TranslateInfra(&in)
			yamlInfraIR, _ := yaml.Marshal(&result.InfraIR)
			r.Logger.Info(string(yamlInfraIR))

			// Publish the infra IR. Use the service name as the key
			// to ensure there is always one element in the map.
			// Also validate the ir before sending it.
			if err := result.InfraIR.Validate(); err != nil {
				r.Logger.Error(err, "unable to validate infra ir, skipped sending it")
			} else {
				r.InfraIR.Store(r.Name(), result.InfraIR)
			}
		}
	}
	r.Logger.Info("shutting down")
}

func (r *Runner) subscribeAndTranslateXds(ctx context.Context) {
	r.Logger.Info("done initializing provider xds resources")
	// Subscribe to resources.
	gatewayClassesCh := r.ProviderResources.GatewayClasses.Subscribe(ctx)
	gatewaysCh := r.ProviderResources.Gateways.Subscribe(ctx)
	httpRoutesCh := r.ProviderResources.HTTPRoutes.Subscribe(ctx)
	servicesCh := r.ProviderResources.Services.Subscribe(ctx)
	namespacesCh := r.ProviderResources.Namespaces.Subscribe(ctx)

	// Wait until provider resources have been initialized during startup.
	r.ProviderResources.Initialized.Wait()
	for ctx.Err() == nil {
		var in gatewayapi.Resources
		// Receive subscribed resource notifications
		select {
		case <-gatewayClassesCh:
		case <-gatewaysCh:
		case <-httpRoutesCh:
		case <-servicesCh:
		case <-namespacesCh:
		}
		r.Logger.Info("received a notification")
		// Load all resources required for xds ir translation.
		in.Gateways = r.ProviderResources.GetGateways()
		in.HTTPRoutes = r.ProviderResources.GetHTTPRoutes()
		in.Services = r.ProviderResources.GetServices()
		in.Namespaces = r.ProviderResources.GetNamespaces()
		gatewayClasses := r.ProviderResources.GetGatewayClasses()
		// Fetch the first gatewayclass since there should only be 1
		// linked to this controller.
		switch {
		case gatewayClasses == nil:
			// Envoy Gateway startup.
			continue
		case gatewayClasses[0] == nil:
			// No need to translate, publish empty xds IR to trigger deletion.
			r.XdsIR.Store(r.Name(), &ir.Xds{})
		default:
			r.Logger.Info("translating xds ir")
			// Translate and publish xds IR.
			t := &gatewayapi.Translator{
				GatewayClassName: v1beta1.ObjectName(gatewayClasses[0].GetName()),
			}
			// Translate to xds IR.
			result := t.TranslateXds(&in)
			yamlXdsIR, _ := yaml.Marshal(&result.XdsIR)
			r.Logger.Info(string(yamlXdsIR))

			// Publish the xds IR. Use the service name as the key
			// to ensure there is always one element in the map.
			// Also validate the ir before sending it.
			if err := result.XdsIR.Validate(); err != nil {
				r.Logger.Error(err, "unable to validate xds ir, skipped sending it")
			} else {
				r.XdsIR.Store(r.Name(), result.XdsIR)
			}
		}
	}
	r.Logger.Info("shutting down")
}
