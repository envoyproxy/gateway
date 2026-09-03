// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"sync"
	"time"

	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// debouncingQueue wraps the controller's workqueue and holds requests for a quiet
// period so that a burst of resource changes costs a single reconcile.
//
// Every watch in this controller enqueues the same GatewayClass request, so the
// workqueue already collapses events that arrive while a reconcile is in flight: the
// request sits in the dirty set and is requeued exactly once when Done is called.
// What the workqueue does not do is wait. When a reconcile is fast relative to the
// event rate the queue drains between events, and each event then costs a full
// rebuild of the resource tree even though only the resulting state matters.
//
// This wrapper supplies the missing quiet period. A request is held until no new Add
// has arrived for after, bounded by maxHold so that sustained churn cannot defer a
// reconcile indefinitely.
//
// Only Add is debounced. AddAfter and AddRateLimited are the requeue paths taken when
// a reconcile returns an error, and they reach the embedded queue directly, so
// backoff is never slowed down by debouncing.
type debouncingQueue struct {
	workqueue.TypedRateLimitingInterface[reconcile.Request]

	after   time.Duration
	maxHold time.Duration

	// now and afterFunc are indirected so that tests can drive time directly
	// instead of sleeping through the real quiet period.
	now       func() time.Time
	afterFunc func(time.Duration, func())

	mu sync.Mutex
	// pending is the batch currently being held, or nil when nothing is held.
	// Since every watch enqueues the same GatewayClass request, at most one batch
	// is ever open, so a single slot is enough.
	pending *pendingRequest
	drained bool
}

// pendingRequest is a batch of Adds being held for one request.
//
// No timer is stored. Exactly one timer is outstanding while a batch is open, and a
// timer that fires after the batch is gone finds nothing pending and returns, so
// there is never anything worth cancelling.
type pendingRequest struct {
	item reconcile.Request

	// startedAt is when the batch opened, and fixes the maximum hold deadline.
	startedAt time.Time

	// lastAdd is the most recent Add, and is what moves the quiet period forward.
	lastAdd time.Time

	// coalesced counts the Adds merged into this batch, reported on flush.
	coalesced int
}

func newDebouncingQueue(
	inner workqueue.TypedRateLimitingInterface[reconcile.Request],
	after, maxHold time.Duration,
) *debouncingQueue {
	return &debouncingQueue{
		TypedRateLimitingInterface: inner,
		after:                      after,
		maxHold:                    maxHold,
		now:                        time.Now,
		afterFunc:                  func(d time.Duration, fn func()) { time.AfterFunc(d, fn) },
	}
}

// Add holds item until it has been quiet for after, or until the maximum hold time
// has elapsed, whichever comes first.
func (q *debouncingQueue) Add(item reconcile.Request) {
	q.mu.Lock()

	// Once shut down, hand off directly rather than arming a timer that would fire
	// against a queue nobody is reading from.
	if q.drained {
		q.mu.Unlock()
		q.TypedRateLimitingInterface.Add(item)
		return
	}

	now := q.now()
	switch {
	case q.pending == nil:
		q.pending = &pendingRequest{item: item, startedAt: now, lastAdd: now, coalesced: 1}
		q.arm(now)

	case q.pending.item == item:
		// Extend the batch. The timer already running for it re-arms itself from
		// onTimer once it sees the moved quiet period; resetting a timer here
		// would race a callback that has already started.
		q.pending.lastAdd = now
		q.pending.coalesced++

	default:
		// Not reachable while every watch enqueues the same request. Forward
		// directly rather than dropping it or discarding the open batch, so an
		// added enqueue path would lose debouncing rather than lose reconciles.
		q.mu.Unlock()
		q.TypedRateLimitingInterface.Add(item)
		return
	}

	q.mu.Unlock()
}

// arm schedules the next flush check. Callers must hold q.mu with a batch open.
//
// The wait runs to the end of the quiet period measured from the most recent Add,
// not a fresh after from now. Re-arming with a full after would push the flush later
// every time the batch is extended, making the effective quiet period up to twice
// what was configured.
func (q *debouncingQueue) arm(now time.Time) {
	wait := q.pending.lastAdd.Add(q.after).Sub(now)
	if remaining := q.pending.startedAt.Add(q.maxHold).Sub(now); remaining < wait {
		wait = remaining
	}
	if wait < 0 {
		wait = 0
	}
	q.afterFunc(wait, q.onTimer)
}

// onTimer forwards the batch once the quiet period has elapsed or the maximum hold
// time has been reached, and otherwise re-arms because newer Adds moved the quiet
// period.
func (q *debouncingQueue) onTimer() {
	q.mu.Lock()

	p := q.pending
	if p == nil {
		q.mu.Unlock()
		return
	}

	now := q.now()
	deadlineReached := !p.startedAt.Add(q.maxHold).After(now)
	if !deadlineReached && p.lastAdd.Add(q.after).After(now) {
		q.arm(now)
		q.mu.Unlock()
		return
	}

	q.pending = nil
	q.mu.Unlock()

	reason := debounceReasonQuiet
	if deadlineReached {
		reason = debounceReasonMax
	}
	q.forward(p, now, reason)
}

// forward records the batch and hands the request to the embedded queue.
func (q *debouncingQueue) forward(p *pendingRequest, now time.Time, reason string) {
	reconcileDebouncePending.Record(float64(p.coalesced))
	reconcileDebounceDelaySeconds.Record(now.Sub(p.startedAt).Seconds())
	reconcileDebounceFlushTotal.With(debounceReasonLabel.Value(reason)).Increment()
	q.TypedRateLimitingInterface.Add(p.item)
}

// drainPending forwards a held batch, so that shutting down does not drop a reconcile
// that was already requested. Subsequent Adds bypass debouncing.
func (q *debouncingQueue) drainPending() {
	q.mu.Lock()
	if q.drained {
		q.mu.Unlock()
		return
	}
	q.drained = true

	p, now := q.pending, q.now()
	q.pending = nil
	q.mu.Unlock()

	if p != nil {
		q.forward(p, now, debounceReasonClose)
	}
}

func (q *debouncingQueue) ShutDown() {
	q.drainPending()
	q.TypedRateLimitingInterface.ShutDown()
}

func (q *debouncingQueue) ShutDownWithDrain() {
	q.drainPending()
	q.TypedRateLimitingInterface.ShutDownWithDrain()
}
