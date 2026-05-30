import { TestSuite } from '../types';

// Benchmark data extracted from markdown report for version 1.5.0
// Generated on 2025-11-21T02:41:56.137Z

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.5.0",
    "runId": "1.5.0-1763692916136",
    "date": "2025-11-21T02:41:56.136Z",
    "environment": "GitHub CI",
    "description": "Benchmark results for version 1.5.0",
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
      "throughput": 5231,
      "totalRequests": 156932,
      "latency": {
        "min": 0.37,
        "mean": 7.156,
        "max": 72.495,
        "pstdev": 12.509,
        "percentiles": {
          "p50": 3.426,
          "p75": 5.422,
          "p80": 6.073,
          "p90": 9.416,
          "p95": 48.619,
          "p99": 54.312,
          "p999": 58.814
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 111.95,
            "max": 137.44,
            "mean": 132.26
          },
          "cpu": {
            "min": 0.13,
            "max": 1.13,
            "mean": 0.45
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 0,
            "max": 27.53,
            "mean": 21.99
          },
          "cpu": {
            "min": 0,
            "max": 99.97,
            "mean": 7.47
          }
        }
      },
      "poolOverflow": 360,
      "upstreamConnections": 40,
      "counters": {
        "benchmark.http_2xx": {
          "value": 156932,
          "perSecond": 5231
        },
        "benchmark.pool_overflow": {
          "value": 360,
          "perSecond": 12
        },
        "cluster_manager.cluster_added": {
          "value": 4,
          "perSecond": 0.13
        },
        "default.total_match_count": {
          "value": 4,
          "perSecond": 0.13
        },
        "membership_change": {
          "value": 4,
          "perSecond": 0.13
        },
        "runtime.load_success": {
          "value": 1,
          "perSecond": 0.03
        },
        "runtime.override_dir_not_exists": {
          "value": 1,
          "perSecond": 0.03
        },
        "upstream_cx_http1_total": {
          "value": 40,
          "perSecond": 1.33
        },
        "upstream_cx_rx_bytes_total": {
          "value": 24638324,
          "perSecond": 821267.28
        },
        "upstream_cx_total": {
          "value": 40,
          "perSecond": 1.33
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7063470,
          "perSecond": 235446.08
        },
        "upstream_rq_pending_overflow": {
          "value": 360,
          "perSecond": 12
        },
        "upstream_rq_pending_total": {
          "value": 40,
          "perSecond": 1.33
        },
        "upstream_rq_total": {
          "value": 156966,
          "perSecond": 5232.14
        }
      }
    },
    {
      "testName": "scaling up httproutes to 50 with 10 routes per hostname",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-up",
      "throughput": 5284.46,
      "totalRequests": 158534,
      "latency": {
        "min": 0.375,
        "mean": 7.325,
        "max": 70.68,
        "pstdev": 12.639,
        "percentiles": {
          "p50": 3.504,
          "p75": 5.444,
          "p80": 6.115,
          "p90": 9.472,
          "p95": 48.904,
          "p99": 55.166,
          "p999": 60.913
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 137.04,
            "max": 147.84,
            "mean": 140.78
          },
          "cpu": {
            "min": 0.33,
            "max": 4,
            "mean": 0.81
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 27.38,
            "max": 31.86,
            "mean": 31.1
          },
          "cpu": {
            "min": 0,
            "max": 77.81,
            "mean": 4.21
          }
        }
      },
      "poolOverflow": 359,
      "upstreamConnections": 41,
      "counters": {
        "benchmark.http_2xx": {
          "value": 158534,
          "perSecond": 5284.46
        },
        "benchmark.pool_overflow": {
          "value": 359,
          "perSecond": 11.97
        },
        "cluster_manager.cluster_added": {
          "value": 4,
          "perSecond": 0.13
        },
        "default.total_match_count": {
          "value": 4,
          "perSecond": 0.13
        },
        "membership_change": {
          "value": 4,
          "perSecond": 0.13
        },
        "runtime.load_success": {
          "value": 1,
          "perSecond": 0.03
        },
        "runtime.override_dir_not_exists": {
          "value": 1,
          "perSecond": 0.03
        },
        "upstream_cx_http1_total": {
          "value": 41,
          "perSecond": 1.37
        },
        "upstream_cx_rx_bytes_total": {
          "value": 24889838,
          "perSecond": 829660.38
        },
        "upstream_cx_total": {
          "value": 41,
          "perSecond": 1.37
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7135875,
          "perSecond": 237862.25
        },
        "upstream_rq_pending_overflow": {
          "value": 359,
          "perSecond": 11.97
        },
        "upstream_rq_pending_total": {
          "value": 41,
          "perSecond": 1.37
        },
        "upstream_rq_total": {
          "value": 158575,
          "perSecond": 5285.83
        }
      }
    },
    {
      "testName": "scaling up httproutes to 100 with 20 routes per hostname",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-up",
      "throughput": 5245.7,
      "totalRequests": 157371,
      "latency": {
        "min": 0.393,
        "mean": 7.594,
        "max": 95.199,
        "pstdev": 12.997,
        "percentiles": {
          "p50": 3.589,
          "p75": 5.678,
          "p80": 6.423,
          "p90": 10.334,
          "p95": 49.883,
          "p99": 56.033,
          "p999": 61.966
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 140.89,
            "max": 158.13,
            "mean": 147.17
          },
          "cpu": {
            "min": 0.33,
            "max": 5.13,
            "mean": 0.98
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 31.43,
            "max": 37.71,
            "mean": 37.04
          },
          "cpu": {
            "min": 0,
            "max": 19.93,
            "mean": 1.95
          }
        }
      },
      "poolOverflow": 358,
      "upstreamConnections": 42,
      "counters": {
        "benchmark.http_2xx": {
          "value": 157371,
          "perSecond": 5245.7
        },
        "benchmark.pool_overflow": {
          "value": 358,
          "perSecond": 11.93
        },
        "cluster_manager.cluster_added": {
          "value": 4,
          "perSecond": 0.13
        },
        "default.total_match_count": {
          "value": 4,
          "perSecond": 0.13
        },
        "membership_change": {
          "value": 4,
          "perSecond": 0.13
        },
        "runtime.load_success": {
          "value": 1,
          "perSecond": 0.03
        },
        "runtime.override_dir_not_exists": {
          "value": 1,
          "perSecond": 0.03
        },
        "upstream_cx_http1_total": {
          "value": 42,
          "perSecond": 1.4
        },
        "upstream_cx_rx_bytes_total": {
          "value": 24707247,
          "perSecond": 823574.27
        },
        "upstream_cx_total": {
          "value": 42,
          "perSecond": 1.4
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7083585,
          "perSecond": 236119.32
        },
        "upstream_rq_pending_overflow": {
          "value": 358,
          "perSecond": 11.93
        },
        "upstream_rq_pending_total": {
          "value": 42,
          "perSecond": 1.4
        },
        "upstream_rq_total": {
          "value": 157413,
          "perSecond": 5247.1
        }
      }
    },
    {
      "testName": "scaling up httproutes to 300 with 60 routes per hostname",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-up",
      "throughput": 5168.17,
      "totalRequests": 155048,
      "latency": {
        "min": 0.375,
        "mean": 6.991,
        "max": 103.972,
        "pstdev": 12.378,
        "percentiles": {
          "p50": 3.286,
          "p75": 5.28,
          "p80": 5.974,
          "p90": 9.296,
          "p95": 48.189,
          "p99": 55.103,
          "p999": 64.542
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 148.22,
            "max": 156.47,
            "mean": 153.66
          },
          "cpu": {
            "min": 0.47,
            "max": 16.07,
            "mean": 2.08
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 37.69,
            "max": 60.02,
            "mean": 57.65
          },
          "cpu": {
            "min": 0,
            "max": 99.73,
            "mean": 14.01
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38,
      "counters": {
        "benchmark.http_2xx": {
          "value": 155048,
          "perSecond": 5168.17
        },
        "benchmark.pool_overflow": {
          "value": 362,
          "perSecond": 12.07
        },
        "cluster_manager.cluster_added": {
          "value": 4,
          "perSecond": 0.13
        },
        "default.total_match_count": {
          "value": 4,
          "perSecond": 0.13
        },
        "membership_change": {
          "value": 4,
          "perSecond": 0.13
        },
        "runtime.load_success": {
          "value": 1,
          "perSecond": 0.03
        },
        "runtime.override_dir_not_exists": {
          "value": 1,
          "perSecond": 0.03
        },
        "upstream_cx_http1_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_rx_bytes_total": {
          "value": 24342536,
          "perSecond": 811402.27
        },
        "upstream_cx_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_tx_bytes_total": {
          "value": 6978780,
          "perSecond": 232621.53
        },
        "upstream_rq_pending_overflow": {
          "value": 362,
          "perSecond": 12.07
        },
        "upstream_rq_pending_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_rq_total": {
          "value": 155084,
          "perSecond": 5169.37
        }
      }
    },
    {
      "testName": "scaling up httproutes to 500 with 100 routes per hostname",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-up",
      "throughput": 5173.5,
      "totalRequests": 155210,
      "latency": {
        "min": 0.389,
        "mean": 6.827,
        "max": 83.619,
        "pstdev": 12.261,
        "percentiles": {
          "p50": 3.164,
          "p75": 5.067,
          "p80": 5.727,
          "p90": 8.965,
          "p95": 48.003,
          "p99": 54.861,
          "p999": 64.55
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 160.69,
            "max": 181.29,
            "mean": 168.42
          },
          "cpu": {
            "min": 0.4,
            "max": 14.73,
            "mean": 2.16
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 59.88,
            "max": 80.35,
            "mean": 78.09
          },
          "cpu": {
            "min": 0,
            "max": 100.13,
            "mean": 17.03
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37,
      "counters": {
        "benchmark.http_2xx": {
          "value": 155210,
          "perSecond": 5173.5
        },
        "benchmark.pool_overflow": {
          "value": 363,
          "perSecond": 12.1
        },
        "cluster_manager.cluster_added": {
          "value": 4,
          "perSecond": 0.13
        },
        "default.total_match_count": {
          "value": 4,
          "perSecond": 0.13
        },
        "membership_change": {
          "value": 4,
          "perSecond": 0.13
        },
        "runtime.load_success": {
          "value": 1,
          "perSecond": 0.03
        },
        "runtime.override_dir_not_exists": {
          "value": 1,
          "perSecond": 0.03
        },
        "upstream_cx_http1_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_cx_rx_bytes_total": {
          "value": 24367970,
          "perSecond": 812239.92
        },
        "upstream_cx_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_cx_tx_bytes_total": {
          "value": 6985665,
          "perSecond": 232848.12
        },
        "upstream_rq_pending_overflow": {
          "value": 363,
          "perSecond": 12.1
        },
        "upstream_rq_pending_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_rq_total": {
          "value": 155237,
          "perSecond": 5174.4
        }
      }
    },
    {
      "testName": "scaling up httproutes to 1000 with 200 routes per hostname",
      "routes": 1000,
      "routesPerHostname": 200,
      "phase": "scaling-up",
      "throughput": 4954.6,
      "totalRequests": 148639,
      "latency": {
        "min": 0.362,
        "mean": 6.957,
        "max": 94.007,
        "pstdev": 12.255,
        "percentiles": {
          "p50": 3.293,
          "p75": 5.279,
          "p80": 5.991,
          "p90": 9.369,
          "p95": 46.555,
          "p99": 55.103,
          "p999": 68.62
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 182.77,
            "max": 198.76,
            "mean": 195.11
          },
          "cpu": {
            "min": 0.47,
            "max": 33.07,
            "mean": 4.47
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 112.02,
            "max": 136.36,
            "mean": 133.37
          },
          "cpu": {
            "min": 0,
            "max": 99.75,
            "mean": 7.96
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37,
      "counters": {
        "benchmark.http_2xx": {
          "value": 148639,
          "perSecond": 4954.6
        },
        "benchmark.pool_overflow": {
          "value": 363,
          "perSecond": 12.1
        },
        "cluster_manager.cluster_added": {
          "value": 4,
          "perSecond": 0.13
        },
        "default.total_match_count": {
          "value": 4,
          "perSecond": 0.13
        },
        "membership_change": {
          "value": 4,
          "perSecond": 0.13
        },
        "runtime.load_success": {
          "value": 1,
          "perSecond": 0.03
        },
        "runtime.override_dir_not_exists": {
          "value": 1,
          "perSecond": 0.03
        },
        "upstream_cx_http1_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_cx_rx_bytes_total": {
          "value": 23336323,
          "perSecond": 777871.58
        },
        "upstream_cx_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_cx_tx_bytes_total": {
          "value": 6689475,
          "perSecond": 222980.82
        },
        "upstream_rq_pending_overflow": {
          "value": 363,
          "perSecond": 12.1
        },
        "upstream_rq_pending_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_rq_total": {
          "value": 148655,
          "perSecond": 4955.13
        }
      }
    },
    {
      "testName": "scaling down httproutes to 10 with 2 routes per hostname",
      "routes": 10,
      "routesPerHostname": 2,
      "phase": "scaling-down",
      "throughput": 5243.85,
      "totalRequests": 157319,
      "latency": {
        "min": 0.374,
        "mean": 7.087,
        "max": 80.916,
        "pstdev": 12.384,
        "percentiles": {
          "p50": 3.347,
          "p75": 5.37,
          "p80": 6.082,
          "p90": 9.511,
          "p95": 48.24,
          "p99": 54.714,
          "p999": 60.506
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 140.18,
            "max": 149.41,
            "mean": 143.1
          },
          "cpu": {
            "min": 0.67,
            "max": 3.27,
            "mean": 1
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 127.76,
            "max": 127.94,
            "mean": 127.82
          },
          "cpu": {
            "min": 0,
            "max": 99.8,
            "mean": 5.93
          }
        }
      },
      "poolOverflow": 361,
      "upstreamConnections": 39,
      "counters": {
        "benchmark.http_2xx": {
          "value": 157319,
          "perSecond": 5243.85
        },
        "benchmark.pool_overflow": {
          "value": 361,
          "perSecond": 12.03
        },
        "cluster_manager.cluster_added": {
          "value": 4,
          "perSecond": 0.13
        },
        "default.total_match_count": {
          "value": 4,
          "perSecond": 0.13
        },
        "membership_change": {
          "value": 4,
          "perSecond": 0.13
        },
        "runtime.load_success": {
          "value": 1,
          "perSecond": 0.03
        },
        "runtime.override_dir_not_exists": {
          "value": 1,
          "perSecond": 0.03
        },
        "upstream_cx_http1_total": {
          "value": 39,
          "perSecond": 1.3
        },
        "upstream_cx_rx_bytes_total": {
          "value": 24699083,
          "perSecond": 823284.86
        },
        "upstream_cx_total": {
          "value": 39,
          "perSecond": 1.3
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7080885,
          "perSecond": 236024.37
        },
        "upstream_rq_pending_overflow": {
          "value": 361,
          "perSecond": 12.03
        },
        "upstream_rq_pending_total": {
          "value": 39,
          "perSecond": 1.3
        },
        "upstream_rq_total": {
          "value": 157353,
          "perSecond": 5244.99
        }
      }
    },
    {
      "testName": "scaling down httproutes to 50 with 10 routes per hostname",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-down",
      "throughput": 5233.21,
      "totalRequests": 157002,
      "latency": {
        "min": 0.378,
        "mean": 6.898,
        "max": 74.907,
        "pstdev": 12.301,
        "percentiles": {
          "p50": 3.22,
          "p75": 5.159,
          "p80": 5.828,
          "p90": 9.039,
          "p95": 48.322,
          "p99": 54.702,
          "p999": 60.03
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 145.41,
            "max": 151.13,
            "mean": 148.35
          },
          "cpu": {
            "min": 0.6,
            "max": 4.07,
            "mean": 1.07
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 127.77,
            "max": 128.41,
            "mean": 127.86
          },
          "cpu": {
            "min": 0,
            "max": 99.93,
            "mean": 8.94
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38,
      "counters": {
        "benchmark.http_2xx": {
          "value": 157002,
          "perSecond": 5233.21
        },
        "benchmark.pool_overflow": {
          "value": 362,
          "perSecond": 12.07
        },
        "cluster_manager.cluster_added": {
          "value": 4,
          "perSecond": 0.13
        },
        "default.total_match_count": {
          "value": 4,
          "perSecond": 0.13
        },
        "membership_change": {
          "value": 4,
          "perSecond": 0.13
        },
        "runtime.load_success": {
          "value": 1,
          "perSecond": 0.03
        },
        "runtime.override_dir_not_exists": {
          "value": 1,
          "perSecond": 0.03
        },
        "upstream_cx_http1_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_rx_bytes_total": {
          "value": 24649314,
          "perSecond": 821613.48
        },
        "upstream_cx_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7066125,
          "perSecond": 235528.81
        },
        "upstream_rq_pending_overflow": {
          "value": 362,
          "perSecond": 12.07
        },
        "upstream_rq_pending_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_rq_total": {
          "value": 157025,
          "perSecond": 5233.97
        }
      }
    },
    {
      "testName": "scaling down httproutes to 100 with 20 routes per hostname",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-down",
      "throughput": 5193.37,
      "totalRequests": 155801,
      "latency": {
        "min": 0.356,
        "mean": 6.97,
        "max": 134.422,
        "pstdev": 12.396,
        "percentiles": {
          "p50": 3.245,
          "p75": 5.229,
          "p80": 5.909,
          "p90": 9.186,
          "p95": 48.029,
          "p99": 55.148,
          "p999": 61.863
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 146.94,
            "max": 160.41,
            "mean": 148.54
          },
          "cpu": {
            "min": 0.6,
            "max": 9.93,
            "mean": 1.66
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 127.69,
            "max": 136.18,
            "mean": 127.97
          },
          "cpu": {
            "min": 0,
            "max": 99.76,
            "mean": 19.75
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38,
      "counters": {
        "benchmark.http_2xx": {
          "value": 155801,
          "perSecond": 5193.37
        },
        "benchmark.pool_overflow": {
          "value": 362,
          "perSecond": 12.07
        },
        "cluster_manager.cluster_added": {
          "value": 4,
          "perSecond": 0.13
        },
        "default.total_match_count": {
          "value": 4,
          "perSecond": 0.13
        },
        "membership_change": {
          "value": 4,
          "perSecond": 0.13
        },
        "runtime.load_success": {
          "value": 1,
          "perSecond": 0.03
        },
        "runtime.override_dir_not_exists": {
          "value": 1,
          "perSecond": 0.03
        },
        "upstream_cx_http1_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_rx_bytes_total": {
          "value": 24460757,
          "perSecond": 815358.63
        },
        "upstream_cx_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7012755,
          "perSecond": 233758.52
        },
        "upstream_rq_pending_overflow": {
          "value": 362,
          "perSecond": 12.07
        },
        "upstream_rq_pending_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_rq_total": {
          "value": 155839,
          "perSecond": 5194.63
        }
      }
    },
    {
      "testName": "scaling down httproutes to 300 with 60 routes per hostname",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-down",
      "throughput": 5202.79,
      "totalRequests": 156086,
      "latency": {
        "min": 0.363,
        "mean": 6.942,
        "max": 98.598,
        "pstdev": 12.536,
        "percentiles": {
          "p50": 3.224,
          "p75": 5.05,
          "p80": 5.671,
          "p90": 8.706,
          "p95": 49.022,
          "p99": 55.58,
          "p999": 66.095
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 154.45,
            "max": 171.48,
            "mean": 159.56
          },
          "cpu": {
            "min": 0.53,
            "max": 12.73,
            "mean": 1.91
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 136.16,
            "max": 136.36,
            "mean": 136.23
          },
          "cpu": {
            "min": 0,
            "max": 100.08,
            "mean": 15.35
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38,
      "counters": {
        "benchmark.http_2xx": {
          "value": 156086,
          "perSecond": 5202.79
        },
        "benchmark.pool_overflow": {
          "value": 362,
          "perSecond": 12.07
        },
        "cluster_manager.cluster_added": {
          "value": 4,
          "perSecond": 0.13
        },
        "default.total_match_count": {
          "value": 4,
          "perSecond": 0.13
        },
        "membership_change": {
          "value": 4,
          "perSecond": 0.13
        },
        "runtime.load_success": {
          "value": 1,
          "perSecond": 0.03
        },
        "runtime.override_dir_not_exists": {
          "value": 1,
          "perSecond": 0.03
        },
        "upstream_cx_http1_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_rx_bytes_total": {
          "value": 24505502,
          "perSecond": 816837.99
        },
        "upstream_cx_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7024770,
          "perSecond": 234155.54
        },
        "upstream_rq_pending_overflow": {
          "value": 362,
          "perSecond": 12.07
        },
        "upstream_rq_pending_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_rq_total": {
          "value": 156106,
          "perSecond": 5203.46
        }
      }
    },
    {
      "testName": "scaling down httproutes to 500 with 100 routes per hostname",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-down",
      "throughput": 5146.66,
      "totalRequests": 154400,
      "latency": {
        "min": 0.379,
        "mean": 6.86,
        "max": 92.921,
        "pstdev": 12.588,
        "percentiles": {
          "p50": 3.144,
          "p75": 4.942,
          "p80": 5.573,
          "p90": 8.615,
          "p95": 49.158,
          "p99": 56.338,
          "p999": 65.662
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 161.58,
            "max": 194.96,
            "mean": 170.05
          },
          "cpu": {
            "min": 0.53,
            "max": 23.73,
            "mean": 2.99
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 136.16,
            "max": 136.35,
            "mean": 136.22
          },
          "cpu": {
            "min": 0,
            "max": 100.09,
            "mean": 6.89
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37,
      "counters": {
        "benchmark.http_2xx": {
          "value": 154400,
          "perSecond": 5146.66
        },
        "benchmark.pool_overflow": {
          "value": 363,
          "perSecond": 12.1
        },
        "cluster_manager.cluster_added": {
          "value": 4,
          "perSecond": 0.13
        },
        "default.total_match_count": {
          "value": 4,
          "perSecond": 0.13
        },
        "membership_change": {
          "value": 4,
          "perSecond": 0.13
        },
        "runtime.load_success": {
          "value": 1,
          "perSecond": 0.03
        },
        "runtime.override_dir_not_exists": {
          "value": 1,
          "perSecond": 0.03
        },
        "upstream_cx_http1_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_cx_rx_bytes_total": {
          "value": 24240800,
          "perSecond": 808025.55
        },
        "upstream_cx_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_cx_tx_bytes_total": {
          "value": 6949665,
          "perSecond": 231655.18
        },
        "upstream_rq_pending_overflow": {
          "value": 363,
          "perSecond": 12.1
        },
        "upstream_rq_pending_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_rq_total": {
          "value": 154437,
          "perSecond": 5147.89
        }
      }
    }
  ]
};

export default benchmarkData;
