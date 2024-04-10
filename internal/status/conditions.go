// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// This file contains code derived from Contour,
// https://github.com/projectcontour/contour
// from the source file
// https://github.com/projectcontour/contour/blob/main/internal/status/gatewayclassconditions.go
// and is provided here subject to the following:
// Copyright Project Contour Authors
// SPDX-License-Identifier: Apache-2.0

package status

import (
	"fmt"
	"time"
	"unicode"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

const (
	ReasonOlderGatewayClassExists gwapiv1.GatewayClassConditionReason = "OlderGatewayClassExists"

	MsgOlderGatewayClassExists   = "Invalid GatewayClass: another older GatewayClass with the same Spec.Controller exists"
	MsgValidGatewayClass         = "Valid GatewayClass"
	MsgGatewayClassInvalidParams = "Invalid parametersRef"
)

// computeGatewayClassAcceptedCondition computes the GatewayClass Accepted status condition.
func computeGatewayClassAcceptedCondition(gatewayClass *gwapiv1.GatewayClass,
	accepted bool,
	reason, msg string) metav1.Condition {
	switch accepted {
	case true:
		return metav1.Condition{
			Type:               string(gwapiv1.GatewayClassConditionStatusAccepted),
			Status:             metav1.ConditionTrue,
			Reason:             reason,
			Message:            msg,
			ObservedGeneration: gatewayClass.Generation,
			LastTransitionTime: metav1.NewTime(time.Now()),
		}
	default:
		return metav1.Condition{
			Type:               string(gwapiv1.GatewayClassConditionStatusAccepted),
			Status:             metav1.ConditionFalse,
			Reason:             reason,
			Message:            msg,
			ObservedGeneration: gatewayClass.Generation,
			LastTransitionTime: metav1.NewTime(time.Now()),
		}
	}
}

// computeGatewayAcceptedCondition computes the Gateway Accepted status condition.
func computeGatewayAcceptedCondition(gw *gwapiv1.Gateway, accepted bool) metav1.Condition {
	switch accepted {
	case true:
		return newCondition(string(gwapiv1.GatewayReasonAccepted), metav1.ConditionTrue,
			string(gwapiv1.GatewayReasonAccepted),
			"The Gateway has been scheduled by Envoy Gateway", time.Now(), gw.Generation)
	default:
		return newCondition(string(gwapiv1.GatewayReasonAccepted), metav1.ConditionFalse,
			string(gwapiv1.GatewayReasonAccepted),
			"The Gateway has not been scheduled by Envoy Gateway", time.Now(), gw.Generation)
	}
}

// computeGatewayProgrammedCondition computes the Gateway Programmed status condition.
// Programmed condition surfaces true when the Envoy Deployment status is ready.
func computeGatewayProgrammedCondition(gw *gwapiv1.Gateway, deployment *appsv1.Deployment) metav1.Condition {
	if len(gw.Status.Addresses) == 0 {
		return newCondition(string(gwapiv1.GatewayConditionProgrammed), metav1.ConditionFalse,
			string(gwapiv1.GatewayReasonAddressNotAssigned),
			"No addresses have been assigned to the Gateway", time.Now(), gw.Generation)
	}

	// If there are no available replicas for the Envoy Deployment, don't
	// mark the Gateway as ready yet.

	if deployment == nil || deployment.Status.AvailableReplicas == 0 {
		return newCondition(string(gwapiv1.GatewayConditionProgrammed), metav1.ConditionFalse,
			string(gwapiv1.GatewayReasonNoResources),
			"Deployment replicas unavailable", time.Now(), gw.Generation)
	}

	message := fmt.Sprintf("Address assigned to the Gateway, %d/%d envoy Deployment replicas available",
		deployment.Status.AvailableReplicas, deployment.Status.Replicas)
	return newCondition(string(gwapiv1.GatewayConditionProgrammed), metav1.ConditionTrue,
		string(gwapiv1.GatewayConditionProgrammed), message, time.Now(), gw.Generation)
}

// MergeConditions adds or updates matching conditions, and updates the transition
// time if details of a condition have changed. Returns the updated condition array.
func MergeConditions(conditions []metav1.Condition, updates ...metav1.Condition) []metav1.Condition {
	var additions []metav1.Condition
	for i, update := range updates {
		add := true
		for j, cond := range conditions {
			if cond.Type == update.Type {
				add = false
				if conditionChanged(cond, update) {
					conditions[j].Status = update.Status
					conditions[j].Reason = update.Reason
					conditions[j].Message = update.Message
					conditions[j].ObservedGeneration = update.ObservedGeneration
					conditions[j].LastTransitionTime = update.LastTransitionTime
					break
				}
			}
		}
		if add {
			additions = append(additions, updates[i])
		}
	}
	conditions = append(conditions, additions...)
	return conditions
}

func newCondition(t string, status metav1.ConditionStatus, reason, msg string, lt time.Time, og int64) metav1.Condition {
	return metav1.Condition{
		Type:               t,
		Status:             status,
		Reason:             reason,
		Message:            msg,
		LastTransitionTime: metav1.NewTime(lt),
		ObservedGeneration: og,
	}
}

func conditionChanged(a, b metav1.Condition) bool {
	return (a.Status != b.Status) ||
		(a.Reason != b.Reason) ||
		(a.Message != b.Message) ||
		(a.ObservedGeneration != b.ObservedGeneration)
}

// Error2ConditionMsg format the error string to a Status condition message.
// * Convert the first letter to capital
// * Append "." to the string if it doesn't exit
func Error2ConditionMsg(err error) string {
	if err == nil {
		return ""
	}

	message := err.Error()
	if message == "" {
		return message
	}

	// Convert the string to a rune slice for easier manipulation
	runes := []rune(message)

	// Check if the first rune is a letter and convert it to uppercase
	if unicode.IsLetter(runes[0]) {
		runes[0] = unicode.ToUpper(runes[0])
	}

	// Convert the rune slice back to a string
	return string(runes)
}
