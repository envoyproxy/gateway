// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

// TCPAuthorization defines the authorization configuration for TCP traffic.
//
// Note: if neither `Rules` nor `DefaultAction` is specified, the default action is to deny all connections.
type TCPAuthorization struct {
	// Rules defines a list of authorization rules.
	// These rules are evaluated in order, the first matching rule will be applied,
	// and the rest will be skipped.
	//
	// +optional
	Rules []TCPAuthorizationRule `json:"rules,omitempty"`

	// DefaultAction defines the default action to be taken if no rules match.
	// If not specified, the default action is Deny.
	//
	// +optional
	DefaultAction *AuthorizationAction `json:"defaultAction,omitempty"`
}

// TCPAuthorizationRule defines a single authorization rule for TCP traffic.
type TCPAuthorizationRule struct {
	// Name is a user-friendly name for the rule.
	// If not specified, Envoy Gateway will generate a unique name for the rule.
	//
	// +optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	Name *string `json:"name,omitempty"`

	// Action defines the action to be taken if the rule matches.
	Action AuthorizationAction `json:"action"`

	// Principal specifies the client identity of a connection.
	Principal TCPPrincipal `json:"principal"`
}

// TCPPrincipal specifies the client identity of a TCP connection.
// If there are multiple principal types, all principals must match for the rule to match.
//
// +kubebuilder:validation:XValidation:rule="has(self.clientCIDRs)",message="clientCIDRs must be specified"
type TCPPrincipal struct {
	// ClientCIDRs are the IP CIDR ranges of the client.
	// Valid examples are "192.168.1.0/24" or "2001:db8::/64"
	//
	// If multiple CIDR ranges are specified, one of the CIDR ranges must match
	// the client IP for the rule to match.
	//
	// The client IP is the peer address of the TCP connection, or the address
	// carried by the PROXY protocol when it is enabled on the listener with
	// `ClientTrafficPolicy`.
	//
	// +optional
	// +kubebuilder:validation:MinItems=1
	ClientCIDRs []CIDR `json:"clientCIDRs,omitempty"`
}

// UDPAuthorization defines the authorization configuration for UDP traffic.
//
// Note: if neither `Rules` nor `DefaultAction` is specified, the default action is to deny all datagrams.
type UDPAuthorization struct {
	// Rules defines a list of authorization rules.
	// These rules are evaluated in order, the first matching rule will be applied,
	// and the rest will be skipped.
	//
	// +optional
	Rules []UDPAuthorizationRule `json:"rules,omitempty"`

	// DefaultAction defines the default action to be taken if no rules match.
	// If not specified, the default action is Deny.
	//
	// +optional
	DefaultAction *AuthorizationAction `json:"defaultAction,omitempty"`
}

// UDPAuthorizationRule defines a single authorization rule for UDP traffic.
type UDPAuthorizationRule struct {
	// Name is a user-friendly name for the rule.
	// If not specified, Envoy Gateway will generate a unique name for the rule.
	//
	// +optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	Name *string `json:"name,omitempty"`

	// Action defines the action to be taken if the rule matches.
	Action AuthorizationAction `json:"action"`

	// Principal specifies the client identity of a datagram.
	Principal UDPPrincipal `json:"principal"`
}

// UDPPrincipal specifies the client identity of a UDP datagram.
// If there are multiple principal types, all principals must match for the rule to match.
//
// +kubebuilder:validation:XValidation:rule="has(self.clientCIDRs)",message="clientCIDRs must be specified"
type UDPPrincipal struct {
	// ClientCIDRs are the IP CIDR ranges of the client.
	// Valid examples are "192.168.1.0/24" or "2001:db8::/64"
	//
	// If multiple CIDR ranges are specified, one of the CIDR ranges must match
	// the client IP for the rule to match.
	//
	// The client IP is the source address of the datagram. The PROXY protocol
	// is not available for UDP, so any load balancer or NAT in front of Envoy
	// must preserve the source address for this to match the original client.
	//
	// +optional
	// +kubebuilder:validation:MinItems=1
	ClientCIDRs []CIDR `json:"clientCIDRs,omitempty"`
}
