---
title: "HTTP Request Headers"
---

The [HTTPRoute][] resource can modify the headers of a request before forwarding it to the upstream service. HTTPRoute
rules cannot use both filter types at once. Currently, Envoy Gateway only supports __core__ [HTTPRoute filters][] which
consist of `RequestRedirect` and `RequestHeaderModifier` at the time of this writing. To learn more about HTTP routing,
refer to the [Gateway API documentation][].

A [`RequestHeaderModifier` filter][req_filter] instructs Gateways to modify the headers in requests that match the rule
before forwarding the request upstream. Note that the `RequestHeaderModifier` filter will only modify headers before the
request is sent from Envoy to the upstream service and will not affect response headers returned to the downstream
client.

## Prerequisites

Follow the steps from the [Quickstart Guide](../quickstart) to install Envoy Gateway and the example manifest.
Before proceeding, you should be able to query the example backend using HTTP.

## Adding Request Headers

The `RequestHeaderModifier` filter can add new headers to a request before it is sent to the upstream. If the request
does not have the header configured by the filter, then that header will be added to the request. If the request already
has the header configured by the filter, then the value of the header in the filter will be appended to the value of the
header in the request.

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
    - type: RequestHeaderModifier
      requestHeaderModifier:
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
example app should indicate that the upstream example app received the header `add-header` with the value:
`something,foo`

```console
$ curl -vvv --header "Host: headers.example" "http://${GATEWAY_HOST}/get" --header "add-header: something"
...
> GET /get HTTP/1.1
> Host: headers.example
> User-Agent: curl/7.81.0
> Accept: */*
> add-header: something
>
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< content-type: application/json
< x-content-type-options: nosniff
< content-length: 474
< x-envoy-upstream-service-time: 0
< server: envoy
<
...
 "headers": {
  "Accept": [
   "*/*"
  ],
  "Add-Header": [
   "something",
   "foo"
  ],
...
```

## Setting Request Headers

Setting headers is similar to adding headers. If the request does not have the header configured by the filter, then it
will be added, but unlike [adding request headers](#adding-request-headers) which will append the value of the header if
the request already contains it, setting a header will cause the value to be replaced by the value configured in the
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
    - type: RequestHeaderModifier
      requestHeaderModifier:
        set:
        - name: "set-header"
          value: "foo"
EOF
```

Querying `headers.example/get` should result in a `200` response from the example Gateway and the output from the
example app should indicate that the upstream example app received the header `add-header` with the original value
`something` replaced by `foo`.

```console
$ curl -vvv --header "Host: headers.example" "http://${GATEWAY_HOST}/get" --header "set-header: something"
...
> GET /get HTTP/1.1
> Host: headers.example
> User-Agent: curl/7.81.0
> Accept: */*
> add-header: something
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
  "Set-Header": [
   "foo"
  ],
...
```

## Removing Request Headers

Headers can be removed from a request by simply supplying a list of header names.

Setting headers is similar to adding headers. If the request does not have the header configured by the filter, then it
will be added, but unlike [adding request headers](#adding-request-headers) which will append the value of the header if
the request already contains it, setting a header will cause the value to be replaced by the value configured in the
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
    - type: RequestHeaderModifier
      requestHeaderModifier:
        remove:
        - "remove-header"
EOF
```

Querying `headers.example/get` should result in a `200` response from the example Gateway and the output from the
example app should indicate that the upstream example app received the header `add-header`, but the header
`remove-header` that was sent by curl was removed before the upstream received the request.

```console
$ curl -vvv --header "Host: headers.example" "http://${GATEWAY_HOST}/get" --header "add-header: something" --header "remove-header: foo"
...
> GET /get HTTP/1.1
> Host: headers.example
> User-Agent: curl/7.81.0
> Accept: */*
> add-header: something
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
  "Add-Header": [
   "something"
  ],
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
    - type: RequestHeaderModifier
      requestHeaderModifier:
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
[HTTPRoute filters]: https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.HTTPRouteFilter
[Gateway API documentation]: https://gateway-api.sigs.k8s.io/
[req_filter]: https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.HTTPHeaderFilter
