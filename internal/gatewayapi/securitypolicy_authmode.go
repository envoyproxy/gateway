// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

// applyAuthMode is called after all individual auth methods (JWT, BasicAuth,
// OIDC, APIKeyAuth) have been constructed and before they are aggregated into
// ir.SecurityFeatures. When AuthMode is "Any" and more than one auth method
// is active, it:
//
//  1. Sets AllowMissing = true on every auth method that supports the flag.
//     This prevents a filter from immediately rejecting a request just because
//     the credentials it expects are absent — another filter may still accept
//     the request.
//
//  2. Records the mode in SecurityFeatures.AuthMode so that the downstream
//     xDS translator can emit the composite RBAC policy that enforces the
//     "at least one succeeded" invariant.
//
// Callers: translateSecurityPolicyForRoute, translateSecurityPolicyForGateway.

import (
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
)

// authMethodCount returns the number of non-nil auth methods present.
func authMethodCount(jwt *ir.JWT, basicAuth *ir.BasicAuth, oidc *ir.OIDC, apiKeyAuth *ir.APIKeyAuth) int {
	count := 0
	if jwt != nil {
		count++
	}
	if basicAuth != nil {
		count++
	}
	if oidc != nil {
		count++
	}
	if apiKeyAuth != nil {
		count++
	}
	return count
}

// applyAuthMode mutates the supplied auth-method IR objects and sets the
// AuthMode field on SecurityFeatures according to the policy's AuthMode
// setting.
//
// Parameters:
//   - policy    – the SecurityPolicy being translated
//   - features  – the partially-constructed SecurityFeatures IR (must not be nil)
//   - jwt       – may be nil if JWT is not configured in this policy
//   - basicAuth – may be nil if BasicAuth is not configured in this policy
//   - oidc      – may be nil
//   - apiKey    – may be nil
func applyAuthMode(
	policy *egv1a1.SecurityPolicy,
	features *ir.SecurityFeatures,
	jwt *ir.JWT,
	basicAuth *ir.BasicAuth,
	oidc *ir.OIDC,
	apiKeyAuth *ir.APIKeyAuth,
) {
	// Resolve the effective mode (default: All).
	mode := egv1a1.AuthenticationModeAll
	if policy.Spec.AuthMode != nil {
		mode = *policy.Spec.AuthMode
	}

	if mode != egv1a1.AuthenticationModeAny {
		// Nothing to do — "All" semantics are the existing default behaviour.
		return
	}

	// "Any" mode only makes sense when there are multiple auth methods.
	// If there is only one configured method, OR semantics degenerate to AND,
	// so we skip the mutation to avoid unnecessary complexity in the generated
	// xDS config.
	if authMethodCount(jwt, basicAuth, oidc, apiKeyAuth) < 2 {
		return
	}

	// --- Mutate each auth method to be non-fatal when credentials are absent ---

	// JWT: set AllowMissing so that requests without a Bearer token are not
	// immediately rejected — they may succeed via BasicAuth or another method.
	if jwt != nil {
		jwt.AllowMissing = true
	}

	// BasicAuth: set AllowMissing so that requests without an Authorization:
	// Basic header pass through the filter to the next auth stage.
	//
	// Note: this requires Envoy to expose allow_missing on its basic_auth
	// HTTP filter (tracked in envoyproxy/envoy#43911). The xDS translator
	// emits the flag only when the running Envoy version supports it; see
	// internal/xds/translator/basicauth.go for the version guard.
	if basicAuth != nil {
		basicAuth.AllowMissing = true
	}

	// Record the mode so the xDS translator knows to emit the composite RBAC.
	features.AuthMode = ptr.To(string(egv1a1.AuthenticationModeAny))
}
