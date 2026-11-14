// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import "github.com/envoyproxy/gateway/internal/metrics"

var (
	snapshotPublications = metrics.NewCounter("xds_snapshot_publications_total",
		"Snapshots published to the xDS cache, by publication path (full, fast or delete) and status.")
	endpointRefreshDuration = metrics.NewHistogram("xds_endpoint_refresh_duration_seconds",
		"Time to apply a batch of fast endpoint updates to one xDS snapshot and publish it.",
		[]float64{0.0001, 0.001, 0.01, 0.1, 1})
	publicationPath = metrics.NewLabel("path")
)

// Values of the publication path label.
const (
	pathFull   = "full"
	pathFast   = "fast"
	pathDelete = "delete"
)
