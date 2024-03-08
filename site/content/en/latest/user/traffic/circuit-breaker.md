---
title: "Circuit Breakers"
---

[Envoy circuit breakers] can be used to fail quickly and apply back-pressure in response to upstream service degradation. 

Envoy Gateway supports the following circuit breaker thresholds:
- **Concurrent Connections**: limit the connections that Envoy can establish to the upstream service. When this threshold is met, new connections will not be established, and some requests will be queued until an existing connection becomes available. 
- **Concurrent Requests**: limit on concurrent requests in-flight from Envoy to the upstream service. When this threshold is met, requests will be queued.
- **Pending Requests**: limit the pending request queue size. When this threshold is met, overflowing requests will be terminated with a `503` status code. 

Envoy's circuit breakers are distributed: counters are not synchronized across different Envoy processes. The default Envoy and Envoy Gateway circuit breaker threshold values (1024) may be too strict for high-throughput systems.

Envoy Gateway introduces a new CRD called [BackendTrafficPolicy][] that allows the user to describe their desired circuit breaker thresholds.
This instantiated resource can be linked to a [Gateway][], [HTTPRoute][] or [GRPCRoute][] resource.

**Note**: There are distinct circuit breaker counters for each `BackendReference` in an `xRoute` rule. Even if a `BackendTrafficPolicy` targets a `Gateway`, each `BackendReference` in that gateway still has separate circuit breaker counter.

## Prerequisites

### Install Envoy Gateway

* Follow the installation step from the [Quickstart Guide](../../quickstart) to install Envoy Gateway and sample resources.

### Install the hey load testing tool
* The `hey` CLI will be used to generate load and measure response times. Follow the installation instruction from the [Hey project] docs.   

## Test and customize circuit breaker settings

This example will simulate a degraded backend that responds within 10 seconds by adding the `?delay=10s` query parameter to API calls. The hey tool will be used to generate 100 concurrent requests. 

```shell
hey -n 100 -c 100 -host "www.example.com"  http://${GATEWAY_HOST}/?delay=10s
```

```console
Summary:
  Total:	10.3426 secs
  Slowest:	10.3420 secs
  Fastest:	10.0664 secs
  Average:	10.2145 secs
  Requests/sec:	9.6687

  Total data:	36600 bytes
  Size/request:	366 bytes

Response time histogram:
  10.066 [1]	|■■■■
  10.094 [4]	|■■■■■■■■■■■■■■■
  10.122 [9]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  10.149 [10]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  10.177 [10]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  10.204 [11]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  10.232 [11]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  10.259 [11]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  10.287 [11]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  10.314 [11]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  10.342 [11]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
```

The default circuit breaker threshold (1024) is not met. As a result, requests do not overflow: all requests are proxied upstream and both Envoy and clients wait for 10s.

In order to fail fast, apply a `BackendTrafficPolicy` that limits concurrent requests to 10 and pending requests to 0.  

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: circuitbreaker-for-route
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: backend
    namespace: default
  circuitBreaker:
    maxPendingRequests: 0
    maxParallelRequests: 10
EOF
```

Execute the load simulation again.  

```shell
hey -n 100 -c 100 -host "www.example.com"  http://${GATEWAY_HOST}/?delay=10s
```

```console
Summary:
  Total:	10.1230 secs
  Slowest:	10.1224 secs
  Fastest:	0.0529 secs
  Average:	1.0677 secs
  Requests/sec:	9.8785

  Total data:	10940 bytes
  Size/request:	109 bytes

Response time histogram:
  0.053 [1]	|
  1.060 [89]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  2.067 [0]	|
  3.074 [0]	|
  4.081 [0]	|
  5.088 [0]	|
  6.095 [0]	|
  7.102 [0]	|
  8.109 [0]	|
  9.115 [0]	|
  10.122 [10]	|■■■■
```

With the new circuit breaker settings, and due to the slowness of the backend, only the first 10 concurrent requests were proxied, while the other 90 overflowed.   
* Overflowing Requests failed fast, reducing proxy resource consumption. 
* Upstream traffic was limited, alleviating the pressure on the degraded service. 

[Envoy Circuit Breakers]: https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/upstream/circuit_breaking
[BackendTrafficPolicy]: ../../../api/extension_types#backendtrafficpolicy
[Gateway]: https://gateway-api.sigs.k8s.io/api-types/gateway/
[HTTPRoute]: https://gateway-api.sigs.k8s.io/api-types/httproute/
[GRPCRoute]: https://gateway-api.sigs.k8s.io/api-types/grpcroute/
[Hey project]: https://github.com/rakyll/hey