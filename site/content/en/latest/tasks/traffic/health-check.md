---
title: "Health Checking"
---

Envoy Gateway can check the health of the endpoints behind a backend, so that Envoy only sends requests to endpoints
that are able to serve them. Two kinds of health checking are supported, and they can be used together:

- **Active health checking**: Envoy periodically sends a probe (HTTP, TCP or gRPC) to every endpoint and marks an
  endpoint unhealthy once enough consecutive probes fail. See [Envoy Active Health Checking][].
- **Passive health checking**: Envoy watches the responses of real requests and temporarily ejects endpoints that
  return too many errors. This is Envoy's [Outlier Detection][].

Health checking is configured in the `healthCheck` field of a [BackendTrafficPolicy][], which can target a [Gateway][],
an [HTTPRoute][], a [GRPCRoute][] or the other xRoute kinds. Each backendRef of a route has its own set of endpoints,
and each endpoint is checked independently.

## Prerequisites

{{< boilerplate prerequisites >}}

## Active Health Checking

Apply a `BackendTrafficPolicy` that probes every endpoint of the `backend` HTTPRoute with an HTTP `GET /` request every
3 seconds. An endpoint is marked unhealthy after 3 probes in a row time out or fail to connect, and healthy again after
1 successful probe. A response with an unexpected status marks the endpoint unhealthy right away, unless that status is
listed in `retriableStatuses`, in which case it counts toward `unhealthyThreshold` like a timeout.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: health-check
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: backend
  healthCheck:
    active:
      type: HTTP
      timeout: 1s
      interval: 3s
      unhealthyThreshold: 3
      healthyThreshold: 1
      http:
        path: /
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
  name: health-check
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: backend
  healthCheck:
    active:
      type: HTTP
      timeout: 1s
      interval: 3s
      unhealthyThreshold: 3
      healthyThreshold: 1
      http:
        path: /
```

{{% /tab %}}
{{< /tabpane >}}

Check that the policy was accepted:

```shell
kubectl get backendtrafficpolicy/health-check -o yaml
```

The health of each endpoint is visible through the Envoy admin interface:

```shell
ENVOY_POD=$(kubectl get pods -n envoy-gateway-system -l gateway.envoyproxy.io/owning-gateway-name=eg -o jsonpath='{.items[0].metadata.name}')
kubectl -n envoy-gateway-system port-forward pod/${ENVOY_POD} 19000:19000 &
curl -s localhost:19000/clusters | grep "httproute/default/backend" | grep health_flags
```

The quickstart backend answers `GET /` with a `200`, so the endpoint is reported as healthy:

```console
httproute/default/backend/rule/0::10.244.0.11:3000::health_flags::healthy
```

### Simulate a failing endpoint

By default, only a `200` response counts as healthy. To see what happens when an endpoint fails its health check,
expect a `204` instead, which the quickstart backend never returns:

```shell
kubectl patch backendtrafficpolicy/health-check --type merge \
  -p '{"spec":{"healthCheck":{"active":{"http":{"expectedStatuses":[204]}}}}}'
```

The next probe gets a `200`, which is neither expected nor retriable, so the endpoint is marked as failing the active
health check right away:

```console
httproute/default/backend/rule/0::10.244.0.11:3000::health_flags::/failed_active_hc
```

Requests to the backend still succeed, because of Envoy's [panic threshold][Panic Threshold]: when fewer than 50% of a
backend's endpoints are healthy, Envoy ignores health status and balances across all of its endpoints, so that a
failing health check can't take down a whole backend at once. The quickstart backend has a single endpoint, so it is
always in panic mode while that endpoint is unhealthy.

If you followed the "Without LoadBalancer Support" prerequisite, run `export GATEWAY_HOST=localhost:8888` first.

```shell
curl --verbose --header "Host: www.example.com" http://$GATEWAY_HOST/
```

To stop sending requests to unhealthy endpoints even when all of them are unhealthy, set `panicThreshold` to `0`:

```shell
kubectl patch backendtrafficpolicy/health-check --type merge \
  -p '{"spec":{"healthCheck":{"panicThreshold":0}}}'
```

The same request now fails with `503 Service Unavailable` and the body `no healthy upstream`.

Restore the expected statuses and the default panic threshold before continuing:

```shell
kubectl patch backendtrafficpolicy/health-check --type json \
  -p '[{"op":"remove","path":"/spec/healthCheck/active/http/expectedStatuses"},{"op":"remove","path":"/spec/healthCheck/panicThreshold"}]'
```

### Other active health checkers

The `type` field selects the health checker, and the matching field configures it. These snippets replace the
`healthCheck` block of the policy above.

A TCP health checker that only checks that a connection can be opened, or, with `send` and `receive`, that the endpoint
answers a payload with the expected bytes:

```yaml
  healthCheck:
    active:
      type: TCP
      tcp:
        send:
          type: Text
          text: ping
        receive:
          type: Text
          text: pong
```

A gRPC health checker that uses the [gRPC health checking protocol][], optionally for a specific service:

```yaml
  healthCheck:
    active:
      type: GRPC
      grpc:
        service: helloworld.Greeter
```

The HTTP health checker can also send a different `method`, `hostname` or `requestBody`, accept a list of
`expectedStatuses`, and match the response body with `expectedResponse`. When the health check is served on a different
port than the application, set `overrides.port`:

```yaml
  healthCheck:
    active:
      type: HTTP
      http:
        path: /healthz
        expectedStatuses:
          - 200
          - 204
      overrides:
        port: 9090
```

See [ActiveHealthCheck][] for every field and its default. To log the result of each probe, see
[Health Check Logging](../../observability/proxy-health-check-log).

## Passive Health Checking

Passive health checking needs no probes: Envoy counts the errors returned by each endpoint while serving real requests.
Apply a policy that ejects an endpoint for 30 seconds after 5 consecutive `5xx` responses, checking every 3 seconds.
It never ejects more than half of the endpoints, but `alwaysEjectOneEndpoint` still allows one endpoint to be ejected
when that limit rounds down to zero, as it does for the single endpoint of the quickstart backend:

```shell
kubectl patch backendtrafficpolicy/health-check --type merge \
  -p '{"spec":{"healthCheck":{"passive":{"consecutive5XxErrors":5,"interval":"3s","baseEjectionTime":"30s","maxEjectionPercent":50,"alwaysEjectOneEndpoint":true}}}}'
```

An ejected endpoint is reported with the `failed_outlier_check` health flag. It stays ejected for `baseEjectionTime`
multiplied by the number of times it has been ejected, so an endpoint that keeps failing is ejected for longer. See [PassiveHealthCheck][] for every field and its default.

Active and passive health checking can be combined in the same policy: an endpoint is only considered healthy when it
passes both. Panic mode still applies, so when fewer than `panicThreshold` percent of the endpoints are healthy, Envoy
sends requests to all of them regardless.

## Clean-Up

```shell
kubectl delete backendtrafficpolicy/health-check
```

[Envoy Active Health Checking]: https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/upstream/health_checking
[Outlier Detection]: https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/upstream/outlier
[Panic Threshold]: https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/upstream/load_balancing/panic_threshold
[gRPC health checking protocol]: https://github.com/grpc/grpc/blob/master/doc/health-checking.md
[BackendTrafficPolicy]: ../../../api/extension_types#backendtrafficpolicy
[ActiveHealthCheck]: ../../../api/extension_types#activehealthcheck
[PassiveHealthCheck]: ../../../api/extension_types#passivehealthcheck
[Gateway]: https://gateway-api.sigs.k8s.io/reference/api-types/gateway/
[HTTPRoute]: https://gateway-api.sigs.k8s.io/reference/api-types/httproute/
[GRPCRoute]: https://gateway-api.sigs.k8s.io/reference/api-types/grpcroute/
