// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
	"sigs.k8s.io/yaml"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	extensionregistry "github.com/envoyproxy/gateway/internal/extension/registry"
	extension "github.com/envoyproxy/gateway/internal/extension/types"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/provider/utils"
	"github.com/envoyproxy/gateway/internal/runner"
)

func Register(resources Resources, globalConfig config.Server) {
	runner.Manager().Register(New(resources, globalConfig), runner.RootParentRunner)
}

func New(resources Resources, globalConfig config.Server) *gwapiRunner {
	return &gwapiRunner{runner.New(string(v1alpha1.LogComponentGatewayAPIRunner), resources, globalConfig)}
}

type Resources struct {
	ProviderResources *message.ProviderResources
	XdsIR             *message.XdsIR
	InfraIR           *message.InfraIR
	ExtensionManager  extension.Manager
}

type gwapiRunner struct {
	*runner.GenericRunner[Resources]
}

func (r *gwapiRunner) SubscribeAndTranslate(ctx context.Context) {
	message.HandleSubscription(r.Resources.ProviderResources.GatewayAPIResources.Subscribe(ctx),
		func(update message.Update[string, *gatewayapi.Resources]) {
			val := update.Value

			if update.Delete || val == nil {
				return
			}

			// Translate and publish IRs.
			t := &gatewayapi.Translator{
				GatewayControllerName:  r.Server.EnvoyGateway.Gateway.ControllerName,
				GatewayClassName:       v1beta1.ObjectName(update.Key),
				GlobalRateLimitEnabled: r.EnvoyGateway.RateLimit != nil,
			}

			// If an extension is loaded, pass its supported groups/kinds to the translator
			if r.EnvoyGateway.Extension != nil {
				var extGKs []schema.GroupKind
				for _, gvk := range r.EnvoyGateway.Extension.Resources {
					extGKs = append(extGKs, schema.GroupKind{Group: gvk.Group, Kind: gvk.Kind})
				}
				t.ExtensionGroupKinds = extGKs
			}
			// Translate to IR
			result := t.Translate(val)

			yamlXdsIR, _ := yaml.Marshal(&result.XdsIR)
			r.Logger.WithValues("output", "xds-ir").Info(string(yamlXdsIR))
			yamlInfraIR, _ := yaml.Marshal(&result.InfraIR)
			r.Logger.WithValues("output", "infra-ir").Info(string(yamlInfraIR))

			var curKeys, newKeys []string
			// Get current IR keys
			for key := range r.Resources.InfraIR.LoadAll() {
				curKeys = append(curKeys, key)
			}

			// Publish the IRs.
			// Also validate the ir before sending it.
			for key, val := range result.InfraIR {
				if err := val.Validate(); err != nil {
					r.Logger.Error(err, "unable to validate infra ir, skipped sending it")
				} else {
					r.Resources.InfraIR.Store(key, val)
					newKeys = append(newKeys, key)
				}
			}

			for key, val := range result.XdsIR {
				if err := val.Validate(); err != nil {
					r.Logger.Error(err, "unable to validate xds ir, skipped sending it")
				} else {
					r.Resources.XdsIR.Store(key, val)
				}
			}

			// Delete keys
			// There is a 1:1 mapping between infra and xds IR keys
			delKeys := getIRKeysToDelete(curKeys, newKeys)
			for _, key := range delKeys {
				r.Resources.InfraIR.Delete(key)
				r.Resources.XdsIR.Delete(key)
			}

			// Update Status
			for _, gateway := range result.Gateways {
				key := utils.NamespacedName(gateway)
				r.Resources.ProviderResources.GatewayStatuses.Store(key, &gateway.Status)
			}
			for _, httpRoute := range result.HTTPRoutes {
				key := utils.NamespacedName(httpRoute)
				r.Resources.ProviderResources.HTTPRouteStatuses.Store(key, &httpRoute.Status)
			}
			for _, grpcRoute := range result.GRPCRoutes {
				key := utils.NamespacedName(grpcRoute)
				r.Resources.ProviderResources.GRPCRouteStatuses.Store(key, &grpcRoute.Status)
			}

			for _, tlsRoute := range result.TLSRoutes {
				key := utils.NamespacedName(tlsRoute)
				r.Resources.ProviderResources.TLSRouteStatuses.Store(key, &tlsRoute.Status)
			}
			for _, tcpRoute := range result.TCPRoutes {
				key := utils.NamespacedName(tcpRoute)
				r.Resources.ProviderResources.TCPRouteStatuses.Store(key, &tcpRoute.Status)
			}
			for _, udpRoute := range result.UDPRoutes {
				key := utils.NamespacedName(udpRoute)
				r.Resources.ProviderResources.UDPRouteStatuses.Store(key, &udpRoute.Status)
			}
		},
	)
	r.Logger.Info("shutting down")
}

// Start starts the gateway-api translator runner
func (r *gwapiRunner) Start(ctx context.Context) error {
	r.Init(ctx)
	go r.SubscribeAndTranslate(ctx)

	r.Logger.Info("started")
	return nil
}

// Start starts the gateway-api translator runner
func (r *gwapiRunner) ShutDown(ctx context.Context) {
	r.Resources.XdsIR.Close()
	r.Resources.InfraIR.Close()
	r.Resources.ProviderResources.Close()

	// Close connections to extension services
	if mgr, ok := r.Resources.ExtensionManager.(*extensionregistry.Manager); ok {
		mgr.CleanupHookConns()
	}

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
