// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwapiv1a3 "sigs.k8s.io/gateway-api/apis/v1alpha3"
	mcsapiv1a1 "sigs.k8s.io/mcs-api/pkg/apis/v1alpha1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/utils"
)

// nolint: gosec
const (
	oidcHMACSecretName = "envoy-oidc-hmac"
	envoyTLSSecretName = "envoy"
)

// hasMatchingController returns true if the provided object is a GatewayClass
// with a Spec.Controller string matching this Envoy Gateway's controller string,
// or false otherwise.
func (r *gatewayAPIReconciler) hasMatchingController(gc *gwapiv1.GatewayClass) bool {
	if gc.Spec.ControllerName == r.classController {
		r.log.Info("gatewayclass has matching controller name, processing", "name", gc.Name)
		return true
	}

	r.log.Info("bypassing reconciliation due to controller name", "controller", gc.Spec.ControllerName)
	return false
}

// hasMatchingNamespaceLabels returns true if the namespace of provided object has
// the provided labels or false otherwise.
func (r *gatewayAPIReconciler) hasMatchingNamespaceLabels(obj client.Object) bool {
	ok, err := r.checkObjectNamespaceLabels(obj)
	if err != nil {
		r.log.Error(
			err, "failed to get Namespace",
			"object", obj.GetObjectKind().GroupVersionKind().String(),
			"name", obj.GetName())
		return false
	}
	return ok
}

type NamespaceGetter interface {
	GetNamespace() string
}

// checkObjectNamespaceLabels checks if labels of namespace of the object is a subset of namespaceLabels
func (r *gatewayAPIReconciler) checkObjectNamespaceLabels(obj metav1.Object) (bool, error) {
	var nsString string
	// TODO: it requires extra condition validate cluster resources or resources without namespace?
	if nsString = obj.GetNamespace(); len(nsString) == 0 {
		return false, nil
	}

	ns := &corev1.Namespace{}
	if err := r.client.Get(
		context.Background(),
		client.ObjectKey{
			Namespace: "", // Namespace object should have an empty Namespace
			Name:      nsString,
		},
		ns,
	); err != nil {
		return false, err
	}

	return matchLabelsAndExpressions(r.namespaceLabel, ns.Labels), nil
}

// matchLabelsAndExpressions extracts information from a given label selector and checks whether
// the provided object labels match the selector criteria.
// If the label selector is nil, it returns true, indicating a match.
// It returns false if there is an error while converting the label selector or if the labels do not match.
func matchLabelsAndExpressions(ls *metav1.LabelSelector, objLabels map[string]string) bool {
	if ls == nil {
		return true
	}

	selector, err := metav1.LabelSelectorAsSelector(ls)
	if err != nil {
		return false
	}

	return selector.Matches(labels.Set(objLabels))
}

// validateGatewayForReconcile returns true if the provided object is a Gateway
// using a GatewayClass matching the configured GatewayClass controller name.
func (r *gatewayAPIReconciler) validateGatewayForReconcile(obj client.Object) bool {
	gw, ok := obj.(*gwapiv1.Gateway)
	if !ok {
		r.log.Info("unexpected object type, bypassing reconciliation", "object", obj)
		return false
	}

	gc := &gwapiv1.GatewayClass{}
	key := types.NamespacedName{Name: string(gw.Spec.GatewayClassName)}
	if err := r.client.Get(context.Background(), key, gc); err != nil {
		r.log.Error(err, "failed to get GatewayClass", "name", gw.Spec.GatewayClassName)
		return false
	}

	if gc.Spec.ControllerName != r.classController {
		r.log.Info("GatewayClass name for gateway doesn't match configured name",
			"namespace", gw.Namespace, "name", gw.Name,
			"GatewayClass controller name", string(gc.Spec.ControllerName),
			"class controller name", string(r.classController))
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

	nsName := utils.NamespacedName(secret)

	if r.isGatewayReferencingSecret(&nsName) {
		return true
	}

	if r.spCRDExists {
		if r.isSecurityPolicyReferencingSecret(&nsName) {
			return true
		}
	}

	if r.ctpCRDExists {
		if r.isCtpReferencingSecret(&nsName) {
			return true
		}
	}

	if r.isOIDCHMACSecret(&nsName) {
		return true
	}

	if r.isEnvoyTLSSecret(&nsName) {
		return true
	}

	if r.epCRDExists {
		if r.isEnvoyProxyReferencingSecret(&nsName) {
			return true
		}
	}

	if r.eepCRDExists {
		if r.isExtensionPolicyReferencingSecret(&nsName) {
			return true
		}
	}

	if r.bTLSPolicyCRDExists {
		if r.isBackendTLSPolicyReferencingSecret(&nsName) {
			return true
		}
	}

	if r.hrfCRDExists {
		if r.isHTTPRouteFilterReferencingSecret(&nsName) {
			return true
		}
	}

	return false
}

func (r *gatewayAPIReconciler) isHTTPRouteFilterReferencingSecret(nsName *types.NamespacedName) bool {
	routeFilterList := &egv1a1.HTTPRouteFilterList{}
	if err := r.client.List(context.Background(), routeFilterList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(secretHTTPRouteFilterIndex, nsName.String()),
	}); err != nil {
		r.log.Error(err, "unable to find associated HTTPRouteFilter")
		return false
	}

	if len(routeFilterList.Items) > 0 {
		return true
	}

	return true
}

func (r *gatewayAPIReconciler) isBackendTLSPolicyReferencingSecret(nsName *types.NamespacedName) bool {
	btlsList := &gwapiv1a3.BackendTLSPolicyList{}
	if err := r.client.List(context.Background(), btlsList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(secretBtlsIndex, nsName.String()),
	}); err != nil {
		r.log.Error(err, "unable to find associated BackendTLSPolicy")
		return false
	}

	if len(btlsList.Items) > 0 {
		return true
	}

	return false
}

func (r *gatewayAPIReconciler) isEnvoyProxyReferencingSecret(nsName *types.NamespacedName) bool {
	epList := &egv1a1.EnvoyProxyList{}
	if err := r.client.List(context.Background(), epList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(secretEnvoyProxyIndex, nsName.String()),
	}); err != nil {
		r.log.Error(err, "unable to find associated Gateways")
		return false
	}

	if len(epList.Items) == 0 {
		return false
	}

	for _, ep := range epList.Items {
		if ep.Spec.BackendTLS != nil {
			if ep.Spec.BackendTLS.ClientCertificateRef != nil {
				certRef := ep.Spec.BackendTLS.ClientCertificateRef
				ns := gatewayapi.NamespaceDerefOr(certRef.Namespace, ep.Namespace)
				if nsName.Name == string(certRef.Name) && nsName.Namespace == ns {
					return true
				}
				continue
			}
		}
	}

	return false
}

func (r *gatewayAPIReconciler) isGatewayReferencingSecret(nsName *types.NamespacedName) bool {
	gwList := &gwapiv1.GatewayList{}
	if err := r.client.List(context.Background(), gwList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(secretGatewayIndex, nsName.String()),
	}); err != nil {
		r.log.Error(err, "unable to find associated Gateways")
		return false
	}

	if len(gwList.Items) == 0 {
		return false
	}

	for _, gw := range gwList.Items {
		if !r.validateGatewayForReconcile(&gw) {
			return false
		}
	}
	return true
}

func (r *gatewayAPIReconciler) isSecurityPolicyReferencingSecret(nsName *types.NamespacedName) bool {
	spList := &egv1a1.SecurityPolicyList{}
	if err := r.client.List(context.Background(), spList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(secretSecurityPolicyIndex, nsName.String()),
	}); err != nil {
		r.log.Error(err, "unable to find associated SecurityPolicies")
		return false
	}

	return len(spList.Items) > 0
}

func (r *gatewayAPIReconciler) isCtpReferencingSecret(nsName *types.NamespacedName) bool {
	ctpList := &egv1a1.ClientTrafficPolicyList{}
	if err := r.client.List(context.Background(), ctpList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(secretCtpIndex, nsName.String()),
	}); err != nil {
		r.log.Error(err, "unable to find associated ClientTrafficPolicies")
		return false
	}

	return len(ctpList.Items) > 0
}

func (r *gatewayAPIReconciler) isOIDCHMACSecret(nsName *types.NamespacedName) bool {
	oidcHMACSecret := types.NamespacedName{
		Namespace: r.namespace,
		Name:      oidcHMACSecretName,
	}
	return *nsName == oidcHMACSecret
}

func (r *gatewayAPIReconciler) isEnvoyTLSSecret(nsName *types.NamespacedName) bool {
	envoyTLSSecret := types.NamespacedName{
		Namespace: r.namespace,
		Name:      envoyTLSSecretName,
	}
	return *nsName == envoyTLSSecret
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
	labels := svc.GetLabels()

	// Check if the Service belongs to a Gateway, if so, update the Gateway status.
	gtw := r.findOwningGateway(ctx, labels)
	if gtw != nil {
		r.updateGatewayStatus(gtw)
		return true
	}

	// Merged gateways will have only this label, update status of all Gateways under found GatewayClass.
	gcName, ok := labels[gatewayapi.OwningGatewayClassLabel]
	if ok && r.mergeGateways.Has(gcName) {
		if err := r.updateStatusForGatewaysUnderGatewayClass(ctx, gcName); err != nil {
			r.log.Info("no Gateways found under GatewayClass", "name", gcName)
			return false
		}
		return false
	}

	nsName := utils.NamespacedName(svc)
	if r.isRouteReferencingBackend(&nsName) {
		return true
	}

	if r.spCRDExists {
		if r.isSecurityPolicyReferencingBackend(&nsName) {
			return true
		}
	}

	if r.epCRDExists {
		if r.isEnvoyProxyReferencingBackend(&nsName) {
			return true
		}
	}

	if r.eepCRDExists {
		if r.isEnvoyExtensionPolicyReferencingBackend(&nsName) {
			return true
		}
	}

	return false
}

// validateBackendForReconcile tries finding the owning Gateway of the Backend
// if it exists, finds the Gateway's Deployment, and further updates the Gateway
// status Ready condition. All Services are pushed for reconciliation.
func (r *gatewayAPIReconciler) validateBackendForReconcile(obj client.Object) bool {
	be, ok := obj.(*egv1a1.Backend)
	if !ok {
		r.log.Info("unexpected object type, bypassing reconciliation", "object", obj)
		return false
	}

	nsName := utils.NamespacedName(be)
	if r.isRouteReferencingBackend(&nsName) {
		return true
	}

	if r.spCRDExists {
		if r.isSecurityPolicyReferencingBackend(&nsName) {
			return true
		}
	}

	if r.epCRDExists {
		if r.isEnvoyProxyReferencingBackend(&nsName) {
			return true
		}
	}

	if r.eepCRDExists {
		if r.isEnvoyExtensionPolicyReferencingBackend(&nsName) {
			return true
		}
	}

	if r.isProxyInfraService(&nsName) {
		return true
	}

	return false
}

func (r *gatewayAPIReconciler) isSecurityPolicyReferencingBackend(nsName *types.NamespacedName) bool {
	spList := &egv1a1.SecurityPolicyList{}
	if err := r.client.List(context.Background(), spList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(backendSecurityPolicyIndex, nsName.String()),
	}); err != nil {
		r.log.Error(err, "unable to find associated SecurityPolicies")
		return false
	}

	return len(spList.Items) > 0
}

// validateServiceImportForReconcile tries finding the owning Gateway of the ServiceImport
// if it exists, finds the Gateway's Deployment, and further updates the Gateway
// status Ready condition. All Services are pushed for reconciliation.
func (r *gatewayAPIReconciler) validateServiceImportForReconcile(obj client.Object) bool {
	svcImport, ok := obj.(*mcsapiv1a1.ServiceImport)
	if !ok {
		r.log.Info("unexpected object type, bypassing reconciliation", "object", obj)
		return false
	}

	nsName := utils.NamespacedName(svcImport)
	return r.isRouteReferencingBackend(&nsName)
}

// isRouteReferencingBackend returns true if the backend(service and serviceImport) is referenced by any of the xRoutes
// in the system, else returns false.
func (r *gatewayAPIReconciler) isRouteReferencingBackend(nsName *types.NamespacedName) bool {
	ctx := context.Background()
	httpRouteList := &gwapiv1.HTTPRouteList{}
	if err := r.client.List(ctx, httpRouteList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(backendHTTPRouteIndex, nsName.String()),
	}); err != nil && !kerrors.IsNotFound(err) {
		r.log.Error(err, "failed to find associated HTTPRoutes")
		return false
	}
	if len(httpRouteList.Items) > 0 {
		return true
	}

	if r.grpcRouteCRDExists {
		grpcRouteList := &gwapiv1.GRPCRouteList{}
		if err := r.client.List(ctx, grpcRouteList, &client.ListOptions{
			FieldSelector: fields.OneTermEqualSelector(backendGRPCRouteIndex, nsName.String()),
		}); err != nil && !kerrors.IsNotFound(err) {
			r.log.Error(err, "failed to find associated GRPCRoutes")
			return false
		}
		if len(grpcRouteList.Items) > 0 {
			return true
		}
	}

	if r.tlsRouteCRDExists {
		tlsRouteList := &gwapiv1a2.TLSRouteList{}
		if err := r.client.List(ctx, tlsRouteList, &client.ListOptions{
			FieldSelector: fields.OneTermEqualSelector(backendTLSRouteIndex, nsName.String()),
		}); err != nil && !kerrors.IsNotFound(err) {
			r.log.Error(err, "failed to find associated TLSRoutes")
			return false
		}
		if len(tlsRouteList.Items) > 0 {
			return true
		}
	}

	if r.tcpRouteCRDExists {
		tcpRouteList := &gwapiv1a2.TCPRouteList{}
		if err := r.client.List(ctx, tcpRouteList, &client.ListOptions{
			FieldSelector: fields.OneTermEqualSelector(backendTCPRouteIndex, nsName.String()),
		}); err != nil && !kerrors.IsNotFound(err) {
			r.log.Error(err, "failed to find associated TCPRoutes")
			return false
		}
		if len(tcpRouteList.Items) > 0 {
			return true
		}
	}

	if r.udpRouteCRDExists {
		udpRouteList := &gwapiv1a2.UDPRouteList{}
		if err := r.client.List(ctx, udpRouteList, &client.ListOptions{
			FieldSelector: fields.OneTermEqualSelector(backendUDPRouteIndex, nsName.String()),
		}); err != nil && !kerrors.IsNotFound(err) {
			r.log.Error(err, "failed to find associated UDPRoutes")
			return false
		}
		if len(udpRouteList.Items) > 0 {
			return true
		}
	}

	return false
}

// validateEndpointSliceForReconcile returns true if the endpointSlice references
// a service that is referenced by a xRoute
func (r *gatewayAPIReconciler) validateEndpointSliceForReconcile(obj client.Object) bool {
	ep, ok := obj.(*discoveryv1.EndpointSlice)
	if !ok {
		r.log.Info("unexpected object type, bypassing reconciliation", "object", obj)
		return false
	}

	svcName, ok := ep.GetLabels()[discoveryv1.LabelServiceName]
	multiClusterSvcName, isMCS := ep.GetLabels()[mcsapiv1a1.LabelServiceName]
	if !ok && !isMCS {
		r.log.Info("endpointslice is missing kubernetes.io/service-name or multicluster.kubernetes.io/service-name label", "object", obj)
		return false
	}

	nsName := types.NamespacedName{
		Namespace: obj.GetNamespace(),
		Name:      svcName,
	}

	if isMCS {
		nsName.Name = multiClusterSvcName
	}

	if r.isRouteReferencingBackend(&nsName) {
		return true
	}

	if r.spCRDExists {
		if r.isSecurityPolicyReferencingBackend(&nsName) {
			return true
		}
	}

	if r.epCRDExists {
		if r.isEnvoyProxyReferencingBackend(&nsName) {
			return true
		}
	}

	if r.eepCRDExists {
		if r.isEnvoyExtensionPolicyReferencingBackend(&nsName) {
			return true
		}
	}

	return false
}

// validateObjectForReconcile tries finding the owning Gateway of the Deployment or DaemonSet
// if it exists, finds the Gateway's Service, and further updates the Gateway
// status Ready condition. No Deployments or DaemonSets are pushed for reconciliation.
func (r *gatewayAPIReconciler) validateObjectForReconcile(obj client.Object) bool {
	ctx := context.Background()
	labels := obj.GetLabels()

	// Only objects in the configured namespace should be reconciled.
	if obj.GetNamespace() == r.namespace || r.gatewayNamespaceMode {
		// Check if the obj belongs to a Gateway, if so, update the Gateway status.
		gtw := r.findOwningGateway(ctx, labels)
		if gtw != nil {
			r.updateGatewayStatus(gtw)
			return false
		}
	}

	// Merged gateways will have only this label, update status of all Gateways under found GatewayClass.
	gcName, ok := labels[gatewayapi.OwningGatewayClassLabel]
	if ok && r.mergeGateways.Has(gcName) {
		if err := r.updateStatusForGatewaysUnderGatewayClass(ctx, gcName); err != nil {
			r.log.Info("no Gateways found under GatewayClass", "name", gcName)
			return false
		}
		return true
	}

	// There is no need to reconcile the object any further.
	return false
}

func envoyObjectNamespace(r *gatewayAPIReconciler, gateway *gwapiv1.Gateway) string {
	if r.gatewayNamespaceMode {
		return gateway.Namespace
	}
	return r.namespace
}

// envoyObjectForGateway returns the Envoy Deployment or DaemonSet, returning nil if neither exists.
func (r *gatewayAPIReconciler) envoyObjectForGateway(ctx context.Context, gateway *gwapiv1.Gateway) (client.Object, error) {
	// Helper func to list and return the first object from results
	listResource := func(list client.ObjectList) (client.Object, error) {
		if err := r.client.List(ctx, list, &client.ListOptions{
			LabelSelector: labels.SelectorFromSet(gatewayapi.OwnerLabels(gateway, r.mergeGateways.Has(string(gateway.Spec.GatewayClassName)))),
			Namespace:     envoyObjectNamespace(r, gateway),
		}); err != nil {
			if !kerrors.IsNotFound(err) {
				return nil, err
			}
		}
		items, err := meta.ExtractList(list)
		if err != nil || len(items) == 0 {
			return nil, nil
		}
		return items[0].(client.Object), nil
	}

	// Check for Deployment
	deployments := &appsv1.DeploymentList{}
	if obj, err := listResource(deployments); obj != nil || err != nil {
		return obj, err
	}

	// Check for DaemonSet
	daemonsets := &appsv1.DaemonSetList{}
	if obj, err := listResource(daemonsets); obj != nil || err != nil {
		return obj, err
	}

	return nil, nil
}

// envoyServiceForGateway returns the Envoy service, returning nil if the service doesn't exist.
func (r *gatewayAPIReconciler) envoyServiceForGateway(ctx context.Context, gateway *gwapiv1.Gateway) (*corev1.Service, error) {
	var services corev1.ServiceList
	labelSelector := labels.SelectorFromSet(labels.Set(gatewayapi.OwnerLabels(gateway, r.mergeGateways.Has(string(gateway.Spec.GatewayClassName)))))
	if err := r.client.List(ctx, &services, &client.ListOptions{
		LabelSelector: labelSelector,
		Namespace:     envoyObjectNamespace(r, gateway),
	}); err != nil {
		if kerrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	if len(services.Items) == 0 {
		return nil, nil
	}
	return &services.Items[0], nil
}

// findOwningGateway attempts finds a Gateway using "labels".
func (r *gatewayAPIReconciler) findOwningGateway(ctx context.Context, labels map[string]string) *gwapiv1.Gateway {
	gwName, ok := labels[gatewayapi.OwningGatewayNameLabel]
	if !ok {
		return nil
	}

	gwNamespace, ok := labels[gatewayapi.OwningGatewayNamespaceLabel]
	if !ok {
		return nil
	}

	gatewayKey := types.NamespacedName{Namespace: gwNamespace, Name: gwName}
	gtw := new(gwapiv1.Gateway)
	if err := r.client.Get(ctx, gatewayKey, gtw); err != nil {
		r.log.Info("gateway not found", "namespace", gtw.Namespace, "name", gtw.Name)
		return nil
	}

	return gtw
}

// updateStatusForGatewaysUnderGatewayClass updates status of all Gateways under the GatewayClass.
func (r *gatewayAPIReconciler) updateStatusForGatewaysUnderGatewayClass(ctx context.Context, gatewayClassName string) error {
	gateways := new(gwapiv1.GatewayList)
	if err := r.client.List(ctx, gateways, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(classGatewayIndex, gatewayClassName),
	}); err != nil {
		return err
	}

	if len(gateways.Items) == 0 {
		return fmt.Errorf("no gateways found for gatewayclass: %s", gatewayClassName)
	}

	for _, gateway := range gateways.Items {
		r.updateGatewayStatus(&gateway)
	}

	return nil
}

// updateGatewayStatus triggers a status update for the Gateway.
func (r *gatewayAPIReconciler) updateGatewayStatus(gateway *gwapiv1.Gateway) {
	gwName := utils.NamespacedName(gateway)
	status := &gateway.Status
	// Use the existing status if it exists to avoid losing the status calculated by the Gateway API translator.
	if existing, ok := r.resources.GatewayStatuses.Load(gwName); ok {
		status = existing
	}

	// Since the status does not reflect the actual changed status, we need to delete it first
	// to prevent it from being considered unchanged. This ensures that subscribers receive the update event.
	r.resources.GatewayStatuses.Delete(gwName)
	// The status that is stored in the GatewayStatuses GatewayStatuses is solely used to trigger the status updater
	// and does not reflect the real changed status.
	//
	// The status updater will check the Envoy Proxy service to get the addresses of the Gateway,
	// and check the Envoy Proxy Deployment/DaemonSet to get the status of the Gateway workload.
	r.resources.GatewayStatuses.Store(gwName, status)
}

func (r *gatewayAPIReconciler) handleNode(obj client.Object) bool {
	ctx := context.Background()
	node, ok := obj.(*corev1.Node)
	if !ok {
		r.log.Info("unexpected object type, bypassing reconciliation", "object", obj)
		return false
	}

	key := types.NamespacedName{Name: node.Name}
	if err := r.client.Get(ctx, key, node); err != nil {
		if kerrors.IsNotFound(err) {
			r.store.removeNode(node)
			return true
		}
		r.log.Error(err, "unable to find node", "name", node.Name)
		return false
	}

	r.store.addNode(node)
	return true
}

// validateConfigMapForReconcile checks whether the ConfigMap belongs to a valid EG resource.
func (r *gatewayAPIReconciler) validateConfigMapForReconcile(obj client.Object) bool {
	configMap, ok := obj.(*corev1.ConfigMap)
	if !ok {
		r.log.Info("unexpected object type, bypassing reconciliation", "object", obj)
		return false
	}

	if r.ctpCRDExists {
		ctpList := &egv1a1.ClientTrafficPolicyList{}
		if err := r.client.List(context.Background(), ctpList, &client.ListOptions{
			FieldSelector: fields.OneTermEqualSelector(configMapCtpIndex, utils.NamespacedName(configMap).String()),
		}); err != nil {
			r.log.Error(err, "unable to find associated ClientTrafficPolicy")
			return false
		}

		if len(ctpList.Items) > 0 {
			return true
		}
	}

	if r.bTLSPolicyCRDExists {
		btlsList := &gwapiv1a3.BackendTLSPolicyList{}
		if err := r.client.List(context.Background(), btlsList, &client.ListOptions{
			FieldSelector: fields.OneTermEqualSelector(configMapBtlsIndex, utils.NamespacedName(configMap).String()),
		}); err != nil {
			r.log.Error(err, "unable to find associated BackendTLSPolicy")
			return false
		}

		if len(btlsList.Items) > 0 {
			return true
		}
	}

	if r.btpCRDExists {
		btpList := &egv1a1.BackendTrafficPolicyList{}
		if err := r.client.List(context.Background(), btpList, &client.ListOptions{
			FieldSelector: fields.OneTermEqualSelector(configMapBtpIndex, utils.NamespacedName(configMap).String()),
		}); err != nil {
			r.log.Error(err, "unable to find associated BackendTrafficPolicy")
			return false
		}

		if len(btpList.Items) > 0 {
			return true
		}
	}

	if r.eepCRDExists {
		eepList := &egv1a1.EnvoyExtensionPolicyList{}
		if err := r.client.List(context.Background(), eepList, &client.ListOptions{
			FieldSelector: fields.OneTermEqualSelector(configMapEepIndex, utils.NamespacedName(configMap).String()),
		}); err != nil {
			r.log.Error(err, "unable to find associated EnvoyExtensionPolicy")
			return false
		}

		if len(eepList.Items) > 0 {
			return true
		}
	}

	if r.hrfCRDExists {
		routeFilterList := &egv1a1.HTTPRouteFilterList{}
		if err := r.client.List(context.Background(), routeFilterList, &client.ListOptions{
			FieldSelector: fields.OneTermEqualSelector(configMapHTTPRouteFilterIndex, utils.NamespacedName(configMap).String()),
		}); err != nil {
			r.log.Error(err, "unable to find associated HTTPRouteFilter")
			return false
		}

		if len(routeFilterList.Items) > 0 {
			return true
		}
	}

	if r.spCRDExists {
		spList := &egv1a1.SecurityPolicyList{}
		if err := r.client.List(context.Background(), spList, &client.ListOptions{
			FieldSelector: fields.OneTermEqualSelector(configMapSecurityPolicyIndex, utils.NamespacedName(configMap).String()),
		}); err != nil {
			r.log.Error(err, "unable to find associated SecurityPolicy")
			return false
		}

		if len(spList.Items) > 0 {
			return true
		}
	}

	return false
}

func (r *gatewayAPIReconciler) isEnvoyExtensionPolicyReferencingBackend(nsName *types.NamespacedName) bool {
	spList := &egv1a1.EnvoyExtensionPolicyList{}
	if err := r.client.List(context.Background(), spList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(backendEnvoyExtensionPolicyIndex, nsName.String()),
	}); err != nil {
		r.log.Error(err, "unable to find associated EnvoyExtensionPolicies")
		return false
	}

	return len(spList.Items) > 0
}

func (r *gatewayAPIReconciler) isEnvoyProxyReferencingBackend(nn *types.NamespacedName) bool {
	proxyList := &egv1a1.EnvoyProxyList{}
	if err := r.client.List(context.Background(), proxyList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(backendEnvoyProxyTelemetryIndex, nn.String()),
	}); err != nil {
		r.log.Error(err, "unable to find associated EnvoyProxies")
		return false
	}

	return len(proxyList.Items) > 0
}

func (r *gatewayAPIReconciler) isExtensionPolicyReferencingSecret(nsName *types.NamespacedName) bool {
	eepList := &egv1a1.EnvoyExtensionPolicyList{}
	if err := r.client.List(context.Background(), eepList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(secretEnvoyExtensionPolicyIndex, nsName.String()),
	}); err != nil {
		r.log.Error(err, "unable to find associated ExtensionPolicies")
		return false
	}

	return len(eepList.Items) > 0
}

// isRouteReferencingHTTPRouteFilter returns true if the HTTPRouteFilter is referenced by an HTTPRoute
func (r *gatewayAPIReconciler) isRouteReferencingHTTPRouteFilter(nsName *types.NamespacedName) bool {
	ctx := context.Background()
	httpRouteList := &gwapiv1.HTTPRouteList{}
	if err := r.client.List(ctx, httpRouteList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(httpRouteFilterHTTPRouteIndex, nsName.String()),
	}); err != nil {
		r.log.Error(err, "unable to find associated HTTPRoutes")
		return false
	}

	return len(httpRouteList.Items) != 0
}

func (r *gatewayAPIReconciler) isProxyInfraService(nn *types.NamespacedName) bool {
	ctx := context.Background()
	svc := &corev1.Service{}
	if err := r.client.Get(ctx, *nn, svc); err != nil {
		r.log.Error(err, "unable to find associated ProxyInfra service")
		return false
	}

	svcLabels := svc.GetLabels()

	// Check if the Service belongs to a Gateway, if so, update the Gateway status.
	if gtw := r.findOwningGateway(ctx, svcLabels); gtw != nil {
		return true
	}

	// Merged gateways will have only this label, update status of all Gateways under found GatewayClass.
	gcName, ok := svcLabels[gatewayapi.OwningGatewayClassLabel]
	if ok && r.mergeGateways.Has(gcName) {
		return true
	}
	return false
}

// validateHTTPRouteFilterForReconcile tries finding the referencing HTTPRoute of the filter
func (r *gatewayAPIReconciler) validateHTTPRouteFilterForReconcile(obj client.Object) bool {
	hrf, ok := obj.(*egv1a1.HTTPRouteFilter)
	if !ok {
		r.log.Info("unexpected object type, bypassing reconciliation", "object", obj)
		return false
	}

	nsName := utils.NamespacedName(hrf)
	return r.isRouteReferencingHTTPRouteFilter(&nsName)
}

func commonPredicates[T client.Object]() []predicate.TypedPredicate[T] {
	return []predicate.TypedPredicate[T]{
		metadataPredicate[T](),
	}
}

func metadataPredicate[T client.Object]() predicate.TypedPredicate[T] {
	return predicate.Or(predicate.TypedGenerationChangedPredicate[T]{},
		predicate.TypedLabelChangedPredicate[T]{},
		predicate.TypedAnnotationChangedPredicate[T]{},
	)
}
