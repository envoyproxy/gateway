// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
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
	// DefaultKubernetesClientQPS defines the default QPS limit for the Kubernetes client.
	DefaultKubernetesClientQPS int32 = 50
	// DefaultKubernetesClientBurst defines the default Burst limit for the Kubernetes client.
	DefaultKubernetesClientBurst int32 = 100
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

type KubernetesClient struct {
	// RateLimit defines the rate limit settings for the Kubernetes client.
	RateLimit *KubernetesClientRateLimit `json:"rateLimit,omitempty"`
}

// KubernetesClientRateLimit defines the rate limit settings for the Kubernetes client.
type KubernetesClientRateLimit struct {
	// QPS defines the queries per second limit for the Kubernetes client.
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=50
	QPS *int32 `json:"qps,omitempty"`

	// Burst defines the maximum burst of requests allowed when tokens have accumulated.
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=100
	Burst *int32 `json:"burst,omitempty"`
}

// LeaderElection defines the desired leader election settings.
type LeaderElection struct {
	// LeaseDuration defines the time non-leader contenders will wait before attempting to claim leadership.
	// It's based on the timestamp of the last acknowledged signal. The default setting is 15 seconds.
	LeaseDuration *gwapiv1.Duration `json:"leaseDuration,omitempty"`
	// RenewDeadline represents the time frame within which the current leader will attempt to renew its leadership
	// status before relinquishing its position. The default setting is 10 seconds.
	RenewDeadline *gwapiv1.Duration `json:"renewDeadline,omitempty"`
	// RetryPeriod denotes the interval at which LeaderElector clients should perform action retries.
	// The default setting is 2 seconds.
	RetryPeriod *gwapiv1.Duration `json:"retryPeriod,omitempty"`
	// Disable provides the option to turn off leader election, which is enabled by default.
	Disable *bool `json:"disable,omitempty"`
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
	// EnableBackend enables Envoy Gateway to
	// reconcile and implement the Backend resources.
	EnableBackend bool `json:"enableBackend"`
}

// EnvoyGatewayProvider defines the desired configuration of a provider.
// +union
type EnvoyGatewayProvider struct {
	// Type is the type of provider to use. Supported types are "Kubernetes", "Custom".
	//
	// +unionDiscriminator
	Type ProviderType `json:"type"`

	// Kubernetes defines the configuration of the Kubernetes provider. Kubernetes
	// provides runtime configuration via the Kubernetes API.
	//
	// +optional
	Kubernetes *EnvoyGatewayKubernetesProvider `json:"kubernetes,omitempty"`

	// Custom defines the configuration for the Custom provider. This provider
	// allows you to define a specific resource provider and an infrastructure
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

	// RateLimitHpa defines the Horizontal Pod Autoscaler settings for Envoy ratelimit Deployment.
	// If the HPA is set, Replicas field from RateLimitDeployment will be ignored.
	//
	// +optional
	RateLimitHpa *KubernetesHorizontalPodAutoscalerSpec `json:"rateLimitHpa,omitempty"`

	// Watch holds configuration of which input resources should be watched and reconciled.
	// +optional
	Watch *KubernetesWatchMode `json:"watch,omitempty"`
	// Deploy holds configuration of how output managed resources such as the Envoy Proxy data plane
	// should be deployed
	// +optional
	// +notImplementedHide
	Deploy *KubernetesDeployMode `json:"deploy,omitempty"`
	// LeaderElection specifies the configuration for leader election.
	// If it's not set up, leader election will be active by default, using Kubernetes' standard settings.
	// +optional
	LeaderElection *LeaderElection `json:"leaderElection,omitempty"`

	// ShutdownManager defines the configuration for the shutdown manager.
	// +optional
	ShutdownManager *ShutdownManager `json:"shutdownManager,omitempty"`
	// Client holds the configuration for the Kubernetes client.
	Client *KubernetesClient `json:"client,omitempty"`
	// TopologyInjector defines the configuration for topology injector MutatatingWebhookConfiguration
	// +optional
	TopologyInjector *EnvoyGatewayTopologyInjector `json:"proxyTopologyInjector,omitempty"`

	// CacheSyncPeriod determines the minimum frequency at which watched resources are synced.
	// Note that a sync in the provider layer will not lead to a full reconciliation (including translation),
	// unless there are actual changes in the provider resources.
	// This option can be used to protect against missed events or issues in Envoy Gateway where resources
	// are not requeued when they should be, at the cost of increased resource consumption.
	// Learn more about the implications of this option: https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/cache#Options
	// Default: 10 hours
	// +optional
	CacheSyncPeriod *gwapiv1.Duration `json:"cacheSyncPeriod,omitempty"`
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

const (
	// KubernetesDeployModeTypeControllerNamespace indicates that the controller namespace is used for the infra proxy deployments.
	KubernetesDeployModeTypeControllerNamespace KubernetesDeployModeType = "ControllerNamespace"

	// KubernetesDeployModeTypeGatewayNamespace indicates that the gateway namespace is used for the infra proxy deployments.
	KubernetesDeployModeTypeGatewayNamespace KubernetesDeployModeType = "GatewayNamespace"
)

// KubernetesDeployModeType defines the type of KubernetesDeployMode
type KubernetesDeployModeType string

// KubernetesDeployMode holds configuration for how to deploy managed resources such as the Envoy Proxy
// data plane fleet.
type KubernetesDeployMode struct {
	// Type indicates what deployment mode to use. "ControllerNamespace" and
	// "GatewayNamespace" are currently supported.
	// By default, when this field is unset or empty, Envoy Gateway will deploy Envoy Proxy fleet in the Controller namespace.
	// +optional
	// +kubebuilder:default=ControllerNamespace
	// +kubebuilder:validation:Enum=ControllerNamespace;GatewayNamespace
	Type *KubernetesDeployModeType `json:"type,omitempty"`
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
	//
	// Infrastructure is optional, if provider is not specified,
	// No infrastructure provider is available.
	// +optional
	Infrastructure *EnvoyGatewayInfrastructureProvider `json:"infrastructure,omitempty"`
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
	// Recursive subdirectories are not currently supported.
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

	// Tracing defines traces configuration for RateLimit.
	Tracing *RateLimitTracing `json:"tracing,omitempty"`
}

type RateLimitMetrics struct {
	// Prometheus defines the configuration for prometheus endpoint.
	Prometheus *RateLimitMetricsPrometheusProvider `json:"prometheus,omitempty"`
}

type RateLimitMetricsPrometheusProvider struct {
	// Disable the Prometheus endpoint.
	Disable bool `json:"disable,omitempty"`
}

type RateLimitTracing struct {
	// SamplingRate controls the rate at which traffic will be
	// selected for tracing if no prior sampling decision has been made.
	// Defaults to 100, valid values [0-100]. 100 indicates 100% sampling.
	// +optional
	SamplingRate *uint32 `json:"samplingRate,omitempty"`

	// Provider defines the rateLimit tracing provider.
	// Only OpenTelemetry is supported currently.
	Provider *RateLimitTracingProvider `json:"provider,omitempty"`
}

type RateLimitTracingProviderType string

const (
	RateLimitTracingProviderTypeOpenTelemetry TracingProviderType = "OpenTelemetry"
)

// RateLimitTracingProvider defines the tracing provider configuration of RateLimit
type RateLimitTracingProvider struct {
	// Type defines the tracing provider type.
	// Since to RateLimit Exporter currently using OpenTelemetry, only OpenTelemetry is supported
	Type *RateLimitTracingProviderType `json:"type,omitempty"`

	// URL is the endpoint of the trace collector that supports the OTLP protocol
	URL string `json:"url"`
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
	// Resources defines the set of K8s resources the extension will handle as route
	// filter resources
	//
	// +optional
	Resources []GroupVersionKind `json:"resources,omitempty"`

	// PolicyResources defines the set of K8S resources the extension server will handle
	// as directly attached GatewayAPI policies
	//
	// +optional
	PolicyResources []GroupVersionKind `json:"policyResources,omitempty"`

	// Hooks defines the set of hooks the extension supports
	//
	// +kubebuilder:validation:Required
	Hooks *ExtensionHooks `json:"hooks,omitempty"`

	// Service defines the configuration of the extension service that the Envoy
	// Gateway Control Plane will call through extension hooks.
	//
	// +kubebuilder:validation:Required
	Service *ExtensionService `json:"service,omitempty"`

	// FailOpen defines if Envoy Gateway should ignore errors returned from the Extension Service hooks.
	//
	// When set to false, Envoy Gateway does not ignore extension Service hook errors. As a result,
	// xDS updates are skipped for the relevant envoy proxy fleet and the previous state is preserved.
	//
	// When set to true, if the Extension Service hooks return an error, no changes will be applied to the
	// source of the configuration which was sent to the extension server. The errors are ignored and the resulting
	// xDS configuration is updated in the xDS snapshot.
	//
	// Default: false
	//
	// +optional
	FailOpen bool `json:"failOpen,omitempty"`

	// MaxMessageSize defines the maximum message size in bytes that can be
	// sent to or received from the Extension Service.
	// Default: 4M
	//
	// +kubebuilder:validation:XIntOrString
	// +kubebuilder:validation:Pattern="^[1-9]+[0-9]*([EPTGMK]i|[EPTGMk])?$"
	// +optional
	MaxMessageSize *resource.Quantity `json:"maxMessageSize,omitempty"`
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
	// BackendEndpoint points to where the extension server can be found.
	BackendEndpoint `json:",inline"`

	// Host define the extension service hostname.
	// Deprecated: use the appropriate transport attribute instead (FQDN,IP,Unix)
	//
	// +optional
	Host string `json:"host,omitempty"`

	// Port defines the port the extension service is exposed on.
	// Deprecated: use the appropriate transport attribute instead (FQDN,IP,Unix)
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

	// Retry defines the retry policy for to use when errors are encountered in communication with
	// the extension service.
	//
	// +optional
	Retry *ExtensionServiceRetry `json:"retry,omitempty"`
}

// ExtensionTLS defines the TLS configuration when connecting to an extension service.
type ExtensionTLS struct {
	// CertificateRef is a reference to a Kubernetes Secret with a CA certificate in a key named "tls.crt".
	//
	// The CA certificate is used by Envoy Gateway the verify the server certificate presented by the extension server.
	// At this time, Envoy Gateway does not support Client Certificate authentication of Envoy Gateway towards the extension server (mTLS).
	//
	// +kubebuilder:validation:Required
	CertificateRef gwapiv1.SecretObjectReference `json:"certificateRef"`
}

// GRPCStatus defines grpc status codes as defined in https://github.com/grpc/grpc/blob/master/doc/statuscodes.md.
// +kubebuilder:validation:Enum=CANCELLED;UNKNOWN;INVALID_ARGUMENT;DEADLINE_EXCEEDED;NOT_FOUND;ALREADY_EXISTS;PERMISSION_DENIED;RESOURCE_EXHAUSTED;FAILED_PRECONDITION;ABORTED;OUT_OF_RANGE;UNIMPLEMENTED;INTERNAL;UNAVAILABLE;DATA_LOSS;UNAUTHENTICATED
type RetryableGRPCStatusCode string

// ExtensionServiceRetry defines the retry policy for to use when errors are encountered in communication with the extension service.
type ExtensionServiceRetry struct {
	// MaxAttempts defines the maximum number of retry attempts.
	// Default: 4
	//
	// +optional
	MaxAttempts *int `json:"maxAttempts,omitempty"`

	// InitialBackoff defines the initial backoff in seconds for retries, details: https://github.com/grpc/proposal/blob/master/A6-client-retries.md#integration-with-service-config.
	// Default: 0.1s
	//
	// +optional
	InitialBackoff *gwapiv1.Duration `json:"initialBackoff,omitempty"`

	// MaxBackoff defines the maximum backoff in seconds for retries.
	// Default: 1s
	//
	// +optional
	MaxBackoff *gwapiv1.Duration `json:"maxBackoff,omitempty"`

	// BackoffMultiplier defines the multiplier to use for exponential backoff for retries.
	// Default: 2.0
	//
	// +optional
	BackoffMultiplier *gwapiv1.Fraction `json:"backoffMultiplier,omitempty"`

	// RetryableStatusCodes defines the grpc status code for which retries will be attempted.
	// Default: [ "UNAVAILABLE" ]
	//
	// +optional
	RetryableStatusCodes []RetryableGRPCStatusCode `json:"RetryableStatusCodes,omitempty"`
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

// ShutdownManager defines the configuration for the shutdown manager.
type ShutdownManager struct {
	// Image specifies the ShutdownManager container image to be used, instead of the default image.
	Image *string `json:"image,omitempty"`
}

// EnvoyGatewayTopologyInjector defines the configuration for topology injector MutatatingWebhookConfiguration
type EnvoyGatewayTopologyInjector struct {
	// +optional
	Disable *bool `json:"disabled,omitempty"`
}

func init() {
	SchemeBuilder.Register(&EnvoyGateway{})
}
