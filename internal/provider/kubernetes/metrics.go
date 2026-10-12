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

	reconcileDebouncePending = metrics.NewGauge(
		"reconcile_debounce_pending",
		"Number of enqueue requests merged into a debounced batch when it was flushed.",
	)

	reconcileDebounceFlushTotal = metrics.NewCounter(
		"reconcile_debounce_flush_total",
		"Total number of debounced reconcile batches flushed.",
	)

	reconcileDebounceDelaySeconds = metrics.NewHistogram(
		"reconcile_debounce_delay_seconds",
		"How long in seconds a reconcile was held by the debouncer before being enqueued.",
		// A quiet-period flush lands just above After and a forced flush lands at
		// Max, so the useful range has to span both, and neither bound is fixed:
		// both are configurable.
		[]float64{0.005, 0.025, 0.1, 0.25, 0.5, 1, 2, 4, 8, 15, 30, 60},
	)

	debounceReasonLabel = metrics.NewLabel("reason")
)

// Reasons a debounced batch of reconcile requests was flushed.
const (
	// debounceReasonQuiet indicates the quiet period elapsed with no new request.
	debounceReasonQuiet = "quiet"
	// debounceReasonMax indicates the maximum hold time was reached while requests
	// kept arriving.
	debounceReasonMax = "max"
	// debounceReasonClose indicates the queue was shut down and the pending batch
	// was drained.
	debounceReasonClose = "close"
)

const (
	// statusNoAction means the status of metric is taking no action.
	statusNoAction = "no_action"
)
