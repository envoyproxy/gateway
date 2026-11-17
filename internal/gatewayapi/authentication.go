// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
)

const authenticationAnySuffix = "/authentication-any"

func authenticationAny(policy *egv1a1.SecurityPolicy) bool {
	return policy.Spec.Authentication != nil && len(policy.Spec.Authentication.AnyOf) > 0
}

// validateAuthentication runs on the effective (merged) policy, so it also covers
// configurations that admission cannot see, such as methods inherited from a
// parent policy or policies loaded without CRD validation.
func validateAuthentication(policy *egv1a1.SecurityPolicy) error {
	if !authenticationAny(policy) {
		return nil
	}
	spec := policy.Spec
	anyOf := spec.Authentication.AnyOf
	if len(anyOf) != 2 || !slices.Contains(anyOf, egv1a1.AuthenticationMethodJWT) || !slices.Contains(anyOf, egv1a1.AuthenticationMethodBasicAuth) {
		return errors.New("authentication.anyOf must list JWT and BasicAuth")
	}
	if spec.JWT == nil || len(spec.JWT.Providers) == 0 || spec.BasicAuth == nil {
		return errors.New("authentication.anyOf requires jwt and basicAuth")
	}
	if spec.OIDC != nil {
		// The oauth2 filter strips the client's Authorization header when it
		// forwards the access token, so Basic credentials never reach basic_auth.
		if ptr.Deref(spec.OIDC.ForwardAccessToken, false) {
			return errors.New("authentication.anyOf does not support oidc.forwardAccessToken")
		}
		// Pass-through skips OIDC on the strength of a JWT that jwt_authn must
		// then require, while anyOf lets jwt_authn pass requests without one.
		if ptr.Deref(spec.OIDC.PassThroughAuthHeader, false) {
			return errors.New("authentication.anyOf does not support oidc.passThroughAuthHeader")
		}
		// The oauth2 filter overwrites the forwarded ID token header, replacing
		// the Basic credentials or client JWT that anyOf would accept.
		if spec.OIDC.ForwardIDToken != nil {
			header := spec.OIDC.ForwardIDToken.Header
			if strings.EqualFold(header, "Authorization") {
				return errors.New("authentication.anyOf does not support oidc.forwardIDToken on the Authorization header")
			}
			for _, provider := range spec.JWT.Providers {
				if provider.ExtractFrom == nil {
					continue
				}
				for _, extractor := range provider.ExtractFrom.Headers {
					if strings.EqualFold(header, extractor.Name) {
						return fmt.Errorf("authentication.anyOf does not support oidc.forwardIDToken on header %q, which JWT provider %q reads", header, provider.Name)
					}
				}
			}
		}
	}
	// Envoy enables per-route filters from the original route, while the gate
	// reads its rules from the recomputed one, so the two can disagree. ext_authz
	// and jwt_authn can both recompute the route.
	if spec.ExtAuth != nil && ptr.Deref(spec.ExtAuth.RecomputeRoute, false) {
		return errors.New("authentication.anyOf does not support extAuth.recomputeRoute")
	}
	// basic_auth leaves a client-supplied username header untouched when it
	// passes a request without Basic credentials, so a JWT client could set it.
	if spec.BasicAuth.ForwardUsernameHeader != nil {
		return errors.New("authentication.anyOf does not support basicAuth.forwardUsernameHeader")
	}
	if ptr.Deref(spec.JWT.FailOpen, false) {
		return errors.New("authentication.anyOf does not support jwt.failOpen")
	}
	for _, provider := range spec.JWT.Providers {
		if ptr.Deref(provider.RecomputeRoute, false) {
			return fmt.Errorf("authentication.anyOf does not support recomputeRoute on JWT provider %q", provider.Name)
		}
		// jwt_authn writes the provider's payload under its name, which would
		// overwrite the verified-token marker that the authentication gate reads.
		if provider.Name == ir.AuthenticationAnyJWTMetadataKey {
			return fmt.Errorf("authentication.anyOf reserves the JWT provider name %q", provider.Name)
		}
		// jwt_authn extracts a token wherever valuePrefix occurs in the header
		// value and rejects a token that fails verification, so an extractor that
		// can match a Basic credential rejects every BasicAuth request.
		if provider.ExtractFrom != nil {
			for _, extractor := range provider.ExtractFrom.Headers {
				if strings.EqualFold(extractor.Name, "Authorization") && ptr.Deref(extractor.ValuePrefix, "") != "Bearer " {
					return fmt.Errorf("authentication.anyOf requires valuePrefix \"Bearer \" on the Authorization header extractor of JWT provider %q", provider.Name)
				}
			}
		}
	}
	return nil
}

// basicAuthName keeps anyOf BasicAuth in a separate filter from one built
// without anyOf from the same owner policy, since allow_missing is a
// filter-wide setting.
func basicAuthName(policy, ownerPolicy *egv1a1.SecurityPolicy) string {
	name := irConfigName(ownerPolicy)
	if authenticationAny(policy) {
		name += authenticationAnySuffix
	}
	return name
}
