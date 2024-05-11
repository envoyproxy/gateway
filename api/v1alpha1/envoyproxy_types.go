// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

const (
	// KindEnvoyProxy is the name of the EnvoyProxy kind.
	KindEnvoyProxy = "EnvoyProxy"
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories=envoy-gateway,shortName=eproxy
// +kubebuilder:subresource:status

// EnvoyProxy is the schema for the envoyproxies API.
type EnvoyProxy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// EnvoyProxySpec defines the desired state of EnvoyProxy.
	Spec EnvoyProxySpec `json:"spec,omitempty"`
	// EnvoyProxyStatus defines the actual state of EnvoyProxy.
	Status EnvoyProxyStatus `json:"status,omitempty"`
}

// EnvoyProxySpec defines the desired state of EnvoyProxy.
type EnvoyProxySpec struct {
	// Provider defines the desired resource provider and provider-specific configuration.
	// If unspecified, the "Kubernetes" resource provider is used with default configuration
	// parameters.
	//
	// +optional
	Provider *EnvoyProxyProvider `json:"provider,omitempty"`

	// Logging defines logging parameters for managed proxies.
	// +kubebuilder:default={level: {default: warn}}
	Logging ProxyLogging `json:"logging,omitempty"`

	// Telemetry defines telemetry parameters for managed proxies.
	//
	// +optional
	Telemetry *ProxyTelemetry `json:"telemetry,omitempty"`

	// Bootstrap defines the Envoy Bootstrap as a YAML string.
	// Visit https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/bootstrap/v3/bootstrap.proto#envoy-v3-api-msg-config-bootstrap-v3-bootstrap
	// to learn more about the syntax.
	// If set, this is the Bootstrap configuration used for the managed Envoy Proxy fleet instead of the default Bootstrap configuration
	// set by Envoy Gateway.
	// Some fields within the Bootstrap that are required to communicate with the xDS Server (Envoy Gateway) and receive xDS resources
	// from it are not configurable and will result in the `EnvoyProxy` resource being rejected.
	// Backward compatibility across minor versions is not guaranteed.
	// We strongly recommend using `egctl x translate` to generate a `EnvoyProxy` resource with the `Bootstrap` field set to the default
	// Bootstrap configuration used. You can edit this configuration, and rerun `egctl x translate` to ensure there are no validation errors.
	//
	// +optional
	Bootstrap *ProxyBootstrap `json:"bootstrap,omitempty"`

	// Concurrency defines the number of worker threads to run. If unset, it defaults to
	// the number of cpuset threads on the platform.
	//
	// +optional
	Concurrency *int32 `json:"concurrency,omitempty"`

	// ExtraArgs defines additional command line options that are provided to Envoy.
	// More info: https://www.envoyproxy.io/docs/envoy/latest/operations/cli#command-line-options
	// Note: some command line options are used internally(e.g. --log-level) so they cannot be provided here.
	//
	// +optional
	ExtraArgs []string `json:"extraArgs,omitempty"`

	// MergeGateways defines if Gateway resources should be merged onto the same Envoy Proxy Infrastructure.
	// Setting this field to true would merge all Gateway Listeners under the parent Gateway Class.
	// This means that the port, protocol and hostname tuple must be unique for every listener.
	// If a duplicate listener is detected, the newer listener (based on timestamp) will be rejected and its status will be updated with a "Accepted=False" condition.
	//
	// +optional
	MergeGateways *bool `json:"mergeGateways,omitempty"`

	// Shutdown defines configuration for graceful envoy shutdown process.
	//
	// +optional
	Shutdown *ShutdownConfig `json:"shutdown,omitempty"`

	// FilterOrder defines the order of filters in the Envoy proxy's HTTP filter chain.
	// The FilterPosition in the list will be applied in the order they are defined.
	// If unspecified, the default filter order is applied.
	// Default filter order is:
	//
	// - envoy.filters.http.fault
	//
	// - envoy.filters.http.cors
	//
	// - envoy.filters.http.ext_authz
	//
	// - envoy.filters.http.basic_authn
	//
	// - envoy.filters.http.oauth2
	//
	// - envoy.filters.http.jwt_authn
	//
	// - envoy.filters.http.ext_proc
	//
	// - envoy.filters.http.wasm
	//
	// - envoy.filters.http.local_ratelimit
	//
	// - envoy.filters.http.ratelimit
	//
	// - envoy.filters.http.router
	//
	// +optional
	FilterOrder []FilterPosition `json:"filterOrder,omitempty"`
	// BackendTLS is the TLS configuration for the Envoy proxy to use when connecting to backends.
	// These settings are applied on backends for which TLS policies are specified.
	// +optional
	BackendTLS *BackendTLSConfig `json:"backendTLS,omitempty"`
}

// BackendTLSConfig describes the BackendTLS configuration for Envoy Proxy.
type BackendTLSConfig struct {
	// ClientCertificateRef defines the reference to a Kubernetes Secret that contains
	// the client certificate and private key for Envoy to use when connecting to
	// backend services and external services, such as ExtAuth, ALS, OpenTelemetry, etc.
	// +optional
	ClientCertificateRef *gwapiv1.SecretObjectReference `json:"clientCertificateRef,omitempty"`
	TLSSettings          `json:",inline"`
}

// FilterPosition defines the position of an Envoy HTTP filter in the filter chain.
// +kubebuilder:validation:XValidation:rule="(has(self.before) || has(self.after))",message="one of before or after must be specified"
// +kubebuilder:validation:XValidation:rule="(has(self.before) && !has(self.after)) || (!has(self.before) && has(self.after))",message="only one of before or after can be specified"
type FilterPosition struct {
	// Name of the filter.
	Name EnvoyFilter `json:"name"`

	// Before defines the filter that should come before the filter.
	// Only one of Before or After must be set.
	Before *EnvoyFilter `json:"before,omitempty"`

	// After defines the filter that should come after the filter.
	// Only one of Before or After must be set.
	After *EnvoyFilter `json:"after,omitempty"`
}

// EnvoyFilter defines the type of Envoy HTTP filter.
// +kubebuilder:validation:Enum=envoy.filters.http.cors;envoy.filters.http.ext_authz;envoy.filters.http.basic_authn;envoy.filters.http.oauth2;envoy.filters.http.jwt_authn;envoy.filters.http.fault;envoy.filters.http.local_ratelimit;envoy.filters.http.ratelimit;envoy.filters.http.wasm;envoy.filters.http.ext_proc
type EnvoyFilter string

const (
	// EnvoyFilterFault defines the Envoy HTTP fault filter.
	EnvoyFilterFault EnvoyFilter = "envoy.filters.http.fault"
	// EnvoyFilterCORS defines the Envoy HTTP CORS filter.
	EnvoyFilterCORS EnvoyFilter = "envoy.filters.http.cors"

	// EnvoyFilterExtAuthz defines the Envoy HTTP external authorization filter.
	EnvoyFilterExtAuthz EnvoyFilter = "envoy.filters.http.ext_authz"

	// EnvoyFilterBasicAuthn defines the Envoy HTTP basic authentication filter.
	EnvoyFilterBasicAuthn EnvoyFilter = "envoy.filters.http.basic_authn"

	// EnvoyFilterOAuth2 defines the Envoy HTTP OAuth2 filter.
	EnvoyFilterOAuth2 EnvoyFilter = "envoy.filters.http.oauth2"

	// EnvoyFilterJWTAuthn defines the Envoy HTTP JWT authentication filter.
	EnvoyFilterJWTAuthn EnvoyFilter = "envoy.filters.http.jwt_authn"

	// EnvoyFilterExtProc defines the Envoy HTTP external process filter.
	EnvoyFilterExtProc EnvoyFilter = "envoy.filters.http.ext_proc"

	// EnvoyFilterWasm defines the Envoy HTTP WebAssembly filter.
	EnvoyFilterWasm EnvoyFilter = "envoy.filters.http.wasm"

	// EnvoyFilterLocalRateLimit defines the Envoy HTTP local rate limit filter.
	EnvoyFilterLocalRateLimit EnvoyFilter = "envoy.filters.http.local_ratelimit"

	// EnvoyFilterRateLimit defines the Envoy HTTP rate limit filter.
	EnvoyFilterRateLimit EnvoyFilter = "envoy.filters.http.ratelimit"

	// EnvoyFilterRouter defines the Envoy HTTP router filter.
	EnvoyFilterRouter EnvoyFilter = "envoy.filters.http.router"
)

type ProxyTelemetry struct {
	// AccessLogs defines accesslog parameters for managed proxies.
	// If unspecified, will send default format to stdout.
	// +optional
	AccessLog *ProxyAccessLog `json:"accessLog,omitempty"`
	// Tracing defines tracing configuration for managed proxies.
	// If unspecified, will not send tracing data.
	// +optional
	Tracing *ProxyTracing `json:"tracing,omitempty"`

	// Metrics defines metrics configuration for managed proxies.
	Metrics *ProxyMetrics `json:"metrics,omitempty"`
}

// EnvoyProxyProvider defines the desired state of a resource provider.
// +union
type EnvoyProxyProvider struct {
	// Type is the type of resource provider to use. A resource provider provides
	// infrastructure resources for running the data plane, e.g. Envoy proxy, and
	// optional auxiliary control planes. Supported types are "Kubernetes".
	//
	// +unionDiscriminator
	Type ProviderType `json:"type"`
	// Kubernetes defines the desired state of the Kubernetes resource provider.
	// Kubernetes provides infrastructure resources for running the data plane,
	// e.g. Envoy proxy. If unspecified and type is "Kubernetes", default settings
	// for managed Kubernetes resources are applied.
	//
	// +optional
	Kubernetes *EnvoyProxyKubernetesProvider `json:"kubernetes,omitempty"`
}

// ShutdownConfig defines configuration for graceful envoy shutdown process.
type ShutdownConfig struct {
	// DrainTimeout defines the graceful drain timeout. This should be less than the pod's terminationGracePeriodSeconds.
	// If unspecified, defaults to 600 seconds.
	//
	// +optional
	DrainTimeout *metav1.Duration `json:"drainTimeout,omitempty"`
	// MinDrainDuration defines the minimum drain duration allowing time for endpoint deprogramming to complete.
	// If unspecified, defaults to 5 seconds.
	//
	// +optional
	MinDrainDuration *metav1.Duration `json:"minDrainDuration,omitempty"`
}

// +kubebuilder:validation:XValidation:rule="((has(self.envoyDeployment) && !has(self.envoyDaemonSet)) || (!has(self.envoyDeployment) && has(self.envoyDaemonSet))) || (!has(self.envoyDeployment) && !has(self.envoyDaemonSet))",message="only one of envoyDeployment or envoyDaemonSet can be specified"
// +kubebuilder:validation:XValidation:rule="((has(self.envoyHpa) && !has(self.envoyDaemonSet)) || (!has(self.envoyHpa) && has(self.envoyDaemonSet))) || (!has(self.envoyHpa) && !has(self.envoyDaemonSet))",message="cannot use envoyHpa if envoyDaemonSet is used"
//
// EnvoyProxyKubernetesProvider defines configuration for the Kubernetes resource
// provider.
type EnvoyProxyKubernetesProvider struct {
	// EnvoyDeployment defines the desired state of the Envoy deployment resource.
	// If unspecified, default settings for the managed Envoy deployment resource
	// are applied.
	//
	// +optional
	EnvoyDeployment *KubernetesDeploymentSpec `json:"envoyDeployment,omitempty"`

	// EnvoyDaemonSet defines the desired state of the Envoy daemonset resource.
	// Disabled by default, a deployment resource is used instead to provision the Envoy Proxy fleet
	//
	// +optional
	EnvoyDaemonSet *KubernetesDaemonSetSpec `json:"envoyDaemonSet,omitempty"`

	// EnvoyService defines the desired state of the Envoy service resource.
	// If unspecified, default settings for the managed Envoy service resource
	// are applied.
	//
	// +optional
	EnvoyService *KubernetesServiceSpec `json:"envoyService,omitempty"`

	// EnvoyHpa defines the Horizontal Pod Autoscaler settings for Envoy Proxy Deployment.
	// Once the HPA is being set, Replicas field from EnvoyDeployment will be ignored.
	//
	// +optional
	EnvoyHpa *KubernetesHorizontalPodAutoscalerSpec `json:"envoyHpa,omitempty"`
}

// ProxyLogging defines logging parameters for managed proxies.
type ProxyLogging struct {
	// Level is a map of logging level per component, where the component is the key
	// and the log level is the value. If unspecified, defaults to "default: warn".
	//
	// +kubebuilder:default={default: warn}
	Level map[ProxyLogComponent]LogLevel `json:"level,omitempty"`
}

// ProxyLogComponent defines a component that supports a configured logging level.
// +kubebuilder:validation:Enum=system;upstream;http;connection;admin;client;filter;main;router;runtime
type ProxyLogComponent string

const (
	// LogComponentDefault defines the default logging component.
	// See more details: https://www.envoyproxy.io/docs/envoy/latest/operations/cli#cmdoption-l
	LogComponentDefault ProxyLogComponent = "default"

	// LogComponentUpstream defines the "upstream" logging component.
	LogComponentUpstream ProxyLogComponent = "upstream"

	// LogComponentHTTP defines the "http" logging component.
	LogComponentHTTP ProxyLogComponent = "http"

	// LogComponentConnection defines the "connection" logging component.
	LogComponentConnection ProxyLogComponent = "connection"

	// LogComponentAdmin defines the "admin" logging component.
	LogComponentAdmin ProxyLogComponent = "admin"

	// LogComponentClient defines the "client" logging component.
	LogComponentClient ProxyLogComponent = "client"

	// LogComponentFilter defines the "filter" logging component.
	LogComponentFilter ProxyLogComponent = "filter"

	// LogComponentMain defines the "main" logging component.
	LogComponentMain ProxyLogComponent = "main"

	// LogComponentRouter defines the "router" logging component.
	LogComponentRouter ProxyLogComponent = "router"

	// LogComponentRuntime defines the "runtime" logging component.
	LogComponentRuntime ProxyLogComponent = "runtime"
)

// ProxyBootstrap defines Envoy Bootstrap configuration.
type ProxyBootstrap struct {
	// Type is the type of the bootstrap configuration, it should be either Replace or Merge.
	// If unspecified, it defaults to Replace.
	// +optional
	// +kubebuilder:default=Replace
	Type *BootstrapType `json:"type"`

	// Value is a YAML string of the bootstrap.
	Value string `json:"value"`
}

// BootstrapType defines the types of bootstrap supported by Envoy Gateway.
// +kubebuilder:validation:Enum=Merge;Replace
type BootstrapType string

const (
	// Merge merges the provided bootstrap with the default one. The provided bootstrap can add or override a value
	// within a map, or add a new value to a list.
	// Please note that the provided bootstrap can't override a value within a list.
	BootstrapTypeMerge BootstrapType = "Merge"

	// Replace replaces the default bootstrap with the provided one.
	BootstrapTypeReplace BootstrapType = "Replace"
)

// EnvoyProxyStatus defines the observed state of EnvoyProxy. This type is not implemented
// until https://github.com/envoyproxy/gateway/issues/1007 is fixed.
type EnvoyProxyStatus struct {
	// INSERT ADDITIONAL STATUS FIELDS - define observed state of cluster.
	// Important: Run "make" to regenerate code after modifying this file.
}

// +kubebuilder:object:root=true

// EnvoyProxyList contains a list of EnvoyProxy
type EnvoyProxyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EnvoyProxy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EnvoyProxy{}, &EnvoyProxyList{})
}
