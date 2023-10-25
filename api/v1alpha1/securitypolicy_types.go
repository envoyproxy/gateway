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
	Status SecurityPolicyStatus `json:"status,omitempty"`
}

// SecurityPolicySpec defines the desired state of SecurityPolicy.
type SecurityPolicySpec struct {
	// TargetRef is the name of the Gateway resource this policy
	// is being attached to.
	// This Policy and the TargetRef MUST be in the same namespace
	// for this Policy to have effect and be applied to the Gateway.
	// TargetRef
	TargetRef gwapiv1a2.PolicyTargetReferenceWithSectionName `json:"targetRef"`

	// CORS defines the configuration for Cross-Origin Resource Sharing (CORS).
	CORS *CORS `json:"cors,omitempty"`
}

// CORS defines the configuration for Cross-Origin Resource Sharing (CORS).
type CORS struct {
	// AllowOrigins defines the origins that are allowed to make requests.
	AllowOrigins []StringMatch `json:"allowOrigins,omitempty" yaml:"allowOrigins,omitempty"`
	// AllowMethods defines the methods that are allowed to make requests.
	AllowMethods []string `json:"allowMethods,omitempty" yaml:"allowMethods,omitempty"`
	// AllowHeaders defines the headers that are allowed to be sent with requests.
	AllowHeaders []string `json:"allowHeaders,omitempty" yaml:"allowHeaders,omitempty"`
	// ExposeHeaders defines the headers that can be exposed in the responses.
	ExposeHeaders []string `json:"exposeHeaders,omitempty" yaml:"exposeHeaders,omitempty"`
	// MaxAge defines how long the results of a preflight request can be cached.
	MaxAge *metav1.Duration `json:"maxAge,omitempty" yaml:"maxAge,omitempty"`
}

// StringMatch defines how to match any strings.
// This is a general purpose match condition that can be used by other EG APIs
// that need to match against a string.
type StringMatch struct {
	// Type specifies how to match against a string.
	//
	// +optional
	// +kubebuilder:default=Exact
	Type *MatchType `json:"type,omitempty"`

	// Value specifies the string value that the match must have.
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=1024
	Value string `json:"value"`
}

// MatchType specifies the semantics of how a string value should be compared.
// Valid MatchType values are "Exact", "Prefix", "Suffix", "RegularExpression".
//
// +kubebuilder:validation:Enum=Exact;Prefix;Suffix;Contains;RegularExpression
type MatchType string

const (
	// MatchExact :the input string must match exactly the match value.
	MatchExact MatchType = "Exact"

	// MatchPrefix :the input string must start with the match value.
	MatchPrefix MatchType = "Prefix"

	// MatchSuffix :the input string must end with the match value.
	MatchSuffix MatchType = "Suffix"

	// MatchRegularExpression :The input string must match the regular expression
	// specified in the match value.
	// The regex string must adhere to the syntax documented in
	// https://github.com/google/re2/wiki/Syntax.
	MatchRegularExpression MatchType = "RegularExpression"
)

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
