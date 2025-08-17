// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

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
// +kubebuilder:validation:XValidation:rule="(has(self.recomputeRoute) && self.recomputeRoute) ? size(self.claimToHeaders) > 0 : true", message="claimToHeaders must be specified if recomputeRoute is enabled."
// +kubebuilder:validation:XValidation:rule="has(self.remoteJWKS) || has(self.localJWKS)", message="either remoteJWKS or localJWKS must be specified."
// +kubebuilder:validation:XValidation:rule="!(has(self.remoteJWKS) && has(self.localJWKS))", message="remoteJWKS and localJWKS cannot both be specified."
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
	//
	// +optional
	RemoteJWKS *RemoteJWKS `json:"remoteJWKS,omitempty"`

	// LocalJWKS defines how to get the JSON Web Key Sets (JWKS) from a local source.
	//
	// +optional
	LocalJWKS *LocalJWKS `json:"localJWKS,omitempty"`

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

// RemoteJWKS defines how to fetch and cache JSON Web Key Sets (JWKS) from a remote HTTP/HTTPS endpoint.
// +kubebuilder:validation:XValidation:rule="!has(self.backendRef)",message="BackendRefs must be used, backendRef is not supported."
// +kubebuilder:validation:XValidation:rule="has(self.backendSettings)? (has(self.backendSettings.retry)?(has(self.backendSettings.retry.perRetry)? !has(self.backendSettings.retry.perRetry.timeout):true):true):true",message="Retry timeout is not supported."
// +kubebuilder:validation:XValidation:rule="has(self.backendSettings)? (has(self.backendSettings.retry)?(has(self.backendSettings.retry.retryOn)? !has(self.backendSettings.retry.retryOn.httpStatusCodes):true):true):true",message="HTTPStatusCodes is not supported."
type RemoteJWKS struct {
	// BackendRefs is used to specify the address of the Remote JWKS. The BackendRefs are optional, if not specified,
	// the backend service is extracted from the host and port of the URI field.
	//
	// TLS configuration can be specified in a BackendTLSConfig resource and target the BackendRefs.
	//
	// Other settings for the connection to remote JWKS can be specified in the BackendSettings resource.
	// Currently, only the retry policy is supported.
	//
	// +optional
	BackendCluster `json:",inline"`

	// URI is the HTTPS URI to fetch the JWKS. Envoy's system trust bundle is used to validate the server certificate.
	// If a custom trust bundle is needed, it can be specified in a BackendTLSConfig resource and target the BackendRefs.
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	URI string `json:"uri"`
	// Duration after which the cached JWKS should be expired. If not specified, default cache duration is 5 minutes.

	// +kubebuilder:default="300s"
	// +kubebuilder:validation:Format=duration
	// +optional
	CacheDuration *gwapiv1.Duration `json:"cacheDuration,omitempty"`
}

// LocalJWKSType defines the types of values for Local JWKS.
type LocalJWKSType string

const (
	// LocalJWKSTypeInline defines the "Inline" LocalJWKS type.
	LocalJWKSTypeInline LocalJWKSType = "Inline"

	// LocalJWKSTypeValueRef defines the "ValueRef" LocalJWKS type.
	LocalJWKSTypeValueRef LocalJWKSType = "ValueRef"
)

// LocalJWKS defines how to load a JSON Web Key Sets (JWKS) from a local source, either inline or from a reference to a ConfigMap.
//
// +kubebuilder:validation:XValidation:rule="(self.type == 'Inline' && has(self.inline) && !has(self.valueRef)) || (self.type == 'ValueRef' && !has(self.inline) && has(self.valueRef))",message="Exactly one of inline or valueRef must be set with correct type."
type LocalJWKS struct {
	// Type is the type of method to use to read the body value.
	// Valid values are Inline and ValueRef, default is Inline.
	//
	// +kubebuilder:default=Inline
	// +kubebuilder:validation:Enum=Inline;ValueRef
	// +unionDiscriminator
	Type *LocalJWKSType `json:"type"`

	// Inline contains the value as an inline string.
	//
	// +optional
	Inline *string `json:"inline,omitempty"`

	// ValueRef is a reference to a local ConfigMap that contains the JSON Web Key Sets (JWKS).
	//
	// The value of key `jwks` in the ConfigMap will be used.
	// If the key is not found, the first value in the ConfigMap will be used.
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
