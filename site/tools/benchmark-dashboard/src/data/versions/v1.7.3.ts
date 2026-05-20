import { TestSuite } from '../types';

// Benchmark data extracted from release artifact for version 1.7.3
// Generated from benchmark_result.json

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.7.3",
    "runId": "1.7.3-release-2026-05-09",
    "date": "2026-05-09T23:13:52Z",
    "environment": "GitHub Release",
    "description": "Benchmark results for version 1.7.3 from release artifacts",
    "downloadUrl": "https://github.com/envoyproxy/gateway/releases/download/v1.7.3/benchmark_report.zip",
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
      "throughput": 404.44943820224717,
      "totalRequests": 35996,
      "latency": {
        "max": 46.344191,
        "min": 0.38296,
        "mean": 0.535867,
        "pstdev": 0.58933,
        "percentiles": {
          "p50": 0.47548700000000005,
          "p75": 0.496943,
          "p80": 0.5047670000000001,
          "p90": 0.541695,
          "p95": 0.627551,
          "p99": 2.049343,
          "p999": 7.107583
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 137.67578125,
            "min": 117.6484375,
            "mean": 135.65533854166668
          },
          "cpu": {
            "max": 1.4000000000000006,
            "min": 0.3333333333333336,
            "mean": 0.6252873563218392
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 24.73828125,
            "min": 24.26953125,
            "mean": 24.435807291666666
          },
          "cpu": {
            "max": 10.468080415045396,
            "min": 10.427154828411808,
            "mean": 10.433975759517407
          }
        }
      },
      "poolOverflow": 4,
      "upstreamConnections": 12,
      "counters": {
        "benchmark.http_2xx": {
          "value": 35996,
          "perSecond": 404.44943820224717
        },
        "benchmark.pool_overflow": {
          "value": 4,
          "perSecond": 0.0449438202247191
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
          "value": 5651372,
          "perSecond": 63498.56179775281
        },
        "upstream_cx_total": {
          "value": 12,
          "perSecond": 0.1348314606741573
        },
        "upstream_cx_tx_bytes_total": {
          "value": 1619820,
          "perSecond": 18200.224719101123
        },
        "upstream_rq_pending_overflow": {
          "value": 4,
          "perSecond": 0.0449438202247191
        },
        "upstream_rq_pending_total": {
          "value": 12,
          "perSecond": 0.1348314606741573
        },
        "upstream_rq_total": {
          "value": 35996,
          "perSecond": 404.44943820224717
        }
      }
    },
    {
      "testName": "scaling up httproutes to 50 with 10 routes per hostname at 300 rps",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-up",
      "throughput": 1212.7191011235955,
      "totalRequests": 107932,
      "latency": {
        "max": 56.348671,
        "min": 0.352256,
        "mean": 0.5176000000000001,
        "pstdev": 0.836032,
        "percentiles": {
          "p50": 0.458511,
          "p75": 0.47571100000000005,
          "p80": 0.482431,
          "p90": 0.511647,
          "p95": 0.579231,
          "p99": 1.619263,
          "p999": 10.062335000000001
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 147.06640625,
            "min": 137.546875,
            "mean": 144.12955729166666
          },
          "cpu": {
            "max": 2.0666666666666655,
            "min": 0.40000000000000036,
            "mean": 0.7888888888888892
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 31.00390625,
            "min": 28.80078125,
            "mean": 30.55546875
          },
          "cpu": {
            "max": 31.896429069605237,
            "min": 20.159596644847134,
            "mean": 29.888655497563743
          }
        }
      },
      "poolOverflow": 68,
      "upstreamConnections": 40,
      "counters": {
        "benchmark.http_2xx": {
          "value": 107932,
          "perSecond": 1212.7191011235955
        },
        "benchmark.pool_overflow": {
          "value": 68,
          "perSecond": 0.7640449438202247
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
          "value": 16945324,
          "perSecond": 190396.8988764045
        },
        "upstream_cx_total": {
          "value": 40,
          "perSecond": 0.449438202247191
        },
        "upstream_cx_tx_bytes_total": {
          "value": 4856940,
          "perSecond": 54572.3595505618
        },
        "upstream_rq_pending_overflow": {
          "value": 68,
          "perSecond": 0.7640449438202247
        },
        "upstream_rq_pending_total": {
          "value": 40,
          "perSecond": 0.449438202247191
        },
        "upstream_rq_total": {
          "value": 107932,
          "perSecond": 1212.7191011235955
        }
      }
    },
    {
      "testName": "scaling up httproutes to 100 with 20 routes per hostname at 500 rps",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-up",
      "throughput": 1999.2666666666667,
      "totalRequests": 179934,
      "latency": {
        "max": 982.4829430000001,
        "min": 0.33972800000000003,
        "mean": 2.467626,
        "pstdev": 30.439479,
        "percentiles": {
          "p50": 0.45207899999999995,
          "p75": 0.479439,
          "p80": 0.488863,
          "p90": 0.567295,
          "p95": 0.8977269999999999,
          "p99": 37.341183,
          "p999": 379.46982299999996
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 148.5703125,
            "min": 138.1640625,
            "mean": 145.23307291666666
          },
          "cpu": {
            "max": 2.7999999999999994,
            "min": 0.5999999999999991,
            "mean": 1.0733333333333335
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 47.19140625,
            "min": 34.671875,
            "mean": 43.625390625
          },
          "cpu": {
            "max": 48.564002664002736,
            "min": 28.511621210313688,
            "mean": 42.71705709892529
          }
        }
      },
      "poolOverflow": 66,
      "upstreamConnections": 334,
      "counters": {
        "benchmark.http_2xx": {
          "value": 179934,
          "perSecond": 1999.2666666666667
        },
        "benchmark.pool_overflow": {
          "value": 66,
          "perSecond": 0.7333333333333333
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
          "value": 334,
          "perSecond": 3.7111111111111112
        },
        "upstream_cx_rx_bytes_total": {
          "value": 28249638,
          "perSecond": 313884.86666666664
        },
        "upstream_cx_total": {
          "value": 334,
          "perSecond": 3.7111111111111112
        },
        "upstream_cx_tx_bytes_total": {
          "value": 8097030,
          "perSecond": 89967
        },
        "upstream_rq_pending_overflow": {
          "value": 66,
          "perSecond": 0.7333333333333333
        },
        "upstream_rq_pending_total": {
          "value": 334,
          "perSecond": 3.7111111111111112
        },
        "upstream_rq_total": {
          "value": 179934,
          "perSecond": 1999.2666666666667
        }
      }
    },
    {
      "testName": "scaling up httproutes to 300 with 60 routes per hostname at 800 rps",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-up",
      "throughput": 3199.4,
      "totalRequests": 287946,
      "latency": {
        "max": 3029.991423,
        "min": 0.3276,
        "mean": 4.983300000000001,
        "pstdev": 62.79289000000001,
        "percentiles": {
          "p50": 0.623775,
          "p75": 0.809791,
          "p80": 0.916447,
          "p90": 1.6968949999999998,
          "p95": 11.373567,
          "p99": 65.62815900000001,
          "p999": 194.78937499999998
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 151.1015625,
            "min": 143.8828125,
            "mean": 149.07669270833333
          },
          "cpu": {
            "max": 5.400000000000003,
            "min": 0.86666666666666,
            "mean": 1.8244444444444445
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 71.67578125,
            "min": 63.6171875,
            "mean": 68.65364583333333
          },
          "cpu": {
            "max": 77.8570545746389,
            "min": 51.385197588890605,
            "mean": 68.08387131390853
          }
        }
      },
      "poolOverflow": 52,
      "upstreamConnections": 348,
      "counters": {
        "benchmark.http_2xx": {
          "value": 287946,
          "perSecond": 3199.4
        },
        "benchmark.pool_overflow": {
          "value": 52,
          "perSecond": 0.5777777777777777
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
          "value": 348,
          "perSecond": 3.8666666666666667
        },
        "upstream_cx_rx_bytes_total": {
          "value": 45207522,
          "perSecond": 502305.8
        },
        "upstream_cx_total": {
          "value": 348,
          "perSecond": 3.8666666666666667
        },
        "upstream_cx_tx_bytes_total": {
          "value": 12957660,
          "perSecond": 143974
        },
        "upstream_rq_pending_overflow": {
          "value": 52,
          "perSecond": 0.5777777777777777
        },
        "upstream_rq_pending_total": {
          "value": 348,
          "perSecond": 3.8666666666666667
        },
        "upstream_rq_total": {
          "value": 287948,
          "perSecond": 3199.4222222222224
        }
      }
    },
    {
      "testName": "scaling up httproutes to 500 with 100 routes per hostname at 1000 rps",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-up",
      "throughput": 3997.9777777777776,
      "totalRequests": 359818,
      "latency": {
        "max": 244.654079,
        "min": 0.332464,
        "mean": 2.5921220000000003,
        "pstdev": 9.425811,
        "percentiles": {
          "p50": 0.7054389999999999,
          "p75": 0.918815,
          "p80": 1.020703,
          "p90": 1.9464949999999999,
          "p95": 6.685439,
          "p99": 62.543871,
          "p999": 90.509311
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 168.64453125,
            "min": 162.08203125,
            "mean": 166.54466145833334
          },
          "cpu": {
            "max": 8.066666666666672,
            "min": 0.9999999999999905,
            "mean": 2.2821052102695236
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 89.4921875,
            "min": 85.69140625,
            "mean": 88.01627604166667
          },
          "cpu": {
            "max": 0,
            "min": 0,
            "mean": 0
          }
        }
      },
      "poolOverflow": 179,
      "upstreamConnections": 221,
      "counters": {
        "benchmark.http_2xx": {
          "value": 359818,
          "perSecond": 3997.9777777777776
        },
        "benchmark.pool_overflow": {
          "value": 179,
          "perSecond": 1.988888888888889
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
          "value": 56491426,
          "perSecond": 627682.5111111111
        },
        "upstream_cx_total": {
          "value": 221,
          "perSecond": 2.4555555555555557
        },
        "upstream_cx_tx_bytes_total": {
          "value": 16191945,
          "perSecond": 179910.5
        },
        "upstream_rq_pending_overflow": {
          "value": 179,
          "perSecond": 1.988888888888889
        },
        "upstream_rq_pending_total": {
          "value": 221,
          "perSecond": 2.4555555555555557
        },
        "upstream_rq_total": {
          "value": 359821,
          "perSecond": 3998.011111111111
        }
      }
    },
    {
      "testName": "scaling up httproutes to 1000 with 200 routes per hostname at 2000 rps",
      "routes": 1000,
      "routesPerHostname": 200,
      "phase": "scaling-up",
      "throughput": 5138.511111111111,
      "totalRequests": 462466,
      "latency": {
        "max": 950.337535,
        "min": 1.663616,
        "mean": 49.312019,
        "pstdev": 14.934874,
        "percentiles": {
          "p50": 48.105470999999994,
          "p75": 54.161407000000004,
          "p80": 55.447551,
          "p90": 58.933247,
          "p95": 63.86278300000001,
          "p99": 102.187007,
          "p999": 155.000831
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 188.54296875,
            "min": 179.57421875,
            "mean": 186.373046875
          },
          "cpu": {
            "max": 13.933333333333357,
            "min": 1.0000000000000377,
            "mean": 3.299999999999998
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 148.27734375,
            "min": 134.01953125,
            "mean": 142.86770833333333
          },
          "cpu": {
            "max": 99.96877736845785,
            "min": 51.17226004171184,
            "mean": 86.03022160635253
          }
        }
      },
      "poolOverflow": 145,
      "upstreamConnections": 255,
      "counters": {
        "benchmark.http_2xx": {
          "value": 462466,
          "perSecond": 5138.511111111111
        },
        "benchmark.pool_overflow": {
          "value": 145,
          "perSecond": 1.6111111111111112
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
          "value": 255,
          "perSecond": 2.8333333333333335
        },
        "upstream_cx_rx_bytes_total": {
          "value": 72607162,
          "perSecond": 806746.2444444444
        },
        "upstream_cx_total": {
          "value": 255,
          "perSecond": 2.8333333333333335
        },
        "upstream_cx_tx_bytes_total": {
          "value": 20822220,
          "perSecond": 231358
        },
        "upstream_rq_pending_overflow": {
          "value": 145,
          "perSecond": 1.6111111111111112
        },
        "upstream_rq_pending_total": {
          "value": 255,
          "perSecond": 2.8333333333333335
        },
        "upstream_rq_total": {
          "value": 462716,
          "perSecond": 5141.288888888889
        }
      }
    },
    {
      "testName": "scaling down httproutes to 500 with 100 routes per hostname at 1000 rps",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-down",
      "throughput": 3998.5,
      "totalRequests": 359865,
      "latency": {
        "max": 2133.524479,
        "min": 0.33208,
        "mean": 6.480329,
        "pstdev": 35.650090000000006,
        "percentiles": {
          "p50": 0.727711,
          "p75": 1.060223,
          "p80": 1.356671,
          "p90": 9.771519,
          "p95": 44.038143000000005,
          "p99": 90.968063,
          "p999": 164.429823
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 179.8125,
            "min": 162.421875,
            "mean": 170.29466145833334
          },
          "cpu": {
            "max": 1.2666666666666515,
            "min": 0.8666666666666363,
            "mean": 1.1565217391304292
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 144.08203125,
            "min": 142.3828125,
            "mean": 143.41341145833334
          },
          "cpu": {
            "max": 65.76390265911255,
            "min": 44.97408887731358,
            "mean": 54.14164684504709
          }
        }
      },
      "poolOverflow": 132,
      "upstreamConnections": 268,
      "counters": {
        "benchmark.http_2xx": {
          "value": 359865,
          "perSecond": 3998.5
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
          "value": 268,
          "perSecond": 2.977777777777778
        },
        "upstream_cx_rx_bytes_total": {
          "value": 56498805,
          "perSecond": 627764.5
        },
        "upstream_cx_total": {
          "value": 268,
          "perSecond": 2.977777777777778
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
          "value": 268,
          "perSecond": 2.977777777777778
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
      "throughput": 3182.8333333333335,
      "totalRequests": 286455,
      "latency": {
        "max": 489.89798299999995,
        "min": 0.330016,
        "mean": 2.042146,
        "pstdev": 11.847002,
        "percentiles": {
          "p50": 0.615935,
          "p75": 0.7458229999999999,
          "p80": 0.833087,
          "p90": 1.343807,
          "p95": 2.417791,
          "p99": 46.784511,
          "p999": 224.002047
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 169.71875,
            "min": 156.6015625,
            "mean": 164.18645833333332
          },
          "cpu": {
            "max": 1.3333333333333524,
            "min": 1.066666666666644,
            "mean": 1.1650793650793705
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 142.4296875,
            "min": 141.1640625,
            "mean": 141.56002604166667
          },
          "cpu": {
            "max": 73.5007114191903,
            "min": 43.6545225355711,
            "mean": 65.09343172003922
          }
        }
      },
      "poolOverflow": 126,
      "upstreamConnections": 274,
      "counters": {
        "benchmark.http_2xx": {
          "value": 286455,
          "perSecond": 3182.8333333333335
        },
        "benchmark.pool_overflow": {
          "value": 126,
          "perSecond": 1.4
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
          "value": 274,
          "perSecond": 3.0444444444444443
        },
        "upstream_cx_rx_bytes_total": {
          "value": 44973435,
          "perSecond": 499704.8333333333
        },
        "upstream_cx_total": {
          "value": 274,
          "perSecond": 3.0444444444444443
        },
        "upstream_cx_tx_bytes_total": {
          "value": 12902805,
          "perSecond": 143364.5
        },
        "upstream_rq_pending_overflow": {
          "value": 126,
          "perSecond": 1.4
        },
        "upstream_rq_pending_total": {
          "value": 274,
          "perSecond": 3.0444444444444443
        },
        "upstream_rq_total": {
          "value": 286729,
          "perSecond": 3185.8777777777777
        }
      }
    },
    {
      "testName": "scaling down httproutes to 100 with 20 routes per hostname at 500 rps",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-down",
      "throughput": 1998.388888888889,
      "totalRequests": 179855,
      "latency": {
        "max": 147.406847,
        "min": 0.339904,
        "mean": 0.5224789999999999,
        "pstdev": 0.6352289999999999,
        "percentiles": {
          "p50": 0.453471,
          "p75": 0.47263900000000003,
          "p80": 0.478431,
          "p90": 0.517999,
          "p95": 0.724191,
          "p99": 2.087359,
          "p999": 7.226879
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 156.8125,
            "min": 147.4296875,
            "mean": 150.44973958333333
          },
          "cpu": {
            "max": 1.2666666666666517,
            "min": 1.066666666666644,
            "mean": 1.1277777777777709
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 140.5390625,
            "min": 140.19921875,
            "mean": 140.38294270833333
          },
          "cpu": {
            "max": 51.33240404359542,
            "min": 33.86808667887703,
            "mean": 48.447040965036784
          }
        }
      },
      "poolOverflow": 142,
      "upstreamConnections": 48,
      "counters": {
        "benchmark.http_2xx": {
          "value": 179855,
          "perSecond": 1998.388888888889
        },
        "benchmark.pool_overflow": {
          "value": 142,
          "perSecond": 1.5777777777777777
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
          "value": 28237235,
          "perSecond": 313747.05555555556
        },
        "upstream_cx_total": {
          "value": 48,
          "perSecond": 0.5333333333333333
        },
        "upstream_cx_tx_bytes_total": {
          "value": 8093610,
          "perSecond": 89929
        },
        "upstream_rq_pending_overflow": {
          "value": 142,
          "perSecond": 1.5777777777777777
        },
        "upstream_rq_pending_total": {
          "value": 48,
          "perSecond": 0.5333333333333333
        },
        "upstream_rq_total": {
          "value": 179858,
          "perSecond": 1998.4222222222222
        }
      }
    },
    {
      "testName": "scaling down httproutes to 50 with 10 routes per hostname at 300 rps",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-down",
      "throughput": 1198.9,
      "totalRequests": 107901,
      "latency": {
        "max": 91.013119,
        "min": 0.36744000000000004,
        "mean": 0.51687,
        "pstdev": 1.1865809999999999,
        "percentiles": {
          "p50": 0.455055,
          "p75": 0.471199,
          "p80": 0.476111,
          "p90": 0.499823,
          "p95": 0.548223,
          "p99": 1.549503,
          "p999": 8.814591
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 150.85546875,
            "min": 148.97265625,
            "mean": 149.60885416666667
          },
          "cpu": {
            "max": 1.2666666666666515,
            "min": 1.0000000000000377,
            "mean": 1.1142857142857159
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 140.7421875,
            "min": 139.89453125,
            "mean": 140.21901041666666
          },
          "cpu": {
            "max": 30.88898623279097,
            "min": 19.699947025365148,
            "mean": 27.62057691682817
          }
        }
      },
      "poolOverflow": 99,
      "upstreamConnections": 37,
      "counters": {
        "benchmark.http_2xx": {
          "value": 107901,
          "perSecond": 1198.9
        },
        "benchmark.pool_overflow": {
          "value": 99,
          "perSecond": 1.1
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
          "value": 16940457,
          "perSecond": 188227.3
        },
        "upstream_cx_total": {
          "value": 37,
          "perSecond": 0.4111111111111111
        },
        "upstream_cx_tx_bytes_total": {
          "value": 4855545,
          "perSecond": 53950.5
        },
        "upstream_rq_pending_overflow": {
          "value": 99,
          "perSecond": 1.1
        },
        "upstream_rq_pending_total": {
          "value": 37,
          "perSecond": 0.4111111111111111
        },
        "upstream_rq_total": {
          "value": 107901,
          "perSecond": 1198.9
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
        "max": 64.180223,
        "min": 0.3844,
        "mean": 0.502389,
        "pstdev": 0.691644,
        "percentiles": {
          "p50": 0.45702299999999996,
          "p75": 0.473823,
          "p80": 0.479103,
          "p90": 0.5029750000000001,
          "p95": 0.541279,
          "p99": 1.187135,
          "p999": 8.559103
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 150.90625,
            "min": 144.01953125,
            "mean": 146.19817708333332
          },
          "cpu": {
            "max": 1.2666666666666515,
            "min": 1.066666666666644,
            "mean": 1.1391304347826023
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 140.7421875,
            "min": 139.96875,
            "mean": 140.19140625
          },
          "cpu": {
            "max": 0,
            "min": 0,
            "mean": 0
          }
        }
      },
      "poolOverflow": 9,
      "upstreamConnections": 15,
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
          "value": 15,
          "perSecond": 0.16853932584269662
        },
        "upstream_cx_rx_bytes_total": {
          "value": 5650587,
          "perSecond": 63489.74157303371
        },
        "upstream_cx_total": {
          "value": 15,
          "perSecond": 0.16853932584269662
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
          "value": 15,
          "perSecond": 0.16853932584269662
        },
        "upstream_rq_total": {
          "value": 35991,
          "perSecond": 404.39325842696627
        }
      }
    }
  ]
};
