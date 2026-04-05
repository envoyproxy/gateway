// Template for new version data - Copy this file and rename it
// Example: v1.6.0.ts, v1.7.0.ts, main-branch.ts, etc.

import { TestSuite } from '../types';

export const vXXXTestSuite: TestSuite = {
  metadata: {
    version: '1.X.X', // Update this
    runId: 'v1.X.X-description-YYYY-MM-DD', // Update this with unique ID
    date: '2024-XX-XXTXX:XX:XXZ', // Update with actual test date
    environment: 'test', // or 'staging', 'prod', etc.
    description: 'Description of this benchmark run',
    gitCommit: 'abc123...', // Optional: git commit hash
    testConfiguration: {
      rps: 10000, // Update if different
      connections: 100, // Update if different
      duration: 30, // Update if different
      cpuLimit: '1000m', // Update if different
      memoryLimit: '2000Mi' // Update if different
    }
  },
  results: [
    // Copy your test results here in the same format as v1.4.1
    // Each test result should follow the TestResult interface
    {
      testName: 'scaling up httproutes to 10 with 2 routes per hostname',
      routes: 10,
      routesPerHostname: 2,
      phase: 'scaling-up',
      throughput: 0, // Replace with actual data
      totalRequests: 0, // Replace with actual data
      latency: {
        min: 0, // Replace with actual data
        mean: 0, // Replace with actual data
        max: 0, // Replace with actual data
        pstdev: 0, // Replace with actual data
        percentiles: {
          p50: 0, // Replace with actual data
          p75: 0, // Replace with actual data
          p80: 0, // Replace with actual data
          p90: 0, // Replace with actual data
          p95: 0, // Replace with actual data
          p99: 0, // Replace with actual data
          p999: 0 // Replace with actual data
        }
      },
      resources: {
        envoyGateway: {
          memory: { min: 0, max: 0, mean: 0 }, // Replace with actual data
          cpu: { min: 0, max: 0, mean: 0 } // Replace with actual data
        },
        envoyProxy: {
          memory: { min: 0, max: 0, mean: 0 }, // Replace with actual data
          cpu: { min: 0, max: 0, mean: 0 } // Replace with actual data
        }
      },
      poolOverflow: 0, // Replace with actual data
      upstreamConnections: 0 // Replace with actual data
    }
    // Add more test results...
  ]
};

/*
STEPS TO ADD A NEW VERSION:

1. Copy this template file to a new file named after your version (e.g., v1.6.0.ts)

2. Update the export name (e.g., export const v160TestSuite)

3. Fill in the metadata:
   - version: The version string (e.g., '1.5.0')
   - runId: Unique identifier for this run
   - date: ISO date string when tests were run
   - description: Human-readable description
   - testConfiguration: Update if test parameters changed

4. Replace all the placeholder data (000.00, 0000, etc.) with your actual test results

5. Add your new export to src/data/index.ts:
   import { v160TestSuite } from './versions/v1.6.0';

   export const allTestSuites: TestSuite[] = [
     v133TestSuite,
     v141TestSuite,
     v160TestSuite, // Add your new suite here
   ];

6. Your new data will automatically be available in all comparison functions!
*/
