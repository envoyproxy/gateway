// Portions of this code are based on code from Contour, available at:
// https://github.com/projectcontour/contour/blob/main/internal/controller/httproute.go

package kubernetes

import (
	"context"

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
	gatewayapi_v1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
)

const (
	serviceHTTPRouteIndex = "serviceHTTPRouteBackendRef"
)

type httpRouteReconciler struct {
	client client.Client
	log    logr.Logger
}

// newHTTPRouteController creates the httproute controller from mgr. The controller will be pre-configured
// to watch for HTTPRoute objects across all namespaces.
func newHTTPRouteController(mgr manager.Manager, cfg *config.Server) error {
	r := &httpRouteReconciler{
		client: mgr.GetClient(),
		log:    cfg.Logger,
	}
	c, err := controller.New("httproute", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}
	r.log.Info("created httproute controller")

	if err := c.Watch(&source.Kind{Type: &gatewayapi_v1beta1.HTTPRoute{}}, &handler.EnqueueRequestForObject{}); err != nil {
		return err
	}

	// Add indexing on HTTPRoute, for Service objects that are referenced in HTTPRoute objects
	// via `.spec.rules.backendRefs`. This helps in querying for HTTPRoutes that are affected by
	// a particular Service CRUD.
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &gatewayapi_v1beta1.HTTPRoute{}, serviceHTTPRouteIndex, func(rawObj client.Object) []string {
		httpRoute := rawObj.(*gatewayapi_v1beta1.HTTPRoute)
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
	affectedHTTPRouteList := &gatewayapi_v1beta1.HTTPRouteList{}

	if err := r.client.List(context.Background(), affectedHTTPRouteList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(serviceHTTPRouteIndex, NamespacedNameStr(obj)),
	}); err != nil {
		return []reconcile.Request{}
	}

	requests := make([]reconcile.Request, len(affectedHTTPRouteList.Items))
	for i, item := range affectedHTTPRouteList.Items {
		requests[i] = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: item.GetNamespace(),
				Name:      item.GetName(),
			},
		}
	}

	return requests
}

func (r *httpRouteReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	log := r.log.WithValues("namespace", request.Namespace, "name", request.Name)

	log.Info("reconciling httproute")

	// Fetch the HTTPRoute from the cache.
	httpRoute := &gatewayapi_v1beta1.HTTPRoute{}
	err := r.client.Get(ctx, request.NamespacedName, httpRoute)
	if errors.IsNotFound(err) {
		log.Info("httproute not found, deleting it from the IR")
		// TODO: Delete httproute from the IR.
		// xref: https://github.com/envoyproxy/gateway/issues/34
		// xref: https://github.com/envoyproxy/gateway/issues/38
		return reconcile.Result{}, nil
	}

	log.Info("adding httproute to the IR")
	// TODO: Add httproute to the IR.
	// xref: https://github.com/envoyproxy/gateway/issues/34
	// xref: https://github.com/envoyproxy/gateway/issues/38

	return reconcile.Result{}, nil
}
