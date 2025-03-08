+++
title = "Configuruation Issues"
+++

## Overview
After configuring and applying resources, you might find that Envoy Gateway does not behave as expected. This guide helps troubleshoot configuration issues.

Many **syntax errors and simple semantic issues** are caught during resource validation by the **Kubernetes API Server**, which rejects invalid resources. However, more complex configuration issues require additional debugging.

## Prerequisites

{{< boilerplate prerequisites >}}

To demonstrate debugging techniques, let’s apply an *intentionally incorrect* HTTPRoute configuration with a *non-existent* backend.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: backend
  namespace: default
spec:
  hostnames:
  - www.example.com
  parentRefs:
  - group: gateway.networking.k8s.io
    kind: Gateway
    name: eg
  rules:
  - backendRefs:
    - group: ""
      kind: Service
      name: backend-does-not-exist
      port: 3000
      weight: 1
    matches:
    - path:
        type: PathPrefix
        value: /
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
  name: backend
  namespace: default
spec:
  hostnames:
  - www.example.com
  parentRefs:
  - group: gateway.networking.k8s.io
    kind: Gateway
    name: eg
  rules:
  - backendRefs:
    - group: ""
      kind: Service
      name: backend-does-not-exist
      port: 3000
      weight: 1
    matches:
    - path:
        type: PathPrefix
        value: /
```

{{% /tab %}}
{{< /tabpane >}}

### Checking Resource Status
The `status` field in your resource definitions is your primary troubleshooting tool. It provides insights into whether a resource has been **Accepted** or not, along with the reason for rejection.

#### Using kubectl
This example below shows why an HTTPRoute has been **Accepted** but the **ResolvedRefs** condition is `false` because the backend does not exist. 
**Note**: Almost all resources have a `status` field, so be sure to check them all.

```shell
kubectl get httproute/backend -o yaml
```

```console
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: backend
  namespace: default
spec:
  hostnames:
  - www.example.com
  parentRefs:
  - group: gateway.networking.k8s.io
    kind: Gateway
    name: eg
  rules:
  - backendRefs:
    - group: ""
      kind: Service
      name: backend-does-not-exist
      port: 3000
      weight: 1
    matches:
    - path:
        type: PathPrefix
        value: /
status:
  parents:
  - conditions:
    - lastTransitionTime: "2025-03-08T02:07:53Z"
      message: Route is accepted
      observedGeneration: 3
      reason: Accepted
      status: "True"
      type: Accepted
    - lastTransitionTime: "2025-03-08T02:07:53Z"
      message: Service default/backend-does-not-exist not found
      observedGeneration: 3
      reason: BackendNotFound
      status: "False"
      type: ResolvedRefs
    controllerName: gateway.envoyproxy.io/gatewayclass-controller
    parentRef:
      group: gateway.networking.k8s.io
      kind: Gateway
      name: eg
```

#### Using egctl
The egctl CLI tool can quickly fetch the status of multiple resources at once:

```shell
egctl x status all -A
```

```console
NAME              TYPE       STATUS    REASON
gatewayclass/eg   Accepted   True      Accepted

NAMESPACE   NAME         TYPE         STATUS    REASON
default     gateway/eg   Programmed   True      Programmed
                         Accepted     True      Accepted

NAMESPACE   NAME                TYPE           STATUS    REASON
default     httproute/backend   ResolvedRefs   False     BackendNotFound
                                Accepted       True      Accepted
```

Follow the instructions [here](./../install/install-egctl.md) to install `egctl`.

#### Using kube-state-metrics
For large-scale deployments, kube-state-metrics can help monitor the status of Envoy Gateway resources. For more details, refer to the [Observability Guide](./../tasks/observability/gateway-api-metrics.md).

### Traffic
If a configuration is not accepted, Envoy Gateway assigns a `direct_response` to the affected route, causing clients to receive an **HTTP 500 error**.

```shell
curl --verbose --header "Host: www.example.com" http://$GATEWAY_HOST/get
```

```console
*   Trying 127.0.0.1:80...
* Connected to 127.0.0.1 (127.0.0.1) port 80
> GET /get HTTP/1.1
> Host: www.example.com
> User-Agent: curl/8.7.1
> Accept: */*
> 
* Request completely sent off
< HTTP/1.1 500 Internal Server Error
< date: Sat, 08 Mar 2025 02:25:32 GMT
< content-length: 0
< 
* Connection #0 to host 127.0.0.1 left intact
```

If you inspect the access logs in the pod logs:

```shell
kubectl logs -n envoy-gateway-system -l gateway.envoyproxy.io/owning-gateway-namespace=default,gateway.envoyproxy.io/owning-gateway-name=eg -c envoy | grep start_time | jq
```

```console
{
  ":authority": "www.example.com",
  "bytes_received": 0,
  "bytes_sent": 0,
  "connection_termination_details": null,
  "downstream_local_address": "10.1.21.65:10080",
  "downstream_remote_address": "192.168.65.4:45050",
  "duration": 0,
  "method": "GET",
  "protocol": "HTTP/1.1",
  "requested_server_name": null,
  "response_code": 500,
  "response_code_details": "direct_response",
  "response_flags": "-",
  "route_name": "httproute/default/backend/rule/0/match/0/www_example_com",
  "start_time": "2025-03-08T02:25:32.588Z",
  "upstream_cluster": null,
  "upstream_host": null,
  "upstream_local_address": null,
  "upstream_transport_failure_reason": null,
  "user-agent": "curl/8.7.1",
  "x-envoy-origin-path": "/get",
  "x-envoy-upstream-service-time": null,
  "x-forwarded-for": "192.168.65.4",
  "x-request-id": "54c4c2f9-d209-4aa0-8870-342bc6622d1a"
}
```

and find the following entries

```console
"response_code": "500",
"response_code_details": "direct_response"
```
this likely indicates a configuration issue. Review the relevant [resource status](#checking-resource-status) for resolution steps.
