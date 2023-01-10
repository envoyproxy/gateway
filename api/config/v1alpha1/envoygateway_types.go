// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// KindEnvoyGateway is the name of the EnvoyGateway kind.
	KindEnvoyGateway = "EnvoyGateway"
	// GatewayControllerName is the name of the GatewayClass controller.
	GatewayControllerName = "gateway.envoyproxy.io/gatewayclass-controller"
)

//+kubebuilder:object:root=true

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
	Provider *Provider `json:"provider,omitempty"`

	// RateLimit defines the configuration associated with the Rate Limit service
	// deployed by Envoy Gateway required to implement the Global Rate limiting
	// functionality. The specific rate limit service used here is the reference
	// implementation in Envoy. For more details visit https://github.com/envoyproxy/ratelimit.
	// This configuration is unneeded for "Local" rate limiting.
	//
	// +optional
	RateLimit *RateLimit `json:"rateLimit,omitempty"`

	// Extension defines an extension to register for the Envoy Gateway Control Plane.
	//
	// +optional
	Extension *Extension `json:"extensions,omitempty"`
}

// Gateway defines the desired Gateway API configuration of Envoy Gateway.
type Gateway struct {
	// ControllerName defines the name of the Gateway API controller. If unspecified,
	// defaults to "gateway.envoyproxy.io/gatewayclass-controller". See the following
	// for additional details:
	//   https://gateway-api.sigs.k8s.io/v1alpha2/references/spec/#gateway.networking.k8s.io/v1alpha2.GatewayClass
	//
	// +optional
	ControllerName string `json:"controllerName,omitempty"`
}

// Provider defines the desired configuration of a provider.
// +union
type Provider struct {
	// Type is the type of provider to use. Supported types are "Kubernetes".
	//
	// +unionDiscriminator
	Type ProviderType `json:"type"`

	// Kubernetes defines the configuration of the Kubernetes provider. Kubernetes
	// provides runtime configuration via the Kubernetes API.
	//
	// +optional
	Kubernetes *KubernetesProvider `json:"kubernetes,omitempty"`

	// File defines the configuration of the File provider. File provides runtime
	// configuration defined by one or more files. This type is not implemented
	// until https://github.com/envoyproxy/gateway/issues/1001 is fixed.
	//
	// +optional
	File *FileProvider `json:"file,omitempty"`
}

// KubernetesProvider defines configuration for the Kubernetes provider.
type KubernetesProvider struct {
	// TODO: Add config as use cases are better understood.
}

// FileProvider defines configuration for the File provider.
type FileProvider struct {
	// TODO: Add config as use cases are better understood.
}

// TLSSecret defines the configuration for getting TLS certificates from a Secret.
type TLSSecret struct {
	// Name is the secret name to load the TLS certificate from
	Name string `json:"name"`

	// Namespace is the namespace where the secret is located.
	//
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

// TLSFile defines the configuration for getting TLS certificates from a file.
type TLSFile struct {
	// TODO: Add config when we are ready to support this
}

// RateLimit defines the configuration associated with the Rate Limit Service
// used for Global Rate Limiting.
type RateLimit struct {
	// Backend holds the configuration associated with the
	// database backend used by the rate limit service to store
	// state associated with global ratelimiting.
	Backend RateLimitDatabaseBackend `json:"backend"`
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

// RateLimitRedisSettings defines the configuration for connecting to
// a Redis database.
type RateLimitRedisSettings struct {
	// URL of the Redis Database.
	URL string `json:"url"`
}

// Extension defines the configuration for registering an extension to
// the Envoy Gateway control plane.
type Extension struct {
	// Resources defines the set of K8s resources the extension will handle.
	Resources []*GroupVersionKind `json:"resources,omitempty"`

	// Service defines the configuration of the extension service that the Envoy
	// Gateway Control Plane will call through extension hooks.
	Service *ExtensionService `json:"service"`
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

// ExtensionTLS defines the TLS configuration when connecting to an extension service.
type ExtensionTLS struct {
	// Type is the method for how the TLS certificate is loaded. Supported types are:
	//
	//   * Secret: Load the TLS certificate from a K8s secret.
	//
	// +unionDiscriminator
	Type TLSType `json:"type"`

	// Secret defines which K8s secret to load the TLS certificate from.
	//
	// +optional
	Secret *TLSSecret `json:"secret,omitempty"`

	// File defines the configuration for loading the TLS certificate from the filesystem.
	//
	// +optional
	File *TLSFile `json:"file,omitempty"`
}

func init() {
	SchemeBuilder.Register(&EnvoyGateway{})
}
