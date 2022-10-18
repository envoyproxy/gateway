// Copyright 2022 Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// This file contains code derived from Contour,
// https://github.com/projectcontour/contour
// from the source file
// https://github.com/projectcontour/contour/blob/main/internal/controller/gateway.go// and is provided here subject to the following:
// Copyright Project Contour Authors
// SPDX-License-Identifier: Apache-2.0

package kubernetes

import (
	"context"
	"fmt"

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
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
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

	resources *message.ProviderResources
}

// newGatewayController creates a gateway controller. The controller will watch for
// Gateway objects across all namespaces and reconcile those that match the configured
// gatewayclass controller name.
func newGatewayController(mgr manager.Manager, cfg *config.Server, su status.Updater, resources *message.ProviderResources) error {
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
	if err := c.Watch(&source.Kind{Type: &corev1.Service{}}, r.enqueueRequestForOwningGateway()); err != nil {
		return err
	}
	if err := c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, r.enqueueRequestForOwningGateway()); err != nil {
		return err
	}
	// Trigger gateway reconciliation when a Secret that is referenced
	// by a managed Gateway has changed.
	if err := c.Watch(&source.Kind{Type: &corev1.Secret{}}, r.enqueueRequestForGatewaySecrets()); err != nil {
		return err
	}
	// Trigger gateway reconciliation when a ReferenceGrant that refers
	// to a managed Gateway has changed.
	if err := c.Watch(&source.Kind{Type: &gwapiv1a2.ReferenceGrant{}}, r.enqueueRequestForReferencedGateway()); err != nil {
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

// enqueueRequestForOwningGateway returns an event handler that maps events for
// resources with Gateway owning labels to reconcile requests for those Gateway objects.
func (r *gatewayReconciler) enqueueRequestForOwningGateway() handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(a client.Object) []reconcile.Request {
		labels := a.GetLabels()
		if labels == nil {
			return nil
		}

		gatewayNamespace := labels[gatewayapi.OwningGatewayNamespaceLabel]
		gatewayName := labels[gatewayapi.OwningGatewayNameLabel]

		if len(gatewayNamespace) == 0 || len(gatewayName) == 0 {
			return nil
		}

		return []reconcile.Request{
			{
				NamespacedName: types.NamespacedName{
					Namespace: gatewayNamespace,
					Name:      gatewayName,
				},
			},
		}
	})
}

// enqueueRequestForGatewaySecrets returns an event handler that maps events for
// Secrets referenced by managed Gateways to reconcile requests for those Gateway objects.
func (r *gatewayReconciler) enqueueRequestForGatewaySecrets() handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(a client.Object) []reconcile.Request {
		secret, ok := a.(*corev1.Secret)
		if !ok {
			r.log.Info("bypassing reconciliation due to unexpected object type", "type", a)
			return nil
		}

		ctx := context.Background()
		var gateways gwapiv1b1.GatewayList
		if err := r.client.List(ctx, &gateways); err != nil {
			return nil
		}

		var reqs []reconcile.Request
		for i := range gateways.Items {
			gw := gateways.Items[i]
			if r.hasMatchingController(&gw) {
				for j := range gw.Spec.Listeners {
					if terminatesTLS(&gw.Spec.Listeners[j]) {
						secrets, _, err := r.secretsAndRefGrantsForGateway(ctx, &gw)
						if err != nil {
							return nil
						}
						for _, s := range secrets {
							if s.Namespace == secret.Namespace && s.Name == secret.Name {
								req := reconcile.Request{
									NamespacedName: types.NamespacedName{
										Namespace: gw.Namespace,
										Name:      gw.Name,
									},
								}
								reqs = append(reqs, req)
							}
						}
					}
				}
			}
		}

		return reqs
	})
}

// enqueueRequestForReferencedGateway returns an event handler that maps events for
// resources that reference a managed Gateway to reconcile requests for those Gateway objects.
// Note: A ReferenceGrant is the only supported object type.
func (r *gatewayReconciler) enqueueRequestForReferencedGateway() handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(a client.Object) []reconcile.Request {
		rg, ok := a.(*gwapiv1a2.ReferenceGrant)
		if !ok {
			r.log.Info("bypassing reconciliation due to unexpected object type", "type", a)
			return nil
		}

		var refs []types.NamespacedName
		for _, to := range rg.Spec.To {
			if to.Group == gwapiv1a2.GroupName &&
				to.Kind == gatewayapi.KindGateway &&
				to.Name != nil {
				ref := types.NamespacedName{Namespace: rg.Namespace, Name: string(*to.Name)}
				refs = append(refs, ref)
			}
		}
		for _, from := range rg.Spec.From {
			if from.Group == gwapiv1a2.GroupName &&
				from.Kind == gatewayapi.KindGateway {
				ref := types.NamespacedName{Namespace: string(from.Namespace), Name: rg.Name}
				refs = append(refs, ref)
			}
		}

		ctx := context.Background()
		var gateways gwapiv1b1.GatewayList
		if err := r.client.List(ctx, &gateways); err != nil {
			return nil
		}

		var reqs []reconcile.Request
		for i := range gateways.Items {
			gw := gateways.Items[i]
			for _, ref := range refs {
				if gw.Namespace == ref.Namespace && gw.Name == ref.Name && r.hasMatchingController(&gw) {
					req := reconcile.Request{
						NamespacedName: types.NamespacedName{
							Namespace: gw.Namespace,
							Name:      gw.Name,
						},
					}
					reqs = append(reqs, req)
				}
			}
		}

		return reqs
	})
}

// Reconcile finds all the Gateways for the GatewayClass with an "Accepted: true" condition
// and passes all Gateways for the configured GatewayClass to the IR for processing.
func (r *gatewayReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	r.log.Info("reconciling gateway", "namespace", request.Namespace, "name", request.Name)

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

		// Get the secret and referenceGrants of the Gateway's TLS configuration.
		secrets, refGrants, err := r.secretsAndRefGrantsForGateway(ctx, &gw)
		if err != nil {
			r.log.Info("failed to get secrets and referencegrants for gateway",
				"namespace", gw.Namespace, "name", gw.Name)
		}
		for i := range secrets {
			secret := secrets[i]
			// Store the secrets in the resource map.
			key := utils.NamespacedName(&secret)
			r.resources.Secrets.Store(key, &secret)
		}
		for i := range refGrants {
			rg := refGrants[i]
			// Store the referencegrants in the resource map.
			key := utils.NamespacedName(&rg)
			r.resources.ReferenceGrants.Store(key, &rg)
			// Store the referencegrant namespace in the resource map.
			key = types.NamespacedName{Name: rg.Namespace}
			refNs := corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: rg.Namespace,
				},
			}
			if err := r.client.Get(ctx, key, &refNs); err != nil {
				r.log.Info("failed to get referencegrant namespace", "name", refNs.Name)
				return reconcile.Result{}, nil
			}
			r.resources.Namespaces.Store(refNs.Name, &refNs)
		}

		// update scheduled condition
		status.UpdateGatewayStatusScheduledCondition(&gw, true)
		// update address field and ready condition
		status.UpdateGatewayStatusReadyCondition(&gw, svc, deployment)

		key := utils.NamespacedName(&gw)
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
				gCopy.Status.Conditions = status.MergeConditions(gCopy.Status.Conditions, gw.Status.Conditions...)
				gCopy.Status.Addresses = gw.Status.Addresses
				return gCopy

			}),
		})

		// only store the resource if it does not exist or it has a newer spec.
		if v, ok := r.resources.Gateways.Load(key); !ok || (gw.Generation > v.Generation) {
			r.resources.Gateways.Store(key, &gw)
		}
		if key == request.NamespacedName {
			found = true
		}
	}

	if !found {
		gw, ok := r.resources.Gateways.Load(request.NamespacedName)
		if !ok {
			r.log.Info("failed to find accepted gateway in the watchable map", "namespace", request.Namespace, "name", request.Name)
			return reconcile.Result{}, nil
		}

		r.resources.Gateways.Delete(request.NamespacedName)
		// Delete the TLS secrets from the resource map if no other managed
		// Gateways reference them.
		secrets, _, err := r.secretsAndRefGrantsForGateway(ctx, gw)
		if err != nil {
			r.log.Info("failed to get secrets and referencegrants for gateway",
				"namespace", gw.Namespace, "name", gw.Name)
		}
		for i := range secrets {
			secret := secrets[i]
			referenced, err := r.gatewaysRefSecret(ctx, &secret)
			switch {
			case err != nil:
				r.log.Error(err, "failed to verify if other gateways reference secret")
			case !referenced:
				r.log.Info("no other gateways reference secret; deleting from resource map",
					"namespace", secret.Namespace, "name", secret.Name)
				key := utils.NamespacedName(&secret)
				r.resources.Secrets.Delete(key)
			default:
				r.log.Info("other gateways reference secret; keeping the secret in the resource map",
					"namespace", secret.Namespace, "name", secret.Name)
			}
		}
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

// gatewaysRefSecret returns true if a managed Gateway references the provided secret.
// An error is returned if an error is encountered while checking.
func (r *gatewayReconciler) gatewaysRefSecret(ctx context.Context, secret *corev1.Secret) (bool, error) {
	if secret == nil {
		return false, fmt.Errorf("secret is nil")
	}
	gateways := &gwapiv1b1.GatewayList{}
	if err := r.client.List(ctx, gateways); err != nil {
		return false, fmt.Errorf("error listing gatewayclasses: %v", err)
	}
	for i := range gateways.Items {
		gw := gateways.Items[i]
		if r.hasMatchingController(&gw) {
			secrets, _, err := r.secretsAndRefGrantsForGateway(ctx, &gw)
			if err != nil {
				return false, err
			}
			for _, s := range secrets {
				if s.Namespace == secret.Namespace && s.Name == secret.Name {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

// secretsAndRefGrantsForGateway returns the Secrets referenced by the provided gateway listeners.
// If the provided Gateway references a Secret in a different namespace, a list of
// ReferenceGrants is returned that permit the cross namespace Secret reference.
func (r *gatewayReconciler) secretsAndRefGrantsForGateway(ctx context.Context, gateway *gwapiv1b1.Gateway) ([]corev1.Secret, []gwapiv1a2.ReferenceGrant, error) {
	var secrets []corev1.Secret
	var returnedGrants []gwapiv1a2.ReferenceGrant
	for i := range gateway.Spec.Listeners {
		listener := gateway.Spec.Listeners[i]
		if terminatesTLS(&listener) {
			for j := range listener.TLS.CertificateRefs {
				ref := listener.TLS.CertificateRefs[j]
				if refsSecret(&ref) {
					if ref.Namespace != nil {
						// A ReferenceGrant is required for cross namespace secret references.
						refGrants := &gwapiv1a2.ReferenceGrantList{}
						opts := client.ListOptions{Namespace: string(*ref.Namespace)}
						if err := r.client.List(ctx, refGrants, &opts); err != nil {
							return nil, nil, fmt.Errorf("error listing referencegrants")
						}
						var gwRefd, secretRefd bool
						for _, rg := range refGrants.Items {
							for _, from := range rg.Spec.From {
								if from.Group == gwapiv1a2.GroupName &&
									from.Kind == gatewayapi.KindGateway {
									gwRefd = true
									break
								}
							}
							for _, to := range rg.Spec.To {
								if to.Group == corev1.GroupName &&
									to.Kind == gatewayapi.KindSecret {
									if to.Name == nil || *to.Name == gwapiv1a2.ObjectName(ref.Name) {
										secretRefd = true
										break
									}
								}
							}
							if gwRefd && secretRefd {
								returnedGrants = append(returnedGrants, rg)
								key := types.NamespacedName{
									Namespace: string(*ref.Namespace),
									Name:      string(ref.Name),
								}
								secret := new(corev1.Secret)
								if err := r.client.Get(ctx, key, secret); err != nil {
									r.resources.Secrets.Delete(key)
									return nil, nil, fmt.Errorf("failed to get secret: %v", err)
								}
								secrets = append(secrets, *secret)
							}
						}
					} else {
						// The secret is in the Gateway's namespace.
						key := types.NamespacedName{
							Namespace: gateway.Namespace,
							Name:      string(ref.Name),
						}
						secret := new(corev1.Secret)
						if err := r.client.Get(ctx, key, secret); err != nil {
							r.resources.Secrets.Delete(key)
							return nil, nil, fmt.Errorf("failed to get secret: %v", err)
						}
						secrets = append(secrets, *secret)
					}
				}
			}
		}
	}

	return secrets, returnedGrants, nil
}

// terminatesTLS returns true if the provided gateway contains a listener configured
// for TLS termination.
func terminatesTLS(listener *gwapiv1b1.Listener) bool {
	if listener.TLS != nil &&
		listener.Protocol == gwapiv1b1.HTTPSProtocolType &&
		listener.TLS.Mode != nil &&
		*listener.TLS.Mode == gwapiv1b1.TLSModeTerminate {
		return true
	}
	return false
}

// refsSecret returns true if ref refers to a Secret.
func refsSecret(ref *gwapiv1b1.SecretObjectReference) bool {
	return (ref.Group == nil || *ref.Group == corev1.GroupName) &&
		(ref.Kind == nil || *ref.Kind == gatewayapi.KindSecret)
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
	r.log.Info("status subscriber shutting down")
}

func infraServiceName(gateway *gwapiv1b1.Gateway) string {
	infraName := utils.GetHashedName(fmt.Sprintf("%s-%s", gateway.Namespace, gateway.Name))
	return fmt.Sprintf("%s-%s", config.EnvoyPrefix, infraName)
}

func infraDeploymentName(gateway *gwapiv1b1.Gateway) string {
	infraName := utils.GetHashedName(fmt.Sprintf("%s-%s", gateway.Namespace, gateway.Name))
	return fmt.Sprintf("%s-%s", config.EnvoyPrefix, infraName)
}
