package translator

import (
	"testing"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	oauth2v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/oauth2/v3"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
)

func TestOIDCCookieConfigSameSite(t *testing.T) {
	tests := []struct {
		name   string
		input  ir.OIDC
		expect oauth2v3.CookieConfigs
	}{
		{
			name:  "defaults all cookie to disabled",
			input: ir.OIDC{},
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
			name: "partial configuration on bearer token",
			input: ir.OIDC{
				CookieConfigs: &egv1a1.OIDCCookieConfigs{
					BearerToken: &egv1a1.OIDCCookieConfig{SameSite: ptr.To("Lax")},
				},
			},
			expect: oauth2v3.CookieConfigs{
				BearerTokenCookieConfig:  &oauth2v3.CookieConfig{SameSite: oauth2v3.CookieConfig_LAX},
				OauthHmacCookieConfig:    &oauth2v3.CookieConfig{SameSite: oauth2v3.CookieConfig_DISABLED},
				OauthExpiresCookieConfig: &oauth2v3.CookieConfig{SameSite: oauth2v3.CookieConfig_DISABLED},
				IdTokenCookieConfig:      &oauth2v3.CookieConfig{SameSite: oauth2v3.CookieConfig_DISABLED},
				RefreshTokenCookieConfig: &oauth2v3.CookieConfig{SameSite: oauth2v3.CookieConfig_DISABLED},
				OauthNonceCookieConfig:   &oauth2v3.CookieConfig{SameSite: oauth2v3.CookieConfig_DISABLED},
				CodeVerifierCookieConfig: &oauth2v3.CookieConfig{SameSite: oauth2v3.CookieConfig_DISABLED},
			},
		},
		{
			name: "all cookie configs set to None",
			input: ir.OIDC{
				CookieConfigs: &egv1a1.OIDCCookieConfigs{
					BearerToken:  &egv1a1.OIDCCookieConfig{SameSite: ptr.To("None")},
					OauthHmac:    &egv1a1.OIDCCookieConfig{SameSite: ptr.To("None")},
					OauthExpires: &egv1a1.OIDCCookieConfig{SameSite: ptr.To("None")},
					IdToken:      &egv1a1.OIDCCookieConfig{SameSite: ptr.To("None")},
					RefreshToken: &egv1a1.OIDCCookieConfig{SameSite: ptr.To("None")},
					OauthNonce:   &egv1a1.OIDCCookieConfig{SameSite: ptr.To("None")},
					CodeVerifier: &egv1a1.OIDCCookieConfig{SameSite: ptr.To("None")},
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
	}

	for _, tc := range tests {
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
