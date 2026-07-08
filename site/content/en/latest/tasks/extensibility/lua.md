---
title: "Lua Extensions"
---

This task provides instructions for extending Envoy Gateway with Lua extensions.

Lua extensions allow you to extend the functionality of Envoy Gateway by running custom code against HTTP requests and responses,
without modifying the Envoy Gateway binary. These comparatively light-weight extensions are written in the Lua scripting language using APIs defined [here](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/lua_filter#stream-handle-api).

Envoy Gateway allows the user to configure Lua extensions using the [EnvoyExtensionPolicy][] CRD.
This instantiated resource can be linked to a [Gateway][Gateway] or [HTTPRoute][HTTPRoute] resource. If linked to both, the resource linked to the route takes precedence over those linked to Gateway.

{{% alert title="Warning" color="warning" %}}
Lua scripts execute inside the Envoy proxy process without strong sandboxing. While Envoy Gateway
sanitizes scripts and restricts the available Lua API surface, the Lua runtime is inherently less
isolated than a dedicated extension process.

When enabling Lua extensions, admins should take additional measures to reduce risk, including:
* Using K8s [RBAC][] to restrict who can create or modify `EnvoyExtensionPolicy` resources with Lua scripts.
* Using [AdmissionControl][] tools (e.g. OPA Gatekeeper, Kyverno) to validate and review Lua scripts before they are admitted.
* Auditing `EnvoyExtensionPolicy` resources periodically, and enabling [AuditLog][] for API server operations on these resources.
* Not enabling Lua when it is not needed by omitting the `enableLua` field or setting it to `false` in the [EnvoyGateway][] configuration.
{{% /alert %}}

## Enable Lua

Lua extensions are disabled by default. To enable Lua, set `enableLua: true` in the [EnvoyGateway][] configuration:

```yaml
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
      enableLua: true
```

## Prerequisites

{{< boilerplate prerequisites >}}

## Configuration

Envoy Gateway supports Lua in [EnvoyExtensionPolicy][] in two modes:
* Inline Lua: The extension defines the Lua script as an inline string.
* ValueRef Lua: The extension points to an in-cluster ConfigMap resource that contains the Lua script in it's data.

The following example demonstrates how to configure an [EnvoyExtensionPolicy][] to attach a Lua extension to a [HTTPRoute][HTTPRoute].
This Lua extension adds a custom header `x-lua-custom: FOO` to the response.

### Lua Extension - Inline

This [EnvoyExtensionPolicy][] configuration defines the Lua script as an inline string.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyExtensionPolicy
metadata:
  name: lua-inline-test
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: backend
  lua:
    - type: Inline
      inline: |
        function envoy_on_response(response_handle)
          response_handle:headers():add("x-lua-custom", "FOO")
        end
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
  name: lua-inline-test
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: backend
  lua:
    - type: Inline
      inline: |
        function envoy_on_response(response_handle)
          response_handle:headers():add("x-lua-custom", "FOO")
        end
```

{{% /tab %}}
{{< /tabpane >}}

Verify the EnvoyExtensionPolicy status:

```shell
kubectl get envoyextensionpolicy/lua-inline-test -o yaml
```

### Lua Extension - ValueRef

This [EnvoyExtensionPolicy][] configuration defines the Lua extension in a ConfigMap resource.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyExtensionPolicy
metadata:
  name: lua-valueref-test
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: backend
  lua:
    - type: ValueRef
      valueRef:
        name: cm-lua-valueref
        kind: ConfigMap
        group: v1
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm-lua-valueref
data:
  lua: |
    function envoy_on_response(response_handle)
      response_handle:headers():add("x-lua-custom", "FOO")
    end
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resources to your cluster:

```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyExtensionPolicy
metadata:
  name: lua-valueref-test
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: backend
  lua:
    - type: ValueRef
      valueRef:
        name: cm-lua-valueref
        kind: ConfigMap
        group: v1
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm-lua-valueref
data:
  lua: |
    function envoy_on_response(response_handle)
      response_handle:headers():add("x-lua-custom", "FOO")
    end
```

{{% /tab %}}
{{< /tabpane >}}

Verify the EnvoyExtensionPolicy status:

```shell
kubectl get envoyextensionpolicy/lua-valueref-test -o yaml
```

### Lua Extension - FilterContext

The `filterContext` field allows you to pass key/value pairs to a shared Lua script, so it can be parameterized differently per route.
The Lua script accesses these values via the [`filterContext()` stream handle API](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/lua_filter#stream-handle-api).

This example uses a shared ConfigMap Lua script that reads a configurable header name from filter context and copies its value to the `authorization` header.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm-lua-shared-auth
data:
  lua: |
    function envoy_on_request(request_handle)
      local ctx = request_handle:filterContext()
      local token = request_handle:headers():get(ctx:get("token_header"))
      if token and token ~= "" then
        request_handle:headers():replace("authorization", "Bearer " .. token)
      end
    end
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyExtensionPolicy
metadata:
  name: lua-filter-context-test
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: backend
  lua:
    - type: ValueRef
      valueRef:
        name: cm-lua-shared-auth
        kind: ConfigMap
        group: v1
      filterContext:
        token_header: x-api-key
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resources to your cluster:

```yaml
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm-lua-shared-auth
data:
  lua: |
    function envoy_on_request(request_handle)
      local ctx = request_handle:filterContext()
      local token = request_handle:headers():get(ctx:get("token_header"))
      if token and token ~= "" then
        request_handle:headers():replace("authorization", "Bearer " .. token)
      end
    end
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyExtensionPolicy
metadata:
  name: lua-filter-context-test
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: backend
  lua:
    - type: ValueRef
      valueRef:
        name: cm-lua-shared-auth
        kind: ConfigMap
        group: v1
      filterContext:
        token_header: x-api-key
```

{{% /tab %}}
{{< /tabpane >}}

A different route can reuse the same ConfigMap with different filter context values:

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyExtensionPolicy
metadata:
  name: lua-filter-context-route-b
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: backend-b
  lua:
    - type: ValueRef
      valueRef:
        name: cm-lua-shared-auth
        kind: ConfigMap
        group: v1
      filterContext:
        token_header: x-session-token
```

Verify the EnvoyExtensionPolicy status:

```shell
kubectl get envoyextensionpolicy/lua-filter-context-test -o yaml
```

### Testing

Ensure the `GATEWAY_HOST` environment variable from the [Quickstart](../../quickstart) is set. If not, follow the
Quickstart instructions to set the variable.

```shell
echo $GATEWAY_HOST
```

Send a request to the backend service:

```shell
curl -i -H "Host: www.example.com" "http://${GATEWAY_HOST}"
```

You should see that the lua extension has added this header to the response:

```
x-lua-custom: FOO
```

## Clean-Up

Follow the steps from the [Quickstart](../../quickstart) to uninstall Envoy Gateway and the example manifest.

Delete the EnvoyExtensionPolicy:

```shell
kubectl delete envoyextensionpolicy/lua-inline-test
kubectl delete envoyextensionpolicy/lua-valueref-test
kubectl delete envoyextensionpolicy/lua-filter-context-test
kubectl delete configmap/cm-lua-valueref
kubectl delete configmap/cm-lua-shared-auth
```

## Next Steps

Checkout the [Developer Guide](/community/develop) to get involved in the project.

[EnvoyExtensionPolicy]: ../../../api/extension_types#envoyextensionpolicy
[EnvoyGateway]: ../../../api/extension_types#envoygateway
[Gateway]: https://gateway-api.sigs.k8s.io/reference/api-types/gateway/
[HTTPRoute]: https://gateway-api.sigs.k8s.io/reference/api-types/httproute/
[RBAC]: https://kubernetes.io/docs/reference/access-authn-authz/rbac/
[AdmissionControl]: https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/
[AuditLog]: https://kubernetes.io/docs/tasks/debug/debug-cluster/audit/
