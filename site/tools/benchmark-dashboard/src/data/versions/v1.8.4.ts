import { TestSuite } from '../types';

// Benchmark data extracted from release artifact for version 1.8.4
// Generated from benchmark_result.json

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.8.4",
    "runId": "1.8.4-release-2026-08-28",
    "date": "2026-08-28T11:03:57Z",
    "environment": "GitHub Release",
    "description": "Benchmark results for version 1.8.4 from release artifacts",
    "downloadUrl": "https://github.com/envoyproxy/gateway/releases/download/v1.8.4/benchmark_report.zip",
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
      "throughput": 404.3820224719101,
      "totalRequests": 35990,
      "latency": {
        "max": 119.939071,
        "min": 0.21596,
        "mean": 0.335991,
        "pstdev": 0.709551,
        "percentiles": {
          "p50": 0.307039,
          "p75": 0.339119,
          "p80": 0.349535,
          "p90": 0.382623,
          "p95": 0.420399,
          "p99": 0.723583,
          "p999": 3.1151350000000004
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 134.03125,
            "min": 110.859375,
            "mean": 129.82916666666668
          },
          "cpu": {
            "max": 0.3333333333333337,
            "min": 0.1333333333333327,
            "mean": 0.26944444444444454
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 20.62890625,
            "min": 7.10546875,
            "mean": 19.091145833333332
          },
          "cpu": {
            "max": 7.527106454212908,
            "min": 7.340302364646593,
            "mean": 7.414313003920752
          }
        }
      },
      "poolOverflow": 10,
      "upstreamConnections": 9,
      "counters": {
        "benchmark.http_2xx": {
          "value": 35990,
          "perSecond": 404.3820224719101
        },
        "benchmark.pool_overflow": {
          "value": 10,
          "perSecond": 0.11235955056179775
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
          "value": 9,
          "perSecond": 0.10112359550561797
        },
        "upstream_cx_rx_bytes_total": {
          "value": 5650430,
          "perSecond": 63487.97752808989
        },
        "upstream_cx_total": {
          "value": 9,
          "perSecond": 0.10112359550561797
        },
        "upstream_cx_tx_bytes_total": {
          "value": 1619550,
          "perSecond": 18197.191011235955
        },
        "upstream_rq_pending_overflow": {
          "value": 10,
          "perSecond": 0.11235955056179775
        },
        "upstream_rq_pending_total": {
          "value": 9,
          "perSecond": 0.10112359550561797
        },
        "upstream_rq_total": {
          "value": 35990,
          "perSecond": 404.3820224719101
        }
      }
    },
    {
      "testName": "scaling up httproutes to 50 with 10 routes per hostname at 300 rps",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-up",
      "throughput": 1213.2359550561798,
      "totalRequests": 107978,
      "latency": {
        "max": 42.352638999999996,
        "min": 0.17556,
        "mean": 0.29641500000000004,
        "pstdev": 0.42214900000000005,
        "percentiles": {
          "p50": 0.277631,
          "p75": 0.300223,
          "p80": 0.30684700000000004,
          "p90": 0.331695,
          "p95": 0.37281499999999995,
          "p99": 0.660159,
          "p999": 2.787071
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 138.171875,
            "min": 132.578125,
            "mean": 136.0296875
          },
          "cpu": {
            "max": 0.40000000000000047,
            "min": 0.26666666666666694,
            "mean": 0.3478260869565219
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 24.93359375,
            "min": 20.62890625,
            "mean": 23.972005208333332
          },
          "cpu": {
            "max": 19.59109641517803,
            "min": 19.16007167906278,
            "mean": 19.375584047120405
          }
        }
      },
      "poolOverflow": 22,
      "upstreamConnections": 19,
      "counters": {
        "benchmark.http_2xx": {
          "value": 107978,
          "perSecond": 1213.2359550561798
        },
        "benchmark.pool_overflow": {
          "value": 22,
          "perSecond": 0.24719101123595505
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
          "value": 16952546,
          "perSecond": 190478.04494382022
        },
        "upstream_cx_total": {
          "value": 19,
          "perSecond": 0.21348314606741572
        },
        "upstream_cx_tx_bytes_total": {
          "value": 4859010,
          "perSecond": 54595.61797752809
        },
        "upstream_rq_pending_overflow": {
          "value": 22,
          "perSecond": 0.24719101123595505
        },
        "upstream_rq_pending_total": {
          "value": 19,
          "perSecond": 0.21348314606741572
        },
        "upstream_rq_total": {
          "value": 107978,
          "perSecond": 1213.2359550561798
        }
      }
    },
    {
      "testName": "scaling up httproutes to 100 with 20 routes per hostname at 500 rps",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-up",
      "throughput": 2021.7865168539327,
      "totalRequests": 179939,
      "latency": {
        "max": 76.02175899999999,
        "min": 0.14932800000000002,
        "mean": 0.270185,
        "pstdev": 0.36920600000000003,
        "percentiles": {
          "p50": 0.235039,
          "p75": 0.27135899999999996,
          "p80": 0.281615,
          "p90": 0.323967,
          "p95": 0.388767,
          "p99": 0.767807,
          "p999": 3.1786230000000004
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 144.93359375,
            "min": 137.2265625,
            "mean": 142.44401041666666
          },
          "cpu": {
            "max": 0.4666666666666685,
            "min": 0.26666666666666694,
            "mean": 0.4151515151515149
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 29.03515625,
            "min": 24.8203125,
            "mean": 28.054817708333335
          },
          "cpu": {
            "max": 26.763908866607157,
            "min": 26.536555706521764,
            "mean": 26.574447899869327
          }
        }
      },
      "poolOverflow": 61,
      "upstreamConnections": 35,
      "counters": {
        "benchmark.http_2xx": {
          "value": 179939,
          "perSecond": 2021.7865168539327
        },
        "benchmark.pool_overflow": {
          "value": 61,
          "perSecond": 0.6853932584269663
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
          "value": 35,
          "perSecond": 0.39325842696629215
        },
        "upstream_cx_rx_bytes_total": {
          "value": 28250423,
          "perSecond": 317420.4831460674
        },
        "upstream_cx_total": {
          "value": 35,
          "perSecond": 0.39325842696629215
        },
        "upstream_cx_tx_bytes_total": {
          "value": 8097255,
          "perSecond": 90980.39325842696
        },
        "upstream_rq_pending_overflow": {
          "value": 61,
          "perSecond": 0.6853932584269663
        },
        "upstream_rq_pending_total": {
          "value": 35,
          "perSecond": 0.39325842696629215
        },
        "upstream_rq_total": {
          "value": 179939,
          "perSecond": 2021.7865168539327
        }
      }
    },
    {
      "testName": "scaling up httproutes to 300 with 60 routes per hostname at 800 rps",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-up",
      "throughput": 3199.5444444444443,
      "totalRequests": 287959,
      "latency": {
        "max": 38.354943000000006,
        "min": 0.13914400000000002,
        "mean": 0.266782,
        "pstdev": 0.416913,
        "percentiles": {
          "p50": 0.22828700000000002,
          "p75": 0.259431,
          "p80": 0.267759,
          "p90": 0.293839,
          "p95": 0.345151,
          "p99": 0.938175,
          "p999": 4.387327
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 155.93359375,
            "min": 147.65234375,
            "mean": 152.7953125
          },
          "cpu": {
            "max": 21.066666666666663,
            "min": 0.4666666666666626,
            "mean": 4.8066666666666675
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 41.57421875,
            "min": 35.33203125,
            "mean": 40.686458333333334
          },
          "cpu": {
            "max": 41.85195254002993,
            "min": 28.207594329848668,
            "mean": 39.56249482663179
          }
        }
      },
      "poolOverflow": 40,
      "upstreamConnections": 76,
      "counters": {
        "benchmark.http_2xx": {
          "value": 287959,
          "perSecond": 3199.5444444444443
        },
        "benchmark.pool_overflow": {
          "value": 40,
          "perSecond": 0.4444444444444444
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
          "value": 76,
          "perSecond": 0.8444444444444444
        },
        "upstream_cx_rx_bytes_total": {
          "value": 45209563,
          "perSecond": 502328.47777777776
        },
        "upstream_cx_total": {
          "value": 76,
          "perSecond": 0.8444444444444444
        },
        "upstream_cx_tx_bytes_total": {
          "value": 12958200,
          "perSecond": 143980
        },
        "upstream_rq_pending_overflow": {
          "value": 40,
          "perSecond": 0.4444444444444444
        },
        "upstream_rq_pending_total": {
          "value": 76,
          "perSecond": 0.8444444444444444
        },
        "upstream_rq_total": {
          "value": 287960,
          "perSecond": 3199.5555555555557
        }
      }
    },
    {
      "testName": "scaling up httproutes to 500 with 100 routes per hostname at 1000 rps",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-up",
      "throughput": 4043.4606741573034,
      "totalRequests": 359868,
      "latency": {
        "max": 51.077118999999996,
        "min": 0.13896,
        "mean": 0.357221,
        "pstdev": 0.485623,
        "percentiles": {
          "p50": 0.312367,
          "p75": 0.387791,
          "p80": 0.409503,
          "p90": 0.487663,
          "p95": 0.569759,
          "p99": 1.1097590000000002,
          "p999": 4.848639
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 171.75,
            "min": 160.265625,
            "mean": 168.56770833333334
          },
          "cpu": {
            "max": 36.199999999999996,
            "min": 0.5333333333333456,
            "mean": 7.077777777777774
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 57.6015625,
            "min": 45.01171875,
            "mean": 55.521484375
          },
          "cpu": {
            "max": 46.250173445276715,
            "min": 23.767157814197066,
            "mean": 38.85715305474797
          }
        }
      },
      "poolOverflow": 131,
      "upstreamConnections": 68,
      "counters": {
        "benchmark.http_2xx": {
          "value": 359868,
          "perSecond": 4043.4606741573034
        },
        "benchmark.pool_overflow": {
          "value": 131,
          "perSecond": 1.4719101123595506
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
          "value": 68,
          "perSecond": 0.7640449438202247
        },
        "upstream_cx_rx_bytes_total": {
          "value": 56499276,
          "perSecond": 634823.3258426966
        },
        "upstream_cx_total": {
          "value": 68,
          "perSecond": 0.7640449438202247
        },
        "upstream_cx_tx_bytes_total": {
          "value": 16194105,
          "perSecond": 181956.23595505618
        },
        "upstream_rq_pending_overflow": {
          "value": 131,
          "perSecond": 1.4719101123595506
        },
        "upstream_rq_pending_total": {
          "value": 68,
          "perSecond": 0.7640449438202247
        },
        "upstream_rq_total": {
          "value": 359869,
          "perSecond": 4043.4719101123596
        }
      }
    },
    {
      "testName": "scaling up httproutes to 1000 with 200 routes per hostname at 2000 rps",
      "routes": 1000,
      "routesPerHostname": 200,
      "phase": "scaling-up",
      "throughput": 7999.222222222223,
      "totalRequests": 719930,
      "latency": {
        "max": 297.467903,
        "min": 0.12615200000000001,
        "mean": 1.27271,
        "pstdev": 6.602606,
        "percentiles": {
          "p50": 0.343039,
          "p75": 0.522191,
          "p80": 0.614527,
          "p90": 1.161919,
          "p95": 2.786431,
          "p99": 25.415679,
          "p999": 75.26809499999999
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 190.375,
            "min": 174.55078125,
            "mean": 184.20924479166666
          },
          "cpu": {
            "max": 70.79999999999998,
            "min": 0.5999999999999753,
            "mean": 12.76666666666666
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 102.0859375,
            "min": 77.51171875,
            "mean": 97.30403645833333
          },
          "cpu": {
            "max": 81.8812197636608,
            "min": 46.92947128945742,
            "mean": 75.75004598315773
          }
        }
      },
      "poolOverflow": 68,
      "upstreamConnections": 332,
      "counters": {
        "benchmark.http_2xx": {
          "value": 719930,
          "perSecond": 7999.222222222223
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
          "value": 113029010,
          "perSecond": 1255877.888888889
        },
        "upstream_cx_total": {
          "value": 332,
          "perSecond": 3.688888888888889
        },
        "upstream_cx_tx_bytes_total": {
          "value": 32396940,
          "perSecond": 359966
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
          "value": 719932,
          "perSecond": 7999.2444444444445
        }
      }
    },
    {
      "testName": "scaling down httproutes to 500 with 100 routes per hostname at 1000 rps",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-down",
      "throughput": 3998.5333333333333,
      "totalRequests": 359868,
      "latency": {
        "max": 57.128959,
        "min": 0.13352,
        "mean": 0.269393,
        "pstdev": 0.349684,
        "percentiles": {
          "p50": 0.22182300000000002,
          "p75": 0.250807,
          "p80": 0.261735,
          "p90": 0.34990299999999996,
          "p95": 0.461871,
          "p99": 1.044575,
          "p999": 3.168511
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 204.4296875,
            "min": 168.5078125,
            "mean": 176.57421875
          },
          "cpu": {
            "max": 5.533333333333322,
            "min": 0.7333333333333295,
            "mean": 1.7333333333333294
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 103.16796875,
            "min": 101.12109375,
            "mean": 102.660546875
          },
          "cpu": {
            "max": 48.8563925617659,
            "min": 48.8563925617659,
            "mean": 48.8563925617659
          }
        }
      },
      "poolOverflow": 132,
      "upstreamConnections": 48,
      "counters": {
        "benchmark.http_2xx": {
          "value": 359868,
          "perSecond": 3998.5333333333333
        },
        "benchmark.pool_overflow": {
          "value": 132,
          "perSecond": 1.4666666666666666
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
          "value": 56499276,
          "perSecond": 627769.7333333333
        },
        "upstream_cx_total": {
          "value": 48,
          "perSecond": 0.5333333333333333
        },
        "upstream_cx_tx_bytes_total": {
          "value": 16194060,
          "perSecond": 179934
        },
        "upstream_rq_pending_overflow": {
          "value": 132,
          "perSecond": 1.4666666666666666
        },
        "upstream_rq_pending_total": {
          "value": 48,
          "perSecond": 0.5333333333333333
        },
        "upstream_rq_total": {
          "value": 359868,
          "perSecond": 3998.5333333333333
        }
      }
    },
    {
      "testName": "scaling down httproutes to 300 with 60 routes per hostname at 800 rps",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-down",
      "throughput": 3199.277777777778,
      "totalRequests": 287935,
      "latency": {
        "max": 38.062079,
        "min": 0.150232,
        "mean": 0.260147,
        "pstdev": 0.334088,
        "percentiles": {
          "p50": 0.226327,
          "p75": 0.258351,
          "p80": 0.265855,
          "p90": 0.290351,
          "p95": 0.336735,
          "p99": 0.8700789999999999,
          "p999": 3.254399
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 173.04296875,
            "min": 158.703125,
            "mean": 164.41171875
          },
          "cpu": {
            "max": 0.8000000000000304,
            "min": 0.7333333333333295,
            "mean": 0.778787878787881
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 102.953125,
            "min": 102.49609375,
            "mean": 102.73307291666667
          },
          "cpu": {
            "max": 41.558809195889545,
            "min": 38.93407127429819,
            "mean": 39.52988357994586
          }
        }
      },
      "poolOverflow": 65,
      "upstreamConnections": 49,
      "counters": {
        "benchmark.http_2xx": {
          "value": 287935,
          "perSecond": 3199.277777777778
        },
        "benchmark.pool_overflow": {
          "value": 65,
          "perSecond": 0.7222222222222222
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
          "value": 49,
          "perSecond": 0.5444444444444444
        },
        "upstream_cx_rx_bytes_total": {
          "value": 45205795,
          "perSecond": 502286.6111111111
        },
        "upstream_cx_total": {
          "value": 49,
          "perSecond": 0.5444444444444444
        },
        "upstream_cx_tx_bytes_total": {
          "value": 12957075,
          "perSecond": 143967.5
        },
        "upstream_rq_pending_overflow": {
          "value": 65,
          "perSecond": 0.7222222222222222
        },
        "upstream_rq_pending_total": {
          "value": 49,
          "perSecond": 0.5444444444444444
        },
        "upstream_rq_total": {
          "value": 287935,
          "perSecond": 3199.277777777778
        }
      }
    },
    {
      "testName": "scaling down httproutes to 100 with 20 routes per hostname at 500 rps",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-down",
      "throughput": 2021.685393258427,
      "totalRequests": 179930,
      "latency": {
        "max": 25.540607,
        "min": 0.16216,
        "mean": 0.26710500000000004,
        "pstdev": 0.255345,
        "percentiles": {
          "p50": 0.24746300000000002,
          "p75": 0.274111,
          "p80": 0.280095,
          "p90": 0.298975,
          "p95": 0.331743,
          "p99": 0.697407,
          "p999": 2.913919
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 164.578125,
            "min": 147.62109375,
            "mean": 153.78059895833334
          },
          "cpu": {
            "max": 0.7333333333333297,
            "min": 0.6666666666666762,
            "mean": 0.716666666666666
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 103.0625,
            "min": 102.37109375,
            "mean": 102.634765625
          },
          "cpu": {
            "max": 30.921305144652383,
            "min": 15.096517370517995,
            "mean": 22.580741766870943
          }
        }
      },
      "poolOverflow": 70,
      "upstreamConnections": 36,
      "counters": {
        "benchmark.http_2xx": {
          "value": 179930,
          "perSecond": 2021.685393258427
        },
        "benchmark.pool_overflow": {
          "value": 70,
          "perSecond": 0.7865168539325843
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
          "value": 36,
          "perSecond": 0.4044943820224719
        },
        "upstream_cx_rx_bytes_total": {
          "value": 28249010,
          "perSecond": 317404.606741573
        },
        "upstream_cx_total": {
          "value": 36,
          "perSecond": 0.4044943820224719
        },
        "upstream_cx_tx_bytes_total": {
          "value": 8096850,
          "perSecond": 90975.84269662922
        },
        "upstream_rq_pending_overflow": {
          "value": 70,
          "perSecond": 0.7865168539325843
        },
        "upstream_rq_pending_total": {
          "value": 36,
          "perSecond": 0.4044943820224719
        },
        "upstream_rq_total": {
          "value": 179930,
          "perSecond": 2021.685393258427
        }
      }
    },
    {
      "testName": "scaling down httproutes to 50 with 10 routes per hostname at 300 rps",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-down",
      "throughput": 1213,
      "totalRequests": 107957,
      "latency": {
        "max": 71.528447,
        "min": 0.18027200000000002,
        "mean": 0.308976,
        "pstdev": 0.703636,
        "percentiles": {
          "p50": 0.281231,
          "p75": 0.302415,
          "p80": 0.308991,
          "p90": 0.334063,
          "p95": 0.375839,
          "p99": 0.700223,
          "p999": 3.311359
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 154.62109375,
            "min": 146.0703125,
            "mean": 149.11302083333334
          },
          "cpu": {
            "max": 0.8666666666666836,
            "min": 0.7333333333333295,
            "mean": 0.7454545454545435
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 103.32421875,
            "min": 102.3515625,
            "mean": 102.66393229166667
          },
          "cpu": {
            "max": 20.1480397327706,
            "min": 13.323227073956772,
            "mean": 18.36455007420628
          }
        }
      },
      "poolOverflow": 43,
      "upstreamConnections": 29,
      "counters": {
        "benchmark.http_2xx": {
          "value": 107957,
          "perSecond": 1213
        },
        "benchmark.pool_overflow": {
          "value": 43,
          "perSecond": 0.48314606741573035
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
          "value": 29,
          "perSecond": 0.3258426966292135
        },
        "upstream_cx_rx_bytes_total": {
          "value": 16949249,
          "perSecond": 190441
        },
        "upstream_cx_total": {
          "value": 29,
          "perSecond": 0.3258426966292135
        },
        "upstream_cx_tx_bytes_total": {
          "value": 4858065,
          "perSecond": 54585
        },
        "upstream_rq_pending_overflow": {
          "value": 43,
          "perSecond": 0.48314606741573035
        },
        "upstream_rq_pending_total": {
          "value": 29,
          "perSecond": 0.3258426966292135
        },
        "upstream_rq_total": {
          "value": 107957,
          "perSecond": 1213
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
        "max": 12.648447,
        "min": 0.218536,
        "mean": 0.338598,
        "pstdev": 0.243919,
        "percentiles": {
          "p50": 0.31575899999999996,
          "p75": 0.34166300000000005,
          "p80": 0.350511,
          "p90": 0.383759,
          "p95": 0.419871,
          "p99": 0.757855,
          "p999": 3.9430389999999997
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 148.30078125,
            "min": 135.6015625,
            "mean": 143.54713541666666
          },
          "cpu": {
            "max": 0.8666666666666836,
            "min": 0.7333333333333295,
            "mean": 0.7749999999999978
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 102.8828125,
            "min": 102.3515625,
            "mean": 102.51809895833334
          },
          "cpu": {
            "max": 7.531821335646289,
            "min": 4.956691620988822,
            "mean": 7.045568808088292
          }
        }
      },
      "poolOverflow": 0,
      "upstreamConnections": 6,
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
          "value": 6,
          "perSecond": 0.06741573033707865
        },
        "upstream_cx_rx_bytes_total": {
          "value": 5652000,
          "perSecond": 63505.61797752809
        },
        "upstream_cx_total": {
          "value": 6,
          "perSecond": 0.06741573033707865
        },
        "upstream_cx_tx_bytes_total": {
          "value": 1620000,
          "perSecond": 18202.247191011236
        },
        "upstream_rq_pending_total": {
          "value": 6,
          "perSecond": 0.06741573033707865
        },
        "upstream_rq_total": {
          "value": 36000,
          "perSecond": 404.4943820224719
        }
      }
    }
  ]
};
