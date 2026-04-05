import { TestSuite } from '../types';

// Benchmark data extracted from markdown report for version 1.2.1
// Generated on 2025-06-17T19:50:26.759Z

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.2.1",
    "runId": "1.2.1-1750189826759",
    "date": "2024-11-07",
    "environment": "GitHub CI",
    "description": "Benchmark report for Envoy Gateway 1.2.1",
    "downloadUrl": "https://github.com/envoyproxy/gateway/releases/download/v1.2.1/benchmark_report.zip",
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
      "throughput": 6110.53,
      "totalRequests": 183316,
      "latency": {
        "min": 0.337,
        "mean": 6.415,
        "max": 92.647,
        "pstdev": 11.35,
        "percentiles": {
          "p50": 3.107,
          "p75": 5.037,
          "p80": 5.711,
          "p90": 8.782,
          "p95": 45.103,
          "p99": 53.942,
          "p999": 60.79
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 111.18,
            "max": 111.18,
            "mean": 111.18
          },
          "cpu": {
            "min": 0.79,
            "max": 0.79,
            "mean": 0.79
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 24.25,
            "max": 24.25,
            "mean": 24.25
          },
          "cpu": {
            "min": 30.62,
            "max": 30.62,
            "mean": 30.62
          }
        }
      },
      "poolOverflow": 359,
      "upstreamConnections": 41
    },
    {
      "testName": "scale-up-httproutes-50",
      "routes": 50,
      "routesPerHostname": 1,
      "phase": "scaling-up",
      "throughput": 6011.43,
      "totalRequests": 180343,
      "latency": {
        "min": 0.377,
        "mean": 6.244,
        "max": 72.724,
        "pstdev": 11.87,
        "percentiles": {
          "p50": 2.835,
          "p75": 4.408,
          "p80": 4.939,
          "p90": 7.412,
          "p95": 47.611,
          "p99": 54.769,
          "p999": 61.716
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 116.45,
            "max": 116.45,
            "mean": 116.45
          },
          "cpu": {
            "min": 1.59,
            "max": 1.59,
            "mean": 1.59
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 30.41,
            "max": 30.41,
            "mean": 30.41
          },
          "cpu": {
            "min": 61.42,
            "max": 61.42,
            "mean": 61.42
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
      "throughput": 5961.65,
      "totalRequests": 178854,
      "latency": {
        "min": 0.37,
        "mean": 5.979,
        "max": 81.096,
        "pstdev": 11.586,
        "percentiles": {
          "p50": 2.705,
          "p75": 4.189,
          "p80": 4.699,
          "p90": 7.03,
          "p95": 46.987,
          "p99": 54.095,
          "p999": 61.222
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 128.02,
            "max": 128.02,
            "mean": 128.02
          },
          "cpu": {
            "min": 3.04,
            "max": 3.04,
            "mean": 3.04
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 36.55,
            "max": 36.55,
            "mean": 36.55
          },
          "cpu": {
            "min": 92.49,
            "max": 92.49,
            "mean": 92.49
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
      "throughput": 5921.3,
      "totalRequests": 177639,
      "latency": {
        "min": 0.376,
        "mean": 6.208,
        "max": 100.098,
        "pstdev": 11.901,
        "percentiles": {
          "p50": 2.77,
          "p75": 4.413,
          "p80": 5.009,
          "p90": 7.647,
          "p95": 46.934,
          "p99": 55.273,
          "p999": 68.509
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 144.93,
            "max": 144.93,
            "mean": 144.93
          },
          "cpu": {
            "min": 15.11,
            "max": 15.11,
            "mean": 15.11
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 58.75,
            "max": 58.75,
            "mean": 58.75
          },
          "cpu": {
            "min": 125.9,
            "max": 125.9,
            "mean": 125.9
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
      "throughput": 5876.24,
      "totalRequests": 176290,
      "latency": {
        "min": 0.365,
        "mean": 6.248,
        "max": 80.101,
        "pstdev": 12.134,
        "percentiles": {
          "p50": 2.759,
          "p75": 4.348,
          "p80": 4.922,
          "p90": 7.422,
          "p95": 47.855,
          "p99": 56.041,
          "p999": 67.145
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 155.94,
            "max": 155.94,
            "mean": 155.94
          },
          "cpu": {
            "min": 27.72,
            "max": 27.72,
            "mean": 27.72
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 80.95,
            "max": 80.95,
            "mean": 80.95
          },
          "cpu": {
            "min": 159.58,
            "max": 159.58,
            "mean": 159.58
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
      "throughput": 5744.1,
      "totalRequests": 172323,
      "latency": {
        "min": 0.38,
        "mean": 6.374,
        "max": 100.245,
        "pstdev": 12.424,
        "percentiles": {
          "p50": 2.775,
          "p75": 4.337,
          "p80": 4.899,
          "p90": 7.542,
          "p95": 47.699,
          "p99": 59.248,
          "p999": 71.761
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 184.41,
            "max": 184.41,
            "mean": 184.41
          },
          "cpu": {
            "min": 61.5,
            "max": 61.5,
            "mean": 61.5
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 129.19,
            "max": 129.19,
            "mean": 129.19
          },
          "cpu": {
            "min": 199.75,
            "max": 199.75,
            "mean": 199.75
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
      "throughput": 6028.67,
      "totalRequests": 180864,
      "latency": {
        "min": 0.361,
        "mean": 5.919,
        "max": 75.882,
        "pstdev": 11.427,
        "percentiles": {
          "p50": 2.729,
          "p75": 4.244,
          "p80": 4.783,
          "p90": 6.934,
          "p95": 46.608,
          "p99": 53.897,
          "p999": 59.893
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 229.41,
            "max": 229.41,
            "mean": 229.41
          },
          "cpu": {
            "min": 114.16,
            "max": 114.16,
            "mean": 114.16
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 121.69,
            "max": 121.69,
            "mean": 121.69
          },
          "cpu": {
            "min": 363.78,
            "max": 363.78,
            "mean": 363.78
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
      "throughput": 6093.21,
      "totalRequests": 182799,
      "latency": {
        "min": 0.364,
        "mean": 5.859,
        "max": 73.809,
        "pstdev": 11.454,
        "percentiles": {
          "p50": 2.634,
          "p75": 4.171,
          "p80": 4.702,
          "p90": 7.002,
          "p95": 46.548,
          "p99": 53.979,
          "p999": 62.793
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 236.91,
            "max": 236.91,
            "mean": 236.91
          },
          "cpu": {
            "min": 113.6,
            "max": 113.6,
            "mean": 113.6
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 122.3,
            "max": 122.3,
            "mean": 122.3
          },
          "cpu": {
            "min": 333.22,
            "max": 333.22,
            "mean": 333.22
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
      "throughput": 6057.56,
      "totalRequests": 181727,
      "latency": {
        "min": 0.366,
        "mean": 5.898,
        "max": 83.623,
        "pstdev": 11.56,
        "percentiles": {
          "p50": 2.628,
          "p75": 4.159,
          "p80": 4.695,
          "p90": 6.996,
          "p95": 46.897,
          "p99": 54.495,
          "p999": 61.849
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 143.71,
            "max": 143.71,
            "mean": 143.71
          },
          "cpu": {
            "min": 112.58,
            "max": 112.58,
            "mean": 112.58
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 121.84,
            "max": 121.84,
            "mean": 121.84
          },
          "cpu": {
            "min": 302.58,
            "max": 302.58,
            "mean": 302.58
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
      "throughput": 5899.93,
      "totalRequests": 176998,
      "latency": {
        "min": 0.329,
        "mean": 6.218,
        "max": 88.965,
        "pstdev": 12.007,
        "percentiles": {
          "p50": 2.818,
          "p75": 4.388,
          "p80": 4.92,
          "p90": 7.204,
          "p95": 47.871,
          "p99": 55.517,
          "p999": 65.886
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 154.24,
            "max": 154.24,
            "mean": 154.24
          },
          "cpu": {
            "min": 102.3,
            "max": 102.3,
            "mean": 102.3
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 129.34,
            "max": 129.34,
            "mean": 129.34
          },
          "cpu": {
            "min": 270.21,
            "max": 270.21,
            "mean": 270.21
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
      "throughput": 5862.2,
      "totalRequests": 175868,
      "latency": {
        "min": 0.359,
        "mean": 6.106,
        "max": 107.622,
        "pstdev": 11.962,
        "percentiles": {
          "p50": 2.648,
          "p75": 4.343,
          "p80": 4.933,
          "p90": 7.617,
          "p95": 47.341,
          "p99": 55.508,
          "p999": 69.103
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 153.72,
            "max": 153.72,
            "mean": 153.72
          },
          "cpu": {
            "min": 91.07,
            "max": 91.07,
            "mean": 91.07
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 129.18,
            "max": 129.18,
            "mean": 129.18
          },
          "cpu": {
            "min": 237.04,
            "max": 237.04,
            "mean": 237.04
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37
    }
  ]
};

export default benchmarkData;
