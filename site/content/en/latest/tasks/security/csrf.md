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

Note: Envoy's CSRF filter compares against the host and port of the origin only (the scheme is stripped
before matching). Additional origins must be specified as `host` or `host:port` values, not full URLs.
For example, use `www.example.com` instead of `https://www.example.com`. A SecurityPolicy whose
`additionalOrigins` contain a scheme or a path is rejected at admission, since such a value could never match.

The filter supports gradual rollout via `shadowFraction`: the fraction of requests that are evaluated in
dry-run mode instead of being enforced. It is expressed as a `numerator` and an optional `denominator` that
defaults to `100`, and defaults to 0%, i.e. all requests are enforced.

The below example defines a SecurityPolicy that enables CSRF protection and allows additional origins
matching `www.example.com` exactly and any subdomain of `trusted.com` via regex.

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
    - type: Exact
      value: "www.example.com"
    - type: RegularExpression
      value: ".*\\.trusted\\.com$"
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
    - type: Exact
      value: "www.example.com"
    - type: RegularExpression
      value: ".*\\.trusted\\.com$"
```

{{% /tab %}}
{{< /tabpane >}}

With this configuration:

- A `POST` request with `Origin: https://www.example.com` will be **allowed** (Envoy extracts `www.example.com` and matches the exact origin).
- A `POST` request with `Origin: https://app.trusted.com` will be **allowed** (Envoy extracts `app.trusted.com` which matches the regex).
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
    - type: Exact
      value: "www.example.com"
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
