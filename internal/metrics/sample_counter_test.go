// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package metrics_test

import "github.com/envoyproxy/gateway/internal/metrics"

var (
	irUpdates = metrics.NewCounter(
		"ir_updates_total",
		"Number of IR updates, by ir type",
	)
)

func NewCounter() {
	// increment on every xds ir update
	irUpdates.With(irType.Value("xds")).Increment()

	// xds ir updates double
	irUpdates.With(irType.Value("xds")).Record(2)
}
