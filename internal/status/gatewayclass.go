// Portions of this code are based on code from Contour, available at:
// https://github.com/projectcontour/contour/blob/main/internal/status/gatewayclass.go

package status

import (
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

// SetGatewayClassAccepted inserts or updates the Accepted condition
// for the provided GatewayClass.
func SetGatewayClassAccepted(gc *gwapiv1b1.GatewayClass, accepted bool) *gwapiv1b1.GatewayClass {
	gc.Status.Conditions = mergeConditions(gc.Status.Conditions, computeGatewayClassAcceptedCondition(gc, accepted))
	return gc
}
