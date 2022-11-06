// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"fmt"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/provider/utils"

	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

// HTTPRoute processing

// processHTTPRoute processes the HTTPRoute coming from the watcher and further
// processes parent Gateway objects to eventually reconcile GatewayClass.
func (r *gatewayAPIReconciler) processHTTPRoute(obj client.Object) []reconcile.Request {
	r.log.Info("processing httproute", "namespace", obj.GetNamespace(), "name", obj.GetName())
	ctx := context.Background()
	requests := []reconcile.Request{}

	hrkey := types.NamespacedName{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}

	httpRouteParentReferences := []gwapiv1b1.ParentReference{}
	routeDeleted := false

	httproute := new(gwapiv1b1.HTTPRoute)
	if err := r.client.Get(ctx, hrkey, httproute); err != nil {
		if !kerrors.IsNotFound(err) {
			r.log.Error(err, "failed to get httproute")
			return requests
		}

		routeDeleted = true
		// Remove the HTTPRoute from watchable map.
		if resourceRoute, found := r.resources.HTTPRoutes.Load(hrkey); found {
			httpRouteParentReferences = append(httpRouteParentReferences, resourceRoute.Spec.ParentRefs...)
			r.resources.HTTPRoutes.Delete(hrkey)
			r.log.Info("deleted httproute from watchable map")
		}

		// Delete the Namespace and Service from the watchable maps if no other
		// routes (TLSRoute or HTTPRoute) exist in the namespace.
		found, err := isRoutePresentInNamespace(ctx, r.client, hrkey.Namespace)
		if err != nil {
			return requests
		}
		if !found {
			r.resources.Namespaces.Delete(hrkey.Namespace)
			r.log.Info("deleted namespace from watchable map")
		}

		// Delete the Service from the watchable maps if no other
		// routes (TLSRoute or HTTPRoute) reference that Service.
		routeServices := r.referenceStore.getRouteToServicesMapping(ObjectKindNamespacedName{kindHTTPRoute, hrkey.Namespace, hrkey.Name})
		for svc := range routeServices {
			r.referenceStore.removeRouteToServicesMapping(ObjectKindNamespacedName{kindHTTPRoute, hrkey.Namespace, hrkey.Name}, svc)
			if !r.referenceStore.isServiceReferredByRoutes(svc) {
				r.resources.Services.Delete(svc)
				r.log.Info("deleted service from watchable map", "namespace", svc.Namespace, "name", svc.Name)
			}
		}
	}

	if !routeDeleted {
		v, ok := r.resources.HTTPRoutes.Load(hrkey)
		// Donot process further if the resource is unchanged.
		if ok && httproute.Generation == v.Generation {
			return requests
		}
		if !ok {
			r.resources.HTTPRoutes.Store(hrkey, httproute)
			r.log.Info("added httproute to watchable map")
		}

		// Get the route's namespace from the cache.
		if err := r.checkNamespaceForRoute(ctx, hrkey.Namespace); err != nil {
			return requests
		}

		// Get the route's backendRefs from the cache. Note that a Service is the
		// only supported kind.
		routeBackendReferences := []gwapiv1b1.BackendRef{}
		for i := range httproute.Spec.Rules {
			for j := range httproute.Spec.Rules[i].BackendRefs {
				ref := httproute.Spec.Rules[i].BackendRefs[j].BackendRef
				routeBackendReferences = append(routeBackendReferences, ref)
			}
		}
		if err := r.checkAndValidateRouteBackendRefs(ctx, kindHTTPRoute, hrkey, routeBackendReferences); err != nil {
			return requests
		}

		httpRouteParentReferences = append(httpRouteParentReferences, httproute.Spec.ParentRefs...)
	}

	// Find the parent Gateway objects for httproute.
	gateways, err := validateParentRefs(ctx, r.client, hrkey.Namespace, r.classController, httpRouteParentReferences)
	if err != nil {
		r.log.Info("invalid parentRefs for httproute, bypassing reconciliation", "object", obj)
		return requests
	}

	// Remove the route from the watchable map since it doesn't reference
	// a managed Gateway.
	if len(gateways) == 0 {
		r.log.Info("httproute doesn't reference any managed gateways")
		r.resources.HTTPRoutes.Delete(hrkey)
		return requests
	}

	for j := range gateways {
		requests = append(requests, r.processGateway(gateways[j].DeepCopy())...)
	}

	return requests
}

// addHTTPRouteIndexers adds indexing on HTTPRoute, for Service objects that are
// referenced in HTTPRoute objects via `.spec.rules.backendRefs`. This helps in
// querying for HTTPRoutes that are affected by a particular Service CRUD.
func addHTTPRouteIndexers(ctx context.Context, mgr manager.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(ctx, &gwapiv1b1.HTTPRoute{}, serviceHTTPRouteIndex, func(rawObj client.Object) []string {
		httpRoute := rawObj.(*gwapiv1b1.HTTPRoute)
		var backendServices []string
		for _, rule := range httpRoute.Spec.Rules {
			for _, backend := range rule.BackendRefs {
				if string(*backend.Kind) == gatewayapi.KindService {
					// If an explicit Service namespace is not provided, use the HTTPRoute namespace to
					// lookup the provided Service Name.
					backendServices = append(backendServices,
						types.NamespacedName{
							Namespace: gatewayapi.NamespaceDerefOr(backend.Namespace, httpRoute.Namespace),
							Name:      string(backend.Name),
						}.String(),
					)
				}
			}
		}
		return backendServices
	}); err != nil {
		return err
	}
	return nil
}

// TLSRoute processing

// processTLSRoute processes the TLSRoute coming from the watcher and further
// processes parent Gateway objects to eventually reconcile GatewayClass.
func (r *gatewayAPIReconciler) processTLSRoute(obj client.Object) []reconcile.Request {
	r.log.Info("processing tlsroute", "namespace", obj.GetNamespace(), "name", obj.GetName())
	ctx := context.Background()
	requests := []reconcile.Request{}

	trkey := types.NamespacedName{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}

	tlsRouteParentReferences := []gwapiv1b1.ParentReference{}
	routeDeleted := false

	tlsroute := new(gwapiv1a2.TLSRoute)
	if err := r.client.Get(ctx, trkey, tlsroute); err != nil {
		if !kerrors.IsNotFound(err) {
			r.log.Error(err, "failed to get tlsroute")
			return requests
		}

		routeDeleted = true
		// Remove the TLSRoute from watchable map.
		if resourceRoute, found := r.resources.TLSRoutes.Load(trkey); found {
			tlsRouteParentReferences = append(tlsRouteParentReferences, gatewayapi.UpgradeParentReferences(resourceRoute.Spec.ParentRefs)...)
			r.resources.TLSRoutes.Delete(trkey)
			r.log.Info("deleted tlsroute from watchable map")
		}

		// Delete the Namespace and Service from the watchable maps if no other
		// routes (TLSRoute or HTTPRoute) exist in the namespace.
		found, err := isRoutePresentInNamespace(ctx, r.client, trkey.Namespace)
		if err != nil {
			return requests
		}
		if !found {
			r.resources.Namespaces.Delete(trkey.Namespace)
			r.log.Info("deleted namespace from watchable map")
		}

		// Delete the Service from the watchable maps if no other
		// routes (TLSRoute or HTTPRoute) reference that Service.
		routeServices := r.referenceStore.getRouteToServicesMapping(ObjectKindNamespacedName{kindTLSRoute, trkey.Namespace, trkey.Name})
		for svc := range routeServices {
			r.referenceStore.removeRouteToServicesMapping(ObjectKindNamespacedName{kindTLSRoute, trkey.Namespace, trkey.Name}, svc)
			if !r.referenceStore.isServiceReferredByRoutes(svc) {
				r.resources.Services.Delete(svc)
				r.log.Info("deleted service from watchable map", "namespace", svc.Namespace, "name", svc.Name)
			}
		}
	}

	fmt.Printf("heyoo %v\n", routeDeleted)

	if !routeDeleted {
		v, ok := r.resources.TLSRoutes.Load(trkey)
		// Donot process further if the resource is unchanged.
		if ok && tlsroute.Generation == v.Generation {
			return requests
		}

		if !ok {
			r.resources.TLSRoutes.Store(trkey, tlsroute)
			r.log.Info("added tlsroute to watchable map")
		}

		// Get the route's namespace from the cache.
		if err := r.checkNamespaceForRoute(ctx, trkey.Namespace); err != nil {
			return requests
		}

		// Get the route's backendRefs from the cache. Note that a Service is the
		// only supported kind.
		routeBackendReferences := []gwapiv1b1.BackendRef{}
		for i := range tlsroute.Spec.Rules {
			for j := range tlsroute.Spec.Rules[i].BackendRefs {
				ref := gatewayapi.UpgradeBackendRef(tlsroute.Spec.Rules[i].BackendRefs[j])
				routeBackendReferences = append(routeBackendReferences, ref)
			}
		}
		if err := r.checkAndValidateRouteBackendRefs(ctx, kindTLSRoute, trkey, routeBackendReferences); err != nil {
			return requests
		}

		tlsRouteParentReferences = append(tlsRouteParentReferences, gatewayapi.UpgradeParentReferences(tlsroute.Spec.ParentRefs)...)
	}

	// Find the parent Gateway objects for tlsroute.
	gateways, err := validateParentRefs(ctx, r.client, trkey.Namespace, r.classController, tlsRouteParentReferences)
	if err != nil {
		r.log.Info("invalid parentRefs for tlsroute, bypassing reconciliation", "object", obj)
		return requests
	}

	// Remove the route from the watchable map since it doesn't reference
	// a managed Gateway.
	if len(gateways) == 0 {
		r.log.Info("tlsroute doesn't reference any managed gateways")
		r.resources.TLSRoutes.Delete(trkey)
		return requests
	}

	for j := range gateways {
		requests = append(requests, r.processGateway(gateways[j].DeepCopy())...)
	}

	return requests
}

// addTLSRouteIndexers adds indexing on TLSRoute, for Service objects that are
// referenced in TLSRoute objects via `.spec.rules.backendRefs`. This helps in
// querying for TLSRoutes that are affected by a particular Service CRUD.
func addTLSRouteIndexers(ctx context.Context, mgr manager.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(ctx, &gwapiv1a2.TLSRoute{}, serviceTLSRouteIndex, func(rawObj client.Object) []string {
		tlsRoute := rawObj.(*gwapiv1a2.TLSRoute)
		var backendServices []string
		for _, rule := range tlsRoute.Spec.Rules {
			for _, backend := range rule.BackendRefs {
				if string(*backend.Kind) == gatewayapi.KindService {
					// If an explicit Service namespace is not provided, use the TLSRoute namespace to
					// lookup the provided Service Name.
					backendServices = append(backendServices,
						types.NamespacedName{
							Namespace: gatewayapi.NamespaceDerefOrAlpha(backend.Namespace, tlsRoute.Namespace),
							Name:      string(backend.Name),
						}.String(),
					)
				}
			}
		}
		return backendServices
	}); err != nil {
		return err
	}
	return nil
}

// Common Route processing functions

// processService processes the Service coming from the watcher and further
// processes parent Route objects to eventually reconcile GatewayClass.
func (r *gatewayAPIReconciler) processService(obj client.Object) []reconcile.Request {
	r.log.Info("processing service", "namespace", obj.GetNamespace(), "name", obj.GetName())
	ctx := context.Background()
	requests := []reconcile.Request{}

	svckey := types.NamespacedName{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}

	serviceDeleted := false

	service := new(corev1.Service)
	if err := r.client.Get(ctx, svckey, service); err != nil {
		if !kerrors.IsNotFound(err) {
			r.log.Error(err, "failed to get service")
			return requests
		}

		serviceDeleted = true
		// Remove the Service from watchable map.
		r.resources.Services.Delete(svckey)
	}

	if !serviceDeleted {
		// Store the Service in watchable map.
		r.resources.Services.Store(svckey, service.DeepCopy())
	}

	// Find the HTTPRoutes that reference this Service.
	httpRouteList := &gwapiv1b1.HTTPRouteList{}
	if err := r.client.List(context.Background(), httpRouteList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(serviceHTTPRouteIndex, utils.NamespacedName(obj).String()),
	}); err != nil {
		return []reconcile.Request{}
	}

	// Find the TLSRoutes that reference this Service.
	tlsRouteList := &gwapiv1a2.TLSRouteList{}
	if err := r.client.List(context.Background(), tlsRouteList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(serviceTLSRouteIndex, utils.NamespacedName(obj).String()),
	}); err != nil {
		return []reconcile.Request{}
	}

	for _, h := range httpRouteList.Items {
		requests = append(requests, r.processHTTPRoute(h.DeepCopy())...)
	}
	for _, t := range tlsRouteList.Items {
		requests = append(requests, r.processTLSRoute(t.DeepCopy())...)
	}

	return requests
}

// checkNamespaceForRoute
func (r *gatewayAPIReconciler) checkNamespaceForRoute(ctx context.Context, name string) error {
	nsKey := types.NamespacedName{Name: name}
	ns := new(corev1.Namespace)
	if err := r.client.Get(ctx, nsKey, ns); err != nil {
		if kerrors.IsNotFound(err) {
			// The route's namespace doesn't exist in the cache, so remove it from
			// the namespace watchable map if it exists.
			if _, ok := r.resources.Namespaces.Load(nsKey.Name); ok {
				r.resources.Namespaces.Delete(nsKey.Name)
				r.log.Info("deleted namespace from watchable map")
			}
		}
		return fmt.Errorf("failed to get namespace %s", nsKey.Name)
	}

	// The route's namespace exists, so add it to the watchable map.
	r.resources.Namespaces.Store(nsKey.Name, ns)
	r.log.Info("added namespace to watchable map")
	return nil
}

// checkAndValidateRouteBackendRefs
func (r *gatewayAPIReconciler) checkAndValidateRouteBackendRefs(ctx context.Context, routeKind string, routekey types.NamespacedName, backendReferences []gwapiv1b1.BackendRef) error {
	for j := range backendReferences {
		ref := backendReferences[j]
		if err := validateBackendRef(&ref); err != nil {
			return fmt.Errorf("invalid backendRef: %w", err)
		}

		// The backendRef is valid, so get the referenced service from the cache.
		svcKey := types.NamespacedName{Namespace: routekey.Namespace, Name: string(ref.Name)}
		svc := new(corev1.Service)
		if err := r.client.Get(ctx, svcKey, svc); err != nil {
			if kerrors.IsNotFound(err) {
				// The ref's service doesn't exist in the cache, so remove it from
				// the watchable map if it exists.
				if _, ok := r.resources.Services.Load(svcKey); ok {
					r.resources.Services.Delete(svcKey)
					r.referenceStore.removeRouteToServicesMapping(
						ObjectKindNamespacedName{routeKind, routekey.Namespace, routekey.Name},
						svcKey,
					)
					r.log.Info("deleted service from watchable map")
				}
			}
			return fmt.Errorf("failed to get service %s/%s",
				svcKey.Namespace, svcKey.Name)
		}

		// The backendRef Service exists, so add it to the watchable map.
		r.resources.Services.Store(svcKey, svc)
		r.referenceStore.updateRouteToServicesMapping(
			ObjectKindNamespacedName{routeKind, routekey.Namespace, routekey.Name},
			svcKey,
		)
		r.log.Info("added service to watchable map")
	}
	return nil
}
