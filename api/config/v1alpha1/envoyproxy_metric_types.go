// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

type ProxyMetrics struct {
	// Prometheus defines the configuration for Admin endpoint `/stats/prometheus`.
	Prometheus *PrometheusProvider `json:"prometheus,omitempty"`
	// Sinks defines the metric sinks where metrics are sent to.
	Sinks []ProxyMetricSink `json:"sinks,omitempty"`
	// Matches defines configuration for selecting specific metrics instead of generating all metrics stats
	// that are enabled by default. This helps reduce CPU and memory overhead in Envoy.
	Matches []Match `json:"matches,omitempty"`
}

type MetricSinkType string

const (
	MetricSinkTypeOpenTelemetry MetricSinkType = "OpenTelemetry"
)

type ProxyMetricSink struct {
	// Type defines the metric sink type.
	// EG currently only supports OpenTelemetry.
	// +kubebuilder:validation:Enum=OpenTelemetry
	// +kubebuilder:default=OpenTelemetry
	Type MetricSinkType `json:"type"`
	// OpenTelemetry defines the configuration for OpenTelemetry sink.
	// It's required if the sink type is OpenTelemetry.
	OpenTelemetry *OpenTelemetrySink `json:"openTelemetry,omitempty"`
}

// Match defines the stats match configuration.
type Match struct {
	// MatcherType defines the stats matcher type
	//
	// +kubebuilder:validation:Enum=RegularExpression;Prefix;Suffix
	Type  MatcherType `json:"type"`
	Value string      `json:"value"`
}

type MatcherType string

const (
	Prefix            MatcherType = "Prefix"
	RegularExpression MatcherType = "RegularExpression"
	Suffix            MatcherType = "Suffix"
)

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
