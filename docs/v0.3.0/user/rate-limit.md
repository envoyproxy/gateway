# Rate limit

Rate limit is a feature that allows the user to limit the number of incoming requests to a predefined value based on attributes within the traffic flow.

Here are some reasons why you may want to implements Rate limits

* To prevent malicious activity such as DDoS attacks.
* To prevent applications and its resources (such as a database) from getting overloaded.
* To create API limits based on user entitlements.

Envoy Gateway supports [Global rate limiting][], where the rate limit is common across all the instances of Envoy proxies where its applied
i.e. if the data plane has 2 replicas of Envoy running, and the rate limit is 10 requests/second, this limit is common and will be hit
if 5 requests pass through the first replica and 5 requests pass through the second replica within the same second.

Envoy Gateway introduces a new CRD called [RateLimitFilter][] that allows the user to describe their rate limit intent. This instantiated resource
can be linked to a [HTTPRoute][] resource using an [ExtensionRef][] filter.

## Prerequisites

### Install Envoy Gateway

* Follow the steps from the [Quickstart Guide](quickstart.md) to install Envoy Gateway and the example manifest.
Before proceeding, you should be able to query the example backend using HTTP.

### Install Redis

* The global rate limit feature is based on [Envoy Ratelimit][] which requires a Redis instance as its caching layer.
Lets install a Redis deployment in the `redis-system` namespce.

```shell
cat <<EOF | kubectl apply -f -
kind: Namespace
apiVersion: v1
metadata:
  name: redis-system 
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis
  namespace: redis-system
  labels:
    app: redis
spec:
  replicas: 1
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
      - image: redis:6.0.6
        imagePullPolicy: IfNotPresent
        name: redis
        resources:
          limits:
            cpu: 1500m
            memory: 512Mi
          requests:
            cpu: 200m
            memory: 256Mi
---
apiVersion: v1
kind: Service
metadata:
  name: redis
  namespace: redis-system 
  labels:
    app: redis
  annotations:
spec:
  ports:
  - name: redis
    port: 6379
    protocol: TCP
    targetPort: 6379
  selector:
    app: redis
---

EOF
```

### Enable Global Rate limit in Envoy Gateway

* The default installation of Envoy Gateway installs a default [EnvoyGateway][] configuration and attaches it
using a `ConfigMap`. In the next step, we will update this resource to enable rate limit in Envoy Gateway
as well as configure the URL for the Redis instance used for Global rate limiting.


```shell
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: envoy-gateway-config
  namespace: envoy-gateway-system
data:
  envoy-gateway.yaml: |
    apiVersion: config.gateway.envoyproxy.io/v1alpha1
    kind: EnvoyGateway
    provider:
      type: Kubernetes
    gateway:
      controllerName: gateway.envoyproxy.io/gatewayclass-controller
    rateLimit:
      backend:
        type: Redis
        redis:
          url: redis.redis-system.svc.cluster.local:6379
EOF
```

* After updating the `ConfigMap`, you will need to restart the `envoy-gateway` deployment so the configuration kicks in

```shell
kubectl rollout restart deployment envoy-gateway -n envoy-gateway-system
```

## Rate limit specific user 

Here is an example of a rate limit implemented by the application developer to limit a specific user by matching on a custom `x-user-id` header
with a value set to `one`.

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: RateLimitFilter
metadata:
  name: ratelimit-specific-user
spec:
  type: Global
  global:
    rules:
    - clientSelectors:
      - headers:
        - name: x-user-id
          value: one
      limit:
        requests: 3
        unit: Hour
---
apiVersion: gateway.networking.k8s.io/v1beta1
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
    filters:
    - type: ExtensionRef
      extensionRef:
        group: gateway.envoyproxy.io
        kind: RateLimitFilter
        name: ratelimit-specific-user
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

Lets query `ratelimit.example/get` 4 times. We should receive a `200` response from the example Gateway for the first 3 requests 
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

## Rate limit distinct users 

Here is an example of a rate limit implemented by the application developer to limit distinct users who can be differentiated based on the 
value in the `x-user-id` header. Here, user `one` (recognised from the traffic flow using the header `x-user-id` and value `one`) will be rate limited at 3 requests/hour
and so will user `two` (recognised from the traffic flow using the header `x-user-id` and value `two`).

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: RateLimitFilter
metadata:
  name: ratelimit-distinct-users
spec:
  type: Global
  global:
    rules:
    - clientSelectors:
      - headers:
        - type: Distinct
          name: x-user-id
      limit:
        requests: 3
        unit: Hour
---
apiVersion: gateway.networking.k8s.io/v1beta1
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
    filters:
    - type: ExtensionRef
      extensionRef:
        group: gateway.envoyproxy.io
        kind: RateLimitFilter
        name: ratelimit-distinct-users
    backendRefs:
    - group: ""
      kind: Service
      name: backend
      port: 3000
EOF
```


Lets run the same command again with the header `x-user-id` and value `one` set in the request. We should the first 3 requests succeeding and
the 4th request being rate limited.

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

You should see the same behavior when the value for header `x-user-id` is set to `two` and 4 requests are sent.

```shell
for i in {1..4}; do curl -I --header "Host: ratelimit.example" --header "x-user-id: two" http://${GATEWAY_HOST}/get ; sleep 1; done
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

## Rate limit all requests 

This example shows you how to rate limit all requests matching the HTTPRoute rule at 3 requests/Hour by leaving the `clientSelectors` field unset.

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: RateLimitFilter
metadata:
  name: ratelimit-all-requests
spec:
  type: Global
  global:
    rules:
    - limit:
        requests: 3
        unit: Hour
---
apiVersion: gateway.networking.k8s.io/v1beta1
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
    filters:
    - type: ExtensionRef
      extensionRef:
        group: gateway.envoyproxy.io
        kind: RateLimitFilter
        name: ratelimit-all-requests
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


[Global rate limiting]: https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/other_features/global_rate_limiting
[RateLimitFilter]: https://gateway.envoyproxy.io/v0.3.0/api/extension_types.html#ratelimitfilter
[Envoy Ratelimit]: https://github.com/envoyproxy/ratelimit
[EnvoyGateway]: https://gateway.envoyproxy.io/v0.3.0/api/config_types.html#envoygateway
[HTTPRoute]: https://gateway-api.sigs.k8s.io/api-types/httproute/
[ExtensionRef]: https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io%2fv1beta1.HTTPRouteFilter
