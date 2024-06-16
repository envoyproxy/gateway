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
	// StreamedExtProcBodyProcessingMode will stream the body to the server in pieces as they arrive at the proxy.
	StreamedExtProcBodyProcessingMode ExtProcBodyProcessingMode = "Streamed"
	// BufferedExtProcBodyProcessingMode will buffer the message body in memory and send the entire body at once. If the body exceeds the configured buffer limit, then the downstream system will receive an error.
	BufferedExtProcBodyProcessingMode ExtProcBodyProcessingMode = "Buffered"
	// BufferedPartialExtBodyHeaderProcessingMode will buffer the message body in memory and send the entire body in one chunk. If the body exceeds the configured buffer limit, then the body contents up to the buffer limit will be sent.
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

// ExtProc defines the configuration for External Processing filter.
type ExtProc struct {
	// BackendRefs defines the configuration of the external processing service
	//
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=1
	// +kubebuilder:validation:XValidation:message="BackendRefs only supports Service and Backend kind.",rule="self.all(f, f.kind == 'Service' || f.kind == 'Backend')"
	// +kubebuilder:validation:XValidation:message="BackendRefs only supports Core and gateway.envoyproxy.io group.",rule="self.all(f, f.group == '' || f.group == 'gateway.envoyproxy.io')"
	BackendRefs []BackendRef `json:"backendRefs"`

	// MessageTimeout is the timeout for a response to be returned from the external processor
	// Default: 200ms
	//
	// +optional
	MessageTimeout *gwapiv1.Duration `json:"messageTimeout,omitempty"`

	// FailOpen defines if requests or responses that cannot be processed due to connectivity to the
	// external processor are terminated or passed-through.
	// Default: false
	//
	// +optional
	FailOpen *bool `json:"failOpen,omitempty"`

	// ProcessingMode defines how request and response body is processed
	// Default: header and body are not sent to the external processor
	//
	// +optional
	ProcessingMode *ExtProcProcessingMode `json:"processingMode,omitempty"`
}
