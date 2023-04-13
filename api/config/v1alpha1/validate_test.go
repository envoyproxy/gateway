// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	// Register embed
	_ "embed"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	//go:embed testdata/valid-user-bootstrap.yaml
	validUserBootstrap string
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
		obj      *EnvoyProxy
		expected bool
	}{
		{
			name:     "nil envoyproxy",
			obj:      nil,
			expected: false,
		},
		{
			name: "nil provider",
			obj: &EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: EnvoyProxySpec{
					Provider: nil,
				},
			},
			expected: true,
		},
		{
			name: "unsupported provider",
			obj: &EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: EnvoyProxySpec{
					Provider: &EnvoyProxyProvider{
						Type: ProviderTypeFile,
					},
				},
			},
			expected: false,
		},
		{
			name: "nil envoy service",
			obj: &EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: EnvoyProxySpec{
					Provider: &EnvoyProxyProvider{
						Type: ProviderTypeKubernetes,
						Kubernetes: &EnvoyProxyKubernetesProvider{
							EnvoyService: nil,
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "unsupported envoy service type \"\" ",
			obj: &EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: EnvoyProxySpec{
					Provider: &EnvoyProxyProvider{
						Type: ProviderTypeKubernetes,
						Kubernetes: &EnvoyProxyKubernetesProvider{
							EnvoyService: &KubernetesServiceSpec{
								Type: GetKubernetesServiceType(""),
							},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "unsupported envoy service type 'NodePort'",
			obj: &EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: EnvoyProxySpec{
					Provider: &EnvoyProxyProvider{
						Type: ProviderTypeKubernetes,
						Kubernetes: &EnvoyProxyKubernetesProvider{
							EnvoyService: &KubernetesServiceSpec{
								Type: GetKubernetesServiceType(ServiceType(corev1.ServiceTypeNodePort)),
							},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "valid envoy service type 'LoadBalancer'",
			obj: &EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: EnvoyProxySpec{
					Provider: &EnvoyProxyProvider{
						Type: ProviderTypeKubernetes,
						Kubernetes: &EnvoyProxyKubernetesProvider{
							EnvoyService: &KubernetesServiceSpec{
								Type: GetKubernetesServiceType(ServiceTypeLoadBalancer),
							},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "valid envoy service type 'ClusterIP'",
			obj: &EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: EnvoyProxySpec{
					Provider: &EnvoyProxyProvider{
						Type: ProviderTypeKubernetes,
						Kubernetes: &EnvoyProxyKubernetesProvider{
							EnvoyService: &KubernetesServiceSpec{
								Type: GetKubernetesServiceType(ServiceTypeClusterIP),
							},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "valid user bootstrap",
			obj: &EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: EnvoyProxySpec{
					Bootstrap: &validUserBootstrap,
				},
			},
			expected: true,
		},
		{
			name: "user bootstrap with missing admin address",
			obj: &EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: EnvoyProxySpec{
					Bootstrap: &missingAdminAddressUserBootstrap,
				},
			},
			expected: false,
		},
		{
			name: "user bootstrap with different dynamic resources",
			obj: &EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: EnvoyProxySpec{
					Bootstrap: &differentDynamicResourcesUserBootstrap,
				},
			},
			expected: false,
		},
		{
			name: "user bootstrap with different xds_cluster endpoint",
			obj: &EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: EnvoyProxySpec{
					Bootstrap: &differentXdsClusterAddressBootstrap,
				},
			},
			expected: false,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			err := tc.obj.Validate()
			if tc.expected {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestEnvoyGatewayProvider(t *testing.T) {
	envoyGateway := &EnvoyGateway{
		TypeMeta:         metav1.TypeMeta{},
		EnvoyGatewaySpec: EnvoyGatewaySpec{Provider: DefaultEnvoyGatewayProvider()},
	}
	assert.True(t, envoyGateway.Provider != nil)

	envoyGatewayProvider := envoyGateway.GetEnvoyGatewayProvider()
	assert.True(t, envoyGatewayProvider.Kubernetes == nil)
	assert.Equal(t, envoyGateway.Provider, envoyGatewayProvider)

	envoyGatewayProvider.Kubernetes = DefaultEnvoyGatewayKubeProvider()
	assert.Equal(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment, DefaultKubernetesDeployment())

	envoyGatewayProvider.Kubernetes = &EnvoyGatewayKubernetesProvider{}
	assert.True(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment == nil)

	envoyGatewayProvider.Kubernetes = &EnvoyGatewayKubernetesProvider{RateLimitDeployment: &KubernetesDeploymentSpec{
		Replicas:  nil,
		Pod:       nil,
		Container: nil,
	}}
	assert.True(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Replicas == nil)
	assert.True(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Pod == nil)
	assert.True(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Container == nil)
	envoyGatewayKubeProvider := envoyGatewayProvider.GetEnvoyGatewayKubeProvider()

	envoyGatewayProvider.Kubernetes = &EnvoyGatewayKubernetesProvider{RateLimitDeployment: &KubernetesDeploymentSpec{
		Replicas: nil,
		Pod:      nil,
		Container: &KubernetesContainerSpec{
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
	assert.Equal(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment, DefaultKubernetesDeployment())
	assert.True(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Replicas != nil)
	assert.Equal(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Replicas, DefaultKubernetesDeploymentReplicas())
	assert.True(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Pod != nil)
	assert.Equal(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Pod, DefaultKubernetesPod())
	assert.True(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Container != nil)
	assert.Equal(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Container, DefaultKubernetesContainer())
	assert.True(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Container.Resources != nil)
	assert.Equal(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Container.Resources, DefaultResourceRequirements())
	assert.True(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Container.Image != nil)
	assert.Equal(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Container.Image, DefaultKubernetesContainerImage())
}

func TestEnvoyProxyProvider(t *testing.T) {
	envoyProxy := &EnvoyProxy{Spec: EnvoyProxySpec{Provider: DefaultEnvoyProxyProvider()}}
	assert.True(t, envoyProxy.Spec.Provider != nil)

	envoyProxyProvider := envoyProxy.GetEnvoyProxyProvider()
	assert.True(t, envoyProxyProvider.Kubernetes == nil)
	assert.True(t, reflect.DeepEqual(envoyProxy.Spec.Provider, envoyProxyProvider))

	envoyProxyKubeProvider := envoyProxyProvider.GetEnvoyProxyKubeProvider()

	assert.True(t, envoyProxyProvider.Kubernetes != nil)
	assert.True(t, reflect.DeepEqual(envoyProxyProvider.Kubernetes, envoyProxyKubeProvider))

	envoyProxyProvider.GetEnvoyProxyKubeProvider()

	assert.True(t, envoyProxyProvider.Kubernetes.EnvoyDeployment != nil)
	assert.True(t, reflect.DeepEqual(envoyProxyProvider.Kubernetes.EnvoyDeployment, DefaultKubernetesDeployment()))
	assert.True(t, envoyProxyProvider.Kubernetes.EnvoyDeployment.Replicas != nil)
	assert.True(t, reflect.DeepEqual(envoyProxyProvider.Kubernetes.EnvoyDeployment.Replicas, DefaultKubernetesDeploymentReplicas()))
	assert.True(t, envoyProxyProvider.Kubernetes.EnvoyDeployment.Pod != nil)
	assert.True(t, reflect.DeepEqual(envoyProxyProvider.Kubernetes.EnvoyDeployment.Pod, DefaultKubernetesPod()))
	assert.True(t, envoyProxyProvider.Kubernetes.EnvoyDeployment.Container != nil)
	assert.True(t, reflect.DeepEqual(envoyProxyProvider.Kubernetes.EnvoyDeployment.Container, DefaultKubernetesContainer()))
	assert.True(t, envoyProxyProvider.Kubernetes.EnvoyDeployment.Container.Resources != nil)
	assert.True(t, reflect.DeepEqual(envoyProxyProvider.Kubernetes.EnvoyDeployment.Container.Resources, DefaultResourceRequirements()))

	assert.True(t, envoyProxyProvider.Kubernetes.EnvoyService != nil)
	assert.True(t, reflect.DeepEqual(envoyProxyProvider.Kubernetes.EnvoyService.Type, GetKubernetesServiceType(ServiceTypeLoadBalancer)))
}
