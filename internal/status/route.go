// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package status

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func SetConditionForRoute(route *gwv1.RouteStatus, routeParentStatusIdx int, routeGeneration int64,
	conditionType gwv1.RouteConditionType, status metav1.ConditionStatus, reason gwv1.RouteConditionReason, message string) {
	cond := metav1.Condition{
		Type:               string(conditionType),
		Status:             status,
		Reason:             string(reason),
		Message:            message,
		ObservedGeneration: routeGeneration,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}

	idx := -1
	for i, existing := range route.Parents[routeParentStatusIdx].Conditions {
		if existing.Type == cond.Type {
			// return early if the condition is unchanged
			if existing.Status == cond.Status &&
				existing.Reason == cond.Reason &&
				existing.Message == cond.Message &&
				existing.ObservedGeneration == cond.ObservedGeneration {
				return
			}
			idx = i
			break
		}
	}

	if idx > -1 {
		route.Parents[routeParentStatusIdx].Conditions[idx] = cond
	} else {
		route.Parents[routeParentStatusIdx].Conditions = append(route.Parents[routeParentStatusIdx].Conditions, cond)
	}
}
