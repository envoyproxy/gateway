// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// AuthenticationFilterKind is the name of the AuthenticationFilter kind.
	AuthenticationFilterKind = "AuthenticationFilter"
)

//+kubebuilder:object:root=true

type AuthenticationFilter struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of the AuthenticationFilter type.
	Spec AuthenticationFilterSpec `json:"spec"`

	// Note: The status sub-resource has been excluded but may be added in the future.
}

// AuthenticationFilterSpec defines the desired state of the AuthenticationFilter type.
// +union
type AuthenticationFilterSpec struct {
	// Type defines the type of authentication provider to use. Supported provider types are:
	//
	//   * JWT: A provider that uses JSON Web Token (JWT) for authenticating requests.
	//
	// +unionDiscriminator
	Type AuthenticationFilterType `json:"type"`

	// JWT defines the JSON Web Token (JWT) authentication provider type. When multiple
	// jwtProviders are specified, the JWT is considered valid if any of the providers
	// successfully validate the JWT. For additional details, see:
	//
	//   https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/jwt_authn_filter.html
	//
	// +kubebuilder:validation:MaxItems=4
	// +optional
	JwtProviders []JwtAuthenticationFilterProvider `json:"jwtProviders,omitempty"`
}

// AuthenticationFilterType is a type of authentication provider.
// +kubebuilder:validation:Enum=JWT
type AuthenticationFilterType string

const (
	// JwtAuthenticationFilterProviderType is the JWT authentication provider type.
	JwtAuthenticationFilterProviderType AuthenticationFilterType = "JWT"
)

// JwtAuthenticationFilterProvider defines the JSON Web Token (JWT) authentication provider type
// and how JWTs should be verified:
type JwtAuthenticationFilterProvider struct {
	// Name defines a unique name for the JWT provider. A name can have a variety of forms,
	// including RFC1123 subdomains, RFC 1123 labels, or RFC 1035 labels.
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	Name string `json:"name"`

	// Issuer is the principal that issued the JWT.	For additional details, see:
	//
	//   https://tools.ietf.org/html/rfc7519#section-4.1.1
	//
	// Example:
	//  issuer: https://auth.example.com
	//
	// If not provided, the JWT issuer is not checked.
	//
	// +kubebuilder:validation:MaxLength=253
	// +optional
	Issuer string `json:"issuer,omitempty"`

	// Audiences is a list of JWT audiences allowed to access. For additional details, see:
	//
	//   https://tools.ietf.org/html/rfc7519#section-4.1.3
	//
	// Example:
	//   audiences:
	//   - foo.apps.example.com
	//     bar.apps.example.com
	//
	// If not provided, JWT audiences are not checked.
	//
	// +kubebuilder:validation:MaxItems=8
	// +optional
	Audiences []string `json:"audiences,omitempty"`

	// RemoteJWKS defines how to fetch and cache JSON Web Key Sets (JWKS) from a remote
	// HTTP/HTTPS endpoint.
	RemoteJWKS RemoteJWKS `json:"remoteJWKS"`

	// TODO: Add TBD JWT fields based on defined use cases.
}

// RemoteJWKS defines how to fetch and cache JSON Web Key Sets (JWKS) from a remote
// HTTP/HTTPS endpoint.
type RemoteJWKS struct {
	// URI is the HTTP/HTTPS URI to fetch the JWKS. When using an HTTPS endpoint,
	// Envoy's system trust bundle is used to validate the server certificate.
	//
	// Example:
	//  uri: https://www.foo.com/oauth2/v1/certs
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	URI string `json:"uri"`

	// TODO: Add TBD remote JWKS fields based on defined use cases.
}

//+kubebuilder:object:root=true

// AuthenticationFilterList contains a list of AuthenticationFilter.
type AuthenticationFilterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AuthenticationFilter `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AuthenticationFilter{}, &AuthenticationFilterList{})
}
