---
title: "Circuit Breaking"
---

## Overview

Circuit breaking is a resilience mechanism used to prevent system overload and quickly fail requests when upstream services are degraded or unresponsive. By defining thresholds for concurrent connections, concurrent requests, and pending requests, circuit breakers help maintain system stability and apply back-pressure to protect backend services.

Envoy Gateway leverages circuit breaking to limit resource consumption and ensure reliable traffic handling. Through configurable thresholds and integration with Kubernetes via the BackendTrafficPolicy custom resource, users can fine-tune circuit breaking behavior to fit their specific resilience and performance requirements.

## Key Concepts

| Concept                      | Description                                                                                            |
| ---------------------------- | ------------------------------------------------------------------------------------------------------ |
| **Circuit Breaker**          | A safety control that limits the number of concurrent or pending requests sent to an upstream service. |
| **BackendTrafficPolicy CRD** | Configures circuit breaking thresholds for Gateway API resources.                                      |
| **Concurrent Connections**   | Maximum number of active connections Envoy can maintain with an upstream service.                      |
| **Concurrent Requests**      | Maximum number of in-flight requests permitted to a backend.                                           |
| **Pending Requests**         | Maximum size of the request queue; requests beyond this limit return `503`.                            |



## Use Cases

Use circuit breaking to:

- Protect backend services from overload by limiting concurrent connections, requests, and pending requests.
- Fail fast and apply back-pressure when upstream services are degraded or unresponsive.
- Improve system resilience by isolating failing backends and preventing cascading failures.

## Circuit Breaking in Envoy Gateway

Envoy Gateway implements circuit breaking through the `BackendTrafficPolicy` Custom Resource, which defines per-backend thresholds for traffic flow control.

- Policies can target `Gateway`, `HTTPRoute`, or `GRPCRoute` resources using `targetRefs` or `targetSelectors`.

- Each backend reference in a route has its own independent circuit breaker counters, preventing one failing service from impacting others.

- Default thresholds (1024 connections/requests) are conservative and should be tuned for high-throughput systems.

- Circuit breaker counters are distributed, meaning they are not synchronized across Envoy instances.

## Examples

- Limit concurrent requests and pending requests for a route:

  ```yaml
  apiVersion: gateway.envoyproxy.io/v1alpha1
  kind: BackendTrafficPolicy
  metadata:
    name: circuitbreaker-for-route
  spec:
    targetRefs:
      - group: gateway.networking.k8s.io
        kind: HTTPRoute
        name: backend
    circuitBreaker:
      maxPendingRequests: 0
      maxParallelRequests: 10
  ```

- Apply circuit breaking to all routes in a Gateway, with a maximum of 100 connections:

  ```yaml
  apiVersion: gateway.envoyproxy.io/v1alpha1
  kind: BackendTrafficPolicy
  metadata:
    name: gateway-policy
  spec:
    targetRefs:
      - kind: Gateway
        name: my-gateway
    circuitBreaker:
      maxConnections: 100
  ```

- Set a lower maximum connection count for a specific HTTPRoute:

  ```yaml
  apiVersion: gateway.envoyproxy.io/v1alpha1
  kind: BackendTrafficPolicy
  metadata:
    name: route-policy
  spec:
    targetRefs:
      - kind: HTTPRoute
        name: my-route
    circuitBreaker:
      maxConnections: 50
  ```

## Related Resources

- [Circuit Breakers](../tasks/traffic/circuit-breaker.md)
- [BackendTrafficPolicy API Reference](../api/extension_types#backendtrafficpolicy)
- [Gateway](https://gateway-api.sigs.k8s.io/api-types/gateway/)
- [HTTPRoute](https://gateway-api.sigs.k8s.io/api-types/httproute/)
- [GRPCRoute](https://gateway-api.sigs.k8s.io/api-types/grpcroute/)
- [Envoy Circuit Breakers](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/upstream/circuit_breaking)
- [Failover](../tasks/traffic/failover)
- [Fault Injection](../tasks/traffic/fault-injection)
- [Global Rate Limit](../tasks/traffic/global-rate-limit)
- [Local Rate Limit](../tasks/traffic/local-rate-limit)
- [Load Balancing](../tasks/traffic/load-balancing)
- [Response Compression](../tasks/traffic/response-compression)
- [Response Override](../tasks/traffic/response-override)
