import { TestSuite } from '../types';

// Benchmark data extracted from markdown report for version 1.1.0
// Generated on 2025-06-17T19:58:49.977Z

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.1.0",
    "runId": "1.1.0-1750190329977",
    "date": "2024-07-23",
    "environment": "GitHub CI",
    "description": "Benchmark report for Envoy Gateway 1.1.0",
    "downloadUrl": "https://github.com/envoyproxy/gateway/releases/download/v1.1.0/benchmark_report.zip",
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
      "throughput": 6181.97,
      "totalRequests": 185463,
      "latency": {
        "min": 362,
        "mean": 5902,
        "max": 73084,
        "pstdev": 11039,
        "percentiles": {
          "p50": 2765,
          "p75": 4364,
          "p80": 4935,
          "p90": 7504,
          "p95": 41244,
          "p99": 53929,
          "p999": 61147
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 86.2,
            "max": 86.2,
            "mean": 86.2
          },
          "cpu": {
            "min": 1.4333333333333333,
            "max": 1.4333333333333333,
            "mean": 1.4333333333333333
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 24.09,
            "max": 24.09,
            "mean": 24.09
          },
          "cpu": {
            "min": 101.36666666666667,
            "max": 101.36666666666667,
            "mean": 101.36666666666667
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
      "throughput": 6103.97,
      "totalRequests": 183121,
      "latency": {
        "min": 354,
        "mean": 5852,
        "max": 75943,
        "pstdev": 11382,
        "percentiles": {
          "p50": 2681,
          "p75": 4083,
          "p80": 4601,
          "p90": 6825,
          "p95": 46262,
          "p99": 53862,
          "p999": 60651
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 100.14,
            "max": 100.14,
            "mean": 100.14
          },
          "cpu": {
            "min": 7.233333333333333,
            "max": 7.233333333333333,
            "mean": 7.233333333333333
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 32.25,
            "max": 32.25,
            "mean": 32.25
          },
          "cpu": {
            "min": 203.6,
            "max": 203.6,
            "mean": 203.6
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
      "throughput": 6074.87,
      "totalRequests": 182247,
      "latency": {
        "min": 371,
        "mean": 5877,
        "max": 94171,
        "pstdev": 11399,
        "percentiles": {
          "p50": 2732,
          "p75": 4191,
          "p80": 4677,
          "p90": 6801,
          "p95": 46481,
          "p99": 53866,
          "p999": 61036
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 114.69,
            "max": 114.69,
            "mean": 114.69
          },
          "cpu": {
            "min": 30,
            "max": 30,
            "mean": 30
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 46.42,
            "max": 46.42,
            "mean": 46.42
          },
          "cpu": {
            "min": 308.70000000000005,
            "max": 308.70000000000005,
            "mean": 308.70000000000005
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
      "throughput": 5993.57,
      "totalRequests": 179811,
      "latency": {
        "min": 368,
        "mean": 6123,
        "max": 96571,
        "pstdev": 11831,
        "percentiles": {
          "p50": 2753,
          "p75": 4261,
          "p80": 4799,
          "p90": 7247,
          "p95": 47368,
          "p99": 55248,
          "p999": 66586
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 762.8,
            "max": 762.8,
            "mean": 762.8
          },
          "cpu": {
            "min": 600.8666666666667,
            "max": 600.8666666666667,
            "mean": 600.8666666666667
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 152.78,
            "max": 152.78,
            "mean": 152.78
          },
          "cpu": {
            "min": 488.70000000000005,
            "max": 488.70000000000005,
            "mean": 488.70000000000005
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
      "throughput": 5812.8,
      "totalRequests": 174396,
      "latency": {
        "min": 388,
        "mean": 6310,
        "max": 95719,
        "pstdev": 12246,
        "percentiles": {
          "p50": 2729,
          "p75": 4334,
          "p80": 4905,
          "p90": 7607,
          "p95": 47970,
          "p99": 56811,
          "p999": 68165
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 1593.56,
            "max": 1593.56,
            "mean": 1593.56
          },
          "cpu": {
            "min": 36.56666666666667,
            "max": 36.56666666666667,
            "mean": 36.56666666666667
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 289.79,
            "max": 289.79,
            "mean": 289.79
          },
          "cpu": {
            "min": 715.6333333333333,
            "max": 715.6333333333333,
            "mean": 715.6333333333333
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
      "throughput": 5916.93,
      "totalRequests": 177517,
      "latency": {
        "min": 385,
        "mean": 6352,
        "max": 82747,
        "pstdev": 12011,
        "percentiles": {
          "p50": 2822,
          "p75": 4461,
          "p80": 5061,
          "p90": 7769,
          "p95": 47847,
          "p99": 55068,
          "p999": 62578
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 124.07,
            "max": 124.07,
            "mean": 124.07
          },
          "cpu": {
            "min": 519.6333333333332,
            "max": 519.6333333333332,
            "mean": 519.6333333333332
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 284.03,
            "max": 284.03,
            "mean": 284.03
          },
          "cpu": {
            "min": 1194.8666666666666,
            "max": 1194.8666666666666,
            "mean": 1194.8666666666666
          }
        }
      },
      "poolOverflow": 361,
      "upstreamConnections": 39
    },
    {
      "testName": "scale-down-httproutes-50",
      "routes": 50,
      "routesPerHostname": 1,
      "phase": "scaling-down",
      "throughput": 5876.35,
      "totalRequests": 176294,
      "latency": {
        "min": 371,
        "mean": 6226,
        "max": 83869,
        "pstdev": 11816,
        "percentiles": {
          "p50": 2770,
          "p75": 4396,
          "p80": 5000,
          "p90": 7848,
          "p95": 47163,
          "p99": 55156,
          "p999": 64434
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 131.62,
            "max": 131.62,
            "mean": 131.62
          },
          "cpu": {
            "min": 513.6,
            "max": 513.6,
            "mean": 513.6
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 291.73,
            "max": 291.73,
            "mean": 291.73
          },
          "cpu": {
            "min": 1092.9333333333334,
            "max": 1092.9333333333334,
            "mean": 1092.9333333333334
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
      "throughput": 4820.43,
      "totalRequests": 144617,
      "latency": {
        "min": 342,
        "mean": 7088,
        "max": 172031,
        "pstdev": 12642,
        "percentiles": {
          "p50": 3156,
          "p75": 5480,
          "p80": 6443,
          "p90": 12007,
          "p95": 39675,
          "p99": 63842,
          "p999": 86532
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 374.36,
            "max": 374.36,
            "mean": 374.36
          },
          "cpu": {
            "min": 462.8,
            "max": 462.8,
            "mean": 462.8
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 289.98,
            "max": 289.98,
            "mean": 289.98
          },
          "cpu": {
            "min": 980.4,
            "max": 980.4,
            "mean": 980.4
          }
        }
      },
      "poolOverflow": 364,
      "upstreamConnections": 36
    },
    {
      "testName": "scale-down-httproutes-300",
      "routes": 300,
      "routesPerHostname": 1,
      "phase": "scaling-down",
      "throughput": 5746.71,
      "totalRequests": 172402,
      "latency": {
        "min": 386,
        "mean": 6313,
        "max": 101535,
        "pstdev": 11578,
        "percentiles": {
          "p50": 2875,
          "p75": 4579,
          "p80": 5222,
          "p90": 8617,
          "p95": 44341,
          "p99": 54996,
          "p999": 67178
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 670.44,
            "max": 670.44,
            "mean": 670.44
          },
          "cpu": {
            "min": 24.333333333333332,
            "max": 24.333333333333332,
            "mean": 24.333333333333332
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 285.86,
            "max": 285.86,
            "mean": 285.86
          },
          "cpu": {
            "min": 824.5666666666667,
            "max": 824.5666666666667,
            "mean": 824.5666666666667
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38
    }
  ]
};

export default benchmarkData;
