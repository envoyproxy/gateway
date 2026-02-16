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
        "min": 384,
        "mean": 7440,
        "max": 68399,
        "pstdev": 12191,
        "percentiles": {
          "p50": 3692,
          "p75": 5903,
          "p80": 6670,
          "p90": 10777,
          "p95": 47519,
          "p99": 54405,
          "p999": 59346
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
        "min": 392,
        "mean": 7067,
        "max": 77467,
        "pstdev": 12552,
        "percentiles": {
          "p50": 3330,
          "p75": 5225,
          "p80": 5887,
          "p90": 8906,
          "p95": 49113,
          "p99": 55103,
          "p999": 60295
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
        "min": 387,
        "mean": 7350,
        "max": 76185,
        "pstdev": 12732,
        "percentiles": {
          "p50": 3508,
          "p75": 5524,
          "p80": 6220,
          "p90": 9652,
          "p95": 49100,
          "p99": 55363,
          "p999": 62654
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
        "min": 390,
        "mean": 7376,
        "max": 78098,
        "pstdev": 12993,
        "percentiles": {
          "p50": 3429,
          "p75": 5431,
          "p80": 6116,
          "p90": 9633,
          "p95": 49809,
          "p99": 56043,
          "p999": 64651
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
        "min": 374,
        "mean": 7385,
        "max": 105918,
        "pstdev": 13085,
        "percentiles": {
          "p50": 3382,
          "p75": 5336,
          "p80": 6024,
          "p90": 9658,
          "p95": 50104,
          "p99": 56487,
          "p999": 66418
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
        "min": 374,
        "mean": 7335,
        "max": 94580,
        "pstdev": 13294,
        "percentiles": {
          "p50": 3255,
          "p75": 5279,
          "p80": 5998,
          "p90": 9532,
          "p95": 50294,
          "p99": 58908,
          "p999": 70729
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
        "min": 395,
        "mean": 7530,
        "max": 81936,
        "pstdev": 13142,
        "percentiles": {
          "p50": 3438,
          "p75": 5497,
          "p80": 6226,
          "p90": 10218,
          "p95": 50345,
          "p99": 56436,
          "p999": 61919
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
        "min": 368,
        "mean": 6911,
        "max": 72273,
        "pstdev": 12396,
        "percentiles": {
          "p50": 3227,
          "p75": 5075,
          "p80": 5695,
          "p90": 8751,
          "p95": 48676,
          "p99": 54822,
          "p999": 59865
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
        "min": 357,
        "mean": 6637,
        "max": 68100,
        "pstdev": 12012,
        "percentiles": {
          "p50": 3167,
          "p75": 4972,
          "p80": 5580,
          "p90": 8302,
          "p95": 47650,
          "p99": 53993,
          "p999": 59299
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
        "min": 379,
        "mean": 7234,
        "max": 111517,
        "pstdev": 12814,
        "percentiles": {
          "p50": 3381,
          "p75": 5331,
          "p80": 5993,
          "p90": 9269,
          "p95": 49465,
          "p99": 55562,
          "p999": 64862
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
        "min": 365,
        "mean": 7098,
        "max": 95834,
        "pstdev": 12843,
        "percentiles": {
          "p50": 3267,
          "p75": 5110,
          "p80": 5752,
          "p90": 8975,
          "p95": 49702,
          "p99": 56408,
          "p999": 67330
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
