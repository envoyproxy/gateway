import { TestSuite } from '../types';

// Benchmark data extracted from markdown report for version 1.2.2
// Generated on 2025-06-17T19:50:26.761Z

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.2.2",
    "runId": "1.2.2-1750189826761",
    "date": "2024-11-28",
    "environment": "GitHub CI",
    "description": "Benchmark report for Envoy Gateway 1.2.2",
    "downloadUrl": "https://github.com/envoyproxy/gateway/releases/download/v1.2.2/benchmark_report.zip",
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
      "throughput": 6249.74,
      "totalRequests": 187495,
      "latency": {
        "min": 333,
        "mean": 5937,
        "max": 75075,
        "pstdev": 10895,
        "percentiles": {
          "p50": 2840,
          "p75": 4501,
          "p80": 5077,
          "p90": 7640,
          "p95": 41435,
          "p99": 53243,
          "p999": 58066
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 106.26,
            "max": 106.26,
            "mean": 106.26
          },
          "cpu": {
            "min": 0.73,
            "max": 0.73,
            "mean": 0.73
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 26.09,
            "max": 26.09,
            "mean": 26.09
          },
          "cpu": {
            "min": 30.41,
            "max": 30.41,
            "mean": 30.41
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
      "throughput": 6140.58,
      "totalRequests": 184222,
      "latency": {
        "min": 373,
        "mean": 5812,
        "max": 79740,
        "pstdev": 11513,
        "percentiles": {
          "p50": 2590,
          "p75": 4071,
          "p80": 4583,
          "p90": 6812,
          "p95": 46811,
          "p99": 54179,
          "p999": 61151
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 131.24,
            "max": 131.24,
            "mean": 131.24
          },
          "cpu": {
            "min": 1.34,
            "max": 1.34,
            "mean": 1.34
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 32.26,
            "max": 32.26,
            "mean": 32.26
          },
          "cpu": {
            "min": 60.74,
            "max": 60.74,
            "mean": 60.74
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
      "throughput": 6149.72,
      "totalRequests": 184494,
      "latency": {
        "min": 366,
        "mean": 5813,
        "max": 89935,
        "pstdev": 11434,
        "percentiles": {
          "p50": 2637,
          "p75": 4073,
          "p80": 4561,
          "p90": 6680,
          "p95": 46505,
          "p99": 53872,
          "p999": 60348
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 133.05,
            "max": 133.05,
            "mean": 133.05
          },
          "cpu": {
            "min": 2.75,
            "max": 2.75,
            "mean": 2.75
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 38.38,
            "max": 38.38,
            "mean": 38.38
          },
          "cpu": {
            "min": 91.57,
            "max": 91.57,
            "mean": 91.57
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
      "throughput": 6076.78,
      "totalRequests": 182305,
      "latency": {
        "min": 361,
        "mean": 6031,
        "max": 87502,
        "pstdev": 11819,
        "percentiles": {
          "p50": 2675,
          "p75": 4203,
          "p80": 4751,
          "p90": 7118,
          "p95": 47147,
          "p99": 55078,
          "p999": 67106
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 153.62,
            "max": 153.62,
            "mean": 153.62
          },
          "cpu": {
            "min": 14.77,
            "max": 14.77,
            "mean": 14.77
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 60.45,
            "max": 60.45,
            "mean": 60.45
          },
          "cpu": {
            "min": 124.04,
            "max": 124.04,
            "mean": 124.04
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
      "throughput": 5928.89,
      "totalRequests": 177873,
      "latency": {
        "min": 336,
        "mean": 6196,
        "max": 85643,
        "pstdev": 12032,
        "percentiles": {
          "p50": 2803,
          "p75": 4338,
          "p80": 4852,
          "p90": 7143,
          "p95": 47726,
          "p99": 55619,
          "p999": 67780
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 145.79,
            "max": 145.79,
            "mean": 145.79
          },
          "cpu": {
            "min": 27.17,
            "max": 27.17,
            "mean": 27.17
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 80.64,
            "max": 80.64,
            "mean": 80.64
          },
          "cpu": {
            "min": 157.65,
            "max": 157.65,
            "mean": 157.65
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38
    },
    {
      "testName": "scale-up-httproutes-1000",
      "routes": 1000,
      "routesPerHostname": 1,
      "phase": "scaling-up",
      "throughput": 5843.64,
      "totalRequests": 175314,
      "latency": {
        "min": 339,
        "mean": 6452,
        "max": 101912,
        "pstdev": 12482,
        "percentiles": {
          "p50": 2790,
          "p75": 4509,
          "p80": 5136,
          "p90": 8012,
          "p95": 47702,
          "p99": 58542,
          "p999": 74014
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 173.7,
            "max": 173.7,
            "mean": 173.7
          },
          "cpu": {
            "min": 61.12,
            "max": 61.12,
            "mean": 61.12
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 130.89,
            "max": 130.89,
            "mean": 130.89
          },
          "cpu": {
            "min": 195.74,
            "max": 195.74,
            "mean": 195.74
          }
        }
      },
      "poolOverflow": 361,
      "upstreamConnections": 39
    },
    {
      "testName": "scale-down-httproutes-10",
      "routes": 10,
      "routesPerHostname": 1,
      "phase": "scaling-down",
      "throughput": 6096.9,
      "totalRequests": 182907,
      "latency": {
        "min": 365,
        "mean": 5840,
        "max": 99377,
        "pstdev": 11326,
        "percentiles": {
          "p50": 2735,
          "p75": 4188,
          "p80": 4674,
          "p90": 6654,
          "p95": 46344,
          "p99": 53444,
          "p999": 58591
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 230.99,
            "max": 230.99,
            "mean": 230.99
          },
          "cpu": {
            "min": 114.22,
            "max": 114.22,
            "mean": 114.22
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 121.56,
            "max": 121.56,
            "mean": 121.56
          },
          "cpu": {
            "min": 358.71,
            "max": 358.71,
            "mean": 358.71
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37
    },
    {
      "testName": "scale-down-httproutes-50",
      "routes": 50,
      "routesPerHostname": 1,
      "phase": "scaling-down",
      "throughput": 6151.76,
      "totalRequests": 184556,
      "latency": {
        "min": 376,
        "mean": 5805,
        "max": 84885,
        "pstdev": 11572,
        "percentiles": {
          "p50": 2520,
          "p75": 3969,
          "p80": 4498,
          "p90": 6842,
          "p95": 46837,
          "p99": 54761,
          "p999": 62570
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 233.63,
            "max": 233.63,
            "mean": 233.63
          },
          "cpu": {
            "min": 113.65,
            "max": 113.65,
            "mean": 113.65
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 121.86,
            "max": 121.86,
            "mean": 121.86
          },
          "cpu": {
            "min": 328.3,
            "max": 328.3,
            "mean": 328.3
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37
    },
    {
      "testName": "scale-down-httproutes-100",
      "routes": 100,
      "routesPerHostname": 1,
      "phase": "scaling-down",
      "throughput": 6157.53,
      "totalRequests": 184729,
      "latency": {
        "min": 379,
        "mean": 5964,
        "max": 75988,
        "pstdev": 11502,
        "percentiles": {
          "p50": 2743,
          "p75": 4204,
          "p80": 4695,
          "p90": 6941,
          "p95": 46497,
          "p99": 54179,
          "p999": 62005
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 141.36,
            "max": 141.36,
            "mean": 141.36
          },
          "cpu": {
            "min": 112.67,
            "max": 112.67,
            "mean": 112.67
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 122.04,
            "max": 122.04,
            "mean": 122.04
          },
          "cpu": {
            "min": 297.59,
            "max": 297.59,
            "mean": 297.59
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38
    },
    {
      "testName": "scale-down-httproutes-300",
      "routes": 300,
      "routesPerHostname": 1,
      "phase": "scaling-down",
      "throughput": 6060.5,
      "totalRequests": 181818,
      "latency": {
        "min": 360,
        "mean": 6022,
        "max": 92483,
        "pstdev": 11657,
        "percentiles": {
          "p50": 2685,
          "p75": 4265,
          "p80": 4811,
          "p90": 7313,
          "p95": 46495,
          "p99": 54495,
          "p999": 64219
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 150.71,
            "max": 150.71,
            "mean": 150.71
          },
          "cpu": {
            "min": 102.4,
            "max": 102.4,
            "mean": 102.4
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 130.9,
            "max": 130.9,
            "mean": 130.9
          },
          "cpu": {
            "min": 265.29,
            "max": 265.29,
            "mean": 265.29
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38
    },
    {
      "testName": "scale-down-httproutes-500",
      "routes": 500,
      "routesPerHostname": 1,
      "phase": "scaling-down",
      "throughput": 5966.83,
      "totalRequests": 179008,
      "latency": {
        "min": 343,
        "mean": 5962,
        "max": 102715,
        "pstdev": 11763,
        "percentiles": {
          "p50": 2668,
          "p75": 4125,
          "p80": 4633,
          "p90": 6863,
          "p95": 46741,
          "p99": 54929,
          "p999": 67018
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 162.62,
            "max": 162.62,
            "mean": 162.62
          },
          "cpu": {
            "min": 91.03,
            "max": 91.03,
            "mean": 91.03
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 131.05,
            "max": 131.05,
            "mean": 131.05
          },
          "cpu": {
            "min": 232.55,
            "max": 232.55,
            "mean": 232.55
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37
    }
  ]
};

export default benchmarkData;
