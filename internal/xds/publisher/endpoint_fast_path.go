// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package publisher

import (
	"maps"
	"time"

	"github.com/envoyproxy/gateway/internal/endpoint"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/cache"
	"github.com/envoyproxy/gateway/internal/xds/translator"
)

// processEndpointUpdates takes the pending endpoint updates, rebuilds the EDS
// resources they affect using the new endpoints, and publishes the updated
// snapshots.
func (p *Publisher) processEndpointUpdates() {
	// Group one drained batch by IR key so each affected cluster is rebuilt once.
	updatesByKey := make(map[string]map[ir.ServiceKey]*endpoint.BackendEndpoints)
	for _, backendEndpoints := range p.endpointUpdates.Drain() {
		// Find all the IR keys that are affected by each backend update.
		for key := range p.publishedEndpoints.clustersUsing(backendEndpoints.Backend) {
			if updatesByKey[key] == nil {
				updatesByKey[key] = make(map[ir.ServiceKey]*endpoint.BackendEndpoints)
			}
			updatesByKey[key][backendEndpoints.Backend] = backendEndpoints
		}
	}
	for key, updates := range updatesByKey {
		p.applyEndpointUpdates(key, updates)
	}
}

// applyEndpointUpdates rebuilds one snapshot's EDS resources from the updated
// backends and publishes the result.
func (p *Publisher) applyEndpointUpdates(key string, updates map[ir.ServiceKey]*endpoint.BackendEndpoints) {
	start := time.Now()
	defer func() { endpointRefreshDuration.Record(time.Since(start).Seconds()) }()
	published, current := p.publishedEndpoints.get(key), p.snapshots[key]
	affected := make(map[string]struct{})
	for backend := range updates {
		for name := range p.publishedEndpoints.clustersUsing(backend)[key] {
			affected[name] = struct{}{}
		}
	}
	updatedContexts, replacements, err := translator.ApplyEndpointUpdates(published, maps.Keys(affected), updates)
	if err != nil {
		p.log.Error(err, "failed to refresh endpoints", "ir-key", key)
		return
	}
	snapshot, err := cache.ReplaceEndpoints(current, replacements)
	if err != nil {
		p.log.Error(err, "failed to prepare endpoint publication", "ir-key", key)
		return
	}

	if len(replacements) != 0 {
		// Save the new snapshot as the published one. Only its endpoint
		// assignments differ; listeners, routes and clusters are shared.
		p.snapshots[key] = snapshot

		// Save the endpoints just applied, so the next update compares its
		// revisions against them. The fast path only swaps endpoints, so the
		// backend index stays the same and isn't rebuilt.
		p.publishedEndpoints.updateEndpoints(key, updatedContexts)
	}
	if snapshot != current {
		p.publish(p.ctx, pathFast, key, snapshot)
	}
}
