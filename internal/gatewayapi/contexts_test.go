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

func TestContexts(t *testing.T) {
	gateway := &gwapiv1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "envoy-gateway",
			Name:      "gateway-1",
		},
		Spec: gwapiv1.GatewaySpec{
			Listeners: []gwapiv1.Listener{
				{
					Name: "http",
				},
			},
		},
	}

	gctx := &GatewayContext{
		Gateway: gateway,
	}
	gctx.ResetListeners()
	require.Len(t, gctx.listeners, 1)

	lctx := gctx.listeners[0]
	require.NotNil(t, lctx)

	lctx.SetCondition(gwapiv1.ListenerConditionAccepted, metav1.ConditionFalse, gwapiv1.ListenerReasonUnsupportedProtocol, "HTTPS protocol is not supported yet")

	require.Len(t, gateway.Status.Listeners, 1)
	require.EqualValues(t, "http", gateway.Status.Listeners[0].Name)
	require.Len(t, gateway.Status.Listeners[0].Conditions, 1)
	require.EqualValues(t, gwapiv1.ListenerConditionAccepted, gateway.Status.Listeners[0].Conditions[0].Type)
	require.EqualValues(t, metav1.ConditionFalse, gateway.Status.Listeners[0].Conditions[0].Status)
	require.EqualValues(t, gwapiv1.ListenerReasonUnsupportedProtocol, gateway.Status.Listeners[0].Conditions[0].Reason)
	require.EqualValues(t, "HTTPS protocol is not supported yet", gateway.Status.Listeners[0].Conditions[0].Message)

	lctx.SetSupportedKinds(gwapiv1.RouteGroupKind{Group: GroupPtr(gwapiv1.GroupName), Kind: "HTTPRoute"})

	require.Len(t, gateway.Status.Listeners, 1)
	require.Len(t, gateway.Status.Listeners[0].SupportedKinds, 1)
	require.EqualValues(t, "HTTPRoute", gateway.Status.Listeners[0].SupportedKinds[0].Kind)

	gctx.ResetListeners()
	require.Empty(t, gateway.Status.Listeners[0].Conditions)
}

func TestContextsStaleListener(t *testing.T) {
	gateway := &gwapiv1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "envoy-gateway",
			Name:      "gateway-1",
		},
		Spec: gwapiv1.GatewaySpec{
			Listeners: []gwapiv1.Listener{
				{
					Name: "https",
				},
				{
					Name: "http",
				},
			},
		},
		Status: gwapiv1.GatewayStatus{
			Listeners: []gwapiv1.ListenerStatus{
				{
					Name: "https",
					Conditions: []metav1.Condition{
						{
							Status: metav1.ConditionStatus(gwapiv1.ListenerConditionProgrammed),
						},
					},
				},
				{
					Name: "http",
					Conditions: []metav1.Condition{
						{
							Status: metav1.ConditionStatus(gwapiv1.ListenerConditionProgrammed),
						},
					},
				},
			},
		},
	}

	gCtx := &GatewayContext{Gateway: gateway}

	httpsListenerCtx := &ListenerContext{
		Listener: &gwapiv1.Listener{
			Name: "https",
		},
		gateway:           gateway,
		listenerStatusIdx: 0,
	}

	httpListenerCtx := &ListenerContext{
		Listener: &gwapiv1.Listener{
			Name: "http",
		},
		gateway:           gateway,
		listenerStatusIdx: 1,
	}

	gCtx.ResetListeners()

	require.Len(t, gCtx.listeners, 2)

	expectedListenerContexts := []*ListenerContext{
		httpsListenerCtx,
		httpListenerCtx,
	}
	require.EqualValues(t, expectedListenerContexts, gCtx.listeners)

	require.Len(t, gCtx.Status.Listeners, 2)

	expectedListenerStatuses := []gwapiv1.ListenerStatus{
		{
			Name: "https",
		},
		{
			Name: "http",
		},
	}
	require.EqualValues(t, expectedListenerStatuses, gCtx.Status.Listeners)

	// Remove one of the listeners
	gateway.Spec.Listeners = gateway.Spec.Listeners[:1]

	gCtx.ResetListeners()

	// Ensure the listener status has been updated and the stale listener has been
	// removed.
	expectedListenerStatus := []gwapiv1.ListenerStatus{{Name: "https"}}
	require.EqualValues(t, expectedListenerStatus, gCtx.Gateway.Status.Listeners)

	// Ensure that the listeners within GatewayContext have been properly updated.
	expectedGCtxListeners := []*ListenerContext{httpsListenerCtx}
	require.EqualValues(t, expectedGCtxListeners, gCtx.listeners)
}
