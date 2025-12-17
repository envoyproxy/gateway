import { TestSuite } from '../types';

// Benchmark data extracted from markdown report for version 1.6.0
// Generated on 2025-12-17T03:01:22.910Z

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.6.0",
    "runId": "1.6.0-1765940482907",
    "date": "2025-12-17T03:01:22.907Z",
    "environment": "GitHub CI",
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
      "throughput": 5339.06,
      "totalRequests": 160172,
      "latency": {
        "min": 377,
        "mean": 6730,
        "max": 89411,
        "pstdev": 11378,
        "percentiles": {
          "p50": 3326,
          "p75": 5429,
          "p80": 6219,
          "p90": 9755,
          "p95": 43753,
          "p99": 53676,
          "p999": 58869
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 116.49,
            "max": 137.48,
            "mean": 136.01
          },
          "cpu": {
            "min": 0,
            "max": 1.26,
            "mean": 0.45
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 4.3,
            "max": 27.61,
            "mean": 24.84
          },
          "cpu": {
            "min": 0,
            "max": 99.91,
            "mean": 3.8
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38,
      "counters": {
        "benchmark.http_2xx": {
          "value": 160172,
          "perSecond": 5339.06
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
          "value": 25147004,
          "perSecond": 838232.6
        },
        "upstream_cx_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7209450,
          "perSecond": 240314.75
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
          "value": 160210,
          "perSecond": 5340.33
        }
      }
    },
    {
      "testName": "scaling up httproutes to 50 with 10 routes per hostname",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-up",
      "throughput": 5301.93,
      "totalRequests": 159058,
      "latency": {
        "min": 376,
        "mean": 6607,
        "max": 73531,
        "pstdev": 11890,
        "percentiles": {
          "p50": 3156,
          "p75": 4917,
          "p80": 5531,
          "p90": 8259,
          "p95": 47218,
          "p99": 54067,
          "p999": 59246
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 134.82,
            "max": 148.47,
            "mean": 145.25
          },
          "cpu": {
            "min": 0.33,
            "max": 2,
            "mean": 0.62
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 29.24,
            "max": 33.7,
            "mean": 32.86
          },
          "cpu": {
            "min": 0,
            "max": 87.87,
            "mean": 5.11
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37,
      "counters": {
        "benchmark.http_2xx": {
          "value": 159058,
          "perSecond": 5301.93
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
          "value": 24972106,
          "perSecond": 832402.72
        },
        "upstream_cx_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7159275,
          "perSecond": 238642.27
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
          "value": 159095,
          "perSecond": 5303.16
        }
      }
    },
    {
      "testName": "scaling up httproutes to 100 with 20 routes per hostname",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-up",
      "throughput": 5336.06,
      "totalRequests": 160089,
      "latency": {
        "min": 384,
        "mean": 6790,
        "max": 76795,
        "pstdev": 12170,
        "percentiles": {
          "p50": 3197,
          "p75": 5026,
          "p80": 5667,
          "p90": 8665,
          "p95": 48017,
          "p99": 54579,
          "p999": 60415
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 138.84,
            "max": 147.74,
            "mean": 145.34
          },
          "cpu": {
            "min": 0.4,
            "max": 2.67,
            "mean": 0.81
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 37.31,
            "max": 39.85,
            "mean": 39.28
          },
          "cpu": {
            "min": 0,
            "max": 99.98,
            "mean": 2.72
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38,
      "counters": {
        "benchmark.http_2xx": {
          "value": 160089,
          "perSecond": 5336.06
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
          "value": 25133973,
          "perSecond": 837762.12
        },
        "upstream_cx_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7205040,
          "perSecond": 240157.4
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
          "value": 160112,
          "perSecond": 5336.83
        }
      }
    },
    {
      "testName": "scaling up httproutes to 300 with 60 routes per hostname",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-up",
      "throughput": 5270.65,
      "totalRequests": 158120,
      "latency": {
        "min": 379,
        "mean": 6688,
        "max": 80261,
        "pstdev": 12107,
        "percentiles": {
          "p50": 3160,
          "p75": 4938,
          "p80": 5551,
          "p90": 8416,
          "p95": 47781,
          "p99": 55164,
          "p999": 61943
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 143.84,
            "max": 153.12,
            "mean": 152.05
          },
          "cpu": {
            "min": 0.47,
            "max": 4.8,
            "mean": 1.07
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 59.5,
            "max": 61.68,
            "mean": 61.56
          },
          "cpu": {
            "min": 0,
            "max": 100.02,
            "mean": 9.25
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37,
      "counters": {
        "benchmark.http_2xx": {
          "value": 158120,
          "perSecond": 5270.65
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
          "value": 24824840,
          "perSecond": 827492.11
        },
        "upstream_cx_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7117065,
          "perSecond": 237234.77
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
          "value": 158157,
          "perSecond": 5271.88
        }
      }
    },
    {
      "testName": "scaling up httproutes to 500 with 100 routes per hostname",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-up",
      "throughput": 5247.62,
      "totalRequests": 157429,
      "latency": {
        "min": 356,
        "mean": 6872,
        "max": 81539,
        "pstdev": 12331,
        "percentiles": {
          "p50": 3228,
          "p75": 5062,
          "p80": 5705,
          "p90": 8678,
          "p95": 48199,
          "p99": 55171,
          "p999": 65744
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 161.45,
            "max": 169.55,
            "mean": 166.8
          },
          "cpu": {
            "min": 0.6,
            "max": 6.87,
            "mean": 1.46
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 81.71,
            "max": 86.02,
            "mean": 85.62
          },
          "cpu": {
            "min": 0,
            "max": 99.81,
            "mean": 6.02
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38,
      "counters": {
        "benchmark.http_2xx": {
          "value": 157429,
          "perSecond": 5247.62
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
          "value": 24716353,
          "perSecond": 823876.99
        },
        "upstream_cx_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7086015,
          "perSecond": 236200.09
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
          "value": 157467,
          "perSecond": 5248.89
        }
      }
    },
    {
      "testName": "scaling up httproutes to 1000 with 200 routes per hostname",
      "routes": 1000,
      "routesPerHostname": 200,
      "phase": "scaling-up",
      "throughput": 5124.53,
      "totalRequests": 153737,
      "latency": {
        "min": 357,
        "mean": 7041,
        "max": 88129,
        "pstdev": 12624,
        "percentiles": {
          "p50": 3241,
          "p75": 5175,
          "p80": 5882,
          "p90": 9462,
          "p95": 47947,
          "p99": 57892,
          "p999": 69079
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 181.93,
            "max": 192.07,
            "mean": 188.87
          },
          "cpu": {
            "min": 0.6,
            "max": 12.67,
            "mean": 2.17
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 140.27,
            "max": 140.47,
            "mean": 140.43
          },
          "cpu": {
            "min": 0,
            "max": 99.81,
            "mean": 14.62
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38,
      "counters": {
        "benchmark.http_2xx": {
          "value": 153737,
          "perSecond": 5124.53
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
          "value": 24136709,
          "perSecond": 804550.82
        },
        "upstream_cx_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_tx_bytes_total": {
          "value": 6918750,
          "perSecond": 230623.24
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
          "value": 153750,
          "perSecond": 5124.96
        }
      }
    },
    {
      "testName": "scaling down httproutes to 10 with 2 routes per hostname",
      "routes": 10,
      "routesPerHostname": 2,
      "phase": "scaling-down",
      "throughput": 5332.22,
      "totalRequests": 159967,
      "latency": {
        "min": 377,
        "mean": 6769,
        "max": 68108,
        "pstdev": 12093,
        "percentiles": {
          "p50": 3195,
          "p75": 5134,
          "p80": 5793,
          "p90": 8793,
          "p95": 47841,
          "p99": 54325,
          "p999": 59437
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 140.07,
            "max": 148.71,
            "mean": 144.23
          },
          "cpu": {
            "min": 0,
            "max": 1.07,
            "mean": 0.68
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 135.8,
            "max": 135.99,
            "mean": 135.88
          },
          "cpu": {
            "min": 0,
            "max": 37.96,
            "mean": 0.81
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38,
      "counters": {
        "benchmark.http_2xx": {
          "value": 159967,
          "perSecond": 5332.22
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
          "value": 25114819,
          "perSecond": 837157.78
        },
        "upstream_cx_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7200225,
          "perSecond": 240006.68
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
          "value": 160005,
          "perSecond": 5333.48
        }
      }
    },
    {
      "testName": "scaling down httproutes to 50 with 10 routes per hostname",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-down",
      "throughput": 5360.96,
      "totalRequests": 160829,
      "latency": {
        "min": 385,
        "mean": 6567,
        "max": 78565,
        "pstdev": 12020,
        "percentiles": {
          "p50": 3047,
          "p75": 4812,
          "p80": 5446,
          "p90": 8291,
          "p95": 47761,
          "p99": 54624,
          "p999": 59953
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 143.64,
            "max": 152.11,
            "mean": 147.81
          },
          "cpu": {
            "min": 0,
            "max": 1.2,
            "mean": 0.76
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 135.8,
            "max": 136.22,
            "mean": 135.88
          },
          "cpu": {
            "min": 0,
            "max": 99.95,
            "mean": 4.35
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37,
      "counters": {
        "benchmark.http_2xx": {
          "value": 160829,
          "perSecond": 5360.96
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
          "value": 25250153,
          "perSecond": 841670.33
        },
        "upstream_cx_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7238970,
          "perSecond": 241298.59
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
          "value": 160866,
          "perSecond": 5362.19
        }
      }
    },
    {
      "testName": "scaling down httproutes to 100 with 20 routes per hostname",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-down",
      "throughput": 5387.07,
      "totalRequests": 161620,
      "latency": {
        "min": 368,
        "mean": 6902,
        "max": 91717,
        "pstdev": 12257,
        "percentiles": {
          "p50": 3235,
          "p75": 5117,
          "p80": 5779,
          "p90": 8942,
          "p95": 48250,
          "p99": 55025,
          "p999": 61599
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 152.11,
            "max": 163.4,
            "mean": 154.49
          },
          "cpu": {
            "min": 0,
            "max": 1.07,
            "mean": 0.68
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 135.81,
            "max": 136.21,
            "mean": 135.88
          },
          "cpu": {
            "min": 0,
            "max": 0.47,
            "mean": 0.06
          }
        }
      },
      "poolOverflow": 361,
      "upstreamConnections": 39,
      "counters": {
        "benchmark.http_2xx": {
          "value": 161620,
          "perSecond": 5387.07
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
          "value": 25374340,
          "perSecond": 845769.67
        },
        "upstream_cx_total": {
          "value": 39,
          "perSecond": 1.3
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7274070,
          "perSecond": 242457.06
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
          "value": 161646,
          "perSecond": 5387.93
        }
      }
    },
    {
      "testName": "scaling down httproutes to 300 with 60 routes per hostname",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-down",
      "throughput": 5326.66,
      "totalRequests": 159800,
      "latency": {
        "min": 381,
        "mean": 6971,
        "max": 99508,
        "pstdev": 12521,
        "percentiles": {
          "p50": 3196,
          "p75": 5124,
          "p80": 5775,
          "p90": 9183,
          "p95": 48947,
          "p99": 55918,
          "p999": 63801
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 156.4,
            "max": 175.45,
            "mean": 164.47
          },
          "cpu": {
            "min": 0,
            "max": 1.27,
            "mean": 0.76
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 135.8,
            "max": 140.45,
            "mean": 136.18
          },
          "cpu": {
            "min": 0,
            "max": 99.85,
            "mean": 9.76
          }
        }
      },
      "poolOverflow": 361,
      "upstreamConnections": 39,
      "counters": {
        "benchmark.http_2xx": {
          "value": 159800,
          "perSecond": 5326.66
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
          "value": 25088600,
          "perSecond": 836285.42
        },
        "upstream_cx_total": {
          "value": 39,
          "perSecond": 1.3
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7192755,
          "perSecond": 239758.14
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
          "value": 159839,
          "perSecond": 5327.96
        }
      }
    },
    {
      "testName": "scaling down httproutes to 500 with 100 routes per hostname",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-down",
      "throughput": 5258.42,
      "totalRequests": 157753,
      "latency": {
        "min": 372,
        "mean": 6861,
        "max": 98947,
        "pstdev": 12533,
        "percentiles": {
          "p50": 3083,
          "p75": 4948,
          "p80": 5617,
          "p90": 8963,
          "p95": 48871,
          "p99": 56164,
          "p999": 66125
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 168.89,
            "max": 187.7,
            "mean": 172.5
          },
          "cpu": {
            "min": 0,
            "max": 1.2,
            "mean": 0.73
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 140.43,
            "max": 140.85,
            "mean": 140.5
          },
          "cpu": {
            "min": 0,
            "max": 91.32,
            "mean": 1.79
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38,
      "counters": {
        "benchmark.http_2xx": {
          "value": 157753,
          "perSecond": 5258.42
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
          "value": 24767221,
          "perSecond": 825571.62
        },
        "upstream_cx_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7100595,
          "perSecond": 236685.81
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
          "value": 157791,
          "perSecond": 5259.68
        }
      }
    }
  ]
};

export default benchmarkData;
