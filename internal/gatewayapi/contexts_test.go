// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
)

func TestContexts(t *testing.T) {
	gateway := &v1beta1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "envoy-gateway",
			Name:      "gateway-1",
		},
		Spec: v1beta1.GatewaySpec{
			Listeners: []v1beta1.Listener{
				{
					Name: "http",
				},
			},
		},
	}

	gctx := &GatewayContext{
		Gateway: gateway,
	}

	lctx := gctx.GetListenerContext("http")
	require.NotNil(t, lctx)

	lctx.SetCondition(v1beta1.ListenerConditionDetached, metav1.ConditionTrue, v1beta1.ListenerReasonUnsupportedProtocol, "HTTPS protocol is not supported yet")

	require.Len(t, gateway.Status.Listeners, 1)
	require.EqualValues(t, gateway.Status.Listeners[0].Name, "http")
	require.Len(t, gateway.Status.Listeners[0].Conditions, 1)
	require.EqualValues(t, gateway.Status.Listeners[0].Conditions[0].Type, v1beta1.ListenerConditionDetached)
	require.EqualValues(t, gateway.Status.Listeners[0].Conditions[0].Status, metav1.ConditionTrue)
	require.EqualValues(t, gateway.Status.Listeners[0].Conditions[0].Reason, v1beta1.ListenerReasonUnsupportedProtocol)
	require.EqualValues(t, gateway.Status.Listeners[0].Conditions[0].Message, "HTTPS protocol is not supported yet")

	lctx.SetSupportedKinds(v1beta1.RouteGroupKind{Group: GroupPtr(v1beta1.GroupName), Kind: "HTTPRoute"})

	require.Len(t, gateway.Status.Listeners, 1)
	require.Len(t, gateway.Status.Listeners[0].SupportedKinds, 1)
	require.EqualValues(t, gateway.Status.Listeners[0].SupportedKinds[0].Kind, "HTTPRoute")

	lctx.ResetConditions()
	require.Len(t, gateway.Status.Listeners[0].Conditions, 0)
}
