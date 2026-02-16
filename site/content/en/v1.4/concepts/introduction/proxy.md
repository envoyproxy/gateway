---
title: "Proxy"
weight: 4
---

## Overview
**A proxy server is an intermediary between a client (like a web browser) and another server (like an API server).** When the client makes a request, the proxy forwards it to the destination server, receives the response, and then sends it back to the client.

Proxies are used to enhance security, manage traffic, anonymize user activity, or optimize performance through caching and load balancing features. In cloud environments, they often handle critical tasks such as request routing, TLS termination, authentication, and traffic shaping.

## Use Cases

**Use Envoy Proxy to:**
- Manage internal or external traffic with a powerful L3/L4/L7 proxy
- Control HTTP, gRPC, or TLS routing with fine-grained match and rewrite rules
- Gain full observability via built-in metrics, tracing, and logging
- Implement intelligent load balancing and resilient failover strategies
- Integrate seamlessly with service meshes, API gateways, and other control planes

## Proxy in Envoy Gateway
Envoy Gateway is a system made up of two main parts:
- A _data plane_, which handles the actual network traffic
- A _control plane_, which manages and configures the _data plane_

Envoy Gateway uses the Envoy Proxy, which was originally developed at Lyft. This proxy is the foundation of the Envoy project, of which Envoy Gateway is a part, and is now a graduated project within the Cloud Native Computing Foundation (CNCF).

Envoy Proxy is a high-performance, open-source proxy designed for cloud-native applications. Envoy supports use cases for edge and service proxies, routing traffic at the system’s boundary or between internal services.

The control plane uses the Kubernetes Gateway API to understand your settings and then translates them into the format Envoy Proxy needs (called _xDS configuration_). It also runs and updates the Envoy Proxy instances inside your Kubernetes cluster.

## Related Resources
- [Getting Started with Envoy Gateway](../../tasks/quickstart.md)
- [Envoy Proxy](https://www.envoyproxy.io/)
