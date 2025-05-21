---
title: "Client Traffic Policy"
---

This task explains the usage of the [ClientTrafficPolicy][] API.

## Introduction

The [ClientTrafficPolicy][] API allows system administrators to configure
the behavior for how the Envoy Proxy server behaves with downstream clients.

## Motivation

This API was added as a new policy attachment resource that can be applied to Gateway resources and it is meant to hold settings for configuring behavior of the connection between the downstream client and Envoy Proxy listener. It is the counterpart to the [BackendTrafficPolicy][] API resource.

## Quickstart

### Prerequisites

{{< boilerplate prerequisites >}}

### Support TCP keepalive for downstream client

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: ClientTrafficPolicy
metadata:
  name: enable-tcp-keepalive-policy
  namespace: default
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: Gateway
      name: eg
  tcpKeepalive:
    idleTime: 20m
    interval: 60s
    probes: 3
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: ClientTrafficPolicy
metadata:
  name: enable-tcp-keepalive-policy
  namespace: default
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: Gateway
      name: eg
  tcpKeepalive:
    idleTime: 20m
    interval: 60s
    probes: 3
```

{{% /tab %}}
{{< /tabpane >}}

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

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: ClientTrafficPolicy
metadata:
  name: enable-proxy-protocol-policy
  namespace: default
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: Gateway
      name: eg
  enableProxyProtocol: true
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: ClientTrafficPolicy
metadata:
  name: enable-proxy-protocol-policy
  namespace: default
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: Gateway
      name: eg
  enableProxyProtocol: true
```

{{% /tab %}}
{{< /tabpane >}}

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

### Configure Client IP Detection

This example configures the number of additional ingress proxy hops from the right side of XFF HTTP headers to trust when determining the origin client's IP address and determines whether or not `x-forwarded-proto` headers will be trusted. Refer to https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_conn_man/headers#x-forwarded-for for details.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: ClientTrafficPolicy
metadata:
  name: http-client-ip-detection
  namespace: default
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: Gateway
      name: eg
  clientIPDetection:
    xForwardedFor:
      numTrustedHops: 2
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: ClientTrafficPolicy
metadata:
  name: http-client-ip-detection
  namespace: default
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: Gateway
      name: eg
  clientIPDetection:
    xForwardedFor:
      numTrustedHops: 2
```

{{% /tab %}}
{{< /tabpane >}}

Verify that ClientTrafficPolicy is Accepted:

```shell
kubectl get clienttrafficpolicies.gateway.envoyproxy.io -n default
```

You should see the policy marked as accepted like this:

```shell
NAME                          STATUS     AGE
http-client-ip-detection   Accepted   5s
```

Open port-forward to the admin interface port:

```shell
kubectl port-forward deploy/${ENVOY_DEPLOYMENT} -n envoy-gateway-system 19000:19000
```

Curl the admin interface port to fetch the configured value for `xff_num_trusted_hops`:

```shell
curl -s 'http://localhost:19000/config_dump?resource=dynamic_listeners' \
  | jq -r '.configs[1].active_state.listener.default_filter_chain.filters[0].typed_config
    | {use_remote_address, original_ip_detection_extensions}'
```

You should expect to see the following:

```json
{
  "use_remote_address": false,
  "original_ip_detection_extensions": [
    {
      "name": "envoy.extensions.http.original_ip_detection.xff",
      "typed_config": {
        "@type": "type.googleapis.com/envoy.extensions.http.original_ip_detection.xff.v3.XffConfig",
        "xff_num_trusted_hops": 1,
        "skip_xff_append": false
      }
    }
  ]
}
```

Curl the example app through Envoy proxy:

```shell
curl -v http://$GATEWAY_HOST/get \
  -H "Host: www.example.com" \
  -H "X-Forwarded-Proto: https" \
  -H "X-Forwarded-For: 1.1.1.1,2.2.2.2"
```

You should expect 200 response status, see that `X-Forwarded-Proto` was preserved and `X-Envoy-External-Address` was set to the leftmost address in the `X-Forwarded-For` header:

```shell
*   Trying [::1]:8888...
* Connected to localhost (::1) port 8888
> GET /get HTTP/1.1
> Host: www.example.com
> User-Agent: curl/8.4.0
> Accept: */*
> X-Forwarded-Proto: https
> X-Forwarded-For: 1.1.1.1,2.2.2.2
> 
Handling connection for 8888
< HTTP/1.1 200 OK
< content-type: application/json
< x-content-type-options: nosniff
< date: Tue, 30 Jan 2024 15:19:22 GMT
< content-length: 535
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
   "curl/8.4.0"
  ],
  "X-Envoy-Expected-Rq-Timeout-Ms": [
   "15000"
  ],
  "X-Envoy-External-Address": [
   "1.1.1.1"
  ],
  "X-Forwarded-For": [
   "1.1.1.1,2.2.2.2,10.244.0.9"
  ],
  "X-Forwarded-Proto": [
   "https"
  ],
  "X-Request-Id": [
   "53ccfad7-1899-40fa-9322-ddb833aa1ac3"
  ]
 },
 "namespace": "default",
 "ingress": "",
 "service": "",
 "pod": "backend-58d58f745-8psnc"
* Connection #0 to host localhost left intact
}
```

### Enable HTTP Request Received Timeout

This feature allows you to limit the time taken by the Envoy Proxy fleet to receive the entire request from the client, which is useful in preventing certain clients from consuming too much memory in Envoy
This example configures the HTTP request timeout for the client, please check out the details [here](https://www.envoyproxy.io/docs/envoy/latest/faq/configuration/timeouts#stream-timeouts). 

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: ClientTrafficPolicy
metadata:
  name: client-timeout
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: Gateway
      name: eg
  timeout:
    http:
      requestReceivedTimeout: 2s
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: ClientTrafficPolicy
metadata:
  name: client-timeout
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: Gateway
      name: eg
  timeout:
    http:
      requestReceivedTimeout: 2s
```

{{% /tab %}}
{{< /tabpane >}}

Curl the example app through Envoy proxy:

```shell
curl -v http://$GATEWAY_HOST/get \
  -H "Host: www.example.com" \
  -H "Content-Length: 10000"
```

You should expect `428` response status after 2s:

```shell
curl -v http://$GATEWAY_HOST/get \
  -H "Host: www.example.com" \
  -H "Content-Length: 10000"
*   Trying 172.18.255.200:80...
* Connected to 172.18.255.200 (172.18.255.200) port 80
> GET /get HTTP/1.1
> Host: www.example.com
> User-Agent: curl/8.4.0
> Accept: */*
> Content-Length: 10000
>
< HTTP/1.1 408 Request Timeout
< content-length: 15
< content-type: text/plain
< date: Tue, 27 Feb 2024 07:38:27 GMT
< connection: close
<
* Closing connection
request timeout
```

### Configure Client HTTP Idle Timeout

The idle timeout is defined as the period in which there are no active requests. When the idle timeout is reached the connection will be closed.
For more details see [here](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/protocol.proto#envoy-v3-api-field-config-core-v3-httpprotocoloptions-idle-timeout:~:text=...%7D%0A%7D-,idle_timeout,-(Duration)%20The).

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: ClientTrafficPolicy
metadata:
  name: client-timeout
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: Gateway
      name: eg
  timeout:
    http:
      idleTimeout: 5s
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: ClientTrafficPolicy
metadata:
  name: client-timeout
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: Gateway
      name: eg
  timeout:
    http:
      idleTimeout: 5s
```

{{% /tab %}}
{{< /tabpane >}}

Curl the example app through Envoy proxy:

```shell
openssl s_client -crlf -connect $GATEWAY_HOST:443
```

You should expect the connection to be closed after 5s.

You can also check the number of connections closed due to idle timeout by using the following query:

```shell
envoy_http_downstream_cx_idle_timeout{envoy_http_conn_manager_prefix="<name of connection manager>"} 
```

The number of connections closed due to idle timeout should be increased by 1.


### Configure Downstream Per Connection Buffer Limit

This feature allows you to set a soft limit on size of the listener’s new connection read and write buffers.
The size is configured using the `resource.Quantity` format, see examples [here](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#meaning-of-memory).

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: ClientTrafficPolicy
metadata:
  name: client-timeout
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: Gateway
      name: eg
  connection:
    bufferLimit: 1024
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: ClientTrafficPolicy
metadata:
  name: client-timeout
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: Gateway
      name: eg
  connection:
    bufferLimit: 1024
```

{{% /tab %}}
{{< /tabpane >}}

[ClientTrafficPolicy]: ../../../api/extension_types#clienttrafficpolicy
[BackendTrafficPolicy]: ../../../api/extension_types#backendtrafficpolicy
