# Benchmark Report

| RPS  | Connections | Duration (Seconds) | CPU Limits (m) | Memory Limits (MiB) |
|------|-------------|--------------------|----------------|---------------------|
| 1000 | 100         | 90                 | 1000           | 2048                |

## Test: ScaleHTTPRoute

Fixed one Gateway and different scales of HTTPRoutes.


### Results

Click to see the full results.


<details>
<summary>scale-up-httproutes-10</summary>

```plaintext

benchmark_http_client.latency_2xx (359873 samples)
  min: 0s 000ms 313us | mean: 0s 000ms 519us | max: 0s 072ms 552us | pstdev: 0s 000ms 793us

  Percentile  Count       Value          
  0.5         179946      0s 000ms 439us 
  0.75        269912      0s 000ms 513us 
  0.8         287903      0s 000ms 533us 
  0.9         323889      0s 000ms 614us 
  0.95        341884      0s 000ms 736us 
  0.990625    356500      0s 001ms 559us 
  0.99902344  359522      0s 007ms 985us 

Queueing and connection setup latency (359874 samples)
  min: 0s 000ms 002us | mean: 0s 000ms 012us | max: 0s 063ms 952us | pstdev: 0s 000ms 131us

  Percentile  Count       Value          
  0.5         180120      0s 000ms 010us 
  0.75        270199      0s 000ms 011us 
  0.8         288323      0s 000ms 011us 
  0.9         324255      0s 000ms 011us 
  0.95        341934      0s 000ms 012us 
  0.990625    356501      0s 000ms 031us 
  0.99902344  359523      0s 000ms 167us 

Request start to response end (359873 samples)
  min: 0s 000ms 312us | mean: 0s 000ms 518us | max: 0s 072ms 552us | pstdev: 0s 000ms 793us

  Percentile  Count       Value          
  0.5         179950      0s 000ms 438us 
  0.75        269913      0s 000ms 512us 
  0.8         287916      0s 000ms 533us 
  0.9         323890      0s 000ms 614us 
  0.95        341883      0s 000ms 736us 
  0.990625    356500      0s 001ms 558us 
  0.99902344  359522      0s 007ms 984us 

Response body size in bytes (359873 samples)
  min: 10 | mean: 10 | max: 10 | pstdev: 0

Response header size in bytes (359873 samples)
  min: 110 | mean: 110 | max: 110 | pstdev: 0

Blocking. Results are skewed when significant numbers are reported here. (1 samples)
  min: 0s 001ms 903us | mean: 0s 001ms 903us | max: 0s 001ms 903us | pstdev: 0s 000ms 000us

Initiation to completion (359999 samples)
  min: 0s 000ms 009us | mean: 0s 000ms 537us | max: 0s 072ms 634us | pstdev: 0s 000ms 841us

  Percentile  Count       Value          
  0.5         180023      0s 000ms 455us 
  0.75        270019      0s 000ms 530us 
  0.8         288001      0s 000ms 551us 
  0.9         324001      0s 000ms 634us 
  0.95        342004      0s 000ms 755us 
  0.990625    356625      0s 001ms 601us 
  0.99902344  359648      0s 008ms 163us 

Counter                                 Value       Per second
benchmark.http_2xx                      359873      3998.59
benchmark.pool_overflow                 126         1.40
cluster_manager.cluster_added           4           0.04
default.total_match_count               4           0.04
membership_change                       4           0.04
runtime.load_success                    1           0.01
runtime.override_dir_not_exists         1           0.01
upstream_cx_http1_total                 102         1.13
upstream_cx_rx_bytes_total              56500061    627778.28
upstream_cx_total                       102         1.13
upstream_cx_tx_bytes_total              15474582    171939.75
upstream_rq_pending_overflow            126         1.40
upstream_rq_pending_total               102         1.13
upstream_rq_total                       359874      3998.60

[09:24:22.102587][1][I] Done.

```

</details>

<details>
<summary>scale-up-httproutes-50</summary>

```plaintext

benchmark_http_client.latency_2xx (359812 samples)
  min: 0s 000ms 309us | mean: 0s 000ms 509us | max: 0s 038ms 768us | pstdev: 0s 000ms 426us

  Percentile  Count       Value          
  0.5         179913      0s 000ms 445us 
  0.75        269869      0s 000ms 519us 
  0.8         287867      0s 000ms 541us 
  0.9         323837      0s 000ms 663us 
  0.95        341824      0s 000ms 741us 
  0.990625    356439      0s 001ms 300us 
  0.99902344  359461      0s 005ms 292us 

Queueing and connection setup latency (359813 samples)
  min: 0s 000ms 001us | mean: 0s 000ms 011us | max: 0s 029ms 871us | pstdev: 0s 000ms 075us

  Percentile  Count       Value          
  0.5         180015      0s 000ms 010us 
  0.75        270757      0s 000ms 011us 
  0.8         287880      0s 000ms 011us 
  0.9         323906      0s 000ms 011us 
  0.95        341867      0s 000ms 012us 
  0.990625    356440      0s 000ms 027us 
  0.99902344  359462      0s 000ms 163us 

Request start to response end (359812 samples)
  min: 0s 000ms 305us | mean: 0s 000ms 509us | max: 0s 038ms 768us | pstdev: 0s 000ms 426us

  Percentile  Count       Value          
  0.5         179911      0s 000ms 445us 
  0.75        269873      0s 000ms 519us 
  0.8         287861      0s 000ms 540us 
  0.9         323831      0s 000ms 662us 
  0.95        341823      0s 000ms 741us 
  0.990625    356439      0s 001ms 299us 
  0.99902344  359461      0s 005ms 291us 

Response body size in bytes (359812 samples)
  min: 10 | mean: 10 | max: 10 | pstdev: 0

Response header size in bytes (359812 samples)
  min: 110 | mean: 110 | max: 110 | pstdev: 0

Initiation to completion (359999 samples)
  min: 0s 000ms 007us | mean: 0s 000ms 534us | max: 0s 038ms 785us | pstdev: 0s 000ms 618us

  Percentile  Count       Value          
  0.5         180006      0s 000ms 462us 
  0.75        270010      0s 000ms 537us 
  0.8         288017      0s 000ms 558us 
  0.9         324002      0s 000ms 682us 
  0.95        342008      0s 000ms 760us 
  0.990625    356625      0s 001ms 355us 
  0.99902344  359648      0s 006ms 391us 

Counter                                 Value       Per second
benchmark.http_2xx                      359812      3997.91
benchmark.pool_overflow                 187         2.08
cluster_manager.cluster_added           4           0.04
default.total_match_count               4           0.04
membership_change                       4           0.04
runtime.load_success                    1           0.01
runtime.override_dir_not_exists         1           0.01
upstream_cx_http1_total                 56          0.62
upstream_cx_rx_bytes_total              56490484    627671.58
upstream_cx_total                       56          0.62
upstream_cx_tx_bytes_total              15471959    171910.53
upstream_rq_pending_overflow            187         2.08
upstream_rq_pending_total               56          0.62
upstream_rq_total                       359813      3997.92

[09:26:13.509310][1][I] Done.

```

</details>

<details>
<summary>scale-up-httproutes-100</summary>

```plaintext

benchmark_http_client.latency_2xx (359853 samples)
  min: 0s 000ms 314us | mean: 0s 000ms 516us | max: 0s 035ms 172us | pstdev: 0s 000ms 430us

  Percentile  Count       Value          
  0.5         179943      0s 000ms 449us 
  0.75        269895      0s 000ms 523us 
  0.8         287893      0s 000ms 546us 
  0.9         323868      0s 000ms 666us 
  0.95        341865      0s 000ms 750us 
[09:27:57.164977][1][I] Done.
  0.990625    356480      0s 001ms 430us 
  0.99902344  359502      0s 005ms 762us 

Queueing and connection setup latency (359853 samples)
  min: 0s 000ms 002us | mean: 0s 000ms 011us | max: 0s 026ms 753us | pstdev: 0s 000ms 060us

  Percentile  Count       Value          
  0.5         180211      0s 000ms 010us 
  0.75        270918      0s 000ms 011us 
  0.8         288356      0s 000ms 011us 
  0.9         323954      0s 000ms 011us 
  0.95        341865      0s 000ms 012us 
  0.990625    356481      0s 000ms 029us 
  0.99902344  359502      0s 000ms 166us 

Request start to response end (359853 samples)
  min: 0s 000ms 313us | mean: 0s 000ms 516us | max: 0s 035ms 172us | pstdev: 0s 000ms 430us

  Percentile  Count       Value          
  0.5         179945      0s 000ms 448us 
  0.75        269895      0s 000ms 523us 
  0.8         287887      0s 000ms 545us 
  0.9         323881      0s 000ms 666us 
  0.95        341864      0s 000ms 749us 
  0.990625    356480      0s 001ms 430us 
  0.99902344  359502      0s 005ms 762us 

Response body size in bytes (359853 samples)
  min: 10 | mean: 10 | max: 10 | pstdev: 0

Response header size in bytes (359853 samples)
  min: 110 | mean: 110 | max: 110 | pstdev: 0

Initiation to completion (360000 samples)
  min: 0s 000ms 005us | mean: 0s 000ms 534us | max: 0s 035ms 190us | pstdev: 0s 000ms 447us

  Percentile  Count       Value          
  0.5         180008      0s 000ms 466us 
  0.75        270010      0s 000ms 540us 
  0.8         288009      0s 000ms 563us 
  0.9         324005      0s 000ms 686us 
  0.95        342003      0s 000ms 769us 
  0.990625    356625      0s 001ms 481us 
  0.99902344  359649      0s 006ms 056us 

Counter                                 Value       Per second
benchmark.http_2xx                      359853      3998.37
benchmark.pool_overflow                 147         1.63
cluster_manager.cluster_added           4           0.04
default.total_match_count               4           0.04
membership_change                       4           0.04
runtime.load_success                    1           0.01
runtime.override_dir_not_exists         1           0.01
upstream_cx_http1_total                 64          0.71
upstream_cx_rx_bytes_total              56496921    627743.57
upstream_cx_total                       64          0.71
upstream_cx_tx_bytes_total              15473679    171929.77
upstream_rq_pending_overflow            147         1.63
upstream_rq_pending_total               64          0.71
upstream_rq_total                       359853      3998.37


```

</details>

<details>
<summary>scale-up-httproutes-300</summary>

```plaintext

benchmark_http_client.latency_2xx (359847 samples)
  min: 0s 000ms 304us | mean: 0s 000ms 521us | max: 0s 024ms 524us | pstdev: 0s 000ms 424us

  Percentile  Count       Value          
  0.5         179935      0s 000ms 454us 
  0.75        269895      0s 000ms 527us 
  0.8         287881      0s 000ms 548us 
  0.9         323864      0s 000ms 662us 
  0.95        341856      0s 000ms 754us 
  0.990625    356474      0s 001ms 504us 
  0.99902344  359496      0s 006ms 789us 

Queueing and connection setup latency (359848 samples)
  min: 0s 000ms 001us | mean: 0s 000ms 011us | max: 0s 015ms 363us | pstdev: 0s 000ms 045us

  Percentile  Count       Value          
  0.5         179968      0s 000ms 010us 
  0.75        269895      0s 000ms 011us 
  0.8         288448      0s 000ms 011us 
  0.9         324002      0s 000ms 011us 
  0.95        341858      0s 000ms 012us 
  0.990625    356477      0s 000ms 029us 
  0.99902344  359497      0s 000ms 167us 

Request start to response end (359847 samples)
  min: 0s 000ms 304us | mean: 0s 000ms 521us | max: 0s 024ms 524us | pstdev: 0s 000ms 423us

  Percentile  Count       Value          
  0.5         179945      0s 000ms 454us 
  0.75        269914      0s 000ms 526us 
  0.8         287893      0s 000ms 547us 
  0.9         323864      0s 000ms 661us 
  0.95        341856      0s 000ms 754us 
  0.990625    356474      0s 001ms 503us 
  0.99902344  359496      0s 006ms 786us 

Response body size in bytes (359847 samples)
  min: 10 | mean: 10 | max: 10 | pstdev: 0

Response header size in bytes (359847 samples)
  min: 110 | mean: 110 | max: 110 | pstdev: 0

Initiation to completion (359999 samples)
  min: 0s 000ms 006us | mean: 0s 000ms 538us | max: 0s 024ms 556us | pstdev: 0s 000ms 433us

  Percentile  Count       Value          
  0.5         180015      0s 000ms 471us 
  0.75        270008      0s 000ms 544us 
  0.8         288021      0s 000ms 565us 
  0.9         324004      0s 000ms 681us 
  0.95        342002      0s 000ms 773us 
  0.990625    356625      0s 001ms 545us 
  0.99902344  359648      0s 006ms 943us 

Counter                                 Value       Per second
benchmark.http_2xx                      359847      3998.30
benchmark.pool_overflow                 152         1.69
cluster_manager.cluster_added           4           0.04
default.total_match_count               4           0.04
membership_change                       4           0.04
runtime.load_success                    1           0.01
runtime.override_dir_not_exists         1           0.01
upstream_cx_http1_total                 70          0.78
upstream_cx_rx_bytes_total              56495979    627732.89
upstream_cx_total                       70          0.78
upstream_cx_tx_bytes_total              15473464    171927.32
upstream_rq_pending_overflow            152         1.69
upstream_rq_pending_total               70          0.78
upstream_rq_total                       359848      3998.31

[09:33:09.468007][1][I] Done.

```

</details>

<details>
<summary>scale-up-httproutes-500</summary>

```plaintext

benchmark_http_client.latency_2xx (359807 samples)
  min: 0s 000ms 310us | mean: 0s 000ms 560us | max: 0s 060ms 688us | pstdev: 0s 000ms 942us

  Percentile  Count       Value          
  0.5         179905      0s 000ms 456us 
  0.75        269868      0s 000ms 529us 
  0.8         287867      0s 000ms 550us 
  0.9         323830      0s 000ms 678us 
  0.95        341818      0s 000ms 791us 
  0.990625    356434      0s 002ms 103us 
  0.99902344  359456      0s 012ms 998us 

Queueing and connection setup latency (359808 samples)
  min: 0s 000ms 001us | mean: 0s 000ms 012us | max: 0s 027ms 772us | pstdev: 0s 000ms 104us

  Percentile  Count       Value          
  0.5         180194      0s 000ms 010us 
  0.75        270932      0s 000ms 011us 
  0.8         288826      0s 000ms 011us 
  0.9         324092      0s 000ms 011us 
  0.95        341848      0s 000ms 012us 
  0.990625    356435      0s 000ms 030us 
  0.99902344  359457      0s 000ms 170us 

Request start to response end (359807 samples)
  min: 0s 000ms 309us | mean: 0s 000ms 560us | max: 0s 060ms 688us | pstdev: 0s 000ms 942us

  Percentile  Count       Value          
  0.5         179908      0s 000ms 456us 
  0.75        269856      0s 000ms 529us 
  0.8         287848      0s 000ms 550us 
  0.9         323827      0s 000ms 677us 
  0.95        341817      0s 000ms 791us 
  0.990625    356434      0s 002ms 103us 
  0.99902344  359456      0s 012ms 998us 

Response body size in bytes (359807 samples)
  min: 10 | mean: 10 | max: 10 | pstdev: 0

Response header size in bytes (359807 samples)
  min: 110 | mean: 110 | max: 110 | pstdev: 0

Blocking. Results are skewed when significant numbers are reported here. (6 samples)
  min: 0s 000ms 201us | mean: 0s 000ms 928us | max: 0s 002ms 801us | pstdev: 0s 000ms 880us

  Percentile  Count       Value          
  0.5         3           0s 000ms 614us 
  0.75        5           0s 000ms 914us 
  0.8         5           0s 000ms 914us 

Initiation to completion (359999 samples)
  min: 0s 000ms 006us | mean: 0s 000ms 581us | max: 0s 060ms 710us | pstdev: 0s 000ms 978us

  Percentile  Count       Value          
  0.5         180010      0s 000ms 473us 
  0.75        270021      0s 000ms 547us 
  0.8         288019      0s 000ms 568us 
  0.9         324004      0s 000ms 697us 
  0.95        342000      0s 000ms 814us 
  0.990625    356625      0s 002ms 210us 
  0.99902344  359648      0s 014ms 801us 

Counter                                 Value       Per second
benchmark.http_2xx                      359807      3997.85
benchmark.pool_overflow                 192         2.13
cluster_manager.cluster_added           4           0.04
default.total_match_count               4           0.04
membership_change                       4           0.04
runtime.load_success                    1           0.01
runtime.override_dir_not_exists         1           0.01
upstream_cx_http1_total                 112         1.24
upstream_cx_rx_bytes_total              56489699    627663.06
upstream_cx_total                       112         1.24
upstream_cx_tx_bytes_total              15471744    171908.20
upstream_rq_pending_overflow            192         2.13
upstream_rq_pending_total               112         1.24
upstream_rq_total                       359808      3997.87

[09:40:33.045512][1][I] Done.

```

</details>

<details>
<summary>scale-down-httproutes-300</summary>

```plaintext

benchmark_http_client.latency_2xx (359812 samples)
  min: 0s 000ms 302us | mean: 0s 001ms 445us | max: 0s 183ms 287us | pstdev: 0s 007ms 612us

  Percentile  Count       Value          
  0.5         179921      0s 000ms 473us 
  0.75        269867      0s 000ms 556us 
  0.8         287850      0s 000ms 592us 
  0.9         323832      0s 000ms 760us 
  0.95        341822      0s 001ms 226us 
  0.990625    356439      0s 043ms 878us 
  0.99902344  359461      0s 095ms 641us 

Queueing and connection setup latency (359812 samples)
  min: 0s 000ms 001us | mean: 0s 000ms 012us | max: 0s 027ms 984us | pstdev: 0s 000ms 113us

  Percentile  Count       Value          
  0.5         180354      0s 000ms 010us 
  0.75        270110      0s 000ms 011us 
  0.8         288028      0s 000ms 011us 
  0.9         324203      0s 000ms 011us 
  0.95        341828      0s 000ms 012us 
  0.990625    356439      0s 000ms 033us 
  0.99902344  359461      0s 000ms 194us 

Request start to response end (359812 samples)
  min: 0s 000ms 302us | mean: 0s 001ms 444us | max: 0s 183ms 287us | pstdev: 0s 007ms 612us

  Percentile  Count       Value          
  0.5         179909      0s 000ms 472us 
  0.75        269876      0s 000ms 555us 
  0.8         287853      0s 000ms 592us 
  0.9         323833      0s 000ms 759us 
  0.95        341822      0s 001ms 226us 
  0.990625    356439      0s 043ms 878us 
  0.99902344  359461      0s 095ms 641us 

Response body size in bytes (359812 samples)
  min: 10 | mean: 10 | max: 10 | pstdev: 0

Response header size in bytes (359812 samples)
  min: 110 | mean: 110 | max: 110 | pstdev: 0

Blocking. Results are skewed when significant numbers are reported here. (565 samples)
  min: 0s 000ms 044us | mean: 0s 006ms 788us | max: 0s 095ms 117us | pstdev: 0s 015ms 888us

  Percentile  Count       Value          
  0.5         283         0s 000ms 744us 
  0.75        424         0s 002ms 919us 
  0.8         452         0s 005ms 369us 
  0.9         509         0s 018ms 845us 
  0.95        537         0s 047ms 767us 
  0.990625    560         0s 071ms 987us 

Initiation to completion (360000 samples)
  min: 0s 000ms 007us | mean: 0s 001ms 468us | max: 0s 183ms 320us | pstdev: 0s 007ms 622us

  Percentile  Count       Value          
  0.5         180014      0s 000ms 490us 
  0.75        270017      0s 000ms 573us 
  0.8         288000      0s 000ms 611us 
  0.9         324002      0s 000ms 780us 
  0.95        342000      0s 001ms 259us 
  0.990625    356625      0s 043ms 876us 
  0.99902344  359649      0s 095ms 735us 

Counter                                 Value       Per second
benchmark.http_2xx                      359812      3997.91
benchmark.pool_overflow                 188         2.09
cluster_manager.cluster_added           4           0.04
default.total_match_count               4           0.04
membership_change                       4           0.04
runtime.load_success                    1           0.01
runtime.override_dir_not_exists         1           0.01
upstream_cx_http1_total                 212         2.36
upstream_cx_rx_bytes_total              56490484    627671.56
upstream_cx_total                       212         2.36
upstream_cx_tx_bytes_total              15471916    171910.04
upstream_rq_pending_overflow            188         2.09
upstream_rq_pending_total               212         2.36
upstream_rq_total                       359812      3997.91


```

</details>

<details>
<summary>scale-down-httproutes-100</summary>

```plaintext

benchmark_http_client.latency_2xx (359688 samples)
  min: 0s 000ms 308us | mean: 0s 009ms 727us | max: 0s 186ms 793us | pstdev: 0s 019ms 123us

  Percentile  Count       Value          
  0.5         179844      0s 001ms 784us 
  0.75        269768      0s 007ms 040us 
  0.8         287752      0s 009ms 941us 
  0.9         323723      0s 033ms 730us 
  0.95        341704      0s 061ms 620us 
  0.990625    356316      0s 083ms 664us 
  0.99902344  359337      0s 110ms 288us 

Queueing and connection setup latency (359702 samples)
  min: 0s 000ms 001us | mean: 0s 000ms 014us | max: 0s 049ms 557us | pstdev: 0s 000ms 258us

  Percentile  Count       Value          
  0.5         180026      0s 000ms 008us 
  0.75        269966      0s 000ms 010us 
  0.8         287847      0s 000ms 010us 
  0.9         323820      0s 000ms 011us 
  0.95        341720      0s 000ms 018us 
  0.990625    356330      0s 000ms 113us 
  0.99902344  359351      0s 000ms 988us 

Request start to response end (359688 samples)
  min: 0s 000ms 308us | mean: 0s 009ms 726us | max: 0s 186ms 793us | pstdev: 0s 019ms 123us

  Percentile  Count       Value          
  0.5         179845      0s 001ms 783us 
  0.75        269766      0s 007ms 039us 
  0.8         287751      0s 009ms 940us 
  0.9         323723      0s 033ms 730us 
  0.95        341705      0s 061ms 620us 
  0.990625    356316      0s 083ms 664us 
  0.99902344  359337      0s 110ms 288us 

Response body size in bytes (359688 samples)
  min: 10 | mean: 10 | max: 10 | pstdev: 0

Response header size in bytes (359688 samples)
  min: 110 | mean: 110 | max: 110 | pstdev: 0

Blocking. Results are skewed when significant numbers are reported here. (21695 samples)
  min: 0s 000ms 035us | mean: 0s 004ms 952us | max: 0s 143ms 073us | pstdev: 0s 012ms 407us

  Percentile  Count       Value          
  0.5         10848       0s 001ms 029us 
  0.75        16272       0s 003ms 006us 
  0.8         17356       0s 003ms 915us 
  0.9         19526       0s 009ms 562us 
  0.95        20611       0s 031ms 671us 
  0.990625    21492       0s 065ms 878us 
  0.99902344  21674       0s 083ms 513us 

Initiation to completion (359969 samples)
  min: 0s 000ms 135us | mean: 0s 009ms 787us | max: 0s 186ms 834us | pstdev: 0s 019ms 151us

  Percentile  Count       Value          
  0.5         179987      0s 001ms 820us 
  0.75        269979      0s 007ms 143us 
  0.8         287978      0s 010ms 083us 
  0.9         323973      0s 033ms 902us 
  0.95        341972      0s 061ms 700us 
  0.990625    356596      0s 083ms 812us 
  0.99902344  359618      0s 110ms 374us 

Counter                                 Value       Per second
benchmark.http_2xx                      359688      3996.53
benchmark.pool_overflow                 281         3.12
cluster_manager.cluster_added           4           0.04
default.total_match_count               4           0.04
membership_change                       4           0.04
runtime.load_success                    1           0.01
runtime.override_dir_not_exists         1           0.01
upstream_cx_http1_total                 119         1.32
upstream_cx_rx_bytes_total              56471016    627455.05
upstream_cx_total                       119         1.32
upstream_cx_tx_bytes_total              15467186    171857.44
upstream_rq_pending_overflow            281         3.12
upstream_rq_pending_total               119         1.32
upstream_rq_total                       359702      3996.68

[09:44:35.525681][19][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[09:44:40.533295][21][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[09:44:40.549139][1][I] Done.

```

</details>

<details>
<summary>scale-down-httproutes-50</summary>

```plaintext

benchmark_http_client.latency_2xx (359809 samples)
  min: 0s 000ms 298us | mean: 0s 000ms 544us | max: 0s 082ms 124us | pstdev: 0s 000ms 771us

  Percentile  Count       Value          
  0.5         179920      0s 000ms 458us 
  0.75        269873      0s 000ms 534us 
  0.8         287849      0s 000ms 558us 
  0.9         323834      0s 000ms 699us 
  0.95        341819      0s 000ms 778us 
  0.990625    356436      0s 001ms 748us 
  0.99902344  359458      0s 009ms 125us 

Queueing and connection setup latency (359811 samples)
  min: 0s 000ms 001us | mean: 0s 000ms 011us | max: 0s 019ms 383us | pstdev: 0s 000ms 048us

  Percentile  Count       Value          
  0.5         180077      0s 000ms 010us 
  0.75        271002      0s 000ms 011us 
  0.8         288142      0s 000ms 011us 
  0.9         323906      0s 000ms 011us 
  0.95        341880      0s 000ms 012us 
  0.990625    356438      0s 000ms 030us 
  0.99902344  359460      0s 000ms 172us 

Request start to response end (359809 samples)
  min: 0s 000ms 297us | mean: 0s 000ms 543us | max: 0s 082ms 124us | pstdev: 0s 000ms 771us

  Percentile  Count       Value          
  0.5         179929      0s 000ms 458us 
  0.75        269861      0s 000ms 534us 
  0.8         287865      0s 000ms 558us 
  0.9         323830      0s 000ms 698us 
  0.95        341819      0s 000ms 778us 
  0.990625    356436      0s 001ms 748us 
  0.99902344  359458      0s 009ms 124us 

Response body size in bytes (359809 samples)
  min: 10 | mean: 10 | max: 10 | pstdev: 0

Response header size in bytes (359809 samples)
  min: 110 | mean: 110 | max: 110 | pstdev: 0

Blocking. Results are skewed when significant numbers are reported here. (6 samples)
  min: 0s 000ms 140us | mean: 0s 001ms 861us | max: 0s 003ms 722us | pstdev: 0s 001ms 492us

  Percentile  Count       Value          
  0.5         3           0s 000ms 854us 
  0.75        5           0s 003ms 630us 
  0.8         5           0s 003ms 630us 

Initiation to completion (359998 samples)
  min: 0s 000ms 007us | mean: 0s 000ms 562us | max: 0s 082ms 153us | pstdev: 0s 000ms 779us

  Percentile  Count       Value          
  0.5         180001      0s 000ms 476us 
  0.75        270017      0s 000ms 552us 
  0.8         288016      0s 000ms 576us 
  0.9         324003      0s 000ms 717us 
  0.95        342000      0s 000ms 798us 
  0.990625    356624      0s 001ms 806us 
  0.99902344  359647      0s 009ms 598us 

Counter                                 Value       Per second
benchmark.http_2xx                      359809      3997.88
benchmark.pool_overflow                 189         2.10
cluster_manager.cluster_added           4           0.04
default.total_match_count               4           0.04
membership_change                       4           0.04
runtime.load_success                    1           0.01
runtime.override_dir_not_exists         1           0.01
upstream_cx_http1_total                 84          0.93
upstream_cx_rx_bytes_total              56490013    627666.63
upstream_cx_total                       84          0.93
upstream_cx_tx_bytes_total              15471873    171909.65
upstream_rq_pending_overflow            189         2.10
upstream_rq_pending_total               84          0.93
upstream_rq_total                       359811      3997.90

[09:47:56.790350][21][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[09:47:56.792949][1][I] Done.

```

</details>

<details>
<summary>scale-down-httproutes-10</summary>

```plaintext

benchmark_http_client.latency_2xx (359744 samples)
  min: 0s 000ms 300us | mean: 0s 000ms 552us | max: 0s 062ms 459us | pstdev: 0s 000ms 650us

  Percentile  Count       Value          
  0.5         179903      0s 000ms 462us 
  0.75        269816      0s 000ms 540us 
  0.8         287814      0s 000ms 563us 
  0.9         323771      0s 000ms 709us 
  0.95        341759      0s 000ms 807us 
  0.990625    356373      0s 001ms 984us 
  0.99902344  359393      0s 008ms 610us 

Queueing and connection setup latency (359746 samples)
  min: 0s 000ms 001us | mean: 0s 000ms 011us | max: 0s 017ms 278us | pstdev: 0s 000ms 043us

  Percentile  Count       Value          
  0.5         179917      0s 000ms 010us 
  0.75        270948      0s 000ms 011us 
  0.8         288210      0s 000ms 011us 
  0.9         323879      0s 000ms 011us 
  0.95        341762      0s 000ms 012us 
  0.990625    356374      0s 000ms 031us 
  0.99902344  359395      0s 000ms 173us 

Request start to response end (359744 samples)
  min: 0s 000ms 299us | mean: 0s 000ms 552us | max: 0s 062ms 459us | pstdev: 0s 000ms 650us

  Percentile  Count       Value          
  0.5         179873      0s 000ms 461us 
  0.75        269814      0s 000ms 540us 
  0.8         287814      0s 000ms 562us 
  0.9         323778      0s 000ms 709us 
  0.95        341757      0s 000ms 807us 
  0.990625    356372      0s 001ms 984us 
  0.99902344  359393      0s 008ms 610us 

Response body size in bytes (359744 samples)
  min: 10 | mean: 10 | max: 10 | pstdev: 0

Response header size in bytes (359744 samples)
  min: 110 | mean: 110 | max: 110 | pstdev: 0

Blocking. Results are skewed when significant numbers are reported here. (9 samples)
  min: 0s 000ms 088us | mean: 0s 000ms 567us | max: 0s 001ms 563us | pstdev: 0s 000ms 563us

  Percentile  Count       Value          
  0.5         5           0s 000ms 257us 
  0.75        7           0s 001ms 061us 
  0.8         8           0s 001ms 402us 
  0.9         9           0s 001ms 563us 

Initiation to completion (359998 samples)
  min: 0s 000ms 006us | mean: 0s 000ms 570us | max: 0s 062ms 482us | pstdev: 0s 000ms 661us

  Percentile  Count       Value          
  0.5         180035      0s 000ms 479us 
  0.75        270019      0s 000ms 558us 
  0.8         288010      0s 000ms 580us 
  0.9         324007      0s 000ms 725us 
  0.95        341999      0s 000ms 829us 
  0.990625    356624      0s 002ms 038us 
  0.99902344  359647      0s 008ms 940us 

Counter                                 Value       Per second
benchmark.http_2xx                      359744      3997.14
benchmark.pool_overflow                 254         2.82
cluster_manager.cluster_added           4           0.04
default.total_match_count               4           0.04
membership_change                       4           0.04
runtime.load_success                    1           0.01
runtime.override_dir_not_exists         1           0.01
upstream_cx_http1_total                 74          0.82
upstream_cx_rx_bytes_total              56479808    627551.51
upstream_cx_total                       74          0.82
upstream_cx_tx_bytes_total              15469078    171878.12
upstream_rq_pending_overflow            254         2.82
upstream_rq_pending_total               74          0.82
upstream_rq_total                       359746      3997.17

[09:49:49.788004][23][I] Wait for the connection pool drain timed out, proceeding to hard shutdown.
[09:49:49.789745][1][I] Done.

```

</details>

### Metrics

|Benchmark Name           |Envoy Gateway Memory (MiB)|Envoy Gateway Total CPU (Seconds)|
|-                        |-                         |-                                |
|scale-up-httproutes-10   |77                        |0.35                             |
|scale-up-httproutes-50   |90                        |1.96                             |
|scale-up-httproutes-100  |161                       |9.97                             |
|scale-up-httproutes-300  |883                       |185.43                           |
|scale-up-httproutes-500  |1473                      |10.42                            |
|scale-down-httproutes-300|697                       |5.73                             |
|scale-down-httproutes-100|463                       |103.44                           |
|scale-down-httproutes-50 |112                       |161.11                           |
|scale-down-httproutes-10 |116                       |162.41                           |
