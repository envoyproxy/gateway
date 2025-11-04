---
title: "Sessions"
---

## Overview
Session management maintains continuity across multiple client requests. While Envoy is stateless by default, Envoy Gateway supports session persistence for sticky connections.

## Key Concepts

| Concept | Description |
|----------|--------------|
| [Session Persistence](./../tasks/traffic/session-persistence.md) | Allows client requests to be consistently routed to the same backend service instance. |
| Strong Session Affinity | Routes a client’s requests to the same backend. Cookie-based persistence with a permanent cookie. |
| Weak Session Affinity | Routes to the same backend is not guaranteed. Header-based persistence or cookie persistence with a short TTL. |
| [Load Balancer State](./../tasks/traffic/load-balancer-state.md) | Maintains session mappings across connections. |

## Use Cases
- Maintain login state in web applications.  
- Preserve shopping cart data.  
- Improve cache locality in microservices.  

## Implementation
Envoy Gateway supports session persistence via cookie-based and consistent-hash load balancing policies defined in `HTTPRoute` configurations.

## Examples
- Enable cookie-based session affinity.  
- Use consistent hashing by header.  
- Route same user to same backend pod.

## Related Resources

- [Load Balancing Reference](load-balancing.md)
