---
title: "Response Override"
---

Response Override allows you to override the response from the backend with a custom one. This can be useful for scenarios such as returning a custom 404 page when the requested resource is not found or a custom 500 error message when the backend is failing.

## Installation

Follow the steps from the [Quickstart](../../quickstart) to install Envoy Gateway and the example manifest.
Before proceeding, you should be able to query the example backend using HTTP.

## Testing Response Override

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: response-override
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: backend
  responseOverride:
    - match:
        statusCodes:
          - type: Value
            value: 404
      response:
        contentType: text/plain
        body:
          type: Inline
          inline: "Oops! Your request is not found."
    - match:
        statusCodes:
          - type: Value
            value: 500
          - type: Range
            range:
              start: 501
              end: 511
      response:
        contentType: application/json
        body:
          type: ValueRef
          valueRef:
            group: ""
            kind: ConfigMap
            name: response-override-config
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: response-override-config
data:
  response.body: '{"error": "Internal Server Error"}'
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
  name: response-override
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: backend
  responseOverride:
    - match:
        statusCodes:
          - type: Value
            value: 404
      response:
        contentType: text/plain
        body:
          type: Inline
          inline: "Oops! Your request is not found."
    - match:
        statusCodes:
          - type: Value
            value: 500
          - type: Range
            range:
              start: 501
              end: 511
      response:
        contentType: application/json
        body:
          type: ValueRef
          valueRef:
            group: ""
            kind: ConfigMap
            name: response-override-config
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: response-override-config
data:
  response.body: '{"error": "Internal Server Error"}'
```

{{% /tab %}}
{{< /tabpane >}}

```shell
curl --verbose --header "Host: www.example.com" http://$GATEWAY_HOST/status/404
```

```console
*   Trying 127.0.0.1:80...
* Connected to 172.18.0.200 (172.18.0.200) port 80
> GET /status/404 HTTP/1.1
> Host: www.example.com
> User-Agent: curl/8.5.0
> Accept: */*
>
< HTTP/1.1 404 Not Found
< content-type: text/plain
< content-length: 32
< date: Thu, 07 Nov 2024 09:22:29 GMT
<
* Connection #0 to host 172.18.0.200 left intact
Oops! Your request is not found.
```

```shell
curl --verbose --header "Host: www.example.com" http://$GATEWAY_HOST/status/500
```

```console
*   Trying 127.0.0.1:80...
* Connected to 172.18.0.200 (172.18.0.200) port 80
> GET /status/500 HTTP/1.1
> Host: www.example.com
> User-Agent: curl/8.5.0
> Accept: */*
>
< HTTP/1.1 500 Internal Server Error
< content-type: application/json
< content-length: 34
< date: Thu, 07 Nov 2024 09:23:02 GMT
<
* Connection #0 to host 172.18.0.200 left intact
{"error": "Internal Server Error"}
```