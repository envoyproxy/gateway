// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package status

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
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
		if string(ancestor.ControllerName) == controllerName && cmp.Equal(ancestor.AncestorRef, ancestorRef) {
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
		if string(ancestor.ControllerName) == controllerName && cmp.Equal(ancestor.AncestorRef, ancestorRef) {
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

func TruncatePolicyAncestors(policyStatus *gwapiv1a2.PolicyStatus, ancestorRef gwapiv1a2.ParentReference, controllerName string, generation int64) {
	aggregatedPolicyConditions := map[metav1.Condition][]string{}
	for _, ancestor := range policyStatus.Ancestors {
		for _, condition := range ancestor.Conditions {
			apc := metav1.Condition{
				Type:   condition.Type,
				Status: condition.Status,
				Reason: condition.Reason,
			}
			aggregatedPolicyConditions[apc] = append(aggregatedPolicyConditions[apc], string(ancestor.AncestorRef.Name))
		}
	}

	policyStatus.Ancestors = nil

	for apc, parents := range aggregatedPolicyConditions {
		sort.Strings(parents)

		SetConditionForPolicyAncestor(policyStatus,
			ancestorRef,
			controllerName,
			gwapiv1a2.PolicyConditionType(apc.Type),
			apc.Status,
			gwapiv1a2.PolicyConditionReason(apc.Reason),
			fmt.Sprintf("This policy has %d ancestors with %s condition in %s status due to %s reason. Aggregated ancestors: %s.",
				len(parents), apc.Type, apc.Status, apc.Reason, strings.Join(parents, ", ")),
			generation,
		)
	}

	SetConditionForPolicyAncestor(policyStatus,
		ancestorRef,
		controllerName,
		egv1a1.PolicyConditionAggregated,
		metav1.ConditionTrue,
		egv1a1.PolicyReasonAggregated,
		"Ancestors have been aggregated because the number of policy ancestors exceeds 16.",
		generation,
	)
}
