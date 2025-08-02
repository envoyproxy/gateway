import { useState, useMemo } from 'react';
import {
  allTestSuites,
  getAvailableVersions
} from '@/data';
import { TestSuite, TestResult, TestConfiguration } from '@/data/types';

interface UseVersionDataReturn {
  selectedVersion: string;
  setSelectedVersion: (version: string) => void;
  availableVersions: string[];
  benchmarkResults: TestResult[];
  testConfiguration: TestConfiguration;
  performanceSummary: any;
  latencyPercentileComparison: any[];
  resourceTrends: any[];
  performanceMatrix: any[];
  metadata: any;
}

export const useVersionData = (): UseVersionDataReturn => {
  const availableVersions = getAvailableVersions();
  const [selectedVersion, setSelectedVersion] = useState<string>(availableVersions[0] || '');

  // Get the current version's test suite
  const currentTestSuite = useMemo(() => {
    return allTestSuites.find(suite => suite.metadata.version === selectedVersion);
  }, [selectedVersion]);

  // Extract data in the same format as the old structure for backward compatibility
  const versionData = useMemo(() => {
    if (!currentTestSuite) {
      return {
        benchmarkResults: [],
        testConfiguration: {
          rps: 10000,
          connections: 100,
          duration: 30,
          cpuLimit: '1000m',
          memoryLimit: '2000Mi'
        },
        performanceSummary: {
          totalTests: 0,
          scaleUpTests: 0,
          scaleDownTests: 0,
          maxRoutes: 0,
          minRoutes: 0,
          avgThroughput: 0,
          avgLatency: 0
        },
        latencyPercentileComparison: [],
        resourceTrends: [],
        performanceMatrix: [],
        metadata: null
      };
    }

    const results = currentTestSuite.results;

    return {
      benchmarkResults: results,
      testConfiguration: currentTestSuite.metadata.testConfiguration,
      performanceSummary: {
        totalTests: results.length,
        scaleUpTests: results.filter(r => r.phase === 'scaling-up').length,
        scaleDownTests: results.filter(r => r.phase === 'scaling-down').length,
        maxRoutes: results.length > 0 ? Math.max(...results.map(r => r.routes)) : 0,
        minRoutes: results.length > 0 ? Math.min(...results.map(r => r.routes)) : 0,
        avgThroughput: results.length > 0 ? results.reduce((sum, r) => sum + r.throughput, 0) / results.length : 0,
        avgLatency: results.length > 0 ? results.reduce((sum, r) => sum + r.latency.mean, 0) / results.length : 0
      },
      latencyPercentileComparison: results.map(result => ({
        routes: result.routes,
        phase: result.phase,
        p50: result.latency.percentiles.p50 / 1000, // convert to ms
        p75: result.latency.percentiles.p75 / 1000,
        p90: result.latency.percentiles.p90 / 1000,
        p95: result.latency.percentiles.p95 / 1000,
        p99: result.latency.percentiles.p99 / 1000,
        p999: result.latency.percentiles.p999 / 1000
      })),
      resourceTrends: results.map(result => ({
        routes: result.routes,
        phase: result.phase,
        envoyGatewayMemory: result.resources.envoyGateway.memory.mean,
        envoyGatewayCpu: result.resources.envoyGateway.cpu.mean,
        envoyProxyMemory: result.resources.envoyProxy.memory.mean,
        envoyProxyCpu: result.resources.envoyProxy.cpu.mean
      })),
      performanceMatrix: results.map(result => ({
        testName: result.testName,
        routes: result.routes,
        phase: result.phase,
        throughput: result.throughput,
        meanLatency: result.latency.mean / 1000, // convert to ms
        p95Latency: result.latency.percentiles.p95 / 1000, // convert to ms
        totalMemory: result.resources.envoyGateway.memory.mean + result.resources.envoyProxy.memory.mean,
        totalCpu: result.resources.envoyGateway.cpu.mean + result.resources.envoyProxy.cpu.mean
      })),
      metadata: currentTestSuite.metadata
    };
  }, [currentTestSuite]);

  return {
    selectedVersion,
    setSelectedVersion,
    availableVersions,
    ...versionData
  };
};
