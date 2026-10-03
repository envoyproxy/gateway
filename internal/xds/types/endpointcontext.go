// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package types

import (
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/message"
)

// ClusterLoadAssignmentContext captures, for one EDS cluster, the inputs needed to rebuild its
// ClusterLoadAssignment outside a full translation. It is recorded during full
// translation and consumed by the xDS runner's endpoint fast path.
type ClusterLoadAssignmentContext struct {
	// ClusterName is the xDS cluster (and CLA) name.
	ClusterName string
	// Settings are the cluster's destination settings from the translated IR.
	// Their order matters: the per-endpoint TLS transport socket match metadata
	// embeds the setting index.
	Settings []*ir.DestinationSetting
	// HealthCheck carries the active health check overrides applied to endpoints.
	HealthCheck *ir.HealthCheck
	// PreferLocal is the cluster-level zone-aware routing config.
	PreferLocal *ir.PreferLocalZone
	// WeightedZones is the explicit per-zone weighting config.
	WeightedZones []ir.WeightedZoneConfig
}

// AddClusterLoadAssignmentContext records the context for an EDS cluster.
func (t *ResourceVersionTable) AddClusterLoadAssignmentContext(ec *ClusterLoadAssignmentContext) {
	if t.EDSContexts == nil {
		t.EDSContexts = make(map[string][]*ClusterLoadAssignmentContext)
	}
	seen := make(map[string]struct{})
	for _, setting := range ec.Settings {
		if setting.EndpointSource == nil {
			continue
		}
		key := message.BackendKey(setting.EndpointSource.Kind, setting.EndpointSource.Namespace, setting.EndpointSource.Name)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		t.EDSContexts[key] = append(t.EDSContexts[key], ec)
	}
}
