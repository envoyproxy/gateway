---
title: "Global Rate Limit"
---

Rate limit is a feature that allows the user to limit the number of incoming requests to a predefined value based on attributes within the traffic flow.

Here are some reasons why you may want to implement Rate limits

* To prevent malicious activity such as DDoS attacks.
* To prevent applications and its resources (such as a database) from getting overloaded.
* To create API limits based on user entitlements.

Envoy Gateway supports two types of rate limiting: [Global rate limiting][] and [Local rate limiting][].

[Global rate limiting][] applies a shared rate limit to the traffic flowing through all the instances of Envoy proxies where it is configured.
i.e. if the data plane has 2 replicas of Envoy running, and the rate limit is 10 requests/second, this limit is shared and will be hit
if 5 requests pass through the first replica and 5 requests pass through the second replica within the same second.

Envoy Gateway introduces a new CRD called [BackendTrafficPolicy][] that allows the user to describe their rate limit intent. This instantiated resource
can be linked to a [Gateway][], [HTTPRoute][] or [GRPCRoute][] resource.

**Note:** Limit is applied per route. Even if a [BackendTrafficPolicy][] targets a gateway, each route in that gateway
still has a separate rate limit bucket. For example, if a gateway has 2 routes, and the limit is 100r/s, then each route
has its own 100r/s rate limit bucket.

## Prerequisites

### Install Envoy Gateway

{{< boilerplate prerequisites >}}

### Install Redis

* The global rate limit feature is based on [Envoy Ratelimit][] which requires a Redis instance as its caching layer.
Lets install a Redis deployment in the `redis-system` namespce.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

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
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resources to your cluster:

```yaml
---
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
```

{{% /tab %}}
{{< /tabpane >}}

### Enable Global Rate limit in Envoy Gateway

* The default installation of Envoy Gateway installs a default [EnvoyGateway][] configuration and attaches it
using a `ConfigMap`. To enable global rate limit, we need to update this configuration to configure the Redis service
as the backend for rate limit.

You can manually edit the `ConfigMap`:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: envoy-gateway-config
  namespace: envoy-gateway-system
data:
  envoy-gateway.yaml: |
    apiVersion: gateway.envoyproxy.io/v1alpha1
    kind: EnvoyGateway
    ... keep the existing configuration ...
    rateLimit:
      backend:
        type: Redis
        redis:
          url: redis.redis-system.svc.cluster.local:6379
```

Or use the `helm upgrade` command to update the configuration if you have installed Envoy Gateway using Helm:

```bash
helm upgrade eg oci://docker.io/envoyproxy/gateway-helm \
  --set config.envoyGateway.rateLimit.backend.type=Redis \
  --set config.envoyGateway.rateLimit.backend.redis.url="redis.redis-system.svc.cluster.local:6379" \
  --reuse-values \
  -n envoy-gateway-system
```

{{< boilerplate rollout-envoy-gateway >}}

## Rate Limit Specific User

Here is an example of a rate limit implemented by the application developer to limit a specific user by matching on a custom `x-user-id` header
with a value set to `one`.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: policy-httproute
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: http-ratelimit
  rateLimit:
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
  name: policy-httproute
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: http-ratelimit
  rateLimit:
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
```

{{% /tab %}}
{{< /tabpane >}}

### HTTPRoute

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

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

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
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
```

{{% /tab %}}
{{< /tabpane >}}

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

## Rate Limit Distinct Users Except Admin

Here is an example of a rate limit implemented by the application developer to limit distinct users who can be differentiated based on the
value in the `x-user-id` header. Here, user `one` (recognised from the traffic flow using the header `x-user-id` and value `one`) will be rate limited at 3 requests/hour
and so will user `two` (recognised from the traffic flow using the header `x-user-id` and value `two`). But if `x-user-id` is `admin`, it will not be rate limited even beyond 3 requests/hour.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: policy-httproute
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: http-ratelimit
  rateLimit:
    type: Global
    global:
      rules:
      - clientSelectors:
        - headers:
          - type: Distinct
            name: x-user-id
          - name: x-user-id
            value: admin
            invert: true
        limit:
          requests: 3
          unit: Hour
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
  name: policy-httproute
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: http-ratelimit
  rateLimit:
    type: Global
    global:
      rules:
      - clientSelectors:
        - headers:
          - type: Distinct
            name: x-user-id
          - name: x-user-id
            value: admin
            invert: true
        limit:
          requests: 3
          unit: Hour
```

{{% /tab %}}
{{< /tabpane >}}

### HTTPRoute

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

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

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
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
```

{{% /tab %}}
{{< /tabpane >}}

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

But when the value for header `x-user-id` is set to `admin` and 4 requests are sent, all 4 of them should respond with 200 OK.

```shell
for i in {1..4}; do curl -I --header "Host: ratelimit.example" --header "x-user-id: admin" http://${GATEWAY_HOST}/get ; sleep 1; done
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

HTTP/1.1 200 OK
content-type: application/json
x-content-type-options: nosniff
date: Wed, 08 Feb 2023 02:33:33 GMT
content-length: 460
x-envoy-upstream-service-time: 0
server: envoy

```

## Rate Limit All Requests

This example shows you how to rate limit all requests matching the HTTPRoute rule at 3 requests/Hour by leaving the `clientSelectors` field unset.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: policy-httproute
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: http-ratelimit
  rateLimit:
    type: Global
    global:
      rules:
      - limit:
          requests: 3
          unit: Hour
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
  name: policy-httproute
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: http-ratelimit
  rateLimit:
    type: Global
    global:
      rules:
      - limit:
          requests: 3
          unit: Hour
```

{{% /tab %}}
{{< /tabpane >}}

### HTTPRoute

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

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

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
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
```

{{% /tab %}}
{{< /tabpane >}}

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

## Rate Limit Client IP Addresses

Here is an example of a rate limit implemented by the application developer to limit distinct users who can be differentiated based on their
 IP address (also reflected in the  `X-Forwarded-For` header).

Note: EG supports two kinds of rate limit for the IP address: Exact and Distinct.
* Exact means that all IP addresses within the specified Source IP CIDR share the same rate limit bucket.
* Distinct means that each IP address within the specified Source IP CIDR has its own rate limit bucket.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: policy-httproute
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: http-ratelimit
  rateLimit:
    type: Global
    global:
      rules:
      - clientSelectors:
        - sourceCIDR:
            value: 0.0.0.0/0
            type: Distinct
        limit:
          requests: 3
          unit: Hour
---
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

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resources to your cluster:

```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: policy-httproute
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: http-ratelimit
  rateLimit:
    type: Global
    global:
      rules:
      - clientSelectors:
        - sourceCIDR:
            value: 0.0.0.0/0
            type: Distinct
        limit:
          requests: 3
          unit: Hour
---
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
```

{{% /tab %}}
{{< /tabpane >}}

```shell
for i in {1..4}; do curl -I --header "Host: ratelimit.example" http://${GATEWAY_HOST}/get ; sleep 1; done
```

```console
HTTP/1.1 200 OK
content-type: application/json
x-content-type-options: nosniff
date: Tue, 28 Mar 2023 08:28:45 GMT
content-length: 512
x-envoy-upstream-service-time: 0
server: envoy

HTTP/1.1 200 OK
content-type: application/json
x-content-type-options: nosniff
date: Tue, 28 Mar 2023 08:28:46 GMT
content-length: 512
x-envoy-upstream-service-time: 0
server: envoy

HTTP/1.1 200 OK
content-type: application/json
x-content-type-options: nosniff
date: Tue, 28 Mar 2023 08:28:48 GMT
content-length: 512
x-envoy-upstream-service-time: 0
server: envoy

HTTP/1.1 429 Too Many Requests
x-envoy-ratelimited: true
date: Tue, 28 Mar 2023 08:28:48 GMT
server: envoy
transfer-encoding: chunked

```

## Rate Limit Jwt Claims

Here is an example of a rate limit implemented by the application developer to limit distinct users who can be differentiated based on the value of the Jwt claims carried.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: jwt-example
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: example
  jwt:
    providers:
    - name: example
      remoteJWKS:
        uri: https://raw.githubusercontent.com/envoyproxy/gateway/main/examples/kubernetes/jwt/jwks.json
      claimToHeaders:
      - claim: name
        header: x-claim-name
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: policy-httproute
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: example
  rateLimit:
    type: Global
    global:
      rules:
      - clientSelectors:
        - headers:
          - name: x-claim-name
            value: John Doe
        limit:
          requests: 3
          unit: Hour
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: example
spec:
  parentRefs:
  - name: eg
  hostnames:
  - ratelimit.example
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
kind: SecurityPolicy
metadata:
  name: jwt-example
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: example
  jwt:
    providers:
    - name: example
      remoteJWKS:
        uri: https://raw.githubusercontent.com/envoyproxy/gateway/main/examples/kubernetes/jwt/jwks.json
      claimToHeaders:
      - claim: name
        header: x-claim-name
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: policy-httproute
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: example
  rateLimit:
    type: Global
    global:
      rules:
      - clientSelectors:
        - headers:
          - name: x-claim-name
            value: John Doe
        limit:
          requests: 3
          unit: Hour
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: example
spec:
  parentRefs:
  - name: eg
  hostnames:
  - ratelimit.example
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

Get the JWT used for testing request authentication:

```shell
TOKEN=$(curl https://raw.githubusercontent.com/envoyproxy/gateway/main/examples/kubernetes/jwt/test.jwt -s) && echo "$TOKEN" | cut -d '.' -f2 - | base64 --decode
```

```shell
TOKEN1=$(curl https://raw.githubusercontent.com/envoyproxy/gateway/main/examples/kubernetes/jwt/with-different-claim.jwt -s) && echo "$TOKEN1" | cut -d '.' -f2 - | base64 --decode
```

### Rate limit by carrying `TOKEN`

```shell
for i in {1..4}; do curl -I --header "Host: ratelimit.example" --header "Authorization: Bearer $TOKEN" http://${GATEWAY_HOST}/foo ; sleep 1; done
```

```console
HTTP/1.1 200 OK
content-type: application/json
x-content-type-options: nosniff
date: Mon, 12 Jun 2023 12:00:25 GMT
content-length: 561
x-envoy-upstream-service-time: 0
server: envoy


HTTP/1.1 200 OK
content-type: application/json
x-content-type-options: nosniff
date: Mon, 12 Jun 2023 12:00:26 GMT
content-length: 561
x-envoy-upstream-service-time: 0
server: envoy


HTTP/1.1 200 OK
content-type: application/json
x-content-type-options: nosniff
date: Mon, 12 Jun 2023 12:00:27 GMT
content-length: 561
x-envoy-upstream-service-time: 0
server: envoy


HTTP/1.1 429 Too Many Requests
x-envoy-ratelimited: true
date: Mon, 12 Jun 2023 12:00:28 GMT
server: envoy
transfer-encoding: chunked

```

### No Rate Limit by carrying `TOKEN1`

```shell
for i in {1..4}; do curl -I --header "Host: ratelimit.example" --header "Authorization: Bearer $TOKEN1" http://${GATEWAY_HOST}/foo ; sleep 1; done
```

```console
HTTP/1.1 200 OK
content-type: application/json
x-content-type-options: nosniff
date: Mon, 12 Jun 2023 12:02:34 GMT
content-length: 556
x-envoy-upstream-service-time: 0
server: envoy

HTTP/1.1 200 OK
content-type: application/json
x-content-type-options: nosniff
date: Mon, 12 Jun 2023 12:02:35 GMT
content-length: 556
x-envoy-upstream-service-time: 0
server: envoy

HTTP/1.1 200 OK
content-type: application/json
x-content-type-options: nosniff
date: Mon, 12 Jun 2023 12:02:36 GMT
content-length: 556
x-envoy-upstream-service-time: 1
server: envoy

HTTP/1.1 200 OK
content-type: application/json
x-content-type-options: nosniff
date: Mon, 12 Jun 2023 12:02:37 GMT
content-length: 556
x-envoy-upstream-service-time: 0
server: envoy

```

## Rate Limit Policy Merging

This example demonstrates how to implement a layered rate limiting strategy using BackendTrafficPolicy merging. This approach allows platform teams to set global abuse prevention limits at the Gateway level, while application teams can define more specific rate limits for their individual routes.

In this scenario:
- **Platform Team**: Manages a Gateway-level BackendTrafficPolicy that applies a global abuse rate limit (100 requests/second) across all traffic
- **Application Team**: Manages a Route-level BackendTrafficPolicy that applies a more restrictive limit (5 requests/minute) for their specific signup service

The route-level policy uses `mergeType: StrategicMerge` to combine with the gateway-level policy, ensuring both limits are enforced. The global limit acts as an abuse prevention mechanism, while the route-specific limit protects the signup service from overload.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: global-backendtrafficpolicy
spec:
  rateLimit:
    type: Global
    global:
      rules:
      - clientSelectors:
        - sourceCIDR:
            type: Distinct
            value: 0.0.0.0/0
        limit:
          requests: 100
          unit: Second
        shared: true
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: Gateway
    name: eg
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: route-backendtrafficpolicy
spec:
  mergeType: StrategicMerge
  rateLimit:
    type: Global
    global:
      rules:
      - clientSelectors:
        - sourceCIDR:
            type: Distinct
            value: 0.0.0.0/0
        limit:
          requests: 5
          unit: Minute
        shared: false
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: signup-service-httproute
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: signup-service-httproute
spec:
  parentRefs:
  - name: eg
  hostnames:
  - signup.example
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /signup
    backendRefs:
    - group: ""
      kind: Service
      name: backend
      port: 3000
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
  name: global-backendtrafficpolicy
spec:
  rateLimit:
    type: Global
    global:
      rules:
      - clientSelectors:
        - sourceCIDR:
            type: Distinct
            value: 0.0.0.0/0
        limit:
          requests: 100
          unit: Second
        shared: true
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: Gateway
    name: eg
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: route-backendtrafficpolicy
spec:
  mergeType: StrategicMerge
  rateLimit:
    type: Global
    global:
      rules:
      - clientSelectors:
        - sourceCIDR:
            type: Distinct
            value: 0.0.0.0/0
        limit:
          requests: 5
          unit: Minute
        shared: false
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: signup-service-httproute
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: signup-service-httproute
spec:
  parentRefs:
  - name: eg
  hostnames:
  - signup.example
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /signup
    backendRefs:
    - group: ""
      kind: Service
      name: backend
      port: 3000
```

{{% /tab %}}
{{< /tabpane >}}

### Key Configuration Details

- **Gateway-level Policy**: Uses `shared: true` to create a shared rate limit bucket that applies across all routes in the Gateway
- **Route-level Policy**: Uses `shared: false` to create a dedicated rate limit bucket for this specific route
- **Merge Type**: The `mergeType: StrategicMerge` field enables the route policy to merge with the gateway policy rather than override it

### Testing the Merged Rate Limits

Let's verify that both rate limits are enforced. The signup service should be limited to 5 requests per minute, but also subject to the global 100 requests/second abuse limit.

First, let's test the route-specific limit by sending 6 requests within a minute:

```shell
for i in {1..6}; do curl -I --header "Host: signup.example" http://${GATEWAY_HOST}/signup ; sleep 1; done
```

You should see the first 5 requests succeed and the 6th request receive a `429 Too Many Requests` response:

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

...

HTTP/1.1 429 Too Many Requests
x-envoy-ratelimited: true
date: Wed, 08 Feb 2023 02:33:36 GMT
server: envoy
transfer-encoding: chunked
```

The global abuse limit would activate if traffic across all routes in the Gateway exceeds 100 requests/second, providing an additional layer of protection managed by the platform team.

### Use Cases and Team Responsibilities

**Platform Team Responsibilities:**
- Define Gateway-level policies with global abuse rate limits
- Set reasonable baseline protections that apply to all applications
- Use `shared: true` for limits that should be enforced across multiple routes
- Monitor and adjust global limits based on infrastructure capacity

**Application Team Responsibilities:**
- Define Route-level policies with application-specific rate limits
- Use `mergeType: StrategicMerge` to combine with platform policies
- Set `shared: false` for limits specific to their application's requirements
- Implement rate limits that protect their specific service from overload

This layered approach ensures both infrastructure protection and application-specific controls work together effectively.

### (Optional) Editing Kubernetes Resources settings for the Rate Limit Service

* The default installation of Envoy Gateway installs a default [EnvoyGateway][] configuration and provides the initial rate
limit kubernetes resources settings. such as `replicas` is 1, requests resources cpu is `100m`, memory is `512Mi`. the others
like container `image`, `securityContext`, `env` and pod `annotations` and `securityContext` can be modified by modifying the `ConfigMap`.

* `tls.certificateRef` set the client certificate for redis server TLS connections.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: envoy-gateway-config
  namespace: envoy-gateway-system
data:
  envoy-gateway.yaml: |
    apiVersion: gateway.envoyproxy.io/v1alpha1
    kind: EnvoyGateway
    provider:
      type: Kubernetes
      kubernetes:
        rateLimitDeployment:
          replicas: 1
          container:
            image: envoyproxy/ratelimit:master
            env:
            - name: CACHE_KEY_PREFIX
              value: "eg:rl:"
            resources:
              requests:
                cpu: 100m
                memory: 512Mi
            securityContext:
              runAsUser: 2000
              allowPrivilegeEscalation: false
          pod:
            annotations:
              key1: val1
              key2: val2
            securityContext:
              runAsUser: 1000
              runAsGroup: 3000
              fsGroup: 2000
              fsGroupChangePolicy: "OnRootMismatch"
    gateway:
      controllerName: gateway.envoyproxy.io/gatewayclass-controller
    rateLimit:
      backend:
        type: Redis
        redis:
          url: redis.redis-system.svc.cluster.local:6379
          tls:
            certificateRef:
              name: ratelimit-cert
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: envoy-gateway-config
  namespace: envoy-gateway-system
data:
  envoy-gateway.yaml: |
    apiVersion: gateway.envoyproxy.io/v1alpha1
    kind: EnvoyGateway
    provider:
      type: Kubernetes
      kubernetes:
        rateLimitDeployment:
          replicas: 1
          container:
            image: envoyproxy/ratelimit:master
            env:
            - name: CACHE_KEY_PREFIX
              value: "eg:rl:"
            resources:
              requests:
                cpu: 100m
                memory: 512Mi
            securityContext:
              runAsUser: 2000
              allowPrivilegeEscalation: false
          pod:
            annotations:
              key1: val1
              key2: val2
            securityContext:
              runAsUser: 1000
              runAsGroup: 3000
              fsGroup: 2000
              fsGroupChangePolicy: "OnRootMismatch"
    gateway:
      controllerName: gateway.envoyproxy.io/gatewayclass-controller
    rateLimit:
      backend:
        type: Redis
        redis:
          url: redis.redis-system.svc.cluster.local:6379
          tls:
            certificateRef:
              name: ratelimit-cert
```

{{% /tab %}}
{{< /tabpane >}}

{{< boilerplate rollout-envoy-gateway >}}

[Global Rate Limiting]: https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/other_features/global_rate_limiting
[Local rate limiting]: https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/other_features/local_rate_limiting
[BackendTrafficPolicy]: ../../../api/extension_types#backendtrafficpolicy
[Envoy Ratelimit]: https://github.com/envoyproxy/ratelimit
[EnvoyGateway]: ../../api/extension_types#envoygateway
[Gateway]: https://gateway-api.sigs.k8s.io/api-types/gateway/
[HTTPRoute]: https://gateway-api.sigs.k8s.io/api-types/httproute/
[GRPCRoute]: https://gateway-api.sigs.k8s.io/api-types/grpcroute/
