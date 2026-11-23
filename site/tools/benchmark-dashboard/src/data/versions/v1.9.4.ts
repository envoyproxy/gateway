import { TestSuite } from '../types';

// Benchmark data extracted from release artifact for version 1.9.4
// Generated from benchmark_result.json

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.9.4",
    "runId": "1.9.4-release-2026-11-17",
    "date": "2026-11-17T11:15:53Z",
    "environment": "GitHub Release",
    "description": "Benchmark results for version 1.9.4 from release artifacts",
    "downloadUrl": "https://github.com/envoyproxy/gateway/releases/download/v1.9.4/benchmark_report.zip",
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
      "throughput": 399.97777777777776,
      "totalRequests": 35998,
      "latency": {
        "max": 14.721535,
        "min": 0.14976799999999998,
        "mean": 0.255804,
        "pstdev": 0.20713299999999998,
        "percentiles": {
          "p50": 0.238407,
          "p75": 0.255327,
          "p80": 0.26071900000000003,
          "p90": 0.283231,
          "p95": 0.318623,
          "p99": 0.8322229999999999,
          "p999": 2.335615
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 132.26171875,
            "min": 114.40234375,
            "mean": 127.215234375
          },
          "cpu": {
            "max": 0.6666666666666665,
            "min": 0.13333333333333347,
            "mean": 0.345679012345679
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 18.01953125,
            "min": 4.57421875,
            "mean": 14.285286458333333
          },
          "cpu": {
            "max": 5.288261343012706,
            "min": 5.154227341882197,
            "mean": 5.207840942334402
          }
        }
      },
      "poolOverflow": 2,
      "upstreamConnections": 7,
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
          "value": 7,
          "perSecond": 0.07777777777777778
        },
        "upstream_cx_rx_bytes_total": {
          "value": 5651686,
          "perSecond": 62796.51111111111
        },
        "upstream_cx_total": {
          "value": 7,
          "perSecond": 0.07777777777777778
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
          "value": 7,
          "perSecond": 0.07777777777777778
        },
        "upstream_rq_total": {
          "value": 35998,
          "perSecond": 399.97777777777776
        }
      }
    },
    {
      "testName": "scaling up httproutes to 50 with 10 routes per hostname at 300 rps",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-up",
      "throughput": 1199.8,
      "totalRequests": 107982,
      "latency": {
        "max": 32.942079,
        "min": 0.121764,
        "mean": 0.246454,
        "pstdev": 0.326897,
        "percentiles": {
          "p50": 0.229287,
          "p75": 0.249087,
          "p80": 0.254479,
          "p90": 0.27358299999999997,
          "p95": 0.310175,
          "p99": 0.8710070000000001,
          "p999": 3.0053110000000003
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 137.3203125,
            "min": 130.89453125,
            "mean": 134.68684895833334
          },
          "cpu": {
            "max": 2.4,
            "min": 0.3333333333333336,
            "mean": 0.786666666666666
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 21.19921875,
            "min": 17.83203125,
            "mean": 20.256640625
          },
          "cpu": {
            "max": 15.399397665955577,
            "min": 10.120688801111301,
            "mean": 14.679929396181027
          }
        }
      },
      "poolOverflow": 18,
      "upstreamConnections": 23,
      "counters": {
        "benchmark.http_2xx": {
          "value": 107982,
          "perSecond": 1199.8
        },
        "benchmark.pool_overflow": {
          "value": 18,
          "perSecond": 0.2
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
          "value": 23,
          "perSecond": 0.25555555555555554
        },
        "upstream_cx_rx_bytes_total": {
          "value": 16953174,
          "perSecond": 188368.6
        },
        "upstream_cx_total": {
          "value": 23,
          "perSecond": 0.25555555555555554
        },
        "upstream_cx_tx_bytes_total": {
          "value": 4859190,
          "perSecond": 53991
        },
        "upstream_rq_pending_overflow": {
          "value": 18,
          "perSecond": 0.2
        },
        "upstream_rq_pending_total": {
          "value": 23,
          "perSecond": 0.25555555555555554
        },
        "upstream_rq_total": {
          "value": 107982,
          "perSecond": 1199.8
        }
      }
    },
    {
      "testName": "scaling up httproutes to 100 with 20 routes per hostname at 500 rps",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-up",
      "throughput": 2022.1685393258426,
      "totalRequests": 179973,
      "latency": {
        "max": 22.725631,
        "min": 0.111156,
        "mean": 0.214912,
        "pstdev": 0.300435,
        "percentiles": {
          "p50": 0.186663,
          "p75": 0.217759,
          "p80": 0.223487,
          "p90": 0.239703,
          "p95": 0.27980700000000003,
          "p99": 1.002207,
          "p999": 3.248767
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 143.96875,
            "min": 132.31640625,
            "mean": 139.76184895833333
          },
          "cpu": {
            "max": 4.133333333333334,
            "min": 0.3999999999999974,
            "mean": 1.1160493827160498
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 25.44140625,
            "min": 20.66796875,
            "mean": 24.165885416666665
          },
          "cpu": {
            "max": 20.066395292600372,
            "min": 11.436525166511617,
            "mean": 16.46306469213024
          }
        }
      },
      "poolOverflow": 27,
      "upstreamConnections": 32,
      "counters": {
        "benchmark.http_2xx": {
          "value": 179973,
          "perSecond": 2022.1685393258426
        },
        "benchmark.pool_overflow": {
          "value": 27,
          "perSecond": 0.30337078651685395
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
          "value": 28255761,
          "perSecond": 317480.4606741573
        },
        "upstream_cx_total": {
          "value": 32,
          "perSecond": 0.3595505617977528
        },
        "upstream_cx_tx_bytes_total": {
          "value": 8098785,
          "perSecond": 90997.58426966293
        },
        "upstream_rq_pending_overflow": {
          "value": 27,
          "perSecond": 0.30337078651685395
        },
        "upstream_rq_pending_total": {
          "value": 32,
          "perSecond": 0.3595505617977528
        },
        "upstream_rq_total": {
          "value": 179973,
          "perSecond": 2022.1685393258426
        }
      }
    },
    {
      "testName": "scaling up httproutes to 300 with 60 routes per hostname at 800 rps",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-up",
      "throughput": 3197.6666666666665,
      "totalRequests": 287790,
      "latency": {
        "max": 66.44735899999999,
        "min": 0.103464,
        "mean": 0.229231,
        "pstdev": 0.512822,
        "percentiles": {
          "p50": 0.181847,
          "p75": 0.220663,
          "p80": 0.228215,
          "p90": 0.249807,
          "p95": 0.321839,
          "p99": 1.271039,
          "p999": 4.321535
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 163.94921875,
            "min": 146.1171875,
            "mean": 154.69140625
          },
          "cpu": {
            "max": 10.066666666666652,
            "min": 0.4666666666666685,
            "mean": 2.2511111111111126
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 39.421875,
            "min": 37.265625,
            "mean": 38.98515625
          },
          "cpu": {
            "max": 32.53456712232737,
            "min": 19.118885485120767,
            "mean": 30.757506366729693
          }
        }
      },
      "poolOverflow": 210,
      "upstreamConnections": 50,
      "counters": {
        "benchmark.http_2xx": {
          "value": 287790,
          "perSecond": 3197.6666666666665
        },
        "benchmark.pool_overflow": {
          "value": 210,
          "perSecond": 2.3333333333333335
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
          "value": 45183030,
          "perSecond": 502033.6666666667
        },
        "upstream_cx_total": {
          "value": 50,
          "perSecond": 0.5555555555555556
        },
        "upstream_cx_tx_bytes_total": {
          "value": 12950550,
          "perSecond": 143895
        },
        "upstream_rq_pending_overflow": {
          "value": 210,
          "perSecond": 2.3333333333333335
        },
        "upstream_rq_pending_total": {
          "value": 50,
          "perSecond": 0.5555555555555556
        },
        "upstream_rq_total": {
          "value": 287790,
          "perSecond": 3197.6666666666665
        }
      }
    },
    {
      "testName": "scaling up httproutes to 500 with 100 routes per hostname at 1000 rps",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-up",
      "throughput": 4043.0224719101125,
      "totalRequests": 359829,
      "latency": {
        "max": 39.022591000000006,
        "min": 0.10126,
        "mean": 0.212948,
        "pstdev": 0.37268,
        "percentiles": {
          "p50": 0.16443100000000002,
          "p75": 0.207279,
          "p80": 0.215079,
          "p90": 0.23458299999999999,
          "p95": 0.306575,
          "p99": 1.1616630000000001,
          "p999": 4.175103
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 185.484375,
            "min": 173.234375,
            "mean": 179.35859375
          },
          "cpu": {
            "max": 10.266666666666662,
            "min": 0.533333333333322,
            "mean": 2.3600000000000048
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 53.84765625,
            "min": 50.65234375,
            "mean": 52.756770833333334
          },
          "cpu": {
            "max": 38.37441511893124,
            "min": 20.51171162517767,
            "mean": 33.13848302315136
          }
        }
      },
      "poolOverflow": 171,
      "upstreamConnections": 59,
      "counters": {
        "benchmark.http_2xx": {
          "value": 359829,
          "perSecond": 4043.0224719101125
        },
        "benchmark.pool_overflow": {
          "value": 171,
          "perSecond": 1.9213483146067416
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
          "value": 59,
          "perSecond": 0.6629213483146067
        },
        "upstream_cx_rx_bytes_total": {
          "value": 56493153,
          "perSecond": 634754.5280898877
        },
        "upstream_cx_total": {
          "value": 59,
          "perSecond": 0.6629213483146067
        },
        "upstream_cx_tx_bytes_total": {
          "value": 16192305,
          "perSecond": 181936.01123595505
        },
        "upstream_rq_pending_overflow": {
          "value": 171,
          "perSecond": 1.9213483146067416
        },
        "upstream_rq_pending_total": {
          "value": 59,
          "perSecond": 0.6629213483146067
        },
        "upstream_rq_total": {
          "value": 359829,
          "perSecond": 4043.0224719101125
        }
      }
    },
    {
      "testName": "scaling up httproutes to 1000 with 200 routes per hostname at 2000 rps",
      "routes": 1000,
      "routesPerHostname": 200,
      "phase": "scaling-up",
      "throughput": 7998.055555555556,
      "totalRequests": 719825,
      "latency": {
        "max": 77.602815,
        "min": 0.09966,
        "mean": 0.585122,
        "pstdev": 2.635966,
        "percentiles": {
          "p50": 0.194631,
          "p75": 0.290863,
          "p80": 0.340463,
          "p90": 0.629503,
          "p95": 1.482367,
          "p99": 9.083390999999999,
          "p999": 46.157823
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 262.97265625,
            "min": 241.4609375,
            "mean": 251.319140625
          },
          "cpu": {
            "max": 11.066666666666643,
            "min": 0.6666666666666288,
            "mean": 2.7488888888889007
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 95.01171875,
            "min": 83.63671875,
            "mean": 91.3734375
          },
          "cpu": {
            "max": 69.03577296094127,
            "min": 40.48571111036484,
            "mean": 63.65463177478578
          }
        }
      },
      "poolOverflow": 175,
      "upstreamConnections": 225,
      "counters": {
        "benchmark.http_2xx": {
          "value": 719825,
          "perSecond": 7998.055555555556
        },
        "benchmark.pool_overflow": {
          "value": 175,
          "perSecond": 1.9444444444444444
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
          "value": 225,
          "perSecond": 2.5
        },
        "upstream_cx_rx_bytes_total": {
          "value": 113012525,
          "perSecond": 1255694.7222222222
        },
        "upstream_cx_total": {
          "value": 225,
          "perSecond": 2.5
        },
        "upstream_cx_tx_bytes_total": {
          "value": 32392125,
          "perSecond": 359912.5
        },
        "upstream_rq_pending_overflow": {
          "value": 175,
          "perSecond": 1.9444444444444444
        },
        "upstream_rq_pending_total": {
          "value": 225,
          "perSecond": 2.5
        },
        "upstream_rq_total": {
          "value": 719825,
          "perSecond": 7998.055555555556
        }
      }
    },
    {
      "testName": "scaling down httproutes to 500 with 100 routes per hostname at 1000 rps",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-down",
      "throughput": 3996.4777777777776,
      "totalRequests": 359683,
      "latency": {
        "max": 93.782015,
        "min": 0.10714,
        "mean": 0.36561699999999997,
        "pstdev": 1.708116,
        "percentiles": {
          "p50": 0.190991,
          "p75": 0.233135,
          "p80": 0.24544699999999997,
          "p90": 0.385855,
          "p95": 0.791199,
          "p99": 3.655551,
          "p999": 22.879231
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 263.59765625,
            "min": 188.66796875,
            "mean": 204.13255208333334
          },
          "cpu": {
            "max": 5.733333333333235,
            "min": 0.6666666666666288,
            "mean": 1.9936507936507755
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 95.01171875,
            "min": 94.71484375,
            "mean": 94.890234375
          },
          "cpu": {
            "max": 45.240416380976036,
            "min": 27.770848387666526,
            "mean": 39.213026927341346
          }
        }
      },
      "poolOverflow": 317,
      "upstreamConnections": 83,
      "counters": {
        "benchmark.http_2xx": {
          "value": 359683,
          "perSecond": 3996.4777777777776
        },
        "benchmark.pool_overflow": {
          "value": 317,
          "perSecond": 3.522222222222222
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
          "value": 83,
          "perSecond": 0.9222222222222223
        },
        "upstream_cx_rx_bytes_total": {
          "value": 56470231,
          "perSecond": 627447.0111111111
        },
        "upstream_cx_total": {
          "value": 83,
          "perSecond": 0.9222222222222223
        },
        "upstream_cx_tx_bytes_total": {
          "value": 16185735,
          "perSecond": 179841.5
        },
        "upstream_rq_pending_overflow": {
          "value": 317,
          "perSecond": 3.522222222222222
        },
        "upstream_rq_pending_total": {
          "value": 83,
          "perSecond": 0.9222222222222223
        },
        "upstream_rq_total": {
          "value": 359683,
          "perSecond": 3996.4777777777776
        }
      }
    },
    {
      "testName": "scaling down httproutes to 300 with 60 routes per hostname at 800 rps",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-down",
      "throughput": 3232.5842696629215,
      "totalRequests": 287700,
      "latency": {
        "max": 80.277503,
        "min": 0.102532,
        "mean": 0.277121,
        "pstdev": 1.0003090000000001,
        "percentiles": {
          "p50": 0.179791,
          "p75": 0.232415,
          "p80": 0.244863,
          "p90": 0.318895,
          "p95": 0.463487,
          "p99": 1.9096950000000001,
          "p999": 10.950655
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 188.66796875,
            "min": 158.9609375,
            "mean": 166.45234375
          },
          "cpu": {
            "max": 8.199999999999932,
            "min": 0.6666666666666288,
            "mean": 2.327777777777761
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 95.11328125,
            "min": 94.1328125,
            "mean": 94.8015625
          },
          "cpu": {
            "max": 30.19899725917513,
            "min": 29.49879968103417,
            "mean": 30.058957743546937
          }
        }
      },
      "poolOverflow": 299,
      "upstreamConnections": 72,
      "counters": {
        "benchmark.http_2xx": {
          "value": 287700,
          "perSecond": 3232.5842696629215
        },
        "benchmark.pool_overflow": {
          "value": 299,
          "perSecond": 3.359550561797753
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
          "value": 72,
          "perSecond": 0.8089887640449438
        },
        "upstream_cx_rx_bytes_total": {
          "value": 45168900,
          "perSecond": 507515.7303370786
        },
        "upstream_cx_total": {
          "value": 72,
          "perSecond": 0.8089887640449438
        },
        "upstream_cx_tx_bytes_total": {
          "value": 12946500,
          "perSecond": 145466.29213483146
        },
        "upstream_rq_pending_overflow": {
          "value": 299,
          "perSecond": 3.359550561797753
        },
        "upstream_rq_pending_total": {
          "value": 72,
          "perSecond": 0.8089887640449438
        },
        "upstream_rq_total": {
          "value": 287700,
          "perSecond": 3232.5842696629215
        }
      }
    },
    {
      "testName": "scaling down httproutes to 100 with 20 routes per hostname at 500 rps",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-down",
      "throughput": 2019.6516853932585,
      "totalRequests": 179749,
      "latency": {
        "max": 101.289983,
        "min": 0.11330799999999999,
        "mean": 0.266825,
        "pstdev": 1.447675,
        "percentiles": {
          "p50": 0.195159,
          "p75": 0.22639099999999998,
          "p80": 0.233199,
          "p90": 0.257639,
          "p95": 0.327135,
          "p99": 1.3599350000000001,
          "p999": 8.865791
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 168.8359375,
            "min": 150.015625,
            "mean": 154.7328125
          },
          "cpu": {
            "max": 0.8666666666668259,
            "min": 0.7333333333332348,
            "mean": 0.7682539682539821
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 94.99609375,
            "min": 94.42578125,
            "mean": 94.75481770833333
          },
          "cpu": {
            "max": 13.593500446873257,
            "min": 13.593500446873257,
            "mean": 13.593500446873257
          }
        }
      },
      "poolOverflow": 251,
      "upstreamConnections": 45,
      "counters": {
        "benchmark.http_2xx": {
          "value": 179749,
          "perSecond": 2019.6516853932585
        },
        "benchmark.pool_overflow": {
          "value": 251,
          "perSecond": 2.8202247191011236
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
          "value": 45,
          "perSecond": 0.5056179775280899
        },
        "upstream_cx_rx_bytes_total": {
          "value": 28220593,
          "perSecond": 317085.3146067416
        },
        "upstream_cx_total": {
          "value": 45,
          "perSecond": 0.5056179775280899
        },
        "upstream_cx_tx_bytes_total": {
          "value": 8088705,
          "perSecond": 90884.32584269663
        },
        "upstream_rq_pending_overflow": {
          "value": 251,
          "perSecond": 2.8202247191011236
        },
        "upstream_rq_pending_total": {
          "value": 45,
          "perSecond": 0.5056179775280899
        },
        "upstream_rq_total": {
          "value": 179749,
          "perSecond": 2019.6516853932585
        }
      }
    },
    {
      "testName": "scaling down httproutes to 50 with 10 routes per hostname at 300 rps",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-down",
      "throughput": 1199.3666666666666,
      "totalRequests": 107943,
      "latency": {
        "max": 71.811071,
        "min": 0.118576,
        "mean": 0.249425,
        "pstdev": 0.5145660000000001,
        "percentiles": {
          "p50": 0.227335,
          "p75": 0.245511,
          "p80": 0.250007,
          "p90": 0.268527,
          "p95": 0.309487,
          "p99": 0.957759,
          "p999": 3.483519
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 154.765625,
            "min": 142.01171875,
            "mean": 150.2203125
          },
          "cpu": {
            "max": 0.8666666666666363,
            "min": 0.7999999999998407,
            "mean": 0.8144927536231809
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 94.83203125,
            "min": 94.21875,
            "mean": 94.48984375
          },
          "cpu": {
            "max": 14.797097151775374,
            "min": 10.67585897672007,
            "mean": 14.344927245939049
          }
        }
      },
      "poolOverflow": 57,
      "upstreamConnections": 20,
      "counters": {
        "benchmark.http_2xx": {
          "value": 107943,
          "perSecond": 1199.3666666666666
        },
        "benchmark.pool_overflow": {
          "value": 57,
          "perSecond": 0.6333333333333333
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
          "value": 16947051,
          "perSecond": 188300.56666666668
        },
        "upstream_cx_total": {
          "value": 20,
          "perSecond": 0.2222222222222222
        },
        "upstream_cx_tx_bytes_total": {
          "value": 4857435,
          "perSecond": 53971.5
        },
        "upstream_rq_pending_overflow": {
          "value": 57,
          "perSecond": 0.6333333333333333
        },
        "upstream_rq_pending_total": {
          "value": 20,
          "perSecond": 0.2222222222222222
        },
        "upstream_rq_total": {
          "value": 107943,
          "perSecond": 1199.3666666666666
        }
      }
    },
    {
      "testName": "scaling down httproutes to 10 with 2 routes per hostname at 100 rps",
      "routes": 10,
      "routesPerHostname": 2,
      "phase": "scaling-down",
      "throughput": 404.4943820224719,
      "totalRequests": 36000,
      "latency": {
        "max": 30.033919,
        "min": 0.16045600000000002,
        "mean": 0.27472199999999997,
        "pstdev": 0.471438,
        "percentiles": {
          "p50": 0.243783,
          "p75": 0.27115100000000003,
          "p80": 0.27780699999999997,
          "p90": 0.304271,
          "p95": 0.343903,
          "p99": 0.9235829999999999,
          "p999": 3.439871
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 149.375,
            "min": 145.91015625,
            "mean": 147.73567708333334
          },
          "cpu": {
            "max": 0.8666666666668259,
            "min": 0.7333333333332348,
            "mean": 0.8166666666666345
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 94.4296875,
            "min": 94.1328125,
            "mean": 94.28671875
          },
          "cpu": {
            "max": 5.485480122324283,
            "min": 3.3376959210565964,
            "mean": 4.402358673280198
          }
        }
      },
      "poolOverflow": 0,
      "upstreamConnections": 11,
      "counters": {
        "benchmark.http_2xx": {
          "value": 36000,
          "perSecond": 404.4943820224719
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
          "value": 5652000,
          "perSecond": 63505.61797752809
        },
        "upstream_cx_total": {
          "value": 11,
          "perSecond": 0.12359550561797752
        },
        "upstream_cx_tx_bytes_total": {
          "value": 1620000,
          "perSecond": 18202.247191011236
        },
        "upstream_rq_pending_total": {
          "value": 11,
          "perSecond": 0.12359550561797752
        },
        "upstream_rq_total": {
          "value": 36000,
          "perSecond": 404.4943820224719
        }
      }
    }
  ]
};
