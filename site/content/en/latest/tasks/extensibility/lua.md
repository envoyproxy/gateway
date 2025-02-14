---
title: "Lua Extensions"
---

This task provides instructions for extending Envoy Gateway with Lua extensions.

Lua extensions allow you to extend the functionality of Envoy Gateway by running custom code against HTTP requests and responses,
without modifying the Envoy Gateway binary. These comparatively light-weight extensions are written in the Lua scripting language using APIs defined [here](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/lua_filter#stream-handle-api).

Envoy Gateway allows the user to configure Lua extensions using the [EnvoyExtensionPolicy][] CRD.
This instantiated resource can be linked to a [Gateway][Gateway] or [HTTPRoute][HTTPRoute] resource. If linked to both, the resource linked to the route takes precedence over those linked to Gateway.

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
kubectl delete configmap/cm-lua-valueref
```

## Next Steps

Checkout the [Developer Guide](../../../contributions/develop) to get involved in the project.

[EnvoyExtensionPolicy]: ../../../api/extension_types#envoyextensionpolicy
[Gateway]: https://gateway-api.sigs.k8s.io/api-types/gateway
[HTTPRoute]: https://gateway-api.sigs.k8s.io/api-types/httproute
