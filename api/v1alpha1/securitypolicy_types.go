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
	// KindSecurityPolicy is the name of the SecurityPolicy kind.
	KindSecurityPolicy = "SecurityPolicy"
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories=envoy-gateway,shortName=sp
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.conditions[?(@.type=="Accepted")].reason`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// SecurityPolicy allows the user to configure various security settings for a
// Gateway.
type SecurityPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of SecurityPolicy.
	Spec SecurityPolicySpec `json:"spec"`

	// Status defines the current status of SecurityPolicy.
	Status gwapiv1a2.PolicyStatus `json:"status,omitempty"`
}

// SecurityPolicySpec defines the desired state of SecurityPolicy.
type SecurityPolicySpec struct {
	// +kubebuilder:validation:XValidation:rule="self.group == 'gateway.networking.k8s.io'", message="this policy can only have a targetRef.group of gateway.networking.k8s.io"
	// +kubebuilder:validation:XValidation:rule="self.kind in ['Gateway', 'HTTPRoute', 'GRPCRoute']", message="this policy can only have a targetRef.kind of Gateway/HTTPRoute/GRPCRoute"
	// +kubebuilder:validation:XValidation:rule="!has(self.sectionName)",message="this policy does not yet support the sectionName field"
	//
	// TargetRef is the name of the Gateway resource this policy
	// is being attached to.
	// This Policy and the TargetRef MUST be in the same namespace
	// for this Policy to have effect and be applied to the Gateway.
	TargetRef gwapiv1a2.PolicyTargetReferenceWithSectionName `json:"targetRef"`

	// CORS defines the configuration for Cross-Origin Resource Sharing (CORS).
	//
	// +optional
	CORS *CORS `json:"cors,omitempty"`

	// BasicAuth defines the configuration for the HTTP Basic Authentication.
	//
	// +optional
	BasicAuth *BasicAuth `json:"basicAuth,omitempty"`

	// JWT defines the configuration for JSON Web Token (JWT) authentication.
	//
	// +optional
	JWT *JWT `json:"jwt,omitempty"`

	// OIDC defines the configuration for the OpenID Connect (OIDC) authentication.
	//
	// +optional
	OIDC *OIDC `json:"oidc,omitempty"`

	// ExtAuth defines the configuration for External Authorization.
	//
	// +optional
	ExtAuth *ExtAuth `json:"extAuth,omitempty"`
}

// SecurityPolicyStatus defines the state of SecurityPolicy
type SecurityPolicyStatus struct {
	// Conditions describe the current conditions of the SecurityPolicy.
	//
	// +optional
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MaxItems=8
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true

// SecurityPolicyList contains a list of SecurityPolicy resources.
type SecurityPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SecurityPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SecurityPolicy{}, &SecurityPolicyList{})
}
