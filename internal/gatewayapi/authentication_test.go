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
			Authentication: &egv1a1.Authentication{Mode: egv1a1.AuthenticationModeAny},
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
		{name: "any", change: func(*egv1a1.SecurityPolicy) {}},
		{name: "omitted", change: func(policy *egv1a1.SecurityPolicy) { policy.Spec.Authentication = nil }},
		{name: "all", change: func(policy *egv1a1.SecurityPolicy) { policy.Spec.Authentication.Mode = egv1a1.AuthenticationModeAll }},
		{
			name: "all ignores any-only restrictions",
			change: func(policy *egv1a1.SecurityPolicy) {
				policy.Spec.Authentication.Mode = ""
				policy.Spec.BasicAuth.ForwardUsernameHeader = new("x-user")
			},
		},
		{name: "unknown mode", change: func(policy *egv1a1.SecurityPolicy) { policy.Spec.Authentication.Mode = "Unknown" }, wantErr: "unsupported authentication mode"},
		{name: "missing jwt", change: func(policy *egv1a1.SecurityPolicy) { policy.Spec.JWT = nil }, wantErr: "requires jwt and basicAuth"},
		{name: "missing providers", change: func(policy *egv1a1.SecurityPolicy) { policy.Spec.JWT.Providers = nil }, wantErr: "requires jwt and basicAuth"},
		{name: "missing basic", change: func(policy *egv1a1.SecurityPolicy) { policy.Spec.BasicAuth = nil }, wantErr: "requires jwt and basicAuth"},
		{name: "oidc", change: func(policy *egv1a1.SecurityPolicy) { policy.Spec.OIDC = &egv1a1.OIDC{} }, wantErr: "oidc or apiKeyAuth"},
		{name: "api key", change: func(policy *egv1a1.SecurityPolicy) { policy.Spec.APIKeyAuth = &egv1a1.APIKeyAuth{} }, wantErr: "oidc or apiKeyAuth"},
		{name: "forwarded username", change: func(policy *egv1a1.SecurityPolicy) { policy.Spec.BasicAuth.ForwardUsernameHeader = new("x-user") }, wantErr: "forwardUsernameHeader"},
		{name: "jwt fail open", change: func(policy *egv1a1.SecurityPolicy) { policy.Spec.JWT.FailOpen = new(true) }, wantErr: "jwt.failOpen"},
		{name: "jwt route recomputation", change: func(policy *egv1a1.SecurityPolicy) { policy.Spec.JWT.Providers[0].RecomputeRoute = new(true) }, wantErr: "recomputeRoute"},
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

	require.NotEqual(t, authenticationBasicName(allPolicy, owners), authenticationBasicName(anyPolicy, owners))
	require.Equal(t, irConfigName(allPolicy), authenticationBasicName(allPolicy, owners))
	require.NotEqual(t, jwtName(allPolicy, owners), jwtName(anyPolicy, owners))

	parent := anyPolicy.DeepCopy()
	parent.Name = "parent"
	parent.Spec.Authentication = nil
	inherited := &securityPolicyOwners{basicAuth: parent}
	require.Equal(t, irConfigName(parent), authenticationBasicName(allPolicy, inherited))
	require.Equal(t, irConfigName(parent)+authenticationAnySuffix, authenticationBasicName(anyPolicy, inherited))
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
