import { TestSuite } from '../types';

// Benchmark data extracted from release artifact for version 1.6.4
// Generated from benchmark_result.json

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.6.4",
    "runId": "1.6.4-release-2026-02-11",
    "date": "2026-02-11T18:48:14Z",
    "environment": "GitHub Release",
    "description": "Benchmark results for version 1.6.4 from release artifacts",
    "downloadUrl": "https://github.com/envoyproxy/gateway/releases/download/v1.6.4/benchmark_report.zip",
    "testConfiguration": {
      "rps": 100,
      "connections": 100,
      "duration": 90,
      "cpuLimit": "1000m",
      "memoryLimit": "2000Mi"
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
        "max": 11.386367,
        "min": 0.35,
        "mean": 0.46965,
        "pstdev": 0.225788,
        "percentiles": {
          "p50": 0.442879,
          "p75": 0.456559,
          "p80": 0.461311,
          "p90": 0.480175,
          "p95": 0.520079,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 135.9765625,
            "min": 123.1640625,
            "mean": 132.58528645833334
          },
          "cpu": {
            "max": 0.8,
            "min": 0.33333333333333287,
            "mean": 0.5222222222222221
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 27.64453125,
            "min": 6.15234375,
            "mean": 23.934375
          },
          "cpu": {
            "max": 9.529624248827844,
            "min": 5.990168017599307,
            "mean": 8.839583181101693
          }
        }
      },
      "poolOverflow": 3,
      "upstreamConnections": 7,
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
          "value": 7,
          "perSecond": 0.07865168539325842
        },
        "upstream_cx_rx_bytes_total": {
          "value": 5651529,
          "perSecond": 63500.32584269663
        },
        "upstream_cx_total": {
          "value": 7,
          "perSecond": 0.07865168539325842
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
          "value": 7,
          "perSecond": 0.07865168539325842
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
      "throughput": 1199.4555555555555,
      "totalRequests": 107951,
      "latency": {
        "max": 62.380031,
        "min": 0.340896,
        "mean": 0.470731,
        "pstdev": 0.613115,
        "percentiles": {
          "p50": 0.43300700000000003,
          "p75": 0.447071,
          "p80": 0.451215,
          "p90": 0.46831900000000004,
          "p95": 0.5087349999999999,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 145.69921875,
            "min": 136.0625,
            "mean": 142.76380208333333
          },
          "cpu": {
            "max": 1.9333333333333333,
            "min": 0.40000000000000036,
            "mean": 0.7777777777777778
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 33.57421875,
            "min": 29.44140625,
            "mean": 32.14075520833333
          },
          "cpu": {
            "max": 28.238767669101605,
            "min": 16.703645871077857,
            "mean": 26.294669115997362
          }
        }
      },
      "poolOverflow": 49,
      "upstreamConnections": 26,
      "counters": {
        "benchmark.http_2xx": {
          "value": 107951,
          "perSecond": 1199.4555555555555
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
          "value": 26,
          "perSecond": 0.28888888888888886
        },
        "upstream_cx_rx_bytes_total": {
          "value": 16948307,
          "perSecond": 188314.52222222224
        },
        "upstream_cx_total": {
          "value": 26,
          "perSecond": 0.28888888888888886
        },
        "upstream_cx_tx_bytes_total": {
          "value": 4857795,
          "perSecond": 53975.5
        },
        "upstream_rq_pending_overflow": {
          "value": 49,
          "perSecond": 0.5444444444444444
        },
        "upstream_rq_pending_total": {
          "value": 26,
          "perSecond": 0.28888888888888886
        },
        "upstream_rq_total": {
          "value": 107951,
          "perSecond": 1199.4555555555555
        }
      }
    },
    {
      "testName": "scaling up httproutes to 100 with 20 routes per hostname at 500 rps",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-up",
      "throughput": 2021.2359550561798,
      "totalRequests": 179890,
      "latency": {
        "max": 60.870655,
        "min": 0.33552000000000004,
        "mean": 0.471221,
        "pstdev": 0.568649,
        "percentiles": {
          "p50": 0.435999,
          "p75": 0.449839,
          "p80": 0.454463,
          "p90": 0.47441500000000003,
          "p95": 0.524127,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 151.2734375,
            "min": 138.15234375,
            "mean": 144.94466145833334
          },
          "cpu": {
            "max": 2.4666666666666672,
            "min": 0.4666666666666685,
            "mean": 0.9777777777777779
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 37.6328125,
            "min": 37.53515625,
            "mean": 37.585286458333336
          },
          "cpu": {
            "max": 46.866473690981564,
            "min": 46.622352764535854,
            "mean": 46.82578686990728
          }
        }
      },
      "poolOverflow": 110,
      "upstreamConnections": 36,
      "counters": {
        "benchmark.http_2xx": {
          "value": 179890,
          "perSecond": 2021.2359550561798
        },
        "benchmark.pool_overflow": {
          "value": 110,
          "perSecond": 1.2359550561797752
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
          "value": 28242730,
          "perSecond": 317334.0449438202
        },
        "upstream_cx_total": {
          "value": 36,
          "perSecond": 0.4044943820224719
        },
        "upstream_cx_tx_bytes_total": {
          "value": 8095050,
          "perSecond": 90955.61797752809
        },
        "upstream_rq_pending_overflow": {
          "value": 110,
          "perSecond": 1.2359550561797752
        },
        "upstream_rq_pending_total": {
          "value": 36,
          "perSecond": 0.4044943820224719
        },
        "upstream_rq_total": {
          "value": 179890,
          "perSecond": 2021.2359550561798
        }
      }
    },
    {
      "testName": "scaling up httproutes to 300 with 60 routes per hostname at 800 rps",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-up",
      "throughput": 3199.3444444444444,
      "totalRequests": 287941,
      "latency": {
        "max": 148.733951,
        "min": 0.31295999999999996,
        "mean": 0.97614,
        "pstdev": 4.529151000000001,
        "percentiles": {
          "p50": 0.47313500000000003,
          "p75": 0.554079,
          "p80": 0.591647,
          "p90": 0.764671,
          "p95": 1.004255,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 152.23046875,
            "min": 145.31640625,
            "mean": 150.138671875
          },
          "cpu": {
            "max": 4.666666666666662,
            "min": 0.6666666666666643,
            "mean": 1.455555555555555
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 71.12109375,
            "min": 61.76171875,
            "mean": 66.30546875
          },
          "cpu": {
            "max": 74.6521586726646,
            "min": 73.32263545627377,
            "mean": 73.88321261511852
          }
        }
      },
      "poolOverflow": 59,
      "upstreamConnections": 256,
      "counters": {
        "benchmark.http_2xx": {
          "value": 287941,
          "perSecond": 3199.3444444444444
        },
        "benchmark.pool_overflow": {
          "value": 59,
          "perSecond": 0.6555555555555556
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
          "value": 256,
          "perSecond": 2.8444444444444446
        },
        "upstream_cx_rx_bytes_total": {
          "value": 45206737,
          "perSecond": 502297.0777777778
        },
        "upstream_cx_total": {
          "value": 256,
          "perSecond": 2.8444444444444446
        },
        "upstream_cx_tx_bytes_total": {
          "value": 12957345,
          "perSecond": 143970.5
        },
        "upstream_rq_pending_overflow": {
          "value": 59,
          "perSecond": 0.6555555555555556
        },
        "upstream_rq_pending_total": {
          "value": 256,
          "perSecond": 2.8444444444444446
        },
        "upstream_rq_total": {
          "value": 287941,
          "perSecond": 3199.3444444444444
        }
      }
    },
    {
      "testName": "scaling up httproutes to 500 with 100 routes per hostname at 1000 rps",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-up",
      "throughput": 3993.5444444444443,
      "totalRequests": 359419,
      "latency": {
        "max": 126.963711,
        "min": 0.326976,
        "mean": 1.743744,
        "pstdev": 7.812901000000001,
        "percentiles": {
          "p50": 0.5075029999999999,
          "p75": 0.622399,
          "p80": 0.701055,
          "p90": 0.975903,
          "p95": 2.395647,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 162.53125,
            "min": 160.515625,
            "mean": 161.974609375
          },
          "cpu": {
            "max": 6.933333333333351,
            "min": 0.933333333333337,
            "mean": 1.9777777777777799
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 95.48046875,
            "min": 90.8203125,
            "mean": 93.844140625
          },
          "cpu": {
            "max": 91.90152781338112,
            "min": 51.79498024497793,
            "mean": 71.5604851909052
          }
        }
      },
      "poolOverflow": 224,
      "upstreamConnections": 176,
      "counters": {
        "benchmark.http_2xx": {
          "value": 359419,
          "perSecond": 3993.5444444444443
        },
        "benchmark.pool_overflow": {
          "value": 224,
          "perSecond": 2.488888888888889
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
          "value": 176,
          "perSecond": 1.9555555555555555
        },
        "upstream_cx_rx_bytes_total": {
          "value": 56428783,
          "perSecond": 626986.4777777778
        },
        "upstream_cx_total": {
          "value": 176,
          "perSecond": 1.9555555555555555
        },
        "upstream_cx_tx_bytes_total": {
          "value": 16175970,
          "perSecond": 179733
        },
        "upstream_rq_pending_overflow": {
          "value": 224,
          "perSecond": 2.488888888888889
        },
        "upstream_rq_pending_total": {
          "value": 176,
          "perSecond": 1.9555555555555555
        },
        "upstream_rq_total": {
          "value": 359466,
          "perSecond": 3994.0666666666666
        }
      }
    },
    {
      "testName": "scaling up httproutes to 1000 with 200 routes per hostname at 2000 rps",
      "routes": 1000,
      "routesPerHostname": 200,
      "phase": "scaling-up",
      "throughput": 5014.222222222223,
      "totalRequests": 451280,
      "latency": {
        "max": 276.578303,
        "min": 0.3776,
        "mean": 51.310144,
        "pstdev": 37.19974,
        "percentiles": {
          "p50": 59.969535,
          "p75": 81.895423,
          "p80": 86.302719,
          "p90": 97.402879,
          "p95": 104.501247,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 191.44140625,
            "min": 179.79296875,
            "mean": 185.4
          },
          "cpu": {
            "max": 13.333333333333334,
            "min": 0.8666666666667312,
            "mean": 3.011111111111104
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 153.73828125,
            "min": 147.21484375,
            "mean": 151.28802083333332
          },
          "cpu": {
            "max": 100.07782742565323,
            "min": 66.06429997077066,
            "mean": 91.7528157677912
          }
        }
      },
      "poolOverflow": 134,
      "upstreamConnections": 266,
      "counters": {
        "benchmark.http_2xx": {
          "value": 451280,
          "perSecond": 5014.222222222223
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
          "value": 266,
          "perSecond": 2.9555555555555557
        },
        "upstream_cx_rx_bytes_total": {
          "value": 70850960,
          "perSecond": 787232.8888888889
        },
        "upstream_cx_total": {
          "value": 266,
          "perSecond": 2.9555555555555557
        },
        "upstream_cx_tx_bytes_total": {
          "value": 20316195,
          "perSecond": 225735.5
        },
        "upstream_rq_pending_overflow": {
          "value": 134,
          "perSecond": 1.488888888888889
        },
        "upstream_rq_pending_total": {
          "value": 266,
          "perSecond": 2.9555555555555557
        },
        "upstream_rq_total": {
          "value": 451471,
          "perSecond": 5016.344444444445
        }
      }
    },
    {
      "testName": "scaling down httproutes to 500 with 100 routes per hostname at 1000 rps",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-down",
      "throughput": 3999.2555555555555,
      "totalRequests": 359933,
      "latency": {
        "max": 254.033919,
        "min": 0.267488,
        "mean": 2.765009,
        "pstdev": 12.463928000000001,
        "percentiles": {
          "p50": 0.5370550000000001,
          "p75": 0.7140789999999999,
          "p80": 0.7666869999999999,
          "p90": 1.1643510000000001,
          "p95": 3.600639,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 193.22265625,
            "min": 169.76171875,
            "mean": 174.6515625
          },
          "cpu": {
            "max": 1.1997600479904476,
            "min": 1.066666666666644,
            "mean": 1.1083277966664442
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 153.9453125,
            "min": 152.50390625,
            "mean": 153.59830729166666
          },
          "cpu": {
            "max": 90.6643189111433,
            "min": 48.332151206771016,
            "mean": 72.5650898137253
          }
        }
      },
      "poolOverflow": 67,
      "upstreamConnections": 333,
      "counters": {
        "benchmark.http_2xx": {
          "value": 359933,
          "perSecond": 3999.2555555555555
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
          "value": 56509481,
          "perSecond": 627883.1222222223
        },
        "upstream_cx_total": {
          "value": 333,
          "perSecond": 3.7
        },
        "upstream_cx_tx_bytes_total": {
          "value": 16196985,
          "perSecond": 179966.5
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
          "value": 359933,
          "perSecond": 3999.2555555555555
        }
      }
    },
    {
      "testName": "scaling down httproutes to 300 with 60 routes per hostname at 800 rps",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-down",
      "throughput": 3198.9555555555557,
      "totalRequests": 287906,
      "latency": {
        "max": 124.383231,
        "min": 0.312,
        "mean": 0.91919,
        "pstdev": 3.678308,
        "percentiles": {
          "p50": 0.509023,
          "p75": 0.601823,
          "p80": 0.643743,
          "p90": 0.834751,
          "p95": 1.167167,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 184.0546875,
            "min": 155.8671875,
            "mean": 166.537109375
          },
          "cpu": {
            "max": 1.2666666666666515,
            "min": 0.9999999999999432,
            "mean": 1.1272727272727225
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 153.81640625,
            "min": 148.70703125,
            "mean": 149.38411458333334
          },
          "cpu": {
            "max": 74.81835549167545,
            "min": 70.37882085331354,
            "mean": 72.42984265413129
          }
        }
      },
      "poolOverflow": 94,
      "upstreamConnections": 188,
      "counters": {
        "benchmark.http_2xx": {
          "value": 287906,
          "perSecond": 3198.9555555555557
        },
        "benchmark.pool_overflow": {
          "value": 94,
          "perSecond": 1.0444444444444445
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
          "value": 188,
          "perSecond": 2.088888888888889
        },
        "upstream_cx_rx_bytes_total": {
          "value": 45201242,
          "perSecond": 502236.02222222224
        },
        "upstream_cx_total": {
          "value": 188,
          "perSecond": 2.088888888888889
        },
        "upstream_cx_tx_bytes_total": {
          "value": 12955770,
          "perSecond": 143953
        },
        "upstream_rq_pending_overflow": {
          "value": 94,
          "perSecond": 1.0444444444444445
        },
        "upstream_rq_pending_total": {
          "value": 188,
          "perSecond": 2.088888888888889
        },
        "upstream_rq_total": {
          "value": 287906,
          "perSecond": 3198.9555555555557
        }
      }
    },
    {
      "testName": "scaling down httproutes to 100 with 20 routes per hostname at 500 rps",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-down",
      "throughput": 2021.112359550562,
      "totalRequests": 179879,
      "latency": {
        "max": 75.083775,
        "min": 0.332592,
        "mean": 0.471157,
        "pstdev": 0.623542,
        "percentiles": {
          "p50": 0.432367,
          "p75": 0.447295,
          "p80": 0.45182300000000003,
          "p90": 0.468383,
          "p95": 0.512367,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 163.9921875,
            "min": 151.484375,
            "mean": 153.562109375
          },
          "cpu": {
            "max": 1.1333333333333446,
            "min": 0.9999999999999432,
            "mean": 1.0416666666666703
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 149.2265625,
            "min": 148.9765625,
            "mean": 149.03190104166666
          },
          "cpu": {
            "max": 47.40828974480921,
            "min": 33.23458396978556,
            "mean": 45.411667373423406
          }
        }
      },
      "poolOverflow": 121,
      "upstreamConnections": 38,
      "counters": {
        "benchmark.http_2xx": {
          "value": 179879,
          "perSecond": 2021.112359550562
        },
        "benchmark.pool_overflow": {
          "value": 121,
          "perSecond": 1.3595505617977528
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
          "value": 28241003,
          "perSecond": 317314.6404494382
        },
        "upstream_cx_total": {
          "value": 38,
          "perSecond": 0.42696629213483145
        },
        "upstream_cx_tx_bytes_total": {
          "value": 8094555,
          "perSecond": 90950.05617977527
        },
        "upstream_rq_pending_overflow": {
          "value": 121,
          "perSecond": 1.3595505617977528
        },
        "upstream_rq_pending_total": {
          "value": 38,
          "perSecond": 0.42696629213483145
        },
        "upstream_rq_total": {
          "value": 179879,
          "perSecond": 2021.112359550562
        }
      }
    },
    {
      "testName": "scaling down httproutes to 50 with 10 routes per hostname at 300 rps",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-down",
      "throughput": 1212.8426966292134,
      "totalRequests": 107943,
      "latency": {
        "max": 71.524351,
        "min": 0.35416000000000003,
        "mean": 0.475783,
        "pstdev": 0.648528,
        "percentiles": {
          "p50": 0.438783,
          "p75": 0.455471,
          "p80": 0.460527,
          "p90": 0.478863,
          "p95": 0.520927,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 153.9140625,
            "min": 149.5703125,
            "mean": 150.66848958333333
          },
          "cpu": {
            "max": 1.2000000000000457,
            "min": 0.9999999999999433,
            "mean": 1.095238095238087
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 149.0078125,
            "min": 147.71484375,
            "mean": 148.062890625
          },
          "cpu": {
            "max": 29.494176952217742,
            "min": 29.494176952217742,
            "mean": 29.494176952217742
          }
        }
      },
      "poolOverflow": 57,
      "upstreamConnections": 27,
      "counters": {
        "benchmark.http_2xx": {
          "value": 107943,
          "perSecond": 1212.8426966292134
        },
        "benchmark.pool_overflow": {
          "value": 57,
          "perSecond": 0.6404494382022472
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
          "value": 16947051,
          "perSecond": 190416.3033707865
        },
        "upstream_cx_total": {
          "value": 27,
          "perSecond": 0.30337078651685395
        },
        "upstream_cx_tx_bytes_total": {
          "value": 4857435,
          "perSecond": 54577.92134831461
        },
        "upstream_rq_pending_overflow": {
          "value": 57,
          "perSecond": 0.6404494382022472
        },
        "upstream_rq_pending_total": {
          "value": 27,
          "perSecond": 0.30337078651685395
        },
        "upstream_rq_total": {
          "value": 107943,
          "perSecond": 1212.8426966292134
        }
      }
    },
    {
      "testName": "scaling down httproutes to 10 with 2 routes per hostname at 100 rps",
      "routes": 10,
      "routesPerHostname": 2,
      "phase": "scaling-down",
      "throughput": 399.9222222222222,
      "totalRequests": 35993,
      "latency": {
        "max": 57.382911,
        "min": 0.35481599999999996,
        "mean": 0.4928620000000001,
        "pstdev": 0.7162310000000001,
        "percentiles": {
          "p50": 0.450783,
          "p75": 0.470527,
          "p80": 0.479391,
          "p90": 0.512191,
          "p95": 0.562975,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 152.5078125,
            "min": 144.46484375,
            "mean": 146.828515625
          },
          "cpu": {
            "max": 1.2000000000000457,
            "min": 0.933333333333337,
            "mean": 1.0724637681159428
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 147.734375,
            "min": 147.44921875,
            "mean": 147.50325520833334
          },
          "cpu": {
            "max": 10.827605653415667,
            "min": 10.123649291472635,
            "mean": 10.546023108638455
          }
        }
      },
      "poolOverflow": 7,
      "upstreamConnections": 14,
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
          "value": 14,
          "perSecond": 0.15555555555555556
        },
        "upstream_cx_rx_bytes_total": {
          "value": 5650901,
          "perSecond": 62787.78888888889
        },
        "upstream_cx_total": {
          "value": 14,
          "perSecond": 0.15555555555555556
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
          "value": 14,
          "perSecond": 0.15555555555555556
        },
        "upstream_rq_total": {
          "value": 35993,
          "perSecond": 399.9222222222222
        }
      }
    }
  ]
};

export default benchmarkData;
