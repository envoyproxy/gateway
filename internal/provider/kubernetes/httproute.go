// Portions of this code are based on code from Contour, available at:
// https://github.com/projectcontour/contour/blob/main/internal/controller/httproute.go

package kubernetes

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/provider/utils"
	"github.com/envoyproxy/gateway/internal/status"
)

const (
	serviceHTTPRouteIndex = "serviceHTTPRouteBackendRef"
)

type httpRouteReconciler struct {
	client          client.Client
	log             logr.Logger
	statusUpdater   status.Updater
	classController gwapiv1b1.GatewayController

	initializeOnce sync.Once
	resources      *message.ProviderResources
}

// newHTTPRouteController creates the httproute controller from mgr. The controller will be pre-configured
// to watch for HTTPRoute objects across all namespaces.
func newHTTPRouteController(mgr manager.Manager, cfg *config.Server, su status.Updater, resources *message.ProviderResources) error {
	resources.HTTPRoutesInitialized.Add(1)
	r := &httpRouteReconciler{
		client:          mgr.GetClient(),
		log:             cfg.Logger,
		classController: gwapiv1b1.GatewayController(cfg.EnvoyGateway.Gateway.ControllerName),
		statusUpdater:   su,
		resources:       resources,
	}

	c, err := controller.New("httproute", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}
	r.log.Info("created httproute controller")

	if err := c.Watch(&source.Kind{Type: &gwapiv1b1.HTTPRoute{}}, &handler.EnqueueRequestForObject{}); err != nil {
		return err
	}

	// Subscribe to status updates
	go r.subscribeAndUpdateStatus(context.Background())

	// Add indexing on HTTPRoute, for Service objects that are referenced in HTTPRoute objects
	// via `.spec.rules.backendRefs`. This helps in querying for HTTPRoutes that are affected by
	// a particular Service CRUD.
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &gwapiv1b1.HTTPRoute{}, serviceHTTPRouteIndex, func(rawObj client.Object) []string {
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

	// Watch Service CRUDs and reconcile affected HTTPRoutes.
	if err := c.Watch(
		&source.Kind{Type: &corev1.Service{}},
		handler.EnqueueRequestsFromMapFunc(r.getHTTPRoutesForService),
	); err != nil {
		return err
	}

	r.log.Info("watching httproute objects")
	return nil
}

// getHTTPRoutesForService uses a Service obj to fetch HTTPRoutes that references
// the Service using `.spec.rules.backendRefs`. The affected HTTPRoutes are then
// pushed for reconciliation.
func (r *httpRouteReconciler) getHTTPRoutesForService(obj client.Object) []reconcile.Request {
	affectedHTTPRouteList := &gwapiv1b1.HTTPRouteList{}

	if err := r.client.List(context.Background(), affectedHTTPRouteList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(serviceHTTPRouteIndex, utils.NamespacedName(obj).String()),
	}); err != nil {
		return []reconcile.Request{}
	}

	requests := make([]reconcile.Request, len(affectedHTTPRouteList.Items))
	for i, item := range affectedHTTPRouteList.Items {
		item := item
		requests[i] = reconcile.Request{
			NamespacedName: utils.NamespacedName(&item),
		}
	}

	return requests
}

func (r *httpRouteReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	log := r.log.WithValues("namespace", request.Namespace, "name", request.Name)

	log.Info("reconciling httproute")

	defer r.initializeOnce.Do(r.resources.HTTPRoutesInitialized.Done)

	// Fetch all HTTPRoutes from the cache.
	routeList := &gwapiv1b1.HTTPRouteList{}
	if err := r.client.List(ctx, routeList); err != nil {
		return reconcile.Result{}, fmt.Errorf("error listing httproutes")
	}

	found := false
	for i := range routeList.Items {
		// See if this route from the list matched the reconciled route.
		route := routeList.Items[i]
		routeKey := utils.NamespacedName(&route)
		if routeKey == request.NamespacedName {
			found = true
		}

		// Validate the route.
		gws, err := r.validateParentRefs(ctx, &route)
		if err != nil {
			// Remove the route from the watchable map since it's invalid.
			r.resources.HTTPRoutes.Delete(routeKey)
			r.log.Error(err, "invalid parentRefs for httproute")
			return reconcile.Result{}, nil
		}
		log.Info("validated httproute parentRefs")

		if len(gws) == 0 {
			// Remove the route from the watchable map since it doesn't reference
			// a managed Gateway.
			log.Info("httproute doesn't reference any managed gateways")
			r.resources.HTTPRoutes.Delete(routeKey)
			return reconcile.Result{}, nil
		}

		// Store the httproute in the resource map.
		r.resources.HTTPRoutes.Store(routeKey, &route)
		log.Info("added httproute to resource map")

		// Get the route's namespace from the cache.
		nsKey := types.NamespacedName{Name: route.Namespace}
		ns := new(corev1.Namespace)
		if err := r.client.Get(ctx, nsKey, ns); err != nil {
			if errors.IsNotFound(err) {
				// The route's namespace doesn't exist in the cache, so remove it from
				// the namespace resource map if it exists.
				if _, ok := r.resources.Namespaces.Load(nsKey.Name); ok {
					r.resources.Namespaces.Delete(nsKey.Name)
					log.Info("deleted namespace from resource map")
				}
			}
			return reconcile.Result{}, fmt.Errorf("failed to get namespace %s", nsKey.Name)
		}

		// The route's namespace exists, so add it to the resource map.
		r.resources.Namespaces.Store(nsKey.Name, ns)
		log.Info("added namespace to resource map")

		// Get the route's backendRefs from the cache. Note that a Service is the
		// only supported kind.
		for i := range route.Spec.Rules {
			for j := range route.Spec.Rules[i].BackendRefs {
				ref := route.Spec.Rules[i].BackendRefs[j]
				if err := validateBackendRef(&ref); err != nil {
					return reconcile.Result{}, fmt.Errorf("invalid backendRef: %w", err)
				}

				// The backendRef is valid, so get the referenced service from the cache.
				svcKey := types.NamespacedName{Namespace: route.Namespace, Name: string(ref.Name)}
				svc := new(corev1.Service)
				if err := r.client.Get(ctx, svcKey, svc); err != nil {
					if errors.IsNotFound(err) {
						// The ref's service doesn't exist in the cache, so remove it from
						// the resource map if it exists.
						if _, ok := r.resources.Services.Load(svcKey); ok {
							r.resources.Services.Delete(svcKey)
							log.Info("deleted service from resource map")
						}
					}
					return reconcile.Result{}, fmt.Errorf("failed to get service %s/%s",
						svcKey.Namespace, svcKey.Name)
				}

				// The backendRef Service exists, so add it to the resource map.
				r.resources.Services.Store(svcKey, svc)
				log.Info("added service to resource map")
			}
		}
	}

	if !found {
		// Delete the httproute from the resource map.
		r.resources.HTTPRoutes.Delete(request.NamespacedName)
		log.Info("deleted httproute from resource map")

		// Delete the Namespace and Service from the resource maps if no other
		// routes exist in the namespace.
		routeList = &gwapiv1b1.HTTPRouteList{}
		if err := r.client.List(ctx, routeList, &client.ListOptions{Namespace: request.Namespace}); err != nil {
			return reconcile.Result{}, fmt.Errorf("error listing httproutes")
		}
		if len(routeList.Items) == 0 {
			r.resources.Namespaces.Delete(request.Namespace)
			log.Info("deleted namespace from resource map")
			r.resources.Services.Delete(request.NamespacedName)
			log.Info("deleted service from resource map")
		}
	}

	log.Info("reconciled httproute")

	return reconcile.Result{}, nil
}

// validateBackendRef validates that ref is a reference to a local Service.
// TODO: Add support for:
//   - Validating weights.
//   - Validating ports.
//   - Referencing HTTPRoutes.
//   - Referencing Services/HTTPRoutes from other namespaces using ReferenceGrant.
func validateBackendRef(ref *gwapiv1b1.HTTPBackendRef) error {
	switch {
	case ref == nil:
		return nil
	case ref.Group != nil && *ref.Group != corev1.GroupName:
		return fmt.Errorf("invalid group; must be nil or empty string")
	case ref.Kind != nil && *ref.Kind != gatewayapi.KindService:
		return fmt.Errorf("invalid kind %q; must be %q",
			*ref.BackendRef.BackendObjectReference.Kind, gatewayapi.KindService)
	case ref.Namespace != nil:
		return fmt.Errorf("invalid namespace; must be nil")
	}

	return nil
}

// subscribeAndUpdateStatus subscribes to httproute status updates and writes it into the
// Kubernetes API Server
func (r *httpRouteReconciler) subscribeAndUpdateStatus(ctx context.Context) {
	// Subscribe to resources
	for snapshot := range r.resources.HTTPRouteStatuses.Subscribe(ctx) {
		r.log.Info("received a status notification")
		updates := snapshot.Updates
		for _, update := range updates {
			// skip delete updates.
			if update.Delete {
				continue
			}
			key := update.Key
			val := update.Value
			r.statusUpdater.Send(status.Update{
				NamespacedName: key,
				Resource:       new(gwapiv1b1.HTTPRoute),
				Mutator: status.MutatorFunc(func(obj client.Object) client.Object {
					if _, ok := obj.(*gwapiv1b1.HTTPRoute); !ok {
						panic(fmt.Sprintf("unsupported object type %T", obj))
					}
					return val
				}),
			})
		}
	}
	r.log.Info("status subscriber shutting down")
}

// validateParentRefs validates parentRefs for the provided route, returning the referenced Gateways
// managed by Envoy Gateway. The only supported parentRef is a Gateway.
func (r *httpRouteReconciler) validateParentRefs(ctx context.Context, route *gwapiv1b1.HTTPRoute) ([]gwapiv1b1.Gateway, error) {
	if route == nil {
		return nil, fmt.Errorf("httproute is nil")
	}

	var ret []gwapiv1b1.Gateway
	for i := range route.Spec.ParentRefs {
		ref := route.Spec.ParentRefs[i]
		if ref.Kind != nil && *ref.Kind != "Gateway" {
			return nil, fmt.Errorf("invalid Kind %q", *ref.Kind)
		}
		if ref.Group != nil && *ref.Group != gwapiv1b1.GroupName {
			return nil, fmt.Errorf("invalid Group %q", *ref.Group)
		}
		// Ensure the referenced Gateway exists, using the route's namespace unless
		// specified by the parentRef.
		ns := route.Namespace
		if ref.Namespace != nil {
			ns = string(*ref.Namespace)
		}
		gwKey := types.NamespacedName{
			Namespace: ns,
			Name:      string(ref.Name),
		}
		gw := new(gwapiv1b1.Gateway)
		if err := r.client.Get(ctx, gwKey, gw); err != nil {
			return nil, fmt.Errorf("failed to get gateway %s/%s: %v", gwKey.Namespace, gwKey.Name, err)
		}
		gcKey := types.NamespacedName{Name: string(gw.Spec.GatewayClassName)}
		gc := new(gwapiv1b1.GatewayClass)
		if err := r.client.Get(ctx, gcKey, gc); err != nil {
			return nil, fmt.Errorf("failed to get gatewayclass %s: %v", gcKey.Name, err)
		}
		if gc.Spec.ControllerName == r.classController {
			ret = append(ret, *gw)
		}
	}

	return ret, nil
}
