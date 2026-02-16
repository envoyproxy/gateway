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
        "min": 353,
        "mean": 6263,
        "max": 87117,
        "pstdev": 11274,
        "percentiles": {
          "p50": 2994,
          "p75": 4783,
          "p80": 5441,
          "p90": 8506,
          "p95": 44029,
          "p99": 54130,
          "p999": 59291
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
        "min": 368,
        "mean": 6117,
        "max": 76840,
        "pstdev": 11675,
        "percentiles": {
          "p50": 2803,
          "p75": 4339,
          "p80": 4903,
          "p90": 7204,
          "p95": 47128,
          "p99": 54317,
          "p999": 60557
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
        "min": 390,
        "mean": 5889,
        "max": 94826,
        "pstdev": 11642,
        "percentiles": {
          "p50": 2599,
          "p75": 4075,
          "p80": 4583,
          "p90": 6945,
          "p95": 47128,
          "p99": 54872,
          "p999": 62212
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
        "min": 378,
        "mean": 6220,
        "max": 100757,
        "pstdev": 12033,
        "percentiles": {
          "p50": 2723,
          "p75": 4341,
          "p80": 4931,
          "p90": 7644,
          "p95": 47828,
          "p99": 55521,
          "p999": 67383
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
        "min": 356,
        "mean": 6109,
        "max": 107913,
        "pstdev": 11285,
        "percentiles": {
          "p50": 2832,
          "p75": 4538,
          "p80": 5163,
          "p90": 8134,
          "p95": 42639,
          "p99": 54546,
          "p999": 67543
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
        "min": 367,
        "mean": 6127,
        "max": 77688,
        "pstdev": 11777,
        "percentiles": {
          "p50": 2745,
          "p75": 4292,
          "p80": 4835,
          "p90": 7317,
          "p95": 47597,
          "p99": 54706,
          "p999": 61323
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
        "min": 336,
        "mean": 4876,
        "max": 76279,
        "pstdev": 10072,
        "percentiles": {
          "p50": 2290,
          "p75": 3444,
          "p80": 3846,
          "p90": 5476,
          "p95": 24546,
          "p99": 51763,
          "p999": 57040
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
        "min": 362,
        "mean": 6219,
        "max": 144302,
        "pstdev": 11133,
        "percentiles": {
          "p50": 2808,
          "p75": 4741,
          "p80": 5579,
          "p90": 10689,
          "p95": 37480,
          "p99": 53626,
          "p999": 73998
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
        "min": 369,
        "mean": 6266,
        "max": 106577,
        "pstdev": 11127,
        "percentiles": {
          "p50": 2908,
          "p75": 4793,
          "p80": 5511,
          "p90": 9568,
          "p95": 36169,
          "p99": 54472,
          "p999": 68628
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
