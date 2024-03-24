// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"context"
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
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

			// Get all status keys from watchable and save them in this StatusesToDelete structure.
			// Iterating through the controller resources, any valid keys will be removed from statusesToDelete.
			// Remaining keys will be deleted from watchable before we exit this function.
			statusesToDelete := r.getAllStatuses()

			for _, resources := range *val {
				// Translate and publish IRs.
				t := &gatewayapi.Translator{
					GatewayControllerName:   r.Server.EnvoyGateway.Gateway.ControllerName,
					GatewayClassName:        v1.ObjectName(resources.GatewayClass.Name),
					GlobalRateLimitEnabled:  r.EnvoyGateway.RateLimit != nil,
					EnvoyPatchPolicyEnabled: r.EnvoyGateway.ExtensionAPIs != nil && r.EnvoyGateway.ExtensionAPIs.EnableEnvoyPatchPolicy,
					Namespace:               r.Namespace,
					MergeGateways:           gatewayapi.IsMergeGatewaysEnabled(resources),
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
					delete(statusesToDelete.GatewayStatusKeys, key)
				}
				for _, httpRoute := range result.HTTPRoutes {
					httpRoute := httpRoute
					key := utils.NamespacedName(httpRoute)
					r.ProviderResources.HTTPRouteStatuses.Store(key, &httpRoute.Status)
					delete(statusesToDelete.HTTPRouteStatusKeys, key)
				}
				for _, grpcRoute := range result.GRPCRoutes {
					grpcRoute := grpcRoute
					key := utils.NamespacedName(grpcRoute)
					r.ProviderResources.GRPCRouteStatuses.Store(key, &grpcRoute.Status)
					delete(statusesToDelete.GRPCRouteStatusKeys, key)
				}
				for _, tlsRoute := range result.TLSRoutes {
					tlsRoute := tlsRoute
					key := utils.NamespacedName(tlsRoute)
					r.ProviderResources.TLSRouteStatuses.Store(key, &tlsRoute.Status)
					delete(statusesToDelete.TLSRouteStatusKeys, key)
				}
				for _, tcpRoute := range result.TCPRoutes {
					tcpRoute := tcpRoute
					key := utils.NamespacedName(tcpRoute)
					r.ProviderResources.TCPRouteStatuses.Store(key, &tcpRoute.Status)
					delete(statusesToDelete.TCPRouteStatusKeys, key)
				}
				for _, udpRoute := range result.UDPRoutes {
					udpRoute := udpRoute
					key := utils.NamespacedName(udpRoute)
					r.ProviderResources.UDPRouteStatuses.Store(key, &udpRoute.Status)
					delete(statusesToDelete.UDPRouteStatusKeys, key)
				}

				// Skip updating status for policies with empty status
				// They may have been skipped in this translation because
				// their target is not found (not relevant)

				for _, backendTLSPolicy := range result.BackendTLSPolicies {
					backendTLSPolicy := backendTLSPolicy
					key := utils.NamespacedName(backendTLSPolicy)
					if !(reflect.ValueOf(backendTLSPolicy.Status).IsZero()) {
						r.ProviderResources.BackendTLSPolicyStatuses.Store(key, &backendTLSPolicy.Status)
					}
					delete(statusesToDelete.BackendTLSPolicyStatusKeys, key)
				}

				for _, clientTrafficPolicy := range result.ClientTrafficPolicies {
					clientTrafficPolicy := clientTrafficPolicy
					key := utils.NamespacedName(clientTrafficPolicy)
					if !(reflect.ValueOf(clientTrafficPolicy.Status).IsZero()) {
						r.ProviderResources.ClientTrafficPolicyStatuses.Store(key, &clientTrafficPolicy.Status)
					}
					delete(statusesToDelete.ClientTrafficPolicyStatusKeys, key)
				}
				for _, backendTrafficPolicy := range result.BackendTrafficPolicies {
					backendTrafficPolicy := backendTrafficPolicy
					key := utils.NamespacedName(backendTrafficPolicy)
					if !(reflect.ValueOf(backendTrafficPolicy.Status).IsZero()) {
						r.ProviderResources.BackendTrafficPolicyStatuses.Store(key, &backendTrafficPolicy.Status)
					}
					delete(statusesToDelete.BackendTrafficPolicyStatusKeys, key)
				}
				for _, securityPolicy := range result.SecurityPolicies {
					securityPolicy := securityPolicy
					key := utils.NamespacedName(securityPolicy)
					if !(reflect.ValueOf(securityPolicy.Status).IsZero()) {
						r.ProviderResources.SecurityPolicyStatuses.Store(key, &securityPolicy.Status)
					}
					delete(statusesToDelete.SecurityPolicyStatusKeys, key)
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
			r.deleteStatusKeys(statusesToDelete)
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

type StatusesToDelete struct {
	GatewayStatusKeys          map[types.NamespacedName]bool
	HTTPRouteStatusKeys        map[types.NamespacedName]bool
	GRPCRouteStatusKeys        map[types.NamespacedName]bool
	TLSRouteStatusKeys         map[types.NamespacedName]bool
	TCPRouteStatusKeys         map[types.NamespacedName]bool
	UDPRouteStatusKeys         map[types.NamespacedName]bool
	BackendTLSPolicyStatusKeys map[types.NamespacedName]bool

	ClientTrafficPolicyStatusKeys  map[types.NamespacedName]bool
	BackendTrafficPolicyStatusKeys map[types.NamespacedName]bool
	SecurityPolicyStatusKeys       map[types.NamespacedName]bool
}

func (r *Runner) getAllStatuses() *StatusesToDelete {
	// Maps storing status keys to be deleted
	ds := &StatusesToDelete{
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
	for key := range r.ProviderResources.BackendTLSPolicyStatuses.LoadAll() {
		ds.BackendTLSPolicyStatusKeys[key] = true
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

	return ds
}

func (r *Runner) deleteStatusKeys(ds *StatusesToDelete) {
	for key := range ds.GatewayStatusKeys {
		r.ProviderResources.GatewayStatuses.Delete(key)
		delete(ds.GatewayStatusKeys, key)
	}
	for key := range ds.HTTPRouteStatusKeys {
		r.ProviderResources.HTTPRouteStatuses.Delete(key)
		delete(ds.HTTPRouteStatusKeys, key)
	}
	for key := range ds.GRPCRouteStatusKeys {
		r.ProviderResources.GRPCRouteStatuses.Delete(key)
		delete(ds.GRPCRouteStatusKeys, key)
	}
	for key := range ds.TLSRouteStatusKeys {
		r.ProviderResources.TLSRouteStatuses.Delete(key)
		delete(ds.TLSRouteStatusKeys, key)
	}
	for key := range ds.TCPRouteStatusKeys {
		r.ProviderResources.TCPRouteStatuses.Delete(key)
		delete(ds.TCPRouteStatusKeys, key)
	}
	for key := range ds.UDPRouteStatusKeys {
		r.ProviderResources.UDPRouteStatuses.Delete(key)
		delete(ds.UDPRouteStatusKeys, key)
	}

	for key := range ds.ClientTrafficPolicyStatusKeys {
		r.ProviderResources.ClientTrafficPolicyStatuses.Delete(key)
		delete(ds.ClientTrafficPolicyStatusKeys, key)
	}
	for key := range ds.BackendTrafficPolicyStatusKeys {
		r.ProviderResources.BackendTrafficPolicyStatuses.Delete(key)
		delete(ds.BackendTrafficPolicyStatusKeys, key)
	}
	for key := range ds.SecurityPolicyStatusKeys {
		r.ProviderResources.SecurityPolicyStatuses.Delete(key)
		delete(ds.SecurityPolicyStatusKeys, key)
	}
	for key := range ds.BackendTLSPolicyStatusKeys {
		r.ProviderResources.BackendTLSPolicyStatuses.Delete(key)
		delete(ds.BackendTLSPolicyStatusKeys, key)
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
	for key := range r.ProviderResources.BackendTLSPolicyStatuses.LoadAll() {
		r.ProviderResources.BackendTLSPolicyStatuses.Delete(key)
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
}

// getIRKeysToDelete returns the list of IR keys to delete
// based on the difference between the current keys and the
// new keys parameters passed to the function.
func getIRKeysToDelete(curKeys, newKeys []string) []string {
	curSet := sets.NewString(curKeys...)
	newSet := sets.NewString(newKeys...)

	delSet := curSet.Difference(newSet)

	return delSet.List()
}
