// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

type ProxyMetrics struct {
	// Prometheus defines the configuration for Admin endpoint `/stats/prometheus`.
	Prometheus *PrometheusProvider `json:"prometheus,omitempty"`
	// Sinks defines the metric sinks where metrics are sent to.
	Sinks []MetricSink `json:"sinks,omitempty"`
	// Matches defines configuration for metric matching rules
	// to reduce memory and CPU overhead from Envoy stats system.
	// When Matches is nil(not set), gateway by default create and expose only a subset of
	// Envoy stats with prefixes(cluster_manager,listener_manager,server,cluster.xds-grpc)
	// expressions match on the name of stats.
	// When Matches is not empty, this option is to control creation of additional Envoy stats
	// with prefix, suffix, and regex expressions match on the name of the stats.
	Matches []Match `json:"matches,omitempty"`
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

type Match struct {
	Type  MatcherType `json:"type"`
	Value string      `json:"value"`
}

type MatcherType string

const (
	Prefix MatcherType = "Prefix"
	Regex  MatcherType = "Regex"
	Suffix MatcherType = "Suffix"
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
