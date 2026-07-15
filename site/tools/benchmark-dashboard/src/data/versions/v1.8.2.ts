import { TestSuite } from '../types';

// Benchmark data extracted from release artifact for version 1.8.2
// Generated from benchmark_result.json

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.8.2",
    "runId": "1.8.2-release-2026-07-01",
    "date": "2026-07-01T07:14:42Z",
    "environment": "GitHub Release",
    "description": "Benchmark results for version 1.8.2 from release artifacts",
    "downloadUrl": "https://github.com/envoyproxy/gateway/releases/download/v1.8.2/benchmark_report.zip",
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
        "max": 36.020222999999994,
        "min": 0.312,
        "mean": 0.41179699999999997,
        "pstdev": 0.449004,
        "percentiles": {
          "p50": 0.380687,
          "p75": 0.39697499999999997,
          "p80": 0.40220700000000004,
          "p90": 0.422943,
          "p95": 0.457727,
          "p99": 0.942815,
          "p999": 5.765119
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 132.5859375,
            "min": 106.25390625,
            "mean": 125.273828125
          },
          "cpu": {
            "max": 0.8,
            "min": 0.26666666666666655,
            "mean": 0.4357142857142859
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 20.3125,
            "min": 7.51171875,
            "mean": 17.907462284482758
          },
          "cpu": {
            "max": 8.761964562188197,
            "min": 5.422322740352064,
            "mean": 7.468832454853503
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
      "throughput": 1199.4333333333334,
      "totalRequests": 107949,
      "latency": {
        "max": 64.907263,
        "min": 0.27233599999999997,
        "mean": 0.38906599999999997,
        "pstdev": 0.7291580000000001,
        "percentiles": {
          "p50": 0.351375,
          "p75": 0.366735,
          "p80": 0.371327,
          "p90": 0.38687900000000003,
          "p95": 0.417103,
          "p99": 1.061375,
          "p999": 6.537727
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 136.16796875,
            "min": 129.78125,
            "mean": 133.76419270833333
          },
          "cpu": {
            "max": 0.46666666666666706,
            "min": 0.40000000000000036,
            "mean": 0.4515151515151514
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 24.875,
            "min": 20.3046875,
            "mean": 24.182291666666668
          },
          "cpu": {
            "max": 23.617121248692584,
            "min": 15.34443729396305,
            "mean": 21.76731003192191
          }
        }
      },
      "poolOverflow": 51,
      "upstreamConnections": 37,
      "counters": {
        "benchmark.http_2xx": {
          "value": 107949,
          "perSecond": 1199.4333333333334
        },
        "benchmark.pool_overflow": {
          "value": 51,
          "perSecond": 0.5666666666666667
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
          "value": 37,
          "perSecond": 0.4111111111111111
        },
        "upstream_cx_rx_bytes_total": {
          "value": 16947993,
          "perSecond": 188311.03333333333
        },
        "upstream_cx_total": {
          "value": 37,
          "perSecond": 0.4111111111111111
        },
        "upstream_cx_tx_bytes_total": {
          "value": 4857705,
          "perSecond": 53974.5
        },
        "upstream_rq_pending_overflow": {
          "value": 51,
          "perSecond": 0.5666666666666667
        },
        "upstream_rq_pending_total": {
          "value": 37,
          "perSecond": 0.4111111111111111
        },
        "upstream_rq_total": {
          "value": 107949,
          "perSecond": 1199.4333333333334
        }
      }
    },
    {
      "testName": "scaling up httproutes to 100 with 20 routes per hostname at 500 rps",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-up",
      "throughput": 2021.5168539325844,
      "totalRequests": 179915,
      "latency": {
        "max": 127.25452700000001,
        "min": 0.263232,
        "mean": 0.378263,
        "pstdev": 1.1435600000000001,
        "percentiles": {
          "p50": 0.332655,
          "p75": 0.354879,
          "p80": 0.360607,
          "p90": 0.377151,
          "p95": 0.422591,
          "p99": 1.157119,
          "p999": 4.588799
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 137.95703125,
            "min": 128.9453125,
            "mean": 134.10416666666666
          },
          "cpu": {
            "max": 4.800000000000002,
            "min": 0.3999999999999974,
            "mean": 1.335897435897436
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 28.83984375,
            "min": 24.78515625,
            "mean": 27.861458333333335
          },
          "cpu": {
            "max": 36.54449081803006,
            "min": 35.403415543690116,
            "mean": 36.08806070829409
          }
        }
      },
      "poolOverflow": 85,
      "upstreamConnections": 42,
      "counters": {
        "benchmark.http_2xx": {
          "value": 179915,
          "perSecond": 2021.5168539325844
        },
        "benchmark.pool_overflow": {
          "value": 85,
          "perSecond": 0.9550561797752809
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
          "value": 42,
          "perSecond": 0.47191011235955055
        },
        "upstream_cx_rx_bytes_total": {
          "value": 28246655,
          "perSecond": 317378.1460674157
        },
        "upstream_cx_total": {
          "value": 42,
          "perSecond": 0.47191011235955055
        },
        "upstream_cx_tx_bytes_total": {
          "value": 8096175,
          "perSecond": 90968.25842696629
        },
        "upstream_rq_pending_overflow": {
          "value": 85,
          "perSecond": 0.9550561797752809
        },
        "upstream_rq_pending_total": {
          "value": 42,
          "perSecond": 0.47191011235955055
        },
        "upstream_rq_total": {
          "value": 179915,
          "perSecond": 2021.5168539325844
        }
      }
    },
    {
      "testName": "scaling up httproutes to 300 with 60 routes per hostname at 800 rps",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-up",
      "throughput": 3199.233333333333,
      "totalRequests": 287931,
      "latency": {
        "max": 44.654591,
        "min": 0.227192,
        "mean": 0.416971,
        "pstdev": 0.692846,
        "percentiles": {
          "p50": 0.325727,
          "p75": 0.35556699999999997,
          "p80": 0.37772700000000003,
          "p90": 0.482255,
          "p95": 0.652991,
          "p99": 2.004095,
          "p999": 9.937918999999999
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 152.89453125,
            "min": 142.5703125,
            "mean": 148.78268229166667
          },
          "cpu": {
            "max": 23.06666666666667,
            "min": 0.5999999999999991,
            "mean": 4.562222222222224
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 45.5,
            "min": 28.19921875,
            "mean": 40.841796875
          },
          "cpu": {
            "max": 56.956964607078575,
            "min": 55.81466109688724,
            "mean": 56.115334620281175
          }
        }
      },
      "poolOverflow": 69,
      "upstreamConnections": 91,
      "counters": {
        "benchmark.http_2xx": {
          "value": 287931,
          "perSecond": 3199.233333333333
        },
        "benchmark.pool_overflow": {
          "value": 69,
          "perSecond": 0.7666666666666667
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
          "value": 91,
          "perSecond": 1.011111111111111
        },
        "upstream_cx_rx_bytes_total": {
          "value": 45205167,
          "perSecond": 502279.63333333336
        },
        "upstream_cx_total": {
          "value": 91,
          "perSecond": 1.011111111111111
        },
        "upstream_cx_tx_bytes_total": {
          "value": 12956895,
          "perSecond": 143965.5
        },
        "upstream_rq_pending_overflow": {
          "value": 69,
          "perSecond": 0.7666666666666667
        },
        "upstream_rq_pending_total": {
          "value": 91,
          "perSecond": 1.011111111111111
        },
        "upstream_rq_total": {
          "value": 287931,
          "perSecond": 3199.233333333333
        }
      }
    },
    {
      "testName": "scaling up httproutes to 500 with 100 routes per hostname at 1000 rps",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-up",
      "throughput": 3998.511111111111,
      "totalRequests": 359866,
      "latency": {
        "max": 107.909119,
        "min": 0.213488,
        "mean": 0.846032,
        "pstdev": 3.164183,
        "percentiles": {
          "p50": 0.46446299999999996,
          "p75": 0.518143,
          "p80": 0.548575,
          "p90": 0.7605430000000001,
          "p95": 1.322303,
          "p99": 9.396735,
          "p999": 49.727487
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 169,
            "min": 151.01953125,
            "mean": 165.20091145833334
          },
          "cpu": {
            "max": 32.93333333333333,
            "min": 0.5999999999999991,
            "mean": 6.317777777777774
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 64.9921875,
            "min": 44.9375,
            "mean": 58.34322916666667
          },
          "cpu": {
            "max": 67.03195836840888,
            "min": 46.42560149423007,
            "mean": 63.367608419813365
          }
        }
      },
      "poolOverflow": 134,
      "upstreamConnections": 253,
      "counters": {
        "benchmark.http_2xx": {
          "value": 359866,
          "perSecond": 3998.511111111111
        },
        "benchmark.pool_overflow": {
          "value": 134,
          "perSecond": 1.488888888888889
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
          "value": 253,
          "perSecond": 2.811111111111111
        },
        "upstream_cx_rx_bytes_total": {
          "value": 56498962,
          "perSecond": 627766.2444444444
        },
        "upstream_cx_total": {
          "value": 253,
          "perSecond": 2.811111111111111
        },
        "upstream_cx_tx_bytes_total": {
          "value": 16193970,
          "perSecond": 179933
        },
        "upstream_rq_pending_overflow": {
          "value": 134,
          "perSecond": 1.488888888888889
        },
        "upstream_rq_pending_total": {
          "value": 253,
          "perSecond": 2.811111111111111
        },
        "upstream_rq_total": {
          "value": 359866,
          "perSecond": 3998.511111111111
        }
      }
    },
    {
      "testName": "scaling up httproutes to 1000 with 200 routes per hostname at 2000 rps",
      "routes": 1000,
      "routesPerHostname": 200,
      "phase": "scaling-up",
      "throughput": 7443.9111111111115,
      "totalRequests": 669952,
      "latency": {
        "max": 584.450047,
        "min": 0.286336,
        "mean": 42.145749,
        "pstdev": 11.921429000000002,
        "percentiles": {
          "p50": 41.594879,
          "p75": 48.228351,
          "p80": 50.016255,
          "p90": 54.749183,
          "p95": 59.179007,
          "p99": 74.428415,
          "p999": 106.749951
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 188.53125,
            "min": 166.06640625,
            "mean": 183.54153645833333
          },
          "cpu": {
            "max": 67.19999999999999,
            "min": 0.7333333333333295,
            "mean": 13.73333333333333
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 102.2421875,
            "min": 91.796875,
            "mean": 98.63997395833333
          },
          "cpu": {
            "max": 99.32751039159913,
            "min": 98.86982770503084,
            "mean": 99.02721402466419
          }
        }
      },
      "poolOverflow": 77,
      "upstreamConnections": 323,
      "counters": {
        "benchmark.http_2xx": {
          "value": 669952,
          "perSecond": 7443.9111111111115
        },
        "benchmark.pool_overflow": {
          "value": 77,
          "perSecond": 0.8555555555555555
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
          "value": 323,
          "perSecond": 3.588888888888889
        },
        "upstream_cx_rx_bytes_total": {
          "value": 105182464,
          "perSecond": 1168694.0444444444
        },
        "upstream_cx_total": {
          "value": 323,
          "perSecond": 3.588888888888889
        },
        "upstream_cx_tx_bytes_total": {
          "value": 30162375,
          "perSecond": 335137.5
        },
        "upstream_rq_pending_overflow": {
          "value": 77,
          "perSecond": 0.8555555555555555
        },
        "upstream_rq_pending_total": {
          "value": 323,
          "perSecond": 3.588888888888889
        },
        "upstream_rq_total": {
          "value": 670275,
          "perSecond": 7447.5
        }
      }
    },
    {
      "testName": "scaling down httproutes to 500 with 100 routes per hostname at 1000 rps",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-down",
      "throughput": 3999.511111111111,
      "totalRequests": 359956,
      "latency": {
        "max": 69.97196699999999,
        "min": 0.222976,
        "mean": 0.789628,
        "pstdev": 2.5729770000000003,
        "percentiles": {
          "p50": 0.466191,
          "p75": 0.525215,
          "p80": 0.568799,
          "p90": 0.8516469999999999,
          "p95": 1.4764789999999999,
          "p99": 6.115327000000001,
          "p999": 45.881343
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 169.31640625,
            "min": 165.83984375,
            "mean": 167.48033854166667
          },
          "cpu": {
            "max": 0.9333333333333372,
            "min": 0.8666666666666363,
            "mean": 0.904347826086966
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 102.65625,
            "min": 100.359375,
            "mean": 101.98138020833333
          },
          "cpu": {
            "max": 67.3141865613595,
            "min": 47.342977440931314,
            "mean": 65.13734998065154
          }
        }
      },
      "poolOverflow": 44,
      "upstreamConnections": 258,
      "counters": {
        "benchmark.http_2xx": {
          "value": 359956,
          "perSecond": 3999.511111111111
        },
        "benchmark.pool_overflow": {
          "value": 44,
          "perSecond": 0.4888888888888889
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
          "value": 258,
          "perSecond": 2.8666666666666667
        },
        "upstream_cx_rx_bytes_total": {
          "value": 56513092,
          "perSecond": 627923.2444444444
        },
        "upstream_cx_total": {
          "value": 258,
          "perSecond": 2.8666666666666667
        },
        "upstream_cx_tx_bytes_total": {
          "value": 16198020,
          "perSecond": 179978
        },
        "upstream_rq_pending_overflow": {
          "value": 44,
          "perSecond": 0.4888888888888889
        },
        "upstream_rq_pending_total": {
          "value": 258,
          "perSecond": 2.8666666666666667
        },
        "upstream_rq_total": {
          "value": 359956,
          "perSecond": 3999.511111111111
        }
      }
    },
    {
      "testName": "scaling down httproutes to 300 with 60 routes per hostname at 800 rps",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-down",
      "throughput": 3233.7303370786517,
      "totalRequests": 287802,
      "latency": {
        "max": 67.01670299999999,
        "min": 0.201728,
        "mean": 0.495236,
        "pstdev": 0.942877,
        "percentiles": {
          "p50": 0.378687,
          "p75": 0.482031,
          "p80": 0.506575,
          "p90": 0.632319,
          "p95": 0.808319,
          "p99": 2.2458869999999997,
          "p999": 12.598783000000001
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 169.31640625,
            "min": 151.1328125,
            "mean": 157.56510416666666
          },
          "cpu": {
            "max": 1.066666666666644,
            "min": 0.8666666666666363,
            "mean": 0.9333333333333373
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 102.3828125,
            "min": 101.3515625,
            "mean": 101.960546875
          },
          "cpu": {
            "max": 54.84321253527425,
            "min": 39.7540166597358,
            "mean": 53.0728229770368
          }
        }
      },
      "poolOverflow": 197,
      "upstreamConnections": 87,
      "counters": {
        "benchmark.http_2xx": {
          "value": 287802,
          "perSecond": 3233.7303370786517
        },
        "benchmark.pool_overflow": {
          "value": 197,
          "perSecond": 2.2134831460674156
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
          "value": 87,
          "perSecond": 0.9775280898876404
        },
        "upstream_cx_rx_bytes_total": {
          "value": 45184914,
          "perSecond": 507695.6629213483
        },
        "upstream_cx_total": {
          "value": 87,
          "perSecond": 0.9775280898876404
        },
        "upstream_cx_tx_bytes_total": {
          "value": 12951090,
          "perSecond": 145517.8651685393
        },
        "upstream_rq_pending_overflow": {
          "value": 197,
          "perSecond": 2.2134831460674156
        },
        "upstream_rq_pending_total": {
          "value": 87,
          "perSecond": 0.9775280898876404
        },
        "upstream_rq_total": {
          "value": 287802,
          "perSecond": 3233.7303370786517
        }
      }
    },
    {
      "testName": "scaling down httproutes to 100 with 20 routes per hostname at 500 rps",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-down",
      "throughput": 2021.0337078651685,
      "totalRequests": 179872,
      "latency": {
        "max": 159.129599,
        "min": 0.248888,
        "mean": 0.388351,
        "pstdev": 1.219139,
        "percentiles": {
          "p50": 0.338255,
          "p75": 0.358415,
          "p80": 0.363711,
          "p90": 0.382207,
          "p95": 0.437503,
          "p99": 1.319487,
          "p999": 6.5948150000000005
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 155.79296875,
            "min": 137.40234375,
            "mean": 145.04557291666666
          },
          "cpu": {
            "max": 0.933333333333337,
            "min": 0.8000000000000304,
            "mean": 0.8666666666666759
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 101.8828125,
            "min": 101.41015625,
            "mean": 101.60481770833333
          },
          "cpu": {
            "max": 22.885429953798287,
            "min": 22.885429953798287,
            "mean": 22.885429953798287
          }
        }
      },
      "poolOverflow": 128,
      "upstreamConnections": 42,
      "counters": {
        "benchmark.http_2xx": {
          "value": 179872,
          "perSecond": 2021.0337078651685
        },
        "benchmark.pool_overflow": {
          "value": 128,
          "perSecond": 1.4382022471910112
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
          "value": 42,
          "perSecond": 0.47191011235955055
        },
        "upstream_cx_rx_bytes_total": {
          "value": 28239904,
          "perSecond": 317302.2921348315
        },
        "upstream_cx_total": {
          "value": 42,
          "perSecond": 0.47191011235955055
        },
        "upstream_cx_tx_bytes_total": {
          "value": 8094240,
          "perSecond": 90946.51685393258
        },
        "upstream_rq_pending_overflow": {
          "value": 128,
          "perSecond": 1.4382022471910112
        },
        "upstream_rq_pending_total": {
          "value": 42,
          "perSecond": 0.47191011235955055
        },
        "upstream_rq_total": {
          "value": 179872,
          "perSecond": 2021.0337078651685
        }
      }
    },
    {
      "testName": "scaling down httproutes to 50 with 10 routes per hostname at 300 rps",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-down",
      "throughput": 1213.3820224719102,
      "totalRequests": 107991,
      "latency": {
        "max": 19.398655,
        "min": 0.266,
        "mean": 0.371821,
        "pstdev": 0.225718,
        "percentiles": {
          "p50": 0.35195099999999996,
          "p75": 0.366655,
          "p80": 0.37027099999999996,
          "p90": 0.382367,
          "p95": 0.409743,
          "p99": 0.9092790000000001,
          "p999": 3.3223670000000003
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 150.1171875,
            "min": 142.828125,
            "mean": 145.550390625
          },
          "cpu": {
            "max": 0.8666666666666363,
            "min": 0.7333333333333295,
            "mean": 0.8285714285714285
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 102.02734375,
            "min": 101.15625,
            "mean": 101.46640625
          },
          "cpu": {
            "max": 23.23340219877513,
            "min": 23.14379762476824,
            "mean": 23.21548128397375
          }
        }
      },
      "poolOverflow": 9,
      "upstreamConnections": 21,
      "counters": {
        "benchmark.http_2xx": {
          "value": 107991,
          "perSecond": 1213.3820224719102
        },
        "benchmark.pool_overflow": {
          "value": 9,
          "perSecond": 0.10112359550561797
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
          "value": 21,
          "perSecond": 0.23595505617977527
        },
        "upstream_cx_rx_bytes_total": {
          "value": 16954587,
          "perSecond": 190500.9775280899
        },
        "upstream_cx_total": {
          "value": 21,
          "perSecond": 0.23595505617977527
        },
        "upstream_cx_tx_bytes_total": {
          "value": 4859595,
          "perSecond": 54602.191011235955
        },
        "upstream_rq_pending_overflow": {
          "value": 9,
          "perSecond": 0.10112359550561797
        },
        "upstream_rq_pending_total": {
          "value": 21,
          "perSecond": 0.23595505617977527
        },
        "upstream_rq_total": {
          "value": 107991,
          "perSecond": 1213.3820224719102
        }
      }
    },
    {
      "testName": "scaling down httproutes to 10 with 2 routes per hostname at 100 rps",
      "routes": 10,
      "routesPerHostname": 2,
      "phase": "scaling-down",
      "throughput": 399.97777777777776,
      "totalRequests": 35998,
      "latency": {
        "max": 29.652991,
        "min": 0.297424,
        "mean": 0.40774499999999997,
        "pstdev": 0.5738810000000001,
        "percentiles": {
          "p50": 0.371663,
          "p75": 0.384431,
          "p80": 0.388447,
          "p90": 0.403167,
          "p95": 0.435055,
          "p99": 0.9181750000000001,
          "p999": 7.789054999999999
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 147.078125,
            "min": 142.0546875,
            "mean": 143.29973958333332
          },
          "cpu": {
            "max": 0.933333333333337,
            "min": 0.8000000000000304,
            "mean": 0.8579710144927519
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 101.5859375,
            "min": 101.1484375,
            "mean": 101.260546875
          },
          "cpu": {
            "max": 8.503712683739975,
            "min": 5.653550666357093,
            "mean": 8.023664163938273
          }
        }
      },
      "poolOverflow": 2,
      "upstreamConnections": 11,
      "counters": {
        "benchmark.http_2xx": {
          "value": 35998,
          "perSecond": 399.97777777777776
        },
        "benchmark.pool_overflow": {
          "value": 2,
          "perSecond": 0.022222222222222223
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
          "value": 11,
          "perSecond": 0.12222222222222222
        },
        "upstream_cx_rx_bytes_total": {
          "value": 5651686,
          "perSecond": 62796.51111111111
        },
        "upstream_cx_total": {
          "value": 11,
          "perSecond": 0.12222222222222222
        },
        "upstream_cx_tx_bytes_total": {
          "value": 1619910,
          "perSecond": 17999
        },
        "upstream_rq_pending_overflow": {
          "value": 2,
          "perSecond": 0.022222222222222223
        },
        "upstream_rq_pending_total": {
          "value": 11,
          "perSecond": 0.12222222222222222
        },
        "upstream_rq_total": {
          "value": 35998,
          "perSecond": 399.97777777777776
        }
      }
    }
  ]
};
