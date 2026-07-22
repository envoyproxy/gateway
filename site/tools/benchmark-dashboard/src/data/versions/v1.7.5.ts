import { TestSuite } from '../types';

// Benchmark data extracted from release artifact for version 1.7.5
// Generated from benchmark_result.json

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.7.5",
    "runId": "1.7.5-release-2026-07-08",
    "date": "2026-07-08T15:36:49Z",
    "environment": "GitHub Release",
    "description": "Benchmark results for version 1.7.5 from release artifacts",
    "downloadUrl": "https://github.com/envoyproxy/gateway/releases/download/v1.7.5/benchmark_report.zip",
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
      "throughput": 404.46067415730334,
      "totalRequests": 35997,
      "latency": {
        "max": 40.740863000000004,
        "min": 0.378976,
        "mean": 0.5088929999999999,
        "pstdev": 0.7275590000000001,
        "percentiles": {
          "p50": 0.475471,
          "p75": 0.49051100000000003,
          "p80": 0.49487900000000007,
          "p90": 0.510031,
          "p95": 0.5354230000000001,
          "p99": 1.0564470000000001,
          "p999": 4.872703
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 130.08984375,
            "min": 125.99609375,
            "mean": 128.72135416666666
          },
          "cpu": {
            "max": 1.0666666666666662,
            "min": 0.13333333333333347,
            "mean": 0.3555555555555555
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 23.609375,
            "min": 23.11328125,
            "mean": 23.408333333333335
          },
          "cpu": {
            "max": 10.531097611033577,
            "min": 6.74744975768273,
            "mean": 10.076916047509108
          }
        }
      },
      "poolOverflow": 3,
      "upstreamConnections": 13,
      "counters": {
        "benchmark.http_2xx": {
          "value": 35997,
          "perSecond": 404.46067415730334
        },
        "benchmark.pool_overflow": {
          "value": 3,
          "perSecond": 0.033707865168539325
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
          "value": 5651529,
          "perSecond": 63500.32584269663
        },
        "upstream_cx_total": {
          "value": 13,
          "perSecond": 0.14606741573033707
        },
        "upstream_cx_tx_bytes_total": {
          "value": 1619865,
          "perSecond": 18200.73033707865
        },
        "upstream_rq_pending_overflow": {
          "value": 3,
          "perSecond": 0.033707865168539325
        },
        "upstream_rq_pending_total": {
          "value": 13,
          "perSecond": 0.14606741573033707
        },
        "upstream_rq_total": {
          "value": 35997,
          "perSecond": 404.46067415730334
        }
      }
    },
    {
      "testName": "scaling up httproutes to 50 with 10 routes per hostname at 300 rps",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-up",
      "throughput": 1213.0224719101125,
      "totalRequests": 107959,
      "latency": {
        "max": 63.63545499999999,
        "min": 0.35260800000000003,
        "mean": 0.487317,
        "pstdev": 0.512447,
        "percentiles": {
          "p50": 0.450335,
          "p75": 0.470415,
          "p80": 0.47551899999999997,
          "p90": 0.494223,
          "p95": 0.538207,
          "p99": 1.389247,
          "p999": 5.619967
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 137.29296875,
            "min": 135.15234375,
            "mean": 135.28619791666668
          },
          "cpu": {
            "max": 2.0666666666666655,
            "min": 0.13333333333333347,
            "mean": 0.513333333333333
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 29.578125,
            "min": 27.1796875,
            "mean": 29.144401041666665
          },
          "cpu": {
            "max": 30.144766966475878,
            "min": 17.716992856688087,
            "mean": 26.46859586575745
          }
        }
      },
      "poolOverflow": 41,
      "upstreamConnections": 32,
      "counters": {
        "benchmark.http_2xx": {
          "value": 107959,
          "perSecond": 1213.0224719101125
        },
        "benchmark.pool_overflow": {
          "value": 41,
          "perSecond": 0.4606741573033708
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
          "value": 32,
          "perSecond": 0.3595505617977528
        },
        "upstream_cx_rx_bytes_total": {
          "value": 16949563,
          "perSecond": 190444.52808988764
        },
        "upstream_cx_total": {
          "value": 32,
          "perSecond": 0.3595505617977528
        },
        "upstream_cx_tx_bytes_total": {
          "value": 4858155,
          "perSecond": 54586.011235955055
        },
        "upstream_rq_pending_overflow": {
          "value": 41,
          "perSecond": 0.4606741573033708
        },
        "upstream_rq_pending_total": {
          "value": 32,
          "perSecond": 0.3595505617977528
        },
        "upstream_rq_total": {
          "value": 107959,
          "perSecond": 1213.0224719101125
        }
      }
    },
    {
      "testName": "scaling up httproutes to 100 with 20 routes per hostname at 500 rps",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-up",
      "throughput": 2020.2696629213483,
      "totalRequests": 179804,
      "latency": {
        "max": 85.991423,
        "min": 0.290608,
        "mean": 0.49629199999999996,
        "pstdev": 0.7787850000000001,
        "percentiles": {
          "p50": 0.43433499999999997,
          "p75": 0.457519,
          "p80": 0.464767,
          "p90": 0.495855,
          "p95": 0.6559670000000001,
          "p99": 1.870719,
          "p999": 7.932671000000001
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 138.1015625,
            "min": 138.1015625,
            "mean": 138.1015625
          },
          "cpu": {
            "max": 2.533333333333333,
            "min": 0.13333333333333347,
            "mean": 0.6111111111111112
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 36.19140625,
            "min": 33.421875,
            "mean": 35.618489583333336
          },
          "cpu": {
            "max": 46.994710251098866,
            "min": 46.994710251098866,
            "mean": 46.994710251098866
          }
        }
      },
      "poolOverflow": 196,
      "upstreamConnections": 54,
      "counters": {
        "benchmark.http_2xx": {
          "value": 179804,
          "perSecond": 2020.2696629213483
        },
        "benchmark.pool_overflow": {
          "value": 196,
          "perSecond": 2.202247191011236
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
          "value": 54,
          "perSecond": 0.6067415730337079
        },
        "upstream_cx_rx_bytes_total": {
          "value": 28229228,
          "perSecond": 317182.3370786517
        },
        "upstream_cx_total": {
          "value": 54,
          "perSecond": 0.6067415730337079
        },
        "upstream_cx_tx_bytes_total": {
          "value": 8091180,
          "perSecond": 90912.13483146067
        },
        "upstream_rq_pending_overflow": {
          "value": 196,
          "perSecond": 2.202247191011236
        },
        "upstream_rq_pending_total": {
          "value": 54,
          "perSecond": 0.6067415730337079
        },
        "upstream_rq_total": {
          "value": 179804,
          "perSecond": 2020.2696629213483
        }
      }
    },
    {
      "testName": "scaling up httproutes to 300 with 60 routes per hostname at 800 rps",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-up",
      "throughput": 3199.4555555555557,
      "totalRequests": 287951,
      "latency": {
        "max": 138.452991,
        "min": 0.264784,
        "mean": 1.207951,
        "pstdev": 5.162012,
        "percentiles": {
          "p50": 0.5925750000000001,
          "p75": 0.6785909999999999,
          "p80": 0.7471030000000001,
          "p90": 1.249087,
          "p95": 2.074879,
          "p99": 12.977151,
          "p999": 89.747455
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 144.7265625,
            "min": 142.16015625,
            "mean": 143.97526041666666
          },
          "cpu": {
            "max": 5.066666666666677,
            "min": 0.1333333333333305,
            "mean": 0.9888888888888895
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 68,
            "min": 58.19140625,
            "mean": 64.08984375
          },
          "cpu": {
            "max": 68.22084858569042,
            "min": 41.70343299524441,
            "mean": 61.86047234637115
          }
        }
      },
      "poolOverflow": 49,
      "upstreamConnections": 326,
      "counters": {
        "benchmark.http_2xx": {
          "value": 287951,
          "perSecond": 3199.4555555555557
        },
        "benchmark.pool_overflow": {
          "value": 49,
          "perSecond": 0.5444444444444444
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
          "value": 326,
          "perSecond": 3.6222222222222222
        },
        "upstream_cx_rx_bytes_total": {
          "value": 45208307,
          "perSecond": 502314.52222222224
        },
        "upstream_cx_total": {
          "value": 326,
          "perSecond": 3.6222222222222222
        },
        "upstream_cx_tx_bytes_total": {
          "value": 12957795,
          "perSecond": 143975.5
        },
        "upstream_rq_pending_overflow": {
          "value": 49,
          "perSecond": 0.5444444444444444
        },
        "upstream_rq_pending_total": {
          "value": 326,
          "perSecond": 3.6222222222222222
        },
        "upstream_rq_total": {
          "value": 287951,
          "perSecond": 3199.4555555555557
        }
      }
    },
    {
      "testName": "scaling up httproutes to 500 with 100 routes per hostname at 1000 rps",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-up",
      "throughput": 3999.177777777778,
      "totalRequests": 359926,
      "latency": {
        "max": 443.858943,
        "min": 0.264192,
        "mean": 3.7524490000000004,
        "pstdev": 15.938013000000002,
        "percentiles": {
          "p50": 0.711615,
          "p75": 1.2355829999999999,
          "p80": 1.5825909999999999,
          "p90": 3.236863,
          "p95": 7.336958999999999,
          "p99": 85.340159,
          "p999": 157.859839
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 161.8671875,
            "min": 152.77734375,
            "mean": 160.35221354166666
          },
          "cpu": {
            "max": 7.000000000000004,
            "min": 0.19999999999998386,
            "mean": 1.3555555555555538
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 92.54296875,
            "min": 84.4609375,
            "mean": 89.76236979166667
          },
          "cpu": {
            "max": 80.72387478849407,
            "min": 44.87404051042136,
            "mean": 62.13861607744482
          }
        }
      },
      "poolOverflow": 71,
      "upstreamConnections": 329,
      "counters": {
        "benchmark.http_2xx": {
          "value": 359926,
          "perSecond": 3999.177777777778
        },
        "benchmark.pool_overflow": {
          "value": 71,
          "perSecond": 0.7888888888888889
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
          "value": 329,
          "perSecond": 3.6555555555555554
        },
        "upstream_cx_rx_bytes_total": {
          "value": 56508382,
          "perSecond": 627870.911111111
        },
        "upstream_cx_total": {
          "value": 329,
          "perSecond": 3.6555555555555554
        },
        "upstream_cx_tx_bytes_total": {
          "value": 16196805,
          "perSecond": 179964.5
        },
        "upstream_rq_pending_overflow": {
          "value": 71,
          "perSecond": 0.7888888888888889
        },
        "upstream_rq_pending_total": {
          "value": 329,
          "perSecond": 3.6555555555555554
        },
        "upstream_rq_total": {
          "value": 359929,
          "perSecond": 3999.211111111111
        }
      }
    },
    {
      "testName": "scaling up httproutes to 1000 with 200 routes per hostname at 2000 rps",
      "routes": 1000,
      "routesPerHostname": 200,
      "phase": "scaling-up",
      "throughput": 5385.4,
      "totalRequests": 484686,
      "latency": {
        "max": 1876.885503,
        "min": 1.8483200000000002,
        "mean": 58.245123,
        "pstdev": 26.072029,
        "percentiles": {
          "p50": 56.051711,
          "p75": 65.304575,
          "p80": 67.309567,
          "p90": 71.97081499999999,
          "p95": 77.377535,
          "p99": 109.314047,
          "p999": 174.579711
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 183.9609375,
            "min": 173.19921875,
            "mean": 174.90078125
          },
          "cpu": {
            "max": 13.666666666666744,
            "min": 0.19999999999991283,
            "mean": 2.4843792691346436
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 143.34765625,
            "min": 135.33984375,
            "mean": 140.876171875
          },
          "cpu": {
            "max": 99.63894164326153,
            "min": 98.67385725741774,
            "mean": 99.26863566184787
          }
        }
      },
      "poolOverflow": 84,
      "upstreamConnections": 316,
      "counters": {
        "benchmark.http_2xx": {
          "value": 484686,
          "perSecond": 5385.4
        },
        "benchmark.pool_overflow": {
          "value": 84,
          "perSecond": 0.9333333333333333
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
          "value": 316,
          "perSecond": 3.511111111111111
        },
        "upstream_cx_rx_bytes_total": {
          "value": 76095702,
          "perSecond": 845507.8
        },
        "upstream_cx_total": {
          "value": 316,
          "perSecond": 3.511111111111111
        },
        "upstream_cx_tx_bytes_total": {
          "value": 21825090,
          "perSecond": 242501
        },
        "upstream_rq_pending_overflow": {
          "value": 84,
          "perSecond": 0.9333333333333333
        },
        "upstream_rq_pending_total": {
          "value": 316,
          "perSecond": 3.511111111111111
        },
        "upstream_rq_total": {
          "value": 485002,
          "perSecond": 5388.9111111111115
        }
      }
    },
    {
      "testName": "scaling down httproutes to 500 with 100 routes per hostname at 1000 rps",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-down",
      "throughput": 3999.0222222222224,
      "totalRequests": 359912,
      "latency": {
        "max": 332.857343,
        "min": 0.26344,
        "mean": 2.509089,
        "pstdev": 9.775503,
        "percentiles": {
          "p50": 0.6791670000000001,
          "p75": 1.0431030000000001,
          "p80": 1.311679,
          "p90": 2.827135,
          "p95": 5.955071,
          "p99": 55.834623,
          "p999": 98.39411100000001
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 190.32421875,
            "min": 158.69921875,
            "mean": 164.38671875
          },
          "cpu": {
            "max": 0.2666666666667084,
            "min": 0.1333333333333068,
            "mean": 0.20000000000000753
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 148.25390625,
            "min": 141.0546875,
            "mean": 143.53841145833334
          },
          "cpu": {
            "max": 80.86865352647818,
            "min": 44.001111287746824,
            "mean": 74.12282990985476
          }
        }
      },
      "poolOverflow": 86,
      "upstreamConnections": 314,
      "counters": {
        "benchmark.http_2xx": {
          "value": 359912,
          "perSecond": 3999.0222222222224
        },
        "benchmark.pool_overflow": {
          "value": 86,
          "perSecond": 0.9555555555555556
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
          "value": 314,
          "perSecond": 3.488888888888889
        },
        "upstream_cx_rx_bytes_total": {
          "value": 56506184,
          "perSecond": 627846.4888888889
        },
        "upstream_cx_total": {
          "value": 314,
          "perSecond": 3.488888888888889
        },
        "upstream_cx_tx_bytes_total": {
          "value": 16196130,
          "perSecond": 179957
        },
        "upstream_rq_pending_overflow": {
          "value": 86,
          "perSecond": 0.9555555555555556
        },
        "upstream_rq_pending_total": {
          "value": 314,
          "perSecond": 3.488888888888889
        },
        "upstream_rq_total": {
          "value": 359914,
          "perSecond": 3999.0444444444443
        }
      }
    },
    {
      "testName": "scaling down httproutes to 300 with 60 routes per hostname at 800 rps",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-down",
      "throughput": 3198.988888888889,
      "totalRequests": 287909,
      "latency": {
        "max": 116.113407,
        "min": 0.277152,
        "mean": 1.150514,
        "pstdev": 3.81955,
        "percentiles": {
          "p50": 0.591263,
          "p75": 0.679039,
          "p80": 0.7616949999999999,
          "p90": 1.3676789999999999,
          "p95": 2.382847,
          "p99": 14.532606999999999,
          "p999": 55.408639
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 159.09375,
            "min": 157.96875,
            "mean": 158.69596354166666
          },
          "cpu": {
            "max": 0.2668979782478565,
            "min": 0.1333333333333068,
            "mean": 0.1939919647532957
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 142.52734375,
            "min": 141.1953125,
            "mean": 141.98515625
          },
          "cpu": {
            "max": 68.17021856339512,
            "min": 46.6719064597314,
            "mean": 66.0963438187755
          }
        }
      },
      "poolOverflow": 90,
      "upstreamConnections": 221,
      "counters": {
        "benchmark.http_2xx": {
          "value": 287909,
          "perSecond": 3198.988888888889
        },
        "benchmark.pool_overflow": {
          "value": 90,
          "perSecond": 1
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
          "value": 221,
          "perSecond": 2.4555555555555557
        },
        "upstream_cx_rx_bytes_total": {
          "value": 45201713,
          "perSecond": 502241.2555555556
        },
        "upstream_cx_total": {
          "value": 221,
          "perSecond": 2.4555555555555557
        },
        "upstream_cx_tx_bytes_total": {
          "value": 12955950,
          "perSecond": 143955
        },
        "upstream_rq_pending_overflow": {
          "value": 90,
          "perSecond": 1
        },
        "upstream_rq_pending_total": {
          "value": 221,
          "perSecond": 2.4555555555555557
        },
        "upstream_rq_total": {
          "value": 287910,
          "perSecond": 3199
        }
      }
    },
    {
      "testName": "scaling down httproutes to 100 with 20 routes per hostname at 500 rps",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-down",
      "throughput": 2020.6966292134832,
      "totalRequests": 179842,
      "latency": {
        "max": 83.697663,
        "min": 0.286272,
        "mean": 0.5179819999999999,
        "pstdev": 0.9698060000000001,
        "percentiles": {
          "p50": 0.442751,
          "p75": 0.464303,
          "p80": 0.470031,
          "p90": 0.495839,
          "p95": 0.679071,
          "p99": 2.114815,
          "p999": 10.049535
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 168.8359375,
            "min": 149.0625,
            "mean": 153.08723958333334
          },
          "cpu": {
            "max": 4.066666666666663,
            "min": 0.13333333333340155,
            "mean": 0.9599999999999984
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 141.69921875,
            "min": 141.21875,
            "mean": 141.47200520833334
          },
          "cpu": {
            "max": 48.569813084112255,
            "min": 48.077679610985385,
            "mean": 48.32794707266626
          }
        }
      },
      "poolOverflow": 158,
      "upstreamConnections": 63,
      "counters": {
        "benchmark.http_2xx": {
          "value": 179842,
          "perSecond": 2020.6966292134832
        },
        "benchmark.pool_overflow": {
          "value": 158,
          "perSecond": 1.7752808988764044
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
          "value": 63,
          "perSecond": 0.7078651685393258
        },
        "upstream_cx_rx_bytes_total": {
          "value": 28235194,
          "perSecond": 317249.37078651687
        },
        "upstream_cx_total": {
          "value": 63,
          "perSecond": 0.7078651685393258
        },
        "upstream_cx_tx_bytes_total": {
          "value": 8092890,
          "perSecond": 90931.34831460674
        },
        "upstream_rq_pending_overflow": {
          "value": 158,
          "perSecond": 1.7752808988764044
        },
        "upstream_rq_pending_total": {
          "value": 63,
          "perSecond": 0.7078651685393258
        },
        "upstream_rq_total": {
          "value": 179842,
          "perSecond": 2020.6966292134832
        }
      }
    },
    {
      "testName": "scaling down httproutes to 50 with 10 routes per hostname at 300 rps",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-down",
      "throughput": 1212.5280898876404,
      "totalRequests": 107915,
      "latency": {
        "max": 66.39001499999999,
        "min": 0.34712000000000004,
        "mean": 0.501021,
        "pstdev": 0.9397420000000001,
        "percentiles": {
          "p50": 0.45027100000000003,
          "p75": 0.468511,
          "p80": 0.473327,
          "p90": 0.490223,
          "p95": 0.5424950000000001,
          "p99": 1.3708150000000001,
          "p999": 9.019903
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 150.3125,
            "min": 138.4140625,
            "mean": 140.31223958333334
          },
          "cpu": {
            "max": 0.2000000000000076,
            "min": 0.1333333333333068,
            "mean": 0.17878787878789712
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 141.70703125,
            "min": 140.8828125,
            "mean": 141.17109375
          },
          "cpu": {
            "max": 30.12063364308174,
            "min": 19.692582564011218,
            "mean": 28.7662632231051
          }
        }
      },
      "poolOverflow": 85,
      "upstreamConnections": 38,
      "counters": {
        "benchmark.http_2xx": {
          "value": 107915,
          "perSecond": 1212.5280898876404
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
          "value": 38,
          "perSecond": 0.42696629213483145
        },
        "upstream_cx_rx_bytes_total": {
          "value": 16942655,
          "perSecond": 190366.91011235956
        },
        "upstream_cx_total": {
          "value": 38,
          "perSecond": 0.42696629213483145
        },
        "upstream_cx_tx_bytes_total": {
          "value": 4856175,
          "perSecond": 54563.76404494382
        },
        "upstream_rq_pending_overflow": {
          "value": 85,
          "perSecond": 0.9550561797752809
        },
        "upstream_rq_pending_total": {
          "value": 38,
          "perSecond": 0.42696629213483145
        },
        "upstream_rq_total": {
          "value": 107915,
          "perSecond": 1212.5280898876404
        }
      }
    },
    {
      "testName": "scaling down httproutes to 10 with 2 routes per hostname at 100 rps",
      "routes": 10,
      "routesPerHostname": 2,
      "phase": "scaling-down",
      "throughput": 399.96666666666664,
      "totalRequests": 35997,
      "latency": {
        "max": 57.362431,
        "min": 0.38779199999999997,
        "mean": 0.49359100000000006,
        "pstdev": 0.429145,
        "percentiles": {
          "p50": 0.468223,
          "p75": 0.486559,
          "p80": 0.491615,
          "p90": 0.507215,
          "p95": 0.528639,
          "p99": 1.005791,
          "p999": 4.174847
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 145.31640625,
            "min": 139.6640625,
            "mean": 144.648828125
          },
          "cpu": {
            "max": 0.20000000000000764,
            "min": 0.1333333333333068,
            "mean": 0.17222222222221562
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 140.9921875,
            "min": 140.70703125,
            "mean": 140.89322916666666
          },
          "cpu": {
            "max": 10.635300213744282,
            "min": 7.031060889842074,
            "mean": 9.981794070354459
          }
        }
      },
      "poolOverflow": 3,
      "upstreamConnections": 11,
      "counters": {
        "benchmark.http_2xx": {
          "value": 35997,
          "perSecond": 399.96666666666664
        },
        "benchmark.pool_overflow": {
          "value": 3,
          "perSecond": 0.03333333333333333
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
          "value": 5651529,
          "perSecond": 62794.76666666667
        },
        "upstream_cx_total": {
          "value": 11,
          "perSecond": 0.12222222222222222
        },
        "upstream_cx_tx_bytes_total": {
          "value": 1619865,
          "perSecond": 17998.5
        },
        "upstream_rq_pending_overflow": {
          "value": 3,
          "perSecond": 0.03333333333333333
        },
        "upstream_rq_pending_total": {
          "value": 11,
          "perSecond": 0.12222222222222222
        },
        "upstream_rq_total": {
          "value": 35997,
          "perSecond": 399.96666666666664
        }
      }
    }
  ]
};
