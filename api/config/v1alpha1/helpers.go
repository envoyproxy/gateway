// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	"fmt"
	"sort"
	"strings"

	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

// DefaultEnvoyGateway returns a new EnvoyGateway with default configuration parameters.
func DefaultEnvoyGateway() *EnvoyGateway {
	return &EnvoyGateway{
		metav1.TypeMeta{
			Kind:       KindEnvoyGateway,
			APIVersion: GroupVersion.String(),
		},
		EnvoyGatewaySpec{
			Gateway:  DefaultGateway(),
			Provider: DefaultEnvoyGatewayProvider(),
			Logging:  DefaultEnvoyGatewayLogging(),
			Admin:    DefaultEnvoyGatewayAdmin(),
		},
	}
}

// SetEnvoyGatewayDefaults sets default EnvoyGateway configuration parameters.
func (e *EnvoyGateway) SetEnvoyGatewayDefaults() {
	if e.TypeMeta.Kind == "" {
		e.TypeMeta.Kind = KindEnvoyGateway
	}
	if e.TypeMeta.APIVersion == "" {
		e.TypeMeta.APIVersion = GroupVersion.String()
	}
	if e.Provider == nil {
		e.Provider = DefaultEnvoyGatewayProvider()
	}
	if e.Gateway == nil {
		e.Gateway = DefaultGateway()
	}
	if e.Logging == nil {
		e.Logging = DefaultEnvoyGatewayLogging()
	}
	if e.Admin == nil {
		e.Admin = DefaultEnvoyGatewayAdmin()
	}
}

// GetEnvoyGatewayAdmin returns the EnvoyGatewayAdmin of EnvoyGateway or a default EnvoyGatewayAdmin if unspecified.
func (e *EnvoyGateway) GetEnvoyGatewayAdmin() *EnvoyGatewayAdmin {
	if e.Admin != nil {
		if e.Admin.Address == nil {
			e.Admin.Address = DefaultEnvoyGatewayAdminAddress()
		}
		return e.Admin
	}
	e.Admin = DefaultEnvoyGatewayAdmin()

	return e.Admin
}

// DefaultGateway returns a new Gateway with default configuration parameters.
func DefaultGateway() *Gateway {
	return &Gateway{
		ControllerName: GatewayControllerName,
	}
}

// DefaultEnvoyGatewayLogging returns a new EnvoyGatewayLogging with default configuration parameters.
func DefaultEnvoyGatewayLogging() *EnvoyGatewayLogging {
	return &EnvoyGatewayLogging{
		Level: map[EnvoyGatewayLogComponent]LogLevel{
			LogComponentGatewayDefault: LogLevelInfo,
		},
	}
}

// DefaultEnvoyGatewayLoggingLevel returns a new EnvoyGatewayLogging with default configuration parameters.
// When v1alpha1.LogComponentGatewayDefault specified, all other logging components are ignored.
func (logging *EnvoyGatewayLogging) DefaultEnvoyGatewayLoggingLevel(level LogLevel) LogLevel {
	if level != "" {
		return level
	}

	if logging.Level[LogComponentGatewayDefault] != "" {
		return logging.Level[LogComponentGatewayDefault]
	}

	return LogLevelInfo
}

// SetEnvoyGatewayLoggingDefaults sets default EnvoyGatewayLogging configuration parameters.
func (logging *EnvoyGatewayLogging) SetEnvoyGatewayLoggingDefaults() {
	if logging != nil && logging.Level != nil && logging.Level[LogComponentGatewayDefault] == "" {
		logging.Level[LogComponentGatewayDefault] = LogLevelInfo
	}
}

// DefaultEnvoyGatewayProvider returns a new EnvoyGatewayProvider with default configuration parameters.
func DefaultEnvoyGatewayProvider() *EnvoyGatewayProvider {
	return &EnvoyGatewayProvider{
		Type: ProviderTypeKubernetes,
	}
}

// GetEnvoyGatewayProvider returns the EnvoyGatewayProvider of EnvoyGateway or a default EnvoyGatewayProvider if unspecified.
func (e *EnvoyGateway) GetEnvoyGatewayProvider() *EnvoyGatewayProvider {
	if e.Provider != nil {
		return e.Provider
	}
	e.Provider = DefaultEnvoyGatewayProvider()

	return e.Provider
}

// DefaultEnvoyGatewayKubeProvider returns a new EnvoyGatewayKubernetesProvider with default settings.
func DefaultEnvoyGatewayKubeProvider() *EnvoyGatewayKubernetesProvider {
	return &EnvoyGatewayKubernetesProvider{
		RateLimitDeployment: DefaultKubernetesDeployment(DefaultRateLimitImage),
	}
}

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

// DefaultKubernetesDeploymentReplicas returns the default replica settings.
func DefaultKubernetesDeploymentReplicas() *int32 {
	repl := int32(DefaultDeploymentReplicas)
	return &repl
}

// DefaultKubernetesDeploymentStrategy returns the default deployment strategy settings.
func DefaultKubernetesDeploymentStrategy() *appv1.DeploymentStrategy {
	return &appv1.DeploymentStrategy{
		Type: appv1.RollingUpdateDeploymentStrategyType,
	}
}

// DefaultKubernetesContainerImage returns the default envoyproxy image.
func DefaultKubernetesContainerImage(image string) *string {
	return pointer.String(image)
}

// DefaultKubernetesDeployment returns a new KubernetesDeploymentSpec with default settings.
func DefaultKubernetesDeployment(image string) *KubernetesDeploymentSpec {
	return &KubernetesDeploymentSpec{
		Replicas:  DefaultKubernetesDeploymentReplicas(),
		Strategy:  DefaultKubernetesDeploymentStrategy(),
		Pod:       DefaultKubernetesPod(),
		Container: DefaultKubernetesContainer(image),
	}
}

// DefaultKubernetesPod returns a new KubernetesPodSpec with default settings.
func DefaultKubernetesPod() *KubernetesPodSpec {
	return &KubernetesPodSpec{}
}

// DefaultKubernetesContainer returns a new KubernetesContainerSpec with default settings.
func DefaultKubernetesContainer(image string) *KubernetesContainerSpec {
	return &KubernetesContainerSpec{
		Resources: DefaultResourceRequirements(),
		Image:     DefaultKubernetesContainerImage(image),
	}
}

// DefaultResourceRequirements returns a new ResourceRequirements with default settings.
func DefaultResourceRequirements() *corev1.ResourceRequirements {
	return &corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(DefaultDeploymentCPUResourceRequests),
			corev1.ResourceMemory: resource.MustParse(DefaultDeploymentMemoryResourceRequests),
		},
	}
}

// DefaultKubernetesService returns a new KubernetesServiceSpec with default settings.
func DefaultKubernetesService() *KubernetesServiceSpec {
	return &KubernetesServiceSpec{
		Type: DefaultKubernetesServiceType(),
	}
}

// DefaultKubernetesServiceType returns a new KubernetesServiceType with default settings.
func DefaultKubernetesServiceType() *ServiceType {
	return GetKubernetesServiceType(ServiceTypeLoadBalancer)
}

// GetKubernetesServiceType returns the KubernetesServiceType pointer.
func GetKubernetesServiceType(serviceType ServiceType) *ServiceType {
	return &serviceType
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

	return r.Kubernetes
}

// GetEnvoyGatewayKubeProvider returns the EnvoyGatewayKubernetesProvider of Provider or
// a default EnvoyGatewayKubernetesProvider if unspecified. If EnvoyGatewayProvider is not of
// type "Kubernetes", a nil EnvoyGatewayKubernetesProvider is returned.
func (r *EnvoyGatewayProvider) GetEnvoyGatewayKubeProvider() *EnvoyGatewayKubernetesProvider {
	if r.Type != ProviderTypeKubernetes {
		return nil
	}

	if r.Kubernetes == nil {
		r.Kubernetes = DefaultEnvoyGatewayKubeProvider()
		return r.Kubernetes
	}

	if r.Kubernetes.RateLimitDeployment == nil {
		r.Kubernetes.RateLimitDeployment = DefaultKubernetesDeployment(DefaultRateLimitImage)
	}

	r.Kubernetes.RateLimitDeployment.defaultKubernetesDeploymentSpec(DefaultRateLimitImage)

	return r.Kubernetes
}

// defaultKubernetesDeploymentSpec fill a default KubernetesDeploymentSpec if unspecified.
func (deployment *KubernetesDeploymentSpec) defaultKubernetesDeploymentSpec(image string) {
	if deployment.Replicas == nil {
		deployment.Replicas = DefaultKubernetesDeploymentReplicas()
	}

	if deployment.Strategy == nil {
		deployment.Strategy = DefaultKubernetesDeploymentStrategy()
	}

	if deployment.Pod == nil {
		deployment.Pod = DefaultKubernetesPod()
	}

	if deployment.Container == nil {
		deployment.Container = DefaultKubernetesContainer(image)
	}

	if deployment.Container.Resources == nil {
		deployment.Container.Resources = DefaultResourceRequirements()
	}

	if deployment.Container.Image == nil {
		deployment.Container.Image = DefaultKubernetesContainerImage(image)
	}
}

// DefaultEnvoyGatewayAdmin returns a new EnvoyGatewayAdmin with default configuration parameters.
func DefaultEnvoyGatewayAdmin() *EnvoyGatewayAdmin {
	return &EnvoyGatewayAdmin{
		Debug:   false,
		Address: DefaultEnvoyGatewayAdminAddress(),
	}
}

// DefaultEnvoyGatewayAdminAddress returns a new EnvoyGatewayAdminAddress with default configuration parameters.
func DefaultEnvoyGatewayAdminAddress() *EnvoyGatewayAdminAddress {
	return &EnvoyGatewayAdminAddress{
		Port: GatewayAdminPort,
		Host: GatewayAdminHost,
	}
}
