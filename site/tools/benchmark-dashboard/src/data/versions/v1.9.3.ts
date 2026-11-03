import { TestSuite } from '../types';

// Benchmark data extracted from release artifact for version 1.9.3
// Generated from benchmark_result.json

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.9.3",
    "runId": "1.9.3-release-2026-10-26",
    "date": "2026-10-26T09:58:38Z",
    "environment": "GitHub Release",
    "description": "Benchmark results for version 1.9.3 from release artifacts",
    "downloadUrl": "https://github.com/envoyproxy/gateway/releases/download/v1.9.3/benchmark_report.zip",
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
        "max": 86.949887,
        "min": 0.38632,
        "mean": 0.533984,
        "pstdev": 0.773031,
        "percentiles": {
          "p50": 0.474127,
          "p75": 0.497519,
          "p80": 0.506575,
          "p90": 0.547679,
          "p95": 0.6237429999999999,
          "p99": 1.670527,
          "p999": 8.793087
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 137.81640625,
            "min": 117.84375,
            "mean": 132.46614583333334
          },
          "cpu": {
            "max": 1.4666666666666666,
            "min": 0.3333333333333336,
            "mean": 0.6400000000000001
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 20.55078125,
            "min": 6.8046875,
            "mean": 19.016666666666666
          },
          "cpu": {
            "max": 10.96166199337242,
            "min": 7.9546242655367205,
            "mean": 10.639389985814882
          }
        }
      },
      "poolOverflow": 6,
      "upstreamConnections": 14,
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
          "value": 14,
          "perSecond": 0.15730337078651685
        },
        "upstream_cx_rx_bytes_total": {
          "value": 5651058,
          "perSecond": 63495.03370786517
        },
        "upstream_cx_total": {
          "value": 14,
          "perSecond": 0.15730337078651685
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
          "value": 14,
          "perSecond": 0.15730337078651685
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
      "throughput": 1213.191011235955,
      "totalRequests": 107974,
      "latency": {
        "max": 30.068735,
        "min": 0.35344,
        "mean": 0.547377,
        "pstdev": 0.418346,
        "percentiles": {
          "p50": 0.496703,
          "p75": 0.556703,
          "p80": 0.569215,
          "p90": 0.627647,
          "p95": 0.755679,
          "p99": 1.4656630000000002,
          "p999": 3.869951
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 143.40234375,
            "min": 133.8984375,
            "mean": 140.37981770833332
          },
          "cpu": {
            "max": 0.600000000000002,
            "min": 0.46666666666666556,
            "mean": 0.5722222222222217
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 24.95703125,
            "min": 20.1171875,
            "mean": 23.909505208333332
          },
          "cpu": {
            "max": 30.32818612922349,
            "min": 20.18092544667118,
            "mean": 26.382853717672553
          }
        }
      },
      "poolOverflow": 26,
      "upstreamConnections": 19,
      "counters": {
        "benchmark.http_2xx": {
          "value": 107974,
          "perSecond": 1213.191011235955
        },
        "benchmark.pool_overflow": {
          "value": 26,
          "perSecond": 0.29213483146067415
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
          "value": 19,
          "perSecond": 0.21348314606741572
        },
        "upstream_cx_rx_bytes_total": {
          "value": 16951918,
          "perSecond": 190470.98876404495
        },
        "upstream_cx_total": {
          "value": 19,
          "perSecond": 0.21348314606741572
        },
        "upstream_cx_tx_bytes_total": {
          "value": 4858830,
          "perSecond": 54593.59550561798
        },
        "upstream_rq_pending_overflow": {
          "value": 26,
          "perSecond": 0.29213483146067415
        },
        "upstream_rq_pending_total": {
          "value": 19,
          "perSecond": 0.21348314606741572
        },
        "upstream_rq_total": {
          "value": 107974,
          "perSecond": 1213.191011235955
        }
      }
    },
    {
      "testName": "scaling up httproutes to 100 with 20 routes per hostname at 500 rps",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-up",
      "throughput": 1999.4333333333334,
      "totalRequests": 179949,
      "latency": {
        "max": 45.539327,
        "min": 0.345888,
        "mean": 0.5195259999999999,
        "pstdev": 0.44477,
        "percentiles": {
          "p50": 0.461135,
          "p75": 0.48484700000000003,
          "p80": 0.495007,
          "p90": 0.566079,
          "p95": 0.800735,
          "p99": 1.691839,
          "p999": 3.9591670000000003
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 146.4453125,
            "min": 138.2421875,
            "mean": 143.99609375
          },
          "cpu": {
            "max": 3.600000000000001,
            "min": 0.6666666666666643,
            "mean": 1.2487179487179492
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 29.1171875,
            "min": 24.6328125,
            "mean": 28.356640625
          },
          "cpu": {
            "max": 49.494606223588136,
            "min": 36.52454466269687,
            "mean": 47.870622429693036
          }
        }
      },
      "poolOverflow": 51,
      "upstreamConnections": 32,
      "counters": {
        "benchmark.http_2xx": {
          "value": 179949,
          "perSecond": 1999.4333333333334
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
          "value": 32,
          "perSecond": 0.35555555555555557
        },
        "upstream_cx_rx_bytes_total": {
          "value": 28251993,
          "perSecond": 313911.0333333333
        },
        "upstream_cx_total": {
          "value": 32,
          "perSecond": 0.35555555555555557
        },
        "upstream_cx_tx_bytes_total": {
          "value": 8097705,
          "perSecond": 89974.5
        },
        "upstream_rq_pending_overflow": {
          "value": 51,
          "perSecond": 0.5666666666666667
        },
        "upstream_rq_pending_total": {
          "value": 32,
          "perSecond": 0.35555555555555557
        },
        "upstream_rq_total": {
          "value": 179949,
          "perSecond": 1999.4333333333334
        }
      }
    },
    {
      "testName": "scaling up httproutes to 300 with 60 routes per hostname at 800 rps",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-up",
      "throughput": 3199.7444444444445,
      "totalRequests": 287977,
      "latency": {
        "max": 65.908735,
        "min": 0.324768,
        "mean": 0.8322890000000001,
        "pstdev": 1.9964190000000002,
        "percentiles": {
          "p50": 0.613599,
          "p75": 0.7053429999999999,
          "p80": 0.751679,
          "p90": 0.963295,
          "p95": 1.3108469999999999,
          "p99": 3.387775,
          "p999": 40.212479
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 163.91015625,
            "min": 151.06640625,
            "mean": 161.14140625
          },
          "cpu": {
            "max": 16.33333333333333,
            "min": 0.933333333333337,
            "mean": 3.708888888888887
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 48.328125,
            "min": 40.3046875,
            "mean": 46.32369791666667
          },
          "cpu": {
            "max": 71.57732422900884,
            "min": 47.53847596648045,
            "mean": 67.1966056265135
          }
        }
      },
      "poolOverflow": 22,
      "upstreamConnections": 194,
      "counters": {
        "benchmark.http_2xx": {
          "value": 287977,
          "perSecond": 3199.7444444444445
        },
        "benchmark.pool_overflow": {
          "value": 22,
          "perSecond": 0.24444444444444444
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
          "value": 194,
          "perSecond": 2.1555555555555554
        },
        "upstream_cx_rx_bytes_total": {
          "value": 45212389,
          "perSecond": 502359.8777777778
        },
        "upstream_cx_total": {
          "value": 194,
          "perSecond": 2.1555555555555554
        },
        "upstream_cx_tx_bytes_total": {
          "value": 12959010,
          "perSecond": 143989
        },
        "upstream_rq_pending_overflow": {
          "value": 22,
          "perSecond": 0.24444444444444444
        },
        "upstream_rq_pending_total": {
          "value": 194,
          "perSecond": 2.1555555555555554
        },
        "upstream_rq_total": {
          "value": 287978,
          "perSecond": 3199.7555555555555
        }
      }
    },
    {
      "testName": "scaling up httproutes to 500 with 100 routes per hostname at 1000 rps",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-up",
      "throughput": 3998.0555555555557,
      "totalRequests": 359825,
      "latency": {
        "max": 191.889407,
        "min": 0.325056,
        "mean": 1.381377,
        "pstdev": 4.744625,
        "percentiles": {
          "p50": 0.696607,
          "p75": 0.8915190000000001,
          "p80": 0.9794230000000002,
          "p90": 1.4031989999999999,
          "p95": 2.3550709999999997,
          "p99": 22.130686999999998,
          "p999": 64.819199
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 184.546875,
            "min": 179.27734375,
            "mean": 182.73111979166666
          },
          "cpu": {
            "max": 14.866666666666692,
            "min": 0.933333333333337,
            "mean": 3.511111111111122
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 67,
            "min": 57.59375,
            "mean": 63.62669270833333
          },
          "cpu": {
            "max": 51.48676200012127,
            "min": 51.48676200012127,
            "mean": 51.48676200012127
          }
        }
      },
      "poolOverflow": 173,
      "upstreamConnections": 227,
      "counters": {
        "benchmark.http_2xx": {
          "value": 359825,
          "perSecond": 3998.0555555555557
        },
        "benchmark.pool_overflow": {
          "value": 173,
          "perSecond": 1.9222222222222223
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
          "value": 227,
          "perSecond": 2.522222222222222
        },
        "upstream_cx_rx_bytes_total": {
          "value": 56492525,
          "perSecond": 627694.7222222222
        },
        "upstream_cx_total": {
          "value": 227,
          "perSecond": 2.522222222222222
        },
        "upstream_cx_tx_bytes_total": {
          "value": 16192215,
          "perSecond": 179913.5
        },
        "upstream_rq_pending_overflow": {
          "value": 173,
          "perSecond": 1.9222222222222223
        },
        "upstream_rq_pending_total": {
          "value": 227,
          "perSecond": 2.522222222222222
        },
        "upstream_rq_total": {
          "value": 359827,
          "perSecond": 3998.0777777777776
        }
      }
    },
    {
      "testName": "scaling up httproutes to 1000 with 200 routes per hostname at 2000 rps",
      "routes": 1000,
      "routesPerHostname": 200,
      "phase": "scaling-up",
      "throughput": 4717.966666666666,
      "totalRequests": 424617,
      "latency": {
        "max": 1965.228031,
        "min": 1.7781120000000001,
        "mean": 73.709777,
        "pstdev": 26.440853,
        "percentiles": {
          "p50": 72.60159900000001,
          "p75": 83.218431,
          "p80": 85.622783,
          "p90": 90.853375,
          "p95": 94.79782300000001,
          "p99": 111.947775,
          "p999": 184.066047
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 262.859375,
            "min": 244.62890625,
            "mean": 254.54505208333333
          },
          "cpu": {
            "max": 19.933333333333394,
            "min": 0.9998666844423018,
            "mean": 4.028888894814789
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 103.796875,
            "min": 95.66015625,
            "mean": 101.69622395833333
          },
          "cpu": {
            "max": 98.74243346754432,
            "min": 53.54759885646671,
            "mean": 77.38316011334125
          }
        }
      },
      "poolOverflow": 50,
      "upstreamConnections": 350,
      "counters": {
        "benchmark.http_2xx": {
          "value": 424617,
          "perSecond": 4717.966666666666
        },
        "benchmark.pool_overflow": {
          "value": 50,
          "perSecond": 0.5555555555555556
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
          "value": 350,
          "perSecond": 3.888888888888889
        },
        "upstream_cx_rx_bytes_total": {
          "value": 66664869,
          "perSecond": 740720.7666666667
        },
        "upstream_cx_total": {
          "value": 350,
          "perSecond": 3.888888888888889
        },
        "upstream_cx_tx_bytes_total": {
          "value": 19121670,
          "perSecond": 212463
        },
        "upstream_rq_pending_overflow": {
          "value": 50,
          "perSecond": 0.5555555555555556
        },
        "upstream_rq_pending_total": {
          "value": 350,
          "perSecond": 3.888888888888889
        },
        "upstream_rq_total": {
          "value": 424926,
          "perSecond": 4721.4
        }
      }
    },
    {
      "testName": "scaling down httproutes to 500 with 100 routes per hostname at 1000 rps",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-down",
      "throughput": 3996.9333333333334,
      "totalRequests": 359724,
      "latency": {
        "max": 160.21913500000002,
        "min": 0.330384,
        "mean": 1.271993,
        "pstdev": 3.465788,
        "percentiles": {
          "p50": 0.703327,
          "p75": 0.975071,
          "p80": 1.101567,
          "p90": 1.790335,
          "p95": 3.035775,
          "p99": 12.363263,
          "p999": 47.124478999999994
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 243.046875,
            "min": 187.890625,
            "mean": 199.41666666666666
          },
          "cpu": {
            "max": 7.066666666666682,
            "min": 1.0666666666668334,
            "mean": 2.539393939393973
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 105.11328125,
            "min": 101.94921875,
            "mean": 104.24283854166667
          },
          "cpu": {
            "max": 84.14002721088441,
            "min": 55.13337939752684,
            "mean": 71.71722480646932
          }
        }
      },
      "poolOverflow": 273,
      "upstreamConnections": 127,
      "counters": {
        "benchmark.http_2xx": {
          "value": 359724,
          "perSecond": 3996.9333333333334
        },
        "benchmark.pool_overflow": {
          "value": 273,
          "perSecond": 3.033333333333333
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
          "value": 127,
          "perSecond": 1.4111111111111112
        },
        "upstream_cx_rx_bytes_total": {
          "value": 56476668,
          "perSecond": 627518.5333333333
        },
        "upstream_cx_total": {
          "value": 127,
          "perSecond": 1.4111111111111112
        },
        "upstream_cx_tx_bytes_total": {
          "value": 16187715,
          "perSecond": 179863.5
        },
        "upstream_rq_pending_overflow": {
          "value": 273,
          "perSecond": 3.033333333333333
        },
        "upstream_rq_pending_total": {
          "value": 127,
          "perSecond": 1.4111111111111112
        },
        "upstream_rq_total": {
          "value": 359727,
          "perSecond": 3996.9666666666667
        }
      }
    },
    {
      "testName": "scaling down httproutes to 300 with 60 routes per hostname at 800 rps",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-down",
      "throughput": 3196.2555555555555,
      "totalRequests": 287663,
      "latency": {
        "max": 157.073407,
        "min": 0.324416,
        "mean": 0.919661,
        "pstdev": 2.132761,
        "percentiles": {
          "p50": 0.623999,
          "p75": 0.773247,
          "p80": 0.854943,
          "p90": 1.1633909999999998,
          "p95": 1.6647669999999999,
          "p99": 7.266047,
          "p999": 28.302335000000003
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 210.203125,
            "min": 162.78125,
            "mean": 176.39700520833333
          },
          "cpu": {
            "max": 12.06666666666668,
            "min": 1.0666666666668334,
            "mean": 3.4972222222222307
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 104.56640625,
            "min": 103.890625,
            "mean": 104.33020833333333
          },
          "cpu": {
            "max": 71.57665017129798,
            "min": 70.93686382393375,
            "mean": 71.21194568222268
          }
        }
      },
      "poolOverflow": 336,
      "upstreamConnections": 64,
      "counters": {
        "benchmark.http_2xx": {
          "value": 287663,
          "perSecond": 3196.2555555555555
        },
        "benchmark.pool_overflow": {
          "value": 336,
          "perSecond": 3.7333333333333334
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
          "value": 64,
          "perSecond": 0.7111111111111111
        },
        "upstream_cx_rx_bytes_total": {
          "value": 45163091,
          "perSecond": 501812.1222222222
        },
        "upstream_cx_total": {
          "value": 64,
          "perSecond": 0.7111111111111111
        },
        "upstream_cx_tx_bytes_total": {
          "value": 12944880,
          "perSecond": 143832
        },
        "upstream_rq_pending_overflow": {
          "value": 336,
          "perSecond": 3.7333333333333334
        },
        "upstream_rq_pending_total": {
          "value": 64,
          "perSecond": 0.7111111111111111
        },
        "upstream_rq_total": {
          "value": 287664,
          "perSecond": 3196.266666666667
        }
      }
    },
    {
      "testName": "scaling down httproutes to 100 with 20 routes per hostname at 500 rps",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-down",
      "throughput": 2018.9213483146068,
      "totalRequests": 179684,
      "latency": {
        "max": 123.37151899999999,
        "min": 0.33172799999999997,
        "mean": 0.638771,
        "pstdev": 1.373748,
        "percentiles": {
          "p50": 0.5365749999999999,
          "p75": 0.6207349999999999,
          "p80": 0.647231,
          "p90": 0.759135,
          "p95": 0.968383,
          "p99": 1.9668469999999998,
          "p999": 7.4739189999999995
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 174.63671875,
            "min": 155.23828125,
            "mean": 163.27890625
          },
          "cpu": {
            "max": 1.8666666666664846,
            "min": 1.1999999999996664,
            "mean": 1.3969696969697012
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 104.6484375,
            "min": 103.91796875,
            "mean": 104.41940104166666
          },
          "cpu": {
            "max": 0,
            "min": 0,
            "mean": 0
          }
        }
      },
      "poolOverflow": 315,
      "upstreamConnections": 47,
      "counters": {
        "benchmark.http_2xx": {
          "value": 179684,
          "perSecond": 2018.9213483146068
        },
        "benchmark.pool_overflow": {
          "value": 315,
          "perSecond": 3.539325842696629
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
          "value": 47,
          "perSecond": 0.5280898876404494
        },
        "upstream_cx_rx_bytes_total": {
          "value": 28210388,
          "perSecond": 316970.65168539324
        },
        "upstream_cx_total": {
          "value": 47,
          "perSecond": 0.5280898876404494
        },
        "upstream_cx_tx_bytes_total": {
          "value": 8085825,
          "perSecond": 90851.96629213484
        },
        "upstream_rq_pending_overflow": {
          "value": 315,
          "perSecond": 3.539325842696629
        },
        "upstream_rq_pending_total": {
          "value": 47,
          "perSecond": 0.5280898876404494
        },
        "upstream_rq_total": {
          "value": 179685,
          "perSecond": 2018.9325842696628
        }
      }
    },
    {
      "testName": "scaling down httproutes to 50 with 10 routes per hostname at 300 rps",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-down",
      "throughput": 1213.4044943820224,
      "totalRequests": 107993,
      "latency": {
        "max": 13.798399,
        "min": 0.349696,
        "mean": 0.490119,
        "pstdev": 0.242724,
        "percentiles": {
          "p50": 0.454639,
          "p75": 0.472319,
          "p80": 0.479359,
          "p90": 0.514591,
          "p95": 0.586239,
          "p99": 1.411071,
          "p999": 3.425535
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 175.23046875,
            "min": 157.94921875,
            "mean": 160.79557291666666
          },
          "cpu": {
            "max": 1.667000066680003,
            "min": 1.2000000000000455,
            "mean": 1.3361700328912465
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 104.53125,
            "min": 103.78125,
            "mean": 104.05481770833333
          },
          "cpu": {
            "max": 30.264162977018728,
            "min": 19.40162430595029,
            "mean": 25.375699329124853
          }
        }
      },
      "poolOverflow": 7,
      "upstreamConnections": 16,
      "counters": {
        "benchmark.http_2xx": {
          "value": 107993,
          "perSecond": 1213.4044943820224
        },
        "benchmark.pool_overflow": {
          "value": 7,
          "perSecond": 0.07865168539325842
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
          "value": 16,
          "perSecond": 0.1797752808988764
        },
        "upstream_cx_rx_bytes_total": {
          "value": 16954901,
          "perSecond": 190504.50561797753
        },
        "upstream_cx_total": {
          "value": 16,
          "perSecond": 0.1797752808988764
        },
        "upstream_cx_tx_bytes_total": {
          "value": 4859685,
          "perSecond": 54603.20224719101
        },
        "upstream_rq_pending_overflow": {
          "value": 7,
          "perSecond": 0.07865168539325842
        },
        "upstream_rq_pending_total": {
          "value": 16,
          "perSecond": 0.1797752808988764
        },
        "upstream_rq_total": {
          "value": 107993,
          "perSecond": 1213.4044943820224
        }
      }
    },
    {
      "testName": "scaling down httproutes to 10 with 2 routes per hostname at 100 rps",
      "routes": 10,
      "routesPerHostname": 2,
      "phase": "scaling-down",
      "throughput": 404.4719101123595,
      "totalRequests": 35998,
      "latency": {
        "max": 50.925567,
        "min": 0.37795199999999995,
        "mean": 0.525442,
        "pstdev": 0.681925,
        "percentiles": {
          "p50": 0.47334299999999996,
          "p75": 0.48910299999999995,
          "p80": 0.49443100000000006,
          "p90": 0.5247350000000001,
          "p95": 0.592511,
          "p99": 1.716287,
          "p999": 6.911999000000001
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 261.6015625,
            "min": 186.73046875,
            "mean": 232.37005208333332
          },
          "cpu": {
            "max": 1.3333333333332575,
            "min": 1.2000000000000455,
            "mean": 1.2793913547953244
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 104.6171875,
            "min": 103.953125,
            "mean": 104.22174479166667
          },
          "cpu": {
            "max": 0,
            "min": 0,
            "mean": 0
          }
        }
      },
      "poolOverflow": 2,
      "upstreamConnections": 13,
      "counters": {
        "benchmark.http_2xx": {
          "value": 35998,
          "perSecond": 404.4719101123595
        },
        "benchmark.pool_overflow": {
          "value": 2,
          "perSecond": 0.02247191011235955
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
          "value": 13,
          "perSecond": 0.14606741573033707
        },
        "upstream_cx_rx_bytes_total": {
          "value": 5651686,
          "perSecond": 63502.089887640446
        },
        "upstream_cx_total": {
          "value": 13,
          "perSecond": 0.14606741573033707
        },
        "upstream_cx_tx_bytes_total": {
          "value": 1619910,
          "perSecond": 18201.23595505618
        },
        "upstream_rq_pending_overflow": {
          "value": 2,
          "perSecond": 0.02247191011235955
        },
        "upstream_rq_pending_total": {
          "value": 13,
          "perSecond": 0.14606741573033707
        },
        "upstream_rq_total": {
          "value": 35998,
          "perSecond": 404.4719101123595
        }
      }
    }
  ]
};
