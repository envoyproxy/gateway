---
title: "HTTP URL Rewrite"
---

[HTTPURLRewriteFilter][] defines a filter that modifies a request during forwarding. At most one of these filters may be
used on a Route rule. This MUST NOT be used on the same Route rule as a HTTPRequestRedirect filter.

## Prerequisites

Follow the steps from the [Quickstart Guide](../../quickstart) to install Envoy Gateway and the example manifest.
Before proceeding, you should be able to query the example backend using HTTP.

## Rewrite URL Prefix Path

You can configure to rewrite the prefix in the url like below. In this example, any curls to
`http://${GATEWAY_HOST}/get/xxx` will be rewritten to `http://${GATEWAY_HOST}/replace/xxx`.

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

## Rewrite Host Name

You can configure to rewrite the hostname like below. In this example, any requests sent to
`http://${GATEWAY_HOST}/get` with `--header "Host: path.rewrite.example"` will rewrite host into `envoygateway.io`.

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

[HTTPURLRewriteFilter]: https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.HTTPURLRewriteFilter
