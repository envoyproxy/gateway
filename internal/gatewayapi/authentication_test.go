// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
)

func authenticationPolicy() *egv1a1.SecurityPolicy {
	return &egv1a1.SecurityPolicy{ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"}, Spec: egv1a1.SecurityPolicySpec{
		Authentication: &egv1a1.Authentication{Mode: egv1a1.AuthenticationModeAny},
		JWT:            &egv1a1.JWT{Providers: []egv1a1.JWTProvider{{Name: "issuer", Audiences: []string{"test-api"}}}}, BasicAuth: &egv1a1.BasicAuth{},
	}}
}

func TestAuthenticationValidation(t *testing.T) {
	for _, test := range []struct {
		name    string
		change  func(*egv1a1.SecurityPolicy)
		invalid bool
	}{
		{"Any", func(*egv1a1.SecurityPolicy) {}, false},
		{"omitted mode", func(policy *egv1a1.SecurityPolicy) { policy.Spec.Authentication = nil }, false},
		{"All", func(policy *egv1a1.SecurityPolicy) { policy.Spec.Authentication.Mode = egv1a1.AuthenticationModeAll }, false},
		{"unknown mode", func(policy *egv1a1.SecurityPolicy) { policy.Spec.Authentication.Mode = "Unknown" }, true},
		{"missing JWT", func(policy *egv1a1.SecurityPolicy) { policy.Spec.JWT = nil }, true},
		{"missing providers", func(policy *egv1a1.SecurityPolicy) { policy.Spec.JWT.Providers = nil }, true},
		{"missing Basic", func(policy *egv1a1.SecurityPolicy) { policy.Spec.BasicAuth = nil }, true},
		{"OIDC", func(policy *egv1a1.SecurityPolicy) { policy.Spec.OIDC = &egv1a1.OIDC{} }, true},
		{"API key", func(policy *egv1a1.SecurityPolicy) { policy.Spec.APIKeyAuth = &egv1a1.APIKeyAuth{} }, true},
		{"JWT fail open", func(policy *egv1a1.SecurityPolicy) { policy.Spec.JWT.FailOpen = new(true) }, true},
		{"JWT route recomputation", func(policy *egv1a1.SecurityPolicy) { policy.Spec.JWT.Providers[0].RecomputeRoute = new(true) }, true},
		{"external route recomputation", func(policy *egv1a1.SecurityPolicy) { policy.Spec.ExtAuth = &egv1a1.ExtAuth{RecomputeRoute: new(true)} }, true},
		{"external gRPC", func(policy *egv1a1.SecurityPolicy) {
			policy.Spec.ExtAuth = &egv1a1.ExtAuth{GRPC: &egv1a1.GRPCExtAuthService{}, IncludeAuthenticationMetadata: new(true)}
		}, false},
		{"external HTTP", func(policy *egv1a1.SecurityPolicy) {
			policy.Spec.ExtAuth = &egv1a1.ExtAuth{HTTP: &egv1a1.HTTPExtAuthService{}, IncludeAuthenticationMetadata: new(true)}
		}, true},
		{"external fail open", func(policy *egv1a1.SecurityPolicy) {
			policy.Spec.ExtAuth = &egv1a1.ExtAuth{GRPC: &egv1a1.GRPCExtAuthService{}, IncludeAuthenticationMetadata: new(true), FailOpen: new(true)}
		}, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			policy := authenticationPolicy()
			test.change(policy)
			err := validateAuthentication(policy)
			if test.invalid {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAuthenticationPrincipalAndOwnership(t *testing.T) {
	policy := authenticationPolicy()
	parent := policy.DeepCopy()
	parent.Name = "parent"
	parent.Spec.Authentication = nil
	parent.Spec.Authorization = &egv1a1.Authorization{Rules: []egv1a1.AuthorizationRule{{Action: egv1a1.AuthorizationActionAllow, Principal: &egv1a1.Principal{BasicAuth: &egv1a1.BasicAuthPrincipal{Users: []string{"alice"}}}}}}
	policy.Spec.Authorization = parent.Spec.Authorization.DeepCopy()
	owners := &securityPolicyOwners{basicAuth: parent}
	translator := &Translator{}
	authorization, err := translator.buildAuthorization(policy, owners)
	require.NoError(t, err)
	require.Equal(t, "envoy.filters.http.basic_auth/"+authenticationBasicName(policy, owners), authorization.Rules[0].Principal.BasicAuthFilter)
	require.NotEqual(t, authenticationBasicName(parent, &securityPolicyOwners{}), authenticationBasicName(policy, owners))
	policy.Spec.BasicAuth = nil
	_, err = translator.buildAuthorization(policy, owners)
	require.ErrorContains(t, err, "requires basicAuth")
}

func TestAuthenticationMetadataAndOrdering(t *testing.T) {
	policy := authenticationPolicy()
	policy.Spec.ExtAuth = &egv1a1.ExtAuth{GRPC: &egv1a1.GRPCExtAuthService{}, IncludeAuthenticationMetadata: new(true)}
	external := &ir.ExtAuth{Name: "external"}
	basic := &ir.BasicAuth{Name: "basic"}
	jwt := &ir.JWT{Name: "jwt"}
	result := authenticationExtAuthMetadata(external, jwt, basic, policy)
	require.Empty(t, external.AuthenticationMetadataNamespaces)
	require.Equal(t, []string{"envoy.filters.http.jwt_authn", "envoy.filters.http.basic_auth/basic"}, result.AuthenticationMetadataNamespaces)
	require.Error(t, validateAuthenticationOrder(policy, nil))
	proxy := &egv1a1.EnvoyProxy{Spec: egv1a1.EnvoyProxySpec{FilterOrder: []egv1a1.FilterPosition{{Name: egv1a1.EnvoyFilterExtAuthz, After: new(egv1a1.EnvoyFilterJWTAuthn)}}}}
	require.NoError(t, validateAuthenticationOrder(policy, proxy))
	proxy.Spec.FilterOrder = append(proxy.Spec.FilterOrder, egv1a1.FilterPosition{Name: egv1a1.EnvoyFilterRBAC, Before: new(egv1a1.EnvoyFilterJWTAuthn)})
	require.Error(t, validateAuthenticationOrder(policy, proxy))
}

func TestAuthenticationLongJWTNames(t *testing.T) {
	policy := authenticationPolicy()
	policy.Name = strings.Repeat("a", 253)
	parent := policy.DeepCopy()
	parent.Name = strings.Repeat("b", 253)
	owners := &securityPolicyOwners{mergedParent: parent}
	first := authenticationJWTName(policy, owners)
	require.LessOrEqual(t, len(first), maxJWTNameLength)
	parent.Name = strings.Repeat("c", 253)
	require.NotEqual(t, first, authenticationJWTName(policy, owners))
}

func TestAuthenticationMetadataForSingleMethod(t *testing.T) {
	for _, method := range []string{"jwt", "basic"} {
		t.Run(method, func(t *testing.T) {
			policy := authenticationPolicy()
			policy.Spec.Authentication = nil
			policy.Spec.ExtAuth = &egv1a1.ExtAuth{GRPC: &egv1a1.GRPCExtAuthService{}, IncludeAuthenticationMetadata: new(true)}
			basic := &ir.BasicAuth{Name: "basic"}
			jwt := &ir.JWT{Name: "jwt"}
			after := egv1a1.EnvoyFilterJWTAuthn
			expected := []string{"envoy.filters.http.jwt_authn"}
			if method == "jwt" {
				policy.Spec.BasicAuth = nil
				basic = nil
			} else {
				policy.Spec.JWT = nil
				jwt = nil
				after = egv1a1.EnvoyFilterBasicAuth
				expected = []string{"envoy.filters.http.basic_auth/basic"}
			}
			require.NoError(t, validateAuthentication(policy))
			proxy := &egv1a1.EnvoyProxy{Spec: egv1a1.EnvoyProxySpec{FilterOrder: []egv1a1.FilterPosition{{Name: egv1a1.EnvoyFilterExtAuthz, After: &after}}}}
			require.NoError(t, validateAuthenticationOrder(policy, proxy))
			result := authenticationExtAuthMetadata(&ir.ExtAuth{Name: "external"}, jwt, basic, policy)
			require.Equal(t, expected, result.AuthenticationMetadataNamespaces)
			policy.Spec.JWT = nil
			policy.Spec.BasicAuth = nil
			require.Error(t, validateAuthentication(policy))
		})
	}
}
