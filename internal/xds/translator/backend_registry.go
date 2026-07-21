// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"github.com/envoyproxy/gateway/internal/ir"
)

// backendClusterIndex resolves a BackendClusterRef's Name against ir.Xds.Backends.
// Built once per Translate() call.
type backendClusterIndex map[string]*ir.BackendCluster

func newBackendClusterIndex(xdsIR *ir.Xds) backendClusterIndex {
	idx := make(backendClusterIndex, len(xdsIR.Backends))
	for _, bc := range xdsIR.Backends {
		idx[bc.Name] = bc
	}
	return idx
}

// resolve returns the BackendCluster for each ref, in order. A ref whose Name isn't found
// in the index is silently dropped.
func (idx backendClusterIndex) resolve(refs []*ir.BackendClusterRef) []*ir.BackendCluster {
	return ir.ResolveBackendClusterRefs(idx, refs)
}

// resolveMergedBackendClusters resolves rd's BackendClusterRefs (the genuinely merged,
// identity-deduplicated backends for this destination) against idx. A ref whose Name isn't found
// in idx is silently dropped (matches this package's existing convention for stale/missing
// references). Non-merged backends never appear here — they live in rd.Settings, handled entirely
// separately by the reverted pre-PR code paths in translator.go/cluster.go/route.go.
func (t *Translator) resolveMergedBackendClusters(rd *ir.RouteDestination) []*ir.BackendCluster {
	if rd == nil || len(rd.BackendClusterRefs) == 0 {
		return nil
	}
	bcs := make([]*ir.BackendCluster, 0, len(rd.BackendClusterRefs))
	for _, ref := range rd.BackendClusterRefs {
		if bc, ok := t.backendIndex[ref.Name]; ok {
			bcs = append(bcs, bc)
		}
	}
	return bcs
}

// mergedBackendClusterRef returns bc's original BackendClusterRef from rd, so callers can recover
// its route-scoped Weight (a merged BackendCluster's own Setting.Weight is always nil).
func mergedBackendClusterRef(rd *ir.RouteDestination, bc *ir.BackendCluster) *ir.BackendClusterRef {
	for _, ref := range rd.BackendClusterRefs {
		if ref.Name == bc.Name {
			return ref
		}
	}
	return nil
}

// tcpDestinationClusterName returns rd's single Envoy Cluster name. TCP/UDP/TLS routes have no
// weighted-clusters mechanism, so a destination must resolve to exactly one cluster: its own
// route-scoped name (non-merged, rd.Settings) or its one merged backend's name.
func tcpDestinationClusterName(rd *ir.RouteDestination) string {
	if len(rd.Settings) > 0 {
		return rd.Name
	}
	if len(rd.BackendClusterRefs) > 0 {
		return rd.BackendClusterRefs[0].Name
	}
	return rd.Name
}
