// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package status

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
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
		LastTransitionTime: metav1.NewTime(time.Now()),
	}

	route.Parents[routeParentStatusIdx].Conditions = MergeConditions(route.Parents[routeParentStatusIdx].Conditions, cond)
}
