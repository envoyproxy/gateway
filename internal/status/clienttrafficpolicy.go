// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package status

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func SetClientTrafficPolicyCondition(c *egv1a1.ClientTrafficPolicy, conditionType gwv1a2.PolicyConditionType, status metav1.ConditionStatus, reason gwv1a2.PolicyConditionReason, message string) {
	cond := newCondition(string(conditionType), status, string(reason), message, time.Now(), c.Generation)
	c.Status.Conditions = MergeConditions(c.Status.Conditions, cond)
}
