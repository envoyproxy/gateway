import { TestSuite } from '../types';

// Benchmark data extracted from release artifact for version 1.7.0
// Generated from benchmark_result.json

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.7.0",
    "runId": "1.7.0-release-2026-02-05",
    "date": "2026-02-05T22:23:47Z",
    "environment": "GitHub Release",
    "description": "Benchmark results for version 1.7.0 from release artifacts",
    "downloadUrl": "https://github.com/envoyproxy/gateway/releases/download/v1.7.0/benchmark_report.zip",
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
      "throughput": 404.4943820224719,
      "totalRequests": 36000,
      "latency": {
        "max": 20.996095,
        "min": 0.29254399999999997,
        "mean": 0.392865,
        "pstdev": 0.234214,
        "percentiles": {
          "p50": 0.376687,
          "p75": 0.391071,
          "p80": 0.396143,
          "p90": 0.41241500000000003,
          "p95": 0.435103,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 138.95703125,
            "min": 116.44921875,
            "mean": 132.88359375
          },
          "cpu": {
            "max": 1.1333333333333337,
            "min": 0.26666666666666694,
            "mean": 0.5023809523809526
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 22.7890625,
            "min": 6.25,
            "mean": 21.124088541666666
          },
          "cpu": {
            "max": 8.392692429188166,
            "min": 8.392692429188166,
            "mean": 8.392692429188166
          }
        }
      },
      "poolOverflow": 0,
      "upstreamConnections": 8,
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
          "value": 8,
          "perSecond": 0.0898876404494382
        },
        "upstream_cx_rx_bytes_total": {
          "value": 5652000,
          "perSecond": 63505.61797752809
        },
        "upstream_cx_total": {
          "value": 8,
          "perSecond": 0.0898876404494382
        },
        "upstream_cx_tx_bytes_total": {
          "value": 1620000,
          "perSecond": 18202.247191011236
        },
        "upstream_rq_pending_total": {
          "value": 8,
          "perSecond": 0.0898876404494382
        },
        "upstream_rq_total": {
          "value": 36000,
          "perSecond": 404.4943820224719
        }
      }
    },
    {
      "testName": "scaling up httproutes to 50 with 10 routes per hostname at 300 rps",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-up",
      "throughput": 1213.0449438202247,
      "totalRequests": 107961,
      "latency": {
        "max": 76.619775,
        "min": 0.260456,
        "mean": 0.38059699999999996,
        "pstdev": 0.742124,
        "percentiles": {
          "p50": 0.35684699999999997,
          "p75": 0.369695,
          "p80": 0.37345500000000004,
          "p90": 0.386127,
          "p95": 0.407487,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 141.91015625,
            "min": 132.375,
            "mean": 139.46731770833333
          },
          "cpu": {
            "max": 1.9333333333333333,
            "min": 0.33333333333333215,
            "mean": 0.7222222222222222
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 28.9453125,
            "min": 26.76953125,
            "mean": 28.652213541666665
          },
          "cpu": {
            "max": 24.348176941615478,
            "min": 16.925555708766215,
            "mean": 22.918668536759185
          }
        }
      },
      "poolOverflow": 39,
      "upstreamConnections": 34,
      "counters": {
        "benchmark.http_2xx": {
          "value": 107961,
          "perSecond": 1213.0449438202247
        },
        "benchmark.pool_overflow": {
          "value": 39,
          "perSecond": 0.43820224719101125
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
          "value": 16949877,
          "perSecond": 190448.05617977527
        },
        "upstream_cx_total": {
          "value": 34,
          "perSecond": 0.38202247191011235
        },
        "upstream_cx_tx_bytes_total": {
          "value": 4858245,
          "perSecond": 54587.02247191011
        },
        "upstream_rq_pending_overflow": {
          "value": 39,
          "perSecond": 0.43820224719101125
        },
        "upstream_rq_pending_total": {
          "value": 34,
          "perSecond": 0.38202247191011235
        },
        "upstream_rq_total": {
          "value": 107961,
          "perSecond": 1213.0449438202247
        }
      }
    },
    {
      "testName": "scaling up httproutes to 100 with 20 routes per hostname at 500 rps",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-up",
      "throughput": 2021.9775280898875,
      "totalRequests": 179956,
      "latency": {
        "max": 35.938303,
        "min": 0.24958400000000003,
        "mean": 0.360931,
        "pstdev": 0.352214,
        "percentiles": {
          "p50": 0.345583,
          "p75": 0.360431,
          "p80": 0.363695,
          "p90": 0.376079,
          "p95": 0.404991,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 147.76171875,
            "min": 140.00390625,
            "mean": 143.31015625
          },
          "cpu": {
            "max": 2.4666666666666672,
            "min": 0.46666666666666556,
            "mean": 0.862222222222222
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 35.27734375,
            "min": 32.88671875,
            "mean": 34.832942708333334
          },
          "cpu": {
            "max": 38.243800513752305,
            "min": 36.42319860496981,
            "mean": 37.312253068263566
          }
        }
      },
      "poolOverflow": 44,
      "upstreamConnections": 28,
      "counters": {
        "benchmark.http_2xx": {
          "value": 179956,
          "perSecond": 2021.9775280898875
        },
        "benchmark.pool_overflow": {
          "value": 44,
          "perSecond": 0.4943820224719101
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
          "value": 28253092,
          "perSecond": 317450.47191011236
        },
        "upstream_cx_total": {
          "value": 28,
          "perSecond": 0.3146067415730337
        },
        "upstream_cx_tx_bytes_total": {
          "value": 8098020,
          "perSecond": 90988.98876404495
        },
        "upstream_rq_pending_overflow": {
          "value": 44,
          "perSecond": 0.4943820224719101
        },
        "upstream_rq_pending_total": {
          "value": 28,
          "perSecond": 0.3146067415730337
        },
        "upstream_rq_total": {
          "value": 179956,
          "perSecond": 2021.9775280898875
        }
      }
    },
    {
      "testName": "scaling up httproutes to 300 with 60 routes per hostname at 800 rps",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-up",
      "throughput": 3199.5222222222224,
      "totalRequests": 287957,
      "latency": {
        "max": 69.099519,
        "min": 0.229736,
        "mean": 0.562037,
        "pstdev": 2.4478809999999998,
        "percentiles": {
          "p50": 0.33948700000000004,
          "p75": 0.384655,
          "p80": 0.43838299999999997,
          "p90": 0.49603099999999994,
          "p95": 0.628447,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 156.15234375,
            "min": 145.50390625,
            "mean": 152.67447916666666
          },
          "cpu": {
            "max": 4.733333333333327,
            "min": 0.7333333333333295,
            "mean": 1.3666666666666663
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 62.18359375,
            "min": 57.375,
            "mean": 60.349609375
          },
          "cpu": {
            "max": 57.686706657779716,
            "min": 56.90183246073298,
            "mean": 57.38613645993947
          }
        }
      },
      "poolOverflow": 43,
      "upstreamConnections": 177,
      "counters": {
        "benchmark.http_2xx": {
          "value": 287957,
          "perSecond": 3199.5222222222224
        },
        "benchmark.pool_overflow": {
          "value": 43,
          "perSecond": 0.4777777777777778
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
          "value": 177,
          "perSecond": 1.9666666666666666
        },
        "upstream_cx_rx_bytes_total": {
          "value": 45209249,
          "perSecond": 502324.9888888889
        },
        "upstream_cx_total": {
          "value": 177,
          "perSecond": 1.9666666666666666
        },
        "upstream_cx_tx_bytes_total": {
          "value": 12958065,
          "perSecond": 143978.5
        },
        "upstream_rq_pending_overflow": {
          "value": 43,
          "perSecond": 0.4777777777777778
        },
        "upstream_rq_pending_total": {
          "value": 177,
          "perSecond": 1.9666666666666666
        },
        "upstream_rq_total": {
          "value": 287957,
          "perSecond": 3199.5222222222224
        }
      }
    },
    {
      "testName": "scaling up httproutes to 500 with 100 routes per hostname at 1000 rps",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-up",
      "throughput": 3999.088888888889,
      "totalRequests": 359918,
      "latency": {
        "max": 153.878527,
        "min": 0.20744,
        "mean": 1.240927,
        "pstdev": 6.1772670000000005,
        "percentiles": {
          "p50": 0.444351,
          "p75": 0.504159,
          "p80": 0.537535,
          "p90": 0.7239030000000001,
          "p95": 1.070271,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 165.05078125,
            "min": 158.30078125,
            "mean": 162.92018229166666
          },
          "cpu": {
            "max": 7.266666666666666,
            "min": 0.6666666666666762,
            "mean": 1.968876744256573
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 89.16796875,
            "min": 79.9453125,
            "mean": 85.46458333333334
          },
          "cpu": {
            "max": 64.9697469866435,
            "min": 64.3614700614124,
            "mean": 64.53535528621765
          }
        }
      },
      "poolOverflow": 81,
      "upstreamConnections": 314,
      "counters": {
        "benchmark.http_2xx": {
          "value": 359918,
          "perSecond": 3999.088888888889
        },
        "benchmark.pool_overflow": {
          "value": 81,
          "perSecond": 0.9
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
          "value": 56507126,
          "perSecond": 627856.9555555555
        },
        "upstream_cx_total": {
          "value": 314,
          "perSecond": 3.488888888888889
        },
        "upstream_cx_tx_bytes_total": {
          "value": 16196355,
          "perSecond": 179959.5
        },
        "upstream_rq_pending_overflow": {
          "value": 81,
          "perSecond": 0.9
        },
        "upstream_rq_pending_total": {
          "value": 314,
          "perSecond": 3.488888888888889
        },
        "upstream_rq_total": {
          "value": 359919,
          "perSecond": 3999.1
        }
      }
    },
    {
      "testName": "scaling up httproutes to 1000 with 200 routes per hostname at 2000 rps",
      "routes": 1000,
      "routesPerHostname": 200,
      "phase": "scaling-up",
      "throughput": 7998.9,
      "totalRequests": 719901,
      "latency": {
        "max": 612.2045430000001,
        "min": 0.224096,
        "mean": 10.839886,
        "pstdev": 17.041939,
        "percentiles": {
          "p50": 2.2106869999999996,
          "p75": 16.740863,
          "p80": 24.484863,
          "p90": 33.835007,
          "p95": 39.696383000000004,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 184.7734375,
            "min": 168.8125,
            "mean": 179.18958333333333
          },
          "cpu": {
            "max": 12.533333333333301,
            "min": 0.7999999999999354,
            "mean": 2.7400000000000064
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 142.8515625,
            "min": 137.015625,
            "mean": 141.06419270833334
          },
          "cpu": {
            "max": 99.09098152526053,
            "min": 59.47936166053764,
            "mean": 90.50754321831067
          }
        }
      },
      "poolOverflow": 95,
      "upstreamConnections": 305,
      "counters": {
        "benchmark.http_2xx": {
          "value": 719901,
          "perSecond": 7998.9
        },
        "benchmark.pool_overflow": {
          "value": 95,
          "perSecond": 1.0555555555555556
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
          "value": 305,
          "perSecond": 3.388888888888889
        },
        "upstream_cx_rx_bytes_total": {
          "value": 113024457,
          "perSecond": 1255827.3
        },
        "upstream_cx_total": {
          "value": 305,
          "perSecond": 3.388888888888889
        },
        "upstream_cx_tx_bytes_total": {
          "value": 32395725,
          "perSecond": 359952.5
        },
        "upstream_rq_pending_overflow": {
          "value": 95,
          "perSecond": 1.0555555555555556
        },
        "upstream_rq_pending_total": {
          "value": 305,
          "perSecond": 3.388888888888889
        },
        "upstream_rq_total": {
          "value": 719905,
          "perSecond": 7998.944444444444
        }
      }
    },
    {
      "testName": "scaling down httproutes to 500 with 100 routes per hostname at 1000 rps",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-down",
      "throughput": 3998.9777777777776,
      "totalRequests": 359908,
      "latency": {
        "max": 96.98918300000001,
        "min": 0.21712,
        "mean": 1.214108,
        "pstdev": 5.909787,
        "percentiles": {
          "p50": 0.446463,
          "p75": 0.48827099999999996,
          "p80": 0.5066390000000001,
          "p90": 0.617279,
          "p95": 0.841503,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 180.0546875,
            "min": 168.95703125,
            "mean": 172.00481770833332
          },
          "cpu": {
            "max": 1.000000000000038,
            "min": 0.8000000000000304,
            "mean": 0.966666666666674
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 142.19921875,
            "min": 141.73828125,
            "mean": 142.04947916666666
          },
          "cpu": {
            "max": 65.55879881534545,
            "min": 38.38425750701383,
            "mean": 57.566019955372525
          }
        }
      },
      "poolOverflow": 92,
      "upstreamConnections": 308,
      "counters": {
        "benchmark.http_2xx": {
          "value": 359908,
          "perSecond": 3998.9777777777776
        },
        "benchmark.pool_overflow": {
          "value": 92,
          "perSecond": 1.0222222222222221
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
          "value": 308,
          "perSecond": 3.422222222222222
        },
        "upstream_cx_rx_bytes_total": {
          "value": 56505556,
          "perSecond": 627839.5111111111
        },
        "upstream_cx_total": {
          "value": 308,
          "perSecond": 3.422222222222222
        },
        "upstream_cx_tx_bytes_total": {
          "value": 16195860,
          "perSecond": 179954
        },
        "upstream_rq_pending_overflow": {
          "value": 92,
          "perSecond": 1.0222222222222221
        },
        "upstream_rq_pending_total": {
          "value": 308,
          "perSecond": 3.422222222222222
        },
        "upstream_rq_total": {
          "value": 359908,
          "perSecond": 3998.9777777777776
        }
      }
    },
    {
      "testName": "scaling down httproutes to 300 with 60 routes per hostname at 800 rps",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-down",
      "throughput": 3199.1111111111113,
      "totalRequests": 287920,
      "latency": {
        "max": 70.762495,
        "min": 0.240672,
        "mean": 0.617723,
        "pstdev": 2.804425,
        "percentiles": {
          "p50": 0.352303,
          "p75": 0.441039,
          "p80": 0.450911,
          "p90": 0.5074390000000001,
          "p95": 0.623903,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 171.9140625,
            "min": 154.85546875,
            "mean": 159.69309895833334
          },
          "cpu": {
            "max": 1.0000000000000377,
            "min": 0.933333333333337,
            "mean": 0.9472222222222332
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 142.1953125,
            "min": 137.5625,
            "mean": 138.707421875
          },
          "cpu": {
            "max": 60.48572130986353,
            "min": 34.594808303176436,
            "mean": 50.88592625565946
          }
        }
      },
      "poolOverflow": 80,
      "upstreamConnections": 197,
      "counters": {
        "benchmark.http_2xx": {
          "value": 287920,
          "perSecond": 3199.1111111111113
        },
        "benchmark.pool_overflow": {
          "value": 80,
          "perSecond": 0.8888888888888888
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
          "value": 197,
          "perSecond": 2.188888888888889
        },
        "upstream_cx_rx_bytes_total": {
          "value": 45203440,
          "perSecond": 502260.44444444444
        },
        "upstream_cx_total": {
          "value": 197,
          "perSecond": 2.188888888888889
        },
        "upstream_cx_tx_bytes_total": {
          "value": 12956400,
          "perSecond": 143960
        },
        "upstream_rq_pending_overflow": {
          "value": 80,
          "perSecond": 0.8888888888888888
        },
        "upstream_rq_pending_total": {
          "value": 197,
          "perSecond": 2.188888888888889
        },
        "upstream_rq_total": {
          "value": 287920,
          "perSecond": 3199.1111111111113
        }
      }
    },
    {
      "testName": "scaling down httproutes to 100 with 20 routes per hostname at 500 rps",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-down",
      "throughput": 2022,
      "totalRequests": 179958,
      "latency": {
        "max": 21.025790999999998,
        "min": 0.248144,
        "mean": 0.366898,
        "pstdev": 0.15592,
        "percentiles": {
          "p50": 0.355983,
          "p75": 0.371327,
          "p80": 0.375567,
          "p90": 0.38636699999999996,
          "p95": 0.408463,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 163.04296875,
            "min": 146.859375,
            "mean": 150.65572916666667
          },
          "cpu": {
            "max": 1.000000000000038,
            "min": 0.933333333333337,
            "mean": 0.9787878787878627
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 137.84765625,
            "min": 136.77734375,
            "mean": 137.0859375
          },
          "cpu": {
            "max": 39.49899838449119,
            "min": 39.465970330002804,
            "mean": 39.49074137086909
          }
        }
      },
      "poolOverflow": 42,
      "upstreamConnections": 19,
      "counters": {
        "benchmark.http_2xx": {
          "value": 179958,
          "perSecond": 2022
        },
        "benchmark.pool_overflow": {
          "value": 42,
          "perSecond": 0.47191011235955055
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
          "value": 28253406,
          "perSecond": 317454
        },
        "upstream_cx_total": {
          "value": 19,
          "perSecond": 0.21348314606741572
        },
        "upstream_cx_tx_bytes_total": {
          "value": 8098110,
          "perSecond": 90990
        },
        "upstream_rq_pending_overflow": {
          "value": 42,
          "perSecond": 0.47191011235955055
        },
        "upstream_rq_pending_total": {
          "value": 19,
          "perSecond": 0.21348314606741572
        },
        "upstream_rq_total": {
          "value": 179958,
          "perSecond": 2022
        }
      }
    },
    {
      "testName": "scaling down httproutes to 50 with 10 routes per hostname at 300 rps",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-down",
      "throughput": 1213.3033707865168,
      "totalRequests": 107984,
      "latency": {
        "max": 25.883647,
        "min": 0.25404,
        "mean": 0.371704,
        "pstdev": 0.24943300000000002,
        "percentiles": {
          "p50": 0.358335,
          "p75": 0.370575,
          "p80": 0.373935,
          "p90": 0.384879,
          "p95": 0.40491099999999997,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 149.109375,
            "min": 142.48046875,
            "mean": 145.869921875
          },
          "cpu": {
            "max": 1.066666666666644,
            "min": 0.933333333333337,
            "mean": 0.9749999999999859
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 136.78515625,
            "min": 136.7421875,
            "mean": 136.769140625
          },
          "cpu": {
            "max": 24.220989802525185,
            "min": 24.20874876604165,
            "mean": 24.214869284283417
          }
        }
      },
      "poolOverflow": 16,
      "upstreamConnections": 20,
      "counters": {
        "benchmark.http_2xx": {
          "value": 107984,
          "perSecond": 1213.3033707865168
        },
        "benchmark.pool_overflow": {
          "value": 16,
          "perSecond": 0.1797752808988764
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
          "value": 20,
          "perSecond": 0.2247191011235955
        },
        "upstream_cx_rx_bytes_total": {
          "value": 16953488,
          "perSecond": 190488.62921348316
        },
        "upstream_cx_total": {
          "value": 20,
          "perSecond": 0.2247191011235955
        },
        "upstream_cx_tx_bytes_total": {
          "value": 4859280,
          "perSecond": 54598.651685393255
        },
        "upstream_rq_pending_overflow": {
          "value": 16,
          "perSecond": 0.1797752808988764
        },
        "upstream_rq_pending_total": {
          "value": 20,
          "perSecond": 0.2247191011235955
        },
        "upstream_rq_total": {
          "value": 107984,
          "perSecond": 1213.3033707865168
        }
      }
    },
    {
      "testName": "scaling down httproutes to 10 with 2 routes per hostname at 100 rps",
      "routes": 10,
      "routesPerHostname": 2,
      "phase": "scaling-down",
      "throughput": 400,
      "totalRequests": 36000,
      "latency": {
        "max": 18.101247,
        "min": 0.295056,
        "mean": 0.396895,
        "pstdev": 0.25685399999999997,
        "percentiles": {
          "p50": 0.37620699999999996,
          "p75": 0.39150300000000005,
          "p80": 0.396223,
          "p90": 0.413279,
          "p95": 0.437551,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 148.10546875,
            "min": 140.08203125,
            "mean": 143.51692708333334
          },
          "cpu": {
            "max": 1.1333333333333446,
            "min": 0.8666666666666363,
            "mean": 1.0095238095238068
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 138.34765625,
            "min": 136.5234375,
            "mean": 137.123046875
          },
          "cpu": {
            "max": 8.597541519550166,
            "min": 6.052206989843309,
            "mean": 8.068733008098855
          }
        }
      },
      "poolOverflow": 0,
      "upstreamConnections": 7,
      "counters": {
        "benchmark.http_2xx": {
          "value": 36000,
          "perSecond": 400
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
          "value": 5652000,
          "perSecond": 62800
        },
        "upstream_cx_total": {
          "value": 7,
          "perSecond": 0.07777777777777778
        },
        "upstream_cx_tx_bytes_total": {
          "value": 1620000,
          "perSecond": 18000
        },
        "upstream_rq_pending_total": {
          "value": 7,
          "perSecond": 0.07777777777777778
        },
        "upstream_rq_total": {
          "value": 36000,
          "perSecond": 400
        }
      }
    }
  ]
};

export default benchmarkData;
