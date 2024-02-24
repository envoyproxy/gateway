// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

// Authorization defines the authorization configuration.
type Authorization struct {
	// Rules contains all the authorization rules.
	// If rules contains at least one Allow rule and none of them
	// matches the action for the request is deny.
	// If rules contains at least one Deny rule and none of them
	// matches the action for the request is allow.
	//
	// +kubebuilder:validation:MinItems=1
	Rules []Rule `json:"rules,omitempty"`
}

// Rule defines the single authorization rule.
type Rule struct {
	// ClientSelectors contains the client selector configuration.
	// All selectors are and together and only if all selector are valid
	// the Action is performed.
	//
	// +kubebuilder:validation:MinItems=1
	ClientSelectors []ClientSelector `json:"clientSelector,omitempty"`

	// Action defines the action to be taken if the rule matches.
	Action RuleActionType `json:"action"`
}

// ClientSelector contains the client selector configuration.
type ClientSelector struct {
	// ClientCIDRs is a list of CIDRs.
	// Valid examples are "192.168.1.0/24" or "2001:db8::/64"
	//
	// +optional
	ClientCIDRs []string `json:"clientCIDR,omitempty"`
}

// RuleActionType specifies the types of authorization rule action.
// +kubebuilder:validation:Enum=Allow;Deny;Log
type RuleActionType string

const (
	// Allow is the action to allow the request.
	Allow RuleActionType = "Allow"
	// Deny is the action to deny the request.
	Deny RuleActionType = "Deny"
	// Log is the action to log the request.
	Log RuleActionType = "Log"
)
