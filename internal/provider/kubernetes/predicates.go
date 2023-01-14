// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/provider/utils"
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

	if len(gwList.Items) == 0 {
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

	// Check if the Service belongs to a Gateway, if so, update the Gateway status.
	gtw := r.findOwningGateway(ctx, svc.GetLabels())
	if gtw != nil {
		r.statusUpdateForGateway(ctx, gtw)
		return false
	}

	httpRouteList := &gwapiv1b1.HTTPRouteList{}
	if err := r.client.List(ctx, httpRouteList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(serviceHTTPRouteIndex, utils.NamespacedName(svc).String()),
	}); err != nil {
		r.log.Error(err, "unable to find associated HTTPRoutes")
		return false
	}

	grpcRouteList := &gwapiv1a2.GRPCRouteList{}
	if err := r.client.List(ctx, grpcRouteList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(serviceGRPCRouteIndex, utils.NamespacedName(svc).String()),
	}); err != nil {
		r.log.Error(err, "unable to find associated GRPCRoutes")
		return false
	}

	tlsRouteList := &gwapiv1a2.TLSRouteList{}
	if err := r.client.List(ctx, tlsRouteList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(serviceTLSRouteIndex, utils.NamespacedName(svc).String()),
	}); err != nil {
		r.log.Error(err, "unable to find associated TLSRoutes")
		return false
	}

	tcpRouteList := &gwapiv1a2.TCPRouteList{}
	if err := r.client.List(ctx, tcpRouteList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(serviceTCPRouteIndex, utils.NamespacedName(svc).String()),
	}); err != nil {
		r.log.Error(err, "unable to find associated TCPRoutes")
		return false
	}

	udpRouteList := &gwapiv1a2.UDPRouteList{}
	if err := r.client.List(ctx, udpRouteList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(serviceUDPRouteIndex, utils.NamespacedName(svc).String()),
	}); err != nil {
		r.log.Error(err, "unable to find associated UDPRoutes")
		return false
	}

	// Check how many Route objects refer this Service
	allAssociatedRoutes := len(httpRouteList.Items) +
		len(grpcRouteList.Items) +
		len(tlsRouteList.Items) +
		len(tcpRouteList.Items) +
		len(udpRouteList.Items)

	return allAssociatedRoutes != 0
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

	// Only deployments in the configured namespace should be reconciled.
	if deployment.Namespace == r.namespace {
		// Check if the deployment belongs to a Gateway, if so, update the Gateway status.
		gtw := r.findOwningGateway(ctx, deployment.GetLabels())
		if gtw != nil {
			r.statusUpdateForGateway(ctx, gtw)
			return false
		}
	}

	// There is no need to reconcile the Deployment any further.
	return false
}

// httpRoutesForAuthenticationFilter tries finding HTTPRoute referents of the provided
// AuthenticationFilter and returns true if any exist.
func (r *gatewayAPIReconciler) httpRoutesForAuthenticationFilter(obj client.Object) bool {
	ctx := context.Background()
	filter, ok := obj.(*egv1a1.AuthenticationFilter)
	if !ok {
		r.log.Info("unexpected object type, bypassing reconciliation", "object", obj)
		return false
	}

	// Check if the AuthenticationFilter belongs to a managed HTTPRoute.
	httpRouteList := &gwapiv1b1.HTTPRouteList{}
	if err := r.client.List(ctx, httpRouteList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(authenFilterHTTPRouteIndex, utils.NamespacedName(filter).String()),
	}); err != nil {
		r.log.Error(err, "unable to find associated HTTPRoutes")
		return false
	}

	return len(httpRouteList.Items) != 0
}

// httpRoutesForRateLimitFilter tries finding HTTPRoute referents of the provided
// RateLimitFilter and returns true if any exist.
func (r *gatewayAPIReconciler) httpRoutesForRateLimitFilter(obj client.Object) bool {
	ctx := context.Background()
	filter, ok := obj.(*egv1a1.RateLimitFilter)
	if !ok {
		r.log.Info("unexpected object type, bypassing reconciliation", "object", obj)
		return false
	}

	// Check if the RateLimitFilter belongs to a managed HTTPRoute.
	httpRouteList := &gwapiv1b1.HTTPRouteList{}
	if err := r.client.List(ctx, httpRouteList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(rateLimitFilterHTTPRouteIndex, utils.NamespacedName(filter).String()),
	}); err != nil {
		r.log.Error(err, "unable to find associated HTTPRoutes")
		return false
	}

	return len(httpRouteList.Items) != 0
}

// envoyDeploymentForGateway returns the Envoy Deployment, returning nil if the Deployment doesn't exist.
func (r *gatewayAPIReconciler) envoyDeploymentForGateway(ctx context.Context, gateway *gwapiv1b1.Gateway) (*appsv1.Deployment, error) {
	key := types.NamespacedName{
		Namespace: r.namespace,
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
		Namespace: r.namespace,
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

// findOwningGateway attempts finds a Gateway using "labels".
func (r gatewayAPIReconciler) findOwningGateway(ctx context.Context, labels map[string]string) *gwapiv1b1.Gateway {
	gwName, ok := labels[gatewayapi.OwningGatewayNameLabel]
	if !ok {
		return nil
	}

	gwNamespace, ok := labels[gatewayapi.OwningGatewayNamespaceLabel]
	if !ok {
		return nil
	}

	gatewayKey := types.NamespacedName{Namespace: gwNamespace, Name: gwName}
	gtw := new(gwapiv1b1.Gateway)
	if err := r.client.Get(ctx, gatewayKey, gtw); err != nil {
		r.log.Error(err, "gateway not found")
		return nil
	}

	return gtw
}
