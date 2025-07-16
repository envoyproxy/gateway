// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package status

import (
	"cmp"
	"fmt"
	"slices"
	"strings"
	"time"

	gocmp "github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

const (
	conditionMessageMaxLength = 32768
)

type PolicyResolveError struct {
	Reason  gwapiv1a2.PolicyConditionReason
	Message string

	error
}

func SetResolveErrorForPolicyAncestors(policyStatus *gwapiv1a2.PolicyStatus, ancestorRefs []gwapiv1a2.ParentReference, controllerName string, generation int64, resolveErr *PolicyResolveError) {
	for _, ancestorRef := range ancestorRefs {
		SetConditionForPolicyAncestor(policyStatus, ancestorRef, controllerName,
			gwapiv1a2.PolicyConditionAccepted, metav1.ConditionFalse, resolveErr.Reason, resolveErr.Message, generation)
	}
}

func SetTranslationErrorForPolicyAncestors(policyStatus *gwapiv1a2.PolicyStatus, ancestorRefs []gwapiv1a2.ParentReference, controllerName string, generation int64, errMsg string) {
	for _, ancestorRef := range ancestorRefs {
		SetConditionForPolicyAncestor(policyStatus, ancestorRef, controllerName,
			gwapiv1a2.PolicyConditionAccepted, metav1.ConditionFalse, gwapiv1a2.PolicyReasonInvalid, errMsg, generation)
	}
}

// SetAcceptedForPolicyAncestors sets accepted conditions for each ancestor reference if it is unset.
func SetAcceptedForPolicyAncestors(policyStatus *gwapiv1a2.PolicyStatus, ancestorRefs []gwapiv1a2.ParentReference, controllerName string, generation int64) {
	for _, ancestorRef := range ancestorRefs {
		setAcceptedForPolicyAncestor(policyStatus, ancestorRef, controllerName, generation)
	}
}

func setAcceptedForPolicyAncestor(policyStatus *gwapiv1a2.PolicyStatus, ancestorRef gwapiv1a2.ParentReference, controllerName string, generation int64) {
	// Return early if Accepted condition is already set for specific ancestor.
	for _, ancestor := range policyStatus.Ancestors {
		if string(ancestor.ControllerName) == controllerName && gocmp.Equal(ancestor.AncestorRef, ancestorRef) {
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

func SetConditionForPolicyAncestors(policyStatus *gwapiv1a2.PolicyStatus, ancestorRefs []gwapiv1a2.ParentReference, controllerName string,
	conditionType gwapiv1a2.PolicyConditionType, status metav1.ConditionStatus, reason gwapiv1a2.PolicyConditionReason, message string, generation int64,
) {
	for _, ancestorRef := range ancestorRefs {
		SetConditionForPolicyAncestor(policyStatus, ancestorRef, controllerName, conditionType, status, reason, message, generation)
	}
}

func SetConditionForPolicyAncestor(policyStatus *gwapiv1a2.PolicyStatus, ancestorRef gwapiv1a2.ParentReference, controllerName string,
	conditionType gwapiv1a2.PolicyConditionType, status metav1.ConditionStatus, reason gwapiv1a2.PolicyConditionReason, message string, generation int64,
) {
	if policyStatus.Ancestors == nil {
		policyStatus.Ancestors = []gwapiv1a2.PolicyAncestorStatus{}
	}

	cond := newCondition(string(conditionType), status, string(reason), message, time.Now(), generation)

	// Add condition for exist PolicyAncestorStatus.
	for i, ancestor := range policyStatus.Ancestors {
		if string(ancestor.ControllerName) == controllerName && gocmp.Equal(ancestor.AncestorRef, ancestorRef) {
			policyStatus.Ancestors[i].Conditions = MergeConditions(policyStatus.Ancestors[i].Conditions, cond)
			return
		}
	}

	// Add condition for new PolicyAncestorStatus.
	policyStatus.Ancestors = append(policyStatus.Ancestors, gwapiv1a2.PolicyAncestorStatus{
		AncestorRef:    ancestorRef,
		ControllerName: gwapiv1a2.GatewayController(controllerName),
		Conditions:     []metav1.Condition{cond},
	})
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
	slices.SortStableFunc(policyStatus.Ancestors, func(a, b gwapiv1a2.PolicyAncestorStatus) int {
		if r := cmp.Compare(sortRankForPolicyAncestor(a), sortRankForPolicyAncestor(b)); r != 0 {
			return r
		}
		return strings.Compare(string(a.AncestorRef.Name), string(b.AncestorRef.Name))
	})

	aggregated := make([]string, len(policyStatus.Ancestors)-16)
	for i, ancestor := range policyStatus.Ancestors[16:] {
		aggregated[i] = string(ancestor.AncestorRef.Name)
	}
	aggregatedMessage := fmt.Sprintf("Ancestors have been aggregated because the number of policy ancestors exceeds 16. "+
		"The aggregated ancestors: %s", strings.Join(aggregated, ", "))
	if len(aggregatedMessage) > conditionMessageMaxLength {
		aggregatedMessage = aggregatedMessage[:conditionMessageMaxLength]
	}

	policyStatus.Ancestors = policyStatus.Ancestors[:16]
	SetConditionForPolicyAncestor(policyStatus,
		policyStatus.Ancestors[15].AncestorRef,
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
func sortRankForPolicyAncestor(ancestor gwapiv1a2.PolicyAncestorStatus) int {
	switch {
	case meta.IsStatusConditionFalse(ancestor.Conditions, string(gwapiv1a2.PolicyConditionAccepted)):
		return 0
	case meta.IsStatusConditionTrue(ancestor.Conditions, string(egv1a1.PolicyReasonOverridden)):
		return 1
	default:
		return 2
	}
}
