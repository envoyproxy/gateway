---
title: "Dynamic Forward Proxy"
---

Envoy Gateway can be configured as a dynamic forward proxy using the [Backend][] API by setting its type to `DynamicResolver`. This allows Envoy Gateway to act as an HTTP proxy without needing prior knowledge of destination hostnames or IP addresses, while still maintaining its advanced routing and traffic management capabilities.

## Warning

When a Backend resource is configured as `DynamicResolver`, it can route traffic to any destination, effectively exposing all potential endpoints to clients. This can introduce security risks if not properly managed. For example:
* Sending requests with injected credentials to an attacker-controlled endpoint
* Being exploited as part of a botnet to send traffic to a targeted service

Please exercise caution when using this backend type.

## Prerequisites

{{< boilerplate prerequisites >}}

### Enable Backend

Dynamic Forward Proxy relies on the [Backend][] API to route traffic. By default, the Backend API is disabled in Envoy Gateway.
To enable it, follow the instructions in the [Backend Routing][] task.

## Dynamic Forward Proxy

In this example, we will configure Envoy Gateway to act as a dynamic forward proxy using the Backend API. The configuration will include a `Backend` resource of type `DynamicResolver`, which allows Envoy Gateway to resolve and route traffic to any destination.

Note: the TLS configuration in the example is optional. It's only required if you want to use TLS to connect to the backend service. The example uses the system well-known CA certificate to validate the backend service's certificate. You can also use a custom CA certificate by specifying the `caCertificate` field in the `tls` section.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: dynamic-forward-proxy
spec:
  parentRefs:
    - name: eg
  rules:
    - backendRefs:
      - group: gateway.envoyproxy.io
        kind: Backend
        name: backend-dynamic-resolver
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: Backend
metadata:
  name: backend-dynamic-resolver
spec:
  type: DynamicResolver
  tls:
    wellKnownCACertificates: System
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resources to your cluster:

```yaml
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: dynamic-forward-proxy
spec:
  parentRefs:
    - name: eg
  rules:
    - backendRefs:
      - group: gateway.envoyproxy.io
        kind: Backend
        name: backend-dynamic-resolver
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: Backend
metadata:
  name: backend-dynamic-resolver
spec:
  type: DynamicResolver
  tls:
    wellKnownCACertificates: System
```

{{% /tab %}}
{{< /tabpane >}}

Get the Gateway address:

```shell
export GATEWAY_HOST=$(kubectl get gateway/eg -o jsonpath='{.status.addresses[0].value}')
```

Send a request to `gateway.envoyproxy.io` and view the response:

```shell
curl -HHost:gateway.envoyproxy.io http://${GATEWAY_HOST}
```

You can also send a request to any other domain, and Envoy Gateway will resolve the hostname and route the traffic accordingly:

```shell
curl -HHost:httpbin.org http://${GATEWAY_HOST}/get
```

[Backend]: ../../../api/extension_types#backend
[Backend Routing]: ../backend/#enable-backend
