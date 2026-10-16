// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"slices"

	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

// BuildClusterLoadAssignment rebuilds a cluster's ClusterLoadAssignment from a
// ClusterLoadAssignment context captured during full translation. It is the endpoint fast
// path's entry point into the same CLA construction used by full translation,
// so both paths produce identical output for the same inputs.
func BuildClusterLoadAssignment(ec *types.ClusterLoadAssignmentContext) *endpointv3.ClusterLoadAssignment {
	return buildXdsClusterLoadAssignment(ec.ClusterName, ec.Settings, ec.HealthCheck, ec.PreferLocal, ec.WeightedZones)
}

// endpointFastPathEnabled reports whether the endpoint fast path can serve this
// cluster. The gatewayapi translation only populates EndpointSource when the
// EndpointFastPath runtime flag is on and the backend is EndpointSlice-backed.
func endpointFastPathEnabled(settings []*ir.DestinationSetting) bool {
	return slices.ContainsFunc(settings, func(ds *ir.DestinationSetting) bool {
		return ds != nil && ds.EndpointSource != nil
	})
}

// buildClusterLoadAssignmentContext captures everything the endpoint fast path needs to rebuild
// this cluster's CLA without a full translation, or nil when the fast path is off.
// The context owns deep copies: the IR it is built from stays live in the watchable
// layer, which compares it for equality on the next publish.
func buildClusterLoadAssignmentContext(args *xdsClusterArgs, lb *ir.LoadBalancer) *types.ClusterLoadAssignmentContext {
	if !endpointFastPathEnabled(args.settings) {
		return nil
	}
	settings := make([]*ir.DestinationSetting, len(args.settings))
	for i, s := range args.settings {
		settings[i] = s.DeepCopy()
	}
	ec := &types.ClusterLoadAssignmentContext{
		ClusterName: args.name,
		Settings:    settings,
		HealthCheck: args.healthCheck.DeepCopy(),
		PreferLocal: lb.PreferLocal.DeepCopy(),
	}
	if lb.WeightedZones != nil {
		ec.WeightedZones = make([]ir.WeightedZoneConfig, len(lb.WeightedZones))
		for i, wz := range lb.WeightedZones {
			wz.DeepCopyInto(&ec.WeightedZones[i])
		}
	}
	return ec
}
