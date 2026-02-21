---
title: "Observability"
---


## Overview

Observability in Envoy Gateway provides comprehensive insight into both the control plane and Envoy Proxy instances. It allows users to collect, monitor, and analyze metrics, logs, and traces to ensure system health, optimize performance, and troubleshoot issues efficiently.

Envoy Gateway leverages standards like Prometheus and OpenTelemetry to export metrics and supports integration with external observability tools such as Grafana and Jaeger. Observability features are configurable to meet diverse operational requirements, empowering operators with actionable visibility across their infrastructure.

## Key Concepts

| Concept      | Description |
|--------------|-------------|
| Metrics      | Quantitative data from Envoy Gateway and Envoy Proxy. Supports Prometheus (pull) and OpenTelemetry (push) sinks. Includes counters, gauges, histograms, and labels for component/resource identification. |
| Logs         | Structured logs from the Envoy Gateway control plane, managed by an internal zap-based library and written to `/dev/stdout`. Logging levels can be configured per component. |
| Traces       | Distributed tracing using OpenTelemetry, Zipkin, or Datadog tracers. Configurable via the EnvoyProxy CRD with support for sampling rate and custom tags. |
| Runners   | Core telemetry sources: Resource Provider, xDS Translator, Infra Manager, xDS Server, Status Updater. Each emits labeled metrics to support monitoring and troubleshooting. |

## Use Cases

Use observability in Envoy Gateway to:

- Monitor control plane health and proxy behavior.
- Troubleshoot latency, errors, and traffic anomalies.
- Integrate with external tools like Prometheus, Grafana, and OpenTelemetry Collector.
- Enable auditing and compliance via centralized logging.
- Set up real-time alerting and dashboards with Prometheus and OpenTelemetry metrics.

## Observability in Envoy Gateway

Envoy Gateway implements observability as a core capability, covering metrics, logs, and traces, all of which can be surfaced via standard interfaces and integrated with external platforms. Metrics are collected from key components (Resource Provider, xDS Translator, Infra Manager, Status Updater, xDS Server) and exported to Prometheus or pushed to OpenTelemetry. Logging is written to standard output, with configurable verbosity. Distributed tracing is supported for OpenTelemetry, Zipkin, and Datadog tracers (OpenTelemetry is currently supported), with configuration at the EnvoyProxy CRD level. This ensures deep visibility into both Envoy Gateway’s operations and the traffic handled by Envoy Proxy instances.

## Examples

- Query Envoy Gateway metrics using Prometheus by port-forwarding to the Prometheus service and running a `curl` command against the `/api/v1/query` endpoint.
- Disable Prometheus metrics in EnvoyProxy by setting `telemetry.metrics.prometheus.disable` to `true` in the EnvoyProxy CRD.
- Configure Envoy Gateway to export metrics to an OpenTelemetry Collector by specifying a sink of type `OpenTelemetry` in the EnvoyProxy resource.
- Enable the debug exporter in the OpenTelemetry Collector to view raw metrics output in pod logs.

## Related Resources

- [Proxy Metrics (latest)](https://gateway.envoyproxy.io/latest/tasks/observability/proxy-metric/)
- [Grafana Integration](https://gateway.envoyproxy.io/latest/tasks/observability/grafana-integration/)
- [Gateway Exported Metrics List](https://gateway.envoyproxy.io/latest/tasks/observability/gateway-exported-metrics/)
- [Addons Helm Chart Installation Guide](https://gateway.envoyproxy.io/docs/install/gateway-addons-helm-api/)
- [OpenTelemetry](https://opentelemetry.io/)
- [Envoy Admin Endpoint Documentation](https://www.envoyproxy.io/docs/envoy/latest/operations/admin)
