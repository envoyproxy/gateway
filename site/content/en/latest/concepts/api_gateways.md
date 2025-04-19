---
title: "API Gateways"
---

## Before You Begin
You may want to be familiar with:
- [What is Envoy Proxy?](envoy_proxy)

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

## Configuration in Envoy Gateway

Under the hood, Envoy Proxy is a powerful, production-grade proxy that supports many of the capabilities you'd expect from an API Gateway — like traffic routing, retries, TLS termination, observability, and more. But configuring Envoy directly can be complex and verbose.

Envoy Gateway solves this by providing a Kubernetes-native interface, using the standard Gateway API and its own custom resources, to make configuring Envoy Proxy simpler and more approachable. You define high-level traffic rules using resources like Gateway, HTTPRoute, or TLSRoute, and Envoy Gateway automatically translates them into detailed Envoy Proxy configurations.


## Related Resources

- [Getting Started with Envoy Gateway](https://gateway.envoyproxy.io/docs/latest/getting-started/)
- [Kubernetes Gateway API](https://gateway-api.sigs.k8s.io/)
