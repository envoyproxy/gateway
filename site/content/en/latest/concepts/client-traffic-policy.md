---
title: "Client Traffic Policy"
---


## Overview

Client Traffic Policy is a way to control how Envoy Gateway handles incoming traffic from clients. Think of it as a set of rules that helps you manage the connection between your clients and the gateway. This policy is particularly useful when you need to ensure your gateway remains secure, performant, and resilient under different conditions.

## Use Cases

Client Traffic Policy is particularly useful in scenarios where you need to:

1. **Protect Gateway Resources**
   - Limit the number of concurrent connections
   - Control TCP keepalive settings
   - Manage connection timeouts
   - Prevent resource exhaustion

2. **Enhance Security**
   - Implement IP allow/deny lists
   - Configure mutual TLS for client connections
   - Set up JWT-based authorization
   - Enable proxy protocol support

3. **Optimize Performance**
   - Configure HTTP/3 support
   - Manage request headers
   - Control connection behavior
   - Optimize protocol settings

4. **Monitor and Control**
   - Track connection metrics
   - Implement rate limiting
   - Monitor client behavior
   - Control access patterns

## Client Traffic Policy in Envoy Gateway

Client Traffic Policy is part of the Envoy Gateway API suite, which extends the Kubernetes Gateway API with additional capabilities. It's implemented as a Custom Resource Definition (CRD) that you can use to configure how Envoy Gateway manages incoming client traffic.

You can apply a Client Traffic Policy to a Gateway resource, and it will affect how that gateway handles client connections. The policy's effects are applied at the listener level, with some important distinctions:

- For HTTP listeners: All HTTP listeners in a Gateway share a common connection counter
- For HTTPS/TLS listeners: Each listener maintains its own separate connection counter

This separation ensures that different types of traffic can be managed independently while still maintaining overall control over gateway resources.

## Best Practices

When implementing Client Traffic Policy, consider the following best practices:

1. **Start Conservative**
   - Begin with higher connection limits
   - Monitor system behavior
   - Adjust based on metrics
   - Test in non-production first

2. **Monitor and Adjust**
   - Track connection patterns
   - Monitor resource usage
   - Adjust limits based on usage
   - Review and update regularly

3. **Security First**
   - Implement appropriate connection limits
   - Use mutual TLS where possible
   - Configure IP allow/deny lists
   - Monitor for suspicious activity

4. **Performance Considerations**
   - Configure appropriate keepalive settings
   - Enable HTTP/3 for modern clients
   - Optimize protocol settings
   - Monitor connection metrics

## Related Resources

- [Connection Limit](../tasks/traffic/connection-limit)
- [HTTP Request Headers](../tasks/traffic/http-request-headers)
- [HTTP/3](../tasks/traffic/http3)
- [IP Allowlist/Denylist](../tasks/traffic/ip-allowlist-denylist)
- [JWT Claim-Based Authorization](../tasks/traffic/jwt-claim-based-authorization)
- [Mutual TLS: External Clients to the Gateway](../tasks/traffic/mutual-tls-external-clients)
- [Secure Gateways](../tasks/traffic/secure-gateways)
- [Client Traffic Policy API Reference](../../api/extension_types#clienttrafficpolicy)