// Portions of this code are based on code from Contour, available at:
// https://github.com/projectcontour/contour/blob/main/internal/status/gatewayclassconditions.go

package status

import (
	"time"

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

// mergeConditions adds or updates matching conditions, and updates the transition
// time if details of a condition have changed. Returns the updated condition array.
func mergeConditions(conditions []metav1.Condition, updates ...metav1.Condition) []metav1.Condition {
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

// TODO: Pass the Generation so we can set ObservedGeneration.
// xref: https://github.com/envoyproxy/gateway/issues/166
func newCondition(t string, status metav1.ConditionStatus, reason, msg string, lt time.Time) metav1.Condition {
	return metav1.Condition{
		Type:               t,
		Status:             status,
		Reason:             reason,
		Message:            msg,
		LastTransitionTime: metav1.NewTime(lt),
	}
}

func conditionChanged(a, b metav1.Condition) bool {
	return a.Status != b.Status || a.Reason != b.Reason || a.Message != b.Message
}
