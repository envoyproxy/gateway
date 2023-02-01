// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package validation

import (
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func TestValidateAuthenticationFilter(t *testing.T) {
	testCases := []struct {
		name     string
		filter   *egv1a1.AuthenticationFilter
		expected bool
	}{
		{
			name:     "nil authentication filter",
			filter:   nil,
			expected: false,
		},
		{
			name: "valid authentication filter with url",
			filter: &egv1a1.AuthenticationFilter{
				TypeMeta: metav1.TypeMeta{
					Kind:       egv1a1.KindAuthenticationFilter,
					APIVersion: egv1a1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.AuthenticationFilterSpec{
					Type: egv1a1.JwtAuthenticationFilterProviderType,
					JwtProviders: []egv1a1.JwtAuthenticationFilterProvider{
						{
							Name:      "test",
							Issuer:    "https://www.test.local",
							Audiences: []string{"test.local"},
							RemoteJWKS: egv1a1.RemoteJWKS{
								URI: "https://test.local/jwt/public-key/jwks.json",
							},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "valid authentication filter with email",
			filter: &egv1a1.AuthenticationFilter{
				TypeMeta: metav1.TypeMeta{
					Kind:       egv1a1.KindAuthenticationFilter,
					APIVersion: egv1a1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.AuthenticationFilterSpec{
					Type: egv1a1.JwtAuthenticationFilterProviderType,
					JwtProviders: []egv1a1.JwtAuthenticationFilterProvider{
						{
							Name:      "test",
							Issuer:    "test@test.local",
							Audiences: []string{"test.local"},
							RemoteJWKS: egv1a1.RemoteJWKS{
								URI: "https://test.local/jwt/public-key/jwks.json",
							},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "unqualified authentication provider name",
			filter: &egv1a1.AuthenticationFilter{
				TypeMeta: metav1.TypeMeta{
					Kind:       egv1a1.KindAuthenticationFilter,
					APIVersion: egv1a1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.AuthenticationFilterSpec{
					Type: egv1a1.JwtAuthenticationFilterProviderType,
					JwtProviders: []egv1a1.JwtAuthenticationFilterProvider{
						{
							Name:      "unqualified_...",
							Issuer:    "https://www.test.local",
							Audiences: []string{"test.local"},
							RemoteJWKS: egv1a1.RemoteJWKS{
								URI: "https://test.local/jwt/public-key/jwks.json",
							},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "unspecified provider name",
			filter: &egv1a1.AuthenticationFilter{
				TypeMeta: metav1.TypeMeta{
					Kind:       egv1a1.KindAuthenticationFilter,
					APIVersion: egv1a1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.AuthenticationFilterSpec{
					Type: egv1a1.JwtAuthenticationFilterProviderType,
					JwtProviders: []egv1a1.JwtAuthenticationFilterProvider{
						{
							Name:      "",
							Issuer:    "https://www.test.local",
							Audiences: []string{"test.local"},
							RemoteJWKS: egv1a1.RemoteJWKS{
								URI: "https://test.local/jwt/public-key/jwks.json",
							},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "non unique provider names",
			filter: &egv1a1.AuthenticationFilter{
				TypeMeta: metav1.TypeMeta{
					Kind:       egv1a1.KindAuthenticationFilter,
					APIVersion: egv1a1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.AuthenticationFilterSpec{
					Type: egv1a1.JwtAuthenticationFilterProviderType,
					JwtProviders: []egv1a1.JwtAuthenticationFilterProvider{
						{
							Name:      "unique",
							Issuer:    "https://www.test.local",
							Audiences: []string{"test.local"},
							RemoteJWKS: egv1a1.RemoteJWKS{
								URI: "https://test.local/jwt/public-key/jwks.json",
							},
						},
						{
							Name:      "non-unique",
							Issuer:    "https://www.test.local",
							Audiences: []string{"test.local"},
							RemoteJWKS: egv1a1.RemoteJWKS{
								URI: "https://test.local/jwt/public-key/jwks.json",
							},
						},
						{
							Name:      "non-unique",
							Issuer:    "https://www.test.local",
							Audiences: []string{"test.local"},
							RemoteJWKS: egv1a1.RemoteJWKS{
								URI: "https://test.local/jwt/public-key/jwks.json",
							},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "invalid issuer uri",
			filter: &egv1a1.AuthenticationFilter{
				TypeMeta: metav1.TypeMeta{
					Kind:       egv1a1.KindAuthenticationFilter,
					APIVersion: egv1a1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.AuthenticationFilterSpec{
					Type: egv1a1.JwtAuthenticationFilterProviderType,
					JwtProviders: []egv1a1.JwtAuthenticationFilterProvider{
						{
							Name:      "test",
							Issuer:    "http://invalid url.local",
							Audiences: []string{"test.local"},
							RemoteJWKS: egv1a1.RemoteJWKS{
								URI: "http://www.test.local",
							},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "inivalid issuer email",
			filter: &egv1a1.AuthenticationFilter{
				TypeMeta: metav1.TypeMeta{
					Kind:       egv1a1.KindAuthenticationFilter,
					APIVersion: egv1a1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.AuthenticationFilterSpec{
					Type: egv1a1.JwtAuthenticationFilterProviderType,
					JwtProviders: []egv1a1.JwtAuthenticationFilterProvider{
						{
							Name:      "test",
							Issuer:    "test@!123...",
							Audiences: []string{"test.local"},
							RemoteJWKS: egv1a1.RemoteJWKS{
								URI: "https://test.local/jwt/public-key/jwks.json",
							},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "invalid remote jwks uri",
			filter: &egv1a1.AuthenticationFilter{
				TypeMeta: metav1.TypeMeta{
					Kind:       egv1a1.KindAuthenticationFilter,
					APIVersion: egv1a1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.AuthenticationFilterSpec{
					Type: egv1a1.JwtAuthenticationFilterProviderType,
					JwtProviders: []egv1a1.JwtAuthenticationFilterProvider{
						{
							Name:      "test",
							Issuer:    "http://www.test.local",
							Audiences: []string{"test.local"},
							RemoteJWKS: egv1a1.RemoteJWKS{
								URI: "invalid/local",
							},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "unspecified remote jwks uri",
			filter: &egv1a1.AuthenticationFilter{
				TypeMeta: metav1.TypeMeta{
					Kind:       egv1a1.KindAuthenticationFilter,
					APIVersion: egv1a1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.AuthenticationFilterSpec{
					Type: egv1a1.JwtAuthenticationFilterProviderType,
					JwtProviders: []egv1a1.JwtAuthenticationFilterProvider{
						{
							Name:      "test",
							Audiences: []string{"test.local"},
							RemoteJWKS: egv1a1.RemoteJWKS{
								URI: "",
							},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "unspecified issuer",
			filter: &egv1a1.AuthenticationFilter{
				TypeMeta: metav1.TypeMeta{
					Kind:       egv1a1.KindAuthenticationFilter,
					APIVersion: egv1a1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.AuthenticationFilterSpec{
					Type: egv1a1.JwtAuthenticationFilterProviderType,
					JwtProviders: []egv1a1.JwtAuthenticationFilterProvider{
						{
							Name:      "test",
							Audiences: []string{"test.local"},
							RemoteJWKS: egv1a1.RemoteJWKS{
								URI: "https://test.local/jwt/public-key/jwks.json",
							},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "unspecified audiences",
			filter: &egv1a1.AuthenticationFilter{
				TypeMeta: metav1.TypeMeta{
					Kind:       egv1a1.KindAuthenticationFilter,
					APIVersion: egv1a1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.AuthenticationFilterSpec{
					Type: egv1a1.JwtAuthenticationFilterProviderType,
					JwtProviders: []egv1a1.JwtAuthenticationFilterProvider{
						{
							Name:   "test",
							Issuer: "https://www.test.local",
							RemoteJWKS: egv1a1.RemoteJWKS{
								URI: "https://test.local/jwt/public-key/jwks.json",
							},
						},
					},
				},
			},
			expected: true,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateAuthenticationFilter(tc.filter)
			if tc.expected {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
