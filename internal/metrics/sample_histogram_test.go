// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package metrics_test

import "github.com/envoyproxy/gateway/internal/metrics"

var (
	method = metrics.NewLabel("method")

	sentBytes = metrics.NewHistogram(
		"sent_bytes_total",
		"Histogram of sent bytes by method",
		[]float64{10, 50, 100, 1000, 10000},
		metrics.WithUnit(metrics.Bytes),
	)
)

func NewHistogram() {
	sentBytes.With(method.Value("/request/path/1")).Record(458)
}
