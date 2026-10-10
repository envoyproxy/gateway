---
title: "Backend Dependencies for Extensions"
---

An extension may need to call another service before allowing a request to
continue, for example to check authorization or look up a key. Declare that
service in `EnvoyExtensionPolicy.spec.backends` so Envoy Gateway can configure
the connection for Lua, Wasm or dynamic module HTTP callouts.

Envoy sends these calls through a cluster, which holds the backend's endpoints
and connection settings. Your extension uses an alias you choose, such as
`resolver`, to find the cluster name during a request. This keeps generated
cluster names out of your extension configuration and lets you change the
backend reference while keeping the same alias in your code. Envoy Gateway
resolves endpoints, authorizes references and configures backend TLS.

For the Lua example, first [enable Lua support](../lua/) in Envoy Gateway.

## Declare a backend

This policy binds the alias `resolver` to a Kubernetes Service. The Lua script
looks up that alias and calls the resulting cluster. The Service does not need
to appear in the HTTPRoute's `backendRefs`. Your route can send client requests
to the application while the extension uses a separate service for its lookup.

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyExtensionPolicy
metadata:
  name: resolverpolicy
  namespace: default
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: app
  backends:
  - name: resolver
    backendRef:
      name: resolver
      port: 8080
  lua:
  - type: Inline
    inline: |
      function envoy_on_request(handle)
        local cluster = handle:route():metadata():get("resolver")
        if cluster == nil then
          handle:respond({[":status"] = "502"}, "missing resolver binding")
          return
        end
        local headers = handle:httpCall(cluster, {
          [":method"] = "GET",
          [":path"] = "/lookup",
          [":authority"] = "resolver"
        }, nil, 2000)
        if headers[":status"] ~= "200" then
          handle:respond({[":status"] = "502"}, "resolver callout failed")
        end
      end
```

If the lookup returns HTTP 200, the original request continues to the route's
backend. Otherwise, the script responds with HTTP 502.

Aliases must be lowercase DNS labels and unique within the policy, with at most
16 bindings. `resolver` is an example alias chosen by the operator. Separate
policies can use the same alias for different Services on the same proxy. Lua,
Wasm and dynamic modules using the same effective policy share its backend bindings.
The HTTP authority, path, timeout and response handling remain part of extension
configuration or code. The alias supplies only the Envoy cluster name.

## Configure backend connections

Use `backendSettings` to control how Envoy connects to the callout service.
For example, this binding allows three seconds to establish a connection and
limits the cluster to 100 connections:

```yaml
backends:
- name: resolver
  backendRef:
    name: resolver
    port: 8080
  backendSettings:
    timeout:
      tcp:
        connectTimeout: 3s
    circuitBreaker:
      maxConnections: 100
```

These settings configure the callout cluster. Omitted settings use existing
defaults. Inherited bindings keep their settings. BackendTrafficPolicy settings
for route traffic do not apply to these callouts. A connection timeout limits
how long Envoy waits to connect. The extension controls how long it waits for
an individual callout and whether to retry. For that reason, `requestTimeout`
and `streamIdleTimeout` are rejected here. `DynamicModule` load balancing is
not supported for these backends.

Backend references support the same resources as external processing services:
Service, ServiceImport and Envoy Gateway Backend. Enable the Backend API before
referencing Backend resources. References to another namespace require a ReferenceGrant
allowing `gateway.envoyproxy.io/EnvoyExtensionPolicy` from the declaring policy's
namespace. BackendTLSPolicy and applicable backend TLS settings are resolved by
Envoy Gateway.
DynamicResolver backends and mixed DNS/static endpoint sets are not supported.

## Use a Unix socket backend

Reference a Backend resource to reach a sidecar over a Unix socket:

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: Backend
metadata:
  name: sidecar
  namespace: default
spec:
  endpoints:
  - unix:
      path: /run/resolver/socket
```

Bind it in the policy:

```yaml
backends:
- name: resolver
  backendRef:
    group: gateway.envoyproxy.io
    kind: Backend
    name: sidecar
```

Enable the Backend API and mount the socket directory in both Envoy and the
sidecar. Operators manage the sidecar and socket lifecycle. Extensions use the
same alias lookup and HTTP callout APIs as for a Service.

## Read a binding

The alias map is stored in [Envoy's route metadata][] with the route
configuration. Reading it is a local lookup, with no request to Envoy Gateway
or Kubernetes. It is not an HTTP header and is not sent to the backend service.

Envoy Gateway publishes a flat map of aliases to cluster names in
the selected route's `metadata.filter_metadata` under
`gateway.envoyproxy.io/extension-backends`. It also publishes the map under each
Lua filter's exact name, because Lua's metadata accessor selects that namespace.

Lua uses `handle:route():metadata():get("resolver")`, as above.

A Wasm extension uses its SDK's property getter with this path:

```text
["xds", "route_metadata", "filter_metadata",
 "gateway.envoyproxy.io/extension-backends", "resolver"]
```

Decode the returned bytes as a UTF-8 cluster name and pass it to
`dispatch_http_call` (or the equivalent call in the SDK).

The dynamic module Go SDK provides this lookup:

```go
cluster, found := handle.GetMetadataString(
    shared.MetadataSourceTypeRoute,
    "gateway.envoyproxy.io/extension-backends",
    "resolver",
)
if !found {
    // Return a local error for the missing binding.
    return
}
// Use cluster.ToString() as the cluster argument to handle.HttpCallout.
```

The Rust SDK provides `get_metadata_string` with metadata source `Route` and the
same namespace and key. SDK buffers can borrow Envoy memory. Copy the string if
it must outlive the callback. See the [dynamic module test example][] for a complete
HTTP callout using the Go SDK.

## Scope and lifecycle

Bindings belong to the selected route, so a shared extension can call different
services for different applications. Always resolve the alias for the current
request rather than keeping one cluster name for the lifetime of the extension.

Look up bindings while processing request headers or later in the request's
context. Initialization and independent background tasks have no request route.
Asynchronous callout completion remains supported. Resolve the binding in request
context and use it only for that request, copying borrowed strings if needed.
Do not cache a binding across requests: another route can map the same
alias to a different backend, including when the same filter instance serves both
routes. Handle missing bindings explicitly and treat generated cluster names as
opaque. Route changes can change which bindings are available.

Normal EnvoyExtensionPolicy attachment precedence applies. When merging a route
policy with a parent policy, an omitted or empty `backends` list inherits the parent's
bindings and reference namespace. A supplied nonempty list replaces the whole
list, so repeat every binding needed by inherited extensions. This also lets a
route intentionally rebind an alias used by an inherited extension. Without
merging, only the selected policy's bindings apply.

For example, a Gateway policy can provide a common lookup service. A route
policy that merges with it can replace the `resolver` binding to use a service
for that application while retaining the parent's extension code.

Invalid references set the policy's `Accepted` condition to `False` and affected
routes return HTTP 500 until the dependency is corrected. This applies even when
a Wasm extension has `failOpen` enabled: the policy's declared dependency is
invalid. Unrelated policies continue to operate.

[dynamic module test example]: https://github.com/envoyproxy/gateway/tree/main/examples/dynamic-module-test
[Envoy's route metadata]: https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#envoy-v3-api-field-config-route-v3-route-metadata
