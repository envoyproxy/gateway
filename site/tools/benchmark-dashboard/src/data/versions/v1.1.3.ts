import { TestSuite } from '../types';

// Benchmark data extracted from markdown report for version 1.1.3
// Generated on 2025-06-17T19:58:49.984Z

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.1.3",
    "runId": "1.1.3-1750190329984",
    "date": "2024-11-04",
    "environment": "GitHub CI",
    "description": "Benchmark report for Envoy Gateway 1.1.3",
    "downloadUrl": "https://github.com/envoyproxy/gateway/releases/download/v1.1.3/benchmark_report.zip",
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
      "throughput": 6070.53,
      "totalRequests": 182116,
      "latency": {
        "min": 0.353,
        "mean": 6.263,
        "max": 87.117,
        "pstdev": 11.274,
        "percentiles": {
          "p50": 2.994,
          "p75": 4.783,
          "p80": 5.441,
          "p90": 8.506,
          "p95": 44.029,
          "p99": 54.13,
          "p999": 59.291
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 85.16,
            "max": 85.16,
            "mean": 85.16
          },
          "cpu": {
            "min": 1.3666666666666665,
            "max": 1.3666666666666665,
            "mean": 1.3666666666666665
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 26,
            "max": 26,
            "mean": 26
          },
          "cpu": {
            "min": 101.43333333333334,
            "max": 101.43333333333334,
            "mean": 101.43333333333334
          }
        }
      },
      "poolOverflow": 360,
      "upstreamConnections": 40
    },
    {
      "testName": "scale-up-httproutes-50",
      "routes": 50,
      "routesPerHostname": 1,
      "phase": "scaling-up",
      "throughput": 5999.45,
      "totalRequests": 179985,
      "latency": {
        "min": 0.368,
        "mean": 6.117,
        "max": 76.84,
        "pstdev": 11.675,
        "percentiles": {
          "p50": 2.803,
          "p75": 4.339,
          "p80": 4.903,
          "p90": 7.204,
          "p95": 47.128,
          "p99": 54.317,
          "p999": 60.557
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 102.18,
            "max": 102.18,
            "mean": 102.18
          },
          "cpu": {
            "min": 7.033333333333333,
            "max": 7.033333333333333,
            "mean": 7.033333333333333
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 34.18,
            "max": 34.18,
            "mean": 34.18
          },
          "cpu": {
            "min": 203.56666666666666,
            "max": 203.56666666666666,
            "mean": 203.56666666666666
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38
    },
    {
      "testName": "scale-up-httproutes-100",
      "routes": 100,
      "routesPerHostname": 1,
      "phase": "scaling-up",
      "throughput": 6053.06,
      "totalRequests": 181592,
      "latency": {
        "min": 0.39,
        "mean": 5.889,
        "max": 94.826,
        "pstdev": 11.642,
        "percentiles": {
          "p50": 2.599,
          "p75": 4.075,
          "p80": 4.583,
          "p90": 6.945,
          "p95": 47.128,
          "p99": 54.872,
          "p999": 62.212
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 151.34,
            "max": 151.34,
            "mean": 151.34
          },
          "cpu": {
            "min": 31.633333333333336,
            "max": 31.633333333333336,
            "mean": 31.633333333333336
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 48.34,
            "max": 48.34,
            "mean": 48.34
          },
          "cpu": {
            "min": 309.06666666666666,
            "max": 309.06666666666666,
            "mean": 309.06666666666666
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
      "throughput": 5895.77,
      "totalRequests": 176873,
      "latency": {
        "min": 0.378,
        "mean": 6.22,
        "max": 100.757,
        "pstdev": 12.033,
        "percentiles": {
          "p50": 2.723,
          "p75": 4.341,
          "p80": 4.931,
          "p90": 7.644,
          "p95": 47.828,
          "p99": 55.521,
          "p999": 67.383
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 687.16,
            "max": 687.16,
            "mean": 687.16
          },
          "cpu": {
            "min": 578.2333333333332,
            "max": 578.2333333333332,
            "mean": 578.2333333333332
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 150.69,
            "max": 150.69,
            "mean": 150.69
          },
          "cpu": {
            "min": 488.20000000000005,
            "max": 488.20000000000005,
            "mean": 488.20000000000005
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
      "throughput": 5818.09,
      "totalRequests": 174543,
      "latency": {
        "min": 0.356,
        "mean": 6.109,
        "max": 107.913,
        "pstdev": 11.285,
        "percentiles": {
          "p50": 2.832,
          "p75": 4.538,
          "p80": 5.163,
          "p90": 8.134,
          "p95": 42.639,
          "p99": 54.546,
          "p999": 67.543
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 1329.35,
            "max": 1329.35,
            "mean": 1329.35
          },
          "cpu": {
            "min": 61.33333333333333,
            "max": 61.33333333333333,
            "mean": 61.33333333333333
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 303.55,
            "max": 303.55,
            "mean": 303.55
          },
          "cpu": {
            "min": 731.4333333333334,
            "max": 731.4333333333334,
            "mean": 731.4333333333334
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
      "throughput": 5996.79,
      "totalRequests": 179907,
      "latency": {
        "min": 0.367,
        "mean": 6.127,
        "max": 77.688,
        "pstdev": 11.777,
        "percentiles": {
          "p50": 2.745,
          "p75": 4.292,
          "p80": 4.835,
          "p90": 7.317,
          "p95": 47.597,
          "p99": 54.706,
          "p999": 61.323
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 125.62,
            "max": 125.62,
            "mean": 125.62
          },
          "cpu": {
            "min": 525.5,
            "max": 525.5,
            "mean": 525.5
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 289.68,
            "max": 289.68,
            "mean": 289.68
          },
          "cpu": {
            "min": 1239.6,
            "max": 1239.6,
            "mean": 1239.6
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38
    },
    {
      "testName": "scale-down-httproutes-50",
      "routes": 50,
      "routesPerHostname": 1,
      "phase": "scaling-down",
      "throughput": 5732.14,
      "totalRequests": 171956,
      "latency": {
        "min": 0.336,
        "mean": 4.876,
        "max": 76.279,
        "pstdev": 10.072,
        "percentiles": {
          "p50": 2.29,
          "p75": 3.444,
          "p80": 3.846,
          "p90": 5.476,
          "p95": 24.546,
          "p99": 51.763,
          "p999": 57.04
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 132.64,
            "max": 132.64,
            "mean": 132.64
          },
          "cpu": {
            "min": 520.8333333333333,
            "max": 520.8333333333333,
            "mean": 520.8333333333333
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 291.68,
            "max": 291.68,
            "mean": 291.68
          },
          "cpu": {
            "min": 1138.0333333333335,
            "max": 1138.0333333333335,
            "mean": 1138.0333333333335
          }
        }
      },
      "poolOverflow": 371,
      "upstreamConnections": 29
    },
    {
      "testName": "scale-down-httproutes-100",
      "routes": 100,
      "routesPerHostname": 1,
      "phase": "scaling-down",
      "throughput": 5311.34,
      "totalRequests": 159332,
      "latency": {
        "min": 0.362,
        "mean": 6.219,
        "max": 144.302,
        "pstdev": 11.133,
        "percentiles": {
          "p50": 2.808,
          "p75": 4.741,
          "p80": 5.579,
          "p90": 10.689,
          "p95": 37.48,
          "p99": 53.626,
          "p999": 73.998
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 179.01,
            "max": 179.01,
            "mean": 179.01
          },
          "cpu": {
            "min": 497.9666666666666,
            "max": 497.9666666666666,
            "mean": 497.9666666666666
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 298.77,
            "max": 298.77,
            "mean": 298.77
          },
          "cpu": {
            "min": 1034.033333333333,
            "max": 1034.033333333333,
            "mean": 1034.033333333333
          }
        }
      },
      "poolOverflow": 365,
      "upstreamConnections": 35
    },
    {
      "testName": "scale-down-httproutes-300",
      "routes": 300,
      "routesPerHostname": 1,
      "phase": "scaling-down",
      "throughput": 5633.85,
      "totalRequests": 169021,
      "latency": {
        "min": 0.369,
        "mean": 6.266,
        "max": 106.577,
        "pstdev": 11.127,
        "percentiles": {
          "p50": 2.908,
          "p75": 4.793,
          "p80": 5.511,
          "p90": 9.568,
          "p95": 36.169,
          "p99": 54.472,
          "p999": 68.628
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 487.12,
            "max": 487.12,
            "mean": 487.12
          },
          "cpu": {
            "min": 27.133333333333336,
            "max": 27.133333333333336,
            "mean": 27.133333333333336
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 299.84,
            "max": 299.84,
            "mean": 299.84
          },
          "cpu": {
            "min": 837.8666666666668,
            "max": 837.8666666666668,
            "mean": 837.8666666666668
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37
    }
  ]
};

export default benchmarkData;
