---
title: "BackendTrafficPolicy"
---
## Before you Begin
- [Gateway API Extensions](gateway-api-extensions.md)

## Overview
`BackendTrafficPolicy` is an extension to the Kubernetes Gateway API that controls how Envoy Gateway communicates with your backend services. It can configure connection behavior, resilience mechanisms, and performance optimizations without requiring changes to your applications.

Think of it as a traffic controller that sits between your gateway and backend services. It can detect problems, prevent failures from spreading, and optimize request handling to improve system stability.


## Use Cases

`BackendTrafficPolicy` is particularly useful in scenarios where you need to:

- **Protect your services:** Limit connections and reject excess traffic when necessary
- **Build resilient systems:** Detect failing services and redirect traffic
- **Improve performance:** Optimize how requests are distributed and responses are handled
- **Test system behavior:** Inject faults and validate your recovery mechanisms

## `BackendTrafficPolicy` in Envoy Gateway

`BackendTrafficPolicy` is part of the Envoy Gateway API suite, which extends the Kubernetes Gateway API with additional capabilities. It's implemented as a Custom Resource Definition (CRD) that you can use to configure how Envoy Gateway manages traffic to your backend services.

`BackendTrafficPolicy` uses the `targetRefs` field to specify which Gateway API resources the policy should apply to. You can target:
- A Gateway: to apply the policy to all routes in that gateway
- A specific Route (HTTPRoute or GRPCRoute): to apply the policy to that route's traffic

When multiple policies target the same resource, the most specific configuration wins.

Lastly, it's important to note that even when you apply a policy to a Gateway, the policy's effects are tracked separately for each backend service referenced in your routes. For example, if you set up circuit breaking on a Gateway with multiple backend services, each backend service will have its own independent circuit breaker counter. This ensures that issues with one backend service don't affect the others.

## Best Practices
1. **Start Conservative**
   - Begin with higher thresholds
   - Monitor system behavior
   - Adjust based on metrics
   - Test in non-production first

2. **Monitor and Adjust**
   - Track policy effectiveness
   - Monitor backend health
   - Adjust thresholds based on usage
   - Review and update regularly

3. **Plan for Failures**
   - Implement proper fallback mechanisms
   - Test failover scenarios
   - Document recovery procedures
   - Monitor failover events

4. **Security Considerations**
   - Implement appropriate rate limits
   - Monitor for abuse patterns
   - Review security implications
   - Update policies regularly

5. **Performance Optimization**
   - Choose appropriate load balancing strategy
   - Enable response compression when beneficial
   - Monitor and adjust timeouts
   - Consider protocol-specific optimizations

## Related Resources

- [Circuit Breakers](../tasks/traffic/circuit-breaker)
- [Failover](../tasks/traffic/failover)
- [Fault Injection](../tasks/traffic/fault-injection)
- [Global Rate Limit](../tasks/traffic/global-rate-limit)
- [Local Rate Limit](../tasks/traffic/local-rate-limit)
- [Load Balancing](../tasks/traffic/load-balancing)
- [Response Compression](../tasks/traffic/response-compression)
- [Response Override](../tasks/traffic/response-override)
- [gRPC Routing](../tasks/traffic/grpc-routing)
- [BackendTrafficPolicy API Reference](../api/extension_types#backendtrafficpolicy)
- [ClientTrafficPolicy](client-traffic-policy.md)
