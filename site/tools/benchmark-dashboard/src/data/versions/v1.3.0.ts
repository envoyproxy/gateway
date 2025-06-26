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
        "min": 371,
        "mean": 6861,
        "max": 69562,
        "pstdev": 12096,
        "percentiles": {
          "p50": 3223,
          "p75": 5219,
          "p80": 5926,
          "p90": 9360,
          "p95": 48029,
          "p99": 54945,
          "p999": 60096
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
        "min": 372,
        "mean": 6781,
        "max": 81436,
        "pstdev": 12685,
        "percentiles": {
          "p50": 3035,
          "p75": 4783,
          "p80": 5410,
          "p90": 8210,
          "p95": 50042,
          "p99": 55861,
          "p999": 61739
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
        "min": 386,
        "mean": 6832,
        "max": 83939,
        "pstdev": 12879,
        "percentiles": {
          "p50": 2957,
          "p75": 4826,
          "p80": 5504,
          "p90": 8587,
          "p95": 50454,
          "p99": 56524,
          "p999": 63150
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
        "min": 374,
        "mean": 6735,
        "max": 105820,
        "pstdev": 12667,
        "percentiles": {
          "p50": 2997,
          "p75": 4715,
          "p80": 5323,
          "p90": 8205,
          "p95": 49692,
          "p99": 56199,
          "p999": 64403
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
        "min": 345,
        "mean": 7139,
        "max": 89923,
        "pstdev": 13210,
        "percentiles": {
          "p50": 3191,
          "p75": 5039,
          "p80": 5680,
          "p90": 8824,
          "p95": 50767,
          "p99": 57397,
          "p999": 67665
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
        "min": 371,
        "mean": 6950,
        "max": 98525,
        "pstdev": 13111,
        "percentiles": {
          "p50": 3057,
          "p75": 4816,
          "p80": 5416,
          "p90": 8383,
          "p95": 50270,
          "p99": 57946,
          "p999": 69038
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
        "min": 376,
        "mean": 6870,
        "max": 73203,
        "pstdev": 12827,
        "percentiles": {
          "p50": 3060,
          "p75": 4817,
          "p80": 5428,
          "p90": 8280,
          "p95": 50618,
          "p99": 56178,
          "p999": 61280
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
        "min": 383,
        "mean": 6991,
        "max": 88027,
        "pstdev": 12998,
        "percentiles": {
          "p50": 3091,
          "p75": 4901,
          "p80": 5532,
          "p90": 8568,
          "p95": 50655,
          "p99": 56272,
          "p999": 62160
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
        "min": 380,
        "mean": 7027,
        "max": 78663,
        "pstdev": 12995,
        "percentiles": {
          "p50": 3129,
          "p75": 4963,
          "p80": 5600,
          "p90": 8642,
          "p95": 50726,
          "p99": 56254,
          "p999": 61509
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
        "min": 386,
        "mean": 6893,
        "max": 92049,
        "pstdev": 12874,
        "percentiles": {
          "p50": 3088,
          "p75": 4898,
          "p80": 5548,
          "p90": 8444,
          "p95": 50268,
          "p99": 56334,
          "p999": 65038
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
        "min": 375,
        "mean": 6786,
        "max": 84877,
        "pstdev": 12812,
        "percentiles": {
          "p50": 3009,
          "p75": 4734,
          "p80": 5340,
          "p90": 8236,
          "p95": 50247,
          "p99": 56647,
          "p999": 65042
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
