---
title: "API Gateway"
---

## Before You Begin
You may want to be familiar with:
- [What is Envoy Proxy?](../envoy_proxy)

## Overview

An API Gateway acts as the front door to your system — it receives incoming traffic, applies rules, and forwards requests to the right backend services.

In modern cloud-native systems, especially those built on microservices, an API Gateway helps simplify architecture, improve security, and streamline traffic management. Instead of each service dealing with these responsibilities individually, the API Gateway offloads and centralizes them, allowing your services to focus on core business logic.

## Use Cases

Use an API Gateway when:
- You have microservices and want to avoid duplicating logic across them.
- You want a central point of control for access, monitoring, and traffic rules.
- You’re exposing internal services to the public internet.
- You need protocol support for HTTP, gRPC, or TLS.
- You want to enforce policy and see traffic metrics at the edge.

## How API Gateways fit into Envoy Gateway

Envoy Gateway makes it easier to use Envoy Proxy as an API Gateway by combining the Kubernetes Gateway API with its own set of custom resources. This approach allows users to manage routing, security, and traffic control using familiar Kubernetes tools—without needing to learn Envoy's low-level configuration model.

The Kubernetes Gateway API offers a standard, Kubernetes-native way to define how external traffic is handled within your cluster. Resources like Gateway and HTTPRoute let you describe how requests should be routed. Envoy Gateway builds on top of this API by adding its own enhancements, making it easier to access Envoy’s powerful capabilities in a more user-friendly and declarative way.

Behind the scenes, Envoy Gateway watches the resources you create and translates them into configuration for Envoy Proxy. Envoy Proxy then handles the actual traffic based on that config—routing requests, enforcing policies, and collecting metrics.


## Related Resources

- [Getting Started with Envoy Gateway](../../tasks/quickstart)
- [Kubernetes Gateway API](https://gateway-api.sigs.k8s.io/)
