// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"github.com/envoyproxy/gateway/internal/ir"
)

// backendClusterIndex resolves a BackendClusterRef's Name against ir.Xds.Backends.
type backendClusterIndex map[string]*ir.BackendCluster

func newBackendClusterIndex(xdsIR *ir.Xds) backendClusterIndex {
	idx := make(backendClusterIndex, len(xdsIR.Backends))
	for _, bc := range xdsIR.Backends {
		idx[bc.Name] = bc
	}
	return idx
}

// resolveBackendClusters resolves rd's BackendClusterRefs against backendIndex, dropping any ref
// whose Name isn't found.
func resolveBackendClusters(rd *ir.RouteDestination, backendIndex backendClusterIndex) []*ir.BackendCluster {
	if rd == nil || len(rd.BackendClusterRefs) == 0 {
		return nil
	}
	bcs := make([]*ir.BackendCluster, 0, len(rd.BackendClusterRefs))
	for _, ref := range rd.BackendClusterRefs {
		if bc, ok := backendIndex[ref.Name]; ok {
			bcs = append(bcs, bc)
		}
	}
	return bcs
}

// routeDestinationSettings returns rd's own Settings plus every resolved backend's Setting.
func routeDestinationSettings(rd *ir.RouteDestination, backendIndex backendClusterIndex) []*ir.DestinationSetting {
	backendClusters := resolveBackendClusters(rd, backendIndex)
	settings := make([]*ir.DestinationSetting, 0, len(rd.Settings)+len(backendClusters))
	settings = append(settings, rd.Settings...)
	for _, bc := range backendClusters {
		settings = append(settings, bc.Setting)
	}
	return settings
}

// backendClusterRef returns bc's original BackendClusterRef from rd. A BackendCluster's own
// Setting.Weight is always nil; the ref carries the real, route-scoped Weight instead.
func backendClusterRef(rd *ir.RouteDestination, bc *ir.BackendCluster) *ir.BackendClusterRef {
	for _, ref := range rd.BackendClusterRefs {
		if ref.Name == bc.Name {
			return ref
		}
	}
	return nil
}

// needsRouteCluster reports whether rd needs its own route-scoped cluster: it has non-merged
// Settings, or no backends at all (a placeholder destination still needs a real EDS cluster).
func needsRouteCluster(rd *ir.RouteDestination) bool {
	return len(rd.Settings) > 0 || len(rd.BackendClusterRefs) == 0
}

// singleClusterDestinationName returns rd's single Envoy Cluster name — TCP/UDP/TLS routes have no
// weighted-clusters mechanism, so exactly one of rd.Settings/rd.BackendClusterRefs must resolve it.
func singleClusterDestinationName(rd *ir.RouteDestination) string {
	if len(rd.Settings) > 0 {
		return rd.Name
	}
	if len(rd.BackendClusterRefs) > 0 {
		return rd.BackendClusterRefs[0].Name
	}
	return rd.Name
}
