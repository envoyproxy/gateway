// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"errors"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/utils"
)

// processTLSRoutes finds TLSRoutes corresponding to a gatewayNamespaceName, further checks for
// the backend references and pushes the TLSRoutes to the resourceTree.
func (r *gatewayAPIReconciler) processTLSRoutes(ctx context.Context, gatewayNamespaceName string,
	resourceMap *resourceMappings, resourceTree *gatewayapi.Resources) error {
	tlsRouteList := &gwapiv1a2.TLSRouteList{}
	if err := r.client.List(ctx, tlsRouteList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(gatewayTLSRouteIndex, gatewayNamespaceName),
	}); err != nil {
		r.log.Error(err, "unable to find associated TLSRoutes")
		return err
	}

	for _, tlsRoute := range tlsRouteList.Items {
		tlsRoute := tlsRoute
		if r.namespaceLabel != nil {
			if ok, err := r.checkObjectNamespaceLabels(&tlsRoute); err != nil {
				r.log.Error(err, "failed to check namespace labels for TLSRoute %s in namespace %s: %w", tlsRoute.GetName(), tlsRoute.GetNamespace())
				continue
			} else if !ok {
				continue
			}
		}
		r.log.Info("processing TLSRoute", "namespace", tlsRoute.Namespace, "name", tlsRoute.Name)

		for _, rule := range tlsRoute.Spec.Rules {
			for _, backendRef := range rule.BackendRefs {
				backendRef := backendRef
				ref := gatewayapi.UpgradeBackendRef(backendRef)
				if err := validateBackendRef(&ref); err != nil {
					r.log.Error(err, "invalid backendRef")
					continue
				}

				backendNamespace := gatewayapi.NamespaceDerefOrAlpha(backendRef.Namespace, tlsRoute.Namespace)
				resourceMap.allAssociatedBackendRefs[gwapiv1.BackendObjectReference{
					Group:     backendRef.BackendObjectReference.Group,
					Kind:      backendRef.BackendObjectReference.Kind,
					Namespace: gatewayapi.NamespacePtrV1Alpha2(backendNamespace),
					Name:      backendRef.Name,
				}] = struct{}{}

				if backendNamespace != tlsRoute.Namespace {
					from := ObjectKindNamespacedName{kind: gatewayapi.KindTLSRoute, namespace: tlsRoute.Namespace, name: tlsRoute.Name}
					to := ObjectKindNamespacedName{kind: gatewayapi.KindDerefOr(backendRef.Kind, gatewayapi.KindService), namespace: backendNamespace, name: string(backendRef.Name)}
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

		resourceMap.allAssociatedNamespaces[tlsRoute.Namespace] = struct{}{}
		// Discard Status to reduce memory consumption in watchable
		// It will be recomputed by the gateway-api layer
		tlsRoute.Status = gwapiv1a2.TLSRouteStatus{}
		resourceTree.TLSRoutes = append(resourceTree.TLSRoutes, &tlsRoute)
	}

	return nil
}

// processGRPCRoutes finds GRPCRoutes corresponding to a gatewayNamespaceName, further checks for
// the backend references and pushes the GRPCRoutes to the resourceTree.
func (r *gatewayAPIReconciler) processGRPCRoutes(ctx context.Context, gatewayNamespaceName string,
	resourceMap *resourceMappings, resourceTree *gatewayapi.Resources) error {
	grpcRouteList := &gwapiv1a2.GRPCRouteList{}

	if err := r.client.List(ctx, grpcRouteList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(gatewayGRPCRouteIndex, gatewayNamespaceName),
	}); err != nil {
		r.log.Error(err, "failed to list GRPCRoutes")
		return err
	}

	for _, grpcRoute := range grpcRouteList.Items {
		grpcRoute := grpcRoute
		if r.namespaceLabel != nil {
			if ok, err := r.checkObjectNamespaceLabels(&grpcRoute); err != nil {
				r.log.Error(err, "failed to check namespace labels for GRPCRoute %s in namespace %s: %w", grpcRoute.GetName(), grpcRoute.GetNamespace())
				continue
			} else if !ok {
				continue
			}
		}
		r.log.Info("processing GRPCRoute", "namespace", grpcRoute.Namespace, "name", grpcRoute.Name)

		for _, rule := range grpcRoute.Spec.Rules {
			for _, backendRef := range rule.BackendRefs {
				backendRef := backendRef
				if err := validateBackendRef(&backendRef.BackendRef); err != nil {
					r.log.Error(err, "invalid backendRef")
					continue
				}

				backendNamespace := gatewayapi.NamespaceDerefOr(backendRef.Namespace, grpcRoute.Namespace)
				resourceMap.allAssociatedBackendRefs[gwapiv1.BackendObjectReference{
					Group:     backendRef.BackendObjectReference.Group,
					Kind:      backendRef.BackendObjectReference.Kind,
					Namespace: gatewayapi.NamespacePtrV1Alpha2(backendNamespace),
					Name:      backendRef.Name,
				}] = struct{}{}

				if backendNamespace != grpcRoute.Namespace {
					from := ObjectKindNamespacedName{
						kind:      gatewayapi.KindGRPCRoute,
						namespace: grpcRoute.Namespace,
						name:      grpcRoute.Name,
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

			for i := range rule.Filters {
				filter := rule.Filters[i]
				var extGKs []schema.GroupKind
				for _, gvk := range r.extGVKs {
					extGKs = append(extGKs, gvk.GroupKind())
				}
				if err := gatewayapi.ValidateGRPCRouteFilter(&filter, extGKs...); err != nil {
					r.log.Error(err, "bypassing filter rule", "index", i)
					continue
				}
				if filter.Type == gwapiv1a2.GRPCRouteFilterExtensionRef {
					// NOTE: filters must be in the same namespace as the GRPCRoute
					// Check if it's a Kind managed by an extension and add to resourceTree
					key := types.NamespacedName{
						Namespace: grpcRoute.Namespace,
						Name:      string(filter.ExtensionRef.Name),
					}
					extRefFilter, ok := resourceMap.extensionRefFilters[key]
					if !ok {
						r.log.Error(
							errors.New("filter not found; bypassing rule"),
							"Filter not found; bypassing rule",
							"name",
							filter.ExtensionRef.Name, "index", i)
						continue
					}

					resourceTree.ExtensionRefFilters = append(resourceTree.ExtensionRefFilters, extRefFilter)

				}
			}
		}

		resourceMap.allAssociatedNamespaces[grpcRoute.Namespace] = struct{}{}
		// Discard Status to reduce memory consumption in watchable
		// It will be recomputed by the gateway-api layer
		grpcRoute.Status = gwapiv1a2.GRPCRouteStatus{}
		resourceTree.GRPCRoutes = append(resourceTree.GRPCRoutes, &grpcRoute)
	}

	return nil
}

// processHTTPRoutes finds HTTPRoutes corresponding to a gatewayNamespaceName, further checks for
// the backend references and pushes the HTTPRoutes to the resourceTree.
func (r *gatewayAPIReconciler) processHTTPRoutes(ctx context.Context, gatewayNamespaceName string,
	resourceMap *resourceMappings, resourceTree *gatewayapi.Resources) error {
	httpRouteList := &gwapiv1.HTTPRouteList{}

	extensionRefFilters, err := r.getExtensionRefFilters(ctx)
	if err != nil {
		return err
	}
	for i := range extensionRefFilters {
		filter := extensionRefFilters[i]
		resourceMap.extensionRefFilters[utils.NamespacedName(&filter)] = filter
	}

	if err := r.client.List(ctx, httpRouteList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(gatewayHTTPRouteIndex, gatewayNamespaceName),
	}); err != nil {
		r.log.Error(err, "failed to list HTTPRoutes")
		return err
	}

	for _, httpRoute := range httpRouteList.Items {
		httpRoute := httpRoute
		if r.namespaceLabel != nil {
			if ok, err := r.checkObjectNamespaceLabels(&httpRoute); err != nil {
				r.log.Error(err, "failed to check namespace labels for HTTPRoute %s in namespace %s: %w", httpRoute.GetName(), httpRoute.GetNamespace())
				continue
			} else if !ok {
				continue
			}
		}
		r.log.Info("processing HTTPRoute", "namespace", httpRoute.Namespace, "name", httpRoute.Name)

		for _, rule := range httpRoute.Spec.Rules {
			for _, backendRef := range rule.BackendRefs {
				backendRef := backendRef
				if err := validateBackendRef(&backendRef.BackendRef); err != nil {
					r.log.Error(err, "invalid backendRef")
					continue
				}

				backendNamespace := gatewayapi.NamespaceDerefOr(backendRef.Namespace, httpRoute.Namespace)
				resourceMap.allAssociatedBackendRefs[gwapiv1.BackendObjectReference{
					Group:     backendRef.BackendObjectReference.Group,
					Kind:      backendRef.BackendObjectReference.Kind,
					Namespace: gatewayapi.NamespacePtrV1Alpha2(backendNamespace),
					Name:      backendRef.Name,
				}] = struct{}{}

				if backendNamespace != httpRoute.Namespace {
					from := ObjectKindNamespacedName{
						kind:      gatewayapi.KindHTTPRoute,
						namespace: httpRoute.Namespace,
						name:      httpRoute.Name,
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

			for i := range rule.Filters {
				filter := rule.Filters[i]
				var extGKs []schema.GroupKind
				for _, gvk := range r.extGVKs {
					extGKs = append(extGKs, gvk.GroupKind())
				}
				if err := gatewayapi.ValidateHTTPRouteFilter(&filter, extGKs...); err != nil {
					r.log.Error(err, "bypassing filter rule", "index", i)
					continue
				}

				// Load in the backendRefs from any requestMirrorFilters on the HTTPRoute
				if filter.Type == gwapiv1.HTTPRouteFilterRequestMirror {
					// Make sure the config actually exists
					mirrorFilter := filter.RequestMirror
					if mirrorFilter == nil {
						r.log.Error(errors.New("invalid requestMirror filter"), "bypassing filter rule", "index", i)
						continue
					}

					mirrorBackendObj := mirrorFilter.BackendRef
					// Wrap the filter's BackendObjectReference into a BackendRef so we can use existing tooling to check it
					weight := int32(1)
					mirrorBackendRef := gwapiv1.BackendRef{
						BackendObjectReference: mirrorBackendObj,
						Weight:                 &weight,
					}

					if err := validateBackendRef(&mirrorBackendRef); err != nil {
						r.log.Error(err, "invalid backendRef")
						continue
					}

					backendNamespace := gatewayapi.NamespaceDerefOr(mirrorBackendRef.Namespace, httpRoute.Namespace)
					resourceMap.allAssociatedBackendRefs[gwapiv1.BackendObjectReference{
						Group:     mirrorBackendRef.BackendObjectReference.Group,
						Kind:      mirrorBackendRef.BackendObjectReference.Kind,
						Namespace: gatewayapi.NamespacePtrV1Alpha2(backendNamespace),
						Name:      mirrorBackendRef.Name,
					}] = struct{}{}

					if backendNamespace != httpRoute.Namespace {
						from := ObjectKindNamespacedName{
							kind:      gatewayapi.KindHTTPRoute,
							namespace: httpRoute.Namespace,
							name:      httpRoute.Name,
						}
						to := ObjectKindNamespacedName{
							kind:      gatewayapi.KindDerefOr(mirrorBackendRef.Kind, gatewayapi.KindService),
							namespace: backendNamespace,
							name:      string(mirrorBackendRef.Name),
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
				} else if filter.Type == gwapiv1.HTTPRouteFilterExtensionRef {
					// NOTE: filters must be in the same namespace as the HTTPRoute
					// Check if it's a Kind managed by an extension and add to resourceTree
					key := types.NamespacedName{
						Namespace: httpRoute.Namespace,
						Name:      string(filter.ExtensionRef.Name),
					}
					extRefFilter, ok := resourceMap.extensionRefFilters[key]
					if !ok {
						r.log.Error(
							errors.New("filter not found; bypassing rule"),
							"Filter not found; bypassing rule",
							"name", filter.ExtensionRef.Name,
							"index", i)
						continue
					}

					resourceTree.ExtensionRefFilters = append(resourceTree.ExtensionRefFilters, extRefFilter)
				}
			}
		}

		resourceMap.allAssociatedNamespaces[httpRoute.Namespace] = struct{}{}
		// Discard Status to reduce memory consumption in watchable
		// It will be recomputed by the gateway-api layer
		httpRoute.Status = gwapiv1.HTTPRouteStatus{}
		resourceTree.HTTPRoutes = append(resourceTree.HTTPRoutes, &httpRoute)
	}

	return nil
}

// processTCPRoutes finds TCPRoutes corresponding to a gatewayNamespaceName, further checks for
// the backend references and pushes the TCPRoutes to the resourceTree.
func (r *gatewayAPIReconciler) processTCPRoutes(ctx context.Context, gatewayNamespaceName string,
	resourceMap *resourceMappings, resourceTree *gatewayapi.Resources) error {
	tcpRouteList := &gwapiv1a2.TCPRouteList{}
	if err := r.client.List(ctx, tcpRouteList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(gatewayTCPRouteIndex, gatewayNamespaceName),
	}); err != nil {
		r.log.Error(err, "unable to find associated UDPRoutes")
		return err
	}

	for _, tcpRoute := range tcpRouteList.Items {
		tcpRoute := tcpRoute
		if r.namespaceLabel != nil {
			if ok, err := r.checkObjectNamespaceLabels(&tcpRoute); err != nil {
				r.log.Error(err, "failed to check namespace labels for TCPRoute %s in namespace %s: %w", tcpRoute.GetName(), tcpRoute.GetNamespace())
				continue
			} else if !ok {
				continue
			}
		}
		r.log.Info("processing TCPRoute", "namespace", tcpRoute.Namespace, "name", tcpRoute.Name)

		for _, rule := range tcpRoute.Spec.Rules {
			for _, backendRef := range rule.BackendRefs {
				backendRef := backendRef
				ref := gatewayapi.UpgradeBackendRef(backendRef)
				if err := validateBackendRef(&ref); err != nil {
					r.log.Error(err, "invalid backendRef")
					continue
				}

				backendNamespace := gatewayapi.NamespaceDerefOrAlpha(backendRef.Namespace, tcpRoute.Namespace)
				resourceMap.allAssociatedBackendRefs[gwapiv1.BackendObjectReference{
					Group:     backendRef.BackendObjectReference.Group,
					Kind:      backendRef.BackendObjectReference.Kind,
					Namespace: gatewayapi.NamespacePtrV1Alpha2(backendNamespace),
					Name:      backendRef.Name,
				}] = struct{}{}

				if backendNamespace != tcpRoute.Namespace {
					from := ObjectKindNamespacedName{kind: gatewayapi.KindTCPRoute, namespace: tcpRoute.Namespace, name: tcpRoute.Name}
					to := ObjectKindNamespacedName{kind: gatewayapi.KindDerefOr(backendRef.Kind, gatewayapi.KindService), namespace: backendNamespace, name: string(backendRef.Name)}
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

		resourceMap.allAssociatedNamespaces[tcpRoute.Namespace] = struct{}{}
		// Discard Status to reduce memory consumption in watchable
		// It will be recomputed by the gateway-api layer
		tcpRoute.Status = gwapiv1a2.TCPRouteStatus{}
		resourceTree.TCPRoutes = append(resourceTree.TCPRoutes, &tcpRoute)
	}

	return nil
}

// processUDPRoutes finds UDPRoutes corresponding to a gatewayNamespaceName, further checks for
// the backend references and pushes the UDPRoutes to the resourceTree.
func (r *gatewayAPIReconciler) processUDPRoutes(ctx context.Context, gatewayNamespaceName string,
	resourceMap *resourceMappings, resourceTree *gatewayapi.Resources) error {
	udpRouteList := &gwapiv1a2.UDPRouteList{}
	if err := r.client.List(ctx, udpRouteList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(gatewayUDPRouteIndex, gatewayNamespaceName),
	}); err != nil {
		r.log.Error(err, "unable to find associated UDPRoutes")
		return err
	}

	for _, udpRoute := range udpRouteList.Items {
		udpRoute := udpRoute
		if r.namespaceLabel != nil {
			if ok, err := r.checkObjectNamespaceLabels(&udpRoute); err != nil {
				r.log.Error(err, "failed to check namespace labels for UDPRoute %s in namespace %s: %w", udpRoute.GetName(), udpRoute.GetNamespace())
				continue
			} else if !ok {
				continue
			}
		}
		r.log.Info("processing UDPRoute", "namespace", udpRoute.Namespace, "name", udpRoute.Name)

		for _, rule := range udpRoute.Spec.Rules {
			for _, backendRef := range rule.BackendRefs {
				backendRef := backendRef
				ref := gatewayapi.UpgradeBackendRef(backendRef)
				if err := validateBackendRef(&ref); err != nil {
					r.log.Error(err, "invalid backendRef")
					continue
				}

				backendNamespace := gatewayapi.NamespaceDerefOrAlpha(backendRef.Namespace, udpRoute.Namespace)
				resourceMap.allAssociatedBackendRefs[gwapiv1.BackendObjectReference{
					Group:     backendRef.BackendObjectReference.Group,
					Kind:      backendRef.BackendObjectReference.Kind,
					Namespace: gatewayapi.NamespacePtrV1Alpha2(backendNamespace),
					Name:      backendRef.Name,
				}] = struct{}{}

				if backendNamespace != udpRoute.Namespace {
					from := ObjectKindNamespacedName{kind: gatewayapi.KindUDPRoute, namespace: udpRoute.Namespace, name: udpRoute.Name}
					to := ObjectKindNamespacedName{kind: gatewayapi.KindDerefOr(backendRef.Kind, gatewayapi.KindService), namespace: backendNamespace, name: string(backendRef.Name)}
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

		resourceMap.allAssociatedNamespaces[udpRoute.Namespace] = struct{}{}
		// Discard Status to reduce memory consumption in watchable
		// It will be recomputed by the gateway-api layer
		udpRoute.Status = gwapiv1a2.UDPRouteStatus{}
		resourceTree.UDPRoutes = append(resourceTree.UDPRoutes, &udpRoute)
	}

	return nil
}
