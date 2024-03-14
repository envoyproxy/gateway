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

type ProcessingModeOptions struct {
	// Defines header processing mode
	//
	// +kubebuilder:default:=Send
	// +optional
	Headers *ExtProcHeaderProcessingMode `json:"header,omitempty"`
	// Defines body processing mode
	//
	// +kubebuilder:default:=None
	// +optional
	Body *ExtProcBodyProcessingMode `json:"body,omitempty"`
}

// ExtProcProcessingMode defines if and how headers and bodies are sent to the service.
// https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/ext_proc/v3/processing_mode.proto#envoy-v3-api-msg-extensions-filters-http-ext-proc-v3-processingmode
type ExtProcProcessingMode struct {
	// Defines header and body treatment for requests
	//
	// +optional
	Request *ProcessingModeOptions `json:"request,omitempty"`
	// Defines header and body treatment for responses
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
	// Specifies a list of metadata namespaces whose values, if present, will be passed to the ext_proc service as an opaque protobuf::Struct.
	//
	// +optional
	Untyped []string `json:"untyped,omitempty"`
	// Specifies a list of metadata namespaces whose values, if present, will be passed to the ext_proc service as a protobuf::Any.
	//
	// +optional
	Typed []string `json:"typed,omitempty"`
}

// ExtProcMetadataOptions defines options related to the sending and receiving of dynamic metadata
type ExtProcMetadataOptions struct {
	// metadata namespaces forwarded to external processor
	//
	// +optional
	ForwardingNamespaces []MetadataNamespaces `json:"forwardingNamespaces,omitempty"`
	// metadata namespaces updatable by external processor
	//
	// +optional
	ReceivingNamespaces []MetadataNamespaces `json:"receivingNamespaces,omitempty"`
}

// +kubebuilder:validation:XValidation:rule="has(self.service) ? (!has(self.service.backendRef.group) || self.service.backendRef.group == \"\") : true", message="group is invalid, only the core API group (specified by omitting the group field or setting it to an empty string) is supported"
// +kubebuilder:validation:XValidation:rule="has(self.service) ? (!has(self.service.backendRef.kind) || self.service.backendRef.kind == 'Service') : true", message="kind is invalid, only Service (specified by omitting the kind field or setting it to 'Service') is supported"
//
// ExtProc defines the configuration for External Processing.
type ExtProc struct {
	// Service defines the configuration of the external processing service
	Service ExtProcService `json:"service"`
	// ProcessingMode defines how request and response headers and body are processed
	// Default: request and response headers are sent, bodies are not sent
	//
	// +optional
	ProcessingMode *ExtProcProcessingMode `json:"processingMode,omitempty"`
	// Attributes defines which envoy request and response attributes are provided as context to external processor
	// Default: no attributes are sent
	//
	// +optional
	Attributes *ExtProcAttributes `json:"attributes,omitempty"`
	// MetadataOptions defines options related to the sending and receiving of dynamic metadata
	// Default: no metadata context is sent or received
	//
	// +optional
	MetadataOptions *ExtProcMetadataOptions `json:"metadataOptions,omitempty"`
	// The timeout for a response to be returned from the external processor
	// Default: 200ms
	//
	// +optional
	MessageTimeout *gwapiv1.Duration `json:"messageTimeout,omitempty"`
}

// ExtProcService defines the gRPC External Processing service using the envoy grpc client
// The processing request and response messages are defined in
// https://www.envoyproxy.io/docs/envoy/latest/api-v3/service/ext_proc/v3/external_processor.proto
type ExtProcService struct {
	// BackendObjectReference references a Kubernetes object that represents the
	// backend server to which the processing requests will be sent.
	// Only service Kind is supported for now.
	BackendRef gwapiv1.BackendObjectReference `json:"backendRef"`
}
