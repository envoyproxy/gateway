// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// HealthCheck defines the health check configuration.
type HealthCheck struct {
	// Timeout defines the time to wait for a health check response.
	//
	// +kubebuilder:validation:Format=duration
	// +kubebuilder:default="1s"
	// +optional
	Timeout *metav1.Duration `json:"timeout"`

	// Interval defines the time between health checks.
	//
	// +kubebuilder:validation:Format=duration
	// +kubebuilder:default="3s"
	// +optional
	Interval *metav1.Duration `json:"interval"`

	// UnhealthyThreshold defines the number of unhealthy health checks required before a backend host is marked unhealthy.
	//
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=3
	// +optional
	UnhealthyThreshold *uint32 `json:"unhealthyThreshold"`

	// HealthyThreshold defines the number of healthy health checks required before a backend host is marked healthy.
	//
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=1
	// +optional
	HealthyThreshold *uint32 `json:"healthyThreshold"`

	// HealthChecker defines the concrete health checker to do health checking.
	HealthChecker HealthChecker `json:"healthChecker"`
}

// HealthCheckerType is the type of health checker.
// +kubebuilder:validation:Enum=HTTP;GRPC;TCP
type HealthCheckerType string

const (
	// HealthCheckerTypeHTTP defines the HTTP type of health checking.
	HealthCheckerTypeHTTP HealthCheckerType = "HTTP"
	// HealthCheckerTypeGRPC defines the GRPC type of health checking.
	HealthCheckerTypeGRPC HealthCheckerType = "GRPC"
	// HealthCheckerTypeTCP defines the TCP type of health checking.
	HealthCheckerTypeTCP HealthCheckerType = "TCP"
)

// HealthChecker defines the configuration of concrete health checker.
// EG supports various types of health checking including HTTP, GRPC, TCP.
// +union
//
// +kubebuilder:validation:XValidation:rule="self.type == 'HTTP' ? has(self.http) : !has(self.http)",message="If Health Checker type is HTTP, http field needs to be set."
// +kubebuilder:validation:XValidation:rule="self.type == 'GRPC' ? has(self.grpc) : !has(self.grpc)",message="If Health Checker type is GRPC, grpc field needs to be set."
// +kubebuilder:validation:XValidation:rule="self.type == 'TCP' ? has(self.tcp) : !has(self.tcp)",message="If Health Checker type is TCP, tcp field needs to be set."
type HealthChecker struct {
	// Type defines the type of health checker.
	// +kubebuilder:validation:Enum=HTTP;GRPC;TCP
	// +unionDiscriminator
	Type HealthCheckerType `json:"type" yaml:"type"`
	// HTTP defines the configuration of http health checker.
	// It's required while the health checker type is HTTP.
	// +optional
	HTTP *HTTPHealthChecker `json:"http,omitempty" yaml:"http,omitempty"`
	// GRPC defines the configuration of grpc health checker.
	// It's required while the health checker type is GRPC.
	// +optional
	GRPC *GRPCHealthChecker `json:"grpc,omitempty" yaml:"grpc,omitempty"`
	// TCP defines the configuration of tcp health checker.
	// It's required while the health checker type is TCP.
	// +optional
	TCP *TCPHealthChecker `json:"tcp,omitempty" yaml:"tcp,omitempty"`
}

// HTTPHealthChecker defines the settings of http health check.
type HTTPHealthChecker struct {
	// Path defines the HTTP path that will be requested during health checking.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=1024
	Path string `json:"path" yaml:"path"`
	// Method defines the HTTP method used for health checking.
	// Defaults to GET
	// +optional
	Method *string `json:"method,omitempty" yaml:"method,omitempty"`
	// ExpectedStatuses defines a list of HTTP response statuses considered healthy.
	// +optional
	ExpectedStatuses []HTTPStatusRange `json:"expectedStatuses,omitempty" yaml:"expectedStatuses,omitempty"`
	// ExpectedResponses defines a list of HTTP expected responses to match.
	// +optional
	ExpectedResponses []HealthCheckPayload `json:"expectedResponses,omitempty" yaml:"expectedResponses,omitempty"`
}

// GRPCHealthChecker defines the settings of grpc health check.
type GRPCHealthChecker struct {
	// ServiceName defines the value of parameter service in the gRPC health check request.
	// +optional
	ServiceName *string `json:"serviceName,omitempty" yaml:"omitempty"`
	// Authority defines the value of :authority header in the gRPC health check request.
	// +optional
	Authority *string `json:"authority,omitempty" yaml:"authority,omitempty"`
	// +optional
	Metadata map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// TCPHealthChecker defines the settings of tcp health check.
type TCPHealthChecker struct {
	// Send defines the request payload.
	// +optional
	Send *HealthCheckPayload `json:"send,omitempty" yaml:"send,omitempty"`
	// Receive defines the expected response payloads.
	// +optional
	Receive []HealthCheckPayload `json:"receive,omitempty" yaml:"receive,omitempty"`
}

// HTTPStatusRange defines the start and end of the http status range using half-open interval semantics [start, end). Only statuses in the range [100, 600) are allowed.
// +kubebuilder:validation:XValidation:rule="self.start <= self.end",message="start should be not greater than end"
type HTTPStatusRange struct {
	// Start defines start of the range (inclusive)
	// +kubebuilder:validation:Minimum=100
	// +kubebuilder:validation:Maximum=600
	// +kubebuilder:validation:ExclusiveMaximum=true
	Start int64 `json:"start" yaml:"start"`
	// End defines end of the range (exclusive)
	// +kubebuilder:validation:Minimum=100
	// +kubebuilder:validation:Maximum=600
	End int64 `json:"end" yaml:"end"`
}

// HealthCheckPayloadType is the type of the payload.
// +kubebuilder:validation:Enum=Text;Binary
type HealthCheckPayloadType string

const (
	// HealthCheckPayloadTypeText defines the Text type payload.
	HealthCheckPayloadTypeText HealthCheckPayloadType = "Text"
	// HealthCheckPayloadTypeBinary defines the Binary type payload.
	HealthCheckPayloadTypeBinary HealthCheckPayloadType = "Binary"
)

// HealthCheckPayload defines the encoding of the payload bytes in the payload.
// +union
// +kubebuilder:validation:XValidation:rule="self.type == 'Text' ? has(self.text) : !has(self.text)",message="If payload type is Text, text field needs to be set."
// +kubebuilder:validation:XValidation:rule="self.type == 'Binary' ? has(self.binary) : !has(self.binary)",message="If payload type is Binary, binary field needs to be set."
type HealthCheckPayload struct {
	// Type defines the type of the payload.
	// +kubebuilder:validation:Enum=Text;Binary
	// +unionDiscriminator
	Type HealthCheckPayloadType `json:"type" yaml:"type"`
	// Text payload in plain text.
	// +optional
	Text *string `json:"text,omitempty" yaml:"text,omitempty"`
	// Binary payload base64 encoded.
	// +optional
	Binary []byte `json:"binary,omitempty" yaml:"binary,omitempty"`
}
