// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

// Authorization defines the authorization configuration.
type Authorization struct {
	// Rules contains all the authorization rules.
	//
	// +kubebuilder:validation:MinItems=1
	Rules []Rule `json:"rules,omitempty"`
}

// Rule defines the single authorization rule.
type Rule struct {
	// Subjects contains the subject configuration.
	// If empty, all subjects are included.
	//
	// +optional
	Subjects []Subject `json:"subjects,omitempty"`

	// Permissions contains allowed HTTP methods.
	// If empty, all methods are matching.
	//
	// +optional
	Permissions []string `json:"permissions,omitempty"`

	// Action defines the action to be taken if the rule matches.
	Action RuleActionType `json:"action"`
}

// Subject contains the subject configuration.
type Subject struct {
	// ClientCIDR contains client cidr configuration.
	// Valid examples are "192.168.1.0/24" or "2001:db8::/64"
	//
	// +optional
	ClientCIDR *string `json:"clientCIDR,omitempty"`
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
