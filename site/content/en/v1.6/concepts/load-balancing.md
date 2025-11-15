---
title: "Load Balancing"
---

## Overview

Load balancing distributes incoming requests across multiple backend services to improve availability, responsiveness, and scalability. Instead of directing all traffic to a single backend, which can cause slowdowns or outages, load balancing spreads the load across multiple instances, helping your applications stay fast and reliable under pressure.

## Use Cases

Use load balancing to:

- Handle high traffic by distributing it across multiple service instances  
- Keep services available even if one or more backends go down  
- Improve response time by routing to less busy or closer backends  

## Load Balancing in Envoy Gateway

Envoy Gateway supports several load balancing strategies that determine how traffic is distributed across backend services. These strategies are configured using the `BackendTrafficPolicy` resource and can be applied to `Gateway`, `HTTPRoute`, or `GRPCRoute` resources either by directly referencing them using the targetRefs field or by dynamically selecting them using the targetSelectors field, which matches resources based on Kubernetes labels.

**Supported load balancing types:**
- **Round Robin** – Sends requests sequentially to all available backends
- **Random** – Chooses a backend at random to balance load
- **Least Request** – Sends the request to the backend with the fewest active requests (this is the default)
- **Consistent Hash** – Routes requests based on a hash (e.g., client IP or header), which helps keep repeat requests going to the same backend (useful for session affinity)

If no load balancing strategy is specified, Envoy Gateway uses **Least Request** by default.

## Example: Round Robin Load Balancing

This example shows how to apply the Round Robin strategy using a `BackendTrafficPolicy` that targets a specific `HTTPRoute`:

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: round-robin-policy
  namespace: default
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: round-robin-route
  loadBalancer:
    type: RoundRobin
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: round-robin-route
  namespace: default
spec:
  parentRefs:
  - name: eg
  hostnames:
  - "www.example.com"
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /round
    backendRefs:
    - name: backend
      port: 3000
```
In this setup, traffic matching /round is distributed evenly across all available backend service instances. For example, if there are four replicas of the backend service, each one should receive roughly 25% of the requests.

## Related Resources
- [BackendTrafficPolicy](gateway_api_extensions/backend-traffic-policy.md)
- [Task: Load Balancing](../tasks/traffic/load-balancing.md)
