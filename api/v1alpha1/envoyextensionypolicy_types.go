// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

const (
	// KindEnvoyExtensionPolicy is the name of the EnvoyExtensionPolicy kind.
	KindEnvoyExtensionPolicy = "EnvoyExtensionPolicy"
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=eep
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// EnvoyExtensionPolicy allows the user to configure various envoy extensibility options for the Gateway.
type EnvoyExtensionPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of EnvoyExtensionPolicy.
	Spec EnvoyExtensionPolicySpec `json:"spec"`

	// Status defines the current status of EnvoyExtensionPolicy.
	Status gwapiv1a2.PolicyStatus `json:"status,omitempty"`
}

// EnvoyExtensionPolicySpec defines the desired state of EnvoyExtensionPolicy.
//
// +kubebuilder:validation:XValidation:rule="(has(self.targetRef) && !has(self.targetRefs)) || (!has(self.targetRef) && has(self.targetRefs)) || (has(self.targetSelectors) && self.targetSelectors.size() > 0) ", message="either targetRef or targetRefs must be used"
// +kubebuilder:validation:XValidation:rule="has(self.targetRef) ? self.targetRef.group == 'gateway.networking.k8s.io' : true", message="this policy can only have a targetRef.group of gateway.networking.k8s.io"
// +kubebuilder:validation:XValidation:rule="has(self.targetRef) ? self.targetRef.kind in ['Gateway', 'HTTPRoute', 'GRPCRoute', 'UDPRoute', 'TCPRoute', 'TLSRoute'] : true", message="this policy can only have a targetRef.kind of Gateway/HTTPRoute/GRPCRoute/TCPRoute/UDPRoute/TLSRoute"
// +kubebuilder:validation:XValidation:rule="has(self.targetRef) ? !(has(self.targetRef.sectionName) && self.targetRef.kind == 'Gateway') : true",message="this policy does not yet support the sectionName field for Gateway"
// +kubebuilder:validation:XValidation:rule="has(self.targetRefs) ? self.targetRefs.all(ref, ref.group == 'gateway.networking.k8s.io') : true ", message="this policy can only have a targetRefs[*].group of gateway.networking.k8s.io"
// +kubebuilder:validation:XValidation:rule="has(self.targetRefs) ? self.targetRefs.all(ref, ref.kind in ['Gateway', 'HTTPRoute', 'GRPCRoute', 'UDPRoute', 'TCPRoute', 'TLSRoute']) : true ", message="this policy can only have a targetRefs[*].kind of Gateway/HTTPRoute/GRPCRoute/TCPRoute/UDPRoute/TLSRoute"
// +kubebuilder:validation:XValidation:rule="has(self.targetRefs) ? self.targetRefs.all(ref, !(has(ref.sectionName) && ref.kind == 'Gateway')) : true",message="this policy does not yet support the sectionName field for Gateway"
type EnvoyExtensionPolicySpec struct {
	PolicyTargetReferences `json:",inline"`

	// Wasm is a list of Wasm extensions to be loaded by the Gateway.
	// Order matters, as the extensions will be loaded in the order they are
	// defined in this list.
	//
	// +kubebuilder:validation:MaxItems=16
	// +optional
	Wasm []Wasm `json:"wasm,omitempty"`

	// ExtProc is an ordered list of external processing filters
	// that should be added to the envoy filter chain
	//
	// +kubebuilder:validation:MaxItems=16
	// +optional
	ExtProc []ExtProc `json:"extProc,omitempty"`

	// Lua is an ordered list of Lua filters
	// that should be added to the envoy filter chain
	//
	// +kubebuilder:validation:MaxItems=16
	// +optional
	Lua []Lua `json:"lua,omitempty"`
}

//+kubebuilder:object:root=true

// EnvoyExtensionPolicyList contains a list of EnvoyExtensionPolicy resources.
type EnvoyExtensionPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EnvoyExtensionPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EnvoyExtensionPolicy{}, &EnvoyExtensionPolicyList{})
}
