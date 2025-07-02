import { TestSuite } from '../types';

// Benchmark data extracted from markdown report for version 1.3.1
// Generated on 2025-06-17T19:50:26.775Z

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.3.1",
    "runId": "1.3.1-1750189826775",
    "date": "2025-03-05",
    "environment": "GitHub CI",
    "description": "Benchmark report for Envoy Gateway 1.3.1",
    "downloadUrl": "https://github.com/envoyproxy/gateway/releases/download/v1.3.1/benchmark_report.zip",
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
      "throughput": 5115.35,
      "totalRequests": 153465,
      "latency": {
        "min": 389,
        "mean": 6604,
        "max": 74543,
        "pstdev": 11843,
        "percentiles": {
          "p50": 3094,
          "p75": 5007,
          "p80": 5671,
          "p90": 8723,
          "p95": 46997,
          "p99": 55033,
          "p999": 60336
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 122.23,
            "max": 122.23,
            "mean": 122.23
          },
          "cpu": {
            "min": 1.04,
            "max": 1.04,
            "mean": 1.04
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 25.28,
            "max": 25.28,
            "mean": 25.28
          },
          "cpu": {
            "min": 30.51,
            "max": 30.51,
            "mean": 30.51
          }
        }
      },
      "poolOverflow": 364,
      "upstreamConnections": 36
    },
    {
      "testName": "scale-up-httproutes-50",
      "routes": 50,
      "routesPerHostname": 1,
      "phase": "scaling-up",
      "throughput": 5150.43,
      "totalRequests": 154513,
      "latency": {
        "min": 386,
        "mean": 6919,
        "max": 87556,
        "pstdev": 12645,
        "percentiles": {
          "p50": 3126,
          "p75": 5009,
          "p80": 5646,
          "p90": 8620,
          "p95": 49580,
          "p99": 55326,
          "p999": 60073
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 151.59,
            "max": 151.59,
            "mean": 151.59
          },
          "cpu": {
            "min": 1.83,
            "max": 1.83,
            "mean": 1.83
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 31.45,
            "max": 31.45,
            "mean": 31.45
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
      "throughput": 5211.7,
      "totalRequests": 156354,
      "latency": {
        "min": 376,
        "mean": 6687,
        "max": 88604,
        "pstdev": 12657,
        "percentiles": {
          "p50": 2913,
          "p75": 4676,
          "p80": 5309,
          "p90": 8259,
          "p95": 49930,
          "p99": 55853,
          "p999": 63514
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 145.61,
            "max": 145.61,
            "mean": 145.61
          },
          "cpu": {
            "min": 3.19,
            "max": 3.19,
            "mean": 3.19
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 35.5,
            "max": 35.5,
            "mean": 35.5
          },
          "cpu": {
            "min": 91.72,
            "max": 91.72,
            "mean": 91.72
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
      "throughput": 5066.99,
      "totalRequests": 152016,
      "latency": {
        "min": 386,
        "mean": 7628,
        "max": 93048,
        "pstdev": 13700,
        "percentiles": {
          "p50": 3310,
          "p75": 5414,
          "p80": 6156,
          "p90": 10289,
          "p95": 51750,
          "p99": 58007,
          "p999": 66938
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 156.12,
            "max": 156.12,
            "mean": 156.12
          },
          "cpu": {
            "min": 15.73,
            "max": 15.73,
            "mean": 15.73
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 55.56,
            "max": 55.56,
            "mean": 55.56
          },
          "cpu": {
            "min": 124.54,
            "max": 124.54,
            "mean": 124.54
          }
        }
      },
      "poolOverflow": 359,
      "upstreamConnections": 41
    },
    {
      "testName": "scale-up-httproutes-500",
      "routes": 500,
      "routesPerHostname": 1,
      "phase": "scaling-up",
      "throughput": 5098.53,
      "totalRequests": 152960,
      "latency": {
        "min": 373,
        "mean": 7203,
        "max": 108142,
        "pstdev": 13209,
        "percentiles": {
          "p50": 3161,
          "p75": 5083,
          "p80": 5791,
          "p90": 9292,
          "p95": 50372,
          "p99": 56864,
          "p999": 67850
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 181.23,
            "max": 181.23,
            "mean": 181.23
          },
          "cpu": {
            "min": 28.35,
            "max": 28.35,
            "mean": 28.35
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 73.6,
            "max": 73.6,
            "mean": 73.6
          },
          "cpu": {
            "min": 158.01,
            "max": 158.01,
            "mean": 158.01
          }
        }
      },
      "poolOverflow": 361,
      "upstreamConnections": 39
    },
    {
      "testName": "scale-up-httproutes-1000",
      "routes": 1000,
      "routesPerHostname": 1,
      "phase": "scaling-up",
      "throughput": 5048.96,
      "totalRequests": 151469,
      "latency": {
        "min": 366,
        "mean": 7462,
        "max": 99799,
        "pstdev": 13676,
        "percentiles": {
          "p50": 3122,
          "p75": 5245,
          "p80": 6050,
          "p90": 10398,
          "p95": 50685,
          "p99": 59639,
          "p999": 74547
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 196.57,
            "max": 196.57,
            "mean": 196.57
          },
          "cpu": {
            "min": 61.62,
            "max": 61.62,
            "mean": 61.62
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 119.73,
            "max": 119.73,
            "mean": 119.73
          },
          "cpu": {
            "min": 196.43,
            "max": 196.43,
            "mean": 196.43
          }
        }
      },
      "poolOverflow": 360,
      "upstreamConnections": 40
    },
    {
      "testName": "scale-down-httproutes-10",
      "routes": 10,
      "routesPerHostname": 1,
      "phase": "scaling-down",
      "throughput": 5116.49,
      "totalRequests": 153495,
      "latency": {
        "min": 373,
        "mean": 7373,
        "max": 79749,
        "pstdev": 13231,
        "percentiles": {
          "p50": 3334,
          "p75": 5268,
          "p80": 5935,
          "p90": 9436,
          "p95": 50972,
          "p99": 56680,
          "p999": 62969
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 213.36,
            "max": 213.36,
            "mean": 213.36
          },
          "cpu": {
            "min": 114.56,
            "max": 114.56,
            "mean": 114.56
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 109.79,
            "max": 109.79,
            "mean": 109.79
          },
          "cpu": {
            "min": 357.84,
            "max": 357.84,
            "mean": 357.84
          }
        }
      },
      "poolOverflow": 360,
      "upstreamConnections": 40
    },
    {
      "testName": "scale-down-httproutes-50",
      "routes": 50,
      "routesPerHostname": 1,
      "phase": "scaling-down",
      "throughput": 5128.63,
      "totalRequests": 153859,
      "latency": {
        "min": 380,
        "mean": 6787,
        "max": 76259,
        "pstdev": 12813,
        "percentiles": {
          "p50": 2982,
          "p75": 4725,
          "p80": 5330,
          "p90": 8194,
          "p95": 50354,
          "p99": 56367,
          "p999": 61741
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 217.34,
            "max": 217.34,
            "mean": 217.34
          },
          "cpu": {
            "min": 113.9,
            "max": 113.9,
            "mean": 113.9
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 119.73,
            "max": 119.73,
            "mean": 119.73
          },
          "cpu": {
            "min": 327.37,
            "max": 327.37,
            "mean": 327.37
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
      "throughput": 5217.06,
      "totalRequests": 156516,
      "latency": {
        "min": 386,
        "mean": 7182,
        "max": 93491,
        "pstdev": 12999,
        "percentiles": {
          "p50": 3211,
          "p75": 5188,
          "p80": 5872,
          "p90": 9449,
          "p95": 50083,
          "p99": 56135,
          "p999": 63309
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 220.2,
            "max": 220.2,
            "mean": 220.2
          },
          "cpu": {
            "min": 112.64,
            "max": 112.64,
            "mean": 112.64
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 119.73,
            "max": 119.73,
            "mean": 119.73
          },
          "cpu": {
            "min": 296.73,
            "max": 296.73,
            "mean": 296.73
          }
        }
      },
      "poolOverflow": 360,
      "upstreamConnections": 40
    },
    {
      "testName": "scale-down-httproutes-300",
      "routes": 300,
      "routesPerHostname": 1,
      "phase": "scaling-down",
      "throughput": 5033.4,
      "totalRequests": 151002,
      "latency": {
        "min": 380,
        "mean": 6522,
        "max": 92626,
        "pstdev": 12472,
        "percentiles": {
          "p50": 2849,
          "p75": 4643,
          "p80": 5272,
          "p90": 8064,
          "p95": 49350,
          "p99": 55799,
          "p999": 64323
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 201.77,
            "max": 201.77,
            "mean": 201.77
          },
          "cpu": {
            "min": 102.25,
            "max": 102.25,
            "mean": 102.25
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 119.73,
            "max": 119.73,
            "mean": 119.73
          },
          "cpu": {
            "min": 264.74,
            "max": 264.74,
            "mean": 264.74
          }
        }
      },
      "poolOverflow": 365,
      "upstreamConnections": 35
    },
    {
      "testName": "scale-down-httproutes-500",
      "routes": 500,
      "routesPerHostname": 1,
      "phase": "scaling-down",
      "throughput": 5061.34,
      "totalRequests": 151843,
      "latency": {
        "min": 393,
        "mean": 7429,
        "max": 99934,
        "pstdev": 13552,
        "percentiles": {
          "p50": 3227,
          "p75": 5180,
          "p80": 5867,
          "p90": 9747,
          "p95": 51320,
          "p99": 58009,
          "p999": 68829
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 173.04,
            "max": 173.04,
            "mean": 173.04
          },
          "cpu": {
            "min": 91.12,
            "max": 91.12,
            "mean": 91.12
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 119.73,
            "max": 119.73,
            "mean": 119.73
          },
          "cpu": {
            "min": 231.98,
            "max": 231.98,
            "mean": 231.98
          }
        }
      },
      "poolOverflow": 360,
      "upstreamConnections": 40
    }
  ]
};

export default benchmarkData;
