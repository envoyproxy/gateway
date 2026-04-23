---
title: "CSRF Protection"
---

This task explains how to configure Cross-Site Request Forgery (CSRF) protection
using the [SecurityPolicy][] API.

## Prerequisites

{{< boilerplate prerequisites >}}

## Enabling CSRF Protection

The CSRF filter validates the `Origin` header on mutating requests (POST, PUT, DELETE, PATCH)
against the request destination and any additional allowed origins. Requests without an `Origin`
header or with a non-matching origin are rejected with a `403 Forbidden` response.
GET and HEAD requests are always allowed.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: csrf-protection
  namespace: default
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: my-route
  csrf:
    additionalOrigins:
      - type: Suffix
        value: example.com
      - type: Exact
        value: https://trusted.partner.com
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
  name: csrf-protection
  namespace: default
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: my-route
  csrf:
    additionalOrigins:
      - type: Suffix
        value: example.com
      - type: Exact
        value: https://trusted.partner.com
```

{{% /tab %}}
{{< /tabpane >}}

Verify the policy is accepted:

```shell
kubectl get securitypolicies.gateway.envoyproxy.io -n default
```

```
NAME               STATUS     AGE
csrf-protection    Accepted   5s
```

## Origin Matching

The `additionalOrigins` field supports the same match types as other Envoy Gateway string matchers:

| Type | Description | Example |
|------|-------------|---------|
| `Exact` | Exact string match | `https://example.com` |
| `Prefix` | Prefix match | `https://app.` |
| `Suffix` | Suffix match | `example.com` |
| `RegularExpression` | RE2 regex match | `https://.*\.example\.com` |

The request destination is always allowed implicitly — you only need to specify *additional* origins.

[SecurityPolicy]: ../../../api/extension_types#securitypolicy
