---
title: "Traffic Basics"
---

## Overview
Traffic in Envoy Gateway represents the flow of network requests between clients and services. It defines how requests are received, processed, routed, and returned. Envoy Gateway uses the xDS API and Kubernetes Gateway resources (`HTTPRoute`, `Gateway`) to manage and direct traffic dynamically.

## Key Concepts

| Concept | Description |
|----------|--------------|
| Listener | Defines where and how Envoy receives connections (ports, protocols, TLS). |
| Route | Maps incoming requests to destinations based on host, path, or headers. |
| Cluster | Represents a group of upstream endpoints that serve traffic. |

## Use Cases
- Direct HTTP/HTTPS traffic to backend services.  
- Implement canary or blue-green deployments.  
- Balance load across multiple upstreams.  

## Implementation
Envoy Gateway translates Gateway API resources into Envoy configuration using its control plane. Listeners, routes, and clusters are automatically synchronized without manual config.

## Examples
- Route `/api` requests to `backend-svc`.  
- Terminate TLS at the Gateway.  
- Configure weighted routing for traffic shifting.

## Related Resources

- [Gateway API Reference](https://gateway-api.sigs.k8s.io/)
