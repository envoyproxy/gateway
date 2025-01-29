// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

type PolicyTargetReferences struct {
	// TargetRef is the name of the resource this policy is being attached to.
	// This policy and the TargetRef MUST be in the same namespace for this
	// Policy to have effect
	//
	// Deprecated: use targetRefs/targetSelectors instead
	TargetRef *gwapiv1a2.LocalPolicyTargetReferenceWithSectionName `json:"targetRef,omitempty"`

	// TargetRefs are the names of the Gateway resources this policy
	// is being attached to.
	TargetRefs []gwapiv1a2.LocalPolicyTargetReferenceWithSectionName `json:"targetRefs,omitempty"`

	// TargetSelectors allow targeting resources for this policy based on labels
	TargetSelectors []TargetSelector `json:"targetSelectors,omitempty"`
}

// +kubebuilder:validation:XValidation:rule="has(self.group) ? self.group == 'gateway.networking.k8s.io' : true ", message="group must be gateway.networking.k8s.io"
type TargetSelector struct {
	// Group is the group that this selector targets. Defaults to gateway.networking.k8s.io
	//
	// +kubebuilder:default:="gateway.networking.k8s.io"
	Group *gwapiv1a2.Group `json:"group,omitempty"`

	// Kind is the resource kind that this selector targets.
	Kind gwapiv1a2.Kind `json:"kind"`

	// MatchLabels are the set of label selectors for identifying the targeted resource
	MatchLabels map[string]string `json:"matchLabels"`

	// Namespaces determines if the resource from all namespaces or the current namespace
	// are considered when matching by label selectors specified in MatchLabels.
	// Note: when referencing targets in a different namespace, appropriate ReferenceGrants must be
	// configured to allow the selection.
	// +optional
	// +notImplementedHide
	Namespaces *TargetSelectorNamespaces `json:"namespaces,omitempty"`
}

type FromNamespaces string

const (
	// FromNamespacesAll indicates that the target selector should apply to targets from all namespaces
	FromNamespacesAll FromNamespaces = "All"
)

// TargetSelectorNamespaces determines which namespaces are used when selecting policy targets.
type TargetSelectorNamespaces struct {
	// Indicates where targets would be selected for the Policy's TargetSelector.
	// +kubebuilder:validation:Enum=All
	// +kubebuilder:validation:Required
	FromNamespaces FromNamespaces `json:"from,omitempty"`
}

func (p PolicyTargetReferences) GetTargetRefs() []gwapiv1a2.LocalPolicyTargetReferenceWithSectionName {
	if p.TargetRef != nil {
		return []gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{*p.TargetRef}
	}
	return p.TargetRefs
}
