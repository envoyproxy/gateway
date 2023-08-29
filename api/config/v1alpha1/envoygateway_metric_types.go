// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

type EnvoyGatewayMetrics struct {
	// Sinks defines the metric sinks where metrics are sent to.
	Sinks []EnvoyGatewayMetricSink `json:"sinks,omitempty"`
	// Matches defines configuration for selecting specific metrics instead of generating all metrics stats
	// that are enabled by default. This helps reduce CPU and memory overhead in Envoy Gateway.
	Matches []Match `json:"matches,omitempty"`
}

type EnvoyGatewayMetricSink struct {
	// Type defines the metric sink type.
	// EG currently supports OpenTelemetry.
	// +kubebuilder:validation:Enum=OpenTelemetry
	// +kubebuilder:default=OpenTelemetry
	Type MetricSinkType `json:"type"`
	// Host define the sink service hostname.
	Host string `json:"host"`
	// Port defines the port the sink service is exposed on.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=4317
	Port int32 `json:"port,omitempty"`
}
