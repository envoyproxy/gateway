---
title: "Overload Manager"
---

Envoy Gateway configures the Envoy [overload manager][] in the bootstrap of every managed Envoy proxy. The overload manager protects the proxy from running out of memory or file descriptors by shedding load before the process fails.

This page describes the defaults Envoy Gateway applies today. They are implementation defaults and may change in future releases. There is no dedicated `EnvoyProxy` field to configure them yet, see [#3604](https://github.com/envoyproxy/gateway/issues/3604).

## Downstream connection limit

Envoy Gateway always configures the `envoy.resource_monitors.global_downstream_max_connections` resource monitor with a limit of 50000 active downstream connections, shared by all listeners of an Envoy proxy. Once the limit is reached, Envoy rejects new connections until existing ones close.

## Heap memory limits

When the Envoy container has a memory limit, Envoy Gateway also configures the `envoy.resource_monitors.fixed_heap` resource monitor. Its maximum heap size is set to 80% of the container memory limit, which leaves headroom for memory that is not allocated on the heap.

The monitor reports heap pressure as the heap memory used by Envoy divided by that maximum heap size. Two overload actions are triggered by it:

| Overload action | Heap pressure | Share of the container memory limit | Effect |
|---|---|---|---|
| `envoy.overload_actions.shrink_heap` | 95% | 76% | Envoy periodically releases free heap memory back to the operating system. |
| `envoy.overload_actions.stop_accepting_requests` | 98% | 78.4% | Envoy responds to every new HTTP request with a `503` and the body `envoy overloaded`. Requests already in progress and TCP or UDP traffic are not affected. |

For example, with a memory limit of `1Gi`, the maximum heap size is about 819Mi, and Envoy stops accepting requests when its heap usage reaches about 803Mi.

The heap usage measured by Envoy is lower than the memory usage of the container reported by Kubernetes, which also counts the Envoy binary, thread stacks and kernel memory charged to the container. Container memory metrics can therefore be above 78.4% of the limit before Envoy rejects requests.

If the Envoy container has no memory limit, which is the default, no heap monitor and no heap overload actions are configured. Only the downstream connection limit applies.

You can set the memory limit as described in [Customize EnvoyProxy Deployment Resources](../customize-envoyproxy#customize-envoyproxy-deployment-resources). The same applies to an Envoy proxy deployed as a DaemonSet.

## Readiness and metrics

The listener that serves the readiness probe bypasses the overload manager, so an overloaded Envoy proxy keeps reporting ready while it rejects traffic. When Prometheus metrics are enabled, the listener that serves them bypasses the overload manager as well.

## Autoscaling

If you autoscale the Envoy proxy on memory usage, as described in [Customize EnvoyProxy Horizontal Pod Autoscaler](../customize-envoyproxy#customize-envoyproxy-horizontal-pod-autoscaler), scale out before Envoy stops accepting requests. The HPA computes memory utilization against the memory request, and measures container memory, which is higher than the heap usage Envoy measures. Keep the target utilization multiplied by the memory request well below 78.4% of the memory limit, and leave room for the time new pods take to become ready.

For example, with a memory request and limit of `1Gi`, a target utilization of 70% scales out at about 717Mi of container memory, before Envoy stops accepting requests at about 803Mi of heap usage.

## Monitoring

The overload manager emits the following statistics. The heap statistics only exist when the Envoy container has a memory limit.

| Statistic | Prometheus metric | Description |
|---|---|---|
| `overload.envoy.resource_monitors.fixed_heap.pressure` | `envoy_overload_envoy_resource_monitors_fixed_heap_pressure` | Heap pressure, as a percent of the maximum heap size. |
| `overload.envoy.resource_monitors.global_downstream_max_connections.pressure` | `envoy_overload_envoy_resource_monitors_global_downstream_max_connections_pressure` | Active downstream connections, as a percent of the limit. |
| `overload.envoy.overload_actions.shrink_heap.active` | `envoy_overload_envoy_overload_actions_shrink_heap_active` | `1` while the shrink heap action is active. |
| `overload.envoy.overload_actions.stop_accepting_requests.active` | `envoy_overload_envoy_overload_actions_stop_accepting_requests_active` | `1` while Envoy rejects new requests. |

Alert on `stop_accepting_requests.active` to detect an Envoy proxy that is shedding load, and warn on a heap pressure above 90 to act before it does. If you restrict the exported statistics with `telemetry.metrics.matches` in the `EnvoyProxy` resource, include these statistics in the match list.

[overload manager]: https://www.envoyproxy.io/docs/envoy/latest/configuration/operations/overload_manager/overload_manager
