// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package cache

import "github.com/envoyproxy/gateway/internal/metrics"

var (
	xdsSnapshotCreationTotal   = metrics.NewCounter("xds_snapshot_creation_total", "Total number of xds snapshot cache creation.")
	xdsSnapshotCreationFailed  = metrics.NewCounter("xds_snapshot_creation_failed", "Number of xds snapshot cache creation that failed.")
	xdsSnapshotCreationSuccess = metrics.NewCounter("xds_snapshot_creation_success", "Number of xds snapshot cache creation that succeeded.")

	xdsSnapshotUpdateTotal   = metrics.NewCounter("xds_snapshot_update_total", "Total number of xds snapshot cache updates by node id.")
	xdsSnapshotUpdateFailed  = metrics.NewCounter("xds_snapshot_update_failed", "Number of xds snapshot cache updates that failed by node id.")
	xdsSnapshotUpdateSuccess = metrics.NewCounter("xds_snapshot_update_success", "Number of xds snapshot cache updates that succeeded by node id.")
	nodeIDLabel              = metrics.NewLabel("nodeID")

	xdsStreamDurationSeconds      = metrics.NewHistogram("xds_stream_duration_seconds", "How long a xds stream takes to finish.", []float64{0.1, 10, 50, 100, 1000, 10000})
	xdsDeltaStreamDurationSeconds = metrics.NewHistogram("xds_delta_stream_duration_seconds", "How long a xds delta stream takes to finish.", []float64{0.1, 10, 50, 100, 1000, 10000})
	streamIDLabel                 = metrics.NewLabel("streamID")
)
