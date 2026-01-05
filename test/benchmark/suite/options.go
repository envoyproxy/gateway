// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build benchmark

package suite

// BenchmarkOptions for nighthawk-client.
type BenchmarkOptions struct {
	BaselineRPS string
	Connections string
	Duration    string
	Concurrency string
}

func NewBenchmarkOptions(baselineRPS, connections, duration, concurrency string) BenchmarkOptions {
	return BenchmarkOptions{
		BaselineRPS: baselineRPS,
		Connections: connections,
		Duration:    duration,
		Concurrency: concurrency,
	}
}
