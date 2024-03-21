---
date: 2023-10-10
title: "Control Plane Observability: Metrics"
---

This document aims to cover all aspects of envoy gateway control plane metrics observability.

{{% alert title="Note" color="secondary" %}}
**Data plane** observability (while important) is outside of scope for this document. For dataplane observability, refer to [here](../metrics).
{{% /alert %}}

## Current State

At present, the Envoy Gateway control plane provides logs and controller-runtime metrics, without traces. Logs are managed through our proprietary library (`internal/logging`, a shim to `zap`) and are written to `/dev/stdout`.

## Goals

Our objectives include:

+ Supporting **PULL** mode for Prometheus metrics and exposing these metrics on the admin address.
+ Supporting **PUSH** mode for Prometheus metrics, thereby sending metrics to the Open Telemetry Stats sink via gRPC or HTTP.

## Non-Goals

Our non-goals include:

+ Supporting other stats sinks.

## Use-Cases

The use-cases include:

+ Exposing Prometheus metrics in the Envoy Gateway Control Plane.
+ Pushing Envoy Gateway Control Plane metrics via the Open Telemetry Sink.

## Design

### Standards

Our metrics, will be built upon the [OpenTelemetry][] standards. All metrics will be configured via the [OpenTelemetry SDK][], which offers neutral libraries that can be connected to various backends.

This approach allows the Envoy Gateway code to concentrate on the crucial aspect - generating the metrics - and delegate all other tasks to systems designed for telemetry ingestion.

### Attributes

OpenTelemetry defines a set of [Semantic Conventions][], including [Kubernetes specific ones][].

These attributes can be expressed in logs (as keys of structured logs), traces (as attributes), and metrics (as labels).

We aim to use attributes consistently where applicable. Where possible, these should adhere to codified Semantic Conventions; when not possible, they should maintain consistency across the project.

### Extensibility

Envoy Gateway supports both **PULL/PUSH** mode metrics, with Metrics exported via Prometheus by default.

Additionally, Envoy Gateway can export metrics using both the [OTEL gRPC metrics exporter][] and [OTEL HTTP metrics exporter][], which pushes metrics by grpc/http to a remote OTEL collector.

Users can extend these in two ways:

#### Downstream Collection

Based on the exported data, other tools can collect, process, and export telemetry as needed. Some examples include:

+ Metrics in **PULL** mode: The OTEL collector can scrape Prometheus and export to X.
+ Metrics in **PUSH** mode: The OTEL collector can receive OTEL gRPC/HTTP exporter metrics and export to X.

While the examples above involve OTEL collectors, there are numerous other systems available.

#### Vendor extensions

The OTEL libraries allow for the registration of Providers/Handlers. While we will offer the default ones (PULL via Prometheus, PUSH via OTEL HTTP metrics exporter) mentioned in Envoy Gateway's extensibility, we can easily allow custom builds of Envoy Gateway to plug in alternatives if the default options don't meet their needs.

For instance, users may prefer to write metrics over the OTLP gRPC metrics exporter instead of the HTTP metrics exporter. This is perfectly acceptable -- and almost impossible to prevent. The OTEL has ways to register their providers/exporters, and Envoy Gateway can ensure its usage is such that it's not overly difficult to swap out a different provider/exporter.

### Stability

Observability is, in essence, a user-facing API. Its primary purpose is to be consumed - by both humans and tooling. Therefore, having well-defined guarantees around their formats is crucial.

Please note that this refers only to the contents of the telemetry - what we emit, the names of things, semantics, etc. Other settings like Prometheus vs OTLP, JSON vs plaintext, logging levels, etc., are not considered.

I propose the following:

#### Metrics

Metrics offer the greatest potential for providing guarantees. They often directly influence alerts and dashboards, making changes highly impactful. This contrasts with traces and logs, which are often used for ad-hoc analysis, where minor changes to information can be easily understood by a human.

Moreover, there is precedent for this: [Kubernetes Metrics Lifecycle][] has well-defined processes, and Envoy Gateway's dataplane (Envoy Proxy) metrics are de facto stable.

Currently, all Envoy Gateway metrics lack defined stability. I suggest we categorize all existing metrics as either:

+ ***Deprecated***: a metric that is intended to be phased out.
+ ***Experimental***: a metric that is off by default.
+ ***Alpha***: a metric that is on by default.

We should aim to promote a core set of metrics to **Stable** within a few releases.

## Envoy Gateway API Types

New APIs will be added to Envoy Gateway config, which are used to manage Control Plane Telemetry bootstrap configs.

### EnvoyGatewayTelemetry

``` go
// EnvoyGatewayTelemetry defines telemetry configurations for envoy gateway control plane.
// Control plane will focus on metrics observability telemetry and tracing telemetry later.
type EnvoyGatewayTelemetry struct {
	// Metrics defines metrics configuration for envoy gateway.
	Metrics *EnvoyGatewayMetrics `json:"metrics,omitempty"`
}
```

### EnvoyGatewayMetrics

> Prometheus will be exposed on 0.0.0.0:19001, which is not supported to be configured yet.

``` go
// EnvoyGatewayMetrics defines control plane push/pull metrics configurations.
type EnvoyGatewayMetrics struct {
	// Sinks defines the metric sinks where metrics are sent to.
	Sinks []EnvoyGatewayMetricSink `json:"sinks,omitempty"`
	// Prometheus defines the configuration for prometheus endpoint.
	Prometheus *EnvoyGatewayPrometheusProvider `json:"prometheus,omitempty"`
}

// EnvoyGatewayMetricSink defines control plane
// metric sinks where metrics are sent to.
type EnvoyGatewayMetricSink struct {
	// Type defines the metric sink type.
	// EG control plane currently supports OpenTelemetry.
	// +kubebuilder:validation:Enum=OpenTelemetry
	// +kubebuilder:default=OpenTelemetry
	Type MetricSinkType `json:"type"`
	// OpenTelemetry defines the configuration for OpenTelemetry sink.
	// It's required if the sink type is OpenTelemetry.
	OpenTelemetry *EnvoyGatewayOpenTelemetrySink `json:"openTelemetry,omitempty"`
}

type EnvoyGatewayOpenTelemetrySink struct {
	// Host define the sink service hostname.
	Host string `json:"host"`
	// Protocol define the sink service protocol.
	// +kubebuilder:validation:Enum=grpc;http
	Protocol string `json:"protocol"`
	// Port defines the port the sink service is exposed on.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=4317
	Port int32 `json:"port,omitempty"`
}

// EnvoyGatewayPrometheusProvider will expose prometheus endpoint in pull mode.
type EnvoyGatewayPrometheusProvider struct {
	// Disable defines if disables the prometheus metrics in pull mode.
	//
	Disable bool `json:"disable,omitempty"`
}

```

#### Example

+ The following is an example to disable prometheus metric.

``` yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyGateway
gateway:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
logging:
  level: null
  default: info
provider:
  type: Kubernetes
telemetry:
  metrics:
    prometheus:
      disable: true
```

+ The following is an example to send metric via Open Telemetry sink to OTEL gRPC Collector.

``` yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyGateway
gateway:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
logging:
  level: null
  default: info
provider:
  type: Kubernetes
telemetry:
  metrics:
    sinks:
      - type: OpenTelemetry
        openTelemetry:
          host: otel-collector.monitoring.svc.cluster.local
          port: 4317
          protocol: grpc
```

+ The following is an example to disable prometheus metric and send metric via Open Telemetry sink to OTEL HTTP Collector at the same time.

``` yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyGateway
gateway:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
logging:
  level: null
  default: info
provider:
  type: Kubernetes
telemetry:
  metrics:
    prometheus:
      disable: false
    sinks:
      - type: OpenTelemetry
        openTelemetry:
          host: otel-collector.monitoring.svc.cluster.local
          port: 4318
          protocol: http
```

[OpenTelemetry]: https://opentelemetry.io/
[OpenTelemetry SDK]: https://opentelemetry.io/docs/specs/otel/metrics/sdk/
[Semantic Conventions]: https://opentelemetry.io/docs/concepts/semantic-conventions/
[Kubernetes specific ones]: https://opentelemetry.io/docs/specs/otel/resource/semantic_conventions/k8s/
[OTEL gRPC metrics exporter]: https://opentelemetry.io/docs/specs/otel/metrics/sdk_exporters/otlp/#general
[OTEL HTTP metrics exporter]: https://opentelemetry.io/docs/specs/otel/metrics/sdk_exporters/otlp/#general
[Kubernetes Metrics Lifecycle]: https://kubernetes.io/docs/concepts/cluster-administration/system-metrics/#metric-lifecycle
