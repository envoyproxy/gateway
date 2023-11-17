---
title: "Gateway API Support"
---

As mentioned in the [system design][] document, Envoy Gateway's managed data plane is configured dynamically through
Kubernetes resources, primarily [Gateway API][] objects. Envoy Gateway supports configuration using the following Gateway API resources.

## GatewayClass

A [GatewayClass][] represents a "class" of gateways, i.e. which Gateways should be managed by Envoy Gateway.
Envoy Gateway supports managing __a single__ GatewayClass resource that matches its configured `controllerName` and
follows Gateway API guidelines for [resolving conflicts][] when multiple GatewayClasses exist with a matching
`controllerName`.

__Note:__ If specifying GatewayClass [parameters reference][], it must refer to an [EnvoyProxy][] resource.

## Gateway

When a [Gateway][] resource is created that references the managed GatewayClass, Envoy Gateway will create and manage a
new Envoy Proxy deployment. Gateway API resources that reference this Gateway will configure this managed Envoy Proxy
deployment.

## HTTPRoute

An [HTTPRoute][] configures routing of HTTP traffic through one or more Gateways. The following HTTPRoute filters are
supported by Envoy Gateway:

- `requestHeaderModifier`: [RequestHeaderModifiers][http-filter]
  can be used to modify or add request headers before the request is proxied to its destination.
- `responseHeaderModifier`: [ResponseHeaderModifiers][http-filter]
  can be used to modify or add response headers before the response is sent back to the client.
- `requestMirror`: [RequestMirrors][http-filter]
  configure destinations where the requests should also be mirrored to. Responses to mirrored requests will be ignored.
- `requestRedirect`: [RequestRedirects][http-filter]
  configure policied for how requests that match the HTTPRoute should be modified and then redirected.
- `urlRewrite`: [UrlRewrites][http-filter]
  allow for modification of the request's hostname and path before it is proxied to its destination.
- `extensionRef`: [ExtensionRefs][] are used by Envoy Gateway to implement extended filters. Currently, Envoy Gateway
  supports rate limiting and request authentication filters. For more information about these filters, refer to the
  [rate limiting][] and [request authentication][] documentation.

__Notes:__
- The only [BackendRef][] kind supported by Envoy Gateway is a [Service][]. Routing traffic to other destinations such
  as arbitrary URLs is not possible.
- The `filters` field within [HTTPBackendRef][] is not supported.

## TCPRoute

A [TCPRoute][] configures routing of raw TCP traffic through one or more Gateways. Traffic can be forwarded to the
desired BackendRefs based on a TCP port number.

__Note:__ A TCPRoute only supports proxying in non-transparent mode, i.e. the backend will see the source IP and port of
the Envoy Proxy instance instead of the client.

## UDPRoute

A [UDPRoute][] configures routing of raw UDP traffic through one or more Gateways. Traffic can be forwarded to the
desired BackendRefs based on a UDP port number.

__Note:__ Similar to TCPRoutes, UDPRoutes only support proxying in non-transparent mode i.e. the backend will see the
source IP and port of the Envoy Proxy instance instead of the client.

## GRPCRoute

A [GRPCRoute][] configures routing of [gRPC][] requests through one or more Gateways. They offer request matching by
hostname, gRPC service, gRPC method, or HTTP/2 Header. Envoy Gateway supports the following filters on GRPCRoutes to
provide additional traffic processing:

- `requestHeaderModifier`: [RequestHeaderModifiers][grpc-filter]
  can be used to modify or add request headers before the request is proxied to its destination.
- `responseHeaderModifier`: [ResponseHeaderModifiers][grpc-filter]
  can be used to modify or add response headers before the response is sent back to the client.
- `requestMirror`: [RequestMirrors][grpc-filter]
  configure destinations where the requests should also be mirrored to. Responses to mirrored requests will be ignored.

__Notes:__
- The only [BackendRef][grpc-filter] kind supported by Envoy Gateway is a [Service][]. Routing traffic to other
  destinations such as arbitrary URLs is not currently possible.
- The `filters` field within [HTTPBackendRef][] is not supported.

## TLSRoute

A [TLSRoute][] configures routing of TCP traffic through one or more Gateways. However, unlike TCPRoutes, TLSRoutes
can match against TLS-specific metadata.

## ReferenceGrant

A [ReferenceGrant][] is used to allow a resource to reference another resource in a different namespace. Normally an
HTTPRoute created in namespace `foo` is not allowed to reference a Service in namespace `bar`. A ReferenceGrant permits
these types of cross-namespace references. Envoy Gateway supports the following ReferenceGrant use-cases:

- Allowing an HTTPRoute, GRPCRoute, TLSRoute, UDPRoute, or TCPRoute to reference a Service in a different namespace.
- Allowing an HTTPRoute's `requestMirror` filter to include a BackendRef that references a Service in a different
  namespace.
- Allowing a Gateway's [SecretObjectReference][] to reference a secret in a different namespace.

[system design]: ../../design/system-design/
[Gateway API]: https://gateway-api.sigs.k8s.io/
[GatewayClass]: https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.GatewayClass
[parameters reference]: https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.ParametersReference
[Gateway]: https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.Gateway
[HTTPRoute]: https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.HTTPRoute
[Service]: https://kubernetes.io/docs/concepts/services-networking/service/
[BackendRef]: https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.BackendRef
[HTTPBackendRef]: https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.HTTPBackendRef
[TCPRoute]: https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.TCPRoute
[UDPRoute]: https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.UDPRoute
[GRPCRoute]: https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.GRPCRoute
[gRPC]: https://grpc.io/
[TLSRoute]: https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.TLSRoute
[ReferenceGrant]: https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.ReferenceGrant
[SecretObjectReference]: https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.SecretObjectReference
[rate limiting]: ../rate-limit/
[request authentication]: ../jwt-authentication/
[EnvoyProxy]: ../../api/extension_types#envoyproxy
[resolving conflicts]: https://gateway-api.sigs.k8s.io/concepts/guidelines/?h=conflict#conflicts
[ExtensionRefs]: https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.HTTPRouteFilterType
[grpc-filter]: https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.GRPCRouteFilter
[http-filter]: https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.HTTPRouteFilter
