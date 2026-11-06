// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"time"

	"github.com/envoyproxy/gateway/internal/endpoint"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/cache"
	"github.com/envoyproxy/gateway/internal/xds/translator"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

// processEndpointUpdates takes the pending endpoint updates, rebuilds the EDS
// resources they affect using the new endpoints, and publishes the updated
// snapshots.
func (p *snapshotPublisher) processEndpointUpdates() {
	// Group one drained batch by IR key so each affected cluster is rebuilt once.
	updatesByKey := make(map[string]map[ir.EndpointBackend]*endpoint.BackendEndpoints)
	for _, backendEndpoints := range p.endpointUpdates.Drain() {
		// Find all the IR keys that are affected by each backend update.
		for key := range p.publishedEndpoints.snapshotsUsing(backendEndpoints.Backend) {
			if updatesByKey[key] == nil {
				updatesByKey[key] = make(map[ir.EndpointBackend]*endpoint.BackendEndpoints)
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
func (p *snapshotPublisher) applyEndpointUpdates(key string, updates map[ir.EndpointBackend]*endpoint.BackendEndpoints) {
	start := time.Now()
	defer func() { endpointRefreshDuration.Record(time.Since(start).Seconds()) }()
	endpointContexts, published, current := p.endpointContexts[key], p.publishedEndpoints.get(key), p.snapshots[key]
	affected := make(types.EndpointContexts)
	for backend := range updates {
		for name := range p.publishedEndpoints.snapshotsUsing(backend)[key] {
			affected[name] = endpointContexts[name]
		}
	}
	next, replacements, err := translator.ApplyEndpointUpdates(affected, published, updates)
	if err != nil {
		p.log.Error(err, "failed to refresh endpoints", "ir-key", key)
		return
	}
	snapshot, err := cache.ReplaceEndpoints(current, replacements)
	if err != nil {
		p.log.Error(err, "failed to prepare endpoint publication", "ir-key", key)
		return
	}

	// Update the published endpoints and snapshot in the publisher.
	if len(replacements) != 0 {
		p.snapshots[key] = snapshot
		p.publishedEndpoints.set(key, next)
	}
	if snapshot != current {
		endpointPublications.With(publicationPath.Value("fast")).Increment()
		p.publish(p.ctx, key, snapshot)
	}
}
