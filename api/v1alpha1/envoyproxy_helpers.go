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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
)

// DefaultEnvoyProxyProvider returns a new EnvoyProxyProvider with default settings.
func DefaultEnvoyProxyProvider() *EnvoyProxyProvider {
	return &EnvoyProxyProvider{
		Type: EnvoyProxyProviderTypeKubernetes,
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

// DefaultEnvoyProxyHpaMetrics returns a default HPA metrics spec for EnvoyProxy.
func DefaultEnvoyProxyHpaMetrics() []autoscalingv2.MetricSpec {
	return []autoscalingv2.MetricSpec{
		{
			Resource: &autoscalingv2.ResourceMetricSource{
				Name: corev1.ResourceCPU,
				Target: autoscalingv2.MetricTarget{
					Type:               autoscalingv2.UtilizationMetricType,
					AverageUtilization: ptr.To[int32](80),
				},
			},
			Type: autoscalingv2.ResourceMetricSourceType,
		},
	}
}

// NeedToSwitchPorts returns true if the EnvoyProxy needs to switch ports.
func (e *EnvoyProxy) NeedToSwitchPorts() bool {
	if e.Spec.Provider == nil {
		return true
	}

	if e.Spec.Provider.Kubernetes == nil {
		return true
	}

	if e.Spec.Provider.Kubernetes.UseListenerPortAsContainerPort == nil {
		return true
	}

	return !*e.Spec.Provider.Kubernetes.UseListenerPortAsContainerPort
}

// GetEnvoyProxyHostProvider returns the EnvoyProxyHostProvider of EnvoyProxyProvider or
// a nil EnvoyProxyHostProvider if unspecified. If EnvoyProxyProvider is not of
// type "Host", a nil EnvoyProxyHostProvider is returned.
func (r *EnvoyProxyProvider) GetEnvoyProxyHostProvider() *EnvoyProxyHostProvider {
	if r.Type != EnvoyProxyProviderTypeHost {
		return nil
	}
	return r.Host
}

// GetEnvoyProxyKubeProvider returns the EnvoyProxyKubernetesProvider of EnvoyProxyProvider or
// a default EnvoyProxyKubernetesProvider if unspecified. If EnvoyProxyProvider is not of
// type "Kubernetes", a nil EnvoyProxyKubernetesProvider is returned.
func (r *EnvoyProxyProvider) GetEnvoyProxyKubeProvider() *EnvoyProxyKubernetesProvider {
	if r.Type != EnvoyProxyProviderTypeKubernetes {
		return nil
	}

	if r.Kubernetes == nil {
		r.Kubernetes = DefaultEnvoyProxyKubeProvider()
		return r.Kubernetes
	}

	// if EnvoyDeployment and EnvoyDaemonSet are both nil, use EnvoyDeployment
	if r.Kubernetes.EnvoyDeployment == nil && r.Kubernetes.EnvoyDaemonSet == nil {
		r.Kubernetes.EnvoyDeployment = DefaultKubernetesDeployment(DefaultEnvoyProxyImage)
	}

	// if use EnvoyDeployment, set default values
	if r.Kubernetes.EnvoyDeployment != nil {
		r.Kubernetes.EnvoyDeployment.defaultKubernetesDeploymentSpec(DefaultEnvoyProxyImage)
	}

	// if use EnvoyDaemonSet, set default values
	if r.Kubernetes.EnvoyDaemonSet != nil {
		r.Kubernetes.EnvoyDaemonSet.defaultKubernetesDaemonSetSpec(DefaultEnvoyProxyImage)
	}

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

// GetEnvoyVersion returns the version of Envoy to use.
// This method gracefully handles nil pointers.
func (e *EnvoyProxyHostProvider) GetEnvoyVersion() string {
	if e == nil || e.EnvoyVersion == nil {
		return ""
	}
	return *e.EnvoyVersion
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

	if len(args) == 0 {
		// use "misc:error" as default
		args = []string{"misc:error"}
	}

	sort.Strings(args)

	return strings.Join(args, ",")
}

// DefaultShutdownManagerContainerResourceRequirements returns a new ResourceRequirements with default settings.
func DefaultShutdownManagerContainerResourceRequirements() *corev1.ResourceRequirements {
	return &corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(DefaultShutdownManagerCPUResourceRequests),
			corev1.ResourceMemory: resource.MustParse(DefaultShutdownManagerMemoryResourceRequests),
		},
	}
}

// String returns the string representation of the EnvoyFilter type.
func (f EnvoyFilter) String() string {
	return string(f)
}
