---
title: "Client Traffic Policy"
---

This guide explains the usage of the [ClientTrafficPolicy][] API.


## Introduction

The [ClientTrafficPolicy][] API allows system administrators to configure
the behavior for how the Envoy Proxy server behaves with downstream clients.

## Motivation

This API was added as a new policy attachment resource that can be applied to Gateway resources and it is meant to hold settings for configuring behavior of the connection between the downstream client and Envoy Proxy listener. It is the counterpart to the [BackendTrafficPolicy][] API resource.

## Quickstart

### Prerequisites

* Follow the steps from the [Quickstart](../quickstart) guide to install Envoy Gateway and the example manifest.
Before proceeding, you should be able to query the example backend using HTTP.

### Support TCP keepalive for downstream client

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: ClientTrafficPolicy
metadata:
  name: enable-tcp-keepalive-policy
  namespace: default
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: Gateway
    name: eg
    namespace: default
  tcpKeepalive:
    idleTime: 20m
    interval: 60s
    probes: 3
EOF
```

Verify that ClientTrafficPolicy is Accepted:

```shell
kubectl get clienttrafficpolicies.gateway.envoyproxy.io -n default
```

You should see the policy marked as accepted like this:

```shell
NAME                          STATUS     AGE
enable-tcp-keepalive-policy   Accepted   5s
```

Curl the example app through Envoy proxy once again:

```shell
curl --verbose  --header "Host: www.example.com" http://$GATEWAY_HOST/get --next --header "Host: www.example.com" http://$GATEWAY_HOST/get
```

You should see the output like this:

```shell
*   Trying 172.18.255.202:80...
* Connected to 172.18.255.202 (172.18.255.202) port 80 (#0)
> GET /get HTTP/1.1
> Host: www.example.com
> User-Agent: curl/8.1.2
> Accept: */*
>
< HTTP/1.1 200 OK
< content-type: application/json
< x-content-type-options: nosniff
< date: Fri, 01 Dec 2023 10:17:04 GMT
< content-length: 507
< x-envoy-upstream-service-time: 0
< server: envoy
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
   "curl/8.1.2"
  ],
  "X-Envoy-Expected-Rq-Timeout-Ms": [
   "15000"
  ],
  "X-Envoy-Internal": [
   "true"
  ],
  "X-Forwarded-For": [
   "172.18.0.2"
  ],
  "X-Forwarded-Proto": [
   "http"
  ],
  "X-Request-Id": [
   "4d0d33e8-d611-41f0-9da0-6458eec20fa5"
  ]
 },
 "namespace": "default",
 "ingress": "",
 "service": "",
 "pod": "backend-58d58f745-2zwvn"
* Connection #0 to host 172.18.255.202 left intact
}* Found bundle for host: 0x7fb9f5204ea0 [serially]
* Can not multiplex, even if we wanted to
* Re-using existing connection #0 with host 172.18.255.202
> GET /headers HTTP/1.1
> Host: www.example.com
> User-Agent: curl/8.1.2
> Accept: */*
>
< HTTP/1.1 200 OK
< content-type: application/json
< x-content-type-options: nosniff
< date: Fri, 01 Dec 2023 10:17:04 GMT
< content-length: 511
< x-envoy-upstream-service-time: 0
< server: envoy
<
{
 "path": "/headers",
 "host": "www.example.com",
 "method": "GET",
 "proto": "HTTP/1.1",
 "headers": {
  "Accept": [
   "*/*"
  ],
  "User-Agent": [
   "curl/8.1.2"
  ],
  "X-Envoy-Expected-Rq-Timeout-Ms": [
   "15000"
  ],
  "X-Envoy-Internal": [
   "true"
  ],
  "X-Forwarded-For": [
   "172.18.0.2"
  ],
  "X-Forwarded-Proto": [
   "http"
  ],
  "X-Request-Id": [
   "9a8874c0-c117-481c-9b04-933571732ca5"
  ]
 },
 "namespace": "default",
 "ingress": "",
 "service": "",
 "pod": "backend-58d58f745-2zwvn"
* Connection #0 to host 172.18.255.202 left intact
}
```

You can see keepalive connection marked by the output in:

```shell
* Connection #0 to host 172.18.255.202 left intact
* Re-using existing connection #0 with host 172.18.255.202
```

### Enable Proxy Protocol for downstream client

This example configures Proxy Protocol for downstream clients.

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: ClientTrafficPolicy
metadata:
  name: enable-proxy-protocol-policy
  namespace: default
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: Gateway
    name: eg
    namespace: default
  enableProxyProtocol: true
EOF
```

Verify that ClientTrafficPolicy is Accepted:

```shell
kubectl get clienttrafficpolicies.gateway.envoyproxy.io -n default
```

You should see the policy marked as accepted like this:

```shell
NAME                          STATUS     AGE
enable-proxy-protocol-policy   Accepted   5s
```

Try the endpoint without using PROXY protocol with curl:

```shell
curl -v --header "Host: www.example.com" http://$GATEWAY_HOST/get
```

```shell
*   Trying 172.18.255.202:80...
* Connected to 172.18.255.202 (172.18.255.202) port 80 (#0)
> GET /get HTTP/1.1
> Host: www.example.com
> User-Agent: curl/8.1.2
> Accept: */*
>
* Recv failure: Connection reset by peer
* Closing connection 0
curl: (56) Recv failure: Connection reset by peer
```

Curl the example app through Envoy proxy once again, now sending HAProxy PROXY protocol header at the beginning of the connection with --haproxy-protocol flag:

```shell
curl --verbose --haproxy-protocol --header "Host: www.example.com" http://$GATEWAY_HOST/get
```

You should now expect 200 response status and also see that source IP was preserved in the X-Forwarded-For header.

```shell
*   Trying 172.18.255.202:80...
* Connected to 172.18.255.202 (172.18.255.202) port 80 (#0)
> GET /get HTTP/1.1
> Host: www.example.com
> User-Agent: curl/8.1.2
> Accept: */*
>
< HTTP/1.1 200 OK
< content-type: application/json
< x-content-type-options: nosniff
< date: Mon, 04 Dec 2023 21:11:43 GMT
< content-length: 510
< x-envoy-upstream-service-time: 0
< server: envoy
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
   "curl/8.1.2"
  ],
  "X-Envoy-Expected-Rq-Timeout-Ms": [
   "15000"
  ],
  "X-Envoy-Internal": [
   "true"
  ],
  "X-Forwarded-For": [
   "192.168.255.6"
  ],
  "X-Forwarded-Proto": [
   "http"
  ],
  "X-Request-Id": [
   "290e4b61-44b7-4e5c-a39c-0ec76784e897"
  ]
 },
 "namespace": "default",
 "ingress": "",
 "service": "",
 "pod": "backend-58d58f745-2zwvn"
* Connection #0 to host 172.18.255.202 left intact
}
```

### Configure Number of X-Forwarded-For Trusted Hops

This example configures the number of additional ingress proxy hops from the right side of XFF HTTP headers to trust when determining the origin client's IP address and determines whether or not `x-forwarded-proto` headers will be trusted. Refer to https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_conn_man/headers#x-forwarded-for for details.

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: ClientTrafficPolicy
metadata:
  name: http-xff-num-trusted-hops
  namespace: default
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: Gateway
    name: eg
    namespace: default
  httpConnectionManager:
    xffNumTrustedHops: 2
EOF
```

Verify that ClientTrafficPolicy is Accepted:

```shell
kubectl get clienttrafficpolicies.gateway.envoyproxy.io -n default
```

You should see the policy marked as accepted like this:

```shell
NAME                          STATUS     AGE
http-xff-num-trusted-hops   Accepted   5s
```

Open port-forward to the admin interface port:

```shell
kubectl port-forward deploy/${ENVOY_DEPLOYMENT} -n envoy-gateway-system 19000:19000
```

Curl the admin interface port to fetch the configured value for `xff_num_trusted_hops`:

```shell
curl -s 'http://localhost:19000/config_dump?resource=dynamic_listeners' \
  | jq -r '.configs[0].active_state.listener.default_filter_chain.filters[0].typed_config 
      | with_entries(select(.key == "xff_num_trusted_hops"))'
```

You should expect to see the following:

```json
{
  "xff_num_trusted_hops": 2
}
```

[ClientTrafficPolicy]: ../../api/extension_types#clienttrafficpolicy
[BackendTrafficPolicy]: ../../api/extension_types#backendtrafficpolicy
