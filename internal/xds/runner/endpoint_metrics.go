// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import "github.com/envoyproxy/gateway/internal/metrics"

var (
	endpointPublications = metrics.NewCounter("xds_endpoint_publications_total",
		"Desired snapshots committed by endpoint publishers, by publication path.")
	endpointRefreshDuration = metrics.NewHistogram("xds_endpoint_refresh_duration_seconds",
		"Time spent rebuilding, validating, hashing and committing fast endpoint updates.",
		[]float64{0.0001, 0.001, 0.01, 0.1, 1})
	endpointDeliveryRetries = metrics.NewCounter("xds_endpoint_delivery_retries_total",
		"Retries delivering the current desired snapshot or deletion.")
	publicationPath = metrics.NewLabel("path")
)
