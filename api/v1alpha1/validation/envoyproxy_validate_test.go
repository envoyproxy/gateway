// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package validation

import (
	// Register embed
	_ "embed"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/utils/ptr"
)

var (
	//go:embed testdata/valid-user-bootstrap.yaml
	validUserBootstrap string
	//go:embed testdata/merge-user-bootstrap.yaml
	mergeUserBootstrap string
	//go:embed testdata/missing-admin-address-user-bootstrap.yaml
	missingAdminAddressUserBootstrap string
	//go:embed testdata/different-dynamic-resources-user-bootstrap.yaml
	differentDynamicResourcesUserBootstrap string
	//go:embed testdata/different-xds-cluster-address-bootstrap.yaml
	differentXdsClusterAddressBootstrap string
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
						Type: egv1a1.ProviderTypeFile,
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
			name: "valid user bootstrap replace type",
			proxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Bootstrap: &egv1a1.ProxyBootstrap{
						Value: validUserBootstrap,
					},
				},
			},
			expected: true,
		},
		{
			name: "valid user bootstrap merge type",
			proxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Bootstrap: &egv1a1.ProxyBootstrap{
						Type:  (*egv1a1.BootstrapType)(pointer.String(string(egv1a1.BootstrapTypeMerge))),
						Value: mergeUserBootstrap,
					},
				},
			},
			expected: true,
		},
		{
			name: "user bootstrap with missing admin address",
			proxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Bootstrap: &egv1a1.ProxyBootstrap{
						Value: missingAdminAddressUserBootstrap,
					},
				},
			},
			expected: false,
		},
		{
			name: "user bootstrap with different dynamic resources",
			proxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Bootstrap: &egv1a1.ProxyBootstrap{
						Value: differentDynamicResourcesUserBootstrap,
					},
				},
			},
			expected: false,
		},
		{
			name: "user bootstrap with different xds_cluster endpoint",
			proxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Bootstrap: &egv1a1.ProxyBootstrap{
						Value: differentXdsClusterAddressBootstrap,
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
									Format: egv1a1.ProxyAccessLogFormat{
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
									Format: egv1a1.ProxyAccessLogFormat{
										Type: egv1a1.ProxyAccessLogFormatTypeText,
										Text: pointer.String("[%START_TIME%]"),
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
		}, {
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
		}, {
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
										Host: "0.0.0.0",
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

func TestEnvoyGateway(t *testing.T) {
	envoyGateway := egv1a1.DefaultEnvoyGateway()
	assert.True(t, envoyGateway.Provider != nil)
	assert.True(t, envoyGateway.Gateway != nil)
	assert.True(t, envoyGateway.Logging != nil)
	envoyGateway.SetEnvoyGatewayDefaults()
	assert.Equal(t, envoyGateway.Logging, egv1a1.DefaultEnvoyGatewayLogging())

	logging := egv1a1.DefaultEnvoyGatewayLogging()
	assert.True(t, logging != nil)
	assert.True(t, logging.Level[egv1a1.LogComponentGatewayDefault] == egv1a1.LogLevelInfo)

	gatewayLogging := &egv1a1.EnvoyGatewayLogging{
		Level: logging.Level,
	}
	gatewayLogging.SetEnvoyGatewayLoggingDefaults()
	assert.True(t, gatewayLogging != nil)
	assert.True(t, gatewayLogging.Level[egv1a1.LogComponentGatewayDefault] == egv1a1.LogLevelInfo)
}

func TestDefaultEnvoyGatewayLoggingLevel(t *testing.T) {
	type args struct {
		component string
		level     egv1a1.LogLevel
	}
	tests := []struct {
		name string
		args args
		want egv1a1.LogLevel
	}{
		{
			name: "test default info level for empty level",
			args: args{component: "", level: ""},
			want: egv1a1.LogLevelInfo,
		},
		{
			name: "test default info level for empty level",
			args: args{component: string(egv1a1.LogComponentGatewayDefault), level: ""},
			want: egv1a1.LogLevelInfo,
		},
		{
			name: "test default info level for info level",
			args: args{component: string(egv1a1.LogComponentGatewayDefault), level: egv1a1.LogLevelInfo},
			want: egv1a1.LogLevelInfo,
		},
		{
			name: "test default error level for error level",
			args: args{component: string(egv1a1.LogComponentGatewayDefault), level: egv1a1.LogLevelError},
			want: egv1a1.LogLevelError,
		},
		{
			name: "test gateway-api error level for error level",
			args: args{component: string(egv1a1.LogComponentGatewayAPIRunner), level: egv1a1.LogLevelError},
			want: egv1a1.LogLevelError,
		},
		{
			name: "test gateway-api info level for info level",
			args: args{component: string(egv1a1.LogComponentGatewayAPIRunner), level: egv1a1.LogLevelInfo},
			want: egv1a1.LogLevelInfo,
		},
		{
			name: "test default gateway-api warn level for info level",
			args: args{component: string(egv1a1.LogComponentGatewayAPIRunner), level: ""},
			want: egv1a1.LogLevelInfo,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logging := &egv1a1.EnvoyGatewayLogging{}
			if got := logging.DefaultEnvoyGatewayLoggingLevel(tt.args.level); got != tt.want {
				t.Errorf("defaultLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnvoyGatewayProvider(t *testing.T) {
	envoyGateway := &egv1a1.EnvoyGateway{
		TypeMeta:         metav1.TypeMeta{},
		EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{Provider: egv1a1.DefaultEnvoyGatewayProvider()},
	}
	assert.True(t, envoyGateway.Provider != nil)

	envoyGatewayProvider := envoyGateway.GetEnvoyGatewayProvider()
	assert.True(t, envoyGatewayProvider.Kubernetes == nil)
	assert.Equal(t, envoyGateway.Provider, envoyGatewayProvider)

	envoyGatewayProvider.Kubernetes = egv1a1.DefaultEnvoyGatewayKubeProvider()
	assert.Equal(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment, egv1a1.DefaultKubernetesDeployment(egv1a1.DefaultRateLimitImage))

	envoyGatewayProvider.Kubernetes = &egv1a1.EnvoyGatewayKubernetesProvider{}
	assert.True(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment == nil)

	envoyGatewayProvider.Kubernetes = &egv1a1.EnvoyGatewayKubernetesProvider{
		RateLimitDeployment: &egv1a1.KubernetesDeploymentSpec{
			Replicas:  nil,
			Pod:       nil,
			Container: nil,
		}}
	assert.True(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Replicas == nil)
	assert.True(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Pod == nil)
	assert.True(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Container == nil)
	envoyGatewayKubeProvider := envoyGatewayProvider.GetEnvoyGatewayKubeProvider()

	envoyGatewayProvider.Kubernetes = &egv1a1.EnvoyGatewayKubernetesProvider{
		RateLimitDeployment: &egv1a1.KubernetesDeploymentSpec{
			Replicas: nil,
			Pod:      nil,
			Container: &egv1a1.KubernetesContainerSpec{
				Resources:       nil,
				SecurityContext: nil,
				Image:           nil,
			},
		}}
	assert.True(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Container.Resources == nil)
	envoyGatewayProvider.GetEnvoyGatewayKubeProvider()

	assert.True(t, envoyGatewayProvider.Kubernetes != nil)
	assert.Equal(t, envoyGatewayProvider.Kubernetes, envoyGatewayKubeProvider)

	assert.True(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment != nil)
	assert.Equal(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment, egv1a1.DefaultKubernetesDeployment(egv1a1.DefaultRateLimitImage))
	assert.True(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Replicas != nil)
	assert.Equal(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Replicas, egv1a1.DefaultKubernetesDeploymentReplicas())
	assert.True(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Pod != nil)
	assert.Equal(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Pod, egv1a1.DefaultKubernetesPod())
	assert.True(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Container != nil)
	assert.Equal(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Container, egv1a1.DefaultKubernetesContainer(egv1a1.DefaultRateLimitImage))
	assert.True(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Container.Resources != nil)
	assert.Equal(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Container.Resources, egv1a1.DefaultResourceRequirements())
	assert.True(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Container.Image != nil)
	assert.Equal(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Container.Image, egv1a1.DefaultKubernetesContainerImage(egv1a1.DefaultRateLimitImage))
}

func TestEnvoyProxyProvider(t *testing.T) {
	envoyProxy := &egv1a1.EnvoyProxy{
		Spec: egv1a1.EnvoyProxySpec{
			Provider: egv1a1.DefaultEnvoyProxyProvider(),
		},
	}
	assert.True(t, envoyProxy.Spec.Provider != nil)

	envoyProxyProvider := envoyProxy.GetEnvoyProxyProvider()
	assert.True(t, envoyProxyProvider.Kubernetes == nil)
	assert.True(t, reflect.DeepEqual(envoyProxy.Spec.Provider, envoyProxyProvider))

	envoyProxyKubeProvider := envoyProxyProvider.GetEnvoyProxyKubeProvider()

	assert.True(t, envoyProxyProvider.Kubernetes != nil)
	assert.True(t, reflect.DeepEqual(envoyProxyProvider.Kubernetes, envoyProxyKubeProvider))

	envoyProxyProvider.GetEnvoyProxyKubeProvider()

	assert.True(t, envoyProxyProvider.Kubernetes.EnvoyDeployment != nil)
	assert.Equal(t, envoyProxyProvider.Kubernetes.EnvoyDeployment, egv1a1.DefaultKubernetesDeployment(egv1a1.DefaultEnvoyProxyImage))
	assert.True(t, envoyProxyProvider.Kubernetes.EnvoyDeployment.Replicas != nil)
	assert.Equal(t, envoyProxyProvider.Kubernetes.EnvoyDeployment.Replicas, egv1a1.DefaultKubernetesDeploymentReplicas())
	assert.True(t, envoyProxyProvider.Kubernetes.EnvoyDeployment.Pod != nil)
	assert.Equal(t, envoyProxyProvider.Kubernetes.EnvoyDeployment.Pod, egv1a1.DefaultKubernetesPod())
	assert.True(t, envoyProxyProvider.Kubernetes.EnvoyDeployment.Container != nil)
	assert.Equal(t, envoyProxyProvider.Kubernetes.EnvoyDeployment.Container, egv1a1.DefaultKubernetesContainer(egv1a1.DefaultEnvoyProxyImage))
	assert.True(t, envoyProxyProvider.Kubernetes.EnvoyDeployment.Container.Resources != nil)
	assert.Equal(t, envoyProxyProvider.Kubernetes.EnvoyDeployment.Container.Resources, egv1a1.DefaultResourceRequirements())
	assert.True(t, envoyProxyProvider.Kubernetes.EnvoyDeployment.Container.Image != nil)
	assert.Equal(t, envoyProxyProvider.Kubernetes.EnvoyDeployment.Container.Image, egv1a1.DefaultKubernetesContainerImage(egv1a1.DefaultEnvoyProxyImage))

	assert.True(t, envoyProxyProvider.Kubernetes.EnvoyService != nil)
	assert.True(t, reflect.DeepEqual(envoyProxyProvider.Kubernetes.EnvoyService.Type, egv1a1.GetKubernetesServiceType(egv1a1.ServiceTypeLoadBalancer)))
}

func TestEnvoyGatewayAdmin(t *testing.T) {
	// default envoygateway config admin should not be nil
	eg := egv1a1.DefaultEnvoyGateway()
	assert.True(t, eg.Admin != nil)

	// get default admin config from envoygateway
	// values should be set in default
	egAdmin := eg.GetEnvoyGatewayAdmin()
	assert.True(t, egAdmin != nil)
	assert.True(t, egAdmin.Address.Port == egv1a1.GatewayAdminPort)
	assert.True(t, egAdmin.Address.Host == egv1a1.GatewayAdminHost)
	assert.True(t, egAdmin.EnableDumpConfig == false)
	assert.True(t, egAdmin.EnablePprof == false)

	// override the admin config
	// values should be updated
	eg.Admin = &egv1a1.EnvoyGatewayAdmin{
		Address: &egv1a1.EnvoyGatewayAdminAddress{
			Host: "0.0.0.0",
			Port: 19010,
		},
		EnableDumpConfig: true,
		EnablePprof:      true,
	}

	assert.True(t, eg.GetEnvoyGatewayAdmin().Address.Port == 19010)
	assert.True(t, eg.GetEnvoyGatewayAdmin().Address.Host == "0.0.0.0")
	assert.True(t, eg.GetEnvoyGatewayAdmin().EnableDumpConfig == true)
	assert.True(t, eg.GetEnvoyGatewayAdmin().EnablePprof == true)

	// set eg defaults when admin is nil
	// the admin should not be nil
	eg.Admin = nil
	eg.SetEnvoyGatewayDefaults()
	assert.True(t, eg.Admin != nil)
	assert.True(t, eg.Admin.Address.Port == egv1a1.GatewayAdminPort)
	assert.True(t, eg.Admin.Address.Host == egv1a1.GatewayAdminHost)
	assert.True(t, eg.Admin.EnableDumpConfig == false)
	assert.True(t, eg.Admin.EnablePprof == false)
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
			expected: "",
		},
		{
			logging: egv1a1.ProxyLogging{
				Level: map[egv1a1.ProxyLogComponent]egv1a1.LogLevel{
					egv1a1.LogComponentDefault: egv1a1.LogLevelInfo,
				},
			},
			expected: "",
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
