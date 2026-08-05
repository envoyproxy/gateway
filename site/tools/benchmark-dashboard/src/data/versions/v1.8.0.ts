import { TestSuite } from '../types';

// Benchmark data extracted from release artifact for version 1.8.0
// Generated from benchmark_result.json

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.8.0",
    "runId": "1.8.0-release-2026-05-13",
    "date": "2026-05-13T15:45:46Z",
    "environment": "GitHub Release",
    "description": "Benchmark results for version 1.8.0 from release artifacts",
    "downloadUrl": "https://github.com/envoyproxy/gateway/releases/download/v1.8.0/benchmark_report.zip",
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
      "throughput": 404.39325842696627,
      "totalRequests": 35991,
      "latency": {
        "max": 60.844031,
        "min": 0.37550399999999995,
        "mean": 0.48515800000000003,
        "pstdev": 0.5689510000000001,
        "percentiles": {
          "p50": 0.442959,
          "p75": 0.460783,
          "p80": 0.464623,
          "p90": 0.481535,
          "p95": 0.519087,
          "p99": 1.1913589999999998,
          "p999": 6.606847
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 129.33984375,
            "min": 101.58984375,
            "mean": 122.695703125
          },
          "cpu": {
            "max": 0.46666666666666673,
            "min": 0.33333333333333287,
            "mean": 0.4111111111111112
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 20.86328125,
            "min": 7.546875,
            "mean": 18.339574353448278
          },
          "cpu": {
            "max": 10.287521430320846,
            "min": 5.4345973605157,
            "mean": 8.106397471790913
          }
        }
      },
      "poolOverflow": 9,
      "upstreamConnections": 12,
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
          "value": 12,
          "perSecond": 0.1348314606741573
        },
        "upstream_cx_rx_bytes_total": {
          "value": 5650587,
          "perSecond": 63489.74157303371
        },
        "upstream_cx_total": {
          "value": 12,
          "perSecond": 0.1348314606741573
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
          "value": 12,
          "perSecond": 0.1348314606741573
        },
        "upstream_rq_total": {
          "value": 35991,
          "perSecond": 404.39325842696627
        }
      }
    },
    {
      "testName": "scaling up httproutes to 50 with 10 routes per hostname at 300 rps",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-up",
      "throughput": 1212.4943820224719,
      "totalRequests": 107912,
      "latency": {
        "max": 92.65151900000001,
        "min": 0.339872,
        "mean": 0.498101,
        "pstdev": 0.9952330000000001,
        "percentiles": {
          "p50": 0.445919,
          "p75": 0.461375,
          "p80": 0.466127,
          "p90": 0.49267099999999997,
          "p95": 0.547999,
          "p99": 1.456383,
          "p999": 5.356287
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 134.60546875,
            "min": 127.90625,
            "mean": 132.28541666666666
          },
          "cpu": {
            "max": 0.6000000000000005,
            "min": 0.39999999999999886,
            "mean": 0.4869565217391301
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 24.96875,
            "min": 20.359375,
            "mean": 23.803125
          },
          "cpu": {
            "max": 29.920842185542302,
            "min": 21.45938352063402,
            "mean": 28.947064689830615
          }
        }
      },
      "poolOverflow": 88,
      "upstreamConnections": 32,
      "counters": {
        "benchmark.http_2xx": {
          "value": 107912,
          "perSecond": 1212.4943820224719
        },
        "benchmark.pool_overflow": {
          "value": 88,
          "perSecond": 0.9887640449438202
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
          "value": 16942184,
          "perSecond": 190361.61797752808
        },
        "upstream_cx_total": {
          "value": 32,
          "perSecond": 0.3595505617977528
        },
        "upstream_cx_tx_bytes_total": {
          "value": 4856040,
          "perSecond": 54562.247191011236
        },
        "upstream_rq_pending_overflow": {
          "value": 88,
          "perSecond": 0.9887640449438202
        },
        "upstream_rq_pending_total": {
          "value": 32,
          "perSecond": 0.3595505617977528
        },
        "upstream_rq_total": {
          "value": 107912,
          "perSecond": 1212.4943820224719
        }
      }
    },
    {
      "testName": "scaling up httproutes to 100 with 20 routes per hostname at 500 rps",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-up",
      "throughput": 1999.0222222222221,
      "totalRequests": 179912,
      "latency": {
        "max": 63.74195100000001,
        "min": 0.348576,
        "mean": 0.518629,
        "pstdev": 0.836542,
        "percentiles": {
          "p50": 0.44159899999999996,
          "p75": 0.46043100000000003,
          "p80": 0.46492700000000003,
          "p90": 0.497487,
          "p95": 0.683295,
          "p99": 2.061119,
          "p999": 11.345407
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 140.12890625,
            "min": 131.4453125,
            "mean": 137.73307291666666
          },
          "cpu": {
            "max": 0.6666666666666672,
            "min": 0.5333333333333339,
            "mean": 0.5909090909090912
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 29.26953125,
            "min": 24.78125,
            "mean": 28.626041666666666
          },
          "cpu": {
            "max": 49.50165070307723,
            "min": 35.43637304454524,
            "mean": 48.11188505844771
          }
        }
      },
      "poolOverflow": 88,
      "upstreamConnections": 61,
      "counters": {
        "benchmark.http_2xx": {
          "value": 179912,
          "perSecond": 1999.0222222222221
        },
        "benchmark.pool_overflow": {
          "value": 88,
          "perSecond": 0.9777777777777777
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
          "value": 28246184,
          "perSecond": 313846.4888888889
        },
        "upstream_cx_total": {
          "value": 61,
          "perSecond": 0.6777777777777778
        },
        "upstream_cx_tx_bytes_total": {
          "value": 8096040,
          "perSecond": 89956
        },
        "upstream_rq_pending_overflow": {
          "value": 88,
          "perSecond": 0.9777777777777777
        },
        "upstream_rq_pending_total": {
          "value": 61,
          "perSecond": 0.6777777777777778
        },
        "upstream_rq_total": {
          "value": 179912,
          "perSecond": 1999.0222222222221
        }
      }
    },
    {
      "testName": "scaling up httproutes to 300 with 60 routes per hostname at 800 rps",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-up",
      "throughput": 3199.5777777777776,
      "totalRequests": 287962,
      "latency": {
        "max": 72.74905500000001,
        "min": 0.317536,
        "mean": 0.86458,
        "pstdev": 2.317196,
        "percentiles": {
          "p50": 0.594783,
          "p75": 0.6669430000000001,
          "p80": 0.711903,
          "p90": 0.9254709999999999,
          "p95": 1.382655,
          "p99": 4.547071,
          "p999": 41.443327000000004
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 148.8984375,
            "min": 142.625,
            "mean": 146.22877604166666
          },
          "cpu": {
            "max": 28.4,
            "min": 0.6666666666666643,
            "mean": 6.008873485589029
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 48.43359375,
            "min": 32.796875,
            "mean": 43.454166666666666
          },
          "cpu": {
            "max": 68.6065377149196,
            "min": 68.6065377149196,
            "mean": 68.6065377149196
          }
        }
      },
      "poolOverflow": 35,
      "upstreamConnections": 211,
      "counters": {
        "benchmark.http_2xx": {
          "value": 287962,
          "perSecond": 3199.5777777777776
        },
        "benchmark.pool_overflow": {
          "value": 35,
          "perSecond": 0.3888888888888889
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
          "value": 211,
          "perSecond": 2.3444444444444446
        },
        "upstream_cx_rx_bytes_total": {
          "value": 45210034,
          "perSecond": 502333.7111111111
        },
        "upstream_cx_total": {
          "value": 211,
          "perSecond": 2.3444444444444446
        },
        "upstream_cx_tx_bytes_total": {
          "value": 12958425,
          "perSecond": 143982.5
        },
        "upstream_rq_pending_overflow": {
          "value": 35,
          "perSecond": 0.3888888888888889
        },
        "upstream_rq_pending_total": {
          "value": 211,
          "perSecond": 2.3444444444444446
        },
        "upstream_rq_total": {
          "value": 287965,
          "perSecond": 3199.6111111111113
        }
      }
    },
    {
      "testName": "scaling up httproutes to 500 with 100 routes per hostname at 1000 rps",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-up",
      "throughput": 3999.133333333333,
      "totalRequests": 359922,
      "latency": {
        "max": 130.183167,
        "min": 0.326256,
        "mean": 1.359798,
        "pstdev": 5.120169,
        "percentiles": {
          "p50": 0.644479,
          "p75": 0.814079,
          "p80": 0.8828790000000001,
          "p90": 1.238335,
          "p95": 2.279423,
          "p99": 21.554174999999997,
          "p999": 76.931071
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 167.62109375,
            "min": 153.66015625,
            "mean": 164.03059895833334
          },
          "cpu": {
            "max": 49.60000000000001,
            "min": 0.7333333333333295,
            "mean": 9.057777777777776
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 69.98828125,
            "min": 49.10546875,
            "mean": 63.48033854166667
          },
          "cpu": {
            "max": 82.9537167555781,
            "min": 58.941972332206745,
            "mean": 78.03539083542282
          }
        }
      },
      "poolOverflow": 75,
      "upstreamConnections": 325,
      "counters": {
        "benchmark.http_2xx": {
          "value": 359922,
          "perSecond": 3999.133333333333
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
          "value": 56507754,
          "perSecond": 627863.9333333333
        },
        "upstream_cx_total": {
          "value": 325,
          "perSecond": 3.611111111111111
        },
        "upstream_cx_tx_bytes_total": {
          "value": 16196580,
          "perSecond": 179962
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
          "value": 359924,
          "perSecond": 3999.1555555555556
        }
      }
    },
    {
      "testName": "scaling up httproutes to 1000 with 200 routes per hostname at 2000 rps",
      "routes": 1000,
      "routesPerHostname": 200,
      "phase": "scaling-up",
      "throughput": 6043.5,
      "totalRequests": 543915,
      "latency": {
        "max": 1135.345663,
        "min": 2.0096,
        "mean": 51.45764500000001,
        "pstdev": 15.012057,
        "percentiles": {
          "p50": 51.800063,
          "p75": 57.352191,
          "p80": 58.351615,
          "p90": 60.934143000000006,
          "p95": 63.979518999999996,
          "p99": 90.02188699999999,
          "p999": 135.438335
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 194.875,
            "min": 165.51953125,
            "mean": 192.38671875
          },
          "cpu": {
            "max": 94.19999999999997,
            "min": 0.8000000000000304,
            "mean": 16.59333333333334
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 105.08984375,
            "min": 80.03125,
            "mean": 100.88229166666666
          },
          "cpu": {
            "max": 99.56727514596123,
            "min": 68.13554583696704,
            "mean": 93.12984261229636
          }
        }
      },
      "poolOverflow": 87,
      "upstreamConnections": 313,
      "counters": {
        "benchmark.http_2xx": {
          "value": 543915,
          "perSecond": 6043.5
        },
        "benchmark.pool_overflow": {
          "value": 87,
          "perSecond": 0.9666666666666667
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
          "value": 313,
          "perSecond": 3.477777777777778
        },
        "upstream_cx_rx_bytes_total": {
          "value": 85394655,
          "perSecond": 948829.5
        },
        "upstream_cx_total": {
          "value": 313,
          "perSecond": 3.477777777777778
        },
        "upstream_cx_tx_bytes_total": {
          "value": 24490260,
          "perSecond": 272114
        },
        "upstream_rq_pending_overflow": {
          "value": 87,
          "perSecond": 0.9666666666666667
        },
        "upstream_rq_pending_total": {
          "value": 313,
          "perSecond": 3.477777777777778
        },
        "upstream_rq_total": {
          "value": 544228,
          "perSecond": 6046.977777777778
        }
      }
    },
    {
      "testName": "scaling down httproutes to 500 with 100 routes per hostname at 1000 rps",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-down",
      "throughput": 3999.1222222222223,
      "totalRequests": 359921,
      "latency": {
        "max": 124.64947099999999,
        "min": 0.308336,
        "mean": 1.323936,
        "pstdev": 5.229215,
        "percentiles": {
          "p50": 0.634143,
          "p75": 0.8154549999999999,
          "p80": 0.889439,
          "p90": 1.2160630000000001,
          "p95": 2.128127,
          "p99": 18.499583,
          "p999": 82.882559
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 173.19140625,
            "min": 163.875,
            "mean": 167.57057291666666
          },
          "cpu": {
            "max": 1.1999999999999507,
            "min": 0.933333333333337,
            "mean": 1.0363636363636282
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 104.51171875,
            "min": 102.859375,
            "mean": 103.88216145833333
          },
          "cpu": {
            "max": 82.94537447151188,
            "min": 59.30601867489608,
            "mean": 79.5770549883814
          }
        }
      },
      "poolOverflow": 75,
      "upstreamConnections": 322,
      "counters": {
        "benchmark.http_2xx": {
          "value": 359921,
          "perSecond": 3999.1222222222223
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
          "value": 322,
          "perSecond": 3.577777777777778
        },
        "upstream_cx_rx_bytes_total": {
          "value": 56507597,
          "perSecond": 627862.1888888889
        },
        "upstream_cx_total": {
          "value": 322,
          "perSecond": 3.577777777777778
        },
        "upstream_cx_tx_bytes_total": {
          "value": 16196625,
          "perSecond": 179962.5
        },
        "upstream_rq_pending_overflow": {
          "value": 75,
          "perSecond": 0.8333333333333334
        },
        "upstream_rq_pending_total": {
          "value": 322,
          "perSecond": 3.577777777777778
        },
        "upstream_rq_total": {
          "value": 359925,
          "perSecond": 3999.1666666666665
        }
      }
    },
    {
      "testName": "scaling down httproutes to 300 with 60 routes per hostname at 800 rps",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-down",
      "throughput": 3199.411111111111,
      "totalRequests": 287947,
      "latency": {
        "max": 69.144575,
        "min": 0.31392,
        "mean": 0.8586320000000001,
        "pstdev": 2.1950279999999998,
        "percentiles": {
          "p50": 0.5969909999999999,
          "p75": 0.683039,
          "p80": 0.7361909999999999,
          "p90": 0.993631,
          "p95": 1.4826869999999999,
          "p99": 4.3184629999999995,
          "p999": 40.722431
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 172.1328125,
            "min": 153.3828125,
            "mean": 159.8359375
          },
          "cpu": {
            "max": 1.0666666666667386,
            "min": 0.933333333333337,
            "mean": 1.0166666666666662
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 103.9296875,
            "min": 102.8984375,
            "mean": 103.58268229166667
          },
          "cpu": {
            "max": 68.8044860276349,
            "min": 44.415199491315,
            "mean": 62.765503557186285
          }
        }
      },
      "poolOverflow": 52,
      "upstreamConnections": 189,
      "counters": {
        "benchmark.http_2xx": {
          "value": 287947,
          "perSecond": 3199.411111111111
        },
        "benchmark.pool_overflow": {
          "value": 52,
          "perSecond": 0.5777777777777777
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
          "value": 189,
          "perSecond": 2.1
        },
        "upstream_cx_rx_bytes_total": {
          "value": 45207679,
          "perSecond": 502307.5444444445
        },
        "upstream_cx_total": {
          "value": 189,
          "perSecond": 2.1
        },
        "upstream_cx_tx_bytes_total": {
          "value": 12957660,
          "perSecond": 143974
        },
        "upstream_rq_pending_overflow": {
          "value": 52,
          "perSecond": 0.5777777777777777
        },
        "upstream_rq_pending_total": {
          "value": 189,
          "perSecond": 2.1
        },
        "upstream_rq_total": {
          "value": 287948,
          "perSecond": 3199.4222222222224
        }
      }
    },
    {
      "testName": "scaling down httproutes to 100 with 20 routes per hostname at 500 rps",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-down",
      "throughput": 2020.887640449438,
      "totalRequests": 179859,
      "latency": {
        "max": 66.33676700000001,
        "min": 0.32340800000000003,
        "mean": 0.550293,
        "pstdev": 0.861711,
        "percentiles": {
          "p50": 0.44910300000000003,
          "p75": 0.540383,
          "p80": 0.549759,
          "p90": 0.6192949999999999,
          "p95": 0.944383,
          "p99": 1.894911,
          "p999": 8.984063
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 159.85546875,
            "min": 141.671875,
            "mean": 146.059765625
          },
          "cpu": {
            "max": 1.1333333333333449,
            "min": 0.933333333333337,
            "mean": 1.0363636363636408
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 103.4765625,
            "min": 102.9609375,
            "mean": 103.1703125
          },
          "cpu": {
            "max": 45.96425033467191,
            "min": 45.96425033467191,
            "mean": 45.96425033467191
          }
        }
      },
      "poolOverflow": 140,
      "upstreamConnections": 53,
      "counters": {
        "benchmark.http_2xx": {
          "value": 179859,
          "perSecond": 2020.887640449438
        },
        "benchmark.pool_overflow": {
          "value": 140,
          "perSecond": 1.5730337078651686
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
          "value": 53,
          "perSecond": 0.5955056179775281
        },
        "upstream_cx_rx_bytes_total": {
          "value": 28237863,
          "perSecond": 317279.3595505618
        },
        "upstream_cx_total": {
          "value": 53,
          "perSecond": 0.5955056179775281
        },
        "upstream_cx_tx_bytes_total": {
          "value": 8093655,
          "perSecond": 90939.94382022473
        },
        "upstream_rq_pending_overflow": {
          "value": 140,
          "perSecond": 1.5730337078651686
        },
        "upstream_rq_pending_total": {
          "value": 53,
          "perSecond": 0.5955056179775281
        },
        "upstream_rq_total": {
          "value": 179859,
          "perSecond": 2020.887640449438
        }
      }
    },
    {
      "testName": "scaling down httproutes to 50 with 10 routes per hostname at 300 rps",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-down",
      "throughput": 1199.6555555555556,
      "totalRequests": 107969,
      "latency": {
        "max": 29.548543,
        "min": 0.342528,
        "mean": 0.49396,
        "pstdev": 0.46276,
        "percentiles": {
          "p50": 0.444223,
          "p75": 0.463871,
          "p80": 0.46807899999999997,
          "p90": 0.496095,
          "p95": 0.548479,
          "p99": 1.618303,
          "p999": 7.135999
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 148.67578125,
            "min": 145.35546875,
            "mean": 147.05924479166666
          },
          "cpu": {
            "max": 1.066666666666644,
            "min": 0.8666666666666363,
            "mean": 0.9666666666666638
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 103.4765625,
            "min": 102.55859375,
            "mean": 102.98919270833333
          },
          "cpu": {
            "max": 20.25151466666635,
            "min": 20.25151466666635,
            "mean": 20.25151466666635
          }
        }
      },
      "poolOverflow": 31,
      "upstreamConnections": 30,
      "counters": {
        "benchmark.http_2xx": {
          "value": 107969,
          "perSecond": 1199.6555555555556
        },
        "benchmark.pool_overflow": {
          "value": 31,
          "perSecond": 0.34444444444444444
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
          "value": 16951133,
          "perSecond": 188345.92222222223
        },
        "upstream_cx_total": {
          "value": 30,
          "perSecond": 0.3333333333333333
        },
        "upstream_cx_tx_bytes_total": {
          "value": 4858605,
          "perSecond": 53984.5
        },
        "upstream_rq_pending_overflow": {
          "value": 31,
          "perSecond": 0.34444444444444444
        },
        "upstream_rq_pending_total": {
          "value": 30,
          "perSecond": 0.3333333333333333
        },
        "upstream_rq_total": {
          "value": 107969,
          "perSecond": 1199.6555555555556
        }
      }
    },
    {
      "testName": "scaling down httproutes to 10 with 2 routes per hostname at 100 rps",
      "routes": 10,
      "routesPerHostname": 2,
      "phase": "scaling-down",
      "throughput": 399.9888888888889,
      "totalRequests": 35999,
      "latency": {
        "max": 30.369791,
        "min": 0.365792,
        "mean": 0.49008500000000005,
        "pstdev": 0.506989,
        "percentiles": {
          "p50": 0.44926299999999997,
          "p75": 0.467951,
          "p80": 0.47259100000000004,
          "p90": 0.493983,
          "p95": 0.5300469999999999,
          "p99": 1.070783,
          "p999": 7.098879
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 147.73046875,
            "min": 138.51171875,
            "mean": 141.698046875
          },
          "cpu": {
            "max": 1.1333333333333446,
            "min": 0.933333333333337,
            "mean": 1.0272727272727271
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 103.16015625,
            "min": 102.6015625,
            "mean": 102.71276041666667
          },
          "cpu": {
            "max": 10.248671815326777,
            "min": 6.000252904630225,
            "mean": 9.154973976397287
          }
        }
      },
      "poolOverflow": 1,
      "upstreamConnections": 10,
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
          "value": 10,
          "perSecond": 0.1111111111111111
        },
        "upstream_cx_rx_bytes_total": {
          "value": 5651843,
          "perSecond": 62798.25555555556
        },
        "upstream_cx_total": {
          "value": 10,
          "perSecond": 0.1111111111111111
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
          "value": 10,
          "perSecond": 0.1111111111111111
        },
        "upstream_rq_total": {
          "value": 35999,
          "perSecond": 399.9888888888889
        }
      }
    }
  ]
};
