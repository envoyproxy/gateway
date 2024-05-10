// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

// Authorization defines the authorization configuration.
// +notImplementedHide
type Authorization struct {
	// Rules contains all the authorization rules.
	// Rules are evaluated in order, the first matching rule will be applied,
	// and the rest will be skipped.
	//
	// For example, if there are two rules, the first rule allows the request,
	// and the second rule denies the request, the request will be allowed.
	// If the first rule denies the request, and the second rule allows it,
	// the request will be denied.
	//
	// +optional
	Rules []Rule `json:"rules"`

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

	// Subject contains the subject of the rule.
	Subject Subject `json:"subjects,omitempty"`

	// Permissions contains allowed HTTP methods.
	// If empty, all methods are matching.
	//
	// +optional
	// Permissions []string `json:"permissions,omitempty"`
}

// Subject is the subject of the rule.
// +kubebuilder:validation:XValidation:rule="has(self.clientCIDR) || has(self.notClientCIDR)",message="subject must not be empty"
// +notImplementedHide
type Subject struct {
	// ClientCIDR is the IP CIDR range of the client.
	// Valid examples are "192.168.1.0/24" or "2001:db8::/64"
	ClientCIDR []string `json:"clientCIDR,omitempty"`

	// NotClientCIDR is the IP CIDR range of the client that should not match.
	// Valid examples are "192.168.1.0/24" or "2001:db8::/64"
	NotClientCIDR []string `json:"notClientCIDR,omitempty"`
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
