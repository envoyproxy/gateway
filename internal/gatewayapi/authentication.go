// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"errors"
	"fmt"

	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

const authenticationAnySuffix = "/authentication-any"

func authenticationAny(policy *egv1a1.SecurityPolicy) bool {
	return policy.Spec.Authentication != nil && policy.Spec.Authentication.Mode == egv1a1.AuthenticationModeAny
}

// validateAuthentication runs on the effective (merged) policy, so it also covers
// configurations that admission cannot see, such as JWT or BasicAuth inherited
// from a parent policy.
func validateAuthentication(policy *egv1a1.SecurityPolicy) error {
	if policy.Spec.Authentication != nil {
		switch policy.Spec.Authentication.Mode {
		case "", egv1a1.AuthenticationModeAll, egv1a1.AuthenticationModeAny:
		default:
			return fmt.Errorf("unsupported authentication mode %q", policy.Spec.Authentication.Mode)
		}
	}
	if !authenticationAny(policy) {
		return nil
	}
	spec := policy.Spec
	if spec.JWT == nil || len(spec.JWT.Providers) == 0 || spec.BasicAuth == nil {
		return errors.New("authentication mode Any requires jwt and basicAuth")
	}
	if spec.OIDC != nil || spec.APIKeyAuth != nil {
		return errors.New("authentication mode Any does not support oidc or apiKeyAuth")
	}
	// basic_auth leaves a client-supplied username header untouched when it
	// passes a request without Basic credentials, so a JWT client could set it.
	if spec.BasicAuth.ForwardUsernameHeader != nil {
		return errors.New("authentication mode Any does not support basicAuth.forwardUsernameHeader")
	}
	if ptr.Deref(spec.JWT.FailOpen, false) {
		return errors.New("authentication mode Any does not support jwt.failOpen")
	}
	for _, provider := range spec.JWT.Providers {
		// The gate's per-route config is selected before a recomputed route takes effect.
		if ptr.Deref(provider.RecomputeRoute, false) {
			return fmt.Errorf("authentication mode Any does not support recomputeRoute on JWT provider %q", provider.Name)
		}
	}
	return nil
}

// authenticationBasicName keeps Any-mode BasicAuth in a separate filter from an
// All-mode filter built from the same owner policy, since allow_missing is a
// filter-wide setting.
func authenticationBasicName(policy *egv1a1.SecurityPolicy, owners *securityPolicyOwners) string {
	name := irConfigName(policyOwnerOr(owners.basicAuth, policy))
	if authenticationAny(policy) {
		name += authenticationAnySuffix
	}
	return name
}
