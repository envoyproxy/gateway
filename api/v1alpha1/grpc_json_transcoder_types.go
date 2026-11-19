// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// GRPCJSONTranscoder defines the configuration for gRPC-JSON transcoding.
type GRPCJSONTranscoder struct {
	// ProtoDescriptor defines where to find the protobuf descriptor set of the gRPC services.
	ProtoDescriptor ProtoDescriptor `json:"protoDescriptor"`

	// Services defines the gRPC services to transcode.
	//
	// If not specified, all the services defined in the descriptor are transcoded,
	// except the ones that come from imported files.
	//
	// +optional
	// +listType=set
	Services []string `json:"services,omitempty"`

	// PrintOptions defines how the gRPC response is printed as JSON.
	//
	// +optional
	PrintOptions *JSONPrintOptions `json:"printOptions,omitempty"`

	// IgnoreUnknownQueryParameters ignores the query parameters that can't be mapped to
	// a field of the gRPC request.
	//
	// If false, a request with such a parameter is not transcoded and is sent to the
	// backend as is.
	//
	// +optional
	IgnoreUnknownQueryParameters *bool `json:"ignoreUnknownQueryParameters,omitempty"`

	// ConvertGRPCStatus converts the gRPC status of an error response to the HTTP
	// status code and a JSON body.
	//
	// +optional
	ConvertGRPCStatus *bool `json:"convertGRPCStatus,omitempty"`

	// BodySize defines the maximum size of the request and response bodies to transcode.
	//
	// +optional
	BodySize *GRPCJSONTranscoderBodySize `json:"bodySize,omitempty"`
}

// GRPCJSONTranscoderBodySize defines the maximum size of the bodies to transcode.
type GRPCJSONTranscoderBodySize struct {
	// MaxRequest is the maximum size of a request body to transcode. A larger
	// body is rejected with HTTP 413.
	// If not set, the buffer limit of the stream is used, which is 32Ki for HTTP/1.1 and
	// 64Ki for HTTP/2 by default.
	// For example, 1Mi, 64Mi etc.
	// Note that when the suffix is not provided, the value is interpreted as bytes.
	//
	// +kubebuilder:validation:XIntOrString
	// +kubebuilder:validation:Pattern="^[1-9]+[0-9]*([EPTGMK]i|[EPTGMk])?$"
	// +optional
	MaxRequest *resource.Quantity `json:"maxRequest,omitempty"`

	// MaxResponse is the maximum size of a response body to transcode. A larger
	// body is replaced with HTTP 500, or resets the stream of a streaming response.
	// If not set, the buffer limit of the stream is used, which is 32Ki for HTTP/1.1 and
	// 64Ki for HTTP/2 by default.
	// For example, 1Mi, 64Mi etc.
	// Note that when the suffix is not provided, the value is interpreted as bytes.
	//
	// +kubebuilder:validation:XIntOrString
	// +kubebuilder:validation:Pattern="^[1-9]+[0-9]*([EPTGMK]i|[EPTGMk])?$"
	// +optional
	MaxResponse *resource.Quantity `json:"maxResponse,omitempty"`
}

// ProtoDescriptor defines where to find the protobuf descriptor set.
// +kubebuilder:validation:XValidation:rule="size(self.valueRef.group) == 0 && self.valueRef.kind == 'ConfigMap'",message="valueRef must refer to a core ConfigMap"
type ProtoDescriptor struct {
	// ValueRef references a ConfigMap in the same namespace that holds the descriptor set.
	//
	// The descriptor set is read from the `proto-descriptor` key, either in `binaryData`
	// or in `data`, where it can be base64-encoded. If the ConfigMap has no `data` and only
	// one `binaryData` entry, that entry is used.
	//
	// The descriptor set must include all the imported files. It can be generated with
	// `protoc --include_imports --descriptor_set_out=...`.
	ValueRef gwapiv1.LocalObjectReference `json:"valueRef"`
}

// JSONPrintOptions defines how the gRPC response is printed as JSON.
// +kubebuilder:validation:XValidation:rule="!(has(self.streamNewlineDelimited) && self.streamNewlineDelimited && has(self.streamSSEStyleDelimited) && self.streamSSEStyleDelimited)",message="streamNewlineDelimited and streamSSEStyleDelimited are mutually exclusive"
type JSONPrintOptions struct {
	// AddWhitespace adds spaces, line breaks and indentation to the JSON output.
	//
	// +optional
	AddWhitespace *bool `json:"addWhitespace,omitempty"`

	// AlwaysPrintPrimitiveFields prints the primitive fields even when they have
	// default values.
	//
	// +optional
	AlwaysPrintPrimitiveFields *bool `json:"alwaysPrintPrimitiveFields,omitempty"`

	// PreserveProtoFieldNames uses the field names in the proto file instead of
	// lowerCamelCase names.
	//
	// +optional
	PreserveProtoFieldNames *bool `json:"preserveProtoFieldNames,omitempty"`

	// AlwaysPrintEnumsAsInts prints enums as integers instead of names.
	//
	// +optional
	AlwaysPrintEnumsAsInts *bool `json:"alwaysPrintEnumsAsInts,omitempty"`

	// StreamNewlineDelimited prints a streaming response as newline-delimited JSON
	// messages instead of a JSON array.
	//
	// Mutually exclusive with StreamSSEStyleDelimited.
	//
	// +optional
	StreamNewlineDelimited *bool `json:"streamNewlineDelimited,omitempty"`

	// StreamSSEStyleDelimited prints a streaming response as Server-Sent Events, with
	// each message framed as `data: <message>`.
	//
	// Mutually exclusive with StreamNewlineDelimited.
	//
	// +optional
	StreamSSEStyleDelimited *bool `json:"streamSSEStyleDelimited,omitempty"`
}
