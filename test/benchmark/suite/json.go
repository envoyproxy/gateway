// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build benchmark

package suite

import (
	"encoding/json"
	"math"
)

func ToJSON(report *BenchmarkSuiteReport) []byte {
	if report == nil {
		return nil
	}

	var results []*JSONTestResult
	for _, r := range report.Reports {
		for _, caseReport := range r.Reports {
			results = append(results, convertCaseReport(caseReport))
		}
	}

	suiteResult := &JSONSuiteResult{
		Metadata: Metadata{
			TestConfiguration: report.Settings,
		},
		Results: results,
	}
	data, _ := json.MarshalIndent(suiteResult, "", "  ")
	return data
}

func convertCaseReport(report *BenchmarkCaseReport) *JSONTestResult {
	if report == nil {
		return nil
	}

	throughput := float64(0)

	totalRequest := 0
	latency := &LatencyMetrics{}

	seconds := report.Result.ExecutionDuration.Seconds
	for _, stat := range report.Result.Statistics {
		if stat.Id == "benchmark_http_client.latency_2xx" {
			totalRequest = int(stat.Count)
			throughput = float64(stat.Count) / float64(seconds)
			latency.Pstdev = stat.GetPstdev().AsDuration().Seconds() * 1000
			latency.Max = stat.GetMax().AsDuration().Seconds() * 1000
			latency.Min = stat.GetMin().AsDuration().Seconds() * 1000
			latency.Mean = stat.GetMean().AsDuration().Seconds() * 1000
			latency.Percentiles = Percentiles{}
			for _, p := range stat.Percentiles {
				switch p.Percentile {
				case 0.5:
					latency.Percentiles.P50 = p.GetDuration().AsDuration().Seconds() * 1000
				case 0.75:
					latency.Percentiles.P75 = p.GetDuration().AsDuration().Seconds() * 1000
				case 0.8:
					latency.Percentiles.P80 = p.GetDuration().AsDuration().Seconds() * 1000
				case 0.9:
					latency.Percentiles.P90 = p.GetDuration().AsDuration().Seconds() * 1000
				case 0.95:
					latency.Percentiles.P95 = p.GetDuration().AsDuration().Seconds() * 1000
				case 0.99:
					latency.Percentiles.P99 = p.GetDuration().AsDuration().Seconds() * 1000
				case 0.999:
					latency.Percentiles.P999 = p.GetDuration().AsDuration().Seconds() * 1000
				}
			}
		}
	}

	counters := map[string]Counter{}
	poolOverflow := uint64(0)
	upstreamConnection := uint64(0)
	for _, c := range report.Result.Counters {
		switch c.Name {
		case "benchmark.pool_overflow":
			poolOverflow = c.Value
		case "upstream_cx_http1_total":
			upstreamConnection = c.Value
		}
		counters[c.Name] = Counter{
			Value:     c.Value,
			PerSecond: float64(c.Value) / float64(seconds),
		}
	}

	r := map[string]ResourceMetrics{}
	r["envoyGateway"], r["envoyProxy"] = getResourceMetrics(report.Samples)

	testResult := &JSONTestResult{
		TestName:           report.Title,
		Routes:             report.Routes,
		RoutesPerHostname:  report.RoutesPerHostname,
		Phase:              report.Phase,
		Throughput:         throughput,
		TotalRequests:      totalRequest,
		Latency:            latency,
		Resources:          r,
		PoolOverflow:       poolOverflow,
		UpstreamConnection: upstreamConnection,
		Counters:           counters,
	}
	return testResult
}

func getResourceMetrics(samples []BenchmarkMetricSample) (ResourceMetrics, ResourceMetrics) {
	cpcMem := make([]float64, 0, len(samples))
	cpCPU := make([]float64, 0, len(samples))
	dpMem := make([]float64, 0, len(samples))
	dpCPU := make([]float64, 0, len(samples))
	for _, sample := range samples {
		cpcMem = append(cpcMem, sample.ControlPlaneContainerMem)
		cpCPU = append(cpCPU, sample.ControlPlaneCPU)
		dpMem = append(dpMem, sample.DataPlaneMem)
		dpCPU = append(dpCPU, sample.DataPlaneCPU)
	}

	return ResourceMetrics{
			CPU:    getResourceUsage(cpCPU),
			Memory: getResourceUsage(cpcMem),
		}, ResourceMetrics{
			CPU:    getResourceUsage(dpCPU),
			Memory: getResourceUsage(dpMem),
		}
}

func getResourceUsage(metrics []float64) *ResourceUsage {
	var minVal, maxVal, avg float64 = math.MaxFloat64, 0, 0
	for _, v := range metrics {
		minVal = math.Min(v, minVal)
		maxVal = math.Max(v, maxVal)
		avg += v
	}
	if minVal == math.MaxFloat64 {
		minVal = 0
	}
	if len(metrics) > 0 {
		avg /= float64(len(metrics))
	}

	return &ResourceUsage{
		Mean: avg,
		Min:  minVal,
		Max:  maxVal,
	}
}
