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

// getBackendClusters resolves rd's BackendClusterRefs into their BackendCluster data. If none
// resolve but rd has a Name, a placeholder BackendCluster with no Settings is synthesized instead
// of returning empty, since TCP/UDP/TLS routes still need an EDS cluster to route to.
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

// resolvedBackendRef pairs a BackendClusterRef with its resolved BackendCluster, so callers can
// get this route's own weight for it without a separate lookup.
type resolvedBackendRef struct {
	ref *ir.BackendClusterRef
	bc  *ir.BackendCluster
}

// ResolvedWeight returns the weight bc's setting s contributes for ref: ref's own Weight when bc
// is merged, or s's own Weight otherwise.
func (p resolvedBackendRef) ResolvedWeight(s *ir.DestinationSetting) *uint32 {
	if p.bc.Merged {
		return p.ref.Weight
	}
	return s.Weight
}

// ToBackendWeights returns bc's weights as seen by ref (see ResolvedWeight).
func (p resolvedBackendRef) ToBackendWeights() *ir.BackendWeights {
	w := &ir.BackendWeights{}
	for _, s := range p.bc.Settings {
		w.AddWeighted(s, p.ResolvedWeight(s))
	}
	return w
}

// resolveBackendClusterRefs resolves each of rd's BackendClusterRefs to its BackendCluster,
// keeping the ref paired with its cluster. Mirrors getBackendClusters' fallback, but the
// fallback pair carries a synthetic ref too, so callers never need to nil-check ref.
func (t *Translator) resolveBackendClusterRefs(rd *ir.RouteDestination) []resolvedBackendRef {
	if rd == nil {
		return nil
	}
	pairs := make([]resolvedBackendRef, 0, len(rd.BackendClusterRefs))
	for _, ref := range rd.BackendClusterRefs {
		if bc, ok := t.backendIndex[ref.Name]; ok {
			pairs = append(pairs, resolvedBackendRef{ref: ref, bc: bc})
		}
	}
	if len(pairs) > 0 {
		return pairs
	}
	if rd.Name == "" {
		return nil
	}
	return []resolvedBackendRef{{
		ref: &ir.BackendClusterRef{Name: rd.Name},
		bc:  &ir.BackendCluster{Name: rd.Name, Metadata: rd.Metadata},
	}}
}

// singleResolvedClusterName returns the name of rd's one resolved BackendCluster. Falls back to
// rd's own route-scoped name if that's not exactly one cluster (should be impossible here).
func (t *Translator) singleResolvedClusterName(rd *ir.RouteDestination) string {
	if bcs := t.getBackendClusters(rd); len(bcs) == 1 {
		return bcs[0].Name
	}
	return rd.Name
}
