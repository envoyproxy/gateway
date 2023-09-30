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
	extension "github.com/envoyproxy/gateway/internal/extension/types"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/provider/utils"
)

type Config struct {
	config.Server
	ProviderResources *message.ProviderResources
	XdsIR             *message.XdsIR
	InfraIR           *message.InfraIR
	ExtensionManager  extension.Manager
}

type Runner struct {
	Config
}

func New(cfg *Config) *Runner {
	return &Runner{Config: *cfg}
}

func (r *Runner) Name() string {
	return string(v1alpha1.LogComponentGatewayAPIRunner)
}

// Start starts the gateway-api translator runner
func (r *Runner) Start(ctx context.Context) error {
	r.Logger = r.Logger.WithName(r.Name()).WithValues("runner", r.Name())
	go r.subscribeAndTranslate(ctx)
	r.Logger.Info("started")
	return nil
}

func (r *Runner) subscribeAndTranslate(ctx context.Context) {
	message.HandleSubscription(r.ProviderResources.GatewayAPIResources.Subscribe(ctx),
		func(update message.Update[string, *gatewayapi.Resources]) {
			r.Logger.Info("received an update")

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
			if r.EnvoyGateway.ExtensionManager != nil {
				var extGKs []schema.GroupKind
				for _, gvk := range r.EnvoyGateway.ExtensionManager.Resources {
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
				gateway := gateway
				key := utils.NamespacedName(gateway)
				r.ProviderResources.GatewayStatuses.Store(key, &gateway.Status)
			}
			for _, httpRoute := range result.HTTPRoutes {
				httpRoute := httpRoute
				key := utils.NamespacedName(httpRoute)
				r.ProviderResources.HTTPRouteStatuses.Store(key, &httpRoute.Status)
			}
			for _, grpcRoute := range result.GRPCRoutes {
				grpcRoute := grpcRoute
				key := utils.NamespacedName(grpcRoute)
				r.ProviderResources.GRPCRouteStatuses.Store(key, &grpcRoute.Status)
			}

			for _, tlsRoute := range result.TLSRoutes {
				tlsRoute := tlsRoute
				key := utils.NamespacedName(tlsRoute)
				r.ProviderResources.TLSRouteStatuses.Store(key, &tlsRoute.Status)
			}
			for _, tcpRoute := range result.TCPRoutes {
				tcpRoute := tcpRoute
				key := utils.NamespacedName(tcpRoute)
				r.ProviderResources.TCPRouteStatuses.Store(key, &tcpRoute.Status)
			}
			for _, udpRoute := range result.UDPRoutes {
				udpRoute := udpRoute
				key := utils.NamespacedName(udpRoute)
				r.ProviderResources.UDPRouteStatuses.Store(key, &udpRoute.Status)
			}
			for _, clientTrafficPolicy := range result.ClientTrafficPolicies {
				clientTrafficPolicy := clientTrafficPolicy
				key := utils.NamespacedName(clientTrafficPolicy)
				r.ProviderResources.ClientTrafficPolicyStatuses.Store(key, &clientTrafficPolicy.Status)
			}

		},
	)
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
