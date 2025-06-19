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
	RouteAcceptedTime time.Duration // RouteAccepted Duration: route creation → RouteConditionAccepted=True (control plane processing)
	RouteReadyTime    time.Duration // RouteReady Duration: T(Apply) → T(Route in Envoy / 200 Status on Route Traffic) (total propagation)
	RouteCount        int           // Number of routes being tested

	// Additional detailed timing metrics
	DataPlaneTime time.Duration // Data Plane Time: T(accepted) → T(traffic ready) - xDS propagation
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
