// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	"fmt"
	"sort"
	"strings"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
)

// DefaultEnvoyProxyProvider returns a new EnvoyProxyProvider with default settings.
func DefaultEnvoyProxyProvider() *EnvoyProxyProvider {
	return &EnvoyProxyProvider{
		Type: ProviderTypeKubernetes,
	}
}

// GetEnvoyProxyProvider returns the EnvoyProxyProvider of EnvoyProxy or a default EnvoyProxyProvider
// if unspecified.
func (e *EnvoyProxy) GetEnvoyProxyProvider() *EnvoyProxyProvider {
	if e.Spec.Provider != nil {
		return e.Spec.Provider
	}
	e.Spec.Provider = DefaultEnvoyProxyProvider()

	return e.Spec.Provider
}

// DefaultEnvoyProxyKubeProvider returns a new EnvoyProxyKubernetesProvider with default settings.
func DefaultEnvoyProxyKubeProvider() *EnvoyProxyKubernetesProvider {
	return &EnvoyProxyKubernetesProvider{
		EnvoyDeployment: DefaultKubernetesDeployment(DefaultEnvoyProxyImage),
		EnvoyService:    DefaultKubernetesService(),
	}
}

func DefaultEnvoyProxyHpaMetrics() []autoscalingv2.MetricSpec {
	return []autoscalingv2.MetricSpec{
		{
			Resource: &autoscalingv2.ResourceMetricSource{
				Name: v1.ResourceCPU,
				Target: autoscalingv2.MetricTarget{
					Type:               autoscalingv2.UtilizationMetricType,
					AverageUtilization: ptr.To[int32](80),
				},
			},
			Type: autoscalingv2.ResourceMetricSourceType,
		},
	}
}

// GetEnvoyProxyKubeProvider returns the EnvoyProxyKubernetesProvider of EnvoyProxyProvider or
// a default EnvoyProxyKubernetesProvider if unspecified. If EnvoyProxyProvider is not of
// type "Kubernetes", a nil EnvoyProxyKubernetesProvider is returned.
func (r *EnvoyProxyProvider) GetEnvoyProxyKubeProvider() *EnvoyProxyKubernetesProvider {
	if r.Type != ProviderTypeKubernetes {
		return nil
	}

	if r.Kubernetes == nil {
		r.Kubernetes = DefaultEnvoyProxyKubeProvider()
		return r.Kubernetes
	}

	if r.Kubernetes.EnvoyDeployment == nil {
		r.Kubernetes.EnvoyDeployment = DefaultKubernetesDeployment(DefaultEnvoyProxyImage)
	}

	r.Kubernetes.EnvoyDeployment.defaultKubernetesDeploymentSpec(DefaultEnvoyProxyImage)

	if r.Kubernetes.EnvoyService == nil {
		r.Kubernetes.EnvoyService = DefaultKubernetesService()
	}

	if r.Kubernetes.EnvoyService.Type == nil {
		r.Kubernetes.EnvoyService.Type = GetKubernetesServiceType(ServiceTypeLoadBalancer)
	}

	if r.Kubernetes.EnvoyHpa != nil {
		r.Kubernetes.EnvoyHpa.setDefault()
	}

	return r.Kubernetes
}

// DefaultEnvoyProxyLoggingLevel returns envoy proxy  v1alpha1.LogComponentGatewayDefault log level.
// If unspecified, defaults to "warn". When specified, all other logging components are ignored.
func (logging *ProxyLogging) DefaultEnvoyProxyLoggingLevel() LogLevel {
	if logging.Level[LogComponentDefault] != "" {
		return logging.Level[LogComponentDefault]
	}

	return LogLevelWarn
}

// GetEnvoyProxyComponentLevel returns envoy proxy component log level args.
// xref: https://www.envoyproxy.io/docs/envoy/latest/operations/cli#cmdoption-component-log-level
func (logging *ProxyLogging) GetEnvoyProxyComponentLevel() string {
	var args []string

	for component, level := range logging.Level {
		if component == LogComponentDefault {
			// Skip default component
			continue
		}

		if level != "" {
			args = append(args, fmt.Sprintf("%s:%s", component, level))
		}
	}

	sort.Strings(args)

	return strings.Join(args, ",")
}

// DefaultShutdownManagerContainerResourceRequirements returns a new ResourceRequirements with default settings.
func DefaultShutdownManagerContainerResourceRequirements() *v1.ResourceRequirements {
	return &v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse(DefaultShutdownManagerCPUResourceRequests),
			v1.ResourceMemory: resource.MustParse(DefaultShutdownManagerMemoryResourceRequests),
		},
	}
}
