// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

type ProxyTracing struct {
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
	CustomTags map[string]CustomTag `json:"customTags,omitempty"`
	// Provider defines the tracing provider.
	// Only OpenTelemetry is supported currently.
	Provider TracingProvider `json:"provider"`
}

type TracingProviderType string

const (
	TracingProviderTypeOpenTelemetry TracingProviderType = "OpenTelemetry"
)

// TracingProvider defines the tracing provider configuration.
//
// +kubebuilder:validation:XValidation:message="BackendRef only support Service Kind.",rule="!has(self.backendRef) || !has(self.backendRef.kind) || self.backendRef.kind == 'Service'"
type TracingProvider struct {
	// Type defines the tracing provider type.
	// EG currently only supports OpenTelemetry.
	// +kubebuilder:validation:Enum=OpenTelemetry
	// +kubebuilder:default=OpenTelemetry
	Type TracingProviderType `json:"type"`
	// Host define the provider service hostname.
	// Deprecated: Use BackendRef instead.
	Host string `json:"host"`
	// Port defines the port the provider service is exposed on.
	// Deprecated: Use BackendRef instead.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=4317
	Port int32 `json:"port,omitempty"`
	// BackendRef references a Kubernetes object that represents the
	// backend server to which the accesslog will be sent.
	// Only service Kind is supported for now.
	//
	// +optional
	BackendRef *gwapiv1.BackendObjectReference `json:"backendRef"`
}

type CustomTagType string

const (
	// CustomTagTypeLiteral adds hard-coded value to each span.
	CustomTagTypeLiteral CustomTagType = "Literal"
	// CustomTagTypeEnvironment adds value from environment variable to each span.
	CustomTagTypeEnvironment CustomTagType = "Environment"
	// CustomTagTypeRequestHeader adds value from request header to each span.
	CustomTagTypeRequestHeader CustomTagType = "RequestHeader"
)

type CustomTag struct {
	// Type defines the type of custom tag.
	// +kubebuilder:validation:Enum=Literal;Environment;RequestHeader
	// +unionDiscriminator
	// +kubebuilder:default=Literal
	Type CustomTagType `json:"type"`
	// Literal adds hard-coded value to each span.
	// It's required when the type is "Literal".
	Literal *LiteralCustomTag `json:"literal,omitempty"`
	// Environment adds value from environment variable to each span.
	// It's required when the type is "Environment".
	Environment *EnvironmentCustomTag `json:"environment,omitempty"`
	// RequestHeader adds value from request header to each span.
	// It's required when the type is "RequestHeader".
	RequestHeader *RequestHeaderCustomTag `json:"requestHeader,omitempty"`

	// TODO: add support for Metadata tags in the future.
	// EG currently doesn't support metadata for route or cluster.
}

// LiteralCustomTag adds hard-coded value to each span.
type LiteralCustomTag struct {
	// Value defines the hard-coded value to add to each span.
	Value string `json:"value"`
}

// EnvironmentCustomTag adds value from environment variable to each span.
type EnvironmentCustomTag struct {
	// Name defines the name of the environment variable which to extract the value from.
	Name string `json:"name"`
	// DefaultValue defines the default value to use if the environment variable is not set.
	// +optional
	DefaultValue *string `json:"defaultValue,omitempty"`
}

// RequestHeaderCustomTag adds value from request header to each span.
type RequestHeaderCustomTag struct {
	// Name defines the name of the request header which to extract the value from.
	Name string `json:"name"`
	// DefaultValue defines the default value to use if the request header is not set.
	// +optional
	DefaultValue *string `json:"defaultValue,omitempty"`
}
