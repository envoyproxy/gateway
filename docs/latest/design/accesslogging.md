# Observability: Access Logging

## Overview

Envoy supports extensible access logging to different sinks, File, gRPC etc. Envoy supports customizable access log formats using predefined fields as well as arbitrary HTTP request and response headers. Envoy supports several built-in access log filters and extension filters that are registered at runtime.

Envoy Gateway leverages [Gateway API](https://gateway-api.sigs.k8s.io/) for configuring managed Envoy proxies. Gateway API defines core, extended, and implementation-specific API [support levels](https://gateway-api.sigs.k8s.io/concepts/conformance/?h=extended#2-support-levels) for implementors such as Envoy Gateway to expose features. Since access logging is not covered by `Core` or `Extended` APIs, EG should provide an easy to config access log formats and sinks per `EnvoyProxy`.

WARNING:

*Envoy Gateway will disable access log unless user config it via `EnvoyProxy`.*

## Goals

- Support send access logging to `File` or `OpenTelemetry` backend
- TODO: Support access log filters base on [CEL](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/access_loggers/filters/cel/v3/cel.proto#extension-envoy-access-loggers-extension-filters-cel) expression

## Non-Goals

- Support non-CEL filters, e.g. `status_code_filter`, `response_flag_filter`
- Support [HttpGrpcAccessLogConfig](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/access_loggers/grpc/v3/als.proto#extensions-access-loggers-grpc-v3-httpgrpcaccesslogconfig) or [TcpGrpcAccessLogConfig](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/access_loggers/grpc/v3/als.proto#extensions-access-loggers-grpc-v3-tcpgrpcaccesslogconfig)

## Use-Cases

- Configure access logging for a `EnvoyProxy` to `File`
- Configure access logging for a `EnvoyProxy` to `OpenTelemetry` backend
- Configure multi access logging providers for a `EnvoyProxy`

### ProxyAccessLogging API Type

```golang mdox-exec="sed '1,7d' api/config/v1alpha1/accesslogging_types.go"
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
	// can be used as values for json within the Struct.
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
```

### Example

1. The following is an example with text format access log.

```yaml
apiVersion: config.gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: text-access-logging
spec:
  accessLoggings:
    - format:
        type: text
        text: |
          [%START_TIME%] "%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%" %RESPONSE_CODE% %RESPONSE_FLAGS% %BYTES_RECEIVED% %BYTES_SENT% %DURATION% "%REQ(X-FORWARDED-FOR)%" "%REQ(USER-AGENT)%" "%REQ(X-REQUEST-ID)%" "%REQ(:AUTHORITY)%" "%UPSTREAM_HOST%"
      sinks:
        - type: "file"
          file:
            path: /dev/stdout

```

1. The following is an example with json format access log.

```yaml
apiVersion: config.gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: json-access-logging
spec:
  accessLoggings:
    - format:
        type: text
        json:
          status: "%RESPONSE_CODE%"
          message: "%LOCAL_REPLY_BODY%"
      sinks:
        - type: "file"
          file:
            path: /dev/stdout

```

1. The following is an example with OpenTelemetry format access log.

```yaml
apiVersion: config.gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: otel-access-logging
spec:
  accessLoggings:
    - format:
        type: text
        text: |
          [%START_TIME%] "%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%" %RESPONSE_CODE% %RESPONSE_FLAGS% %BYTES_RECEIVED% %BYTES_SENT% %DURATION% "%REQ(X-FORWARDED-FOR)%" "%REQ(USER-AGENT)%" "%REQ(X-REQUEST-ID)%" "%REQ(:AUTHORITY)%" "%UPSTREAM_HOST%"
      sinks:
        - type: "opentelemetry"
          opentelemetry:
            address: otel-collector.monitoring.svc.cluster.local:4317
          resources:
            k8s.cluster.name: "cluster-1"

```

1. The following is an example of sending same format to different sinks.

```yaml
apiVersion: config.gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: multi-providers
spec:
  accessLoggings:
    - format:
        type: text
        text: |
          [%START_TIME%] "%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%" %RESPONSE_CODE% %RESPONSE_FLAGS% %BYTES_RECEIVED% %BYTES_SENT% %DURATION% "%REQ(X-FORWARDED-FOR)%" "%REQ(USER-AGENT)%" "%REQ(X-REQUEST-ID)%" "%REQ(:AUTHORITY)%" "%UPSTREAM_HOST%"
      sinks:
        - type: "file"
          file:
            path: /dev/stdout
        - type: "opentelemetry"
          opentelemetry:
            address: otel-collector.monitoring.svc.cluster.local:4317
            resources:
              k8s.cluster.name: "cluster-1"

```
