---
title: "Gateway API Extensions"
weight: 2
---
## Before You Begin
- [The Gateway API](https://gateway-api.sigs.k8s.io/)

## Overview
Gateway API Extensions let you configure extra features that aren’t part of the standard Kubernetes Gateway API. These extensions are built by the teams that create and maintain Gateway API implementations.
The Gateway API was designed to be extensible safe, and reliable. In the old Ingress API, people had to use custom annotations to add new features, but those weren’t type-safe, making it hard to check if their configuration was correct.
With Gateway API Extensions, implementers provide type-safe Custom Resource Definitions (CRDs). This means every configuration you write has a clear structure and strict rules, making it easier to catch mistakes early and be confident your setup is valid.
## Use Cases

Here are some examples of what kind of features extensions include:

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

The Envoy Gateway API introduces a set of Gateway API extensions that enable users to leverage the power of the Envoy proxy. Envoy Gateway uses a policy attachment model, where custom policies are applied to standard Gateway API resources (like HTTPRoute or Gateway) without modifying the core API. This approach provides separation of concerns and makes it easier to manage configurations across teams.

{{% alert title="Current Extensions" color="info" %}}
Currently supported extensions include
[`Backend`](../../api/extension_types#backend),
[`BackendTrafficPolicy`](../../api/extension_types#backendtrafficpolicy),
[`ClientTrafficPolicy`](../../api/extension_types#clienttrafficpolicy),
[`EnvoyExtensionPolicy`](../../api/extension_types#envoyextensionpolicy),
[`EnvoyGateway`](../../api/extension_types#envoygateway),
[`EnvoyPatchPolicy`](../../api/extension_types#envoypatchpolicy),
[`EnvoyProxy`](../../api/extension_types#envoyproxy),
[`HTTPRouteFilter`](../../api/extension_types#httproutefilter), and
[`SecurityPolicy`](../../api/extension_types#securitypolicy),
{{% /alert %}}

These extensions are processed through Envoy Gateway's control plane, translating them into xDS configurations applied to Envoy Proxy instances. This layered architecture allows for consistent, scalable, and production-grade traffic control without needing to manage raw Envoy configuration directly.

## Related Resources
