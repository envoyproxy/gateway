// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

// RateLimitSpec defines the desired state of RateLimitSpec.
// +union
type RateLimitSpec struct {
	// Type decides the scope for the RateLimits.
	// Valid RateLimitType values are "Global".
	//
	// +unionDiscriminator
	Type RateLimitType `json:"type"`
	// Global defines global rate limit configuration.
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

// GlobalRateLimit defines global rate limit configuration.
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
	Headers []RateLimitHeaderMatch `json:"headers,omitempty"`

	// SourceCIDR is the client IP Address range to match on.
	//
	// +optional
	SourceCIDR *RateLimitSourceMatch `json:"sourceCIDR,omitempty"`
}

type RateLimitSourceMatchType string

const (
	// SourceMatchExact All IP Addresses within the specified Source IP CIDR are treated as a single client selector
	// and share the same rate limit bucket.
	SourceMatchExact RateLimitSourceMatchType = "Exact"
	// SourceMatchDistinct Each IP Address within the specified Source IP CIDR is treated as a distinct client selector
	// and uses a separate rate limit bucket/counter.
	SourceMatchDistinct RateLimitSourceMatchType = "Distinct"
)

type RateLimitSourceMatch struct {
	// +optional
	// +kubebuilder:default=Exact
	Type *RateLimitSourceMatchType `json:"type,omitempty"`

	// Value is the IP CIDR that represents the range of Source IP Addresses of the client.
	// These could also be the intermediate addresses through which the request has flown through and is part of the  `X-Forwarded-For` header.
	// For example, `192.168.0.1/32`, `192.168.0.0/24`, `001:db8::/64`.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=256
	Value string `json:"value"`
}

// RateLimitHeaderMatch defines the match attributes within the HTTP Headers of the request.
type RateLimitHeaderMatch struct {
	// Name of the HTTP header.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=256
	Name string `json:"name"`

	// Distinct matches any and all possible unique values encountered in the
	// specified HTTP Header.
	// Note that each unique value will receive its own rate limit bucket.
	// Only one of Distinct or Match can be set.
	Distinct *bool `json:"distinct,omitempty"`

	// Match specifies how to match against the value of the header.
	// Do not set this field when Type="Distinct", implying matching on any/all unique
	// values within the header.
	// Only one of Distinct or Match can be set.
	Match *StringMatch `json:"match,omitempty"`
}

// RateLimitValue defines the limits for rate limiting.
type RateLimitValue struct {
	Requests uint          `json:"requests"`
	Unit     RateLimitUnit `json:"unit"`
}

// RateLimitUnit specifies the intervals for setting rate limits.
// Valid RateLimitUnit values are "Second", "Minute", "Hour", and "Day".
//
// +kubebuilder:validation:Enum=Second;Minute;Hour;Day
type RateLimitUnit string
