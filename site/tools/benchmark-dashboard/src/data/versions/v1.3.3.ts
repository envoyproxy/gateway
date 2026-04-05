import { TestSuite } from '../types';

// Benchmark data extracted from markdown report for version 1.3.3
// Generated on 2025-06-17T19:50:26.778Z

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.3.3",
    "runId": "1.3.3-1750189826778",
    "date": "2025-05-09",
    "environment": "GitHub CI",
    "description": "Benchmark report for Envoy Gateway 1.3.3",
    "downloadUrl": "https://github.com/envoyproxy/gateway/releases/download/v1.3.3/benchmark_report.zip",
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
      "throughput": 5302.05,
      "totalRequests": 159067,
      "latency": {
        "min": 0.384,
        "mean": 7.44,
        "max": 68.399,
        "pstdev": 12.191,
        "percentiles": {
          "p50": 3.692,
          "p75": 5.903,
          "p80": 6.67,
          "p90": 10.777,
          "p95": 47.519,
          "p99": 54.405,
          "p999": 59.346
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 126.17,
            "max": 126.17,
            "mean": 126.17
          },
          "cpu": {
            "min": 0.95,
            "max": 0.95,
            "mean": 0.95
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 25.93,
            "max": 25.93,
            "mean": 25.93
          },
          "cpu": {
            "min": 30.42,
            "max": 30.42,
            "mean": 30.42
          }
        }
      },
      "poolOverflow": 358,
      "upstreamConnections": 42
    },
    {
      "testName": "scale-up-httproutes-50",
      "routes": 50,
      "routesPerHostname": 1,
      "phase": "scaling-up",
      "throughput": 5270.25,
      "totalRequests": 158112,
      "latency": {
        "min": 0.392,
        "mean": 7.067,
        "max": 77.467,
        "pstdev": 12.552,
        "percentiles": {
          "p50": 3.33,
          "p75": 5.225,
          "p80": 5.887,
          "p90": 8.906,
          "p95": 49.113,
          "p99": 55.103,
          "p999": 60.295
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 145.89,
            "max": 145.89,
            "mean": 145.89
          },
          "cpu": {
            "min": 1.66,
            "max": 1.66,
            "mean": 1.66
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 32.1,
            "max": 32.1,
            "mean": 32.1
          },
          "cpu": {
            "min": 60.71,
            "max": 60.71,
            "mean": 60.71
          }
        }
      },
      "poolOverflow": 361,
      "upstreamConnections": 39
    },
    {
      "testName": "scale-up-httproutes-100",
      "routes": 100,
      "routesPerHostname": 1,
      "phase": "scaling-up",
      "throughput": 5307.61,
      "totalRequests": 159230,
      "latency": {
        "min": 0.387,
        "mean": 7.35,
        "max": 76.185,
        "pstdev": 12.732,
        "percentiles": {
          "p50": 3.508,
          "p75": 5.524,
          "p80": 6.22,
          "p90": 9.652,
          "p95": 49.1,
          "p99": 55.363,
          "p999": 62.654
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 153.05,
            "max": 153.05,
            "mean": 153.05
          },
          "cpu": {
            "min": 3.08,
            "max": 3.08,
            "mean": 3.08
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 38.29,
            "max": 38.29,
            "mean": 38.29
          },
          "cpu": {
            "min": 91.5,
            "max": 91.5,
            "mean": 91.5
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
      "throughput": 5139.63,
      "totalRequests": 154189,
      "latency": {
        "min": 0.39,
        "mean": 7.376,
        "max": 78.098,
        "pstdev": 12.993,
        "percentiles": {
          "p50": 3.429,
          "p75": 5.431,
          "p80": 6.116,
          "p90": 9.633,
          "p95": 49.809,
          "p99": 56.043,
          "p999": 64.651
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 152.28,
            "max": 152.28,
            "mean": 152.28
          },
          "cpu": {
            "min": 15.6,
            "max": 15.6,
            "mean": 15.6
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 56.33,
            "max": 56.33,
            "mean": 56.33
          },
          "cpu": {
            "min": 124.15,
            "max": 124.15,
            "mean": 124.15
          }
        }
      },
      "poolOverflow": 360,
      "upstreamConnections": 40
    },
    {
      "testName": "scale-up-httproutes-500",
      "routes": 500,
      "routesPerHostname": 1,
      "phase": "scaling-up",
      "throughput": 5169.8,
      "totalRequests": 155094,
      "latency": {
        "min": 0.374,
        "mean": 7.385,
        "max": 105.918,
        "pstdev": 13.085,
        "percentiles": {
          "p50": 3.382,
          "p75": 5.336,
          "p80": 6.024,
          "p90": 9.658,
          "p95": 50.104,
          "p99": 56.487,
          "p999": 66.418
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 166.29,
            "max": 166.29,
            "mean": 166.29
          },
          "cpu": {
            "min": 28.13,
            "max": 28.13,
            "mean": 28.13
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 76.53,
            "max": 76.53,
            "mean": 76.53
          },
          "cpu": {
            "min": 157.45,
            "max": 157.45,
            "mean": 157.45
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
      "throughput": 4963.46,
      "totalRequests": 148908,
      "latency": {
        "min": 0.374,
        "mean": 7.335,
        "max": 94.58,
        "pstdev": 13.294,
        "percentiles": {
          "p50": 3.255,
          "p75": 5.279,
          "p80": 5.998,
          "p90": 9.532,
          "p95": 50.294,
          "p99": 58.908,
          "p999": 70.729
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 199.79,
            "max": 199.79,
            "mean": 199.79
          },
          "cpu": {
            "min": 61.67,
            "max": 61.67,
            "mean": 61.67
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 120.77,
            "max": 120.77,
            "mean": 120.77
          },
          "cpu": {
            "min": 195.1,
            "max": 195.1,
            "mean": 195.1
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
      "throughput": 5200.07,
      "totalRequests": 156001,
      "latency": {
        "min": 0.395,
        "mean": 7.53,
        "max": 81.936,
        "pstdev": 13.142,
        "percentiles": {
          "p50": 3.438,
          "p75": 5.497,
          "p80": 6.226,
          "p90": 10.218,
          "p95": 50.345,
          "p99": 56.436,
          "p999": 61.919
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 221.87,
            "max": 221.87,
            "mean": 221.87
          },
          "cpu": {
            "min": 114.77,
            "max": 114.77,
            "mean": 114.77
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 121.27,
            "max": 121.27,
            "mean": 121.27
          },
          "cpu": {
            "min": 356.4,
            "max": 356.4,
            "mean": 356.4
          }
        }
      },
      "poolOverflow": 359,
      "upstreamConnections": 41
    },
    {
      "testName": "scale-down-httproutes-50",
      "routes": 50,
      "routesPerHostname": 1,
      "phase": "scaling-down",
      "throughput": 5229.66,
      "totalRequests": 156893,
      "latency": {
        "min": 0.368,
        "mean": 6.911,
        "max": 72.273,
        "pstdev": 12.396,
        "percentiles": {
          "p50": 3.227,
          "p75": 5.075,
          "p80": 5.695,
          "p90": 8.751,
          "p95": 48.676,
          "p99": 54.822,
          "p999": 59.865
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 221.8,
            "max": 221.8,
            "mean": 221.8
          },
          "cpu": {
            "min": 114.13,
            "max": 114.13,
            "mean": 114.13
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 121.22,
            "max": 121.22,
            "mean": 121.22
          },
          "cpu": {
            "min": 325.9,
            "max": 325.9,
            "mean": 325.9
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
      "throughput": 5099.34,
      "totalRequests": 152980,
      "latency": {
        "min": 0.357,
        "mean": 6.637,
        "max": 68.1,
        "pstdev": 12.012,
        "percentiles": {
          "p50": 3.167,
          "p75": 4.972,
          "p80": 5.58,
          "p90": 8.302,
          "p95": 47.65,
          "p99": 53.993,
          "p999": 59.299
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 182.69,
            "max": 182.69,
            "mean": 182.69
          },
          "cpu": {
            "min": 113.02,
            "max": 113.02,
            "mean": 113.02
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 121.2,
            "max": 121.2,
            "mean": 121.2
          },
          "cpu": {
            "min": 295.3,
            "max": 295.3,
            "mean": 295.3
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
      "throughput": 5145.13,
      "totalRequests": 154357,
      "latency": {
        "min": 0.379,
        "mean": 7.234,
        "max": 111.517,
        "pstdev": 12.814,
        "percentiles": {
          "p50": 3.381,
          "p75": 5.331,
          "p80": 5.993,
          "p90": 9.269,
          "p95": 49.465,
          "p99": 55.562,
          "p999": 64.862
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 159.09,
            "max": 159.09,
            "mean": 159.09
          },
          "cpu": {
            "min": 102.57,
            "max": 102.57,
            "mean": 102.57
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 121.08,
            "max": 121.08,
            "mean": 121.08
          },
          "cpu": {
            "min": 263.29,
            "max": 263.29,
            "mean": 263.29
          }
        }
      },
      "poolOverflow": 361,
      "upstreamConnections": 39
    },
    {
      "testName": "scale-down-httproutes-500",
      "routes": 500,
      "routesPerHostname": 1,
      "phase": "scaling-down",
      "throughput": 5130.95,
      "totalRequests": 153929,
      "latency": {
        "min": 0.365,
        "mean": 7.098,
        "max": 95.834,
        "pstdev": 12.843,
        "percentiles": {
          "p50": 3.267,
          "p75": 5.11,
          "p80": 5.752,
          "p90": 8.975,
          "p95": 49.702,
          "p99": 56.408,
          "p999": 67.33
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 163.91,
            "max": 163.91,
            "mean": 163.91
          },
          "cpu": {
            "min": 91.18,
            "max": 91.18,
            "mean": 91.18
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 120.91,
            "max": 120.91,
            "mean": 120.91
          },
          "cpu": {
            "min": 230.77,
            "max": 230.77,
            "mean": 230.77
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38
    }
  ]
};

export default benchmarkData;
