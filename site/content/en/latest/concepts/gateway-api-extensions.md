---
title: "Gateway API Extensions"
---
## Before you Begin
- [The Gateway API](https://gateway-api.sigs.k8s.io/)

## Overview

Gateway Extensions are mechanisms that allow implementation-specific features to be added to the Kubernetes Gateway API without modifying its core specification. They enable Gateway API implementations like Envoy Gateway to expose their unique capabilities while maintaining compatibility with the standard API. Extensions follow defined patterns that preserve the role-oriented design of Gateway API while allowing for customization to address specific use cases not covered by the core API.

## Use Cases

1. **Advanced Traffic Management:** 
    Implementing sophisticated load balancing algorithms, circuit breaking, or retries not defined in the core API
2. **Enhanced Security Controls:** 
    Adding implementation-specific TLS configurations, authentication mechanisms, or access control rules
3. **Observability Integration:** 
    Connecting Gateway resources to monitoring systems, logging pipelines, or tracing frameworks
4. **Custom Protocol Support:** 
    Extending beyond HTTP/TCP/UDP with specialized protocol handling
5. **Rate Limiting and Compression:** 
    Implementing traffic policing specific to the implementation's capabilities

## Gateway API Extensions in Envoy Gateway

The Envoy Gateway API is a set of Gateway API extensions that enable advanced traffic management capabilities in Envoy Gateway. These extensions build on top of the Kubernetes Gateway API by introducing custom resources that expose powerful features of Envoy Proxy in a Kubernetes-native way.

Envoy Gateway uses a policy attachment model, where custom policies are applied to standard Gateway API resources (like HTTPRoute or Gateway) without modifying the core API. This approach provides separation of concerns and makes it easier to manage configurations across teams.

{{% alert title="Current Extensions" color="info" %}}
Currently supported extensions include [`ClientTrafficPolicy`](../api/extension_types#clienttrafficpolicy), [`BackendTrafficPolicy`](../api/extension_types#backendtrafficpolicy), [`SecurityPolicy`](../api/extension_types#securitypolicy), [`EnvoyExtensionPolicy`](../api/extension_types#envoyextensionpolicy), [`EnvoyProxy`](../api/extension_types#envoyproxy), [`HTTPRouteFilter`](../api/extension_types#httproutefilter), and [`Backend`](../api/extension_types#backend).
{{% /alert %}}

These extensions are processed through Envoy Gateway's control plane, which translates them into xDS configurations that are applied to Envoy Proxy instances. This layered architecture allows for consistent, scalable, and production-grade traffic control without needing to manage raw Envoy configuration directly.

## Related Resources
- [ClientTrafficPolicy](client-traffic-policy.md)
- [BackendTrafficPolicy](backend-traffic-policy.md)
- [SecurityPolicy](security-policy.md)