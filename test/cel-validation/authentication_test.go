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

func TestAuthenticationMode(t *testing.T) {
	routeTarget := egv1a1.PolicyTargetReferences{TargetRefs: []gwapiv1.LocalPolicyTargetReferenceWithSectionName{{
		LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{Group: "gateway.networking.k8s.io", Kind: "HTTPRoute", Name: "backend"},
	}}}
	jwt := &egv1a1.JWT{Providers: []egv1a1.JWTProvider{{
		Name:       "example",
		Issuer:     "https://issuer.example",
		RemoteJWKS: &egv1a1.RemoteJWKS{URI: "https://issuer.example/keys"},
	}}}
	basicAuth := func() *egv1a1.BasicAuth {
		return &egv1a1.BasicAuth{Users: gwapiv1.SecretObjectReference{Name: "users"}}
	}

	for _, test := range []struct {
		name    string
		spec    func(*egv1a1.SecurityPolicySpec)
		wantErr string
	}{
		{name: "any", spec: func(*egv1a1.SecurityPolicySpec) {}},
		{name: "all", spec: func(spec *egv1a1.SecurityPolicySpec) { spec.Authentication.Mode = egv1a1.AuthenticationModeAll }},
		{name: "default", spec: func(spec *egv1a1.SecurityPolicySpec) { spec.Authentication.Mode = "" }},
		{name: "unknown", spec: func(spec *egv1a1.SecurityPolicySpec) { spec.Authentication.Mode = "Unknown" }, wantErr: "supported values"},
		{name: "any without jwt", spec: func(spec *egv1a1.SecurityPolicySpec) { spec.JWT = nil }, wantErr: "authentication.mode Any requires jwt and basicAuth"},
		{name: "any without basic", spec: func(spec *egv1a1.SecurityPolicySpec) { spec.BasicAuth = nil }, wantErr: "authentication.mode Any requires jwt and basicAuth"},
		{
			name: "any inherits methods with mergeType",
			spec: func(spec *egv1a1.SecurityPolicySpec) {
				spec.JWT, spec.BasicAuth = nil, nil
				spec.MergeType = new(egv1a1.JSONMerge)
			},
		},
		{
			name: "any with apiKeyAuth",
			spec: func(spec *egv1a1.SecurityPolicySpec) {
				spec.APIKeyAuth = &egv1a1.APIKeyAuth{
					CredentialRefs: []gwapiv1.SecretObjectReference{{Name: "keys"}},
					ExtractFrom:    []*egv1a1.ExtractFrom{{Headers: []string{"x-api-key"}}},
				}
			},
			wantErr: "authentication.mode Any does not support oidc, apiKeyAuth or extAuth",
		},
		{
			name: "any with extAuth",
			spec: func(spec *egv1a1.SecurityPolicySpec) {
				spec.ExtAuth = &egv1a1.ExtAuth{HTTP: &egv1a1.HTTPExtAuthService{
					BackendCluster: egv1a1.BackendCluster{BackendRefs: []egv1a1.BackendRef{{
						BackendObjectReference: gwapiv1.BackendObjectReference{Name: "authz", Port: new(gwapiv1.PortNumber(80))},
					}}},
				}}
			},
			wantErr: "authentication.mode Any does not support oidc, apiKeyAuth or extAuth",
		},
		{
			name:    "any with jwt failOpen",
			spec:    func(spec *egv1a1.SecurityPolicySpec) { spec.JWT.FailOpen = new(true) },
			wantErr: "authentication.mode Any does not support jwt.failOpen or recomputeRoute on JWT providers",
		},
		{
			name:    "any with jwt recomputeRoute",
			spec:    func(spec *egv1a1.SecurityPolicySpec) { spec.JWT.Providers[0].RecomputeRoute = new(true) },
			wantErr: "authentication.mode Any does not support jwt.failOpen or recomputeRoute on JWT providers",
		},
		{
			name: "all with jwt failOpen",
			spec: func(spec *egv1a1.SecurityPolicySpec) {
				spec.Authentication.Mode = egv1a1.AuthenticationModeAll
				spec.JWT.FailOpen = new(true)
			},
		},
		{
			name:    "any with forwarded username",
			spec:    func(spec *egv1a1.SecurityPolicySpec) { spec.BasicAuth.ForwardUsernameHeader = new("x-user") },
			wantErr: "authentication.mode Any does not support basicAuth.forwardUsernameHeader",
		},
		{
			name: "all with forwarded username",
			spec: func(spec *egv1a1.SecurityPolicySpec) {
				spec.Authentication.Mode = egv1a1.AuthenticationModeAll
				spec.BasicAuth.ForwardUsernameHeader = new("x-user")
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			policy := &egv1a1.SecurityPolicy{
				ObjectMeta: metav1.ObjectMeta{GenerateName: "authentication-", Namespace: "default"},
				Spec: egv1a1.SecurityPolicySpec{
					PolicyTargetReferences: routeTarget,
					Authentication:         &egv1a1.Authentication{Mode: egv1a1.AuthenticationModeAny},
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
			if test.name == "default" {
				require.Equal(t, egv1a1.AuthenticationModeAll, policy.Spec.Authentication.Mode)
			}
		})
	}
}
