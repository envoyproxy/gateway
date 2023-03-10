// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatewayv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

const (
	// KindAccessLoggingPolicy is the name of the AccessLoggingPolicy kind.
	KindAccessLoggingPolicy = "AccessLoggingPolicy"
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster

// AccessLoggingPolicy allows the user to configure AccessLogging for Listener.
type AccessLoggingPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of AccessLoggingPolicy.
	Spec AccessLoggingPolicySpec `json:"spec"`
}

// AccessLoggingPolicySpec defines the desired state of AccessLoggingPolicy.
type AccessLoggingPolicySpec struct {
	// TargetRef identifies an API object to apply policy to.
	TargetRef gatewayv1a2.PolicyTargetReference `json:"targetRef"`
	Filters   []AccessLoggingFilter             `json:"filters,omitempty"`
	// Text defines text based access logs.
	// +optional
	Text *TextFileEnvoyProxyAccessLog `json:"text,omitempty"`
	// JSON defines structured json based access logs.
	// +optional
	JSON *JSONFileEnvoyProxyAccessLog `json:"json,omitempty"`
	// Otel defines configuration for OpenTelemetry log provider.
	// +optional
	Otel *OpenTelemetryEnvoyProxyAccessLog `json:"otel,omitempty"`
}

type AccessLoggingFilter struct {
	// Expression is a valid CEL expression for selecting when requests/connections should be logged.
	//
	// Examples:
	// ```
	// response.code >= 400
	// request.url_path.contains('v1beta3')
	// ```
	Expression string `json:"expression"`
}

type TextFileEnvoyProxyAccessLog struct {
	// Path defines the file path used to expose envoy access log(e.g. /dev/stdout).
	// Empty value disables access logging.
	Path string `json:"path"`
	// Format is the format for the proxy access log, following Envoy access logging formatting, empty value results in proxy's default access log format.
	// Envoy [command operators](https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#command-operators) may be used in the format.
	// The [format string documentation](https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#config-access-log-format-strings) provides more information.
	// +optional
	Format string `json:"format,omitempty"`
}

type JSONFileEnvoyProxyAccessLog struct {
	// Path defines the file path used to expose envoy access log(e.g. /dev/stdout).
	// Empty value disables access logging.
	Path string `json:"path"`
	// Fields is additional attributes that describe the specific event occurrence.
	// Structured format for the envoy access logs. Envoy [command operators](https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#command-operators)
	// can be used as values for fields within the Struct.
	Fields map[string]string `json:"fields"`
}

type OpenTelemetryEnvoyProxyAccessLog struct {
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

//+kubebuilder:object:root=true

// AccessLoggingPolicyList contains a list of AccessLoggingPolicy resources.
type AccessLoggingPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AccessLoggingPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AccessLoggingPolicy{}, &AccessLoggingPolicyList{})
}
