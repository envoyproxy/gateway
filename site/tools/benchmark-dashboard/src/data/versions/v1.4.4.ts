import { TestSuite } from '../types';

// Benchmark data extracted from markdown report for version 1.4.4
// Generated on 2025-11-21T02:41:45.325Z

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.4.4",
    "runId": "1.4.4-1763692905323",
    "date": "2025-11-21T02:41:45.323Z",
    "environment": "production",
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
        "min": 385,
        "mean": 7177,
        "max": 74239,
        "pstdev": 12629,
        "percentiles": {
          "p50": 3424,
          "p75": 5345,
          "p80": 6000,
          "p90": 9149,
          "p95": 49100,
          "p99": 55275,
          "p999": 59793
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
        "min": 367,
        "mean": 6673,
        "max": 75735,
        "pstdev": 11892,
        "percentiles": {
          "p50": 3219,
          "p75": 5053,
          "p80": 5669,
          "p90": 8493,
          "p95": 47017,
          "p99": 53731,
          "p999": 59070
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
        "min": 392,
        "mean": 6716,
        "max": 74948,
        "pstdev": 12218,
        "percentiles": {
          "p50": 3086,
          "p75": 4917,
          "p80": 5540,
          "p90": 8473,
          "p95": 48504,
          "p99": 54824,
          "p999": 60995
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
        "min": 386,
        "mean": 6733,
        "max": 87748,
        "pstdev": 12052,
        "percentiles": {
          "p50": 3188,
          "p75": 5058,
          "p80": 5703,
          "p90": 8646,
          "p95": 47349,
          "p99": 54261,
          "p999": 62337
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
        "min": 390,
        "mean": 6838,
        "max": 87953,
        "pstdev": 12514,
        "percentiles": {
          "p50": 3154,
          "p75": 4940,
          "p80": 5547,
          "p90": 8546,
          "p95": 48750,
          "p99": 55814,
          "p999": 66682
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
        "min": 369,
        "mean": 7157,
        "max": 95969,
        "pstdev": 13025,
        "percentiles": {
          "p50": 3187,
          "p75": 5196,
          "p80": 5933,
          "p90": 9716,
          "p95": 49264,
          "p99": 58398,
          "p999": 70475
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
        "min": 352,
        "mean": 6847,
        "max": 77496,
        "pstdev": 12213,
        "percentiles": {
          "p50": 3276,
          "p75": 5100,
          "p80": 5725,
          "p90": 8528,
          "p95": 48064,
          "p99": 54683,
          "p999": 58912
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
        "min": 399,
        "mean": 6732,
        "max": 77856,
        "pstdev": 12300,
        "percentiles": {
          "p50": 3112,
          "p75": 4945,
          "p80": 5568,
          "p90": 8439,
          "p95": 48916,
          "p99": 54951,
          "p999": 60592
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
        "min": 359,
        "mean": 6712,
        "max": 85757,
        "pstdev": 12224,
        "percentiles": {
          "p50": 3131,
          "p75": 4899,
          "p80": 5530,
          "p90": 8450,
          "p95": 48300,
          "p99": 55074,
          "p999": 61050
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
        "min": 398,
        "mean": 6830,
        "max": 80662,
        "pstdev": 12280,
        "percentiles": {
          "p50": 3200,
          "p75": 5016,
          "p80": 5664,
          "p90": 8694,
          "p95": 48179,
          "p99": 55212,
          "p999": 63549
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
        "min": 380,
        "mean": 6971,
        "max": 107388,
        "pstdev": 12508,
        "percentiles": {
          "p50": 3261,
          "p75": 5130,
          "p80": 5781,
          "p90": 8858,
          "p95": 48461,
          "p99": 55611,
          "p999": 67166
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
