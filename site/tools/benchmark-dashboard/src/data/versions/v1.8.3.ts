import { TestSuite } from '../types';

// Benchmark data extracted from release artifact for version 1.8.3
// Generated from benchmark_result.json

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.8.3",
    "runId": "1.8.3-release-2026-07-22",
    "date": "2026-07-22T18:59:16Z",
    "environment": "GitHub Release",
    "description": "Benchmark results for version 1.8.3 from release artifacts",
    "downloadUrl": "https://github.com/envoyproxy/gateway/releases/download/v1.8.3/benchmark_report.zip",
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
        "max": 51.132415,
        "min": 0.365536,
        "mean": 0.5075770000000001,
        "pstdev": 0.444223,
        "percentiles": {
          "p50": 0.472159,
          "p75": 0.49329500000000004,
          "p80": 0.500303,
          "p90": 0.526463,
          "p95": 0.575455,
          "p99": 1.173119,
          "p999": 4.875263
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 131.94921875,
            "min": 110.19921875,
            "mean": 125.26692708333333
          },
          "cpu": {
            "max": 1.0666666666666664,
            "min": 0.3999999999999996,
            "mean": 0.5866666666666666
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 21.28515625,
            "min": 7.125,
            "mean": 19.7890625
          },
          "cpu": {
            "max": 0,
            "min": 0,
            "mean": 0
          }
        }
      },
      "poolOverflow": 7,
      "upstreamConnections": 9,
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
          "value": 9,
          "perSecond": 0.1
        },
        "upstream_cx_rx_bytes_total": {
          "value": 5650901,
          "perSecond": 62787.78888888889
        },
        "upstream_cx_total": {
          "value": 9,
          "perSecond": 0.1
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
          "value": 9,
          "perSecond": 0.1
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
      "throughput": 1213.0674157303372,
      "totalRequests": 107963,
      "latency": {
        "max": 114.163711,
        "min": 0.32768,
        "mean": 0.517349,
        "pstdev": 0.606409,
        "percentiles": {
          "p50": 0.461327,
          "p75": 0.5228309999999999,
          "p80": 0.543999,
          "p90": 0.585855,
          "p95": 0.707199,
          "p99": 1.271103,
          "p999": 4.477951
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 135.23046875,
            "min": 128.62890625,
            "mean": 133.09075520833332
          },
          "cpu": {
            "max": 0.6666666666666672,
            "min": 0.39999999999999886,
            "mean": 0.5583333333333336
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 25.62890625,
            "min": 20.78515625,
            "mean": 24.804817708333335
          },
          "cpu": {
            "max": 29.86671699062718,
            "min": 29.61201171177602,
            "mean": 29.771014316353924
          }
        }
      },
      "poolOverflow": 37,
      "upstreamConnections": 28,
      "counters": {
        "benchmark.http_2xx": {
          "value": 107963,
          "perSecond": 1213.0674157303372
        },
        "benchmark.pool_overflow": {
          "value": 37,
          "perSecond": 0.4157303370786517
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
          "value": 16950191,
          "perSecond": 190451.5842696629
        },
        "upstream_cx_total": {
          "value": 28,
          "perSecond": 0.3146067415730337
        },
        "upstream_cx_tx_bytes_total": {
          "value": 4858335,
          "perSecond": 54588.03370786517
        },
        "upstream_rq_pending_overflow": {
          "value": 37,
          "perSecond": 0.4157303370786517
        },
        "upstream_rq_pending_total": {
          "value": 28,
          "perSecond": 0.3146067415730337
        },
        "upstream_rq_total": {
          "value": 107963,
          "perSecond": 1213.0674157303372
        }
      }
    },
    {
      "testName": "scaling up httproutes to 100 with 20 routes per hostname at 500 rps",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-up",
      "throughput": 1998.588888888889,
      "totalRequests": 179873,
      "latency": {
        "max": 96.337919,
        "min": 0.34664,
        "mean": 0.5038170000000001,
        "pstdev": 0.744036,
        "percentiles": {
          "p50": 0.45004700000000003,
          "p75": 0.47902300000000003,
          "p80": 0.488255,
          "p90": 0.5363829999999999,
          "p95": 0.714111,
          "p99": 1.4520309999999998,
          "p999": 4.068095
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 144.8125,
            "min": 132.7578125,
            "mean": 139.57135416666668
          },
          "cpu": {
            "max": 0.7999999999999977,
            "min": 0.4666666666666685,
            "mean": 0.6608695652173914
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 29.76171875,
            "min": 25.0078125,
            "mean": 28.669921875
          },
          "cpu": {
            "max": 50.28680540957189,
            "min": 49.74902974451187,
            "mean": 50.00520828577484
          }
        }
      },
      "poolOverflow": 127,
      "upstreamConnections": 30,
      "counters": {
        "benchmark.http_2xx": {
          "value": 179873,
          "perSecond": 1998.588888888889
        },
        "benchmark.pool_overflow": {
          "value": 127,
          "perSecond": 1.4111111111111112
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
          "value": 30,
          "perSecond": 0.3333333333333333
        },
        "upstream_cx_rx_bytes_total": {
          "value": 28240061,
          "perSecond": 313778.4555555555
        },
        "upstream_cx_total": {
          "value": 30,
          "perSecond": 0.3333333333333333
        },
        "upstream_cx_tx_bytes_total": {
          "value": 8094285,
          "perSecond": 89936.5
        },
        "upstream_rq_pending_overflow": {
          "value": 127,
          "perSecond": 1.4111111111111112
        },
        "upstream_rq_pending_total": {
          "value": 30,
          "perSecond": 0.3333333333333333
        },
        "upstream_rq_total": {
          "value": 179873,
          "perSecond": 1998.588888888889
        }
      }
    },
    {
      "testName": "scaling up httproutes to 300 with 60 routes per hostname at 800 rps",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-up",
      "throughput": 3198.777777777778,
      "totalRequests": 287890,
      "latency": {
        "max": 90.623999,
        "min": 0.321952,
        "mean": 0.879421,
        "pstdev": 2.6107489999999998,
        "percentiles": {
          "p50": 0.6053430000000001,
          "p75": 0.679487,
          "p80": 0.715039,
          "p90": 0.884159,
          "p95": 1.155903,
          "p99": 5.443071,
          "p999": 48.566271
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 155.81640625,
            "min": 139.4140625,
            "mean": 150.62825520833334
          },
          "cpu": {
            "max": 24.6,
            "min": 0.7333333333333295,
            "mean": 5.0809523809523816
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 49.01953125,
            "min": 39.38671875,
            "mean": 46.757161458333336
          },
          "cpu": {
            "max": 71.12555608516043,
            "min": 45.798966963013314,
            "mean": 66.45335458648839
          }
        }
      },
      "poolOverflow": 109,
      "upstreamConnections": 183,
      "counters": {
        "benchmark.http_2xx": {
          "value": 287890,
          "perSecond": 3198.777777777778
        },
        "benchmark.pool_overflow": {
          "value": 109,
          "perSecond": 1.211111111111111
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
          "value": 183,
          "perSecond": 2.033333333333333
        },
        "upstream_cx_rx_bytes_total": {
          "value": 45198730,
          "perSecond": 502208.1111111111
        },
        "upstream_cx_total": {
          "value": 183,
          "perSecond": 2.033333333333333
        },
        "upstream_cx_tx_bytes_total": {
          "value": 12955095,
          "perSecond": 143945.5
        },
        "upstream_rq_pending_overflow": {
          "value": 109,
          "perSecond": 1.211111111111111
        },
        "upstream_rq_pending_total": {
          "value": 183,
          "perSecond": 2.033333333333333
        },
        "upstream_rq_total": {
          "value": 287891,
          "perSecond": 3198.788888888889
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
        "max": 257.10591900000003,
        "min": 0.32776,
        "mean": 1.966448,
        "pstdev": 8.562240000000001,
        "percentiles": {
          "p50": 0.6871349999999999,
          "p75": 0.857279,
          "p80": 0.9281590000000001,
          "p90": 1.254015,
          "p95": 2.1233910000000003,
          "p99": 51.021823,
          "p999": 103.79263900000001
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 166.37109375,
            "min": 151.65625,
            "mean": 163.66588541666667
          },
          "cpu": {
            "max": 37.79999999999999,
            "min": 0.8666666666666836,
            "mean": 7.569047619047619
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 69.76953125,
            "min": 56.01171875,
            "mean": 64.37630208333333
          },
          "cpu": {
            "max": 85.0582403306762,
            "min": 54.320248558863284,
            "mean": 76.69461944531828
          }
        }
      },
      "poolOverflow": 108,
      "upstreamConnections": 292,
      "counters": {
        "benchmark.http_2xx": {
          "value": 359888,
          "perSecond": 3998.7555555555555
        },
        "benchmark.pool_overflow": {
          "value": 108,
          "perSecond": 1.2
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
          "value": 292,
          "perSecond": 3.2444444444444445
        },
        "upstream_cx_rx_bytes_total": {
          "value": 56502416,
          "perSecond": 627804.6222222223
        },
        "upstream_cx_total": {
          "value": 292,
          "perSecond": 3.2444444444444445
        },
        "upstream_cx_tx_bytes_total": {
          "value": 16195140,
          "perSecond": 179946
        },
        "upstream_rq_pending_overflow": {
          "value": 108,
          "perSecond": 1.2
        },
        "upstream_rq_pending_total": {
          "value": 292,
          "perSecond": 3.2444444444444445
        },
        "upstream_rq_total": {
          "value": 359892,
          "perSecond": 3998.8
        }
      }
    },
    {
      "testName": "scaling up httproutes to 1000 with 200 routes per hostname at 2000 rps",
      "routes": 1000,
      "routesPerHostname": 200,
      "phase": "scaling-up",
      "throughput": 4517.711111111111,
      "totalRequests": 406594,
      "latency": {
        "max": 2134.704127,
        "min": 1.891712,
        "mean": 73.911531,
        "pstdev": 29.014477,
        "percentiles": {
          "p50": 72.912895,
          "p75": 83.714047,
          "p80": 85.831679,
          "p90": 89.993215,
          "p95": 93.27820700000001,
          "p99": 115.322879,
          "p999": 180.101119
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 185.81640625,
            "min": 170.34375,
            "mean": 180.59635416666666
          },
          "cpu": {
            "max": 97.33333333333334,
            "min": 0.8666666666666363,
            "mean": 16.279999999999998
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 104.75390625,
            "min": 88.48046875,
            "mean": 100.72252604166667
          },
          "cpu": {
            "max": 99.34067522187586,
            "min": 68.32545834029196,
            "mean": 95.7042618494685
          }
        }
      },
      "poolOverflow": 64,
      "upstreamConnections": 336,
      "counters": {
        "benchmark.http_2xx": {
          "value": 406594,
          "perSecond": 4517.711111111111
        },
        "benchmark.pool_overflow": {
          "value": 64,
          "perSecond": 0.7111111111111111
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
          "value": 336,
          "perSecond": 3.7333333333333334
        },
        "upstream_cx_rx_bytes_total": {
          "value": 63835258,
          "perSecond": 709280.6444444444
        },
        "upstream_cx_total": {
          "value": 336,
          "perSecond": 3.7333333333333334
        },
        "upstream_cx_tx_bytes_total": {
          "value": 18311850,
          "perSecond": 203465
        },
        "upstream_rq_pending_overflow": {
          "value": 64,
          "perSecond": 0.7111111111111111
        },
        "upstream_rq_pending_total": {
          "value": 336,
          "perSecond": 3.7333333333333334
        },
        "upstream_rq_total": {
          "value": 406930,
          "perSecond": 4521.444444444444
        }
      }
    },
    {
      "testName": "scaling down httproutes to 500 with 100 routes per hostname at 1000 rps",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-down",
      "throughput": 3999.3,
      "totalRequests": 359937,
      "latency": {
        "max": 259.268607,
        "min": 0.322704,
        "mean": 2.020829,
        "pstdev": 9.292394,
        "percentiles": {
          "p50": 0.711231,
          "p75": 0.899839,
          "p80": 0.9801269999999999,
          "p90": 1.384511,
          "p95": 2.4189429999999996,
          "p99": 47.151103,
          "p999": 111.84947100000001
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 185.05078125,
            "min": 166.546875,
            "mean": 171.54231770833334
          },
          "cpu": {
            "max": 9.133333333333269,
            "min": 1.0000000000000377,
            "mean": 2.746666666666671
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 104.9453125,
            "min": 103.35546875,
            "mean": 104.48984375
          },
          "cpu": {
            "max": 0,
            "min": 0,
            "mean": 0
          }
        }
      },
      "poolOverflow": 60,
      "upstreamConnections": 340,
      "counters": {
        "benchmark.http_2xx": {
          "value": 359937,
          "perSecond": 3999.3
        },
        "benchmark.pool_overflow": {
          "value": 60,
          "perSecond": 0.6666666666666666
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
          "value": 340,
          "perSecond": 3.7777777777777777
        },
        "upstream_cx_rx_bytes_total": {
          "value": 56510109,
          "perSecond": 627890.1
        },
        "upstream_cx_total": {
          "value": 340,
          "perSecond": 3.7777777777777777
        },
        "upstream_cx_tx_bytes_total": {
          "value": 16197300,
          "perSecond": 179970
        },
        "upstream_rq_pending_overflow": {
          "value": 60,
          "perSecond": 0.6666666666666666
        },
        "upstream_rq_pending_total": {
          "value": 340,
          "perSecond": 3.7777777777777777
        },
        "upstream_rq_total": {
          "value": 359940,
          "perSecond": 3999.3333333333335
        }
      }
    },
    {
      "testName": "scaling down httproutes to 300 with 60 routes per hostname at 800 rps",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-down",
      "throughput": 3197.8444444444444,
      "totalRequests": 287806,
      "latency": {
        "max": 154.599423,
        "min": 0.32588799999999996,
        "mean": 0.934964,
        "pstdev": 3.035537,
        "percentiles": {
          "p50": 0.610079,
          "p75": 0.691455,
          "p80": 0.7298870000000001,
          "p90": 0.9088949999999999,
          "p95": 1.200511,
          "p99": 6.1445110000000005,
          "p999": 49.352703
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 172.7265625,
            "min": 156.8515625,
            "mean": 160.60690104166667
          },
          "cpu": {
            "max": 1.3333333333333524,
            "min": 1.066666666666644,
            "mean": 1.1826086956521698
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 104.6640625,
            "min": 103.58984375,
            "mean": 104.109375
          },
          "cpu": {
            "max": 71.65725755667467,
            "min": 47.0471566150785,
            "mean": 64.86687438895122
          }
        }
      },
      "poolOverflow": 192,
      "upstreamConnections": 200,
      "counters": {
        "benchmark.http_2xx": {
          "value": 287806,
          "perSecond": 3197.8444444444444
        },
        "benchmark.pool_overflow": {
          "value": 192,
          "perSecond": 2.1333333333333333
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
          "value": 200,
          "perSecond": 2.2222222222222223
        },
        "upstream_cx_rx_bytes_total": {
          "value": 45185542,
          "perSecond": 502061.5777777778
        },
        "upstream_cx_total": {
          "value": 200,
          "perSecond": 2.2222222222222223
        },
        "upstream_cx_tx_bytes_total": {
          "value": 12951360,
          "perSecond": 143904
        },
        "upstream_rq_pending_overflow": {
          "value": 192,
          "perSecond": 2.1333333333333333
        },
        "upstream_rq_pending_total": {
          "value": 200,
          "perSecond": 2.2222222222222223
        },
        "upstream_rq_total": {
          "value": 287808,
          "perSecond": 3197.866666666667
        }
      }
    },
    {
      "testName": "scaling down httproutes to 100 with 20 routes per hostname at 500 rps",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-down",
      "throughput": 1999.4777777777779,
      "totalRequests": 179953,
      "latency": {
        "max": 61.347839,
        "min": 0.350352,
        "mean": 0.5134029999999999,
        "pstdev": 0.723512,
        "percentiles": {
          "p50": 0.453935,
          "p75": 0.478639,
          "p80": 0.48667099999999996,
          "p90": 0.532191,
          "p95": 0.724031,
          "p99": 1.5605110000000002,
          "p999": 6.3656950000000005
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 159.91015625,
            "min": 148.28125,
            "mean": 152.69010416666666
          },
          "cpu": {
            "max": 1.1333333333333446,
            "min": 1.066666666666644,
            "mean": 1.0984126984126914
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 103.7421875,
            "min": 103.0078125,
            "mean": 103.44661458333333
          },
          "cpu": {
            "max": 48.42706517227081,
            "min": 46.78797542871779,
            "mean": 47.43269954796724
          }
        }
      },
      "poolOverflow": 47,
      "upstreamConnections": 37,
      "counters": {
        "benchmark.http_2xx": {
          "value": 179953,
          "perSecond": 1999.4777777777779
        },
        "benchmark.pool_overflow": {
          "value": 47,
          "perSecond": 0.5222222222222223
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
          "value": 28252621,
          "perSecond": 313918.0111111111
        },
        "upstream_cx_total": {
          "value": 37,
          "perSecond": 0.4111111111111111
        },
        "upstream_cx_tx_bytes_total": {
          "value": 8097885,
          "perSecond": 89976.5
        },
        "upstream_rq_pending_overflow": {
          "value": 47,
          "perSecond": 0.5222222222222223
        },
        "upstream_rq_pending_total": {
          "value": 37,
          "perSecond": 0.4111111111111111
        },
        "upstream_rq_total": {
          "value": 179953,
          "perSecond": 1999.4777777777779
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
        "max": 23.735295,
        "min": 0.37063999999999997,
        "mean": 0.49363100000000004,
        "pstdev": 0.28475799999999996,
        "percentiles": {
          "p50": 0.45791899999999996,
          "p75": 0.477263,
          "p80": 0.483407,
          "p90": 0.5164949999999999,
          "p95": 0.575231,
          "p99": 1.3491829999999998,
          "p999": 3.8708470000000004
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 154.03125,
            "min": 137.33203125,
            "mean": 144.51171875
          },
          "cpu": {
            "max": 1.2666666666666515,
            "min": 1.066666666666644,
            "mean": 1.1333333333333244
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 103.5078125,
            "min": 103.0078125,
            "mean": 103.279296875
          },
          "cpu": {
            "max": 31.498524849265436,
            "min": 31.498524849265436,
            "mean": 31.498524849265436
          }
        }
      },
      "poolOverflow": 7,
      "upstreamConnections": 22,
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
          "value": 22,
          "perSecond": 0.24719101123595505
        },
        "upstream_cx_rx_bytes_total": {
          "value": 16954901,
          "perSecond": 190504.50561797753
        },
        "upstream_cx_total": {
          "value": 22,
          "perSecond": 0.24719101123595505
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
          "value": 22,
          "perSecond": 0.24719101123595505
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
      "throughput": 404.4943820224719,
      "totalRequests": 36000,
      "latency": {
        "max": 19.070975,
        "min": 0.393632,
        "mean": 0.521772,
        "pstdev": 0.373193,
        "percentiles": {
          "p50": 0.48030300000000004,
          "p75": 0.505807,
          "p80": 0.514799,
          "p90": 0.548639,
          "p95": 0.598719,
          "p99": 1.2974709999999998,
          "p999": 5.102591
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 148.15234375,
            "min": 134.84375,
            "mean": 141.89518229166666
          },
          "cpu": {
            "max": 1.1333333333333449,
            "min": 1.0000000000000377,
            "mean": 1.066666666666667
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 103.5546875,
            "min": 103.04296875,
            "mean": 103.24778645833334
          },
          "cpu": {
            "max": 11.451342242194794,
            "min": 6.591708918838274,
            "mean": 9.783392879522161
          }
        }
      },
      "poolOverflow": 0,
      "upstreamConnections": 10,
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
          "value": 10,
          "perSecond": 0.11235955056179775
        },
        "upstream_cx_rx_bytes_total": {
          "value": 5652000,
          "perSecond": 63505.61797752809
        },
        "upstream_cx_total": {
          "value": 10,
          "perSecond": 0.11235955056179775
        },
        "upstream_cx_tx_bytes_total": {
          "value": 1620000,
          "perSecond": 18202.247191011236
        },
        "upstream_rq_pending_total": {
          "value": 10,
          "perSecond": 0.11235955056179775
        },
        "upstream_rq_total": {
          "value": 36000,
          "perSecond": 404.4943820224719
        }
      }
    }
  ]
};
