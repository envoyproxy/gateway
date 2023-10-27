// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package message

import (
	"time"

	"github.com/telepresenceio/watchable"

	"github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/metrics"
)

type Update[K comparable, V any] watchable.Update[K, V]

var logger = logging.DefaultLogger(v1alpha1.LogLevelInfo).WithName("watchable")

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
	//TODO: find a suitable value
	errChans := make(chan error, 10)
	go func() {
		for err := range errChans {
			logger.WithValues("runner", meta.Runner).Error(err, "observed an error")
			watchableSubscribedErrorsTotal.With(meta.LabelValues()...).Increment()
		}
	}()

	if snapshot, ok := <-subscription; ok {
		for k, v := range snapshot.State {
			startHandleTime := time.Now()
			handle(Update[K, V]{
				Key:   k,
				Value: v,
			}, errChans)
			watchableSubscribedTotal.With(meta.LabelValues()...).Increment()
			watchableSubscribedDurationSeconds.With(meta.LabelValues()...).Record(time.Since(startHandleTime).Seconds())
		}
	}
	for snapshot := range subscription {
		watchableDepth.With(meta.LabelValues()...).Record(float64(len(subscription)))
		for _, update := range snapshot.Updates {
			startHandleTime := time.Now()
			handle(Update[K, V](update), errChans)
			watchableSubscribedTotal.With(meta.LabelValues()...).Increment()
			watchableSubscribedDurationSeconds.With(meta.LabelValues()...).Record(time.Since(startHandleTime).Seconds())
		}
	}
}
