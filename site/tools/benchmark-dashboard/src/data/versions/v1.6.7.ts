import { TestSuite } from '../types';

// Benchmark data extracted from release artifact for version 1.6.7
// Generated from benchmark_result.json

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.6.7",
    "runId": "1.6.7-release-2026-04-27",
    "date": "2026-04-27T20:16:54Z",
    "environment": "GitHub Release",
    "description": "Benchmark results for version 1.6.7 from release artifacts",
    "downloadUrl": "https://github.com/envoyproxy/gateway/releases/download/v1.6.7/benchmark_report.zip",
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
      "throughput": 399.9222222222222,
      "totalRequests": 35993,
      "latency": {
        "max": 47.067135,
        "min": 0.378944,
        "mean": 0.49701,
        "pstdev": 0.540569,
        "percentiles": {
          "p50": 0.455535,
          "p75": 0.474767,
          "p80": 0.480495,
          "p90": 0.503663,
          "p95": 0.547071,
          "p99": 1.226175,
          "p999": 6.373119
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 139.1484375,
            "min": 119.875,
            "mean": 134.18802083333333
          },
          "cpu": {
            "max": 0.9333333333333335,
            "min": 0.3333333333333336,
            "mean": 0.5288888888888891
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 28.12109375,
            "min": 4.859375,
            "mean": 25.583072916666666
          },
          "cpu": {
            "max": 10.268738200125863,
            "min": 5.4939283145112325,
            "mean": 7.653110203221817
          }
        }
      },
      "poolOverflow": 7,
      "upstreamConnections": 12,
      "counters": {
        "benchmark.http_2xx": {
          "value": 35993,
          "perSecond": 399.9222222222222
        },
        "benchmark.pool_overflow": {
          "value": 7,
          "perSecond": 0.07777777777777778
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
          "value": 12,
          "perSecond": 0.13333333333333333
        },
        "upstream_cx_rx_bytes_total": {
          "value": 5650901,
          "perSecond": 62787.78888888889
        },
        "upstream_cx_total": {
          "value": 12,
          "perSecond": 0.13333333333333333
        },
        "upstream_cx_tx_bytes_total": {
          "value": 1619685,
          "perSecond": 17996.5
        },
        "upstream_rq_pending_overflow": {
          "value": 7,
          "perSecond": 0.07777777777777778
        },
        "upstream_rq_pending_total": {
          "value": 12,
          "perSecond": 0.13333333333333333
        },
        "upstream_rq_total": {
          "value": 35993,
          "perSecond": 399.9222222222222
        }
      }
    },
    {
      "testName": "scaling up httproutes to 50 with 10 routes per hostname at 300 rps",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-up",
      "throughput": 1212.438202247191,
      "totalRequests": 107907,
      "latency": {
        "max": 61.036543,
        "min": 0.345056,
        "mean": 0.49985100000000005,
        "pstdev": 0.832388,
        "percentiles": {
          "p50": 0.45359900000000003,
          "p75": 0.467951,
          "p80": 0.472831,
          "p90": 0.49551900000000004,
          "p95": 0.5390389999999999,
          "p99": 1.290431,
          "p999": 7.4393590000000005
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 147.07421875,
            "min": 140.265625,
            "mean": 143.76028645833333
          },
          "cpu": {
            "max": 2.1333333333333337,
            "min": 0.40000000000000036,
            "mean": 0.7977777777777771
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 33.90625,
            "min": 27.37890625,
            "mean": 32.03515625
          },
          "cpu": {
            "max": 30.00747150259066,
            "min": 30.00747150259066,
            "mean": 30.00747150259066
          }
        }
      },
      "poolOverflow": 93,
      "upstreamConnections": 34,
      "counters": {
        "benchmark.http_2xx": {
          "value": 107907,
          "perSecond": 1212.438202247191
        },
        "benchmark.pool_overflow": {
          "value": 93,
          "perSecond": 1.0449438202247192
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
          "value": 34,
          "perSecond": 0.38202247191011235
        },
        "upstream_cx_rx_bytes_total": {
          "value": 16941399,
          "perSecond": 190352.79775280898
        },
        "upstream_cx_total": {
          "value": 34,
          "perSecond": 0.38202247191011235
        },
        "upstream_cx_tx_bytes_total": {
          "value": 4855815,
          "perSecond": 54559.7191011236
        },
        "upstream_rq_pending_overflow": {
          "value": 93,
          "perSecond": 1.0449438202247192
        },
        "upstream_rq_pending_total": {
          "value": 34,
          "perSecond": 0.38202247191011235
        },
        "upstream_rq_total": {
          "value": 107907,
          "perSecond": 1212.438202247191
        }
      }
    },
    {
      "testName": "scaling up httproutes to 100 with 20 routes per hostname at 500 rps",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-up",
      "throughput": 1997.611111111111,
      "totalRequests": 179785,
      "latency": {
        "max": 90.980351,
        "min": 0.3476,
        "mean": 0.521672,
        "pstdev": 0.9244789999999999,
        "percentiles": {
          "p50": 0.452687,
          "p75": 0.474287,
          "p80": 0.48307100000000003,
          "p90": 0.5064470000000001,
          "p95": 0.586111,
          "p99": 1.869439,
          "p999": 10.968063
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 147.62890625,
            "min": 138.10546875,
            "mean": 145.30078125
          },
          "cpu": {
            "max": 2.7999999999999994,
            "min": 0.5333333333333339,
            "mean": 0.9888888888888888
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 40.01171875,
            "min": 37.4765625,
            "mean": 38.424869791666666
          },
          "cpu": {
            "max": 49.459333478544984,
            "min": 34.124251895301164,
            "mean": 46.837576796026426
          }
        }
      },
      "poolOverflow": 213,
      "upstreamConnections": 59,
      "counters": {
        "benchmark.http_2xx": {
          "value": 179785,
          "perSecond": 1997.611111111111
        },
        "benchmark.pool_overflow": {
          "value": 213,
          "perSecond": 2.3666666666666667
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
          "value": 59,
          "perSecond": 0.6555555555555556
        },
        "upstream_cx_rx_bytes_total": {
          "value": 28226245,
          "perSecond": 313624.94444444444
        },
        "upstream_cx_total": {
          "value": 59,
          "perSecond": 0.6555555555555556
        },
        "upstream_cx_tx_bytes_total": {
          "value": 8090415,
          "perSecond": 89893.5
        },
        "upstream_rq_pending_overflow": {
          "value": 213,
          "perSecond": 2.3666666666666667
        },
        "upstream_rq_pending_total": {
          "value": 59,
          "perSecond": 0.6555555555555556
        },
        "upstream_rq_total": {
          "value": 179787,
          "perSecond": 1997.6333333333334
        }
      }
    },
    {
      "testName": "scaling up httproutes to 300 with 60 routes per hostname at 800 rps",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-up",
      "throughput": 3198.7555555555555,
      "totalRequests": 287888,
      "latency": {
        "max": 151.90425499999998,
        "min": 0.326016,
        "mean": 1.168468,
        "pstdev": 5.222306000000001,
        "percentiles": {
          "p50": 0.531871,
          "p75": 0.633407,
          "p80": 0.693887,
          "p90": 0.9335669999999999,
          "p95": 1.418751,
          "p99": 21.561343,
          "p999": 76.386303
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 159.078125,
            "min": 144.9375,
            "mean": 153.85286458333334
          },
          "cpu": {
            "max": 5,
            "min": 0.6666666666666643,
            "mean": 1.5444444444444447
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 72.27734375,
            "min": 59.99609375,
            "mean": 66.97721354166667
          },
          "cpu": {
            "max": 77.5605855855856,
            "min": 46.243881959734054,
            "mean": 65.97234530390695
          }
        }
      },
      "poolOverflow": 112,
      "upstreamConnections": 240,
      "counters": {
        "benchmark.http_2xx": {
          "value": 287888,
          "perSecond": 3198.7555555555555
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
          "value": 240,
          "perSecond": 2.6666666666666665
        },
        "upstream_cx_rx_bytes_total": {
          "value": 45198416,
          "perSecond": 502204.6222222222
        },
        "upstream_cx_total": {
          "value": 240,
          "perSecond": 2.6666666666666665
        },
        "upstream_cx_tx_bytes_total": {
          "value": 12954960,
          "perSecond": 143944
        },
        "upstream_rq_pending_overflow": {
          "value": 112,
          "perSecond": 1.2444444444444445
        },
        "upstream_rq_pending_total": {
          "value": 240,
          "perSecond": 2.6666666666666665
        },
        "upstream_rq_total": {
          "value": 287888,
          "perSecond": 3198.7555555555555
        }
      }
    },
    {
      "testName": "scaling up httproutes to 500 with 100 routes per hostname at 1000 rps",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-up",
      "throughput": 3998.777777777778,
      "totalRequests": 359890,
      "latency": {
        "max": 202.014719,
        "min": 0.329824,
        "mean": 3.292303,
        "pstdev": 13.431645000000001,
        "percentiles": {
          "p50": 0.566399,
          "p75": 0.771711,
          "p80": 0.860031,
          "p90": 1.660223,
          "p95": 8.354559,
          "p99": 81.367039,
          "p999": 141.672447
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 164.703125,
            "min": 157.4453125,
            "mean": 161.54361979166666
          },
          "cpu": {
            "max": 7.666666666666658,
            "min": 0.933333333333337,
            "mean": 2.0888889022222217
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 98.875,
            "min": 89.01171875,
            "mean": 94.38020833333333
          },
          "cpu": {
            "max": 95.24843349258227,
            "min": 92.57570059190199,
            "mean": 93.24388381707206
          }
        }
      },
      "poolOverflow": 107,
      "upstreamConnections": 293,
      "counters": {
        "benchmark.http_2xx": {
          "value": 359890,
          "perSecond": 3998.777777777778
        },
        "benchmark.pool_overflow": {
          "value": 107,
          "perSecond": 1.1888888888888889
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
          "value": 293,
          "perSecond": 3.2555555555555555
        },
        "upstream_cx_rx_bytes_total": {
          "value": 56502730,
          "perSecond": 627808.1111111111
        },
        "upstream_cx_total": {
          "value": 293,
          "perSecond": 3.2555555555555555
        },
        "upstream_cx_tx_bytes_total": {
          "value": 16195185,
          "perSecond": 179946.5
        },
        "upstream_rq_pending_overflow": {
          "value": 107,
          "perSecond": 1.1888888888888889
        },
        "upstream_rq_pending_total": {
          "value": 293,
          "perSecond": 3.2555555555555555
        },
        "upstream_rq_total": {
          "value": 359893,
          "perSecond": 3998.811111111111
        }
      }
    },
    {
      "testName": "scaling up httproutes to 1000 with 200 routes per hostname at 2000 rps",
      "routes": 1000,
      "routesPerHostname": 200,
      "phase": "scaling-up",
      "throughput": 4983.211111111111,
      "totalRequests": 448489,
      "latency": {
        "max": 216.915967,
        "min": 0.41708799999999996,
        "mean": 46.134673,
        "pstdev": 35.280172,
        "percentiles": {
          "p50": 33.787903,
          "p75": 76.439551,
          "p80": 80.596991,
          "p90": 92.073983,
          "p95": 100.651007,
          "p99": 120.58623899999999,
          "p999": 180.092927
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 196.02734375,
            "min": 184.22265625,
            "mean": 190.77174479166666
          },
          "cpu": {
            "max": 14.133333333333365,
            "min": 0.7330889703432152,
            "mean": 2.902090894976589
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 155.6484375,
            "min": 149.3828125,
            "mean": 154.801953125
          },
          "cpu": {
            "max": 100.0690987454784,
            "min": 99.85133052721676,
            "mean": 99.94811640199971
          }
        }
      },
      "poolOverflow": 161,
      "upstreamConnections": 239,
      "counters": {
        "benchmark.http_2xx": {
          "value": 448489,
          "perSecond": 4983.211111111111
        },
        "benchmark.pool_overflow": {
          "value": 161,
          "perSecond": 1.788888888888889
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
          "value": 239,
          "perSecond": 2.6555555555555554
        },
        "upstream_cx_rx_bytes_total": {
          "value": 70412773,
          "perSecond": 782364.1444444444
        },
        "upstream_cx_total": {
          "value": 239,
          "perSecond": 2.6555555555555554
        },
        "upstream_cx_tx_bytes_total": {
          "value": 20186415,
          "perSecond": 224293.5
        },
        "upstream_rq_pending_overflow": {
          "value": 161,
          "perSecond": 1.788888888888889
        },
        "upstream_rq_pending_total": {
          "value": 239,
          "perSecond": 2.6555555555555554
        },
        "upstream_rq_total": {
          "value": 448587,
          "perSecond": 4984.3
        }
      }
    },
    {
      "testName": "scaling down httproutes to 500 with 100 routes per hostname at 1000 rps",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-down",
      "throughput": 3997.8555555555554,
      "totalRequests": 359807,
      "latency": {
        "max": 166.436863,
        "min": 0.32673599999999997,
        "mean": 2.4847099999999998,
        "pstdev": 10.041276,
        "percentiles": {
          "p50": 0.543615,
          "p75": 0.7288629999999999,
          "p80": 0.817535,
          "p90": 1.6024310000000002,
          "p95": 5.248511,
          "p99": 65.257471,
          "p999": 98.025471
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 191.90234375,
            "min": 172.84375,
            "mean": 178.55078125
          },
          "cpu": {
            "max": 3.1333333333333253,
            "min": 0.9999999999999432,
            "mean": 1.520000000000001
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 154.890625,
            "min": 153.66015625,
            "mean": 154.521484375
          },
          "cpu": {
            "max": 96.44166481933075,
            "min": 52.320209673316576,
            "mean": 88.50193902863825
          }
        }
      },
      "poolOverflow": 193,
      "upstreamConnections": 207,
      "counters": {
        "benchmark.http_2xx": {
          "value": 359807,
          "perSecond": 3997.8555555555554
        },
        "benchmark.pool_overflow": {
          "value": 193,
          "perSecond": 2.1444444444444444
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
          "value": 207,
          "perSecond": 2.3
        },
        "upstream_cx_rx_bytes_total": {
          "value": 56489699,
          "perSecond": 627663.3222222222
        },
        "upstream_cx_total": {
          "value": 207,
          "perSecond": 2.3
        },
        "upstream_cx_tx_bytes_total": {
          "value": 16191315,
          "perSecond": 179903.5
        },
        "upstream_rq_pending_overflow": {
          "value": 193,
          "perSecond": 2.1444444444444444
        },
        "upstream_rq_pending_total": {
          "value": 207,
          "perSecond": 2.3
        },
        "upstream_rq_total": {
          "value": 359807,
          "perSecond": 3997.8555555555554
        }
      }
    },
    {
      "testName": "scaling down httproutes to 300 with 60 routes per hostname at 800 rps",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-down",
      "throughput": 3198.222222222222,
      "totalRequests": 287840,
      "latency": {
        "max": 110.284799,
        "min": 0.321408,
        "mean": 1.079807,
        "pstdev": 4.7962739999999995,
        "percentiles": {
          "p50": 0.512159,
          "p75": 0.591199,
          "p80": 0.6246389999999999,
          "p90": 0.794207,
          "p95": 1.124415,
          "p99": 21.080063,
          "p999": 69.353471
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 171.6640625,
            "min": 155.78515625,
            "mean": 160.75598958333333
          },
          "cpu": {
            "max": 1.1999999999999509,
            "min": 1.066666666666644,
            "mean": 1.0956521739130352
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 154.83984375,
            "min": 148.24609375,
            "mean": 149.51236979166666
          },
          "cpu": {
            "max": 77.58169163932948,
            "min": 42.21342170869475,
            "mean": 61.567339702727715
          }
        }
      },
      "poolOverflow": 158,
      "upstreamConnections": 209,
      "counters": {
        "benchmark.http_2xx": {
          "value": 287840,
          "perSecond": 3198.222222222222
        },
        "benchmark.pool_overflow": {
          "value": 158,
          "perSecond": 1.7555555555555555
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
          "value": 209,
          "perSecond": 2.3222222222222224
        },
        "upstream_cx_rx_bytes_total": {
          "value": 45190880,
          "perSecond": 502120.8888888889
        },
        "upstream_cx_total": {
          "value": 209,
          "perSecond": 2.3222222222222224
        },
        "upstream_cx_tx_bytes_total": {
          "value": 12952845,
          "perSecond": 143920.5
        },
        "upstream_rq_pending_overflow": {
          "value": 158,
          "perSecond": 1.7555555555555555
        },
        "upstream_rq_pending_total": {
          "value": 209,
          "perSecond": 2.3222222222222224
        },
        "upstream_rq_total": {
          "value": 287841,
          "perSecond": 3198.233333333333
        }
      }
    },
    {
      "testName": "scaling down httproutes to 100 with 20 routes per hostname at 500 rps",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-down",
      "throughput": 1996.7222222222222,
      "totalRequests": 179705,
      "latency": {
        "max": 111.64876699999999,
        "min": 0.353088,
        "mean": 0.518911,
        "pstdev": 1.016197,
        "percentiles": {
          "p50": 0.456271,
          "p75": 0.469279,
          "p80": 0.473775,
          "p90": 0.501503,
          "p95": 0.581151,
          "p99": 1.7482229999999999,
          "p999": 8.471039
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 158.65234375,
            "min": 149.59765625,
            "mean": 153.23697916666666
          },
          "cpu": {
            "max": 1.1333333333333446,
            "min": 1.066666666666644,
            "mean": 1.0833333333333186
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 148.0859375,
            "min": 147.0859375,
            "mean": 147.59361979166667
          },
          "cpu": {
            "max": 50.94044791886718,
            "min": 50.70775568181835,
            "mean": 50.8628838398509
          }
        }
      },
      "poolOverflow": 295,
      "upstreamConnections": 59,
      "counters": {
        "benchmark.http_2xx": {
          "value": 179705,
          "perSecond": 1996.7222222222222
        },
        "benchmark.pool_overflow": {
          "value": 295,
          "perSecond": 3.2777777777777777
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
          "value": 59,
          "perSecond": 0.6555555555555556
        },
        "upstream_cx_rx_bytes_total": {
          "value": 28213685,
          "perSecond": 313485.3888888889
        },
        "upstream_cx_total": {
          "value": 59,
          "perSecond": 0.6555555555555556
        },
        "upstream_cx_tx_bytes_total": {
          "value": 8086725,
          "perSecond": 89852.5
        },
        "upstream_rq_pending_overflow": {
          "value": 295,
          "perSecond": 3.2777777777777777
        },
        "upstream_rq_pending_total": {
          "value": 59,
          "perSecond": 0.6555555555555556
        },
        "upstream_rq_total": {
          "value": 179705,
          "perSecond": 1996.7222222222222
        }
      }
    },
    {
      "testName": "scaling down httproutes to 50 with 10 routes per hostname at 300 rps",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-down",
      "throughput": 1212.5056179775281,
      "totalRequests": 107913,
      "latency": {
        "max": 79.536127,
        "min": 0.34924799999999995,
        "mean": 0.540723,
        "pstdev": 0.816151,
        "percentiles": {
          "p50": 0.468959,
          "p75": 0.522239,
          "p80": 0.537631,
          "p90": 0.5840310000000001,
          "p95": 0.6754870000000001,
          "p99": 1.557567,
          "p999": 9.940991
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 151.6484375,
            "min": 148.71484375,
            "mean": 150.14205729166667
          },
          "cpu": {
            "max": 1.066666666666644,
            "min": 1.0000000000000377,
            "mean": 1.0515151515151426
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 147.75390625,
            "min": 146.85546875,
            "mean": 147.240625
          },
          "cpu": {
            "max": 30.839552297012645,
            "min": 20.589903930428164,
            "mean": 27.965147793636593
          }
        }
      },
      "poolOverflow": 86,
      "upstreamConnections": 40,
      "counters": {
        "benchmark.http_2xx": {
          "value": 107913,
          "perSecond": 1212.5056179775281
        },
        "benchmark.pool_overflow": {
          "value": 86,
          "perSecond": 0.9662921348314607
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
          "value": 40,
          "perSecond": 0.449438202247191
        },
        "upstream_cx_rx_bytes_total": {
          "value": 16942341,
          "perSecond": 190363.38202247192
        },
        "upstream_cx_total": {
          "value": 40,
          "perSecond": 0.449438202247191
        },
        "upstream_cx_tx_bytes_total": {
          "value": 4856085,
          "perSecond": 54562.752808988764
        },
        "upstream_rq_pending_overflow": {
          "value": 86,
          "perSecond": 0.9662921348314607
        },
        "upstream_rq_pending_total": {
          "value": 40,
          "perSecond": 0.449438202247191
        },
        "upstream_rq_total": {
          "value": 107913,
          "perSecond": 1212.5056179775281
        }
      }
    },
    {
      "testName": "scaling down httproutes to 10 with 2 routes per hostname at 100 rps",
      "routes": 10,
      "routesPerHostname": 2,
      "phase": "scaling-down",
      "throughput": 404.39325842696627,
      "totalRequests": 35991,
      "latency": {
        "max": 43.788287000000004,
        "min": 0.36227200000000004,
        "mean": 0.493254,
        "pstdev": 0.590515,
        "percentiles": {
          "p50": 0.455775,
          "p75": 0.47031900000000004,
          "p80": 0.47356699999999996,
          "p90": 0.489807,
          "p95": 0.522671,
          "p99": 1.0246389999999999,
          "p999": 8.825343
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 150.1953125,
            "min": 145.38671875,
            "mean": 147.28645833333334
          },
          "cpu": {
            "max": 1.1333333333333446,
            "min": 0.933333333333337,
            "mean": 1.0499933546663929
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 147.5703125,
            "min": 146.84375,
            "mean": 147.04609375
          },
          "cpu": {
            "max": 10.200383393028265,
            "min": 9.953618421052257,
            "mean": 10.07700090704026
          }
        }
      },
      "poolOverflow": 9,
      "upstreamConnections": 12,
      "counters": {
        "benchmark.http_2xx": {
          "value": 35991,
          "perSecond": 404.39325842696627
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
          "value": 12,
          "perSecond": 0.1348314606741573
        },
        "upstream_cx_rx_bytes_total": {
          "value": 5650587,
          "perSecond": 63489.74157303371
        },
        "upstream_cx_total": {
          "value": 12,
          "perSecond": 0.1348314606741573
        },
        "upstream_cx_tx_bytes_total": {
          "value": 1619595,
          "perSecond": 18197.696629213482
        },
        "upstream_rq_pending_overflow": {
          "value": 9,
          "perSecond": 0.10112359550561797
        },
        "upstream_rq_pending_total": {
          "value": 12,
          "perSecond": 0.1348314606741573
        },
        "upstream_rq_total": {
          "value": 35991,
          "perSecond": 404.39325842696627
        }
      }
    }
  ]
};
