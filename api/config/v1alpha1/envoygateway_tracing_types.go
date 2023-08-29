// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

type EnvoyGatewayTracing struct {
	// SamplingRate controls the rate at which traffic will be
	// selected for tracing if no prior sampling decision has been made.
	// Defaults to 100, valid values [0-100]. 100 indicates 100% sampling.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	// +kubebuilder:default=100
	// +optional
	SamplingRate *uint32 `json:"samplingRate,omitempty"`
	// CustomTags defines the custom tags to add to each span.
	// If provider is kubernetes, pod name and namespace are added by default.
	CustomTags map[string]EnvoyGatewayCustomTag `json:"customTags,omitempty"`
	// Provider defines the tracing provider.
	// Only OpenTelemetry is supported currently.
	Provider EnvoyGatewayTracingProvider `json:"provider"`
}

type EnvoyGatewayCustomTag struct {
	// Type defines the type of custom tag.
	// +kubebuilder:validation:Enum=Literal;Environment
	// +unionDiscriminator
	// +kubebuilder:default=Literal
	Type CustomTagType `json:"type"`
	// Literal adds hard-coded value to each span.
	// It's required when the type is "Literal".
	Literal *LiteralCustomTag `json:"literal,omitempty"`
	// Environment adds value from environment variable to each span.
	// It's required when the type is "Environment".
	Environment *EnvironmentCustomTag `json:"environment,omitempty"`
}

type EnvoyGatewayTracingProvider struct {
	// Type defines the tracing provider type.
	// EG currently only supports OpenTelemetry.
	// +kubebuilder:validation:Enum=OpenTelemetry
	// +kubebuilder:default=OpenTelemetry
	Type TracingProviderType `json:"type"`
	// Host define the provider service hostname.
	Host string `json:"host"`
	// Port defines the port the provider service is exposed on.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=4317
	Port int32 `json:"port,omitempty"`
}
