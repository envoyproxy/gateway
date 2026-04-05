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
        "min": 0.367,
        "mean": 6.485,
        "max": 65.673,
        "pstdev": 11.042,
        "percentiles": {
          "p50": 3.222,
          "p75": 5.166,
          "p80": 5.879,
          "p90": 9.253,
          "p95": 43.97,
          "p99": 52.201,
          "p999": 56.854
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
        "min": 0.355,
        "mean": 6.582,
        "max": 88.674,
        "pstdev": 11.849,
        "percentiles": {
          "p50": 3.134,
          "p75": 4.984,
          "p80": 5.597,
          "p90": 8.545,
          "p95": 46.972,
          "p99": 53.663,
          "p999": 59.625
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
        "min": 0.384,
        "mean": 4.221,
        "max": 63.5,
        "pstdev": 9.204,
        "percentiles": {
          "p50": 2.039,
          "p75": 3.063,
          "p80": 3.399,
          "p90": 4.605,
          "p95": 7.844,
          "p99": 49.758,
          "p999": 54.032
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
        "min": 0.393,
        "mean": 7.209,
        "max": 93.327,
        "pstdev": 12.64,
        "percentiles": {
          "p50": 3.281,
          "p75": 5.443,
          "p80": 6.241,
          "p90": 10.258,
          "p95": 48.494,
          "p99": 55.793,
          "p999": 65.497
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
        "min": 0.37,
        "mean": 6.797,
        "max": 91.549,
        "pstdev": 12.276,
        "percentiles": {
          "p50": 3.145,
          "p75": 5.052,
          "p80": 5.728,
          "p90": 8.916,
          "p95": 47.915,
          "p99": 54.947,
          "p999": 65.368
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
        "min": 0.367,
        "mean": 6.908,
        "max": 97.939,
        "pstdev": 12.608,
        "percentiles": {
          "p50": 3.176,
          "p75": 5.046,
          "p80": 5.68,
          "p90": 8.77,
          "p95": 48.472,
          "p99": 56.449,
          "p999": 68.317
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
        "min": 0.386,
        "mean": 6.69,
        "max": 71.688,
        "pstdev": 12.144,
        "percentiles": {
          "p50": 3.075,
          "p75": 4.937,
          "p80": 5.61,
          "p90": 8.727,
          "p95": 48.252,
          "p99": 54.595,
          "p999": 59.408
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
        "min": 0.378,
        "mean": 7.119,
        "max": 85.032,
        "pstdev": 12.404,
        "percentiles": {
          "p50": 3.39,
          "p75": 5.34,
          "p80": 6.015,
          "p90": 9.3,
          "p95": 48.457,
          "p99": 55.005,
          "p999": 60.284
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
        "min": 0.357,
        "mean": 6.664,
        "max": 77.307,
        "pstdev": 12.03,
        "percentiles": {
          "p50": 3.14,
          "p75": 4.935,
          "p80": 5.56,
          "p90": 8.456,
          "p95": 47.661,
          "p99": 54.544,
          "p999": 61.11
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
        "min": 0.376,
        "mean": 6.544,
        "max": 86.54,
        "pstdev": 11.908,
        "percentiles": {
          "p50": 3.079,
          "p75": 4.927,
          "p80": 5.554,
          "p90": 8.36,
          "p95": 47.069,
          "p99": 54.265,
          "p999": 62.601
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
        "min": 0.395,
        "mean": 6.772,
        "max": 99.7,
        "pstdev": 12.345,
        "percentiles": {
          "p50": 3.118,
          "p75": 4.954,
          "p80": 5.598,
          "p90": 8.672,
          "p95": 48.261,
          "p99": 55.498,
          "p999": 65.902
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
