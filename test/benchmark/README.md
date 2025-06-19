# Envoy Gateway Benchmark Framework - Proxy Latency Analysis

This document explains how the benchmark framework extracts proxy latency from total request processing time, isolating it from client-side latency.

## Overview

The enhanced benchmark framework now provides detailed latency breakdown, specifically extracting proxy latency from the total request processing time. This helps identify performance bottlenecks and understand the impact of proxy processing on overall request latency.

## How Proxy Latency Extraction Works

### 1. Enhanced Data Collection

The framework now uses Nighthawk's JSON output format to collect structured latency data:

```bash
nighthawk_client --output-format json --verbosity warn [other args...]
```

This provides access to detailed histogram data for different latency components:
- `benchmark_http_client.latency_2xx` - Total request latency for successful requests
- `benchmark_http_client.request_to_response` - Time from request start to response completion
- `benchmark_http_client.queue_to_connect` - Connection establishment time

### 2. Latency Component Breakdown

The total request processing time consists of several components:

```
Total Latency = Connection Latency + Request Processing + Network Overhead + Client Processing
```

Where Request Processing includes:
- **Proxy Processing**: Time spent in Envoy Gateway proxy (filtering, routing, load balancing)
- **Server Processing**: Time spent in the backend server
- **Network Transport**: Time for request/response transmission

### 3. Proxy Latency Calculation Methods

The framework uses multiple methods to estimate proxy latency:

#### Method 1: Subtraction from Total Latency
```
Proxy Latency = Total Latency - Connection Latency - Network Overhead - Server Processing
```

#### Method 2: Request-to-Response Analysis
```
Proxy Latency = Request-to-Response Latency - Network Overhead - Server Processing
```

The framework takes the more conservative (lower) estimate to avoid overestimating proxy overhead.

### 4. Baseline Assumptions

For in-cluster Kubernetes deployments, the framework uses these baseline estimates:
- **Network Latency**: 0.5ms (pod-to-pod communication)
- **Server Processing**: 0.1ms (simple test responses)
- **Client Processing**: 0.3ms (Nighthawk client overhead)

These values can be calibrated based on your specific environment.

## Understanding the Report

### Latency Analysis Table

The benchmark report now includes a comprehensive latency analysis:

| Metric | Description |
|--------|-------------|
| **Total Latency** | End-to-end request latency (P50/P95/P99) |
| **Estimated Proxy Latency** | Proxy processing time isolated from total latency |
| **Proxy % of Total** | Percentage of total latency attributed to proxy processing |
| **Success Rate** | Percentage of successful (2xx) responses |
| **Throughput** | Requests per second achieved |

### Detailed Breakdown

The detailed breakdown shows latency components for the P95 percentile:

- **Client Processing**: Time spent in the Nighthawk client
- **Network Round-trip**: Network transmission time
- **Proxy Processing**: Time spent in Envoy Gateway proxy
- **Server Processing**: Time spent in the backend server

## Usage Examples

### Running Benchmarks with Proxy Latency Analysis

```bash
# Run standard benchmark with enhanced latency analysis
make benchmark

# Run with custom parameters
make run-benchmark BENCHMARK_RPS=1000 BENCHMARK_CONNECTIONS=100 BENCHMARK_DURATION=60
```

### Interpreting Results

Example output:
```
| Test Name | Total Latency (ms) | Estimated Proxy Latency (ms) | Proxy % of Total | Success Rate (%) |
|-----------|-------------------|------------------------------|------------------|------------------|
| HTTPRoute-1000 | 2.50 / 5.20 / 8.10 | 1.20 / 3.50 / 6.20 | 67.3% | 99.8 |
```

This indicates:
- P95 total latency is 5.20ms
- P95 proxy latency is 3.50ms (67.3% of total)
- The proxy is the dominant factor in request latency
- 99.8% success rate indicates stable performance

## Limitations and Considerations

### 1. Estimation Accuracy

Proxy latency extraction is based on estimation techniques. The accuracy depends on:
- Consistency of network latency
- Predictability of server processing time
- Stability of client-side processing

### 2. Environment-Specific Factors

The baseline values should be calibrated for your environment:
- **Network topology**: Different Kubernetes CNIs have varying latencies
- **Server characteristics**: Backend service complexity affects processing time
- **Load patterns**: High concurrency may affect latency profiles

### 3. Nighthawk Limitations

- All measurements are from the client perspective
- Server-side timing requires additional instrumentation
- Connection reuse affects connection latency measurements

## Advanced Configuration

### Custom Timing Headers

The framework includes an enhanced test server configuration that adds timing headers:

```yaml
# nighthawk-test-server-timing-config.yaml
v3_response_headers:
  - {header: {key: "Server-Timing", value: "processing;dur=0.1"}}
  - {header: {key: "X-Server-Start-Time", value: "%START_TIME%"}}
  - {header: {key: "X-Server-End-Time", value: "%END_TIME%"}}
```

### Calibrating Baseline Values

To improve accuracy, measure baseline latencies in your environment:

1. **Network Latency**: Run simple connectivity tests between pods
2. **Server Processing**: Benchmark the test server directly
3. **Client Overhead**: Compare Nighthawk with other load testing tools

## Troubleshooting

### Missing Latency Data

If latency metrics show "N/A":
1. Check that Nighthawk is outputting JSON format
2. Verify the JSON parsing is successful
3. Ensure benchmark completed successfully

### Unexpected Proxy Latency Values

If proxy latency seems too high or low:
1. Review baseline assumptions for your environment
2. Check for concurrent system load
3. Verify Envoy Gateway configuration complexity

### Performance Issues

High proxy latency (>50% of total) may indicate:
- Complex routing rules
- Heavy authentication/authorization processing
- Resource constraints (CPU/memory)
- External service dependencies (rate limiting, etc.)

## Contributing

To improve the latency analysis:
1. Add server-side timing measurements
2. Implement environment-specific calibration
3. Add support for additional Nighthawk metrics
4. Enhance visualization and reporting

## References

- [Nighthawk Documentation](https://github.com/envoyproxy/nighthawk)
- [Envoy Proxy Performance Guide](https://www.envoyproxy.io/docs/envoy/latest/faq/performance/how_fast_is_envoy)
- [Gateway API Performance Testing](https://gateway-api.sigs.k8s.io/guides/traffic-splitting/)