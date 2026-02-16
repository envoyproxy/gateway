import { TestSuite } from '../types';

// Benchmark data extracted from markdown report for version 1.2.6
// Generated on 2025-06-17T19:50:26.767Z

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.2.6",
    "runId": "1.2.6-1750189826767",
    "date": "2025-01-23",
    "environment": "GitHub CI",
    "description": "Benchmark report for Envoy Gateway 1.2.6",
    "downloadUrl": "https://github.com/envoyproxy/gateway/releases/download/v1.2.6/benchmark_report.zip",
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
      "throughput": 5668.03,
      "totalRequests": 170041,
      "latency": {
        "min": 354,
        "mean": 6713,
        "max": 73658,
        "pstdev": 11906,
        "percentiles": {
          "p50": 3142,
          "p75": 5170,
          "p80": 5865,
          "p90": 9278,
          "p95": 47425,
          "p99": 54870,
          "p999": 59602
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 108.82,
            "max": 108.82,
            "mean": 108.82
          },
          "cpu": {
            "min": 0.76,
            "max": 0.76,
            "mean": 0.76
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 25.38,
            "max": 25.38,
            "mean": 25.38
          },
          "cpu": {
            "min": 30.51,
            "max": 30.51,
            "mean": 30.51
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
      "throughput": 5542.06,
      "totalRequests": 166262,
      "latency": {
        "min": 380,
        "mean": 6117,
        "max": 73658,
        "pstdev": 11949,
        "percentiles": {
          "p50": 2749,
          "p75": 4310,
          "p80": 4842,
          "p90": 7067,
          "p95": 48484,
          "p99": 54499,
          "p999": 59725
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 135.08,
            "max": 135.08,
            "mean": 135.08
          },
          "cpu": {
            "min": 1.53,
            "max": 1.53,
            "mean": 1.53
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 31.53,
            "max": 31.53,
            "mean": 31.53
          },
          "cpu": {
            "min": 61.17,
            "max": 61.17,
            "mean": 61.17
          }
        }
      },
      "poolOverflow": 364,
      "upstreamConnections": 36
    },
    {
      "testName": "scale-up-httproutes-100",
      "routes": 100,
      "routesPerHostname": 1,
      "phase": "scaling-up",
      "throughput": 5556.06,
      "totalRequests": 166682,
      "latency": {
        "min": 363,
        "mean": 6439,
        "max": 100163,
        "pstdev": 12297,
        "percentiles": {
          "p50": 2891,
          "p75": 4506,
          "p80": 5050,
          "p90": 7595,
          "p95": 49078,
          "p99": 54935,
          "p999": 60456
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 140.09,
            "max": 140.09,
            "mean": 140.09
          },
          "cpu": {
            "min": 2.88,
            "max": 2.88,
            "mean": 2.88
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 37.7,
            "max": 37.7,
            "mean": 37.7
          },
          "cpu": {
            "min": 91.98,
            "max": 91.98,
            "mean": 91.98
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
      "throughput": 5462.58,
      "totalRequests": 163882,
      "latency": {
        "min": 372,
        "mean": 6376,
        "max": 78544,
        "pstdev": 12289,
        "percentiles": {
          "p50": 2853,
          "p75": 4454,
          "p80": 5002,
          "p90": 7458,
          "p95": 49076,
          "p99": 55132,
          "p999": 63879
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 146.71,
            "max": 146.71,
            "mean": 146.71
          },
          "cpu": {
            "min": 15.64,
            "max": 15.64,
            "mean": 15.64
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 59.89,
            "max": 59.89,
            "mean": 59.89
          },
          "cpu": {
            "min": 125.41,
            "max": 125.41,
            "mean": 125.41
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
      "throughput": 5427.5,
      "totalRequests": 162825,
      "latency": {
        "min": 359,
        "mean": 6562,
        "max": 97148,
        "pstdev": 12598,
        "percentiles": {
          "p50": 2881,
          "p75": 4518,
          "p80": 5091,
          "p90": 7774,
          "p95": 49301,
          "p99": 56102,
          "p999": 67338
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 155.26,
            "max": 155.26,
            "mean": 155.26
          },
          "cpu": {
            "min": 28.52,
            "max": 28.52,
            "mean": 28.52
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 82.08,
            "max": 82.08,
            "mean": 82.08
          },
          "cpu": {
            "min": 159.72,
            "max": 159.72,
            "mean": 159.72
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
      "throughput": 5261.62,
      "totalRequests": 157849,
      "latency": {
        "min": 373,
        "mean": 6972,
        "max": 115060,
        "pstdev": 13251,
        "percentiles": {
          "p50": 3015,
          "p75": 4770,
          "p80": 5385,
          "p90": 8293,
          "p95": 50296,
          "p99": 59930,
          "p999": 70590
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 188.13,
            "max": 188.13,
            "mean": 188.13
          },
          "cpu": {
            "min": 61.78,
            "max": 61.78,
            "mean": 61.78
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 130.35,
            "max": 130.35,
            "mean": 130.35
          },
          "cpu": {
            "min": 199.1,
            "max": 199.1,
            "mean": 199.1
          }
        }
      },
      "poolOverflow": 361,
      "upstreamConnections": 39
    },
    {
      "testName": "scale-down-httproutes-10",
      "routes": 10,
      "routesPerHostname": 1,
      "phase": "scaling-down",
      "throughput": 5543.17,
      "totalRequests": 166299,
      "latency": {
        "min": 360,
        "mean": 6447,
        "max": 75276,
        "pstdev": 12396,
        "percentiles": {
          "p50": 2866,
          "p75": 4486,
          "p80": 5040,
          "p90": 7527,
          "p95": 49663,
          "p99": 55240,
          "p999": 60647
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 129.36,
            "max": 129.36,
            "mean": 129.36
          },
          "cpu": {
            "min": 115.21,
            "max": 115.21,
            "mean": 115.21
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 122.58,
            "max": 122.58,
            "mean": 122.58
          },
          "cpu": {
            "min": 362.6,
            "max": 362.6,
            "mean": 362.6
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
      "throughput": 5522.73,
      "totalRequests": 165682,
      "latency": {
        "min": 384,
        "mean": 6098,
        "max": 68321,
        "pstdev": 11922,
        "percentiles": {
          "p50": 2776,
          "p75": 4326,
          "p80": 4838,
          "p90": 6958,
          "p95": 48431,
          "p99": 54296,
          "p999": 58857
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 140.97,
            "max": 140.97,
            "mean": 140.97
          },
          "cpu": {
            "min": 114.57,
            "max": 114.57,
            "mean": 114.57
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 122.83,
            "max": 122.83,
            "mean": 122.83
          },
          "cpu": {
            "min": 332.12,
            "max": 332.12,
            "mean": 332.12
          }
        }
      },
      "poolOverflow": 364,
      "upstreamConnections": 36
    },
    {
      "testName": "scale-down-httproutes-100",
      "routes": 100,
      "routesPerHostname": 1,
      "phase": "scaling-down",
      "throughput": 5500.66,
      "totalRequests": 165023,
      "latency": {
        "min": 360,
        "mean": 6301,
        "max": 80400,
        "pstdev": 12064,
        "percentiles": {
          "p50": 2833,
          "p75": 4483,
          "p80": 5035,
          "p90": 7574,
          "p95": 48777,
          "p99": 54876,
          "p999": 60850
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 132.45,
            "max": 132.45,
            "mean": 132.45
          },
          "cpu": {
            "min": 113.5,
            "max": 113.5,
            "mean": 113.5
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 122.84,
            "max": 122.84,
            "mean": 122.84
          },
          "cpu": {
            "min": 301.26,
            "max": 301.26,
            "mean": 301.26
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
      "throughput": 5453.33,
      "totalRequests": 163603,
      "latency": {
        "min": 357,
        "mean": 6324,
        "max": 86024,
        "pstdev": 12138,
        "percentiles": {
          "p50": 2842,
          "p75": 4429,
          "p80": 4967,
          "p90": 7393,
          "p95": 48201,
          "p99": 54683,
          "p999": 65251
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 150.62,
            "max": 150.62,
            "mean": 150.62
          },
          "cpu": {
            "min": 103.11,
            "max": 103.11,
            "mean": 103.11
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 130.34,
            "max": 130.34,
            "mean": 130.34
          },
          "cpu": {
            "min": 269.03,
            "max": 269.03,
            "mean": 269.03
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
      "throughput": 5402.96,
      "totalRequests": 162089,
      "latency": {
        "min": 371,
        "mean": 6575,
        "max": 90804,
        "pstdev": 12621,
        "percentiles": {
          "p50": 2877,
          "p75": 4543,
          "p80": 5114,
          "p90": 7832,
          "p95": 49375,
          "p99": 56092,
          "p999": 69132
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 159.25,
            "max": 159.25,
            "mean": 159.25
          },
          "cpu": {
            "min": 91.66,
            "max": 91.66,
            "mean": 91.66
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 130.34,
            "max": 130.34,
            "mean": 130.34
          },
          "cpu": {
            "min": 235.83,
            "max": 235.83,
            "mean": 235.83
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38
    }
  ]
};

export default benchmarkData;
