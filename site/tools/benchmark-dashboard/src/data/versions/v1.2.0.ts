import { TestSuite } from '../types';

// Benchmark data extracted from markdown report for version 1.2.0
// Generated on 2025-06-17T19:50:26.754Z

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.2.0",
    "runId": "1.2.0-1750189826754",
    "date": "2024-11-06",
    "environment": "GitHub CI",
    "description": "Benchmark report for Envoy Gateway 1.2.0",
    "downloadUrl": "https://github.com/envoyproxy/gateway/releases/download/v1.2.0/benchmark_report.zip",
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
      "throughput": 6006.27,
      "totalRequests": 180191,
      "latency": {
        "min": 0.359,
        "mean": 6.202,
        "max": 74.358,
        "pstdev": 11.262,
        "percentiles": {
          "p50": 2.976,
          "p75": 4.719,
          "p80": 5.33,
          "p90": 8.159,
          "p95": 44.003,
          "p99": 54.19,
          "p999": 59.357
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 113.66,
            "max": 113.66,
            "mean": 113.66
          },
          "cpu": {
            "min": 0.75,
            "max": 0.75,
            "mean": 0.75
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 25.55,
            "max": 25.55,
            "mean": 25.55
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
      "throughput": 5988.15,
      "totalRequests": 179650,
      "latency": {
        "min": 0.366,
        "mean": 6.093,
        "max": 82.386,
        "pstdev": 11.67,
        "percentiles": {
          "p50": 2.838,
          "p75": 4.283,
          "p80": 4.798,
          "p90": 6.972,
          "p95": 47.523,
          "p99": 54.038,
          "p999": 59.813
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 119.16,
            "max": 119.16,
            "mean": 119.16
          },
          "cpu": {
            "min": 1.53,
            "max": 1.53,
            "mean": 1.53
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 31.71,
            "max": 31.71,
            "mean": 31.71
          },
          "cpu": {
            "min": 61.03,
            "max": 61.03,
            "mean": 61.03
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
      "throughput": 5930.13,
      "totalRequests": 177905,
      "latency": {
        "min": 0.374,
        "mean": 6.182,
        "max": 93.884,
        "pstdev": 11.844,
        "percentiles": {
          "p50": 2.812,
          "p75": 4.42,
          "p80": 4.967,
          "p90": 7.339,
          "p95": 47.618,
          "p99": 54.661,
          "p999": 61.865
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 123.35,
            "max": 123.35,
            "mean": 123.35
          },
          "cpu": {
            "min": 2.89,
            "max": 2.89,
            "mean": 2.89
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 37.86,
            "max": 37.86,
            "mean": 37.86
          },
          "cpu": {
            "min": 91.87,
            "max": 91.87,
            "mean": 91.87
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
      "throughput": 5713.27,
      "totalRequests": 171401,
      "latency": {
        "min": 0.384,
        "mean": 6.405,
        "max": 105.545,
        "pstdev": 12.313,
        "percentiles": {
          "p50": 2.832,
          "p75": 4.45,
          "p80": 5.057,
          "p90": 7.773,
          "p95": 48.732,
          "p99": 56.084,
          "p999": 66.22
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 150.92,
            "max": 150.92,
            "mean": 150.92
          },
          "cpu": {
            "min": 15.08,
            "max": 15.08,
            "mean": 15.08
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 60.05,
            "max": 60.05,
            "mean": 60.05
          },
          "cpu": {
            "min": 125.09,
            "max": 125.09,
            "mean": 125.09
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
      "throughput": 5689.03,
      "totalRequests": 170675,
      "latency": {
        "min": 0.365,
        "mean": 5.939,
        "max": 98.639,
        "pstdev": 11.585,
        "percentiles": {
          "p50": 2.715,
          "p75": 4.258,
          "p80": 4.772,
          "p90": 7.055,
          "p95": 45.996,
          "p99": 54.368,
          "p999": 68.464
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 156.97,
            "max": 156.97,
            "mean": 156.97
          },
          "cpu": {
            "min": 27.62,
            "max": 27.62,
            "mean": 27.62
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 82.26,
            "max": 82.26,
            "mean": 82.26
          },
          "cpu": {
            "min": 158.86,
            "max": 158.86,
            "mean": 158.86
          }
        }
      },
      "poolOverflow": 365,
      "upstreamConnections": 35
    },
    {
      "testName": "scale-up-httproutes-1000",
      "routes": 1000,
      "routesPerHostname": 1,
      "phase": "scaling-up",
      "throughput": 5407.08,
      "totalRequests": 162220,
      "latency": {
        "min": 0.371,
        "mean": 6.424,
        "max": 131.579,
        "pstdev": 12.473,
        "percentiles": {
          "p50": 2.692,
          "p75": 4.503,
          "p80": 5.177,
          "p90": 8.264,
          "p95": 47.54,
          "p99": 58.488,
          "p999": 72.72
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 215.93,
            "max": 215.93,
            "mean": 215.93
          },
          "cpu": {
            "min": 61.56,
            "max": 61.56,
            "mean": 61.56
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 130.5,
            "max": 130.5,
            "mean": 130.5
          },
          "cpu": {
            "min": 197.45,
            "max": 197.45,
            "mean": 197.45
          }
        }
      },
      "poolOverflow": 364,
      "upstreamConnections": 36
    },
    {
      "testName": "scale-down-httproutes-10",
      "routes": 10,
      "routesPerHostname": 1,
      "phase": "scaling-down",
      "throughput": 5905.23,
      "totalRequests": 177157,
      "latency": {
        "min": 0.396,
        "mean": 6.205,
        "max": 92.979,
        "pstdev": 12.065,
        "percentiles": {
          "p50": 2.713,
          "p75": 4.262,
          "p80": 4.832,
          "p90": 7.415,
          "p95": 48.306,
          "p99": 55.736,
          "p999": 63.318
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 233.71,
            "max": 233.71,
            "mean": 233.71
          },
          "cpu": {
            "min": 114.53,
            "max": 114.53,
            "mean": 114.53
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 118.94,
            "max": 118.94,
            "mean": 118.94
          },
          "cpu": {
            "min": 360.39,
            "max": 360.39,
            "mean": 360.39
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
      "throughput": 5793.19,
      "totalRequests": 173800,
      "latency": {
        "min": 0.385,
        "mean": 6.501,
        "max": 86.843,
        "pstdev": 12.452,
        "percentiles": {
          "p50": 2.744,
          "p75": 4.568,
          "p80": 5.227,
          "p90": 8.494,
          "p95": 48.965,
          "p99": 56.467,
          "p999": 65.382
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 202.08,
            "max": 202.08,
            "mean": 202.08
          },
          "cpu": {
            "min": 113.96,
            "max": 113.96,
            "mean": 113.96
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 121.24,
            "max": 121.24,
            "mean": 121.24
          },
          "cpu": {
            "min": 329.91,
            "max": 329.91,
            "mean": 329.91
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
      "throughput": 5783.8,
      "totalRequests": 173514,
      "latency": {
        "min": 0.385,
        "mean": 6.179,
        "max": 85.405,
        "pstdev": 11.889,
        "percentiles": {
          "p50": 2.707,
          "p75": 4.462,
          "p80": 5.071,
          "p90": 7.76,
          "p95": 47.423,
          "p99": 54.994,
          "p999": 64.567
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 144.59,
            "max": 144.59,
            "mean": 144.59
          },
          "cpu": {
            "min": 112.93,
            "max": 112.93,
            "mean": 112.93
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 121.22,
            "max": 121.22,
            "mean": 121.22
          },
          "cpu": {
            "min": 299.28,
            "max": 299.28,
            "mean": 299.28
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37
    },
    {
      "testName": "scale-down-httproutes-300",
      "routes": 300,
      "routesPerHostname": 1,
      "phase": "scaling-down",
      "throughput": 5804.51,
      "totalRequests": 174139,
      "latency": {
        "min": 0.384,
        "mean": 6.147,
        "max": 82.661,
        "pstdev": 11.892,
        "percentiles": {
          "p50": 2.807,
          "p75": 4.284,
          "p80": 4.791,
          "p90": 7.039,
          "p95": 47.558,
          "p99": 55.162,
          "p999": 66.723
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 160.23,
            "max": 160.23,
            "mean": 160.23
          },
          "cpu": {
            "min": 102.85,
            "max": 102.85,
            "mean": 102.85
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 130.62,
            "max": 130.62,
            "mean": 130.62
          },
          "cpu": {
            "min": 267.09,
            "max": 267.09,
            "mean": 267.09
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
      "throughput": 5804.87,
      "totalRequests": 174149,
      "latency": {
        "min": 0.368,
        "mean": 5.828,
        "max": 106.119,
        "pstdev": 11.561,
        "percentiles": {
          "p50": 2.647,
          "p75": 4.078,
          "p80": 4.553,
          "p90": 6.656,
          "p95": 46.047,
          "p99": 55.025,
          "p999": 66.633
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 155.88,
            "max": 155.88,
            "mean": 155.88
          },
          "cpu": {
            "min": 91.5,
            "max": 91.5,
            "mean": 91.5
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 130.61,
            "max": 130.61,
            "mean": 130.61
          },
          "cpu": {
            "min": 234.47,
            "max": 234.47,
            "mean": 234.47
          }
        }
      },
      "poolOverflow": 365,
      "upstreamConnections": 35
    }
  ]
};

export default benchmarkData;
