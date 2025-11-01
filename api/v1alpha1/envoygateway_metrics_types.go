// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

// EnvoyGatewayMetrics defines control plane push/pull metrics configurations.
type EnvoyGatewayMetrics struct {
	// Sinks defines the metric sinks where metrics are sent to.
	Sinks []EnvoyGatewayMetricSink `json:"sinks,omitempty"`
	// Prometheus defines the configuration for prometheus endpoint.
	Prometheus *EnvoyGatewayPrometheusProvider `json:"prometheus,omitempty"`
}

// EnvoyGatewayTraces defines control plane tracing configurations.
type EnvoyGatewayTraces struct {
	// Sink defines the trace sink where traces are sent to.
	Sink EnvoyGatewayTraceSink `json:"sink,omitempty"`
	// Disable disables the traces.
	// TODO: implement disability
	Disable bool `json:"enable,omitempty"`
	// SamplingRate controls the rate at which traces are sampled.
	// Defaults to 1.0 (100% sampling). Valid values are between 0.0 and 1.0.
	// 0.0 means no sampling, 1.0 means all traces are sampled.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0.0
	// +kubebuilder:validation:Maximum=1.0
	SamplingRate *float64 `json:"samplingRate,omitempty"`
	// BatchSpanProcessorConfig defines the configuration for the batch span processor.
	// This processor batches spans before exporting them to the configured sink.
	//
	// +optional
	BatchSpanProcessorConfig *BatchSpanProcessorConfig `json:"batchSpanProcessor,omitempty"`
}

// BatchSpanProcessorConfig defines the configuration for the OpenTelemetry batch span processor.
// The batch span processor batches spans before sending them to the exporter.
type BatchSpanProcessorConfig struct {
	// BatchTimeout is the maximum duration for constructing a batch. Spans are
	// exported when either the batch is full or this timeout is reached.
	// Default is 5s. For e2e testing, a lower value like 100ms is recommended.
	//
	// +optional
	BatchTimeout *gwapiv1.Duration `json:"batchTimeout,omitempty"`
	// MaxExportBatchSize is the maximum number of spans to export in a single batch.
	// Default is 512.
	//
	// +optional
	// +kubebuilder:validation:Minimum=1
	MaxExportBatchSize *int `json:"maxExportBatchSize,omitempty"`
	// MaxQueueSize is the maximum queue size to buffer spans for delayed processing.
	// If the queue gets full it drops the spans. Default is 2048.
	//
	// +optional
	// +kubebuilder:validation:Minimum=1
	MaxQueueSize *int `json:"maxQueueSize,omitempty"`
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

// EnvoyGatewayTraceSink defines control plane
// trace sinks where traces are sent to.
type EnvoyGatewayTraceSink struct {
	// Type defines the trace sink type.
	// EG control plane currently supports OpenTelemetry.
	// +kubebuilder:validation:Enum=OpenTelemetry
	// +kubebuilder:default=OpenTelemetry
	Type TraceSinkType `json:"type"` // TODO: is this even needed?
	// OpenTelemetry defines the configuration for OpenTelemetry sink.
	// It's required if the sink type is OpenTelemetry.
	OpenTelemetry *EnvoyGatewayOpenTelemetrySink `json:"openTelemetry,omitempty"`
}

type EnvoyGatewayTracingSink struct {
	// Host define the sink service hostname.
	Host string `json:"host"`
	// Protocol define the sink service protocol.
	// +kubebuilder:validation:Enum=grpc;http
	Protocol string `json:"protocol"`
	// Port defines the port the sink service is exposed on.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=4319
	Port int32 `json:"port,omitempty"`
	// ExportInterval configures the intervening time between exports for a
	// Sink. This option overrides any value set for the
	// OTEL_METRIC_EXPORT_INTERVAL environment variable.
	// If ExportInterval is less than or equal to zero, 60 seconds
	// is used as the default.
	ExportInterval *gwapiv1.Duration `json:"exportInterval,omitempty"`
	// ExportTimeout configures the time a Sink waits for an export to
	// complete before canceling it. This option overrides any value set for the
	// OTEL_METRIC_EXPORT_TIMEOUT environment variable.
	// If ExportTimeout is less than or equal to zero, 30 seconds
	// is used as the default.
	ExportTimeout *gwapiv1.Duration `json:"exportTimeout,omitempty"`
	// TODO sampling rate
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
	// ExportInterval configures the intervening time between exports for a
	// Sink. This option overrides any value set for the
	// OTEL_METRIC_EXPORT_INTERVAL environment variable.
	// If ExportInterval is less than or equal to zero, 60 seconds
	// is used as the default.
	ExportInterval *gwapiv1.Duration `json:"exportInterval,omitempty"`
	// ExportTimeout configures the time a Sink waits for an export to
	// complete before canceling it. This option overrides any value set for the
	// OTEL_METRIC_EXPORT_TIMEOUT environment variable.
	// If ExportTimeout is less than or equal to zero, 30 seconds
	// is used as the default.
	ExportTimeout *gwapiv1.Duration `json:"exportTimeout,omitempty"`
}

// EnvoyGatewayPrometheusProvider will expose prometheus endpoint in pull mode.
type EnvoyGatewayPrometheusProvider struct {
	// Disable defines if disables the prometheus metrics in pull mode.
	//
	Disable bool `json:"disable,omitempty"`
}
