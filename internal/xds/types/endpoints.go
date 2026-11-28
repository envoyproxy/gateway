// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package types

import (
	"github.com/envoyproxy/gateway/internal/ir"
)

// EndpointContexts maps each EDS cluster name to its endpoint context: the args
// that build its assignment. The publisher keeps them per snapshot to rebuild
// cluster endpoints without full translation.
type EndpointContexts map[string]*LoadAssignmentArgs

// LoadAssignmentArgs is everything that decides a cluster's endpoint assignment.
type LoadAssignmentArgs struct {
	// Settings are the cluster's destinations. The fast path replaces only each
	// setting's endpoints, address type and endpoint source, in a copy.
	Settings []*ir.DestinationSetting

	// The fields below come from the last full translation. The fast path keeps
	// them as they are when it rebuilds the assignment.
	HealthCheck   *ir.HealthCheck
	PreferLocal   *ir.PreferLocalZone
	WeightedZones []ir.WeightedZoneConfig
}
