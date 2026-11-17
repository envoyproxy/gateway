---
title: "Authentication Composition"
---

By default, each authentication method configured in a [SecurityPolicy][] applies on its
own. With JWT and Basic authentication on the same route, both are required unless
`jwt.optional` relaxes JWT, so a client that sends either a Bearer token or Basic
credentials is rejected by the method that does not find its credentials.

Listing JWT and Basic authentication in `authentication.anyOf` accepts a request
authenticated by either of them. Any other configured method, such as `oidc`, `extAuth`
or `apiKeyAuth`, stays required.

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
    anyOf: [JWT, BasicAuth]
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

Each listed method lets requests without its own credentials through, and a check placed
after both methods rejects a request that neither authenticated. That check responds with 403
and does not send a `WWW-Authenticate` challenge, because Envoy's RBAC filter has no
configurable deny status.

If a filter placed before the check changes the request's route, for example a Lua
script placed by `EnvoyProxy.spec.filterOrder` that rewrites the path, the check applies
the rules of the new route, and credentials verified for another policy's route do not
count. A filter placed after the check that changes the route bypasses it, as it does
every other authentication method.

Authorization rules still apply after authentication. A rule with a `jwt` principal only
matches requests authenticated by JWT:

- With `defaultAction: Deny`, Basic-authenticated requests need an `Allow` rule that
  matches them by other attributes.
- With `defaultAction: Allow`, a `Deny` rule with a `jwt` principal never matches a
  Basic-authenticated request, so it does not restrict Basic users. Match them by other
  attributes, or use `defaultAction: Deny`.

`jwt.optional` has no effect when JWT is listed in `anyOf`, since JWT already lets
requests without a token through.

## Combining with required methods

Methods not listed in `anyOf` apply as usual. For example, adding `extAuth` to the policy
above requires the external authorization service to allow the request in addition to
JWT or Basic authentication. By default the external authorization service runs before
JWT and Basic authentication, so it also sees requests that they later reject.

## Restrictions

`anyOf` currently accepts exactly `JWT` and `BasicAuth`, and both must be configured,
either in the policy itself or inherited from a parent policy through `mergeType`. It
does not support:

- `oidc.forwardAccessToken`. The OIDC filter replaces the client's `Authorization`
  header with the access token, so Basic credentials never reach Basic authentication.
- `oidc.passThroughAuthHeader`, which skips OIDC for requests carrying a JWT header and
  relies on JWT authentication rejecting requests without a valid token.
- `oidc.forwardIDToken` on the `Authorization` header or on a header a JWT provider
  reads. The OIDC filter overwrites that header with the ID token, replacing the Basic
  credentials or JWT the client sent.
- `extAuth.recomputeRoute`.
- `jwt.failOpen`, which would accept requests with invalid tokens.
- `recomputeRoute` on JWT providers.
- A JWT header extractor on the `Authorization` header with a `valuePrefix` other than
  `Bearer `. JWT authentication would read Basic credentials as an invalid token and
  reject the request.
- `basicAuth.forwardUsernameHeader`. Envoy's Basic authentication filter leaves that
  header untouched when it lets a request through without Basic credentials, so a client
  authenticated by JWT could set it to any username
  ([envoyproxy/envoy#48957](https://github.com/envoyproxy/envoy/issues/48957)).

A policy that violates these restrictions through an inherited configuration is marked
invalid, and the route returns 500.

Without `anyOf`, every configured method is required, as before.

[SecurityPolicy]: ../../../api/extension_types#securitypolicy
[JWT Authentication]: ../jwt-authentication
[Basic Authentication]: ../basic-auth
