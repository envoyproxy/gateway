// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
)

func makePolicy(mode *egv1a1.AuthenticationMode) *egv1a1.SecurityPolicy {
	return &egv1a1.SecurityPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec:       egv1a1.SecurityPolicySpec{AuthMode: mode},
	}
}

func makeJWT() *ir.JWT {
	return &ir.JWT{AllowMissing: false, Providers: []ir.JWTProvider{{Name: "p1"}}}
}

func makeBasicAuth() *ir.BasicAuth {
	return &ir.BasicAuth{Name: "ba", AllowMissing: false}
}

func makeOIDC() *ir.OIDC { return &ir.OIDC{} }

func makeAPIKeyAuth() *ir.APIKeyAuth { return &ir.APIKeyAuth{} }

func TestAuthMethodCount(t *testing.T) {
	tests := []struct {
		name      string
		jwt       *ir.JWT
		basicAuth *ir.BasicAuth
		oidc      *ir.OIDC
		apiKey    *ir.APIKeyAuth
		want      int
	}{
		{name: "none configured", want: 0},
		{name: "only JWT", jwt: makeJWT(), want: 1},
		{name: "JWT and BasicAuth", jwt: makeJWT(), basicAuth: makeBasicAuth(), want: 2},
		{name: "JWT, BasicAuth and OIDC", jwt: makeJWT(), basicAuth: makeBasicAuth(), oidc: makeOIDC(), want: 3},
		{name: "all four", jwt: makeJWT(), basicAuth: makeBasicAuth(), oidc: makeOIDC(), apiKey: makeAPIKeyAuth(), want: 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := authMethodCount(tt.jwt, tt.basicAuth, tt.oidc, tt.apiKey)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestApplyAuthMode_AllMode_DoesNotSetAllowMissing(t *testing.T) {
	policy := makePolicy(ptr.To(egv1a1.AuthenticationModeAll))
	features := &ir.SecurityFeatures{}
	jwt := makeJWT()
	ba := makeBasicAuth()
	applyAuthMode(policy, features, jwt, ba, nil, nil)
	assert.False(t, jwt.AllowMissing)
	assert.False(t, ba.AllowMissing)
	assert.Nil(t, features.AuthMode)
}

func TestApplyAuthMode_NilMode_DefaultsToAll(t *testing.T) {
	policy := makePolicy(nil)
	features := &ir.SecurityFeatures{}
	jwt := makeJWT()
	ba := makeBasicAuth()
	applyAuthMode(policy, features, jwt, ba, nil, nil)
	assert.False(t, jwt.AllowMissing)
	assert.False(t, ba.AllowMissing)
	assert.Nil(t, features.AuthMode)
}

func TestApplyAuthMode_AnyMode_SingleMethod_NoOp(t *testing.T) {
	policy := makePolicy(ptr.To(egv1a1.AuthenticationModeAny))
	features := &ir.SecurityFeatures{}
	jwt := makeJWT()
	applyAuthMode(policy, features, jwt, nil, nil, nil)
	assert.False(t, jwt.AllowMissing)
	assert.Nil(t, features.AuthMode)
}

func TestApplyAuthMode_AnyMode_JWTAndBasicAuth_SetsAllowMissing(t *testing.T) {
	policy := makePolicy(ptr.To(egv1a1.AuthenticationModeAny))
	features := &ir.SecurityFeatures{}
	jwt := makeJWT()
	ba := makeBasicAuth()
	applyAuthMode(policy, features, jwt, ba, nil, nil)
	assert.True(t, jwt.AllowMissing)
	assert.True(t, ba.AllowMissing)
	require.NotNil(t, features.AuthMode)
	assert.Equal(t, string(egv1a1.AuthenticationModeAny), *features.AuthMode)
}

func TestApplyAuthMode_AnyMode_PreservesJWTProviders(t *testing.T) {
	policy := makePolicy(ptr.To(egv1a1.AuthenticationModeAny))
	features := &ir.SecurityFeatures{}
	jwt := makeJWT()
	ba := makeBasicAuth()
	originalProviders := jwt.Providers
	applyAuthMode(policy, features, jwt, ba, nil, nil)
	assert.Equal(t, originalProviders, jwt.Providers)
}

func TestApplyAuthMode_AnyMode_ThreeMethods(t *testing.T) {
	policy := makePolicy(ptr.To(egv1a1.AuthenticationModeAny))
	features := &ir.SecurityFeatures{}
	jwt := makeJWT()
	ba := makeBasicAuth()
	applyAuthMode(policy, features, jwt, ba, makeOIDC(), nil)
	assert.True(t, jwt.AllowMissing)
	assert.True(t, ba.AllowMissing)
	require.NotNil(t, features.AuthMode)
}

func TestApplyAuthMode_AnyMode_NilAuthObjects_NoPanic(t *testing.T) {
	policy := makePolicy(ptr.To(egv1a1.AuthenticationModeAny))
	features := &ir.SecurityFeatures{}
	assert.NotPanics(t, func() {
		applyAuthMode(policy, features, nil, nil, nil, nil)
	})
	assert.Nil(t, features.AuthMode)
}

func TestApplyAuthMode_AnyMode_Idempotent(t *testing.T) {
	policy := makePolicy(ptr.To(egv1a1.AuthenticationModeAny))
	features := &ir.SecurityFeatures{}
	jwt := makeJWT()
	ba := makeBasicAuth()
	applyAuthMode(policy, features, jwt, ba, nil, nil)
	applyAuthMode(policy, features, jwt, ba, nil, nil)
	assert.True(t, jwt.AllowMissing)
	assert.True(t, ba.AllowMissing)
	require.NotNil(t, features.AuthMode)
}

func TestApplyAuthMode_AnyMode_OnlyBasicAuth_NoMutation(t *testing.T) {
	policy := makePolicy(ptr.To(egv1a1.AuthenticationModeAny))
	features := &ir.SecurityFeatures{}
	ba := makeBasicAuth()
	applyAuthMode(policy, features, nil, ba, nil, nil)
	assert.False(t, ba.AllowMissing)
	assert.Nil(t, features.AuthMode)
}
