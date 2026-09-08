---
title: "FAQ"
---

Answers to recurring questions from GitHub issues and community discussions. If your question
isn't covered here, check the rest of the [Troubleshooting](../) section or open a
[GitHub issue](https://github.com/envoyproxy/gateway/issues/new/choose).

## Why does `kubectl exec`/`port-forward` fail when exposing kube-apiserver through Envoy Gateway?

You've routed the Kubernetes API server through Envoy Gateway, and plain calls work fine
(`kubectl get`, `list`, `watch`), but `kubectl exec`, `attach`, or `port-forward` fail with:

```console
error: unable to upgrade connection: empty server response
```

The Envoy Gateway access log for the failing request shows a protocol error followed by a
rejected upgrade:

```json
{ "method": "GET", "protocol": "HTTP/1.1", "response_code": 502, "response_code_details": "upstream_reset_before_response_started{protocol_error}", "response_flags": "UPE" }
{ "method": "POST", "protocol": "HTTP/1.1", "response_code": 403, "response_code_details": "upgrade_failed" }
```

**Root cause:** `exec`, `attach`, and `port-forward` aren't ordinary request/response calls — they
need a bidirectional streaming channel, which kubectl gets by asking the server to switch
protocols mid-connection (the `SPDY/3.1` or `websocket` upgrade). Protocol upgrade is an
HTTP/1.1-only handshake (a `101 Switching Protocols` response). If the connection between Envoy
Gateway and kube-apiserver negotiates HTTP/2 instead, that upgrade request has nowhere to go and
the request resets. `get`/`list`/`watch` don't ask for an upgrade, so HTTP/2 handles them without
issue.

There are three ways to fix this:

1. **Match the upstream protocol to the client's, on the route (recommended)**

   Attach a `BackendTrafficPolicy` with `useClientProtocol: true` and an explicit `httpUpgrade`
   list. Envoy then negotiates the upstream connection per-request to match what the client asked
   for — ordinary calls still get HTTP/2, and upgrade requests get the HTTP/1.1 handshake they
   need.

   ```yaml
   apiVersion: gateway.envoyproxy.io/v1alpha1
   kind: BackendTrafficPolicy
   metadata:
     name: kube-api
     namespace: default
   spec:
     targetRefs:
       - group: gateway.networking.k8s.io
         kind: HTTPRoute
         name: kube-api
     httpUpgrade:
       - type: "spdy/3.1"
       - type: "websocket"
     useClientProtocol: true
   ```

   `httpUpgrade` must name every subprotocol you need passed through — kubectl has used both
   `spdy/3.1` (older versions) and `websocket` (current versions) across its history, so list both
   unless you control every client's kubectl version.

2. **Force HTTP/1.1 on the Gateway**

   A `ClientTrafficPolicy` with `http1: {}` targeting the `Gateway` avoids the negotiation
   question for that listener entirely, at the cost of also affecting every other route on it.

   ```yaml
   apiVersion: gateway.envoyproxy.io/v1alpha1
   kind: ClientTrafficPolicy
   metadata:
     name: kube-api
     namespace: default
   spec:
     targetRefs:
       - group: gateway.networking.k8s.io
         kind: Gateway
         name: envoy
     http1: {}
   ```

3. **Use a `TLSRoute` instead of an `HTTPRoute`**

   Passing the connection through at L4 avoids terminating TLS (and negotiating HTTP/2) at the
   gateway at all. This only works if kube-apiserver's own certificate is valid for the hostname
   your clients connect to — on a managed control plane, that certificate usually belongs to the
   cloud provider and won't match a custom domain.

{{% alert title="Recommendation" color="primary" %}}
Option 1 (`useClientProtocol: true`) is the best default: it's scoped to a single route and keeps
HTTP/2 for every call that doesn't need an upgrade. Reach for option 2 only if you're fine forcing
HTTP/1.1 gateway-wide, and option 3 only if you control the apiserver's certificate.
{{% /alert %}}

As a one-off, client-side-only workaround that needs no gateway changes, you can also force
kubectl's own HTTP client down to HTTP/1.1 for a single command:

```shell
GODEBUG=http2client=0 kubectl exec -ti <pod> -- sh
```

## Why does `kubectl logs -f` fail or get cut off after a while?

This is a separate issue with the same symptom family: a long-lived streaming response getting
cut off. `HTTPRoute` rules apply a default request timeout, which eventually terminates a
follow-mode log stream. Set an explicit, unbounded timeout on the route rule:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: kube-api
  namespace: default
spec:
  rules:
    - backendRefs: [...]
      timeouts:
        request: 0s
```
