---
title: "Dynamic Modules"
---

This task provides instructions for configuring [Dynamic Modules](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/advanced/dynamic_modules).

Dynamic Modules are a critical extension mechanism that allows functionality to be loaded directly into the Envoy proxy process at runtime, typically as shared object files (`.so`). This approach enables customized filtering and request processing without requiring a full recompile of the core Envoy binary, streamlining deployments and upgrades.

Envoy Gateway is able to load dynamic modules from the local filesystem using [EnvoyExtensionPolicy][]. This example demonstrates it's working by loading the [Coraza](https://www.coraza.io/) Web Application Firewall (WAF) using [Built On Envoy](https://builtonenvoy.io/) development toolkit. The module's full documentation can be found on the [Coraza WAF extension page](https://builtonenvoy.io/extensions/coraza-waf/).

## Prerequisites

{{< boilerplate prerequisites >}}

## Coraza WAF Extension

### Installation

Add the dynamic module to the Envoy proxy container's filesystem and configure the [DynamicModules][] spec to load it into Envoy.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: my-proxy
spec:
  provider:
    type: Kubernetes
    kubernetes:
      envoyDeployment:
        pod:
          volumes:
            - name: dynamic-modules
              image:
                reference: ghcr.io/tetratelabs/built-on-envoy/composer:0.6.0-dev
                pullPolicy: IfNotPresent
        container:
          env:
            - name: GODEBUG
              value: "cgocheck=0"
          volumeMounts:
            - name: dynamic-modules
              mountPath: /etc/envoy/dynamic-modules
              readOnly: true
  dynamicModules:
    - name: composer
      source: 
        type: Local
        local:
          path: /etc/envoy/dynamic-modules/libcomposer.so
      doNotClose: true
      loadGlobally: false
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: my-proxy
spec:
  provider:
    type: Kubernetes
    kubernetes:
      envoyDeployment:
        pod:
          volumes:
            - name: dynamic-modules
              image:
                reference: ghcr.io/tetratelabs/built-on-envoy/composer:0.6.0-dev
                pullPolicy: IfNotPresent
        container:
          env:
            - name: GODEBUG
              value: "cgocheck=0"
          volumeMounts:
            - name: dynamic-modules
              mountPath: /etc/envoy/dynamic-modules
              readOnly: true
  dynamicModules:
    - name: composer
      source: 
        type: Local
        local:
          path: /etc/envoy/dynamic-modules/libcomposer.so
      doNotClose: true
      loadGlobally: false
```

{{% /tab %}}
{{< /tabpane >}}

**Note: verify version compatibility of the extension because of dynamic module forward compatibility.**

[Image Volumes](https://kubernetes.io/docs/tasks/configure-pod-container/image-volumes/) are relatively new and only supported from Kubernetes `1.35` onwards.
Alternative ways of loading the dynamic module are:
- Building a custom docker image
- Copying from `InitContainer` to a shared volume
- Letting Envoy Gateway fetch the module from an OCI image with the `Image` source type, described below

#### Loading the module from an OCI image

With the `Image` source type, Envoy Gateway pulls the OCI image, extracts the shared library, and serves it to Envoy
from the same built-in HTTP server it uses for Wasm modules. Envoy verifies the module with its SHA256 checksum.
This works on any Kubernetes version and does not require changes to the Envoy proxy pod.

```yaml
  dynamicModules:
    - name: composer
      source:
        type: Image
        image:
          url: ghcr.io/tetratelabs/built-on-envoy/composer:0.6.0-dev
          path: /libcomposer.so
          pullPolicy: IfNotPresent
      doNotClose: true
      loadGlobally: false
```

- `path` is the path of the shared library inside the image. It can be omitted if the image contains exactly one `.so` file,
  for example an OCI artifact pushed with `oras push <registry>/<repository>:<tag> libmodule.so`.
- `sha256` optionally pins the digest of the image. For a multi-platform image, this is the digest of the image index.
- `pullSecretRef` references a `kubernetes.io/dockerconfigjson` Secret for private registries. A Secret in another
  namespace than the EnvoyProxy requires a ReferenceGrant.
- For a multi-platform image, Envoy Gateway uses the `linux` image for its own CPU architecture, so the Envoy Gateway
  controller and the Envoy proxies must run on the same architecture.
- The environment variables required by the module, such as `GODEBUG` above, still need to be set on the Envoy container.


Verify the [EnvoyProxy][] status:

```shell
kubectl get envoyproxy/my-proxy -o yaml
```

Attach to Gateway via a GatewayClass with `spec.parametersRef`

```shell
kubectl patch gatewayclass eg --type=merge -p '{
  "spec": {
    "parametersRef": {
      "group": "gateway.envoyproxy.io",
      "kind": "EnvoyProxy",
      "name": "my-proxy",
      "namespace": "envoy-gateway-system"
    }
  }
}'
```

The entire configuration can also be specified directly on the Gateway instead by using `spec.envoyProxy`.

### Configuration

Create a new EnvoyExtensionPolicy resource to configure the dynamic module for an entire Gateway or per HTTPRoute.

This EnvoyExtensionPolicy targets the Gateway "eg" created with the quickstart. It loads the Coraza WAF extension with its configuration.


{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyExtensionPolicy
metadata:
  name: waf-extension
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: Gateway
      name: eg
  dynamicModule:
    - name: composer
      filterName: coraza-waf
      config:
        directives:
          - Include @coraza.conf
          - SecRuleEngine On
          - SecResponseBodyAccess Off
          - Include @crs-setup.conf
          - Include @owasp_crs/*.conf
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
  name: waf-extension
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: Gateway # or e.g. HTTPRoute
      name: eg

  dynamicModule:
    - name: composer
      filterName: coraza-waf
      config:
        directives:
          - Include @coraza.conf
          - SecRuleEngine On
          - SecResponseBodyAccess Off
          - Include @crs-setup.conf
          - Include @owasp_crs/*.conf
```

{{% /tab %}}
{{< /tabpane >}}

Verify the Envoy Extension Policy configuration:

```shell
kubectl get envoyextensionpolicy/waf-extension -o yaml
```

### Testing

Ensure the `GATEWAY_HOST` environment variable from the [Quickstart](../../quickstart) is set. If not, follow the
Quickstart instructions to set the variable.

```shell
echo $GATEWAY_HOST
```

Send a normal request to the backend service:

```shell
curl -v -H "Host: www.example.com" "http://${GATEWAY_HOST}/"
```

You should get a `200 OK` response from the backend.

Now send a request with a SQL injection payload to trigger the WAF:

```shell
curl -v -H "Host: www.example.com" "http://${GATEWAY_HOST}/?id=1'+OR+'1'%3D'1"
```

The Coraza WAF should block the request and return a `403 Forbidden` response:

```
> GET /?id=1'+OR+'1'='1 HTTP/1.1
> Host: www.example.com
[...]
< HTTP/1.1 403 Forbidden
< date: Sat, 03 May 2026 12:00:00 GMT
< content-length: 0
<
```

## Clean-Up

Follow the steps from the [Quickstart](../../quickstart) to uninstall Envoy Gateway and the example manifest.

Delete the [EnvoyProxy][] and [EnvoyExtensionPolicy][]:

```shell
kubectl delete envoyproxy/my-proxy
kubectl delete envoyextensionpolicy/waf-extension
```

## Next Steps

Checkout the [Developer Guide](/community/develop) to get involved in the project.

[EnvoyProxy]: ../../../api/extension_types#envoyproxy
[EnvoyExtensionPolicy]: ../../../api/extension_types#envoyextensionpolicy
[DynamicModules]: ../../../api/extension_types#dynamicmoduleentry
