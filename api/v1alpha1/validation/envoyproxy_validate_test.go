// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package validation

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func TestValidateEnvoyProxy(t *testing.T) {
	testCases := []struct {
		name     string
		proxy    *egv1a1.EnvoyProxy
		expected bool
	}{
		{
			name:     "nil egv1a1.EnvoyProxy",
			proxy:    nil,
			expected: false,
		},
		{
			name: "nil provider",
			proxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Provider: nil,
				},
			},
			expected: true,
		},
		{
			name: "unsupported provider",
			proxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.ProviderTypeCustom,
					},
				},
			},
			expected: false,
		},
		{
			name: "nil envoy service",
			proxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyService: nil,
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "unsupported envoy service type \"\" ",
			proxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyService: &egv1a1.KubernetesServiceSpec{
								Type: egv1a1.GetKubernetesServiceType(""),
							},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "valid envoy service type 'NodePort'",
			proxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyService: &egv1a1.KubernetesServiceSpec{
								Type: egv1a1.GetKubernetesServiceType(egv1a1.ServiceType(corev1.ServiceTypeNodePort)),
							},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "valid envoy service type 'LoadBalancer'",
			proxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyService: &egv1a1.KubernetesServiceSpec{
								Type: egv1a1.GetKubernetesServiceType(egv1a1.ServiceTypeLoadBalancer),
							},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "valid envoy service type 'ClusterIP'",
			proxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyService: &egv1a1.KubernetesServiceSpec{
								Type: egv1a1.GetKubernetesServiceType(egv1a1.ServiceTypeClusterIP),
							},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "envoy service type 'LoadBalancer' with allocateLoadBalancerNodePorts",
			proxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyService: &egv1a1.KubernetesServiceSpec{
								Type:                          egv1a1.GetKubernetesServiceType(egv1a1.ServiceTypeLoadBalancer),
								AllocateLoadBalancerNodePorts: ptr.To(false),
							},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "non envoy service type 'LoadBalancer' with allocateLoadBalancerNodePorts",
			proxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyService: &egv1a1.KubernetesServiceSpec{
								Type:                          egv1a1.GetKubernetesServiceType(egv1a1.ServiceTypeClusterIP),
								AllocateLoadBalancerNodePorts: ptr.To(false),
							},
						},
					},
				},
			},
			expected: false,
		},

		{
			name: "envoy service type 'LoadBalancer' with loadBalancerSourceRanges",
			proxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyService: &egv1a1.KubernetesServiceSpec{
								Type:                     egv1a1.GetKubernetesServiceType(egv1a1.ServiceTypeLoadBalancer),
								LoadBalancerSourceRanges: []string{"1.1.1.1/32"},
							},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "non envoy service type 'LoadBalancer' with loadBalancerSourceRanges",
			proxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyService: &egv1a1.KubernetesServiceSpec{
								Type:                     egv1a1.GetKubernetesServiceType(egv1a1.ServiceTypeClusterIP),
								LoadBalancerSourceRanges: []string{"1.1.1.1/32"},
							},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "envoy service type 'LoadBalancer' with valid loadBalancerIP",
			proxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyService: &egv1a1.KubernetesServiceSpec{
								Type:           egv1a1.GetKubernetesServiceType(egv1a1.ServiceTypeLoadBalancer),
								LoadBalancerIP: ptr.To("10.11.12.13"),
							},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "envoy service type 'LoadBalancer' with invalid loadBalancerIP",
			proxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyService: &egv1a1.KubernetesServiceSpec{
								Type:           egv1a1.GetKubernetesServiceType(egv1a1.ServiceTypeLoadBalancer),
								LoadBalancerIP: ptr.To("invalid-ip"),
							},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "envoy service type 'LoadBalancer' with ipv6 loadBalancerIP",
			proxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyService: &egv1a1.KubernetesServiceSpec{
								Type:           egv1a1.GetKubernetesServiceType(egv1a1.ServiceTypeLoadBalancer),
								LoadBalancerIP: ptr.To("2001:db8::68"),
							},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "should invalid when accesslog enabled using Text format, but `text` field being empty",
			proxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Telemetry: &egv1a1.ProxyTelemetry{
						AccessLog: &egv1a1.ProxyAccessLog{
							Settings: []egv1a1.ProxyAccessLogSetting{
								{
									Format: &egv1a1.ProxyAccessLogFormat{
										Type: egv1a1.ProxyAccessLogFormatTypeText,
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
			name: "should invalid when accesslog enabled using File sink, but `file` field being empty",
			proxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Telemetry: &egv1a1.ProxyTelemetry{
						AccessLog: &egv1a1.ProxyAccessLog{
							Settings: []egv1a1.ProxyAccessLogSetting{
								{
									Format: &egv1a1.ProxyAccessLogFormat{
										Type: egv1a1.ProxyAccessLogFormatTypeText,
										Text: ptr.To("[%START_TIME%]"),
									},
									Sinks: []egv1a1.ProxyAccessLogSink{
										{
											Type: egv1a1.ProxyAccessLogSinkTypeFile,
										},
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
			name: "should invalid when metrics type is OpenTelemetry, but `OpenTelemetry` field being empty",
			proxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Telemetry: &egv1a1.ProxyTelemetry{
						Metrics: &egv1a1.ProxyMetrics{
							Sinks: []egv1a1.ProxyMetricSink{
								{
									Type: egv1a1.MetricSinkTypeOpenTelemetry,
								},
							},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "should valid when metrics type is OpenTelemetry and `OpenTelemetry` field being not empty",
			proxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Telemetry: &egv1a1.ProxyTelemetry{
						Metrics: &egv1a1.ProxyMetrics{
							Sinks: []egv1a1.ProxyMetricSink{
								{
									Type: egv1a1.MetricSinkTypeOpenTelemetry,
									OpenTelemetry: &egv1a1.ProxyOpenTelemetrySink{
										Host: ptr.To("0.0.0.0"),
										Port: 3217,
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
			name: "should be valid when service patch is empty",
			proxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyService: &egv1a1.KubernetesServiceSpec{
								Patch: &egv1a1.KubernetesPatchSpec{
									Value: apiextensionsv1.JSON{
										Raw: []byte{},
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
			name: "should be valid when deployment patch is empty",
			proxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyDeployment: &egv1a1.KubernetesDeploymentSpec{
								Patch: &egv1a1.KubernetesPatchSpec{
									Value: apiextensionsv1.JSON{
										Raw: []byte{},
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
			name: "should be valid when pdb patch type and patch are empty",
			proxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyPDB: &egv1a1.KubernetesPodDisruptionBudgetSpec{
								Patch: &egv1a1.KubernetesPatchSpec{
									Value: apiextensionsv1.JSON{
										Raw: []byte{},
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
			name: "should be valid when pdb patch and type are set",
			proxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyPDB: &egv1a1.KubernetesPodDisruptionBudgetSpec{
								Patch: &egv1a1.KubernetesPatchSpec{
									Type: ptr.To(egv1a1.StrategicMerge),
									Value: apiextensionsv1.JSON{
										Raw: []byte("{}"),
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
			name: "should be invalid when pdb patch object is empty",
			proxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyPDB: &egv1a1.KubernetesPodDisruptionBudgetSpec{
								Patch: &egv1a1.KubernetesPatchSpec{
									Type: ptr.To(egv1a1.StrategicMerge),
								},
							},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "should be valid when pdb type not set",
			proxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyPDB: &egv1a1.KubernetesPodDisruptionBudgetSpec{
								Patch: &egv1a1.KubernetesPatchSpec{
									Value: apiextensionsv1.JSON{
										Raw: []byte("{}"),
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
			name: "should be valid when hpa patch and type are empty",
			proxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyHpa: &egv1a1.KubernetesHorizontalPodAutoscalerSpec{
								Patch: &egv1a1.KubernetesPatchSpec{
									Value: apiextensionsv1.JSON{
										Raw: []byte{},
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
			name: "should be valid when hpa patch and type are set",
			proxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyHpa: &egv1a1.KubernetesHorizontalPodAutoscalerSpec{
								Patch: &egv1a1.KubernetesPatchSpec{
									Type: ptr.To(egv1a1.StrategicMerge),
									Value: apiextensionsv1.JSON{
										Raw: []byte("{}"),
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
			name: "should be invalid when hpa patch object is empty",
			proxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyHpa: &egv1a1.KubernetesHorizontalPodAutoscalerSpec{
								Patch: &egv1a1.KubernetesPatchSpec{
									Type: ptr.To(egv1a1.StrategicMerge),
								},
							},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "should be valid when hpa type not set",
			proxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyHpa: &egv1a1.KubernetesHorizontalPodAutoscalerSpec{
								Patch: &egv1a1.KubernetesPatchSpec{
									Value: apiextensionsv1.JSON{
										Raw: []byte("{}"),
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
			name: "should invalid when deployment patch object is empty",
			proxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyDeployment: &egv1a1.KubernetesDeploymentSpec{
								Patch: &egv1a1.KubernetesPatchSpec{
									Type: ptr.To(egv1a1.StrategicMerge),
								},
							},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "should valid when deployment patch type and object are both not empty",
			proxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyDeployment: &egv1a1.KubernetesDeploymentSpec{
								Patch: &egv1a1.KubernetesPatchSpec{
									Type: ptr.To(egv1a1.StrategicMerge),
									Value: apiextensionsv1.JSON{
										Raw: []byte("{}"),
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
			name: "should valid when deployment patch type is empty and object is not empty",
			proxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyDeployment: &egv1a1.KubernetesDeploymentSpec{
								Patch: &egv1a1.KubernetesPatchSpec{
									Value: apiextensionsv1.JSON{
										Raw: []byte("{}"),
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
			name: "valid filter order",
			proxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.EnvoyProxySpec{
					FilterOrder: []egv1a1.FilterPosition{
						{
							Name:   egv1a1.EnvoyFilterOAuth2,
							Before: ptr.To(egv1a1.EnvoyFilterJWTAuthn),
						},
						{
							Name:  egv1a1.EnvoyFilterExtProc,
							After: ptr.To(egv1a1.EnvoyFilterJWTAuthn),
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "invalid filter order with circular dependency",
			proxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.EnvoyProxySpec{
					FilterOrder: []egv1a1.FilterPosition{
						{
							Name:   egv1a1.EnvoyFilterOAuth2,
							Before: ptr.To(egv1a1.EnvoyFilterJWTAuthn),
						},
						{
							Name:   egv1a1.EnvoyFilterJWTAuthn,
							Before: ptr.To(egv1a1.EnvoyFilterExtProc),
						},
						{
							Name:   egv1a1.EnvoyFilterExtProc,
							Before: ptr.To(egv1a1.EnvoyFilterOAuth2),
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "valid operators in ClusterStatName",
			proxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Telemetry: &egv1a1.ProxyTelemetry{
						Metrics: &egv1a1.ProxyMetrics{
							ClusterStatName: ptr.To(fmt.Sprintf("%s/%s/%s/%s/%s/%s/%s", egv1a1.StatFormatterRouteName,
								egv1a1.StatFormatterRouteName, egv1a1.StatFormatterRouteNamespace, egv1a1.StatFormatterRouteKind,
								egv1a1.StatFormatterRouteRuleName, egv1a1.StatFormatterRouteRuleNumber, egv1a1.StatFormatterBackendRefs)),
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "invalid operators in ClusterStatName",
			proxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Telemetry: &egv1a1.ProxyTelemetry{
						Metrics: &egv1a1.ProxyMetrics{
							ClusterStatName: ptr.To("%ROUTE_NAME%.%FOO%.%BAR%/my/%BACKEND_REFS%/%FOOBAR%"),
						},
					},
				},
			},
			expected: false,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateEnvoyProxy(tc.proxy)
			if tc.expected {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestEnvoyProxyProvider(t *testing.T) {
	envoyProxy := &egv1a1.EnvoyProxy{
		Spec: egv1a1.EnvoyProxySpec{
			Provider: egv1a1.DefaultEnvoyProxyProvider(),
		},
	}
	assert.NotNil(t, envoyProxy.Spec.Provider)

	envoyProxyProvider := envoyProxy.GetEnvoyProxyProvider()
	assert.Nil(t, envoyProxyProvider.Kubernetes)
	assert.True(t, reflect.DeepEqual(envoyProxy.Spec.Provider, envoyProxyProvider))

	envoyProxyKubeProvider := envoyProxyProvider.GetEnvoyProxyKubeProvider()

	assert.NotNil(t, envoyProxyProvider.Kubernetes)
	assert.True(t, reflect.DeepEqual(envoyProxyProvider.Kubernetes, envoyProxyKubeProvider))

	envoyProxyProvider.GetEnvoyProxyKubeProvider()

	assert.NotNil(t, envoyProxyProvider.Kubernetes.EnvoyDeployment)
	assert.Equal(t, envoyProxyProvider.Kubernetes.EnvoyDeployment, egv1a1.DefaultKubernetesDeployment(egv1a1.DefaultEnvoyProxyImage))
	assert.NotNil(t, envoyProxyProvider.Kubernetes.EnvoyDeployment.Pod)
	assert.Equal(t, envoyProxyProvider.Kubernetes.EnvoyDeployment.Pod, egv1a1.DefaultKubernetesPod())
	assert.NotNil(t, envoyProxyProvider.Kubernetes.EnvoyDeployment.Container)
	assert.Equal(t, envoyProxyProvider.Kubernetes.EnvoyDeployment.Container, egv1a1.DefaultKubernetesContainer(egv1a1.DefaultEnvoyProxyImage))
	assert.NotNil(t, envoyProxyProvider.Kubernetes.EnvoyDeployment.Container.Resources)
	assert.Equal(t, envoyProxyProvider.Kubernetes.EnvoyDeployment.Container.Resources, egv1a1.DefaultResourceRequirements())
	assert.NotNil(t, envoyProxyProvider.Kubernetes.EnvoyDeployment.Container.Image)
	assert.Equal(t, envoyProxyProvider.Kubernetes.EnvoyDeployment.Container.Image, egv1a1.DefaultKubernetesContainerImage(egv1a1.DefaultEnvoyProxyImage))

	assert.NotNil(t, envoyProxyProvider.Kubernetes.EnvoyService)
	assert.True(t, reflect.DeepEqual(envoyProxyProvider.Kubernetes.EnvoyService.Type, egv1a1.GetKubernetesServiceType(egv1a1.ServiceTypeLoadBalancer)))
}

func TestGetEnvoyProxyDefaultComponentLevel(t *testing.T) {
	cases := []struct {
		logging   egv1a1.ProxyLogging
		component egv1a1.ProxyLogComponent
		expected  egv1a1.LogLevel
	}{
		{
			logging: egv1a1.ProxyLogging{
				Level: map[egv1a1.ProxyLogComponent]egv1a1.LogLevel{
					egv1a1.LogComponentDefault: egv1a1.LogLevelInfo,
				},
			},
			expected: egv1a1.LogLevelInfo,
		},
		{
			logging: egv1a1.ProxyLogging{
				Level: map[egv1a1.ProxyLogComponent]egv1a1.LogLevel{
					egv1a1.LogComponentDefault: egv1a1.LogLevelInfo,
				},
			},
			expected: egv1a1.LogLevelInfo,
		},
	}

	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			got := tc.logging.DefaultEnvoyProxyLoggingLevel()
			require.Equal(t, tc.expected, got)
		})
	}
}

func TestGetEnvoyProxyComponentLevelArgs(t *testing.T) {
	cases := []struct {
		logging  egv1a1.ProxyLogging
		expected string
	}{
		{
			logging:  egv1a1.ProxyLogging{},
			expected: "misc:error",
		},
		{
			logging: egv1a1.ProxyLogging{
				Level: map[egv1a1.ProxyLogComponent]egv1a1.LogLevel{
					egv1a1.LogComponentDefault: egv1a1.LogLevelInfo,
				},
			},
			expected: "misc:error",
		},
		{
			logging: egv1a1.ProxyLogging{
				Level: map[egv1a1.ProxyLogComponent]egv1a1.LogLevel{
					egv1a1.LogComponentDefault: egv1a1.LogLevelInfo,
					egv1a1.LogComponentAdmin:   egv1a1.LogLevelWarn,
				},
			},
			expected: "admin:warn",
		},
		{
			logging: egv1a1.ProxyLogging{
				Level: map[egv1a1.ProxyLogComponent]egv1a1.LogLevel{
					egv1a1.LogComponentDefault: egv1a1.LogLevelInfo,
					egv1a1.LogComponentAdmin:   egv1a1.LogLevelWarn,
					egv1a1.LogComponentFilter:  egv1a1.LogLevelDebug,
				},
			},
			expected: "admin:warn,filter:debug",
		},
		{
			logging: egv1a1.ProxyLogging{
				Level: map[egv1a1.ProxyLogComponent]egv1a1.LogLevel{
					egv1a1.LogComponentDefault: egv1a1.LogLevelInfo,
					egv1a1.LogComponentAdmin:   egv1a1.LogLevelWarn,
					egv1a1.LogComponentFilter:  egv1a1.LogLevelDebug,
					egv1a1.LogComponentClient:  "",
				},
			},
			expected: "admin:warn,filter:debug",
		},
	}

	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			got := tc.logging.GetEnvoyProxyComponentLevel()
			require.Equal(t, tc.expected, got)
		})
	}
}
