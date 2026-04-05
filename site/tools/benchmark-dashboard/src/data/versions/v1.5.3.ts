import { TestSuite } from '../types';

// Benchmark data extracted from markdown report for version 1.5.3
// Generated on 2025-11-21T02:42:14.480Z

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.5.3",
    "runId": "1.5.3-1763692934477",
    "date": "2025-11-21T02:42:14.477Z",
    "environment": "GitHub CI",
    "description": "Benchmark results for version 1.5.3",
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
      "throughput": 5307.12,
      "totalRequests": 159214,
      "latency": {
        "min": 0.355,
        "mean": 8.822,
        "max": 84.43,
        "pstdev": 13.817,
        "percentiles": {
          "p50": 4.302,
          "p75": 7.114,
          "p80": 8.131,
          "p90": 15.852,
          "p95": 50.37,
          "p99": 56.598,
          "p999": 62.789
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 108.98,
            "max": 135.71,
            "mean": 131.81
          },
          "cpu": {
            "min": 0.13,
            "max": 0.93,
            "mean": 0.44
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 0,
            "max": 27.21,
            "mean": 22.38
          },
          "cpu": {
            "min": 0,
            "max": 100.07,
            "mean": 7.4
          }
        }
      },
      "poolOverflow": 350,
      "upstreamConnections": 50,
      "counters": {
        "benchmark.http_2xx": {
          "value": 159214,
          "perSecond": 5307.12
        },
        "benchmark.pool_overflow": {
          "value": 350,
          "perSecond": 11.67
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
          "value": 50,
          "perSecond": 1.67
        },
        "upstream_cx_rx_bytes_total": {
          "value": 24996598,
          "perSecond": 833217.57
        },
        "upstream_cx_total": {
          "value": 50,
          "perSecond": 1.67
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7166880,
          "perSecond": 238895.32
        },
        "upstream_rq_pending_overflow": {
          "value": 350,
          "perSecond": 11.67
        },
        "upstream_rq_pending_total": {
          "value": 50,
          "perSecond": 1.67
        },
        "upstream_rq_total": {
          "value": 159264,
          "perSecond": 5308.78
        }
      }
    },
    {
      "testName": "scaling up httproutes to 50 with 10 routes per hostname",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-up",
      "throughput": 5208.91,
      "totalRequests": 156270,
      "latency": {
        "min": 0.362,
        "mean": 6.755,
        "max": 82.063,
        "pstdev": 12.17,
        "percentiles": {
          "p50": 3.138,
          "p75": 4.998,
          "p80": 5.665,
          "p90": 8.703,
          "p95": 48.013,
          "p99": 54.568,
          "p999": 60.248
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 135.71,
            "max": 146.09,
            "mean": 142.6
          },
          "cpu": {
            "min": 0.33,
            "max": 3.93,
            "mean": 0.8
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 26.76,
            "max": 30.96,
            "mean": 30.48
          },
          "cpu": {
            "min": 0,
            "max": 99.94,
            "mean": 4.65
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37,
      "counters": {
        "benchmark.http_2xx": {
          "value": 156270,
          "perSecond": 5208.91
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
          "value": 24534390,
          "perSecond": 817799.54
        },
        "upstream_cx_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7033635,
          "perSecond": 234450.64
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
          "value": 156303,
          "perSecond": 5210.01
        }
      }
    },
    {
      "testName": "scaling up httproutes to 100 with 20 routes per hostname",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-up",
      "throughput": 5217.66,
      "totalRequests": 156530,
      "latency": {
        "min": 0.386,
        "mean": 7.652,
        "max": 99.11,
        "pstdev": 13.119,
        "percentiles": {
          "p50": 3.49,
          "p75": 5.773,
          "p80": 6.596,
          "p90": 11.249,
          "p95": 49.59,
          "p99": 56.573,
          "p999": 64.243
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 141.46,
            "max": 152.33,
            "mean": 150.33
          },
          "cpu": {
            "min": 0.4,
            "max": 5.93,
            "mean": 1.05
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 30.82,
            "max": 37.48,
            "mean": 36.09
          },
          "cpu": {
            "min": 0,
            "max": 92.26,
            "mean": 1.94
          }
        }
      },
      "poolOverflow": 358,
      "upstreamConnections": 42,
      "counters": {
        "benchmark.http_2xx": {
          "value": 156530,
          "perSecond": 5217.66
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
          "value": 24575210,
          "perSecond": 819172.58
        },
        "upstream_cx_total": {
          "value": 42,
          "perSecond": 1.4
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7045740,
          "perSecond": 234857.69
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
          "value": 156572,
          "perSecond": 5219.06
        }
      }
    },
    {
      "testName": "scaling up httproutes to 300 with 60 routes per hostname",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-up",
      "throughput": 5120.99,
      "totalRequests": 153632,
      "latency": {
        "min": 0.363,
        "mean": 7.027,
        "max": 92.614,
        "pstdev": 12.387,
        "percentiles": {
          "p50": 3.32,
          "p75": 5.314,
          "p80": 6.018,
          "p90": 9.487,
          "p95": 48.113,
          "p99": 54.697,
          "p999": 63.485
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 146.21,
            "max": 154.59,
            "mean": 152.49
          },
          "cpu": {
            "min": 0.4,
            "max": 22.79,
            "mean": 3.29
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 36.96,
            "max": 59.15,
            "mean": 54.38
          },
          "cpu": {
            "min": 0,
            "max": 1.28,
            "mean": 0.03
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38,
      "counters": {
        "benchmark.http_2xx": {
          "value": 153632,
          "perSecond": 5120.99
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
          "value": 24120224,
          "perSecond": 803994.95
        },
        "upstream_cx_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_tx_bytes_total": {
          "value": 6915015,
          "perSecond": 230496.91
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
          "value": 153667,
          "perSecond": 5122.15
        }
      }
    },
    {
      "testName": "scaling up httproutes to 500 with 100 routes per hostname",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-up",
      "throughput": 5080.76,
      "totalRequests": 152425,
      "latency": {
        "min": 0.375,
        "mean": 7.037,
        "max": 95.44,
        "pstdev": 12.408,
        "percentiles": {
          "p50": 3.346,
          "p75": 5.267,
          "p80": 5.929,
          "p90": 9.267,
          "p95": 48.084,
          "p99": 55.326,
          "p999": 65.15
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 152.76,
            "max": 168.61,
            "mean": 164.59
          },
          "cpu": {
            "min": 0.47,
            "max": 22.2,
            "mean": 2.76
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 59.11,
            "max": 79.46,
            "mean": 77.33
          },
          "cpu": {
            "min": 0,
            "max": 100.07,
            "mean": 8.55
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38,
      "counters": {
        "benchmark.http_2xx": {
          "value": 152425,
          "perSecond": 5080.76
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
          "value": 23930725,
          "perSecond": 797679.97
        },
        "upstream_cx_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_tx_bytes_total": {
          "value": 6860790,
          "perSecond": 228689.89
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
          "value": 152462,
          "perSecond": 5082
        }
      }
    },
    {
      "testName": "scaling up httproutes to 1000 with 200 routes per hostname",
      "routes": 1000,
      "routesPerHostname": 200,
      "phase": "scaling-up",
      "throughput": 5007.79,
      "totalRequests": 150239,
      "latency": {
        "min": 0.379,
        "mean": 7.327,
        "max": 121.139,
        "pstdev": 12.913,
        "percentiles": {
          "p50": 3.33,
          "p75": 5.423,
          "p80": 6.152,
          "p90": 10.043,
          "p95": 48.586,
          "p99": 56.95,
          "p999": 69.328
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 162.51,
            "max": 197.73,
            "mean": 192.58
          },
          "cpu": {
            "min": 0.47,
            "max": 54.6,
            "mean": 7.15
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 99.33,
            "max": 135.09,
            "mean": 131.17
          },
          "cpu": {
            "min": 0,
            "max": 47.18,
            "mean": 5.45
          }
        }
      },
      "poolOverflow": 361,
      "upstreamConnections": 39,
      "counters": {
        "benchmark.http_2xx": {
          "value": 150239,
          "perSecond": 5007.79
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
          "value": 23587523,
          "perSecond": 786223.43
        },
        "upstream_cx_total": {
          "value": 39,
          "perSecond": 1.3
        },
        "upstream_cx_tx_bytes_total": {
          "value": 6761970,
          "perSecond": 225391.16
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
          "value": 150266,
          "perSecond": 5008.69
        }
      }
    },
    {
      "testName": "scaling down httproutes to 10 with 2 routes per hostname",
      "routes": 10,
      "routesPerHostname": 2,
      "phase": "scaling-down",
      "throughput": 5202.6,
      "totalRequests": 156078,
      "latency": {
        "min": 0.396,
        "mean": 6.739,
        "max": 69.812,
        "pstdev": 12.158,
        "percentiles": {
          "p50": 3.166,
          "p75": 5.022,
          "p80": 5.647,
          "p90": 8.593,
          "p95": 48.123,
          "p99": 54.417,
          "p999": 59.23
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 137.63,
            "max": 148.54,
            "mean": 142.77
          },
          "cpu": {
            "min": 0.6,
            "max": 3.27,
            "mean": 0.98
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 121.41,
            "max": 121.58,
            "mean": 121.45
          },
          "cpu": {
            "min": 0,
            "max": 98.16,
            "mean": 12.67
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37,
      "counters": {
        "benchmark.http_2xx": {
          "value": 156078,
          "perSecond": 5202.6
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
          "value": 24504246,
          "perSecond": 816807.83
        },
        "upstream_cx_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7025175,
          "perSecond": 234172.39
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
          "value": 156115,
          "perSecond": 5203.83
        }
      }
    },
    {
      "testName": "scaling down httproutes to 50 with 10 routes per hostname",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-down",
      "throughput": 5206.36,
      "totalRequests": 156191,
      "latency": {
        "min": 0.374,
        "mean": 6.733,
        "max": 75.362,
        "pstdev": 12.051,
        "percentiles": {
          "p50": 3.238,
          "p75": 5.085,
          "p80": 5.693,
          "p90": 8.505,
          "p95": 48.011,
          "p99": 54.202,
          "p999": 59.555
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 144.43,
            "max": 147.82,
            "mean": 145.9
          },
          "cpu": {
            "min": 0.6,
            "max": 4.27,
            "mean": 1.06
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 121.41,
            "max": 121.54,
            "mean": 121.45
          },
          "cpu": {
            "min": 0,
            "max": 96.17,
            "mean": 7.55
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37,
      "counters": {
        "benchmark.http_2xx": {
          "value": 156191,
          "perSecond": 5206.36
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
          "value": 24521987,
          "perSecond": 817398.06
        },
        "upstream_cx_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7030260,
          "perSecond": 234341.57
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
          "value": 156228,
          "perSecond": 5207.59
        }
      }
    },
    {
      "testName": "scaling down httproutes to 100 with 20 routes per hostname",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-down",
      "throughput": 5249.49,
      "totalRequests": 157487,
      "latency": {
        "min": 0.372,
        "mean": 6.913,
        "max": 72.269,
        "pstdev": 12.304,
        "percentiles": {
          "p50": 3.269,
          "p75": 5.099,
          "p80": 5.722,
          "p90": 8.753,
          "p95": 48.592,
          "p99": 54.806,
          "p999": 61.022
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 143.72,
            "max": 159.16,
            "mean": 149.43
          },
          "cpu": {
            "min": 0.6,
            "max": 13.4,
            "mean": 2.26
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 121.41,
            "max": 135.12,
            "mean": 123.41
          },
          "cpu": {
            "min": 0,
            "max": 100.04,
            "mean": 4.96
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38,
      "counters": {
        "benchmark.http_2xx": {
          "value": 157487,
          "perSecond": 5249.49
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
          "value": 24725459,
          "perSecond": 824169.93
        },
        "upstream_cx_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7087770,
          "perSecond": 236255.55
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
          "value": 157506,
          "perSecond": 5250.12
        }
      }
    },
    {
      "testName": "scaling down httproutes to 300 with 60 routes per hostname",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-down",
      "throughput": 5094.86,
      "totalRequests": 152846,
      "latency": {
        "min": 0.37,
        "mean": 6.845,
        "max": 83.345,
        "pstdev": 12.159,
        "percentiles": {
          "p50": 3.146,
          "p75": 5.469,
          "p80": 6.243,
          "p90": 9.495,
          "p95": 47.437,
          "p99": 54.419,
          "p999": 64.133
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 151.69,
            "max": 164.31,
            "mean": 158.32
          },
          "cpu": {
            "min": 0.67,
            "max": 14.33,
            "mean": 2.06
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 135.06,
            "max": 135.27,
            "mean": 135.13
          },
          "cpu": {
            "min": 0,
            "max": 83.56,
            "mean": 4.59
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37,
      "counters": {
        "benchmark.http_2xx": {
          "value": 152846,
          "perSecond": 5094.86
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
          "value": 23996822,
          "perSecond": 799893.07
        },
        "upstream_cx_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_cx_tx_bytes_total": {
          "value": 6879735,
          "perSecond": 229324.21
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
          "value": 152883,
          "perSecond": 5096.09
        }
      }
    },
    {
      "testName": "scaling down httproutes to 500 with 100 routes per hostname",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-down",
      "throughput": 5119.62,
      "totalRequests": 153596,
      "latency": {
        "min": 0.38,
        "mean": 6.992,
        "max": 88.113,
        "pstdev": 12.19,
        "percentiles": {
          "p50": 3.336,
          "p75": 5.257,
          "p80": 5.945,
          "p90": 9.203,
          "p95": 47.157,
          "p99": 54.669,
          "p999": 65.861
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 162.61,
            "max": 197.09,
            "mean": 168.15
          },
          "cpu": {
            "min": 0.53,
            "max": 33.8,
            "mean": 3.91
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 135.07,
            "max": 135.25,
            "mean": 135.12
          },
          "cpu": {
            "min": 0,
            "max": 63.39,
            "mean": 1.52
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38,
      "counters": {
        "benchmark.http_2xx": {
          "value": 153596,
          "perSecond": 5119.62
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
          "value": 24114572,
          "perSecond": 803779.88
        },
        "upstream_cx_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_tx_bytes_total": {
          "value": 6913125,
          "perSecond": 230426.27
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
          "value": 153625,
          "perSecond": 5120.58
        }
      }
    }
  ]
};

export default benchmarkData;
