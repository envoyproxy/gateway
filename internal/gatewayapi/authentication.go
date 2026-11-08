// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
)

func authenticationAny(policy *egv1a1.SecurityPolicy) bool {
	return policy.Spec.Authentication != nil && policy.Spec.Authentication.Mode == egv1a1.AuthenticationModeAny
}

func authenticationMetadataRequired(policy *egv1a1.SecurityPolicy) bool {
	if authenticationAny(policy) || (policy.Spec.ExtAuth != nil && ptr.Deref(policy.Spec.ExtAuth.IncludeAuthenticationMetadata, false)) {
		return true
	}
	if policy.Spec.Authorization != nil {
		for _, rule := range policy.Spec.Authorization.Rules {
			if rule.Principal != nil && rule.Principal.BasicAuth != nil {
				return true
			}
		}
	}
	return false
}

func validateAuthentication(policy *egv1a1.SecurityPolicy) error {
	if policy.Spec.Authentication != nil {
		switch policy.Spec.Authentication.Mode {
		case "", egv1a1.AuthenticationModeAll, egv1a1.AuthenticationModeAny:
		default:
			return errors.New("unsupported authentication mode")
		}
	}
	any := authenticationAny(policy)
	delivery := policy.Spec.ExtAuth != nil && ptr.Deref(policy.Spec.ExtAuth.IncludeAuthenticationMetadata, false)
	if any && (policy.Spec.JWT == nil || len(policy.Spec.JWT.Providers) == 0 || policy.Spec.BasicAuth == nil || policy.Spec.OIDC != nil || policy.Spec.APIKeyAuth != nil) {
		return errors.New("authentication Any requires JWT and BasicAuth without OIDC or APIKeyAuth")
	}
	if delivery {
		external := policy.Spec.ExtAuth
		if external.GRPC == nil || external.HTTP != nil || ptr.Deref(external.FailOpen, false) || ptr.Deref(external.RecomputeRoute, false) {
			return errors.New("authentication metadata requires gRPC extAuth with failOpen and recomputeRoute disabled")
		}
		if policy.Spec.JWT == nil && policy.Spec.BasicAuth == nil {
			return errors.New("authentication metadata requires JWT or BasicAuth")
		}
		if policy.Spec.OIDC != nil || policy.Spec.APIKeyAuth != nil {
			return errors.New("authentication metadata supports JWT and BasicAuth only")
		}
	}
	if any || delivery {
		if policy.Spec.JWT != nil {
			if ptr.Deref(policy.Spec.JWT.FailOpen, false) {
				return errors.New("verified authentication cannot use jwt.failOpen")
			}
			for _, provider := range policy.Spec.JWT.Providers {
				if ptr.Deref(provider.RecomputeRoute, false) {
					return fmt.Errorf("verified authentication cannot recompute routes for JWT provider %q", provider.Name)
				}
			}
		}
		if policy.Spec.ExtAuth != nil && ptr.Deref(policy.Spec.ExtAuth.RecomputeRoute, false) {
			return errors.New("verified authentication cannot recompute routes in extAuth")
		}
	}
	return nil
}

func authenticationExtAuthMetadata(external *ir.ExtAuth, jwt *ir.JWT, basic *ir.BasicAuth, policy *egv1a1.SecurityPolicy) *ir.ExtAuth {
	if external == nil || policy.Spec.ExtAuth == nil || !ptr.Deref(policy.Spec.ExtAuth.IncludeAuthenticationMetadata, false) {
		return external
	}
	external = external.DeepCopy()
	external.Name += "/authentication-metadata"
	if basic != nil {
		external.Name += "/" + basic.Name
	}
	if jwt != nil {
		external.Name += "/" + jwt.Name
	}
	if jwt != nil {
		external.AuthenticationMetadataNamespaces = append(external.AuthenticationMetadataNamespaces, "envoy.filters.http.jwt_authn")
	}
	if basic != nil {
		external.AuthenticationMetadataNamespaces = append(external.AuthenticationMetadataNamespaces, "envoy.filters.http.basic_auth/"+basic.Name)
	}
	return external
}

func authenticationBasicName(policy *egv1a1.SecurityPolicy, owners *securityPolicyOwners) string {
	name := irConfigName(policyOwnerOr(owners.basicAuth, policy))
	if authenticationAny(policy) {
		return name + "/authentication-any"
	}
	if authenticationMetadataRequired(policy) {
		return name + "/authentication-metadata"
	}
	return name
}

func authenticationJWTName(policy *egv1a1.SecurityPolicy, owners *securityPolicyOwners) string {
	name := jwtName(policy, owners)
	if authenticationAny(policy) {
		name += "/authentication-any"
		if len(name) > maxJWTNameLength {
			sum := sha256.Sum256([]byte(name))
			suffix := "_" + hex.EncodeToString(sum[:8])
			name = name[:maxJWTNameLength-len(suffix)] + suffix
		}
		return name
	}
	return name
}

func validateAuthenticationOrder(policy *egv1a1.SecurityPolicy, proxy *egv1a1.EnvoyProxy) error {
	if !authenticationMetadataRequired(policy) {
		return nil
	}
	delivery := policy.Spec.ExtAuth != nil && ptr.Deref(policy.Spec.ExtAuth.IncludeAuthenticationMetadata, false)
	externalAfterAuthentication := false
	after := egv1a1.EnvoyFilterJWTAuthn
	if policy.Spec.JWT == nil {
		after = egv1a1.EnvoyFilterBasicAuth
	}
	if proxy != nil {
		for _, position := range proxy.Spec.FilterOrder {
			switch position.Name {
			case egv1a1.EnvoyFilterJWTAuthn, egv1a1.EnvoyFilterBasicAuth, egv1a1.EnvoyFilterRBAC:
				return errors.New("verified authentication does not support custom JWT, BasicAuth or RBAC ordering")
			case egv1a1.EnvoyFilterExtAuthz:
				if delivery {
					if position.After == nil || *position.After != after || position.Before != nil {
						return errors.New("authentication metadata requires extAuth filterOrder after the final authentication method")
					}
					externalAfterAuthentication = true
				}
			}
		}
	}
	if delivery && !externalAfterAuthentication {
		return errors.New("authentication metadata requires extAuth filterOrder after the final authentication method")
	}
	return nil
}
