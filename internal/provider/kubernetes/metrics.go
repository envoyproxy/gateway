// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import "github.com/envoyproxy/gateway/internal/metrics"

var (
	statusUpdateTotal           = metrics.NewCounter("status_update_total", "Total number of status updates by object kind.")
	statusUpdateFailed          = metrics.NewCounter("status_update_failed_total", "Number of status updates that failed by object kind.")
	statusUpdateConflict        = metrics.NewCounter("status_update_conflict_total", "Number of status update conflicts encountered by object kind.")
	statusUpdateSuccess         = metrics.NewCounter("status_update_success_total", "Number of status updates that succeeded by object kind.")
	statusUpdateNoop            = metrics.NewCounter("status_update_noop_total", "Number of status updates that are no-ops by object kind. This is a subset of successful status updates.")
	statusUpdateDurationSeconds = metrics.NewHistogram("status_update_duration_seconds", "How long a status update takes to finish.", []float64{0.001, 0.01, 0.1, 1, 5, 10})

	kindLabel = metrics.NewLabel("kind")
)
