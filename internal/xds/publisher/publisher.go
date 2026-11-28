// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package publisher

import (
	"context"

	cachev3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/telepresenceio/watchable"

	"github.com/envoyproxy/gateway/internal/endpoint"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/metrics"
	"github.com/envoyproxy/gateway/internal/xds/cache"
	"github.com/envoyproxy/gateway/internal/xds/translator"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

type fullPublication struct {
	snapshot *cachev3.Snapshot
	// endpointContexts holds the cluster settings, endpoints and revisions this
	// full result was built from.
	endpointContexts types.EndpointContexts
	ctx              context.Context
}

// Publisher publishes full results and fast endpoint updates. Three rules
// keep the two paths consistent:
//  1. The full path still runs for every endpoint change, so the fast path only
//     publishes the same endpoints earlier, and can skip anything it can't handle.
//  2. The provider gives both paths' endpoints revisions from one counter, so a
//     backend's revisions grow in the order its EndpointSlices were read.
//  3. This one goroutine publishes everything, and on both paths a cluster
//     setting only takes endpoints with a newer revision than it has.
type Publisher struct {
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

	// publishedEndpoints holds, per IR key, the endpoint contexts of the published
	// snapshot: each cluster's settings, and the endpoints and revision it was
	// built with, indexed by backend. A full result sets it, and fast updates
	// advance its endpoints.
	publishedEndpoints *publishedEndpointStore

	// publishSnapshot hands a snapshot to the xDS cache, which sends it
	// to the connected proxies.
	publishSnapshot func(string, *cachev3.Snapshot, context.Context) error
}

// New returns a publisher that sends snapshots with publishSnapshot and takes
// fast-path endpoint updates from endpointUpdates. Start it with Run.
func New(ctx context.Context, endpointUpdates *endpoint.UpdateQueue, publishSnapshot func(string, *cachev3.Snapshot, context.Context) error, log logging.Logger) *Publisher {
	return &Publisher{
		ctx: ctx, endpointUpdates: endpointUpdates, publishSnapshot: publishSnapshot, log: log,
		fullResults:        make(chan watchable.Update[string, *fullPublication], 2),
		snapshots:          make(map[string]*cachev3.Snapshot),
		publishedEndpoints: newPublishedEndpointStore(),
	}
}

// Run handles full results and endpoint updates one at a time until ctx is done.
func (p *Publisher) Run() {
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

// Publish queues a full translation result for the key. It blocks while the
// queue is full, and fails only once the publisher's context is done.
func (p *Publisher) Publish(ctx context.Context, key string, snapshot *cachev3.Snapshot, contexts types.EndpointContexts) error {
	return p.enqueueFullResult(key, &fullPublication{snapshot: snapshot, endpointContexts: contexts, ctx: ctx})
}

// Delete queues the removal of the key's snapshot, in order with its results.
func (p *Publisher) Delete(ctx context.Context, key string) error {
	return p.enqueueFullResult(key, &fullPublication{ctx: ctx})
}

func (p *Publisher) enqueueFullResult(key string, publication *fullPublication) error {
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

func (p *Publisher) processFullResults(first watchable.Update[string, *fullPublication]) {
	batch := []watchable.Update[string, *fullPublication]{first}
	for remaining := len(p.fullResults); remaining > 0; remaining-- {
		batch = append(batch, <-p.fullResults)
	}
	for _, update := range message.CoalesceUpdates(p.log, batch) {
		if update.Value.snapshot == nil {
			p.processDelete(update.Key, update.Value.ctx)
		} else {
			p.processFull(update.Key, update.Value)
		}
	}
}

func (p *Publisher) processFull(key string, publication *fullPublication) {
	// Published endpoints win over the full result's when their revision is newer
	// (rule 3), so a full build that started before a fast update can't roll it back.
	contexts, replacements, err := translator.ReplayPublishedEndpoints(publication.endpointContexts, p.publishedEndpoints.get(key))
	snapshot := publication.snapshot
	if err == nil {
		snapshot, err = cache.ReplaceEndpoints(publication.snapshot, replacements)
	}
	if err != nil {
		// Unlikely, mostly a guard against code bugs. Publish the full result as
		// built rather than drop it, so its other changes still go out.
		p.log.Error(err, "failed to replay published endpoints, publishing the full result as built", "ir-key", key)
		snapshot, contexts = publication.snapshot, publication.endpointContexts
	}
	p.snapshots[key] = snapshot
	p.publishedEndpoints.replaceContexts(key, contexts)
	p.publish(publication.ctx, pathFull, key, snapshot)
}

func (p *Publisher) processDelete(key string, ctx context.Context) {
	p.forget(key)
	p.publish(ctx, pathDelete, key, nil)
}

// forget drops a key's snapshot and published endpoint contexts.
func (p *Publisher) forget(key string) {
	p.publishedEndpoints.forget(key)
	delete(p.snapshots, key)
}

func (p *Publisher) publish(ctx context.Context, path, key string, snapshot *cachev3.Snapshot) {
	// The xDS cache stores the snapshot before it can fail, so later Envoy
	// requests still get it. Its errors are deterministic, so a retry would
	// fail the same way.
	if err := p.publishSnapshot(key, snapshot, ctx); err != nil {
		p.log.Error(err, "failed to publish snapshot", "ir-key", key)
		snapshotPublications.WithFailure(metrics.ReasonError, publicationPath.Value(path)).Increment()
		return
	}
	snapshotPublications.WithSuccess(publicationPath.Value(path)).Increment()
}
