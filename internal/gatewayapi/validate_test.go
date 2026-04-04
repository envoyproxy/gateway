// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestValidateConflictedProtocolsListenersIgnoresUnsupportedProtocols(t *testing.T) {
	unsupported := gwapiv1.ProtocolType("INVALID")

	gateway := &gwapiv1.Gateway{}
	gateway.Status.Listeners = []gwapiv1.ListenerStatus{{}, {}}

	gatewayCtx := &GatewayContext{Gateway: gateway}
	gatewayCtx.listeners = []*ListenerContext{
		{
			Listener: &gwapiv1.Listener{
				Name:     "invalid",
				Port:     80,
				Protocol: unsupported,
			},
			gateway:           gatewayCtx,
			listenerStatusIdx: 0,
		},
		{
			Listener: &gwapiv1.Listener{
				Name:     "http",
				Port:     80,
				Protocol: gwapiv1.HTTPProtocolType,
			},
			gateway:           gatewayCtx,
			listenerStatusIdx: 1,
		},
	}

	translator := &Translator{}
	translator.validateConflictedProtocolsListeners([]*GatewayContext{gatewayCtx})

	httpConds := gatewayCtx.Status.Listeners[1].Conditions
	require.False(t, hasListenerCondition(httpConds, gwapiv1.ListenerConditionConflicted, gwapiv1.ListenerReasonProtocolConflict, metav1.ConditionTrue))
	require.False(t, hasListenerCondition(httpConds, gwapiv1.ListenerConditionAccepted, gwapiv1.ListenerReasonProtocolConflict, metav1.ConditionFalse))
	require.False(t, hasListenerCondition(httpConds, gwapiv1.ListenerConditionProgrammed, gwapiv1.ListenerReasonProtocolConflict, metav1.ConditionFalse))
}

func hasListenerCondition(conditions []metav1.Condition, condType gwapiv1.ListenerConditionType, reason gwapiv1.ListenerConditionReason, status metav1.ConditionStatus) bool {
	for _, cond := range conditions {
		if cond.Type == string(condType) && cond.Reason == string(reason) && cond.Status == status {
			return true
		}
	}

	return false
}
