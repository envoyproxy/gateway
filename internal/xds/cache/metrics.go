// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package cache

import "github.com/envoyproxy/gateway/internal/metrics"

var (
	xdsSnapshotCreateTotal   = metrics.NewCounter("xds_snapshot_create_total", "Total number of xds snapshot cache creates.")
	xdsSnapshotUpdateTotal   = metrics.NewCounter("xds_snapshot_update_total", "Total number of xds snapshot cache updates by node id.")
	xdsStreamDurationSeconds = metrics.NewHistogram("xds_stream_duration_seconds", "How long a xds stream takes to finish.", []float64{0.1, 10, 50, 100, 1000, 10000})

	nodeIDLabel        = metrics.NewLabel("nodeID")
	streamIDLabel      = metrics.NewLabel("streamID")
	isDeltaStreamLabel = metrics.NewLabel("isDeltaStream")
)
