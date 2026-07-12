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

// getBackendClusters resolves rd's BackendClusterRefs into their BackendCluster data. If rd has a
// Name but no refs resolve to anything, a single placeholder BackendCluster with no Settings is
// synthesized instead of returning empty: TCP/UDP/TLS routes have no per-request fallback, so
// Envoy still needs an EDS cluster to route to, even one with zero endpoints.
func (t *Translator) getBackendClusters(rd *ir.RouteDestination) []*ir.BackendCluster {
	if rd == nil {
		return nil
	}
	if clusters := t.backendIndex.resolve(rd.BackendClusterRefs); len(clusters) > 0 {
		return clusters
	}
	if rd.Name == "" {
		return nil
	}
	return []*ir.BackendCluster{{Name: rd.Name, Metadata: rd.Metadata}}
}

// singleResolvedClusterName returns the name of rd's one resolved BackendCluster, when there
// is exactly one. This layer doesn't assume data from earlier layers is always well-formed:
// if the ref count isn't exactly one (should be impossible in this call path) or the ref
// doesn't resolve to a registered BackendCluster, something unexpected has happened, and rd's
// own route-scoped name is the best available fallback.
func (t *Translator) singleResolvedClusterName(rd *ir.RouteDestination) string {
	if bcs := t.getBackendClusters(rd); len(bcs) == 1 {
		return bcs[0].Name
	}
	return rd.Name
}
