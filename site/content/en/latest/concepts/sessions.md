---
title: "Sessions"
---

## Overview
Session management maintains continuity across multiple client requests. While Envoy is stateless by default, Envoy Gateway supports session persistence for sticky connections.

## Key Concepts
| Concept | Description |
|----------|--------------|
| Session Affinity | Routes a client’s requests to the same backend. |
| Cookies | Identify and track session information. |
| Load Balancer State | Maintains session mappings across connections. |

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
