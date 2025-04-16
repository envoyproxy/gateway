---
title: "Envoy Proxy"
---

## Before You Begin
You may want to be familiar with:
- [What is a Proxy?](https://developer.mozilla.org/en-US/docs/Glossary/Proxy_server)

## Overview

A proxy server acts as an intermediary between a client (like a web browser) and another server (such as a website). When the client makes a request, the proxy forwards it to the destination server, receives the response, and then sends it back to the client.

Proxies are commonly used to enhance security, manage traffic, anonymize user activity, or optimize performance through features like caching and load balancing. In modern cloud environments, proxies often handle critical tasks such as request routing, TLS termination, authentication, and traffic shaping.

Envoy Proxy is a high-performance, open source edge and service proxy built for cloud-native applications. Originally developed at Lyft and now part of the Cloud Native Computing Foundation (CNCF), Envoy is widely adopted as a data plane component in service meshes, API gateways, and ingress solutions.

Thanks to its powerful features and extensibility, Envoy is often embedded in larger platforms—like Istio, Consul, and Envoy Gateway—where it takes care of the complex traffic management and security tasks behind the scenes.

## Use Cases

Use Envoy Proxy when you need:
- A powerful L3/L4/L7 proxy to manage internal or external traffic
- Fine-grained control over HTTP/gRPC/TLS request routing
- Built-in observability (metrics, tracing, logging) for all traffic
- Smart load balancing and failover strategies
- A proxy that integrates with service meshes or API gateways

## How Envoy Proxy fits into Envoy Gateway

Envoy Gateway uses Envoy Proxy as its data plane. That means Envoy Proxy is the component actually handling traffic—terminating TLS, routing requests, and applying policies.

Here’s how the interaction works:

1. You define traffic rules using Kubernetes Gateway API resources like `Gateway`, `HTTPRoute`, or Envoy Gateway's own CRDs (e.g., `RateLimitPolicy`, `AuthenticationPolicy`).

2. Envoy Gateway translates those rules into Envoy Proxy configuration behind the scenes.

3. Envoy Proxy enforces those rules, acting on real-world traffic—balancing requests, rejecting unauthorized ones, collecting metrics, etc.

This separation of concerns allows users to configure traffic behavior declaratively (with CRDs), while leveraging Envoy Proxy's robust capabilities under the hood.


## Related Resources

- [Getting Started with Envoy Gateway](../../tasks/quickstart)
- [API Gateway](../api_gateway)
