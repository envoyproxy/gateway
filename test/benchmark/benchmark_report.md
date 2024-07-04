# Benchmark Report

Benchmark test settings:

|RPS  |Connections|Duration (Seconds)|CPU Limits (m)|Memory Limits (MiB)|
|-    |-          |-                 |-             |-                  |
|10000|100        |90                |1000          |2048               |

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
Warning: 30 09:50:00.929][1][warning][misc] [external/envoy/source/common/protobuf/message_validator_impl.cc:21] Deprecated field: type envoy.config.core.v3.HeaderValueOption Using deprecated option 'envoy.config.core.v3.HeaderValueOption.append' from file base.proto. This configuration will be removed from Envoy soon. Please see https://www.envoyproxy.io/docs/envoy/latest/version_history/version_history for details. If continued use of this field is absolutely necessary, see https://www.envoyproxy.io/docs/envoy/latest/configuration/operations/runtime#using-runtime-overrides-for-deprecated-features for how to apply a temporary and highly discouraged override.
[09:50:00.930463][1][I] Detected 4 (v)CPUs with affinity..
[09:50:00.930476][1][I] Starting 4 threads / event loops. Time limit: 90 seconds.
[09:50:00.930478][1][I] Global targets: 400 connections and 40000 calls per second.
[09:50:00.930480][1][I]    (Per-worker targets: 100 connections and 10000 calls per second)
[09:51:31.632205][18][I] Stopping after 90000 ms. Initiated: 132984 / Completed: 132975. (Completion rate was 1477.4999671666674 per second.)
[09:51:31.632253][19][I] Stopping after 90000 ms. Initiated: 184337 / Completed: 184323. (Completion rate was 2048.0327871912564 per second.)
[09:51:31.632393][23][I] Stopping after 90000 ms. Initiated: 119466 / Completed: 119458. (Completion rate was 1327.309400356773 per second.)
[09:51:31.632493][21][I] Stopping after 90000 ms. Initiated: 98553 / Completed: 98547. (Completion rate was 1094.963746763342 per second.)
[09:51:37.272832][18][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
Nighthawk - A layer 7 protocol benchmarking tool.

benchmark_http_client.latency_2xx (534940 samples)
  min: 0s 000ms 324us | mean: 0s 006ms 013us | max: 0s 080ms 613us | pstdev: 0s 011ms 899us

  Percentile  Count       Value          
  0.5         267470      0s 002ms 663us 
  0.75        401207      0s 004ms 128us 
  0.8         427956      0s 004ms 645us 
  0.9         481449      0s 006ms 891us 
  0.95        508194      0s 048ms 232us 
  0.990625    529927      0s 055ms 224us 
  0.99902344  534418      0s 061ms 255us 

Queueing and connection setup latency (534977 samples)
  min: 0s 000ms 001us | mean: 0s 000ms 012us | max: 0s 026ms 379us | pstdev: 0s 000ms 131us

  Percentile  Count       Value          
  0.5         267828      0s 000ms 003us 
  0.75        401520      0s 000ms 008us 
  0.8         427993      0s 000ms 009us 
  0.9         482090      0s 000ms 009us 
  0.95        508250      0s 000ms 010us 
  0.990625    529962      0s 000ms 132us 
  0.99902344  534455      0s 001ms 630us 

Request start to response end (534940 samples)
  min: 0s 000ms 324us | mean: 0s 006ms 012us | max: 0s 080ms 613us | pstdev: 0s 011ms 899us

  Percentile  Count       Value          
  0.5         267484      0s 002ms 663us 
  0.75        401210      0s 004ms 127us 
  0.8         427955      0s 004ms 645us 
  0.9         481446      0s 006ms 889us 
  0.95        508195      0s 048ms 232us 
  0.990625    529929      0s 055ms 224us 
  0.99902344  534418      0s 061ms 251us 

Response body size in bytes (534940 samples)
  min: 10 | mean: 10 | max: 10 | pstdev: 0

Response header size in bytes (534940 samples)
  min: 110 | mean: 110 | max: 110 | pstdev: 0

Blocking. Results are skewed when significant numbers are reported here. (176008 samples)
  min: 0s 000ms 036us | mean: 0s 002ms 044us | max: 0s 073ms 248us | pstdev: 0s 006ms 744us

  Percentile  Count       Value          
  0.5         88005       0s 000ms 734us 
  0.75        132008      0s 001ms 534us 
  0.8         140807      0s 001ms 802us 
  0.9         158410      0s 002ms 792us 
  0.95        167208      0s 004ms 190us 
  0.990625    174359      0s 047ms 886us 
  0.99902344  175837      0s 053ms 557us 

Initiation to completion (535303 samples)
  min: 0s 000ms 004us | mean: 0s 006ms 072us | max: 0s 080ms 633us | pstdev: 0s 011ms 900us

  Percentile  Count       Value          
  0.5         267665      0s 002ms 717us 
  0.75        401487      0s 004ms 204us 
  0.8         428258      0s 004ms 732us 
  0.9         481775      0s 007ms 022us 
  0.95        508538      0s 048ms 248us 
  0.990625    530287      0s 055ms 298us 
  0.99902344  534781      0s 061ms 421us 

Counter                                 Value       Per second
benchmark.http_2xx                      534940      5943.77
benchmark.pool_overflow                 363         4.03
cluster_manager.cluster_added           4           0.04
default.total_match_count               4           0.04
membership_change                       4           0.04
runtime.load_success                    1           0.01
runtime.override_dir_not_exists         1           0.01
upstream_cx_http1_total                 37          0.41
upstream_cx_rx_bytes_total              83985580    933172.12
upstream_cx_total                       37          0.41
upstream_cx_tx_bytes_total              23004011    255599.85
upstream_rq_pending_overflow            363         4.03
upstream_rq_pending_total               37          0.41
upstream_rq_total                       534977      5944.18

[09:51:42.273918][19][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[09:51:47.274994][21][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[09:51:52.275699][23][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[09:51:52.277010][1][I] Done.

```

</details>

<details>
<summary>scale-up-httproutes-100</summary>

```plaintext
Warning: 30 09:52:08.843][1][warning][misc] [external/envoy/source/common/protobuf/message_validator_impl.cc:21] Deprecated field: type envoy.config.core.v3.HeaderValueOption Using deprecated option 'envoy.config.core.v3.HeaderValueOption.append' from file base.proto. This configuration will be removed from Envoy soon. Please see https://www.envoyproxy.io/docs/envoy/latest/version_history/version_history for details. If continued use of this field is absolutely necessary, see https://www.envoyproxy.io/docs/envoy/latest/configuration/operations/runtime#using-runtime-overrides-for-deprecated-features for how to apply a temporary and highly discouraged override.
[09:52:08.843704][1][I] Detected 4 (v)CPUs with affinity..
[09:52:08.843716][1][I] Starting 4 threads / event loops. Time limit: 90 seconds.
[09:52:08.843719][1][I] Global targets: 400 connections and 40000 calls per second.
[09:52:08.843720][1][I]    (Per-worker targets: 100 connections and 10000 calls per second)
[09:53:39.545479][18][I] Stopping after 90000 ms. Initiated: 60881 / Completed: 60878. (Completion rate was 676.421921590257 per second.)
[09:53:39.545649][23][I] Stopping after 90000 ms. Initiated: 77409 / Completed: 77407. (Completion rate was 860.076458993874 per second.)
[09:53:39.545773][19][I] Stopping after 90000 ms. Initiated: 271217 / Completed: 271198. (Completion rate was 3013.300631521137 per second.)
[09:53:39.546416][21][I] Stopping after 90000 ms. Initiated: 123329 / Completed: 123321. (Completion rate was 1370.219189626365 per second.)
[09:53:45.183623][18][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
Nighthawk - A layer 7 protocol benchmarking tool.

benchmark_http_client.latency_2xx (532441 samples)
  min: 0s 000ms 359us | mean: 0s 006ms 029us | max: 0s 090ms 263us | pstdev: 0s 011ms 870us

  Percentile  Count       Value          
  0.5         266227      0s 002ms 698us 
  0.75        399344      0s 004ms 228us 
  0.8         425959      0s 004ms 757us 
  0.9         479199      0s 007ms 042us 
  0.95        505821      0s 047ms 982us 
  0.990625    527453      0s 055ms 048us 
  0.99902344  531922      0s 060ms 678us 

Queueing and connection setup latency (532473 samples)
  min: 0s 000ms 001us | mean: 0s 000ms 011us | max: 0s 023ms 204us | pstdev: 0s 000ms 126us

  Percentile  Count       Value          
  0.5         267630      0s 000ms 003us 
  0.75        399579      0s 000ms 008us 
  0.8         426987      0s 000ms 009us 
  0.9         479713      0s 000ms 009us 
  0.95        505878      0s 000ms 010us 
  0.990625    527482      0s 000ms 040us 
  0.99902344  531955      0s 001ms 391us 

Request start to response end (532441 samples)
  min: 0s 000ms 359us | mean: 0s 006ms 028us | max: 0s 090ms 263us | pstdev: 0s 011ms 870us

  Percentile  Count       Value          
  0.5         266221      0s 002ms 697us 
  0.75        399332      0s 004ms 227us 
  0.8         425959      0s 004ms 757us 
  0.9         479199      0s 007ms 042us 
  0.95        505824      0s 047ms 982us 
  0.990625    527450      0s 055ms 046us 
  0.99902344  531922      0s 060ms 678us 

Response body size in bytes (532441 samples)
  min: 10 | mean: 10 | max: 10 | pstdev: 0

Response header size in bytes (532441 samples)
  min: 110 | mean: 110 | max: 110 | pstdev: 0

Blocking. Results are skewed when significant numbers are reported here. (183352 samples)
  min: 0s 000ms 033us | mean: 0s 001ms 962us | max: 0s 084ms 799us | pstdev: 0s 006ms 606us

  Percentile  Count       Value          
  0.5         91683       0s 000ms 726us 
  0.75        137514      0s 001ms 420us 
  0.8         146685      0s 001ms 667us 
  0.9         165017      0s 002ms 581us 
  0.95        174186      0s 003ms 882us 
  0.990625    181634      0s 047ms 661us 
  0.99902344  183173      0s 053ms 168us 

Initiation to completion (532804 samples)
  min: 0s 000ms 001us | mean: 0s 006ms 088us | max: 0s 090ms 275us | pstdev: 0s 011ms 873us

  Percentile  Count       Value          
  0.5         266426      0s 002ms 753us 
  0.75        399604      0s 004ms 306us 
  0.8         426251      0s 004ms 841us 
  0.9         479526      0s 007ms 164us 
  0.95        506172      0s 048ms 007us 
  0.990625    527811      0s 055ms 113us 
  0.99902344  532284      0s 060ms 801us 

Counter                                 Value       Per second
benchmark.http_2xx                      532441      5915.99
benchmark.pool_overflow                 363         4.03
cluster_manager.cluster_added           4           0.04
default.total_match_count               4           0.04
membership_change                       4           0.04
runtime.load_success                    1           0.01
runtime.override_dir_not_exists         1           0.01
upstream_cx_http1_total                 37          0.41
upstream_cx_rx_bytes_total              83593237    928810.08
upstream_cx_total                       37          0.41
upstream_cx_tx_bytes_total              22896339    254402.76
upstream_rq_pending_overflow            363         4.03
upstream_rq_pending_total               37          0.41
upstream_rq_total                       532473      5916.34

[09:53:50.184479][19][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[09:53:55.185619][21][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[09:54:00.186411][23][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[09:54:00.187569][1][I] Done.

```

</details>

<details>
<summary>scale-up-httproutes-300</summary>

```plaintext
Warning: 30 09:57:17.615][1][warning][misc] [external/envoy/source/common/protobuf/message_validator_impl.cc:21] Deprecated field: type envoy.config.core.v3.HeaderValueOption Using deprecated option 'envoy.config.core.v3.HeaderValueOption.append' from file base.proto. This configuration will be removed from Envoy soon. Please see https://www.envoyproxy.io/docs/envoy/latest/version_history/version_history for details. If continued use of this field is absolutely necessary, see https://www.envoyproxy.io/docs/envoy/latest/configuration/operations/runtime#using-runtime-overrides-for-deprecated-features for how to apply a temporary and highly discouraged override.
[09:57:17.615763][1][I] Detected 4 (v)CPUs with affinity..
[09:57:17.615775][1][I] Starting 4 threads / event loops. Time limit: 90 seconds.
[09:57:17.615778][1][I] Global targets: 400 connections and 40000 calls per second.
[09:57:17.615779][1][I]    (Per-worker targets: 100 connections and 10000 calls per second)
[09:58:48.317519][18][I] Stopping after 90000 ms. Initiated: 141701 / Completed: 141691. (Completion rate was 1574.3439371558425 per second.)
[09:58:48.317519][19][I] Stopping after 90000 ms. Initiated: 142076 / Completed: 142066. (Completion rate was 1578.5110584940758 per second.)
[09:58:48.317557][21][I] Stopping after 90000 ms. Initiated: 169359 / Completed: 169346. (Completion rate was 1881.6218249909482 per second.)
[09:58:48.317726][23][I] Stopping after 90000 ms. Initiated: 73891 / Completed: 73887. (Completion rate was 820.9651798075076 per second.)
[09:58:53.963287][18][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
Nighthawk - A layer 7 protocol benchmarking tool.

benchmark_http_client.latency_2xx (526627 samples)
  min: 0s 000ms 368us | mean: 0s 006ms 116us | max: 0s 085ms 676us | pstdev: 0s 012ms 060us

  Percentile  Count       Value          
  0.5         263331      0s 002ms 692us 
  0.75        394974      0s 004ms 203us 
  0.8         421305      0s 004ms 735us 
  0.9         473966      0s 007ms 143us 
  0.95        500299      0s 048ms 519us 
  0.990625    521693      0s 055ms 746us 
  0.99902344  526114      0s 062ms 912us 

Queueing and connection setup latency (526664 samples)
  min: 0s 000ms 001us | mean: 0s 000ms 012us | max: 0s 019ms 310us | pstdev: 0s 000ms 136us

  Percentile  Count       Value          
  0.5         263616      0s 000ms 003us 
  0.75        395207      0s 000ms 008us 
  0.8         421679      0s 000ms 009us 
  0.9         474013      0s 000ms 009us 
  0.95        500359      0s 000ms 010us 
  0.990625    521727      0s 000ms 127us 
  0.99902344  526150      0s 001ms 624us 

Request start to response end (526627 samples)
  min: 0s 000ms 368us | mean: 0s 006ms 116us | max: 0s 085ms 676us | pstdev: 0s 012ms 060us

  Percentile  Count       Value          
  0.5         263324      0s 002ms 692us 
  0.75        394984      0s 004ms 203us 
  0.8         421305      0s 004ms 734us 
  0.9         473966      0s 007ms 142us 
  0.95        500299      0s 048ms 519us 
  0.990625    521693      0s 055ms 746us 
  0.99902344  526113      0s 062ms 910us 

Response body size in bytes (526627 samples)
  min: 10 | mean: 10 | max: 10 | pstdev: 0

Response header size in bytes (526627 samples)
  min: 110 | mean: 110 | max: 110 | pstdev: 0

Blocking. Results are skewed when significant numbers are reported here. (178472 samples)
  min: 0s 000ms 034us | mean: 0s 002ms 016us | max: 0s 085ms 700us | pstdev: 0s 006ms 753us

  Percentile  Count       Value          
  0.5         89237       0s 000ms 732us 
  0.75        133855      0s 001ms 469us 
  0.8         142779      0s 001ms 727us 
  0.9         160630      0s 002ms 687us 
  0.95        169550      0s 004ms 042us 
  0.990625    176799      0s 048ms 060us 
  0.99902344  178298      0s 054ms 435us 

Initiation to completion (526990 samples)
  min: 0s 000ms 004us | mean: 0s 006ms 174us | max: 0s 085ms 700us | pstdev: 0s 012ms 062us

  Percentile  Count       Value          
  0.5         263496      0s 002ms 745us 
  0.75        395254      0s 004ms 278us 
  0.8         421594      0s 004ms 818us 
  0.9         474295      0s 007ms 277us 
  0.95        500644      0s 048ms 537us 
  0.990625    522050      0s 055ms 818us 
  0.99902344  526476      0s 062ms 992us 

Counter                                 Value       Per second
benchmark.http_2xx                      526627      5851.41
benchmark.pool_overflow                 363         4.03
cluster_manager.cluster_added           4           0.04
default.total_match_count               4           0.04
membership_change                       4           0.04
runtime.load_success                    1           0.01
runtime.override_dir_not_exists         1           0.01
upstream_cx_http1_total                 37          0.41
upstream_cx_rx_bytes_total              82680439    918671.00
upstream_cx_total                       37          0.41
upstream_cx_tx_bytes_total              22646552    251628.21
upstream_rq_pending_overflow            363         4.03
upstream_rq_pending_total               37          0.41
upstream_rq_total                       526664      5851.82

[09:58:58.964269][19][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[09:59:03.965258][21][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[09:59:08.966199][23][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[09:59:08.967339][1][I] Done.

```

</details>

<details>
<summary>scale-up-httproutes-500</summary>

```plaintext
Warning: 30 10:05:50.181][1][warning][misc] [external/envoy/source/common/protobuf/message_validator_impl.cc:21] Deprecated field: type envoy.config.core.v3.HeaderValueOption Using deprecated option 'envoy.config.core.v3.HeaderValueOption.append' from file base.proto. This configuration will be removed from Envoy soon. Please see https://www.envoyproxy.io/docs/envoy/latest/version_history/version_history for details. If continued use of this field is absolutely necessary, see https://www.envoyproxy.io/docs/envoy/latest/configuration/operations/runtime#using-runtime-overrides-for-deprecated-features for how to apply a temporary and highly discouraged override.
[10:05:50.182352][1][I] Detected 4 (v)CPUs with affinity..
[10:05:50.182365][1][I] Starting 4 threads / event loops. Time limit: 90 seconds.
[10:05:50.182367][1][I] Global targets: 400 connections and 40000 calls per second.
[10:05:50.182369][1][I]    (Per-worker targets: 100 connections and 10000 calls per second)
[10:07:20.884276][21][I] Stopping after 90000 ms. Initiated: 93542 / Completed: 93539. (Completion rate was 1039.3210789690352 per second.)
[10:07:20.884582][19][I] Stopping after 90000 ms. Initiated: 180072 / Completed: 180060. (Completion rate was 2000.6570635127619 per second.)
[10:07:20.884601][18][I] Stopping after 90000 ms. Initiated: 101770 / Completed: 101764. (Completion rate was 1130.7051560639559 per second.)
[10:07:20.884602][23][I] Stopping after 90000 ms. Initiated: 140594 / Completed: 140587. (Completion rate was 1562.0708005282022 per second.)
[10:07:26.528168][18][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
Nighthawk - A layer 7 protocol benchmarking tool.

benchmark_http_client.latency_2xx (515587 samples)
  min: 0s 000ms 352us | mean: 0s 006ms 243us | max: 0s 103ms 198us | pstdev: 0s 012ms 343us

  Percentile  Count       Value          
  0.5         257799      0s 002ms 698us 
  0.75        386691      0s 004ms 155us 
  0.8         412474      0s 004ms 697us 
  0.9         464029      0s 007ms 193us 
  0.95        489815      0s 048ms 922us 
  0.990625    510755      0s 056ms 528us 
  0.99902344  515084      0s 067ms 747us 

Queueing and connection setup latency (515615 samples)
  min: 0s 000ms 001us | mean: 0s 000ms 012us | max: 0s 020ms 386us | pstdev: 0s 000ms 125us

  Percentile  Count       Value          
  0.5         258414      0s 000ms 003us 
  0.75        386829      0s 000ms 008us 
  0.8         412620      0s 000ms 009us 
  0.9         464074      0s 000ms 009us 
  0.95        489886      0s 000ms 010us 
  0.990625    510782      0s 000ms 123us 
  0.99902344  515112      0s 001ms 602us 

Request start to response end (515587 samples)
  min: 0s 000ms 352us | mean: 0s 006ms 243us | max: 0s 103ms 194us | pstdev: 0s 012ms 343us

  Percentile  Count       Value          
  0.5         257802      0s 002ms 698us 
  0.75        386698      0s 004ms 155us 
  0.8         412470      0s 004ms 697us 
  0.9         464031      0s 007ms 192us 
  0.95        489808      0s 048ms 920us 
  0.990625    510756      0s 056ms 528us 
  0.99902344  515084      0s 067ms 747us 

Response body size in bytes (515587 samples)
  min: 10 | mean: 10 | max: 10 | pstdev: 0

Response header size in bytes (515587 samples)
  min: 110 | mean: 110 | max: 110 | pstdev: 0

Blocking. Results are skewed when significant numbers are reported here. (165055 samples)
  min: 0s 000ms 036us | mean: 0s 002ms 180us | max: 0s 090ms 710us | pstdev: 0s 007ms 128us

  Percentile  Count       Value          
  0.5         82528       0s 000ms 765us 
  0.75        123792      0s 001ms 622us 
  0.8         132045      0s 001ms 906us 
  0.9         148550      0s 002ms 915us 
  0.95        156804      0s 004ms 402us 
  0.990625    163510      0s 048ms 775us 
  0.99902344  164894      0s 058ms 019us 

Initiation to completion (515950 samples)
  min: 0s 000ms 004us | mean: 0s 006ms 300us | max: 0s 103ms 215us | pstdev: 0s 012ms 342us

  Percentile  Count       Value          
  0.5         257987      0s 002ms 750us 
  0.75        386970      0s 004ms 235us 
  0.8         412760      0s 004ms 784us 
  0.9         464357      0s 007ms 323us 
  0.95        490153      0s 048ms 945us 
  0.990625    511115      0s 056ms 580us 
  0.99902344  515447      0s 067ms 813us 

Counter                                 Value       Per second
benchmark.http_2xx                      515587      5728.72
benchmark.pool_overflow                 363         4.03
cluster_manager.cluster_added           4           0.04
default.total_match_count               4           0.04
membership_change                       4           0.04
runtime.load_success                    1           0.01
runtime.override_dir_not_exists         1           0.01
upstream_cx_http1_total                 37          0.41
upstream_cx_rx_bytes_total              80947159    899409.36
upstream_cx_total                       37          0.41
upstream_cx_tx_bytes_total              22171445    246348.42
upstream_rq_pending_overflow            363         4.03
upstream_rq_pending_total               37          0.41
upstream_rq_total                       515615      5729.03

[10:07:31.529027][19][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[10:07:36.530025][21][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[10:07:41.530832][23][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[10:07:41.532165][1][I] Done.

```

</details>

<details>
<summary>scale-down-httproutes-300</summary>

```plaintext
Warning: 30 10:08:14.153][1][warning][misc] [external/envoy/source/common/protobuf/message_validator_impl.cc:21] Deprecated field: type envoy.config.core.v3.HeaderValueOption Using deprecated option 'envoy.config.core.v3.HeaderValueOption.append' from file base.proto. This configuration will be removed from Envoy soon. Please see https://www.envoyproxy.io/docs/envoy/latest/version_history/version_history for details. If continued use of this field is absolutely necessary, see https://www.envoyproxy.io/docs/envoy/latest/configuration/operations/runtime#using-runtime-overrides-for-deprecated-features for how to apply a temporary and highly discouraged override.
[10:08:14.154484][1][I] Detected 4 (v)CPUs with affinity..
[10:08:14.154499][1][I] Starting 4 threads / event loops. Time limit: 90 seconds.
[10:08:14.154502][1][I] Global targets: 400 connections and 40000 calls per second.
[10:08:14.154504][1][I]    (Per-worker targets: 100 connections and 10000 calls per second)
[10:09:44.857305][21][I] Stopping after 90000 ms. Initiated: 120108 / Completed: 120100. (Completion rate was 1334.4432137912584 per second.)
[10:09:44.857423][23][I] Stopping after 90000 ms. Initiated: 54818 / Completed: 54817. (Completion rate was 609.0765799271705 per second.)
[10:09:44.857655][18][I] Stopping after 90000 ms. Initiated: 191822 / Completed: 191822. (Completion rate was 2131.3440462977055 per second.)
[10:09:44.858450][19][I] Stopping after 90001 ms. Initiated: 150329 / Completed: 150327. (Completion rate was 1670.2767089192257 per second.)
[10:09:50.501933][19][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
Nighthawk - A layer 7 protocol benchmarking tool.

benchmark_http_client.latency_2xx (516703 samples)
  min: 0s 000ms 339us | mean: 0s 006ms 221us | max: 0s 096ms 722us | pstdev: 0s 011ms 937us

  Percentile  Count       Value          
  0.5         258358      0s 002ms 750us 
  0.75        387543      0s 004ms 382us 
  0.8         413372      0s 004ms 969us 
  0.9         465035      0s 007ms 728us 
  0.95        490869      0s 047ms 640us 
  0.990625    511859      0s 055ms 703us 
  0.99902344  516199      0s 065ms 038us 

Queueing and connection setup latency (516714 samples)
  min: 0s 000ms 001us | mean: 0s 000ms 012us | max: 0s 019ms 922us | pstdev: 0s 000ms 135us

  Percentile  Count       Value          
  0.5         259348      0s 000ms 003us 
  0.75        387676      0s 000ms 008us 
  0.8         413511      0s 000ms 009us 
  0.9         465475      0s 000ms 009us 
  0.95        490880      0s 000ms 010us 
  0.990625    511870      0s 000ms 042us 
  0.99902344  516210      0s 001ms 596us 

Request start to response end (516703 samples)
  min: 0s 000ms 339us | mean: 0s 006ms 221us | max: 0s 096ms 722us | pstdev: 0s 011ms 937us

  Percentile  Count       Value          
  0.5         258360      0s 002ms 750us 
  0.75        387541      0s 004ms 382us 
  0.8         413365      0s 004ms 968us 
  0.9         465034      0s 007ms 726us 
  0.95        490870      0s 047ms 640us 
  0.990625    511859      0s 055ms 703us 
  0.99902344  516199      0s 065ms 038us 

Response body size in bytes (516703 samples)
  min: 10 | mean: 10 | max: 10 | pstdev: 0

Response header size in bytes (516703 samples)
  min: 110 | mean: 110 | max: 110 | pstdev: 0

Blocking. Results are skewed when significant numbers are reported here. (170680 samples)
  min: 0s 000ms 034us | mean: 0s 002ms 108us | max: 0s 090ms 091us | pstdev: 0s 006ms 769us

  Percentile  Count       Value          
  0.5         85340       0s 000ms 756us 
  0.75        128013      0s 001ms 581us 
  0.8         136546      0s 001ms 864us 
  0.9         153612      0s 002ms 914us 
  0.95        162148      0s 004ms 492us 
  0.990625    169080      0s 047ms 810us 
  0.99902344  170514      0s 055ms 848us 

Initiation to completion (517066 samples)
  min: 0s 000ms 003us | mean: 0s 006ms 281us | max: 0s 096ms 731us | pstdev: 0s 011ms 939us

  Percentile  Count       Value          
  0.5         258539      0s 002ms 805us 
  0.75        387802      0s 004ms 465us 
  0.8         413664      0s 005ms 058us 
  0.9         465360      0s 007ms 867us 
  0.95        491214      0s 047ms 654us 
  0.990625    512220      0s 055ms 760us 
  0.99902344  516562      0s 065ms 097us 

Counter                                 Value       Per second
benchmark.http_2xx                      516703      5741.11
benchmark.pool_overflow                 363         4.03
cluster_manager.cluster_added           4           0.04
default.total_match_count               4           0.04
membership_change                       4           0.04
runtime.load_success                    1           0.01
runtime.override_dir_not_exists         1           0.01
upstream_cx_http1_total                 37          0.41
upstream_cx_rx_bytes_total              81122371    901354.66
upstream_cx_total                       37          0.41
upstream_cx_tx_bytes_total              22218702    246873.09
upstream_rq_pending_overflow            363         4.03
upstream_rq_pending_total               37          0.41
upstream_rq_total                       516714      5741.23

[10:09:55.503051][21][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[10:10:00.503918][23][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[10:10:00.505088][1][I] Done.

```

</details>

<details>
<summary>scale-down-httproutes-100</summary>

```plaintext
Warning: 30 10:10:18.056][1][warning][misc] [external/envoy/source/common/protobuf/message_validator_impl.cc:21] Deprecated field: type envoy.config.core.v3.HeaderValueOption Using deprecated option 'envoy.config.core.v3.HeaderValueOption.append' from file base.proto. This configuration will be removed from Envoy soon. Please see https://www.envoyproxy.io/docs/envoy/latest/version_history/version_history for details. If continued use of this field is absolutely necessary, see https://www.envoyproxy.io/docs/envoy/latest/configuration/operations/runtime#using-runtime-overrides-for-deprecated-features for how to apply a temporary and highly discouraged override.
[10:10:18.057130][1][I] Detected 4 (v)CPUs with affinity..
[10:10:18.057143][1][I] Starting 4 threads / event loops. Time limit: 90 seconds.
[10:10:18.057145][1][I] Global targets: 400 connections and 40000 calls per second.
[10:10:18.057147][1][I]    (Per-worker targets: 100 connections and 10000 calls per second)
[10:11:48.759104][18][I] Stopping after 90000 ms. Initiated: 209011 / Completed: 208988. (Completion rate was 2322.087186024952 per second.)
[10:11:48.759212][21][I] Stopping after 90000 ms. Initiated: 49528 / Completed: 49526. (Completion rate was 550.2881062569155 per second.)
[10:11:48.759225][19][I] Stopping after 90000 ms. Initiated: 82374 / Completed: 82369. (Completion rate was 915.2096671136363 per second.)
[10:11:48.759235][22][I] Stopping after 90000 ms. Initiated: 83103 / Completed: 83098. (Completion rate was 923.3098082183817 per second.)
[10:11:54.582462][18][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
Nighthawk - A layer 7 protocol benchmarking tool.

benchmark_http_client.latency_2xx (423617 samples)
  min: 0s 000ms 341us | mean: 0s 007ms 208us | max: 0s 137ms 691us | pstdev: 0s 011ms 926us

  Percentile  Count       Value          
  0.5         211815      0s 003ms 194us 
  0.75        317722      0s 006ms 279us 
  0.8         338894      0s 007ms 606us 
  0.9         381256      0s 016ms 511us 
  0.95        402437      0s 035ms 002us 
  0.990625    419646      0s 062ms 793us 
  0.99902344  423204      0s 082ms 636us 

Queueing and connection setup latency (423652 samples)
  min: 0s 000ms 001us | mean: 0s 000ms 014us | max: 0s 035ms 192us | pstdev: 0s 000ms 210us

  Percentile  Count       Value          
  0.5         212529      0s 000ms 003us 
  0.75        317890      0s 000ms 008us 
  0.8         339187      0s 000ms 009us 
  0.9         381367      0s 000ms 009us 
  0.95        402534      0s 000ms 011us 
  0.990625    419681      0s 000ms 121us 
  0.99902344  423239      0s 001ms 867us 

Request start to response end (423617 samples)
  min: 0s 000ms 341us | mean: 0s 007ms 207us | max: 0s 137ms 691us | pstdev: 0s 011ms 926us

  Percentile  Count       Value          
  0.5         211809      0s 003ms 194us 
  0.75        317719      0s 006ms 279us 
  0.8         338894      0s 007ms 606us 
  0.9         381256      0s 016ms 510us 
  0.95        402437      0s 035ms 002us 
  0.990625    419646      0s 062ms 793us 
  0.99902344  423204      0s 082ms 636us 

Response body size in bytes (423617 samples)
  min: 10 | mean: 10 | max: 10 | pstdev: 0

Response header size in bytes (423617 samples)
  min: 110 | mean: 110 | max: 110 | pstdev: 0

Blocking. Results are skewed when significant numbers are reported here. (142751 samples)
  min: 0s 000ms 035us | mean: 0s 002ms 521us | max: 0s 105ms 373us | pstdev: 0s 006ms 732us

  Percentile  Count       Value          
  0.5         71376       0s 000ms 888us 
  0.75        107064      0s 001ms 861us 
  0.8         114202      0s 002ms 248us 
  0.9         128476      0s 004ms 006us 
  0.95        135614      0s 008ms 008us 
  0.990625    141413      0s 041ms 697us 
  0.99902344  142612      0s 064ms 958us 

Initiation to completion (423981 samples)
  min: 0s 000ms 005us | mean: 0s 007ms 283us | max: 0s 137ms 756us | pstdev: 0s 011ms 948us

  Percentile  Count       Value          
  0.5         211995      0s 003ms 260us 
  0.75        317989      0s 006ms 396us 
  0.8         339189      0s 007ms 761us 
  0.9         381583      0s 016ms 687us 
  0.95        402787      0s 035ms 110us 
  0.990625    420007      0s 062ms 932us 
  0.99902344  423567      0s 082ms 649us 

Counter                                 Value       Per second
benchmark.http_2xx                      423617      4706.85
benchmark.pool_overflow                 364         4.04
cluster_manager.cluster_added           4           0.04
default.total_match_count               4           0.04
membership_change                       4           0.04
runtime.load_success                    1           0.01
runtime.override_dir_not_exists         1           0.01
upstream_cx_http1_total                 36          0.40
upstream_cx_rx_bytes_total              66507869    738975.37
upstream_cx_total                       36          0.40
upstream_cx_tx_bytes_total              18217036    202411.25
upstream_rq_pending_overflow            364         4.04
upstream_rq_pending_total               36          0.40
upstream_rq_total                       423652      4707.24

[10:11:59.584852][19][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[10:12:04.585643][21][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[10:12:09.586270][22][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[10:12:09.587436][1][I] Done.

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
Warning: 30 10:15:03.038][1][warning][misc] [external/envoy/source/common/protobuf/message_validator_impl.cc:21] Deprecated field: type envoy.config.core.v3.HeaderValueOption Using deprecated option 'envoy.config.core.v3.HeaderValueOption.append' from file base.proto. This configuration will be removed from Envoy soon. Please see https://www.envoyproxy.io/docs/envoy/latest/version_history/version_history for details. If continued use of this field is absolutely necessary, see https://www.envoyproxy.io/docs/envoy/latest/configuration/operations/runtime#using-runtime-overrides-for-deprecated-features for how to apply a temporary and highly discouraged override.
[10:15:03.039299][1][I] Detected 4 (v)CPUs with affinity..
[10:15:03.039310][1][I] Starting 4 threads / event loops. Time limit: 90 seconds.
[10:15:03.039313][1][I] Global targets: 400 connections and 40000 calls per second.
[10:15:03.039314][1][I]    (Per-worker targets: 100 connections and 10000 calls per second)
[10:16:33.741189][23][I] Stopping after 90000 ms. Initiated: 258274 / Completed: 258258. (Completion rate was 2869.5316753816987 per second.)
[10:16:33.741450][18][I] Stopping after 90000 ms. Initiated: 35338 / Completed: 35338. (Completion rate was 392.64273862987994 per second.)
[10:16:33.741587][21][I] Stopping after 90000 ms. Initiated: 196085 / Completed: 196074. (Completion rate was 2178.588477687607 per second.)
[10:16:33.742737][19][I] Stopping after 90001 ms. Initiated: 35313 / Completed: 35313. (Completion rate was 392.3594516123065 per second.)
[10:16:39.388073][21][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
Nighthawk - A layer 7 protocol benchmarking tool.

benchmark_http_client.latency_2xx (524619 samples)
  min: 0s 000ms 363us | mean: 0s 005ms 917us | max: 0s 093ms 347us | pstdev: 0s 011ms 507us

  Percentile  Count       Value          
  0.5         262326      0s 002ms 779us 
  0.75        393472      0s 004ms 063us 
  0.8         419697      0s 004ms 496us 
  0.9         472160      0s 006ms 384us 
  0.95        498391      0s 046ms 964us 
  0.990625    519704      0s 054ms 102us 
  0.99902344  524107      0s 058ms 939us 

Queueing and connection setup latency (524646 samples)
  min: 0s 000ms 001us | mean: 0s 000ms 010us | max: 0s 019ms 046us | pstdev: 0s 000ms 096us

  Percentile  Count       Value          
  0.5         262650      0s 000ms 003us 
  0.75        393534      0s 000ms 008us 
  0.8         420137      0s 000ms 009us 
  0.9         472311      0s 000ms 009us 
  0.95        498511      0s 000ms 011us 
  0.990625    519728      0s 000ms 076us 
  0.99902344  524134      0s 001ms 276us 

Request start to response end (524619 samples)
  min: 0s 000ms 363us | mean: 0s 005ms 917us | max: 0s 093ms 347us | pstdev: 0s 011ms 507us

  Percentile  Count       Value          
  0.5         262310      0s 002ms 778us 
  0.75        393465      0s 004ms 062us 
  0.8         419696      0s 004ms 496us 
  0.9         472160      0s 006ms 384us 
  0.95        498389      0s 046ms 962us 
  0.990625    519707      0s 054ms 102us 
  0.99902344  524107      0s 058ms 939us 

Response body size in bytes (524619 samples)
  min: 10 | mean: 10 | max: 10 | pstdev: 0

Response header size in bytes (524619 samples)
  min: 110 | mean: 110 | max: 110 | pstdev: 0

Blocking. Results are skewed when significant numbers are reported here. (158918 samples)
  min: 0s 000ms 034us | mean: 0s 002ms 264us | max: 0s 069ms 718us | pstdev: 0s 007ms 119us

  Percentile  Count       Value          
  0.5         79459       0s 000ms 812us 
  0.75        119189      0s 001ms 826us 
  0.8         127135      0s 002ms 141us 
  0.9         143027      0s 003ms 136us 
  0.95        150973      0s 004ms 425us 
  0.990625    157429      0s 048ms 680us 
  0.99902344  158763      0s 055ms 619us 

Initiation to completion (524983 samples)
  min: 0s 000ms 005us | mean: 0s 005ms 969us | max: 0s 093ms 364us | pstdev: 0s 011ms 505us

  Percentile  Count       Value          
  0.5         262493      0s 002ms 832us 
  0.75        393739      0s 004ms 131us 
  0.8         419987      0s 004ms 571us 
  0.9         472485      0s 006ms 477us 
  0.95        498740      0s 046ms 983us 
  0.990625    520063      0s 054ms 147us 
  0.99902344  524471      0s 059ms 058us 

Counter                                 Value       Per second
benchmark.http_2xx                      524619      5829.06
benchmark.pool_overflow                 364         4.04
cluster_manager.cluster_added           4           0.04
default.total_match_count               4           0.04
membership_change                       4           0.04
runtime.load_success                    1           0.01
runtime.override_dir_not_exists         1           0.01
upstream_cx_http1_total                 36          0.40
upstream_cx_rx_bytes_total              82365183    915162.15
upstream_cx_total                       36          0.40
upstream_cx_tx_bytes_total              22559778    250662.41
upstream_rq_pending_overflow            364         4.04
upstream_rq_pending_total               36          0.40
upstream_rq_total                       524646      5829.36

[10:16:44.389294][23][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[10:16:44.391101][1][I] Done.

```

</details>

### Metrics

|Benchmark Name           |Envoy Gateway Memory (MiB)|Envoy Gateway Total CPU (Seconds)|Envoy Proxy Memory: 4lvn5<sup>[1]</sup> (MiB)|
|-                        |-                         |-                                |-                                            |
|scale-up-httproutes-10   |78                        |0.39                             |7                                            |
|scale-up-httproutes-50   |87                        |1.87                             |11                                           |
|scale-up-httproutes-100  |167                       |8.9                              |20                                           |
|scale-up-httproutes-300  |673                       |167.86                           |81                                           |
|scale-up-httproutes-500  |1358                      |10.98                            |190                                          |
|scale-down-httproutes-300|643                       |5.59                             |84                                           |
|scale-down-httproutes-100|455                       |112.24                           |55                                           |
|scale-down-httproutes-50 |147                       |150.17                           |20                                           |
|scale-down-httproutes-10 |120                       |151.69                           |12                                           |
1. envoy-gateway-system/envoy-benchmark-test-benchmark-0520098c-dbf5d95fb-4lvn5
