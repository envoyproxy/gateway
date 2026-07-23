---
title: "Wasm Extensions"
---

This task provides instructions for extending Envoy Gateway with WebAssembly (Wasm) extensions.

Wasm extensions allow you to extend the functionality of Envoy Gateway by running custom code against HTTP requests and responses,
without modifying the Envoy Gateway binary. These extensions can be written in any language that compiles to Wasm, such as C++, Rust, AssemblyScript, or TinyGo.

Envoy Gateway introduces a new CRD called [EnvoyExtensionPolicy][] that allows the user to configure Wasm extensions.
This instantiated resource can be linked to a [Gateway][Gateway] and [HTTPRoute][HTTPRoute] resource.

## Prerequisites

{{< boilerplate prerequisites >}}

## Configuration

Envoy Gateway supports three types of Wasm extensions:
* HTTP Wasm Extension: The Wasm extension is fetched from a remote URL.
* Image Wasm Extension: The Wasm extension is packaged as an OCI image and fetched from an image registry.
* EnvoyProxyModule Wasm Extension: The Wasm extension is loaded from a module registered on [EnvoyProxy][] (`spec.wasmModules`). Today only a Local filesystem path is supported; the policy references the module by name.

The following example demonstrates how to configure an [EnvoyExtensionPolicy][] to attach a Wasm extension to an [EnvoyExtensionPolicy][] .
This Wasm extension adds a custom header `x-wasm-custom: FOO` to the response.

### HTTP Wasm Extension

This [EnvoyExtensionPolicy][] configuration fetches the Wasm extension from an HTTP URL.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyExtensionPolicy
metadata:
  name: wasm-test
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: backend
  wasm:
  - name: wasm-filter
    rootID: my_root_id
    code:
      type: HTTP
      http:
        url: https://raw.githubusercontent.com/envoyproxy/examples/main/wasm-cc/lib/envoy_filter_http_wasm_example.wasm
        sha256: 79c9f85128bb0177b6511afa85d587224efded376ac0ef76df56595f1e6315c0
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
  name: wasm-test
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: backend
  wasm:
  - name: wasm-filter
    rootID: my_root_id
    code:
      type: HTTP
      http:
        url: https://raw.githubusercontent.com/envoyproxy/examples/main/wasm-cc/lib/envoy_filter_http_wasm_example.wasm
        sha256: 79c9f85128bb0177b6511afa85d587224efded376ac0ef76df56595f1e6315c0
```

{{% /tab %}}
{{< /tabpane >}}

Verify the EnvoyExtensionPolicy status:

```shell
kubectl get envoyextensionpolicy/wasm-test -o yaml
```

### Image Wasm Extension

This [EnvoyExtensionPolicy][] configuration fetches the Wasm extension from an OCI image.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyExtensionPolicy
metadata:
  name: wasm-test
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: backend
  wasm:
  - name: wasm-filter
    rootID: my_root_id
    code:
      type: Image
      image:
        url: zhaohuabing/testwasm:v0.0.1
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
  name: wasm-test
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: backend
  wasm:
  - name: wasm-filter
    rootID: my_root_id
    code:
      type: Image
      image:
        url: zhaohuabing/testwasm:v0.0.1
```

{{% /tab %}}
{{< /tabpane >}}

### EnvoyProxyModule Wasm Extension

Register the module on the [EnvoyProxy][] attached to the Gateway, then reference it by name from the [EnvoyExtensionPolicy][]. Envoy Gateway does not place Local modules on the proxy; provision them with a custom Envoy image or a volume mount. Local modules skip the control-plane download path, which avoids a fail-closed load window when the file is already on the proxy.

Update the EnvoyProxy used by the Gateway:

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: custom-proxy-config
  namespace: envoy-gateway-system
spec:
  wasmModules:
  - name: example-filter
    source:
      type: Local
      local:
        path: /var/lib/envoy/example-filter.wasm
```

Then apply the EnvoyExtensionPolicy:

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyExtensionPolicy
metadata:
  name: wasm-test
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: backend
  wasm:
  - name: wasm-filter
    rootID: my_root_id
    code:
      type: EnvoyProxyModule
      envoyProxyModule:
        name: example-filter
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
  name: wasm-test
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: backend
  wasm:
  - name: wasm-filter
    rootID: my_root_id
    code:
      type: EnvoyProxyModule
      envoyProxyModule:
        name: example-filter
```

{{% /tab %}}
{{< /tabpane >}}

Verify the EnvoyExtensionPolicy status:

```shell
kubectl get envoyextensionpolicy/wasm-test -o yaml
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

You should see that the wasm extension has added this header to the response:

```
x-wasm-custom: FOO
```

## Clean-Up

Follow the steps from the [Quickstart](../../quickstart) to uninstall Envoy Gateway and the example manifest.

Delete the EnvoyExtensionPolicy:

```shell
kubectl delete envoyextensionpolicy/wasm-test
```

## Next Steps

Checkout the [Developer Guide](/community/develop) to get involved in the project.

[EnvoyExtensionPolicy]: ../../../api/extension_types#envoyextensionpolicy
[EnvoyProxy]: ../../../api/extension_types#envoyproxy
[Gateway]: https://gateway-api.sigs.k8s.io/reference/api-types/gateway/
[HTTPRoute]: https://gateway-api.sigs.k8s.io/reference/api-types/httproute/
