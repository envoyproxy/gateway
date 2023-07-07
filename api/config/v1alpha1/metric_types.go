// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

type ProxyMetric struct {
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
	// Backend defines the backend to send OpenTelemetry metrics to.
	// +kubebuilder:default={port: 3417}
	Backend BackendService `json:"backend"`

	// TODO: add support for customizing OpenTelemetry sink in https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/stat_sinks/open_telemetry/v3/open_telemetry.proto#envoy-v3-api-msg-extensions-stat-sinks-open-telemetry-v3-sinkconfig
}

type PrometheusProvider struct {
	// Enable defines whether to enable the Prometheus endpoint.
	// Prometheus' annotations will be added to pod if enabled.
	Enable bool `json:"enable,omitempty"`

	// TODO: add support for customizing scrape path, e.g. rewrite `/stats/prometheus` to `/metrics`?
}

type HistogramBucketSetting struct {
	// Regex defines the regex for the stats name.
	// This use RE2 engine.
	// +kubebuilder:validation:Pattern=^/.*$
	// +kubebuilder:validation:MinLength=1
	Regex string `json:"regex"`
	// Buckets defines the buckets for the histogram.
	// +kubebuilder:validation:MinItems=1
	Buckets []float64 `json:"buckets"`
}
