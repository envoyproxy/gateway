// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

// JWT defines the configuration for JSON Web Token (JWT) authentication.
type JWT struct {

	// Providers defines the JSON Web Token (JWT) authentication provider type.
	//
	// When multiple JWT providers are specified, the JWT is considered valid if
	// any of the providers successfully validate the JWT. For additional details,
	// see https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/jwt_authn_filter.html.
	//
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=4
	Providers []JWTProvider `json:"providers"`
}

// JWTProvider defines how a JSON Web Token (JWT) can be verified.
type JWTProvider struct {
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

// ClaimToHeader defines a configuration to convert JWT claims into HTTP headers
type ClaimToHeader struct {

	// Header defines the name of the HTTP request header that the JWT Claim will be saved into.
	Header string `json:"header"`

	// Claim is the JWT Claim that should be saved into the header : it can be a nested claim of type
	// (eg. "claim.nested.key", "sub"). The nested claim name must use dot "."
	// to separate the JSON name path.
	Claim string `json:"claim"`
}
