---
title: "HTTP Response Headers"
---

The [HTTPRoute][] resource can modify the headers of a response before responding it to the downstream service. To learn
more about HTTP routing, refer to the [Gateway API documentation][].

A [`ResponseHeaderModifier` filter][req_filter] instructs Gateways to modify the headers in responses that match the
rule before responding to the downstream. Note that the `ResponseHeaderModifier` filter will only modify headers before
the response is returned from Envoy to the downstream client and will not affect request headers forwarding to the
upstream service.

## Prerequisites

Follow the steps from the [Quickstart Guide](../quickstart) to install Envoy Gateway and the example manifest.
Before proceeding, you should be able to query the example backend using HTTP.

## Adding Response Headers

The `ResponseHeaderModifier` filter can add new headers to a response before it is sent to the upstream. If the response
does not have the header configured by the filter, then that header will be added to the response. If the response
already has the header configured by the filter, then the value of the header in the filter will be appended to the
value of the header in the response.

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: http-headers
spec:
  parentRefs:
  - name: eg
  hostnames:
  - headers.example
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
      weight: 1
    filters:
    - type: ResponseHeaderModifier
      responseHeaderModifier:
        add:
        - name: "add-header"
          value: "foo"
EOF
```

The HTTPRoute status should indicate that it has been accepted and is bound to the example Gateway.

```shell
kubectl get httproute/http-headers -o yaml
```

Get the Gateway's address:

```shell
export GATEWAY_HOST=$(kubectl get gateway/eg -o jsonpath='{.status.addresses[0].value}')
```

Querying `headers.example/get` should result in a `200` response from the example Gateway and the output from the
example app should indicate that the downstream client received the header `add-header` with the value: `foo`

```console
$ curl -vvv --header "Host: headers.example" "http://${GATEWAY_HOST}/get" -H 'X-Echo-Set-Header: X-Foo: value1'
...
> GET /get HTTP/1.1
> Host: headers.example
> User-Agent: curl/7.81.0
> Accept: */*
> X-Echo-Set-Header: X-Foo: value1
>
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< content-type: application/json
< x-content-type-options: nosniff
< content-length: 474
< x-envoy-upstream-service-time: 0
< server: envoy
< x-foo: value1
< add-header: foo
<
...
 "headers": {
  "Accept": [
   "*/*"
  ],
  "X-Echo-Set-Header": [
   "X-Foo: value1"
  ]
...
```

## Setting Response Headers

Setting headers is similar to adding headers. If the response does not have the header configured by the filter, then it
will be added, but unlike [adding response headers](#adding-response-headers) which will append the value of the header
if the response already contains it, setting a header will cause the value to be replaced by the value configured in the
filter.

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: http-headers
spec:
  parentRefs:
  - name: eg
  hostnames:
  - headers.example
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
    filters:
    - type: ResponseHeaderModifier
      responseHeaderModifier:
        set:
        - name: "set-header"
          value: "foo"
EOF
```

Querying `headers.example/get` should result in a `200` response from the example Gateway and the output from the
example app should indicate that the downstream client received the header `set-header` with the original value `value1`
replaced by `foo`.

```console
$ curl -vvv --header "Host: headers.example" "http://${GATEWAY_HOST}/get" -H 'X-Echo-Set-Header: set-header: value1'
...
> GET /get HTTP/1.1
> Host: headers.example
> User-Agent: curl/7.81.0
> Accept: */*
> X-Echo-Set-Header: set-header: value1
>
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< content-type: application/json
< x-content-type-options: nosniff
< content-length: 474
< x-envoy-upstream-service-time: 0
< server: envoy
< set-header: foo
<
 "headers": {
  "Accept": [
   "*/*"
  ],
  "X-Echo-Set-Header": [
    "set-header": value1"
  ]
...
```

## Removing Response Headers

Headers can be removed from a response by simply supplying a list of header names.

Setting headers is similar to adding headers. If the response does not have the header configured by the filter, then it
will be added, but unlike [adding response headers](#adding-response-headers) which will append the value of the header
if the response already contains it, setting a header will cause the value to be replaced by the value configured in the
filter.

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: http-headers
spec:
  parentRefs:
  - name: eg
  hostnames:
  - headers.example
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /
    backendRefs:
    - group: ""
      name: backend
      port: 3000
      weight: 1
    filters:
    - type: ResponseHeaderModifier
      responseHeaderModifier:
        remove:
        - "remove-header"
EOF
```

Querying `headers.example/get` should result in a `200` response from the example Gateway and the output from the
example app should indicate that the header `remove-header` that was sent by curl was removed before the upstream
received the response.

```console
$ curl -vvv --header "Host: headers.example" "http://${GATEWAY_HOST}/get" -H 'X-Echo-Set-Header: remove-header: value1'
...
> GET /get HTTP/1.1
> Host: headers.example
> User-Agent: curl/7.81.0
> Accept: */*
> X-Echo-Set-Header: remove-header: value1
>
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< content-type: application/json
< x-content-type-options: nosniff
< content-length: 474
< x-envoy-upstream-service-time: 0
< server: envoy
<

 "headers": {
  "Accept": [
   "*/*"
  ],
  "X-Echo-Set-Header": [
    "remove-header": value1"
  ]
...
```

## Combining Filters

Headers can be added/set/removed in a single filter on the same HTTPRoute and they will all perform as expected

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: http-headers
spec:
  parentRefs:
  - name: eg
  hostnames:
  - headers.example
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
      weight: 1
    filters:
    - type: ResponseHeaderModifier
      responseHeaderModifier:
        add:
        - name: "add-header-1"
          value: "foo"
        set:
        - name: "set-header-1"
          value: "bar"
        remove:
        - "removed-header"
EOF
```

[HTTPRoute]: https://gateway-api.sigs.k8s.io/api-types/httproute/
[Gateway API documentation]: https://gateway-api.sigs.k8s.io/
[req_filter]: https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.HTTPHeaderFilter
