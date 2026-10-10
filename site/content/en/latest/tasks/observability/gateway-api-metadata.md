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
| [Virtual Host][] | `Gateway`                                              | Kind, Namespace, Name, Annotations, SectionName (`spec.listeners.<listener>.name`)  |                                                                                   |
| [Route][]        | `HTTPRoute`, `GRPCRoute`                               | Kind, Namespace, Name, Annotations, SectionName (`spec.listener.rules.<rule>.name`) |                                                                                   |
| [Cluster][]      | `xRoute`                                               | Kind, Namespace, Name, Annotations, SectionName (`spec.listener.rules.<rule>.name`) |                                                                                   |
| [Cluster][]      | `EnvoyProxy`, `EnvoyExtensionPolicy`, `SecurityPolicy` | Kind, Namespace, Name, Annotations, SectionName (`spec.listener.rules.<rule>.name`) |  When a non-xRoute BackendRef is used (e.g. ext_auth, observabiltiy sink, ... )   |
| [Locality Endpoints][] | `Service`, `ServiceImport`, `Backend`            | Kind, Namespace, Name, Annotations, SectionName (`backendRef.port`)                 |  Attached to each entry of the cluster load assignment's `endpoints`              |
| [LbEndpoint][]   | `Service`, `ServiceImport`, `Backend`                  | Kind, Namespace, Name, SectionName (`backendRef.port`)                              |  Optional, requires the `LbEndpointMetadata` runtime flag, see [Per-Endpoint Backend Metadata](#per-endpoint-backend-metadata) |

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

## Per-Endpoint Backend Metadata
The backend metadata described above is attached to each [Locality Endpoints][] entry (`endpoints[].metadata` in the cluster load assignment) under `envoy-gateway.resources`. Envoy Gateway can additionally attach backend metadata to each individual [LbEndpoint][] (`endpoints[].lbEndpoints[].metadata`). This is needed for Envoy features that read the upstream host's metadata, such as the `%METADATA(UPSTREAM_HOST:...)%` access log command operator, which does not see locality-level metadata. It is disabled by default, since it adds metadata to every endpoint and increases the size of the XDS configuration sent to each proxy. Enable it with the `LbEndpointMetadata` runtime flag in the [EnvoyGateway][] configuration:

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyGateway
gateway:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
runtimeFlags:
  enabled:
  - LbEndpointMetadata
```

Each `lbEndpoints` entry is then annotated with the originating backend under `metadata.filter_metadata.envoy-gateway.backend`. Unlike the locality-level metadata, this contains only the kind, name, namespace and sectionName (no annotations):

```yaml
metadata:
  filter_metadata:
    envoy-gateway:
      backend:
        kind: Service
        name: my-service
        namespace: default
        sectionName: "8080"
```

This allows access logs to record which backend served a request, for example with the `%METADATA(UPSTREAM_HOST:envoy-gateway:backend:name)%` command operator in an [EnvoyProxy access log format](./proxy-accesslog.md).

## Use Cases
XDS Metadata serves multiple purposes:
- Observability: Envoy proxy access logs can be [enriched with Gateway-API resource](./proxy-accesslog.md) context and custom annotations, creating an association with relevant Application Developers personas.
- Troubleshooting: users and tools that analyze the envoy proxy XDS config can identify the Gateway API resources that lead to the XDS configuration's creation.
- Extensibility:
  - Envoy Gateway [Extension Servers][] can leverage Gateway API metadata as additional context annotating XDS resources sent for mutation.
  - Envoy Proxy extensions can leverage XDS metadata as additional context when processing traffic:
    - [lua][] extensions can access metadata using the [Metadata Stream handle API][] .
    - [ext_proc][] extensions can access metadata using the `xds.*_metadata` [ext_proc attribute][].
  

[EnvoyGateway]: ../../api/extension_types#envoygateway
[Static Metadata]: https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/base.proto#envoy-v3-api-msg-config-core-v3-metadata
[Virtual Host]: https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#config-route-v3-virtualhost
[Route]: https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#envoy-v3-api-msg-config-route-v3-route
[Cluster]: https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/cluster/v3/cluster.proto.html
[Locality Endpoints]: https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/endpoint/v3/endpoint_components.proto#config-endpoint-v3-localitylbendpoints
[LbEndpoint]: https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/endpoint/v3/endpoint_components.proto#envoy-v3-api-msg-config-endpoint-v3-lbendpoint
[Metadata Stream handle API]: https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/lua_filter#metadata
[lua]: ../../api/extension_types#lua
[ext_proc]: ../../api/extension_types#extproc
[ext_proc attribute]: ../../api/extension_types#processingmodeoptions
[Extension Servers]: ../extensibility/extension-server.md



