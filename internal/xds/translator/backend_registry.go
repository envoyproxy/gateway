// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"github.com/envoyproxy/gateway/internal/ir"
)

// backendClusterIndex resolves a BackendClusterRef's Name against the gateway's Backends
// registry (ir.Xds.Backends). Built once per Translate() call.
type backendClusterIndex map[string]*ir.BackendCluster

func newBackendClusterIndex(xdsIR *ir.Xds) backendClusterIndex {
	idx := make(backendClusterIndex, len(xdsIR.Backends))
	for _, bc := range xdsIR.Backends {
		idx[bc.Name] = bc
	}
	return idx
}

// resolve returns the BackendCluster for each ref, in order, by looking up its Name in the
// index. A ref whose Name isn't found in the index (e.g. it was pruned from Xds.Backends) is
// silently dropped.
func (idx backendClusterIndex) resolve(refs []*ir.BackendClusterRef) []*ir.BackendCluster {
	return ir.ResolveBackendClusterRefs(idx, refs)
}

// getBackendClusters resolves rd's BackendClusterRefs into their BackendCluster data via
// the registry. Falls back to rd.Settings (pre-BackendClusterRefs legacy shape) if rd has
// no refs at all. The fallback's returned BackendCluster is non-empty even when rd.Settings
// is empty (e.g. a route destination with zero healthy endpoints still needs an EDS cluster
// built) — callers that must distinguish "no backend configured at all" (e.g. ext-service/
// JWT/OIDC provider destinations, where an empty Settings means no backendRef was actually
// selected) need to check rd.Settings directly rather than relying on this method's slice
// length.
func (t *Translator) getBackendClusters(rd *ir.RouteDestination) []*ir.BackendCluster {
	if rd == nil {
		return nil
	}
	if len(rd.BackendClusterRefs) > 0 {
		return t.backendIndex.resolve(rd.BackendClusterRefs)
	}
	return []*ir.BackendCluster{{Name: rd.Name, Settings: rd.Settings, Metadata: rd.Metadata}}
}

// routeDestinationNeedsClusterPerSetting reports whether rd's backends must each get their
// own Envoy cluster (ZoneAware routing, per-setting filters, mixed address types, or
// invalid/empty settings) rather than being combined into one cluster.
func (t *Translator) routeDestinationNeedsClusterPerSetting(rd *ir.RouteDestination) bool {
	if len(rd.BackendClusterRefs) > 1 {
		return true
	}
	bcs := t.getBackendClusters(rd)
	if len(bcs) != 1 {
		return false
	}
	return bcs[0].NeedsClusterPerSetting()
}

// httpRouteNeedsClusterPerSetting reports whether h's backends must each get their own
// Envoy cluster, additionally accounting for HTTPRoute-level zone-aware routing config.
func (t *Translator) httpRouteNeedsClusterPerSetting(h *ir.HTTPRoute) bool {
	if h.Traffic != nil &&
		h.Traffic.LoadBalancer != nil &&
		(h.Traffic.LoadBalancer.PreferLocal != nil || len(h.Traffic.LoadBalancer.WeightedZones) > 0) {
		return true
	}
	return t.routeDestinationNeedsClusterPerSetting(h.Destination)
}

// backendWeights sums the valid/invalid/no-endpoints weight across rd's resolved backends.
func (t *Translator) backendWeights(rd *ir.RouteDestination) *ir.BackendWeights {
	bcs := t.getBackendClusters(rd)
	w := &ir.BackendWeights{Name: rd.Name}
	for _, bc := range bcs {
		bw := bc.ToBackendWeights()
		w.Valid += bw.Valid
		w.Invalid += bw.Invalid
		w.NoEndpoints += bw.NoEndpoints
	}
	return w
}
