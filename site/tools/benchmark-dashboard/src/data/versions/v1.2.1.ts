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
        "min": 337,
        "mean": 6415,
        "max": 92647,
        "pstdev": 11350,
        "percentiles": {
          "p50": 3107,
          "p75": 5037,
          "p80": 5711,
          "p90": 8782,
          "p95": 45103,
          "p99": 53942,
          "p999": 60790
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
        "min": 377,
        "mean": 6244,
        "max": 72724,
        "pstdev": 11870,
        "percentiles": {
          "p50": 2835,
          "p75": 4408,
          "p80": 4939,
          "p90": 7412,
          "p95": 47611,
          "p99": 54769,
          "p999": 61716
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
        "min": 370,
        "mean": 5979,
        "max": 81096,
        "pstdev": 11586,
        "percentiles": {
          "p50": 2705,
          "p75": 4189,
          "p80": 4699,
          "p90": 7030,
          "p95": 46987,
          "p99": 54095,
          "p999": 61222
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
        "min": 376,
        "mean": 6208,
        "max": 100098,
        "pstdev": 11901,
        "percentiles": {
          "p50": 2770,
          "p75": 4413,
          "p80": 5009,
          "p90": 7647,
          "p95": 46934,
          "p99": 55273,
          "p999": 68509
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
        "min": 365,
        "mean": 6248,
        "max": 80101,
        "pstdev": 12134,
        "percentiles": {
          "p50": 2759,
          "p75": 4348,
          "p80": 4922,
          "p90": 7422,
          "p95": 47855,
          "p99": 56041,
          "p999": 67145
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
        "min": 380,
        "mean": 6374,
        "max": 100245,
        "pstdev": 12424,
        "percentiles": {
          "p50": 2775,
          "p75": 4337,
          "p80": 4899,
          "p90": 7542,
          "p95": 47699,
          "p99": 59248,
          "p999": 71761
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
        "min": 361,
        "mean": 5919,
        "max": 75882,
        "pstdev": 11427,
        "percentiles": {
          "p50": 2729,
          "p75": 4244,
          "p80": 4783,
          "p90": 6934,
          "p95": 46608,
          "p99": 53897,
          "p999": 59893
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
        "min": 364,
        "mean": 5859,
        "max": 73809,
        "pstdev": 11454,
        "percentiles": {
          "p50": 2634,
          "p75": 4171,
          "p80": 4702,
          "p90": 7002,
          "p95": 46548,
          "p99": 53979,
          "p999": 62793
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
        "min": 366,
        "mean": 5898,
        "max": 83623,
        "pstdev": 11560,
        "percentiles": {
          "p50": 2628,
          "p75": 4159,
          "p80": 4695,
          "p90": 6996,
          "p95": 46897,
          "p99": 54495,
          "p999": 61849
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
        "min": 329,
        "mean": 6218,
        "max": 88965,
        "pstdev": 12007,
        "percentiles": {
          "p50": 2818,
          "p75": 4388,
          "p80": 4920,
          "p90": 7204,
          "p95": 47871,
          "p99": 55517,
          "p999": 65886
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
        "min": 359,
        "mean": 6106,
        "max": 107622,
        "pstdev": 11962,
        "percentiles": {
          "p50": 2648,
          "p75": 4343,
          "p80": 4933,
          "p90": 7617,
          "p95": 47341,
          "p99": 55508,
          "p999": 69103
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
