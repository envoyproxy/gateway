// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"fmt"

	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/provider/utils"
	"github.com/envoyproxy/gateway/internal/status"
	"github.com/envoyproxy/gateway/internal/utils/slice"

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
)

const (
	classGatewayIndex        = "classGatewayIndex"
	gatewayTLSRouteIndex     = "gatewayTLSRouteIndex"
	gatewayHTTPRouteIndex    = "gatewayHTTPRouteIndex"
	secretGatewayIndex       = "secretGatewayIndex"
	targetRefGrantRouteIndex = "targetRefGrantRouteIndex"
)

type gatewayAPIReconciler struct {
	client          client.Client
	log             logr.Logger
	statusUpdater   status.Updater
	classController gwapiv1b1.GatewayController

	resources      *message.ProviderResources
	referenceStore *providerReferenceStore
}

// newGatewayAPIController
func newGatewayAPIController(mgr manager.Manager, cfg *config.Server, su status.Updater, resources *message.ProviderResources, referenceStore *providerReferenceStore) error {
	ctx := context.Background()

	r := &gatewayAPIReconciler{
		client:          mgr.GetClient(),
		log:             cfg.Logger,
		classController: gwapiv1b1.GatewayController(cfg.EnvoyGateway.Gateway.ControllerName),
		statusUpdater:   su,
		resources:       resources,
		referenceStore:  referenceStore,
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
		predicate.NewPredicateFuncs(r.validateGatewayForReconcile)); err != nil {
		return err
	}
	if err := addGatewayIndexers(ctx, mgr); err != nil {
		return err
	}

	// Watch HTTPRoute CRUDs and process affected Gateways.
	if err := c.Watch(
		&source.Kind{Type: &gwapiv1b1.HTTPRoute{}},
		&handler.EnqueueRequestForObject{},
		predicate.NewPredicateFuncs(r.validateHTTPRouteForReconcile),
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
		predicate.NewPredicateFuncs(r.validateTLSRouteForReconcile),
	); err != nil {
		return err
	}
	if err := addTLSRouteIndexers(ctx, mgr); err != nil {
		return err
	}

	// Watch Service CRUDs and process affected *Route objects.
	if err := c.Watch(&source.Kind{Type: &corev1.Service{}}, handler.EnqueueRequestsFromMapFunc(r.processServiceForOwningGateway)); err != nil {
		return err
	}

	// Watch Secret CRUDs and process affected Gateways.
	if err := c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForObject{}); err != nil {
		return err
	}

	// Watch ReferenceGrant CRUDs and process affected Gateways.
	if err := c.Watch(&source.Kind{Type: &gwapiv1a2.ReferenceGrant{}}, &handler.EnqueueRequestForObject{}); err != nil {
		return err
	}
	if err := addReferenceGrantIndexers(ctx, mgr); err != nil {
		return err
	}

	// Watch Deployment CRUDs and process affected Gateways.
	if err := c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, handler.EnqueueRequestsFromMapFunc(r.processDeploymentForOwningGateway)); err != nil {
		return err
	}

	r.log.Info("watching gatewayAPI related objects")
	return nil
}

func (r *gatewayAPIReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	r.log.WithName(request.Name).Info("reconciling gatewayAPI object", "namespace", request.Namespace, "name", request.Name)

	var gatewayClasses gwapiv1b1.GatewayClassList
	if err := r.client.List(ctx, &gatewayClasses); err != nil {
		return reconcile.Result{}, fmt.Errorf("error listing gatewayclasses: %v", err)
	}

	var cc controlledClasses

	for _, gwClass := range gatewayClasses.Items {
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

			if err := r.client.Status().Update(ctx, copy); err != nil {
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
	if acceptedGC == nil {
		r.log.Info("no accepted GatewayClass available.")
		return reconcile.Result{}, nil
	}

	resourceTree := &gatewayapi.Resources{
		Gateways:        []*gwapiv1b1.Gateway{},
		HTTPRoutes:      []*gwapiv1b1.HTTPRoute{},
		TLSRoutes:       []*gwapiv1a2.TLSRoute{},
		Services:        []*corev1.Service{},
		Secrets:         []*corev1.Secret{},
		ReferenceGrants: []*gwapiv1a2.ReferenceGrant{},
		Namespaces:      []*corev1.Namespace{},
	}

	// Add objects' namespaces into the map.
	// Make sure to add only those objects' namespaces, that exist.
	allAssociatedNamespaces := map[string]struct{}{}

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
		r.log.Info("processing Gateway", "namespace", gtw.Namespace, "name", gtw.Name)
		allAssociatedNamespaces[gtw.Namespace] = struct{}{}

		for _, listener := range gtw.Spec.Listeners {
			// Get Secret for gateway if it exists.
			if terminatesTLS(&listener) {
				for _, certRef := range listener.TLS.CertificateRefs {
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
							from := ObjectKindNamespacedName{kind: gatewayapi.KindGateway, namespace: gtw.Namespace, name: gtw.Name}
							to := ObjectKindNamespacedName{kind: gatewayapi.KindSecret, namespace: secretNamespace, name: string(certRef.Name)}
							refGrant, err := r.findReferenceGrant(ctx, from, to)
							if err != nil {
								r.log.Error(err, "unable to find ReferenceGrant that links the Secret to Gateway")
								continue
							}

							resourceTree.ReferenceGrants = append(resourceTree.ReferenceGrants, refGrant)
						}

						allAssociatedNamespaces[secretNamespace] = struct{}{}
						resourceTree.Secrets = append(resourceTree.Secrets, secret)
					}
				}
			}

			routeServicesList := map[types.NamespacedName]struct{}{}

			// Get TLSRoute objects and check if it exists.
			tlsRouteList := &gwapiv1a2.TLSRouteList{}
			if err := r.client.List(ctx, tlsRouteList, &client.ListOptions{
				FieldSelector: fields.OneTermEqualSelector(gatewayTLSRouteIndex, utils.NamespacedName(&gtw).String()),
			}); err != nil {
				r.log.Error(err, "unable to find associated TLSRoutes")
				return reconcile.Result{}, err
			}
			for _, tlsRoute := range tlsRouteList.Items {
				r.log.Info("processing TLSRoute", "namespace", tlsRoute.Namespace, "name", tlsRoute.Name)

				for _, rule := range tlsRoute.Spec.Rules {
					for _, backendRef := range rule.BackendRefs {
						ref := gatewayapi.UpgradeBackendRef(backendRef)
						if err := validateBackendRef(&ref); err != nil {
							r.log.Error(err, "invalid backendRef")
							continue
						}

						backendNamespace := gatewayapi.NamespaceDerefOrAlpha(backendRef.Namespace, gtw.Namespace)
						routeServicesList[types.NamespacedName{
							Namespace: backendNamespace,
							Name:      string(backendRef.Name),
						}] = struct{}{}

						if backendNamespace != tlsRoute.Namespace {
							from := ObjectKindNamespacedName{kind: gatewayapi.KindTLSRoute, namespace: tlsRoute.Namespace, name: tlsRoute.Name}
							to := ObjectKindNamespacedName{kind: gatewayapi.KindService, namespace: backendNamespace, name: string(backendRef.Name)}
							refGrant, err := r.findReferenceGrant(ctx, from, to)
							if err != nil {
								r.log.Error(err, "unable to find ReferenceGrant that links the Service to TLSRoute")
								continue
							}

							resourceTree.ReferenceGrants = append(resourceTree.ReferenceGrants, refGrant)
						}
					}
				}

				allAssociatedNamespaces[tlsRoute.Namespace] = struct{}{}
				resourceTree.TLSRoutes = append(resourceTree.TLSRoutes, &tlsRoute)
			}

			// Get HTTPRoute objects and check if it exists.
			httpRouteList := &gwapiv1b1.HTTPRouteList{}
			if err := r.client.List(ctx, httpRouteList, &client.ListOptions{
				FieldSelector: fields.OneTermEqualSelector(gatewayHTTPRouteIndex, utils.NamespacedName(&gtw).String()),
			}); err != nil {
				r.log.Error(err, "unable to find associated HTTPRoutes")
				return reconcile.Result{}, err
			}
			for _, httpRoute := range httpRouteList.Items {
				r.log.Info("processing HTTPRoute", "namespace", httpRoute.Namespace, "name", httpRoute.Name)

				for _, rule := range httpRoute.Spec.Rules {
					for _, backendRef := range rule.BackendRefs {
						if err := validateBackendRef(&backendRef.BackendRef); err != nil {
							r.log.Error(err, "invalid backendRef")
							continue
						}

						backendNamespace := gatewayapi.NamespaceDerefOr(backendRef.Namespace, httpRoute.Namespace)
						routeServicesList[types.NamespacedName{
							Namespace: backendNamespace,
							Name:      string(backendRef.Name),
						}] = struct{}{}

						if backendNamespace != httpRoute.Namespace {
							from := ObjectKindNamespacedName{kind: gatewayapi.KindHTTPRoute, namespace: httpRoute.Namespace, name: httpRoute.Name}
							to := ObjectKindNamespacedName{kind: gatewayapi.KindService, namespace: backendNamespace, name: string(backendRef.Name)}
							refGrant, err := r.findReferenceGrant(ctx, from, to)
							if err != nil {
								r.log.Error(err, "unable to find ReferenceGrant that links the Service to TLSRoute")
								continue
							}

							resourceTree.ReferenceGrants = append(resourceTree.ReferenceGrants, refGrant)
						}
					}
				}

				allAssociatedNamespaces[httpRoute.Namespace] = struct{}{}
				resourceTree.HTTPRoutes = append(resourceTree.HTTPRoutes, &httpRoute)
			}

			for serviceNamespaceName := range routeServicesList {
				r.log.Info("processing Service", "namespace", serviceNamespaceName.Namespace,
					"name", serviceNamespaceName.Name)

				service := new(corev1.Service)
				err := r.client.Get(ctx, serviceNamespaceName, service)
				if err != nil {
					if kerrors.IsNotFound(err) {
						continue
					}
					r.log.Error(err, "unable to find associated Services")
					return reconcile.Result{}, err
				}

				allAssociatedNamespaces[service.Namespace] = struct{}{}
				resourceTree.Services = append(resourceTree.Services, service)
			}
		}

		// For this particular Gateway, and all associated objects, check whether the
		// namespace exists. Add to the resourceTree.
		for ns := range allAssociatedNamespaces {
			namespace, err := r.getNamespace(ctx, ns)
			if err != nil {
				if kerrors.IsNotFound(err) {
					continue
				}
				r.log.Error(err, "unable to find the namespace")
				return reconcile.Result{}, err
			}

			resourceTree.Namespaces = append(resourceTree.Namespaces, namespace)
		}

		resourceTree.Gateways = append(resourceTree.Gateways, &gtw)
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

		// No further processing is required as there are no Gateways for this GatewayClass
		return reconcile.Result{}, nil
	}

	// If needed, finalize the accepted GatewayClass.
	if err := r.addFinalizer(ctx, acceptedGC); err != nil {
		r.log.Error(err, fmt.Sprintf("failed adding finalizer to gatewayclass %s",
			acceptedGC.Name))
		return reconcile.Result{}, err
	}

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

// processDeploymentForOwningGateway tries finding the owning Gateway of the Deployment
// if it exists, finds the Gateway's Service, and further updates the Gateway
// status Ready condition.
func (r *gatewayAPIReconciler) processDeploymentForOwningGateway(obj client.Object) (request []reconcile.Request) {
	// Process Deployment Reconcile nothing.
	ctx := context.Background()
	deployment := obj.(*appsv1.Deployment)
	if deployment == nil {
		return
	}

	// Check if the deployment belongs to a Gateway, if so, find the Gateway.
	gtw := r.findOwningGateway(ctx, deployment.GetLabels())
	if gtw == nil {
		return
	}

	// Check if the Service for the Gateway also exists, if it does, proceed with
	// the Gateway status update.
	svc, err := r.envoyServiceForGateway(ctx, gtw)
	if err != nil {
		r.log.Info("failed to get Service for gateway",
			"namespace", gtw.Namespace, "name", gtw.Name)
		return
	}

	r.statusUpdateForGateway(gtw, svc, deployment)
	return
}

// processServiceForOwningGateway tries finding the owning Gateway of the Service
// if it exists, finds the Gateway's Deployment, and further updates the Gateway
// status Ready condition.
func (r *gatewayAPIReconciler) processServiceForOwningGateway(obj client.Object) (request []reconcile.Request) {
	// Process Service Reconcile nothing.
	ctx := context.Background()
	svc := obj.(*corev1.Service)
	if svc == nil {
		return
	}

	// Check if the Service belongs to a Gateway, if so, find the Gateway.
	gtw := r.findOwningGateway(ctx, svc.GetLabels())
	if gtw == nil {
		return
	}

	// Check if the Deployment for the Gateway also exists, if it does, proceed with
	// the Gateway status update.
	deployment, err := r.envoyDeploymentForGateway(ctx, gtw)
	if err != nil {
		r.log.Info("failed to get Deployment for gateway",
			"namespace", gtw.Namespace, "name", gtw.Name)
		return
	}

	r.statusUpdateForGateway(gtw, svc, deployment)
	return
}

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

func (r *gatewayAPIReconciler) statusUpdateForGateway(gtw *gwapiv1b1.Gateway, svc *corev1.Service, deploy *appsv1.Deployment) {
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
	if err := r.client.List(ctx, refGrantList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(targetRefGrantRouteIndex, to.kind),
	}); err != nil {
		return nil, err
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

	return nil, fmt.Errorf("no reference grants found that target %s in namespace %s for kind %s in namespace %s",
		to.kind, to.namespace, from.kind, from.namespace)
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

// addHTTPRouteIndexers adds indexing on HTTPRoute, for Service objects that are
// referenced in HTTPRoute objects via `.spec.rules.backendRefs`. This helps in
// querying for HTTPRoutes that are affected by a particular Service CRUD.
func addHTTPRouteIndexers(ctx context.Context, mgr manager.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(ctx, &gwapiv1b1.HTTPRoute{}, gatewayHTTPRouteIndex, func(rawObj client.Object) []string {
		httproute := rawObj.(*gwapiv1b1.HTTPRoute)
		var gateways []string
		for _, parent := range httproute.Spec.ParentRefs {
			if string(*parent.Kind) == gatewayapi.KindGateway {
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
	}); err != nil {
		return err
	}
	return nil
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
	return nil
}

// addGatewayIndexers adds indexing on Gateway, for Secret objects that are
// referenced in Gateway objects. This helps in querying for Gateways that are
// affected by a particular Secret CRUD.
func addGatewayIndexers(ctx context.Context, mgr manager.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(ctx, &gwapiv1b1.Gateway{}, secretGatewayIndex, func(rawObj client.Object) []string {
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
	}); err != nil {
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

// removeFinalizer removes the gatewayclass finalizer from the provided gc, if it exists.
func (r *gatewayAPIReconciler) removeFinalizer(ctx context.Context, gc *gwapiv1b1.GatewayClass) error {
	firstAttempt := true
	if err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		if !firstAttempt {
			// Get the resource.
			if err := r.client.Get(ctx, utils.NamespacedName(gc), gc); err != nil {
				return err
			}
		}

		if slice.ContainsString(gc.Finalizers, gatewayClassFinalizer) {
			updated := gc.DeepCopy()
			updated.Finalizers = slice.RemoveString(updated.Finalizers, gatewayClassFinalizer)
			if err := r.client.Update(ctx, updated); err != nil {
				firstAttempt = false
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
	firstAttempt := true
	if err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		if !firstAttempt {
			// Get the resource.
			if err := r.client.Get(ctx, utils.NamespacedName(gc), gc); err != nil {
				return err
			}
		}

		if !slice.ContainsString(gc.Finalizers, gatewayClassFinalizer) {
			updated := gc.DeepCopy()
			updated.Finalizers = append(updated.Finalizers, gatewayClassFinalizer)
			if err := r.client.Update(ctx, updated); err != nil {
				firstAttempt = false
				return fmt.Errorf("failed to add finalizer to gatewayclass %s: %w", gc.Name, err)
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
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
	}()

	r.log.Info("status subscriber shutting down")
}
