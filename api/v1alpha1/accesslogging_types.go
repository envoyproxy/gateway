// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

type ProxyAccessLog struct {
	// Disable disables access logging for managed proxies if set to true.
	Disable bool `json:"disable,omitempty"`
	// Settings defines accesslog settings for managed proxies.
	// If unspecified, will send default format to stdout.
	// +optional
	Settings []ProxyAccessLogSetting `json:"settings,omitempty"`
}

type ProxyAccessLogSetting struct {
	// Format defines the format of accesslog.
	Format ProxyAccessLogFormat `json:"format"`
	// Sinks defines the sinks of accesslog.
	// +kubebuilder:validation:MinItems=1
	Sinks []ProxyAccessLogSink `json:"sinks"`
}

type ProxyAccessLogFormatType string

const (
	// ProxyAccessLogFormatTypeText defines the text accesslog format.
	ProxyAccessLogFormatTypeText ProxyAccessLogFormatType = "Text"
	// ProxyAccessLogFormatTypeJSON defines the JSON accesslog format.
	ProxyAccessLogFormatTypeJSON ProxyAccessLogFormatType = "JSON"
	// TODO: support format type "mix" in the future.
)

// ProxyAccessLogFormat defines the format of accesslog.
// By default accesslogs are written to standard output.
// +union
//
// +kubebuilder:validation:XValidation:rule="self.type == 'Text' ? has(self.text) : !has(self.text)",message="If AccessLogFormat type is Text, text field needs to be set."
// +kubebuilder:validation:XValidation:rule="self.type == 'JSON' ? has(self.json) : !has(self.json)",message="If AccessLogFormat type is JSON, json field needs to be set."
type ProxyAccessLogFormat struct {
	// Type defines the type of accesslog format.
	// +kubebuilder:validation:Enum=Text;JSON
	// +unionDiscriminator
	Type ProxyAccessLogFormatType `json:"type,omitempty"`
	// Text defines the text accesslog format, following Envoy accesslog formatting,
	// It's required when the format type is "Text".
	// Envoy [command operators](https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#command-operators) may be used in the format.
	// The [format string documentation](https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#config-access-log-format-strings) provides more information.
	// +optional
	Text *string `json:"text,omitempty"`
	// JSON is additional attributes that describe the specific event occurrence.
	// Structured format for the envoy access logs. Envoy [command operators](https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#command-operators)
	// can be used as values for fields within the Struct.
	// It's required when the format type is "JSON".
	// +optional
	JSON map[string]string `json:"json,omitempty"`
}

type ProxyAccessLogSinkType string

const (
	// ProxyAccessLogSinkTypeFile defines the file accesslog sink.
	ProxyAccessLogSinkTypeFile ProxyAccessLogSinkType = "File"
	// ProxyAccessLogSinkTypeOpenTelemetry defines the OpenTelemetry accesslog sink.
	// When the provider is Kubernetes, EnvoyGateway always sends `k8s.namespace.name`
	// and `k8s.pod.name` as additional attributes.
	ProxyAccessLogSinkTypeOpenTelemetry ProxyAccessLogSinkType = "OpenTelemetry"
)

// ProxyAccessLogSink defines the sink of accesslog.
// +union
//
// +kubebuilder:validation:XValidation:rule="self.type == 'File' ? has(self.file) : !has(self.file)",message="If AccessLogSink type is File, file field needs to be set."
// +kubebuilder:validation:XValidation:rule="self.type == 'OpenTelemetry' ? has(self.openTelemetry) : !has(self.openTelemetry)",message="If AccessLogSink type is OpenTelemetry, openTelemetry field needs to be set."
type ProxyAccessLogSink struct {
	// Type defines the type of accesslog sink.
	// +kubebuilder:validation:Enum=File;OpenTelemetry
	// +unionDiscriminator
	Type ProxyAccessLogSinkType `json:"type,omitempty"`
	// File defines the file accesslog sink.
	// +optional
	File *FileEnvoyProxyAccessLog `json:"file,omitempty"`
	// OpenTelemetry defines the OpenTelemetry accesslog sink.
	// +optional
	OpenTelemetry *OpenTelemetryEnvoyProxyAccessLog `json:"openTelemetry,omitempty"`
}

type FileEnvoyProxyAccessLog struct {
	// Path defines the file path used to expose envoy access log(e.g. /dev/stdout).
	// +kubebuilder:validation:MinLength=1
	Path string `json:"path,omitempty"`
}

// OpenTelemetryEnvoyProxyAccessLog defines the OpenTelemetry access log sink.
//
// +kubebuilder:validation:XValidation:message="BackendRef only support Service Kind.",rule="!has(self.backendRef) || !has(self.backendRef.kind) || self.backendRef.kind == 'Service'"
type OpenTelemetryEnvoyProxyAccessLog struct {
	// Host define the extension service hostname.
	// Deprecated: Use BackendRef instead.
	Host string `json:"host"`
	// Port defines the port the extension service is exposed on.
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
	BackendRef *gwapiv1.BackendObjectReference `json:"backendRef,omitempty"`
	// Resources is a set of labels that describe the source of a log entry, including envoy node info.
	// It's recommended to follow [semantic conventions](https://opentelemetry.io/docs/reference/specification/resource/semantic_conventions/).
	// +optional
	Resources map[string]string `json:"resources,omitempty"`

	// TODO: support more OpenTelemetry accesslog options(e.g. TLS, auth etc.) in the future.
}
