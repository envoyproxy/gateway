// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"

	"github.com/envoyproxy/gateway/internal/xds/types"
)

// BuildClusterLoadAssignment rebuilds a cluster's ClusterLoadAssignment from an
// endpoint context captured during full translation. It is the endpoint fast
// path's entry point into the same CLA construction used by full translation,
// so both paths produce identical output for the same inputs.
func BuildClusterLoadAssignment(ec *types.EndpointContext) *endpointv3.ClusterLoadAssignment {
	return buildXdsClusterLoadAssignment(ec.ClusterName, ec.Settings, ec.HealthCheck, ec.PreferLocal, ec.WeightedZones)
}
