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

// EndpointContexts maps each EDS cluster name to the configuration needed to
// rebuild its assignment. Contexts do not change after translation.
type EndpointContexts map[string]*EndpointContext

// EndpointContext keeps the configuration needed to rebuild one assignment.
// The endpoints its settings were built with live in EndpointInputs.
type EndpointContext struct {
	// Settings is each destination's configuration: name, weight, priority,
	// metadata, TLS presence. Its endpoint fields are left empty.
	Settings      []*ir.DestinationSetting
	HealthCheck   *ir.HealthCheck
	PreferLocal   *ir.PreferLocalZone
	WeightedZones []ir.WeightedZoneConfig
	// Baselines are used during full translation to detect incompatible patches.
	ClusterBaseline    *clusterv3.Cluster
	AssignmentBaseline *endpointv3.ClusterLoadAssignment
}

// EndpointInputs maps each EDS cluster name, then each destination setting
// name, to the endpoints and revision that setting is built from. Setting
// names are unique within a cluster.
type EndpointInputs map[string]map[string]EndpointInput

// EndpointInput is the endpoint part of a destination setting, which the fast
// path replaces.
type EndpointInput struct {
	Endpoints   []*ir.DestinationEndpoint
	AddressType *ir.DestinationAddressType
	Source      *ir.EndpointSource
}

// BuildSettings merges the context's configuration with its settings'
// endpoint inputs, for the assignment builder.
func (c *EndpointContext) BuildSettings(inputs map[string]EndpointInput) []*ir.DestinationSetting {
	settings := make([]*ir.DestinationSetting, len(c.Settings))
	for i, setting := range c.Settings {
		merged := *setting
		entry := inputs[setting.Name]
		merged.Endpoints, merged.AddressType, merged.EndpointSource = entry.Endpoints, entry.AddressType, entry.Source
		settings[i] = &merged
	}
	return settings
}
