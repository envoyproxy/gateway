---
title: "GRPC Timeouts"
---

Unlike [HTTPRoute][], the Gateway API [GRPCRoute][] resource does not (yet) expose a native
`timeouts` field — see the upstream tracking issue [Support Request Timeouts for GRPCRoute][gapi-3139].
Until that lands, Envoy Gateway lets you configure timeouts for gRPC traffic with a
[BackendTrafficPolicy][] that targets the `GRPCRoute`.

The default request timeout is 15 seconds in Envoy Proxy, which will terminate long-lived
streaming RPCs. The relevant `spec.timeout.http` fields are:

- **requestTimeout**: the maximum duration for the entire response to be received from the
  upstream. This bounds **unary** RPCs. Set it to `"0s"` to disable it for **streaming** RPCs,
  which otherwise would be cut off once the timeout elapses.
- **maxStreamDuration**: the maximum duration of a stream, measured from when the request is sent
  until the response stream is fully consumed. It does not apply to non-streaming requests. Set it
  to `"0s"` to allow streams to run indefinitely.
- **streamIdleTimeout**: the amount of time a stream may exist with no upstream or downstream
  activity. Use this to reclaim idle streams without capping a healthy long-lived stream's total
  duration.

## Prerequisites

{{< boilerplate prerequisites >}}

Follow the [GRPC Routing](../grpc-routing) task to set up a `Gateway` and a `GRPCRoute` named
`yages` before configuring timeouts.

__Note:__ A `GRPCRoute` can have at most one `BackendTrafficPolicy` attached to it; a second policy
targeting the same route is rejected as `Conflicted`. The two examples below are therefore
alternatives that reuse the same policy name (`grpc-timeouts`) — pick the one that matches your
workload. Re-applying with the same `metadata.name` updates the existing policy rather than creating
a conflicting second one.

## Unary RPCs

Set `requestTimeout` to bound the duration of unary RPCs. Here, unary calls that take longer than
5 seconds are terminated with a timeout.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: grpc-timeouts
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: GRPCRoute
    name: yages
  timeout:
    http:
      requestTimeout: "5s"
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
  name: grpc-timeouts
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: GRPCRoute
    name: yages
  timeout:
    http:
      requestTimeout: "5s"
```

{{% /tab %}}
{{< /tabpane >}}

## Streaming RPCs

For server-streaming, client-streaming, or bidirectional-streaming RPCs, `requestTimeout` would
terminate the stream once it elapses. Disable it with `"0s"` and, if you want an upper bound, use
`maxStreamDuration` (or `streamIdleTimeout` to reclaim only idle streams). The example below lets
streams run indefinitely while reclaiming streams that are idle for more than 1 hour.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: grpc-timeouts
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: GRPCRoute
    name: yages
  timeout:
    http:
      # Disable the per-request timeout so long-lived streams are not cut off.
      requestTimeout: "0s"
      # Allow streams to run indefinitely; set a non-zero value to cap them.
      maxStreamDuration: "0s"
      # Reclaim streams with no activity for more than 1 hour.
      streamIdleTimeout: "1h"
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
  name: grpc-timeouts
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: GRPCRoute
    name: yages
  timeout:
    http:
      # Disable the per-request timeout so long-lived streams are not cut off.
      requestTimeout: "0s"
      # Allow streams to run indefinitely; set a non-zero value to cap them.
      maxStreamDuration: "0s"
      # Reclaim streams with no activity for more than 1 hour.
      streamIdleTimeout: "1h"
```

{{% /tab %}}
{{< /tabpane >}}

## Verification

Confirm the policy is accepted:

```shell
kubectl get backendtrafficpolicy/grpc-timeouts -o yaml
```

The status should reflect `Accepted=True` on the targeted `GRPCRoute` ancestor.

[HTTPRoute]: https://gateway-api.sigs.k8s.io/api-types/httproute/
[GRPCRoute]: https://gateway-api.sigs.k8s.io/api-types/grpcroute/
[BackendTrafficPolicy]: ../../../api/extension_types#backendtrafficpolicy
[gapi-3139]: https://github.com/kubernetes-sigs/gateway-api/issues/3139
