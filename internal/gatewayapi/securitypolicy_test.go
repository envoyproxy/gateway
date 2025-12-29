// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
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

func TestValidateCIDRs_ErrorOnBadCIDR(t *testing.T) {
	if err := validateCIDRs([]egv1a1.CIDR{"10.0.0.0/33"}); err == nil {
		t.Fatal("expected invalid ClientCIDR error")
	}
}

func TestTranslatorFetchEndpointsFromIssuerCache(t *testing.T) {
	var (
		callCount atomic.Int32
		server    *httptest.Server
	)

	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/.well-known/openid-configuration" {
			http.NotFound(w, r)
			return
		}

		callCount.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"token_endpoint":%q,"authorization_endpoint":%q}`, server.URL+"/token", server.URL+"/authorize")
	}))
	defer server.Close()

	tr := &Translator{GatewayControllerName: "gateway.envoyproxy.io/gatewayclass-controller"}
	tr.oidcDiscoveryCache = newOIDCDiscoveryCache()

	cfg, err := tr.fetchEndpointsFromIssuer(server.URL, nil)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.Equal(t, int32(1), callCount.Load())

	cfgCached, err := tr.fetchEndpointsFromIssuer(server.URL, nil)
	require.NoError(t, err)
	require.NotNil(t, cfgCached)
	require.Equal(t, int32(1), callCount.Load(), "second fetch should use cache")

	cfgAgain, err := tr.fetchEndpointsFromIssuer(server.URL, nil)
	require.NoError(t, err)
	require.NotNil(t, cfgAgain)
	require.Equal(t, int32(1), callCount.Load(), "subsequent fetch should continue using cache")
}

func TestTranslatorFetchEndpointsFromIssuerCacheError(t *testing.T) {
	var (
		callCount atomic.Int32
		server    *httptest.Server
	)

	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/.well-known/openid-configuration" {
			http.NotFound(w, r)
			return
		}

		callCount.Add(1)
		http.NotFound(w, r)
	}))
	defer server.Close()

	tr := &Translator{GatewayControllerName: "gateway.envoyproxy.io/gatewayclass-controller"}
	tr.oidcDiscoveryCache = newOIDCDiscoveryCache()

	cfg, err := tr.fetchEndpointsFromIssuer(server.URL, nil)
	require.Error(t, err)
	require.Nil(t, cfg)
	require.Equal(t, int32(1), callCount.Load())

	cfgCached, err := tr.fetchEndpointsFromIssuer(server.URL, nil)
	require.Error(t, err)
	require.Nil(t, cfgCached)
	require.Equal(t, int32(1), callCount.Load(), "second fetch should use cached error")

	cfgAfter, err := tr.fetchEndpointsFromIssuer(server.URL, nil)
	require.Error(t, err)
	require.Nil(t, cfgAfter)
	require.Equal(t, int32(1), callCount.Load(), "subsequent fetch should continue using cached error")
}

// / tiny helper to build a minimal SecurityPolicy
func sp(ns, name string) *egv1a1.SecurityPolicy {
	return &egv1a1.SecurityPolicy{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: name},
	}
}

// looks for any False condition on PolicyStatus parents (how EG surfaces translation errors)
func hasParentFalseCondition(p *egv1a1.SecurityPolicy) bool {
	for _, pr := range p.Status.Ancestors {
		for _, c := range pr.Conditions {
			if c.Status == metav1.ConditionFalse {
				return true
			}
		}
	}
	return false
}

// --- TCP branch: validateSecurityPolicyForTCP(...) returns err -> SetTranslationErrorForPolicyAncestors(...) + return
func Test_SecurityPolicy_TCP_Invalid_setsStatus_and_returns(t *testing.T) {
	tr := &Translator{GatewayControllerName: "gateway.envoyproxy.io/gatewayclass-controller"}
	trContext := &TranslatorContext{}

	// Create an invalid TCP policy (has CORS which is not allowed for TCP)
	policy := sp("default", "bad-tcp")
	policy.Spec.CORS = &egv1a1.CORS{}

	// Create a mock TCP route
	tcpRoute := &TCPRouteContext{
		TCPRoute: &gwapiv1a2.TCPRoute{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "tcp-route",
			},
			Spec: gwapiv1a2.TCPRouteSpec{
				CommonRouteSpec: gwapiv1a2.CommonRouteSpec{
					ParentRefs: []gwapiv1a2.ParentReference{
						{
							Name: "test-gateway",
						},
					},
				},
				Rules: []gwapiv1a2.TCPRouteRule{
					{
						BackendRefs: []gwapiv1a2.BackendRef{
							{
								BackendObjectReference: gwapiv1a2.BackendObjectReference{
									Name: "test-service",
									Port: ptr.To(gwapiv1a2.PortNumber(80)),
								},
							},
						},
					},
				},
			},
		},
	}

	// Create the target reference
	target := gwapiv1.LocalPolicyTargetReferenceWithSectionName{
		LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
			Group: gwapiv1.Group(gwapiv1.GroupVersion.Group),
			Kind:  resource.KindTCPRoute,
			Name:  "tcp-route",
		},
	}

	// Create route map
	routeMap := make(map[policyTargetRouteKey]*policyRouteTargetContext)
	key := policyTargetRouteKey{
		Kind:      string(resource.KindTCPRoute),
		Name:      "tcp-route",
		Namespace: "default",
	}
	routeMap[key] = &policyRouteTargetContext{RouteContext: tcpRoute}

	gatewayRouteMap := make(map[string]map[string]sets.Set[string])
	resources := resource.NewResources()
	xdsIR := make(resource.XdsIRMap)
	trContext.SetServices(resources.Services)
	tr.TranslatorContext = trContext

	// Process the policy - this should set error status
	tr.processSecurityPolicyForRoute(resources, xdsIR, routeMap, gatewayRouteMap, policy, target)

	// Assert that the policy has a False condition (error was set)
	require.True(t, hasParentFalseCondition(policy))
}

// --- non-TCP branch: malformed CIDR should return err -> SetTranslationErrorForPolicyAncestors(...) + return
func Test_SecurityPolicy_HTTP_Invalid_setsStatus_and_returns(t *testing.T) {
	tr := &Translator{GatewayControllerName: "gateway.envoyproxy.io/gatewayclass-controller"}
	trContext := &TranslatorContext{}

	// Create an invalid HTTP policy (malformed CIDR)
	policy := sp("default", "bad-http")
	policy.Spec.Authorization = &egv1a1.Authorization{
		Rules: []egv1a1.AuthorizationRule{
			{Principal: egv1a1.Principal{ClientCIDRs: []egv1a1.CIDR{"not-a-cidr"}}},
		},
	}

	// Create a mock HTTP route
	httpRoute := &HTTPRouteContext{
		HTTPRoute: &gwapiv1.HTTPRoute{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "http-route",
			},
			Spec: gwapiv1.HTTPRouteSpec{
				CommonRouteSpec: gwapiv1a2.CommonRouteSpec{
					ParentRefs: []gwapiv1.ParentReference{
						{
							Name: "test-gateway",
						},
					},
				},
				Rules: []gwapiv1.HTTPRouteRule{
					{
						BackendRefs: []gwapiv1.HTTPBackendRef{
							{
								BackendRef: gwapiv1.BackendRef{
									BackendObjectReference: gwapiv1.BackendObjectReference{
										Name: "test-service",
										Port: ptr.To(gwapiv1.PortNumber(80)),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Create the target reference
	target := gwapiv1.LocalPolicyTargetReferenceWithSectionName{
		LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
			Group: gwapiv1.Group(gwapiv1.GroupVersion.Group),
			Kind:  resource.KindHTTPRoute,
			Name:  "http-route",
		},
	}

	// Create route map
	routeMap := make(map[policyTargetRouteKey]*policyRouteTargetContext)
	key := policyTargetRouteKey{
		Kind:      string(resource.KindHTTPRoute),
		Name:      "http-route",
		Namespace: "default",
	}
	routeMap[key] = &policyRouteTargetContext{RouteContext: httpRoute}

	gatewayRouteMap := make(map[string]map[string]sets.Set[string])
	resources := resource.NewResources()
	xdsIR := make(resource.XdsIRMap)
	trContext.SetServices(resources.Services)
	tr.TranslatorContext = trContext

	// Process the policy - this should set error status
	tr.processSecurityPolicyForRoute(resources, xdsIR, routeMap, gatewayRouteMap, policy, target)

	// Assert that the policy has a False condition (error was set)
	require.True(t, hasParentFalseCondition(policy))
}

func Test_validateSecurityPolicyForTCP_Table(t *testing.T) {
	tests := []struct {
		name    string
		spec    egv1a1.SecurityPolicySpec
		wantErr bool
	}{
		{
			name:    "no authorization (nil)",
			spec:    egv1a1.SecurityPolicySpec{},
			wantErr: false,
		},
		{
			name: "authorization present but no rules",
			spec: egv1a1.SecurityPolicySpec{
				Authorization: &egv1a1.Authorization{
					Rules: []egv1a1.AuthorizationRule{},
				},
			},
			wantErr: false,
		},
		{
			name: "deny rule ok without cidrs",
			spec: egv1a1.SecurityPolicySpec{
				Authorization: &egv1a1.Authorization{
					Rules: []egv1a1.AuthorizationRule{
						{Action: egv1a1.AuthorizationActionDeny},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "allow with valid cidr ok",
			spec: egv1a1.SecurityPolicySpec{
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
			wantErr: false,
		},
		{
			name: "allow with invalid cidr errors",
			spec: egv1a1.SecurityPolicySpec{
				Authorization: &egv1a1.Authorization{
					Rules: []egv1a1.AuthorizationRule{
						{
							Action: egv1a1.AuthorizationActionAllow,
							Principal: egv1a1.Principal{
								ClientCIDRs: []egv1a1.CIDR{"10.0.0.0/99"},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "deny with invalid cidr errors",
			spec: egv1a1.SecurityPolicySpec{
				Authorization: &egv1a1.Authorization{
					Rules: []egv1a1.AuthorizationRule{
						{
							Action: egv1a1.AuthorizationActionDeny,
							Principal: egv1a1.Principal{
								ClientCIDRs: []egv1a1.CIDR{"10.0.0.0/99"},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "jwt principal not supported on tcp",
			spec: egv1a1.SecurityPolicySpec{
				Authorization: &egv1a1.Authorization{
					Rules: []egv1a1.AuthorizationRule{
						{
							Action: egv1a1.AuthorizationActionAllow,
							Principal: egv1a1.Principal{
								ClientCIDRs: []egv1a1.CIDR{"10.0.0.0/8"},
								JWT:         &egv1a1.JWTPrincipal{},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "headers principal not supported on tcp",
			spec: egv1a1.SecurityPolicySpec{
				Authorization: &egv1a1.Authorization{
					Rules: []egv1a1.AuthorizationRule{
						{
							Action: egv1a1.AuthorizationActionAllow,
							Principal: egv1a1.Principal{
								ClientCIDRs: []egv1a1.CIDR{"10.0.0.0/8"},
								Headers:     []egv1a1.AuthorizationHeaderMatch{{}},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "mixed allow and deny ok",
			spec: egv1a1.SecurityPolicySpec{
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
							Principal: egv1a1.Principal{},
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := &egv1a1.SecurityPolicy{Spec: tc.spec}
			err := validateSecurityPolicyForTCP(p)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_buildContextExtensions(t *testing.T) {
	policyNs := "default"
	tests := []struct {
		name              string
		contextExtensions []*egv1a1.ContextExtension
		translatorContext *TranslatorContext
		want              []*ir.ContextExtention
		wantErr           bool
	}{
		{
			name:              "Nil",
			contextExtensions: nil,
			want:              nil,
		},
		{
			name:              "Empty",
			contextExtensions: []*egv1a1.ContextExtension{},
			want:              nil,
		},
		{
			name: "TypeValue",
			contextExtensions: []*egv1a1.ContextExtension{
				{Name: "foo", Value: ptr.To("bar")},
			},
			want: []*ir.ContextExtention{{Name: "foo", Value: ir.PrivateBytes("bar")}},
		},
		{
			name:              "TypeValueEmpty",
			contextExtensions: []*egv1a1.ContextExtension{{Name: "foo"}},
			want:              []*ir.ContextExtention{{Name: "foo", Value: nil}},
		},
		{
			name: "TypeValueExplicit",
			contextExtensions: []*egv1a1.ContextExtension{
				{
					Name:  "foo",
					Type:  egv1a1.ContextExtensionValueTypeValue,
					Value: ptr.To("bar"),
				},
			},
			want: []*ir.ContextExtention{{Name: "foo", Value: ir.PrivateBytes("bar")}},
		},
		{
			name: "TypeValueRefNil",
			contextExtensions: []*egv1a1.ContextExtension{
				{Name: "foo", Type: egv1a1.ContextExtensionValueTypeValueRef},
			},
			wantErr: true,
		},
		{
			name: "TypeValueRefConfigMapNotFound",
			contextExtensions: []*egv1a1.ContextExtension{{
				Name: "foo",
				Type: egv1a1.ContextExtensionValueTypeValueRef,
				ValueRef: &egv1a1.LocalObjectKeyReference{
					LocalObjectReference: gwapiv1.LocalObjectReference{
						Kind: resource.KindConfigMap,
						Name: "test-cm",
					},
					Key: "test-key",
				},
			}},
			translatorContext: &TranslatorContext{},
			wantErr:           true,
		},
		{
			name: "TypeValueRefConfigMapKeyNotFound",
			contextExtensions: []*egv1a1.ContextExtension{{
				Name: "foo",
				Type: egv1a1.ContextExtensionValueTypeValueRef,
				ValueRef: &egv1a1.LocalObjectKeyReference{
					LocalObjectReference: gwapiv1.LocalObjectReference{
						Kind: resource.KindConfigMap,
						Name: "test-cm",
					},
					Key: "test-key",
				},
			}},
			translatorContext: &TranslatorContext{
				ConfigMapMap: map[types.NamespacedName]*corev1.ConfigMap{
					{Namespace: policyNs, Name: "test-cm"}: {},
				},
			},
			wantErr: true,
		},
		{
			name: "TypeValueRefConfigMap",
			contextExtensions: []*egv1a1.ContextExtension{{
				Name: "foo",
				Type: egv1a1.ContextExtensionValueTypeValueRef,
				ValueRef: &egv1a1.LocalObjectKeyReference{
					LocalObjectReference: gwapiv1.LocalObjectReference{
						Kind: resource.KindConfigMap,
						Name: "test-cm",
					},
					Key: "test-key",
				},
			}},
			translatorContext: &TranslatorContext{
				ConfigMapMap: map[types.NamespacedName]*corev1.ConfigMap{
					{Namespace: policyNs, Name: "test-cm"}: {
						Data: map[string]string{"test-key": "bar"},
					},
				},
			},
			want: []*ir.ContextExtention{{Name: "foo", Value: ir.PrivateBytes("bar")}},
		},
		{
			name: "TypeValueRefSecretNotFound",
			contextExtensions: []*egv1a1.ContextExtension{{
				Name: "foo",
				Type: egv1a1.ContextExtensionValueTypeValueRef,
				ValueRef: &egv1a1.LocalObjectKeyReference{
					LocalObjectReference: gwapiv1.LocalObjectReference{
						Kind: resource.KindSecret,
						Name: "test-secret",
					},
					Key: "test-key",
				},
			}},
			translatorContext: &TranslatorContext{},
			wantErr:           true,
		},
		{
			name: "TypeValueRefSecretKeyNotFound",
			contextExtensions: []*egv1a1.ContextExtension{{
				Name: "foo",
				Type: egv1a1.ContextExtensionValueTypeValueRef,
				ValueRef: &egv1a1.LocalObjectKeyReference{
					LocalObjectReference: gwapiv1.LocalObjectReference{
						Kind: resource.KindSecret,
						Name: "test-secret",
					},
					Key: "test-key",
				},
			}},
			translatorContext: &TranslatorContext{
				SecretMap: map[types.NamespacedName]*corev1.Secret{
					{Namespace: policyNs, Name: "test-secret"}: {},
				},
			},
			wantErr: true,
		},
		{
			name: "TypeValueRefSecret",
			contextExtensions: []*egv1a1.ContextExtension{{
				Name: "foo",
				Type: egv1a1.ContextExtensionValueTypeValueRef,
				ValueRef: &egv1a1.LocalObjectKeyReference{
					LocalObjectReference: gwapiv1.LocalObjectReference{
						Kind: resource.KindSecret,
						Name: "test-secret",
					},
					Key: "test-key",
				},
			}},
			translatorContext: &TranslatorContext{
				SecretMap: map[types.NamespacedName]*corev1.Secret{
					{Namespace: policyNs, Name: "test-secret"}: {
						Data: map[string][]byte{"test-key": []byte("YmFy")},
					},
				},
			},
			want: []*ir.ContextExtention{{Name: "foo", Value: ir.PrivateBytes("bar")}},
		},
		{
			name: "TypeValueRefUnexpectedKind",
			contextExtensions: []*egv1a1.ContextExtension{{
				Name: "foo",
				Type: egv1a1.ContextExtensionValueTypeValueRef,
				ValueRef: &egv1a1.LocalObjectKeyReference{
					LocalObjectReference: gwapiv1.LocalObjectReference{
						Kind: resource.KindService,
						Name: "test-secret",
					},
					Key: "test-key",
				},
			}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			translator := &Translator{TranslatorContext: tt.translatorContext}
			got, err := translator.buildContextExtensions(tt.contextExtensions, policyNs)
			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, got)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
