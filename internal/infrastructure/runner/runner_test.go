// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"sync"
	"syscall"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/message"
)

// retryWithin bounds, in the bubble's virtual time, how long a failed
// CreateOrUpdateProxyInfra call may wait before it is retried.
const retryWithin = time.Minute

// webhookErr is the error reported in
// https://github.com/envoyproxy/gateway/issues/10000, wrapped as
// kubernetes.Infra wraps the errors of the resources it applies.
var webhookErr = wrapApplyErr(kerrors.NewInternalError(errors.New(`failed calling webhook "mservice.elbv2.k8s.aws": ` +
	`tls: failed to verify certificate: x509: certificate signed by unknown authority`)))

func wrapApplyErr(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("failed to create or update service envoy-gateway-system/envoy-default-eg: %w", err)
}

// fakeManager is an infrastructure.Manager that records every
// CreateOrUpdateProxyInfra and DeleteProxyInfra call. The nth
// CreateOrUpdateProxyInfra call returns createErr(ctx, infra, n) if createErr
// is set, and otherwise the first calls fail with createErrs in turn. If lock
// is set, it also counts the calls made while lock is not held.
type fakeManager struct {
	createErr func(ctx context.Context, infra *ir.Infra, n int) error
	lock      *sync.Mutex

	mu            sync.Mutex
	createErrs    []error
	creates       []*ir.Infra
	createdAt     []time.Time
	deletes       []*ir.Infra
	inFlight      int
	maxInFlight   int
	unlockedCalls int
}

// enter records the start of a proxy infra call, and the func it returns
// records its end.
func (m *fakeManager) enter() func() {
	// TryLock fails while the runner holds lock, as it must during the call.
	unlocked := m.lock != nil && m.lock.TryLock()
	if unlocked {
		m.lock.Unlock()
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if unlocked {
		m.unlockedCalls++
	}
	m.inFlight++
	m.maxInFlight = max(m.maxInFlight, m.inFlight)
	return func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		m.inFlight--
	}
}

func (m *fakeManager) CreateOrUpdateProxyInfra(ctx context.Context, infra *ir.Infra) error {
	exit := m.enter()
	defer exit()

	m.mu.Lock()
	n := len(m.creates)
	m.creates = append(m.creates, infra)
	m.createdAt = append(m.createdAt, time.Now())
	var err error
	if len(m.createErrs) > 0 {
		err, m.createErrs = m.createErrs[0], m.createErrs[1:]
	}
	m.mu.Unlock()

	if m.createErr != nil {
		return m.createErr(ctx, infra, n)
	}
	return err
}

func (m *fakeManager) DeleteProxyInfra(_ context.Context, infra *ir.Infra) error {
	exit := m.enter()
	defer exit()

	m.mu.Lock()
	defer m.mu.Unlock()
	m.deletes = append(m.deletes, infra)
	return nil
}

func (m *fakeManager) CreateOrUpdateRateLimitInfra(context.Context) error { return nil }
func (m *fakeManager) DeleteRateLimitInfra(context.Context) error         { return nil }

// Close fails if a proxy infra call is still running, as one would be if
// Runner.Close did not wait for the goroutines that make them.
func (m *fakeManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.inFlight > 0 {
		return errors.New("manager closed during a proxy infra call")
	}
	return nil
}

func (m *fakeManager) createCalls() []*ir.Infra {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]*ir.Infra(nil), m.creates...)
}

func (m *fakeManager) deleteCalls() []*ir.Infra {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]*ir.Infra(nil), m.deletes...)
}

// createOffsets returns how long after start each CreateOrUpdateProxyInfra
// call for the proxy named name was made.
func (m *fakeManager) createOffsets(start time.Time, name string) []time.Duration {
	m.mu.Lock()
	defer m.mu.Unlock()
	offsets := make([]time.Duration, 0, len(m.creates))
	for i, infra := range m.creates {
		if infra.Proxy.Name == name {
			offsets = append(offsets, m.createdAt[i].Sub(start))
		}
	}
	return offsets
}

func seconds(secs ...int) []time.Duration {
	ds := make([]time.Duration, 0, len(secs))
	for _, s := range secs {
		ds = append(ds, time.Duration(s)*time.Second)
	}
	return ds
}

// startTestRunner subscribes a Runner backed by mgr to a new InfraIR map, as
// Start does once the Manager exists. When the test ends, it stops the runner
// and checks that mgr was only called with r.mu held, never concurrently.
func startTestRunner(t *testing.T, mgr *fakeManager) (*Runner, context.CancelFunc) {
	t.Helper()
	ctx, cancel := context.WithCancel(t.Context())
	r := New(&Config{
		Server: config.Server{
			Logger: logging.DefaultLogger(t.Output(), egv1a1.LogLevelInfo),
		},
		InfraIR: new(message.InfraIR),
	})
	r.mgr = mgr
	mgr.lock = &r.mu
	sub := r.InfraIR.Subscribe(ctx)
	r.done.Go(func() { r.updateProxyInfraFromSubscription(ctx, sub) })
	t.Cleanup(func() {
		stopRunner(t, r, cancel)
		mgr.mu.Lock()
		defer mgr.mu.Unlock()
		require.Zero(t, mgr.unlockedCalls, "the runner called the manager without holding r.mu")
		require.LessOrEqual(t, mgr.maxInFlight, 1, "the runner called the manager concurrently")
	})
	return r, cancel
}

// stopRunner cancels the context of r and fails the test unless r.Close
// returns within a minute. Without that bound, a goroutine that never exits
// would hang the test instead of failing it, because the heartbeat of the
// retry queue keeps the bubble's clock running.
func stopRunner(t *testing.T, r *Runner, cancel context.CancelFunc) {
	t.Helper()
	cancel()
	closed := make(chan error, 1)
	go func() { closed <- r.Close() }()
	select {
	case err := <-closed:
		if err != nil {
			t.Error(err)
		}
	case <-time.After(time.Minute):
		t.Error("Runner.Close did not return within a minute of its context being canceled")
	}
}

// requireNoRetry asserts that r neither holds an Infra IR to retry for key
// nor tracks failures of it. It must be called while r is idle, such as after
// synctest.Wait.
func requireNoRetry(t *testing.T, r *Runner, key string) {
	t.Helper()
	if !r.mu.TryLock() {
		t.Fatal("r.mu is held while the runner should be idle")
	}
	defer r.mu.Unlock()
	require.NotContains(t, r.failed, key)
	require.Zero(t, r.retries.NumRequeues(key))
}

func testInfra(name string) *ir.Infra {
	return &ir.Infra{
		Proxy: &ir.ProxyInfra{
			Name:      name,
			Namespace: "envoy-gateway-system",
			Listeners: []*ir.ProxyListener{{
				Name:  name + "/http",
				Ports: []ir.ListenerPort{{Name: "http-80", Protocol: ir.HTTPProtocolType, ServicePort: 80, ContainerPort: 10080}},
			}},
		},
	}
}

// TestCreateOrUpdateProxyInfraRetriesTransientError reproduces
// https://github.com/envoyproxy/gateway/issues/10000: a CreateOrUpdateProxyInfra
// call that fails because the API server could not call an admission webhook
// must be retried with the latest Infra IR, without waiting for the IR to change.
func TestCreateOrUpdateProxyInfraRetriesTransientError(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		mgr := &fakeManager{createErrs: []error{webhookErr}}
		r, _ := startTestRunner(t, mgr)

		want := testInfra("default/eg")
		r.InfraIR.Store("default/eg", &message.InfraIRWithContext{Infra: want})
		synctest.Wait()
		require.Len(t, mgr.createCalls(), 1)

		time.Sleep(retryWithin)
		synctest.Wait()
		calls := mgr.createCalls()
		require.Lenf(t, calls, 2, "a CreateOrUpdateProxyInfra call that failed with %q was not retried within %v", webhookErr, retryWithin)
		require.Equal(t, want, calls[1])
		// The handler's Infra IR is shared with the watchable map, which reads
		// it concurrently, so the retry must not hand the same one to the Manager.
		require.NotSame(t, calls[0], calls[1])
	})
}

// TestRetryBacksOffUntilWebhookRecovers replays the timeline of
// https://github.com/envoyproxy/gateway/issues/10000, where the webhook
// recovered 283s after the first failure. Retries back off from 1s to once a
// minute, the first one after the recovery succeeds, and none follow it.
func TestRetryBacksOffUntilWebhookRecovers(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		start := time.Now()
		mgr := &fakeManager{createErr: func(context.Context, *ir.Infra, int) error {
			if time.Since(start) < 283*time.Second {
				return webhookErr
			}
			return nil
		}}
		r, _ := startTestRunner(t, mgr)

		r.InfraIR.Store("default/eg", &message.InfraIRWithContext{Infra: testInfra("default/eg")})
		time.Sleep(10 * time.Minute)
		synctest.Wait()
		require.Equal(t, seconds(0, 1, 3, 7, 15, 31, 63, 123, 183, 243, 303), mgr.createOffsets(start, "default/eg"))
		requireNoRetry(t, r, "default/eg")
	})
}

// TestRetryAppliesOnlyNewestInfraIR checks that an update stored while a
// retry is pending supersedes the Infra IR that the retry holds, so an older
// Infra IR is never applied after a newer one, and that a retry of the newer
// one waits for its own backoff.
func TestRetryAppliesOnlyNewestInfraIR(t *testing.T) {
	older := testInfra("default/eg")
	newer := testInfra("default/eg")
	newer.Proxy.Listeners[0].Ports[0].ServicePort = 8080
	noListeners := testInfra("default/eg")
	noListeners.Proxy.Listeners = nil

	testCases := []struct {
		name        string
		newer       *ir.Infra
		createErrs  []error
		want        []*ir.Infra
		wantOffsets []time.Duration
	}{
		{
			name:        "newer update succeeds",
			newer:       newer,
			createErrs:  []error{webhookErr},
			want:        []*ir.Infra{older, newer},
			wantOffsets: []time.Duration{0, 500 * time.Millisecond},
		},
		{
			// The retry queue still holds the key for the older Infra IR, due
			// at 1s, but the newer one is retried 1s after it failed.
			name:        "newer update fails too",
			newer:       newer,
			createErrs:  []error{webhookErr, webhookErr},
			want:        []*ir.Infra{older, newer, newer},
			wantOffsets: []time.Duration{0, 500 * time.Millisecond, 1500 * time.Millisecond},
		},
		{
			name:        "newer update has no listeners",
			newer:       noListeners,
			createErrs:  []error{webhookErr},
			want:        []*ir.Infra{older},
			wantOffsets: []time.Duration{0},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				start := time.Now()
				mgr := &fakeManager{createErrs: tc.createErrs}
				r, _ := startTestRunner(t, mgr)

				r.InfraIR.Store("default/eg", &message.InfraIRWithContext{Infra: older})
				time.Sleep(500 * time.Millisecond)
				r.InfraIR.Store("default/eg", &message.InfraIRWithContext{Infra: tc.newer})
				time.Sleep(10 * time.Minute)
				synctest.Wait()
				require.Equal(t, tc.want, mgr.createCalls())
				require.Equal(t, tc.wantOffsets, mgr.createOffsets(start, "default/eg"))
				requireNoRetry(t, r, "default/eg")
			})
		})
	}
}

// TestUpdateDuringSlowRetry checks an update stored while a retry is being
// applied, when each failing call takes as long as a webhook timeout: the
// update is applied as soon as that retry is done, an older Infra IR is never
// applied after a newer one, and each retry waits for its backoff after the
// failure it retries. The retry stores the update itself, as the last thing it
// does, because under synctest time cannot pass while a goroutine waits for
// r.mu.
func TestUpdateDuringSlowRetry(t *testing.T) {
	older := testInfra("default/eg")
	newer := testInfra("default/eg")
	newer.Proxy.Listeners[0].Ports[0].ServicePort = 8080
	other := testInfra("default/b")

	testCases := []struct {
		name string
		// stored is stored at the end of the first retry of older, and its
		// first storedFails applies fail.
		stored      *ir.Infra
		storedFails int
		wantCreates []*ir.Infra
		wantOffsets map[string][]time.Duration
	}{
		{
			name:        "newer update succeeds",
			stored:      newer,
			wantCreates: []*ir.Infra{older, older, newer},
			wantOffsets: map[string][]time.Duration{"default/eg": seconds(0, 11, 21)},
		},
		{
			// The newer update restarts the backoff, so it is retried 1s after
			// it fails, before the retry of older that was due at 23s.
			name:        "newer update fails once",
			stored:      newer,
			storedFails: 1,
			wantCreates: []*ir.Infra{older, older, newer, newer},
			wantOffsets: map[string][]time.Duration{"default/eg": seconds(0, 11, 21, 22)},
		},
		{
			name:        "update of another key",
			stored:      other,
			wantCreates: []*ir.Infra{older, older, other, older, older, older, older},
			wantOffsets: map[string][]time.Duration{
				"default/eg": seconds(0, 11, 23, 27, 35, 51),
				"default/b":  seconds(21),
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				start := time.Now()
				storedFails := tc.storedFails
				var r *Runner
				mgr := &fakeManager{createErr: func(_ context.Context, infra *ir.Infra, n int) error {
					switch {
					case n <= 1: // the first apply of older and its first retry
						time.Sleep(10 * time.Second)
						if n == 1 {
							r.InfraIR.Store(tc.stored.Proxy.Name, &message.InfraIRWithContext{Infra: tc.stored})
						}
						return webhookErr
					case infra.Proxy.Name != tc.stored.Proxy.Name:
						return webhookErr
					case storedFails > 0:
						storedFails--
						return webhookErr
					}
					return nil
				}}
				r, _ = startTestRunner(t, mgr)

				r.InfraIR.Store("default/eg", &message.InfraIRWithContext{Infra: older})
				time.Sleep(time.Minute)
				synctest.Wait()
				require.Equal(t, tc.wantCreates, mgr.createCalls())
				for name, want := range tc.wantOffsets {
					require.Equal(t, want, mgr.createOffsets(start, name), name)
				}
				requireNoRetry(t, r, tc.stored.Proxy.Name)
			})
		})
	}
}

// TestDeleteCancelsPendingRetry checks that deleting a key cancels the retry
// pending for it, so its infra is not created again after it is deleted, and
// that a key created again afterwards is retried with its new Infra IR only.
func TestDeleteCancelsPendingRetry(t *testing.T) {
	older := testInfra("default/eg")
	same := testInfra("default/eg")
	changed := testInfra("default/eg")
	changed.Proxy.Listeners[0].Ports[0].ServicePort = 8080

	testCases := []struct {
		name        string
		recreated   *ir.Infra // stored 200ms after the delete, unless nil
		wantOffsets []time.Duration
		wantCreates []*ir.Infra
	}{
		{
			name:        "deleted",
			wantOffsets: seconds(0),
			wantCreates: []*ir.Infra{older},
		},
		{
			name:        "created again with the same Infra IR",
			recreated:   same,
			wantOffsets: []time.Duration{0, 700 * time.Millisecond, 1700 * time.Millisecond},
			wantCreates: []*ir.Infra{older, same, same},
		},
		{
			name:        "created again with a changed Infra IR",
			recreated:   changed,
			wantOffsets: []time.Duration{0, 700 * time.Millisecond, 1700 * time.Millisecond},
			wantCreates: []*ir.Infra{older, changed, changed},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				start := time.Now()
				mgr := &fakeManager{createErrs: []error{webhookErr, webhookErr}}
				r, _ := startTestRunner(t, mgr)

				r.InfraIR.Store("default/eg", &message.InfraIRWithContext{Infra: older})
				time.Sleep(500 * time.Millisecond)
				r.InfraIR.Delete("default/eg")
				if tc.recreated != nil {
					time.Sleep(200 * time.Millisecond)
					r.InfraIR.Store("default/eg", &message.InfraIRWithContext{Infra: tc.recreated})
				}
				time.Sleep(10 * time.Minute)
				synctest.Wait()
				require.Equal(t, tc.wantOffsets, mgr.createOffsets(start, "default/eg"))
				require.Equal(t, tc.wantCreates, mgr.createCalls())
				require.Equal(t, []*ir.Infra{older}, mgr.deleteCalls())
				requireNoRetry(t, r, "default/eg")
			})
		})
	}
}

// TestFailingKeyDoesNotDelayOtherKeys checks that a key whose infra keeps
// failing is retried with capped backoff, at most once a minute, while the
// updates of other keys are applied as soon as they are stored, including at
// the instant the failing key is retried.
func TestFailingKeyDoesNotDelayOtherKeys(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		start := time.Now()
		mgr := &fakeManager{createErr: func(_ context.Context, infra *ir.Infra, _ int) error {
			if infra.Proxy.Name == "default/a" {
				return webhookErr
			}
			return nil
		}}
		r, _ := startTestRunner(t, mgr)

		r.InfraIR.Store("default/a", &message.InfraIRWithContext{Infra: testInfra("default/a")})
		time.Sleep(2 * time.Second)
		r.InfraIR.Store("default/b", &message.InfraIRWithContext{Infra: testInfra("default/b")})
		time.Sleep(61 * time.Second) // default/a is retried at 63s.
		r.InfraIR.Store("default/c", &message.InfraIRWithContext{Infra: testInfra("default/c")})
		time.Sleep(10*time.Minute - 63*time.Second)
		synctest.Wait()
		require.Equal(t, seconds(2), mgr.createOffsets(start, "default/b"))
		require.Equal(t, seconds(63), mgr.createOffsets(start, "default/c"))
		require.Equal(t, seconds(0, 1, 3, 7, 15, 31, 63, 123, 183, 243, 303, 363, 423, 483, 543),
			mgr.createOffsets(start, "default/a"))
		requireNoRetry(t, r, "default/b")
		requireNoRetry(t, r, "default/c")
	})
}

// TestCancelStopsRetries checks that cancelling the runner's context stops
// retries, whether one is pending or in flight, and that Close then returns
// once nothing calls the manager. synctest fails the test if a goroutine, such
// as the retry worker or the retry queue's own, is left blocked.
func TestCancelStopsRetries(t *testing.T) {
	testCases := []struct {
		name      string
		createErr func(ctx context.Context, infra *ir.Infra, n int) error
		cancelAt  time.Duration
		wantCalls int
	}{
		{
			name: "retry pending",
			createErr: func(context.Context, *ir.Infra, int) error {
				return webhookErr
			},
			cancelAt:  500 * time.Millisecond,
			wantCalls: 1,
		},
		{
			name: "retry in flight",
			createErr: func(ctx context.Context, _ *ir.Infra, n int) error {
				if n == 0 {
					return webhookErr
				}
				<-ctx.Done()
				time.Sleep(time.Second) // the request takes a while to unwind
				return ctx.Err()
			},
			cancelAt:  2 * time.Second,
			wantCalls: 2,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				mgr := &fakeManager{createErr: tc.createErr}
				r, cancel := startTestRunner(t, mgr)

				r.InfraIR.Store("default/eg", &message.InfraIRWithContext{Infra: testInfra("default/eg")})
				time.Sleep(tc.cancelAt)
				synctest.Wait()
				stopRunner(t, r, cancel)
				time.Sleep(10 * time.Minute)
				synctest.Wait()
				require.Len(t, mgr.createCalls(), tc.wantCalls)
			})
		})
	}
}

// TestRetryKeySkipsCanceledContext checks that the retry worker does not apply
// a key it is handed once the runner's context is canceled, such as a key that
// was waiting for the handler to release r.mu when the runner stopped. It calls
// retryKey directly, because under synctest a goroutine waiting for a
// sync.Mutex is not durably blocked, so the test cannot hold the worker there.
// For the same reason, it cannot tell whether the check runs after r.mu is
// acquired.
func TestRetryKeySkipsCanceledContext(t *testing.T) {
	testCases := []struct {
		name      string
		canceled  bool
		wantCalls int
	}{
		{name: "running", wantCalls: 1},
		{name: "canceled", canceled: true, wantCalls: 0},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mgr := &fakeManager{}
			r := New(&Config{Server: config.Server{Logger: logging.DefaultLogger(t.Output(), egv1a1.LogLevelInfo)}})
			r.mgr = mgr
			mgr.lock = &r.mu
			r.failed = map[string]*failedInfra{"default/eg": {infra: testInfra("default/eg")}}
			r.backoff = workqueue.NewTypedItemExponentialFailureRateLimiter[string](time.Second, time.Minute)
			r.retries = workqueue.NewTypedRateLimitingQueue(r.backoff)
			t.Cleanup(r.retries.ShutDown)
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			if tc.canceled {
				cancel()
			}

			r.retryKey(ctx, "default/eg")
			require.Len(t, mgr.createCalls(), tc.wantCalls)
			require.Zero(t, mgr.unlockedCalls)
		})
	}
}

// TestNonRetryableErrorIsNotRetried checks that errors outside the retry
// policy keep the previous behavior: they are reported once, not retried.
func TestNonRetryableErrorIsNotRetried(t *testing.T) {
	gr := schema.GroupResource{Resource: "services"}
	testCases := []struct {
		name string
		err  error
	}{
		{"invalid", kerrors.NewInvalid(schema.GroupKind{Kind: "Service"}, "envoy-default-eg",
			field.ErrorList{field.Invalid(field.NewPath("spec", "ports"), 0, "must be greater than 0")})},
		{"webhook denial", kerrors.NewForbidden(gr, "envoy-default-eg",
			errors.New(`admission webhook "vservice.elbv2.k8s.aws" denied the request`))},
		{"plain error", errors.New("failed to render the envoy proxy service")},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				mgr := &fakeManager{createErrs: []error{wrapApplyErr(tc.err)}}
				r, _ := startTestRunner(t, mgr)

				r.InfraIR.Store("default/eg", &message.InfraIRWithContext{Infra: testInfra("default/eg")})
				time.Sleep(10 * time.Minute)
				synctest.Wait()
				require.Len(t, mgr.createCalls(), 1)
				requireNoRetry(t, r, "default/eg")
			})
		})
	}
}

// TestRetriesStartOnlyAfterElected checks that, with leader election enabled,
// neither the first attempt nor its retries happen before the runner is elected.
func TestRetriesStartOnlyAfterElected(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		cli := config.NewKubernetesClientHolder()
		cli.Set(fake.NewClientBuilder().Build())
		r := New(&Config{
			Server: config.Server{
				EnvoyGateway: &egv1a1.EnvoyGateway{
					EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{Provider: egv1a1.DefaultEnvoyGatewayProvider()},
				},
				Logger:           logging.DefaultLogger(t.Output(), egv1a1.LogLevelInfo),
				Elected:          make(chan struct{}),
				ProviderReady:    make(chan struct{}),
				KubernetesClient: cli,
			},
			InfraIR: new(message.InfraIR),
		})
		require.NoError(t, r.Start(ctx))
		t.Cleanup(func() { stopRunner(t, r, cancel) })
		// Start's goroutines call the manager only once Elected is closed.
		mgr := &fakeManager{createErrs: []error{webhookErr}}
		r.mgr = mgr
		mgr.lock = &r.mu

		r.InfraIR.Store("default/eg", &message.InfraIRWithContext{Infra: testInfra("default/eg")})
		time.Sleep(10 * time.Minute)
		synctest.Wait()
		require.Empty(t, mgr.createCalls())

		close(r.Elected)
		time.Sleep(retryWithin)
		synctest.Wait()
		require.Len(t, mgr.createCalls(), 2)
	})
}

func TestIsRetryable(t *testing.T) {
	gr := schema.GroupResource{Resource: "services"}
	urlErr := func(err error) error {
		return &url.Error{Op: "Patch", URL: "https://10.96.0.1:443/api/v1/namespaces/envoy-gateway-system/services/envoy-default-eg", Err: err}
	}
	stopped, stop := context.WithCancel(t.Context())
	stop()

	testCases := []struct {
		name string
		ctx  context.Context
		err  error
		want bool
	}{
		{"no error", t.Context(), nil, false},
		{"webhook call failed", t.Context(), webhookErr, true},
		{"too many requests", t.Context(), kerrors.NewTooManyRequests("throttled", 1), true},
		{"service unavailable", t.Context(), kerrors.NewServiceUnavailable("etcd leader changed"), true},
		{"gateway timeout", t.Context(), kerrors.NewTimeoutError("request timed out", 1), true},
		{"server timeout", t.Context(), kerrors.NewServerTimeout(gr, "patch", 1), true},
		{"client deadline exceeded", t.Context(), context.DeadlineExceeded, true},
		{"inner context canceled", t.Context(), context.Canceled, false},
		{"connection refused", t.Context(), urlErr(&net.OpError{Op: "dial", Net: "tcp", Err: os.NewSyscallError("connect", syscall.ECONNREFUSED)}), true},
		{"connection reset", t.Context(), urlErr(&net.OpError{Op: "read", Net: "tcp", Err: os.NewSyscallError("read", syscall.ECONNRESET)}), true},
		{"EOF", t.Context(), urlErr(io.EOF), true},
		{"GOAWAY", t.Context(), urlErr(errors.New("http2: server sent GOAWAY and closed the connection; LastStreamID=3, ErrCode=NO_ERROR, debug=\"\"")), true},
		{"I/O timeout", t.Context(), urlErr(os.ErrDeadlineExceeded), true},
		{"HTTP/2 connection lost", t.Context(), urlErr(errors.New("http2: client connection lost")), true},
		{"no route to host", t.Context(), urlErr(&net.OpError{Op: "dial", Net: "tcp", Err: os.NewSyscallError("connect", syscall.EHOSTUNREACH)}), true},
		{"invalid", t.Context(), kerrors.NewInvalid(schema.GroupKind{Kind: "Service"}, "envoy-default-eg", nil), false},
		{"forbidden", t.Context(), kerrors.NewForbidden(gr, "envoy-default-eg", errors.New("denied")), false},
		{"not found", t.Context(), kerrors.NewNotFound(gr, "envoy-default-eg"), false},
		{"conflict", t.Context(), kerrors.NewConflict(gr, "envoy-default-eg", errors.New("modified")), false},
		{"bad request", t.Context(), kerrors.NewBadRequest("bad request"), false},
		{"plain error", t.Context(), errors.New("failed to render"), false},
		{"gRPC unavailable", t.Context(), status.Error(codes.Unavailable, "remote infra provider unavailable"), false},
		{"gRPC unavailable after a reset", t.Context(), status.Error(codes.Unavailable, "error reading from server: read tcp 10.0.0.1:4321->10.0.0.2:9002: read: connection reset by peer"), false},
		{"forbidden, message quotes a reset", t.Context(), kerrors.NewForbidden(gr, "envoy-default-eg", errors.New("denied: upstream connection reset by peer")), false},
		{"runner stopping", stopped, webhookErr, false},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, isRetryable(tc.ctx, wrapApplyErr(tc.err)))
		})
	}
}

// TestHandlerPanicDoesNotBlockRetries checks that a panic recovered in the
// handler, here on an Infra IR without a proxy, leaves later updates and
// retries running.
func TestHandlerPanicDoesNotBlockRetries(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		start := time.Now()
		failed := false // startTestRunner checks that the manager is never called concurrently
		mgr := &fakeManager{createErr: func(_ context.Context, infra *ir.Infra, _ int) error {
			if infra.Proxy.Name == "default/eg" && !failed {
				failed = true
				return webhookErr
			}
			return nil
		}}
		r, _ := startTestRunner(t, mgr)

		r.InfraIR.Store("default/eg", &message.InfraIRWithContext{Infra: testInfra("default/eg")})
		r.InfraIR.Store("default/bad", &message.InfraIRWithContext{Infra: &ir.Infra{}})
		r.InfraIR.Store("default/b", &message.InfraIRWithContext{Infra: testInfra("default/b")})
		time.Sleep(time.Minute)
		synctest.Wait()
		require.Equal(t, seconds(0, 1), mgr.createOffsets(start, "default/eg"))
		require.Equal(t, seconds(0), mgr.createOffsets(start, "default/b"))
	})
}

// TestPanicInRetryIsRecovered checks that a retry that panics stops retrying
// its key but leaves the retries of other keys and the handler running.
func TestPanicInRetryIsRecovered(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		start := time.Now()
		calls := map[string]int{} // startTestRunner checks that the manager is never called concurrently
		mgr := &fakeManager{createErr: func(_ context.Context, infra *ir.Infra, _ int) error {
			calls[infra.Proxy.Name]++
			switch {
			case infra.Proxy.Name == "default/a" && calls["default/a"] == 2:
				panic("failed to render the HPA")
			case calls[infra.Proxy.Name] == 1:
				return webhookErr
			}
			return nil
		}}
		r, _ := startTestRunner(t, mgr)

		r.InfraIR.Store("default/a", &message.InfraIRWithContext{Infra: testInfra("default/a")})
		time.Sleep(2 * time.Second) // the retry of default/a panics at 1s.
		r.InfraIR.Store("default/b", &message.InfraIRWithContext{Infra: testInfra("default/b")})
		time.Sleep(10 * time.Minute)
		synctest.Wait()
		require.Equal(t, seconds(0, 1), mgr.createOffsets(start, "default/a"))
		require.Equal(t, seconds(2, 3), mgr.createOffsets(start, "default/b"))
		requireNoRetry(t, r, "default/a")
		requireNoRetry(t, r, "default/b")
	})
}

// TestRetrySpanLinksToFailedUpdate checks that each retry starts a new trace
// whose span links to the span of the update that failed, rather than
// becoming a child of that completed span.
func TestRetrySpanLinksToFailedUpdate(t *testing.T) {
	sr := tracetest.NewSpanRecorder()
	original := tracer
	tracer = sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr)).Tracer("test")
	t.Cleanup(func() { tracer = original })

	synctest.Test(t, func(t *testing.T) {
		mgr := &fakeManager{createErrs: []error{webhookErr}}
		r, _ := startTestRunner(t, mgr)

		r.InfraIR.Store("default/eg", &message.InfraIRWithContext{Infra: testInfra("default/eg")})
		time.Sleep(retryWithin)
		synctest.Wait()
	})
	var update trace.SpanContext
	var retries []sdktrace.ReadOnlySpan
	for _, s := range sr.Ended() {
		switch s.Name() {
		case "InfrastructureRunner.updateProxyInfraFromSubscription":
			update = s.SpanContext()
		case "InfrastructureRunner.retryProxyInfra":
			retries = append(retries, s)
		}
	}
	require.True(t, update.IsValid())
	require.Len(t, retries, 1)
	require.False(t, retries[0].Parent().IsValid())
	require.Len(t, retries[0].Links(), 1)
	require.Equal(t, update, retries[0].Links()[0].SpanContext)
}
