// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

const (
	// KindSecurityPolicy is the name of the SecurityPolicy kind.
	KindSecurityPolicy = "SecurityPolicy"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.conditions[?(@.type=="Accepted")].reason`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// SecurityPolicy allows the user to configure various security settings for a
// Gateway.
type SecurityPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of SecurityPolicy.
	Spec SecurityPolicySpec `json:"spec"`

	// Status defines the current status of SecurityPolicy.
	Status SecurityPolicyStatus `json:"status,omitempty"`
}

// SecurityPolicySpec defines the desired state of SecurityPolicy.
type SecurityPolicySpec struct {
	// TargetRef is the name of the Gateway resource this policy
	// is being attached to.
	// This Policy and the TargetRef MUST be in the same namespace
	// for this Policy to have effect and be applied to the Gateway.
	// TargetRef
	TargetRef gwapiv1a2.PolicyTargetReferenceWithSectionName `json:"targetRef"`

	// CORS defines the configuration for Cross-Origin Resource Sharing (CORS).
	//
	// +optional
	CORS *CORS `json:"cors,omitempty"`

	// JWTAuthentication defines the configuration for JSON Web Token (JWT)
	// authentication.
	//
	// +optional
	JWTAuthentication *JWTAuthentication `json:"jwtAuthentication,omitempty"`
}

// CORS defines the configuration for Cross-Origin Resource Sharing (CORS).
type CORS struct {
	// AllowOrigins defines the origins that are allowed to make requests.
	// +kubebuilder:validation:MinItems=1
	AllowOrigins []StringMatch `json:"allowOrigins,omitempty" yaml:"allowOrigins"`
	// AllowMethods defines the methods that are allowed to make requests.
	// +kubebuilder:validation:MinItems=1
	AllowMethods []string `json:"allowMethods,omitempty" yaml:"allowMethods"`
	// AllowHeaders defines the headers that are allowed to be sent with requests.
	AllowHeaders []string `json:"allowHeaders,omitempty" yaml:"allowHeaders,omitempty"`
	// ExposeHeaders defines the headers that can be exposed in the responses.
	ExposeHeaders []string `json:"exposeHeaders,omitempty" yaml:"exposeHeaders,omitempty"`
	// MaxAge defines how long the results of a preflight request can be cached.
	MaxAge *metav1.Duration `json:"maxAge,omitempty" yaml:"maxAge,omitempty"`
}

// JWTAuthentication defines the configuration for JSON Web Token (JWT) authentication.
type JWTAuthentication struct {

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

// StringMatch defines how to match any strings.
// This is a general purpose match condition that can be used by other EG APIs
// that need to match against a string.
type StringMatch struct {
	// Type specifies how to match against a string.
	//
	// +optional
	// +kubebuilder:default=Exact
	Type *MatchType `json:"type,omitempty"`

	// Value specifies the string value that the match must have.
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=1024
	Value string `json:"value"`
}

// MatchType specifies the semantics of how a string value should be compared.
// Valid MatchType values are "Exact", "Prefix", "Suffix", "RegularExpression".
//
// +kubebuilder:validation:Enum=Exact;Prefix;Suffix;RegularExpression
type MatchType string

const (
	// MatchExact :the input string must match exactly the match value.
	MatchExact MatchType = "Exact"

	// MatchPrefix :the input string must start with the match value.
	MatchPrefix MatchType = "Prefix"

	// MatchSuffix :the input string must end with the match value.
	MatchSuffix MatchType = "Suffix"

	// MatchRegularExpression :The input string must match the regular expression
	// specified in the match value.
	// The regex string must adhere to the syntax documented in
	// https://github.com/google/re2/wiki/Syntax.
	MatchRegularExpression MatchType = "RegularExpression"
)

// SecurityPolicyStatus defines the state of SecurityPolicy
type SecurityPolicyStatus struct {
	// Conditions describe the current conditions of the SecurityPolicy.
	//
	// +optional
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MaxItems=8
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true

// SecurityPolicyList contains a list of SecurityPolicy resources.
type SecurityPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SecurityPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SecurityPolicy{}, &SecurityPolicyList{})
}
