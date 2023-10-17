// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package metrics_test

import "github.com/envoyproxy/gateway/internal/metrics"

var (
	irType        = metrics.NewLabel("ir-type")
	currentIRsNum = metrics.NewGauge(
		"current_irs_queue_num",
		"current number of ir in queue, by ir type",
	)
)

func NewGauge() {
	// only the last recorded value (2) will be exported for this gauge
	currentIRsNum.With(irType.Value("xds")).Record(1)
	currentIRsNum.With(irType.Value("xds")).Record(3)
	currentIRsNum.With(irType.Value("xds")).Record(2)

	currentIRsNum.With(irType.Value("infra")).Record(1)
	currentIRsNum.With(irType.Value("infra")).Record(3)
	currentIRsNum.With(irType.Value("infra")).Record(2)
}
