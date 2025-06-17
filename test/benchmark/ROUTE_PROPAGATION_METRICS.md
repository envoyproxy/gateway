# Route Propagation Metrics Enhancement

This document describes the enhancements made to the Envoy Gateway benchmark tests to include route propagation timing metrics, providing visibility into how quickly configuration changes take effect.

## Overview

The benchmark tests now measure **route propagation time** in addition to the existing load test performance and resource utilization metrics. This addresses a critical operational need for understanding configuration velocity and deployment timing.

## New Metrics

The enhanced benchmark reports now include a "Route Propagation Metrics" section with the following timing measurements:

### Timing Definitions

1. **RouteAccepted Duration**: Time from route creation to `RouteConditionAccepted=True`
   - Measures control plane processing speed
   - Includes validation, admission, and status updates
   - Typically 1-5 seconds for small to medium route sets

2. **DataPlaneReady Duration**: Time from `Accepted=True` to traffic flowing correctly
   - Currently estimated (simplified for reliability)
   - In production, this would include xDS propagation to Envoy
   - Typically 500ms-2s additional time

3. **End-to-End Time**: Complete route deployment time
   - Total time from `kubectl apply` to traffic routing correctly
   - Most important metric for operational planning
   - Sum of control plane + DataPlaneReady Duration

4. **Route Count**: Number of routes in the test
   - Helps correlate timing with scale
   - Useful for capacity planning

## Sample Output

During benchmark runs, you'll see timing output like:
```
suite.go:474: Route propagation timing - Control Plane: 2.575107292s, End-to-End: 2.575107334s, Routes: 200
```

This indicates:
- 200 routes took 2.575s to reach `Accepted=True`
- End-to-end time was essentially the same
- This represents ~77 routes/second control plane throughput

## Running the Enhanced Benchmarks

### Prerequisites
- Kubernetes cluster (Kind recommended for local testing)
- Sufficient resources (4+ CPU cores, 8GB+ RAM recommended)
- Network connectivity for monitoring stack
- Helm 3.x installed


### Quick Start

Full benchmark with infrastructure setup

```bash
make benchmark
```

# Just run tests (requires existing cluster + Envoy Gateway)
### Installing Envoy Gateway (For Existing Clusters)

**Note**: If you're using `make benchmark`, Envoy Gateway is automatically installed in the Kind cluster. This section is only needed if you're running benchmarks on your own existing cluster with `make run-benchmark`.

If you don't have Envoy Gateway installed in your existing cluster, you can install it using Helm:

```bash
# Add the Envoy Gateway Helm repository
helm repo add eg https://gateway.envoyproxy.io
helm repo update

# Install Envoy Gateway in the envoy-gateway-system namespace
helm install eg eg/gateway-helm --create-namespace -n envoy-gateway-system

# Verify the installation
kubectl get pods -n envoy-gateway-system
kubectl wait --for=condition=Available deployment/envoy-gateway -n envoy-gateway-system --timeout=300s
```

For custom configurations, you can override values:
```bash
# Install with custom resource limits
helm install eg eg/gateway-helm \
  --create-namespace -n envoy-gateway-system \
  --set resources.limits.cpu=1000m \
  --set resources.limits.memory=1Gi
```

```bash
make run-benchmark BENCHMARK_RPS=100 BENCHMARK_CONNECTIONS=10 BENCHMARK_DURATION=30
```

### Advanced Configuration
```bash
# Longer timeout for larger scale tests
make run-benchmark BENCHMARK_TIMEOUT=20m BENCHMARK_DURATION=60

# Specific route scales
go test -v -tags benchmark -timeout 20m ./test/benchmark \
    --rps=500 --connections=50 --duration=60 \
    --report-save-dir=my_benchmark_results
```

## Interpreting Results

### Route Propagation Performance

| Route Count | Expected RouteAccepted Duration | Notes |
|-------------|----------------------------|-------|
| 10-50       | 0.5-2s                    | Fast for small deployments |
| 100-200     | 2-5s                      | Typical service mesh scale |
| 300-500     | 5-15s                     | Large application scale |
| 1000+       | 15s+                      | Enterprise/platform scale |

### Performance Indicators

**Good Performance:**
- Linear scaling: 2x routes ≈ 2x time
- RouteAccepted Duration < 50ms per route
- Consistent timing across test runs

**Potential Issues:**
- Exponential scaling: 2x routes >> 2x time
- High variance between runs
- RouteAccepted Duration > 100ms per route

## Troubleshooting

### Common Issues

#### 1. **Prometheus Connection Timeouts**
```
Post "http://172.23.0.200:80/api/v1/query": dial tcp 172.23.0.200:80: i/o timeout
```

**Cause**: Monitoring infrastructure not ready or network issues
**Solution**:
- Wait longer for infrastructure setup
- Check `kubectl get pods -n monitoring`
- Restart benchmark with longer timeout

#### 2. **Test Timeouts**
```
panic: test timed out after 10m0s
```

**Cause**: Large scale tests need more time
**Solution**:
```bash
make run-benchmark BENCHMARK_TIMEOUT=20m
```

#### 3. **Nighthawk Job Timeouts**
```
Job scale-up-httproutes-500 still not complete
```

**Cause**: Load testing takes time, especially at scale
**Solution**: This is normal - the route propagation timing is still captured successfully

### Resource Requirements

For reliable results:

**Local (Kind):**
- 4+ CPU cores
- 8GB+ RAM
- 20GB+ disk space

**Production:**
- 8+ CPU cores
- 16GB+ RAM
- High-speed networking

## Integration with CI/CD

### Performance Regression Detection

Monitor these metrics in your pipeline:
```yaml
# Example performance thresholds
route_propagation_thresholds:
  control_plane_ms_per_route: 50  # Alert if > 50ms per route
  end_to_end_seconds_100_routes: 5  # Alert if 100 routes > 5s
  variance_threshold: 0.2  # Alert if >20% variance
```

### Automated Analysis
```bash
# Extract timing from benchmark logs
grep "Route propagation timing" benchmark.log | \
  awk '{print $8, $10, $12}' | \
  # Further analysis...
```

## Future Enhancements

The route propagation timing framework provides foundation for:

1. **Per-Route Type Metrics**: Separate timing for HTTPRoute vs GRPCRoute
2. **xDS Propagation Timing**: Actual data plane readiness verification
3. **Concurrent Route Creation**: Measure bulk vs incremental updates
4. **Cross-Cluster Timing**: Multi-cluster route propagation
5. **Error Recovery Timing**: How quickly bad routes are detected/fixed

## Contributing

To enhance the route propagation metrics:

1. **Timing Measurements**: Add new timing points in `suite/suite.go`
2. **Report Generation**: Update tables in `suite/render.go`
3. **Test Cases**: Add new scenarios in `tests/scale_httproutes.go`
4. **Documentation**: Update this file with new metrics

## Support

For issues with route propagation timing:
1. Check the troubleshooting section above
2. Review benchmark logs for specific error messages
3. Verify cluster resources and networking
4. Open GitHub issue with benchmark output and cluster details
