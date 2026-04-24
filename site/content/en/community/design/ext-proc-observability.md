---
title: "Ext-Proc Observability: Named Instances"
---

## Overview

Envoy's [ext-proc filter](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/ext_proc_filter) emits per-stream statistics (latency, gRPC status, bytes transferred, etc.) and writes per-stream data into [filter state](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/advanced/filter_state), both keyed by the filter's instance name. Envoy Gateway generates those instance names from policy identity (e.g. `envoy.filters.http.ext_proc/namespace/extproc/my-policy/0`), following a convention that is opaque to users.

This creates two related observability problems:

1. **Access logs**: users who want to include ext-proc data in access logs must hand-write `%FILTER_STATE(envoy.filters.http.ext_proc/...:FIELD:request_header_latency_us)%` with the full internal name, which is verbose and breaks whenever the policy is renamed or reordered.

2. **Metrics**: Envoy emits stats under `ext_proc.<stat_prefix>.*` where `stat_prefix` defaults to the same opaque internal name, making dashboards and alerts hard to write and maintain.

This document describes two complementary features introduced on the `extProc` entry:

- **`name`**: an optional user-defined identifier for an ext-proc instance. Setting it enables the `%EG_EXT_PROC_FILTER_STATE(name:args)%` access log operator, which EG resolves at xDS translation time to the real `%FILTER_STATE(...)%` operator. Users reference instances by the friendly name they assigned rather than by EG's internal filter names.
- **`statPrefix`**: an optional field that controls the Envoy stat prefix for the filter instance (e.g. `ext_proc.<statPrefix>.streams_started`). Defaults to `name` when `name` is set, otherwise Envoy uses its own default. Shared prefixes aggregate stats across deployments intentionally; distinct prefixes isolate per-deployment metrics.

Both features are opt-in and independent of each other.

## Goals

- Let users reference ext-proc filter state in access log format strings by a name they control.
- Let users control the Envoy stat prefix for ext-proc filter instances to get readable, stable metric names.
- Allow independent control of the log alias (`name`) and the stat prefix (`statPrefix`), supporting use cases where logs should aggregate but metrics should isolate (or vice versa).
- No coupling between `EnvoyProxy` (format strings) and `EnvoyExtensionPolicy` (extension config) beyond the user-chosen name.

## Non Goals

- Runtime resolution (all substitution happens at xDS generation time).
- Letting users control internal filter names (they assign aliases only).
- Automatic expansion into multiple matching instances for a single operator.
- Support for WASM, Dynamic Modules, and Lua in this iteration (they don't write filter state under their instance name by default).

## Design Decisions

| # | Decision | Alternatives considered |
|---|---|---|
| 1 | **Two independent fields** `name` and `statPrefix` with separate conflict semantics; `statPrefix` defaults to `name` when set, otherwise Envoy's own default is used. | Single field doubling as both (rejected: can't isolate per-deployment metrics while sharing a log name). |
| 2 | **`name` is a user alias only**; EG owns internal filter names generated from policy identity. | Letting users set the internal filter name (rejected: risks naming conflicts and breaks HCM uniqueness). |
| 3 | **Duplicate `statPrefix` values are allowed** and aggregate intentionally, mirroring `clusterStatName` semantics; no warning issued. | Treating duplicates as an error like `name` (rejected: aggregation across deployments is a valid pattern). |
| 4 | **`args` are passed verbatim** to `%FILTER_STATE(key:args)%`; EG does not inject or validate them. | EG-managed attribute selection (rejected: EG can't know which filter state fields are meaningful per service). |
| 5 | **Unresolved operators become `[EG_UNRESOLVED:name]`** rather than passing through or being dropped. | Pass-through (rejected: causes xDS NACK); silent drop (rejected: hides misconfiguration; the sentinel is distinct from Envoy's runtime `-`). |
| 6 | **First-claimant-wins per IR listener** using a `nameRegistry` in creation-timestamp order; losers get `CustomName` cleared and an `AmbiguousDefinition` warning. Under `MergeGateways`, cross-listener collisions are undetected (replicating FilterChain-sharing knowledge in the GW-API layer would leak xDS concerns upward). | Rejecting all duplicates (rejected: breaks the oldest, authoritative policy). |
| 7 | **Operator requires explicit attribute selection**; no automatic inclusion of all ext-proc filter state fields. | Auto-include all fields (rejected: EG can't know which are meaningful; would pollute logs). |
| 8 | **Operator named `EG_EXT_PROC_FILTER_STATE`**, scoped to the extension type. | Generic `EG_FILTER_STATE` (rejected: only ext-proc writes filter state keyed by its instance name; WASM, Lua, and Dynamic Modules use user-chosen keys, so a generic name would mislead). |

## Scenarios

| Scenario | `name` | `statPrefix` | Access log | Stats |
|---|---|---|---|---|
| Simple: one deployment, one listener | `auth-proc` | _(unset, defaults to `name`)_ | `%EG_EXT_PROC_FILTER_STATE(auth-proc:FIELD:bytes_sent)%` | `ext_proc.auth-proc.streams_started` |
| Isolated metrics: same logical service, multiple deployments | `auth-proc` (same on all) | `auth-proc-east` / `auth-proc-west` | Same operator resolves on every listener | Per-deployment metric families; aggregate by querying both prefixes |
| Multiple EEPs in one format string: a single `EnvoyProxy` access log format references ext-proc instances from different EEPs by name | `auth-proc` on EEP-1, `enrich-proc` on EEP-2 | _(unset)_ | `%EG_EXT_PROC_FILTER_STATE(auth-proc:...)%` and `%EG_EXT_PROC_FILTER_STATE(enrich-proc:...)%` both resolve in the same format string | `ext_proc.auth-proc.*` and `ext_proc.enrich-proc.*` independently |

## Scope and Future Work

**Header mutation** (`AddRequestHeaders`, `AddResponseHeaders`) is a natural next target; implementation requires threading the extension list through `buildXdsAddedHeaders` and applying the same resolver.

**Other extension types** (WASM, Dynamic Modules, Lua) could use the same mechanism if they wrote filter state under their Envoy filter instance name, but are deferred until there's a concrete use case.

**MergeGateways conflict detection** is a known limitation; the `nameRegistry` is scoped to individual IR listeners. Long-term, the IR itself should express which listeners share a FilterChain, removing the need to re-derive grouping at xDS translation time.

[Gateway API]: https://gateway-api.sigs.k8s.io/
[Policy Attachment]: https://gateway-api.sigs.k8s.io/reference/policy-attachment/
