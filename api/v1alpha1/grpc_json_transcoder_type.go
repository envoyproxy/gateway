// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// KindGrpcJsonTranscpderFilterKind is the name of the CorsFilter kind.
	KindGrpcJsonTranscpderFilter = "KindGrpcJsonTranscpderFilter"
)

//+kubebuilder:object:root=true

type GrpcJsonTranscpderFilter struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of the GrpcJsonTranscpderFilter type.
	Spec GrpcJsonTranscpderFilterSpec `json:"spec"`
}

// GrpcJsonTranscpderFilterSpec defines the desired state of the GrpcJsonTranscpderFilter type.
// +union
//
//	proto_descriptor_bin: "CtgECg1hcGkveWluLnByb3RvEgN5aW4iHAoKR2V0UmVxdWVzdBIOCgJpZBgBIAEoA1ICaWQiKQoIR2V0UmVwbHkSHQoEaXRlbRgBIAEoCzIJLnlpbi5JdGVtUgRpdGVtIhwKBEl0ZW0SFAoFYnl0ZXMYASABKAlSBWJ5dGVzMi4KA1lpbhInCgNHZXQSDy55aW4uR2V0UmVxdWVzdBoNLnlpbi5HZXRSZXBseSIAQjNaMWdpdGh1Yi5jb20vR2VvQ29tcGx5L21vbm9yZXBvL2djaS95aW4vcGtnL2FwaTt5aW5K7QIKBhIEAAATAQoICgEMEgMAABIKCAoBAhIDAgAMCggKAQgSAwQASAoJCgIICxIDBABICgoKAgYAEgQGAAgBCgoKAwYAARIDBggLCgsKBAYAAgASAwcCKwoMCgUGAAIAARIDBwYJCgwKBQYAAgACEgMHChQKDAoFBgACAAMSAwcfJwoKCgIEABIECgAMAQoKCgMEAAESAwoIEgoLCgQEAAIAEgMLAg8KDAoFBAACAAUSAwsCBwoMCgUEAAIAARIDCwgKCgwKBQQAAgADEgMLDQ4KCgoCBAESBA0ADwEKCgoDBAEBEgMNCBAKCwoEBAECABIDDgIQCgwKBQQBAgAGEgMOAgYKDAoFBAECAAESAw4HCwoMCgUEAQIAAxIDDg4PCgoKAgQCEgQRABMBCgoKAwQCARIDEQgMCgsKBAQCAgASAxICEwoMCgUEAgIABRIDEgIICgwKBQQCAgABEgMSCQ4KDAoFBAICAAMSAxIREmIGcHJvdG8z"
//	services:
//	- yin.Yin
//	auto_mapping: true
//	print_options:
//	add_whitespace: true
//	always_print_primitive_fields: true
//	always_print_enums_as_ints: false
//	preserve_proto_field_names: false
type GrpcJsonTranscpderFilterSpec struct {
	// ProtoDescriptorBin is the base64 encoded binary representation of the proto descriptor.
	// +kubebuilder:validation:Required
	ProtoDescriptorBin string   `json:"proto_descriptor_bin"`
	Services           []string `json:"services"`
	// AutoMapping is a flag that indicates whether the filter should automatically map the incoming request to the appropriate method in the proto descriptor.
	// +kubebuilder:validation:Required
	AutoMapping bool `json:"auto_mapping"`
	// PrintOptions is a set of options that controls how the filter generates the JSON response.
	PrintOptions *PrintOptions `json:"print_options,omitempty"`
}

type PrintOptions struct {
	// AddWhitespace is a flag that indicates whether the filter should add whitespace to the JSON response.
	// +kubebuilder:validation:Required
	AddWhitespace bool `json:"add_whitespace"`
	// AlwaysPrintPrimitiveFields is a flag that indicates whether the filter should always print primitive fields in the JSON response.
	// +kubebuilder:validation:Required
	AlwaysPrintPrimitiveFields bool `json:"always_print_primitive_fields"`
	// AlwaysPrintEnumsAsInts is a flag that indicates whether the filter should always print enums as ints in the JSON response.
	// +kubebuilder:validation:Required
	AlwaysPrintEnumsAsInts bool `json:"always_print_enums_as_ints"`
	// PreserveProtoFieldNames is a flag that indicates whether the filter should preserve proto field names in the JSON response.
	// +kubebuilder:validation:Required
	PreserveProtoFieldNames bool `json:"preserve_proto_field_names"`
}

// +kubebuilder:object:root=true
type GrpcJsonTranscpderFilterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GrpcJsonTranscpderFilter `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GrpcJsonTranscpderFilter{}, &GrpcJsonTranscpderFilterList{})
}
