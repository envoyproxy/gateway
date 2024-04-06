// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

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
	// ExporterInterval configures the intervening time between exports for a
	// Sink. This option overrides any value set for the
	// OTEL_METRIC_EXPORT_INTERVAL environment variable.
	// If ExporterInterval is less than or equal to zero, 60 seconds
	// is used as the default.
	ExporterInterval *ExporterDuration `json:"exporterInterval"`
	// ExporterTimeout configures the time a Sink waits for an export to
	// complete before canceling it. This option overrides any value set for the
	// OTEL_METRIC_EXPORT_TIMEOUT environment variable.
	// If ExporterTimeout is less than or equal to zero, 30 seconds
	// is used as the default.
	ExporterTimeout *ExporterDuration `json:"exporterTimeout"`
}

// ExporterDuration is used to represent the duration of the exporter.
// It is used to represent duration as a string and parsed by time.ParseDuration function.
// This type uses the time format in Go.
// example: "1h30m" means a duration of 1 hour and 30 minutes, "10s" means a duration of 10 seconds.
type ExporterDuration string

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
