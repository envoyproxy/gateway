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
// and which attributes are sent to the processor
type ProcessingModeOptions struct {
	// Defines body processing mode
	//
	// +optional
	Body *ExtProcBodyProcessingMode `json:"body,omitempty"`

	// Defines which attributes are sent to the external processor
	// https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/advanced/attributes
	//
	// +optional
	Attributes []string `json:"attributes,omitempty"`
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
// +kubebuilder:validation:XValidation:message="BackendRefs must be used, backendRef is not supported.",rule="!has(self.backendRef)"
// +kubebuilder:validation:XValidation:message="BackendRefs only supports Service and Backend kind.",rule="has(self.backendRefs) ? self.backendRefs.all(f, f.kind == 'Service' || f.kind == 'Backend') : true"
// +kubebuilder:validation:XValidation:message="BackendRefs only supports Core and gateway.envoyproxy.io group.",rule="has(self.backendRefs) ? (self.backendRefs.all(f, f.group == \"\" || f.group == 'gateway.envoyproxy.io')) : true"
type ExtProc struct {
	BackendCluster `json:",inline"`

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

	// MetadataOptions defines options related to the sending and receiving of dynamic metadata.
	// These options define which metadata namespaces would be sent to the processor and which dynamic metadata
	// namespaces the processor would be permitted to emit metadata to.
	// Users can specify custom namespaces or well-known envoy metadata namespace (such as envoy.filters.http.ext_authz)
	// documented here: https://www.envoyproxy.io/docs/envoy/latest/configuration/advanced/well_known_dynamic_metadata#well-known-dynamic-metadata
	// Default: no metadata context is sent or received from the external processor
	//
	// +optional
	MetadataOptions *ExtProcMetadataOptions `json:"metadataOptions,omitempty"`
}

// ExtProcAttributes defines which envoy attributes are sent for requests and responses to the external processor
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

// ExtProcMetadataOptions defines options related to the sending and receiving of dynamic metadata to and from the
// external processor service
type ExtProcMetadataOptions struct {
	// metadata namespaces forwarded to external processor
	//
	// +optional
	ForwardingNamespaces []string `json:"forwardingNamespaces,omitempty"`

	// metadata namespaces updatable by external processor
	//
	// +optional
	ReceivingNamespaces []string `json:"receivingNamespaces,omitempty"`
}
