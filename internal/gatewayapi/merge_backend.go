// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"fmt"
	"strings"

	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/ir"
)

// mergeBackendClusters implements the MergeBackends (CDS deduplication) optimization: routes that
// reference the same backend (kind/namespace/name/port) are collapsed onto a single Envoy cluster
// instead of Envoy Gateway generating one cluster per route rule.
//
// The decision is best-effort. A backendRef's DestinationSetting is only merged when it is safe to
// share the resulting cluster; otherwise the setting keeps its route-scoped name (current behavior)
// and falls back to a dedicated cluster. The actual deduplication happens naturally at the xDS
// layer: merged settings are renamed to a backend-identity key, so settings from different routes
// collapse via the existing name-based cluster dedup.
//
// It runs after all routes and BackendTrafficPolicies have been processed, so that route-targeted
// cluster-scoped settings (recorded on RouteDestination.RouteLevelClusterSettings) are known.
func (t *Translator) mergeBackendClusters(xdsIR resource.XdsIRMap) {
	for _, x := range xdsIR {
		// HTTP/GRPC routes translate multiple backends via weighted clusters, so each backend can be
		// merged into its own shared cluster (unless split-incompatible features are in use).
		for _, l := range x.HTTP {
			for _, r := range l.Routes {
				mergeRouteDestination(r.Destination, httpRouteSplitIncompatibleWithMerge(r), true)
			}
		}
		// TCP/UDP filter chains route to a single cluster, so only single-backend destinations are
		// merged; multi-backend splits keep today's aggregated cluster.
		for _, l := range x.TCP {
			for _, r := range l.Routes {
				mergeRouteDestination(r.Destination, loadBalancerSplitIncompatibleWithMerge(r.LoadBalancer), false)
			}
		}
		for _, l := range x.UDP {
			if l.Route != nil {
				mergeRouteDestination(l.Route.Destination, loadBalancerSplitIncompatibleWithMerge(l.Route.LoadBalancer), false)
			}
		}
	}
}

// mergeRouteDestination merges the eligible DestinationSettings of a single route destination onto
// shared, backend-identity-named clusters.
//
// The destination opts out of merging (keeping today's per-route-rule cluster) when:
//   - a route-targeted BackendTrafficPolicy set backend-cluster-scoped settings for it, or
//   - the rule splits traffic across multiple backends using a feature that does not work with
//     weighted clusters (consistent-hash load balancing or session persistence).
//
// Dynamic resolvers and custom backends are never mergeable and are left untouched.
//
// allowWeightedMerge must be true for route types that translate multiple backends into weighted
// clusters (HTTP/GRPC). When false (TCP/UDP), a destination with more than one backend is not
// merged because those filter chains route to a single cluster.
func mergeRouteDestination(d *ir.RouteDestination, splitIncompatible, allowWeightedMerge bool) {
	if d == nil || len(d.Settings) == 0 {
		return
	}
	if d.RouteLevelClusterSettings {
		return
	}
	if len(d.Settings) > 1 && (splitIncompatible || !allowWeightedMerge) {
		return
	}
	for _, s := range d.Settings {
		name, ok := mergedBackendClusterName(s)
		if !ok {
			continue
		}
		s.Name = name
		s.Merged = true
	}
}

// mergedBackendClusterName returns the backend-identity cluster name for a mergeable
// DestinationSetting. Routes referencing the same backend (kind/namespace/name/port) produce the
// same name and therefore collapse onto a single Envoy cluster. It returns ok=false for settings
// that must never be merged (dynamic resolvers, custom backends, or settings lacking backend
// identity metadata).
func mergedBackendClusterName(s *ir.DestinationSetting) (string, bool) {
	if s == nil || s.IsDynamicResolver || s.IsCustomBackend {
		return "", false
	}
	md := s.Metadata
	if md == nil || md.Kind == "" || md.Name == "" {
		return "", false
	}
	name := fmt.Sprintf("backend/%s/%s/%s", strings.ToLower(md.Kind), md.Namespace, md.Name)
	if md.SectionName != "" {
		name += "/" + md.SectionName
	}
	return name, true
}

// httpRouteSplitIncompatibleWithMerge reports whether an HTTPRoute uses a traffic-splitting feature
// that cannot be represented once each backend becomes its own cluster (weighted clusters).
func httpRouteSplitIncompatibleWithMerge(r *ir.HTTPRoute) bool {
	if r == nil {
		return false
	}
	if r.SessionPersistence != nil {
		return true
	}
	if r.Traffic != nil {
		return loadBalancerSplitIncompatibleWithMerge(r.Traffic.LoadBalancer)
	}
	return false
}

// loadBalancerSplitIncompatibleWithMerge reports whether a load balancer configuration relies on
// host selection (consistent hashing) that is broken by splitting traffic across weighted clusters.
func loadBalancerSplitIncompatibleWithMerge(lb *ir.LoadBalancer) bool {
	return lb != nil && lb.ConsistentHash != nil
}
