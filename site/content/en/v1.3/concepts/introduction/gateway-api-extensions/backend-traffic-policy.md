---
title: "BackendTrafficPolicy"
---
## Before you Begin
- [Gateway API Extensions](gateway-api-extensions.md)

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

You can attach it to Gateway API resources in two ways:

1. Using `targetRefs` to directly reference specific Gateway resources
2. Using `targetSelectors` to match Gateway resources based on labels

The policy applies to all resources that match either targeting method. When multiple policies target the same resource, the most specific configuration wins.

For example, consider these two policies:

```yaml
# Policy 1: Applies to all routes in the gateway
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
# Policy 2: Applies to a specific route
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

In this example `my-route` and `my-gateway` would both affect the route. However, since Policy 2 targets the route directly while Policy 1 targets the gateway, Policy 2's configuration (`maxConnections: 50`) will take precedence for that specific route.

Lastly, it's important to note that even when you apply a policy to a Gateway, the policy's effects are tracked separately for each backend service referenced in your routes. For example, if you set up circuit breaking on a Gateway with multiple backend services, each backend service will have its own independent circuit breaker counter. This ensures that issues with one backend service don't affect the others.

## Related Resources

- [Circuit Breakers](../tasks/traffic/circuit-breaker)
- [Failover](../tasks/traffic/failover)
- [Fault Injection](../tasks/traffic/fault-injection)
- [Global Rate Limit](../tasks/traffic/global-rate-limit)
- [Local Rate Limit](../tasks/traffic/local-rate-limit)
- [Load Balancing](../tasks/traffic/load-balancing)
- [Response Compression](../tasks/traffic/response-compression)
- [Response Override](../tasks/traffic/response-override)
- [BackendTrafficPolicy API Reference](../api/extension_types#backendtrafficpolicy)
