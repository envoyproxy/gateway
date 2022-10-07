package runner

import (
	"context"

	"sigs.k8s.io/gateway-api/apis/v1beta1"
	"sigs.k8s.io/yaml"

	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
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
	secretsCh := r.ProviderResources.Secrets.Subscribe(ctx)
	refGrantsCh := r.ProviderResources.ReferenceGrants.Subscribe(ctx)
	httpRoutesCh := r.ProviderResources.HTTPRoutes.Subscribe(ctx)
	tlsRoutesCh := r.ProviderResources.TLSRoutes.Subscribe(ctx)
	servicesCh := r.ProviderResources.Services.Subscribe(ctx)
	namespacesCh := r.ProviderResources.Namespaces.Subscribe(ctx)

	for ctx.Err() == nil {
		var in gatewayapi.Resources
		// Receive subscribed resource notifications
		select {
		case <-gatewayClassesCh:
		case <-gatewaysCh:
		case <-secretsCh:
		case <-refGrantsCh:
		case <-httpRoutesCh:
		case <-tlsRoutesCh:
		case <-servicesCh:
		case <-namespacesCh:
		}
		r.Logger.Info("received a notification")
		// Load all resources required for translation
		in.Gateways = r.ProviderResources.GetGateways()
		in.Secrets = r.ProviderResources.GetSecrets()
		in.ReferenceGrants = r.ProviderResources.GetReferenceGrants()
		in.HTTPRoutes = r.ProviderResources.GetHTTPRoutes()
		in.TLSRoutes = r.ProviderResources.GetTLSRoutes()
		in.Services = r.ProviderResources.GetServices()
		in.Namespaces = r.ProviderResources.GetNamespaces()
		gatewayClasses := r.ProviderResources.GetGatewayClasses()
		// Fetch the first gateway class since there should be only 1
		// gateway class linked to this controller
		switch {
		case gatewayClasses == nil:
			// Envoy Gateway startup.
			continue
		default:
			// Translate and publish IRs.
			t := &gatewayapi.Translator{
				GatewayClassName: v1beta1.ObjectName(gatewayClasses[0].GetName()),
			}
			// Translate to IR
			result := t.Translate(&in)

			yamlInfraIR, _ := yaml.Marshal(&result.InfraIR)
			r.Logger.WithValues("output", "infra-ir").Info(string(yamlInfraIR))

			var curKeys, newKeys []string
			// Get current IR keys
			for key := range r.InfraIR.LoadAll() {
				curKeys = append(curKeys, key)
			}

			// Publish the IRs.
			// Also validate the ir before sending it.
			for key, val := range result.InfraIR {
				if err := val.Validate(); err != nil {
					r.Logger.Error(err, "unable to validate infra ir, skipped sending it")
				} else {
					r.InfraIR.Store(key, val)
					newKeys = append(newKeys, key)
				}
			}

			for key, val := range result.XdsIR {
				if err := val.Validate(); err != nil {
					r.Logger.Error(err, "unable to validate xds ir, skipped sending it")
				} else {
					r.XdsIR.Store(key, val)
				}
			}

			// Delete keys
			// There is a 1:1 mapping between infra and xds IR keys
			delKeys := getIRKeysToDelete(curKeys, newKeys)
			for _, key := range delKeys {
				r.InfraIR.Delete(key)
				r.XdsIR.Delete(key)
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
			for _, tlsRoute := range result.TLSRoutes {
				key := utils.NamespacedName(tlsRoute)
				r.ProviderResources.TLSRouteStatuses.Store(key, tlsRoute)
			}
		}
	}
	r.Logger.Info("shutting down")
}

// getIRKeysToDelete returns the list of IR keys to delete
// based on the difference between the current keys and the
// new keys parameters passed to the function.
func getIRKeysToDelete(curKeys, newKeys []string) []string {
	var delKeys []string
	remaining := make(map[string]bool)

	// Add all current keys to the remaining map
	for _, key := range curKeys {
		remaining[key] = true
	}

	// Delete newKeys from the remaining map
	// to get keys that need to be deleted
	for _, key := range newKeys {
		delete(remaining, key)
	}

	for key := range remaining {
		delKeys = append(delKeys, key)
	}

	return delKeys
}
