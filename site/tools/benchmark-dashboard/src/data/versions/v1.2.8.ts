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
        "min": 376,
        "mean": 9154,
        "max": 75661,
        "pstdev": 14181,
        "percentiles": {
          "p50": 4294,
          "p75": 7182,
          "p80": 8292,
          "p90": 22716,
          "p95": 52056,
          "p99": 58071,
          "p999": 64841
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
        "min": 359,
        "mean": 6378,
        "max": 69967,
        "pstdev": 12106,
        "percentiles": {
          "p50": 2864,
          "p75": 4529,
          "p80": 5113,
          "p90": 7717,
          "p95": 48431,
          "p99": 54521,
          "p999": 59967
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
        "min": 368,
        "mean": 6738,
        "max": 72437,
        "pstdev": 12216,
        "percentiles": {
          "p50": 3318,
          "p75": 5037,
          "p80": 5581,
          "p90": 8092,
          "p95": 47886,
          "p99": 53829,
          "p999": 60200
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
        "min": 375,
        "mean": 6453,
        "max": 103043,
        "pstdev": 12191,
        "percentiles": {
          "p50": 2887,
          "p75": 4603,
          "p80": 5209,
          "p90": 7967,
          "p95": 48150,
          "p99": 54630,
          "p999": 63743
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
        "min": 377,
        "mean": 6516,
        "max": 93298,
        "pstdev": 12549,
        "percentiles": {
          "p50": 2857,
          "p75": 4496,
          "p80": 5076,
          "p90": 7615,
          "p95": 49512,
          "p99": 55822,
          "p999": 65318
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
        "min": 358,
        "mean": 6968,
        "max": 108298,
        "pstdev": 13304,
        "percentiles": {
          "p50": 2979,
          "p75": 4802,
          "p80": 5441,
          "p90": 8343,
          "p95": 50370,
          "p99": 60203,
          "p999": 71446
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
        "min": 362,
        "mean": 6570,
        "max": 73609,
        "pstdev": 12562,
        "percentiles": {
          "p50": 2923,
          "p75": 4626,
          "p80": 5209,
          "p90": 7751,
          "p95": 49960,
          "p99": 55525,
          "p999": 60127
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
        "min": 363,
        "mean": 6397,
        "max": 73404,
        "pstdev": 12347,
        "percentiles": {
          "p50": 2884,
          "p75": 4452,
          "p80": 4989,
          "p90": 7300,
          "p95": 49477,
          "p99": 55369,
          "p999": 61495
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
        "min": 366,
        "mean": 6575,
        "max": 127266,
        "pstdev": 12484,
        "percentiles": {
          "p50": 2940,
          "p75": 4600,
          "p80": 5162,
          "p90": 7727,
          "p95": 49549,
          "p99": 55334,
          "p999": 61478
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
        "min": 349,
        "mean": 6325,
        "max": 84213,
        "pstdev": 12296,
        "percentiles": {
          "p50": 2789,
          "p75": 4396,
          "p80": 4944,
          "p90": 7293,
          "p95": 48961,
          "p99": 55212,
          "p999": 65019
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
        "min": 361,
        "mean": 6727,
        "max": 98807,
        "pstdev": 12828,
        "percentiles": {
          "p50": 2916,
          "p75": 4650,
          "p80": 5273,
          "p90": 8161,
          "p95": 49971,
          "p99": 56840,
          "p999": 67182
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
