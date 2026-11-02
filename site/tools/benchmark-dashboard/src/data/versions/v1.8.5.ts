import { TestSuite } from '../types';

// Benchmark data extracted from release artifact for version 1.8.5
// Generated from benchmark_result.json

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.8.5",
    "runId": "1.8.5-release-2026-09-28",
    "date": "2026-09-28T17:28:21Z",
    "environment": "GitHub Release",
    "description": "Benchmark results for version 1.8.5 from release artifacts",
    "downloadUrl": "https://github.com/envoyproxy/gateway/releases/download/v1.8.5/benchmark_report.zip",
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
        "max": 28.309503,
        "min": 0.37888,
        "mean": 0.5223909999999999,
        "pstdev": 0.583671,
        "percentiles": {
          "p50": 0.475423,
          "p75": 0.497679,
          "p80": 0.505087,
          "p90": 0.538559,
          "p95": 0.589503,
          "p99": 1.333119,
          "p999": 8.294143
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 133.03125,
            "min": 109.53125,
            "mean": 129.85143229166667
          },
          "cpu": {
            "max": 0.5334400213376006,
            "min": 0.26661334399786735,
            "mean": 0.4277926027165601
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 21,
            "min": 7.12890625,
            "mean": 18.471174568965516
          },
          "cpu": {
            "max": 11.034468999386132,
            "min": 10.910008320456432,
            "mean": 10.962416082067348
          }
        }
      },
      "poolOverflow": 2,
      "upstreamConnections": 12,
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
          "value": 12,
          "perSecond": 0.13333333333333333
        },
        "upstream_cx_rx_bytes_total": {
          "value": 5651686,
          "perSecond": 62796.51111111111
        },
        "upstream_cx_total": {
          "value": 12,
          "perSecond": 0.13333333333333333
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
          "value": 12,
          "perSecond": 0.13333333333333333
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
      "throughput": 1199.7,
      "totalRequests": 107973,
      "latency": {
        "max": 43.857919,
        "min": 0.3524,
        "mean": 0.485019,
        "pstdev": 0.333984,
        "percentiles": {
          "p50": 0.452511,
          "p75": 0.472111,
          "p80": 0.47918299999999997,
          "p90": 0.508063,
          "p95": 0.5586869999999999,
          "p99": 1.249599,
          "p999": 3.8081270000000003
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 139.93359375,
            "min": 129.73046875,
            "mean": 137.158203125
          },
          "cpu": {
            "max": 0.6000800106680881,
            "min": 0.46666666666666706,
            "mean": 0.5565217437681158
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 24.8515625,
            "min": 20.5234375,
            "mean": 24.280078125
          },
          "cpu": {
            "max": 30.55081032412965,
            "min": 30.50368630046325,
            "mean": 30.51939430835206
          }
        }
      },
      "poolOverflow": 27,
      "upstreamConnections": 22,
      "counters": {
        "benchmark.http_2xx": {
          "value": 107973,
          "perSecond": 1199.7
        },
        "benchmark.pool_overflow": {
          "value": 27,
          "perSecond": 0.3
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
          "value": 22,
          "perSecond": 0.24444444444444444
        },
        "upstream_cx_rx_bytes_total": {
          "value": 16951761,
          "perSecond": 188352.9
        },
        "upstream_cx_total": {
          "value": 22,
          "perSecond": 0.24444444444444444
        },
        "upstream_cx_tx_bytes_total": {
          "value": 4858785,
          "perSecond": 53986.5
        },
        "upstream_rq_pending_overflow": {
          "value": 27,
          "perSecond": 0.3
        },
        "upstream_rq_pending_total": {
          "value": 22,
          "perSecond": 0.24444444444444444
        },
        "upstream_rq_total": {
          "value": 107973,
          "perSecond": 1199.7
        }
      }
    },
    {
      "testName": "scaling up httproutes to 100 with 20 routes per hostname at 500 rps",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-up",
      "throughput": 1999.3333333333333,
      "totalRequests": 179940,
      "latency": {
        "max": 55.035903,
        "min": 0.340144,
        "mean": 0.488811,
        "pstdev": 0.470499,
        "percentiles": {
          "p50": 0.443935,
          "p75": 0.468287,
          "p80": 0.47561499999999995,
          "p90": 0.510047,
          "p95": 0.6575030000000001,
          "p99": 1.371199,
          "p999": 4.205823
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 143.62890625,
            "min": 136.91796875,
            "mean": 141.77122395833334
          },
          "cpu": {
            "max": 0.6666666666666672,
            "min": 0.5999999999999991,
            "mean": 0.6363636363636365
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 29.21875,
            "min": 24.51953125,
            "mean": 28.356770833333332
          },
          "cpu": {
            "max": 47.32861251525185,
            "min": 28.185683750210693,
            "mean": 37.67328734912401
          }
        }
      },
      "poolOverflow": 60,
      "upstreamConnections": 45,
      "counters": {
        "benchmark.http_2xx": {
          "value": 179940,
          "perSecond": 1999.3333333333333
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
          "value": 45,
          "perSecond": 0.5
        },
        "upstream_cx_rx_bytes_total": {
          "value": 28250580,
          "perSecond": 313895.3333333333
        },
        "upstream_cx_total": {
          "value": 45,
          "perSecond": 0.5
        },
        "upstream_cx_tx_bytes_total": {
          "value": 8097300,
          "perSecond": 89970
        },
        "upstream_rq_pending_overflow": {
          "value": 60,
          "perSecond": 0.6666666666666666
        },
        "upstream_rq_pending_total": {
          "value": 45,
          "perSecond": 0.5
        },
        "upstream_rq_total": {
          "value": 179940,
          "perSecond": 1999.3333333333333
        }
      }
    },
    {
      "testName": "scaling up httproutes to 300 with 60 routes per hostname at 800 rps",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-up",
      "throughput": 3199.8333333333335,
      "totalRequests": 287985,
      "latency": {
        "max": 78.761983,
        "min": 0.31656,
        "mean": 0.905476,
        "pstdev": 3.280247,
        "percentiles": {
          "p50": 0.595679,
          "p75": 0.661631,
          "p80": 0.695679,
          "p90": 0.836511,
          "p95": 1.0761589999999999,
          "p99": 3.935615,
          "p999": 58.712063
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 155.9296875,
            "min": 144.59765625,
            "mean": 152.173828125
          },
          "cpu": {
            "max": 30.866666666666664,
            "min": 0.7332355685908507,
            "mean": 7.15332059523928
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 49.26953125,
            "min": 32.89453125,
            "mean": 44.54127604166667
          },
          "cpu": {
            "max": 69.45105416801813,
            "min": 45.02621132976314,
            "mean": 65.49879866762801
          }
        }
      },
      "poolOverflow": 15,
      "upstreamConnections": 244,
      "counters": {
        "benchmark.http_2xx": {
          "value": 287985,
          "perSecond": 3199.8333333333335
        },
        "benchmark.pool_overflow": {
          "value": 15,
          "perSecond": 0.16666666666666666
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
          "value": 244,
          "perSecond": 2.7111111111111112
        },
        "upstream_cx_rx_bytes_total": {
          "value": 45213645,
          "perSecond": 502373.8333333333
        },
        "upstream_cx_total": {
          "value": 244,
          "perSecond": 2.7111111111111112
        },
        "upstream_cx_tx_bytes_total": {
          "value": 12959325,
          "perSecond": 143992.5
        },
        "upstream_rq_pending_overflow": {
          "value": 15,
          "perSecond": 0.16666666666666666
        },
        "upstream_rq_pending_total": {
          "value": 244,
          "perSecond": 2.7111111111111112
        },
        "upstream_rq_total": {
          "value": 287985,
          "perSecond": 3199.8333333333335
        }
      }
    },
    {
      "testName": "scaling up httproutes to 500 with 100 routes per hostname at 1000 rps",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-up",
      "throughput": 3999.3555555555554,
      "totalRequests": 359942,
      "latency": {
        "max": 298.31987100000003,
        "min": 0.323232,
        "mean": 2.0105679999999997,
        "pstdev": 10.06255,
        "percentiles": {
          "p50": 0.633951,
          "p75": 0.800735,
          "p80": 0.860159,
          "p90": 1.144511,
          "p95": 1.985151,
          "p99": 52.416511,
          "p999": 121.61023899999999
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 163.6796875,
            "min": 159.640625,
            "mean": 161.84557291666667
          },
          "cpu": {
            "max": 54.933333333333344,
            "min": 0.86666666666666,
            "mean": 11.566666666666665
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 72.08984375,
            "min": 53.890625,
            "mean": 65.66458333333334
          },
          "cpu": {
            "max": 83.72186754422064,
            "min": 57.602040234882864,
            "mean": 81.23554295865046
          }
        }
      },
      "poolOverflow": 58,
      "upstreamConnections": 342,
      "counters": {
        "benchmark.http_2xx": {
          "value": 359942,
          "perSecond": 3999.3555555555554
        },
        "benchmark.pool_overflow": {
          "value": 58,
          "perSecond": 0.6444444444444445
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
          "value": 342,
          "perSecond": 3.8
        },
        "upstream_cx_rx_bytes_total": {
          "value": 56510894,
          "perSecond": 627898.8222222222
        },
        "upstream_cx_total": {
          "value": 342,
          "perSecond": 3.8
        },
        "upstream_cx_tx_bytes_total": {
          "value": 16197390,
          "perSecond": 179971
        },
        "upstream_rq_pending_overflow": {
          "value": 58,
          "perSecond": 0.6444444444444445
        },
        "upstream_rq_pending_total": {
          "value": 342,
          "perSecond": 3.8
        },
        "upstream_rq_total": {
          "value": 359942,
          "perSecond": 3999.3555555555554
        }
      }
    },
    {
      "testName": "scaling up httproutes to 1000 with 200 routes per hostname at 2000 rps",
      "routes": 1000,
      "routesPerHostname": 200,
      "phase": "scaling-up",
      "throughput": 4946.966666666666,
      "totalRequests": 445227,
      "latency": {
        "max": 1763.835903,
        "min": 0.384736,
        "mean": 63.26142,
        "pstdev": 22.778221000000002,
        "percentiles": {
          "p50": 62.351359,
          "p75": 71.32364700000001,
          "p80": 73.146367,
          "p90": 76.828671,
          "p95": 80.101375,
          "p99": 103.043071,
          "p999": 172.965887
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 190.4765625,
            "min": 176.97265625,
            "mean": 185.11067708333334
          },
          "cpu": {
            "max": 99.20000000000006,
            "min": 0.933333333333337,
            "mean": 16.035555555555554
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 102.48828125,
            "min": 88.5703125,
            "mean": 99.080078125
          },
          "cpu": {
            "max": 0,
            "min": 0,
            "mean": 0
          }
        }
      },
      "poolOverflow": 85,
      "upstreamConnections": 315,
      "counters": {
        "benchmark.http_2xx": {
          "value": 445227,
          "perSecond": 4946.966666666666
        },
        "benchmark.pool_overflow": {
          "value": 85,
          "perSecond": 0.9444444444444444
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
          "value": 315,
          "perSecond": 3.5
        },
        "upstream_cx_rx_bytes_total": {
          "value": 69900639,
          "perSecond": 776673.7666666667
        },
        "upstream_cx_total": {
          "value": 315,
          "perSecond": 3.5
        },
        "upstream_cx_tx_bytes_total": {
          "value": 20049345,
          "perSecond": 222770.5
        },
        "upstream_rq_pending_overflow": {
          "value": 85,
          "perSecond": 0.9444444444444444
        },
        "upstream_rq_pending_total": {
          "value": 315,
          "perSecond": 3.5
        },
        "upstream_rq_total": {
          "value": 445541,
          "perSecond": 4950.455555555555
        }
      }
    },
    {
      "testName": "scaling down httproutes to 500 with 100 routes per hostname at 1000 rps",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-down",
      "throughput": 3999.4444444444443,
      "totalRequests": 359950,
      "latency": {
        "max": 182.97651100000002,
        "min": 0.323744,
        "mean": 2.0387600000000003,
        "pstdev": 9.709877,
        "percentiles": {
          "p50": 0.6461429999999999,
          "p75": 0.811103,
          "p80": 0.868831,
          "p90": 1.167807,
          "p95": 2.0366709999999997,
          "p99": 56.735743,
          "p999": 118.243327
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 190.84765625,
            "min": 171.92578125,
            "mean": 173.98841145833333
          },
          "cpu": {
            "max": 1.2666666666666515,
            "min": 1.13333333333325,
            "mean": 1.2027741086417063
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 104.75,
            "min": 100.7265625,
            "mean": 103.78854166666666
          },
          "cpu": {
            "max": 83.5370124701514,
            "min": 48.631071809498636,
            "mean": 74.96610873031307
          }
        }
      },
      "poolOverflow": 49,
      "upstreamConnections": 351,
      "counters": {
        "benchmark.http_2xx": {
          "value": 359950,
          "perSecond": 3999.4444444444443
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
          "value": 351,
          "perSecond": 3.9
        },
        "upstream_cx_rx_bytes_total": {
          "value": 56512150,
          "perSecond": 627912.7777777778
        },
        "upstream_cx_total": {
          "value": 351,
          "perSecond": 3.9
        },
        "upstream_cx_tx_bytes_total": {
          "value": 16197795,
          "perSecond": 179975.5
        },
        "upstream_rq_pending_overflow": {
          "value": 49,
          "perSecond": 0.5444444444444444
        },
        "upstream_rq_pending_total": {
          "value": 351,
          "perSecond": 3.9
        },
        "upstream_rq_total": {
          "value": 359951,
          "perSecond": 3999.4555555555557
        }
      }
    },
    {
      "testName": "scaling down httproutes to 300 with 60 routes per hostname at 800 rps",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-down",
      "throughput": 3199.7,
      "totalRequests": 287973,
      "latency": {
        "max": 74.93222300000001,
        "min": 0.31241599999999997,
        "mean": 0.8155169999999999,
        "pstdev": 2.42428,
        "percentiles": {
          "p50": 0.593215,
          "p75": 0.653407,
          "p80": 0.685855,
          "p90": 0.816543,
          "p95": 1.0412469999999998,
          "p99": 3.234047,
          "p999": 45.158399
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 178.296875,
            "min": 157.63671875,
            "mean": 165.49388020833334
          },
          "cpu": {
            "max": 1.3333333333332575,
            "min": 1.0000000000000377,
            "mean": 1.196969696969685
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 104.0078125,
            "min": 103.1796875,
            "mean": 103.6453125
          },
          "cpu": {
            "max": 69.4224337585867,
            "min": 44.27671200761932,
            "mean": 65.03689245662095
          }
        }
      },
      "poolOverflow": 27,
      "upstreamConnections": 198,
      "counters": {
        "benchmark.http_2xx": {
          "value": 287973,
          "perSecond": 3199.7
        },
        "benchmark.pool_overflow": {
          "value": 27,
          "perSecond": 0.3
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
          "value": 198,
          "perSecond": 2.2
        },
        "upstream_cx_rx_bytes_total": {
          "value": 45211761,
          "perSecond": 502352.9
        },
        "upstream_cx_total": {
          "value": 198,
          "perSecond": 2.2
        },
        "upstream_cx_tx_bytes_total": {
          "value": 12958785,
          "perSecond": 143986.5
        },
        "upstream_rq_pending_overflow": {
          "value": 27,
          "perSecond": 0.3
        },
        "upstream_rq_pending_total": {
          "value": 198,
          "perSecond": 2.2
        },
        "upstream_rq_total": {
          "value": 287973,
          "perSecond": 3199.7
        }
      }
    },
    {
      "testName": "scaling down httproutes to 100 with 20 routes per hostname at 500 rps",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-down",
      "throughput": 1999.4555555555555,
      "totalRequests": 179951,
      "latency": {
        "max": 57.098239,
        "min": 0.34784,
        "mean": 0.499194,
        "pstdev": 0.501315,
        "percentiles": {
          "p50": 0.452671,
          "p75": 0.468847,
          "p80": 0.477055,
          "p90": 0.523135,
          "p95": 0.6975990000000001,
          "p99": 1.4784629999999999,
          "p999": 4.070143
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 163.01171875,
            "min": 150.0078125,
            "mean": 152.2109375
          },
          "cpu": {
            "max": 1.1333333333333449,
            "min": 1.066666666666644,
            "mean": 1.1083333333333472
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 103.8984375,
            "min": 102.79296875,
            "mean": 103.319921875
          },
          "cpu": {
            "max": 50.835053217353774,
            "min": 28.143773420074297,
            "mean": 43.73657018847577
          }
        }
      },
      "poolOverflow": 49,
      "upstreamConnections": 37,
      "counters": {
        "benchmark.http_2xx": {
          "value": 179951,
          "perSecond": 1999.4555555555555
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
          "value": 37,
          "perSecond": 0.4111111111111111
        },
        "upstream_cx_rx_bytes_total": {
          "value": 28252307,
          "perSecond": 313914.52222222224
        },
        "upstream_cx_total": {
          "value": 37,
          "perSecond": 0.4111111111111111
        },
        "upstream_cx_tx_bytes_total": {
          "value": 8097795,
          "perSecond": 89975.5
        },
        "upstream_rq_pending_overflow": {
          "value": 49,
          "perSecond": 0.5444444444444444
        },
        "upstream_rq_pending_total": {
          "value": 37,
          "perSecond": 0.4111111111111111
        },
        "upstream_rq_total": {
          "value": 179951,
          "perSecond": 1999.4555555555555
        }
      }
    },
    {
      "testName": "scaling down httproutes to 50 with 10 routes per hostname at 300 rps",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-down",
      "throughput": 1213.0112359550562,
      "totalRequests": 107958,
      "latency": {
        "max": 30.490623000000003,
        "min": 0.329776,
        "mean": 0.567446,
        "pstdev": 0.485993,
        "percentiles": {
          "p50": 0.5001909999999999,
          "p75": 0.6102069999999999,
          "p80": 0.634015,
          "p90": 0.6951350000000001,
          "p95": 0.7618870000000001,
          "p99": 1.396607,
          "p999": 5.311230999999999
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 153.33203125,
            "min": 149.78515625,
            "mean": 151.39908854166666
          },
          "cpu": {
            "max": 1.2666666666667465,
            "min": 1.066666666666644,
            "mean": 1.1575757575757597
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 103.09765625,
            "min": 102.80859375,
            "mean": 102.92330729166666
          },
          "cpu": {
            "max": 28.328316549912476,
            "min": 16.572242765684877,
            "mean": 26.1927209497026
          }
        }
      },
      "poolOverflow": 41,
      "upstreamConnections": 25,
      "counters": {
        "benchmark.http_2xx": {
          "value": 107958,
          "perSecond": 1213.0112359550562
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
          "value": 25,
          "perSecond": 0.2808988764044944
        },
        "upstream_cx_rx_bytes_total": {
          "value": 16949406,
          "perSecond": 190442.76404494382
        },
        "upstream_cx_total": {
          "value": 25,
          "perSecond": 0.2808988764044944
        },
        "upstream_cx_tx_bytes_total": {
          "value": 4858110,
          "perSecond": 54585.50561797753
        },
        "upstream_rq_pending_overflow": {
          "value": 41,
          "perSecond": 0.4606741573033708
        },
        "upstream_rq_pending_total": {
          "value": 25,
          "perSecond": 0.2808988764044944
        },
        "upstream_rq_total": {
          "value": 107958,
          "perSecond": 1213.0112359550562
        }
      }
    },
    {
      "testName": "scaling down httproutes to 10 with 2 routes per hostname at 100 rps",
      "routes": 10,
      "routesPerHostname": 2,
      "phase": "scaling-down",
      "throughput": 404.4157303370786,
      "totalRequests": 35993,
      "latency": {
        "max": 45.230078999999996,
        "min": 0.384688,
        "mean": 0.526502,
        "pstdev": 0.6780539999999999,
        "percentiles": {
          "p50": 0.474383,
          "p75": 0.49903899999999995,
          "p80": 0.508303,
          "p90": 0.543167,
          "p95": 0.591167,
          "p99": 1.2552310000000002,
          "p999": 8.875519
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 151.73828125,
            "min": 138.58203125,
            "mean": 144.96744791666666
          },
          "cpu": {
            "max": 1.1333333333333449,
            "min": 1.066666666666644,
            "mean": 1.0944444444444354
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 103.0625,
            "min": 102.6015625,
            "mean": 102.75377604166667
          },
          "cpu": {
            "max": 10.957607594936164,
            "min": 10.741062506938954,
            "mean": 10.784371524538395
          }
        }
      },
      "poolOverflow": 7,
      "upstreamConnections": 12,
      "counters": {
        "benchmark.http_2xx": {
          "value": 35993,
          "perSecond": 404.4157303370786
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
          "value": 12,
          "perSecond": 0.1348314606741573
        },
        "upstream_cx_rx_bytes_total": {
          "value": 5650901,
          "perSecond": 63493.269662921346
        },
        "upstream_cx_total": {
          "value": 12,
          "perSecond": 0.1348314606741573
        },
        "upstream_cx_tx_bytes_total": {
          "value": 1619685,
          "perSecond": 18198.70786516854
        },
        "upstream_rq_pending_overflow": {
          "value": 7,
          "perSecond": 0.07865168539325842
        },
        "upstream_rq_pending_total": {
          "value": 12,
          "perSecond": 0.1348314606741573
        },
        "upstream_rq_total": {
          "value": 35993,
          "perSecond": 404.4157303370786
        }
      }
    }
  ]
};
