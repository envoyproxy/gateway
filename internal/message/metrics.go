// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package message

import "github.com/envoyproxy/gateway/internal/metrics"

var (
	// watchableDepth is recorded as the length of the snapshot channel. Note that the
	// watchable library delivers snapshots over an unbuffered channel, so this gauge is
	// always 0 in practice; it is kept for backwards compatibility only. To observe how
	// many updates are being merged, use watchableDebouncePending.
	watchableDepth = metrics.NewGauge(
		"watchable_depth",
		"Current depth of watchable queue.",
	)

	panicCounter = metrics.NewCounter(
		"watchable_panics_recovered_total",
		"Total number of panics recovered while handling items in queue.",
	)

	// The upper buckets are deliberately coarse and wide: in large clusters a
	// single translation can take tens of seconds, and buckets that top out at
	// 10s make every such build indistinguishable, hiding the real p95/p99.
	watchableSubscribeDurationSeconds = metrics.NewHistogram(
		"watchable_subscribe_duration_seconds",
		"How long in seconds a subscribed watchable queue is handled.",
		[]float64{0.001, 0.01, 0.1, 1, 5, 10, 30, 60, 120},
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

	watchableCoalescedTotal = metrics.NewCounter(
		"watchable_coalesced_updates_total",
		"Total number of updates dropped by coalescing because a newer update for the same key superseded them.",
	)

	watchableDebouncePending = metrics.NewGauge(
		"watchable_debounce_pending",
		"Number of updates accumulated in a debounced batch when it was flushed.",
	)

	watchableDebounceFlushTotal = metrics.NewCounter(
		"watchable_debounce_flush_total",
		"Total number of debounced batches flushed.",
	)

	watchableDebounceDelaySeconds = metrics.NewHistogram(
		"watchable_debounce_delay_seconds",
		"How long in seconds updates were held by the debouncer before being flushed.",
		// Matches the bucket spacing used for rest_client_request_duration_seconds.
		// A quiet-period flush lands just above After and a forced flush lands at Max,
		// so the useful range spans both, and neither bound is fixed: both are
		// configurable.
		[]float64{0.005, 0.025, 0.1, 0.25, 0.5, 1, 2, 4, 8, 15, 30, 60},
	)

	runnerLabel          = metrics.NewLabel("runner")
	messageLabel         = metrics.NewLabel("message")
	runnerEventTypeLabel = metrics.NewLabel("event_type")
	debounceReasonLabel  = metrics.NewLabel("reason")
)

// Reasons a debounced batch of updates was flushed.
const (
	// debounceReasonQuiet indicates the quiet period elapsed with no new update.
	debounceReasonQuiet = "quiet"
	// debounceReasonMax indicates the maximum delay was reached while updates
	// kept arriving.
	debounceReasonMax = "max"
	// debounceReasonClose indicates the subscription was closed and the pending
	// batch was drained.
	debounceReasonClose = "close"
)
