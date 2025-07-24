import { TestSuite } from '../types';

// Benchmark data extracted from markdown report for version 1.2.7
// Generated on 2025-06-17T19:50:26.769Z

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.2.7",
    "runId": "1.2.7-1750189826769",
    "date": "2025-03-06",
    "environment": "GitHub CI",
    "description": "Benchmark report for Envoy Gateway 1.2.7",
    "downloadUrl": "https://github.com/envoyproxy/gateway/releases/download/v1.2.7/benchmark_report.zip",
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
      "throughput": 5984.83,
      "totalRequests": 179545,
      "latency": {
        "min": 402,
        "mean": 12033,
        "max": 92835,
        "pstdev": 16725,
        "percentiles": {
          "p50": 5655,
          "p75": 9809,
          "p80": 11804,
          "p90": 49971,
          "p95": 55760,
          "p99": 64241,
          "p999": 73961
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 112.35,
            "max": 112.35,
            "mean": 112.35
          },
          "cpu": {
            "min": 0.76,
            "max": 0.76,
            "mean": 0.76
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 28.29,
            "max": 28.29,
            "mean": 28.29
          },
          "cpu": {
            "min": 30.49,
            "max": 30.49,
            "mean": 30.49
          }
        }
      },
      "poolOverflow": 323,
      "upstreamConnections": 77
    },
    {
      "testName": "scale-up-httproutes-50",
      "routes": 50,
      "routesPerHostname": 1,
      "phase": "scaling-up",
      "throughput": 5564.85,
      "totalRequests": 166949,
      "latency": {
        "min": 362,
        "mean": 6416,
        "max": 73261,
        "pstdev": 12268,
        "percentiles": {
          "p50": 2855,
          "p75": 4514,
          "p80": 5094,
          "p90": 7660,
          "p95": 49031,
          "p99": 54939,
          "p999": 60436
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 125.94,
            "max": 125.94,
            "mean": 125.94
          },
          "cpu": {
            "min": 1.52,
            "max": 1.52,
            "mean": 1.52
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 34.46,
            "max": 34.46,
            "mean": 34.46
          },
          "cpu": {
            "min": 61.11,
            "max": 61.11,
            "mean": 61.11
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
      "throughput": 5618.23,
      "totalRequests": 168547,
      "latency": {
        "min": 372,
        "mean": 6834,
        "max": 85766,
        "pstdev": 12643,
        "percentiles": {
          "p50": 3107,
          "p75": 4954,
          "p80": 5560,
          "p90": 8454,
          "p95": 49485,
          "p99": 55513,
          "p999": 61272
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 132.05,
            "max": 132.05,
            "mean": 132.05
          },
          "cpu": {
            "min": 2.92,
            "max": 2.92,
            "mean": 2.92
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 40.63,
            "max": 40.63,
            "mean": 40.63
          },
          "cpu": {
            "min": 92.02,
            "max": 92.02,
            "mean": 92.02
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
      "throughput": 5506.26,
      "totalRequests": 165193,
      "latency": {
        "min": 383,
        "mean": 6670,
        "max": 85905,
        "pstdev": 12677,
        "percentiles": {
          "p50": 2946,
          "p75": 4659,
          "p80": 5243,
          "p90": 7891,
          "p95": 49846,
          "p99": 56096,
          "p999": 65185
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 146.09,
            "max": 146.09,
            "mean": 146.09
          },
          "cpu": {
            "min": 15.43,
            "max": 15.43,
            "mean": 15.43
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 62.86,
            "max": 62.86,
            "mean": 62.86
          },
          "cpu": {
            "min": 125.21,
            "max": 125.21,
            "mean": 125.21
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
      "throughput": 5580.92,
      "totalRequests": 167428,
      "latency": {
        "min": 310,
        "mean": 6716,
        "max": 89178,
        "pstdev": 12233,
        "percentiles": {
          "p50": 3157,
          "p75": 4861,
          "p80": 5446,
          "p90": 8231,
          "p95": 47419,
          "p99": 54634,
          "p999": 65587
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 163.49,
            "max": 163.49,
            "mean": 163.49
          },
          "cpu": {
            "min": 28.15,
            "max": 28.15,
            "mean": 28.15
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 83.07,
            "max": 83.07,
            "mean": 83.07
          },
          "cpu": {
            "min": 159.15,
            "max": 159.15,
            "mean": 159.15
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
      "throughput": 5298.3,
      "totalRequests": 158951,
      "latency": {
        "min": 369,
        "mean": 6537,
        "max": 122015,
        "pstdev": 12654,
        "percentiles": {
          "p50": 2914,
          "p75": 4472,
          "p80": 5012,
          "p90": 7402,
          "p95": 48639,
          "p99": 58097,
          "p999": 69103
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 187.81,
            "max": 187.81,
            "mean": 187.81
          },
          "cpu": {
            "min": 62.41,
            "max": 62.41,
            "mean": 62.41
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 133.27,
            "max": 133.27,
            "mean": 133.27
          },
          "cpu": {
            "min": 199.13,
            "max": 199.13,
            "mean": 199.13
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
      "throughput": 5623.57,
      "totalRequests": 168711,
      "latency": {
        "min": 380,
        "mean": 7190,
        "max": 75063,
        "pstdev": 13103,
        "percentiles": {
          "p50": 3211,
          "p75": 5102,
          "p80": 5772,
          "p90": 9144,
          "p95": 50722,
          "p99": 56260,
          "p999": 61515
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 132.38,
            "max": 132.38,
            "mean": 132.38
          },
          "cpu": {
            "min": 116.08,
            "max": 116.08,
            "mean": 116.08
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 128.69,
            "max": 128.69,
            "mean": 128.69
          },
          "cpu": {
            "min": 362.3,
            "max": 362.3,
            "mean": 362.3
          }
        }
      },
      "poolOverflow": 357,
      "upstreamConnections": 43
    },
    {
      "testName": "scale-down-httproutes-50",
      "routes": 50,
      "routesPerHostname": 1,
      "phase": "scaling-down",
      "throughput": 5516.59,
      "totalRequests": 165498,
      "latency": {
        "min": 329,
        "mean": 6044,
        "max": 69902,
        "pstdev": 11632,
        "percentiles": {
          "p50": 2893,
          "p75": 4337,
          "p80": 4790,
          "p90": 6742,
          "p95": 47155,
          "p99": 53372,
          "p999": 58073
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 144.02,
            "max": 144.02,
            "mean": 144.02
          },
          "cpu": {
            "min": 115.5,
            "max": 115.5,
            "mean": 115.5
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 129.14,
            "max": 129.14,
            "mean": 129.14
          },
          "cpu": {
            "min": 331.76,
            "max": 331.76,
            "mean": 331.76
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
      "throughput": 5515.82,
      "totalRequests": 165475,
      "latency": {
        "min": 380,
        "mean": 6483,
        "max": 79159,
        "pstdev": 12363,
        "percentiles": {
          "p50": 2872,
          "p75": 4545,
          "p80": 5131,
          "p90": 7760,
          "p95": 49190,
          "p99": 55408,
          "p999": 61890
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 135.79,
            "max": 135.79,
            "mean": 135.79
          },
          "cpu": {
            "min": 114.36,
            "max": 114.36,
            "mean": 114.36
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 109,
            "max": 109,
            "mean": 109
          },
          "cpu": {
            "min": 301.13,
            "max": 301.13,
            "mean": 301.13
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
      "throughput": 5403.53,
      "totalRequests": 162106,
      "latency": {
        "min": 354,
        "mean": 6429,
        "max": 83423,
        "pstdev": 12305,
        "percentiles": {
          "p50": 2864,
          "p75": 4496,
          "p80": 5072,
          "p90": 7504,
          "p95": 48885,
          "p99": 55414,
          "p999": 63565
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 155.42,
            "max": 155.42,
            "mean": 155.42
          },
          "cpu": {
            "min": 103.9,
            "max": 103.9,
            "mean": 103.9
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 133.59,
            "max": 133.59,
            "mean": 133.59
          },
          "cpu": {
            "min": 268.85,
            "max": 268.85,
            "mean": 268.85
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
      "throughput": 5405.24,
      "totalRequests": 162162,
      "latency": {
        "min": 387,
        "mean": 6615,
        "max": 90812,
        "pstdev": 12660,
        "percentiles": {
          "p50": 2913,
          "p75": 4593,
          "p80": 5174,
          "p90": 7866,
          "p95": 49461,
          "p99": 56510,
          "p999": 67248
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 165.58,
            "max": 165.58,
            "mean": 165.58
          },
          "cpu": {
            "min": 92.35,
            "max": 92.35,
            "mean": 92.35
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 133.42,
            "max": 133.42,
            "mean": 133.42
          },
          "cpu": {
            "min": 235.87,
            "max": 235.87,
            "mean": 235.87
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38
    }
  ]
};

export default benchmarkData;
