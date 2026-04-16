import { TestSuite } from '../types';

// Benchmark data extracted from release artifact for version 1.6.6
// Generated from benchmark_result.json

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.6.6",
    "runId": "1.6.6-release-2026-04-16",
    "date": "2026-04-16T14:08:03Z",
    "environment": "GitHub Release",
    "description": "Benchmark results for version 1.6.6 from release artifacts",
    "downloadUrl": "https://github.com/envoyproxy/gateway/releases/download/v1.6.6/benchmark_report.zip",
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
        "max": 33.880063,
        "min": 0.387376,
        "mean": 0.507736,
        "pstdev": 0.411868,
        "percentiles": {
          "p50": 0.470607,
          "p75": 0.49054299999999995,
          "p80": 0.49811099999999997,
          "p90": 0.5253429999999999,
          "p95": 0.569183,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 134.1796875,
            "min": 116.68359375,
            "mean": 132.01341145833334
          },
          "cpu": {
            "max": 1.266666666666667,
            "min": 0.33333333333333287,
            "mean": 0.5738095238095238
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 28.56640625,
            "min": 7.24609375,
            "mean": 26.650390625
          },
          "cpu": {
            "max": 10.481698907670713,
            "min": 10.198609044927105,
            "mean": 10.315512665874808
          }
        }
      },
      "poolOverflow": 5,
      "upstreamConnections": 9,
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
          "value": 9,
          "perSecond": 0.10112359550561797
        },
        "upstream_cx_rx_bytes_total": {
          "value": 5651215,
          "perSecond": 63496.79775280899
        },
        "upstream_cx_total": {
          "value": 9,
          "perSecond": 0.10112359550561797
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
          "value": 9,
          "perSecond": 0.10112359550561797
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
      "throughput": 1199.2333333333333,
      "totalRequests": 107931,
      "latency": {
        "max": 90.86975899999999,
        "min": 0.35481599999999996,
        "mean": 0.50866,
        "pstdev": 0.876099,
        "percentiles": {
          "p50": 0.457231,
          "p75": 0.476111,
          "p80": 0.481439,
          "p90": 0.506111,
          "p95": 0.5612790000000001,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 146.0625,
            "min": 134.296875,
            "mean": 142.66171875
          },
          "cpu": {
            "max": 2.066666666666667,
            "min": 0.46666666666666556,
            "mean": 0.8133333333333337
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 32.96875,
            "min": 28.46484375,
            "mean": 32.34921875
          },
          "cpu": {
            "max": 30.693209207037853,
            "min": 22.027248979189356,
            "mean": 29.453267114194706
          }
        }
      },
      "poolOverflow": 69,
      "upstreamConnections": 34,
      "counters": {
        "benchmark.http_2xx": {
          "value": 107931,
          "perSecond": 1199.2333333333333
        },
        "benchmark.pool_overflow": {
          "value": 69,
          "perSecond": 0.7666666666666667
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
          "value": 16945167,
          "perSecond": 188279.63333333333
        },
        "upstream_cx_total": {
          "value": 34,
          "perSecond": 0.37777777777777777
        },
        "upstream_cx_tx_bytes_total": {
          "value": 4856895,
          "perSecond": 53965.5
        },
        "upstream_rq_pending_overflow": {
          "value": 69,
          "perSecond": 0.7666666666666667
        },
        "upstream_rq_pending_total": {
          "value": 34,
          "perSecond": 0.37777777777777777
        },
        "upstream_rq_total": {
          "value": 107931,
          "perSecond": 1199.2333333333333
        }
      }
    },
    {
      "testName": "scaling up httproutes to 100 with 20 routes per hostname at 500 rps",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-up",
      "throughput": 2021.1348314606741,
      "totalRequests": 179881,
      "latency": {
        "max": 56.614911,
        "min": 0.353712,
        "mean": 0.516452,
        "pstdev": 0.659549,
        "percentiles": {
          "p50": 0.460671,
          "p75": 0.476447,
          "p80": 0.482431,
          "p90": 0.514799,
          "p95": 0.611711,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 148.16015625,
            "min": 137.08203125,
            "mean": 144.21393229166668
          },
          "cpu": {
            "max": 2.8666666666666676,
            "min": 0.4666666666666685,
            "mean": 1.0355555555555556
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 41.40234375,
            "min": 34.8359375,
            "mean": 38.9046875
          },
          "cpu": {
            "max": 49.46994247110364,
            "min": 49.11130632911394,
            "mean": 49.326488014307756
          }
        }
      },
      "poolOverflow": 119,
      "upstreamConnections": 49,
      "counters": {
        "benchmark.http_2xx": {
          "value": 179881,
          "perSecond": 2021.1348314606741
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
          "value": 49,
          "perSecond": 0.550561797752809
        },
        "upstream_cx_rx_bytes_total": {
          "value": 28241317,
          "perSecond": 317318.1685393258
        },
        "upstream_cx_total": {
          "value": 49,
          "perSecond": 0.550561797752809
        },
        "upstream_cx_tx_bytes_total": {
          "value": 8094645,
          "perSecond": 90951.06741573034
        },
        "upstream_rq_pending_overflow": {
          "value": 119,
          "perSecond": 1.3370786516853932
        },
        "upstream_rq_pending_total": {
          "value": 49,
          "perSecond": 0.550561797752809
        },
        "upstream_rq_total": {
          "value": 179881,
          "perSecond": 2021.1348314606741
        }
      }
    },
    {
      "testName": "scaling up httproutes to 300 with 60 routes per hostname at 800 rps",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-up",
      "throughput": 3234.921348314607,
      "totalRequests": 287908,
      "latency": {
        "max": 214.679551,
        "min": 0.33377599999999996,
        "mean": 1.347763,
        "pstdev": 5.953906,
        "percentiles": {
          "p50": 0.5782069999999999,
          "p75": 0.736575,
          "p80": 0.784127,
          "p90": 0.959999,
          "p95": 1.368447,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 157.796875,
            "min": 142.00390625,
            "mean": 153.67513020833334
          },
          "cpu": {
            "max": 4.933333333333323,
            "min": 0.7999999999999948,
            "mean": 1.5888888888888897
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 73.23828125,
            "min": 60.9609375,
            "mean": 69.82369791666666
          },
          "cpu": {
            "max": 78.84926882768706,
            "min": 39.2908866931036,
            "mean": 62.03672197900912
          }
        }
      },
      "poolOverflow": 91,
      "upstreamConnections": 270,
      "counters": {
        "benchmark.http_2xx": {
          "value": 287908,
          "perSecond": 3234.921348314607
        },
        "benchmark.pool_overflow": {
          "value": 91,
          "perSecond": 1.0224719101123596
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
          "value": 270,
          "perSecond": 3.033707865168539
        },
        "upstream_cx_rx_bytes_total": {
          "value": 45201556,
          "perSecond": 507882.65168539324
        },
        "upstream_cx_total": {
          "value": 270,
          "perSecond": 3.033707865168539
        },
        "upstream_cx_tx_bytes_total": {
          "value": 12955860,
          "perSecond": 145571.4606741573
        },
        "upstream_rq_pending_overflow": {
          "value": 91,
          "perSecond": 1.0224719101123596
        },
        "upstream_rq_pending_total": {
          "value": 270,
          "perSecond": 3.033707865168539
        },
        "upstream_rq_total": {
          "value": 287908,
          "perSecond": 3234.921348314607
        }
      }
    },
    {
      "testName": "scaling up httproutes to 500 with 100 routes per hostname at 1000 rps",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-up",
      "throughput": 3998.777777777778,
      "totalRequests": 359890,
      "latency": {
        "max": 245.112831,
        "min": 0.33624000000000004,
        "mean": 4.780056,
        "pstdev": 17.543810999999998,
        "percentiles": {
          "p50": 0.589215,
          "p75": 0.869631,
          "p80": 1.021663,
          "p90": 3.309439,
          "p95": 25.322495,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 161.43359375,
            "min": 156.67578125,
            "mean": 159.50208333333333
          },
          "cpu": {
            "max": 8.266666666666655,
            "min": 0.86666666666666,
            "mean": 2.1822222222222245
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 100.265625,
            "min": 91.65234375,
            "mean": 97.87916666666666
          },
          "cpu": {
            "max": 98.35603575528931,
            "min": 59.25775073715499,
            "mean": 84.76149181860085
          }
        }
      },
      "poolOverflow": 108,
      "upstreamConnections": 292,
      "counters": {
        "benchmark.http_2xx": {
          "value": 359890,
          "perSecond": 3998.777777777778
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
          "value": 56502730,
          "perSecond": 627808.1111111111
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
      "throughput": 4689.977777777778,
      "totalRequests": 422098,
      "latency": {
        "max": 248.496127,
        "min": 0.398128,
        "mean": 59.083398,
        "pstdev": 39.084657,
        "percentiles": {
          "p50": 70.328319,
          "p75": 87.94931100000001,
          "p80": 92.20505499999999,
          "p90": 101.740543,
          "p95": 110.870527,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 187.17578125,
            "min": 174.92578125,
            "mean": 183.34283854166668
          },
          "cpu": {
            "max": 14.933333333333392,
            "min": 0.8666666666667312,
            "mean": 3.491111111111108
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 163.9296875,
            "min": 150.73046875,
            "mean": 158.60247395833332
          },
          "cpu": {
            "max": 69.85045922786442,
            "min": 69.85045922786442,
            "mean": 69.85045922786442
          }
        }
      },
      "poolOverflow": 113,
      "upstreamConnections": 287,
      "counters": {
        "benchmark.http_2xx": {
          "value": 422098,
          "perSecond": 4689.977777777778
        },
        "benchmark.pool_overflow": {
          "value": 113,
          "perSecond": 1.2555555555555555
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
          "value": 287,
          "perSecond": 3.188888888888889
        },
        "upstream_cx_rx_bytes_total": {
          "value": 66269386,
          "perSecond": 736326.5111111111
        },
        "upstream_cx_total": {
          "value": 287,
          "perSecond": 3.188888888888889
        },
        "upstream_cx_tx_bytes_total": {
          "value": 18998505,
          "perSecond": 211094.5
        },
        "upstream_rq_pending_overflow": {
          "value": 113,
          "perSecond": 1.2555555555555555
        },
        "upstream_rq_pending_total": {
          "value": 287,
          "perSecond": 3.188888888888889
        },
        "upstream_rq_total": {
          "value": 422189,
          "perSecond": 4690.988888888889
        }
      }
    },
    {
      "testName": "scaling down httproutes to 500 with 100 routes per hostname at 1000 rps",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-down",
      "throughput": 3999.288888888889,
      "totalRequests": 359936,
      "latency": {
        "max": 458.833919,
        "min": 0.331216,
        "mean": 4.969953,
        "pstdev": 19.039579,
        "percentiles": {
          "p50": 0.596607,
          "p75": 0.886719,
          "p80": 1.034399,
          "p90": 2.897279,
          "p95": 24.657919,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 174.984375,
            "min": 162.4921875,
            "mean": 167.33841145833333
          },
          "cpu": {
            "max": 1.2659914712153366,
            "min": 1.13333333333325,
            "mean": 1.1709512339401278
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 156.515625,
            "min": 154.5625,
            "mean": 155.71731770833333
          },
          "cpu": {
            "max": 97.94922054530775,
            "min": 94.7142316561226,
            "mean": 95.79061787275042
          }
        }
      },
      "poolOverflow": 61,
      "upstreamConnections": 339,
      "counters": {
        "benchmark.http_2xx": {
          "value": 359936,
          "perSecond": 3999.288888888889
        },
        "benchmark.pool_overflow": {
          "value": 61,
          "perSecond": 0.6777777777777778
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
          "value": 339,
          "perSecond": 3.7666666666666666
        },
        "upstream_cx_rx_bytes_total": {
          "value": 56509952,
          "perSecond": 627888.3555555556
        },
        "upstream_cx_total": {
          "value": 339,
          "perSecond": 3.7666666666666666
        },
        "upstream_cx_tx_bytes_total": {
          "value": 16197255,
          "perSecond": 179969.5
        },
        "upstream_rq_pending_overflow": {
          "value": 61,
          "perSecond": 0.6777777777777778
        },
        "upstream_rq_pending_total": {
          "value": 339,
          "perSecond": 3.7666666666666666
        },
        "upstream_rq_total": {
          "value": 359939,
          "perSecond": 3999.322222222222
        }
      }
    },
    {
      "testName": "scaling down httproutes to 300 with 60 routes per hostname at 800 rps",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-down",
      "throughput": 3198.5777777777776,
      "totalRequests": 287872,
      "latency": {
        "max": 92.098559,
        "min": 0.3348,
        "mean": 1.216339,
        "pstdev": 5.12706,
        "percentiles": {
          "p50": 0.5432629999999999,
          "p75": 0.6351669999999999,
          "p80": 0.684223,
          "p90": 0.906527,
          "p95": 1.413311,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 166.23828125,
            "min": 160.7265625,
            "mean": 162.56966145833334
          },
          "cpu": {
            "max": 1.2000000000000455,
            "min": 1.066666666666644,
            "mean": 1.1166666666666698
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 154.7578125,
            "min": 150.59765625,
            "mean": 152.107421875
          },
          "cpu": {
            "max": 80.57323272090986,
            "min": 80.57323272090986,
            "mean": 80.57323272090986
          }
        }
      },
      "poolOverflow": 128,
      "upstreamConnections": 235,
      "counters": {
        "benchmark.http_2xx": {
          "value": 287872,
          "perSecond": 3198.5777777777776
        },
        "benchmark.pool_overflow": {
          "value": 128,
          "perSecond": 1.4222222222222223
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
          "value": 235,
          "perSecond": 2.611111111111111
        },
        "upstream_cx_rx_bytes_total": {
          "value": 45195904,
          "perSecond": 502176.7111111111
        },
        "upstream_cx_total": {
          "value": 235,
          "perSecond": 2.611111111111111
        },
        "upstream_cx_tx_bytes_total": {
          "value": 12954240,
          "perSecond": 143936
        },
        "upstream_rq_pending_overflow": {
          "value": 128,
          "perSecond": 1.4222222222222223
        },
        "upstream_rq_pending_total": {
          "value": 235,
          "perSecond": 2.611111111111111
        },
        "upstream_rq_total": {
          "value": 287872,
          "perSecond": 3198.5777777777776
        }
      }
    },
    {
      "testName": "scaling down httproutes to 100 with 20 routes per hostname at 500 rps",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-down",
      "throughput": 2019.629213483146,
      "totalRequests": 179747,
      "latency": {
        "max": 96.550911,
        "min": 0.349776,
        "mean": 0.577981,
        "pstdev": 1.2729519999999999,
        "percentiles": {
          "p50": 0.474671,
          "p75": 0.518367,
          "p80": 0.539359,
          "p90": 0.611071,
          "p95": 0.763871,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 157.94140625,
            "min": 148.09765625,
            "mean": 150.53111979166667
          },
          "cpu": {
            "max": 1.2000000000000457,
            "min": 1.066666666666644,
            "mean": 1.1391304347826228
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 151.9296875,
            "min": 150.359375,
            "mean": 151.08776041666667
          },
          "cpu": {
            "max": 51.70564292946298,
            "min": 50.91727267696056,
            "mean": 51.180062761128035
          }
        }
      },
      "poolOverflow": 253,
      "upstreamConnections": 62,
      "counters": {
        "benchmark.http_2xx": {
          "value": 179747,
          "perSecond": 2019.629213483146
        },
        "benchmark.pool_overflow": {
          "value": 253,
          "perSecond": 2.842696629213483
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
          "value": 62,
          "perSecond": 0.6966292134831461
        },
        "upstream_cx_rx_bytes_total": {
          "value": 28220279,
          "perSecond": 317081.78651685396
        },
        "upstream_cx_total": {
          "value": 62,
          "perSecond": 0.6966292134831461
        },
        "upstream_cx_tx_bytes_total": {
          "value": 8088615,
          "perSecond": 90883.31460674158
        },
        "upstream_rq_pending_overflow": {
          "value": 253,
          "perSecond": 2.842696629213483
        },
        "upstream_rq_pending_total": {
          "value": 62,
          "perSecond": 0.6966292134831461
        },
        "upstream_rq_total": {
          "value": 179747,
          "perSecond": 2019.629213483146
        }
      }
    },
    {
      "testName": "scaling down httproutes to 50 with 10 routes per hostname at 300 rps",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-down",
      "throughput": 1199.4888888888888,
      "totalRequests": 107954,
      "latency": {
        "max": 52.932607,
        "min": 0.366016,
        "mean": 0.513018,
        "pstdev": 0.546949,
        "percentiles": {
          "p50": 0.46732700000000005,
          "p75": 0.48377499999999996,
          "p80": 0.48961499999999997,
          "p90": 0.515983,
          "p95": 0.569087,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 152.47265625,
            "min": 142.44921875,
            "mean": 148.01822916666666
          },
          "cpu": {
            "max": 1.2666666666666515,
            "min": 1.066666666666644,
            "mean": 1.1666666666666716
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 151.0078125,
            "min": 149.953125,
            "mean": 150.43489583333334
          },
          "cpu": {
            "max": 31.921542598843654,
            "min": 18.91236776541521,
            "mean": 25.88042318797759
          }
        }
      },
      "poolOverflow": 46,
      "upstreamConnections": 32,
      "counters": {
        "benchmark.http_2xx": {
          "value": 107954,
          "perSecond": 1199.4888888888888
        },
        "benchmark.pool_overflow": {
          "value": 46,
          "perSecond": 0.5111111111111111
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
          "value": 32,
          "perSecond": 0.35555555555555557
        },
        "upstream_cx_rx_bytes_total": {
          "value": 16948778,
          "perSecond": 188319.75555555554
        },
        "upstream_cx_total": {
          "value": 32,
          "perSecond": 0.35555555555555557
        },
        "upstream_cx_tx_bytes_total": {
          "value": 4857930,
          "perSecond": 53977
        },
        "upstream_rq_pending_overflow": {
          "value": 46,
          "perSecond": 0.5111111111111111
        },
        "upstream_rq_pending_total": {
          "value": 32,
          "perSecond": 0.35555555555555557
        },
        "upstream_rq_total": {
          "value": 107954,
          "perSecond": 1199.4888888888888
        }
      }
    },
    {
      "testName": "scaling down httproutes to 10 with 2 routes per hostname at 100 rps",
      "routes": 10,
      "routesPerHostname": 2,
      "phase": "scaling-down",
      "throughput": 404.1123595505618,
      "totalRequests": 35966,
      "latency": {
        "max": 111.333375,
        "min": 0.38016,
        "mean": 0.545542,
        "pstdev": 1.075619,
        "percentiles": {
          "p50": 0.485487,
          "p75": 0.527999,
          "p80": 0.539679,
          "p90": 0.573791,
          "p95": 0.621375,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 145.6328125,
            "min": 143.19921875,
            "mean": 144.47630208333334
          },
          "cpu": {
            "max": 1.2666666666666517,
            "min": 1.066666666666644,
            "mean": 1.1128344152949756
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 150.8671875,
            "min": 150.2421875,
            "mean": 150.51067708333332
          },
          "cpu": {
            "max": 12.051869716020317,
            "min": 10.812681257807123,
            "mean": 11.105701068624029
          }
        }
      },
      "poolOverflow": 34,
      "upstreamConnections": 14,
      "counters": {
        "benchmark.http_2xx": {
          "value": 35966,
          "perSecond": 404.1123595505618
        },
        "benchmark.pool_overflow": {
          "value": 34,
          "perSecond": 0.38202247191011235
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
          "value": 14,
          "perSecond": 0.15730337078651685
        },
        "upstream_cx_rx_bytes_total": {
          "value": 5646662,
          "perSecond": 63445.6404494382
        },
        "upstream_cx_total": {
          "value": 14,
          "perSecond": 0.15730337078651685
        },
        "upstream_cx_tx_bytes_total": {
          "value": 1618470,
          "perSecond": 18185.05617977528
        },
        "upstream_rq_pending_overflow": {
          "value": 34,
          "perSecond": 0.38202247191011235
        },
        "upstream_rq_pending_total": {
          "value": 14,
          "perSecond": 0.15730337078651685
        },
        "upstream_rq_total": {
          "value": 35966,
          "perSecond": 404.1123595505618
        }
      }
    }
  ]
};
