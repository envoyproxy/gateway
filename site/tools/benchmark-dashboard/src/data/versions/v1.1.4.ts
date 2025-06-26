import { TestSuite } from '../types';

// Benchmark data extracted from markdown report for version 1.1.4
// Generated on 2025-06-17T19:58:49.987Z

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.1.4",
    "runId": "1.1.4-1750190329987",
    "date": "2024-12-13",
    "environment": "GitHub CI",
    "description": "Benchmark report for Envoy Gateway 1.1.4",
    "downloadUrl": "https://github.com/envoyproxy/gateway/releases/download/v1.1.4/benchmark_report.zip",
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
      "throughput": 5971.73,
      "totalRequests": 179152,
      "latency": {
        "min": 354,
        "mean": 6247,
        "max": 97591,
        "pstdev": 11371,
        "percentiles": {
          "p50": 2948,
          "p75": 4694,
          "p80": 5352,
          "p90": 8420,
          "p95": 44519,
          "p99": 54398,
          "p999": 62244
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 82.79,
            "max": 82.79,
            "mean": 82.79
          },
          "cpu": {
            "min": 1.5000000000000002,
            "max": 1.5000000000000002,
            "mean": 1.5000000000000002
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 24.11,
            "max": 24.11,
            "mean": 24.11
          },
          "cpu": {
            "min": 101.53333333333335,
            "max": 101.53333333333335,
            "mean": 101.53333333333335
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
      "throughput": 5957.22,
      "totalRequests": 178717,
      "latency": {
        "min": 385,
        "mean": 6145,
        "max": 93360,
        "pstdev": 11838,
        "percentiles": {
          "p50": 2751,
          "p75": 4352,
          "p80": 4922,
          "p90": 7436,
          "p95": 47679,
          "p99": 54878,
          "p999": 62414
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 103.29,
            "max": 103.29,
            "mean": 103.29
          },
          "cpu": {
            "min": 6.800000000000001,
            "max": 6.800000000000001,
            "mean": 6.800000000000001
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 32.29,
            "max": 32.29,
            "mean": 32.29
          },
          "cpu": {
            "min": 204,
            "max": 204,
            "mean": 204
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
      "throughput": 5896.53,
      "totalRequests": 176899,
      "latency": {
        "min": 360,
        "mean": 6376,
        "max": 93995,
        "pstdev": 12077,
        "percentiles": {
          "p50": 2905,
          "p75": 4542,
          "p80": 5112,
          "p90": 7699,
          "p95": 48019,
          "p99": 55472,
          "p999": 63707
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 151.7,
            "max": 151.7,
            "mean": 151.7
          },
          "cpu": {
            "min": 32.599999999999994,
            "max": 32.599999999999994,
            "mean": 32.599999999999994
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 46.46,
            "max": 46.46,
            "mean": 46.46
          },
          "cpu": {
            "min": 309.8,
            "max": 309.8,
            "mean": 309.8
          }
        }
      },
      "poolOverflow": 361,
      "upstreamConnections": 39
    },
    {
      "testName": "scale-up-httproutes-300",
      "routes": 300,
      "routesPerHostname": 1,
      "phase": "scaling-up",
      "throughput": 5885.1,
      "totalRequests": 176553,
      "latency": {
        "min": 370,
        "mean": 6378,
        "max": 102457,
        "pstdev": 12096,
        "percentiles": {
          "p50": 2834,
          "p75": 4508,
          "p80": 5093,
          "p90": 7817,
          "p95": 48099,
          "p99": 55572,
          "p999": 64329
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 763.38,
            "max": 763.38,
            "mean": 763.38
          },
          "cpu": {
            "min": 605.6333333333333,
            "max": 605.6333333333333,
            "mean": 605.6333333333333
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 150.81,
            "max": 150.81,
            "mean": 150.81
          },
          "cpu": {
            "min": 494.66666666666674,
            "max": 494.66666666666674,
            "mean": 494.66666666666674
          }
        }
      },
      "poolOverflow": 361,
      "upstreamConnections": 39
    },
    {
      "testName": "scale-up-httproutes-500",
      "routes": 500,
      "routesPerHostname": 1,
      "phase": "scaling-up",
      "throughput": 5642.37,
      "totalRequests": 169274,
      "latency": {
        "min": 314,
        "mean": 6430,
        "max": 135987,
        "pstdev": 12216,
        "percentiles": {
          "p50": 2801,
          "p75": 4489,
          "p80": 5175,
          "p90": 8673,
          "p95": 46690,
          "p99": 56358,
          "p999": 73814
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 1603.28,
            "max": 1603.28,
            "mean": 1603.28
          },
          "cpu": {
            "min": 41.63333333333333,
            "max": 41.63333333333333,
            "mean": 41.63333333333333
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 289.75,
            "max": 289.75,
            "mean": 289.75
          },
          "cpu": {
            "min": 727.1666666666666,
            "max": 727.1666666666666,
            "mean": 727.1666666666666
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
      "throughput": 5993.8,
      "totalRequests": 179822,
      "latency": {
        "min": 381,
        "mean": 5965,
        "max": 84086,
        "pstdev": 11660,
        "percentiles": {
          "p50": 2666,
          "p75": 4169,
          "p80": 4715,
          "p90": 6995,
          "p95": 47218,
          "p99": 54775,
          "p999": 61200
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 117.16,
            "max": 117.16,
            "mean": 117.16
          },
          "cpu": {
            "min": 518.0333333333333,
            "max": 518.0333333333333,
            "mean": 518.0333333333333
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 278.72,
            "max": 278.72,
            "mean": 278.72
          },
          "cpu": {
            "min": 1205.8,
            "max": 1205.8,
            "mean": 1205.8
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
      "throughput": 5947.1,
      "totalRequests": 178412,
      "latency": {
        "min": 346,
        "mean": 5809,
        "max": 76681,
        "pstdev": 11155,
        "percentiles": {
          "p50": 2707,
          "p75": 4164,
          "p80": 4684,
          "p90": 6938,
          "p95": 45451,
          "p99": 53227,
          "p999": 59303
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 123.23,
            "max": 123.23,
            "mean": 123.23
          },
          "cpu": {
            "min": 512.5333333333333,
            "max": 512.5333333333333,
            "mean": 512.5333333333333
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 291.25,
            "max": 291.25,
            "mean": 291.25
          },
          "cpu": {
            "min": 1103.6333333333334,
            "max": 1103.6333333333334,
            "mean": 1103.6333333333334
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
      "throughput": 4884.46,
      "totalRequests": 146534,
      "latency": {
        "min": 362,
        "mean": 7148,
        "max": 137863,
        "pstdev": 11237,
        "percentiles": {
          "p50": 3375,
          "p75": 6266,
          "p80": 7621,
          "p90": 17699,
          "p95": 31992,
          "p99": 59576,
          "p999": 83935
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 338.12,
            "max": 338.12,
            "mean": 338.12
          },
          "cpu": {
            "min": 488.1333333333333,
            "max": 488.1333333333333,
            "mean": 488.1333333333333
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 294.41,
            "max": 294.41,
            "mean": 294.41
          },
          "cpu": {
            "min": 993,
            "max": 993,
            "mean": 993
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
      "throughput": 5737.62,
      "totalRequests": 172120,
      "latency": {
        "min": 345,
        "mean": 6192,
        "max": 117288,
        "pstdev": 11530,
        "percentiles": {
          "p50": 2822,
          "p75": 4535,
          "p80": 5155,
          "p90": 8463,
          "p95": 42631,
          "p99": 55584,
          "p999": 70291
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 679.73,
            "max": 679.73,
            "mean": 679.73
          },
          "cpu": {
            "min": 18.76666666666667,
            "max": 18.76666666666667,
            "mean": 18.76666666666667
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 285.93,
            "max": 285.93,
            "mean": 285.93
          },
          "cpu": {
            "min": 833.4333333333334,
            "max": 833.4333333333334,
            "mean": 833.4333333333334
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37
    }
  ]
};

export default benchmarkData;
