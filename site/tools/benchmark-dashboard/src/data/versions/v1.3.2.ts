import { TestSuite } from '../types';

// Benchmark data extracted from markdown report for version 1.3.2
// Generated on 2025-06-17T19:50:26.777Z

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.3.2",
    "runId": "1.3.2-1750189826777",
    "date": "2025-03-24",
    "environment": "GitHub CI",
    "description": "Benchmark report for Envoy Gateway 1.3.2",
    "downloadUrl": "https://github.com/envoyproxy/gateway/releases/download/v1.3.2/benchmark_report.zip",
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
      "throughput": 5203.06,
      "totalRequests": 156092,
      "latency": {
        "min": 0.371,
        "mean": 6.97,
        "max": 73.949,
        "pstdev": 12.106,
        "percentiles": {
          "p50": 3.331,
          "p75": 5.396,
          "p80": 6.08,
          "p90": 9.715,
          "p95": 47.564,
          "p99": 55.666,
          "p999": 60.872
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 126.68,
            "max": 126.68,
            "mean": 126.68
          },
          "cpu": {
            "min": 0.99,
            "max": 0.99,
            "mean": 0.99
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 28.17,
            "max": 28.17,
            "mean": 28.17
          },
          "cpu": {
            "min": 30.49,
            "max": 30.49,
            "mean": 30.49
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
      "throughput": 5157.35,
      "totalRequests": 154724,
      "latency": {
        "min": 0.38,
        "mean": 6.752,
        "max": 83.992,
        "pstdev": 12.526,
        "percentiles": {
          "p50": 3.089,
          "p75": 4.875,
          "p80": 5.484,
          "p90": 8.253,
          "p95": 49.682,
          "p99": 55.398,
          "p999": 61.507
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 129.05,
            "max": 129.05,
            "mean": 129.05
          },
          "cpu": {
            "min": 1.79,
            "max": 1.79,
            "mean": 1.79
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 32.33,
            "max": 32.33,
            "mean": 32.33
          },
          "cpu": {
            "min": 61.02,
            "max": 61.02,
            "mean": 61.02
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
      "throughput": 5052.19,
      "totalRequests": 151566,
      "latency": {
        "min": 0.363,
        "mean": 7.564,
        "max": 72.122,
        "pstdev": 13.492,
        "percentiles": {
          "p50": 3.384,
          "p75": 5.472,
          "p80": 6.191,
          "p90": 9.941,
          "p95": 51.32,
          "p99": 57.128,
          "p999": 63.107
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 146.92,
            "max": 146.92,
            "mean": 146.92
          },
          "cpu": {
            "min": 3.17,
            "max": 3.17,
            "mean": 3.17
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 40.52,
            "max": 40.52,
            "mean": 40.52
          },
          "cpu": {
            "min": 91.79,
            "max": 91.79,
            "mean": 91.79
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
      "throughput": 5101.42,
      "totalRequests": 153051,
      "latency": {
        "min": 0.397,
        "mean": 6.832,
        "max": 82.624,
        "pstdev": 12.894,
        "percentiles": {
          "p50": 2.99,
          "p75": 4.798,
          "p80": 5.418,
          "p90": 8.366,
          "p95": 50.391,
          "p99": 56.737,
          "p999": 65.114
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 154.98,
            "max": 154.98,
            "mean": 154.98
          },
          "cpu": {
            "min": 15.2,
            "max": 15.2,
            "mean": 15.2
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 58.7,
            "max": 58.7,
            "mean": 58.7
          },
          "cpu": {
            "min": 124.54,
            "max": 124.54,
            "mean": 124.54
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
      "throughput": 4948.44,
      "totalRequests": 148459,
      "latency": {
        "min": 0.398,
        "mean": 7.2,
        "max": 94.949,
        "pstdev": 13.256,
        "percentiles": {
          "p50": 3.208,
          "p75": 5.057,
          "p80": 5.652,
          "p90": 8.78,
          "p95": 50.821,
          "p99": 57.442,
          "p999": 66.746
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 163.01,
            "max": 163.01,
            "mean": 163.01
          },
          "cpu": {
            "min": 27.68,
            "max": 27.68,
            "mean": 27.68
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 78.9,
            "max": 78.9,
            "mean": 78.9
          },
          "cpu": {
            "min": 158.2,
            "max": 158.2,
            "mean": 158.2
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
      "throughput": 4933.76,
      "totalRequests": 148015,
      "latency": {
        "min": 0.361,
        "mean": 7.088,
        "max": 127.959,
        "pstdev": 13.349,
        "percentiles": {
          "p50": 3.052,
          "p75": 4.885,
          "p80": 5.529,
          "p90": 8.785,
          "p95": 50.7,
          "p99": 59.123,
          "p999": 70.66
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 196.05,
            "max": 196.05,
            "mean": 196.05
          },
          "cpu": {
            "min": 62.32,
            "max": 62.32,
            "mean": 62.32
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 127.19,
            "max": 127.19,
            "mean": 127.19
          },
          "cpu": {
            "min": 196.79,
            "max": 196.79,
            "mean": 196.79
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
      "throughput": 5116.93,
      "totalRequests": 153515,
      "latency": {
        "min": 0.387,
        "mean": 9.017,
        "max": 78.245,
        "pstdev": 15.046,
        "percentiles": {
          "p50": 3.922,
          "p75": 6.583,
          "p80": 7.609,
          "p90": 16.301,
          "p95": 54.007,
          "p99": 59.99,
          "p999": 67.956
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 210.06,
            "max": 210.06,
            "mean": 210.06
          },
          "cpu": {
            "min": 115.56,
            "max": 115.56,
            "mean": 115.56
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 110.31,
            "max": 110.31,
            "mean": 110.31
          },
          "cpu": {
            "min": 358.89,
            "max": 358.89,
            "mean": 358.89
          }
        }
      },
      "poolOverflow": 351,
      "upstreamConnections": 49
    },
    {
      "testName": "scale-down-httproutes-50",
      "routes": 50,
      "routesPerHostname": 1,
      "phase": "scaling-down",
      "throughput": 4991.13,
      "totalRequests": 149734,
      "latency": {
        "min": 0.391,
        "mean": 7.071,
        "max": 80.433,
        "pstdev": 12.841,
        "percentiles": {
          "p50": 3.199,
          "p75": 5.005,
          "p80": 5.635,
          "p90": 8.73,
          "p95": 49.813,
          "p99": 55.756,
          "p999": 61.407
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 199.41,
            "max": 199.41,
            "mean": 199.41
          },
          "cpu": {
            "min": 114.89,
            "max": 114.89,
            "mean": 114.89
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 127.44,
            "max": 127.44,
            "mean": 127.44
          },
          "cpu": {
            "min": 328.33,
            "max": 328.33,
            "mean": 328.33
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
      "throughput": 5138.59,
      "totalRequests": 154161,
      "latency": {
        "min": 0.377,
        "mean": 6.954,
        "max": 86.982,
        "pstdev": 13.049,
        "percentiles": {
          "p50": 3.045,
          "p75": 4.886,
          "p80": 5.525,
          "p90": 8.466,
          "p95": 51.048,
          "p99": 56.733,
          "p999": 62.337
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 219.76,
            "max": 219.76,
            "mean": 219.76
          },
          "cpu": {
            "min": 113.71,
            "max": 113.71,
            "mean": 113.71
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 127.31,
            "max": 127.31,
            "mean": 127.31
          },
          "cpu": {
            "min": 297.73,
            "max": 297.73,
            "mean": 297.73
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
      "throughput": 5051.68,
      "totalRequests": 151551,
      "latency": {
        "min": 0.389,
        "mean": 7.977,
        "max": 86.577,
        "pstdev": 14.09,
        "percentiles": {
          "p50": 3.498,
          "p75": 5.731,
          "p80": 6.471,
          "p90": 10.947,
          "p95": 52.355,
          "p99": 58.748,
          "p999": 67.379
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 259.14,
            "max": 259.14,
            "mean": 259.14
          },
          "cpu": {
            "min": 103.22,
            "max": 103.22,
            "mean": 103.22
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 127.31,
            "max": 127.31,
            "mean": 127.31
          },
          "cpu": {
            "min": 265.59,
            "max": 265.59,
            "mean": 265.59
          }
        }
      },
      "poolOverflow": 357,
      "upstreamConnections": 43
    },
    {
      "testName": "scale-down-httproutes-500",
      "routes": 500,
      "routesPerHostname": 1,
      "phase": "scaling-down",
      "throughput": 4962.4,
      "totalRequests": 148872,
      "latency": {
        "min": 0.384,
        "mean": 7.209,
        "max": 110.186,
        "pstdev": 13.357,
        "percentiles": {
          "p50": 3.112,
          "p75": 5.018,
          "p80": 5.694,
          "p90": 9.338,
          "p95": 51.267,
          "p99": 57.384,
          "p999": 67.244
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 177.76,
            "max": 177.76,
            "mean": 177.76
          },
          "cpu": {
            "min": 91.86,
            "max": 91.86,
            "mean": 91.86
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 127.24,
            "max": 127.24,
            "mean": 127.24
          },
          "cpu": {
            "min": 233.07,
            "max": 233.07,
            "mean": 233.07
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38
    }
  ]
};

export default benchmarkData;
