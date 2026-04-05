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
        "min": 0.333,
        "mean": 5.937,
        "max": 75.075,
        "pstdev": 10.895,
        "percentiles": {
          "p50": 2.84,
          "p75": 4.501,
          "p80": 5.077,
          "p90": 7.64,
          "p95": 41.435,
          "p99": 53.243,
          "p999": 58.066
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
        "min": 0.373,
        "mean": 5.812,
        "max": 79.74,
        "pstdev": 11.513,
        "percentiles": {
          "p50": 2.59,
          "p75": 4.071,
          "p80": 4.583,
          "p90": 6.812,
          "p95": 46.811,
          "p99": 54.179,
          "p999": 61.151
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
        "min": 0.366,
        "mean": 5.813,
        "max": 89.935,
        "pstdev": 11.434,
        "percentiles": {
          "p50": 2.637,
          "p75": 4.073,
          "p80": 4.561,
          "p90": 6.68,
          "p95": 46.505,
          "p99": 53.872,
          "p999": 60.348
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
        "min": 0.361,
        "mean": 6.031,
        "max": 87.502,
        "pstdev": 11.819,
        "percentiles": {
          "p50": 2.675,
          "p75": 4.203,
          "p80": 4.751,
          "p90": 7.118,
          "p95": 47.147,
          "p99": 55.078,
          "p999": 67.106
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
        "min": 0.336,
        "mean": 6.196,
        "max": 85.643,
        "pstdev": 12.032,
        "percentiles": {
          "p50": 2.803,
          "p75": 4.338,
          "p80": 4.852,
          "p90": 7.143,
          "p95": 47.726,
          "p99": 55.619,
          "p999": 67.78
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
        "min": 0.339,
        "mean": 6.452,
        "max": 101.912,
        "pstdev": 12.482,
        "percentiles": {
          "p50": 2.79,
          "p75": 4.509,
          "p80": 5.136,
          "p90": 8.012,
          "p95": 47.702,
          "p99": 58.542,
          "p999": 74.014
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
        "min": 0.365,
        "mean": 5.84,
        "max": 99.377,
        "pstdev": 11.326,
        "percentiles": {
          "p50": 2.735,
          "p75": 4.188,
          "p80": 4.674,
          "p90": 6.654,
          "p95": 46.344,
          "p99": 53.444,
          "p999": 58.591
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
        "min": 0.376,
        "mean": 5.805,
        "max": 84.885,
        "pstdev": 11.572,
        "percentiles": {
          "p50": 2.52,
          "p75": 3.969,
          "p80": 4.498,
          "p90": 6.842,
          "p95": 46.837,
          "p99": 54.761,
          "p999": 62.57
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
        "min": 0.379,
        "mean": 5.964,
        "max": 75.988,
        "pstdev": 11.502,
        "percentiles": {
          "p50": 2.743,
          "p75": 4.204,
          "p80": 4.695,
          "p90": 6.941,
          "p95": 46.497,
          "p99": 54.179,
          "p999": 62.005
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
        "min": 0.36,
        "mean": 6.022,
        "max": 92.483,
        "pstdev": 11.657,
        "percentiles": {
          "p50": 2.685,
          "p75": 4.265,
          "p80": 4.811,
          "p90": 7.313,
          "p95": 46.495,
          "p99": 54.495,
          "p999": 64.219
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
        "min": 0.343,
        "mean": 5.962,
        "max": 102.715,
        "pstdev": 11.763,
        "percentiles": {
          "p50": 2.668,
          "p75": 4.125,
          "p80": 4.633,
          "p90": 6.863,
          "p95": 46.741,
          "p99": 54.929,
          "p999": 67.018
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
