import { TestSuite } from '../types';

// Benchmark data extracted from markdown report for version 1.2.4
// Generated on 2025-06-17T19:50:26.764Z

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.2.4",
    "runId": "1.2.4-1750189826764",
    "date": "2024-12-13",
    "environment": "GitHub CI",
    "description": "Benchmark report for Envoy Gateway 1.2.4",
    "downloadUrl": "https://github.com/envoyproxy/gateway/releases/download/v1.2.4/benchmark_report.zip",
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
      "throughput": 5982.38,
      "totalRequests": 179472,
      "latency": {
        "min": 0.349,
        "mean": 6.011,
        "max": 83.349,
        "pstdev": 11.061,
        "percentiles": {
          "p50": 2.829,
          "p75": 4.477,
          "p80": 5.098,
          "p90": 7.842,
          "p95": 43.255,
          "p99": 53.35,
          "p999": 59.346
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 110.54,
            "max": 110.54,
            "mean": 110.54
          },
          "cpu": {
            "min": 0.77,
            "max": 0.77,
            "mean": 0.77
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 24.25,
            "max": 24.25,
            "mean": 24.25
          },
          "cpu": {
            "min": 30.44,
            "max": 30.44,
            "mean": 30.44
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38
    },
    {
      "testName": "scale-up-httproutes-50",
      "routes": 50,
      "routesPerHostname": 1,
      "phase": "scaling-up",
      "throughput": 5917.46,
      "totalRequests": 177524,
      "latency": {
        "min": 0.361,
        "mean": 6.164,
        "max": 70.909,
        "pstdev": 11.809,
        "percentiles": {
          "p50": 2.843,
          "p75": 4.366,
          "p80": 4.876,
          "p90": 7.109,
          "p95": 47.788,
          "p99": 54.372,
          "p999": 59.932
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 119.89,
            "max": 119.89,
            "mean": 119.89
          },
          "cpu": {
            "min": 1.61,
            "max": 1.61,
            "mean": 1.61
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 30.43,
            "max": 30.43,
            "mean": 30.43
          },
          "cpu": {
            "min": 61.13,
            "max": 61.13,
            "mean": 61.13
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
      "throughput": 5930.25,
      "totalRequests": 177910,
      "latency": {
        "min": 0.398,
        "mean": 6.027,
        "max": 92.041,
        "pstdev": 11.76,
        "percentiles": {
          "p50": 2.652,
          "p75": 4.198,
          "p80": 4.747,
          "p90": 7.135,
          "p95": 47.513,
          "p99": 55.083,
          "p999": 61.542
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 129.08,
            "max": 129.08,
            "mean": 129.08
          },
          "cpu": {
            "min": 3.05,
            "max": 3.05,
            "mean": 3.05
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 36.57,
            "max": 36.57,
            "mean": 36.57
          },
          "cpu": {
            "min": 92.01,
            "max": 92.01,
            "mean": 92.01
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37
    },
    {
      "testName": "scale-up-httproutes-300",
      "routes": 300,
      "routesPerHostname": 1,
      "phase": "scaling-up",
      "throughput": 5612.73,
      "totalRequests": 168382,
      "latency": {
        "min": 0.368,
        "mean": 6.361,
        "max": 108.085,
        "pstdev": 12.378,
        "percentiles": {
          "p50": 2.817,
          "p75": 4.383,
          "p80": 4.929,
          "p90": 7.356,
          "p95": 48.795,
          "p99": 56.586,
          "p999": 68.399
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 140.63,
            "max": 140.63,
            "mean": 140.63
          },
          "cpu": {
            "min": 15.35,
            "max": 15.35,
            "mean": 15.35
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 58.77,
            "max": 58.77,
            "mean": 58.77
          },
          "cpu": {
            "min": 125.44,
            "max": 125.44,
            "mean": 125.44
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
      "throughput": 5799.01,
      "totalRequests": 173974,
      "latency": {
        "min": 0.378,
        "mean": 6.166,
        "max": 97.771,
        "pstdev": 11.964,
        "percentiles": {
          "p50": 2.818,
          "p75": 4.251,
          "p80": 4.742,
          "p90": 6.922,
          "p95": 47.56,
          "p99": 55.281,
          "p999": 68.112
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 152.5,
            "max": 152.5,
            "mean": 152.5
          },
          "cpu": {
            "min": 28.38,
            "max": 28.38,
            "mean": 28.38
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 78.94,
            "max": 78.94,
            "mean": 78.94
          },
          "cpu": {
            "min": 159.67,
            "max": 159.67,
            "mean": 159.67
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
      "throughput": 5474.43,
      "totalRequests": 164233,
      "latency": {
        "min": 0.386,
        "mean": 6.534,
        "max": 109.809,
        "pstdev": 12.936,
        "percentiles": {
          "p50": 2.656,
          "p75": 4.361,
          "p80": 5.013,
          "p90": 8.074,
          "p95": 49.516,
          "p99": 60.213,
          "p999": 73.678
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 183.98,
            "max": 183.98,
            "mean": 183.98
          },
          "cpu": {
            "min": 62.29,
            "max": 62.29,
            "mean": 62.29
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 131.16,
            "max": 131.16,
            "mean": 131.16
          },
          "cpu": {
            "min": 200.46,
            "max": 200.46,
            "mean": 200.46
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
      "throughput": 5816.39,
      "totalRequests": 174492,
      "latency": {
        "min": 0.34,
        "mean": 6.131,
        "max": 85.209,
        "pstdev": 11.868,
        "percentiles": {
          "p50": 2.73,
          "p75": 4.264,
          "p80": 4.816,
          "p90": 7.369,
          "p95": 47.933,
          "p99": 54.984,
          "p999": 61.82
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 195,
            "max": 195,
            "mean": 195
          },
          "cpu": {
            "min": 116.53,
            "max": 116.53,
            "mean": 116.53
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 120.37,
            "max": 120.37,
            "mean": 120.37
          },
          "cpu": {
            "min": 363.61,
            "max": 363.61,
            "mean": 363.61
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
      "throughput": 5848.8,
      "totalRequests": 175470,
      "latency": {
        "min": 0.372,
        "mean": 6.285,
        "max": 77.979,
        "pstdev": 12.136,
        "percentiles": {
          "p50": 2.76,
          "p75": 4.334,
          "p80": 4.906,
          "p90": 7.492,
          "p95": 48.553,
          "p99": 55.732,
          "p999": 63.037
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 205.2,
            "max": 205.2,
            "mean": 205.2
          },
          "cpu": {
            "min": 115.88,
            "max": 115.88,
            "mean": 115.88
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 120.26,
            "max": 120.26,
            "mean": 120.26
          },
          "cpu": {
            "min": 333.06,
            "max": 333.06,
            "mean": 333.06
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
      "throughput": 5897.55,
      "totalRequests": 176927,
      "latency": {
        "min": 0.365,
        "mean": 6.054,
        "max": 78.041,
        "pstdev": 11.77,
        "percentiles": {
          "p50": 2.761,
          "p75": 4.253,
          "p80": 4.77,
          "p90": 6.934,
          "p95": 47.628,
          "p99": 54.935,
          "p999": 62.232
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 201.78,
            "max": 201.78,
            "mean": 201.78
          },
          "cpu": {
            "min": 114.84,
            "max": 114.84,
            "mean": 114.84
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 122.23,
            "max": 122.23,
            "mean": 122.23
          },
          "cpu": {
            "min": 302.66,
            "max": 302.66,
            "mean": 302.66
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
      "throughput": 5764.69,
      "totalRequests": 172941,
      "latency": {
        "min": 0.376,
        "mean": 6.54,
        "max": 87.822,
        "pstdev": 12.528,
        "percentiles": {
          "p50": 2.847,
          "p75": 4.511,
          "p80": 5.118,
          "p90": 7.907,
          "p95": 49.336,
          "p99": 56.737,
          "p999": 66.895
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 237.36,
            "max": 237.36,
            "mean": 237.36
          },
          "cpu": {
            "min": 104.3,
            "max": 104.3,
            "mean": 104.3
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 131.2,
            "max": 131.2,
            "mean": 131.2
          },
          "cpu": {
            "min": 270.32,
            "max": 270.32,
            "mean": 270.32
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
      "throughput": 5709.58,
      "totalRequests": 171290,
      "latency": {
        "min": 0.35,
        "mean": 6.172,
        "max": 93.663,
        "pstdev": 11.978,
        "percentiles": {
          "p50": 2.864,
          "p75": 4.285,
          "p80": 4.782,
          "p90": 6.98,
          "p95": 47.689,
          "p99": 55.115,
          "p999": 67.44
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 158.41,
            "max": 158.41,
            "mean": 158.41
          },
          "cpu": {
            "min": 92.69,
            "max": 92.69,
            "mean": 92.69
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 131.32,
            "max": 131.32,
            "mean": 131.32
          },
          "cpu": {
            "min": 236.85,
            "max": 236.85,
            "mean": 236.85
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37
    }
  ]
};

export default benchmarkData;
