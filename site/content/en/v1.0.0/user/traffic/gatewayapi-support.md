---
title: Gateway API Support
---

As mentioned in the [system design](../../contributions/design/system-design) document, Envoy Gateway's managed data plane is configured dynamically through Kubernetes resources, primarily [Gateway API](https://gateway-api.sigs.k8s.io/) objects. Envoy Gateway supports configuration using the following Gateway API resources.

## GatewayClass

A [GatewayClass](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.GatewayClass) represents a "class" of gateways, i.e. which Gateways should be managed by Envoy Gateway. Envoy Gateway supports managing **a single** GatewayClass resource that matches its configured `controllerName` and follows Gateway API guidelines for [resolving conflicts](https://gateway-api.sigs.k8s.io/concepts/guidelines/?h=conflict#conflicts) when multiple GatewayClasses exist with a matching `controllerName`.

**Note:** If specifying GatewayClass [parameters reference](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.ParametersReference), it must refer to an [EnvoyProxy](../../api/extension_types#envoyproxy) resource.

## Gateway

When a [Gateway](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.Gateway) resource is created that references the managed GatewayClass, Envoy Gateway will create and manage a new Envoy Proxy deployment. Gateway API resources that reference this Gateway will configure this managed Envoy Proxy deployment.

## HTTPRoute

A [HTTPRoute](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.HTTPRoute) configures routing of HTTP traffic through one or more Gateways. The following HTTPRoute filters are supported by Envoy Gateway:

- `requestHeaderModifier`: [RequestHeaderModifiers](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.HTTPRouteFilter) can be used to modify or add request headers before the request is proxied to its destination.
- `responseHeaderModifier`: [ResponseHeaderModifiers](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.HTTPRouteFilter) can be used to modify or add response headers before the response is sent back to the client.
- `requestMirror`: [RequestMirrors](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.HTTPRouteFilter) configure destinations where the requests should also be mirrored to. Responses to mirrored requests will be ignored.
- `requestRedirect`: [RequestRedirects](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.HTTPRouteFilter) configure policied for how requests that match the HTTPRoute should be modified and then redirected.
- `urlRewrite`: [UrlRewrites](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.HTTPRouteFilter) allow for modification of the request's hostname and path before it is proxied to its destination.
- `extensionRef`: [ExtensionRefs](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.HTTPRouteFilterType) are used by Envoy Gateway to implement extended filters. Currently, Envoy Gateway supports rate limiting and request authentication filters. For more information about these filters, refer to the [rate limiting](../traffic/global-rate-limit) and [request authentication](../security/jwt-authentication) documentation.

**Notes:**
- The only [BackendRef](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.BackendRef) kind supported by Envoy Gateway is a [Service](https://kubernetes.io/docs/concepts/services-networking/service/). Routing traffic to other destinations such as arbitrary URLs is not possible.
- The `filters` field within [HTTPBackendRef](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.HTTPBackendRef) is not supported.

## TCPRoute

A [TCPRoute](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.TCPRoute) configures routing of raw TCP traffic through one or more Gateways. Traffic can be forwarded to the desired BackendRefs based on a TCP port number.

**Note:** A TCPRoute only supports proxying in non-transparent mode, i.e. the backend will see the source IP and port of the Envoy Proxy instance instead of the client.

## UDPRoute

A [UDPRoute](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.UDPRoute) configures routing of raw UDP traffic through one or more Gateways. Traffic can be forwarded to the desired BackendRefs based on a UDP port number.

**Note:** Similar to TCPRoutes, UDPRoutes only support proxying in non-transparent mode i.e. the backend will see the source IP and port of the Envoy Proxy instance instead of the client.

## GRPCRoute

A [GRPCRoute](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.GRPCRoute) configures routing of [gRPC](https://grpc.io/) requests through one or more Gateways. They offer request matching by hostname, gRPC service, gRPC method, or HTTP/2 Header. Envoy Gateway supports the following filters on GRPCRoutes to provide additional traffic processing:

- `requestHeaderModifier`: [RequestHeaderModifiers](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.GRPCRouteFilter) can be used to modify or add request headers before the request is proxied to its destination.
- `responseHeaderModifier`: [ResponseHeaderModifiers](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.GRPCRouteFilter) can be used to modify or add response headers before the response is sent back to the client.
- `requestMirror`: [RequestMirrors](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.GRPCRouteFilter) configure destinations where the requests should also be mirrored to. Responses to mirrored requests will be ignored.

**Notes:**
- The only [BackendRef](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.GRPCRouteFilter) kind supported by Envoy Gateway is a [Service](https://kubernetes.io/docs/concepts/services-networking/service/). Routing traffic to other destinations such as arbitrary URLs is not currently possible.
- The `filters` field within [HTTPBackendRef](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.HTTPBackendRef) is not supported.

## TLSRoute

A [TLSRoute](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.TLSRoute) configures routing of TCP traffic through one or more Gateways. However, unlike TCPRoutes, TLSRoutes can match against TLS-specific metadata.

## ReferenceGrant

A [ReferenceGrant](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.ReferenceGrant) is used to allow a resource to reference another resource in a different namespace. Normally an HTTPRoute created in namespace `foo` is not allowed to reference a Service in namespace `bar`. A ReferenceGrant permits these types of cross-namespace references. Envoy Gateway supports the following ReferenceGrant use-cases:

- Allowing an HTTPRoute, GRPCRoute, TLSRoute, UDPRoute, or TCPRoute to reference a Service in a different namespace.
- Allowing an HTTPRoute's `requestMirror` filter to include a BackendRef that references a Service in a different namespace.
- Allowing a Gateway's [SecretObjectReference](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.SecretObjectReference) to reference a secret in a different namespace.
