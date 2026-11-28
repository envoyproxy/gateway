---
title: "Dedicated Rate Limit Service"
---

## Overview

This design lets one Envoy Gateway control plane run **dedicated rate limit services (RLS)**, each with its
own Redis connection, for individual Gateways. Today a single RLS and Redis are shared by every Gateway
the control plane manages.

Envoy Gateway provisions the RLS; users provision Redis. Several RLS instances may reference the same
Redis, so dedicated Redis capacity and failure isolation depend on the supplied backends.

This proposal covers only the Gateway scope. A dedicated RLS shared by a GatewayClass is out of scope:
`rateLimit` on a GatewayClass's `EnvoyProxy` is ignored, and may be supported in the future (see
[TODO](#gatewayclass-scope)).

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

* Let a Gateway use its own RLS, optionally backed by its own Redis.
* Let `timeout`, `failClosed`, telemetry and connection settings to the RLS differ between Gateways.
* Keep the shared RLS working exactly as today; dedicated services are opt-in, both for the admin
  (a control-plane setting) and per Gateway.
* Allow global rate limiting with a dedicated RLS even when no shared RLS is configured.
* Accept calls to a dedicated RLS only from its own Gateway's proxies, authenticated with a
  per-Gateway client certificate rather than inferred from the network.

## Non-Goals

* A dedicated RLS per GatewayClass, or one shared by any other set of Gateways, in the first iteration.
  Each dedicated RLS serves exactly one Gateway.
* Dedicated RLS for merged Gateways (`mergeGateways`) in the first iteration.
* User-managed (bring-your-own) RLS in the first iteration. Envoy Gateway still deploys and configures
  the RLS. A user-managed RLS may be proposed separately as a follow-up.
* Rate limit backends other than Redis.
* Dedicated RLS with any infrastructure other than Envoy Gateway's built-in Kubernetes infrastructure
  in the first iteration.
* Changing how rate limit rules or domains are generated.
* Validating RLS workload overrides or guaranteeing they are safe. Overrides are gated behind a
  default-off switch, and admins enforce their own requirements with tools such as Kyverno (see
  [Trust boundary](#trust-boundary)).
* Per-Gateway `timeout`, `failClosed` or other proxy-side settings while keeping the shared RLS. That
  needs a separate shared-mode design.
* Redis migration, including data transfer, counter synchronization, backend cutover orchestration
  and rollback. Users provision and operate Redis outside Envoy Gateway.
* Seamless switching between shared and dedicated RLS, or uninterrupted enforcement during a scope
  change. These are planned, potentially disruptive operations.
* Per-Gateway client identities for the shared RLS. Proxies keep presenting the shared client
  certificate to the shared RLS, which keeps accepting every proxy as today.
* Automatic certificate renewal or revocation before expiry. Operators manage certificate expiry (see
  [RLS identities and certificates](#rls-identities-and-certificates)).
* Cleanup of changes Envoy Gateway did not observe. A dedicated RLS is removed when Envoy Gateway
  observes its Gateway being deleted or dropping `rateLimit`, as proxy resources are today. Changes
  made while Envoy Gateway is stopped, turning off `enableDedicatedRateLimit` and switching the
  deployment mode leave dedicated RLS resources behind; admins remove them manually (see [Cleanup](#cleanup)). An automatic sweep is
  deferred (see [TODO](#sweeping-stale-dedicated-rls-resources)).
* Multiple Envoy Gateway installations managing Gateways in the same namespace while dedicated RLS is
  enabled. This design assumes one controller per namespace in the first iteration; that combination
  is unsupported.
* Hardening tenant namespaces or proxy workloads. Proxy `patch`, privileged containers, host
  networking and host path mounts can already let a tenant's proxy escape its namespace's network
  isolation, for every namespace-based NetworkPolicy and not only for the RLS. This proposal relies on
  operator's existing multi-tenant controls. This proposal introduces no admission hook or built-in
  validation for tenant hardening. Operators choose and maintain those controls and accept the risks
  of enabling the feature without them. Changing the `EnvoyProxyPatch` default is also out of scope.

## Assumptions

A dedicated RLS is infrastructure: a platform or Gateway owner sets it up once, together with its Redis,
and it normally stays unchanged for the life of the Gateway. Switching a Gateway between the shared and a
dedicated RLS, or changing its Redis backend or credentials, is expected to be rare and planned.

The first iteration optimizes for a correct steady state and ordinary Kubernetes reconciliation.
Changes to RLS scope are planned maintenance operations, not coordinated migrations. Operators who
require continuous rate-limit enforcement must stop or externally gate affected traffic before a scope
change and restore it only after verifying the new configuration. A low-traffic window alone does not
provide that guarantee. Credential rotation remains supported without changing the selected RLS.

**xDS clients are trusted.** This proposal requires that every client of the Envoy Gateway xDS server
requests only its own Gateway's configuration. In practice, proxy pods run only the Envoy image and
containers that Envoy Gateway generates, and no other process can authenticate to the xDS server. This
is a precondition, not a property the design provides: if it does not hold, isolation between dedicated
RLS does not hold either, and the guarantees in this proposal no longer apply.

This is not new with this proposal; Envoy Gateway already relies on it today. The xDS server does not
bind clients to their Gateway: in either deployment mode, a client chooses the snapshot it receives.
A custom proxy image, an extra container in the proxy pod, or any other process with valid xDS client
credentials can already request another Gateway's configuration, including the private keys of its
listener certificates and backend client certificates. A dedicated RLS client certificate is one more
secret in that snapshot and is exposed the same way. Operators who let Gateway owners run such
workloads must already prevent this, as part of proxy hardening (see [Non-Goals](#non-goals)).
Authenticating xDS clients per Gateway is a separate xDS change, out of scope for this proposal.

## API

The setting lives on `EnvoyProxy`, the existing API for per-Gateway infrastructure. A Gateway already
references an `EnvoyProxy` in its own namespace through `spec.infrastructure.parametersRef`, so no new
attachment mechanism is needed. Only this Gateway-level attachment enables a dedicated RLS; see
[GatewayClass EnvoyProxy](#gatewayclass-envoyproxy).

A new optional `spec.rateLimit` field holds the same settings as `EnvoyGateway.rateLimit`:

| Field             | Description                                                                     |
| ----------------- | ------------------------------------------------------------------------------- |
| `backend`         | Required Redis connection for the dedicated RLS (URL or Secret reference, TLS). |
| `timeout`         | How long proxies wait for the RLS. Defaults to 20ms.                            |
| `failClosed`      | Deny traffic when the RLS can't be reached. Defaults to false.                  |
| `telemetry`       | Metrics and tracing for the dedicated RLS.                                      |
| `serviceSettings` | Connection settings from proxies to the RLS (circuit breakers, etc.).           |

`spec.rateLimit` reuses the `EnvoyGateway.rateLimit` type. Its `failClosed` is a plain boolean
defaulting to `false`, so an omitted field and an explicit `false` behave the same way. Today that field
has neither `omitempty` nor an `+optional` marker, so the generated EnvoyProxy CRD would make it required
inside `rateLimit` and reject an `EnvoyProxy` that omits it. As part of this API change,
`RateLimit.FailClosed` gains `+optional` and `omitempty`. It stays a `bool`, so the meaning of an
omitted value is unchanged and existing `EnvoyGateway` configurations remain valid.

Reusing the `RateLimit` type also reuses its API documentation, which today describes only the shared
RLS. In particular, `RateLimitRedisSettings.urlRef` and `RedisURLSource.secretKeyRef` say the Secret
must be in the namespace of the Envoy Gateway rate limit deployment, which is wrong for `EnvoyProxy`
(see [Redis credentials](#redis-credentials)). As part of this API change, these descriptions are made
to cover both uses: in `EnvoyGateway.rateLimit` the Secret is in the controller namespace, and in
`EnvoyProxy.spec.rateLimit` it is in the `EnvoyProxy`'s namespace. The same review applies to every
other description in the reused types, including the TLS `certificateRef`, whose description must
state that in `EnvoyProxy.spec.rateLimit` it must be in the `EnvoyProxy`'s namespace, except for a
granted controller-namespace Secret in the default mode (see [Redis credentials](#redis-credentials)).

Deployment settings for the dedicated RLS (`rateLimitDeployment`, `rateLimitHpa`, `rateLimitPDB`) sit under
`spec.provider.kubernetes`, next to the existing proxy deployment settings.

These use the same types as the proxy settings (`KubernetesDeploymentSpec` and friends). `patch` is applied
only when the `EnvoyProxyPatch` runtime flag is enabled, the same as for the proxy Deployment.

> **Security risk.** A dedicated RLS runs in the same namespace as its Gateway's proxies (see
> [Namespace](#namespace)), and its workload settings come from `EnvoyProxy`, which is often authored
> by Gateway owners. Whoever can author that `EnvoyProxy` can use `patch`, `volumes`, `securityContext`
> and similar fields to configure the RLS pod, for example to mount a Secret in its namespace or run
> with elevated privileges. This is the same authority that `EnvoyProxy` already grants over the proxy
> pods in that namespace:
>
> * In Gateway namespace mode, the RLS runs in the Gateway's namespace. Its workload settings reach
>   only what the Gateway owner's proxies can already reach there.
> * In the default deployment mode, the RLS runs in the controller namespace, as the proxies do. Proxy
>   Deployment settings already let `EnvoyProxy` authors configure pods there, and dedicated RLS
>   workload settings add one more pod with the same reach. That includes the RLS CA Secret (see
>   [RLS identities and certificates](#rls-identities-and-certificates)): a pod that reads it can
>   impersonate any Gateway to any dedicated RLS, just as a proxy pod that mounts the control plane
>   Secrets there can already impersonate the control plane.
>
> **Workload overrides are not validated.** Envoy Gateway does not make them safe; that is why they
> sit behind a separate switch, `enableDedicatedRateLimitWorkloadOverrides`, that is off by default.
> Admins who enable it accept these risks and choose their own controls, such as restricting who can
> author the supplying `EnvoyProxy` through RBAC, or with admission policy tools such as Kyverno, OPA
> Gatekeeper or Kubernetes `ValidatingAdmissionPolicy`, as they would for proxy workload settings.
>
> With only `enableDedicatedRateLimit` enabled, Gateway owners still choose addresses that the RLS
> connects to: the Redis URL and the OTLP endpoint in `telemetry.tracing.provider.url`. In Gateway
> namespace mode these connections leave from the Gateway's namespace and are subject to its
> NetworkPolicies and quota, like the proxies' backend connections. In the default mode they leave from
> the controller namespace, as proxy backend connections already do in that mode.

### Operator responsibility for tenant isolation

Both dedicated RLS and its workload overrides are disabled by default. Enabling either switch is an
administrator decision to delegate the corresponding capabilities and accept their risks. Envoy Gateway
does not validate workload overrides for safety, enforce tenant hardening, or verify operator security
policies. This proposal adds no admission hook or tenant-hardening checks.

A dedicated RLS follows the proxies' namespace, so the controls operators already apply to a Gateway's
proxies also cover its RLS. In Gateway namespace mode, the namespace's ResourceQuota, LimitRange and
NetworkPolicies apply to the RLS pods, and the RLS pods carry the labels listed in
[Service identity](#service-identity) for policies that need to select them. Nothing new is placed in
the controller namespace.

The default mode offers no namespace boundary between tenants, for proxies or for RLS. Operators who
run multiple tenants in that mode should, as they already must for proxies:

* Limit who can read Secrets in the controller namespace, including the RLS CA Secret, and keep
  workload overrides, which can mount Secrets, away from tenant authors. Caller isolation between
  dedicated RLS relies on that key, not on namespaces or network policy.
* Plan capacity and quota for the controller namespace: each Gateway with a dedicated RLS adds a
  Deployment there, and a shared ResourceQuota alone does not reserve headroom for the controller.
* Restrict egress from proxy and RLS pods with NetworkPolicy if tenants must not reach internal
  services, using the RLS labels above.

These are operator-owned mitigations, not feature enablement requirements or implementation acceptance
checks. The installation guide documents the risks per deployment mode. Neither enabling the feature
nor `RateLimitReady=True` guarantees tenant isolation.

### Enabling dedicated RLS

Because of this risk, dedicated RLS is disabled by default. Admins opt in through a new field in the
`EnvoyGateway` configuration, next to the other switches for risky features such as
`enableEnvoyPatchPolicy`, `enableLua` and `enableBackend`:

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyGateway
extensionApis:
  enableDedicatedRateLimit: true   # defaults to false
  enableDedicatedRateLimitWorkloadOverrides: false   # defaults to false
```

The switch does not grant the controller the RBAC that dedicated RLS needs. That comes from a separate
Helm value, `dedicatedRateLimit.rbac.enabled`, because the switch can be changed at runtime without a
Helm upgrade (see [Namespace](#namespace)). Envoy Gateway does not change its own RBAC. Admins must
therefore confirm the controller's permissions each time they turn the feature on or off:

* **After turning it on**, confirm that the controller can write Secrets where dedicated RLS run, and,
  in the default mode with `watch.namespaces`, read `ReferenceGrant` resources in the controller
  namespace. For example, with a default Helm installation:

  ```shell
  SA=system:serviceaccount:envoy-gateway-system:envoy-gateway
  # Default mode: the controller namespace. Gateway namespace mode: each Gateway namespace.
  kubectl auth can-i create secrets --as=$SA -n <namespace>
  kubectl auth can-i delete secrets --as=$SA -n <namespace>
  kubectl auth can-i list referencegrants.gateway.networking.k8s.io --as=$SA -n envoy-gateway-system
  ```

  Missing permissions also surface as a startup warning and as `RateLimitReady=False` with reason
  `ProvisioningFailed` on each affected Gateway.
* **After turning it off**, check for dedicated RLS left behind in the namespaces this installation
  manages, for example with
  `kubectl get deployments -n <namespace> -l gateway.envoyproxy.io/ratelimit-scope=gateway`. Envoy
  Gateway does not remove them when the switch is turned off (see [Cleanup](#cleanup)). Once every
  proxy has moved to the shared RLS, delete them with the command in [Cleanup](#cleanup).

  Then decide whether to keep the permission. Keep it if the feature may be turned on again; otherwise set
  `dedicatedRateLimit.rbac.enabled: false` to return the controller to its previous privileges.
  Removal does not depend on the permission (see [Cleanup](#cleanup)).

The table shows what happens to a Gateway whose own `EnvoyProxy` sets `rateLimit`, for each
combination of the two switches. For an otherwise valid Gateway and `EnvoyProxy` configuration, the
switches alone do not change `Accepted=True`. The table describes steady-state behavior after
reconciliation.

| `enableDedicatedRateLimit` | `enableDedicatedRateLimitWorkloadOverrides` | Dedicated RLS | Workload fields¹ | Routes with global rate limits                                                       | Other routes | Gateway status                                              |
| -------------------------- | ------------------------------------------- | ------------- | ---------------- | ------------------------------------------------------------------------------------ | ------------ | ----------------------------------------------------------- |
| `false`                    | any (no effect)                             | Not deployed  | Ignored          | Same as without `rateLimit`: the shared RLS, or rejected if none is configured       | Unaffected   | `EnvoyProxy` warning; no dedicated `RateLimitReady`         |
| `true`                     | `false`                                     | Deployed      | Ignored, warning | Rate limited by the dedicated RLS                                                    | Unaffected   | `RateLimitReady` reflects the RLS                           |
| `true`                     | `true`                                      | Deployed      | Applied          | Rate limited by the dedicated RLS                                                    | Unaffected   | `RateLimitReady` reflects the RLS                           |

¹ `rateLimitDeployment`, `rateLimitHpa` and `rateLimitPDB`. When they are ignored, the RLS uses
built-in workload defaults.

Ignored settings produce warnings on `EnvoyProxy` status naming the switch to enable. A disabled
switch does not reject the Gateway or remove its listeners. With dedicated RLS disabled, adding a
`rateLimit` block cannot move a Gateway off the shared RLS or bypass its limits. If no shared RLS is
configured, global rate limit policies are rejected as today.

Disabling the setting after dedicated RLS is in use has the same effect: each affected Gateway moves
back to the shared RLS, if configured, with the transition behavior in
[Changing the configuration](#changing-the-configuration). The dedicated RLS is **not** removed: the
control plane reloads with the new setting and starts without the previous desired state, so nothing
deletes it. Admins remove it manually (see [Cleanup](#cleanup)).
Re-enabling the setting moves it to a dedicated RLS again. RLS workload fields on an `EnvoyProxy`
without `rateLimit` have no effect either way.

> **Warning: turning off `enableDedicatedRateLimit` is disruptive.** It is a control-plane-wide
> change that affects every Gateway with a dedicated RLS at once, not a per-Gateway rollback.
> Before turning it off, admins should understand that:
>
> * **Enforcement can be lost or traffic denied.** Proxies move to the shared RLS without
>   coordination. Until each proxy has switched, it keeps calling a dedicated RLS that no longer
>   receives rule updates, and once that RLS is removed its calls fail: with `failClosed: true` those
>   requests are denied, and with `failClosed: false` they pass unlimited. Unknown domains can also
>   bypass limits (see [Changing the configuration](#changing-the-configuration)).
> * **Leftover resources.** The dedicated RLS Deployments, Services and Secrets stay until an admin
>   deletes them (see [Cleanup](#cleanup)).
> * **Without a shared RLS, global rate limits stop.** If `EnvoyGateway.rateLimit` is not set, the
>   global rate limit policies of those Gateways are rejected, and their routes serve without global
>   rate limiting until the setting is re-enabled.
> * **Per-Gateway behavior reverts to the shared settings.** Each Gateway takes the shared RLS's
>   `timeout`, `failClosed`, telemetry and connection settings. A Gateway configured to fail closed can
>   start failing open.
> * **Isolation and capacity are gone.** All of that traffic moves to the shared RLS and its Redis,
>   which must have the capacity for it. Counters move to the shared keyspace, and `shared: true`
>   buckets that were separate per RLS merge into one, so combined limits can become stricter (see
>   [Shared rules](#shared-rules)).
> * **It is not an emergency switch.** It does not revoke Redis credentials or stop an RLS faster than
>   other changes; revoke credentials at Redis instead (see [Redis credentials](#redis-credentials)).
>
> To retire dedicated RLS safely, externally gate traffic to each affected Gateway during planned
> maintenance, then remove `rateLimit` from its `EnvoyProxy` one at a time. Reconciliation does not
> wait for proxies to switch before deleting the dedicated RLS; disconnected proxies can retain old
> configuration. Before restoring traffic, verify that every proxy that will serve it has the intended
> configuration and that the intended rate limits are enforced. Keep proxies whose configuration
> cannot be verified out of service. Turn the setting off only after no Gateway uses a dedicated RLS.

The two settings are control-plane-wide opt-ins. `enableDedicatedRateLimit` allows Gateway owners to
configure a dedicated RLS using built-in workload defaults. It does not permit RLS workload customization.
When `enableDedicatedRateLimitWorkloadOverrides` is false, `rateLimitDeployment`, `rateLimitHpa` and
`rateLimitPDB` on a Gateway-level `EnvoyProxy` are ignored with a warning naming the setting, and the
dedicated RLS runs with built-in workload defaults. Rate limiting keeps working. Fields on a
GatewayClass or without `rateLimit` retain the scope behavior described below.

Admins may additionally enable `enableDedicatedRateLimitWorkloadOverrides` to accept those workload
fields. It only takes effect together with `enableDedicatedRateLimit: true`. While
`enableDedicatedRateLimit` is false, `enableDedicatedRateLimitWorkloadOverrides` has no effect whatever
its value, and Envoy Gateway logs a warning at startup when it is true. This lets admins disable
dedicated RLS by changing only `enableDedicatedRateLimit`, as described above. The RBAC or admission
controls described above remain optional operator-owned mitigations when customization is enabled.
Disabling customization rolls existing dedicated RLS back to built-in workload defaults and reports the
ignored fields as warnings; Gateways are not rejected.

`rateLimit` is not accepted in the control-plane default `EnvoyGateway.envoyProxy`. Control-plane-wide
rate limiting stays configured in one place, `EnvoyGateway.rateLimit`.

### Example

This example customizes replicas and therefore requires both
`extensionApis.enableDedicatedRateLimit: true` and
`extensionApis.enableDedicatedRateLimitWorkloadOverrides: true`
(see [Enabling dedicated RLS](#enabling-dedicated-rls)).
Omit the `provider` block to use built-in workload defaults with only the first switch enabled.

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

A dedicated RLS always serves exactly one Gateway. Which RLS a Gateway uses is decided as follows:

| Configuration                                              | Result                                                   |
| ---------------------------------------------------------- | -------------------------------------------------------- |
| The Gateway's own `EnvoyProxy` sets `rateLimit`            | A dedicated RLS for that Gateway.                        |
| Only the GatewayClass's `EnvoyProxy` sets `rateLimit`      | Ignored; one of the last two rows applies.               |
| `mergeGateways` enabled                                    | No Gateway-level `EnvoyProxy`; one of the last two rows. |
| No Gateway-level `rateLimit`, `EnvoyGateway.rateLimit` set | The shared control-plane RLS (today's behaviour).        |
| No Gateway-level `rateLimit`, no shared RLS                | Global rate limits are rejected (today's behaviour).     |

### Shared rules

Ordinary rules keep separate counters through their Gateway listener domains, as today.

Rules with `shared: true` share one bucket among the policy's targets **that use the same RLS**. Each
dedicated RLS keeps its own bucket for the rule, and the shared control-plane RLS keeps one bucket for
all targets on it. For example, a shared 100rps rule whose targets are on Gateway A (dedicated RLS),
Gateway B (dedicated RLS) and Gateways C and D (shared RLS) allows up to 100rps on A, 100rps on B and
100rps across C and D together.

Domains are generated exactly as today, on every RLS; proxies and RLS configuration see no change.
Buckets are kept apart in Redis instead. The RLS supports a `CACHE_KEY_PREFIX` setting that it prepends
to every Redis key it reads or writes. Envoy Gateway sets it on each dedicated RLS Deployment to
`<gatewayNamespace>/<gatewayName>:` and leaves it unset on the shared RLS, as today. All keys of a
dedicated RLS, for shared and ordinary rules alike, are therefore separate from those of every other RLS,
even when several RLS use the same Redis database. Whether targets share a bucket depends only on which
RLS their Gateways use, never on the Redis backends behind them.

The prefix cannot collide with existing keys. The RLS builds each key from the prefix, the domain, `_`
and the descriptors. Domains consist only of Kubernetes names, listener section names and `/`, so
without a prefix a shared RLS key contains only lowercase letters, digits, `-`, `.` and `/` before its
first `_`. A dedicated RLS key always contains `:` before that point, so it never matches a shared RLS
key. Two dedicated RLS keys differ in the text before their first `:`, which is their Gateway's
namespace and name.

Admins can already set a custom `CACHE_KEY_PREFIX` on the shared RLS through its workload settings, for
example `eg:rl:` as shown in the global rate limit task. Envoy Gateway does not check it; like any
workload override, it is the admin's responsibility (see [Trust boundary](#trust-boundary)). A custom
shared prefix must not have the `<namespace>/<name>:` form that dedicated RLS use, or shared and
dedicated counters in the same Redis database can merge. Leaving the shared prefix unset, or giving
dedicated RLS a different Redis database, avoids the question.

The prefix is derived from the Gateway, not from the RLS settings. It stays the same when the RLS is
updated, including when its Redis backend changes.

Users who need one bucket across Gateways should keep those Gateways on the shared RLS. A shared bucket
never spans more than one RLS.

Whenever a shared rule's effective targets use more than one RLS, the policy reports
`SharedRateLimitCrossScope=True` with reason `MultipleRLSScopes`. The message lists the RLS involved
and states that each enforces the limit separately. The policy remains accepted, and the condition is
removed when no shared rule's targets span more than one RLS. The user guide must explain this
behaviour.

### GatewayClass EnvoyProxy

`rateLimit` and the RLS workload fields (`rateLimitDeployment`, `rateLimitHpa`, `rateLimitPDB`) are
ignored on an `EnvoyProxy` referenced by a GatewayClass's `spec.parametersRef`. Gateways in the class
are not affected by them. The GatewayClass stays `Accepted=True`, and its condition message notes that
these fields are ignored, so the configuration is not silently dropped.

These fields are also never inherited from the class through `mergeType`. With StrategicMerge or
JSONMerge, Envoy Gateway resolves `rateLimit` and the RLS workload fields from the Gateway's own,
unmerged `EnvoyProxy`; class values for them are discarded before merging. Normal merge rules still
apply to unrelated proxy fields. A Gateway therefore gets a dedicated RLS only when its own `EnvoyProxy`
sets a complete `rateLimit` block, including `backend`.

### Merged Gateways

With `mergeGateways` enabled, a Gateway's `spec.infrastructure.parametersRef` must be unset, as required
today, and the class `rateLimit` block is ignored as above. Merged Gateways therefore keep using the
shared RLS. This design does not change that attachment restriction.

### Workload settings

A dedicated RLS uses the `rateLimitDeployment`, `rateLimitHpa` and `rateLimitPDB` of the Gateway's
`EnvoyProxy`, then built-in defaults for omitted settings. Neither the GatewayClass `EnvoyProxy`,
`EnvoyGateway.envoyProxy` nor the shared RLS settings under `EnvoyGateway.provider.kubernetes` supply
defaults for dedicated RLS workloads. RLS workload fields in `EnvoyGateway.envoyProxy` are rejected when
loading the control-plane config. RLS workload fields on a Gateway `EnvoyProxy` without `rateLimit` have
no effect and are reported as a warning on the `EnvoyProxy` status. The shared RLS continues to use its existing
control-plane workload settings.

The exception is where the RLS image comes from. Private-registry and air-gapped installations set the
RLS image, pull secrets and pull policy in `EnvoyGateway.provider.kubernetes.rateLimitDeployment`, for
example through the Helm values `global.imageRegistry`, `global.images.ratelimit` and
`global.imagePullSecrets`. Without them, a built-in dedicated RLS would pull the compiled default image
and fail with `ImagePullBackOff`, with no admin-level fix while workload overrides are disabled.
The built-in defaults for a dedicated RLS therefore take these settings from the shared RLS
configuration:

* the container image;
* the pod's `imagePullSecrets`;
* the container's `imagePullPolicy`, including when set through a patch as the Helm chart does.

These are admin-authored and apply to every dedicated RLS whether or not workload overrides are
enabled. No other shared RLS setting is inherited. `EnvoyGateway.rateLimit` does not need to be set for
these defaults to apply, so dedicated-only installations can use them too. When workload overrides
are enabled, a Gateway's `EnvoyProxy` can still override them, subject to the
[security note](#api).

Pull secrets are referenced by name from the namespace where the RLS runs (see
[Namespace](#namespace)). In the default mode that is the controller namespace, as for the shared RLS
today. In Gateway namespace mode it is the Gateway's namespace: Envoy Gateway does not copy pull
secrets, so admins who set them must make them available in every Gateway namespace that uses a
dedicated RLS, as they already must for proxies that pull from a private registry.

## How it works

1. **Resolution.** When translating each Gateway, Envoy Gateway works out which RLS it uses (dedicated or
   shared) from the scope rules above. Global rate limit policies are accepted for a Gateway when it has
   either kind of RLS.
2. **Deployment.** Envoy Gateway deploys one RLS per Gateway with a dedicated RLS, named after that
   Gateway (for example `envoy-ratelimit-<hash>`), with a cache key prefix derived from the Gateway (see
   [Shared rules](#shared-rules)). It removes the RLS when the Gateway is deleted or no longer sets
   `rateLimit`, through ordinary reconciliation as described below.
3. **Rate limit configuration.** Each dedicated RLS receives only the rate limit rules of its Gateway.
   The shared RLS keeps receiving the rules of all Gateways without a dedicated RLS. A shared policy
   domain is included in the configuration of every RLS that serves at least one of the policy's
   targets. The configuration server runs whenever any RLS exists, even when no shared RLS is
   configured.
4. **Proxy configuration.** Each Gateway's proxies are pointed at their RLS, with that RLS's `timeout`,
   `failClosed` and connection settings.

```mermaid
flowchart LR
  EG["Envoy Gateway<br/>resolves RLS · deploys · pushes rules"]

  subgraph nsa["Namespace team-a"]
    A["Gateway A proxies<br/>(own RLS)"]
    RA["RLS-A<br/>prefix team-a/gw-a:"]
  end

  subgraph nsb["Namespace team-b"]
    B["Gateway B proxies<br/>(own RLS)"]
    RB["RLS-B<br/>prefix team-b/gw-b:"]
  end

  subgraph nsc["Namespace team-c"]
    CD["Gateways C, D proxies<br/>(no rateLimit)"]
  end

  subgraph cp["Controller namespace"]
    RS["shared RLS<br/>no prefix"]
  end

  RedisA[("Redis-A")]
  RedisS[("shared Redis")]

  A --> RA --> RedisA
  B --> RB --> RedisS
  CD --> RS --> RedisS

  EG -. rules .-> RA
  EG -. rules .-> RB
  EG -. rules .-> RS
```

Solid arrows are rate limit checks; dotted arrows are rules pushed by Envoy Gateway. RLS-B and the
shared RLS use the same Redis database but keep separate counters through their cache key prefixes.
The diagram shows Gateway namespace mode, where each dedicated RLS runs next to its Gateway's proxies.
In the default mode, proxies and all RLS run in the controller namespace.

### Service identity

Each RLS has a stable identity: the Gateway (namespace/name) for a dedicated RLS, or the existing shared
identity. Resource names and the configuration each RLS receives are keyed by this identity, so several
Gateways referencing the same `EnvoyProxy` each get their own RLS rather than sharing one.

The identity depends only on the Gateway, not on settings. Changing the settings of an existing RLS, including
its Redis backend, therefore updates that RLS in place rather than creating a new one; see
[Changing the configuration](#changing-the-configuration).

For dedicated RLS, explicit `rateLimitDeployment.name`, `rateLimitHpa.name` and `rateLimitPDB.name` are
not supported; they are ignored and reported as a warning on the `EnvoyProxy`. Envoy Gateway generates the names of the Deployment,
Service, HPA, PDB, certificate Secret and copied credential Secrets from the scope identity. Generated identities reserve the
existing shared RLS identity and include the scope kind, leaving room for a GatewayClass scope later. The
HPA target and Service and PDB selectors refer only to that scope's Deployment and pods; no selector
spans dedicated scopes or the shared RLS.

Today the shared RLS pods also depend on two support resources with fixed names in the controller
namespace: the `envoy-ratelimit` ServiceAccount and, with default telemetry, the
`statsd-exporter-config` ConfigMap. They exist only while the shared RLS is configured: Envoy Gateway
never creates them, or deletes them, when `EnvoyGateway.rateLimit` is not set. Dedicated RLS pods
therefore cannot use them, and must not depend on resources whose lifecycle belongs to another RLS, so
each dedicated RLS gets its own ServiceAccount and, when needed,
its own statsd exporter ConfigMap, with names generated from the scope identity like its other
resources. They are created and deleted with that dedicated RLS. Deleting the shared RLS, or another
dedicated RLS, therefore never removes resources a dedicated RLS uses, including in dedicated-only
installations without `EnvoyGateway.rateLimit`. The shared RLS keeps its existing fixed-name resources.
For the same reason, a dedicated RLS does not use the `envoy-rate-limit` TLS Secret, which may not
exist when only dedicated RLS are configured. Each dedicated RLS gets its own certificate Secret,
named from the scope identity (see [RLS identities and certificates](#rls-identities-and-certificates)).

Every dedicated RLS resource, and its pods, carries the following labels:

| Label                                            | Value                                             |
| ------------------------------------------------ | ------------------------------------------------- |
| `app.kubernetes.io/name`                         | The generated RLS name, never `envoy-ratelimit`   |
| `app.kubernetes.io/component`                    | `ratelimit`                                       |
| `app.kubernetes.io/managed-by`                   | `envoy-gateway`                                   |
| `gateway.envoyproxy.io/ratelimit-scope`          | `gateway`                                         |
| `gateway.envoyproxy.io/owning-gateway-namespace` | The Gateway's namespace                           |
| `gateway.envoyproxy.io/owning-gateway-name`      | The Gateway's name                                |

The shared RLS Service, Deployment and PDB select `app.kubernetes.io/name: envoy-ratelimit`,
`app.kubernetes.io/component: ratelimit` and `app.kubernetes.io/managed-by: envoy-gateway`. Adding the
scope labels alone would not exclude dedicated pods from that selector, so dedicated pods use a
different `app.kubernetes.io/name` value instead. The shared selectors therefore never match a
dedicated pod, and they stay unchanged, which matters because a Deployment's selector is immutable.
Each dedicated RLS selects all six labels above, so its selectors match neither the shared RLS nor any
other dedicated RLS.

The owning-gateway labels are the same ones Envoy Gateway already sets on proxy resources. The shared
RLS does not carry them or the `ratelimit-scope` label. A future GatewayClass scope would use a different
`ratelimit-scope` value. The labels are part of the supported interface, so admins can select dedicated
RLS pods in NetworkPolicies and admission policies (see
[Operator responsibility for tenant isolation](#operator-responsibility-for-tenant-isolation)).

#### Trust boundary

Envoy Gateway generates everything a dedicated RLS needs: resource names, labels and selectors, the
credential version annotation, environment variables such as the cache key prefix and the Redis and TLS
settings, and volumes and mounts. With built-in workload settings, Gateway owners cannot change any of
it, and this design describes exactly what runs.

Workload overrides and patches are applied on top without validation, as they are for the shared RLS
today. Changing generated values through them is unsupported: it can break isolation between RLS,
credential rotation or readiness reporting, and the behavior in this design then no longer applies.
The same is true of additions such as extra environment variables, `envFrom` sources, custom images,
other Deployment strategies or extra containers. This is why overrides need the separate, default-off
`enableDedicatedRateLimitWorkloadOverrides` switch, and why admins who enable it own the resulting
risk (see the [security note](#api)).

### Namespace

A dedicated RLS is part of its Gateway's data path, so it runs in the same namespace as that
Gateway's proxies:

| Deployment mode                 | Proxies run in           | Dedicated RLS runs in    |
| ------------------------------- | ------------------------ | ------------------------ |
| Default (`ControllerNamespace`) | The controller namespace | The controller namespace |
| `GatewayNamespace`              | The Gateway's namespace  | The Gateway's namespace  |

The shared RLS stays in the controller namespace in both modes. All resources of a dedicated RLS
(Deployment, Service, HPA, PDB, ServiceAccount, statsd ConfigMap and Secrets) live in its namespace
and, in Gateway namespace mode, count against that namespace's quota rather than the controller's.
Resource ownership and garbage collection are described in [Cleanup](#cleanup).

Following the proxies means a dedicated RLS adds no new kind of exposure in either mode. In the default
mode, Gateway owners can already configure pods in the controller namespace through proxy settings. In
Gateway namespace mode, the RLS stays inside the namespace whose proxies the Gateway owner already
controls, and the RLS CA Secret never shares a namespace with a pod that tenants configure.

**Secret write access for the controller.** Today the controller writes no Secrets in any namespace;
the certgen job creates the control plane and shared RLS certificates. Gateway namespace mode in
particular was built so the controller never writes Secrets in Gateway namespaces: proxies receive the
xDS CA through a ConfigMap and authenticate to xDS with a projected ServiceAccount token. A dedicated
RLS cannot follow that pattern:

* It serves mTLS to its proxies, so it needs its certificate and private key in its own namespace.
  This mTLS, like the TLS between the RLS and the configuration server, is always on, as it is for the
  shared RLS today; no setting turns it off, because caller isolation depends on it. Only TLS to Redis
  is optional. Every dedicated RLS therefore has a certificate Secret, whatever its Redis settings.
* The `envoyproxy/ratelimit` configuration client authenticates only with a client certificate or
  static headers; it cannot reload a short-lived projected ServiceAccount token. It therefore uses a
  configuration client certificate stored next to its serving certificate, as described in
  [RLS identities and certificates](#rls-identities-and-certificates).

The controller therefore needs permission to create, update and delete Secrets in the namespace where
each dedicated RLS runs:

* **Default mode:** in the controller namespace, for RLS certificate Secrets and Redis credential
  copies. The Helm chart adds it to the infra-manager Role in that namespace.
* **Gateway namespace mode:** in Gateway namespaces, for RLS certificate Secrets. The Helm chart adds
  it to whichever role already grants infrastructure access there: the cluster infra-manager
  ClusterRole when the controller watches all namespaces or uses a namespace selector, or each
  per-namespace infra-manager Role when `provider.kubernetes.watch.namespaces` lists namespaces.

**A separate Helm knob grants this permission.** The chart adds it only when
`dedicatedRateLimit.rbac.enabled` is true (default `false`), independently of
`enableDedicatedRateLimit`:

```yaml
dedicatedRateLimit:
  rbac:
    enabled: true   # defaults to false
```

The RBAC cannot follow `enableDedicatedRateLimit` itself, because that switch lives in the
`EnvoyGateway` configuration, which admins can change and reload without a Helm upgrade. RBAC derived
from it would drift from the running setting: turning the switch on outside Helm would leave every
dedicated RLS unable to start. With a separate knob, installations that never use dedicated RLS keep
today's least-privilege controller, and admins who plan to offer it enable the knob once and can then
turn `enableDedicatedRateLimit` on and off at runtime.

Enabling the knob is a standing increase in the controller's privileges: it can create, modify and
delete Secrets wherever it manages infrastructure, not only read them as today. The added risk is
limited to a compromised or faulty controller, which can already read those Secrets and create pods
that mount them. Envoy Gateway writes only the Secrets it owns for a dedicated RLS, selected by the
labels in [Service identity](#service-identity), and never modifies Secrets it did not create.
Installations that manage RBAC themselves grant the same permissions before enabling dedicated RLS.

If `enableDedicatedRateLimit` is true but the permission is missing, Envoy Gateway logs a warning at
startup, and each dedicated RLS reports `RateLimitReady=False` with reason `ProvisioningFailed` and a
message naming the missing permission. Turning the knob off while dedicated RLS exist does not block
their removal (see [Cleanup](#cleanup)).

#### Cleanup

Envoy Gateway removes a dedicated RLS the way it removes proxy resources today: when it observes the
Gateway being deleted or no longer setting `rateLimit`, it deletes that Gateway's RLS resources,
selected by that Gateway's labels. Removal does not depend on the controller holding Secret
permissions:

* **Secrets are owned by the RLS Deployment.** Every Secret of a dedicated RLS (its certificate Secret
  and, in the default mode, its credential copies) has a single owner reference, to that RLS's
  Deployment in the same namespace. Deleting the Deployment lets Kubernetes garbage-collect the
  Secrets, so removal needs only the infrastructure permissions the controller always has, and still
  works where an admin has removed the Secret permission. In Gateway namespace mode, all other RLS
  resources are owned by the Gateway, so deleting the Gateway also removes the Deployment and its Secrets.

**Changes Envoy Gateway does not observe leave resources behind.** The first iteration has no sweep
of stale resources. A configuration reload starts a new infrastructure runner without the previous
desired state, and changes made while Envoy Gateway is stopped are never seen. Dedicated RLS resources
therefore stay behind when:

* `enableDedicatedRateLimit` is turned off;
* a Gateway drops `rateLimit` while Envoy Gateway is stopped;
* in the default mode, a Gateway is deleted while Envoy Gateway is stopped. In Gateway namespace mode,
  the owner reference removes them;
* the deployment mode is switched and the RLS namespace changes. The reloaded control plane provisions
  the new placement without deleting the old one. Existing Gateways keep the old resources' owner
  references live, so garbage collection does not remove them.

Leftover RLS keep consuming resources and holding Redis credentials. For each Gateway, verify that
its proxies no longer use the old RLS and that the selected resources belong to this installation and
are no longer desired. Then delete that Gateway's stale placement; deleting the Deployment also
removes its Secrets:

```shell
kubectl delete deployment,service,hpa,pdb,serviceaccount,configmap -n <old-rls-namespace> \
    -l 'gateway.envoyproxy.io/ratelimit-scope=gateway,gateway.envoyproxy.io/owning-gateway-namespace=<gateway-namespace>,gateway.envoyproxy.io/owning-gateway-name=<gateway-name>'
```

Use this procedure for feature disablement, deployment mode changes and individual Gateway cleanup.
After a mode change, only the placement outside the Gateway's current RLS namespace is stale.
Gateways in the controller namespace keep the same placement in either mode; do not delete their RLS.
Always specify the old RLS namespace and both Gateway labels; do not use `-A`.


Because label-selected deletion relies on each Gateway's namespace and name being unique to one
controller, multiple Envoy Gateway installations managing Gateways in the same namespace are not
supported with dedicated RLS (see [Non-Goals](#non-goals)). An automatic sweep is deferred (see
[TODO](#sweeping-stale-dedicated-rls-resources)).

**Configuration server address.** The shared RLS reaches the configuration server through the short
Service name `envoy-gateway:18001`, which only resolves from the controller namespace. Every dedicated
RLS instead uses the fully qualified name `envoy-gateway.<controller-namespace>.svc.<dns-domain>:18001`,
in both modes, using the configured `DNSDomain`. That name is already a SAN of the control plane
certificate that certgen issues, so the RLS can verify the server against it. A wrong address fails
quietly. With `FORCE_START_WITHOUT_INITIAL_CONFIG`, which Envoy Gateway sets so the RLS does not block
on the configuration server at startup, the RLS serves without any configuration and still reports
healthy. Every check then returns an error, which proxies treat according to `failClosed`: requests
pass unlimited by default, or are all denied with `failClosed: true`. The rendered address is therefore
covered by the namespace placement tests in the [acceptance plan](#acceptance-and-validation-plan).

In Gateway namespace mode, the configuration server also receives connections from Gateway namespaces,
not only from the controller namespace. Every request is authenticated and limited to the caller's own
snapshot (see [RLS identities and certificates](#rls-identities-and-certificates)), so this exposes no
configuration beyond what the Gateway owner already has.

### RLS identities and certificates

The shared RLS keeps today's certificates: it serves with the `envoy-rate-limit` Secret and accepts the
client certificate that every proxy presents. Dedicated RLS use a separate certificate authority, the
**RLS CA**, and a separate identity per Gateway.

**The RLS CA.** certgen creates the RLS CA together with the existing control plane certificates and
stores its certificate and key in the `envoy-rate-limit-ca` Secret in the controller namespace. The Secret has type
`kubernetes.io/tls` and requires `tls.crt` containing a PEM-encoded CA certificate and `tls.key`
containing its matching PEM-encoded private key. The certificate must have CA basic constraints and
certificate-signing key usage. V1 uses a single self-signed root CA; intermediate CA chains are outside
scope. Custom installations supply the same Secret format. Only the controller reads it; no pod mounts
it. It is separate from the control plane CA for two reasons:

* The existing Secrets contain only certificates signed by the control plane CA, not the CA private key,
  which certgen discards. Their keys cannot sign new certificates.
* Storing the control plane CA key in the cluster would let anyone who reads it sign certificates that
  the xDS server and every proxy trust. A leaked RLS CA key only affects dedicated RLS identities.

certgen creates Secrets that do not exist yet and leaves existing ones alone, so upgrading an
installation adds the RLS CA without touching the existing certificates. Installations that bring their
own certificates must also provide this Secret before enabling dedicated RLS.

**Gateway identity.** Each dedicated RLS uses its Service FQDN as its Gateway's certificate identity:
`<rls-service>.<rls-namespace>.svc.<dns-domain>`, where `<rls-namespace>` is the namespace the RLS
runs in (see [Namespace](#namespace)). The Service name is derived from the Gateway's namespace/name,
so the FQDN is unique in both deployment modes. `<dns-domain>` uses the configured `DNSDomain`, as
control plane certificate generation does today. Proxies verify this DNS SAN in the RLS serving
certificate; the RLS requires it in proxy client certificates through `GRPC_CLIENT_TLS_SAN`.

Configuration clients use a separate URI SAN derived from the Gateway's namespace/name:
`urn:envoy-gateway:ratelimit:<gateway-namespace>:<gateway-name>`. The configuration server maps this
identity to the Gateway's scope and authorizes requests for that scope only. The RLS CA never issues
wildcard SANs.

Resource names, Redis key prefixes and certificate identities remain stable when a Gateway is
recreated with the same namespace and name in the same installation and deployment mode.

Reusing a Gateway name assumes the same trust ownership. Certificates from the previous Gateway remain
accepted by the recreated RLS until expiry, provided they are signed by the current RLS CA; a former
owner retaining a client certificate can still call it. Deletion does not revoke certificates or create
a fresh security boundary. Operators accept this risk when reusing names, or choose a different Gateway
name when ownership changes. An existing valid certificate may be reused if it is still available and
matches the CA, identity and required SANs. This does not retain certificate Secrets beyond the normal
RLS cleanup lifecycle or persist proxy client certificates: if cleanup removed them, reconciliation
issues new certificates with the same identity.

**Certificates the controller issues.** For each dedicated RLS, the controller signs three certificates
with the RLS CA:

| Certificate | SANs | Extended key usage | Delivery and storage |
| --- | --- | --- | --- |
| Proxy client | Gateway identity FQDN as its only DNS SAN | `clientAuth` | Controller memory; delivered only to that Gateway's proxies over SDS |
| RLS serving | RLS Service DNS names, including the identity FQDN | `serverAuth` | RLS certificate Secret |
| RLS configuration client | Scope URI above as its only SAN; no DNS SAN | `clientAuth` | RLS certificate Secret |

Each controller replica issues its own proxy client certificate. All are accepted because the RLS
checks the CA and SAN, not a specific certificate. A controller restart issues a new certificate and
delivers it over SDS without restarting proxies or changing their pod specification.

Both RLS certificates are stored in the RLS's own certificate Secret in the RLS's namespace, together
with the RLS CA certificate and the control plane CA certificate, and are written by the controller
that provisions infrastructure. Reconciliation reuses existing valid certificates matching the current
CA, Gateway identity and Service names; it does not periodically renew certificates. Restarts and
infrastructure-leader changes do not themselves reissue them or change the certificate version
annotation.

The serving and configuration client roles use separate certificates so that nothing in the RLS Secret
is accepted as a caller by any dedicated RLS. A dedicated RLS accepts a client only with the
`clientAuth` usage and a DNS SAN equal to its `GRPC_CLIENT_TLS_SAN`. The serving certificate has the
right DNS SAN but no `clientAuth`; the configuration client certificate has `clientAuth` but no DNS
SAN. In Gateway namespace mode this matters when Gateways share a namespace: someone who can read
Gateway B's RLS Secret there still cannot call B's RLS. A single certificate with both usages would
carry B's SAN and `clientAuth` and be accepted.

**What each side checks.**

* The **dedicated RLS** serves with its RLS serving certificate, trusts only the RLS CA for clients,
  and sets `GRPC_CLIENT_TLS_SAN` to its Gateway identity. It thereby requires the `clientAuth` usage
  and that DNS SAN from every caller.
* **Proxies** verify the dedicated RLS against the RLS CA, delivered over xDS rather than read from a
  file, and require its Service DNS name as the SAN. A proxy cannot be pointed at another RLS that
  holds a valid RLS CA certificate.
* The **configuration server** accepts client certificates from both the control plane CA (the shared
  RLS) and the RLS CA (dedicated RLS). For a client with an RLS CA certificate, it accepts only a
  configuration client certificate, with `clientAuth` and the URI SAN above, and serves only the
  snapshot whose node ID maps to the scope identity in that URI SAN, so a dedicated RLS cannot fetch
  another Gateway's rules. Clients authenticated by the control plane CA may request only the shared
  snapshot; they cannot access dedicated snapshots. This authorization applies to every configuration
  request, including requests on an existing stream. The dedicated RLS presents its configuration
  client certificate and verifies the configuration server against the control plane CA, at the
  address in [Namespace](#namespace).

**The configuration server follows the RLS CA Secret.** The server reloads RLS CA trust from the same
watched `envoy-rate-limit-ca` Secret used for certificate issuance. Creating, fixing or replacing the
Secret takes effect for new TLS handshakes without a restart; existing connections retain their
verified identity until they reconnect. With no valid RLS CA, the server rejects new dedicated RLS TLS handshakes.
Its control plane certificates and shared RLS authentication remain unchanged.

```mermaid
sequenceDiagram
    participant CG as certgen
    participant EG as Envoy Gateway
    participant PA as Gateway A proxies
    participant RA as Gateway A RLS
    participant RB as Gateway B RLS

    CG->>EG: RLS CA Secret (certificate and key)
    EG->>RA: RLS Secret (serving and config client certificates), GRPC_CLIENT_TLS_SAN = identity A
    EG->>PA: client certificate (SAN identity A) and RLS CA over SDS
    PA->>RA: mTLS, SAN matches, accepted
    PA-->>RB: mTLS, SAN mismatch, rejected
    RA->>EG: config request, node ID matches URI SAN, rules for A only
```

**Certificate lifetime.** The RLS CA, proxy client certificates and RLS certificates use the existing
certgen lifetime of 1,825 days (five years). Every issued leaf certificate's expiry is capped at its
CA's expiry; custom CAs may have a shorter remaining lifetime. V1 does not automatically renew
certificates. Provisioning issues missing certificates when a valid CA is available and reuses valid
matching certificates. Expiry alone does not trigger replacement: an expired persisted RLS serving
or configuration client certificate remains an operator-managed failure. Reissuing proxy client
certificates on controller restart does not extend the CA's lifetime.

Operators are responsible for monitoring certificate expiry and arranging replacement. An expired or
not-yet-valid CA reports `CertificateUnavailable`; expired certificates clear readiness, and TLS calls
that cannot be served follow `failClosed`. Deleting a Gateway does not revoke its certificates.
Expired or compromised certificate material requires operator-managed recovery, which may disrupt
service. V1 provides no coordinated replacement workflow. Automatic renewal for both shared and
dedicated RLS is deferred (see [RLS certificate renewal](#rls-certificate-renewal)).

When reconciliation writes a new RLS certificate Secret, it updates the certificate version annotation
on the pod template only after the Secret write succeeds, so the workload rolls to use that version.

**Missing RLS CA.** If the RLS CA Secret is missing or invalid, the controller cannot issue
certificates. Dedicated RLS that need a certificate report `RateLimitReady=False` with reason
`CertificateUnavailable`, and their proxies keep any certificate they already have until it expires.
This failure does not change RLS selection, fall back to the shared RLS, or skip infrastructure, rule
or proxy reconciliation. On first creation without a usable CA, the Service and Deployment are still
provisioned, but RLS pods cannot start without their certificate Secret. The proxy configuration keeps
the dedicated RLS cluster and rate limit filter; without a client certificate, the cluster cannot make
authenticated calls. Unavailable calls follow `failClosed`. Once the CA becomes available, ordinary
reconciliation issues the certificates and delivers them to the RLS and proxies, and the configuration
server trusts the new CA without a restart. The shared RLS is not affected.

### Isolation between Gateways

A dedicated RLS accepts only its own Gateway's proxies. Other Gateways' proxies present a certificate
with a different SAN, and the shared client certificate is signed by a different CA, so both are
rejected during the TLS handshake. This holds in every deployment mode, including the default mode where
all proxies share the controller namespace, and for Gateways that share a namespace. Pointing a proxy at
another Gateway's RLS, for example through `EnvoyPatchPolicy`, does not help, because a proxy only ever
receives its own Gateway's client certificate. NetworkPolicy ingress rules are optional defense in
depth.

This isolation relies on the RLS CA private key and on the certificates delivered to proxies staying
confidential:

* Anyone who can read the RLS CA Secret can impersonate any Gateway. The Secret stays in the controller
  namespace in both modes and no RLS pod mounts it. Reading Secrets in the controller namespace, and,
  in the default mode, workload overrides that mount them, must be limited to administrators (see
  [Operator responsibility for tenant isolation](#operator-responsibility-for-tenant-isolation)).
* A Gateway owner can read their own proxies' client certificate, but that only lets them call their
  own Gateway's RLS.
* In Gateway namespace mode, anyone who can read Secrets in a Gateway's namespace can read the RLS
  Secrets of every Gateway there. Neither certificate in them is accepted as a caller by any dedicated
  RLS (see [RLS identities and certificates](#rls-identities-and-certificates)), so caller isolation
  between Gateways that share a namespace still holds. What remains is limited: the configuration
  client certificate lets the reader fetch that Gateway's rate limit rules, and the serving key lets
  them pose as that Gateway's RLS, but only to proxies whose traffic they can already redirect, which
  needs write access to that Gateway's resources in the namespace. Operators who place Gateways of
  different owners in one namespace should restrict Secret reads there.
* Every xDS client requests only its own Gateway's configuration. This is a precondition Envoy
  Gateway already relies on for listener and backend certificates (see [Assumptions](#assumptions));
  without it, a custom proxy image or extra container can obtain another Gateway's client certificate.

The shared RLS keeps today's behavior: it accepts every proxy and is not a security boundary between
Gateways. Egress from a dedicated RLS, for example to an internal address set as its Redis URL, is not
covered by certificates; it is subject to the network controls of the namespace the RLS runs in (see
the [security note](#api)).

### Redis credentials

Redis references are resolved against the Gateway's `EnvoyProxy` that supplied the `rateLimit` block.
In both deployment modes, `urlRef.secretKeyRef` and the TLS `certificateRef` must be in that
`EnvoyProxy`'s namespace, and a `ReferenceGrant` does not lift this limit. One rule for both modes
means a configuration stays valid when an admin switches deployment mode, and it can be relaxed later
without breaking anyone. Platform teams that keep one Redis client certificate for several Gateways
replicate it into each Gateway namespace, for example with an external secrets tool.

**Exception: controller-namespace TLS Secrets in the default mode.** In the default mode, the TLS
`certificateRef` may also name a Secret in the controller namespace, where the dedicated RLS runs. This
lets admins keep Redis client certificates out of tenant namespaces. It requires a `ReferenceGrant` in
the controller namespace that allows the `EnvoyProxy`'s namespace to reference that Secret. Without the
grant, any Gateway owner could use any Redis client certificate an admin keeps there, including the
shared RLS's, to authenticate to that Redis. The grant is the admin's consent, and its absence or
revocation is reported like any other failed reference. Secrets that Envoy Gateway itself owns or
requires in the controller namespace, such as the RLS CA, the control plane certificates and the
Secrets of any RLS, are never accepted, even with a grant. The exception does not apply to `urlRef`,
and it does not apply in Gateway namespace mode, where a Secret in the controller namespace could only
reach the RLS by copying it to where the Gateway owner can read it.

The controller must be able to see these grants and Secrets even when it watches only selected
namespaces. When `provider.kubernetes.watch.namespaces` omits the controller namespace, Gateway API
informers, including `ReferenceGrant`, watch only the listed namespaces today, and the Helm chart gives
the controller only Secret and EndpointSlice reads in its own namespace. While `enableDedicatedRateLimit`
is true, Envoy Gateway therefore also watches `ReferenceGrant` resources in the controller namespace,
as it already does for Secrets there. The Helm chart adds `get`, `list` and `watch` on
`referencegrants` to the infra-manager Role in the controller namespace when
`dedicatedRateLimit.rbac.enabled` is true, together with the Secret permission (see
[Namespace](#namespace)). Without this, a valid grant
would be invisible and the reference would report `InvalidCredentials`.

How the RLS reads them depends on where the Secret is relative to the RLS (see [Namespace](#namespace)):

* **Same namespace as the RLS.** The RLS reads the source Secret directly, selecting only the
  referenced keys. Nothing is copied. This covers every reference in Gateway namespace mode, and the
  controller-namespace TLS exception in the default mode.
* **`EnvoyProxy`'s namespace in the default mode.** The RLS runs in the controller namespace, so Envoy
  Gateway copies only the referenced keys into a Secret it owns for that RLS. It keeps the copy in sync
  with the source Secret and deletes it when the RLS is removed. The Gateway owner cannot read the copy,
  but it holds only credentials the Gateway owner supplied.

A reference outside these rules is reported as `RateLimitReady=False` with reason `InvalidCredentials`
(see [Validation and status](#validation-and-status)).

In both modes a credential version is set as an annotation on the RLS pod template, so a credential
rotation triggers a normal rolling update and new pods pick up the new credentials.

The RLS reads credentials once, at startup. When it reads the source Secret directly, the Secret is
already current when the annotation changes. When credentials are copied, the copied Secret must be
current before any new pod starts. Envoy Gateway therefore changes the copied Secret first and changes the pod template only after
that write has succeeded: on rotation, it updates the copy and then the annotation; on deletion or
revocation, it deletes the copy and then updates the template as described below. If the Secret write fails, the pod
template is left unchanged and the change is retried. A pod started for the new template thus never
reads a stale controller-owned copy. Direct mounts have the revocation limitation described below.

The annotation is never derived from the credential values themselves. A plain hash of a Redis URL
containing a human-chosen password would let anyone who can read Deployments in the RLS's namespace,
but not Secrets, test password guesses offline. Instead, it is a hash of non-secret version
data: a canonical, unambiguous encoding of each reference's purpose, resolved namespace and name,
selected key names, source Secret UID and `resourceVersion`, and authorization state. Reference and key
selection are included even when two references resolve to the same Secret version. Changing a URL key
or TLS reference therefore changes the pod template version even if the source Secret is unchanged.
Credential values are never included. Any change to a source Secret therefore rolls the RLS, including changes to keys
it does not use, which is acceptable for configuration that rarely changes (see
[Assumptions](#assumptions)).

**Missing or revoked credentials.** If a referenced Secret is deleted or its `ReferenceGrant` is revoked,
Envoy Gateway deletes its copy, if any, then updates the annotation. The Gateway reports
`RateLimitReady=False` with reason `InvalidCredentials` and a message naming the failed reference.
Envoy Gateway never modifies or deletes the administrator-owned source Secret.

If the source Secret or required copy is absent, new pods cannot start. With the built-in
`RollingUpdate` strategy, the rollout stalls and old pods keep serving with credentials they already
loaded. Once the references are valid again, reconciliation restores any required copy and the rollout
completes. A different strategy set through workload overrides, such as `Recreate`, can stop the old
pods first (see [Trust boundary](#trust-boundary)).

Revoking a grant for an existing direct mount only reports the authorization failure; it does not
block the workload from using that mount. The source Secret remains present, and Kubernetes does not
check ReferenceGrants when mounting Secrets, so existing and replacement pods may continue using it.
There is no stalled-rollout guarantee for this case. Reconciliation retains the existing workload's
direct reference, including after controller restart, but never adds a new denied reference to a pod
template. An initially denied reference reports `InvalidCredentials` and leaves the RLS without the
required credential; it does not gain access to the Secret. Restoring the grant clears the reference
failure through normal reconciliation.

Deleting the source Secret or revoking its ReferenceGrant is therefore **not** a reliable way to cut
off Redis access. Deletion also does not reliably keep the RLS running: the outcome depends on pod churn. Envoy Gateway does not
provide an RLS stop switch in the first iteration. Removing the `rateLimit` block isn't one either: the
Gateway moves to the shared RLS, if configured, and the dedicated RLS is removed only after that change
converges. To stop a dedicated RLS from using credentials that must no longer be used, revoke them at Redis
(change the password or ACL, or revoke the client certificate). The running pods then lose access on
their next connection attempt, and rate limit checks follow `failClosed`.

Credential resolution failures do not change RLS selection or remove its desired infrastructure or
rules. Envoy Gateway continues translating the Gateway's listeners and routes, including its RLS cluster
and rate limit filter, and never falls back to another RLS because credentials are unavailable.
This also applies on first creation with a missing Secret: the Service and Deployment are provisioned,
but pods requiring the missing credential cannot start. Checks that cannot be served follow
`failClosed`. After deletion or revocation, existing pods may continue serving, as described above.
On restart, reference validation uses current Secrets and grants; existing direct mounts are read from
the persisted workload. No in-memory snapshot is required.

### Validation and status

A `rateLimit` block without `backend` is rejected by schema validation when the `EnvoyProxy` is applied.
Other invalid settings, such as an unsupported `serviceSettings` field, are reported the same way as
other invalid `EnvoyProxy` settings. The `serviceSettings` restrictions are the same as for the shared
RLS. Settings not allowed by a disabled `extensionApis` switch are not invalid parameters: they are
ignored with an `EnvoyProxy` warning and never set the Gateway `Accepted=False` (see
[Enabling dedicated RLS](#enabling-dedicated-rls)).

Missing Secrets, missing or invalid credential keys, denied or revoked Secret references, and
references outside the namespace rules in [Redis credentials](#redis-credentials) are
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

* All credential references resolve successfully and any owned credential copies are up to date.
* The RLS certificate Secret holds current serving and configuration client certificates issued by
  the RLS CA, and a proxy client
  certificate for the Gateway is currently valid under that CA.
* Infrastructure reconciliation has succeeded. The Deployment has observed its latest generation,
  completed the desired rollout, and has at least one available replica, with no old replicas remaining.
  Its pod template includes the current credential and certificate versions.
* The Service has at least one ready endpoint belonging to the current rollout.

The controller recomputes readiness as desired workload settings, credentials, certificates, workloads
and endpoints change, and when certificates expire. It does not retain `True` from the previous desired workload configuration while a new rollout
is pending. On restart, readiness remains unknown until the controller has observed the required state.
This condition describes credential resolution and Kubernetes infrastructure readiness. It does not
confirm that RLS endpoints have loaded the desired rules, can connect to Redis, or provide uninterrupted
enforcement. ACK-based configuration readiness is deferred to [TODO](#todo). Operators must verify
actual enforcement separately before restoring traffic after planned maintenance.

`RateLimitReady=False` uses the following reasons, in priority order when multiple failures apply:

| Reason                   | Meaning                                                                                            |
| ------------------------ | -------------------------------------------------------------------------------------------------- |
| `InvalidCredentials`     | A required credential or its authorization cannot be resolved.                                     |
| `CertificateUnavailable` | The RLS CA is missing, invalid or expired, or a leaf certificate is invalid or expired.¹           |
| `ProvisioningFailed`     | Creation or update of a required Kubernetes resource failed, including the RLS certificate Secret. |
| `RolloutFailed`          | The Deployment reports that the desired rollout exceeded its progress deadline.                    |
| `Provisioning`           | Required resources have not yet been observed.                                                     |
| `RollingOut`             | The current Deployment rollout has not completed.                                                  |
| `NoReadyEndpoints`       | The Service has no ready endpoint from the current rollout.                                        |

¹ A leaf certificate that is missing only because writing its Secret failed, for example because the
controller lacks Secret write permission (see [Namespace](#namespace)), is reported as
`ProvisioningFailed`, not `CertificateUnavailable`. A missing leaf counts as `CertificateUnavailable`
only when the controller cannot issue it, because the RLS CA is unusable.

An unknown observation state uses `RateLimitReady=Unknown` with reason `PendingObservation`. Successful
convergence uses reason `Ready`. A Gateway that no longer uses global rate limiting has this condition
removed, as does a Gateway that switches to the shared RLS. Shared RLS status and credential validation
retain their existing control-plane behavior.

### Supported infrastructure

Dedicated RLS is supported only when Envoy Gateway provisions infrastructure itself with its built-in
Kubernetes infrastructure manager, that is, with `provider.type: Kubernetes`. That manager is the only
one that can create a per-Gateway RLS and report the Deployment and Service state that `RateLimitReady`
requires. Every other configuration is unsupported in the first iteration, including:

* the host provider;
* a custom provider with remote infrastructure: the remote contract cannot request a per-Gateway RLS,
  because `CreateOrUpdateRateLimitInfraRequest` carries no scope, backend, settings or address;
* a custom provider without infrastructure, where nothing provisions the RLS at all;
* any infrastructure type added in the future, until it is explicitly supported.

While `enableDedicatedRateLimit` is true, setting `rateLimit` on a Gateway's `EnvoyProxy` with
unsupported infrastructure makes the Gateway `Accepted=False`, with a message saying so, and no RLS
infrastructure request is sent. Supporting more infrastructure needs its own provisioning and readiness
contract.

While `enableDedicatedRateLimit` is false, this check does not run. `rateLimit` is ignored as for any
other provider, and the Gateway stays accepted, as described in
[Enabling dedicated RLS](#enabling-dedicated-rls). Rejecting it then would let a block the admin never
enabled take down all of the Gateway's listeners. Admins should therefore not enable dedicated RLS with
unsupported infrastructure; if they do, the Gateways that set `rateLimit` are rejected, which is visible and
limited to Gateways that asked for a dedicated RLS.

## Changing the configuration

Envoy Gateway reconciles the desired RLS selection, infrastructure, rules and proxy settings directly.
There is no migration record, ACK gate, retained source configuration or coordinated cutover in v1.
Controller restart reconstructs desired state from current resources.

Adding or removing `rateLimit`, or toggling `enableDedicatedRateLimit`, can switch a Gateway between
shared and dedicated RLS. Treat these as planned maintenance operations. During convergence:

* Calls to an unavailable RLS follow `failClosed`.
* A proxy may still call the previous RLS after its rules have moved. An unknown domain is allowed by
  the RLS, so those requests can bypass limits **even with `failClosed: true`**. Disconnected proxies
  or rejected updates can prolong this gap without a guaranteed upper bound.

Operators requiring continuous enforcement must stop or externally gate affected traffic throughout
the change and verify proxy configuration and actual rate-limit behavior before restoring it.
`RateLimitReady=True` alone is insufficient. Feature disablement requires the same planning; it is
not an emergency enforcement or credential-revocation mechanism.

Updating Redis connection settings rolls the existing RLS Deployment normally. Redis provisioning,
data transfer, cutover, rollback and counter continuity are outside this proposal. Old and new pods
may temporarily use different backends and enforce independent buckets. Credential updates use the
copy-and-rollout procedure above and do not change RLS selection.

**Counters after a change.** Moving a Gateway between the shared RLS and a dedicated one changes the
cache key prefix of its keys (see [Shared rules](#shared-rules)), selecting a different keyspace even
if both use the same Redis database. Counters start from zero only when matching, unexpired keys do not
already exist in the destination keyspace. Switching shared → dedicated → shared can reuse surviving
shared counters. Likewise, re-enabling the dedicated RLS for the same Gateway, or recreating a Gateway
with the same namespace and name, can reuse its earlier dedicated counters. A shared policy bucket on
the destination shared RLS may also already contain usage from other targets. The transition does not
reset counters of other targets of the same policy. These are RLS keyspace semantics, not a promise
to transfer or synchronize counters.

## Upgrade and compatibility

Without `EnvoyProxy.spec.rateLimit` on a Gateway's `EnvoyProxy`, nothing changes: same shared RLS, same
resource names, same behaviour. Dedicated RLS is disabled by default, so upgrading does not let anyone
create RLS workloads until an admin sets `extensionApis.enableDedicatedRateLimit`. The controller's
RBAC is unchanged too until an admin sets the Helm value `dedicatedRateLimit.rbac.enabled`, which
must be set before, or together with, the switch (see [Namespace](#namespace)). The certgen job of an upgraded installation creates the RLS CA
Secret and leaves existing certificates unchanged; installations with custom certificates must create
it before enabling dedicated RLS. `rateLimit` on a GatewayClass's `EnvoyProxy` is ignored in this
iteration, so adding GatewayClass support later only changes behaviour for configurations that are
explicitly reported as ignored today. The GatewayClass condition message makes this visible before
support lands.

## Alternatives considered

* **Support GatewayClass scope in the first iteration.** A class `rateLimit` block would deploy one RLS
  shared by every Gateway in the class, including merged Gateways. This reduces the number of RLS
  deployments, but requires deciding how class and Gateway blocks interact under each `mergeType`,
  tracking which level supplied the merged block, and handling the class RLS lifecycle as Gateways join
  and leave. Deferred; see [TODO](#gatewayclass-scope).
* **Secret RBAC tied to `enableDedicatedRateLimit`, or granted unconditionally.** Deriving the chart's
  RBAC from the runtime switch breaks when admins change the switch without a Helm upgrade. Granting
  it in every installation would raise the controller's privileges even where dedicated RLS is never
  used. A separate, default-off Helm knob avoids both.
* **Dedicated RLS in the controller namespace in every mode.** The controller would never write Secrets
  in Gateway namespaces and the RLS would reach Redis Secrets only through copies. But in Gateway
  namespace mode it would place tenant-configured pods next to the RLS CA Secret and the control plane,
  charge their resources to the controller namespace's quota, send tenant-chosen Redis and telemetry
  traffic from the controller namespace past the tenant's NetworkPolicies, and require admins to write
  per-Gateway NetworkPolicies there. Running the RLS next to its proxies keeps all of that inside the
  Gateway's namespace, at the cost of Secret write access in Gateway namespaces.
* **No switch; rely on RBAC and admission policy only.** Less configuration, but every upgraded control
  plane would immediately let anyone who can author an `EnvoyProxy` create RLS pods, protected only by
  guidance in the docs.
* **A runtime flag instead of an `extensionApis` field.** Runtime flags guard breaking changes and
  experimental behaviour, are often enabled by default and are removed once the change settles; they
  are not meant as permanent security switches. `EnvoyProxyPatch` is enabled by default for this
  reason. `extensionApis` already holds the permanent, disabled-by-default switches for risky features.
* **The switch under `EnvoyGateway.rateLimit`.** That block configures the shared RLS, and one goal is
  to allow dedicated RLS without a shared RLS, so users would have to set it only to reach the switch.
* **A single switch for dedicated RLS and workload customization.** Simpler API, but enabling per-Gateway
  Redis configuration would also allow RLS workload customization. Separate switches
  let admins offer built-in workloads without requiring workload admission policies.
* **Named rate limit profiles in the `EnvoyGateway` config**, referenced by name from `EnvoyProxy`. This
  keeps Redis credentials admin-only, but `EnvoyGateway` changes need a control plane restart, and it adds
  a second place to configure per-Gateway infrastructure.
* **A new `RateLimitService` resource** referenced from `EnvoyProxy` or `BackendTrafficPolicy`. This lets
  any set of Gateways share one dedicated RLS, but it is a new API. Per-Gateway isolation, the main
  need, is covered by the existing `EnvoyProxy` attachment; sharing across Gateways is left to future work.
* **Choosing the RLS per `BackendTrafficPolicy`.** A proxy would then talk to several RLS instances, which
  makes the proxy configuration much more complex and gives Gateways no isolation by default.
* **One ServiceAccount and statsd exporter ConfigMap shared by all RLS instances.** Fewer objects, but
  the shared resources would need their own owner and a rule for deleting them once no RLS uses them,
  independent of `EnvoyGateway.rateLimit`. Per-RLS copies are created and deleted with their RLS, so
  no RLS depends on another's lifecycle. The copies are cheap: the ServiceAccount has no RBAC
  bindings, and the ConfigMap content is the same static statsd mapping.
* **Network-only caller isolation.** Keep the shared client certificate and rely on NetworkPolicy to
  restrict each dedicated RLS to its Gateway's proxies. No certificate handling is added, but
  NetworkPolicy identifies callers by namespace only: it cannot separate Gateways that share a
  namespace, or any Gateways in the default mode where all proxies run in the controller namespace, and
  it fails open when a tenant's proxy escapes its namespace's network.
* **Sign per-Gateway certificates with the existing control plane CA**, or with the `envoy-rate-limit`
  Secret. The existing Secrets hold only certificates without the `CertSign` usage; certgen discards
  the CA private key. Storing that key would let its holder sign certificates trusted by the xDS server
  and every proxy, and putting it in `envoy-rate-limit` would expose it to the shared RLS pod. The
  `envoy-rate-limit` Secret may also not exist in dedicated-only installations.
* **Admin-supplied certificate per Gateway**, for example from cert-manager, in a Secret with an
  expected name. Envoy Gateway would never hold a CA key, but every Gateway would need a manual step
  before its RLS works, and the proxy client certificate would have to be stored in a Secret.
* **Proxy client certificates stored in Secrets** instead of controller memory. Certificates would
  stay stable across controller restarts and replicas, but every Gateway would get a stored private key
  in the controller namespace, and stored copies would need cleanup. Reissuing after a restart only
  makes proxies reconnect to their RLS.
* **Multiple control planes** (today's workaround). It works, but it is heavy when only the rate limit
  backend needs to be isolated.

## Proposed decisions

These are the author's recommendations and are open for review.

* Support Gateway scope on the built-in Kubernetes infrastructure, with each dedicated RLS in the same
  namespace as its Gateway's proxies. GatewayClass scope and merged Gateways remain on the shared RLS.
* Reuse `EnvoyProxy` for configuration, with separate default-off switches for dedicated RLS and
  workload overrides. Disabled settings are ignored with warnings, as defined in the
  [behavior table](#enabling-dedicated-rls).
* Resolve dedicated settings from the Gateway's unmerged `EnvoyProxy`. Use built-in workload defaults,
  inheriting only administrator-configured image, pull secrets and pull policy from shared RLS settings.
* Keep workload customization and tenant hardening operator-owned, without new admission hooks,
  safety validation or generated NetworkPolicies.
* Keep shared buckets within each RLS, isolate Redis keys by Gateway namespace/name, and warn when a
  shared rule spans RLS scopes.
* Authenticate each dedicated RLS scope using a separate RLS CA and its Service FQDN as identity.
  Namespace/name reuse preserves identity. Use five-year certificates capped by CA expiry, with
  operator-managed expiry and no automatic renewal or revocation in v1.
* Require Redis credential references to be in the `EnvoyProxy`'s namespace in both modes, even with
  a `ReferenceGrant`. In the default mode only, also allow a TLS `certificateRef` to a granted Secret
  in the controller namespace, never one Envoy Gateway owns. Read Secrets in the RLS's namespace
  directly and copy the rest into the controller namespace in the default mode. Roll workloads on
  credential changes. Credential revocation is enforced at Redis.
* Report infrastructure and credential readiness through `RateLimitReady`. Operational failures
  preserve RLS selection and rate limit filters; configuration ACK tracking is deferred.
* Grant the controller's Secret write and controller-namespace `ReferenceGrant` read permissions
  through a separate, default-off Helm value, `dedicatedRateLimit.rbac.enabled`, not through
  `enableDedicatedRateLimit`, which can change at runtime. Admins confirm the permissions after
  turning the feature on or off.
* Make each dedicated RLS's Secrets owned only by its Deployment, so removal never needs Secret
  permissions. Remove dedicated RLS on observed changes only; leave resources from unobserved changes,
  feature disablement and deployment mode switches to per-Gateway manual cleanup, and defer an
  automatic sweep. Have the configuration server follow the RLS CA Secret without a restart. Multiple installations
  managing one namespace are unsupported.
* Use ordinary reconciliation for scope changes and require planned maintenance when continuous
  enforcement is needed, including with `failClosed: true`.

## Acceptance and validation plan

Implementation must cover the following cases with meaningful automated tests:

* **Opt-ins and scope:** both switches default off; built-in workloads work with only dedicated RLS
  enabled and use the image, pull secrets and pull policy from the shared RLS settings, also without
  `EnvoyGateway.rateLimit`; workload overrides require both switches; the overrides switch alone has no effect and does
  not block startup, so disabling only `enableDedicatedRateLimit` works; class fields are ignored and
  never inherited.
* **Disabled switches:** Gateways stay accepted and serving with an `EnvoyProxy` warning; with dedicated
  RLS disabled, a Gateway with `rateLimit` (including `failClosed` false or unset) stays on the shared
  RLS and its shared limits keep being enforced, or its policies are rejected as today without a
  shared RLS; disabling the switch moves dedicated Gateways back to the shared RLS; with workload
  overrides disabled, the RLS runs with built-in defaults.
* **Coexistence:** shared and dedicated RLS operate together; dedicated-only operation works without a
  shared RLS; multiple Gateways referencing one `EnvoyProxy` receive distinct services and snapshots.
* **Counter isolation:** two dedicated RLS and the shared RLS using the same Redis database keep
  separate counters, including for `shared: true` rules and a shared RLS with an admin-set prefix such
  as `eg:rl:`. Cross-scope policy warnings appear and clear correctly.
* **Caller isolation:** without any NetworkPolicy, a Gateway's proxies can call its dedicated RLS;
  another Gateway's proxies, including one in the same namespace and in the default deployment mode,
  are rejected during the handshake, also when pointed at that RLS through `EnvoyPatchPolicy`; the
  shared client certificate is rejected by every dedicated RLS and still accepted by the shared RLS.
  Proxies reject a dedicated RLS whose certificate does not match its Service name. The configuration
  server refuses a dedicated RLS that requests another identity's node ID, and refuses an RLS serving
  certificate used as a client. Neither certificate in a dedicated RLS's Secret is accepted as a caller
  by that RLS or by another Gateway's RLS in the same namespace.
* **Certificates:** a fresh install and an upgraded install both get the RLS CA Secret without changing
  existing certificates; a missing or invalid RLS CA reports `CertificateUnavailable` and leaves the
  shared RLS working; when the RLS CA Secret is created or fixed after the controller started, the
  dedicated RLS loads its rules from the configuration server without a controller restart; controller restarts deliver reissued proxy client certificates without proxy
  restarts; new RLS certificates are written to the Secret before updating the pod template; dedicated-only
  installations work without the `envoy-rate-limit` Secret; with several controller replicas, every
  replica's certificates are accepted. Recreating a Gateway with the same namespace/name accepts its previous
  unexpired certificates under the current CA and reuses matching certificates when available;
  control-plane-CA clients cannot fetch dedicated snapshots. Infrastructure-leader changes reuse valid
  RLS certificate Secrets without triggering rollouts. Leaf expiry never exceeds CA expiry, and an
  expired CA reports `CertificateUnavailable`. No periodic certificate renewal or rollout occurs;
  expiry clears readiness, and readiness waits for the current certificate version
  to complete its rollout.
* **Namespace placement:** in the default mode every dedicated RLS resource is created in the controller
  namespace; in Gateway namespace mode every dedicated RLS resource, including its certificate Secret,
  is created in the Gateway's namespace, and nothing for it in the controller namespace. Every resource
  except the Secrets has an owner reference to the Gateway; the Secrets have only their Deployment as
  owner. The Helm chart grants Secret write access, and `ReferenceGrant` reads in
  the controller namespace, only with `dedicatedRateLimit.rbac.enabled`, in the controller namespace
  in the default mode and in Gateway namespaces in Gateway namespace mode, and the controller never
  modifies Secrets it did not create. With the knob enabled, turning `enableDedicatedRateLimit` on by
  editing the `EnvoyGateway` configuration alone provisions working dedicated RLS; without it, each
  dedicated RLS reports `ProvisioningFailed`, not `CertificateUnavailable`; with `watch.namespaces` in Gateway
  namespace mode, the grant is on each per-namespace Role. In both modes every dedicated RLS is
  rendered with the fully qualified configuration server address; in Gateway namespace mode it loads
  its rules from a pod in a namespace other than the controller's. With `watch.namespaces` that omits the controller namespace,
  a `ReferenceGrant` there still enables the default-mode TLS exception.
* **Counter reuse:** switching back to a previously used shared or dedicated keyspace preserves matching
  unexpired counters rather than promising a reset.
* **Credentials:** initial missing references, rotation, deletion and ReferenceGrant revocation update
  copies, rollouts and status as specified, without fallback or removal of rate limit filters; the
  copied Secret is written before the pod template changes, and a failed Secret write leaves the
  template unchanged; deletion keeps old pods serving under the built-in strategy. Secrets in the RLS's
  namespace are read without a copy, and their rotation still rolls the RLS. In both modes a reference
  to another tenant namespace reports `InvalidCredentials` even with a `ReferenceGrant`. In the default
  mode, a TLS `certificateRef` to a controller-namespace Secret works only with a grant, revoking the
  grant reports `InvalidCredentials` and leaves the administrator-owned Secret and existing direct
  mount intact; replacement pods may still use that mount. Verify this behavior after controller
  restart, that an initially denied reference never gains a mount, and that restoring the grant clears
  the reference failure. Envoy Gateway's own Secrets are refused even with a grant;
  in Gateway namespace mode the same reference is refused.
* **Reconciliation:** controller restart reconstructs scope selection and readiness; Gateway deletion
  removes only its resources; updates preserve scope-specific selectors and HPA targets; the shared RLS
  Service and PDB never select dedicated pods; in dedicated-only mode, and when one of several dedicated
  RLS is deleted, the remaining dedicated RLS keep their ServiceAccounts and ConfigMaps. Every dedicated
  RLS Secret is owned only by its Deployment and is garbage-collected with it, also when
  `dedicatedRateLimit.rbac.enabled` is off. Removing `rateLimit` or deleting a Gateway while Envoy
  Gateway runs removes its dedicated RLS and Secrets. Turning off `enableDedicatedRateLimit` leaves
  dedicated RLS in place, and the documented per-Gateway delete removes them with their Secrets
  without touching other Gateways or another installation's dedicated RLS. Switching the deployment
  mode in either direction provisions the RLS in the new namespace and leaves the old placement, which
  the documented per-Gateway delete removes without touching the new one. Gateways in the
  controller namespace retain their RLS when switching modes and are excluded from cleanup.
* **API:** an `EnvoyProxy` whose `rateLimit` omits `failClosed` is accepted and behaves as `false`;
  generated API docs for `urlRef` and `certificateRef` describe both the `EnvoyGateway` and the
  `EnvoyProxy` namespace rules.
* **Unsupported infrastructure:** with the host provider, a custom provider with remote infrastructure,
  or a custom provider without infrastructure, and `enableDedicatedRateLimit` true, `rateLimit` sets the Gateway `Accepted=False` and no RLS
  infrastructure request is sent; with the switch false, the Gateway stays accepted and follows the
  disabled-switch behavior.
* **Transitions and failures:** exercise shared/dedicated switches, feature disablement, missing shared
  RLS, controller restart, unavailable services and delayed or rejected proxy updates. Verify ordinary
  convergence and the documented unknown-domain enforcement gap even with `failClosed: true`.
* **Credential selection:** switching keys within an unchanged Secret, changing TLS references, and
  changing reference purpose update the credential version and roll pods without hashing secret data.
* **Tenant safeguards documentation:** explain, per deployment mode, where dedicated RLS run and which
  operator-owned capacity, network and workload security controls apply, including the default-mode
  risks of controller-namespace capacity and selecting privileged override configurations.
  Implementing or testing an operator's admission policies is outside this proposal. Document that
  caller isolation depends on keeping the RLS CA Secret away from tenants, how to supply the RLS CA
  with custom certificates, and that pull secrets must exist in each Gateway namespace in Gateway
  namespace mode. The installation guide repeats the warning about turning off
  `enableDedicatedRateLimit` next to the setting, with the per-Gateway retirement procedure, and
  documents the `dedicatedRateLimit.rbac.enabled` Helm value and the permission checks to run after
  turning the feature on or off.

Use schema/CEL tests for API validation, Gateway translation fixtures for selection and conditions,
infrastructure tests for rendered resources, and xDS/configuration tests for
per-scope routing and rule distribution. End-to-end tests must demonstrate shared/dedicated coexistence,
dedicated-only operation, same-Redis counter isolation and credential rotation. Transition tests
must control configuration delivery rather than depend on a timing race. These checks are requirements
for implementation; this proposal itself changes no executable behavior.

## TODO

The following are follow-ups, not requirements for the supported v1 deployment modes.

### Sweeping stale dedicated RLS resources

Automatically remove dedicated RLS left behind by changes Envoy Gateway did not observe, such as
turning off `enableDedicatedRateLimit`, switching the deployment mode or changes made while it was
stopped. After each start,
including a reload, once the infrastructure leader has processed the first complete set of desired
infrastructure, Envoy Gateway could list dedicated RLS resources by label and delete those whose scope
is not desired. The design must:

* Add an installation identity label, for example a hash of the controller name and namespace, and
  restrict the sweep to it, so that installations sharing a namespace never delete each other's RLS.
  This would also lift the single-installation restriction.
* Use direct API reads, because the label-filtered informer caches select only the shared RLS and
  proxy names.
* Never run before the first complete desired set, so a partial view cannot delete a live RLS.
* Consider covering per-Gateway proxy resources at the same time.

### Coordinated RLS scope changes

A future design could retain source rules until proxies have switched and wait for destination RLS
configuration ACKs before publishing proxy updates. It must address disconnected proxies, multiple
controller replicas, restart recovery and cleanup. Redis migration remains separate from this work.

### ACK-based configuration readiness

A future condition could report whether all serving RLS endpoints have loaded the desired rules.
This requires ACK observations across controller replicas, with stale observations invalidated on
restart or disconnect. V1 reports Kubernetes infrastructure and credential readiness only.

### Validating RLS workload overrides

Neither shared nor dedicated RLS workload overrides are validated in this proposal (see
[Trust boundary](#trust-boundary)). If users need it, a follow-up could reject overrides of generated
settings, for both modes, instead of leaving that to admission policy.

### RLS certificate renewal

A future proposal may add automatic certificate renewal for both shared and dedicated RLS, covering
proxy client certificates and RLS server/configuration-client certificates consistently. It should
address certificate delivery, workload reloads or rollouts, controller restarts and custom certificates.
CA rotation is also deferred; a future workflow could temporarily trust old and new CAs while replacing
leaf certificates, then remove the old CA. Neither renewal nor coordinated CA replacement is required
for v1.

### Generated NetworkPolicies for RLS

Consider letting Envoy Gateway generate egress NetworkPolicies for RLS pods, for example one per
dedicated RLS that allows only its configured Redis and telemetry endpoints. This would apply to the
shared RLS as well, and needs a way to resolve Redis addresses to namespaces or IP ranges, so it is
left to a follow-up.

### GatewayClass scope

Only needed if users ask for it. A dedicated RLS shared by every Gateway in a GatewayClass could be
added later, by honouring `rateLimit`
on the class `EnvoyProxy`. It trades isolation for footprint: Gateways in the class share capacity and a
failure domain again, but with fewer RLS deployments and a common Redis connection, which fits when the
GatewayClass is already the team boundary. It would also give merged Gateways a dedicated RLS.

The identity scheme above already leaves room for a GatewayClass scope. Open questions for that design:

* **Inheritance.** Whether a Gateway-level `rateLimit` may override the class RLS, and how the existing
  `mergeType` modes (replace, StrategicMerge, JSONMerge) apply to the block. One option treats
  `rateLimit` as a single unit and lets the Gateway block win; another makes the class RLS mandatory and
  rejects Gateway-level blocks in that class.
* **Unrelated Gateway overrides.** With replace semantics, a Gateway bringing its own `EnvoyProxy` for an
  unrelated reason would silently leave the class RLS.
* **Workload settings.** Whether Gateways on the class RLS may influence its workload settings.
* **Lifecycle.** Keeping the class RLS while at least one Gateway uses it, and rule migration
  when Gateways move between the class RLS and others; coordinated cutover needs a separate design.
* **Shared buckets.** Gateway overrides within a class can split `shared: true` buckets across Redis
  backends.
* **Migration from ignored configuration.** Class `rateLimit` blocks that are ignored today would start
  taking effect, so the rollout needs a clear upgrade note.
