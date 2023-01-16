// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// KindRateLimitFilter is the name of the RateLimitFilter kind.
	KindRateLimitFilter = "RateLimitFilter"
)

// +kubebuilder:object:root=true

// RateLimitFilter allows the user to limit the number of incoming requests
// to a predefined value based on attributes within the traffic flow.
type RateLimitFilter struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of RateLimitFilter.
	Spec RateLimitFilterSpec `json:"spec"`
}

// RateLimitFilterSpec defines the desired state of RateLimitFilter.
// +union
type RateLimitFilterSpec struct {
	// Type decides the scope for the RateLimits.
	// Valid RateLimitType values are:
	//
	// * "Global" - In this mode, the rate limits are applied across all Envoy proxy instances.
	//
	// +unionDiscriminator
	Type RateLimitType `json:"type"`
	// Global rate limit configuration.
	//
	// +optional
	Global *GlobalRateLimit `json:"global,omitempty"`
}

// RateLimitType specifies the types of RateLimiting.
// +kubebuilder:validation:Enum=Global
type RateLimitType string

const (
	// GlobalRateLimitType allows the rate limits to be applied across all Envoy proxy instances.
	GlobalRateLimitType RateLimitType = "Global"
)

// GlobalRateLimit defines the global rate limit configuration.
type GlobalRateLimit struct {
	// Rules are a list of RateLimit selectors and limits.
	// Each rule and its associated limit is applied
	// in a mutually exclusive way i.e. if multiple
	// rules get selected, each of their associated
	// limits get applied, so a single traffic request
	// might increase the rate limit counters for multiple
	// rules if selected.
	//
	// +kubebuilder:validation:MaxItems=16
	Rules []RateLimitRule `json:"rules"`
}

// RateLimitRule defines the semantics for matching attributes
// from the incoming requests, and setting limits for them.
type RateLimitRule struct {
	// ClientSelectors holds the list of select conditions to select
	// specific clients using attributes from the traffic flow.
	// All individual select conditions must hold True for this rule
	// and its limit to be applied.
	// If this field is empty, it is equivalent to True, and
	// the limit is applied.
	//
	// +optional
	// +kubebuilder:validation:MaxItems=8
	ClientSelectors []RateLimitSelectCondition `json:"clientSelectors,omitempty"`
	// Limit holds the rate limit values.
	// This limit is applied for traffic flows when the selectors
	// compute to True, causing the request to be counted towards the limit.
	// The limit is enforced and the request is ratelimited, i.e. a response with
	// 429 HTTP status code is sent back to the client when
	// the selected requests have reached the limit.
	Limit RateLimitValue `json:"limit"`
}

// RateLimitSelectCondition specifies the attributes within the traffic flow that can
// be used to select a subset of clients to be ratelimited.
// All the individual conditions must hold True for the overall condition to hold True.
type RateLimitSelectCondition struct {
	// Headers is a list of request headers to match. Multiple header values are ANDed together,
	// meaning, a request MUST match all the specified headers.
	//
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
	// Do not set this field when Type="Distinct", implying matching on any/all unique values within the header.
	// +optional
	// +kubebuilder:validation:MaxLength=1024
	Value *string `json:"value,omitempty"`
}

// HeaderMatchType specifies the semantics of how HTTP header values should be
// compared. Valid HeaderMatchType values are:
//
//   - "Exact": Use this type to match the exact value of the Value field against the value of the specified HTTP Header.
//   - "RegularExpression": Use this type to match a regular expression against the value of the specified HTTP Header.
//     The regex string must adhere to the syntax documented in https://github.com/google/re2/wiki/Syntax.
//   - "Distinct": Use this type to match any and all possible unique values encountered in the specified HTTP Header.
//     Note that each unique value will receive its own rate limit bucket.
//
// +kubebuilder:validation:Enum=Exact;RegularExpression;Distinct
type HeaderMatchType string

// HeaderMatchType constants.
const (
	HeaderMatchExact             HeaderMatchType = "Exact"
	HeaderMatchRegularExpression HeaderMatchType = "RegularExpression"
	HeaderMatchDistinct          HeaderMatchType = "Distinct"
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

//+kubebuilder:object:root=true

// RateLimitFilterList contains a list of RateLimitFilter resources.
type RateLimitFilterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RateLimitFilter `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RateLimitFilter{}, &RateLimitFilterList{})
}
