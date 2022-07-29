// Portions of this code are based on code from Contour, available at:
// https://github.com/projectcontour/contour/blob/main/internal/controller/httproute.go

package kubernetes

import (
	"context"

	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	gatewayapi_v1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
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

	r.log.Info("watching httproute objects")
	return nil
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
