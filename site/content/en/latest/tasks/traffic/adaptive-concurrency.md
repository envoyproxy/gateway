---
title: "Adaptive Concurrency"
---

[Envoy Adaptive Concurrency][] dynamically limits in-flight requests based on
sampled latencies. Envoy measures a baseline round-trip time, then lowers the
concurrency limit when latency rises so overloaded backends can recover.

Envoy Gateway exposes this through [BackendTrafficPolicy][]. Attach the policy
to a [Gateway][], [HTTPRoute][], or [GRPCRoute][]. Each policy gets its own
filter instance so concurrency state is not shared across unrelated routes.

The initial API is intentionally small. `sampleAggregatePercentile`,
`minConcurrencyLimit`, `maxConcurrencyLimit`, and
`concurrencyUpdateInterval` are the knobs operators typically need to tune.
Advanced minRTT sampling settings use Envoy defaults.

## Prerequisites

{{< boilerplate prerequisites >}}

## Configuration

An empty `adaptiveConcurrency` object enables the filter with Envoy defaults
(p50 latency, min concurrency 3, max concurrency 1000, update interval 100ms,
exceeded requests return 503).

The following policy enables adaptive concurrency on the example HTTPRoute and
overrides percentile and the concurrency bounds:

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: adaptive-concurrency-for-route
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: backend
  adaptiveConcurrency:
    sampleAggregatePercentile: 50
    minConcurrencyLimit: 3
    maxConcurrencyLimit: 1000
    concurrencyUpdateInterval: 100ms
    concurrencyLimitExceededStatus: 503
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: adaptive-concurrency-for-route
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: backend
  adaptiveConcurrency:
    sampleAggregatePercentile: 50
    minConcurrencyLimit: 3
    maxConcurrencyLimit: 1000
    concurrencyUpdateInterval: 100ms
    concurrencyLimitExceededStatus: 503
```

{{% /tab %}}
{{< /tabpane >}}

When sampled latency rises, Envoy reduces the concurrency limit between
`minConcurrencyLimit` and `maxConcurrencyLimit`. Requests that exceed the
limit receive `concurrencyLimitExceededStatus` (503 by default) instead of
being forwarded. Raise `minConcurrencyLimit` when a backend needs a higher
in-flight floor so load-based autoscaling can still observe enough traffic.

## Next Steps

- [Circuit Breakers](../circuit-breaker)
- [Admission Control](../../../api/extension_types#admissioncontrol)
- [BackendTrafficPolicy API](../../../api/extension_types#backendtrafficpolicy)

[Envoy Adaptive Concurrency]: https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/adaptive_concurrency_filter
[BackendTrafficPolicy]: ../../../api/extension_types#backendtrafficpolicy
[Gateway]: https://gateway-api.sigs.k8s.io/reference/api-types/gateway/
[HTTPRoute]: https://gateway-api.sigs.k8s.io/reference/api-types/httproute/
[GRPCRoute]: https://gateway-api.sigs.k8s.io/reference/api-types/grpcroute/
