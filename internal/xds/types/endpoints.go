// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package types

import (
	"github.com/envoyproxy/gateway/internal/ir"
)

// EndpointState is what the publisher keeps per snapshot to rebuild cluster
// endpoints without full translation. Both maps are keyed by EDS cluster name.
type EndpointState struct {
	// Contexts keeps each cluster's settings. Only a full result changes them.
	Contexts EndpointContexts
	// Endpoints keeps the endpoints and revision each destination was built from.
	// Fast updates advance them.
	Endpoints EndpointsByCluster
}

// EndpointContexts maps each EDS cluster name to the args that build its
// assignment, saved without endpoints. Endpoints are kept separately because
// they are the part that changes, and are added back on each rebuild.
type EndpointContexts map[string]*LoadAssignmentArgs

// LoadAssignmentArgs is everything that decides a cluster's endpoint assignment.
type LoadAssignmentArgs struct {
	Settings      []*ir.DestinationSetting
	HealthCheck   *ir.HealthCheck
	PreferLocal   *ir.PreferLocalZone
	WeightedZones []ir.WeightedZoneConfig
}

// EndpointsByCluster maps each EDS cluster name, then each destination setting
// name, to the endpoints and revision that setting is built from. Setting
// names are unique within a cluster.
type EndpointsByCluster map[string]map[string]SettingEndpoints

// SettingEndpoints is the endpoint part of a destination setting, which the fast
// path replaces.
type SettingEndpoints struct {
	Endpoints   []*ir.DestinationEndpoint
	AddressType *ir.DestinationAddressType
	Source      *ir.EndpointSource
}

// WithEndpoints returns a copy of the args with each setting's endpoints
// filled in from endpoints, ready for the assignment builder.
func (a LoadAssignmentArgs) WithEndpoints(endpoints map[string]SettingEndpoints) *LoadAssignmentArgs {
	settings := make([]*ir.DestinationSetting, len(a.Settings))
	for i, setting := range a.Settings {
		merged := *setting
		entry := endpoints[setting.Name]
		merged.Endpoints, merged.AddressType, merged.EndpointSource = entry.Endpoints, entry.AddressType, entry.Source
		settings[i] = &merged
	}
	a.Settings = settings
	return &a
}
