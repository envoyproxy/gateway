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
	Message string
}

func (m Metadata) LabelValues() []metrics.LabelValue {
	labels := make([]metrics.LabelValue, 0, 2)
	if m.Runner != "" {
		labels = append(labels, runnerLabel.Value(m.Runner))
	}
	if m.Message != "" {
		labels = append(labels, messageLabel.Value(m.Message))
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
		for _, update := range snapshot.Updates {
			handleWithCrashRecovery(handle, Update[K, V](update), meta, errChans)
		}
	}
}

func HandleStore[K comparable, V any](meta Metadata, key K, value V, publish *watchable.Map[K, V]) {
	publish.Store(key, value)
	watchablePublishTotal.WithSuccess(meta.LabelValues()...).Increment()
}
