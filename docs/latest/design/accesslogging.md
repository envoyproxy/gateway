# Observability: Access Logging

## Overview

Envoy supports extensible access logging to different sinks, File, gRPC etc.
Envoy supports customizable access log formats using predefined fields as well as arbitrary HTTP request and response headers.
Envoy supports several built-in access log filters and extension filters that are registered at runtime.

Envoy Gateway leverages [Gateway API][] for configuring managed Envoy proxies. Gateway API defines core, extended, and
implementation-specific API [support levels][] for implementors such as Envoy Gateway to expose features. 
Since access logging is not covered by `Core` or `Extended` APIs, EG should provide an easy to config access log formats and sinks per `EnvoyProxy`.

## Goals

- Support send access logging to `File` or `OpenTelemetry` backend
- Support access log filters base on [CEL][] expression

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
	// Text defines text based access logs.
	// +optional
	Text *TextFileEnvoyProxyAccessLogging `json:"text,omitempty"`
	// Json defines structured json based access logs.
	// +optional
	Json *JsonFileEnvoyProxyAccessLogging `json:"json,omitempty"`
	// Otel defines configuration for OpenTelemetry log provider.
	// +optional
	Otel *OpenTelemetryEnvoyProxyAccessLogging `json:"otel,omitempty"`
}

type TextFileEnvoyProxyAccessLogging struct {
	// Path defines the file path used to expose envoy access log(e.g. /dev/stdout).
	// Empty value disables access logging.
	Path string `json:"path,omitempty"`
	// Format is the format for the proxy access log, following Envoy access logging formatting, empty value results in proxy's default access log format.
	// Envoy [command operators](https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#command-operators) may be used in the format.
	// The [format string documentation](https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#config-access-log-format-strings) provides more information.
	Format string `json:"format,omitempty"`
}

type JsonFileEnvoyProxyAccessLogging struct {
	// Path defines the file path used to expose envoy access log(e.g. /dev/stdout).
	// Empty value disables access logging.
	Path string `json:"path,omitempty"`
	// Fields is additional attributes that describe the specific event occurrence.
	// Structured format for the envoy access logs. Envoy [command operators](https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#command-operators)
	// can be used as values for fields within the Struct.
	// +optional
	Fields map[string]string `json:"fields"`
}

type OpenTelemetryEnvoyProxyAccessLogging struct {
	// Service specifies the service that implements the Envoy ALS gRPC authorization service.
	//
	// Example: "otel-collector.monitoring.svc.cluster.local".
	Service string `json:"service,omitempty"`
	// Port specifies the port of the service.
	Port uint32 `json:"port,omitempty"`
	// LogName is the friendly name of the access log, empty value results in default `otel_envoy_accesslog`.
	// +optional
	LogName string `json:"logName,omitempty"`
	// Resources is a set of labels that describe the source of a log entry, including envoy node info.
	// It's recommended to follow [semantic conventions](https://opentelemetry.io/docs/reference/specification/resource/semantic_conventions/).
	//
	//
	// Example:
	// ```
	// resources:
	//
	//	k8s.cluster.name: "cluster-xxxx"
	//
	// ```
	// +optional
	Resources map[string]string `json:"resources,omitempty"`
	// Text is the format for the proxy access log, following Envoy access logging formatting, empty value results in proxy's default access log format.
	// Envoy [command operators](https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#command-operators) may be used in the format.
	// The [format string documentation](https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#config-access-log-format-strings) provides more information.
	// Alias to `body` filed in [Open Telemetry](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/access_loggers/open_telemetry/v3/logs_service.proto)
	//
	// Example: `text: "%LOCAL_REPLY_BODY%:%RESPONSE_CODE%:path=%REQ(:path)%"`
	// +optional
	Text string `json:"text,omitempty"`
	// Fields is additional attributes that describe the specific event occurrence.
	// Structured format for the envoy access logs. Envoy [command operators](https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#command-operators)
	// can be used as values for fields within the Struct.
	// Alias to `attributes` filed in [Open Telemetry](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/access_loggers/open_telemetry/v3/logs_service.proto)
	//
	// Example:
	// ```
	// fields:
	//
	//	status: "%RESPONSE_CODE%"
	//	message: "%LOCAL_REPLY_BODY%"
	//
	// ```
	// +optional
	Fields map[string]string `json:"fields,omitempty"`
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
  - text:
      format: |
		[%START_TIME%] "%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%" %RESPONSE_CODE% %RESPONSE_FLAGS% %BYTES_RECEIVED% %BYTES_SENT% %DURATION% "%REQ(X-FORWARDED-FOR)%" "%REQ(USER-AGENT)%" "%REQ(X-REQUEST-ID)%" "%REQ(:AUTHORITY)%" "%UPSTREAM_HOST%"
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
  - json:
      fields:
        status: "%RESPONSE_CODE%"
        message: "%LOCAL_REPLY_BODY%"
      path: /dev/stdout
```

1. The following is an example with OpenTelemetry format access log.

```yaml
apiVersion: config.gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: otel-access-logging
spec:
  - otel:
      service: otel-collector.monitoring.svc.cluster.local
      port: 4317
      resources:
        k8s.cluster.name: "cluster-1"
      text: |
        [%START_TIME%] "%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%" %RESPONSE_CODE% %RESPONSE_FLAGS% %BYTES_RECEIVED% %BYTES_SENT% %DURATION% "%REQ(X-FORWARDED-FOR)%" "%REQ(USER-AGENT)%" "%REQ(X-REQUEST-ID)%" "%REQ(:AUTHORITY)%" "%UPSTREAM_HOST%"
      fields:
        status: "%RESPONSE_CODE%"
        message: "%LOCAL_REPLY_BODY%"
```

1. The following is an example with multi providers.

```yaml
apiVersion: config.gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: multi-providers
spec:
  - text:
      format: |
		[%START_TIME%] "%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%" %RESPONSE_CODE% %RESPONSE_FLAGS% %BYTES_RECEIVED% %BYTES_SENT% %DURATION% "%REQ(X-FORWARDED-FOR)%" "%REQ(USER-AGENT)%" "%REQ(X-REQUEST-ID)%" "%REQ(:AUTHORITY)%" "%UPSTREAM_HOST%"
      path: /dev/stdout
  - otel:
      service: otel-collector.monitoring.svc.cluster.local
      port: 4317
      resources:
        k8s.cluster.name: "cluster-1"
      text: |
        [%START_TIME%] "%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%" %RESPONSE_CODE% %RESPONSE_FLAGS% %BYTES_RECEIVED% %BYTES_SENT% %DURATION% "%REQ(X-FORWARDED-FOR)%" "%REQ(USER-AGENT)%" "%REQ(X-REQUEST-ID)%" "%REQ(:AUTHORITY)%" "%UPSTREAM_HOST%"
      fields:
        status: "%RESPONSE_CODE%"
        message: "%LOCAL_REPLY_BODY%"
```

1. The following is an example show how to disabled access log.

```yaml
apiVersion: config.gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: text-access-logging
spec:
  accessLoggings:
  - text:
      path: /dev/null
```

1. The following is an example with text format access log, only logging when response code is 400.

```yaml
apiVersion: config.gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: text-access-logging
spec:
  accessLoggings:
  - filter:
      expression: response.code == 400
    text:
      format: |
		[%START_TIME%] "%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%" %RESPONSE_CODE% %RESPONSE_FLAGS% %BYTES_RECEIVED% %BYTES_SENT% %DURATION% "%REQ(X-FORWARDED-FOR)%" "%REQ(USER-AGENT)%" "%REQ(X-REQUEST-ID)%" "%REQ(:AUTHORITY)%" "%UPSTREAM_HOST%"
      path: /dev/stdout
```

1. The following is an example with multi providers, but send to open telemetry provider when response code is greater than 500.

```yaml
apiVersion: config.gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: multi-providers
spec:
  - text:
      format: |
		[%START_TIME%] "%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%" %RESPONSE_CODE% %RESPONSE_FLAGS% %BYTES_RECEIVED% %BYTES_SENT% %DURATION% "%REQ(X-FORWARDED-FOR)%" "%REQ(USER-AGENT)%" "%REQ(X-REQUEST-ID)%" "%REQ(:AUTHORITY)%" "%UPSTREAM_HOST%"
      path: /dev/stdout
  - filter:
      expression: response.code > 500
    otel:
      service: otel-collector.monitoring.svc.cluster.local
      port: 4317
      resources:
        k8s.cluster.name: "cluster-1"
      text: |
        [%START_TIME%] "%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%" %RESPONSE_CODE% %RESPONSE_FLAGS% %BYTES_RECEIVED% %BYTES_SENT% %DURATION% "%REQ(X-FORWARDED-FOR)%" "%REQ(USER-AGENT)%" "%REQ(X-REQUEST-ID)%" "%REQ(:AUTHORITY)%" "%UPSTREAM_HOST%"
      fields:
        status: "%RESPONSE_CODE%"
        message: "%LOCAL_REPLY_BODY%"
```

[Gateway API]: https://gateway-api.sigs.k8s.io/
[support levels]: https://gateway-api.sigs.k8s.io/concepts/conformance/?h=extended#2-support-levels
[CEL]: https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/access_loggers/filters/cel/v3/cel.proto#extension-envoy-access-loggers-extension-filters-cel
[HttpGrpcAccessLogConfig]: https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/access_loggers/grpc/v3/als.proto#extensions-access-loggers-grpc-v3-httpgrpcaccesslogconfig
[TcpGrpcAccessLogConfig]: https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/access_loggers/grpc/v3/als.proto#extensions-access-loggers-grpc-v3-tcpgrpcaccesslogconfig
