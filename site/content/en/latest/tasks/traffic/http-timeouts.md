---
title: "HTTP Timeouts"
---

The default request timeout is set to 15 seconds in Envoy Proxy.
The [HTTPRouteTimeouts][] resource allows users to configure request timeouts for an [HTTPRouteRule][].
This task shows you how to configure timeouts.

The [HTTPRouteTimeouts][] supports two kinds of timeouts:
- **request**: Request specifies the maximum duration for a gateway to respond to an HTTP request. 
- **backendRequest**: BackendRequest specifies a timeout for an individual request from the gateway to a backend.

__Note:__  The Request duration must be >= BackendRequest duration

## Prerequisites

{{< boilerplate prerequisites >}}

## Verification

backend has the ability to delay responses; we use it as the backend to control response time.

### request timeout
We configure the backend to delay responses by 3 seconds, then we set the request timeout to 4 seconds. Envoy Gateway will successfully respond to the request.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: backend
spec:
  hostnames:
  - timeout.example.com
  parentRefs:
  - group: gateway.networking.k8s.io
    kind: Gateway
    name: eg
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
        value: /
    timeouts:
      request: "4s"
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: backend
spec:
  hostnames:
  - timeout.example.com
  parentRefs:
  - group: gateway.networking.k8s.io
    kind: Gateway
    name: eg
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
        value: /
    timeouts:
      request: "4s"
```

{{% /tab %}}
{{< /tabpane >}}

```shell
curl --header "Host: timeout.example.com" http://${GATEWAY_HOST}/?delay=3s  -I
```

```console
HTTP/1.1 200 OK
content-type: application/json
x-content-type-options: nosniff
date: Mon, 04 Mar 2024 02:34:21 GMT
content-length: 480
```

Then we set the request timeout to 2 seconds. In this case, Envoy Gateway will respond with a timeout.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: backend
spec:
  hostnames:
  - timeout.example.com
  parentRefs:
  - group: gateway.networking.k8s.io
    kind: Gateway
    name: eg
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
        value: /
    timeouts:
      request: "2s"
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: backend
spec:
  hostnames:
  - timeout.example.com
  parentRefs:
  - group: gateway.networking.k8s.io
    kind: Gateway
    name: eg
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
        value: /
    timeouts:
      request: "2s"
```

{{% /tab %}}
{{< /tabpane >}}

```shell
curl --header "Host: timeout.example.com" http://${GATEWAY_HOST}/?delay=3s  -v
```

```console
*   Trying 127.0.0.1:80...
* Connected to 127.0.0.1 (127.0.0.1) port 80
> GET /?delay=3s HTTP/1.1
> Host: timeout.example.com
> User-Agent: curl/8.6.0
> Accept: */*
>


< HTTP/1.1 504 Gateway Timeout
< content-length: 24
< content-type: text/plain
< date: Mon, 04 Mar 2024 02:35:03 GMT
<
* Connection #0 to host 127.0.0.1 left intact
upstream request timeout
```

[HTTPRouteTimeouts]: https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.HTTPRouteTimeouts
[HTTPRouteRule]: https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.HTTPRouteRule
