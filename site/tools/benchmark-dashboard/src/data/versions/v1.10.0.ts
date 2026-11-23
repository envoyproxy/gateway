import { TestSuite } from '../types';

// Benchmark data extracted from release artifact for version 1.10.0
// Generated from benchmark_result.json

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.10.0",
    "runId": "1.10.0-release-2026-11-06",
    "date": "2026-11-06T11:48:01Z",
    "environment": "GitHub Release",
    "description": "Benchmark results for version 1.10.0 from release artifacts",
    "downloadUrl": "https://github.com/envoyproxy/gateway/releases/download/v1.10.0/benchmark_report.zip",
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
      "throughput": 400,
      "totalRequests": 36000,
      "latency": {
        "max": 18.256895,
        "min": 0.294304,
        "mean": 0.387363,
        "pstdev": 0.232102,
        "percentiles": {
          "p50": 0.367903,
          "p75": 0.380303,
          "p80": 0.383775,
          "p90": 0.395727,
          "p95": 0.42505499999999996,
          "p99": 0.928127,
          "p999": 2.873727
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 133.82421875,
            "min": 114.51953125,
            "mean": 127.89609375
          },
          "cpu": {
            "max": 0.933333333333333,
            "min": 0.26666666666666616,
            "mean": 0.44615384615384623
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 18.328125,
            "min": 14.90625,
            "mean": 17.280729166666667
          },
          "cpu": {
            "max": 8.135989821294315,
            "min": 8.099383395727818,
            "mean": 8.125089762414552
          }
        }
      },
      "poolOverflow": 0,
      "upstreamConnections": 8,
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
          "value": 8,
          "perSecond": 0.08888888888888889
        },
        "upstream_cx_rx_bytes_total": {
          "value": 5652000,
          "perSecond": 62800
        },
        "upstream_cx_total": {
          "value": 8,
          "perSecond": 0.08888888888888889
        },
        "upstream_cx_tx_bytes_total": {
          "value": 1620000,
          "perSecond": 18000
        },
        "upstream_rq_pending_total": {
          "value": 8,
          "perSecond": 0.08888888888888889
        },
        "upstream_rq_total": {
          "value": 36000,
          "perSecond": 400
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
        "max": 50.227199,
        "min": 0.266992,
        "mean": 0.368185,
        "pstdev": 0.443388,
        "percentiles": {
          "p50": 0.346815,
          "p75": 0.358399,
          "p80": 0.36247900000000005,
          "p90": 0.37500700000000003,
          "p95": 0.410959,
          "p99": 0.888159,
          "p999": 2.651007
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 140.09765625,
            "min": 133.82421875,
            "mean": 137.76354166666667
          },
          "cpu": {
            "max": 0.40000000000000036,
            "min": 0.3333333333333336,
            "mean": 0.38611111111111107
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 21.50390625,
            "min": 17.94140625,
            "mean": 20.298828125
          },
          "cpu": {
            "max": 23.00534682080925,
            "min": 14.97980593177511,
            "mean": 21.379554429593956
          }
        }
      },
      "poolOverflow": 18,
      "upstreamConnections": 17,
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
          "value": 17,
          "perSecond": 0.18888888888888888
        },
        "upstream_cx_rx_bytes_total": {
          "value": 16953174,
          "perSecond": 188368.6
        },
        "upstream_cx_total": {
          "value": 17,
          "perSecond": 0.18888888888888888
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
          "value": 17,
          "perSecond": 0.18888888888888888
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
      "throughput": 1999.9555555555555,
      "totalRequests": 179996,
      "latency": {
        "max": 7.618815000000001,
        "min": 0.24870399999999998,
        "mean": 0.365344,
        "pstdev": 0.16806300000000002,
        "percentiles": {
          "p50": 0.344671,
          "p75": 0.35983899999999996,
          "p80": 0.363679,
          "p90": 0.377311,
          "p95": 0.42484700000000003,
          "p99": 1.048895,
          "p999": 2.544255
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 143.2109375,
            "min": 133.5859375,
            "mean": 140.65677083333333
          },
          "cpu": {
            "max": 3.1333333333333346,
            "min": 0.40000000000000036,
            "mean": 0.9666666666666663
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 25.86328125,
            "min": 21.19921875,
            "mean": 24.626692708333334
          },
          "cpu": {
            "max": 38.12149596068019,
            "min": 26.79154171428569,
            "mean": 36.226106955775435
          }
        }
      },
      "poolOverflow": 4,
      "upstreamConnections": 18,
      "counters": {
        "benchmark.http_2xx": {
          "value": 179996,
          "perSecond": 1999.9555555555555
        },
        "benchmark.pool_overflow": {
          "value": 4,
          "perSecond": 0.044444444444444446
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
          "value": 18,
          "perSecond": 0.2
        },
        "upstream_cx_rx_bytes_total": {
          "value": 28259372,
          "perSecond": 313993.02222222224
        },
        "upstream_cx_total": {
          "value": 18,
          "perSecond": 0.2
        },
        "upstream_cx_tx_bytes_total": {
          "value": 8099820,
          "perSecond": 89998
        },
        "upstream_rq_pending_overflow": {
          "value": 4,
          "perSecond": 0.044444444444444446
        },
        "upstream_rq_pending_total": {
          "value": 18,
          "perSecond": 0.2
        },
        "upstream_rq_total": {
          "value": 179996,
          "perSecond": 1999.9555555555555
        }
      }
    },
    {
      "testName": "scaling up httproutes to 300 with 60 routes per hostname at 800 rps",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-up",
      "throughput": 3198.511111111111,
      "totalRequests": 287866,
      "latency": {
        "max": 79.23711899999999,
        "min": 0.21308000000000002,
        "mean": 0.396919,
        "pstdev": 0.65739,
        "percentiles": {
          "p50": 0.33897499999999997,
          "p75": 0.383055,
          "p80": 0.430335,
          "p90": 0.489807,
          "p95": 0.629247,
          "p99": 1.241663,
          "p999": 3.6415990000000003
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 158.28515625,
            "min": 150.8359375,
            "mean": 154.995703125
          },
          "cpu": {
            "max": 9.733333333333338,
            "min": 0.4666666666666685,
            "mean": 2.3422222222222215
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 41.390625,
            "min": 37.94921875,
            "mean": 40.144921875
          },
          "cpu": {
            "max": 58.18939313984169,
            "min": 38.160174937874864,
            "mean": 55.60421330836275
          }
        }
      },
      "poolOverflow": 134,
      "upstreamConnections": 42,
      "counters": {
        "benchmark.http_2xx": {
          "value": 287866,
          "perSecond": 3198.511111111111
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
          "value": 42,
          "perSecond": 0.4666666666666667
        },
        "upstream_cx_rx_bytes_total": {
          "value": 45194962,
          "perSecond": 502166.2444444444
        },
        "upstream_cx_total": {
          "value": 42,
          "perSecond": 0.4666666666666667
        },
        "upstream_cx_tx_bytes_total": {
          "value": 12953970,
          "perSecond": 143933
        },
        "upstream_rq_pending_overflow": {
          "value": 134,
          "perSecond": 1.488888888888889
        },
        "upstream_rq_pending_total": {
          "value": 42,
          "perSecond": 0.4666666666666667
        },
        "upstream_rq_total": {
          "value": 287866,
          "perSecond": 3198.511111111111
        }
      }
    },
    {
      "testName": "scaling up httproutes to 500 with 100 routes per hostname at 1000 rps",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-up",
      "throughput": 4044.887640449438,
      "totalRequests": 359995,
      "latency": {
        "max": 46.878719,
        "min": 0.211168,
        "mean": 0.576632,
        "pstdev": 1.1048580000000001,
        "percentiles": {
          "p50": 0.48795099999999997,
          "p75": 0.5631670000000001,
          "p80": 0.581727,
          "p90": 0.668095,
          "p95": 0.800863,
          "p99": 1.739263,
          "p999": 20.645887000000002
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 183.87890625,
            "min": 176.0546875,
            "mean": 181.13932291666666
          },
          "cpu": {
            "max": 9.600000000000032,
            "min": 0.5999999999999753,
            "mean": 2.3444444444444503
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 61.01171875,
            "min": 53.3046875,
            "mean": 57.701953125
          },
          "cpu": {
            "max": 63.71739130434779,
            "min": 63.39084723259477,
            "mean": 63.60798818224919
          }
        }
      },
      "poolOverflow": 2,
      "upstreamConnections": 156,
      "counters": {
        "benchmark.http_2xx": {
          "value": 359995,
          "perSecond": 4044.887640449438
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
          "value": 156,
          "perSecond": 1.752808988764045
        },
        "upstream_cx_rx_bytes_total": {
          "value": 56519215,
          "perSecond": 635047.3595505618
        },
        "upstream_cx_total": {
          "value": 156,
          "perSecond": 1.752808988764045
        },
        "upstream_cx_tx_bytes_total": {
          "value": 16199910,
          "perSecond": 182021.4606741573
        },
        "upstream_rq_pending_overflow": {
          "value": 2,
          "perSecond": 0.02247191011235955
        },
        "upstream_rq_pending_total": {
          "value": 156,
          "perSecond": 1.752808988764045
        },
        "upstream_rq_total": {
          "value": 359998,
          "perSecond": 4044.921348314607
        }
      }
    },
    {
      "testName": "scaling up httproutes to 1000 with 200 routes per hostname at 2000 rps",
      "routes": 1000,
      "routesPerHostname": 200,
      "phase": "scaling-up",
      "throughput": 7905.766666666666,
      "totalRequests": 711519,
      "latency": {
        "max": 719.683583,
        "min": 0.217528,
        "mean": 13.094173000000001,
        "pstdev": 21.331111,
        "percentiles": {
          "p50": 2.655359,
          "p75": 11.913727000000002,
          "p80": 36.429823,
          "p90": 45.764607,
          "p95": 52.527103,
          "p99": 64.30310300000001,
          "p999": 112.570367
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 250.03515625,
            "min": 228.36328125,
            "mean": 241.38203125
          },
          "cpu": {
            "max": 11.866666666666674,
            "min": 0.7333333333332348,
            "mean": 2.72000000000001
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 107.37890625,
            "min": 92.5703125,
            "mean": 102.38216145833333
          },
          "cpu": {
            "max": 98.74200844673862,
            "min": 56.27800771421867,
            "mean": 90.86690039394128
          }
        }
      },
      "poolOverflow": 48,
      "upstreamConnections": 352,
      "counters": {
        "benchmark.http_2xx": {
          "value": 711519,
          "perSecond": 7905.766666666666
        },
        "benchmark.pool_overflow": {
          "value": 48,
          "perSecond": 0.5333333333333333
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
          "value": 352,
          "perSecond": 3.911111111111111
        },
        "upstream_cx_rx_bytes_total": {
          "value": 111708483,
          "perSecond": 1241205.3666666667
        },
        "upstream_cx_total": {
          "value": 352,
          "perSecond": 3.911111111111111
        },
        "upstream_cx_tx_bytes_total": {
          "value": 32034195,
          "perSecond": 355935.5
        },
        "upstream_rq_pending_overflow": {
          "value": 48,
          "perSecond": 0.5333333333333333
        },
        "upstream_rq_pending_total": {
          "value": 352,
          "perSecond": 3.911111111111111
        },
        "upstream_rq_total": {
          "value": 711871,
          "perSecond": 7909.677777777778
        }
      }
    },
    {
      "testName": "scaling down httproutes to 500 with 100 routes per hostname at 1000 rps",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-down",
      "throughput": 3998.311111111111,
      "totalRequests": 359848,
      "latency": {
        "max": 64.378879,
        "min": 0.22332,
        "mean": 0.620247,
        "pstdev": 1.394215,
        "percentiles": {
          "p50": 0.464799,
          "p75": 0.530335,
          "p80": 0.5671670000000001,
          "p90": 0.756415,
          "p95": 1.044223,
          "p99": 2.6776310000000003,
          "p999": 25.581567
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 216.47265625,
            "min": 186.39453125,
            "mean": 194.10169270833333
          },
          "cpu": {
            "max": 3.800000000000144,
            "min": 0.8000000000000304,
            "mean": 1.5682539682539762
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 106.62109375,
            "min": 106.015625,
            "mean": 106.39192708333333
          },
          "cpu": {
            "max": 67.49039730349978,
            "min": 67.47633403678032,
            "mean": 67.48336567014005
          }
        }
      },
      "poolOverflow": 147,
      "upstreamConnections": 136,
      "counters": {
        "benchmark.http_2xx": {
          "value": 359848,
          "perSecond": 3998.311111111111
        },
        "benchmark.pool_overflow": {
          "value": 147,
          "perSecond": 1.6333333333333333
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
          "value": 136,
          "perSecond": 1.511111111111111
        },
        "upstream_cx_rx_bytes_total": {
          "value": 56496136,
          "perSecond": 627734.8444444444
        },
        "upstream_cx_total": {
          "value": 136,
          "perSecond": 1.511111111111111
        },
        "upstream_cx_tx_bytes_total": {
          "value": 16193385,
          "perSecond": 179926.5
        },
        "upstream_rq_pending_overflow": {
          "value": 147,
          "perSecond": 1.6333333333333333
        },
        "upstream_rq_pending_total": {
          "value": 136,
          "perSecond": 1.511111111111111
        },
        "upstream_rq_total": {
          "value": 359853,
          "perSecond": 3998.366666666667
        }
      }
    },
    {
      "testName": "scaling down httproutes to 300 with 60 routes per hostname at 800 rps",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-down",
      "throughput": 3196.3444444444444,
      "totalRequests": 287671,
      "latency": {
        "max": 69.672959,
        "min": 0.207712,
        "mean": 0.44837899999999997,
        "pstdev": 1.094921,
        "percentiles": {
          "p50": 0.34646299999999997,
          "p75": 0.434431,
          "p80": 0.451439,
          "p90": 0.558111,
          "p95": 0.728927,
          "p99": 1.5311350000000001,
          "p999": 11.941375
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 186.171875,
            "min": 160.8984375,
            "mean": 167.50026041666666
          },
          "cpu": {
            "max": 6.733333333333462,
            "min": 0.8666666666666363,
            "mean": 2.1507246376811806
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 106.46484375,
            "min": 105.6171875,
            "mean": 106.02408854166667
          },
          "cpu": {
            "max": 0,
            "min": 0,
            "mean": 0
          }
        }
      },
      "poolOverflow": 329,
      "upstreamConnections": 71,
      "counters": {
        "benchmark.http_2xx": {
          "value": 287671,
          "perSecond": 3196.3444444444444
        },
        "benchmark.pool_overflow": {
          "value": 329,
          "perSecond": 3.6555555555555554
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
          "value": 71,
          "perSecond": 0.7888888888888889
        },
        "upstream_cx_rx_bytes_total": {
          "value": 45164347,
          "perSecond": 501826.0777777778
        },
        "upstream_cx_total": {
          "value": 71,
          "perSecond": 0.7888888888888889
        },
        "upstream_cx_tx_bytes_total": {
          "value": 12945195,
          "perSecond": 143835.5
        },
        "upstream_rq_pending_overflow": {
          "value": 329,
          "perSecond": 3.6555555555555554
        },
        "upstream_rq_pending_total": {
          "value": 71,
          "perSecond": 0.7888888888888889
        },
        "upstream_rq_total": {
          "value": 287671,
          "perSecond": 3196.3444444444444
        }
      }
    },
    {
      "testName": "scaling down httproutes to 100 with 20 routes per hostname at 500 rps",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-down",
      "throughput": 2020.5393258426966,
      "totalRequests": 179828,
      "latency": {
        "max": 112.967679,
        "min": 0.2044,
        "mean": 0.433003,
        "pstdev": 0.731167,
        "percentiles": {
          "p50": 0.368783,
          "p75": 0.447903,
          "p80": 0.468719,
          "p90": 0.526111,
          "p95": 0.5988150000000001,
          "p99": 1.280575,
          "p999": 4.9054709999999995
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 170.625,
            "min": 145.703125,
            "mean": 154.51666666666668
          },
          "cpu": {
            "max": 1.066666666666644,
            "min": 0.8666666666666363,
            "mean": 0.9142857142857311
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 106.0546875,
            "min": 105.86328125,
            "mean": 105.93020833333334
          },
          "cpu": {
            "max": 34.36459663377853,
            "min": 22.64464919519858,
            "mean": 32.83407448123581
          }
        }
      },
      "poolOverflow": 171,
      "upstreamConnections": 43,
      "counters": {
        "benchmark.http_2xx": {
          "value": 179828,
          "perSecond": 2020.5393258426966
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
          "value": 43,
          "perSecond": 0.48314606741573035
        },
        "upstream_cx_rx_bytes_total": {
          "value": 28232996,
          "perSecond": 317224.67415730335
        },
        "upstream_cx_total": {
          "value": 43,
          "perSecond": 0.48314606741573035
        },
        "upstream_cx_tx_bytes_total": {
          "value": 8092305,
          "perSecond": 90924.77528089887
        },
        "upstream_rq_pending_overflow": {
          "value": 171,
          "perSecond": 1.9213483146067416
        },
        "upstream_rq_pending_total": {
          "value": 43,
          "perSecond": 0.48314606741573035
        },
        "upstream_rq_total": {
          "value": 179829,
          "perSecond": 2020.5505617977528
        }
      }
    },
    {
      "testName": "scaling down httproutes to 50 with 10 routes per hostname at 300 rps",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-down",
      "throughput": 1213.2022471910113,
      "totalRequests": 107975,
      "latency": {
        "max": 35.670015,
        "min": 0.26528,
        "mean": 0.434393,
        "pstdev": 0.45309099999999997,
        "percentiles": {
          "p50": 0.39854300000000004,
          "p75": 0.470655,
          "p80": 0.48329500000000003,
          "p90": 0.514831,
          "p95": 0.559071,
          "p99": 1.074943,
          "p999": 2.639103
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 155.921875,
            "min": 143.59765625,
            "mean": 149.91744791666667
          },
          "cpu": {
            "max": 1.066666666666644,
            "min": 0.8000000000000304,
            "mean": 0.9043478260869583
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 106.078125,
            "min": 105.59375,
            "mean": 105.748046875
          },
          "cpu": {
            "max": 21.399908759124088,
            "min": 12.949655387828349,
            "mean": 18.927623074297017
          }
        }
      },
      "poolOverflow": 25,
      "upstreamConnections": 18,
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
          "value": 18,
          "perSecond": 0.20224719101123595
        },
        "upstream_cx_rx_bytes_total": {
          "value": 16952075,
          "perSecond": 190472.75280898876
        },
        "upstream_cx_total": {
          "value": 18,
          "perSecond": 0.20224719101123595
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
          "value": 18,
          "perSecond": 0.20224719101123595
        },
        "upstream_rq_total": {
          "value": 107975,
          "perSecond": 1213.2022471910113
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
        "max": 44.040191,
        "min": 0.30449600000000004,
        "mean": 0.40148,
        "pstdev": 0.474409,
        "percentiles": {
          "p50": 0.37343899999999997,
          "p75": 0.38814299999999996,
          "p80": 0.391679,
          "p90": 0.404495,
          "p95": 0.436975,
          "p99": 0.974495,
          "p999": 3.582975
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 151.11328125,
            "min": 149.25,
            "mean": 150.15690104166666
          },
          "cpu": {
            "max": 1.0000000000000377,
            "min": 0.8000000000000304,
            "mean": 0.8833333333333353
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 105.80078125,
            "min": 105.54296875,
            "mean": 105.671484375
          },
          "cpu": {
            "max": 8.190930414385756,
            "min": 5.300359866645494,
            "mean": 7.470460062058507
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
