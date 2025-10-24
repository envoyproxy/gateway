---
title: "Authorization"
---


## Overview
Authorization controls what an authenticated identity is allowed to do. Envoy Gateway supports both local RBAC and external authorization integrations.

## Key Concepts

| Concept | Description |
|----------|--------------|
| RBAC | Role-based access control rules for routes or clusters. |
| External AuthZ | Delegates authorization to an external service. |
| Policy | Defines access rules for requests (allow/deny). |

## Use Cases
- Restrict access to `/admin` routes.  
- Apply route-level access control.  
- Integrate with an external OPA or custom AuthZ service.  

## Implementation
`AuthorizationPolicy` resources are implemented using Envoy RBAC and `ext_authz` filters, evaluated per request.

## Examples
- Deny all except specific IPs.  
- Use an external OPA server.  
- Apply per-route RBAC policies.

## Related Resources
- [Authorization Guide](../howto/authorization.md)  
- [External Authorization Reference](../reference/ext-authz.md)
