---
title: "HTTP URL Rewrite"
---

[HTTPURLRewriteFilter][] defines a filter that modifies a request during forwarding. At most one of these filters may be
used on a Route rule. This MUST NOT be used on the same Route rule as a HTTPRequestRedirect filter.

## Prerequisites

{{< boilerplate prerequisites >}}

## Rewrite URL Prefix Path

You can configure to rewrite the prefix in the url like below. In this example, any curls to
`http://${GATEWAY_HOST}/get/xxx` will be rewritten to `http://${GATEWAY_HOST}/replace/xxx`.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: http-filter-url-rewrite
spec:
  parentRefs:
    - name: eg
  hostnames:
    - path.rewrite.example
  rules:
    - matches:
      - path:
          value: "/get"
      filters:
      - type: URLRewrite
        urlRewrite:
          path:
            type: ReplacePrefixMatch
            replacePrefixMatch: /replace
      backendRefs:
      - name: backend
        port: 3000
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
  name: http-filter-url-rewrite
spec:
  parentRefs:
    - name: eg
  hostnames:
    - path.rewrite.example
  rules:
    - matches:
      - path:
          value: "/get"
      filters:
      - type: URLRewrite
        urlRewrite:
          path:
            type: ReplacePrefixMatch
            replacePrefixMatch: /replace
      backendRefs:
      - name: backend
        port: 3000
```

{{% /tab %}}
{{< /tabpane >}}

The HTTPRoute status should indicate that it has been accepted and is bound to the example Gateway.

```shell
kubectl get httproute/http-filter-url-rewrite -o yaml
```

Get the Gateway's address:

```shell
export GATEWAY_HOST=$(kubectl get gateway/eg -o jsonpath='{.status.addresses[0].value}')
```

Querying `http://${GATEWAY_HOST}/get/origin/path` should rewrite to
`http://${GATEWAY_HOST}/replace/origin/path`.

```console
$ curl -L -vvv --header "Host: path.rewrite.example" "http://${GATEWAY_HOST}/get/origin/path"
...
> GET /get/origin/path HTTP/1.1
> Host: path.rewrite.example
> User-Agent: curl/7.85.0
> Accept: */*
>

< HTTP/1.1 200 OK
< content-type: application/json
< x-content-type-options: nosniff
< date: Wed, 21 Dec 2022 11:03:28 GMT
< content-length: 503
< x-envoy-upstream-service-time: 0
< server: envoy
<
{
 "path": "/replace/origin/path",
 "host": "path.rewrite.example",
 "method": "GET",
 "proto": "HTTP/1.1",
 "headers": {
  "Accept": [
   "*/*"
  ],
  "User-Agent": [
   "curl/7.85.0"
  ],
  "X-Envoy-Expected-Rq-Timeout-Ms": [
   "15000"
  ],
  "X-Envoy-Original-Path": [
   "/get/origin/path"
  ],
  "X-Forwarded-Proto": [
   "http"
  ],
  "X-Request-Id": [
   "fd84b842-9937-4fb5-83c7-61470d854b90"
  ]
 },
 "namespace": "default",
 "ingress": "",
 "service": "",
 "pod": "backend-6fdd4b9bd8-8vlc5"
...
```

You can see that the `X-Envoy-Original-Path` is `/get/origin/path`, but the actual path is `/replace/origin/path`.

## Rewrite URL Full Path

You can configure to rewrite the fullpath in the url like below. In this example, any request sent to
`http://${GATEWAY_HOST}/get/origin/path/xxxx` will be rewritten to
`http://${GATEWAY_HOST}/force/replace/fullpath`.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: http-filter-url-rewrite
spec:
  parentRefs:
    - name: eg
  hostnames:
    - path.rewrite.example
  rules:
    - matches:
      - path:
          type: PathPrefix
          value: "/get/origin/path"
      filters:
      - type: URLRewrite
        urlRewrite:
          path:
            type: ReplaceFullPath
            replaceFullPath: /force/replace/fullpath
      backendRefs:
      - name: backend
        port: 3000
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
  name: http-filter-url-rewrite
spec:
  parentRefs:
    - name: eg
  hostnames:
    - path.rewrite.example
  rules:
    - matches:
      - path:
          type: PathPrefix
          value: "/get/origin/path"
      filters:
      - type: URLRewrite
        urlRewrite:
          path:
            type: ReplaceFullPath
            replaceFullPath: /force/replace/fullpath
      backendRefs:
      - name: backend
        port: 3000
```

{{% /tab %}}
{{< /tabpane >}}

The HTTPRoute status should indicate that it has been accepted and is bound to the example Gateway.

```shell
kubectl get httproute/http-filter-url-rewrite -o yaml
```

Querying `http://${GATEWAY_HOST}/get/origin/path/extra` should rewrite the request to
`http://${GATEWAY_HOST}/force/replace/fullpath`.

```console
$ curl -L -vvv --header "Host: path.rewrite.example" "http://${GATEWAY_HOST}/get/origin/path/extra"
...
> GET /get/origin/path/extra HTTP/1.1
> Host: path.rewrite.example
> User-Agent: curl/7.85.0
> Accept: */*
>
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< content-type: application/json
< x-content-type-options: nosniff
< date: Wed, 21 Dec 2022 11:09:31 GMT
< content-length: 512
< x-envoy-upstream-service-time: 0
< server: envoy
<
{
 "path": "/force/replace/fullpath",
 "host": "path.rewrite.example",
 "method": "GET",
 "proto": "HTTP/1.1",
 "headers": {
  "Accept": [
   "*/*"
  ],
  "User-Agent": [
   "curl/7.85.0"
  ],
  "X-Envoy-Expected-Rq-Timeout-Ms": [
   "15000"
  ],
  "X-Envoy-Original-Path": [
   "/get/origin/path/extra"
  ],
  "X-Forwarded-Proto": [
   "http"
  ],
  "X-Request-Id": [
   "8ab774d6-9ffa-4faa-abbb-f45b0db00895"
  ]
 },
 "namespace": "default",
 "ingress": "",
 "service": "",
 "pod": "backend-6fdd4b9bd8-8vlc5"
...
```

You can see that the `X-Envoy-Original-Path` is `/get/origin/path/extra`, but the actual path is
`/force/replace/fullpath`.

## Rewrite URL Path with Regex

In addition to core Gateway-API rewrite options, Envoy Gateway supports extended rewrite options through the [HTTPRouteFilter][] API.
The `HTTPRouteFilter` API can be configured to use [RE2][]-compatible regex matchers and substitutions to rewrite a portion of the url.
In the example below, requests sent to `http://${GATEWAY_HOST}/service/xxx/yyy` (where `xxx` is a single path portion and `yyy` is one or more path portions)
are rewritten to `http://${GATEWAY_HOST}/yyy/instance/xxx`. The entire path is matched and rewritten using capture groups.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: http-filter-url-regex-rewrite
spec:
  parentRefs:
    - name: eg
  hostnames:
    - path.regex.rewrite.example
  rules:
    - matches:
      - path:
          type: PathPrefix
          value: "/"
      filters:
        - type: ExtensionRef
          extensionRef:
            group: gateway.envoyproxy.io
            kind: HTTPRouteFilter
            name: regex-path-rewrite
      backendRefs:
      - name: backend
        port: 3000
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: HTTPRouteFilter
metadata:
  name: regex-path-rewrite
spec:
  urlRewrite:
    path:
      type: ReplaceRegexMatch
      replaceRegexMatch:
        pattern: '^/service/([^/]+)(/.*)$'
        substitution: '\2/instance/\1'
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
  name: http-filter-url-regex-rewrite
spec:
  parentRefs:
    - name: eg
  hostnames:
    - path.regex.rewrite.example
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: "/"
      filters:
        - type: ExtensionRef
          extensionRef:
            group: gateway.envoyproxy.io
            kind: HTTPRouteFilter
            name: regex-path-rewrite
      backendRefs:
        - name: backend
          port: 3000
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: HTTPRouteFilter
metadata:
  name: regex-path-rewrite
spec:
  urlRewrite:
    path:
      type: ReplaceRegexMatch
      replaceRegexMatch:
        pattern: '^/service/([^/]+)(/.*)$'
        substitution: '\2/instance/\1'
```

{{% /tab %}}
{{< /tabpane >}}

The HTTPRoute status should indicate that it has been accepted and is bound to the example Gateway.

```shell
kubectl get httproute/http-filter-url-regex-rewrite -o yaml
```

Querying `http://${GATEWAY_HOST}/service/foo/v1/api` should rewrite the request to
`http://${GATEWAY_HOST}/service/foo/v1/api`.

```console
$ curl -L -vvv --header "Host: path.regex.rewrite.example" "http://${GATEWAY_HOST}/service/foo/v1/api"
...
> GET /service/foo/v1/api HTTP/1.1
> Host: path.regex.rewrite.example
> User-Agent: curl/8.7.1
> Accept: */*
>
* Request completely sent off
< HTTP/1.1 200 OK
< content-type: application/json
< x-content-type-options: nosniff
< date: Mon, 16 Sep 2024 18:49:48 GMT
< content-length: 482
<
{
 "path": "/v1/api/instance/foo",
 "host": "path.regex.rewrite.example",
 "method": "GET",
 "proto": "HTTP/1.1",
 "headers": {
  "Accept": [
   "*/*"
  ],
  "User-Agent": [
   "curl/8.7.1"
  ],
  "X-Envoy-Internal": [
   "true"
  ],
  "X-Forwarded-For": [
   "10.244.0.37"
  ],
  "X-Forwarded-Proto": [
   "http"
  ],
  "X-Request-Id": [
   "24a5958f-1bfa-4694-a9c1-807d5139a18a"
  ]
 },
 "namespace": "default",
 "ingress": "",
 "service": "",
 "pod": "backend-765694d47f-lzmpm"
...
```

You can see that the path is rewritten from `/service/foo/v1/api`, to `/v1/api/instance/foo`.

## Rewrite Host Name

You can configure to rewrite the hostname like below. In this example, any requests sent to
`http://${GATEWAY_HOST}/get` with `--header "Host: path.rewrite.example"` will rewrite host into `envoygateway.io`.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: http-filter-url-rewrite
spec:
  parentRefs:
    - name: eg
  hostnames:
    - path.rewrite.example
  rules:
    - matches:
      - path:
          type: PathPrefix
          value: "/get"
      filters:
      - type: URLRewrite
        urlRewrite:
          hostname: "envoygateway.io"
      backendRefs:
      - name: backend
        port: 3000
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
  name: http-filter-url-rewrite
spec:
  parentRefs:
    - name: eg
  hostnames:
    - path.rewrite.example
  rules:
    - matches:
      - path:
          type: PathPrefix
          value: "/get"
      filters:
      - type: URLRewrite
        urlRewrite:
          hostname: "envoygateway.io"
      backendRefs:
      - name: backend
        port: 3000
```

{{% /tab %}}
{{< /tabpane >}}

The HTTPRoute status should indicate that it has been accepted and is bound to the example Gateway.

```shell
kubectl get httproute/http-filter-url-rewrite -o yaml
```

Querying `http://${GATEWAY_HOST}/get` with `--header "Host: path.rewrite.example"` will rewrite host into
`envoygateway.io`.

```console
$ curl -L -vvv --header "Host: path.rewrite.example" "http://${GATEWAY_HOST}/get"
...
> GET /get HTTP/1.1
> Host: path.rewrite.example
> User-Agent: curl/7.85.0
> Accept: */*
>
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< content-type: application/json
< x-content-type-options: nosniff
< date: Wed, 21 Dec 2022 11:15:15 GMT
< content-length: 481
< x-envoy-upstream-service-time: 0
< server: envoy
<
{
 "path": "/get",
 "host": "envoygateway.io",
 "method": "GET",
 "proto": "HTTP/1.1",
 "headers": {
  "Accept": [
   "*/*"
  ],
  "User-Agent": [
   "curl/7.85.0"
  ],
  "X-Envoy-Expected-Rq-Timeout-Ms": [
   "15000"
  ],
  "X-Forwarded-Host": [
   "path.rewrite.example"
  ],
  "X-Forwarded-Proto": [
   "http"
  ],
  "X-Request-Id": [
   "39aa447c-97b9-45a3-a675-9fb266ab1af0"
  ]
 },
 "namespace": "default",
 "ingress": "",
 "service": "",
 "pod": "backend-6fdd4b9bd8-8vlc5"
...
```

You can see that the `X-Forwarded-Host` is `path.rewrite.example`, but the actual host is `envoygateway.io`.

## Rewrite URL Host Name by Header or Backend

In addition to core Gateway-API rewrite options, Envoy Gateway supports extended rewrite options through the [HTTPRouteFilter][] API.
The `HTTPRouteFilter` API can be configured to rewrite the Host header value to:
- The value of a different request header
- The DNS name of the backend that the request is routed to

In the following example, the host header is rewritten to the value of the x-custom-host header.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: http-filter-hostname-header-rewrite
spec:
  parentRefs:
    - name: eg
  hostnames:
    - host.header.rewrite.example
  rules:
    - matches:
      - path:
          type: PathPrefix
          value: "/header"
      filters:
        - type: ExtensionRef
          extensionRef:
            group: gateway.envoyproxy.io
            kind: HTTPRouteFilter
            name: header-host-rewrite
      backendRefs:
      - name: backend
        port: 3000
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: HTTPRouteFilter
metadata:
  name: header-host-rewrite
spec:
  urlRewrite:
    hostname:
      type: Header
      header: x-custom-host
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
  name: http-filter-hostname-header-rewrite
spec:
  parentRefs:
    - name: eg
  hostnames:
    - host.header.rewrite.example
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: "/header"
      filters:
        - type: ExtensionRef
          extensionRef:
            group: gateway.envoyproxy.io
            kind: HTTPRouteFilter
            name: header-host-rewrite
      backendRefs:
        - name: backend
          port: 3000
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: HTTPRouteFilter
metadata:
  name: header-host-rewrite
spec:
  urlRewrite:
    hostname:
      type: Header
      header: x-custom-host
```

{{% /tab %}}
{{< /tabpane >}}

The HTTPRoute status should indicate that it has been accepted and is bound to the example Gateway.

```shell
kubectl get httproute/http-filter-header-host-rewrite -o yaml
```

Querying `http://${GATEWAY_HOST}/header` and providing a custom host rewrite header x-custom-host should rewrite the
request host header to the value of the x-custom-host header.

```console
$ curl -L -vvv --header "Host: host.header.rewrite.example" --header "x-custom-host: foo" "http://${GATEWAY_HOST}/header"
...
> GET /header HTTP/1.1
> Host: host.header.rewrite.example
> User-Agent: curl/8.7.1
> Accept: */*
> x-custom-host: foo
>
* Request completely sent off
< HTTP/1.1 200 OK
<
{
 "path": "/header",
 "host": "foo",
 "method": "GET",
 "proto": "HTTP/1.1",
 "headers": {
  "X-Custom-Host": [
   "foo"
  ],
  "X-Forwarded-Host": [
   "host.header.rewrite.example"
  ],
 },
 "namespace": "default",
 "ingress": "",
 "service": "",
 "pod": "backend-765694d47f-5t6f2"
...
```

You can see that the host is rewritten from `host.header.rewrite.example`, to the value of the provided
`x-custom-host` header `foo`. The original host header is preserved in the `X-Forwarded-Host` header.


[HTTPURLRewriteFilter]: https://gateway-api.sigs.k8s.io/reference/spec#gateway.networking.k8s.io/v1.HTTPURLRewriteFilter
[HTTPRouteFilter]: ../../../api/extension_types#httproutefilter
[RE2]: https://github.com/google/re2/wiki/Syntax
