---
title: "Request Body Size Limit"
---

Envoy Gateway's [BackendTrafficPolicy][] CRD lets you cap the size of the request body on a route with the `requestBodyLimit` field.
This instantiated resource can be linked to a [Gateway][] or [HTTPRoute][].

If the target of the BackendTrafficPolicy is a Gateway, the configuration will be applied to all xRoutes under that Gateway.

Configuring `requestBodyLimit` requires specifying `maxSize`. Envoy enforces it with the [body size limit filter][] and rejects a request
as soon as it knows the body is too large:

- When the request carries a `Content-Length` header that announces a larger body, it is rejected with an HTTP 413 Content Too Large
  response before any body byte is read.
- Otherwise the body is streamed to the backend and the request is rejected once the bytes received pass the limit. The client gets a
  413 if no response headers have been sent to it yet, which is the case when the backend only responds after reading the whole body.
  If the response has already started, it can no longer be replaced, so Envoy resets the stream instead: it closes an HTTP/1.1
  connection and resets an HTTP/2 or HTTP/3 stream.

The body is never buffered. This is what sets `requestBodyLimit` apart from [request buffering][] with mode `BufferAndLimit`, which also
caps the body size but only by holding the whole request in memory first, and which disables HTTP upgrades on the route. A request body
limit costs no memory and leaves upgrades such as WebSocket enabled, so it is the right choice when all you need is a maximum request
body size, for example to replace an ingress controller's body size annotation.

Warning: the limit counts every request body byte of a stream. Envoy passes the client-to-server payload of an upgraded connection and
of a streaming request through the HTTP filter chain as request body data, so a long-lived WebSocket connection or gRPC stream whose
payload grows past `maxSize` is terminated as well. By then a WebSocket has always received its `101 Switching Protocols` response, and a
gRPC stream has often received its response headers, so the client sees a closed connection or a reset stream rather than a 413. Do not
set `requestBodyLimit` on routes that carry such traffic.

Note: `requestBodyLimit` requires Envoy proxy v1.40 or later, the first release that ships the body size limit filter.

## Prerequisites

{{< boilerplate prerequisites >}}

## Configuration

Cap the request body size by creating a [BackendTrafficPolicy][] and attaching it to the example HTTPRoute.

### HTTPRoute

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: request-body-limit
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: foo
  requestBodyLimit:
    maxSize: 4 # Supports SI units e.g. 4Ki, 1Mi
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: foo
spec:
  parentRefs:
  - name: eg
  hostnames:
  - "www.example.com"
  rules:
  - backendRefs:
    - group: ""
      kind: Service
      name: backend
      port: 3000
      weight: 1
    matches:
    - path:
        type: PathPrefix
        value: /foo
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resources to your cluster:

```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: request-body-limit
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: foo
  requestBodyLimit:
    maxSize: 4 # Supports SI units e.g. 4Ki, 1Mi
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: foo
spec:
  parentRefs:
  - name: eg
  hostnames:
  - "www.example.com"
  rules:
  - backendRefs:
    - group: ""
      kind: Service
      name: backend
      port: 3000
      weight: 1
    matches:
    - path:
        type: PathPrefix
        value: /foo
```

{{% /tab %}}
{{< /tabpane >}}

A HTTPRoute resource is created for the `/foo` path prefix. The `request-body-limit` BackendTrafficPolicy targets HTTPRoute foo and
caps its request body at 4 bytes. Such a small limit is purposely chosen to make testing easier.

Verify the HTTPRoute configuration and status:

```shell
kubectl get httproute/foo -o yaml
```

Verify the BackendTrafficPolicy configuration:

```shell
kubectl get backendtrafficpolicy/request-body-limit -o yaml
```

## Testing

Ensure the `GATEWAY_HOST` environment variable from the [Quickstart](../../quickstart) is set. If not, follow the
Quickstart instructions to set the variable.

```shell
echo $GATEWAY_HOST
```

### HTTPRoute

Send a request with an empty JSON object, which is smaller than the 4 byte limit:

```shell
curl -H "Host: www.example.com" "http://${GATEWAY_HOST}/foo" -XPOST -d '{}' -w "\nStatus Code: %{http_code}"
```

The request is forwarded to the backend, which echoes it back with a `200` status code.

Next, send a JSON object that is larger than 4 bytes. curl sets a `Content-Length` header, so Envoy rejects the request from the header
alone, before reading the body:

```shell
curl -H "Host: www.example.com" "http://${GATEWAY_HOST}/foo" -XPOST -d '{"key": "value"}' -w "\nStatus Code: %{http_code}"
```

```
content length too large
Status Code: 413
```

Finally, send the same body with chunked transfer encoding, so that no `Content-Length` header is sent. Envoy now streams the body and
rejects the request once the bytes received pass the limit:

```shell
curl -H "Host: www.example.com" -H "Transfer-Encoding: chunked" "http://${GATEWAY_HOST}/foo" -XPOST -d '{"key": "value"}' -w "\nStatus Code: %{http_code}"
```

```
request too long
Status Code: 413
```

## Choosing between requestBodyLimit and requestBuffer

Both fields reject a request whose body is larger than a configured size with HTTP 413, but they do so differently:

| | `requestBodyLimit` | `requestBuffer` with mode `BufferAndLimit` |
| --- | --- | --- |
| Body buffered in the gateway | No, it is streamed to the backend | Yes, the whole body is received before it is forwarded |
| Memory cost per in-flight request | None | Up to the limit |
| HTTP upgrades such as WebSocket | Stay enabled | Disabled on the route |
| Oversized `Content-Length` | Rejected before the body is read | Rejected before the body is read |
| Upper bound of the limit | 64-bit | 32-bit (`max_request_bytes` of the buffer filter) |

Use `requestBodyLimit` when you only need a maximum request body size. Use `requestBuffer` when the backend must also be shielded from
slow or partial requests. The two can be combined on a route, and the body size limit filter runs first:

- A request whose `Content-Length` exceeds `maxSize` is rejected before the buffer filter starts buffering it.
- A streamed body is counted as it arrives, and every chunk below the limit is passed on to the buffer filter. By the time the body
  passes `maxSize` and is rejected, the buffer filter already holds the bytes received so far. Each streamed request can therefore
  buffer up to the lower of `maxSize` and the `requestBuffer` limit.

## Combining with external authorization

The body size limit filter also runs before the external authorization filter. When a [SecurityPolicy][] on the same route sends the
request body to the authorization service with `extAuth.bodyToExtAuth` (see [External Authorization][]), a body larger than `maxSize`
is rejected with HTTP 413 before it is sent to the authorization service:

- An oversized `Content-Length` is rejected without calling the authorization service.
- A streamed body is rejected once it passes `maxSize`, so the external authorization filter never buffers more than `maxSize` bytes.

Because the size check comes first, a client that sends an oversized body receives a 413 even if the request would have failed
authorization.

## Clean-Up

Follow the steps from the [Quickstart](../../quickstart) to uninstall Envoy Gateway and the example manifest.

Delete the BackendTrafficPolicy and HTTPRoute:

```shell
kubectl delete httproute/foo
kubectl delete backendtrafficpolicy/request-body-limit
```

## Next Steps

Checkout the [Developer Guide](../../../contributions/develop) to get involved in the project.

[BackendTrafficPolicy]: ../../../api/extension_types#backendtrafficpolicy
[Gateway]: https://gateway-api.sigs.k8s.io/reference/api-types/gateway/
[HTTPRoute]: https://gateway-api.sigs.k8s.io/reference/api-types/httproute/
[body size limit filter]: https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/body_size_limit_filter
[request buffering]: ../request-buffering
[SecurityPolicy]: ../../../api/extension_types#securitypolicy
[External Authorization]: ../../security/ext-auth
