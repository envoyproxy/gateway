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
	// Disable disables the traces.
	//
	// +optional
	Disable bool `json:"disable,omitempty"`
	// SamplingRate controls the rate at which traces are sampled.
	// Defaults to 1.0 (100% sampling). Valid values are between 0.0 and 1.0.
	// 0.0 means no sampling, 1.0 means all traces are sampled.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0.0
	// +kubebuilder:validation:Maximum=1.0
	SamplingRate *float64 `json:"samplingRate,omitempty"`
	// BatchSpanProcessorConfig defines the configuration for the batch span processor.
	// This processor batches spans before exporting them to the configured sink.
	//
	// +optional
	BatchSpanProcessorConfig *BatchSpanProcessorConfig `json:"batchSpanProcessor,omitempty"`
}

// BatchSpanProcessorConfig defines the configuration for the OpenTelemetry batch span processor.
// The batch span processor batches spans before sending them to the exporter.
type BatchSpanProcessorConfig struct {
	// BatchTimeout is the maximum duration for constructing a batch. Spans are
	// exported when either the batch is full or this timeout is reached.
	//
	// +optional
	BatchTimeout *gwapiv1.Duration `json:"batchTimeout,omitempty"`
	// MaxExportBatchSize is the maximum number of spans to export in a single batch.
	// Default is 512.
	//
	// +optional
	// +kubebuilder:validation:Minimum=1
	MaxExportBatchSize *int `json:"maxExportBatchSize,omitempty"`
	// MaxQueueSize is the maximum queue size to buffer spans for delayed processing.
	// If the queue gets full it drops the spans. Default is 2048.
	//
	// +optional
	// +kubebuilder:validation:Minimum=1
	MaxQueueSize *int `json:"maxQueueSize,omitempty"`
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
