---
title: "Rate Limiting"
---

## Overview

Rate limiting is a technique for controlling the number of incoming requests over a defined period. It can be used to control usage for business purposes, like agreed usage quotas, or to ensure the stability of a system, preventing overload and protecting the system from, e.g., Denial of Service attacks.

## Use Cases

Rate limiting is commonly used to:

- **Prevent Overload:** Protect internal systems like databases from excessive traffic.
- **Enhance Security:** Block or limit abusive behavior such as brute-force attempts or DDoS attacks.
- **Ensure Fair Usage:** Enforce quotas and prevent resource hogging by individual clients.
- **Implement Entitlements:** Define API usage limits based on user identity or role.

## Rate Limiting in Envoy Gateway

Envoy Gateway supports two types of rate limiting:

- **Global Rate Limiting:** Shared limits across all Envoy instances.
- **Local Rate Limiting:** Independent limits per Envoy instance.

Envoy Gateway supports rate limiting through the `BackendTrafficPolicy` custom resource. You can define rate-limiting rules and apply them to `HTTPRoute`, `GRPCRoute`, or `Gateway` resources either by directly referencing them with the targetRefs field or by dynamically selecting them using the targetSelectors field, which matches resources based on Kubernetes labels.

{{% alert title="Note" color="primary" %}}
Rate limits are applied per route, even if the `BackendTrafficPolicy` targets a `Gateway`. For example, if the limit is 100r/s and a Gateway has 3 routes, each route has its own 100r/s bucket.
{{% /alert %}}

---

## Global Rate Limiting

Global rate limiting ensures a consistent request limit across the entire Envoy fleet. This is ideal for shared resources or distributed environments where coordinated enforcement is critical.

Global limits are enforced via Envoy’s external Rate Limit Service, which is automatically deployed and managed by the Envoy Gateway system. The Rate Limit Service requires a datastore component (commonly Redis). When a request is received, Envoy sends a descriptor to this external service to determine if the request should be allowed.

**Benefits of global limits:**

- Centralized control across instances
- Fair sharing of backend capacity
- Burst resistance during autoscaling

### Example

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: global-ratelimit
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: my-api
  rateLimit:
    type: Global
    global:
      rules:
      - limit:
          requests: 100
          unit: Minute

```
This configuration limits all requests across all Envoy instances for the my-api route to 100 requests per minute total. If there are multiple replicas of Envoy, the limit is shared across all of them.

---

## Local Rate Limiting


Local rate limiting applies limits independently within each Envoy Proxy instance. It does not rely on external services, making it lightweight and efficient—especially for blocking abusive traffic early.

**Benefits of local limits:**

- Lightweight and does not require an external rate limit service
- Fast enforcement with rate limiting at the edge
- Effective as a first line of defense against traffic bursts

### Example

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: local-ratelimit
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: my-api
  rateLimit:
    type: Local
    local:
      rules:
      - limit:
          requests: 50
          unit: Minute

```
This configuration limits traffic to 50 requests per minute per Envoy instance for the my-api route. If there are two Envoy replicas, up to 100 total requests per minute may be allowed (50 per replica).

## Related Resources
- [BackendTrafficPolicy](gateway_api_extensions/backend-traffic-policy.md)
- [Task: Global Rate Limit](../../tasks/traffic/global-rate-limit.md)
- [Task: Local Rate Limit](../../tasks/traffic/local-rate-limit.md)
