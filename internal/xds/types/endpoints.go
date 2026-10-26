// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package types

import (
	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"

	"github.com/envoyproxy/gateway/internal/ir"
)

type EndpointContexts map[string]*EndpointContext

// EndpointContext keeps only inputs needed to rebuild one assignment. Settings
// also hold the applied endpoint data and revisions; both paths replace them at commit.
type EndpointContext struct {
	Settings      []*ir.DestinationSetting
	HealthCheck   *ir.HealthCheck
	PreferLocal   *ir.PreferLocalZone
	WeightedZones []ir.WeightedZoneConfig
	// Baselines are used during full translation to detect incompatible patches.
	ClusterBaseline    *clusterv3.Cluster
	AssignmentBaseline *endpointv3.ClusterLoadAssignment
}
