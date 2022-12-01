// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/provider/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

// TODO: all predicate functions are unti test candidates.

// hasMatchingController returns true if the provided object is a GatewayClass
// with a Spec.Controller string matching this Envoy Gateway's controller string,
// or false otherwise.
func (r *gatewayAPIReconciler) hasMatchingController(obj client.Object) bool {
	gc, ok := obj.(*gwapiv1b1.GatewayClass)
	if !ok {
		r.log.Info("bypassing reconciliation due to unexpected object type", "type", obj)
		return false
	}

	if gc.Spec.ControllerName == r.classController {
		r.log.Info("gatewayclass has matching controller name, processing", "name", gc.Name)
		return true
	}

	r.log.Info("bypassing reconciliation due to controller name", "controller", gc.Spec.ControllerName)
	return false
}

// validateGatewayForReconcile returns true if the provided object is a Gateway
// using a GatewayClass matching the configured gatewayclass controller name.
func (r *gatewayAPIReconciler) validateGatewayForReconcile(obj client.Object) bool {
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

// validateTLSRouteForReconcile checks whether the HTTPRoute refers any valid Gateway.
func (r *gatewayAPIReconciler) validateHTTPRouteForReconcile(obj client.Object) bool {
	hr, ok := obj.(*gwapiv1b1.HTTPRoute)
	if !ok {
		r.log.Info("unexpected object type, bypassing reconciliation", "object", obj)
		return false
	}

	parentReferences := hr.Spec.ParentRefs
	return r.validateRouteParentReferences(parentReferences, hr.Namespace)
}

// validateTLSRouteForReconcile checks whether the TLSRoute refers any valid Gateway.
func (r *gatewayAPIReconciler) validateTLSRouteForReconcile(obj client.Object) bool {
	tr, ok := obj.(*gwapiv1a2.TLSRoute)
	if !ok {
		r.log.Info("unexpected object type, bypassing reconciliation", "object", obj)
		return false
	}

	parentReferences := gatewayapi.UpgradeParentReferences(tr.Spec.ParentRefs)
	return r.validateRouteParentReferences(parentReferences, tr.Namespace)
}

// validateRouteParentReferences checks whether the parent references of a given Route
// object, point to valid Gateways.
func (r *gatewayAPIReconciler) validateRouteParentReferences(refs []gwapiv1b1.ParentReference, defaultNamespace string) bool {
	for _, ref := range refs {
		if ref.Kind != nil && *ref.Kind == gatewayapi.KindGateway {
			key := types.NamespacedName{
				Namespace: gatewayapi.NamespaceDerefOr(ref.Namespace, defaultNamespace),
				Name:      string(ref.Name),
			}

			gw := &gwapiv1b1.Gateway{}
			if err := r.client.Get(context.Background(), key, gw); err != nil {
				r.log.Error(err, "failed to get gateway", "namespace", key.Namespace, "name", key.Name)
				return false
			}

			if !r.validateGatewayForReconcile(gw) {
				return false
			}

			// Even if one of the parent references points to a valid Gateway, we
			// must reconcile the Route object.
			return true
		}
	}

	return true
}

// validateSecretForReconcile checks whether the Secret belongs to a valid Gateway.
func (r *gatewayAPIReconciler) validateSecretForReconcile(obj client.Object) bool {
	secret, ok := obj.(*corev1.Secret)
	if !ok {
		r.log.Info("unexpected object type, bypassing reconciliation", "object", obj)
		return false
	}

	gwList := &gwapiv1b1.GatewayList{}
	if err := r.client.List(context.Background(), gwList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(secretGatewayIndex, utils.NamespacedName(secret).String()),
	}); err != nil {
		r.log.Error(err, "unable to find associated HTTPRoutes")
		return false
	}

	for _, gw := range gwList.Items {
		gw := gw
		if !r.validateGatewayForReconcile(&gw) {
			return false
		}
	}

	return true
}

// validateServiceForReconcile tries finding the owning Gateway of the Service
// if it exists, finds the Gateway's Deployment, and further updates the Gateway
// status Ready condition. All Services are pushed for reconciliation.
func (r *gatewayAPIReconciler) validateServiceForReconcile(obj client.Object) bool {
	ctx := context.Background()
	svc, ok := obj.(*corev1.Service)
	if !ok {
		r.log.Info("unexpected object type, bypassing reconciliation", "object", obj)
		return false
	}

	// Check if the Service belongs to a Gateway, if so, find the Gateway. If
	gtw := r.findOwningGateway(ctx, svc.GetLabels())
	if gtw != nil {
		// Check if the Deployment for the Gateway also exists, if it does, proceed with
		// the Gateway status update.
		deployment, err := r.envoyDeploymentForGateway(ctx, gtw)
		if err != nil {
			r.log.Info("failed to get Deployment for gateway",
				"namespace", gtw.Namespace, "name", gtw.Name)
			return false
		}

		r.statusUpdateForGateway(gtw, svc, deployment)
		return true
	}

	// TODO: further filter only those services that are referred by HTTPRoutes
	return true
}

// validateDeploymentForReconcile tries finding the owning Gateway of the Deployment
// if it exists, finds the Gateway's Service, and further updates the Gateway
// status Ready condition. No Deployments are pushed for reconciliation.
func (r *gatewayAPIReconciler) validateDeploymentForReconcile(obj client.Object) bool {
	ctx := context.Background()
	deployment, ok := obj.(*appsv1.Deployment)
	if !ok {
		r.log.Info("unexpected object type, bypassing reconciliation", "object", obj)
		return false
	}

	// Check if the deployment belongs to a Gateway, if so, find the Gateway.
	gtw := r.findOwningGateway(ctx, deployment.GetLabels())
	if gtw != nil {
		// Check if the Service for the Gateway also exists, if it does, proceed with
		// the Gateway status update.
		svc, err := r.envoyServiceForGateway(ctx, gtw)
		if err != nil {
			r.log.Info("failed to get Service for gateway",
				"namespace", gtw.Namespace, "name", gtw.Name)
			return false
		}

		r.statusUpdateForGateway(gtw, svc, deployment)
	}

	// There is no need to reconcile the Deployment any further.
	return false
}
