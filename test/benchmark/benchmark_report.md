# Benchmark Report

Benchmark test settings:

|RPS |Connections|Duration (Seconds)|CPU Limits (m)|Memory Limits (MiB)|
|-   |-          |-                 |-             |-                  |
|1000|100        |90                |1000          |2048               |

## Test: ScaleHTTPRoute

Fixed one Gateway and different scales of HTTPRoutes.


### Results

Click to see the full results.


<details>
<summary>scale-up-httproutes-10</summary>

```plaintext
[2024-06-25 14:23:36.545][1][warning][misc] [external/envoy/source/common/protobuf/message_validator_impl.cc:21] Deprecated field: type envoy.config.core.v3.HeaderValueOption Using deprecated option 'envoy.config.core.v3.HeaderValueOption.append' from file base.proto. This configuration will be removed from Envoy soon. Please see https://www.envoyproxy.io/docs/envoy/latest/version_history/version_history for details. If continued use of this field is absolutely necessary, see https://www.envoyproxy.io/docs/envoy/latest/configuration/operations/runtime#using-runtime-overrides-for-deprecated-features for how to apply a temporary and highly discouraged override.
[14:23:36.545677][1][I] Detected 4 (v)CPUs with affinity..
[14:23:36.545689][1][I] Starting 4 threads / event loops. Time limit: 90 seconds.
[14:23:36.545691][1][I] Global targets: 400 connections and 4000 calls per second.
[14:23:36.545692][1][I]    (Per-worker targets: 100 connections and 1000 calls per second)
[14:25:07.247563][19][I] Stopping after 90000 ms. Initiated: 90000 / Completed: 90000. (Completion rate was 999.9998444444686 per second.)
[14:25:07.248019][20][I] Stopping after 90000 ms. Initiated: 90000 / Completed: 90000. (Completion rate was 999.9975333394178 per second.)
[14:25:07.248086][22][I] Stopping after 90000 ms. Initiated: 90000 / Completed: 90000. (Completion rate was 999.999577777956 per second.)
[14:25:07.248380][24][I] Stopping after 90000 ms. Initiated: 90000 / Completed: 90000. (Completion rate was 999.9990666675377 per second.)
[14:25:07.818186][1][I] Done.
Nighthawk - A layer 7 protocol benchmarking tool.

benchmark_http_client.latency_2xx (359802 samples)
  min: 0s 000ms 290us | mean: 0s 000ms 544us | max: 0s 062ms 072us | pstdev: 0s 000ms 863us

  Percentile  Count       Value          
  0.5         179926      0s 000ms 448us 
  0.75        269854      0s 000ms 530us 
  0.8         287861      0s 000ms 552us 
  0.9         323828      0s 000ms 684us 
  0.95        341812      0s 000ms 793us 
  0.990625    356429      0s 001ms 863us 
  0.99902344  359451      0s 009ms 407us 

Queueing and connection setup latency (359802 samples)
  min: 0s 000ms 002us | mean: 0s 000ms 011us | max: 0s 023ms 417us | pstdev: 0s 000ms 100us

  Percentile  Count       Value          
  0.5         179902      0s 000ms 010us 
  0.75        269909      0s 000ms 010us 
  0.8         288003      0s 000ms 011us 
  0.9         323850      0s 000ms 011us 
  0.95        341822      0s 000ms 012us 
  0.990625    356429      0s 000ms 029us 
  0.99902344  359451      0s 000ms 170us 

Request start to response end (359802 samples)
  min: 0s 000ms 289us | mean: 0s 000ms 543us | max: 0s 062ms 072us | pstdev: 0s 000ms 863us

  Percentile  Count       Value          
  0.5         179915      0s 000ms 448us 
  0.75        269861      0s 000ms 530us 
  0.8         287848      0s 000ms 552us 
  0.9         323824      0s 000ms 683us 
  0.95        341813      0s 000ms 792us 
  0.990625    356429      0s 001ms 862us 
  0.99902344  359451      0s 009ms 406us 

Response body size in bytes (359802 samples)
  min: 10 | mean: 10 | max: 10 | pstdev: 0

Response header size in bytes (359802 samples)
  min: 110 | mean: 110 | max: 110 | pstdev: 0

Blocking. Results are skewed when significant numbers are reported here. (2 samples)
  min: 0s 000ms 901us | mean: 0s 001ms 551us | max: 0s 002ms 202us | pstdev: 0s 000ms 650us

  Percentile  Count       Value          
  0.5         1           0s 000ms 901us 

Initiation to completion (360000 samples)
  min: 0s 000ms 006us | mean: 0s 000ms 562us | max: 0s 062ms 310us | pstdev: 0s 000ms 883us

  Percentile  Count       Value          
  0.5         180016      0s 000ms 465us 
  0.75        270003      0s 000ms 547us 
  0.8         288024      0s 000ms 570us 
  0.9         324002      0s 000ms 702us 
  0.95        342002      0s 000ms 815us 
  0.990625    356625      0s 001ms 912us 
  0.99902344  359649      0s 009ms 617us 

Counter                                 Value       Per second
benchmark.http_2xx                      359802      3997.80
benchmark.pool_overflow                 198         2.20
cluster_manager.cluster_added           4           0.04
default.total_match_count               4           0.04
membership_change                       4           0.04
runtime.load_success                    1           0.01
runtime.override_dir_not_exists         1           0.01
upstream_cx_http1_total                 108         1.20
upstream_cx_rx_bytes_total              56488914    627653.97
upstream_cx_total                       108         1.20
upstream_cx_tx_bytes_total              15471486    171905.23
upstream_rq_pending_overflow            198         2.20
upstream_rq_pending_total               108         1.20
upstream_rq_total                       359802      3997.80


```

</details>

<details>
<summary>scale-up-httproutes-50</summary>

```plaintext
[2024-06-25 14:25:18.533][1][warning][misc] [external/envoy/source/common/protobuf/message_validator_impl.cc:21] Deprecated field: type envoy.config.core.v3.HeaderValueOption Using deprecated option 'envoy.config.core.v3.HeaderValueOption.append' from file base.proto. This configuration will be removed from Envoy soon. Please see https://www.envoyproxy.io/docs/envoy/latest/version_history/version_history for details. If continued use of this field is absolutely necessary, see https://www.envoyproxy.io/docs/envoy/latest/configuration/operations/runtime#using-runtime-overrides-for-deprecated-features for how to apply a temporary and highly discouraged override.
[14:25:18.533824][1][I] Detected 4 (v)CPUs with affinity..
[14:25:18.533833][1][I] Starting 4 threads / event loops. Time limit: 90 seconds.
[14:25:18.533836][1][I] Global targets: 400 connections and 4000 calls per second.
[14:25:18.533837][1][I]    (Per-worker targets: 100 connections and 1000 calls per second)
[14:26:49.235731][18][I] Stopping after 89999 ms. Initiated: 90000 / Completed: 90000. (Completion rate was 1000.0000222222228 per second.)
[14:26:49.235977][19][I] Stopping after 90000 ms. Initiated: 90000 / Completed: 90000. (Completion rate was 1000 per second.)
[14:26:49.236240][21][I] Stopping after 90000 ms. Initiated: 90000 / Completed: 89999. (Completion rate was 999.9888111119815 per second.)
[14:26:49.236565][23][I] Stopping after 90000 ms. Initiated: 90000 / Completed: 89999. (Completion rate was 999.9879333448638 per second.)
[14:26:54.772502][21][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
Nighthawk - A layer 7 protocol benchmarking tool.

benchmark_http_client.latency_2xx (359912 samples)
  min: 0s 000ms 303us | mean: 0s 000ms 508us | max: 0s 031ms 259us | pstdev: 0s 000ms 449us

  Percentile  Count       Value          
  0.5         179967      0s 000ms 437us 
  0.75        269938      0s 000ms 509us 
  0.8         287950      0s 000ms 533us 
  0.9         323925      0s 000ms 643us 
  0.95        341919      0s 000ms 735us 
  0.990625    356538      0s 001ms 526us 
  0.99902344  359561      0s 006ms 589us 

Queueing and connection setup latency (359914 samples)
  min: 0s 000ms 001us | mean: 0s 000ms 011us | max: 0s 023ms 449us | pstdev: 0s 000ms 080us

  Percentile  Count       Value          
  0.5         180198      0s 000ms 010us 
  0.75        270795      0s 000ms 011us 
  0.8         288165      0s 000ms 011us 
  0.9         324019      0s 000ms 011us 
  0.95        341989      0s 000ms 011us 
  0.990625    356540      0s 000ms 025us 
  0.99902344  359564      0s 000ms 159us 

Request start to response end (359912 samples)
  min: 0s 000ms 302us | mean: 0s 000ms 508us | max: 0s 031ms 258us | pstdev: 0s 000ms 448us

  Percentile  Count       Value          
  0.5         179970      0s 000ms 437us 
  0.75        269934      0s 000ms 509us 
  0.8         287943      0s 000ms 533us 
  0.9         323921      0s 000ms 643us 
  0.95        341917      0s 000ms 735us 
  0.990625    356538      0s 001ms 525us 
  0.99902344  359561      0s 006ms 589us 

Response body size in bytes (359912 samples)
  min: 10 | mean: 10 | max: 10 | pstdev: 0

Response header size in bytes (359912 samples)
  min: 110 | mean: 110 | max: 110 | pstdev: 0

Initiation to completion (359998 samples)
  min: 0s 000ms 006us | mean: 0s 000ms 525us | max: 0s 032ms 263us | pstdev: 0s 000ms 461us

  Percentile  Count       Value          
  0.5         180024      0s 000ms 453us 
  0.75        270017      0s 000ms 526us 
  0.8         288004      0s 000ms 550us 
  0.9         323999      0s 000ms 663us 
  0.95        341999      0s 000ms 753us 
  0.990625    356624      0s 001ms 571us 
  0.99902344  359647      0s 006ms 655us 

Counter                                 Value       Per second
benchmark.http_2xx                      359912      3999.02
benchmark.pool_overflow                 86          0.96
cluster_manager.cluster_added           4           0.04
default.total_match_count               4           0.04
membership_change                       4           0.04
runtime.load_success                    1           0.01
runtime.override_dir_not_exists         1           0.01
upstream_cx_http1_total                 61          0.68
upstream_cx_rx_bytes_total              56506184    627846.33
upstream_cx_total                       61          0.68
upstream_cx_tx_bytes_total              15476302    171958.87
upstream_rq_pending_overflow            86          0.96
upstream_rq_pending_total               61          0.68
upstream_rq_total                       359914      3999.04

[14:26:59.773726][23][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[14:26:59.775576][1][I] Done.

```

</details>

<details>
<summary>scale-up-httproutes-100</summary>

```plaintext
[2024-06-25 14:27:18.024][1][warning][misc] [external/envoy/source/common/protobuf/message_validator_impl.cc:21] Deprecated field: type envoy.config.core.v3.HeaderValueOption Using deprecated option 'envoy.config.core.v3.HeaderValueOption.append' from file base.proto. This configuration will be removed from Envoy soon. Please see https://www.envoyproxy.io/docs/envoy/latest/version_history/version_history for details. If continued use of this field is absolutely necessary, see https://www.envoyproxy.io/docs/envoy/latest/configuration/operations/runtime#using-runtime-overrides-for-deprecated-features for how to apply a temporary and highly discouraged override.
[14:27:18.024969][1][I] Detected 4 (v)CPUs with affinity..
[14:27:18.024981][1][I] Starting 4 threads / event loops. Time limit: 90 seconds.
[14:27:18.024984][1][I] Global targets: 400 connections and 4000 calls per second.
[14:27:18.024985][1][I]    (Per-worker targets: 100 connections and 1000 calls per second)
[14:28:48.726723][18][I] Stopping after 89999 ms. Initiated: 90000 / Completed: 90000. (Completion rate was 1000.0000111111112 per second.)
[14:28:48.726970][19][I] Stopping after 89999 ms. Initiated: 90000 / Completed: 90000. (Completion rate was 1000.0000222222228 per second.)
[14:28:48.727250][21][I] Stopping after 90000 ms. Initiated: 90000 / Completed: 89999. (Completion rate was 999.9885555593705 per second.)
[14:28:48.727579][23][I] Stopping after 90000 ms. Initiated: 90000 / Completed: 89999. (Completion rate was 999.9878444571402 per second.)
[14:28:54.260999][21][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
Nighthawk - A layer 7 protocol benchmarking tool.

benchmark_http_client.latency_2xx (359869 samples)
  min: 0s 000ms 301us | mean: 0s 000ms 514us | max: 0s 029ms 035us | pstdev: 0s 000ms 431us

  Percentile  Count       Value          
  0.5         179965      0s 000ms 444us 
  0.75        269904      0s 000ms 518us 
  0.8         287910      0s 000ms 540us 
  0.9         323888      0s 000ms 641us 
  0.95        341879      0s 000ms 744us 
  0.990625    356497      0s 001ms 522us 
  0.99902344  359518      0s 006ms 553us 

Queueing and connection setup latency (359871 samples)
  min: 0s 000ms 001us | mean: 0s 000ms 011us | max: 0s 027ms 366us | pstdev: 0s 000ms 058us

  Percentile  Count       Value          
  0.5         179998      0s 000ms 010us 
  0.75        271128      0s 000ms 011us 
  0.8         288699      0s 000ms 011us 
  0.9         324073      0s 000ms 011us 
  0.95        341895      0s 000ms 012us 
  0.990625    356498      0s 000ms 029us 
  0.99902344  359520      0s 000ms 166us 

Request start to response end (359869 samples)
  min: 0s 000ms 301us | mean: 0s 000ms 513us | max: 0s 029ms 035us | pstdev: 0s 000ms 431us

  Percentile  Count       Value          
  0.5         179935      0s 000ms 444us 
  0.75        269906      0s 000ms 517us 
  0.8         287896      0s 000ms 539us 
  0.9         323883      0s 000ms 641us 
  0.95        341876      0s 000ms 743us 
  0.990625    356496      0s 001ms 521us 
  0.99902344  359518      0s 006ms 552us 

Response body size in bytes (359869 samples)
  min: 10 | mean: 10 | max: 10 | pstdev: 0

Response header size in bytes (359869 samples)
  min: 110 | mean: 110 | max: 110 | pstdev: 0

Initiation to completion (359998 samples)
  min: 0s 000ms 007us | mean: 0s 000ms 532us | max: 0s 029ms 055us | pstdev: 0s 000ms 451us

  Percentile  Count       Value          
  0.5         180023      0s 000ms 461us 
  0.75        270008      0s 000ms 535us 
  0.8         288030      0s 000ms 557us 
  0.9         324004      0s 000ms 661us 
  0.95        341999      0s 000ms 763us 
  0.990625    356624      0s 001ms 570us 
  0.99902344  359647      0s 007ms 023us 

Counter                                 Value       Per second
benchmark.http_2xx                      359869      3998.54
benchmark.pool_overflow                 129         1.43
cluster_manager.cluster_added           4           0.04
default.total_match_count               4           0.04
membership_change                       4           0.04
runtime.load_success                    1           0.01
runtime.override_dir_not_exists         1           0.01
upstream_cx_http1_total                 63          0.70
upstream_cx_rx_bytes_total              56499433    627771.26
upstream_cx_total                       63          0.70
upstream_cx_tx_bytes_total              15474453    171938.31
upstream_rq_pending_overflow            129         1.43
upstream_rq_pending_total               63          0.70
upstream_rq_total                       359871      3998.57

[14:28:59.262084][23][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[14:28:59.263795][1][I] Done.

```

</details>

<details>
<summary>scale-up-httproutes-300</summary>

```plaintext
[2024-06-25 14:32:23.491][1][warning][misc] [external/envoy/source/common/protobuf/message_validator_impl.cc:21] Deprecated field: type envoy.config.core.v3.HeaderValueOption Using deprecated option 'envoy.config.core.v3.HeaderValueOption.append' from file base.proto. This configuration will be removed from Envoy soon. Please see https://www.envoyproxy.io/docs/envoy/latest/version_history/version_history for details. If continued use of this field is absolutely necessary, see https://www.envoyproxy.io/docs/envoy/latest/configuration/operations/runtime#using-runtime-overrides-for-deprecated-features for how to apply a temporary and highly discouraged override.
[14:32:23.491963][1][I] Detected 4 (v)CPUs with affinity..
[14:32:23.491978][1][I] Starting 4 threads / event loops. Time limit: 90 seconds.
[14:32:23.491980][1][I] Global targets: 400 connections and 4000 calls per second.
[14:32:23.491981][1][I]    (Per-worker targets: 100 connections and 1000 calls per second)
[14:33:54.193887][18][I] Stopping after 90000 ms. Initiated: 90000 / Completed: 90000. (Completion rate was 999.9999888888891 per second.)
[14:33:54.194133][19][I] Stopping after 89999 ms. Initiated: 90000 / Completed: 90000. (Completion rate was 1000.0000222222228 per second.)
[14:33:54.194387][21][I] Stopping after 90000 ms. Initiated: 90000 / Completed: 90000. (Completion rate was 999.9999888888891 per second.)
[14:33:54.194633][23][I] Stopping after 90000 ms. Initiated: 90000 / Completed: 90000. (Completion rate was 1000 per second.)
[14:33:54.712898][1][I] Done.
Nighthawk - A layer 7 protocol benchmarking tool.

benchmark_http_client.latency_2xx (359845 samples)
  min: 0s 000ms 302us | mean: 0s 000ms 507us | max: 0s 019ms 987us | pstdev: 0s 000ms 380us

  Percentile  Count       Value          
  0.5         179931      0s 000ms 440us 
  0.75        269886      0s 000ms 512us 
  0.8         287876      0s 000ms 533us 
  0.9         323864      0s 000ms 637us 
  0.95        341854      0s 000ms 734us 
  0.990625    356472      0s 001ms 469us 
  0.99902344  359494      0s 006ms 259us 

Queueing and connection setup latency (359845 samples)
  min: 0s 000ms 001us | mean: 0s 000ms 011us | max: 0s 017ms 634us | pstdev: 0s 000ms 043us

  Percentile  Count       Value          
  0.5         180160      0s 000ms 010us 
  0.75        270791      0s 000ms 010us 
  0.8         287918      0s 000ms 011us 
  0.9         324365      0s 000ms 011us 
  0.95        341940      0s 000ms 011us 
  0.990625    356473      0s 000ms 027us 
  0.99902344  359494      0s 000ms 167us 

Request start to response end (359845 samples)
  min: 0s 000ms 302us | mean: 0s 000ms 506us | max: 0s 019ms 987us | pstdev: 0s 000ms 380us

  Percentile  Count       Value          
  0.5         179923      0s 000ms 440us 
  0.75        269893      0s 000ms 511us 
  0.8         287905      0s 000ms 533us 
  0.9         323861      0s 000ms 636us 
  0.95        341854      0s 000ms 733us 
  0.990625    356472      0s 001ms 468us 
  0.99902344  359494      0s 006ms 259us 

Response body size in bytes (359845 samples)
  min: 10 | mean: 10 | max: 10 | pstdev: 0

Response header size in bytes (359845 samples)
  min: 110 | mean: 110 | max: 110 | pstdev: 0

Initiation to completion (360000 samples)
  min: 0s 000ms 007us | mean: 0s 000ms 525us | max: 0s 020ms 035us | pstdev: 0s 000ms 409us

  Percentile  Count       Value          
  0.5         180011      0s 000ms 456us 
  0.75        270028      0s 000ms 529us 
  0.8         288004      0s 000ms 550us 
  0.9         324006      0s 000ms 656us 
  0.95        342005      0s 000ms 751us 
  0.990625    356625      0s 001ms 525us 
  0.99902344  359649      0s 006ms 569us 

Counter                                 Value       Per second
benchmark.http_2xx                      359845      3998.28
benchmark.pool_overflow                 155         1.72
cluster_manager.cluster_added           4           0.04
default.total_match_count               4           0.04
membership_change                       4           0.04
runtime.load_success                    1           0.01
runtime.override_dir_not_exists         1           0.01
upstream_cx_http1_total                 62          0.69
upstream_cx_rx_bytes_total              56495665    627729.61
upstream_cx_total                       62          0.69
upstream_cx_tx_bytes_total              15473335    171925.94
upstream_rq_pending_overflow            155         1.72
upstream_rq_pending_total               62          0.69
upstream_rq_total                       359845      3998.28


```

</details>

<details>
<summary>scale-up-httproutes-500</summary>

```plaintext
[2024-06-25 14:38:43.691][1][warning][misc] [external/envoy/source/common/protobuf/message_validator_impl.cc:21] Deprecated field: type envoy.config.core.v3.HeaderValueOption Using deprecated option 'envoy.config.core.v3.HeaderValueOption.append' from file base.proto. This configuration will be removed from Envoy soon. Please see https://www.envoyproxy.io/docs/envoy/latest/version_history/version_history for details. If continued use of this field is absolutely necessary, see https://www.envoyproxy.io/docs/envoy/latest/configuration/operations/runtime#using-runtime-overrides-for-deprecated-features for how to apply a temporary and highly discouraged override.
[14:38:43.691938][1][I] Detected 4 (v)CPUs with affinity..
[14:38:43.691951][1][I] Starting 4 threads / event loops. Time limit: 90 seconds.
[14:38:43.691953][1][I] Global targets: 400 connections and 4000 calls per second.
[14:38:43.691954][1][I]    (Per-worker targets: 100 connections and 1000 calls per second)
[14:40:14.393764][18][I] Stopping after 90000 ms. Initiated: 90000 / Completed: 89999. (Completion rate was 999.9885666703507 per second.)
[14:40:14.393981][19][I] Stopping after 89999 ms. Initiated: 90000 / Completed: 90000. (Completion rate was 1000.0000555555587 per second.)
[14:40:14.394312][21][I] Stopping after 90000 ms. Initiated: 90000 / Completed: 90000. (Completion rate was 999.9992666672044 per second.)
[14:40:14.394484][23][I] Stopping after 89999 ms. Initiated: 90000 / Completed: 90000. (Completion rate was 1000.0000333333344 per second.)
[14:40:19.954533][18][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
Nighthawk - A layer 7 protocol benchmarking tool.

benchmark_http_client.latency_2xx (359743 samples)
  min: 0s 000ms 311us | mean: 0s 000ms 566us | max: 0s 064ms 673us | pstdev: 0s 001ms 003us

  Percentile  Count       Value          
  0.5         179890      0s 000ms 441us 
  0.75        269813      0s 000ms 519us 
  0.8         287802      0s 000ms 544us 
  0.9         323771      0s 000ms 684us 
  0.95        341757      0s 000ms 790us 
  0.990625    356372      0s 002ms 465us 
  0.99902344  359392      0s 016ms 404us 

Queueing and connection setup latency (359744 samples)
  min: 0s 000ms 001us | mean: 0s 000ms 011us | max: 0s 040ms 116us | pstdev: 0s 000ms 093us

  Percentile  Count       Value          
  0.5         180015      0s 000ms 010us 
  0.75        269817      0s 000ms 011us 
  0.8         288758      0s 000ms 011us 
  0.9         324218      0s 000ms 011us 
  0.95        341791      0s 000ms 012us 
  0.990625    356372      0s 000ms 029us 
  0.99902344  359393      0s 000ms 172us 

Request start to response end (359743 samples)
  min: 0s 000ms 311us | mean: 0s 000ms 566us | max: 0s 064ms 673us | pstdev: 0s 001ms 003us

  Percentile  Count       Value          
  0.5         179874      0s 000ms 441us 
  0.75        269820      0s 000ms 519us 
  0.8         287800      0s 000ms 544us 
  0.9         323772      0s 000ms 684us 
  0.95        341758      0s 000ms 790us 
  0.990625    356372      0s 002ms 464us 
  0.99902344  359392      0s 016ms 404us 

Response body size in bytes (359743 samples)
  min: 10 | mean: 10 | max: 10 | pstdev: 0

Response header size in bytes (359743 samples)
  min: 110 | mean: 110 | max: 110 | pstdev: 0

Blocking. Results are skewed when significant numbers are reported here. (1 samples)
  min: 0s 001ms 331us | mean: 0s 001ms 331us | max: 0s 001ms 331us | pstdev: 0s 000ms 000us

Initiation to completion (359999 samples)
  min: 0s 000ms 005us | mean: 0s 000ms 593us | max: 0s 064ms 710us | pstdev: 0s 001ms 145us

  Percentile  Count       Value          
  0.5         180029      0s 000ms 458us 
  0.75        270014      0s 000ms 537us 
  0.8         288011      0s 000ms 562us 
  0.9         324008      0s 000ms 703us 
  0.95        342000      0s 000ms 813us 
  0.990625    356625      0s 002ms 640us 
  0.99902344  359648      0s 018ms 319us 

Counter                                 Value       Per second
benchmark.http_2xx                      359743      3997.14
benchmark.pool_overflow                 256         2.84
cluster_manager.cluster_added           4           0.04
default.total_match_count               4           0.04
membership_change                       4           0.04
runtime.load_success                    1           0.01
runtime.override_dir_not_exists         1           0.01
upstream_cx_http1_total                 96          1.07
upstream_cx_rx_bytes_total              56479651    627551.52
upstream_cx_total                       96          1.07
upstream_cx_tx_bytes_total              15468992    171877.65
upstream_rq_pending_overflow            256         2.84
upstream_rq_pending_total               96          1.07
upstream_rq_total                       359744      3997.15

[14:40:19.962562][1][I] Done.

```

</details>

<details>
<summary>scale-down-httproutes-300</summary>

```plaintext
[2024-06-25 14:40:47.629][1][warning][misc] [external/envoy/source/common/protobuf/message_validator_impl.cc:21] Deprecated field: type envoy.config.core.v3.HeaderValueOption Using deprecated option 'envoy.config.core.v3.HeaderValueOption.append' from file base.proto. This configuration will be removed from Envoy soon. Please see https://www.envoyproxy.io/docs/envoy/latest/version_history/version_history for details. If continued use of this field is absolutely necessary, see https://www.envoyproxy.io/docs/envoy/latest/configuration/operations/runtime#using-runtime-overrides-for-deprecated-features for how to apply a temporary and highly discouraged override.
[14:40:47.629688][1][I] Detected 4 (v)CPUs with affinity..
[14:40:47.629700][1][I] Starting 4 threads / event loops. Time limit: 90 seconds.
[14:40:47.629703][1][I] Global targets: 400 connections and 4000 calls per second.
[14:40:47.629704][1][I]    (Per-worker targets: 100 connections and 1000 calls per second)
[14:42:18.331523][18][I] Stopping after 90000 ms. Initiated: 90000 / Completed: 90000. (Completion rate was 999.9990222231783 per second.)
[14:42:18.331681][19][I] Stopping after 90000 ms. Initiated: 90000 / Completed: 90000. (Completion rate was 999.9999777777783 per second.)
[14:42:18.331935][21][I] Stopping after 90000 ms. Initiated: 90000 / Completed: 89999. (Completion rate was 999.988855555927 per second.)
[14:42:18.332255][23][I] Stopping after 90000 ms. Initiated: 90000 / Completed: 90000. (Completion rate was 999.999166667361 per second.)
[14:42:23.961708][21][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
Nighthawk - A layer 7 protocol benchmarking tool.

benchmark_http_client.latency_2xx (359722 samples)
  min: 0s 000ms 303us | mean: 0s 001ms 203us | max: 0s 121ms 311us | pstdev: 0s 006ms 029us

  Percentile  Count       Value          
  0.5         179885      0s 000ms 456us 
  0.75        269810      0s 000ms 543us 
  0.8         287781      0s 000ms 570us 
  0.9         323750      0s 000ms 749us 
  0.95        341736      0s 001ms 344us 
  0.990625    356350      0s 016ms 743us 
  0.99902344  359371      0s 083ms 783us 

Queueing and connection setup latency (359723 samples)
  min: 0s 000ms 001us | mean: 0s 000ms 012us | max: 0s 041ms 897us | pstdev: 0s 000ms 131us

  Percentile  Count       Value          
  0.5         179921      0s 000ms 010us 
  0.75        270953      0s 000ms 011us 
  0.8         288381      0s 000ms 011us 
  0.9         323867      0s 000ms 011us 
  0.95        341759      0s 000ms 012us 
  0.990625    356351      0s 000ms 032us 
  0.99902344  359372      0s 000ms 206us 

Request start to response end (359722 samples)
  min: 0s 000ms 302us | mean: 0s 001ms 202us | max: 0s 121ms 311us | pstdev: 0s 006ms 029us

  Percentile  Count       Value          
  0.5         179877      0s 000ms 456us 
  0.75        269810      0s 000ms 542us 
  0.8         287783      0s 000ms 570us 
  0.9         323752      0s 000ms 748us 
  0.95        341736      0s 001ms 344us 
  0.990625    356350      0s 016ms 742us 
  0.99902344  359371      0s 083ms 783us 

Response body size in bytes (359722 samples)
  min: 10 | mean: 10 | max: 10 | pstdev: 0

Response header size in bytes (359722 samples)
  min: 110 | mean: 110 | max: 110 | pstdev: 0

Blocking. Results are skewed when significant numbers are reported here. (1022 samples)
  min: 0s 000ms 045us | mean: 0s 006ms 403us | max: 0s 099ms 049us | pstdev: 0s 015ms 130us

  Percentile  Count       Value          
  0.5         511         0s 001ms 012us 
  0.75        767         0s 003ms 262us 
  0.8         818         0s 004ms 782us 
  0.9         920         0s 015ms 482us 
  0.95        971         0s 047ms 316us 
  0.990625    1013        0s 075ms 657us 
  0.99902344  1022        0s 099ms 049us 

Initiation to completion (359999 samples)
  min: 0s 000ms 002us | mean: 0s 001ms 224us | max: 0s 121ms 405us | pstdev: 0s 006ms 041us

  Percentile  Count       Value          
  0.5         180002      0s 000ms 473us 
  0.75        270009      0s 000ms 560us 
  0.8         288002      0s 000ms 589us 
  0.9         324002      0s 000ms 770us 
  0.95        342001      0s 001ms 390us 
  0.990625    356625      0s 017ms 278us 
  0.99902344  359648      0s 083ms 849us 

Counter                                 Value       Per second
benchmark.http_2xx                      359722      3996.91
benchmark.pool_overflow                 277         3.08
cluster_manager.cluster_added           4           0.04
default.total_match_count               4           0.04
membership_change                       4           0.04
runtime.load_success                    1           0.01
runtime.override_dir_not_exists         1           0.01
upstream_cx_http1_total                 123         1.37
upstream_cx_rx_bytes_total              56476354    627514.75
upstream_cx_total                       123         1.37
upstream_cx_tx_bytes_total              15468089    171867.57
upstream_rq_pending_overflow            277         3.08
upstream_rq_pending_total               123         1.37
upstream_rq_total                       359723      3996.92

[14:42:23.965353][1][I] Done.

```

</details>

<details>
<summary>scale-down-httproutes-100</summary>

```plaintext
[2024-06-25 14:42:41.429][1][warning][misc] [external/envoy/source/common/protobuf/message_validator_impl.cc:21] Deprecated field: type envoy.config.core.v3.HeaderValueOption Using deprecated option 'envoy.config.core.v3.HeaderValueOption.append' from file base.proto. This configuration will be removed from Envoy soon. Please see https://www.envoyproxy.io/docs/envoy/latest/version_history/version_history for details. If continued use of this field is absolutely necessary, see https://www.envoyproxy.io/docs/envoy/latest/configuration/operations/runtime#using-runtime-overrides-for-deprecated-features for how to apply a temporary and highly discouraged override.
[14:42:41.430321][1][I] Detected 4 (v)CPUs with affinity..
[14:42:41.430334][1][I] Starting 4 threads / event loops. Time limit: 90 seconds.
[14:42:41.430336][1][I] Global targets: 400 connections and 4000 calls per second.
[14:42:41.430338][1][I]    (Per-worker targets: 100 connections and 1000 calls per second)
[14:44:12.132884][18][I] Stopping after 89999 ms. Initiated: 90000 / Completed: 89996. (Completion rate was 999.9555777767906 per second.)
[14:44:12.133405][21][I] Stopping after 90000 ms. Initiated: 90000 / Completed: 90000. (Completion rate was 999.999777777827 per second.)
[14:44:12.133681][23][I] Stopping after 90000 ms. Initiated: 89995 / Completed: 89995. (Completion rate was 999.9439111410252 per second.)
[14:44:12.138437][19][I] Stopping after 90005 ms. Initiated: 89979 / Completed: 89958. (Completion rate was 999.474453182769 per second.)
[14:44:17.976650][18][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
Nighthawk - A layer 7 protocol benchmarking tool.

benchmark_http_client.latency_2xx (359649 samples)
  min: 0s 000ms 297us | mean: 0s 009ms 493us | max: 0s 163ms 553us | pstdev: 0s 017ms 560us

  Percentile  Count       Value          
  0.5         179826      0s 002ms 042us 
  0.75        269740      0s 008ms 322us 
  0.8         287721      0s 011ms 606us 
  0.9         323685      0s 030ms 779us 
  0.95        341667      0s 054ms 808us 
  0.990625    356278      0s 079ms 679us 
  0.99902344  359299      0s 106ms 098us 

Queueing and connection setup latency (359674 samples)
  min: 0s 000ms 001us | mean: 0s 000ms 015us | max: 0s 033ms 947us | pstdev: 0s 000ms 269us

  Percentile  Count       Value          
  0.5         179931      0s 000ms 008us 
  0.75        269789      0s 000ms 010us 
  0.8         287838      0s 000ms 011us 
  0.9         323819      0s 000ms 011us 
  0.95        341697      0s 000ms 015us 
  0.990625    356303      0s 000ms 059us 
  0.99902344  359323      0s 001ms 303us 

Request start to response end (359649 samples)
  min: 0s 000ms 296us | mean: 0s 009ms 493us | max: 0s 163ms 553us | pstdev: 0s 017ms 560us

  Percentile  Count       Value          
  0.5         179829      0s 002ms 041us 
  0.75        269740      0s 008ms 321us 
  0.8         287722      0s 011ms 606us 
  0.9         323687      0s 030ms 779us 
  0.95        341667      0s 054ms 808us 
  0.990625    356278      0s 079ms 679us 
  0.99902344  359299      0s 106ms 098us 

Response body size in bytes (359649 samples)
  min: 10 | mean: 10 | max: 10 | pstdev: 0

Response header size in bytes (359649 samples)
  min: 110 | mean: 110 | max: 110 | pstdev: 0

Blocking. Results are skewed when significant numbers are reported here. (19052 samples)
  min: 0s 000ms 039us | mean: 0s 006ms 142us | max: 0s 146ms 767us | pstdev: 0s 012ms 746us

  Percentile  Count       Value          
  0.5         9526        0s 001ms 273us 
  0.75        14289       0s 004ms 540us 
  0.8         15242       0s 006ms 354us 
  0.9         17147       0s 018ms 240us 
  0.95        18100       0s 036ms 243us 
  0.990625    18874       0s 063ms 096us 
  0.99902344  19034       0s 087ms 937us 

Initiation to completion (359949 samples)
  min: 0s 000ms 006us | mean: 0s 009ms 592us | max: 0s 163ms 594us | pstdev: 0s 017ms 643us

  Percentile  Count       Value          
  0.5         179976      0s 002ms 093us 
  0.75        269962      0s 008ms 477us 
  0.8         287962      0s 011ms 824us 
  0.9         323955      0s 030ms 991us 
  0.95        341957      0s 055ms 142us 
  0.990625    356578      0s 080ms 228us 
  0.99902344  359598      0s 106ms 164us 

Counter                                 Value       Per second
benchmark.http_2xx                      359649      3996.04
benchmark.pool_overflow                 300         3.33
cluster_manager.cluster_added           4           0.04
default.total_match_count               4           0.04
membership_change                       4           0.04
runtime.load_success                    1           0.01
runtime.override_dir_not_exists         1           0.01
upstream_cx_http1_total                 100         1.11
upstream_cx_rx_bytes_total              56464893    627378.34
upstream_cx_total                       100         1.11
upstream_cx_tx_bytes_total              15465939    171841.20
upstream_rq_pending_overflow            300         3.33
upstream_rq_pending_total               100         1.11
upstream_rq_total                       359674      3996.32

[14:44:22.981221][19][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[14:44:22.985214][1][I] Done.

```

</details>

<details>
<summary>scale-down-httproutes-50</summary>

```plaintext
[2024-06-25 14:44:35.399][1][warning][misc] [external/envoy/source/common/protobuf/message_validator_impl.cc:21] Deprecated field: type envoy.config.core.v3.HeaderValueOption Using deprecated option 'envoy.config.core.v3.HeaderValueOption.append' from file base.proto. This configuration will be removed from Envoy soon. Please see https://www.envoyproxy.io/docs/envoy/latest/version_history/version_history for details. If continued use of this field is absolutely necessary, see https://www.envoyproxy.io/docs/envoy/latest/configuration/operations/runtime#using-runtime-overrides-for-deprecated-features for how to apply a temporary and highly discouraged override.
[14:44:35.400594][1][I] Detected 4 (v)CPUs with affinity..
[14:44:35.400604][1][I] Starting 4 threads / event loops. Time limit: 90 seconds.
[14:44:35.400607][1][I] Global targets: 400 connections and 4000 calls per second.
[14:44:35.400609][1][I]    (Per-worker targets: 100 connections and 1000 calls per second)
[14:45:50.873708][19][E] Exiting due to failing termination predicate
[14:45:50.873747][19][I] Stopping after 74767 ms. Initiated: 74767 / Completed: 74746. (Completion rate was 999.716774109537 per second.)
[14:45:50.874918][18][E] Exiting due to failing termination predicate
[14:45:50.875120][18][I] Stopping after 74769 ms. Initiated: 74770 / Completed: 74749. (Completion rate was 999.7209036132042 per second.)
[14:45:50.878184][20][E] Exiting due to failing termination predicate
[14:45:50.878203][20][I] Stopping after 74772 ms. Initiated: 74772 / Completed: 74750. (Completion rate was 999.6972288218956 per second.)
[14:45:50.880264][21][E] Exiting due to failing termination predicate
[14:45:50.880286][21][I] Stopping after 74773 ms. Initiated: 74774 / Completed: 74747. (Completion rate was 999.6424011911454 per second.)
[14:45:51.840758][1][E] Terminated early because of a failure predicate.
[14:45:51.840780][1][I] Check the output for problematic counter values. The default Nighthawk failure predicates report failure if (1) Nighthawk could not connect to the target (see 'benchmark.pool_connection_failure' counter; check the address and port number, and try explicitly setting --address-family v4 or v6, especially when using DNS; instead of localhost try 127.0.0.1 or ::1 explicitly), (2) the protocol was not supported by the target (see 'benchmark.stream_resets' counter; check http/https in the URI, --h2), (3) the target returned a 4xx or 5xx HTTP response code (see 'benchmark.http_4xx' and 'benchmark.http_5xx' counters; check the URI path and the server config), or (4) a custom gRPC RequestSource failed. --failure-predicate can be used to relax expectations.
[14:45:56.841540][18][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
Nighthawk - A layer 7 protocol benchmarking tool.

benchmark_http_client.latency_2xx (298738 samples)
  min: 0s 000ms 307us | mean: 0s 011ms 515us | max: 0s 187ms 269us | pstdev: 0s 019ms 783us

  Percentile  Count       Value          
  0.5         149369      0s 002ms 416us 
  0.75        224054      0s 011ms 230us 
  0.8         238992      0s 016ms 571us 
  0.9         268865      0s 041ms 666us 
  0.95        283803      0s 059ms 402us 
  0.990625    295939      0s 085ms 483us 
  0.99902344  298447      0s 120ms 631us 

benchmark_http_client.latency_5xx (7 samples)
  min: 0s 000ms 210us | mean: 0s 000ms 884us | max: 0s 001ms 337us | pstdev: 0s 000ms 404us

  Percentile  Count       Value          
  0.5         4           0s 000ms 683us 
  0.75        6           0s 001ms 309us 
  0.8         6           0s 001ms 309us 

Queueing and connection setup latency (298836 samples)
  min: 0s 000ms 001us | mean: 0s 000ms 016us | max: 0s 075ms 845us | pstdev: 0s 000ms 315us

  Percentile  Count       Value          
  0.5         149531      0s 000ms 008us 
  0.75        224310      0s 000ms 010us 
  0.8         239171      0s 000ms 010us 
  0.9         268973      0s 000ms 011us 
  0.95        283896      0s 000ms 016us 
  0.990625    296035      0s 000ms 061us 
  0.99902344  298545      0s 001ms 244us 

Request start to response end (298745 samples)
  min: 0s 000ms 210us | mean: 0s 011ms 514us | max: 0s 187ms 269us | pstdev: 0s 019ms 783us

  Percentile  Count       Value          
  0.5         149373      0s 002ms 415us 
  0.75        224059      0s 011ms 229us 
  0.8         238996      0s 016ms 569us 
  0.9         268871      0s 041ms 664us 
  0.95        283810      0s 059ms 402us 
  0.990625    295946      0s 085ms 483us 
  0.99902344  298454      0s 120ms 631us 

Response body size in bytes (298745 samples)
  min: 0 | mean: 9.999765686455003 | max: 10 | pstdev: 0.048405377254271256

Response header size in bytes (298745 samples)
  min: 58 | mean: 109.99878156956602 | max: 110 | pstdev: 0.25170796172221055

Blocking. Results are skewed when significant numbers are reported here. (8534 samples)
  min: 0s 000ms 040us | mean: 0s 006ms 742us | max: 0s 101ms 433us | pstdev: 0s 012ms 626us

  Percentile  Count       Value          
  0.5         4267        0s 001ms 369us 
  0.75        6401        0s 005ms 750us 
  0.8         6828        0s 008ms 406us 
  0.9         7681        0s 022ms 739us 
  0.95        8108        0s 037ms 554us 
  0.990625    8454        0s 058ms 765us 
  0.99902344  8526        0s 078ms 221us 

Initiation to completion (298992 samples)
  min: 0s 000ms 005us | mean: 0s 011ms 618us | max: 0s 187ms 277us | pstdev: 0s 019ms 857us

  Percentile  Count       Value          
  0.5         149496      0s 002ms 465us 
  0.75        224244      0s 011ms 490us 
  0.8         239195      0s 016ms 867us 
  0.9         269093      0s 041ms 924us 
  0.95        284043      0s 059ms 586us 
  0.990625    296189      0s 085ms 766us 
  0.99902344  298701      0s 120ms 954us 

Counter                                 Value       Per second
benchmark.http_2xx                      298738      3995.38
benchmark.http_5xx                      7           0.09
benchmark.pool_overflow                 247         3.30
cluster_manager.cluster_added           4           0.05
default.total_match_count               4           0.05
membership_change                       4           0.05
runtime.load_success                    1           0.01
runtime.override_dir_not_exists         1           0.01
sequencer.failed_terminations           4           0.05
upstream_cx_http1_total                 153         2.05
upstream_cx_rx_bytes_total              46902510    627283.31
upstream_cx_total                       153         2.05
upstream_cx_tx_bytes_total              12849948    171857.71
upstream_rq_pending_overflow            247         3.30
upstream_rq_pending_total               153         2.05
upstream_rq_total                       298836      3996.69

[14:46:01.842997][19][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[14:46:06.845068][20][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[14:46:11.847277][21][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[14:46:11.850206][1][E] An error ocurred.

```

</details>

<details>
<summary>scale-down-httproutes-10</summary>

```plaintext
[2024-06-25 14:48:06.510][1][warning][misc] [external/envoy/source/common/protobuf/message_validator_impl.cc:21] Deprecated field: type envoy.config.core.v3.HeaderValueOption Using deprecated option 'envoy.config.core.v3.HeaderValueOption.append' from file base.proto. This configuration will be removed from Envoy soon. Please see https://www.envoyproxy.io/docs/envoy/latest/version_history/version_history for details. If continued use of this field is absolutely necessary, see https://www.envoyproxy.io/docs/envoy/latest/configuration/operations/runtime#using-runtime-overrides-for-deprecated-features for how to apply a temporary and highly discouraged override.
[14:48:06.510881][1][I] Detected 4 (v)CPUs with affinity..
[14:48:06.510893][1][I] Starting 4 threads / event loops. Time limit: 90 seconds.
[14:48:06.510895][1][I] Global targets: 400 connections and 4000 calls per second.
[14:48:06.510897][1][I]    (Per-worker targets: 100 connections and 1000 calls per second)
[14:49:37.212970][19][I] Stopping after 90000 ms. Initiated: 90000 / Completed: 90000. (Completion rate was 999.9998000000401 per second.)
[14:49:37.213444][21][I] Stopping after 90000 ms. Initiated: 90000 / Completed: 90000. (Completion rate was 999.9973000072899 per second.)
[14:49:37.213541][23][I] Stopping after 90000 ms. Initiated: 90000 / Completed: 90000. (Completion rate was 999.9989777788228 per second.)
[14:49:37.213971][18][I] Stopping after 90001 ms. Initiated: 89999 / Completed: 89996. (Completion rate was 999.9414452707166 per second.)
Nighthawk - A layer 7 protocol benchmarking tool.

benchmark_http_client.latency_2xx (359809 samples)
  min: 0s 000ms 297us | mean: 0s 000ms 525us | max: 0s 048ms 130us | pstdev: 0s 000ms 466us

[14:49:42.728878][18][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
  Percentile  Count       Value          
  0.5         179925      0s 000ms 450us 
  0.75        269871      0s 000ms 525us 
  0.8         287855      0s 000ms 549us 
  0.9         323830      0s 000ms 661us 
  0.95        341819      0s 000ms 762us 
  0.990625    356436      0s 001ms 585us 
  0.99902344  359458      0s 007ms 084us 

Queueing and connection setup latency (359812 samples)
  min: 0s 000ms 002us | mean: 0s 000ms 012us | max: 0s 034ms 091us | pstdev: 0s 000ms 067us

  Percentile  Count       Value          
  0.5         180343      0s 000ms 010us 
  0.75        269955      0s 000ms 011us 
  0.8         288513      0s 000ms 011us 
  0.9         323857      0s 000ms 012us 
  0.95        341824      0s 000ms 023us 
  0.990625    356439      0s 000ms 050us 
  0.99902344  359461      0s 000ms 189us 

Request start to response end (359809 samples)
  min: 0s 000ms 297us | mean: 0s 000ms 524us | max: 0s 048ms 130us | pstdev: 0s 000ms 466us

  Percentile  Count       Value          
  0.5         179906      0s 000ms 450us 
  0.75        269861      0s 000ms 524us 
  0.8         287864      0s 000ms 549us 
  0.9         323829      0s 000ms 661us 
  0.95        341820      0s 000ms 761us 
  0.990625    356436      0s 001ms 585us 
  0.99902344  359458      0s 007ms 083us 

Response body size in bytes (359809 samples)
  min: 10 | mean: 10 | max: 10 | pstdev: 0

Response header size in bytes (359809 samples)
  min: 110 | mean: 110 | max: 110 | pstdev: 0

Initiation to completion (359996 samples)
  min: 0s 000ms 005us | mean: 0s 000ms 543us | max: 0s 048ms 146us | pstdev: 0s 000ms 476us

  Percentile  Count       Value          
  0.5         179999      0s 000ms 468us 
  0.75        269997      0s 000ms 543us 
  0.8         288002      0s 000ms 568us 
  0.9         324002      0s 000ms 683us 
  0.95        341998      0s 000ms 783us 
  0.990625    356622      0s 001ms 632us 
  0.99902344  359645      0s 007ms 247us 

Counter                                 Value       Per second
benchmark.http_2xx                      359809      3997.86
benchmark.pool_overflow                 187         2.08
cluster_manager.cluster_added           4           0.04
default.total_match_count               4           0.04
membership_change                       4           0.04
runtime.load_success                    1           0.01
runtime.override_dir_not_exists         1           0.01
upstream_cx_http1_total                 66          0.73
upstream_cx_rx_bytes_total              56490013    627663.98
upstream_cx_total                       66          0.73
upstream_cx_tx_bytes_total              15471916    171909.40
upstream_rq_pending_overflow            187         2.08
upstream_rq_pending_total               66          0.73
upstream_rq_total                       359812      3997.89

[14:49:42.732798][1][I] Done.

```

</details>

### Metrics

|Benchmark Name           |Envoy Gateway Memory (MiB)|Envoy Gateway Total CPU (Seconds)|Envoy Proxy Memory: k5s2s<sup>[1]</sup> (MiB)|
|-                        |-                         |-                                |-                                            |
|scale-up-httproutes-10   |76                        |0.34                             |7                                            |
|scale-up-httproutes-50   |114                       |1.94                             |12                                           |
|scale-up-httproutes-100  |195                       |10.85                            |20                                           |
|scale-up-httproutes-300  |1124                      |176.36                           |81                                           |
|scale-up-httproutes-500  |1588                      |16.69                            |190                                          |
|scale-down-httproutes-300|661                       |6.04                             |87                                           |
|scale-down-httproutes-100|679                       |104.73                           |95                                           |
|scale-down-httproutes-50 |143                       |172.84                           |53                                           |
|scale-down-httproutes-10 |118                       |174.16                           |18                                           |
1. envoy-gateway-system/envoy-benchmark-test-benchmark-0520098c-7668c94dd5-k5s2s
