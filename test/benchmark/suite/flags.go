// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build benchmark
// +build benchmark

package suite

import "flag"

var (
	RPS                 = flag.String("rps", "1000", "RPS of benchmark test client")
	Connections         = flag.String("connections", "10", "Connections of benchmark test client")
	Duration            = flag.String("duration", "60", "Duration of benchmark test client")
	Concurrency         = flag.String("concurrency", "auto", "Concurrency of benchmark test client")
	PrefetchConnections = flag.Bool("prefetch-connections", true, "PrefetchConnections of benchmark test client")
)
