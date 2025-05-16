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
	// +kubebuilder:validation:XValidation:rule="self.all(foo, !has(foo.cost) || !has(foo.cost.response))", message="response cost is not supported for Local Rate Limits"
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
	// Cost specifies the cost of requests and responses for the rule.
	//
	// This is optional and if not specified, the default behavior is to reduce the rate limit counters by 1 on
	// the request path and do not reduce the rate limit counters on the response path.
	//
	// +optional
	Cost *RateLimitCost `json:"cost,omitempty"`
	// Shared determines whether this rate limit rule applies across all the policy targets.
	// If set to true, the rule is treated as a common bucket and is shared across all policy targets (xRoutes).
	// Default: false.
	//
	// +optional
	Shared *bool `json:"shared,omitempty"`
}

type RateLimitCost struct {
	// Request specifies the number to reduce the rate limit counters
	// on the request path. If this is not specified, the default behavior
	// is to reduce the rate limit counters by 1.
	//
	// When Envoy receives a request that matches the rule, it tries to reduce the
	// rate limit counters by the specified number. If the counter doesn't have
	// enough capacity, the request is rate limited.
	//
	// +optional
	Request *RateLimitCostSpecifier `json:"request,omitempty"`
	// Response specifies the number to reduce the rate limit counters
	// after the response is sent back to the client or the request stream is closed.
	//
	// The cost is used to reduce the rate limit counters for the matching requests.
	// Since the reduction happens after the request stream is complete, the rate limit
	// won't be enforced for the current request, but for the subsequent matching requests.
	//
	// This is optional and if not specified, the rate limit counters are not reduced
	// on the response path.
	//
	// Currently, this is only supported for HTTP Global Rate Limits.
	//
	// +optional
	Response *RateLimitCostSpecifier `json:"response,omitempty"`
}

// RateLimitCostSpecifier specifies where the Envoy retrieves the number to reduce the rate limit counters.
//
// +kubebuilder:validation:XValidation:rule="!(has(self.number) && has(self.metadata))",message="only one of number or metadata can be specified"
type RateLimitCostSpecifier struct {
	// From specifies where to get the rate limit cost. Currently, only "Number" and "Metadata" are supported.
	//
	// +kubebuilder:validation:Required
	From RateLimitCostFrom `json:"from"`
	// Number specifies the fixed usage number to reduce the rate limit counters.
	// Using zero can be used to only check the rate limit counters without reducing them.
	//
	// +optional
	Number *uint64 `json:"number,omitempty"`
	// Metadata specifies the per-request metadata to retrieve the usage number from.
	//
	// +optional
	Metadata *RateLimitCostMetadata `json:"metadata,omitempty"`
}

// RateLimitCostFrom specifies the source of the rate limit cost.
// Valid RateLimitCostType values are "Number" and "Metadata".
//
// +kubebuilder:validation:Enum=Number;Metadata
type RateLimitCostFrom string

const (
	// RateLimitCostFromNumber specifies the rate limit cost to be a fixed number.
	RateLimitCostFromNumber RateLimitCostFrom = "Number"
	// RateLimitCostFromMetadata specifies the rate limit cost to be retrieved from the per-request dynamic metadata.
	RateLimitCostFromMetadata RateLimitCostFrom = "Metadata"
	// TODO: add headers, etc. Anything that can be represented in "Format" can be added here.
	// 	https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#config-access-log-format
)

// RateLimitCostMetadata specifies the filter metadata to retrieve the usage number from.
type RateLimitCostMetadata struct {
	// Namespace is the namespace of the dynamic metadata.
	//
	// +kubebuilder:validation:Required
	Namespace string `json:"namespace"`
	// Key is the key to retrieve the usage number from the filter metadata.
	//
	// +kubebuilder:validation:Required
	Key string `json:"key"`
}

// RateLimitSelectCondition specifies the attributes within the traffic flow that can
// be used to select a subset of clients to be ratelimited.
// All the individual conditions must hold True for the overall condition to hold True.
type RateLimitSelectCondition struct {
	// Headers is a list of request headers to match. Multiple header values are ANDed together,
	// meaning, a request MUST match all the specified headers.
	// At least one of headers or sourceCIDR condition must be specified.
	//
	// +optional
	// +kubebuilder:validation:MaxItems=16
	Headers []HeaderMatch `json:"headers,omitempty"`

	// SourceCIDR is the client IP Address range to match on.
	// At least one of headers or sourceCIDR condition must be specified.
	//
	// +optional
	SourceCIDR *SourceMatch `json:"sourceCIDR,omitempty"`
}

// +kubebuilder:validation:Enum=Exact;Distinct
type SourceMatchType string

const (
	// SourceMatchExact All IP Addresses within the specified Source IP CIDR are treated as a single client selector
	// and share the same rate limit bucket.
	SourceMatchExact SourceMatchType = "Exact"
	// SourceMatchDistinct Each IP Address within the specified Source IP CIDR is treated as a distinct client selector
	// and uses a separate rate limit bucket/counter.
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
type HeaderMatch struct {
	// Type specifies how to match against the value of the header.
	//
	// +optional
	// +kubebuilder:default=Exact
	Type *HeaderMatchType `json:"type,omitempty"`

	// Name of the HTTP header.
	// The header name is case-insensitive unless PreserveHeaderCase is set to true.
	// For example, "Foo" and "foo" are considered the same header.
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=256
	Name string `json:"name"`

	// Value within the HTTP header.
	// Do not set this field when Type="Distinct", implying matching on any/all unique
	// values within the header.
	//
	// +optional
	// +kubebuilder:validation:MaxLength=1024
	Value *string `json:"value,omitempty"`

	// Invert specifies whether the value match result will be inverted.
	// Do not set this field when Type="Distinct", implying matching on any/all unique
	// values within the header.
	//
	// +optional
	// +kubebuilder:default=false
	Invert *bool `json:"invert,omitempty"`
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
