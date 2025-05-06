---
title: "The Gateway API"
---

## Before You Begin
You may want to be familiar with:
- [Kubernetes Gateway API](https://gateway-api.sigs.k8s.io/)
- [Kubernetes Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/)

## Overview

Kubernetes users may already be familiar with Ingress, which provides a simple mechanism for exposing HTTP(S) services to external traffic. While effective for basic use cases such as host and path-based routing, the Ingress API lacks advanced features like regex path matching. To compensate for this deficiency, custom annotations are needed. However, this non-standard approach leads to fragmentation across Ingress Controllers, making portability challenging.

The Gateway API is a modern Kubernetes API designed to provide a more consistent, expressive, and extensible method for managing network traffic into and within a Kubernetes cluster, compared to the legacy Ingress API. It introduces core resources such as `GatewayClass`, `Gateway`, and various route types like `HTTPRoute` and `TLSRoute`, which allow you to define how traffic is routed, secured, and exposed.

## Use Cases
Use The Gateway API to:
- Define how external traffic enters and is routed within your cluster
- Configure HTTP(S), TLS, and TCP traffic routing in a standardized, Kubernetes-native way
- Apply host-based, path-based, and header-based routing rules using HTTPRoute
- Terminate TLS at the edge using Gateway TLS configuration
- Separate responsibilities between infrastructure and application teams through role-oriented resource design
- Improve portability and consistency across different gateway implementations

## The Gateway API in Envoy Gateway
In essence, the Gateway API provides a standard interface, and Envoy Gateway adds production-grade capabilities to that interface—bridging the gap between simplicity and power while keeping everything Kubernetes-native.

One of the key strengths of the Gateway API is its extensibility. While it provides a solid foundation for routing and traffic control, it also enables implementations to introduce custom resources that enhance and tailor functionality for specific use cases.

Envoy Gateway leverages this model by introducing a suite of Gateway API extensions—implemented as Kubernetes Custom Resource Definitions (CRDs)—to expose powerful features from Envoy Proxy. These features include enhanced support for rate limiting, authentication, traffic shaping, and more. By utilizing these extensions, users can access production-grade functionality in a Kubernetes-native and declarative manner, without needing to write low-level Envoy configuration.

## Related Resources

- [Getting Started with Envoy Gateway](../../tasks/quickstart.md)
- [Envoy Gateway API Reference](../../api/extension_types)
- [Extensibility Tasks](../../tasks/extensibility/_index.md)
