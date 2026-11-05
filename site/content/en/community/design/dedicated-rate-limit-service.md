---
title: "Dedicated Rate Limit Service"
---

## Overview

This design lets one Envoy Gateway control plane run **dedicated rate limit services (RLS)**, each with its
own Redis connection, for individual Gateways or GatewayClasses. Today a single RLS and Redis are shared by
every Gateway the control plane manages.

Envoy Gateway provisions the RLS; users provision Redis. Several RLS instances may reference the same
Redis, so dedicated Redis capacity and failure isolation depend on the supplied backends.

Tracking issue: [#10520](https://github.com/envoyproxy/gateway/issues/10520)

## Background

Global rate limiting is configured once for the whole control plane, in `EnvoyGateway.rateLimit`. Envoy
Gateway deploys one `envoy-ratelimit` service backed by one Redis, and every Gateway's proxies call it with
the same timeout and failure behaviour.

Ordinary rate limit domains (`<namespace>/<gateway>/<listener>`) keep each Gateway's counters separate;
`shared: true` rules use policy domains to share counters across targets. Infrastructure and service
behaviour are shared in either case:

* **Capacity:** one Redis and one RLS deployment for all traffic.
* **Blast radius:** if one Gateway overloads or breaks the RLS or Redis, every Gateway is affected.
* **Behaviour:** `timeout` and `failClosed` apply to all Gateways.

In multi-team clusters this is often not enough isolation. The only workaround today is running a
separate control plane per Gateway, which is heavy when the rate limit backend is the only thing that
needs to be isolated.

## Goals

* Let a Gateway or a GatewayClass use its own RLS, optionally backed by its own Redis.
* Let `timeout`, `failClosed`, telemetry and connection settings to the RLS differ between them.
* Keep the shared RLS working exactly as today; dedicated services are opt-in.
* Allow global rate limiting with a dedicated RLS even when no shared RLS is configured.

## Non-Goals

* User-managed (bring-your-own) RLS in the first iteration. Envoy Gateway still deploys and configures
  the RLS. A user-managed RLS may be proposed separately as a follow-up.
* Rate limit backends other than Redis.
* Dedicated RLS on the host (standalone) provider in the first iteration.
* Changing how rate limit rules or domains are generated.
* Uninterrupted rate limiting while the dedicated RLS configuration is being changed.

## Assumptions

A dedicated RLS is infrastructure: a platform or Gateway owner sets it up once, together with its Redis,
and it normally stays unchanged for the life of the Gateway or GatewayClass. Changing its scope, Redis
backend or credentials is expected to be rare and planned.

This design therefore optimizes for a simple, correct steady state rather than for seamless changes.
Changes are reconciled the way other Kubernetes resources are: Envoy Gateway converges to the new desired
state, and rate limiting may be briefly degraded, or briefly not enforced, while it does (see
[Changing the configuration](#changing-the-configuration)). If this assumption turns out to be wrong,
smoother changes (for example, migrating between services before removing the old one) can be added
later without changing the API.

## API

The setting lives on `EnvoyProxy`, the existing API for per-Gateway and per-GatewayClass infrastructure.
`EnvoyProxy` already attaches to a GatewayClass (`spec.parametersRef`) or a Gateway
(`spec.infrastructure.parametersRef`), so no new attachment mechanism is needed.

A new optional `spec.rateLimit` field holds the same settings as `EnvoyGateway.rateLimit`:

| Field             | Description                                                                     |
| ----------------- | ------------------------------------------------------------------------------- |
| `backend`         | Required Redis connection for the dedicated RLS (URL or Secret reference, TLS). |
| `timeout`         | How long proxies wait for the RLS. Defaults to 20ms.                            |
| `failClosed`      | Deny traffic when the RLS can't be reached. Defaults to false.                  |
| `telemetry`       | Metrics and tracing for the dedicated RLS.                                      |
| `serviceSettings` | Connection settings from proxies to the RLS (circuit breakers, etc.).           |

Deployment settings for the dedicated RLS (`rateLimitDeployment`, `rateLimitHpa`, `rateLimitPDB`) sit under
`spec.provider.kubernetes`, next to the existing proxy deployment settings.

These use the same types as the proxy settings (`KubernetesDeploymentSpec` and friends). `patch` is applied
only when the `EnvoyProxyPatch` runtime flag is enabled, the same as for the proxy Deployment.

> **Security risk.** The dedicated RLS runs in the controller namespace (see
> [Namespace and certificates](#namespace-and-certificates)), but its workload settings come from
> `EnvoyProxy`, which is often authored by Gateway owners. Whoever can author that `EnvoyProxy` can use
> `patch`, `volumes`, `securityContext` and similar fields to configure a pod in the controller namespace,
> for example to mount a Secret there or run with elevated privileges. Proxy Deployment settings already
> carry this risk outside Gateway namespace mode. Dedicated RLS introduces it even in Gateway namespace
> mode, because the RLS still runs in the controller namespace.
>
> Before enabling tenant-authored dedicated RLS configuration, admins must either restrict creation and
> updates of the supplying `EnvoyProxy` to trusted platform administrators through RBAC, or enforce an
> admission policy that allows only approved workload settings. The policy must cover deployment, HPA
> and PDB settings, including patches, and allowlist safe fields and values rather than rely on a partial
> denylist. Controller-namespace Secret references, custom images, environment variables, volumes and
> mounts, security settings and ServiceAccount overrides all need to be covered. The same restrictions
> must apply to every `EnvoyProxy` attachment that can supply these settings.
>
> Disabling `EnvoyProxyPatch` is an additional restriction, not sufficient protection: ordinary
> `KubernetesDeploymentSpec` fields can also select images and reference controller-namespace Secrets.
> Tenant-authored configuration is supported only with the admission restrictions above; otherwise
> dedicated RLS configuration must remain administrator-authored.

`rateLimit` is not accepted in the control-plane default `EnvoyGateway.envoyProxy`. Control-plane-wide
rate limiting stays configured in one place, `EnvoyGateway.rateLimit`.

### Example

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: team-a
  namespace: team-a
spec:
  rateLimit:
    backend:
      type: Redis
      redis:
        urlRef:
          secretKeyRef:
            name: team-a-redis
            key: url
    timeout: 50ms
    failClosed: true
  provider:
    type: Kubernetes
    kubernetes:
      rateLimitDeployment:
        replicas: 3
---
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: team-a
  namespace: team-a
spec:
  gatewayClassName: eg
  infrastructure:
    parametersRef:
      group: gateway.envoyproxy.io
      kind: EnvoyProxy
      name: team-a
  listeners:
  - name: http
    protocol: HTTP
    port: 80
```

## Scope

Where `rateLimit` is set decides who shares the dedicated RLS:

| `rateLimit` set on                                  | Result                                                          |
| --------------------------------------------------- | --------------------------------------------------------------- |
| The Gateway's `EnvoyProxy`                          | One RLS for that Gateway.                                       |
| The GatewayClass's `EnvoyProxy`                     | One RLS shared by every Gateway in the class.                   |
| Class configuration, with `mergeGateways` enabled   | One RLS for the merged Gateways; Gateway-level `parametersRef` is invalid. |
| No `EnvoyProxy` level, `EnvoyGateway.rateLimit` set | The shared control-plane RLS (today's behaviour).               |
| No `EnvoyProxy` level, no shared RLS                | Global rate limits are rejected (today's behaviour).            |

Sharing an RLS per GatewayClass trades isolation for footprint: Gateways in the class share capacity and
a failure domain again, but with fewer RLS deployments and a common Redis connection. This fits when the
GatewayClass is already the team boundary.

Ordinary rules keep separate counters through their Gateway listener domains. Rules with `shared: true`
instead use the existing policy domain and share a bucket across that policy's targets using the same
Redis backend. This design preserves those domains and Redis keys; an RLS scope is not an additional
counter namespace.

When a policy's targets use RLS scopes backed by different Redis databases, its shared bucket splits into
one independent bucket per database. For example, a shared 100rps rule spanning two such backends permits
up to 100rps on each backend, rather than 100rps in total.

Whenever a shared rule's effective targets span multiple RLS scopes, including the shared control-plane
scope, the policy reports `SharedRateLimitCrossScope=True` with reason `MultipleRLSScopes`. The warning
identifies the scopes and explains that the bucket may split depending on their Redis backends; the
policy remains accepted. It is based only on resolved scope identities, without comparing Redis URLs,
Secret contents or database identities, and is reported even when the scopes have identical backend
settings. The condition is removed when no shared rule's effective targets span multiple scopes. Users
needing a common bucket should keep those targets on the same RLS scope.

Separate RLS scopes pointing at the same Redis database can still share counters for `shared: true`
rules with the same policy domain and descriptors. Dedicated deployments alone do not isolate those
counters. Conversely, different Redis backends cannot provide a common bucket merely because the policy
is shared. The user guide must explain both cases; rules and domains remain generated as today.

### Inheritance from the GatewayClass

**Replace (no `mergeType`).** A Gateway-level `EnvoyProxy` replaces the class one completely. A Gateway
that brings its own `EnvoyProxy` for an unrelated reason, such as replicas, therefore leaves the class RLS:
it falls back to the shared RLS, or gets its own if it repeats the `rateLimit` block. Users who want such a
Gateway to stay on the class RLS set `mergeType`. The user guide should call this out.

**StrategicMerge and JSONMerge.** These merge `EnvoyProxy` field by field today. This design adds new
behaviour for `rateLimit`: the block is treated as a single unit instead of being merged field by field.

* A Gateway that does not set `rateLimit` inherits the class block and stays on the class RLS.
* A Gateway that sets `rateLimit` replaces the class block in full and gets its own RLS, even when the
  values equal the class configuration. The block must include `backend`, so an `EnvoyProxy` patch that
  only sets `timeout` is rejected when it is applied.

This requires Envoy Gateway to track which level supplied the `rateLimit` block, not just the merged result.
Scope is determined by attachment and inheritance; there is no explicit `scope` field.

**Both levels set `rateLimit`.** In every mode above, the Gateway's block wins: the Gateway gets its own
RLS and the class RLS keeps serving the other Gateways in the class. Shared rules spanning this Gateway and
Gateways on the class RLS report `SharedRateLimitCrossScope`, as described in [Scope](#scope). If the
Gateway's backend uses a different Redis database from the class, adding the Gateway block splits those
shared buckets and so loosens their effective total limit; the user guide should call this out.

**Merged Gateways.** With `mergeGateways` enabled, the class `rateLimit` block applies to the merged
Gateways. A Gateway's `spec.infrastructure.parametersRef` must be unset, as required today. Attaching a
Gateway-level `EnvoyProxy` is invalid, including one supplying `rateLimit`; it is rejected rather than
ignored with a warning. This design does not change that attachment restriction.

**Workload settings.** RLS workload settings follow the scope of the selected `rateLimit` block. A class
RLS uses the class `EnvoyProxy`'s `rateLimitDeployment`, `rateLimitHpa` and `rateLimitPDB`. If a Gateway
on the class RLS sets these fields, they are ignored and the Gateway reports a warning condition; the
Gateway is not rejected. A Gateway RLS uses only its own `EnvoyProxy`'s workload settings, without
inheriting class workload settings, even with StrategicMerge or JSONMerge. For example, a Gateway that
supplies a new `rateLimit.backend` but no workload settings gets the built-in RLS workload defaults,
not the class's replicas, HPA or PDB.

All three workload fields are resolved from the unmerged `EnvoyProxy` that supplied the selected
`rateLimit` block, then built-in defaults are applied. Neither `EnvoyGateway.envoyProxy` nor the shared
RLS settings under `EnvoyGateway.provider.kubernetes` supply defaults for dedicated RLS workloads.
RLS workload fields in `EnvoyGateway.envoyProxy` are rejected when loading the control-plane config.
The shared RLS continues to use its existing control-plane workload settings. Normal merge rules still
apply to unrelated proxy fields.

## How it works

1. **Resolution.** When translating each Gateway, Envoy Gateway works out which RLS it uses (dedicated or
   shared) from the scope rules above. Global rate limit policies are accepted for a Gateway when it has
   either kind of RLS.
2. **Deployment.** Envoy Gateway deploys one RLS per distinct dedicated scope, named after that scope
   (for example `envoy-ratelimit-<hash>`). It keeps the RLS while at least one Gateway uses it and removes
   it once the last Gateway leaves.
3. **Rate limit configuration.** Each RLS receives only the rate limit rules of the Gateways that use it.
   The shared RLS keeps receiving the rules of all Gateways
   without a dedicated RLS. A shared policy domain is included in the configuration of every RLS that
   serves at least one of the policy's targets. The configuration server runs whenever any RLS exists, even when no shared
   RLS is configured.
4. **Proxy configuration.** Each Gateway's proxies are pointed at their RLS, with that scope's `timeout`,
   `failClosed` and connection settings.

```text
                    ┌──────────────── Envoy Gateway ────────────────┐
                    │  resolves scope · deploys RLS · pushes rules  │
                    └──────┬──────────────────┬──────────────────┬──┘
                           │                  │                  │
 Gateway A (own RLS) ──▶ RLS-A ──▶ Redis-A    │                  │
 Gateways B, C (class) ─────────────▶ RLS-class ──▶ Redis-class  │
 Gateway D (no rateLimit) ──────────────────────────────▶ shared RLS ──▶ shared Redis
```

### Service identity

Each RLS has a stable identity: the Gateway (namespace/name), the GatewayClass name, or the existing shared
identity. Resource names and the configuration each RLS receives are keyed by this identity, so reusing
one `EnvoyProxy` across several attachments does not merge their services.

The identity depends only on scope, not on settings. Changing the settings of an existing RLS, including
its Redis backend, therefore updates that RLS in place rather than creating a new one; see
[Changing the configuration](#changing-the-configuration).

For dedicated RLS, explicit `rateLimitDeployment.name`, `rateLimitHpa.name` and `rateLimitPDB.name` are
not supported and are rejected as invalid parameters. Envoy Gateway generates the names of the Deployment,
Service, HPA, PDB and copied credential Secrets from the scope identity. Generated identities distinguish
Gateway and GatewayClass scopes and reserve the existing shared RLS identity. The HPA target and Service
and PDB selectors refer only to that scope's Deployment and pods; no selector spans dedicated scopes or
the shared RLS.

These identity fields are controller-owned even when `EnvoyProxyPatch` is enabled. Envoy Gateway validates
the rendered resources after applying workload settings and patches, rejecting changes to generated
resource names or namespaces, ownership metadata, scope labels and selectors, or the HPA target. Custom
pod labels cannot overwrite scope labels. These checks also apply to administrator-authored configuration
and prevent one scope's settings from redirecting reconciliation or traffic to another scope's resources.

### Namespace and certificates

In the first iteration, dedicated RLS deployments run in the controller namespace, like the shared one.
They reuse the existing mTLS certificates between proxies, the RLS and the control plane, so no new
certificate handling is needed. Proxies in Gateway namespaces can reach them as they reach the shared RLS
today. Running the RLS in the Gateway namespace (in Gateway namespace mode) is a possible follow-up.

### Redis credentials

Redis references are resolved against the `EnvoyProxy` that supplied the `rateLimit` block, even when a
Gateway inherits it. `urlRef.secretKeyRef` must be in that `EnvoyProxy`'s namespace. A TLS `certificateRef`
defaults to the same namespace; a cross-namespace reference requires a `ReferenceGrant` that allows that
`EnvoyProxy` to reference the Secret.

Because the RLS runs in the controller namespace, Envoy Gateway copies only the referenced keys into a
Secret it owns for that RLS. It keeps the copy in sync with the source Secret and deletes it when the RLS
is removed. A hash of the copied values is set as an annotation on the RLS pod template, so a credential
rotation triggers a normal rolling update and new pods pick up the new credentials.

**Missing or revoked credentials.** If a referenced Secret is deleted or its `ReferenceGrant` is revoked,
Envoy Gateway deletes its copy, updates the hash and the affected Gateways report a condition naming the
problem. The new pods cannot start
without the Secret, so the rolling update stalls and running pods keep serving with the credentials they
already loaded until they are replaced. Once the references are valid again, the rollout completes.

Deleting the source Secret is therefore **not** a way to cut off Redis access, and Envoy Gateway does not
provide an RLS stop switch in the first iteration. Removing the `rateLimit` block isn't one either: the
Gateways may move to the class or shared RLS instead, and a class RLS keeps running while other Gateways
use it. To stop a dedicated RLS from using credentials that must no longer be used, revoke them at Redis
(change the password or ACL, or revoke the client certificate). The running pods then lose access on
their next connection attempt, and rate limit checks follow `failClosed`.

Credential resolution failures do not change RLS selection or remove its desired infrastructure or
rules. Envoy Gateway continues translating the Gateway's listeners and routes, including its RLS cluster
and rate limit filter, and never falls back to another RLS because credentials are unavailable.
This also applies on first creation with a missing Secret: the Service and Deployment are provisioned,
but pods requiring the missing credential copy cannot start. Checks that cannot be served follow
`failClosed`. After deletion or revocation, existing pods may continue serving as described above.
These rules are derived from current resources on restart and do not depend on retaining an old snapshot.

### Validation and status

A `rateLimit` block without `backend` is rejected by schema validation when the `EnvoyProxy` is applied.
Other invalid settings, such as an unsupported `serviceSettings` field, are
reported the same way as other invalid `EnvoyProxy` settings: the affected Gateways are `Accepted=False`
with reason `InvalidParameters`. The `serviceSettings` restrictions are the same as for the
shared RLS.

Missing Secrets, missing or invalid credential keys, and denied or revoked Secret references are
operational credential failures, not fatal `EnvoyProxy` validation errors. They do not set Gateway
`Accepted=False` or enter the existing path that skips Gateway translation. Instead, the Gateway reports
a separate `RateLimitReady=False` condition with reason `InvalidCredentials` and a message identifying
the failed reference without exposing credential values. Selection, provisioning and proxy translation
continue as specified in [Redis credentials](#redis-credentials). Valid credentials clear this failure;
`RateLimitReady` becomes true only when all the readiness criteria below are met.

Each Gateway using global rate limiting with a dedicated RLS reports which RLS it uses and whether
that RLS infrastructure is ready.
The condition message identifies the selected scope and Service without exposing credentials. For a
dedicated RLS, `RateLimitReady=True` requires all of the following for the current desired configuration:

* All credential references resolve successfully and the owned credential copies are up to date.
* Infrastructure reconciliation has succeeded. The Deployment has observed its latest generation,
  completed the desired rollout, and has at least one available replica, with no old replicas remaining.
  Its pod template includes the current credential hash.
* The Service has at least one ready endpoint belonging to the current rollout.

The controller recomputes readiness as desired workload settings, credentials, workloads and endpoints
change. It does not retain `True` from the previous desired workload configuration while a new rollout
is pending. On restart, readiness remains unknown until the controller has observed the required state.
This condition describes credential resolution and Kubernetes infrastructure readiness. It does not
confirm that RLS endpoints have loaded the desired rules, can connect to Redis, or provide uninterrupted
enforcement. ACK-based configuration readiness is deferred to [TODO](#todo); rule migration still has the
limitations described below.

`RateLimitReady=False` uses the following reasons, in priority order when multiple failures apply:

| Reason | Meaning |
| ------ | ------- |
| `InvalidCredentials` | A required credential or its authorization cannot be resolved. |
| `ProvisioningFailed` | Creation or update of a required Kubernetes resource failed. |
| `RolloutFailed` | The Deployment reports that the desired rollout exceeded its progress deadline. |
| `Provisioning` | Required resources have not yet been observed. |
| `RollingOut` | The current Deployment rollout has not completed. |
| `NoReadyEndpoints` | The Service has no ready endpoint from the current rollout. |

An unknown observation state uses `RateLimitReady=Unknown` with reason `PendingObservation`. Successful
convergence uses reason `Ready`. A Gateway that no longer uses global rate limiting has this condition
removed, as does a Gateway that switches to the shared RLS. Shared RLS status and credential validation
retain their existing control-plane behavior.

### Host provider

Not supported in the first iteration. Setting `rateLimit` on an `EnvoyProxy` used with the host provider
makes the Gateway `Accepted=False` with a message saying so.

### Tooling

`egctl` can show the rate limit configuration of a dedicated RLS as well as the shared one, and
`egctl x translate` picks the settings up from the `EnvoyProxy` input.

## Changing the configuration

Following the [assumptions](#assumptions), any change to a dedicated RLS (new settings, a new Redis
backend, or a new scope such as moving a Gateway from the class RLS to its own) is reconciled toward the
desired state, like other Kubernetes resources. Envoy Gateway creates or updates the RLS, publishes the
Gateway's rules to it, points the proxies at it, and removes a dedicated RLS once no Gateway uses it. A
control plane restart simply reconciles again.

While a change rolls out, rate limiting may be degraded in two ways:

* **The RLS can't be reached** (not ready yet, or already removed). Those requests follow `failClosed`.
* **The RLS is reachable but no longer has the Gateway's rules.** This happens when the rules have
  already moved to the new RLS but some proxies still call the old one, which is still running for other
  Gateways or not yet removed. The RLS treats an unknown domain as having no limits and allows the
  request, so **rate limits are not enforced for those requests even with `failClosed: true`**.

Both gaps last until the proxies receive their new configuration, normally seconds. Envoy Gateway does
not keep old rules on the previous RLS until proxies have switched; that would need the migration support
this design leaves out. Users who need continuous enforcement should make such changes during a low-traffic window.

## Upgrade and compatibility

Without `EnvoyProxy.spec.rateLimit`, nothing changes: same shared RLS, same resource names, same
behaviour.

## Alternatives considered

* **Make the GatewayClass RLS mandatory when configured.** If the GatewayClass sets `rateLimit`, every
  Gateway in that class uses the class RLS, regardless of `mergeType`. A Gateway-level `rateLimit` is
  rejected with Gateway `Accepted=False`, reason `InvalidParameters`. Gateway-level configuration of
  unrelated proxy fields remains subject to existing rules. If the class has no `rateLimit`, a Gateway
  can configure its own RLS or fall back to the shared control-plane RLS.

  **Pros:** simpler scope resolution, with no need to track whether a merged `rateLimit` block came from
  the class or Gateway; a clear platform-owned RLS boundary; and unrelated Gateway proxy settings cannot
  accidentally move a Gateway off the class RLS. Gateway overrides cannot split shared buckets within
  that class across different Redis backends.

  **Cons:** no per-Gateway RLS exceptions within a class, so a Gateway needing separate capacity, Redis
  or failure behaviour must move to another GatewayClass. Adding `rateLimit` to a class invalidates
  existing Gateway-level `rateLimit` configurations, requiring a coordinated change. It also introduces
  a special inheritance rule: the class RLS remains effective even when the Gateway's `EnvoyProxy`
  otherwise replaces the class configuration.

* **Named rate limit profiles in the `EnvoyGateway` config**, referenced by name from `EnvoyProxy`. This
  keeps Redis credentials admin-only, but `EnvoyGateway` changes need a control plane restart, and it adds
  a second place to configure per-Gateway infrastructure.
* **A new `RateLimitService` resource** referenced from `EnvoyProxy` or `BackendTrafficPolicy`. This lets
  any set of Gateways share one dedicated RLS, but it is a new API. Attaching `rateLimit` to a
  GatewayClass's `EnvoyProxy` or merging Gateways covers the common sharing cases with the existing API.
* **Choosing the RLS per `BackendTrafficPolicy`.** A proxy would then talk to several RLS instances, which
  makes the proxy configuration much more complex and gives Gateways no isolation by default.
* **Multiple control planes** (today's workaround). It works, but it is heavy when only the rate limit
  backend needs to be isolated.

## Proposed decisions

These are the author's recommendations and are open for review.

* `backend` is required in every `rateLimit` block. Overriding only `timeout` or `failClosed` while
  keeping the shared RLS is left to a separate shared-mode design.
* Scope comes from attachment and inheritance; no explicit `scope` field is added.
* Shared rules retain their policy domains and Redis keys. Shared targets spanning multiple RLS scopes
  produce a policy warning that their bucket may split depending on the Redis backends. The warning
  uses scope identities; separate scopes using the same database can still share counters.
* `rateLimit` is not allowed in the control-plane default `EnvoyGateway.envoyProxy`.
* Gateway overrides of class RLS workload settings are ignored with a warning, not rejected.
* When both the Gateway and its GatewayClass set `rateLimit`, the Gateway block wins and gets its own RLS.
  With `mergeGateways`, Gateway-level `parametersRef` remains invalid, so only the class can supply
  dedicated RLS settings.
* Dedicated RLS workload settings come only from the unmerged `EnvoyProxy` supplying `rateLimit`, with
  built-in defaults for omitted settings; class and control-plane workload settings are not inherited.
* Dedicated RLS resource names, ownership, scope labels and selectors, and HPA targets are
  controller-owned. Explicit workload names and patches that change those fields are rejected.
* Dedicated RLS workload settings use the existing Kubernetes types. `patch` is gated by
  `EnvoyProxyPatch`, but disabling patches alone is insufficient protection. Configuration must be
  administrator-authored or constrained by an admission allowlist covering all workload settings.
* Dedicated RLS deployments run in the controller namespace in the first iteration. Running them in the
  Gateway namespace, and the authentication that needs, is deferred.
* Envoy Gateway copies the referenced Redis Secret keys into the controller namespace and keeps the copy
  in sync; a config hash on the pod template rolls the RLS on rotation. Revoked credentials stall that
  rollout, so running pods are not actively stopped; credentials are revoked at Redis.
* Credential resolution failures report `RateLimitReady=False` without rejecting the Gateway, changing
  RLS selection, or skipping infrastructure, rule or proxy reconciliation, including on first creation.
* Dedicated `RateLimitReady=True` requires the current workload rollout, ready Service endpoints and
  valid credentials. ACK-based configuration readiness and changes to shared RLS status are deferred.
* Dedicated RLS configuration is assumed to be set up once and rarely changed. Changes are plain
  reconciliation with no migration support. While they converge, unreachable RLS calls follow
  `failClosed`, and calls to an RLS that no longer has the Gateway's rules are allowed.
* User-managed RLS is deferred to a separate proposal.

## TODO

The following concerns also apply to the existing shared RLS mode. They are follow-up work and are not
requirements for introducing dedicated RLS.

### ACK-based configuration readiness

Consider a separate design for configuration readiness across both shared and dedicated RLS. It would
require every ready RLS endpoint to acknowledge the desired snapshot on its current stream, correlating
pod and scope identities and distinguishing pending configuration from a rejected configuration.
Infrastructure readiness alone does not provide this guarantee.

The cross-replica approach below is a candidate for that follow-up, rather than an implementation
requirement of this proposal. Its RPC, timing and identity choices need separate review.

#### ACK observation with multiple controller replicas

RLS configuration streams may connect to any Envoy Gateway replica. Only the elected status writer
publishes Gateway conditions, so readiness cannot depend solely on streams connected to that replica.
Each configuration server exposes a controller-internal, mutually authenticated observation RPC that
reports its current RLS streams. The elected status writer queries every live controller replica serving
the configuration service and combines these observations with Kubernetes workload and endpoint state.
This RPC is restricted to controller identities and does not expose configuration contents or credentials.

Each observation includes the controller pod UID and process incarnation, RLS pod UID, scope identity,
stream identity, and the last ACK or NACK for that stream. RLS pods supply their pod UID separately from
the scope used to select their snapshot. Snapshot versions use a deterministic hash of the canonical
scope configuration, including an empty configuration, rather than a replica-local incrementing counter.
The server matches response nonces to these versions before recording an ACK or NACK. Thus observations
from different replicas can be compared against the status writer's desired snapshot without treating
unrelated local version numbers as equivalent.

Observations are fetched at least every five seconds, expire after fifteen seconds, and are invalidated
immediately when a stream closes or its server process changes. Missing or expired observations produce
`PendingObservation`; a known current stream without the desired ACK produces `ConfigurationPending`.
During a reconnect, multiple observed streams for one RLS pod are treated conservatively: all must have
ACKed the desired snapshot until closure of the superseded stream is observed. A failed replica must be
observed as terminated before its streams are discarded; an unreachable replica is not assumed dead.
This bounds stale readiness after an unobserved disconnection without requiring RLS streams to connect
to the elected controller.

On leadership change, the new status writer starts with no trusted ACK observations, publishes
`PendingObservation`, and obtains fresh responses from all serving replicas before reporting readiness.
Persisted Gateway conditions and observations from a previous controller incarnation are never used as
evidence of an ACK. The same aggregation applies to dedicated and shared RLS scopes.

### Counter behavior when changing Redis backends

Clarify counter reuse in general rate-limit operational guidance. Changing a Redis backend does not
guarantee counters start at zero: matching, unexpired keys on the destination still apply, including
when returning to a previously used backend. This also applies to changes to the existing shared RLS
backend and does not require dedicated-mode counter migration or reset support.

### Workload overrides of RLS runtime configuration

Review which runtime settings should remain controller-owned across both shared and dedicated RLS.
Existing deployment settings can override generated container environment variables, including
`CONFIG_GRPC_XDS_NODE_ID`, `CONFIG_GRPC_XDS_SERVER_URL` and `CONFIG_TYPE`; workload patches can also
change them. Such overrides can already prevent the shared RLS from loading its intended rules.
Dedicated mode adds the possibility of selecting another scope's snapshot, but the general ability to
replace runtime configuration predates this mode.

A follow-up should define supported overrides and compatibility before introducing restrictions on
these fields. Dedicated mode must generate the correct snapshot identity for each scope; broader
validation of user-supplied runtime overrides is deferred and is not an additional requirement here.
