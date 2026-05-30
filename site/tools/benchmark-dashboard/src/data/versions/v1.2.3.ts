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
        "min": 0.349,
        "mean": 6.177,
        "max": 70.643,
        "pstdev": 11.464,
        "percentiles": {
          "p50": 2.879,
          "p75": 4.52,
          "p80": 5.082,
          "p90": 7.684,
          "p95": 46.213,
          "p99": 54.147,
          "p999": 60.25
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
        "min": 0.37,
        "mean": 5.77,
        "max": 72.798,
        "pstdev": 11.358,
        "percentiles": {
          "p50": 2.657,
          "p75": 4.054,
          "p80": 4.535,
          "p90": 6.537,
          "p95": 46.632,
          "p99": 53.751,
          "p999": 59.174
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
        "min": 0.365,
        "mean": 5.863,
        "max": 84.721,
        "pstdev": 11.54,
        "percentiles": {
          "p50": 2.605,
          "p75": 4.127,
          "p80": 4.662,
          "p90": 7.048,
          "p95": 46.759,
          "p99": 54.556,
          "p999": 61.736
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
        "min": 0.362,
        "mean": 6.079,
        "max": 79.421,
        "pstdev": 11.806,
        "percentiles": {
          "p50": 2.707,
          "p75": 4.272,
          "p80": 4.832,
          "p90": 7.327,
          "p95": 47.441,
          "p99": 54.743,
          "p999": 65.808
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
        "min": 0.344,
        "mean": 6.36,
        "max": 79.355,
        "pstdev": 12.473,
        "percentiles": {
          "p50": 2.707,
          "p75": 4.299,
          "p80": 4.888,
          "p90": 7.643,
          "p95": 49.309,
          "p99": 57.382,
          "p999": 68.141
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
        "min": 0.361,
        "mean": 6.194,
        "max": 89.714,
        "pstdev": 12.241,
        "percentiles": {
          "p50": 2.703,
          "p75": 4.241,
          "p80": 4.791,
          "p90": 7.296,
          "p95": 47.755,
          "p99": 58.537,
          "p999": 70.676
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
        "min": 0.351,
        "mean": 5.824,
        "max": 73.633,
        "pstdev": 11.47,
        "percentiles": {
          "p50": 2.578,
          "p75": 4.038,
          "p80": 4.57,
          "p90": 6.823,
          "p95": 46.927,
          "p99": 54.392,
          "p999": 60.213
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
        "min": 0.378,
        "mean": 5.877,
        "max": 82.526,
        "pstdev": 11.643,
        "percentiles": {
          "p50": 2.588,
          "p75": 4.048,
          "p80": 4.587,
          "p90": 6.876,
          "p95": 47.429,
          "p99": 54.71,
          "p999": 62.111
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
        "min": 0.388,
        "mean": 6.024,
        "max": 80.216,
        "pstdev": 11.596,
        "percentiles": {
          "p50": 2.802,
          "p75": 4.253,
          "p80": 4.761,
          "p90": 6.951,
          "p95": 47.183,
          "p99": 54.065,
          "p999": 59.455
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
        "min": 0.379,
        "mean": 6.153,
        "max": 85.528,
        "pstdev": 12.032,
        "percentiles": {
          "p50": 2.64,
          "p75": 4.362,
          "p80": 4.961,
          "p90": 7.625,
          "p95": 47.886,
          "p99": 55.965,
          "p999": 65.695
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
        "min": 0.362,
        "mean": 5.967,
        "max": 83.48,
        "pstdev": 11.671,
        "percentiles": {
          "p50": 2.714,
          "p75": 4.156,
          "p80": 4.658,
          "p90": 6.839,
          "p95": 46.901,
          "p99": 54.601,
          "p999": 65.464
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
