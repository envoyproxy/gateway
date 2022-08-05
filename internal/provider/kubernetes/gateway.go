// Portions of this code are based on code from Contour, available at:
// https://github.com/projectcontour/contour/blob/main/internal/controller/gateway.go

package kubernetes

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/envoyproxy/gateway/internal/envoygateway/config"
)

type gatewayReconciler struct {
	client client.Client
	// classController is the configured gatewayclass controller name.
	classController gwapiv1b1.GatewayController
	log             logr.Logger

	initializeOnce sync.Once
	resourceTable  *ResourceTable
}

// newGatewayController creates a gateway controller. The controller will watch for
// Gateway objects across all namespaces and reconcile those that match the configured
// gatewayclass controller name.
func newGatewayController(mgr manager.Manager, cfg *config.Server, resourceTable *ResourceTable) error {
	resourceTable.Initialized.Add(1)
	r := &gatewayReconciler{
		client:          mgr.GetClient(),
		classController: gwapiv1b1.GatewayController(cfg.EnvoyGateway.Gateway.ControllerName),
		log:             cfg.Logger,
		resourceTable:   resourceTable,
	}

	c, err := controller.New("gateway", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}
	r.log.Info("created gateway controller")

	// Only enqueue Gateway objects that match this Envoy Gateway's controller name.
	if err := c.Watch(
		&source.Kind{Type: &gwapiv1b1.Gateway{}},
		&handler.EnqueueRequestForObject{},
		predicate.NewPredicateFuncs(r.hasMatchingController),
	); err != nil {
		return err
	}
	r.log.Info("watching gateway objects")

	return nil
}

// hasMatchingController returns true if the provided object is a Gateway
// using a GatewayClass matching the configured gatewayclass controller name.
func (r *gatewayReconciler) hasMatchingController(obj client.Object) bool {
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

// Reconcile finds all the Gateways for the GatewayClass with an "Accepted: true" condition
// and passes all Gateways for the configured GatewayClass to the IR for processing.
func (r *gatewayReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	r.log.Info("reconciling gateway", "namespace", request.Namespace, "name", request.Name)

	var allClasses *gwapiv1b1.GatewayClassList
	if err := r.client.List(ctx, allClasses); err != nil {
		return reconcile.Result{}, fmt.Errorf("error listing gatewayclasses")
	}
	// Find the GatewayClass for this controller with Accepted=true status condition.
	acceptedClass := r.acceptedClass(allClasses)
	if acceptedClass == nil {
		r.log.Info("No accepted gatewayclass found for gateway", "namespace", request.Namespace,
			"name", request.Name)
		for namespacedName := range r.resourceTable.Gateways.LoadAll() {
			r.resourceTable.Gateways.Delete(namespacedName)
		}
		return reconcile.Result{}, nil
	}

	var allGateways *gwapiv1b1.GatewayList
	if err := r.client.List(ctx, allGateways); err != nil {
		return reconcile.Result{}, fmt.Errorf("error listing gateways")
	}

	// Get all the Gateways for the Accepted=true GatewayClass.
	acceptedGateways := gatewaysOfClass(acceptedClass, allGateways)
	if len(acceptedGateways) == 0 {
		r.log.Info("No gateways found for accepted gatewayclass")
	}
	found := false
	for i := range acceptedGateways {
		key := types.NamespacedName{
			Name:      acceptedGateways[i].GetName(),
			Namespace: acceptedGateways[i].GetNamespace(),
		}
		r.resourceTable.Gateways.Store(key, &acceptedGateways[i])
		if key == request.NamespacedName {
			found = true
		}
	}
	if !found {
		r.resourceTable.Gateways.Delete(request.NamespacedName)
	}

	// Once we've processed `allGateways`, record that we've fully initialized.
	defer r.initializeOnce.Do(r.resourceTable.Initialized.Done)

	return reconcile.Result{}, nil
}

// acceptedClass returns the GatewayClass from the provided list that matches
// the configured controller name and contains the Accepted=true status condition.
func (r *gatewayReconciler) acceptedClass(gcList *gwapiv1b1.GatewayClassList) *gwapiv1b1.GatewayClass {
	if gcList == nil {
		return nil
	}
	for i := range gcList.Items {
		gc := &gcList.Items[i]
		if gc.Spec.ControllerName == r.classController && isAccepted(gc) {
			return gc
		}
	}
	return nil
}

// isAccepted returns true if the provided gatewayclass contains the Accepted=true
// status condition.
func isAccepted(gc *gwapiv1b1.GatewayClass) bool {
	if gc == nil {
		return false
	}
	for _, cond := range gc.Status.Conditions {
		if cond.Type == string(gwapiv1b1.GatewayClassConditionStatusAccepted) && cond.Status == metav1.ConditionTrue {
			return true
		}
	}
	return false
}

// gatewaysOfClass returns a list of gateways that reference gc from the provided gwList.
func gatewaysOfClass(gc *gwapiv1b1.GatewayClass, gwList *gwapiv1b1.GatewayList) []gwapiv1b1.Gateway {
	var ret []gwapiv1b1.Gateway
	if gwList == nil || gc == nil {
		return ret
	}
	for i := range gwList.Items {
		gw := gwList.Items[i]
		if string(gw.Spec.GatewayClassName) == gc.Name {
			ret = append(ret, gw)
		}
	}
	return ret
}
