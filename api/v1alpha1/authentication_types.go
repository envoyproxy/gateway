// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

// AuthenticationMode defines how multiple authentication methods within a
// SecurityPolicy are evaluated against an incoming request.
//
// +kubebuilder:validation:Enum=All;Any
type AuthenticationMode string

const (
	// AuthenticationModeAll (default) requires ALL configured authentication
	// methods to succeed for a request to be accepted.
	//
	// Example: if both JWT and BasicAuth are configured, the request must carry
	// a valid Bearer token AND valid Basic credentials. Requests that satisfy
	// only one method are rejected.
	AuthenticationModeAll AuthenticationMode = "All"

	// AuthenticationModeAny requires AT LEAST ONE configured authentication
	// method to succeed for a request to be accepted.
	//
	// Example: if both JWT and BasicAuth are configured, a request carrying a
	// valid Bearer token is accepted even if no Basic credentials are present,
	// and vice-versa.
	//
	// When AuthenticationModeAny is set, the gateway automatically enables
	// AllowMissing on every individual auth filter so that the absence of
	// credentials for a given method is not immediately fatal. A downstream
	// Authorization policy then enforces that at least one method produced a
	// positive authentication signal.
	//
	// Note: full OR-semantics for BasicAuth require Envoy to support the
	// allow_missing flag on the basic_auth HTTP filter
	// (envoyproxy/envoy#43911). Until that upstream change is merged, bearer
	// requests are allowed to bypass BasicAuth only when AllowMissing is
	// propagated; requests that carry a malformed Basic header are still
	// rejected by Envoy before the auth decision reaches the RBAC layer.
	AuthenticationModeAny AuthenticationMode = "Any"
)
