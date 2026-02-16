import { TestSuite } from '../types';

// Benchmark data extracted from markdown report for version 1.1.1
// Generated on 2025-06-17T19:58:49.981Z

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.1.1",
    "runId": "1.1.1-1750190329981",
    "date": "2024-09-12",
    "environment": "GitHub CI",
    "description": "Benchmark report for Envoy Gateway 1.1.1",
    "downloadUrl": "https://github.com/envoyproxy/gateway/releases/download/v1.1.1/benchmark_report.zip",
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
      "throughput": 6258.99,
      "totalRequests": 187774,
      "latency": {
        "min": 333,
        "mean": 5933,
        "max": 71266,
        "pstdev": 10816,
        "percentiles": {
          "p50": 2898,
          "p75": 4513,
          "p80": 5099,
          "p90": 7661,
          "p95": 40607,
          "p99": 53149,
          "p999": 58789
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 84.04,
            "max": 84.04,
            "mean": 84.04
          },
          "cpu": {
            "min": 1.5333333333333334,
            "max": 1.5333333333333334,
            "mean": 1.5333333333333334
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 25.57,
            "max": 25.57,
            "mean": 25.57
          },
          "cpu": {
            "min": 101.33333333333331,
            "max": 101.33333333333331,
            "mean": 101.33333333333331
          }
        }
      },
      "poolOverflow": 361,
      "upstreamConnections": 39
    },
    {
      "testName": "scale-up-httproutes-50",
      "routes": 50,
      "routesPerHostname": 1,
      "phase": "scaling-up",
      "throughput": 6118.28,
      "totalRequests": 183552,
      "latency": {
        "min": 359,
        "mean": 5828,
        "max": 81178,
        "pstdev": 11437,
        "percentiles": {
          "p50": 2646,
          "p75": 4049,
          "p80": 4561,
          "p90": 6666,
          "p95": 46723,
          "p99": 53913,
          "p999": 60094
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 101.33,
            "max": 101.33,
            "mean": 101.33
          },
          "cpu": {
            "min": 6.8999999999999995,
            "max": 6.8999999999999995,
            "mean": 6.8999999999999995
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 33.75,
            "max": 33.75,
            "mean": 33.75
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
      "throughput": 6045.36,
      "totalRequests": 181361,
      "latency": {
        "min": 377,
        "mean": 5908,
        "max": 88063,
        "pstdev": 11539,
        "percentiles": {
          "p50": 2659,
          "p75": 4143,
          "p80": 4675,
          "p90": 6863,
          "p95": 46733,
          "p99": 54091,
          "p999": 63211
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 115.34,
            "max": 115.34,
            "mean": 115.34
          },
          "cpu": {
            "min": 32.800000000000004,
            "max": 32.800000000000004,
            "mean": 32.800000000000004
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 47.79,
            "max": 47.79,
            "mean": 47.79
          },
          "cpu": {
            "min": 308.96666666666664,
            "max": 308.96666666666664,
            "mean": 308.96666666666664
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
      "throughput": 5881.18,
      "totalRequests": 176436,
      "latency": {
        "min": 327,
        "mean": 6376,
        "max": 130330,
        "pstdev": 12048,
        "percentiles": {
          "p50": 2843,
          "p75": 4548,
          "p80": 5181,
          "p90": 8168,
          "p95": 47517,
          "p99": 55470,
          "p999": 66041
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 608.65,
            "max": 608.65,
            "mean": 608.65
          },
          "cpu": {
            "min": 625.6666666666666,
            "max": 625.6666666666666,
            "mean": 625.6666666666666
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 152.29,
            "max": 152.29,
            "mean": 152.29
          },
          "cpu": {
            "min": 492.5666666666667,
            "max": 492.5666666666667,
            "mean": 492.5666666666667
          }
        }
      },
      "poolOverflow": 361,
      "upstreamConnections": 39
    },
    {
      "testName": "scale-up-httproutes-500",
      "routes": 500,
      "routesPerHostname": 1,
      "phase": "scaling-up",
      "throughput": 5875.93,
      "totalRequests": 176280,
      "latency": {
        "min": 381,
        "mean": 6086,
        "max": 86654,
        "pstdev": 11932,
        "percentiles": {
          "p50": 2704,
          "p75": 4181,
          "p80": 4727,
          "p90": 7177,
          "p95": 47091,
          "p99": 55814,
          "p999": 68415
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 1308.52,
            "max": 1308.52,
            "mean": 1308.52
          },
          "cpu": {
            "min": 39.49999999999999,
            "max": 39.49999999999999,
            "mean": 39.49999999999999
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 283.24,
            "max": 283.24,
            "mean": 283.24
          },
          "cpu": {
            "min": 707.1333333333332,
            "max": 707.1333333333332,
            "mean": 707.1333333333332
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37
    },
    {
      "testName": "scale-down-httproutes-10",
      "routes": 10,
      "routesPerHostname": 1,
      "phase": "scaling-down",
      "throughput": 6077.56,
      "totalRequests": 182327,
      "latency": {
        "min": 362,
        "mean": 6197,
        "max": 92372,
        "pstdev": 11720,
        "percentiles": {
          "p50": 2824,
          "p75": 4438,
          "p80": 5000,
          "p90": 7503,
          "p95": 47110,
          "p99": 54169,
          "p999": 60426
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 121.13,
            "max": 121.13,
            "mean": 121.13
          },
          "cpu": {
            "min": 544.5333333333334,
            "max": 544.5333333333334,
            "mean": 544.5333333333334
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 277.29,
            "max": 277.29,
            "mean": 277.29
          },
          "cpu": {
            "min": 1206.4666666666665,
            "max": 1206.4666666666665,
            "mean": 1206.4666666666665
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
      "throughput": 6047.67,
      "totalRequests": 181438,
      "latency": {
        "min": 370,
        "mean": 6062,
        "max": 92168,
        "pstdev": 11743,
        "percentiles": {
          "p50": 2733,
          "p75": 4268,
          "p80": 4801,
          "p90": 7226,
          "p95": 47095,
          "p99": 54886,
          "p999": 61917
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 126.22,
            "max": 126.22,
            "mean": 126.22
          },
          "cpu": {
            "min": 538.8333333333334,
            "max": 538.8333333333334,
            "mean": 538.8333333333334
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 285.62,
            "max": 285.62,
            "mean": 285.62
          },
          "cpu": {
            "min": 1104.3000000000002,
            "max": 1104.3000000000002,
            "mean": 1104.3000000000002
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
      "throughput": 4654.8,
      "totalRequests": 139655,
      "latency": {
        "min": 336,
        "mean": 4675,
        "max": 107409,
        "pstdev": 9843,
        "percentiles": {
          "p50": 2048,
          "p75": 3459,
          "p80": 4015,
          "p90": 6531,
          "p95": 19138,
          "p99": 57518,
          "p999": 73269
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 480.36,
            "max": 480.36,
            "mean": 480.36
          },
          "cpu": {
            "min": 377,
            "max": 377,
            "mean": 377
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 289.51,
            "max": 289.51,
            "mean": 289.51
          },
          "cpu": {
            "min": 969.2333333333332,
            "max": 969.2333333333332,
            "mean": 969.2333333333332
          }
        }
      },
      "poolOverflow": 377,
      "upstreamConnections": 23
    },
    {
      "testName": "scale-down-httproutes-300",
      "routes": 300,
      "routesPerHostname": 1,
      "phase": "scaling-down",
      "throughput": 5676.68,
      "totalRequests": 170301,
      "latency": {
        "min": 376,
        "mean": 6244,
        "max": 153264,
        "pstdev": 11618,
        "percentiles": {
          "p50": 2797,
          "p75": 4618,
          "p80": 5324,
          "p90": 8807,
          "p95": 40232,
          "p99": 56496,
          "p999": 73506
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 436.58,
            "max": 436.58,
            "mean": 436.58
          },
          "cpu": {
            "min": 28.066666666666666,
            "max": 28.066666666666666,
            "mean": 28.066666666666666
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 283.47,
            "max": 283.47,
            "mean": 283.47
          },
          "cpu": {
            "min": 813.3666666666667,
            "max": 813.3666666666667,
            "mean": 813.3666666666667
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37
    }
  ]
};

export default benchmarkData;
