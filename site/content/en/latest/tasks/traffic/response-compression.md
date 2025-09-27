---
title: "Response Compression"
---

Response Compression allows you to compress the response from the backend before sending it to the client. This can be useful for scenarios where the backend sends large responses that can be compressed to reduce the network bandwidth. However, this comes with a trade-off of increased CPU usage on the Envoy side to compress the response.

## Prerequisites

{{< boilerplate prerequisites >}}

## Testing Response Compression

You can enable compression by specifying the compression types in the `BackendTrafficPolicy` resource.
Multiple compression types can be defined within the resource, allowing Envoy Gateway to choose the most appropriate option based on the `Accept-Encoding header` provided by the client..

Envoy Gateway currently supports Brotli and Gzip compression algorithms. Additional compression algorithms may be supported in the future.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: response-compression
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: backend
  compression:
    - type: Brotli
    - type: Gzip
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
  name: response-compression
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: backend
  compression:
    - type: Brotli
    - type: Gzip
```

{{% /tab %}}
{{< /tabpane >}}


To specify the desired compression type, include the `Accept-Encoding` header in requests sent to the Envoy Gateway. The quality value (`q`) in the `Accept-Encoding` header determines the priority of each compression type. Envoy Gateway will select the compression type with the highest `q` value that matches a type configured in the `BackendTrafficPolicy` resource.

```shell
curl --verbose --header "Host: www.example.com" --header "Accept-Encoding: br;q=1.0, gzip;q=0.8" http://$GATEWAY_HOST/
```

```console
*   Trying 172.18.0.200:80...
* Connected to 172.18.0.200 (172.18.0.200) port 80
> GET / HTTP/1.1
> Host: www.example.com
> User-Agent: curl/8.5.0
> Accept: */*
> Accept-Encoding: br;q=1.0, gzip;q=0.8
>
< HTTP/1.1 200 OK
< content-type: application/json
< x-content-type-options: nosniff
< date: Wed, 15 Jan 2025 08:23:42 GMT
< vary: Accept-Encoding
< content-encoding: br
< transfer-encoding: chunked
```

```shell
curl --verbose --header "Host: www.example.com" --header "Accept-Encoding: br;q=0.8, gzip;q=1.0" http://$GATEWAY_HOST/
```

```console
*   Trying 172.18.0.200:80...
* Connected to 172.18.0.200 (172.18.0.200) port 80
> GET / HTTP/1.1
> Host: www.example.com
> User-Agent: curl/8.5.0
> Accept: */*
> Accept-Encoding: br;q=0.8, gzip;q=1.0
>
< HTTP/1.1 200 OK
< content-type: application/json
< x-content-type-options: nosniff
< date: Wed, 15 Jan 2025 08:29:22 GMT
< content-encoding: gzip
< vary: Accept-Encoding
< transfer-encoding: chunked
```

## Disabling Compression

To disable compression for a specific route, set the `compression` field to an empty array in your `BackendTrafficPolicy`:

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: disable-compression
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: backend
  compression: []  # Empty array disables compression
```

This is particularly useful when you have a gateway-level compression policy but need to disable compression for specific routes.