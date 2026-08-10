// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/client-go/util/workqueue"
	clocktesting "k8s.io/utils/clock/testing"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func TestReconcileCoalesceWindowFor(t *testing.T) {
	t.Run("unset falls back to the default", func(t *testing.T) {
		d, err := reconcileCoalesceWindowFor(egv1a1.DefaultEnvoyGateway())
		require.NoError(t, err)
		require.Equal(t, defaultReconcileCoalesceWindow, d)
	})

	t.Run("configured value is parsed and used", func(t *testing.T) {
		eg := egv1a1.DefaultEnvoyGateway()
		eg.Provider.Kubernetes.ReconcileCoalesceWindow = new(gwapiv1.Duration("250ms"))
		d, err := reconcileCoalesceWindowFor(eg)
		require.NoError(t, err)
		require.Equal(t, 250*time.Millisecond, d)
	})

	t.Run("invalid duration is rejected", func(t *testing.T) {
		eg := egv1a1.DefaultEnvoyGateway()
		eg.Provider.Kubernetes.ReconcileCoalesceWindow = new(gwapiv1.Duration("not-a-duration"))
		_, err := reconcileCoalesceWindowFor(eg)
		require.Error(t, err)
	})
}

// newTestCoalescingQueue builds a coalescingQueue backed by a fake clock, so the test can
// deterministically control when the coalesce window elapses instead of sleeping.
func newTestCoalescingQueue(t *testing.T, fakeClock *clocktesting.FakeClock) *coalescingQueue {
	t.Helper()
	base := workqueue.NewTypedRateLimitingQueueWithConfig(
		workqueue.DefaultTypedControllerRateLimiter[reconcile.Request](),
		workqueue.TypedRateLimitingQueueConfig[reconcile.Request]{Clock: fakeClock},
	)
	t.Cleanup(base.ShutDown)
	return &coalescingQueue{TypedRateLimitingInterface: base, window: defaultReconcileCoalesceWindow}
}

// getWithTimeout calls Get() in a goroutine and fails the test if no item shows up within d,
// instead of hanging forever when the coalescing behavior under test regresses.
func getWithTimeout(t *testing.T, q *coalescingQueue, d time.Duration) (reconcile.Request, bool) {
	t.Helper()
	type result struct {
		item     reconcile.Request
		shutdown bool
	}
	ch := make(chan result, 1)
	go func() {
		item, shutdown := q.Get()
		ch <- result{item, shutdown}
	}()
	select {
	case r := <-ch:
		return r.item, r.shutdown
	case <-time.After(d):
		return reconcile.Request{}, false
	}
}

func TestCoalescingQueue(t *testing.T) {
	req := reconcile.Request{}

	t.Run("repeated Add within the window collapses into a single Reconcile", func(t *testing.T) {
		fakeClock := clocktesting.NewFakeClock(time.Now())
		q := newTestCoalescingQueue(t, fakeClock)

		// Simulate a burst of watch events for the same key, spread across the window.
		for range 5 {
			q.Add(req)
			fakeClock.Step(defaultReconcileCoalesceWindow / 10)
		}
		require.Equal(t, 0, q.Len(), "item must not be visible to a worker before the window elapses")

		// Advance past the window; only the first Add()'s (earliest) deadline should fire.
		fakeClock.Step(defaultReconcileCoalesceWindow)
		require.Eventually(t, func() bool { return q.Len() == 1 }, time.Second, time.Millisecond)

		item, shutdown := getWithTimeout(t, q, time.Second)
		require.False(t, shutdown)
		require.Equal(t, req, item)
		q.Done(item)

		// No second item should ever surface for this burst.
		require.Equal(t, 0, q.Len())
	})

	t.Run("Add after the window schedules a fresh Reconcile", func(t *testing.T) {
		fakeClock := clocktesting.NewFakeClock(time.Now())
		q := newTestCoalescingQueue(t, fakeClock)

		q.Add(req)
		fakeClock.Step(defaultReconcileCoalesceWindow)
		item, shutdown := getWithTimeout(t, q, time.Second)
		require.False(t, shutdown)
		q.Done(item)

		// A later, independent burst gets its own Reconcile.
		q.Add(req)
		fakeClock.Step(defaultReconcileCoalesceWindow)
		_, shutdown = getWithTimeout(t, q, time.Second)
		require.False(t, shutdown)
	})

	t.Run("AddRateLimited is not redirected through the coalesce window", func(t *testing.T) {
		fakeClock := clocktesting.NewFakeClock(time.Now())
		q := newTestCoalescingQueue(t, fakeClock)

		q.AddRateLimited(req)
		// DefaultTypedControllerRateLimiter's base delay is 5ms, well under the coalesce window;
		// if AddRateLimited were being redirected to the window, this would still be empty.
		fakeClock.Step(10 * time.Millisecond)
		require.Eventually(t, func() bool { return q.Len() == 1 }, time.Second, time.Millisecond)
	})
}
