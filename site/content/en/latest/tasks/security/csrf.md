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
For example, use `www.example.com` instead of `https://www.example.com`.

The filter supports gradual rollout via `filterEnabled` (percentage of requests enforced, defaults to 100)
and `shadowEnabled` (percentage of requests evaluated in dry-run mode without enforcing).

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

To evaluate CSRF policies without enforcing them (useful for gradual rollout), set `filterEnabled` to 0
and `shadowEnabled` to the desired percentage:

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
    filterEnabled: 0
    shadowEnabled: 100
    additionalOrigins:
    - type: Exact
      value: "www.example.com"
```

In this mode, all requests are allowed but Envoy tracks CSRF metrics (`request_valid` / `request_invalid`)
so you can monitor the impact before enabling enforcement.

[csrf]: https://owasp.org/www-community/attacks/csrf
[SecurityPolicy]: ../../../api/extension_types#securitypolicy
[Gateway]: https://gateway-api.sigs.k8s.io/api-types/gateway
[HTTPRoute]: https://gateway-api.sigs.k8s.io/api-types/httproute
[GRPCRoute]: https://gateway-api.sigs.k8s.io/api-types/grpcroute
