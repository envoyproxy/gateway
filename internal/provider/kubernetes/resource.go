// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/sets"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/envoyproxy/gateway/internal/utils"
)

type resourceMappings struct {
	// Set for storing Gateways' NamespacedNames.
	allAssociatedGateways sets.Set[string]
	// Set for storing ReferenceGrants' NamespacedNames.
	allAssociatedReferenceGrants sets.Set[string]
	// Set for storing ServiceImports' NamespacedNames.
	allAssociatedServiceImports sets.Set[string]
	// Set for storing EndpointSlices' NamespacedNames.
	allAssociatedEndpointSlices sets.Set[string]
	// Set for storing Backends' NamespacedNames.
	allAssociatedBackends sets.Set[string]
	// Set for storing Secrets' NamespacedNames.
	allAssociatedSecrets sets.Set[string]
	// Set for storing ConfigMaps' NamespacedNames.
	allAssociatedConfigMaps sets.Set[string]
	// Set for storing namespaces for Route, Service and Gateway objects.
	allAssociatedNamespaces sets.Set[string]
	// Set for storing EnvoyProxies' NamespacedNames attaching to Gateway or GatewayClass.
	allAssociatedEnvoyProxies sets.Set[string]
	// Set for storing EnvoyPatchPolicies' NamespacedNames attaching to Gateway.
	allAssociatedEnvoyPatchPolicies sets.Set[string]
	// Set for storing TLSRoutes' NamespacedNames attaching to various Gateway objects.
	allAssociatedTLSRoutes sets.Set[string]
	// Set for storing HTTPRoutes' NamespacedNames attaching to various Gateway objects.
	allAssociatedHTTPRoutes sets.Set[string]
	// Set for storing GRPCRoutes' NamespacedNames attaching to various Gateway objects.
	allAssociatedGRPCRoutes sets.Set[string]
	// Set for storing TCPRoutes' NamespacedNames attaching to various Gateway objects.
	allAssociatedTCPRoutes sets.Set[string]
	// Set for storing UDPRoutes' NamespacedNames attaching to various Gateway objects.
	allAssociatedUDPRoutes sets.Set[string]
	// Set for storing backendRefs' BackendObjectReference referred by various Route objects.
	allAssociatedBackendRefs sets.Set[gwapiv1.BackendObjectReference]
	// Set for storing ClientTrafficPolicies' NamespacedNames referred by various Route objects.
	allAssociatedClientTrafficPolicies sets.Set[string]
	// Set for storing BackendTrafficPolicies' NamespacedNames referred by various Route objects.
	allAssociatedBackendTrafficPolicies sets.Set[string]
	// Set for storing SecurityPolicies' NamespacedNames referred by various Route objects.
	allAssociatedSecurityPolicies sets.Set[string]
	// Set for storing BackendTLSPolicies' NamespacedNames referred by various Backend objects.
	allAssociatedBackendTLSPolicies sets.Set[string]
	// Set for storing EnvoyExtensionPolicies' NamespacedNames attaching to various Gateway objects.
	allAssociatedEnvoyExtensionPolicies sets.Set[string]
	// extensionRefFilters is a map of filters managed by an extension.
	// The key is the namespaced name, group and kind of the filter and the value is the
	// unstructured form of the resource.
	extensionRefFilters map[utils.NamespacedNameWithGroupKind]unstructured.Unstructured
	// Set for storing HTTPRouteExtensions (Envoy Gateway or Custom) NamespacedNames referenced by various
	// route rules objects.
	allAssociatedHTTPRouteExtensionFilters sets.Set[utils.NamespacedNameWithGroupKind]

	// Set for storing BackendTrafficPolicies' NamespacedNames referred by various Backend objects.
	allAssociatedXBackendTrafficPolicies sets.Set[string]
}

func newResourceMapping() *resourceMappings {
	return &resourceMappings{
		allAssociatedGateways:                  sets.New[string](),
		allAssociatedReferenceGrants:           sets.New[string](),
		allAssociatedServiceImports:            sets.New[string](),
		allAssociatedEndpointSlices:            sets.New[string](),
		allAssociatedBackends:                  sets.New[string](),
		allAssociatedSecrets:                   sets.New[string](),
		allAssociatedConfigMaps:                sets.New[string](),
		allAssociatedNamespaces:                sets.New[string](),
		allAssociatedEnvoyProxies:              sets.New[string](),
		allAssociatedEnvoyPatchPolicies:        sets.New[string](),
		allAssociatedTLSRoutes:                 sets.New[string](),
		allAssociatedHTTPRoutes:                sets.New[string](),
		allAssociatedGRPCRoutes:                sets.New[string](),
		allAssociatedTCPRoutes:                 sets.New[string](),
		allAssociatedUDPRoutes:                 sets.New[string](),
		allAssociatedBackendRefs:               sets.New[gwapiv1.BackendObjectReference](),
		allAssociatedClientTrafficPolicies:     sets.New[string](),
		allAssociatedBackendTrafficPolicies:    sets.New[string](),
		allAssociatedSecurityPolicies:          sets.New[string](),
		allAssociatedBackendTLSPolicies:        sets.New[string](),
		allAssociatedEnvoyExtensionPolicies:    sets.New[string](),
		extensionRefFilters:                    map[utils.NamespacedNameWithGroupKind]unstructured.Unstructured{},
		allAssociatedHTTPRouteExtensionFilters: sets.New[utils.NamespacedNameWithGroupKind](),
	}
}
