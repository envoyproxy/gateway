// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// fakeQueue records what the debouncer forwards, standing in for the real workqueue.
type fakeQueue struct {
	workqueue.TypedRateLimitingInterface[reconcile.Request]

	mu    sync.Mutex
	added []reconcile.Request
	shut  bool
}

func (q *fakeQueue) Add(item reconcile.Request) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.added = append(q.added, item)
}

func (q *fakeQueue) ShutDown() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.shut = true
}

func (q *fakeQueue) adds() []reconcile.Request {
	q.mu.Lock()
	defer q.mu.Unlock()
	return append([]reconcile.Request(nil), q.added...)
}

// fakeClock drives the debouncer without sleeping. Timers are collected rather than
// scheduled, and advance fires the ones that have come due.
type fakeClock struct {
	mu     sync.Mutex
	now    time.Time
	timers []*fakeTimer
}

type fakeTimer struct {
	at time.Time
	fn func()
}

func newFakeClock() *fakeClock {
	return &fakeClock{now: time.Date(2026, 8, 19, 0, 0, 0, 0, time.UTC)}
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeClock) AfterFunc(d time.Duration, fn func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.timers = append(c.timers, &fakeTimer{at: c.now.Add(d), fn: fn})
}

// advance moves the clock forward and runs every timer that came due, repeating
// until none remain so that a timer which re-arms itself at the same instant still
// settles.
func (c *fakeClock) advance(d time.Duration) {
	c.mu.Lock()
	c.now = c.now.Add(d)
	c.mu.Unlock()

	for {
		c.mu.Lock()
		var due []*fakeTimer
		remaining := c.timers[:0]
		for _, t := range c.timers {
			if !t.at.After(c.now) {
				due = append(due, t)
				continue
			}
			remaining = append(remaining, t)
		}
		c.timers = remaining
		c.mu.Unlock()

		if len(due) == 0 {
			return
		}
		for _, t := range due {
			t.fn()
		}
	}
}

func newTestQueue(after, maxHold time.Duration) (*debouncingQueue, *fakeQueue, *fakeClock) {
	inner := &fakeQueue{}
	clock := newFakeClock()
	q := newDebouncingQueue(inner, after, maxHold)
	q.now = clock.Now
	q.afterFunc = clock.AfterFunc
	return q, inner, clock
}

var testRequest = reconcile.Request{NamespacedName: types.NamespacedName{Name: "gatewayclass-controller"}}

func TestDebouncingQueue(t *testing.T) {
	t.Run("holds a single add for the quiet period", func(t *testing.T) {
		q, inner, clock := newTestQueue(100*time.Millisecond, 10*time.Second)

		q.Add(testRequest)
		clock.advance(50 * time.Millisecond)
		require.Empty(t, inner.adds(), "should still be held before the quiet period elapses")

		clock.advance(50 * time.Millisecond)
		require.Equal(t, []reconcile.Request{testRequest}, inner.adds())
	})

	t.Run("coalesces a burst into one reconcile", func(t *testing.T) {
		q, inner, clock := newTestQueue(100*time.Millisecond, 10*time.Second)

		// Ten adds, each arriving before the quiet period elapses, as a burst of
		// EndpointSlice events would.
		for range 10 {
			q.Add(testRequest)
			clock.advance(50 * time.Millisecond)
		}
		require.Empty(t, inner.adds(), "continued adds should keep moving the quiet period")

		clock.advance(100 * time.Millisecond)
		require.Equal(t, []reconcile.Request{testRequest}, inner.adds())
	})

	t.Run("quiet period is measured from the last add", func(t *testing.T) {
		// Re-arming must run to lastAdd+after, not a fresh after from the moment
		// the timer fired, or extending a batch pushes the flush progressively
		// later than the configured quiet period.
		q, inner, clock := newTestQueue(100*time.Millisecond, 10*time.Second)

		q.Add(testRequest)
		clock.advance(50 * time.Millisecond)
		q.Add(testRequest) // quiet period now ends 100ms from here

		clock.advance(99 * time.Millisecond)
		require.Empty(t, inner.adds(), "should not flush before the quiet period ends")

		clock.advance(1 * time.Millisecond)
		require.Equal(t, []reconcile.Request{testRequest}, inner.adds(),
			"should flush exactly 100ms after the last add, not 100ms after the timer fired")
	})

	t.Run("forces a flush at the maximum hold time", func(t *testing.T) {
		q, inner, clock := newTestQueue(100*time.Millisecond, 1*time.Second)

		// Sustained churn: the quiet period never elapses on its own.
		for range 40 {
			q.Add(testRequest)
			clock.advance(50 * time.Millisecond)
		}

		adds := inner.adds()
		require.NotEmpty(t, adds, "max hold should force a flush under sustained churn")
		// 2s of churn against a 1s bound: two flushes, not one per add.
		require.Less(t, len(adds), 4)
	})

	t.Run("max hold is not extended by new adds", func(t *testing.T) {
		q, inner, clock := newTestQueue(500*time.Millisecond, 1*time.Second)

		q.Add(testRequest)
		clock.advance(400 * time.Millisecond)
		q.Add(testRequest)
		clock.advance(400 * time.Millisecond)
		q.Add(testRequest)
		require.Empty(t, inner.adds())

		// 1s after the first add, the batch must flush even though the last add
		// was only 200ms ago and the quiet period has not elapsed.
		clock.advance(200 * time.Millisecond)
		require.Equal(t, []reconcile.Request{testRequest}, inner.adds())
	})

	t.Run("forwards an unexpected second request rather than dropping it", func(t *testing.T) {
		// Every watch enqueues the same GatewayClass request, so only one batch is
		// ever open. Should another enqueue path ever appear, it must lose
		// debouncing rather than lose reconciles.
		q, inner, clock := newTestQueue(100*time.Millisecond, 10*time.Second)
		other := reconcile.Request{NamespacedName: types.NamespacedName{Name: "other"}}

		q.Add(testRequest)
		q.Add(other)
		require.Equal(t, []reconcile.Request{other}, inner.adds(),
			"the request that cannot be batched should go straight through")

		clock.advance(100 * time.Millisecond)
		require.Equal(t, []reconcile.Request{other, testRequest}, inner.adds(),
			"the held batch should still flush on its own schedule")
	})

	t.Run("drains pending requests on shutdown", func(t *testing.T) {
		q, inner, clock := newTestQueue(10*time.Second, 30*time.Second)

		q.Add(testRequest)
		clock.advance(time.Second)
		require.Empty(t, inner.adds(), "still within the quiet period")

		q.ShutDown()
		require.Equal(t, []reconcile.Request{testRequest}, inner.adds(),
			"a requested reconcile must not be dropped by shutdown")
		require.True(t, inner.shut)
	})

	t.Run("passes through after shutdown", func(t *testing.T) {
		q, inner, _ := newTestQueue(10*time.Second, 30*time.Second)

		q.ShutDown()
		q.Add(testRequest)
		require.Equal(t, []reconcile.Request{testRequest}, inner.adds(),
			"adds after shutdown must not be parked on a timer that nobody drains")
	})

	t.Run("requeue paths bypass the debouncer", func(t *testing.T) {
		// AddAfter and AddRateLimited are the error-requeue paths. They must reach
		// the embedded queue's own Add rather than the debounced one, so that
		// backoff is not delayed. A real queue is used here because the bypass is
		// a property of embedding, not of the wrapper's own code.
		inner := workqueue.NewTypedRateLimitingQueue[reconcile.Request](
			workqueue.DefaultTypedControllerRateLimiter[reconcile.Request]())
		q := newDebouncingQueue(inner, time.Hour, time.Hour)
		t.Cleanup(inner.ShutDown)

		q.AddRateLimited(testRequest)

		require.Eventually(t, func() bool { return inner.Len() == 1 }, 5*time.Second, 10*time.Millisecond,
			"AddRateLimited should not be held by the debouncer")
	})
}
