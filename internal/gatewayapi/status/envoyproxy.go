// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package status

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func UpdateEnvoyProxyStatusAccepted(ep *egv1a1.EnvoyProxy, ancestor *gwapiv1.ParentReference,
	reason egv1a1.EnvoyProxyConditionReason, msg string,
) {
	if ep == nil || ancestor == nil {
		return
	}
	status := metav1.ConditionTrue
	if reason != egv1a1.EnvoyProxyReasonAccepted {
		status = metav1.ConditionFalse
	}

	cond := newCondition(string(egv1a1.EnvoyProxyConditionAccepted), status,
		string(reason), msg, ep.Generation)

	for _, item := range ep.Status.Ancestors {
		if ancestorRefsEqual(&item.AncestorRef, ancestor) {
			item.Conditions = MergeConditions(item.Conditions, cond)
			return
		}
	}

	// ancestor not found, append a new one
	ep.Status.Ancestors = append(ep.Status.Ancestors, egv1a1.EnvoyProxyAncestorStatus{
		AncestorRef: *ancestor,
		Conditions: []metav1.Condition{
			cond,
		},
	})
}
