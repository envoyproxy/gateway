// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

// ExtAuthServiceType specifies the types of External Authorization.
// +kubebuilder:validation:Enum=GRPC;HTTP
type ExtAuthServiceType string

const (
	// GRPC external authorization service.
	GRPCExtAuthServiceType ExtAuthServiceType = "GRPC"

	// HTTP external authorization service.
	HTTPExtAuthServiceType ExtAuthServiceType = "HTTP"
)

// +kubebuilder:validation:XValidation:message="http must be specified if type is HTTP",rule="self.type == 'HTTP' ? has(self.http) : true"
// +kubebuilder:validation:XValidation:message="grpc must be specified if type is GRPC",rule="self.type == 'GRPC' ? has(self.grpc) : true"
// +kubebuilder:validation:XValidation:message="only one of grpc or http can be specified",rule="!(has(self.grpc) && has(self.http))"
//
// ExtAuth defines the configuration for External Authorization.
type ExtAuth struct {
	// Type decides the type of External Authorization.
	// Valid ExtAuthServiceType values are "GRPC" or "HTTP".
	// +kubebuilder:validation:Enum=GRPC;HTTP
	// +unionDiscriminator
	Type ExtAuthServiceType `json:"type"`

	// GRPC defines the gRPC External Authorization service.
	// Only one of GRPCService or HTTPService may be specified.
	GRPC *GRPCExtAuthService `json:"grpc,omitempty"`

	// HTTP defines the HTTP External Authorization service.
	// Only one of GRPCService or HTTPService may be specified.
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
}

// GRPCExtAuthService defines the gRPC External Authorization service
// The authorization request message is defined in
// https://www.envoyproxy.io/docs/envoy/latest/api-v3/service/auth/v3/external_auth.proto
type GRPCExtAuthService struct {
	// Host is the hostname of the gRPC External Authorization service.
	Host gwapiv1a2.PreciseHostname `json:"host"`

	// Port is the network port of the gRPC External Authorization service.
	Port gwapiv1a2.PortNumber `json:"port"`

	// TLS defines the TLS configuration for the gRPC External Authorization service.
	// Note: If not specified, the proxy will talk to the gRPC External Authorization
	// service in plaintext.
	// +optional
	TLS *TLSConfig `json:"tls,omitempty"`
}

// HTTPExtAuthService defines the HTTP External Authorization service
type HTTPExtAuthService struct {
	// Host is the hostname of the HTTP External Authorization service.
	Host gwapiv1a2.PreciseHostname `json:"host"`

	// Port is the network port of the HTTP External Authorization service.
	// If port is not specified, 80 for http and 443 for https are assumed.
	Port *gwapiv1a2.PortNumber `json:"port,omitempty"`

	// Path is the path of the HTTP External Authorization service.
	// If path is specified, the authorization request will be sent to that path,
	// or else the authorization request will be sent to the root path.
	Path *string `json:"path,omitempty"`

	// TLS defines the TLS configuration for the HTTP External Authorization service.
	// Note: If not specified, the proxy will talk to the HTTP External Authorization
	// service in plaintext.
	// +optional
	TLS *TLSConfig `json:"tls,omitempty"`

	// HeadersToBackend are the authorization response headers that will be added
	// to the original client request before sending it to the backend server.
	// Note that coexisting headers will be overridden.
	// If not specified, no authorization response headers will be added to the
	// original client request.
	// +optional
	HeadersToBackend []string `json:"headersToBackend,omitempty"`
}

// TLSConfig describes a TLS configuration.
type TLSConfig struct {
}
