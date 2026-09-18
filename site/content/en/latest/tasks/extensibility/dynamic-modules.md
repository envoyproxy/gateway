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

## Backend dependencies

A module that makes HTTP callouts can declare its backends in `dynamicModule[].backends`.
Envoy Gateway creates a cluster with each declared name. The module must use that same
name when making a callout. Envoy Gateway does not read or rewrite the module's `config`.

For example, a registered `envoy-web-bot-auth` module can call a resolver exposed by a
Service named `web-bot-auth-resolver` on port 8081:

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyExtensionPolicy
metadata:
  name: bot-auth
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: application
  dynamicModule:
    - name: envoy-web-bot-auth
      filterName: web-bot-auth
      backends:
        - name: web-bot-auth-key-resolver
          backendRef:
            name: web-bot-auth-resolver
            port: 8081
      config:
        resolver:
          cluster: web-bot-auth-key-resolver
```

Each entry references one Service, ServiceImport, or Envoy Gateway Backend. References
to another namespace require a ReferenceGrant.

### Resolving conflicts

Cluster names are shared throughout an Envoy deployment, including Gateways combined
with `mergeGateways`. Policies that declare the same name and the same backend reference
share one cluster. Different backend references conflict, even if their endpoints match.

The oldest policy keeps the name. A conflicting policy receives `Accepted=False` with
reason `Conflicted`, and its affected routes return HTTP 500. Use the same backend
reference or choose another cluster name and update the module configuration to match.

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
