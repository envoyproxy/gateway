import { TestSuite } from '../types';

// Benchmark data extracted from release artifact for version 1.9.0
// Generated from benchmark_result.json

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.9.0",
    "runId": "1.9.0-release-2026-08-15",
    "date": "2026-08-15T12:31:24Z",
    "environment": "GitHub Release",
    "description": "Benchmark results for version 1.9.0 from release artifacts",
    "downloadUrl": "https://github.com/envoyproxy/gateway/releases/download/v1.9.0/benchmark_report.zip",
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
      "throughput": 404.4831460674157,
      "totalRequests": 35999,
      "latency": {
        "max": 26.657791,
        "min": 0.365776,
        "mean": 0.49240700000000004,
        "pstdev": 0.47715,
        "percentiles": {
          "p50": 0.444463,
          "p75": 0.456015,
          "p80": 0.460751,
          "p90": 0.484255,
          "p95": 0.536895,
          "p99": 1.668223,
          "p999": 7.041279
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 134.3046875,
            "min": 110.484375,
            "mean": 129.667578125
          },
          "cpu": {
            "max": 0.46666666666666706,
            "min": 0.3333333333333336,
            "mean": 0.41388888888888875
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 20.90625,
            "min": 7.12109375,
            "mean": 19.226302083333334
          },
          "cpu": {
            "max": 9.832710152838425,
            "min": 6.043408072986803,
            "mean": 9.407696136791584
          }
        }
      },
      "poolOverflow": 1,
      "upstreamConnections": 10,
      "counters": {
        "benchmark.http_2xx": {
          "value": 35999,
          "perSecond": 404.4831460674157
        },
        "benchmark.pool_overflow": {
          "value": 1,
          "perSecond": 0.011235955056179775
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
          "value": 5651843,
          "perSecond": 63503.85393258427
        },
        "upstream_cx_total": {
          "value": 10,
          "perSecond": 0.11235955056179775
        },
        "upstream_cx_tx_bytes_total": {
          "value": 1619955,
          "perSecond": 18201.74157303371
        },
        "upstream_rq_pending_overflow": {
          "value": 1,
          "perSecond": 0.011235955056179775
        },
        "upstream_rq_pending_total": {
          "value": 10,
          "perSecond": 0.11235955056179775
        },
        "upstream_rq_total": {
          "value": 35999,
          "perSecond": 404.4831460674157
        }
      }
    },
    {
      "testName": "scaling up httproutes to 50 with 10 routes per hostname at 300 rps",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-up",
      "throughput": 1199.5666666666666,
      "totalRequests": 107961,
      "latency": {
        "max": 51.097599,
        "min": 0.345792,
        "mean": 0.482854,
        "pstdev": 0.60319,
        "percentiles": {
          "p50": 0.436591,
          "p75": 0.456047,
          "p80": 0.460079,
          "p90": 0.49004699999999995,
          "p95": 0.5474870000000001,
          "p99": 1.3834229999999998,
          "p999": 4.798463
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 139.96484375,
            "min": 131.7421875,
            "mean": 137.19557291666666
          },
          "cpu": {
            "max": 0.5999999999999991,
            "min": 0.4666666666666685,
            "mean": 0.5420289855072469
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 25.0078125,
            "min": 20.3671875,
            "mean": 24.186197916666668
          },
          "cpu": {
            "max": 29.670826210826213,
            "min": 22.780171872013455,
            "mean": 26.225499041419834
          }
        }
      },
      "poolOverflow": 38,
      "upstreamConnections": 31,
      "counters": {
        "benchmark.http_2xx": {
          "value": 107961,
          "perSecond": 1199.5666666666666
        },
        "benchmark.pool_overflow": {
          "value": 38,
          "perSecond": 0.4222222222222222
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
          "value": 31,
          "perSecond": 0.34444444444444444
        },
        "upstream_cx_rx_bytes_total": {
          "value": 16949877,
          "perSecond": 188331.96666666667
        },
        "upstream_cx_total": {
          "value": 31,
          "perSecond": 0.34444444444444444
        },
        "upstream_cx_tx_bytes_total": {
          "value": 4858245,
          "perSecond": 53980.5
        },
        "upstream_rq_pending_overflow": {
          "value": 38,
          "perSecond": 0.4222222222222222
        },
        "upstream_rq_pending_total": {
          "value": 31,
          "perSecond": 0.34444444444444444
        },
        "upstream_rq_total": {
          "value": 107961,
          "perSecond": 1199.5666666666666
        }
      }
    },
    {
      "testName": "scaling up httproutes to 100 with 20 routes per hostname at 500 rps",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-up",
      "throughput": 1998.6333333333334,
      "totalRequests": 179877,
      "latency": {
        "max": 56.977407,
        "min": 0.340432,
        "mean": 0.488172,
        "pstdev": 0.391262,
        "percentiles": {
          "p50": 0.44089500000000004,
          "p75": 0.46265500000000004,
          "p80": 0.468751,
          "p90": 0.509503,
          "p95": 0.691199,
          "p99": 1.543487,
          "p999": 3.7213429999999996
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 147.06640625,
            "min": 138.8125,
            "mean": 144.365234375
          },
          "cpu": {
            "max": 0.6666666666666704,
            "min": 0.5999999999999991,
            "mean": 0.6388888888888891
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 29.41015625,
            "min": 24.44140625,
            "mean": 28.741145833333334
          },
          "cpu": {
            "max": 47.396353655013016,
            "min": 47.335301008786196,
            "mean": 47.36360580055597
          }
        }
      },
      "poolOverflow": 123,
      "upstreamConnections": 30,
      "counters": {
        "benchmark.http_2xx": {
          "value": 179877,
          "perSecond": 1998.6333333333334
        },
        "benchmark.pool_overflow": {
          "value": 123,
          "perSecond": 1.3666666666666667
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
          "value": 28240689,
          "perSecond": 313785.43333333335
        },
        "upstream_cx_total": {
          "value": 30,
          "perSecond": 0.3333333333333333
        },
        "upstream_cx_tx_bytes_total": {
          "value": 8094465,
          "perSecond": 89938.5
        },
        "upstream_rq_pending_overflow": {
          "value": 123,
          "perSecond": 1.3666666666666667
        },
        "upstream_rq_pending_total": {
          "value": 30,
          "perSecond": 0.3333333333333333
        },
        "upstream_rq_total": {
          "value": 179877,
          "perSecond": 1998.6333333333334
        }
      }
    },
    {
      "testName": "scaling up httproutes to 300 with 60 routes per hostname at 800 rps",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-up",
      "throughput": 3199.788888888889,
      "totalRequests": 287981,
      "latency": {
        "max": 42.504191,
        "min": 0.31223999999999996,
        "mean": 0.706733,
        "pstdev": 0.9379059999999999,
        "percentiles": {
          "p50": 0.5929909999999999,
          "p75": 0.654591,
          "p80": 0.691071,
          "p90": 0.843999,
          "p95": 1.095359,
          "p99": 2.4816629999999997,
          "p999": 16.161279
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 156.625,
            "min": 148.38671875,
            "mean": 153.39778645833334
          },
          "cpu": {
            "max": 12.799999999999986,
            "min": 0.7333333333333295,
            "mean": 2.9111111111111114
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 48.046875,
            "min": 40.9453125,
            "mean": 44.668359375
          },
          "cpu": {
            "max": 68.44566304658476,
            "min": 68.12625820568928,
            "mean": 68.33946768493264
          }
        }
      },
      "poolOverflow": 17,
      "upstreamConnections": 100,
      "counters": {
        "benchmark.http_2xx": {
          "value": 287981,
          "perSecond": 3199.788888888889
        },
        "benchmark.pool_overflow": {
          "value": 17,
          "perSecond": 0.18888888888888888
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
          "value": 100,
          "perSecond": 1.1111111111111112
        },
        "upstream_cx_rx_bytes_total": {
          "value": 45213017,
          "perSecond": 502366.85555555555
        },
        "upstream_cx_total": {
          "value": 100,
          "perSecond": 1.1111111111111112
        },
        "upstream_cx_tx_bytes_total": {
          "value": 12959235,
          "perSecond": 143991.5
        },
        "upstream_rq_pending_overflow": {
          "value": 17,
          "perSecond": 0.18888888888888888
        },
        "upstream_rq_pending_total": {
          "value": 100,
          "perSecond": 1.1111111111111112
        },
        "upstream_rq_total": {
          "value": 287983,
          "perSecond": 3199.811111111111
        }
      }
    },
    {
      "testName": "scaling up httproutes to 500 with 100 routes per hostname at 1000 rps",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-up",
      "throughput": 3998.8333333333335,
      "totalRequests": 359895,
      "latency": {
        "max": 147.996671,
        "min": 0.323872,
        "mean": 1.344672,
        "pstdev": 5.540702,
        "percentiles": {
          "p50": 0.620095,
          "p75": 0.799935,
          "p80": 0.860447,
          "p90": 1.158207,
          "p95": 1.928639,
          "p99": 21.194751,
          "p999": 82.444287
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 181.6953125,
            "min": 170.97265625,
            "mean": 176.509765625
          },
          "cpu": {
            "max": 14.06666666666671,
            "min": 0.8666666666666363,
            "mean": 3.393333347555553
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 74.30078125,
            "min": 57.40625,
            "mean": 64.9140625
          },
          "cpu": {
            "max": 82.77148470955721,
            "min": 47.87701941325333,
            "mean": 73.5128278378039
          }
        }
      },
      "poolOverflow": 102,
      "upstreamConnections": 298,
      "counters": {
        "benchmark.http_2xx": {
          "value": 359895,
          "perSecond": 3998.8333333333335
        },
        "benchmark.pool_overflow": {
          "value": 102,
          "perSecond": 1.1333333333333333
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
          "value": 298,
          "perSecond": 3.311111111111111
        },
        "upstream_cx_rx_bytes_total": {
          "value": 56503515,
          "perSecond": 627816.8333333334
        },
        "upstream_cx_total": {
          "value": 298,
          "perSecond": 3.311111111111111
        },
        "upstream_cx_tx_bytes_total": {
          "value": 16195410,
          "perSecond": 179949
        },
        "upstream_rq_pending_overflow": {
          "value": 102,
          "perSecond": 1.1333333333333333
        },
        "upstream_rq_pending_total": {
          "value": 298,
          "perSecond": 3.311111111111111
        },
        "upstream_rq_total": {
          "value": 359898,
          "perSecond": 3998.866666666667
        }
      }
    },
    {
      "testName": "scaling up httproutes to 1000 with 200 routes per hostname at 2000 rps",
      "routes": 1000,
      "routesPerHostname": 200,
      "phase": "scaling-up",
      "throughput": 5773.944444444444,
      "totalRequests": 519655,
      "latency": {
        "max": 1396.572159,
        "min": 2.3439360000000002,
        "mean": 55.933875,
        "pstdev": 17.828335,
        "percentiles": {
          "p50": 55.871486999999995,
          "p75": 63.182847,
          "p80": 64.331775,
          "p90": 67.108863,
          "p95": 69.894143,
          "p99": 87.130111,
          "p999": 137.38803099999998
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 262.515625,
            "min": 230.34375,
            "mean": 255.52760416666666
          },
          "cpu": {
            "max": 15.466666666667,
            "min": 0.9335822886102052,
            "mean": 3.5644077223480233
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 105.00390625,
            "min": 100.4375,
            "mean": 104.04557291666667
          },
          "cpu": {
            "max": 99.30173165984326,
            "min": 98.62741163675885,
            "mean": 98.98021125987295
          }
        }
      },
      "poolOverflow": 75,
      "upstreamConnections": 325,
      "counters": {
        "benchmark.http_2xx": {
          "value": 519655,
          "perSecond": 5773.944444444444
        },
        "benchmark.pool_overflow": {
          "value": 75,
          "perSecond": 0.8333333333333334
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
          "value": 325,
          "perSecond": 3.611111111111111
        },
        "upstream_cx_rx_bytes_total": {
          "value": 81585835,
          "perSecond": 906509.2777777778
        },
        "upstream_cx_total": {
          "value": 325,
          "perSecond": 3.611111111111111
        },
        "upstream_cx_tx_bytes_total": {
          "value": 23397975,
          "perSecond": 259977.5
        },
        "upstream_rq_pending_overflow": {
          "value": 75,
          "perSecond": 0.8333333333333334
        },
        "upstream_rq_pending_total": {
          "value": 325,
          "perSecond": 3.611111111111111
        },
        "upstream_rq_total": {
          "value": 519955,
          "perSecond": 5777.277777777777
        }
      }
    },
    {
      "testName": "scaling down httproutes to 500 with 100 routes per hostname at 1000 rps",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-down",
      "throughput": 3996.233333333333,
      "totalRequests": 359661,
      "latency": {
        "max": 164.241407,
        "min": 0.317312,
        "mean": 1.1612580000000001,
        "pstdev": 3.110231,
        "percentiles": {
          "p50": 0.666143,
          "p75": 0.900703,
          "p80": 1.002335,
          "p90": 1.5483509999999998,
          "p95": 2.750079,
          "p99": 10.365951,
          "p999": 46.364671
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 273.81640625,
            "min": 184.06640625,
            "mean": 203.72825520833334
          },
          "cpu": {
            "max": 25.27099820442915,
            "min": 0.9999999999998486,
            "mean": 6.734423708389329
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 105.640625,
            "min": 102.82421875,
            "mean": 104.92734375
          },
          "cpu": {
            "max": 82.48518576092052,
            "min": 80.61823333743985,
            "mean": 82.11847587184953
          }
        }
      },
      "poolOverflow": 338,
      "upstreamConnections": 62,
      "counters": {
        "benchmark.http_2xx": {
          "value": 359661,
          "perSecond": 3996.233333333333
        },
        "benchmark.pool_overflow": {
          "value": 338,
          "perSecond": 3.7555555555555555
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
          "value": 62,
          "perSecond": 0.6888888888888889
        },
        "upstream_cx_rx_bytes_total": {
          "value": 56466777,
          "perSecond": 627408.6333333333
        },
        "upstream_cx_total": {
          "value": 62,
          "perSecond": 0.6888888888888889
        },
        "upstream_cx_tx_bytes_total": {
          "value": 16184790,
          "perSecond": 179831
        },
        "upstream_rq_pending_overflow": {
          "value": 338,
          "perSecond": 3.7555555555555555
        },
        "upstream_rq_pending_total": {
          "value": 62,
          "perSecond": 0.6888888888888889
        },
        "upstream_rq_total": {
          "value": 359662,
          "perSecond": 3996.2444444444445
        }
      }
    },
    {
      "testName": "scaling down httproutes to 300 with 60 routes per hostname at 800 rps",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-down",
      "throughput": 3231.887640449438,
      "totalRequests": 287638,
      "latency": {
        "max": 143.441919,
        "min": 0.31872,
        "mean": 0.801465,
        "pstdev": 1.487045,
        "percentiles": {
          "p50": 0.649823,
          "p75": 0.781183,
          "p80": 0.8397749999999999,
          "p90": 1.045855,
          "p95": 1.383615,
          "p99": 3.3299190000000003,
          "p999": 11.514367
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 179.99609375,
            "min": 163.65234375,
            "mean": 171.115625
          },
          "cpu": {
            "max": 9.53524038140966,
            "min": 0.9999999999998485,
            "mean": 2.9743276191470143
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 105.27734375,
            "min": 104.69921875,
            "mean": 105.15533854166667
          },
          "cpu": {
            "max": 66.41848263787573,
            "min": 65.46887444336396,
            "mean": 66.03477159617294
          }
        }
      },
      "poolOverflow": 355,
      "upstreamConnections": 45,
      "counters": {
        "benchmark.http_2xx": {
          "value": 287638,
          "perSecond": 3231.887640449438
        },
        "benchmark.pool_overflow": {
          "value": 355,
          "perSecond": 3.9887640449438204
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
          "value": 45,
          "perSecond": 0.5056179775280899
        },
        "upstream_cx_rx_bytes_total": {
          "value": 45159166,
          "perSecond": 507406.3595505618
        },
        "upstream_cx_total": {
          "value": 45,
          "perSecond": 0.5056179775280899
        },
        "upstream_cx_tx_bytes_total": {
          "value": 12943845,
          "perSecond": 145436.4606741573
        },
        "upstream_rq_pending_overflow": {
          "value": 355,
          "perSecond": 3.9887640449438204
        },
        "upstream_rq_pending_total": {
          "value": 45,
          "perSecond": 0.5056179775280899
        },
        "upstream_rq_total": {
          "value": 287641,
          "perSecond": 3231.921348314607
        }
      }
    },
    {
      "testName": "scaling down httproutes to 100 with 20 routes per hostname at 500 rps",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-down",
      "throughput": 2020.056179775281,
      "totalRequests": 179785,
      "latency": {
        "max": 87.142399,
        "min": 0.315712,
        "mean": 0.542065,
        "pstdev": 0.8061929999999999,
        "percentiles": {
          "p50": 0.477087,
          "p75": 0.542879,
          "p80": 0.568415,
          "p90": 0.653119,
          "p95": 0.773695,
          "p99": 1.675199,
          "p999": 4.007295
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 173.77734375,
            "min": 146.9296875,
            "mean": 156.85013020833333
          },
          "cpu": {
            "max": 1.3333333333332575,
            "min": 1.1333333333330606,
            "mean": 1.2158730158729565
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 105.60546875,
            "min": 104.62890625,
            "mean": 105.17330729166666
          },
          "cpu": {
            "max": 44.28930005737239,
            "min": 23.724222343464607,
            "mean": 36.2524203725847
          }
        }
      },
      "poolOverflow": 214,
      "upstreamConnections": 37,
      "counters": {
        "benchmark.http_2xx": {
          "value": 179785,
          "perSecond": 2020.056179775281
        },
        "benchmark.pool_overflow": {
          "value": 214,
          "perSecond": 2.404494382022472
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
          "value": 37,
          "perSecond": 0.4157303370786517
        },
        "upstream_cx_rx_bytes_total": {
          "value": 28226245,
          "perSecond": 317148.8202247191
        },
        "upstream_cx_total": {
          "value": 37,
          "perSecond": 0.4157303370786517
        },
        "upstream_cx_tx_bytes_total": {
          "value": 8090325,
          "perSecond": 90902.52808988764
        },
        "upstream_rq_pending_overflow": {
          "value": 214,
          "perSecond": 2.404494382022472
        },
        "upstream_rq_pending_total": {
          "value": 37,
          "perSecond": 0.4157303370786517
        },
        "upstream_rq_total": {
          "value": 179785,
          "perSecond": 2020.056179775281
        }
      }
    },
    {
      "testName": "scaling down httproutes to 50 with 10 routes per hostname at 300 rps",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-down",
      "throughput": 1212.8988764044943,
      "totalRequests": 107948,
      "latency": {
        "max": 54.050815,
        "min": 0.351776,
        "mean": 0.490955,
        "pstdev": 0.555363,
        "percentiles": {
          "p50": 0.447391,
          "p75": 0.46265500000000004,
          "p80": 0.468607,
          "p90": 0.501391,
          "p95": 0.5712630000000001,
          "p99": 1.5144309999999999,
          "p999": 4.261631
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 158.01171875,
            "min": 149.8671875,
            "mean": 152.21119791666666
          },
          "cpu": {
            "max": 1.2666666666666515,
            "min": 1.0666666666668334,
            "mean": 1.1768115942029156
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 105.16015625,
            "min": 104.58203125,
            "mean": 104.906640625
          },
          "cpu": {
            "max": 29.542877352688034,
            "min": 18.159243472358067,
            "mean": 24.333210107585668
          }
        }
      },
      "poolOverflow": 52,
      "upstreamConnections": 22,
      "counters": {
        "benchmark.http_2xx": {
          "value": 107948,
          "perSecond": 1212.8988764044943
        },
        "benchmark.pool_overflow": {
          "value": 52,
          "perSecond": 0.5842696629213483
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
          "value": 16947836,
          "perSecond": 190425.1235955056
        },
        "upstream_cx_total": {
          "value": 22,
          "perSecond": 0.24719101123595505
        },
        "upstream_cx_tx_bytes_total": {
          "value": 4857660,
          "perSecond": 54580.449438202246
        },
        "upstream_rq_pending_overflow": {
          "value": 52,
          "perSecond": 0.5842696629213483
        },
        "upstream_rq_pending_total": {
          "value": 22,
          "perSecond": 0.24719101123595505
        },
        "upstream_rq_total": {
          "value": 107948,
          "perSecond": 1212.8988764044943
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
        "max": 26.066943,
        "min": 0.36608,
        "mean": 0.47985,
        "pstdev": 0.43344,
        "percentiles": {
          "p50": 0.43689500000000003,
          "p75": 0.44676699999999997,
          "p80": 0.452975,
          "p90": 0.48123099999999996,
          "p95": 0.5360309999999999,
          "p99": 1.473087,
          "p999": 5.483263
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 246.296875,
            "min": 152.87109375,
            "mean": 199.55924479166666
          },
          "cpu": {
            "max": 1.2666666666666515,
            "min": 1.0666666666668334,
            "mean": 1.1666666666666479
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 105.015625,
            "min": 104.625,
            "mean": 104.77864583333333
          },
          "cpu": {
            "max": 9.945251066623168,
            "min": 6.111743037219268,
            "mean": 8.656984557582316
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
