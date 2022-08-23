package status

import gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

// SetGatewayScheduled adds or updates the Scheduled condition for the provided Gateway.
func SetGatewayScheduled(gw *gwapiv1b1.Gateway, scheduled bool) *gwapiv1b1.Gateway {
	gw.Status.Conditions = mergeConditions(gw.Status.Conditions, computeGatewayScheduledCondition(gw, scheduled))
	return gw
}
