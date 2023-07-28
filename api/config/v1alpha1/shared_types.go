// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	// DefaultDeploymentReplicas is the default number of deployment replicas.
	DefaultDeploymentReplicas = 1
	// DefaultDeploymentCPUResourceRequests for deployment cpu resource
	DefaultDeploymentCPUResourceRequests = "100m"
	// DefaultDeploymentMemoryResourceRequests for deployment memory resource
	DefaultDeploymentMemoryResourceRequests = "512Mi"
	// DefaultEnvoyProxyImage is the default image used by envoyproxy
	DefaultEnvoyProxyImage = "envoyproxy/envoy:v1.27-latest"
	// DefaultRateLimitImage is the default image used by ratelimit.
	DefaultRateLimitImage = "envoyproxy/ratelimit:e059638d"
)

// GroupVersionKind unambiguously identifies a Kind.
// It can be converted to k8s.io/apimachinery/pkg/runtime/schema.GroupVersionKind
type GroupVersionKind struct {
	Group   string `json:"group"`
	Version string `json:"version"`
	Kind    string `json:"kind"`
}

// ProviderType defines the types of providers supported by Envoy Gateway.
//
// +kubebuilder:validation:Enum=Kubernetes
type ProviderType string

const (
	// ProviderTypeKubernetes defines the "Kubernetes" provider.
	ProviderTypeKubernetes ProviderType = "Kubernetes"

	// ProviderTypeFile defines the "File" provider. This type is not implemented
	// until https://github.com/envoyproxy/gateway/issues/1001 is fixed.
	ProviderTypeFile ProviderType = "File"
)

// KubernetesDeploymentSpec defines the desired state of the Kubernetes deployment resource.
type KubernetesDeploymentSpec struct {
	// Replicas is the number of desired pods. Defaults to 1.
	//
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// The deployment strategy to use to replace existing pods with new ones.
	// +optional
	Strategy *appv1.DeploymentStrategy `json:"strategy,omitempty"`

	// Pod defines the desired annotations and securityContext of container.
	//
	// +optional
	Pod *KubernetesPodSpec `json:"pod,omitempty"`

	// Container defines the resources and securityContext of container.
	//
	// +optional
	Container *KubernetesContainerSpec `json:"container,omitempty"`

	// TODO: Expose config as use cases are better understood, e.g. labels.
}

// KubernetesPodSpec defines the desired state of the Kubernetes pod resource.
type KubernetesPodSpec struct {
	// Annotations are the annotations that should be appended to the pods.
	// By default, no pod annotations are appended.
	//
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// Labels are the additional labels that should be tagged to the pods.
	// By default, no additional pod labels are tagged.
	//
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// SecurityContext holds pod-level security attributes and common container settings.
	// Optional: Defaults to empty.  See type description for default values of each field.
	//
	// +optional
	SecurityContext *corev1.PodSecurityContext `json:"securityContext,omitempty"`

	// If specified, the pod's scheduling constraints.
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// If specified, the pod's tolerations.
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// Volumes that can be mounted by containers belonging to the pod.
	// More info: https://kubernetes.io/docs/concepts/storage/volumes
	//
	// +optional
	Volumes []corev1.Volume `json:"volumes,omitempty"`
}

// KubernetesContainerSpec defines the desired state of the Kubernetes container resource.
type KubernetesContainerSpec struct {
	// List of environment variables to set in the container.
	//
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`

	// Resources required by this container.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	//
	// +optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`

	// SecurityContext defines the security options the container should be run with.
	// If set, the fields of SecurityContext override the equivalent fields of PodSecurityContext.
	// More info: https://kubernetes.io/docs/tasks/configure-pod-container/security-context/
	//
	// +optional
	SecurityContext *corev1.SecurityContext `json:"securityContext,omitempty"`

	// Image specifies the EnvoyProxy container image to be used, instead of the default image.
	//
	// +optional
	Image *string `json:"image,omitempty"`

	// VolumeMounts are volumes to mount into the container's filesystem.
	// Cannot be updated.
	//
	// +optional
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`
}

// ServiceType string describes ingress methods for a service
// +enum
// +kubebuilder:validation:Enum=ClusterIP;LoadBalancer;NodePort
type ServiceType string

const (
	// ServiceTypeClusterIP means a service will only be accessible inside the
	// cluster, via the cluster IP.
	ServiceTypeClusterIP ServiceType = "ClusterIP"

	// ServiceTypeLoadBalancer means a service will be exposed via an
	// external load balancer (if the cloud provider supports it).
	ServiceTypeLoadBalancer ServiceType = "LoadBalancer"

	// ServiceTypeNodePort means a service will be exposed on each Kubernetes Node
	// at a static Port, common across all Nodes.
	ServiceTypeNodePort ServiceType = "NodePort"
)

// KubernetesServiceSpec defines the desired state of the Kubernetes service resource.
type KubernetesServiceSpec struct {
	// Annotations that should be appended to the service.
	// By default, no annotations are appended.
	//
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// Type determines how the Service is exposed. Defaults to LoadBalancer.
	// Valid options are ClusterIP, LoadBalancer and NodePort.
	// "LoadBalancer" means a service will be exposed via an external load balancer (if the cloud provider supports it).
	// "ClusterIP" means a service will only be accessible inside the cluster, via the cluster IP.
	// "NodePort" means a service will be exposed on a static Port on all Nodes of the cluster.
	// +kubebuilder:default:="LoadBalancer"
	// +optional
	Type *ServiceType `json:"type,omitempty"`

	// TODO: Expose config as use cases are better understood, e.g. labels.
}

// LogLevel defines a log level for Envoy Gateway and EnvoyProxy system logs.
// This type is not implemented for EnvoyProxy until
// https://github.com/envoyproxy/gateway/issues/280 is fixed.
// +kubebuilder:validation:Enum=debug;info;error;warn
type LogLevel string

const (
	// LogLevelDebug defines the "debug" logging level.
	LogLevelDebug LogLevel = "debug"

	// LogLevelInfo defines the "Info" logging level.
	LogLevelInfo LogLevel = "info"

	// LogLevelWarn defines the "Warn" logging level.
	LogLevelWarn LogLevel = "warn"

	// LogLevelError defines the "Error" logging level.
	LogLevelError LogLevel = "error"
)

// XDSTranslatorHook defines the types of hooks that an Envoy Gateway extension may support
// for the xds-translator
//
// +kubebuilder:validation:Enum=VirtualHost;Route;HTTPListener;Translation
type XDSTranslatorHook string

const (
	XDSVirtualHost  XDSTranslatorHook = "VirtualHost"
	XDSRoute        XDSTranslatorHook = "Route"
	XDSHTTPListener XDSTranslatorHook = "HTTPListener"
	XDSTranslation  XDSTranslatorHook = "Translation"
)
