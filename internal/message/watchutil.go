// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package message

import (
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/telepresenceio/watchable"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/metrics"
)

type Update[K comparable, V any] watchable.Update[K, V]

// TODO: Remove the global logger and localize the scope of the logger.
var logger = logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo).WithName("watchable")

type Metadata struct {
	Runner  string
	Message MessageName
}

func PublishMetric(meta Metadata, count int) {
	watchablePublishTotal.WithSuccess(meta.LabelValues()...).Add(float64(count))
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
	handle func(updateFunc Update[K, V], errChans chan error),
	update Update[K, V],
	meta Metadata,
	errChans chan error,
) {
	defer func() {
		if r := recover(); r != nil {
			logger.WithValues("runner", meta.Runner).Error(fmt.Errorf("%+v", r), "observed a panic",
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

// HandleSubscription takes a channel returned by
// watchable.Map.Subscribe() (or .SubscribeSubset()), and calls the
// given function for each initial value in the map, and for any
// updates.
//
// This is better than simply iterating over snapshot.Updates because
// it handles the case where the watchable.Map already contains
// entries before .Subscribe is called.
func HandleSubscription[K comparable, V any](
	meta Metadata,
	subscription <-chan watchable.Snapshot[K, V],
	handle func(updateFunc Update[K, V], errChans chan error),
) {
	// TODO: find a suitable value
	errChans := make(chan error, 10)
	go func() {
		for err := range errChans {
			logger.WithValues("runner", meta.Runner).Error(err, "observed an error")
			watchableSubscribeTotal.WithFailure(metrics.ReasonError, meta.LabelValues()...).Increment()
		}
	}()
	defer close(errChans)

	if snapshot, ok := <-subscription; ok {
		for k, v := range snapshot.State {
			handleWithCrashRecovery(handle, Update[K, V]{
				Key:   k,
				Value: v,
			}, meta, errChans)
		}
	}
	for snapshot := range subscription {
		watchableDepth.With(meta.LabelValues()...).Record(float64(len(subscription)))

		for _, update := range coalesceUpdates(meta.Runner, snapshot.Updates) {
			handleWithCrashRecovery(handle, Update[K, V](update), meta, errChans)
		}
	}
}

// coalesceUpdates merges multiple updates for the same key into a single update,
// preserving the latest state for each key.
// This helps reduce redundant processing and ensures that only the most recent update per key is handled.
func coalesceUpdates[K comparable, V any](runner string, updates []watchable.Update[K, V]) []watchable.Update[K, V] {
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
		logger.WithValues("runner", runner).Info(
			"coalesced updates",
			"count", len(result),
			"before", len(updates),
		)
	}
	return result
}
