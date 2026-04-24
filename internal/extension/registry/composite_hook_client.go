// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package registry

import (
	"fmt"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/extension/types"
	"github.com/envoyproxy/gateway/internal/ir"
)

var _ types.XDSHookClient = (*compositeXDSHookClient)(nil)

// hookClientEntry pairs an XDSHookClient with its parent manager's metadata.
// Matching is by group/kind only, mirroring runner.ExtensionGroupKinds and
// Manager.HasExtension; exact-GVK matching would silently drop resources
// when CRDs serve multiple versions.
type hookClientEntry struct {
	name              string
	client            types.XDSHookClient
	failOpen          bool
	resourceGKSet     sets.Set[schema.GroupKind] // used for per-extension resource filtering in PostRouteModifyHook, PostClusterModifyHook
	policyGKSet       sets.Set[schema.GroupKind] // used for per-extension policy filtering in PostHTTPListenerModifyHook, PostTranslateModifyHook
	translationConfig *egv1a1.TranslationConfig  // used for per-extension resource-type gating in PostTranslateModifyHook
}

// compositeXDSHookClient chains multiple XDSHookClient calls sequentially.
// Each client's output becomes the next client's input.
type compositeXDSHookClient struct {
	entries []hookClientEntry
}

func (c *compositeXDSHookClient) PostRouteModifyHook(r *route.Route, routeHostnames []string, extensionResources []*unstructured.Unstructured) (*route.Route, error) {
	current := r
	for _, entry := range c.entries {
		filtered := filterResourcesByGK(extensionResources, entry.resourceGKSet)
		result, err := entry.client.PostRouteModifyHook(current, routeHostnames, filtered)
		if err != nil {
			if entry.failOpen {
				continue
			}
			return nil, fmt.Errorf("extension %q: %w", entry.name, err)
		}
		current = result
	}
	return current, nil
}

func (c *compositeXDSHookClient) PostVirtualHostModifyHook(vh *route.VirtualHost) (*route.VirtualHost, error) {
	current := vh
	for _, entry := range c.entries {
		result, err := entry.client.PostVirtualHostModifyHook(current)
		if err != nil {
			if entry.failOpen {
				continue
			}
			return nil, fmt.Errorf("extension %q: %w", entry.name, err)
		}
		current = result
	}
	return current, nil
}

func (c *compositeXDSHookClient) PostEndpointsModifyHook(loadAssignment *endpoint.ClusterLoadAssignment) (*endpoint.ClusterLoadAssignment, error) {
	current := loadAssignment
	for _, entry := range c.entries {
		result, err := entry.client.PostEndpointsModifyHook(current)
		if err != nil {
			if entry.failOpen {
				continue
			}
			return nil, fmt.Errorf("extension %q: %w", entry.name, err)
		}
		current = result
	}
	return current, nil
}

func (c *compositeXDSHookClient) PostHTTPListenerModifyHook(l *listener.Listener, extensionResources []*unstructured.Unstructured) (*listener.Listener, error) {
	current := l
	for _, entry := range c.entries {
		filtered := filterResourcesByGK(extensionResources, entry.policyGKSet)
		result, err := entry.client.PostHTTPListenerModifyHook(current, filtered)
		if err != nil {
			if entry.failOpen {
				continue
			}
			return nil, fmt.Errorf("extension %q: %w", entry.name, err)
		}
		current = result
	}
	return current, nil
}

func (c *compositeXDSHookClient) PostClusterModifyHook(cl *cluster.Cluster, extensionResources []*unstructured.Unstructured) (*cluster.Cluster, error) {
	current := cl
	for _, entry := range c.entries {
		filtered := filterResourcesByGK(extensionResources, entry.resourceGKSet)
		result, err := entry.client.PostClusterModifyHook(current, filtered)
		if err != nil {
			if entry.failOpen {
				continue
			}
			return nil, fmt.Errorf("extension %q: %w", entry.name, err)
		}
		current = result
	}
	return current, nil
}

// PostTranslateModifyHook chains the PostTranslateModifyHook call across all extensions sequentially.
// Each extension only receives the resource types it declared interest in via its TranslationConfig.
// The result for each resource type is only taken back if the extension declared interest in it.
//
// Note: policies are passed as []*ir.UnstructuredRef (slice of pointers). While each extension
// receives a filtered subset via gRPC (which serializes/deserializes, preventing mutation),
// the interface signature technically allows in-process implementations to mutate the underlying
// objects. Earlier extensions in the chain could therefore affect policies seen by later extensions.
func (c *compositeXDSHookClient) PostTranslateModifyHook(
	clusters []*cluster.Cluster,
	secrets []*tls.Secret,
	listeners []*listener.Listener,
	routes []*route.RouteConfiguration,
	extensionPolicies []*ir.UnstructuredRef,
) ([]*cluster.Cluster, []*tls.Secret, []*listener.Listener, []*route.RouteConfiguration, error) {
	currentClusters := clusters
	currentSecrets := secrets
	currentListeners := listeners
	currentRoutes := routes

	for _, entry := range c.entries {
		// Per-extension policy filtering: only send policies matching this extension's declared GKs
		filteredPolicies := filterPoliciesByGK(extensionPolicies, entry.policyGKSet)

		// Per-extension resource-type gating: only pass resource types this extension declared interest in.
		// This mirrors the behavior in processExtensionPostTranslationHook (extension.go) where
		// resources are only populated and reassigned based on GetTranslationHookConfig().
		tc := entry.translationConfig
		var entryClusters []*cluster.Cluster
		var entrySecrets []*tls.Secret
		var entryListeners []*listener.Listener
		var entryRoutes []*route.RouteConfiguration
		if tc.ShouldIncludeClusters() {
			entryClusters = currentClusters
		}
		if tc.ShouldIncludeSecrets() {
			entrySecrets = currentSecrets
		}
		if tc.ShouldIncludeListeners() {
			entryListeners = currentListeners
		}
		if tc.ShouldIncludeRoutes() {
			entryRoutes = currentRoutes
		}

		rc, rs, rl, rr, err := entry.client.PostTranslateModifyHook(
			entryClusters, entrySecrets, entryListeners, entryRoutes, filteredPolicies,
		)
		if err != nil {
			if entry.failOpen {
				continue
			}
			return nil, nil, nil, nil, fmt.Errorf("extension %q: %w", entry.name, err)
		}

		// Only take back resource types the extension declared interest in
		if tc.ShouldIncludeClusters() {
			currentClusters = rc
		}
		if tc.ShouldIncludeSecrets() {
			currentSecrets = rs
		}
		if tc.ShouldIncludeListeners() {
			currentListeners = rl
		}
		if tc.ShouldIncludeRoutes() {
			currentRoutes = rr
		}
	}

	return currentClusters, currentSecrets, currentListeners, currentRoutes, nil
}

// filterResourcesByGK returns only those unstructured resources whose group/kind matches
// the given set. Version is intentionally ignored so behavior matches single-manager mode,
// where runner.ExtensionGroupKinds and Manager.HasExtension also match by group+kind only.
// If the set is nil or empty, all resources are returned (for backward compatibility).
func filterResourcesByGK(resources []*unstructured.Unstructured, gkSet sets.Set[schema.GroupKind]) []*unstructured.Unstructured {
	if gkSet.Len() == 0 {
		return resources
	}
	var filtered []*unstructured.Unstructured
	for _, r := range resources {
		if r == nil {
			continue
		}
		if gkSet.Has(r.GroupVersionKind().GroupKind()) {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

// filterPoliciesByGK returns only those policies whose group/kind matches the given set.
// Version is intentionally ignored; see filterResourcesByGK for rationale.
// If the set is nil or empty, all policies are returned (for backward compatibility).
func filterPoliciesByGK(policies []*ir.UnstructuredRef, gkSet sets.Set[schema.GroupKind]) []*ir.UnstructuredRef {
	if gkSet.Len() == 0 {
		return policies
	}
	var filtered []*ir.UnstructuredRef
	for _, p := range policies {
		if p == nil || p.Object == nil {
			continue
		}
		if gkSet.Has(p.Object.GroupVersionKind().GroupKind()) {
			filtered = append(filtered, p)
		}
	}
	return filtered
}
