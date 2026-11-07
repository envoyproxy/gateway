// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"context"

	cachev3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/telepresenceio/watchable"

	"github.com/envoyproxy/gateway/internal/endpoint"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/xds/cache"
	"github.com/envoyproxy/gateway/internal/xds/translator"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

type fullPublication struct {
	snapshot         *cachev3.Snapshot
	endpointContexts types.EndpointContexts
	// endpointInputs holds the endpoints and revisions this full result was
	// built from.
	endpointInputs types.EndpointInputs
	ctx            context.Context
}

// snapshotPublisher serializes full results, endpoint updates, and deletions.
// Full translation submits immutable results through fullResults; endpoint
// controller notifies endpointUpdates. Only run accesses the maps below.
type snapshotPublisher struct {
	ctx context.Context
	log logging.Logger

	// endpointUpdates carries fast-path endpoint updates from the endpoint
	// controller. It signals when updates are pending, and the fast path drains them.
	endpointUpdates *endpoint.UpdateQueue

	// fullResults queues full translation results and deletions in the order
	// the translator submitted them.
	fullResults chan watchable.Update[string, *fullPublication]

	// snapshots holds the published snapshot per IR key.
	snapshots map[string]*cachev3.Snapshot

	// endpointContexts holds, per IR key, the configuration of every EDS cluster
	// in the published snapshot.
	//
	// The full translation sets endpointContexts when it submits a full result,
	// and the endpoint fast path uses it to rebuild assignments from updated
	// endpoints, so the fast path doesn't rebuild the cluster configuration itself.
	endpointContexts map[string]types.EndpointContexts

	// publishedEndpoints holds, per IR key, the endpoints and revision each
	// cluster in the published snapshot was built with, indexed by backend.
	//
	// The endpoint fast path sets publishedEndpoints when it receives endpoint
	// updates, and full translation replays them with a newer revision over its
	// own, so a translation that started before a fast-path update and finished
	// after it cannot roll that update back.
	publishedEndpoints *publishedEndpointStore

	// publishSnapshot hands a snapshot to the xDS cache, which sends it
	// to the connected proxies.
	publishSnapshot func(string, *cachev3.Snapshot, context.Context) error
}

func newSnapshotPublisher(ctx context.Context, endpointUpdates *endpoint.UpdateQueue, publishSnapshot func(string, *cachev3.Snapshot, context.Context) error, log logging.Logger) *snapshotPublisher {
	return &snapshotPublisher{
		ctx: ctx, endpointUpdates: endpointUpdates, publishSnapshot: publishSnapshot, log: log,
		fullResults:        make(chan watchable.Update[string, *fullPublication], 2),
		snapshots:          make(map[string]*cachev3.Snapshot),
		endpointContexts:   make(map[string]types.EndpointContexts),
		publishedEndpoints: newPublishedEndpointStore(),
	}
}

func (p *snapshotPublisher) run() {
	defer func() {
		for key := range p.snapshots {
			p.forget(key)
		}
	}()
	for {
		select {
		case <-p.ctx.Done():
			return
		case result := <-p.fullResults:
			p.processFullResults(result)
		case <-p.endpointUpdates.Ready():
			p.processEndpointUpdates()
		}
	}
}

func (p *snapshotPublisher) enqueueFullResult(key string, publication *fullPublication) error {
	if err := p.ctx.Err(); err != nil {
		return err
	}
	select {
	case <-p.ctx.Done():
		return p.ctx.Err()
	case p.fullResults <- watchable.Update[string, *fullPublication]{Key: key, Value: publication, Delete: publication.snapshot == nil}:
		return nil
	}
}

func (p *snapshotPublisher) processFullResults(first watchable.Update[string, *fullPublication]) {
	batch := []watchable.Update[string, *fullPublication]{first}
	for remaining := len(p.fullResults); remaining > 0; remaining-- {
		batch = append(batch, <-p.fullResults)
	}
	for _, update := range message.CoalesceUpdatesPreservingDeletes(p.log, batch) {
		if update.Value.snapshot == nil {
			p.processDelete(update.Key, update.Value.ctx)
		} else {
			p.processFull(update.Key, update.Value)
		}
	}
}

func (p *snapshotPublisher) processFull(key string, publication *fullPublication) {
	// Published endpoints win over the full result's when their revision is newer.
	inputs, replacements, err := translator.ReplayPublishedEndpoints(publication.endpointContexts, publication.endpointInputs, p.publishedEndpoints.get(key))
	if err != nil {
		p.log.Error(err, "failed to refresh full publication", "ir-key", key)
		return
	}
	snapshot, err := cache.ReplaceEndpoints(publication.snapshot, replacements)
	if err != nil {
		p.log.Error(err, "failed to prepare full publication", "ir-key", key)
		return
	}
	p.snapshots[key], p.endpointContexts[key] = snapshot, publication.endpointContexts
	p.publishedEndpoints.set(key, inputs)
	endpointPublications.With(publicationPath.Value("full")).Increment()
	p.publish(publication.ctx, key, snapshot)
}

func (p *snapshotPublisher) processDelete(key string, ctx context.Context) {
	p.forget(key)
	endpointPublications.With(publicationPath.Value("delete")).Increment()
	p.publish(ctx, key, nil)
}

// forget drops a key's snapshot, endpoint contexts, and published endpoints.
func (p *snapshotPublisher) forget(key string) {
	p.publishedEndpoints.forget(key)
	delete(p.endpointContexts, key)
	delete(p.snapshots, key)
}

func (p *snapshotPublisher) publish(ctx context.Context, key string, snapshot *cachev3.Snapshot) {
	// The xDS cache stores the snapshot before it can fail, so later Envoy
	// requests still get it. Its errors are deterministic, so a retry would
	// fail the same way.
	if err := p.publishSnapshot(key, snapshot, ctx); err != nil {
		p.log.Error(err, "failed to publish snapshot", "ir-key", key)
	}
}
