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
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
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

func Test_OIDC_PassThroughAuthHeader(t *testing.T) {
	tests := []struct {
		name      string
		OIDC      egv1a1.OIDC
		JWT       *egv1a1.JWT
		wantError bool
	}{
		{
			name: "oidc and jwt with PassThroughAuthHeader configured",
			OIDC: egv1a1.OIDC{
				PassThroughAuthHeader: ToPointer(true),
			},
			JWT: &egv1a1.JWT{
				Providers: []egv1a1.JWTProvider{
					{
						Name: "test",
					},
				},
			},
			wantError: false,
		},
		{
			name: "jwt configured to read a non-standard header is ok",
			OIDC: egv1a1.OIDC{
				PassThroughAuthHeader: ToPointer(true),
			},
			JWT: &egv1a1.JWT{
				Providers: []egv1a1.JWTProvider{
					{
						Name: "test",
						ExtractFrom: &egv1a1.JWTExtractor{
							Headers: []egv1a1.JWTHeaderExtractor{{Name: "SomeHeader", ValuePrefix: ToPointer("Bearer ")}},
						},
					},
				},
			},
			wantError: false,
		},
		{
			name: "jwt configured to read a non-standard header without valuePrefix is ok",
			OIDC: egv1a1.OIDC{
				PassThroughAuthHeader: ToPointer(true),
			},
			JWT: &egv1a1.JWT{
				Providers: []egv1a1.JWTProvider{
					{
						Name: "test",
						ExtractFrom: &egv1a1.JWTExtractor{
							Headers: []egv1a1.JWTHeaderExtractor{{Name: "SomeHeader", ValuePrefix: nil}},
						},
					},
				},
			},
			wantError: false,
		},
		{
			name: "oidc with PassThroughAuthHeader configured requires jwt configured too",
			OIDC: egv1a1.OIDC{
				PassThroughAuthHeader: ToPointer(true),
			},
			JWT:       nil,
			wantError: true,
		},
		{
			name: "jwt configured to read cookie only is not ok",
			OIDC: egv1a1.OIDC{
				PassThroughAuthHeader: ToPointer(true),
			},
			JWT: &egv1a1.JWT{
				Providers: []egv1a1.JWTProvider{
					{
						Name: "test",
						ExtractFrom: &egv1a1.JWTExtractor{
							Cookies: []string{"SomeCookie"},
						},
					},
				},
			},
			wantError: true,
		},
		{
			name: "jwt configured with multiple providers is ok",
			OIDC: egv1a1.OIDC{
				PassThroughAuthHeader: ToPointer(true),
			},
			JWT: &(egv1a1.JWT{
				Providers: []egv1a1.JWTProvider{
					{
						Name: "test",
						ExtractFrom: &egv1a1.JWTExtractor{
							Headers: []egv1a1.JWTHeaderExtractor{{Name: "Blah"}},
						},
					},
					{
						Name: "test2",
						ExtractFrom: &egv1a1.JWTExtractor{
							Cookies: []string{"SomeCookie"},
						},
					},
				},
			}),
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			securityPolicy := egv1a1.SecurityPolicy{
				Spec: egv1a1.SecurityPolicySpec{
					OIDC: &tt.OIDC,
					JWT:  tt.JWT,
				},
			}
			err := validateSecurityPolicy(&securityPolicy)
			if (err != nil) != tt.wantError {
				t.Errorf("validateSecurityPolicy() error = %v, wantErr %v", err, tt.wantError)
				return
			}
		})
	}
}

func ToPointer[T any](v T) *T {
	return &v
}

func Test_validateHtpasswdFormat(t *testing.T) {
	tests := []struct {
		name      string
		htpasswd  string
		wantError bool
	}{
		{
			name:      "valid htpasswd with SHA format",
			htpasswd:  "user1:{SHA}hashed_user1_password\nuser2:{SHA}hashed_user2_password",
			wantError: false,
		},
		{
			name:      "valid htpasswd with SHA format and empty lines",
			htpasswd:  "user1:{SHA}hashed_user1_password\n\nuser2:{SHA}hashed_user2_password\n",
			wantError: false,
		},
		{
			name:      "invalid htpasswd with missing SHA prefix",
			htpasswd:  "user1:hashed_user1_password",
			wantError: true,
		},
		{
			name:      "invalid htpasswd with MD5 format",
			htpasswd:  "user1:$apr1$hashed_user1_password",
			wantError: true,
		},
		{
			name:      "invalid htpasswd with bcrypt format",
			htpasswd:  "user1:$2y$hashed_user1_password",
			wantError: true,
		},
		{
			name:      "invalid htpasswd with missing colon",
			htpasswd:  "user1{SHA}hashed_user1_password",
			wantError: true,
		},
		{
			name:      "mixed valid and invalid formats",
			htpasswd:  "user1:{SHA}hashed_user1_password\nuser2:$apr1$hashed_user2_password",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateHtpasswdFormat([]byte(tt.htpasswd))
			if (err != nil) != tt.wantError {
				t.Errorf("validateHtpasswdFormat() error = %v, wantErr %v", err, tt.wantError)
				return
			}
		})
	}
}

func Test_parseExtAuthTimeout(t *testing.T) {
	tests := []struct {
		name      string
		timeout   *gwapiv1.Duration
		wantValid bool
		wantValue string
	}{
		{
			name:      "valid timeout",
			timeout:   ptr.To(gwapiv1.Duration("10s")),
			wantValid: true,
			wantValue: "10s",
		},
		{
			name:      "invalid timeout format",
			timeout:   ptr.To(gwapiv1.Duration("invalid-duration")),
			wantValid: false,
			wantValue: "",
		},
		{
			name:      "nil timeout",
			timeout:   nil,
			wantValid: false,
			wantValue: "",
		},
		{
			name:      "complex valid timeout",
			timeout:   ptr.To(gwapiv1.Duration("1h30m45s")),
			wantValid: true,
			wantValue: "1h30m45s",
		},
		{
			name:      "millisecond timeout",
			timeout:   ptr.To(gwapiv1.Duration("500ms")),
			wantValid: true,
			wantValue: "500ms",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseExtAuthTimeout(tt.timeout)

			// Verify the timeout parsing behavior
			if tt.wantValid {
				assert.NotNil(t, result)
				assert.Equal(t, tt.wantValue, result.Duration.String())
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

func Test_getRouteProtocol(t *testing.T) {
	// nil route should default to HTTP
	require.Equal(t, ir.HTTP, getRouteProtocol(nil))
}

func Test_validateSecurityPolicyForTCP_AllowsNilOrEmptyAuth(t *testing.T) {
	// No Authorization present
	p := &egv1a1.SecurityPolicy{
		Spec: egv1a1.SecurityPolicySpec{},
	}
	require.NoError(t, validateSecurityPolicyForTCP(p))

	// Authorization present but no rules
	p = &egv1a1.SecurityPolicy{
		Spec: egv1a1.SecurityPolicySpec{
			Authorization: &egv1a1.Authorization{
				Rules: []egv1a1.AuthorizationRule{},
			},
		},
	}
	require.NoError(t, validateSecurityPolicyForTCP(p))
}

func Test_validateSecurityPolicyForTCP_UnsupportedTopLevelFeatures(t *testing.T) {
	// Any non-authorization feature should fail for TCP
	cases := []struct {
		name string
		spec egv1a1.SecurityPolicySpec
	}{
		{"JWT", egv1a1.SecurityPolicySpec{JWT: &egv1a1.JWT{}}},
		{"OIDC", egv1a1.SecurityPolicySpec{OIDC: &egv1a1.OIDC{}}},
		{"CORS", egv1a1.SecurityPolicySpec{CORS: &egv1a1.CORS{}}},
		{"APIKeyAuth", egv1a1.SecurityPolicySpec{APIKeyAuth: &egv1a1.APIKeyAuth{}}},
		{"BasicAuth", egv1a1.SecurityPolicySpec{BasicAuth: &egv1a1.BasicAuth{}}},
		{"ExtAuth", egv1a1.SecurityPolicySpec{ExtAuth: &egv1a1.ExtAuth{HTTP: &egv1a1.HTTPExtAuthService{}}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := &egv1a1.SecurityPolicy{Spec: tc.spec}
			require.Error(t, validateSecurityPolicyForTCP(p))
		})
	}
}

func Test_validateSecurityPolicyForTCP_AllowRequiresCIDR(t *testing.T) {
	// Allow rule without ClientCIDRs should error
	p := &egv1a1.SecurityPolicy{
		Spec: egv1a1.SecurityPolicySpec{
			Authorization: &egv1a1.Authorization{
				Rules: []egv1a1.AuthorizationRule{
					{
						Action: egv1a1.AuthorizationActionAllow,
						// Principal with zero ClientCIDRs (default) → invalid for TCP Allow
					},
				},
			},
		},
	}
	require.Error(t, validateSecurityPolicyForTCP(p))
}

func Test_validateSecurityPolicyForTCP_DenyRuleOK(t *testing.T) {
	// Deny rule is fine without CIDRs
	p := &egv1a1.SecurityPolicy{
		Spec: egv1a1.SecurityPolicySpec{
			Authorization: &egv1a1.Authorization{
				Rules: []egv1a1.AuthorizationRule{
					{
						Action: egv1a1.AuthorizationActionDeny,
					},
				},
			},
		},
	}
	require.NoError(t, validateSecurityPolicyForTCP(p))
}

func Test_validateSecurityPolicyForTCP_AllowWithCIDR_OK(t *testing.T) {
	// Allow rule with at least one ClientCIDR should be valid
	p := &egv1a1.SecurityPolicy{
		Spec: egv1a1.SecurityPolicySpec{
			Authorization: &egv1a1.Authorization{
				Rules: []egv1a1.AuthorizationRule{
					{
						Action: egv1a1.AuthorizationActionAllow,
						Principal: egv1a1.Principal{
							ClientCIDRs: []egv1a1.CIDR{"10.0.0.0/8"},
						},
					},
				},
			},
		},
	}
	require.NoError(t, validateSecurityPolicyForTCP(p))
}

func Test_validateSecurityPolicyForTCP_AllowWithInvalidCIDR_Errors(t *testing.T) {
	p := &egv1a1.SecurityPolicy{
		Spec: egv1a1.SecurityPolicySpec{
			Authorization: &egv1a1.Authorization{
				Rules: []egv1a1.AuthorizationRule{
					{
						Action: egv1a1.AuthorizationActionAllow,
						Principal: egv1a1.Principal{
							ClientCIDRs: []egv1a1.CIDR{"10.0.0.0/99"}, // invalid
						},
					},
				},
			},
		},
	}
	require.Error(t, validateSecurityPolicyForTCP(p))
}

func Test_validateSecurityPolicyForTCP_PrincipalJWT_NotSupported(t *testing.T) {
	p := &egv1a1.SecurityPolicy{
		Spec: egv1a1.SecurityPolicySpec{
			Authorization: &egv1a1.Authorization{
				Rules: []egv1a1.AuthorizationRule{{
					Action: egv1a1.AuthorizationActionAllow,
					Principal: egv1a1.Principal{
						ClientCIDRs: []egv1a1.CIDR{"10.0.0.0/8"},
						JWT:         &egv1a1.JWTPrincipal{}, // zero value is enough to be non-nil
					},
				}},
			},
		},
	}
	require.Error(t, validateSecurityPolicyForTCP(p))
}

func Test_validateSecurityPolicyForTCP_PrincipalHeaders_NotSupported(t *testing.T) {
	p := &egv1a1.SecurityPolicy{
		Spec: egv1a1.SecurityPolicySpec{
			Authorization: &egv1a1.Authorization{
				Rules: []egv1a1.AuthorizationRule{{
					Action: egv1a1.AuthorizationActionAllow,
					Principal: egv1a1.Principal{
						ClientCIDRs: []egv1a1.CIDR{"10.0.0.0/8"},
						// any non-empty slice will trigger the "headers not supported" branch
						Headers: []egv1a1.AuthorizationHeaderMatch{{}},
					},
				}},
			},
		},
	}
	require.Error(t, validateSecurityPolicyForTCP(p))
}

func Test_validateSecurityPolicyForTCP_DenyWithInvalidCIDR_Errors(t *testing.T) {
	p := &egv1a1.SecurityPolicy{
		Spec: egv1a1.SecurityPolicySpec{
			Authorization: &egv1a1.Authorization{
				Rules: []egv1a1.AuthorizationRule{{
					Action: egv1a1.AuthorizationActionDeny,
					Principal: egv1a1.Principal{
						ClientCIDRs: []egv1a1.CIDR{"10.0.0.0/99"}, // invalid mask
					},
				}},
			},
		},
	}
	require.Error(t, validateSecurityPolicyForTCP(p))
}

func Test_validateSecurityPolicyForTCP_MixedRules_OK(t *testing.T) {
	p := &egv1a1.SecurityPolicy{
		Spec: egv1a1.SecurityPolicySpec{
			Authorization: &egv1a1.Authorization{
				Rules: []egv1a1.AuthorizationRule{
					{
						Action: egv1a1.AuthorizationActionAllow,
						Principal: egv1a1.Principal{
							ClientCIDRs: []egv1a1.CIDR{"192.168.0.0/16"},
						},
					},
					{
						Action:    egv1a1.AuthorizationActionDeny,
						Principal: egv1a1.Principal{}, // no CIDR is fine
					},
				},
			},
		},
	}
	require.NoError(t, validateSecurityPolicyForTCP(p))
}
