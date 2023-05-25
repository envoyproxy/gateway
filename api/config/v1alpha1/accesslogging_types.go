// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

type ProxyAccessLogging struct {
	// Format defines the format of access logging.
	Format ProxyAccessLoggingFormat `json:"format"`
	// Sinks defines the sinks of access logging.
	// +kubebuilder:validation:MinItems=1
	Sinks []ProxyAccessLoggingSink `json:"sinks"`
}

type ProxyAccessLoggingFormatType string

const (
	// ProxyAccessLoggingFormatTypeText defines the text access logging format.
	ProxyAccessLoggingFormatTypeText ProxyAccessLoggingFormatType = "Text"
	// ProxyAccessLoggingFormatTypeJSON defines the JSON access logging format.
	ProxyAccessLoggingFormatTypeJSON ProxyAccessLoggingFormatType = "JSON"
	// TODO: support format type "mix" in the future.
)

// ProxyAccessLoggingFormat defines the format of access logging.
// +union
type ProxyAccessLoggingFormat struct {
	// Type defines the type of access logging format.
	// +kubebuilder:validation:Enum=Text;JSON
	// +unionDiscriminator
	Type ProxyAccessLoggingFormatType `json:"type,omitempty"`
	// Text defines the text access logging format, following Envoy access logging formatting,
	// empty value results in proxy's default access log format.
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

type ProxyAccessLoggingSinkType string

const (
	// ProxyAccessLoggingSinkTypeFile defines the file access logging sink.
	ProxyAccessLoggingSinkTypeFile ProxyAccessLoggingSinkType = "File"
	// ProxyAccessLoggingSinkTypeOpenTelemetry defines the OpenTelemetry access logging sink.
	ProxyAccessLoggingSinkTypeOpenTelemetry ProxyAccessLoggingSinkType = "OpenTelemetry"
)

type ProxyAccessLoggingSink struct {
	// Type defines the type of access logging sink.
	// +kubebuilder:validation:Enum=File;OpenTelemetry
	Type ProxyAccessLoggingSinkType `json:"type,omitempty"`
	// File defines the file access logging sink.
	// +optional
	File *FileEnvoyProxyAccessLogging `json:"file,omitempty"`
	// OpenTelemetry defines the OpenTelemetry access logging sink.
	// +optional
	OpenTelemetry *OpenTelemetryEnvoyProxyAccessLogging `json:"openTelemetry,omitempty"`
}

type FileEnvoyProxyAccessLogging struct {
	// Path defines the file path used to expose envoy access log(e.g. /dev/stdout).
	// Empty value disables access logging.
	Path string `json:"path,omitempty"`
}

// TODO: consider reuse ExtensionService?
type OpenTelemetryEnvoyProxyAccessLogging struct {
	// Host define the extension service hostname.
	Host string `json:"host"`
	// Port defines the port the extension service is exposed on.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=4317
	Port int32 `json:"port,omitempty"`
	// Resources is a set of labels that describe the source of a log entry, including envoy node info.
	// It's recommended to follow [semantic conventions](https://opentelemetry.io/docs/reference/specification/resource/semantic_conventions/).
	// +optional
	Resources map[string]string `json:"resources,omitempty"`

	// TODO: support more OpenTelemetry access logging options(e.g. TLS, auth etc.) in the future.
}
