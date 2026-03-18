---
title: "SecurityPolicy"
---

## Before you Begin
- [Gateway API Extensions](_index.md)

## Overview

`SecurityPolicy` is an Envoy Gateway extension to the Kubernetes Gateway API that allows you to define authentication and authorization requirements for traffic entering your gateway. It acts as a security layer that only properly authenticated and authorized requests are allowed through your backend services.

`SecurityPolicy` is designed for you to enforce access controls through configuration at the edge of your infrastructure in a declarative, Kubernetes-native way, without needing to configure complex proxy rules manually.

## Use Cases

1. **Authentication Methods:**
    - Authenticate client apps using mTLS, JWTs, API keys, or Basic Auth
    - Authenticate users with OIDC Provider integration

2. **Authorization Controls:**
    - Define and enforce authorization rules based on user roles and permissions
    - Integrate with external authorization services for real-time policy decisions
    - JWT Token Authorization Checks

3. **Cross-Origin Security:**
    - Configure CORS to allow or restrict cross-origin requests for APIs

## SecurityPolicy in Envoy Gateway

`SecurityPolicy` is implemented as a Kubernetes Custom Resource Definition (CRD) and follows the policy attachment model.

### Targets

SecurityPolicy can be attached to Gateway API resources using two targeting mechanisms:

1. **Direct Reference (`targetRefs`)**: Explicitly reference specific resources by name and kind.
2. **Label Selection (`targetSelectors`)**: Match resources based on their labels (see [targetSelectors API reference](../../api/extension_types#targetselectors))

The policy applies to all resources that match either targeting method. You can target various Gateway API resource types including `Gateway`, `HTTPRoute`, `GRPCRoute`, and `TCPRoute`.

Note: TCPRoute support is limited to authorization using client IP allow/deny lists (IP-based authorization). Other SecurityPolicy features such as JWT, API Key, Basic Auth, or OIDC are not applicable to TCPRoute targets.

**Important**: A SecurityPolicy can only target resources in the same namespace as the policy itself.

### Precedence

When multiple SecurityPolicies apply to the same resource, Envoy Gateway resolves conflicts using a precedence hierarchy based on the target resource type and section-level specificity:

1. **Route rule-level policies** (HTTPRoute, GRPCRoute, or TCPRoute with `sectionName` targeting specific rules) - Highest precedence
2. **Route-level policies** (HTTPRoute, GRPCRoute, or TCPRoute without `sectionName`) - High precedence  
3. **Listener-level policies** (Gateway with `sectionName` targeting specific listeners) - Medium precedence
4. **Gateway-level policies** (Gateway without `sectionName`) - Lowest precedence

#### Multiple Policies at the Same Level

When multiple SecurityPolicies target the same resource at the same hierarchy level (e.g., multiple policies targeting the same HTTPRoute), Envoy Gateway uses the following tie-breaking rules:

1. **Creation Time Priority**: The oldest policy (earliest `creationTimestamp`) takes precedence
2. **Name-based Sorting**: If policies have identical creation timestamps, they are sorted alphabetically by namespaced name, with the first policy taking precedence

```yaml
# Policy created first - takes precedence
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: alpha-policy
  creationTimestamp: "2023-01-01T10:00:00Z"
spec:
  targetRefs:
    - kind: HTTPRoute
      name: my-route
  cors:
    allowOrigins:
      - exact: https://example.com

---
# Policy created later - lower precedence
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: beta-policy
  creationTimestamp: "2023-01-01T11:00:00Z"
spec:
  targetRefs:
    - kind: HTTPRoute
      name: my-route
  cors:
    allowOrigins:
      - exact: https://different.com
```

In this example, `alpha-policy` would take precedence due to its earlier creation time, so the HTTPRoute would use the CORS setting from `alpha-policy`.

```yaml
# HTTPRoute with named rules
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: my-route
spec:
  rules:
  - name: rule-1  # Named rule for sectionName targeting
    matches:
    - path:
        value: "/api"
    backendRefs:
    - name: api-service
      port: 80

---
# Route rule-level policy (highest precedence)
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: rule-policy
spec:
  targetRef:
    kind: HTTPRoute
    name: my-route
    sectionName: rule-1  # Targets specific named rule
  cors:
    allowOrigins:
    - exact: https://rule.example.com

---
# Route-level policy (high precedence)
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: route-policy
spec:
  targetRef:
    kind: HTTPRoute
    name: my-route  # No sectionName = entire route
  cors:
    allowOrigins:
    - exact: https://route.example.com

---
# Listener-level policy (medium precedence)
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: listener-policy
spec:
  targetRef:
    kind: Gateway
    name: my-gateway
    sectionName: https-listener  # Targets specific listener
  cors:
    allowOrigins:
    - exact: https://listener.example.com

---
# Gateway-level policy (lowest precedence)
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: gateway-policy
spec:
  targetRef:
    kind: Gateway
    name: my-gateway  # No sectionName = entire gateway
  cors:
    allowOrigins:
    - exact: https://gateway.example.com
```

In this example, the specific rule `rule-1` within HTTPRoute `my-route` would use the CORS settings from the route rule-level policy (`https://rule.example.com`), overriding the route-level, listener-level, and gateway-level settings.

For section-specific targeting, consider these policies with different hierarchy levels targeting the same Gateway:

```yaml
# Policy A: Applies to a specific listener
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: listener-policy
  namespace: default
spec:
  targetRefs:
    - kind: Gateway
      name: my-gateway
      sectionName: https  # Applies only to "https" listener
  cors:
    allowOrigins:
      - exact: https://example.com
---
# Policy B: Applies to the entire gateway
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: gateway-policy
  namespace: default
spec:
  targetRefs:
    - kind: Gateway
      name: my-gateway  # Applies to all listeners
  cors:
    allowOrigins:
      - exact: https://default.com
```

In the example, policy A affects only the HTTPS listener, while policy B applies to the rest of the listeners in the gateway. Since Policy A is more specific, the system will show Overridden=True for Policy B on the https listener.

When the `mergeType` field is unset, no merging occurs and only the most specific configuration takes effect. However, policies can be configured to merge with parent policies using the `mergeType` field (see [Policy Merging](#policy-merging) section below).

## Policy Merging

SecurityPolicy supports merging configurations using the `mergeType` field, which allows route-level or route rule-level policies to combine with gateway-level or listener-level policies rather than completely overriding them. This enables layered security strategies where platform teams can set baseline security configurations at the Gateway level, while application teams can add specific security policies for their routes.

When merging occurs, route-level policies will merge with either a gateway-level or listener-level policy, but not both. If both gateway and listener policies exist, the listener-level policy takes precedence.

### Merge Types

- **StrategicMerge**: Uses Kubernetes strategic merge patch semantics, providing intelligent merging for complex data structures including arrays
- **JSONMerge**: Uses RFC 7396 JSON Merge Patch semantics, with simple replacement strategy where arrays are completely replaced

### Example Usage

Here's an example demonstrating policy merging for combining authentication and CORS policies:

```yaml
# Platform team: Gateway-level policy with baseline authentication
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: gateway-security
  namespace: default
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: Gateway
    name: my-gateway
    sectionName: https-listener
  basicAuth:
    users:
      name: basic-auth-secret

---
# Application team: Route-level policy with CORS configuration
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: route-security
  namespace: default
spec:
  mergeType: StrategicMerge  # Enables merging with gateway policy
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: my-route
  cors:
    allowOrigins:
    - exact: https://example.com
    allowMethods:
    - GET
    - POST
    allowHeaders:
    - x-header-1
```

In this example, the route-level policy merges with the gateway-level policy, resulting in both security controls being enforced: the baseline BasicAuth (from Gateway) and the route-specific CORS policy (from Route). This allows platform teams to enforce organization-wide authentication requirements while enabling application teams to configure route-specific cross-origin policies.

### Key Constraints

- The `mergeType` field can only be set on policies targeting child resources (like HTTPRoute), not parent resources (like Gateway)
- When `mergeType` is unset, no merging occurs - only the most specific policy takes effect
- The merged configuration combines both policies, enabling layered security strategies
- When the same security feature is configured in both parent and child policies (e.g., both define CORS), the child policy's configuration takes precedence for that specific feature

## Related Resources
- [API Key Authentication](../../tasks/security/apikey-auth.md)
- [Basic Authentication](../../tasks/security/basic-auth.md)
- [CORS](../../tasks/security/cors.md)
- [External Authorization](../../tasks/security/ext-auth.md)
- [IP Allowlist/Denylist](../../tasks/security/restrict-ip-access.md)
- [JWT Authentication](../../tasks/security/jwt-authentication.md)
- [JWT Claim Based Authorization](../../tasks/security/jwt-claim-authorization.md)
- [OIDC Authorization](../../tasks/security/oidc.md)
- [SecurityPolicy API Reference](../../api/extension_types#securitypolicy)
