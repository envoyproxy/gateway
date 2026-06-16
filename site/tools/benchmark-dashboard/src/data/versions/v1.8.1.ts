import { TestSuite } from '../types';

// Benchmark data extracted from release artifact for version 1.8.1
// Generated from benchmark_result.json

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.8.1",
    "runId": "1.8.1-release-2026-06-05",
    "date": "2026-06-05T07:32:33Z",
    "environment": "GitHub Release",
    "description": "Benchmark results for version 1.8.1 from release artifacts",
    "downloadUrl": "https://github.com/envoyproxy/gateway/releases/download/v1.8.1/benchmark_report.zip",
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
      "throughput": 404.40449438202245,
      "totalRequests": 35992,
      "latency": {
        "max": 109.498367,
        "min": 0.38012799999999997,
        "mean": 0.515294,
        "pstdev": 0.980444,
        "percentiles": {
          "p50": 0.464495,
          "p75": 0.48273499999999997,
          "p80": 0.487215,
          "p90": 0.503199,
          "p95": 0.534847,
          "p99": 1.253183,
          "p999": 10.898431
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 131.48046875,
            "min": 109.91796875,
            "mean": 125.279296875
          },
          "cpu": {
            "max": 0.9333333333333335,
            "min": 0.3333333333333336,
            "mean": 0.5111111111111112
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 21.76171875,
            "min": 7.1015625,
            "mean": 18.87365301724138
          },
          "cpu": {
            "max": 10.554019329822324,
            "min": 7.540699450203251,
            "mean": 10.166335275663474
          }
        }
      },
      "poolOverflow": 8,
      "upstreamConnections": 12,
      "counters": {
        "benchmark.http_2xx": {
          "value": 35992,
          "perSecond": 404.40449438202245
        },
        "benchmark.pool_overflow": {
          "value": 8,
          "perSecond": 0.0898876404494382
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
          "value": 5650744,
          "perSecond": 63491.50561797753
        },
        "upstream_cx_total": {
          "value": 12,
          "perSecond": 0.1348314606741573
        },
        "upstream_cx_tx_bytes_total": {
          "value": 1619640,
          "perSecond": 18198.20224719101
        },
        "upstream_rq_pending_overflow": {
          "value": 8,
          "perSecond": 0.0898876404494382
        },
        "upstream_rq_pending_total": {
          "value": 12,
          "perSecond": 0.1348314606741573
        },
        "upstream_rq_total": {
          "value": 35992,
          "perSecond": 404.40449438202245
        }
      }
    },
    {
      "testName": "scaling up httproutes to 50 with 10 routes per hostname at 300 rps",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-up",
      "throughput": 1212.9550561797753,
      "totalRequests": 107953,
      "latency": {
        "max": 83.43961499999999,
        "min": 0.33288,
        "mean": 0.49536899999999995,
        "pstdev": 0.762126,
        "percentiles": {
          "p50": 0.44383900000000004,
          "p75": 0.462367,
          "p80": 0.466623,
          "p90": 0.485423,
          "p95": 0.5404150000000001,
          "p99": 1.544191,
          "p999": 9.935359
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 133.84765625,
            "min": 127.07421875,
            "mean": 131.03697916666667
          },
          "cpu": {
            "max": 0.5999999999999991,
            "min": 0.40000000000000036,
            "mean": 0.5111111111111107
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 25.69921875,
            "min": 21.16015625,
            "mean": 25.233854166666667
          },
          "cpu": {
            "max": 29.37750977835723,
            "min": 29.288211204542968,
            "mean": 29.352440857053452
          }
        }
      },
      "poolOverflow": 47,
      "upstreamConnections": 36,
      "counters": {
        "benchmark.http_2xx": {
          "value": 107953,
          "perSecond": 1212.9550561797753
        },
        "benchmark.pool_overflow": {
          "value": 47,
          "perSecond": 0.5280898876404494
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
          "value": 16948621,
          "perSecond": 190433.94382022473
        },
        "upstream_cx_total": {
          "value": 36,
          "perSecond": 0.4044943820224719
        },
        "upstream_cx_tx_bytes_total": {
          "value": 4857885,
          "perSecond": 54582.97752808989
        },
        "upstream_rq_pending_overflow": {
          "value": 47,
          "perSecond": 0.5280898876404494
        },
        "upstream_rq_pending_total": {
          "value": 36,
          "perSecond": 0.4044943820224719
        },
        "upstream_rq_total": {
          "value": 107953,
          "perSecond": 1212.9550561797753
        }
      }
    },
    {
      "testName": "scaling up httproutes to 100 with 20 routes per hostname at 500 rps",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-up",
      "throughput": 2021.4943820224719,
      "totalRequests": 179913,
      "latency": {
        "max": 45.336574999999996,
        "min": 0.276032,
        "mean": 0.5828209999999999,
        "pstdev": 0.817276,
        "percentiles": {
          "p50": 0.487327,
          "p75": 0.549631,
          "p80": 0.5931190000000001,
          "p90": 0.738015,
          "p95": 0.8929269999999999,
          "p99": 2.2096630000000004,
          "p999": 10.760703
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 139.8203125,
            "min": 130.8671875,
            "mean": 134.37369791666666
          },
          "cpu": {
            "max": 1.7333333333333318,
            "min": 0.5333333333333339,
            "mean": 0.8266666666666667
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 30.21875,
            "min": 25.56640625,
            "mean": 29.29453125
          },
          "cpu": {
            "max": 43.00184937340304,
            "min": 42.62349434320169,
            "mean": 42.807061855525625
          }
        }
      },
      "poolOverflow": 85,
      "upstreamConnections": 76,
      "counters": {
        "benchmark.http_2xx": {
          "value": 179913,
          "perSecond": 2021.4943820224719
        },
        "benchmark.pool_overflow": {
          "value": 85,
          "perSecond": 0.9550561797752809
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
          "value": 76,
          "perSecond": 0.8539325842696629
        },
        "upstream_cx_rx_bytes_total": {
          "value": 28246341,
          "perSecond": 317374.6179775281
        },
        "upstream_cx_total": {
          "value": 76,
          "perSecond": 0.8539325842696629
        },
        "upstream_cx_tx_bytes_total": {
          "value": 8096130,
          "perSecond": 90967.75280898876
        },
        "upstream_rq_pending_overflow": {
          "value": 85,
          "perSecond": 0.9550561797752809
        },
        "upstream_rq_pending_total": {
          "value": 76,
          "perSecond": 0.8539325842696629
        },
        "upstream_rq_total": {
          "value": 179914,
          "perSecond": 2021.5056179775281
        }
      }
    },
    {
      "testName": "scaling up httproutes to 300 with 60 routes per hostname at 800 rps",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-up",
      "throughput": 3199.2,
      "totalRequests": 287928,
      "latency": {
        "max": 40.372223,
        "min": 0.2696,
        "mean": 0.760329,
        "pstdev": 0.8954399999999999,
        "percentiles": {
          "p50": 0.578367,
          "p75": 0.648383,
          "p80": 0.689471,
          "p90": 1.035039,
          "p95": 1.6718709999999999,
          "p99": 3.783935,
          "p999": 13.377535
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 153.76171875,
            "min": 144.046875,
            "mean": 150.990234375
          },
          "cpu": {
            "max": 32.06666666666667,
            "min": 0.6666666666666643,
            "mean": 5.988888893629631
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 46.6875,
            "min": 33.8515625,
            "mean": 43.86145833333333
          },
          "cpu": {
            "max": 65.97595594217421,
            "min": 65.81858578052547,
            "mean": 65.89727086134984
          }
        }
      },
      "poolOverflow": 72,
      "upstreamConnections": 87,
      "counters": {
        "benchmark.http_2xx": {
          "value": 287928,
          "perSecond": 3199.2
        },
        "benchmark.pool_overflow": {
          "value": 72,
          "perSecond": 0.8
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
          "value": 87,
          "perSecond": 0.9666666666666667
        },
        "upstream_cx_rx_bytes_total": {
          "value": 45204696,
          "perSecond": 502274.4
        },
        "upstream_cx_total": {
          "value": 87,
          "perSecond": 0.9666666666666667
        },
        "upstream_cx_tx_bytes_total": {
          "value": 12956760,
          "perSecond": 143964
        },
        "upstream_rq_pending_overflow": {
          "value": 72,
          "perSecond": 0.8
        },
        "upstream_rq_pending_total": {
          "value": 87,
          "perSecond": 0.9666666666666667
        },
        "upstream_rq_total": {
          "value": 287928,
          "perSecond": 3199.2
        }
      }
    },
    {
      "testName": "scaling up httproutes to 500 with 100 routes per hostname at 1000 rps",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-up",
      "throughput": 4043.4606741573034,
      "totalRequests": 359868,
      "latency": {
        "max": 107.41350299999999,
        "min": 0.253432,
        "mean": 1.37783,
        "pstdev": 4.171488999999999,
        "percentiles": {
          "p50": 0.690591,
          "p75": 0.9181429999999999,
          "p80": 1.038527,
          "p90": 2.045887,
          "p95": 3.422079,
          "p99": 12.497919,
          "p999": 66.467839
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 161.8359375,
            "min": 146.98046875,
            "mean": 160.31276041666666
          },
          "cpu": {
            "max": 49.13333333333334,
            "min": 0.86666666666666,
            "mean": 9.222988505747129
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 70.37890625,
            "min": 48.23046875,
            "mean": 62.147526041666666
          },
          "cpu": {
            "max": 76.39781104541453,
            "min": 75.68597611597177,
            "mean": 75.94452978166471
          }
        }
      },
      "poolOverflow": 119,
      "upstreamConnections": 281,
      "counters": {
        "benchmark.http_2xx": {
          "value": 359868,
          "perSecond": 4043.4606741573034
        },
        "benchmark.pool_overflow": {
          "value": 119,
          "perSecond": 1.3370786516853932
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
          "value": 281,
          "perSecond": 3.157303370786517
        },
        "upstream_cx_rx_bytes_total": {
          "value": 56499276,
          "perSecond": 634823.3258426966
        },
        "upstream_cx_total": {
          "value": 281,
          "perSecond": 3.157303370786517
        },
        "upstream_cx_tx_bytes_total": {
          "value": 16194600,
          "perSecond": 181961.79775280898
        },
        "upstream_rq_pending_overflow": {
          "value": 119,
          "perSecond": 1.3370786516853932
        },
        "upstream_rq_pending_total": {
          "value": 281,
          "perSecond": 3.157303370786517
        },
        "upstream_rq_total": {
          "value": 359880,
          "perSecond": 4043.5955056179773
        }
      }
    },
    {
      "testName": "scaling up httproutes to 1000 with 200 routes per hostname at 2000 rps",
      "routes": 1000,
      "routesPerHostname": 200,
      "phase": "scaling-up",
      "throughput": 6508.044444444445,
      "totalRequests": 585724,
      "latency": {
        "max": 705.036287,
        "min": 1.559616,
        "mean": 45.79942,
        "pstdev": 13.919675,
        "percentiles": {
          "p50": 44.763135000000005,
          "p75": 52.000767,
          "p80": 53.450751,
          "p90": 57.329663000000004,
          "p95": 61.442047,
          "p99": 87.625727,
          "p999": 156.680191
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 189.63671875,
            "min": 175.828125,
            "mean": 182.88255208333334
          },
          "cpu": {
            "max": 96.26666666666665,
            "min": 0.8666666666666363,
            "mean": 15.506662238222047
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 105.33203125,
            "min": 89.0859375,
            "mean": 100.22604166666666
          },
          "cpu": {
            "max": 98.10031793745301,
            "min": 98.100317937453,
            "mean": 98.100317937453
          }
        }
      },
      "poolOverflow": 99,
      "upstreamConnections": 301,
      "counters": {
        "benchmark.http_2xx": {
          "value": 585724,
          "perSecond": 6508.044444444445
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
          "value": 301,
          "perSecond": 3.3444444444444446
        },
        "upstream_cx_rx_bytes_total": {
          "value": 91958668,
          "perSecond": 1021762.9777777778
        },
        "upstream_cx_total": {
          "value": 301,
          "perSecond": 3.3444444444444446
        },
        "upstream_cx_tx_bytes_total": {
          "value": 26370360,
          "perSecond": 293004
        },
        "upstream_rq_pending_overflow": {
          "value": 99,
          "perSecond": 1.1
        },
        "upstream_rq_pending_total": {
          "value": 301,
          "perSecond": 3.3444444444444446
        },
        "upstream_rq_total": {
          "value": 586008,
          "perSecond": 6511.2
        }
      }
    },
    {
      "testName": "scaling down httproutes to 500 with 100 routes per hostname at 1000 rps",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-down",
      "throughput": 3999.0555555555557,
      "totalRequests": 359915,
      "latency": {
        "max": 141.844479,
        "min": 0.262272,
        "mean": 1.5387339999999998,
        "pstdev": 4.385846,
        "percentiles": {
          "p50": 0.648831,
          "p75": 0.982335,
          "p80": 1.1812470000000002,
          "p90": 2.536063,
          "p95": 4.634111,
          "p99": 18.376703000000003,
          "p999": 65.419263
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 185.24609375,
            "min": 163.96875,
            "mean": 170.17838541666666
          },
          "cpu": {
            "max": 1.3333333333333524,
            "min": 1.0000000000000377,
            "mean": 1.1199573418649655
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 104.74609375,
            "min": 103.33203125,
            "mean": 104.26653645833333
          },
          "cpu": {
            "max": 78.74553133975333,
            "min": 54.886713963210774,
            "mean": 66.81612265148205
          }
        }
      },
      "poolOverflow": 81,
      "upstreamConnections": 296,
      "counters": {
        "benchmark.http_2xx": {
          "value": 359915,
          "perSecond": 3999.0555555555557
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
          "value": 296,
          "perSecond": 3.2888888888888888
        },
        "upstream_cx_rx_bytes_total": {
          "value": 56506655,
          "perSecond": 627851.7222222222
        },
        "upstream_cx_total": {
          "value": 296,
          "perSecond": 3.2888888888888888
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
          "value": 296,
          "perSecond": 3.2888888888888888
        },
        "upstream_rq_total": {
          "value": 359919,
          "perSecond": 3999.1
        }
      }
    },
    {
      "testName": "scaling down httproutes to 300 with 60 routes per hostname at 800 rps",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-down",
      "throughput": 3199.4555555555557,
      "totalRequests": 287951,
      "latency": {
        "max": 55.562239,
        "min": 0.27488,
        "mean": 1.172193,
        "pstdev": 1.6265640000000001,
        "percentiles": {
          "p50": 0.603935,
          "p75": 1.207871,
          "p80": 1.733567,
          "p90": 2.645375,
          "p95": 3.0668789999999997,
          "p99": 6.285823000000001,
          "p999": 20.894719
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 160.8046875,
            "min": 149.265625,
            "mean": 154.90598958333334
          },
          "cpu": {
            "max": 1.266244585138272,
            "min": 0.9336445481827313,
            "mean": 1.096944499997193
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 104.10546875,
            "min": 103.3125,
            "mean": 103.80494791666666
          },
          "cpu": {
            "max": 66.13249527856466,
            "min": 64.65317981276254,
            "mean": 65.46358015461453
          }
        }
      },
      "poolOverflow": 48,
      "upstreamConnections": 123,
      "counters": {
        "benchmark.http_2xx": {
          "value": 287951,
          "perSecond": 3199.4555555555557
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
          "value": 123,
          "perSecond": 1.3666666666666667
        },
        "upstream_cx_rx_bytes_total": {
          "value": 45208307,
          "perSecond": 502314.52222222224
        },
        "upstream_cx_total": {
          "value": 123,
          "perSecond": 1.3666666666666667
        },
        "upstream_cx_tx_bytes_total": {
          "value": 12957840,
          "perSecond": 143976
        },
        "upstream_rq_pending_overflow": {
          "value": 48,
          "perSecond": 0.5333333333333333
        },
        "upstream_rq_pending_total": {
          "value": 123,
          "perSecond": 1.3666666666666667
        },
        "upstream_rq_total": {
          "value": 287952,
          "perSecond": 3199.4666666666667
        }
      }
    },
    {
      "testName": "scaling down httproutes to 100 with 20 routes per hostname at 500 rps",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-down",
      "throughput": 2021.2808988764045,
      "totalRequests": 179894,
      "latency": {
        "max": 63.080447,
        "min": 0.27105599999999996,
        "mean": 0.6560739999999999,
        "pstdev": 1.06362,
        "percentiles": {
          "p50": 0.548959,
          "p75": 0.6565749999999999,
          "p80": 0.690783,
          "p90": 0.823327,
          "p95": 0.9779190000000001,
          "p99": 2.508799,
          "p999": 17.105919
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 162.5546875,
            "min": 146.77734375,
            "mean": 151.49283854166666
          },
          "cpu": {
            "max": 1.066666666666644,
            "min": 0.933333333333337,
            "mean": 1.0166666666666657
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 103.87109375,
            "min": 103.01953125,
            "mean": 103.50442708333334
          },
          "cpu": {
            "max": 41.79160047230567,
            "min": 21.789793713683935,
            "mean": 31.509512905644222
          }
        }
      },
      "poolOverflow": 104,
      "upstreamConnections": 66,
      "counters": {
        "benchmark.http_2xx": {
          "value": 179894,
          "perSecond": 2021.2808988764045
        },
        "benchmark.pool_overflow": {
          "value": 104,
          "perSecond": 1.1685393258426966
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
          "value": 66,
          "perSecond": 0.7415730337078652
        },
        "upstream_cx_rx_bytes_total": {
          "value": 28243358,
          "perSecond": 317341.1011235955
        },
        "upstream_cx_total": {
          "value": 66,
          "perSecond": 0.7415730337078652
        },
        "upstream_cx_tx_bytes_total": {
          "value": 8095230,
          "perSecond": 90957.6404494382
        },
        "upstream_rq_pending_overflow": {
          "value": 104,
          "perSecond": 1.1685393258426966
        },
        "upstream_rq_pending_total": {
          "value": 66,
          "perSecond": 0.7415730337078652
        },
        "upstream_rq_total": {
          "value": 179894,
          "perSecond": 2021.2808988764045
        }
      }
    },
    {
      "testName": "scaling down httproutes to 50 with 10 routes per hostname at 300 rps",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-down",
      "throughput": 1198.9222222222222,
      "totalRequests": 107903,
      "latency": {
        "max": 97.869823,
        "min": 0.34808,
        "mean": 0.5089359999999999,
        "pstdev": 1.073334,
        "percentiles": {
          "p50": 0.442463,
          "p75": 0.462575,
          "p80": 0.467695,
          "p90": 0.486479,
          "p95": 0.5455990000000001,
          "p99": 1.8232949999999999,
          "p999": 10.969599
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 150.85546875,
            "min": 142.01953125,
            "mean": 144.434375
          },
          "cpu": {
            "max": 1.1999999999999507,
            "min": 0.933333333333337,
            "mean": 1.0606060606060561
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 103.63671875,
            "min": 103.2734375,
            "mean": 103.49453125
          },
          "cpu": {
            "max": 29.558607972838676,
            "min": 19.90093760440843,
            "mean": 27.73225719409582
          }
        }
      },
      "poolOverflow": 97,
      "upstreamConnections": 42,
      "counters": {
        "benchmark.http_2xx": {
          "value": 107903,
          "perSecond": 1198.9222222222222
        },
        "benchmark.pool_overflow": {
          "value": 97,
          "perSecond": 1.0777777777777777
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
          "value": 16940771,
          "perSecond": 188230.7888888889
        },
        "upstream_cx_total": {
          "value": 42,
          "perSecond": 0.4666666666666667
        },
        "upstream_cx_tx_bytes_total": {
          "value": 4855635,
          "perSecond": 53951.5
        },
        "upstream_rq_pending_overflow": {
          "value": 97,
          "perSecond": 1.0777777777777777
        },
        "upstream_rq_pending_total": {
          "value": 42,
          "perSecond": 0.4666666666666667
        },
        "upstream_rq_total": {
          "value": 107903,
          "perSecond": 1198.9222222222222
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
        "max": 27.911167,
        "min": 0.371904,
        "mean": 0.497451,
        "pstdev": 0.427144,
        "percentiles": {
          "p50": 0.466335,
          "p75": 0.482623,
          "p80": 0.486383,
          "p90": 0.49972700000000003,
          "p95": 0.530815,
          "p99": 1.201855,
          "p999": 5.028607
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 145.3046875,
            "min": 141.98046875,
            "mean": 143.59661458333332
          },
          "cpu": {
            "max": 1.1999999999999507,
            "min": 0.933333333333337,
            "mean": 1.03055555555555
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 103.58203125,
            "min": 103.0703125,
            "mean": 103.26497395833333
          },
          "cpu": {
            "max": 10.618555732253945,
            "min": 5.543974256020859,
            "mean": 9.807505359943042
          }
        }
      },
      "poolOverflow": 9,
      "upstreamConnections": 13,
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
          "value": 13,
          "perSecond": 0.14606741573033707
        },
        "upstream_cx_rx_bytes_total": {
          "value": 5650587,
          "perSecond": 63489.74157303371
        },
        "upstream_cx_total": {
          "value": 13,
          "perSecond": 0.14606741573033707
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
          "value": 13,
          "perSecond": 0.14606741573033707
        },
        "upstream_rq_total": {
          "value": 35991,
          "perSecond": 404.39325842696627
        }
      }
    }
  ]
};
