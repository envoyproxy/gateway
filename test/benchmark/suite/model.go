// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build benchmark

package suite

type JSONSuiteResult struct {
	Metadata Metadata          `json:"metadata"`
	Results  []*JSONTestResult `json:"results"`
}

type Metadata struct {
	TestConfiguration map[string]string `json:"testConfiguration"`
}

type JSONTestResult struct {
	TestName           string                     `json:"testName"`
	Routes             int                        `json:"routes"`
	RoutesPerHostname  int                        `json:"routesPerHostname"`
	Phase              string                     `json:"phase"`
	Throughput         float64                    `json:"throughput"`
	TotalRequests      int                        `json:"totalRequests"`
	Latency            *LatencyMetrics            `json:"latency,omitempty"`
	Resources          map[string]ResourceMetrics `json:"resources"`
	PoolOverflow       uint64                     `json:"poolOverflow"`
	UpstreamConnection uint64                     `json:"upstreamConnections"`
	Counters           map[string]Counter         `json:"counters"`
}

type Counter struct {
	Value     uint64  `json:"value"`
	PerSecond float64 `json:"perSecond"`
}

type LatencyMetrics struct {
	Max         float64     `json:"max"`
	Min         float64     `json:"min"`
	Mean        float64     `json:"mean"`
	Pstdev      float64     `json:"pstdev"`
	Percentiles Percentiles `json:"percentiles"`
}

type Percentiles struct {
	P50  float64 `json:"p50"`
	P75  float64 `json:"p75"`
	P80  float64 `json:"p80"`
	P90  float64 `json:"p90"`
	P95  float64 `json:"p95"`
	P99  float64 `json:"p99"`
	P999 float64 `json:"p999"`
}

type ResourceMetrics struct {
	Memory *ResourceUsage `json:"memory"`
	CPU    *ResourceUsage `json:"cpu"`
}

type ResourceUsage struct {
	Max  float64 `json:"max"`
	Min  float64 `json:"min"`
	Mean float64 `json:"mean"`
}
