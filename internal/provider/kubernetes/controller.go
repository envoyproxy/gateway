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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	egcfgv1a1 "github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/api/config/v1alpha1/validation"
	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/provider/utils"
	"github.com/envoyproxy/gateway/internal/status"
	"github.com/envoyproxy/gateway/internal/utils/slice"
)

const (
	classGatewayIndex             = "classGatewayIndex"
	gatewayTLSRouteIndex          = "gatewayTLSRouteIndex"
	gatewayHTTPRouteIndex         = "gatewayHTTPRouteIndex"
	gatewayGRPCRouteIndex         = "gatewayGRPCRouteIndex"
	gatewayTCPRouteIndex          = "gatewayTCPRouteIndex"
	gatewayUDPRouteIndex          = "gatewayUDPRouteIndex"
	secretGatewayIndex            = "secretGatewayIndex"
	targetRefGrantRouteIndex      = "targetRefGrantRouteIndex"
	serviceHTTPRouteIndex         = "serviceHTTPRouteIndex"
	serviceGRPCRouteIndex         = "serviceGRPCRouteIndex"
	serviceTLSRouteIndex          = "serviceTLSRouteIndex"
	serviceTCPRouteIndex          = "serviceTCPRouteIndex"
	serviceUDPRouteIndex          = "serviceUDPRouteIndex"
	authenFilterHTTPRouteIndex    = "authenHTTPRouteIndex"
	rateLimitFilterHTTPRouteIndex = "rateLimitHTTPRouteIndex"
	authenFilterGRPCRouteIndex    = "authenGRPCRouteIndex"
	rateLimitFilterGRPCRouteIndex = "rateLimitGRPCRouteIndex"
)

type gatewayAPIReconciler struct {
	client          client.Client
	log             logging.Logger
	statusUpdater   status.Updater
	classController gwapiv1b1.GatewayController
	store           *kubernetesProviderStore
	namespace       string
	envoyGateway    *egcfgv1a1.EnvoyGateway

	resources                *message.ProviderResources
	envoyPatchPolicyStatuses *message.EnvoyPatchPolicyStatuses
	extGVKs                  []schema.GroupVersionKind
}

// newGatewayAPIController
func newGatewayAPIController(mgr manager.Manager, cfg *config.Server, su status.Updater,
	resources *message.ProviderResources, eStatuses *message.EnvoyPatchPolicyStatuses) error {
	ctx := context.Background()

	// Gather additional resources to watch from registered extensions
	var extGVKs []schema.GroupVersionKind
	if cfg.EnvoyGateway.ExtensionManager != nil {
		for _, rsrc := range cfg.EnvoyGateway.ExtensionManager.Resources {
			gvk := schema.GroupVersionKind(rsrc)
			extGVKs = append(extGVKs, gvk)
		}
	}

	r := &gatewayAPIReconciler{
		client:                   mgr.GetClient(),
		log:                      cfg.Logger,
		classController:          gwapiv1b1.GatewayController(cfg.EnvoyGateway.Gateway.ControllerName),
		namespace:                cfg.Namespace,
		statusUpdater:            su,
		resources:                resources,
		envoyPatchPolicyStatuses: eStatuses,
		extGVKs:                  extGVKs,
		store:                    newProviderStore(),
		envoyGateway:             cfg.EnvoyGateway,
	}

	c, err := controller.New("gatewayapi", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}
	r.log.Info("created gatewayapi controller")

	// Subscribe to status updates
	r.subscribeAndUpdateStatus(ctx)

	// Watch resources
	if err := r.watchResources(ctx, mgr, c); err != nil {
		return err
	}
	return nil
}

type resourceMappings struct {
	// Map for storing namespaces for Route, Service and Gateway objects.
	allAssociatedNamespaces map[string]struct{}
	// Map for storing service NamespaceNames referred by various Route objects.
	allAssociatedBackendRefs map[types.NamespacedName]struct{}
	// Map for storing referenceGrant NamespaceNames for BackendRefs, SecretRefs.
	allAssociatedRefGrants map[types.NamespacedName]*gwapiv1a2.ReferenceGrant
	// authenFilters is a map of AuthenticationFilters, where the key is the
	// namespaced name of the AuthenticationFilter.
	authenFilters map[types.NamespacedName]*egv1a1.AuthenticationFilter
	// rateLimitFilters is a map of RateLimitFilters, where the key is the
	// namespaced name of the RateLimitFilter.
	rateLimitFilters map[types.NamespacedName]*egv1a1.RateLimitFilter
	// extensionRefFilters is a map of filters managed by an extension.
	// The key is the namespaced name of the filter and the value is the
	// unstructured form of the resource.
	extensionRefFilters map[types.NamespacedName]unstructured.Unstructured
}

func newResourceMapping() *resourceMappings {
	return &resourceMappings{
		allAssociatedNamespaces:  map[string]struct{}{},
		allAssociatedBackendRefs: map[types.NamespacedName]struct{}{},
		allAssociatedRefGrants:   map[types.NamespacedName]*gwapiv1a2.ReferenceGrant{},
		authenFilters:            map[types.NamespacedName]*egv1a1.AuthenticationFilter{},
		rateLimitFilters:         map[types.NamespacedName]*egv1a1.RateLimitFilter{},
		extensionRefFilters:      map[types.NamespacedName]unstructured.Unstructured{},
	}
}

func (r *gatewayAPIReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	r.log.WithName(request.Name).Info("reconciling gateways", "namespace")

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
		r.log.Info("no accepted gatewayclass")
		return reconcile.Result{}, nil
	}

	// Update status for all gateway classes
	for _, gc := range cc.notAcceptedClasses() {
		if err := r.gatewayClassUpdater(ctx, gc, false, string(status.ReasonOlderGatewayClassExists),
			status.MsgOlderGatewayClassExists); err != nil {
			r.resources.GatewayAPIResources.Delete(acceptedGC.Name)
			return reconcile.Result{}, err
		}
	}

	// Initialize resource types.
	resourceTree := gatewayapi.NewResources()
	resourceMap := newResourceMapping()

	if err := r.processGateways(ctx, acceptedGC, resourceMap, resourceTree); err != nil {
		return reconcile.Result{}, err
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

			// Retrieve the EndpointSlices associated with the service
			endpointSliceList := new(discoveryv1.EndpointSliceList)
			opts := []client.ListOption{
				client.MatchingLabels(map[string]string{
					discoveryv1.LabelServiceName: serviceNamespaceName.Name,
				}),
				client.InNamespace(serviceNamespaceName.Namespace),
			}
			if err := r.client.List(ctx, endpointSliceList, opts...); err != nil {
				r.log.Error(err, "failed to get EndpointSlices", "namespace", serviceNamespaceName.Namespace,
					"service", serviceNamespaceName.Name)
			} else {
				for _, endpointSlice := range endpointSliceList.Items {
					endpointSlice := endpointSlice
					r.log.Info("added EndpointSlice to resource tree", "namespace", endpointSlice.Namespace,
						"name", endpointSlice.Name)
					resourceTree.EndpointSlices = append(resourceTree.EndpointSlices, &endpointSlice)
				}
			}
		}
	}

	// Add all ReferenceGrants to the resourceTree
	for _, referenceGrant := range resourceMap.allAssociatedRefGrants {
		resourceTree.ReferenceGrants = append(resourceTree.ReferenceGrants, referenceGrant)
	}

	// Add all EnvoyPatchPolicies
	if r.envoyGateway.ExtensionAPIs != nil && r.envoyGateway.ExtensionAPIs.EnableEnvoyPatchPolicy {
		envoyPatchPolicies := egv1a1.EnvoyPatchPolicyList{}
		if err := r.client.List(ctx, &envoyPatchPolicies); err != nil {
			return reconcile.Result{}, fmt.Errorf("error listing envoypatchpolicies: %v", err)
		}

		for _, policy := range envoyPatchPolicies.Items {
			policy := policy
			// Discard Status to reduce memory consumption in watchable
			// It will be recomputed by the gateway-api layer
			policy.Status = egv1a1.EnvoyPatchPolicyStatus{}
			resourceTree.EnvoyPatchPolicies = append(resourceTree.EnvoyPatchPolicies, &policy)
		}
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

	// Process the parametersRef of the accepted GatewayClass.
	if acceptedGC.Spec.ParametersRef != nil && acceptedGC.DeletionTimestamp == nil {
		if err := r.processParamsRef(ctx, acceptedGC, resourceTree); err != nil {
			msg := fmt.Sprintf("%s: %v", status.MsgGatewayClassInvalidParams, err)
			if err := r.gatewayClassUpdater(ctx, acceptedGC, false, string(gwapiv1b1.GatewayClassReasonInvalidParameters), msg); err != nil {
				r.log.Error(err, "unable to update GatewayClass status")
			}
			r.log.Error(err, "failed to process parametersRef for gatewayclass", "name", acceptedGC.Name)
			return reconcile.Result{}, err
		}
	}

	if err := r.gatewayClassUpdater(ctx, acceptedGC, true, string(gwapiv1b1.GatewayClassReasonAccepted), status.MsgValidGatewayClass); err != nil {
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

	r.log.WithName(request.Name).Info("reconciled gateways successfully")
	return reconcile.Result{}, nil
}

func (r *gatewayAPIReconciler) gatewayClassUpdater(ctx context.Context, gc *gwapiv1b1.GatewayClass, accepted bool, reason, msg string) error {
	if r.statusUpdater != nil {
		r.statusUpdater.Send(status.Update{
			NamespacedName: types.NamespacedName{Name: gc.Name},
			Resource:       &gwapiv1b1.GatewayClass{},
			Mutator: status.MutatorFunc(func(obj client.Object) client.Object {
				gc, ok := obj.(*gwapiv1b1.GatewayClass)
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

func (r *gatewayAPIReconciler) getNamespace(ctx context.Context, name string) (*corev1.Namespace, error) {
	nsKey := types.NamespacedName{Name: name}
	ns := new(corev1.Namespace)
	if err := r.client.Get(ctx, nsKey, ns); err != nil {
		r.log.Error(err, "unable to get Namespace")
		return nil, err
	}
	return ns, nil
}

func (r *gatewayAPIReconciler) statusUpdateForGateway(ctx context.Context, gtw *gwapiv1b1.Gateway) {
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
		Resource:       new(gwapiv1b1.Gateway),
		Mutator: status.MutatorFunc(func(obj client.Object) client.Object {
			g, ok := obj.(*gwapiv1b1.Gateway)
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

func (r *gatewayAPIReconciler) findReferenceGrant(ctx context.Context, from, to ObjectKindNamespacedName) (*gwapiv1a2.ReferenceGrant, error) {
	refGrantList := new(gwapiv1a2.ReferenceGrantList)
	opts := &client.ListOptions{FieldSelector: fields.OneTermEqualSelector(targetRefGrantRouteIndex, to.kind)}
	if err := r.client.List(ctx, refGrantList, opts); err != nil {
		return nil, fmt.Errorf("failed to list ReferenceGrants: %v", err)
	}

	for _, refGrant := range refGrantList.Items {
		if refGrant.Namespace == to.namespace {
			for _, src := range refGrant.Spec.From {
				if src.Kind == gwapiv1a2.Kind(from.kind) && string(src.Namespace) == from.namespace {
					return &refGrant, nil
				}
			}
		}
	}

	// No ReferenceGrant found.
	return nil, nil
}

func (r *gatewayAPIReconciler) processGateways(ctx context.Context, acceptedGC *gwapiv1b1.GatewayClass, resourceMap *resourceMappings, resourceTree *gatewayapi.Resources) error {
	// Find gateways for the acceptedGC
	// Find the Gateways that reference this Class.
	gatewayList := &gwapiv1b1.GatewayList{}
	if err := r.client.List(ctx, gatewayList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(classGatewayIndex, acceptedGC.Name),
	}); err != nil {
		r.log.Info("no associated Gateways found for GatewayClass", "name", acceptedGC.Name)
		return err
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
							return err
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
			return err
		}

		// Get HTTPRoute objects and check if it exists.
		if err := r.processHTTPRoutes(ctx, utils.NamespacedName(&gtw).String(), resourceMap, resourceTree); err != nil {
			return err
		}

		// Get GRPCRoute objects and check if it exists.
		if err := r.processGRPCRoutes(ctx, utils.NamespacedName(&gtw).String(), resourceMap, resourceTree); err != nil {
			return err
		}

		// Get TCPRoute objects and check if it exists.
		if err := r.processTCPRoutes(ctx, utils.NamespacedName(&gtw).String(), resourceMap, resourceTree); err != nil {
			return err
		}

		// Get UDPRoute objects and check if it exists.
		if err := r.processUDPRoutes(ctx, utils.NamespacedName(&gtw).String(), resourceMap, resourceTree); err != nil {
			return err
		}

		// Discard Status to reduce memory consumption in watchable
		// It will be recomputed by the gateway-api layer
		gtw.Status = gwapiv1b1.GatewayStatus{}
		resourceTree.Gateways = append(resourceTree.Gateways, &gtw)
	}
	return nil
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
//   - For AuthenticationFilter and RateLimitFilter objects that are referenced in HTTPRoute objects via
//     `.spec.rules[].filters`. This helps in querying for HTTPRoutes that are affected by a
//     particular AuthenticationFilter CRUD.
func addHTTPRouteIndexers(ctx context.Context, mgr manager.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(ctx, &gwapiv1b1.HTTPRoute{}, gatewayHTTPRouteIndex, gatewayHTTPRouteIndexFunc); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(ctx, &gwapiv1b1.HTTPRoute{}, serviceHTTPRouteIndex, serviceHTTPRouteIndexFunc); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(ctx, &gwapiv1b1.HTTPRoute{}, authenFilterHTTPRouteIndex, authenFilterHTTPRouteIndexFunc); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(ctx, &gwapiv1b1.HTTPRoute{}, rateLimitFilterHTTPRouteIndex, rateLimitFilterHTTPRouteIndexFunc); err != nil {
		return err
	}
	return nil
}

func authenFilterHTTPRouteIndexFunc(rawObj client.Object) []string {
	httproute := rawObj.(*gwapiv1b1.HTTPRoute)
	var filters []string
	for _, rule := range httproute.Spec.Rules {
		for i := range rule.Filters {
			filter := rule.Filters[i]
			if gatewayapi.IsAuthnHTTPFilter(&filter) {
				if err := gatewayapi.ValidateHTTPRouteFilter(&filter); err == nil {
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
}

func rateLimitFilterHTTPRouteIndexFunc(rawObj client.Object) []string {
	httproute := rawObj.(*gwapiv1b1.HTTPRoute)
	var filters []string
	for _, rule := range httproute.Spec.Rules {
		for i := range rule.Filters {
			filter := rule.Filters[i]
			if gatewayapi.IsRateLimitHTTPFilter(&filter) {
				if err := gatewayapi.ValidateHTTPRouteFilter(&filter); err == nil {
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

// addGRPCRouteIndexers adds indexing on GRPCRoute, for Service objects that are
// referenced in GRPCRoute objects via `.spec.rules.backendRefs`. This helps in
// querying for GRPCRoutes that are affected by a particular Service CRUD.
func addGRPCRouteIndexers(ctx context.Context, mgr manager.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(ctx, &gwapiv1a2.GRPCRoute{}, gatewayGRPCRouteIndex, gatewayGRPCRouteIndexFunc); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(ctx, &gwapiv1a2.GRPCRoute{}, serviceGRPCRouteIndex, serviceGRPCRouteIndexFunc); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(ctx, &gwapiv1a2.GRPCRoute{}, authenFilterGRPCRouteIndex, authenFilterGRPCRouteIndexFunc); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(ctx, &gwapiv1a2.GRPCRoute{}, rateLimitFilterGRPCRouteIndex, rateLimitFilterGRPCRouteIndexFunc); err != nil {
		return err
	}

	return nil
}

func gatewayGRPCRouteIndexFunc(rawObj client.Object) []string {
	grpcroute := rawObj.(*gwapiv1a2.GRPCRoute)
	var gateways []string
	for _, parent := range grpcroute.Spec.ParentRefs {
		if parent.Kind == nil || string(*parent.Kind) == gatewayapi.KindGateway {
			// If an explicit Gateway namespace is not provided, use the GRPCRoute namespace to
			// lookup the provided Gateway Name.
			gateways = append(gateways,
				types.NamespacedName{
					Namespace: gatewayapi.NamespaceDerefOr(parent.Namespace, grpcroute.Namespace),
					Name:      string(parent.Name),
				}.String(),
			)
		}
	}
	return gateways
}

func serviceGRPCRouteIndexFunc(rawObj client.Object) []string {
	grpcroute := rawObj.(*gwapiv1a2.GRPCRoute)
	var services []string
	for _, rule := range grpcroute.Spec.Rules {
		for _, backend := range rule.BackendRefs {
			if backend.Kind == nil || string(*backend.Kind) == gatewayapi.KindService {
				// If an explicit Service namespace is not provided, use the GRPCRoute namespace to
				// lookup the provided Gateway Name.
				services = append(services,
					types.NamespacedName{
						Namespace: gatewayapi.NamespaceDerefOr(backend.Namespace, grpcroute.Namespace),
						Name:      string(backend.Name),
					}.String(),
				)
			}
		}
	}
	return services
}

func authenFilterGRPCRouteIndexFunc(rawObj client.Object) []string {
	grpcroute := rawObj.(*gwapiv1a2.GRPCRoute)
	var filters []string
	for _, rule := range grpcroute.Spec.Rules {
		for i := range rule.Filters {
			filter := rule.Filters[i]
			if gatewayapi.IsAuthnGRPCFilter(&filter) {
				if err := gatewayapi.ValidateGRPCRouteFilter(&filter); err == nil {
					filters = append(filters,
						types.NamespacedName{
							Namespace: grpcroute.Namespace,
							Name:      string(filter.ExtensionRef.Name),
						}.String(),
					)
				}
			}
		}
	}
	return filters
}

func rateLimitFilterGRPCRouteIndexFunc(rawObj client.Object) []string {
	grpcroute := rawObj.(*gwapiv1a2.GRPCRoute)
	var filters []string
	for _, rule := range grpcroute.Spec.Rules {
		for i := range rule.Filters {
			filter := rule.Filters[i]
			if gatewayapi.IsRateLimitGRPCFilter(&filter) {
				if err := gatewayapi.ValidateGRPCRouteFilter(&filter); err == nil {
					filters = append(filters,
						types.NamespacedName{
							Namespace: grpcroute.Namespace,
							Name:      string(filter.ExtensionRef.Name),
						}.String(),
					)
				}
			}
		}
	}
	return filters
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

// addTCPRouteIndexers adds indexing on TCPRoute, for Service objects that are
// referenced in TCPRoute objects via `.spec.rules.backendRefs`. This helps in
// querying for TCPRoutes that are affected by a particular Service CRUD.
func addTCPRouteIndexers(ctx context.Context, mgr manager.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(ctx, &gwapiv1a2.TCPRoute{}, gatewayTCPRouteIndex, func(rawObj client.Object) []string {
		tcpRoute := rawObj.(*gwapiv1a2.TCPRoute)
		var gateways []string
		for _, parent := range tcpRoute.Spec.ParentRefs {
			if string(*parent.Kind) == gatewayapi.KindGateway {
				// If an explicit Gateway namespace is not provided, use the TCPRoute namespace to
				// lookup the provided Gateway Name.
				gateways = append(gateways,
					types.NamespacedName{
						Namespace: gatewayapi.NamespaceDerefOrAlpha(parent.Namespace, tcpRoute.Namespace),
						Name:      string(parent.Name),
					}.String(),
				)
			}
		}
		return gateways
	}); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(ctx, &gwapiv1a2.TCPRoute{}, serviceTCPRouteIndex, serviceTCPRouteIndexFunc); err != nil {
		return err
	}
	return nil
}

func serviceTCPRouteIndexFunc(rawObj client.Object) []string {
	tcpRoute := rawObj.(*gwapiv1a2.TCPRoute)
	var services []string
	for _, rule := range tcpRoute.Spec.Rules {
		for _, backend := range rule.BackendRefs {
			if backend.Kind == nil || string(*backend.Kind) == gatewayapi.KindService {
				// If an explicit Service namespace is not provided, use the TCPRoute namespace to
				// lookup the provided Gateway Name.
				services = append(services,
					types.NamespacedName{
						Namespace: gatewayapi.NamespaceDerefOrAlpha(backend.Namespace, tcpRoute.Namespace),
						Name:      string(backend.Name),
					}.String(),
				)
			}
		}
	}
	return services
}

// addUDPRouteIndexers adds indexing on UDPRoute, for Service objects that are
// referenced in UDPRoute objects via `.spec.rules.backendRefs`. This helps in
// querying for UDPRoutes that are affected by a particular Service CRUD.
func addUDPRouteIndexers(ctx context.Context, mgr manager.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(ctx, &gwapiv1a2.UDPRoute{}, gatewayUDPRouteIndex, func(rawObj client.Object) []string {
		udpRoute := rawObj.(*gwapiv1a2.UDPRoute)
		var gateways []string
		for _, parent := range udpRoute.Spec.ParentRefs {
			if string(*parent.Kind) == gatewayapi.KindGateway {
				// If an explicit Gateway namespace is not provided, use the UDPRoute namespace to
				// lookup the provided Gateway Name.
				gateways = append(gateways,
					types.NamespacedName{
						Namespace: gatewayapi.NamespaceDerefOrAlpha(parent.Namespace, udpRoute.Namespace),
						Name:      string(parent.Name),
					}.String(),
				)
			}
		}
		return gateways
	}); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(ctx, &gwapiv1a2.UDPRoute{}, serviceUDPRouteIndex, serviceUDPRouteIndexFunc); err != nil {
		return err
	}
	return nil
}

func serviceUDPRouteIndexFunc(rawObj client.Object) []string {
	udproute := rawObj.(*gwapiv1a2.UDPRoute)
	var services []string
	for _, rule := range udproute.Spec.Rules {
		for _, backend := range rule.BackendRefs {
			if backend.Kind == nil || string(*backend.Kind) == gatewayapi.KindService {
				// If an explicit Service namespace is not provided, use the UDPRoute namespace to
				// lookup the provided Gateway Name.
				services = append(services,
					types.NamespacedName{
						Namespace: gatewayapi.NamespaceDerefOrAlpha(backend.Namespace, udproute.Namespace),
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
	if slice.ContainsString(gc.Finalizers, gatewayClassFinalizer) {
		base := client.MergeFrom(gc.DeepCopy())
		gc.Finalizers = slice.RemoveString(gc.Finalizers, gatewayClassFinalizer)
		if err := r.client.Patch(ctx, gc, base); err != nil {
			return fmt.Errorf("failed to remove finalizer from gatewayclass %s: %w", gc.Name, err)
		}
	}
	return nil
}

// addFinalizer adds the gatewayclass finalizer to the provided gc, if it doesn't exist.
func (r *gatewayAPIReconciler) addFinalizer(ctx context.Context, gc *gwapiv1b1.GatewayClass) error {
	if !slice.ContainsString(gc.Finalizers, gatewayClassFinalizer) {
		base := client.MergeFrom(gc.DeepCopy())
		gc.Finalizers = append(gc.Finalizers, gatewayClassFinalizer)
		if err := r.client.Patch(ctx, gc, base); err != nil {
			return fmt.Errorf("failed to add finalizer to gatewayclass %s: %w", gc.Name, err)
		}
	}
	return nil
}

// subscribeAndUpdateStatus subscribes to gateway API object status updates and
// writes it into the Kubernetes API Server.
func (r *gatewayAPIReconciler) subscribeAndUpdateStatus(ctx context.Context) {
	// Gateway object status updater
	go func() {
		message.HandleSubscription(r.resources.GatewayStatuses.Subscribe(ctx),
			func(update message.Update[types.NamespacedName, *gwapiv1b1.GatewayStatus]) {
				// skip delete updates.
				if update.Delete {
					return
				}
				// Get gateway object
				gtw := new(gwapiv1b1.Gateway)
				if err := r.client.Get(ctx, update.Key, gtw); err != nil {
					r.log.Error(err, "gateway not found", "namespace", gtw.Namespace, "name", gtw.Name)
					return
				}
				// Set the updated Status and call the status update
				gtw.Status = *update.Value
				r.statusUpdateForGateway(ctx, gtw)
			},
		)
		r.log.Info("gateway status subscriber shutting down")
	}()

	// HTTPRoute object status updater
	go func() {
		message.HandleSubscription(r.resources.HTTPRouteStatuses.Subscribe(ctx),
			func(update message.Update[types.NamespacedName, *gwapiv1b1.HTTPRouteStatus]) {
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
		message.HandleSubscription(r.resources.GRPCRouteStatuses.Subscribe(ctx),
			func(update message.Update[types.NamespacedName, *gwapiv1a2.GRPCRouteStatus]) {
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
							panic(fmt.Sprintf("unsupported object type %T", obj))
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
		message.HandleSubscription(r.resources.TLSRouteStatuses.Subscribe(ctx),
			func(update message.Update[types.NamespacedName, *gwapiv1a2.TLSRouteStatus]) {
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
		message.HandleSubscription(r.resources.TCPRouteStatuses.Subscribe(ctx),
			func(update message.Update[types.NamespacedName, *gwapiv1a2.TCPRouteStatus]) {
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
							panic(fmt.Sprintf("unsupported object type %T", obj))
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
		message.HandleSubscription(r.resources.UDPRouteStatuses.Subscribe(ctx),
			func(update message.Update[types.NamespacedName, *gwapiv1a2.UDPRouteStatus]) {
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
							panic(fmt.Sprintf("unsupported object type %T", obj))
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
		message.HandleSubscription(r.envoyPatchPolicyStatuses.Subscribe(ctx),
			func(update message.Update[types.NamespacedName, *egv1a1.EnvoyPatchPolicyStatus]) {
				// skip delete updates.
				if update.Delete {
					return
				}
				key := update.Key
				val := update.Value
				r.statusUpdater.Send(status.Update{
					NamespacedName: key,
					Resource:       new(egv1a1.EnvoyPatchPolicy),
					Mutator: status.MutatorFunc(func(obj client.Object) client.Object {
						t, ok := obj.(*egv1a1.EnvoyPatchPolicy)
						if !ok {
							panic(fmt.Sprintf("unsupported object type %T", obj))
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
}

// watchResources watches gateway api resources.
func (r *gatewayAPIReconciler) watchResources(ctx context.Context, mgr manager.Manager, c controller.Controller) error {
	// Only enqueue GatewayClass objects that match this Envoy Gateway's controller name.
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &gwapiv1b1.GatewayClass{}),
		handler.EnqueueRequestsFromMapFunc(r.enqueueClass),
		predicate.NewPredicateFuncs(r.hasMatchingController),
	); err != nil {
		return err
	}

	// Only enqueue EnvoyProxy objects that match this Envoy Gateway's GatewayClass.
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &egcfgv1a1.EnvoyProxy{}),
		handler.EnqueueRequestsFromMapFunc(r.enqueueClass),
		predicate.ResourceVersionChangedPredicate{},
		predicate.NewPredicateFuncs(r.hasManagedClass),
	); err != nil {
		return err
	}

	// Watch Gateway CRUDs and reconcile affected GatewayClass.
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &gwapiv1b1.Gateway{}),
		handler.EnqueueRequestsFromMapFunc(r.enqueueClass),
		predicate.NewPredicateFuncs(r.validateGatewayForReconcile),
	); err != nil {
		return err
	}
	if err := addGatewayIndexers(ctx, mgr); err != nil {
		return err
	}

	// Watch HTTPRoute CRUDs and process affected Gateways.
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &gwapiv1b1.HTTPRoute{}),
		handler.EnqueueRequestsFromMapFunc(r.enqueueClass),
	); err != nil {
		return err
	}
	if err := addHTTPRouteIndexers(ctx, mgr); err != nil {
		return err
	}

	// Watch GRPCRoute CRUDs and process affected Gateways.
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &gwapiv1a2.GRPCRoute{}),
		handler.EnqueueRequestsFromMapFunc(r.enqueueClass),
	); err != nil {
		return err
	}
	if err := addGRPCRouteIndexers(ctx, mgr); err != nil {
		return err
	}

	// Watch TLSRoute CRUDs and process affected Gateways.
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &gwapiv1a2.TLSRoute{}),
		handler.EnqueueRequestsFromMapFunc(r.enqueueClass),
	); err != nil {
		return err
	}
	if err := addTLSRouteIndexers(ctx, mgr); err != nil {
		return err
	}

	// Watch UDPRoute CRUDs and process affected Gateways.
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &gwapiv1a2.UDPRoute{}),
		handler.EnqueueRequestsFromMapFunc(r.enqueueClass),
	); err != nil {
		return err
	}
	if err := addUDPRouteIndexers(ctx, mgr); err != nil {
		return err
	}

	// Watch TCPRoute CRUDs and process affected Gateways.
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &gwapiv1a2.TCPRoute{}),
		handler.EnqueueRequestsFromMapFunc(r.enqueueClass),
	); err != nil {
		return err
	}
	if err := addTCPRouteIndexers(ctx, mgr); err != nil {
		return err
	}

	// Watch Service CRUDs and process affected *Route objects.
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &corev1.Service{}),
		handler.EnqueueRequestsFromMapFunc(r.enqueueClass),
		predicate.NewPredicateFuncs(r.validateServiceForReconcile)); err != nil {
		return err
	}

	// Watch EndpointSlice CRUDs and process affected *Route objects.
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &discoveryv1.EndpointSlice{}),
		handler.EnqueueRequestsFromMapFunc(r.enqueueClass),
		predicate.NewPredicateFuncs(r.validateEndpointSliceForReconcile)); err != nil {
		return err
	}

	// Watch Node CRUDs to update Gateway Address exposed by Service of type NodePort.
	// Node creation/deletion and ExternalIP updates would require update in the Gateway
	// resource address.
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &corev1.Node{}),
		handler.EnqueueRequestsFromMapFunc(r.enqueueClass),
		predicate.NewPredicateFuncs(r.handleNode),
	); err != nil {
		return err
	}

	// Watch Secret CRUDs and process affected Gateways.
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &corev1.Secret{}),
		handler.EnqueueRequestsFromMapFunc(r.enqueueClass),
		predicate.NewPredicateFuncs(r.validateSecretForReconcile),
	); err != nil {
		return err
	}

	// Watch ReferenceGrant CRUDs and process affected Gateways.
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &gwapiv1a2.ReferenceGrant{}),
		handler.EnqueueRequestsFromMapFunc(r.enqueueClass),
	); err != nil {
		return err
	}
	if err := addReferenceGrantIndexers(ctx, mgr); err != nil {
		return err
	}

	// Watch Deployment CRUDs and process affected Gateways.
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &appsv1.Deployment{}),
		handler.EnqueueRequestsFromMapFunc(r.enqueueClass),
		predicate.NewPredicateFuncs(r.validateDeploymentForReconcile),
	); err != nil {
		return err
	}

	// Watch AuthenticationFilter CRUDs and enqueue associated HTTPRoute objects.
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &egv1a1.AuthenticationFilter{}),
		handler.EnqueueRequestsFromMapFunc(r.enqueueClass),
		predicate.NewPredicateFuncs(r.httpRoutesForAuthenticationFilter)); err != nil {
		return err
	}

	// Watch RateLimitFilter CRUDs and enqueue associated HTTPRoute objects.
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &egv1a1.RateLimitFilter{}),
		handler.EnqueueRequestsFromMapFunc(r.enqueueClass),
		predicate.NewPredicateFuncs(r.httpRoutesForRateLimitFilter)); err != nil {
		return err
	}

	// Watch EnvoyPatchPolicy if enabled in config
	if r.envoyGateway.ExtensionAPIs != nil && r.envoyGateway.ExtensionAPIs.EnableEnvoyPatchPolicy {
		// Watch EnvoyPatchPolicy CRUDs
		if err := c.Watch(
			source.Kind(mgr.GetCache(), &egv1a1.EnvoyPatchPolicy{}),
			&handler.EnqueueRequestForObject{}); err != nil {
			return err
		}
	}

	r.log.Info("Watching gatewayAPI related objects")

	// Watch any additional GVKs from the registered extension.
	for _, gvk := range r.extGVKs {
		u := &unstructured.Unstructured{}
		u.SetGroupVersionKind(gvk)
		if err := c.Watch(source.Kind(mgr.GetCache(), u),
			handler.EnqueueRequestsFromMapFunc(r.enqueueClass)); err != nil {
			return err
		}
		r.log.Info("Watching additional resource", "resource", gvk.String())
	}
	return nil
}

func (r *gatewayAPIReconciler) enqueueClass(_ context.Context, _ client.Object) []reconcile.Request {
	return []reconcile.Request{{NamespacedName: types.NamespacedName{
		Name: string(r.classController),
	}}}
}

func (r *gatewayAPIReconciler) hasManagedClass(obj client.Object) bool {
	ep, ok := obj.(*egcfgv1a1.EnvoyProxy)
	if !ok {
		panic(fmt.Sprintf("unsupported object type %T", obj))
	}

	// The EnvoyProxy must be in the same namespace as EG.
	if ep.Namespace != r.namespace {
		r.log.Info("envoyproxy namespace does not match Envoy Gateway's namespace",
			"namespace", ep.Namespace, "name", ep.Name)
		return false
	}

	gcList := new(gwapiv1b1.GatewayClassList)
	err := r.client.List(context.TODO(), gcList)
	if err != nil {
		r.log.Error(err, "failed to list gatewayclasses")
		return false
	}

	for i := range gcList.Items {
		gc := gcList.Items[i]
		// Reconcile the managed GatewayClass if it's referenced by the EnvoyProxy.
		if r.hasMatchingController(&gc) &&
			classAccepted(&gc) &&
			classRefsEnvoyProxy(&gc, ep) {
			return true
		}
	}

	return false
}

// processParamsRef processes the parametersRef of the provided GatewayClass.
func (r *gatewayAPIReconciler) processParamsRef(ctx context.Context, gc *gwapiv1b1.GatewayClass, resourceTree *gatewayapi.Resources) error {
	if !refsEnvoyProxy(gc) {
		return fmt.Errorf("unsupported parametersRef for gatewayclass %s", gc.Name)
	}

	epList := new(egcfgv1a1.EnvoyProxyList)
	// The EnvoyProxy must be in the same namespace as EG.
	if err := r.client.List(ctx, epList, &client.ListOptions{Namespace: r.namespace}); err != nil {
		return fmt.Errorf("failed to list envoyproxies in namespace %s: %v", r.namespace, err)
	}

	if len(epList.Items) == 0 {
		r.log.Info("no envoyproxies exist in", "namespace", r.namespace)
		return nil
	}

	found := false
	valid := false
	var validationErr error
	for i := range epList.Items {
		ep := epList.Items[i]
		r.log.Info("processing envoyproxy", "namespace", ep.Namespace, "name", ep.Name)
		if classRefsEnvoyProxy(gc, &ep) {
			found = true
			if err := validation.ValidateEnvoyProxy(&ep); err != nil {
				validationErr = fmt.Errorf("invalid envoyproxy: %v", err)
				continue
			}
			valid = true
			resourceTree.EnvoyProxy = &ep
			break
		}
	}

	if !found {
		return fmt.Errorf("failed to find envoyproxy referenced by gatewayclass: %s", gc.Name)
	}

	if !valid {
		return fmt.Errorf("invalid gatewayclass %s: %v", gc.Name, validationErr)
	}

	return nil
}
