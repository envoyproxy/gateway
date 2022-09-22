// Portions of this code are based on code from Contour, available at:
// https://github.com/projectcontour/contour/blob/main/internal/controller/gateway.go

package kubernetes

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
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
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/provider/utils"
	"github.com/envoyproxy/gateway/internal/status"
	"github.com/envoyproxy/gateway/internal/utils/slice"
)

const gatewayClassFinalizer = gwapiv1b1.GatewayClassFinalizerGatewaysExist

type gatewayReconciler struct {
	client client.Client
	// classController is the configured gatewayclass controller name.
	classController gwapiv1b1.GatewayController
	statusUpdater   status.Updater
	log             logr.Logger

	initializeOnce sync.Once
	resources      *message.ProviderResources
}

// newGatewayController creates a gateway controller. The controller will watch for
// Gateway objects across all namespaces and reconcile those that match the configured
// gatewayclass controller name.
func newGatewayController(mgr manager.Manager, cfg *config.Server, su status.Updater, resources *message.ProviderResources) error {
	resources.GatewaysInitialized.Add(1)
	r := &gatewayReconciler{
		client:          mgr.GetClient(),
		classController: gwapiv1b1.GatewayController(cfg.EnvoyGateway.Gateway.ControllerName),
		statusUpdater:   su,
		log:             cfg.Logger,
		resources:       resources,
	}

	c, err := controller.New("gateway", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}
	r.log.Info("created gateway controller")

	// Subscribe to status updates
	go r.subscribeAndUpdateStatus(context.Background())

	// Only enqueue Gateway objects that match this Envoy Gateway's controller name.
	if err := c.Watch(
		&source.Kind{Type: &gwapiv1b1.Gateway{}},
		&handler.EnqueueRequestForObject{},
		predicate.NewPredicateFuncs(r.hasMatchingController),
	); err != nil {
		return err
	}
	r.log.Info("watching gateway objects")

	// Trigger gateway reconciliation when the Envoy Service or Deployment has changed.
	if err := c.Watch(&source.Kind{Type: &corev1.Service{}}, r.enqueueRequestForOwningGatewayClass()); err != nil {
		return err
	}
	if err := c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, r.enqueueRequestForOwningGatewayClass()); err != nil {
		return err
	}

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

// enqueueRequestForOwningGatewayClass returns an event handler that maps events with
// the GatewayClass owning label to Gateway objects.
func (r *gatewayReconciler) enqueueRequestForOwningGatewayClass() handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(a client.Object) []reconcile.Request {
		labels := a.GetLabels()
		gcName, found := labels[gatewayapi.OwningGatewayLabel]
		if found {
			var reqs []reconcile.Request
			for _, gw := range r.resources.Gateways.LoadAll() {
				if gw != nil && gw.Spec.GatewayClassName == gwapiv1b1.ObjectName(gcName) {
					req := reconcile.Request{
						NamespacedName: types.NamespacedName{
							Namespace: gw.Namespace,
							Name:      gw.Name,
						},
					}
					reqs = append(reqs, req)
					r.log.Info("queueing gateway", "namespace", gw.Namespace, "name", gw.Name)
				}
			}
			return reqs
		}
		return []reconcile.Request{}
	})
}

// Reconcile finds all the Gateways for the GatewayClass with an "Accepted: true" condition
// and passes all Gateways for the configured GatewayClass to the IR for processing.
func (r *gatewayReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	r.log.Info("reconciling gateway", "namespace", request.Namespace, "name", request.Name)

	// Once we've processed `allGateways`, record that we've fully initialized.
	defer r.initializeOnce.Do(r.resources.GatewaysInitialized.Done)

	allClasses := &gwapiv1b1.GatewayClassList{}
	if err := r.client.List(ctx, allClasses); err != nil {
		return reconcile.Result{}, fmt.Errorf("error listing gatewayclasses")
	}
	// Find the GatewayClass for this controller with Accepted=true status condition.
	acceptedClass := r.acceptedClass(allClasses)
	if acceptedClass == nil {
		r.log.Info("No accepted gatewayclass found for gateway", "namespace", request.Namespace,
			"name", request.Name)
		for namespacedName := range r.resources.Gateways.LoadAll() {
			r.resources.Gateways.Delete(namespacedName)
		}
		return reconcile.Result{}, nil
	}

	allGateways := &gwapiv1b1.GatewayList{}
	if err := r.client.List(ctx, allGateways); err != nil {
		return reconcile.Result{}, fmt.Errorf("error listing gateways")
	}

	// Get all the Gateways for the Accepted=true GatewayClass.
	acceptedGateways := gatewaysOfClass(acceptedClass, allGateways)
	if len(acceptedGateways) == 0 {
		r.log.Info("No gateways found for accepted gatewayclass")
		// If needed, remove the finalizer from the accepted GatewayClass.
		if err := r.removeFinalizer(ctx, acceptedClass); err != nil {
			return reconcile.Result{}, fmt.Errorf("failed to remove finalizer from gatewayclass %s: %w",
				acceptedClass.Name, err)
		}
	} else {
		// If needed, finalize the accepted GatewayClass.
		if err := r.addFinalizer(ctx, acceptedClass); err != nil {
			return reconcile.Result{}, fmt.Errorf("failed adding finalizer to gatewayclass %s: %w",
				acceptedClass.Name, err)
		}
	}

	found := false
	for i := range acceptedGateways {
		key := utils.NamespacedName(&acceptedGateways[i])
		r.resources.Gateways.Store(key, &acceptedGateways[i])
		if key == request.NamespacedName {
			found = true
		}
	}
	if !found {
		r.resources.Gateways.Delete(request.NamespacedName)
	}

	// Set status conditions for all accepted gateways.
	for i := range acceptedGateways {
		gw := acceptedGateways[i]

		// Get the status of the Gateway's associated Envoy Deployment.
		deployment, err := r.envoyDeploymentForGateway(ctx, &gw)
		if err != nil {
			r.log.Info("failed to get deployment for gateway",
				"namespace", gw.Namespace, "name", gw.Name)
		}

		// Get the status address of the Gateway's associated Envoy Service.
		svc, err := r.envoyServiceForGateway(ctx, &gw)
		if err != nil {
			r.log.Info("failed to get service for gateway",
				"namespace", gw.Namespace, "name", gw.Name)
		}

		// update scheduled condition
		status.UpdateGatewayStatusScheduledCondition(&gw, true)
		// update address field and ready condition
		status.UpdateGatewayStatusReadyCondition(&gw, svc, deployment)
		// publish status
		key := utils.NamespacedName(&gw)
		r.resources.GatewayStatuses.Store(key, &gw)
	}

	r.log.WithName(request.Namespace).WithName(request.Name).Info("reconciled gateway")

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

// envoyServiceForGateway returns the Envoy service, returning nil if the service doesn't exist.
func (r *gatewayReconciler) envoyServiceForGateway(ctx context.Context, gateway *gwapiv1b1.Gateway) (*corev1.Service, error) {
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

// addFinalizer adds the gatewayclass finalizer to the provided gc, if it doesn't exist.
func (r *gatewayReconciler) addFinalizer(ctx context.Context, gc *gwapiv1b1.GatewayClass) error {
	if !slice.ContainsString(gc.Finalizers, gatewayClassFinalizer) {
		updated := gc.DeepCopy()
		updated.Finalizers = append(updated.Finalizers, gatewayClassFinalizer)
		if err := r.client.Update(ctx, updated); err != nil {
			return fmt.Errorf("failed to add finalizer to gatewayclass %s: %w", gc.Name, err)
		}
	}
	return nil
}

// removeFinalizer removes the gatewayclass finalizer from the provided gc, if it exists.
func (r *gatewayReconciler) removeFinalizer(ctx context.Context, gc *gwapiv1b1.GatewayClass) error {
	if slice.ContainsString(gc.Finalizers, gatewayClassFinalizer) {
		updated := gc.DeepCopy()
		updated.Finalizers = slice.RemoveString(updated.Finalizers, gatewayClassFinalizer)
		if err := r.client.Update(ctx, updated); err != nil {
			return fmt.Errorf("failed to remove finalizer from gatewayclass %s: %w", gc.Name, err)
		}
	}
	return nil
}

// envoyDeploymentForGateway returns the Envoy Deployment, returning nil if the Deployment doesn't exist.
func (r *gatewayReconciler) envoyDeploymentForGateway(ctx context.Context, gateway *gwapiv1b1.Gateway) (*appsv1.Deployment, error) {
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

// subscribeAndUpdateStatus subscribes to gateway status updates and writes it into the
// Kubernetes API Server
func (r *gatewayReconciler) subscribeAndUpdateStatus(ctx context.Context) {
	// Subscribe to resources
	for snapshot := range r.resources.GatewayStatuses.Subscribe(ctx) {
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
				Resource:       new(gwapiv1b1.Gateway),
				Mutator: status.MutatorFunc(func(obj client.Object) client.Object {
					g, ok := obj.(*gwapiv1b1.Gateway)
					if !ok {
						panic(fmt.Sprintf("unsupported object type %T", obj))
					}
					gCopy := g.DeepCopy()
					gCopy.Status.Conditions = status.MergeConditions(gCopy.Status.Conditions, val.Status.Conditions...)
					gCopy.Status.Addresses = val.Status.Addresses
					gCopy.Status.Listeners = val.Status.Listeners
					return gCopy
				}),
			})
		}
	}
	r.log.Info("status subscriber shutting down")
}

func infraServiceName(gateway *gwapiv1b1.Gateway) string {
	return fmt.Sprintf("%s-%s-%s", config.EnvoyServicePrefix, gateway.Namespace, gateway.Name)
}

func infraDeploymentName(gateway *gwapiv1b1.Gateway) string {
	return fmt.Sprintf("%s-%s-%s", config.EnvoyDeploymentPrefix, gateway.Namespace, gateway.Name)
}
