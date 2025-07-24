// Enhanced types to support multiple versions and test runs

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

// NEW: Version/Run metadata
export interface TestRunMetadata {
  version: string; // e.g., "1.4.1", "1.5.0", "main"
  runId: string; // unique identifier for this test run
  date: string; // ISO date string
  environment?: string; // e.g., "staging", "prod", "local"
  description?: string; // optional description of this run
  gitCommit?: string; // git commit hash if applicable
  downloadUrl?: string; // URL to download the raw benchmark report ZIP file
  testConfiguration: TestConfiguration;
}

// NEW: Complete test suite data
export interface TestSuite {
  metadata: TestRunMetadata;
  results: TestResult[];
}

// NEW: Multi-version dataset
export interface BenchmarkDataset {
  testSuites: TestSuite[];
}

// Utility types for data processing
export interface PerformanceComparison {
  version: string;
  runId: string;
  date: string;
  routes: number;
  phase: 'scaling-up' | 'scaling-down';
  throughput: number;
  meanLatency: number;
  p95Latency: number;
  totalMemory: number;
  totalCpu: number;
}

export interface LatencyComparison {
  version: string;
  runId: string;
  routes: number;
  phase: 'scaling-up' | 'scaling-down';
  p50: number;
  p75: number;
  p90: number;
  p95: number;
  p99: number;
  p999: number;
}

export interface ResourceComparison {
  version: string;
  runId: string;
  routes: number;
  phase: 'scaling-up' | 'scaling-down';
  envoyGatewayMemory: number;
  envoyGatewayCpu: number;
  envoyProxyMemory: number;
  envoyProxyCpu: number;
}
