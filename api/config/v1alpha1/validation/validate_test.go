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

	egcfgv1a1 "github.com/envoyproxy/gateway/api/config/v1alpha1"
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
		proxy    *egcfgv1a1.EnvoyProxy
		expected bool
	}{
		{
			name:     "nil egcfgv1a1.EnvoyProxy",
			proxy:    nil,
			expected: false,
		},
		{
			name: "nil provider",
			proxy: &egcfgv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egcfgv1a1.EnvoyProxySpec{
					Provider: nil,
				},
			},
			expected: true,
		},
		{
			name: "unsupported provider",
			proxy: &egcfgv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egcfgv1a1.EnvoyProxySpec{
					Provider: &egcfgv1a1.EnvoyProxyProvider{
						Type: egcfgv1a1.ProviderTypeFile,
					},
				},
			},
			expected: false,
		},
		{
			name: "nil envoy service",
			proxy: &egcfgv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egcfgv1a1.EnvoyProxySpec{
					Provider: &egcfgv1a1.EnvoyProxyProvider{
						Type: egcfgv1a1.ProviderTypeKubernetes,
						Kubernetes: &egcfgv1a1.EnvoyProxyKubernetesProvider{
							EnvoyService: nil,
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "unsupported envoy service type \"\" ",
			proxy: &egcfgv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egcfgv1a1.EnvoyProxySpec{
					Provider: &egcfgv1a1.EnvoyProxyProvider{
						Type: egcfgv1a1.ProviderTypeKubernetes,
						Kubernetes: &egcfgv1a1.EnvoyProxyKubernetesProvider{
							EnvoyService: &egcfgv1a1.KubernetesServiceSpec{
								Type: egcfgv1a1.GetKubernetesServiceType(""),
							},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "unsupported envoy service type 'NodePort'",
			proxy: &egcfgv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egcfgv1a1.EnvoyProxySpec{
					Provider: &egcfgv1a1.EnvoyProxyProvider{
						Type: egcfgv1a1.ProviderTypeKubernetes,
						Kubernetes: &egcfgv1a1.EnvoyProxyKubernetesProvider{
							EnvoyService: &egcfgv1a1.KubernetesServiceSpec{
								Type: egcfgv1a1.GetKubernetesServiceType(egcfgv1a1.ServiceType(corev1.ServiceTypeNodePort)),
							},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "valid envoy service type 'LoadBalancer'",
			proxy: &egcfgv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egcfgv1a1.EnvoyProxySpec{
					Provider: &egcfgv1a1.EnvoyProxyProvider{
						Type: egcfgv1a1.ProviderTypeKubernetes,
						Kubernetes: &egcfgv1a1.EnvoyProxyKubernetesProvider{
							EnvoyService: &egcfgv1a1.KubernetesServiceSpec{
								Type: egcfgv1a1.GetKubernetesServiceType(egcfgv1a1.ServiceTypeLoadBalancer),
							},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "valid envoy service type 'ClusterIP'",
			proxy: &egcfgv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egcfgv1a1.EnvoyProxySpec{
					Provider: &egcfgv1a1.EnvoyProxyProvider{
						Type: egcfgv1a1.ProviderTypeKubernetes,
						Kubernetes: &egcfgv1a1.EnvoyProxyKubernetesProvider{
							EnvoyService: &egcfgv1a1.KubernetesServiceSpec{
								Type: egcfgv1a1.GetKubernetesServiceType(egcfgv1a1.ServiceTypeClusterIP),
							},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "valid user bootstrap replace type",
			proxy: &egcfgv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egcfgv1a1.EnvoyProxySpec{
					Bootstrap: &egcfgv1a1.ProxyBootstrap{
						Value: validUserBootstrap,
					},
				},
			},
			expected: true,
		},
		{
			name: "valid user bootstrap merge type",
			proxy: &egcfgv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egcfgv1a1.EnvoyProxySpec{
					Bootstrap: &egcfgv1a1.ProxyBootstrap{
						Type:  (*egcfgv1a1.BootstrapType)(pointer.String(string(egcfgv1a1.BootstrapTypeMerge))),
						Value: mergeUserBootstrap,
					},
				},
			},
			expected: true,
		},
		{
			name: "user bootstrap with missing admin address",
			proxy: &egcfgv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egcfgv1a1.EnvoyProxySpec{
					Bootstrap: &egcfgv1a1.ProxyBootstrap{
						Value: missingAdminAddressUserBootstrap,
					},
				},
			},
			expected: false,
		},
		{
			name: "user bootstrap with different dynamic resources",
			proxy: &egcfgv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egcfgv1a1.EnvoyProxySpec{
					Bootstrap: &egcfgv1a1.ProxyBootstrap{
						Value: differentDynamicResourcesUserBootstrap,
					},
				},
			},
			expected: false,
		},
		{
			name: "user bootstrap with different xds_cluster endpoint",
			proxy: &egcfgv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egcfgv1a1.EnvoyProxySpec{
					Bootstrap: &egcfgv1a1.ProxyBootstrap{
						Value: differentXdsClusterAddressBootstrap,
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

func TestEnvoyGateway(t *testing.T) {
	envoyGateway := egcfgv1a1.DefaultEnvoyGateway()
	assert.True(t, envoyGateway.Provider != nil)
	assert.True(t, envoyGateway.Gateway != nil)
	assert.True(t, envoyGateway.Logging != nil)
	envoyGateway.SetEnvoyGatewayDefaults()
	assert.Equal(t, envoyGateway.Logging, egcfgv1a1.DefaultEnvoyGatewayLogging())

	logging := egcfgv1a1.DefaultEnvoyGatewayLogging()
	assert.True(t, logging != nil)
	assert.True(t, logging.Level[egcfgv1a1.LogComponentGatewayDefault] == egcfgv1a1.LogLevelInfo)

	gatewayLogging := &egcfgv1a1.EnvoyGatewayLogging{
		Level: logging.Level,
	}
	gatewayLogging.SetEnvoyGatewayLoggingDefaults()
	assert.True(t, gatewayLogging != nil)
	assert.True(t, gatewayLogging.Level[egcfgv1a1.LogComponentGatewayDefault] == egcfgv1a1.LogLevelInfo)
}

func TestDefaultEnvoyGatewayLoggingLevel(t *testing.T) {
	type args struct {
		component string
		level     egcfgv1a1.LogLevel
	}
	tests := []struct {
		name string
		args args
		want egcfgv1a1.LogLevel
	}{
		{
			name: "test default info level for empty level",
			args: args{component: "", level: ""},
			want: egcfgv1a1.LogLevelInfo,
		},
		{
			name: "test default info level for empty level",
			args: args{component: string(egcfgv1a1.LogComponentGatewayDefault), level: ""},
			want: egcfgv1a1.LogLevelInfo,
		},
		{
			name: "test default info level for info level",
			args: args{component: string(egcfgv1a1.LogComponentGatewayDefault), level: egcfgv1a1.LogLevelInfo},
			want: egcfgv1a1.LogLevelInfo,
		},
		{
			name: "test default error level for error level",
			args: args{component: string(egcfgv1a1.LogComponentGatewayDefault), level: egcfgv1a1.LogLevelError},
			want: egcfgv1a1.LogLevelError,
		},
		{
			name: "test gateway-api error level for error level",
			args: args{component: string(egcfgv1a1.LogComponentGatewayAPIRunner), level: egcfgv1a1.LogLevelError},
			want: egcfgv1a1.LogLevelError,
		},
		{
			name: "test gateway-api info level for info level",
			args: args{component: string(egcfgv1a1.LogComponentGatewayAPIRunner), level: egcfgv1a1.LogLevelInfo},
			want: egcfgv1a1.LogLevelInfo,
		},
		{
			name: "test default gateway-api warn level for info level",
			args: args{component: string(egcfgv1a1.LogComponentGatewayAPIRunner), level: ""},
			want: egcfgv1a1.LogLevelInfo,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logging := &egcfgv1a1.EnvoyGatewayLogging{}
			if got := logging.DefaultEnvoyGatewayLoggingLevel(tt.args.level); got != tt.want {
				t.Errorf("defaultLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnvoyGatewayProvider(t *testing.T) {
	envoyGateway := &egcfgv1a1.EnvoyGateway{
		TypeMeta:         metav1.TypeMeta{},
		EnvoyGatewaySpec: egcfgv1a1.EnvoyGatewaySpec{Provider: egcfgv1a1.DefaultEnvoyGatewayProvider()},
	}
	assert.True(t, envoyGateway.Provider != nil)

	envoyGatewayProvider := envoyGateway.GetEnvoyGatewayProvider()
	assert.True(t, envoyGatewayProvider.Kubernetes == nil)
	assert.Equal(t, envoyGateway.Provider, envoyGatewayProvider)

	envoyGatewayProvider.Kubernetes = egcfgv1a1.DefaultEnvoyGatewayKubeProvider()
	assert.Equal(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment, egcfgv1a1.DefaultKubernetesDeployment(egcfgv1a1.DefaultRateLimitImage))

	envoyGatewayProvider.Kubernetes = &egcfgv1a1.EnvoyGatewayKubernetesProvider{}
	assert.True(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment == nil)

	envoyGatewayProvider.Kubernetes = &egcfgv1a1.EnvoyGatewayKubernetesProvider{
		RateLimitDeployment: &egcfgv1a1.KubernetesDeploymentSpec{
			Replicas:  nil,
			Pod:       nil,
			Container: nil,
		}}
	assert.True(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Replicas == nil)
	assert.True(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Pod == nil)
	assert.True(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Container == nil)
	envoyGatewayKubeProvider := envoyGatewayProvider.GetEnvoyGatewayKubeProvider()

	envoyGatewayProvider.Kubernetes = &egcfgv1a1.EnvoyGatewayKubernetesProvider{
		RateLimitDeployment: &egcfgv1a1.KubernetesDeploymentSpec{
			Replicas: nil,
			Pod:      nil,
			Container: &egcfgv1a1.KubernetesContainerSpec{
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
	assert.Equal(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment, egcfgv1a1.DefaultKubernetesDeployment(egcfgv1a1.DefaultRateLimitImage))
	assert.True(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Replicas != nil)
	assert.Equal(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Replicas, egcfgv1a1.DefaultKubernetesDeploymentReplicas())
	assert.True(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Pod != nil)
	assert.Equal(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Pod, egcfgv1a1.DefaultKubernetesPod())
	assert.True(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Container != nil)
	assert.Equal(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Container, egcfgv1a1.DefaultKubernetesContainer(egcfgv1a1.DefaultRateLimitImage))
	assert.True(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Container.Resources != nil)
	assert.Equal(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Container.Resources, egcfgv1a1.DefaultResourceRequirements())
	assert.True(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Container.Image != nil)
	assert.Equal(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Container.Image, egcfgv1a1.DefaultKubernetesContainerImage(egcfgv1a1.DefaultRateLimitImage))
}

func TestEnvoyProxyProvider(t *testing.T) {
	envoyProxy := &egcfgv1a1.EnvoyProxy{
		Spec: egcfgv1a1.EnvoyProxySpec{
			Provider: egcfgv1a1.DefaultEnvoyProxyProvider(),
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
	assert.Equal(t, envoyProxyProvider.Kubernetes.EnvoyDeployment, egcfgv1a1.DefaultKubernetesDeployment(egcfgv1a1.DefaultEnvoyProxyImage))
	assert.True(t, envoyProxyProvider.Kubernetes.EnvoyDeployment.Replicas != nil)
	assert.Equal(t, envoyProxyProvider.Kubernetes.EnvoyDeployment.Replicas, egcfgv1a1.DefaultKubernetesDeploymentReplicas())
	assert.True(t, envoyProxyProvider.Kubernetes.EnvoyDeployment.Pod != nil)
	assert.Equal(t, envoyProxyProvider.Kubernetes.EnvoyDeployment.Pod, egcfgv1a1.DefaultKubernetesPod())
	assert.True(t, envoyProxyProvider.Kubernetes.EnvoyDeployment.Container != nil)
	assert.Equal(t, envoyProxyProvider.Kubernetes.EnvoyDeployment.Container, egcfgv1a1.DefaultKubernetesContainer(egcfgv1a1.DefaultEnvoyProxyImage))
	assert.True(t, envoyProxyProvider.Kubernetes.EnvoyDeployment.Container.Resources != nil)
	assert.Equal(t, envoyProxyProvider.Kubernetes.EnvoyDeployment.Container.Resources, egcfgv1a1.DefaultResourceRequirements())
	assert.True(t, envoyProxyProvider.Kubernetes.EnvoyDeployment.Container.Image != nil)
	assert.Equal(t, envoyProxyProvider.Kubernetes.EnvoyDeployment.Container.Image, egcfgv1a1.DefaultKubernetesContainerImage(egcfgv1a1.DefaultEnvoyProxyImage))

	assert.True(t, envoyProxyProvider.Kubernetes.EnvoyService != nil)
	assert.True(t, reflect.DeepEqual(envoyProxyProvider.Kubernetes.EnvoyService.Type, egcfgv1a1.GetKubernetesServiceType(egcfgv1a1.ServiceTypeLoadBalancer)))
}

func TestEnvoyGatewayAdmin(t *testing.T) {
	// default envoygateway config admin should not be nil
	eg := egcfgv1a1.DefaultEnvoyGateway()
	assert.True(t, eg.Admin != nil)

	// get default admin config from envoygateway
	// values should be set in default
	egAdmin := eg.GetEnvoyGatewayAdmin()
	assert.True(t, egAdmin != nil)
	assert.True(t, egAdmin.Debug == false)
	assert.True(t, egAdmin.Address.Port == egcfgv1a1.GatewayAdminPort)
	assert.True(t, egAdmin.Address.Host == egcfgv1a1.GatewayAdminHost)

	// override the admin config
	// values should be updated
	eg.Admin.Debug = true
	eg.Admin.Address = nil
	assert.True(t, eg.Admin.Debug == true)
	assert.True(t, eg.GetEnvoyGatewayAdmin().Address.Port == egcfgv1a1.GatewayAdminPort)
	assert.True(t, eg.GetEnvoyGatewayAdmin().Address.Host == egcfgv1a1.GatewayAdminHost)

	// set eg defaults when admin is nil
	// the admin should not be nil
	eg.Admin = nil
	eg.SetEnvoyGatewayDefaults()
	assert.True(t, eg.Admin != nil)
	assert.True(t, eg.Admin.Debug == false)
	assert.True(t, eg.Admin.Address.Port == egcfgv1a1.GatewayAdminPort)
	assert.True(t, eg.Admin.Address.Host == egcfgv1a1.GatewayAdminHost)
}

func TestGetEnvoyProxyDefaultComponentLevel(t *testing.T) {
	cases := []struct {
		logging   egcfgv1a1.ProxyLogging
		component egcfgv1a1.LogComponent
		expected  egcfgv1a1.LogLevel
	}{
		{
			logging: egcfgv1a1.ProxyLogging{
				Level: map[egcfgv1a1.LogComponent]egcfgv1a1.LogLevel{
					egcfgv1a1.LogComponentDefault: egcfgv1a1.LogLevelInfo,
				},
			},
			expected: egcfgv1a1.LogLevelInfo,
		},
		{
			logging: egcfgv1a1.ProxyLogging{
				Level: map[egcfgv1a1.LogComponent]egcfgv1a1.LogLevel{
					egcfgv1a1.LogComponentDefault: egcfgv1a1.LogLevelInfo,
				},
			},
			expected: egcfgv1a1.LogLevelInfo,
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
		logging  egcfgv1a1.ProxyLogging
		expected string
	}{
		{
			logging:  egcfgv1a1.ProxyLogging{},
			expected: "",
		},
		{
			logging: egcfgv1a1.ProxyLogging{
				Level: map[egcfgv1a1.LogComponent]egcfgv1a1.LogLevel{
					egcfgv1a1.LogComponentDefault: egcfgv1a1.LogLevelInfo,
				},
			},
			expected: "",
		},
		{
			logging: egcfgv1a1.ProxyLogging{
				Level: map[egcfgv1a1.LogComponent]egcfgv1a1.LogLevel{
					egcfgv1a1.LogComponentDefault: egcfgv1a1.LogLevelInfo,
					egcfgv1a1.LogComponentAdmin:   egcfgv1a1.LogLevelWarn,
				},
			},
			expected: "admin:warn",
		},
		{
			logging: egcfgv1a1.ProxyLogging{
				Level: map[egcfgv1a1.LogComponent]egcfgv1a1.LogLevel{
					egcfgv1a1.LogComponentDefault: egcfgv1a1.LogLevelInfo,
					egcfgv1a1.LogComponentAdmin:   egcfgv1a1.LogLevelWarn,
					egcfgv1a1.LogComponentFilter:  egcfgv1a1.LogLevelDebug,
				},
			},
			expected: "admin:warn,filter:debug",
		},
		{
			logging: egcfgv1a1.ProxyLogging{
				Level: map[egcfgv1a1.LogComponent]egcfgv1a1.LogLevel{
					egcfgv1a1.LogComponentDefault: egcfgv1a1.LogLevelInfo,
					egcfgv1a1.LogComponentAdmin:   egcfgv1a1.LogLevelWarn,
					egcfgv1a1.LogComponentFilter:  egcfgv1a1.LogLevelDebug,
					egcfgv1a1.LogComponentClient:  "",
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
