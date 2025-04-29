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

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

var inPath = "./testdata/decoder/in/"

func TestDecode(t *testing.T) {
	testCases := []struct {
		in     string
		out    *egv1a1.EnvoyGateway
		expect bool
	}{
		{
			in: inPath + "kube-provider.yaml",
			out: &egv1a1.EnvoyGateway{
				TypeMeta: metav1.TypeMeta{
					Kind:       egv1a1.KindEnvoyGateway,
					APIVersion: egv1a1.GroupVersion.String(),
				},
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Provider: egv1a1.DefaultEnvoyGatewayProvider(),
				},
			},
			expect: true,
		},
		{
			in: inPath + "gateway-controller-name.yaml",
			out: &egv1a1.EnvoyGateway{
				TypeMeta: metav1.TypeMeta{
					Kind:       egv1a1.KindEnvoyGateway,
					APIVersion: egv1a1.GroupVersion.String(),
				},
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway: egv1a1.DefaultGateway(),
				},
			},
			expect: true,
		},
		{
			in: inPath + "provider-with-gateway.yaml",
			out: &egv1a1.EnvoyGateway{
				TypeMeta: metav1.TypeMeta{
					Kind:       egv1a1.KindEnvoyGateway,
					APIVersion: egv1a1.GroupVersion.String(),
				},
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway:  egv1a1.DefaultGateway(),
					Provider: egv1a1.DefaultEnvoyGatewayProvider(),
				},
			},
			expect: true,
		},
		{
			in: inPath + "provider-mixing-gateway.yaml",
			out: &egv1a1.EnvoyGateway{
				TypeMeta: metav1.TypeMeta{
					Kind:       egv1a1.KindEnvoyGateway,
					APIVersion: egv1a1.GroupVersion.String(),
				},
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Provider: egv1a1.DefaultEnvoyGatewayProvider(),
				},
			},
			expect: true,
		},
		{
			in: inPath + "gateway-mixing-provider.yaml",
			out: &egv1a1.EnvoyGateway{
				TypeMeta: metav1.TypeMeta{
					Kind:       egv1a1.KindEnvoyGateway,
					APIVersion: egv1a1.GroupVersion.String(),
				},
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway: egv1a1.DefaultGateway(),
				},
			},
			expect: true,
		},
		{
			in: inPath + "provider-mixing-gateway.yaml",
			out: &egv1a1.EnvoyGateway{
				TypeMeta: metav1.TypeMeta{
					Kind:       egv1a1.KindEnvoyGateway,
					APIVersion: egv1a1.GroupVersion.String(),
				},
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Provider: egv1a1.DefaultEnvoyGatewayProvider(),
					Gateway:  egv1a1.DefaultGateway(),
				},
			},
			expect: false,
		},
		{
			in: inPath + "gateway-mixing-provider.yaml",
			out: &egv1a1.EnvoyGateway{
				TypeMeta: metav1.TypeMeta{
					Kind:       egv1a1.KindEnvoyGateway,
					APIVersion: egv1a1.GroupVersion.String(),
				},
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Provider: egv1a1.DefaultEnvoyGatewayProvider(),
					Gateway:  egv1a1.DefaultGateway(),
				},
			},
			expect: false,
		},
		{
			in: inPath + "gateway-ratelimit.yaml",
			out: &egv1a1.EnvoyGateway{
				TypeMeta: metav1.TypeMeta{
					Kind:       egv1a1.KindEnvoyGateway,
					APIVersion: egv1a1.GroupVersion.String(),
				},
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway: egv1a1.DefaultGateway(),
					Provider: &egv1a1.EnvoyGatewayProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyGatewayKubernetesProvider{
							RateLimitDeployment: &egv1a1.KubernetesDeploymentSpec{
								Strategy: egv1a1.DefaultKubernetesDeploymentStrategy(),
								Container: &egv1a1.KubernetesContainerSpec{
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
									Resources: egv1a1.DefaultResourceRequirements(),
									SecurityContext: &corev1.SecurityContext{
										RunAsUser:                ptr.To[int64](2000),
										AllowPrivilegeEscalation: ptr.To(false),
									},
								},
								Pod: &egv1a1.KubernetesPodSpec{
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
					RateLimit: &egv1a1.RateLimit{
						Backend: egv1a1.RateLimitDatabaseBackend{
							Type: egv1a1.RedisBackendType,
							Redis: &egv1a1.RateLimitRedisSettings{
								URL: "localhost:6379",
								TLS: &egv1a1.RedisTLSSettings{
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
			out: &egv1a1.EnvoyGateway{
				TypeMeta: metav1.TypeMeta{
					Kind:       egv1a1.KindEnvoyGateway,
					APIVersion: egv1a1.GroupVersion.String(),
				},
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Provider: egv1a1.DefaultEnvoyGatewayProvider(),
					Gateway:  egv1a1.DefaultGateway(),
					RateLimit: &egv1a1.RateLimit{
						Timeout: &metav1.Duration{
							Duration: 10000000,
						},
						FailClosed: true,
						Backend: egv1a1.RateLimitDatabaseBackend{
							Type: egv1a1.RedisBackendType,
							Redis: &egv1a1.RateLimitRedisSettings{
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
			out: &egv1a1.EnvoyGateway{
				TypeMeta: metav1.TypeMeta{
					Kind:       egv1a1.KindEnvoyGateway,
					APIVersion: egv1a1.GroupVersion.String(),
				},
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Provider: &egv1a1.EnvoyGatewayProvider{
						Type: egv1a1.ProviderTypeKubernetes,
					},
					Gateway: egv1a1.DefaultGateway(),
					Logging: &egv1a1.EnvoyGatewayLogging{
						Level: map[egv1a1.EnvoyGatewayLogComponent]egv1a1.LogLevel{
							egv1a1.LogComponentGatewayDefault: egv1a1.LogLevelInfo,
						},
					},
				},
			},
			expect: true,
		},
		{
			in: inPath + "gateway-ns-watch.yaml",
			out: &egv1a1.EnvoyGateway{
				TypeMeta: metav1.TypeMeta{
					Kind:       egv1a1.KindEnvoyGateway,
					APIVersion: egv1a1.GroupVersion.String(),
				},
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Provider: &egv1a1.EnvoyGatewayProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyGatewayKubernetesProvider{
							Watch: &egv1a1.KubernetesWatchMode{
								Type: egv1a1.KubernetesWatchModeTypeNamespaces,
								Namespaces: []string{
									"ns-a",
									"ns-b",
								},
							},
						},
					},
					Gateway: egv1a1.DefaultGateway(),
				},
			},
			expect: true,
		},
		{
			in: inPath + "gateway-nsselector-watch.yaml",
			out: &egv1a1.EnvoyGateway{
				TypeMeta: metav1.TypeMeta{
					Kind:       egv1a1.KindEnvoyGateway,
					APIVersion: egv1a1.GroupVersion.String(),
				},
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Provider: &egv1a1.EnvoyGatewayProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyGatewayKubernetesProvider{
							Watch: &egv1a1.KubernetesWatchMode{
								Type: egv1a1.KubernetesWatchModeTypeNamespaceSelector,
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
					Gateway: egv1a1.DefaultGateway(),
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
		{
			in: inPath + "gateway-leaderelection.yaml",
			out: &egv1a1.EnvoyGateway{
				TypeMeta: metav1.TypeMeta{
					Kind:       egv1a1.KindEnvoyGateway,
					APIVersion: egv1a1.GroupVersion.String(),
				},
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway: egv1a1.DefaultGateway(),
					Provider: &egv1a1.EnvoyGatewayProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyGatewayKubernetesProvider{
							LeaderElection: &egv1a1.LeaderElection{
								Disable:       ptr.To(true),
								LeaseDuration: ptr.To(gwapiv1.Duration("1s")),
								RenewDeadline: ptr.To(gwapiv1.Duration("2s")),
								RetryPeriod:   ptr.To(gwapiv1.Duration("3s")),
							},
						},
					},
				},
			},
			expect: true,
		},
		{
			in: inPath + "gateway-k8s-client-ratelimit.yaml",
			out: &egv1a1.EnvoyGateway{
				TypeMeta: metav1.TypeMeta{
					Kind:       egv1a1.KindEnvoyGateway,
					APIVersion: egv1a1.GroupVersion.String(),
				},
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway: egv1a1.DefaultGateway(),
					Provider: &egv1a1.EnvoyGatewayProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyGatewayKubernetesProvider{
							Client: &egv1a1.KubernetesClient{
								RateLimit: &egv1a1.KubernetesClientRateLimit{
									QPS:   ptr.To[int32](500),
									Burst: ptr.To[int32](1000),
								},
							},
						},
					},
				},
			},
			expect: true,
		},
	}

	for _, tc := range testCases {
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
