import { TestSuite } from '../types';

// Benchmark data extracted from markdown report for version 1.6.0
// Generated on 2025-11-21T02:48:35.722Z

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.6.0",
    "runId": "1.6.0-1763693315721",
    "date": "2025-11-21T02:48:35.721Z",
    "environment": "production",
    "description": "Benchmark results for version 1.6.0",
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
      "throughput": 5405.78,
      "totalRequests": 162183,
      "latency": {
        "min": 371,
        "mean": 6510,
        "max": 71950,
        "pstdev": 11469,
        "percentiles": {
          "p50": 3176,
          "p75": 5009,
          "p80": 5637,
          "p90": 8605,
          "p95": 45916,
          "p99": 53184,
          "p999": 58396
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 117.88,
            "max": 139.78,
            "mean": 136.88
          },
          "cpu": {
            "min": 0,
            "max": 0.8,
            "mean": 0.47
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 22.73,
            "max": 27.1,
            "mean": 26.67
          },
          "cpu": {
            "min": 0,
            "max": 60.52,
            "mean": 31.28
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37,
      "counters": {
        "benchmark.http_2xx": {
          "value": 162183,
          "perSecond": 5405.78
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
          "value": 25462731,
          "perSecond": 848706.75
        },
        "upstream_cx_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7299720,
          "perSecond": 243309.39
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
          "value": 162216,
          "perSecond": 5406.88
        }
      }
    },
    {
      "testName": "scaling up httproutes to 50 with 10 routes per hostname",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-up",
      "throughput": 5321.53,
      "totalRequests": 159646,
      "latency": {
        "min": 380,
        "mean": 6965,
        "max": 74780,
        "pstdev": 12324,
        "percentiles": {
          "p50": 3279,
          "p75": 5212,
          "p80": 5857,
          "p90": 8962,
          "p95": 48390,
          "p99": 54839,
          "p999": 60753
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 137.93,
            "max": 147.15,
            "mean": 144.92
          },
          "cpu": {
            "min": 0.77,
            "max": 1.97,
            "mean": 1.13
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 31,
            "max": 33.25,
            "mean": 32.95
          },
          "cpu": {
            "min": 0.68,
            "max": 48.9,
            "mean": 24.82
          }
        }
      },
      "poolOverflow": 361,
      "upstreamConnections": 39,
      "counters": {
        "benchmark.http_2xx": {
          "value": 159646,
          "perSecond": 5321.53
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
          "value": 25064422,
          "perSecond": 835479.94
        },
        "upstream_cx_total": {
          "value": 39,
          "perSecond": 1.3
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7185825,
          "perSecond": 239527.27
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
          "value": 159685,
          "perSecond": 5322.83
        }
      }
    },
    {
      "testName": "scaling up httproutes to 100 with 20 routes per hostname",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-up",
      "throughput": 5174.45,
      "totalRequests": 155239,
      "latency": {
        "min": 391,
        "mean": 6953,
        "max": 75833,
        "pstdev": 12267,
        "percentiles": {
          "p50": 3289,
          "p75": 5252,
          "p80": 5923,
          "p90": 9042,
          "p95": 48228,
          "p99": 54372,
          "p999": 61143
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 142.65,
            "max": 148.24,
            "mean": 145.29
          },
          "cpu": {
            "min": 1.02,
            "max": 2.42,
            "mean": 1.5
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 37.06,
            "max": 39.35,
            "mean": 38.85
          },
          "cpu": {
            "min": 0.76,
            "max": 38.87,
            "mean": 22.23
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38,
      "counters": {
        "benchmark.http_2xx": {
          "value": 155239,
          "perSecond": 5174.45
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
          "value": 24372523,
          "perSecond": 812388.87
        },
        "upstream_cx_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_tx_bytes_total": {
          "value": 6987240,
          "perSecond": 232899.81
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
          "value": 155272,
          "perSecond": 5175.55
        }
      }
    },
    {
      "testName": "scaling up httproutes to 300 with 60 routes per hostname",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-up",
      "throughput": 5137.17,
      "totalRequests": 154122,
      "latency": {
        "min": 374,
        "mean": 7227,
        "max": 98656,
        "pstdev": 12782,
        "percentiles": {
          "p50": 3366,
          "p75": 5364,
          "p80": 6057,
          "p90": 9492,
          "p95": 49248,
          "p99": 55939,
          "p999": 66674
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 145.75,
            "max": 155.92,
            "mean": 154.26
          },
          "cpu": {
            "min": 2.78,
            "max": 4.03,
            "mean": 3.35
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 61.35,
            "max": 61.58,
            "mean": 61.46
          },
          "cpu": {
            "min": 1.67,
            "max": 15.27,
            "mean": 10.25
          }
        }
      },
      "poolOverflow": 361,
      "upstreamConnections": 39,
      "counters": {
        "benchmark.http_2xx": {
          "value": 154122,
          "perSecond": 5137.17
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
          "value": 24197154,
          "perSecond": 806535.05
        },
        "upstream_cx_total": {
          "value": 39,
          "perSecond": 1.3
        },
        "upstream_cx_tx_bytes_total": {
          "value": 6936210,
          "perSecond": 231196.47
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
          "value": 154138,
          "perSecond": 5137.7
        }
      }
    },
    {
      "testName": "scaling up httproutes to 500 with 100 routes per hostname",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-up",
      "throughput": 5143.02,
      "totalRequests": 154291,
      "latency": {
        "min": 377,
        "mean": 7218,
        "max": 90238,
        "pstdev": 12857,
        "percentiles": {
          "p50": 3304,
          "p75": 5256,
          "p80": 5932,
          "p90": 9426,
          "p95": 49238,
          "p99": 56430,
          "p999": 67078
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 159.84,
            "max": 171.14,
            "mean": 168.88
          },
          "cpu": {
            "min": 3.95,
            "max": 6.35,
            "mean": 5.04
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 83.48,
            "max": 85.93,
            "mean": 85.44
          },
          "cpu": {
            "min": 2.8,
            "max": 14.06,
            "mean": 11.15
          }
        }
      },
      "poolOverflow": 361,
      "upstreamConnections": 39,
      "counters": {
        "benchmark.http_2xx": {
          "value": 154291,
          "perSecond": 5143.02
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
          "value": 24223687,
          "perSecond": 807454.86
        },
        "upstream_cx_total": {
          "value": 39,
          "perSecond": 1.3
        },
        "upstream_cx_tx_bytes_total": {
          "value": 6944850,
          "perSecond": 231494.61
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
          "value": 154330,
          "perSecond": 5144.32
        }
      }
    },
    {
      "testName": "scaling up httproutes to 1000 with 200 routes per hostname",
      "routes": 1000,
      "routesPerHostname": 200,
      "phase": "scaling-up",
      "throughput": 5030.53,
      "totalRequests": 150916,
      "latency": {
        "min": 388,
        "mean": 7148,
        "max": 94916,
        "pstdev": 12846,
        "percentiles": {
          "p50": 3256,
          "p75": 5277,
          "p80": 5991,
          "p90": 9554,
          "p95": 48789,
          "p99": 58165,
          "p999": 70217
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 182.29,
            "max": 190.1,
            "mean": 188.54
          },
          "cpu": {
            "min": 8.38,
            "max": 10.53,
            "mean": 9.4
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 139.7,
            "max": 140.13,
            "mean": 139.81
          },
          "cpu": {
            "min": 4.34,
            "max": 9.59,
            "mean": 8.24
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38,
      "counters": {
        "benchmark.http_2xx": {
          "value": 150916,
          "perSecond": 5030.53
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
          "value": 23693812,
          "perSecond": 789792.97
        },
        "upstream_cx_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_tx_bytes_total": {
          "value": 6792930,
          "perSecond": 226430.78
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
          "value": 150954,
          "perSecond": 5031.8
        }
      }
    },
    {
      "testName": "scaling down httproutes to 10 with 2 routes per hostname",
      "routes": 10,
      "routesPerHostname": 2,
      "phase": "scaling-down",
      "throughput": 5159.79,
      "totalRequests": 154794,
      "latency": {
        "min": 387,
        "mean": 7884,
        "max": 88625,
        "pstdev": 12967,
        "percentiles": {
          "p50": 3761,
          "p75": 6219,
          "p80": 7083,
          "p90": 12018,
          "p95": 48920,
          "p99": 55654,
          "p999": 62513
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 142.16,
            "max": 154.78,
            "mean": 146.94
          },
          "cpu": {
            "min": 0,
            "max": 1.27,
            "mean": 0.89
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 133.71,
            "max": 133.92,
            "mean": 133.77
          },
          "cpu": {
            "min": 0,
            "max": 70.81,
            "mean": 19.99
          }
        }
      },
      "poolOverflow": 357,
      "upstreamConnections": 43,
      "counters": {
        "benchmark.http_2xx": {
          "value": 154794,
          "perSecond": 5159.79
        },
        "benchmark.pool_overflow": {
          "value": 357,
          "perSecond": 11.9
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
          "value": 43,
          "perSecond": 1.43
        },
        "upstream_cx_rx_bytes_total": {
          "value": 24302658,
          "perSecond": 810086.94
        },
        "upstream_cx_total": {
          "value": 43,
          "perSecond": 1.43
        },
        "upstream_cx_tx_bytes_total": {
          "value": 6967665,
          "perSecond": 232255.02
        },
        "upstream_rq_pending_overflow": {
          "value": 357,
          "perSecond": 11.9
        },
        "upstream_rq_pending_total": {
          "value": 43,
          "perSecond": 1.43
        },
        "upstream_rq_total": {
          "value": 154837,
          "perSecond": 5161.22
        }
      }
    },
    {
      "testName": "scaling down httproutes to 50 with 10 routes per hostname",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-down",
      "throughput": 5250.83,
      "totalRequests": 157525,
      "latency": {
        "min": 373,
        "mean": 7073,
        "max": 83021,
        "pstdev": 12387,
        "percentiles": {
          "p50": 3327,
          "p75": 5364,
          "p80": 6068,
          "p90": 9392,
          "p95": 48283,
          "p99": 54970,
          "p999": 61716
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 145.25,
            "max": 154.5,
            "mean": 149.3
          },
          "cpu": {
            "min": 0,
            "max": 1.33,
            "mean": 0.83
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 133.71,
            "max": 134.11,
            "mean": 133.78
          },
          "cpu": {
            "min": 0,
            "max": 100,
            "mean": 26.95
          }
        }
      },
      "poolOverflow": 361,
      "upstreamConnections": 39,
      "counters": {
        "benchmark.http_2xx": {
          "value": 157525,
          "perSecond": 5250.83
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
          "value": 24731425,
          "perSecond": 824380.14
        },
        "upstream_cx_total": {
          "value": 39,
          "perSecond": 1.3
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7090380,
          "perSecond": 236345.8
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
          "value": 157564,
          "perSecond": 5252.13
        }
      }
    },
    {
      "testName": "scaling down httproutes to 100 with 20 routes per hostname",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-down",
      "throughput": 5098.66,
      "totalRequests": 152960,
      "latency": {
        "min": 381,
        "mean": 7047,
        "max": 85135,
        "pstdev": 12463,
        "percentiles": {
          "p50": 3247,
          "p75": 5250,
          "p80": 5945,
          "p90": 9435,
          "p95": 48533,
          "p99": 55283,
          "p999": 61874
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 149.52,
            "max": 166.6,
            "mean": 154.82
          },
          "cpu": {
            "min": 0,
            "max": 3.93,
            "mean": 1.66
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 133.69,
            "max": 133.83,
            "mean": 133.75
          },
          "cpu": {
            "min": 0,
            "max": 99.94,
            "mean": 33.64
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38,
      "counters": {
        "benchmark.http_2xx": {
          "value": 152960,
          "perSecond": 5098.66
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
          "value": 24014720,
          "perSecond": 800489.45
        },
        "upstream_cx_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_tx_bytes_total": {
          "value": 6884910,
          "perSecond": 229496.65
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
          "value": 152998,
          "perSecond": 5099.93
        }
      }
    },
    {
      "testName": "scaling down httproutes to 300 with 60 routes per hostname",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-down",
      "throughput": 5184.86,
      "totalRequests": 155546,
      "latency": {
        "min": 381,
        "mean": 6806,
        "max": 122974,
        "pstdev": 12476,
        "percentiles": {
          "p50": 3071,
          "p75": 4981,
          "p80": 5673,
          "p90": 9101,
          "p95": 48928,
          "p99": 55502,
          "p999": 65976
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 162.01,
            "max": 173.05,
            "mean": 165.19
          },
          "cpu": {
            "min": 0,
            "max": 1.13,
            "mean": 0.8
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 133.27,
            "max": 139.72,
            "mean": 134.05
          },
          "cpu": {
            "min": 0,
            "max": 88.97,
            "mean": 36.48
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37,
      "counters": {
        "benchmark.http_2xx": {
          "value": 155546,
          "perSecond": 5184.86
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
          "value": 24420722,
          "perSecond": 814023.45
        },
        "upstream_cx_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7001235,
          "perSecond": 233374.32
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
          "value": 155583,
          "perSecond": 5186.1
        }
      }
    },
    {
      "testName": "scaling down httproutes to 500 with 100 routes per hostname",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-down",
      "throughput": 5183.69,
      "totalRequests": 155511,
      "latency": {
        "min": 382,
        "mean": 6773,
        "max": 81068,
        "pstdev": 12373,
        "percentiles": {
          "p50": 3103,
          "p75": 4965,
          "p80": 5616,
          "p90": 8698,
          "p95": 48418,
          "p99": 55670,
          "p999": 65275
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 167.8,
            "max": 212.05,
            "mean": 174.66
          },
          "cpu": {
            "min": 0,
            "max": 3,
            "mean": 1.41
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 139.71,
            "max": 139.9,
            "mean": 139.78
          },
          "cpu": {
            "min": 0,
            "max": 79.69,
            "mean": 20.46
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37,
      "counters": {
        "benchmark.http_2xx": {
          "value": 155511,
          "perSecond": 5183.69
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
          "value": 24415227,
          "perSecond": 813839.77
        },
        "upstream_cx_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_cx_tx_bytes_total": {
          "value": 6999660,
          "perSecond": 233321.67
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
          "value": 155548,
          "perSecond": 5184.93
        }
      }
    }
  ]
};

export default benchmarkData;
