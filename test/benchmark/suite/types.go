// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build benchmark

package suite

import (
	"time"

	kube "github.com/envoyproxy/gateway/internal/kubernetes"
	prom "github.com/envoyproxy/gateway/test/utils/prometheus"
)

// RoutePropagationTiming contains timing measurements for route propagation
type RoutePropagationTiming struct {
	RouteAcceptedTime  time.Duration // Route creation → RouteConditionAccepted=True
	DataPlaneReadyTime time.Duration // Accepted=True → First successful request
	EndToEndTime       time.Duration // Route creation → Traffic flowing correctly
	RouteCount         int           // Number of routes being tested
}

// BenchmarkMetricSample contains sampled metrics and profiles data.
type BenchmarkMetricSample struct {
	ControlPlaneMem float64
	ControlPlaneCPU float64
	DataPlaneMem    float64
	DataPlaneCPU    float64

	HeapProfile []byte
}

type BenchmarkReport struct {
	Name              string
	ProfilesOutputDir string
	// Nighthawk benchmark result
	Result []byte
	// Prometheus metrics and pprof profiles sampled data
	Samples []BenchmarkMetricSample
	// Route propagation timing data
	PropagationTiming *RoutePropagationTiming

	kubeClient kube.CLIClient
	promClient *prom.Client
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
