// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import "github.com/envoyproxy/gateway/internal/metrics"

var (
	statusUpdateTotal = metrics.NewCounter(
		"status_update_total",
		"Total number of status update by object kind.",
	)

	statusUpdateDurationSeconds = metrics.NewHistogram(
		"status_update_duration_seconds",
		"How long a status update takes to finish.",
		[]float64{0.001, 0.01, 0.1, 1, 5, 10},
	)

	kindLabel = metrics.NewLabel("kind")

	topologyInjectorEventsTotal = metrics.NewCounter(
		"topology_injector_webhook_events_total",
		"Total number of topology injector webhook events.",
	)
)

const (
	// statusNoAction means the status of metric is taking no action.
	statusNoAction = "no_action"
)
