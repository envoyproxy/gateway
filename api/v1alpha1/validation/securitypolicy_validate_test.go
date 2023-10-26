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

func TestValidateSecurityPolicy(t *testing.T) {
	testCases := []struct {
		name     string
		policy   *egv1a1.SecurityPolicy
		expected bool
	}{
		{
			name:     "nil security policy",
			policy:   nil,
			expected: false,
		},
		{
			name: "empty security policy",
			policy: &egv1a1.SecurityPolicy{
				TypeMeta: metav1.TypeMeta{
					Kind:       egv1a1.KindSecurityPolicy,
					APIVersion: egv1a1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.SecurityPolicySpec{},
			},
			expected: false,
		},
		{
			name: "valid security policy with url",
			policy: &egv1a1.SecurityPolicy{
				TypeMeta: metav1.TypeMeta{
					Kind:       egv1a1.KindSecurityPolicy,
					APIVersion: egv1a1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.SecurityPolicySpec{
					JWT: &egv1a1.JWT{
						Providers: []egv1a1.JWTProvider{
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
			},
			expected: true,
		},
		{
			name: "valid security policy with email",
			policy: &egv1a1.SecurityPolicy{
				TypeMeta: metav1.TypeMeta{
					Kind:       egv1a1.KindSecurityPolicy,
					APIVersion: egv1a1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.SecurityPolicySpec{
					JWT: &egv1a1.JWT{
						Providers: []egv1a1.JWTProvider{
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
			},
			expected: true,
		},
		{
			name: "valid security policy with jwtClaimToHeader",
			policy: &egv1a1.SecurityPolicy{
				TypeMeta: metav1.TypeMeta{
					Kind:       egv1a1.KindSecurityPolicy,
					APIVersion: egv1a1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.SecurityPolicySpec{
					JWT: &egv1a1.JWT{
						Providers: []egv1a1.JWTProvider{
							{
								Name:      "test",
								Issuer:    "test@test.local",
								Audiences: []string{"test.local"},
								RemoteJWKS: egv1a1.RemoteJWKS{
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
				},
			},
			expected: true,
		},
		{
			name: "unqualified authentication provider name",
			policy: &egv1a1.SecurityPolicy{
				TypeMeta: metav1.TypeMeta{
					Kind:       egv1a1.KindSecurityPolicy,
					APIVersion: egv1a1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.SecurityPolicySpec{
					JWT: &egv1a1.JWT{
						Providers: []egv1a1.JWTProvider{
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
			},
			expected: false,
		},
		{
			name: "unspecified provider name",
			policy: &egv1a1.SecurityPolicy{
				TypeMeta: metav1.TypeMeta{
					Kind:       egv1a1.KindSecurityPolicy,
					APIVersion: egv1a1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.SecurityPolicySpec{
					JWT: &egv1a1.JWT{
						Providers: []egv1a1.JWTProvider{
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
			},
			expected: false,
		},
		{
			name: "non unique provider names",
			policy: &egv1a1.SecurityPolicy{
				TypeMeta: metav1.TypeMeta{
					Kind:       egv1a1.KindSecurityPolicy,
					APIVersion: egv1a1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.SecurityPolicySpec{
					JWT: &egv1a1.JWT{
						Providers: []egv1a1.JWTProvider{
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
			},
			expected: false,
		},
		{
			name: "invalid issuer uri",
			policy: &egv1a1.SecurityPolicy{
				TypeMeta: metav1.TypeMeta{
					Kind:       egv1a1.KindSecurityPolicy,
					APIVersion: egv1a1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.SecurityPolicySpec{
					JWT: &egv1a1.JWT{
						Providers: []egv1a1.JWTProvider{
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
			},
			expected: false,
		},
		{
			name: "inivalid issuer email",
			policy: &egv1a1.SecurityPolicy{
				TypeMeta: metav1.TypeMeta{
					Kind:       egv1a1.KindSecurityPolicy,
					APIVersion: egv1a1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.SecurityPolicySpec{
					JWT: &egv1a1.JWT{
						Providers: []egv1a1.JWTProvider{
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
			},
			expected: false,
		},
		{
			name: "invalid remote jwks uri",
			policy: &egv1a1.SecurityPolicy{
				TypeMeta: metav1.TypeMeta{
					Kind:       egv1a1.KindSecurityPolicy,
					APIVersion: egv1a1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.SecurityPolicySpec{
					JWT: &egv1a1.JWT{
						Providers: []egv1a1.JWTProvider{
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
			},
			expected: false,
		},
		{
			name: "unspecified remote jwks uri",
			policy: &egv1a1.SecurityPolicy{
				TypeMeta: metav1.TypeMeta{
					Kind:       egv1a1.KindSecurityPolicy,
					APIVersion: egv1a1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.SecurityPolicySpec{
					JWT: &egv1a1.JWT{
						Providers: []egv1a1.JWTProvider{
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
			},
			expected: false,
		},
		{
			name: "unspecified jwtClaimToHeader headerName",
			policy: &egv1a1.SecurityPolicy{
				TypeMeta: metav1.TypeMeta{
					Kind:       egv1a1.KindSecurityPolicy,
					APIVersion: egv1a1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.SecurityPolicySpec{
					JWT: &egv1a1.JWT{
						Providers: []egv1a1.JWTProvider{
							{
								Name:      "test",
								Issuer:    "test@test.local",
								Audiences: []string{"test.local"},
								RemoteJWKS: egv1a1.RemoteJWKS{
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
					},
				},
			},
			expected: false,
		},
		{
			name: "unspecified jwtClaimToHeader claimName",
			policy: &egv1a1.SecurityPolicy{
				TypeMeta: metav1.TypeMeta{
					Kind:       egv1a1.KindSecurityPolicy,
					APIVersion: egv1a1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.SecurityPolicySpec{
					JWT: &egv1a1.JWT{
						Providers: []egv1a1.JWTProvider{
							{
								Name:      "test",
								Issuer:    "test@test.local",
								Audiences: []string{"test.local"},
								RemoteJWKS: egv1a1.RemoteJWKS{
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
					},
				},
			},
			expected: false,
		},
		{
			name: "unspecified issuer",
			policy: &egv1a1.SecurityPolicy{
				TypeMeta: metav1.TypeMeta{
					Kind:       egv1a1.KindSecurityPolicy,
					APIVersion: egv1a1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.SecurityPolicySpec{
					JWT: &egv1a1.JWT{
						Providers: []egv1a1.JWTProvider{
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
			},
			expected: true,
		},
		{
			name: "unspecified audiences",
			policy: &egv1a1.SecurityPolicy{
				TypeMeta: metav1.TypeMeta{
					Kind:       egv1a1.KindSecurityPolicy,
					APIVersion: egv1a1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.SecurityPolicySpec{
					JWT: &egv1a1.JWT{
						Providers: []egv1a1.JWTProvider{
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
			},
			expected: true,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateSecurityPolicy(tc.policy)
			if tc.expected {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
