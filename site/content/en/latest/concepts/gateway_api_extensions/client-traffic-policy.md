---
title: "ClientTrafficPolicy"
---
## Before you Begin

- [Gateway API Extensions](_index.md)

## Overview

`ClientTrafficPolicy` is an extension to the Kubernetes Gateway API that allows system administrators to configure how the Envoy Proxy server behaves with downstream clients. It is a policy attachment resource that can be applied to `Gateway` and `ListenerSet` resources and holds settings for configuring the behavior of the connection between the downstream client and Envoy Proxy listener.

Think of `ClientTrafficPolicy` as a set of rules for your Gateway's entry points, it lets you configure specific behaviors for each listener in your Gateway, with more specific rules taking precedence over general ones.

## Use Cases

`ClientTrafficPolicy` is particularly useful in scenarios where you need to:

1. **Enforce TLS Security**
   Configure TLS termination, mutual TLS (mTLS), and certificate validation at the edge.

2. **Manage Client Connections**
   Control TCP keepalive behavior and connection timeouts for optimal resource usage.

3. **Handle Client Identity**
   Configure trusted proxy chains to correctly resolve client IPs for logging and access control.

4. **Normalize Request Paths**
   Sanitize incoming request paths to ensure compatibility with backend routing rules.

5. **Tune HTTP Protocols**
   Configure HTTP/1, HTTP/2, and HTTP/3 settings for compatibility and performance.

6. **Monitor Listener Health**
   Set up health checks for integration with load balancers and failover mechanisms.

## ClientTrafficPolicy in Envoy Gateway

`ClientTrafficPolicy` is part of the Envoy Gateway API suite, which extends the Kubernetes Gateway API with additional capabilities. It's implemented as a Custom Resource Definition (CRD) that you can use to configure how Envoy Gateway manages incoming client traffic.

### Targets

ClientTrafficPolicy can be attached to Gateway API resources using two targeting mechanisms:

1. **Direct Reference (`targetRefs`)**: Explicitly reference specific resources by name and kind. Supported kinds are:
   - `Gateway` — apply the policy to all listeners on the Gateway (or a specific listener via `sectionName`)
   - `ListenerSet` — apply the policy to all listeners in the ListenerSet (or a specific listener via `sectionName`)
2. **Label Selection (`targetSelectors`)**: Match resources based on their labels (see [targetSelectors API reference](../../api/extension_types#targetselectors)). Both `Gateway` and `ListenerSet` kinds are supported.

The policy applies to all resources that match either targeting method.

Targeting a `ListenerSet` is useful when you want to apply client traffic settings to a group of listeners managed independently from the parent Gateway, without affecting Gateway-owned listeners.

**Important**: A ClientTrafficPolicy can only target resources in the same namespace as the policy itself.

### Precedence

When multiple ClientTrafficPolicies apply to the same listener, Envoy Gateway resolves conflicts using target specificity and creation-time priority:

1. **Section-specific policies** (targeting a specific listener via `sectionName` on a `Gateway` or `ListenerSet`) - Highest precedence
2. **ListenerSet-wide policies** (targeting an entire `ListenerSet`) - Medium precedence
3. **Gateway-wide policies** (targeting an entire `Gateway`) - Lowest precedence

A ListenerSet-wide policy takes precedence over a Gateway-wide policy for any listener that belongs to that ListenerSet. The Gateway-wide policy is still applied to listeners that are owned directly by the Gateway and not covered by a ListenerSet-wide policy.

#### Multiple Policies at the Same Level

When multiple ClientTrafficPolicies target the same resource at the same specificity level (e.g., multiple policies targeting the same Gateway listener section), Envoy Gateway uses the following tie-breaking rules:

1. **Creation Time Priority**: The oldest policy (earliest `creationTimestamp`) takes precedence
2. **Name-based Sorting**: If policies have identical creation timestamps, they are sorted alphabetically by namespaced name, with the first policy taking precedence

```yaml
# Policy created first - takes precedence
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: ClientTrafficPolicy
metadata:
  name: alpha-policy
  creationTimestamp: "2023-01-01T10:00:00Z"
spec:
  targetRefs:
    - kind: Gateway
      name: my-gateway
      sectionName: https-listener
  timeout:
    http:
      idleTimeout: 30s

---
# Policy created later - lower precedence
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: ClientTrafficPolicy
metadata:
  name: beta-policy
  creationTimestamp: "2023-01-01T11:00:00Z"
spec:
  targetRefs:
    - kind: Gateway
      name: my-gateway
      sectionName: https-listener
  timeout:
    http:
      idleTimeout: 40s
```

In this example, `alpha-policy` would take precedence due to its earlier creation time, so the listener would use `idleTimeout: 30s`.

For example, consider these policies with different specificity levels targeting the same Gateway:

```yaml
# Policy A: Targets a specific listener in the gateway
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: ClientTrafficPolicy
metadata:
  name: listener-specific-policy
spec:
  targetRefs:
    - kind: Gateway
      name: my-gateway
      sectionName: https-listener  # Targets specific listener
  timeout:
    http:
      idleTimeout: 30s

---
# Policy B: Targets the entire gateway
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: ClientTrafficPolicy
metadata:
  name: gateway-wide-policy
spec:
  targetRefs:
    - kind: Gateway
      name: my-gateway  # Targets all listeners
  timeout:
    http:
      idleTimeout: 60s
```

In this case:

- Policy A will be applied/attached to the specific Listener defined in the `targetRef.SectionName`
- Policy B will be applied to the remaining Listeners within the Gateway. Policy B will have an additional status condition Overridden=True.

The same specificity rules apply when a `ListenerSet` is involved. Consider this example where a Gateway and a ListenerSet each have a policy, plus a section-specific policy on the ListenerSet:

```yaml
# Policy A: Targets a specific listener in a ListenerSet (highest precedence for that listener)
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: ClientTrafficPolicy
metadata:
  name: listenerset-section-policy
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: ListenerSet
      name: my-listener-set
      sectionName: ext-https  # Targets a single listener
  tls:
    minVersion: "1.3"

---
# Policy B: Targets the entire ListenerSet (beats Gateway-wide for ListenerSet listeners)
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: ClientTrafficPolicy
metadata:
  name: listenerset-wide-policy
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: ListenerSet
      name: my-listener-set
  timeout:
    http:
      idleTimeout: 45s

---
# Policy C: Targets the entire Gateway (lowest precedence)
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: ClientTrafficPolicy
metadata:
  name: gateway-wide-policy
spec:
  targetRefs:
    - kind: Gateway
      name: my-gateway
  timeout:
    http:
      idleTimeout: 60s
```

In this case:

- Policy A is applied to the `ext-https` listener in the ListenerSet.
- Policy B is applied to the remaining listeners in `my-listener-set`. Policy B will have `Overridden=True` for the `ext-https` listener covered by Policy A.
- Policy C is applied to listeners owned directly by `my-gateway` that are not part of `my-listener-set`. Policy C will have `Overridden=True` for every listener covered by Policy A or Policy B.

When a ClientTrafficPolicy targets a `ListenerSet`, the policy's status `ancestorRef` is set to the `ListenerSet` itself rather than the parent Gateway, so you can observe attachment status scoped to the ListenerSet directly.

## Related Resources

- [Connection Limit](../../tasks/traffic/connection-limit.md)
- [HTTP Request Headers](../../tasks/traffic/http-request-headers)
- [HTTP Response Headers](../../tasks/traffic/http-response-headers)
- [HTTP/3](../../tasks/traffic/http3)
- [Mutual TLS: External Clients to the Gateway](../../tasks/security/mutual-tls/)
- [ClientTrafficPolicy API Reference](../../api/extension_types#clienttrafficpolicy)
