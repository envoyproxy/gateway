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

type UpdateMetadata struct {
	Component string
	Resource  string
}

func (m UpdateMetadata) LabelValues() []metrics.LabelValue {
	labels := []metrics.LabelValue{}
	if m.Component != "" {
		labels = append(labels, componentNameLabel.Value(m.Component))
	}
	if m.Resource != "" {
		labels = append(labels, resourceTypeLabel.Value(m.Resource))
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
	meta UpdateMetadata,
	subscription <-chan watchable.Snapshot[K, V],
	handle func(updateFunc Update[K, V], errChans chan error),
) {
	errChans := make(chan error, 10)
	go func() {
		for err := range errChans {
			logger.WithValues("component", meta.Component).Error(err, "observed an error")
			watchableHandleUpdateErrors.With(meta.LabelValues()...).Increment()
		}
	}()

	if snapshot, ok := <-subscription; ok {
		for k, v := range snapshot.State {
			startHandleTime := time.Now()
			handle(Update[K, V]{Key: k, Value: v}, errChans)
			watchableHandleUpdates.With(meta.LabelValues()...).Increment()
			watchableHandleUpdateTimeSeconds.With(meta.LabelValues()...).Record(time.Since(startHandleTime).Seconds())
		}
	}
	for snapshot := range subscription {
		watchableDepth.With(meta.LabelValues()...).RecordInt(int64(len(subscription)))
		for _, update := range snapshot.Updates {
			startHandleTime := time.Now()
			handle(Update[K, V](update), errChans)
			watchableHandleUpdates.With(meta.LabelValues()...).Increment()
			watchableHandleUpdateTimeSeconds.With(meta.LabelValues()...).Record(time.Since(startHandleTime).Seconds())
		}
	}
}
