# Envoy Gateway Resources

There are several resources that play a part in enabling you to meet your Kubernetes ingress traffic handling needs. This page provides a brief overview of the resources you’ll be working with.

# Overview

## Kubernetes Gateway API Resources
- **GatewayClass:** Defines a class of Gateways with common configuration.
- **Gateway:** Specifies how traffic can enter the cluster.
- **Routes:** **HTTPRoute, GRPCRoute, TLSRoute, TCPRoute, UDPRoute:** Define routing rules for different types of traffic.
## Envoy Gateway (EG) API Resources
- **EnvoyProxy:** Represents the deployment and configuration of the Envoy proxy within a Kubernetes cluster, managing its lifecycle and settings.
- **EnvoyPatchPolicy, ClientTrafficPolicy, SecurityPolicy, BackendTrafficPolicy, EnvoyExtensionPolicy, BackendTLSPolicy:** Additional policies and configurations specific to Envoy Gateway.

| Resource                                       | API         | Required | Purpose            | References           | Description                                                                                                                                                                                                 |
| ---------------------------------------------- | ----------- | -------- | ------------------ | -------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [GatewayClass][]                                   | Gateway API | Yes      | Gateway Config     | Core                 | Defines a class of Gateways with common configuration.                                                                                                                                                      |
| [Gateway][]                                        | Gateway API | Yes      | Gateway Config     | GatewayClass         | Specifies how traffic can enter the cluster.                                                                                                                                                                |
| [HTTPRoute][] [GRPCRoute][] [TLSRoute][] [TCPRoute][] [UDPRoute][] | Gateway API | Yes      | Routing            | Gateway              | Define routing rules for different types of traffic. **Note:**_For simplicity these resources are referenced collectively as Route in the References column_                                                |
| [ClientTrafficPolicy][]                            | EG API      | No       | Traffic Handling   | Gateway              | Specifies policies for handling client traffic, including rate limiting, retries, and other client-specific configurations.                                                                                 |
| [BackendTrafficPolicy][]                           | EG API      | No       | Traffic Handling   | Gateway Route        | Specifies policies for traffic directed towards backend services, including load balancing, health checks, and failover strategies. **Note:**_Most specific configuration wins_                             |
| [SecurityPolicy][]                                 | EG API      | No       | Security           | Gateway Route        | Defines security-related policies such as authentication, authorization, and encryption settings for traffic handled by Envoy Gateway. **Note:**_Most specific configuration wins_                          |
| [BackendTLSPolicy][]                               | Gateway API | No       | Security           | Service              | Defines TLS settings for backend connections, including certificate management, TLS version settings, and other security configurations. This policy is applied to Kubernetes Services.                     |
| [EnvoyProxy][]                                     | EG API      | No       | Customize & Extend | GatewayClass Gateway | The EnvoyProxy resource represents the deployment and configuration of the Envoy proxy itself within a Kubernetes cluster, managing its lifecycle and settings. **Note:**_Most specific configuration wins_ |
| [EnvoyPatchPolicy][]                               | EG API      | No       | Customize & Extend | GatewayClass Gateway | This policy defines custom patches to be applied to Envoy Gateway resources, allowing users to tailor the configuration to their specific needs. **Note:**_Most specific configuration wins_                |
| [EnvoyExtensionPolicy][]                           | EG API      | No       | Customize & Extend | Gateway Route        | Allows for the configuration of Envoy proxy extensions, enabling custom behavior and functionality. **Note:**_Most specific configuration wins_                                                             |



[BackendTrafficPolicy]: ../api/extension_types#backendtrafficpolicy
[ClientTrafficPolicy]: ../api/extension_types#clienttrafficpolicy
[SecurityPolicy]: ../api/extension_types#securitypolicy
[EnvoyProxy]: ../api/extension_types#envoyproxy
[SecurityPolicy]: ../api/extension_types#securitypolicy
[EnvoyPatchPolicy]: ../api/extension_types#envoypatchpolicy
[EnvoyExtensionPolicy]: ../api/extension_types#envoyextensionpolicy
[Gateway]: https://gateway-api.sigs.k8s.io/api-types/gateway/
[GatewayClass]: https://gateway-api.sigs.k8s.io/api-types/gatewayclass/
[HTTPRoute]: https://gateway-api.sigs.k8s.io/api-types/httproute/
[GRPCRoute]: https://gateway-api.sigs.k8s.io/api-types/grpcroute/
[TLSRoute]: https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.TLSRoute
[UDPRoute]: https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.UDPRoute
[TCPRoute]: https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.TCPRoute
[BackendTLSPolicy]:https://gateway-api.sigs.k8s.io/api-types/backendtlspolicy/

