---
title: "Endpoint Fast Path"
---

The experimental endpoint fast path can publish eligible Kubernetes EndpointSlice
changes while full route and policy translation is still running. It is disabled
by default. It supports endpoint routing for Services and ServiceImports.

Enable it in the Envoy Gateway configuration:

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyGateway
provider:
  type: Kubernetes
runtimeFlags:
  enabled:
  - EndpointFastPath
```

For a Helm installation, set `config.envoyGateway.runtimeFlags.enabled` to include
`EndpointFastPath`. Keep any other runtime flags you already enabled. Restart the
Envoy Gateway deployment after changing its configuration.

## Behavior

Endpoint updates use the currently published routes and policies. A later full
translation installs configuration changes and preserves newer endpoints where
they remain compatible. Both paths record the endpoint revisions they publish,
so a delayed notification cannot replace newer applied endpoints.

Full reconciliation still runs for every endpoint change. The fast path reduces
the wait for eligible endpoint updates; it does not eliminate full translation
work or immediately update route status.

The following cases use full reconciliation:

- A backend changes between zero and nonzero endpoints.
- The endpoint address type changes, such as from IP to DNS.
- A patch changes the cluster or its endpoint assignment.
- An extension manager is configured.
- A destination uses ClusterIP routing, static Backend addresses, or DNS resolution.
- No successful full configuration with an eligible endpoint assignment exists yet.

When a full result cannot safely reuse newer applied endpoints, it publishes its
own coherent result. For example, an empty backend may produce a direct 503
response. A subsequent full reconciliation restores forwarding after endpoints
return; this transition can briefly return 503.

## Observability

The control-plane metrics include:

| Metric | Meaning |
| --- | --- |
| `xds_endpoint_publications_total` | Desired publications, labeled `path=full`, `fast`, or `delete`. |
| `xds_endpoint_refresh_duration_seconds` | Fast refresh time, including validation, hashing, and cache installation. |
| `xds_endpoint_refresh_skipped_total` | Contributions skipped because the revision was already applied, endpoint emptiness changed, or the address type changed. |
| `xds_endpoint_delivery_retries_total` | Delivery retries using the current desired state. |

Publication counters describe control-plane commits, not Envoy acknowledgements.
Use the existing xDS NACK metrics and traffic checks to verify data-plane behavior.

## Development testing

The `endpoint-fast-path` deployment profile enables the flag. With a test cluster
already installed using that profile, run:

```shell
KUBE_DEPLOY_PROFILE=endpoint-fast-path make e2e-prepare
KUBE_DEPLOY_PROFILE=endpoint-fast-path make run-e2e E2E_RUN_TEST=EndpointFastPath
```

The test moves traffic between two backends, removes all endpoints, and restores
them. It checks convergence through both endpoint refresh and full fallback.
