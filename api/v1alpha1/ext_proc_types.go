// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// +kubebuilder:validation:Enum=Streamed;Buffered;BufferedPartial
type ExtProcBodyProcessingMode string

const (
	StreamedExtProcBodyProcessingMode          ExtProcBodyProcessingMode = "Streamed"
	BufferedExtProcBodyProcessingMode          ExtProcBodyProcessingMode = "Buffered"
	BufferedPartialExtBodyHeaderProcessingMode ExtProcBodyProcessingMode = "BufferedPartial"
)

// ProcessingModeOptions defines if headers or body should be processed by the external service
type ProcessingModeOptions struct {
	// Defines body processing mode
	//
	// +optional
	Body *ExtProcBodyProcessingMode `json:"body,omitempty"`
}

// ExtProcProcessingMode defines if and how headers and bodies are sent to the service.
// https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/ext_proc/v3/processing_mode.proto#envoy-v3-api-msg-extensions-filters-http-ext-proc-v3-processingmode
type ExtProcProcessingMode struct {
	// Defines processing mode for requests. If present, request headers are sent. Request body is processed according
	// to the specified mode.
	//
	// +optional
	Request *ProcessingModeOptions `json:"request,omitempty"`

	// Defines processing mode for responses. If present, response headers are sent. Response body is processed according
	// to the specified mode.
	//
	// +optional
	Response *ProcessingModeOptions `json:"response,omitempty"`
}

// +kubebuilder:validation:XValidation:rule="has(self.backendRef) ? (!has(self.backendRef.group) || self.backendRef.group == \"\") : true", message="group is invalid, only the core API group (specified by omitting the group field or setting it to an empty string) is supported"
// +kubebuilder:validation:XValidation:rule="has(self.backendRef) ? (!has(self.backendRef.kind) || self.backendRef.kind == 'Service') : true", message="kind is invalid, only Service (specified by omitting the kind field or setting it to 'Service') is supported"
//
// ExtProc defines the configuration for External Processing filter.
type ExtProc struct {
	// BackendRef defines the configuration of the external processing service
	BackendRef ExtProcBackendRef `json:"backendRef"`

	// ProcessingMode defines how request and response body is processed
	// Default: header and body are not sent to the external processor
	//
	// +optional
	ProcessingMode *ExtProcProcessingMode `json:"processingMode,omitempty"`
}

// ExtProcService defines the gRPC External Processing service using the envoy grpc client
// The processing request and response messages are defined in
// https://www.envoyproxy.io/docs/envoy/latest/api-v3/service/ext_proc/v3/external_processor.proto
type ExtProcBackendRef struct {
	// BackendObjectReference references a Kubernetes object that represents the
	// backend server to which the processing requests will be sent.
	// Only service Kind is supported for now.
	gwapiv1.BackendObjectReference `json:",inline"`
}
