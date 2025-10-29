// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"fmt"
	"reflect"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwapiv1a3 "sigs.k8s.io/gateway-api/apis/v1alpha3"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/utils"
)

// updateStatusFromSubscriptions writes gateway API object status updates to the Kubernetes API server.
func (r *gatewayAPIReconciler) updateStatusFromSubscriptions(ctx context.Context, extensionManagerEnabled bool) {
	// GatewayClass object status updater
	go func() {
		message.HandleSubscription(
			message.Metadata{Runner: string(egv1a1.LogComponentProviderRunner), Message: message.GatewayClassStatusMessageName},
			r.subscriptions.gatewayClassStatuses,
			func(update message.Update[types.NamespacedName, *gwapiv1.GatewayClassStatus], errChan chan error) {
				// skip delete updates.
				if update.Delete {
					return
				}

				r.statusUpdater.Send(Update{
					NamespacedName: update.Key,
					Resource:       new(gwapiv1.GatewayClass),
					Mutator: MutatorFunc(func(obj client.Object) client.Object {
						gc, ok := obj.(*gwapiv1.GatewayClass)
						if !ok {
							panic(fmt.Sprintf("unsupported object type %T", obj))
						}
						gcCopy := gc.DeepCopy()
						gcCopy.Status = *update.Value
						return gcCopy
					}),
				})
			},
		)
		r.log.Info("gatewayclass status subscriber shutting down")
	}()

	// Gateway object status updater
	go func() {
		message.HandleSubscription(
			message.Metadata{Runner: string(egv1a1.LogComponentProviderRunner), Message: message.GatewayStatusMessageName},
			r.subscriptions.gatewayStatuses,
			func(update message.Update[types.NamespacedName, *gwapiv1.GatewayStatus], errChan chan error) {
				// skip delete updates.
				if update.Delete {
					return
				}
				// Get gateway object
				gtw := new(gwapiv1.Gateway)
				if err := r.client.Get(ctx, update.Key, gtw); err != nil {
					r.log.Error(err, "gateway not found", "namespace", gtw.Namespace, "name", gtw.Name)
					errChan <- err
					return
				}
				// Set the updated Status and call the status update
				gtw.Status = *update.Value
				r.updateStatusForGateway(ctx, gtw)
			},
		)
		r.log.Info("gateway status subscriber shutting down")
	}()

	// HTTPRoute object status updater
	go func() {
		message.HandleSubscription(
			message.Metadata{Runner: string(egv1a1.LogComponentProviderRunner), Message: message.HTTPRouteStatusMessageName},
			r.subscriptions.httpRouteStatuses,
			func(update message.Update[types.NamespacedName, *gwapiv1.HTTPRouteStatus], errChan chan error) {
				// skip delete updates.
				if update.Delete {
					return
				}
				key := update.Key
				val := update.Value
				r.statusUpdater.Send(Update{
					NamespacedName: key,
					Resource:       new(gwapiv1.HTTPRoute),
					Mutator: MutatorFunc(func(obj client.Object) client.Object {
						h, ok := obj.(*gwapiv1.HTTPRoute)
						if !ok {
							err := fmt.Errorf("unsupported object type %T", obj)
							errChan <- err
							panic(err)
						}
						hCopy := &gwapiv1.HTTPRoute{
							TypeMeta:   h.TypeMeta,
							ObjectMeta: h.ObjectMeta,
							Spec:       h.Spec,
							Status: gwapiv1.HTTPRouteStatus{
								RouteStatus: gwapiv1.RouteStatus{
									Parents: mergeStatus(h.Namespace, r.envoyGateway.Gateway.ControllerName, h.Status.Parents, val.Parents),
								},
							},
						}
						return hCopy
					}),
				})
			},
		)
		r.log.Info("httpRoute status subscriber shutting down")
	}()

	// GRPCRoute object status updater
	go func() {
		message.HandleSubscription(message.Metadata{Runner: string(egv1a1.LogComponentProviderRunner), Message: message.GRPCRouteStatusMessageName}, r.subscriptions.grpcRouteStatuses,
			func(update message.Update[types.NamespacedName, *gwapiv1.GRPCRouteStatus], errChan chan error) {
				// skip delete updates.
				if update.Delete {
					return
				}
				key := update.Key
				val := update.Value
				r.statusUpdater.Send(Update{
					NamespacedName: key,
					Resource:       new(gwapiv1.GRPCRoute),
					Mutator: MutatorFunc(func(obj client.Object) client.Object {
						g, ok := obj.(*gwapiv1.GRPCRoute)
						if !ok {
							err := fmt.Errorf("unsupported object type %T", obj)
							errChan <- err
							panic(err)
						}
						gCopy := &gwapiv1.GRPCRoute{
							TypeMeta:   g.TypeMeta,
							ObjectMeta: g.ObjectMeta,
							Spec:       g.Spec,
							Status: gwapiv1.GRPCRouteStatus{
								RouteStatus: gwapiv1.RouteStatus{
									Parents: mergeStatus(g.Namespace, r.envoyGateway.Gateway.ControllerName, g.Status.Parents, val.Parents),
								},
							},
						}
						return gCopy
					}),
				})
			},
		)
		r.log.Info("grpcRoute status subscriber shutting down")
	}()

	// TLSRoute object status updater
	go func() {
		message.HandleSubscription(
			message.Metadata{Runner: string(egv1a1.LogComponentProviderRunner), Message: message.TLSRouteStatusMessageName},
			r.subscriptions.tlsRouteStatuses,
			func(update message.Update[types.NamespacedName, *gwapiv1a2.TLSRouteStatus], errChan chan error) {
				// skip delete updates.
				if update.Delete {
					return
				}
				key := update.Key
				val := update.Value
				r.statusUpdater.Send(Update{
					NamespacedName: key,
					Resource:       new(gwapiv1a3.TLSRoute),
					Mutator: MutatorFunc(func(obj client.Object) client.Object {
						t, ok := obj.(*gwapiv1a3.TLSRoute)
						if !ok {
							err := fmt.Errorf("unsupported object type %T", obj)
							errChan <- err
							panic(err)
						}
						tCopy := &gwapiv1a3.TLSRoute{
							TypeMeta:   t.TypeMeta,
							ObjectMeta: t.ObjectMeta,
							Spec:       t.Spec,
							Status: gwapiv1a2.TLSRouteStatus{
								RouteStatus: gwapiv1.RouteStatus{
									Parents: mergeStatus(t.Namespace, r.envoyGateway.Gateway.ControllerName, t.Status.Parents, val.Parents),
								},
							},
						}
						return tCopy
					}),
				})
			},
		)
		r.log.Info("tlsRoute status subscriber shutting down")
	}()

	// TCPRoute object status updater
	go func() {
		message.HandleSubscription(
			message.Metadata{Runner: string(egv1a1.LogComponentProviderRunner), Message: message.TCPRouteStatusMessageName},
			r.subscriptions.tcpRouteStatuses,
			func(update message.Update[types.NamespacedName, *gwapiv1a2.TCPRouteStatus], errChan chan error) {
				// skip delete updates.
				if update.Delete {
					return
				}
				key := update.Key
				val := update.Value
				r.statusUpdater.Send(Update{
					NamespacedName: key,
					Resource:       new(gwapiv1a2.TCPRoute),
					Mutator: MutatorFunc(func(obj client.Object) client.Object {
						t, ok := obj.(*gwapiv1a2.TCPRoute)
						if !ok {
							err := fmt.Errorf("unsupported object type %T", obj)
							errChan <- err
							panic(err)
						}
						tCopy := &gwapiv1a2.TCPRoute{
							TypeMeta:   t.TypeMeta,
							ObjectMeta: t.ObjectMeta,
							Spec:       t.Spec,
							Status: gwapiv1a2.TCPRouteStatus{
								RouteStatus: gwapiv1.RouteStatus{
									Parents: mergeStatus(t.Namespace, r.envoyGateway.Gateway.ControllerName, t.Status.Parents, val.Parents),
								},
							},
						}
						return tCopy
					}),
				})
			},
		)
		r.log.Info("tcpRoute status subscriber shutting down")
	}()

	// UDPRoute object status updater
	go func() {
		message.HandleSubscription(
			message.Metadata{Runner: string(egv1a1.LogComponentProviderRunner), Message: message.UDPRouteStatusMessageName},
			r.subscriptions.udpRouteStatuses,
			func(update message.Update[types.NamespacedName, *gwapiv1a2.UDPRouteStatus], errChan chan error) {
				// skip delete updates.
				if update.Delete {
					return
				}
				key := update.Key
				val := update.Value
				r.statusUpdater.Send(Update{
					NamespacedName: key,
					Resource:       new(gwapiv1a2.UDPRoute),
					Mutator: MutatorFunc(func(obj client.Object) client.Object {
						u, ok := obj.(*gwapiv1a2.UDPRoute)
						if !ok {
							err := fmt.Errorf("unsupported object type %T", obj)
							errChan <- err
							panic(err)
						}
						uCopy := &gwapiv1a2.UDPRoute{
							TypeMeta:   u.TypeMeta,
							ObjectMeta: u.ObjectMeta,
							Spec:       u.Spec,
							Status: gwapiv1a2.UDPRouteStatus{
								RouteStatus: gwapiv1.RouteStatus{
									Parents: mergeStatus(u.Namespace, r.envoyGateway.Gateway.ControllerName, u.Status.Parents, val.Parents),
								},
							},
						}
						return uCopy
					}),
				})
			},
		)
		r.log.Info("udpRoute status subscriber shutting down")
	}()

	// EnvoyPatchPolicy object status updater
	go func() {
		message.HandleSubscription(
			message.Metadata{Runner: string(egv1a1.LogComponentProviderRunner), Message: message.EnvoyPatchPolicyStatusMessageName},
			r.subscriptions.envoyPatchPolicyStatuses,
			func(update message.Update[types.NamespacedName, *gwapiv1.PolicyStatus], errChan chan error) {
				// skip delete updates.
				if update.Delete {
					return
				}
				key := update.Key
				val := update.Value
				r.statusUpdater.Send(Update{
					NamespacedName: key,
					Resource:       new(egv1a1.EnvoyPatchPolicy),
					Mutator: MutatorFunc(func(obj client.Object) client.Object {
						t, ok := obj.(*egv1a1.EnvoyPatchPolicy)
						if !ok {
							err := fmt.Errorf("unsupported object type %T", obj)
							errChan <- err
							panic(err)
						}
						tCopy := &egv1a1.EnvoyPatchPolicy{
							TypeMeta:   t.TypeMeta,
							ObjectMeta: t.ObjectMeta,
							Spec:       t.Spec,
							Status: gwapiv1.PolicyStatus{
								Ancestors: mergeStatus(t.Namespace, r.envoyGateway.Gateway.ControllerName, t.Status.Ancestors, val.Ancestors),
							},
						}
						return tCopy
					}),
				})
			},
		)
		r.log.Info("envoyPatchPolicy status subscriber shutting down")
	}()

	// ClientTrafficPolicy object status updater
	go func() {
		message.HandleSubscription(
			message.Metadata{Runner: string(egv1a1.LogComponentProviderRunner), Message: message.ClientTrafficPolicyStatusMessageName},
			r.subscriptions.clientTrafficPolicyStatuses,
			func(update message.Update[types.NamespacedName, *gwapiv1.PolicyStatus], errChan chan error) {
				// skip delete updates.
				if update.Delete {
					return
				}
				key := update.Key
				val := update.Value
				r.statusUpdater.Send(Update{
					NamespacedName: key,
					Resource:       new(egv1a1.ClientTrafficPolicy),
					Mutator: MutatorFunc(func(obj client.Object) client.Object {
						t, ok := obj.(*egv1a1.ClientTrafficPolicy)
						if !ok {
							err := fmt.Errorf("unsupported object type %T", obj)
							errChan <- err
							panic(err)
						}
						tCopy := &egv1a1.ClientTrafficPolicy{
							TypeMeta:   t.TypeMeta,
							ObjectMeta: t.ObjectMeta,
							Spec:       t.Spec,
							Status: gwapiv1.PolicyStatus{
								Ancestors: mergeStatus(t.Namespace, r.envoyGateway.Gateway.ControllerName, t.Status.Ancestors, val.Ancestors),
							},
						}
						return tCopy
					}),
				})
			},
		)
		r.log.Info("clientTrafficPolicy status subscriber shutting down")
	}()

	// BackendTrafficPolicy object status updater
	go func() {
		message.HandleSubscription(
			message.Metadata{Runner: string(egv1a1.LogComponentProviderRunner), Message: message.BackendTrafficPolicyStatusMessageName},
			r.subscriptions.backendTrafficPolicyStatuses,
			func(update message.Update[types.NamespacedName, *gwapiv1.PolicyStatus], errChan chan error) {
				// skip delete updates.
				if update.Delete {
					return
				}
				key := update.Key
				val := update.Value
				r.statusUpdater.Send(Update{
					NamespacedName: key,
					Resource:       new(egv1a1.BackendTrafficPolicy),
					Mutator: MutatorFunc(func(obj client.Object) client.Object {
						t, ok := obj.(*egv1a1.BackendTrafficPolicy)
						if !ok {
							err := fmt.Errorf("unsupported object type %T", obj)
							errChan <- err
							panic(err)
						}
						tCopy := &egv1a1.BackendTrafficPolicy{
							TypeMeta:   t.TypeMeta,
							ObjectMeta: t.ObjectMeta,
							Spec:       t.Spec,
							Status: gwapiv1.PolicyStatus{
								Ancestors: mergeStatus(t.Namespace, r.envoyGateway.Gateway.ControllerName, t.Status.Ancestors, val.Ancestors),
							},
						}
						return tCopy
					}),
				})
			},
		)
		r.log.Info("backendTrafficPolicy status subscriber shutting down")
	}()

	// SecurityPolicy object status updater
	go func() {
		message.HandleSubscription(
			message.Metadata{Runner: string(egv1a1.LogComponentProviderRunner), Message: message.SecurityPolicyStatusMessageName},
			r.subscriptions.securityPolicyStatuses,
			func(update message.Update[types.NamespacedName, *gwapiv1.PolicyStatus], errChan chan error) {
				// skip delete updates.
				if update.Delete {
					return
				}
				key := update.Key
				val := update.Value
				r.statusUpdater.Send(Update{
					NamespacedName: key,
					Resource:       new(egv1a1.SecurityPolicy),
					Mutator: MutatorFunc(func(obj client.Object) client.Object {
						t, ok := obj.(*egv1a1.SecurityPolicy)
						if !ok {
							err := fmt.Errorf("unsupported object type %T", obj)
							errChan <- err
							panic(err)
						}
						tCopy := &egv1a1.SecurityPolicy{
							TypeMeta:   t.TypeMeta,
							ObjectMeta: t.ObjectMeta,
							Spec:       t.Spec,
							Status: gwapiv1.PolicyStatus{
								Ancestors: mergeStatus(t.Namespace, r.envoyGateway.Gateway.ControllerName, t.Status.Ancestors, val.Ancestors),
							},
						}
						return tCopy
					}),
				})
			},
		)
		r.log.Info("securityPolicy status subscriber shutting down")
	}()

	// BackendTLSPolicy object status updater
	go func() {
		message.HandleSubscription(message.Metadata{Runner: string(egv1a1.LogComponentProviderRunner), Message: message.BackendTLSPolicyStatusMessageName}, r.subscriptions.backendTLSPolicyStatuses,
			func(update message.Update[types.NamespacedName, *gwapiv1.PolicyStatus], errChan chan error) {
				// skip delete updates.
				if update.Delete {
					return
				}
				key := update.Key
				val := update.Value
				r.statusUpdater.Send(Update{
					NamespacedName: key,
					Resource:       new(gwapiv1.BackendTLSPolicy),
					Mutator: MutatorFunc(func(obj client.Object) client.Object {
						t, ok := obj.(*gwapiv1.BackendTLSPolicy)
						if !ok {
							err := fmt.Errorf("unsupported object type %T", obj)
							errChan <- err
							panic(err)
						}
						tCopy := &gwapiv1.BackendTLSPolicy{
							TypeMeta:   t.TypeMeta,
							ObjectMeta: t.ObjectMeta,
							Spec:       t.Spec,
							Status: gwapiv1.PolicyStatus{
								Ancestors: mergeStatus(t.Namespace, r.envoyGateway.Gateway.ControllerName, t.Status.Ancestors, val.Ancestors),
							},
						}
						return tCopy
					}),
				})
			},
		)
		r.log.Info("backendTlsPolicy status subscriber shutting down")
	}()

	// EnvoyExtensionPolicy object status updater
	go func() {
		message.HandleSubscription(
			message.Metadata{Runner: string(egv1a1.LogComponentProviderRunner), Message: message.EnvoyExtensionPolicyStatusMessageName},
			r.subscriptions.envoyExtensionPolicyStatuses,
			func(update message.Update[types.NamespacedName, *gwapiv1.PolicyStatus], errChan chan error) {
				// skip delete updates.
				if update.Delete {
					return
				}
				key := update.Key
				val := update.Value
				r.statusUpdater.Send(Update{
					NamespacedName: key,
					Resource:       new(egv1a1.EnvoyExtensionPolicy),
					Mutator: MutatorFunc(func(obj client.Object) client.Object {
						t, ok := obj.(*egv1a1.EnvoyExtensionPolicy)
						if !ok {
							err := fmt.Errorf("unsupported object type %T", obj)
							errChan <- err
							panic(err)
						}
						tCopy := &egv1a1.EnvoyExtensionPolicy{
							TypeMeta:   t.TypeMeta,
							ObjectMeta: t.ObjectMeta,
							Spec:       t.Spec,
							Status: gwapiv1.PolicyStatus{
								Ancestors: mergeStatus(t.Namespace, r.envoyGateway.Gateway.ControllerName, t.Status.Ancestors, val.Ancestors),
							},
						}
						return tCopy
					}),
				})
			},
		)
		r.log.Info("envoyExtensionPolicy status subscriber shutting down")
	}()

	// Backend object status updater
	go func() {
		message.HandleSubscription(
			message.Metadata{Runner: string(egv1a1.LogComponentProviderRunner), Message: message.BackendStatusMessageName},
			r.subscriptions.backendStatuses,
			func(update message.Update[types.NamespacedName, *egv1a1.BackendStatus], errChan chan error) {
				// skip delete updates.
				if update.Delete {
					return
				}
				key := update.Key
				val := update.Value
				r.statusUpdater.Send(Update{
					NamespacedName: key,
					Resource:       new(egv1a1.Backend),
					Mutator: MutatorFunc(func(obj client.Object) client.Object {
						t, ok := obj.(*egv1a1.Backend)
						if !ok {
							err := fmt.Errorf("unsupported object type %T", obj)
							errChan <- err
							panic(err)
						}
						tCopy := &egv1a1.Backend{
							TypeMeta:   t.TypeMeta,
							ObjectMeta: t.ObjectMeta,
							Spec:       t.Spec,
							Status:     *val,
						}
						return tCopy
					}),
				})
			},
		)
		r.log.Info("backend status subscriber shutting down")
	}()

	if extensionManagerEnabled {
		// ExtensionServerPolicy object status updater
		go func() {
			message.HandleSubscription(
				message.Metadata{Runner: string(egv1a1.LogComponentProviderRunner), Message: message.ExtensionServerPoliciesStatusMessageName},
				r.subscriptions.extensionPolicyStatuses,
				func(update message.Update[message.NamespacedNameAndGVK, *gwapiv1.PolicyStatus], errChan chan error) {
					// skip delete updates.
					if update.Delete {
						return
					}
					key := update.Key
					val := update.Value
					obj := unstructured.Unstructured{}
					obj.SetGroupVersionKind(key.GroupVersionKind)

					r.statusUpdater.Send(Update{
						NamespacedName: key.NamespacedName,
						Resource:       &obj,
						Mutator: MutatorFunc(func(obj client.Object) client.Object {
							t, ok := obj.(*unstructured.Unstructured)
							if !ok {
								err := fmt.Errorf("unsupported object type %T", obj)
								errChan <- err
								panic(err)
							}
							tCopy := t.DeepCopy()
							var oldAncestors []gwapiv1.PolicyAncestorStatus
							o, found, err := unstructured.NestedFieldCopy(tCopy.Object, "status", "ancestors")
							if found && err == nil {
								oldAncestors, ok = o.([]gwapiv1.PolicyAncestorStatus)
								if ok {
									tCopy.Object["status"] = mergeStatus(t.GetNamespace(), r.envoyGateway.Gateway.ControllerName, oldAncestors, val.Ancestors)
									return tCopy
								}
							}
							tCopy.Object["status"] = *val
							return tCopy
						}),
					})
				},
			)
			r.log.Info("extensionServerPolicies status subscriber shutting down")
		}()
	}
}

// mergeStatus merges the old and new `RouteParentStatus`/`PolicyAncestorStatus`.
// This is needed because the `RouteParentStatus`/`PolicyAncestorStatus` doesn't support strategic merge patch yet.
// This depends on the fact that we get the full updated status of the route/policy (all parents/ancestors), and will break otherwise.
func mergeStatus[K interface{}](ns, controllerName string, old, new []K) []K {
	// Allocating with worst-case capacity to avoid reallocation.
	merged := make([]K, 0, len(old)+len(new))

	// Range over old status parentRefs in order:
	// 1. The parentRef exists in the new status: append the new one to the final status.
	// 2. The parentRef doesn't exist in the new status and it's not our controller: append it to the final status.
	// 3. The parentRef doesn't exist in the new status, and it is our controller: don't append it to the final status.
	for _, oldP := range old {
		found := -1
		for newI, newP := range new {
			if isParentOrAncestorRefEqual(oldP, newP, ns) {
				found = newI
				break
			}
		}
		if found >= 0 {
			merged = append(merged, new[found])
		} else if parentOrAncestorControllerName(oldP) != gwapiv1.GatewayController(controllerName) {
			merged = append(merged, oldP)
		}
	}

	// Range over new status parentRefs and make sure every parentRef exists in the final status. If not, append it.
	for _, newP := range new {
		found := false
		for _, mergedP := range merged {
			if isParentOrAncestorRefEqual(newP, mergedP, ns) {
				found = true
				break
			}
		}

		if !found {
			merged = append(merged, newP)
		}
	}
	return merged
}

func isParentOrAncestorRefEqual[K any](firstRef, secondRef K, ns string) bool {
	switch reflect.TypeOf(firstRef) {
	case reflect.TypeOf(gwapiv1.RouteParentStatus{}):
		return gatewayapi.IsParentRefEqual(any(firstRef).(gwapiv1.RouteParentStatus).ParentRef, any(secondRef).(gwapiv1.RouteParentStatus).ParentRef, ns)
	case reflect.TypeOf(gwapiv1.PolicyAncestorStatus{}):
		return gatewayapi.IsParentRefEqual(any(firstRef).(gwapiv1.PolicyAncestorStatus).AncestorRef, any(secondRef).(gwapiv1.PolicyAncestorStatus).AncestorRef, ns)
	default:
		return false
	}
}

func parentOrAncestorControllerName[K any](ref K) gwapiv1.GatewayController {
	switch reflect.TypeOf(ref) {
	case reflect.TypeOf(gwapiv1.RouteParentStatus{}):
		return any(ref).(gwapiv1.RouteParentStatus).ControllerName
	case reflect.TypeOf(gwapiv1.PolicyAncestorStatus{}):
		return any(ref).(gwapiv1.PolicyAncestorStatus).ControllerName
	default:
		return gwapiv1.GatewayController("")
	}
}

func (r *gatewayAPIReconciler) updateStatusForGateway(ctx context.Context, gtw *gwapiv1.Gateway) {
	// nil check for unit tests.
	if r.statusUpdater == nil {
		return
	}

	// Get envoyObjects
	envoyObj, err := r.envoyObjectForGateway(ctx, gtw)
	if err != nil {
		r.log.Info("failed to get Deployment for gateway",
			"namespace", gtw.Namespace, "name", gtw.Name)
	}

	// Get service
	svc, err := r.envoyServiceForGateway(ctx, gtw)
	if err != nil {
		r.log.Info("failed to get Service for gateway",
			"namespace", gtw.Namespace, "name", gtw.Name)
	}

	if status.GatewayAccepted(gtw) {
		// update accepted condition to true if it is not false
		// this is needed because the accepted condition is not set to true by the Gateway API translator
		// TODO (huabing): this is tricky and confusing for later readers, we should remove this and set the accepted condition
		// to true in the Gateway API translator
		status.UpdateGatewayStatusAccepted(gtw)
		// update address field and programmed condition
		status.UpdateGatewayStatusProgrammedCondition(gtw, svc, envoyObj, r.store.listNodeAddresses())
	}

	key := utils.NamespacedName(gtw)

	// publish status
	r.statusUpdater.Send(Update{
		NamespacedName: key,
		Resource:       new(gwapiv1.Gateway),
		Mutator: MutatorFunc(func(obj client.Object) client.Object {
			g, ok := obj.(*gwapiv1.Gateway)
			if !ok {
				panic(fmt.Sprintf("unsupported object type %T", obj))
			}
			gCopy := g.DeepCopy()
			gCopy.Status.Conditions = gtw.Status.Conditions
			gCopy.Status.Addresses = gtw.Status.Addresses
			gCopy.Status.Listeners = gtw.Status.Listeners
			return gCopy
		}),
	})
}
