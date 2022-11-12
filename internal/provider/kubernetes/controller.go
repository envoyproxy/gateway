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
	serviceTLSRouteIndex  = "serviceTLSRouteBackendRef"
	serviceHTTPRouteIndex = "serviceHTTPRouteBackendRef"
	secretGatewayIndex    = "secretGatewayIndex"
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
		handler.EnqueueRequestsFromMapFunc(r.processGateway),
	); err != nil {
		return err
	}
	if err := addGatewayIndexers(ctx, mgr); err != nil {
		return err
	}

	// Watch HTTPRoute CRUDs and process affected Gateways.
	if err := c.Watch(
		&source.Kind{Type: &gwapiv1b1.HTTPRoute{}},
		handler.EnqueueRequestsFromMapFunc(r.processHTTPRoute),
	); err != nil {
		return err
	}
	if err := addHTTPRouteIndexers(ctx, mgr); err != nil {
		return err
	}

	// Watch TLSRoute CRUDs and process affected Gateways.
	if err := c.Watch(
		&source.Kind{Type: &gwapiv1a2.TLSRoute{}},
		handler.EnqueueRequestsFromMapFunc(r.processTLSRoute),
	); err != nil {
		return err
	}
	if err := addTLSRouteIndexers(ctx, mgr); err != nil {
		return err
	}

	// Watch Service CRUDs and process affected *Route objects.
	if err := c.Watch(
		&source.Kind{Type: &corev1.Service{}},
		handler.EnqueueRequestsFromMapFunc(r.processService),
	); err != nil {
		return err
	}

	// Watch Secret CRUDs and process affected Gateways.
	if err := c.Watch(
		&source.Kind{Type: &corev1.Secret{}},
		handler.EnqueueRequestsFromMapFunc(r.processSecret),
	); err != nil {
		return err
	}

	// Watch ReferenceGrant CRUDs and process affected Gateways.
	if err := c.Watch(
		&source.Kind{Type: &gwapiv1a2.ReferenceGrant{}},
		handler.EnqueueRequestsFromMapFunc(r.processReferenceGrant),
	); err != nil {
		return err
	}

	// Watch Deployment CRUDs and process affected Gateways.
	if err := c.Watch(
		&source.Kind{Type: &appsv1.Deployment{}},
		handler.EnqueueRequestsFromMapFunc(r.processDeployment),
	); err != nil {
		return err
	}

	r.log.Info("watching gatewayAPI related objects")
	return nil
}

func (r *gatewayAPIReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	r.log.WithName(request.Name).Info("reconciling gatewayclass")

	var gatewayClasses gwapiv1b1.GatewayClassList
	if err := r.client.List(ctx, &gatewayClasses); err != nil {
		return reconcile.Result{}, fmt.Errorf("error listing gatewayclasses: %v", err)
	}

	var cc controlledClasses

	for i := range gatewayClasses.Items {
		if gatewayClasses.Items[i].Spec.ControllerName == r.classController {
			// The gatewayclass was marked for deletion and the finalizer removed,
			// so clean-up dependents.
			if !gatewayClasses.Items[i].DeletionTimestamp.IsZero() &&
				!slice.ContainsString(gatewayClasses.Items[i].Finalizers, gatewayClassFinalizer) {
				r.log.Info("gatewayclass marked for deletion")
				cc.removeMatch(&gatewayClasses.Items[i])
				// Delete the gatewayclass from the watchable map.
				r.resources.GatewayClasses.Delete(request.Name)
				continue
			}

			cc.addMatch(&gatewayClasses.Items[i])
		}
	}

	// The gatewayclass was already deleted/finalized and there are stale queue entries.
	acceptedGC := cc.acceptedClass()
	if acceptedGC == nil {
		r.log.Info("failed to find an accepted gatewayclass")
		return reconcile.Result{}, nil
	}

	// TODO: do not store unless you have a gateway, route and service in the store already,
	// for this gatewayclass
	// if one of them is not present, remove the gatewayclass if it exists.

	// Store the accepted gatewayclass in the watchable map.
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

	r.log.WithName(request.Name).Info("reconciled gatewayclass")
	return reconcile.Result{}, nil
}

// hasMatchingController returns true if the provided object is a GatewayClass
// with a Spec.Controller string matching this Envoy Gateway's controller string,
// or false otherwise.
func (r *gatewayAPIReconciler) hasMatchingController(obj client.Object) bool {
	gc, ok := obj.(*gwapiv1b1.GatewayClass)
	if !ok {
		r.log.Info("bypassing reconciliation due to unexpected object type", "type", obj)
		return false
	}

	if gc.Spec.ControllerName == r.classController {
		r.log.Info("gatewayclass has matching controller name, processing", "name", gc.Name)
		return true
	}

	r.log.Info("bypassing reconciliation due to controller name", "controller", gc.Spec.ControllerName)
	return false
}

// processSecret processes the Secret coming from the watcher and further
// processes parent Gateway objects to eventually reconcile GatewayClass.
func (r *gatewayAPIReconciler) processSecret(obj client.Object) []reconcile.Request {
	r.log.Info("processing secret", "namespace", obj.GetNamespace(), "name", obj.GetName())
	ctx := context.Background()
	requests := []reconcile.Request{}

	secretkey := types.NamespacedName{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}

	secretDeleted := false

	secret := new(corev1.Secret)
	if err := r.client.Get(ctx, secretkey, secret); err != nil {
		if !kerrors.IsNotFound(err) {
			r.log.Error(err, "failed to get secret")
			return requests
		}

		secretDeleted = true
		// Remove the Secret from watchable map.
		r.resources.Secrets.Delete(secretkey)
	}

	if !secretDeleted {
		// Store the Secret in watchable map.
		r.resources.Secrets.Store(secretkey, secret.DeepCopy())
	}

	// Find the Gateways that reference this Secret.
	gatewayList := &gwapiv1b1.GatewayList{}
	if err := r.client.List(context.Background(), gatewayList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(secretGatewayIndex, utils.NamespacedName(obj).String()),
	}); err != nil {
		return requests
	}

	for _, gw := range gatewayList.Items {
		requests = append(requests, r.processGateway(gw.DeepCopy())...)
	}

	return requests
}

// processReferenceGrant processes the ReferenceGrant coming from the watcher
// and further processes parent Gateway objects to eventually reconcile GatewayClass.
func (r *gatewayAPIReconciler) processReferenceGrant(obj client.Object) []reconcile.Request {
	r.log.Info("processing reference grant", "namespace", obj.GetNamespace(), "name", obj.GetName())
	ctx := context.Background()
	requests := []reconcile.Request{}

	refgrantkey := types.NamespacedName{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}

	gatewayReferences := []types.NamespacedName{}
	refGrantDeleted := false

	refgrant := new(gwapiv1a2.ReferenceGrant)
	if err := r.client.Get(ctx, refgrantkey, refgrant); err != nil {
		if !kerrors.IsNotFound(err) {
			r.log.Error(err, "failed to get reference grant")
			return requests
		}

		refGrantDeleted = true
		// Remove the Reference Grant from watchable map.
		if resourceRefGrant, found := r.resources.ReferenceGrants.Load(refgrantkey); found {
			r.resources.ReferenceGrants.Delete(refgrantkey)
			gatewayReferences = append(gatewayReferences, findGatewayReferencesFromRefGrant(resourceRefGrant)...)
		}
	}

	if !refGrantDeleted {
		// Store the Reference Grant in watchable map.
		r.resources.ReferenceGrants.Store(refgrantkey, refgrant.DeepCopy())
		gatewayReferences = append(gatewayReferences, findGatewayReferencesFromRefGrant(refgrant)...)
	}

	for _, gwref := range gatewayReferences {
		var gateway gwapiv1b1.Gateway
		if err := r.client.Get(ctx, gwref, &gateway); err != nil {
			return requests
		}
		requests = append(requests, r.processGateway(gateway.DeepCopy())...)
	}

	return requests
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
				if *cert.Kind == kindSecret {
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
	return nil
}

// removeFinalizer removes the gatewayclass finalizer from the provided gc, if it exists.
func (r *gatewayAPIReconciler) removeFinalizer(ctx context.Context, gc *gwapiv1b1.GatewayClass) error {
	firstAttempt := true
	if err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		if !firstAttempt {
			// Get the resource.
			if err := r.client.Get(context.Background(), utils.NamespacedName(gc), gc); err != nil {
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
			if err := r.client.Get(context.Background(), utils.NamespacedName(gc), gc); err != nil {
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

// acceptedClass returns the GatewayClass from the provided list that matches
// the configured controller name and contains the Accepted=true status condition.
func (r *gatewayAPIReconciler) acceptedClass(gcList *gwapiv1b1.GatewayClassList) *gwapiv1b1.GatewayClass {
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
