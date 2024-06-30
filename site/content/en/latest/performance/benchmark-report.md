---
title: Benchmark Report
description: The latest benchmarking report of Envoy Gateway.
---

{{% alert title="Where does this report come from?" color="warning" %}}

This report is auto-generated after running benchmark test, 
you can refer to [how to run benchmark test](./benchmark#run-benchmark-test) to learn more.

{{% /alert %}}

---

Benchmark test settings:

|RPS  |Connections|Duration (Seconds)|CPU Limits (m)|Memory Limits (MiB)|
|-    |-          |-                 |-             |-                  |
|10000|100        |30                |1000          |2000               |

## Test: ScaleHTTPRoute

Fixed one Gateway and different scales of HTTPRoutes.


### Results

Click to see the full results.


<details>
<summary>scale-up-httproutes-10</summary>

```plaintext
Warning: 29 02:48:14.546][1][warning][misc] [external/envoy/source/common/protobuf/message_validator_impl.cc:21] Deprecated field: type envoy.config.core.v3.HeaderValueOption Using deprecated option 'envoy.config.core.v3.HeaderValueOption.append' from file base.proto. This configuration will be removed from Envoy soon. Please see https://www.envoyproxy.io/docs/envoy/latest/version_history/version_history for details. If continued use of this field is absolutely necessary, see https://www.envoyproxy.io/docs/envoy/latest/configuration/operations/runtime#using-runtime-overrides-for-deprecated-features for how to apply a temporary and highly discouraged override.
[02:48:14.547222][1][I] Detected 4 (v)CPUs with affinity..
[02:48:14.547235][1][I] Starting 4 threads / event loops. Time limit: 30 seconds.
[02:48:14.547237][1][I] Global targets: 400 connections and 40000 calls per second.
[02:48:14.547239][1][I]    (Per-worker targets: 100 connections and 10000 calls per second)
[02:48:45.249233][18][I] Stopping after 30000 ms. Initiated: 63756 / Completed: 63740. (Completion rate was 2124.6656043338644 per second.)
[02:48:45.249278][21][I] Stopping after 30000 ms. Initiated: 13846 / Completed: 13844. (Completion rate was 461.4664974622843 per second.)
[02:48:45.249278][19][I] Stopping after 30000 ms. Initiated: 79156 / Completed: 79137. (Completion rate was 2637.8965707344582 per second.)
[02:48:45.249494][23][I] Stopping after 30000 ms. Initiated: 20224 / Completed: 20222. (Completion rate was 674.0621055130861 per second.)
Nighthawk - A layer 7 protocol benchmarking tool.

benchmark_http_client.latency_2xx (176582 samples)
  min: 0s 000ms 358us | mean: 0s 006ms 272us | max: 0s 070ms 623us | pstdev: 0s 011ms 440us

  Percentile  Count       Value          
  0.5         88291       0s 002ms 898us 
  0.75        132439      0s 004ms 643us 
  0.8         141270      0s 005ms 297us 
  0.9         158924      0s 008ms 790us 
  0.95        167754      0s 045ms 185us 
  0.990625    174928      0s 054ms 206us 
  0.99902344  176410      0s 059ms 482us 

Queueing and connection setup latency (176621 samples)
  min: 0s 000ms 001us | mean: 0s 000ms 012us | max: 0s 024ms 204us | pstdev: 0s 000ms 165us
[02:48:50.882431][18][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.

  Percentile  Count       Value          
  0.5         88528       0s 000ms 003us 
  0.75        132525      0s 000ms 008us 
  0.8         141340      0s 000ms 009us 
  0.9         159044      0s 000ms 010us 
  0.95        167794      0s 000ms 011us 
  0.990625    174966      0s 000ms 077us 
  0.99902344  176449      0s 001ms 431us 

Request start to response end (176582 samples)
  min: 0s 000ms 358us | mean: 0s 006ms 271us | max: 0s 070ms 623us | pstdev: 0s 011ms 440us

  Percentile  Count       Value          
  0.5         88293       0s 002ms 897us 
  0.75        132441      0s 004ms 643us 
  0.8         141267      0s 005ms 296us 
  0.9         158924      0s 008ms 788us 
  0.95        167754      0s 045ms 185us 
  0.990625    174928      0s 054ms 204us 
  0.99902344  176410      0s 059ms 473us 

Response body size in bytes (176582 samples)
  min: 10 | mean: 10 | max: 10 | pstdev: 0

Response header size in bytes (176582 samples)
  min: 110 | mean: 110 | max: 110 | pstdev: 0

Blocking. Results are skewed when significant numbers are reported here. (53988 samples)
  min: 0s 000ms 035us | mean: 0s 002ms 220us | max: 0s 067ms 391us | pstdev: 0s 006ms 778us

  Percentile  Count       Value          
  0.5         26994       0s 000ms 766us 
  0.75        40492       0s 001ms 732us 
  0.8         43191       0s 002ms 062us 
  0.9         48590       0s 003ms 222us 
  0.95        51289       0s 004ms 958us 
  0.990625    53482       0s 048ms 125us 
  0.99902344  53936       0s 055ms 597us 

Initiation to completion (176943 samples)
  min: 0s 000ms 003us | mean: 0s 006ms 351us | max: 0s 071ms 114us | pstdev: 0s 011ms 455us

  Percentile  Count       Value          
  0.5         88472       0s 002ms 960us 
  0.75        132709      0s 004ms 735us 
  0.8         141556      0s 005ms 408us 
  0.9         159249      0s 009ms 012us 
  0.95        168096      0s 045ms 185us 
  0.990625    175285      0s 054ms 288us 
  0.99902344  176771      0s 059ms 691us 

Counter                                 Value       Per second
benchmark.http_2xx                      176582      5886.05
benchmark.pool_overflow                 361         12.03
cluster_manager.cluster_added           4           0.13
default.total_match_count               4           0.13
membership_change                       4           0.13
runtime.load_success                    1           0.03
runtime.override_dir_not_exists         1           0.03
upstream_cx_http1_total                 39          1.30
upstream_cx_rx_bytes_total              27723374    924110.39
upstream_cx_total                       39          1.30
upstream_cx_tx_bytes_total              7594703     253156.20
upstream_rq_pending_overflow            361         12.03
upstream_rq_pending_total               39          1.30
upstream_rq_total                       176621      5887.35

[02:48:55.883737][19][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[02:49:00.884904][21][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[02:49:05.885509][23][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[02:49:05.886786][1][I] Done.

```

</details>

<details>
<summary>scale-up-httproutes-50</summary>

```plaintext
Warning: 29 02:49:16.557][1][warning][misc] [external/envoy/source/common/protobuf/message_validator_impl.cc:21] Deprecated field: type envoy.config.core.v3.HeaderValueOption Using deprecated option 'envoy.config.core.v3.HeaderValueOption.append' from file base.proto. This configuration will be removed from Envoy soon. Please see https://www.envoyproxy.io/docs/envoy/latest/version_history/version_history for details. If continued use of this field is absolutely necessary, see https://www.envoyproxy.io/docs/envoy/latest/configuration/operations/runtime#using-runtime-overrides-for-deprecated-features for how to apply a temporary and highly discouraged override.
[02:49:16.558120][1][I] Detected 4 (v)CPUs with affinity..
[02:49:16.558133][1][I] Starting 4 threads / event loops. Time limit: 30 seconds.
[02:49:16.558136][1][I] Global targets: 400 connections and 40000 calls per second.
[02:49:16.558137][1][I]    (Per-worker targets: 100 connections and 10000 calls per second)
[02:49:47.260193][18][I] Stopping after 30000 ms. Initiated: 54652 / Completed: 54639. (Completion rate was 1821.296539536575 per second.)
[02:49:47.260240][23][I] Stopping after 30000 ms. Initiated: 47061 / Completed: 47051. (Completion rate was 1568.3648891864589 per second.)
[02:49:47.260255][21][I] Stopping after 30000 ms. Initiated: 19683 / Completed: 19680. (Completion rate was 655.998381870658 per second.)
[02:49:47.260634][19][I] Stopping after 30000 ms. Initiated: 52114 / Completed: 52104. (Completion rate was 1736.7723853190735 per second.)
[02:49:52.907560][18][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
Nighthawk - A layer 7 protocol benchmarking tool.

benchmark_http_client.latency_2xx (173112 samples)
  min: 0s 000ms 390us | mean: 0s 006ms 344us | max: 0s 082ms 546us | pstdev: 0s 012ms 275us

  Percentile  Count       Value          
  0.5         86560       0s 002ms 771us 
  0.75        129838      0s 004ms 367us 
  0.8         138490      0s 004ms 925us 
  0.9         155803      0s 007ms 559us 
  0.95        164457      0s 049ms 076us 
  0.990625    171491      0s 055ms 959us 
  0.99902344  172943      0s 062ms 951us 

Queueing and connection setup latency (173148 samples)
  min: 0s 000ms 001us | mean: 0s 000ms 012us | max: 0s 027ms 359us | pstdev: 0s 000ms 173us

  Percentile  Count       Value          
  0.5         86613       0s 000ms 003us 
  0.75        129901      0s 000ms 008us 
  0.8         138585      0s 000ms 009us 
  0.9         155919      0s 000ms 010us 
  0.95        164492      0s 000ms 011us 
  0.990625    171525      0s 000ms 094us 
  0.99902344  172979      0s 001ms 449us 

Request start to response end (173112 samples)
  min: 0s 000ms 390us | mean: 0s 006ms 344us | max: 0s 082ms 546us | pstdev: 0s 012ms 275us

  Percentile  Count       Value          
  0.5         86562       0s 002ms 770us 
  0.75        129835      0s 004ms 366us 
  0.8         138493      0s 004ms 925us 
  0.9         155801      0s 007ms 558us 
  0.95        164457      0s 049ms 076us 
  0.990625    171491      0s 055ms 959us 
  0.99902344  172943      0s 062ms 951us 

Response body size in bytes (173112 samples)
  min: 10 | mean: 10 | max: 10 | pstdev: 0

Response header size in bytes (173112 samples)
  min: 110 | mean: 110 | max: 110 | pstdev: 0

Blocking. Results are skewed when significant numbers are reported here. (54045 samples)
  min: 0s 000ms 034us | mean: 0s 002ms 217us | max: 0s 080ms 617us | pstdev: 0s 007ms 139us

  Percentile  Count       Value          
  0.5         27023       0s 000ms 799us 
  0.75        40535       0s 001ms 657us 
  0.8         43236       0s 001ms 933us 
  0.9         48642       0s 002ms 955us 
  0.95        51343       0s 004ms 441us 
  0.990625    53539       0s 048ms 988us 
  0.99902344  53993       0s 055ms 539us 

Initiation to completion (173474 samples)
  min: 0s 000ms 002us | mean: 0s 006ms 420us | max: 0s 085ms 741us | pstdev: 0s 012ms 283us

  Percentile  Count       Value          
  0.5         86741       0s 002ms 832us 
  0.75        130106      0s 004ms 457us 
  0.8         138781      0s 005ms 030us 
  0.9         156127      0s 007ms 771us 
  0.95        164802      0s 049ms 082us 
  0.990625    171849      0s 056ms 018us 
  0.99902344  173305      0s 062ms 982us 

Counter                                 Value       Per second
benchmark.http_2xx                      173112      5770.37
benchmark.pool_overflow                 362         12.07
cluster_manager.cluster_added           4           0.13
default.total_match_count               4           0.13
membership_change                       4           0.13
runtime.load_success                    1           0.03
runtime.override_dir_not_exists         1           0.03
upstream_cx_http1_total                 38          1.27
upstream_cx_rx_bytes_total              27178584    905947.93
upstream_cx_total                       38          1.27
upstream_cx_tx_bytes_total              7445364     248177.47
upstream_rq_pending_overflow            362         12.07
upstream_rq_pending_total               38          1.27
upstream_rq_total                       173148      5771.57

[02:49:57.908877][19][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[02:50:02.909755][21][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[02:50:07.910451][23][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[02:50:07.911821][1][I] Done.

```

</details>

<details>
<summary>scale-up-httproutes-100</summary>

```plaintext
Warning: 29 02:50:24.428][1][warning][misc] [external/envoy/source/common/protobuf/message_validator_impl.cc:21] Deprecated field: type envoy.config.core.v3.HeaderValueOption Using deprecated option 'envoy.config.core.v3.HeaderValueOption.append' from file base.proto. This configuration will be removed from Envoy soon. Please see https://www.envoyproxy.io/docs/envoy/latest/version_history/version_history for details. If continued use of this field is absolutely necessary, see https://www.envoyproxy.io/docs/envoy/latest/configuration/operations/runtime#using-runtime-overrides-for-deprecated-features for how to apply a temporary and highly discouraged override.
[02:50:24.429389][1][I] Detected 4 (v)CPUs with affinity..
[02:50:24.429401][1][I] Starting 4 threads / event loops. Time limit: 30 seconds.
[02:50:24.429403][1][I] Global targets: 400 connections and 40000 calls per second.
[02:50:24.429404][1][I]    (Per-worker targets: 100 connections and 10000 calls per second)
[02:50:55.131171][18][I] Stopping after 30000 ms. Initiated: 67695 / Completed: 67679. (Completion rate was 2255.966516268899 per second.)
[02:50:55.131279][23][I] Stopping after 30000 ms. Initiated: 51702 / Completed: 51691. (Completion rate was 1723.0310359586188 per second.)
[02:50:55.131365][21][I] Stopping after 30000 ms. Initiated: 13596 / Completed: 13594. (Completion rate was 453.131052573702 per second.)
[02:50:55.131197][19][I] Stopping after 30000 ms. Initiated: 41261 / Completed: 41252. (Completion rate was 1375.0663458178528 per second.)
[02:51:00.765603][18][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
Nighthawk - A layer 7 protocol benchmarking tool.

benchmark_http_client.latency_2xx (173854 samples)
  min: 0s 000ms 371us | mean: 0s 006ms 326us | max: 0s 081ms 301us | pstdev: 0s 012ms 060us

  Percentile  Count       Value          
  0.5         86929       0s 002ms 870us 
  0.75        130391      0s 004ms 436us 
  0.8         139084      0s 004ms 988us 
  0.9         156469      0s 007ms 453us 
  0.95        165164      0s 048ms 289us 
  0.990625    172225      0s 055ms 343us 
  0.99902344  173685      0s 060ms 573us 

Queueing and connection setup latency (173892 samples)
  min: 0s 000ms 001us | mean: 0s 000ms 012us | max: 0s 022ms 520us | pstdev: 0s 000ms 151us

  Percentile  Count       Value          
  0.5         87000       0s 000ms 003us 
  0.75        130576      0s 000ms 009us 
  0.8         139114      0s 000ms 009us 
  0.9         156579      0s 000ms 010us 
  0.95        165215      0s 000ms 011us 
  0.990625    172262      0s 000ms 052us 
  0.99902344  173723      0s 001ms 574us 

Request start to response end (173854 samples)
  min: 0s 000ms 371us | mean: 0s 006ms 325us | max: 0s 081ms 301us | pstdev: 0s 012ms 060us

  Percentile  Count       Value          
  0.5         86928       0s 002ms 870us 
  0.75        130392      0s 004ms 435us 
  0.8         139085      0s 004ms 987us 
  0.9         156470      0s 007ms 451us 
  0.95        165164      0s 048ms 289us 
  0.990625    172225      0s 055ms 343us 
  0.99902344  173685      0s 060ms 573us 

Response body size in bytes (173854 samples)
  min: 10 | mean: 10 | max: 10 | pstdev: 0

Response header size in bytes (173854 samples)
  min: 110 | mean: 110 | max: 110 | pstdev: 0

Blocking. Results are skewed when significant numbers are reported here. (57112 samples)
  min: 0s 000ms 036us | mean: 0s 002ms 099us | max: 0s 066ms 932us | pstdev: 0s 006ms 814us

  Percentile  Count       Value          
  0.5         28556       0s 000ms 745us 
  0.75        42835       0s 001ms 585us 
  0.8         45690       0s 001ms 856us 
  0.9         51404       0s 002ms 898us 
  0.95        54257       0s 004ms 315us 
  0.990625    56577       0s 047ms 876us 
  0.99902344  57057       0s 054ms 708us 

Initiation to completion (174216 samples)
  min: 0s 000ms 002us | mean: 0s 006ms 394us | max: 0s 081ms 338us | pstdev: 0s 012ms 059us

  Percentile  Count       Value          
  0.5         87109       0s 002ms 927us 
  0.75        130664      0s 004ms 514us 
  0.8         139373      0s 005ms 082us 
  0.9         156795      0s 007ms 623us 
  0.95        165506      0s 048ms 308us 
  0.990625    172584      0s 055ms 402us 
  0.99902344  174046      0s 060ms 628us 

Counter                                 Value       Per second
benchmark.http_2xx                      173854      5795.12
benchmark.pool_overflow                 362         12.07
cluster_manager.cluster_added           4           0.13
default.total_match_count               4           0.13
membership_change                       4           0.13
runtime.load_success                    1           0.03
runtime.override_dir_not_exists         1           0.03
upstream_cx_http1_total                 38          1.27
upstream_cx_rx_bytes_total              27295078    909834.41
upstream_cx_total                       38          1.27
upstream_cx_tx_bytes_total              7477356     249244.78
upstream_rq_pending_overflow            362         12.07
upstream_rq_pending_total               38          1.27
upstream_rq_total                       173892      5796.39

[02:51:05.767184][19][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[02:51:10.768179][21][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[02:51:15.768984][23][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[02:51:15.770420][1][I] Done.

```

</details>

<details>
<summary>scale-up-httproutes-300</summary>

```plaintext
Warning: 29 02:54:16.918][1][warning][misc] [external/envoy/source/common/protobuf/message_validator_impl.cc:21] Deprecated field: type envoy.config.core.v3.HeaderValueOption Using deprecated option 'envoy.config.core.v3.HeaderValueOption.append' from file base.proto. This configuration will be removed from Envoy soon. Please see https://www.envoyproxy.io/docs/envoy/latest/version_history/version_history for details. If continued use of this field is absolutely necessary, see https://www.envoyproxy.io/docs/envoy/latest/configuration/operations/runtime#using-runtime-overrides-for-deprecated-features for how to apply a temporary and highly discouraged override.
[02:54:16.919515][1][I] Detected 4 (v)CPUs with affinity..
[02:54:16.919526][1][I] Starting 4 threads / event loops. Time limit: 30 seconds.
[02:54:16.919528][1][I] Global targets: 400 connections and 40000 calls per second.
[02:54:16.919530][1][I]    (Per-worker targets: 100 connections and 10000 calls per second)
[02:54:47.621901][18][I] Stopping after 30000 ms. Initiated: 30621 / Completed: 30615. (Completion rate was 1020.4993877003675 per second.)
[02:54:47.621932][20][I] Stopping after 30000 ms. Initiated: 57498 / Completed: 57485. (Completion rate was 1916.1664750500192 per second.)
[02:54:47.622002][22][I] Stopping after 30000 ms. Initiated: 18978 / Completed: 18975. (Completion rate was 632.4993464173419 per second.)
[02:54:47.622083][19][I] Stopping after 30000 ms. Initiated: 62489 / Completed: 62473. (Completion rate was 2082.420977635533 per second.)
Nighthawk - A layer 7 protocol benchmarking tool.

benchmark_http_client.latency_2xx (169186 samples)
  min: 0s 000ms 377us | mean: 0s 006ms 498us | max: 0s 076ms 939us | pstdev: 0s 012ms 515us

  Percentile  Count       Value          
  0.5         84593       0s 002ms 805us 
  0.75        126890      0s 004ms 518us 
  0.8         135349      0s 005ms 134us 
  0.9         152268      0s 007ms 831us 
  0.95        160728      0s 049ms 625us 
  0.990625    167601      0s 056ms 295us 
  0.99902344  169021      0s 064ms 280us 

Queueing and connection setup latency (169224 samples)
  min: 0s 000ms 001us | mean: 0s 000ms 012us | max: 0s 023ms 327us | pstdev: 0s 000ms 159us

  Percentile  Count       Value          
  0.5         84850       0s 000ms 003us 
[02:54:53.257557][18][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
  0.75        126988      0s 000ms 008us 
  0.8         135486      0s 000ms 009us 
  0.9         152379      0s 000ms 009us 
  0.95        160789      0s 000ms 011us 
  0.990625    167638      0s 000ms 094us 
  0.99902344  169059      0s 001ms 577us 

Request start to response end (169186 samples)
  min: 0s 000ms 376us | mean: 0s 006ms 497us | max: 0s 076ms 939us | pstdev: 0s 012ms 514us

  Percentile  Count       Value          
  0.5         84597       0s 002ms 805us 
  0.75        126895      0s 004ms 517us 
  0.8         135352      0s 005ms 133us 
  0.9         152269      0s 007ms 830us 
  0.95        160727      0s 049ms 618us 
  0.990625    167601      0s 056ms 293us 
  0.99902344  169021      0s 064ms 280us 

Response body size in bytes (169186 samples)
  min: 10 | mean: 10 | max: 10 | pstdev: 0

Response header size in bytes (169186 samples)
  min: 110 | mean: 110 | max: 110 | pstdev: 0

Blocking. Results are skewed when significant numbers are reported here. (56249 samples)
  min: 0s 000ms 033us | mean: 0s 002ms 131us | max: 0s 073ms 695us | pstdev: 0s 007ms 030us

  Percentile  Count       Value          
  0.5         28125       0s 000ms 745us 
  0.75        42188       0s 001ms 540us 
  0.8         45000       0s 001ms 812us 
  0.9         50625       0s 002ms 844us 
  0.95        53437       0s 004ms 336us 
  0.990625    55723       0s 048ms 930us 
  0.99902344  56195       0s 055ms 592us 

Initiation to completion (169548 samples)
  min: 0s 000ms 005us | mean: 0s 006ms 559us | max: 0s 076ms 967us | pstdev: 0s 012ms 513us

  Percentile  Count       Value          
  0.5         84780       0s 002ms 862us 
  0.75        127162      0s 004ms 606us 
  0.8         135642      0s 005ms 239us 
  0.9         152594      0s 007ms 989us 
  0.95        161071      0s 049ms 643us 
  0.990625    167959      0s 056ms 363us 
  0.99902344  169383      0s 064ms 391us 

Counter                                 Value       Per second
benchmark.http_2xx                      169186      5639.52
benchmark.pool_overflow                 362         12.07
cluster_manager.cluster_added           4           0.13
default.total_match_count               4           0.13
membership_change                       4           0.13
runtime.load_success                    1           0.03
runtime.override_dir_not_exists         1           0.03
upstream_cx_http1_total                 38          1.27
upstream_cx_rx_bytes_total              26562202    885405.02
upstream_cx_total                       38          1.27
upstream_cx_tx_bytes_total              7276632     242553.93
upstream_rq_pending_overflow            362         12.07
upstream_rq_pending_total               38          1.27
upstream_rq_total                       169224      5640.79

[02:54:58.258626][19][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[02:55:03.259870][20][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[02:55:08.260826][22][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[02:55:08.262048][1][I] Done.

```

</details>

<details>
<summary>scale-up-httproutes-500</summary>

```plaintext
Warning: 29 03:00:41.969][1][warning][misc] [external/envoy/source/common/protobuf/message_validator_impl.cc:21] Deprecated field: type envoy.config.core.v3.HeaderValueOption Using deprecated option 'envoy.config.core.v3.HeaderValueOption.append' from file base.proto. This configuration will be removed from Envoy soon. Please see https://www.envoyproxy.io/docs/envoy/latest/version_history/version_history for details. If continued use of this field is absolutely necessary, see https://www.envoyproxy.io/docs/envoy/latest/configuration/operations/runtime#using-runtime-overrides-for-deprecated-features for how to apply a temporary and highly discouraged override.
[03:00:41.970303][1][I] Detected 4 (v)CPUs with affinity..
[03:00:41.970315][1][I] Starting 4 threads / event loops. Time limit: 30 seconds.
[03:00:41.970318][1][I] Global targets: 400 connections and 40000 calls per second.
[03:00:41.970319][1][I]    (Per-worker targets: 100 connections and 10000 calls per second)
[03:01:12.672360][23][I] Stopping after 30000 ms. Initiated: 54564 / Completed: 54557. (Completion rate was 1818.5578769702613 per second.)
[03:01:12.672390][21][I] Stopping after 30000 ms. Initiated: 52563 / Completed: 52552. (Completion rate was 1751.7214216276664 per second.)
[03:01:12.672806][18][I] Stopping after 30000 ms. Initiated: 52292 / Completed: 52285. (Completion rate was 1742.794585200389 per second.)
[03:01:12.673438][19][I] Stopping after 30001 ms. Initiated: 13309 / Completed: 13309. (Completion rate was 443.6144797179453 per second.)
Nighthawk - A layer 7 protocol benchmarking tool.

benchmark_http_client.latency_2xx (172340 samples)
  min: 0s 000ms 356us | mean: 0s 006ms 210us | max: 0s 101ms 638us | pstdev: 0s 012ms 211us

  Percentile  Count       Value          
  0.5         86177       0s 002ms 665us 
  0.75        129260      0s 004ms 228us 
  0.8         137873      0s 004ms 775us 
  0.9         155106      0s 007ms 354us 
  0.95        163724      0s 048ms 760us 
  0.990625    170726      0s 055ms 965us 
  0.99902344  172172      0s 064ms 825us 

[03:01:18.296362][18][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
Queueing and connection setup latency (172365 samples)
  min: 0s 000ms 001us | mean: 0s 000ms 012us | max: 0s 021ms 094us | pstdev: 0s 000ms 143us

  Percentile  Count       Value          
  0.5         86185       0s 000ms 003us 
  0.75        129311      0s 000ms 008us 
  0.8         138040      0s 000ms 009us 
  0.9         155190      0s 000ms 009us 
  0.95        163751      0s 000ms 011us 
  0.990625    170750      0s 000ms 118us 
  0.99902344  172197      0s 001ms 587us 

Request start to response end (172340 samples)
  min: 0s 000ms 356us | mean: 0s 006ms 210us | max: 0s 101ms 638us | pstdev: 0s 012ms 211us

  Percentile  Count       Value          
  0.5         86173       0s 002ms 664us 
  0.75        129262      0s 004ms 228us 
  0.8         137876      0s 004ms 775us 
  0.9         155107      0s 007ms 354us 
  0.95        163724      0s 048ms 760us 
  0.990625    170726      0s 055ms 963us 
  0.99902344  172172      0s 064ms 825us 

Response body size in bytes (172340 samples)
  min: 10 | mean: 10 | max: 10 | pstdev: 0

Response header size in bytes (172340 samples)
  min: 110 | mean: 110 | max: 110 | pstdev: 0

Blocking. Results are skewed when significant numbers are reported here. (56978 samples)
  min: 0s 000ms 034us | mean: 0s 002ms 104us | max: 0s 082ms 956us | pstdev: 0s 006ms 958us

  Percentile  Count       Value          
  0.5         28489       0s 000ms 737us 
  0.75        42735       0s 001ms 556us 
  0.8         45583       0s 001ms 826us 
  0.9         51281       0s 002ms 782us 
  0.95        54130       0s 004ms 249us 
  0.990625    56444       0s 048ms 467us 
  0.99902344  56923       0s 057ms 014us 

Initiation to completion (172703 samples)
  min: 0s 000ms 004us | mean: 0s 006ms 277us | max: 0s 101ms 695us | pstdev: 0s 012ms 210us

  Percentile  Count       Value          
  0.5         86352       0s 002ms 719us 
  0.75        129528      0s 004ms 311us 
  0.8         138165      0s 004ms 865us 
  0.9         155433      0s 007ms 529us 
  0.95        164068      0s 048ms 779us 
  0.990625    171087      0s 056ms 068us 
  0.99902344  172535      0s 064ms 860us 

Counter                                 Value       Per second
benchmark.http_2xx                      172340      5744.56
benchmark.pool_overflow                 363         12.10
cluster_manager.cluster_added           4           0.13
default.total_match_count               4           0.13
membership_change                       4           0.13
runtime.load_success                    1           0.03
runtime.override_dir_not_exists         1           0.03
upstream_cx_http1_total                 37          1.23
upstream_cx_rx_bytes_total              27057380    901895.44
upstream_cx_total                       37          1.23
upstream_cx_tx_bytes_total              7411695     247051.78
upstream_rq_pending_overflow            363         12.10
upstream_rq_pending_total               37          1.23
upstream_rq_total                       172365      5745.39

[03:01:23.297822][21][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[03:01:28.298754][23][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[03:01:28.300386][1][I] Done.

```

</details>

<details>
<summary>scale-down-httproutes-300</summary>

```plaintext
Warning: 29 03:01:55.901][1][warning][misc] [external/envoy/source/common/protobuf/message_validator_impl.cc:21] Deprecated field: type envoy.config.core.v3.HeaderValueOption Using deprecated option 'envoy.config.core.v3.HeaderValueOption.append' from file base.proto. This configuration will be removed from Envoy soon. Please see https://www.envoyproxy.io/docs/envoy/latest/version_history/version_history for details. If continued use of this field is absolutely necessary, see https://www.envoyproxy.io/docs/envoy/latest/configuration/operations/runtime#using-runtime-overrides-for-deprecated-features for how to apply a temporary and highly discouraged override.
[03:01:55.901704][1][I] Detected 4 (v)CPUs with affinity..
[03:01:55.901717][1][I] Starting 4 threads / event loops. Time limit: 30 seconds.
[03:01:55.901720][1][I] Global targets: 400 connections and 40000 calls per second.
[03:01:55.901723][1][I]    (Per-worker targets: 100 connections and 10000 calls per second)
[03:02:26.605529][23][I] Stopping after 29995 ms. Initiated: 49635 / Completed: 49627. (Completion rate was 1654.46716469261 per second.)
[03:02:26.609785][18][I] Stopping after 30004 ms. Initiated: 46794 / Completed: 46787. (Completion rate was 1559.3076656634 per second.)
[03:02:26.612242][21][I] Stopping after 30008 ms. Initiated: 28252 / Completed: 28252. (Completion rate was 941.461721626644 per second.)
[03:02:26.613171][19][I] Stopping after 30009 ms. Initiated: 37451 / Completed: 37446. (Completion rate was 1247.801452216162 per second.)
[03:02:32.320914][18][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
Nighthawk - A layer 7 protocol benchmarking tool.

benchmark_http_client.latency_2xx (161736 samples)
  min: 0s 000ms 368us | mean: 0s 004ms 241us | max: 0s 073ms 580us | pstdev: 0s 008ms 730us

  Percentile  Count       Value          
  0.5         80872       0s 002ms 059us 
  0.75        121304      0s 003ms 210us 
  0.8         129390      0s 003ms 631us 
  0.9         145564      0s 005ms 406us 
  0.95        153650      0s 011ms 199us 
  0.990625    160220      0s 051ms 089us 
  0.99902344  161580      0s 057ms 368us 

Queueing and connection setup latency (161756 samples)
  min: 0s 000ms 001us | mean: 0s 000ms 014us | max: 0s 032ms 240us | pstdev: 0s 000ms 155us

  Percentile  Count       Value          
  0.5         80984       0s 000ms 003us 
  0.75        121381      0s 000ms 009us 
  0.8         129453      0s 000ms 009us 
  0.9         145649      0s 000ms 010us 
  0.95        153680      0s 000ms 011us 
  0.990625    160240      0s 000ms 169us 
  0.99902344  161599      0s 001ms 933us 

Request start to response end (161736 samples)
  min: 0s 000ms 368us | mean: 0s 004ms 241us | max: 0s 073ms 580us | pstdev: 0s 008ms 730us

  Percentile  Count       Value          
  0.5         80868       0s 002ms 059us 
  0.75        121302      0s 003ms 209us 
  0.8         129390      0s 003ms 631us 
  0.9         145563      0s 005ms 403us 
  0.95        153650      0s 011ms 196us 
  0.990625    160221      0s 051ms 089us 
  0.99902344  161580      0s 057ms 368us 

Response body size in bytes (161736 samples)
  min: 10 | mean: 10 | max: 10 | pstdev: 0

Response header size in bytes (161736 samples)
  min: 110 | mean: 110 | max: 110 | pstdev: 0

Blocking. Results are skewed when significant numbers are reported here. (64868 samples)
  min: 0s 000ms 040us | mean: 0s 001ms 849us | max: 0s 072ms 249us | pstdev: 0s 005ms 548us

  Percentile  Count       Value          
  0.5         32435       0s 000ms 817us 
  0.75        48651       0s 001ms 542us 
  0.8         51896       0s 001ms 782us 
  0.9         58382       0s 002ms 681us 
  0.95        61625       0s 004ms 049us 
  0.990625    64261       0s 044ms 695us 
  0.99902344  64805       0s 051ms 093us 

Initiation to completion (162112 samples)
  min: 0s 000ms 005us | mean: 0s 004ms 291us | max: 0s 073ms 601us | pstdev: 0s 008ms 726us

  Percentile  Count       Value          
  0.5         81057       0s 002ms 105us 
  0.75        121585      0s 003ms 273us 
  0.8         129691      0s 003ms 700us 
  0.9         145903      0s 005ms 518us 
  0.95        154007      0s 011ms 286us 
  0.990625    160593      0s 051ms 124us 
  0.99902344  161954      0s 057ms 366us 

Counter                                 Value       Per second
benchmark.http_2xx                      161736      5390.35
benchmark.pool_overflow                 376         12.53
cluster_manager.cluster_added           4           0.13
default.total_match_count               4           0.13
membership_change                       4           0.13
runtime.load_success                    1           0.03
runtime.override_dir_not_exists         1           0.03
upstream_cx_http1_total                 24          0.80
upstream_cx_rx_bytes_total              25392552    846284.54
upstream_cx_total                       24          0.80
upstream_cx_tx_bytes_total              6955508     231813.60
upstream_rq_pending_overflow            376         12.53
upstream_rq_pending_total               24          0.80
upstream_rq_total                       161756      5391.01

[03:02:37.324872][19][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[03:02:42.326139][23][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[03:02:42.327893][1][I] Done.

```

</details>

<details>
<summary>scale-down-httproutes-100</summary>

```plaintext
Warning: 29 03:02:59.726][1][warning][misc] [external/envoy/source/common/protobuf/message_validator_impl.cc:21] Deprecated field: type envoy.config.core.v3.HeaderValueOption Using deprecated option 'envoy.config.core.v3.HeaderValueOption.append' from file base.proto. This configuration will be removed from Envoy soon. Please see https://www.envoyproxy.io/docs/envoy/latest/version_history/version_history for details. If continued use of this field is absolutely necessary, see https://www.envoyproxy.io/docs/envoy/latest/configuration/operations/runtime#using-runtime-overrides-for-deprecated-features for how to apply a temporary and highly discouraged override.
[03:02:59.726741][1][I] Detected 4 (v)CPUs with affinity..
[03:02:59.726755][1][I] Starting 4 threads / event loops. Time limit: 30 seconds.
[03:02:59.726757][1][I] Global targets: 400 connections and 40000 calls per second.
[03:02:59.726759][1][I]    (Per-worker targets: 100 connections and 10000 calls per second)
[03:03:30.428772][18][I] Stopping after 30000 ms. Initiated: 53855 / Completed: 53837. (Completion rate was 1794.5648721017947 per second.)
[03:03:30.429263][22][I] Stopping after 30000 ms. Initiated: 29991 / Completed: 29986. (Completion rate was 999.5176075896406 per second.)
[03:03:30.429653][19][I] Stopping after 30000 ms. Initiated: 25627 / Completed: 25625. (Completion rate was 854.1414694933167 per second.)
[03:03:30.430085][23][I] Stopping after 30001 ms. Initiated: 34111 / Completed: 34111. (Completion rate was 1136.9852009598262 per second.)
Nighthawk - A layer 7 protocol benchmarking tool.

benchmark_http_client.latency_2xx (143196 samples)
  min: 0s 000ms 345us | mean: 0s 007ms 260us | max: 0s 107ms 692us | pstdev: 0s 010ms 819us

  Percentile  Count       Value          
  0.5         71598       0s 003ms 409us 
  0.75        107399      0s 006ms 646us 
  0.8         114557      0s 008ms 150us 
  0.9         128877      0s 019ms 938us 
  0.95        136039      0s 031ms 923us 
[03:03:36.209040][18][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
  0.990625    141854      0s 054ms 925us 
  0.99902344  143058      0s 082ms 124us 

Queueing and connection setup latency (143221 samples)
  min: 0s 000ms 001us | mean: 0s 000ms 017us | max: 0s 042ms 967us | pstdev: 0s 000ms 345us

  Percentile  Count       Value          
  0.5         71628       0s 000ms 003us 
  0.75        107479      0s 000ms 008us 
  0.8         114584      0s 000ms 009us 
  0.9         128905      0s 000ms 010us 
  0.95        136066      0s 000ms 011us 
  0.990625    141879      0s 000ms 126us 
  0.99902344  143082      0s 002ms 497us 

Request start to response end (143196 samples)
  min: 0s 000ms 345us | mean: 0s 007ms 260us | max: 0s 107ms 692us | pstdev: 0s 010ms 819us

  Percentile  Count       Value          
  0.5         71598       0s 003ms 408us 
  0.75        107400      0s 006ms 646us 
  0.8         114557      0s 008ms 148us 
  0.9         128878      0s 019ms 938us 
  0.95        136037      0s 031ms 922us 
  0.990625    141854      0s 054ms 925us 
  0.99902344  143058      0s 082ms 124us 

Response body size in bytes (143196 samples)
  min: 10 | mean: 10 | max: 10 | pstdev: 0

Response header size in bytes (143196 samples)
  min: 110 | mean: 110 | max: 110 | pstdev: 0

Blocking. Results are skewed when significant numbers are reported here. (43316 samples)
  min: 0s 000ms 034us | mean: 0s 002ms 767us | max: 0s 095ms 743us | pstdev: 0s 006ms 458us

  Percentile  Count       Value          
  0.5         21658       0s 000ms 985us 
  0.75        32487       0s 002ms 212us 
  0.8         34654       0s 002ms 720us 
  0.9         38985       0s 005ms 062us 
  0.95        41151       0s 010ms 924us 
  0.990625    42910       0s 036ms 741us 
  0.99902344  43274       0s 057ms 384us 

Initiation to completion (143559 samples)
  min: 0s 000ms 003us | mean: 0s 007ms 357us | max: 0s 107ms 704us | pstdev: 0s 010ms 878us

  Percentile  Count       Value          
  0.5         71781       0s 003ms 486us 
  0.75        107671      0s 006ms 787us 
  0.8         114851      0s 008ms 298us 
  0.9         129204      0s 020ms 202us 
  0.95        136383      0s 032ms 024us 
  0.990625    142214      0s 055ms 140us 
  0.99902344  143419      0s 081ms 698us 

Counter                                 Value       Per second
benchmark.http_2xx                      143196      4773.09
benchmark.pool_overflow                 363         12.10
cluster_manager.cluster_added           4           0.13
default.total_match_count               4           0.13
membership_change                       4           0.13
runtime.load_success                    1           0.03
runtime.override_dir_not_exists         1           0.03
upstream_cx_http1_total                 37          1.23
upstream_cx_rx_bytes_total              22481772    749375.79
upstream_cx_total                       37          1.23
upstream_cx_tx_bytes_total              6158503     205278.88
upstream_rq_pending_overflow            363         12.10
upstream_rq_pending_total               37          1.23
upstream_rq_total                       143221      4773.93

[03:03:41.210523][19][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[03:03:46.211509][22][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[03:03:46.214098][1][I] Done.

```

</details>

<details>
<summary>scale-down-httproutes-50</summary>

```plaintext
Warning: 29 03:03:53.663][1][warning][misc] [external/envoy/source/common/protobuf/message_validator_impl.cc:21] Deprecated field: type envoy.config.core.v3.HeaderValueOption Using deprecated option 'envoy.config.core.v3.HeaderValueOption.append' from file base.proto. This configuration will be removed from Envoy soon. Please see https://www.envoyproxy.io/docs/envoy/latest/version_history/version_history for details. If continued use of this field is absolutely necessary, see https://www.envoyproxy.io/docs/envoy/latest/configuration/operations/runtime#using-runtime-overrides-for-deprecated-features for how to apply a temporary and highly discouraged override.
[03:03:53.664682][1][I] Detected 4 (v)CPUs with affinity..
[03:03:53.664695][1][I] Starting 4 threads / event loops. Time limit: 30 seconds.
[03:03:53.664698][1][I] Global targets: 400 connections and 40000 calls per second.
[03:03:53.664700][1][I]    (Per-worker targets: 100 connections and 10000 calls per second)
[03:04:24.367434][23][I] Stopping after 30000 ms. Initiated: 46577 / Completed: 46567. (Completion rate was 1552.2191046582072 per second.)
[03:04:24.367829][18][I] Stopping after 30000 ms. Initiated: 42300 / Completed: 42290. (Completion rate was 1409.6315198541051 per second.)
[03:04:24.368549][21][I] Stopping after 30001 ms. Initiated: 39600 / Completed: 39591. (Completion rate was 1319.637581142412 per second.)
[03:04:24.367485][19][I] Stopping after 29994 ms. Initiated: 14551 / Completed: 14549. (Completion rate was 485.05192262626696 per second.)
[03:04:30.226109][18][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
Nighthawk - A layer 7 protocol benchmarking tool.

benchmark_http_client.latency_2xx (142634 samples)
  min: 0s 000ms 350us | mean: 0s 007ms 306us | max: 0s 108ms 011us | pstdev: 0s 010ms 777us

  Percentile  Count       Value          
  0.5         71317       0s 003ms 525us 
  0.75        106976      0s 006ms 682us 
  0.8         114108      0s 008ms 188us 
  0.9         128372      0s 019ms 436us 
  0.95        135503      0s 031ms 845us 
  0.990625    141297      0s 055ms 373us 
  0.99902344  142495      0s 080ms 084us 

Queueing and connection setup latency (142665 samples)
  min: 0s 000ms 001us | mean: 0s 000ms 016us | max: 0s 030ms 452us | pstdev: 0s 000ms 301us

  Percentile  Count       Value          
  0.5         71653       0s 000ms 003us 
  0.75        107071      0s 000ms 008us 
  0.8         114147      0s 000ms 009us 
  0.9         128437      0s 000ms 010us 
  0.95        135540      0s 000ms 011us 
  0.990625    141328      0s 000ms 076us 
  0.99902344  142526      0s 002ms 365us 

Request start to response end (142634 samples)
  min: 0s 000ms 350us | mean: 0s 007ms 305us | max: 0s 108ms 011us | pstdev: 0s 010ms 777us

  Percentile  Count       Value          
  0.5         71318       0s 003ms 525us 
  0.75        106977      0s 006ms 681us 
  0.8         114108      0s 008ms 187us 
  0.9         128371      0s 019ms 434us 
  0.95        135503      0s 031ms 842us 
  0.990625    141297      0s 055ms 373us 
  0.99902344  142495      0s 080ms 084us 

Response body size in bytes (142634 samples)
  min: 10 | mean: 10 | max: 10 | pstdev: 0

Response header size in bytes (142634 samples)
  min: 110 | mean: 110 | max: 110 | pstdev: 0

Blocking. Results are skewed when significant numbers are reported here. (42514 samples)
  min: 0s 000ms 036us | mean: 0s 002ms 819us | max: 0s 097ms 796us | pstdev: 0s 006ms 582us

  Percentile  Count       Value          
  0.5         21257       0s 000ms 979us 
  0.75        31886       0s 002ms 264us 
  0.8         34012       0s 002ms 802us 
  0.9         38263       0s 005ms 232us 
  0.95        40389       0s 011ms 347us 
  0.990625    42116       0s 036ms 845us 
  0.99902344  42473       0s 061ms 714us 

Initiation to completion (142997 samples)
  min: 0s 000ms 003us | mean: 0s 007ms 382us | max: 0s 108ms 048us | pstdev: 0s 010ms 803us

  Percentile  Count       Value          
  0.5         71499       0s 003ms 591us 
  0.75        107249      0s 006ms 808us 
  0.8         114399      0s 008ms 328us 
  0.9         128698      0s 019ms 615us 
  0.95        135848      0s 031ms 933us 
  0.990625    141657      0s 055ms 416us 
  0.99902344  142858      0s 080ms 097us 

Counter                                 Value       Per second
benchmark.http_2xx                      142634      4754.58
benchmark.pool_overflow                 363         12.10
cluster_manager.cluster_added           4           0.13
default.total_match_count               4           0.13
membership_change                       4           0.13
runtime.load_success                    1           0.03
runtime.override_dir_not_exists         1           0.03
upstream_cx_http1_total                 37          1.23
upstream_cx_rx_bytes_total              22393538    746468.86
upstream_cx_total                       37          1.23
upstream_cx_tx_bytes_total              6134595     204491.32
upstream_rq_pending_overflow            363         12.10
upstream_rq_pending_total               37          1.23
upstream_rq_total                       142665      4755.61

[03:04:35.227368][19][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[03:04:40.228272][21][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[03:04:45.229345][23][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[03:04:45.231026][1][I] Done.

```

</details>

<details>
<summary>scale-down-httproutes-10</summary>

```plaintext
Warning: 29 03:04:54.906][1][warning][misc] [external/envoy/source/common/protobuf/message_validator_impl.cc:21] Deprecated field: type envoy.config.core.v3.HeaderValueOption Using deprecated option 'envoy.config.core.v3.HeaderValueOption.append' from file base.proto. This configuration will be removed from Envoy soon. Please see https://www.envoyproxy.io/docs/envoy/latest/version_history/version_history for details. If continued use of this field is absolutely necessary, see https://www.envoyproxy.io/docs/envoy/latest/configuration/operations/runtime#using-runtime-overrides-for-deprecated-features for how to apply a temporary and highly discouraged override.
[03:04:54.907652][1][I] Detected 4 (v)CPUs with affinity..
[03:04:54.907700][1][I] Starting 4 threads / event loops. Time limit: 30 seconds.
[03:04:54.907720][1][I] Global targets: 400 connections and 40000 calls per second.
[03:04:54.907741][1][I]    (Per-worker targets: 100 connections and 10000 calls per second)
[03:05:25.610341][17][I] Stopping after 29998 ms. Initiated: 16275 / Completed: 16274. (Completion rate was 542.4883119503135 per second.)
[03:05:25.610352][18][I] Stopping after 29994 ms. Initiated: 23300 / Completed: 23298. (Completion rate was 776.7367056601962 per second.)
[03:05:25.610457][19][I] Stopping after 29977 ms. Initiated: 48060 / Completed: 48053. (Completion rate was 1602.9754704689472 per second.)
[03:05:25.610482][21][I] Stopping after 29977 ms. Initiated: 38690 / Completed: 38686. (Completion rate was 1290.507795769919 per second.)
[03:05:31.462778][17][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
Nighthawk - A layer 7 protocol benchmarking tool.

benchmark_http_client.latency_2xx (125925 samples)
  min: 0s 000ms 332us | mean: 0s 003ms 142us | max: 0s 171ms 196us | pstdev: 0s 006ms 668us

  Percentile  Count       Value          
  0.5         62964       0s 001ms 457us 
  0.75        94445       0s 002ms 468us 
  0.8         100740      0s 002ms 889us 
  0.9         113335      0s 004ms 824us 
  0.95        119630      0s 010ms 471us 
  0.990625    124745      0s 039ms 651us 
  0.99902344  125803      0s 065ms 499us 

Queueing and connection setup latency (125939 samples)
  min: 0s 000ms 001us | mean: 0s 000ms 014us | max: 0s 059ms 512us | pstdev: 0s 000ms 250us

  Percentile  Count       Value          
  0.5         62993       0s 000ms 007us 
  0.75        94584       0s 000ms 009us 
  0.8         100762      0s 000ms 009us 
  0.9         113362      0s 000ms 010us 
  0.95        119645      0s 000ms 011us 
  0.990625    124759      0s 000ms 156us 
  0.99902344  125817      0s 001ms 272us 

Request start to response end (125925 samples)
  min: 0s 000ms 331us | mean: 0s 003ms 141us | max: 0s 171ms 196us | pstdev: 0s 006ms 668us

  Percentile  Count       Value          
  0.5         62963       0s 001ms 456us 
  0.75        94446       0s 002ms 467us 
  0.8         100742      0s 002ms 888us 
  0.9         113333      0s 004ms 823us 
  0.95        119629      0s 010ms 469us 
  0.990625    124745      0s 039ms 651us 
  0.99902344  125803      0s 065ms 499us 

Response body size in bytes (125925 samples)
  min: 10 | mean: 10 | max: 10 | pstdev: 0

Response header size in bytes (125925 samples)
  min: 110 | mean: 110 | max: 110 | pstdev: 0

Blocking. Results are skewed when significant numbers are reported here. (66167 samples)
  min: 0s 000ms 037us | mean: 0s 001ms 808us | max: 0s 099ms 373us | pstdev: 0s 004ms 798us

  Percentile  Count       Value          
  0.5         33084       0s 000ms 815us 
  0.75        49626       0s 001ms 426us 
  0.8         52934       0s 001ms 644us 
  0.9         59551       0s 002ms 601us 
  0.95        62859       0s 004ms 569us 
  0.990625    65547       0s 028ms 648us 
  0.99902344  66103       0s 054ms 749us 

Initiation to completion (126311 samples)
  min: 0s 000ms 003us | mean: 0s 003ms 182us | max: 0s 171ms 368us | pstdev: 0s 006ms 690us

  Percentile  Count       Value          
  0.5         63156       0s 001ms 486us 
  0.75        94735       0s 002ms 520us 
  0.8         101049      0s 002ms 952us 
  0.9         113680      0s 004ms 910us 
  0.95        119996      0s 010ms 608us 
  0.990625    125128      0s 039ms 723us 
  0.99902344  126188      0s 065ms 523us 

Counter                                 Value       Per second
benchmark.http_2xx                      125925      4199.31
benchmark.pool_overflow                 386         12.87
cluster_manager.cluster_added           4           0.13
default.total_match_count               4           0.13
membership_change                       4           0.13
runtime.load_success                    1           0.03
runtime.override_dir_not_exists         1           0.03
upstream_cx_http1_total                 14          0.47
upstream_cx_rx_bytes_total              19770225    659291.83
upstream_cx_total                       14          0.47
upstream_cx_tx_bytes_total              5415377     180590.45
upstream_rq_pending_overflow            386         12.87
upstream_rq_pending_total               14          0.47
upstream_rq_total                       125939      4199.78

[03:05:36.463346][18][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[03:05:41.463985][19][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[03:05:46.464754][21][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[03:05:46.466019][1][I] Done.

```

</details>

### Metrics

|Benchmark Name           |Envoy Gateway Memory (MiB)|Envoy Gateway Total CPU (Seconds)|Envoy Proxy Memory: crhxz<sup>[1]</sup> (MiB)|
|-                        |-                         |-                                |-                                            |
|scale-up-httproutes-10   |80                        |0.33                             |7                                            |
|scale-up-httproutes-50   |104                       |1.93                             |12                                           |
|scale-up-httproutes-100  |167                       |9.03                             |20                                           |
|scale-up-httproutes-300  |773                       |156.58                           |80                                           |
|scale-up-httproutes-500  |1921                      |11.06                            |188                                          |
|scale-down-httproutes-300|724                       |7.94                             |82                                           |
|scale-down-httproutes-100|801                       |60.46                            |67                                           |
|scale-down-httproutes-50 |474                       |112.37                           |48                                           |
|scale-down-httproutes-10 |125                       |152.13                           |12                                           |
1. envoy-gateway-system/envoy-benchmark-test-benchmark-0520098c-7fb5487b9f-crhxz