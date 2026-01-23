// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
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

	status.SetGatewayListenerStatusCondition(lctx.gateway.Gateway, lctx.listenerStatusIdx,
		gwapiv1.ListenerConditionAccepted, metav1.ConditionFalse, gwapiv1.ListenerReasonUnsupportedProtocol, "HTTPS protocol is not supported yet")

	require.Len(t, gateway.Status.Listeners, 1)
	require.EqualValues(t, "http", gateway.Status.Listeners[0].Name)
	require.Len(t, gateway.Status.Listeners[0].Conditions, 1)
	require.EqualValues(t, gwapiv1.ListenerConditionAccepted, gateway.Status.Listeners[0].Conditions[0].Type)
	require.Equal(t, metav1.ConditionFalse, gateway.Status.Listeners[0].Conditions[0].Status)
	require.EqualValues(t, gwapiv1.ListenerReasonUnsupportedProtocol, gateway.Status.Listeners[0].Conditions[0].Reason)
	require.Equal(t, "HTTPS protocol is not supported yet", gateway.Status.Listeners[0].Conditions[0].Message)

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
		gateway:           gCtx,
		listenerStatusIdx: 0,
	}

	httpListenerCtx := &ListenerContext{
		Listener: &gwapiv1.Listener{
			Name: "http",
		},
		gateway:           gCtx,
		listenerStatusIdx: 1,
	}

	gCtx.ResetListeners()

	require.Len(t, gCtx.listeners, 2)

	expectedListenerContexts := []*ListenerContext{
		httpsListenerCtx,
		httpListenerCtx,
	}
	require.Equal(t, expectedListenerContexts, gCtx.listeners)

	require.Len(t, gCtx.Status.Listeners, 2)

	expectedListenerStatuses := []gwapiv1.ListenerStatus{
		{
			Name: "https",
		},
		{
			Name: "http",
		},
	}
	require.Equal(t, expectedListenerStatuses, gCtx.Status.Listeners)

	// Remove one of the listeners
	gateway.Spec.Listeners = gateway.Spec.Listeners[:1]

	gCtx.ResetListeners()

	// Ensure the listener status has been updated and the stale listener has been
	// removed.
	expectedListenerStatus := []gwapiv1.ListenerStatus{{Name: "https"}}
	require.Equal(t, expectedListenerStatus, gCtx.Status.Listeners)

	// Ensure that the listeners within GatewayContext have been properly updated.
	expectedGCtxListeners := []*ListenerContext{httpsListenerCtx}
	require.Equal(t, expectedGCtxListeners, gCtx.listeners)
}

func TestAttachEnvoyProxy(t *testing.T) {
	testCases := []struct {
		name                  string
		gatewayParametersRef  *gwapiv1.LocalParametersReference
		envoyProxyForGateway  *egv1a1.EnvoyProxy
		envoyProxyForGWClass  *egv1a1.EnvoyProxy
		envoyProxyDefaultSpec *egv1a1.EnvoyProxySpec
		expectedMergeGateways *bool
		expectedConcurrency   *int32
		expectEnvoyProxyNil   bool
	}{
		{
			name:                "no envoy proxy at any level",
			expectEnvoyProxyNil: true,
		},
		{
			name: "only default spec - should use default",
			envoyProxyDefaultSpec: &egv1a1.EnvoyProxySpec{
				Concurrency: ptr.To[int32](4),
			},
			expectedConcurrency: ptr.To[int32](4),
		},
		{
			name: "gatewayclass envoy proxy overrides default spec",
			envoyProxyForGWClass: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "envoy-gateway-system",
					Name:      "gc-proxy",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Concurrency: ptr.To[int32](8),
				},
			},
			envoyProxyDefaultSpec: &egv1a1.EnvoyProxySpec{
				Concurrency: ptr.To[int32](4),
			},
			expectedConcurrency: ptr.To[int32](8),
		},
		{
			name: "gateway envoy proxy overrides gatewayclass",
			gatewayParametersRef: &gwapiv1.LocalParametersReference{
				Group: gwapiv1.Group(egv1a1.GroupVersion.Group),
				Kind:  gwapiv1.Kind(egv1a1.KindEnvoyProxy),
				Name:  "gw-proxy",
			},
			envoyProxyForGateway: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "gw-proxy",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Concurrency: ptr.To[int32](16),
				},
			},
			envoyProxyForGWClass: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "envoy-gateway-system",
					Name:      "gc-proxy",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Concurrency: ptr.To[int32](8),
				},
			},
			envoyProxyDefaultSpec: &egv1a1.EnvoyProxySpec{
				Concurrency: ptr.To[int32](4),
			},
			expectedConcurrency: ptr.To[int32](16),
		},
		{
			name: "default spec with merge gateways enabled",
			envoyProxyDefaultSpec: &egv1a1.EnvoyProxySpec{
				MergeGateways: ptr.To(true),
				Concurrency:   ptr.To[int32](4),
			},
			expectedMergeGateways: ptr.To(true),
			expectedConcurrency:   ptr.To[int32](4),
		},
		{
			name: "gatewayclass overrides default merge gateways setting",
			envoyProxyForGWClass: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "envoy-gateway-system",
					Name:      "gc-proxy",
				},
				Spec: egv1a1.EnvoyProxySpec{
					MergeGateways: ptr.To(false),
				},
			},
			envoyProxyDefaultSpec: &egv1a1.EnvoyProxySpec{
				MergeGateways: ptr.To(true),
			},
			expectedMergeGateways: ptr.To(false),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create gateway
			gateway := &gwapiv1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "test-gateway",
				},
				Spec: gwapiv1.GatewaySpec{
					GatewayClassName: "test-gc",
				},
			}
			if tc.gatewayParametersRef != nil {
				gateway.Spec.Infrastructure = &gwapiv1.GatewayInfrastructure{
					ParametersRef: tc.gatewayParametersRef,
				}
			}

			gCtx := &GatewayContext{Gateway: gateway}

			// Build resources
			resources := &resource.Resources{
				EnvoyProxyForGatewayClass: tc.envoyProxyForGWClass,
				EnvoyProxyDefaultSpec:     tc.envoyProxyDefaultSpec,
			}

			// Build envoy proxy map for gateway-level proxies
			epMap := make(map[types.NamespacedName]*egv1a1.EnvoyProxy)
			if tc.envoyProxyForGateway != nil {
				key := types.NamespacedName{
					Namespace: tc.envoyProxyForGateway.Namespace,
					Name:      tc.envoyProxyForGateway.Name,
				}
				epMap[key] = tc.envoyProxyForGateway
			}

			// Call attachEnvoyProxy
			gCtx.attachEnvoyProxy(resources, epMap)

			// Verify results
			if tc.expectEnvoyProxyNil {
				require.Nil(t, gCtx.envoyProxy)
				return
			}

			require.NotNil(t, gCtx.envoyProxy)

			if tc.expectedConcurrency != nil {
				require.NotNil(t, gCtx.envoyProxy.Spec.Concurrency)
				require.Equal(t, *tc.expectedConcurrency, *gCtx.envoyProxy.Spec.Concurrency)
			}

			if tc.expectedMergeGateways != nil {
				require.NotNil(t, gCtx.envoyProxy.Spec.MergeGateways)
				require.Equal(t, *tc.expectedMergeGateways, *gCtx.envoyProxy.Spec.MergeGateways)
			}
		})
	}
}

func TestIsMergeGatewaysEnabled(t *testing.T) {
	testCases := []struct {
		name                  string
		envoyProxyForGWClass  *egv1a1.EnvoyProxy
		envoyProxyDefaultSpec *egv1a1.EnvoyProxySpec
		expected              bool
	}{
		{
			name:     "no envoy proxy configured",
			expected: false,
		},
		{
			name: "default spec with merge gateways true",
			envoyProxyDefaultSpec: &egv1a1.EnvoyProxySpec{
				MergeGateways: ptr.To(true),
			},
			expected: true,
		},
		{
			name: "default spec with merge gateways false",
			envoyProxyDefaultSpec: &egv1a1.EnvoyProxySpec{
				MergeGateways: ptr.To(false),
			},
			expected: false,
		},
		{
			name: "gatewayclass proxy with merge gateways true",
			envoyProxyForGWClass: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					MergeGateways: ptr.To(true),
				},
			},
			expected: true,
		},
		{
			name: "gatewayclass proxy overrides default - gc true, default false",
			envoyProxyForGWClass: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					MergeGateways: ptr.To(true),
				},
			},
			envoyProxyDefaultSpec: &egv1a1.EnvoyProxySpec{
				MergeGateways: ptr.To(false),
			},
			expected: true,
		},
		{
			name: "gatewayclass proxy overrides default - gc false, default true",
			envoyProxyForGWClass: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					MergeGateways: ptr.To(false),
				},
			},
			envoyProxyDefaultSpec: &egv1a1.EnvoyProxySpec{
				MergeGateways: ptr.To(true),
			},
			expected: false,
		},
		{
			name: "gatewayclass proxy nil merge gateways falls back to default",
			envoyProxyForGWClass: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					Concurrency: ptr.To[int32](4), // some other setting
				},
			},
			envoyProxyDefaultSpec: &egv1a1.EnvoyProxySpec{
				MergeGateways: ptr.To(true),
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resources := &resource.Resources{
				EnvoyProxyForGatewayClass: tc.envoyProxyForGWClass,
				EnvoyProxyDefaultSpec:     tc.envoyProxyDefaultSpec,
			}

			result := IsMergeGatewaysEnabled(resources)
			require.Equal(t, tc.expected, result)
		})
	}
}
