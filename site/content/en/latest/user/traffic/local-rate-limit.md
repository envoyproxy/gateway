---
title: "Local Rate Limit"
---

Rate limit is a feature that allows the user to limit the number of incoming requests to a predefined value based on attributes within the traffic flow.

Here are some reasons why you may want to implement Rate limits

* To prevent malicious activity such as DDoS attacks.
* To prevent applications and its resources (such as a database) from getting overloaded.
* To create API limits based on user entitlements.

Envoy Gateway supports two types of rate limiting: [Global rate limiting][] and [Local rate limiting][].

[Local rate limiting][] applies rate limits to the traffic flowing through a single instance of Envoy proxy. This means 
that if the data plane has 2 replicas of Envoy running, and the rate limit is 10 requests/second, each replica will allow
10 requests/second. This is in contrast to [Global Rate Limiting][] which applies rate limits to the traffic flowing through
all instances of Envoy proxy.

Envoy Gateway introduces a new CRD called [BackendTrafficPolicy][] that allows the user to describe their rate limit intent. 
This instantiated resource can be linked to a [Gateway][], [HTTPRoute][] or [GRPCRoute][] resource.

**Note:** Limit is applied per route. Even if a [BackendTrafficPolicy][] targets a gateway, each route in that gateway 
still has a separate rate limit bucket. For example, if a gateway has 2 routes, and the limit is 100r/s, then each route
has its own 100r/s rate limit bucket.

## Prerequisites

### Install Envoy Gateway

* Follow the steps from the [Quickstart Guide](../../quickstart) to install Envoy Gateway and the HTTPRoute example manifest.
Before proceeding, you should be able to query the example backend using HTTP.

## Rate Limit Specific User 

Here is an example of a rate limit implemented by the application developer to limit a specific user by matching on a custom `x-user-id` header
with a value set to `one`.

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy 
metadata:
  name: policy-httproute
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: http-ratelimit
    namespace: default
  rateLimit:
    type: Local
    local:
      rules:
      - clientSelectors:
        - headers:
          - name: x-user-id
            value: one
        limit:
          requests: 3
          unit: Hour
EOF
```

### HTTPRoute

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: http-ratelimit
spec:
  parentRefs:
  - name: eg
  hostnames:
  - ratelimit.example 
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
EOF
```

The HTTPRoute status should indicate that it has been accepted and is bound to the example Gateway.

```shell
kubectl get httproute/http-ratelimit -o yaml
```

Get the Gateway's address:

```shell
export GATEWAY_HOST=$(kubectl get gateway/eg -o jsonpath='{.status.addresses[0].value}')
```

Let's query `ratelimit.example/get` 4 times. We should receive a `200` response from the example Gateway for the first 3 requests 
and then receive a `429` status code for the 4th request since the limit is set at 3 requests/Hour for the request which contains the header `x-user-id`
and value `one`.

```shell
for i in {1..4}; do curl -I --header "Host: ratelimit.example" --header "x-user-id: one" http://${GATEWAY_HOST}/get ; sleep 1; done
```

```console
HTTP/1.1 200 OK
content-type: application/json
x-content-type-options: nosniff
date: Wed, 08 Feb 2023 02:33:31 GMT
content-length: 460
x-envoy-upstream-service-time: 4
server: envoy

HTTP/1.1 200 OK
content-type: application/json
x-content-type-options: nosniff
date: Wed, 08 Feb 2023 02:33:32 GMT
content-length: 460
x-envoy-upstream-service-time: 2
server: envoy

HTTP/1.1 200 OK
content-type: application/json
x-content-type-options: nosniff
date: Wed, 08 Feb 2023 02:33:33 GMT
content-length: 460
x-envoy-upstream-service-time: 0
server: envoy

HTTP/1.1 429 Too Many Requests
x-envoy-ratelimited: true
date: Wed, 08 Feb 2023 02:33:34 GMT
server: envoy
transfer-encoding: chunked

```

You should be able to send requests with the `x-user-id` header and a different value and receive successful responses from the server.

```shell
for i in {1..4}; do curl -I --header "Host: ratelimit.example" --header "x-user-id: two" http://${GATEWAY_HOST}/get ; sleep 1; done
```

```console
HTTP/1.1 200 OK
content-type: application/json
x-content-type-options: nosniff
date: Wed, 08 Feb 2023 02:34:36 GMT
content-length: 460
x-envoy-upstream-service-time: 0
server: envoy

HTTP/1.1 200 OK
content-type: application/json
x-content-type-options: nosniff
date: Wed, 08 Feb 2023 02:34:37 GMT
content-length: 460
x-envoy-upstream-service-time: 0
server: envoy

HTTP/1.1 200 OK
content-type: application/json
x-content-type-options: nosniff
date: Wed, 08 Feb 2023 02:34:38 GMT
content-length: 460
x-envoy-upstream-service-time: 0
server: envoy

HTTP/1.1 200 OK
content-type: application/json
x-content-type-options: nosniff
date: Wed, 08 Feb 2023 02:34:39 GMT
content-length: 460
x-envoy-upstream-service-time: 0
server: envoy

```

## Rate Limit All Requests 

This example shows you how to rate limit all requests matching the HTTPRoute rule at 3 requests/Hour by leaving the `clientSelectors` field unset.

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy 
metadata:
  name: policy-httproute
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: http-ratelimit
    namespace: default
  rateLimit:
    type: Local
    local:
      rules:
      - limit:
          requests: 3
          unit: Hour
EOF
```

### HTTPRoute

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: http-ratelimit
spec:
  parentRefs:
  - name: eg
  hostnames:
  - ratelimit.example 
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
EOF
```

```shell
for i in {1..4}; do curl -I --header "Host: ratelimit.example" http://${GATEWAY_HOST}/get ; sleep 1; done
```

```console
HTTP/1.1 200 OK
content-type: application/json
x-content-type-options: nosniff
date: Wed, 08 Feb 2023 02:33:31 GMT
content-length: 460
x-envoy-upstream-service-time: 4
server: envoy

HTTP/1.1 200 OK
content-type: application/json
x-content-type-options: nosniff
date: Wed, 08 Feb 2023 02:33:32 GMT
content-length: 460
x-envoy-upstream-service-time: 2
server: envoy

HTTP/1.1 200 OK
content-type: application/json
x-content-type-options: nosniff
date: Wed, 08 Feb 2023 02:33:33 GMT
content-length: 460
x-envoy-upstream-service-time: 0
server: envoy

HTTP/1.1 429 Too Many Requests
x-envoy-ratelimited: true
date: Wed, 08 Feb 2023 02:33:34 GMT
server: envoy
transfer-encoding: chunked

```

**Note:** Local rate limiting does not support `distinct` matching. If you want to rate limit based on distinct values, 
you should use [Global Rate Limiting][]. 

[Global Rate Limiting]: https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/other_features/global_rate_limiting
[Local rate limiting]: https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/other_features/local_rate_limiting
[BackendTrafficPolicy]: ../../../api/extension_types#backendtrafficpolicy
[Gateway]: https://gateway-api.sigs.k8s.io/api-types/gateway/
[HTTPRoute]: https://gateway-api.sigs.k8s.io/api-types/httproute/
[GRPCRoute]: https://gateway-api.sigs.k8s.io/api-types/grpcroute/
