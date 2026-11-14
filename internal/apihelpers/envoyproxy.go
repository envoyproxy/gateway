// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package apihelpers

import (
	"fmt"
	"sort"
	"strings"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

// DefaultEnvoyProxyProvider returns a new EnvoyProxyProvider with default settings.
func DefaultEnvoyProxyProvider() *egv1a1.EnvoyProxyProvider {
	return &egv1a1.EnvoyProxyProvider{
		Type: egv1a1.EnvoyProxyProviderTypeKubernetes,
	}
}

// GetEnvoyProxyProvider returns the EnvoyProxyProvider of EnvoyProxy or a default EnvoyProxyProvider
// if unspecified.
func GetEnvoyProxyProvider(e *egv1a1.EnvoyProxy) *egv1a1.EnvoyProxyProvider {
	if e.Spec.Provider != nil {
		return e.Spec.Provider
	}
	e.Spec.Provider = DefaultEnvoyProxyProvider()

	return e.Spec.Provider
}

// DefaultEnvoyProxyKubeProvider returns a new EnvoyProxyKubernetesProvider with default settings.
func DefaultEnvoyProxyKubeProvider() *egv1a1.EnvoyProxyKubernetesProvider {
	return &egv1a1.EnvoyProxyKubernetesProvider{
		EnvoyDeployment: DefaultKubernetesDeployment(egv1a1.DefaultEnvoyProxyImage),
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
					AverageUtilization: new(int32(80)),
				},
			},
			Type: autoscalingv2.ResourceMetricSourceType,
		},
	}
}

// NeedToSwitchPorts returns true if the EnvoyProxy needs to switch ports.
func NeedToSwitchPorts(e *egv1a1.EnvoyProxy) bool {
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
func GetEnvoyProxyHostProvider(r *egv1a1.EnvoyProxyProvider) *egv1a1.EnvoyProxyHostProvider {
	if r.Type != egv1a1.EnvoyProxyProviderTypeHost {
		return nil
	}
	return r.Host
}

// GetEnvoyProxyKubeProvider returns the EnvoyProxyKubernetesProvider of EnvoyProxyProvider or
// a default EnvoyProxyKubernetesProvider if unspecified. If EnvoyProxyProvider is not of
// type "Kubernetes", a nil EnvoyProxyKubernetesProvider is returned.
func GetEnvoyProxyKubeProvider(r *egv1a1.EnvoyProxyProvider) *egv1a1.EnvoyProxyKubernetesProvider {
	if r.Type != egv1a1.EnvoyProxyProviderTypeKubernetes {
		return nil
	}

	if r.Kubernetes == nil {
		r.Kubernetes = DefaultEnvoyProxyKubeProvider()
		return r.Kubernetes
	}

	// if EnvoyDeployment and EnvoyDaemonSet are both nil, use EnvoyDeployment
	if r.Kubernetes.EnvoyDeployment == nil && r.Kubernetes.EnvoyDaemonSet == nil {
		r.Kubernetes.EnvoyDeployment = DefaultKubernetesDeployment(egv1a1.DefaultEnvoyProxyImage)
	}

	// if use EnvoyDeployment, set default values
	if r.Kubernetes.EnvoyDeployment != nil {
		defaultKubernetesDeploymentSpec(r.Kubernetes.EnvoyDeployment, egv1a1.DefaultEnvoyProxyImage)
	}

	// if use EnvoyDaemonSet, set default values
	if r.Kubernetes.EnvoyDaemonSet != nil {
		defaultKubernetesDaemonSetSpec(r.Kubernetes.EnvoyDaemonSet, egv1a1.DefaultEnvoyProxyImage)
	}

	if r.Kubernetes.EnvoyService == nil {
		r.Kubernetes.EnvoyService = DefaultKubernetesService()
	}

	if r.Kubernetes.EnvoyService.Type == nil {
		r.Kubernetes.EnvoyService.Type = GetKubernetesServiceType(egv1a1.ServiceTypeLoadBalancer)
	}

	if r.Kubernetes.EnvoyHpa != nil {
		setHPADefaults(r.Kubernetes.EnvoyHpa)
	}

	return r.Kubernetes
}

// GetEnvoyVersion returns the version of Envoy to use.
// This function gracefully handles nil pointers.
func GetEnvoyVersion(e *egv1a1.EnvoyProxyHostProvider) string {
	if e == nil || e.EnvoyVersion == nil {
		return ""
	}
	return *e.EnvoyVersion
}

// DefaultEnvoyProxyLoggingLevel returns envoy proxy  v1alpha1.LogComponentGatewayDefault log level.
// If unspecified, defaults to "warn". When specified, all other logging components are ignored.
func DefaultEnvoyProxyLoggingLevel(logging *egv1a1.ProxyLogging) egv1a1.LogLevel {
	if logging.Level[egv1a1.LogComponentDefault] != "" {
		return logging.Level[egv1a1.LogComponentDefault]
	}

	return egv1a1.LogLevelWarn
}

// GetEnvoyProxyComponentLevel returns envoy proxy component log level args.
// xref: https://www.envoyproxy.io/docs/envoy/latest/operations/cli#cmdoption-component-log-level
func GetEnvoyProxyComponentLevel(logging *egv1a1.ProxyLogging) string {
	var args []string

	for component, level := range logging.Level {
		if component == egv1a1.LogComponentDefault {
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
			corev1.ResourceCPU:    resource.MustParse(egv1a1.DefaultShutdownManagerCPUResourceRequests),
			corev1.ResourceMemory: resource.MustParse(egv1a1.DefaultShutdownManagerMemoryResourceRequests),
		},
	}
}
