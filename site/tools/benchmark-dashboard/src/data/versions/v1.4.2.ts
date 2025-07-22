import { TestSuite } from '../types';

// Benchmark data extracted from markdown report for version 1.4.2
// Generated on 2025-07-18T02:23:12.475Z

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.4.2",
    "runId": "1.4.2-1752805392474",
    "date": "2025-07-18T02:23:12.474Z",
    "environment": "production",
    "description": "Benchmark results for version 1.4.2",
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
      "throughput": 5153.92,
      "totalRequests": 154618,
      "latency": {
        "min": 383,
        "mean": 7923,
        "max": 80375,
        "pstdev": 13254,
        "percentiles": {
          "p50": 3722,
          "p75": 6072,
          "p80": 6915,
          "p90": 11685,
          "p95": 49872,
          "p99": 56281,
          "p999": 62722
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 131.64,
            "max": 158.27,
            "mean": 153.65
          },
          "cpu": {
            "min": 0.13,
            "max": 1,
            "mean": 0.43
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 0,
            "max": 27.04,
            "mean": 24.5
          },
          "cpu": {
            "min": 0,
            "max": 92.27,
            "mean": 13.11
          }
        }
      },
      "poolOverflow": 357,
      "upstreamConnections": 43
    },
    {
      "testName": "scaling up httproutes to 50 with 10 routes per hostname",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-up",
      "throughput": 5214.96,
      "totalRequests": 156456,
      "latency": {
        "min": 392,
        "mean": 6917,
        "max": 94703,
        "pstdev": 12289,
        "percentiles": {
          "p50": 3265,
          "p75": 5168,
          "p80": 5850,
          "p90": 9014,
          "p95": 48109,
          "p99": 54501,
          "p999": 61138
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 151.75,
            "max": 162.37,
            "mean": 156.16
          },
          "cpu": {
            "min": 0.27,
            "max": 4.07,
            "mean": 0.93
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 26.85,
            "max": 33.23,
            "mean": 31.93
          },
          "cpu": {
            "min": 0,
            "max": 99.8,
            "mean": 2.75
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38
    },
    {
      "testName": "scaling up httproutes to 100 with 20 routes per hostname",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-up",
      "throughput": 5286.96,
      "totalRequests": 158609,
      "latency": {
        "min": 368,
        "mean": 7734,
        "max": 92332,
        "pstdev": 12933,
        "percentiles": {
          "p50": 3787,
          "p75": 5897,
          "p80": 6637,
          "p90": 10518,
          "p95": 49209,
          "p99": 55494,
          "p999": 62668
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 154.29,
            "max": 162.45,
            "mean": 160.12
          },
          "cpu": {
            "min": 0.4,
            "max": 8.13,
            "mean": 1.15
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 33.08,
            "max": 37.43,
            "mean": 36.96
          },
          "cpu": {
            "min": 0,
            "max": 99.94,
            "mean": 12.84
          }
        }
      },
      "poolOverflow": 357,
      "upstreamConnections": 43
    },
    {
      "testName": "scaling up httproutes to 300 with 60 routes per hostname",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-up",
      "throughput": 5071.92,
      "totalRequests": 152158,
      "latency": {
        "min": 367,
        "mean": 7108,
        "max": 96673,
        "pstdev": 12611,
        "percentiles": {
          "p50": 3259,
          "p75": 5273,
          "p80": 5983,
          "p90": 9540,
          "p95": 48895,
          "p99": 55816,
          "p999": 64583
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 178.72,
            "max": 189.43,
            "mean": 185.02
          },
          "cpu": {
            "min": 0.53,
            "max": 49.47,
            "mean": 4.51
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 53.3,
            "max": 57.54,
            "mean": 57.04
          },
          "cpu": {
            "min": 0,
            "max": 100.16,
            "mean": 15
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38
    },
    {
      "testName": "scaling up httproutes to 500 with 100 routes per hostname",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-up",
      "throughput": 4987.92,
      "totalRequests": 149640,
      "latency": {
        "min": 394,
        "mean": 7071,
        "max": 108810,
        "pstdev": 12666,
        "percentiles": {
          "p50": 3283,
          "p75": 5167,
          "p80": 5821,
          "p90": 9217,
          "p95": 48912,
          "p99": 56031,
          "p999": 68722
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 187.59,
            "max": 203.43,
            "mean": 199.65
          },
          "cpu": {
            "min": 0.27,
            "max": 48.25,
            "mean": 1.88
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 73.41,
            "max": 79.6,
            "mean": 79.35
          },
          "cpu": {
            "min": 0,
            "max": 99.91,
            "mean": 11.81
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37
    },
    {
      "testName": "scaling up httproutes to 1000 with 200 routes per hostname",
      "routes": 1000,
      "routesPerHostname": 200,
      "phase": "scaling-up",
      "throughput": 4945.69,
      "totalRequests": 148373,
      "latency": {
        "min": 411,
        "mean": 7306,
        "max": 101023,
        "pstdev": 12990,
        "percentiles": {
          "p50": 3249,
          "p75": 5440,
          "p80": 6223,
          "p90": 10246,
          "p95": 49207,
          "p99": 57444,
          "p999": 69570
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 231.85,
            "max": 243.1,
            "mean": 237.75
          },
          "cpu": {
            "min": 0.2,
            "max": 1.2,
            "mean": 0.84
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 129.84,
            "max": 130.36,
            "mean": 129.96
          },
          "cpu": {
            "min": 0,
            "max": 99.89,
            "mean": 11.78
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
      "throughput": 5200.02,
      "totalRequests": 156013,
      "latency": {
        "min": 397,
        "mean": 7832,
        "max": 75223,
        "pstdev": 13154,
        "percentiles": {
          "p50": 3651,
          "p75": 5953,
          "p80": 6758,
          "p90": 11656,
          "p95": 49915,
          "p99": 56150,
          "p999": 62173
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 262.52,
            "max": 266.5,
            "mean": 263.99
          },
          "cpu": {
            "min": 0.67,
            "max": 3.53,
            "mean": 1.15
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 123.49,
            "max": 123.65,
            "mean": 123.54
          },
          "cpu": {
            "min": 0,
            "max": 100.22,
            "mean": 7.55
          }
        }
      },
      "poolOverflow": 357,
      "upstreamConnections": 43
    },
    {
      "testName": "scaling down httproutes to 50 with 10 routes per hostname",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-down",
      "throughput": 5116.68,
      "totalRequests": 153503,
      "latency": {
        "min": 404,
        "mean": 7069,
        "max": 71688,
        "pstdev": 12498,
        "percentiles": {
          "p50": 3313,
          "p75": 5203,
          "p80": 5888,
          "p90": 9173,
          "p95": 48934,
          "p99": 55580,
          "p999": 61552
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 266.25,
            "max": 277.75,
            "mean": 267.54
          },
          "cpu": {
            "min": 0.67,
            "max": 7.13,
            "mean": 1.51
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 123.47,
            "max": 123.67,
            "mean": 123.53
          },
          "cpu": {
            "min": 0,
            "max": 99.78,
            "mean": 4.47
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38
    },
    {
      "testName": "scaling down httproutes to 100 with 20 routes per hostname",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-down",
      "throughput": 5153.5,
      "totalRequests": 154605,
      "latency": {
        "min": 376,
        "mean": 6989,
        "max": 80056,
        "pstdev": 12394,
        "percentiles": {
          "p50": 3284,
          "p75": 5189,
          "p80": 5871,
          "p90": 9047,
          "p95": 48644,
          "p99": 55019,
          "p999": 61106
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 170.85,
            "max": 277.75,
            "mean": 210.44
          },
          "cpu": {
            "min": 0.67,
            "max": 68.27,
            "mean": 7.39
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 123.46,
            "max": 123.75,
            "mean": 123.55
          },
          "cpu": {
            "min": 0,
            "max": 100.02,
            "mean": 16.06
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
      "throughput": 5125.52,
      "totalRequests": 153766,
      "latency": {
        "min": 369,
        "mean": 6827,
        "max": 85323,
        "pstdev": 12196,
        "percentiles": {
          "p50": 3193,
          "p75": 5036,
          "p80": 5699,
          "p90": 8862,
          "p95": 47597,
          "p99": 54747,
          "p999": 64284
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 181.68,
            "max": 196.62,
            "mean": 192.08
          },
          "cpu": {
            "min": 0.4,
            "max": 72.6,
            "mean": 3.62
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 129.86,
            "max": 130.06,
            "mean": 129.95
          },
          "cpu": {
            "min": 0,
            "max": 100.07,
            "mean": 13.11
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
      "throughput": 5130.15,
      "totalRequests": 153916,
      "latency": {
        "min": 389,
        "mean": 7004,
        "max": 103731,
        "pstdev": 12505,
        "percentiles": {
          "p50": 3244,
          "p75": 5247,
          "p80": 5944,
          "p90": 9184,
          "p95": 48279,
          "p99": 55420,
          "p999": 66549
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 205.04,
            "max": 208.2,
            "mean": 206.96
          },
          "cpu": {
            "min": 0.33,
            "max": 54.27,
            "mean": 2.52
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 129.85,
            "max": 130.02,
            "mean": 129.9
          },
          "cpu": {
            "min": 0,
            "max": 91.79,
            "mean": 22.69
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38
    }
  ]
};

export default benchmarkData;
