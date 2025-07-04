// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

const (
	// PolicyConditionAggregated indicates whether the policy has been aggregated
	// to satisfy CEL constraints in PolicyAncestorStatus (not exceeding 16).
	//
	// Possible reasons for this condition to be True are:
	//
	// * "Aggregated"
	//
	PolicyConditionAggregated gwapiv1a2.PolicyConditionType = "Aggregated"

	// PolicyReasonAggregated is used with the "Aggregated" condition when the policy
	// is aggregated to satisfy CEL constraints in PolicyAncestorStatus (not exceeding 16).
	PolicyReasonAggregated gwapiv1a2.PolicyConditionReason = "Aggregated"
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
	Group *gwapiv1.Group `json:"group,omitempty"`

	// Kind is the resource kind that this selector targets.
	Kind gwapiv1.Kind `json:"kind"`

	// MatchLabels are the set of label selectors for identifying the targeted resource
	// +optional
	MatchLabels map[string]string `json:"matchLabels,omitempty"`

	// MatchExpressions is a list of label selector requirements. The requirements are ANDed.
	//
	// +optional
	// +listType=atomic
	MatchExpressions []metav1.LabelSelectorRequirement `json:"matchExpressions,omitempty"`
}

func (p PolicyTargetReferences) GetTargetRefs() []gwapiv1a2.LocalPolicyTargetReferenceWithSectionName {
	if p.TargetRef != nil {
		return []gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{*p.TargetRef}
	}
	return p.TargetRefs
}
