// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build celvalidation

package celvalidation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func TestAuthenticationAnyOf(t *testing.T) {
	routeTarget := egv1a1.PolicyTargetReferences{TargetRefs: []gwapiv1.LocalPolicyTargetReferenceWithSectionName{{
		LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{Group: "gateway.networking.k8s.io", Kind: "HTTPRoute", Name: "backend"},
	}}}
	jwt := &egv1a1.JWT{Providers: []egv1a1.JWTProvider{{
		Name:       "example",
		Issuer:     "https://issuer.example",
		RemoteJWKS: &egv1a1.RemoteJWKS{URI: "https://issuer.example/keys"},
	}}}
	extAuth := func() *egv1a1.ExtAuth {
		return &egv1a1.ExtAuth{HTTP: &egv1a1.HTTPExtAuthService{
			BackendCluster: egv1a1.BackendCluster{BackendRefs: []egv1a1.BackendRef{{
				BackendObjectReference: gwapiv1.BackendObjectReference{Name: "authz", Port: new(gwapiv1.PortNumber(80))},
			}}},
		}}
	}
	oidc := func() *egv1a1.OIDC {
		return &egv1a1.OIDC{
			Provider:     egv1a1.OIDCProvider{Issuer: "https://issuer.example"},
			ClientID:     new("client"),
			ClientSecret: gwapiv1.SecretObjectReference{Name: "client-secret"},
		}
	}
	basicAuth := func() *egv1a1.BasicAuth {
		return &egv1a1.BasicAuth{Users: gwapiv1.SecretObjectReference{Name: "users"}}
	}

	for _, test := range []struct {
		name    string
		spec    func(*egv1a1.SecurityPolicySpec)
		wantErr string
	}{
		{name: "any of", spec: func(*egv1a1.SecurityPolicySpec) {}},
		{name: "without anyOf", spec: func(spec *egv1a1.SecurityPolicySpec) { spec.Authentication.AnyOf = nil }},
		{
			name:    "unknown method",
			spec:    func(spec *egv1a1.SecurityPolicySpec) { spec.Authentication.AnyOf[1] = "APIKeyAuth" },
			wantErr: "supported values",
		},
		{
			name:    "single method",
			spec:    func(spec *egv1a1.SecurityPolicySpec) { spec.Authentication.AnyOf = spec.Authentication.AnyOf[:1] },
			wantErr: "should have at least 2 items",
		},
		{
			name:    "duplicate method",
			spec:    func(spec *egv1a1.SecurityPolicySpec) { spec.Authentication.AnyOf[1] = egv1a1.AuthenticationMethodJWT },
			wantErr: "Duplicate value",
		},
		{name: "without jwt", spec: func(spec *egv1a1.SecurityPolicySpec) { spec.JWT = nil }, wantErr: "every method in authentication.anyOf must be configured"},
		{name: "without basic", spec: func(spec *egv1a1.SecurityPolicySpec) { spec.BasicAuth = nil }, wantErr: "every method in authentication.anyOf must be configured"},
		{
			name: "inherits methods with mergeType",
			spec: func(spec *egv1a1.SecurityPolicySpec) {
				spec.JWT, spec.BasicAuth = nil, nil
				spec.MergeType = new(egv1a1.JSONMerge)
			},
		},
		{
			name: "required apiKeyAuth",
			spec: func(spec *egv1a1.SecurityPolicySpec) {
				spec.APIKeyAuth = &egv1a1.APIKeyAuth{
					CredentialRefs: []gwapiv1.SecretObjectReference{{Name: "keys"}},
					ExtractFrom:    []*egv1a1.ExtractFrom{{Headers: []string{"x-api-key"}}},
				}
			},
		},
		{
			name: "required extAuth",
			spec: func(spec *egv1a1.SecurityPolicySpec) { spec.ExtAuth = extAuth() },
		},
		{
			name: "extAuth recomputeRoute",
			spec: func(spec *egv1a1.SecurityPolicySpec) {
				spec.ExtAuth = extAuth()
				spec.ExtAuth.RecomputeRoute = new(true)
			},
			wantErr: "authentication.anyOf does not support extAuth.recomputeRoute",
		},
		{name: "required oidc", spec: func(spec *egv1a1.SecurityPolicySpec) { spec.OIDC = oidc() }},
		{
			name: "oidc forwardAccessToken",
			spec: func(spec *egv1a1.SecurityPolicySpec) {
				spec.OIDC = oidc()
				spec.OIDC.ForwardAccessToken = new(true)
			},
			wantErr: "authentication.anyOf does not support oidc.forwardAccessToken or oidc.passThroughAuthHeader",
		},
		{
			name: "oidc passThroughAuthHeader",
			spec: func(spec *egv1a1.SecurityPolicySpec) {
				spec.OIDC = oidc()
				spec.OIDC.PassThroughAuthHeader = new(true)
			},
			wantErr: "authentication.anyOf does not support oidc.forwardAccessToken or oidc.passThroughAuthHeader",
		},
		{
			name: "oidc forwardIDToken on Authorization",
			spec: func(spec *egv1a1.SecurityPolicySpec) {
				spec.OIDC = oidc()
				spec.OIDC.ForwardIDToken = &egv1a1.OIDCTokenForwarding{Header: "Authorization"}
			},
			wantErr: "authentication.anyOf does not support oidc.forwardIDToken on the Authorization header",
		},
		{
			name: "oidc forwardIDToken on other header",
			spec: func(spec *egv1a1.SecurityPolicySpec) {
				spec.OIDC = oidc()
				spec.OIDC.ForwardIDToken = &egv1a1.OIDCTokenForwarding{Header: "x-id-token"}
			},
		},
		{
			name:    "jwt failOpen",
			spec:    func(spec *egv1a1.SecurityPolicySpec) { spec.JWT.FailOpen = new(true) },
			wantErr: "authentication.anyOf with JWT does not support jwt.failOpen or recomputeRoute on JWT providers",
		},
		{
			name:    "jwt recomputeRoute",
			spec:    func(spec *egv1a1.SecurityPolicySpec) { spec.JWT.Providers[0].RecomputeRoute = new(true) },
			wantErr: "authentication.anyOf with JWT does not support jwt.failOpen or recomputeRoute on JWT providers",
		},
		{
			name: "jwt failOpen without anyOf",
			spec: func(spec *egv1a1.SecurityPolicySpec) {
				spec.Authentication = nil
				spec.JWT.FailOpen = new(true)
			},
		},
		{
			name:    "reserved jwt provider name",
			spec:    func(spec *egv1a1.SecurityPolicySpec) { spec.JWT.Providers[0].Name = "authentication-any" },
			wantErr: "authentication.anyOf with JWT reserves the JWT provider name authentication-any",
		},
		{
			name: "reserved jwt provider name without anyOf",
			spec: func(spec *egv1a1.SecurityPolicySpec) {
				spec.Authentication = nil
				spec.JWT.Providers[0].Name = "authentication-any"
			},
		},
		{
			name:    "forwarded username",
			spec:    func(spec *egv1a1.SecurityPolicySpec) { spec.BasicAuth.ForwardUsernameHeader = new("x-user") },
			wantErr: "authentication.anyOf with BasicAuth does not support basicAuth.forwardUsernameHeader",
		},
		{
			name: "forwarded username without anyOf",
			spec: func(spec *egv1a1.SecurityPolicySpec) {
				spec.Authentication = nil
				spec.BasicAuth.ForwardUsernameHeader = new("x-user")
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			policy := &egv1a1.SecurityPolicy{
				ObjectMeta: metav1.ObjectMeta{GenerateName: "authentication-", Namespace: "default"},
				Spec: egv1a1.SecurityPolicySpec{
					PolicyTargetReferences: routeTarget,
					Authentication:         &egv1a1.Authentication{AnyOf: []egv1a1.AuthenticationMethod{egv1a1.AuthenticationMethodJWT, egv1a1.AuthenticationMethodBasicAuth}},
					JWT:                    jwt.DeepCopy(),
					BasicAuth:              basicAuth(),
				},
			}
			test.spec(&policy.Spec)

			err := c.Create(t.Context(), policy)
			if test.wantErr != "" {
				require.ErrorContains(t, err, test.wantErr)
				return
			}
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, c.Delete(context.WithoutCancel(t.Context()), policy)) })
		})
	}
}
