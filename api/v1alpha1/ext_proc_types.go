// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// +kubebuilder:validation:Enum=Streamed;Buffered;BufferedPartial;FullDuplexStreamed
type ExtProcBodyProcessingMode string

const (
	// StreamedExtProcBodyProcessingMode will stream the body to the server in pieces as they arrive at the proxy.
	StreamedExtProcBodyProcessingMode ExtProcBodyProcessingMode = "Streamed"
	// BufferedExtProcBodyProcessingMode will buffer the message body in memory and send the entire body at once. If the body exceeds the configured buffer limit, then the downstream system will receive an error.
	BufferedExtProcBodyProcessingMode ExtProcBodyProcessingMode = "Buffered"
	// FullDuplexStreamedExtBodyProcessingMode will send the body in pieces, to be read in a stream. When enabled, trailers are also sent, and failOpen must be false.
	// Full details here: https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/ext_proc/v3/processing_mode.proto.html#enum-extensions-filters-http-ext-proc-v3-processingmode-bodysendmode
	FullDuplexStreamedExtBodyProcessingMode ExtProcBodyProcessingMode = "FullDuplexStreamed"
	// BufferedPartialExtBodyHeaderProcessingMode will buffer the message body in memory and send the entire body in one chunk. If the body exceeds the configured buffer limit, then the body contents up to the buffer limit will be sent.
	BufferedPartialExtBodyHeaderProcessingMode ExtProcBodyProcessingMode = "BufferedPartial"
)

// +kubebuilder:validation:Enum=Route;Backend;All
type ExtProcStage string

const (
	// ExtProcStageRoute configures ExtProc to execute at the route phase (downstream, listener HCM filter)
	ExtProcStageRoute ExtProcStage = "Route"
	// ExtProcStageBackend configures ExtProc to execute at the backend phase (upstream, cluster filter)
	ExtProcStageBackend ExtProcStage = "Backend"
	// ExtProcStageAll configures ExtProc to execute at both route and backend phases
	ExtProcStageAll ExtProcStage = "All"
)

// ProcessingModeOptions defines if headers or body should be processed by the external service
// and which attributes are sent to the processor
type ProcessingModeOptions struct {
	// Defines body processing mode
	//
	// +optional
	Body *ExtProcBodyProcessingMode `json:"body,omitempty"`

	// Defines which attributes are sent to the external processor. Envoy Gateway currently
	// supports only the following attribute prefixes: connection, source, destination,
	// request, response, upstream and xds.route.
	// https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/advanced/attributes
	//
	// +optional
	// +kubebuilder:validation:items:Pattern=`^(connection\.|source\.|destination\.|request\.|response\.|upstream\.|xds\.route_)[a-z_1-9]*$`
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

	// AllowModeOverride allows the external processor to override the processing mode set via the
	// `mode_override` field in the gRPC response message. This defaults to false.
	//
	// +optional
	AllowModeOverride bool `json:"allowModeOverride,omitempty"`
}

// ExtProc defines the configuration for External Processing filter.
// +kubebuilder:validation:XValidation:message="BackendRefs must be used, backendRef is not supported.",rule="!has(self.backendRef)"
// +kubebuilder:validation:XValidation:message="BackendRefs only supports Service and Backend kind.",rule="has(self.backendRefs) ? self.backendRefs.all(f, f.kind == 'Service' || f.kind == 'Backend') : true"
// +kubebuilder:validation:XValidation:message="BackendRefs only supports Core and gateway.envoyproxy.io group.",rule="has(self.backendRefs) ? (self.backendRefs.all(f, f.group == \"\" || f.group == 'gateway.envoyproxy.io')) : true"
// +kubebuilder:validation:XValidation:message="If FullDuplexStreamed body processing mode is used, FailOpen must be false.",rule="!(has(self.failOpen) && self.failOpen == true && ((has(self.processingMode.request.body) && self.processingMode.request.body == 'FullDuplexStreamed') || (has(self.processingMode.response.body) && self.processingMode.response.body == 'FullDuplexStreamed')))"
type ExtProc struct {
	BackendCluster `json:",inline"`

	// Stage defines at which stage of request processing the ExtProc filter is executed.
	// Default: Route
	//
	// +optional
	// +kubebuilder:default=Route
	Stage *ExtProcStage `json:"stage,omitempty"`

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

	// Metadata defines options related to the sending and receiving of dynamic metadata.
	// These options define which metadata namespaces would be sent to the processor and which dynamic metadata
	// namespaces the processor would be permitted to emit metadata to.
	// Users can specify custom namespaces or well-known envoy metadata namespace (such as envoy.filters.http.ext_authz)
	// documented here: https://www.envoyproxy.io/docs/envoy/latest/configuration/advanced/well_known_dynamic_metadata#well-known-dynamic-metadata
	// Default: no metadata context is sent or received from the external processor
	//
	// +optional
	Metadata *ExtProcMetadata `json:"metadata,omitempty"`
}

// ExtProcMetadata defines options related to the sending and receiving of dynamic metadata to and from the
// external processor service
type ExtProcMetadata struct {
	// AccessibleNamespaces are metadata namespaces that are sent to the external processor as context
	//
	// +optional
	AccessibleNamespaces []string `json:"accessibleNamespaces,omitempty"`

	// WritableNamespaces are metadata namespaces that the external processor can write to
	//
	// +kubebuilder:validation:XValidation:rule="self.all(f, !f.startsWith('envoy.filters.http'))",message="writableNamespaces cannot contain well-known Envoy HTTP filter namespaces"
	// +kubebuilder:validation:MaxItems=8
	// +optional
	WritableNamespaces []string `json:"writableNamespaces,omitempty"`
}
