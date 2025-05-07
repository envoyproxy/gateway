---
title: "SecurityPolicy"
---

## Before you Begin
- [Gateway API Extensions](gateway-api-extensions.md)

## Overview

`SecurityPolicy` is an Envoy Gateway extension to the Kubernetes Gateway API that allows you to define authentication and authorization requirements for traffic entering your gateway. It acts as a security layer that ensures only properly authenticated and authorized requests are allowed through to your backend services.

`SecurityPolicy` is designed for users who want to enforce access controls at the edge of their infrastructure in a declarative, Kubernetes-native way—without needing to manually configure complex proxy rules.

## Use Cases

- **JWT Authentication:** Require valid JSON Web Tokens to access specific routes or services.
- **OIDC Integration:** Use OpenID Connect providers to authenticate clients at the gateway.
- **CORS Configuration:** Allow or restrict cross-origin requests for APIs.
- **API Key or Basic Auth:** Enforce simple forms of authentication for legacy systems.
- **External Authorization:** Integrate with external authorization services to make policy decisions in real time.

## `SecurityPolicy` in Envoy Gateway
`SecurityPolicy` is implemented as a Kubernetes Custom Resource Definition (CRD) and follows the policy attachment model—meaning you attach it to Gateway API resources like Gateway, HTTPRoute, or GRPCRoute.

This model allows you to apply fine-grained security controls to specific traffic flows. You can:

Attach a policy to a Gateway to apply authentication/authorization to all traffic entering the cluster.

Attach a policy to an individual Route to control access to specific services or paths.

Envoy Gateway processes the attached policy and translates it into the appropriate Envoy configuration, enabling enforcement of the defined rules.

## Best Practices
1. Apply the Principle of Least Privilege
    - Only expose the necessary routes or resources
    - Restrict access based on roles or claims

2. Validate Your Identity Sources
    - Use trusted JWT issuers
    - Keep keys and token lifetimes in sync
    - Monitor token expiration and failures

3. Test in Lower Environments
    - Test authentication flows in staging
    - Validate that protected routes behave as expected

## Related Resources
- [API Key Authentication](../tasks/security/apikey-auth.md)
- [Basic Authentication](../tasks/security/basic-auth.md)
- [CORS](../tasks/security/cors.md)
- [External Authorization](../tasks/security/ext-auth.md)
- [IP Allowlist/Denylist](../tasks/security/restrict-ip-access.md)
- [JWT Authentication](../tasks/security/jwt-authentication.md)
- [JWT Claim Based Authorization](../tasks/security/jwt-claim-authorization.md)
- [OIDC Authorization](../tasks/security/oidc.md)
- [Threat Model](../tasks/security/threat-model.md)
- [SecurityPolicy API Reference](../api/extension_types#securitypolicy)
