// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

// CSRF defines the configuration for the Cross-Site Request Forgery (CSRF) filter.
// The CSRF filter checks that the Origin header in HTTP requests matches the destination,
// preventing cross-origin mutating requests (POST, PUT, DELETE, PATCH) from being processed.
// GET and HEAD requests are always allowed.
type CSRF struct {
	// AdditionalOrigins specifies additional origins that are allowed to make requests.
	// These are checked against the Origin header and if matched, the request is allowed.
	// Each origin can be an exact, prefix, suffix, or regex match using StringMatch.
	//
	// +optional
	AdditionalOrigins []StringMatch `json:"additionalOrigins,omitempty"`
}
