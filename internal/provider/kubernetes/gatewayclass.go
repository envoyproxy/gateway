// Portions of this code are based on code from Contour, available at:
// https://github.com/projectcontour/contour/blob/main/internal/controller/gatewayclass.go

package kubernetes

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-logr/logr"
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
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/status"
)

type gatewayClassReconciler struct {
	client        client.Client
	controller    gwapiv1b1.GatewayController
	statusUpdater status.Updater
	log           logr.Logger

	initializeOnce sync.Once
	resources      *message.ProviderResources
}

// newGatewayClassController creates the gatewayclass controller. The controller
// will be pre-configured to watch for cluster-scoped GatewayClass objects with
// a controller field that matches name.
func newGatewayClassController(mgr manager.Manager, cfg *config.Server, su status.Updater, resources *message.ProviderResources) error {
	resources.Initialized.Add(1)
	r := &gatewayClassReconciler{
		client:        mgr.GetClient(),
		controller:    gwapiv1b1.GatewayController(cfg.EnvoyGateway.Gateway.ControllerName),
		statusUpdater: su,
		log:           cfg.Logger,
		resources:     resources,
	}

	c, err := controller.New("gatewayclass", mgr, controller.Options{Reconciler: r})
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

	log.Info("bypassing reconciliation due to controller name", "controller", gc.Spec.ControllerName)
	return false
}

func (r *gatewayClassReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	r.log.WithName(request.Name).Info("reconciling gatewayclass")

	var gatewayClasses gwapiv1b1.GatewayClassList
	if err := r.client.List(ctx, &gatewayClasses); err != nil {
		return reconcile.Result{}, fmt.Errorf("error listing gatewayclasses: %v", err)
	}

	var cc controlledClasses

	for i := range gatewayClasses.Items {
		if gatewayClasses.Items[i].Spec.ControllerName == r.controller {
			cc.addMatch(&gatewayClasses.Items[i])
		}
	}

	acceptedGC := cc.acceptedClass()
	// Reset gatewayclasses since this Reconcile function never performs a Delete and
	// we are only interested in the first element.
	r.resources.DeleteGatewayClasses()
	if acceptedGC == nil {
		// A nil gatewayclass removes managed proxy infra, if it exists.
		r.log.Info("failed to find an accepted gatewayclass")
		r.resources.GatewayClasses.Store(request.Name, nil)
		return reconcile.Result{}, nil
	}
	// Store the accepted gatewayclass in the resource map.
	r.resources.GatewayClasses.Store(acceptedGC.GetName(), acceptedGC)

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
			return reconcile.Result{}, err
		}
	}
	if acceptedGC != nil {
		if err := updater(acceptedGC, true); err != nil {
			return reconcile.Result{}, err
		}
	}
	// Once we've iterated over all listed classes, mark that we've fully initialized.
	r.initializeOnce.Do(r.resources.Initialized.Done)

	r.log.WithName(request.Name).Info("reconciled gatewayclass")
	return reconcile.Result{}, nil
}

type controlledClasses struct {
	matchedClasses []*gwapiv1b1.GatewayClass
	oldestClass    *gwapiv1b1.GatewayClass
}

func (cc *controlledClasses) addMatch(gc *gwapiv1b1.GatewayClass) {
	cc.matchedClasses = append(cc.matchedClasses, gc)

	switch {
	case cc.oldestClass == nil:
		cc.oldestClass = gc
	case gc.CreationTimestamp.Time.Before(cc.oldestClass.CreationTimestamp.Time):
		cc.oldestClass = gc
	case gc.CreationTimestamp.Time.Equal(cc.oldestClass.CreationTimestamp.Time) && gc.Name < cc.oldestClass.Name:
		// tie-breaker: first one in alphabetical order is considered oldest/accepted
		cc.oldestClass = gc
	}
}

func (cc *controlledClasses) acceptedClass() *gwapiv1b1.GatewayClass {
	return cc.oldestClass
}

func (cc *controlledClasses) notAcceptedClasses() []*gwapiv1b1.GatewayClass {
	var res []*gwapiv1b1.GatewayClass
	for _, gc := range cc.matchedClasses {
		// skip the oldest one since it will be accepted.
		if gc.Name != cc.oldestClass.Name {
			res = append(res, gc)
		}
	}

	return res
}
