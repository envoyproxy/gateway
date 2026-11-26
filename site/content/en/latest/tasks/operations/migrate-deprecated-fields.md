---
title: "Migrate Deprecated API Fields"
---

Several Envoy Gateway API fields have been deprecated for more than a year and are removed in v1.11.
Envoy Gateway reports them in a `DeprecatedField` warning condition on the status of the policy or
EnvoyProxy that uses them, and in its logs for the EnvoyGateway configuration.

`egctl x migrate deprecated` rewrites these fields to their replacements, so you can update your
manifests before upgrading. Follow the instructions [here](./../../install/install-egctl.md) to
install `egctl`.

## Migrate manifest files

Point `-f` at a file, a directory (searched recursively for `.yaml` and `.yml` files), or `-` for
stdin. By default, the migrated YAML is printed to stdout, and a report of every deprecated field is
written to stderr:

```shell
egctl x migrate deprecated -f manifests/
```

Use individual resource documents, separated by `---` when a file contains multiple resources.
`List` and typed lists such as `SecurityPolicyList` are rejected before any files are changed;
split their `items` into individual resource documents first.

Given this ClientTrafficPolicy:

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: ClientTrafficPolicy
metadata:
  name: edge
  namespace: default
spec:
  # Attach to the edge gateway.
  targetRef:
    group: gateway.networking.k8s.io
    kind: Gateway
    name: eg
  enableProxyProtocol: true  # behind an NLB
  headers:
    preserveXRequestID: true
```

The command prints:

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: ClientTrafficPolicy
metadata:
  name: edge
  namespace: default
spec:
  # Attach to the edge gateway.
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: Gateway
    name: eg
  proxyProtocol: {}  # behind an NLB
  headers:
    requestID: PreserveOrGenerate
```

And reports:

```console
KIND                 NAMESPACE/NAME                     FIELD                            RESULT    NOTE
ClientTrafficPolicy  default/edge (manifests/ctp.yaml)  spec.enableProxyProtocol         migrated  replaced with spec.proxyProtocol: {}
ClientTrafficPolicy  default/edge (manifests/ctp.yaml)  spec.headers.preserveXRequestID  migrated  replaced with spec.headers.requestID: PreserveOrGenerate
ClientTrafficPolicy  default/edge (manifests/ctp.yaml)  spec.targetRef                   migrated  moved to spec.targetRefs
```

Comments and field order are kept, and documents without deprecated fields are written back
unchanged.

### Modes

`--mode` selects what to do with the migrated YAML:

| Mode | Migrated YAML | Report |
|---|---|---|
| `print` (default) | printed to stdout | stderr |
| `in-place` | written back to the files, only the ones that changed | stderr |
| `check-only` | not written | stderr |

To rewrite the files of a Git repository:

```shell
egctl x migrate deprecated -f manifests/ --mode in-place
git diff
```

To fail a CI job while deprecated fields remain:

```shell
egctl x migrate deprecated -f manifests/ --mode check-only
```

### Exit codes

| Code | Meaning |
|---|---|
| `0` | Nothing to migrate, or everything was migrated. |
| `1` | `--mode check-only` found fields to migrate. |
| `2` | At least one field can't be migrated automatically. It is left unchanged and reported as `error`. |

## Read resources from the cluster

`--from-cluster` reads ClientTrafficPolicy, BackendTrafficPolicy, SecurityPolicy,
EnvoyExtensionPolicy and EnvoyProxy resources, and the `envoy-gateway-config` ConfigMap, from the
cluster. It prints them migrated, without status and server-set metadata, but doesn't change the
cluster:

```shell
# All namespaces
egctl x migrate deprecated --from-cluster -A

# One namespace
egctl x migrate deprecated --from-cluster -n default --mode check-only
```

Use `--eg-namespace` if Envoy Gateway isn't installed in `envoy-gateway-system`.

Resources managed by Helm, Argo CD or Flux are printed unchanged and their deprecated fields are
reported as `error`: the tool would revert the change on its next sync, so fix them in their source
instead. A resource is treated as managed by Argo CD if it has the `app.kubernetes.io/instance` or
`argocd.argoproj.io/instance` label, or the `argocd.argoproj.io/tracking-id` annotation.

## Fields that are reported but not migrated

Some fields can't be migrated without a decision from you. They're left unchanged, reported as
`error`, and the command exits with code `2`.

### Telemetry `host` and `port`

The `host` and `port` of the tracing provider, the OpenTelemetry metrics sink and the OpenTelemetry
access log sink in EnvoyProxy are replaced by `backendRefs`.

If `host` is the DNS name of a Service (`<service>.<namespace>.svc` or
`<service>.<namespace>.svc.<cluster-domain>`, including custom cluster domains), it is migrated to a `backendRefs` entry for that Service.
Envoy then sends to the Service endpoints instead of resolving its DNS name. A reference to a
Service in another namespace needs a [ReferenceGrant][] that allows the EnvoyProxy to reference it.

Any other host is reported as `error`. Set `backendRefs` to a Service, or to a [Backend][] with an
`fqdn` or `ip` endpoint, which requires `extensionApis.enableBackend: true` in the EnvoyGateway
configuration:

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: Backend
metadata:
  name: otel-collector
  namespace: envoy-gateway-system
spec:
  endpoints:
  - fqdn:
      hostname: collector.example.com
      port: 4317
```

```yaml
telemetry:
  tracing:
    provider:
      type: OpenTelemetry
      backendRefs:
      - group: gateway.envoyproxy.io
        kind: Backend
        name: otel-collector
```

### EnvoyGateway configuration

The EnvoyGateway configuration is usually rendered from the `config.envoyGateway` values of the
`gateway-helm` chart, so the command never rewrites it, whether it reads it from the
`envoy-gateway-config` ConfigMap, an EnvoyGateway file or a Helm values file. Update these fields in
your Helm values instead:

| Deprecated | Replacement |
|---|---|
| `extensionManager.service.host`, `extensionManager.service.port` (and the same under `extensionManagers` and `provider.custom.infrastructure.remote.service`) | `fqdn` with `hostname` and `port`, or `ip` with `address` and `port` |
| `extensionApis.disableLua` | `extensionApis.enableLua` |
| `envoyProxy.luaValidation` | `envoyProxy.lua.validationType` |
| `envoyProxy.telemetry...` `host`, `port` | `backendRefs` |

### Conflicting fields

If a deprecated field and its replacement are both set and the result would change, the field is
reported as `error`. For example, when `targetRef` and `targetRefs` reference different resources,
only `targetRef` is used today, so decide which references to keep and move them to `targetRefs`.

## Migrated fields

| Resource | Deprecated | Replacement |
|---|---|---|
| All policies | `spec.targetRef` | `spec.targetRefs` |
| ClientTrafficPolicy | `spec.enableProxyProtocol: true` | `spec.proxyProtocol: {}` |
| ClientTrafficPolicy | `spec.headers.preserveXRequestID: true` | `spec.headers.requestID: PreserveOrGenerate` |
| ClientTrafficPolicy | `spec.tls.clientValidation.optional: true` | `spec.tls.clientValidation.mode: VerifyIfGiven` |
| BackendTrafficPolicy | `spec.compression` | `spec.compressor`, with an empty settings block such as `gzip: {}` added where it is missing, since `compressor` ignores an entry without one |
| All load balancers | `consistentHash.type: Header` and `consistentHash.header` | `consistentHash.type: Headers` and `consistentHash.headers` |
| SecurityPolicy | External auth `backendRef` | `backendRefs` |
| EnvoyProxy | `spec.luaValidation` | `spec.lua.validationType` |
| EnvoyProxy | Telemetry `host` and `port` | `backendRefs`, see [above](#telemetry-host-and-port) |

A deprecated field that is set to its default, such as `enableProxyProtocol: false`, or that is
ignored because its replacement is also set, is removed.

[ReferenceGrant]: https://gateway-api.sigs.k8s.io/reference/api-types/referencegrant/
[Backend]: ../../tasks/traffic/backend.md
