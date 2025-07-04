import { TestSuite } from '../types';

// Benchmark data extracted from markdown report for version 1.4.0
// Generated on 2025-06-17T19:50:26.781Z

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.4.0",
    "runId": "1.4.0-1750189826781",
    "date": "2025-05-14",
    "environment": "GitHub CI",
    "description": "Benchmark report for Envoy Gateway 1.4.0",
    "downloadUrl": "https://github.com/envoyproxy/gateway/releases/download/v1.4.0/benchmark_report.zip",
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
      "testName": "scaling up httproutes to 10 with 2 routes per hostname",
      "routes": 10,
      "routesPerHostname": 2,
      "phase": "scaling-up",
      "throughput": 5433.23,
      "totalRequests": 163000,
      "latency": {
        "min": 367,
        "mean": 6485,
        "max": 65673,
        "pstdev": 11042,
        "percentiles": {
          "p50": 3222,
          "p75": 5166,
          "p80": 5879,
          "p90": 9253,
          "p95": 43970,
          "p99": 52201,
          "p999": 56854
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 128.52,
            "max": 156.24,
            "mean": 150.28
          },
          "cpu": {
            "min": 0.13,
            "max": 0.67,
            "mean": 0.44
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 0,
            "max": 27.01,
            "mean": 21.73
          },
          "cpu": {
            "min": 0,
            "max": 72.36,
            "mean": 2.32
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38
    },
    {
      "testName": "scaling up httproutes to 50 with 10 routes per hostname",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-up",
      "throughput": 5340.69,
      "totalRequests": 160221,
      "latency": {
        "min": 355,
        "mean": 6582,
        "max": 88674,
        "pstdev": 11849,
        "percentiles": {
          "p50": 3134,
          "p75": 4984,
          "p80": 5597,
          "p90": 8545,
          "p95": 46972,
          "p99": 53663,
          "p999": 59625
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 155.3,
            "max": 169.25,
            "mean": 159.23
          },
          "cpu": {
            "min": 0.27,
            "max": 4.73,
            "mean": 0.91
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 26.64,
            "max": 33.18,
            "mean": 32.22
          },
          "cpu": {
            "min": 0,
            "max": 99.91,
            "mean": 11.1
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37
    },
    {
      "testName": "scaling up httproutes to 100 with 20 routes per hostname",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-up",
      "throughput": 5205.41,
      "totalRequests": 156162,
      "latency": {
        "min": 384,
        "mean": 4221,
        "max": 63500,
        "pstdev": 9204,
        "percentiles": {
          "p50": 2039,
          "p75": 3063,
          "p80": 3399,
          "p90": 4605,
          "p95": 7844,
          "p99": 49758,
          "p999": 54032
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 156.77,
            "max": 174.36,
            "mean": 168.53
          },
          "cpu": {
            "min": 0.4,
            "max": 8.8,
            "mean": 1.37
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 32.83,
            "max": 37.05,
            "mean": 36.58
          },
          "cpu": {
            "min": 0,
            "max": 99.97,
            "mean": 4.39
          }
        }
      },
      "poolOverflow": 377,
      "upstreamConnections": 23
    },
    {
      "testName": "scaling up httproutes to 300 with 60 routes per hostname",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-up",
      "throughput": 5271.84,
      "totalRequests": 158157,
      "latency": {
        "min": 393,
        "mean": 7209,
        "max": 93327,
        "pstdev": 12640,
        "percentiles": {
          "p50": 3281,
          "p75": 5443,
          "p80": 6241,
          "p90": 10258,
          "p95": 48494,
          "p99": 55793,
          "p999": 65497
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 179.66,
            "max": 187.8,
            "mean": 184.81
          },
          "cpu": {
            "min": 0.53,
            "max": 27.13,
            "mean": 4.87
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 53.06,
            "max": 57.31,
            "mean": 56.58
          },
          "cpu": {
            "min": 0,
            "max": 99.96,
            "mean": 15.82
          }
        }
      },
      "poolOverflow": 360,
      "upstreamConnections": 40
    },
    {
      "testName": "scaling up httproutes to 500 with 100 routes per hostname",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-up",
      "throughput": 5351.64,
      "totalRequests": 160555,
      "latency": {
        "min": 370,
        "mean": 6797,
        "max": 91549,
        "pstdev": 12276,
        "percentiles": {
          "p50": 3145,
          "p75": 5052,
          "p80": 5728,
          "p90": 8916,
          "p95": 47915,
          "p99": 54947,
          "p999": 65368
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 186.4,
            "max": 200.27,
            "mean": 195.8
          },
          "cpu": {
            "min": 0.4,
            "max": 26.07,
            "mean": 1.71
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 71.16,
            "max": 77.36,
            "mean": 77
          },
          "cpu": {
            "min": 0,
            "max": 94.88,
            "mean": 3.39
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38
    },
    {
      "testName": "scaling up httproutes to 1000 with 200 routes per hostname",
      "routes": 1000,
      "routesPerHostname": 200,
      "phase": "scaling-up",
      "throughput": 5239.17,
      "totalRequests": 157179,
      "latency": {
        "min": 367,
        "mean": 6908,
        "max": 97939,
        "pstdev": 12608,
        "percentiles": {
          "p50": 3176,
          "p75": 5046,
          "p80": 5680,
          "p90": 8770,
          "p95": 48472,
          "p99": 56449,
          "p999": 68317
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 220.03,
            "max": 237.27,
            "mean": 233.38
          },
          "cpu": {
            "min": 0.07,
            "max": 1.2,
            "mean": 0.78
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 129.13,
            "max": 129.59,
            "mean": 129.34
          },
          "cpu": {
            "min": 0,
            "max": 74.43,
            "mean": 2.24
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38
    },
    {
      "testName": "scaling down httproutes to 10 with 2 routes per hostname",
      "routes": 10,
      "routesPerHostname": 2,
      "phase": "scaling-down",
      "throughput": 5409.79,
      "totalRequests": 162296,
      "latency": {
        "min": 386,
        "mean": 6690,
        "max": 71688,
        "pstdev": 12144,
        "percentiles": {
          "p50": 3075,
          "p75": 4937,
          "p80": 5610,
          "p90": 8727,
          "p95": 48252,
          "p99": 54595,
          "p999": 59408
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 165.91,
            "max": 176.04,
            "mean": 168.11
          },
          "cpu": {
            "min": 0.67,
            "max": 3.73,
            "mean": 1.23
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 123.28,
            "max": 123.45,
            "mean": 123.35
          },
          "cpu": {
            "min": 0,
            "max": 99.72,
            "mean": 9.08
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38
    },
    {
      "testName": "scaling down httproutes to 50 with 10 routes per hostname",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-down",
      "throughput": 5480.62,
      "totalRequests": 164426,
      "latency": {
        "min": 378,
        "mean": 7119,
        "max": 85032,
        "pstdev": 12404,
        "percentiles": {
          "p50": 3390,
          "p75": 5340,
          "p80": 6015,
          "p90": 9300,
          "p95": 48457,
          "p99": 55005,
          "p999": 60284
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 165.36,
            "max": 171.88,
            "mean": 169.47
          },
          "cpu": {
            "min": 0.6,
            "max": 6.8,
            "mean": 1.44
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 123.27,
            "max": 123.97,
            "mean": 123.42
          },
          "cpu": {
            "min": 0,
            "max": 100.05,
            "mean": 11.5
          }
        }
      },
      "poolOverflow": 359,
      "upstreamConnections": 41
    },
    {
      "testName": "scaling down httproutes to 100 with 20 routes per hostname",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-down",
      "throughput": 5413.63,
      "totalRequests": 162409,
      "latency": {
        "min": 357,
        "mean": 6664,
        "max": 77307,
        "pstdev": 12030,
        "percentiles": {
          "p50": 3140,
          "p75": 4935,
          "p80": 5560,
          "p90": 8456,
          "p95": 47661,
          "p99": 54544,
          "p999": 61110
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 170.86,
            "max": 186.68,
            "mean": 173.53
          },
          "cpu": {
            "min": 0.67,
            "max": 67.27,
            "mean": 7.29
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 123.27,
            "max": 129.61,
            "mean": 124.04
          },
          "cpu": {
            "min": 0,
            "max": 100,
            "mean": 5.99
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38
    },
    {
      "testName": "scaling down httproutes to 300 with 60 routes per hostname",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-down",
      "throughput": 5356.13,
      "totalRequests": 160684,
      "latency": {
        "min": 376,
        "mean": 6544,
        "max": 86540,
        "pstdev": 11908,
        "percentiles": {
          "p50": 3079,
          "p75": 4927,
          "p80": 5554,
          "p90": 8360,
          "p95": 47069,
          "p99": 54265,
          "p999": 62601
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 176.87,
            "max": 192.99,
            "mean": 189.97
          },
          "cpu": {
            "min": 0.6,
            "max": 73.6,
            "mean": 5.05
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 129.54,
            "max": 129.76,
            "mean": 129.65
          },
          "cpu": {
            "min": 0,
            "max": 100.08,
            "mean": 11.45
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37
    },
    {
      "testName": "scaling down httproutes to 500 with 100 routes per hostname",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-down",
      "throughput": 5339.99,
      "totalRequests": 160200,
      "latency": {
        "min": 395,
        "mean": 6772,
        "max": 99700,
        "pstdev": 12345,
        "percentiles": {
          "p50": 3118,
          "p75": 4954,
          "p80": 5598,
          "p90": 8672,
          "p95": 48261,
          "p99": 55498,
          "p999": 65902
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 199.81,
            "max": 210.76,
            "mean": 205.85
          },
          "cpu": {
            "min": 0.4,
            "max": 31.13,
            "mean": 2.01
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 129.54,
            "max": 129.74,
            "mean": 129.6
          },
          "cpu": {
            "min": 0,
            "max": 99.87,
            "mean": 8.99
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38
    }
  ]
};

export default benchmarkData;
