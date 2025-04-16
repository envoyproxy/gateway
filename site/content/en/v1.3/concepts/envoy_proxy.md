---
title: "Envoy Proxy"
---

## Before You Begin
You may want to be familiar with:
- [What is a Proxy?](https://developer.mozilla.org/en-US/docs/Glossary/Proxy_server)

## Overview

Envoy Proxy is a high-performance, open source edge and service proxy designed for cloud-native applications. Originally developed at Lyft, Envoy is now part of the CNCF and widely used as a data plane component in service meshes, API gateways, and ingress solutions.

Envoy functions as a powerful intermediary for service-to-service and edge traffic—managing, routing, and observing requests with built-in support for features like load balancing, retries, circuit breaking, traffic shifting, and comprehensive observability.

Because of its rich feature set and extensibility, Envoy is often embedded in larger systems (like Istio, Consul, or Envoy Gateway) where it takes care of the heavy lifting for traffic management and security.

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
