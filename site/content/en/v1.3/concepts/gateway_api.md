---
title: "The Gateway API"
---

## Before You Begin
You may want to be familiar with:
- [Kubernetes Gateway API](https://gateway-api.sigs.k8s.io/)

## Overview

The **Kubernetes Gateway API** is a set of standard APIs designed to provide a unified way to manage traffic into and within a Kubernetes cluster. It defines resources like `GatewayClass`, `Gateway`, and various route types (e.g., `HTTPRoute`, `TLSRoute`) that allow users to configure how traffic is routed and managed. This API is designed to be extensible, allowing projects like Envoy Gateway to build on top of it to provide additional functionality.

The **Envoy Gateway API** is a set of Kubernetes Custom Resource Definitions (CRDs) created by the Envoy Gateway project. It enhances the Kubernetes Gateway API by adding features that are essential for real-world production traffic management, such as security policies, rate limits, and traffic shaping. 

Gateway APIs are useful because they provide a standardized, Kubernetes-native way to manage traffic. They allow users to define their desired behavior using familiar Kubernetes manifests, making it easier to configure and manage traffic routing and policies. In the context of the Envoy Gateway project, the Gateway API enables users to take advantage of Envoy Proxy's advanced features in a user-friendly way, without needing to write complex Envoy configurations.

## Use Cases

Use the Envoy Gateway API when:
- Apply **security policies** (e.g., JWT, OIDC auth) across routes or gateways
- Set **rate limits** or enable **traffic shaping** per client or backend
- Define **mTLS requirements** or configure **TLS passthrough**
- Centralize routing and policy logic using Kubernetes-native resources
- Customize request/response behavior with filters, retries, mirroring, etc.
- Extend your routing logic without writing raw Envoy config

## Configuration in Envoy Gateway

Envoy Gateway uses the **Kubernetes Gateway API** as its foundation and layers its own **CRDs** on top to expose advanced features in a user-friendly way.

You define your desired behavior using Kubernetes manifests. For example:

- Use a `SecurityPolicy` to enforce authentication on an `HTTPRoute`.
- Apply a `BackendTrafficPolicy` to configure circuit breaking or retries.
- Attach a `ClientTrafficPolicy` to rate limit based on client identity.

Envoy Gateway (the controller) watches these custom resources and converts them into the correct configuration for **Envoy Proxy** (the data plane). Envoy Proxy then applies these configurations to real-time traffic.

## Related Resources

- [Envoy Gateway API Reference](../api/extension_types)
- [Kubernetes Gateway API Reference](../api/gateway_api/_index.md)
