---
title: "Metadata in XDS resources"
---

## Overview

In Envoy, [static metadata][] can be configured on various resources: [listener][], [virtual host][], [route][] and [cluster][].

Static metadata can be used for various purposes:
- Observability: enrichment of access logs and traces with [metadata formatters][] and [custom tags][]. 
- Processing: provide configuration context to filters in a certain scope (e.g. vhost, route, etc.).  

This document describes how Envoy Gateway manages [static metadata][] for various XDS resource such as listeners, virtual hosts, routes, clusters and endpoints.

## Configuration 

Envoy Gateway propagates certain attributes of [Gateway-API][gw_api] resources to XDS resources. Attributes include:
- Metadata: Kind, Group/Version, Name, Namespace and Annotations (belonging to the `gateway.envoyproxy.io` namespace)
- Spec: SectionName (Listener Name, RouteRule Name, Port Name), in-spec annotations (e.g. Gateway Annotations)

Future enhancements may include:
- Additional attribute propagation
- Supporting section-specific metadata, e.g. HTTPRoute Metadata annotations that are propagated only to a specific route rule XDS metadata.
- Supporting additional XDS resource, e.g. endpoints and filter chains.

## Translation 

Envoy Gateway uses the following namespace for envoy resource metadata: `gateway.envoyproxy.io/`. For example, an envoy [route][] resource may have the following metadata structure:

Kubernetes resource:

```yaml
kind: HTTPRoute
apiVersion: gateway.networking.k8s.io/v1
metadata:
  annotations:
    gateway.envoyproxy.io/foo: bar
  name: myroute
  namespace: gateway-conformance-infra
spec:
  rules:
    matches:
    - path:
        type: PathPrefix
        value: /mypath
```

Metadata structure:

```yaml
name: httproute/gateway-conformance-infra/myroute/rule/0/match/0/*
match:
  path_separated_prefix: "/mypath"
route:
  cluster: httproute/gateway-conformance-infra/myroute/rule/0
metadata:
  filter_metadata:
    envoy-gateway:
      resources:
        - namespace: gateway-conformance-infra
          groupVersion: gateway.networking.k8s.io/v1
          kind: HTTPRoute
          annotations:
            foo: bar
          name: myroute
```

Envoy Gateway translates [Gateway-API][gw_api] in the following manner:
- [Gateway][gw] metadata is propagated to envoy [listener][] metadata. If merge-gateways is enabled, [Gateway Class][gc] is used instead. 
- [Gateway][gw] metadata and Listener Section name are propagated to envoy [virtual host][] metadata
- [HTTPRoute][httpr] and [GRPCRoute][grpcr] metadata is propagated to envoy [route][] metadata. When Gateway-API adds support [named route rules][], the route rule name
- [TCP/UDPRoute][tcpr] and [TLSRoute][tlsr] resource attributes are not propagated. These resources are translated to envoy filter chains, which do not currently support static metadata. 
- [Service][svc], [ServiceImport][svci] and [Backend][] metadata and port name are propagated to envoy [cluster] metadata.

## Usage

Users can consume metadata in various ways:
- Adding metadata to access logs using the metadata operator, e.g. `%METADATA(ROUTE:envoy-gateway:resources)`
- Accessing metadata in CEL expressions through the `xds.*_metadata` attribute

[static metadata]: https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/base.proto#envoy-v3-api-msg-config-core-v3-metadata
[metadata formatters]: https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/formatter/metadata/v3/metadata.proto.html#formatter-extension-for-printing-various-types-of-metadata-proto
[custom tags]: https://www.envoyproxy.io/docs/envoy/latest/api-v3/type/tracing/v3/custom_tag.proto.html#envoy-v3-api-msg-type-tracing-v3-customtag-metadata
[gw_api]: https://gateway-api.sigs.k8s.io
[gc]: https://gateway-api.sigs.k8s.io/concepts/api-overview/#gatewayclass
[gw]: https://gateway-api.sigs.k8s.io/concepts/api-overview/#gateway
[tlsr]: https://gateway-api.sigs.k8s.io/concepts/api-overview/#tlsroute
[tcpr]: https://gateway-api.sigs.k8s.io/concepts/api-overview/#tcproute-and-udproute
[listener]: https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/listener/v3/listener.proto#envoy-v3-api-msg-config-listener-v3-listener
[virtual host]: https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#config-route-v3-virtualhost
[grpcr]: https://gateway-api.sigs.k8s.io/concepts/api-overview/#grpcroute
[httpr]: https://gateway-api.sigs.k8s.io/concepts/api-overview/#httproute
[named route rules]: https://gateway-api.sigs.k8s.io/geps/gep-995/
[route]: https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#envoy-v3-api-msg-config-route-v3-route
[cluster]: https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/cluster/v3/cluster.proto
[svc]: https://kubernetes.io/docs/concepts/services-networking/service/
[svci]: https://multicluster.sigs.k8s.io/concepts/multicluster-services-api/#serviceimport-and-endpointslices
[Backend]: ../../latest/api/extension_types#backend
