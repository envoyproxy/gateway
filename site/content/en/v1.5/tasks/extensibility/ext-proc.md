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


## Choosing a Deployment Pattern

You can deploy your external processing service in two main ways:  
as a **separate Kubernetes Service** or as a **sidecar container** within the Envoy pod.  
The best choice depends on your needs for latency, resource management, and operational simplicity.

### Pattern 1: Separate Service Deployment

**Pros:**
- Standard Kubernetes approach.
- Independently scalable and deployable from the gateway.

**Cons:**
- Higher network latency between Envoy and the processor, since traffic traverses the pod network.

### Pattern 2: Sidecar Deployment (Localhost ExtProc)

**Pros:**
- Ultra-low latency — communication happens over `localhost`.
- Processor lifecycle is tied to the Envoy pod, simplifying operations.

**Cons:**
- Processor shares CPU and memory with the Envoy container.
- Scales together with the gateway, which may not suit all workloads.


## Pattern 1: Separate Service Deployment

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

Optional: Enable TLS Between Envoy and the gRPC Processor

If your gRPC service uses TLS, create a `BackendTLSPolicy` to configure trust and validation.

```yaml
---
apiVersion: gateway.networking.k8s.io/v1alpha2
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
    hostname: grpc-ext-proc.default.svc.cluster.local
```

## Pattern 2: Sidecar Deployment (Localhost ExtProc)

For latency-sensitive use cases, you can deploy the external processing gRPC service as a sidecar in the same Pod as Envoy.

This allows communication over `localhost`, which avoids pod-to-pod networking overhead.

### Step 1: Enable Extension APIs
Enable the `EnvoyPatchPolicy` and `Backend` APIs in your Envoy Gateway configuration.
This is done by editing the `envoy-gateway-config` ConfigMap in the `envoy-gateway-system` namespace.

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
    gateway:
      controllerName: gateway.envoyproxy.io/gatewayclass-controller
    extensionApis:
      enableEnvoyPatchPolicy: true
      enableBackend: true
```

### Step 2: Add the Sidecar Container with EnvoyProxy
Use an `EnvoyProxy` resource to patch the Envoy Deployment and include your gRPC service container:

```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: gateway
  namespace: envoy-gateway-system
spec:
  provider:
    type: Kubernetes
    kubernetes:
      envoyDeployment:
        patch:
          type: StrategicMerge
          value:
            spec:
              template:
                spec:
                  containers:
                    - name: my-ext-proc-image
                      image: my-ext-image:latest
                      ports:
                        - containerPort: 9000
```

### Step 3: Define the localhost Backend

Create a Backend resource that defines the localhost endpoint:

```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: Backend
metadata:
  name: extproc
  namespace: envoy-gateway-system
spec:
  endpoints:
    - ip:
        address: "127.0.0.1"
        port: 9000
```

### Step 4: Configure the EnvoyExtensionPolicy

Link your route to the localhost backend:

```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyExtensionPolicy
metadata:
  name: add-ext-proc-server
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: myapp
  extProc:
    backendRefs:
      - name: extproc
        kind: Backend
        namespace: envoy-gateway-system
        port: 9000

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

Follow the steps from the [Quickstart](../../quickstart) to uninstall Envoy Gateway and the example manifest. Then delete the resources specific to the deployment pattern you used.

### For the Separate Service Pattern

```shell
kubectl delete -f https://raw.githubusercontent.com/envoyproxy/gateway/latest/examples/kubernetes/ext-proc-grpc-service.yaml
kubectl delete httproute/myapp
kubectl delete envoyextensionpolicy/ext-proc-example
kubectl delete backendtlspolicy/grpc-ext-proc-btls
```

### For the Sidecar Pattern

To clean up the sidecar configuration, remove the patch from your `EnvoyProxy` resource and delete the `Backend` and `EnvoyExtensionPolicy` resources you created:

```shell
kubectl delete envoyextensionpolicy/add-ext-proc-server
kubectl delete backend/extproc -n envoy-gateway-system
kubectl delete envoyproxy/gateway -n envoy-gateway-system
```

If you modified the `envoy-gateway-config` ConfigMap to enable `extensionApis`, you can revert it by restoring the original configuration from the [Quickstart](../../quickstart).

## Next Steps

Checkout the [Developer Guide](../../../contributions/develop) to get involved in the project.

[EnvoyExtensionPolicy]: ../../../api/extension_types#envoyextensionpolicy
[BackendTLSPolicy]: https://gateway-api.sigs.k8s.io/api-types/backendtlspolicy/
[Gateway]: https://gateway-api.sigs.k8s.io/api-types/gateway
[HTTPRoute]: https://gateway-api.sigs.k8s.io/api-types/httproute
