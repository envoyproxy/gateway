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
	// KindBackendTrafficPolicy is the name of the BackendTrafficPolicy kind.
	KindBackendTrafficPolicy = "BackendTrafficPolicy"
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=btpolicy
// +kubebuilder:subresource:status
// +kubebuilder:subresource:overrideStrategy
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.conditions[?(@.type=="Accepted")].reason`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
//
// BackendTrafficPolicy allows the user to configure the behavior of the connection
// between the downstream client and Envoy Proxy listener.
type BackendTrafficPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// spec defines the desired state of BackendTrafficPolicy.
	Spec BackendTrafficPolicySpec `json:"spec"`

	// status defines the current status of BackendTrafficPolicy.
	Status BackendTrafficPolicyStatus `json:"status,omitempty"`
}

// spec defines the desired state of BackendTrafficPolicy.
type BackendTrafficPolicySpec struct {

	// +kubebuilder:validation:XValidation:rule="self.kind == 'Gateway' || self.kind == 'HTTPRoute' || self.kind == 'GRPCRoute' || self.kind == 'UDPRoute' || self.kind == 'TCPRoute' || self.kind == 'TLSRoute'", message="this policy can only have a targetRef.kind of Gateway/HTTPRoute/GRPCRoute/TCPRoute/UDPRoute/TLSRoute"
	//
	// targetRef is the name of the resource this policy
	// is being attached to.
	// This Policy and the TargetRef MUST be in the same namespace
	// for this Policy to have effect and be applied to the Gateway.
	TargetRef gwapiv1a2.PolicyTargetReferenceWithSectionName `json:"targetRef"`
}

// BackendTrafficPolicyStatus defines the state of BackendTrafficPolicy
type BackendTrafficPolicyStatus struct {
	// Conditions describe the current conditions of the BackendTrafficPolicy.
	//
	// +optional
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MaxItems=8
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// BackendTrafficPolicyList contains a list of BackendTrafficPolicy resources.
type BackendTrafficPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BackendTrafficPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BackendTrafficPolicy{}, &BackendTrafficPolicyList{})
}
