// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

type PolicyTargetReferences struct {
	// TargetRef is the name of the Gateway resource this policy
	// is being attached to.
	// This Policy and the TargetRef MUST be in the same namespace
	// for this Policy to have effect and be applied to the Gateway.
	// TargetRef
	//
	// Deprecated: use targetRefs instead
	TargetRef *gwapiv1a2.LocalPolicyTargetReferenceWithSectionName `json:"targetRef,omitempty"`

	// TargetRefs are the names of the Gateway resources this policy
	// is being attached to.
	TargetRefs []gwapiv1a2.LocalPolicyTargetReferenceWithSectionName `json:"targetRefs,omitempty"`
}

func (p PolicyTargetReferences) GetTargetRefs() []gwapiv1a2.LocalPolicyTargetReferenceWithSectionName {
	if p.TargetRef != nil {
		return []gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{*p.TargetRef}
	}
	return p.TargetRefs
}
