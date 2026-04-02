import { TestSuite } from '../types';

// Benchmark data extracted from markdown report for version 1.2.5
// Generated on 2025-06-17T19:50:26.766Z

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.2.5",
    "runId": "1.2.5-1750189826766",
    "date": "2025-01-14",
    "environment": "GitHub CI",
    "description": "Benchmark report for Envoy Gateway 1.2.5",
    "downloadUrl": "https://github.com/envoyproxy/gateway/releases/download/v1.2.5/benchmark_report.zip",
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
      "throughput": 5704.46,
      "totalRequests": 171136,
      "latency": {
        "min": 0.37,
        "mean": 6.52,
        "max": 74.477,
        "pstdev": 11.781,
        "percentiles": {
          "p50": 3.051,
          "p75": 4.913,
          "p80": 5.565,
          "p90": 8.7,
          "p95": 47.091,
          "p99": 54.638,
          "p999": 59.404
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 112.07,
            "max": 112.07,
            "mean": 112.07
          },
          "cpu": {
            "min": 0.76,
            "max": 0.76,
            "mean": 0.76
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 24.38,
            "max": 24.38,
            "mean": 24.38
          },
          "cpu": {
            "min": 30.47,
            "max": 30.47,
            "mean": 30.47
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
      "throughput": 5647.86,
      "totalRequests": 169439,
      "latency": {
        "min": 0.379,
        "mean": 6.313,
        "max": 80.498,
        "pstdev": 11.935,
        "percentiles": {
          "p50": 2.854,
          "p75": 4.55,
          "p80": 5.158,
          "p90": 7.82,
          "p95": 48.025,
          "p99": 54.39,
          "p999": 60.321
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 123.4,
            "max": 123.4,
            "mean": 123.4
          },
          "cpu": {
            "min": 1.48,
            "max": 1.48,
            "mean": 1.48
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 30.55,
            "max": 30.55,
            "mean": 30.55
          },
          "cpu": {
            "min": 61.09,
            "max": 61.09,
            "mean": 61.09
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
      "throughput": 5529.29,
      "totalRequests": 165879,
      "latency": {
        "min": 0.379,
        "mean": 6.24,
        "max": 73.539,
        "pstdev": 12.017,
        "percentiles": {
          "p50": 2.818,
          "p75": 4.51,
          "p80": 5.068,
          "p90": 7.413,
          "p95": 48.3,
          "p99": 54.171,
          "p999": 59.308
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 139.7,
            "max": 139.7,
            "mean": 139.7
          },
          "cpu": {
            "min": 2.92,
            "max": 2.92,
            "mean": 2.92
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 36.73,
            "max": 36.73,
            "mean": 36.73
          },
          "cpu": {
            "min": 91.9,
            "max": 91.9,
            "mean": 91.9
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
      "throughput": 5542.3,
      "totalRequests": 166274,
      "latency": {
        "min": 0.365,
        "mean": 6.463,
        "max": 81.305,
        "pstdev": 12.357,
        "percentiles": {
          "p50": 2.89,
          "p75": 4.545,
          "p80": 5.116,
          "p90": 7.585,
          "p95": 49.043,
          "p99": 55.365,
          "p999": 64.399
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 146.11,
            "max": 146.11,
            "mean": 146.11
          },
          "cpu": {
            "min": 15.2,
            "max": 15.2,
            "mean": 15.2
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 58.91,
            "max": 58.91,
            "mean": 58.91
          },
          "cpu": {
            "min": 124.96,
            "max": 124.96,
            "mean": 124.96
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
      "throughput": 5407.59,
      "totalRequests": 162228,
      "latency": {
        "min": 0.363,
        "mean": 6.396,
        "max": 88.236,
        "pstdev": 12.419,
        "percentiles": {
          "p50": 2.816,
          "p75": 4.45,
          "p80": 5.012,
          "p90": 7.485,
          "p95": 48.762,
          "p99": 55.666,
          "p999": 67.457
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 154.01,
            "max": 154.01,
            "mean": 154.01
          },
          "cpu": {
            "min": 27.98,
            "max": 27.98,
            "mean": 27.98
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 81.11,
            "max": 81.11,
            "mean": 81.11
          },
          "cpu": {
            "min": 158.85,
            "max": 158.85,
            "mean": 158.85
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37
    },
    {
      "testName": "scale-up-httproutes-1000",
      "routes": 1000,
      "routesPerHostname": 1,
      "phase": "scaling-up",
      "throughput": 5247.32,
      "totalRequests": 157425,
      "latency": {
        "min": 0.374,
        "mean": 6.604,
        "max": 106.491,
        "pstdev": 12.903,
        "percentiles": {
          "p50": 2.88,
          "p75": 4.514,
          "p80": 5.089,
          "p90": 7.639,
          "p95": 49.453,
          "p99": 58.75,
          "p999": 71.97
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 183.52,
            "max": 183.52,
            "mean": 183.52
          },
          "cpu": {
            "min": 61.85,
            "max": 61.85,
            "mean": 61.85
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 129.33,
            "max": 129.33,
            "mean": 129.33
          },
          "cpu": {
            "min": 198.98,
            "max": 198.98,
            "mean": 198.98
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
      "throughput": 5668.05,
      "totalRequests": 170048,
      "latency": {
        "min": 0.366,
        "mean": 6.143,
        "max": 73.551,
        "pstdev": 12.054,
        "percentiles": {
          "p50": 2.726,
          "p75": 4.279,
          "p80": 4.799,
          "p90": 7.123,
          "p95": 48.988,
          "p99": 54.704,
          "p999": 58.812
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 138.57,
            "max": 138.57,
            "mean": 138.57
          },
          "cpu": {
            "min": 115.2,
            "max": 115.2,
            "mean": 115.2
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 121.88,
            "max": 121.88,
            "mean": 121.88
          },
          "cpu": {
            "min": 361.89,
            "max": 361.89,
            "mean": 361.89
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
      "throughput": 5579.16,
      "totalRequests": 167375,
      "latency": {
        "min": 0.353,
        "mean": 6.19,
        "max": 87.244,
        "pstdev": 12.028,
        "percentiles": {
          "p50": 2.803,
          "p75": 4.357,
          "p80": 4.889,
          "p90": 7.11,
          "p95": 48.482,
          "p99": 54.329,
          "p999": 59.113
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 132.83,
            "max": 132.83,
            "mean": 132.83
          },
          "cpu": {
            "min": 114.63,
            "max": 114.63,
            "mean": 114.63
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 122.16,
            "max": 122.16,
            "mean": 122.16
          },
          "cpu": {
            "min": 331.47,
            "max": 331.47,
            "mean": 331.47
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
      "throughput": 5581.7,
      "totalRequests": 167453,
      "latency": {
        "min": 0.377,
        "mean": 6.525,
        "max": 72.724,
        "pstdev": 12.286,
        "percentiles": {
          "p50": 2.929,
          "p75": 4.641,
          "p80": 5.219,
          "p90": 7.915,
          "p95": 48.732,
          "p99": 54.81,
          "p999": 61.45
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 133.91,
            "max": 133.91,
            "mean": 133.91
          },
          "cpu": {
            "min": 113.44,
            "max": 113.44,
            "mean": 113.44
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 122.02,
            "max": 122.02,
            "mean": 122.02
          },
          "cpu": {
            "min": 300.79,
            "max": 300.79,
            "mean": 300.79
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
      "throughput": 5447.53,
      "totalRequests": 163426,
      "latency": {
        "min": 0.368,
        "mean": 6.54,
        "max": 88.612,
        "pstdev": 12.451,
        "percentiles": {
          "p50": 2.904,
          "p75": 4.546,
          "p80": 5.128,
          "p90": 7.692,
          "p95": 49.305,
          "p99": 55.422,
          "p999": 64.116
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 145.28,
            "max": 145.28,
            "mean": 145.28
          },
          "cpu": {
            "min": 102.98,
            "max": 102.98,
            "mean": 102.98
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 129.93,
            "max": 129.93,
            "mean": 129.93
          },
          "cpu": {
            "min": 268.5,
            "max": 268.5,
            "mean": 268.5
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
      "throughput": 5437.61,
      "totalRequests": 163132,
      "latency": {
        "min": 0.383,
        "mean": 6.527,
        "max": 102.998,
        "pstdev": 12.6,
        "percentiles": {
          "p50": 2.827,
          "p75": 4.47,
          "p80": 5.07,
          "p90": 7.714,
          "p95": 49.029,
          "p99": 56.436,
          "p999": 66.684
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 164.07,
            "max": 164.07,
            "mean": 164.07
          },
          "cpu": {
            "min": 91.64,
            "max": 91.64,
            "mean": 91.64
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 129.93,
            "max": 129.93,
            "mean": 129.93
          },
          "cpu": {
            "min": 235.61,
            "max": 235.61,
            "mean": 235.61
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38
    }
  ]
};

export default benchmarkData;
