// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

// JWT defines the configuration for JSON Web Token (JWT) authentication.
type JWT struct {
	// Optional determines whether a missing JWT is acceptable, defaulting to false if not specified.
	// Note: Even if optional is set to true, JWT authentication will still fail if an invalid JWT is presented.
	Optional *bool `json:"optional,omitempty"`

	// Providers defines the JSON Web Token (JWT) authentication provider type.
	// When multiple JWT providers are specified, the JWT is considered valid if
	// any of the providers successfully validate the JWT. For additional details,
	// see https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/jwt_authn_filter.html.
	//
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=4
	Providers []JWTProvider `json:"providers"`
}

// JWTProvider defines how a JSON Web Token (JWT) can be verified.
// +kubebuilder:validation:XValidation:rule="(has(self.recomputeRoute) && self.recomputeRoute) ? size(self.claimToHeaders) > 0 : true", message="claimToHeaders must be specified if recomputeRoute is enabled"
// +kubebuilder:validation:XValidation:rule="((has(self.remoteJWKS) && self.remoteJWKS) && (has(self.jwksSource) && self.jwksSource)", message="only one of jwksSource or remoteJWKS, should be specfied. remoteJWKS can be replaced by jwksSource."
// +kubebuilder:validation:XValidation:rule="(!has(self.remoteJWKS) && !has(self.jwksSource))", message="a jwksSource is required."
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
	// HTTP/HTTPS endpoint. This field is deprecated, https retrieval is supported by the uri
	// type in JWKSSource
	//
	// +optional
	RemoteJWKS *RemoteJWKS `json:"remoteJWKS"`

	// JWKSSource defines how to fetch and cache JSON Web Key Sets (JWKS) from a given source.
	// Inline, URI and Configmap/Secret contents are supported.
	//
	// +optional
	JWKSSource *JWKSSource `json:"jwksSource"`

	// ClaimToHeaders is a list of JWT claims that must be extracted into HTTP request headers
	// For examples, following config:
	// The claim must be of type; string, int, double, bool. Array type claims are not supported
	//
	// +optional
	ClaimToHeaders []ClaimToHeader `json:"claimToHeaders,omitempty"`

	// RecomputeRoute clears the route cache and recalculates the routing decision.
	// This field must be enabled if the headers generated from the claim are used for
	// route matching decisions. If the recomputation selects a new route, features targeting
	// the new matched route will be applied.
	//
	// +optional
	RecomputeRoute *bool `json:"recomputeRoute,omitempty"`

	// ExtractFrom defines different ways to extract the JWT token from HTTP request.
	// If empty, it defaults to extract JWT token from the Authorization HTTP request header using Bearer schema
	// or access_token from query parameters.
	//
	// +optional
	ExtractFrom *JWTExtractor `json:"extractFrom,omitempty"`
}

// RemoteJWKS defines how to fetch and cache JSON Web Key Sets (JWKS) from a remote
// HTTP/HTTPS endpoint. Note this has been replaced by JWKSSource and is deprecated.
type RemoteJWKS struct {
	// URI is the HTTPS URI to fetch the JWKS. Envoy's system trust bundle is used to
	// validate the server certificate.
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	URI string `json:"uri"`

	// TODO: Add TBD remote JWKS fields based on defined use cases.
}

// JWKSSourceType defines the types of values for the source of JSON Web Key Sets (JWKS) for use with a JWTProvider.
// +kubebuilder:validation:Enum=Inline;URI;ValueRef
type JWKSSourceType string

const (
	// JWKSSourceTypeInline defines the "Inline" type, indicating that the JWKS json is provided inline
	JWKSSourceTypeInline JWKSSourceType = "Inline"

	// JWKSSourceTypeURI defines the "URI" type, indicating that the JWKS json should be retrieved from the given uri
	// typically a https endpoint or a local file.
	JWKSSourceTypeURI JWKSSourceType = "URI"

	// JWKSSourceTypeValueRefdefines the "ValueRef" type, indicating that the JWKS json should be loaded from
	// the given LocalObjectReference
	JWKSSourceTypeValueRef JWKSSourceType = "ValueRef"
)

// JWKSSource defines how to load a JSON Web Key Sets (JWKS) from various source locations.
//
// +kubebuilder:validation:XValidation:message="uri must be set for type URI",rule="(!has(self.type) || self.type == 'URI')? has(self.path) : true"
// +kubebuilder:validation:XValidation:message="inline must be set for type Inline",rule="(!has(self.type) || self.type == 'Inline')? has(self.inline) : true"
// +kubebuilder:validation:XValidation:message="valueRef must be set for type ValueRef",rule="(has(self.type) && self.type == 'ValueRef')? has(self.valueRef) : true"
type JWKSSource struct {

	// Type is the type of method to use to read the contents of the JWKS.
	// Valid values are Inline, URI and ValueRef, default is URI.
	//
	// +kubebuilder:default=Inline
	// +kubebuilder:validation:Enum=Inline;URI;ValueRef
	// +unionDiscriminator
	Type *JWKSSourceType `json:"type"`

	// URI is a location of the JWKS for File or HTTPS based retrieval. The location of the contents
	// are determined by examining the scheme of the given URI. This covers the HTTPS usecase supported
	// by RemoteJWKS
	//
	// Envoy's system trust bundle is used to validate the server certificate for HTTPS retrieval
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	// +optional
	URI *string `json:"uri,omitempty"`

	// Inline contains the value as an inline string.
	//
	// +optional
	Inline *string `json:"inline,omitempty"`

	// ValueRef contains the contents of the JWKS specified as a local object reference specifically
	// a ConfigMap or Secret.
	//
	// The first value in the Secret/ConfigMap will be used as the contents for the JWKS
	//
	// +optional
	ValueRef *gwapiv1.LocalObjectReference `json:"valueRef,omitempty"`
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

// JWTExtractor defines a custom JWT token extraction from HTTP request.
// If specified, Envoy will extract the JWT token from the listed extractors (headers, cookies, or params) and validate each of them.
// If any value extracted is found to be an invalid JWT, a 401 error will be returned.
type JWTExtractor struct {
	// Headers represents a list of HTTP request headers to extract the JWT token from.
	//
	// +optional
	Headers []JWTHeaderExtractor `json:"headers,omitempty"`

	// Cookies represents a list of cookie names to extract the JWT token from.
	//
	// +optional
	Cookies []string `json:"cookies,omitempty"`

	// Params represents a list of query parameters to extract the JWT token from.
	//
	// +optional
	Params []string `json:"params,omitempty"`
}

// JWTHeaderExtractor defines an HTTP header location to extract JWT token
type JWTHeaderExtractor struct {
	// Name is the HTTP header name to retrieve the token
	//
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// ValuePrefix is the prefix that should be stripped before extracting the token.
	// The format would be used by Envoy like "{ValuePrefix}<TOKEN>".
	// For example, "Authorization: Bearer <TOKEN>", then the ValuePrefix="Bearer " with a space at the end.
	//
	// +optional
	ValuePrefix *string `json:"valuePrefix,omitempty"`
}
