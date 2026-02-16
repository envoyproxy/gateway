import { TestSuite } from '../types';

// Benchmark data extracted from markdown report for version 1.1.2
// Generated on 2025-06-17T19:58:49.982Z

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.1.2",
    "runId": "1.1.2-1750190329982",
    "date": "2024-09-24",
    "environment": "GitHub CI",
    "description": "Benchmark report for Envoy Gateway 1.1.2",
    "downloadUrl": "https://github.com/envoyproxy/gateway/releases/download/v1.1.2/benchmark_report.zip",
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
      "throughput": 5649.72,
      "totalRequests": 169493,
      "latency": {
        "min": 356,
        "mean": 6560,
        "max": 72347,
        "pstdev": 11796,
        "percentiles": {
          "p50": 3044,
          "p75": 5011,
          "p80": 5690,
          "p90": 8937,
          "p95": 46880,
          "p99": 54560,
          "p999": 59316
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 86.72,
            "max": 86.72,
            "mean": 86.72
          },
          "cpu": {
            "min": 1.5666666666666667,
            "max": 1.5666666666666667,
            "mean": 1.5666666666666667
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 24.26,
            "max": 24.26,
            "mean": 24.26
          },
          "cpu": {
            "min": 101.63333333333333,
            "max": 101.63333333333333,
            "mean": 101.63333333333333
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
      "throughput": 5594.14,
      "totalRequests": 167826,
      "latency": {
        "min": 377,
        "mean": 6225,
        "max": 71024,
        "pstdev": 11980,
        "percentiles": {
          "p50": 2812,
          "p75": 4378,
          "p80": 4905,
          "p90": 7214,
          "p95": 48160,
          "p99": 54513,
          "p999": 59850
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 106.76,
            "max": 106.76,
            "mean": 106.76
          },
          "cpu": {
            "min": 7.7,
            "max": 7.7,
            "mean": 7.7
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 32.43,
            "max": 32.43,
            "mean": 32.43
          },
          "cpu": {
            "min": 204.53333333333333,
            "max": 204.53333333333333,
            "mean": 204.53333333333333
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
      "throughput": 5523.65,
      "totalRequests": 165711,
      "latency": {
        "min": 379,
        "mean": 6455,
        "max": 77815,
        "pstdev": 12317,
        "percentiles": {
          "p50": 2868,
          "p75": 4510,
          "p80": 5086,
          "p90": 7645,
          "p95": 49127,
          "p99": 55048,
          "p999": 61448
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 149.03,
            "max": 149.03,
            "mean": 149.03
          },
          "cpu": {
            "min": 29.333333333333332,
            "max": 29.333333333333332,
            "mean": 29.333333333333332
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 46.61,
            "max": 46.61,
            "mean": 46.61
          },
          "cpu": {
            "min": 309.7666666666667,
            "max": 309.7666666666667,
            "mean": 309.7666666666667
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
      "throughput": 5554.01,
      "totalRequests": 166624,
      "latency": {
        "min": 357,
        "mean": 6246,
        "max": 99753,
        "pstdev": 12051,
        "percentiles": {
          "p50": 2841,
          "p75": 4416,
          "p80": 4960,
          "p90": 7177,
          "p95": 48371,
          "p99": 54800,
          "p999": 63275
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 375.51,
            "max": 375.51,
            "mean": 375.51
          },
          "cpu": {
            "min": 615.1,
            "max": 615.1,
            "mean": 615.1
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 152.97,
            "max": 152.97,
            "mean": 152.97
          },
          "cpu": {
            "min": 492.4666666666667,
            "max": 492.4666666666667,
            "mean": 492.4666666666667
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
      "throughput": 5475.04,
      "totalRequests": 164256,
      "latency": {
        "min": 365,
        "mean": 6546,
        "max": 95539,
        "pstdev": 12646,
        "percentiles": {
          "p50": 2834,
          "p75": 4484,
          "p80": 5075,
          "p90": 7891,
          "p95": 49397,
          "p99": 56231,
          "p999": 67629
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 1340.6,
            "max": 1340.6,
            "mean": 1340.6
          },
          "cpu": {
            "min": 38.333333333333336,
            "max": 38.333333333333336,
            "mean": 38.333333333333336
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 281.71,
            "max": 281.71,
            "mean": 281.71
          },
          "cpu": {
            "min": 718.0666666666666,
            "max": 718.0666666666666,
            "mean": 718.0666666666666
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
      "throughput": 5657.68,
      "totalRequests": 169734,
      "latency": {
        "min": 355,
        "mean": 7320,
        "max": 87252,
        "pstdev": 13162,
        "percentiles": {
          "p50": 3307,
          "p75": 5218,
          "p80": 5885,
          "p90": 9252,
          "p95": 50671,
          "p99": 56332,
          "p999": 62406
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 371.29,
            "max": 371.29,
            "mean": 371.29
          },
          "cpu": {
            "min": 578.7666666666667,
            "max": 578.7666666666667,
            "mean": 578.7666666666667
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 273.57,
            "max": 273.57,
            "mean": 273.57
          },
          "cpu": {
            "min": 1224,
            "max": 1224,
            "mean": 1224
          }
        }
      },
      "poolOverflow": 356,
      "upstreamConnections": 44
    },
    {
      "testName": "scale-down-httproutes-50",
      "routes": 50,
      "routesPerHostname": 1,
      "phase": "scaling-down",
      "throughput": 5587.8,
      "totalRequests": 167636,
      "latency": {
        "min": 356,
        "mean": 6346,
        "max": 68095,
        "pstdev": 12097,
        "percentiles": {
          "p50": 2890,
          "p75": 4509,
          "p80": 5041,
          "p90": 7495,
          "p95": 48560,
          "p99": 54345,
          "p999": 60436
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 342.8,
            "max": 342.8,
            "mean": 342.8
          },
          "cpu": {
            "min": 573.3000000000001,
            "max": 573.3000000000001,
            "mean": 573.3000000000001
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 274.84,
            "max": 274.84,
            "mean": 274.84
          },
          "cpu": {
            "min": 1122.2333333333333,
            "max": 1122.2333333333333,
            "mean": 1122.2333333333333
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
      "throughput": 4098.74,
      "totalRequests": 122966,
      "latency": {
        "min": 347,
        "mean": 3660,
        "max": 99745,
        "pstdev": 7371,
        "percentiles": {
          "p50": 1658,
          "p75": 2735,
          "p80": 3147,
          "p90": 5306,
          "p95": 18030,
          "p99": 42778,
          "p999": 60440
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 433.01,
            "max": 433.01,
            "mean": 433.01
          },
          "cpu": {
            "min": 486.96666666666664,
            "max": 486.96666666666664,
            "mean": 486.96666666666664
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 281.27,
            "max": 281.27,
            "mean": 281.27
          },
          "cpu": {
            "min": 996.0333333333333,
            "max": 996.0333333333333,
            "mean": 996.0333333333333
          }
        }
      },
      "poolOverflow": 384,
      "upstreamConnections": 16
    },
    {
      "testName": "scale-down-httproutes-300",
      "routes": 300,
      "routesPerHostname": 1,
      "phase": "scaling-down",
      "throughput": 5306.46,
      "totalRequests": 159194,
      "latency": {
        "min": 360,
        "mean": 6884,
        "max": 106807,
        "pstdev": 12850,
        "percentiles": {
          "p50": 2983,
          "p75": 4873,
          "p80": 5573,
          "p90": 9215,
          "p95": 49102,
          "p99": 58521,
          "p999": 70213
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 232.76,
            "max": 232.76,
            "mean": 232.76
          },
          "cpu": {
            "min": 16.299999999999997,
            "max": 16.299999999999997,
            "mean": 16.299999999999997
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 282.68,
            "max": 282.68,
            "mean": 282.68
          },
          "cpu": {
            "min": 825.4333333333333,
            "max": 825.4333333333333,
            "mean": 825.4333333333333
          }
        }
      },
      "poolOverflow": 361,
      "upstreamConnections": 39
    }
  ]
};

export default benchmarkData;
