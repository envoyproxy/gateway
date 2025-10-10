// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package status

import (
	"sort"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

type PolicyResolveError struct {
	Reason  gwapiv1a2.PolicyConditionReason
	Message string

	error
}

func SetResolveErrorForPolicyAncestors(policyStatus *gwapiv1a2.PolicyStatus, ancestorRefs []*gwapiv1a2.ParentReference, controllerName string, generation int64, resolveErr *PolicyResolveError) {
	for _, ancestorRef := range ancestorRefs {
		SetConditionForPolicyAncestor(policyStatus, ancestorRef, controllerName,
			gwapiv1a2.PolicyConditionAccepted, metav1.ConditionFalse, resolveErr.Reason, resolveErr.Message, generation)
	}
}

func SetResolveErrorForPolicyAncestor(policyStatus *gwapiv1a2.PolicyStatus, ancestorRef *gwapiv1a2.ParentReference, controllerName string, generation int64, resolveErr *PolicyResolveError) {
	SetConditionForPolicyAncestor(policyStatus, ancestorRef, controllerName,
		gwapiv1a2.PolicyConditionAccepted, metav1.ConditionFalse, resolveErr.Reason, resolveErr.Message, generation)
}

func SetTranslationErrorForPolicyAncestors(policyStatus *gwapiv1a2.PolicyStatus, ancestorRefs []*gwapiv1a2.ParentReference, controllerName string, generation int64, errMsg string) {
	for _, ancestorRef := range ancestorRefs {
		SetConditionForPolicyAncestor(policyStatus, ancestorRef, controllerName,
			gwapiv1a2.PolicyConditionAccepted, metav1.ConditionFalse, gwapiv1a2.PolicyReasonInvalid, errMsg, generation)
	}
}

func SetTranslationErrorForPolicyAncestor(policyStatus *gwapiv1a2.PolicyStatus, ancestorRef *gwapiv1a2.ParentReference, controllerName string, generation int64, errMsg string) {
	SetConditionForPolicyAncestor(policyStatus, ancestorRef, controllerName,
		gwapiv1a2.PolicyConditionAccepted, metav1.ConditionFalse, gwapiv1a2.PolicyReasonInvalid, errMsg, generation)
}

// SetAcceptedForPolicyAncestors sets accepted conditions for each ancestor reference if it is unset.
func SetAcceptedForPolicyAncestors(policyStatus *gwapiv1a2.PolicyStatus, ancestorRefs []*gwapiv1a2.ParentReference, controllerName string, generation int64) {
	for _, ancestorRef := range ancestorRefs {
		SetAcceptedForPolicyAncestor(policyStatus, ancestorRef, controllerName, generation)
	}
}

func SetAcceptedForPolicyAncestor(policyStatus *gwapiv1a2.PolicyStatus, ancestorRef *gwapiv1a2.ParentReference, controllerName string, generation int64) {
	// Return early if Accepted condition is already set for specific ancestor.
	for _, ancestor := range policyStatus.Ancestors {
		if string(ancestor.ControllerName) == controllerName && ancestorRefsEqual(&ancestor.AncestorRef, ancestorRef) {
			for _, c := range ancestor.Conditions {
				if c.Type == string(gwapiv1a2.PolicyConditionAccepted) {
					return
				}
			}
		}
	}

	message := "Policy has been accepted."
	SetConditionForPolicyAncestor(policyStatus, ancestorRef, controllerName,
		gwapiv1a2.PolicyConditionAccepted, metav1.ConditionTrue, gwapiv1a2.PolicyReasonAccepted, message, generation)
}

func SetConditionForPolicyAncestors(policyStatus *gwapiv1a2.PolicyStatus, ancestorRefs []*gwapiv1a2.ParentReference, controllerName string,
	conditionType gwapiv1a2.PolicyConditionType, status metav1.ConditionStatus, reason gwapiv1a2.PolicyConditionReason, message string, generation int64,
) {
	for _, ancestorRef := range ancestorRefs {
		SetConditionForPolicyAncestor(policyStatus, ancestorRef, controllerName, conditionType, status, reason, message, generation)
	}
}

func SetConditionForPolicyAncestor(policyStatus *gwapiv1a2.PolicyStatus, ancestorRef *gwapiv1a2.ParentReference, controllerName string,
	conditionType gwapiv1a2.PolicyConditionType, status metav1.ConditionStatus, reason gwapiv1a2.PolicyConditionReason, message string, generation int64,
) {
	if policyStatus.Ancestors == nil {
		policyStatus.Ancestors = []gwapiv1a2.PolicyAncestorStatus{}
	}

	sanitizedMessage := truncateConditionMessage(message)

	// Find existing ancestor first
	for i, ancestor := range policyStatus.Ancestors {
		if string(ancestor.ControllerName) == controllerName && ancestorRefsEqual(&ancestor.AncestorRef, ancestorRef) {
			// if condition already exists and is unchanged, exit early
			for _, existingCond := range ancestor.Conditions {
				if existingCond.Type == string(conditionType) &&
					existingCond.Status == status &&
					existingCond.Reason == string(reason) &&
					existingCond.Message == sanitizedMessage &&
					existingCond.ObservedGeneration == generation {
					return
				}
			}

			// Only create condition and merge if needed
			cond := newCondition(string(conditionType), status, string(reason), sanitizedMessage, time.Now(), generation)
			policyStatus.Ancestors[i].Conditions = MergeConditions(policyStatus.Ancestors[i].Conditions, cond)
			return
		}
	}

	// Add condition for new PolicyAncestorStatus
	cond := newCondition(string(conditionType), status, string(reason), sanitizedMessage, time.Now(), generation)
	policyStatus.Ancestors = append(policyStatus.Ancestors, gwapiv1a2.PolicyAncestorStatus{
		AncestorRef:    *ancestorRef,
		ControllerName: gwapiv1a2.GatewayController(controllerName),
		Conditions:     []metav1.Condition{cond},
	})
}

func ancestorRefsEqual(a, b *gwapiv1a2.ParentReference) bool {
	if a == nil || b == nil {
		return a == b
	}

	// Compare non-pointer fields first (fastest)
	if a.Name != b.Name {
		return false
	}

	// Compare Group pointers
	if (a.Group == nil) != (b.Group == nil) {
		return false
	}
	if a.Group != nil && *a.Group != *b.Group {
		return false
	}

	// Compare Kind pointers
	if (a.Kind == nil) != (b.Kind == nil) {
		return false
	}
	if a.Kind != nil && *a.Kind != *b.Kind {
		return false
	}

	// Compare Namespace pointers
	if (a.Namespace == nil) != (b.Namespace == nil) {
		return false
	}
	if a.Namespace != nil && *a.Namespace != *b.Namespace {
		return false
	}

	// Compare SectionName pointers
	if (a.SectionName == nil) != (b.SectionName == nil) {
		return false
	}
	if a.SectionName != nil && *a.SectionName != *b.SectionName {
		return false
	}

	return true
}

// TruncatePolicyAncestors trims PolicyStatus.Ancestors down to at most 16 entries.
// The first 15 ancestors are shown as is.
// The last 16th ancestor is shown as is and add Aggregated condition.
func TruncatePolicyAncestors(policyStatus *gwapiv1a2.PolicyStatus, controllerName string, generation int64) {
	if len(policyStatus.Ancestors) <= 16 {
		return
	}

	// we need to truncate policy ancestor status due to the item limit (max 16).
	// so we are choosing to preserve the 16 most important ancestors.
	// negative polarity (Conflicted, Overridden...) should be clearly indicated to the user.
	sort.Slice(policyStatus.Ancestors, func(i, j int) bool {
		a, b := policyStatus.Ancestors[i], policyStatus.Ancestors[j]
		aRank := sortRankForPolicyAncestor(&a)
		bRank := sortRankForPolicyAncestor(&b)

		if aRank != bRank {
			return aRank < bRank
		}
		// First compare by namespace, then by name
		aNamespace := ""
		if a.AncestorRef.Namespace != nil {
			aNamespace = string(*a.AncestorRef.Namespace)
		}
		bNamespace := ""
		if b.AncestorRef.Namespace != nil {
			bNamespace = string(*b.AncestorRef.Namespace)
		}

		if aNamespace != bNamespace {
			return aNamespace < bNamespace
		}
		return string(a.AncestorRef.Name) < string(b.AncestorRef.Name)
	})

	aggregatedMessage := "Ancestors have been truncated because the number of policy ancestors exceeds 16."

	policyStatus.Ancestors = policyStatus.Ancestors[:16]
	SetConditionForPolicyAncestor(policyStatus,
		&policyStatus.Ancestors[15].AncestorRef,
		controllerName,
		egv1a1.PolicyConditionAggregated,
		metav1.ConditionTrue,
		egv1a1.PolicyReasonAggregated,
		aggregatedMessage,
		generation,
	)
}

// sortRankForPolicyAncestor returns an integer sort key.
// ranking rules (smaller value -> higher priority):
//
//	– The ancestor is not accepted (Accepted == false).
//	– The ancestor is accepted and overridden (Override == true).
//	– All other cases.
func sortRankForPolicyAncestor(ancestor *gwapiv1a2.PolicyAncestorStatus) int {
	switch {
	case meta.IsStatusConditionFalse(ancestor.Conditions, string(gwapiv1a2.PolicyConditionAccepted)):
		return 0
	case meta.IsStatusConditionTrue(ancestor.Conditions, string(egv1a1.PolicyReasonOverridden)):
		return 1
	default:
		return 2
	}
}
