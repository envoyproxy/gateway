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
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/utils"
)

// subscribeAndUpdateStatus subscribes to gateway API object status updates and
// writes it into the Kubernetes API Server.
func (r *gatewayAPIReconciler) subscribeAndUpdateStatus(ctx context.Context, extensionManagerEnabled bool) {
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
									Parents: mergeRouteParentStatus(h.Namespace, h.Status.Parents, val.Parents),
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
									Parents: mergeRouteParentStatus(g.Namespace, g.Status.Parents, val.Parents),
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
					Resource:       new(gwapiv1a2.TLSRoute),
					Mutator: MutatorFunc(func(obj client.Object) client.Object {
						t, ok := obj.(*gwapiv1a2.TLSRoute)
						if !ok {
							err := fmt.Errorf("unsupported object type %T", obj)
							errChan <- err
							panic(err)
						}
						tCopy := &gwapiv1a2.TLSRoute{
							TypeMeta:   t.TypeMeta,
							ObjectMeta: t.ObjectMeta,
							Spec:       t.Spec,
							Status: gwapiv1a2.TLSRouteStatus{
								RouteStatus: gwapiv1.RouteStatus{
									Parents: mergeRouteParentStatus(t.Namespace, t.Status.Parents, val.Parents),
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
									Parents: mergeRouteParentStatus(t.Namespace, t.Status.Parents, val.Parents),
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
									Parents: mergeRouteParentStatus(u.Namespace, u.Status.Parents, val.Parents),
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
			func(update message.Update[types.NamespacedName, *gwapiv1a2.PolicyStatus], errChan chan error) {
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
							Status:     *val,
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
			func(update message.Update[types.NamespacedName, *gwapiv1a2.PolicyStatus], errChan chan error) {
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
							Status:     *val,
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
			func(update message.Update[types.NamespacedName, *gwapiv1a2.PolicyStatus], errChan chan error) {
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
							Status:     *val,
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
			func(update message.Update[types.NamespacedName, *gwapiv1a2.PolicyStatus], errChan chan error) {
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
							Status:     *val,
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
			func(update message.Update[types.NamespacedName, *gwapiv1a2.PolicyStatus], errChan chan error) {
				// skip delete updates.
				if update.Delete {
					return
				}
				key := update.Key
				val := update.Value
				r.statusUpdater.Send(Update{
					NamespacedName: key,
					Resource:       new(gwapiv1a3.BackendTLSPolicy),
					Mutator: MutatorFunc(func(obj client.Object) client.Object {
						t, ok := obj.(*gwapiv1a3.BackendTLSPolicy)
						if !ok {
							err := fmt.Errorf("unsupported object type %T", obj)
							errChan <- err
							panic(err)
						}
						tCopy := &gwapiv1a3.BackendTLSPolicy{
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
		r.log.Info("backendTlsPolicy status subscriber shutting down")
	}()

	// EnvoyExtensionPolicy object status updater
	go func() {
		message.HandleSubscription(
			message.Metadata{Runner: string(egv1a1.LogComponentProviderRunner), Message: message.EnvoyExtensionPolicyStatusMessageName},
			r.subscriptions.envoyExtensionPolicyStatuses,
			func(update message.Update[types.NamespacedName, *gwapiv1a2.PolicyStatus], errChan chan error) {
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
							Status:     *val,
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
				func(update message.Update[message.NamespacedNameAndGVK, *gwapiv1a2.PolicyStatus], errChan chan error) {
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

// mergeRouteParentStatus merges the old and new RouteParentStatus.
// This is needed because the RouteParentStatus doesn't support strategic merge patch yet.
func mergeRouteParentStatus(ns string, old, new []gwapiv1.RouteParentStatus) []gwapiv1.RouteParentStatus {
	// Allocating with worst-case capacity to avoid reallocation.
	merged := make([]gwapiv1.RouteParentStatus, 0, len(old)+len(new))

	// Range over old status parentRefs in order:
	// 1. The parentRef exists in the new status: append the new one to the final status.
	// 2. The parentRef doesn't exist in the new status and it's not our controller: append it to the final status.
	// 3. The parentRef doesn't exist in the new status, and it is our controller: keep it in the final status.
	//    This is important for routes with multiple parent references - not all parents are updated in each reconciliation.
	for _, oldP := range old {
		found := -1
		for newI, newP := range new {
			if isParentRefEqual(oldP.ParentRef, newP.ParentRef, ns) {
				found = newI
				break
			}
		}
		if found >= 0 {
			merged = append(merged, new[found])
		} else {
			// Keep all old parent statuses, regardless of controller.
			// For routes with multiple parents managed by the same controller,
			// not all parents are necessarily updated in each reconciliation.
			merged = append(merged, oldP)
		}
	}

	// Range over new status parentRefs and make sure every parentRef exists in the final status. If not, append it.
	for _, newP := range new {
		found := false
		for _, mergedP := range merged {
			if isParentRefEqual(newP.ParentRef, mergedP.ParentRef, ns) {
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

func isParentRefEqual(ref1, ref2 gwapiv1.ParentReference, routeNS string) bool {
	defaultGroup := (*gwapiv1.Group)(&gwapiv1.GroupVersion.Group)
	if ref1.Group == nil {
		ref1.Group = defaultGroup
	}
	if ref2.Group == nil {
		ref2.Group = defaultGroup
	}

	defaultKind := gwapiv1.Kind(resource.KindGateway)
	if ref1.Kind == nil {
		ref1.Kind = &defaultKind
	}
	if ref2.Kind == nil {
		ref2.Kind = &defaultKind
	}

	// If the parent's namespace is not set, default to the namespace of the Route.
	defaultNS := gwapiv1.Namespace(routeNS)
	if ref1.Namespace == nil {
		ref1.Namespace = &defaultNS
	}
	if ref2.Namespace == nil {
		ref2.Namespace = &defaultNS
	}
	return reflect.DeepEqual(ref1, ref2)
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
