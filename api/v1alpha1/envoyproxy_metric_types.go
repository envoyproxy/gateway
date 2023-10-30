// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

type MetricSinkType string

const (
	MetricSinkTypeOpenTelemetry MetricSinkType = "OpenTelemetry"
)

type ProxyMetrics struct {
	// Prometheus defines the configuration for Admin endpoint `/stats/prometheus`.
	Prometheus *ProxyPrometheusProvider `json:"prometheus,omitempty"`
	// Sinks defines the metric sinks where metrics are sent to.
	Sinks []ProxyMetricSink `json:"sinks,omitempty"`
	// Matches defines configuration for selecting specific metrics instead of generating all metrics stats
	// that are enabled by default. This helps reduce CPU and memory overhead in Envoy, but eliminating some stats
	// may after critical functionality. Here are the stats that we strongly recommend not disabling:
	// `cluster_manager.warming_clusters`, `cluster.<cluster_name>.membership_total`,`cluster.<cluster_name>.membership_healthy`,
	// `cluster.<cluster_name>.membership_degraded`，reference  https://github.com/envoyproxy/envoy/issues/9856,
	// https://github.com/envoyproxy/envoy/issues/14610
	//
	Matches []StringMatch `json:"matches,omitempty"`

	// EnableVirtualHostStats enables envoy stat metrics for virtual hosts.
	EnableVirtualHostStats bool `json:"enableVirtualHostStats,omitempty"`
}

type ProxyMetricSink struct {
	// Type defines the metric sink type.
	// EG currently only supports OpenTelemetry.
	// +kubebuilder:validation:Enum=OpenTelemetry
	// +kubebuilder:default=OpenTelemetry
	Type MetricSinkType `json:"type"`
	// OpenTelemetry defines the configuration for OpenTelemetry sink.
	// It's required if the sink type is OpenTelemetry.
	OpenTelemetry *ProxyOpenTelemetrySink `json:"openTelemetry,omitempty"`
}

type ProxyOpenTelemetrySink struct {
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

type ProxyPrometheusProvider struct {
	// Disable the Prometheus endpoint.
	Disable bool `json:"disable,omitempty"`
}
