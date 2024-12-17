// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

const (
	// DefaultDeploymentReplicas is the default number of deployment replicas.
	DefaultDeploymentReplicas = 1
	// DefaultDeploymentCPUResourceRequests for deployment cpu resource
	DefaultDeploymentCPUResourceRequests = "100m"
	// DefaultDeploymentMemoryResourceRequests for deployment memory resource
	DefaultDeploymentMemoryResourceRequests = "512Mi"
	// DefaultEnvoyProxyImage is the default image used by envoyproxy
	DefaultEnvoyProxyImage = "docker.io/envoyproxy/envoy:distroless-dev"
	// DefaultShutdownManagerCPUResourceRequests for shutdown manager cpu resource
	DefaultShutdownManagerCPUResourceRequests = "10m"
	// DefaultShutdownManagerMemoryResourceRequests for shutdown manager memory resource
	DefaultShutdownManagerMemoryResourceRequests = "32Mi"
	// DefaultShutdownManagerImage is the default image used for the shutdown manager.
	DefaultShutdownManagerImage = "docker.io/envoyproxy/gateway-dev:latest"
	// DefaultRateLimitImage is the default image used by ratelimit.
	DefaultRateLimitImage = "docker.io/envoyproxy/ratelimit:master"
	// HTTPProtocol is the common-used http protocol.
	HTTPProtocol = "http"
	// GRPCProtocol is the common-used grpc protocol.
	GRPCProtocol = "grpc"
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
// +kubebuilder:validation:Enum=Kubernetes;Custom
type ProviderType string

const (
	// ProviderTypeKubernetes defines the "Kubernetes" provider.
	ProviderTypeKubernetes ProviderType = "Kubernetes"

	// ProviderTypeCustom defines the "Custom" provider.
	ProviderTypeCustom ProviderType = "Custom"
)

// KubernetesDeploymentSpec defines the desired state of the Kubernetes deployment resource.
type KubernetesDeploymentSpec struct {
	// Patch defines how to perform the patch operation to deployment
	//
	// +optional
	Patch *KubernetesPatchSpec `json:"patch,omitempty"`

	// Replicas is the number of desired pods. Defaults to 1.
	//
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// The deployment strategy to use to replace existing pods with new ones.
	// +optional
	Strategy *appsv1.DeploymentStrategy `json:"strategy,omitempty"`

	// Pod defines the desired specification of pod.
	//
	// +optional
	Pod *KubernetesPodSpec `json:"pod,omitempty"`

	// Container defines the desired specification of main container.
	//
	// +optional
	Container *KubernetesContainerSpec `json:"container,omitempty"`

	// List of initialization containers belonging to the pod.
	// More info: https://kubernetes.io/docs/concepts/workloads/pods/init-containers/
	//
	// +optional
	InitContainers []corev1.Container `json:"initContainers,omitempty"`

	// Name of the deployment.
	// When unset, this defaults to an autogenerated name.
	//
	// +optional
	Name *string `json:"name,omitempty"`
	// TODO: Expose config as use cases are better understood, e.g. labels.
}

// KubernetesDaemonSetSpec defines the desired state of the Kubernetes daemonset resource.
type KubernetesDaemonSetSpec struct {
	// Patch defines how to perform the patch operation to daemonset
	//
	// +optional
	Patch *KubernetesPatchSpec `json:"patch,omitempty"`

	// The daemonset strategy to use to replace existing pods with new ones.
	// +optional
	Strategy *appsv1.DaemonSetUpdateStrategy `json:"strategy,omitempty"`

	// Pod defines the desired specification of pod.
	//
	// +optional
	Pod *KubernetesPodSpec `json:"pod,omitempty"`

	// Container defines the desired specification of main container.
	//
	// +optional
	Container *KubernetesContainerSpec `json:"container,omitempty"`

	// Name of the daemonSet.
	// When unset, this defaults to an autogenerated name.
	//
	// +optional
	Name *string `json:"name,omitempty"`
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

	// ImagePullSecrets is an optional list of references to secrets
	// in the same namespace to use for pulling any of the images used by this PodSpec.
	// If specified, these secrets will be passed to individual puller implementations for them to use.
	// More info: https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod
	//
	// +optional
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`

	// NodeSelector is a selector which must be true for the pod to fit on a node.
	// Selector which must match a node's labels for the pod to be scheduled on that node.
	// More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/
	//
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// TopologySpreadConstraints describes how a group of pods ought to spread across topology
	// domains. Scheduler will schedule pods in a way which abides by the constraints.
	// All topologySpreadConstraints are ANDed.
	//
	// +optional
	TopologySpreadConstraints []corev1.TopologySpreadConstraint `json:"topologySpreadConstraints,omitempty"`
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

// ServiceExternalTrafficPolicy describes how nodes distribute service traffic they
// receive on one of the Service's "externally-facing" addresses (NodePorts, ExternalIPs,
// and LoadBalancer IPs.
// +enum
// +kubebuilder:validation:Enum=Local;Cluster
type ServiceExternalTrafficPolicy string

const (
	// ServiceExternalTrafficPolicyCluster routes traffic to all endpoints.
	ServiceExternalTrafficPolicyCluster ServiceExternalTrafficPolicy = "Cluster"

	// ServiceExternalTrafficPolicyLocal preserves the source IP of the traffic by
	// routing only to endpoints on the same node as the traffic was received on
	// (dropping the traffic if there are no local endpoints).
	ServiceExternalTrafficPolicyLocal ServiceExternalTrafficPolicy = "Local"
)

// KubernetesServiceSpec defines the desired state of the Kubernetes service resource.
// +kubebuilder:validation:XValidation:message="allocateLoadBalancerNodePorts can only be set for LoadBalancer type",rule="!has(self.allocateLoadBalancerNodePorts) || self.type == 'LoadBalancer'"
// +kubebuilder:validation:XValidation:message="loadBalancerSourceRanges can only be set for LoadBalancer type",rule="!has(self.loadBalancerSourceRanges) || self.type == 'LoadBalancer'"
// +kubebuilder:validation:XValidation:message="loadBalancerIP can only be set for LoadBalancer type",rule="!has(self.loadBalancerIP) || self.type == 'LoadBalancer'"
type KubernetesServiceSpec struct {
	// Annotations that should be appended to the service.
	// By default, no annotations are appended.
	//
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// Labels that should be appended to the service.
	// By default, no labels are appended.
	//
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// Type determines how the Service is exposed. Defaults to LoadBalancer.
	// Valid options are ClusterIP, LoadBalancer and NodePort.
	// "LoadBalancer" means a service will be exposed via an external load balancer (if the cloud provider supports it).
	// "ClusterIP" means a service will only be accessible inside the cluster, via the cluster IP.
	// "NodePort" means a service will be exposed on a static Port on all Nodes of the cluster.
	// +kubebuilder:default:="LoadBalancer"
	// +optional
	Type *ServiceType `json:"type,omitempty"`

	// LoadBalancerClass, when specified, allows for choosing the LoadBalancer provider
	// implementation if more than one are available or is otherwise expected to be specified
	// +optional
	LoadBalancerClass *string `json:"loadBalancerClass,omitempty"`

	// AllocateLoadBalancerNodePorts defines if NodePorts will be automatically allocated for
	// services with type LoadBalancer. Default is "true". It may be set to "false" if the cluster
	// load-balancer does not rely on NodePorts. If the caller requests specific NodePorts (by specifying a
	// value), those requests will be respected, regardless of this field. This field may only be set for
	// services with type LoadBalancer and will be cleared if the type is changed to any other type.
	// +optional
	AllocateLoadBalancerNodePorts *bool `json:"allocateLoadBalancerNodePorts,omitempty"`

	// LoadBalancerSourceRanges defines a list of allowed IP addresses which will be configured as
	// firewall rules on the platform providers load balancer. This is not guaranteed to be working as
	// it happens outside of kubernetes and has to be supported and handled by the platform provider.
	// This field may only be set for services with type LoadBalancer and will be cleared if the type
	// is changed to any other type.
	// +optional
	LoadBalancerSourceRanges []string `json:"loadBalancerSourceRanges,omitempty"`

	// LoadBalancerIP defines the IP Address of the underlying load balancer service. This field
	// may be ignored if the load balancer provider does not support this feature.
	// This field has been deprecated in Kubernetes, but it is still used for setting the IP Address in some cloud
	// providers such as GCP.
	//
	// +kubebuilder:validation:XValidation:message="loadBalancerIP must be a valid IPv4 address",rule="self.matches(r\"^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$\")"
	// +optional
	LoadBalancerIP *string `json:"loadBalancerIP,omitempty"`

	// ExternalTrafficPolicy determines the externalTrafficPolicy for the Envoy Service. Valid options
	// are Local and Cluster. Default is "Local". "Local" means traffic will only go to pods on the node
	// receiving the traffic. "Cluster" means connections are loadbalanced to all pods in the cluster.
	// +kubebuilder:default:="Local"
	// +optional
	ExternalTrafficPolicy *ServiceExternalTrafficPolicy `json:"externalTrafficPolicy,omitempty"`

	// Patch defines how to perform the patch operation to the service
	//
	// +optional
	Patch *KubernetesPatchSpec `json:"patch,omitempty"`

	// Name of the service.
	// When unset, this defaults to an autogenerated name.
	//
	// +optional
	Name *string `json:"name,omitempty"`

	// TODO: Expose config as use cases are better understood, e.g. labels.
}

// LogLevel defines a log level for Envoy Gateway and EnvoyProxy system logs.
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

// StringMatch defines how to match any strings.
// This is a general purpose match condition that can be used by other EG APIs
// that need to match against a string.
type StringMatch struct {
	// Type specifies how to match against a string.
	//
	// +optional
	// +kubebuilder:default=Exact
	Type *StringMatchType `json:"type,omitempty"`

	// Value specifies the string value that the match must have.
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=1024
	Value string `json:"value"`
}

// StringMatchType specifies the semantics of how a string value should be compared.
// Valid MatchType values are "Exact", "Prefix", "Suffix", "RegularExpression".
//
// +kubebuilder:validation:Enum=Exact;Prefix;Suffix;RegularExpression
type StringMatchType string

const (
	// StringMatchExact :the input string must match exactly the match value.
	StringMatchExact StringMatchType = "Exact"

	// StringMatchPrefix :the input string must start with the match value.
	StringMatchPrefix StringMatchType = "Prefix"

	// StringMatchSuffix :the input string must end with the match value.
	StringMatchSuffix StringMatchType = "Suffix"

	// StringMatchRegularExpression :The input string must match the regular expression
	// specified in the match value.
	// The regex string must adhere to the syntax documented in
	// https://github.com/google/re2/wiki/Syntax.
	StringMatchRegularExpression StringMatchType = "RegularExpression"
)

// KubernetesPodDisruptionBudgetSpec defines Kubernetes PodDisruptionBudget settings of Envoy Proxy Deployment.
type KubernetesPodDisruptionBudgetSpec struct {
	// MinAvailable specifies the minimum number of pods that must be available at all times during voluntary disruptions,
	// such as node drains or updates. This setting ensures that your envoy proxy maintains a certain level of availability
	// and resilience during maintenance operations.
	// +optional
	MinAvailable *int32 `json:"minAvailable,omitempty"`

	// Patch defines how to perform the patch operation to the PodDisruptionBudget
	//
	// +optional
	Patch *KubernetesPatchSpec `json:"patch,omitempty"`
}

// KubernetesHorizontalPodAutoscalerSpec defines Kubernetes Horizontal Pod Autoscaler settings of Envoy Proxy Deployment.
// When HPA is enabled, it is recommended that the value in `KubernetesDeploymentSpec.replicas` be removed, otherwise
// Envoy Gateway will revert back to this value every time reconciliation occurs.
// See k8s.io.autoscaling.v2.HorizontalPodAutoScalerSpec.
//
// +kubebuilder:validation:XValidation:message="maxReplicas cannot be less than minReplicas",rule="!has(self.minReplicas) || self.maxReplicas >= self.minReplicas"
type KubernetesHorizontalPodAutoscalerSpec struct {
	// minReplicas is the lower limit for the number of replicas to which the autoscaler
	// can scale down. It defaults to 1 replica.
	//
	// +kubebuilder:validation:XValidation:message="minReplicas must be greater than 0",rule="self > 0"
	// +optional
	MinReplicas *int32 `json:"minReplicas,omitempty"`

	// maxReplicas is the upper limit for the number of replicas to which the autoscaler can scale up.
	// It cannot be less that minReplicas.
	//
	// +kubebuilder:validation:XValidation:message="maxReplicas must be greater than 0",rule="self > 0"
	MaxReplicas *int32 `json:"maxReplicas"`

	// metrics contains the specifications for which to use to calculate the
	// desired replica count (the maximum replica count across all metrics will
	// be used).
	// If left empty, it defaults to being based on CPU utilization with average on 80% usage.
	//
	// +optional
	Metrics []autoscalingv2.MetricSpec `json:"metrics,omitempty"`

	// behavior configures the scaling behavior of the target
	// in both Up and Down directions (scaleUp and scaleDown fields respectively).
	// If not set, the default HPAScalingRules for scale up and scale down are used.
	// See k8s.io.autoscaling.v2.HorizontalPodAutoScalerBehavior.
	//
	// +optional
	Behavior *autoscalingv2.HorizontalPodAutoscalerBehavior `json:"behavior,omitempty"`

	// Patch defines how to perform the patch operation to the HorizontalPodAutoscaler
	//
	// +optional
	Patch *KubernetesPatchSpec `json:"patch,omitempty"`
}

// HTTPStatus defines the http status code.
// +kubebuilder:validation:Minimum=100
// +kubebuilder:validation:Maximum=600
// +kubebuilder:validation:ExclusiveMaximum=true
type HTTPStatus int

// MergeType defines the type of merge operation
type MergeType string

const (
	// StrategicMerge indicates a strategic merge patch type
	StrategicMerge MergeType = "StrategicMerge"
	// JSONMerge indicates a JSON merge patch type
	JSONMerge MergeType = "JSONMerge"
)

// KubernetesPatchSpec defines how to perform the patch operation.
// Note that `value` can be an in-line YAML document, as can be seen in e.g. (the example of patching the Envoy proxy Deployment)[https://gateway.envoyproxy.io/docs/tasks/operations/customize-envoyproxy/#patching-deployment-for-envoyproxy].
// Note also that, currently, strings containing literal JSON are _rejected_.
type KubernetesPatchSpec struct {
	// Type is the type of merge operation to perform
	//
	// By default, StrategicMerge is used as the patch type.
	// +optional
	Type *MergeType `json:"type,omitempty"`

	// Object contains the raw configuration for merged object
	Value apiextensionsv1.JSON `json:"value"`
}

// BackendRef defines how an ObjectReference that is specific to BackendRef.
type BackendRef struct {
	// BackendObjectReference references a Kubernetes object that represents the backend.
	// Only Service kind is supported for now.
	gwapiv1.BackendObjectReference `json:",inline"`
	// Fallback indicates whether the backend is designated as a fallback.
	// Multiple fallback backends can be configured.
	// It is highly recommended to configure active or passive health checks to ensure that failover can be detected
	// when the active backends become unhealthy and to automatically readjust once the primary backends are healthy again.
	// The overprovisioning factor is set to 1.4, meaning the fallback backends will only start receiving traffic when
	// the health of the active backends falls below 72%.
	//
	// +optional
	Fallback *bool `json:"fallback,omitempty"`
}

// BackendCluster contains all the configuration required for configuring access
// to a backend. This can include multiple endpoints, and settings that apply for
// managing the connection to all these endpoints.
type BackendCluster struct {
	// BackendRef references a Kubernetes object that represents the
	// backend server to which the authorization request will be sent.
	//
	// Deprecated: Use BackendRefs instead.
	// +optional
	BackendRef *gwapiv1.BackendObjectReference `json:"backendRef,omitempty"`

	// BackendRefs references a Kubernetes object that represents the
	// backend server to which the authorization request will be sent.
	//
	// +kubebuilder:validation:MaxItems=16
	// +optional
	BackendRefs []BackendRef `json:"backendRefs,omitempty"`

	// BackendSettings holds configuration for managing the connection
	// to the backend.
	//
	// +optional
	BackendSettings *ClusterSettings `json:"backendSettings,omitempty"`
}

// ClusterSettings provides the various knobs that can be set to control how traffic to a given
// backend will be configured.
type ClusterSettings struct {
	// LoadBalancer policy to apply when routing traffic from the gateway to
	// the backend endpoints. Defaults to `LeastRequest`.
	// +optional
	LoadBalancer *LoadBalancer `json:"loadBalancer,omitempty"`

	// Retry provides more advanced usage, allowing users to customize the number of retries, retry fallback strategy, and retry triggering conditions.
	// If not set, retry will be disabled.
	// +optional
	Retry *Retry `json:"retry,omitempty"`

	// ProxyProtocol enables the Proxy Protocol when communicating with the backend.
	// +optional
	ProxyProtocol *ProxyProtocol `json:"proxyProtocol,omitempty"`

	// TcpKeepalive settings associated with the upstream client connection.
	// Disabled by default.
	//
	// +optional
	TCPKeepalive *TCPKeepalive `json:"tcpKeepalive,omitempty"`

	// HealthCheck allows gateway to perform active health checking on backends.
	//
	// +optional
	HealthCheck *HealthCheck `json:"healthCheck,omitempty"`

	// Circuit Breaker settings for the upstream connections and requests.
	// If not set, circuit breakers will be enabled with the default thresholds
	//
	// +optional
	CircuitBreaker *CircuitBreaker `json:"circuitBreaker,omitempty"`

	// Timeout settings for the backend connections.
	//
	// +optional
	Timeout *Timeout `json:"timeout,omitempty"`

	// Connection includes backend connection settings.
	//
	// +optional
	Connection *BackendConnection `json:"connection,omitempty"`

	// DNS includes dns resolution settings.
	//
	// +optional
	DNS *DNS `json:"dns,omitempty"`

	// HTTP2 provides HTTP/2 configuration for backend connections.
	//
	// +optional
	HTTP2 *HTTP2Settings `json:"http2,omitempty"`
}

// CIDR defines a CIDR Address range.
// A CIDR can be an IPv4 address range such as "192.168.1.0/24" or an IPv6 address range such as "2001:0db8:11a3:09d7::/64".
// +kubebuilder:validation:Pattern=`((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\/([0-9]+))|((([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))\/([0-9]+))`
type CIDR string

type InvalidMessageAction string

const (
	InvalidMessageActionTerminateConnection InvalidMessageAction = "TerminateConnection"
	InvalidMessageActionTerminateStream     InvalidMessageAction = "TerminateStream"
)

// HTTP2Settings provides HTTP/2 configuration for listeners and backends.
type HTTP2Settings struct {
	// InitialStreamWindowSize sets the initial window size for HTTP/2 streams.
	// If not set, the default value is 64 KiB(64*1024).
	//
	// +kubebuilder:validation:XIntOrString
	// +kubebuilder:validation:Pattern="^[1-9]+[0-9]*([EPTGMK]i|[EPTGMk])?$"
	// +optional
	InitialStreamWindowSize *resource.Quantity `json:"initialStreamWindowSize,omitempty"`

	// InitialConnectionWindowSize sets the initial window size for HTTP/2 connections.
	// If not set, the default value is 1 MiB.
	//
	// +kubebuilder:validation:XIntOrString
	// +kubebuilder:validation:Pattern="^[1-9]+[0-9]*([EPTGMK]i|[EPTGMk])?$"
	// +optional
	InitialConnectionWindowSize *resource.Quantity `json:"initialConnectionWindowSize,omitempty"`

	// MaxConcurrentStreams sets the maximum number of concurrent streams allowed per connection.
	// If not set, the default value is 100.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=2147483647
	// +optional
	MaxConcurrentStreams *uint32 `json:"maxConcurrentStreams,omitempty"`

	// OnInvalidMessage determines if Envoy will terminate the connection or just the offending stream in the event of HTTP messaging error
	// It's recommended for L2 Envoy deployments to set this value to TerminateStream.
	// https://www.envoyproxy.io/docs/envoy/latest/configuration/best_practices/level_two
	// Default: TerminateConnection
	// +optional
	OnInvalidMessage *InvalidMessageAction `json:"onInvalidMessage,omitempty"`
}

// ResponseOverride defines the configuration to override specific responses with a custom one.
type ResponseOverride struct {
	// Match configuration.
	Match CustomResponseMatch `json:"match"`
	// Response configuration.
	Response CustomResponse `json:"response"`
}

// CustomResponseMatch defines the configuration for matching a user response to return a custom one.
type CustomResponseMatch struct {
	// Status code to match on. The match evaluates to true if any of the matches are successful.
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=50
	StatusCodes []StatusCodeMatch `json:"statusCodes"`
}

// StatusCodeValueType defines the types of values for the status code match supported by Envoy Gateway.
// +kubebuilder:validation:Enum=Value;Range
type StatusCodeValueType string

const (
	// StatusCodeValueTypeValue defines the "Value" status code match type.
	StatusCodeValueTypeValue StatusCodeValueType = "Value"

	// StatusCodeValueTypeRange defines the "Range" status code match type.
	StatusCodeValueTypeRange StatusCodeValueType = "Range"
)

// StatusCodeMatch defines the configuration for matching a status code.
// +kubebuilder:validation:XValidation:message="value must be set for type Value",rule="(!has(self.type) || self.type == 'Value')? has(self.value) : true"
// +kubebuilder:validation:XValidation:message="range must be set for type Range",rule="(has(self.type) && self.type == 'Range')? has(self.range) : true"
type StatusCodeMatch struct {
	// Type is the type of value.
	// Valid values are Value and Range, default is Value.
	//
	// +kubebuilder:default=Value
	// +kubebuilder:validation:Enum=Value;Range
	// +unionDiscriminator
	Type *StatusCodeValueType `json:"type"`

	// Value contains the value of the status code.
	//
	// +optional
	Value *int `json:"value,omitempty"`

	// Range contains the range of status codes.
	//
	// +optional
	Range *StatusCodeRange `json:"range,omitempty"`
}

// StatusCodeRange defines the configuration for define a range of status codes.
// +kubebuilder:validation:XValidation: message="end must be greater than start",rule="self.end > self.start"
type StatusCodeRange struct {
	// Start of the range, including the start value.
	Start int `json:"start"`
	// End of the range, including the end value.
	End int `json:"end"`
}

// CustomResponse defines the configuration for returning a custom response.
type CustomResponse struct {
	// Content Type of the response. This will be set in the Content-Type header.
	//
	// +optional
	ContentType *string `json:"contentType,omitempty"`

	// Body of the Custom Response
	Body CustomResponseBody `json:"body"`
}

// ResponseValueType defines the types of values for the response body supported by Envoy Gateway.
// +kubebuilder:validation:Enum=Inline;ValueRef
type ResponseValueType string

const (
	// ResponseValueTypeInline defines the "Inline" response body type.
	ResponseValueTypeInline ResponseValueType = "Inline"

	// ResponseValueTypeValueRef defines the "ValueRef" response body type.
	ResponseValueTypeValueRef ResponseValueType = "ValueRef"
)

// CustomResponseBody
// +kubebuilder:validation:XValidation:message="inline must be set for type Inline",rule="(!has(self.type) || self.type == 'Inline')? has(self.inline) : true"
// +kubebuilder:validation:XValidation:message="valueRef must be set for type ValueRef",rule="(has(self.type) && self.type == 'ValueRef')? has(self.valueRef) : true"
// +kubebuilder:validation:XValidation:message="only ConfigMap is supported for ValueRef",rule="has(self.valueRef) ? self.valueRef.kind == 'ConfigMap' : true"
type CustomResponseBody struct {
	// Type is the type of method to use to read the body value.
	// Valid values are Inline and ValueRef, default is Inline.
	//
	// +kubebuilder:default=Inline
	// +kubebuilder:validation:Enum=Inline;ValueRef
	// +unionDiscriminator
	Type *ResponseValueType `json:"type"`

	// Inline contains the value as an inline string.
	//
	// +optional
	Inline *string `json:"inline,omitempty"`

	// ValueRef contains the contents of the body
	// specified as a local object reference.
	// Only a reference to ConfigMap is supported.
	//
	// The value of key `response.body` in the ConfigMap will be used as the response body.
	// If the key is not found, the first value in the ConfigMap will be used.
	//
	// +optional
	ValueRef *gwapiv1.LocalObjectReference `json:"valueRef,omitempty"`
}
