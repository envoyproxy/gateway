// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"
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
	Provider *EnvoyGatewayProvider `json:"provider,omitempty"`

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
	Extension *Extension `json:"extension,omitempty"`
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

	// File defines the configuration of the File provider. File provides runtime
	// configuration defined by one or more files. This type is not implemented
	// until https://github.com/envoyproxy/gateway/issues/1001 is fixed.
	//
	// +optional
	File *EnvoyGatewayFileProvider `json:"file,omitempty"`
}

// EnvoyGatewayKubernetesProvider defines configuration for the Kubernetes provider.
type EnvoyGatewayKubernetesProvider struct {
	// RateLimitDeployment defines the desired state of the Envoy ratelimit deployment resource.
	// If unspecified, default settings for the manged Envoy ratelimit deployment resource
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
}

// KubernetesWatchMode holds the configuration for which input resources to watch and reconcile.
type KubernetesWatchMode struct {
	// Namespaces holds the list of namespaces that Envoy Gateway will watch for namespaced scoped
	// resources such as Gateway, HTTPRoute and Service.
	// Note that Envoy Gateway will continue to reconcile relevant cluster scoped resources such as
	// GatewayClass that it is linked to.
	// By default, when this field is unset or empty, Envoy Gateway will watch for input namespaced resources
	// from all namespaces.
	Namespaces []string
}

// KubernetesDeployMode holds configuration for how to deploy managed resources such as the Envoy Proxy
// data plane fleet.
type KubernetesDeployMode struct {
	// EnvoyNamespaceType decides the namespace in which the managed Envoy Proxy data plane fleet
	// will be deployed in.
	// +optional
	EnvoyNamespaceType *EnvoyKubernetesNamespaceType `json:"envoyNamespaceType,omitempty"`
}

// EnvoyGatewayFileProvider defines configuration for the File provider.
type EnvoyGatewayFileProvider struct {
	// TODO: Add config as use cases are better understood.
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
	CertificateRef gwapiv1b1.SecretObjectReference `json:"certificateRef"`
}

func init() {
	SchemeBuilder.Register(&EnvoyGateway{})
}
