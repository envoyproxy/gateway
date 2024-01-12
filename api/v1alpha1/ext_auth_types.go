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
	GRPC ExtAuthServiceType = "GRPC"

	// HTTP external authorization service.
	HTTP ExtAuthServiceType = "HTTP"
)

// ExtAuth defines the configuration for External Authorization.
type ExtAuth struct {
	// Type decides the type of External Authorization.
	// Valid ExtAuthServiceType values are "GRPC" or "HTTP".
	//
	// +unionDiscriminator
	Type ExtAuthServiceType `json:"type"`

	// GRPCService defines the gRPC External Authorization service
	// Only one of GRPCService or HTTPService may be specified.
	GRPCService *GRPCExtAuthService `json:"grpcService,omitempty" yaml:"grpcService"`

	// HTTPService defines the HTTP External Authorization service
	// Only one of GRPCService or HTTPService may be specified.
	HTTPService *HTTPExtAuthService `json:"httpService,omitempty" yaml:"httpService"`

	// AllowWhenFailed indicates whether the request should be allowed when the
	// authorization service has failed or has returned an HTTP 5xx error.
	// When set to true, the request will be allowed.
	// When set to false, the request will be denied with an HTTP 403 Forbidden error.
	// Defaults to false.
	AllowWhenFailed bool `json:"allowWhenFailed,omitempty" yaml:"allowWhenFailed"`

	// AllowedHeaders defines the client request headers that will be included
	// in the request to the external authorization service.
	// Note: If not specified, the default behavior of different external authorization
	// services is different. All headers will be included in the check request
	// to a gRPC authorization server, whereas no headers will be included in the
	// check request to an HTTP authorization server.
	AllowedHeaders []string `json:"allowedHeaders,omitempty" yaml:"allowedHeaders,omitempty"`
}

// GRPCExtAuthService defines the gRPC External Authorization service
type GRPCExtAuthService struct {
	// Host is the hostname of the gRPC External Authorization service
	Host gwapiv1a2.Hostname `json:"host,omitempty" yaml:"host,omitempty"`

	// Port is the network port of the gRPC External Authorization service
	Port gwapiv1a2.PortNumber `json:"port"`

	// TLS defines the TLS configuration for the gRPC External Authorization service.
	// Note: If not specified, the proxy will talk to the gRPC External
	// Authorization service in plaintext.
	TLS *TLSConfig `json:"tlsSettings,omitempty" yaml:"tlsSettings"`
}

// HTTPExtAuthService defines the HTTP External Authorization service
type HTTPExtAuthService struct {
	// URL is the URL of the HTTP External Authorization service.
	// The URL must be a fully qualified URL with a scheme, hostname,
	// and optional port and path. Parameters are not allowed.
	// The URL must use either the http or https scheme.
	// If port is not specified, 80 for http and 443 for https are assumed.
	// If path is specified, the authorization request will be sent to that path,
	// or else the authorization request will be sent to the root path.
	URL string `json:"url" yaml:"url"`

	// TLS defines the TLS configuration for the HTTP External Authorization service.
	// TLS is only valid when the URL scheme is https. If the URL scheme is
	// https, and TLS is not specified, the proxy will use the system default
	// certificate pool to verify the server certificate.
	TLS *TLSConfig `json:"tlsSettings,omitempty" yaml:"tlsSettings"`

	// Authorization response headers that will be added to the original client request
	// before sending it to the upstream server.
	// Note that coexistent headers will be overridden.
	AllowedUpstreamHeaders []string `json:"allowedUpstreamHeaders,omitempty"`
}

// TLSConfig describes a TLS configuration.
type TLSConfig struct {
	// CertificateRefs contains a series of references to Kubernetes objects that
	// contains TLS certificates and private keys. These certificates are used to
	// establish a TLS handshake with the external authorization server.
	//
	// If this field is not specified, the proxy will not present a client certificate
	// and will use the system default certificate pool to verify the server certificate.
	// +optional
	// +kubebuilder:validation:MaxItems=64
	CertificateRefs []gwapiv1a2.SecretObjectReference `json:"certificateRefs,omitempty"`
}

// AuthorizationResponse defines how the authorization response impacts the client request and response.
type AuthorizationResponse struct {
	// Authorization response headers that will be added to the original client request.
	// Note that coexistent headers will be overridden.
	AllowedUpstreamHeaders []string `json:"allowedUpstreamHeaders,omitempty"`
}
