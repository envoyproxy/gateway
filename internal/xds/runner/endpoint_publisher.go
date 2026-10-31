// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"context"
	"maps"
	"sync"
	"time"

	cachev3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/telepresenceio/watchable"

	"github.com/envoyproxy/gateway/internal/endpoint"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/xds/cache"
	"github.com/envoyproxy/gateway/internal/xds/translator"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

type fullPublication struct {
	snapshot *cachev3.Snapshot
	contexts types.EndpointContexts
	ctx      context.Context
}

type endpointPublishers struct {
	ctx      context.Context
	source   *endpoint.Source
	install  func(string, *cachev3.Snapshot, context.Context) error
	log      logging.Logger
	mu       sync.Mutex
	keys     map[string]*endpointPublisher
	index    map[ir.EndpointBackend]map[string]*endpointPublisher
	stopping bool
	done     sync.WaitGroup
}

func newEndpointPublishers(ctx context.Context, source *endpoint.Source, install func(string, *cachev3.Snapshot, context.Context) error, log logging.Logger) *endpointPublishers {
	return &endpointPublishers{
		ctx: ctx, source: source, install: install, log: log,
		keys: make(map[string]*endpointPublisher), index: make(map[ir.EndpointBackend]map[string]*endpointPublisher),
	}
}

func (p *endpointPublishers) run() {
	defer func() {
		// Synchronize with submit before waiting; it cannot create an owner once
		// cancellation is visible under the routing lock.
		p.mu.Lock()
		p.stopping = true
		p.mu.Unlock()
		p.done.Wait()
	}()
	for {
		select {
		case <-p.ctx.Done():
			return
		case <-p.source.Updates():
			updates := p.source.Drain()
			p.mu.Lock()
			for _, update := range updates {
				for _, publisher := range p.index[update.Backend] {
					publisher.offer(update)
				}
			}
			p.mu.Unlock()
		}
	}
}

func (p *endpointPublishers) submit(key string, publication *fullPublication) error {
	for {
		p.mu.Lock()
		if err := p.ctx.Err(); err != nil {
			p.mu.Unlock()
			return err
		}
		if p.stopping {
			p.mu.Unlock()
			return context.Canceled
		}
		publisher := p.keys[key]
		if publisher == nil {
			publisher = &endpointPublisher{
				pool: p, key: key,
				full:  make(chan watchable.Update[string, *fullPublication], 2),
				space: make(chan struct{}, 1), wake: make(chan struct{}, 1), stopped: make(chan struct{}),
				pending: make(map[ir.EndpointBackend]*endpoint.Snapshot),
			}
			p.keys[key] = publisher
			p.done.Go(publisher.run)
		}
		select {
		case publisher.full <- watchable.Update[string, *fullPublication]{Key: key, Value: publication, Delete: publication.snapshot == nil}:
			p.mu.Unlock()
			return nil
		default:
			p.mu.Unlock()
		}
		select {
		case <-p.ctx.Done():
			return p.ctx.Err()
		case <-publisher.stopped:
		case <-publisher.space:
		}
	}
}

func (p *endpointPublishers) updateIndex(publisher *endpointPublisher, contexts types.EndpointContexts) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.keys[publisher.key] != publisher {
		return
	}
	for backend := range publisher.backends {
		delete(p.index[backend], publisher.key)
		if len(p.index[backend]) == 0 {
			delete(p.index, backend)
		}
	}
	publisher.backends = make(map[ir.EndpointBackend]struct{})
	publisher.clusters = make(map[ir.EndpointBackend]map[string]struct{})
	var keys []endpoint.Key
	for name, context := range contexts {
		for _, setting := range context.Settings {
			if setting.EndpointSource == nil {
				continue
			}
			backend := setting.EndpointSource.Backend
			if publisher.clusters[backend] == nil {
				publisher.clusters[backend] = make(map[string]struct{})
			}
			publisher.clusters[backend][name] = struct{}{}
			if _, found := publisher.backends[backend]; found {
				continue
			}
			publisher.backends[backend] = struct{}{}
			keys = append(keys, endpoint.BackendKey(backend))
			if p.index[backend] == nil {
				p.index[backend] = make(map[string]*endpointPublisher)
			}
			p.index[backend][publisher.key] = publisher
		}
	}
	p.source.SetReferences("publisher/"+publisher.key, keys)
}

// retire runs after deletion is installed. submit holds the same routing lock
// while enqueueing, so a recreated key cannot lose a result to a retiring owner.
func (p *endpointPublishers) retire(publisher *endpointPublisher) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(publisher.full) != 0 {
		return false
	}
	delete(p.keys, publisher.key)
	return true
}

type endpointPublisher struct {
	pool    *endpointPublishers
	key     string
	full    chan watchable.Update[string, *fullPublication]
	space   chan struct{}
	wake    chan struct{}
	stopped chan struct{}
	// Only the input mailbox is shared. Refresh and committed state are owned
	// exclusively by run, with no mutex held during endpoint rebuilding.
	inboxMu sync.Mutex
	pending map[ir.EndpointBackend]*endpoint.Snapshot

	snapshot *cachev3.Snapshot
	contexts types.EndpointContexts
	backends map[ir.EndpointBackend]struct{}
	clusters map[ir.EndpointBackend]map[string]struct{}
	deleted  bool
	retry    bool
}

func (p *endpointPublisher) offer(update *endpoint.Snapshot) {
	p.inboxMu.Lock()
	if previous := p.pending[update.Backend]; previous == nil || previous.Revision < update.Revision {
		p.pending[update.Backend] = update
	}
	p.inboxMu.Unlock()
	select {
	case p.wake <- struct{}{}:
	default:
	}
}

func (p *endpointPublisher) drain() map[ir.EndpointBackend]*endpoint.Snapshot {
	p.inboxMu.Lock()
	defer p.inboxMu.Unlock()
	updates := p.pending
	p.pending = make(map[ir.EndpointBackend]*endpoint.Snapshot)
	return updates
}

func (p *endpointPublisher) run() {
	defer close(p.stopped)
	defer p.pool.updateIndex(p, nil)
	timer := time.NewTimer(time.Hour)
	timer.Stop()
	defer timer.Stop()
	var retryTick <-chan time.Time
	for {
		select {
		case <-p.pool.ctx.Done():
			return
		case update := <-p.full:
			p.processFullBatch(update)
			// Give pending endpoints a turn even during continuous full results.
			select {
			case <-p.wake:
				p.processEndpoints()
			default:
			}
		case <-p.wake:
			p.processEndpoints()
			// Endpoint churn cannot starve full results and deletion.
			select {
			case update := <-p.full:
				p.processFullBatch(update)
			default:
			}
		case <-retryTick:
			retryTick = nil
			if p.retry {
				endpointDeliveryRetries.Increment()
				p.publish(p.pool.ctx)
			}
		}
		if p.retry && retryTick == nil {
			timer.Reset(time.Second)
			retryTick = timer.C
		} else if !p.retry && retryTick != nil {
			timer.Stop()
			retryTick = nil
		}
		if p.deleted && !p.retry && p.pool.retire(p) {
			return
		}
	}
}

func (p *endpointPublisher) processFullBatch(first watchable.Update[string, *fullPublication]) {
	batch := []watchable.Update[string, *fullPublication]{first}
	for remaining := len(p.full); remaining > 0; remaining-- {
		batch = append(batch, <-p.full)
	}
	select {
	case p.space <- struct{}{}:
	default:
	}
	for _, update := range message.CoalesceUpdatesPreservingDeletes(p.pool.log, batch) {
		publication := update.Value
		if update.Delete {
			p.snapshot, p.contexts, p.deleted = nil, nil, true
			p.pool.updateIndex(p, nil)
			endpointPublications.With(publicationPath.Value("delete")).Increment()
			p.publish(publication.ctx)
			continue
		}
		contexts, replacements, err := translator.ReplayEndpoints(publication.contexts, p.contexts)
		if err != nil {
			p.pool.log.Error(err, "failed to refresh full publication", "ir-key", p.key)
			continue
		}
		snapshot, err := cache.ReplaceEndpoints(publication.snapshot, replacements)
		if err != nil {
			p.pool.log.Error(err, "failed to prepare full publication", "ir-key", p.key)
			continue
		}
		p.snapshot, p.contexts, p.deleted = snapshot, contexts, false
		p.pool.updateIndex(p, contexts)
		endpointPublications.With(publicationPath.Value("full")).Increment()
		p.publish(publication.ctx)
	}
}

func (p *endpointPublisher) processEndpoints() {
	updates := p.drain()
	if p.snapshot == nil || len(updates) == 0 {
		return
	}
	start := time.Now()
	defer func() { endpointRefreshDuration.Record(time.Since(start).Seconds()) }()
	affected := make(types.EndpointContexts)
	for backend := range updates {
		for name := range p.clusters[backend] {
			affected[name] = p.contexts[name]
		}
	}
	contexts, replacements, err := translator.RefreshEndpoints(affected, updates)
	if err != nil {
		p.pool.log.Error(err, "failed to refresh endpoints", "ir-key", p.key)
		return
	}
	snapshot, err := cache.ReplaceEndpoints(p.snapshot, replacements)
	if err != nil {
		p.pool.log.Error(err, "failed to prepare endpoint publication", "ir-key", p.key)
		return
	}
	changed := snapshot != p.snapshot
	if len(replacements) != 0 {
		next := maps.Clone(p.contexts)
		maps.Copy(next, contexts)
		p.snapshot, p.contexts = snapshot, next
	}
	if changed {
		endpointPublications.With(publicationPath.Value("fast")).Increment()
		p.publish(p.pool.ctx)
	}
}

func (p *endpointPublisher) publish(ctx context.Context) {
	// Desired state is already committed. A delivery failure retries that current
	// state, including deletion, rather than retaining an older snapshot to resend.
	err := p.pool.install(p.key, p.snapshot, ctx)
	p.retry = err != nil
	if err != nil {
		p.pool.log.Error(err, "failed to deliver desired snapshot; will retry", "ir-key", p.key)
	}
}
