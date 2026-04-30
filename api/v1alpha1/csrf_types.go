// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

// CSRF defines the configuration for the Cross-Site Request Forgery (CSRF) filter.
// The CSRF filter checks that the Origin header in HTTP requests matches the destination,
// preventing cross-origin mutating requests (POST, PUT, DELETE, PATCH) from being processed.
// GET and HEAD requests are always allowed.
//
// Note: Envoy's CSRF filter compares against the host and port of the origin only
// (the scheme is stripped before matching). Additional origins must be specified as
// host or host:port values, not full URLs. For example, use "www.example.com"
// instead of "https://www.example.com".
type CSRF struct {
	// FilterEnabled specifies the percentage of requests for which the CSRF filter is enabled.
	// When set, only the given percentage of requests will have CSRF protection enforced.
	// Defaults to 100 (fully enabled) if not specified.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	FilterEnabled *int32 `json:"filterEnabled,omitempty"`

	// ShadowEnabled specifies the percentage of requests for which the CSRF filter is in
	// shadow/dry-run mode. In this mode, the filter evaluates requests and tracks whether
	// they would be allowed or rejected, but does not enforce the policy.
	// This is useful for rolling out CSRF protection gradually while monitoring the impact.
	// Only takes effect when FilterEnabled is not set or is 0.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	ShadowEnabled *int32 `json:"shadowEnabled,omitempty"`

	// AdditionalOrigins specifies additional origins that are allowed to make requests,
	// beyond the destination origin. These are checked against the Origin header (host:port only,
	// not the full URL) and if matched, the request is allowed.
	// Each origin supports Exact, Prefix, Suffix, and RegularExpression matching.
	//
	// +optional
	// +kubebuilder:validation:MaxItems=16
	AdditionalOrigins []StringMatch `json:"additionalOrigins,omitempty"`
}
