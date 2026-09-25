---
title: Client Certificate Authorization
description:
  Authorize requests based on the validated mTLS client certificate using
  SecurityPolicy clientCert principal.
---

## Overview

Envoy Gateway supports authorizing requests based on the TLS client certificate
presented during the mTLS handshake. When a `SecurityPolicy` rule uses
`principal.clientCert`, Envoy evaluates the validated client certificate against
one or more match criteria before allowing or denying the request.

**Supported match types:**

| Field | Matches |
|-------|---------|
| `subject` | Subject Distinguished Name (DN) of the certificate, RFC 4514 string form |
| `subjectAltNames.uris` | URI Subject Alternative Names (e.g. SPIFFE IDs) |
| `subjectAltNames.dnsNames` | DNS Subject Alternative Names |

> **Note:** Email address, IP address, and OtherName SAN types are not currently
> supported. Specifying them will produce a validation error.

## Prerequisites

### mTLS must be configured on the gateway listener

`clientCert` authorization matches against the certificate that the client
presents during the TLS handshake. This certificate is only available when
mutual TLS (mTLS) is enabled on the listener.

Configure mTLS by creating a `ClientTrafficPolicy` that targets the relevant
listener and sets `spec.tls.clientValidation.caCertificateRefs`:

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: ClientTrafficPolicy
metadata:
  name: mtls-policy
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: Gateway
      name: my-gateway
      sectionName: https
  tls:
    clientValidation:
      caCertificateRefs:
        - kind: Secret
          group: ""
          name: my-ca-cert
```

Without mTLS configured, no client certificate is available and
`clientCert` principals will never match.

{{< boilerplate prerequisites >}}

---

## URI SAN Authorization

Use `subjectAltNames.uris` to allow requests from clients whose certificate
carries a specific URI SAN (e.g. a SPIFFE ID).

### Example

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: authz-uri-san
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: my-route
  authorization:
    defaultAction: Deny
    rules:
      - name: allow-spiffe-workload
        action: Allow
        principal:
          clientCert:
            subjectAltNames:
              uris:
                - type: Exact
                  value: "spiffe://my-trust-domain/ns/prod/sa/frontend"
```

Only clients whose certificate contains the URI SAN
`spiffe://my-trust-domain/ns/prod/sa/frontend` (exact match) are allowed.
All other requests are denied by the default action.

`StringMatch` supports `Exact`, `Prefix`, `Suffix`, and `RegularExpression`
match types for URI SANs.

---

## DNS SAN Authorization

Use `subjectAltNames.dnsNames` to allow requests from clients whose certificate
carries a specific DNS SAN.

### Example

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: authz-dns-san
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: my-route
  authorization:
    defaultAction: Deny
    rules:
      - name: allow-client-dns
        action: Allow
        principal:
          clientCert:
            subjectAltNames:
              dnsNames:
                - type: Exact
                  value: "client.example.com"
```

Only clients presenting a certificate with the DNS SAN `client.example.com`
are permitted.

---

## Subject DN Authorization

Use `subject` to match against the certificate's Subject Distinguished Name.
The DN is represented as an RFC 4514 string, for example
`CN=client.example.com,O=Example Inc.,C=US`. Use `RegularExpression` when you
need to match a subset of the DN.

### Example: exact Subject DN

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: authz-subject-dn
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: my-route
  authorization:
    defaultAction: Deny
    rules:
      - name: allow-exact-subject
        action: Allow
        principal:
          clientCert:
            subject:
              type: Exact
              value: "CN=allowed-client,O=Example Inc."
```

### Example: regex Subject DN

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: authz-subject-regex
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: my-route
  authorization:
    defaultAction: Deny
    rules:
      - name: allow-org-clients
        action: Allow
        principal:
          clientCert:
            subject:
              type: RegularExpression
              value: ".*O=Example Inc\\..*"
```

---

## Combining Subject DN and SANs

When both `subject` and `subjectAltNames` are specified in the same
`clientCert` entry, **both** must match for the principal to be satisfied.

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: authz-combined
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: my-route
  authorization:
    defaultAction: Deny
    rules:
      - name: allow-specific-workload
        action: Allow
        principal:
          clientCert:
            subject:
              type: RegularExpression
              value: ".*O=Acme Corp.*"
            subjectAltNames:
              uris:
                - type: Prefix
                  value: "spiffe://acme.example/ns/prod/"
```

---

## Behavior Notes

- **mTLS required:** If mTLS is not configured, no certificate is presented and
  `clientCert` principals will never match any request.
- **AND semantics between `subject` and `subjectAltNames`:** When both `subject`
  and `subjectAltNames` are set, the Subject DN must match **and** the SAN group
  must match.
- **OR semantics within `subjectAltNames`:** All `uris` and `dnsNames` entries
  OR-combine into a single group — a certificate matches if **any one** of the
  listed URI or DNS identities matches. URI and DNS SANs are not AND-combined
  across types, because a workload cert typically carries a URI SAN and a service
  cert a DNS SAN, rarely both.
- **Rule evaluation order:** Rules are evaluated in the order they are defined.
  The first matching rule is applied.
- **HTTPRoute / GRPCRoute only:** `clientCert` principal is not applicable to
  TCPRoute targets.

[SecurityPolicy]: ../../../api/extension_types#securitypolicy
[ClientTrafficPolicy]: ../../../api/extension_types#clienttrafficpolicy
[Gateway]: https://gateway-api.sigs.k8s.io/api-types/gateway
[HTTPRoute]: https://gateway-api.sigs.k8s.io/api-types/httproute
[GRPCRoute]: https://gateway-api.sigs.k8s.io/api-types/grpcroute
