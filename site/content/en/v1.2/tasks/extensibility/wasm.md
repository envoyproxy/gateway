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

Envoy Gateway supports two types of Wasm extensions:
* HTTP Wasm Extension: The Wasm extension is fetched from a remote URL.
* Image Wasm Extension: The Wasm extension is packaged as an OCI image and fetched from an image registry.

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
kubectl get envoyextensionpolicy/http-wasm-source-test -o yaml
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

Verify the EnvoyExtensionPolicy status:

```shell
kubectl get envoyextensionpolicy/http-wasm-source-test -o yaml
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

Checkout the [Developer Guide](../../../contributions/develop) to get involved in the project.

[EnvoyExtensionPolicy]: ../../../api/extension_types#envoyextensionpolicy
[Gateway]: https://gateway-api.sigs.k8s.io/api-types/gateway
[HTTPRoute]: https://gateway-api.sigs.k8s.io/api-types/httproute
