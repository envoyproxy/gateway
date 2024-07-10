# Envoy Gateway Resources

There are several resources that play a part in enabling you to meet your Kubernetes ingress traffic handling needs. This page provides a brief overview of the resources you’ll be working with.

# Overview

## Kubernetes Gateway API Resources
- **GatewayClass:** Defines a class of Gateways with common configuration.
- **Gateway:** Specifies how traffic can enter the cluster.
- **Routes:** **HTTPRoute, GRPCRoute, TLSRoute, TCPRoute, UDPRoute:** Define routing rules for different types of traffic.
## Envoy Gateway (EG) API Resources
- **EnvoyProxy:** Represents the deployment and configuration of the Envoy proxy within a Kubernetes cluster, managing its lifecycle and settings.
- **EnvoyGatewayPatchPolicy, ClientTrafficPolicy, SecurityPolicy, BackendTrafficPolicy, EnvoyExtensionPolicy, BackendTLSPolicy:** Additional policies and configurations specific to Envoy Gateway.

| API         | Resource                                       | Required | Purpose            | References           | Description                                                                                                                                                                                                 |
| ----------- | ---------------------------------------------- | -------- | ------------------ | -------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Gateway API | GatewayClass                                   | Yes      | Gateway Config     | Core                 | Defines a class of Gateways with common configuration.                                                                                                                                                      |
| Gateway API | Gateway                                        | Yes      | Gateway Config     | GatewayClass         | Specifies how traffic can enter the cluster.                                                                                                                                                                |
| Gateway API | HTTPRoute GRPCRoute TLSRoute TCPRoute UDPRoute | Yes      | Routing            | Gateway              | Define routing rules for different types of traffic. **Note:**_For simplicity these resources are referenced collectively as Route in the References column_                                                |
| EG API      | ClientTrafficPolicy                            | No       | Traffic Handling   | Gateway              | Specifies policies for handling client traffic, including rate limiting, retries, and other client-specific configurations.                                                                                 |
| EG API      | BackendTrafficPolicy                           | No       | Traffic Handling   | Gateway Route        | Specifies policies for traffic directed towards backend services, including load balancing, health checks, and failover strategies. **Note:**_Most specific configuration wins_                             |
| EG API      | SecurityPolicy                                 | No       | Security           | Gateway Route        | Defines security-related policies such as authentication, authorization, and encryption settings for traffic handled by Envoy Gateway. **Note:**_Most specific configuration wins_                          |
| Gateway API | BackendTLSPolicy                               | No       | Security           | Service              | Defines TLS settings for backend connections, including certificate management, TLS version settings, and other security configurations. This policy is applied to Kubernetes Services.                     |
| EG API      | EnvoyProxy                                     | No       | Customize & Extend | GatewayClass Gateway | The EnvoyProxy resource represents the deployment and configuration of the Envoy proxy itself within a Kubernetes cluster, managing its lifecycle and settings. **Note:**_Most specific configuration wins_ |
| EG API      | EnvoyGatewayPatchPolicy                        | No       | Customize & Extend | GatewayClass Gateway | This policy defines custom patches to be applied to Envoy Gateway resources, allowing users to tailor the configuration to their specific needs. **Note:**_Most specific configuration wins_                |
| EG API      | EnvoyExtensionPolicy                           | No       | Customize & Extend | Gateway Route        | Allows for the configuration of Envoy proxy extensions, enabling custom behavior and functionality. **Note:**_Most specific configuration wins_                                                             |

# Resources Relationship Diagram

![resources-visual](/img/envoy-gateway-resources-visual.png)
