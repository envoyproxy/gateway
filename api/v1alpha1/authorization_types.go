// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

// Authorization defines the authorization configuration.
// +notImplementedHide
type Authorization struct {
	// Rules defines a list of authorization rules.
	// These rules are evaluated in order, the first matching rule will be applied,
	// and the rest will be skipped.
	//
	// For example, if there are two rules: the first rule allows the request
	// and the second rule denies it, when a request matches both rules, it will be allowed.
	//
	// +optional
	Rules []Rule `json:"rules,omitempty"`

	// DefaultAction defines the default action to be taken if no rules match.
	// If not specified, the default action is Deny.
	// +optional
	DefaultAction *RuleActionType `json:"defaultAction"`
}

// Rule defines the single authorization rule.
// +notImplementedHide
type Rule struct {
	// Action defines the action to be taken if the rule matches.
	Action RuleActionType `json:"action"`

	// Principal specifies the client identity of a request.
	Principal Principal `json:"principal"`

	// Permissions contains allowed HTTP methods.
	// If empty, all methods are matching.
	//
	// +optional
	// Permissions []string `json:"permissions,omitempty"`
}

// Principal specifies the client identity of a request.
// +notImplementedHide
type Principal struct {
	// ClientCIDR is the IP CIDR range of the client.
	// Valid examples are "192.168.1.0/24" or "2001:db8::/64"
	//
	// By default, the client IP is inferred from the x-forwarder-for header and proxy protocol.
	// You can use the `EnableProxyProtocol` and `ClientIPDetection` options in
	// the `ClientTrafficPolicy` to configure how the client IP is detected.
	ClientCIDR []string `json:"clientCIDR,omitempty"`
}

// RuleActionType specifies the types of authorization rule action.
// +kubebuilder:validation:Enum=Allow;Deny
// +notImplementedHide
type RuleActionType string

const (
	// Allow is the action to allow the request.
	Allow RuleActionType = "Allow"
	// Deny is the action to deny the request.
	Deny RuleActionType = "Deny"
)
