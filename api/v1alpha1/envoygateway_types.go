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
	// KindEnvoyGateway is the name of the EnvoyGateway kind.
	KindEnvoyGateway = "EnvoyGateway"
	// GatewayControllerName is the name of the GatewayClass controller.
	GatewayControllerName = "gateway.envoyproxy.io/gatewayclass-controller"
	// GatewayAdminPort is the port which envoy gateway admin server is listening on.
	GatewayAdminPort = 19000
	// GatewayAdminHost is the host of envoy gateway admin server.
	GatewayAdminHost = "127.0.0.1"
	// GatewayMetricsPort is the port which envoy gateway metrics server is listening on.
	GatewayMetricsPort = 19001
	// GatewayMetricsHost is the host of envoy gateway metrics server.
	GatewayMetricsHost = "0.0.0.0"
)

// +kubebuilder:object:root=true

// EnvoyGateway is the schema for the envoygateways API.
type EnvoyGateway struct {
	metav1.TypeMeta `json:",inline"`

	// EnvoyGatewaySpec defines the desired state of EnvoyGateway.
	EnvoyGatewaySpec `json:",inline"`
}

// EnvoyGatewaySpec defines the desired state of Envoy Gateway.
type EnvoyGatewaySpec struct {
	// Gateway defines desired Gateway API specific configuration. If unset,
	// default configuration parameters will apply.
	//
	// +optional
	Gateway *Gateway `json:"gateway,omitempty"`

	// Provider defines the desired provider and provider-specific configuration.
	// If unspecified, the Kubernetes provider is used with default configuration
	// parameters.
	//
	// +optional
	Provider *EnvoyGatewayProvider `json:"provider,omitempty"`

	// Logging defines logging parameters for Envoy Gateway.
	//
	// +optional
	// +kubebuilder:default={default: info}
	Logging *EnvoyGatewayLogging `json:"logging,omitempty"`

	// Admin defines the desired admin related abilities.
	// If unspecified, the Admin is used with default configuration
	// parameters.
	//
	// +optional
	Admin *EnvoyGatewayAdmin `json:"admin,omitempty"`

	// Telemetry defines the desired control plane telemetry related abilities.
	// If unspecified, the telemetry is used with default configuration.
	//
	// +optional
	Telemetry *EnvoyGatewayTelemetry `json:"telemetry,omitempty"`

	// RateLimit defines the configuration associated with the Rate Limit service
	// deployed by Envoy Gateway required to implement the Global Rate limiting
	// functionality. The specific rate limit service used here is the reference
	// implementation in Envoy. For more details visit https://github.com/envoyproxy/ratelimit.
	// This configuration is unneeded for "Local" rate limiting.
	//
	// +optional
	RateLimit *RateLimit `json:"rateLimit,omitempty"`

	// ExtensionManager defines an extension manager to register for the Envoy Gateway Control Plane.
	//
	// +optional
	ExtensionManager *ExtensionManager `json:"extensionManager,omitempty"`

	// ExtensionAPIs defines the settings related to specific Gateway API Extensions
	// implemented by Envoy Gateway
	//
	// +optional
	ExtensionAPIs *ExtensionAPISettings `json:"extensionApis,omitempty"`
}

// EnvoyGatewayTelemetry defines telemetry configurations for envoy gateway control plane.
// Control plane will focus on metrics observability telemetry and tracing telemetry later.
type EnvoyGatewayTelemetry struct {
	// Metrics defines metrics configuration for envoy gateway.
	Metrics *EnvoyGatewayMetrics `json:"metrics,omitempty"`
}

// EnvoyGatewayLogging defines logging for Envoy Gateway.
type EnvoyGatewayLogging struct {
	// Level is the logging level. If unspecified, defaults to "info".
	// EnvoyGatewayLogComponent options: default/provider/gateway-api/xds-translator/xds-server/infrastructure/global-ratelimit.
	// LogLevel options: debug/info/error/warn.
	//
	// +kubebuilder:default={default: info}
	Level map[EnvoyGatewayLogComponent]LogLevel `json:"level,omitempty"`
}

// EnvoyGatewayLogComponent defines a component that supports a configured logging level.
// +kubebuilder:validation:Enum=default;provider;gateway-api;xds-translator;xds-server;infrastructure;global-ratelimit
type EnvoyGatewayLogComponent string

const (
	// LogComponentGatewayDefault defines the "default"-wide logging component. When specified,
	// all other logging components are ignored.
	LogComponentGatewayDefault EnvoyGatewayLogComponent = "default"

	// LogComponentProviderRunner defines the "provider" runner component.
	LogComponentProviderRunner EnvoyGatewayLogComponent = "provider"

	// LogComponentGatewayAPIRunner defines the "gateway-api" runner component.
	LogComponentGatewayAPIRunner EnvoyGatewayLogComponent = "gateway-api"

	// LogComponentXdsTranslatorRunner defines the "xds-translator" runner component.
	LogComponentXdsTranslatorRunner EnvoyGatewayLogComponent = "xds-translator"

	// LogComponentXdsServerRunner defines the "xds-server" runner component.
	LogComponentXdsServerRunner EnvoyGatewayLogComponent = "xds-server"

	// LogComponentInfrastructureRunner defines the "infrastructure" runner component.
	LogComponentInfrastructureRunner EnvoyGatewayLogComponent = "infrastructure"

	// LogComponentGlobalRateLimitRunner defines the "global-ratelimit" runner component.
	LogComponentGlobalRateLimitRunner EnvoyGatewayLogComponent = "global-ratelimit"
)

// Gateway defines the desired Gateway API configuration of Envoy Gateway.
type Gateway struct {
	// ControllerName defines the name of the Gateway API controller. If unspecified,
	// defaults to "gateway.envoyproxy.io/gatewayclass-controller". See the following
	// for additional details:
	//   https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.GatewayClass
	//
	// +optional
	ControllerName string `json:"controllerName,omitempty"`
}

// ExtensionAPISettings defines the settings specific to Gateway API Extensions.
type ExtensionAPISettings struct {
	// EnableEnvoyPatchPolicy enables Envoy Gateway to
	// reconcile and implement the EnvoyPatchPolicy resources.
	EnableEnvoyPatchPolicy bool `json:"enableEnvoyPatchPolicy"`

	// EnableEnvoyExtensionPolicy enables Envoy Gateway to
	// reconcile and implement the EnvoyExtensionPolicy resources.
	EnableEnvoyExtensionPolicy bool `json:"enableEnvoyExtensionPolicy"`
}

// EnvoyGatewayProvider defines the desired configuration of a provider.
// +union
type EnvoyGatewayProvider struct {
	// Type is the type of provider to use. Supported types are "Kubernetes".
	//
	// +unionDiscriminator
	Type ProviderType `json:"type"`

	// Kubernetes defines the configuration of the Kubernetes provider. Kubernetes
	// provides runtime configuration via the Kubernetes API.
	//
	// +optional
	Kubernetes *EnvoyGatewayKubernetesProvider `json:"kubernetes,omitempty"`

	// Custom defines the configuration for the Custom provider. This provider
	// allows you to define a specific resource provider and a infrastructure
	// provider.
	//
	// +optional
	Custom *EnvoyGatewayCustomProvider `json:"custom,omitempty"`
}

// EnvoyGatewayKubernetesProvider defines configuration for the Kubernetes provider.
type EnvoyGatewayKubernetesProvider struct {
	// RateLimitDeployment defines the desired state of the Envoy ratelimit deployment resource.
	// If unspecified, default settings for the managed Envoy ratelimit deployment resource
	// are applied.
	//
	// +optional
	RateLimitDeployment *KubernetesDeploymentSpec `json:"rateLimitDeployment,omitempty"`

	// Watch holds configuration of which input resources should be watched and reconciled.
	// +optional
	Watch *KubernetesWatchMode `json:"watch,omitempty"`
	// Deploy holds configuration of how output managed resources such as the Envoy Proxy data plane
	// should be deployed
	// +optional
	Deploy *KubernetesDeployMode `json:"deploy,omitempty"`
	// OverwriteControlPlaneCerts updates the secrets containing the control plane certs, when set.
	// +optional
	OverwriteControlPlaneCerts *bool `json:"overwriteControlPlaneCerts,omitempty"`
}

const (
	// KubernetesWatchModeTypeNamespaces indicates that the namespace watch mode is used.
	KubernetesWatchModeTypeNamespaces = "Namespaces"

	// KubernetesWatchModeTypeNamespaceSelector indicates that namespaceSelector watch
	// mode is used.
	KubernetesWatchModeTypeNamespaceSelector = "NamespaceSelector"
)

// KubernetesWatchModeType defines the type of KubernetesWatchMode
type KubernetesWatchModeType string

// KubernetesWatchMode holds the configuration for which input resources to watch and reconcile.
type KubernetesWatchMode struct {
	// Type indicates what watch mode to use. KubernetesWatchModeTypeNamespaces and
	// KubernetesWatchModeTypeNamespaceSelector are currently supported
	// By default, when this field is unset or empty, Envoy Gateway will watch for input namespaced resources
	// from all namespaces.
	Type KubernetesWatchModeType `json:"type,omitempty"`

	// Namespaces holds the list of namespaces that Envoy Gateway will watch for namespaced scoped
	// resources such as Gateway, HTTPRoute and Service.
	// Note that Envoy Gateway will continue to reconcile relevant cluster scoped resources such as
	// GatewayClass that it is linked to. Precisely one of Namespaces and NamespaceSelector must be set.
	Namespaces []string `json:"namespaces,omitempty"`

	// NamespaceSelector holds the label selector used to dynamically select namespaces.
	// Envoy Gateway will watch for namespaces matching the specified label selector.
	// Precisely one of Namespaces and NamespaceSelector must be set.
	NamespaceSelector *metav1.LabelSelector `json:"namespaceSelector,omitempty"`
}

// KubernetesDeployMode holds configuration for how to deploy managed resources such as the Envoy Proxy
// data plane fleet.
type KubernetesDeployMode struct {
	// TODO
}

// EnvoyGatewayCustomProvider defines configuration for the Custom provider.
type EnvoyGatewayCustomProvider struct {
	// Resource defines the desired resource provider.
	// This provider is used to specify the provider to be used
	// to retrieve the resource configurations such as Gateway API
	// resources
	Resource EnvoyGatewayResourceProvider `json:"resource"`
	// Infrastructure defines the desired infrastructure provider.
	// This provider is used to specify the provider to be used
	// to provide an environment to deploy the out resources like
	// the Envoy Proxy data plane.
	Infrastructure EnvoyGatewayInfrastructureProvider `json:"infrastructure"`
}

// ResourceProviderType defines the types of custom resource providers supported by Envoy Gateway.
//
// +kubebuilder:validation:Enum=File
type ResourceProviderType string

const (
	// ResourceProviderTypeFile defines the "File" provider.
	ResourceProviderTypeFile ResourceProviderType = "File"
)

// EnvoyGatewayResourceProvider defines configuration for the Custom Resource provider.
type EnvoyGatewayResourceProvider struct {
	// Type is the type of resource provider to use. Supported types are "File".
	//
	// +unionDiscriminator
	Type ResourceProviderType `json:"type"`
	// File defines the configuration of the File provider. File provides runtime
	// configuration defined by one or more files.
	//
	// +optional
	File *EnvoyGatewayFileResourceProvider `json:"file,omitempty"`
}

// EnvoyGatewayFileResourceProvider defines configuration for the File Resource provider.
type EnvoyGatewayFileResourceProvider struct {
	// Paths are the paths to a directory or file containing the resource configuration.
	// Recursive sub directories are not currently supported.
	Paths []string `json:"paths"`
}

// InfrastructureProviderType defines the types of custom infrastructure providers supported by Envoy Gateway.
//
// +kubebuilder:validation:Enum=Host
type InfrastructureProviderType string

const (
	// InfrastructureProviderTypeHost defines the "Host" provider.
	InfrastructureProviderTypeHost InfrastructureProviderType = "Host"
)

// EnvoyGatewayInfrastructureProvider defines configuration for the Custom Infrastructure provider.
type EnvoyGatewayInfrastructureProvider struct {
	// Type is the type of infrastructure providers to use. Supported types are "Host".
	//
	// +unionDiscriminator
	Type InfrastructureProviderType `json:"type"`
	// Host defines the configuration of the Host provider. Host provides runtime
	// deployment of the data plane as a child process on the host environment.
	//
	// +optional
	Host *EnvoyGatewayHostInfrastructureProvider `json:"host,omitempty"`
}

// EnvoyGatewayHostInfrastructureProvider defines configuration for the Host Infrastructure provider.
type EnvoyGatewayHostInfrastructureProvider struct {
	// TODO: Add config as use cases are better understood.
}

// RateLimit defines the configuration associated with the Rate Limit Service
// used for Global Rate Limiting.
type RateLimit struct {
	// Backend holds the configuration associated with the
	// database backend used by the rate limit service to store
	// state associated with global ratelimiting.
	Backend RateLimitDatabaseBackend `json:"backend"`

	// Timeout specifies the timeout period for the proxy to access the ratelimit server
	// If not set, timeout is 20ms.
	// +optional
	// +kubebuilder:validation:Format=duration
	Timeout *metav1.Duration `json:"timeout,omitempty"`

	// FailClosed is a switch used to control the flow of traffic
	// when the response from the ratelimit server cannot be obtained.
	// If FailClosed is false, let the traffic pass,
	// otherwise, don't let the traffic pass and return 500.
	// If not set, FailClosed is False.
	FailClosed bool `json:"failClosed"`

	// Telemetry defines telemetry configuration for RateLimit.
	// +optional
	Telemetry *RateLimitTelemetry `json:"telemetry,omitempty"`
}

type RateLimitTelemetry struct {
	// Metrics defines metrics configuration for RateLimit.
	Metrics *RateLimitMetrics `json:"metrics,omitempty"`
}

type RateLimitMetrics struct {
	// Prometheus defines the configuration for prometheus endpoint.
	Prometheus *RateLimitMetricsPrometheusProvider `json:"prometheus,omitempty"`
}

type RateLimitMetricsPrometheusProvider struct {
	// Disable the Prometheus endpoint.
	Disable bool `json:"disable,omitempty"`
}

// RateLimitDatabaseBackend defines the configuration associated with
// the database backend used by the rate limit service.
// +union
type RateLimitDatabaseBackend struct {
	// Type is the type of database backend to use. Supported types are:
	//	* Redis: Connects to a Redis database.
	//
	// +unionDiscriminator
	Type RateLimitDatabaseBackendType `json:"type"`
	// Redis defines the settings needed to connect to a Redis database.
	//
	// +optional
	Redis *RateLimitRedisSettings `json:"redis,omitempty"`
}

// RateLimitDatabaseBackendType specifies the types of database backend
// to be used by the rate limit service.
// +kubebuilder:validation:Enum=Redis
type RateLimitDatabaseBackendType string

const (
	// RedisBackendType uses a redis database for the rate limit service.
	RedisBackendType RateLimitDatabaseBackendType = "Redis"
)

// RedisTLSSettings defines the TLS configuration for connecting to redis database.
type RedisTLSSettings struct {
	// CertificateRef defines the client certificate reference for TLS connections.
	// Currently only a Kubernetes Secret of type TLS is supported.
	// +optional
	CertificateRef *gwapiv1.SecretObjectReference `json:"certificateRef,omitempty"`
}

// RateLimitRedisSettings defines the configuration for connecting to redis database.
type RateLimitRedisSettings struct {
	// URL of the Redis Database.
	URL string `json:"url"`

	// TLS defines TLS configuration for connecting to redis database.
	//
	// +optional
	TLS *RedisTLSSettings `json:"tls,omitempty"`
}

// ExtensionManager defines the configuration for registering an extension manager to
// the Envoy Gateway control plane.
type ExtensionManager struct {
	// Resources defines the set of K8s resources the extension will handle.
	//
	// +optional
	Resources []GroupVersionKind `json:"resources,omitempty"`

	// Hooks defines the set of hooks the extension supports
	//
	// +kubebuilder:validation:Required
	Hooks *ExtensionHooks `json:"hooks,omitempty"`

	// Service defines the configuration of the extension service that the Envoy
	// Gateway Control Plane will call through extension hooks.
	//
	// +kubebuilder:validation:Required
	Service *ExtensionService `json:"service,omitempty"`
}

// ExtensionHooks defines extension hooks across all supported runners
type ExtensionHooks struct {
	// XDSTranslator defines all the supported extension hooks for the xds-translator runner
	XDSTranslator *XDSTranslatorHooks `json:"xdsTranslator,omitempty"`
}

// XDSTranslatorHooks contains all the pre and post hooks for the xds-translator runner.
type XDSTranslatorHooks struct {
	Pre  []XDSTranslatorHook `json:"pre,omitempty"`
	Post []XDSTranslatorHook `json:"post,omitempty"`
}

// ExtensionService defines the configuration for connecting to a registered extension service.
type ExtensionService struct {
	// Host define the extension service hostname.
	Host string `json:"host"`

	// Port defines the port the extension service is exposed on.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=80
	Port int32 `json:"port,omitempty"`

	// TLS defines TLS configuration for communication between Envoy Gateway and
	// the extension service.
	//
	// +optional
	TLS *ExtensionTLS `json:"tls,omitempty"`
}

// ExtensionTLS defines the TLS configuration when connecting to an extension service
type ExtensionTLS struct {
	// CertificateRef contains a references to objects (Kubernetes objects or otherwise) that
	// contains a TLS certificate and private keys. These certificates are used to
	// establish a TLS handshake to the extension server.
	//
	// CertificateRef can only reference a Kubernetes Secret at this time.
	//
	// +kubebuilder:validation:Required
	CertificateRef gwapiv1.SecretObjectReference `json:"certificateRef"`
}

// EnvoyGatewayAdmin defines the Envoy Gateway Admin configuration.
type EnvoyGatewayAdmin struct {

	// Address defines the address of Envoy Gateway Admin Server.
	//
	// +optional
	Address *EnvoyGatewayAdminAddress `json:"address,omitempty"`
	// EnableDumpConfig defines if enable dump config in Envoy Gateway logs.
	//
	// +optional
	EnableDumpConfig bool `json:"enableDumpConfig,omitempty"`
	// EnablePprof defines if enable pprof in Envoy Gateway Admin Server.
	//
	// +optional
	EnablePprof bool `json:"enablePprof,omitempty"`
}

// EnvoyGatewayAdminAddress defines the Envoy Gateway Admin Address configuration.
type EnvoyGatewayAdminAddress struct {
	// Port defines the port the admin server is exposed on.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=19000
	Port int `json:"port,omitempty"`
	// Host defines the admin server hostname.
	//
	// +optional
	// +kubebuilder:default="127.0.0.1"
	Host string `json:"host,omitempty"`
}

func init() {
	SchemeBuilder.Register(&EnvoyGateway{})
}
