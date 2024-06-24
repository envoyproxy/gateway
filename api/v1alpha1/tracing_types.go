// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

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
	Provider TracingProvider `json:"provider"`
}

type TracingProviderType string

const (
	TracingProviderTypeOpenTelemetry TracingProviderType = "OpenTelemetry"
	TracingProviderTypeZipkin        TracingProviderType = "Zipkin"
)

// TracingProvider defines the tracing provider configuration.
//
// +kubebuilder:validation:XValidation:message="host or backendRefs needs to be set",rule="has(self.host) || self.backendRefs.size() > 0"
type TracingProvider struct {
	// Type defines the tracing provider type.
	// +kubebuilder:validation:Enum=OpenTelemetry;Zipkin
	// +kubebuilder:default=OpenTelemetry
	Type TracingProviderType `json:"type"`
	// Host define the provider service hostname.
	// Deprecated: Use BackendRefs instead.
	//
	// +optional
	Host *string `json:"host,omitempty"`
	// Port defines the port the provider service is exposed on.
	// Deprecated: Use BackendRefs instead.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=4317
	Port int32 `json:"port,omitempty"`
	// BackendRefs references a Kubernetes object that represents the
	// backend server to which the trace will be sent.
	// Only Service kind is supported for now.
	//
	// +optional
	// +kubebuilder:validation:MaxItems=1
	// +kubebuilder:validation:XValidation:message="only support Service kind.",rule="self.all(f, f.kind == 'Service')"
	// +kubebuilder:validation:XValidation:message="BackendRefs only supports Core group.",rule="self.all(f, f.group == '')"
	BackendRefs []BackendRef `json:"backendRefs,omitempty"`
	// Zipkin defines the Zipkin tracing provider configuration
	// +optional
	Zipkin *ZipkinTracingProvider `json:"zipkin,omitempty"`
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

// ZipkinTracingProvider defines the Zipkin tracing provider configuration.
type ZipkinTracingProvider struct {
	// Enable128BitTraceID determines whether a 128bit trace id will be used
	// when creating a new trace instance. If set to false, a 64bit trace
	// id will be used.
	// +optional
	Enable128BitTraceID *bool `json:"enable128BitTraceId,omitempty"`
	// DisableSharedSpanContext determines whether the default Envoy behaviour of
	// client and server spans sharing the same span context should be disabled.
	// +optional
	DisableSharedSpanContext *bool `json:"disableSharedSpanContext,omitempty"`
}
