---
title: Benchmark
---

Envoy Gateway uses [nighthawk][nighthawk] for benchmarking, and mainly concerned with 
its performance and scalability as a control-plane.

The performance and scalability concerns come from several aspects for control-plane:

- The consumption of memory and CPU.
- The rate of configuration changes.

## Run Benchmark Test

The benchmark test is running on [Kind][kind] cluster, you can run the following command
to start a Kind cluster and run benchmark test on it.

```shell
make benchmark-test
```

By default, a benchmark report will be generated under `test/benchmark` after test finished.


[nighthawk]: https://github.com/envoyproxy/nighthawk
[kind]: https://kind.sigs.k8s.io/
