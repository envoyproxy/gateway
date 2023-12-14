// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package message

import "github.com/envoyproxy/gateway/internal/metrics"

var (
	watchableDepth = metrics.NewGauge("watchable_depth", "Current depth of watchable queue.")

	watchableSubscribedDurationSeconds = metrics.NewHistogram("watchable_subscribed_duration_seconds", "How long in seconds a subscribed watchable is handled.", []float64{0.001, 0.01, 0.1, 1, 5, 10})

	watchableSubscribedTotal = metrics.NewCounter("watchable_subscribed_total", "Total number of subscribed watchable.")

	watchableSubscribedErrorsTotal = metrics.NewCounter("watchable_subscribed_errors_total", "Total number of subscribed watchable errors.")

	runnerLabel = metrics.NewLabel("runner")

	messageLabel = metrics.NewLabel("message")
)
