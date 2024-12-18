// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

// Lua defines a Lua extension
// Only one of Inline or ValueRef must be set
//
// +kubebuilder:validation:XValidation:rule="has(self.inline) ? !has(self.valueRef) : has(self.valueRef)",message="Exactly one of inline or valueRef must be set."
type Lua struct {
	// Inline contains the source code as an inline string.
	//
	// +optional
	Inline *string `json:"inline,omitempty"`
	// ValueRef has the source code specified as a local object reference.
	// Only a reference to ConfigMap is supported.
	// The value of key `lua` in the ConfigMap will be used.
	// If the key is not found, the first value in the ConfigMap will be used.
	//
	// +kubebuilder:validation:XValidation:rule="self.kind == 'ConfigMap' && (!has(self.group) || self.group == '')",message="Only a reference to an object of kind ConfigMap belonging to default core API group is supported."
	// +optional
	ValueRef *gwapiv1.LocalObjectReference `json:"valueRef,omitempty"`
}
