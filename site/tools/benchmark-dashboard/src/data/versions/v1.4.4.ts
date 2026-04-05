import { TestSuite } from '../types';

// Benchmark data extracted from markdown report for version 1.4.4
// Generated on 2025-11-21T02:41:45.325Z

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.4.4",
    "runId": "1.4.4-1763692905323",
    "date": "2025-11-21T02:41:45.323Z",
    "environment": "GitHub CI",
    "description": "Benchmark results for version 1.4.4",
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
      "throughput": 5276.66,
      "totalRequests": 158309,
      "latency": {
        "min": 0.385,
        "mean": 7.177,
        "max": 74.239,
        "pstdev": 12.629,
        "percentiles": {
          "p50": 3.424,
          "p75": 5.345,
          "p80": 6.0,
          "p90": 9.149,
          "p95": 49.1,
          "p99": 55.275,
          "p999": 59.793
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 129.45,
            "max": 153.16,
            "mean": 149.11
          },
          "cpu": {
            "min": 0.27,
            "max": 0.87,
            "mean": 0.44
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 0,
            "max": 28.43,
            "mean": 24.06
          },
          "cpu": {
            "min": 0,
            "max": 99.95,
            "mean": 9.16
          }
        }
      },
      "poolOverflow": 360,
      "upstreamConnections": 40,
      "counters": {
        "benchmark.http_2xx": {
          "value": 158309,
          "perSecond": 5276.66
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
          "value": 24854513,
          "perSecond": 828435.92
        },
        "upstream_cx_total": {
          "value": 40,
          "perSecond": 1.33
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7125435,
          "perSecond": 237500.78
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
          "value": 158343,
          "perSecond": 5277.8
        }
      }
    },
    {
      "testName": "scaling up httproutes to 50 with 10 routes per hostname",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-up",
      "throughput": 5230.73,
      "totalRequests": 156922,
      "latency": {
        "min": 0.367,
        "mean": 6.673,
        "max": 75.735,
        "pstdev": 11.892,
        "percentiles": {
          "p50": 3.219,
          "p75": 5.053,
          "p80": 5.669,
          "p90": 8.493,
          "p95": 47.017,
          "p99": 53.731,
          "p999": 59.07
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 153.16,
            "max": 162.19,
            "mean": 159.96
          },
          "cpu": {
            "min": 0.33,
            "max": 4.33,
            "mean": 0.88
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 28.28,
            "max": 34.65,
            "mean": 34.01
          },
          "cpu": {
            "min": 0,
            "max": 100.09,
            "mean": 12.12
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37,
      "counters": {
        "benchmark.http_2xx": {
          "value": 156922,
          "perSecond": 5230.73
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
          "value": 24636754,
          "perSecond": 821224.78
        },
        "upstream_cx_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7063155,
          "perSecond": 235438.4
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
          "value": 156959,
          "perSecond": 5231.96
        }
      }
    },
    {
      "testName": "scaling up httproutes to 100 with 20 routes per hostname",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-up",
      "throughput": 5233.77,
      "totalRequests": 157018,
      "latency": {
        "min": 0.392,
        "mean": 6.716,
        "max": 74.948,
        "pstdev": 12.218,
        "percentiles": {
          "p50": 3.086,
          "p75": 4.917,
          "p80": 5.54,
          "p90": 8.473,
          "p95": 48.504,
          "p99": 54.824,
          "p999": 60.995
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 154.81,
            "max": 164.16,
            "mean": 158.99
          },
          "cpu": {
            "min": 0.47,
            "max": 9.67,
            "mean": 2.02
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 34.51,
            "max": 38.95,
            "mean": 38.19
          },
          "cpu": {
            "min": 0,
            "max": 99.98,
            "mean": 12.96
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37,
      "counters": {
        "benchmark.http_2xx": {
          "value": 157018,
          "perSecond": 5233.77
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
          "value": 24651826,
          "perSecond": 821701.88
        },
        "upstream_cx_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7066755,
          "perSecond": 235551.15
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
          "value": 157039,
          "perSecond": 5234.47
        }
      }
    },
    {
      "testName": "scaling up httproutes to 300 with 60 routes per hostname",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-up",
      "throughput": 5200.61,
      "totalRequests": 156022,
      "latency": {
        "min": 0.386,
        "mean": 6.733,
        "max": 87.748,
        "pstdev": 12.052,
        "percentiles": {
          "p50": 3.188,
          "p75": 5.058,
          "p80": 5.703,
          "p90": 8.646,
          "p95": 47.349,
          "p99": 54.261,
          "p999": 62.337
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 178.22,
            "max": 190.05,
            "mean": 185.94
          },
          "cpu": {
            "min": 0.53,
            "max": 63.2,
            "mean": 5.17
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 58.72,
            "max": 59.32,
            "mean": 59.03
          },
          "cpu": {
            "min": 0,
            "max": 100.06,
            "mean": 8.71
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37,
      "counters": {
        "benchmark.http_2xx": {
          "value": 156022,
          "perSecond": 5200.61
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
          "value": 24495454,
          "perSecond": 816495.26
        },
        "upstream_cx_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7022115,
          "perSecond": 234064.8
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
          "value": 156047,
          "perSecond": 5201.44
        }
      }
    },
    {
      "testName": "scaling up httproutes to 500 with 100 routes per hostname",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-up",
      "throughput": 5161.32,
      "totalRequests": 154842,
      "latency": {
        "min": 0.39,
        "mean": 6.838,
        "max": 87.953,
        "pstdev": 12.514,
        "percentiles": {
          "p50": 3.154,
          "p75": 4.94,
          "p80": 5.547,
          "p90": 8.546,
          "p95": 48.75,
          "p99": 55.814,
          "p999": 66.682
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 188.56,
            "max": 205.36,
            "mean": 199.35
          },
          "cpu": {
            "min": 0.2,
            "max": 1,
            "mean": 0.72
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 80.99,
            "max": 81.17,
            "mean": 81.05
          },
          "cpu": {
            "min": 0,
            "max": 100.11,
            "mean": 13.39
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37,
      "counters": {
        "benchmark.http_2xx": {
          "value": 154842,
          "perSecond": 5161.32
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
          "value": 24310194,
          "perSecond": 810327.12
        },
        "upstream_cx_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_cx_tx_bytes_total": {
          "value": 6968970,
          "perSecond": 232295.37
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
          "value": 154866,
          "perSecond": 5162.12
        }
      }
    },
    {
      "testName": "scaling up httproutes to 1000 with 200 routes per hostname",
      "routes": 1000,
      "routesPerHostname": 200,
      "phase": "scaling-up",
      "throughput": 5070.93,
      "totalRequests": 152131,
      "latency": {
        "min": 0.369,
        "mean": 7.157,
        "max": 95.969,
        "pstdev": 13.025,
        "percentiles": {
          "p50": 3.187,
          "p75": 5.196,
          "p80": 5.933,
          "p90": 9.716,
          "p95": 49.264,
          "p99": 58.398,
          "p999": 70.475
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 235.82,
            "max": 250.67,
            "mean": 243.33
          },
          "cpu": {
            "min": 0.2,
            "max": 1.27,
            "mean": 0.81
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 127.18,
            "max": 127.36,
            "mean": 127.27
          },
          "cpu": {
            "min": 0,
            "max": 99.82,
            "mean": 15.31
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38,
      "counters": {
        "benchmark.http_2xx": {
          "value": 152131,
          "perSecond": 5070.93
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
          "value": 23884567,
          "perSecond": 796135.93
        },
        "upstream_cx_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_tx_bytes_total": {
          "value": 6847380,
          "perSecond": 228241.33
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
          "value": 152164,
          "perSecond": 5072.03
        }
      }
    },
    {
      "testName": "scaling down httproutes to 10 with 2 routes per hostname",
      "routes": 10,
      "routesPerHostname": 2,
      "phase": "scaling-down",
      "throughput": 5231.33,
      "totalRequests": 156943,
      "latency": {
        "min": 0.352,
        "mean": 6.847,
        "max": 77.496,
        "pstdev": 12.213,
        "percentiles": {
          "p50": 3.276,
          "p75": 5.1,
          "p80": 5.725,
          "p90": 8.528,
          "p95": 48.064,
          "p99": 54.683,
          "p999": 58.912
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 244.6,
            "max": 250.95,
            "mean": 247.09
          },
          "cpu": {
            "min": 0.8,
            "max": 3.87,
            "mean": 1.17
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 113.82,
            "max": 114,
            "mean": 113.88
          },
          "cpu": {
            "min": 0,
            "max": 99.92,
            "mean": 6.36
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38,
      "counters": {
        "benchmark.http_2xx": {
          "value": 156943,
          "perSecond": 5231.33
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
          "value": 24640051,
          "perSecond": 821318.13
        },
        "upstream_cx_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7063740,
          "perSecond": 235453.15
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
          "value": 156972,
          "perSecond": 5232.29
        }
      }
    },
    {
      "testName": "scaling down httproutes to 50 with 10 routes per hostname",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-down",
      "throughput": 5261.75,
      "totalRequests": 157855,
      "latency": {
        "min": 0.399,
        "mean": 6.732,
        "max": 77.856,
        "pstdev": 12.3,
        "percentiles": {
          "p50": 3.112,
          "p75": 4.945,
          "p80": 5.568,
          "p90": 8.439,
          "p95": 48.916,
          "p99": 54.951,
          "p999": 60.592
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 224.61,
            "max": 255.07,
            "mean": 250.77
          },
          "cpu": {
            "min": 0.73,
            "max": 7.13,
            "mean": 1.49
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 113.8,
            "max": 113.98,
            "mean": 113.86
          },
          "cpu": {
            "min": 0,
            "max": 94.87,
            "mean": 7.39
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37,
      "counters": {
        "benchmark.http_2xx": {
          "value": 157855,
          "perSecond": 5261.75
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
          "value": 24783235,
          "perSecond": 826094.04
        },
        "upstream_cx_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7104870,
          "perSecond": 236825.04
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
          "value": 157886,
          "perSecond": 5262.78
        }
      }
    },
    {
      "testName": "scaling down httproutes to 100 with 20 routes per hostname",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-down",
      "throughput": 5237.44,
      "totalRequests": 157127,
      "latency": {
        "min": 0.359,
        "mean": 6.712,
        "max": 85.757,
        "pstdev": 12.224,
        "percentiles": {
          "p50": 3.131,
          "p75": 4.899,
          "p80": 5.53,
          "p90": 8.45,
          "p95": 48.3,
          "p99": 55.074,
          "p999": 61.05
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 172.34,
            "max": 224.61,
            "mean": 184.21
          },
          "cpu": {
            "min": 0.67,
            "max": 65.47,
            "mean": 5.97
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 113.81,
            "max": 114,
            "mean": 113.88
          },
          "cpu": {
            "min": 0,
            "max": 99.95,
            "mean": 9.23
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37,
      "counters": {
        "benchmark.http_2xx": {
          "value": 157127,
          "perSecond": 5237.44
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
          "value": 24668939,
          "perSecond": 822278.65
        },
        "upstream_cx_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7071930,
          "perSecond": 235725.46
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
          "value": 157154,
          "perSecond": 5238.34
        }
      }
    },
    {
      "testName": "scaling down httproutes to 300 with 60 routes per hostname",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-down",
      "throughput": 5179.69,
      "totalRequests": 155395,
      "latency": {
        "min": 0.398,
        "mean": 6.83,
        "max": 80.662,
        "pstdev": 12.28,
        "percentiles": {
          "p50": 3.2,
          "p75": 5.016,
          "p80": 5.664,
          "p90": 8.694,
          "p95": 48.179,
          "p99": 55.212,
          "p999": 63.549
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 179.46,
            "max": 205.75,
            "mean": 190.6
          },
          "cpu": {
            "min": 0.6,
            "max": 62.6,
            "mean": 3.68
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 127.33,
            "max": 127.48,
            "mean": 127.37
          },
          "cpu": {
            "min": 0,
            "max": 90.4,
            "mean": 7.51
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37,
      "counters": {
        "benchmark.http_2xx": {
          "value": 155395,
          "perSecond": 5179.69
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
          "value": 24397015,
          "perSecond": 813211.27
        },
        "upstream_cx_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_cx_tx_bytes_total": {
          "value": 6993540,
          "perSecond": 233111.53
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
          "value": 155412,
          "perSecond": 5180.26
        }
      }
    },
    {
      "testName": "scaling down httproutes to 500 with 100 routes per hostname",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-down",
      "throughput": 5186.01,
      "totalRequests": 155581,
      "latency": {
        "min": 0.38,
        "mean": 6.971,
        "max": 107.388,
        "pstdev": 12.508,
        "percentiles": {
          "p50": 3.261,
          "p75": 5.13,
          "p80": 5.781,
          "p90": 8.858,
          "p95": 48.461,
          "p99": 55.611,
          "p999": 67.166
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 197.3,
            "max": 213.13,
            "mean": 208.59
          },
          "cpu": {
            "min": 0.33,
            "max": 69.58,
            "mean": 2.16
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 127.33,
            "max": 127.53,
            "mean": 127.38
          },
          "cpu": {
            "min": 0,
            "max": 93.66,
            "mean": 10.66
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38,
      "counters": {
        "benchmark.http_2xx": {
          "value": 155581,
          "perSecond": 5186.01
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
          "value": 24426217,
          "perSecond": 814203.16
        },
        "upstream_cx_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7002855,
          "perSecond": 233427.33
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
          "value": 155619,
          "perSecond": 5187.27
        }
      }
    }
  ]
};

export default benchmarkData;
