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

// EnvoyGateway is the Schema for the envoygateways API.
type EnvoyGateway struct {
	metav1.TypeMeta `json:",inline"`

	// EnvoyGatewaySpec defines the desired state of Envoy Gateway.
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
}

// Gateway defines the desired Gateway API configuration of Envoy Gateway.
type Gateway struct {
	// ControllerName defines the name of the Gateway API controller. If unspecified,
	// defaults to "gateway.envoyproxy.io/gatewayclass-controller". See the following
	// for additional details:
	//
	// https://gateway-api.sigs.k8s.io/v1alpha2/references/spec/#gateway.networking.k8s.io/v1alpha2.GatewayClass
	//
	// +optional
	ControllerName string `json:"controllerName,omitempty"`
}

// Provider defines the desired configuration of a provider.
// +union
type Provider struct {
	// Type is the type of provider to use.
	//
	// +unionDiscriminator
	Type ProviderType `json:"type"`
	// Kubernetes defines the configuration of the Kubernetes provider. Kubernetes
	// provides runtime configuration via the Kubernetes API.
	//
	// +optional
	Kubernetes *KubernetesProvider `json:"kubernetes,omitempty"`

	// File defines the configuration of the File provider. File provides runtime
	// configuration defined by one or more files.
	//
	// +optional
	File *FileProvider `json:"file,omitempty"`
}

// ProviderType defines the types of providers supported by Envoy Gateway.
type ProviderType string

const (
	// ProviderTypeKubernetes defines the "Kubernetes" provider.
	ProviderTypeKubernetes ProviderType = "Kubernetes"

	// ProviderTypeFile defines the "File" provider.
	ProviderTypeFile ProviderType = "File"
)

// KubernetesProvider defines configuration for the Kubernetes provider.
type KubernetesProvider struct {
	// TODO: Add config as use cases are better understood.
}

// FileProvider defines configuration for the File provider.
type FileProvider struct {
	// TODO: Add config as use cases are better understood.
}

// DefaultEnvoyGateway returns a new EnvoyGateway with default configuration parameters.
func DefaultEnvoyGateway() *EnvoyGateway {
	gw := DefaultGateway()
	p := DefaultProvider()
	return &EnvoyGateway{
		metav1.TypeMeta{
			Kind:       KindEnvoyGateway,
			APIVersion: GroupVersion.String(),
		},
		EnvoyGatewaySpec{
			Gateway:  gw,
			Provider: p,
		},
	}
}

// SetDefaults sets default EnvoyGateway configuration parameters.
func (e *EnvoyGateway) SetDefaults() {
	if e.TypeMeta.Kind == "" {
		e.TypeMeta.Kind = KindEnvoyGateway
	}
	if e.TypeMeta.APIVersion == "" {
		e.TypeMeta.APIVersion = GroupVersion.String()
	}
	if e.Provider == nil {
		e.Provider = DefaultProvider()
	}
	if e.Gateway == nil {
		e.Gateway = DefaultGateway()
	}
}

// DefaultGateway returns a new Gateway with default configuration parameters.
func DefaultGateway() *Gateway {
	return &Gateway{
		ControllerName: GatewayControllerName,
	}
}

// DefaultProvider returns a new Provider with default configuration parameters.
func DefaultProvider() *Provider {
	return &Provider{
		Type: ProviderTypeKubernetes,
	}
}

func init() {
	SchemeBuilder.Register(&EnvoyGateway{})
}
