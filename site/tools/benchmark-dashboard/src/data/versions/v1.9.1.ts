import { TestSuite } from '../types';

// Benchmark data extracted from release artifact for version 1.9.1
// Generated from benchmark_result.json

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.9.1",
    "runId": "1.9.1-release-2026-08-28",
    "date": "2026-08-28T11:34:36Z",
    "environment": "GitHub Release",
    "description": "Benchmark results for version 1.9.1 from release artifacts",
    "downloadUrl": "https://github.com/envoyproxy/gateway/releases/download/v1.9.1/benchmark_report.zip",
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
      "throughput": 399.9555555555556,
      "totalRequests": 35996,
      "latency": {
        "max": 77.910015,
        "min": 0.375072,
        "mean": 0.50444,
        "pstdev": 0.577751,
        "percentiles": {
          "p50": 0.46014299999999997,
          "p75": 0.474223,
          "p80": 0.478895,
          "p90": 0.505439,
          "p95": 0.563295,
          "p99": 1.510847,
          "p999": 5.835263
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 138.44921875,
            "min": 123.19140625,
            "mean": 135.57317708333332
          },
          "cpu": {
            "max": 0.9999999999999993,
            "min": 0.33333333333333287,
            "mean": 0.5333333333333334
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 20.9609375,
            "min": 8.80859375,
            "mean": 19.4359375
          },
          "cpu": {
            "max": 10.31448833592535,
            "min": 6.362209290343487,
            "mean": 9.775797993982346
          }
        }
      },
      "poolOverflow": 4,
      "upstreamConnections": 10,
      "counters": {
        "benchmark.http_2xx": {
          "value": 35996,
          "perSecond": 399.9555555555556
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
          "value": 10,
          "perSecond": 0.1111111111111111
        },
        "upstream_cx_rx_bytes_total": {
          "value": 5651372,
          "perSecond": 62793.02222222222
        },
        "upstream_cx_total": {
          "value": 10,
          "perSecond": 0.1111111111111111
        },
        "upstream_cx_tx_bytes_total": {
          "value": 1619820,
          "perSecond": 17998
        },
        "upstream_rq_pending_overflow": {
          "value": 4,
          "perSecond": 0.044444444444444446
        },
        "upstream_rq_pending_total": {
          "value": 10,
          "perSecond": 0.1111111111111111
        },
        "upstream_rq_total": {
          "value": 35996,
          "perSecond": 399.9555555555556
        }
      }
    },
    {
      "testName": "scaling up httproutes to 50 with 10 routes per hostname at 300 rps",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-up",
      "throughput": 1199.888888888889,
      "totalRequests": 107990,
      "latency": {
        "max": 22.923263,
        "min": 0.357696,
        "mean": 0.496908,
        "pstdev": 0.308301,
        "percentiles": {
          "p50": 0.457759,
          "p75": 0.47529499999999997,
          "p80": 0.481951,
          "p90": 0.516463,
          "p95": 0.588959,
          "p99": 1.498815,
          "p999": 3.718399
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 142.02734375,
            "min": 135.8125,
            "mean": 139.61145833333333
          },
          "cpu": {
            "max": 5.2,
            "min": 0.46666666666666556,
            "mean": 1.4538461538461538
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 25.3515625,
            "min": 20.5625,
            "mean": 24.327864583333334
          },
          "cpu": {
            "max": 30.64405834914612,
            "min": 23.798477049221198,
            "mean": 29.62114258323851
          }
        }
      },
      "poolOverflow": 10,
      "upstreamConnections": 20,
      "counters": {
        "benchmark.http_2xx": {
          "value": 107990,
          "perSecond": 1199.888888888889
        },
        "benchmark.pool_overflow": {
          "value": 10,
          "perSecond": 0.1111111111111111
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
          "value": 20,
          "perSecond": 0.2222222222222222
        },
        "upstream_cx_rx_bytes_total": {
          "value": 16954430,
          "perSecond": 188382.55555555556
        },
        "upstream_cx_total": {
          "value": 20,
          "perSecond": 0.2222222222222222
        },
        "upstream_cx_tx_bytes_total": {
          "value": 4859550,
          "perSecond": 53995
        },
        "upstream_rq_pending_overflow": {
          "value": 10,
          "perSecond": 0.1111111111111111
        },
        "upstream_rq_pending_total": {
          "value": 20,
          "perSecond": 0.2222222222222222
        },
        "upstream_rq_total": {
          "value": 107990,
          "perSecond": 1199.888888888889
        }
      }
    },
    {
      "testName": "scaling up httproutes to 100 with 20 routes per hostname at 500 rps",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-up",
      "throughput": 1999.1333333333334,
      "totalRequests": 179922,
      "latency": {
        "max": 73.310207,
        "min": 0.329568,
        "mean": 0.51358,
        "pstdev": 0.49629700000000004,
        "percentiles": {
          "p50": 0.458383,
          "p75": 0.48324700000000004,
          "p80": 0.49276699999999996,
          "p90": 0.5533750000000001,
          "p95": 0.776159,
          "p99": 1.637311,
          "p999": 3.839103
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 148.90625,
            "min": 134.8359375,
            "mean": 145.43580729166666
          },
          "cpu": {
            "max": 8.400000000000002,
            "min": 0.6666666666666702,
            "mean": 2.1283950617283947
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 29.3984375,
            "min": 24.5546875,
            "mean": 28.605598958333335
          },
          "cpu": {
            "max": 49.40438007771103,
            "min": 26.110288373890516,
            "mean": 34.374298546291456
          }
        }
      },
      "poolOverflow": 78,
      "upstreamConnections": 34,
      "counters": {
        "benchmark.http_2xx": {
          "value": 179922,
          "perSecond": 1999.1333333333334
        },
        "benchmark.pool_overflow": {
          "value": 78,
          "perSecond": 0.8666666666666667
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
          "value": 34,
          "perSecond": 0.37777777777777777
        },
        "upstream_cx_rx_bytes_total": {
          "value": 28247754,
          "perSecond": 313863.93333333335
        },
        "upstream_cx_total": {
          "value": 34,
          "perSecond": 0.37777777777777777
        },
        "upstream_cx_tx_bytes_total": {
          "value": 8096490,
          "perSecond": 89961
        },
        "upstream_rq_pending_overflow": {
          "value": 78,
          "perSecond": 0.8666666666666667
        },
        "upstream_rq_pending_total": {
          "value": 34,
          "perSecond": 0.37777777777777777
        },
        "upstream_rq_total": {
          "value": 179922,
          "perSecond": 1999.1333333333334
        }
      }
    },
    {
      "testName": "scaling up httproutes to 300 with 60 routes per hostname at 800 rps",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-up",
      "throughput": 3199.722222222222,
      "totalRequests": 287975,
      "latency": {
        "max": 59.658239,
        "min": 0.33262400000000003,
        "mean": 0.7722230000000001,
        "pstdev": 1.462783,
        "percentiles": {
          "p50": 0.602591,
          "p75": 0.734111,
          "p80": 0.788959,
          "p90": 0.9867509999999999,
          "p95": 1.266559,
          "p99": 3.017215,
          "p999": 29.073407
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 165.73828125,
            "min": 155.859375,
            "mean": 162.00885416666668
          },
          "cpu": {
            "max": 14.466666666666656,
            "min": 0.86666666666666,
            "mean": 3.3022222222222215
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 48.390625,
            "min": 40.6640625,
            "mean": 43.76875
          },
          "cpu": {
            "max": 72.39989733059555,
            "min": 71.9838002171552,
            "mean": 72.17059582761327
          }
        }
      },
      "poolOverflow": 21,
      "upstreamConnections": 150,
      "counters": {
        "benchmark.http_2xx": {
          "value": 287975,
          "perSecond": 3199.722222222222
        },
        "benchmark.pool_overflow": {
          "value": 21,
          "perSecond": 0.23333333333333334
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
          "value": 150,
          "perSecond": 1.6666666666666667
        },
        "upstream_cx_rx_bytes_total": {
          "value": 45212075,
          "perSecond": 502356.3888888889
        },
        "upstream_cx_total": {
          "value": 150,
          "perSecond": 1.6666666666666667
        },
        "upstream_cx_tx_bytes_total": {
          "value": 12959010,
          "perSecond": 143989
        },
        "upstream_rq_pending_overflow": {
          "value": 21,
          "perSecond": 0.23333333333333334
        },
        "upstream_rq_pending_total": {
          "value": 150,
          "perSecond": 1.6666666666666667
        },
        "upstream_rq_total": {
          "value": 287978,
          "perSecond": 3199.7555555555555
        }
      }
    },
    {
      "testName": "scaling up httproutes to 500 with 100 routes per hostname at 1000 rps",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-up",
      "throughput": 3998.233333333333,
      "totalRequests": 359841,
      "latency": {
        "max": 152.444927,
        "min": 0.324048,
        "mean": 1.414823,
        "pstdev": 5.364163,
        "percentiles": {
          "p50": 0.691711,
          "p75": 0.876799,
          "p80": 0.958783,
          "p90": 1.3818229999999998,
          "p95": 2.3194869999999996,
          "p99": 25.915391,
          "p999": 70.38566300000001
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 189.32421875,
            "min": 178.6484375,
            "mean": 186.11822916666668
          },
          "cpu": {
            "max": 14.000000000000057,
            "min": 0.8666666666666363,
            "mean": 3.546666666666665
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 69.41015625,
            "min": 57.1953125,
            "mean": 64.64830729166667
          },
          "cpu": {
            "max": 85.34406553290002,
            "min": 59.30669878561757,
            "mean": 81.9806609448706
          }
        }
      },
      "poolOverflow": 136,
      "upstreamConnections": 264,
      "counters": {
        "benchmark.http_2xx": {
          "value": 359841,
          "perSecond": 3998.233333333333
        },
        "benchmark.pool_overflow": {
          "value": 136,
          "perSecond": 1.511111111111111
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
          "value": 56495037,
          "perSecond": 627722.6333333333
        },
        "upstream_cx_total": {
          "value": 264,
          "perSecond": 2.933333333333333
        },
        "upstream_cx_tx_bytes_total": {
          "value": 16193880,
          "perSecond": 179932
        },
        "upstream_rq_pending_overflow": {
          "value": 136,
          "perSecond": 1.511111111111111
        },
        "upstream_rq_pending_total": {
          "value": 264,
          "perSecond": 2.933333333333333
        },
        "upstream_rq_total": {
          "value": 359864,
          "perSecond": 3998.488888888889
        }
      }
    },
    {
      "testName": "scaling up httproutes to 1000 with 200 routes per hostname at 2000 rps",
      "routes": 1000,
      "routesPerHostname": 200,
      "phase": "scaling-up",
      "throughput": 5033.877777777778,
      "totalRequests": 453049,
      "latency": {
        "max": 1579.614207,
        "min": 2.169856,
        "mean": 63.414631,
        "pstdev": 20.780189,
        "percentiles": {
          "p50": 62.693374999999996,
          "p75": 71.675903,
          "p80": 73.490431,
          "p90": 77.57414299999999,
          "p95": 80.990207,
          "p99": 95.285247,
          "p999": 162.840575
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 260.15234375,
            "min": 238.921875,
            "mean": 252.5546875
          },
          "cpu": {
            "max": 13.933333333333167,
            "min": 0.9999999999998485,
            "mean": 3.255555555555563
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 104.828125,
            "min": 98.015625,
            "mean": 103.42317708333333
          },
          "cpu": {
            "max": 99.11910950149134,
            "min": 99.11910950149132,
            "mean": 99.11910950149132
          }
        }
      },
      "poolOverflow": 79,
      "upstreamConnections": 321,
      "counters": {
        "benchmark.http_2xx": {
          "value": 453049,
          "perSecond": 5033.877777777778
        },
        "benchmark.pool_overflow": {
          "value": 79,
          "perSecond": 0.8777777777777778
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
          "value": 321,
          "perSecond": 3.566666666666667
        },
        "upstream_cx_rx_bytes_total": {
          "value": 71128693,
          "perSecond": 790318.8111111111
        },
        "upstream_cx_total": {
          "value": 321,
          "perSecond": 3.566666666666667
        },
        "upstream_cx_tx_bytes_total": {
          "value": 20401650,
          "perSecond": 226685
        },
        "upstream_rq_pending_overflow": {
          "value": 79,
          "perSecond": 0.8777777777777778
        },
        "upstream_rq_pending_total": {
          "value": 321,
          "perSecond": 3.566666666666667
        },
        "upstream_rq_total": {
          "value": 453370,
          "perSecond": 5037.444444444444
        }
      }
    },
    {
      "testName": "scaling down httproutes to 500 with 100 routes per hostname at 1000 rps",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-down",
      "throughput": 3997.1666666666665,
      "totalRequests": 359745,
      "latency": {
        "max": 119.324671,
        "min": 0.329344,
        "mean": 1.466548,
        "pstdev": 3.969409,
        "percentiles": {
          "p50": 0.732255,
          "p75": 1.0117749999999999,
          "p80": 1.145663,
          "p90": 1.885311,
          "p95": 3.481855,
          "p99": 19.318783,
          "p999": 55.023615
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 260.234375,
            "min": 193.13671875,
            "mean": 201.78346354166666
          },
          "cpu": {
            "max": 12.999999999999922,
            "min": 1.0666666666664544,
            "mean": 3.669444444444455
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 103.82421875,
            "min": 102.78515625,
            "mean": 103.57369791666666
          },
          "cpu": {
            "max": 85.67922944580592,
            "min": 44.92754001016751,
            "mean": 72.75851749097717
          }
        }
      },
      "poolOverflow": 253,
      "upstreamConnections": 147,
      "counters": {
        "benchmark.http_2xx": {
          "value": 359745,
          "perSecond": 3997.1666666666665
        },
        "benchmark.pool_overflow": {
          "value": 253,
          "perSecond": 2.811111111111111
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
          "value": 147,
          "perSecond": 1.6333333333333333
        },
        "upstream_cx_rx_bytes_total": {
          "value": 56479965,
          "perSecond": 627555.1666666666
        },
        "upstream_cx_total": {
          "value": 147,
          "perSecond": 1.6333333333333333
        },
        "upstream_cx_tx_bytes_total": {
          "value": 16188615,
          "perSecond": 179873.5
        },
        "upstream_rq_pending_overflow": {
          "value": 253,
          "perSecond": 2.811111111111111
        },
        "upstream_rq_pending_total": {
          "value": 147,
          "perSecond": 1.6333333333333333
        },
        "upstream_rq_total": {
          "value": 359747,
          "perSecond": 3997.188888888889
        }
      }
    },
    {
      "testName": "scaling down httproutes to 300 with 60 routes per hostname at 800 rps",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-down",
      "throughput": 3232.14606741573,
      "totalRequests": 287661,
      "latency": {
        "max": 137.887743,
        "min": 0.33240000000000003,
        "mean": 0.886362,
        "pstdev": 1.9994460000000003,
        "percentiles": {
          "p50": 0.652735,
          "p75": 0.805759,
          "p80": 0.855807,
          "p90": 1.079807,
          "p95": 1.504255,
          "p99": 5.465343,
          "p999": 24.437759
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 201.88671875,
            "min": 165.61328125,
            "mean": 175.87825520833334
          },
          "cpu": {
            "max": 5.088036419629429,
            "min": 1.1333333333330606,
            "mean": 2.1542943856260455
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 103.4453125,
            "min": 102.6484375,
            "mean": 103.11315104166667
          },
          "cpu": {
            "max": 71.47829536904024,
            "min": 43.776501550183156,
            "mean": 65.07638751993836
          }
        }
      },
      "poolOverflow": 336,
      "upstreamConnections": 64,
      "counters": {
        "benchmark.http_2xx": {
          "value": 287661,
          "perSecond": 3232.14606741573
        },
        "benchmark.pool_overflow": {
          "value": 336,
          "perSecond": 3.7752808988764044
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
          "value": 64,
          "perSecond": 0.7191011235955056
        },
        "upstream_cx_rx_bytes_total": {
          "value": 45162777,
          "perSecond": 507446.93258426967
        },
        "upstream_cx_total": {
          "value": 64,
          "perSecond": 0.7191011235955056
        },
        "upstream_cx_tx_bytes_total": {
          "value": 12944790,
          "perSecond": 145447.07865168538
        },
        "upstream_rq_pending_overflow": {
          "value": 336,
          "perSecond": 3.7752808988764044
        },
        "upstream_rq_pending_total": {
          "value": 64,
          "perSecond": 0.7191011235955056
        },
        "upstream_rq_total": {
          "value": 287662,
          "perSecond": 3232.1573033707864
        }
      }
    },
    {
      "testName": "scaling down httproutes to 100 with 20 routes per hostname at 500 rps",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-down",
      "throughput": 1996.5,
      "totalRequests": 179685,
      "latency": {
        "max": 126.390271,
        "min": 0.34728,
        "mean": 0.57243,
        "pstdev": 1.541711,
        "percentiles": {
          "p50": 0.469327,
          "p75": 0.49329500000000004,
          "p80": 0.504335,
          "p90": 0.6470389999999999,
          "p95": 0.912063,
          "p99": 2.025535,
          "p999": 7.168255
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 176.046875,
            "min": 151.49609375,
            "mean": 157.49934895833334
          },
          "cpu": {
            "max": 6.733333333333273,
            "min": 1.1333333333334394,
            "mean": 2.3805555555555538
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 103.50390625,
            "min": 102.7265625,
            "mean": 103.05885416666666
          },
          "cpu": {
            "max": 52.653675744320395,
            "min": 32.823886508356296,
            "mean": 48.774853021114204
          }
        }
      },
      "poolOverflow": 314,
      "upstreamConnections": 39,
      "counters": {
        "benchmark.http_2xx": {
          "value": 179685,
          "perSecond": 1996.5
        },
        "benchmark.pool_overflow": {
          "value": 314,
          "perSecond": 3.488888888888889
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
          "value": 39,
          "perSecond": 0.43333333333333335
        },
        "upstream_cx_rx_bytes_total": {
          "value": 28210545,
          "perSecond": 313450.5
        },
        "upstream_cx_total": {
          "value": 39,
          "perSecond": 0.43333333333333335
        },
        "upstream_cx_tx_bytes_total": {
          "value": 8085870,
          "perSecond": 89843
        },
        "upstream_rq_pending_overflow": {
          "value": 314,
          "perSecond": 3.488888888888889
        },
        "upstream_rq_pending_total": {
          "value": 39,
          "perSecond": 0.43333333333333335
        },
        "upstream_rq_total": {
          "value": 179686,
          "perSecond": 1996.5111111111112
        }
      }
    },
    {
      "testName": "scaling down httproutes to 50 with 10 routes per hostname at 300 rps",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-down",
      "throughput": 1212.5842696629213,
      "totalRequests": 107920,
      "latency": {
        "max": 125.378559,
        "min": 0.359936,
        "mean": 0.5203220000000001,
        "pstdev": 1.131535,
        "percentiles": {
          "p50": 0.459279,
          "p75": 0.48033499999999996,
          "p80": 0.488623,
          "p90": 0.5460470000000001,
          "p95": 0.6728949999999999,
          "p99": 1.583871,
          "p999": 3.929727
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 160.91015625,
            "min": 151.16796875,
            "mean": 155.06901041666666
          },
          "cpu": {
            "max": 1.3999999999998636,
            "min": 1.1333333333330606,
            "mean": 1.2634920634920515
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 103.06640625,
            "min": 102.48828125,
            "mean": 102.75911458333333
          },
          "cpu": {
            "max": 31.010104037091104,
            "min": 20.73928894544472,
            "mean": 29.460187498476223
          }
        }
      },
      "poolOverflow": 79,
      "upstreamConnections": 23,
      "counters": {
        "benchmark.http_2xx": {
          "value": 107920,
          "perSecond": 1212.5842696629213
        },
        "benchmark.pool_overflow": {
          "value": 79,
          "perSecond": 0.8876404494382022
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
          "value": 16943440,
          "perSecond": 190375.73033707865
        },
        "upstream_cx_total": {
          "value": 23,
          "perSecond": 0.25842696629213485
        },
        "upstream_cx_tx_bytes_total": {
          "value": 4856400,
          "perSecond": 54566.29213483146
        },
        "upstream_rq_pending_overflow": {
          "value": 79,
          "perSecond": 0.8876404494382022
        },
        "upstream_rq_pending_total": {
          "value": 23,
          "perSecond": 0.25842696629213485
        },
        "upstream_rq_total": {
          "value": 107920,
          "perSecond": 1212.5842696629213
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
        "max": 26.908671,
        "min": 0.371264,
        "mean": 0.507798,
        "pstdev": 0.417179,
        "percentiles": {
          "p50": 0.46487900000000004,
          "p75": 0.479583,
          "p80": 0.48347100000000004,
          "p90": 0.5088630000000001,
          "p95": 0.564767,
          "p99": 1.603967,
          "p999": 5.539071
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 208.40234375,
            "min": 152.953125,
            "mean": 168.809765625
          },
          "cpu": {
            "max": 1.3999999999998636,
            "min": 1.1333333333334394,
            "mean": 1.2579710144927463
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 103.06640625,
            "min": 102.5234375,
            "mean": 102.66536458333333
          },
          "cpu": {
            "max": 10.406962201998446,
            "min": 6.026136998279408,
            "mean": 8.34873453076639
          }
        }
      },
      "poolOverflow": 0,
      "upstreamConnections": 11,
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
          "value": 11,
          "perSecond": 0.12359550561797752
        },
        "upstream_cx_rx_bytes_total": {
          "value": 5652000,
          "perSecond": 63505.61797752809
        },
        "upstream_cx_total": {
          "value": 11,
          "perSecond": 0.12359550561797752
        },
        "upstream_cx_tx_bytes_total": {
          "value": 1620000,
          "perSecond": 18202.247191011236
        },
        "upstream_rq_pending_total": {
          "value": 11,
          "perSecond": 0.12359550561797752
        },
        "upstream_rq_total": {
          "value": 36000,
          "perSecond": 404.4943820224719
        }
      }
    }
  ]
};
