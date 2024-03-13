// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package config

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/envoyproxy/gateway/api/v1alpha1"
)

var (
	inPath = "./testdata/decoder/in/"
)

func TestDecode(t *testing.T) {
	testCases := []struct {
		in     string
		out    *v1alpha1.EnvoyGateway
		expect bool
	}{
		{
			in: inPath + "kube-provider.yaml",
			out: &v1alpha1.EnvoyGateway{
				TypeMeta: metav1.TypeMeta{
					Kind:       v1alpha1.KindEnvoyGateway,
					APIVersion: v1alpha1.GroupVersion.String(),
				},
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Provider: v1alpha1.DefaultEnvoyGatewayProvider(),
				},
			},
			expect: true,
		},
		{
			in: inPath + "gateway-controller-name.yaml",
			out: &v1alpha1.EnvoyGateway{
				TypeMeta: metav1.TypeMeta{
					Kind:       v1alpha1.KindEnvoyGateway,
					APIVersion: v1alpha1.GroupVersion.String(),
				},
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Gateway: v1alpha1.DefaultGateway(),
				},
			},
			expect: true,
		},
		{
			in: inPath + "provider-with-gateway.yaml",
			out: &v1alpha1.EnvoyGateway{
				TypeMeta: metav1.TypeMeta{
					Kind:       v1alpha1.KindEnvoyGateway,
					APIVersion: v1alpha1.GroupVersion.String(),
				},
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Gateway:  v1alpha1.DefaultGateway(),
					Provider: v1alpha1.DefaultEnvoyGatewayProvider(),
				},
			},
			expect: true,
		},
		{
			in: inPath + "provider-mixing-gateway.yaml",
			out: &v1alpha1.EnvoyGateway{
				TypeMeta: metav1.TypeMeta{
					Kind:       v1alpha1.KindEnvoyGateway,
					APIVersion: v1alpha1.GroupVersion.String(),
				},
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Provider: v1alpha1.DefaultEnvoyGatewayProvider(),
				},
			},
			expect: true,
		},
		{
			in: inPath + "gateway-mixing-provider.yaml",
			out: &v1alpha1.EnvoyGateway{
				TypeMeta: metav1.TypeMeta{
					Kind:       v1alpha1.KindEnvoyGateway,
					APIVersion: v1alpha1.GroupVersion.String(),
				},
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Gateway: v1alpha1.DefaultGateway(),
				},
			},
			expect: true,
		},
		{
			in: inPath + "provider-mixing-gateway.yaml",
			out: &v1alpha1.EnvoyGateway{
				TypeMeta: metav1.TypeMeta{
					Kind:       v1alpha1.KindEnvoyGateway,
					APIVersion: v1alpha1.GroupVersion.String(),
				},
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Provider: v1alpha1.DefaultEnvoyGatewayProvider(),
					Gateway:  v1alpha1.DefaultGateway(),
				},
			},
			expect: false,
		},
		{
			in: inPath + "gateway-mixing-provider.yaml",
			out: &v1alpha1.EnvoyGateway{
				TypeMeta: metav1.TypeMeta{
					Kind:       v1alpha1.KindEnvoyGateway,
					APIVersion: v1alpha1.GroupVersion.String(),
				},
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Provider: v1alpha1.DefaultEnvoyGatewayProvider(),
					Gateway:  v1alpha1.DefaultGateway(),
				},
			},
			expect: false,
		},
		{
			in: inPath + "gateway-ratelimit.yaml",
			out: &v1alpha1.EnvoyGateway{
				TypeMeta: metav1.TypeMeta{
					Kind:       v1alpha1.KindEnvoyGateway,
					APIVersion: v1alpha1.GroupVersion.String(),
				},
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Gateway: v1alpha1.DefaultGateway(),
					Provider: &v1alpha1.EnvoyGatewayProvider{
						Type: v1alpha1.ProviderTypeKubernetes,
						Kubernetes: &v1alpha1.EnvoyGatewayKubernetesProvider{
							RateLimitDeployment: &v1alpha1.KubernetesDeploymentSpec{
								Strategy: v1alpha1.DefaultKubernetesDeploymentStrategy(),
								Container: &v1alpha1.KubernetesContainerSpec{
									Env: []corev1.EnvVar{
										{
											Name:  "env_a",
											Value: "env_a_value",
										},
										{
											Name:  "env_b",
											Value: "env_b_value",
										},
									},
									Image:     ptr.To("envoyproxy/ratelimit:latest"),
									Resources: v1alpha1.DefaultResourceRequirements(),
									SecurityContext: &corev1.SecurityContext{
										RunAsUser:                ptr.To[int64](2000),
										AllowPrivilegeEscalation: ptr.To(false),
									},
								},
								Pod: &v1alpha1.KubernetesPodSpec{
									Annotations: map[string]string{
										"key1": "val1",
										"key2": "val2",
									},
									SecurityContext: &corev1.PodSecurityContext{
										RunAsUser:           ptr.To[int64](1000),
										RunAsGroup:          ptr.To[int64](3000),
										FSGroup:             ptr.To[int64](2000),
										FSGroupChangePolicy: func(s corev1.PodFSGroupChangePolicy) *corev1.PodFSGroupChangePolicy { return &s }(corev1.FSGroupChangeOnRootMismatch),
									},
								},
							},
						},
					},
					RateLimit: &v1alpha1.RateLimit{
						Backend: v1alpha1.RateLimitDatabaseBackend{
							Type: v1alpha1.RedisBackendType,
							Redis: &v1alpha1.RateLimitRedisSettings{
								URL: "localhost:6379",
								TLS: &v1alpha1.RedisTLSSettings{
									CertificateRef: &gwapiv1.SecretObjectReference{
										Name: "ratelimit-cert",
									},
								},
							},
						},
					},
				},
			},
			expect: true,
		},
		{
			in: inPath + "gateway-global-ratelimit.yaml",
			out: &v1alpha1.EnvoyGateway{
				TypeMeta: metav1.TypeMeta{
					Kind:       v1alpha1.KindEnvoyGateway,
					APIVersion: v1alpha1.GroupVersion.String(),
				},
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Provider: v1alpha1.DefaultEnvoyGatewayProvider(),
					Gateway:  v1alpha1.DefaultGateway(),
					RateLimit: &v1alpha1.RateLimit{
						Timeout: &metav1.Duration{
							Duration: 10000000,
						},
						FailClosed: true,
						Backend: v1alpha1.RateLimitDatabaseBackend{
							Type: v1alpha1.RedisBackendType,
							Redis: &v1alpha1.RateLimitRedisSettings{
								URL: "localhost:6379",
							},
						},
					},
				},
			},
			expect: true,
		},
		{
			in: inPath + "gateway-logging.yaml",
			out: &v1alpha1.EnvoyGateway{
				TypeMeta: metav1.TypeMeta{
					Kind:       v1alpha1.KindEnvoyGateway,
					APIVersion: v1alpha1.GroupVersion.String(),
				},
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Provider: &v1alpha1.EnvoyGatewayProvider{
						Type: v1alpha1.ProviderTypeKubernetes,
					},
					Gateway: v1alpha1.DefaultGateway(),
					Logging: &v1alpha1.EnvoyGatewayLogging{
						Level: map[v1alpha1.EnvoyGatewayLogComponent]v1alpha1.LogLevel{
							v1alpha1.LogComponentGatewayDefault: v1alpha1.LogLevelInfo,
						},
					},
				},
			},
			expect: true,
		},
		{
			in: inPath + "gateway-ns-watch.yaml",
			out: &v1alpha1.EnvoyGateway{
				TypeMeta: metav1.TypeMeta{
					Kind:       v1alpha1.KindEnvoyGateway,
					APIVersion: v1alpha1.GroupVersion.String(),
				},
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Provider: &v1alpha1.EnvoyGatewayProvider{
						Type: v1alpha1.ProviderTypeKubernetes,
						Kubernetes: &v1alpha1.EnvoyGatewayKubernetesProvider{
							Watch: &v1alpha1.KubernetesWatchMode{
								Type: v1alpha1.KubernetesWatchModeTypeNamespaces,
								Namespaces: []string{
									"ns-a",
									"ns-b",
								},
							},
						},
					},
					Gateway: v1alpha1.DefaultGateway(),
				},
			},
			expect: true,
		},
		{
			in: inPath + "gateway-nsselector-watch.yaml",
			out: &v1alpha1.EnvoyGateway{
				TypeMeta: metav1.TypeMeta{
					Kind:       v1alpha1.KindEnvoyGateway,
					APIVersion: v1alpha1.GroupVersion.String(),
				},
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Provider: &v1alpha1.EnvoyGatewayProvider{
						Type: v1alpha1.ProviderTypeKubernetes,
						Kubernetes: &v1alpha1.EnvoyGatewayKubernetesProvider{
							Watch: &v1alpha1.KubernetesWatchMode{
								Type: v1alpha1.KubernetesWatchModeTypeNamespaceSelector,
								NamespaceSelector: &metav1.LabelSelector{
									MatchLabels: map[string]string{"label-a": "foo"},
									MatchExpressions: []metav1.LabelSelectorRequirement{
										{
											Key:      "tier",
											Operator: metav1.LabelSelectorOpIn,
											Values:   []string{"cache"},
										},
										{
											Key:      "environment",
											Operator: metav1.LabelSelectorOpNotIn,
											Values:   []string{"dev"},
										},
									},
								},
							},
						},
					},
					Gateway: v1alpha1.DefaultGateway(),
				},
			},
			expect: true,
		},
		{
			in:     inPath + "invalid-gateway-logging.yaml",
			expect: false,
		},
		{
			in:     inPath + "no-api-version.yaml",
			expect: false,
		},
		{
			in:     inPath + "no-kind.yaml",
			expect: false,
		},
		{
			in:     "/non/existent/config.yaml",
			expect: false,
		},
		{
			in:     inPath + "invalid-gateway-group.yaml",
			expect: false,
		},
		{
			in:     inPath + "invalid-gateway-kind.yaml",
			expect: false,
		},
		{
			in:     inPath + "invalid-gateway-version.yaml",
			expect: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.in, func(t *testing.T) {
			eg, err := Decode(tc.in)
			if tc.expect {
				require.NoError(t, err)
				require.Equal(t, tc.out, eg)
			} else {
				require.True(t, !reflect.DeepEqual(tc.out, eg) || err != nil)
			}
		})
	}
}
