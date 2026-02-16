import { TestSuite } from '../types';

// Benchmark data extracted from markdown report for version 1.2.0
// Generated on 2025-06-17T19:50:26.754Z

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.2.0",
    "runId": "1.2.0-1750189826754",
    "date": "2024-11-06",
    "environment": "GitHub CI",
    "description": "Benchmark report for Envoy Gateway 1.2.0",
    "downloadUrl": "https://github.com/envoyproxy/gateway/releases/download/v1.2.0/benchmark_report.zip",
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
      "throughput": 6006.27,
      "totalRequests": 180191,
      "latency": {
        "min": 359,
        "mean": 6202,
        "max": 74358,
        "pstdev": 11262,
        "percentiles": {
          "p50": 2976,
          "p75": 4719,
          "p80": 5330,
          "p90": 8159,
          "p95": 44003,
          "p99": 54190,
          "p999": 59357
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 113.66,
            "max": 113.66,
            "mean": 113.66
          },
          "cpu": {
            "min": 0.75,
            "max": 0.75,
            "mean": 0.75
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 25.55,
            "max": 25.55,
            "mean": 25.55
          },
          "cpu": {
            "min": 30.41,
            "max": 30.41,
            "mean": 30.41
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
      "throughput": 5988.15,
      "totalRequests": 179650,
      "latency": {
        "min": 366,
        "mean": 6093,
        "max": 82386,
        "pstdev": 11670,
        "percentiles": {
          "p50": 2838,
          "p75": 4283,
          "p80": 4798,
          "p90": 6972,
          "p95": 47523,
          "p99": 54038,
          "p999": 59813
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 119.16,
            "max": 119.16,
            "mean": 119.16
          },
          "cpu": {
            "min": 1.53,
            "max": 1.53,
            "mean": 1.53
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 31.71,
            "max": 31.71,
            "mean": 31.71
          },
          "cpu": {
            "min": 61.03,
            "max": 61.03,
            "mean": 61.03
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
      "throughput": 5930.13,
      "totalRequests": 177905,
      "latency": {
        "min": 374,
        "mean": 6182,
        "max": 93884,
        "pstdev": 11844,
        "percentiles": {
          "p50": 2812,
          "p75": 4420,
          "p80": 4967,
          "p90": 7339,
          "p95": 47618,
          "p99": 54661,
          "p999": 61865
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 123.35,
            "max": 123.35,
            "mean": 123.35
          },
          "cpu": {
            "min": 2.89,
            "max": 2.89,
            "mean": 2.89
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 37.86,
            "max": 37.86,
            "mean": 37.86
          },
          "cpu": {
            "min": 91.87,
            "max": 91.87,
            "mean": 91.87
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
      "throughput": 5713.27,
      "totalRequests": 171401,
      "latency": {
        "min": 384,
        "mean": 6405,
        "max": 105545,
        "pstdev": 12313,
        "percentiles": {
          "p50": 2832,
          "p75": 4450,
          "p80": 5057,
          "p90": 7773,
          "p95": 48732,
          "p99": 56084,
          "p999": 66220
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 150.92,
            "max": 150.92,
            "mean": 150.92
          },
          "cpu": {
            "min": 15.08,
            "max": 15.08,
            "mean": 15.08
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 60.05,
            "max": 60.05,
            "mean": 60.05
          },
          "cpu": {
            "min": 125.09,
            "max": 125.09,
            "mean": 125.09
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
      "throughput": 5689.03,
      "totalRequests": 170675,
      "latency": {
        "min": 365,
        "mean": 5939,
        "max": 98639,
        "pstdev": 11585,
        "percentiles": {
          "p50": 2715,
          "p75": 4258,
          "p80": 4772,
          "p90": 7055,
          "p95": 45996,
          "p99": 54368,
          "p999": 68464
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 156.97,
            "max": 156.97,
            "mean": 156.97
          },
          "cpu": {
            "min": 27.62,
            "max": 27.62,
            "mean": 27.62
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 82.26,
            "max": 82.26,
            "mean": 82.26
          },
          "cpu": {
            "min": 158.86,
            "max": 158.86,
            "mean": 158.86
          }
        }
      },
      "poolOverflow": 365,
      "upstreamConnections": 35
    },
    {
      "testName": "scale-up-httproutes-1000",
      "routes": 1000,
      "routesPerHostname": 1,
      "phase": "scaling-up",
      "throughput": 5407.08,
      "totalRequests": 162220,
      "latency": {
        "min": 371,
        "mean": 6424,
        "max": 131579,
        "pstdev": 12473,
        "percentiles": {
          "p50": 2692,
          "p75": 4503,
          "p80": 5177,
          "p90": 8264,
          "p95": 47540,
          "p99": 58488,
          "p999": 72720
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 215.93,
            "max": 215.93,
            "mean": 215.93
          },
          "cpu": {
            "min": 61.56,
            "max": 61.56,
            "mean": 61.56
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 130.5,
            "max": 130.5,
            "mean": 130.5
          },
          "cpu": {
            "min": 197.45,
            "max": 197.45,
            "mean": 197.45
          }
        }
      },
      "poolOverflow": 364,
      "upstreamConnections": 36
    },
    {
      "testName": "scale-down-httproutes-10",
      "routes": 10,
      "routesPerHostname": 1,
      "phase": "scaling-down",
      "throughput": 5905.23,
      "totalRequests": 177157,
      "latency": {
        "min": 396,
        "mean": 6205,
        "max": 92979,
        "pstdev": 12065,
        "percentiles": {
          "p50": 2713,
          "p75": 4262,
          "p80": 4832,
          "p90": 7415,
          "p95": 48306,
          "p99": 55736,
          "p999": 63318
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 233.71,
            "max": 233.71,
            "mean": 233.71
          },
          "cpu": {
            "min": 114.53,
            "max": 114.53,
            "mean": 114.53
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 118.94,
            "max": 118.94,
            "mean": 118.94
          },
          "cpu": {
            "min": 360.39,
            "max": 360.39,
            "mean": 360.39
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
      "throughput": 5793.19,
      "totalRequests": 173800,
      "latency": {
        "min": 385,
        "mean": 6501,
        "max": 86843,
        "pstdev": 12452,
        "percentiles": {
          "p50": 2744,
          "p75": 4568,
          "p80": 5227,
          "p90": 8494,
          "p95": 48965,
          "p99": 56467,
          "p999": 65382
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 202.08,
            "max": 202.08,
            "mean": 202.08
          },
          "cpu": {
            "min": 113.96,
            "max": 113.96,
            "mean": 113.96
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 121.24,
            "max": 121.24,
            "mean": 121.24
          },
          "cpu": {
            "min": 329.91,
            "max": 329.91,
            "mean": 329.91
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
      "throughput": 5783.8,
      "totalRequests": 173514,
      "latency": {
        "min": 385,
        "mean": 6179,
        "max": 85405,
        "pstdev": 11889,
        "percentiles": {
          "p50": 2707,
          "p75": 4462,
          "p80": 5071,
          "p90": 7760,
          "p95": 47423,
          "p99": 54994,
          "p999": 64567
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 144.59,
            "max": 144.59,
            "mean": 144.59
          },
          "cpu": {
            "min": 112.93,
            "max": 112.93,
            "mean": 112.93
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 121.22,
            "max": 121.22,
            "mean": 121.22
          },
          "cpu": {
            "min": 299.28,
            "max": 299.28,
            "mean": 299.28
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
      "throughput": 5804.51,
      "totalRequests": 174139,
      "latency": {
        "min": 384,
        "mean": 6147,
        "max": 82661,
        "pstdev": 11892,
        "percentiles": {
          "p50": 2807,
          "p75": 4284,
          "p80": 4791,
          "p90": 7039,
          "p95": 47558,
          "p99": 55162,
          "p999": 66723
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 160.23,
            "max": 160.23,
            "mean": 160.23
          },
          "cpu": {
            "min": 102.85,
            "max": 102.85,
            "mean": 102.85
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 130.62,
            "max": 130.62,
            "mean": 130.62
          },
          "cpu": {
            "min": 267.09,
            "max": 267.09,
            "mean": 267.09
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
      "throughput": 5804.87,
      "totalRequests": 174149,
      "latency": {
        "min": 368,
        "mean": 5828,
        "max": 106119,
        "pstdev": 11561,
        "percentiles": {
          "p50": 2647,
          "p75": 4078,
          "p80": 4553,
          "p90": 6656,
          "p95": 46047,
          "p99": 55025,
          "p999": 66633
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 155.88,
            "max": 155.88,
            "mean": 155.88
          },
          "cpu": {
            "min": 91.5,
            "max": 91.5,
            "mean": 91.5
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 130.61,
            "max": 130.61,
            "mean": 130.61
          },
          "cpu": {
            "min": 234.47,
            "max": 234.47,
            "mean": 234.47
          }
        }
      },
      "poolOverflow": 365,
      "upstreamConnections": 35
    }
  ]
};

export default benchmarkData;
