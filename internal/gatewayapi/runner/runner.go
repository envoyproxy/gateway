// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	v1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	extension "github.com/envoyproxy/gateway/internal/extension/types"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/utils"
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
func (r *Runner) Start(ctx context.Context) (err error) {
	r.Logger = r.Logger.WithName(r.Name()).WithValues("runner", r.Name())
	go r.subscribeAndTranslate(ctx)
	r.Logger.Info("started")
	return
}

func (r *Runner) subscribeAndTranslate(ctx context.Context) {
	message.HandleSubscription(message.Metadata{Runner: string(v1alpha1.LogComponentGatewayAPIRunner), Message: "provider-resources"}, r.ProviderResources.GatewayAPIResources.Subscribe(ctx),
		func(update message.Update[string, *gatewayapi.ControllerResources], errChan chan error) {
			r.Logger.Info("received an update")
			val := update.Value
			// There is only 1 key which is the controller name
			// so when a delete is triggered, delete all IR keys
			if update.Delete || val == nil {
				r.deleteAllIRKeys()
				r.deleteAllStatusKeys()
				return
			}

			// IR keys for watchable
			var curIRKeys, newIRKeys []string

			// Get current IR keys
			for key := range r.InfraIR.LoadAll() {
				curIRKeys = append(curIRKeys, key)
			}

			// Get the current DeletableStatus which manages status keys to be deleted
			deletableStatus := r.getDeletableStatus()

			for _, resources := range *val {
				// Translate and publish IRs.
				t := &gatewayapi.Translator{
					GatewayControllerName:   r.Server.EnvoyGateway.Gateway.ControllerName,
					GatewayClassName:        v1.ObjectName(resources.GatewayClass.Name),
					GlobalRateLimitEnabled:  r.EnvoyGateway.RateLimit != nil,
					EnvoyPatchPolicyEnabled: r.EnvoyGateway.ExtensionAPIs != nil && r.EnvoyGateway.ExtensionAPIs.EnableEnvoyPatchPolicy,
					Namespace:               r.Namespace,
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
				result := t.Translate(resources)

				// Publish the IRs.
				// Also validate the ir before sending it.
				for key, val := range result.InfraIR {
					r.Logger.WithValues("infra-ir", key).Info(val.YAMLString())
					if err := val.Validate(); err != nil {
						r.Logger.Error(err, "unable to validate infra ir, skipped sending it")
						errChan <- err
					} else {
						r.InfraIR.Store(key, val)
						newIRKeys = append(newIRKeys, key)
					}
				}

				for key, val := range result.XdsIR {
					r.Logger.WithValues("xds-ir", key).Info(val.YAMLString())
					if err := val.Validate(); err != nil {
						r.Logger.Error(err, "unable to validate xds ir, skipped sending it")
						errChan <- err
					} else {
						r.XdsIR.Store(key, val)
					}
				}

				// Update Status
				for _, gateway := range result.Gateways {
					gateway := gateway
					key := utils.NamespacedName(gateway)
					r.ProviderResources.GatewayStatuses.Store(key, &gateway.Status)
					delete(deletableStatus.GatewayStatusKeys, key)
				}
				for _, httpRoute := range result.HTTPRoutes {
					httpRoute := httpRoute
					key := utils.NamespacedName(httpRoute)
					r.ProviderResources.HTTPRouteStatuses.Store(key, &httpRoute.Status)
					delete(deletableStatus.HTTPRouteStatusKeys, key)
				}
				for _, grpcRoute := range result.GRPCRoutes {
					grpcRoute := grpcRoute
					key := utils.NamespacedName(grpcRoute)
					r.ProviderResources.GRPCRouteStatuses.Store(key, &grpcRoute.Status)
					delete(deletableStatus.GRPCRouteStatusKeys, key)
				}

				for _, tlsRoute := range result.TLSRoutes {
					tlsRoute := tlsRoute
					key := utils.NamespacedName(tlsRoute)
					r.ProviderResources.TLSRouteStatuses.Store(key, &tlsRoute.Status)
					delete(deletableStatus.TLSRouteStatusKeys, key)
				}
				for _, tcpRoute := range result.TCPRoutes {
					tcpRoute := tcpRoute
					key := utils.NamespacedName(tcpRoute)
					r.ProviderResources.TCPRouteStatuses.Store(key, &tcpRoute.Status)
					delete(deletableStatus.TCPRouteStatusKeys, key)
				}
				for _, udpRoute := range result.UDPRoutes {
					udpRoute := udpRoute
					key := utils.NamespacedName(udpRoute)
					r.ProviderResources.UDPRouteStatuses.Store(key, &udpRoute.Status)
					delete(deletableStatus.UDPRouteStatusKeys, key)
				}
				for _, clientTrafficPolicy := range result.ClientTrafficPolicies {
					clientTrafficPolicy := clientTrafficPolicy
					key := utils.NamespacedName(clientTrafficPolicy)
					r.ProviderResources.ClientTrafficPolicyStatuses.Store(key, &clientTrafficPolicy.Status)
					delete(deletableStatus.ClientTrafficPolicyStatusKeys, key)
				}
				for _, backendTrafficPolicy := range result.BackendTrafficPolicies {
					backendTrafficPolicy := backendTrafficPolicy
					key := utils.NamespacedName(backendTrafficPolicy)
					r.ProviderResources.BackendTrafficPolicyStatuses.Store(key, &backendTrafficPolicy.Status)
					delete(deletableStatus.BackendTrafficPolicyStatusKeys, key)
				}
				for _, securityPolicy := range result.SecurityPolicies {
					securityPolicy := securityPolicy
					key := utils.NamespacedName(securityPolicy)
					r.ProviderResources.SecurityPolicyStatuses.Store(key, &securityPolicy.Status)
					delete(deletableStatus.SecurityPolicyStatusKeys, key)
				}
				for _, backendTLSPolicy := range result.BackendTLSPolicies {
					backendTLSPolicy := backendTLSPolicy
					key := utils.NamespacedName(backendTLSPolicy)
					r.ProviderResources.BackendTLSPolicyStatuses.Store(key, &backendTLSPolicy.Status)
					delete(deletableStatus.BackendTLSPolicyStatusKeys, key)
				}
			}

			// Delete IR keys
			// There is a 1:1 mapping between infra and xds IR keys
			delKeys := getIRKeysToDelete(curIRKeys, newIRKeys)
			for _, key := range delKeys {
				r.InfraIR.Delete(key)
				r.XdsIR.Delete(key)
			}

			// Delete status keys
			r.deleteStatusKeys(deletableStatus)
		},
	)
	r.Logger.Info("shutting down")
}

// deleteAllIRKeys deletes all XdsIR and InfraIR
func (r *Runner) deleteAllIRKeys() {
	for key := range r.InfraIR.LoadAll() {
		r.InfraIR.Delete(key)
		r.XdsIR.Delete(key)
	}
}

type DeletableStatus struct {
	GatewayStatusKeys   map[types.NamespacedName]bool
	HTTPRouteStatusKeys map[types.NamespacedName]bool
	GRPCRouteStatusKeys map[types.NamespacedName]bool
	TLSRouteStatusKeys  map[types.NamespacedName]bool
	TCPRouteStatusKeys  map[types.NamespacedName]bool
	UDPRouteStatusKeys  map[types.NamespacedName]bool

	ClientTrafficPolicyStatusKeys  map[types.NamespacedName]bool
	BackendTrafficPolicyStatusKeys map[types.NamespacedName]bool
	SecurityPolicyStatusKeys       map[types.NamespacedName]bool
	BackendTLSPolicyStatusKeys     map[types.NamespacedName]bool
}

func (r *Runner) getDeletableStatus() *DeletableStatus {
	// Maps storing status keys to be deleted
	ds := &DeletableStatus{
		GatewayStatusKeys:   make(map[types.NamespacedName]bool),
		HTTPRouteStatusKeys: make(map[types.NamespacedName]bool),
		GRPCRouteStatusKeys: make(map[types.NamespacedName]bool),
		TLSRouteStatusKeys:  make(map[types.NamespacedName]bool),
		TCPRouteStatusKeys:  make(map[types.NamespacedName]bool),
		UDPRouteStatusKeys:  make(map[types.NamespacedName]bool),

		ClientTrafficPolicyStatusKeys:  make(map[types.NamespacedName]bool),
		BackendTrafficPolicyStatusKeys: make(map[types.NamespacedName]bool),
		SecurityPolicyStatusKeys:       make(map[types.NamespacedName]bool),
		BackendTLSPolicyStatusKeys:     make(map[types.NamespacedName]bool),
	}

	// Get current status keys
	for key := range r.ProviderResources.GatewayStatuses.LoadAll() {
		ds.GatewayStatusKeys[key] = true
	}
	for key := range r.ProviderResources.HTTPRouteStatuses.LoadAll() {
		ds.HTTPRouteStatusKeys[key] = true
	}
	for key := range r.ProviderResources.GRPCRouteStatuses.LoadAll() {
		ds.GRPCRouteStatusKeys[key] = true
	}
	for key := range r.ProviderResources.TLSRouteStatuses.LoadAll() {
		ds.TLSRouteStatusKeys[key] = true
	}
	for key := range r.ProviderResources.TCPRouteStatuses.LoadAll() {
		ds.TCPRouteStatusKeys[key] = true
	}
	for key := range r.ProviderResources.UDPRouteStatuses.LoadAll() {
		ds.UDPRouteStatusKeys[key] = true
	}

	for key := range r.ProviderResources.ClientTrafficPolicyStatuses.LoadAll() {
		ds.ClientTrafficPolicyStatusKeys[key] = true
	}
	for key := range r.ProviderResources.BackendTrafficPolicyStatuses.LoadAll() {
		ds.BackendTrafficPolicyStatusKeys[key] = true
	}
	for key := range r.ProviderResources.SecurityPolicyStatuses.LoadAll() {
		ds.SecurityPolicyStatusKeys[key] = true
	}
	for key := range r.ProviderResources.BackendTLSPolicyStatuses.LoadAll() {
		ds.BackendTLSPolicyStatusKeys[key] = true
	}

	return ds
}

func (r *Runner) deleteStatusKeys(ds *DeletableStatus) {
	for key := range ds.GatewayStatusKeys {
		r.ProviderResources.GatewayStatuses.Delete(key)
	}
	for key := range ds.HTTPRouteStatusKeys {
		r.ProviderResources.HTTPRouteStatuses.Delete(key)
	}
	for key := range ds.GRPCRouteStatusKeys {
		r.ProviderResources.GRPCRouteStatuses.Delete(key)
	}
	for key := range ds.TLSRouteStatusKeys {
		r.ProviderResources.TLSRouteStatuses.Delete(key)
	}
	for key := range ds.TCPRouteStatusKeys {
		r.ProviderResources.TCPRouteStatuses.Delete(key)
	}
	for key := range ds.UDPRouteStatusKeys {
		r.ProviderResources.UDPRouteStatuses.Delete(key)
	}

	for key := range ds.ClientTrafficPolicyStatusKeys {
		r.ProviderResources.ClientTrafficPolicyStatuses.Delete(key)
	}
	for key := range ds.BackendTrafficPolicyStatusKeys {
		r.ProviderResources.BackendTrafficPolicyStatuses.Delete(key)
	}
	for key := range ds.SecurityPolicyStatusKeys {
		r.ProviderResources.SecurityPolicyStatuses.Delete(key)
	}
	for key := range ds.BackendTLSPolicyStatusKeys {
		r.ProviderResources.BackendTLSPolicyStatuses.Delete(key)
	}
}

// deleteAllStatusKeys deletes all status keys stored by the subscriber.
func (r *Runner) deleteAllStatusKeys() {
	// Fields of GatewayAPIStatuses
	for key := range r.ProviderResources.GatewayStatuses.LoadAll() {
		r.ProviderResources.GatewayStatuses.Delete(key)
	}
	for key := range r.ProviderResources.HTTPRouteStatuses.LoadAll() {
		r.ProviderResources.HTTPRouteStatuses.Delete(key)
	}
	for key := range r.ProviderResources.GRPCRouteStatuses.LoadAll() {
		r.ProviderResources.GRPCRouteStatuses.Delete(key)
	}
	for key := range r.ProviderResources.TLSRouteStatuses.LoadAll() {
		r.ProviderResources.TLSRouteStatuses.Delete(key)
	}
	for key := range r.ProviderResources.TCPRouteStatuses.LoadAll() {
		r.ProviderResources.TCPRouteStatuses.Delete(key)
	}
	for key := range r.ProviderResources.UDPRouteStatuses.LoadAll() {
		r.ProviderResources.UDPRouteStatuses.Delete(key)
	}

	// Fields of PolicyStatuses
	for key := range r.ProviderResources.ClientTrafficPolicyStatuses.LoadAll() {
		r.ProviderResources.ClientTrafficPolicyStatuses.Delete(key)
	}
	for key := range r.ProviderResources.BackendTrafficPolicyStatuses.LoadAll() {
		r.ProviderResources.BackendTrafficPolicyStatuses.Delete(key)
	}
	for key := range r.ProviderResources.SecurityPolicyStatuses.LoadAll() {
		r.ProviderResources.SecurityPolicyStatuses.Delete(key)
	}
	for key := range r.ProviderResources.BackendTLSPolicyStatuses.LoadAll() {
		r.ProviderResources.BackendTLSPolicyStatuses.Delete(key)
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
