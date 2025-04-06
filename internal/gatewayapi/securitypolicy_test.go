// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func Test_wildcard2regex(t *testing.T) {
	tests := []struct {
		name     string
		wildcard string
		origin   string
		want     int
	}{
		{
			name:     "test1",
			wildcard: "http://*.example.com",
			origin:   "http://foo.example.com",
			want:     1,
		},
		{
			name:     "test2",
			wildcard: "http://*.example.com",
			origin:   "http://foo.bar.example.com",
			want:     1,
		},
		{
			name:     "test3",
			wildcard: "http://*.example.com",
			origin:   "http://foo.bar.com",
			want:     0,
		},
		{
			name:     "test4",
			wildcard: "http://*.example.com",
			origin:   "https://foo.example.com",
			want:     0,
		},
		{
			name:     "test5",
			wildcard: "http://*.example.com:8080",
			origin:   "http://foo.example.com:8080",
			want:     1,
		},
		{
			name:     "test6",
			wildcard: "http://*.example.com:8080",
			origin:   "http://foo.bar.example.com:8080",
			want:     1,
		},
		{
			name:     "test7",
			wildcard: "http://*.example.com:8080",
			origin:   "http://foo.example.com",
			want:     0,
		},
		{
			name:     "test8",
			wildcard: "http://*",
			origin:   "http://foo.example.com",
			want:     1,
		},
		{
			name:     "test9",
			wildcard: "http://*",
			origin:   "https://foo.example.com",
			want:     0,
		},
		{
			name:     "test10",
			wildcard: "*",
			origin:   "http://foo.example.com",
			want:     1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			regexStr := wildcard2regex(tt.wildcard)
			regex, err := regexp.Compile(regexStr)
			require.NoError(t, err)
			finds := regex.FindAllString(tt.origin, -1)
			assert.Lenf(t, finds, tt.want, "wildcard2regex(%v)", tt.wildcard)
		})
	}
}

func Test_extractRedirectPath(t *testing.T) {
	tests := []struct {
		name        string
		redirectURL string
		want        string
		wantErr     bool
	}{
		{
			name:        "header value syntax",
			redirectURL: "%REQ(x-forwarded-proto)%://%REQ(:authority)%/petstore/oauth2/callback",
			want:        "/petstore/oauth2/callback",
			wantErr:     false,
		},
		{
			name:        "without header value syntax",
			redirectURL: "https://www.example.com/petstore/oauth2/callback",
			want:        "/petstore/oauth2/callback",
			wantErr:     false,
		},
		{
			name:        "with port",
			redirectURL: "https://www.example.com:9080/petstore/oauth2/callback",
			want:        "/petstore/oauth2/callback",
			wantErr:     false,
		},
		{
			name:        "without path",
			redirectURL: "https://www.example.com/",
			want:        "",
			wantErr:     true,
		},
		{
			name:        "without path",
			redirectURL: "https://www.example.com",
			want:        "",
			wantErr:     true,
		},
		{
			name:        "without scheme",
			redirectURL: "://www.example.com",
			want:        "",
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractRedirectPath(tt.redirectURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractRedirectPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				assert.Equalf(t, tt.want, got, "extractRedirectPath(%v)", tt.redirectURL)
			}
		})
	}
}

func Test_JWTProvider(t *testing.T) {
	tests := []struct {
		name      string
		Providers []egv1a1.JWTProvider
		wantError bool
	}{
		{
			name: "valid security policy with URI issuer",
			Providers: []egv1a1.JWTProvider{
				{
					Name:      "test",
					Issuer:    "https://www.test.local",
					Audiences: []string{"test.local"},
					RemoteJWKS: &egv1a1.RemoteJWKS{
						URI: "https://test.local/jwt/public-key/jwks.json",
					},
				},
			},
		},
		{
			name: "valid security policy with Email issuer",
			Providers: []egv1a1.JWTProvider{
				{
					Name:      "test",
					Issuer:    "test@test.local",
					Audiences: []string{"test.local"},
					RemoteJWKS: &egv1a1.RemoteJWKS{
						URI: "https://test.local/jwt/public-key/jwks.json",
					},
				},
			},
		},
		{
			name: "valid security policy with non URI/Email Issuer",
			Providers: []egv1a1.JWTProvider{
				{
					Name:      "test",
					Issuer:    "foo.bar.local",
					Audiences: []string{"foo.bar.local"},
					RemoteJWKS: &egv1a1.RemoteJWKS{
						URI: "https://test.local/jwt/public-key/jwks.json",
					},
				},
			},
		},
		{
			name: "valid security policy with jwtClaimToHeader",
			Providers: []egv1a1.JWTProvider{
				{
					Name:      "test",
					Issuer:    "test@test.local",
					Audiences: []string{"test.local"},
					RemoteJWKS: &egv1a1.RemoteJWKS{
						URI: "https://test.local/jwt/public-key/jwks.json",
					},
					ClaimToHeaders: []egv1a1.ClaimToHeader{
						{
							Header: "test",
							Claim:  "test",
						},
					},
				},
			},
		},

		{
			name: "unqualified authentication provider name",
			Providers: []egv1a1.JWTProvider{
				{
					Name:      "unqualified_...",
					Issuer:    "https://www.test.local",
					Audiences: []string{"test.local"},
					RemoteJWKS: &egv1a1.RemoteJWKS{
						URI: "https://test.local/jwt/public-key/jwks.json",
					},
				},
			},
			wantError: true,
		},
		{
			name: "unspecified provider name",
			Providers: []egv1a1.JWTProvider{
				{
					Name:      "",
					Issuer:    "https://www.test.local",
					Audiences: []string{"test.local"},
					RemoteJWKS: &egv1a1.RemoteJWKS{
						URI: "https://test.local/jwt/public-key/jwks.json",
					},
				},
			},
			wantError: true,
		},

		{
			name: "non unique provider names",
			Providers: []egv1a1.JWTProvider{
				{
					Name:      "unique",
					Issuer:    "https://www.test.local",
					Audiences: []string{"test.local"},
					RemoteJWKS: &egv1a1.RemoteJWKS{
						URI: "https://test.local/jwt/public-key/jwks.json",
					},
				},
				{
					Name:      "non-unique",
					Issuer:    "https://www.test.local",
					Audiences: []string{"test.local"},
					RemoteJWKS: &egv1a1.RemoteJWKS{
						URI: "https://test.local/jwt/public-key/jwks.json",
					},
				},
				{
					Name:      "non-unique",
					Issuer:    "https://www.test.local",
					Audiences: []string{"test.local"},
					RemoteJWKS: &egv1a1.RemoteJWKS{
						URI: "https://test.local/jwt/public-key/jwks.json",
					},
				},
			},
			wantError: true,
		},

		{
			name: "invalid issuer uri",
			Providers: []egv1a1.JWTProvider{
				{
					Name:      "test",
					Issuer:    "http://invalid url.local",
					Audiences: []string{"test.local"},
					RemoteJWKS: &egv1a1.RemoteJWKS{
						URI: "http://www.test.local",
					},
				},
			},
			wantError: true,
		},
		{
			name: "inivalid issuer email",
			Providers: []egv1a1.JWTProvider{
				{
					Name:      "test",
					Issuer:    "test@!123...",
					Audiences: []string{"test.local"},
					RemoteJWKS: &egv1a1.RemoteJWKS{
						URI: "https://test.local/jwt/public-key/jwks.json",
					},
				},
			},
			wantError: true,
		},
		{
			name: "invalid remote jwks uri",
			Providers: []egv1a1.JWTProvider{
				{
					Name:      "test",
					Issuer:    "http://www.test.local",
					Audiences: []string{"test.local"},
					RemoteJWKS: &egv1a1.RemoteJWKS{
						URI: "invalid/local",
					},
				},
			},
			wantError: true,
		},
		{
			name: "unspecified remote jwks uri",
			Providers: []egv1a1.JWTProvider{
				{
					Name:      "test",
					Audiences: []string{"test.local"},
					RemoteJWKS: &egv1a1.RemoteJWKS{
						URI: "",
					},
				},
			},
			wantError: true,
		},
		{
			name: "unspecified jwtClaimToHeader headerName",
			Providers: []egv1a1.JWTProvider{
				{
					Name:      "test",
					Issuer:    "test@test.local",
					Audiences: []string{"test.local"},
					RemoteJWKS: &egv1a1.RemoteJWKS{
						URI: "https://test.local/jwt/public-key/jwks.json",
					},
					ClaimToHeaders: []egv1a1.ClaimToHeader{
						{
							Header: "",
							Claim:  "test",
						},
					},
				},
			},
			wantError: true,
		},
		{
			name: "unspecified jwtClaimToHeader claimName",
			Providers: []egv1a1.JWTProvider{
				{
					Name:      "test",
					Issuer:    "test@test.local",
					Audiences: []string{"test.local"},
					RemoteJWKS: &egv1a1.RemoteJWKS{
						URI: "https://test.local/jwt/public-key/jwks.json",
					},
					ClaimToHeaders: []egv1a1.ClaimToHeader{
						{
							Header: "test",
							Claim:  "",
						},
					},
				},
			},
			wantError: true,
		},
		{
			name: "unspecified issuer",
			Providers: []egv1a1.JWTProvider{
				{
					Name:      "test",
					Audiences: []string{"test.local"},
					RemoteJWKS: &egv1a1.RemoteJWKS{
						URI: "https://test.local/jwt/public-key/jwks.json",
					},
				},
			},
			wantError: false,
		},
		{
			name: "unspecified audiences",
			Providers: []egv1a1.JWTProvider{
				{
					Name:   "test",
					Issuer: "https://www.test.local",
					RemoteJWKS: &egv1a1.RemoteJWKS{
						URI: "https://test.local/jwt/public-key/jwks.json",
					},
				},
			},
			wantError: false,
		},
		{
			name: "with both remoteJWKS and localJWKS",
			Providers: []egv1a1.JWTProvider{
				{
					Name:      "test",
					Issuer:    "https://www.test.local",
					Audiences: []string{"test.local"},
					RemoteJWKS: &egv1a1.RemoteJWKS{
						URI: "https://test.local/jwt/public-key/jwks.json",
					},
					LocalJWKS: &egv1a1.LocalJWKS{
						Inline: ptr.To("{}"),
					},
				},
			},
			wantError: true,
		},
		{
			name: "without remoteJWKS or localJWKS",
			Providers: []egv1a1.JWTProvider{
				{
					Name:      "test",
					Issuer:    "https://www.test.local",
					Audiences: []string{"test.local"},
				},
			},
			wantError: true,
		},
		{
			name: "localJWKS type without correct value",
			Providers: []egv1a1.JWTProvider{
				{
					Name:      "test",
					Issuer:    "https://www.test.local",
					Audiences: []string{"test.local"},
					LocalJWKS: &egv1a1.LocalJWKS{
						Type:   ptr.To(egv1a1.LocalJWKSTypeValueRef),
						Inline: ptr.To("{}"),
					},
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateJWTProvider(tt.Providers)
			if (err != nil) != tt.wantError {
				t.Errorf("validateJWTProvider() error = %v, wantErr %v", err, tt.wantError)
				return
			}
		})
	}
}

func Test_APIKeyAuth(t *testing.T) {
	tests := []struct {
		name       string
		APIKeyAuth egv1a1.APIKeyAuth
		wantError  bool
	}{
		{
			name: "only one of header, query or cookie is supposed to be specified",
			APIKeyAuth: egv1a1.APIKeyAuth{
				ExtractFrom: []*egv1a1.ExtractFrom{
					{
						Headers: []string{"header"},
						Params:  []string{"param"},
					},
				},
			},
			wantError: true,
		},
		{
			name: "only one of header, query or cookie is supposed to be specified",
			APIKeyAuth: egv1a1.APIKeyAuth{
				ExtractFrom: []*egv1a1.ExtractFrom{
					{
						Headers: []string{"header"},
						Cookies: []string{"cookie"},
					},
				},
			},
			wantError: true,
		},
		{
			name: "only one of header, query or cookie is supposed to be specified",
			APIKeyAuth: egv1a1.APIKeyAuth{
				ExtractFrom: []*egv1a1.ExtractFrom{
					{
						Params:  []string{"param"},
						Cookies: []string{"cookie"},
					},
				},
			},
			wantError: true,
		},
		{
			name: "only one of header, query or cookie is supposed to be specified",
			APIKeyAuth: egv1a1.APIKeyAuth{
				ExtractFrom: []*egv1a1.ExtractFrom{
					{
						Headers: []string{"header"},
						Params:  []string{"param"},
						Cookies: []string{"cookie"},
					},
				},
			},
			wantError: true,
		},
		{
			name: "valid APIKeyAuth",
			APIKeyAuth: egv1a1.APIKeyAuth{
				ExtractFrom: []*egv1a1.ExtractFrom{
					{
						Headers: []string{"header"},
					},
				},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAPIKeyAuth(&tt.APIKeyAuth)
			if (err != nil) != tt.wantError {
				t.Errorf("validateAPIKeyAuth() error = %v, wantErr %v", err, tt.wantError)
				return
			}
		})
	}
}
