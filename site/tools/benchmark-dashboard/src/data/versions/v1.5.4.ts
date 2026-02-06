import { TestSuite } from '../types';

// Benchmark data extracted from markdown report for version 1.5.4
// Generated on 2025-11-21T02:42:19.321Z

export const benchmarkData: TestSuite = {
  "metadata": {
    "version": "1.5.4",
    "runId": "1.5.4-1763692939320",
    "date": "2025-11-21T02:42:19.320Z",
    "environment": "GitHub CI",
    "description": "Benchmark results for version 1.5.4",
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
      "throughput": 5261.72,
      "totalRequests": 157854,
      "latency": {
        "min": 386,
        "mean": 6996,
        "max": 92798,
        "pstdev": 11423,
        "percentiles": {
          "p50": 3467,
          "p75": 5796,
          "p80": 6627,
          "p90": 10950,
          "p95": 42975,
          "p99": 53620,
          "p999": 61476
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 112.64,
            "max": 132.9,
            "mean": 127.62
          },
          "cpu": {
            "min": 0.13,
            "max": 1.13,
            "mean": 0.44
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 0,
            "max": 26.89,
            "mean": 22.19
          },
          "cpu": {
            "min": 0,
            "max": 99.78,
            "mean": 11.73
          }
        }
      },
      "poolOverflow": 361,
      "upstreamConnections": 39,
      "counters": {
        "benchmark.http_2xx": {
          "value": 157854,
          "perSecond": 5261.72
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
          "value": 24783078,
          "perSecond": 826090.43
        },
        "upstream_cx_total": {
          "value": 39,
          "perSecond": 1.3
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7104105,
          "perSecond": 236800.01
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
          "value": 157869,
          "perSecond": 5262.22
        }
      }
    },
    {
      "testName": "scaling up httproutes to 50 with 10 routes per hostname",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-up",
      "throughput": 5190.53,
      "totalRequests": 155722,
      "latency": {
        "min": 387,
        "mean": 6792,
        "max": 87822,
        "pstdev": 11838,
        "percentiles": {
          "p50": 3189,
          "p75": 5155,
          "p80": 5868,
          "p90": 9620,
          "p95": 45627,
          "p99": 53573,
          "p999": 62234
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 132.9,
            "max": 142.34,
            "mean": 138.63
          },
          "cpu": {
            "min": 0.27,
            "max": 4.4,
            "mean": 0.84
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 26.59,
            "max": 30.92,
            "mean": 30.2
          },
          "cpu": {
            "min": 0,
            "max": 100.03,
            "mean": 11.03
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37,
      "counters": {
        "benchmark.http_2xx": {
          "value": 155722,
          "perSecond": 5190.53
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
          "value": 24448354,
          "perSecond": 814913.58
        },
        "upstream_cx_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7008300,
          "perSecond": 233600.96
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
          "value": 155740,
          "perSecond": 5191.13
        }
      }
    },
    {
      "testName": "scaling up httproutes to 100 with 20 routes per hostname",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-up",
      "throughput": 5212.86,
      "totalRequests": 156386,
      "latency": {
        "min": 374,
        "mean": 7452,
        "max": 85016,
        "pstdev": 12613,
        "percentiles": {
          "p50": 3453,
          "p75": 5808,
          "p80": 6677,
          "p90": 11296,
          "p95": 47863,
          "p99": 55584,
          "p999": 64251
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 136.62,
            "max": 139.78,
            "mean": 139
          },
          "cpu": {
            "min": 0.33,
            "max": 4.87,
            "mean": 0.94
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 30.65,
            "max": 37.21,
            "mean": 36.32
          },
          "cpu": {
            "min": 0,
            "max": 46.44,
            "mean": 1.81
          }
        }
      },
      "poolOverflow": 359,
      "upstreamConnections": 41,
      "counters": {
        "benchmark.http_2xx": {
          "value": 156386,
          "perSecond": 5212.86
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
          "value": 24552602,
          "perSecond": 818418.84
        },
        "upstream_cx_total": {
          "value": 41,
          "perSecond": 1.37
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7039215,
          "perSecond": 234640.15
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
          "value": 156427,
          "perSecond": 5214.23
        }
      }
    },
    {
      "testName": "scaling up httproutes to 300 with 60 routes per hostname",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-up",
      "throughput": 5198.93,
      "totalRequests": 155968,
      "latency": {
        "min": 381,
        "mean": 6749,
        "max": 77881,
        "pstdev": 11891,
        "percentiles": {
          "p50": 3164,
          "p75": 5134,
          "p80": 5828,
          "p90": 9403,
          "p95": 46280,
          "p99": 54018,
          "p999": 64313
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 139.78,
            "max": 155.04,
            "mean": 152.03
          },
          "cpu": {
            "min": 0.4,
            "max": 20.87,
            "mean": 2.55
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 36.77,
            "max": 59.16,
            "mean": 55.93
          },
          "cpu": {
            "min": 0,
            "max": 99.96,
            "mean": 4.33
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37,
      "counters": {
        "benchmark.http_2xx": {
          "value": 155968,
          "perSecond": 5198.93
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
          "value": 24486976,
          "perSecond": 816232
        },
        "upstream_cx_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7020225,
          "perSecond": 234007.35
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
          "value": 156005,
          "perSecond": 5200.16
        }
      }
    },
    {
      "testName": "scaling up httproutes to 500 with 100 routes per hostname",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-up",
      "throughput": 5176.03,
      "totalRequests": 155281,
      "latency": {
        "min": 387,
        "mean": 6938,
        "max": 113328,
        "pstdev": 12104,
        "percentiles": {
          "p50": 3172,
          "p75": 5303,
          "p80": 6058,
          "p90": 10265,
          "p95": 46184,
          "p99": 54487,
          "p999": 67004
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 151.23,
            "max": 165.62,
            "mean": 162.13
          },
          "cpu": {
            "min": 0.4,
            "max": 19.13,
            "mean": 2.41
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 58.96,
            "max": 79.86,
            "mean": 77.36
          },
          "cpu": {
            "min": 0,
            "max": 100.12,
            "mean": 7.47
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38,
      "counters": {
        "benchmark.http_2xx": {
          "value": 155281,
          "perSecond": 5176.03
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
          "value": 24379117,
          "perSecond": 812636.63
        },
        "upstream_cx_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_tx_bytes_total": {
          "value": 6989355,
          "perSecond": 232978.33
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
          "value": 155319,
          "perSecond": 5177.3
        }
      }
    },
    {
      "testName": "scaling up httproutes to 1000 with 200 routes per hostname",
      "routes": 1000,
      "routesPerHostname": 200,
      "phase": "scaling-up",
      "throughput": 5137,
      "totalRequests": 154113,
      "latency": {
        "min": 359,
        "mean": 8124,
        "max": 101470,
        "pstdev": 13230,
        "percentiles": {
          "p50": 3824,
          "p75": 6312,
          "p80": 7277,
          "p90": 13760,
          "p95": 48394,
          "p99": 58728,
          "p999": 70451
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 189.07,
            "max": 235.64,
            "mean": 194.15
          },
          "cpu": {
            "min": 0.47,
            "max": 33.73,
            "mean": 5.63
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 79.22,
            "max": 135.98,
            "mean": 130.13
          },
          "cpu": {
            "min": 0,
            "max": 99.94,
            "mean": 11.21
          }
        }
      },
      "poolOverflow": 356,
      "upstreamConnections": 44,
      "counters": {
        "benchmark.http_2xx": {
          "value": 154113,
          "perSecond": 5137
        },
        "benchmark.pool_overflow": {
          "value": 356,
          "perSecond": 11.87
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
          "value": 44,
          "perSecond": 1.47
        },
        "upstream_cx_rx_bytes_total": {
          "value": 24195741,
          "perSecond": 806508.95
        },
        "upstream_cx_total": {
          "value": 44,
          "perSecond": 1.47
        },
        "upstream_cx_tx_bytes_total": {
          "value": 6936390,
          "perSecond": 231208.49
        },
        "upstream_rq_pending_overflow": {
          "value": 356,
          "perSecond": 11.87
        },
        "upstream_rq_pending_total": {
          "value": 44,
          "perSecond": 1.47
        },
        "upstream_rq_total": {
          "value": 154142,
          "perSecond": 5137.97
        }
      }
    },
    {
      "testName": "scaling down httproutes to 10 with 2 routes per hostname",
      "routes": 10,
      "routesPerHostname": 2,
      "phase": "scaling-down",
      "throughput": 5329.16,
      "totalRequests": 159878,
      "latency": {
        "min": 376,
        "mean": 6806,
        "max": 98381,
        "pstdev": 11627,
        "percentiles": {
          "p50": 3310,
          "p75": 5297,
          "p80": 6029,
          "p90": 9460,
          "p95": 45320,
          "p99": 52797,
          "p999": 59246
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 165.02,
            "max": 216.97,
            "mean": 203.73
          },
          "cpu": {
            "min": 0.6,
            "max": 3.13,
            "mean": 1.01
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 128.97,
            "max": 129.21,
            "mean": 129.1
          },
          "cpu": {
            "min": 0,
            "max": 52.71,
            "mean": 5.05
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38,
      "counters": {
        "benchmark.http_2xx": {
          "value": 159878,
          "perSecond": 5329.16
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
          "value": 25100846,
          "perSecond": 836678.51
        },
        "upstream_cx_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7195860,
          "perSecond": 239857.31
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
          "value": 159908,
          "perSecond": 5330.16
        }
      }
    },
    {
      "testName": "scaling down httproutes to 50 with 10 routes per hostname",
      "routes": 50,
      "routesPerHostname": 10,
      "phase": "scaling-down",
      "throughput": 5287.65,
      "totalRequests": 158634,
      "latency": {
        "min": 358,
        "mean": 7578,
        "max": 118231,
        "pstdev": 12572,
        "percentiles": {
          "p50": 3603,
          "p75": 5831,
          "p80": 6664,
          "p90": 11794,
          "p95": 47650,
          "p99": 55255,
          "p999": 65124
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 143.82,
            "max": 154.45,
            "mean": 146.42
          },
          "cpu": {
            "min": 0.6,
            "max": 4.47,
            "mean": 1.17
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 128.95,
            "max": 129.23,
            "mean": 129.05
          },
          "cpu": {
            "min": 0,
            "max": 100.12,
            "mean": 16.99
          }
        }
      },
      "poolOverflow": 358,
      "upstreamConnections": 42,
      "counters": {
        "benchmark.http_2xx": {
          "value": 158634,
          "perSecond": 5287.65
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
          "value": 24905538,
          "perSecond": 830160.73
        },
        "upstream_cx_total": {
          "value": 42,
          "perSecond": 1.4
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7139250,
          "perSecond": 237968.16
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
          "value": 158650,
          "perSecond": 5288.18
        }
      }
    },
    {
      "testName": "scaling down httproutes to 100 with 20 routes per hostname",
      "routes": 100,
      "routesPerHostname": 20,
      "phase": "scaling-down",
      "throughput": 5182.59,
      "totalRequests": 155481,
      "latency": {
        "min": 390,
        "mean": 6947,
        "max": 95182,
        "pstdev": 11994,
        "percentiles": {
          "p50": 3230,
          "p75": 5374,
          "p80": 6156,
          "p90": 10277,
          "p95": 46147,
          "p99": 53874,
          "p999": 64235
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 144.34,
            "max": 159.4,
            "mean": 150.23
          },
          "cpu": {
            "min": 0.53,
            "max": 12.73,
            "mean": 1.84
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 128.98,
            "max": 135.66,
            "mean": 129.5
          },
          "cpu": {
            "min": 0,
            "max": 100,
            "mean": 14.54
          }
        }
      },
      "poolOverflow": 362,
      "upstreamConnections": 38,
      "counters": {
        "benchmark.http_2xx": {
          "value": 155481,
          "perSecond": 5182.59
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
          "value": 24410517,
          "perSecond": 813667.08
        },
        "upstream_cx_total": {
          "value": 38,
          "perSecond": 1.27
        },
        "upstream_cx_tx_bytes_total": {
          "value": 6998130,
          "perSecond": 233266.18
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
          "value": 155514,
          "perSecond": 5183.69
        }
      }
    },
    {
      "testName": "scaling down httproutes to 300 with 60 routes per hostname",
      "routes": 300,
      "routesPerHostname": 60,
      "phase": "scaling-down",
      "throughput": 5243.55,
      "totalRequests": 157311,
      "latency": {
        "min": 377,
        "mean": 6697,
        "max": 79204,
        "pstdev": 11447,
        "percentiles": {
          "p50": 3250,
          "p75": 5266,
          "p80": 5976,
          "p90": 9577,
          "p95": 44281,
          "p99": 52840,
          "p999": 63244
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 149.68,
            "max": 165.16,
            "mean": 156.89
          },
          "cpu": {
            "min": 0.53,
            "max": 13.46,
            "mean": 2.24
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 135.65,
            "max": 135.89,
            "mean": 135.73
          },
          "cpu": {
            "min": 0,
            "max": 99.96,
            "mean": 14.81
          }
        }
      },
      "poolOverflow": 363,
      "upstreamConnections": 37,
      "counters": {
        "benchmark.http_2xx": {
          "value": 157311,
          "perSecond": 5243.55
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
          "value": 24697827,
          "perSecond": 823237.67
        },
        "upstream_cx_total": {
          "value": 37,
          "perSecond": 1.23
        },
        "upstream_cx_tx_bytes_total": {
          "value": 7080075,
          "perSecond": 235995.84
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
          "value": 157335,
          "perSecond": 5244.35
        }
      }
    },
    {
      "testName": "scaling down httproutes to 500 with 100 routes per hostname",
      "routes": 500,
      "routesPerHostname": 100,
      "phase": "scaling-down",
      "throughput": 4881.7,
      "totalRequests": 146451,
      "latency": {
        "min": 364,
        "mean": 5040,
        "max": 106377,
        "pstdev": 10078,
        "percentiles": {
          "p50": 2332,
          "p75": 3922,
          "p80": 4464,
          "p90": 6588,
          "p95": 31911,
          "p99": 50827,
          "p999": 62269
        }
      },
      "resources": {
        "envoyGateway": {
          "memory": {
            "min": 158.79,
            "max": 189.29,
            "mean": 167.36
          },
          "cpu": {
            "min": 0.53,
            "max": 37.46,
            "mean": 4.26
          }
        },
        "envoyProxy": {
          "memory": {
            "min": 135.64,
            "max": 136.08,
            "mean": 135.74
          },
          "cpu": {
            "min": 0,
            "max": 98.66,
            "mean": 11.59
          }
        }
      },
      "poolOverflow": 374,
      "upstreamConnections": 26,
      "counters": {
        "benchmark.http_2xx": {
          "value": 146451,
          "perSecond": 4881.7
        },
        "benchmark.pool_overflow": {
          "value": 374,
          "perSecond": 12.47
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
          "value": 26,
          "perSecond": 0.87
        },
        "upstream_cx_rx_bytes_total": {
          "value": 22992807,
          "perSecond": 766426.81
        },
        "upstream_cx_total": {
          "value": 26,
          "perSecond": 0.87
        },
        "upstream_cx_tx_bytes_total": {
          "value": 6591465,
          "perSecond": 219715.47
        },
        "upstream_rq_pending_overflow": {
          "value": 374,
          "perSecond": 12.47
        },
        "upstream_rq_pending_total": {
          "value": 26,
          "perSecond": 0.87
        },
        "upstream_rq_total": {
          "value": 146477,
          "perSecond": 4882.57
        }
      }
    }
  ]
};

export default benchmarkData;
