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
func computeGatewayClassAcceptedCondition(gatewayClass *gwapiv1b1.GatewayClass, accepted bool, reason, msg string) metav1.Condition {
	switch accepted {
	case true:
		return metav1.Condition{
			Type:               string(gwapiv1b1.GatewayClassConditionStatusAccepted),
			Status:             metav1.ConditionTrue,
			Reason:             reason,
			Message:            msg,
			ObservedGeneration: gatewayClass.Generation,
			LastTransitionTime: metav1.NewTime(time.Now()),
		}
	default:
		return metav1.Condition{
			Type:               string(gwapiv1b1.GatewayClassConditionStatusAccepted),
			Status:             metav1.ConditionFalse,
			Reason:             reason,
			Message:            msg,
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
