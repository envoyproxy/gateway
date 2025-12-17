import { TestSuite } from '../types';

// Benchmark data extracted from markdown report for version 1.6.1
// Generated on 2025-12-17T00:36:52.525Z

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.6.1",
    "runId": "1.6.1-1765931812524",
    "date": "2025-12-17T00:36:52.524Z",
    "environment": "production",
    "description": "Benchmark results for version 1.6.1",
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
      "throughput": 5304.85,
      "totalRequests": 159146,
      "latency": {
        "min": 376,
        "mean": 6595,
        "max": 70885,
        "pstdev": 11480,
        "percentiles": {
          "p50": 3196,
          "p75": 5135,
          "p80": 5825,
          "p90": 9106,
          "p95": 45002,
          "p99": 53958,
          "p999": 58937
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 121.21,
            "max": 137.21,
            "mean": 133.6
          },
          "cpu": {
            "min": 0.27,
            "max": 1.07,
            "mean": 0.45
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 5.99,
            "max": 27.8,
            "mean": 26.65
          },
          "cpu": {
            "min": 0,
            "max": 100,
            "mean": 7.14
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37,
      "counters": {
        "benchmark.http_2xx": {
          "value": 159146,
          "perSecond": 5304.85
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
          "value": 24985922,
          "perSecond": 832861.45
        },
        "upstream_cx_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7163235,
          "perSecond": 238773.75
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
          "value": 159183,
          "perSecond": 5306.08
        }
      }
    },
    {
      "testName": "scaling up httproutes to 50 with 10 routes per hostname",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-up",
      "throughput": 5315.3,
      "totalRequests": 159459,
      "latency": {
        "min": 394,
        "mean": 6608,
        "max": 78028,
        "pstdev": 11984,
        "percentiles": {
          "p50": 3110,
          "p75": 4873,
          "p80": 5487,
          "p90": 8302,
          "p95": 47810,
          "p99": 54222,
          "p999": 59463
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 137.2,
            "max": 147.08,
            "mean": 142.14
          },
          "cpu": {
            "min": 0.27,
            "max": 1.93,
            "mean": 0.63
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 29.66,
            "max": 33.89,
            "mean": 33.5
          },
          "cpu": {
            "min": 0,
            "max": 97.78,
            "mean": 9.59
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37,
      "counters": {
        "benchmark.http_2xx": {
          "value": 159459,
          "perSecond": 5315.3
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
          "value": 25035063,
          "perSecond": 834501.78
        },
        "upstream_cx_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7177320,
          "perSecond": 239243.91
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
          "value": 159496,
          "perSecond": 5316.53
        }
      }
    },
    {
      "testName": "scaling up httproutes to 100 with 20 routes per hostname",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-up",
      "throughput": 5283.74,
      "totalRequests": 158514,
      "latency": {
        "min": 378,
        "mean": 6628,
        "max": 93491,
        "pstdev": 11944,
        "percentiles": {
          "p50": 3137,
          "p75": 5002,
          "p80": 5624,
          "p90": 8369,
          "p95": 47605,
          "p99": 53665,
          "p999": 58869
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 140.7,
            "max": 151.43,
            "mean": 145.16
          },
          "cpu": {
            "min": 0.4,
            "max": 2.67,
            "mean": 0.76
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 37.73,
            "max": 40.02,
            "mean": 39.47
          },
          "cpu": {
            "min": 0,
            "max": 25.9,
            "mean": 0.59
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37,
      "counters": {
        "benchmark.http_2xx": {
          "value": 158514,
          "perSecond": 5283.74
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
          "value": 24886698,
          "perSecond": 829546.43
        },
        "upstream_cx_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7134615,
          "perSecond": 237817.59
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
          "value": 158547,
          "perSecond": 5284.84
        }
      }
    },
    {
      "testName": "scaling up httproutes to 300 with 60 routes per hostname",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-up",
      "throughput": 5204.25,
      "totalRequests": 156128,
      "latency": {
        "min": 373,
        "mean": 6726,
        "max": 91820,
        "pstdev": 12218,
        "percentiles": {
          "p50": 3131,
          "p75": 4960,
          "p80": 5612,
          "p90": 8593,
          "p95": 47857,
          "p99": 54527,
          "p999": 63778
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 147.88,
            "max": 153.92,
            "mean": 151.47
          },
          "cpu": {
            "min": 0.47,
            "max": 4.73,
            "mean": 1.07
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 62.02,
            "max": 62.22,
            "mean": 62.08
          },
          "cpu": {
            "min": 0,
            "max": 45.07,
            "mean": 1.13
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37,
      "counters": {
        "benchmark.http_2xx": {
          "value": 156128,
          "perSecond": 5204.25
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
          "value": 24512096,
          "perSecond": 817067.71
        },
        "upstream_cx_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7027425,
          "perSecond": 234246.88
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
          "value": 156165,
          "perSecond": 5205.49
        }
      }
    },
    {
      "testName": "scaling up httproutes to 500 with 100 routes per hostname",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-up",
      "throughput": 5194.25,
      "totalRequests": 155829,
      "latency": {
        "min": 383,
        "mean": 6935,
        "max": 79613,
        "pstdev": 12512,
        "percentiles": {
          "p50": 3204,
          "p75": 5117,
          "p80": 5808,
          "p90": 8937,
          "p95": 48531,
          "p99": 55611,
          "p999": 67059
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 157.16,
            "max": 172.57,
            "mean": 168.57
          },
          "cpu": {
            "min": 0.53,
            "max": 7.2,
            "mean": 1.47
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 84.09,
            "max": 86.29,
            "mean": 86.04
          },
          "cpu": {
            "min": 0,
            "max": 100.05,
            "mean": 12.87
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38,
      "counters": {
        "benchmark.http_2xx": {
          "value": 155829,
          "perSecond": 5194.25
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
          "value": 24465153,
          "perSecond": 815496.96
        },
        "upstream_cx_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7013565,
          "perSecond": 233783.17
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
          "value": 155857,
          "perSecond": 5195.18
        }
      }
    },
    {
      "testName": "scaling up httproutes to 1000 with 200 routes per hostname",
      "routes": 1000,
      "routesPerHostname": 200,
      "phase": "scaling-up",
      "throughput": 5115.06,
      "totalRequests": 153454,
      "latency": {
        "min": 393,
        "mean": 7085,
        "max": 109756,
        "pstdev": 12950,
        "percentiles": {
          "p50": 3195,
          "p75": 5039,
          "p80": 5692,
          "p90": 8994,
          "p95": 49426,
          "p99": 58230,
          "p999": 70422
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 173.94,
            "max": 196.39,
            "mean": 190.77
          },
          "cpu": {
            "min": 0.6,
            "max": 13.93,
            "mean": 1.97
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 140.53,
            "max": 140.7,
            "mean": 140.57
          },
          "cpu": {
            "min": 0,
            "max": 5.67,
            "mean": 1.57
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38,
      "counters": {
        "benchmark.http_2xx": {
          "value": 153454,
          "perSecond": 5115.06
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
          "value": 24092278,
          "perSecond": 803064.79
        },
        "upstream_cx_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_tx_bytes_total": {
          "value": 6906420,
          "perSecond": 230210.81
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
          "value": 153476,
          "perSecond": 5115.8
        }
      }
    },
    {
      "testName": "scaling down httproutes to 10 with 2 routes per hostname",
      "routes": 10,
      "routesPerHostname": 2,
      "phase": "scaling-down",
      "throughput": 5338.31,
      "totalRequests": 160153,
      "latency": {
        "min": 379,
        "mean": 6782,
        "max": 75825,
        "pstdev": 11999,
        "percentiles": {
          "p50": 3274,
          "p75": 5094,
          "p80": 5709,
          "p90": 8727,
          "p95": 47667,
          "p99": 54355,
          "p999": 59471
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 145.52,
            "max": 147.96,
            "mean": 146.23
          },
          "cpu": {
            "min": 0,
            "max": 1.07,
            "mean": 0.75
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 138.12,
            "max": 138.57,
            "mean": 138.22
          },
          "cpu": {
            "min": 0,
            "max": 100.01,
            "mean": 9.46
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38,
      "counters": {
        "benchmark.http_2xx": {
          "value": 160153,
          "perSecond": 5338.31
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
          "value": 25144021,
          "perSecond": 838114.93
        },
        "upstream_cx_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7207920,
          "perSecond": 240258.52
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
          "value": 160176,
          "perSecond": 5339.08
        }
      }
    },
    {
      "testName": "scaling down httproutes to 50 with 10 routes per hostname",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-down",
      "throughput": 5292.35,
      "totalRequests": 158769,
      "latency": {
        "min": 379,
        "mean": 6680,
        "max": 76271,
        "pstdev": 12093,
        "percentiles": {
          "p50": 3116,
          "p75": 4881,
          "p80": 5514,
          "p90": 8447,
          "p95": 48091,
          "p99": 54738,
          "p999": 61007
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 141.33,
            "max": 151.66,
            "mean": 147.23
          },
          "cpu": {
            "min": 0,
            "max": 1.2,
            "mean": 0.71
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 138.14,
            "max": 138.32,
            "mean": 138.19
          },
          "cpu": {
            "min": 0,
            "max": 99.95,
            "mean": 4.5
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37,
      "counters": {
        "benchmark.http_2xx": {
          "value": 158769,
          "perSecond": 5292.35
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
          "value": 24926733,
          "perSecond": 830898.42
        },
        "upstream_cx_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7145865,
          "perSecond": 238197.6
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
          "value": 158797,
          "perSecond": 5293.28
        }
      }
    },
    {
      "testName": "scaling down httproutes to 100 with 20 routes per hostname",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-down",
      "throughput": 5353.39,
      "totalRequests": 160607,
      "latency": {
        "min": 379,
        "mean": 7767,
        "max": 94674,
        "pstdev": 12965,
        "percentiles": {
          "p50": 3649,
          "p75": 5985,
          "p80": 6804,
          "p90": 11396,
          "p95": 49006,
          "p99": 55625,
          "p999": 63916
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 146.41,
            "max": 158.55,
            "mean": 150.46
          },
          "cpu": {
            "min": 0,
            "max": 1.2,
            "mean": 0.75
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 138.3,
            "max": 138.52,
            "mean": 138.36
          },
          "cpu": {
            "min": 0,
            "max": 100.2,
            "mean": 9.9
          }
        }
      },
      "poolOverflow": 356,
      "upstreamConnections": 44,
      "counters": {
        "benchmark.http_2xx": {
          "value": 160607,
          "perSecond": 5353.39
        },
        "benchmark.pool_overflow": {
          "value": 356,
          "perSecond": 11.87
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
          "value": 44,
          "perSecond": 1.47
        },
        "upstream_cx_rx_bytes_total": {
          "value": 25215299,
          "perSecond": 840482.78
        },
        "upstream_cx_total": {
          "value": 44,
          "perSecond": 1.47
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7228575,
          "perSecond": 240944.71
        },
        "upstream_rq_pending_overflow": {
          "value": 356,
          "perSecond": 11.87
        },
        "upstream_rq_pending_total": {
          "value": 44,
          "perSecond": 1.47
        },
        "upstream_rq_total": {
          "value": 160635,
          "perSecond": 5354.33
        }
      }
    },
    {
      "testName": "scaling down httproutes to 300 with 60 routes per hostname",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-down",
      "throughput": 5251.66,
      "totalRequests": 157550,
      "latency": {
        "min": 362,
        "mean": 6850,
        "max": 102064,
        "pstdev": 12185,
        "percentiles": {
          "p50": 3190,
          "p75": 5111,
          "p80": 5817,
          "p90": 9126,
          "p95": 47685,
          "p99": 54695,
          "p999": 63205
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 158.54,
            "max": 168.72,
            "mean": 160.6
          },
          "cpu": {
            "min": 0,
            "max": 1.2,
            "mean": 0.73
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 138.3,
            "max": 140.56,
            "mean": 138.63
          },
          "cpu": {
            "min": 0,
            "max": 99.87,
            "mean": 4.61
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38,
      "counters": {
        "benchmark.http_2xx": {
          "value": 157550,
          "perSecond": 5251.66
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
          "value": 24735350,
          "perSecond": 824510.95
        },
        "upstream_cx_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7091460,
          "perSecond": 236381.79
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
          "value": 157588,
          "perSecond": 5252.93
        }
      }
    },
    {
      "testName": "scaling down httproutes to 500 with 100 routes per hostname",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-down",
      "throughput": 5226.15,
      "totalRequests": 156791,
      "latency": {
        "min": 367,
        "mean": 6930,
        "max": 87855,
        "pstdev": 12538,
        "percentiles": {
          "p50": 3176,
          "p75": 5094,
          "p80": 5780,
          "p90": 9033,
          "p95": 48709,
          "p99": 55670,
          "p999": 65673
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 167.79,
            "max": 196.18,
            "mean": 171.16
          },
          "cpu": {
            "min": 0,
            "max": 1.13,
            "mean": 0.74
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 140.56,
            "max": 149.79,
            "mean": 141.48
          },
          "cpu": {
            "min": 0,
            "max": 99.89,
            "mean": 9.34
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38,
      "counters": {
        "benchmark.http_2xx": {
          "value": 156791,
          "perSecond": 5226.15
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
          "value": 24616187,
          "perSecond": 820506.23
        },
        "upstream_cx_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7056720,
          "perSecond": 235214.44
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
          "value": 156816,
          "perSecond": 5226.99
        }
      }
    }
  ]
};

export default benchmarkData;
