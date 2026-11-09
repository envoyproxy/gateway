---
title: "Authentication Composition"
---

By default, each authentication method configured in a [SecurityPolicy][] applies on its
own. With JWT and Basic authentication on the same route, both are required unless
`jwt.optional` relaxes JWT, so a client that sends either a Bearer token or Basic
credentials is rejected by the method that does not find its credentials.

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
matches requests authenticated by JWT:

- With `defaultAction: Deny`, Basic-authenticated requests need an `Allow` rule that
  matches them by other attributes.
- With `defaultAction: Allow`, a `Deny` rule with a `jwt` principal never matches a
  Basic-authenticated request, so it does not restrict Basic users. Match them by other
  attributes, or use `defaultAction: Deny`.

`jwt.optional` has no effect in `Any` mode, since JWT already lets requests without a
token through.

## Restrictions

`mode: Any` requires both `jwt` and `basicAuth`, either in the policy itself or inherited
from a parent policy through `mergeType`. It does not support:

- `oidc`, `apiKeyAuth` or `extAuth`.
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
