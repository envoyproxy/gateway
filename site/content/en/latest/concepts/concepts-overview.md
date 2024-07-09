# Envoy Gateway Resources

There are several resources that play a part in enabling you to meet your Kubernetes ingress traffic handling needs. This page provides a brief overview of the resources you’ll be working with.

# High Level Summary

## Kubernetes Gateway API Resources
- **GatewayClass:** Defines a class of Gateways with common configuration.
- **Gateway: **Specifies how traffic can enter the cluster.
- **HTTPRoute, GRPCRoute, TLSRoute, TCPRoute, UDPRoute:** Define routing rules for different types of traffic.
## Envoy Gateway (EG) API Resources
- **EnvoyProxy:** Represents the deployment and configuration of the Envoy proxy within a Kubernetes cluster, managing its lifecycle and settings.
- **EnvoyGatewayPatchPolicy, ClientTrafficPolicy, SecurityPolicy, BackendTrafficPolicy, EnvoyExtensionPolicy, BackendTLSPolicy:** Additional policies and configurations specific to Envoy Gateway.

![resources](/img/envoy-gateway-resources-overview.png)

# Required Resources
The below listed Resources are necessary for you to create a minimal deployment of Envoy Gateway in your Kubernetes cluster.

## EnvoyProxy 
The EnvoyProxy resource represents the deployment and configuration of the Envoy proxy itself within a Kubernetes cluster, managing its lifecycle and settings.

## GatewayClass
A GatewayClass resource defines a class of Gateways that share a common configuration and behavior. It acts as a template for creating Gateway resources.

## Gateway
The Gateway resource specifies how traffic can be routed at the edge of the Kubernetes cluster. It acts as a top-level resource that binds various routing rules and policies.

## At least one Route
You need at least one route to define how incoming traffic should be handled. The specific type of route will depend on the traffic requirements:

### Routes
Routes are used to define how traffic should be routed based on different criteria. There are several types of routes.

#### HTTPRoute
Defines routing rules for HTTP traffic, including path matching, header manipulation, and forwarding policies.

#### GRPCRoute
Defines routing rules for gRPC traffic, allowing for matching on gRPC-specific attributes and forwarding policies.

#### TLSRoute
Defines routing rules for TLS traffic, including SNI matching and TLS termination settings.

#### TCPRoute
Defines routing rules for TCP traffic, specifying how connections should be handled and forwarded.

#### UDPRoute
Defines routing rules for UDP traffic, detailing how packets should be processed and forwarded.

# Optional

## Traffic Management

### ClientTrafficPolicy
Specifies policies for handling client traffic, including rate limiting, retries, and other client-specific configurations.

### BackendTrafficPolicy
Specifies policies for traffic directed towards backend services, including load balancing, health checks, and failover strategies.

## Security

### SecurityPolicy
Defines security-related policies such as authentication, authorization, and encryption settings for traffic handled by Envoy Gateway.

### BackendTLSPolicy
Defines TLS settings for backend connections, including certificate management, TLS version settings, and other security configurations. This policy is applied to Kubernetes Services.

## Customize & Extend Envoy

### EnvoyGatewayPatchPolicy
This policy defines custom patches to be applied to Envoy Gateway resources, allowing users to tailor the configuration to their specific needs.

### EnvoyExtensionPolicy
Allows for the configuration of Envoy proxy extensions, enabling custom behavior and functionality.