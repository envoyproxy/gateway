---
title: "API Gateways"
weight: 3
---

## Overview
An API gateway is a centralized entry point for managing, securing, and routing requests to backend services. It handles cross-cutting concerns, like authentication, rate limiting, and protocol translation, so individual services don’t have to. Decoupling clients from internal systems simplifies scaling, enforces consistency, and reduces redundancy.

## Use Cases

Use an API Gateway to:
- Avoid duplicating logic across microservices.
- Create a central point of control for access, monitoring, and traffic rules.
- Expose internal services to the public internet.
- Provide protocol support for HTTP, gRPC, or TLS.
- Enforce policies and see traffic metrics at the edge.

## API Gateways in relation to Envoy Gateway

Under the hood, Envoy Proxy is a powerful, production-grade proxy that supports many of the capabilities you'd expect from an API Gateway, like traffic routing, retries, TLS termination, observability, and more. However, configuring Envoy directly can be complex and verbose.

Envoy Gateway makes configuring Envoy Proxy simple by implementing and extending the Kubernetes-native Gateway API. You define high-level traffic rules using resources like Gateway, HTTPRoute, or TLSRoute, and Envoy Gateway automatically translates them into detailed Envoy Proxy configurations.

## Related Resources

- [Getting Started with Envoy Gateway](../../tasks/quickstart.md)
- [Kubernetes Gateway API](https://gateway-api.sigs.k8s.io/)