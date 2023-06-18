// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

const (
	// KindAuthenticationFilter is the name of the AuthenticationFilter kind.
	KindAuthenticationFilter = "AuthenticationFilter"
)

//+kubebuilder:object:root=true

type AuthenticationFilter struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of the AuthenticationFilter type.
	Spec AuthenticationFilterSpec `json:"spec"`

	// Note: The status sub-resource has been excluded but may be added in the future.
}

// ClaimToHeader defines a configuration to convert JWT claims into HTTP headers
type ClaimToHeader struct {

	// Header defines the name of the HTTP request header that the JWT Claim will be saved into.
	Header string `json:"header"`

	// Claim is the JWT Claim that should be saved into the header : it can be a nested claim of type
	// (eg. "claim.nested.key", "sub"). The nested claim name must use dot "."
	// to separate the JSON name path.
	Claim string `json:"claim"`
}

// AuthenticationFilterSpec defines the desired state of the AuthenticationFilter type.
// +union
type AuthenticationFilterSpec struct {
	// Type defines the type of authentication provider to use. Supported provider types
	// are "JWT".
	//
	// +unionDiscriminator
	Type AuthenticationFilterType `json:"type"`

	// JWT defines the JSON Web Token (JWT) authentication provider type. When multiple
	// jwtProviders are specified, the JWT is considered valid if any of the providers
	// successfully validate the JWT. For additional details, see
	// https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/jwt_authn_filter.html.
	//
	// +kubebuilder:validation:MaxItems=4
	// +optional
	JwtProviders []JwtAuthenticationFilterProvider `json:"jwtProviders,omitempty"`
}

// AuthenticationFilterType is a type of authentication provider.
// +kubebuilder:validation:Enum=JWT
type AuthenticationFilterType string

const (
	// JwtAuthenticationFilterProviderType is a provider that uses JSON Web Token (JWT)
	// for authenticating requests..
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

	// Issuer is the principal that issued the JWT and takes the form of a URL or email address.
	// For additional details, see https://tools.ietf.org/html/rfc7519#section-4.1.1 for
	// URL format and https://rfc-editor.org/rfc/rfc5322.html for email format. If not provided,
	// the JWT issuer is not checked.
	//
	// +kubebuilder:validation:MaxLength=253
	// +optional
	Issuer string `json:"issuer,omitempty"`

	// Audiences is a list of JWT audiences allowed access. For additional details, see
	// https://tools.ietf.org/html/rfc7519#section-4.1.3. If not provided, JWT audiences
	// are not checked.
	//
	// +kubebuilder:validation:MaxItems=8
	// +optional
	Audiences []string `json:"audiences,omitempty"`

	// RemoteJWKS defines how to fetch and cache JSON Web Key Sets (JWKS) from a remote
	// HTTP/HTTPS endpoint.
	RemoteJWKS RemoteJWKS `json:"remoteJWKS"`

	// ClaimToHeaders is a list of JWT claims that must be extracted into HTTP request headers
	// For examples, following config:
	// The claim must be of type; string, int, double, bool. Array type claims are not supported
	//
	ClaimToHeaders []ClaimToHeader `json:"claimToHeaders,omitempty"`
	// TODO: Add TBD JWT fields based on defined use cases.
}

// RemoteJWKS defines how to fetch and cache JSON Web Key Sets (JWKS) from a remote
// HTTP/HTTPS endpoint.
type RemoteJWKS struct {
	// URI is the HTTPS URI to fetch the JWKS. Envoy's system trust bundle is used to
	// validate the server certificate.
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	URI string `json:"uri"`

	// TODO: Add TBD remote JWKS fields based on defined use cases.
}

// OIDCAuthenticationFilterProvider defines the OpenID Connect (OIDC) authentication provider type
type OIDCAuthenticationFilterProvider struct {
	// The OIDC Provider configuration.
	Provider OIDCProvider `json:"provider"`

	// The OIDC client ID assigned to the filter to be used in the
	// [Authentication Request](https://openid.net/specs/openid-connect-core-1_0.html#AuthRequest).
	//
	// +kubebuilder:validation:MinLength=1
	ClientID string `json:"clientId"`

	// The Kubernetes secret which contains the OIDC client secret assigned to the filter to be used in the
	// [Authentication Request](https://openid.net/specs/openid-connect-core-1_0.html#AuthRequest).
	//
	// +kubebuilder:validation:Required
	ClientSecret gwapiv1b1.SecretObjectReference `json:"clientSecret"`

	// The redirect URI passed to the authorization endpoint in the
	// [Authentication Request](https://openid.net/specs/openid-connect-core-1_0.html#AuthRequest).
	//
	// +kubebuilder:validation:MinLength=1
	RedirectURI string `json:"redirectUri"`
}

type OIDCProvider struct {
	// The OIDC Provider's [issuer identifier](https://openid.net/specs/openid-connect-discovery-1_0.html#IssuerDiscovery).
	//
	// +kubebuilder:validation:MinLength=1
	Issuer string `json:"issuer"`

	// The OIDC Provider's [authorization endpoint](https://openid.net/specs/openid-connect-core-1_0.html#AuthorizationEndpoint).
	// If not provided, EG will try to discover it from the provider's [Well-Known Configuration Endpoint](https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderConfigurationResponse).
	//
	// +optional
	AuthorizationEndpoint string `json:"authorizationEndpoint,omitempty"`

	// The OIDC Provider's [token endpoint](https://openid.net/specs/openid-connect-core-1_0.html#TokenEndpoint).
	// If not provided, EG will try to discover it from the provider's [Well-Known Configuration Endpoint](https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderConfigurationResponse).
	//
	// +optional
	TokenEndpoint string `json:"tokenEndpoint,omitempty"`

	// The JSON JWKS response from the OIDC provider’s `jwks_uri` URI which can be found in the OIDC provider's
	// [configuration response](https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderConfigurationResponse).
	// Note that this JSON value must be escaped when embedded in a json configmap
	// (see [example](https://github.com/istio-ecosystem/authservice/blob/master/bookinfo-example/config/authservice-configmap-template.yaml)).
	// Used during token verification.
	// If not provided, EG will try to discover it from the provider's [Well-Known Configuration Endpoint](https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderConfigurationResponse).
	//
	// +optional
	Jwks string `json:"jwks,omitempty"`
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
