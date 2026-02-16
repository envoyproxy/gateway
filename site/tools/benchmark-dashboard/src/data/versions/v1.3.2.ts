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
        "min": 371,
        "mean": 6970,
        "max": 73949,
        "pstdev": 12106,
        "percentiles": {
          "p50": 3331,
          "p75": 5396,
          "p80": 6080,
          "p90": 9715,
          "p95": 47564,
          "p99": 55666,
          "p999": 60872
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
        "min": 380,
        "mean": 6752,
        "max": 83992,
        "pstdev": 12526,
        "percentiles": {
          "p50": 3089,
          "p75": 4875,
          "p80": 5484,
          "p90": 8253,
          "p95": 49682,
          "p99": 55398,
          "p999": 61507
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
        "min": 363,
        "mean": 7564,
        "max": 72122,
        "pstdev": 13492,
        "percentiles": {
          "p50": 3384,
          "p75": 5472,
          "p80": 6191,
          "p90": 9941,
          "p95": 51320,
          "p99": 57128,
          "p999": 63107
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
        "min": 397,
        "mean": 6832,
        "max": 82624,
        "pstdev": 12894,
        "percentiles": {
          "p50": 2990,
          "p75": 4798,
          "p80": 5418,
          "p90": 8366,
          "p95": 50391,
          "p99": 56737,
          "p999": 65114
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
        "min": 398,
        "mean": 7200,
        "max": 94949,
        "pstdev": 13256,
        "percentiles": {
          "p50": 3208,
          "p75": 5057,
          "p80": 5652,
          "p90": 8780,
          "p95": 50821,
          "p99": 57442,
          "p999": 66746
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
        "min": 361,
        "mean": 7088,
        "max": 127959,
        "pstdev": 13349,
        "percentiles": {
          "p50": 3052,
          "p75": 4885,
          "p80": 5529,
          "p90": 8785,
          "p95": 50700,
          "p99": 59123,
          "p999": 70660
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
        "min": 387,
        "mean": 9017,
        "max": 78245,
        "pstdev": 15046,
        "percentiles": {
          "p50": 3922,
          "p75": 6583,
          "p80": 7609,
          "p90": 16301,
          "p95": 54007,
          "p99": 59990,
          "p999": 67956
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
        "min": 391,
        "mean": 7071,
        "max": 80433,
        "pstdev": 12841,
        "percentiles": {
          "p50": 3199,
          "p75": 5005,
          "p80": 5635,
          "p90": 8730,
          "p95": 49813,
          "p99": 55756,
          "p999": 61407
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
        "min": 377,
        "mean": 6954,
        "max": 86982,
        "pstdev": 13049,
        "percentiles": {
          "p50": 3045,
          "p75": 4886,
          "p80": 5525,
          "p90": 8466,
          "p95": 51048,
          "p99": 56733,
          "p999": 62337
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
        "min": 389,
        "mean": 7977,
        "max": 86577,
        "pstdev": 14090,
        "percentiles": {
          "p50": 3498,
          "p75": 5731,
          "p80": 6471,
          "p90": 10947,
          "p95": 52355,
          "p99": 58748,
          "p999": 67379
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
        "min": 384,
        "mean": 7209,
        "max": 110186,
        "pstdev": 13357,
        "percentiles": {
          "p50": 3112,
          "p75": 5018,
          "p80": 5694,
          "p90": 9338,
          "p95": 51267,
          "p99": 57384,
          "p999": 67244
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
