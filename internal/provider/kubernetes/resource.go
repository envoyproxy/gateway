// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/utils"
	"github.com/envoyproxy/gateway/internal/utils/safeset"
)

type resourceMappings struct {
	// Map for storing Gateways' NamespacedNames.
	allAssociatedGateways *safeset.SafeSet[string]
	// Map for storing ReferenceGrants' NamespacedNames.
	allAssociatedReferenceGrants *safeset.SafeSet[string]
	// Map for storing ServiceImports' NamespacedNames.
	allAssociatedServiceImports *safeset.SafeSet[string]
	// Map for storing EndpointSlices' NamespacedNames.
	allAssociatedEndpointSlices *safeset.SafeSet[string]
	// Map for storing Secrets' NamespacedNames.
	allAssociatedSecrets *safeset.SafeSet[string]
	// Map for storing ConfigMaps' NamespacedNames.
	allAssociatedConfigMaps *safeset.SafeSet[string]
	// Map for storing namespaces for Route, Service and Gateway objects.
	allAssociatedNamespaces *safeset.SafeSet[string]
	// Map for storing EnvoyProxies' NamespacedNames attaching to Gateway or GatewayClass.
	allAssociatedEnvoyProxies *safeset.SafeSet[string]
	// Map for storing EnvoyPatchPolicies' NamespacedNames attaching to Gateway.
	allAssociatedEnvoyPatchPolicies *safeset.SafeSet[string]
	// Map for storing TLSRoutes' NamespacedNames attaching to various Gateway objects.
	allAssociatedTLSRoutes *safeset.SafeSet[string]
	// Map for storing HTTPRoutes' NamespacedNames attaching to various Gateway objects.
	allAssociatedHTTPRoutes *safeset.SafeSet[string]
	// Map for storing GRPCRoutes' NamespacedNames attaching to various Gateway objects.
	allAssociatedGRPCRoutes *safeset.SafeSet[string]
	// Map for storing TCPRoutes' NamespacedNames attaching to various Gateway objects.
	allAssociatedTCPRoutes *safeset.SafeSet[string]
	// Map for storing UDPRoutes' NamespacedNames attaching to various Gateway objects.
	allAssociatedUDPRoutes *safeset.SafeSet[string]
	// Map for storing backendRefs' BackendObjectReference referred by various Route objects.
	allAssociatedBackendRefs *safeset.SafeSet[gwapiv1.BackendObjectReference]
	// Map for storing ClientTrafficPolicies' NamespacedNames referred by various Route objects.
	allAssociatedClientTrafficPolicies *safeset.SafeSet[string]
	// Map for storing BackendTrafficPolicies' NamespacedNames referred by various Route objects.
	allAssociatedBackendTrafficPolicies *safeset.SafeSet[string]
	// Map for storing SecurityPolicies' NamespacedNames referred by various Route objects.
	allAssociatedSecurityPolicies *safeset.SafeSet[string]
	// Map for storing BackendTLSPolicies' NamespacedNames referred by various Backend objects.
	allAssociatedBackendTLSPolicies *safeset.SafeSet[string]
	// Map for storing EnvoyExtensionPolicies' NamespacedNames attaching to various Gateway objects.
	allAssociatedEnvoyExtensionPolicies *safeset.SafeSet[string]
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
		allAssociatedGateways:               safeset.NewSafeSet[string](),
		allAssociatedReferenceGrants:        safeset.NewSafeSet[string](),
		allAssociatedServiceImports:         safeset.NewSafeSet[string](),
		allAssociatedEndpointSlices:         safeset.NewSafeSet[string](),
		allAssociatedSecrets:                safeset.NewSafeSet[string](),
		allAssociatedConfigMaps:             safeset.NewSafeSet[string](),
		allAssociatedNamespaces:             safeset.NewSafeSet[string](),
		allAssociatedEnvoyProxies:           safeset.NewSafeSet[string](),
		allAssociatedEnvoyPatchPolicies:     safeset.NewSafeSet[string](),
		allAssociatedTLSRoutes:              safeset.NewSafeSet[string](),
		allAssociatedHTTPRoutes:             safeset.NewSafeSet[string](),
		allAssociatedGRPCRoutes:             safeset.NewSafeSet[string](),
		allAssociatedTCPRoutes:              safeset.NewSafeSet[string](),
		allAssociatedUDPRoutes:              safeset.NewSafeSet[string](),
		allAssociatedBackendRefs:            safeset.NewSafeSet[gwapiv1.BackendObjectReference](),
		allAssociatedClientTrafficPolicies:  safeset.NewSafeSet[string](),
		allAssociatedBackendTrafficPolicies: safeset.NewSafeSet[string](),
		allAssociatedSecurityPolicies:       safeset.NewSafeSet[string](),
		allAssociatedBackendTLSPolicies:     safeset.NewSafeSet[string](),
		allAssociatedEnvoyExtensionPolicies: safeset.NewSafeSet[string](),
		extensionRefFilters:                 map[utils.NamespacedNameWithGroupKind]unstructured.Unstructured{},
		httpRouteFilters:                    map[utils.NamespacedNameWithGroupKind]*egv1a1.HTTPRouteFilter{},
	}
}
