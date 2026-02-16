---
title: "Gateway Exported Metrics"
---

The Envoy Gateway provides a collection of self-monitoring metrics in [Prometheus format][prom-format]. 

These metrics allow monitoring of the behavior of Envoy Gateway itself (as distinct from that of the EnvoyProxy it managed).

{{% alert title="EnvoyProxy Metrics" color="warning" %}}
For EnvoyProxy Metrics, please refer to the [EnvoyProxy Metrics](./proxy-metric) to learn more.
{{% /alert %}}

## Watching Components

The Resource Provider, xDS Translator and Infra Manager etc. are key components that made up of Envoy Gateway,
they all follow the design of [Watching Components](../../../contributions/design/watching).

Envoy Gateway collects the following metrics in Watching Components:

| Name                                   | Description                                                  |
|----------------------------------------|--------------------------------------------------------------|
| `watchable_depth`                      | Current depth of watchable map.                              |
| `watchable_subscribe_duration_seconds` | How long in seconds a subscribed watchable queue is handled. |
| `watchable_subscribe_total`            | Total number of subscribed watchable queue.                  |
| `watchable_panics_recovered_total`     | Total recovered panics in the watchable infrastructure.      |

Each metric includes the `runner` label to identify the corresponding components,
the relationship between label values and components is as follows:

| Value              | Components                      |
|--------------------|---------------------------------|
| `gateway-api`      | Gateway API Translator          |
| `infrastructure`   | Infrastructure Manager          |
| `xds-server`       | xDS Server                      |
| `xds-translator`   | xDS Translator                  |
| `global-ratelimit` | Global RateLimit xDS Translator |

Metrics may include one or more additional labels, such as `message`, `status` and `reason` etc.

## Status Updater

Envoy Gateway monitors the status updates of various resources (like `GatewayClass`, `Gateway` and `HTTPRoute` etc.) through Status Updater.

Envoy Gateway collects the following metrics in Status Updater:

| Name                             | Description                                    |
|----------------------------------|------------------------------------------------|
| `status_update_total`            | Total number of status update by object kind.  |
| `status_update_duration_seconds` | How long a status update takes to finish.      |

Each metric includes `kind` label to identify the corresponding resources.

## xDS Server

Envoy Gateway monitors the cache and xDS connection status in xDS Server.

Envoy Gateway collects the following metrics in xDS Server:

| Name                          | Description                                            |
|-------------------------------|--------------------------------------------------------|
| `xds_snapshot_create_total`   | Total number of xds snapshot cache creates.            |
| `xds_snapshot_update_total`   | Total number of xds snapshot cache updates by node id. |
| `xds_stream_duration_seconds` | How long a xds stream takes to finish.                 |

- For xDS snapshot cache update and xDS stream connection status, each metric includes `nodeID` label to identify the connection peer.
- For xDS stream connection status, each metric also includes `streamID` label to identify the connection stream, and `isDeltaStream` label to identify the delta connection stream.

## Infrastructure Manager

Envoy Gateway monitors the `apply` (`create` or `update`) and `delete` operations in Infrastructure Manager.

Envoy Gateway collects the following metrics in Infrastructure Manager:

| Name                               | Description                                             |
|------------------------------------|---------------------------------------------------------|
| `resource_apply_total`             | Total number of applied resources.                      |
| `resource_apply_duration_seconds`  | How long in seconds a resource be applied successfully. |
| `resource_delete_total`            | Total number of deleted resources.                      |
| `resource_delete_duration_seconds` | How long in seconds a resource be deleted successfully. |

Each metric includes the `kind` label to identify the corresponding resources being applied or deleted by Infrastructure Manager.

Metrics may also include `name` and `namespace` label to identify the name and namespace of corresponding Infrastructure Manager.

## Wasm

Envoy Gateway monitors the status of Wasm remote fetch cache.

| Name                      | Description                                      |
|---------------------------|--------------------------------------------------|
| `wasm_cache_entries`      | Number of Wasm remote fetch cache entries.       | 
| `wasm_cache_lookup_total` | Total number of Wasm remote fetch cache lookups. |
| `wasm_remote_fetch_total` | Total number of Wasm remote fetches and results. |

For metric `wasm_cache_lookup_total`, we are using `hit` label (boolean) to indicate whether the Wasm cache has been hit.


[prom-format]: https://prometheus.io/docs/instrumenting/exposition_formats/#text-based-format
