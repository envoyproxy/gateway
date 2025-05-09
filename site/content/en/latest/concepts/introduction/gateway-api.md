---
title: "The Gateway API"
---

## Before You Begin
You may want to be familiar with:
- [Kubernetes Gateway API](https://gateway-api.sigs.k8s.io/)
- [Kubernetes Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/)

## Overview
The Gateway API is a Kubernetes API designed to provide a consistent, expressive, and extensible method for managing network traffic into and within a Kubernetes cluster, compared to the legacy Ingress API. It introduces core resources such as `GatewayClass` and `Gateway` and various route types like `HTTPRoute` and `TLSRoute`, which allow you to define how traffic is routed, secured, and exposed.

The Gateway API succeeds the Ingress API, which many Kubernetes users may already be familiar with. The Ingress API provided a mechanism for exposing HTTP(S) services to external traffic. The lack of advanced features like regex path matching led to custom annotations to compensate for these deficiencies. This non-standard approach led to fragmentation across Ingress Controllers, challenging portability.

## Use Cases
Use The Gateway API to:
- Define how external traffic enters and is routed within your cluster
- Configure HTTP(S), TLS, and TCP traffic routing in a standardized, Kubernetes-native way
- Apply host-based, path-based, and header-based routing rules using HTTPRoute
- Terminate TLS at the edge using Gateway TLS configuration
- Separate responsibilities between infrastructure and application teams through role-oriented resource design
- Improve portability and consistency across different gateway implementations

## The Gateway API in Envoy Gateway
In essence, the Gateway API provides a standard interface. Envoy Gateway adds production-grade capabilities to that interface, bridging the gap between simplicity and power while keeping everything Kubernetes-native.

One of the Gateway API's key strengths is that implementers can extend it. While providing a foundation for standard routing and traffic control needs, it enables implementations to introduce custom resources that address specific use cases.

Envoy Gateway leverages this model by introducing a suite of Gateway API extensions—implemented as Kubernetes Custom Resource Definitions (CRDs)—to expose powerful features from Envoy Proxy. These features include enhanced support for rate limiting, authentication, traffic shaping, and more. By utilizing these extensions, users can access production-grade functionality in a Kubernetes-native and declarative manner, without needing to write a low-level Envoy configuration.

## Related Resources
- [Getting Started with Envoy Gateway](../../tasks/quickstart.md)
- [Envoy Gateway API Reference](../../api/extension_types)
- [Extensibility Tasks](../../tasks/extensibility/_index.md)
