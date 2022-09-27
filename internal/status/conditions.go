// Portions of this code are based on code from Contour, available at:
// https://github.com/projectcontour/contour/blob/main/internal/status/gatewayclassconditions.go

package status

import (
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

const ReasonOlderGatewayClassExists gwapiv1b1.GatewayClassConditionReason = "OlderGatewayClassExists"

// computeGatewayClassAcceptedCondition computes the GatewayClass Accepted status condition.
func computeGatewayClassAcceptedCondition(gatewayClass *gwapiv1b1.GatewayClass, accepted bool) metav1.Condition {
	switch accepted {
	case true:
		return metav1.Condition{
			Type:               string(gwapiv1b1.GatewayClassConditionStatusAccepted),
			Status:             metav1.ConditionTrue,
			Reason:             string(gwapiv1b1.GatewayClassReasonAccepted),
			Message:            "Valid GatewayClass",
			ObservedGeneration: gatewayClass.Generation,
			LastTransitionTime: metav1.NewTime(time.Now()),
		}
	default:
		return metav1.Condition{
			Type:               string(gwapiv1b1.GatewayClassConditionStatusAccepted),
			Status:             metav1.ConditionFalse,
			Reason:             string(ReasonOlderGatewayClassExists),
			Message:            "Invalid GatewayClass: another older GatewayClass with the same Spec.Controller exists",
			ObservedGeneration: gatewayClass.Generation,
			LastTransitionTime: metav1.NewTime(time.Now()),
		}
	}
}

// computeGatewayScheduledCondition computes the Gateway Scheduled status condition.
func computeGatewayScheduledCondition(gw *gwapiv1b1.Gateway, scheduled bool) metav1.Condition {
	switch scheduled {
	case true:
		return newCondition(string(gwapiv1b1.GatewayConditionScheduled), metav1.ConditionTrue,
			string(gwapiv1b1.GatewayReasonScheduled),
			"The Gateway has been scheduled by Envoy Gateway", time.Now(), gw.Generation)
	default:
		return newCondition(string(gwapiv1b1.GatewayConditionScheduled), metav1.ConditionFalse,
			string(gwapiv1b1.GatewayReasonScheduled),
			"The Gateway has not been scheduled by Envoy Gateway", time.Now(), gw.Generation)
	}
}

// computeGatewayReadyCondition computes the Gateway Ready status condition.
// Ready condition surfaces true when the Envoy Deployment status is ready.
func computeGatewayReadyCondition(gw *gwapiv1b1.Gateway, deployment *appsv1.Deployment) metav1.Condition {
	if len(gw.Status.Addresses) == 0 {
		return newCondition(string(gwapiv1b1.GatewayConditionReady), metav1.ConditionFalse,
			string(gwapiv1b1.GatewayReasonAddressNotAssigned),
			"No addresses have been assigned to the Gateway", time.Now(), gw.Generation)
	}

	// If there are no available replicas for the Envoy Deployment, don't
	// mark the Gateway as ready yet.

	if deployment == nil || deployment.Status.AvailableReplicas == 0 {
		return newCondition(string(gwapiv1b1.GatewayConditionReady), metav1.ConditionFalse,
			string(gwapiv1b1.GatewayReasonNoResources),
			"Deployment replicas unavailable", time.Now(), gw.Generation)
	}

	message := fmt.Sprintf("Address assigned to the Gateway, %d/%d envoy Deployment replicas available",
		deployment.Status.AvailableReplicas, deployment.Status.Replicas)
	return newCondition(string(gwapiv1b1.GatewayConditionReady), metav1.ConditionTrue,
		string(gwapiv1b1.GatewayReasonReady), message, time.Now(), gw.Generation)
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
	return a.Status != b.Status || a.Reason != b.Reason || a.Message != b.Message
}
