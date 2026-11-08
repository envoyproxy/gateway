---
title: Authentication Composition
---

`SecurityPolicy` can accept either JWT or HTTP Basic credentials on the same
route. Authorization remains independent: successful authentication does not
bypass configured authorization rules or external authorization.

## Accept JWT or Basic credentials

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: alternative-authentication
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
      issuer: https://issuer.example.com
      audiences: [example-api]
      remoteJWKS:
        uri: https://issuer.example.com/keys
  basicAuth:
    users:
      name: example-users
  authorization:
    defaultAction: Deny
    rules:
    - name: basic-readers
      action: Allow
      principal:
        basicAuth:
          users: [example-reader]
    - name: token-readers
      action: Allow
      principal:
        jwt:
          provider: example
          scopes: [read]
```

`Any` requires both JWT and BasicAuth. OIDC and API key authentication are not
supported in this mode. Omitted mode, or `All`, preserves sequential enforcement
of the configured authentication filters. This does not make credential schemes
sharing one Authorization header simultaneously usable.

Missing credentials are tolerated by each authentication filter individually.
An independent RBAC gate requires successful authentication by at least one
configured method. This gate remains enforced even without authorization rules
or when the authorization default action is Allow.

Invalid presented credentials reject the request, including when a valid
credential for another method is supplied using a separate JWT extractor.
`jwt.failOpen` is incompatible with verified authentication composition.
JWT extraction otherwise follows the existing provider configuration.

The native filter response codes apply: no successful authentication is rejected
with 403 by the RBAC gate; invalid Basic credentials and most invalid JWTs return
401; an unaccepted JWT audience returns 403. The gate does not implement a 401
challenge negotiation flow.

## Authorize verified Basic usernames

`authorization.rules[].principal.basicAuth.users` matches exact, case-sensitive
usernames emitted after successful Basic authentication. It requires a BasicAuth
configuration in the effective policy. It also works with BasicAuth alone,
without `authentication.mode: Any`.

Multiple principal types in a rule must all match. Use separate Allow rules for
alternative identities. Existing operations, first-match rule ordering, CEL
expressions and default actions retain their behavior.

The controller enables username metadata automatically when it is consumed.
Identity matches use the applicable BasicAuth filter's metadata namespace;
client headers cannot supply that identity. JWT identity remains associated with
the configured provider and verified payload. A JWT does not need a `sub` claim
to satisfy the authentication gate.

## Send verified identity to external authorization

A gRPC external authorization service can opt into receiving verified identity:

```yaml
spec:
  extAuth:
    includeAuthenticationMetadata: true
    grpc:
      backendRefs:
      - name: authorization
        port: 9000
```

The service receives Envoy's standard CheckRequest metadata context, containing
JWT payloads under `envoy.filters.http.jwt_authn`, keyed by provider name, and
Basic usernames under the generated BasicAuth filter namespace. This discloses
JWT payload claims to the service. Route metadata is controlled independently by
`includeRouteMetadata`. HTTP external authorization does not support this option.

External authorization runs before local authentication by default. Configure
`EnvoyProxy.spec.filterOrder` explicitly:

```yaml
spec:
  filterOrder:
  - name: envoy.filters.http.ext_authz
    after: envoy.filters.http.jwt_authn
```

For a BasicAuth-only policy, use `after: envoy.filters.http.basic_auth` instead.
The controller places the authentication gate immediately after the local
authentication filters and validates the effective filter chain.

With this option, the required sequence is local authentication, authentication
gate, external authorization, then local authorization. Reordering JWT,
BasicAuth or RBAC is unsupported for policies consuming verified authentication
metadata. Invalid combinations produce a policy translation error and a closed
route, rather than silently omitting authentication.

The service must fail closed. Route recomputation by JWT or external
authorization is unsupported in verified composition. External responses cannot
modify Authorization, Host or pseudoheaders, and their dynamic metadata is not
ingested. Other response headers remain available for application-specific
policy results. Trusted extensions that mutate authentication metadata or clear
the route cache require a separate security review; this feature does not make
arbitrary extension chains safe.

## Standards and boundaries

These fields extend Envoy Gateway's SecurityPolicy API. They are not portable
Kubernetes Gateway API fields or tenant-management primitives.

The mechanisms are HTTP Basic ([RFC 7617](https://www.rfc-editor.org/rfc/rfc7617)),
OAuth Bearer ([RFC 6750](https://www.rfc-editor.org/rfc/rfc6750)), Envoy's JWT and
BasicAuth filters, RBAC, and the existing external authorization gRPC protocol.
Use TLS, explicit accepted audiences, trusted issuers and appropriate token
lifetimes. OAuth access tokens belong at API boundaries; OIDC ID tokens are not a
substitute for API access tokens. RFC 9068 describes an interoperable JWT access
token profile, not a universal tenant authorization schema.

Application owners configure their own authorization rules using these
primitives. No tenant header or group claim grants access without an explicit
policy binding.

## Verification

Run the translator tests and optional live Envoy protocol suite:

```shell
go test ./internal/gatewayapi ./internal/xds/translator
ENVOY_BINARY=/path/to/envoy go test ./internal/xds/translator -run TestAuthenticationProtocol -count=1
```

The live suite uses temporary test credentials, a local HTTP backend and a gRPC
authorization service. It checks authentication alternatives, verified username
matching, JWT authorization, identity delivery, service failures, forbidden
header mutations and route isolation.
