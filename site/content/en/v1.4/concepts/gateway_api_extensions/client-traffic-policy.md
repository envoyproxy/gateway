---
title: "ClientTrafficPolicy"
---
## Before you Begin
- [Gateway API Extensions](_index.md)

## Overview

`ClientTrafficPolicy` is an extension to the Kubernetes Gateway API that allows system administrators to configure how the Envoy Proxy server behaves with downstream clients. It is a policy attachment resource that can be applied to Gateway resources and holds settings for configuring the behavior of the connection between the downstream client and Envoy Proxy listener.

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

You can attach it to Gateway API resources in two ways:

1. Using `targetRefs` to directly reference specific Gateway resources
2. Using `targetSelectors` to match Gateway resources based on labels

The policy applies to all Gateway resources that match either targeting method. When multiple policies target the same resource, the most specific configuration wins.

For example, consider these policies targeting the same Gateway Listener:

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

## Related Resources

- [Connection Limit](../../tasks/traffic/connection-limit.md)
- [HTTP Request Headers](../../tasks/traffic/http-request-headers)
- [HTTP/3](../../tasks/traffic/http3)
- [Mutual TLS: External Clients to the Gateway](../../tasks/security/mutual-tls/)
- [ClientTrafficPolicy API Reference](../../api/extension_types#clienttrafficpolicy)
