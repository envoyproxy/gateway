import { TestSuite } from '../types';

// Benchmark data extracted from release artifact for version 1.7.4
// Generated from benchmark_result.json

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.7.4",
    "runId": "1.7.4-release-2026-06-05",
    "date": "2026-06-05T07:55:33Z",
    "environment": "GitHub Release",
    "description": "Benchmark results for version 1.7.4 from release artifacts",
    "downloadUrl": "https://github.com/envoyproxy/gateway/releases/download/v1.7.4/benchmark_report.zip",
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
      "throughput": 399.9888888888889,
      "totalRequests": 35999,
      "latency": {
        "max": 33.251326999999996,
        "min": 0.308688,
        "mean": 0.426933,
        "pstdev": 0.39075,
        "percentiles": {
          "p50": 0.399887,
          "p75": 0.423567,
          "p80": 0.43081499999999995,
          "p90": 0.453743,
          "p95": 0.482719,
          "p99": 0.888159,
          "p999": 4.870911
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 137.75,
            "min": 119.390625,
            "mean": 135.84114583333334
          },
          "cpu": {
            "max": 1.2000000000000004,
            "min": 0.26666666666666694,
            "mean": 0.523809523809524
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 24.57421875,
            "min": 6.71484375,
            "mean": 22.530598958333332
          },
          "cpu": {
            "max": 9.426398721169218,
            "min": 5.449132527399091,
            "mean": 8.087672764585205
          }
        }
      },
      "poolOverflow": 1,
      "upstreamConnections": 9,
      "counters": {
        "benchmark.http_2xx": {
          "value": 35999,
          "perSecond": 399.9888888888889
        },
        "benchmark.pool_overflow": {
          "value": 1,
          "perSecond": 0.011111111111111112
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
          "value": 5651843,
          "perSecond": 62798.25555555556
        },
        "upstream_cx_total": {
          "value": 9,
          "perSecond": 0.1
        },
        "upstream_cx_tx_bytes_total": {
          "value": 1619955,
          "perSecond": 17999.5
        },
        "upstream_rq_pending_overflow": {
          "value": 1,
          "perSecond": 0.011111111111111112
        },
        "upstream_rq_pending_total": {
          "value": 9,
          "perSecond": 0.1
        },
        "upstream_rq_total": {
          "value": 35999,
          "perSecond": 399.9888888888889
        }
      }
    },
    {
      "testName": "scaling up httproutes to 50 with 10 routes per hostname at 300 rps",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-up",
      "throughput": 1213,
      "totalRequests": 107957,
      "latency": {
        "max": 83.54201499999999,
        "min": 0.27854399999999996,
        "mean": 0.416908,
        "pstdev": 0.76607,
        "percentiles": {
          "p50": 0.386223,
          "p75": 0.40463899999999997,
          "p80": 0.410015,
          "p90": 0.427055,
          "p95": 0.457007,
          "p99": 1.030719,
          "p999": 4.182271
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 143.77734375,
            "min": 138.45703125,
            "mean": 140.41380208333334
          },
          "cpu": {
            "max": 1.866666666666667,
            "min": 0.40000000000000036,
            "mean": 0.7466666666666664
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 30.94140625,
            "min": 28.3046875,
            "mean": 30.134635416666665
          },
          "cpu": {
            "max": 25.687688289072725,
            "min": 15.110782636047029,
            "mean": 23.981684291342074
          }
        }
      },
      "poolOverflow": 43,
      "upstreamConnections": 29,
      "counters": {
        "benchmark.http_2xx": {
          "value": 107957,
          "perSecond": 1213
        },
        "benchmark.pool_overflow": {
          "value": 43,
          "perSecond": 0.48314606741573035
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
          "value": 29,
          "perSecond": 0.3258426966292135
        },
        "upstream_cx_rx_bytes_total": {
          "value": 16949249,
          "perSecond": 190441
        },
        "upstream_cx_total": {
          "value": 29,
          "perSecond": 0.3258426966292135
        },
        "upstream_cx_tx_bytes_total": {
          "value": 4858065,
          "perSecond": 54585
        },
        "upstream_rq_pending_overflow": {
          "value": 43,
          "perSecond": 0.48314606741573035
        },
        "upstream_rq_pending_total": {
          "value": 29,
          "perSecond": 0.3258426966292135
        },
        "upstream_rq_total": {
          "value": 107957,
          "perSecond": 1213
        }
      }
    },
    {
      "testName": "scaling up httproutes to 100 with 20 routes per hostname at 500 rps",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-up",
      "throughput": 2021.9325842696628,
      "totalRequests": 179952,
      "latency": {
        "max": 39.815166999999995,
        "min": 0.25008,
        "mean": 0.393862,
        "pstdev": 0.291014,
        "percentiles": {
          "p50": 0.359743,
          "p75": 0.403711,
          "p80": 0.416543,
          "p90": 0.460687,
          "p95": 0.544991,
          "p99": 0.997439,
          "p999": 3.0465269999999998
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 146.54296875,
            "min": 138.87109375,
            "mean": 144.67838541666666
          },
          "cpu": {
            "max": 2.2666666666666657,
            "min": 0.4666666666666685,
            "mean": 0.8977777777777782
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 36.9375,
            "min": 36.58203125,
            "mean": 36.77981770833333
          },
          "cpu": {
            "max": 36.94134174742694,
            "min": 26.86066527358325,
            "mean": 35.05044941527442
          }
        }
      },
      "poolOverflow": 48,
      "upstreamConnections": 27,
      "counters": {
        "benchmark.http_2xx": {
          "value": 179952,
          "perSecond": 2021.9325842696628
        },
        "benchmark.pool_overflow": {
          "value": 48,
          "perSecond": 0.5393258426966292
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
          "value": 28252464,
          "perSecond": 317443.4157303371
        },
        "upstream_cx_total": {
          "value": 27,
          "perSecond": 0.30337078651685395
        },
        "upstream_cx_tx_bytes_total": {
          "value": 8097840,
          "perSecond": 90986.96629213484
        },
        "upstream_rq_pending_overflow": {
          "value": 48,
          "perSecond": 0.5393258426966292
        },
        "upstream_rq_pending_total": {
          "value": 27,
          "perSecond": 0.30337078651685395
        },
        "upstream_rq_total": {
          "value": 179952,
          "perSecond": 2021.9325842696628
        }
      }
    },
    {
      "testName": "scaling up httproutes to 300 with 60 routes per hostname at 800 rps",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-up",
      "throughput": 3199.588888888889,
      "totalRequests": 287963,
      "latency": {
        "max": 74.407935,
        "min": 0.2362,
        "mean": 0.636105,
        "pstdev": 2.963721,
        "percentiles": {
          "p50": 0.33427100000000004,
          "p75": 0.424143,
          "p80": 0.44676699999999997,
          "p90": 0.550239,
          "p95": 0.731583,
          "p99": 3.988735,
          "p999": 51.693567
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 152.87109375,
            "min": 146.68359375,
            "mean": 150.56302083333333
          },
          "cpu": {
            "max": 4.800000000000004,
            "min": 0.5333333333333339,
            "mean": 1.4555555555555566
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 68.76953125,
            "min": 56.9609375,
            "mean": 63.75859375
          },
          "cpu": {
            "max": 58.60384001418934,
            "min": 34.71300790173815,
            "mean": 53.306176495585646
          }
        }
      },
      "poolOverflow": 37,
      "upstreamConnections": 226,
      "counters": {
        "benchmark.http_2xx": {
          "value": 287963,
          "perSecond": 3199.588888888889
        },
        "benchmark.pool_overflow": {
          "value": 37,
          "perSecond": 0.4111111111111111
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
          "value": 226,
          "perSecond": 2.511111111111111
        },
        "upstream_cx_rx_bytes_total": {
          "value": 45210191,
          "perSecond": 502335.4555555555
        },
        "upstream_cx_total": {
          "value": 226,
          "perSecond": 2.511111111111111
        },
        "upstream_cx_tx_bytes_total": {
          "value": 12958335,
          "perSecond": 143981.5
        },
        "upstream_rq_pending_overflow": {
          "value": 37,
          "perSecond": 0.4111111111111111
        },
        "upstream_rq_pending_total": {
          "value": 226,
          "perSecond": 2.511111111111111
        },
        "upstream_rq_total": {
          "value": 287963,
          "perSecond": 3199.588888888889
        }
      }
    },
    {
      "testName": "scaling up httproutes to 500 with 100 routes per hostname at 1000 rps",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-up",
      "throughput": 3999.266666666667,
      "totalRequests": 359934,
      "latency": {
        "max": 191.356927,
        "min": 0.21468,
        "mean": 1.664008,
        "pstdev": 8.160039,
        "percentiles": {
          "p50": 0.45743900000000004,
          "p75": 0.523199,
          "p80": 0.564031,
          "p90": 0.905023,
          "p95": 1.908159,
          "p99": 47.011839,
          "p999": 96.559103
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 174.10546875,
            "min": 161.8203125,
            "mean": 169.17161458333334
          },
          "cpu": {
            "max": 7.266666666666666,
            "min": 0.6666666666666762,
            "mean": 1.8755555555555554
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 93.60546875,
            "min": 83.7734375,
            "mean": 90.74986979166667
          },
          "cpu": {
            "max": 66.83496625843536,
            "min": 34.83865624987582,
            "mean": 54.83174977413697
          }
        }
      },
      "poolOverflow": 66,
      "upstreamConnections": 334,
      "counters": {
        "benchmark.http_2xx": {
          "value": 359934,
          "perSecond": 3999.266666666667
        },
        "benchmark.pool_overflow": {
          "value": 66,
          "perSecond": 0.7333333333333333
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
          "value": 334,
          "perSecond": 3.7111111111111112
        },
        "upstream_cx_rx_bytes_total": {
          "value": 56509638,
          "perSecond": 627884.8666666667
        },
        "upstream_cx_total": {
          "value": 334,
          "perSecond": 3.7111111111111112
        },
        "upstream_cx_tx_bytes_total": {
          "value": 16197030,
          "perSecond": 179967
        },
        "upstream_rq_pending_overflow": {
          "value": 66,
          "perSecond": 0.7333333333333333
        },
        "upstream_rq_pending_total": {
          "value": 334,
          "perSecond": 3.7111111111111112
        },
        "upstream_rq_total": {
          "value": 359934,
          "perSecond": 3999.266666666667
        }
      }
    },
    {
      "testName": "scaling up httproutes to 1000 with 200 routes per hostname at 2000 rps",
      "routes": 1000,
      "routesPerHostname": 200,
      "phase": "scaling-up",
      "throughput": 7976.5,
      "totalRequests": 717885,
      "latency": {
        "max": 233.98809500000002,
        "min": 0.229776,
        "mean": 14.102452999999999,
        "pstdev": 15.460493000000001,
        "percentiles": {
          "p50": 3.779839,
          "p75": 25.521151,
          "p80": 27.223039,
          "p90": 31.238143,
          "p95": 35.014655,
          "p99": 77.422591,
          "p999": 90.357759
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 192.08984375,
            "min": 174.54296875,
            "mean": 186.37473958333334
          },
          "cpu": {
            "max": 13.133333333333324,
            "min": 0.8000000000000304,
            "mean": 2.880000000000004
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 143.33984375,
            "min": 138.078125,
            "mean": 141.55286458333333
          },
          "cpu": {
            "max": 98.90207227264949,
            "min": 62.16400644335651,
            "mean": 90.35971139102087
          }
        }
      },
      "poolOverflow": 150,
      "upstreamConnections": 250,
      "counters": {
        "benchmark.http_2xx": {
          "value": 717885,
          "perSecond": 7976.5
        },
        "benchmark.pool_overflow": {
          "value": 150,
          "perSecond": 1.6666666666666667
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
          "value": 250,
          "perSecond": 2.7777777777777777
        },
        "upstream_cx_rx_bytes_total": {
          "value": 112707945,
          "perSecond": 1252310.5
        },
        "upstream_cx_total": {
          "value": 250,
          "perSecond": 2.7777777777777777
        },
        "upstream_cx_tx_bytes_total": {
          "value": 32315175,
          "perSecond": 359057.5
        },
        "upstream_rq_pending_overflow": {
          "value": 150,
          "perSecond": 1.6666666666666667
        },
        "upstream_rq_pending_total": {
          "value": 250,
          "perSecond": 2.7777777777777777
        },
        "upstream_rq_total": {
          "value": 718115,
          "perSecond": 7979.055555555556
        }
      }
    },
    {
      "testName": "scaling down httproutes to 500 with 100 routes per hostname at 1000 rps",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-down",
      "throughput": 3999.488888888889,
      "totalRequests": 359954,
      "latency": {
        "max": 233.693183,
        "min": 0.231872,
        "mean": 1.694018,
        "pstdev": 8.011149,
        "percentiles": {
          "p50": 0.461503,
          "p75": 0.522703,
          "p80": 0.558111,
          "p90": 0.840639,
          "p95": 1.9069429999999998,
          "p99": 48.564223,
          "p999": 89.624575
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 173.07421875,
            "min": 167.5703125,
            "mean": 171.127734375
          },
          "cpu": {
            "max": 1.1333333333333446,
            "min": 0.9999999999999432,
            "mean": 1.0515151515151429
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 143.82421875,
            "min": 142.14453125,
            "mean": 143.23541666666668
          },
          "cpu": {
            "max": 67.09508363517963,
            "min": 67.09508363517963,
            "mean": 67.09508363517963
          }
        }
      },
      "poolOverflow": 46,
      "upstreamConnections": 354,
      "counters": {
        "benchmark.http_2xx": {
          "value": 359954,
          "perSecond": 3999.488888888889
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
          "value": 354,
          "perSecond": 3.933333333333333
        },
        "upstream_cx_rx_bytes_total": {
          "value": 56512778,
          "perSecond": 627919.7555555556
        },
        "upstream_cx_total": {
          "value": 354,
          "perSecond": 3.933333333333333
        },
        "upstream_cx_tx_bytes_total": {
          "value": 16197930,
          "perSecond": 179977
        },
        "upstream_rq_pending_overflow": {
          "value": 46,
          "perSecond": 0.5111111111111111
        },
        "upstream_rq_pending_total": {
          "value": 354,
          "perSecond": 3.933333333333333
        },
        "upstream_rq_total": {
          "value": 359954,
          "perSecond": 3999.488888888889
        }
      }
    },
    {
      "testName": "scaling down httproutes to 300 with 60 routes per hostname at 800 rps",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-down",
      "throughput": 3199.5,
      "totalRequests": 287955,
      "latency": {
        "max": 122.720255,
        "min": 0.235768,
        "mean": 0.640906,
        "pstdev": 3.1200669999999997,
        "percentiles": {
          "p50": 0.32247899999999996,
          "p75": 0.366303,
          "p80": 0.417903,
          "p90": 0.529279,
          "p95": 0.753823,
          "p99": 5.074687,
          "p999": 50.528255
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 178.47265625,
            "min": 161.37890625,
            "mean": 164.63216145833334
          },
          "cpu": {
            "max": 2.6000000000000036,
            "min": 0.9999999999999432,
            "mean": 1.333333333333333
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 143.71484375,
            "min": 133.97265625,
            "mean": 136.52604166666666
          },
          "cpu": {
            "max": 56.556327110947876,
            "min": 34.234786377497414,
            "mean": 51.36810909110853
          }
        }
      },
      "poolOverflow": 41,
      "upstreamConnections": 197,
      "counters": {
        "benchmark.http_2xx": {
          "value": 287955,
          "perSecond": 3199.5
        },
        "benchmark.pool_overflow": {
          "value": 41,
          "perSecond": 0.45555555555555555
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
          "value": 45208935,
          "perSecond": 502321.5
        },
        "upstream_cx_total": {
          "value": 197,
          "perSecond": 2.188888888888889
        },
        "upstream_cx_tx_bytes_total": {
          "value": 12958155,
          "perSecond": 143979.5
        },
        "upstream_rq_pending_overflow": {
          "value": 41,
          "perSecond": 0.45555555555555555
        },
        "upstream_rq_pending_total": {
          "value": 197,
          "perSecond": 2.188888888888889
        },
        "upstream_rq_total": {
          "value": 287959,
          "perSecond": 3199.5444444444443
        }
      }
    },
    {
      "testName": "scaling down httproutes to 100 with 20 routes per hostname at 500 rps",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-down",
      "throughput": 1999.4,
      "totalRequests": 179946,
      "latency": {
        "max": 40.019966999999994,
        "min": 0.255168,
        "mean": 0.386061,
        "pstdev": 0.370692,
        "percentiles": {
          "p50": 0.361807,
          "p75": 0.379759,
          "p80": 0.384175,
          "p90": 0.398287,
          "p95": 0.428847,
          "p99": 1.0680310000000002,
          "p999": 4.079359
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 158.55078125,
            "min": 151.36328125,
            "mean": 153.33072916666666
          },
          "cpu": {
            "max": 1.1333333333333446,
            "min": 0.933333333333337,
            "mean": 0.9848484848484845
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 134.265625,
            "min": 133.7421875,
            "mean": 133.99661458333333
          },
          "cpu": {
            "max": 39.644203192640646,
            "min": 22.70835085046973,
            "mean": 36.96250951395476
          }
        }
      },
      "poolOverflow": 54,
      "upstreamConnections": 36,
      "counters": {
        "benchmark.http_2xx": {
          "value": 179946,
          "perSecond": 1999.4
        },
        "benchmark.pool_overflow": {
          "value": 54,
          "perSecond": 0.6
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
          "value": 36,
          "perSecond": 0.4
        },
        "upstream_cx_rx_bytes_total": {
          "value": 28251522,
          "perSecond": 313905.8
        },
        "upstream_cx_total": {
          "value": 36,
          "perSecond": 0.4
        },
        "upstream_cx_tx_bytes_total": {
          "value": 8097570,
          "perSecond": 89973
        },
        "upstream_rq_pending_overflow": {
          "value": 54,
          "perSecond": 0.6
        },
        "upstream_rq_pending_total": {
          "value": 36,
          "perSecond": 0.4
        },
        "upstream_rq_total": {
          "value": 179946,
          "perSecond": 1999.4
        }
      }
    },
    {
      "testName": "scaling down httproutes to 50 with 10 routes per hostname at 300 rps",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-down",
      "throughput": 1212.7640449438202,
      "totalRequests": 107936,
      "latency": {
        "max": 56.717311,
        "min": 0.26424000000000003,
        "mean": 0.40393999999999997,
        "pstdev": 0.624905,
        "percentiles": {
          "p50": 0.375199,
          "p75": 0.391215,
          "p80": 0.395695,
          "p90": 0.411343,
          "p95": 0.434719,
          "p99": 1.033279,
          "p999": 4.856318999999999
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 153.9609375,
            "min": 148.01171875,
            "mean": 149.52682291666667
          },
          "cpu": {
            "max": 1.1999999999999509,
            "min": 0.8666666666666363,
            "mean": 0.9888888888888814
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 134.21875,
            "min": 133.6484375,
            "mean": 133.89127604166666
          },
          "cpu": {
            "max": 0,
            "min": 0,
            "mean": 0
          }
        }
      },
      "poolOverflow": 64,
      "upstreamConnections": 29,
      "counters": {
        "benchmark.http_2xx": {
          "value": 107936,
          "perSecond": 1212.7640449438202
        },
        "benchmark.pool_overflow": {
          "value": 64,
          "perSecond": 0.7191011235955056
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
          "value": 29,
          "perSecond": 0.3258426966292135
        },
        "upstream_cx_rx_bytes_total": {
          "value": 16945952,
          "perSecond": 190403.95505617978
        },
        "upstream_cx_total": {
          "value": 29,
          "perSecond": 0.3258426966292135
        },
        "upstream_cx_tx_bytes_total": {
          "value": 4857120,
          "perSecond": 54574.38202247191
        },
        "upstream_rq_pending_overflow": {
          "value": 64,
          "perSecond": 0.7191011235955056
        },
        "upstream_rq_pending_total": {
          "value": 29,
          "perSecond": 0.3258426966292135
        },
        "upstream_rq_total": {
          "value": 107936,
          "perSecond": 1212.7640449438202
        }
      }
    },
    {
      "testName": "scaling down httproutes to 10 with 2 routes per hostname at 100 rps",
      "routes": 10,
      "routesPerHostname": 2,
      "phase": "scaling-down",
      "throughput": 404.438202247191,
      "totalRequests": 35995,
      "latency": {
        "max": 48.982015,
        "min": 0.316272,
        "mean": 0.47786,
        "pstdev": 0.880794,
        "percentiles": {
          "p50": 0.428367,
          "p75": 0.45654300000000003,
          "p80": 0.47054300000000004,
          "p90": 0.496687,
          "p95": 0.526687,
          "p99": 1.1299190000000001,
          "p999": 8.804863
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 148.640625,
            "min": 143.78125,
            "mean": 145.92122395833334
          },
          "cpu": {
            "max": 1.1333333333333446,
            "min": 0.933333333333337,
            "mean": 1.0222222222222215
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 134.15234375,
            "min": 133.6875,
            "mean": 133.94466145833334
          },
          "cpu": {
            "max": 9.708975160130983,
            "min": 6.287532313700823,
            "mean": 8.870891522527844
          }
        }
      },
      "poolOverflow": 5,
      "upstreamConnections": 16,
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
          "value": 16,
          "perSecond": 0.1797752808988764
        },
        "upstream_cx_rx_bytes_total": {
          "value": 5651215,
          "perSecond": 63496.79775280899
        },
        "upstream_cx_total": {
          "value": 16,
          "perSecond": 0.1797752808988764
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
          "value": 16,
          "perSecond": 0.1797752808988764
        },
        "upstream_rq_total": {
          "value": 35995,
          "perSecond": 404.438202247191
        }
      }
    }
  ]
};
