// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

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

// ProxyMetricSink defines the sink of metrics.
// Default metrics sink is OpenTelemetry.
// +union
//
// +kubebuilder:validation:XValidation:rule="self.type == 'OpenTelemetry' ? has(self.openTelemetry) : !has(self.openTelemetry)",message="If MetricSink type is OpenTelemetry, openTelemetry field needs to be set."
type ProxyMetricSink struct {
	// Type defines the metric sink type.
	// EG currently only supports OpenTelemetry.
	// +kubebuilder:validation:Enum=OpenTelemetry
	// +kubebuilder:default=OpenTelemetry
	// +unionDiscriminator
	Type MetricSinkType `json:"type"`
	// OpenTelemetry defines the configuration for OpenTelemetry sink.
	// It's required if the sink type is OpenTelemetry.
	// +optional
	OpenTelemetry *ProxyOpenTelemetrySink `json:"openTelemetry,omitempty"`
}

// ProxyOpenTelemetrySink defines the configuration for OpenTelemetry sink.
//
// +kubebuilder:validation:XValidation:message="BackendRef only support Service Kind.",rule="!has(self.backendRef) || !has(self.backendRef.kind) || self.backendRef.kind == 'Service'"
type ProxyOpenTelemetrySink struct {
	// Host define the service hostname.
	// Deprecated: Use BackendRef instead.
	Host string `json:"host"`
	// Port defines the port the service is exposed on.
	// Deprecated: Use BackendRef instead.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:default=4317
	Port int32 `json:"port,omitempty"`
	// BackendRef references a Kubernetes object that represents the
	// backend server to which the metric will be sent.
	// Only service Kind is supported for now.
	//
	// +optional
	BackendRef *gwapiv1.BackendObjectReference `json:"backendRef,omitempty"`

	// TODO: add support for customizing OpenTelemetry sink in https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/stat_sinks/open_telemetry/v3/open_telemetry.proto#envoy-v3-api-msg-extensions-stat-sinks-open-telemetry-v3-sinkconfig
}

type ProxyPrometheusProvider struct {
	// Disable the Prometheus endpoint.
	Disable bool `json:"disable,omitempty"`
}
