import { TestSuite } from '../types';

// Benchmark data extracted from release artifact for version 1.8.6
// Generated from benchmark_result.json

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.8.6",
    "runId": "1.8.6-release-2026-10-26",
    "date": "2026-10-26T06:47:43Z",
    "environment": "GitHub Release",
    "description": "Benchmark results for version 1.8.6 from release artifacts",
    "downloadUrl": "https://github.com/envoyproxy/gateway/releases/download/v1.8.6/benchmark_report.zip",
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
      "throughput": 404.438202247191,
      "totalRequests": 35995,
      "latency": {
        "max": 49.174527000000005,
        "min": 0.368224,
        "mean": 0.49030000000000007,
        "pstdev": 0.733786,
        "percentiles": {
          "p50": 0.444415,
          "p75": 0.45903900000000003,
          "p80": 0.462895,
          "p90": 0.481983,
          "p95": 0.522431,
          "p99": 1.188607,
          "p999": 6.849023
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 136.5859375,
            "min": 116.66796875,
            "mean": 131.84205729166666
          },
          "cpu": {
            "max": 0.9999999999999993,
            "min": 0.3334444814938308,
            "mean": 0.545234141533952
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 22.04296875,
            "min": 7.19140625,
            "mean": 20.562630208333335
          },
          "cpu": {
            "max": 10.017454793585808,
            "min": 5.965580456852793,
            "mean": 9.07078371892435
          }
        }
      },
      "poolOverflow": 5,
      "upstreamConnections": 11,
      "counters": {
        "benchmark.http_2xx": {
          "value": 35995,
          "perSecond": 404.438202247191
        },
        "benchmark.pool_overflow": {
          "value": 5,
          "perSecond": 0.056179775280898875
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
          "value": 5651215,
          "perSecond": 63496.79775280899
        },
        "upstream_cx_total": {
          "value": 11,
          "perSecond": 0.12359550561797752
        },
        "upstream_cx_tx_bytes_total": {
          "value": 1619775,
          "perSecond": 18199.719101123595
        },
        "upstream_rq_pending_overflow": {
          "value": 5,
          "perSecond": 0.056179775280898875
        },
        "upstream_rq_pending_total": {
          "value": 11,
          "perSecond": 0.12359550561797752
        },
        "upstream_rq_total": {
          "value": 35995,
          "perSecond": 404.438202247191
        }
      }
    },
    {
      "testName": "scaling up httproutes to 50 with 10 routes per hostname at 300 rps",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-up",
      "throughput": 1213.3595505617977,
      "totalRequests": 107989,
      "latency": {
        "max": 29.393919,
        "min": 0.349504,
        "mean": 0.47622000000000003,
        "pstdev": 0.312723,
        "percentiles": {
          "p50": 0.443247,
          "p75": 0.45899100000000004,
          "p80": 0.463007,
          "p90": 0.490799,
          "p95": 0.542207,
          "p99": 1.219135,
          "p999": 3.939199
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 142.23046875,
            "min": 135.0546875,
            "mean": 139.65768229166667
          },
          "cpu": {
            "max": 0.6001200240048016,
            "min": 0.39999999999999886,
            "mean": 0.5333365187302865
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 26.1875,
            "min": 21.53515625,
            "mean": 24.445442708333335
          },
          "cpu": {
            "max": 30.2450942812983,
            "min": 30.2450942812983,
            "mean": 30.2450942812983
          }
        }
      },
      "poolOverflow": 11,
      "upstreamConnections": 23,
      "counters": {
        "benchmark.http_2xx": {
          "value": 107989,
          "perSecond": 1213.3595505617977
        },
        "benchmark.pool_overflow": {
          "value": 11,
          "perSecond": 0.12359550561797752
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
          "value": 23,
          "perSecond": 0.25842696629213485
        },
        "upstream_cx_rx_bytes_total": {
          "value": 16954273,
          "perSecond": 190497.44943820225
        },
        "upstream_cx_total": {
          "value": 23,
          "perSecond": 0.25842696629213485
        },
        "upstream_cx_tx_bytes_total": {
          "value": 4859505,
          "perSecond": 54601.1797752809
        },
        "upstream_rq_pending_overflow": {
          "value": 11,
          "perSecond": 0.12359550561797752
        },
        "upstream_rq_pending_total": {
          "value": 23,
          "perSecond": 0.25842696629213485
        },
        "upstream_rq_total": {
          "value": 107989,
          "perSecond": 1213.3595505617977
        }
      }
    },
    {
      "testName": "scaling up httproutes to 100 with 20 routes per hostname at 500 rps",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-up",
      "throughput": 1999.3555555555556,
      "totalRequests": 179942,
      "latency": {
        "max": 59.346942999999996,
        "min": 0.33608,
        "mean": 0.49446199999999996,
        "pstdev": 0.463185,
        "percentiles": {
          "p50": 0.448687,
          "p75": 0.468479,
          "p80": 0.474927,
          "p90": 0.511023,
          "p95": 0.674655,
          "p99": 1.448255,
          "p999": 4.453631000000001
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 144.82421875,
            "min": 134.390625,
            "mean": 142.09765625
          },
          "cpu": {
            "max": 3.3333333333333335,
            "min": 0.5999999999999991,
            "mean": 1.1733333333333331
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 30.0234375,
            "min": 25.359375,
            "mean": 28.770182291666668
          },
          "cpu": {
            "max": 50.64683050516873,
            "min": 28.145607959738207,
            "mean": 37.31180429460918
          }
        }
      },
      "poolOverflow": 58,
      "upstreamConnections": 38,
      "counters": {
        "benchmark.http_2xx": {
          "value": 179942,
          "perSecond": 1999.3555555555556
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
          "value": 38,
          "perSecond": 0.4222222222222222
        },
        "upstream_cx_rx_bytes_total": {
          "value": 28250894,
          "perSecond": 313898.8222222222
        },
        "upstream_cx_total": {
          "value": 38,
          "perSecond": 0.4222222222222222
        },
        "upstream_cx_tx_bytes_total": {
          "value": 8097390,
          "perSecond": 89971
        },
        "upstream_rq_pending_overflow": {
          "value": 58,
          "perSecond": 0.6444444444444445
        },
        "upstream_rq_pending_total": {
          "value": 38,
          "perSecond": 0.4222222222222222
        },
        "upstream_rq_total": {
          "value": 179942,
          "perSecond": 1999.3555555555556
        }
      }
    },
    {
      "testName": "scaling up httproutes to 300 with 60 routes per hostname at 800 rps",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-up",
      "throughput": 3199.0666666666666,
      "totalRequests": 287916,
      "latency": {
        "max": 47.302655,
        "min": 0.323216,
        "mean": 0.726439,
        "pstdev": 1.12138,
        "percentiles": {
          "p50": 0.598495,
          "p75": 0.659903,
          "p80": 0.694271,
          "p90": 0.8338869999999999,
          "p95": 1.0524149999999999,
          "p99": 2.680831,
          "p999": 21.057534999999998
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 156.80859375,
            "min": 143.95703125,
            "mean": 152.86328125
          },
          "cpu": {
            "max": 34.06666666666667,
            "min": 0.6666666666666643,
            "mean": 6.900000000000002
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 47.1484375,
            "min": 39.63671875,
            "mean": 44.444661458333336
          },
          "cpu": {
            "max": 69.94817441717043,
            "min": 69.4867531003382,
            "mean": 69.6980050921912
          }
        }
      },
      "poolOverflow": 84,
      "upstreamConnections": 102,
      "counters": {
        "benchmark.http_2xx": {
          "value": 287916,
          "perSecond": 3199.0666666666666
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
          "value": 102,
          "perSecond": 1.1333333333333333
        },
        "upstream_cx_rx_bytes_total": {
          "value": 45202812,
          "perSecond": 502253.4666666667
        },
        "upstream_cx_total": {
          "value": 102,
          "perSecond": 1.1333333333333333
        },
        "upstream_cx_tx_bytes_total": {
          "value": 12956220,
          "perSecond": 143958
        },
        "upstream_rq_pending_overflow": {
          "value": 84,
          "perSecond": 0.9333333333333333
        },
        "upstream_rq_pending_total": {
          "value": 102,
          "perSecond": 1.1333333333333333
        },
        "upstream_rq_total": {
          "value": 287916,
          "perSecond": 3199.0666666666666
        }
      }
    },
    {
      "testName": "scaling up httproutes to 500 with 100 routes per hostname at 1000 rps",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-up",
      "throughput": 4044.258426966292,
      "totalRequests": 359939,
      "latency": {
        "max": 315.752447,
        "min": 0.322784,
        "mean": 1.937796,
        "pstdev": 9.359589,
        "percentiles": {
          "p50": 0.689503,
          "p75": 0.853663,
          "p80": 0.9182389999999999,
          "p90": 1.2664309999999999,
          "p95": 2.439295,
          "p99": 46.370815,
          "p999": 122.474495
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 175.1875,
            "min": 164.91796875,
            "mean": 170.701171875
          },
          "cpu": {
            "max": 59.06666666666667,
            "min": 0.933333333333337,
            "mean": 11.26666666666667
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 72.9765625,
            "min": 48.2890625,
            "mean": 63.471354166666664
          },
          "cpu": {
            "max": 84.48096554902627,
            "min": 55.77155108460453,
            "mean": 70.1262583168154
          }
        }
      },
      "poolOverflow": 58,
      "upstreamConnections": 342,
      "counters": {
        "benchmark.http_2xx": {
          "value": 359939,
          "perSecond": 4044.258426966292
        },
        "benchmark.pool_overflow": {
          "value": 58,
          "perSecond": 0.651685393258427
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
          "value": 342,
          "perSecond": 3.842696629213483
        },
        "upstream_cx_rx_bytes_total": {
          "value": 56510423,
          "perSecond": 634948.5730337079
        },
        "upstream_cx_total": {
          "value": 342,
          "perSecond": 3.842696629213483
        },
        "upstream_cx_tx_bytes_total": {
          "value": 16197345,
          "perSecond": 181992.64044943822
        },
        "upstream_rq_pending_overflow": {
          "value": 58,
          "perSecond": 0.651685393258427
        },
        "upstream_rq_pending_total": {
          "value": 342,
          "perSecond": 3.842696629213483
        },
        "upstream_rq_total": {
          "value": 359941,
          "perSecond": 4044.2808988764045
        }
      }
    },
    {
      "testName": "scaling up httproutes to 1000 with 200 routes per hostname at 2000 rps",
      "routes": 1000,
      "routesPerHostname": 200,
      "phase": "scaling-up",
      "throughput": 4833.066666666667,
      "totalRequests": 434976,
      "latency": {
        "max": 1604.321279,
        "min": 1.905152,
        "mean": 63.343183999999994,
        "pstdev": 21.100161,
        "percentiles": {
          "p50": 62.357503,
          "p75": 71.626751,
          "p80": 73.23647899999999,
          "p90": 76.69759900000001,
          "p95": 80.150527,
          "p99": 103.473151,
          "p999": 168.493055
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 192.48828125,
            "min": 170.69140625,
            "mean": 181.10442708333332
          },
          "cpu": {
            "max": 96.06666666666669,
            "min": 0.933333333333337,
            "mean": 17.273333333333323
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 105.1796875,
            "min": 89.0859375,
            "mean": 101.73567708333333
          },
          "cpu": {
            "max": 98.85747445913444,
            "min": 67.21523505474924,
            "mean": 92.80137366578352
          }
        }
      },
      "poolOverflow": 92,
      "upstreamConnections": 308,
      "counters": {
        "benchmark.http_2xx": {
          "value": 434976,
          "perSecond": 4833.066666666667
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
          "value": 68291232,
          "perSecond": 758791.4666666667
        },
        "upstream_cx_total": {
          "value": 308,
          "perSecond": 3.422222222222222
        },
        "upstream_cx_tx_bytes_total": {
          "value": 19587780,
          "perSecond": 217642
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
          "value": 435284,
          "perSecond": 4836.488888888889
        }
      }
    },
    {
      "testName": "scaling down httproutes to 500 with 100 routes per hostname at 1000 rps",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-down",
      "throughput": 3999.4777777777776,
      "totalRequests": 359953,
      "latency": {
        "max": 240.78745500000002,
        "min": 0.32576000000000005,
        "mean": 2.087761,
        "pstdev": 9.442183,
        "percentiles": {
          "p50": 0.682495,
          "p75": 0.863039,
          "p80": 0.945855,
          "p90": 1.417663,
          "p95": 2.991231,
          "p99": 50.276351,
          "p999": 120.029183
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 187.7734375,
            "min": 167.88671875,
            "mean": 172.470703125
          },
          "cpu": {
            "max": 1.2000000000000455,
            "min": 0.9999999999999432,
            "mean": 1.1594202898550767
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 105.41015625,
            "min": 103.5546875,
            "mean": 104.74674479166667
          },
          "cpu": {
            "max": 85.57953983318974,
            "min": 60.43722032185072,
            "mean": 80.55107593092194
          }
        }
      },
      "poolOverflow": 44,
      "upstreamConnections": 356,
      "counters": {
        "benchmark.http_2xx": {
          "value": 359953,
          "perSecond": 3999.4777777777776
        },
        "benchmark.pool_overflow": {
          "value": 44,
          "perSecond": 0.4888888888888889
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
          "value": 356,
          "perSecond": 3.9555555555555557
        },
        "upstream_cx_rx_bytes_total": {
          "value": 56512621,
          "perSecond": 627918.0111111111
        },
        "upstream_cx_total": {
          "value": 356,
          "perSecond": 3.9555555555555557
        },
        "upstream_cx_tx_bytes_total": {
          "value": 16198020,
          "perSecond": 179978
        },
        "upstream_rq_pending_overflow": {
          "value": 44,
          "perSecond": 0.4888888888888889
        },
        "upstream_rq_pending_total": {
          "value": 356,
          "perSecond": 3.9555555555555557
        },
        "upstream_rq_total": {
          "value": 359956,
          "perSecond": 3999.511111111111
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
        "max": 36.669439,
        "min": 0.323264,
        "mean": 0.72551,
        "pstdev": 0.884742,
        "percentiles": {
          "p50": 0.6039990000000001,
          "p75": 0.678047,
          "p80": 0.714751,
          "p90": 0.8924789999999999,
          "p95": 1.133887,
          "p99": 2.641151,
          "p999": 15.983615
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 175.15625,
            "min": 164.38671875,
            "mean": 167.263671875
          },
          "cpu": {
            "max": 1.1333333333333446,
            "min": 1.066666666666644,
            "mean": 1.0833333333333186
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 104.26953125,
            "min": 103.26953125,
            "mean": 103.943359375
          },
          "cpu": {
            "max": 69.97441389896726,
            "min": 69.97441389896726,
            "mean": 69.97441389896726
          }
        }
      },
      "poolOverflow": 26,
      "upstreamConnections": 97,
      "counters": {
        "benchmark.http_2xx": {
          "value": 287973,
          "perSecond": 3199.7
        },
        "benchmark.pool_overflow": {
          "value": 26,
          "perSecond": 0.28888888888888886
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
          "value": 97,
          "perSecond": 1.0777777777777777
        },
        "upstream_cx_rx_bytes_total": {
          "value": 45211761,
          "perSecond": 502352.9
        },
        "upstream_cx_total": {
          "value": 97,
          "perSecond": 1.0777777777777777
        },
        "upstream_cx_tx_bytes_total": {
          "value": 12958830,
          "perSecond": 143987
        },
        "upstream_rq_pending_overflow": {
          "value": 26,
          "perSecond": 0.28888888888888886
        },
        "upstream_rq_pending_total": {
          "value": 97,
          "perSecond": 1.0777777777777777
        },
        "upstream_rq_total": {
          "value": 287974,
          "perSecond": 3199.711111111111
        }
      }
    },
    {
      "testName": "scaling down httproutes to 100 with 20 routes per hostname at 500 rps",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-down",
      "throughput": 2020.9325842696628,
      "totalRequests": 179863,
      "latency": {
        "max": 116.969471,
        "min": 0.343568,
        "mean": 0.5027900000000001,
        "pstdev": 0.879488,
        "percentiles": {
          "p50": 0.449119,
          "p75": 0.47257499999999997,
          "p80": 0.480447,
          "p90": 0.524863,
          "p95": 0.707935,
          "p99": 1.474047,
          "p999": 4.035583
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 168.1015625,
            "min": 152.44921875,
            "mean": 156.88671875
          },
          "cpu": {
            "max": 1.2000000000000455,
            "min": 1.066666666666644,
            "mean": 1.1101449275362312
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 103.96875,
            "min": 103.328125,
            "mean": 103.60026041666667
          },
          "cpu": {
            "max": 48.5367403314919,
            "min": 28.940529475587816,
            "mean": 44.45150091159167
          }
        }
      },
      "poolOverflow": 137,
      "upstreamConnections": 32,
      "counters": {
        "benchmark.http_2xx": {
          "value": 179863,
          "perSecond": 2020.9325842696628
        },
        "benchmark.pool_overflow": {
          "value": 137,
          "perSecond": 1.5393258426966292
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
          "value": 28238491,
          "perSecond": 317286.4157303371
        },
        "upstream_cx_total": {
          "value": 32,
          "perSecond": 0.3595505617977528
        },
        "upstream_cx_tx_bytes_total": {
          "value": 8093835,
          "perSecond": 90941.96629213484
        },
        "upstream_rq_pending_overflow": {
          "value": 137,
          "perSecond": 1.5393258426966292
        },
        "upstream_rq_pending_total": {
          "value": 32,
          "perSecond": 0.3595505617977528
        },
        "upstream_rq_total": {
          "value": 179863,
          "perSecond": 2020.9325842696628
        }
      }
    },
    {
      "testName": "scaling down httproutes to 50 with 10 routes per hostname at 300 rps",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-down",
      "throughput": 1199.9444444444443,
      "totalRequests": 107995,
      "latency": {
        "max": 24.568831,
        "min": 0.35208,
        "mean": 0.47963999999999996,
        "pstdev": 0.23649,
        "percentiles": {
          "p50": 0.449839,
          "p75": 0.467263,
          "p80": 0.47289499999999995,
          "p90": 0.501663,
          "p95": 0.555615,
          "p99": 1.2311670000000001,
          "p999": 3.451903
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 154.71875,
            "min": 149.3203125,
            "mean": 152.18098958333334
          },
          "cpu": {
            "max": 1.066666666666644,
            "min": 0.9999999999999432,
            "mean": 1.0333333333333174
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 104.23828125,
            "min": 103.3203125,
            "mean": 103.74674479166667
          },
          "cpu": {
            "max": 30.417310579126177,
            "min": 21.761050943483184,
            "mean": 27.464019054584
          }
        }
      },
      "poolOverflow": 5,
      "upstreamConnections": 16,
      "counters": {
        "benchmark.http_2xx": {
          "value": 107995,
          "perSecond": 1199.9444444444443
        },
        "benchmark.pool_overflow": {
          "value": 5,
          "perSecond": 0.05555555555555555
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
          "value": 16,
          "perSecond": 0.17777777777777778
        },
        "upstream_cx_rx_bytes_total": {
          "value": 16955215,
          "perSecond": 188391.27777777778
        },
        "upstream_cx_total": {
          "value": 16,
          "perSecond": 0.17777777777777778
        },
        "upstream_cx_tx_bytes_total": {
          "value": 4859775,
          "perSecond": 53997.5
        },
        "upstream_rq_pending_overflow": {
          "value": 5,
          "perSecond": 0.05555555555555555
        },
        "upstream_rq_pending_total": {
          "value": 16,
          "perSecond": 0.17777777777777778
        },
        "upstream_rq_total": {
          "value": 107995,
          "perSecond": 1199.9444444444443
        }
      }
    },
    {
      "testName": "scaling down httproutes to 10 with 2 routes per hostname at 100 rps",
      "routes": 10,
      "routesPerHostname": 2,
      "phase": "scaling-down",
      "throughput": 404.46067415730334,
      "totalRequests": 35997,
      "latency": {
        "max": 40.593407,
        "min": 0.364272,
        "mean": 0.522306,
        "pstdev": 0.45468800000000004,
        "percentiles": {
          "p50": 0.459391,
          "p75": 0.5370550000000001,
          "p80": 0.555743,
          "p90": 0.606143,
          "p95": 0.663007,
          "p99": 1.234431,
          "p999": 5.294847
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 152.5703125,
            "min": 142.671875,
            "mean": 147.36536458333333
          },
          "cpu": {
            "max": 1.2002400480096473,
            "min": 1.066666666666644,
            "mean": 1.1091905887090083
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 103.80859375,
            "min": 103.28125,
            "mean": 103.47591145833333
          },
          "cpu": {
            "max": 12.533122028526384,
            "min": 10.097419137465744,
            "mean": 10.792540065476103
          }
        }
      },
      "poolOverflow": 3,
      "upstreamConnections": 10,
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
          "value": 10,
          "perSecond": 0.11235955056179775
        },
        "upstream_cx_rx_bytes_total": {
          "value": 5651529,
          "perSecond": 63500.32584269663
        },
        "upstream_cx_total": {
          "value": 10,
          "perSecond": 0.11235955056179775
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
          "value": 10,
          "perSecond": 0.11235955056179775
        },
        "upstream_rq_total": {
          "value": 35997,
          "perSecond": 404.46067415730334
        }
      }
    }
  ]
};
