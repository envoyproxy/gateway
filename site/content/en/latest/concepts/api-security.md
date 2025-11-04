---
title: "API Security"
---

## Overview

API security protects applications and services from unauthorized access, data breaches, and abuse. In Envoy Gateway, API security includes mechanisms for authentication, authorization, rate limiting, encryption, and access control at the edge of your infrastructure.

Envoy Gateway extends the Kubernetes Gateway API with features such as `SecurityPolicy`, allowing declarative and Kubernetes-native configuration of security requirements. This model ensures that only authenticated and authorized requests reach backend services, providing a consistent and auditable way to  address modern API threats.

## Key Concepts

| Concept                     | Summary                                                                                  |
|-----------------------------|-----------------------------------------------------------------------------------------|
| Authentication              | Verifies client identity using mechanisms such as API keys, JWTs, Basic Auth, mTLS, and OIDC. |
| Authorization               | Controls access to resources based on user roles, JWT claims, or external services.      |
| Access Control              | IP allowlist/denylist to restrict or permit access from specific addresses.              |
| SecurityPolicy CRD          | Defines authentication, authorization, and encryption settings for Gateway API resources.|
| Rate Limiting               | Limits requests to protect APIs from abuse; supports global and local enforcement.       |
| CORS                        | Configurable policies to allow or restrict cross-origin API requests.                    |

## Use Cases

Use API security to:

- Authenticate and authorize client access to APIs.
- Restrict API access based on IP allowlist/denylist.
- Enforce rate limiting to protect APIs from abuse.
- Apply CORS policies to control cross-origin API requests.
- Encrypt traffic with TLS and manage certificate securely.

## API Security in Envoy Gateway

API security is implemented through Kubernetes-native policies and Envoy Gateway API extensions. The core mechanism is the [SecurityPolicy](./../concepts/gateway_api_extensions/security-policy.md) Custom Resource Definition (CRD), which defines how authentication, authorization, and encryption are applied to incoming traffic. Policies attach to Gateway API resources (such as Gateways, HTTPRoutes, or GRPCRoutes) using `targetRefs` or `targetSelectors`. The most specific policy applies when multiple policies target the same resource.

Envoy Gateway supports API key, JWT, mTLS, Basic Auth, and OIDC authentication; authorization via external services or JWT claims; access control with IP allow/deny lists; CORS configuration; and rate limiting. This model enables secure, declarative access control at the edge, aligning with Kubernetes best practices.

## Examples

- Attach a `SecurityPolicy` with JWT authentication to an `HTTPRoute`.
- Configure CORS on a Gateway listener using a `SecurityPolicy`.
- Enforce global or local rate limiting using `BackendTrafficPolicy`.
- Restrict access to APIs using IP allowlist/denylist in `SecurityPolicy`.

## Related Resources

- [API Key Authentication](../tasks/security/apikey-auth.md)
- [Basic Authentication](../tasks/security/basic-auth.md)
- [CORS](../tasks/security/cors.md)
- [External Authorization](../tasks/security/ext-auth.md)
- [IP Allowlist/Denylist](../tasks/security/restrict-ip-access.md)
- [JWT Authentication](../tasks/security/jwt-authentication.md)
- [JWT Claim Based Authorization](../tasks/security/jwt-claim-authorization.md)
- [OIDC Authorization](../tasks/security/oidc.md)
- [SecurityPolicy API Reference](../api/extension_types#securitypolicy)


