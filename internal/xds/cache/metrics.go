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

	xdsSnapshotUpdateTotal   = metrics.NewCounter("xds_snapshot_update_total", "Total number of xds snapshot cache updates by node name.")
	xdsSnapshotUpdateFailed  = metrics.NewCounter("xds_snapshot_update_failed", "Number of xds snapshot cache updates that failed by node name.")
	xdsSnapshotUpdateSuccess = metrics.NewCounter("xds_snapshot_update_success", "Number of xds snapshot cache updates that succeeded by node name.")

	nodeLabel = metrics.NewLabel("node")
)
