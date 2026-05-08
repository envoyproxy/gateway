---
title: "Bandwidth Limit"
---

Bandwidth limit allows users to control the throughput of traffic sent to and received from backends.
Envoy throttles matching request or response bodies by pausing data transfer when the configured bandwidth is exhausted, resuming as tokens are replenished, so clients generally see slower transfers rather than bandwidth-limit rejections.

This feature is implemented using the [Envoy bandwidth limit filter][envoy-bandwidth-limit-filter].

Users may want to limit bandwidth for several reasons:

* Prevent a single backend or route from consuming all available network capacity.
* Protect upstream services from being overwhelmed by large request or response bodies.
* Enforce fair usage across multiple routes or tenants sharing the same gateway.

Envoy Gateway uses the [BackendTrafficPolicy][] CRD to express bandwidth limit settings.
This instantiated resource can be linked to a [Gateway][], [HTTPRoute][], or [GRPCRoute][].

**Note:** The bandwidth limit is applied per Envoy proxy instance.
If the data plane runs multiple replicas, each replica enforces the limit independently.

## Prerequisites

### Install Envoy Gateway

{{< boilerplate prerequisites >}}

## Limit Request Bandwidth

The following example limits inbound request traffic for an HTTPRoute to **10 KiB/s**.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: bandwidth-limit-route
spec:
  parentRefs:
  - name: eg
  hostnames:
  - "www.example.com"
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /
    backendRefs:
    - group: ""
      kind: Service
      name: backend
      port: 3000
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: bandwidth-limit-policy
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: bandwidth-limit-route
  bandwidthLimit:
    request:
      limit:
        value: "10Ki"
        unit: Second
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resources to your cluster:

```yaml
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: bandwidth-limit-route
spec:
  parentRefs:
  - name: eg
  hostnames:
  - "www.example.com"
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /
    backendRefs:
    - group: ""
      kind: Service
      name: backend
      port: 3000
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: bandwidth-limit-policy
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: bandwidth-limit-route
  bandwidthLimit:
    request:
      limit:
        value: "10Ki"
        unit: Second
```

{{% /tab %}}
{{< /tabpane >}}

Verify that the policy was accepted:

```shell
kubectl get backendtrafficpolicy bandwidth-limit-policy -o yaml
```

Upload a large file to trigger request throttling:

```shell
dd if=/dev/zero of=/tmp/test.bin bs=1K count=50

export GATEWAY_HOST=$(kubectl get gateway/eg -o jsonpath='{.status.addresses[0].value}')

curl -X POST \
  -H "Host: www.example.com" \
  --data-binary @/tmp/test.bin \
  http://${GATEWAY_HOST}/
```

Verify throttling occurred by checking Envoy's Prometheus metrics:

```shell
egctl experimental stats envoy-proxy \
  -n envoy-gateway-system \
  -l gateway.envoyproxy.io/owning-gateway-name=eg,gateway.envoyproxy.io/owning-gateway-namespace=default \
  | grep bandwidth_limit
```

Key metrics to check:

| Metric | Description |
|--------|-------------|
| `envoy_http_bandwidth_limiter_http_bandwidth_limit_request_enabled` | Total number of request streams for which the bandwidth limiter was consulted |
| `envoy_http_bandwidth_limiter_http_bandwidth_limit_request_enforced` | Total number of request streams for which the bandwidth limiter was enforced |
| `envoy_http_bandwidth_limiter_http_bandwidth_limit_request_incoming_total_size` | Total size in bytes of incoming request data to bandwidth limiter |
| `envoy_http_bandwidth_limiter_http_bandwidth_limit_request_allowed_total_size` | Total size in bytes of outgoing request data from bandwidth limiter |

## Limit Response Bandwidth

The following example limits outbound response traffic to **1 KiB/s** and adds response trailers, so clients can observe the delay introduced by the filter.

Deploy [httpbin][] as a backend. Its `/bytes/{n}` endpoint returns exactly `n` bytes, making it easy to generate large responses on demand:

```shell
kubectl create deployment httpbin --image=kennethreitz/httpbin --port=80
kubectl expose deployment httpbin --port=80
kubectl wait --for=condition=Available deployment/httpbin --timeout=60s
```

Apply the HTTPRoute and BackendTrafficPolicy:

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: bandwidth-limit-response-route
spec:
  parentRefs:
  - name: eg
  hostnames:
  - "www.example.com"
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /bytes
    backendRefs:
    - group: ""
      kind: Service
      name: httpbin
      port: 80
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: bandwidth-limit-response-policy
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: bandwidth-limit-response-route
  bandwidthLimit:
    response:
      limit:
        value: "1Ki"
        unit: Second
      responseTrailers:
        prefix: "x-eg"
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}

Save and apply the following resources to your cluster:

```yaml
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: bandwidth-limit-response-route
spec:
  parentRefs:
  - name: eg
  hostnames:
  - "www.example.com"
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /bytes
    backendRefs:
    - group: ""
      kind: Service
      name: httpbin
      port: 80
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: bandwidth-limit-response-policy
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: bandwidth-limit-response-route
  bandwidthLimit:
    response:
      limit:
        value: "1Ki"
        unit: Second
      responseTrailers:
        prefix: "x-eg"
```

{{% /tab %}}
{{< /tabpane >}}

Download 10 KiB through the throttled route.

```shell
curl -v --http2-prior-knowledge -sS \
  -o /dev/null \
  -H "Host: www.example.com" \
  http://${GATEWAY_HOST}/bytes/10240
```

Look for lines starting with `< x-eg` in the verbose output:

```
< x-eg-bandwidth-response-delay-ms: 9047
< x-eg-bandwidth-response-filter-delay-ms: 8850
```

With `prefix: "x-eg"`, Envoy appends the following trailers to each throttled response:

| Trailer | Description |
|---------|-------------|
| `x-eg-bandwidth-request-delay-ms` | Total request-stream transfer delay in milliseconds, including body transfer time and filter-added delay. |
| `x-eg-bandwidth-response-delay-ms` | Total response-stream transfer delay in milliseconds, including body transfer time and filter-added delay. |
| `x-eg-bandwidth-request-filter-delay-ms` | Delay in milliseconds added to the request stream by the filter only. |
| `x-eg-bandwidth-response-filter-delay-ms` | Delay in milliseconds added to the response stream by the filter only. |

## Limit Both Directions

`request` and `response` can be combined in a single policy:

```yaml
bandwidthLimit:
  request:
    limit:
      value: "10Ki"
      unit: Second
  response:
    limit:
      value: "1Mi"
      unit: Second
```

At least one of `request` or `response` must be specified.

## Bandwidth Limit Values

The `value` field accepts [Kubernetes resource quantity][resource-quantity] notation.
The `unit` field controls the time window: `Second`, `Minute`, or `Hour`.

[envoy-bandwidth-limit-filter]: https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/bandwidth_limit_filter
[BackendTrafficPolicy]: ../../../api/extension_types#backendtrafficpolicy
[Gateway]: https://gateway-api.sigs.k8s.io/api-types/gateway/
[HTTPRoute]: https://gateway-api.sigs.k8s.io/api-types/httproute/
[GRPCRoute]: https://gateway-api.sigs.k8s.io/api-types/grpcroute/
[resource-quantity]: https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/quantity/
[httpbin]: https://httpbin.org
