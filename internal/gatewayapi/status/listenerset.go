// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package status

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// UpdateListenerSetStatusAccepted updates the Accepted condition for the ListenerSet.
func UpdateListenerSetStatusAccepted(xls *gwapiv1.ListenerSet, accepted bool, reason gwapiv1.ListenerSetConditionReason, msg string) {
	status := metav1.ConditionFalse
	if accepted {
		status = metav1.ConditionTrue
	}
	cond := newCondition(string(gwapiv1.ListenerSetConditionAccepted), status, string(reason), msg, xls.Generation)
	xls.Status.Conditions = MergeConditions(xls.Status.Conditions, cond)
}

// UpdateListenerSetStatusProgrammed updates the Programmed condition for the ListenerSet.
func UpdateListenerSetStatusProgrammed(xls *gwapiv1.ListenerSet, programmed bool, reason gwapiv1.ListenerSetConditionReason, msg string) {
	status := metav1.ConditionFalse
	if programmed {
		status = metav1.ConditionTrue
	}
	cond := newCondition(string(gwapiv1.ListenerSetConditionProgrammed), status, string(reason), msg, xls.Generation)
	xls.Status.Conditions = MergeConditions(xls.Status.Conditions, cond)
}

// SetListenerSetListenerStatusCondition sets a condition for a specific listener in the ListenerSet.
func SetListenerSetListenerStatusCondition(xls *gwapiv1.ListenerSet, listenerStatusIdx int,
	conditionType gwapiv1.ListenerEntryConditionType, status metav1.ConditionStatus, reason gwapiv1.ListenerEntryConditionReason, message string,
) {
	cond := metav1.Condition{
		Type:               string(conditionType),
		Status:             status,
		Reason:             string(reason),
		Message:            message,
		ObservedGeneration: xls.Generation,
	}
	xls.Status.Listeners[listenerStatusIdx].Conditions = MergeConditions(xls.Status.Listeners[listenerStatusIdx].Conditions, cond)
}
