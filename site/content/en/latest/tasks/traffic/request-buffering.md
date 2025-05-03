---
title: "Request Buffering"
---

The [Envoy buffer filter] is used to stop filter iteration and wait for a fully buffered complete request. This is useful in different situations including protecting some applications from having to deal with partial requests and high network latency.

Enabling request buffering requires specifying a size limit for the buffer. Any requests that are larger than the limit will stop the buffering and return a HTTP 413 Content Too Large response. 

Envoy Gateway introduces a new CRD called [BackendTrafficPolicy][] that allows the user to enable request buffering.
This instantiated resource can be linked to a [Gateway][], or [HTTPRoute][].

If the target of the BackendTrafficPolicy is a Gateway, the request buffering will be applied to all xRoutes under that Gateway.

## Prerequisites

{{< boilerplate prerequisites >}}

## Configuration

Enable request buffering by creating an [BackendTrafficPolicy][BackendTrafficPolicy] and attaching it to the example HTTPRoute.

### HTTPRoute

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: request-buffer
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: foo
  requestBuffer:
    limit: 4Ki
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
  name: request-buffer
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: foo
  requestBuffer:
    limit: 4 # Supports SI units e.g. 4Ki, 1Mi
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

A HTTPRoute resource is created for the `/foo` path prefix. The `request-buffer` BackendTrafficPolicy has been created and targeted HTTPRoute foo to enable request buffering. A small buffer limit of `4` bytes is purposely chosen to make testing easier.

Verify the HTTPRoute configuration and status:

```shell
kubectl get httproute/foo -o yaml
```

Verify the BackendTrafficPolicy configuration:

```shell
kubectl get backendtrafficpolicy/request-buffer -o yaml
```

## Testing

Ensure the `GATEWAY_HOST` environment variable from the [Quickstart](../../quickstart) is set. If not, follow the
Quickstart instructions to set the variable.

```shell
echo $GATEWAY_HOST
```

### HTTPRoute

We will try sending a request with an empty json object that is less than the buffer limit of 4 bytes

```shell
curl -H "Host: www.example.com" "http://${GATEWAY_HOST}/foo" -XPOST -d '{}'
```

We will see the following output. The `Content-Length` header will be added by the buffer filter.

```
{
 "path": "/foo",
 "host": "www.example.com",
 "method": "POST",
 "proto": "HTTP/1.1",
 "headers": {
  "Accept": [
   "*/*"
  ],
  "Content-Length": [
   "2"
  ],
  "Content-Type": [
   "application/x-www-form-urlencoded"
  ],
  "User-Agent": [
   "curl/8.7.1"
  ],
  "X-Envoy-External-Address": [
   "127.0.0.1"
  ],
  "X-Forwarded-For": [
   "10.244.0.2"
  ],
  "X-Forwarded-Proto": [
   "http"
  ],
  "X-Request-Id": [
   "daf7067e-a9e5-48da-86d2-6f5d9ccfb57e"
  ]
 },
 "namespace": "default",
 "ingress": "",
 "service": "",
 "pod": "backend-869c8646c5-9vm4l"
}
```

Next we will try sending a json object that is larger than 4 bytes. We will also write the status code to make it clear.

```shell
curl -H "Host: www.example.com" "http://${GATEWAY_HOST}/foo" -XPOST -d '{"key": "value"}' -w "\nStatus Code: %{http_code}"
```

We will now see that sending a payload of `{"key": "value"}` which is larger than the request buffer limit of 4 bytes returns a 
HTTP 413 Payload Too Large response

```
Payload Too Large
Status Code: 413
```

## Clean-Up

Follow the steps from the [Quickstart](../../quickstart) to uninstall Envoy Gateway and the example manifest.

Delete the BackendTrafficPolicy and HTTPRoute:

```shell
kubectl delete httproute/foo
kubectl delete backendtrafficpolicy/request-buffer
```

[Envoy buffer filter]: https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/buffer_filter
[BackendTrafficPolicy]: ../../../api/extension_types#backendtrafficpolicy
[Gateway]: https://gateway-api.sigs.k8s.io/api-types/gateway/
[HTTPRoute]: https://gateway-api.sigs.k8s.io/api-types/httproute/
