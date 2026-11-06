// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package endpoint

import (
	"sync"

	"github.com/envoyproxy/gateway/internal/ir"
)

// UpdateQueue hands fast-path endpoint updates from the endpoint controller to
// the xDS publisher. It keeps only the newest revision per backend until the
// publisher drains them, and has its own lock so neither side waits on
// Kubernetes reads.
type UpdateQueue struct {
	once    sync.Once
	mu      sync.Mutex
	pending map[ir.ServiceKey]*BackendEndpoints
	ready   chan struct{}
}

func (q *UpdateQueue) init() {
	q.once.Do(func() {
		q.pending = make(map[ir.ServiceKey]*BackendEndpoints)
		q.ready = make(chan struct{}, 1)
	})
}

// Add queues an update, replacing any older pending revision of the same
// backend, and wakes the publisher.
func (q *UpdateQueue) Add(backendEndpoints *BackendEndpoints) {
	q.init()
	q.mu.Lock()
	defer q.mu.Unlock()
	key := backendEndpoints.Backend
	if previous := q.pending[key]; previous != nil && previous.Revision >= backendEndpoints.Revision {
		return
	}
	q.pending[key] = backendEndpoints
	select {
	case q.ready <- struct{}{}:
	default:
	}
}

// Ready signals that updates are pending; take them with Drain.
func (q *UpdateQueue) Ready() <-chan struct{} {
	q.init()
	return q.ready
}

// Drain returns all pending updates and empties the queue.
func (q *UpdateQueue) Drain() map[ir.ServiceKey]*BackendEndpoints {
	q.init()
	q.mu.Lock()
	defer q.mu.Unlock()
	updates := q.pending
	q.pending = make(map[ir.ServiceKey]*BackendEndpoints)
	return updates
}
