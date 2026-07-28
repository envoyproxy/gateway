---
title: "CSRF"
---

This task provides instructions for configuring [Cross-Site Request Forgery (CSRF)][csrf] protection on Envoy Gateway.
CSRF is an attack that tricks a user's browser into making unintended requests to a different site where the user is
authenticated.

Envoy Gateway introduces a new field in the [SecurityPolicy][] CRD that allows the user to configure CSRF protection.
This instantiated resource can be linked to a [Gateway][], [HTTPRoute][] or [GRPCRoute][] resource.

## Prerequisites

{{< boilerplate prerequisites >}}

## Configuration

When CSRF protection is enabled, the Envoy CSRF filter validates that the `Origin` header of mutating requests
(POST, PUT, DELETE, PATCH) matches the destination or one of the configured additional origins.
Non-mutating requests (GET, HEAD, OPTIONS) are not affected.

`additionalOrigins` uses the same origin syntax as [`cors.allowOrigins`](./cors): a full origin such as
`https://www.example.com`, a single wildcard label such as `https://*.trusted.com`, an explicit port such as
`http://www.example.com:8080`, or `*` to allow any origin.

Note: Envoy's CSRF filter compares against the host and port of the origin only, so the scheme is ignored.
`https://www.example.com` and `http://www.example.com` are equivalent, and either one allows the request
whichever scheme the client used. The scheme is still required by the syntax so that origins read the same
way here as they do in `cors.allowOrigins`.

The filter supports gradual rollout via `shadowFraction`: the fraction of requests that are evaluated in
dry-run mode instead of being enforced. It is expressed as a `numerator` and an optional `denominator` that
defaults to `100`, and defaults to 0%, i.e. all requests are enforced.

The below example defines a SecurityPolicy that enables CSRF protection and allows `www.example.com` and any
subdomain of `trusted.com` as additional origins.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: csrf-example
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: backend
  csrf:
    additionalOrigins:
    - "https://www.example.com"
    - "https://*.trusted.com"
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: csrf-example
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: backend
  csrf:
    additionalOrigins:
    - "https://www.example.com"
    - "https://*.trusted.com"
```

{{% /tab %}}
{{< /tabpane >}}

With this configuration:

- A `POST` request with `Origin: https://www.example.com` will be **allowed** (Envoy extracts `www.example.com` and matches the exact origin).
- A `POST` request with `Origin: http://www.example.com` will also be **allowed**, since the scheme is not compared.
- A `POST` request with `Origin: https://app.trusted.com` will be **allowed** (Envoy extracts `app.trusted.com` which matches the wildcard).
- A `POST` request with `Origin: https://trusted.com` will be **rejected**: the wildcard matches subdomains, not the apex domain.
- A `POST` request with `Origin: https://www.malicious.com` will be **rejected** with a `403 Forbidden`.
- A `GET` request from any origin will be **allowed** (non-mutating).

### Shadow mode (dry-run)

To evaluate CSRF policies without enforcing them, set `shadowFraction` to 100%:

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: csrf-shadow
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: backend
  csrf:
    shadowFraction:
      numerator: 100
    additionalOrigins:
    - "https://www.example.com"
```

In this mode, all requests are allowed but Envoy tracks CSRF metrics (`request_valid` / `request_invalid`)
so you can monitor the impact before enabling enforcement.

Requests that are not selected for shadowing are enforced, so `shadowFraction` also doubles as the knob for
rolling enforcement out gradually: `numerator: 25` shadows a quarter of the requests and enforces the
remaining three quarters. Lower it towards 0 as the metrics confirm that no legitimate origin is rejected.

[csrf]: https://owasp.org/www-community/attacks/csrf
[SecurityPolicy]: ../../../api/extension_types#securitypolicy
[Gateway]: https://gateway-api.sigs.k8s.io/reference/api-types/gateway/
[HTTPRoute]: https://gateway-api.sigs.k8s.io/reference/api-types/httproute/
[GRPCRoute]: https://gateway-api.sigs.k8s.io/reference/api-types/grpcroute/
