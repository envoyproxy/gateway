---
title: "Direct Response"
---

Direct responses are valuable in cases where you want the gateway itself
to handle certain requests without forwarding them to backend services.
This task shows you how to configure them.

## Installation

Follow the steps from the [Quickstart](../../quickstart) to install Envoy Gateway and the example manifest.
Before proceeding, you should be able to query the example backend using HTTP.

## Testing Direct Response 

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: direct-response
spec:
  parentRefs:
  - name: eg
  hostnames:
  - "www.example.com"    
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /inline
    filters:
    - type: ExtensionRef
      extensionRef:
        group: gateway.envoyproxy.io
        kind: HTTPRouteFilter
        name: direct-response-inline
  - matches:
    - path:
        type: PathPrefix
        value: /value-ref
    filters:
    - type: ExtensionRef
      extensionRef:
        group: gateway.envoyproxy.io
        kind: HTTPRouteFilter
        name: direct-response-value-ref
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: value-ref-response
data:
  response.body: '{"error": "Internal Server Error"}'
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: HTTPRouteFilter
metadata:
  name: direct-response-inline
spec:
  directResponse:
    contentType: text/plain
    statusCode: 503
    body:
      type: Inline
      inline: "Oops! Your request is not found."
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: HTTPRouteFilter
metadata:
  name: direct-response-value-ref
spec:
  directResponse:
    contentType: application/json
    statusCode: 500
    body:
      type: ValueRef
      valueRef:
        group: ""
        kind: ConfigMap
        name: value-ref-response
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
  name: direct-response
spec:
  parentRefs:
  - name: eg
  hostnames:
  - "www.example.com"    
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /inline
    filters:
    - type: ExtensionRef
      extensionRef:
        group: gateway.envoyproxy.io
        kind: HTTPRouteFilter
        name: direct-response-inline
  - matches:
    - path:
        type: PathPrefix
        value: /value-ref
    filters:
    - type: ExtensionRef
      extensionRef:
        group: gateway.envoyproxy.io
        kind: HTTPRouteFilter
        name: direct-response-value-ref
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: value-ref-response
data:
  response.body: '{"error": "Internal Server Error"}'
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: HTTPRouteFilter
metadata:
  name: direct-response-inline
spec:
  directResponse:
    contentType: text/plain
    statusCode: 503
    body:
      type: Inline
      inline: "Oops! Your request is not found."
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: HTTPRouteFilter
metadata:
  name: direct-response-value-ref
spec:
  directResponse:
    contentType: application/json
    statusCode: 500
    body:
      type: ValueRef
      valueRef:
        group: ""
        kind: ConfigMap
        name: value-ref-response
```

{{% /tab %}}
{{< /tabpane >}}

```shell
curl --verbose --header "Host: www.example.com" http://$GATEWAY_HOST/inline
```

```console
*   Trying 127.0.0.1:80...
* Connected to 127.0.0.1 (127.0.0.1) port 80
> GET /inline HTTP/1.1
> Host: www.example.com
> User-Agent: curl/8.4.0
> Accept: */*
> 
< HTTP/1.1 503 Service Unavailable
< content-type: text/plain
< content-length: 32
< date: Sat, 02 Nov 2024 00:35:48 GMT
< 
* Connection #0 to host 127.0.0.1 left intact
Oops! Your request is not found.
```

```shell
curl --verbose --header "Host: www.example.com" http://$GATEWAY_HOST/value-ref
```

```console
*   Trying 127.0.0.1:80...
* Connected to 127.0.0.1 (127.0.0.1) port 80
> GET /value-ref HTTP/1.1
> Host: www.example.com
> User-Agent: curl/8.4.0
> Accept: */*
> 
< HTTP/1.1 500 Internal Server Error
< content-type: application/json
< content-length: 34
< date: Sat, 02 Nov 2024 00:35:55 GMT
< 
* Connection #0 to host 127.0.0.1 left intact
{"error": "Internal Server Error"}
```
