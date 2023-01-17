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

	lctx.SetCondition(v1beta1.ListenerConditionAccepted, metav1.ConditionFalse, v1beta1.ListenerReasonUnsupportedProtocol, "HTTPS protocol is not supported yet")

	require.Len(t, gateway.Status.Listeners, 1)
	require.EqualValues(t, gateway.Status.Listeners[0].Name, "http")
	require.Len(t, gateway.Status.Listeners[0].Conditions, 1)
	require.EqualValues(t, gateway.Status.Listeners[0].Conditions[0].Type, v1beta1.ListenerConditionAccepted)
	require.EqualValues(t, gateway.Status.Listeners[0].Conditions[0].Status, metav1.ConditionFalse)
	require.EqualValues(t, gateway.Status.Listeners[0].Conditions[0].Reason, v1beta1.ListenerReasonUnsupportedProtocol)
	require.EqualValues(t, gateway.Status.Listeners[0].Conditions[0].Message, "HTTPS protocol is not supported yet")

	lctx.SetSupportedKinds(v1beta1.RouteGroupKind{Group: GroupPtr(v1beta1.GroupName), Kind: "HTTPRoute"})

	require.Len(t, gateway.Status.Listeners, 1)
	require.Len(t, gateway.Status.Listeners[0].SupportedKinds, 1)
	require.EqualValues(t, gateway.Status.Listeners[0].SupportedKinds[0].Kind, "HTTPRoute")

	gctx.ResetListeners()
	require.Len(t, gateway.Status.Listeners[0].Conditions, 0)
}

func TestContextsStaleListener(t *testing.T) {
	gateway := &v1beta1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "envoy-gateway",
			Name:      "gateway-1",
		},
		Spec: v1beta1.GatewaySpec{
			Listeners: []v1beta1.Listener{
				{
					Name: "https",
				},
				{
					Name: "http",
				},
			},
		},
		Status: v1beta1.GatewayStatus{
			Listeners: []v1beta1.ListenerStatus{
				{
					Name: "https",
					Conditions: []metav1.Condition{
						{
							Status: metav1.ConditionStatus(v1beta1.ListenerConditionProgrammed),
						},
					},
				},
				{
					Name: "http",
					Conditions: []metav1.Condition{
						{
							Status: metav1.ConditionStatus(v1beta1.ListenerConditionProgrammed),
						},
					},
				},
			},
		},
	}

	gCtx := &GatewayContext{Gateway: gateway}

	httpsListenerCtx := &ListenerContext{
		Listener: &v1beta1.Listener{
			Name: "https",
		},
		gateway:           gateway,
		listenerStatusIdx: 0,
	}

	httpListenerCtx := &ListenerContext{
		Listener: &v1beta1.Listener{
			Name: "http",
		},
		gateway:           gateway,
		listenerStatusIdx: 1,
	}

	gCtx.ResetListeners()

	require.EqualValues(t, 2, len(gCtx.listeners))

	expectedListenerContexts := []*ListenerContext{
		httpsListenerCtx,
		httpListenerCtx,
	}
	require.EqualValues(t, expectedListenerContexts, gCtx.listeners)

	require.EqualValues(t, 2, len(gCtx.Status.Listeners))

	expectedListenerStatuses := []v1beta1.ListenerStatus{
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
	expectedListenerStatus := []v1beta1.ListenerStatus{{Name: "https"}}
	require.EqualValues(t, expectedListenerStatus, gCtx.Gateway.Status.Listeners)

	// Ensure that the listeners within GatewayContext have been properly updated.
	expectedGCtxListeners := []*ListenerContext{httpsListenerCtx}
	require.EqualValues(t, expectedGCtxListeners, gCtx.listeners)
}
