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
type ProxyHealthCheckLog struct {
	// Sinks defines where health check events are written.
	// When omitted, events are written to /dev/stdout.
	//
	// +kubebuilder:validation:MaxItems=1
	// +optional
	Sinks []ProxyHealthCheckLogSink `json:"sinks,omitempty"`

	// Matches defines which health check probe outcomes produce a log entry.
	// When omitted or empty, all events are logged.
	//
	// Each value must be unique. Multiple values are ORed. If any failure type is
	// specified then a success type must also be specified, and vice versa.
	//
	// +kubebuilder:validation:MaxItems=4
	// +kubebuilder:validation:XValidation:rule="self.exists(e, e == 'Failure' || e == 'FailureTransition') == self.exists(e, e == 'Success' || e == 'SuccessTransition')",message="a failure type and a success type must both be specified together"
	// +listType=set
	// +optional
	Matches []ProxyHealthCheckLogEventType `json:"matches,omitempty"`
}

// ProxyHealthCheckLogEventType specifies which health check probe outcomes produce a log entry.
//
// +kubebuilder:validation:Enum=Failure;FailureTransition;Success;SuccessTransition
type ProxyHealthCheckLogEventType string

const (
	// ProxyHealthCheckLogEventTypeFailure logs every failed probe regardless of
	// the host's current health state.
	ProxyHealthCheckLogEventTypeFailure ProxyHealthCheckLogEventType = "Failure"

	// ProxyHealthCheckLogEventTypeFailureTransition logs only when a host
	// transitions from healthy to unhealthy.
	ProxyHealthCheckLogEventTypeFailureTransition ProxyHealthCheckLogEventType = "FailureTransition"

	// ProxyHealthCheckLogEventTypeSuccess logs every successful probe regardless
	// of the host's current health state.
	ProxyHealthCheckLogEventTypeSuccess ProxyHealthCheckLogEventType = "Success"

	// ProxyHealthCheckLogEventTypeSuccessTransition logs only when a host
	// transitions from unhealthy to healthy.
	ProxyHealthCheckLogEventTypeSuccessTransition ProxyHealthCheckLogEventType = "SuccessTransition"
)

// ProxyHealthCheckLogSinkType is the type of a ProxyHealthCheckLog sink.
// +kubebuilder:validation:Enum=File
type ProxyHealthCheckLogSinkType string

const (
	// ProxyHealthCheckLogSinkTypeFile writes health check events as JSON to a local file.
	ProxyHealthCheckLogSinkTypeFile ProxyHealthCheckLogSinkType = "File"
)

// ProxyHealthCheckLogSink defines a destination for health check event logs.
// +union
//
// +kubebuilder:validation:XValidation:rule="self.type == 'File' ? has(self.file) : !has(self.file)",message="If ProxyHealthCheckLogSink type is File, file field needs to be set."
type ProxyHealthCheckLogSink struct {
	// Type defines the type of sink.
	//
	// +kubebuilder:validation:Enum=File
	// +unionDiscriminator
	Type ProxyHealthCheckLogSinkType `json:"type"`

	// File defines the file sink configuration.
	// Required when type is File.
	//
	// +optional
	File *FileEnvoyProxyHealthCheckLog `json:"file,omitempty"`
}

// FileEnvoyProxyHealthCheckLog writes health check events as JSON to a local file path.
//
// See: https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/health_check/event_sinks/file/v3/file.proto
type FileEnvoyProxyHealthCheckLog struct {
	// Path specifies the file path for health check event output.
	// Use /dev/stdout to write to standard output.
	//
	// +kubebuilder:validation:MinLength=1
	Path string `json:"path"`
}
