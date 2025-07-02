import { TestSuite } from '../types';

// Benchmark data extracted from markdown report for version 1.2.5
// Generated on 2025-06-17T19:50:26.766Z

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.2.5",
    "runId": "1.2.5-1750189826766",
    "date": "2025-01-14",
    "environment": "GitHub CI",
    "description": "Benchmark report for Envoy Gateway 1.2.5",
    "downloadUrl": "https://github.com/envoyproxy/gateway/releases/download/v1.2.5/benchmark_report.zip",
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
      "throughput": 5704.46,
      "totalRequests": 171136,
      "latency": {
        "min": 370,
        "mean": 6520,
        "max": 74477,
        "pstdev": 11781,
        "percentiles": {
          "p50": 3051,
          "p75": 4913,
          "p80": 5565,
          "p90": 8700,
          "p95": 47091,
          "p99": 54638,
          "p999": 59404
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 112.07,
            "max": 112.07,
            "mean": 112.07
          },
          "cpu": {
            "min": 0.76,
            "max": 0.76,
            "mean": 0.76
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 24.38,
            "max": 24.38,
            "mean": 24.38
          },
          "cpu": {
            "min": 30.47,
            "max": 30.47,
            "mean": 30.47
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
      "throughput": 5647.86,
      "totalRequests": 169439,
      "latency": {
        "min": 379,
        "mean": 6313,
        "max": 80498,
        "pstdev": 11935,
        "percentiles": {
          "p50": 2854,
          "p75": 4550,
          "p80": 5158,
          "p90": 7820,
          "p95": 48025,
          "p99": 54390,
          "p999": 60321
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 123.4,
            "max": 123.4,
            "mean": 123.4
          },
          "cpu": {
            "min": 1.48,
            "max": 1.48,
            "mean": 1.48
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 30.55,
            "max": 30.55,
            "mean": 30.55
          },
          "cpu": {
            "min": 61.09,
            "max": 61.09,
            "mean": 61.09
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
      "throughput": 5529.29,
      "totalRequests": 165879,
      "latency": {
        "min": 379,
        "mean": 6240,
        "max": 73539,
        "pstdev": 12017,
        "percentiles": {
          "p50": 2818,
          "p75": 4510,
          "p80": 5068,
          "p90": 7413,
          "p95": 48300,
          "p99": 54171,
          "p999": 59308
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 139.7,
            "max": 139.7,
            "mean": 139.7
          },
          "cpu": {
            "min": 2.92,
            "max": 2.92,
            "mean": 2.92
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 36.73,
            "max": 36.73,
            "mean": 36.73
          },
          "cpu": {
            "min": 91.9,
            "max": 91.9,
            "mean": 91.9
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
      "throughput": 5542.3,
      "totalRequests": 166274,
      "latency": {
        "min": 365,
        "mean": 6463,
        "max": 81305,
        "pstdev": 12357,
        "percentiles": {
          "p50": 2890,
          "p75": 4545,
          "p80": 5116,
          "p90": 7585,
          "p95": 49043,
          "p99": 55365,
          "p999": 64399
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 146.11,
            "max": 146.11,
            "mean": 146.11
          },
          "cpu": {
            "min": 15.2,
            "max": 15.2,
            "mean": 15.2
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 58.91,
            "max": 58.91,
            "mean": 58.91
          },
          "cpu": {
            "min": 124.96,
            "max": 124.96,
            "mean": 124.96
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
      "throughput": 5407.59,
      "totalRequests": 162228,
      "latency": {
        "min": 363,
        "mean": 6396,
        "max": 88236,
        "pstdev": 12419,
        "percentiles": {
          "p50": 2816,
          "p75": 4450,
          "p80": 5012,
          "p90": 7485,
          "p95": 48762,
          "p99": 55666,
          "p999": 67457
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 154.01,
            "max": 154.01,
            "mean": 154.01
          },
          "cpu": {
            "min": 27.98,
            "max": 27.98,
            "mean": 27.98
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 81.11,
            "max": 81.11,
            "mean": 81.11
          },
          "cpu": {
            "min": 158.85,
            "max": 158.85,
            "mean": 158.85
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
      "throughput": 5247.32,
      "totalRequests": 157425,
      "latency": {
        "min": 374,
        "mean": 6604,
        "max": 106491,
        "pstdev": 12903,
        "percentiles": {
          "p50": 2880,
          "p75": 4514,
          "p80": 5089,
          "p90": 7639,
          "p95": 49453,
          "p99": 58750,
          "p999": 71970
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 183.52,
            "max": 183.52,
            "mean": 183.52
          },
          "cpu": {
            "min": 61.85,
            "max": 61.85,
            "mean": 61.85
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 129.33,
            "max": 129.33,
            "mean": 129.33
          },
          "cpu": {
            "min": 198.98,
            "max": 198.98,
            "mean": 198.98
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
      "throughput": 5668.05,
      "totalRequests": 170048,
      "latency": {
        "min": 366,
        "mean": 6143,
        "max": 73551,
        "pstdev": 12054,
        "percentiles": {
          "p50": 2726,
          "p75": 4279,
          "p80": 4799,
          "p90": 7123,
          "p95": 48988,
          "p99": 54704,
          "p999": 58812
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 138.57,
            "max": 138.57,
            "mean": 138.57
          },
          "cpu": {
            "min": 115.2,
            "max": 115.2,
            "mean": 115.2
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 121.88,
            "max": 121.88,
            "mean": 121.88
          },
          "cpu": {
            "min": 361.89,
            "max": 361.89,
            "mean": 361.89
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
      "throughput": 5579.16,
      "totalRequests": 167375,
      "latency": {
        "min": 353,
        "mean": 6190,
        "max": 87244,
        "pstdev": 12028,
        "percentiles": {
          "p50": 2803,
          "p75": 4357,
          "p80": 4889,
          "p90": 7110,
          "p95": 48482,
          "p99": 54329,
          "p999": 59113
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 132.83,
            "max": 132.83,
            "mean": 132.83
          },
          "cpu": {
            "min": 114.63,
            "max": 114.63,
            "mean": 114.63
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 122.16,
            "max": 122.16,
            "mean": 122.16
          },
          "cpu": {
            "min": 331.47,
            "max": 331.47,
            "mean": 331.47
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
      "throughput": 5581.7,
      "totalRequests": 167453,
      "latency": {
        "min": 377,
        "mean": 6525,
        "max": 72724,
        "pstdev": 12286,
        "percentiles": {
          "p50": 2929,
          "p75": 4641,
          "p80": 5219,
          "p90": 7915,
          "p95": 48732,
          "p99": 54810,
          "p999": 61450
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 133.91,
            "max": 133.91,
            "mean": 133.91
          },
          "cpu": {
            "min": 113.44,
            "max": 113.44,
            "mean": 113.44
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 122.02,
            "max": 122.02,
            "mean": 122.02
          },
          "cpu": {
            "min": 300.79,
            "max": 300.79,
            "mean": 300.79
          }
        }
      },
      "poolOverflow": 361,
      "upstreamConnections": 39
    },
    {
      "testName": "scale-down-httproutes-300",
      "routes": 300,
      "routesPerHostname": 1,
      "phase": "scaling-down",
      "throughput": 5447.53,
      "totalRequests": 163426,
      "latency": {
        "min": 368,
        "mean": 6540,
        "max": 88612,
        "pstdev": 12451,
        "percentiles": {
          "p50": 2904,
          "p75": 4546,
          "p80": 5128,
          "p90": 7692,
          "p95": 49305,
          "p99": 55422,
          "p999": 64116
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 145.28,
            "max": 145.28,
            "mean": 145.28
          },
          "cpu": {
            "min": 102.98,
            "max": 102.98,
            "mean": 102.98
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 129.93,
            "max": 129.93,
            "mean": 129.93
          },
          "cpu": {
            "min": 268.5,
            "max": 268.5,
            "mean": 268.5
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
      "throughput": 5437.61,
      "totalRequests": 163132,
      "latency": {
        "min": 383,
        "mean": 6527,
        "max": 102998,
        "pstdev": 12600,
        "percentiles": {
          "p50": 2827,
          "p75": 4470,
          "p80": 5070,
          "p90": 7714,
          "p95": 49029,
          "p99": 56436,
          "p999": 66684
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 164.07,
            "max": 164.07,
            "mean": 164.07
          },
          "cpu": {
            "min": 91.64,
            "max": 91.64,
            "mean": 91.64
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 129.93,
            "max": 129.93,
            "mean": 129.93
          },
          "cpu": {
            "min": 235.61,
            "max": 235.61,
            "mean": 235.61
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38
    }
  ]
};

export default benchmarkData;
