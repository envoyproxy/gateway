---
title: "Envoy Proxy"
---

## Overview

A proxy server acts as an intermediary between a client (like a web browser) and another server (such as a website). When the client makes a request, the proxy forwards it to the destination server, receives the response, and then sends it back to the client.

Proxies are commonly used to enhance security, manage traffic, anonymize user activity, or optimize performance through features like caching and load balancing. In modern cloud environments, they often handle critical tasks such as request routing, TLS termination, authentication, and traffic shaping.

Envoy Proxy is a high-performance, open source proxy built for cloud-native applications. Originally developed at Lyft and now a graduated project in the Cloud Native Computing Foundation (CNCF), Envoy supports both edge proxy and service proxy use cases—routing traffic at the boundary of a system or between internal services.

Because of its rich feature set and extensibility, Envoy is widely used as a data plane component in platforms like Istio, Consul, and Envoy Gateway, where it handles complex networking and security operations behind the scenes.

## Use Cases

Use Envoy Proxy when you need:
- A powerful L3/L4/L7 proxy to manage internal or external traffic
- Fine-grained control over HTTP/gRPC/TLS request routing
- Built-in observability (metrics, tracing, logging) for all traffic
- Smart load balancing and failover strategies
- A proxy that integrates with service meshes or API gateways

## Configuration in Envoy Gateway

Envoy Gateway uses Envoy Proxy as its data plane. That means Envoy Proxy is the component actually handling traffic—terminating TLS, routing requests, and applying policies.

Here’s how the interaction works:

1. You define traffic rules using Kubernetes Gateway API resources like `Gateway`, `HTTPRoute`, or Envoy Gateway's own CRDs (e.g., `RateLimitPolicy`, `AuthenticationPolicy`).

2. Envoy Gateway translates those rules into Envoy Proxy configuration behind the scenes.

3. Envoy Proxy enforces those rules, acting on real-world traffic—balancing requests, rejecting unauthorized ones, collecting metrics, etc.

This separation of concerns allows users to configure traffic behavior declaratively (with CRDs), while leveraging Envoy Proxy's robust capabilities under the hood.


## Related Resources

- [Getting Started with Envoy Gateway](../tasks/quickstart.md)
- [API Gateway](api_gateways.md)
