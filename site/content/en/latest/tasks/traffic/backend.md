---
title: "Backend Routing"
---

Envoy Gateway supports routing to native K8s resources such as `Service` and `ServiceImport`. The `Backend` API is a custom Envoy Gateway [extension resource][] that can used in Gateway-API [BackendObjectReference][].

## Motivation
The Backend API was added to support several use cases:
- Allowing users to integrate Envoy with services (Ext Auth, Rate Limit, ALS, ...) using Unix Domain Sockets, which are currently not supported by K8s.
- Simplify [routing to cluster-external backends][], which currently requires users to maintain both K8s `Service` and `EndpointSlice` resources.

## Warning

Similar to the K8s EndpointSlice API, the Backend API can be misused to allow traffic to be sent to otherwise restricted destinations, as described in [CVE-2021-25740][].
A Backend resource can be used to:
- Expose a Service or Pod that should not be accessible
- Reference a Service or Pod by a Route without appropriate Reference Grants
- Expose the Envoy Proxy localhost (including the Envoy admin endpoint)

For these reasons, the Backend API is disabled by default in Envoy Gateway configuration. Envoy Gateway admins are advised to follow [upstream recommendations][] and restrict access to the Backend API using K8s RBAC.

## Restrictions

The Backend API is currently supported only in the following BackendReferences:
- [HTTPRoute]: IP and FQDN endpoints
- [TLSRoute]: IP and FQDN endpoints
- [Envoy Extension Policy] (ExtProc): IP, FQDN and unix domain socket endpoints
- [Security Policy]: IP and FQDN endpoints for the OIDC providers

The Backend API supports attachment the following policies:
- [Backend TLS Policy][]

Certain restrictions apply on the value of hostnames and addresses. For example, the loopback IP address range and the localhost hostname are forbidden.

Envoy Gateway does not manage the lifecycle of unix domain sockets referenced by the Backend resource. Envoy Gateway admins are responsible for creating and mounting the sockets into the envoy proxy pod. The latter can be achieved by patching the envoy deployment using the [EnvoyProxy][] resource.

## Quickstart

### Prerequisites

{{< boilerplate prerequisites >}}

### Enable Backend

* By default [Backend][] is disabled. Lets enable it in the [EnvoyGateway][] startup configuration

* The default installation of Envoy Gateway installs a default [EnvoyGateway][] configuration and attaches it
  using a `ConfigMap`. In the next step, we will update this resource to enable Backend.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
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
      enableBackend: true
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

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
      enableBackend: true
```

{{% /tab %}}
{{< /tabpane >}}

{{< boilerplate rollout-envoy-gateway >}}

## Testing

### Route to External Backend

* In the following example, we will create a `Backend` resource that routes to httpbin.org:80 and a `HTTPRoute` that references this backend.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: backend
spec:
  parentRefs:
    - name: eg
  hostnames:
    - "www.example.com"
  rules:
    - backendRefs:
        - group: gateway.envoyproxy.io
          kind: Backend
          name: httpbin
      matches:
        - path:
            type: PathPrefix
            value: /
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: Backend
metadata:
  name: httpbin
  namespace: default
spec:
  endpoints:
    - fqdn:
        hostname: httpbin.org
        port: 80
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
  name: backend
spec:
  parentRefs:
    - name: eg
  hostnames:
    - "www.example.com"
  rules:
    - backendRefs:
        - group: gateway.envoyproxy.io
          kind: Backend
          name: httpbin
      matches:
        - path:
            type: PathPrefix
            value: /
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: Backend
metadata:
  name: httpbin
  namespace: default
spec:
  endpoints:
    - fqdn:
        hostname: httpbin.org
        port: 80

```

{{% /tab %}}
{{< /tabpane >}}

Get the Gateway address:

```shell
export GATEWAY_HOST=$(kubectl get gateway/eg -o jsonpath='{.status.addresses[0].value}')
```

Send a request and view the response:

```shell
curl -I -HHost:www.example.com http://${GATEWAY_HOST}/headers
```

[Backend]: ../../../api/extension_types#backend
[routing to cluster-external backends]: ./../../tasks/traffic/routing-outside-kubernetes.md
[BackendObjectReference]: https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.BackendObjectReference
[extension resource]: https://gateway-api.sigs.k8s.io/guides/migrating-from-ingress/#approach-to-extensibility
[CVE-2021-25740]: https://nvd.nist.gov/vuln/detail/CVE-2021-25740
[upstream recommendations]: https://github.com/kubernetes/kubernetes/issues/103675
[HTTPRoute]: https://gateway-api.sigs.k8s.io/api-types/httproute
[TLSRoute]: https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.TLSRoute
[Envoy Extension Policy]: ../../../api/extension_types#envoyextensionpolicy
[Security Policy]: ../../../api/extension_types#oidcprovider
[Backend TLS Policy]: https://gateway-api.sigs.k8s.io/api-types/backendtlspolicy/
[EnvoyProxy]: ../../../api/extension_types#envoyproxy
[EnvoyGateway]: ../../../api/extension_types#envoygateway
