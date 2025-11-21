import { TestSuite } from '../types';

// Benchmark data extracted from markdown report for version 1.4.3
// Generated on 2025-11-21T02:41:39.819Z

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.4.3",
    "runId": "1.4.3-1763692899811",
    "date": "2025-11-21T02:41:39.811Z",
    "environment": "production",
    "description": "Benchmark results for version 1.4.3",
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
      "throughput": 5355.55,
      "totalRequests": 160667,
      "latency": {
        "min": 358,
        "mean": 7267,
        "max": 75100,
        "pstdev": 12628,
        "percentiles": {
          "p50": 3399,
          "p75": 5544,
          "p80": 6276,
          "p90": 9990,
          "p95": 49121,
          "p99": 55177,
          "p999": 60772
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 130.65,
            "max": 152.97,
            "mean": 147.82
          },
          "cpu": {
            "min": 0.2,
            "max": 1,
            "mean": 0.45
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 0,
            "max": 27.79,
            "mean": 23.38
          },
          "cpu": {
            "min": 0,
            "max": 99.91,
            "mean": 7.25
          }
        }
      },
      "poolOverflow": 359,
      "upstreamConnections": 41,
      "counters": {
        "benchmark.http_2xx": {
          "value": 160667,
          "perSecond": 5355.55
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
          "value": 25224719,
          "perSecond": 840822.06
        },
        "upstream_cx_total": {
          "value": 41,
          "perSecond": 1.37
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7231860,
          "perSecond": 241061.45
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
          "value": 160708,
          "perSecond": 5356.92
        }
      }
    },
    {
      "testName": "scaling up httproutes to 50 with 10 routes per hostname",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-up",
      "throughput": 5340.22,
      "totalRequests": 160216,
      "latency": {
        "min": 390,
        "mean": 6790,
        "max": 86708,
        "pstdev": 12266,
        "percentiles": {
          "p50": 3114,
          "p75": 5075,
          "p80": 5778,
          "p90": 8996,
          "p95": 48447,
          "p99": 54765,
          "p999": 59922
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 153.72,
            "max": 164.07,
            "mean": 160.07
          },
          "cpu": {
            "min": 0.4,
            "max": 4.4,
            "mean": 0.9
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 27.43,
            "max": 31.89,
            "mean": 31.33
          },
          "cpu": {
            "min": 0,
            "max": 100.05,
            "mean": 11.23
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38,
      "counters": {
        "benchmark.http_2xx": {
          "value": 160216,
          "perSecond": 5340.22
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
          "value": 25153912,
          "perSecond": 838414.8
        },
        "upstream_cx_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7211070,
          "perSecond": 240354.97
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
          "value": 160246,
          "perSecond": 5341.22
        }
      }
    },
    {
      "testName": "scaling up httproutes to 100 with 20 routes per hostname",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-up",
      "throughput": 5324.98,
      "totalRequests": 159753,
      "latency": {
        "min": 366,
        "mean": 6637,
        "max": 92008,
        "pstdev": 12072,
        "percentiles": {
          "p50": 3127,
          "p75": 4922,
          "p80": 5533,
          "p90": 8308,
          "p95": 48070,
          "p99": 54439,
          "p999": 60065
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 164.91,
            "max": 170.05,
            "mean": 166.44
          },
          "cpu": {
            "min": 0.4,
            "max": 8.2,
            "mean": 1.32
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 31.52,
            "max": 37.83,
            "mean": 37.06
          },
          "cpu": {
            "min": 0,
            "max": 92.36,
            "mean": 5.35
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37,
      "counters": {
        "benchmark.http_2xx": {
          "value": 159753,
          "perSecond": 5324.98
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
          "value": 25081221,
          "perSecond": 836022.33
        },
        "upstream_cx_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7190550,
          "perSecond": 239679.73
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
          "value": 159790,
          "perSecond": 5326.22
        }
      }
    },
    {
      "testName": "scaling up httproutes to 300 with 60 routes per hostname",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-up",
      "throughput": 5216.76,
      "totalRequests": 156503,
      "latency": {
        "min": 377,
        "mean": 6720,
        "max": 78270,
        "pstdev": 12030,
        "percentiles": {
          "p50": 3196,
          "p75": 5018,
          "p80": 5633,
          "p90": 8582,
          "p95": 47130,
          "p99": 54605,
          "p999": 64385
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 179.89,
            "max": 190.89,
            "mean": 188.89
          },
          "cpu": {
            "min": 0.53,
            "max": 37.94,
            "mean": 4.32
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 51.75,
            "max": 58.28,
            "mean": 57.56
          },
          "cpu": {
            "min": 0,
            "max": 99.97,
            "mean": 10.27
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37,
      "counters": {
        "benchmark.http_2xx": {
          "value": 156503,
          "perSecond": 5216.76
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
          "value": 24570971,
          "perSecond": 819031.62
        },
        "upstream_cx_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7044300,
          "perSecond": 234809.79
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
          "value": 156540,
          "perSecond": 5218
        }
      }
    },
    {
      "testName": "scaling up httproutes to 500 with 100 routes per hostname",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-up",
      "throughput": 5209.26,
      "totalRequests": 156278,
      "latency": {
        "min": 390,
        "mean": 6375,
        "max": 101388,
        "pstdev": 11924,
        "percentiles": {
          "p50": 2960,
          "p75": 4644,
          "p80": 5225,
          "p90": 7832,
          "p95": 47497,
          "p99": 54636,
          "p999": 65597
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 198.36,
            "max": 209.82,
            "mean": 204.77
          },
          "cpu": {
            "min": 0.47,
            "max": 39.55,
            "mean": 2.27
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 77.96,
            "max": 78.17,
            "mean": 78.04
          },
          "cpu": {
            "min": 0,
            "max": 99.98,
            "mean": 5.25
          }
        }
      },
      "poolOverflow": 365,
      "upstreamConnections": 35,
      "counters": {
        "benchmark.http_2xx": {
          "value": 156278,
          "perSecond": 5209.26
        },
        "benchmark.pool_overflow": {
          "value": 365,
          "perSecond": 12.17
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
          "value": 35,
          "perSecond": 1.17
        },
        "upstream_cx_rx_bytes_total": {
          "value": 24535646,
          "perSecond": 817853.19
        },
        "upstream_cx_total": {
          "value": 35,
          "perSecond": 1.17
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7034085,
          "perSecond": 234469.02
        },
        "upstream_rq_pending_overflow": {
          "value": 365,
          "perSecond": 12.17
        },
        "upstream_rq_pending_total": {
          "value": 35,
          "perSecond": 1.17
        },
        "upstream_rq_total": {
          "value": 156313,
          "perSecond": 5210.42
        }
      }
    },
    {
      "testName": "scaling up httproutes to 1000 with 200 routes per hostname",
      "routes": 1000,
      "routesPerHostname": 200,
      "phase": "scaling-up",
      "throughput": 5098.94,
      "totalRequests": 152971,
      "latency": {
        "min": 359,
        "mean": 6888,
        "max": 82657,
        "pstdev": 12606,
        "percentiles": {
          "p50": 3124,
          "p75": 4997,
          "p80": 5658,
          "p90": 8935,
          "p95": 48592,
          "p99": 56840,
          "p999": 68640
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 228.83,
            "max": 240.91,
            "mean": 237.53
          },
          "cpu": {
            "min": 0.13,
            "max": 1.2,
            "mean": 0.8
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 126.34,
            "max": 126.55,
            "mean": 126.44
          },
          "cpu": {
            "min": 0,
            "max": 100.05,
            "mean": 11.87
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37,
      "counters": {
        "benchmark.http_2xx": {
          "value": 152971,
          "perSecond": 5098.94
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
          "value": 24016447,
          "perSecond": 800534.22
        },
        "upstream_cx_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_cx_tx_bytes_total": {
          "value": 6884370,
          "perSecond": 229474.98
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
          "value": 152986,
          "perSecond": 5099.44
        }
      }
    },
    {
      "testName": "scaling down httproutes to 10 with 2 routes per hostname",
      "routes": 10,
      "routesPerHostname": 2,
      "phase": "scaling-down",
      "throughput": 5357.63,
      "totalRequests": 160736,
      "latency": {
        "min": 377,
        "mean": 6733,
        "max": 73097,
        "pstdev": 11999,
        "percentiles": {
          "p50": 3204,
          "p75": 5032,
          "p80": 5678,
          "p90": 8677,
          "p95": 47718,
          "p99": 54052,
          "p999": 59168
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 238.38,
            "max": 246.22,
            "mean": 241.42
          },
          "cpu": {
            "min": 0.73,
            "max": 3.93,
            "mean": 1.17
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 122.09,
            "max": 122.25,
            "mean": 122.14
          },
          "cpu": {
            "min": 0,
            "max": 99.72,
            "mean": 9.75
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38,
      "counters": {
        "benchmark.http_2xx": {
          "value": 160736,
          "perSecond": 5357.63
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
          "value": 25235552,
          "perSecond": 841148.66
        },
        "upstream_cx_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7234650,
          "perSecond": 241144.56
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
          "value": 160770,
          "perSecond": 5358.77
        }
      }
    },
    {
      "testName": "scaling down httproutes to 50 with 10 routes per hostname",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-down",
      "throughput": 5332.35,
      "totalRequests": 159971,
      "latency": {
        "min": 390,
        "mean": 6800,
        "max": 72916,
        "pstdev": 12278,
        "percentiles": {
          "p50": 3170,
          "p75": 4982,
          "p80": 5616,
          "p90": 8533,
          "p95": 48674,
          "p99": 55203,
          "p999": 59807
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 240.27,
            "max": 254.06,
            "mean": 244.17
          },
          "cpu": {
            "min": 0.67,
            "max": 7.07,
            "mean": 1.48
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 122.07,
            "max": 122.28,
            "mean": 122.13
          },
          "cpu": {
            "min": 0,
            "max": 100.09,
            "mean": 9.45
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38,
      "counters": {
        "benchmark.http_2xx": {
          "value": 159971,
          "perSecond": 5332.35
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
          "value": 25115447,
          "perSecond": 837179.18
        },
        "upstream_cx_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7200405,
          "perSecond": 240012.82
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
          "value": 160009,
          "perSecond": 5333.62
        }
      }
    },
    {
      "testName": "scaling down httproutes to 100 with 20 routes per hostname",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-down",
      "throughput": 5327.29,
      "totalRequests": 159819,
      "latency": {
        "min": 369,
        "mean": 6798,
        "max": 75177,
        "pstdev": 12404,
        "percentiles": {
          "p50": 3117,
          "p75": 4972,
          "p80": 5617,
          "p90": 8628,
          "p95": 49203,
          "p99": 55308,
          "p999": 61265
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 183.42,
            "max": 257.85,
            "mean": 238.95
          },
          "cpu": {
            "min": 0.67,
            "max": 67.53,
            "mean": 7.3
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 122.02,
            "max": 122.51,
            "mean": 122.15
          },
          "cpu": {
            "min": 0,
            "max": 99.85,
            "mean": 14.48
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38,
      "counters": {
        "benchmark.http_2xx": {
          "value": 159819,
          "perSecond": 5327.29
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
          "value": 25091583,
          "perSecond": 836384.72
        },
        "upstream_cx_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7193565,
          "perSecond": 239785.1
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
          "value": 159857,
          "perSecond": 5328.56
        }
      }
    },
    {
      "testName": "scaling down httproutes to 300 with 60 routes per hostname",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-down",
      "throughput": 5289.39,
      "totalRequests": 158686,
      "latency": {
        "min": 368,
        "mean": 6822,
        "max": 81756,
        "pstdev": 12344,
        "percentiles": {
          "p50": 3163,
          "p75": 4985,
          "p80": 5606,
          "p90": 8640,
          "p95": 48357,
          "p99": 55128,
          "p999": 63125
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 181.54,
            "max": 195.97,
            "mean": 192.74
          },
          "cpu": {
            "min": 0.33,
            "max": 72.2,
            "mean": 3.57
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 126.37,
            "max": 126.8,
            "mean": 126.47
          },
          "cpu": {
            "min": 0,
            "max": 62.71,
            "mean": 1.45
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38,
      "counters": {
        "benchmark.http_2xx": {
          "value": 158686,
          "perSecond": 5289.39
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
          "value": 24913702,
          "perSecond": 830433.73
        },
        "upstream_cx_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7142175,
          "perSecond": 238065.91
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
          "value": 158715,
          "perSecond": 5290.35
        }
      }
    },
    {
      "testName": "scaling down httproutes to 500 with 100 routes per hostname",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-down",
      "throughput": 5169.42,
      "totalRequests": 155086,
      "latency": {
        "min": 391,
        "mean": 7030,
        "max": 99708,
        "pstdev": 12686,
        "percentiles": {
          "p50": 3210,
          "p75": 5119,
          "p80": 5775,
          "p90": 9064,
          "p95": 49074,
          "p99": 56016,
          "p999": 65034
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 192.96,
            "max": 208.64,
            "mean": 205.72
          },
          "cpu": {
            "min": 0.33,
            "max": 40.87,
            "mean": 2.37
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 126.37,
            "max": 126.55,
            "mean": 126.42
          },
          "cpu": {
            "min": 0,
            "max": 67.89,
            "mean": 1.84
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38,
      "counters": {
        "benchmark.http_2xx": {
          "value": 155086,
          "perSecond": 5169.42
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
          "value": 24348502,
          "perSecond": 811598.23
        },
        "upstream_cx_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_tx_bytes_total": {
          "value": 6979860,
          "perSecond": 232656.7
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
          "value": 155108,
          "perSecond": 5170.15
        }
      }
    }
  ]
};

export default benchmarkData;
