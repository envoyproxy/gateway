// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/provider/utils"
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

func (r *gatewayAPIReconciler) validateHTTPRouteForReconcile(obj client.Object) bool {
	hr, ok := obj.(*gwapiv1b1.HTTPRoute)
	if !ok {
		r.log.Info("unexpected object type, bypassing reconciliation", "object", obj)
		return false
	}

	parentReferences := hr.Spec.ParentRefs
	return r.validateRouteParentReferences(parentReferences, hr.Namespace)
}

func (r *gatewayAPIReconciler) validateTLSRouteForReconcile(obj client.Object) bool {
	tr, ok := obj.(*gwapiv1a2.TLSRoute)
	if !ok {
		r.log.Info("unexpected object type, bypassing reconciliation", "object", obj)
		return false
	}

	parentReferences := gatewayapi.UpgradeParentReferences(tr.Spec.ParentRefs)
	return r.validateRouteParentReferences(parentReferences, tr.Namespace)
}

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
		}
	}

	return true
}

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
		if !r.validateGatewayForReconcile(&gw) {
			return false
		}
	}

	return true
}
