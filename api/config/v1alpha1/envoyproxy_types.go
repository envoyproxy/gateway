// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// EnvoyProxy is the Schema for the envoyproxies API
type EnvoyProxy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EnvoyProxySpec   `json:"spec,omitempty"`
	Status EnvoyProxyStatus `json:"status,omitempty"`
}

// EnvoyProxySpec defines the desired state of EnvoyProxy.
type EnvoyProxySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - define desired state of cluster.
	// Important: Run "make" to regenerate code after modifying this file.

	// AccessLogs defines the access log configuration for envoy listener/filter class.
	AccessLogs []EnvoyProxyAccessLog `json:"accessLogs,omitempty"`
}

// AccessLogs defines the access log configuration for envoy listener class.
type EnvoyProxyAccessLog struct {
	// Text defines text based access logs.
	// +optional
	Text *TextFileEnvoyProxyAccessLog `json:"text,omitempty"`
	// Json defines structured json based access logs.
	// +optional
	Json *JsonFileEnvoyProxyAccessLog `json:"json,omitempty"`
	// Otel defines configuration for OpenTelemetry log provider.
	// +optional
	Otel *OpenTelemetryEnvoyProxyAccessLog `json:"otel,omitempty"`
}

type TextFileEnvoyProxyAccessLog struct {
	// Path defines the file path used to expose envoy access log, empty value results in default `/dev/stdout`.
	// +optional
	Path string `json:"path,omitempty"`
	// Format for envoy text defaulat access logs, empty value results in default access log format.
	// +optional
	Format string `json:"format,omitempty"`
}

type JsonFileEnvoyProxyAccessLog struct {
	// Path defines the file path used to expose envoy access log, empty value results in default `/dev/stdout`.
	// +optional
	Path string `json:"path,omitempty"`
	// Fields for envoy structured json based access logs.
	// +optional
	Fields map[string]string `json:"fields"`
}

type OpenTelemetryEnvoyProxyAccessLog struct {
	// Specifies the service that implements the Envoy ALS gRPC authorization service.
	//
	// Example: "otel-collector.monitoring.svc.cluster.local".
	Service string `json:"service,omitempty"`
	// Specifies the port of the service.
	Port uint32 `json:"port,omitempty"`
	// The friendly name of the access log, empty value results in default `otel_envoy_accesslog`.
	// +optional
	LogName string `json:"logName,omitempty"`
	// OpenTelemetry `Resource <https://github.com/open-telemetry/opentelemetry-proto/blob/main/opentelemetry/proto/logs/v1/logs.proto#L51>`_
	// attributes are filled with Envoy node info.
	//
	// Example:
	// ```
	// resources:
	//
	//	k8s.pod.name: "sample-pod-xxxxx"
	//	k8s.pod.namespace: "default"
	//
	// ```
	// +optional
	Resources map[string]string `json:"resources,omitempty"`
	// Format for the proxy access log, following Envoy access logging formatting, empty value results in proxy's default access log format.
	// Envoy [command operators](https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#command-operators) may be used in the format.
	// The [format string documentation](https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#config-access-log-format-strings) provides more information.
	// Alias to `body` filed in [Open Telemetry](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/access_loggers/open_telemetry/v3/logs_service.proto)
	//
	// Example: `text: "%LOCAL_REPLY_BODY%:%RESPONSE_CODE%:path=%REQ(:path)%"`
	// +optional
	Text string `json:"text,omitempty"`
	// Additional attributes that describe the specific event occurrence.
	// Structured format for the envoy access logs. Envoy [command operators](https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#command-operators)
	// can be used as values for fields within the Struct.
	// Alias to `attributes` filed in [Open Telemetry](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/access_loggers/open_telemetry/v3/logs_service.proto)
	//
	// Example:
	// ```
	// fileds:
	//
	//	status: "%RESPONSE_CODE%"
	//	message: "%LOCAL_REPLY_BODY%"
	//
	// ```
	// +optional
	Fields map[string]string `json:"fields,omitempty"`
}

// EnvoyProxyStatus defines the observed state of EnvoyProxy
type EnvoyProxyStatus struct {
	// INSERT ADDITIONAL STATUS FIELDS - define observed state of cluster.
	// Important: Run "make" to regenerate code after modifying this file.
}

//+kubebuilder:object:root=true

// EnvoyProxyList contains a list of EnvoyProxy
type EnvoyProxyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EnvoyProxy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EnvoyProxy{}, &EnvoyProxyList{})
}
