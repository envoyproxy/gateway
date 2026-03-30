import { TestSuite } from '../types';

// Benchmark data extracted from release artifact for version 1.7.1
// Generated from benchmark_result.json

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.7.1",
    "runId": "1.7.1-release-2026-03-12",
    "date": "2026-03-12T17:15:44Z",
    "environment": "GitHub Release",
    "description": "Benchmark results for version 1.7.1 from release artifacts",
    "downloadUrl": "https://github.com/envoyproxy/gateway/releases/download/v1.7.1/benchmark_report.zip",
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
      "throughput": 404.46067415730334,
      "totalRequests": 35997,
      "latency": {
        "max": 21.831679,
        "min": 0.36027200000000004,
        "mean": 0.50792,
        "pstdev": 0.327603,
        "percentiles": {
          "p50": 0.47825500000000004,
          "p75": 0.494671,
          "p80": 0.500831,
          "p90": 0.524479,
          "p95": 0.565567,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 141.2109375,
            "min": 126.21484375,
            "mean": 134.95377604166666
          },
          "cpu": {
            "max": 1.2666666666666664,
            "min": 0.3999999999999996,
            "mean": 0.6333333333333335
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 23.01953125,
            "min": 6.2890625,
            "mean": 21.337369791666667
          },
          "cpu": {
            "max": 10.880065789473683,
            "min": 6.7053471008693455,
            "mean": 9.803489410005396
          }
        }
      },
      "poolOverflow": 3,
      "upstreamConnections": 9,
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
          "value": 9,
          "perSecond": 0.10112359550561797
        },
        "upstream_cx_rx_bytes_total": {
          "value": 5651529,
          "perSecond": 63500.32584269663
        },
        "upstream_cx_total": {
          "value": 9,
          "perSecond": 0.10112359550561797
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
          "value": 9,
          "perSecond": 0.10112359550561797
        },
        "upstream_rq_total": {
          "value": 35997,
          "perSecond": 404.46067415730334
        }
      }
    },
    {
      "testName": "scaling up httproutes to 50 with 10 routes per hostname at 300 rps",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-up",
      "throughput": 1199.7777777777778,
      "totalRequests": 107980,
      "latency": {
        "max": 33.160191,
        "min": 0.350496,
        "mean": 0.502856,
        "pstdev": 0.381932,
        "percentiles": {
          "p50": 0.46670300000000003,
          "p75": 0.48307100000000003,
          "p80": 0.488031,
          "p90": 0.509791,
          "p95": 0.563007,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 147.07421875,
            "min": 133.85546875,
            "mean": 142.88971354166668
          },
          "cpu": {
            "max": 2.066666666666667,
            "min": 0.46666666666666556,
            "mean": 0.8288888888888886
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 29.1484375,
            "min": 26.9921875,
            "mean": 28.961067708333335
          },
          "cpu": {
            "max": 31.337084666860616,
            "min": 16.902799943566244,
            "mean": 27.581361126197127
          }
        }
      },
      "poolOverflow": 20,
      "upstreamConnections": 28,
      "counters": {
        "benchmark.http_2xx": {
          "value": 107980,
          "perSecond": 1199.7777777777778
        },
        "benchmark.pool_overflow": {
          "value": 20,
          "perSecond": 0.2222222222222222
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
          "value": 28,
          "perSecond": 0.3111111111111111
        },
        "upstream_cx_rx_bytes_total": {
          "value": 16952860,
          "perSecond": 188365.11111111112
        },
        "upstream_cx_total": {
          "value": 28,
          "perSecond": 0.3111111111111111
        },
        "upstream_cx_tx_bytes_total": {
          "value": 4859100,
          "perSecond": 53990
        },
        "upstream_rq_pending_overflow": {
          "value": 20,
          "perSecond": 0.2222222222222222
        },
        "upstream_rq_pending_total": {
          "value": 28,
          "perSecond": 0.3111111111111111
        },
        "upstream_rq_total": {
          "value": 107980,
          "perSecond": 1199.7777777777778
        }
      }
    },
    {
      "testName": "scaling up httproutes to 100 with 20 routes per hostname at 500 rps",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-up",
      "throughput": 2020.8988764044943,
      "totalRequests": 179860,
      "latency": {
        "max": 159.686655,
        "min": 0.34865599999999997,
        "mean": 0.555528,
        "pstdev": 0.899074,
        "percentiles": {
          "p50": 0.465487,
          "p75": 0.49182299999999995,
          "p80": 0.499215,
          "p90": 0.559263,
          "p95": 0.823231,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 148.53125,
            "min": 141.1015625,
            "mean": 146.11067708333334
          },
          "cpu": {
            "max": 2.666666666666666,
            "min": 0.5333333333333339,
            "mean": 1.0111111111111104
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 35.53515625,
            "min": 33.25390625,
            "mean": 35.204166666666666
          },
          "cpu": {
            "max": 50.98887138280554,
            "min": 28.235218195210045,
            "mean": 46.107534023712816
          }
        }
      },
      "poolOverflow": 140,
      "upstreamConnections": 53,
      "counters": {
        "benchmark.http_2xx": {
          "value": 179860,
          "perSecond": 2020.8988764044943
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
          "value": 28238020,
          "perSecond": 317281.1235955056
        },
        "upstream_cx_total": {
          "value": 53,
          "perSecond": 0.5955056179775281
        },
        "upstream_cx_tx_bytes_total": {
          "value": 8093700,
          "perSecond": 90940.44943820225
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
          "value": 179860,
          "perSecond": 2020.8988764044943
        }
      }
    },
    {
      "testName": "scaling up httproutes to 300 with 60 routes per hostname at 800 rps",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-up",
      "throughput": 3199.5333333333333,
      "totalRequests": 287958,
      "latency": {
        "max": 195.354623,
        "min": 0.340448,
        "mean": 2.3120990000000003,
        "pstdev": 9.991765999999998,
        "percentiles": {
          "p50": 0.643839,
          "p75": 0.961279,
          "p80": 1.1813749999999998,
          "p90": 2.493439,
          "p95": 4.838911,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 155.48828125,
            "min": 145.59765625,
            "mean": 151.63020833333334
          },
          "cpu": {
            "max": 5.266666666666662,
            "min": 0.86666666666666,
            "mean": 1.6333333333333317
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 70.109375,
            "min": 57.375,
            "mean": 62.9921875
          },
          "cpu": {
            "max": 73.13433104940499,
            "min": 72.0589681943401,
            "mean": 72.51389466052387
          }
        }
      },
      "poolOverflow": 40,
      "upstreamConnections": 360,
      "counters": {
        "benchmark.http_2xx": {
          "value": 287958,
          "perSecond": 3199.5333333333333
        },
        "benchmark.pool_overflow": {
          "value": 40,
          "perSecond": 0.4444444444444444
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
          "value": 360,
          "perSecond": 4
        },
        "upstream_cx_rx_bytes_total": {
          "value": 45209406,
          "perSecond": 502326.73333333334
        },
        "upstream_cx_total": {
          "value": 360,
          "perSecond": 4
        },
        "upstream_cx_tx_bytes_total": {
          "value": 12958200,
          "perSecond": 143980
        },
        "upstream_rq_pending_overflow": {
          "value": 40,
          "perSecond": 0.4444444444444444
        },
        "upstream_rq_pending_total": {
          "value": 360,
          "perSecond": 4
        },
        "upstream_rq_total": {
          "value": 287960,
          "perSecond": 3199.5555555555557
        }
      }
    },
    {
      "testName": "scaling up httproutes to 500 with 100 routes per hostname at 1000 rps",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-up",
      "throughput": 3999.788888888889,
      "totalRequests": 359981,
      "latency": {
        "max": 1183.7767669999998,
        "min": 0.348768,
        "mean": 8.674752000000002,
        "pstdev": 30.238035,
        "percentiles": {
          "p50": 1.351807,
          "p75": 3.699583,
          "p80": 5.259263,
          "p90": 16.358911,
          "p95": 57.106431,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 167.328125,
            "min": 162.40625,
            "mean": 165.52278645833334
          },
          "cpu": {
            "max": 7.1999999999999895,
            "min": 0.9999999999999905,
            "mean": 2.0555318550119233
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 94.5546875,
            "min": 84.5390625,
            "mean": 90.65416666666667
          },
          "cpu": {
            "max": 85.74224964628205,
            "min": 85.12550198150596,
            "mean": 85.30171560001341
          }
        }
      },
      "poolOverflow": 15,
      "upstreamConnections": 385,
      "counters": {
        "benchmark.http_2xx": {
          "value": 359981,
          "perSecond": 3999.788888888889
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
          "value": 385,
          "perSecond": 4.277777777777778
        },
        "upstream_cx_rx_bytes_total": {
          "value": 56517017,
          "perSecond": 627966.8555555556
        },
        "upstream_cx_total": {
          "value": 385,
          "perSecond": 4.277777777777778
        },
        "upstream_cx_tx_bytes_total": {
          "value": 16199325,
          "perSecond": 179992.5
        },
        "upstream_rq_pending_overflow": {
          "value": 15,
          "perSecond": 0.16666666666666666
        },
        "upstream_rq_pending_total": {
          "value": 385,
          "perSecond": 4.277777777777778
        },
        "upstream_rq_total": {
          "value": 359985,
          "perSecond": 3999.8333333333335
        }
      }
    },
    {
      "testName": "scaling up httproutes to 1000 with 200 routes per hostname at 2000 rps",
      "routes": 1000,
      "routesPerHostname": 200,
      "phase": "scaling-up",
      "throughput": 4313.4157303370785,
      "totalRequests": 383894,
      "latency": {
        "max": 2317.615103,
        "min": 2.001856,
        "mean": 80.580335,
        "pstdev": 34.233291,
        "percentiles": {
          "p50": 78.331903,
          "p75": 91.250687,
          "p80": 93.556735,
          "p90": 99.053567,
          "p95": 106.225663,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 189.859375,
            "min": 177.27734375,
            "mean": 184.05078125
          },
          "cpu": {
            "max": 14.133333333333365,
            "min": 0.8666666666667312,
            "mean": 3.2955555555555662
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 146.84375,
            "min": 137.078125,
            "mean": 144.76744791666667
          },
          "cpu": {
            "max": 99.10046568276267,
            "min": 56.52672818023091,
            "mean": 88.91415783040362
          }
        }
      },
      "poolOverflow": 54,
      "upstreamConnections": 346,
      "counters": {
        "benchmark.http_2xx": {
          "value": 383894,
          "perSecond": 4313.4157303370785
        },
        "benchmark.pool_overflow": {
          "value": 54,
          "perSecond": 0.6067415730337079
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
          "value": 346,
          "perSecond": 3.8876404494382024
        },
        "upstream_cx_rx_bytes_total": {
          "value": 60271358,
          "perSecond": 677206.2696629213
        },
        "upstream_cx_total": {
          "value": 346,
          "perSecond": 3.8876404494382024
        },
        "upstream_cx_tx_bytes_total": {
          "value": 17290800,
          "perSecond": 194278.65168539327
        },
        "upstream_rq_pending_overflow": {
          "value": 54,
          "perSecond": 0.6067415730337079
        },
        "upstream_rq_pending_total": {
          "value": 346,
          "perSecond": 3.8876404494382024
        },
        "upstream_rq_total": {
          "value": 384240,
          "perSecond": 4317.303370786517
        }
      }
    },
    {
      "testName": "scaling down httproutes to 500 with 100 routes per hostname at 1000 rps",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-down",
      "throughput": 3998.7,
      "totalRequests": 359883,
      "latency": {
        "max": 576.323583,
        "min": 0.337072,
        "mean": 6.515112,
        "pstdev": 17.650902,
        "percentiles": {
          "p50": 1.259839,
          "p75": 3.451263,
          "p80": 4.744959,
          "p90": 13.520383,
          "p95": 35.483647,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 182.9140625,
            "min": 171.34375,
            "mean": 175.60807291666666
          },
          "cpu": {
            "max": 1.466666666666659,
            "min": 0.933333333333337,
            "mean": 1.1466666666666656
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 146.94140625,
            "min": 144.984375,
            "mean": 145.9390625
          },
          "cpu": {
            "max": 86.54882356499097,
            "min": 60.09245426968198,
            "mean": 80.62355319860941
          }
        }
      },
      "poolOverflow": 95,
      "upstreamConnections": 305,
      "counters": {
        "benchmark.http_2xx": {
          "value": 359883,
          "perSecond": 3998.7
        },
        "benchmark.pool_overflow": {
          "value": 95,
          "perSecond": 1.0555555555555556
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
          "value": 305,
          "perSecond": 3.388888888888889
        },
        "upstream_cx_rx_bytes_total": {
          "value": 56501631,
          "perSecond": 627795.9
        },
        "upstream_cx_total": {
          "value": 305,
          "perSecond": 3.388888888888889
        },
        "upstream_cx_tx_bytes_total": {
          "value": 16195725,
          "perSecond": 179952.5
        },
        "upstream_rq_pending_overflow": {
          "value": 95,
          "perSecond": 1.0555555555555556
        },
        "upstream_rq_pending_total": {
          "value": 305,
          "perSecond": 3.388888888888889
        },
        "upstream_rq_total": {
          "value": 359905,
          "perSecond": 3998.9444444444443
        }
      }
    },
    {
      "testName": "scaling down httproutes to 300 with 60 routes per hostname at 800 rps",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-down",
      "throughput": 3199.266666666667,
      "totalRequests": 287934,
      "latency": {
        "max": 292.716543,
        "min": 0.32511999999999996,
        "mean": 2.174077,
        "pstdev": 8.738519,
        "percentiles": {
          "p50": 0.648095,
          "p75": 1.035167,
          "p80": 1.354623,
          "p90": 2.638975,
          "p95": 5.187583,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 172.57421875,
            "min": 157.23046875,
            "mean": 163.800390625
          },
          "cpu": {
            "max": 1.2000000000000455,
            "min": 1.066666666666644,
            "mean": 1.1333340751525947
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 145.29296875,
            "min": 141.73828125,
            "mean": 142.549609375
          },
          "cpu": {
            "max": 73.15572640218886,
            "min": 44.218809476424916,
            "mean": 65.55157206751501
          }
        }
      },
      "poolOverflow": 61,
      "upstreamConnections": 339,
      "counters": {
        "benchmark.http_2xx": {
          "value": 287934,
          "perSecond": 3199.266666666667
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
          "value": 45205638,
          "perSecond": 502284.86666666664
        },
        "upstream_cx_total": {
          "value": 339,
          "perSecond": 3.7666666666666666
        },
        "upstream_cx_tx_bytes_total": {
          "value": 12957255,
          "perSecond": 143969.5
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
          "value": 287939,
          "perSecond": 3199.322222222222
        }
      }
    },
    {
      "testName": "scaling down httproutes to 100 with 20 routes per hostname at 500 rps",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-down",
      "throughput": 1999.0222222222221,
      "totalRequests": 179912,
      "latency": {
        "max": 66.826239,
        "min": 0.358272,
        "mean": 0.551723,
        "pstdev": 0.927127,
        "percentiles": {
          "p50": 0.468031,
          "p75": 0.487647,
          "p80": 0.49355099999999996,
          "p90": 0.531871,
          "p95": 0.759775,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 161.9765625,
            "min": 149.421875,
            "mean": 152.75846354166666
          },
          "cpu": {
            "max": 1.2000000000000455,
            "min": 1.066666666666644,
            "mean": 1.099999999999994
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 141.02734375,
            "min": 138.734375,
            "mean": 139.32994791666667
          },
          "cpu": {
            "max": 51.284291616457324,
            "min": 32.19114984765026,
            "mean": 47.86547970682163
          }
        }
      },
      "poolOverflow": 88,
      "upstreamConnections": 60,
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
          "value": 60,
          "perSecond": 0.6666666666666666
        },
        "upstream_cx_rx_bytes_total": {
          "value": 28246184,
          "perSecond": 313846.4888888889
        },
        "upstream_cx_total": {
          "value": 60,
          "perSecond": 0.6666666666666666
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
          "value": 60,
          "perSecond": 0.6666666666666666
        },
        "upstream_rq_total": {
          "value": 179912,
          "perSecond": 1999.0222222222221
        }
      }
    },
    {
      "testName": "scaling down httproutes to 50 with 10 routes per hostname at 300 rps",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-down",
      "throughput": 1212.6404494382023,
      "totalRequests": 107925,
      "latency": {
        "max": 65.267711,
        "min": 0.353296,
        "mean": 0.527068,
        "pstdev": 0.932132,
        "percentiles": {
          "p50": 0.46964700000000004,
          "p75": 0.486287,
          "p80": 0.490863,
          "p90": 0.5107510000000001,
          "p95": 0.5663349999999999,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 152.046875,
            "min": 148.2265625,
            "mean": 149.540625
          },
          "cpu": {
            "max": 1.1999999999999507,
            "min": 1.0000000000000377,
            "mean": 1.0757575757575657
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 138.80859375,
            "min": 138.703125,
            "mean": 138.75911458333334
          },
          "cpu": {
            "max": 31.886488894168092,
            "min": 19.332278206834392,
            "mean": 27.251537958002597
          }
        }
      },
      "poolOverflow": 75,
      "upstreamConnections": 40,
      "counters": {
        "benchmark.http_2xx": {
          "value": 107925,
          "perSecond": 1212.6404494382023
        },
        "benchmark.pool_overflow": {
          "value": 75,
          "perSecond": 0.8426966292134831
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
          "value": 40,
          "perSecond": 0.449438202247191
        },
        "upstream_cx_rx_bytes_total": {
          "value": 16944225,
          "perSecond": 190384.55056179775
        },
        "upstream_cx_total": {
          "value": 40,
          "perSecond": 0.449438202247191
        },
        "upstream_cx_tx_bytes_total": {
          "value": 4856625,
          "perSecond": 54568.8202247191
        },
        "upstream_rq_pending_overflow": {
          "value": 75,
          "perSecond": 0.8426966292134831
        },
        "upstream_rq_pending_total": {
          "value": 40,
          "perSecond": 0.449438202247191
        },
        "upstream_rq_total": {
          "value": 107925,
          "perSecond": 1212.6404494382023
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
        "max": 52.789246999999996,
        "min": 0.379296,
        "mean": 0.535933,
        "pstdev": 0.861002,
        "percentiles": {
          "p50": 0.480127,
          "p75": 0.499503,
          "p80": 0.506415,
          "p90": 0.534079,
          "p95": 0.5791029999999999,
          "p99": 0,
          "p999": 0
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "max": 151.6875,
            "min": 142.6875,
            "mean": 146.77421875
          },
          "cpu": {
            "max": 1.2000000000000455,
            "min": 1.0000000000000377,
            "mean": 1.0833333333333348
          }
        },
        "envoyProxy": {
          "memory": {
            "max": 138.79296875,
            "min": 138.6875,
            "mean": 138.7109375
          },
          "cpu": {
            "max": 10.990164787510622,
            "min": 6.4548566601374775,
            "mean": 9.83564108772025
          }
        }
      },
      "poolOverflow": 9,
      "upstreamConnections": 15,
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
          "value": 15,
          "perSecond": 0.16853932584269662
        },
        "upstream_cx_rx_bytes_total": {
          "value": 5650587,
          "perSecond": 63489.74157303371
        },
        "upstream_cx_total": {
          "value": 15,
          "perSecond": 0.16853932584269662
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
          "value": 15,
          "perSecond": 0.16853932584269662
        },
        "upstream_rq_total": {
          "value": 35991,
          "perSecond": 404.39325842696627
        }
      }
    }
  ]
};

export default benchmarkData;
