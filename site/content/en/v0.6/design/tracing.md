---
title: "Observability: Tracing"
---

## Overview

Envoy supports extensible tracing to different sinks, Zipkin, OpenTelemetry etc. Overview of Envoy tracing can be found [here](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/observability/tracing).

Envoy Gateway leverages [Gateway API](https://gateway-api.sigs.k8s.io/) for configuring managed Envoy proxies. Gateway API defines core, extended, and implementation-specific API [support levels](https://gateway-api.sigs.k8s.io/concepts/conformance/?h=extended#2-support-levels) for implementers such as Envoy Gateway to expose features. Since tracing is not covered by `Core` or `Extended` APIs, EG should provide an easy to config tracing per `EnvoyProxy`.

Only OpenTelemetry sink can be configured currently, you can use [OpenTelemetry Collector](https://opentelemetry.io/docs/collector/) to export to other tracing backends.

## Goals

- Support send tracing to `OpenTelemetry` backend
- Support configurable sampling rate
- Support propagate tag from `Literal`, `Environment` and `Request Header`

## Non-Goals

- Support other tracing backend, e.g. `Zipkin`, `Jaeger`

## Use-Cases

- Configure accesslog for a `EnvoyProxy` to `File`

### ProxyAccessLog API Type

```golang mdox-exec="sed '1,7d' api/config/v1alpha1/tracing_types.go"
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

type TracingProvider struct {
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
```

### Example

1. The following is an example to config tracing.

```yaml mdox-exec="sed '1,12d' examples/kubernetes/tracing/default.yaml"
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: tracing
  namespace: envoy-gateway-system
spec:
  telemetry:
    tracing:
      # sample 100% of requests
      samplingRate: 100
      provider:
        host: otel-collector.monitoring.svc.cluster.local
        port: 4317
      customTags:
        # This is an example of using a literal as a tag value
        key1:
          type: Literal
          literal:
            value: "val1"
        # This is an example of using an environment variable as a tag value
        env1:
          type: Environment
          environment:
            name: ENV1
            defaultValue: "-"
        # This is an example of using a header value as a tag value
        header1:
          type: RequestHeader
          requestHeader:
            name: X-Header-1
            defaultValue: "-"
```
