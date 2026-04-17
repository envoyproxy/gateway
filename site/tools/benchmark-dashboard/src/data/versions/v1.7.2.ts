import { TestSuite } from '../types';

// Benchmark data extracted from release artifact for version 1.7.2
// Generated from benchmark_result.json

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.7.2",
    "runId": "1.7.2-release-2026-04-17",
    "date": "2026-04-17T01:33:18Z",
    "environment": "GitHub Release",
    "description": "Benchmark results for version 1.7.2 from release artifacts",
    "downloadUrl": "https://github.com/envoyproxy/gateway/releases/download/v1.7.2/benchmark_report.zip",
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
      "throughput": 404.35955056179773,
      "totalRequests": 35988,
      "latency": {
        "max": 101.68729499999999,
        "min": 0.355904,
        "mean": 0.49871699999999997,
        "pstdev": 0.9805560000000001,
        "percentiles": {
          "p50": 0.457103,
          "p75": 0.47526300000000005,
          "p80": 0.48102300000000003,
          "p90": 0.501631,
          "p95": 0.535551,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 134.72265625,
            "min": 116.70703125,
            "mean": 131.82174479166667
          },
          "cpu": {
            "max": 1.1999999999999995,
            "min": 0.33333333333333287,
            "mean": 0.5690476190476189
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 25.734375,
            "min": 6.9375,
            "mean": 22.484895833333333
          },
          "cpu": {
            "max": 10.60714949795463,
            "min": 6.173356863557859,
            "mean": 8.817763315828271
          }
        }
      },
      "poolOverflow": 12,
      "upstreamConnections": 11,
      "counters": {
        "benchmark.http_2xx": {
          "value": 35988,
          "perSecond": 404.35955056179773
        },
        "benchmark.pool_overflow": {
          "value": 12,
          "perSecond": 0.1348314606741573
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
          "value": 5650116,
          "perSecond": 63484.449438202246
        },
        "upstream_cx_total": {
          "value": 11,
          "perSecond": 0.12359550561797752
        },
        "upstream_cx_tx_bytes_total": {
          "value": 1619460,
          "perSecond": 18196.1797752809
        },
        "upstream_rq_pending_overflow": {
          "value": 12,
          "perSecond": 0.1348314606741573
        },
        "upstream_rq_pending_total": {
          "value": 11,
          "perSecond": 0.12359550561797752
        },
        "upstream_rq_total": {
          "value": 35988,
          "perSecond": 404.35955056179773
        }
      }
    },
    {
      "testName": "scaling up httproutes to 50 with 10 routes per hostname at 300 rps",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-up",
      "throughput": 1213.2022471910113,
      "totalRequests": 107975,
      "latency": {
        "max": 52.019199,
        "min": 0.346128,
        "mean": 0.48446300000000003,
        "pstdev": 0.478466,
        "percentiles": {
          "p50": 0.447967,
          "p75": 0.463743,
          "p80": 0.468719,
          "p90": 0.49182299999999995,
          "p95": 0.539551,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 146.11328125,
            "min": 135.58984375,
            "mean": 141.50143229166667
          },
          "cpu": {
            "max": 1.9999999999999987,
            "min": 0.40000000000000036,
            "mean": 0.7711111111111112
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 29.76953125,
            "min": 27.9296875,
            "mean": 29.26328125
          },
          "cpu": {
            "max": 30.24145243537894,
            "min": 30.22350314191992,
            "mean": 30.22948623973959
          }
        }
      },
      "poolOverflow": 25,
      "upstreamConnections": 24,
      "counters": {
        "benchmark.http_2xx": {
          "value": 107975,
          "perSecond": 1213.2022471910113
        },
        "benchmark.pool_overflow": {
          "value": 25,
          "perSecond": 0.2808988764044944
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
          "value": 24,
          "perSecond": 0.2696629213483146
        },
        "upstream_cx_rx_bytes_total": {
          "value": 16952075,
          "perSecond": 190472.75280898876
        },
        "upstream_cx_total": {
          "value": 24,
          "perSecond": 0.2696629213483146
        },
        "upstream_cx_tx_bytes_total": {
          "value": 4858875,
          "perSecond": 54594.10112359551
        },
        "upstream_rq_pending_overflow": {
          "value": 25,
          "perSecond": 0.2808988764044944
        },
        "upstream_rq_pending_total": {
          "value": 24,
          "perSecond": 0.2696629213483146
        },
        "upstream_rq_total": {
          "value": 107975,
          "perSecond": 1213.2022471910113
        }
      }
    },
    {
      "testName": "scaling up httproutes to 100 with 20 routes per hostname at 500 rps",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-up",
      "throughput": 2020.9550561797753,
      "totalRequests": 179865,
      "latency": {
        "max": 104.263679,
        "min": 0.33550399999999997,
        "mean": 0.539038,
        "pstdev": 0.863241,
        "percentiles": {
          "p50": 0.465887,
          "p75": 0.521359,
          "p80": 0.549887,
          "p90": 0.658143,
          "p95": 0.761823,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 148.65625,
            "min": 137.18359375,
            "mean": 144.56315104166666
          },
          "cpu": {
            "max": 2.7333333333333343,
            "min": 0.5333333333333339,
            "mean": 0.9666666666666668
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 36.41015625,
            "min": 33.765625,
            "mean": 35.787109375
          },
          "cpu": {
            "max": 45.66901459854013,
            "min": 45.04842244701349,
            "mean": 45.243921833694976
          }
        }
      },
      "poolOverflow": 135,
      "upstreamConnections": 48,
      "counters": {
        "benchmark.http_2xx": {
          "value": 179865,
          "perSecond": 2020.9550561797753
        },
        "benchmark.pool_overflow": {
          "value": 135,
          "perSecond": 1.5168539325842696
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
          "value": 48,
          "perSecond": 0.5393258426966292
        },
        "upstream_cx_rx_bytes_total": {
          "value": 28238805,
          "perSecond": 317289.9438202247
        },
        "upstream_cx_total": {
          "value": 48,
          "perSecond": 0.5393258426966292
        },
        "upstream_cx_tx_bytes_total": {
          "value": 8093925,
          "perSecond": 90942.97752808989
        },
        "upstream_rq_pending_overflow": {
          "value": 135,
          "perSecond": 1.5168539325842696
        },
        "upstream_rq_pending_total": {
          "value": 48,
          "perSecond": 0.5393258426966292
        },
        "upstream_rq_total": {
          "value": 179865,
          "perSecond": 2020.9550561797753
        }
      }
    },
    {
      "testName": "scaling up httproutes to 300 with 60 routes per hostname at 800 rps",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-up",
      "throughput": 3197.4222222222224,
      "totalRequests": 287768,
      "latency": {
        "max": 99.41401499999999,
        "min": 0.319616,
        "mean": 1.147704,
        "pstdev": 4.4813860000000005,
        "percentiles": {
          "p50": 0.594687,
          "p75": 0.687487,
          "p80": 0.742047,
          "p90": 1.042047,
          "p95": 1.562431,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 154.125,
            "min": 140.8671875,
            "mean": 151.02200520833333
          },
          "cpu": {
            "max": 4.9333333333333345,
            "min": 0.5999999999999991,
            "mean": 1.5511111111111098
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 68.3515625,
            "min": 57.90234375,
            "mean": 64.46614583333333
          },
          "cpu": {
            "max": 68.30009668842146,
            "min": 43.403573110547036,
            "mean": 62.29201721165111
          }
        }
      },
      "poolOverflow": 56,
      "upstreamConnections": 264,
      "counters": {
        "benchmark.http_2xx": {
          "value": 287768,
          "perSecond": 3197.4222222222224
        },
        "benchmark.pool_overflow": {
          "value": 56,
          "perSecond": 0.6222222222222222
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
          "value": 264,
          "perSecond": 2.933333333333333
        },
        "upstream_cx_rx_bytes_total": {
          "value": 45179576,
          "perSecond": 501995.2888888889
        },
        "upstream_cx_total": {
          "value": 264,
          "perSecond": 2.933333333333333
        },
        "upstream_cx_tx_bytes_total": {
          "value": 12957480,
          "perSecond": 143972
        },
        "upstream_rq_pending_overflow": {
          "value": 56,
          "perSecond": 0.6222222222222222
        },
        "upstream_rq_pending_total": {
          "value": 264,
          "perSecond": 2.933333333333333
        },
        "upstream_rq_total": {
          "value": 287944,
          "perSecond": 3199.3777777777777
        }
      }
    },
    {
      "testName": "scaling up httproutes to 500 with 100 routes per hostname at 1000 rps",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-up",
      "throughput": 3999.5777777777776,
      "totalRequests": 359962,
      "latency": {
        "max": 386.252799,
        "min": 0.307408,
        "mean": 2.964271,
        "pstdev": 14.94732,
        "percentiles": {
          "p50": 0.611135,
          "p75": 0.7879989999999999,
          "p80": 0.863615,
          "p90": 1.354687,
          "p95": 3.6558070000000003,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 171.26953125,
            "min": 161.11328125,
            "mean": 166.31549479166668
          },
          "cpu": {
            "max": 6.866666666666673,
            "min": 0.86666666666666,
            "mean": 2.0600000000000005
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 95.4609375,
            "min": 84.73046875,
            "mean": 91.400390625
          },
          "cpu": {
            "max": 82.5813865584955,
            "min": 45.283503319351,
            "mean": 72.3024265227598
          }
        }
      },
      "poolOverflow": 30,
      "upstreamConnections": 370,
      "counters": {
        "benchmark.http_2xx": {
          "value": 359962,
          "perSecond": 3999.5777777777776
        },
        "benchmark.pool_overflow": {
          "value": 30,
          "perSecond": 0.3333333333333333
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
          "value": 370,
          "perSecond": 4.111111111111111
        },
        "upstream_cx_rx_bytes_total": {
          "value": 56514034,
          "perSecond": 627933.7111111111
        },
        "upstream_cx_total": {
          "value": 370,
          "perSecond": 4.111111111111111
        },
        "upstream_cx_tx_bytes_total": {
          "value": 16198650,
          "perSecond": 179985
        },
        "upstream_rq_pending_overflow": {
          "value": 30,
          "perSecond": 0.3333333333333333
        },
        "upstream_rq_pending_total": {
          "value": 370,
          "perSecond": 4.111111111111111
        },
        "upstream_rq_total": {
          "value": 359970,
          "perSecond": 3999.6666666666665
        }
      }
    },
    {
      "testName": "scaling up httproutes to 1000 with 200 routes per hostname at 2000 rps",
      "routes": 1000,
      "routesPerHostname": 200,
      "phase": "scaling-up",
      "throughput": 5057.266666666666,
      "totalRequests": 455154,
      "latency": {
        "max": 1285.9473910000002,
        "min": 1.371584,
        "mean": 59.338332,
        "pstdev": 19.470976,
        "percentiles": {
          "p50": 57.772031,
          "p75": 66.177023,
          "p80": 67.76831899999999,
          "p90": 71.90118299999999,
          "p95": 78.733311,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 188.16015625,
            "min": 171.28515625,
            "mean": 182.01666666666668
          },
          "cpu": {
            "max": 14.13333333333327,
            "min": 0.8666666666666363,
            "mean": 2.964444444444452
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 145.08984375,
            "min": 139.8203125,
            "mean": 143.387890625
          },
          "cpu": {
            "max": 99.22699520543796,
            "min": 60.78907448875107,
            "mean": 92.6174535512916
          }
        }
      },
      "poolOverflow": 98,
      "upstreamConnections": 302,
      "counters": {
        "benchmark.http_2xx": {
          "value": 455154,
          "perSecond": 5057.266666666666
        },
        "benchmark.pool_overflow": {
          "value": 98,
          "perSecond": 1.0888888888888888
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
          "value": 302,
          "perSecond": 3.3555555555555556
        },
        "upstream_cx_rx_bytes_total": {
          "value": 71459178,
          "perSecond": 793990.8666666667
        },
        "upstream_cx_total": {
          "value": 302,
          "perSecond": 3.3555555555555556
        },
        "upstream_cx_tx_bytes_total": {
          "value": 20495520,
          "perSecond": 227728
        },
        "upstream_rq_pending_overflow": {
          "value": 98,
          "perSecond": 1.0888888888888888
        },
        "upstream_rq_pending_total": {
          "value": 302,
          "perSecond": 3.3555555555555556
        },
        "upstream_rq_total": {
          "value": 455456,
          "perSecond": 5060.622222222222
        }
      }
    },
    {
      "testName": "scaling down httproutes to 500 with 100 routes per hostname at 1000 rps",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-down",
      "throughput": 3999.233333333333,
      "totalRequests": 359931,
      "latency": {
        "max": 385.597439,
        "min": 0.31702400000000003,
        "mean": 2.6073619999999997,
        "pstdev": 12.422353000000001,
        "percentiles": {
          "p50": 0.617951,
          "p75": 0.7978230000000001,
          "p80": 0.873503,
          "p90": 1.409791,
          "p95": 3.412095,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 176.8359375,
            "min": 171.98046875,
            "mean": 173.68333333333334
          },
          "cpu": {
            "max": 1.1333333333333446,
            "min": 1.0000000000000377,
            "mean": 1.106060606060612
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 145.45703125,
            "min": 143.30859375,
            "mean": 144.875
          },
          "cpu": {
            "max": 82.85775245913221,
            "min": 44.665104008759954,
            "mean": 76.23169817649351
          }
        }
      },
      "poolOverflow": 68,
      "upstreamConnections": 332,
      "counters": {
        "benchmark.http_2xx": {
          "value": 359931,
          "perSecond": 3999.233333333333
        },
        "benchmark.pool_overflow": {
          "value": 68,
          "perSecond": 0.7555555555555555
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
          "value": 332,
          "perSecond": 3.688888888888889
        },
        "upstream_cx_rx_bytes_total": {
          "value": 56509167,
          "perSecond": 627879.6333333333
        },
        "upstream_cx_total": {
          "value": 332,
          "perSecond": 3.688888888888889
        },
        "upstream_cx_tx_bytes_total": {
          "value": 16196940,
          "perSecond": 179966
        },
        "upstream_rq_pending_overflow": {
          "value": 68,
          "perSecond": 0.7555555555555555
        },
        "upstream_rq_pending_total": {
          "value": 332,
          "perSecond": 3.688888888888889
        },
        "upstream_rq_total": {
          "value": 359932,
          "perSecond": 3999.2444444444445
        }
      }
    },
    {
      "testName": "scaling down httproutes to 300 with 60 routes per hostname at 800 rps",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-down",
      "throughput": 3199.6111111111113,
      "totalRequests": 287965,
      "latency": {
        "max": 80.625663,
        "min": 0.322336,
        "mean": 1.0790170000000001,
        "pstdev": 3.7298549999999997,
        "percentiles": {
          "p50": 0.598207,
          "p75": 0.7040310000000001,
          "p80": 0.780127,
          "p90": 1.153535,
          "p95": 1.6243830000000001,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 175.10546875,
            "min": 161.109375,
            "mean": 163.90182291666667
          },
          "cpu": {
            "max": 1.1333333333333449,
            "min": 0.9333333333332423,
            "mean": 1.0638888888888844
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 143.6328125,
            "min": 141.3671875,
            "mean": 142.62200520833332
          },
          "cpu": {
            "max": 68.99807624218096,
            "min": 67.9146420263634,
            "mean": 68.48326157169933
          }
        }
      },
      "poolOverflow": 35,
      "upstreamConnections": 230,
      "counters": {
        "benchmark.http_2xx": {
          "value": 287965,
          "perSecond": 3199.6111111111113
        },
        "benchmark.pool_overflow": {
          "value": 35,
          "perSecond": 0.3888888888888889
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
          "value": 230,
          "perSecond": 2.5555555555555554
        },
        "upstream_cx_rx_bytes_total": {
          "value": 45210505,
          "perSecond": 502338.94444444444
        },
        "upstream_cx_total": {
          "value": 230,
          "perSecond": 2.5555555555555554
        },
        "upstream_cx_tx_bytes_total": {
          "value": 12958425,
          "perSecond": 143982.5
        },
        "upstream_rq_pending_overflow": {
          "value": 35,
          "perSecond": 0.3888888888888889
        },
        "upstream_rq_pending_total": {
          "value": 230,
          "perSecond": 2.5555555555555554
        },
        "upstream_rq_total": {
          "value": 287965,
          "perSecond": 3199.6111111111113
        }
      }
    },
    {
      "testName": "scaling down httproutes to 100 with 20 routes per hostname at 500 rps",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-down",
      "throughput": 1998.3777777777777,
      "totalRequests": 179854,
      "latency": {
        "max": 62.019583000000004,
        "min": 0.327328,
        "mean": 0.485828,
        "pstdev": 0.668839,
        "percentiles": {
          "p50": 0.423551,
          "p75": 0.443055,
          "p80": 0.449919,
          "p90": 0.479887,
          "p95": 0.620959,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 165.8046875,
            "min": 152.2734375,
            "mean": 156.540234375
          },
          "cpu": {
            "max": 1.2000000000000455,
            "min": 0.9999999999999432,
            "mean": 1.0666666666666655
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 141.15234375,
            "min": 139.859375,
            "mean": 140.26106770833334
          },
          "cpu": {
            "max": 45.485545566719395,
            "min": 31.71803855746116,
            "mean": 44.091576143376344
          }
        }
      },
      "poolOverflow": 146,
      "upstreamConnections": 48,
      "counters": {
        "benchmark.http_2xx": {
          "value": 179854,
          "perSecond": 1998.3777777777777
        },
        "benchmark.pool_overflow": {
          "value": 146,
          "perSecond": 1.6222222222222222
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
          "value": 48,
          "perSecond": 0.5333333333333333
        },
        "upstream_cx_rx_bytes_total": {
          "value": 28237078,
          "perSecond": 313745.31111111114
        },
        "upstream_cx_total": {
          "value": 48,
          "perSecond": 0.5333333333333333
        },
        "upstream_cx_tx_bytes_total": {
          "value": 8093430,
          "perSecond": 89927
        },
        "upstream_rq_pending_overflow": {
          "value": 146,
          "perSecond": 1.6222222222222222
        },
        "upstream_rq_pending_total": {
          "value": 48,
          "perSecond": 0.5333333333333333
        },
        "upstream_rq_total": {
          "value": 179854,
          "perSecond": 1998.3777777777777
        }
      }
    },
    {
      "testName": "scaling down httproutes to 50 with 10 routes per hostname at 300 rps",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-down",
      "throughput": 1213.0224719101125,
      "totalRequests": 107959,
      "latency": {
        "max": 64.030719,
        "min": 0.34096000000000004,
        "mean": 0.484964,
        "pstdev": 0.641611,
        "percentiles": {
          "p50": 0.440847,
          "p75": 0.45943900000000004,
          "p80": 0.464143,
          "p90": 0.48572699999999996,
          "p95": 0.531871,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 155.21484375,
            "min": 146.46484375,
            "mean": 150.92682291666668
          },
          "cpu": {
            "max": 1.2666666666666515,
            "min": 0.9999999999999432,
            "mean": 1.1166666666666694
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 140.078125,
            "min": 139.0859375,
            "mean": 139.482421875
          },
          "cpu": {
            "max": 30.431386861314387,
            "min": 28.98025197235646,
            "mean": 30.07745151829415
          }
        }
      },
      "poolOverflow": 41,
      "upstreamConnections": 28,
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
          "value": 28,
          "perSecond": 0.3146067415730337
        },
        "upstream_cx_rx_bytes_total": {
          "value": 16949563,
          "perSecond": 190444.52808988764
        },
        "upstream_cx_total": {
          "value": 28,
          "perSecond": 0.3146067415730337
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
          "value": 28,
          "perSecond": 0.3146067415730337
        },
        "upstream_rq_total": {
          "value": 107959,
          "perSecond": 1213.0224719101125
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
        "max": 45.062143,
        "min": 0.365328,
        "mean": 0.48585700000000004,
        "pstdev": 0.49808199999999997,
        "percentiles": {
          "p50": 0.452351,
          "p75": 0.46663899999999997,
          "p80": 0.46991900000000003,
          "p90": 0.48545499999999997,
          "p95": 0.515775,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 150.08984375,
            "min": 146.44921875,
            "mean": 147.88671875
          },
          "cpu": {
            "max": 1.2000000000000455,
            "min": 0.933333333333337,
            "mean": 1.0603174603174614
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 141.55078125,
            "min": 139.140625,
            "mean": 139.66901041666668
          },
          "cpu": {
            "max": 10.07732827919727,
            "min": 6.752312228571138,
            "mean": 9.510461390746068
          }
        }
      },
      "poolOverflow": 2,
      "upstreamConnections": 11,
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
          "value": 11,
          "perSecond": 0.12359550561797752
        },
        "upstream_cx_rx_bytes_total": {
          "value": 5651686,
          "perSecond": 63502.089887640446
        },
        "upstream_cx_total": {
          "value": 11,
          "perSecond": 0.12359550561797752
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
          "value": 11,
          "perSecond": 0.12359550561797752
        },
        "upstream_rq_total": {
          "value": 35998,
          "perSecond": 404.4719101123595
        }
      }
    }
  ]
};
