// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

// RateLimitSpec defines the desired state of RateLimitSpec.
// +union
type RateLimitSpec struct {
	// Type decides the scope for the RateLimits.
	// Valid RateLimitType values are "Global" or "Local".
	//
	// +unionDiscriminator
	Type RateLimitType `json:"type"`
	// Global defines global rate limit configuration.
	//
	// +optional
	Global *GlobalRateLimit `json:"global,omitempty"`

	// Local defines local rate limit configuration.
	//
	// +optional
	Local *LocalRateLimit `json:"local,omitempty"`
}

// RateLimitType specifies the types of RateLimiting.
// +kubebuilder:validation:Enum=Global;Local
type RateLimitType string

const (
	// GlobalRateLimitType allows the rate limits to be applied across all Envoy
	// proxy instances.
	GlobalRateLimitType RateLimitType = "Global"

	// LocalRateLimitType allows the rate limits to be applied on a per Envoy
	// proxy instance basis.
	LocalRateLimitType RateLimitType = "Local"
)

// GlobalRateLimit defines global rate limit configuration.
type GlobalRateLimit struct {
	// Rules are a list of RateLimit selectors and limits. Each rule and its
	// associated limit is applied in a mutually exclusive way. If a request
	// matches multiple rules, each of their associated limits get applied, so a
	// single request might increase the rate limit counters for multiple rules
	// if selected. The rate limit service will return a logical OR of the individual
	// rate limit decisions of all matching rules. For example, if a request
	// matches two rules, one rate limited and one not, the final decision will be
	// to rate limit the request.
	//
	// +kubebuilder:validation:MaxItems=64
	Rules []RateLimitRule `json:"rules"`
}

// LocalRateLimit defines local rate limit configuration.
type LocalRateLimit struct {
	// Rules are a list of RateLimit selectors and limits. If a request matches
	// multiple rules, the strictest limit is applied. For example, if a request
	// matches two rules, one with 10rps and one with 20rps, the final limit will
	// be based on the rule with 10rps.
	//
	// +optional
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
	//
	// If no client selectors are specified, the rule applies to all traffic of
	// the targeted Route.
	//
	// If the policy targets a Gateway, the rule applies to each Route of the Gateway.
	// Please note that each Route has its own rate limit counters. For example,
	// if a Gateway has two Routes, and the policy has a rule with limit 10rps,
	// each Route will have its own 10rps limit.
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
	// At least one of headers or sourceCIDR condition must be specified.
	//
	// +listType=map
	// +listMapKey=name
	// +optional
	// +kubebuilder:validation:MaxItems=16
	Headers []HeaderMatch `json:"headers,omitempty"`

	// SourceCIDR is the client IP Address range to match on.
	// At least one of headers or sourceCIDR condition must be specified.
	//
	// +optional
	SourceCIDR *SourceMatch `json:"sourceCIDR,omitempty"`
}

type SourceMatchType string

const (
	// SourceMatchExact All IP Addresses within the specified Source IP CIDR are treated as a single client selector
	// and share the same rate limit bucket.
	SourceMatchExact SourceMatchType = "Exact"
	// SourceMatchDistinct Each IP Address within the specified Source IP CIDR is treated as a distinct client selector
	// and uses a separate rate limit bucket/counter.
	// Note: This is only supported for Global Rate Limits.
	SourceMatchDistinct SourceMatchType = "Distinct"
)

type SourceMatch struct {
	// +optional
	// +kubebuilder:default=Exact
	Type *SourceMatchType `json:"type,omitempty"`

	// Value is the IP CIDR that represents the range of Source IP Addresses of the client.
	// These could also be the intermediate addresses through which the request has flown through and is part of the  `X-Forwarded-For` header.
	// For example, `192.168.0.1/32`, `192.168.0.0/24`, `001:db8::/64`.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=256
	Value string `json:"value"`
}

// HeaderMatch defines the match attributes within the HTTP Headers of the request.
type HeaderMatch struct { // TODO: zhaohuabing this type could be replaced with a general purpose StringMatch type.
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
	// Do not set this field when Type="Distinct", implying matching on any/all unique
	// values within the header.
	//
	// +optional
	// +kubebuilder:validation:MaxLength=1024
	Value *string `json:"value,omitempty"`
}

// HeaderMatchType specifies the semantics of how HTTP header values should be compared.
// Valid HeaderMatchType values are "Exact", "RegularExpression", and "Distinct".
//
// +kubebuilder:validation:Enum=Exact;RegularExpression;Distinct
type HeaderMatchType string

// HeaderMatchType constants.
const (
	// HeaderMatchExact matches the exact value of the Value field against the value of
	// the specified HTTP Header.
	HeaderMatchExact HeaderMatchType = "Exact"
	// HeaderMatchRegularExpression matches a regular expression against the value of the
	// specified HTTP Header. The regex string must adhere to the syntax documented in
	// https://github.com/google/re2/wiki/Syntax.
	HeaderMatchRegularExpression HeaderMatchType = "RegularExpression"
	// HeaderMatchDistinct matches any and all possible unique values encountered in the
	// specified HTTP Header. Note that each unique value will receive its own rate limit
	// bucket.
	// Note: This is only supported for Global Rate Limits.
	HeaderMatchDistinct HeaderMatchType = "Distinct"
)

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

// RateLimitUnit constants.
const (
	// RateLimitUnitSecond specifies the rate limit interval to be 1 second.
	RateLimitUnitSecond RateLimitUnit = "Second"

	// RateLimitUnitMinute specifies the rate limit interval to be 1 minute.
	RateLimitUnitMinute RateLimitUnit = "Minute"

	// RateLimitUnitHour specifies the rate limit interval to be 1 hour.
	RateLimitUnitHour RateLimitUnit = "Hour"

	// RateLimitUnitDay specifies the rate limit interval to be 1 day.
	RateLimitUnitDay RateLimitUnit = "Day"
)
