// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

// EnvoyGatewayTraces defines control plane tracing configurations.
type EnvoyGatewayTraces struct {
	// Sink defines the trace sink where traces are sent to.
	Sink EnvoyGatewayTraceSink `json:"sink,omitempty"`
	// SamplingRate controls the fraction of traces that are sampled.
	// The value is expressed as a Gateway API Fraction (numerator/denominator).
	// If denominator is omitted, it defaults to 100.
	//
	// +optional
	SamplingRate *gwapiv1.Fraction `json:"samplingRate,omitempty"`
}

// TraceSinkType specifies the types of trace sinks supported by Envoy Gateway.
// +kubebuilder:validation:Enum=OpenTelemetry
type TraceSinkType string

const (
	// TraceSinkTypeOpenTelemetry captures traces for the OpenTelemetry sink.
	TraceSinkTypeOpenTelemetry TraceSinkType = "OpenTelemetry"
)

// EnvoyGatewayTraceSink defines control plane
// trace sinks where traces are sent to.
type EnvoyGatewayTraceSink struct {
	// Type defines the trace sink type.
	// EG control plane currently supports OpenTelemetry.
	// +kubebuilder:validation:Enum=OpenTelemetry
	// +kubebuilder:default=OpenTelemetry
	Type TraceSinkType `json:"type"`
	// OpenTelemetry defines the configuration for OpenTelemetry sink.
	// It's required if the sink type is OpenTelemetry.
	OpenTelemetry *EnvoyGatewayOpenTelemetrySink `json:"openTelemetry,omitempty"`
}
