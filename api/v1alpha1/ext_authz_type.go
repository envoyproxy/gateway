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

// ExtAuthz defines the configuration for External Authorization.
type ExtAuthz struct {
	// Type decides the type of External Authorization.
	// Valid ExtAuthServiceType values are "GRPC" or "HTTP".
	//
	// +unionDiscriminator
	Type ExtAuthServiceType `json:"type"`

	// GRPCService defines the gRPC External Authorization service
	// Only one of GRPCService or HTTPService may be specified.
	GRPCService *GRPCService `json:"grpcService,omitempty" yaml:"grpcService"`

	// HTTPService defines the HTTP External Authorization service
	// Only one of GRPCService or HTTPService may be specified.
	HTTPService *HTTPService `json:"httpService,omitempty" yaml:"httpService"`
}

// GRPCService defines the gRPC External Authorization service
type GRPCService struct {
	// Host ist the hostname of the gRPC External Authorization service
	Host gwapiv1a2.Hostname `json:"host,omitempty" yaml:"host,omitempty"`

	// Port is the network port of the gRPC External Authorization service
	Port gwapiv1a2.PortNumber `json:"port"`

	// TLS defines the TLS configuration for the gRPC External Authorization service
	TLS *TLSConfig `json:"tlsSettings,omitempty" yaml:"tlsSettings"`
}

// HTTPService defines the HTTP External Authorization service
type HTTPService struct {
	// URL is the URL of the HTTP External Authorization service.
	// The URL must be a fully qualified URL with a scheme, hostname, and optional port and path. Parameters are not allowed.
	// The URL must use either the http or https scheme.
	// If port is not specified, 80 for http and 443 for https are assumed.
	// If path is specified, the authorization request will be sent to the path.
	URl string `json:"url" yaml:"url"`

	// TLS defines the TLS configuration for the HTTP External Authorization service.
	TLS *TLSConfig `json:"tlsSettings,omitempty" yaml:"tlsSettings"`

	// AuthorizationRequest defines how the authorization response impacts the client request and response.
	AuthorizationResponse *AuthorizationResponse `json:"authorizationResponse,omitempty" yaml:"authorizationResponse"`
}

// TLSConfig describes a TLS configuration.
type TLSConfig struct {
	// CertificateRefs contains a series of references to Kubernetes objects that
	// contains TLS certificates and private keys. These certificates are used to
	// establish a TLS handshake with the external authorization server.
	//
	// +kubebuilder:validation:MaxItems=64
	CertificateRefs []gwapiv1a2.SecretObjectReference `json:"certificateRefs,omitempty"`
}

// AuthorizationResponse defines how the authorization response impacts the client request and response.
type AuthorizationResponse struct {
	// Authorization response headers that have a correspondent match will be added to the original client request.
	// Note that coexistent headers will be overridden.
	AllowedUpstreamHeaders []string `json:"allowedUpstreamHeaders,omitempty"`

	// Authorization response headers that have a correspondent match will be added to the original client request.
	// Note that coexistent headers will be appended.
	AllowedUpstreamHeadersToAppend []string `json:"allowedUpstreamMethods,omitempty"`
}
