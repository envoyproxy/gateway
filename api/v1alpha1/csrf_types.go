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
type CSRF struct {
	// ShadowFraction represents the fraction of requests for which the CSRF policy is
	// evaluated in shadow (dry-run) mode. For these requests, the filter records whether
	// the request would have been allowed or rejected in the `csrf.request_valid` and
	// `csrf.request_invalid` stats, but always lets the request through. The remaining
	// requests are enforced, i.e. a mutating request with a missing or non-matching
	// Origin header is rejected with a 403.
	//
	// Defaults to 0% (all requests are enforced) if not specified. Set it to 100% to
	// dry run the filter, watch the stats to find origins that would be rejected, then
	// lower it to roll enforcement out gradually.
	//
	// +optional
	ShadowFraction *gwapiv1.Fraction `json:"shadowFraction,omitempty"`

	// AdditionalOrigins specifies additional origins that are allowed to make mutating
	// requests, beyond the destination origin. A request whose Origin header matches one
	// of them is allowed. The value "*" allows any origin, which effectively disables
	// origin validation.
	//
	// Note: Envoy's CSRF filter compares the host and port of the origin only, so the
	// scheme is ignored: "https://www.example.com" and "http://www.example.com" are
	// equivalent here, and both allow the request regardless of the scheme the client
	// used.
	//
	// +optional
	// +kubebuilder:validation:MaxItems=16
	AdditionalOrigins []Origin `json:"additionalOrigins,omitempty"`
}
