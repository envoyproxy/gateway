// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package status

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapixv1a1 "sigs.k8s.io/gateway-api/apisx/v1alpha1"
)

// UpdateXListenerSetStatusAccepted updates the Accepted condition for the XListenerSet.
func UpdateXListenerSetStatusAccepted(xls *gwapixv1a1.XListenerSet, accepted bool, reason gwapixv1a1.ListenerSetConditionReason, msg string) {
	status := metav1.ConditionFalse
	if accepted {
		status = metav1.ConditionTrue
	}
	cond := newCondition(string(gwapixv1a1.ListenerSetConditionAccepted), status, string(reason), msg, xls.Generation)
	xls.Status.Conditions = MergeConditions(xls.Status.Conditions, cond)
}

// UpdateXListenerSetStatusProgrammed updates the Programmed condition for the XListenerSet.
func UpdateXListenerSetStatusProgrammed(xls *gwapixv1a1.XListenerSet, programmed bool, reason gwapixv1a1.ListenerSetConditionReason, msg string) {
	status := metav1.ConditionFalse
	if programmed {
		status = metav1.ConditionTrue
	}
	cond := newCondition(string(gwapixv1a1.ListenerSetConditionProgrammed), status, string(reason), msg, xls.Generation)
	xls.Status.Conditions = MergeConditions(xls.Status.Conditions, cond)
}

// SetXListenerSetListenerStatusCondition sets a condition for a specific listener in the XListenerSet.
func SetXListenerSetListenerStatusCondition(xls *gwapixv1a1.XListenerSet, listenerStatusIdx int,
	conditionType gwapixv1a1.ListenerEntryConditionType, status metav1.ConditionStatus, reason gwapixv1a1.ListenerEntryConditionReason, message string,
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
