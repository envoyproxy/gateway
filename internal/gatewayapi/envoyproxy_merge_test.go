// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func TestMergeEnvoyProxyConfigs(t *testing.T) {
	testCases := []struct {
		name              string
		defaultSpec       *egv1a1.EnvoyProxySpec
		gatewayClassProxy *egv1a1.EnvoyProxy
		gatewayProxy      *egv1a1.EnvoyProxy
		expectedSpec      *egv1a1.EnvoyProxySpec
		expectError       bool
	}{
		{
			name:         "no configs provided",
			expectedSpec: nil,
		},
		{
			name: "only default spec",
			defaultSpec: &egv1a1.EnvoyProxySpec{
				Concurrency: new(int32(4)),
			},
			expectedSpec: &egv1a1.EnvoyProxySpec{
				Concurrency: new(int32(4)),
			},
		},
		{
			name: "replace mode - gatewayclass overrides default",
			defaultSpec: &egv1a1.EnvoyProxySpec{
				Concurrency: new(int32(4)),
			},
			gatewayClassProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					Concurrency: new(int32(8)),
				},
			},
			expectedSpec: &egv1a1.EnvoyProxySpec{
				Concurrency: new(int32(8)),
			},
		},
		{
			name: "replace mode - gateway overrides all",
			defaultSpec: &egv1a1.EnvoyProxySpec{
				Concurrency: new(int32(4)),
			},
			gatewayClassProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					Concurrency: new(int32(8)),
				},
			},
			gatewayProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					Concurrency: new(int32(16)),
				},
			},
			expectedSpec: &egv1a1.EnvoyProxySpec{
				Concurrency: new(int32(16)),
			},
		},
		{
			name: "gateway mergeType controls gateway-over-gatewayclass step",
			defaultSpec: &egv1a1.EnvoyProxySpec{
				Concurrency: new(int32(4)),
			},
			gatewayClassProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					Concurrency: new(int32(8)),
				},
			},
			gatewayProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					MergeType:   new(egv1a1.StrategicMerge),
					Concurrency: new(int32(16)),
				},
			},
			expectedSpec: &egv1a1.EnvoyProxySpec{
				MergeType:   new(egv1a1.StrategicMerge),
				Concurrency: new(int32(16)),
			},
		},
		{
			name: "gateway nil mergeType - step1 Replace discards gatewayclass fields",
			defaultSpec: &egv1a1.EnvoyProxySpec{
				Concurrency: new(int32(4)),
			},
			gatewayClassProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					Concurrency: new(int32(8)),
					Logging: egv1a1.ProxyLogging{
						Level: map[egv1a1.ProxyLogComponent]egv1a1.LogLevel{
							egv1a1.LogComponentDefault: egv1a1.LogLevelInfo,
						},
					},
				},
			},
			gatewayProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					Concurrency: new(int32(16)),
				},
			},
			expectedSpec: &egv1a1.EnvoyProxySpec{
				Concurrency: new(int32(16)),
			},
		},
		{
			defaultSpec: &egv1a1.EnvoyProxySpec{
				Concurrency: new(int32(4)),
				Logging: egv1a1.ProxyLogging{
					Level: map[egv1a1.ProxyLogComponent]egv1a1.LogLevel{
						egv1a1.LogComponentDefault: egv1a1.LogLevelInfo,
					},
				},
			},
			gatewayClassProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					Concurrency: new(int32(8)),
				},
			},
			expectedSpec: &egv1a1.EnvoyProxySpec{
				Concurrency: new(int32(8)),
			},
		},
		{
			name: "gateway StrategicMerge - merges gateway+gatewayclass and defaults",
			defaultSpec: &egv1a1.EnvoyProxySpec{
				Concurrency: new(int32(4)),
			},
			gatewayClassProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					Concurrency: new(int32(8)),
					Logging: egv1a1.ProxyLogging{
						Level: map[egv1a1.ProxyLogComponent]egv1a1.LogLevel{
							egv1a1.LogComponentDefault: egv1a1.LogLevelInfo,
						},
					},
				},
			},
			gatewayProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					MergeType:   new(egv1a1.StrategicMerge),
					Concurrency: new(int32(16)),
				},
			},
			expectedSpec: &egv1a1.EnvoyProxySpec{
				MergeType:   new(egv1a1.StrategicMerge),
				Concurrency: new(int32(16)),
				Logging: egv1a1.ProxyLogging{
					Level: map[egv1a1.ProxyLogComponent]egv1a1.LogLevel{
						egv1a1.LogComponentDefault: egv1a1.LogLevelInfo,
					},
				},
			},
		},
		{
			name: "gateway StrategicMerge propagates to defaults merge - preserves default-only fields",
			defaultSpec: &egv1a1.EnvoyProxySpec{
				Concurrency: new(int32(4)),
				Logging: egv1a1.ProxyLogging{
					Level: map[egv1a1.ProxyLogComponent]egv1a1.LogLevel{
						egv1a1.LogComponentDefault: egv1a1.LogLevelInfo,
					},
				},
			},
			gatewayClassProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					Concurrency: new(int32(8)),
				},
			},
			gatewayProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					MergeType:   new(egv1a1.StrategicMerge),
					Concurrency: new(int32(16)),
				},
			},
			expectedSpec: &egv1a1.EnvoyProxySpec{
				MergeType:   new(egv1a1.StrategicMerge),
				Concurrency: new(int32(16)),
				Logging: egv1a1.ProxyLogging{
					Level: map[egv1a1.ProxyLogComponent]egv1a1.LogLevel{
						egv1a1.LogComponentDefault: egv1a1.LogLevelInfo,
					},
				},
			},
		},
		{
			name: "gatewayclass StrategicMerge - merges result with defaults",
			defaultSpec: &egv1a1.EnvoyProxySpec{
				Concurrency: new(int32(4)),
				Logging: egv1a1.ProxyLogging{
					Level: map[egv1a1.ProxyLogComponent]egv1a1.LogLevel{
						egv1a1.LogComponentDefault: egv1a1.LogLevelInfo,
					},
				},
			},
			gatewayClassProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					MergeType: new(egv1a1.StrategicMerge),
					Logging: egv1a1.ProxyLogging{
						Level: map[egv1a1.ProxyLogComponent]egv1a1.LogLevel{
							egv1a1.LogComponentAdmin: egv1a1.LogLevelDebug,
						},
					},
				},
			},
			expectedSpec: &egv1a1.EnvoyProxySpec{
				MergeType:   new(egv1a1.StrategicMerge),
				Concurrency: new(int32(4)),
				Logging: egv1a1.ProxyLogging{
					Level: map[egv1a1.ProxyLogComponent]egv1a1.LogLevel{
						egv1a1.LogComponentDefault: egv1a1.LogLevelInfo,
						egv1a1.LogComponentAdmin:   egv1a1.LogLevelDebug,
					},
				},
			},
		},
		{
			name: "defaultSpec mergeType has no effect on merge strategy",
			defaultSpec: &egv1a1.EnvoyProxySpec{
				MergeType:   new(egv1a1.StrategicMerge),
				Concurrency: new(int32(4)),
				Logging: egv1a1.ProxyLogging{
					Level: map[egv1a1.ProxyLogComponent]egv1a1.LogLevel{
						egv1a1.LogComponentDefault: egv1a1.LogLevelInfo,
					},
				},
			},
			gatewayClassProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					Concurrency: new(int32(8)),
				},
			},
			expectedSpec: &egv1a1.EnvoyProxySpec{
				Concurrency: new(int32(8)),
			},
		},
		{
			name: "StrategicMerge with backendRefs",
			defaultSpec: &egv1a1.EnvoyProxySpec{
				MergeType:   new(egv1a1.StrategicMerge),
				Concurrency: new(int32(4)),
				Logging: egv1a1.ProxyLogging{
					Level: map[egv1a1.ProxyLogComponent]egv1a1.LogLevel{
						egv1a1.LogComponentDefault: egv1a1.LogLevelInfo,
					},
				},
				Telemetry: &egv1a1.ProxyTelemetry{
					Tracing: &egv1a1.ProxyTracing{
						Provider: egv1a1.TracingProvider{
							BackendRefs: []egv1a1.BackendRef{
								{
									Group: new(gwapiv1.Group("gateway.envoyproxy.io")),
									Name:  gwapiv1.ObjectName("otel-collector1"),
								},
							},
						},
					},
				},
			},
			gatewayClassProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					MergeType:   new(egv1a1.StrategicMerge),
					Concurrency: new(int32(8)),
					Telemetry: &egv1a1.ProxyTelemetry{
						Tracing: &egv1a1.ProxyTracing{
							Provider: egv1a1.TracingProvider{
								ServiceName: new("fake-service-name"),
								BackendRefs: []egv1a1.BackendRef{
									{
										Kind: new(gwapiv1.Kind("Backend")),
										Name: gwapiv1.ObjectName("otel-collector"),
									},
								},
							},
						},
					},
				},
			},
			expectedSpec: &egv1a1.EnvoyProxySpec{
				Concurrency: new(int32(8)),
				MergeType:   new(egv1a1.StrategicMerge),
				Logging: egv1a1.ProxyLogging{
					Level: map[egv1a1.ProxyLogComponent]egv1a1.LogLevel{
						egv1a1.LogComponentDefault: egv1a1.LogLevelInfo,
					},
				},
				Telemetry: &egv1a1.ProxyTelemetry{
					Tracing: &egv1a1.ProxyTracing{
						Provider: egv1a1.TracingProvider{
							ServiceName: new("fake-service-name"),
							BackendRefs: []egv1a1.BackendRef{
								{
									Kind: new(gwapiv1.Kind("Backend")),
									Name: gwapiv1.ObjectName("otel-collector"),
								},
							},
						},
					},
				},
			},
		},
		{
			name: "JSONMerge with backendRefs",
			defaultSpec: &egv1a1.EnvoyProxySpec{
				Concurrency: new(int32(4)),
				Logging: egv1a1.ProxyLogging{
					Level: map[egv1a1.ProxyLogComponent]egv1a1.LogLevel{
						egv1a1.LogComponentDefault: egv1a1.LogLevelInfo,
					},
				},
				Telemetry: &egv1a1.ProxyTelemetry{
					Tracing: &egv1a1.ProxyTracing{
						Provider: egv1a1.TracingProvider{
							BackendRefs: []egv1a1.BackendRef{
								{
									Group: new(gwapiv1.Group("gateway.envoyproxy.io")),
									Name:  gwapiv1.ObjectName("otel-collector1"),
								},
							},
						},
					},
				},
			},
			gatewayClassProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					MergeType:   new(egv1a1.JSONMerge),
					Concurrency: new(int32(8)),
					Telemetry: &egv1a1.ProxyTelemetry{
						Tracing: &egv1a1.ProxyTracing{
							Provider: egv1a1.TracingProvider{
								ServiceName: new("fake-service-name"),
								BackendRefs: []egv1a1.BackendRef{
									{
										Kind: new(gwapiv1.Kind("Backend")),
										Name: gwapiv1.ObjectName("otel-collector"),
									},
								},
							},
						},
					},
				},
			},
			expectedSpec: &egv1a1.EnvoyProxySpec{
				Concurrency: new(int32(8)),
				MergeType:   new(egv1a1.JSONMerge),
				Logging: egv1a1.ProxyLogging{
					Level: map[egv1a1.ProxyLogComponent]egv1a1.LogLevel{
						egv1a1.LogComponentDefault: egv1a1.LogLevelInfo,
					},
				},
				Telemetry: &egv1a1.ProxyTelemetry{
					Tracing: &egv1a1.ProxyTracing{
						Provider: egv1a1.TracingProvider{
							ServiceName: new("fake-service-name"),
							BackendRefs: []egv1a1.BackendRef{
								{
									Kind: new(gwapiv1.Kind("Backend")),
									Name: gwapiv1.ObjectName("otel-collector"),
								},
							},
						},
					},
				},
			},
		},
		{
			name: "gatewayclass StrategicMerge daemonset - converts default deployment and drops its patch and hpa",
			defaultSpec: &egv1a1.EnvoyProxySpec{
				Provider: &egv1a1.EnvoyProxyProvider{
					Type: egv1a1.EnvoyProxyProviderTypeKubernetes,
					Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
						EnvoyDeployment: &egv1a1.KubernetesDeploymentSpec{
							Replicas:  new(int32(2)),
							Container: &egv1a1.KubernetesContainerSpec{Image: new("mirror.example/envoy:dev")},
							Patch: &egv1a1.KubernetesPatchSpec{
								Value: apiextensionsv1.JSON{Raw: []byte(`{"spec":{"replicas":3}}`)},
							},
						},
						EnvoyHpa: &egv1a1.KubernetesHorizontalPodAutoscalerSpec{MaxReplicas: new(int32(5))},
					},
				},
			},
			gatewayClassProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					MergeType: new(egv1a1.StrategicMerge),
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.EnvoyProxyProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyDaemonSet: &egv1a1.KubernetesDaemonSetSpec{
								Pod: &egv1a1.KubernetesPodSpec{NodeSelector: map[string]string{"role": "edge"}},
							},
						},
					},
				},
			},
			expectedSpec: &egv1a1.EnvoyProxySpec{
				MergeType: new(egv1a1.StrategicMerge),
				Provider: &egv1a1.EnvoyProxyProvider{
					Type: egv1a1.EnvoyProxyProviderTypeKubernetes,
					Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
						EnvoyDaemonSet: &egv1a1.KubernetesDaemonSetSpec{
							Pod:       &egv1a1.KubernetesPodSpec{NodeSelector: map[string]string{"role": "edge"}},
							Container: &egv1a1.KubernetesContainerSpec{Image: new("mirror.example/envoy:dev")},
						},
					},
				},
			},
		},
		{
			name: "gatewayclass JSONMerge daemonset - converts default deployment",
			defaultSpec: &egv1a1.EnvoyProxySpec{
				Provider: &egv1a1.EnvoyProxyProvider{
					Type: egv1a1.EnvoyProxyProviderTypeKubernetes,
					Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
						EnvoyDeployment: &egv1a1.KubernetesDeploymentSpec{
							Container: &egv1a1.KubernetesContainerSpec{Image: new("mirror.example/envoy:dev")},
						},
					},
				},
			},
			gatewayClassProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					MergeType: new(egv1a1.JSONMerge),
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.EnvoyProxyProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyDaemonSet: &egv1a1.KubernetesDaemonSetSpec{},
						},
					},
				},
			},
			expectedSpec: &egv1a1.EnvoyProxySpec{
				MergeType: new(egv1a1.JSONMerge),
				Provider: &egv1a1.EnvoyProxyProvider{
					Type: egv1a1.EnvoyProxyProviderTypeKubernetes,
					Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
						EnvoyDaemonSet: &egv1a1.KubernetesDaemonSetSpec{
							Container: &egv1a1.KubernetesContainerSpec{Image: new("mirror.example/envoy:dev")},
						},
					},
				},
			},
		},
		{
			name: "gateway StrategicMerge deployment - converts default daemonset",
			defaultSpec: &egv1a1.EnvoyProxySpec{
				Provider: &egv1a1.EnvoyProxyProvider{
					Type: egv1a1.EnvoyProxyProviderTypeKubernetes,
					Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
						EnvoyDaemonSet: &egv1a1.KubernetesDaemonSetSpec{
							Container: &egv1a1.KubernetesContainerSpec{Image: new("mirror.example/envoy:dev")},
						},
					},
				},
			},
			gatewayProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					MergeType: new(egv1a1.StrategicMerge),
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.EnvoyProxyProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyDeployment: &egv1a1.KubernetesDeploymentSpec{Replicas: new(int32(3))},
						},
					},
				},
			},
			expectedSpec: &egv1a1.EnvoyProxySpec{
				MergeType: new(egv1a1.StrategicMerge),
				Provider: &egv1a1.EnvoyProxyProvider{
					Type: egv1a1.EnvoyProxyProviderTypeKubernetes,
					Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
						EnvoyDeployment: &egv1a1.KubernetesDeploymentSpec{
							Replicas:  new(int32(3)),
							Container: &egv1a1.KubernetesContainerSpec{Image: new("mirror.example/envoy:dev")},
						},
					},
				},
			},
		},
		{
			name: "gateway StrategicMerge daemonset - converts gatewayclass deployment",
			gatewayClassProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.EnvoyProxyProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyDeployment: &egv1a1.KubernetesDeploymentSpec{
								Name:      new("edge-proxy"),
								Container: &egv1a1.KubernetesContainerSpec{Image: new("mirror.example/envoy:dev")},
							},
						},
					},
				},
			},
			gatewayProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					MergeType: new(egv1a1.StrategicMerge),
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.EnvoyProxyProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyDaemonSet: &egv1a1.KubernetesDaemonSetSpec{},
						},
					},
				},
			},
			expectedSpec: &egv1a1.EnvoyProxySpec{
				MergeType: new(egv1a1.StrategicMerge),
				Provider: &egv1a1.EnvoyProxyProvider{
					Type: egv1a1.EnvoyProxyProviderTypeKubernetes,
					Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
						EnvoyDaemonSet: &egv1a1.KubernetesDaemonSetSpec{
							Name:      new("edge-proxy"),
							Container: &egv1a1.KubernetesContainerSpec{Image: new("mirror.example/envoy:dev")},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := MergeEnvoyProxyConfigs(tc.defaultSpec, tc.gatewayClassProxy, tc.gatewayProxy)

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tc.expectedSpec == nil {
				require.Nil(t, result)
				return
			}

			diff := cmp.Diff(result.Spec, *tc.expectedSpec)
			require.Empty(t, diff)
		})
	}
}

func TestMergeEnvoyProxyConfigsWorkloadKindDoesNotMutateDefault(t *testing.T) {
	defaultSpec := &egv1a1.EnvoyProxySpec{
		Provider: &egv1a1.EnvoyProxyProvider{
			Type: egv1a1.EnvoyProxyProviderTypeKubernetes,
			Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
				EnvoyDeployment: &egv1a1.KubernetesDeploymentSpec{Replicas: new(int32(2))},
				EnvoyHpa:        &egv1a1.KubernetesHorizontalPodAutoscalerSpec{MaxReplicas: new(int32(5))},
			},
		},
	}
	original := defaultSpec.DeepCopy()
	gatewayClassProxy := &egv1a1.EnvoyProxy{
		Spec: egv1a1.EnvoyProxySpec{
			MergeType: new(egv1a1.StrategicMerge),
			Provider: &egv1a1.EnvoyProxyProvider{
				Type: egv1a1.EnvoyProxyProviderTypeKubernetes,
				Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
					EnvoyDaemonSet: &egv1a1.KubernetesDaemonSetSpec{},
				},
			},
		},
	}

	_, err := MergeEnvoyProxyConfigs(defaultSpec, gatewayClassProxy, nil)
	require.NoError(t, err)
	require.Empty(t, cmp.Diff(original, defaultSpec))
}
