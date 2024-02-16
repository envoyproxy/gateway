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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"
	mcsapi "sigs.k8s.io/mcs-api/pkg/apis/v1alpha1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/api/v1alpha1/validation"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/provider/utils"
	"github.com/envoyproxy/gateway/internal/status"
	"github.com/envoyproxy/gateway/internal/utils/slice"
)

type gatewayAPIReconciler struct {
	client          client.Client
	log             logging.Logger
	statusUpdater   status.Updater
	classController gwapiv1.GatewayController
	store           *kubernetesProviderStore
	namespace       string
	namespaceLabel  *metav1.LabelSelector
	envoyGateway    *egv1a1.EnvoyGateway
	mergeGateways   map[string]bool
	resources       *message.ProviderResources
	extGVKs         []schema.GroupVersionKind
}

// newGatewayAPIController
func newGatewayAPIController(mgr manager.Manager, cfg *config.Server, su status.Updater,
	resources *message.ProviderResources) error {
	ctx := context.Background()

	// Gather additional resources to watch from registered extensions
	var extGVKs []schema.GroupVersionKind
	if cfg.EnvoyGateway.ExtensionManager != nil {
		for _, rsrc := range cfg.EnvoyGateway.ExtensionManager.Resources {
			gvk := schema.GroupVersionKind(rsrc)
			extGVKs = append(extGVKs, gvk)
		}
	}

	byNamespaceSelector := cfg.EnvoyGateway.Provider != nil &&
		cfg.EnvoyGateway.Provider.Kubernetes != nil &&
		cfg.EnvoyGateway.Provider.Kubernetes.Watch != nil &&
		cfg.EnvoyGateway.Provider.Kubernetes.Watch.Type == egv1a1.KubernetesWatchModeTypeNamespaceSelector &&
		(cfg.EnvoyGateway.Provider.Kubernetes.Watch.NamespaceSelector.MatchLabels != nil ||
			len(cfg.EnvoyGateway.Provider.Kubernetes.Watch.NamespaceSelector.MatchExpressions) > 0)

	r := &gatewayAPIReconciler{
		client:          mgr.GetClient(),
		log:             cfg.Logger,
		classController: gwapiv1.GatewayController(cfg.EnvoyGateway.Gateway.ControllerName),
		namespace:       cfg.Namespace,
		statusUpdater:   su,
		resources:       resources,
		extGVKs:         extGVKs,
		store:           newProviderStore(),
		envoyGateway:    cfg.EnvoyGateway,
		mergeGateways:   map[string]bool{},
	}

	if byNamespaceSelector {
		r.namespaceLabel = cfg.EnvoyGateway.Provider.Kubernetes.Watch.NamespaceSelector
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
	// Map for storing backendRefs' NamespaceNames referred by various Route objects.
	allAssociatedBackendRefs map[gwapiv1.BackendObjectReference]struct{}
	// Map for storing referenceGrant NamespaceNames for BackendRefs, SecretRefs, ConfigMapRefs.
	allAssociatedRefGrants map[types.NamespacedName]*gwapiv1b1.ReferenceGrant
	// extensionRefFilters is a map of filters managed by an extension.
	// The key is the namespaced name of the filter and the value is the
	// unstructured form of the resource.
	extensionRefFilters map[types.NamespacedName]unstructured.Unstructured
}

func newResourceMapping() *resourceMappings {
	return &resourceMappings{
		allAssociatedNamespaces:  map[string]struct{}{},
		allAssociatedBackendRefs: map[gwapiv1.BackendObjectReference]struct{}{},
		allAssociatedRefGrants:   map[types.NamespacedName]*gwapiv1b1.ReferenceGrant{},
		extensionRefFilters:      map[types.NamespacedName]unstructured.Unstructured{},
	}
}

// Reconcile handles reconciling all resources in a single call. Any resource event should enqueue the
// same reconcile.Request containing the gateway controller name. This allows multiple resource updates to
// be handled by a single call to Reconcile. The reconcile.Request DOES NOT map to a specific resource.
func (r *gatewayAPIReconciler) Reconcile(ctx context.Context, _ reconcile.Request) (reconcile.Result, error) {
	r.log.Info("reconciling gateways")

	var gatewayClasses gwapiv1.GatewayClassList
	if err := r.client.List(ctx, &gatewayClasses); err != nil {
		return reconcile.Result{}, fmt.Errorf("error listing gatewayclasses: %w", err)
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
	acceptedGCs := cc.matchedClasses
	if acceptedGCs == nil {
		r.log.Info("no accepted gatewayclass")
		return reconcile.Result{}, nil
	}

	resourcesMap := make(gatewayapi.GatewayClassResources)
	for _, acceptedGC := range acceptedGCs {
		// Initialize resource types.
		acceptedGC := acceptedGC
		resourcesMap[acceptedGC.Name] = gatewayapi.NewResources()
		resourceMappings := newResourceMapping()

		if err := r.processGateways(ctx, acceptedGC, resourceMappings, resourcesMap[acceptedGC.Name]); err != nil {
			return reconcile.Result{}, err
		}

		for backendRef := range resourceMappings.allAssociatedBackendRefs {
			backendRefKind := gatewayapi.KindDerefOr(backendRef.Kind, gatewayapi.KindService)
			r.log.Info("processing Backend", "kind", backendRefKind, "namespace", string(*backendRef.Namespace),
				"name", string(backendRef.Name))

			var endpointSliceLabelKey string
			switch backendRefKind {
			case gatewayapi.KindService:
				service := new(corev1.Service)
				err := r.client.Get(ctx, types.NamespacedName{Namespace: string(*backendRef.Namespace), Name: string(backendRef.Name)}, service)
				if err != nil {
					r.log.Error(err, "failed to get Service", "namespace", string(*backendRef.Namespace),
						"name", string(backendRef.Name))
				} else {
					resourceMappings.allAssociatedNamespaces[service.Namespace] = struct{}{}
					resourcesMap[acceptedGC.Name].Services = append(resourcesMap[acceptedGC.Name].Services, service)
					r.log.Info("added Service to resource tree", "namespace", string(*backendRef.Namespace),
						"name", string(backendRef.Name))
				}
				endpointSliceLabelKey = discoveryv1.LabelServiceName

			case gatewayapi.KindServiceImport:
				serviceImport := new(mcsapi.ServiceImport)
				err := r.client.Get(ctx, types.NamespacedName{Namespace: string(*backendRef.Namespace), Name: string(backendRef.Name)}, serviceImport)
				if err != nil {
					r.log.Error(err, "failed to get ServiceImport", "namespace", string(*backendRef.Namespace),
						"name", string(backendRef.Name))
				} else {
					resourceMappings.allAssociatedNamespaces[serviceImport.Namespace] = struct{}{}
					resourcesMap[acceptedGC.Name].ServiceImports = append(resourcesMap[acceptedGC.Name].ServiceImports, serviceImport)
					r.log.Info("added ServiceImport to resource tree", "namespace", string(*backendRef.Namespace),
						"name", string(backendRef.Name))
				}
				endpointSliceLabelKey = mcsapi.LabelServiceName
			}

			// Retrieve the EndpointSlices associated with the service
			endpointSliceList := new(discoveryv1.EndpointSliceList)
			opts := []client.ListOption{
				client.MatchingLabels(map[string]string{
					endpointSliceLabelKey: string(backendRef.Name),
				}),
				client.InNamespace(string(*backendRef.Namespace)),
			}
			if err := r.client.List(ctx, endpointSliceList, opts...); err != nil {
				r.log.Error(err, "failed to get EndpointSlices", "namespace", string(*backendRef.Namespace),
					backendRefKind, string(backendRef.Name))
			} else {
				for _, endpointSlice := range endpointSliceList.Items {
					endpointSlice := endpointSlice
					r.log.Info("added EndpointSlice to resource tree", "namespace", endpointSlice.Namespace,
						"name", endpointSlice.Name)
					resourcesMap[acceptedGC.Name].EndpointSlices = append(resourcesMap[acceptedGC.Name].EndpointSlices, &endpointSlice)
				}
			}
		}

		// Add all ReferenceGrants to the resourceTree
		for _, referenceGrant := range resourceMappings.allAssociatedRefGrants {
			resourcesMap[acceptedGC.Name].ReferenceGrants = append(resourcesMap[acceptedGC.Name].ReferenceGrants, referenceGrant)
		}

		// Add all EnvoyPatchPolicies
		envoyPatchPolicies := egv1a1.EnvoyPatchPolicyList{}
		if err := r.client.List(ctx, &envoyPatchPolicies); err != nil {
			return reconcile.Result{}, fmt.Errorf("error listing EnvoyPatchPolicies: %w", err)
		}

		for _, policy := range envoyPatchPolicies.Items {
			policy := policy
			// Discard Status to reduce memory consumption in watchable
			// It will be recomputed by the gateway-api layer
			policy.Status = egv1a1.EnvoyPatchPolicyStatus{}

			resourcesMap[acceptedGC.Name].EnvoyPatchPolicies = append(resourcesMap[acceptedGC.Name].EnvoyPatchPolicies, &policy)
		}

		// Add all ClientTrafficPolicies
		clientTrafficPolicies := egv1a1.ClientTrafficPolicyList{}
		if err := r.client.List(ctx, &clientTrafficPolicies); err != nil {
			return reconcile.Result{}, fmt.Errorf("error listing ClientTrafficPolicies: %w", err)
		}

		for _, policy := range clientTrafficPolicies.Items {
			policy := policy
			// Discard Status to reduce memory consumption in watchable
			// It will be recomputed by the gateway-api layer
			policy.Status = egv1a1.ClientTrafficPolicyStatus{}
			resourcesMap[acceptedGC.Name].ClientTrafficPolicies = append(resourcesMap[acceptedGC.Name].ClientTrafficPolicies, &policy)

		}

		// Add the referenced ConfigMaps in ClientTrafficPolicies to the resourceTree
		r.processCtpConfigMapRefs(ctx, resourcesMap[acceptedGC.Name], resourceMappings)

		// Add all BackendTrafficPolicies
		backendTrafficPolicies := egv1a1.BackendTrafficPolicyList{}
		if err := r.client.List(ctx, &backendTrafficPolicies); err != nil {
			return reconcile.Result{}, fmt.Errorf("error listing BackendTrafficPolicies: %w", err)
		}

		for _, policy := range backendTrafficPolicies.Items {
			policy := policy
			// Discard Status to reduce memory consumption in watchable
			// It will be recomputed by the gateway-api layer
			policy.Status = egv1a1.BackendTrafficPolicyStatus{}
			resourcesMap[acceptedGC.Name].BackendTrafficPolicies = append(resourcesMap[acceptedGC.Name].BackendTrafficPolicies, &policy)
		}

		// Add all SecurityPolicies
		securityPolicies := egv1a1.SecurityPolicyList{}
		if err := r.client.List(ctx, &securityPolicies); err != nil {
			return reconcile.Result{}, fmt.Errorf("error listing SecurityPolicies: %w", err)
		}

		for _, policy := range securityPolicies.Items {
			policy := policy
			// Discard Status to reduce memory consumption in watchable
			// It will be recomputed by the gateway-api layer
			policy.Status = egv1a1.SecurityPolicyStatus{}
			resourcesMap[acceptedGC.Name].SecurityPolicies = append(resourcesMap[acceptedGC.Name].SecurityPolicies, &policy)
		}

		// Add the referenced Secrets in SecurityPolicies to the resourceTree
		r.processSecurityPolicySecretRefs(ctx, resourcesMap[acceptedGC.Name], resourceMappings)

		// For this particular Gateway, and all associated objects, check whether the
		// namespace exists. Add to the resourceTree.
		for ns := range resourceMappings.allAssociatedNamespaces {
			namespace, err := r.getNamespace(ctx, ns)
			if err != nil {
				r.log.Error(err, "unable to find the namespace")
				if kerrors.IsNotFound(err) {
					return reconcile.Result{}, nil
				}
				return reconcile.Result{}, err
			}

			resourcesMap[acceptedGC.Name].Namespaces = append(resourcesMap[acceptedGC.Name].Namespaces, namespace)
		}

		// Process the parametersRef of the accepted GatewayClass.
		if acceptedGC.Spec.ParametersRef != nil && acceptedGC.DeletionTimestamp == nil {
			if err := r.processParamsRef(ctx, acceptedGC, resourcesMap[acceptedGC.Name]); err != nil {
				msg := fmt.Sprintf("%s: %v", status.MsgGatewayClassInvalidParams, err)
				if err := r.updateStatusForGatewayClass(ctx, acceptedGC, false, string(gwapiv1.GatewayClassReasonInvalidParameters), msg); err != nil {
					r.log.Error(err, "unable to update GatewayClass status")
				}
				r.log.Error(err, "failed to process parametersRef for gatewayclass", "name", acceptedGC.Name)
				return reconcile.Result{}, err
			}
		}

		if resourcesMap[acceptedGC.Name].EnvoyProxy != nil && resourcesMap[acceptedGC.Name].EnvoyProxy.Spec.MergeGateways != nil {
			r.mergeGateways[acceptedGC.Name] = *resourcesMap[acceptedGC.Name].EnvoyProxy.Spec.MergeGateways
		}

		if err := r.updateStatusForGatewayClass(ctx, acceptedGC, true, string(gwapiv1.GatewayClassReasonAccepted), status.MsgValidGatewayClass); err != nil {
			r.log.Error(err, "unable to update GatewayClass status")
			return reconcile.Result{}, err
		}

		if len(resourcesMap[acceptedGC.Name].Gateways) == 0 {
			r.log.Info("No gateways found for accepted gatewayclass")

			// If needed, remove the finalizer from the accepted GatewayClass.
			if err := r.removeFinalizer(ctx, acceptedGC); err != nil {
				r.log.Error(err, fmt.Sprintf("failed to remove finalizer from gatewayclass %s",
					acceptedGC.Name))
				return reconcile.Result{}, err
			} else {
				// finalize the accepted GatewayClass.
				if err := r.addFinalizer(ctx, acceptedGC); err != nil {
					r.log.Error(err, fmt.Sprintf("failed adding finalizer to gatewayclass %s",
						acceptedGC.Name))
					return reconcile.Result{}, err
				}
			}
		}
	}
	// The Store is triggered even when there are no Gateways associated to the
	// GatewayClass. This would happen in case the last Gateway is removed and the
	// Store will be required to trigger a cleanup of envoy infra resources.
	r.resources.GatewayAPIResources.Store(string(r.classController), resourcesMap.DeepCopy())

	r.log.Info("reconciled gateways successfully")
	return reconcile.Result{}, nil
}

// processSecurityPolicySecretRefs adds the referenced Secrets in SecurityPolicies
// to the resourceTree
func (r *gatewayAPIReconciler) processSecurityPolicySecretRefs(
	ctx context.Context, resourceTree *gatewayapi.Resources, resourceMap *resourceMappings) {
	for _, policy := range resourceTree.SecurityPolicies {
		oidc := policy.Spec.OIDC

		if oidc != nil {
			if err := r.processSecretRef(
				ctx,
				resourceMap,
				resourceTree,
				gatewayapi.KindSecurityPolicy,
				policy.Namespace,
				policy.Name,
				oidc.ClientSecret); err != nil {
				// we don't return an error here, because we want to continue
				// reconciling the rest of the SecurityPolicies despite that this
				// secret reference is invalid.
				// This SecurityPolicy will be marked as invalid in its status
				// when translating to IR because the referenced secret can't be
				// found.
				r.log.Error(err,
					"failed to process OIDC SecretRef for SecurityPolicy",
					"policy", policy, "secretRef", oidc.ClientSecret)
			}
		}
		basicAuth := policy.Spec.BasicAuth
		if basicAuth != nil {
			if err := r.processSecretRef(
				ctx,
				resourceMap,
				resourceTree,
				gatewayapi.KindSecurityPolicy,
				policy.Namespace,
				policy.Name,
				basicAuth.Users); err != nil {
				r.log.Error(err,
					"failed to process BasicAuth SecretRef for SecurityPolicy",
					"policy", policy, "secretRef", basicAuth.Users)
			}
		}
	}
}

// processSecretRef adds the referenced Secret to the resourceTree if it's valid.
// - If it exists in the same namespace as the owner.
// - If it exists in a different namespace, and there is a ReferenceGrant.
func (r *gatewayAPIReconciler) processSecretRef(
	ctx context.Context,
	resourceMap *resourceMappings,
	resourceTree *gatewayapi.Resources,
	ownerKind string,
	ownerNS string,
	ownerName string,
	secretRef gwapiv1b1.SecretObjectReference,
) error {
	secret := new(corev1.Secret)
	secretNS := gatewayapi.NamespaceDerefOr(secretRef.Namespace, ownerNS)
	err := r.client.Get(ctx,
		types.NamespacedName{Namespace: secretNS, Name: string(secretRef.Name)},
		secret,
	)
	if err != nil && !kerrors.IsNotFound(err) {
		return fmt.Errorf("unable to find the Secret: %s/%s", secretNS, string(secretRef.Name))
	}

	if secretNS != ownerNS {
		from := ObjectKindNamespacedName{
			kind:      ownerKind,
			namespace: ownerNS,
			name:      ownerName,
		}
		to := ObjectKindNamespacedName{
			kind:      gatewayapi.KindSecret,
			namespace: secretNS,
			name:      secret.Name,
		}
		refGrant, err := r.findReferenceGrant(ctx, from, to)
		switch {
		case err != nil:
			return fmt.Errorf("failed to find ReferenceGrant: %w", err)
		case refGrant == nil:
			return fmt.Errorf(
				"no matching ReferenceGrants found: from %s/%s to %s/%s",
				from.kind, from.namespace, to.kind, to.namespace)
		default:
			// RefGrant found
			resourceMap.allAssociatedRefGrants[utils.NamespacedName(refGrant)] = refGrant
			r.log.Info("added ReferenceGrant to resource map", "namespace", refGrant.Namespace,
				"name", refGrant.Name)
		}
	}
	resourceMap.allAssociatedNamespaces[secretNS] = struct{}{} // TODO Zhaohuabing do we need this line?
	resourceTree.Secrets = append(resourceTree.Secrets, secret)
	r.log.Info("processing Secret", "namespace", secretNS, "name", string(secretRef.Name))
	return nil
}

// processCtpConfigMapRefs adds the referenced ConfigMaps in ClientTrafficPolicies
// to the resourceTree
func (r *gatewayAPIReconciler) processCtpConfigMapRefs(
	ctx context.Context, resourceTree *gatewayapi.Resources, resourceMap *resourceMappings) {
	for _, policy := range resourceTree.ClientTrafficPolicies {
		tls := policy.Spec.TLS

		if tls != nil && tls.ClientValidation != nil {
			for _, caCertRef := range tls.ClientValidation.CACertificateRefs {
				if caCertRef.Kind != nil && string(*caCertRef.Kind) == gatewayapi.KindConfigMap {
					if err := r.processConfigMapRef(
						ctx,
						resourceMap,
						resourceTree,
						gatewayapi.KindClientTrafficPolicy,
						policy.Namespace,
						policy.Name,
						caCertRef); err != nil {
						// we don't return an error here, because we want to continue
						// reconciling the rest of the ClientTrafficPolicies despite that this
						// reference is invalid.
						// This ClientTrafficPolicy will be marked as invalid in its status
						// when translating to IR because the referenced configmap can't be
						// found.
						r.log.Error(err,
							"failed to process CACertificateRef for ClientTrafficPolicy",
							"policy", policy, "caCertificateRef", caCertRef.Name)
					}
				} else if caCertRef.Kind == nil || string(*caCertRef.Kind) == gatewayapi.KindSecret {
					if err := r.processSecretRef(
						ctx,
						resourceMap,
						resourceTree,
						gatewayapi.KindClientTrafficPolicy,
						policy.Namespace,
						policy.Name,
						caCertRef); err != nil {
						r.log.Error(err,
							"failed to process CACertificateRef for SecurityPolicy",
							"policy", policy, "caCertificateRef", caCertRef.Name)
					}
				}
			}
		}
	}
}

// processConfigMapRef adds the referenced ConfigMap to the resourceTree if it's valid.
// - If it exists in the same namespace as the owner.
// - If it exists in a different namespace, and there is a ReferenceGrant.
func (r *gatewayAPIReconciler) processConfigMapRef(
	ctx context.Context,
	resourceMap *resourceMappings,
	resourceTree *gatewayapi.Resources,
	ownerKind string,
	ownerNS string,
	ownerName string,
	configMapRef gwapiv1b1.SecretObjectReference,
) error {
	configMap := new(corev1.ConfigMap)
	configMapNS := gatewayapi.NamespaceDerefOr(configMapRef.Namespace, ownerNS)
	err := r.client.Get(ctx,
		types.NamespacedName{Namespace: configMapNS, Name: string(configMapRef.Name)},
		configMap,
	)
	if err != nil && !kerrors.IsNotFound(err) {
		return fmt.Errorf("unable to find the ConfigMap: %s/%s", configMapNS, string(configMapRef.Name))
	}

	if configMapNS != ownerNS {
		from := ObjectKindNamespacedName{
			kind:      ownerKind,
			namespace: ownerNS,
			name:      ownerName,
		}
		to := ObjectKindNamespacedName{
			kind:      gatewayapi.KindConfigMap,
			namespace: configMapNS,
			name:      configMap.Name,
		}
		refGrant, err := r.findReferenceGrant(ctx, from, to)
		switch {
		case err != nil:
			return fmt.Errorf("failed to find ReferenceGrant: %w", err)
		case refGrant == nil:
			return fmt.Errorf(
				"no matching ReferenceGrants found: from %s/%s to %s/%s",
				from.kind, from.namespace, to.kind, to.namespace)
		default:
			// RefGrant found
			resourceMap.allAssociatedRefGrants[utils.NamespacedName(refGrant)] = refGrant
			r.log.Info("added ReferenceGrant to resource map", "namespace", refGrant.Namespace,
				"name", refGrant.Name)
		}
	}
	resourceMap.allAssociatedNamespaces[configMapNS] = struct{}{} // TODO Zhaohuabing do we need this line?
	resourceTree.ConfigMaps = append(resourceTree.ConfigMaps, configMap)
	r.log.Info("processing ConfigMap", "namespace", configMapNS, "name", string(configMapRef.Name))
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

func (r *gatewayAPIReconciler) findReferenceGrant(ctx context.Context, from, to ObjectKindNamespacedName) (*gwapiv1b1.ReferenceGrant, error) {
	refGrantList := new(gwapiv1b1.ReferenceGrantList)
	opts := &client.ListOptions{FieldSelector: fields.OneTermEqualSelector(targetRefGrantRouteIndex, to.kind)}
	if err := r.client.List(ctx, refGrantList, opts); err != nil {
		return nil, fmt.Errorf("failed to list ReferenceGrants: %w", err)
	}

	refGrants := refGrantList.Items
	if r.namespaceLabel != nil {
		var rgs []gwapiv1b1.ReferenceGrant
		for _, refGrant := range refGrants {
			refGrant := refGrant
			if ok, err := r.checkObjectNamespaceLabels(&refGrant); err != nil {
				r.log.Error(err, "failed to check namespace labels for ReferenceGrant %s in namespace %s: %w", refGrant.GetName(), refGrant.GetNamespace())
				continue
			} else if !ok {
				continue
			}
			rgs = append(rgs, refGrant)
		}
		refGrants = rgs
	}

	for _, refGrant := range refGrants {
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

func (r *gatewayAPIReconciler) processGateways(ctx context.Context, acceptedGC *gwapiv1.GatewayClass, resourceMap *resourceMappings, resourceTree *gatewayapi.Resources) error {
	// Find gateways for the acceptedGC
	// Find the Gateways that reference this Class.
	gatewayList := &gwapiv1.GatewayList{}
	if err := r.client.List(ctx, gatewayList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(classGatewayIndex, acceptedGC.Name),
	}); err != nil {
		r.log.Info("no associated Gateways found for GatewayClass", "name", acceptedGC.Name)
		return err
	}

	for _, gtw := range gatewayList.Items {
		gtw := gtw
		if r.namespaceLabel != nil {
			if ok, err := r.checkObjectNamespaceLabels(&gtw); err != nil {
				r.log.Error(err, "failed to check namespace labels for gateway %s in namespace %s: %w", gtw.GetName(), gtw.GetNamespace())
				continue
			} else if !ok {
				continue
			}
		}
		r.log.Info("processing Gateway", "namespace", gtw.Namespace, "name", gtw.Name)
		resourceMap.allAssociatedNamespaces[gtw.Namespace] = struct{}{}

		for _, listener := range gtw.Spec.Listeners {
			listener := listener
			// Get Secret for gateway if it exists.
			if terminatesTLS(&listener) {
				for _, certRef := range listener.TLS.CertificateRefs {
					certRef := certRef
					if refsSecret(&certRef) {
						if err := r.processSecretRef(
							ctx,
							resourceMap,
							resourceTree,
							gatewayapi.KindGateway,
							gtw.Namespace,
							gtw.Name,
							certRef); err != nil {

							r.log.Error(err,
								"failed to process TLS SecretRef for gateway",
								"gateway", gtw, "secretRef", certRef)
						}
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
		gtw.Status = gwapiv1.GatewayStatus{}
		resourceTree.Gateways = append(resourceTree.Gateways, &gtw)
	}

	return nil
}

// removeFinalizer removes the gatewayclass finalizer from the provided gc, if it exists.
func (r *gatewayAPIReconciler) removeFinalizer(ctx context.Context, gc *gwapiv1.GatewayClass) error {
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
func (r *gatewayAPIReconciler) addFinalizer(ctx context.Context, gc *gwapiv1.GatewayClass) error {
	if !slice.ContainsString(gc.Finalizers, gatewayClassFinalizer) {
		base := client.MergeFrom(gc.DeepCopy())
		gc.Finalizers = append(gc.Finalizers, gatewayClassFinalizer)
		if err := r.client.Patch(ctx, gc, base); err != nil {
			return fmt.Errorf("failed to add finalizer to gatewayclass %s: %w", gc.Name, err)
		}
	}
	return nil
}

// watchResources watches gateway api resources.
func (r *gatewayAPIReconciler) watchResources(ctx context.Context, mgr manager.Manager, c controller.Controller) error {
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &gwapiv1.GatewayClass{}),
		handler.EnqueueRequestsFromMapFunc(r.enqueueClass),
		predicate.GenerationChangedPredicate{},
		predicate.NewPredicateFuncs(r.hasMatchingController),
	); err != nil {
		return err
	}

	// Only enqueue EnvoyProxy objects that match this Envoy Gateway's GatewayClass.
	epPredicates := []predicate.Predicate{
		predicate.GenerationChangedPredicate{},
		predicate.ResourceVersionChangedPredicate{},
		predicate.NewPredicateFuncs(r.hasManagedClass),
	}
	if r.namespaceLabel != nil {
		epPredicates = append(epPredicates, predicate.NewPredicateFuncs(r.hasMatchingNamespaceLabels))
	}
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &egv1a1.EnvoyProxy{}),
		handler.EnqueueRequestsFromMapFunc(r.enqueueClass),
		epPredicates...,
	); err != nil {
		return err
	}

	// Watch Gateway CRUDs and reconcile affected GatewayClass.
	gPredicates := []predicate.Predicate{
		predicate.GenerationChangedPredicate{},
		predicate.NewPredicateFuncs(r.validateGatewayForReconcile),
	}
	if r.namespaceLabel != nil {
		gPredicates = append(gPredicates, predicate.NewPredicateFuncs(r.hasMatchingNamespaceLabels))
	}
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &gwapiv1.Gateway{}),
		handler.EnqueueRequestsFromMapFunc(r.enqueueClass),
		gPredicates...,
	); err != nil {
		return err
	}
	if err := addGatewayIndexers(ctx, mgr); err != nil {
		return err
	}

	// Watch HTTPRoute CRUDs and process affected Gateways.
	httprPredicates := []predicate.Predicate{predicate.GenerationChangedPredicate{}}
	if r.namespaceLabel != nil {
		httprPredicates = append(httprPredicates, predicate.NewPredicateFuncs(r.hasMatchingNamespaceLabels))
	}
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &gwapiv1.HTTPRoute{}),
		handler.EnqueueRequestsFromMapFunc(r.enqueueClass),
		httprPredicates...,
	); err != nil {
		return err
	}
	if err := addHTTPRouteIndexers(ctx, mgr); err != nil {
		return err
	}

	// Watch GRPCRoute CRUDs and process affected Gateways.
	grpcrPredicates := []predicate.Predicate{predicate.GenerationChangedPredicate{}}
	if r.namespaceLabel != nil {
		grpcrPredicates = append(grpcrPredicates, predicate.NewPredicateFuncs(r.hasMatchingNamespaceLabels))
	}
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &gwapiv1a2.GRPCRoute{}),
		handler.EnqueueRequestsFromMapFunc(r.enqueueClass),
		grpcrPredicates...,
	); err != nil {
		return err
	}
	if err := addGRPCRouteIndexers(ctx, mgr); err != nil {
		return err
	}

	// Watch TLSRoute CRUDs and process affected Gateways.
	tlsrPredicates := []predicate.Predicate{predicate.GenerationChangedPredicate{}}
	if r.namespaceLabel != nil {
		tlsrPredicates = append(tlsrPredicates, predicate.NewPredicateFuncs(r.hasMatchingNamespaceLabels))
	}
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &gwapiv1a2.TLSRoute{}),
		handler.EnqueueRequestsFromMapFunc(r.enqueueClass),
		tlsrPredicates...,
	); err != nil {
		return err
	}
	if err := addTLSRouteIndexers(ctx, mgr); err != nil {
		return err
	}

	// Watch UDPRoute CRUDs and process affected Gateways.
	udprPredicates := []predicate.Predicate{predicate.GenerationChangedPredicate{}}
	if r.namespaceLabel != nil {
		udprPredicates = append(udprPredicates, predicate.NewPredicateFuncs(r.hasMatchingNamespaceLabels))
	}
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &gwapiv1a2.UDPRoute{}),
		handler.EnqueueRequestsFromMapFunc(r.enqueueClass),
		udprPredicates...,
	); err != nil {
		return err
	}
	if err := addUDPRouteIndexers(ctx, mgr); err != nil {
		return err
	}

	// Watch TCPRoute CRUDs and process affected Gateways.
	tcprPredicates := []predicate.Predicate{predicate.GenerationChangedPredicate{}}
	if r.namespaceLabel != nil {
		tcprPredicates = append(tcprPredicates, predicate.NewPredicateFuncs(r.hasMatchingNamespaceLabels))
	}
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &gwapiv1a2.TCPRoute{}),
		handler.EnqueueRequestsFromMapFunc(r.enqueueClass),
		tcprPredicates...,
	); err != nil {
		return err
	}
	if err := addTCPRouteIndexers(ctx, mgr); err != nil {
		return err
	}

	// Watch Service CRUDs and process affected *Route objects.
	servicePredicates := []predicate.Predicate{predicate.NewPredicateFuncs(r.validateServiceForReconcile)}
	if r.namespaceLabel != nil {
		servicePredicates = append(servicePredicates, predicate.NewPredicateFuncs(r.hasMatchingNamespaceLabels))
	}
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &corev1.Service{}),
		handler.EnqueueRequestsFromMapFunc(r.enqueueClass),
		servicePredicates...,
	); err != nil {
		return err
	}

	serviceImportCRDExists := r.serviceImportCRDExists(mgr)
	if !serviceImportCRDExists {
		r.log.Info("ServiceImport CRD not found, skipping ServiceImport watch")
	}

	// Watch ServiceImport CRUDs and process affected *Route objects.
	if serviceImportCRDExists {
		if err := c.Watch(
			source.Kind(mgr.GetCache(), &mcsapi.ServiceImport{}),
			handler.EnqueueRequestsFromMapFunc(r.enqueueClass),
			predicate.GenerationChangedPredicate{},
			predicate.NewPredicateFuncs(r.validateServiceImportForReconcile)); err != nil {
			// ServiceImport is not available in the cluster, skip the watch and not throw error.
			r.log.Info("unable to watch ServiceImport: %s", err.Error())
		}
	}

	// Watch EndpointSlice CRUDs and process affected *Route objects.
	esPredicates := []predicate.Predicate{
		predicate.GenerationChangedPredicate{},
		predicate.NewPredicateFuncs(r.validateEndpointSliceForReconcile),
	}
	if r.namespaceLabel != nil {
		esPredicates = append(esPredicates, predicate.NewPredicateFuncs(r.hasMatchingNamespaceLabels))
	}
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &discoveryv1.EndpointSlice{}),
		handler.EnqueueRequestsFromMapFunc(r.enqueueClass),
		esPredicates...,
	); err != nil {
		return err
	}

	// Watch Node CRUDs to update Gateway Address exposed by Service of type NodePort.
	// Node creation/deletion and ExternalIP updates would require update in the Gateway
	nPredicates := []predicate.Predicate{
		predicate.GenerationChangedPredicate{},
		predicate.NewPredicateFuncs(r.handleNode),
	}
	if r.namespaceLabel != nil {
		nPredicates = append(nPredicates, predicate.NewPredicateFuncs(r.hasMatchingNamespaceLabels))
	}
	// resource address.
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &corev1.Node{}),
		handler.EnqueueRequestsFromMapFunc(r.enqueueClass),
		nPredicates...,
	); err != nil {
		return err
	}

	// Watch Secret CRUDs and process affected EG CRs (Gateway, SecurityPolicy, more in the future).
	secretPredicates := []predicate.Predicate{
		predicate.GenerationChangedPredicate{},
		predicate.NewPredicateFuncs(r.validateSecretForReconcile),
	}
	if r.namespaceLabel != nil {
		secretPredicates = append(secretPredicates, predicate.NewPredicateFuncs(r.hasMatchingNamespaceLabels))
	}
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &corev1.Secret{}),
		handler.EnqueueRequestsFromMapFunc(r.enqueueClass),
		secretPredicates...,
	); err != nil {
		return err
	}

	// Watch ConfigMap CRUDs and process affected ClienTraffiPolicies.
	configMapPredicates := []predicate.Predicate{
		predicate.GenerationChangedPredicate{},
		predicate.NewPredicateFuncs(r.validateConfigMapForReconcile),
	}
	if r.namespaceLabel != nil {
		configMapPredicates = append(configMapPredicates, predicate.NewPredicateFuncs(r.hasMatchingNamespaceLabels))
	}
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &corev1.ConfigMap{}),
		handler.EnqueueRequestsFromMapFunc(r.enqueueClass),
		configMapPredicates...,
	); err != nil {
		return err
	}

	// Watch ReferenceGrant CRUDs and process affected Gateways.
	rgPredicates := []predicate.Predicate{predicate.GenerationChangedPredicate{}}
	if r.namespaceLabel != nil {
		rgPredicates = append(rgPredicates, predicate.NewPredicateFuncs(r.hasMatchingNamespaceLabels))
	}
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &gwapiv1b1.ReferenceGrant{}),
		handler.EnqueueRequestsFromMapFunc(r.enqueueClass),
		rgPredicates...,
	); err != nil {
		return err
	}
	if err := addReferenceGrantIndexers(ctx, mgr); err != nil {
		return err
	}

	// Watch Deployment CRUDs and process affected Gateways.
	dPredicates := []predicate.Predicate{predicate.NewPredicateFuncs(r.validateDeploymentForReconcile)}
	if r.namespaceLabel != nil {
		dPredicates = append(dPredicates, predicate.NewPredicateFuncs(r.hasMatchingNamespaceLabels))
	}
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &appsv1.Deployment{}),
		handler.EnqueueRequestsFromMapFunc(r.enqueueClass),
		dPredicates...,
	); err != nil {
		return err
	}

	// Watch EnvoyPatchPolicy if enabled in config
	eppPredicates := []predicate.Predicate{predicate.GenerationChangedPredicate{}}
	if r.namespaceLabel != nil {
		eppPredicates = append(eppPredicates, predicate.NewPredicateFuncs(r.hasMatchingNamespaceLabels))
	}
	if r.envoyGateway.ExtensionAPIs != nil && r.envoyGateway.ExtensionAPIs.EnableEnvoyPatchPolicy {
		// Watch EnvoyPatchPolicy CRUDs
		if err := c.Watch(
			source.Kind(mgr.GetCache(), &egv1a1.EnvoyPatchPolicy{}),
			handler.EnqueueRequestsFromMapFunc(r.enqueueClass),
			eppPredicates...,
		); err != nil {
			return err
		}
	}

	// Watch ClientTrafficPolicy
	ctpPredicates := []predicate.Predicate{predicate.GenerationChangedPredicate{}}
	if r.namespaceLabel != nil {
		ctpPredicates = append(ctpPredicates, predicate.NewPredicateFuncs(r.hasMatchingNamespaceLabels))
	}

	if err := c.Watch(
		source.Kind(mgr.GetCache(), &egv1a1.ClientTrafficPolicy{}),
		handler.EnqueueRequestsFromMapFunc(r.enqueueClass),
		ctpPredicates...,
	); err != nil {
		return err
	}

	if err := addCtpIndexers(ctx, mgr); err != nil {
		return err
	}

	// Watch BackendTrafficPolicy
	btpPredicates := []predicate.Predicate{predicate.GenerationChangedPredicate{}}
	if r.namespaceLabel != nil {
		btpPredicates = append(btpPredicates, predicate.NewPredicateFuncs(r.hasMatchingNamespaceLabels))
	}

	if err := c.Watch(
		source.Kind(mgr.GetCache(), &egv1a1.BackendTrafficPolicy{}),
		handler.EnqueueRequestsFromMapFunc(r.enqueueClass),
		btpPredicates...,
	); err != nil {
		return err
	}

	// Watch SecurityPolicy
	spPredicates := []predicate.Predicate{predicate.GenerationChangedPredicate{}}
	if r.namespaceLabel != nil {
		spPredicates = append(spPredicates, predicate.NewPredicateFuncs(r.hasMatchingNamespaceLabels))
	}

	if err := c.Watch(
		source.Kind(mgr.GetCache(), &egv1a1.SecurityPolicy{}),
		handler.EnqueueRequestsFromMapFunc(r.enqueueClass),
		spPredicates...,
	); err != nil {
		return err
	}
	if err := addSecurityPolicyIndexers(ctx, mgr); err != nil {
		return err
	}

	r.log.Info("Watching gatewayAPI related objects")

	// Watch any additional GVKs from the registered extension.
	uPredicates := []predicate.Predicate{predicate.GenerationChangedPredicate{}}
	if r.namespaceLabel != nil {
		uPredicates = append(uPredicates, predicate.NewPredicateFuncs(r.hasMatchingNamespaceLabels))
	}
	for _, gvk := range r.extGVKs {
		u := &unstructured.Unstructured{}
		u.SetGroupVersionKind(gvk)
		if err := c.Watch(source.Kind(mgr.GetCache(), u),
			handler.EnqueueRequestsFromMapFunc(r.enqueueClass),
			uPredicates...,
		); err != nil {
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
	ep, ok := obj.(*egv1a1.EnvoyProxy)
	if !ok {
		panic(fmt.Sprintf("unsupported object type %T", obj))
	}

	// The EnvoyProxy must be in the same namespace as EG.
	if ep.Namespace != r.namespace {
		r.log.Info("envoyproxy namespace does not match Envoy Gateway's namespace",
			"namespace", ep.Namespace, "name", ep.Name)
		return false
	}

	gcList := new(gwapiv1.GatewayClassList)
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
func (r *gatewayAPIReconciler) processParamsRef(ctx context.Context, gc *gwapiv1.GatewayClass, resourceTree *gatewayapi.Resources) error {
	if !refsEnvoyProxy(gc) {
		return fmt.Errorf("unsupported parametersRef for gatewayclass %s", gc.Name)
	}

	epList := new(egv1a1.EnvoyProxyList)

	// The EnvoyProxy must be in the same namespace as EG.
	if err := r.client.List(ctx, epList, &client.ListOptions{Namespace: r.namespace}); err != nil {
		return fmt.Errorf("failed to list envoyproxies in namespace %s: %w", r.namespace, err)
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
				validationErr = fmt.Errorf("invalid envoyproxy: %w", err)
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
		return fmt.Errorf("invalid gatewayclass %s: %w", gc.Name, validationErr)
	}

	return nil
}

// serviceImportCRDExists checks for the existence of the ServiceImport CRD in k8s APIServer before watching it
func (r *gatewayAPIReconciler) serviceImportCRDExists(mgr manager.Manager) bool {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(mgr.GetConfig())
	if err != nil {
		r.log.Error(err, "failed to create discovery client")
	}
	apiResourceList, err := discoveryClient.ServerPreferredResources()
	if err != nil {
		r.log.Error(err, "failed to get API resource list")
	}
	serviceImportFound := false
	for _, list := range apiResourceList {
		for _, resource := range list.APIResources {
			if list.GroupVersion == mcsapi.GroupVersion.String() && resource.Kind == gatewayapi.KindServiceImport {
				serviceImportFound = true
				break
			}
		}
	}

	return serviceImportFound
}
