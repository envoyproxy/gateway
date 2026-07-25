// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// CSRF defines the configuration for the Cross-Site Request Forgery (CSRF) filter.
// The CSRF filter checks that the Origin header in HTTP requests matches the destination,
// preventing cross-origin mutating requests (POST, PUT, DELETE, PATCH) from being processed.
// GET and HEAD requests are always allowed.
//
// Note: Envoy's CSRF filter compares against the host and port of the origin only
// (the scheme is stripped before matching). Additional origins must be specified as
// host or host:port values, not full URLs. For example, use "www.example.com"
// instead of "https://www.example.com".
//
// +kubebuilder:validation:XValidation:message="additionalOrigins must be host or host:port values without a scheme or path, for example www.example.com instead of https://www.example.com",rule="!has(self.additionalOrigins) || self.additionalOrigins.all(o, !o.value.contains('/'))"
type CSRF struct {
	// EnforcedFraction represents the fraction of requests for which the CSRF
	// policy is enforced. Requests that are not selected are allowed through
	// without any origin validation.
	// Defaults to 100% (all requests are enforced) if not specified.
	//
	// +optional
	EnforcedFraction *gwapiv1.Fraction `json:"enforcedFraction,omitempty"`

	// ShadowFraction represents the fraction of requests for which the CSRF
	// policy is evaluated in shadow (dry-run) mode. In this mode, the filter
	// evaluates requests and tracks whether they would be allowed or rejected in
	// the `csrf.request_invalid` and `csrf.request_valid` stats, but does not
	// enforce the policy. This is useful for rolling out CSRF protection
	// gradually while monitoring the impact.
	// Only takes effect for requests that are not selected by EnforcedFraction.
	//
	// +optional
	ShadowFraction *gwapiv1.Fraction `json:"shadowFraction,omitempty"`

	// AdditionalOrigins specifies additional origins that are allowed to make requests,
	// beyond the destination origin. These are checked against the Origin header (host:port only,
	// not the full URL) and if matched, the request is allowed.
	// Each origin supports Exact, Prefix, Suffix, and RegularExpression matching.
	//
	// +optional
	// +kubebuilder:validation:MaxItems=16
	AdditionalOrigins []StringMatch `json:"additionalOrigins,omitempty"`
}
