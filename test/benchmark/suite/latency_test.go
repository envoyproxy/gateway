// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build benchmark

package suite

import (
	"testing"
)

func TestParseNighthawkResults(t *testing.T) {
	// Sample Nighthawk JSON output for testing
	sampleJSON := `{
		"results": [
			{
				"name": "nighthawk_client",
				"statistics": [
					{
						"id": "benchmark_http_client.latency_2xx",
						"name": "benchmark_http_client.latency_2xx",
						"count": 1000,
						"mean": 2500000.0,
						"pstdev": 500000.0,
						"frequency": [
							{"upper_bound": 1000000.0, "count": 100},
							{"upper_bound": 2000000.0, "count": 400},
							{"upper_bound": 3000000.0, "count": 300},
							{"upper_bound": 5000000.0, "count": 150},
							{"upper_bound": 10000000.0, "count": 50}
						]
					},
					{
						"id": "benchmark_http_client.request_to_response",
						"name": "benchmark_http_client.request_to_response",
						"count": 1000,
						"mean": 2000000.0,
						"pstdev": 400000.0,
						"frequency": [
							{"upper_bound": 800000.0, "count": 100},
							{"upper_bound": 1500000.0, "count": 400},
							{"upper_bound": 2500000.0, "count": 300},
							{"upper_bound": 4000000.0, "count": 150},
							{"upper_bound": 8000000.0, "count": 50}
						]
					},
					{
						"id": "benchmark_http_client.queue_to_connect",
						"name": "benchmark_http_client.queue_to_connect",
						"count": 1000,
						"mean": 100000.0,
						"pstdev": 50000.0,
						"frequency": [
							{"upper_bound": 50000.0, "count": 200},
							{"upper_bound": 100000.0, "count": 600},
							{"upper_bound": 200000.0, "count": 200}
						]
					}
				],
				"counters": [
					{
						"name": "benchmark_http_client.upstream_rq_2xx",
						"value": 990
					},
					{
						"name": "benchmark_http_client.upstream_rq_5xx",
						"value": 10
					}
				]
			}
		]
	}`

	// Parse the JSON
	metrics, err := ParseNighthawkResults([]byte(sampleJSON))
	if err != nil {
		t.Fatalf("Failed to parse Nighthawk results: %v", err)
	}

	// Validate basic metrics
	if metrics == nil {
		t.Fatal("Metrics should not be nil")
	}

	// Check total latency metrics
	if metrics.TotalLatency == nil {
		t.Fatal("TotalLatency should not be nil")
	}

	if metrics.TotalLatency.Count != 1000 {
		t.Errorf("Expected count 1000, got %d", metrics.TotalLatency.Count)
	}

	expectedMean := 2.5 // 2500000.0 nanoseconds = 2.5 milliseconds
	if metrics.TotalLatency.Mean != expectedMean {
		t.Errorf("Expected mean %.2f ms, got %.2f ms", expectedMean, metrics.TotalLatency.Mean)
	}

	// Check request-to-response metrics
	if metrics.RequestToResponse == nil {
		t.Fatal("RequestToResponse should not be nil")
	}

	expectedRequestMean := 2.0 // 2000000.0 nanoseconds = 2.0 milliseconds
	if metrics.RequestToResponse.Mean != expectedRequestMean {
		t.Errorf("Expected request-to-response mean %.2f ms, got %.2f ms", expectedRequestMean, metrics.RequestToResponse.Mean)
	}

	// Check connection latency metrics
	if metrics.ConnectionLatency == nil {
		t.Fatal("ConnectionLatency should not be nil")
	}

	expectedConnMean := 0.1 // 100000.0 nanoseconds = 0.1 milliseconds
	if metrics.ConnectionLatency.Mean != expectedConnMean {
		t.Errorf("Expected connection mean %.2f ms, got %.2f ms", expectedConnMean, metrics.ConnectionLatency.Mean)
	}

	// Check derived proxy latency
	if metrics.EstimatedProxyLatency == nil {
		t.Fatal("EstimatedProxyLatency should not be nil")
	}

	// Proxy latency should be positive and reasonable
	if metrics.EstimatedProxyLatency.Mean <= 0 {
		t.Errorf("Proxy latency should be positive, got %.2f ms", metrics.EstimatedProxyLatency.Mean)
	}

	// Check success rate calculation
	expectedSuccessRate := 99.0 // 990 out of 1000 requests successful
	if metrics.SuccessRate != expectedSuccessRate {
		t.Errorf("Expected success rate %.1f%%, got %.1f%%", expectedSuccessRate, metrics.SuccessRate)
	}

	// Check response codes
	if metrics.ResponseCodes["2xx"] != 990 {
		t.Errorf("Expected 990 successful responses, got %d", metrics.ResponseCodes["2xx"])
	}

	if metrics.ResponseCodes["5xx"] != 10 {
		t.Errorf("Expected 10 error responses, got %d", metrics.ResponseCodes["5xx"])
	}
}

func TestLatencyBreakdown(t *testing.T) {
	// Create sample metrics
	metrics := &LatencyMetrics{
		TotalLatency: &LatencyStats{
			Percentiles: map[string]float64{
				"P95": 5.0, // 5ms total latency
			},
		},
		EstimatedProxyLatency: &LatencyStats{
			Percentiles: map[string]float64{
				"P95": 3.5, // 3.5ms proxy latency
			},
		},
	}

	// Get latency breakdown
	breakdown := metrics.GetLatencyBreakdown("P95")
	if breakdown == nil {
		t.Fatal("Breakdown should not be nil")
	}

	// Validate breakdown components
	if breakdown.ProxyLatency != 3.5 {
		t.Errorf("Expected proxy latency 3.5ms, got %.2f ms", breakdown.ProxyLatency)
	}

	// Check that proxy percentage is calculated correctly
	expectedPercent := (3.5 / 5.0) * 100.0 // 70%
	if breakdown.ProxyLatencyPercent != expectedPercent {
		t.Errorf("Expected proxy percentage %.1f%%, got %.1f%%", expectedPercent, breakdown.ProxyLatencyPercent)
	}

	// Validate that total components approximately match
	totalComponents := breakdown.ClientLatency + breakdown.NetworkLatency + 
		breakdown.ProxyLatency + breakdown.ServerLatency
	
	// Should be reasonably close to the total (within 20% tolerance for estimation)
	tolerance := 5.0 * 0.2 // 20% of 5ms
	if totalComponents < 5.0-tolerance || totalComponents > 5.0+tolerance {
		t.Errorf("Component sum %.2f ms is not close to total 5.0 ms", totalComponents)
	}
}

func TestSummaryString(t *testing.T) {
	metrics := &LatencyMetrics{
		TotalLatency: &LatencyStats{
			Percentiles: map[string]float64{
				"P95": 4.2,
				"P99": 8.5,
			},
		},
		EstimatedProxyLatency: &LatencyStats{
			Percentiles: map[string]float64{
				"P95": 2.8,
				"P99": 6.1,
			},
		},
		SuccessRate: 99.5,
	}

	summary := metrics.SummaryString()
	
	// Check that summary contains expected values
	expectedSubstrings := []string{
		"P95=4.20ms",
		"P99=8.50ms",
		"P95=2.80ms",
		"P99=6.10ms",
		"99.5%",
	}

	for _, substr := range expectedSubstrings {
		if !contains(summary, substr) {
			t.Errorf("Summary should contain '%s', got: %s", substr, summary)
		}
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		(s == substr || (len(s) > len(substr) && 
			(s[:len(substr)] == substr || 
			 s[len(s)-len(substr):] == substr || 
			 containsAt(s, substr))))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}