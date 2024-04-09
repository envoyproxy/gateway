// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

// AuthorizationAction defines the action to take if a request matches the rule.
type AuthorizationAction string

const (
	// AuthorizationActionAllow allows the request.
	AuthorizationActionAllow AuthorizationAction = "ALLOW"
	// AuthorizationActionDeny denies the request.
	AuthorizationActionDeny AuthorizationAction = "DENY"
)

// Authorization defines the configuration for request authorization.
type Authorization struct {
	// Rules define a list of authorization rules.
	// +kubebuilder:validation:MinItems=1
	Rules []AuthorizationRule `json:"policies"`
}

// AuthorizationRule defines the configuration for an authorization rule.
type AuthorizationRule struct {
	// Action defines the action to take if a request matches the rule.
	// ALLOW allows the request, and DENY denies the request.
	// +kubebuilder:validation:Enum=ALLOW;DENY
	Action AuthorizationAction `json:"action"`

	// Policies define which principals are allowed/denied access to specified permissions based on the action.
	Policies []AuthorizationPolicy `json:"policy"`
}

// AuthorizationPolicy defines the configuration for an authorization policy.
type AuthorizationPolicy struct {
	// Principals define the list of principals that are allowed/denied access based on the action.
	// If no principals are specified, all requests are matched.
	Principles []Principle `json:"principles"`

	// Permissions define the list of permissions that are allowed/denied access based on the action.
	// If no permissions are specified, all requests are matched.
	Permissions []Permission `json:"permissions"`
}

// Principle defines the configuration for a principle.
type Principle struct {
	// SourceIP is an IP CIDR that represents the range of the source IP addresses for clients.
	// This address will honor proxy protocol, but will not honor XFF.
	// For example, `192.168.0.1/32`, `192.168.0.0/24`, `001:db8::/64`.
	// One of SourceIP or JWTClaim must be specified.
	SourceIP *string `json:"sourceIP,omitempty"`

	// JWTClaims is a JWT claim that can be used to authorize requests.
	// To use this claim, the JWT settings must be configured in the SecurityPolicy,
	// and the specified claim must be present in the JWT token.
	// One of SourceIP or JWTClaim must be specified.
	JWTClaim *JWTClaim
}

// JWTClaim defines a JWT claim that can be used to authorize requests.
type JWTClaim struct {
	// Name is the name of the claim.
	Name string `json:"name"`

	// Value is the value of the claim.
	Value string `json:"value"`
}

// Permission defines actions that the authorized principals can take.
type Permission struct {
	// Methods define the list of HTTP methods that are allowed/denied access based on the action.
	// If no methods are specified, all methods are matched.
	Method []string `json:"method"`
}
