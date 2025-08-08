---
title: "BackendTrafficPolicy"
---
## Before you Begin
- [Gateway API Extensions](_index.md)

## Overview
`BackendTrafficPolicy` is an extension to the Kubernetes Gateway API that controls how Envoy Gateway communicates with your backend services. It can configure connection behavior, resilience mechanisms, and performance optimizations without requiring changes to your applications.

Think of it as a traffic controller between your gateway and backend services. It can detect problems, prevent failures from spreading, and optimize request handling to improve system stability.

## Use Cases

`BackendTrafficPolicy` is particularly useful in scenarios where you need to:

1. **Protect your services:**
   Limit connections and reject excess traffic when necessary

2. **Build resilient systems:**
   Detect failing services and redirect traffic

3. **Improve performance:**
   Optimize how requests are distributed and responses are handled

4. **Test system behavior:**
   Inject faults and validate your recovery mechanisms

## BackendTrafficPolicy in Envoy Gateway

`BackendTrafficPolicy` is part of the Envoy Gateway API suite, which extends the Kubernetes Gateway API with additional capabilities. It's implemented as a Custom Resource Definition (CRD) that you can use to configure how Envoy Gateway manages traffic to your backend services.

### Targets

BackendTrafficPolicy can be attached to Gateway API resources using two targeting mechanisms:

1. **Direct Reference (`targetRefs`)**: Explicitly reference specific resources by name and kind.
2. **Label Selection (`targetSelectors`)**: Match resources based on their labels (see [targetSelectors API reference](../../api/extension_types#targetselectors))

```yaml
# Direct reference targeting
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: direct-policy
spec:
  targetRefs:
    - kind: HTTPRoute
      name: my-route
  circuitBreaker:
    maxConnections: 50

---
# Label-based targeting
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: selector-policy
spec:
  targetSelectors:
    - kind: HTTPRoute
      matchLabels:
        app: payment-service
  rateLimit:
    type: Local
    local:
      requests: 10
      unit: Second
```

The policy applies to all resources that match either targeting method. You can target various Gateway API resource types including
`Gateway`, `HTTPRoute`, `GRPCRoute`, `TCPRoute`, `UDPRoute`, `TLSRoute`.

**Important**: A BackendTrafficPolicy can only target resources in the same namespace as the policy itself.

### Precedence

When multiple BackendTrafficPolicies apply to the same resource, Envoy Gateway resolves conflicts using a precedence hierarchy based on the target resource type, regardless of how the policy was attached:

1. **Route-level policies** (HTTPRoute, GRPCRoute, etc.) - Highest precedence
2. **Gateway-level policies** - Lower precedence

```yaml
# Gateway-level policy (lower precedence) - Applies to all routes in the gateway
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

---
# Route-level policy (higher precedence)
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

In this example, the HTTPRoute `my-route` would use `maxConnections: 50` from the route-level policy, overriding the gateway-level setting of 100.

#### Multiple Policies at the Same Level

When multiple BackendTrafficPolicies target the same resource at the same hierarchy level (e.g., multiple policies targeting the same HTTPRoute), Envoy Gateway uses the following tie-breaking rules:

1. **Creation Time Priority**: The oldest policy (earliest `creationTimestamp`) takes precedence
2. **Name-based Sorting**: If policies have identical creation timestamps, they are sorted alphabetically by namespaced name, with the first policy taking precedence

```yaml
# Policy created first - takes precedence
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: alpha-policy
  creationTimestamp: "2023-01-01T10:00:00Z"
spec:
  targetRefs:
    - kind: HTTPRoute
      name: my-route
  circuitBreaker:
    maxConnections: 30

---
# Policy created later - lower precedence
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: beta-policy
  creationTimestamp: "2023-01-01T11:00:00Z"
spec:
  targetRefs:
    - kind: HTTPRoute
      name: my-route
  circuitBreaker:
    maxConnections: 40
```

In this example, `alpha-policy` would take precedence due to its earlier creation time, so the HTTPRoute would use `maxConnections: 30`.

When the `mergeType` field is unset, no merging occurs and only the most specific configuration takes effect. However, policies can be configured to merge with parent policies using the `mergeType` field (see [Policy Merging](#policy-merging) section below).

## Policy Merging

BackendTrafficPolicy supports merging configurations using the `mergeType` field, which allows route-level policies to combine with gateway-level policies rather than completely overriding them. This enables layered policy strategies where platform teams can set baseline configurations at the Gateway level, while application teams can add specific policies for their routes.

### Merge Types

- **StrategicMerge**: Uses Kubernetes strategic merge patch semantics, providing intelligent merging for complex data structures including arrays
- **JSONMerge**: Uses RFC 7396 JSON Merge Patch semantics, with simple replacement strategy where arrays are completely replaced

### Example Usage

Here's an example demonstrating policy merging for rate limiting:

```yaml
# Platform team: Gateway-level policy with global abuse prevention
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: global-backendtrafficpolicy
spec:
  rateLimit:
    type: Global
    global:
      rules:
      - clientSelectors:
        - sourceCIDR:
            type: Distinct
            value: 0.0.0.0/0
        limit:
          requests: 100
          unit: Second
        shared: true
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: Gateway
    name: eg

---
# Application team: Route-level policy with specific limits
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: route-backendtrafficpolicy
spec:
  mergeType: StrategicMerge  # Enables merging with gateway policy
  rateLimit:
    type: Global
    global:
      rules:
      - clientSelectors:
        - sourceCIDR:
            type: Distinct
            value: 0.0.0.0/0
        limit:
          requests: 5
          unit: Minute
        shared: false
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: signup-service-httproute
```

In this example, the route-level policy merges with the gateway-level policy, resulting in both rate limits being enforced: the global 100 requests/second abuse limit and the route-specific 5 requests/minute limit.

### Key Constraints

- The `mergeType` field can only be set on policies targeting child resources (like HTTPRoute), not parent resources (like Gateway)
- When `mergeType` is unset, no merging occurs - only the most specific policy takes effect
- The merged configuration combines both policies, enabling layered protection strategies

## Related Resources

- [Circuit Breakers](../../tasks/traffic/circuit-breaker.md)
- [Failover](../../tasks/traffic/failover)
- [Fault Injection](../../tasks/traffic/fault-injection)
- [Global Rate Limit](../../tasks/traffic/global-rate-limit)
- [Local Rate Limit](../../tasks/traffic/local-rate-limit)
- [Load Balancing](../../tasks/traffic/load-balancing)
- [Response Compression](../../tasks/traffic/response-compression)
- [Response Override](../../tasks/traffic/response-override)
- [BackendTrafficPolicy API Reference](../../api/extension_types#backendtrafficpolicy)
