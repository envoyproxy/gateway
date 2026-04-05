import { TestSuite } from '../types';

// Benchmark data extracted from markdown report for version 1.3.0
// Generated on 2025-06-17T19:50:26.773Z

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.3.0",
    "runId": "1.3.0-1750189826772",
    "date": "2025-01-31",
    "environment": "GitHub CI",
    "description": "Benchmark report for Envoy Gateway 1.3.0",
    "downloadUrl": "https://github.com/envoyproxy/gateway/releases/download/v1.3.0/benchmark_report.zip",
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
      "throughput": 5297.23,
      "totalRequests": 158917,
      "latency": {
        "min": 0.371,
        "mean": 6.861,
        "max": 69.562,
        "pstdev": 12.096,
        "percentiles": {
          "p50": 3.223,
          "p75": 5.219,
          "p80": 5.926,
          "p90": 9.36,
          "p95": 48.029,
          "p99": 54.945,
          "p999": 60.096
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 116.36,
            "max": 116.36,
            "mean": 116.36
          },
          "cpu": {
            "min": 0.77,
            "max": 0.77,
            "mean": 0.77
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 25.3,
            "max": 25.3,
            "mean": 25.3
          },
          "cpu": {
            "min": 30.42,
            "max": 30.42,
            "mean": 30.42
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
      "throughput": 5276.06,
      "totalRequests": 158285,
      "latency": {
        "min": 0.372,
        "mean": 6.781,
        "max": 81.436,
        "pstdev": 12.685,
        "percentiles": {
          "p50": 3.035,
          "p75": 4.783,
          "p80": 5.41,
          "p90": 8.21,
          "p95": 50.042,
          "p99": 55.861,
          "p999": 61.739
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 135.97,
            "max": 135.97,
            "mean": 135.97
          },
          "cpu": {
            "min": 1.53,
            "max": 1.53,
            "mean": 1.53
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 31.45,
            "max": 31.45,
            "mean": 31.45
          },
          "cpu": {
            "min": 60.97,
            "max": 60.97,
            "mean": 60.97
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
      "throughput": 5248.15,
      "totalRequests": 157445,
      "latency": {
        "min": 0.386,
        "mean": 6.832,
        "max": 83.939,
        "pstdev": 12.879,
        "percentiles": {
          "p50": 2.957,
          "p75": 4.826,
          "p80": 5.504,
          "p90": 8.587,
          "p95": 50.454,
          "p99": 56.524,
          "p999": 63.15
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 128.67,
            "max": 128.67,
            "mean": 128.67
          },
          "cpu": {
            "min": 2.88,
            "max": 2.88,
            "mean": 2.88
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 35.62,
            "max": 35.62,
            "mean": 35.62
          },
          "cpu": {
            "min": 91.69,
            "max": 91.69,
            "mean": 91.69
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38
    },
    {
      "testName": "scale-up-httproutes-300",
      "routes": 300,
      "routesPerHostname": 1,
      "phase": "scaling-up",
      "throughput": 5190.05,
      "totalRequests": 155708,
      "latency": {
        "min": 0.374,
        "mean": 6.735,
        "max": 105.82,
        "pstdev": 12.667,
        "percentiles": {
          "p50": 2.997,
          "p75": 4.715,
          "p80": 5.323,
          "p90": 8.205,
          "p95": 49.692,
          "p99": 56.199,
          "p999": 64.403
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 147.27,
            "max": 147.27,
            "mean": 147.27
          },
          "cpu": {
            "min": 15.8,
            "max": 15.8,
            "mean": 15.8
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 61.52,
            "max": 61.52,
            "mean": 61.52
          },
          "cpu": {
            "min": 124.52,
            "max": 124.52,
            "mean": 124.52
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37
    },
    {
      "testName": "scale-up-httproutes-500",
      "routes": 500,
      "routesPerHostname": 1,
      "phase": "scaling-up",
      "throughput": 5171.63,
      "totalRequests": 155149,
      "latency": {
        "min": 0.345,
        "mean": 7.139,
        "max": 89.923,
        "pstdev": 13.21,
        "percentiles": {
          "p50": 3.191,
          "p75": 5.039,
          "p80": 5.68,
          "p90": 8.824,
          "p95": 50.767,
          "p99": 57.397,
          "p999": 67.665
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 158.88,
            "max": 158.88,
            "mean": 158.88
          },
          "cpu": {
            "min": 28.37,
            "max": 28.37,
            "mean": 28.37
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 76,
            "max": 76,
            "mean": 76
          },
          "cpu": {
            "min": 157.59,
            "max": 157.59,
            "mean": 157.59
          }
        }
      },
      "poolOverflow": 361,
      "upstreamConnections": 39
    },
    {
      "testName": "scale-up-httproutes-1000",
      "routes": 1000,
      "routesPerHostname": 1,
      "phase": "scaling-up",
      "throughput": 4998.06,
      "totalRequests": 149944,
      "latency": {
        "min": 0.371,
        "mean": 6.95,
        "max": 98.525,
        "pstdev": 13.111,
        "percentiles": {
          "p50": 3.057,
          "p75": 4.816,
          "p80": 5.416,
          "p90": 8.383,
          "p95": 50.27,
          "p99": 57.946,
          "p999": 69.038
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 182.71,
            "max": 182.71,
            "mean": 182.71
          },
          "cpu": {
            "min": 61.79,
            "max": 61.79,
            "mean": 61.79
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 122.12,
            "max": 122.12,
            "mean": 122.12
          },
          "cpu": {
            "min": 194.7,
            "max": 194.7,
            "mean": 194.7
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
      "throughput": 5224.35,
      "totalRequests": 156731,
      "latency": {
        "min": 0.376,
        "mean": 6.87,
        "max": 73.203,
        "pstdev": 12.827,
        "percentiles": {
          "p50": 3.06,
          "p75": 4.817,
          "p80": 5.428,
          "p90": 8.28,
          "p95": 50.618,
          "p99": 56.178,
          "p999": 61.28
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 209.85,
            "max": 209.85,
            "mean": 209.85
          },
          "cpu": {
            "min": 114.89,
            "max": 114.89,
            "mean": 114.89
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 110.5,
            "max": 110.5,
            "mean": 110.5
          },
          "cpu": {
            "min": 356.13,
            "max": 356.13,
            "mean": 356.13
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
      "throughput": 5247.75,
      "totalRequests": 157436,
      "latency": {
        "min": 0.383,
        "mean": 6.991,
        "max": 88.027,
        "pstdev": 12.998,
        "percentiles": {
          "p50": 3.091,
          "p75": 4.901,
          "p80": 5.532,
          "p90": 8.568,
          "p95": 50.655,
          "p99": 56.272,
          "p999": 62.16
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 208.95,
            "max": 208.95,
            "mean": 208.95
          },
          "cpu": {
            "min": 114.35,
            "max": 114.35,
            "mean": 114.35
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 122.19,
            "max": 122.19,
            "mean": 122.19
          },
          "cpu": {
            "min": 325.78,
            "max": 325.78,
            "mean": 325.78
          }
        }
      },
      "poolOverflow": 361,
      "upstreamConnections": 39
    },
    {
      "testName": "scale-down-httproutes-100",
      "routes": 100,
      "routesPerHostname": 1,
      "phase": "scaling-down",
      "throughput": 5201.69,
      "totalRequests": 156051,
      "latency": {
        "min": 0.38,
        "mean": 7.027,
        "max": 78.663,
        "pstdev": 12.995,
        "percentiles": {
          "p50": 3.129,
          "p75": 4.963,
          "p80": 5.6,
          "p90": 8.642,
          "p95": 50.726,
          "p99": 56.254,
          "p999": 61.509
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 211.34,
            "max": 211.34,
            "mean": 211.34
          },
          "cpu": {
            "min": 113.23,
            "max": 113.23,
            "mean": 113.23
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 122.17,
            "max": 122.17,
            "mean": 122.17
          },
          "cpu": {
            "min": 295.15,
            "max": 295.15,
            "mean": 295.15
          }
        }
      },
      "poolOverflow": 361,
      "upstreamConnections": 39
    },
    {
      "testName": "scale-down-httproutes-300",
      "routes": 300,
      "routesPerHostname": 1,
      "phase": "scaling-down",
      "throughput": 5206.95,
      "totalRequests": 156214,
      "latency": {
        "min": 0.386,
        "mean": 6.893,
        "max": 92.049,
        "pstdev": 12.874,
        "percentiles": {
          "p50": 3.088,
          "p75": 4.898,
          "p80": 5.548,
          "p90": 8.444,
          "p95": 50.268,
          "p99": 56.334,
          "p999": 65.038
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 158.73,
            "max": 158.73,
            "mean": 158.73
          },
          "cpu": {
            "min": 102.79,
            "max": 102.79,
            "mean": 102.79
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 122.15,
            "max": 122.15,
            "mean": 122.15
          },
          "cpu": {
            "min": 263,
            "max": 263,
            "mean": 263
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
      "throughput": 5150.23,
      "totalRequests": 154507,
      "latency": {
        "min": 0.375,
        "mean": 6.786,
        "max": 84.877,
        "pstdev": 12.812,
        "percentiles": {
          "p50": 3.009,
          "p75": 4.734,
          "p80": 5.34,
          "p90": 8.236,
          "p95": 50.247,
          "p99": 56.647,
          "p999": 65.042
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 168.04,
            "max": 168.04,
            "mean": 168.04
          },
          "cpu": {
            "min": 91.51,
            "max": 91.51,
            "mean": 91.51
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 122.14,
            "max": 122.14,
            "mean": 122.14
          },
          "cpu": {
            "min": 230.49,
            "max": 230.49,
            "mean": 230.49
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37
    }
  ]
};

export default benchmarkData;
