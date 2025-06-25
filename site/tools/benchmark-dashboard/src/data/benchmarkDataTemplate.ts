// Benchmark Data Template - Use this as a template for generating new test data

export interface TestConfiguration {
  rps: number;
  connections: number;
  duration: number;
  cpuLimit: string;
  memoryLimit: string;
}

export interface LatencyMetrics {
  min: number; // microseconds
  mean: number; // microseconds
  max: number; // microseconds
  pstdev: number; // microseconds
  percentiles: {
    p50: number;
    p75: number;
    p80: number;
    p90: number;
    p95: number;
    p99: number;
    p999: number;
  };
}

export interface ResourceMetrics {
  envoyGateway: {
    memory: { min: number; max: number; mean: number }; // MiB
    cpu: { min: number; max: number; mean: number }; // percentage
  };
  envoyProxy: {
    memory: { min: number; max: number; mean: number }; // MiB
    cpu: { min: number; max: number; mean: number }; // percentage
  };
}

export interface TestResult {
  testName: string;
  routes: number;
  routesPerHostname: number;
  phase: 'scaling-up' | 'scaling-down';

  // Performance metrics
  throughput: number; // requests per second
  totalRequests: number;

  // Latency data
  latency: LatencyMetrics;

  // Resource usage
  resources: ResourceMetrics;

  // Additional counters
  poolOverflow: number;
  upstreamConnections: number;
}

// ========================================
// TEMPLATE CONFIGURATION
// ========================================

export const testConfigurationTemplate: TestConfiguration = {
  rps: 10000,        // Target requests per second
  connections: 100,   // Number of connections
  duration: 30,      // Test duration in seconds
  cpuLimit: '1000m', // CPU limit for containers
  memoryLimit: '2000Mi' // Memory limit for containers
};

// ========================================
// TEMPLATE TEST RESULTS
// ========================================

// Example template for a single test result
export const testResultTemplate: TestResult = {
  testName: 'scaling up httproutes to {ROUTES} with {ROUTES_PER_HOSTNAME} routes per hostname',
  routes: 0, // FILL: Number of routes in this test
  routesPerHostname: 0, // FILL: Routes per hostname
  phase: 'scaling-up', // FILL: 'scaling-up' or 'scaling-down'

  // FILL: Extract from benchmark tool output
  throughput: 0, // requests per second (from benchmark.http_2xx counter / duration)
  totalRequests: 0, // total successful requests (from benchmark.http_2xx counter)

  // FILL: Extract from latency section of benchmark output
  latency: {
    min: 0,    // minimum latency in microseconds
    mean: 0,   // mean latency in microseconds
    max: 0,    // maximum latency in microseconds
    pstdev: 0, // standard deviation in microseconds
    percentiles: {
      p50: 0,   // 50th percentile (median)
      p75: 0,   // 75th percentile
      p80: 0,   // 80th percentile
      p90: 0,   // 90th percentile
      p95: 0,   // 95th percentile
      p99: 0,   // 99th percentile
      p999: 0   // 99.9th percentile
    }
  },

  // FILL: Extract from metrics table
  resources: {
    envoyGateway: {
      memory: { min: 0, max: 0, mean: 0 }, // Memory usage in MiB
      cpu: { min: 0, max: 0, mean: 0 }     // CPU usage in percentage
    },
    envoyProxy: {
      memory: { min: 0, max: 0, mean: 0 }, // Memory usage in MiB
      cpu: { min: 0, max: 0, mean: 0 }     // CPU usage in percentage
    }
  },

  // FILL: Extract from counters section
  poolOverflow: 0,         // benchmark.pool_overflow value
  upstreamConnections: 0   // upstream_cx_total value
};

// ========================================
// DATA EXTRACTION HELPERS
// ========================================

/**
 * Helper function to parse latency data from benchmark output
 * Use this to extract percentile data from the benchmark latency section
 */
export function parseLatencyData(
  min: string,           // "0s 000ms 335us"
  mean: string,          // "0s 006ms 565us"
  max: string,           // "0s 066ms 668us"
  pstdev: string,        // "0s 011ms 480us"
  p50: string,           // "0s 003ms 258us"
  p75: string,           // "0s 005ms 079us"
  p80: string,           // "0s 005ms 722us"
  p90: string,           // "0s 008ms 679us"
  p95: string,           // "0s 045ms 924us"
  p99: string,           // "0s 053ms 512us"
  p999: string           // "0s 058ms 220us"
): LatencyMetrics {

  // Helper to convert time string to microseconds
  const parseTimeToMicroseconds = (timeStr: string): number => {
    const match = timeStr.match(/(\d+)s (\d+)ms (\d+)us/);
    if (!match) return 0;
    const [, seconds, milliseconds, microseconds] = match;
    return parseInt(seconds) * 1000000 + parseInt(milliseconds) * 1000 + parseInt(microseconds);
  };

  return {
    min: parseTimeToMicroseconds(min),
    mean: parseTimeToMicroseconds(mean),
    max: parseTimeToMicroseconds(max),
    pstdev: parseTimeToMicroseconds(pstdev),
    percentiles: {
      p50: parseTimeToMicroseconds(p50),
      p75: parseTimeToMicroseconds(p75),
      p80: parseTimeToMicroseconds(p80),
      p90: parseTimeToMicroseconds(p90),
      p95: parseTimeToMicroseconds(p95),
      p99: parseTimeToMicroseconds(p99),
      p999: parseTimeToMicroseconds(p999)
    }
  };
}

/**
 * Helper function to parse resource metrics from the metrics table
 */
export function parseResourceMetrics(
  envoyGatewayMemory: string,  // "128.02 / 151.26 / 147.41"
  envoyGatewayCpu: string,     // "0.27 / 0.67 / 0.45"
  envoyProxyMemory: string,    // "0.00 / 26.92 / 22.58"
  envoyProxyCpu: string        // "0.00 / 99.73 / 6.02"
): ResourceMetrics {

  const parseMinMaxMean = (str: string) => {
    const [min, max, mean] = str.split(' / ').map(s => parseFloat(s.trim()));
    return { min, max, mean };
  };

  return {
    envoyGateway: {
      memory: parseMinMaxMean(envoyGatewayMemory),
      cpu: parseMinMaxMean(envoyGatewayCpu)
    },
    envoyProxy: {
      memory: parseMinMaxMean(envoyProxyMemory),
      cpu: parseMinMaxMean(envoyProxyCpu)
    }
  };
}

/**
 * Helper to calculate throughput from counter data
 */
export function calculateThroughput(httpSuccessCount: number, testDuration: number): number {
  return Number((httpSuccessCount / testDuration).toFixed(2));
}

// ========================================
// TEMPLATE DATA ARRAY
// ========================================

// Template array structure - replace with your actual test results
export const benchmarkResultsTemplate: TestResult[] = [
  // Scaling Up Tests
  {
    testName: 'scaling up httproutes to 10 with 2 routes per hostname',
    routes: 10,
    routesPerHostname: 2,
    phase: 'scaling-up',
    throughput: 5440.31, // EXAMPLE - replace with actual
    totalRequests: 163209, // EXAMPLE - replace with actual
    latency: {
      // EXAMPLE DATA - replace with actual parsed values
      min: 335,
      mean: 6565,
      max: 66668,
      pstdev: 11480,
      percentiles: {
        p50: 3258,
        p75: 5079,
        p80: 5722,
        p90: 8679,
        p95: 45924,
        p99: 53512,
        p999: 58220
      }
    },
    resources: {
      // EXAMPLE DATA - replace with actual parsed values
      envoyGateway: {
        memory: { min: 128.02, max: 151.26, mean: 147.41 },
        cpu: { min: 0.27, max: 0.67, mean: 0.45 }
      },
      envoyProxy: {
        memory: { min: 0.00, max: 26.92, mean: 22.58 },
        cpu: { min: 0.00, max: 99.73, mean: 6.02 }
      }
    },
    poolOverflow: 362, // EXAMPLE - replace with actual
    upstreamConnections: 38 // EXAMPLE - replace with actual
  }
  // ADD MORE TEST RESULTS HERE...
];

// ========================================
// AUTO-GENERATED SUMMARY DATA
// ========================================

// These will be automatically calculated from your test results
export function generatePerformanceSummary(results: TestResult[]) {
  return {
    totalTests: results.length,
    scaleUpTests: results.filter(r => r.phase === 'scaling-up').length,
    scaleDownTests: results.filter(r => r.phase === 'scaling-down').length,
    maxRoutes: Math.max(...results.map(r => r.routes)),
    minRoutes: Math.min(...results.map(r => r.routes)),
    avgThroughput: results.reduce((sum, r) => sum + r.throughput, 0) / results.length,
    avgLatency: results.reduce((sum, r) => sum + r.latency.mean, 0) / results.length
  };
}

export function generateLatencyPercentileComparison(results: TestResult[]) {
  return results.map(result => ({
    routes: result.routes,
    phase: result.phase,
    p50: result.latency.percentiles.p50 / 1000, // convert to ms
    p75: result.latency.percentiles.p75 / 1000,
    p90: result.latency.percentiles.p90 / 1000,
    p95: result.latency.percentiles.p95 / 1000,
    p99: result.latency.percentiles.p99 / 1000,
    p999: result.latency.percentiles.p999 / 1000
  }));
}

export function generateResourceTrends(results: TestResult[]) {
  return results.map(result => ({
    routes: result.routes,
    phase: result.phase,
    envoyGatewayMemory: result.resources.envoyGateway.memory.mean,
    envoyGatewayCpu: result.resources.envoyGateway.cpu.mean,
    envoyProxyMemory: result.resources.envoyProxy.memory.mean,
    envoyProxyCpu: result.resources.envoyProxy.cpu.mean
  }));
}

export function generatePerformanceMatrix(results: TestResult[]) {
  return results.map(result => ({
    testName: result.testName,
    routes: result.routes,
    phase: result.phase,
    throughput: result.throughput,
    meanLatency: result.latency.mean / 1000, // convert to ms
    p95Latency: result.latency.percentiles.p95 / 1000, // convert to ms
    totalMemory: result.resources.envoyGateway.memory.mean + result.resources.envoyProxy.memory.mean,
    totalCpu: result.resources.envoyGateway.cpu.mean + result.resources.envoyProxy.cpu.mean
  }));
}

// ========================================
// USAGE INSTRUCTIONS
// ========================================

/*
HOW TO USE THIS TEMPLATE:

1. Copy this file to a new file (e.g., `newBenchmarkData.ts`)

2. Update testConfigurationTemplate with your test settings

3. For each test run:
   - Copy the testResultTemplate structure
   - Fill in the route configuration (routes, routesPerHostname, phase)
   - Extract performance data from benchmark output using the helper functions
   - Parse latency data using parseLatencyData()
   - Parse resource metrics using parseResourceMetrics()
   - Extract counter values for poolOverflow and upstreamConnections

4. Add all your test results to the benchmarkResultsTemplate array

5. Generate summary data using the provided helper functions:
   ```typescript
   export const performanceSummary = generatePerformanceSummary(benchmarkResultsTemplate);
   export const latencyPercentileComparison = generateLatencyPercentileComparison(benchmarkResultsTemplate);
   export const resourceTrends = generateResourceTrends(benchmarkResultsTemplate);
   export const performanceMatrix = generatePerformanceMatrix(benchmarkResultsTemplate);
   ```

6. Update your components to import from the new data file

EXAMPLE EXTRACTION FROM BENCHMARK OUTPUT:

From this benchmark output:
```
benchmark_http_client.latency_2xx (163209 samples)
  min: 0s 000ms 335us | mean: 0s 006ms 565us | max: 0s 066ms 668us

Counter                   Value       Per second
benchmark.http_2xx        163209      5440.31
```

You would extract:
- throughput: 5440.31
- totalRequests: 163209
- latency.min: parseTimeToMicroseconds("0s 000ms 335us") = 335
- latency.mean: parseTimeToMicroseconds("0s 006ms 565us") = 6565
*/
