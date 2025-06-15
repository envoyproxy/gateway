// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"

	oauth2v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/oauth2/v3"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
)

func TestOIDCCookieConfigSameSite(t *testing.T) {
	tests := []struct {
		name   string
		input  ir.OIDC
		expect oauth2v3.CookieConfigs
	}{
		{
			name:  "defaults all cookie to strict",
			input: ir.OIDC{},
			expect: oauth2v3.CookieConfigs{
				BearerTokenCookieConfig:  &oauth2v3.CookieConfig{SameSite: oauth2v3.CookieConfig_STRICT},
				OauthHmacCookieConfig:    &oauth2v3.CookieConfig{SameSite: oauth2v3.CookieConfig_STRICT},
				OauthExpiresCookieConfig: &oauth2v3.CookieConfig{SameSite: oauth2v3.CookieConfig_STRICT},
				IdTokenCookieConfig:      &oauth2v3.CookieConfig{SameSite: oauth2v3.CookieConfig_STRICT},
				RefreshTokenCookieConfig: &oauth2v3.CookieConfig{SameSite: oauth2v3.CookieConfig_STRICT},
				OauthNonceCookieConfig:   &oauth2v3.CookieConfig{SameSite: oauth2v3.CookieConfig_STRICT},
				CodeVerifierCookieConfig: &oauth2v3.CookieConfig{SameSite: oauth2v3.CookieConfig_STRICT},
			},
		},
		{
			name: "all cookie configs set to None",
			input: ir.OIDC{
				CookieConfig: &egv1a1.OIDCCookieConfig{
					SameSite: ptr.To("None"),
				},
			},
			expect: oauth2v3.CookieConfigs{
				BearerTokenCookieConfig:  &oauth2v3.CookieConfig{SameSite: oauth2v3.CookieConfig_NONE},
				OauthHmacCookieConfig:    &oauth2v3.CookieConfig{SameSite: oauth2v3.CookieConfig_NONE},
				OauthExpiresCookieConfig: &oauth2v3.CookieConfig{SameSite: oauth2v3.CookieConfig_NONE},
				IdTokenCookieConfig:      &oauth2v3.CookieConfig{SameSite: oauth2v3.CookieConfig_NONE},
				RefreshTokenCookieConfig: &oauth2v3.CookieConfig{SameSite: oauth2v3.CookieConfig_NONE},
				OauthNonceCookieConfig:   &oauth2v3.CookieConfig{SameSite: oauth2v3.CookieConfig_NONE},
				CodeVerifierCookieConfig: &oauth2v3.CookieConfig{SameSite: oauth2v3.CookieConfig_NONE},
			},
		},
		{
			name: "all cookie configs set to Lax",
			input: ir.OIDC{
				CookieConfig: &egv1a1.OIDCCookieConfig{
					SameSite: ptr.To("Lax"),
				},
			},
			expect: oauth2v3.CookieConfigs{
				BearerTokenCookieConfig:  &oauth2v3.CookieConfig{SameSite: oauth2v3.CookieConfig_LAX},
				OauthHmacCookieConfig:    &oauth2v3.CookieConfig{SameSite: oauth2v3.CookieConfig_LAX},
				OauthExpiresCookieConfig: &oauth2v3.CookieConfig{SameSite: oauth2v3.CookieConfig_LAX},
				IdTokenCookieConfig:      &oauth2v3.CookieConfig{SameSite: oauth2v3.CookieConfig_LAX},
				RefreshTokenCookieConfig: &oauth2v3.CookieConfig{SameSite: oauth2v3.CookieConfig_LAX},
				OauthNonceCookieConfig:   &oauth2v3.CookieConfig{SameSite: oauth2v3.CookieConfig_LAX},
				CodeVerifierCookieConfig: &oauth2v3.CookieConfig{SameSite: oauth2v3.CookieConfig_LAX},
			},
		},
		{
			name: "all cookie configs set to Strict",
			input: ir.OIDC{
				CookieConfig: &egv1a1.OIDCCookieConfig{
					SameSite: ptr.To("Strict"),
				},
			},
			expect: oauth2v3.CookieConfigs{
				BearerTokenCookieConfig:  &oauth2v3.CookieConfig{SameSite: oauth2v3.CookieConfig_STRICT},
				OauthHmacCookieConfig:    &oauth2v3.CookieConfig{SameSite: oauth2v3.CookieConfig_STRICT},
				OauthExpiresCookieConfig: &oauth2v3.CookieConfig{SameSite: oauth2v3.CookieConfig_STRICT},
				IdTokenCookieConfig:      &oauth2v3.CookieConfig{SameSite: oauth2v3.CookieConfig_STRICT},
				RefreshTokenCookieConfig: &oauth2v3.CookieConfig{SameSite: oauth2v3.CookieConfig_STRICT},
				OauthNonceCookieConfig:   &oauth2v3.CookieConfig{SameSite: oauth2v3.CookieConfig_STRICT},
				CodeVerifierCookieConfig: &oauth2v3.CookieConfig{SameSite: oauth2v3.CookieConfig_STRICT},
			},
		},
		{
			name: "all cookie configs set to Disabled",
			input: ir.OIDC{
				CookieConfig: &egv1a1.OIDCCookieConfig{
					SameSite: ptr.To("Disabled"),
				},
			},
			expect: oauth2v3.CookieConfigs{
				BearerTokenCookieConfig:  &oauth2v3.CookieConfig{SameSite: oauth2v3.CookieConfig_DISABLED},
				OauthHmacCookieConfig:    &oauth2v3.CookieConfig{SameSite: oauth2v3.CookieConfig_DISABLED},
				OauthExpiresCookieConfig: &oauth2v3.CookieConfig{SameSite: oauth2v3.CookieConfig_DISABLED},
				IdTokenCookieConfig:      &oauth2v3.CookieConfig{SameSite: oauth2v3.CookieConfig_DISABLED},
				RefreshTokenCookieConfig: &oauth2v3.CookieConfig{SameSite: oauth2v3.CookieConfig_DISABLED},
				OauthNonceCookieConfig:   &oauth2v3.CookieConfig{SameSite: oauth2v3.CookieConfig_DISABLED},
				CodeVerifierCookieConfig: &oauth2v3.CookieConfig{SameSite: oauth2v3.CookieConfig_DISABLED},
			},
		},
		{
			name: "cookie config received invalid SameSite value will default to Strict",
			input: ir.OIDC{
				CookieConfig: &egv1a1.OIDCCookieConfig{
					SameSite: ptr.To("InvalidValue"),
				},
			},
			expect: oauth2v3.CookieConfigs{
				BearerTokenCookieConfig:  &oauth2v3.CookieConfig{SameSite: oauth2v3.CookieConfig_STRICT},
				OauthHmacCookieConfig:    &oauth2v3.CookieConfig{SameSite: oauth2v3.CookieConfig_STRICT},
				OauthExpiresCookieConfig: &oauth2v3.CookieConfig{SameSite: oauth2v3.CookieConfig_STRICT},
				IdTokenCookieConfig:      &oauth2v3.CookieConfig{SameSite: oauth2v3.CookieConfig_STRICT},
				RefreshTokenCookieConfig: &oauth2v3.CookieConfig{SameSite: oauth2v3.CookieConfig_STRICT},
				OauthNonceCookieConfig:   &oauth2v3.CookieConfig{SameSite: oauth2v3.CookieConfig_STRICT},
				CodeVerifierCookieConfig: &oauth2v3.CookieConfig{SameSite: oauth2v3.CookieConfig_STRICT},
			},
		},
	}

	for i := range tests {
		tc := &tests[i]
		t.Run(tc.name, func(t *testing.T) {
			actual := buildCookieConfigs(&tc.input)
			require.Equal(t, tc.expect.BearerTokenCookieConfig.SameSite, actual.BearerTokenCookieConfig.SameSite)
			require.Equal(t, tc.expect.OauthHmacCookieConfig.SameSite, actual.OauthHmacCookieConfig.SameSite)
			require.Equal(t, tc.expect.OauthExpiresCookieConfig.SameSite, actual.OauthExpiresCookieConfig.SameSite)
			require.Equal(t, tc.expect.IdTokenCookieConfig.SameSite, actual.IdTokenCookieConfig.SameSite)
			require.Equal(t, tc.expect.RefreshTokenCookieConfig.SameSite, actual.RefreshTokenCookieConfig.SameSite)
			require.Equal(t, tc.expect.OauthNonceCookieConfig.SameSite, actual.OauthNonceCookieConfig.SameSite)
			require.Equal(t, tc.expect.CodeVerifierCookieConfig.SameSite, actual.CodeVerifierCookieConfig.SameSite)
		})
	}
}
