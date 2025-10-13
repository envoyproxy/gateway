// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

// ExtAuth defines the configuration for External Authorization.
//
// +kubebuilder:validation:XValidation:rule="(has(self.grpc) || has(self.http))",message="one of grpc or http must be specified"
// +kubebuilder:validation:XValidation:rule="(has(self.grpc) && !has(self.http)) || (!has(self.grpc) && has(self.http))",message="only one of grpc or http can be specified"
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

	// BodyToExtAuth defines the Body to Ext Auth configuration.
	// +optional
	BodyToExtAuth *BodyToExtAuth `json:"bodyToExtAuth,omitempty"`

	// Timeout defines the timeout for requests to the external authorization service.
	// If not specified, defaults to 10 seconds.
	// +optional
	Timeout *gwapiv1.Duration `json:"timeout,omitempty"`

	// FailOpen is a switch used to control the behavior when a response from the External Authorization service cannot be obtained.
	// If FailOpen is set to true, the system allows the traffic to pass through.
	// Otherwise, if it is set to false or not set (defaulting to false),
	// the system blocks the traffic and returns a HTTP 5xx error, reflecting a fail-closed approach.
	// This setting determines whether to prioritize accessibility over strict security in case of authorization service failure.
	//
	// If set to true, the External Authorization will also be bypassed if its configuration is invalid.
	//
	// +optional
	// +kubebuilder:default=false
	FailOpen *bool `json:"failOpen,omitempty"`

	// RecomputeRoute clears the route cache and recalculates the routing decision.
	// This field must be enabled if the headers added or modified by the ExtAuth are used for
	// route matching decisions. If the recomputation selects a new route, features targeting
	// the new matched route will be applied.
	//
	// +optional
	RecomputeRoute *bool `json:"recomputeRoute,omitempty"`
}

// GRPCExtAuthService defines the gRPC External Authorization service
// The authorization request message is defined in
// https://www.envoyproxy.io/docs/envoy/latest/api-v3/service/auth/v3/external_auth.proto
// +kubebuilder:validation:XValidation:message="backendRef or backendRefs needs to be set",rule="has(self.backendRef) || self.backendRefs.size() > 0"
// +kubebuilder:validation:XValidation:message="BackendRefs only supports Service, ServiceImport, and Backend kind.",rule="has(self.backendRefs) ? self.backendRefs.all(f, f.kind == 'Service' || f.kind == 'ServiceImport' || f.kind == 'Backend') : true"
// +kubebuilder:validation:XValidation:message="BackendRefs only supports Core, multicluster.x-k8s.io, and gateway.envoyproxy.io groups.",rule="has(self.backendRefs) ? (self.backendRefs.all(f, f.group == \"\" || f.group == 'multicluster.x-k8s.io' || f.group == 'gateway.envoyproxy.io')) : true"
type GRPCExtAuthService struct {
	// Only Service kind is supported for now.
	BackendCluster `json:",inline"`
}

// HTTPExtAuthService defines the HTTP External Authorization service
//
// +kubebuilder:validation:XValidation:message="backendRef or backendRefs needs to be set",rule="has(self.backendRef) || self.backendRefs.size() > 0"
// +kubebuilder:validation:XValidation:message="BackendRefs only supports Service, ServiceImport, and Backend kind.",rule="has(self.backendRefs) ? self.backendRefs.all(f, f.kind == 'Service' || f.kind == 'ServiceImport' || f.kind == 'Backend') : true"
// +kubebuilder:validation:XValidation:message="BackendRefs only supports Core, multicluster.x-k8s.io, and gateway.envoyproxy.io groups.",rule="has(self.backendRefs) ? (self.backendRefs.all(f, f.group == \"\" || f.group == 'multicluster.x-k8s.io' || f.group == 'gateway.envoyproxy.io')) : true"
type HTTPExtAuthService struct {
	// Only Service kind is supported for now.
	BackendCluster `json:",inline"`

	// Path is the path of the HTTP External Authorization service.
	// If path is specified, the authorization request will be sent to that path,
	// or else the authorization request will use the path of the original request.
	//
	// Please note that the original request path will be appended to the path specified here.
	// For example, if the original request path is "/hello", and the path specified here is "/auth",
	// then the path of the authorization request will be "/auth/hello". If the path is not specified,
	// the path of the authorization request will be "/hello".
	// +optional
	Path *string `json:"path,omitempty"`

	// HeadersToBackend are the authorization response headers that will be added
	// to the original client request before sending it to the backend server.
	// Note that coexisting headers will be overridden.
	// If not specified, no authorization response headers will be added to the
	// original client request.
	// +optional
	HeadersToBackend []string `json:"headersToBackend,omitempty"`
}

// BodyToExtAuth defines the Body to Ext Auth configuration
type BodyToExtAuth struct {
	// MaxRequestBytes is the maximum size of a message body that the filter will hold in memory.
	// Envoy will return HTTP 413 and will not initiate the authorization process when buffer
	// reaches the number set in this field.
	// Note that this setting will have precedence over failOpen mode.
	//
	// +kubebuilder:validation:Minimum=1
	MaxRequestBytes uint32 `json:"maxRequestBytes"`
}
