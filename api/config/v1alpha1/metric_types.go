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
	// ProxyStatMatcher defines configuration for reporting custom Envoy stats.
	// To reduce memory and CPU overhead from Envoy stats system.
	ProxyStatsMatcher *StatsMatcher `json:"statsMatcher,omitempty"`
	// HistogramBucketSettings defines rules for setting the histogram buckets.
	// Default buckets are used if not set. See more details at
	// https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/metrics/v3/stats.proto.html#config-metrics-v3-histogrambucketsettings.
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

type StatsMatcher struct {
	// Proxy stats name prefix matcher for inclusion.
	InclusionPrefixes []string `json:"inclusionPrefixes,omitempty"`
	// Proxy stats name suffix matcher for inclusion.
	InclusionSuffixes []string `json:"inclusionSuffixes,omitempty"`
	// Proxy stats name regexps matcher for inclusion.
	InclusionRegexps []string `json:"inclusionRegexps,omitempty"`
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
