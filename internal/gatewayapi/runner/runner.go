package runner

import (
	"context"

	"sigs.k8s.io/gateway-api/apis/v1beta1"
	"sigs.k8s.io/yaml"

	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/provider/utils"
)

type Config struct {
	config.Server
	ProviderResources *message.ProviderResources
	XdsIR             *message.XdsIR
	InfraIR           *message.InfraIR
}

type Runner struct {
	Config
	xdsIRReady bool
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
	go r.subscribeAndTranslate(ctx)
	r.Logger.Info("started")
	return nil
}

func (r *Runner) subscribeAndTranslate(ctx context.Context) {
	// Subscribe to resources
	gatewayClassesCh := r.ProviderResources.GatewayClasses.Subscribe(ctx)
	gatewaysCh := r.ProviderResources.Gateways.Subscribe(ctx)
	httpRoutesCh := r.ProviderResources.HTTPRoutes.Subscribe(ctx)
	servicesCh := r.ProviderResources.Services.Subscribe(ctx)
	namespacesCh := r.ProviderResources.Namespaces.Subscribe(ctx)

	for ctx.Err() == nil {
		var in gatewayapi.Resources
		// Receive subscribed resource notifications
		select {
		case <-gatewayClassesCh:
			r.waitUntilGCAndGatewaysInitialized()
		case <-gatewaysCh:
			r.waitUntilGCAndGatewaysInitialized()
		case <-httpRoutesCh:
			r.waitUntilAllGAPIInitialized()
		case <-servicesCh:
			r.waitUntilAllGAPIInitialized()
		case <-namespacesCh:
			r.waitUntilGCAndGatewaysInitialized()
		}
		r.Logger.Info("received a notification")
		// Load all resources required for translation
		in.Gateways = r.ProviderResources.GetGateways()
		in.HTTPRoutes = r.ProviderResources.GetHTTPRoutes()
		in.Services = r.ProviderResources.GetServices()
		in.Namespaces = r.ProviderResources.GetNamespaces()
		gatewayClasses := r.ProviderResources.GetGatewayClasses()
		// Fetch the first gateway class since there should be only 1
		// gateway class linked to this controller
		switch {
		case gatewayClasses == nil:
			// Envoy Gateway startup.
			continue
		case gatewayClasses[0] == nil:
			// No need to translate, publish empty IRs to trigger a delete operation.
			r.XdsIR.Store(r.Name(), &ir.Xds{})
			// A nil ProxyInfra tells the Infra Manager to delete the managed proxy infra.
			r.InfraIR.Store(r.Name(), &ir.Infra{Proxy: nil})
		default:
			// Translate and publish IRs.
			t := &gatewayapi.Translator{
				GatewayClassName: v1beta1.ObjectName(gatewayClasses[0].GetName()),
			}
			// Translate to IR
			result := t.Translate(&in)

			yamlXdsIR, _ := yaml.Marshal(&result.XdsIR)
			r.Logger.WithValues("output", "xds-ir").Info(string(yamlXdsIR))
			yamlInfraIR, _ := yaml.Marshal(&result.InfraIR)
			r.Logger.WithValues("output", "infra-ir").Info(string(yamlInfraIR))

			// Publish the IRs.
			// Also validate the ir before sending it.
			for key, val := range result.InfraIR {
				if err := val.Validate(); err != nil {
					r.Logger.Error(err, "unable to validate infra ir, skipped sending it")
				} else {
					r.InfraIR.Store(key, val)
				}
			}
			// Wait until all HTTPRoutes have been reconciled , else the translation
			// result will be incomplete, and might cause churn in the data plane.
			if r.xdsIRReady {
				for key, val := range result.XdsIR {
					if err := val.Validate(); err != nil {
						r.Logger.Error(err, "unable to validate xds ir, skipped sending it")
					} else {
						r.XdsIR.Store(key, val)
					}
				}
			}

			// Update Status
			for _, gateway := range result.Gateways {
				key := utils.NamespacedName(gateway)
				r.ProviderResources.GatewayStatuses.Store(key, gateway)
			}
			for _, httpRoute := range result.HTTPRoutes {
				key := utils.NamespacedName(httpRoute)
				r.ProviderResources.HTTPRouteStatuses.Store(key, httpRoute)
			}
		}
	}
	r.Logger.Info("shutting down")
}

// waitUntilGCAndGatewaysInitialized waits until gateway classes and
// gateways have been initialized during startup
func (r *Runner) waitUntilGCAndGatewaysInitialized() {
	r.ProviderResources.GatewayClassesInitialized.Wait()
	r.ProviderResources.GatewaysInitialized.Wait()
}

// waitUntilAllGAPIInitialized waits until gateway classes,
// gateways and httproutes have been initialized during startup
func (r *Runner) waitUntilAllGAPIInitialized() {
	r.waitUntilGCAndGatewaysInitialized()
	r.ProviderResources.HTTPRoutesInitialized.Wait()
	// Now that the httproute resources have been initialized,
	// allow the runner to publish the translated xdsIR.
	r.xdsIRReady = true
}
