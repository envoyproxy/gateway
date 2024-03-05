// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package status

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

func SetBackendTLSPolicyCondition(c *gwv1a2.BackendTLSPolicy, policyAnces gwv1a2.PolicyAncestorStatus, conditionType gwv1a2.PolicyConditionType, status metav1.ConditionStatus, reason gwv1a2.PolicyConditionReason, message string) {

	if c.Status.Ancestors == nil {
		c.Status.Ancestors = []gwv1a2.PolicyAncestorStatus{}
	}

	cond := newCondition(string(conditionType), status, string(reason), message, time.Now(), c.Generation)
	for i, ancestor := range c.Status.Ancestors {
		if ancestor.AncestorRef.Name == policyAnces.AncestorRef.Name &&
			(ancestor.AncestorRef.Namespace == nil || *ancestor.AncestorRef.Namespace == *policyAnces.AncestorRef.Namespace) {
			c.Status.Ancestors[i].Conditions = MergeConditions(c.Status.Ancestors[i].Conditions, cond)
			return
		}
	}
	len := len(c.Status.Ancestors)
	c.Status.Ancestors = append(c.Status.Ancestors, policyAnces)
	c.Status.Ancestors[len].Conditions = MergeConditions(c.Status.Ancestors[len].Conditions, cond)
}
