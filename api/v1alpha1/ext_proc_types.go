// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// +kubebuilder:validation:Enum=Default;Send;Skip
type ExtProcHeaderProcessingMode string

const (
	DefaultExtProcHeaderProcessingMode ExtProcHeaderProcessingMode = "Default"
	SendExtProcHeaderProcessingMode    ExtProcHeaderProcessingMode = "Send"
	SkipExtProcHeaderProcessingMode    ExtProcHeaderProcessingMode = "Skip"
)

// +kubebuilder:validation:Enum=None;Streamed;Buffered;BufferedPartial
type ExtProcBodyProcessingMode string

const (
	NoneExtProcHeaderProcessingMode            ExtProcBodyProcessingMode = "None"
	StreamedExtProcHeaderProcessingMode        ExtProcBodyProcessingMode = "Streamed"
	BufferedExtProcHeaderProcessingMode        ExtProcBodyProcessingMode = "Buffered"
	BufferedPartialExtProcHeaderProcessingMode ExtProcBodyProcessingMode = "BufferedPartial"
)

// ProcessingModeOptions defines if headers or body should be processed by the external service
type ProcessingModeOptions struct {
	// Defines header processing mode
	//
	// +optional
	Headers *ExtProcHeaderProcessingMode `json:"headers,omitempty"`

	// Defines body processing mode
	//
	// +optional
	Body *ExtProcBodyProcessingMode `json:"body,omitempty"`
}

// ExtProcProcessingMode defines if and how headers and bodies are sent to the service.
// https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/ext_proc/v3/processing_mode.proto#envoy-v3-api-msg-extensions-filters-http-ext-proc-v3-processingmode
type ExtProcProcessingMode struct {
	// Defines header and body processing for requests
	//
	// +optional
	Request *ProcessingModeOptions `json:"request,omitempty"`

	// Defines header and body processing for responses
	//
	// +optional
	Response *ProcessingModeOptions `json:"response,omitempty"`
}

// ExtProcAttributes defines which attributes are
type ExtProcAttributes struct {
	// defines attributes to send for Request processing
	//
	// +optional
	Request []string `json:"request,omitempty"`

	// defines attributes to send for Response processing
	//
	// +optional
	Response []string `json:"response,omitempty"`
}

// MetadataNamespaces defines metadata namespaces that can be used to forward or receive dynamic metadata
type MetadataNamespaces struct {
	// Specifies a list of metadata namespaces whose values, if present, will be passed to the ext_proc service
	// as an opaque protobuf::Struct.
	//
	// +optional
	Untyped []string `json:"untyped,omitempty"`
}

// ExtProcMetadataOptions defines options related to the sending and receiving of dynamic metadata to and from the
// external processor service
type ExtProcMetadataOptions struct {
	// metadata namespaces forwarded to external processor
	//
	// +optional
	ForwardingNamespaces MetadataNamespaces `json:"forwardingNamespaces,omitempty"`

	// metadata namespaces updatable by external processor
	//
	// +optional
	ReceivingNamespaces MetadataNamespaces `json:"receivingNamespaces,omitempty"`
}

// +kubebuilder:validation:XValidation:rule="has(self.backendRef) ? (!has(self.backendRef.group) || self.backendRef.group == \"\") : true", message="group is invalid, only the core API group (specified by omitting the group field or setting it to an empty string) is supported"
// +kubebuilder:validation:XValidation:rule="has(self.backendRef) ? (!has(self.backendRef.kind) || self.backendRef.kind == 'Service') : true", message="kind is invalid, only Service (specified by omitting the kind field or setting it to 'Service') is supported"
//
// ExtProc defines the configuration for External Processing filter.
type ExtProc struct {
	// BackendRef defines the configuration of the external processing service
	BackendRef ExtProcBackendRef `json:"backendRef"`

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
