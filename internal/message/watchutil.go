// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package message

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/telepresenceio/watchable"

	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/metrics"
)

type Update[K comparable, V any] watchable.Update[K, V]

type Metadata struct {
	Runner  string
	Message MessageName
}

// DebounceOptions configures time-based coalescing of watchable updates.
type DebounceOptions struct {
	// After is the quiet period. A pending batch of updates is flushed once no
	// new update has arrived for this duration.
	After time.Duration

	// Max bounds how long an update may be held before a flush is forced, so
	// that sustained churn cannot delay propagation indefinitely.
	Max time.Duration
}

func PublishMetric(meta Metadata, count int) {
	watchablePublishTotal.WithSuccess(meta.LabelValues()...).Add(float64(count))
}

func PublishRunnerEventMetric(runnerName string, isDelete bool) {
	eventType := "update"
	if isDelete {
		eventType = "delete"
	}
	watchableEventTotal.With(
		runnerEventTypeLabel.Value(eventType),
		runnerLabel.Value(runnerName),
	).Add(1)
}

func (m Metadata) LabelValues() []metrics.LabelValue {
	labels := make([]metrics.LabelValue, 0, 2)
	if m.Runner != "" {
		labels = append(labels, runnerLabel.Value(m.Runner))
	}
	if m.Message != "" {
		labels = append(labels, messageLabel.Value(string(m.Message)))
	}

	return labels
}

// handleWithCrashRecovery calls the provided handle function and gracefully recovers from any panics
// that might occur when the handle function is called.
func handleWithCrashRecovery[K comparable, V any](
	l logging.Logger,
	handle func(updateFunc Update[K, V], errChans chan error),
	update Update[K, V],
	meta Metadata,
	errChans chan error,
) {
	logger := l.WithValues("runner", meta.Runner)
	defer func() {
		if r := recover(); r != nil {
			logger.Error(fmt.Errorf("%+v", r), "observed a panic",
				"stackTrace", string(debug.Stack()))
			watchableSubscribeTotal.WithFailure(metrics.ReasonError, meta.LabelValues()...).Increment()
			panicCounter.WithFailure(metrics.ReasonError, meta.LabelValues()...).Increment()
		}
	}()
	startHandleTime := time.Now()
	handle(update, errChans)
	watchableSubscribeTotal.WithSuccess(meta.LabelValues()...).Increment()
	watchableSubscribeDurationSeconds.With(meta.LabelValues()...).Record(time.Since(startHandleTime).Seconds())
}

// HandleSubscriptionWithDebounce takes a channel returned by
// watchable.Map.Subscribe() (or .SubscribeSubset()), and calls the
// given function for each initial value in the map, and for any
// updates.
//
// This is better than simply iterating over snapshot.Updates because
// it handles the case where the watchable.Map already contains
// entries before .Subscribe is called.
//
// A non-nil debounce merges bursts of updates before handing them off, which bounds
// how often handle runs when the upstream source is churning. A pending batch is
// flushed when either no new update has arrived for After, or the batch has been held
// for Max. The former keeps propagation fast for isolated changes; the latter caps how
// long propagation can be delayed while changes keep arriving.
func HandleSubscriptionWithDebounce[K comparable, V any](l logging.Logger,
	meta Metadata,
	subscription <-chan watchable.Snapshot[K, V],
	handle func(updateFunc Update[K, V], errChans chan error),
	debounce *DebounceOptions,
) {
	// TODO: find a suitable value
	errChans := make(chan error, 10)
	go func() {
		for err := range errChans {
			l.Error(err, "observed an error")
			watchableSubscribeTotal.WithFailure(metrics.ReasonError, meta.LabelValues()...).Increment()
		}
	}()
	defer close(errChans)

	// The initial snapshot is never debounced: it is the bootstrap state, and
	// delaying it would delay the first translation for no benefit.
	if snapshot, ok := <-subscription; ok {
		for k, v := range snapshot.State {
			handleWithCrashRecovery(l, handle, Update[K, V]{
				Key:   k,
				Value: v,
			}, meta, errChans)
		}
	}

	if debounce != nil {
		handleDebounced(l, meta, subscription, handle, errChans, *debounce)
		return
	}

	for snapshot := range subscription {
		watchableDepth.With(meta.LabelValues()...).Record(float64(len(subscription)))

		for _, update := range coalesceUpdates(l, snapshot.Updates) {
			handleWithCrashRecovery(l, handle, Update[K, V](update), meta, errChans)
		}
	}
}

// HandleSubscription is HandleSubscriptionWithDebounce without debouncing: every
// delivered snapshot is handled as soon as it arrives.
func HandleSubscription[K comparable, V any](l logging.Logger,
	meta Metadata,
	subscription <-chan watchable.Snapshot[K, V],
	handle func(updateFunc Update[K, V], errChans chan error),
) {
	HandleSubscriptionWithDebounce(l, meta, subscription, handle, nil)
}

// handleDebounced consumes the subscription, merging updates that arrive close
// together into a single batch before calling handle.
//
// A batch is flushed when the quiet period elapses with no new update, when the
// maximum delay is reached, or when the subscription closes. Note that while
// handle is running the subscription is not being read, so further updates
// accumulate in the watchable library's own coalescer; the next batch therefore
// starts fresh once handle returns.
func handleDebounced[K comparable, V any](
	l logging.Logger,
	meta Metadata,
	subscription <-chan watchable.Snapshot[K, V],
	handle func(updateFunc Update[K, V], errChans chan error),
	errChans chan error,
	debounce DebounceOptions,
) {
	var (
		pending []watchable.Update[K, V]
		// quietTimer is reset by every new update; maxTimer is armed once per
		// batch and deliberately not reset, so it caps the total hold time.
		quietTimer, maxTimer *time.Timer
		quietC, maxC         <-chan time.Time
		batchStartedAt       time.Time
	)

	// Go 1.23 and later guarantee that Stop and Reset never leave a stale value
	// in the timer channel, so no draining is needed here.
	disarm := func() {
		if quietTimer != nil {
			quietTimer.Stop()
		}
		if maxTimer != nil {
			maxTimer.Stop()
		}
		quietC, maxC = nil, nil
	}

	flush := func(reason string) {
		disarm()
		if len(pending) == 0 {
			return
		}

		batched := len(pending)
		updates := coalesceUpdates(l, pending)
		pending = nil

		// coalesceUpdates already logs when it merges anything, and the flush
		// reason is carried on the metric below, so nothing is logged here: this
		// runs on the hot path of a churning cluster.
		labels := meta.LabelValues()
		watchableDebouncePending.With(labels...).Record(float64(batched))
		watchableDebounceDelaySeconds.With(labels...).Record(time.Since(batchStartedAt).Seconds())
		watchableDebounceFlushTotal.With(append(meta.LabelValues(), debounceReasonLabel.Value(reason))...).Increment()

		for _, update := range updates {
			handleWithCrashRecovery(l, handle, Update[K, V](update), meta, errChans)
		}
	}

	for {
		select {
		case snapshot, ok := <-subscription:
			if !ok {
				// Deliver whatever is still pending rather than dropping it.
				flush(debounceReasonClose)
				return
			}
			watchableDepth.With(meta.LabelValues()...).Record(float64(len(subscription)))

			if len(snapshot.Updates) == 0 {
				continue
			}

			if len(pending) == 0 {
				batchStartedAt = time.Now()
				if maxTimer == nil {
					maxTimer = time.NewTimer(debounce.Max)
				} else {
					maxTimer.Reset(debounce.Max)
				}
				maxC = maxTimer.C
			}
			pending = append(pending, snapshot.Updates...)

			if quietTimer == nil {
				quietTimer = time.NewTimer(debounce.After)
			} else {
				quietTimer.Stop()
				quietTimer.Reset(debounce.After)
			}
			quietC = quietTimer.C

		case <-quietC:
			flush(debounceReasonQuiet)

		case <-maxC:
			flush(debounceReasonMax)
		}
	}
}

// coalesceUpdates merges multiple updates for the same key into a single update,
// preserving the latest state for each key.
// This helps reduce redundant processing and ensures that only the most recent update per key is handled.
func coalesceUpdates[K comparable, V any](logger logging.Logger, updates []watchable.Update[K, V]) []watchable.Update[K, V] {
	if len(updates) <= 1 {
		return updates
	}

	seen := make(map[K]struct{}, len(updates))
	write := len(updates) - 1

	for read := len(updates) - 1; read >= 0; read-- {
		update := updates[read]
		if _, ok := seen[update.Key]; ok {
			continue
		}
		seen[update.Key] = struct{}{}
		updates[write] = update
		write--
	}

	result := updates[write+1:]
	if len(result) != len(updates) {
		logger.Info(
			"coalesced updates",
			"count", len(result),
			"before", len(updates),
		)
	}
	return result
}
