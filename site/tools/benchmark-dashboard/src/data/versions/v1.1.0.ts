import { TestSuite } from '../types';

// Benchmark data extracted from markdown report for version 1.1.0
// Generated on 2025-06-17T19:58:49.977Z

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.1.0",
    "runId": "1.1.0-1750190329977",
    "date": "2024-07-23",
    "environment": "GitHub CI",
    "description": "Benchmark report for Envoy Gateway 1.1.0",
    "downloadUrl": "https://github.com/envoyproxy/gateway/releases/download/v1.1.0/benchmark_report.zip",
    "testConfiguration": {
      "rps": 10000,
      "connections": 100,
      "duration": 30,
      "cpuLimit": "1000m",
      "memoryLimit": "2000Mi"
    }
  },
  "results": [
    {
      "testName": "scale-up-httproutes-10",
      "routes": 10,
      "routesPerHostname": 1,
      "phase": "scaling-up",
      "throughput": 6181.97,
      "totalRequests": 185463,
      "latency": {
        "min": 0.362,
        "mean": 5.902,
        "max": 73.084,
        "pstdev": 11.039,
        "percentiles": {
          "p50": 2.765,
          "p75": 4.364,
          "p80": 4.935,
          "p90": 7.504,
          "p95": 41.244,
          "p99": 53.929,
          "p999": 61.147
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 86.2,
            "max": 86.2,
            "mean": 86.2
          },
          "cpu": {
            "min": 1.4333333333333333,
            "max": 1.4333333333333333,
            "mean": 1.4333333333333333
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 24.09,
            "max": 24.09,
            "mean": 24.09
          },
          "cpu": {
            "min": 101.36666666666667,
            "max": 101.36666666666667,
            "mean": 101.36666666666667
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38
    },
    {
      "testName": "scale-up-httproutes-50",
      "routes": 50,
      "routesPerHostname": 1,
      "phase": "scaling-up",
      "throughput": 6103.97,
      "totalRequests": 183121,
      "latency": {
        "min": 0.354,
        "mean": 5.852,
        "max": 75.943,
        "pstdev": 11.382,
        "percentiles": {
          "p50": 2.681,
          "p75": 4.083,
          "p80": 4.601,
          "p90": 6.825,
          "p95": 46.262,
          "p99": 53.862,
          "p999": 60.651
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 100.14,
            "max": 100.14,
            "mean": 100.14
          },
          "cpu": {
            "min": 7.233333333333333,
            "max": 7.233333333333333,
            "mean": 7.233333333333333
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 32.25,
            "max": 32.25,
            "mean": 32.25
          },
          "cpu": {
            "min": 203.6,
            "max": 203.6,
            "mean": 203.6
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37
    },
    {
      "testName": "scale-up-httproutes-100",
      "routes": 100,
      "routesPerHostname": 1,
      "phase": "scaling-up",
      "throughput": 6074.87,
      "totalRequests": 182247,
      "latency": {
        "min": 0.371,
        "mean": 5.877,
        "max": 94.171,
        "pstdev": 11.399,
        "percentiles": {
          "p50": 2.732,
          "p75": 4.191,
          "p80": 4.677,
          "p90": 6.801,
          "p95": 46.481,
          "p99": 53.866,
          "p999": 61.036
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 114.69,
            "max": 114.69,
            "mean": 114.69
          },
          "cpu": {
            "min": 30,
            "max": 30,
            "mean": 30
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 46.42,
            "max": 46.42,
            "mean": 46.42
          },
          "cpu": {
            "min": 308.70000000000005,
            "max": 308.70000000000005,
            "mean": 308.70000000000005
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37
    },
    {
      "testName": "scale-up-httproutes-300",
      "routes": 300,
      "routesPerHostname": 1,
      "phase": "scaling-up",
      "throughput": 5993.57,
      "totalRequests": 179811,
      "latency": {
        "min": 0.368,
        "mean": 6.123,
        "max": 96.571,
        "pstdev": 11.831,
        "percentiles": {
          "p50": 2.753,
          "p75": 4.261,
          "p80": 4.799,
          "p90": 7.247,
          "p95": 47.368,
          "p99": 55.248,
          "p999": 66.586
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 762.8,
            "max": 762.8,
            "mean": 762.8
          },
          "cpu": {
            "min": 600.8666666666667,
            "max": 600.8666666666667,
            "mean": 600.8666666666667
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 152.78,
            "max": 152.78,
            "mean": 152.78
          },
          "cpu": {
            "min": 488.70000000000005,
            "max": 488.70000000000005,
            "mean": 488.70000000000005
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38
    },
    {
      "testName": "scale-up-httproutes-500",
      "routes": 500,
      "routesPerHostname": 1,
      "phase": "scaling-up",
      "throughput": 5812.8,
      "totalRequests": 174396,
      "latency": {
        "min": 0.388,
        "mean": 6.31,
        "max": 95.719,
        "pstdev": 12.246,
        "percentiles": {
          "p50": 2.729,
          "p75": 4.334,
          "p80": 4.905,
          "p90": 7.607,
          "p95": 47.97,
          "p99": 56.811,
          "p999": 68.165
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 1593.56,
            "max": 1593.56,
            "mean": 1593.56
          },
          "cpu": {
            "min": 36.56666666666667,
            "max": 36.56666666666667,
            "mean": 36.56666666666667
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 289.79,
            "max": 289.79,
            "mean": 289.79
          },
          "cpu": {
            "min": 715.6333333333333,
            "max": 715.6333333333333,
            "mean": 715.6333333333333
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38
    },
    {
      "testName": "scale-down-httproutes-10",
      "routes": 10,
      "routesPerHostname": 1,
      "phase": "scaling-down",
      "throughput": 5916.93,
      "totalRequests": 177517,
      "latency": {
        "min": 0.385,
        "mean": 6.352,
        "max": 82.747,
        "pstdev": 12.011,
        "percentiles": {
          "p50": 2.822,
          "p75": 4.461,
          "p80": 5.061,
          "p90": 7.769,
          "p95": 47.847,
          "p99": 55.068,
          "p999": 62.578
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 124.07,
            "max": 124.07,
            "mean": 124.07
          },
          "cpu": {
            "min": 519.6333333333332,
            "max": 519.6333333333332,
            "mean": 519.6333333333332
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 284.03,
            "max": 284.03,
            "mean": 284.03
          },
          "cpu": {
            "min": 1194.8666666666666,
            "max": 1194.8666666666666,
            "mean": 1194.8666666666666
          }
        }
      },
      "poolOverflow": 361,
      "upstreamConnections": 39
    },
    {
      "testName": "scale-down-httproutes-50",
      "routes": 50,
      "routesPerHostname": 1,
      "phase": "scaling-down",
      "throughput": 5876.35,
      "totalRequests": 176294,
      "latency": {
        "min": 0.371,
        "mean": 6.226,
        "max": 83.869,
        "pstdev": 11.816,
        "percentiles": {
          "p50": 2.77,
          "p75": 4.396,
          "p80": 5.0,
          "p90": 7.848,
          "p95": 47.163,
          "p99": 55.156,
          "p999": 64.434
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 131.62,
            "max": 131.62,
            "mean": 131.62
          },
          "cpu": {
            "min": 513.6,
            "max": 513.6,
            "mean": 513.6
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 291.73,
            "max": 291.73,
            "mean": 291.73
          },
          "cpu": {
            "min": 1092.9333333333334,
            "max": 1092.9333333333334,
            "mean": 1092.9333333333334
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38
    },
    {
      "testName": "scale-down-httproutes-100",
      "routes": 100,
      "routesPerHostname": 1,
      "phase": "scaling-down",
      "throughput": 4820.43,
      "totalRequests": 144617,
      "latency": {
        "min": 0.342,
        "mean": 7.088,
        "max": 172.031,
        "pstdev": 12.642,
        "percentiles": {
          "p50": 3.156,
          "p75": 5.48,
          "p80": 6.443,
          "p90": 12.007,
          "p95": 39.675,
          "p99": 63.842,
          "p999": 86.532
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 374.36,
            "max": 374.36,
            "mean": 374.36
          },
          "cpu": {
            "min": 462.8,
            "max": 462.8,
            "mean": 462.8
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 289.98,
            "max": 289.98,
            "mean": 289.98
          },
          "cpu": {
            "min": 980.4,
            "max": 980.4,
            "mean": 980.4
          }
        }
      },
      "poolOverflow": 364,
      "upstreamConnections": 36
    },
    {
      "testName": "scale-down-httproutes-300",
      "routes": 300,
      "routesPerHostname": 1,
      "phase": "scaling-down",
      "throughput": 5746.71,
      "totalRequests": 172402,
      "latency": {
        "min": 0.386,
        "mean": 6.313,
        "max": 101.535,
        "pstdev": 11.578,
        "percentiles": {
          "p50": 2.875,
          "p75": 4.579,
          "p80": 5.222,
          "p90": 8.617,
          "p95": 44.341,
          "p99": 54.996,
          "p999": 67.178
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 670.44,
            "max": 670.44,
            "mean": 670.44
          },
          "cpu": {
            "min": 24.333333333333332,
            "max": 24.333333333333332,
            "mean": 24.333333333333332
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 285.86,
            "max": 285.86,
            "mean": 285.86
          },
          "cpu": {
            "min": 824.5666666666667,
            "max": 824.5666666666667,
            "mean": 824.5666666666667
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38
    }
  ]
};

export default benchmarkData;
