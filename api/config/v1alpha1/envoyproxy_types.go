// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// KindEnvoyProxy is the name of the EnvoyProxy kind.
	KindEnvoyProxy = "EnvoyProxy"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

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

	// Logging defines logging parameters for managed proxies. If unspecified,
	// default settings apply. This type is not implemented until
	// https://github.com/envoyproxy/gateway/issues/280 is fixed.
	//
	// +kubebuilder:default={level: {system: info}}
	Logging ProxyLogging `json:"logging,omitempty"`

	// AccessLoggings defines access logging parameters for managed proxies.
	// If unspecified, access log is disabled.
	AccessLoggings []ProxyAccessLogging `json:"accessLoggings,omitempty"`

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
	Bootstrap *string `json:"bootstrap,omitempty"`
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

// EnvoyProxyKubernetesProvider defines configuration for the Kubernetes resource
// provider.
type EnvoyProxyKubernetesProvider struct {
	// EnvoyDeployment defines the desired state of the Envoy deployment resource.
	// If unspecified, default settings for the manged Envoy deployment resource
	// are applied.
	//
	// +optional
	EnvoyDeployment *KubernetesDeploymentSpec `json:"envoyDeployment,omitempty"`

	// EnvoyService defines the desired state of the Envoy service resource.
	// If unspecified, default settings for the manged Envoy service resource
	// are applied.
	//
	// +optional
	EnvoyService *KubernetesServiceSpec `json:"envoyService,omitempty"`
}

// ProxyLogging defines logging parameters for managed proxies. This type is not
// implemented until https://github.com/envoyproxy/gateway/issues/280 is fixed.
type ProxyLogging struct {
	// Level is a map of logging level per component, where the component is the key
	// and the log level is the value. If unspecified, defaults to "System: Info".
	//
	// +kubebuilder:default={system: info}
	Level map[LogComponent]LogLevel `json:"level,omitempty"`
}

type ProxyAccessLogging struct {
	// Format defines the format of access logging.
	Format ProxyAccessLoggingFormat `json:"format"`
	// Sink defines the sink of access logging.
	Sink ProxyAccessLoggingSink `json:"sink"`
}

type ProxyAccessLoggingFormatType string

const (
	// ProxyAccessLoggingFormatTypeText defines the text access logging format.
	ProxyAccessLoggingFormatTypeText ProxyAccessLoggingFormatType = "Text"
	// ProxyAccessLoggingFormatTypeJSON defines the JSON access logging format.
	ProxyAccessLoggingFormatTypeJSON ProxyAccessLoggingFormatType = "JSON"
	// TODO: support format type "mix" in the future.
)

type ProxyAccessLoggingFormat struct {
	// Type defines the type of access logging format.
	// +kubebuilder:validation:Enum=Text;JSON
	Type ProxyAccessLoggingFormatType `json:"type,omitempty"`
	// Text defines the text access logging format, following Envoy access logging formatting,
	// empty value results in proxy's default access log format.
	// It's required when the format type is "text".
	// Envoy [command operators](https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#command-operators) may be used in the format.
	// The [format string documentation](https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#config-access-log-format-strings) provides more information.
	// +optional
	Text *string `json:"text,omitempty"`
	// Fields is additional attributes that describe the specific event occurrence.
	// Structured format for the envoy access logs. Envoy [command operators](https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#command-operators)
	// can be used as values for fields within the Struct.
	// It's required when the format type is "json".
	// +optional
	Fields *map[string]string `json:"fields,omitempty"`
}

type ProxyAccessLoggingSinkType string

const (
	// ProxyAccessLoggingSinkTypeFile defines the file access logging sink.
	ProxyAccessLoggingSinkTypeFile ProxyAccessLoggingSinkType = "file"
	// ProxyAccessLoggingSinkTypeOpenTelemetry defines the OpenTelemetry access logging sink.
	ProxyAccessLoggingSinkTypeOpenTelemetry ProxyAccessLoggingSinkType = "Opentelemetry"
)

type ProxyAccessLoggingSink struct {
	// Type defines the type of access logging sink.
	// +kubebuilder:validation:Enum=file;opentelemetry
	Type ProxyAccessLoggingSinkType `json:"type,omitempty"`
	// File defines the file access logging sink.
	// +optional
	File *FileEnvoyProxyAccessLogging `json:"file,omitempty"`
	// OpenTelemetry defines the OpenTelemetry access logging sink.
	// +optional
	OpenTelemetry *OpenTelemetryEnvoyProxyAccessLogging `json:"opentelemetry,omitempty"`
}

type FileEnvoyProxyAccessLogging struct {
	// Path defines the file path used to expose envoy access log(e.g. /dev/stdout).
	// Empty value disables access logging.
	Path string `json:"path,omitempty"`
}

// TODO: consider reuse ExtensionService?
type OpenTelemetryEnvoyProxyAccessLogging struct {
	// Host define the extension service hostname.
	Host string `json:"host"`
	// Port defines the port the extension service is exposed on.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=4317
	Port int32 `json:"port,omitempty"`
	// Resources is a set of labels that describe the source of a log entry, including envoy node info.
	// It's recommended to follow [semantic conventions](https://opentelemetry.io/docs/reference/specification/resource/semantic_conventions/).
	// +optional
	Resources map[string]string `json:"resources,omitempty"`

	// TODO: support more OpenTelemetry access logging options(e.g. TLS, auth etc.) in the future.
}

// LogComponent defines a component that supports a configured logging level.
// This type is not implemented until https://github.com/envoyproxy/gateway/issues/280
// is fixed.
// +kubebuilder:validation:Enum=system;upstream;http;connection;admin;client;filter;main;router;runtime
type LogComponent string

const (
	// LogComponentSystem defines the "system"-wide logging component. When specified,
	// all other logging components are ignored.
	LogComponentSystem LogComponent = "system"

	// LogComponentUpstream defines defines the "upstream" logging component.
	LogComponentUpstream LogComponent = "upstream"

	// LogComponentHTTP defines defines the "http" logging component.
	LogComponentHTTP LogComponent = "http"

	// LogComponentConnection defines defines the "connection" logging component.
	LogComponentConnection LogComponent = "connection"

	// LogComponentAdmin defines defines the "admin" logging component.
	LogComponentAdmin LogComponent = "admin"

	// LogComponentClient defines defines the "client" logging component.
	LogComponentClient LogComponent = "client"

	// LogComponentFilter defines defines the "filter" logging component.
	LogComponentFilter LogComponent = "filter"

	// LogComponentMain defines defines the "main" logging component.
	LogComponentMain LogComponent = "main"

	// LogComponentRouter defines defines the "router" logging component.
	LogComponentRouter LogComponent = "router"

	// LogComponentRuntime defines defines the "runtime" logging component.
	LogComponentRuntime LogComponent = "runtime"
)

// LogLevel defines a log level for system logs. This type is not implemented until
// https://github.com/envoyproxy/gateway/issues/280 is fixed.
// +kubebuilder:validation:Enum=debug;info;error
type LogLevel string

const (
	// LogLevelDebug defines the "debug" logging level.
	LogLevelDebug LogLevel = "debug"

	// LogLevelInfo defines the "Info" logging level.
	LogLevelInfo LogLevel = "info"

	// LogLevelError defines the "Error" logging level.
	LogLevelError LogLevel = "error"
)

// EnvoyProxyStatus defines the observed state of EnvoyProxy. This type is not implemented
// until https://github.com/envoyproxy/gateway/issues/1007 is fixed.
type EnvoyProxyStatus struct {
	// INSERT ADDITIONAL STATUS FIELDS - define observed state of cluster.
	// Important: Run "make" to regenerate code after modifying this file.
}

//+kubebuilder:object:root=true

// EnvoyProxyList contains a list of EnvoyProxy
type EnvoyProxyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EnvoyProxy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EnvoyProxy{}, &EnvoyProxyList{})
}
