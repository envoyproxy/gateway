---
title: "Ext-Proc Access Log Enrichment"
---

## Overview

Envoy's [ext-proc filter](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/ext_proc_filter) writes per-stream statistics (latency, gRPC status, custom attributes) into [filter state](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/advanced/filter_state) keyed by the filter's instance name. Envoy Gateway generates those instance names from policy identity — e.g. `envoy.filters.http.ext_proc/namespace/extproc/my-policy/0` — following a convention that is opaque to users.

This means users who want to include ext-proc data in access logs have to hand-write `%FILTER_STATE(envoy.filters.http.ext_proc/...:latency_ns)%` with the full internal name, which is verbose and breaks whenever the policy is renamed or reordered.

This document describes `%EG_EXT_FILTER_STATE(name:attribute)%`, a synthetic access log operator that EG resolves at xDS translation time so users only need to reference the friendly name they assigned in their `EnvoyExtensionPolicy`.

## Goals

- Let users reference ext-proc filter state in access log format strings by a name they control.
- Resolve at translation time — Envoy only ever sees standard `%FILTER_STATE(...)%` operators in xDS.
- No coupling between `EnvoyProxy` (format strings) and `EnvoyExtensionPolicy` (extension config) beyond the user-chosen name.

## Non Goals

- Runtime resolution (all substitution happens at xDS generation time).
- Letting users control internal filter names — they assign aliases only.
- Automatic expansion into multiple matching instances for a single operator.
- Support for WASM, Dynamic Modules, and Lua in this iteration (they don't write filter state under their instance name by default).

## Design Decisions

### 1. Users assign aliases — they don't control internal names

EG generates internal filter names deterministically from policy identity. Letting users control them would risk naming conflicts and break the uniqueness guarantee within an HCM.

Instead, users set an optional `name` on each `extProc` entry. That name is:

- The lookup key for `%EG_EXT_FILTER_STATE(name:attribute)%` resolution.
- Not visible to Envoy — stripped at translation time.

Name uniqueness is **not enforced at the API level** — duplicates can arise across multiple `EnvoyExtensionPolicy` resources targeting routes on the same listener. This is intentional: a CEL uniqueness rule on a single EEP would not catch the cross-policy case (which is the real conflict surface), and the cost budget on the CRD validation makes even single-EEP checks impractical. Conflicts are instead resolved at HCM-build time by first-match (see [Conflict resolution for shared names](#4-conflict-resolution-for-shared-names)).

Because EG owns the internal names, any user-facing reference to them must go through an operator EG intercepts at translation time. That operator is necessarily non-standard: both `%EG_EXT_FILTER_STATE(...)%` and any alternative form are unrecognized by Envoy and will cause a xDS NACK if they reach it unresolved.

### 2. Protection against unresolved operators

An unresolved `%EG_EXT_FILTER_STATE(name:attr)%` reaching Envoy as-is may cause an xDS NACK. Unresolved operators are replaced by EG with `"-"`. This prevents NACKs and leaves an observable signal in logs that the name wasn't matched.

### 3. `%EG_EXT_FILTER_STATE(name:attr)%` vs. a name-only operator `%EG_FILTER_NAME(name)%`

A narrower alternative would be a `%EG_FILTER_NAME(name)%` operator that resolves only the internal filter name, leaving users to wrap it:

```
%FILTER_STATE(%EG_FILTER_NAME(auth-proc)%:latency_ns)%
```

The `%EG_EXT_FILTER_STATE(name:attr)%` provides **cleaner NACK protection**: an unresolved `%EG_EXT_FILTER_STATE(...)%` can be replaced with `"-"`, leaving a valid format string. An unresolved `%EG_FILTER_NAME(...)%` within a different operator may produce a malformed format string. Safe fallback would require parsing and replacing the entire enclosing expression. See [Protection Against Unresolved Operators](#2-protection-against-unresolved-operators).

### 4. Conflict resolution for shared names

`%EG_EXT_FILTER_STATE(name:attr)%` resolves to a specific Envoy filter state key, which is tied to a single ext-proc instance in the HCM filter chain. Two ext-proc instances sharing the same user-assigned `name` in the same HCM filter chain are inherently ambiguous.

Each Gateway listener translates to an IR `HTTPListener`. At xDS build time, each IR `HTTPListener` becomes an HCM and is attached to either a named `FilterChain` (TLS listeners, where SNI-based matching selects the chain) or the `DefaultFilterChain` (non-TLS listeners). Multiple routes under the same listener all share the same HCM and therefore the same ext-proc filter chain.

An `EnvoyExtensionPolicy` targeting an `HTTPRoute` attaches ext-proc filters to every IR route matched by that route, under every IR listener those routes belong to. Because all those routes share the same HCM at xDS time, two EEPs on the same listener that both assign the same `name` will produce two ext-proc instances with the same user name in the same HCM filter chain — the ambiguous case.

Conflict detection must happen before the IR is written, at the point where EG knows which IR listener a route belongs to. This is `translateEnvoyExtensionPolicyForRoute` in the Gateway API translation layer. EEPs are processed in creation-timestamp order (oldest first). A `nameRegistry` keyed by IR listener name tracks the first owner of each `name` per listener. When a later EEP claims a name already owned by a different ext-proc on the same IR listener, its `CustomName` is cleared in the IR so the access-log resolver never sees the duplicate. A `Warning` condition with reason `AmbiguousDefinition` is set on the losing policy.

When [`MergeGateways`](https://gateway.envoyproxy.io/docs/api/extension_types/#mergegatewaysconfig) is enabled, all gateways under the same `GatewayClass` share a single IR key. Non-TLS listeners on the same address and port are merged into a single `DefaultFilterChain` with a shared HCM at xDS time. However, each gateway still produces its own IR `HTTPListener` — they are distinct entries in the IR, looked up separately during translation.

Because the `nameRegistry` is scoped to individual IR listeners, name collisions **across** IR listeners (i.e. across merged gateways) are not detected, and no `Warning` condition is set. At xDS translation time, the IR listener belonging to the **oldest merged gateway** creates the shared HCM first and resolves its operators. Subsequent IR listeners patch the same HCM and resolve only operators that remain unresolved — first-writer-wins. Operators that go unresolved after all listeners are processed are replaced with `"-"`.

In practice: when using `MergeGateways` with non-TLS listeners, ensure names are unique across all merged gateways, or accept that the oldest gateway's ext-proc instance takes precedence with no warning issued.

### 5. Users opt in to what gets logged

The operator is explicit: users name both the extension instance and the filter state attribute they want. There's no automatic inclusion — EG can't know which attributes are meaningful for every ext-proc service, and blindly surfacing all of them would pollute logs.

### 6. Generic operator name: `EG_EXT_FILTER_STATE`

Named after the underlying Envoy mechanism (`FILTER_STATE`) rather than the specific extension type (`EXT_PROC`), leaving room to support other filter types that write filter state without a new operator. The `EG_` prefix distinguishes it from any native Envoy operator.

## Sharing Names Across Policies

Sharing a `name` across EEPs on the **same listener** is generally not recommended — only the oldest policy's instance is used for access log expansion, and the newer policy receives a `Warning` condition with reason `AmbiguousDefinition`.

The pattern works correctly when the sharing EEPs target routes on **separate listeners** (each listener has its own isolated HCM and filter chain). A common case: the same ext-proc logic is deployed in multiple namespaces for access-control reasons, each team's routes served by their own listener. A single `EnvoyProxy` format string then works uniformly across all of them, with each listener resolving the name against its own ext-proc instance independently.

When `MergeGateways` is enabled, non-TLS listeners on the same port share one HCM, so the isolated-listener guarantee does not hold for those listeners. See [MergeGateways (non-TLS)](#mergeGateways-non-tls) above for ordering and warning behaviour.

## Scope and Future Work

Currently `%EG_EXT_FILTER_STATE(...)%` is resolved only in access log format strings (text and JSON).

**Header mutation** (`AddRequestHeaders`, `AddResponseHeaders`) is a natural next target — Envoy evaluates `HeaderValue.Value` using the same format engine at request time, so resolved `%FILTER_STATE(...)%` operators would just work. Implementation requires threading the extension list through `buildXdsAddedHeaders` and applying the same resolver.

**Other extension types** (WASM, Dynamic Modules, Lua) could use the same mechanism if they wrote filter state under their Envoy filter instance name. Ext-proc does this explicitly; WASM and Lua use user-chosen keys. Extending to those types is deferred until there's a concrete use case.

[Gateway API]: https://gateway-api.sigs.k8s.io/
[Policy Attachment]: https://gateway-api.sigs.k8s.io/reference/policy-attachment/

