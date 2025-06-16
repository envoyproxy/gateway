// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"fmt"
	"time"

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
	"k8s.io/client-go/util/workqueue"
	"k8s.io/utils/ptr"
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
	mcsapiv1a1 "sigs.k8s.io/mcs-api/pkg/apis/v1alpha1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/api/v1alpha1/validation"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/message"
	workqueuemetrics "github.com/envoyproxy/gateway/internal/metrics/workqueue"
	"github.com/envoyproxy/gateway/internal/utils"
	"github.com/envoyproxy/gateway/internal/utils/slice"
	"github.com/envoyproxy/gateway/internal/xds/bootstrap"
)

var skipNameValidation = func() *bool {
	return ptr.To(false)
}

type gatewayAPIReconciler struct {
	client               client.Client
	log                  logging.Logger
	statusUpdater        Updater
	classController      gwapiv1.GatewayController
	store                *kubernetesProviderStore
	namespace            string
	namespaceLabel       *metav1.LabelSelector
	envoyGateway         *egv1a1.EnvoyGateway
	mergeGateways        sets.Set[string]
	resources            *message.ProviderResources
	extGVKs              []schema.GroupVersionKind
	extServerPolicies    []schema.GroupVersionKind
	gatewayNamespaceMode bool

	backendCRDExists       bool
	bTLSPolicyCRDExists    bool
	btpCRDExists           bool
	ctpCRDExists           bool
	eepCRDExists           bool
	epCRDExists            bool
	eppCRDExists           bool
	hrfCRDExists           bool
	grpcRouteCRDExists     bool
	serviceImportCRDExists bool
	spCRDExists            bool
	tcpRouteCRDExists      bool
	tlsRouteCRDExists      bool
	udpRouteCRDExists      bool
}

// newGatewayAPIController
func newGatewayAPIController(ctx context.Context, mgr manager.Manager, cfg *config.Server, su Updater,
	resources *message.ProviderResources,
) error {
	// Gather additional resources to watch from registered extensions
	var extServerPoliciesGVKs []schema.GroupVersionKind
	var extGVKs []schema.GroupVersionKind
	if cfg.EnvoyGateway.ExtensionManager != nil {
		for _, rsrc := range cfg.EnvoyGateway.ExtensionManager.Resources {
			gvk := schema.GroupVersionKind(rsrc)
			extGVKs = append(extGVKs, gvk)
		}
		for _, rsrc := range cfg.EnvoyGateway.ExtensionManager.PolicyResources {
			gvk := schema.GroupVersionKind(rsrc)
			extServerPoliciesGVKs = append(extServerPoliciesGVKs, gvk)
		}
	}

	r := &gatewayAPIReconciler{
		client:               mgr.GetClient(),
		log:                  cfg.Logger,
		classController:      gwapiv1.GatewayController(cfg.EnvoyGateway.Gateway.ControllerName),
		namespace:            cfg.ControllerNamespace,
		statusUpdater:        su,
		resources:            resources,
		extGVKs:              extGVKs,
		store:                newProviderStore(),
		envoyGateway:         cfg.EnvoyGateway,
		mergeGateways:        sets.New[string](),
		extServerPolicies:    extServerPoliciesGVKs,
		gatewayNamespaceMode: cfg.EnvoyGateway.GatewayNamespaceMode(),
	}

	if byNamespaceSelectorEnabled(cfg.EnvoyGateway) {
		r.namespaceLabel = cfg.EnvoyGateway.Provider.Kubernetes.Watch.NamespaceSelector
	}

	// controller-runtime doesn't allow run controller with same name for more than once
	// see https://github.com/kubernetes-sigs/controller-runtime/blob/2b941650bce159006c88bd3ca0d132c7bc40e947/pkg/controller/name.go#L29
	name := fmt.Sprintf("gatewayapi-%d", time.Now().Unix())
	c, err := controller.New(name, mgr, controller.Options{
		Reconciler:         r,
		SkipNameValidation: skipNameValidation(),
		NewQueue: func(controllerName string, rateLimiter workqueue.TypedRateLimiter[reconcile.Request]) workqueue.TypedRateLimitingInterface[reconcile.Request] {
			return workqueue.NewTypedRateLimitingQueueWithConfig(rateLimiter, workqueue.TypedRateLimitingQueueConfig[reconcile.Request]{
				Name:            controllerName,
				MetricsProvider: workqueuemetrics.WorkqueueMetricsProvider{},
			})
		},
	})
	if err != nil {
		return fmt.Errorf("error creating controller: %w", err)
	}
	r.log.Info("created gatewayapi controller")

	// Watch resources
	if err := r.watchResources(ctx, mgr, c); err != nil {
		return fmt.Errorf("error watching resources: %w", err)
	}

	// When leader election is enabled, only subscribe to status updates upon acquiring leadership.
	if cfg.EnvoyGateway.Provider.Type == egv1a1.ProviderTypeKubernetes &&
		!ptr.Deref(cfg.EnvoyGateway.Provider.Kubernetes.LeaderElection.Disable, false) {
		go func() {
			select {
			case <-ctx.Done():
				return
			case <-cfg.Elected:
				r.subscribeAndUpdateStatus(ctx, cfg.EnvoyGateway.ExtensionManager != nil)
			}
		}()
	} else {
		r.subscribeAndUpdateStatus(ctx, cfg.EnvoyGateway.ExtensionManager != nil)
	}
	return nil
}

func byNamespaceSelectorEnabled(eg *egv1a1.EnvoyGateway) bool {
	if eg.Provider == nil ||
		eg.Provider.Kubernetes == nil ||
		eg.Provider.Kubernetes.Watch == nil {
		return false
	}

	watch := eg.Provider.Kubernetes.Watch
	switch watch.Type {
	case egv1a1.KubernetesWatchModeTypeNamespaceSelector:
		// Make sure that the namespace selector has at least one label or expression is set.
		return watch.NamespaceSelector.MatchLabels != nil || len(watch.NamespaceSelector.MatchExpressions) > 0
	default:
		return false
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
	gwcResources := make(resource.ControllerResources, 0, len(managedGCs))
	for _, managedGC := range managedGCs {
		// Initialize resource types.
		gwcResource := resource.NewResources()
		gwcResource.GatewayClass = managedGC
		gwcResources = append(gwcResources, gwcResource)
		resourceMappings := newResourceMapping()

		// Process the parametersRef of the accepted GatewayClass.
		// This should run before processGateways and processBackendRefs
		if managedGC.Spec.ParametersRef != nil && managedGC.DeletionTimestamp == nil {
			if err := r.processGatewayClassParamsRef(ctx, managedGC, resourceMappings, gwcResource); err != nil {
				msg := fmt.Sprintf("%s: %v", status.MsgGatewayClassInvalidParams, err)
				gc := status.SetGatewayClassAccepted(
					managedGC.DeepCopy(),
					false,
					string(gwapiv1.GatewayClassReasonInvalidParameters),
					msg)
				r.resources.GatewayClassStatuses.Store(utils.NamespacedName(gc), &gc.Status)
				continue
			}
		}

		// Add all Gateways, their associated Routes, and referenced resources to the resourceTree
		if err = r.processGateways(ctx, managedGC, resourceMappings, gwcResource); err != nil {
			r.log.Error(err, fmt.Sprintf("failed processGateways for gatewayClass %s, skipping it", managedGC.Name))
			continue
		}

		if r.eppCRDExists {
			// Add all EnvoyPatchPolicies to the resourceTree
			if err = r.processEnvoyPatchPolicies(ctx, gwcResource, resourceMappings); err != nil {
				r.log.Error(err, fmt.Sprintf("failed processEnvoyPatchPolicies for gatewayClass %s, skipping it", managedGC.Name))
				continue
			}
		}
		if r.ctpCRDExists {
			// Add all ClientTrafficPolicies and their referenced resources to the resourceTree
			if err = r.processClientTrafficPolicies(ctx, gwcResource, resourceMappings); err != nil {
				r.log.Error(err, fmt.Sprintf("failed processClientTrafficPolicies for gatewayClass %s, skipping it", managedGC.Name))
				continue
			}
		}

		if r.btpCRDExists {
			// Add all BackendTrafficPolicies to the resourceTree
			if err = r.processBackendTrafficPolicies(ctx, gwcResource, resourceMappings); err != nil {
				r.log.Error(err, fmt.Sprintf("failed processBackendTrafficPolicies for gatewayClass %s, skipping it", managedGC.Name))
				continue
			}
		}

		if r.spCRDExists {
			// Add all SecurityPolicies and their referenced resources to the resourceTree
			if err = r.processSecurityPolicies(ctx, gwcResource, resourceMappings); err != nil {
				r.log.Error(err, fmt.Sprintf("failed processSecurityPolicies for gatewayClass %s, skipping it", managedGC.Name))
				continue
			}
		}

		if r.bTLSPolicyCRDExists {
			// Add all BackendTLSPolies to the resourceTree
			if err = r.processBackendTLSPolicies(ctx, gwcResource, resourceMappings); err != nil {
				r.log.Error(err, fmt.Sprintf("failed processBackendTLSPolicies for gatewayClass %s, skipping it", managedGC.Name))
				continue
			}
		}

		if r.eepCRDExists {
			// Add all EnvoyExtensionPolicies and their referenced resources to the resourceTree
			if err = r.processEnvoyExtensionPolicies(ctx, gwcResource, resourceMappings); err != nil {
				r.log.Error(err, fmt.Sprintf("failed processEnvoyExtensionPolicies for gatewayClass %s, skipping it", managedGC.Name))
				continue
			}
		}

		if err = r.processExtensionServerPolicies(ctx, gwcResource); err != nil {
			r.log.Error(err, fmt.Sprintf("failed processExtensionServerPolicies for gatewayClass %s, skipping it", managedGC.Name))
			continue
		}

		if r.backendCRDExists {
			if err = r.processBackends(ctx, gwcResource); err != nil {
				r.log.Error(err, fmt.Sprintf("failed processBackends for gatewayClass %s, skipping it", managedGC.Name))
				continue
			}
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
					continue
				}
				continue
			}

			gwcResource.Namespaces = append(gwcResource.Namespaces, namespace)
		}

		if gwcResource.EnvoyProxyForGatewayClass != nil && gwcResource.EnvoyProxyForGatewayClass.Spec.MergeGateways != nil {
			if *gwcResource.EnvoyProxyForGatewayClass.Spec.MergeGateways {
				r.mergeGateways.Insert(managedGC.Name)
			} else {
				r.mergeGateways.Delete(managedGC.Name)
			}
		}

		// process envoy gateway secret refs
		r.processEnvoyProxySecretRef(ctx, gwcResource)
		gc := status.SetGatewayClassAccepted(
			managedGC.DeepCopy(),
			true,
			string(gwapiv1.GatewayClassReasonAccepted),
			status.MsgValidGatewayClass)
		r.resources.GatewayClassStatuses.Store(utils.NamespacedName(gc), &gc.Status)

		if len(gwcResource.Gateways) == 0 {
			r.log.Info("No gateways found for accepted gatewayClass")

			// If needed, remove the finalizer from the accepted GatewayClass.
			if err := r.removeFinalizer(ctx, managedGC); err != nil {
				r.log.Error(err, fmt.Sprintf("failed to remove finalizer from gatewayClass %s",
					managedGC.Name))
				continue
			}
		} else {
			// finalize the accepted GatewayClass.
			if err := r.addFinalizer(ctx, managedGC); err != nil {
				r.log.Error(err, fmt.Sprintf("failed adding finalizer to gatewayClass %s",
					managedGC.Name))
				continue
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

func (r *gatewayAPIReconciler) processEnvoyProxySecretRef(ctx context.Context, gwcResource *resource.Resources) {
	if gwcResource.EnvoyProxyForGatewayClass == nil || gwcResource.EnvoyProxyForGatewayClass.Spec.BackendTLS == nil || gwcResource.EnvoyProxyForGatewayClass.Spec.BackendTLS.ClientCertificateRef == nil {
		return
	}
	certRef := gwcResource.EnvoyProxyForGatewayClass.Spec.BackendTLS.ClientCertificateRef
	if refsSecret(certRef) {
		if err := r.processSecretRef(
			ctx,
			newResourceMapping(),
			gwcResource,
			resource.KindGateway,
			gwcResource.EnvoyProxyForGatewayClass.Namespace,
			resource.KindEnvoyProxy,
			*certRef); err != nil {
			r.log.Error(err,
				"failed to process TLS SecretRef for EnvoyProxy",
				"gateway", "issue", "secretRef", certRef)
		}
	}
}

// managedGatewayClasses returns a list of GatewayClass objects that are managed by the Envoy Gateway Controller.
func (r *gatewayAPIReconciler) managedGatewayClasses(ctx context.Context) ([]*gwapiv1.GatewayClass, error) {
	var gatewayClasses gwapiv1.GatewayClassList
	if err := r.client.List(ctx, &gatewayClasses); err != nil {
		return nil, fmt.Errorf("error listing gatewayclasses: %w", err)
	}

	var cc controlledClasses

	for _, gwClass := range gatewayClasses.Items {
		if gwClass.Spec.ControllerName == r.classController {
			// The gatewayclass was marked for deletion and the finalizer removed,
			// so clean-up dependents.
			if !gwClass.DeletionTimestamp.IsZero() &&
				!slice.ContainsString(gwClass.Finalizers, gatewayClassFinalizer) {
				r.log.Info("gatewayclass marked for deletion", "name", gwClass.Name)
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
// - Backends
// - CACertificateRefs in the Backends
func (r *gatewayAPIReconciler) processBackendRefs(ctx context.Context, gwcResource *resource.Resources, resourceMappings *resourceMappings) {
	for backendRef := range resourceMappings.allAssociatedBackendRefs {
		backendRefKind := gatewayapi.KindDerefOr(backendRef.Kind, resource.KindService)
		r.log.Info("processing Backend", "kind", backendRefKind, "namespace", string(*backendRef.Namespace),
			"name", string(backendRef.Name))

		var endpointSliceLabelKey string
		switch backendRefKind {
		case resource.KindService:
			service := new(corev1.Service)
			err := r.client.Get(ctx, types.NamespacedName{Namespace: string(*backendRef.Namespace), Name: string(backendRef.Name)}, service)
			if err != nil {
				r.log.Error(err, "failed to get Service", "namespace", string(*backendRef.Namespace),
					"name", string(backendRef.Name))
			} else {
				resourceMappings.allAssociatedNamespaces.Insert(service.Namespace)
				gwcResource.Services = append(gwcResource.Services, service)
				r.log.Info("added Service to resource tree", "namespace", string(*backendRef.Namespace),
					"name", string(backendRef.Name))
			}
			endpointSliceLabelKey = discoveryv1.LabelServiceName

		case resource.KindServiceImport:
			serviceImport := new(mcsapiv1a1.ServiceImport)
			err := r.client.Get(ctx, types.NamespacedName{Namespace: string(*backendRef.Namespace), Name: string(backendRef.Name)}, serviceImport)
			if err != nil {
				r.log.Error(err, "failed to get ServiceImport", "namespace", string(*backendRef.Namespace),
					"name", string(backendRef.Name))
			} else {
				resourceMappings.allAssociatedNamespaces.Insert(serviceImport.Namespace)
				key := utils.NamespacedName(serviceImport).String()
				if !resourceMappings.allAssociatedServiceImports.Has(key) {
					resourceMappings.allAssociatedServiceImports.Insert(key)
					gwcResource.ServiceImports = append(gwcResource.ServiceImports, serviceImport)
					r.log.Info("added ServiceImport to resource tree", "namespace", string(*backendRef.Namespace),
						"name", string(backendRef.Name))
				}
			}
			endpointSliceLabelKey = mcsapiv1a1.LabelServiceName

		case egv1a1.KindBackend:
			backend := new(egv1a1.Backend)
			err := r.client.Get(ctx, types.NamespacedName{Namespace: string(*backendRef.Namespace), Name: string(backendRef.Name)}, backend)
			if err != nil {
				r.log.Error(err, "failed to get Backend", "namespace", string(*backendRef.Namespace),
					"name", string(backendRef.Name))
			} else {
				resourceMappings.allAssociatedNamespaces.Insert(backend.Namespace)
				key := utils.NamespacedName(backend).String()
				if !resourceMappings.allAssociatedBackends.Has(key) {
					resourceMappings.allAssociatedBackends.Insert(key)
					gwcResource.Backends = append(gwcResource.Backends, backend)
					r.log.Info("added Backend to resource tree", "namespace", string(*backendRef.Namespace),
						"name", string(backendRef.Name))
				}
			}

			if backend.Spec.TLS != nil && backend.Spec.TLS.CACertificateRefs != nil {
				for _, caCertRef := range backend.Spec.TLS.CACertificateRefs {
					// if kind is not Secret or ConfigMap, we skip early to avoid further calculation overhead
					if string(caCertRef.Kind) == resource.KindConfigMap ||
						string(caCertRef.Kind) == resource.KindSecret {

						var err error
						caRefNew := gwapiv1.SecretObjectReference{
							Group:     gatewayapi.GroupPtr(string(caCertRef.Group)),
							Kind:      gatewayapi.KindPtr(string(caCertRef.Kind)),
							Name:      caCertRef.Name,
							Namespace: gatewayapi.NamespacePtr(backend.Namespace),
						}
						switch string(caCertRef.Kind) {
						case resource.KindConfigMap:
							err = r.processConfigMapRef(
								ctx,
								resourceMappings,
								gwcResource,
								resource.KindBackendTLSPolicy,
								backend.Namespace,
								backend.Name,
								caRefNew)

						case resource.KindSecret:
							err = r.processSecretRef(
								ctx,
								resourceMappings,
								gwcResource,
								resource.KindBackendTLSPolicy,
								backend.Namespace,
								backend.Name,
								caRefNew)
						}
						if err != nil {
							r.log.Error(err,
								"failed to process CACertificateRef for Backend",
								"backend", backend, "caCertificateRef", caCertRef.Name)
						}
					}
				}
			}
		}

		// Retrieve the EndpointSlices associated with the Service and ServiceImport
		if endpointSliceLabelKey != "" {
			endpointSliceList := new(discoveryv1.EndpointSliceList)
			opts := []client.ListOption{
				client.MatchingLabels(map[string]string{
					endpointSliceLabelKey: string(backendRef.Name),
				}),
				client.InNamespace(*backendRef.Namespace),
			}
			if err := r.client.List(ctx, endpointSliceList, opts...); err != nil {
				r.log.Error(err, "failed to get EndpointSlices", "namespace", string(*backendRef.Namespace),
					backendRefKind, string(backendRef.Name))
			} else {
				for _, endpointSlice := range endpointSliceList.Items {
					key := utils.NamespacedName(&endpointSlice).String()
					if !resourceMappings.allAssociatedEndpointSlices.Has(key) {
						resourceMappings.allAssociatedEndpointSlices.Insert(key)
						r.log.Info("added EndpointSlice to resource tree",
							"namespace", endpointSlice.Namespace,
							"name", endpointSlice.Name)
						gwcResource.EndpointSlices = append(gwcResource.EndpointSlices, &endpointSlice)
					}
				}
			}
		}
	}
}

// processSecurityPolicyObjectRefs adds the referenced resources in SecurityPolicies
// to the resourceTree
// - Secrets for OIDC and BasicAuth
// - BackendRefs for ExAuth
func (r *gatewayAPIReconciler) processSecurityPolicyObjectRefs(
	ctx context.Context, resourceTree *resource.Resources, resourceMap *resourceMappings,
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
				resource.KindSecurityPolicy,
				policy.Namespace,
				policy.Name,
				oidc.ClientSecret); err != nil {
				r.log.Error(err,
					"failed to process OIDC SecretRef for SecurityPolicy",
					"policy", policy, "secretRef", oidc.ClientSecret)
			}
		}

		// Add the referenced Secretes in APIKeyAuth to the resourceTree.
		apiKeyAuth := policy.Spec.APIKeyAuth
		if apiKeyAuth != nil {
			for _, credRef := range apiKeyAuth.CredentialRefs {
				if err := r.processSecretRef(
					ctx,
					resourceMap,
					resourceTree,
					resource.KindSecurityPolicy,
					policy.Namespace,
					policy.Name,
					credRef); err != nil {
					r.log.Error(err,
						"failed to process APIKeyAuth SecretRef for SecurityPolicy",
						"policy", policy, "secretRef", apiKeyAuth.CredentialRefs)
				}
			}
		}

		// Add the referenced Secrets in BasicAuth to the resourceTree
		basicAuth := policy.Spec.BasicAuth
		if basicAuth != nil {
			if err := r.processSecretRef(
				ctx,
				resourceMap,
				resourceTree,
				resource.KindSecurityPolicy,
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
				if extAuth.GRPC.BackendRef != nil {
					backendRef = *extAuth.GRPC.BackendRef
				}
				if len(extAuth.GRPC.BackendRefs) > 0 {
					if len(extAuth.GRPC.BackendRefs) != 0 {
						backendRef = extAuth.GRPC.BackendRefs[0].BackendObjectReference
					}
				}
			} else if extAuth.HTTP != nil {
				if extAuth.HTTP.BackendRef != nil {
					backendRef = *extAuth.HTTP.BackendRef
				}
				if len(extAuth.HTTP.BackendRefs) > 0 {
					if len(extAuth.HTTP.BackendRefs) != 0 {
						backendRef = extAuth.HTTP.BackendRefs[0].BackendObjectReference
					}
				}
			}
			if err := r.processBackendRef(
				ctx,
				resourceMap,
				resourceTree,
				resource.KindSecurityPolicy,
				policy.Namespace,
				policy.Name,
				backendRef); err != nil {
				r.log.Error(err,
					"failed to process ExtAuth BackendRef for SecurityPolicy",
					"policy", policy, "backendRef", backendRef)
			}
		}

		if policy.Spec.JWT != nil {
			for _, provider := range policy.Spec.JWT.Providers {
				if provider.LocalJWKS != nil &&
					provider.LocalJWKS.Type != nil &&
					*provider.LocalJWKS.Type == egv1a1.LocalJWKSTypeValueRef {
					if err := r.processConfigMapRef(
						ctx,
						resourceMap,
						resourceTree,
						resource.KindClientTrafficPolicy,
						policy.Namespace,
						policy.Name,
						gwapiv1.SecretObjectReference{
							Group: &provider.LocalJWKS.ValueRef.Group,
							Kind:  &provider.LocalJWKS.ValueRef.Kind,
							Name:  provider.LocalJWKS.ValueRef.Name,
						}); err != nil {
						r.log.Error(err, "failed to process LocalJWKS ConfigMap", "policy", policy, "localJWKS", provider.LocalJWKS)
					}
				} else if provider.RemoteJWKS != nil {
					for _, br := range provider.RemoteJWKS.BackendRefs {
						if err := r.processBackendRef(
							ctx,
							resourceMap,
							resourceTree,
							resource.KindSecurityPolicy,
							policy.Namespace,
							policy.Name,
							br.BackendObjectReference); err != nil {
							r.log.Error(err,
								"failed to process RemoteJWKS BackendRef for SecurityPolicy",
								"policy", policy, "backendRef", br.BackendObjectReference)
						}
					}
				}
			}
		}
	}
}

// processBackendRef adds the referenced BackendRef to the resourceMap for later processBackendRefs.
// If BackendRef exists in a different namespace and there is a ReferenceGrant, adds ReferenceGrant to the resourceTree.
func (r *gatewayAPIReconciler) processBackendRef(
	ctx context.Context,
	resourceMap *resourceMappings,
	resourceTree *resource.Resources,
	ownerKind string,
	ownerNS string,
	ownerName string,
	backendRef gwapiv1.BackendObjectReference,
) error {
	backendNamespace := gatewayapi.NamespaceDerefOr(backendRef.Namespace, ownerNS)
	resourceMap.allAssociatedBackendRefs.Insert(gwapiv1.BackendObjectReference{
		Group:     backendRef.Group,
		Kind:      backendRef.Kind,
		Namespace: gatewayapi.NamespacePtr(backendNamespace),
		Name:      backendRef.Name,
	})

	if backendNamespace != ownerNS {
		from := ObjectKindNamespacedName{
			kind:      ownerKind,
			namespace: ownerNS,
			name:      ownerName,
		}
		to := ObjectKindNamespacedName{
			kind:      gatewayapi.KindDerefOr(backendRef.Kind, resource.KindService),
			namespace: backendNamespace,
			name:      string(backendRef.Name),
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
			if !resourceMap.allAssociatedReferenceGrants.Has(utils.NamespacedName(refGrant).String()) {
				resourceMap.allAssociatedReferenceGrants.Insert(utils.NamespacedName(refGrant).String())
				resourceTree.ReferenceGrants = append(resourceTree.ReferenceGrants, refGrant)
				r.log.Info("added ReferenceGrant to resource map", "namespace", refGrant.Namespace,
					"name", refGrant.Name)
			}
		}
	}
	return nil
}

// processOIDCHMACSecret adds the OIDC HMAC Secret to the resourceTree.
// The OIDC HMAC Secret is created by the CertGen job and is used by SecurityPolicy
// to configure OAuth2 filters.
func (r *gatewayAPIReconciler) processOIDCHMACSecret(ctx context.Context, resourceTree *resource.Resources, resourceMap *resourceMappings) {
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

	key := utils.NamespacedName(&secret).String()
	if !resourceMap.allAssociatedSecrets.Has(key) {
		resourceMap.allAssociatedSecrets.Insert(key)
		resourceTree.Secrets = append(resourceTree.Secrets, &secret)
		r.log.Info("processing OIDC HMAC Secret", "namespace", r.namespace, "name", oidcHMACSecretName)
	}
}

// processEnvoyTLSSecret adds the Envoy TLS Secret to the resourceTree.
// The Envoy TLS Secret is created by the CertGen job and is used by envoy to establish
// TLS connections to the rate limit service.
func (r *gatewayAPIReconciler) processEnvoyTLSSecret(ctx context.Context, resourceTree *resource.Resources, resourceMap *resourceMappings) {
	var (
		secret corev1.Secret
		err    error
	)

	err = r.client.Get(ctx,
		types.NamespacedName{Namespace: r.namespace, Name: envoyTLSSecretName},
		&secret,
	)
	if err != nil {
		r.log.Error(err,
			"failed to process Envoy TLS Secret",
			"namespace", r.namespace, "name", envoyTLSSecretName)
		return
	}

	key := utils.NamespacedName(&secret).String()
	if !resourceMap.allAssociatedSecrets.Has(key) {
		resourceMap.allAssociatedSecrets.Insert(key)
		resourceTree.Secrets = append(resourceTree.Secrets, &secret)
		r.log.Info("processing Envoy TLS Secret", "namespace", r.namespace, "name", envoyTLSSecretName)
	}
}

// processSecretRef adds the referenced Secret to the resourceTree if it's valid.
// - If it exists in the same namespace as the owner.
// - If it exists in a different namespace, and there is a ReferenceGrant.
func (r *gatewayAPIReconciler) processSecretRef(
	ctx context.Context,
	resourceMap *resourceMappings,
	resourceTree *resource.Resources,
	ownerKind string,
	ownerNS string,
	ownerName string,
	secretRef gwapiv1.SecretObjectReference,
) error {
	secret := new(corev1.Secret)
	secretNS := gatewayapi.NamespaceDerefOr(secretRef.Namespace, ownerNS)
	err := r.client.Get(ctx,
		types.NamespacedName{Namespace: secretNS, Name: string(secretRef.Name)},
		secret,
	)
	if err != nil && kerrors.IsNotFound(err) {
		return fmt.Errorf("unable to find the Secret: %s/%s", secretNS, string(secretRef.Name))
	}

	if secretNS != ownerNS {
		from := ObjectKindNamespacedName{
			kind:      ownerKind,
			namespace: ownerNS,
			name:      ownerName,
		}
		to := ObjectKindNamespacedName{
			kind:      resource.KindSecret,
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
			if !resourceMap.allAssociatedReferenceGrants.Has(utils.NamespacedName(refGrant).String()) {
				resourceMap.allAssociatedReferenceGrants.Insert(utils.NamespacedName(refGrant).String())
				resourceTree.ReferenceGrants = append(resourceTree.ReferenceGrants, refGrant)
				r.log.Info("added ReferenceGrant to resource map", "namespace", refGrant.Namespace,
					"name", refGrant.Name)
			}
		}
	}
	resourceMap.allAssociatedNamespaces.Insert(secretNS)
	key := utils.NamespacedName(secret).String()
	if !resourceMap.allAssociatedSecrets.Has(key) {
		resourceMap.allAssociatedSecrets.Insert(key)
		resourceTree.Secrets = append(resourceTree.Secrets, secret)
		r.log.Info("processing Secret", "namespace", secretNS, "name", string(secretRef.Name))
	}
	return nil
}

// processCtpConfigMapRefs adds the referenced ConfigMaps in ClientTrafficPolicies
// to the resourceTree
func (r *gatewayAPIReconciler) processCtpConfigMapRefs(
	ctx context.Context, resourceTree *resource.Resources, resourceMap *resourceMappings,
) {
	for _, policy := range resourceTree.ClientTrafficPolicies {
		tls := policy.Spec.TLS

		if tls != nil && tls.ClientValidation != nil {
			for _, caCertRef := range tls.ClientValidation.CACertificateRefs {
				if caCertRef.Kind != nil && string(*caCertRef.Kind) == resource.KindConfigMap {
					if err := r.processConfigMapRef(
						ctx,
						resourceMap,
						resourceTree,
						resource.KindClientTrafficPolicy,
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
				} else if caCertRef.Kind == nil || string(*caCertRef.Kind) == resource.KindSecret {
					if err := r.processSecretRef(
						ctx,
						resourceMap,
						resourceTree,
						resource.KindClientTrafficPolicy,
						policy.Namespace,
						policy.Name,
						caCertRef); err != nil {
						r.log.Error(err,
							"failed to process CACertificateRef for ClientTrafficPolicy",
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
	resourceTree *resource.Resources,
	ownerKind string,
	ownerNS string,
	ownerName string,
	configMapRef gwapiv1.SecretObjectReference,
) error {
	configMap := new(corev1.ConfigMap)
	configMapNS := gatewayapi.NamespaceDerefOr(configMapRef.Namespace, ownerNS)
	err := r.client.Get(ctx,
		types.NamespacedName{Namespace: configMapNS, Name: string(configMapRef.Name)},
		configMap,
	)
	if err != nil && kerrors.IsNotFound(err) {
		return fmt.Errorf("unable to find the ConfigMap: %s/%s", configMapNS, string(configMapRef.Name))
	}

	if configMapNS != ownerNS {
		from := ObjectKindNamespacedName{
			kind:      ownerKind,
			namespace: ownerNS,
			name:      ownerName,
		}
		to := ObjectKindNamespacedName{
			kind:      resource.KindConfigMap,
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
			if !resourceMap.allAssociatedReferenceGrants.Has(utils.NamespacedName(refGrant).String()) {
				resourceMap.allAssociatedReferenceGrants.Insert(utils.NamespacedName(refGrant).String())
				resourceTree.ReferenceGrants = append(resourceTree.ReferenceGrants, refGrant)
				r.log.Info("added ReferenceGrant to resource map", "namespace", refGrant.Namespace,
					"name", refGrant.Name)
			}
		}
	}
	resourceMap.allAssociatedNamespaces.Insert(configMapNS)
	if !resourceMap.allAssociatedConfigMaps.Has(utils.NamespacedName(configMap).String()) {
		resourceMap.allAssociatedConfigMaps.Insert(utils.NamespacedName(configMap).String())
		resourceTree.ConfigMaps = append(resourceTree.ConfigMaps, configMap)
		r.log.Info("processing ConfigMap", "namespace", configMapNS, "name", string(configMapRef.Name))
	}
	return nil
}

// processBtpConfigMapRefs adds the referenced ConfigMaps in BackendTrafficPolicies
// to the resourceTree
func (r *gatewayAPIReconciler) processBtpConfigMapRefs(
	ctx context.Context, resourceTree *resource.Resources, resourceMap *resourceMappings,
) {
	for _, policy := range resourceTree.BackendTrafficPolicies {
		for _, ro := range policy.Spec.ResponseOverride {
			if ro.Response.Body != nil && ro.Response.Body.ValueRef != nil && string(ro.Response.Body.ValueRef.Kind) == resource.KindConfigMap {
				configMap := new(corev1.ConfigMap)
				err := r.client.Get(ctx,
					types.NamespacedName{Namespace: policy.Namespace, Name: string(ro.Response.Body.ValueRef.Name)},
					configMap,
				)
				// we don't return an error here, because we want to continue
				// reconciling the rest of the BackendTrafficPolicies despite that this
				// reference is invalid.
				// This BackendTrafficPolicies will be marked as invalid in its status
				// when translating to IR because the referenced configmap can't be
				// found.
				if err != nil {
					r.log.Error(err,
						"failed to process ResponseOverride ValueRef for BackendTrafficPolicy",
						"policy", policy, "ValueRef", ro.Response.Body.ValueRef.Name)
				}

				resourceMap.allAssociatedNamespaces.Insert(policy.Namespace)
				if !resourceMap.allAssociatedConfigMaps.Has(utils.NamespacedName(configMap).String()) {
					resourceMap.allAssociatedConfigMaps.Insert(utils.NamespacedName(configMap).String())
					resourceTree.ConfigMaps = append(resourceTree.ConfigMaps, configMap)
					r.log.Info("processing ConfigMap", "namespace", policy.Namespace, "name", string(ro.Response.Body.ValueRef.Name))
				}
			}
		}
	}
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
		if refGrant.Namespace != to.namespace {
			continue
		}

		var fromAllowed bool
		for _, refGrantFrom := range refGrant.Spec.From {
			if string(refGrantFrom.Kind) == from.kind && string(refGrantFrom.Namespace) == from.namespace {
				fromAllowed = true
				break
			}
		}

		if !fromAllowed {
			continue
		}

		var toAllowed bool
		for _, refGrantTo := range refGrant.Spec.To {
			if string(refGrantTo.Kind) == to.kind && (refGrantTo.Name == nil || *refGrantTo.Name == "" || string(*refGrantTo.Name) == to.name) {
				toAllowed = true
				break
			}
		}

		if !toAllowed {
			continue
		}

		return &refGrant, nil
	}

	// No ReferenceGrant found.
	return nil, nil
}

func (r *gatewayAPIReconciler) processGateways(ctx context.Context, managedGC *gwapiv1.GatewayClass, resourceMap *resourceMappings, resourceTree *resource.Resources) error {
	// Find gateways for the managedGC
	// Find the Gateways that reference this Class.
	gatewayList := &gwapiv1.GatewayList{}
	if err := r.client.List(ctx, gatewayList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(classGatewayIndex, managedGC.Name),
	}); err != nil {
		r.log.Error(err, "failed to list gateways for gatewayClass %s", managedGC.Name)
		return err
	}

	for _, gtw := range gatewayList.Items {
		gtw := gtw //nolint:copyloopvar
		if r.namespaceLabel != nil {
			if ok, err := r.checkObjectNamespaceLabels(&gtw); err != nil {
				r.log.Error(err, "failed to check namespace labels for gateway %s in namespace %s: %w", gtw.GetName(), gtw.GetNamespace())
				continue
			} else if !ok {
				continue
			}
		}

		r.log.Info("processing Gateway", "namespace", gtw.Namespace, "name", gtw.Name)
		resourceMap.allAssociatedNamespaces.Insert(gtw.Namespace)

		for _, listener := range gtw.Spec.Listeners {
			// Get Secret for gateway if it exists.
			if terminatesTLS(&listener) {
				for _, certRef := range listener.TLS.CertificateRefs {
					if refsSecret(&certRef) {
						if err := r.processSecretRef(
							ctx,
							resourceMap,
							resourceTree,
							resource.KindGateway,
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

		gtwNamespacedName := utils.NamespacedName(&gtw).String()
		// Route Processing

		if r.tlsRouteCRDExists {
			// Get TLSRoute objects and check if it exists.
			if err := r.processTLSRoutes(ctx, gtwNamespacedName, resourceMap, resourceTree); err != nil {
				return err
			}
		}

		// Get HTTPRoute objects and check if it exists.
		if err := r.processHTTPRoutes(ctx, gtwNamespacedName, resourceMap, resourceTree); err != nil {
			return err
		}

		if r.grpcRouteCRDExists {
			// Get GRPCRoute objects and check if it exists.
			if err := r.processGRPCRoutes(ctx, gtwNamespacedName, resourceMap, resourceTree); err != nil {
				return err
			}
		}

		if r.tcpRouteCRDExists {
			// Get TCPRoute objects and check if it exists.
			if err := r.processTCPRoutes(ctx, gtwNamespacedName, resourceMap, resourceTree); err != nil {
				return err
			}
		}

		if r.udpRouteCRDExists {
			// Get UDPRoute objects and check if it exists.
			if err := r.processUDPRoutes(ctx, gtwNamespacedName, resourceMap, resourceTree); err != nil {
				return err
			}
		}

		// Discard Status to reduce memory consumption in watchable
		// It will be recomputed by the gateway-api layer
		gtw.Status = gwapiv1.GatewayStatus{}

		if err := r.processGatewayParamsRef(ctx, &gtw, resourceMap, resourceTree); err != nil {
			// Update the Gateway status to not accepted if there is an error processing the parametersRef.
			// These not-accepted gateways will not be processed by the gateway-api layer, but their status will be
			// updated in the gateway-api layer along with other gateways. This is to avoid the potential race condition
			// of updating the status in both the controller and the gateway-api layer.
			status.UpdateGatewayStatusNotAccepted(&gtw, gwapiv1.GatewayReasonInvalidParameters, err.Error())
			r.log.Error(err, "failed to process infrastructure.parametersRef for gateway", "namespace", gtw.Namespace, "name", gtw.Name)
		}

		if !resourceMap.allAssociatedGateways.Has(gtwNamespacedName) {
			resourceMap.allAssociatedGateways.Insert(gtwNamespacedName)
			resourceTree.Gateways = append(resourceTree.Gateways, &gtw)
		}
	}

	return nil
}

// processEnvoyPatchPolicies adds EnvoyPatchPolicies to the resourceTree
func (r *gatewayAPIReconciler) processEnvoyPatchPolicies(ctx context.Context, resourceTree *resource.Resources, resourceMap *resourceMappings) error {
	envoyPatchPolicies := egv1a1.EnvoyPatchPolicyList{}
	if err := r.client.List(ctx, &envoyPatchPolicies); err != nil {
		return fmt.Errorf("error listing EnvoyPatchPolicies: %w", err)
	}

	for _, policy := range envoyPatchPolicies.Items {
		envoyPatchPolicy := policy //nolint:copyloopvar
		// Discard Status to reduce memory consumption in watchable
		// It will be recomputed by the gateway-api layer
		envoyPatchPolicy.Status = gwapiv1a2.PolicyStatus{}
		if !resourceMap.allAssociatedEnvoyPatchPolicies.Has(utils.NamespacedName(&envoyPatchPolicy).String()) {
			resourceMap.allAssociatedEnvoyPatchPolicies.Insert(utils.NamespacedName(&envoyPatchPolicy).String())
			resourceTree.EnvoyPatchPolicies = append(resourceTree.EnvoyPatchPolicies, &envoyPatchPolicy)
		}
	}
	return nil
}

// processClientTrafficPolicies adds ClientTrafficPolicies to the resourceTree
func (r *gatewayAPIReconciler) processClientTrafficPolicies(
	ctx context.Context, resourceTree *resource.Resources, resourceMap *resourceMappings,
) error {
	clientTrafficPolicies := egv1a1.ClientTrafficPolicyList{}
	if err := r.client.List(ctx, &clientTrafficPolicies); err != nil {
		return fmt.Errorf("error listing ClientTrafficPolicies: %w", err)
	}

	for _, policy := range clientTrafficPolicies.Items {
		clientTrafficPolicy := policy //nolint:copyloopvar
		// Discard Status to reduce memory consumption in watchable
		// It will be recomputed by the gateway-api layer
		clientTrafficPolicy.Status = gwapiv1a2.PolicyStatus{}
		if !resourceMap.allAssociatedClientTrafficPolicies.Has(utils.NamespacedName(&clientTrafficPolicy).String()) {
			resourceMap.allAssociatedClientTrafficPolicies.Insert(utils.NamespacedName(&clientTrafficPolicy).String())
			resourceTree.ClientTrafficPolicies = append(resourceTree.ClientTrafficPolicies, &clientTrafficPolicy)
		}
	}

	r.processCtpConfigMapRefs(ctx, resourceTree, resourceMap)

	return nil
}

// processBackendTrafficPolicies adds BackendTrafficPolicies to the resourceTree
func (r *gatewayAPIReconciler) processBackendTrafficPolicies(ctx context.Context, resourceTree *resource.Resources, resourceMap *resourceMappings,
) error {
	backendTrafficPolicies := egv1a1.BackendTrafficPolicyList{}
	if err := r.client.List(ctx, &backendTrafficPolicies); err != nil {
		return fmt.Errorf("error listing BackendTrafficPolicies: %w", err)
	}

	for _, policy := range backendTrafficPolicies.Items {
		backendTrafficPolicy := policy //nolint:copyloopvar
		// Discard Status to reduce memory consumption in watchable
		// It will be recomputed by the gateway-api layer
		backendTrafficPolicy.Status = gwapiv1a2.PolicyStatus{}
		if !resourceMap.allAssociatedBackendTrafficPolicies.Has(utils.NamespacedName(&backendTrafficPolicy).String()) {
			resourceMap.allAssociatedBackendTrafficPolicies.Insert(utils.NamespacedName(&backendTrafficPolicy).String())
			resourceTree.BackendTrafficPolicies = append(resourceTree.BackendTrafficPolicies, &backendTrafficPolicy)
		}
	}
	r.processBtpConfigMapRefs(ctx, resourceTree, resourceMap)
	return nil
}

// processSecurityPolicies adds SecurityPolicies and their referenced resources to the resourceTree
func (r *gatewayAPIReconciler) processSecurityPolicies(
	ctx context.Context, resourceTree *resource.Resources, resourceMap *resourceMappings,
) error {
	securityPolicies := egv1a1.SecurityPolicyList{}
	if err := r.client.List(ctx, &securityPolicies); err != nil {
		return fmt.Errorf("error listing SecurityPolicies: %w", err)
	}

	for _, policy := range securityPolicies.Items {
		securityPolicy := policy //nolint:copyloopvar
		// Discard Status to reduce memory consumption in watchable
		// It will be recomputed by the gateway-api layer
		securityPolicy.Status = gwapiv1a2.PolicyStatus{}
		if !resourceMap.allAssociatedSecurityPolicies.Has(utils.NamespacedName(&securityPolicy).String()) {
			resourceMap.allAssociatedSecurityPolicies.Insert(utils.NamespacedName(&securityPolicy).String())
			resourceTree.SecurityPolicies = append(resourceTree.SecurityPolicies, &securityPolicy)
		}
	}

	// Add the referenced Resources in SecurityPolicies to the resourceTree
	r.processSecurityPolicyObjectRefs(ctx, resourceTree, resourceMap)

	// Add the OIDC HMAC Secret to the resourceTree
	r.processOIDCHMACSecret(ctx, resourceTree, resourceMap)

	// Add the Envoy TLS Secret to the resourceTree
	r.processEnvoyTLSSecret(ctx, resourceTree, resourceMap)
	return nil
}

// processBackendTLSPolicies adds BackendTLSPolicies and their referenced resources to the resourceTree
func (r *gatewayAPIReconciler) processBackendTLSPolicies(
	ctx context.Context, resourceTree *resource.Resources, resourceMap *resourceMappings,
) error {
	backendTLSPolicies := gwapiv1a3.BackendTLSPolicyList{}
	if err := r.client.List(ctx, &backendTLSPolicies); err != nil {
		return fmt.Errorf("error listing BackendTLSPolicies: %w", err)
	}

	for _, policy := range backendTLSPolicies.Items {
		backendTLSPolicy := policy //nolint:copyloopvar
		// Discard Status to reduce memory consumption in watchable
		// It will be recomputed by the gateway-api layer
		backendTLSPolicy.Status = gwapiv1a2.PolicyStatus{}
		if !resourceMap.allAssociatedBackendTLSPolicies.Has(utils.NamespacedName(&backendTLSPolicy).String()) {
			resourceMap.allAssociatedBackendTLSPolicies.Insert(utils.NamespacedName(&backendTLSPolicy).String())
			resourceTree.BackendTLSPolicies = append(resourceTree.BackendTLSPolicies, &backendTLSPolicy)
		}
	}

	// Add the referenced Secrets and ConfigMaps in BackendTLSPolicies to the resourceTree.
	r.processBackendTLSPolicyRefs(ctx, resourceTree, resourceMap)
	return nil
}

// processBackends adds Backends to the resourceTree
func (r *gatewayAPIReconciler) processBackends(ctx context.Context, resourceTree *resource.Resources) error {
	backends := egv1a1.BackendList{}
	if err := r.client.List(ctx, &backends); err != nil {
		return fmt.Errorf("error listing Backends: %w", err)
	}

	for _, backend := range backends.Items {
		backend := backend //nolint:copyloopvar
		// Discard Status to reduce memory consumption in watchable
		// It will be recomputed by the gateway-api layer
		backend.Status = egv1a1.BackendStatus{}
		resourceTree.Backends = append(resourceTree.Backends, &backend)
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
	// Upon leader election, we retrigger the reconciliation process to allow the elected leader to
	// process status updates and infrastructure changes. This step is crucial for synchronizing resources
	// that may have been altered or introduced while there was no elected leader.
	if err := c.Watch(NewWatchAndReconcileSource(mgr.Elected(), &gwapiv1.GatewayClass{}, handler.EnqueueRequestsFromMapFunc(r.enqueueClass))); err != nil {
		return fmt.Errorf("failed to watch GatewayClass: %w", err)
	}

	if err := c.Watch(
		source.Kind(mgr.GetCache(), &gwapiv1.GatewayClass{},
			handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, gc *gwapiv1.GatewayClass) []reconcile.Request {
				return r.enqueueClass(ctx, gc)
			}),
			&predicate.TypedGenerationChangedPredicate[*gwapiv1.GatewayClass]{},
			predicate.NewTypedPredicateFuncs(r.hasMatchingController))); err != nil {
		return fmt.Errorf("failed to watch GatewayClass: %w", err)
	}

	r.epCRDExists = r.crdExists(mgr, resource.KindEnvoyProxy, egv1a1.GroupVersion.String())
	if !r.epCRDExists {
		r.log.Info("EnvoyProxy CRD not found, skipping EnvoyProxy watch")
	} else {
		epPredicates := []predicate.TypedPredicate[*egv1a1.EnvoyProxy]{
			&predicate.TypedGenerationChangedPredicate[*egv1a1.EnvoyProxy]{},
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
		if err := addEnvoyProxyIndexers(ctx, mgr); err != nil {
			return err
		}
	}

	// Watch Gateway CRUDs and reconcile affected GatewayClass.
	gPredicates := []predicate.TypedPredicate[*gwapiv1.Gateway]{
		metadataPredicate[*gwapiv1.Gateway](),
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
	httprPredicates := commonPredicates[*gwapiv1.HTTPRoute]()
	if r.namespaceLabel != nil {
		httprPredicates = append(httprPredicates, predicate.NewTypedPredicateFuncs(func(hr *gwapiv1.HTTPRoute) bool {
			return r.hasMatchingNamespaceLabels(hr)
		}))
	}
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &gwapiv1.HTTPRoute{},
			handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, route *gwapiv1.HTTPRoute) []reconcile.Request {
				return r.enqueueClass(ctx, route)
			}),
			httprPredicates...)); err != nil {
		return err
	}
	if err := addHTTPRouteIndexers(ctx, mgr); err != nil {
		return err
	}

	// TODO: Remove this optional check once most cloud providers and service meshes support GRPCRoute v1
	r.grpcRouteCRDExists = r.crdExists(mgr, resource.KindGRPCRoute, gwapiv1.GroupVersion.String())
	if !r.grpcRouteCRDExists {
		r.log.Info("GRPCRoute CRD not found, skipping GRPCRoute watch")
	} else {
		// Watch GRPCRoute CRUDs and process affected Gateways.
		grpcrPredicates := commonPredicates[*gwapiv1.GRPCRoute]()
		if r.namespaceLabel != nil {
			grpcrPredicates = append(grpcrPredicates, predicate.NewTypedPredicateFuncs(func(grpc *gwapiv1.GRPCRoute) bool {
				return r.hasMatchingNamespaceLabels(grpc)
			}))
		}
		if err := c.Watch(
			source.Kind(mgr.GetCache(), &gwapiv1.GRPCRoute{},
				handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, route *gwapiv1.GRPCRoute) []reconcile.Request {
					return r.enqueueClass(ctx, route)
				}),
				grpcrPredicates...)); err != nil {
			return err
		}
		if err := addGRPCRouteIndexers(ctx, mgr); err != nil {
			return err
		}
	}

	r.tlsRouteCRDExists = r.crdExists(mgr, resource.KindTLSRoute, gwapiv1a2.GroupVersion.String())
	if !r.tlsRouteCRDExists {
		r.log.Info("TLSRoute CRD not found, skipping TLSRoute watch")
	} else {
		// Watch TLSRoute CRUDs and process affected Gateways.
		tlsrPredicates := commonPredicates[*gwapiv1a2.TLSRoute]()
		if r.namespaceLabel != nil {
			tlsrPredicates = append(tlsrPredicates, predicate.NewTypedPredicateFuncs(func(route *gwapiv1a2.TLSRoute) bool {
				return r.hasMatchingNamespaceLabels(route)
			}))
		}
		if err := c.Watch(
			source.Kind(mgr.GetCache(), &gwapiv1a2.TLSRoute{},
				handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, route *gwapiv1a2.TLSRoute) []reconcile.Request {
					return r.enqueueClass(ctx, route)
				}),
				tlsrPredicates...)); err != nil {
			return err
		}
		if err := addTLSRouteIndexers(ctx, mgr); err != nil {
			return err
		}
	}

	r.udpRouteCRDExists = r.crdExists(mgr, resource.KindUDPRoute, gwapiv1a2.GroupVersion.String())
	if !r.udpRouteCRDExists {
		r.log.Info("UDPRoute CRD not found, skipping UDPRoute watch")
	} else {
		// Watch UDPRoute CRUDs and process affected Gateways.
		udprPredicates := commonPredicates[*gwapiv1a2.UDPRoute]()
		if r.namespaceLabel != nil {
			udprPredicates = append(udprPredicates, predicate.NewTypedPredicateFuncs(func(route *gwapiv1a2.UDPRoute) bool {
				return r.hasMatchingNamespaceLabels(route)
			}))
		}
		if err := c.Watch(
			source.Kind(mgr.GetCache(), &gwapiv1a2.UDPRoute{},
				handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, route *gwapiv1a2.UDPRoute) []reconcile.Request {
					return r.enqueueClass(ctx, route)
				}),
				udprPredicates...)); err != nil {
			return err
		}
		if err := addUDPRouteIndexers(ctx, mgr); err != nil {
			return err
		}
	}

	r.tcpRouteCRDExists = r.crdExists(mgr, resource.KindTCPRoute, gwapiv1a2.GroupVersion.String())
	if !r.tcpRouteCRDExists {
		r.log.Info("TCPRoute CRD not found, skipping TCPRoute watch")
	} else {
		// Watch TCPRoute CRUDs and process affected Gateways.
		tcprPredicates := commonPredicates[*gwapiv1a2.TCPRoute]()
		if r.namespaceLabel != nil {
			tcprPredicates = append(tcprPredicates, predicate.NewTypedPredicateFuncs(func(route *gwapiv1a2.TCPRoute) bool {
				return r.hasMatchingNamespaceLabels(route)
			}))
		}
		if err := c.Watch(
			source.Kind(mgr.GetCache(), &gwapiv1a2.TCPRoute{},
				handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, route *gwapiv1a2.TCPRoute) []reconcile.Request {
					return r.enqueueClass(ctx, route)
				}),
				tcprPredicates...)); err != nil {
			return err
		}
		if err := addTCPRouteIndexers(ctx, mgr); err != nil {
			return err
		}
	}

	// Watch Service CRUDs and process affected *Route objects.
	servicePredicates := []predicate.TypedPredicate[*corev1.Service]{
		predicate.NewTypedPredicateFuncs(func(svc *corev1.Service) bool {
			return r.validateServiceForReconcile(svc)
		}),
	}
	if r.namespaceLabel != nil {
		servicePredicates = append(servicePredicates, predicate.NewTypedPredicateFuncs(func(svc *corev1.Service) bool {
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

	// Watch ServiceImport CRUDs and process affected *Route objects.
	r.serviceImportCRDExists = r.crdExists(mgr, resource.KindServiceImport, mcsapiv1a1.GroupVersion.String())
	if !r.serviceImportCRDExists {
		r.log.Info("ServiceImport CRD not found, skipping ServiceImport watch")
	} else {
		if err := c.Watch(
			source.Kind(mgr.GetCache(), &mcsapiv1a1.ServiceImport{},
				handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, si *mcsapiv1a1.ServiceImport) []reconcile.Request {
					return r.enqueueClass(ctx, si)
				}),
				predicate.TypedGenerationChangedPredicate[*mcsapiv1a1.ServiceImport]{},
				predicate.NewTypedPredicateFuncs(func(si *mcsapiv1a1.ServiceImport) bool {
					return r.validateServiceImportForReconcile(si)
				}))); err != nil {
			// ServiceImport is not available in the cluster, skip the watch and not throw error.
			r.log.Info("unable to watch ServiceImport: %s", err.Error())
		}
	}

	// Watch EndpointSlice CRUDs and process affected *Route objects.
	esPredicates := []predicate.TypedPredicate[*discoveryv1.EndpointSlice]{
		predicate.TypedGenerationChangedPredicate[*discoveryv1.EndpointSlice]{},
		predicate.NewTypedPredicateFuncs(func(eps *discoveryv1.EndpointSlice) bool {
			return r.validateEndpointSliceForReconcile(eps)
		}),
	}
	if r.namespaceLabel != nil {
		esPredicates = append(esPredicates, predicate.NewTypedPredicateFuncs(func(eps *discoveryv1.EndpointSlice) bool {
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

	r.backendCRDExists = r.crdExists(mgr, resource.KindBackend, egv1a1.GroupVersion.String())
	if !r.backendCRDExists {
		r.log.Info("Backend CRD not found, skipping Backend watch")
	} else if r.envoyGateway.ExtensionAPIs != nil && r.envoyGateway.ExtensionAPIs.EnableBackend {
		// Watch Backend CRUDs and process affected *Route objects.
		backendPredicates := []predicate.TypedPredicate[*egv1a1.Backend]{
			predicate.TypedGenerationChangedPredicate[*egv1a1.Backend]{},
			predicate.NewTypedPredicateFuncs(func(be *egv1a1.Backend) bool {
				return r.validateBackendForReconcile(be)
			}),
		}
		if r.namespaceLabel != nil {
			backendPredicates = append(backendPredicates, predicate.NewTypedPredicateFuncs(func(be *egv1a1.Backend) bool {
				return r.hasMatchingNamespaceLabels(be)
			}))
		}
		if err := c.Watch(
			source.Kind(mgr.GetCache(), &egv1a1.Backend{},
				handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, be *egv1a1.Backend) []reconcile.Request {
					return r.enqueueClass(ctx, be)
				}),
				backendPredicates...)); err != nil {
			return err
		}

		if err := addBackendIndexers(ctx, mgr); err != nil {
			return err
		}
	}

	// Watch Node CRUDs to update Gateway Address exposed by Service of type NodePort.
	// Node creation/deletion and ExternalIP updates would require update in the Gateway
	nPredicates := []predicate.TypedPredicate[*corev1.Node]{
		predicate.TypedGenerationChangedPredicate[*corev1.Node]{},
		predicate.NewTypedPredicateFuncs(func(node *corev1.Node) bool {
			return r.handleNode(node)
		}),
	}
	if r.namespaceLabel != nil {
		nPredicates = append(nPredicates, predicate.NewTypedPredicateFuncs(func(node *corev1.Node) bool {
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

	// Watch ConfigMap CRUDs and process affected EG Resources.
	configMapPredicates := []predicate.TypedPredicate[*corev1.ConfigMap]{
		predicate.NewTypedPredicateFuncs(func(cm *corev1.ConfigMap) bool {
			return r.validateConfigMapForReconcile(cm)
		}),
	}
	if r.namespaceLabel != nil {
		configMapPredicates = append(configMapPredicates, predicate.NewTypedPredicateFuncs(func(cm *corev1.ConfigMap) bool {
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
		rgPredicates = append(rgPredicates, predicate.NewTypedPredicateFuncs(func(rg *gwapiv1b1.ReferenceGrant) bool {
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
	deploymentPredicates := []predicate.TypedPredicate[*appsv1.Deployment]{
		predicate.NewTypedPredicateFuncs(func(deploy *appsv1.Deployment) bool {
			return r.validateObjectForReconcile(deploy)
		}),
	}
	if r.namespaceLabel != nil {
		deploymentPredicates = append(deploymentPredicates, predicate.NewTypedPredicateFuncs(func(deploy *appsv1.Deployment) bool {
			return r.hasMatchingNamespaceLabels(deploy)
		}))
	}
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &appsv1.Deployment{},
			handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, deploy *appsv1.Deployment) []reconcile.Request {
				return r.enqueueClass(ctx, deploy)
			}),
			deploymentPredicates...)); err != nil {
		return err
	}

	// Watch DaemonSet CRUDs and process affected Gateways.
	daemonsetPredicates := []predicate.TypedPredicate[*appsv1.DaemonSet]{
		predicate.NewTypedPredicateFuncs(func(daemonset *appsv1.DaemonSet) bool {
			return r.validateObjectForReconcile(daemonset)
		}),
	}
	if r.namespaceLabel != nil {
		daemonsetPredicates = append(daemonsetPredicates, predicate.NewTypedPredicateFuncs(func(daemonset *appsv1.DaemonSet) bool {
			return r.hasMatchingNamespaceLabels(daemonset)
		}))
	}
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &appsv1.DaemonSet{},
			handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, daemonset *appsv1.DaemonSet) []reconcile.Request {
				return r.enqueueClass(ctx, daemonset)
			}),
			daemonsetPredicates...)); err != nil {
		return err
	}

	r.eppCRDExists = r.crdExists(mgr, resource.KindEnvoyPatchPolicy, egv1a1.GroupVersion.String())
	if !r.eppCRDExists {
		r.log.Info("EnvoyPatchPolicy CRD not found, skipping EnvoyPatchPolicy watch")
	} else if r.envoyGateway.ExtensionAPIs != nil && r.envoyGateway.ExtensionAPIs.EnableEnvoyPatchPolicy {
		// Watch EnvoyPatchPolicy if enabled in config
		eppPredicates := []predicate.TypedPredicate[*egv1a1.EnvoyPatchPolicy]{
			predicate.TypedGenerationChangedPredicate[*egv1a1.EnvoyPatchPolicy]{},
		}
		if r.namespaceLabel != nil {
			eppPredicates = append(eppPredicates, predicate.NewTypedPredicateFuncs(func(epp *egv1a1.EnvoyPatchPolicy) bool {
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

	r.ctpCRDExists = r.crdExists(mgr, resource.KindClientTrafficPolicy, egv1a1.GroupVersion.String())
	if !r.ctpCRDExists {
		r.log.Info("ClientTrafficPolicy CRD not found, skipping ClientTrafficPolicy watch")
	} else {
		// Watch ClientTrafficPolicy
		ctpPredicates := []predicate.TypedPredicate[*egv1a1.ClientTrafficPolicy]{
			predicate.TypedGenerationChangedPredicate[*egv1a1.ClientTrafficPolicy]{},
		}
		if r.namespaceLabel != nil {
			ctpPredicates = append(ctpPredicates, predicate.NewTypedPredicateFuncs(func(ctp *egv1a1.ClientTrafficPolicy) bool {
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
	}

	r.btpCRDExists = r.crdExists(mgr, resource.KindBackendTrafficPolicy, egv1a1.GroupVersion.String())
	if !r.btpCRDExists {
		r.log.Info("BackendTrafficPolicy CRD not found, skipping BackendTrafficPolicy watch")
	} else {
		// Watch BackendTrafficPolicy
		btpPredicates := []predicate.TypedPredicate[*egv1a1.BackendTrafficPolicy]{
			predicate.TypedGenerationChangedPredicate[*egv1a1.BackendTrafficPolicy]{},
		}
		if r.namespaceLabel != nil {
			btpPredicates = append(btpPredicates, predicate.NewTypedPredicateFuncs(func(btp *egv1a1.BackendTrafficPolicy) bool {
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

		if err := addBtpIndexers(ctx, mgr); err != nil {
			return err
		}
	}

	r.spCRDExists = r.crdExists(mgr, resource.KindSecurityPolicy, egv1a1.GroupVersion.String())
	if !r.spCRDExists {
		r.log.Info("SecurityPolicy CRD not found, skipping SecurityPolicy watch")
	} else {
		// Watch SecurityPolicy
		spPredicates := []predicate.TypedPredicate[*egv1a1.SecurityPolicy]{
			predicate.TypedGenerationChangedPredicate[*egv1a1.SecurityPolicy]{},
		}
		if r.namespaceLabel != nil {
			spPredicates = append(spPredicates, predicate.NewTypedPredicateFuncs(func(sp *egv1a1.SecurityPolicy) bool {
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
	}

	r.bTLSPolicyCRDExists = r.crdExists(mgr, resource.KindBackendTLSPolicy, gwapiv1a3.GroupVersion.String())
	if !r.bTLSPolicyCRDExists {
		r.log.Info("BackendTLSPolicy CRD not found, skipping BackendTLSPolicy watch")
	} else {
		// Watch BackendTLSPolicy
		btlsPredicates := []predicate.TypedPredicate[*gwapiv1a3.BackendTLSPolicy]{
			predicate.TypedGenerationChangedPredicate[*gwapiv1a3.BackendTLSPolicy]{},
		}
		if r.namespaceLabel != nil {
			btlsPredicates = append(btlsPredicates, predicate.NewTypedPredicateFuncs(func(btp *gwapiv1a3.BackendTLSPolicy) bool {
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
	}

	r.eepCRDExists = r.crdExists(mgr, resource.KindEnvoyExtensionPolicy, egv1a1.GroupVersion.String())
	if !r.eepCRDExists {
		r.log.Info("EnvoyExtensionPolicy CRD not found, skipping EnvoyExtensionPolicy watch")
	} else {
		// Watch EnvoyExtensionPolicy
		eepPredicates := []predicate.TypedPredicate[*egv1a1.EnvoyExtensionPolicy]{
			predicate.TypedGenerationChangedPredicate[*egv1a1.EnvoyExtensionPolicy]{},
		}
		if r.namespaceLabel != nil {
			eepPredicates = append(eepPredicates, predicate.NewTypedPredicateFuncs(func(eep *egv1a1.EnvoyExtensionPolicy) bool {
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
	}

	r.log.Info("Watching gatewayAPI related objects")

	// Watch any additional GVKs from the registered extension.
	uPredicates := []predicate.TypedPredicate[*unstructured.Unstructured]{
		predicate.TypedGenerationChangedPredicate[*unstructured.Unstructured]{},
	}
	if r.namespaceLabel != nil {
		uPredicates = append(uPredicates, predicate.NewTypedPredicateFuncs(func(obj *unstructured.Unstructured) bool {
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
	for _, gvk := range r.extServerPolicies {
		u := &unstructured.Unstructured{}
		u.SetGroupVersionKind(gvk)
		if err := c.Watch(source.Kind(mgr.GetCache(), u,
			handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, si *unstructured.Unstructured) []reconcile.Request {
				return r.enqueueClass(ctx, si)
			}),
			uPredicates...)); err != nil {
			return err
		}
		r.log.Info("Watching additional policy resource", "resource", gvk.String())
	}

	r.hrfCRDExists = r.crdExists(mgr, resource.KindHTTPRouteFilter, egv1a1.GroupVersion.String())
	if !r.hrfCRDExists {
		r.log.Info("HTTPRouteFilter CRD not found, skipping HTTPRouteFilter watch")
	} else {
		// Watch HTTPRouteFilter CRUDs and process affected HTTPRoute objects.
		httpRouteFilter := []predicate.TypedPredicate[*egv1a1.HTTPRouteFilter]{
			predicate.TypedGenerationChangedPredicate[*egv1a1.HTTPRouteFilter]{},
			predicate.NewTypedPredicateFuncs(func(be *egv1a1.HTTPRouteFilter) bool {
				return r.validateHTTPRouteFilterForReconcile(be)
			}),
		}
		if r.namespaceLabel != nil {
			httpRouteFilter = append(httpRouteFilter, predicate.NewTypedPredicateFuncs(func(be *egv1a1.HTTPRouteFilter) bool {
				return r.hasMatchingNamespaceLabels(be)
			}))
		}
		if err := c.Watch(
			source.Kind(mgr.GetCache(), &egv1a1.HTTPRouteFilter{},
				handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, be *egv1a1.HTTPRouteFilter) []reconcile.Request {
					return r.enqueueClass(ctx, be)
				}),
				httpRouteFilter...)); err != nil {
			return err
		}

		if err := addRouteFilterIndexers(ctx, mgr); err != nil {
			return err
		}
	}
	return nil
}

func (r *gatewayAPIReconciler) enqueueClass(_ context.Context, _ client.Object) []reconcile.Request {
	return []reconcile.Request{{NamespacedName: types.NamespacedName{
		Name: string(r.classController),
	}}}
}

// processGatewayParamsRef processes the infrastructure.parametersRef of the provided Gateway.
func (r *gatewayAPIReconciler) processGatewayParamsRef(ctx context.Context, gtw *gwapiv1.Gateway, resourceMap *resourceMappings, resourceTree *resource.Resources) error {
	if gtw == nil || gtw.Spec.Infrastructure == nil || gtw.Spec.Infrastructure.ParametersRef == nil {
		return nil
	}

	if resourceTree.EnvoyProxyForGatewayClass != nil && resourceTree.EnvoyProxyForGatewayClass.Spec.MergeGateways != nil && *resourceTree.EnvoyProxyForGatewayClass.Spec.MergeGateways {
		return fmt.Errorf("infrastructure.parametersref must be nil when MergeGateways feature is enabled by gatewayclass")
	}

	ref := gtw.Spec.Infrastructure.ParametersRef
	if string(ref.Group) != egv1a1.GroupVersion.Group ||
		ref.Kind != egv1a1.KindEnvoyProxy ||
		len(ref.Name) == 0 {
		return fmt.Errorf("unsupported parametersRef for gateway %s/%s", gtw.Namespace, gtw.Name)
	}

	ep := new(egv1a1.EnvoyProxy)
	if err := r.client.Get(ctx, types.NamespacedName{Namespace: gtw.Namespace, Name: ref.Name}, ep); err != nil {
		return fmt.Errorf("failed to find envoyproxy %s/%s: %w", gtw.Namespace, ref.Name, err)
	}

	if err := r.processEnvoyProxy(ep, resourceMap); err != nil {
		return err
	}

	if ep.Spec.BackendTLS != nil && ep.Spec.BackendTLS.ClientCertificateRef != nil {
		certRef := ep.Spec.BackendTLS.ClientCertificateRef
		if refsSecret(certRef) {
			if err := r.processSecretRef(
				ctx,
				resourceMap,
				resourceTree,
				resource.KindGateway,
				gtw.Namespace,
				gtw.Name,
				*certRef); err != nil {
				r.log.Error(err,
					"failed to process TLS SecretRef for gateway",
					"gateway", utils.NamespacedName(gtw).String(), "secretRef", certRef)
			}
		}
	}

	resourceTree.EnvoyProxiesForGateways = append(resourceTree.EnvoyProxiesForGateways, ep)
	return nil
}

// processGatewayClassParamsRef processes the parametersRef of the provided GatewayClass.
func (r *gatewayAPIReconciler) processGatewayClassParamsRef(ctx context.Context, gc *gwapiv1.GatewayClass, resourceMap *resourceMappings, resourceTree *resource.Resources) error {
	if !refsEnvoyProxy(gc) {
		return fmt.Errorf("unsupported parametersRef for gatewayclass %s", gc.Name)
	}

	ep := new(egv1a1.EnvoyProxy)
	if err := r.client.Get(ctx, types.NamespacedName{Namespace: string(*gc.Spec.ParametersRef.Namespace), Name: gc.Spec.ParametersRef.Name}, ep); err != nil {
		if kerrors.IsNotFound(err) {
			return fmt.Errorf("envoyproxy referenced by gatewayclass is not found: %w", err)
		}
		return fmt.Errorf("failed to find envoyproxy %s/%s: %w", r.namespace, gc.Spec.ParametersRef.Name, err)
	}

	// Check for incompatible configuration: both MergeGateways and GatewayNamespaceMode enabled
	if r.gatewayNamespaceMode && ep.Spec.MergeGateways != nil && *ep.Spec.MergeGateways {
		return fmt.Errorf("using Merged Gateways with Gateway Namespace Mode is not supported.")
	}

	if err := r.processEnvoyProxy(ep, resourceMap); err != nil {
		return err
	}
	resourceTree.EnvoyProxyForGatewayClass = ep
	return nil
}

// processEnvoyProxy processes the parametersRef of the provided GatewayClass.
func (r *gatewayAPIReconciler) processEnvoyProxy(ep *egv1a1.EnvoyProxy, resourceMap *resourceMappings) error {
	key := utils.NamespacedName(ep).String()
	if resourceMap.allAssociatedEnvoyProxies.Has(key) {
		r.log.Info("current EnvoyProxy has been processed already", "namespace", ep.Namespace, "name", ep.Name)
		return nil
	}

	r.log.Info("processing EnvoyProxy", "namespace", ep.Namespace, "name", ep.Name)

	if err := validation.ValidateEnvoyProxy(ep); err != nil {
		return fmt.Errorf("invalid EnvoyProxy: %w", err)
	}
	if err := bootstrap.Validate(ep.Spec.Bootstrap); err != nil {
		return fmt.Errorf("invalid EnvoyProxy: %w", err)
	}

	if ep.Spec.Telemetry != nil {
		var backendRefs []egv1a1.BackendRef
		telemetry := ep.Spec.Telemetry

		if telemetry.AccessLog != nil {
			for _, setting := range telemetry.AccessLog.Settings {
				for _, sink := range setting.Sinks {
					if sink.OpenTelemetry != nil {
						backendRefs = append(backendRefs, sink.OpenTelemetry.BackendRefs...)
					}
					if sink.ALS != nil {
						backendRefs = append(backendRefs, sink.ALS.BackendRefs...)
					}
				}
			}
		}

		if telemetry.Metrics != nil {
			for _, sink := range telemetry.Metrics.Sinks {
				if sink.OpenTelemetry != nil {
					backendRefs = append(backendRefs, sink.OpenTelemetry.BackendRefs...)
				}
			}
		}

		if telemetry.Tracing != nil {
			backendRefs = append(backendRefs, telemetry.Tracing.Provider.BackendRefs...)
		}

		for _, backendRef := range backendRefs {
			backendNamespace := gatewayapi.NamespaceDerefOr(backendRef.Namespace, ep.Namespace)
			resourceMap.allAssociatedBackendRefs.Insert(gwapiv1.BackendObjectReference{
				Group:     backendRef.Group,
				Kind:      backendRef.Kind,
				Namespace: gatewayapi.NamespacePtr(backendNamespace),
				Name:      backendRef.Name,
			})
		}
	}

	resourceMap.allAssociatedEnvoyProxies.Insert(key)
	return nil
}

// crdExists checks for the existence of the CRD in k8s APIServer before watching it
func (r *gatewayAPIReconciler) crdExists(mgr manager.Manager, kind, groupVersion string) bool {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(mgr.GetConfig())
	if err != nil {
		r.log.Error(err, "failed to create discovery client")
	}
	apiResourceList, err := discoveryClient.ServerPreferredResources()
	if err != nil {
		r.log.Error(err, "failed to get API resource list")
	}
	found := false
	for _, list := range apiResourceList {
		for _, res := range list.APIResources {
			if list.GroupVersion == groupVersion && res.Kind == kind {
				found = true
				break
			}
		}
	}

	return found
}

func (r *gatewayAPIReconciler) processBackendTLSPolicyRefs(
	ctx context.Context,
	resourceTree *resource.Resources,
	resourceMap *resourceMappings,
) {
	for _, policy := range resourceTree.BackendTLSPolicies {
		tls := policy.Spec.Validation

		if tls.CACertificateRefs != nil {
			for _, caCertRef := range tls.CACertificateRefs {
				// if kind is not Secret or ConfigMap, we skip early to avoid further calculation overhead
				if string(caCertRef.Kind) == resource.KindConfigMap ||
					string(caCertRef.Kind) == resource.KindSecret {

					var err error
					caRefNew := gwapiv1.SecretObjectReference{
						Group:     gatewayapi.GroupPtr(string(caCertRef.Group)),
						Kind:      gatewayapi.KindPtr(string(caCertRef.Kind)),
						Name:      caCertRef.Name,
						Namespace: gatewayapi.NamespacePtr(policy.Namespace),
					}
					switch string(caCertRef.Kind) {
					case resource.KindConfigMap:
						err = r.processConfigMapRef(
							ctx,
							resourceMap,
							resourceTree,
							resource.KindBackendTLSPolicy,
							policy.Namespace,
							policy.Name,
							caRefNew)

					case resource.KindSecret:
						err = r.processSecretRef(
							ctx,
							resourceMap,
							resourceTree,
							resource.KindBackendTLSPolicy,
							policy.Namespace,
							policy.Name,
							caRefNew)
					}
					if err != nil {
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
	ctx context.Context, resourceTree *resource.Resources, resourceMap *resourceMappings,
) error {
	envoyExtensionPolicies := egv1a1.EnvoyExtensionPolicyList{}
	if err := r.client.List(ctx, &envoyExtensionPolicies); err != nil {
		return fmt.Errorf("error listing EnvoyExtensionPolicies: %w", err)
	}

	for _, policy := range envoyExtensionPolicies.Items {
		envoyExtensionPolicy := policy //nolint:copyloopvar
		// Discard Status to reduce memory consumption in watchable
		// It will be recomputed by the gateway-api layer
		envoyExtensionPolicy.Status = gwapiv1a2.PolicyStatus{}
		if !resourceMap.allAssociatedEnvoyExtensionPolicies.Has(utils.NamespacedName(&envoyExtensionPolicy).String()) {
			resourceMap.allAssociatedEnvoyExtensionPolicies.Insert(utils.NamespacedName(&envoyExtensionPolicy).String())
			resourceTree.EnvoyExtensionPolicies = append(resourceTree.EnvoyExtensionPolicies, &envoyExtensionPolicy)
		}
	}

	// Add the referenced Resources in EnvoyExtensionPolicies to the resourceTree
	r.processEnvoyExtensionPolicyObjectRefs(ctx, resourceTree, resourceMap)

	return nil
}

// processExtensionServerPolicies adds directly attached policies intended for the extension server
func (r *gatewayAPIReconciler) processExtensionServerPolicies(
	ctx context.Context, resourceTree *resource.Resources,
) error {
	for _, gvk := range r.extServerPolicies {
		polList := unstructured.UnstructuredList{}
		polList.SetAPIVersion(gvk.GroupVersion().String())
		polList.SetKind(gvk.Kind)

		if err := r.client.List(ctx, &polList); err != nil {
			return fmt.Errorf("error listing extension server policy %s: %w", gvk, err)
		}

		for _, policy := range polList.Items {
			policySpec, found := policy.Object["spec"].(map[string]any)
			if !found {
				return fmt.Errorf("no spec found in %s.%s %s", policy.GetAPIVersion(), policy.GetKind(), policy.GetName())
			}
			_, foundTargetRef := policySpec["targetRef"]
			_, foundTargetRefs := policySpec["targetRefs"]
			if !foundTargetRef && !foundTargetRefs {
				return fmt.Errorf("not a policy object - no targetRef or targetRefs found in %s.%s %s",
					policy.GetAPIVersion(), policy.GetKind(), policy.GetName())
			}

			delete(policy.Object, "status")
			resourceTree.ExtensionServerPolicies = append(resourceTree.ExtensionServerPolicies, policy)
		}
	}

	return nil
}

// processEnvoyExtensionPolicyObjectRefs adds the referenced resources in EnvoyExtensionPolicies
// to the resourceTree
// - BackendRefs for ExtProcs
// - SecretRefs for Wasms
// - ValueRefs for Luas
func (r *gatewayAPIReconciler) processEnvoyExtensionPolicyObjectRefs(
	ctx context.Context, resourceTree *resource.Resources, resourceMap *resourceMappings,
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
				if err := r.processBackendRef(
					ctx,
					resourceMap,
					resourceTree,
					resource.KindEnvoyExtensionPolicy,
					policy.Namespace,
					policy.Name,
					br.BackendObjectReference); err != nil {
					r.log.Error(err,
						"failed to process ExtProc BackendRef for EnvoyExtensionPolicy",
						"policy", policy, "backendRef", br.BackendObjectReference)
				}
			}
		}

		// Add the referenced SecretRefs in EnvoyExtensionPolicies to the resourceTree
		for _, wasm := range policy.Spec.Wasm {
			if wasm.Code.Image != nil && wasm.Code.Image.PullSecretRef != nil {
				if err := r.processSecretRef(
					ctx,
					resourceMap,
					resourceTree,
					resource.KindSecurityPolicy,
					policy.Namespace,
					policy.Name,
					*wasm.Code.Image.PullSecretRef); err != nil {
					r.log.Error(err,
						"failed to process Wasm Image PullSecretRef for EnvoyExtensionPolicy",
						"policy", policy, "secretRef", wasm.Code.Image.PullSecretRef)
				}
			}
		}

		// Add referenced ConfigMaps in Lua EnvoyExtensionPolicies to the resource tree
		for _, lua := range policy.Spec.Lua {
			if lua.Type == egv1a1.LuaValueTypeValueRef {
				if lua.ValueRef != nil && string(lua.ValueRef.Kind) == resource.KindConfigMap {
					configMap := new(corev1.ConfigMap)
					err := r.client.Get(ctx,
						types.NamespacedName{Namespace: policy.Namespace, Name: string(lua.ValueRef.Name)},
						configMap,
					)
					if err != nil {
						r.log.Error(err,
							"failed to process Lua ValueRef for EnvoyExtensionPolicy",
							"policy", policy, "ValueRef", lua.ValueRef.Name)
					}

					resourceMap.allAssociatedNamespaces.Insert(policy.Namespace)
					if !resourceMap.allAssociatedConfigMaps.Has(utils.NamespacedName(configMap).String()) {
						resourceMap.allAssociatedConfigMaps.Insert(utils.NamespacedName(configMap).String())
						resourceTree.ConfigMaps = append(resourceTree.ConfigMaps, configMap)
						r.log.Info("processing ConfigMap", "namespace", policy.Namespace, "name", string(lua.ValueRef.Name))
					}
				}
			}
		}
	}
}
