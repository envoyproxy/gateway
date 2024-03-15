// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"fmt"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/status"
	"github.com/envoyproxy/gateway/internal/utils"
)

// subscribeAndUpdateStatus subscribes to gateway API object status updates and
// writes it into the Kubernetes API Server.
func (r *gatewayAPIReconciler) subscribeAndUpdateStatus(ctx context.Context) {
	// Gateway object status updater
	go func() {
		message.HandleSubscription(
			message.Metadata{Runner: string(v1alpha1.LogComponentProviderRunner), Message: "gateway-status"},
			r.resources.GatewayStatuses.Subscribe(ctx),
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
			message.Metadata{Runner: string(v1alpha1.LogComponentProviderRunner), Message: "httproute-status"},
			r.resources.HTTPRouteStatuses.Subscribe(ctx),
			func(update message.Update[types.NamespacedName, *gwapiv1.HTTPRouteStatus], errChan chan error) {
				// skip delete updates.
				if update.Delete {
					return
				}
				key := update.Key
				val := update.Value
				r.statusUpdater.Send(status.Update{
					NamespacedName: key,
					Resource:       new(gwapiv1.HTTPRoute),
					Mutator: status.MutatorFunc(func(obj client.Object) client.Object {
						h, ok := obj.(*gwapiv1.HTTPRoute)
						if !ok {
							err := fmt.Errorf("unsupported object type %T", obj)
							errChan <- err
							panic(err)
						}
						hCopy := h.DeepCopy()
						hCopy.Status.Parents = val.Parents
						return hCopy
					}),
				})
			},
		)
		r.log.Info("httpRoute status subscriber shutting down")
	}()

	// GRPCRoute object status updater
	go func() {
		message.HandleSubscription(message.Metadata{Runner: string(v1alpha1.LogComponentProviderRunner), Message: "grpcroute-status"}, r.resources.GRPCRouteStatuses.Subscribe(ctx),
			func(update message.Update[types.NamespacedName, *gwapiv1a2.GRPCRouteStatus], errChan chan error) {
				// skip delete updates.
				if update.Delete {
					return
				}
				key := update.Key
				val := update.Value
				r.statusUpdater.Send(status.Update{
					NamespacedName: key,
					Resource:       new(gwapiv1a2.GRPCRoute),
					Mutator: status.MutatorFunc(func(obj client.Object) client.Object {
						h, ok := obj.(*gwapiv1a2.GRPCRoute)
						if !ok {
							err := fmt.Errorf("unsupported object type %T", obj)
							errChan <- err
							panic(err)
						}
						hCopy := h.DeepCopy()
						hCopy.Status.Parents = val.Parents
						return hCopy
					}),
				})
			},
		)
		r.log.Info("grpcRoute status subscriber shutting down")
	}()

	// TLSRoute object status updater
	go func() {
		message.HandleSubscription(
			message.Metadata{Runner: string(v1alpha1.LogComponentProviderRunner), Message: "tlsroute-status"},
			r.resources.TLSRouteStatuses.Subscribe(ctx),
			func(update message.Update[types.NamespacedName, *gwapiv1a2.TLSRouteStatus], errChan chan error) {
				// skip delete updates.
				if update.Delete {
					return
				}
				key := update.Key
				val := update.Value
				r.statusUpdater.Send(status.Update{
					NamespacedName: key,
					Resource:       new(gwapiv1a2.TLSRoute),
					Mutator: status.MutatorFunc(func(obj client.Object) client.Object {
						t, ok := obj.(*gwapiv1a2.TLSRoute)
						if !ok {
							err := fmt.Errorf("unsupported object type %T", obj)
							errChan <- err
							panic(err)
						}
						tCopy := t.DeepCopy()
						tCopy.Status.Parents = val.Parents
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
			message.Metadata{Runner: string(v1alpha1.LogComponentProviderRunner), Message: "tcproute-status"},
			r.resources.TCPRouteStatuses.Subscribe(ctx),
			func(update message.Update[types.NamespacedName, *gwapiv1a2.TCPRouteStatus], errChan chan error) {
				// skip delete updates.
				if update.Delete {
					return
				}
				key := update.Key
				val := update.Value
				r.statusUpdater.Send(status.Update{
					NamespacedName: key,
					Resource:       new(gwapiv1a2.TCPRoute),
					Mutator: status.MutatorFunc(func(obj client.Object) client.Object {
						t, ok := obj.(*gwapiv1a2.TCPRoute)
						if !ok {
							err := fmt.Errorf("unsupported object type %T", obj)
							errChan <- err
							panic(err)
						}
						tCopy := t.DeepCopy()
						tCopy.Status.Parents = val.Parents
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
			message.Metadata{Runner: string(v1alpha1.LogComponentProviderRunner), Message: "udproute-status"},
			r.resources.UDPRouteStatuses.Subscribe(ctx),
			func(update message.Update[types.NamespacedName, *gwapiv1a2.UDPRouteStatus], errChan chan error) {
				// skip delete updates.
				if update.Delete {
					return
				}
				key := update.Key
				val := update.Value
				r.statusUpdater.Send(status.Update{
					NamespacedName: key,
					Resource:       new(gwapiv1a2.UDPRoute),
					Mutator: status.MutatorFunc(func(obj client.Object) client.Object {
						t, ok := obj.(*gwapiv1a2.UDPRoute)
						if !ok {
							err := fmt.Errorf("unsupported object type %T", obj)
							errChan <- err
							panic(err)
						}
						tCopy := t.DeepCopy()
						tCopy.Status.Parents = val.Parents
						return tCopy
					}),
				})
			},
		)
		r.log.Info("udpRoute status subscriber shutting down")
	}()

	// EnvoyPatchPolicy object status updater
	go func() {
		message.HandleSubscription(
			message.Metadata{Runner: string(v1alpha1.LogComponentProviderRunner), Message: "envoypatchpolicy-status"},
			r.resources.EnvoyPatchPolicyStatuses.Subscribe(ctx),
			func(update message.Update[types.NamespacedName, *gwapiv1a2.PolicyStatus], errChan chan error) {
				// skip delete updates.
				if update.Delete {
					return
				}
				key := update.Key
				val := update.Value
				r.statusUpdater.Send(status.Update{
					NamespacedName: key,
					Resource:       new(v1alpha1.EnvoyPatchPolicy),
					Mutator: status.MutatorFunc(func(obj client.Object) client.Object {
						t, ok := obj.(*v1alpha1.EnvoyPatchPolicy)
						if !ok {
							err := fmt.Errorf("unsupported object type %T", obj)
							errChan <- err
							panic(err)
						}
						tCopy := t.DeepCopy()
						tCopy.Status = *val
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
			message.Metadata{Runner: string(v1alpha1.LogComponentProviderRunner), Message: "clienttrafficpolicy-status"},
			r.resources.ClientTrafficPolicyStatuses.Subscribe(ctx),
			func(update message.Update[types.NamespacedName, *gwapiv1a2.PolicyStatus], errChan chan error) {
				// skip delete updates.
				if update.Delete {
					return
				}
				key := update.Key
				val := update.Value
				r.statusUpdater.Send(status.Update{
					NamespacedName: key,
					Resource:       new(v1alpha1.ClientTrafficPolicy),
					Mutator: status.MutatorFunc(func(obj client.Object) client.Object {
						t, ok := obj.(*v1alpha1.ClientTrafficPolicy)
						if !ok {
							err := fmt.Errorf("unsupported object type %T", obj)
							errChan <- err
							panic(err)
						}
						tCopy := t.DeepCopy()
						tCopy.Status = *val
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
			message.Metadata{Runner: string(v1alpha1.LogComponentProviderRunner), Message: "backendtrafficpolicy-status"},
			r.resources.BackendTrafficPolicyStatuses.Subscribe(ctx),
			func(update message.Update[types.NamespacedName, *gwapiv1a2.PolicyStatus], errChan chan error) {
				// skip delete updates.
				if update.Delete {
					return
				}
				key := update.Key
				val := update.Value
				r.statusUpdater.Send(status.Update{
					NamespacedName: key,
					Resource:       new(v1alpha1.BackendTrafficPolicy),
					Mutator: status.MutatorFunc(func(obj client.Object) client.Object {
						t, ok := obj.(*v1alpha1.BackendTrafficPolicy)
						if !ok {
							err := fmt.Errorf("unsupported object type %T", obj)
							errChan <- err
							panic(err)
						}
						tCopy := t.DeepCopy()
						tCopy.Status = *val
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
			message.Metadata{Runner: string(v1alpha1.LogComponentProviderRunner), Message: "securitypolicy-status"},
			r.resources.SecurityPolicyStatuses.Subscribe(ctx),
			func(update message.Update[types.NamespacedName, *gwapiv1a2.PolicyStatus], errChan chan error) {
				// skip delete updates.
				if update.Delete {
					return
				}
				key := update.Key
				val := update.Value
				r.statusUpdater.Send(status.Update{
					NamespacedName: key,
					Resource:       new(v1alpha1.SecurityPolicy),
					Mutator: status.MutatorFunc(func(obj client.Object) client.Object {
						t, ok := obj.(*v1alpha1.SecurityPolicy)
						if !ok {
							err := fmt.Errorf("unsupported object type %T", obj)
							errChan <- err
							panic(err)
						}
						tCopy := t.DeepCopy()
						tCopy.Status = *val
						return tCopy
					}),
				})
			},
		)
		r.log.Info("securityPolicy status subscriber shutting down")
	}()

	// BackendTLSPolicy object status updater
	go func() {
		message.HandleSubscription(message.Metadata{Runner: string(v1alpha1.LogComponentProviderRunner), Message: "backendtlspolicy-status"}, r.resources.BackendTLSPolicyStatuses.Subscribe(ctx),
			func(update message.Update[types.NamespacedName, *gwapiv1a2.PolicyStatus], errChan chan error) {
				// skip delete updates.
				if update.Delete {
					return
				}
				key := update.Key
				val := update.Value
				r.statusUpdater.Send(status.Update{
					NamespacedName: key,
					Resource:       new(gwapiv1a2.BackendTLSPolicy),
					Mutator: status.MutatorFunc(func(obj client.Object) client.Object {
						t, ok := obj.(*gwapiv1a2.BackendTLSPolicy)
						if !ok {
							err := fmt.Errorf("unsupported object type %T", obj)
							errChan <- err
							panic(err)
						}
						tCopy := t.DeepCopy()
						tCopy.Status = *val
						return tCopy
					}),
				})
			},
		)
		r.log.Info("backendTlsPolicy status subscriber shutting down")
	}()
}

func (r *gatewayAPIReconciler) updateStatusForGateway(ctx context.Context, gtw *gwapiv1.Gateway) {
	// nil check for unit tests.
	if r.statusUpdater == nil {
		return
	}

	// Get deployment
	deploy, err := r.envoyDeploymentForGateway(ctx, gtw)
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
	// update accepted condition
	status.UpdateGatewayStatusAcceptedCondition(gtw, true)
	// update address field and programmed condition
	status.UpdateGatewayStatusProgrammedCondition(gtw, svc, deploy, r.store.listNodeAddresses()...)

	key := utils.NamespacedName(gtw)

	// publish status
	r.statusUpdater.Send(status.Update{
		NamespacedName: key,
		Resource:       new(gwapiv1.Gateway),
		Mutator: status.MutatorFunc(func(obj client.Object) client.Object {
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

func (r *gatewayAPIReconciler) updateStatusForGatewayClass(
	ctx context.Context,
	gc *gwapiv1.GatewayClass,
	accepted bool,
	reason,
	msg string) error {
	if r.statusUpdater != nil {
		r.statusUpdater.Send(status.Update{
			NamespacedName: types.NamespacedName{Name: gc.Name},
			Resource:       &gwapiv1.GatewayClass{},
			Mutator: status.MutatorFunc(func(obj client.Object) client.Object {
				gc, ok := obj.(*gwapiv1.GatewayClass)
				if !ok {
					panic(fmt.Sprintf("unsupported object type %T", obj))
				}

				return status.SetGatewayClassAccepted(gc.DeepCopy(), accepted, reason, msg)
			}),
		})
	} else {
		// this branch makes testing easier by not going through the status.Updater.
		duplicate := status.SetGatewayClassAccepted(gc.DeepCopy(), accepted, reason, msg)

		if err := r.client.Status().Update(ctx, duplicate); err != nil && !kerrors.IsNotFound(err) {
			return fmt.Errorf("error updating status of gatewayclass %s: %w", duplicate.Name, err)
		}
	}
	return nil
}
