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
	"k8s.io/apimachinery/pkg/util/sets"
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
	gwapiv1a3 "sigs.k8s.io/gateway-api/apis/v1alpha3"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"
	mcsapi "sigs.k8s.io/mcs-api/pkg/apis/v1alpha1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/api/v1alpha1/validation"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/utils"
	"github.com/envoyproxy/gateway/internal/utils/slice"
)

type gatewayAPIReconciler struct {
	client          client.Client
	log             logging.Logger
	statusUpdater   Updater
	classController gwapiv1.GatewayController
	store           *kubernetesProviderStore
	namespace       string
	namespaceLabel  *metav1.LabelSelector
	envoyGateway    *egv1a1.EnvoyGateway
	mergeGateways   sets.Set[string]
	resources       *message.ProviderResources
	extGVKs         []schema.GroupVersionKind
}

// newGatewayAPIController
func newGatewayAPIController(mgr manager.Manager, cfg *config.Server, su Updater,
	resources *message.ProviderResources,
) error {
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
		mergeGateways:   sets.New[string](),
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
	// extensionRefFilters is a map of filters managed by an extension.
	// The key is the namespaced name of the filter and the value is the
	// unstructured form of the resource.
	extensionRefFilters map[types.NamespacedName]unstructured.Unstructured
}

func newResourceMapping() *resourceMappings {
	return &resourceMappings{
		allAssociatedNamespaces:  map[string]struct{}{},
		allAssociatedBackendRefs: map[gwapiv1.BackendObjectReference]struct{}{},
		extensionRefFilters:      map[types.NamespacedName]unstructured.Unstructured{},
	}
}

// Reconcile handles reconciling all resources in a single call. Any resource event should enqueue the
// same reconcile.Request containing the gateway controller name. This allows multiple resource updates to
// be handled by a single call to Reconcile. The reconcile.Request DOES NOT map to a specific resource.
func (r *gatewayAPIReconciler) Reconcile(ctx context.Context, _ reconcile.Request) (reconcile.Result, error) {
	var (
		managedGCs []*gwapiv1.GatewayClass
		err        error
	)
	r.log.Info("reconciling gateways")

	// Get the GatewayClasses managed by the Envoy Gateway Controller.
	managedGCs, err = r.managedGatewayClasses(ctx)
	if err != nil {
		return reconcile.Result{}, err
	}

	// The gatewayclass was already deleted/finalized and there are stale queue entries.
	if managedGCs == nil {
		r.resources.GatewayAPIResources.Delete(string(r.classController))
		r.log.Info("no accepted gatewayclass")
		return reconcile.Result{}, nil
	}

	// Collect all the Gateway API resources, Envoy Gateway customized resources,
	// and their referenced resources for the managed GatewayClasses, and store
	// them per GatewayClass.
	// For example:
	// - Gateway API resources: Gateways, xRoutes ...
	// - Envoy Gateway customized resources: EnvoyPatchPolicies, ClientTrafficPolicies, BackendTrafficPolicies ...
	// - Referenced resources: Services, ServiceImports, EndpointSlices, Secrets, ConfigMaps ...
	gwcResources := make(gatewayapi.ControllerResources, 0, len(managedGCs))
	for _, managedGC := range managedGCs {
		// Initialize resource types.
		managedGC := managedGC
		gwcResource := gatewayapi.NewResources()
		gwcResource.GatewayClass = managedGC
		gwcResources = append(gwcResources, gwcResource)
		resourceMappings := newResourceMapping()

		// Add all Gateways, their associated Routes, and referenced resources to the resourceTree
		if err = r.processGateways(ctx, managedGC, resourceMappings, gwcResource); err != nil {
			return reconcile.Result{}, err
		}

		// Add all EnvoyPatchPolicies to the resourceTree
		if err = r.processEnvoyPatchPolicies(ctx, gwcResource); err != nil {
			return reconcile.Result{}, err
		}

		// Add all ClientTrafficPolicies and their referenced resources to the resourceTree
		if err = r.processClientTrafficPolicies(ctx, gwcResource, resourceMappings); err != nil {
			return reconcile.Result{}, err
		}

		// Add all BackendTrafficPolicies to the resourceTree
		if err = r.processBackendTrafficPolicies(ctx, gwcResource); err != nil {
			return reconcile.Result{}, err
		}

		// Add all SecurityPolicies and their referenced resources to the resourceTree
		if err = r.processSecurityPolicies(ctx, gwcResource, resourceMappings); err != nil {
			return reconcile.Result{}, err
		}

		// Add all BackendTLSPolies to the resourceTree
		if err = r.processBackendTLSPolicies(ctx, gwcResource, resourceMappings); err != nil {
			return reconcile.Result{}, err
		}

		// Add all EnvoyExtensionPolicies and their referenced resources to the resourceTree
		if err = r.processEnvoyExtensionPolicies(ctx, gwcResource, resourceMappings); err != nil {
			return reconcile.Result{}, err
		}

		// Add the referenced services, ServiceImports, and EndpointSlices in
		// the collected BackendRefs to the resourceTree.
		// BackendRefs are referred by various Route objects and the ExtAuth in SecurityPolicies.
		r.processBackendRefs(ctx, gwcResource, resourceMappings)

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

			gwcResource.Namespaces = append(gwcResource.Namespaces, namespace)
		}

		// Process the parametersRef of the accepted GatewayClass.
		if managedGC.Spec.ParametersRef != nil && managedGC.DeletionTimestamp == nil {
			if err := r.processParamsRef(ctx, managedGC, gwcResource); err != nil {
				msg := fmt.Sprintf("%s: %v", status.MsgGatewayClassInvalidParams, err)
				if err := r.updateStatusForGatewayClass(ctx, managedGC, false, string(gwapiv1.GatewayClassReasonInvalidParameters), msg); err != nil {
					r.log.Error(err, "unable to update GatewayClass status")
				}
				r.log.Error(err, "failed to process parametersRef for gatewayclass", "name", managedGC.Name)
				return reconcile.Result{}, err
			}
		}

		if gwcResource.EnvoyProxy != nil && gwcResource.EnvoyProxy.Spec.MergeGateways != nil {
			if *gwcResource.EnvoyProxy.Spec.MergeGateways {
				r.mergeGateways.Insert(managedGC.Name)
			} else {
				r.mergeGateways.Delete(managedGC.Name)
			}
		}

		if err := r.updateStatusForGatewayClass(ctx, managedGC, true, string(gwapiv1.GatewayClassReasonAccepted), status.MsgValidGatewayClass); err != nil {
			r.log.Error(err, "unable to update GatewayClass status")
			return reconcile.Result{}, err
		}

		if len(gwcResource.Gateways) == 0 {
			r.log.Info("No gateways found for accepted gatewayclass")

			// If needed, remove the finalizer from the accepted GatewayClass.
			if err := r.removeFinalizer(ctx, managedGC); err != nil {
				r.log.Error(err, fmt.Sprintf("failed to remove finalizer from gatewayclass %s",
					managedGC.Name))
				return reconcile.Result{}, err
			}
		} else {
			// finalize the accepted GatewayClass.
			if err := r.addFinalizer(ctx, managedGC); err != nil {
				r.log.Error(err, fmt.Sprintf("failed adding finalizer to gatewayclass %s",
					managedGC.Name))
				return reconcile.Result{}, err
			}
		}
	}

	// Store the Gateway Resources for the GatewayClass.
	// The Store is triggered even when there are no Gateways associated to the
	// GatewayClass. This would happen in case the last Gateway is removed and the
	// Store will be required to trigger a cleanup of envoy infra resources.
	r.resources.GatewayAPIResources.Store(string(r.classController), &gwcResources)

	r.log.Info("reconciled gateways successfully")
	return reconcile.Result{}, nil
}

// managedGatewayClasses returns a list of GatewayClass objects that are managed by the Envoy Gateway Controller.
func (r *gatewayAPIReconciler) managedGatewayClasses(ctx context.Context) ([]*gwapiv1.GatewayClass, error) {
	var gatewayClasses gwapiv1.GatewayClassList
	if err := r.client.List(ctx, &gatewayClasses); err != nil {
		return nil, fmt.Errorf("error listing gatewayclasses: %w", err)
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
				continue
			}

			cc.addMatch(&gwClass)
		}
	}

	return cc.matchedClasses, nil
}

// processBackendRefs adds the referenced resources in BackendRefs to the resourceTree, including:
// - Services
// - ServiceImports
// - EndpointSlices
func (r *gatewayAPIReconciler) processBackendRefs(ctx context.Context, gwcResource *gatewayapi.Resources, resourceMappings *resourceMappings) {
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
				gwcResource.Services = append(gwcResource.Services, service)
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
				gwcResource.ServiceImports = append(gwcResource.ServiceImports, serviceImport)
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
				gwcResource.EndpointSlices = append(gwcResource.EndpointSlices, &endpointSlice)
			}
		}
	}
}

// processSecurityPolicyObjectRefs adds the referenced resources in SecurityPolicies
// to the resourceTree
// - Secrets for OIDC and BasicAuth
// - BackendRefs for ExAuth
func (r *gatewayAPIReconciler) processSecurityPolicyObjectRefs(
	ctx context.Context, resourceTree *gatewayapi.Resources, resourceMap *resourceMappings,
) {
	// we don't return errors from this method, because we want to continue reconciling
	// the rest of the SecurityPolicies despite that one reference is invalid. This
	// allows Envoy Gateway to continue serving traffic even if some SecurityPolicies
	// are invalid.
	//
	// This SecurityPolicy will be marked as invalid in its status when translating
	// to IR because the referenced secret can't be found.
	for _, policy := range resourceTree.SecurityPolicies {
		oidc := policy.Spec.OIDC

		// Add the referenced Secrets in OIDC to the resourceTree
		if oidc != nil {
			if err := r.processSecretRef(
				ctx,
				resourceMap,
				resourceTree,
				gatewayapi.KindSecurityPolicy,
				policy.Namespace,
				policy.Name,
				oidc.ClientSecret); err != nil {
				r.log.Error(err,
					"failed to process OIDC SecretRef for SecurityPolicy",
					"policy", policy, "secretRef", oidc.ClientSecret)
			}
		}

		// Add the referenced Secrets in BasicAuth to the resourceTree
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

		// Add the referenced BackendRefs and ReferenceGrants in ExtAuth to Maps for later processing
		extAuth := policy.Spec.ExtAuth
		if extAuth != nil {
			var backendRef gwapiv1.BackendObjectReference
			if extAuth.GRPC != nil {
				backendRef = extAuth.GRPC.BackendRef
			} else {
				backendRef = extAuth.HTTP.BackendRef
			}

			backendNamespace := gatewayapi.NamespaceDerefOr(backendRef.Namespace, policy.Namespace)
			resourceMap.allAssociatedBackendRefs[gwapiv1.BackendObjectReference{
				Group:     backendRef.Group,
				Kind:      backendRef.Kind,
				Namespace: gatewayapi.NamespacePtrV1Alpha2(backendNamespace),
				Name:      backendRef.Name,
			}] = struct{}{}

			if backendNamespace != policy.Namespace {
				from := ObjectKindNamespacedName{
					kind:      gatewayapi.KindSecurityPolicy,
					namespace: policy.Namespace,
					name:      policy.Name,
				}
				to := ObjectKindNamespacedName{
					kind:      gatewayapi.KindDerefOr(backendRef.Kind, gatewayapi.KindService),
					namespace: backendNamespace,
					name:      string(backendRef.Name),
				}
				refGrant, err := r.findReferenceGrant(ctx, from, to)
				switch {
				case err != nil:
					r.log.Error(err, "failed to find ReferenceGrant")
				case refGrant == nil:
					r.log.Info("no matching ReferenceGrants found", "from", from.kind,
						"from namespace", from.namespace, "target", to.kind, "target namespace", to.namespace)
				default:
					resourceTree.ReferenceGrants = append(resourceTree.ReferenceGrants, refGrant)
					r.log.Info("added ReferenceGrant to resource map", "namespace", refGrant.Namespace,
						"name", refGrant.Name)
				}
			}
		}
	}
}

// processOIDCHMACSecret adds the OIDC HMAC Secret to the resourceTree.
// The OIDC HMAC Secret is created by the CertGen job and is used by SecurityPolicy
// to configure OAuth2 filters.
func (r *gatewayAPIReconciler) processOIDCHMACSecret(ctx context.Context, resourceTree *gatewayapi.Resources) {
	var (
		secret corev1.Secret
		err    error
	)

	err = r.client.Get(ctx,
		types.NamespacedName{Namespace: r.namespace, Name: oidcHMACSecretName},
		&secret,
	)
	// We don't return an error here, because we want to continue reconciling
	// despite that the OIDC HMAC secret can't be found.
	// If the OIDC HMAC Secret is missing, the SecurityPolicy with OIDC will be
	// marked as invalid in its status when translating to IR.
	if err != nil {
		r.log.Error(err,
			"failed to process OIDC HMAC Secret",
			"namespace", r.namespace, "name", oidcHMACSecretName)
		return
	}

	resourceTree.Secrets = append(resourceTree.Secrets, &secret)
	r.log.Info("processing OIDC HMAC Secret", "namespace", r.namespace, "name", oidcHMACSecretName)
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
			resourceTree.ReferenceGrants = append(resourceTree.ReferenceGrants, refGrant)
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
	ctx context.Context, resourceTree *gatewayapi.Resources, resourceMap *resourceMappings,
) {
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
			resourceTree.ReferenceGrants = append(resourceTree.ReferenceGrants, refGrant)
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

func (r *gatewayAPIReconciler) processGateways(ctx context.Context, managedGC *gwapiv1.GatewayClass, resourceMap *resourceMappings, resourceTree *gatewayapi.Resources) error {
	// Find gateways for the managedGC
	// Find the Gateways that reference this Class.
	gatewayList := &gwapiv1.GatewayList{}
	if err := r.client.List(ctx, gatewayList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(classGatewayIndex, managedGC.Name),
	}); err != nil {
		r.log.Info("no associated Gateways found for GatewayClass", "name", managedGC.Name)
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

// processEnvoyPatchPolicies adds EnvoyPatchPolicies to the resourceTree
func (r *gatewayAPIReconciler) processEnvoyPatchPolicies(ctx context.Context, resourceTree *gatewayapi.Resources) error {
	envoyPatchPolicies := egv1a1.EnvoyPatchPolicyList{}
	if err := r.client.List(ctx, &envoyPatchPolicies); err != nil {
		return fmt.Errorf("error listing EnvoyPatchPolicies: %w", err)
	}

	for _, policy := range envoyPatchPolicies.Items {
		policy := policy
		// Discard Status to reduce memory consumption in watchable
		// It will be recomputed by the gateway-api layer
		policy.Status = gwapiv1a2.PolicyStatus{}

		resourceTree.EnvoyPatchPolicies = append(resourceTree.EnvoyPatchPolicies, &policy)
	}
	return nil
}

// processClientTrafficPolicies adds ClientTrafficPolicies to the resourceTree
func (r *gatewayAPIReconciler) processClientTrafficPolicies(
	ctx context.Context, resourceTree *gatewayapi.Resources, resourceMap *resourceMappings,
) error {
	clientTrafficPolicies := egv1a1.ClientTrafficPolicyList{}
	if err := r.client.List(ctx, &clientTrafficPolicies); err != nil {
		return fmt.Errorf("error listing ClientTrafficPolicies: %w", err)
	}

	for _, policy := range clientTrafficPolicies.Items {
		policy := policy
		// Discard Status to reduce memory consumption in watchable
		// It will be recomputed by the gateway-api layer
		policy.Status = gwapiv1a2.PolicyStatus{}
		resourceTree.ClientTrafficPolicies = append(resourceTree.ClientTrafficPolicies, &policy)
	}

	r.processCtpConfigMapRefs(ctx, resourceTree, resourceMap)

	return nil
}

// processBackendTrafficPolicies adds BackendTrafficPolicies to the resourceTree
func (r *gatewayAPIReconciler) processBackendTrafficPolicies(ctx context.Context, resourceTree *gatewayapi.Resources) error {
	backendTrafficPolicies := egv1a1.BackendTrafficPolicyList{}
	if err := r.client.List(ctx, &backendTrafficPolicies); err != nil {
		return fmt.Errorf("error listing BackendTrafficPolicies: %w", err)
	}

	for _, policy := range backendTrafficPolicies.Items {
		policy := policy
		// Discard Status to reduce memory consumption in watchable
		// It will be recomputed by the gateway-api layer
		policy.Status = gwapiv1a2.PolicyStatus{}
		resourceTree.BackendTrafficPolicies = append(resourceTree.BackendTrafficPolicies, &policy)
	}
	return nil
}

// processSecurityPolicies adds SecurityPolicies and their referenced resources to the resourceTree
func (r *gatewayAPIReconciler) processSecurityPolicies(
	ctx context.Context, resourceTree *gatewayapi.Resources, resourceMap *resourceMappings,
) error {
	securityPolicies := egv1a1.SecurityPolicyList{}
	if err := r.client.List(ctx, &securityPolicies); err != nil {
		return fmt.Errorf("error listing SecurityPolicies: %w", err)
	}

	for _, policy := range securityPolicies.Items {
		policy := policy
		// Discard Status to reduce memory consumption in watchable
		// It will be recomputed by the gateway-api layer
		policy.Status = gwapiv1a2.PolicyStatus{}
		resourceTree.SecurityPolicies = append(resourceTree.SecurityPolicies, &policy)
	}

	// Add the referenced Resources in SecurityPolicies to the resourceTree
	r.processSecurityPolicyObjectRefs(ctx, resourceTree, resourceMap)

	// Add the OIDC HMAC Secret to the resourceTree
	r.processOIDCHMACSecret(ctx, resourceTree)
	return nil
}

// processBackendTLSPolicies adds BackendTLSPolicies and their referenced resources to the resourceTree
func (r *gatewayAPIReconciler) processBackendTLSPolicies(
	ctx context.Context, resourceTree *gatewayapi.Resources, resourceMap *resourceMappings,
) error {
	backendTLSPolicies := gwapiv1a3.BackendTLSPolicyList{}
	if err := r.client.List(ctx, &backendTLSPolicies); err != nil {
		return fmt.Errorf("error listing BackendTLSPolicies: %w", err)
	}

	for _, policy := range backendTLSPolicies.Items {
		policy := policy
		// Discard Status to reduce memory consumption in watchable
		// It will be recomputed by the gateway-api layer
		policy.Status = gwapiv1a2.PolicyStatus{}
		resourceTree.BackendTLSPolicies = append(resourceTree.BackendTLSPolicies, &policy)
	}

	// Add the referenced Secrets and ConfigMaps in BackendTLSPolicies to the resourceTree.
	r.processBackendTLSPolicyConfigMapRefs(ctx, resourceTree, resourceMap)
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
	// Upon leader election, we retrigger the reconciliation process to allow the elected leader to
	// process status updates and infrastructure changes. This step is crucial for synchronizing resources
	// that may have been altered or introduced while there was no elected leader.
	if err := c.Watch(NewWatchAndReconcileSource(mgr.Elected(), &gwapiv1.GatewayClass{}, handler.EnqueueRequestsFromMapFunc(r.enqueueClass))); err != nil {
		return err
	}

	if err := c.Watch(
		source.Kind(mgr.GetCache(), &gwapiv1.GatewayClass{},
			handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, gc *gwapiv1.GatewayClass) []reconcile.Request {
				return r.enqueueClass(ctx, gc)
			}),
			&predicate.TypedGenerationChangedPredicate[*gwapiv1.GatewayClass]{},
			predicate.NewTypedPredicateFuncs[*gwapiv1.GatewayClass](r.hasMatchingController))); err != nil {
		return err
	}

	// Only enqueue EnvoyProxy objects that match this Envoy Gateway's GatewayClass.
	epPredicates := []predicate.TypedPredicate[*egv1a1.EnvoyProxy]{
		&predicate.TypedGenerationChangedPredicate[*egv1a1.EnvoyProxy]{},
		predicate.NewTypedPredicateFuncs[*egv1a1.EnvoyProxy](r.hasManagedClass),
	}
	if r.namespaceLabel != nil {
		epPredicates = append(epPredicates, predicate.NewTypedPredicateFuncs(func(ep *egv1a1.EnvoyProxy) bool {
			return r.hasMatchingNamespaceLabels(ep)
		}))
	}
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &egv1a1.EnvoyProxy{},
			handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, t *egv1a1.EnvoyProxy) []reconcile.Request {
				return r.enqueueClass(ctx, t)
			}),
			epPredicates...)); err != nil {
		return err
	}

	// Watch Gateway CRUDs and reconcile affected GatewayClass.
	gPredicates := []predicate.TypedPredicate[*gwapiv1.Gateway]{
		predicate.TypedGenerationChangedPredicate[*gwapiv1.Gateway]{},
		predicate.NewTypedPredicateFuncs(func(gtw *gwapiv1.Gateway) bool {
			return r.validateGatewayForReconcile(gtw)
		}),
	}
	if r.namespaceLabel != nil {
		gPredicates = append(gPredicates, predicate.NewTypedPredicateFuncs(func(gtw *gwapiv1.Gateway) bool {
			return r.hasMatchingNamespaceLabels(gtw)
		}))
	}
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &gwapiv1.Gateway{},
			handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, gtw *gwapiv1.Gateway) []reconcile.Request {
				return r.enqueueClass(ctx, gtw)
			}),
			gPredicates...)); err != nil {
		return err
	}
	if err := addGatewayIndexers(ctx, mgr); err != nil {
		return err
	}

	// Watch HTTPRoute CRUDs and process affected Gateways.
	httprPredicates := []predicate.TypedPredicate[*gwapiv1.HTTPRoute]{
		predicate.TypedGenerationChangedPredicate[*gwapiv1.HTTPRoute]{},
	}
	if r.namespaceLabel != nil {
		httprPredicates = append(httprPredicates, predicate.NewTypedPredicateFuncs(func(hr *gwapiv1.HTTPRoute) bool {
			return r.hasMatchingNamespaceLabels(hr)
		}))
	}
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &gwapiv1.HTTPRoute{},
			handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, t *gwapiv1.HTTPRoute) []reconcile.Request {
				return r.enqueueClass(ctx, t)
			}),
			httprPredicates...)); err != nil {
		return err
	}
	if err := addHTTPRouteIndexers(ctx, mgr); err != nil {
		return err
	}

	// Watch GRPCRoute CRUDs and process affected Gateways.
	grpcrPredicates := []predicate.TypedPredicate[*gwapiv1.GRPCRoute]{
		predicate.TypedGenerationChangedPredicate[*gwapiv1.GRPCRoute]{},
	}
	if r.namespaceLabel != nil {
		grpcrPredicates = append(grpcrPredicates, predicate.NewTypedPredicateFuncs[*gwapiv1.GRPCRoute](func(grpc *gwapiv1.GRPCRoute) bool {
			return r.hasMatchingNamespaceLabels(grpc)
		}))
	}
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &gwapiv1.GRPCRoute{},
			handler.TypedEnqueueRequestsFromMapFunc[*gwapiv1.GRPCRoute](func(ctx context.Context, route *gwapiv1.GRPCRoute) []reconcile.Request {
				return r.enqueueClass(ctx, route)
			}),
			grpcrPredicates...)); err != nil {
		return err
	}
	if err := addGRPCRouteIndexers(ctx, mgr); err != nil {
		return err
	}

	// Watch TLSRoute CRUDs and process affected Gateways.
	tlsrPredicates := []predicate.TypedPredicate[*gwapiv1a2.TLSRoute]{
		predicate.TypedGenerationChangedPredicate[*gwapiv1a2.TLSRoute]{},
	}
	if r.namespaceLabel != nil {
		tlsrPredicates = append(tlsrPredicates, predicate.NewTypedPredicateFuncs[*gwapiv1a2.TLSRoute](func(route *gwapiv1a2.TLSRoute) bool {
			return r.hasMatchingNamespaceLabels(route)
		}))
	}
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &gwapiv1a2.TLSRoute{},
			handler.TypedEnqueueRequestsFromMapFunc[*gwapiv1a2.TLSRoute](func(ctx context.Context, route *gwapiv1a2.TLSRoute) []reconcile.Request {
				return r.enqueueClass(ctx, route)
			}),
			tlsrPredicates...)); err != nil {
		return err
	}
	if err := addTLSRouteIndexers(ctx, mgr); err != nil {
		return err
	}

	// Watch UDPRoute CRUDs and process affected Gateways.
	udprPredicates := []predicate.TypedPredicate[*gwapiv1a2.UDPRoute]{
		predicate.TypedGenerationChangedPredicate[*gwapiv1a2.UDPRoute]{},
	}
	if r.namespaceLabel != nil {
		udprPredicates = append(udprPredicates, predicate.NewTypedPredicateFuncs[*gwapiv1a2.UDPRoute](func(route *gwapiv1a2.UDPRoute) bool {
			return r.hasMatchingNamespaceLabels(route)
		}))
	}
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &gwapiv1a2.UDPRoute{},
			handler.TypedEnqueueRequestsFromMapFunc[*gwapiv1a2.UDPRoute](func(ctx context.Context, route *gwapiv1a2.UDPRoute) []reconcile.Request {
				return r.enqueueClass(ctx, route)
			}),
			udprPredicates...)); err != nil {
		return err
	}
	if err := addUDPRouteIndexers(ctx, mgr); err != nil {
		return err
	}

	// Watch TCPRoute CRUDs and process affected Gateways.
	tcprPredicates := []predicate.TypedPredicate[*gwapiv1a2.TCPRoute]{
		predicate.TypedGenerationChangedPredicate[*gwapiv1a2.TCPRoute]{},
	}
	if r.namespaceLabel != nil {
		tcprPredicates = append(tcprPredicates, predicate.NewTypedPredicateFuncs[*gwapiv1a2.TCPRoute](func(route *gwapiv1a2.TCPRoute) bool {
			return r.hasMatchingNamespaceLabels(route)
		}))
	}
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &gwapiv1a2.TCPRoute{},
			handler.TypedEnqueueRequestsFromMapFunc[*gwapiv1a2.TCPRoute](func(ctx context.Context, route *gwapiv1a2.TCPRoute) []reconcile.Request {
				return r.enqueueClass(ctx, route)
			}),
			tcprPredicates...)); err != nil {
		return err
	}
	if err := addTCPRouteIndexers(ctx, mgr); err != nil {
		return err
	}

	// Watch Service CRUDs and process affected *Route objects.
	servicePredicates := []predicate.TypedPredicate[*corev1.Service]{
		predicate.NewTypedPredicateFuncs[*corev1.Service](func(svc *corev1.Service) bool {
			return r.validateServiceForReconcile(svc)
		}),
	}
	if r.namespaceLabel != nil {
		servicePredicates = append(servicePredicates, predicate.NewTypedPredicateFuncs[*corev1.Service](func(svc *corev1.Service) bool {
			return r.hasMatchingNamespaceLabels(svc)
		}))
	}
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &corev1.Service{},
			handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, svc *corev1.Service) []reconcile.Request {
				return r.enqueueClass(ctx, svc)
			}),
			servicePredicates...)); err != nil {
		return err
	}

	serviceImportCRDExists := r.serviceImportCRDExists(mgr)
	if !serviceImportCRDExists {
		r.log.Info("ServiceImport CRD not found, skipping ServiceImport watch")
	}

	// Watch ServiceImport CRUDs and process affected *Route objects.
	if serviceImportCRDExists {
		if err := c.Watch(
			source.Kind(mgr.GetCache(), &mcsapi.ServiceImport{},
				handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, si *mcsapi.ServiceImport) []reconcile.Request {
					return r.enqueueClass(ctx, si)
				}),
				predicate.TypedGenerationChangedPredicate[*mcsapi.ServiceImport]{},
				predicate.NewTypedPredicateFuncs[*mcsapi.ServiceImport](func(si *mcsapi.ServiceImport) bool {
					return r.validateServiceImportForReconcile(si)
				}))); err != nil {
			// ServiceImport is not available in the cluster, skip the watch and not throw error.
			r.log.Info("unable to watch ServiceImport: %s", err.Error())
		}
	}

	// Watch EndpointSlice CRUDs and process affected *Route objects.
	esPredicates := []predicate.TypedPredicate[*discoveryv1.EndpointSlice]{
		predicate.TypedGenerationChangedPredicate[*discoveryv1.EndpointSlice]{},
		predicate.NewTypedPredicateFuncs[*discoveryv1.EndpointSlice](func(eps *discoveryv1.EndpointSlice) bool {
			return r.validateEndpointSliceForReconcile(eps)
		}),
	}
	if r.namespaceLabel != nil {
		esPredicates = append(esPredicates, predicate.NewTypedPredicateFuncs[*discoveryv1.EndpointSlice](func(eps *discoveryv1.EndpointSlice) bool {
			return r.hasMatchingNamespaceLabels(eps)
		}))
	}
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &discoveryv1.EndpointSlice{},
			handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, si *discoveryv1.EndpointSlice) []reconcile.Request {
				return r.enqueueClass(ctx, si)
			}),
			esPredicates...)); err != nil {
		return err
	}

	// Watch Node CRUDs to update Gateway Address exposed by Service of type NodePort.
	// Node creation/deletion and ExternalIP updates would require update in the Gateway
	nPredicates := []predicate.TypedPredicate[*corev1.Node]{
		predicate.TypedGenerationChangedPredicate[*corev1.Node]{},
		predicate.NewTypedPredicateFuncs[*corev1.Node](func(node *corev1.Node) bool {
			return r.handleNode(node)
		}),
	}
	if r.namespaceLabel != nil {
		nPredicates = append(nPredicates, predicate.NewTypedPredicateFuncs[*corev1.Node](func(node *corev1.Node) bool {
			return r.hasMatchingNamespaceLabels(node)
		}))
	}
	// resource address.
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &corev1.Node{},
			handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, si *corev1.Node) []reconcile.Request {
				return r.enqueueClass(ctx, si)
			}),
			nPredicates...)); err != nil {
		return err
	}

	// Watch Secret CRUDs and process affected EG CRs (Gateway, SecurityPolicy, more in the future).
	secretPredicates := []predicate.TypedPredicate[*corev1.Secret]{
		predicate.TypedGenerationChangedPredicate[*corev1.Secret]{},
		predicate.NewTypedPredicateFuncs(func(s *corev1.Secret) bool {
			return r.validateSecretForReconcile(s)
		}),
	}
	if r.namespaceLabel != nil {
		secretPredicates = append(secretPredicates, predicate.NewTypedPredicateFuncs(func(s *corev1.Secret) bool {
			return r.hasMatchingNamespaceLabels(s)
		}))
	}
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &corev1.Secret{},
			handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, s *corev1.Secret) []reconcile.Request {
				return r.enqueueClass(ctx, s)
			}),
			secretPredicates...)); err != nil {
		return err
	}

	// Watch ConfigMap CRUDs and process affected ClienTraffiPolicies and BackendTLSPolicies.
	configMapPredicates := []predicate.TypedPredicate[*corev1.ConfigMap]{
		predicate.TypedGenerationChangedPredicate[*corev1.ConfigMap]{},
		predicate.NewTypedPredicateFuncs[*corev1.ConfigMap](func(cm *corev1.ConfigMap) bool {
			return r.validateConfigMapForReconcile(cm)
		}),
	}
	if r.namespaceLabel != nil {
		configMapPredicates = append(configMapPredicates, predicate.NewTypedPredicateFuncs[*corev1.ConfigMap](func(cm *corev1.ConfigMap) bool {
			return r.hasMatchingNamespaceLabels(cm)
		}))
	}
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &corev1.ConfigMap{},
			handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, cm *corev1.ConfigMap) []reconcile.Request {
				return r.enqueueClass(ctx, cm)
			}),
			configMapPredicates...)); err != nil {
		return err
	}

	// Watch ReferenceGrant CRUDs and process affected Gateways.
	rgPredicates := []predicate.TypedPredicate[*gwapiv1b1.ReferenceGrant]{
		predicate.TypedGenerationChangedPredicate[*gwapiv1b1.ReferenceGrant]{},
	}
	if r.namespaceLabel != nil {
		rgPredicates = append(rgPredicates, predicate.NewTypedPredicateFuncs[*gwapiv1b1.ReferenceGrant](func(rg *gwapiv1b1.ReferenceGrant) bool {
			return r.hasMatchingNamespaceLabels(rg)
		}))
	}
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &gwapiv1b1.ReferenceGrant{},
			handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, rg *gwapiv1b1.ReferenceGrant) []reconcile.Request {
				return r.enqueueClass(ctx, rg)
			}),
			rgPredicates...)); err != nil {
		return err
	}
	if err := addReferenceGrantIndexers(ctx, mgr); err != nil {
		return err
	}

	// Watch Deployment CRUDs and process affected Gateways.
	dPredicates := []predicate.TypedPredicate[*appsv1.Deployment]{
		predicate.NewTypedPredicateFuncs[*appsv1.Deployment](func(deploy *appsv1.Deployment) bool {
			return r.validateDeploymentForReconcile(deploy)
		}),
	}
	if r.namespaceLabel != nil {
		dPredicates = append(dPredicates, predicate.NewTypedPredicateFuncs[*appsv1.Deployment](func(deploy *appsv1.Deployment) bool {
			return r.hasMatchingNamespaceLabels(deploy)
		}))
	}
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &appsv1.Deployment{},
			handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, deploy *appsv1.Deployment) []reconcile.Request {
				return r.enqueueClass(ctx, deploy)
			}),
			dPredicates...)); err != nil {
		return err
	}

	if r.envoyGateway.ExtensionAPIs != nil && r.envoyGateway.ExtensionAPIs.EnableEnvoyPatchPolicy {
		// Watch EnvoyPatchPolicy if enabled in config
		eppPredicates := []predicate.TypedPredicate[*egv1a1.EnvoyPatchPolicy]{
			predicate.TypedGenerationChangedPredicate[*egv1a1.EnvoyPatchPolicy]{},
		}
		if r.namespaceLabel != nil {
			eppPredicates = append(eppPredicates, predicate.NewTypedPredicateFuncs[*egv1a1.EnvoyPatchPolicy](func(epp *egv1a1.EnvoyPatchPolicy) bool {
				return r.hasMatchingNamespaceLabels(epp)
			}))
		}
		// Watch EnvoyPatchPolicy CRUDs
		if err := c.Watch(
			source.Kind(mgr.GetCache(), &egv1a1.EnvoyPatchPolicy{},
				handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, epp *egv1a1.EnvoyPatchPolicy) []reconcile.Request {
					return r.enqueueClass(ctx, epp)
				}),
				eppPredicates...)); err != nil {
			return err
		}
	}

	// Watch ClientTrafficPolicy
	ctpPredicates := []predicate.TypedPredicate[*egv1a1.ClientTrafficPolicy]{
		predicate.TypedGenerationChangedPredicate[*egv1a1.ClientTrafficPolicy]{},
	}
	if r.namespaceLabel != nil {
		ctpPredicates = append(ctpPredicates, predicate.NewTypedPredicateFuncs[*egv1a1.ClientTrafficPolicy](func(ctp *egv1a1.ClientTrafficPolicy) bool {
			return r.hasMatchingNamespaceLabels(ctp)
		}))
	}

	if err := c.Watch(
		source.Kind(mgr.GetCache(), &egv1a1.ClientTrafficPolicy{},
			handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, ctp *egv1a1.ClientTrafficPolicy) []reconcile.Request {
				return r.enqueueClass(ctx, ctp)
			}),
			ctpPredicates...)); err != nil {
		return err
	}

	if err := addCtpIndexers(ctx, mgr); err != nil {
		return err
	}

	// Watch BackendTrafficPolicy
	btpPredicates := []predicate.TypedPredicate[*egv1a1.BackendTrafficPolicy]{
		predicate.TypedGenerationChangedPredicate[*egv1a1.BackendTrafficPolicy]{},
	}
	if r.namespaceLabel != nil {
		btpPredicates = append(btpPredicates, predicate.NewTypedPredicateFuncs[*egv1a1.BackendTrafficPolicy](func(btp *egv1a1.BackendTrafficPolicy) bool {
			return r.hasMatchingNamespaceLabels(btp)
		}))
	}

	if err := c.Watch(
		source.Kind(mgr.GetCache(), &egv1a1.BackendTrafficPolicy{},
			handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, btp *egv1a1.BackendTrafficPolicy) []reconcile.Request {
				return r.enqueueClass(ctx, btp)
			}),
			btpPredicates...)); err != nil {
		return err
	}

	// Watch SecurityPolicy
	spPredicates := []predicate.TypedPredicate[*egv1a1.SecurityPolicy]{
		predicate.TypedGenerationChangedPredicate[*egv1a1.SecurityPolicy]{},
	}
	if r.namespaceLabel != nil {
		spPredicates = append(spPredicates, predicate.NewTypedPredicateFuncs[*egv1a1.SecurityPolicy](func(sp *egv1a1.SecurityPolicy) bool {
			return r.hasMatchingNamespaceLabels(sp)
		}))
	}

	if err := c.Watch(
		source.Kind(mgr.GetCache(), &egv1a1.SecurityPolicy{},
			handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, sp *egv1a1.SecurityPolicy) []reconcile.Request {
				return r.enqueueClass(ctx, sp)
			}),
			spPredicates...)); err != nil {
		return err
	}
	if err := addSecurityPolicyIndexers(ctx, mgr); err != nil {
		return err
	}

	// Watch BackendTLSPolicy
	btlsPredicates := []predicate.TypedPredicate[*gwapiv1a3.BackendTLSPolicy]{
		predicate.TypedGenerationChangedPredicate[*gwapiv1a3.BackendTLSPolicy]{},
	}
	if r.namespaceLabel != nil {
		btlsPredicates = append(btlsPredicates, predicate.NewTypedPredicateFuncs[*gwapiv1a3.BackendTLSPolicy](func(btp *gwapiv1a3.BackendTLSPolicy) bool {
			return r.hasMatchingNamespaceLabels(btp)
		}))
	}

	if err := c.Watch(
		source.Kind(mgr.GetCache(), &gwapiv1a3.BackendTLSPolicy{},
			handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, btp *gwapiv1a3.BackendTLSPolicy) []reconcile.Request {
				return r.enqueueClass(ctx, btp)
			}),
			btlsPredicates...)); err != nil {
		return err
	}

	if err := addBtlsIndexers(ctx, mgr); err != nil {
		return err
	}

	// Watch EnvoyExtensionPolicy
	eepPredicates := []predicate.TypedPredicate[*egv1a1.EnvoyExtensionPolicy]{
		predicate.TypedGenerationChangedPredicate[*egv1a1.EnvoyExtensionPolicy]{},
	}
	if r.namespaceLabel != nil {
		eepPredicates = append(eepPredicates, predicate.NewTypedPredicateFuncs[*egv1a1.EnvoyExtensionPolicy](func(eep *egv1a1.EnvoyExtensionPolicy) bool {
			return r.hasMatchingNamespaceLabels(eep)
		}))
	}

	// Watch EnvoyExtensionPolicy CRUDs
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &egv1a1.EnvoyExtensionPolicy{},
			handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, eep *egv1a1.EnvoyExtensionPolicy) []reconcile.Request {
				return r.enqueueClass(ctx, eep)
			}),
			eepPredicates...)); err != nil {
		return err
	}
	if err := addEnvoyExtensionPolicyIndexers(ctx, mgr); err != nil {
		return err
	}

	r.log.Info("Watching gatewayAPI related objects")

	// Watch any additional GVKs from the registered extension.
	uPredicates := []predicate.TypedPredicate[*unstructured.Unstructured]{
		predicate.TypedGenerationChangedPredicate[*unstructured.Unstructured]{},
	}
	if r.namespaceLabel != nil {
		uPredicates = append(uPredicates, predicate.NewTypedPredicateFuncs[*unstructured.Unstructured](func(obj *unstructured.Unstructured) bool {
			return r.hasMatchingNamespaceLabels(obj)
		}))
	}
	for _, gvk := range r.extGVKs {
		u := &unstructured.Unstructured{}
		u.SetGroupVersionKind(gvk)
		if err := c.Watch(source.Kind(mgr.GetCache(), u,
			handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, si *unstructured.Unstructured) []reconcile.Request {
				return r.enqueueClass(ctx, si)
			}),
			uPredicates...)); err != nil {
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

func (r *gatewayAPIReconciler) hasManagedClass(ep *egv1a1.EnvoyProxy) bool {
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

func (r *gatewayAPIReconciler) processBackendTLSPolicyConfigMapRefs(ctx context.Context, resourceTree *gatewayapi.Resources, resourceMap *resourceMappings) {
	for _, policy := range resourceTree.BackendTLSPolicies {
		tls := policy.Spec.Validation

		if tls.CACertificateRefs != nil {
			for _, caCertRef := range tls.CACertificateRefs {
				if string(caCertRef.Kind) == gatewayapi.KindConfigMap {
					caRefNew := gwapiv1b1.SecretObjectReference{
						Group:     gatewayapi.GroupPtr(string(caCertRef.Group)),
						Kind:      gatewayapi.KindPtr(string(caCertRef.Kind)),
						Name:      caCertRef.Name,
						Namespace: gatewayapi.NamespacePtr(policy.Namespace),
					}
					if err := r.processConfigMapRef(
						ctx,
						resourceMap,
						resourceTree,
						gatewayapi.KindBackendTLSPolicy,
						policy.Namespace,
						policy.Name,
						caRefNew); err != nil {
						// we don't return an error here, because we want to continue
						// reconciling the rest of the ClientTrafficPolicies despite that this
						// reference is invalid.
						// This ClientTrafficPolicy will be marked as invalid in its status
						// when translating to IR because the referenced configmap can't be
						// found.
						r.log.Error(err,
							"failed to process CACertificateRef for BackendTLSPolicy",
							"policy", policy, "caCertificateRef", caCertRef.Name)
					}
				}
			}
		}
	}
}

// processEnvoyExtensionPolicies adds EnvoyExtensionPolicies and their referenced resources to the resourceTree
func (r *gatewayAPIReconciler) processEnvoyExtensionPolicies(
	ctx context.Context, resourceTree *gatewayapi.Resources, resourceMap *resourceMappings,
) error {
	envoyExtensionPolicies := egv1a1.EnvoyExtensionPolicyList{}
	if err := r.client.List(ctx, &envoyExtensionPolicies); err != nil {
		return fmt.Errorf("error listing EnvoyExtensionPolicies: %w", err)
	}

	for _, policy := range envoyExtensionPolicies.Items {
		policy := policy
		// Discard Status to reduce memory consumption in watchable
		// It will be recomputed by the gateway-api layer
		policy.Status = gwapiv1a2.PolicyStatus{}
		resourceTree.EnvoyExtensionPolicies = append(resourceTree.EnvoyExtensionPolicies, &policy)
	}

	// Add the referenced Resources in EnvoyExtensionPolicies to the resourceTree
	r.processEnvoyExtensionPolicyObjectRefs(ctx, resourceTree, resourceMap)

	return nil
}

// processEnvoyExtensionPolicyObjectRefs adds the referenced resources in EnvoyExtensionPolicies
// to the resourceTree
// - BackendRefs for ExtProcs
func (r *gatewayAPIReconciler) processEnvoyExtensionPolicyObjectRefs(
	ctx context.Context, resourceTree *gatewayapi.Resources, resourceMap *resourceMappings,
) {
	// we don't return errors from this method, because we want to continue reconciling
	// the rest of the EnvoyExtensionPolicies despite that one reference is invalid. This
	// allows Envoy Gateway to continue serving traffic even if some EnvoyExtensionPolicies
	// are invalid.
	//
	// This EnvoyExtensionPolicy will be marked as invalid in its status when translating
	// to IR because the referenced service can't be found.
	for _, policy := range resourceTree.EnvoyExtensionPolicies {
		// Add the referenced BackendRefs and ReferenceGrants in ExtAuth to Maps for later processing
		for _, ep := range policy.Spec.ExtProc {
			for _, br := range ep.BackendRefs {
				backendRef := br.BackendObjectReference

				backendNamespace := gatewayapi.NamespaceDerefOr(backendRef.Namespace, policy.Namespace)
				resourceMap.allAssociatedBackendRefs[gwapiv1.BackendObjectReference{
					Group:     backendRef.Group,
					Kind:      backendRef.Kind,
					Namespace: gatewayapi.NamespacePtrV1Alpha2(backendNamespace),
					Name:      backendRef.Name,
				}] = struct{}{}

				if backendNamespace != policy.Namespace {
					from := ObjectKindNamespacedName{
						kind:      gatewayapi.KindHTTPRoute,
						namespace: policy.Namespace,
						name:      policy.Name,
					}
					to := ObjectKindNamespacedName{
						kind:      gatewayapi.KindDerefOr(backendRef.Kind, gatewayapi.KindService),
						namespace: backendNamespace,
						name:      string(backendRef.Name),
					}
					refGrant, err := r.findReferenceGrant(ctx, from, to)
					switch {
					case err != nil:
						r.log.Error(err, "failed to find ReferenceGrant")
					case refGrant == nil:
						r.log.Info("no matching ReferenceGrants found", "from", from.kind,
							"from namespace", from.namespace, "target", to.kind, "target namespace", to.namespace)
					default:
						resourceTree.ReferenceGrants = append(resourceTree.ReferenceGrants, refGrant)
						r.log.Info("added ReferenceGrant to resource map", "namespace", refGrant.Namespace,
							"name", refGrant.Name)
					}
				}
			}
		}
	}
}
