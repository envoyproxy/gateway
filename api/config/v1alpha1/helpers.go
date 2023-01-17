// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DefaultEnvoyGateway returns a new EnvoyGateway with default configuration parameters.
func DefaultEnvoyGateway() *EnvoyGateway {
	gw := DefaultGateway()
	p := DefaultProvider()
	return &EnvoyGateway{
		metav1.TypeMeta{
			Kind:       KindEnvoyGateway,
			APIVersion: GroupVersion.String(),
		},
		EnvoyGatewaySpec{
			Gateway:  gw,
			Provider: p,
		},
	}
}

// SetDefaults sets default EnvoyGateway configuration parameters.
func (e *EnvoyGateway) SetDefaults() {
	if e.TypeMeta.Kind == "" {
		e.TypeMeta.Kind = KindEnvoyGateway
	}
	if e.TypeMeta.APIVersion == "" {
		e.TypeMeta.APIVersion = GroupVersion.String()
	}
	if e.Provider == nil {
		e.Provider = DefaultProvider()
	}
	if e.Gateway == nil {
		e.Gateway = DefaultGateway()
	}
}

// DefaultGateway returns a new Gateway with default configuration parameters.
func DefaultGateway() *Gateway {
	return &Gateway{
		ControllerName: GatewayControllerName,
	}
}

// DefaultProvider returns a new Provider with default configuration parameters.
func DefaultProvider() *Provider {
	return &Provider{
		Type: ProviderTypeKubernetes,
	}
}

// GetProvider returns the Provider of EnvoyGateway or a default Provider if unspecified.
func (e *EnvoyGateway) GetProvider() *Provider {
	if e.Provider != nil {
		return e.Provider
	}
	e.Provider = DefaultProvider()

	return e.Provider
}

// DefaultResourceProvider returns a new ResourceProvider with default settings.
func DefaultResourceProvider() *ResourceProvider {
	return &ResourceProvider{
		Type: ProviderTypeKubernetes,
	}
}

// GetProvider returns the ResourceProvider of EnvoyProxy or a default ResourceProvider
// if unspecified.
func (e *EnvoyProxy) GetProvider() *ResourceProvider {
	if e.Spec.Provider != nil {
		return e.Spec.Provider
	}
	e.Spec.Provider = DefaultResourceProvider()

	return e.Spec.Provider
}

// DefaultKubeResourceProvider returns a new KubernetesResourceProvider with default settings.
func DefaultKubeResourceProvider() *KubernetesResourceProvider {
	return &KubernetesResourceProvider{
		EnvoyDeployment: DefaultKubernetesDeployment(),
	}
}

// DefaultKubernetesDeployment returns a new KubernetesDeploymentSpec with default settings.
func DefaultKubernetesDeployment() *KubernetesDeploymentSpec {
	repl := int32(DefaultEnvoyReplicas)
	return &KubernetesDeploymentSpec{
		Replicas: &repl,
	}
}

// GetKubeResourceProvider returns the KubernetesResourceProvider of ResourceProvider or
// a default KubernetesResourceProvider if unspecified. If ResourceProvider is not of
// type "Kubernetes", a nil KubernetesResourceProvider is returned.
func (r *ResourceProvider) GetKubeResourceProvider() *KubernetesResourceProvider {
	if r.Type != ProviderTypeKubernetes {
		return nil
	}

	if r.Kubernetes == nil {
		r.Kubernetes = DefaultKubeResourceProvider()
		return r.Kubernetes
	}

	if r.Kubernetes.EnvoyDeployment == nil {
		r.Kubernetes.EnvoyDeployment = DefaultKubernetesDeployment()
		return r.Kubernetes
	}

	return r.Kubernetes
}
