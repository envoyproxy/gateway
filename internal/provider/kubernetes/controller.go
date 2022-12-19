// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/provider/utils"
	"github.com/envoyproxy/gateway/internal/status"
	"github.com/envoyproxy/gateway/internal/utils/slice"
)

const (
	classGatewayIndex          = "classGatewayIndex"
	gatewayTLSRouteIndex       = "gatewayTLSRouteIndex"
	gatewayHTTPRouteIndex      = "gatewayHTTPRouteIndex"
	secretGatewayIndex         = "secretGatewayIndex"
	targetRefGrantRouteIndex   = "targetRefGrantRouteIndex"
	serviceHTTPRouteIndex      = "serviceHTTPRouteIndex"
	serviceTLSRouteIndex       = "serviceTLSRouteIndex"
	authenFilterHTTPRouteIndex = "authenHTTPRouteIndex"
)

type gatewayAPIReconciler struct {
	client          client.Client
	log             logr.Logger
	statusUpdater   status.Updater
	classController gwapiv1b1.GatewayController
	namespace       string

	resources *message.ProviderResources
}

// newGatewayAPIController
func newGatewayAPIController(mgr manager.Manager, cfg *config.Server, su status.Updater, resources *message.ProviderResources) error {
	ctx := context.Background()

	r := &gatewayAPIReconciler{
		client:          mgr.GetClient(),
		log:             cfg.Logger,
		classController: gwapiv1b1.GatewayController(cfg.EnvoyGateway.Gateway.ControllerName),
		namespace:       cfg.Namespace,
		statusUpdater:   su,
		resources:       resources,
	}

	c, err := controller.New("gatewayapi", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}
	r.log.Info("created gatewayapi controller")

	// Subscribe to status updates
	r.subscribeAndUpdateStatus(ctx)

	// Only enqueue GatewayClass objects that match this Envoy Gateway's controller name.
	if err := c.Watch(
		&source.Kind{Type: &gwapiv1b1.GatewayClass{}},
		&handler.EnqueueRequestForObject{},
		predicate.NewPredicateFuncs(r.hasMatchingController),
	); err != nil {
		return err
	}

	// Watch Gateway CRUDs and reconcile affected GatewayClass.
	if err := c.Watch(
		&source.Kind{Type: &gwapiv1b1.Gateway{}},
		&handler.EnqueueRequestForObject{},
		predicate.NewPredicateFuncs(r.validateGatewayForReconcile),
	); err != nil {
		return err
	}
	if err := addGatewayIndexers(ctx, mgr); err != nil {
		return err
	}

	// Watch HTTPRoute CRUDs and process affected Gateways.
	if err := c.Watch(
		&source.Kind{Type: &gwapiv1b1.HTTPRoute{}},
		&handler.EnqueueRequestForObject{},
	); err != nil {
		return err
	}
	if err := addHTTPRouteIndexers(ctx, mgr); err != nil {
		return err
	}

	// Watch TLSRoute CRUDs and process affected Gateways.
	if err := c.Watch(
		&source.Kind{Type: &gwapiv1a2.TLSRoute{}},
		&handler.EnqueueRequestForObject{},
	); err != nil {
		return err
	}
	if err := addTLSRouteIndexers(ctx, mgr); err != nil {
		return err
	}

	// Watch Service CRUDs and process affected *Route objects.
	if err := c.Watch(
		&source.Kind{Type: &corev1.Service{}},
		&handler.EnqueueRequestForObject{},
		predicate.NewPredicateFuncs(r.validateServiceForReconcile)); err != nil {
		return err
	}

	// Watch Secret CRUDs and process affected Gateways.
	if err := c.Watch(
		&source.Kind{Type: &corev1.Secret{}},
		&handler.EnqueueRequestForObject{},
		predicate.NewPredicateFuncs(r.validateSecretForReconcile),
	); err != nil {
		return err
	}

	// Watch ReferenceGrant CRUDs and process affected Gateways.
	if err := c.Watch(
		&source.Kind{Type: &gwapiv1a2.ReferenceGrant{}},
		&handler.EnqueueRequestForObject{},
	); err != nil {
		return err
	}
	if err := addReferenceGrantIndexers(ctx, mgr); err != nil {
		return err
	}

	// Watch Deployment CRUDs and process affected Gateways.
	if err := c.Watch(
		&source.Kind{Type: &appsv1.Deployment{}},
		&handler.EnqueueRequestForObject{},
		predicate.NewPredicateFuncs(r.validateDeploymentForReconcile),
	); err != nil {
		return err
	}

	// Watch AuthenticationFilter CRUDs and enqueue associated HTTPRoute objects.
	if err := c.Watch(
		&source.Kind{Type: &egv1a1.AuthenticationFilter{}},
		&handler.EnqueueRequestForObject{},
		predicate.NewPredicateFuncs(r.httpRoutesForAuthenticationFilter)); err != nil {
		return err
	}

	r.log.Info("watching gatewayAPI related objects")
	return nil
}

type resourceMappings struct {
	// Map for storing namespaces for Route, Service and Gateway objects.
	allAssociatedNamespaces map[string]struct{}
	// Map for storing service NamespaceNames referred by various Route objects.
	allAssociatedBackendRefs map[types.NamespacedName]struct{}
	// Map for storing referenceGrant NamespaceNames for BackendRefs, SecretRefs.
	allAssociatedRefGrants map[types.NamespacedName]*gwapiv1a2.ReferenceGrant
	// httpRouteToAuthenFilters is a map of httproute to authenticationfilter associations,
	// where the key is the httproute namespaced name.
	httpRouteToAuthenFilters map[types.NamespacedName][]*egv1a1.AuthenticationFilter
}

func newResourceMapping() *resourceMappings {
	return &resourceMappings{
		allAssociatedNamespaces:  map[string]struct{}{},
		allAssociatedBackendRefs: map[types.NamespacedName]struct{}{},
		allAssociatedRefGrants:   map[types.NamespacedName]*gwapiv1a2.ReferenceGrant{},
		httpRouteToAuthenFilters: map[types.NamespacedName][]*egv1a1.AuthenticationFilter{},
	}
}

func (r *gatewayAPIReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	r.log.WithName(request.Name).Info("reconciling gatewayAPI object", "namespace", request.Namespace, "name", request.Name)

	var gatewayClasses gwapiv1b1.GatewayClassList
	if err := r.client.List(ctx, &gatewayClasses); err != nil {
		return reconcile.Result{}, fmt.Errorf("error listing gatewayclasses: %v", err)
	}

	var cc controlledClasses

	for _, gwClass := range gatewayClasses.Items {
		gwClass := gwClass
		if gwClass.Spec.ControllerName == r.classController {
			// The gatewayclass was marked for deletion and the finalizer removed,
			// so clean-up dependents.
			if !gwClass.DeletionTimestamp.IsZero() &&
				!slice.ContainsString(gwClass.Finalizers, gatewayClassFinalizer) {
				r.log.Info("gatewayclass marked for deletion")
				cc.removeMatch(&gwClass)

				// Delete the gatewayclass from the watchable map.
				r.resources.GatewayAPIResources.Delete(gwClass.Name)
				continue
			}

			cc.addMatch(&gwClass)
		}
	}

	// The gatewayclass was already deleted/finalized and there are stale queue entries.
	acceptedGC := cc.acceptedClass()
	if acceptedGC == nil {
		r.log.Info("failed to find an accepted gatewayclass")
		return reconcile.Result{}, nil
	}

	updater := func(gc *gwapiv1b1.GatewayClass, accepted bool) error {
		if r.statusUpdater != nil {
			r.statusUpdater.Send(status.Update{
				NamespacedName: types.NamespacedName{Name: gc.Name},
				Resource:       &gwapiv1b1.GatewayClass{},
				Mutator: status.MutatorFunc(func(obj client.Object) client.Object {
					gc, ok := obj.(*gwapiv1b1.GatewayClass)
					if !ok {
						panic(fmt.Sprintf("unsupported object type %T", obj))
					}

					return status.SetGatewayClassAccepted(gc.DeepCopy(), accepted)
				}),
			})
		} else {
			// this branch makes testing easier by not going through the status.Updater.
			copy := status.SetGatewayClassAccepted(gc.DeepCopy(), accepted)

			if err := r.client.Status().Update(ctx, copy); err != nil && !kerrors.IsNotFound(err) {
				return fmt.Errorf("error updating status of gatewayclass %s: %w", copy.Name, err)
			}
		}
		return nil
	}

	// Update status for all gateway classes
	for _, gc := range cc.notAcceptedClasses() {
		if err := updater(gc, false); err != nil {
			r.resources.GatewayAPIResources.Delete(acceptedGC.Name)
			return reconcile.Result{}, err
		}
	}

	// Initialize resource types.
	resourceTree := gatewayapi.NewResources()
	resourceMap := newResourceMapping()

	// Find gateways for the acceptedGC
	// Find the Gateways that reference this Class.
	gatewayList := &gwapiv1b1.GatewayList{}
	if err := r.client.List(ctx, gatewayList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(classGatewayIndex, acceptedGC.Name),
	}); err != nil {
		r.log.Info("no associated Gateways found for GatewayClass", "name", acceptedGC.Name)
		return reconcile.Result{}, err
	}

	for _, gtw := range gatewayList.Items {
		gtw := gtw
		r.log.Info("processing Gateway", "namespace", gtw.Namespace, "name", gtw.Name)
		resourceMap.allAssociatedNamespaces[gtw.Namespace] = struct{}{}

		for _, listener := range gtw.Spec.Listeners {
			listener := listener
			// Get Secret for gateway if it exists.
			if terminatesTLS(&listener) {
				for _, certRef := range listener.TLS.CertificateRefs {
					certRef := certRef
					if refsSecret(&certRef) {
						secret := new(corev1.Secret)
						secretNamespace := gatewayapi.NamespaceDerefOr(certRef.Namespace, gtw.Namespace)
						err := r.client.Get(ctx,
							types.NamespacedName{Namespace: secretNamespace, Name: string(certRef.Name)},
							secret,
						)
						if err != nil && !kerrors.IsNotFound(err) {
							r.log.Error(err, "unable to find Secret")
							return reconcile.Result{}, err
						}

						r.log.Info("processing Secret", "namespace", secretNamespace, "name", string(certRef.Name))

						if secretNamespace != gtw.Namespace {
							from := ObjectKindNamespacedName{
								kind:      gatewayapi.KindGateway,
								namespace: gtw.Namespace,
								name:      gtw.Name,
							}
							to := ObjectKindNamespacedName{
								kind:      gatewayapi.KindSecret,
								namespace: secretNamespace,
								name:      string(certRef.Name),
							}
							refGrant, err := r.findReferenceGrant(ctx, from, to)
							switch {
							case err != nil:
								r.log.Error(err, "failed to find ReferenceGrant")
							case refGrant == nil:
								r.log.Info("no matching ReferenceGrants found", "from", from.kind,
									"from namespace", from.namespace, "target", to.kind, "target namespace", to.namespace)
							default:
								// RefGrant found
								resourceMap.allAssociatedRefGrants[utils.NamespacedName(refGrant)] = refGrant
								r.log.Info("added ReferenceGrant to resource map", "namespace", refGrant.Namespace,
									"name", refGrant.Name)
							}
						}

						resourceMap.allAssociatedNamespaces[secretNamespace] = struct{}{}
						resourceTree.Secrets = append(resourceTree.Secrets, secret)
					}
				}
			}
		}

		// Route Processing
		// Get TLSRoute objects and check if it exists.
		if err := r.processTLSRoutes(ctx, utils.NamespacedName(&gtw).String(), resourceMap, resourceTree); err != nil {
			return reconcile.Result{}, err
		}

		// Get HTTPRoute objects and check if it exists.
		if err := r.processHTTPRoutes(ctx, utils.NamespacedName(&gtw).String(), resourceMap, resourceTree); err != nil {
			return reconcile.Result{}, err
		}

		resourceTree.Gateways = append(resourceTree.Gateways, &gtw)
	}

	for serviceNamespaceName := range resourceMap.allAssociatedBackendRefs {
		r.log.Info("processing Service", "namespace", serviceNamespaceName.Namespace,
			"name", serviceNamespaceName.Name)

		service := new(corev1.Service)
		err := r.client.Get(ctx, serviceNamespaceName, service)
		if err != nil {
			r.log.Error(err, "failed to get Service", "namespace", serviceNamespaceName.Namespace,
				"name", serviceNamespaceName.Name)
		} else {
			resourceMap.allAssociatedNamespaces[service.Namespace] = struct{}{}
			resourceTree.Services = append(resourceTree.Services, service)
			r.log.Info("added Service to resource tree", "namespace", serviceNamespaceName.Namespace,
				"name", serviceNamespaceName.Name)
		}
	}

	// Add all ReferenceGrants to the resourceTree
	for _, referenceGrant := range resourceMap.allAssociatedRefGrants {
		resourceTree.ReferenceGrants = append(resourceTree.ReferenceGrants, referenceGrant)
	}

	// For this particular Gateway, and all associated objects, check whether the
	// namespace exists. Add to the resourceTree.
	for ns := range resourceMap.allAssociatedNamespaces {
		namespace, err := r.getNamespace(ctx, ns)
		if err != nil {
			r.log.Error(err, "unable to find the namespace")
			if kerrors.IsNotFound(err) {
				return reconcile.Result{}, nil
			}
			return reconcile.Result{}, err
		}

		resourceTree.Namespaces = append(resourceTree.Namespaces, namespace)
	}

	// Add all AuthenticationFilters to the resourceTree.
	for _, hroute := range resourceTree.HTTPRoutes {
		if filters, ok := resourceMap.httpRouteToAuthenFilters[utils.NamespacedName(hroute)]; ok {
			resourceTree.AuthenFilters = append(resourceTree.AuthenFilters, filters...)
		}
	}

	if err := updater(acceptedGC, true); err != nil {
		r.log.Error(err, "unable to update GatewayClass status")
		return reconcile.Result{}, err
	}

	// Update finalizer on the gateway class based on the resource tree.
	if len(resourceTree.Gateways) == 0 {
		r.log.Info("No gateways found for accepted gatewayclass")

		// If needed, remove the finalizer from the accepted GatewayClass.
		if err := r.removeFinalizer(ctx, acceptedGC); err != nil {
			r.log.Error(err, fmt.Sprintf("failed to remove finalizer from gatewayclass %s",
				acceptedGC.Name))
			return reconcile.Result{}, err
		}
	} else {
		// finalize the accepted GatewayClass.
		if err := r.addFinalizer(ctx, acceptedGC); err != nil {
			r.log.Error(err, fmt.Sprintf("failed adding finalizer to gatewayclass %s",
				acceptedGC.Name))
			return reconcile.Result{}, err
		}
	}

	// The Store is triggered even when there are no Gateways associated to the
	// GatewayClass. This would happen in case the last Gateway is removed and the
	// Store will be required to trigger a cleanup of envoy infra resources.
	r.resources.GatewayAPIResources.Store(acceptedGC.Name, resourceTree)

	r.log.WithName(request.Name).Info("reconciled gatewayAPI object successfully", "namespace", request.Namespace, "name", request.Name)
	return reconcile.Result{}, nil
}

func (r *gatewayAPIReconciler) getNamespace(ctx context.Context, name string) (*corev1.Namespace, error) {
	nsKey := types.NamespacedName{Name: name}
	ns := new(corev1.Namespace)
	if err := r.client.Get(ctx, nsKey, ns); err != nil {
		r.log.Error(err, "unable to get Namespace")
		return nil, err
	}
	return ns, nil
}

func (r *gatewayAPIReconciler) statusUpdateForGateway(gtw *gwapiv1b1.Gateway, svc *corev1.Service, deploy *appsv1.Deployment) {
	// nil check for unit tests.
	if r.statusUpdater == nil {
		return
	}

	// update scheduled condition
	status.UpdateGatewayStatusScheduledCondition(gtw, true)
	// update address field and ready condition
	status.UpdateGatewayStatusReadyCondition(gtw, svc, deploy)

	key := utils.NamespacedName(gtw)
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
			gCopy.Status.Conditions = status.MergeConditions(gCopy.Status.Conditions, gtw.Status.Conditions...)
			gCopy.Status.Addresses = gtw.Status.Addresses
			return gCopy
		}),
	})
}

func (r *gatewayAPIReconciler) findReferenceGrant(ctx context.Context, from, to ObjectKindNamespacedName) (*gwapiv1a2.ReferenceGrant, error) {
	refGrantList := new(gwapiv1a2.ReferenceGrantList)
	opts := &client.ListOptions{FieldSelector: fields.OneTermEqualSelector(targetRefGrantRouteIndex, to.kind)}
	if err := r.client.List(ctx, refGrantList, opts); err != nil {
		return nil, fmt.Errorf("failed to list ReferenceGrants: %v", err)
	}

	for _, refGrant := range refGrantList.Items {
		if refGrant.Namespace == to.namespace {
			for _, source := range refGrant.Spec.From {
				if source.Kind == gwapiv1a2.Kind(from.kind) && string(source.Namespace) == from.namespace {
					return &refGrant, nil
				}
			}
		}
	}

	// No ReferenceGrant found.
	return nil, nil
}

func (r *gatewayAPIReconciler) getAuthenticationFilter(ctx context.Context, ns, name string) (*egv1a1.AuthenticationFilter, error) {
	filter := new(egv1a1.AuthenticationFilter)
	key := types.NamespacedName{
		Namespace: ns,
		Name:      name,
	}
	if err := r.client.Get(ctx, key, filter); err != nil {
		return nil, err
	}

	return filter, nil
}

func addReferenceGrantIndexers(ctx context.Context, mgr manager.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(ctx, &gwapiv1a2.ReferenceGrant{}, targetRefGrantRouteIndex, func(rawObj client.Object) []string {
		refGrant := rawObj.(*gwapiv1a2.ReferenceGrant)
		var referredServices []string
		for _, target := range refGrant.Spec.To {
			referredServices = append(referredServices, string(target.Kind))
		}
		return referredServices
	}); err != nil {
		return err
	}
	return nil
}

// addHTTPRouteIndexers adds indexing on HTTPRoute.
//   - For Service objects that are referenced in HTTPRoute objects via `.spec.rules.backendRefs`.
//     This helps in querying for HTTPRoutes that are affected by a particular Service CRUD.
//   - For AuthenticationFilter objects that are referenced in HTTPRoute objects via
//     `.spec.rules[].filters`. This helps in querying for HTTPRoutes that are affected by a
//     particular AuthenticationFilter CRUD.
func addHTTPRouteIndexers(ctx context.Context, mgr manager.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(ctx, &gwapiv1b1.HTTPRoute{}, gatewayHTTPRouteIndex, gatewayHTTPRouteIndexFunc); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(ctx, &gwapiv1b1.HTTPRoute{}, serviceHTTPRouteIndex, serviceHTTPRouteIndexFunc); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(ctx, &gwapiv1b1.HTTPRoute{}, authenFilterHTTPRouteIndex, func(obj client.Object) []string {
		httproute := obj.(*gwapiv1b1.HTTPRoute)
		var filters []string
		for _, rule := range httproute.Spec.Rules {
			for i := range rule.Filters {
				filter := rule.Filters[i]
				if filter.Type == gwapiv1b1.HTTPRouteFilterExtensionRef {
					if err := gatewayapi.ValidateHTTPRouteFilter(&filter); err != nil {
						filters = append(filters,
							types.NamespacedName{
								Namespace: httproute.Namespace,
								Name:      string(filter.ExtensionRef.Name),
							}.String(),
						)
					}
				}
			}
		}
		return filters
	}); err != nil {
		return err
	}

	return nil
}

func gatewayHTTPRouteIndexFunc(rawObj client.Object) []string {
	httproute := rawObj.(*gwapiv1b1.HTTPRoute)
	var gateways []string
	for _, parent := range httproute.Spec.ParentRefs {
		if parent.Kind == nil || string(*parent.Kind) == gatewayapi.KindGateway {
			// If an explicit Gateway namespace is not provided, use the HTTPRoute namespace to
			// lookup the provided Gateway Name.
			gateways = append(gateways,
				types.NamespacedName{
					Namespace: gatewayapi.NamespaceDerefOr(parent.Namespace, httproute.Namespace),
					Name:      string(parent.Name),
				}.String(),
			)
		}
	}
	return gateways
}

func serviceHTTPRouteIndexFunc(rawObj client.Object) []string {
	httproute := rawObj.(*gwapiv1b1.HTTPRoute)
	var services []string
	for _, rule := range httproute.Spec.Rules {
		for _, backend := range rule.BackendRefs {
			if backend.Kind == nil || string(*backend.Kind) == gatewayapi.KindService {
				// If an explicit Service namespace is not provided, use the HTTPRoute namespace to
				// lookup the provided Gateway Name.
				services = append(services,
					types.NamespacedName{
						Namespace: gatewayapi.NamespaceDerefOr(backend.Namespace, httproute.Namespace),
						Name:      string(backend.Name),
					}.String(),
				)
			}
		}
	}
	return services
}

// addTLSRouteIndexers adds indexing on TLSRoute, for Service objects that are
// referenced in TLSRoute objects via `.spec.rules.backendRefs`. This helps in
// querying for TLSRoutes that are affected by a particular Service CRUD.
func addTLSRouteIndexers(ctx context.Context, mgr manager.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(ctx, &gwapiv1a2.TLSRoute{}, gatewayTLSRouteIndex, func(rawObj client.Object) []string {
		tlsRoute := rawObj.(*gwapiv1a2.TLSRoute)
		var gateways []string
		for _, parent := range tlsRoute.Spec.ParentRefs {
			if string(*parent.Kind) == gatewayapi.KindGateway {
				// If an explicit Gateway namespace is not provided, use the TLSRoute namespace to
				// lookup the provided Gateway Name.
				gateways = append(gateways,
					types.NamespacedName{
						Namespace: gatewayapi.NamespaceDerefOrAlpha(parent.Namespace, tlsRoute.Namespace),
						Name:      string(parent.Name),
					}.String(),
				)
			}
		}
		return gateways
	}); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(ctx, &gwapiv1a2.TLSRoute{}, serviceTLSRouteIndex, serviceTLSRouteIndexFunc); err != nil {
		return err
	}
	return nil
}

func serviceTLSRouteIndexFunc(rawObj client.Object) []string {
	tlsroute := rawObj.(*gwapiv1a2.TLSRoute)
	var services []string
	for _, rule := range tlsroute.Spec.Rules {
		for _, backend := range rule.BackendRefs {
			if backend.Kind == nil || string(*backend.Kind) == gatewayapi.KindService {
				// If an explicit Service namespace is not provided, use the TLSRoute namespace to
				// lookup the provided Gateway Name.
				services = append(services,
					types.NamespacedName{
						Namespace: gatewayapi.NamespaceDerefOrAlpha(backend.Namespace, tlsroute.Namespace),
						Name:      string(backend.Name),
					}.String(),
				)
			}
		}
	}
	return services
}

// addGatewayIndexers adds indexing on Gateway, for Secret objects that are
// referenced in Gateway objects. This helps in querying for Gateways that are
// affected by a particular Secret CRUD.
func addGatewayIndexers(ctx context.Context, mgr manager.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(ctx, &gwapiv1b1.Gateway{}, secretGatewayIndex, secretGatewayIndexFunc); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(ctx, &gwapiv1b1.Gateway{}, classGatewayIndex, func(rawObj client.Object) []string {
		gateway := rawObj.(*gwapiv1b1.Gateway)
		return []string{string(gateway.Spec.GatewayClassName)}
	}); err != nil {
		return err
	}
	return nil
}

func secretGatewayIndexFunc(rawObj client.Object) []string {
	gateway := rawObj.(*gwapiv1b1.Gateway)
	var secretReferences []string
	for _, listener := range gateway.Spec.Listeners {
		if listener.TLS == nil || *listener.TLS.Mode != gwapiv1b1.TLSModeTerminate {
			continue
		}
		for _, cert := range listener.TLS.CertificateRefs {
			if *cert.Kind == gatewayapi.KindSecret {
				// If an explicit Secret namespace is not provided, use the Gateway namespace to
				// lookup the provided Secret Name.
				secretReferences = append(secretReferences,
					types.NamespacedName{
						Namespace: gatewayapi.NamespaceDerefOr(cert.Namespace, gateway.Namespace),
						Name:      string(cert.Name),
					}.String(),
				)
			}
		}
	}
	return secretReferences
}

// removeFinalizer removes the gatewayclass finalizer from the provided gc, if it exists.
func (r *gatewayAPIReconciler) removeFinalizer(ctx context.Context, gc *gwapiv1b1.GatewayClass) error {
	if err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		// Get the resource.
		if err := r.client.Get(ctx, utils.NamespacedName(gc), gc); err != nil {
			if kerrors.IsNotFound(err) {
				return nil
			}
			return err
		}

		if slice.ContainsString(gc.Finalizers, gatewayClassFinalizer) {
			updated := gc.DeepCopy()
			updated.Finalizers = slice.RemoveString(updated.Finalizers, gatewayClassFinalizer)
			if err := r.client.Update(ctx, updated); err != nil {
				return fmt.Errorf("failed to remove finalizer from gatewayclass %s: %w", gc.Name, err)
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

// addFinalizer adds the gatewayclass finalizer to the provided gc, if it doesn't exist.
func (r *gatewayAPIReconciler) addFinalizer(ctx context.Context, gc *gwapiv1b1.GatewayClass) error {
	if err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		// Get the resource.
		if err := r.client.Get(ctx, utils.NamespacedName(gc), gc); err != nil {
			if kerrors.IsNotFound(err) {
				return nil
			}
			return err
		}

		if !slice.ContainsString(gc.Finalizers, gatewayClassFinalizer) {
			updated := gc.DeepCopy()
			updated.Finalizers = append(updated.Finalizers, gatewayClassFinalizer)
			if err := r.client.Update(ctx, updated); err != nil {
				return fmt.Errorf("failed to add finalizer to gatewayclass %s: %w", gc.Name, err)
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

// subscribeAndUpdateStatus subscribes to gateway API object status updates and
// writes it into the Kubernetes API Server.
func (r *gatewayAPIReconciler) subscribeAndUpdateStatus(ctx context.Context) {
	// Gateway object status updater
	go func() {
		message.HandleSubscription(r.resources.GatewayStatuses.Subscribe(ctx),
			func(update message.Update[types.NamespacedName, *gwapiv1b1.Gateway]) {
				// skip delete updates.
				if update.Delete {
					return
				}
				key := update.Key
				val := update.Value
				r.statusUpdater.Send(status.Update{
					NamespacedName: key,
					Resource:       new(gwapiv1b1.Gateway),
					Mutator: status.MutatorFunc(func(obj client.Object) client.Object {
						g, ok := obj.(*gwapiv1b1.Gateway)
						if !ok {
							panic(fmt.Sprintf("unsupported object type %T", obj))
						}
						gCopy := g.DeepCopy()
						gCopy.Status.Listeners = val.Status.Listeners
						return gCopy
					}),
				})
			},
		)
		r.log.Info("gateway status subscriber shutting down")
	}()

	// HTTPRoute object status updater
	go func() {
		message.HandleSubscription(r.resources.HTTPRouteStatuses.Subscribe(ctx),
			func(update message.Update[types.NamespacedName, *gwapiv1b1.HTTPRoute]) {
				// skip delete updates.
				if update.Delete {
					return
				}
				key := update.Key
				val := update.Value
				r.statusUpdater.Send(status.Update{
					NamespacedName: key,
					Resource:       new(gwapiv1b1.HTTPRoute),
					Mutator: status.MutatorFunc(func(obj client.Object) client.Object {
						h, ok := obj.(*gwapiv1b1.HTTPRoute)
						if !ok {
							panic(fmt.Sprintf("unsupported object type %T", obj))
						}
						hCopy := h.DeepCopy()
						hCopy.Status.Parents = val.Status.Parents
						return hCopy
					}),
				})
			},
		)
		r.log.Info("httpRoute status subscriber shutting down")
	}()

	// TLSRoute object status updater
	go func() {
		message.HandleSubscription(r.resources.TLSRouteStatuses.Subscribe(ctx),
			func(update message.Update[types.NamespacedName, *gwapiv1a2.TLSRoute]) {
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
							panic(fmt.Sprintf("unsupported object type %T", obj))
						}
						tCopy := t.DeepCopy()
						tCopy.Status.Parents = val.Status.Parents
						return tCopy
					}),
				})
			},
		)
		r.log.Info("tlsRoute status subscriber shutting down")
	}()

}
