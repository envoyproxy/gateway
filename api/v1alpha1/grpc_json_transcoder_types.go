// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

// GRPCJSONTranscoder defines the configuration for gRPC-JSON transcoding.
type GRPCJSONTranscoder struct {
	// ProtoDescriptor defines how to obtain the protocol buffer descriptor set.
	// This is required for the transcoder to understand the gRPC service definition.
	ProtoDescriptor ProtoDescriptor `json:"protoDescriptor"`

	// Services defines the gRPC services that should be transcoded.
	//
	// If not specified, every service declared by the proto files in the descriptor is
	// transcoded, excluding services that come from imported files.
	//
	// Entries must be unique. Envoy registers every method of each listed service into one
	// path matcher, and a repeated service silently stops the filter from transcoding.
	// +optional
	// +listType=set
	Services []string `json:"services,omitempty"`

	// PrintOptions defines the output format options for JSON conversion.
	// +optional
	PrintOptions *JSONPrintOptions `json:"printOptions,omitempty"`

	// MatchIncomingRequestRoute keeps the route that matched the incoming request after
	// the transcoder rewrites the path to the gRPC method.
	//
	// When false (the default), the rewritten path is matched against the routing table
	// again, so a route matching the gRPC method must also exist or the request is
	// answered with 404. Either a GRPCRoute for the service, or an HTTPRoute matching
	// `/<package>.<Service>/<Method>`, satisfies this.
	// +optional
	MatchIncomingRequestRoute *bool `json:"matchIncomingRequestRoute,omitempty"`

	// IgnoredQueryParameters defines query parameters to ignore during transcoding.
	// +optional
	// +listType=set
	IgnoredQueryParameters []string `json:"ignoredQueryParameters,omitempty"`

	// AutoMapping enables automatic field mapping for HTTP query parameters and headers
	// to gRPC message fields when explicit field mapping is not present in the proto.
	// +optional
	AutoMapping *bool `json:"autoMapping,omitempty"`

	// IgnoreUnknownQueryParameters ignores query parameters that cannot be mapped to a
	// gRPC request field. Set this when the query parameters are not known in advance;
	// otherwise list them in IgnoredQueryParameters.
	//
	// When false (the default), a request carrying an unmappable parameter is not
	// transcoded at all and is forwarded unchanged, which a gRPC backend rejects.
	// +optional
	IgnoreUnknownQueryParameters *bool `json:"ignoreUnknownQueryParameters,omitempty"`

	// ConvertGRPCStatus enables converting gRPC status to HTTP status codes.
	// If true, gRPC status codes are converted to appropriate HTTP status codes.
	// +optional
	ConvertGRPCStatus *bool `json:"convertGRPCStatus,omitempty"`
}

// ProtoDescriptor locates the protocol buffer descriptor set describing the gRPC services.
// +kubebuilder:validation:XValidation:rule="size(self.valueRef.group) == 0 && self.valueRef.kind == 'ConfigMap'",message="valueRef must refer to a core ConfigMap"
type ProtoDescriptor struct {
	// ValueRef is a reference to a ConfigMap in the same namespace holding the descriptor.
	// The key `proto-descriptor` is used if present, otherwise the ConfigMap must hold
	// exactly one entry. A `binaryData` entry is used as-is; a `data` entry must be
	// base64-encoded.
	//
	// Generate the descriptor with `protoc --include_imports`; without the imported files
	// Envoy cannot build a descriptor pool.
	ValueRef gwapiv1.LocalObjectReference `json:"valueRef"`
}

// JSONPrintOptions defines options for JSON output formatting.
type JSONPrintOptions struct {
	// AddWhitespace adds whitespace for pretty-printing JSON output.
	// +optional
	AddWhitespace *bool `json:"addWhitespace,omitempty"`

	// AlwaysPrintPrimitiveFields always prints primitive fields even if they have default values.
	// +optional
	AlwaysPrintPrimitiveFields *bool `json:"alwaysPrintPrimitiveFields,omitempty"`

	// AlwaysPrintEnumsAsInts always prints enum values as integers instead of strings.
	// +optional
	AlwaysPrintEnumsAsInts *bool `json:"alwaysPrintEnumsAsInts,omitempty"`

	// PreserveProtoFieldNames preserves proto field names in JSON output instead
	// of converting them to camelCase.
	// +optional
	PreserveProtoFieldNames *bool `json:"preserveProtoFieldNames,omitempty"`
}
