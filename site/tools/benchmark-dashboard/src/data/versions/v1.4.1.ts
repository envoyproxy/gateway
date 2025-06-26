import { TestSuite } from '../types';

// Benchmark data extracted from markdown report for version 1.4.1
// Generated on 2025-06-17T19:50:26.783Z

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.4.1",
    "runId": "1.4.1-1750189826783",
    "date": "2025-06-04",
    "environment": "GitHub CI",
    "description": "Benchmark report for Envoy Gateway 1.4.1",
    "downloadUrl": "https://github.com/envoyproxy/gateway/releases/download/v1.4.1/benchmark_report.zip",
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
      "throughput": 5440.31,
      "totalRequests": 163209,
      "latency": {
        "min": 335,
        "mean": 6565,
        "max": 66668,
        "pstdev": 11480,
        "percentiles": {
          "p50": 3258,
          "p75": 5079,
          "p80": 5722,
          "p90": 8679,
          "p95": 45924,
          "p99": 53512,
          "p999": 58220
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 128.02,
            "max": 151.26,
            "mean": 147.41
          },
          "cpu": {
            "min": 0.27,
            "max": 0.67,
            "mean": 0.45
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 0,
            "max": 26.92,
            "mean": 22.58
          },
          "cpu": {
            "min": 0,
            "max": 99.73,
            "mean": 6.02
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38
    },
    {
      "testName": "scaling up httproutes to 50 with 10 routes per hostname",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-up",
      "throughput": 5429.88,
      "totalRequests": 162900,
      "latency": {
        "min": 345,
        "mean": 6468,
        "max": 70791,
        "pstdev": 11793,
        "percentiles": {
          "p50": 3170,
          "p75": 4794,
          "p80": 5311,
          "p90": 7534,
          "p95": 47400,
          "p99": 53536,
          "p999": 58034
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 151.26,
            "max": 170.78,
            "mean": 161.92
          },
          "cpu": {
            "min": 0.4,
            "max": 4.33,
            "mean": 0.97
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 26.73,
            "max": 33.13,
            "mean": 31.63
          },
          "cpu": {
            "min": 0,
            "max": 99.94,
            "mean": 3.35
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37
    },
    {
      "testName": "scaling up httproutes to 100 with 20 routes per hostname",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-up",
      "throughput": 5499.86,
      "totalRequests": 164996,
      "latency": {
        "min": 391,
        "mean": 8147,
        "max": 99663,
        "pstdev": 13319,
        "percentiles": {
          "p50": 3957,
          "p75": 6308,
          "p80": 7120,
          "p90": 11799,
          "p95": 49942,
          "p99": 56190,
          "p999": 62709
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 161.19,
            "max": 165.52,
            "mean": 163.96
          },
          "cpu": {
            "min": 0.4,
            "max": 8.67,
            "mean": 1.36
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 32.98,
            "max": 39.18,
            "mean": 38.14
          },
          "cpu": {
            "min": 0,
            "max": 99.97,
            "mean": 8.67
          }
        }
      },
      "poolOverflow": 353,
      "upstreamConnections": 47
    },
    {
      "testName": "scaling up httproutes to 300 with 60 routes per hostname",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-up",
      "throughput": 5335.1,
      "totalRequests": 160053,
      "latency": {
        "min": 365,
        "mean": 6734,
        "max": 90963,
        "pstdev": 12096,
        "percentiles": {
          "p50": 3228,
          "p75": 5042,
          "p80": 5649,
          "p90": 8465,
          "p95": 47704,
          "p99": 54396,
          "p999": 61714
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 183.87,
            "max": 223.25,
            "mean": 206.96
          },
          "cpu": {
            "min": 0.47,
            "max": 79.87,
            "mean": 3.9
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 57.15,
            "max": 57.52,
            "mean": 57.34
          },
          "cpu": {
            "min": 0,
            "max": 99.97,
            "mean": 13.49
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
      "throughput": 5256.52,
      "totalRequests": 157696,
      "latency": {
        "min": 389,
        "mean": 6320,
        "max": 77848,
        "pstdev": 11829,
        "percentiles": {
          "p50": 2974,
          "p75": 4705,
          "p80": 5292,
          "p90": 7729,
          "p95": 47173,
          "p99": 54317,
          "p999": 64679
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 186.16,
            "max": 199.19,
            "mean": 195.68
          },
          "cpu": {
            "min": 0.4,
            "max": 1.2,
            "mean": 0.75
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 63.26,
            "max": 79.43,
            "mean": 79.04
          },
          "cpu": {
            "min": 0,
            "max": 99.86,
            "mean": 17.64
          }
        }
      },
      "poolOverflow": 365,
      "upstreamConnections": 35
    },
    {
      "testName": "scaling up httproutes to 1000 with 200 routes per hostname",
      "routes": 1000,
      "routesPerHostname": 200,
      "phase": "scaling-up",
      "throughput": 5280.09,
      "totalRequests": 158409,
      "latency": {
        "min": 387,
        "mean": 6871,
        "max": 94064,
        "pstdev": 12441,
        "percentiles": {
          "p50": 3186,
          "p75": 5019,
          "p80": 5651,
          "p90": 8766,
          "p95": 48361,
          "p99": 55758,
          "p999": 67403
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 230.81,
            "max": 243.68,
            "mean": 238.65
          },
          "cpu": {
            "min": 0.13,
            "max": 1.13,
            "mean": 0.8
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 127.63,
            "max": 127.86,
            "mean": 127.72
          },
          "cpu": {
            "min": 0,
            "max": 85.57,
            "mean": 6.19
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
      "throughput": 5430.18,
      "totalRequests": 162909,
      "latency": {
        "min": 376,
        "mean": 6679,
        "max": 79630,
        "pstdev": 12142,
        "percentiles": {
          "p50": 3147,
          "p75": 5015,
          "p80": 5630,
          "p90": 8374,
          "p95": 48377,
          "p99": 54654,
          "p999": 59858
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 164.61,
            "max": 169.75,
            "mean": 165.39
          },
          "cpu": {
            "min": 0.8,
            "max": 3.47,
            "mean": 1.14
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 121.22,
            "max": 121.61,
            "mean": 121.3
          },
          "cpu": {
            "min": 0,
            "max": 99.74,
            "mean": 8.02
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38
    },
    {
      "testName": "scaling down httproutes to 50 with 10 routes per hostname",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-down",
      "throughput": 5506.49,
      "totalRequests": 165195,
      "latency": {
        "min": 386,
        "mean": 8319,
        "max": 85860,
        "pstdev": 13567,
        "percentiles": {
          "p50": 3951,
          "p75": 6305,
          "p80": 7155,
          "p90": 12741,
          "p95": 50714,
          "p99": 57303,
          "p999": 63928
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 165.63,
            "max": 175.23,
            "mean": 169.06
          },
          "cpu": {
            "min": 0.8,
            "max": 7.53,
            "mean": 1.54
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 111.91,
            "max": 121.36,
            "mean": 113.46
          },
          "cpu": {
            "min": 0,
            "max": 99.75,
            "mean": 7.62
          }
        }
      },
      "poolOverflow": 352,
      "upstreamConnections": 48
    },
    {
      "testName": "scaling down httproutes to 100 with 20 routes per hostname",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-down",
      "throughput": 5452.96,
      "totalRequests": 163589,
      "latency": {
        "min": 395,
        "mean": 7694,
        "max": 97460,
        "pstdev": 12949,
        "percentiles": {
          "p50": 3636,
          "p75": 5828,
          "p80": 6627,
          "p90": 10929,
          "p95": 49506,
          "p99": 55681,
          "p999": 62369
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 166.64,
            "max": 195.48,
            "mean": 174.79
          },
          "cpu": {
            "min": 0.8,
            "max": 38.2,
            "mean": 5.63
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 111.8,
            "max": 112.36,
            "mean": 111.95
          },
          "cpu": {
            "min": 0,
            "max": 35.17,
            "mean": 2.77
          }
        }
      },
      "poolOverflow": 356,
      "upstreamConnections": 44
    },
    {
      "testName": "scaling down httproutes to 300 with 60 routes per hostname",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-down",
      "throughput": 5360.33,
      "totalRequests": 160815,
      "latency": {
        "min": 354,
        "mean": 6538,
        "max": 92344,
        "pstdev": 12045,
        "percentiles": {
          "p50": 3068,
          "p75": 4770,
          "p80": 5341,
          "p90": 7974,
          "p95": 47742,
          "p99": 54480,
          "p999": 62144
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 180.27,
            "max": 192.65,
            "mean": 189.01
          },
          "cpu": {
            "min": 0.67,
            "max": 46.53,
            "mean": 4.41
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 127.66,
            "max": 127.84,
            "mean": 127.72
          },
          "cpu": {
            "min": 0,
            "max": 100.03,
            "mean": 9.14
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
      "throughput": 5334.03,
      "totalRequests": 160024,
      "latency": {
        "min": 392,
        "mean": 6782,
        "max": 132493,
        "pstdev": 12373,
        "percentiles": {
          "p50": 3182,
          "p75": 4977,
          "p80": 5586,
          "p90": 8359,
          "p95": 48365,
          "p99": 55252,
          "p999": 64712
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 203.37,
            "max": 217.03,
            "mean": 209.04
          },
          "cpu": {
            "min": 0.67,
            "max": 5.67,
            "mean": 1.33
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 127.65,
            "max": 127.84,
            "mean": 127.72
          },
          "cpu": {
            "min": 0,
            "max": 99.86,
            "mean": 9.91
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38
    }
  ]
};

export default benchmarkData;
