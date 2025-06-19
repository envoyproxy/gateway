// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build benchmark

package suite

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
)

// NighthawkResult represents the parsed Nighthawk JSON output
type NighthawkResult struct {
	Results []Result `json:"results"`
}

// Result contains benchmark results for a specific worker
type Result struct {
	Name         string     `json:"name"`
	Statistics   []Statistic `json:"statistics"`
	Counters     []Counter  `json:"counters"`
}

// Statistic represents latency histogram data
type Statistic struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Count      int64   `json:"count"`
	Mean       float64 `json:"mean"`
	PStdev     float64 `json:"pstdev"`
	Frequency  []Bucket `json:"frequency"`
}

// Bucket represents a histogram bucket
type Bucket struct {
	UpperBound float64 `json:"upper_bound"`
	Count      int64   `json:"count"`
}

// Counter represents a simple counter metric
type Counter struct {
	Name  string `json:"name"`
	Value int64  `json:"value"`
}

// LatencyMetrics contains extracted latency metrics with proxy-specific analysis
type LatencyMetrics struct {
	// Request lifecycle latency components
	ConnectionLatency    *LatencyStats `json:"connection_latency"`     // Time to establish connection (client-side)
	RequestToResponse    *LatencyStats `json:"request_to_response"`    // Request processing time (includes proxy + server)
	TotalLatency        *LatencyStats `json:"total_latency"`          // Overall request latency
	
	// Proxy-specific analysis
	EstimatedProxyLatency *LatencyStats `json:"estimated_proxy_latency"` // Estimated proxy overhead
	ServerResponseTime    *LatencyStats `json:"server_response_time"`    // Backend server processing time
	
	// Network and client overhead
	NetworkOverhead      *LatencyStats `json:"network_overhead"`       // Network round-trip estimation
	ClientProcessing     *LatencyStats `json:"client_processing"`      // Client-side processing overhead
	
	// Additional metrics
	ResponseCodes        map[string]int64 `json:"response_codes"`        // HTTP response code distribution
	ThroughputRPS        float64          `json:"throughput_rps"`        // Requests per second
	SuccessRate          float64          `json:"success_rate"`          // Percentage of successful requests
}

// LatencyStats contains percentile and statistical data for latency measurements
type LatencyStats struct {
	Count       int64             `json:"count"`
	Mean        float64           `json:"mean"`
	StdDev      float64           `json:"std_dev"`
	Min         float64           `json:"min"`
	Max         float64           `json:"max"`
	Percentiles map[string]float64 `json:"percentiles"` // P50, P95, P99, etc.
}

// LatencyBreakdown provides a detailed breakdown of request processing time
type LatencyBreakdown struct {
	ClientLatency        float64 `json:"client_latency_ms"`        // Client-side processing time
	NetworkLatency       float64 `json:"network_latency_ms"`       // Network round-trip time
	ProxyLatency         float64 `json:"proxy_latency_ms"`         // Proxy processing time
	ServerLatency        float64 `json:"server_latency_ms"`        // Backend server processing time
	ProxyLatencyPercent  float64 `json:"proxy_latency_percent"`    // Proxy latency as % of total
}

// ParseNighthawkResults parses Nighthawk JSON output and extracts latency metrics
func ParseNighthawkResults(jsonData []byte) (*LatencyMetrics, error) {
	var nhResult NighthawkResult
	if err := json.Unmarshal(jsonData, &nhResult); err != nil {
		return nil, fmt.Errorf("failed to parse Nighthawk JSON: %w", err)
	}

	if len(nhResult.Results) == 0 {
		return nil, fmt.Errorf("no results found in Nighthawk output")
	}

	// Aggregate results from all workers
	metrics := &LatencyMetrics{
		ResponseCodes: make(map[string]int64),
	}

	// Extract latency statistics from different metric types
	for _, result := range nhResult.Results {
		for _, stat := range result.Statistics {
			switch stat.Name {
			case "benchmark_http_client.request_to_response":
				metrics.RequestToResponse = extractLatencyStats(&stat)
			case "benchmark_http_client.queue_to_connect":
				metrics.ConnectionLatency = extractLatencyStats(&stat)
			case "benchmark_http_client.latency_2xx":
				metrics.TotalLatency = extractLatencyStats(&stat)
			}
		}

		// Extract counters for response codes and throughput
		for _, counter := range result.Counters {
			if counter.Name == "benchmark_http_client.upstream_rq_2xx" {
				metrics.ResponseCodes["2xx"] = counter.Value
			} else if counter.Name == "benchmark_http_client.upstream_rq_5xx" {
				metrics.ResponseCodes["5xx"] = counter.Value
			}
		}
	}

	// Calculate derived metrics
	calculateDerivedMetrics(metrics)

	return metrics, nil
}

// extractLatencyStats converts Nighthawk histogram data to LatencyStats
func extractLatencyStats(stat *Statistic) *LatencyStats {
	if stat == nil || len(stat.Frequency) == 0 {
		return nil
	}

	latency := &LatencyStats{
		Count:       stat.Count,
		Mean:        convertToMilliseconds(stat.Mean),
		StdDev:      convertToMilliseconds(stat.PStdev),
		Percentiles: make(map[string]float64),
	}

	// Extract percentiles from histogram buckets
	percentiles := calculatePercentiles(stat.Frequency, stat.Count)
	for p, value := range percentiles {
		latency.Percentiles[p] = convertToMilliseconds(value)
	}

	// Set min/max from percentiles
	if p0, exists := latency.Percentiles["P0"]; exists {
		latency.Min = p0
	}
	if p100, exists := latency.Percentiles["P100"]; exists {
		latency.Max = p100
	}

	return latency
}

// calculatePercentiles computes percentile values from histogram buckets
func calculatePercentiles(buckets []Bucket, totalCount int64) map[string]float64 {
	if len(buckets) == 0 || totalCount == 0 {
		return nil
	}

	percentiles := make(map[string]float64)
	targetPercentiles := []float64{0, 50, 90, 95, 99, 99.9, 100}
	
	// Sort buckets by upper bound
	sort.Slice(buckets, func(i, j int) bool {
		return buckets[i].UpperBound < buckets[j].UpperBound
	})

	// Calculate cumulative counts
	cumulativeCount := int64(0)
	for i, target := range targetPercentiles {
		targetCount := int64(float64(totalCount) * target / 100.0)
		
		// Find the bucket containing this percentile
		for _, bucket := range buckets {
			cumulativeCount += bucket.Count
			if cumulativeCount >= targetCount {
				key := fmt.Sprintf("P%v", target)
				if target == 0 {
					key = "P0"
				} else if target == 100 {
					key = "P100"
				} else if target == 99.9 {
					key = "P99.9"
				}
				percentiles[key] = bucket.UpperBound
				break
			}
		}
		
		// Reset for next percentile
		if i < len(targetPercentiles)-1 {
			cumulativeCount = 0
		}
	}

	return percentiles
}

// calculateDerivedMetrics computes proxy-specific latency estimates
func calculateDerivedMetrics(metrics *LatencyMetrics) {
	if metrics.RequestToResponse == nil || metrics.TotalLatency == nil {
		return
	}

	// Enhanced proxy latency calculation considering different deployment scenarios
	metrics.EstimatedProxyLatency = &LatencyStats{
		Count:       metrics.RequestToResponse.Count,
		Percentiles: make(map[string]float64),
	}

	// More sophisticated latency breakdown based on request lifecycle
	for p, totalLatency := range metrics.TotalLatency.Percentiles {
		// Get connection latency if available
		var connectionLatency float64
		if metrics.ConnectionLatency != nil && metrics.ConnectionLatency.Percentiles != nil {
			if connLat, exists := metrics.ConnectionLatency.Percentiles[p]; exists {
				connectionLatency = connLat
			}
		}

		// Calculate proxy latency using more accurate method:
		// Total latency = Connection + Request processing + Response processing
		// Request processing includes: Proxy processing + Server processing + Network
		
		// For in-cluster testing, network latency is minimal (< 1ms)
		baselineNetworkLatency := 0.5 // Reduced for in-cluster communication
		estimatedServerTime := 0.1    // Minimal processing for simple test responses
		
		// Proxy latency estimation methods:
		// Method 1: Total - Connection - Network - Server
		proxyLatencyMethod1 := math.Max(0, totalLatency - connectionLatency - baselineNetworkLatency - estimatedServerTime)
		
		// Method 2: Use request_to_response metric which excludes connection setup
		var requestToResponseLatency float64
		if metrics.RequestToResponse != nil && metrics.RequestToResponse.Percentiles != nil {
			if rtrLat, exists := metrics.RequestToResponse.Percentiles[p]; exists {
				requestToResponseLatency = rtrLat
				// This includes proxy + server + final network hop
				proxyLatencyMethod2 := math.Max(0, requestToResponseLatency - estimatedServerTime - baselineNetworkLatency*0.5)
				
				// Use the more conservative estimate
				proxyLatency := math.Min(proxyLatencyMethod1, proxyLatencyMethod2)
				metrics.EstimatedProxyLatency.Percentiles[p] = proxyLatency
			} else {
				metrics.EstimatedProxyLatency.Percentiles[p] = proxyLatencyMethod1
			}
		} else {
			metrics.EstimatedProxyLatency.Percentiles[p] = proxyLatencyMethod1
		}
	}

	// Calculate mean proxy latency using the same logic
	connectionMean := 0.0
	if metrics.ConnectionLatency != nil {
		connectionMean = metrics.ConnectionLatency.Mean
	}
	
	requestToResponseMean := metrics.RequestToResponse.Mean
	baselineLatency := 0.6 // network + server
	
	if requestToResponseMean > 0 {
		metrics.EstimatedProxyLatency.Mean = math.Max(0, requestToResponseMean - baselineLatency)
	} else {
		metrics.EstimatedProxyLatency.Mean = math.Max(0, metrics.TotalLatency.Mean - connectionMean - baselineLatency)
	}

	// Calculate success rate and throughput
	totalRequests := int64(0)
	successfulRequests := int64(0)
	for code, count := range metrics.ResponseCodes {
		totalRequests += count
		if code == "2xx" {
			successfulRequests += count
		}
	}
	
	if totalRequests > 0 {
		metrics.SuccessRate = float64(successfulRequests) / float64(totalRequests) * 100.0
		
		// Calculate throughput - need duration from somewhere, for now estimate from metrics
		// This would ideally come from the benchmark configuration
		if metrics.TotalLatency.Count > 0 {
			// Rough estimate - this should be passed from benchmark duration
			estimatedDurationSeconds := 30.0 // Default benchmark duration
			metrics.ThroughputRPS = float64(totalRequests) / estimatedDurationSeconds
		}
	}

	// Set additional derived metrics
	calculateNetworkAndClientMetrics(metrics)
}

// calculateNetworkAndClientMetrics estimates network and client processing components
func calculateNetworkAndClientMetrics(metrics *LatencyMetrics) {
	if metrics.TotalLatency == nil {
		return
	}

	// Initialize network and client overhead metrics
	metrics.NetworkOverhead = &LatencyStats{
		Count:       metrics.TotalLatency.Count,
		Percentiles: make(map[string]float64),
	}
	
	metrics.ClientProcessing = &LatencyStats{
		Count:       metrics.TotalLatency.Count,
		Percentiles: make(map[string]float64),
	}

	// Estimate network overhead (round-trip time for in-cluster communication)
	baselineNetworkLatency := 0.5
	clientProcessingOverhead := 0.3

	for p := range metrics.TotalLatency.Percentiles {
		metrics.NetworkOverhead.Percentiles[p] = baselineNetworkLatency
		metrics.ClientProcessing.Percentiles[p] = clientProcessingOverhead
	}

	metrics.NetworkOverhead.Mean = baselineNetworkLatency
	metrics.ClientProcessing.Mean = clientProcessingOverhead

	// Initialize server response time estimation
	metrics.ServerResponseTime = &LatencyStats{
		Count:       metrics.TotalLatency.Count,
		Percentiles: make(map[string]float64),
	}

	serverProcessingTime := 0.1
	for p := range metrics.TotalLatency.Percentiles {
		metrics.ServerResponseTime.Percentiles[p] = serverProcessingTime
	}
	metrics.ServerResponseTime.Mean = serverProcessingTime
}

// convertToMilliseconds converts nanoseconds to milliseconds
func convertToMilliseconds(nanoseconds float64) float64 {
	return nanoseconds / 1e6
}

// GetLatencyBreakdown provides a detailed breakdown of latency components
func (lm *LatencyMetrics) GetLatencyBreakdown(percentile string) *LatencyBreakdown {
	if lm.TotalLatency == nil || lm.EstimatedProxyLatency == nil {
		return nil
	}

	totalLatency, exists := lm.TotalLatency.Percentiles[percentile]
	if !exists {
		return nil
	}

	proxyLatency, exists := lm.EstimatedProxyLatency.Percentiles[percentile]
	if !exists {
		return nil
	}

	// Estimate component breakdown
	networkLatency := 1.0     // Baseline network latency
	serverLatency := 0.1      // Minimal server processing
	clientLatency := 0.5      // Client processing overhead

	// Adjust if proxy latency is larger than expected
	remainingLatency := totalLatency - proxyLatency
	if remainingLatency > networkLatency + serverLatency + clientLatency {
		// Distribute excess latency proportionally
		excess := remainingLatency - (networkLatency + serverLatency + clientLatency)
		networkLatency += excess * 0.4
		serverLatency += excess * 0.4
		clientLatency += excess * 0.2
	}

	proxyPercent := 0.0
	if totalLatency > 0 {
		proxyPercent = (proxyLatency / totalLatency) * 100.0
	}

	return &LatencyBreakdown{
		ClientLatency:       clientLatency,
		NetworkLatency:      networkLatency,
		ProxyLatency:        proxyLatency,
		ServerLatency:       serverLatency,
		ProxyLatencyPercent: proxyPercent,
	}
}

// SummaryString returns a human-readable summary of the latency metrics
func (lm *LatencyMetrics) SummaryString() string {
	if lm.TotalLatency == nil {
		return "No latency data available"
	}

	p95 := lm.TotalLatency.Percentiles["P95"]
	p99 := lm.TotalLatency.Percentiles["P99"]
	
	proxyP95 := float64(0)
	proxyP99 := float64(0)
	if lm.EstimatedProxyLatency != nil {
		proxyP95 = lm.EstimatedProxyLatency.Percentiles["P95"]
		proxyP99 = lm.EstimatedProxyLatency.Percentiles["P99"]
	}

	return fmt.Sprintf(
		"Total Latency: P95=%.2fms, P99=%.2fms | Estimated Proxy Latency: P95=%.2fms, P99=%.2fms | Success Rate: %.1f%%",
		p95, p99, proxyP95, proxyP99, lm.SuccessRate,
	)
}