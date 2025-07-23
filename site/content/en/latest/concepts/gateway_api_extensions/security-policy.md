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

`SecurityPolicy` is implemented as a Kubernetes Custom Resource Definition (CRD) and follows the policy attachment model. You can attach it to Gateway API resources in two ways:

1. Using `targetRefs` to directly reference specific Gateway resources
2. Using `targetSelectors` to match Gateway resources based on labels

The policy applies to all resources that match either targeting method. When multiple policies target the same resource, the most specific configuration wins.

For example, consider these policies targeting the same Gateway Listener:

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
