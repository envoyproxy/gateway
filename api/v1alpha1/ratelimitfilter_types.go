// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// RateLimitFilter allows the user to limit the number of incoming requests
// to a predefined value based on attributes within the traffic flow.
type RateLimitFilter struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of HTTPRoute.
	Spec RateLimitFilterSpec `json:"spec"`
}

// +kubebuilder:object:root=true

// RateLimitFilterList contains a list of RateLimitFilter resources.
type RateLimitList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RateLimitFilter `json:"items"`
}

// RateLimitFilterSpec defines the desired state of RateLimitFilter
type RateLimitFilterSpec struct {
	// Type decides the scope for the RateLimits.
	Type RateLimitType `json:"type"`
	// Rules are a list of RateLimit matchers and limits.
	//
	// +kubebuilder:validation:MaxItems=16
	Rules []RateLimitRule `json:"rules"`
}

// RateLimitType specifies the types of RateLimiting.
// Valid RateLimitType values are:
//
// * "Global"
//
// +kubebuilder:validation:Enum=Global
type RateLimitType string

const (
	// In this mode, the rate limits are applied across all Envoy proxy instances.
	RateLimitTypeGlobal RateLimitType = "Global"
)

// RateLimitRule defines the semantics for matching attributes
// from the incoming requests, and setting limits for them.
type RateLimitRule struct {
	// +optional
	// +kubebuilder:validation:MaxItems=8
	Matches []RateLimitMatch `json:"matches,omitempty"`
	Limit   RateLimitValue   `json:"limit"`
}

// RateLimitMatch specifies the attributes within the traffic flow that can
// be matched on.
type RateLimitMatch struct {
	// +listType=map
	// +listMapKey=name
	// +optional
	// +kubebuilder:validation:MaxItems=16
	Headers []HeaderMatch `json:"headers,omitempty"`
}

// HeaderMatch defines the match attributes within the HTTP Headers of the request.
type HeaderMatch struct {
	// Type specifies how to match against the value of the header.
	//
	// +optional
	// +kubebuilder:default=Exact
	Type *HeaderMatchType `json:"type,omitempty"`

	// Name of the HTTP header.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=256
	Name string `json:"name"`

	// Value within the HTTP header. Due to the
	// case-insensitivity of header names, "foo" and "Foo" are considered equivalent.
	// Do not set this field when Type="Any", implying matching on any/all unique values within the header.
	// +optional
	Value *string `json:"value,omitempty"`
}

// HeaderMatchType specifies the semantics of how HTTP header values should be
// compared. Valid HeaderMatchType values are:
//
// * "Exact": Use this type to match the exact value of the Value field against the value of the specified HTTP Header.
// * "RegularExpression": Use this type to match a regular expression against the value of the specified HTTP Header.
// * "Any": Use this type to match any and all possible unique values encountered in the specified HTTP Header.
//
// +kubebuilder:validation:Enum=Exact;RegularExpression;Any
type HeaderMatchType string

// HeaderMatchType constants.
const (
	HeaderMatchExact             HeaderMatchType = "Exact"
	HeaderMatchRegularExpression HeaderMatchType = "RegularExpression"
	HeaderMatchAny               HeaderMatchType = "Any"
)

// RateLimitValue defines the limits for rate limiting.
type RateLimitValue struct {
	Requests uint          `json:"requests"`
	Unit     RateLimitUnit `json:"unit"`
}

// RateLimitUnit specifies the intervals for setting rate limits.
// Valid RateLimitUnit values are:
//
// * "Second"
// * "Minute"
// * "Hour"
// * "Day"
//
// +kubebuilder:validation:Enum=Second;Minute;Hour;Day
type RateLimitUnit string
