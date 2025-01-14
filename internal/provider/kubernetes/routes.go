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

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/utils"
)

// processTLSRoutes finds TLSRoutes corresponding to a gatewayNamespaceName, further checks for
// the backend references and pushes the TLSRoutes to the resourceTree.
func (r *gatewayAPIReconciler) processTLSRoutes(ctx context.Context, gatewayNamespaceName string,
	resourceMap *resourceMappings, resourceTree *resource.Resources,
) error {
	tlsRouteList := &gwapiv1a2.TLSRouteList{}
	if err := r.client.List(ctx, tlsRouteList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(gatewayTLSRouteIndex, gatewayNamespaceName),
	}); err != nil {
		r.log.Error(err, "unable to find associated TLSRoutes")
		return err
	}

	for _, tlsRoute := range tlsRouteList.Items {
		tlsRoute := tlsRoute //nolint:copyloopvar
		if r.namespaceLabel != nil {
			if ok, err := r.checkObjectNamespaceLabels(&tlsRoute); err != nil {
				r.log.Error(err, "failed to check namespace labels for TLSRoute %s in namespace %s: %w", tlsRoute.GetName(), tlsRoute.GetNamespace())
				continue
			} else if !ok {
				continue
			}
		}

		key := utils.NamespacedName(&tlsRoute).String()
		if resourceMap.allAssociatedTLSRoutes.Has(key) {
			r.log.Info("current TLSRoute has been processed already", "namespace", tlsRoute.Namespace, "name", tlsRoute.Name)
			continue
		}

		r.log.Info("processing TLSRoute", "namespace", tlsRoute.Namespace, "name", tlsRoute.Name)

		for _, rule := range tlsRoute.Spec.Rules {
			for _, backendRef := range rule.BackendRefs {
				if err := validateBackendRef(&backendRef); err != nil {
					r.log.Error(err, "invalid backendRef")
					continue
				}

				backendNamespace := gatewayapi.NamespaceDerefOr(backendRef.Namespace, tlsRoute.Namespace)
				resourceMap.allAssociatedBackendRefs.Insert(gwapiv1.BackendObjectReference{
					Group:     backendRef.BackendObjectReference.Group,
					Kind:      backendRef.BackendObjectReference.Kind,
					Namespace: gatewayapi.NamespacePtr(backendNamespace),
					Name:      backendRef.Name,
				})

				if backendNamespace != tlsRoute.Namespace {
					from := ObjectKindNamespacedName{kind: resource.KindTLSRoute, namespace: tlsRoute.Namespace, name: tlsRoute.Name}
					to := ObjectKindNamespacedName{kind: gatewayapi.KindDerefOr(backendRef.Kind, resource.KindService), namespace: backendNamespace, name: string(backendRef.Name)}
					refGrant, err := r.findReferenceGrant(ctx, from, to)
					switch {
					case err != nil:
						r.log.Error(err, "failed to find ReferenceGrant")
					case refGrant == nil:
						r.log.Info("no matching ReferenceGrants found", "from", from.kind,
							"from namespace", from.namespace, "target", to.kind, "target namespace", to.namespace)
					default:
						refGrantNamespacedName := utils.NamespacedName(refGrant).String()
						if !resourceMap.allAssociatedReferenceGrants.Has(refGrantNamespacedName) {
							resourceMap.allAssociatedReferenceGrants.Insert(refGrantNamespacedName)
							resourceTree.ReferenceGrants = append(resourceTree.ReferenceGrants, refGrant)
							r.log.Info("added ReferenceGrant to resource map", "namespace", refGrant.Namespace,
								"name", refGrant.Name)
						}
					}
				}
			}
		}

		resourceMap.allAssociatedNamespaces.Insert(tlsRoute.Namespace)
		resourceMap.allAssociatedTLSRoutes.Insert(key)
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
	resourceMap *resourceMappings, resourceTree *resource.Resources,
) error {
	grpcRouteList := &gwapiv1.GRPCRouteList{}

	if err := r.client.List(ctx, grpcRouteList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(gatewayGRPCRouteIndex, gatewayNamespaceName),
	}); err != nil {
		r.log.Error(err, "failed to list GRPCRoutes")
		return err
	}

	for _, grpcRoute := range grpcRouteList.Items {
		grpcRoute := grpcRoute //nolint:copyloopvar
		if r.namespaceLabel != nil {
			if ok, err := r.checkObjectNamespaceLabels(&grpcRoute); err != nil {
				r.log.Error(err, "failed to check namespace labels for GRPCRoute %s in namespace %s: %w", grpcRoute.GetName(), grpcRoute.GetNamespace())
				continue
			} else if !ok {
				continue
			}
		}

		key := utils.NamespacedName(&grpcRoute).String()
		if resourceMap.allAssociatedGRPCRoutes.Has(key) {
			r.log.Info("current GRPCRoute has been processed already", "namespace", grpcRoute.Namespace, "name", grpcRoute.Name)
			continue
		}

		r.log.Info("processing GRPCRoute", "namespace", grpcRoute.Namespace, "name", grpcRoute.Name)

		for _, rule := range grpcRoute.Spec.Rules {
			for _, backendRef := range rule.BackendRefs {
				if err := validateBackendRef(&backendRef.BackendRef); err != nil {
					r.log.Error(err, "invalid backendRef")
					continue
				}

				backendNamespace := gatewayapi.NamespaceDerefOr(backendRef.Namespace, grpcRoute.Namespace)
				resourceMap.allAssociatedBackendRefs.Insert(gwapiv1.BackendObjectReference{
					Group:     backendRef.BackendObjectReference.Group,
					Kind:      backendRef.BackendObjectReference.Kind,
					Namespace: gatewayapi.NamespacePtr(backendNamespace),
					Name:      backendRef.Name,
				})

				if backendNamespace != grpcRoute.Namespace {
					from := ObjectKindNamespacedName{
						kind:      resource.KindGRPCRoute,
						namespace: grpcRoute.Namespace,
						name:      grpcRoute.Name,
					}
					to := ObjectKindNamespacedName{
						kind:      gatewayapi.KindDerefOr(backendRef.Kind, resource.KindService),
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
						refGrantNamespacedName := utils.NamespacedName(refGrant).String()
						if !resourceMap.allAssociatedReferenceGrants.Has(refGrantNamespacedName) {
							resourceMap.allAssociatedReferenceGrants.Insert(refGrantNamespacedName)
							resourceTree.ReferenceGrants = append(resourceTree.ReferenceGrants, refGrant)
							r.log.Info("added ReferenceGrant to resource map", "namespace", refGrant.Namespace,
								"name", refGrant.Name)
						}
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
				if filter.Type == gwapiv1.GRPCRouteFilterExtensionRef {
					// NOTE: filters must be in the same namespace as the GRPCRoute
					// Check if it's a Kind managed by an extension and add to resourceTree
					key := utils.NamespacedNameWithGroupKind{
						NamespacedName: types.NamespacedName{
							Namespace: grpcRoute.Namespace,
							Name:      string(filter.ExtensionRef.Name),
						},
						GroupKind: schema.GroupKind{
							Group: string(filter.ExtensionRef.Group),
							Kind:  string(filter.ExtensionRef.Kind),
						},
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

		resourceMap.allAssociatedNamespaces.Insert(grpcRoute.Namespace)
		resourceMap.allAssociatedGRPCRoutes.Insert(key)
		// Discard Status to reduce memory consumption in watchable
		// It will be recomputed by the gateway-api layer
		grpcRoute.Status = gwapiv1.GRPCRouteStatus{}
		resourceTree.GRPCRoutes = append(resourceTree.GRPCRoutes, &grpcRoute)
	}

	return nil
}

// processHTTPRoutes finds HTTPRoutes corresponding to a gatewayNamespaceName, further checks for
// the backend references and pushes the HTTPRoutes to the resourceTree.
func (r *gatewayAPIReconciler) processHTTPRoutes(ctx context.Context, gatewayNamespaceName string,
	resourceMap *resourceMappings, resourceTree *resource.Resources,
) error {
	httpRouteList := &gwapiv1.HTTPRouteList{}

	extensionRefFilters, err := r.getExtensionRefFilters(ctx)
	if err != nil {
		return err
	}
	for i := range extensionRefFilters {
		filter := extensionRefFilters[i]
		resourceMap.extensionRefFilters[utils.GetNamespacedNameWithGroupKind(&filter)] = filter
	}

	if err := r.client.List(ctx, httpRouteList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(gatewayHTTPRouteIndex, gatewayNamespaceName),
	}); err != nil {
		r.log.Error(err, "failed to list HTTPRoutes")
		return err
	}

	for _, httpRoute := range httpRouteList.Items {
		httpRoute := httpRoute //nolint:copyloopvar
		if r.namespaceLabel != nil {
			if ok, err := r.checkObjectNamespaceLabels(&httpRoute); err != nil {
				r.log.Error(err, "failed to check namespace labels for HTTPRoute %s in namespace %s: %w", httpRoute.GetName(), httpRoute.GetNamespace())
				continue
			} else if !ok {
				continue
			}
		}

		key := utils.NamespacedName(&httpRoute).String()
		if resourceMap.allAssociatedHTTPRoutes.Has(key) {
			r.log.Info("current HTTPRoute has been processed already", "namespace", httpRoute.Namespace, "name", httpRoute.Name)
			continue
		}

		r.log.Info("processing HTTPRoute", "namespace", httpRoute.Namespace, "name", httpRoute.Name)

		for _, rule := range httpRoute.Spec.Rules {
			for _, backendRef := range rule.BackendRefs {
				if err := validateBackendRef(&backendRef.BackendRef); err != nil {
					r.log.Error(err, "invalid backendRef")
					continue
				}

				backendNamespace := gatewayapi.NamespaceDerefOr(backendRef.Namespace, httpRoute.Namespace)
				resourceMap.allAssociatedBackendRefs.Insert(gwapiv1.BackendObjectReference{
					Group:     backendRef.BackendObjectReference.Group,
					Kind:      backendRef.BackendObjectReference.Kind,
					Namespace: gatewayapi.NamespacePtr(backendNamespace),
					Name:      backendRef.Name,
				})

				if backendNamespace != httpRoute.Namespace {
					from := ObjectKindNamespacedName{
						kind:      resource.KindHTTPRoute,
						namespace: httpRoute.Namespace,
						name:      httpRoute.Name,
					}
					to := ObjectKindNamespacedName{
						kind:      gatewayapi.KindDerefOr(backendRef.Kind, resource.KindService),
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
						refGrantNamespacedName := utils.NamespacedName(refGrant).String()
						if !resourceMap.allAssociatedReferenceGrants.Has(refGrantNamespacedName) {
							resourceMap.allAssociatedReferenceGrants.Insert(refGrantNamespacedName)
							resourceTree.ReferenceGrants = append(resourceTree.ReferenceGrants, refGrant)
							r.log.Info("added ReferenceGrant to resource map", "namespace", refGrant.Namespace,
								"name", refGrant.Name)
						}
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
					resourceMap.allAssociatedBackendRefs.Insert(gwapiv1.BackendObjectReference{
						Group:     mirrorBackendRef.BackendObjectReference.Group,
						Kind:      mirrorBackendRef.BackendObjectReference.Kind,
						Namespace: gatewayapi.NamespacePtr(backendNamespace),
						Name:      mirrorBackendRef.Name,
					})

					if backendNamespace != httpRoute.Namespace {
						from := ObjectKindNamespacedName{
							kind:      resource.KindHTTPRoute,
							namespace: httpRoute.Namespace,
							name:      httpRoute.Name,
						}
						to := ObjectKindNamespacedName{
							kind:      gatewayapi.KindDerefOr(mirrorBackendRef.Kind, resource.KindService),
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
							refGrantNamespacedName := utils.NamespacedName(refGrant).String()
							if !resourceMap.allAssociatedReferenceGrants.Has(refGrantNamespacedName) {
								resourceMap.allAssociatedReferenceGrants.Insert(refGrantNamespacedName)
								resourceTree.ReferenceGrants = append(resourceTree.ReferenceGrants, refGrant)
								r.log.Info("added ReferenceGrant to resource map", "namespace", refGrant.Namespace,
									"name", refGrant.Name)
							}
						}
					}
				} else if filter.Type == gwapiv1.HTTPRouteFilterExtensionRef {
					// NOTE: filters must be in the same namespace as the HTTPRoute
					// Check if it's a Kind managed by an extension and add to resourceTree
					key := utils.NamespacedNameWithGroupKind{
						NamespacedName: types.NamespacedName{
							Namespace: httpRoute.Namespace,
							Name:      string(filter.ExtensionRef.Name),
						},
						GroupKind: schema.GroupKind{
							Group: string(filter.ExtensionRef.Group),
							Kind:  string(filter.ExtensionRef.Kind),
						},
					}

					switch string(filter.ExtensionRef.Kind) {
					case egv1a1.KindHTTPRouteFilter:
						if r.hrfCRDExists {
							httpFilter, err := r.getHTTPRouteFilter(ctx, key.Name, key.Namespace)
							if err != nil {
								r.log.Error(err, "HTTPRouteFilters not found; bypassing rule", "index", i)
								continue
							}
							if !resourceMap.allAssociatedHTTPRouteExtensionFilters.Has(key) {
								r.processRouteFilterConfigMapRef(ctx, httpFilter, resourceMap, resourceTree)
								resourceMap.allAssociatedHTTPRouteExtensionFilters.Insert(key)
								resourceTree.HTTPRouteFilters = append(resourceTree.HTTPRouteFilters, httpFilter)
							}
						}
					default:
						extRefFilter, ok := resourceMap.extensionRefFilters[key]
						if !ok {
							r.log.Error(
								errors.New("filter not found; bypassing rule"),
								"Filter not found; bypassing rule",
								"name", filter.ExtensionRef.Name,
								"index", i)
							continue
						}

						if !resourceMap.allAssociatedHTTPRouteExtensionFilters.Has(key) {
							resourceMap.allAssociatedHTTPRouteExtensionFilters.Insert(key)
							resourceTree.ExtensionRefFilters = append(resourceTree.ExtensionRefFilters, extRefFilter)
						}
					}
				}
			}
		}

		resourceMap.allAssociatedNamespaces.Insert(httpRoute.Namespace)
		resourceMap.allAssociatedHTTPRoutes.Insert(key)
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
	resourceMap *resourceMappings, resourceTree *resource.Resources,
) error {
	tcpRouteList := &gwapiv1a2.TCPRouteList{}
	if err := r.client.List(ctx, tcpRouteList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(gatewayTCPRouteIndex, gatewayNamespaceName),
	}); err != nil {
		r.log.Error(err, "unable to find associated UDPRoutes")
		return err
	}

	for _, tcpRoute := range tcpRouteList.Items {
		tcpRoute := tcpRoute //nolint:copyloopvar
		if r.namespaceLabel != nil {
			if ok, err := r.checkObjectNamespaceLabels(&tcpRoute); err != nil {
				r.log.Error(err, "failed to check namespace labels for TCPRoute %s in namespace %s: %w", tcpRoute.GetName(), tcpRoute.GetNamespace())
				continue
			} else if !ok {
				continue
			}
		}

		key := utils.NamespacedName(&tcpRoute).String()
		if resourceMap.allAssociatedTCPRoutes.Has(key) {
			r.log.Info("current TCPRoute has been processed already", "namespace", tcpRoute.Namespace, "name", tcpRoute.Name)
			continue
		}

		r.log.Info("processing TCPRoute", "namespace", tcpRoute.Namespace, "name", tcpRoute.Name)

		for _, rule := range tcpRoute.Spec.Rules {
			for _, backendRef := range rule.BackendRefs {
				if err := validateBackendRef(&backendRef); err != nil {
					r.log.Error(err, "invalid backendRef")
					continue
				}

				backendNamespace := gatewayapi.NamespaceDerefOr(backendRef.Namespace, tcpRoute.Namespace)
				resourceMap.allAssociatedBackendRefs.Insert(gwapiv1.BackendObjectReference{
					Group:     backendRef.BackendObjectReference.Group,
					Kind:      backendRef.BackendObjectReference.Kind,
					Namespace: gatewayapi.NamespacePtr(backendNamespace),
					Name:      backendRef.Name,
				})

				if backendNamespace != tcpRoute.Namespace {
					from := ObjectKindNamespacedName{kind: resource.KindTCPRoute, namespace: tcpRoute.Namespace, name: tcpRoute.Name}
					to := ObjectKindNamespacedName{kind: gatewayapi.KindDerefOr(backendRef.Kind, resource.KindService), namespace: backendNamespace, name: string(backendRef.Name)}
					refGrant, err := r.findReferenceGrant(ctx, from, to)
					switch {
					case err != nil:
						r.log.Error(err, "failed to find ReferenceGrant")
					case refGrant == nil:
						r.log.Info("no matching ReferenceGrants found", "from", from.kind,
							"from namespace", from.namespace, "target", to.kind, "target namespace", to.namespace)
					default:
						refGrantNamespacedName := utils.NamespacedName(refGrant).String()
						if !resourceMap.allAssociatedReferenceGrants.Has(refGrantNamespacedName) {
							resourceMap.allAssociatedReferenceGrants.Insert(refGrantNamespacedName)
							resourceTree.ReferenceGrants = append(resourceTree.ReferenceGrants, refGrant)
							r.log.Info("added ReferenceGrant to resource map", "namespace", refGrant.Namespace,
								"name", refGrant.Name)
						}
					}
				}
			}
		}

		resourceMap.allAssociatedNamespaces.Insert(tcpRoute.Namespace)
		resourceMap.allAssociatedTCPRoutes.Insert(key)
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
	resourceMap *resourceMappings, resourceTree *resource.Resources,
) error {
	udpRouteList := &gwapiv1a2.UDPRouteList{}
	if err := r.client.List(ctx, udpRouteList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(gatewayUDPRouteIndex, gatewayNamespaceName),
	}); err != nil {
		r.log.Error(err, "unable to find associated UDPRoutes")
		return err
	}

	for _, udpRoute := range udpRouteList.Items {
		udpRoute := udpRoute //nolint:copyloopvar
		if r.namespaceLabel != nil {
			if ok, err := r.checkObjectNamespaceLabels(&udpRoute); err != nil {
				r.log.Error(err, "failed to check namespace labels for UDPRoute %s in namespace %s: %w", udpRoute.GetName(), udpRoute.GetNamespace())
				continue
			} else if !ok {
				continue
			}
		}

		key := utils.NamespacedName(&udpRoute).String()
		if resourceMap.allAssociatedUDPRoutes.Has(key) {
			r.log.Info("current UDPRoute has been processed already", "namespace", udpRoute.Namespace, "name", udpRoute.Name)
			continue
		}

		r.log.Info("processing UDPRoute", "namespace", udpRoute.Namespace, "name", udpRoute.Name)

		for _, rule := range udpRoute.Spec.Rules {
			for _, backendRef := range rule.BackendRefs {
				if err := validateBackendRef(&backendRef); err != nil {
					r.log.Error(err, "invalid backendRef")
					continue
				}

				backendNamespace := gatewayapi.NamespaceDerefOr(backendRef.Namespace, udpRoute.Namespace)
				resourceMap.allAssociatedBackendRefs.Insert(gwapiv1.BackendObjectReference{
					Group:     backendRef.BackendObjectReference.Group,
					Kind:      backendRef.BackendObjectReference.Kind,
					Namespace: gatewayapi.NamespacePtr(backendNamespace),
					Name:      backendRef.Name,
				})

				if backendNamespace != udpRoute.Namespace {
					from := ObjectKindNamespacedName{kind: resource.KindUDPRoute, namespace: udpRoute.Namespace, name: udpRoute.Name}
					to := ObjectKindNamespacedName{kind: gatewayapi.KindDerefOr(backendRef.Kind, resource.KindService), namespace: backendNamespace, name: string(backendRef.Name)}
					refGrant, err := r.findReferenceGrant(ctx, from, to)
					switch {
					case err != nil:
						r.log.Error(err, "failed to find ReferenceGrant")
					case refGrant == nil:
						r.log.Info("no matching ReferenceGrants found", "from", from.kind,
							"from namespace", from.namespace, "target", to.kind, "target namespace", to.namespace)
					default:
						if !resourceMap.allAssociatedReferenceGrants.Has(utils.NamespacedName(refGrant).String()) {
							resourceMap.allAssociatedReferenceGrants.Insert(utils.NamespacedName(refGrant).String())
							resourceTree.ReferenceGrants = append(resourceTree.ReferenceGrants, refGrant)
							r.log.Info("added ReferenceGrant to resource map", "namespace", refGrant.Namespace,
								"name", refGrant.Name)
						}
					}
				}
			}
		}

		resourceMap.allAssociatedNamespaces.Insert(udpRoute.Namespace)
		resourceMap.allAssociatedUDPRoutes.Insert(key)
		// Discard Status to reduce memory consumption in watchable
		// It will be recomputed by the gateway-api layer
		udpRoute.Status = gwapiv1a2.UDPRouteStatus{}
		resourceTree.UDPRoutes = append(resourceTree.UDPRoutes, &udpRoute)
	}

	return nil
}
