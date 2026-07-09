---
title: "Merge Backends (Cluster Deduplication)"
---

{{% alert title="Experimental" color="warning" %}}
`mergeBackends` is an experimental feature and its behavior may change in future releases.
{{% /alert %}}

By default, Envoy Gateway generates **one Envoy cluster per route rule**. When the same backend
(for example a Kubernetes `Service`) is referenced by many routes, this produces many identical
clusters. That inflates the xDS configuration, multiplies active health-check traffic, fragments
passive health-check status, de-optimizes upstream connection pooling, and increases stats
cardinality.

The `mergeBackends` field on [EnvoyProxy][] enables **cluster deduplication**: routes that
reference the same backend — identified by its `kind`, `namespace`, `name`, and `port` — share a
single Envoy cluster.

## Modes

`mergeBackends` accepts one of two modes:

- **`BestEffort`**: a backend is merged into a shared cluster only when it is safe to do so. A
  backend whose route carries backend-cluster-scoped settings (for example a route-targeted
  [BackendTrafficPolicy][] configuring `healthCheck`, `circuitBreaker`, `loadBalancer`, `timeout`,
  `connection`, `dns`, `http2`, `tcpKeepalive`, or `proxyProtocol`), or that uses traffic-splitting
  features incompatible with weighted clusters (consistent-hash load balancing or session
  persistence), falls back to a dedicated per-route cluster. No error is reported. The shared
  cluster's settings come from a backend-targeted BackendTrafficPolicy, falling back to the
  Gateway-targeted BackendTrafficPolicy as the floor.
- **`Force`**: the backend is always merged into a single shared cluster, even when a route-targeted
  BackendTrafficPolicy configures backend-cluster-scoped settings (those settings are ignored for
  the shared cluster).

When `mergeBackends` is unset, cluster deduplication is disabled and Envoy Gateway keeps the
default one-cluster-per-route-rule behavior.

## Example

Enable `BestEffort` cluster deduplication on the GatewayClass-level EnvoyProxy:

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: custom-proxy-config
  namespace: envoy-gateway-system
spec:
  mergeBackends: BestEffort
```

With this configuration, two HTTPRoutes that both forward to `Service` `foo` on port `80` produce a
single Envoy cluster named `backend/service/<namespace>/foo/80`, referenced by both routes, instead
of two separate clusters.

`mergeBackends` can also be set as a **global default** for every GatewayClass through the
EnvoyGateway configuration's default EnvoyProxy spec:

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyGateway
# ...
envoyProxy:
  mergeBackends: BestEffort
```

## Caveats

- Merging is scoped to a single Envoy Proxy configuration. When [merged Gateways][] serve multiple
  Gateways with different Gateway-targeted BackendTrafficPolicy floors from one xDS snapshot, the
  first-written shared cluster configuration wins for a given backend.
- In `BestEffort` mode, a route that keeps a dedicated cluster (because it carries route-level
  cluster settings) does not benefit from deduplication.

[EnvoyProxy]: ../../api/extension_types#envoyproxy
[BackendTrafficPolicy]: ../../api/extension_types#backendtrafficpolicy
[merged Gateways]: ../operations/deployment-mode#merged-gateways-onto-a-single-envoyproxy
