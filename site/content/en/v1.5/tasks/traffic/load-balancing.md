---
title: "Load Balancing"
---

[Envoy load balancing][] is a way of distributing traffic between multiple hosts within a single upstream cluster 
in order to effectively make use of available resources.

Envoy Gateway supports the following load balancing policies:

- **Round Robin**: a simple policy in which each available upstream host is selected in round robin order.
- **Random**: load balancer selects a random available host.
- **Least Request**: load balancer uses different algorithms depending on whether hosts have the same or different weights.
- **Consistent Hash**: load balancer implements consistent hashing to upstream hosts.

Additionally, Envoy Gateway supports **Endpoint Override** functionality that allows endpoint selection based on headers or metadata, which can be used with any of the above load balancing policies as a fallback.

Envoy Gateway introduces a new CRD called [BackendTrafficPolicy][] that allows the user to describe their desired load balancing polices.
This instantiated resource can be linked to a [Gateway][], [HTTPRoute][] or [GRPCRoute][] resource. If `loadBalancer` is not specified in [BackendTrafficPolicy][], the default load balancing policy is `Least Request`.

## Prerequisites

### Install Envoy Gateway

{{< boilerplate prerequisites >}}

For better testing the load balancer, you can add more hosts in upstream cluster by increasing the replicas of one deployment:

```shell
kubectl patch deployment backend -n default -p '{"spec": {"replicas": 4}}'
```

### Install the hey load testing tool

Install the `Hey` CLI tool, this tool will be used to generate load and measure response times. 

Follow the installation instruction from the [Hey project] docs.

## Round Robin

This example will create a Load Balancer with Round Robin policy via [BackendTrafficPolicy][].

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: round-robin-policy
  namespace: default
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: round-robin-route
  loadBalancer:
    type: RoundRobin
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: round-robin-route
  namespace: default
spec:
  parentRefs:
    - name: eg
  hostnames:
    - "www.example.com"
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /round
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
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: round-robin-policy
  namespace: default
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: round-robin-route
  loadBalancer:
    type: RoundRobin
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: round-robin-route
  namespace: default
spec:
  parentRefs:
    - name: eg
  hostnames:
    - "www.example.com"
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /round
      backendRefs:
        - name: backend
          port: 3000
```

{{% /tab %}}
{{< /tabpane >}}

The `hey` tool will be used to generate 100 concurrent requests.

```shell
hey -n 100 -c 100 -host "www.example.com" http://${GATEWAY_HOST}/round
```

```console
Summary:
  Total:	0.0487 secs
  Slowest:	0.0440 secs
  Fastest:	0.0181 secs
  Average:	0.0307 secs
  Requests/sec:	2053.1676

  Total data:	50500 bytes
  Size/request:	505 bytes

Response time histogram:
  0.018 [1]	    |■■
  0.021 [2]  	|■■■■
  0.023 [10]	|■■■■■■■■■■■■■■■■■■■■■■
  0.026 [16]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.028 [7]  	|■■■■■■■■■■■■■■■■
  0.031 [10]	|■■■■■■■■■■■■■■■■■■■■■■
  0.034 [17]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.036 [18]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.039 [11]	|■■■■■■■■■■■■■■■■■■■■■■■■
  0.041 [6] 	|■■■■■■■■■■■■■
  0.044 [2]	    |■■■■
```

As a result, you can see all available upstream hosts receive traffics evenly.

```shell
kubectl get pods -l app=backend --no-headers -o custom-columns=":metadata.name" | while read -r pod; do echo "$pod: received $(($(kubectl logs $pod | wc -l) - 2)) requests"; done
```

```console
backend-69fcff487f-2gfp7: received 26 requests
backend-69fcff487f-69g8c: received 25 requests
backend-69fcff487f-bqwpr: received 24 requests
backend-69fcff487f-kbn8l: received 25 requests
```

You should note that this results may vary, the output here is for reference purpose only.

## Random

This example will create a Load Balancer with Random policy via [BackendTrafficPolicy][].

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: random-policy
  namespace: default
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: random-route
  loadBalancer:
    type: Random
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: random-route
  namespace: default
spec:
  parentRefs:
    - name: eg
  hostnames:
    - "www.example.com"
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /random
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
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: random-policy
  namespace: default
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: random-route
  loadBalancer:
    type: Random
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: random-route
  namespace: default
spec:
  parentRefs:
    - name: eg
  hostnames:
    - "www.example.com"
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /random
      backendRefs:
        - name: backend
          port: 3000
```

{{% /tab %}}
{{< /tabpane >}}

The `hey` tool will be used to generate 1000 concurrent requests.

```shell
hey -n 1000 -c 100 -host "www.example.com" http://${GATEWAY_HOST}/random
```

```console
Summary:
  Total:	0.2624 secs
  Slowest:	0.0851 secs
  Fastest:	0.0007 secs
  Average:	0.0179 secs
  Requests/sec:	3811.3020

  Total data:	506000 bytes
  Size/request:	506 bytes

Response time histogram:
  0.001 [1] 	|
  0.009 [421]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.018 [219]	|■■■■■■■■■■■■■■■■■■■■■
  0.026 [118]	|■■■■■■■■■■■
  0.034 [64]	|■■■■■■
  0.043 [73]	|■■■■■■■
  0.051 [41]	|■■■■
  0.060 [22]	|■■
  0.068 [19]	|■■
  0.077 [13]	|■
  0.085 [9] 	|■
```

As a result, you can see all available upstream hosts receive traffics randomly.

```shell
kubectl get pods -l app=backend --no-headers -o custom-columns=":metadata.name" | while read -r pod; do echo "$pod: received $(($(kubectl logs $pod | wc -l) - 2)) requests"; done
```

```console
backend-69fcff487f-bf6lm: received 246 requests
backend-69fcff487f-gwmqk: received 256 requests
backend-69fcff487f-mzngr: received 230 requests
backend-69fcff487f-xghqq: received 268 requests
```

You should note that this results may vary, the output here is for reference purpose only.

## Least Request

This example will create a Load Balancer with Least Request policy via [BackendTrafficPolicy][].

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: least-request-policy
  namespace: default
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: least-request-route
  loadBalancer:
    type: LeastRequest
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: least-request-route
  namespace: default
spec:
  parentRefs:
    - name: eg
  hostnames:
    - "www.example.com"
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /least
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
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: least-request-policy
  namespace: default
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: least-request-route
  loadBalancer:
    type: LeastRequest
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: least-request-route
  namespace: default
spec:
  parentRefs:
    - name: eg
  hostnames:
    - "www.example.com"
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /least
      backendRefs:
        - name: backend
          port: 3000
```

{{% /tab %}}
{{< /tabpane >}}

The `hey` tool will be used to generate 100 concurrent requests.

```shell
hey -n 100 -c 100 -host "www.example.com" http://${GATEWAY_HOST}/least
```

```console
Summary:
  Total:	0.0489 secs
  Slowest:	0.0479 secs
  Fastest:	0.0054 secs
  Average:	0.0297 secs
  Requests/sec:	2045.9317

  Total data:	50500 bytes
  Size/request:	505 bytes

Response time histogram:
  0.005 [1] 	|■■
  0.010 [1] 	|■■
  0.014 [8] 	|■■■■■■■■■■■■■■■
  0.018 [6] 	|■■■■■■■■■■■
  0.022 [11]	|■■■■■■■■■■■■■■■■■■■■
  0.027 [7] 	|■■■■■■■■■■■■■
  0.031 [15]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.035 [13]	|■■■■■■■■■■■■■■■■■■■■■■■■
  0.039 [22]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.044 [12]	|■■■■■■■■■■■■■■■■■■■■■■
  0.048 [4] 	|■■■■■■■
```

As a result, you can see all available upstream hosts receive traffics randomly, 
and host `backend-69fcff487f-6l2pw` receives fewer requests than others.

```shell
kubectl get pods -l app=backend --no-headers -o custom-columns=":metadata.name" | while read -r pod; do echo "$pod: received $(($(kubectl logs $pod | wc -l) - 2)) requests"; done
```

```console
backend-69fcff487f-59hvs: received 24 requests
backend-69fcff487f-6l2pw: received 19 requests
backend-69fcff487f-ktsx4: received 30 requests
backend-69fcff487f-nqxc7: received 27 requests
```

If you send one more requests to the `${GATEWAY_HOST}/least`, you can tell that host `backend-69fcff487f-6l2pw` is very likely
to get the attention of load balancer and receive this request.

```console
backend-69fcff487f-59hvs: received 24 requests
backend-69fcff487f-6l2pw: received 20 requests
backend-69fcff487f-ktsx4: received 30 requests
backend-69fcff487f-nqxc7: received 27 requests
```

You should note that this results may vary, the output here is for reference purpose only.

## Consistent Hash

This example will create a Load Balancer with Consistent Hash policy via [BackendTrafficPolicy][].

The underlying consistent hash algorithm that Envoy Gateway utilise is [Maglev][], and it can derive hash from following aspects:

- **SourceIP**
- **Header**
- **Cookie**

They are also the supported value as consistent hash type.

### Source IP

This example will create a Load Balancer with Source IP based Consistent Hash policy.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: source-ip-policy
  namespace: default
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: source-ip-route
  loadBalancer:
    type: ConsistentHash
    consistentHash:
      type: SourceIP
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: source-ip-route
  namespace: default
spec:
  parentRefs:
    - name: eg
  hostnames:
    - "www.example.com"
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /source
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
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: source-ip-policy
  namespace: default
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: source-ip-route
  loadBalancer:
    type: ConsistentHash
    consistentHash:
      type: SourceIP
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: source-ip-route
  namespace: default
spec:
  parentRefs:
    - name: eg
  hostnames:
    - "www.example.com"
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /source
      backendRefs:
        - name: backend
          port: 3000
```

{{% /tab %}}
{{< /tabpane >}}

The `hey` tool will be used to generate 100 concurrent requests.

```shell
hey -n 100 -c 100 -host "www.example.com" http://${GATEWAY_HOST}/source
```

```console
Summary:
  Total:	0.0539 secs
  Slowest:	0.0500 secs
  Fastest:	0.0198 secs
  Average:	0.0340 secs
  Requests/sec:	1856.5666

  Total data:	50600 bytes
  Size/request:	506 bytes

Response time histogram:
  0.020 [1] 	|■■
  0.023 [5] 	|■■■■■■■■■■■
  0.026 [12]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.029 [16]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.032 [11]	|■■■■■■■■■■■■■■■■■■■■■■■■
  0.035 [7] 	|■■■■■■■■■■■■■■■■
  0.038 [8] 	|■■■■■■■■■■■■■■■■■■
  0.041 [18]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.044 [15]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.047 [4] 	|■■■■■■■■■
  0.050 [3] 	|■■■■■■■
```

As a result, you can see all traffics are routed to only one upstream host, since the client that send requests
has the same source IP.

```shell
kubectl get pods -l app=backend --no-headers -o custom-columns=":metadata.name" | while read -r pod; do echo "$pod: received $(($(kubectl logs $pod | wc -l) - 2)) requests"; done
```

```console
backend-69fcff487f-grzkj: received 0 requests
backend-69fcff487f-n4d8w: received 100 requests
backend-69fcff487f-tb7zx: received 0 requests
backend-69fcff487f-wbzpg: received 0 requests
```

You can try different client to send out these requests, the upstream host that receives traffics may vary.

### Header

This example will create a Load Balancer with Header based Consistent Hash policy.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: header-policy
  namespace: default
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: header-route
  loadBalancer:
    type: ConsistentHash
    consistentHash:
      type: Header
      header:
        name: FooBar
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: header-route
  namespace: default
spec:
  parentRefs:
    - name: eg
  hostnames:
    - "www.example.com"
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /header
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
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: header-policy
  namespace: default
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: header-route
  loadBalancer:
    type: ConsistentHash
    consistentHash:
      type: Header
      header:
        name: FooBar
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: header-route
  namespace: default
spec:
  parentRefs:
    - name: eg
  hostnames:
    - "www.example.com"
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /header
      backendRefs:
        - name: backend
          port: 3000
```

{{% /tab %}}
{{< /tabpane >}}

The `hey` tool will be used to generate 100 concurrent requests.

```shell
hey -n 100 -c 100 -host "www.example.com" -H "FooBar: 1.2.3.4" http://${GATEWAY_HOST}/header
```

```console
Summary:
  Total:	0.0579 secs
  Slowest:	0.0510 secs
  Fastest:	0.0323 secs
  Average:	0.0431 secs
  Requests/sec:	1728.6064

  Total data:	53800 bytes
  Size/request:	538 bytes

Response time histogram:
  0.032 [1] 	|■■
  0.034 [3] 	|■■■■■■
  0.036 [1] 	|■■
  0.038 [1] 	|■■
  0.040 [7] 	|■■■■■■■■■■■■■■
  0.042 [20]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.044 [20]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.045 [20]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.047 [16]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.049 [9] 	|■■■■■■■■■■■■■■■■■■
  0.051 [2] 	|■■■■
```

As a result, you can see all traffics are routed to only one upstream host, since the header of all requests are the same.

```shell
kubectl get pods -l app=backend --no-headers -o custom-columns=":metadata.name" | while read -r pod; do echo "$pod: received $(($(kubectl logs $pod | wc -l) - 2)) requests"; done
```

```console
backend-69fcff487f-dvt9r: received 0 requests
backend-69fcff487f-f8qdl: received 100 requests
backend-69fcff487f-gnpm4: received 0 requests
backend-69fcff487f-t2pgm: received 0 requests
```

You can try to add different header to these requests, and the upstream host that receives traffics may vary.
The following output happens when you use `hey` to send another 100 requests with header `FooBar: 5.6.7.8`.

```console
backend-69fcff487f-dvt9r: received 0 requests
backend-69fcff487f-f8qdl: received 100 requests
backend-69fcff487f-gnpm4: received 100 requests
backend-69fcff487f-t2pgm: received 0 requests
```

### Cookie

This example will create a Load Balancer with Cookie based Consistent Hash policy.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: cookie-policy
  namespace: default
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: cookie-route
  loadBalancer:
    type: ConsistentHash
    consistentHash:
      type: Cookie
      cookie:
        name: FooBar
        ttl: 60s
        attributes:
          SameSite: Strict
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: cookie-route
  namespace: default
spec:
  parentRefs:
    - name: eg
  hostnames:
    - "www.example.com"
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /cookie
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
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: cookie-policy
  namespace: default
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: cookie-route
  loadBalancer:
    type: ConsistentHash
    consistentHash:
      type: Cookie
      cookie:
        name: FooBar
        ttl: 60s
        attributes:
          SameSite: Strict
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: cookie-route
  namespace: default
spec:
  parentRefs:
    - name: eg
  hostnames:
    - "www.example.com"
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /cookie
      backendRefs:
        - name: backend
          port: 3000
```

{{% /tab %}}
{{< /tabpane >}}

By sending 10 request with `curl` to the `${GATEWAY_HOST}/cookie`, you can see that all requests got routed to only 
one upstream host, since they have same cookie setting.

```shell
for i in {1..10}; do curl -I --header "Host: www.example.com" --cookie "FooBar=1.2.3.4" http://${GATEWAY_HOST}/cookie ; sleep 1; done
```

```shell
kubectl get pods -l app=backend --no-headers -o custom-columns=":metadata.name" | while read -r pod; do echo "$pod: received $(($(kubectl logs $pod | wc -l) - 2)) requests"; done
```

```console
backend-69fcff487f-5dxz9: received 0 requests
backend-69fcff487f-gpvl2: received 0 requests
backend-69fcff487f-pglgv: received 10 requests
backend-69fcff487f-qxr74: received 0 requests
```

You can try to set different cookie to these requests, the upstream host that receives traffics may vary.
The following output happens when you use `curl` to send another 10 requests with cookie `FooBar: 5.6.7.8`.

```console
backend-69fcff487f-dvt9r: received 0 requests
backend-69fcff487f-f8qdl: received 0 requests
backend-69fcff487f-gnpm4: received 10 requests
backend-69fcff487f-t2pgm: received 10 requests
```

If the cookie has not been set in one request, Envoy Gateway will auto-generate a cookie for this request 
according to the `ttl` and `attributes` field.

In this example, the following cookie will be generated (see `set-cookie` header in response) if sending a request without cookie:

```shell
curl -v --header "Host: www.example.com" http://${GATEWAY_HOST}/cookie
```

```console
> GET /cookie HTTP/1.1
> Host: www.example.com
> User-Agent: curl/7.74.0
> Accept: */*
>
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< content-type: application/json
< x-content-type-options: nosniff
< date: Fri, 19 Jul 2024 16:49:57 GMT
< content-length: 458
< set-cookie: FooBar="88358b9442700c56"; Max-Age=60; SameSite=Strict; HttpOnly
<
{
 "path": "/cookie",
 "host": "www.example.com",
 "method": "GET",
 "proto": "HTTP/1.1",
 "headers": {
  "Accept": [
   "*/*"
  ],
  "User-Agent": [
   "curl/7.74.0"
  ],
  "X-Envoy-Internal": [
   "true"
  ],
  "X-Forwarded-For": [
   "10.244.0.1"
  ],
  "X-Forwarded-Proto": [
   "http"
  ],
  "X-Request-Id": [
   "1adeaaf7-d45c-48c8-9a4d-eadbccb2fd50"
  ]
 },
 "namespace": "default",
 "ingress": "",
 "service": "",
 "pod": "backend-69fcff487f-5dxz9"
```

## Endpoint Override

This example will create a Load Balancer with Endpoint Override functionality via [BackendTrafficPolicy][].

The Endpoint Override feature allows endpoint selection based on headers. It can derive the target endpoint from HTTP request headers.

When the specified override endpoint is not available or invalid, the load balancer will fall back to the configured load balancing policy.

### Header-based Endpoint Override

This example will create a Load Balancer with Header-based Endpoint Override functionality.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: endpoint-override-header-policy
  namespace: default
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: endpoint-override-header-route
  loadBalancer:
    type: RoundRobin
    endpointOverride:
      extractFrom:
        - header: x-custom-host
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: endpoint-override-header-route
  namespace: default
spec:
  parentRefs:
    - name: eg
  hostnames:
    - "www.example.com"
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /endpoint-override-header
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
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: endpoint-override-header-policy
  namespace: default
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: endpoint-override-header-route
  loadBalancer:
    type: RoundRobin
    endpointOverride:
      extractFrom:
        - header: x-custom-host
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: endpoint-override-header-route
  namespace: default
spec:
  parentRefs:
    - name: eg
  hostnames:
    - "www.example.com"
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /endpoint-override-header
      backendRefs:
        - name: backend
          port: 3000
```

{{% /tab %}}
{{< /tabpane >}}

First, get one of the backend pod IPs to use as the override endpoint:

```shell
BACKEND_POD_IP=$(kubectl get pods -l app=backend -o jsonpath='{.items[0].status.podIP}')
echo "Backend Pod IP: $BACKEND_POD_IP"
```

Test with a valid pod IP in the header - all requests should go to the specific pod:

```shell
for i in {1..10}; do
  curl -s -H "Host: www.example.com" -H "x-custom-host: $BACKEND_POD_IP:3000" \
    http://${GATEWAY_HOST}/endpoint-override-header | jq -r '.pod'
done
```

All requests should return the same pod name, demonstrating that the endpoint override is working.

Test with an invalid IP in the header - requests should fall back to round robin:

```shell
for i in {1..10}; do
  curl -s -H "Host: www.example.com" -H "x-custom-host: 192.168.1.100:3000" \
    http://${GATEWAY_HOST}/endpoint-override-header | jq -r '.pod'
done
```

You should see requests distributed across different pods using the round robin fallback policy.

[Envoy load balancing]: https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/upstream/load_balancing/overview
[BackendTrafficPolicy]: ../../../api/extension_types#backendtrafficpolicy
[Gateway]: https://gateway-api.sigs.k8s.io/api-types/gateway/
[HTTPRoute]: https://gateway-api.sigs.k8s.io/api-types/httproute/
[GRPCRoute]: https://gateway-api.sigs.k8s.io/api-types/grpcroute/
[Hey project]: https://github.com/rakyll/hey
[Maglev]: https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/upstream/load_balancing/load_balancers#maglev
