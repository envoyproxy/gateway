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
type EnvoyExtensionPolicySpec struct {
	// +kubebuilder:validation:XValidation:rule="self.group == 'gateway.networking.k8s.io'", message="this policy can only have a targetRef.group of gateway.networking.k8s.io"
	// +kubebuilder:validation:XValidation:rule="self.kind in ['Gateway', 'HTTPRoute', 'GRPCRoute', 'UDPRoute', 'TCPRoute', 'TLSRoute']", message="this policy can only have a targetRef.kind of Gateway/HTTPRoute/GRPCRoute/TCPRoute/UDPRoute/TLSRoute"
	// +kubebuilder:validation:XValidation:rule="!has(self.sectionName)",message="this policy does not yet support the sectionName field"
	//
	// TargetRef is the name of the resource this policy
	// is being attached to.
	// This Policy and the TargetRef MUST be in the same namespace
	// for this Policy to have effect and be applied to the Gateway or xRoute.
	TargetRef gwapiv1a2.LocalPolicyTargetReferenceWithSectionName `json:"targetRef"`

	// Wasm is a list of Wasm extensions to be loaded by the Gateway.
	// Order matters, as the extensions will be loaded in the order they are
	// defined in this list.
	//
	// +kubebuilder:validation:MaxItems=16
	// +optional
	Wasm []Wasm `json:"wasm,omitempty"`

	// ExtProc is an ordered list of external processing filters
	// that should added to the envoy filter chain
	//
	// +kubebuilder:validation:MaxItems=16
	// +optional
	ExtProc []ExtProc `json:"extProc,omitempty"`
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
