# Gateway API Support

As mentioned in the [system design][] document, Envoy Gateway's managed data plane is configured dynamically through 
Kubernetes resources, primarily [Gateway API][] objects. Envoy Gateway supports configuration using the following Gateway API resources.

## **GatewayClass**

A [GatewayClass][] is used to configure which Gateways and other reliant resources should be managed by Envoy Gateway.
Envoy Gateway supports a single GatewayClass resource linked to the Envoy Gateway controller and accepts in order of age (oldest first) if there are multiple.
The [ParametersReference][] on the GatewayClass must refer to an EnvoyProxy.

## **Gateway**

When a [Gateway][] resource is created that references the GatwewayClass Envoy Gateway is managing then Envoy Gateway will 
create and manage a new Envoy Proxy deployment. All other Gateway API resources that are managed by this Gateway will be used
to configure the Envoy Proxy deployment that it created. Envoy Gateway does not support Multiple certificate references or  Specifying an [address][]
for the Gateway.

## **HTTPRoute**

[HTTPRoutes][] are supported as the primary way to configure HTTP traffic in Envoy Gateway.
All of the following HTTPRoute filters are supported by Envoy Gateway.

- `requestHeaderModifier`: [RequestHeaderModifiers](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1beta1.HTTPRouteFilter) can be used to modify or add request headers before the request is proxied to its destination.
- `responseHeaderModifier`: [ResponseHeaderModifiers](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1beta1.HTTPRouteFilter) can be used to modify or add response headers before the response is sent back to the client.
- `requestMirror`: [RequestMirrors](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1beta1.HTTPRouteFilter) configure destinations where the requests should also be mirrored to. Responses to mirrored requests will be ignored.
- `requestRedirect`: [RequestRedirects](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1beta1.HTTPRouteFilter) configure policied for how requests that match the HTTPRoute should be modified and then redirected.
- `urlRewrite`: [UrlRewrites](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1beta1.HTTPRouteFilter) allow for modification of the request's hostname and path before it is proxied to its destination.
- `extensionRef`: [ExtensionRefs] are used by Envoy Gateway to add additional support for Ratelimitg and Authentication. For more information about Envoy Gateay's implementation of these filters please refer to the [Ratelimiting][] and [Authentication][] documentation.

**Note:** currently the only [BackendRef][] kind (the destination where traffic should be sent to) that Envoy Gateway supports are [Kubernetes Services][]. Routing traffic to other destinations such as arbitrary URLs is not currently possible.

## **TCPRoute**

[TCPRoutes][] are used to configure routing of raw TCP traffic. Traffic can be forwarded to the desired BackendRef(s) based on a port.

**Note:** TCPRoutes only support proxying in non-transparent mode i.e. the backend will see the source IP and port of the deployed
Envoy instance instead of the client.

## **UDPRoute**

[UDPRoutes][] are used to configure routing of raw UDP traffic. Traffic can be forwarded to the desired BackendRef(s) based on a port.

**Note:** Similar to TCPRoutes, UDPRoutes only support proxying in non-transparent mode i.e. the backend will see the source IP and port of the deployed
Envoy instance instead of the client.

## **GRPCRoute**

[GRPCRoutes][] configure routing of [gRPC][] requests. They offer request matching by hostname, gRPC service, gRPC method, or HTTP/2 Header.
Similar to HTTPRoutes, Envoy Gateway supports the following filters on GRPCRoutes to provide additional traffic processing.

- `requestHeaderModifier`: [RequestHeaderModifiers](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1alpha2.GRPCRouteFilter) can be used to modify or add request headers before the request is proxied to its destination.
- `responseHeaderModifier`: [ResponseHeaderModifiers](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1alpha2.GRPCRouteFilter) can be used to modify or add response headers before the response is sent back to the client.
- `requestMirror`: [RequestMirrors](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1alpha2.GRPCRouteFilter) configure destinations where the requests should also be mirrored to. Responses to mirrored requests will be ignored.
- `extensionRef`: [ExtensionRefs] are used by Envoy Gateway to add additional support for Ratelimitg and Authentication. For more information about Envoy Gateay's implementation of these filters please refer to the [Ratelimiting][] and [Authentication][] documentation.

**Note:** currently the only [BackendRef](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1alpha2.GRPCRouteFilter) kind (the destination where traffic should be sent to) that Envoy Gateway supports are [Kubernetes Services][]. Routing traffic to other destinations such as arbitrary URLs is not currently possible

## **TLSRoute**

[TLSRoutes][] are used similarly to TCPRoutes to configure routing of TCP traffic; however, unlike TCPRoutes, TLSRoutes can match against TLS-Specific Metadata.

## **ReferenceGrant**

[ReferenceGrants][] are used as a way to configure which resources in other namespaces are allowed to reference specific kinds of resources in
the namespace of the ReferenceGrant. Normally an HTTPRoute created in namespace `foo` is not allowed to specify a Service in the `bar` namespace as the
one of its BackendRefs. ReferenceGrants are commonly used to permit these types of cross-namespace references. Envoy Gateway supports the following use-cases for ReferenceGrants.

- Allowing an HTTPRoute, GRPCRoute, TLSRoute, UDPRoute, or TCPRoute to include a BackendRef that references a Service that is not in the same namespace as the HTTPRoute.
- Allowing an HTTPRoute's `requestMirror` filter to include a BackendRef that references a Service that is not in the same namespace as the HTTPRoute.
- Allowing a Gateway's [SecretObjectReference][] to reference a secret that is not in the same namespace as the Gateway when configuring TLS on a Gateway.

[System Design]: https://gateway.envoyproxy.io/latest/design/system-design.html
[Gateway API]: https://gateway-api.sigs.k8s.io/
[GatewayClass]: https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1beta1.GatewayClass
[ParametersReference]: https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1beta1.ParametersReference
[Gateway]: https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1beta1.Gateway
[address]: https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1beta1.GatewayAddress
[HTTPRoutes]: https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1beta1.HTTPRoute
[Kubernetes Services]: https://kubernetes.io/docs/concepts/services-networking/service/
[BackendRef]: https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1beta1.BackendRef
[TCPRoutes]: https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1alpha2.TCPRoute
[UDPRoutes]: https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1alpha2.UDPRoute
[GRPCRoutes]: https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1alpha2.GRPCRoute
[gRPC]: https://grpc.io/
[TLSRoutes]: https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1alpha2.TLSRoute
[ReferenceGrants]: https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io%2fv1beta1.ReferenceGrant
[SecretObjectReference]: https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1beta1.SecretObjectReference
[Ratelimiting]: https://gateway.envoyproxy.io/latest/user/rate-limit.html
[Authentication]: https://gateway.envoyproxy.io/latest/user/authn.html
