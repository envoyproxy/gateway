// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// This file contains code derived from Contour,
// https://github.com/projectcontour/contour
// from the source file
// https://github.com/projectcontour/contour/blob/main/internal/status/gatewayclass.go
// and is provided here subject to the following:
// Copyright Project Contour Authors
// SPDX-License-Identifier: Apache-2.0

package status

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

const (
	ReasonOlderGatewayClassExists gwapiv1.GatewayClassConditionReason = "OlderGatewayClassExists"

	MsgOlderGatewayClassExists   = "Invalid GatewayClass: another older GatewayClass with the same Spec.Controller exists"
	MsgValidGatewayClass         = "Valid GatewayClass"
	MsgGatewayClassInvalidParams = "Invalid parametersRef"
)

// SetGatewayClassAccepted inserts or updates the Accepted condition
// for the provided GatewayClass.
func SetGatewayClassAccepted(gc *gwapiv1.GatewayClass, accepted bool, reason, msg string) *gwapiv1.GatewayClass {
	gc.Status.Conditions = MergeConditions(gc.Status.Conditions, computeGatewayClassAcceptedCondition(gc, accepted, reason, msg))
	return gc
}

// computeGatewayClassAcceptedCondition computes the GatewayClass Accepted status condition.
func computeGatewayClassAcceptedCondition(gatewayClass *gwapiv1.GatewayClass,
	accepted bool,
	reason, msg string,
) metav1.Condition {
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
