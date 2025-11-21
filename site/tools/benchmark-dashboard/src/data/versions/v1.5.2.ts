import { TestSuite } from '../types';

// Benchmark data extracted from markdown report for version 1.5.2
// Generated on 2025-11-21T02:42:01.167Z

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.5.2",
    "runId": "1.5.2-1763692921166",
    "date": "2025-11-21T02:42:01.166Z",
    "environment": "production",
    "description": "Benchmark results for version 1.5.2",
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
      "throughput": 5479.08,
      "totalRequests": 164373,
      "latency": {
        "min": 366,
        "mean": 6532,
        "max": 64706,
        "pstdev": 11722,
        "percentiles": {
          "p50": 3191,
          "p75": 4876,
          "p80": 5443,
          "p90": 7921,
          "p95": 46766,
          "p99": 52940,
          "p999": 57231
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 110.08,
            "max": 139.07,
            "mean": 132.64
          },
          "cpu": {
            "min": 0.2,
            "max": 0.93,
            "mean": 0.4
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 0,
            "max": 28.59,
            "mean": 23.78
          },
          "cpu": {
            "min": 0,
            "max": 99.89,
            "mean": 2.4
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38,
      "counters": {
        "benchmark.http_2xx": {
          "value": 164373,
          "perSecond": 5479.08
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
          "value": 25806561,
          "perSecond": 860215.71
        },
        "upstream_cx_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7398495,
          "perSecond": 246615.64
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
          "value": 164411,
          "perSecond": 5480.35
        }
      }
    },
    {
      "testName": "scaling up httproutes to 50 with 10 routes per hostname",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-up",
      "throughput": 5461.8,
      "totalRequests": 163860,
      "latency": {
        "min": 384,
        "mean": 7323,
        "max": 82927,
        "pstdev": 12697,
        "percentiles": {
          "p50": 3499,
          "p75": 5538,
          "p80": 6239,
          "p90": 9715,
          "p95": 49223,
          "p99": 55181,
          "p999": 60008
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 131.13,
            "max": 142.86,
            "mean": 141.65
          },
          "cpu": {
            "min": 0.27,
            "max": 3.87,
            "mean": 0.86
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 28.2,
            "max": 34.74,
            "mean": 33.02
          },
          "cpu": {
            "min": 0,
            "max": 41.44,
            "mean": 1.79
          }
        }
      },
      "poolOverflow": 358,
      "upstreamConnections": 42,
      "counters": {
        "benchmark.http_2xx": {
          "value": 163860,
          "perSecond": 5461.8
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
          "value": 25726020,
          "perSecond": 857502.4
        },
        "upstream_cx_total": {
          "value": 42,
          "perSecond": 1.4
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7375050,
          "perSecond": 245825.94
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
          "value": 163890,
          "perSecond": 5462.8
        }
      }
    },
    {
      "testName": "scaling up httproutes to 100 with 20 routes per hostname",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-up",
      "throughput": 5433.86,
      "totalRequests": 163016,
      "latency": {
        "min": 367,
        "mean": 6464,
        "max": 72536,
        "pstdev": 11843,
        "percentiles": {
          "p50": 3031,
          "p75": 4849,
          "p80": 5481,
          "p90": 8285,
          "p95": 47396,
          "p99": 54114,
          "p999": 59860
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 137.43,
            "max": 141.93,
            "mean": 140.64
          },
          "cpu": {
            "min": 0.33,
            "max": 5.67,
            "mean": 0.99
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 34.36,
            "max": 38.81,
            "mean": 38.03
          },
          "cpu": {
            "min": 0,
            "max": 99.75,
            "mean": 9.06
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37,
      "counters": {
        "benchmark.http_2xx": {
          "value": 163016,
          "perSecond": 5433.86
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
          "value": 25593512,
          "perSecond": 853116.28
        },
        "upstream_cx_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7337385,
          "perSecond": 244579.28
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
          "value": 163053,
          "perSecond": 5435.1
        }
      }
    },
    {
      "testName": "scaling up httproutes to 300 with 60 routes per hostname",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-up",
      "throughput": 5397.08,
      "totalRequests": 161912,
      "latency": {
        "min": 380,
        "mean": 6525,
        "max": 93134,
        "pstdev": 11997,
        "percentiles": {
          "p50": 3053,
          "p75": 4818,
          "p80": 5400,
          "p90": 8070,
          "p95": 47677,
          "p99": 54564,
          "p999": 64092
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 141.93,
            "max": 154.43,
            "mean": 152.12
          },
          "cpu": {
            "min": 0.4,
            "max": 19.33,
            "mean": 2.39
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 38.43,
            "max": 60.61,
            "mean": 58.95
          },
          "cpu": {
            "min": 0,
            "max": 92.48,
            "mean": 10.62
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37,
      "counters": {
        "benchmark.http_2xx": {
          "value": 161912,
          "perSecond": 5397.08
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
          "value": 25420184,
          "perSecond": 847341.13
        },
        "upstream_cx_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7287660,
          "perSecond": 242922.48
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
          "value": 161948,
          "perSecond": 5398.28
        }
      }
    },
    {
      "testName": "scaling up httproutes to 500 with 100 routes per hostname",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-up",
      "throughput": 5346.58,
      "totalRequests": 160402,
      "latency": {
        "min": 373,
        "mean": 6790,
        "max": 103403,
        "pstdev": 12301,
        "percentiles": {
          "p50": 3153,
          "p75": 4996,
          "p80": 5626,
          "p90": 8574,
          "p95": 48115,
          "p99": 55238,
          "p999": 65812
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 154.42,
            "max": 166.59,
            "mean": 163.3
          },
          "cpu": {
            "min": 0.47,
            "max": 20.73,
            "mean": 2.56
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 60.58,
            "max": 83.01,
            "mean": 80.07
          },
          "cpu": {
            "min": 0,
            "max": 95.15,
            "mean": 8.57
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38,
      "counters": {
        "benchmark.http_2xx": {
          "value": 160402,
          "perSecond": 5346.58
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
          "value": 25183114,
          "perSecond": 839413.75
        },
        "upstream_cx_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7219395,
          "perSecond": 240639.8
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
          "value": 160431,
          "perSecond": 5347.55
        }
      }
    },
    {
      "testName": "scaling up httproutes to 1000 with 200 routes per hostname",
      "routes": 1000,
      "routesPerHostname": 200,
      "phase": "scaling-up",
      "throughput": 5218.06,
      "totalRequests": 156544,
      "latency": {
        "min": 367,
        "mean": 6719,
        "max": 84127,
        "pstdev": 12353,
        "percentiles": {
          "p50": 3065,
          "p75": 4893,
          "p80": 5543,
          "p90": 8557,
          "p95": 47892,
          "p99": 56375,
          "p999": 67162
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 161.57,
            "max": 194.59,
            "mean": 189.96
          },
          "cpu": {
            "min": 0.47,
            "max": 49.87,
            "mean": 5.41
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 82.7,
            "max": 128.93,
            "mean": 125.62
          },
          "cpu": {
            "min": 0,
            "max": 99.38,
            "mean": 16.19
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37,
      "counters": {
        "benchmark.http_2xx": {
          "value": 156544,
          "perSecond": 5218.06
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
          "value": 24577408,
          "perSecond": 819234.78
        },
        "upstream_cx_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7045605,
          "perSecond": 234850.02
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
          "value": 156569,
          "perSecond": 5218.89
        }
      }
    },
    {
      "testName": "scaling down httproutes to 10 with 2 routes per hostname",
      "routes": 10,
      "routesPerHostname": 2,
      "phase": "scaling-down",
      "throughput": 5387.19,
      "totalRequests": 161616,
      "latency": {
        "min": 364,
        "mean": 6542,
        "max": 91033,
        "pstdev": 12047,
        "percentiles": {
          "p50": 3028,
          "p75": 4823,
          "p80": 5436,
          "p90": 8323,
          "p95": 47945,
          "p99": 54667,
          "p999": 61505
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 196.85,
            "max": 204.68,
            "mean": 200.34
          },
          "cpu": {
            "min": 0.6,
            "max": 3.73,
            "mean": 1.01
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 124.9,
            "max": 125.1,
            "mean": 124.96
          },
          "cpu": {
            "min": 0,
            "max": 97.82,
            "mean": 1.97
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37,
      "counters": {
        "benchmark.http_2xx": {
          "value": 161616,
          "perSecond": 5387.19
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
          "value": 25373712,
          "perSecond": 845789.02
        },
        "upstream_cx_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7274385,
          "perSecond": 242479.1
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
          "value": 161653,
          "perSecond": 5388.42
        }
      }
    },
    {
      "testName": "scaling down httproutes to 50 with 10 routes per hostname",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-down",
      "throughput": 5377.52,
      "totalRequests": 161329,
      "latency": {
        "min": 353,
        "mean": 6666,
        "max": 69062,
        "pstdev": 11937,
        "percentiles": {
          "p50": 3235,
          "p75": 5004,
          "p80": 5575,
          "p90": 8050,
          "p95": 47345,
          "p99": 53839,
          "p999": 59277
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 167.66,
            "max": 204.68,
            "mean": 198.45
          },
          "cpu": {
            "min": 0.6,
            "max": 5,
            "mean": 1.14
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 124.91,
            "max": 126.91,
            "mean": 125.07
          },
          "cpu": {
            "min": 0,
            "max": 99.99,
            "mean": 12.3
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38,
      "counters": {
        "benchmark.http_2xx": {
          "value": 161329,
          "perSecond": 5377.52
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
          "value": 25328653,
          "perSecond": 844270.8
        },
        "upstream_cx_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7261290,
          "perSecond": 242037.94
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
          "value": 161362,
          "perSecond": 5378.62
        }
      }
    },
    {
      "testName": "scaling down httproutes to 100 with 20 routes per hostname",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-down",
      "throughput": 5263.08,
      "totalRequests": 157896,
      "latency": {
        "min": 379,
        "mean": 7049,
        "max": 80699,
        "pstdev": 12535,
        "percentiles": {
          "p50": 3279,
          "p75": 5204,
          "p80": 5879,
          "p90": 9138,
          "p95": 48818,
          "p99": 55332,
          "p999": 61927
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 144,
            "max": 167.66,
            "mean": 149.36
          },
          "cpu": {
            "min": 0.67,
            "max": 12,
            "mean": 1.82
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 126.91,
            "max": 134.89,
            "mean": 127.6
          },
          "cpu": {
            "min": 0,
            "max": 68.16,
            "mean": 6.71
          }
        }
      },
      "poolOverflow": 361,
      "upstreamConnections": 39,
      "counters": {
        "benchmark.http_2xx": {
          "value": 157896,
          "perSecond": 5263.08
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
          "value": 24789672,
          "perSecond": 826303.92
        },
        "upstream_cx_total": {
          "value": 39,
          "perSecond": 1.3
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7106580,
          "perSecond": 236880.7
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
          "value": 157924,
          "perSecond": 5264.02
        }
      }
    },
    {
      "testName": "scaling down httproutes to 300 with 60 routes per hostname",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-down",
      "throughput": 5316.8,
      "totalRequests": 159506,
      "latency": {
        "min": 359,
        "mean": 6962,
        "max": 81219,
        "pstdev": 12393,
        "percentiles": {
          "p50": 3266,
          "p75": 5154,
          "p80": 5830,
          "p90": 8940,
          "p95": 48343,
          "p99": 55275,
          "p999": 64294
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 157.35,
            "max": 161.75,
            "mean": 158.67
          },
          "cpu": {
            "min": 0.67,
            "max": 11.27,
            "mean": 2
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 134.9,
            "max": 135.31,
            "mean": 135.02
          },
          "cpu": {
            "min": 0,
            "max": 100.09,
            "mean": 13.37
          }
        }
      },
      "poolOverflow": 361,
      "upstreamConnections": 39,
      "counters": {
        "benchmark.http_2xx": {
          "value": 159506,
          "perSecond": 5316.8
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
          "value": 25042442,
          "perSecond": 834736.89
        },
        "upstream_cx_total": {
          "value": 39,
          "perSecond": 1.3
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7179075,
          "perSecond": 239299.3
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
          "value": 159535,
          "perSecond": 5317.76
        }
      }
    },
    {
      "testName": "scaling down httproutes to 500 with 100 routes per hostname",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-down",
      "throughput": 5381.7,
      "totalRequests": 161451,
      "latency": {
        "min": 351,
        "mean": 6541,
        "max": 88997,
        "pstdev": 12071,
        "percentiles": {
          "p50": 3074,
          "p75": 4769,
          "p80": 5326,
          "p90": 7898,
          "p95": 47615,
          "p99": 54716,
          "p999": 65511
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 154.82,
            "max": 190.15,
            "mean": 165.17
          },
          "cpu": {
            "min": 0.53,
            "max": 33.2,
            "mean": 3.85
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 128.91,
            "max": 135.08,
            "mean": 134.49
          },
          "cpu": {
            "min": 0,
            "max": 99.98,
            "mean": 9.01
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37,
      "counters": {
        "benchmark.http_2xx": {
          "value": 161451,
          "perSecond": 5381.7
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
          "value": 25347807,
          "perSecond": 844926.36
        },
        "upstream_cx_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7266960,
          "perSecond": 242231.85
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
          "value": 161488,
          "perSecond": 5382.93
        }
      }
    }
  ]
};

export default benchmarkData;
