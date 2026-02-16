import { TestSuite } from '../types';

// Benchmark data extracted from markdown report for version 1.2.3
// Generated on 2025-06-17T19:50:26.763Z

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.2.3",
    "runId": "1.2.3-1750189826763",
    "date": "2024-12-02",
    "environment": "GitHub CI",
    "description": "Benchmark report for Envoy Gateway 1.2.3",
    "downloadUrl": "https://github.com/envoyproxy/gateway/releases/download/v1.2.3/benchmark_report.zip",
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
      "throughput": 6222.52,
      "totalRequests": 186680,
      "latency": {
        "min": 349,
        "mean": 6177,
        "max": 70643,
        "pstdev": 11464,
        "percentiles": {
          "p50": 2879,
          "p75": 4520,
          "p80": 5082,
          "p90": 7684,
          "p95": 46213,
          "p99": 54147,
          "p999": 60250
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 110.2,
            "max": 110.2,
            "mean": 110.2
          },
          "cpu": {
            "min": 0.68,
            "max": 0.68,
            "mean": 0.68
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 25.89,
            "max": 25.89,
            "mean": 25.89
          },
          "cpu": {
            "min": 30.33,
            "max": 30.33,
            "mean": 30.33
          }
        }
      },
      "poolOverflow": 360,
      "upstreamConnections": 40
    },
    {
      "testName": "scale-up-httproutes-50",
      "routes": 50,
      "routesPerHostname": 1,
      "phase": "scaling-up",
      "throughput": 6170.38,
      "totalRequests": 185112,
      "latency": {
        "min": 370,
        "mean": 5770,
        "max": 72798,
        "pstdev": 11358,
        "percentiles": {
          "p50": 2657,
          "p75": 4054,
          "p80": 4535,
          "p90": 6537,
          "p95": 46632,
          "p99": 53751,
          "p999": 59174
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 122.48,
            "max": 122.48,
            "mean": 122.48
          },
          "cpu": {
            "min": 1.42,
            "max": 1.42,
            "mean": 1.42
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 32.05,
            "max": 32.05,
            "mean": 32.05
          },
          "cpu": {
            "min": 60.88,
            "max": 60.88,
            "mean": 60.88
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37
    },
    {
      "testName": "scale-up-httproutes-100",
      "routes": 100,
      "routesPerHostname": 1,
      "phase": "scaling-up",
      "throughput": 6085.51,
      "totalRequests": 182567,
      "latency": {
        "min": 365,
        "mean": 5863,
        "max": 84721,
        "pstdev": 11540,
        "percentiles": {
          "p50": 2605,
          "p75": 4127,
          "p80": 4662,
          "p90": 7048,
          "p95": 46759,
          "p99": 54556,
          "p999": 61736
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 131.06,
            "max": 131.06,
            "mean": 131.06
          },
          "cpu": {
            "min": 2.8,
            "max": 2.8,
            "mean": 2.8
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 38.19,
            "max": 38.19,
            "mean": 38.19
          },
          "cpu": {
            "min": 91.64,
            "max": 91.64,
            "mean": 91.64
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
      "throughput": 6032.83,
      "totalRequests": 180985,
      "latency": {
        "min": 362,
        "mean": 6079,
        "max": 79421,
        "pstdev": 11806,
        "percentiles": {
          "p50": 2707,
          "p75": 4272,
          "p80": 4832,
          "p90": 7327,
          "p95": 47441,
          "p99": 54743,
          "p999": 65808
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 142.98,
            "max": 142.98,
            "mean": 142.98
          },
          "cpu": {
            "min": 15.04,
            "max": 15.04,
            "mean": 15.04
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 60.39,
            "max": 60.39,
            "mean": 60.39
          },
          "cpu": {
            "min": 124.47,
            "max": 124.47,
            "mean": 124.47
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
      "throughput": 5773.98,
      "totalRequests": 173223,
      "latency": {
        "min": 344,
        "mean": 6360,
        "max": 79355,
        "pstdev": 12473,
        "percentiles": {
          "p50": 2707,
          "p75": 4299,
          "p80": 4888,
          "p90": 7643,
          "p95": 49309,
          "p99": 57382,
          "p999": 68141
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 160.84,
            "max": 160.84,
            "mean": 160.84
          },
          "cpu": {
            "min": 27.66,
            "max": 27.66,
            "mean": 27.66
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 80.58,
            "max": 80.58,
            "mean": 80.58
          },
          "cpu": {
            "min": 156.97,
            "max": 156.97,
            "mean": 156.97
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
      "throughput": 5776.08,
      "totalRequests": 173283,
      "latency": {
        "min": 361,
        "mean": 6194,
        "max": 89714,
        "pstdev": 12241,
        "percentiles": {
          "p50": 2703,
          "p75": 4241,
          "p80": 4791,
          "p90": 7296,
          "p95": 47755,
          "p99": 58537,
          "p999": 70676
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 186.49,
            "max": 186.49,
            "mean": 186.49
          },
          "cpu": {
            "min": 61.39,
            "max": 61.39,
            "mean": 61.39
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 130.82,
            "max": 130.82,
            "mean": 130.82
          },
          "cpu": {
            "min": 196.17,
            "max": 196.17,
            "mean": 196.17
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
      "throughput": 6128.86,
      "totalRequests": 183866,
      "latency": {
        "min": 351,
        "mean": 5824,
        "max": 73633,
        "pstdev": 11470,
        "percentiles": {
          "p50": 2578,
          "p75": 4038,
          "p80": 4570,
          "p90": 6823,
          "p95": 46927,
          "p99": 54392,
          "p999": 60213
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 218.36,
            "max": 218.36,
            "mean": 218.36
          },
          "cpu": {
            "min": 114.4,
            "max": 114.4,
            "mean": 114.4
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 118.8,
            "max": 118.8,
            "mean": 118.8
          },
          "cpu": {
            "min": 358.03,
            "max": 358.03,
            "mean": 358.03
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
      "throughput": 6065.62,
      "totalRequests": 181969,
      "latency": {
        "min": 378,
        "mean": 5877,
        "max": 82526,
        "pstdev": 11643,
        "percentiles": {
          "p50": 2588,
          "p75": 4048,
          "p80": 4587,
          "p90": 6876,
          "p95": 47429,
          "p99": 54710,
          "p999": 62111
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 222.42,
            "max": 222.42,
            "mean": 222.42
          },
          "cpu": {
            "min": 113.86,
            "max": 113.86,
            "mean": 113.86
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 119.23,
            "max": 119.23,
            "mean": 119.23
          },
          "cpu": {
            "min": 327.6,
            "max": 327.6,
            "mean": 327.6
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
      "throughput": 6045.71,
      "totalRequests": 181360,
      "latency": {
        "min": 388,
        "mean": 6024,
        "max": 80216,
        "pstdev": 11596,
        "percentiles": {
          "p50": 2802,
          "p75": 4253,
          "p80": 4761,
          "p90": 6951,
          "p95": 47183,
          "p99": 54065,
          "p999": 59455
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 142.37,
            "max": 142.37,
            "mean": 142.37
          },
          "cpu": {
            "min": 112.77,
            "max": 112.77,
            "mean": 112.77
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 119.54,
            "max": 119.54,
            "mean": 119.54
          },
          "cpu": {
            "min": 296.95,
            "max": 296.95,
            "mean": 296.95
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
      "throughput": 5956.66,
      "totalRequests": 178700,
      "latency": {
        "min": 379,
        "mean": 6153,
        "max": 85528,
        "pstdev": 12032,
        "percentiles": {
          "p50": 2640,
          "p75": 4362,
          "p80": 4961,
          "p90": 7625,
          "p95": 47886,
          "p99": 55965,
          "p999": 65695
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 148.72,
            "max": 148.72,
            "mean": 148.72
          },
          "cpu": {
            "min": 102.49,
            "max": 102.49,
            "mean": 102.49
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 130.98,
            "max": 130.98,
            "mean": 130.98
          },
          "cpu": {
            "min": 264.96,
            "max": 264.96,
            "mean": 264.96
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38
    },
    {
      "testName": "scale-down-httproutes-500",
      "routes": 500,
      "routesPerHostname": 1,
      "phase": "scaling-down",
      "throughput": 5958.61,
      "totalRequests": 178765,
      "latency": {
        "min": 362,
        "mean": 5967,
        "max": 83480,
        "pstdev": 11671,
        "percentiles": {
          "p50": 2714,
          "p75": 4156,
          "p80": 4658,
          "p90": 6839,
          "p95": 46901,
          "p99": 54601,
          "p999": 65464
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 164.87,
            "max": 164.87,
            "mean": 164.87
          },
          "cpu": {
            "min": 91.25,
            "max": 91.25,
            "mean": 91.25
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 130.98,
            "max": 130.98,
            "mean": 130.98
          },
          "cpu": {
            "min": 231.98,
            "max": 231.98,
            "mean": 231.98
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37
    }
  ]
};

export default benchmarkData;
