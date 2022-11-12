// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// This file contains code derived from Contour,
// https://github.com/projectcontour/contour
// from the source file
// https://github.com/projectcontour/contour/blob/main/internal/controller/gateway.go// and is provided here subject to the following:
// Copyright Project Contour Authors
// SPDX-License-Identifier: Apache-2.0

package kubernetes

import (
	"context"
	"fmt"

	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/provider/utils"
	"github.com/envoyproxy/gateway/internal/status"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

// processGateway processes the Gateway coming from the watcher and eventually
// reconciles the parent GatewayClass.
func (r *gatewayAPIReconciler) processGateway(obj client.Object) []reconcile.Request {
	r.log.Info("processing gateway", "namespace", obj.GetNamespace(), "name", obj.GetName())
	ctx := context.Background()
	requests := []reconcile.Request{}

	gwkey := types.NamespacedName{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}

	gatewayDeleted := false
	gatewayClasses := []string{}

	gw := new(gwapiv1b1.Gateway)
	if err := r.client.Get(ctx, gwkey, gw); err != nil {
		if !kerrors.IsNotFound(err) {
			r.log.Error(err, "failed to get gateway")
			return requests
		}

		gatewayDeleted = true
		if resourceGateway, found := r.resources.Gateways.Load(gwkey); found {
			gatewayClasses = append(gatewayClasses, string(resourceGateway.Spec.GatewayClassName))
			// Remove the Gateway from watchable map.
			r.resources.Gateways.Delete(gwkey)
		}
	}

	allClasses := &gwapiv1b1.GatewayClassList{}
	if err := r.client.List(ctx, allClasses); err != nil {
		r.log.Error(err, "error listing gatewayclasses")
		return requests
	}

	// Find the GatewayClass for this controller with Accepted=true status condition.
	acceptedClass := r.acceptedClass(allClasses)
	if acceptedClass == nil {
		r.log.Info("No accepted gatewayclass found for gateway", "namespace", gwkey.Namespace,
			"name", gwkey.Name)
		r.log.Info(fmt.Sprintf("heyo %v", r.resources == nil))
		if r.resources != nil {
			for namespacedName := range r.resources.Gateways.LoadAll() {
				r.resources.Gateways.Delete(namespacedName)
			}
		}
		return requests
	}

	allGateways := &gwapiv1b1.GatewayList{}
	if err := r.client.List(ctx, allGateways); err != nil {
		r.log.Error(err, "error listing gateways")
		return requests
	}

	// Get all the Gateways for the Accepted=true GatewayClass.
	acceptedGateways := gatewaysOfClass(acceptedClass, allGateways)
	if len(acceptedGateways) == 0 {
		r.log.Info("No gateways found for accepted gatewayclass")
		// If needed, remove the finalizer from the accepted GatewayClass.
		if acceptedClass != nil {
			if err := r.removeFinalizer(ctx, acceptedClass); err != nil {
				r.log.Error(err, fmt.Sprintf("failed to remove finalizer from gatewayclass %s",
					acceptedClass.Name))
				return requests
			}
		}
	} else {
		// If needed, finalize the accepted GatewayClass.
		if err := r.addFinalizer(ctx, acceptedClass); err != nil {
			r.log.Error(err, fmt.Sprintf("failed adding finalizer to gatewayclass %s",
				acceptedClass.Name))
			return requests
		}
	}

	found := false
	// Set status conditions for all accepted gateways.
	for i := range acceptedGateways {
		gw := acceptedGateways[i]

		// Get the status of the Gateway's associated Envoy Deployment.
		deployment, err := r.envoyDeploymentForGateway(ctx, &gw)
		if err != nil {
			r.log.Info("failed to get deployment for gateway",
				"namespace", gw.Namespace, "name", gw.Name)
		}

		// Get the status address of the Gateway's associated Envoy Service.
		svc, err := r.envoyServiceForGateway(ctx, &gw)
		if err != nil {
			r.log.Info("failed to get service for gateway",
				"namespace", gw.Namespace, "name", gw.Name)
		}

		// Get the secret and referenceGrants of the Gateway's TLS configuration.
		secrets, refGrants, err := r.secretsAndRefGrantsForGateway(ctx, &gw)
		if err != nil {
			r.log.Info("failed to get secrets and referencegrants for gateway",
				"namespace", gw.Namespace, "name", gw.Name)
		}
		for i := range secrets {
			secret := secrets[i]
			// Store the secrets in the resource map.
			key := utils.NamespacedName(&secret)
			r.resources.Secrets.Store(key, &secret)
		}
		for i := range refGrants {
			rg := refGrants[i]
			// Store the referencegrants in the resource map.
			key := utils.NamespacedName(&rg)
			r.resources.ReferenceGrants.Store(key, &rg)
			// Store the referencegrant namespace in the resource map.
			key = types.NamespacedName{Name: rg.Namespace}
			refNs := corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: rg.Namespace,
				},
			}
			if err := r.client.Get(ctx, key, &refNs); err != nil {
				r.log.Info("failed to get referencegrant namespace", "name", refNs.Name)
				return requests
			}
			r.resources.Namespaces.Store(refNs.Name, &refNs)
		}

		// update scheduled condition
		status.UpdateGatewayStatusScheduledCondition(&gw, true)
		// update address field and ready condition
		status.UpdateGatewayStatusReadyCondition(&gw, svc, deployment)

		key := utils.NamespacedName(&gw)
		// publish status
		// do it inline since this code flow updates the
		// Status.Addresses field whereas the message bus / subscriber
		// does not.
		r.statusUpdater.Send(status.Update{
			NamespacedName: key,
			Resource:       new(gwapiv1b1.Gateway),
			Mutator: status.MutatorFunc(func(obj client.Object) client.Object {
				g, ok := obj.(*gwapiv1b1.Gateway)
				if !ok {
					panic(fmt.Sprintf("unsupported object type %T", obj))
				}
				gCopy := g.DeepCopy()
				gCopy.Status.Conditions = status.MergeConditions(gCopy.Status.Conditions, gw.Status.Conditions...)
				gCopy.Status.Addresses = gw.Status.Addresses
				return gCopy

			}),
		})

		// only store the resource if it does not exist or it has a newer spec.
		if v, ok := r.resources.Gateways.Load(key); !ok || (gw.Generation > v.Generation) {
			r.resources.Gateways.Store(key, &gw)
		}
		if key == utils.NamespacedName(obj) {
			found = true
		}
	}

	if !found {
		r.resources.Gateways.Delete(utils.NamespacedName(obj))

		// Delete the TLS secrets from the resource map if no other managed
		// Gateways reference them.
		secrets, _, err := r.secretsAndRefGrantsForGateway(ctx, gw)
		if err != nil {
			r.log.Info("failed to get secrets and referencegrants for gateway",
				"namespace", gw.Namespace, "name", gw.Name)
		}
		for i := range secrets {
			secret := secrets[i]
			referenced, err := r.gatewaysRefSecret(ctx, &secret)
			switch {
			case err != nil:
				r.log.Error(err, "failed to verify if other gateways reference secret")
			case !referenced:
				r.log.Info("no other gateways reference secret; deleting from resource map",
					"namespace", secret.Namespace, "name", secret.Name)
				key := utils.NamespacedName(&secret)
				r.resources.Secrets.Delete(key)
			default:
				r.log.Info("other gateways reference secret; keeping the secret in the resource map",
					"namespace", secret.Namespace, "name", secret.Name)
			}
		}
	}

	if !gatewayDeleted {
		// only store the resource if it does not exist or it has a newer spec.
		if v, ok := r.resources.Gateways.Load(gwkey); !ok || (gw.Generation > v.Generation) {
			r.resources.Gateways.Store(gwkey, gw.DeepCopy())
			r.log.Info("added gateway to watchable map")
		}

		if len(gatewayClasses) == 0 || gatewayClasses[0] != string(gw.Spec.GatewayClassName) {
			gatewayClasses = append(gatewayClasses, string(gw.Spec.GatewayClassName))
		}
	}

	// To handle the GatewayClassName update in the, both the old and new
	// Gateway object are passed through this transformation function.
	for _, gwclass := range gatewayClasses {
		requests = append(requests, reconcile.Request{
			NamespacedName: types.NamespacedName{Name: gwclass},
		})
	}

	return requests
}

// gatewaysRefSecret returns true if a managed Gateway references the provided secret.
// An error is returned if an error is encountered while checking.
func (r *gatewayAPIReconciler) gatewaysRefSecret(ctx context.Context, secret *corev1.Secret) (bool, error) {
	if secret == nil {
		return false, fmt.Errorf("secret is nil")
	}
	gateways := &gwapiv1b1.GatewayList{}
	if err := r.client.List(ctx, gateways); err != nil {
		return false, fmt.Errorf("error listing gatewayclasses: %v", err)
	}
	for i := range gateways.Items {
		gw := gateways.Items[i]
		if r.hasMatchingControllerForGateway(&gw) {
			secrets, _, err := r.secretsAndRefGrantsForGateway(ctx, &gw)
			if err != nil {
				return false, err
			}
			for _, s := range secrets {
				if s.Namespace == secret.Namespace && s.Name == secret.Name {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

// hasMatchingControllerForGateway returns true if the provided object is a Gateway
// using a GatewayClass matching the configured gatewayclass controller name.
func (r *gatewayAPIReconciler) hasMatchingControllerForGateway(obj client.Object) bool {
	gw, ok := obj.(*gwapiv1b1.Gateway)
	if !ok {
		r.log.Info("unexpected object type, bypassing reconciliation", "object", obj)
		return false
	}

	gc := &gwapiv1b1.GatewayClass{}
	key := types.NamespacedName{Name: string(gw.Spec.GatewayClassName)}
	if err := r.client.Get(context.Background(), key, gc); err != nil {
		r.log.Error(err, "failed to get gatewayclass", "name", gw.Spec.GatewayClassName)
		return false
	}

	if gc.Spec.ControllerName != r.classController {
		r.log.Info("gatewayclass name for gateway doesn't match configured name",
			"namespace", gw.Namespace, "name", gw.Name)
		return false
	}

	return true
}

// secretsAndRefGrantsForGateway returns the Secrets referenced by the provided gateway listeners.
// If the provided Gateway references a Secret in a different namespace, a list of
// ReferenceGrants is returned that permit the cross namespace Secret reference.
func (r *gatewayAPIReconciler) secretsAndRefGrantsForGateway(ctx context.Context, gateway *gwapiv1b1.Gateway) ([]corev1.Secret, []gwapiv1a2.ReferenceGrant, error) {
	var secrets []corev1.Secret
	var returnedGrants []gwapiv1a2.ReferenceGrant
	for i := range gateway.Spec.Listeners {
		listener := gateway.Spec.Listeners[i]
		if terminatesTLS(&listener) {
			for j := range listener.TLS.CertificateRefs {
				ref := listener.TLS.CertificateRefs[j]
				if refsSecret(&ref) {
					if ref.Namespace != nil {
						// A ReferenceGrant is required for cross namespace secret references.
						refGrants := &gwapiv1a2.ReferenceGrantList{}
						opts := client.ListOptions{Namespace: string(*ref.Namespace)}
						if err := r.client.List(ctx, refGrants, &opts); err != nil {
							return nil, nil, fmt.Errorf("error listing referencegrants")
						}
						var gwRefd, secretRefd bool
						for _, rg := range refGrants.Items {
							for _, from := range rg.Spec.From {
								if from.Group == gwapiv1a2.GroupName &&
									from.Kind == gatewayapi.KindGateway {
									gwRefd = true
									break
								}
							}
							for _, to := range rg.Spec.To {
								if to.Group == corev1.GroupName &&
									to.Kind == gatewayapi.KindSecret {
									if to.Name == nil || *to.Name == gwapiv1a2.ObjectName(ref.Name) {
										secretRefd = true
										break
									}
								}
							}
							if gwRefd && secretRefd {
								returnedGrants = append(returnedGrants, rg)
								key := types.NamespacedName{
									Namespace: string(*ref.Namespace),
									Name:      string(ref.Name),
								}
								secret := new(corev1.Secret)
								if err := r.client.Get(ctx, key, secret); err != nil {
									r.resources.Secrets.Delete(key)
									return nil, nil, fmt.Errorf("failed to get secret: %v", err)
								}
								secrets = append(secrets, *secret)
							}
						}
					} else {
						// The secret is in the Gateway's namespace.
						key := types.NamespacedName{
							Namespace: gateway.Namespace,
							Name:      string(ref.Name),
						}
						secret := new(corev1.Secret)
						if err := r.client.Get(ctx, key, secret); err != nil {
							r.resources.Secrets.Delete(key)
							return nil, nil, fmt.Errorf("failed to get secret: %v", err)
						}
						secrets = append(secrets, *secret)
					}
				}
			}
		}
	}

	return secrets, returnedGrants, nil
}

// envoyDeploymentForGateway returns the Envoy Deployment, returning nil if the Deployment doesn't exist.
func (r *gatewayAPIReconciler) envoyDeploymentForGateway(ctx context.Context, gateway *gwapiv1b1.Gateway) (*appsv1.Deployment, error) {
	key := types.NamespacedName{
		Namespace: config.EnvoyGatewayNamespace,
		Name:      infraDeploymentName(gateway),
	}
	deployment := new(appsv1.Deployment)
	if err := r.client.Get(ctx, key, deployment); err != nil {
		if kerrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return deployment, nil
}

// envoyServiceForGateway returns the Envoy service, returning nil if the service doesn't exist.
func (r *gatewayAPIReconciler) envoyServiceForGateway(ctx context.Context, gateway *gwapiv1b1.Gateway) (*corev1.Service, error) {
	key := types.NamespacedName{
		Namespace: config.EnvoyGatewayNamespace,
		Name:      infraServiceName(gateway),
	}
	svc := new(corev1.Service)
	if err := r.client.Get(ctx, key, svc); err != nil {
		if kerrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return svc, nil
}
