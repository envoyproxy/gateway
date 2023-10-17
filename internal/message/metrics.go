// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package message

import "github.com/envoyproxy/gateway/internal/metrics"

var (
	// metrics definitions
	watchableHandleUpdates = metrics.NewCounter(
		"watchable_queue_handle_updates_total",
		"Total number of updates handled by watchable queue.",
	)

	watchableHandleUpdateErrors = metrics.NewCounter(
		"watchable_queue_handle_updates_errors_total",
		"Total number of update errors handled by watchable queue.",
	)

	watchableDepth = metrics.NewGauge(
		"watchable_queue_depth",
		"Current depth of watchable message queue.",
	)

	watchableHandleUpdateTimeSeconds = metrics.NewHistogram(
		"watchable_queue_handle_update_time_seconds",
		"How long in seconds a update handled by watchable queue.",
		[]float64{0.001, 0.01, 0.1, 1, 5, 10},
	)

	// metrics label definitions
	// component is which component the update belong to.
	componentNameLabel = metrics.NewLabel("component_name")
	// resource is which resource the update belong to.
	resourceTypeLabel = metrics.NewLabel("resource_type")
)
