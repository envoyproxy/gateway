+++
title = "Envoy Gateway Resources"
+++

There are several resources that play a part in enabling you to meet your Kubernetes ingress traffic handling needs. This page provides a brief overview of the resources you’ll be working with.

## Overview

![](/img/envoy-gateway-resources-overview.png)

There are several resources that play a part in enabling you to meet your Kubernetes ingress traffic handling needs. This page provides a brief overview of the resources you’ll be working with.

### Kubernetes Gateway API Resources
- **GatewayClass:** Defines a class of Gateways with common configuration.
- **Gateway:** Specifies how traffic can enter the cluster.
- **Routes:** **HTTPRoute, GRPCRoute, TLSRoute, TCPRoute, UDPRoute:** Define routing rules for different types of traffic.

### Envoy Gateway (EG) API Resources
- **EnvoyProxy:** Represents the deployment and configuration of the Envoy proxy within a Kubernetes cluster, managing its lifecycle and settings.
- **EnvoyPatchPolicy, ClientTrafficPolicy, SecurityPolicy, BackendTrafficPolicy, EnvoyExtensionPolicy, BackendTLSPolicy:** Additional policies and configurations specific to Envoy Gateway.
- **Backend:** A resource that makes routing to cluster-external backends easier and makes access to external processes via Unix Domain Sockets possible.

| Resource                                                                | API         | Required | Purpose            | References             | Description                                                                                                                                                                                                 |
| ----------------------------------------------------------------------- | ----------- | -------- | ------------------ | ---------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [GatewayClass][1]                                                       | Gateway API | Yes      | Gateway Config     | Core                   | Defines a class of Gateways with common configuration.                                                                                                                                                      |
| [Gateway][2]                                                            | Gateway API | Yes      | Gateway Config     | GatewayClass           | Specifies how traffic can enter the cluster.                                                                                                                                                                |
| [HTTPRoute][3] [GRPCRoute][4] [TLSRoute][5] [TCPRoute][6] [UDPRoute][7] | Gateway API | Yes      | Routing            | Gateway                | Define routing rules for different types of traffic. **Note:**_For simplicity these resources are referenced collectively as Route in the References column_                                                |
| [Backend][8]                                                            | EG API      | No       | Routing            | N/A                    | Used for routing to cluster-external backends using FQDN or IP. Can also be used when you want to extend Envoy with external processes accessed via Unix Domain Sockets.                                    |
| [ClientTrafficPolicy][9]                                                | EG API      | No       | Traffic Handling   | Gateway                | Specifies policies for handling client traffic, including rate limiting, retries, and other client-specific configurations.                                                                                 |
| [BackendTrafficPolicy][10]                                              | EG API      | No       | Traffic Handling   | Gateway, Route         | Specifies policies for traffic directed towards backend services, including load balancing, health checks, and failover strategies. **Note:**_Most specific configuration wins_                             |
| [SecurityPolicy][11]                                                    | EG API      | No       | Security           | Gateway, Route         | Defines security-related policies such as authentication, authorization, and encryption settings for traffic handled by Envoy Gateway. **Note:**_Most specific configuration wins_                          |
| [BackendTLSPolicy][12]                                                  | Gateway API | No       | Security           | Service                | Defines TLS settings for backend connections, including certificate management, TLS version settings, and other security configurations. This policy is applied to Kubernetes Services.                     |
| [EnvoyProxy][13]                                                        | EG API      | No       | Customize & Extend | GatewayClass, Gateway  | The EnvoyProxy resource represents the deployment and configuration of the Envoy proxy itself within a Kubernetes cluster, managing its lifecycle and settings. **Note:**_Most specific configuration wins_ |
| [EnvoyPatchPolicy][14]                                                  | EG API      | No       | Customize & Extend | GatewayClass, Gateway  | This policy defines custom patches to be applied to Envoy Gateway resources, allowing users to tailor the configuration to their specific needs. **Note:**_Most specific configuration wins_                |
| [EnvoyExtensionPolicy][15]                                              | EG API      | No       | Customize & Extend | Gateway, Route, Backend| Allows for the configuration of Envoy proxy extensions, enabling custom behavior and functionality. **Note:**_Most specific configuration wins_                                                             |
| [HTTPRouteFilter][16]                                                   | EG API      | No       | Customize & Extend | HTTPRoute              | Allows for the additional request/response processing. |




[1]:	https://gateway-api.sigs.k8s.io/api-types/gatewayclass/
[2]:	https://gateway-api.sigs.k8s.io/api-types/gateway/
[3]:	https://gateway-api.sigs.k8s.io/api-types/httproute/
[4]:	https://gateway-api.sigs.k8s.io/api-types/grpcroute/
[5]:	https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.TLSRoute
[6]:	https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.TCPRoute
[7]:	https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.UDPRoute
[8]:	../tasks/traffic/backend
[9]:	../api/extension_types#clienttrafficpolicy
[10]:	../api/extension_types#backendtrafficpolicy
[11]:	../api/extension_types#securitypolicy
[12]:	https://gateway-api.sigs.k8s.io/api-types/backendtlspolicy/
[13]:	../api/extension_types#envoyproxy
[14]:	../api/extension_types#envoypatchpolicy
[15]:	../api/extension_types#envoyextensionpolicy
[16]:   ../api/extension_types#httproutefilter
