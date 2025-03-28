// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

// ProxyTracing defines the tracing configuration for a proxy.
// +kubebuilder:validation:XValidation:message="only one of SamplingRate or SamplingFraction can be specified",rule="!(has(self.samplingRate) && has(self.samplingFraction))"
type ProxyTracing struct {
	// SamplingRate controls the rate at which traffic will be
	// selected for tracing if no prior sampling decision has been made.
	// Defaults to 100, valid values [0-100]. 100 indicates 100% sampling.
	//
	// Only one of SamplingRate or SamplingFraction may be specified.
	// If neither field is specified, all requests will be sampled.
	//
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	// +optional
	SamplingRate *uint32 `json:"samplingRate,omitempty"`
	// SamplingFraction represents the fraction of requests that should be
	// selected for tracing if no prior sampling decision has been made.
	//
	// Only one of SamplingRate or SamplingFraction may be specified.
	// If neither field is specified, all requests will be sampled.
	//
	// +optional
	SamplingFraction *gwapiv1.Fraction `json:"samplingFraction,omitempty"`
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
	TracingProviderTypeDatadog       TracingProviderType = "Datadog"
)

// TracingProvider defines the tracing provider configuration.
//
// +kubebuilder:validation:XValidation:message="host or backendRefs needs to be set",rule="has(self.host) || self.backendRefs.size() > 0"
// +kubebuilder:validation:XValidation:message="BackendRefs must be used, backendRef is not supported.",rule="!has(self.backendRef)"
// +kubebuilder:validation:XValidation:message="only supports Service kind.",rule="has(self.backendRefs) ? self.backendRefs.all(f, f.kind == 'Service') : true"
// +kubebuilder:validation:XValidation:message="BackendRefs only supports Core group.",rule="has(self.backendRefs) ? (self.backendRefs.all(f, f.group == \"\")) : true"
type TracingProvider struct {
	BackendCluster `json:",inline"`
	// Type defines the tracing provider type.
	// +kubebuilder:validation:Enum=OpenTelemetry;Zipkin;Datadog
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
