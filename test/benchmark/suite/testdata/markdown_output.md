# Benchmark Report

Benchmark test settings:

| RPS      | Connections     | Duration (Seconds) | CPU Limits (m) | Memory Limits (MiB) |
|----------|-----------------|--------------------|----------------|---------------------|
| 1000 | 100 | 30      | 1000m       | 1000Mi         |

## Test: ScaleHTTPRoute

Fixed one Gateway and different scales of HTTPRoutes with different portion of hostnames.


### Results

Expand to see the full results.
<details>
<summary>case-title</summary>

```plaintext

benchmark_http_client.latency_2xx (489025 samples)
  min: 135.832µs | mean: 1.206029ms | max: 473.694207ms | pstdev: 6.470055ms

  Percentile	Count		Value
  0.500000		244542		438.127µs
  0.750000		366785		546.047µs
  0.800000		391239		584.511µs
  0.900000		440125		751.775µs
  0.950000		464574		1.100415ms
  0.990625		484444		55.601151ms
  0.999023		488548		60.461055ms

Queueing and connection setup latency (489026 samples)
  min: 2.25µs | mean: 10.746µs | max: 181.919743ms | pstdev: 822.232µs

  Percentile	Count		Value
  0.500000		245388		5.666µs
  0.750000		367127		6.584µs
  0.800000		392841		6.833µs
  0.900000		440995		7.542µs
  0.950000		464743		8.958µs
  0.990625		484447		37.459µs
  0.999023		488549		206.335µs

Request start to response end (489025 samples)
  min: 135.16µs | mean: 1.205247ms | max: 473.694207ms | pstdev: 6.469448ms

  Percentile	Count		Value
  0.500000		244566		437.679µs
  0.750000		366783		545.567µs
  0.800000		391232		583.935µs
  0.900000		440124		751.071µs
  0.950000		464575		1.099455ms
  0.990625		484442		55.599103ms
  0.999023		488548		60.461055ms

Blocking. Results are skewed when significant numbers are reported here. (489026 samples)
  min: 59.834µs | mean: 1.224014ms | max: 565.084159ms | pstdev: 6.619735ms

  Percentile	Count		Value
  0.500000		244568		452.175µs
  0.750000		366794		561.471µs
  0.800000		391224		600.543µs
  0.900000		440130		772.031µs
  0.950000		464576		1.126783ms
  0.990625		484443		55.615487ms
  0.999023		488549		60.481535ms

Initiation to completion (490015 samples)
  min: 143µs | mean: 1.504828ms | max: 694.747135ms | pstdev: 9.302005ms

  Percentile	Count		Value
  0.500000		245054		448.639µs
  0.750000		367536		558.175µs
  0.800000		392012		597.631µs
  0.900000		441021		772.927µs
  0.950000		465515		1.149055ms
  0.990625		485424		57.298943ms
  0.999023		489541		141.213695ms



```

</details>

### Metrics

The CPU usage statistics of both control-plane and data-plane are the CPU usage per second over the past 30 seconds.

| Test Name | Envoy Gateway Container Memory (MiB) <br> min/max/means | Envoy Gateway Process Memory (MiB) <br> min/max/means | Envoy Gateway CPU (%) <br> min/max/means | Averaged Envoy Proxy Memory (MiB) <br> min/max/means | Averaged Envoy Proxy CPU (%) <br> min/max/means |
|-----------|---------------------------------------------------------|---------------------------------------------------------|------------------------------------------|------------------------------------------------------|-------------------------------------------------|
| case-title | 0.00 / 0.00 / 0.00 | 0.00 / 0.00 / 0.00 |  | 0.00 / 0.00 / 0.00 | 0.00 / 0.00 / 0.00 |

### Profiles

The profiles at different scales are stored under `/profiles` directory in report, with name `{ProfileType}.{TestCase}.pprof`.

You can visualize them in a web page by running:

```shell
go tool pprof -http=: path/to/your.pprof
```

Currently, the supported profile types are:
- heap (memory)


#### Heap

The profiles were sampled when Envoy Gateway Memory is at its maximum.
#### case-title.

![.Name](fake-image)
