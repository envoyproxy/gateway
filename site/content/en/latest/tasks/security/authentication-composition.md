---
title: "Authentication Composition"
---

By default, every authentication method configured in a [SecurityPolicy][] must succeed.
With JWT and Basic authentication on the same route, that means no request can pass: a
client sends either a Bearer token or Basic credentials, and the method that does not
find its credentials rejects the request.

Setting `authentication.mode` to `Any` accepts a request authenticated by either JWT or
Basic authentication.

## Prerequisites

Follow the [JWT Authentication][] and [Basic Authentication][] tasks to set up a JWT
provider and an `.htpasswd` Secret.

## Configuration

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: jwt-or-basic
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: backend
  authentication:
    mode: Any
  jwt:
    providers:
    - name: example
      remoteJWKS:
        uri: https://raw.githubusercontent.com/envoyproxy/gateway/main/examples/kubernetes/jwt/jwks.json
  basicAuth:
    users:
      name: basic-auth
```

| Request | Result |
|---|---|
| Valid JWT, no Basic credentials | Accepted |
| Valid Basic credentials, no JWT | Accepted |
| No credentials for either method | 403 Forbidden |
| Invalid Basic credentials | 401 Unauthorized, even if a valid JWT is also presented |
| Invalid JWT | 401 Unauthorized (403 for an audience mismatch), even if valid Basic credentials are also presented |

Each method lets requests without its own credentials through, and a check placed after
both methods rejects a request that neither authenticated. That check responds with 403
and does not send a `WWW-Authenticate` challenge, because Envoy's RBAC filter has no
configurable deny status.

Authorization rules still apply after authentication. A rule with a `jwt` principal only
matches requests authenticated by JWT, so with `defaultAction: Deny`, Basic-authenticated
requests need a rule that matches them by other attributes.

## Restrictions

`mode: Any` requires both `jwt` and `basicAuth`, either in the policy itself or inherited
from a parent policy through `mergeType`. It does not support:

- `oidc` or `apiKeyAuth`.
- `jwt.failOpen`, which would accept requests with invalid tokens.
- `recomputeRoute` on JWT providers.
- `basicAuth.forwardUsernameHeader`. Envoy's Basic authentication filter leaves that
  header untouched when it lets a request through without Basic credentials, so a client
  authenticated by JWT could set it to any username.

A policy that violates these restrictions through an inherited configuration is marked
invalid, and the route returns 500.

`mode: All`, the default, keeps the existing behavior.

[SecurityPolicy]: ../../../api/extension_types#securitypolicy
[JWT Authentication]: ../jwt-authentication
[Basic Authentication]: ../basic-auth
