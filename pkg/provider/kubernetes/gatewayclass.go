package kubernetes

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/envoyproxy/gateway/pkg/envoygateway/config"
)

type gatewayClassReconciler struct {
	client     client.Client
	controller gwapiv1b1.GatewayController
	log        logr.Logger
}

// newGatewayClassController creates the gatewayclass controller. The controller
// will be pre-configured to watch for cluster-scoped GatewayClass objects with
// a controller field that matches name.
func newGatewayClassController(mgr manager.Manager, cfg *config.Server) error {
	r := &gatewayClassReconciler{
		client:     mgr.GetClient(),
		controller: gwapiv1b1.GatewayController(cfg.EnvoyGateway.Gateway.ControllerName),
		log:        cfg.Logger,
	}

	c, err := controller.New("gatewayclass-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}
	r.log.Info("created gatewayclass controller")

	// Only enqueue GatewayClass objects that match this Envoy Gateway's controller name.
	if err := c.Watch(
		&source.Kind{Type: &gwapiv1b1.GatewayClass{}},
		&handler.EnqueueRequestForObject{},
		predicate.NewPredicateFuncs(r.hasMatchingController),
	); err != nil {
		return err
	}
	r.log.Info("watching gatewayclass objects")

	return nil
}

// hasMatchingController returns true if the provided object is a GatewayClass
// with a Spec.Controller string matching this Envoy Gateway's controller string,
// or false otherwise.
func (r *gatewayClassReconciler) hasMatchingController(obj client.Object) bool {
	log := r.log.WithName(obj.GetName())

	gc, ok := obj.(*gwapiv1b1.GatewayClass)
	if !ok {
		log.Info("bypassing reconciliation due to unexpected object type", "type", obj)
		return false
	}

	if gc.Spec.ControllerName == r.controller {
		log.Info("enqueueing gatewayclass")
		return true
	}

	log.Info("bypassing reconciliation due to controller name", "controllerName", gc.Spec.ControllerName)
	return false
}

func (r *gatewayClassReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	r.log.WithName(request.Name).Info("reconciling gatewayclass")

	var gatewayClasses gwapiv1b1.GatewayClassList
	if err := r.client.List(context.Background(), &gatewayClasses); err != nil {
		return reconcile.Result{}, fmt.Errorf("error listing gatewayclasses: %w", err)
	}

	r.log.WithName(request.Name).Info("reconciled gatewayclass")
	return reconcile.Result{}, nil
}
