# AgentCert ext_authz Example

This example verifies a calling agent's **certificate and live trust score** on
every request through Envoy Gateway, using a [`SecurityPolicy`][secpol] with
`extAuth` — no core change and no custom Envoy filter.

An "AgentCert" is an X.509 certificate issued to an autonomous agent; the
verifier here is the open-source [`agentcert-tag`][tag] sidecar (MIT), which
exposes an Envoy `ext_authz` HTTP endpoint. On each request it:

1. reads the caller's cert from the `x-agent-cert*` request headers,
2. checks signature chain, revocation, and the current trust score, and
3. returns allow/deny plus `x-agentcert-*` verdict headers.

Any HTTP `ext_authz` service that follows the same header contract works the
same way — the pattern, not the sidecar, is the point. This mirrors the existing
[External Authorization task][ext-auth].

## Shadow mode by default

The policy sets `failOpen: true` and the sidecar runs with `TAG_MODE=shadow`, so
it **decides and logs but never blocks**. You can drop it into real traffic,
watch `x-agentcert-verdict` / `x-agentcert-shadow-would` in your access logs to
see what enforcement *would* do, and flip to enforcement only when the deny rate
looks right (set `TAG_MODE=enforce` on the sidecar and `failOpen: false` on the
policy).

## Header contract

Request headers forwarded to the verifier (`headersToExtAuth`):

| Header | Meaning |
|---|---|
| `x-agent-cert` | base64-encoded PEM leaf (the AgentCert) |
| `x-agent-cert-proof` | stapled proof-of-possession (for header carriage) |
| `x-agent-cert-carriage` | `header` (default) or `mtls` |
| `x-agent-id` | claimed agent id, cross-checked to the cert SAN |

Verdict headers injected onto the upstream request (`headersToBackend`):

| Header | Example |
|---|---|
| `x-agentcert-verdict` | `VERIFIED` / `REVOKED` / `UNVERIFIED` / `EXPIRED` / `ERROR` |
| `x-agentcert-agent-id` | `ans://acme/agent/prod` |
| `x-agentcert-trustscore` | `812` |
| `x-agentcert-tier` | `Trusted` / `Caution` / `Untrusted` / `Unknown` |
| `x-agentcert-decision` | `allow` / `deny` |
| `x-agentcert-shadow-would` | `allow` / `deny` |

The sidecar matches exactly `/ext_authz`, so the policy uses `pathOverride`
(replace the path) rather than `path` (append to it).

## Apply it

Assumes the [Quickstart][quickstart] Gateway named `eg` and a `backend` Service
on port 3000 (also from the Quickstart).

```shell
kubectl apply -f agentcert-ext-authz.yaml
kubectl get securitypolicy/agentcert-ext-authz -o yaml
```

Send a request through the gateway; in shadow mode it is always allowed, and the
backend (and your access logs) receive the `x-agentcert-*` verdict headers:

```shell
curl -v -H "Host: www.example.com" \
  -H "x-agent-cert: <base64-PEM>" \
  -H "x-agent-cert-proof: <pop>" \
  "http://${GATEWAY_HOST}/agent"
```

## Files

- [agentcert-ext-authz.yaml](agentcert-ext-authz.yaml) — sidecar Deployment +
  Service and the `SecurityPolicy` with `extAuth`.

[secpol]: https://gateway.envoyproxy.io/docs/api/extension_types/#securitypolicy
[ext-auth]: https://gateway.envoyproxy.io/docs/tasks/security/ext-auth/
[quickstart]: https://gateway.envoyproxy.io/docs/tasks/quickstart/
[tag]: https://github.com/pdxlab/agentcert-tag
