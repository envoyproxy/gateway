// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"sync"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/sets"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/utils"
)

type resourceMappings struct {
	// Map for storing Gateways' NamespacedNames.
	allAssociatedGateways *safeSet[string]
	// Map for storing ReferenceGrants' NamespacedNames.
	allAssociatedReferenceGrants *safeSet[string]
	// Map for storing ServiceImports' NamespacedNames.
	allAssociatedServiceImports *safeSet[string]
	// Map for storing EndpointSlices' NamespacedNames.
	allAssociatedEndpointSlices *safeSet[string]
	// Map for storing Secrets' NamespacedNames.
	allAssociatedSecrets *safeSet[string]
	// Map for storing ConfigMaps' NamespacedNames.
	allAssociatedConfigMaps *safeSet[string]
	// Map for storing namespaces for Route, Service and Gateway objects.
	allAssociatedNamespaces *safeSet[string]
	// Map for storing EnvoyProxies' NamespacedNames attaching to Gateway or GatewayClass.
	allAssociatedEnvoyProxies *safeSet[string]
	// Map for storing EnvoyPatchPolicies' NamespacedNames attaching to Gateway.
	allAssociatedEnvoyPatchPolicies *safeSet[string]
	// Map for storing TLSRoutes' NamespacedNames attaching to various Gateway objects.
	allAssociatedTLSRoutes *safeSet[string]
	// Map for storing HTTPRoutes' NamespacedNames attaching to various Gateway objects.
	allAssociatedHTTPRoutes *safeSet[string]
	// Map for storing GRPCRoutes' NamespacedNames attaching to various Gateway objects.
	allAssociatedGRPCRoutes *safeSet[string]
	// Map for storing TCPRoutes' NamespacedNames attaching to various Gateway objects.
	allAssociatedTCPRoutes *safeSet[string]
	// Map for storing UDPRoutes' NamespacedNames attaching to various Gateway objects.
	allAssociatedUDPRoutes *safeSet[string]
	// Map for storing backendRefs' BackendObjectReference referred by various Route objects.
	allAssociatedBackendRefs *safeSet[gwapiv1.BackendObjectReference]
	// Map for storing ClientTrafficPolicies' NamespacedNames referred by various Route objects.
	allAssociatedClientTrafficPolicies *safeSet[string]
	// Map for storing BackendTrafficPolicies' NamespacedNames referred by various Route objects.
	allAssociatedBackendTrafficPolicies *safeSet[string]
	// Map for storing SecurityPolicies' NamespacedNames referred by various Route objects.
	allAssociatedSecurityPolicies *safeSet[string]
	// Map for storing BackendTLSPolicies' NamespacedNames referred by various Backend objects.
	allAssociatedBackendTLSPolicies *safeSet[string]
	// Map for storing EnvoyExtensionPolicies' NamespacedNames attaching to various Gateway objects.
	allAssociatedEnvoyExtensionPolicies *safeSet[string]
	// extensionRefFilters is a map of filters managed by an extension.
	// The key is the namespaced name, group and kind of the filter and the value is the
	// unstructured form of the resource.
	extensionRefFilters map[utils.NamespacedNameWithGroupKind]unstructured.Unstructured
	// httpRouteFilters is a map of HTTPRouteFilters, where the key is the namespaced name,
	// group and kind of the HTTPFilter.
	httpRouteFilters map[utils.NamespacedNameWithGroupKind]*egv1a1.HTTPRouteFilter
}

type safeSet[T comparable] struct {
	lock   sync.RWMutex
	values sets.Set[T]
}

func newSafeSet[T comparable](items ...T) *safeSet[T] {
	return &safeSet[T]{values: sets.New[T](items...)}
}

func (s *safeSet[T]) Has(item T) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.values.Has(item)
}

func (s *safeSet[T]) Insert(item ...T) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.values.Insert(item...)
}

func newResourceMapping() *resourceMappings {
	return &resourceMappings{
		allAssociatedGateways:               newSafeSet[string](),
		allAssociatedReferenceGrants:        newSafeSet[string](),
		allAssociatedServiceImports:         newSafeSet[string](),
		allAssociatedEndpointSlices:         newSafeSet[string](),
		allAssociatedSecrets:                newSafeSet[string](),
		allAssociatedConfigMaps:             newSafeSet[string](),
		allAssociatedNamespaces:             newSafeSet[string](),
		allAssociatedEnvoyProxies:           newSafeSet[string](),
		allAssociatedEnvoyPatchPolicies:     newSafeSet[string](),
		allAssociatedTLSRoutes:              newSafeSet[string](),
		allAssociatedHTTPRoutes:             newSafeSet[string](),
		allAssociatedGRPCRoutes:             newSafeSet[string](),
		allAssociatedTCPRoutes:              newSafeSet[string](),
		allAssociatedUDPRoutes:              newSafeSet[string](),
		allAssociatedBackendRefs:            newSafeSet[gwapiv1.BackendObjectReference](),
		allAssociatedClientTrafficPolicies:  newSafeSet[string](),
		allAssociatedBackendTrafficPolicies: newSafeSet[string](),
		allAssociatedSecurityPolicies:       newSafeSet[string](),
		allAssociatedBackendTLSPolicies:     newSafeSet[string](),
		allAssociatedEnvoyExtensionPolicies: newSafeSet[string](),
		extensionRefFilters:                 map[utils.NamespacedNameWithGroupKind]unstructured.Unstructured{},
		httpRouteFilters:                    map[utils.NamespacedNameWithGroupKind]*egv1a1.HTTPRouteFilter{},
	}
}
