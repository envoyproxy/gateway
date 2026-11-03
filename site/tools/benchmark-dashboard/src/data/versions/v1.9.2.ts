import { TestSuite } from '../types';

// Benchmark data extracted from release artifact for version 1.9.2
// Generated from benchmark_result.json

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.9.2",
    "runId": "1.9.2-release-2026-09-28",
    "date": "2026-09-28T17:14:42Z",
    "environment": "GitHub Release",
    "description": "Benchmark results for version 1.9.2 from release artifacts",
    "downloadUrl": "https://github.com/envoyproxy/gateway/releases/download/v1.9.2/benchmark_report.zip",
    "testConfiguration": {
      "connections": 100,
      "cpuLimit": "1000m",
      "duration": 90,
      "memoryLimit": "2000Mi",
      "rps": 100
    }
  },
  "results": [
    {
      "testName": "scaling up httproutes to 10 with 2 routes per hostname at 100 rps",
      "routes": 10,
      "routesPerHostname": 2,
      "phase": "scaling-up",
      "throughput": 404.4269662921348,
      "totalRequests": 35994,
      "latency": {
        "max": 65.015807,
        "min": 0.186224,
        "mean": 0.315139,
        "pstdev": 0.681696,
        "percentiles": {
          "p50": 0.282143,
          "p75": 0.31255900000000003,
          "p80": 0.321983,
          "p90": 0.352895,
          "p95": 0.394207,
          "p99": 0.822943,
          "p999": 3.778431
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 138.3671875,
            "min": 117.4453125,
            "mean": 133.910546875
          },
          "cpu": {
            "max": 1.0000000000000002,
            "min": 0.20000000000000018,
            "mean": 0.42142857142857126
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 21.01953125,
            "min": 9.4921875,
            "mean": 19.012369791666668
          },
          "cpu": {
            "max": 7.018474365537923,
            "min": 4.944632165094225,
            "mean": 6.632892040261022
          }
        }
      },
      "poolOverflow": 6,
      "upstreamConnections": 11,
      "counters": {
        "benchmark.http_2xx": {
          "value": 35994,
          "perSecond": 404.4269662921348
        },
        "benchmark.pool_overflow": {
          "value": 6,
          "perSecond": 0.06741573033707865
        },
        "cluster_manager.cluster_added": {
          "value": 4,
          "perSecond": 0.0449438202247191
        },
        "default.total_match_count": {
          "value": 4,
          "perSecond": 0.0449438202247191
        },
        "membership_change": {
          "value": 4,
          "perSecond": 0.0449438202247191
        },
        "runtime.load_success": {
          "value": 1,
          "perSecond": 0.011235955056179775
        },
        "runtime.override_dir_not_exists": {
          "value": 1,
          "perSecond": 0.011235955056179775
        },
        "upstream_cx_http1_total": {
          "value": 11,
          "perSecond": 0.12359550561797752
        },
        "upstream_cx_rx_bytes_total": {
          "value": 5651058,
          "perSecond": 63495.03370786517
        },
        "upstream_cx_total": {
          "value": 11,
          "perSecond": 0.12359550561797752
        },
        "upstream_cx_tx_bytes_total": {
          "value": 1619730,
          "perSecond": 18199.213483146068
        },
        "upstream_rq_pending_overflow": {
          "value": 6,
          "perSecond": 0.06741573033707865
        },
        "upstream_rq_pending_total": {
          "value": 11,
          "perSecond": 0.12359550561797752
        },
        "upstream_rq_total": {
          "value": 35994,
          "perSecond": 404.4269662921348
        }
      }
    },
    {
      "testName": "scaling up httproutes to 50 with 10 routes per hostname at 300 rps",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-up",
      "throughput": 1199.8777777777777,
      "totalRequests": 107989,
      "latency": {
        "max": 23.242751,
        "min": 0.15676,
        "mean": 0.279506,
        "pstdev": 0.275694,
        "percentiles": {
          "p50": 0.254279,
          "p75": 0.279759,
          "p80": 0.287583,
          "p90": 0.318335,
          "p95": 0.364687,
          "p99": 0.827423,
          "p999": 3.2830709999999996
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 141.49609375,
            "min": 137.7890625,
            "mean": 140.074609375
          },
          "cpu": {
            "max": 0.4666666666666657,
            "min": 0.3333333333333336,
            "mean": 0.40000000000000013
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 25.3828125,
            "min": 20.78515625,
            "mean": 24.366276041666666
          },
          "cpu": {
            "max": 17.752909034939204,
            "min": 11.553001436303935,
            "mean": 17.115177224673218
          }
        }
      },
      "poolOverflow": 11,
      "upstreamConnections": 20,
      "counters": {
        "benchmark.http_2xx": {
          "value": 107989,
          "perSecond": 1199.8777777777777
        },
        "benchmark.pool_overflow": {
          "value": 11,
          "perSecond": 0.12222222222222222
        },
        "cluster_manager.cluster_added": {
          "value": 4,
          "perSecond": 0.044444444444444446
        },
        "default.total_match_count": {
          "value": 4,
          "perSecond": 0.044444444444444446
        },
        "membership_change": {
          "value": 4,
          "perSecond": 0.044444444444444446
        },
        "runtime.load_success": {
          "value": 1,
          "perSecond": 0.011111111111111112
        },
        "runtime.override_dir_not_exists": {
          "value": 1,
          "perSecond": 0.011111111111111112
        },
        "upstream_cx_http1_total": {
          "value": 20,
          "perSecond": 0.2222222222222222
        },
        "upstream_cx_rx_bytes_total": {
          "value": 16954273,
          "perSecond": 188380.8111111111
        },
        "upstream_cx_total": {
          "value": 20,
          "perSecond": 0.2222222222222222
        },
        "upstream_cx_tx_bytes_total": {
          "value": 4859505,
          "perSecond": 53994.5
        },
        "upstream_rq_pending_overflow": {
          "value": 11,
          "perSecond": 0.12222222222222222
        },
        "upstream_rq_pending_total": {
          "value": 20,
          "perSecond": 0.2222222222222222
        },
        "upstream_rq_total": {
          "value": 107989,
          "perSecond": 1199.8777777777777
        }
      }
    },
    {
      "testName": "scaling up httproutes to 100 with 20 routes per hostname at 500 rps",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-up",
      "throughput": 2021.8539325842696,
      "totalRequests": 179945,
      "latency": {
        "max": 56.903679,
        "min": 0.138632,
        "mean": 0.25953800000000005,
        "pstdev": 0.49041999999999997,
        "percentiles": {
          "p50": 0.232823,
          "p75": 0.263551,
          "p80": 0.271967,
          "p90": 0.297775,
          "p95": 0.336623,
          "p99": 0.8462390000000001,
          "p999": 2.984191
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 149.25,
            "min": 138.42578125,
            "mean": 146.38606770833334
          },
          "cpu": {
            "max": 0.8000000000000009,
            "min": 0.40000000000000036,
            "mean": 0.5333333333333332
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 29.734375,
            "min": 24.8046875,
            "mean": 28.7671875
          },
          "cpu": {
            "max": 28.15605776291101,
            "min": 13.825065221598607,
            "mean": 21.54805928162631
          }
        }
      },
      "poolOverflow": 55,
      "upstreamConnections": 28,
      "counters": {
        "benchmark.http_2xx": {
          "value": 179945,
          "perSecond": 2021.8539325842696
        },
        "benchmark.pool_overflow": {
          "value": 55,
          "perSecond": 0.6179775280898876
        },
        "cluster_manager.cluster_added": {
          "value": 4,
          "perSecond": 0.0449438202247191
        },
        "default.total_match_count": {
          "value": 4,
          "perSecond": 0.0449438202247191
        },
        "membership_change": {
          "value": 4,
          "perSecond": 0.0449438202247191
        },
        "runtime.load_success": {
          "value": 1,
          "perSecond": 0.011235955056179775
        },
        "runtime.override_dir_not_exists": {
          "value": 1,
          "perSecond": 0.011235955056179775
        },
        "upstream_cx_http1_total": {
          "value": 28,
          "perSecond": 0.3146067415730337
        },
        "upstream_cx_rx_bytes_total": {
          "value": 28251365,
          "perSecond": 317431.06741573033
        },
        "upstream_cx_total": {
          "value": 28,
          "perSecond": 0.3146067415730337
        },
        "upstream_cx_tx_bytes_total": {
          "value": 8097525,
          "perSecond": 90983.42696629213
        },
        "upstream_rq_pending_overflow": {
          "value": 55,
          "perSecond": 0.6179775280898876
        },
        "upstream_rq_pending_total": {
          "value": 28,
          "perSecond": 0.3146067415730337
        },
        "upstream_rq_total": {
          "value": 179945,
          "perSecond": 2021.8539325842696
        }
      }
    },
    {
      "testName": "scaling up httproutes to 300 with 60 routes per hostname at 800 rps",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-up",
      "throughput": 3199.8333333333335,
      "totalRequests": 287985,
      "latency": {
        "max": 18.610174999999998,
        "min": 0.130056,
        "mean": 0.257161,
        "pstdev": 0.336211,
        "percentiles": {
          "p50": 0.212215,
          "p75": 0.242855,
          "p80": 0.251927,
          "p90": 0.28203100000000003,
          "p95": 0.37388699999999997,
          "p99": 1.207679,
          "p999": 4.993278999999999
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 165.03125,
            "min": 151.73046875,
            "mean": 162.00026041666666
          },
          "cpu": {
            "max": 9.266666666666671,
            "min": 0.5999999999999991,
            "mean": 2.033333333333335
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 44.1953125,
            "min": 41.078125,
            "mean": 43.25364583333333
          },
          "cpu": {
            "max": 37.00084581253465,
            "min": 36.41505769772024,
            "mean": 36.74979376332847
          }
        }
      },
      "poolOverflow": 15,
      "upstreamConnections": 50,
      "counters": {
        "benchmark.http_2xx": {
          "value": 287985,
          "perSecond": 3199.8333333333335
        },
        "benchmark.pool_overflow": {
          "value": 15,
          "perSecond": 0.16666666666666666
        },
        "cluster_manager.cluster_added": {
          "value": 4,
          "perSecond": 0.044444444444444446
        },
        "default.total_match_count": {
          "value": 4,
          "perSecond": 0.044444444444444446
        },
        "membership_change": {
          "value": 4,
          "perSecond": 0.044444444444444446
        },
        "runtime.load_success": {
          "value": 1,
          "perSecond": 0.011111111111111112
        },
        "runtime.override_dir_not_exists": {
          "value": 1,
          "perSecond": 0.011111111111111112
        },
        "upstream_cx_http1_total": {
          "value": 50,
          "perSecond": 0.5555555555555556
        },
        "upstream_cx_rx_bytes_total": {
          "value": 45213645,
          "perSecond": 502373.8333333333
        },
        "upstream_cx_total": {
          "value": 50,
          "perSecond": 0.5555555555555556
        },
        "upstream_cx_tx_bytes_total": {
          "value": 12959325,
          "perSecond": 143992.5
        },
        "upstream_rq_pending_overflow": {
          "value": 15,
          "perSecond": 0.16666666666666666
        },
        "upstream_rq_pending_total": {
          "value": 50,
          "perSecond": 0.5555555555555556
        },
        "upstream_rq_total": {
          "value": 287985,
          "perSecond": 3199.8333333333335
        }
      }
    },
    {
      "testName": "scaling up httproutes to 500 with 100 routes per hostname at 1000 rps",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-up",
      "throughput": 3998.7555555555555,
      "totalRequests": 359888,
      "latency": {
        "max": 33.714175,
        "min": 0.13023200000000001,
        "mean": 0.28463,
        "pstdev": 0.429191,
        "percentiles": {
          "p50": 0.214415,
          "p75": 0.24806300000000003,
          "p80": 0.261535,
          "p90": 0.365423,
          "p95": 0.512991,
          "p99": 1.732543,
          "p999": 5.754879
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 183.50390625,
            "min": 176.25390625,
            "mean": 181.03190104166666
          },
          "cpu": {
            "max": 8.400000000000034,
            "min": 0.5999999999999753,
            "mean": 1.9866666666666686
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 58.30078125,
            "min": 55.27734375,
            "mean": 57.44296875
          },
          "cpu": {
            "max": 46.98157847318155,
            "min": 28.669390565200814,
            "mean": 43.0385994033216
          }
        }
      },
      "poolOverflow": 112,
      "upstreamConnections": 69,
      "counters": {
        "benchmark.http_2xx": {
          "value": 359888,
          "perSecond": 3998.7555555555555
        },
        "benchmark.pool_overflow": {
          "value": 112,
          "perSecond": 1.2444444444444445
        },
        "cluster_manager.cluster_added": {
          "value": 4,
          "perSecond": 0.044444444444444446
        },
        "default.total_match_count": {
          "value": 4,
          "perSecond": 0.044444444444444446
        },
        "membership_change": {
          "value": 4,
          "perSecond": 0.044444444444444446
        },
        "runtime.load_success": {
          "value": 1,
          "perSecond": 0.011111111111111112
        },
        "runtime.override_dir_not_exists": {
          "value": 1,
          "perSecond": 0.011111111111111112
        },
        "upstream_cx_http1_total": {
          "value": 69,
          "perSecond": 0.7666666666666667
        },
        "upstream_cx_rx_bytes_total": {
          "value": 56502416,
          "perSecond": 627804.6222222223
        },
        "upstream_cx_total": {
          "value": 69,
          "perSecond": 0.7666666666666667
        },
        "upstream_cx_tx_bytes_total": {
          "value": 16194960,
          "perSecond": 179944
        },
        "upstream_rq_pending_overflow": {
          "value": 112,
          "perSecond": 1.2444444444444445
        },
        "upstream_rq_pending_total": {
          "value": 69,
          "perSecond": 0.7666666666666667
        },
        "upstream_rq_total": {
          "value": 359888,
          "perSecond": 3998.7555555555555
        }
      }
    },
    {
      "testName": "scaling up httproutes to 1000 with 200 routes per hostname at 2000 rps",
      "routes": 1000,
      "routesPerHostname": 200,
      "phase": "scaling-up",
      "throughput": 7999.233333333334,
      "totalRequests": 719931,
      "latency": {
        "max": 216.391679,
        "min": 0.11942,
        "mean": 1.1570809999999998,
        "pstdev": 5.035838,
        "percentiles": {
          "p50": 0.32910300000000003,
          "p75": 0.5369269999999999,
          "p80": 0.656799,
          "p90": 1.5603189999999998,
          "p95": 3.5979509999999997,
          "p99": 18.452479,
          "p999": 71.42195099999999
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 252.1328125,
            "min": 229.0078125,
            "mean": 242.77291666666667
          },
          "cpu": {
            "max": 10.666666666666629,
            "min": 0.7333333333332348,
            "mean": 2.3777777777778013
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 103.4921875,
            "min": 88.11328125,
            "mean": 97.62916666666666
          },
          "cpu": {
            "max": 79.28519733416647,
            "min": 54.769053378168884,
            "mean": 76.88908646510222
          }
        }
      },
      "poolOverflow": 67,
      "upstreamConnections": 333,
      "counters": {
        "benchmark.http_2xx": {
          "value": 719931,
          "perSecond": 7999.233333333334
        },
        "benchmark.pool_overflow": {
          "value": 67,
          "perSecond": 0.7444444444444445
        },
        "cluster_manager.cluster_added": {
          "value": 4,
          "perSecond": 0.044444444444444446
        },
        "default.total_match_count": {
          "value": 4,
          "perSecond": 0.044444444444444446
        },
        "membership_change": {
          "value": 4,
          "perSecond": 0.044444444444444446
        },
        "runtime.load_success": {
          "value": 1,
          "perSecond": 0.011111111111111112
        },
        "runtime.override_dir_not_exists": {
          "value": 1,
          "perSecond": 0.011111111111111112
        },
        "upstream_cx_http1_total": {
          "value": 333,
          "perSecond": 3.7
        },
        "upstream_cx_rx_bytes_total": {
          "value": 113029167,
          "perSecond": 1255879.6333333333
        },
        "upstream_cx_total": {
          "value": 333,
          "perSecond": 3.7
        },
        "upstream_cx_tx_bytes_total": {
          "value": 32396985,
          "perSecond": 359966.5
        },
        "upstream_rq_pending_overflow": {
          "value": 67,
          "perSecond": 0.7444444444444445
        },
        "upstream_rq_pending_total": {
          "value": 333,
          "perSecond": 3.7
        },
        "upstream_rq_total": {
          "value": 719933,
          "perSecond": 7999.2555555555555
        }
      }
    },
    {
      "testName": "scaling down httproutes to 500 with 100 routes per hostname at 1000 rps",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-down",
      "throughput": 3996.4444444444443,
      "totalRequests": 359680,
      "latency": {
        "max": 244.98175899999998,
        "min": 0.132616,
        "mean": 0.409257,
        "pstdev": 2.321205,
        "percentiles": {
          "p50": 0.21848700000000001,
          "p75": 0.262095,
          "p80": 0.285423,
          "p90": 0.467391,
          "p95": 0.920671,
          "p99": 4.077567,
          "p999": 14.153215
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 331.52734375,
            "min": 189.8671875,
            "mean": 233.81184895833334
          },
          "cpu": {
            "max": 6.2000000000000455,
            "min": 0.8666666666666363,
            "mean": 2.9199999999999973
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 104.3046875,
            "min": 103.33984375,
            "mean": 103.86223958333333
          },
          "cpu": {
            "max": 49.45569810063322,
            "min": 29.27915108605534,
            "mean": 41.849868656358545
          }
        }
      },
      "poolOverflow": 320,
      "upstreamConnections": 80,
      "counters": {
        "benchmark.http_2xx": {
          "value": 359680,
          "perSecond": 3996.4444444444443
        },
        "benchmark.pool_overflow": {
          "value": 320,
          "perSecond": 3.5555555555555554
        },
        "cluster_manager.cluster_added": {
          "value": 4,
          "perSecond": 0.044444444444444446
        },
        "default.total_match_count": {
          "value": 4,
          "perSecond": 0.044444444444444446
        },
        "membership_change": {
          "value": 4,
          "perSecond": 0.044444444444444446
        },
        "runtime.load_success": {
          "value": 1,
          "perSecond": 0.011111111111111112
        },
        "runtime.override_dir_not_exists": {
          "value": 1,
          "perSecond": 0.011111111111111112
        },
        "upstream_cx_http1_total": {
          "value": 80,
          "perSecond": 0.8888888888888888
        },
        "upstream_cx_rx_bytes_total": {
          "value": 56469760,
          "perSecond": 627441.7777777778
        },
        "upstream_cx_total": {
          "value": 80,
          "perSecond": 0.8888888888888888
        },
        "upstream_cx_tx_bytes_total": {
          "value": 16185600,
          "perSecond": 179840
        },
        "upstream_rq_pending_overflow": {
          "value": 320,
          "perSecond": 3.5555555555555554
        },
        "upstream_rq_pending_total": {
          "value": 80,
          "perSecond": 0.8888888888888888
        },
        "upstream_rq_total": {
          "value": 359680,
          "perSecond": 3996.4444444444443
        }
      }
    },
    {
      "testName": "scaling down httproutes to 300 with 60 routes per hostname at 800 rps",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-down",
      "throughput": 3196.177777777778,
      "totalRequests": 287656,
      "latency": {
        "max": 93.52806299999999,
        "min": 0.13303199999999998,
        "mean": 0.33526999999999996,
        "pstdev": 1.242871,
        "percentiles": {
          "p50": 0.222599,
          "p75": 0.261791,
          "p80": 0.272383,
          "p90": 0.33639899999999995,
          "p95": 0.579167,
          "p99": 2.327679,
          "p999": 17.328127
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 216.12109375,
            "min": 166.75,
            "mean": 179.96809895833334
          },
          "cpu": {
            "max": 7.533333333333303,
            "min": 0.8666666666666363,
            "mean": 2.5066666666666606
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 104.015625,
            "min": 103.578125,
            "mean": 103.8
          },
          "cpu": {
            "max": 43.39434507345285,
            "min": 37.94940777065254,
            "mean": 42.03311074775277
          }
        }
      },
      "poolOverflow": 344,
      "upstreamConnections": 56,
      "counters": {
        "benchmark.http_2xx": {
          "value": 287656,
          "perSecond": 3196.177777777778
        },
        "benchmark.pool_overflow": {
          "value": 344,
          "perSecond": 3.8222222222222224
        },
        "cluster_manager.cluster_added": {
          "value": 4,
          "perSecond": 0.044444444444444446
        },
        "default.total_match_count": {
          "value": 4,
          "perSecond": 0.044444444444444446
        },
        "membership_change": {
          "value": 4,
          "perSecond": 0.044444444444444446
        },
        "runtime.load_success": {
          "value": 1,
          "perSecond": 0.011111111111111112
        },
        "runtime.override_dir_not_exists": {
          "value": 1,
          "perSecond": 0.011111111111111112
        },
        "upstream_cx_http1_total": {
          "value": 56,
          "perSecond": 0.6222222222222222
        },
        "upstream_cx_rx_bytes_total": {
          "value": 45161992,
          "perSecond": 501799.9111111111
        },
        "upstream_cx_total": {
          "value": 56,
          "perSecond": 0.6222222222222222
        },
        "upstream_cx_tx_bytes_total": {
          "value": 12944520,
          "perSecond": 143828
        },
        "upstream_rq_pending_overflow": {
          "value": 344,
          "perSecond": 3.8222222222222224
        },
        "upstream_rq_pending_total": {
          "value": 56,
          "perSecond": 0.6222222222222222
        },
        "upstream_rq_total": {
          "value": 287656,
          "perSecond": 3196.177777777778
        }
      }
    },
    {
      "testName": "scaling down httproutes to 100 with 20 routes per hostname at 500 rps",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-down",
      "throughput": 1997.9888888888888,
      "totalRequests": 179819,
      "latency": {
        "max": 66.709503,
        "min": 0.15204800000000002,
        "mean": 0.30158799999999997,
        "pstdev": 0.923008,
        "percentiles": {
          "p50": 0.23566299999999998,
          "p75": 0.267775,
          "p80": 0.276943,
          "p90": 0.31847899999999996,
          "p95": 0.404175,
          "p99": 1.499839,
          "p999": 9.945087000000001
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 176.2578125,
            "min": 152.75,
            "mean": 157.94270833333334
          },
          "cpu": {
            "max": 5.33333333333322,
            "min": 0.8000000000000304,
            "mean": 1.8111111111110907
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 104.09765625,
            "min": 103.25,
            "mean": 103.74075520833334
          },
          "cpu": {
            "max": 28.076321673278116,
            "min": 15.20770821331095,
            "mean": 23.141733405882853
          }
        }
      },
      "poolOverflow": 181,
      "upstreamConnections": 51,
      "counters": {
        "benchmark.http_2xx": {
          "value": 179819,
          "perSecond": 1997.9888888888888
        },
        "benchmark.pool_overflow": {
          "value": 181,
          "perSecond": 2.011111111111111
        },
        "cluster_manager.cluster_added": {
          "value": 4,
          "perSecond": 0.044444444444444446
        },
        "default.total_match_count": {
          "value": 4,
          "perSecond": 0.044444444444444446
        },
        "membership_change": {
          "value": 4,
          "perSecond": 0.044444444444444446
        },
        "runtime.load_success": {
          "value": 1,
          "perSecond": 0.011111111111111112
        },
        "runtime.override_dir_not_exists": {
          "value": 1,
          "perSecond": 0.011111111111111112
        },
        "upstream_cx_http1_total": {
          "value": 51,
          "perSecond": 0.5666666666666667
        },
        "upstream_cx_rx_bytes_total": {
          "value": 28231583,
          "perSecond": 313684.2555555556
        },
        "upstream_cx_total": {
          "value": 51,
          "perSecond": 0.5666666666666667
        },
        "upstream_cx_tx_bytes_total": {
          "value": 8091855,
          "perSecond": 89909.5
        },
        "upstream_rq_pending_overflow": {
          "value": 181,
          "perSecond": 2.011111111111111
        },
        "upstream_rq_pending_total": {
          "value": 51,
          "perSecond": 0.5666666666666667
        },
        "upstream_rq_total": {
          "value": 179819,
          "perSecond": 1997.9888888888888
        }
      }
    },
    {
      "testName": "scaling down httproutes to 50 with 10 routes per hostname at 300 rps",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-down",
      "throughput": 1212.8651685393259,
      "totalRequests": 107945,
      "latency": {
        "max": 58.984447,
        "min": 0.16778400000000002,
        "mean": 0.301583,
        "pstdev": 0.542385,
        "percentiles": {
          "p50": 0.260631,
          "p75": 0.288447,
          "p80": 0.297711,
          "p90": 0.33209500000000003,
          "p95": 0.39150300000000005,
          "p99": 1.1170550000000001,
          "p999": 7.045375000000001
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 158.91796875,
            "min": 145.0234375,
            "mean": 152.47955729166668
          },
          "cpu": {
            "max": 0.9333333333334317,
            "min": 0.8666666666666363,
            "mean": 0.8818181818181811
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 103.93359375,
            "min": 103.25,
            "mean": 103.64674479166666
          },
          "cpu": {
            "max": 18.17455505715532,
            "min": 17.865555173599184,
            "mean": 18.04212653563126
          }
        }
      },
      "poolOverflow": 55,
      "upstreamConnections": 27,
      "counters": {
        "benchmark.http_2xx": {
          "value": 107945,
          "perSecond": 1212.8651685393259
        },
        "benchmark.pool_overflow": {
          "value": 55,
          "perSecond": 0.6179775280898876
        },
        "cluster_manager.cluster_added": {
          "value": 4,
          "perSecond": 0.0449438202247191
        },
        "default.total_match_count": {
          "value": 4,
          "perSecond": 0.0449438202247191
        },
        "membership_change": {
          "value": 4,
          "perSecond": 0.0449438202247191
        },
        "runtime.load_success": {
          "value": 1,
          "perSecond": 0.011235955056179775
        },
        "runtime.override_dir_not_exists": {
          "value": 1,
          "perSecond": 0.011235955056179775
        },
        "upstream_cx_http1_total": {
          "value": 27,
          "perSecond": 0.30337078651685395
        },
        "upstream_cx_rx_bytes_total": {
          "value": 16947365,
          "perSecond": 190419.83146067415
        },
        "upstream_cx_total": {
          "value": 27,
          "perSecond": 0.30337078651685395
        },
        "upstream_cx_tx_bytes_total": {
          "value": 4857525,
          "perSecond": 54578.93258426966
        },
        "upstream_rq_pending_overflow": {
          "value": 55,
          "perSecond": 0.6179775280898876
        },
        "upstream_rq_pending_total": {
          "value": 27,
          "perSecond": 0.30337078651685395
        },
        "upstream_rq_total": {
          "value": 107945,
          "perSecond": 1212.8651685393259
        }
      }
    },
    {
      "testName": "scaling down httproutes to 10 with 2 routes per hostname at 100 rps",
      "routes": 10,
      "routesPerHostname": 2,
      "phase": "scaling-down",
      "throughput": 404.438202247191,
      "totalRequests": 35995,
      "latency": {
        "max": 40.247295,
        "min": 0.195208,
        "mean": 0.325752,
        "pstdev": 0.62766,
        "percentiles": {
          "p50": 0.28579099999999996,
          "p75": 0.316415,
          "p80": 0.326479,
          "p90": 0.35884699999999997,
          "p95": 0.400335,
          "p99": 0.970047,
          "p999": 4.993023
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 156.3359375,
            "min": 147.46875,
            "mean": 150.13333333333333
          },
          "cpu": {
            "max": 1.3333333333332575,
            "min": 0.8000000000000304,
            "mean": 0.9916666666666413
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 103.7421875,
            "min": 103.28515625,
            "mean": 103.41015625
          },
          "cpu": {
            "max": 7.12140828707879,
            "min": 3.7227436994447847,
            "mean": 6.34939668103893
          }
        }
      },
      "poolOverflow": 5,
      "upstreamConnections": 11,
      "counters": {
        "benchmark.http_2xx": {
          "value": 35995,
          "perSecond": 404.438202247191
        },
        "benchmark.pool_overflow": {
          "value": 5,
          "perSecond": 0.056179775280898875
        },
        "cluster_manager.cluster_added": {
          "value": 4,
          "perSecond": 0.0449438202247191
        },
        "default.total_match_count": {
          "value": 4,
          "perSecond": 0.0449438202247191
        },
        "membership_change": {
          "value": 4,
          "perSecond": 0.0449438202247191
        },
        "runtime.load_success": {
          "value": 1,
          "perSecond": 0.011235955056179775
        },
        "runtime.override_dir_not_exists": {
          "value": 1,
          "perSecond": 0.011235955056179775
        },
        "upstream_cx_http1_total": {
          "value": 11,
          "perSecond": 0.12359550561797752
        },
        "upstream_cx_rx_bytes_total": {
          "value": 5651215,
          "perSecond": 63496.79775280899
        },
        "upstream_cx_total": {
          "value": 11,
          "perSecond": 0.12359550561797752
        },
        "upstream_cx_tx_bytes_total": {
          "value": 1619775,
          "perSecond": 18199.719101123595
        },
        "upstream_rq_pending_overflow": {
          "value": 5,
          "perSecond": 0.056179775280898875
        },
        "upstream_rq_pending_total": {
          "value": 11,
          "perSecond": 0.12359550561797752
        },
        "upstream_rq_total": {
          "value": 35995,
          "perSecond": 404.438202247191
        }
      }
    }
  ]
};
