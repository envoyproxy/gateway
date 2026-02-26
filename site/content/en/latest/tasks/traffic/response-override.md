---
title: "Response Override"
---

Response Override allows you to override the response from the backend with a custom one. This can be useful for scenarios such as returning a custom 404 page when the requested resource is not found, a custom 500 error message when the backend is failing, or redirecting the client when the backend returns a 403 Forbidden. When using redirect, Envoy applies it internally: Envoy follows the redirect to the new URL, obtains the response from that URL, and sends that response to the client.

## Installation

Follow the steps from the [Quickstart](../../quickstart) to install Envoy Gateway and the example manifest.
Before proceeding, you should be able to query the example backend using HTTP.

## Prerequisites

{{< boilerplate prerequisites >}}

## Testing Response Override

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: response-override
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: backend
  responseOverride:
    - match:
        statusCodes:
          - type: Value
            value: 404
      response:
        contentType: text/plain
        body:
          type: Inline
          inline: "Oops! Your request is not found."
    - match:
        statusCodes:
          - type: Value
            value: 403 # status from backend when envoy will execute the redirect
      redirect:
        statusCode: 302 # status envoy should respond to client
        scheme: https
        hostname: www.example.com
        path:
          type: ReplaceFullPath
          replaceFullPath: "/get"
    - match:
        statusCodes:
          - type: Value
            value: 500
          - type: Range
            range:
              start: 501
              end: 511
      response:
        contentType: application/json
        body:
          type: ValueRef
          valueRef:
            group: ""
            kind: ConfigMap
            name: response-override-config
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: response-override-config
data:
  response.body: '{"error": "Internal Server Error"}'
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
  name: response-override
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: backend
  responseOverride:
    - match:
        statusCodes:
          - type: Value
            value: 404
      response:
        contentType: text/plain
        body:
          type: Inline
          inline: "Oops! Your request is not found."
    - match:
        statusCodes:
          - type: Value
            value: 403 # status from backend when envoy will execute the redirect
      redirect:
        statusCode: 302 # status envoy should respond to client
        scheme: https
        hostname: www.example.com
        path:
          type: ReplaceFullPath
          replaceFullPath: "/get"
    - match:
        statusCodes:
          - type: Value
            value: 500
          - type: Range
            range:
              start: 501
              end: 511
      response:
        contentType: application/json
        body:
          type: ValueRef
          valueRef:
            group: ""
            kind: ConfigMap
            name: response-override-config
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: response-override-config
data:
  response.body: '{"error": "Internal Server Error"}'
```

{{% /tab %}}
{{< /tabpane >}}

```shell
curl --verbose --header "Host: www.example.com" http://$GATEWAY_HOST/status/404
```

```console
*   Trying 127.0.0.1:80...
* Connected to 172.18.0.200 (172.18.0.200) port 80
> GET /status/404 HTTP/1.1
> Host: www.example.com
> User-Agent: curl/8.5.0
> Accept: */*
>
< HTTP/1.1 404 Not Found
< content-type: text/plain
< content-length: 32
< date: Thu, 07 Nov 2024 09:22:29 GMT
<
* Connection #0 to host 172.18.0.200 left intact
Oops! Your request is not found.
```

For 403, the policy redirects to `https://www.example.com/get`. Envoy follows the redirect internally and sends that response to the client:

```shell
curl -L -v --verbose --header "Host: www.example.com" http://$GATEWAY_HOST/status/403
```

```console
* Host localhost:5000 was resolved.
* IPv6: ::1
* IPv4: 127.0.0.1
*   Trying [::1]:5000...
* Connected to localhost (::1) port 5000
> GET /status/403 HTTP/1.1
> Host: www.example.com
> User-Agent: curl/8.7.1
> Accept: */*
>
* Request completely sent off
< HTTP/1.1 302 Found
< content-type: application/json
< x-content-type-options: nosniff
< date: Thu, 26 Feb 2026 14:36:01 GMT
< content-length: 467
<
{
 "path": "/get",
 "host": "www.example.com",
 "method": "GET",
 "proto": "HTTP/1.1",
 "headers": {
  "Accept": [
   "*/*"
  ],
  "User-Agent": [
   "curl/8.7.1"
  ],
  "X-Envoy-External-Address": [
   "127.0.0.1"
  ],
  "X-Forwarded-For": [
   "10.244.2.2"
  ],
  "X-Forwarded-Proto": [
   "http"
  ],
  "X-Request-Id": [
   "d69f627e-c454-46b2-86e1-25d4c18b68e4"
  ]
 },
 "namespace": "default",
 "ingress": "",
 "service": "",
 "pod": "backend-869c8646c5-xfm84"
* Connection #0 to host localhost left intact
}
```

You receive the response body from the redirect target. Then verify the 500 override:

```shell
curl --verbose --header "Host: www.example.com" http://$GATEWAY_HOST/status/500
```

```console
*   Trying 127.0.0.1:80...
* Connected to 172.18.0.200 (172.18.0.200) port 80
> GET /status/500 HTTP/1.1
> Host: www.example.com
> User-Agent: curl/8.5.0
> Accept: */*
>
< HTTP/1.1 500 Internal Server Error
< content-type: application/json
< content-length: 34
< date: Thu, 07 Nov 2024 09:23:02 GMT
<
* Connection #0 to host 172.18.0.200 left intact
{"error": "Internal Server Error"}
```