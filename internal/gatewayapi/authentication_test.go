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
)

func authenticationAnyPolicy() *egv1a1.SecurityPolicy {
	return &egv1a1.SecurityPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: egv1a1.SecurityPolicySpec{
			Authentication: &egv1a1.Authentication{AnyOf: []egv1a1.AuthenticationMethod{egv1a1.AuthenticationMethodJWT, egv1a1.AuthenticationMethodBasicAuth}},
			JWT:            &egv1a1.JWT{Providers: []egv1a1.JWTProvider{{Name: "issuer", Audiences: []string{"test-api"}}}},
			BasicAuth:      &egv1a1.BasicAuth{},
		},
	}
}

func TestValidateAuthentication(t *testing.T) {
	for _, test := range []struct {
		name    string
		change  func(*egv1a1.SecurityPolicy)
		wantErr string
	}{
		{name: "any of", change: func(*egv1a1.SecurityPolicy) {}},
		{name: "omitted", change: func(policy *egv1a1.SecurityPolicy) { policy.Spec.Authentication = nil }},
		{name: "empty", change: func(policy *egv1a1.SecurityPolicy) { policy.Spec.Authentication.AnyOf = nil }},
		{
			name: "without anyOf ignores anyOf restrictions",
			change: func(policy *egv1a1.SecurityPolicy) {
				policy.Spec.Authentication = nil
				policy.Spec.BasicAuth.ForwardUsernameHeader = new("x-user")
			},
		},
		{
			name: "single method",
			change: func(policy *egv1a1.SecurityPolicy) {
				policy.Spec.Authentication.AnyOf = []egv1a1.AuthenticationMethod{egv1a1.AuthenticationMethodJWT}
			},
			wantErr: "must list JWT and BasicAuth",
		},
		{
			name: "duplicate method",
			change: func(policy *egv1a1.SecurityPolicy) {
				policy.Spec.Authentication.AnyOf = []egv1a1.AuthenticationMethod{egv1a1.AuthenticationMethodJWT, egv1a1.AuthenticationMethodJWT}
			},
			wantErr: "must list JWT and BasicAuth",
		},
		{name: "missing jwt", change: func(policy *egv1a1.SecurityPolicy) { policy.Spec.JWT = nil }, wantErr: "requires jwt and basicAuth"},
		{name: "missing providers", change: func(policy *egv1a1.SecurityPolicy) { policy.Spec.JWT.Providers = nil }, wantErr: "requires jwt and basicAuth"},
		{name: "missing basic", change: func(policy *egv1a1.SecurityPolicy) { policy.Spec.BasicAuth = nil }, wantErr: "requires jwt and basicAuth"},
		{name: "required oidc", change: func(policy *egv1a1.SecurityPolicy) { policy.Spec.OIDC = &egv1a1.OIDC{} }},
		{
			name:    "oidc forwarded access token",
			change:  func(policy *egv1a1.SecurityPolicy) { policy.Spec.OIDC = &egv1a1.OIDC{ForwardAccessToken: new(true)} },
			wantErr: "oidc.forwardAccessToken",
		},
		{
			name:    "oidc pass-through",
			change:  func(policy *egv1a1.SecurityPolicy) { policy.Spec.OIDC = &egv1a1.OIDC{PassThroughAuthHeader: new(true)} },
			wantErr: "oidc.passThroughAuthHeader",
		},
		{
			name: "oidc ID token on Authorization",
			change: func(policy *egv1a1.SecurityPolicy) {
				policy.Spec.OIDC = &egv1a1.OIDC{ForwardIDToken: &egv1a1.OIDCTokenForwarding{Header: "authorization"}}
			},
			wantErr: "oidc.forwardIDToken on the Authorization header",
		},
		{
			name: "oidc ID token on JWT extractor header",
			change: func(policy *egv1a1.SecurityPolicy) {
				policy.Spec.OIDC = &egv1a1.OIDC{ForwardIDToken: &egv1a1.OIDCTokenForwarding{Header: "X-Token"}}
				policy.Spec.JWT.Providers[0].ExtractFrom = &egv1a1.JWTExtractor{Headers: []egv1a1.JWTHeaderExtractor{{Name: "x-token"}}}
			},
			wantErr: `oidc.forwardIDToken on header "X-Token"`,
		},
		{
			name: "oidc ID token on other header",
			change: func(policy *egv1a1.SecurityPolicy) {
				policy.Spec.OIDC = &egv1a1.OIDC{ForwardIDToken: &egv1a1.OIDCTokenForwarding{Header: "x-id-token"}}
			},
		},
		{name: "required api key", change: func(policy *egv1a1.SecurityPolicy) { policy.Spec.APIKeyAuth = &egv1a1.APIKeyAuth{} }},
		{name: "required ext auth", change: func(policy *egv1a1.SecurityPolicy) { policy.Spec.ExtAuth = &egv1a1.ExtAuth{} }},
		{
			name:    "ext auth route recomputation",
			change:  func(policy *egv1a1.SecurityPolicy) { policy.Spec.ExtAuth = &egv1a1.ExtAuth{RecomputeRoute: new(true)} },
			wantErr: "extAuth.recomputeRoute",
		},
		{name: "forwarded username", change: func(policy *egv1a1.SecurityPolicy) { policy.Spec.BasicAuth.ForwardUsernameHeader = new("x-user") }, wantErr: "forwardUsernameHeader"},
		{name: "jwt fail open", change: func(policy *egv1a1.SecurityPolicy) { policy.Spec.JWT.FailOpen = new(true) }, wantErr: "jwt.failOpen"},
		{name: "jwt route recomputation", change: func(policy *egv1a1.SecurityPolicy) { policy.Spec.JWT.Providers[0].RecomputeRoute = new(true) }, wantErr: "recomputeRoute on JWT provider"},
		{name: "reserved jwt provider name", change: func(policy *egv1a1.SecurityPolicy) {
			policy.Spec.JWT.Providers[0].Name = "default/test/authentication-any"
		}, wantErr: `reserves JWT provider names ending in "/authentication-any"`},
		{name: "jwt provider name authentication-any", change: func(policy *egv1a1.SecurityPolicy) { policy.Spec.JWT.Providers[0].Name = "authentication-any" }},
		{
			name: "authorization extractor without prefix",
			change: func(policy *egv1a1.SecurityPolicy) {
				policy.Spec.JWT.Providers[0].ExtractFrom = &egv1a1.JWTExtractor{Headers: []egv1a1.JWTHeaderExtractor{{Name: "authorization"}}}
			},
			wantErr: `valuePrefix "Bearer " on the Authorization header extractor of JWT provider "issuer"`,
		},
		{
			name: "authorization extractor with basic prefix",
			change: func(policy *egv1a1.SecurityPolicy) {
				policy.Spec.JWT.Providers[0].ExtractFrom = &egv1a1.JWTExtractor{Headers: []egv1a1.JWTHeaderExtractor{{Name: "Authorization", ValuePrefix: new("Basic ")}}}
			},
			wantErr: `valuePrefix "Bearer "`,
		},
		{
			name: "authorization extractor with bearer prefix",
			change: func(policy *egv1a1.SecurityPolicy) {
				policy.Spec.JWT.Providers[0].ExtractFrom = &egv1a1.JWTExtractor{Headers: []egv1a1.JWTHeaderExtractor{{Name: "Authorization", ValuePrefix: new("Bearer ")}}}
			},
		},
		{
			name: "other header extractor without prefix",
			change: func(policy *egv1a1.SecurityPolicy) {
				policy.Spec.JWT.Providers[0].ExtractFrom = &egv1a1.JWTExtractor{Headers: []egv1a1.JWTHeaderExtractor{{Name: "x-token"}}}
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			policy := authenticationAnyPolicy()
			test.change(policy)
			err := validateAuthentication(policy)
			if test.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, test.wantErr)
		})
	}
}

func TestAuthenticationAnyNames(t *testing.T) {
	anyPolicy := authenticationAnyPolicy()
	allPolicy := anyPolicy.DeepCopy()
	allPolicy.Spec.Authentication = nil
	owners := &securityPolicyOwners{}

	require.NotEqual(t, basicAuthName(allPolicy, allPolicy), basicAuthName(anyPolicy, anyPolicy))
	require.Equal(t, irConfigName(allPolicy), basicAuthName(allPolicy, allPolicy))
	require.NotEqual(t, jwtName(allPolicy, owners), jwtName(anyPolicy, owners))

	parent := anyPolicy.DeepCopy()
	parent.Name = "parent"
	parent.Spec.Authentication = nil
	require.Equal(t, irConfigName(parent), basicAuthName(allPolicy, parent))
	require.Equal(t, irConfigName(parent)+authenticationAnySuffix, basicAuthName(anyPolicy, parent))
}

func TestAuthenticationAnyLongJWTNames(t *testing.T) {
	policy := authenticationAnyPolicy()
	policy.Name = strings.Repeat("a", 253)
	parent := policy.DeepCopy()
	parent.Name = strings.Repeat("b", 253)
	owners := &securityPolicyOwners{mergedParent: parent}

	first := jwtName(policy, owners)
	require.LessOrEqual(t, len(first), maxJWTNameLength)

	all := policy.DeepCopy()
	all.Spec.Authentication = nil
	require.NotEqual(t, first, jwtName(all, owners))

	parent.Name = strings.Repeat("c", 253)
	require.NotEqual(t, first, jwtName(policy, owners))
}
