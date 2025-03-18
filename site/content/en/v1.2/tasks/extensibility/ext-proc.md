---
title: "External Processing"
---

This task provides instructions for configuring external processing.

External processing calls an external gRPC service to process HTTP requests and responses. 
The external processing service can inspect and mutate requests and responses.  

Envoy Gateway introduces a new CRD called [EnvoyExtensionPolicy][] that allows the user to configure external processing.
This instantiated resource can be linked to a [Gateway][Gateway] and [HTTPRoute][HTTPRoute] resource.

## Prerequisites

{{< boilerplate prerequisites >}}

## GRPC External Processing Service

### Installation

Install a demo GRPC service that will be used as the external processing service:

```shell
kubectl apply -f https://raw.githubusercontent.com/envoyproxy/gateway/latest/examples/kubernetes/ext-proc-grpc-service.yaml
```

Create a new HTTPRoute resource to route traffic on the path `/myapp` to the backend service.  

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: myapp
spec:
  parentRefs:
  - name: eg
  hostnames:
  - "www.example.com"
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /myapp
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
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: myapp
spec:
  parentRefs:
  - name: eg
  hostnames:
  - "www.example.com"
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /myapp
    backendRefs:
    - name: backend
      port: 3000   
```

{{% /tab %}}
{{< /tabpane >}}

Verify the HTTPRoute status:

```shell
kubectl get httproute/myapp -o yaml
```

### Configuration

Create a new EnvoyExtensionPolicy resource to configure the external processing service. This EnvoyExtensionPolicy targets the HTTPRoute
"myApp" created in the previous step. It calls the GRPC external processing service "grpc-ext-proc" on port 9002 for
processing. 

By default, requests and responses are not sent to the external processor. The `processingMode` struct is used to define what should be sent to the external processor.
In this example, we configure the following processing modes:
* The empty `request` field configures envoy to send request headers to the external processor.
* The `response` field includes configuration for body processing. As a result, response headers are sent to the external processor. Additionally, the response body is streamed to the external processor.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyExtensionPolicy
metadata:
  name: ext-proc-example
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: myapp
  extProc:
  - backendRefs:
    - name: grpc-ext-proc
      port: 9002
    processingMode:
      request: {}
      response: 
        body: Streamed 
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyExtensionPolicy
metadata:
  name: ext-proc-example
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: myapp
  extProc:
    - backendRefs:
        - name: grpc-ext-proc
          port: 9002
      processingMode:
        request: {}
        response: 
          body: Streamed
```

{{% /tab %}}
{{< /tabpane >}}

Verify the Envoy Extension Policy configuration:

```shell
kubectl get envoyextensionpolicy/ext-proc-example -o yaml
```


Because the gRPC external processing service is enabled with TLS, a [BackendTLSPolicy][] needs to be created to configure
the communication between the Envoy proxy and the gRPC auth service.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1alpha3
kind: BackendTLSPolicy
metadata:
  name: grpc-ext-proc-btls
spec:
  targetRefs:
  - group: ''
    kind: Service
    name: grpc-ext-proc
  validation:
    caCertificateRefs:
    - name: grpc-ext-proc-ca
      group: ''
      kind: ConfigMap
    hostname: grpc-ext-proc.envoygateway
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.networking.k8s.io/v1alpha3
kind: BackendTLSPolicy
metadata:
  name: grpc-ext-proc-btls
spec:
  targetRefs:
    - group: ''
      kind: Service
      name: grpc-ext-proc
  validation:
    caCertificateRefs:
      - name: grpc-ext-proc-ca
        group: ''
        kind: ConfigMap
    hostname: grpc-ext-proc.envoygateway
```

{{% /tab %}}
{{< /tabpane >}}

Verify the BackendTLSPolicy configuration:

```shell
kubectl get backendtlspolicy/grpc-ext-proc-btls -o yaml
```

### Testing

Ensure the `GATEWAY_HOST` environment variable from the [Quickstart](../../quickstart) is set. If not, follow the
Quickstart instructions to set the variable.

```shell
echo $GATEWAY_HOST
```

Send a request to the backend service without `Authentication` header:

```shell
curl -v -H "Host: www.example.com" "http://${GATEWAY_HOST}/myapp"
```

You should see that the external processor added headers:
- `x-request-ext-processed` - this header was added before the request was forwarded to the backend
- `x-response-ext-processed`-  this header was added before the response was returned to the client


```
curl -v -H "Host: www.example.com"  http://localhost:10080/myapp
[...]
< HTTP/1.1 200 OK
< content-type: application/json
< x-content-type-options: nosniff
< date: Fri, 14 Jun 2024 19:30:40 GMT
< content-length: 502
< x-response-ext-processed: true
<
{
 "path": "/myapp",
 "host": "www.example.com",
 "method": "GET",
 "proto": "HTTP/1.1",
 "headers": {
[...] 
  "X-Request-Ext-Processed": [
   "true"
  ],
[...]
 }
```

## Clean-Up

Follow the steps from the [Quickstart](../../quickstart) to uninstall Envoy Gateway and the example manifest.

Delete the demo auth services, HTTPRoute, EnvoyExtensionPolicy and BackendTLSPolicy:

```shell
kubectl delete -f https://raw.githubusercontent.com/envoyproxy/gateway/latest/examples/kubernetes/ext-proc-grpc-service.yaml
kubectl delete httproute/myapp
kubectl delete envoyextensionpolicy/ext-proc-example
kubectl delete backendtlspolicy/grpc-ext-proc-btls
```

## Next Steps

Checkout the [Developer Guide](../../../contributions/develop) to get involved in the project.

[EnvoyExtensionPolicy]: ../../../api/extension_types#envoyextensionpolicy
[BackendTLSPolicy]: https://gateway-api.sigs.k8s.io/api-types/backendtlspolicy/
[Gateway]: https://gateway-api.sigs.k8s.io/api-types/gateway
[HTTPRoute]: https://gateway-api.sigs.k8s.io/api-types/httproute
