// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

type ProxyAccessLog struct {
	// Disable disables access logging for managed proxies if set to true.
	//
	// +optional
	Disable *bool `json:"disable,omitempty"`
	// Settings defines accesslog settings for managed proxies.
	// If unspecified, will send default format to stdout.
	// +optional
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=50
	Settings []ProxyAccessLogSetting `json:"settings,omitempty"`
}

type ProxyAccessLogSetting struct {
	// Format defines the format of accesslog.
	// This will be ignored if sink type is ALS.
	// +optional
	Format *ProxyAccessLogFormat `json:"format,omitempty"`
	// Matches defines the match conditions for accesslog in CEL expression.
	// An accesslog will be emitted only when one or more match conditions are evaluated to true.
	// Invalid [CEL](https://www.envoyproxy.io/docs/envoy/latest/xds/type/v3/cel.proto.html#common-expression-language-cel-proto) expressions will be ignored.
	// +kubebuilder:validation:MaxItems=10
	Matches []string `json:"matches,omitempty"`
	// Sinks defines the sinks of accesslog.
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=50
	Sinks []ProxyAccessLogSink `json:"sinks"`
	// Type defines the component emitting the accesslog, such as Listener and Route.
	// If type not defined, the setting would apply to:
	// (1) All Routes.
	// (2) Listeners if and only if Envoy does not find a matching route for a request.
	// If type is defined, the accesslog settings would apply to the relevant component (as-is).
	// +kubebuilder:validation:Enum=Listener;Route
	// +optional
	Type *ProxyAccessLogType `json:"type,omitempty"`
}

type ProxyAccessLogType string

const (
	// ProxyAccessLogTypeListener defines the accesslog for Listeners.
	// https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/listener/v3/listener.proto#envoy-v3-api-field-config-listener-v3-listener-access-log
	ProxyAccessLogTypeListener ProxyAccessLogType = "Listener"
	// ProxyAccessLogTypeRoute defines the accesslog for HTTP, GRPC, UDP and TCP Routes.
	// https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/udp/udp_proxy/v3/udp_proxy.proto#envoy-v3-api-field-extensions-filters-udp-udp-proxy-v3-udpproxyconfig-access-log
	// https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/network/tcp_proxy/v3/tcp_proxy.proto#envoy-v3-api-field-extensions-filters-network-tcp-proxy-v3-tcpproxy-access-log
	// https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/network/http_connection_manager/v3/http_connection_manager.proto#envoy-v3-api-field-extensions-filters-network-http-connection-manager-v3-httpconnectionmanager-access-log
	ProxyAccessLogTypeRoute ProxyAccessLogType = "Route"
)

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
	// ProxyAccessLogSinkTypeALS defines the gRPC Access Log Service (ALS) sink.
	// The service must implement the Envoy gRPC Access Log Service streaming API:
	// https://www.envoyproxy.io/docs/envoy/latest/api-v3/service/accesslog/v3/als.proto
	ProxyAccessLogSinkTypeALS ProxyAccessLogSinkType = "ALS"
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
// +kubebuilder:validation:XValidation:rule="self.type == 'ALS' ? has(self.als) : !has(self.als)",message="If AccessLogSink type is ALS, als field needs to be set."
// +kubebuilder:validation:XValidation:rule="self.type == 'File' ? has(self.file) : !has(self.file)",message="If AccessLogSink type is File, file field needs to be set."
// +kubebuilder:validation:XValidation:rule="self.type == 'OpenTelemetry' ? has(self.openTelemetry) : !has(self.openTelemetry)",message="If AccessLogSink type is OpenTelemetry, openTelemetry field needs to be set."
type ProxyAccessLogSink struct {
	// Type defines the type of accesslog sink.
	// +kubebuilder:validation:Enum=ALS;File;OpenTelemetry
	// +unionDiscriminator
	Type ProxyAccessLogSinkType `json:"type,omitempty"`
	// ALS defines the gRPC Access Log Service (ALS) sink.
	// +optional
	ALS *ALSEnvoyProxyAccessLog `json:"als,omitempty"`
	// File defines the file accesslog sink.
	// +optional
	File *FileEnvoyProxyAccessLog `json:"file,omitempty"`
	// OpenTelemetry defines the OpenTelemetry accesslog sink.
	// +optional
	OpenTelemetry *OpenTelemetryEnvoyProxyAccessLog `json:"openTelemetry,omitempty"`
}

type ALSEnvoyProxyAccessLogType string

const (
	// ALSEnvoyProxyAccessLogTypeHTTP defines the HTTP access log type and will populate StreamAccessLogsMessage.http_logs.
	ALSEnvoyProxyAccessLogTypeHTTP ALSEnvoyProxyAccessLogType = "HTTP"
	// ALSEnvoyProxyAccessLogTypeTCP defines the TCP access log type and will populate StreamAccessLogsMessage.tcp_logs.
	ALSEnvoyProxyAccessLogTypeTCP ALSEnvoyProxyAccessLogType = "TCP"
)

// ALSEnvoyProxyAccessLog defines the gRPC Access Log Service (ALS) sink.
// The service must implement the Envoy gRPC Access Log Service streaming API:
// https://www.envoyproxy.io/docs/envoy/latest/api-v3/service/accesslog/v3/als.proto
// Access log format information is passed in the form of gRPC metadata when the
// stream is established.
//
// +kubebuilder:validation:XValidation:rule="self.type == 'HTTP' || !has(self.http)",message="The http field may only be set when type is HTTP."
// +kubebuilder:validation:XValidation:message="BackendRefs must be used, backendRef is not supported.",rule="!has(self.backendRef)"
// +kubebuilder:validation:XValidation:message="must have at least one backend in backendRefs",rule="has(self.backendRefs) && self.backendRefs.size() > 0"
// +kubebuilder:validation:XValidation:message="BackendRefs only supports Service kind.",rule="has(self.backendRefs) ? self.backendRefs.all(f, f.kind == 'Service') : true"
// +kubebuilder:validation:XValidation:message="BackendRefs only supports Core group.",rule="has(self.backendRefs) ? (self.backendRefs.all(f, f.group == \"\")) : true"
type ALSEnvoyProxyAccessLog struct {
	BackendCluster `json:",inline"`

	// LogName defines the friendly name of the access log to be returned in
	// StreamAccessLogsMessage.Identifier. This allows the access log server
	// to differentiate between different access logs coming from the same Envoy.
	// +optional
	// +kubebuilder:validation:MinLength=1
	LogName *string `json:"logName,omitempty"`
	// Type defines the type of accesslog. Supported types are "HTTP" and "TCP".
	// +kubebuilder:validation:Enum=HTTP;TCP
	Type ALSEnvoyProxyAccessLogType `json:"type"`
	// HTTP defines additional configuration specific to HTTP access logs.
	// +optional
	HTTP *ALSEnvoyProxyHTTPAccessLogConfig `json:"http,omitempty"`
}

type ALSEnvoyProxyHTTPAccessLogConfig struct {
	// RequestHeaders defines request headers to include in log entries sent to the access log service.
	// +optional
	RequestHeaders []string `json:"requestHeaders,omitempty"`
	// ResponseHeaders defines response headers to include in log entries sent to the access log service.
	// +optional
	ResponseHeaders []string `json:"responseHeaders,omitempty"`
	// ResponseTrailers defines response trailers to include in log entries sent to the access log service.
	// +optional
	ResponseTrailers []string `json:"responseTrailers,omitempty"`
}

type FileEnvoyProxyAccessLog struct {
	// Path defines the file path used to expose envoy access log(e.g. /dev/stdout).
	// +kubebuilder:validation:MinLength=1
	Path string `json:"path,omitempty"`
}

// OpenTelemetryEnvoyProxyAccessLog defines the OpenTelemetry access log sink.
//
// +kubebuilder:validation:XValidation:message="host or backendRefs needs to be set",rule="has(self.host) || self.backendRefs.size() > 0"
// +kubebuilder:validation:XValidation:message="BackendRefs must be used, backendRef is not supported.",rule="!has(self.backendRef)"
// +kubebuilder:validation:XValidation:message="BackendRefs only supports Service kind.",rule="has(self.backendRefs) ? self.backendRefs.all(f, f.kind == 'Service') : true"
// +kubebuilder:validation:XValidation:message="BackendRefs only supports Core group.",rule="has(self.backendRefs) ? (self.backendRefs.all(f, f.group == \"\")) : true"
type OpenTelemetryEnvoyProxyAccessLog struct {
	BackendCluster `json:",inline"`
	// Host define the extension service hostname.
	// Deprecated: Use BackendRefs instead.
	//
	// +optional
	Host *string `json:"host,omitempty"`
	// Port defines the port the extension service is exposed on.
	// Deprecated: Use BackendRefs instead.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=4317
	Port int32 `json:"port,omitempty"`
	// Resources is a set of labels that describe the source of a log entry, including envoy node info.
	// It's recommended to follow [semantic conventions](https://opentelemetry.io/docs/reference/specification/resource/semantic_conventions/).
	// +optional
	Resources map[string]string `json:"resources,omitempty"`

	// TODO: support more OpenTelemetry accesslog options(e.g. TLS, auth etc.) in the future.
}
