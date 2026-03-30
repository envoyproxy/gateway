import { TestSuite } from '../types';

// Benchmark data extracted from release artifact for version 1.6.5
// Generated from benchmark_result.json

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.6.5",
    "runId": "1.6.5-release-2026-03-13",
    "date": "2026-03-13T02:12:48Z",
    "environment": "GitHub Release",
    "description": "Benchmark results for version 1.6.5 from release artifacts",
    "downloadUrl": "https://github.com/envoyproxy/gateway/releases/download/v1.6.5/benchmark_report.zip",
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
      "throughput": 404.4831460674157,
      "totalRequests": 35999,
      "latency": {
        "max": 28.723199,
        "min": 0.39584,
        "mean": 0.53622,
        "pstdev": 0.476311,
        "percentiles": {
          "p50": 0.494831,
          "p75": 0.512095,
          "p80": 0.517439,
          "p90": 0.536639,
          "p95": 0.580063,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 138.1640625,
            "min": 122.62890625,
            "mean": 134.53190104166666
          },
          "cpu": {
            "max": 1.0000000000000002,
            "min": 0.3999999999999996,
            "mean": 0.6000000000000002
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 26.79296875,
            "min": 22.66796875,
            "mean": 25.896614583333335
          },
          "cpu": {
            "max": 10.442485090963993,
            "min": 10.442485090963993,
            "mean": 10.442485090963993
          }
        }
      },
      "poolOverflow": 1,
      "upstreamConnections": 11,
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
          "value": 11,
          "perSecond": 0.12359550561797752
        },
        "upstream_cx_rx_bytes_total": {
          "value": 5651843,
          "perSecond": 63503.85393258427
        },
        "upstream_cx_total": {
          "value": 11,
          "perSecond": 0.12359550561797752
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
          "value": 11,
          "perSecond": 0.12359550561797752
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
      "throughput": 1212.9325842696628,
      "totalRequests": 107951,
      "latency": {
        "max": 40.710142999999995,
        "min": 0.34687999999999997,
        "mean": 0.521477,
        "pstdev": 0.607942,
        "percentiles": {
          "p50": 0.47779099999999997,
          "p75": 0.49417500000000003,
          "p80": 0.49868700000000005,
          "p90": 0.520239,
          "p95": 0.570399,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 148.19921875,
            "min": 138.01171875,
            "mean": 144.55247395833334
          },
          "cpu": {
            "max": 2.066666666666667,
            "min": 0.40000000000000036,
            "mean": 0.8044444444444443
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 33.01953125,
            "min": 26.8125,
            "mean": 31.680078125
          },
          "cpu": {
            "max": 31.36975133214921,
            "min": 18.01566929840301,
            "mean": 26.704873468780946
          }
        }
      },
      "poolOverflow": 49,
      "upstreamConnections": 36,
      "counters": {
        "benchmark.http_2xx": {
          "value": 107951,
          "perSecond": 1212.9325842696628
        },
        "benchmark.pool_overflow": {
          "value": 49,
          "perSecond": 0.550561797752809
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
          "value": 16948307,
          "perSecond": 190430.4157303371
        },
        "upstream_cx_total": {
          "value": 36,
          "perSecond": 0.4044943820224719
        },
        "upstream_cx_tx_bytes_total": {
          "value": 4857795,
          "perSecond": 54581.96629213483
        },
        "upstream_rq_pending_overflow": {
          "value": 49,
          "perSecond": 0.550561797752809
        },
        "upstream_rq_pending_total": {
          "value": 36,
          "perSecond": 0.4044943820224719
        },
        "upstream_rq_total": {
          "value": 107951,
          "perSecond": 1212.9325842696628
        }
      }
    },
    {
      "testName": "scaling up httproutes to 100 with 20 routes per hostname at 500 rps",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-up",
      "throughput": 2020.6629213483145,
      "totalRequests": 179839,
      "latency": {
        "max": 78.09843099999999,
        "min": 0.34444800000000003,
        "mean": 0.5169130000000001,
        "pstdev": 0.690655,
        "percentiles": {
          "p50": 0.470895,
          "p75": 0.48995099999999997,
          "p80": 0.494351,
          "p90": 0.518127,
          "p95": 0.5935670000000001,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 150.58984375,
            "min": 142.328125,
            "mean": 147.31705729166666
          },
          "cpu": {
            "max": 2.8000000000000034,
            "min": 0.5333333333333339,
            "mean": 0.9666666666666673
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 39.15625,
            "min": 36.890625,
            "mean": 37.129557291666664
          },
          "cpu": {
            "max": 50.127887801494964,
            "min": 31.940523292577545,
            "mean": 47.198493725135904
          }
        }
      },
      "poolOverflow": 161,
      "upstreamConnections": 44,
      "counters": {
        "benchmark.http_2xx": {
          "value": 179839,
          "perSecond": 2020.6629213483145
        },
        "benchmark.pool_overflow": {
          "value": 161,
          "perSecond": 1.8089887640449438
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
          "value": 44,
          "perSecond": 0.4943820224719101
        },
        "upstream_cx_rx_bytes_total": {
          "value": 28234723,
          "perSecond": 317244.0786516854
        },
        "upstream_cx_total": {
          "value": 44,
          "perSecond": 0.4943820224719101
        },
        "upstream_cx_tx_bytes_total": {
          "value": 8092755,
          "perSecond": 90929.83146067416
        },
        "upstream_rq_pending_overflow": {
          "value": 161,
          "perSecond": 1.8089887640449438
        },
        "upstream_rq_pending_total": {
          "value": 44,
          "perSecond": 0.4943820224719101
        },
        "upstream_rq_total": {
          "value": 179839,
          "perSecond": 2020.6629213483145
        }
      }
    },
    {
      "testName": "scaling up httproutes to 300 with 60 routes per hostname at 800 rps",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-up",
      "throughput": 3199.011111111111,
      "totalRequests": 287911,
      "latency": {
        "max": 92.168191,
        "min": 0.30192,
        "mean": 0.9269270000000001,
        "pstdev": 3.663421,
        "percentiles": {
          "p50": 0.514383,
          "p75": 0.613663,
          "p80": 0.653407,
          "p90": 0.843391,
          "p95": 1.150847,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 155.2109375,
            "min": 150.078125,
            "mean": 153.849609375
          },
          "cpu": {
            "max": 4.6000000000000085,
            "min": 0.7999999999999948,
            "mean": 1.466666666666667
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 68.71484375,
            "min": 59.22265625,
            "mean": 65.490234375
          },
          "cpu": {
            "max": 77.80208344578179,
            "min": 77.80208344578179,
            "mean": 77.80208344578179
          }
        }
      },
      "poolOverflow": 89,
      "upstreamConnections": 203,
      "counters": {
        "benchmark.http_2xx": {
          "value": 287911,
          "perSecond": 3199.011111111111
        },
        "benchmark.pool_overflow": {
          "value": 89,
          "perSecond": 0.9888888888888889
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
          "value": 203,
          "perSecond": 2.2555555555555555
        },
        "upstream_cx_rx_bytes_total": {
          "value": 45202027,
          "perSecond": 502244.7444444444
        },
        "upstream_cx_total": {
          "value": 203,
          "perSecond": 2.2555555555555555
        },
        "upstream_cx_tx_bytes_total": {
          "value": 12955995,
          "perSecond": 143955.5
        },
        "upstream_rq_pending_overflow": {
          "value": 89,
          "perSecond": 0.9888888888888889
        },
        "upstream_rq_pending_total": {
          "value": 203,
          "perSecond": 2.2555555555555555
        },
        "upstream_rq_total": {
          "value": 287911,
          "perSecond": 3199.011111111111
        }
      }
    },
    {
      "testName": "scaling up httproutes to 500 with 100 routes per hostname at 1000 rps",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-up",
      "throughput": 3999.211111111111,
      "totalRequests": 359929,
      "latency": {
        "max": 184.549375,
        "min": 0.27736,
        "mean": 2.461132,
        "pstdev": 10.065368999999999,
        "percentiles": {
          "p50": 0.570431,
          "p75": 0.785471,
          "p80": 0.858463,
          "p90": 1.5511030000000001,
          "p95": 4.246783,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 172.86328125,
            "min": 162.59375,
            "mean": 168.85026041666666
          },
          "cpu": {
            "max": 6.866666666666673,
            "min": 0.7333333333333295,
            "mean": 1.9111111111111094
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 97.9296875,
            "min": 88.1484375,
            "mean": 94.99869791666667
          },
          "cpu": {
            "max": 92.92058106130418,
            "min": 48.84836610607844,
            "mean": 73.26369989227736
          }
        }
      },
      "poolOverflow": 70,
      "upstreamConnections": 330,
      "counters": {
        "benchmark.http_2xx": {
          "value": 359929,
          "perSecond": 3999.211111111111
        },
        "benchmark.pool_overflow": {
          "value": 70,
          "perSecond": 0.7777777777777778
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
          "value": 330,
          "perSecond": 3.6666666666666665
        },
        "upstream_cx_rx_bytes_total": {
          "value": 56508853,
          "perSecond": 627876.1444444444
        },
        "upstream_cx_total": {
          "value": 330,
          "perSecond": 3.6666666666666665
        },
        "upstream_cx_tx_bytes_total": {
          "value": 16196850,
          "perSecond": 179965
        },
        "upstream_rq_pending_overflow": {
          "value": 70,
          "perSecond": 0.7777777777777778
        },
        "upstream_rq_pending_total": {
          "value": 330,
          "perSecond": 3.6666666666666665
        },
        "upstream_rq_total": {
          "value": 359930,
          "perSecond": 3999.222222222222
        }
      }
    },
    {
      "testName": "scaling up httproutes to 1000 with 200 routes per hostname at 2000 rps",
      "routes": 1000,
      "routesPerHostname": 200,
      "phase": "scaling-up",
      "throughput": 5479.355555555556,
      "totalRequests": 493142,
      "latency": {
        "max": 225.689599,
        "min": 0.33609599999999995,
        "mean": 37.651956,
        "pstdev": 34.037777,
        "percentiles": {
          "p50": 21.173247,
          "p75": 67.024895,
          "p80": 72.24524699999999,
          "p90": 86.859775,
          "p95": 98.26303899999999,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 198.2734375,
            "min": 188.02734375,
            "mean": 194.09921875
          },
          "cpu": {
            "max": 15.53333333333332,
            "min": 0.933333333333337,
            "mean": 3.202222222222225
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 164.03125,
            "min": 152.83984375,
            "mean": 155.52122395833334
          },
          "cpu": {
            "max": 99.97510441629572,
            "min": 99.86972027972055,
            "mean": 99.93038849522263
          }
        }
      },
      "poolOverflow": 184,
      "upstreamConnections": 216,
      "counters": {
        "benchmark.http_2xx": {
          "value": 493142,
          "perSecond": 5479.355555555556
        },
        "benchmark.pool_overflow": {
          "value": 184,
          "perSecond": 2.0444444444444443
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
          "value": 216,
          "perSecond": 2.4
        },
        "upstream_cx_rx_bytes_total": {
          "value": 77423294,
          "perSecond": 860258.8222222222
        },
        "upstream_cx_total": {
          "value": 216,
          "perSecond": 2.4
        },
        "upstream_cx_tx_bytes_total": {
          "value": 22201110,
          "perSecond": 246679
        },
        "upstream_rq_pending_overflow": {
          "value": 184,
          "perSecond": 2.0444444444444443
        },
        "upstream_rq_pending_total": {
          "value": 216,
          "perSecond": 2.4
        },
        "upstream_rq_total": {
          "value": 493358,
          "perSecond": 5481.7555555555555
        }
      }
    },
    {
      "testName": "scaling down httproutes to 500 with 100 routes per hostname at 1000 rps",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-down",
      "throughput": 4043.629213483146,
      "totalRequests": 359883,
      "latency": {
        "max": 155.81183900000002,
        "min": 0.292848,
        "mean": 2.017162,
        "pstdev": 8.171909999999999,
        "percentiles": {
          "p50": 0.5915509999999999,
          "p75": 0.820703,
          "p80": 0.9196789999999999,
          "p90": 1.597695,
          "p95": 3.339007,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 212.12890625,
            "min": 168.421875,
            "mean": 183.72526041666666
          },
          "cpu": {
            "max": 3.533333333333341,
            "min": 1.066666666666644,
            "mean": 1.680000000000007
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 154.5078125,
            "min": 152.875,
            "mean": 153.92044270833333
          },
          "cpu": {
            "max": 94.21263415152491,
            "min": 53.16742164449063,
            "mean": 87.0267242259528
          }
        }
      },
      "poolOverflow": 114,
      "upstreamConnections": 286,
      "counters": {
        "benchmark.http_2xx": {
          "value": 359883,
          "perSecond": 4043.629213483146
        },
        "benchmark.pool_overflow": {
          "value": 114,
          "perSecond": 1.2808988764044944
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
          "value": 286,
          "perSecond": 3.2134831460674156
        },
        "upstream_cx_rx_bytes_total": {
          "value": 56501631,
          "perSecond": 634849.786516854
        },
        "upstream_cx_total": {
          "value": 286,
          "perSecond": 3.2134831460674156
        },
        "upstream_cx_tx_bytes_total": {
          "value": 16194870,
          "perSecond": 181964.83146067415
        },
        "upstream_rq_pending_overflow": {
          "value": 114,
          "perSecond": 1.2808988764044944
        },
        "upstream_rq_pending_total": {
          "value": 286,
          "perSecond": 3.2134831460674156
        },
        "upstream_rq_total": {
          "value": 359886,
          "perSecond": 4043.6629213483147
        }
      }
    },
    {
      "testName": "scaling down httproutes to 300 with 60 routes per hostname at 800 rps",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-down",
      "throughput": 3198.0666666666666,
      "totalRequests": 287826,
      "latency": {
        "max": 132.198399,
        "min": 0.303344,
        "mean": 0.91598,
        "pstdev": 3.57819,
        "percentiles": {
          "p50": 0.508191,
          "p75": 0.608255,
          "p80": 0.647199,
          "p90": 0.8413109999999999,
          "p95": 1.133375,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 171.65234375,
            "min": 160.578125,
            "mean": 163.03046875
          },
          "cpu": {
            "max": 1.1333333333333446,
            "min": 1.066666666666644,
            "mean": 1.0811594202898598
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 153.41796875,
            "min": 148.40625,
            "mean": 149.32838541666666
          },
          "cpu": {
            "max": 77.81052772856602,
            "min": 46.823040725676336,
            "mean": 73.12420330057152
          }
        }
      },
      "poolOverflow": 174,
      "upstreamConnections": 184,
      "counters": {
        "benchmark.http_2xx": {
          "value": 287826,
          "perSecond": 3198.0666666666666
        },
        "benchmark.pool_overflow": {
          "value": 174,
          "perSecond": 1.9333333333333333
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
          "value": 184,
          "perSecond": 2.0444444444444443
        },
        "upstream_cx_rx_bytes_total": {
          "value": 45188682,
          "perSecond": 502096.4666666667
        },
        "upstream_cx_total": {
          "value": 184,
          "perSecond": 2.0444444444444443
        },
        "upstream_cx_tx_bytes_total": {
          "value": 12952170,
          "perSecond": 143913
        },
        "upstream_rq_pending_overflow": {
          "value": 174,
          "perSecond": 1.9333333333333333
        },
        "upstream_rq_pending_total": {
          "value": 184,
          "perSecond": 2.0444444444444443
        },
        "upstream_rq_total": {
          "value": 287826,
          "perSecond": 3198.0666666666666
        }
      }
    },
    {
      "testName": "scaling down httproutes to 100 with 20 routes per hostname at 500 rps",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-down",
      "throughput": 1997.4888888888888,
      "totalRequests": 179774,
      "latency": {
        "max": 99.303423,
        "min": 0.35732800000000003,
        "mean": 0.536752,
        "pstdev": 1.146692,
        "percentiles": {
          "p50": 0.473007,
          "p75": 0.490287,
          "p80": 0.493727,
          "p90": 0.5159509999999999,
          "p95": 0.605375,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 164.15625,
            "min": 152.47265625,
            "mean": 155.97005208333334
          },
          "cpu": {
            "max": 1.2000000000000455,
            "min": 1.0000000000000377,
            "mean": 1.0833333333333426
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 149.203125,
            "min": 148.59765625,
            "mean": 148.73841145833333
          },
          "cpu": {
            "max": 50.68842311583141,
            "min": 50.65819479910333,
            "mean": 50.68168741577659
          }
        }
      },
      "poolOverflow": 226,
      "upstreamConnections": 61,
      "counters": {
        "benchmark.http_2xx": {
          "value": 179774,
          "perSecond": 1997.4888888888888
        },
        "benchmark.pool_overflow": {
          "value": 226,
          "perSecond": 2.511111111111111
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
          "value": 61,
          "perSecond": 0.6777777777777778
        },
        "upstream_cx_rx_bytes_total": {
          "value": 28224518,
          "perSecond": 313605.7555555556
        },
        "upstream_cx_total": {
          "value": 61,
          "perSecond": 0.6777777777777778
        },
        "upstream_cx_tx_bytes_total": {
          "value": 8089830,
          "perSecond": 89887
        },
        "upstream_rq_pending_overflow": {
          "value": 226,
          "perSecond": 2.511111111111111
        },
        "upstream_rq_pending_total": {
          "value": 61,
          "perSecond": 0.6777777777777778
        },
        "upstream_rq_total": {
          "value": 179774,
          "perSecond": 1997.4888888888888
        }
      }
    },
    {
      "testName": "scaling down httproutes to 50 with 10 routes per hostname at 300 rps",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-down",
      "throughput": 1199.4,
      "totalRequests": 107946,
      "latency": {
        "max": 72.323071,
        "min": 0.365152,
        "mean": 0.52208,
        "pstdev": 0.6794760000000001,
        "percentiles": {
          "p50": 0.47329499999999997,
          "p75": 0.490239,
          "p80": 0.494767,
          "p90": 0.5165270000000001,
          "p95": 0.563295,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 156.05859375,
            "min": 144.6796875,
            "mean": 149.99635416666666
          },
          "cpu": {
            "max": 1.1999999999999507,
            "min": 0.933333333333337,
            "mean": 1.033333333333328
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 148.77734375,
            "min": 148.56640625,
            "mean": 148.661328125
          },
          "cpu": {
            "max": 30.93122336812199,
            "min": 17.72173064097516,
            "mean": 24.95206369864552
          }
        }
      },
      "poolOverflow": 54,
      "upstreamConnections": 37,
      "counters": {
        "benchmark.http_2xx": {
          "value": 107946,
          "perSecond": 1199.4
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
          "value": 37,
          "perSecond": 0.4111111111111111
        },
        "upstream_cx_rx_bytes_total": {
          "value": 16947522,
          "perSecond": 188305.8
        },
        "upstream_cx_total": {
          "value": 37,
          "perSecond": 0.4111111111111111
        },
        "upstream_cx_tx_bytes_total": {
          "value": 4857570,
          "perSecond": 53973
        },
        "upstream_rq_pending_overflow": {
          "value": 54,
          "perSecond": 0.6
        },
        "upstream_rq_pending_total": {
          "value": 37,
          "perSecond": 0.4111111111111111
        },
        "upstream_rq_total": {
          "value": 107946,
          "perSecond": 1199.4
        }
      }
    },
    {
      "testName": "scaling down httproutes to 10 with 2 routes per hostname at 100 rps",
      "routes": 10,
      "routesPerHostname": 2,
      "phase": "scaling-down",
      "throughput": 399.96666666666664,
      "totalRequests": 35997,
      "latency": {
        "max": 35.893247,
        "min": 0.395216,
        "mean": 0.526502,
        "pstdev": 0.442801,
        "percentiles": {
          "p50": 0.493535,
          "p75": 0.510991,
          "p80": 0.5156310000000001,
          "p90": 0.531391,
          "p95": 0.560223,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 148.875,
            "min": 146.53125,
            "mean": 147.14166666666668
          },
          "cpu": {
            "max": 1.1333333333333446,
            "min": 0.933333333333337,
            "mean": 1.0305555555555694
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 148.70703125,
            "min": 148.58203125,
            "mean": 148.61848958333334
          },
          "cpu": {
            "max": 11.14762859633839,
            "min": 6.712358765475222,
            "mean": 9.116260686856515
          }
        }
      },
      "poolOverflow": 3,
      "upstreamConnections": 12,
      "counters": {
        "benchmark.http_2xx": {
          "value": 35997,
          "perSecond": 399.96666666666664
        },
        "benchmark.pool_overflow": {
          "value": 3,
          "perSecond": 0.03333333333333333
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
          "value": 5651529,
          "perSecond": 62794.76666666667
        },
        "upstream_cx_total": {
          "value": 12,
          "perSecond": 0.13333333333333333
        },
        "upstream_cx_tx_bytes_total": {
          "value": 1619865,
          "perSecond": 17998.5
        },
        "upstream_rq_pending_overflow": {
          "value": 3,
          "perSecond": 0.03333333333333333
        },
        "upstream_rq_pending_total": {
          "value": 12,
          "perSecond": 0.13333333333333333
        },
        "upstream_rq_total": {
          "value": 35997,
          "perSecond": 399.96666666666664
        }
      }
    }
  ]
};

export default benchmarkData;
