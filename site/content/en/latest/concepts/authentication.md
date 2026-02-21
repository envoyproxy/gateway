---
title: "Authentication"
---

## Overview

Authentication in Envoy Gateway ensures that only verified users or systems can access backend services. By validating credentials such as tokens, certificates, or keys before forwarding requests, Envoy Gateway provides a secure boundary for workloads running in Kubernetes.

Authentication is configured using the `SecurityPolicy` Custom Resource Definition (CRD), which allows you to define authentication requirements for traffic entering your gateway and attach policies to Gateway API resources.


## Key Concepts

| Concept                | Description                                                                                       |
|------------------------|---------------------------------------------------------------------------------------------------|
| Authentication Methods | Envoy Gateway supports mTLS, JWT, API keys, Basic Auth, and OIDC authentication methods.          |
| SecurityPolicy CRD     | The main CRD for configuring authentication and authorization, attached via `targetRefs` or `targetSelectors`.|
| Policy Attachment      | Authentication policies are attached to Gateway API resources (e.g., Gateway, HTTPRoute, GRPCRoute).|


## Use Cases

Use Authentication to:
- Authenticate client applications using mTLS, JWT, API keys, or Basic Auth.
- Integrate with OIDC providers for user authentication.
- Enforce access controls at the edge for APIs and services.


## Implementation

Envoy Gateway expresses authentication using the `SecurityPolicy` CRD, an extension to the Kubernetes Gateway API. The `SecurityPolicy` defines authentication requirements and is attached to Gateway API resources through `targetRefs` or `targetSelectors`. Supported authentication methods include mTLS, JWT, API keys, Basic Auth, and OIDC.


## Examples

- Attach a `SecurityPolicy` to an `HTTPRoute` to require JWT authentication from the `Authorization` header.
- Store Basic Auth credentials in a Kubernetes secret and reference them in a `SecurityPolicy`.
- Use a local JWKS via Kubernetes ConfigMap for JWT validation in a `SecurityPolicy`.
- Configure OIDC authentication at the Gateway or HTTPRoute level with a `SecurityPolicy`.


## Related Resources

- [JWT Authentication Task (latest)](https://gateway.envoyproxy.io/latest/tasks/security/jwt-authentication/)
- [JWT Authentication User Guide (v0.6)](https://gateway.envoyproxy.io/v0.6/user/jwt-authentication/)
- [Gateway API HTTPRoute Documentation](https://gateway-api.sigs.k8s.io/api-types/httproute/)
- [Envoy JWT Authentication Filter Reference](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/jwt_authn_filter#config-http-filters-jwt-authn)
- [JSON Web Token (JWT) RFC 7519](https://tools.ietf.org/html/rfc7519)
- [JSON Web Key Set (JWKS) RFC 7517](https://tools.ietf.org/html/rfc7517)
