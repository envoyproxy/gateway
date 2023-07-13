# Observability: Metrics

## Overview

Envoy provide robust platform for metrics, Envoy support three different kinds of stats: counter, gauges, histograms.

Envoy enables prometheus format output via the `/stats/prometheus` [admin endpoint](https://www.envoyproxy.io/docs/envoy/latest/operations/admin).

Envoy support different kinds of sinks, but EG will only support [Open Telemetry sink](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/stat_sinks/open_telemetry/v3/open_telemetry.proto).

Envoy Gateway leverages [Gateway API](https://gateway-api.sigs.k8s.io/) for configuring managed Envoy proxies. Gateway API defines core, extended, and implementation-specific API [support levels](https://gateway-api.sigs.k8s.io/concepts/conformance/?h=extended#2-support-levels) for implementors such as Envoy Gateway to expose features. Since metrics is not covered by `Core` or `Extended` APIs, EG should provide an easy to config metrics per `EnvoyProxy`.

## Goals

- Support expose metrics in prometheus way(reuse probe port).
- Support Open Telemetry stats sink.

## Non-Goals

- Support other stats sink.

## Use-Cases

- Enable prometheus metric
- Push metrics via Open Telemetry Sink
- Customize histogram buckets of target metric

### ProxyMetric API Type

```golang mdox-exec="sed '1,7d' api/config/v1alpha1/metric_types.go"
type ProxyMetrics struct {
	// Prometheus defines the configuration for Admin endpoint `/stats/prometheus`.
	Prometheus *PrometheusProvider `json:"prometheus,omitempty"`
	// Sinks defines the metric sinks where metrics are sent to.
	Sinks []MetricSink `json:"sinks,omitempty"`
	// HistogramBucketSettings defines rules for setting the histogram buckets.
	// Default buckets are used if not set. See more details at https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/metrics/v3/stats.proto.html#config-metrics-v3-histogrambucketsettings.
	HistogramBucketSettings []HistogramBucketSetting `json:"histogramBucketSettings,omitempty"`
}

type MetricSinkType string

const (
	MetricSinkTypeOpenTelemetry MetricSinkType = "OpenTelemetry"
)

type MetricSink struct {
	// Type defines the metric sink type.
	// EG currently only supports OpenTelemetry.
	// +kubebuilder:validation:Enum=OpenTelemetry
	// +kubebuilder:default=OpenTelemetry
	Type MetricSinkType `json:"type"`
	// OpenTelemetry defines the configuration for OpenTelemetry sink.
	// It's required if the sink type is OpenTelemetry.
	OpenTelemetry *OpenTelemetrySink `json:"openTelemetry,omitempty"`
}

type OpenTelemetrySink struct {
	// Host define the service hostname.
	Host string `json:"host"`
	// Port defines the port the service is exposed on.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:default=4317
	Port int32 `json:"port,omitempty"`

	// TODO: add support for customizing OpenTelemetry sink in https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/stat_sinks/open_telemetry/v3/open_telemetry.proto#envoy-v3-api-msg-extensions-stat-sinks-open-telemetry-v3-sinkconfig
}

type PrometheusProvider struct {
}

type HistogramBucketSetting struct {
	// Regex defines the regex for the stats name.
	// This use RE2 engine.
	// +kubebuilder:validation:Pattern=^/.*$
	// +kubebuilder:validation:MinLength=1
	Regex string `json:"regex"`
	// Buckets defines the buckets for the histogram.
	// +kubebuilder:validation:MinItems=1
	Buckets []float32 `json:"buckets"`
}
```

### Example

1. The following is an example to enable prometheus metric.

```yaml mdox-exec="sed '1,12d' examples/kubernetes/metric/prometheus.yaml"
apiVersion: config.gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: prometheus
  namespace: envoy-gateway-system
spec:
  telemetry:
    metric:
      prometheus: {}
```

1. The following is an example to send metric via Open Telemetry sink.

```yaml mdox-exec="sed '1,12d' examples/kubernetes/metric/otel-sink.yaml"
apiVersion: config.gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: otel-sink
  namespace: envoy-gateway-system
spec:
  telemetry:
    metric:
      sinks:
        - type: OpenTelemetry
          openTelemetry:
            host: otel-collector.monitoring.svc.cluster.local
            port: 4317
```

The following is an example to custom histogram bucket for metrics with prefix `downstream`.

```yaml mdox-exec="sed '1,12d' examples/kubernetes/metric/custom-buckets.yaml"
apiVersion: config.gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: prometheus
  namespace: envoy-gateway-system
spec:
  telemetry:
    metric:
      prometheus:
        enable: true
      histogramBucketSettings:
        - regex: downstream.*
          buckets:
            - 0.5
            - 1
            - 5
            - 10
            - 100
            - 1000
```
