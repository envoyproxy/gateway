---
title: "Gateway API Metadata"
---

## Background
Envoy Gateway translates Gateway API resources to Envoy XDS resources. In this translation process, Envoy Gateway annotates XDS resources with additional metadata from their origin Gateway API resources.

Gateway API Metadata includes:
- K8s Resource Kinds, Names and Namespaces.
- K8s Resource Annotations with the `gateway.envoyproxy.io/` prefix.
- K8s Resource SectionNames (when applicable, e.g. for Route rules and Listeners).

Gateway API Metadata is added to XDS resources using envoy's [Static Metadata][] under `metadata.filter_metadata.envoy-gateway.resources`. Currently, `resources` only contains the primary origin resource.
However, in the future, additional relevant resources (e.g. policies, filters attached to the primary origin resources) may be added.

## Supported Resources
Currently, the following mapping of Gateway API metadata to XDS metadata are supported:

| XDS Resource     | Primary Gateway API Resource                           | XDS Metadata                                                                        | Comments
|------------------|--------------------------------------------------------|-------------------------------------------------------------------------------------|-----------------------------------------------------------------------------------|
| [Filter Chain][] | `Gateway`                                              | Kind, Namespace, Name, Annotations, SectionName (`spec.listeners.<listener>.name`)  |  Only for TCP, TLS and HTTPS Listeners                                            |
| [Virtual Host][] | `Gateway`                                              | Kind, Namespace, Name, Annotations, SectionName (`spec.listeners.<listener>.name`)  |                                                                                   |
| [Route][]        | `HTTPRoute`, `GRPCRoute`                               | Kind, Namespace, Name, Annotations, SectionName (`spec.listener.rules.<rule>.name`) |                                                                                   |
| [Cluster][]      | `xRoute`                                               | Kind, Namespace, Name, Annotations, SectionName (`spec.listener.rules.<rule>.name`) |                                                                                   |
| [Cluster][]      | `EnvoyProxy`, `EnvoyExtensionPolicy`, `SecurityPolicy` | Kind, Namespace, Name, Annotations, SectionName (`spec.listener.rules.<rule>.name`) |  When a non-xRoute BackendRef is used (e.g. ext_auth, observability sink, ... )   |
| [LBEndpoints][]  | `Service`, `ServiceImport`, `Backend`                  | Kind, Namespace, Name, Annotations, SectionName (`backendRef.port`)                 |                                                                                   |

For example, consider the following Gateway API HTTPRoute:

```yaml
kind: HTTPRoute
apiVersion: gateway.networking.k8s.io/v1
metadata:
  annotations:
    gateway.envoyproxy.io/foo: bar
  name: myroute
  namespace: myns
spec:
  rules:
  - name: myrule
    matches:
    - path:
        type: PathPrefix
        value: /mypath
```

The translated XDS [Route][] contains Gateway API metadata under :

```yaml
name: httproute/myns/myroute/rule/0/match/0/*
match:
  path_separated_prefix: "/mypath"
route:
  cluster: httproute/myns/myroute/rule/0
metadata:
  filter_metadata:
    envoy-gateway:
      resources:
      - namespace: myns
        kind: HTTPRoute
        annotations:
          foo: bar
        name: myroute
        sectionName: myrule
```

## Use Cases
XDS Metadata serves multiple purposes:
- Observability: Envoy proxy access logs can be [enriched with Gateway-API resource](./proxy-accesslog.md) context and custom annotations, creating an association with relevant Application Developers personas.
- Troubleshooting: users and tools that analyze the envoy proxy XDS config can identify the Gateway API resources that lead to the XDS configuration's creation.
- Extensibility:
  - Envoy Gateway [Extension Servers][] can leverage Gateway API metadata as additional context annotating XDS resources sent for mutation.
  - Envoy Proxy extensions can leverage XDS metadata as additional context when processing traffic:
    - [lua][] extensions can access metadata using the [Metadata Stream handle API][] .
    - [ext_proc][] extensions can access metadata using the `xds.*_metadata` [ext_proc attribute][].
  
## Operational Considerations
In some cases, changes in XDS metadata may lead to traffic disruption. For example, changing [Filter Chain][] and [Listener][] XDS metadata will lead to a [filter chain drain][] and [listener drain][] accordingly, which may disrupt long-lived connections.

When generating metadata for Listeners and Filter Chains, Envoy Gateway considers the following:
- Multiple Gateway-API `Gateway` resources can be merged into a single XDS Listener. In this case, there is no primary Gateway-API resource for the XDS listener.
- Multiple Gateway-API `Gateway` HTTP listeners can be merged into a single Filter Chain, using distinct HTTP virtual hosts. In this case, there is no primary Gateway-API resource for the XDS Filter Chain.
- A change in Gateway-API `Gateway` resource may lead to multiple filter chains XSD metadata being updated and causing drains:
  - Annotation changes will lead to a drain across all filter chains derived from a gateway
  - SectionName changes will lead to a localized drain in the specific filter chain

As a result, Envoy Gateway currently only supports HTTPS Listener filter chain metadata at this time, with the following resrtictions:
- To propagate K8s annotations from a `Gateway` resources to XDS Filter Chains, a distinct annotation namespace must be used: `ingress.gateway.envoyproxy.io`.
 - A change in an annotation with this unique prefix will trigger a drain of all Filter Chains originating from the `Gateway` resource.
 - Other annotations are ignored for Filter Chains.
- SectionNames are not propagated to avoid unintended drains.


[Static Metadata]: https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/base.proto#envoy-v3-api-msg-config-core-v3-metadata
[Filter Chain]: https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/listener/v3/listener_components.proto#envoy-v3-api-msg-config-listener-v3-filterchain
[Listener]: https://www.envoyproxy.io/docs/envoy/latest/configuration/listeners/listeners#config-listeners
[Virtual Host]: https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#config-route-v3-virtualhost
[Route]: https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#envoy-v3-api-msg-config-route-v3-route
[Cluster]: https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/cluster/v3/cluster.proto.html
[LBEndpoints]: https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/endpoint/v3/endpoint_components.proto#envoy-v3-api-msg-config-endpoint-v3-lbendpoint
[Metadata Stream handle API]: https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/lua_filter#metadata
[lua]: ../../api/extension_types#lua
[ext_proc]: ../../api/extension_types#extproc
[ext_proc attribute]: ../../api/extension_types#processingmodeoptions
[Extension Servers]: ../extensibility/extension-server.md
[filter chain drain]: https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/listeners/listener_filters.html#filter-chain-only-update
[listener drain]: https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/operations/draining

