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
- **Backend Utilization** – Uses client-observed load reports (e.g., Open Resource Cost Aggregation (ORCA) metrics) to dynamically weight endpoints; if no metrics are available, it behaves similar to even-weight round-robin.

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

## Backend Utilization (ORCA)

Backend Utilization load balancing uses Open Resource Cost Application (ORCA) metrics to make load balancing decisions. These metrics are reported by the backend service in response headers or trailers.

### ORCA Load Metrics

The backend service (or its sidecar) reports load metrics in response headers or trailers (for streaming requests). ORCA supports multiple formats for these metrics:

- **JSON**: Use the `endpoint-load-metrics` header with a JSON object.
  ```http
  endpoint-load-metrics: JSON {"cpu_utilization": 0.3, "mem_utilization": 0.8}
  ```
- **TEXT**: Use the `endpoint-load-metrics` header with comma-separated key-value pairs.
  ```http
  endpoint-load-metrics: TEXT cpu=0.3,mem=0.8,foo_bytes=123
  ```
- **Binary Proto**: Use the `endpoint-load-metrics-bin` header with a base64-encoded serialized `OrcaLoadReport` proto.
  ```http
  endpoint-load-metrics-bin: Cg4KCHNvbWUta2V5Eg0AAAAAAADwPw==
  ```

For more details, see:
- [ORCA Load Report Proto](https://www.envoyproxy.io/docs/envoy/latest/xds/data/orca/v3/orca_load_report.proto)
- [ORCA Design Document](https://docs.google.com/document/d/1NSnK3346BkBo1JUU3I9I5NYYnaJZQPt8_Z_XCBCI3uA)

### Automatic Header Removal

By default, Envoy forwards the ORCA response headers/trailers from the upstream cluster to the downstream client. This means that if the downstream client is also configured to use client-side weighted round-robin, it will load balance against Envoy based on upstream weights.

To prevent this, Envoy Gateway automatically removes these headers by default when `BackendUtilization` is enabled. You can change this behavior using the `keepResponseHeaders` field in `backendUtilization`.

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: backend-utilization-policy
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: backend-utilization-route
  loadBalancer:
    type: BackendUtilization
    backendUtilization:
      keepResponseHeaders: true # Keep headers and forward them to the client
```

## Related Resources
- [BackendTrafficPolicy](gateway_api_extensions/backend-traffic-policy.md)
- [Task: Load Balancing](../tasks/traffic/load-balancing.md)
