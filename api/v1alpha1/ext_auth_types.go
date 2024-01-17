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

	// GRPC defines the gRPC External Authorization service
	// Only one of GRPCService or HTTPService may be specified.
	GRPC *GRPCExtAuthService `json:"grpc,omitempty"`

	// HTTP defines the HTTP External Authorization service
	// Only one of GRPCService or HTTPService may be specified.
	HTTP *HTTPExtAuthService `json:"http,omitempty"`

	// AllowedHeaders defines the client request headers that will be included
	// in the request to the external authorization service.
	// Note: If not specified, the default behavior of different external authorization
	// services is different. All headers will be included in the check request
	// to a gRPC authorization server, whereas no headers will be included in the
	// check request to an HTTP authorization server.
	// +optional
	AllowedHeaders []string `json:"allowedHeaders,omitempty"`
}

// GRPCExtAuthService defines the gRPC External Authorization service
type GRPCExtAuthService struct {
	// Host is the hostname of the gRPC External Authorization service
	Host gwapiv1a2.Hostname `json:"host"`

	// Port is the network port of the gRPC External Authorization service
	Port gwapiv1a2.PortNumber `json:"port"`

	// TLS defines the TLS configuration for the gRPC External Authorization service.
	// Note: If not specified, the proxy will talk to the gRPC External
	// Authorization service in plaintext.
	// +optional
	TLS *TLSConfig `json:"tls,omitempty"`
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
	// +kubebuilder:validation:Pattern=`^(https?:)(\/\/[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*)((:\d+)?)(\/[a-zA-Z0-9-._~%]+\/?)*$`
	URL string `json:"url" yaml:"url"`

	// TLS defines the TLS configuration for the HTTP External Authorization service.
	// TLS is only valid when the URL scheme is https. If the URL scheme is
	// https, and TLS is not specified, the proxy will use the system default
	// certificate pool to verify the server certificate.
	// +optional
	TLS *TLSConfig `json:"tls,omitempty"`

	// Authorization response headers that will be added to the original client
	// request before sending it to the backend server.
	// Note that coexisting headers will be overridden.
	// If not specified, no authorization response headers will be added to the
	// original client request.
	// +optional
	AllowedBackendHeaders []string `json:"allowedBackendHeaders,omitempty"`
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
