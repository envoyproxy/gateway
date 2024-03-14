// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// +kubebuilder:validation:XValidation:rule="(has(self.grpc) || has(self.http))",message="one of grpc or http must be specified"
// +kubebuilder:validation:XValidation:rule="(has(self.grpc) && !has(self.http)) || (!has(self.grpc) && has(self.http))",message="only one of grpc or http can be specified"
// +kubebuilder:validation:XValidation:rule="has(self.grpc) ? (!has(self.grpc.backendRef.group) || self.grpc.backendRef.group == \"\") : true", message="group is invalid, only the core API group (specified by omitting the group field or setting it to an empty string) is supported"
// +kubebuilder:validation:XValidation:rule="has(self.grpc) ? (!has(self.grpc.backendRef.kind) || self.grpc.backendRef.kind == 'Service') : true", message="kind is invalid, only Service (specified by omitting the kind field or setting it to 'Service') is supported"
// +kubebuilder:validation:XValidation:rule="has(self.http) ? (!has(self.http.backendRef.group) || self.http.backendRef.group == \"\") : true", message="group is invalid, only the core API group (specified by omitting the group field or setting it to an empty string) is supported"
// +kubebuilder:validation:XValidation:rule="has(self.http) ? (!has(self.http.backendRef.kind) || self.http.backendRef.kind == 'Service') : true", message="kind is invalid, only Service (specified by omitting the kind field or setting it to 'Service') is supported"
//
// ExtAuth defines the configuration for External Authorization.
type ExtAuth struct {
	// GRPC defines the gRPC External Authorization service.
	// Either GRPCService or HTTPService must be specified,
	// and only one of them can be provided.
	GRPC *GRPCExtAuthService `json:"grpc,omitempty"`

	// HTTP defines the HTTP External Authorization service.
	// Either GRPCService or HTTPService must be specified,
	// and only one of them can be provided.
	HTTP *HTTPExtAuthService `json:"http,omitempty"`

	// HeadersToExtAuth defines the client request headers that will be included
	// in the request to the external authorization service.
	// Note: If not specified, the default behavior for gRPC and HTTP external
	// authorization services is different due to backward compatibility reasons.
	// All headers will be included in the check request to a gRPC authorization server.
	// Only the following headers will be included in the check request to an HTTP
	// authorization server: Host, Method, Path, Content-Length, and Authorization.
	// And these headers will always be included to the check request to an HTTP
	// authorization server by default, no matter whether they are specified
	// in HeadersToExtAuth or not.
	// +optional
	HeadersToExtAuth []string `json:"headersToExtAuth,omitempty"`

	// FailOpen is a switch used to control the behavior when a response from the External Authorization service cannot be obtained.
	// If FailOpen is set to true, the system allows the traffic to pass through.
	// Otherwise, if it is set to false or not set (defaulting to false),
	// the system blocks the traffic and returns a HTTP 5xx error, reflecting a fail-closed approach.
	// This setting determines whether to prioritize accessibility over strict security in case of authorization service failure.
	//
	// +optional
	// +kubebuilder:default=false
	FailOpen *bool `json:"failOpen,omitempty"`
}

// GRPCExtAuthService defines the gRPC External Authorization service
// The authorization request message is defined in
// https://www.envoyproxy.io/docs/envoy/latest/api-v3/service/auth/v3/external_auth.proto
type GRPCExtAuthService struct {
	// BackendRef references a Kubernetes object that represents the
	// backend server to which the authorization request will be sent.
	// Only service Kind is supported for now.
	BackendRef gwapiv1.BackendObjectReference `json:"backendRef"`
}

// HTTPExtAuthService defines the HTTP External Authorization service
type HTTPExtAuthService struct {
	// BackendRef references a Kubernetes object that represents the
	// backend server to which the authorization request will be sent.
	// Only service Kind is supported for now.
	BackendRef gwapiv1.BackendObjectReference `json:"backendRef"`

	// Path is the path of the HTTP External Authorization service.
	// If path is specified, the authorization request will be sent to that path,
	// or else the authorization request will be sent to the root path.
	Path *string `json:"path,omitempty"`

	// HeadersToBackend are the authorization response headers that will be added
	// to the original client request before sending it to the backend server.
	// Note that coexisting headers will be overridden.
	// If not specified, no authorization response headers will be added to the
	// original client request.
	// +optional
	HeadersToBackend []string `json:"headersToBackend,omitempty"`
}
