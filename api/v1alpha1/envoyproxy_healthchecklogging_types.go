// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

// ProxyHealthCheckLog configures Envoy health check event logging.
// Health check events (state transitions, failures, successes) are emitted
// to each configured sink.
//
// See the Envoy health check API reference for details on the underlying fields:
// https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/health_check.proto
//
// +kubebuilder:validation:XValidation:rule="size(self.sinks) == 1",message="exactly one sink must be specified"
type ProxyHealthCheckLog struct {
	// Sinks defines where health check events are written.
	// Exactly one sink must be specified.
	//
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=1
	Sinks []HealthCheckEventLogSink `json:"sinks"`

	// AlwaysLogHealthCheckFailures forces a log entry to be written for every
	// failed health check, regardless of the host's current health state.
	//
	// See: https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/health_check.proto#envoy-v3-api-field-config-core-v3-healthcheck-always-log-health-check-failures
	//
	// +optional
	AlwaysLogHealthCheckFailures *bool `json:"alwaysLogHealthCheckFailures,omitempty"`

	// AlwaysLogHealthCheckSuccess forces a log entry to be written for every
	// successful health check, regardless of the host's current health state.
	//
	// See: https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/health_check.proto#envoy-v3-api-field-config-core-v3-healthcheck-always-log-health-check-success
	//
	// +optional
	AlwaysLogHealthCheckSuccess *bool `json:"alwaysLogHealthCheckSuccess,omitempty"`
}

// HealthCheckEventLogSinkType is the type of health check event log sink.
// +kubebuilder:validation:Enum=File
type HealthCheckEventLogSinkType string

const (
	// HealthCheckEventLogSinkTypeFile writes health check events as JSON to a local file.
	HealthCheckEventLogSinkTypeFile HealthCheckEventLogSinkType = "File"
)

// HealthCheckEventLogSink defines a destination for health check event logs.
// +union
//
// +kubebuilder:validation:XValidation:rule="self.type == 'File' ? has(self.file) : !has(self.file)",message="If HealthCheckEventLogSink type is File, file field needs to be set."
type HealthCheckEventLogSink struct {
	// Type defines the type of sink.
	//
	// +kubebuilder:validation:Enum=File
	// +unionDiscriminator
	Type HealthCheckEventLogSinkType `json:"type"`

	// File defines the file sink configuration.
	// Required when type is File.
	//
	// +optional
	File *HealthCheckLoggingFileSink `json:"file,omitempty"`
}

// HealthCheckLoggingFileSink writes health check events as JSON to a local file path.
//
// See: https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/health_check/event_sinks/file/v3/file.proto
type HealthCheckLoggingFileSink struct {
	// Path specifies the file path for health check event output.
	// Use /dev/stdout to write to standard output.
	//
	// +kubebuilder:validation:MinLength=1
	Path string `json:"path"`
}
