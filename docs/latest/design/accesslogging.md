# Observability: Access Logging

## Overview

Envoy supports extensible access logging to different sinks, File, gRPC etc.
Envoy supports customizable access log formats using predefined fields as well as arbitrary HTTP request and response headers.
Envoy supports several built-in access log filters and extension filters that are registered at runtime.

Envoy Gateway leverages [Gateway API][] for configuring managed Envoy proxies. Gateway API defines core, extended, and
implementation-specific API [support levels][] for implementors such as Envoy Gateway to expose features. 
Since access logging is not covered by `Core` or `Extended` APIs, EG should provide an easy to config access log formats and sinks per `EnvoyProxy`.

WARNING:

  _Envoy Gateway will disable access log unless user config it via `EnvoyProxy`._

## Goals

- Support send access logging to `File` or `OpenTelemetry` backend
- TODO: Support access log filters base on [CEL][] expression

## Non-Goals

- Support non-CEL filters, e.g. `status_code_filter`, `response_flag_filter`
- Support [HttpGrpcAccessLogConfig][] or [TcpGrpcAccessLogConfig][]

## Use-Cases

- Configure access logging for a `EnvoyProxy` to `File`
- Configure access logging for a `EnvoyProxy` to `OpenTelemetry` backend
- Configure multi access logging providers for a `EnvoyProxy`

### ProxyAccessLogging API Type

```golang
type ProxyAccessLogging struct {
	// Format defines the format of access logging.
	Format ProxyAccessLoggingFormat `json:"format,omitempty"`
	// Sink defines the sink of access logging.
	Sink ProxyAccessLoggingSink `json:"sink,omitempty"`
}

type ProxyAccessLoggingFormatType string

const (
	// ProxyAccessLoggingFormatTypeText defines the text access logging format.
	ProxyAccessLoggingFormatTypeText ProxyAccessLoggingFormatType = "text"
	// ProxyAccessLoggingFormatTypeJSON defines the JSON access logging format.
	ProxyAccessLoggingFormatTypeJSON ProxyAccessLoggingFormatType = "json"
)

type ProxyAccessLoggingFormat struct {
	Type ProxyAccessLoggingFormatType `json:"type,omitempty"`
}

type ProxyAccessLoggingSinkType string

const (
	// ProxyAccessLoggingSinkTypeFile defines the file access logging sink.
	ProxyAccessLoggingSinkTypeFile ProxyAccessLoggingSinkType = "file"
	// ProxyAccessLoggingSinkTypeOpenTelemetry defines the OpenTelemetry access logging sink.
	ProxyAccessLoggingSinkTypeOpenTelemetry ProxyAccessLoggingSinkType = "opentelemetry"
)

type ProxyAccessLoggingSink struct {
	// Type defines the type of access logging sink.
	// +kubebuilder:validation:Enum=file;opentelemetry
	Type ProxyAccessLoggingSinkType `json:"type,omitempty"`
	// File defines the file access logging sink.
	// +optional
	File *FileEnvoyProxyAccessLogging `json:"file,omitempty"`
	// OpenTelemetry defines the OpenTelemetry access logging sink.
	// +optional
	OpenTelemetry *OpenTelemetryEnvoyProxyAccessLogging `json:"opentelemetry,omitempty"`
}

type FileEnvoyProxyAccessLogging struct {
	// Path defines the file path used to expose envoy access log(e.g. /dev/stdout).
	// Empty value disables access logging.
	Path string `json:"path,omitempty"`
	// Format is the format for the proxy access log, following Envoy access logging formatting, empty value results in proxy's default access log format.
	// Envoy [command operators](https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#command-operators) may be used in the format.
	// The [format string documentation](https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#config-access-log-format-strings) provides more information.
	Format string `json:"format,omitempty"`
}

type OpenTelemetryEnvoyProxyAccessLogging struct {
	// Host define the extension service hostname.
	Host string `json:"host"`
	// Port defines the port the extension service is exposed on.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=80
	Port int32 `json:"port,omitempty"`
	// Resources is a set of labels that describe the source of a log entry, including envoy node info.
	// It's recommended to follow [semantic conventions](https://opentelemetry.io/docs/reference/specification/resource/semantic_conventions/).
	// +optional
	Resources map[string]string `json:"resources,omitempty"`
}
````

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
	sink:
	  type: "file"
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
      fields:
        status: "%RESPONSE_CODE%"
        message: "%LOCAL_REPLY_BODY%"
	sink:
	  type: "file"
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
  - format:
      type: text
      text: |
		[%START_TIME%] "%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%" %RESPONSE_CODE% %RESPONSE_FLAGS% %BYTES_RECEIVED% %BYTES_SENT% %DURATION% "%REQ(X-FORWARDED-FOR)%" "%REQ(USER-AGENT)%" "%REQ(X-REQUEST-ID)%" "%REQ(:AUTHORITY)%" "%UPSTREAM_HOST%"
	sink:
	  type: "opentelemetry"
      opentelemetry:
	    address: otel-collector.monitoring.svc.cluster.local:4317
		resources:
          k8s.cluster.name: "cluster-1"
```

1. The following is an example with multi providers.

```yaml
apiVersion: config.gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: multi-providers
spec:
  - format:
      type: text
      text: |
		[%START_TIME%] "%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%" %RESPONSE_CODE% %RESPONSE_FLAGS% %BYTES_RECEIVED% %BYTES_SENT% %DURATION% "%REQ(X-FORWARDED-FOR)%" "%REQ(USER-AGENT)%" "%REQ(X-REQUEST-ID)%" "%REQ(:AUTHORITY)%" "%UPSTREAM_HOST%"
	sink:
	  type: "file"
      file:
	    path: /dev/stdout
  - format:
      type: text
      text: |
		[%START_TIME%] "%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%" %RESPONSE_CODE% %RESPONSE_FLAGS% %BYTES_RECEIVED% %BYTES_SENT% %DURATION% "%REQ(X-FORWARDED-FOR)%" "%REQ(USER-AGENT)%" "%REQ(X-REQUEST-ID)%" "%REQ(:AUTHORITY)%" "%UPSTREAM_HOST%"
	sink:
	  type: "opentelemetry"
      opentelemetry:
	    address: otel-collector.monitoring.svc.cluster.local:4317
		resources:
          k8s.cluster.name: "cluster-1"
```

[Gateway API]: https://gateway-api.sigs.k8s.io/
[support levels]: https://gateway-api.sigs.k8s.io/concepts/conformance/?h=extended#2-support-levels
[CEL]: https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/access_loggers/filters/cel/v3/cel.proto#extension-envoy-access-loggers-extension-filters-cel
[HttpGrpcAccessLogConfig]: https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/access_loggers/grpc/v3/als.proto#extensions-access-loggers-grpc-v3-httpgrpcaccesslogconfig
[TcpGrpcAccessLogConfig]: https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/access_loggers/grpc/v3/als.proto#extensions-access-loggers-grpc-v3-tcpgrpcaccesslogconfig
