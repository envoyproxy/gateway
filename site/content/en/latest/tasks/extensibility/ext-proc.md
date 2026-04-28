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

### Observability

Each `extProc` entry exposes two optional fields for observability:

- **`name`**: assigns a friendly identifier to the ext-proc instance. Once set, you can reference it in your `EnvoyProxy` access log format strings using the `%EG_EXT_PROC_FILTER_STATE(name:attribute)%` operator. Envoy Gateway resolves this at xDS translation time — Envoy only ever sees the standard `%FILTER_STATE(...)%` form.

- **`statPrefix`**: controls the Envoy stat prefix for this filter instance (e.g. `ext_proc.<statPrefix>.streams_started`). When unset, defaults to `name` if `name` is set, otherwise Envoy uses its own default. Use a shared `statPrefix` across deployments to aggregate metrics, or distinct values to isolate per-deployment counters.

#### Access log operator

Assign a `name` to an ext-proc entry in your `EnvoyExtensionPolicy`. Names consist of lowercase alphanumeric characters and hyphens:

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyExtensionPolicy
metadata:
  name: ext-proc-example
  namespace: default
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: myapp
  extProc:
  - name: auth-service
    backendRefs:
    - name: grpc-ext-proc
      port: 9002
EOF
```

Use the `%EG_EXT_PROC_FILTER_STATE(name:attribute)%` operator in any text or JSON access log format string in your `EnvoyProxy`. Available attributes depend on what the ext-proc service writes to filter state; common ones include `latency_ns` and `grpc_status_code`. The operator also passes through Envoy's optional format arguments verbatim (serialization type and max length):

```
%EG_EXT_PROC_FILTER_STATE(auth-service:latency_ns:TYPED:64)%
```

Text format example:

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: ext-proc-accesslog
  namespace: envoy-gateway-system
spec:
  telemetry:
    accessLog:
      settings:
        - format:
            type: Text
            text: |
              [%START_TIME%] %REQ(:METHOD)% %RESPONSE_CODE% auth_latency=%EG_EXT_PROC_FILTER_STATE(auth-service:latency_ns)%
          sinks:
            - type: File
              file:
                path: /dev/stdout
EOF
```

JSON format is also supported:

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: ext-proc-accesslog-json
  namespace: envoy-gateway-system
spec:
  telemetry:
    accessLog:
      settings:
        - format:
            type: JSON
            json:
              method: "%REQ(:METHOD)%"
              response_code: "%RESPONSE_CODE%"
              auth_latency_ns: "%EG_EXT_PROC_FILTER_STATE(auth-service:latency_ns)%"
              auth_grpc_status: "%EG_EXT_PROC_FILTER_STATE(auth-service:grpc_status_code)%"
          sinks:
            - type: File
              file:
                path: /dev/stdout
EOF
```

A single format string can reference ext-proc instances from multiple EEPs — each operator resolves independently by name.

#### Name conflicts and shared names

Names are not validated for uniqueness across policies. When multiple `EnvoyExtensionPolicy` resources on the same listener claim the same name, the **most-specific target scope wins**: route-rule policies before route policies, route policies before listener policies, listener policies before gateway policies. Within the same scope, the **oldest policy** (earliest creation timestamp) wins. Losers receive a `Warning` condition with reason `AmbiguousDefinition` and their operator resolves to `[EG_UNRESOLVED:name]`. The routes themselves still execute their own ext-proc filters normally.

The winner's filter identity becomes the resolution target for `%EG_EXT_PROC_FILTER_STATE(name:...)%` **across the entire listener** — not just for routes where the winning policy's ext-proc is active. On routes served by a losing policy's ext-proc, the operator will expand to the winner's filter-state key, which will be absent and produce an empty value.

For the grouping pattern to work correctly, routes must reside on **separate listeners** — each listener gets its own HCM and isolated ext-proc filter chain. When routes share a listener/port, they share the same HCM and operator resolution is not per-route.

> **MergeGateways (non-TLS):** When [`MergeGateways`](https://gateway.envoyproxy.io/docs/api/extension_types/#mergegatewaysconfig) is enabled, non-TLS listeners on the same port share a single HCM. Name collisions across gateways are resolved silently with no warning — the oldest merged gateway wins. Use distinct names across merged gateways.

A common safe pattern is duplicating the same ext-proc policy across namespaces, each on a dedicated listener:

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: gateway-team-a
  namespace: namespace-a
spec:
  gatewayClassName: eg
  listeners:
  - name: http
    protocol: HTTP
    port: 8080
---
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: gateway-team-b
  namespace: namespace-b
spec:
  gatewayClassName: eg
  listeners:
  - name: http
    protocol: HTTP
    port: 8081
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: route-a
  namespace: namespace-a
spec:
  parentRefs:
  - name: gateway-team-a
    namespace: namespace-a
  rules:
  - backendRefs:
    - name: service-a
      port: 8080
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: route-b
  namespace: namespace-b
spec:
  parentRefs:
  - name: gateway-team-b
    namespace: namespace-b
  rules:
  - backendRefs:
    - name: service-b
      port: 8080
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyExtensionPolicy
metadata:
  name: auth-policy
  namespace: namespace-a
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: route-a
  extProc:
  - name: auth-service
    backendRefs:
    - name: grpc-auth
      namespace: namespace-a
      port: 9002
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyExtensionPolicy
metadata:
  name: auth-policy
  namespace: namespace-b
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: route-b
  extProc:
  - name: auth-service
    backendRefs:
    - name: grpc-auth
      namespace: namespace-b
      port: 9002
EOF
```

A single access log format covers both gateways because each listener resolves the operator against its own isolated filter chain:

```yaml
spec:
  telemetry:
    accessLog:
      settings:
        - format:
            type: Text
            text: |
              [%START_TIME%] %REQ(:METHOD)% %RESPONSE_CODE% auth_latency=%EG_EXT_PROC_FILTER_STATE(auth-service:latency_ns)%
          sinks:
            - type: File
              file:
                path: /dev/stdout
```

For a full explanation of the design, see the [Ext-Proc Observability design document](/community/design/ext-proc-observability).

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

Checkout the [Developer Guide](/community/develop) to get involved in the project.

[EnvoyExtensionPolicy]: ../../../api/extension_types#envoyextensionpolicy
[BackendTLSPolicy]: https://gateway-api.sigs.k8s.io/api-types/backendtlspolicy/
[Gateway]: https://gateway-api.sigs.k8s.io/api-types/gateway
[HTTPRoute]: https://gateway-api.sigs.k8s.io/api-types/httproute
