// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/sets"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/utils"
)

type resourceMappings struct {
	// Map for storing Gateways' NamespacedNames.
	allAssociatedGateways sets.Set[string]
	// Map for storing ReferenceGrants' NamespacedNames.
	allAssociatedReferenceGrants sets.Set[string]
	// Map for storing ServiceImports' NamespacedNames.
	allAssociatedServiceImports sets.Set[string]
	// Map for storing EndpointSlices' NamespacedNames.
	allAssociatedEndpointSlices sets.Set[string]
	// Map for storing Secrets' NamespacedNames.
	allAssociatedSecrets sets.Set[string]
	// Map for storing ConfigMaps' NamespacedNames.
	allAssociatedConfigMaps sets.Set[string]
	// Map for storing namespaces for Route, Service and Gateway objects.
	allAssociatedNamespaces sets.Set[string]
	// Map for storing EnvoyProxies' NamespacedNames attaching to Gateway or GatewayClass.
	allAssociatedEnvoyProxies sets.Set[string]
	// Map for storing EnvoyPatchPolicies' NamespacedNames attaching to Gateway.
	allAssociatedEnvoyPatchPolicies sets.Set[string]
	// Map for storing TLSRoutes' NamespacedNames attaching to various Gateway objects.
	allAssociatedTLSRoutes sets.Set[string]
	// Map for storing HTTPRoutes' NamespacedNames attaching to various Gateway objects.
	allAssociatedHTTPRoutes sets.Set[string]
	// Map for storing GRPCRoutes' NamespacedNames attaching to various Gateway objects.
	allAssociatedGRPCRoutes sets.Set[string]
	// Map for storing TCPRoutes' NamespacedNames attaching to various Gateway objects.
	allAssociatedTCPRoutes sets.Set[string]
	// Map for storing UDPRoutes' NamespacedNames attaching to various Gateway objects.
	allAssociatedUDPRoutes sets.Set[string]
	// Map for storing backendRefs' BackendObjectReference referred by various Route objects.
	allAssociatedBackendRefs sets.Set[gwapiv1.BackendObjectReference]
	// Map for storing ClientTrafficPolicies' NamespacedNames referred by various Route objects.
	allAssociatedClientTrafficPolicies sets.Set[string]
	// Map for storing BackendTrafficPolicies' NamespacedNames referred by various Route objects.
	allAssociatedBackendTrafficPolicies sets.Set[string]
	// Map for storing SecurityPolicies' NamespacedNames referred by various Route objects.
	allAssociatedSecurityPolicies sets.Set[string]
	// Map for storing BackendTLSPolicies' NamespacedNames referred by various Backend objects.
	allAssociatedBackendTLSPolicies sets.Set[string]
	// Map for storing EnvoyExtensionPolicies' NamespacedNames attaching to various Gateway objects.
	allAssociatedEnvoyExtensionPolicies sets.Set[string]
	// extensionRefFilters is a map of filters managed by an extension.
	// The key is the namespaced name, group and kind of the filter and the value is the
	// unstructured form of the resource.
	extensionRefFilters map[utils.NamespacedNameWithGroupKind]unstructured.Unstructured
	// httpRouteFilters is a map of HTTPRouteFilters, where the key is the namespaced name,
	// group and kind of the HTTPFilter.
	httpRouteFilters map[utils.NamespacedNameWithGroupKind]*egv1a1.HTTPRouteFilter
}

func newResourceMapping() *resourceMappings {
	return &resourceMappings{
		allAssociatedGateways:               sets.New[string](),
		allAssociatedReferenceGrants:        sets.New[string](),
		allAssociatedServiceImports:         sets.New[string](),
		allAssociatedEndpointSlices:         sets.New[string](),
		allAssociatedSecrets:                sets.New[string](),
		allAssociatedConfigMaps:             sets.New[string](),
		allAssociatedNamespaces:             sets.New[string](),
		allAssociatedEnvoyProxies:           sets.New[string](),
		allAssociatedEnvoyPatchPolicies:     sets.New[string](),
		allAssociatedTLSRoutes:              sets.New[string](),
		allAssociatedHTTPRoutes:             sets.New[string](),
		allAssociatedGRPCRoutes:             sets.New[string](),
		allAssociatedTCPRoutes:              sets.New[string](),
		allAssociatedUDPRoutes:              sets.New[string](),
		allAssociatedBackendRefs:            sets.New[gwapiv1.BackendObjectReference](),
		allAssociatedClientTrafficPolicies:  sets.New[string](),
		allAssociatedBackendTrafficPolicies: sets.New[string](),
		allAssociatedSecurityPolicies:       sets.New[string](),
		allAssociatedBackendTLSPolicies:     sets.New[string](),
		allAssociatedEnvoyExtensionPolicies: sets.New[string](),
		extensionRefFilters:                 map[utils.NamespacedNameWithGroupKind]unstructured.Unstructured{},
		httpRouteFilters:                    map[utils.NamespacedNameWithGroupKind]*egv1a1.HTTPRouteFilter{},
	}
}
