// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build benchmark
// +build benchmark

package suite

import "testing"

type BenchmarkTest struct {
	ShortName   string
	Description string
	Test        func(*testing.T, *BenchmarkTestSuite) []*BenchmarkReport
}

// BenchmarkOptions for nighthawk-client.
type BenchmarkOptions struct {
	RPS         string
	Connections string
	Duration    string
	Concurrency string
}

func NewBenchmarkOptions(rps, connections, duration, concurrency string) BenchmarkOptions {
	return BenchmarkOptions{
		RPS:         rps,
		Connections: connections,
		Duration:    duration,
		Concurrency: concurrency,
	}
}
