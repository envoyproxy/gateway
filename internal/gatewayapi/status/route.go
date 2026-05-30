// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package status

import (
	"sort"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

const (
	routeConditionAggregated gwapiv1.RouteConditionType   = "Aggregated"
	routeReasonAggregated    gwapiv1.RouteConditionReason = "Aggregated"
)

func SetRouteStatusCondition(route *gwapiv1.RouteStatus, routeParentStatusIdx int, routeGeneration int64,
	conditionType gwapiv1.RouteConditionType, status metav1.ConditionStatus, reason gwapiv1.RouteConditionReason, message string,
) {
	cond := metav1.Condition{
		Type:               string(conditionType),
		Status:             status,
		Reason:             string(reason),
		Message:            message,
		ObservedGeneration: routeGeneration,
	}

	route.Parents[routeParentStatusIdx].Conditions = MergeConditions(route.Parents[routeParentStatusIdx].Conditions, cond)
}

// TruncateRouteParents trims RouteStatus.Parents down to at most 32 entries.
// The first 31 parents are shown as-is.
// The last 32nd parent is shown as-is and gets the Aggregated condition.
func TruncateRouteParents(routeStatus *gwapiv1.RouteStatus, generation int64) {
	if len(routeStatus.Parents) <= 32 {
		return
	}

	sort.Slice(routeStatus.Parents, func(i, j int) bool {
		a, b := routeStatus.Parents[i], routeStatus.Parents[j]
		aRank := sortRankForRouteParent(&a)
		bRank := sortRankForRouteParent(&b)

		if aRank != bRank {
			return aRank < bRank
		}

		aNamespace := ""
		if a.ParentRef.Namespace != nil {
			aNamespace = string(*a.ParentRef.Namespace)
		}
		bNamespace := ""
		if b.ParentRef.Namespace != nil {
			bNamespace = string(*b.ParentRef.Namespace)
		}
		if aNamespace != bNamespace {
			return aNamespace < bNamespace
		}

		if a.ParentRef.Name != b.ParentRef.Name {
			return string(a.ParentRef.Name) < string(b.ParentRef.Name)
		}

		aSection := ""
		if a.ParentRef.SectionName != nil {
			aSection = string(*a.ParentRef.SectionName)
		}
		bSection := ""
		if b.ParentRef.SectionName != nil {
			bSection = string(*b.ParentRef.SectionName)
		}
		if aSection != bSection {
			return aSection < bSection
		}

		aPort := int32(0)
		if a.ParentRef.Port != nil {
			aPort = *a.ParentRef.Port
		}
		bPort := int32(0)
		if b.ParentRef.Port != nil {
			bPort = *b.ParentRef.Port
		}
		if aPort != bPort {
			return aPort < bPort
		}

		aKind := ""
		if a.ParentRef.Kind != nil {
			aKind = string(*a.ParentRef.Kind)
		}
		bKind := ""
		if b.ParentRef.Kind != nil {
			bKind = string(*b.ParentRef.Kind)
		}
		if aKind != bKind {
			return aKind < bKind
		}

		aGroup := ""
		if a.ParentRef.Group != nil {
			aGroup = string(*a.ParentRef.Group)
		}
		bGroup := ""
		if b.ParentRef.Group != nil {
			bGroup = string(*b.ParentRef.Group)
		}
		if aGroup != bGroup {
			return aGroup < bGroup
		}

		return string(a.ControllerName) < string(b.ControllerName)
	})

	const maxParents = 32
	aggregatedMessage := "Parents have been truncated because the number of route parents exceeds 32."

	routeStatus.Parents = routeStatus.Parents[:maxParents]
	SetRouteStatusCondition(routeStatus,
		maxParents-1,
		generation,
		routeConditionAggregated,
		metav1.ConditionTrue,
		routeReasonAggregated,
		aggregatedMessage,
	)
}

// sortRankForRouteParent returns an integer sort key.
// ranking rules (smaller value -> higher priority):
//
//   - The parent is not accepted (Accepted == false).
//   - The parent has unresolved references (ResolvedRefs == false).
//   - All other cases.
func sortRankForRouteParent(parent *gwapiv1.RouteParentStatus) int {
	switch {
	case meta.IsStatusConditionFalse(parent.Conditions, string(gwapiv1.RouteConditionAccepted)):
		return 0
	case meta.IsStatusConditionFalse(parent.Conditions, string(gwapiv1.RouteConditionResolvedRefs)):
		return 1
	default:
		return 2
	}
}
