// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package message

import "github.com/envoyproxy/gateway/internal/metrics"

var (
	watchableDepth = metrics.NewGauge(
		"watchable_depth",
		"Current depth of watchable queue.",
	)

	panicCounter = metrics.NewCounter(
		"watchable_panics_recovered_total",
		"Total number of panics recovered while handling items in queue.",
	)

	watchableSubscribeDurationSeconds = metrics.NewHistogram(
		"watchable_subscribe_duration_seconds",
		"How long in seconds a subscribed watchable queue is handled.",
		// Same spacing as the k8s rest client metrics, extended to 120s.
		[]float64{0.005, 0.025, 0.1, 0.25, 0.5, 1.0, 2.0, 4.0, 8.0, 15.0, 30.0, 60.0, 120.0},
	)

	watchableSubscribeTotal = metrics.NewCounter(
		"watchable_subscribe_total",
		"Total number of subscribed watchable queue.",
	)

	watchablePublishTotal = metrics.NewCounter(
		"watchable_publish_total",
		"Total number of published updates to watchable queue.",
	)

	watchableEventTotal = metrics.NewCounter(
		"watchable_event_total",
		"Total number of runner events.",
	)

	runnerLabel          = metrics.NewLabel("runner")
	messageLabel         = metrics.NewLabel("message")
	runnerEventTypeLabel = metrics.NewLabel("event_type")
)
