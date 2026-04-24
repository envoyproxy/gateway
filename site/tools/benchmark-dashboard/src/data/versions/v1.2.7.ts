import { TestSuite } from '../types';

// Benchmark data extracted from markdown report for version 1.2.7
// Generated on 2025-06-17T19:50:26.769Z

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.2.7",
    "runId": "1.2.7-1750189826769",
    "date": "2025-03-06",
    "environment": "GitHub CI",
    "description": "Benchmark report for Envoy Gateway 1.2.7",
    "downloadUrl": "https://github.com/envoyproxy/gateway/releases/download/v1.2.7/benchmark_report.zip",
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
      "throughput": 5984.83,
      "totalRequests": 179545,
      "latency": {
        "min": 0.402,
        "mean": 12.033,
        "max": 92.835,
        "pstdev": 16.725,
        "percentiles": {
          "p50": 5.655,
          "p75": 9.809,
          "p80": 11.804,
          "p90": 49.971,
          "p95": 55.76,
          "p99": 64.241,
          "p999": 73.961
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 112.35,
            "max": 112.35,
            "mean": 112.35
          },
          "cpu": {
            "min": 0.76,
            "max": 0.76,
            "mean": 0.76
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 28.29,
            "max": 28.29,
            "mean": 28.29
          },
          "cpu": {
            "min": 30.49,
            "max": 30.49,
            "mean": 30.49
          }
        }
      },
      "poolOverflow": 323,
      "upstreamConnections": 77
    },
    {
      "testName": "scale-up-httproutes-50",
      "routes": 50,
      "routesPerHostname": 1,
      "phase": "scaling-up",
      "throughput": 5564.85,
      "totalRequests": 166949,
      "latency": {
        "min": 0.362,
        "mean": 6.416,
        "max": 73.261,
        "pstdev": 12.268,
        "percentiles": {
          "p50": 2.855,
          "p75": 4.514,
          "p80": 5.094,
          "p90": 7.66,
          "p95": 49.031,
          "p99": 54.939,
          "p999": 60.436
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 125.94,
            "max": 125.94,
            "mean": 125.94
          },
          "cpu": {
            "min": 1.52,
            "max": 1.52,
            "mean": 1.52
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 34.46,
            "max": 34.46,
            "mean": 34.46
          },
          "cpu": {
            "min": 61.11,
            "max": 61.11,
            "mean": 61.11
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
      "throughput": 5618.23,
      "totalRequests": 168547,
      "latency": {
        "min": 0.372,
        "mean": 6.834,
        "max": 85.766,
        "pstdev": 12.643,
        "percentiles": {
          "p50": 3.107,
          "p75": 4.954,
          "p80": 5.56,
          "p90": 8.454,
          "p95": 49.485,
          "p99": 55.513,
          "p999": 61.272
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 132.05,
            "max": 132.05,
            "mean": 132.05
          },
          "cpu": {
            "min": 2.92,
            "max": 2.92,
            "mean": 2.92
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 40.63,
            "max": 40.63,
            "mean": 40.63
          },
          "cpu": {
            "min": 92.02,
            "max": 92.02,
            "mean": 92.02
          }
        }
      },
      "poolOverflow": 359,
      "upstreamConnections": 41
    },
    {
      "testName": "scale-up-httproutes-300",
      "routes": 300,
      "routesPerHostname": 1,
      "phase": "scaling-up",
      "throughput": 5506.26,
      "totalRequests": 165193,
      "latency": {
        "min": 0.383,
        "mean": 6.67,
        "max": 85.905,
        "pstdev": 12.677,
        "percentiles": {
          "p50": 2.946,
          "p75": 4.659,
          "p80": 5.243,
          "p90": 7.891,
          "p95": 49.846,
          "p99": 56.096,
          "p999": 65.185
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 146.09,
            "max": 146.09,
            "mean": 146.09
          },
          "cpu": {
            "min": 15.43,
            "max": 15.43,
            "mean": 15.43
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 62.86,
            "max": 62.86,
            "mean": 62.86
          },
          "cpu": {
            "min": 125.21,
            "max": 125.21,
            "mean": 125.21
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
      "throughput": 5580.92,
      "totalRequests": 167428,
      "latency": {
        "min": 0.31,
        "mean": 6.716,
        "max": 89.178,
        "pstdev": 12.233,
        "percentiles": {
          "p50": 3.157,
          "p75": 4.861,
          "p80": 5.446,
          "p90": 8.231,
          "p95": 47.419,
          "p99": 54.634,
          "p999": 65.587
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 163.49,
            "max": 163.49,
            "mean": 163.49
          },
          "cpu": {
            "min": 28.15,
            "max": 28.15,
            "mean": 28.15
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 83.07,
            "max": 83.07,
            "mean": 83.07
          },
          "cpu": {
            "min": 159.15,
            "max": 159.15,
            "mean": 159.15
          }
        }
      },
      "poolOverflow": 360,
      "upstreamConnections": 40
    },
    {
      "testName": "scale-up-httproutes-1000",
      "routes": 1000,
      "routesPerHostname": 1,
      "phase": "scaling-up",
      "throughput": 5298.3,
      "totalRequests": 158951,
      "latency": {
        "min": 0.369,
        "mean": 6.537,
        "max": 122.015,
        "pstdev": 12.654,
        "percentiles": {
          "p50": 2.914,
          "p75": 4.472,
          "p80": 5.012,
          "p90": 7.402,
          "p95": 48.639,
          "p99": 58.097,
          "p999": 69.103
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 187.81,
            "max": 187.81,
            "mean": 187.81
          },
          "cpu": {
            "min": 62.41,
            "max": 62.41,
            "mean": 62.41
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 133.27,
            "max": 133.27,
            "mean": 133.27
          },
          "cpu": {
            "min": 199.13,
            "max": 199.13,
            "mean": 199.13
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
      "throughput": 5623.57,
      "totalRequests": 168711,
      "latency": {
        "min": 0.38,
        "mean": 7.19,
        "max": 75.063,
        "pstdev": 13.103,
        "percentiles": {
          "p50": 3.211,
          "p75": 5.102,
          "p80": 5.772,
          "p90": 9.144,
          "p95": 50.722,
          "p99": 56.26,
          "p999": 61.515
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 132.38,
            "max": 132.38,
            "mean": 132.38
          },
          "cpu": {
            "min": 116.08,
            "max": 116.08,
            "mean": 116.08
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 128.69,
            "max": 128.69,
            "mean": 128.69
          },
          "cpu": {
            "min": 362.3,
            "max": 362.3,
            "mean": 362.3
          }
        }
      },
      "poolOverflow": 357,
      "upstreamConnections": 43
    },
    {
      "testName": "scale-down-httproutes-50",
      "routes": 50,
      "routesPerHostname": 1,
      "phase": "scaling-down",
      "throughput": 5516.59,
      "totalRequests": 165498,
      "latency": {
        "min": 0.329,
        "mean": 6.044,
        "max": 69.902,
        "pstdev": 11.632,
        "percentiles": {
          "p50": 2.893,
          "p75": 4.337,
          "p80": 4.79,
          "p90": 6.742,
          "p95": 47.155,
          "p99": 53.372,
          "p999": 58.073
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 144.02,
            "max": 144.02,
            "mean": 144.02
          },
          "cpu": {
            "min": 115.5,
            "max": 115.5,
            "mean": 115.5
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 129.14,
            "max": 129.14,
            "mean": 129.14
          },
          "cpu": {
            "min": 331.76,
            "max": 331.76,
            "mean": 331.76
          }
        }
      },
      "poolOverflow": 364,
      "upstreamConnections": 36
    },
    {
      "testName": "scale-down-httproutes-100",
      "routes": 100,
      "routesPerHostname": 1,
      "phase": "scaling-down",
      "throughput": 5515.82,
      "totalRequests": 165475,
      "latency": {
        "min": 0.38,
        "mean": 6.483,
        "max": 79.159,
        "pstdev": 12.363,
        "percentiles": {
          "p50": 2.872,
          "p75": 4.545,
          "p80": 5.131,
          "p90": 7.76,
          "p95": 49.19,
          "p99": 55.408,
          "p999": 61.89
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 135.79,
            "max": 135.79,
            "mean": 135.79
          },
          "cpu": {
            "min": 114.36,
            "max": 114.36,
            "mean": 114.36
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 109,
            "max": 109,
            "mean": 109
          },
          "cpu": {
            "min": 301.13,
            "max": 301.13,
            "mean": 301.13
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
      "throughput": 5403.53,
      "totalRequests": 162106,
      "latency": {
        "min": 0.354,
        "mean": 6.429,
        "max": 83.423,
        "pstdev": 12.305,
        "percentiles": {
          "p50": 2.864,
          "p75": 4.496,
          "p80": 5.072,
          "p90": 7.504,
          "p95": 48.885,
          "p99": 55.414,
          "p999": 63.565
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 155.42,
            "max": 155.42,
            "mean": 155.42
          },
          "cpu": {
            "min": 103.9,
            "max": 103.9,
            "mean": 103.9
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 133.59,
            "max": 133.59,
            "mean": 133.59
          },
          "cpu": {
            "min": 268.85,
            "max": 268.85,
            "mean": 268.85
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37
    },
    {
      "testName": "scale-down-httproutes-500",
      "routes": 500,
      "routesPerHostname": 1,
      "phase": "scaling-down",
      "throughput": 5405.24,
      "totalRequests": 162162,
      "latency": {
        "min": 0.387,
        "mean": 6.615,
        "max": 90.812,
        "pstdev": 12.66,
        "percentiles": {
          "p50": 2.913,
          "p75": 4.593,
          "p80": 5.174,
          "p90": 7.866,
          "p95": 49.461,
          "p99": 56.51,
          "p999": 67.248
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 165.58,
            "max": 165.58,
            "mean": 165.58
          },
          "cpu": {
            "min": 92.35,
            "max": 92.35,
            "mean": 92.35
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 133.42,
            "max": 133.42,
            "mean": 133.42
          },
          "cpu": {
            "min": 235.87,
            "max": 235.87,
            "mean": 235.87
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38
    }
  ]
};

export default benchmarkData;
