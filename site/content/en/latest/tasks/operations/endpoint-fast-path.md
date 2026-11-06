---
title: "Endpoint Fast Path"
---

The experimental endpoint fast path can publish eligible Kubernetes EndpointSlice
changes while full route and policy translation is still running. It is disabled
by default. It supports endpoint routing for Services and ServiceImports.

Without it, an endpoint change, such as a pod scaling up or becoming unready,
reaches Envoy only after a full translation of all routes and policies, which takes
longer as the configuration grows. With the fast path, Envoy Gateway rebuilds only
the endpoints of the clusters that use the changed Service and pushes them over
EDS. How quickly Envoy converges then depends on how many endpoints changed, not
on the size of the rest of the configuration. This helps most in large deployments
and during rollouts or autoscaling, when endpoints change often.

The faster convergence costs more resources. Full translation still runs for every
endpoint change, so the fast path's work is added on top of it. Envoy Gateway uses
more CPU, and more memory to keep the latest endpoints of each referenced Service
and the settings needed to rebuild each cluster's endpoints. Envoy may also receive
more updates, since an endpoint change can be pushed once by the fast path and again
by full translation. See [Performance](#performance) for measured numbers.

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

If you use an extension server, enable the fast path only if the server does not
implement the Endpoints hook. Fast updates rebuild endpoint assignments without
calling the hook, so its changes would be lost until the next full translation.

## Behavior

With the fast path on, Envoy Gateway publishes endpoint changes right away, using
the routes and policies already in effect. Full translation still runs for every
change as before. It still applies route and policy changes and updates route
status, so those are not any faster.

Endpoints never move backward. If a full translation that started before a newer
endpoint change finishes after it, Envoy keeps the newer endpoints.

A newly referenced Service gets its first endpoints through full translation.

The following cases use full reconciliation:

- A backend changes between zero and nonzero endpoints.
- The endpoint address type changes, such as from IP to DNS.
- An EnvoyPatchPolicy targets the endpoint assignment, or a patch or the extension
  server's Translation hook changes it. Other patches, such as a cluster timeout,
  keep the fast path.
- A destination uses ClusterIP routing, static Backend addresses, or DNS resolution.
- No successful full configuration with an eligible endpoint assignment exists yet.

When a full result cannot safely reuse newer applied endpoints, it publishes its
own coherent result. For example, an empty backend may produce a direct 503
response. A subsequent full reconciliation restores forwarding after endpoints
return; this transition can briefly return 503.

## Performance

One test ran on a kind cluster with 2,000 HTTPRoutes (8,000 clusters) and 15
EnvoyPatchPolicies, where a full xDS translation took about 7 seconds. A single
Service's endpoints were changed 40 times, and each time the test measured how long
Envoy took to get the new endpoints after the EndpointSlice was written:

| Other activity during the test | Flag off (p50 / p99) | Flag on (p50 / p99) |
| --- | --- | --- |
| None | 7.1 s / 7.3 s | 46 ms / 71 ms |
| Other Services' endpoints changing 5 times a second | 11.7 s / 15.4 s | 49 ms / 86 ms |
| An HTTPRoute edited every 2 seconds, so full translation never stops | 10.2 s / 14.0 s | 47 ms / 88 ms |

With the flag on, the new endpoints reached Envoy in under 100 ms in every case,
150 to 240 times faster at the median. Changes between one and zero endpoints
always go through full translation, so they took about 7.5 seconds with or without
the flag.

In the same test, the flag added about 10% control plane CPU and 100 to 200 MiB of
memory for the stored endpoint data.

## Observability

The control-plane metrics include:

| Metric | Meaning |
| --- | --- |
| `xds_snapshot_publications_total` | Snapshots published to the xDS cache, labeled `path=full`, `fast`, or `delete`, and `status=success` or `failure`. |
| `xds_endpoint_refresh_duration_seconds` | Time to apply a batch of fast endpoint updates to one xDS snapshot and publish it. |
| `xds_endpoint_refresh_skipped_total` | Endpoint updates not applied to a cluster, by `path` (`fast` or `full`) and `reason`: `already_applied` (fast path only: the cluster already has these or newer endpoints), `empty_transition` (the Service went to or from zero endpoints), or `address_type` (the endpoint address type changed). The last two are left to full translation. |

Publication counters describe control-plane commits, not Envoy acknowledgements.
Use the existing xDS NACK metrics and traffic checks to verify data-plane behavior.
