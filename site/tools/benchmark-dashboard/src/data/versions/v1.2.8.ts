import { TestSuite } from '../types';

// Benchmark data extracted from markdown report for version 1.2.8
// Generated on 2025-06-17T19:50:26.771Z

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.2.8",
    "runId": "1.2.8-1750189826771",
    "date": "2025-03-25",
    "environment": "GitHub CI",
    "description": "Benchmark report for Envoy Gateway 1.2.8",
    "downloadUrl": "https://github.com/envoyproxy/gateway/releases/download/v1.2.8/benchmark_report.zip",
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
      "throughput": 5878.23,
      "totalRequests": 176347,
      "latency": {
        "min": 0.376,
        "mean": 9.154,
        "max": 75.661,
        "pstdev": 14.181,
        "percentiles": {
          "p50": 4.294,
          "p75": 7.182,
          "p80": 8.292,
          "p90": 22.716,
          "p95": 52.056,
          "p99": 58.071,
          "p999": 64.841
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 108.89,
            "max": 108.89,
            "mean": 108.89
          },
          "cpu": {
            "min": 0.77,
            "max": 0.77,
            "mean": 0.77
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 28.23,
            "max": 28.23,
            "mean": 28.23
          },
          "cpu": {
            "min": 30.48,
            "max": 30.48,
            "mean": 30.48
          }
        }
      },
      "poolOverflow": 342,
      "upstreamConnections": 58
    },
    {
      "testName": "scale-up-httproutes-50",
      "routes": 50,
      "routesPerHostname": 1,
      "phase": "scaling-up",
      "throughput": 5582.21,
      "totalRequests": 167469,
      "latency": {
        "min": 0.359,
        "mean": 6.378,
        "max": 69.967,
        "pstdev": 12.106,
        "percentiles": {
          "p50": 2.864,
          "p75": 4.529,
          "p80": 5.113,
          "p90": 7.717,
          "p95": 48.431,
          "p99": 54.521,
          "p999": 59.967
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 132.62,
            "max": 132.62,
            "mean": 132.62
          },
          "cpu": {
            "min": 1.54,
            "max": 1.54,
            "mean": 1.54
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 32.39,
            "max": 32.39,
            "mean": 32.39
          },
          "cpu": {
            "min": 60.99,
            "max": 60.99,
            "mean": 60.99
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
      "throughput": 5454.1,
      "totalRequests": 163623,
      "latency": {
        "min": 0.368,
        "mean": 6.738,
        "max": 72.437,
        "pstdev": 12.216,
        "percentiles": {
          "p50": 3.318,
          "p75": 5.037,
          "p80": 5.581,
          "p90": 8.092,
          "p95": 47.886,
          "p99": 53.829,
          "p999": 60.2
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 135.66,
            "max": 135.66,
            "mean": 135.66
          },
          "cpu": {
            "min": 2.91,
            "max": 2.91,
            "mean": 2.91
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 38.6,
            "max": 38.6,
            "mean": 38.6
          },
          "cpu": {
            "min": 91.91,
            "max": 91.91,
            "mean": 91.91
          }
        }
      },
      "poolOverflow": 360,
      "upstreamConnections": 40
    },
    {
      "testName": "scale-up-httproutes-300",
      "routes": 300,
      "routesPerHostname": 1,
      "phase": "scaling-up",
      "throughput": 5496.5,
      "totalRequests": 164898,
      "latency": {
        "min": 0.375,
        "mean": 6.453,
        "max": 103.043,
        "pstdev": 12.191,
        "percentiles": {
          "p50": 2.887,
          "p75": 4.603,
          "p80": 5.209,
          "p90": 7.967,
          "p95": 48.15,
          "p99": 54.63,
          "p999": 63.743
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 147.15,
            "max": 147.15,
            "mean": 147.15
          },
          "cpu": {
            "min": 15.47,
            "max": 15.47,
            "mean": 15.47
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 60.75,
            "max": 60.75,
            "mean": 60.75
          },
          "cpu": {
            "min": 125.31,
            "max": 125.31,
            "mean": 125.31
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
      "throughput": 5351.8,
      "totalRequests": 160554,
      "latency": {
        "min": 0.377,
        "mean": 6.516,
        "max": 93.298,
        "pstdev": 12.549,
        "percentiles": {
          "p50": 2.857,
          "p75": 4.496,
          "p80": 5.076,
          "p90": 7.615,
          "p95": 49.512,
          "p99": 55.822,
          "p999": 65.318
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 155.79,
            "max": 155.79,
            "mean": 155.79
          },
          "cpu": {
            "min": 28.01,
            "max": 28.01,
            "mean": 28.01
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 82.96,
            "max": 82.96,
            "mean": 82.96
          },
          "cpu": {
            "min": 159.29,
            "max": 159.29,
            "mean": 159.29
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37
    },
    {
      "testName": "scale-up-httproutes-1000",
      "routes": 1000,
      "routesPerHostname": 1,
      "phase": "scaling-up",
      "throughput": 5102.69,
      "totalRequests": 153086,
      "latency": {
        "min": 0.358,
        "mean": 6.968,
        "max": 108.298,
        "pstdev": 13.304,
        "percentiles": {
          "p50": 2.979,
          "p75": 4.802,
          "p80": 5.441,
          "p90": 8.343,
          "p95": 50.37,
          "p99": 60.203,
          "p999": 71.446
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 190.03,
            "max": 190.03,
            "mean": 190.03
          },
          "cpu": {
            "min": 61.69,
            "max": 61.69,
            "mean": 61.69
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 133.22,
            "max": 133.22,
            "mean": 133.22
          },
          "cpu": {
            "min": 199.84,
            "max": 199.84,
            "mean": 199.84
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
      "throughput": 5424.03,
      "totalRequests": 162721,
      "latency": {
        "min": 0.362,
        "mean": 6.57,
        "max": 73.609,
        "pstdev": 12.562,
        "percentiles": {
          "p50": 2.923,
          "p75": 4.626,
          "p80": 5.209,
          "p90": 7.751,
          "p95": 49.96,
          "p99": 55.525,
          "p999": 60.127
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 127.89,
            "max": 127.89,
            "mean": 127.89
          },
          "cpu": {
            "min": 114.69,
            "max": 114.69,
            "mean": 114.69
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 112.11,
            "max": 112.11,
            "mean": 112.11
          },
          "cpu": {
            "min": 363.01,
            "max": 363.01,
            "mean": 363.01
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
      "throughput": 5454.96,
      "totalRequests": 163653,
      "latency": {
        "min": 0.363,
        "mean": 6.397,
        "max": 73.404,
        "pstdev": 12.347,
        "percentiles": {
          "p50": 2.884,
          "p75": 4.452,
          "p80": 4.989,
          "p90": 7.3,
          "p95": 49.477,
          "p99": 55.369,
          "p999": 61.495
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 136.28,
            "max": 136.28,
            "mean": 136.28
          },
          "cpu": {
            "min": 114.05,
            "max": 114.05,
            "mean": 114.05
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 114.79,
            "max": 114.79,
            "mean": 114.79
          },
          "cpu": {
            "min": 332.54,
            "max": 332.54,
            "mean": 332.54
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
      "throughput": 5402.5,
      "totalRequests": 162075,
      "latency": {
        "min": 0.366,
        "mean": 6.575,
        "max": 127.266,
        "pstdev": 12.484,
        "percentiles": {
          "p50": 2.94,
          "p75": 4.6,
          "p80": 5.162,
          "p90": 7.727,
          "p95": 49.549,
          "p99": 55.334,
          "p999": 61.478
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 140.91,
            "max": 140.91,
            "mean": 140.91
          },
          "cpu": {
            "min": 112.87,
            "max": 112.87,
            "mean": 112.87
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 114.76,
            "max": 114.76,
            "mean": 114.76
          },
          "cpu": {
            "min": 301.85,
            "max": 301.85,
            "mean": 301.85
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
      "throughput": 5498.81,
      "totalRequests": 164969,
      "latency": {
        "min": 0.349,
        "mean": 6.325,
        "max": 84.213,
        "pstdev": 12.296,
        "percentiles": {
          "p50": 2.789,
          "p75": 4.396,
          "p80": 4.944,
          "p90": 7.293,
          "p95": 48.961,
          "p99": 55.212,
          "p999": 65.019
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 148.36,
            "max": 148.36,
            "mean": 148.36
          },
          "cpu": {
            "min": 102.53,
            "max": 102.53,
            "mean": 102.53
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 133.35,
            "max": 133.35,
            "mean": 133.35
          },
          "cpu": {
            "min": 269.43,
            "max": 269.43,
            "mean": 269.43
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
      "throughput": 5308.76,
      "totalRequests": 159263,
      "latency": {
        "min": 0.361,
        "mean": 6.727,
        "max": 98.807,
        "pstdev": 12.828,
        "percentiles": {
          "p50": 2.916,
          "p75": 4.65,
          "p80": 5.273,
          "p90": 8.161,
          "p95": 49.971,
          "p99": 56.84,
          "p999": 67.182
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 157.93,
            "max": 157.93,
            "mean": 157.93
          },
          "cpu": {
            "min": 91.26,
            "max": 91.26,
            "mean": 91.26
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 133.2,
            "max": 133.2,
            "mean": 133.2
          },
          "cpu": {
            "min": 236.02,
            "max": 236.02,
            "mean": 236.02
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38
    }
  ]
};

export default benchmarkData;
