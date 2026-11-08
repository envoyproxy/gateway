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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func TestAuthenticationAPI(t *testing.T) {
	for _, test := range []struct {
		name    string
		mode    egv1a1.AuthenticationMode
		users   []string
		invalid bool
	}{
		{"any", egv1a1.AuthenticationModeAny, []string{"alice"}, false},
		{"all", egv1a1.AuthenticationModeAll, []string{"alice"}, false},
		{"default", "", []string{"alice"}, false},
		{"unknown", "Unknown", []string{"alice"}, true},
		{"empty-users", egv1a1.AuthenticationModeAny, []string{}, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			policy := &egv1a1.SecurityPolicy{ObjectMeta: metav1.ObjectMeta{Name: "authentication-" + test.name, Namespace: "default"}, Spec: egv1a1.SecurityPolicySpec{
				PolicyTargetReferences: egv1a1.PolicyTargetReferences{TargetRefs: []gwapiv1.LocalPolicyTargetReferenceWithSectionName{{LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{Group: "gateway.networking.k8s.io", Kind: "HTTPRoute", Name: "backend"}}}},
				Authentication:         &egv1a1.Authentication{Mode: test.mode},
				JWT:                    &egv1a1.JWT{Providers: []egv1a1.JWTProvider{{Name: "example", Issuer: "https://issuer.example", RemoteJWKS: &egv1a1.RemoteJWKS{URI: "https://issuer.example/keys"}}}},
				BasicAuth:              &egv1a1.BasicAuth{Users: gwapiv1.SecretObjectReference{Name: "users"}},
				Authorization:          &egv1a1.Authorization{Rules: []egv1a1.AuthorizationRule{{Action: egv1a1.AuthorizationActionAllow, Principal: &egv1a1.Principal{BasicAuth: &egv1a1.BasicAuthPrincipal{Users: test.users}}}}},
			}}
			err := c.Create(t.Context(), policy)
			if test.invalid {
				require.True(t, apierrors.IsInvalid(err), "expected admission rejection, got %v", err)
				return
			}
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, c.Delete(context.WithoutCancel(t.Context()), policy)) })
			if test.mode == "" {
				require.Equal(t, egv1a1.AuthenticationModeAll, policy.Spec.Authentication.Mode)
			}
		})
	}
}
