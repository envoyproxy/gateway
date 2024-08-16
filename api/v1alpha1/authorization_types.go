// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

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
	// If not specified, Envoy Gateway will generate a unique name for the rule.n
	// +optional
	Name *string `json:"name,omitempty"`

	// Action defines the action to be taken if the rule matches.
	Action AuthorizationAction `json:"action"`

	// Principal specifies the client identity of a request.
	// If there are multiple principal types, all principals must match for the rule to match.
	// For example, if there are two principals: one for client IP and one for JWT claim,
	// the rule will match only if both the client IP and the JWT claim match.
	Principal Principal `json:"principal"`
}

// Principal specifies the client identity of a request.
// A client identity can be a client IP, a JWT claim, username from the Authorization header,
// or any other identity that can be extracted from a custom header.
// Currently, only the client IP is supported.
type Principal struct {
	// ClientCIDRs are the IP CIDR ranges of the client.
	// Valid examples are "192.168.1.0/24" or "2001:db8::/64"
	//
	// If multiple CIDR ranges are specified, one of the CIDR ranges must match
	// the client IP for the rule to match.
	//
	// The client IP is inferred from the X-Forwarded-For header, a custom header,
	// or the proxy protocol.
	// You can use the `ClientIPDetection` or the `EnableProxyProtocol` field in
	// the `ClientTrafficPolicy` to configure how the client IP is detected.
	// +kubebuilder:validation:MinItems=1
	ClientCIDRs []CIDR `json:"clientCIDRs"`

	// JWTClaims are the claims in a JWT token.
	//
	// If multiple claims are specified, all claims must match for the rule to match.
	// For example, if there are two claims: one for the audience and one for the issuer,
	// the rule will match only if both the audience and the issuer match.
	// Note: in order to use JWT claims for authorization, you must configure the
	// JWT authentication in the same `SecurityPolicy`.
	// +optional
	// +notImplementedHide
	JWTClaims []JWTClaim `json:"jwtClaims"`
}

// JWTClaim specifies a claim in a JWT token.
type JWTClaim struct {
	// Type is the type of the claim.
	// Valid values are "String" and "StringArray".
	// For example, sub is a string claim, and groups is a string array claim.
	// +kubebuilder:validation:Enum=String;StringArray
	// +kubebuilder:default=String
	// +unionDiscriminator
	Type JWTClaimType `json:"type"`

	// Name is the name of the claim.
	// If it is a nested claim, use a dot (.) separated string as the name to
	// represent the full path to the claim.
	// For example, if the claim is in the "department" field in the "organization" field,
	// the name should be "organization.department".
	Name string `json:"name"`

	// Values are the values that the claim must match.
	// If the claim is a string type, the specified value must match exactly.
	// If the claim is a string array type, the specified value must match one of the values in the array.
	// If multiple values are specified, one of the values must match for the rule to match.
	Values []string `json:"values"`
}

type JWTClaimType string

const (
	JWTClaimTypeString      JWTClaimType = "String"
	JWTClaimTypeStringArray JWTClaimType = "StringArray"
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
