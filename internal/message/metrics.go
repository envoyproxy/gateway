// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package message

import "github.com/envoyproxy/gateway/internal/metrics"

var (
	messageDepth = metrics.NewGauge("message_depth", "Current depth of message queue.")

	messageSubscribedDurationSeconds = metrics.NewHistogram("message_subscribed_duration_seconds", "How long in seconds a subscribed message is handled.", []float64{0.001, 0.01, 0.1, 1, 5, 10})

	messageSubscribedTotal = metrics.NewCounter("message_subscribed_total", "Total number of subscribed message.")

	messageSubscribedErrorsTotal = metrics.NewCounter("message_subscribed_errors_total", "Total number of subscribed message errors.")

	runnerLabel = metrics.NewLabel("runner")

	resourceLabel = metrics.NewLabel("resource")
)
