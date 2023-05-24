// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetEnvoyGatewayProvider(t *testing.T) {
	envoyGateway := &EnvoyGateway{
		TypeMeta:         metav1.TypeMeta{},
		EnvoyGatewaySpec: EnvoyGatewaySpec{Provider: DefaultEnvoyGatewayProvider()},
	}
	assert.True(t, envoyGateway.Provider != nil)

	envoyGatewayProvider := envoyGateway.GetEnvoyGatewayProvider()
	assert.True(t, envoyGatewayProvider.Kubernetes == nil)
	assert.Equal(t, envoyGateway.Provider, envoyGatewayProvider)

	envoyGatewayProvider.Kubernetes = DefaultEnvoyGatewayKubeProvider()
	assert.Equal(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment, DefaultKubernetesDeployment(DefaultRateLimitImage))

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
	assert.Equal(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment, DefaultKubernetesDeployment(DefaultRateLimitImage))
	assert.True(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Replicas != nil)
	assert.Equal(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Replicas, DefaultKubernetesDeploymentReplicas())
	assert.True(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Pod != nil)
	assert.Equal(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Pod, DefaultKubernetesPod())
	assert.True(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Container != nil)
	assert.Equal(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Container, DefaultKubernetesContainer(DefaultRateLimitImage))
	assert.True(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Container.Resources != nil)
	assert.Equal(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Container.Resources, DefaultResourceRequirements())
	assert.True(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Container.Image != nil)
	assert.Equal(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Container.Image, DefaultKubernetesContainerImage(DefaultRateLimitImage))
}

func TestGetEnvoyProxyProvider(t *testing.T) {
	t.Run("deployment", func(t *testing.T) {
		envoyProxy := &EnvoyProxy{Spec: EnvoyProxySpec{Provider: DefaultEnvoyProxyProvider()}}
		assert.NotNil(t, envoyProxy.Spec.Provider)

		envoyProxyProvider := envoyProxy.GetEnvoyProxyProvider()
		assert.Nil(t, envoyProxyProvider.Kubernetes)
		assert.True(t, reflect.DeepEqual(envoyProxy.Spec.Provider, envoyProxyProvider))

		envoyProxyKubeProvider := envoyProxyProvider.GetEnvoyProxyKubeProvider()
		assert.True(t, reflect.DeepEqual(envoyProxyProvider.Kubernetes, envoyProxyKubeProvider))

		assert.Equal(t, envoyProxyProvider.Kubernetes.EnvoyDeployment, DefaultKubernetesDeployment(DefaultEnvoyProxyImage))
		assert.Equal(t, envoyProxyProvider.Kubernetes.EnvoyDeployment.Replicas, DefaultKubernetesDeploymentReplicas())
		assert.Equal(t, envoyProxyProvider.Kubernetes.EnvoyDeployment.Pod, DefaultKubernetesPod())
		assert.Equal(t, envoyProxyProvider.Kubernetes.EnvoyDeployment.Container, DefaultKubernetesContainer(DefaultEnvoyProxyImage))
		assert.Equal(t, envoyProxyProvider.Kubernetes.EnvoyDeployment.Container.Resources, DefaultResourceRequirements())
		assert.Equal(t, envoyProxyProvider.Kubernetes.EnvoyDeployment.Container.Image, DefaultKubernetesContainerImage(DefaultEnvoyProxyImage))

		assert.NotNil(t, envoyProxyProvider.Kubernetes.EnvoyService)
		assert.True(t, reflect.DeepEqual(envoyProxyProvider.Kubernetes.EnvoyService.Type, GetKubernetesServiceType(ServiceTypeLoadBalancer)))
	})

	t.Run("daemonSet", func(t *testing.T) {
		envoyProxy := &EnvoyProxy{Spec: EnvoyProxySpec{Provider: &EnvoyProxyProvider{
			Type: ProviderTypeKubernetes,
			Kubernetes: &EnvoyProxyKubernetesProvider{
				EnvoyDaemonSet: &KubernetesDaemonSetSpec{},
			},
		}}}
		assert.NotNil(t, envoyProxy.Spec.Provider)

		envoyProxyProvider := envoyProxy.GetEnvoyProxyProvider()
		assert.NotNil(t, envoyProxyProvider.Kubernetes)
		assert.True(t, reflect.DeepEqual(envoyProxy.Spec.Provider, envoyProxyProvider))

		envoyProxyKubeProvider := envoyProxyProvider.GetEnvoyProxyKubeProvider()
		assert.Nil(t, envoyProxyKubeProvider.EnvoyDeployment)
		assert.NotNil(t, envoyProxyKubeProvider.EnvoyDaemonSet)
		assert.Equal(t, envoyProxyKubeProvider.EnvoyDaemonSet.Pod, DefaultKubernetesPod())
		assert.Equal(t, envoyProxyKubeProvider.EnvoyDaemonSet.Container, DefaultKubernetesContainer(DefaultEnvoyProxyImage))
	})
}
