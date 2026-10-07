// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"

	oauth2v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/oauth2/v3"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/testing/protocmp"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
)

// expectedCookieConfigs returns the cookie configs EG emits for the given SameSite
// policy: the nonce and code verifier cookies are always scoped to the redirect path,
// every cookie carries the SameSite policy.
func expectedCookieConfigs(sameSite oauth2v3.CookieConfig_SameSite, redirectPath string) *oauth2v3.CookieConfigs {
	return &oauth2v3.CookieConfigs{
		BearerTokenCookieConfig:  &oauth2v3.CookieConfig{SameSite: sameSite},
		OauthHmacCookieConfig:    &oauth2v3.CookieConfig{SameSite: sameSite},
		OauthExpiresCookieConfig: &oauth2v3.CookieConfig{SameSite: sameSite},
		IdTokenCookieConfig:      &oauth2v3.CookieConfig{SameSite: sameSite},
		RefreshTokenCookieConfig: &oauth2v3.CookieConfig{SameSite: sameSite},
		OauthNonceCookieConfig:   &oauth2v3.CookieConfig{SameSite: sameSite, Path: redirectPath},
		CodeVerifierCookieConfig: &oauth2v3.CookieConfig{SameSite: sameSite, Path: redirectPath},
	}
}

// expectedRedirectPathOnlyCookieConfigs returns the cookie configs EG emits when the
// user did not configure SameSite: only the two flow cookies are configured, so the
// session cookies keep Envoy's defaults.
func expectedRedirectPathOnlyCookieConfigs(redirectPath string) *oauth2v3.CookieConfigs {
	return &oauth2v3.CookieConfigs{
		OauthNonceCookieConfig:   &oauth2v3.CookieConfig{Path: redirectPath},
		CodeVerifierCookieConfig: &oauth2v3.CookieConfig{Path: redirectPath},
	}
}

func TestOIDCCookieConfigSameSite(t *testing.T) {
	tests := []struct {
		name   string
		input  ir.OIDC
		expect *oauth2v3.CookieConfigs
	}{
		{
			name:   "SameSite unset still scopes the flow cookies to the redirect path",
			input:  ir.OIDC{RedirectPath: "/oauth2/callback"},
			expect: expectedRedirectPathOnlyCookieConfigs("/oauth2/callback"),
		},
		{
			name: "all cookie configs set to None",
			input: ir.OIDC{
				RedirectPath: "/oauth2/callback",
				CookieConfig: &egv1a1.OIDCCookieConfig{
					SameSite: new("None"),
				},
			},
			expect: expectedCookieConfigs(oauth2v3.CookieConfig_NONE, "/oauth2/callback"),
		},
		{
			name: "all cookie configs set to Lax",
			input: ir.OIDC{
				RedirectPath: "/oauth2/callback",
				CookieConfig: &egv1a1.OIDCCookieConfig{
					SameSite: new("Lax"),
				},
			},
			expect: expectedCookieConfigs(oauth2v3.CookieConfig_LAX, "/oauth2/callback"),
		},
		{
			name: "all cookie configs set to Strict",
			input: ir.OIDC{
				RedirectPath: "/oauth2/callback",
				CookieConfig: &egv1a1.OIDCCookieConfig{
					SameSite: new("Strict"),
				},
			},
			expect: expectedCookieConfigs(oauth2v3.CookieConfig_STRICT, "/oauth2/callback"),
		},
		{
			name: "all cookie configs set to Disabled",
			input: ir.OIDC{
				RedirectPath: "/oauth2/callback",
				CookieConfig: &egv1a1.OIDCCookieConfig{
					SameSite: new("Disabled"),
				},
			},
			expect: expectedCookieConfigs(oauth2v3.CookieConfig_DISABLED, "/oauth2/callback"),
		},
		{
			name: "cookie config received invalid SameSite value will default to Disabled",
			input: ir.OIDC{
				RedirectPath: "/oauth2/callback",
				CookieConfig: &egv1a1.OIDCCookieConfig{
					SameSite: new("InvalidValue"),
				},
			},
			expect: expectedCookieConfigs(oauth2v3.CookieConfig_DISABLED, "/oauth2/callback"),
		},
		{
			name: "a custom redirect path scopes the flow cookies to that path",
			input: ir.OIDC{
				RedirectPath: "/auth/callback",
				CookieConfig: &egv1a1.OIDCCookieConfig{
					SameSite: new("Lax"),
				},
			},
			expect: expectedCookieConfigs(oauth2v3.CookieConfig_LAX, "/auth/callback"),
		},
		{
			// Envoy defaults an empty cookie path to "/".
			name:   "an empty redirect path leaves the cookie path unset",
			input:  ir.OIDC{},
			expect: expectedRedirectPathOnlyCookieConfigs(""),
		},
		{
			// Envoy rejects ";" in a cookie path, and an invalid path would fail
			// xDS validation and drop the route, so the path is left unset.
			name:   "a redirect path Envoy rejects on a cookie leaves the cookie path unset",
			input:  ir.OIDC{RedirectPath: "/oauth2;v2/callback"},
			expect: expectedRedirectPathOnlyCookieConfigs(""),
		},
	}

	for i := range tests {
		tc := &tests[i]
		t.Run(tc.name, func(t *testing.T) {
			actual := buildCookieConfigs(&tc.input)
			require.Empty(t, cmp.Diff(tc.expect, actual, protocmp.Transform()))
		})
	}
}
