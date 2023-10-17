// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

// EnvoyGatewayMetrics defines control plane push/pull metrics configurations.
type EnvoyGatewayMetrics struct {
	// Address defines the address of Envoy Gateway Metrics Server.
	Address *EnvoyGatewayMetricsAddress
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
	// Enable defines if enables the prometheus metrics in pull mode. Default is true.
	//
	// +optional
	// +kubebuilder:default=true
	Enable bool `json:"enable,omitempty"`
}

// EnvoyGatewayMetricsAddress defines the Envoy Gateway Metrics Address configuration.
type EnvoyGatewayMetricsAddress struct {
	// Port defines the port the metrics server is exposed on.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=19010
	Port int `json:"port,omitempty"`
	// Host defines the metrics server hostname.
	//
	// +optional
	// +kubebuilder:default="0.0.0.0"
	Host string `json:"host,omitempty"`
}
