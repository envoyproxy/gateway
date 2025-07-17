// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// Authorization defines the authorization configuration.
//
// Note: if neither `Rules` nor `DefaultAction` is specified, the default action is to deny all requests.
type Authorization struct {
	// Rules defines a list of authorization rules.
	// These rules are evaluated in order, the first matching rule will be applied,
	// and the rest will be skipped.
	//
	// For example, if there are two rules: the first rule allows the request
	// and the second rule denies it, when a request matches both rules, it will be allowed.
	//
	// +optional
	Rules []AuthorizationRule `json:"rules,omitempty"`

	// DefaultAction defines the default action to be taken if no rules match.
	// If not specified, the default action is Deny.
	// +optional
	DefaultAction *AuthorizationAction `json:"defaultAction"`
}

// AuthorizationRule defines a single authorization rule.
type AuthorizationRule struct {
	// Name is a user-friendly name for the rule.
	// If not specified, Envoy Gateway will generate a unique name for the rule.
	//
	// +optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	Name *string `json:"name,omitempty"`

	// Action defines the action to be taken if the rule matches.
	Action AuthorizationAction `json:"action"`

	// Operation specifies the operation of a request, such as HTTP methods.
	// If not specified, all operations are matched on.
	//
	// +optional
	Operation *Operation `json:"operation,omitempty"`

	// Principal specifies the client identity of a request.
	// If there are multiple principal types, all principals must match for the rule to match.
	// For example, if there are two principals: one for client IP and one for JWT claim,
	// the rule will match only if both the client IP and the JWT claim match.
	Principal Principal `json:"principal"`
}

// Operation specifies the operation of a request.
type Operation struct {
	// Methods are the HTTP methods of the request.
	// If multiple methods are specified, all specified methods are allowed or denied, based on the action of the rule.
	//
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=16
	Methods []gwapiv1.HTTPMethod `json:"methods"`

	// Other fields may be supported in the future, such as path or host.
}

// Principal specifies the client identity of a request.
// A client identity can be a client IP, a JWT claim, username from the Authorization header,
// or any other identity that can be extracted from a custom header.
// If there are multiple principal types, all principals must match for the rule to match.
//
// +kubebuilder:validation:XValidation:rule="(has(self.clientCIDRs) || has(self.jwt) || has(self.headers))",message="at least one of clientCIDRs, jwt, or headers must be specified"
type Principal struct {
	// ClientCIDRs are the IP CIDR ranges of the client.
	// Valid examples are "192.168.1.0/24" or "2001:db8::/64"
	//
	// If multiple CIDR ranges are specified, one of the CIDR ranges must match
	// the client IP for the rule to match.
	//
	// The client IP is inferred from the X-Forwarded-For header, a custom header,
	// or the proxy protocol.
	// You can use the `ClientIPDetection` or the `ProxyProtocol` field in
	// the `ClientTrafficPolicy` to configure how the client IP is detected.
	//
	// +optional
	// +kubebuilder:validation:MinItems=1
	ClientCIDRs []CIDR `json:"clientCIDRs,omitempty"`

	// JWT authorize the request based on the JWT claims and scopes.
	// Note: in order to use JWT claims for authorization, you must configure the
	// JWT authentication in the same `SecurityPolicy`.
	// +optional
	JWT *JWTPrincipal `json:"jwt,omitempty"`

	// Headers authorize the request based on user identity extracted from custom headers.
	// If multiple headers are specified, all headers must match for the rule to match.
	//
	// +optional
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=256
	Headers []AuthorizationHeaderMatch `json:"headers,omitempty"`
}

// AuthorizationHeaderMatch specifies how to match against the value of an HTTP header within a authorization rule.
type AuthorizationHeaderMatch struct {
	// Name of the HTTP header.
	// The header name is case-insensitive unless PreserveHeaderCase is set to true.
	// For example, "Foo" and "foo" are considered the same header.
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=256
	Name string `json:"name"`

	// Values are the values that the header must match.
	// If multiple values are specified, the rule will match if any of the values match.
	//
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=256
	Values []string `json:"values"`

	// Only exact matches are supported for now. It should be good enough for authorization use cases. If use cases for other
	// matching types arise, we can add a MatchingType field here.
}

// JWTPrincipal specifies the client identity of a request based on the JWT claims and scopes.
// At least one of the claims or scopes must be specified.
// Claims and scopes are And-ed together if both are specified.
//
// +kubebuilder:validation:XValidation:rule="(has(self.claims) || has(self.scopes))",message="at least one of claims or scopes must be specified"
type JWTPrincipal struct {
	// Provider is the name of the JWT provider that used to verify the JWT token.
	// In order to use JWT claims for authorization, you must configure the JWT
	// authentication with the same provider in the same `SecurityPolicy`.
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	Provider string `json:"provider"`

	// Claims are the claims in a JWT token.
	//
	// If multiple claims are specified, all claims must match for the rule to match.
	// For example, if there are two claims: one for the audience and one for the issuer,
	// the rule will match only if both the audience and the issuer match.
	//
	// +optional
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=16
	Claims []JWTClaim `json:"claims,omitempty"`

	// Scopes are a special type of claim in a JWT token that represents the permissions of the client.
	//
	// The value of the scopes field should be a space delimited string that is expected in the scope parameter,
	// as defined in RFC 6749: https://datatracker.ietf.org/doc/html/rfc6749#page-23.
	//
	// If multiple scopes are specified, all scopes must match for the rule to match.
	//
	// +optional
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=16
	Scopes []JWTScope `json:"scopes,omitempty"`
}

// JWTClaim specifies a claim in a JWT token.
type JWTClaim struct {
	// Name is the name of the claim.
	// If it is a nested claim, use a dot (.) separated string as the name to
	// represent the full path to the claim.
	// For example, if the claim is in the "department" field in the "organization" field,
	// the name should be "organization.department".
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	Name string `json:"name"`

	// ValueType is the type of the claim value.
	// Only String and StringArray types are supported for now.
	//
	// +kubebuilder:validation:Enum=String;StringArray
	// +kubebuilder:default=String
	// +unionDiscriminator
	// +optional
	ValueType *JWTClaimValueType `json:"valueType,omitempty"`

	// Values are the values that the claim must match.
	// If the claim is a string type, the specified value must match exactly.
	// If the claim is a string array type, the specified value must match one of the values in the array.
	// If multiple values are specified, one of the values must match for the rule to match.
	//
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=16
	Values []string `json:"values"`
}

// +kubebuilder:validation:MinLength=1
// +kubebuilder:validation:MaxLength=253
type JWTScope string

type JWTClaimValueType string

const (
	JWTClaimValueTypeString      JWTClaimValueType = "String"
	JWTClaimValueTypeStringArray JWTClaimValueType = "StringArray"
)

// AuthorizationAction defines the action to be taken if a rule matches.
// +kubebuilder:validation:Enum=Allow;Deny
type AuthorizationAction string

const (
	// AuthorizationActionAllow is the action to allow the request.
	AuthorizationActionAllow AuthorizationAction = "Allow"
	// AuthorizationActionDeny is the action to deny the request.
	AuthorizationActionDeny AuthorizationAction = "Deny"
)
