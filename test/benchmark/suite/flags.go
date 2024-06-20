// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build benchmark
// +build benchmark

package suite

import "flag"

var (
	RPS            = flag.String("rps", "1000", "The target requests-per-second rate.")
	Connections    = flag.String("connections", "10", "The maximum allowed number of concurrent connections per event loop. HTTP/1 only.")
	Duration       = flag.String("duration", "60", "The number of seconds that the test should run.")
	Concurrency    = flag.String("concurrency", "auto", "The number of concurrent event loops that should be used.")
	ReportSavePath = flag.String("report-save-path", "", "The path where to save the benchmark test report.")
)
