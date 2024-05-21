---
title: "Gateway Exported Metrics"
---

The Envoy Gateway provides a collection of self-monitoring metrics. 

These metrics allow monitoring of the behavior of Envoy Gateway itself (as distinct from that of the EnvoyProxy it managed).

{{% alert title="EnvoyProxy Metrics" color="warning" %}}
For EnvoyProxy Metrics, please refer to the [EnvoyProxy Observability](./proxy-observability#metrics) to learn more.
{{% /alert %}}

## Watching Components

The Resource Provider, xDS Translator and Infra Manager etc. are key components that made up of Envoy Gateway,
they all follow the design of [Watching Components](../../../contributions/design/watching).

Envoy Gateway collects the following metrics in Watching Components:

| Name                                    | Description                                            |
|-----------------------------------------|--------------------------------------------------------|
| `watchable_depth`                       | Current depth of watchable map.                        |
| `watchable_subscribed_duration_seconds` | How long in seconds a subscribed watchable is handled. |
| `watchable_subscribed_total`            | Total number of subscribed watchable.                  |
| `watchable_subscribed_errors_total`     | Total number of subscribed watchable errors.           |

Each metrics includes the `runner` label to identify the corresponding components,
the relationship between label values and components is as follows:

| Value              | Components                      |
|--------------------|---------------------------------|
| `gateway-api`      | Gateway API Translator          |
| `infrastructure`   | Infrastructure Manager          |
| `xds-server`       | xDS Server                      |
| `xds-translator`   | xDS Translator                  |
| `global-ratelimit` | Global RateLimit xDS Translator |

Metrics may include one or more additional labels, such as `message` etc.

## Status Updater

Envoy Gateway monitors the status updates of various resources (like `GatewayClass`, `Gateway` and `HTTPRoute` etc.) through Status Updater.

Envoy Gateway collects the following metrics in Status Updater:

| Name                             | Description                                                                                             |
|----------------------------------|---------------------------------------------------------------------------------------------------------|
| `status_update_total`            | Total number of status updates by object kind.                                                          |
| `status_update_failed_total`     | Number of status updates that failed by object kind.                                                    |
| `status_update_conflict_total`   | Number of status update conflicts encountered by object kind.                                           |
| `status_update_success_total`    | Number of status updates that succeeded by object kind.                                                 |
| `status_update_noop_total`       | Number of status updates that are no-ops by object kind. This is a subset of successful status updates. |
| `status_update_duration_seconds` | How long a status update takes to finish.                                                               |

Each metrics includes `kind` label to identify the corresponding resources.

## xDS Server

Envoy Gateway monitors the cache and xDS connection status in xDS Server.

Envoy Gateway collects the following metrics in xDS Server:

| Name                                | Description                                                   |
|-------------------------------------|---------------------------------------------------------------|
| `xds_snapshot_creation_total`       | Total number of xds snapshot cache creation.                  |
| `xds_snapshot_creation_failed`      | Number of xds snapshot cache creation that failed.            |
| `xds_snapshot_creation_success`     | Number of xds snapshot cache creation that succeed.           |
| `xds_snapshot_update_total`         | Total number of xds snapshot cache updates by node id.        |
| `xds_snapshot_update_failed`        | Number of xds snapshot cache updates that failed by node id.  |
| `xds_snapshot_update_success`       | Number of xds snapshot cache updates that succeed by node id. |
| `xds_stream_duration_seconds`       | How long a xds stream takes to finish.                        |
| `xds_delta_stream_duration_seconds` | How long a xds delta stream takes to finish.                  |

For xDS snapshot cache update and xDS stream connection status, each metrics includes `nodeID` label to identify the connection peer.
For xDS stream connection status, each metrics also includes `streamID` label to identify the connection stream.
