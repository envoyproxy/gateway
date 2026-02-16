import { TestSuite } from '../types';

// Benchmark data extracted from markdown report for version 1.5.5
// Generated on 2025-11-21T02:42:24.639Z

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.5.5",
    "runId": "1.5.5-1763692944638",
    "date": "2025-11-21T02:42:24.638Z",
    "environment": "GitHub CI",
    "description": "Benchmark results for version 1.5.5",
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
      "throughput": 5036.75,
      "totalRequests": 151111,
      "latency": {
        "min": 358,
        "mean": 6858,
        "max": 73158,
        "pstdev": 11547,
        "percentiles": {
          "p50": 3378,
          "p75": 5537,
          "p80": 6339,
          "p90": 10011,
          "p95": 45465,
          "p99": 53417,
          "p999": 58159
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 113.75,
            "max": 134.09,
            "mean": 130.21
          },
          "cpu": {
            "min": 0,
            "max": 1.07,
            "mean": 0.45
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 0,
            "max": 28.17,
            "mean": 24.39
          },
          "cpu": {
            "min": 0,
            "max": 99.13,
            "mean": 9.4
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37,
      "counters": {
        "benchmark.http_2xx": {
          "value": 151111,
          "perSecond": 5036.75
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
          "value": 23724427,
          "perSecond": 790770.25
        },
        "upstream_cx_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_cx_tx_bytes_total": {
          "value": 6801480,
          "perSecond": 226703.39
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
          "value": 151144,
          "perSecond": 5037.85
        }
      }
    },
    {
      "testName": "scaling up httproutes to 50 with 10 routes per hostname",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-up",
      "throughput": 5111.15,
      "totalRequests": 153336,
      "latency": {
        "min": 402,
        "mean": 7045,
        "max": 86048,
        "pstdev": 12225,
        "percentiles": {
          "p50": 3316,
          "p75": 5416,
          "p80": 6159,
          "p90": 9640,
          "p95": 47675,
          "p99": 54513,
          "p999": 60006
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 134.09,
            "max": 140.93,
            "mean": 137.68
          },
          "cpu": {
            "min": 0.4,
            "max": 4.13,
            "mean": 0.86
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 27.99,
            "max": 32.19,
            "mean": 31.55
          },
          "cpu": {
            "min": 0,
            "max": 0.39,
            "mean": 0.03
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38,
      "counters": {
        "benchmark.http_2xx": {
          "value": 153336,
          "perSecond": 5111.15
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
          "value": 24073752,
          "perSecond": 802450.43
        },
        "upstream_cx_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_tx_bytes_total": {
          "value": 6901650,
          "perSecond": 230052.72
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
          "value": 153370,
          "perSecond": 5112.28
        }
      }
    },
    {
      "testName": "scaling up httproutes to 100 with 20 routes per hostname",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-up",
      "throughput": 5032,
      "totalRequests": 150960,
      "latency": {
        "min": 383,
        "mean": 6985,
        "max": 85159,
        "pstdev": 12380,
        "percentiles": {
          "p50": 3242,
          "p75": 5302,
          "p80": 6026,
          "p90": 9567,
          "p95": 48363,
          "p99": 54974,
          "p999": 61327
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 137.87,
            "max": 141.86,
            "mean": 139.55
          },
          "cpu": {
            "min": 0.4,
            "max": 5.2,
            "mean": 1.01
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 32.05,
            "max": 38.38,
            "mean": 37.23
          },
          "cpu": {
            "min": 0,
            "max": 0.56,
            "mean": 0.1
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37,
      "counters": {
        "benchmark.http_2xx": {
          "value": 150960,
          "perSecond": 5032
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
          "value": 23700720,
          "perSecond": 790023.81
        },
        "upstream_cx_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_cx_tx_bytes_total": {
          "value": 6794865,
          "perSecond": 226495.45
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
          "value": 150997,
          "perSecond": 5033.23
        }
      }
    },
    {
      "testName": "scaling up httproutes to 300 with 60 routes per hostname",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-up",
      "throughput": 4987.38,
      "totalRequests": 149622,
      "latency": {
        "min": 375,
        "mean": 6978,
        "max": 144031,
        "pstdev": 12357,
        "percentiles": {
          "p50": 3327,
          "p75": 5293,
          "p80": 5983,
          "p90": 9210,
          "p95": 47702,
          "p99": 54665,
          "p999": 65447
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 150.91,
            "max": 166.47,
            "mean": 154.26
          },
          "cpu": {
            "min": 0.47,
            "max": 14.47,
            "mean": 2.05
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 38.27,
            "max": 60.52,
            "mean": 59.21
          },
          "cpu": {
            "min": 0,
            "max": 99.83,
            "mean": 14.85
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37,
      "counters": {
        "benchmark.http_2xx": {
          "value": 149622,
          "perSecond": 4987.38
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
          "value": 23490654,
          "perSecond": 783019.07
        },
        "upstream_cx_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_cx_tx_bytes_total": {
          "value": 6734655,
          "perSecond": 224487.72
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
          "value": 149659,
          "perSecond": 4988.62
        }
      }
    },
    {
      "testName": "scaling up httproutes to 500 with 100 routes per hostname",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-up",
      "throughput": 5054.22,
      "totalRequests": 151627,
      "latency": {
        "min": 380,
        "mean": 6961,
        "max": 86892,
        "pstdev": 12306,
        "percentiles": {
          "p50": 3308,
          "p75": 5212,
          "p80": 5869,
          "p90": 9026,
          "p95": 48003,
          "p99": 54798,
          "p999": 66111
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 154.38,
            "max": 165.47,
            "mean": 160.87
          },
          "cpu": {
            "min": 0.47,
            "max": 15.86,
            "mean": 2.17
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 60.36,
            "max": 80.61,
            "mean": 78.16
          },
          "cpu": {
            "min": 0,
            "max": 62.38,
            "mean": 1.78
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37,
      "counters": {
        "benchmark.http_2xx": {
          "value": 151627,
          "perSecond": 5054.22
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
          "value": 23805439,
          "perSecond": 793512.74
        },
        "upstream_cx_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_cx_tx_bytes_total": {
          "value": 6824880,
          "perSecond": 227495.46
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
          "value": 151664,
          "perSecond": 5055.45
        }
      }
    },
    {
      "testName": "scaling up httproutes to 1000 with 200 routes per hostname",
      "routes": 1000,
      "routesPerHostname": 200,
      "phase": "scaling-up",
      "throughput": 4876.52,
      "totalRequests": 146296,
      "latency": {
        "min": 389,
        "mean": 7785,
        "max": 105218,
        "pstdev": 13242,
        "percentiles": {
          "p50": 3631,
          "p75": 5918,
          "p80": 6770,
          "p90": 11637,
          "p95": 49168,
          "p99": 58429,
          "p999": 73142
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 174.48,
            "max": 189.18,
            "mean": 185.55
          },
          "cpu": {
            "min": 0.53,
            "max": 33.87,
            "mean": 3.98
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 80.41,
            "max": 130.98,
            "mean": 125.04
          },
          "cpu": {
            "min": 0,
            "max": 100.22,
            "mean": 10.6
          }
        }
      },
      "poolOverflow": 360,
      "upstreamConnections": 40,
      "counters": {
        "benchmark.http_2xx": {
          "value": 146296,
          "perSecond": 4876.52
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
          "value": 22968472,
          "perSecond": 765614.39
        },
        "upstream_cx_total": {
          "value": 40,
          "perSecond": 1.33
        },
        "upstream_cx_tx_bytes_total": {
          "value": 6585120,
          "perSecond": 219503.61
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
          "value": 146336,
          "perSecond": 4877.86
        }
      }
    },
    {
      "testName": "scaling down httproutes to 10 with 2 routes per hostname",
      "routes": 10,
      "routesPerHostname": 2,
      "phase": "scaling-down",
      "throughput": 5174,
      "totalRequests": 155220,
      "latency": {
        "min": 383,
        "mean": 6982,
        "max": 84951,
        "pstdev": 12400,
        "percentiles": {
          "p50": 3264,
          "p75": 5191,
          "p80": 5851,
          "p90": 9111,
          "p95": 48539,
          "p99": 55015,
          "p999": 61159
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 200.76,
            "max": 209.96,
            "mean": 204
          },
          "cpu": {
            "min": 0.67,
            "max": 3.2,
            "mean": 1
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 124.56,
            "max": 124.95,
            "mean": 124.66
          },
          "cpu": {
            "min": 0,
            "max": 100.02,
            "mean": 10.45
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38,
      "counters": {
        "benchmark.http_2xx": {
          "value": 155220,
          "perSecond": 5174
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
          "value": 24369540,
          "perSecond": 812318.7
        },
        "upstream_cx_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_tx_bytes_total": {
          "value": 6986610,
          "perSecond": 232887.2
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
          "value": 155258,
          "perSecond": 5175.27
        }
      }
    },
    {
      "testName": "scaling down httproutes to 50 with 10 routes per hostname",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-down",
      "throughput": 5131.76,
      "totalRequests": 153955,
      "latency": {
        "min": 379,
        "mean": 7003,
        "max": 69431,
        "pstdev": 12325,
        "percentiles": {
          "p50": 3309,
          "p75": 5217,
          "p80": 5892,
          "p90": 9196,
          "p95": 48388,
          "p99": 54595,
          "p999": 59465
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 180.64,
            "max": 210.23,
            "mean": 203.47
          },
          "cpu": {
            "min": 0.67,
            "max": 4.47,
            "mean": 1.1
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 124.54,
            "max": 124.73,
            "mean": 124.6
          },
          "cpu": {
            "min": 0,
            "max": 100.06,
            "mean": 15.33
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38,
      "counters": {
        "benchmark.http_2xx": {
          "value": 153955,
          "perSecond": 5131.76
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
          "value": 24170935,
          "perSecond": 805685.76
        },
        "upstream_cx_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_tx_bytes_total": {
          "value": 6929505,
          "perSecond": 230980.04
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
          "value": 153989,
          "perSecond": 5132.89
        }
      }
    },
    {
      "testName": "scaling down httproutes to 100 with 20 routes per hostname",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-down",
      "throughput": 5135.4,
      "totalRequests": 154068,
      "latency": {
        "min": 389,
        "mean": 6840,
        "max": 90390,
        "pstdev": 12116,
        "percentiles": {
          "p50": 3218,
          "p75": 5120,
          "p80": 5774,
          "p90": 8987,
          "p95": 47542,
          "p99": 54333,
          "p999": 60504
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 145.4,
            "max": 167.27,
            "mean": 150.39
          },
          "cpu": {
            "min": 0.67,
            "max": 9.6,
            "mean": 1.6
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 124.55,
            "max": 132.84,
            "mean": 125.87
          },
          "cpu": {
            "min": 0,
            "max": 35.79,
            "mean": 1.32
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37,
      "counters": {
        "benchmark.http_2xx": {
          "value": 154068,
          "perSecond": 5135.4
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
          "value": 24188676,
          "perSecond": 806257.2
        },
        "upstream_cx_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_cx_tx_bytes_total": {
          "value": 6934455,
          "perSecond": 231139.33
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
          "value": 154099,
          "perSecond": 5136.43
        }
      }
    },
    {
      "testName": "scaling down httproutes to 300 with 60 routes per hostname",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-down",
      "throughput": 4805.71,
      "totalRequests": 144106,
      "latency": {
        "min": 355,
        "mean": 3755,
        "max": 71278,
        "pstdev": 8613,
        "percentiles": {
          "p50": 1813,
          "p75": 2693,
          "p80": 2987,
          "p90": 4022,
          "p95": 6251,
          "p99": 48558,
          "p999": 53964
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 158.2,
            "max": 168.5,
            "mean": 160.29
          },
          "cpu": {
            "min": 0.73,
            "max": 11,
            "mean": 1.76
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 132.83,
            "max": 132.92,
            "mean": 132.86
          },
          "cpu": {
            "min": 0,
            "max": 99.81,
            "mean": 5.92
          }
        }
      },
      "poolOverflow": 381,
      "upstreamConnections": 19,
      "counters": {
        "benchmark.http_2xx": {
          "value": 144106,
          "perSecond": 4805.71
        },
        "benchmark.pool_overflow": {
          "value": 381,
          "perSecond": 12.71
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
          "value": 19,
          "perSecond": 0.63
        },
        "upstream_cx_rx_bytes_total": {
          "value": 22624642,
          "perSecond": 754496.28
        },
        "upstream_cx_total": {
          "value": 19,
          "perSecond": 0.63
        },
        "upstream_cx_tx_bytes_total": {
          "value": 6485625,
          "perSecond": 216285.41
        },
        "upstream_rq_pending_overflow": {
          "value": 381,
          "perSecond": 12.71
        },
        "upstream_rq_pending_total": {
          "value": 19,
          "perSecond": 0.63
        },
        "upstream_rq_total": {
          "value": 144125,
          "perSecond": 4806.34
        }
      }
    },
    {
      "testName": "scaling down httproutes to 500 with 100 routes per hostname",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-down",
      "throughput": 4995.81,
      "totalRequests": 149876,
      "latency": {
        "min": 356,
        "mean": 6955,
        "max": 102629,
        "pstdev": 12066,
        "percentiles": {
          "p50": 3365,
          "p75": 5296,
          "p80": 5991,
          "p90": 9226,
          "p95": 46753,
          "p99": 54145,
          "p999": 65548
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 164.96,
            "max": 189.18,
            "mean": 168.91
          },
          "cpu": {
            "min": 0.47,
            "max": 22.33,
            "mean": 2.82
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 130.84,
            "max": 133.12,
            "mean": 132.79
          },
          "cpu": {
            "min": 0,
            "max": 99.83,
            "mean": 13.53
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37,
      "counters": {
        "benchmark.http_2xx": {
          "value": 149876,
          "perSecond": 4995.81
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
          "value": 23530532,
          "perSecond": 784342.88
        },
        "upstream_cx_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_cx_tx_bytes_total": {
          "value": 6746085,
          "perSecond": 224867.15
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
          "value": 149913,
          "perSecond": 4997.05
        }
      }
    }
  ]
};

export default benchmarkData;
