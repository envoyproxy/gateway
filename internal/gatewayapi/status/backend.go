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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

// UpdateBackendStatusAcceptedCondition updates the status condition for the provided Backend based on the accepted state.
func UpdateBackendStatusAcceptedCondition(be *egv1a1.Backend, accepted bool, msg string) *egv1a1.Backend {
	be.Status.Conditions = MergeConditions(be.Status.Conditions, computeBackendAcceptedCondition(be, accepted, msg))
	return be
}

// computeBackendAcceptedCondition computes the Backend Accepted status condition.
func computeBackendAcceptedCondition(be *egv1a1.Backend, accepted bool, msg string) metav1.Condition {
	switch accepted {
	case true:
		return newCondition(string(egv1a1.BackendReasonAccepted), metav1.ConditionTrue,
			string(egv1a1.BackendConditionAccepted),
			"The Backend was accepted", time.Now(), be.Generation)
	default:
		return newCondition(string(egv1a1.BackendReasonInvalid), metav1.ConditionFalse,
			string(egv1a1.BackendConditionAccepted),
			msg, time.Now(), be.Generation)
	}
}
