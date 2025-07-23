// Centralized benchmark data manager
import {
  BenchmarkDataset,
  TestSuite,
  PerformanceComparison,
  LatencyComparison,
  ResourceComparison
} from './types';
import { benchmarkData as v142TestSuite } from './versions/v1.4.2';
import { benchmarkData as v141TestSuite } from './versions/v1.4.1';
import { benchmarkData as v140TestSuite } from './versions/v1.4.0';
import { benchmarkData as v133TestSuite } from './versions/v1.3.3';
import { benchmarkData as v132TestSuite } from './versions/v1.3.2';
import { benchmarkData as v131TestSuite } from './versions/v1.3.1';
import { benchmarkData as v130TestSuite } from './versions/v1.3.0';
import { benchmarkData as v128TestSuite } from './versions/v1.2.8';
import { benchmarkData as v127TestSuite } from './versions/v1.2.7';
import { benchmarkData as v126TestSuite } from './versions/v1.2.6';
import { benchmarkData as v125TestSuite } from './versions/v1.2.5';
import { benchmarkData as v124TestSuite } from './versions/v1.2.4';
import { benchmarkData as v123TestSuite } from './versions/v1.2.3';
import { benchmarkData as v122TestSuite } from './versions/v1.2.2';
import { benchmarkData as v121TestSuite } from './versions/v1.2.1';
import { benchmarkData as v120TestSuite } from './versions/v1.2.0';
import { benchmarkData as v114TestSuite } from './versions/v1.1.4';
import { benchmarkData as v113TestSuite } from './versions/v1.1.3';
import { benchmarkData as v112TestSuite } from './versions/v1.1.2';
import { benchmarkData as v111TestSuite } from './versions/v1.1.1';
import { benchmarkData as v110TestSuite } from './versions/v1.1.0';

// Import all version data
export const allTestSuites: TestSuite[] = [
  v142TestSuite,
  v141TestSuite,
  v140TestSuite,
  v133TestSuite,
  v132TestSuite,
  v131TestSuite,
  v130TestSuite,
  v128TestSuite,
  v127TestSuite,
  v126TestSuite,
  v125TestSuite,
  v124TestSuite,
  v123TestSuite,
  v122TestSuite,
  v121TestSuite,
  v120TestSuite,
  v114TestSuite,
  v113TestSuite,
  v112TestSuite,
  v111TestSuite,
  v110TestSuite,
];

// Complete dataset
export const benchmarkDataset: BenchmarkDataset = {
  testSuites: allTestSuites
};

// ========================================
// DATA PROCESSING UTILITIES
// ========================================

// Get all available versions
export const getAvailableVersions = (): string[] => {
  return allTestSuites.map(suite => suite.metadata.version);
};

// Get all available run IDs
export const getAvailableRunIds = (): string[] => {
  return allTestSuites.map(suite => suite.metadata.runId);
};

// Filter data by version(s)
export const getDataByVersions = (versions: string[]): TestSuite[] => {
  return allTestSuites.filter(suite => versions.includes(suite.metadata.version));
};

// Filter data by date range
export const getDataByDateRange = (startDate: string, endDate: string): TestSuite[] => {
  return allTestSuites.filter(suite => {
    const suiteDate = new Date(suite.metadata.date);
    return suiteDate >= new Date(startDate) && suiteDate <= new Date(endDate);
  });
};

// Get latest data for each version
export const getLatestDataPerVersion = (): TestSuite[] => {
  const versionMap = new Map<string, TestSuite>();

  allTestSuites.forEach(suite => {
    const existing = versionMap.get(suite.metadata.version);
    if (!existing || new Date(suite.metadata.date) > new Date(existing.metadata.date)) {
      versionMap.set(suite.metadata.version, suite);
    }
  });

  return Array.from(versionMap.values());
};

// ========================================
// COMPARISON DATA GENERATORS
// ========================================

// Generate performance comparison data across versions
export const generatePerformanceComparison = (versions?: string[]): PerformanceComparison[] => {
  const suitesToProcess = versions ? getDataByVersions(versions) : allTestSuites;

  return suitesToProcess.flatMap(suite =>
    suite.results.map(result => ({
      version: suite.metadata.version,
      runId: suite.metadata.runId,
      date: suite.metadata.date,
      routes: result.routes,
      phase: result.phase,
      throughput: result.throughput,
      meanLatency: result.latency.mean / 1000, // convert to ms
      p95Latency: result.latency.percentiles.p95 / 1000, // convert to ms
      totalMemory: result.resources.envoyGateway.memory.mean + result.resources.envoyProxy.memory.mean,
      totalCpu: result.resources.envoyGateway.cpu.mean + result.resources.envoyProxy.cpu.mean
    }))
  );
};

// Generate latency comparison data across versions
export const generateLatencyComparison = (versions?: string[]): LatencyComparison[] => {
  const suitesToProcess = versions ? getDataByVersions(versions) : allTestSuites;

  return suitesToProcess.flatMap(suite =>
    suite.results.map(result => ({
      version: suite.metadata.version,
      runId: suite.metadata.runId,
      routes: result.routes,
      phase: result.phase,
      p50: result.latency.percentiles.p50 / 1000, // convert to ms
      p75: result.latency.percentiles.p75 / 1000,
      p90: result.latency.percentiles.p90 / 1000,
      p95: result.latency.percentiles.p95 / 1000,
      p99: result.latency.percentiles.p99 / 1000,
      p999: result.latency.percentiles.p999 / 1000
    }))
  );
};

// Generate resource comparison data across versions
export const generateResourceComparison = (versions?: string[]): ResourceComparison[] => {
  const suitesToProcess = versions ? getDataByVersions(versions) : allTestSuites;

  return suitesToProcess.flatMap(suite =>
    suite.results.map(result => ({
      version: suite.metadata.version,
      runId: suite.metadata.runId,
      routes: result.routes,
      phase: result.phase,
      envoyGatewayMemory: result.resources.envoyGateway.memory.mean,
      envoyGatewayCpu: result.resources.envoyGateway.cpu.mean,
      envoyProxyMemory: result.resources.envoyProxy.memory.mean,
      envoyProxyCpu: result.resources.envoyProxy.cpu.mean
    }))
  );
};

// ========================================
// BACKWARD COMPATIBILITY
// ========================================
// Export data in the old format for existing components

// Get the latest v1.4.1 data for backward compatibility
const currentV141Suite = allTestSuites.find(suite => suite.metadata.version === '1.4.1');

export const benchmarkResults = currentV141Suite?.results || [];
export const testConfiguration = currentV141Suite?.metadata.testConfiguration || {
  rps: 10000,
  connections: 100,
  duration: 30,
  cpuLimit: '1000m',
  memoryLimit: '2000Mi'
};

// Generate backward-compatible processed data
export const performanceSummary = {
  totalTests: benchmarkResults.length,
  scaleUpTests: benchmarkResults.filter(r => r.phase === 'scaling-up').length,
  scaleDownTests: benchmarkResults.filter(r => r.phase === 'scaling-down').length,
  maxRoutes: benchmarkResults.length > 0 ? Math.max(...benchmarkResults.map(r => r.routes)) : 0,
  minRoutes: benchmarkResults.length > 0 ? Math.min(...benchmarkResults.map(r => r.routes)) : 0,
  avgThroughput: benchmarkResults.length > 0 ? benchmarkResults.reduce((sum, r) => sum + r.throughput, 0) / benchmarkResults.length : 0,
  avgLatency: benchmarkResults.length > 0 ? benchmarkResults.reduce((sum, r) => sum + r.latency.mean, 0) / benchmarkResults.length : 0
};

export const latencyPercentileComparison = benchmarkResults.map(result => ({
  routes: result.routes,
  phase: result.phase,
  p50: result.latency.percentiles.p50 / 1000, // convert to ms
  p75: result.latency.percentiles.p75 / 1000,
  p90: result.latency.percentiles.p90 / 1000,
  p95: result.latency.percentiles.p95 / 1000,
  p99: result.latency.percentiles.p99 / 1000,
  p999: result.latency.percentiles.p999 / 1000
}));

export const resourceTrends = benchmarkResults.map(result => ({
  routes: result.routes,
  phase: result.phase,
  envoyGatewayMemory: result.resources.envoyGateway.memory.mean,
  envoyGatewayCpu: result.resources.envoyGateway.cpu.mean,
  envoyProxyMemory: result.resources.envoyProxy.memory.mean,
  envoyProxyCpu: result.resources.envoyProxy.cpu.mean
}));

export const performanceMatrix = benchmarkResults.map(result => ({
  testName: result.testName,
  routes: result.routes,
  phase: result.phase,
  throughput: result.throughput,
  meanLatency: result.latency.mean / 1000, // convert to ms
  p95Latency: result.latency.percentiles.p95 / 1000, // convert to ms
  totalMemory: result.resources.envoyGateway.memory.mean + result.resources.envoyProxy.memory.mean,
  totalCpu: result.resources.envoyGateway.cpu.mean + result.resources.envoyProxy.cpu.mean
}));
