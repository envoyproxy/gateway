// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
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
	Host gwapiv1a2.Hostname `json:"host"`

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
	Host gwapiv1a2.Hostname `json:"host"`

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
	// CertificateRef is the reference to a Kubernetes Secret that contains the
	// TLS certificate and private key. The certificate and private key will be
	// used to establish a TLS handshake between the Envoy proxy and the external
	// authorization server.
	// The referenced Secret must contain two keys: tls.crt and tls.key.
	//
	// If this field is not specified, the proxy will not present a client certificate
	// to the external authorization server.
	//
	// +optional
	CertificateRef *gwapiv1a2.SecretObjectReference `json:"certificateRef,omitempty"`

	// CACertRef is the reference to a Kubernetes ConfigMap that contains a
	// PEM-encoded TLS CA certificate bundle, which is used to validate the
	// certificate presented by the external authorization server.
	// The referenced ConfigMap must contain a key named ca.crt.
	//
	// If not specified, the proxy will use the system default certificate pool to
	// verify the server certificate.
	//
	// +optional
	CACertRef *gwapiv1.LocalObjectReference `json:"caCertRef,omitempty"`

	// Hostname is used for two purposes in the connection between Envoy and the
	// external authorization server:
	//
	// 1. Hostname MUST be used as the SNI to connect to the external authorization server (RFC 6066).
	// 2. Hostname MUST be used for authentication and MUST match the certificate
	//    served by the external authorization server.
	Hostname v1beta1.PreciseHostname `json:"hostname"`
}
