import { TestSuite } from '../types';

// Benchmark data extracted from markdown report for version 1.4.5
// Generated on 2025-11-21T02:41:51.331Z

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.4.5",
    "runId": "1.4.5-1763692911330",
    "date": "2025-11-21T02:41:51.330Z",
    "environment": "production",
    "description": "Benchmark results for version 1.4.5",
    "testConfiguration": {
      "rps": 10000,
      "connections": 100,
      "duration": 30,
      "cpuLimit": "1000m",
      "memoryLimit": "2000Mi"
    }
  },
  "results": [
    {
      "testName": "scaling up httproutes to 10 with 2 routes per hostname",
      "routes": 10,
      "routesPerHostname": 2,
      "phase": "scaling-up",
      "throughput": 5167.23,
      "totalRequests": 155019,
      "latency": {
        "min": 387,
        "mean": 7145,
        "max": 169533,
        "pstdev": 11448,
        "percentiles": {
          "p50": 3498,
          "p75": 6101,
          "p80": 7038,
          "p90": 12180,
          "p95": 43442,
          "p99": 52967,
          "p999": 59131
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 131.36,
            "max": 156.94,
            "mean": 150.91
          },
          "cpu": {
            "min": 0.13,
            "max": 0.93,
            "mean": 0.47
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 0,
            "max": 27.57,
            "mean": 23.8
          },
          "cpu": {
            "min": 0,
            "max": 99.86,
            "mean": 12.25
          }
        }
      },
      "poolOverflow": 360,
      "upstreamConnections": 40,
      "counters": {
        "benchmark.http_2xx": {
          "value": 155019,
          "perSecond": 5167.23
        },
        "benchmark.pool_overflow": {
          "value": 360,
          "perSecond": 12
        },
        "cluster_manager.cluster_added": {
          "value": 4,
          "perSecond": 0.13
        },
        "default.total_match_count": {
          "value": 4,
          "perSecond": 0.13
        },
        "membership_change": {
          "value": 4,
          "perSecond": 0.13
        },
        "runtime.load_success": {
          "value": 1,
          "perSecond": 0.03
        },
        "runtime.override_dir_not_exists": {
          "value": 1,
          "perSecond": 0.03
        },
        "upstream_cx_http1_total": {
          "value": 40,
          "perSecond": 1.33
        },
        "upstream_cx_rx_bytes_total": {
          "value": 24337983,
          "perSecond": 811254.4
        },
        "upstream_cx_total": {
          "value": 40,
          "perSecond": 1.33
        },
        "upstream_cx_tx_bytes_total": {
          "value": 6977115,
          "perSecond": 232567.15
        },
        "upstream_rq_pending_overflow": {
          "value": 360,
          "perSecond": 12
        },
        "upstream_rq_pending_total": {
          "value": 40,
          "perSecond": 1.33
        },
        "upstream_rq_total": {
          "value": 155047,
          "perSecond": 5168.16
        }
      }
    },
    {
      "testName": "scaling up httproutes to 50 with 10 routes per hostname",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-up",
      "throughput": 5091.12,
      "totalRequests": 152734,
      "latency": {
        "min": 325,
        "mean": 7585,
        "max": 117567,
        "pstdev": 12356,
        "percentiles": {
          "p50": 3769,
          "p75": 6053,
          "p80": 6857,
          "p90": 11380,
          "p95": 46624,
          "p99": 54200,
          "p999": 60897
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 155.46,
            "max": 163.91,
            "mean": 163.11
          },
          "cpu": {
            "min": 0.33,
            "max": 4.67,
            "mean": 0.93
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 27.08,
            "max": 33.45,
            "mean": 32.66
          },
          "cpu": {
            "min": 0,
            "max": 99.81,
            "mean": 14.11
          }
        }
      },
      "poolOverflow": 359,
      "upstreamConnections": 41,
      "counters": {
        "benchmark.http_2xx": {
          "value": 152734,
          "perSecond": 5091.12
        },
        "benchmark.pool_overflow": {
          "value": 359,
          "perSecond": 11.97
        },
        "cluster_manager.cluster_added": {
          "value": 4,
          "perSecond": 0.13
        },
        "default.total_match_count": {
          "value": 4,
          "perSecond": 0.13
        },
        "membership_change": {
          "value": 4,
          "perSecond": 0.13
        },
        "runtime.load_success": {
          "value": 1,
          "perSecond": 0.03
        },
        "runtime.override_dir_not_exists": {
          "value": 1,
          "perSecond": 0.03
        },
        "upstream_cx_http1_total": {
          "value": 41,
          "perSecond": 1.37
        },
        "upstream_cx_rx_bytes_total": {
          "value": 23979238,
          "perSecond": 799306.01
        },
        "upstream_cx_total": {
          "value": 41,
          "perSecond": 1.37
        },
        "upstream_cx_tx_bytes_total": {
          "value": 6874875,
          "perSecond": 229161.95
        },
        "upstream_rq_pending_overflow": {
          "value": 359,
          "perSecond": 11.97
        },
        "upstream_rq_pending_total": {
          "value": 41,
          "perSecond": 1.37
        },
        "upstream_rq_total": {
          "value": 152775,
          "perSecond": 5092.49
        }
      }
    },
    {
      "testName": "scaling up httproutes to 100 with 20 routes per hostname",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-up",
      "throughput": 5148.71,
      "totalRequests": 154464,
      "latency": {
        "min": 364,
        "mean": 7180,
        "max": 87302,
        "pstdev": 12250,
        "percentiles": {
          "p50": 3413,
          "p75": 5598,
          "p80": 6370,
          "p90": 10396,
          "p95": 46991,
          "p99": 54482,
          "p999": 63635
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 161.52,
            "max": 166.58,
            "mean": 165.37
          },
          "cpu": {
            "min": 0.4,
            "max": 9,
            "mean": 1.37
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 33.15,
            "max": 37.81,
            "mean": 37.24
          },
          "cpu": {
            "min": 0,
            "max": 94.81,
            "mean": 11.44
          }
        }
      },
      "poolOverflow": 361,
      "upstreamConnections": 39,
      "counters": {
        "benchmark.http_2xx": {
          "value": 154464,
          "perSecond": 5148.71
        },
        "benchmark.pool_overflow": {
          "value": 361,
          "perSecond": 12.03
        },
        "cluster_manager.cluster_added": {
          "value": 4,
          "perSecond": 0.13
        },
        "default.total_match_count": {
          "value": 4,
          "perSecond": 0.13
        },
        "membership_change": {
          "value": 4,
          "perSecond": 0.13
        },
        "runtime.load_success": {
          "value": 1,
          "perSecond": 0.03
        },
        "runtime.override_dir_not_exists": {
          "value": 1,
          "perSecond": 0.03
        },
        "upstream_cx_http1_total": {
          "value": 39,
          "perSecond": 1.3
        },
        "upstream_cx_rx_bytes_total": {
          "value": 24250848,
          "perSecond": 808347.33
        },
        "upstream_cx_total": {
          "value": 39,
          "perSecond": 1.3
        },
        "upstream_cx_tx_bytes_total": {
          "value": 6952410,
          "perSecond": 231742.91
        },
        "upstream_rq_pending_overflow": {
          "value": 361,
          "perSecond": 12.03
        },
        "upstream_rq_pending_total": {
          "value": 39,
          "perSecond": 1.3
        },
        "upstream_rq_total": {
          "value": 154498,
          "perSecond": 5149.84
        }
      }
    },
    {
      "testName": "scaling up httproutes to 300 with 60 routes per hostname",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-up",
      "throughput": 5157.38,
      "totalRequests": 154725,
      "latency": {
        "min": 377,
        "mean": 7029,
        "max": 118747,
        "pstdev": 12310,
        "percentiles": {
          "p50": 3257,
          "p75": 5250,
          "p80": 5989,
          "p90": 10043,
          "p95": 47198,
          "p99": 54951,
          "p999": 65818
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 175.42,
            "max": 187.87,
            "mean": 185.45
          },
          "cpu": {
            "min": 0.53,
            "max": 81.07,
            "mean": 3.8
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 53.46,
            "max": 57.79,
            "mean": 57.44
          },
          "cpu": {
            "min": 0,
            "max": 99.89,
            "mean": 12.24
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38,
      "counters": {
        "benchmark.http_2xx": {
          "value": 154725,
          "perSecond": 5157.38
        },
        "benchmark.pool_overflow": {
          "value": 362,
          "perSecond": 12.07
        },
        "cluster_manager.cluster_added": {
          "value": 4,
          "perSecond": 0.13
        },
        "default.total_match_count": {
          "value": 4,
          "perSecond": 0.13
        },
        "membership_change": {
          "value": 4,
          "perSecond": 0.13
        },
        "runtime.load_success": {
          "value": 1,
          "perSecond": 0.03
        },
        "runtime.override_dir_not_exists": {
          "value": 1,
          "perSecond": 0.03
        },
        "upstream_cx_http1_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_rx_bytes_total": {
          "value": 24291825,
          "perSecond": 809707.91
        },
        "upstream_cx_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_tx_bytes_total": {
          "value": 6963525,
          "perSecond": 232111.88
        },
        "upstream_rq_pending_overflow": {
          "value": 362,
          "perSecond": 12.07
        },
        "upstream_rq_pending_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_rq_total": {
          "value": 154745,
          "perSecond": 5158.04
        }
      }
    },
    {
      "testName": "scaling up httproutes to 500 with 100 routes per hostname",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-up",
      "throughput": 5150.77,
      "totalRequests": 154527,
      "latency": {
        "min": 381,
        "mean": 6971,
        "max": 113311,
        "pstdev": 12316,
        "percentiles": {
          "p50": 3151,
          "p75": 5206,
          "p80": 5983,
          "p90": 10075,
          "p95": 46688,
          "p99": 54849,
          "p999": 68308
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 195.37,
            "max": 203.44,
            "mean": 200.65
          },
          "cpu": {
            "min": 0.13,
            "max": 1,
            "mean": 0.75
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 73.68,
            "max": 79.95,
            "mean": 79.06
          },
          "cpu": {
            "min": 0,
            "max": 99.82,
            "mean": 11.76
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38,
      "counters": {
        "benchmark.http_2xx": {
          "value": 154527,
          "perSecond": 5150.77
        },
        "benchmark.pool_overflow": {
          "value": 362,
          "perSecond": 12.07
        },
        "cluster_manager.cluster_added": {
          "value": 4,
          "perSecond": 0.13
        },
        "default.total_match_count": {
          "value": 4,
          "perSecond": 0.13
        },
        "membership_change": {
          "value": 4,
          "perSecond": 0.13
        },
        "runtime.load_success": {
          "value": 1,
          "perSecond": 0.03
        },
        "runtime.override_dir_not_exists": {
          "value": 1,
          "perSecond": 0.03
        },
        "upstream_cx_http1_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_rx_bytes_total": {
          "value": 24260739,
          "perSecond": 808671.54
        },
        "upstream_cx_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_tx_bytes_total": {
          "value": 6955065,
          "perSecond": 231829.84
        },
        "upstream_rq_pending_overflow": {
          "value": 362,
          "perSecond": 12.07
        },
        "upstream_rq_pending_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_rq_total": {
          "value": 154557,
          "perSecond": 5151.77
        }
      }
    },
    {
      "testName": "scaling up httproutes to 1000 with 200 routes per hostname",
      "routes": 1000,
      "routesPerHostname": 200,
      "phase": "scaling-up",
      "throughput": 4819.38,
      "totalRequests": 144582,
      "latency": {
        "min": 367,
        "mean": 7667,
        "max": 117051,
        "pstdev": 13071,
        "percentiles": {
          "p50": 3465,
          "p75": 5936,
          "p80": 6838,
          "p90": 12270,
          "p95": 47808,
          "p99": 58836,
          "p999": 76193
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 232.1,
            "max": 248.36,
            "mean": 237.82
          },
          "cpu": {
            "min": 0.07,
            "max": 1.47,
            "mean": 0.9
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 129.97,
            "max": 130.7,
            "mean": 130.25
          },
          "cpu": {
            "min": 0,
            "max": 99.96,
            "mean": 10.36
          }
        }
      },
      "poolOverflow": 361,
      "upstreamConnections": 39,
      "counters": {
        "benchmark.http_2xx": {
          "value": 144582,
          "perSecond": 4819.38
        },
        "benchmark.pool_overflow": {
          "value": 361,
          "perSecond": 12.03
        },
        "cluster_manager.cluster_added": {
          "value": 4,
          "perSecond": 0.13
        },
        "default.total_match_count": {
          "value": 4,
          "perSecond": 0.13
        },
        "membership_change": {
          "value": 4,
          "perSecond": 0.13
        },
        "runtime.load_success": {
          "value": 1,
          "perSecond": 0.03
        },
        "runtime.override_dir_not_exists": {
          "value": 1,
          "perSecond": 0.03
        },
        "upstream_cx_http1_total": {
          "value": 39,
          "perSecond": 1.3
        },
        "upstream_cx_rx_bytes_total": {
          "value": 22699374,
          "perSecond": 756642.26
        },
        "upstream_cx_total": {
          "value": 39,
          "perSecond": 1.3
        },
        "upstream_cx_tx_bytes_total": {
          "value": 6507945,
          "perSecond": 216930.48
        },
        "upstream_rq_pending_overflow": {
          "value": 361,
          "perSecond": 12.03
        },
        "upstream_rq_pending_total": {
          "value": 39,
          "perSecond": 1.3
        },
        "upstream_rq_total": {
          "value": 144621,
          "perSecond": 4820.68
        }
      }
    },
    {
      "testName": "scaling down httproutes to 10 with 2 routes per hostname",
      "routes": 10,
      "routesPerHostname": 2,
      "phase": "scaling-down",
      "throughput": 5020.4,
      "totalRequests": 150612,
      "latency": {
        "min": 361,
        "mean": 7165,
        "max": 136364,
        "pstdev": 12568,
        "percentiles": {
          "p50": 3218,
          "p75": 5447,
          "p80": 6252,
          "p90": 10566,
          "p95": 48209,
          "p99": 55255,
          "p999": 64630
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 214.58,
            "max": 217.58,
            "mean": 215.77
          },
          "cpu": {
            "min": 0.87,
            "max": 4,
            "mean": 1.24
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 127.79,
            "max": 128.03,
            "mean": 127.91
          },
          "cpu": {
            "min": 0,
            "max": 100.06,
            "mean": 7.23
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38,
      "counters": {
        "benchmark.http_2xx": {
          "value": 150612,
          "perSecond": 5020.4
        },
        "benchmark.pool_overflow": {
          "value": 362,
          "perSecond": 12.07
        },
        "cluster_manager.cluster_added": {
          "value": 4,
          "perSecond": 0.13
        },
        "default.total_match_count": {
          "value": 4,
          "perSecond": 0.13
        },
        "membership_change": {
          "value": 4,
          "perSecond": 0.13
        },
        "runtime.load_success": {
          "value": 1,
          "perSecond": 0.03
        },
        "runtime.override_dir_not_exists": {
          "value": 1,
          "perSecond": 0.03
        },
        "upstream_cx_http1_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_rx_bytes_total": {
          "value": 23646084,
          "perSecond": 788202.27
        },
        "upstream_cx_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_tx_bytes_total": {
          "value": 6779250,
          "perSecond": 225974.85
        },
        "upstream_rq_pending_overflow": {
          "value": 362,
          "perSecond": 12.07
        },
        "upstream_rq_pending_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_rq_total": {
          "value": 150650,
          "perSecond": 5021.66
        }
      }
    },
    {
      "testName": "scaling down httproutes to 50 with 10 routes per hostname",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-down",
      "throughput": 5156.43,
      "totalRequests": 154695,
      "latency": {
        "min": 398,
        "mean": 7756,
        "max": 94031,
        "pstdev": 13474,
        "percentiles": {
          "p50": 3283,
          "p75": 5911,
          "p80": 6883,
          "p90": 12712,
          "p95": 50251,
          "p99": 57212,
          "p999": 66996
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 209.45,
            "max": 217.88,
            "mean": 215.58
          },
          "cpu": {
            "min": 0.67,
            "max": 7.6,
            "mean": 1.57
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 127.73,
            "max": 128.02,
            "mean": 127.87
          },
          "cpu": {
            "min": 0,
            "max": 99.9,
            "mean": 4.75
          }
        }
      },
      "poolOverflow": 358,
      "upstreamConnections": 42,
      "counters": {
        "benchmark.http_2xx": {
          "value": 154695,
          "perSecond": 5156.43
        },
        "benchmark.pool_overflow": {
          "value": 358,
          "perSecond": 11.93
        },
        "cluster_manager.cluster_added": {
          "value": 4,
          "perSecond": 0.13
        },
        "default.total_match_count": {
          "value": 4,
          "perSecond": 0.13
        },
        "membership_change": {
          "value": 4,
          "perSecond": 0.13
        },
        "runtime.load_success": {
          "value": 1,
          "perSecond": 0.03
        },
        "runtime.override_dir_not_exists": {
          "value": 1,
          "perSecond": 0.03
        },
        "upstream_cx_http1_total": {
          "value": 42,
          "perSecond": 1.4
        },
        "upstream_cx_rx_bytes_total": {
          "value": 24287115,
          "perSecond": 809559.74
        },
        "upstream_cx_total": {
          "value": 42,
          "perSecond": 1.4
        },
        "upstream_cx_tx_bytes_total": {
          "value": 6963075,
          "perSecond": 232099.41
        },
        "upstream_rq_pending_overflow": {
          "value": 358,
          "perSecond": 11.93
        },
        "upstream_rq_pending_total": {
          "value": 42,
          "perSecond": 1.4
        },
        "upstream_rq_total": {
          "value": 154735,
          "perSecond": 5157.76
        }
      }
    },
    {
      "testName": "scaling down httproutes to 100 with 20 routes per hostname",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-down",
      "throughput": 5162.75,
      "totalRequests": 154884,
      "latency": {
        "min": 391,
        "mean": 6928,
        "max": 77328,
        "pstdev": 11972,
        "percentiles": {
          "p50": 3221,
          "p75": 5293,
          "p80": 6082,
          "p90": 10089,
          "p95": 46329,
          "p99": 53424,
          "p999": 61562
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 197.94,
            "max": 231.62,
            "mean": 219.46
          },
          "cpu": {
            "min": 0.73,
            "max": 41.4,
            "mean": 5.76
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 127.8,
            "max": 128.01,
            "mean": 127.89
          },
          "cpu": {
            "min": 0,
            "max": 100.13,
            "mean": 3.03
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38,
      "counters": {
        "benchmark.http_2xx": {
          "value": 154884,
          "perSecond": 5162.75
        },
        "benchmark.pool_overflow": {
          "value": 362,
          "perSecond": 12.07
        },
        "cluster_manager.cluster_added": {
          "value": 4,
          "perSecond": 0.13
        },
        "default.total_match_count": {
          "value": 4,
          "perSecond": 0.13
        },
        "membership_change": {
          "value": 4,
          "perSecond": 0.13
        },
        "runtime.load_success": {
          "value": 1,
          "perSecond": 0.03
        },
        "runtime.override_dir_not_exists": {
          "value": 1,
          "perSecond": 0.03
        },
        "upstream_cx_http1_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_rx_bytes_total": {
          "value": 24316788,
          "perSecond": 810551.95
        },
        "upstream_cx_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_tx_bytes_total": {
          "value": 6971445,
          "perSecond": 232379.31
        },
        "upstream_rq_pending_overflow": {
          "value": 362,
          "perSecond": 12.07
        },
        "upstream_rq_pending_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_rq_total": {
          "value": 154921,
          "perSecond": 5163.98
        }
      }
    },
    {
      "testName": "scaling down httproutes to 300 with 60 routes per hostname",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-down",
      "throughput": 5164.55,
      "totalRequests": 154939,
      "latency": {
        "min": 363,
        "mean": 6813,
        "max": 123748,
        "pstdev": 12135,
        "percentiles": {
          "p50": 3132,
          "p75": 5066,
          "p80": 5786,
          "p90": 9349,
          "p95": 46854,
          "p99": 55068,
          "p999": 64952
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 186.99,
            "max": 214.7,
            "mean": 194.88
          },
          "cpu": {
            "min": 0.73,
            "max": 34.27,
            "mean": 4.92
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 130.14,
            "max": 130.62,
            "mean": 130.27
          },
          "cpu": {
            "min": 0,
            "max": 99.81,
            "mean": 11.5
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37,
      "counters": {
        "benchmark.http_2xx": {
          "value": 154939,
          "perSecond": 5164.55
        },
        "benchmark.pool_overflow": {
          "value": 363,
          "perSecond": 12.1
        },
        "cluster_manager.cluster_added": {
          "value": 4,
          "perSecond": 0.13
        },
        "default.total_match_count": {
          "value": 4,
          "perSecond": 0.13
        },
        "membership_change": {
          "value": 4,
          "perSecond": 0.13
        },
        "runtime.load_success": {
          "value": 1,
          "perSecond": 0.03
        },
        "runtime.override_dir_not_exists": {
          "value": 1,
          "perSecond": 0.03
        },
        "upstream_cx_http1_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_cx_rx_bytes_total": {
          "value": 24325423,
          "perSecond": 810834.78
        },
        "upstream_cx_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_cx_tx_bytes_total": {
          "value": 6972975,
          "perSecond": 232428.87
        },
        "upstream_rq_pending_overflow": {
          "value": 363,
          "perSecond": 12.1
        },
        "upstream_rq_pending_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_rq_total": {
          "value": 154955,
          "perSecond": 5165.09
        }
      }
    },
    {
      "testName": "scaling down httproutes to 500 with 100 routes per hostname",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-down",
      "throughput": 5000.43,
      "totalRequests": 150016,
      "latency": {
        "min": 376,
        "mean": 7369,
        "max": 108912,
        "pstdev": 12666,
        "percentiles": {
          "p50": 3291,
          "p75": 5792,
          "p80": 6722,
          "p90": 11795,
          "p95": 47331,
          "p99": 55957,
          "p999": 71471
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 203.55,
            "max": 219.59,
            "mean": 207.71
          },
          "cpu": {
            "min": 0.13,
            "max": 1.33,
            "mean": 0.86
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 130.13,
            "max": 130.42,
            "mean": 130.26
          },
          "cpu": {
            "min": 0,
            "max": 99.81,
            "mean": 13.03
          }
        }
      },
      "poolOverflow": 361,
      "upstreamConnections": 39,
      "counters": {
        "benchmark.http_2xx": {
          "value": 150016,
          "perSecond": 5000.43
        },
        "benchmark.pool_overflow": {
          "value": 361,
          "perSecond": 12.03
        },
        "cluster_manager.cluster_added": {
          "value": 4,
          "perSecond": 0.13
        },
        "default.total_match_count": {
          "value": 4,
          "perSecond": 0.13
        },
        "membership_change": {
          "value": 4,
          "perSecond": 0.13
        },
        "runtime.load_success": {
          "value": 1,
          "perSecond": 0.03
        },
        "runtime.override_dir_not_exists": {
          "value": 1,
          "perSecond": 0.03
        },
        "upstream_cx_http1_total": {
          "value": 39,
          "perSecond": 1.3
        },
        "upstream_cx_rx_bytes_total": {
          "value": 23552512,
          "perSecond": 785068.13
        },
        "upstream_cx_total": {
          "value": 39,
          "perSecond": 1.3
        },
        "upstream_cx_tx_bytes_total": {
          "value": 6752160,
          "perSecond": 225067.53
        },
        "upstream_rq_pending_overflow": {
          "value": 361,
          "perSecond": 12.03
        },
        "upstream_rq_pending_total": {
          "value": 39,
          "perSecond": 1.3
        },
        "upstream_rq_total": {
          "value": 150048,
          "perSecond": 5001.5
        }
      }
    }
  ]
};

export default benchmarkData;
