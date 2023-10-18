---
title: "Observability: Accesslog"
---

## Overview

Envoy supports extensible accesslog to different sinks, File, gRPC etc. Envoy supports customizable access log formats using predefined fields as well as arbitrary HTTP request and response headers. Envoy supports several built-in access log filters and extension filters that are registered at runtime.

Envoy Gateway leverages [Gateway API](https://gateway-api.sigs.k8s.io/) for configuring managed Envoy proxies. Gateway API defines core, extended, and implementation-specific API [support levels](https://gateway-api.sigs.k8s.io/concepts/conformance/?h=extended#2-support-levels) for implementers such as Envoy Gateway to expose features. Since accesslog is not covered by `Core` or `Extended` APIs, EG should provide an easy to config access log formats and sinks per `EnvoyProxy`.

## Goals

- Support send accesslog to `File` or `OpenTelemetry` backend
- TODO: Support access log filters base on [CEL](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/access_loggers/filters/cel/v3/cel.proto#extension-envoy-access-loggers-extension-filters-cel) expression

## Non-Goals

- Support non-CEL filters, e.g. `status_code_filter`, `response_flag_filter`
- Support [HttpGrpcAccessLogConfig](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/access_loggers/grpc/v3/als.proto#extensions-access-loggers-grpc-v3-httpgrpcaccesslogconfig) or [TcpGrpcAccessLogConfig](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/access_loggers/grpc/v3/als.proto#extensions-access-loggers-grpc-v3-tcpgrpcaccesslogconfig)

## Use-Cases

- Configure accesslog for a `EnvoyProxy` to `File`
- Configure accesslog for a `EnvoyProxy` to `OpenTelemetry` backend
- Configure multi accesslog providers for a `EnvoyProxy`

### ProxyAccessLog API Type

```golang mdox-exec="sed '1,7d' api/config/v1alpha1/accesslogging_types.go"
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
// +union
type ProxyAccessLogFormat struct {
	// Type defines the type of accesslog format.
	// +kubebuilder:validation:Enum=Text;JSON
	// +unionDiscriminator
	Type ProxyAccessLogFormatType `json:"type,omitempty"`
	// Text defines the text accesslog format, following Envoy accesslog formatting,
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

type ProxyAccessLogSinkType string

const (
	// ProxyAccessLogSinkTypeFile defines the file accesslog sink.
	ProxyAccessLogSinkTypeFile ProxyAccessLogSinkType = "File"
	// ProxyAccessLogSinkTypeOpenTelemetry defines the OpenTelemetry accesslog sink.
	ProxyAccessLogSinkTypeOpenTelemetry ProxyAccessLogSinkType = "OpenTelemetry"
)

type ProxyAccessLogSink struct {
	// Type defines the type of accesslog sink.
	// +kubebuilder:validation:Enum=File;OpenTelemetry
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
	// Empty value disables accesslog.
	Path string `json:"path,omitempty"`
}

// TODO: consider reuse ExtensionService?
type OpenTelemetryEnvoyProxyAccessLog struct {
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

	// TODO: support more OpenTelemetry accesslog options(e.g. TLS, auth etc.) in the future.
}
```

### Example

- The following is an example to disable access log.

```yaml mdox-exec="sed '1,12d' examples/kubernetes/accesslog/disable-accesslog.yaml"
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: disable-accesslog
  namespace: envoy-gateway-system
spec:
  telemetry:
    accessLog:
      disable: true
```

- The following is an example with text format access log.

```yaml mdox-exec="sed '1,12d' examples/kubernetes/accesslog/text-accesslog.yaml"
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: text-access-logging
  namespace: envoy-gateway-system
spec:
  telemetry:
    accessLog:
      settings:
        - format:
            type: Text
            text: |
              [%START_TIME%] "%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%" %RESPONSE_CODE% %RESPONSE_FLAGS% %BYTES_RECEIVED% %BYTES_SENT% %DURATION% "%REQ(X-FORWARDED-FOR)%" "%REQ(USER-AGENT)%" "%REQ(X-REQUEST-ID)%" "%REQ(:AUTHORITY)%" "%UPSTREAM_HOST%"
          sinks:
            - type: File
              file:
                path: /dev/stdout
```

- The following is an example with json format access log.

```yaml mdox-exec="sed '1,12d' examples/kubernetes/accesslog/json-accesslog.yaml"
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: json-access-logging
  namespace: envoy-gateway-system
spec:
  telemetry:
    accessLog:
      settings:
        - format:
          type: JSON
          json:
            status: "%RESPONSE_CODE%"
            message: "%LOCAL_REPLY_BODY%"
      sinks:
        - type: File
          file:
            path: /dev/stdout
```

- The following is an example with OpenTelemetry format access log.

```yaml mdox-exec="sed '1,12d' examples/kubernetes/accesslog/otel-accesslog.yaml"
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: otel-access-logging
  namespace: envoy-gateway-system
spec:
  telemetry:
    accessLog:
      settings:
        - format:
            type: Text
            text: |
              [%START_TIME%] "%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%" %RESPONSE_CODE% %RESPONSE_FLAGS% %BYTES_RECEIVED% %BYTES_SENT% %DURATION% "%REQ(X-FORWARDED-FOR)%" "%REQ(USER-AGENT)%" "%REQ(X-REQUEST-ID)%" "%REQ(:AUTHORITY)%" "%UPSTREAM_HOST%"
          sinks:
            - type: OpenTelemetry
              openTelemetry:
                host: otel-collector.monitoring.svc.cluster.local
                port: 4317
                resources:
                  k8s.cluster.name: "cluster-1"
```

- The following is an example of sending same format to different sinks.

```yaml mdox-exec="sed '1,12d' examples/kubernetes/accesslog/multi-sinks.yaml"
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: multi-sinks
  namespace: envoy-gateway-system
spec:
  telemetry:
    accessLog:
      settings:
        - format:
            type: Text
            text: |
              [%START_TIME%] "%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%" %RESPONSE_CODE% %RESPONSE_FLAGS% %BYTES_RECEIVED% %BYTES_SENT% %DURATION% "%REQ(X-FORWARDED-FOR)%" "%REQ(USER-AGENT)%" "%REQ(X-REQUEST-ID)%" "%REQ(:AUTHORITY)%" "%UPSTREAM_HOST%"
          sinks:
            - type: File
              file:
                path: /dev/stdout
            - type: OpenTelemetry
              openTelemetry:
                host: otel-collector.monitoring.svc.cluster.local
                port: 4317
                resources:
                  k8s.cluster.name: "cluster-1"
```
