// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

// LuaValueType defines the types of values for Lua supported by Envoy Gateway.
// +kubebuilder:validation:Enum=Inline;ValueRef
type LuaValueType string

const (
	// LuaValueTypeInline defines the "Inline" Lua type.
	LuaValueTypeInline LuaValueType = "Inline"

	// LuaValueTypeValueRef defines the "ValueRef" Lua type.
	LuaValueTypeValueRef LuaValueType = "ValueRef"
)

// Lua defines a Lua extension
// Only one of Inline or ValueRef must be set
//
// +kubebuilder:validation:XValidation:rule="(self.type == 'Inline' && has(self.inline) && !has(self.valueRef)) || (self.type == 'ValueRef' && !has(self.inline) && has(self.valueRef))",message="Exactly one of inline or valueRef must be set with correct type."
type Lua struct {
	// Type is the type of method to use to read the Lua value.
	// Valid values are Inline and ValueRef, default is Inline.
	//
	// +kubebuilder:default=Inline
	// +unionDiscriminator
	// +required
	Type LuaValueType `json:"type"`
	// Inline contains the source code as an inline string.
	//
	// +optional
	// +unionMember
	Inline *string `json:"inline,omitempty"`
	// ValueRef has the source code specified as a local object reference.
	// Only a reference to ConfigMap is supported.
	// The value of key `lua` in the ConfigMap will be used.
	// If the key is not found, the first value in the ConfigMap will be used.
	//
	// +kubebuilder:validation:XValidation:rule="self.kind == 'ConfigMap' && (self.group == 'v1' || self.group == '')",message="Only a reference to an object of kind ConfigMap belonging to default v1 API group is supported."
	// +optional
	// +unionMember
	ValueRef *gwapiv1.LocalObjectReference `json:"valueRef,omitempty"`
}
