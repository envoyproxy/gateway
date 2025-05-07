---
title: "Envoy Proxy"
---

## Overview

A proxy server acts as an intermediary between a client (like a web browser) and another server (such as a website). When the client makes a request, the proxy forwards it to the destination server, receives the response, and then sends it back to the client.

Proxies are commonly used to enhance security, manage traffic, anonymize user activity, or optimize performance through features like caching and load balancing. In cloud environments, they often handle critical tasks such as request routing, TLS termination, authentication, and traffic shaping.

Envoy Proxy is a high-performance, open-source proxy designed for cloud-native applications. Originally developed at Lyft and now a graduated project in the Cloud Native Computing Foundation (CNCF), Envoy supports use cases for edge and service proxies, routing traffic at the system’s boundary or between internal services.

## Use Cases

Use Envoy Proxy to:
- Manage internal or external traffic with a powerful L3/L4/L7 proxy
- Control HTTP, gRPC, or TLS routing with fine-grained match and rewrite rules
- Gain full observability via built-in metrics, tracing, and logging
- Implement smart load balancing and resilient failover strategies
- Integrate seamlessly with service meshes, API gateways, and other control planes

## Envoy Proxy in Envoy Gateway

Because of its rich feature set and extensibility,Envoy Gateway uses Envoy Proxy as its data plane. That means Envoy Proxy is the component actually handling traffic—terminating TLS, routing requests, and applying policies.

Here’s how the interaction works:

1. You define traffic rules using Kubernetes Gateway API resources like `Gateway`, `HTTPRoute`, or Envoy Gateway's own CRDs (e.g., `BackendTrafficPolicy`, `ClientTrafficPolicy`) implemented as Gateway API extensions.

2. Envoy Gateway translates those rules into Envoy Proxy configuration behind the scenes.

3. Envoy Proxy enforces those rules, acting on real-world traffic—balancing requests, rejecting unauthorized ones, collecting metrics, etc.

This separation of concerns allows users to configure traffic behavior declaratively (with CRDs), while leveraging Envoy Proxy's robust capabilities under the hood.


## Related Resources
- [Getting Started with Envoy Gateway](../../tasks/quickstart.md)
- [Envoy Proxy](https://www.envoyproxy.io/)
