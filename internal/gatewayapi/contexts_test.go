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
	r := &resource.Resources{
		GatewayClass: &gwapiv1.GatewayClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: "foo",
			},
		},
	}
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
	envoyproxyMap := map[types.NamespacedName]*egv1a1.EnvoyProxy{}
	gctx := &GatewayContext{
		Gateway: gateway,
	}
	gctx.ResetListeners(r, envoyproxyMap)
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

	gctx.ResetListeners(r, envoyproxyMap)
	require.Empty(t, gateway.Status.Listeners[0].Conditions)
}

func TestContextsStaleListener(t *testing.T) {
	r := &resource.Resources{
		GatewayClass: &gwapiv1.GatewayClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: "foo",
			},
		},
	}
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
	envoyproxyMap := map[types.NamespacedName]*egv1a1.EnvoyProxy{}
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

	gCtx.ResetListeners(r, envoyproxyMap)

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

	gCtx.ResetListeners(r, envoyproxyMap)

	// Ensure the listener status has been updated and the stale listener has been
	// removed.
	expectedListenerStatus := []gwapiv1.ListenerStatus{{Name: "https"}}
	require.Equal(t, expectedListenerStatus, gCtx.Status.Listeners)

	// Ensure that the listeners within GatewayContext have been properly updated.
	expectedGCtxListeners := []*ListenerContext{httpsListenerCtx}
	require.Equal(t, expectedGCtxListeners, gCtx.listeners)
}

func TestAttachEnvoyProxy(t *testing.T) {
	defaultGateway := &gwapiv1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "gateway-1",
		},
	}

	tests := []struct {
		name              string
		gateway           *gwapiv1.Gateway
		gatewayClassProxy *egv1a1.EnvoyProxy
		gatewayProxy      *egv1a1.EnvoyProxy
		template          *egv1a1.EnvoyProxyTemplateSpec
		expectedReplicas  *int32
	}{
		{
			name:              "no envoy proxy configs - should have nil envoyProxy",
			gateway:           defaultGateway,
			gatewayClassProxy: nil,
			gatewayProxy:      nil,
			template:          nil,
			expectedReplicas:  nil,
		},
		{
			name:              "template only",
			gateway:           defaultGateway,
			gatewayClassProxy: nil,
			gatewayProxy:      nil,
			template: &egv1a1.EnvoyProxyTemplateSpec{
				Spec: &egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.EnvoyProxyProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyDeployment: &egv1a1.KubernetesDeploymentSpec{
								Replicas: ptr.To(int32(2)),
							},
						},
					},
				},
			},
			expectedReplicas: ptr.To(int32(2)),
		},
		{
			name:    "template merged with gatewayClass",
			gateway: defaultGateway,
			gatewayClassProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.EnvoyProxyProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyDeployment: &egv1a1.KubernetesDeploymentSpec{
								Replicas: ptr.To(int32(3)),
							},
						},
					},
				},
			},
			gatewayProxy: nil,
			template: &egv1a1.EnvoyProxyTemplateSpec{
				Spec: &egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.EnvoyProxyProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyDeployment: &egv1a1.KubernetesDeploymentSpec{
								Replicas: ptr.To(int32(2)),
							},
						},
					},
				},
			},
			expectedReplicas: ptr.To(int32(3)), // GatewayClass overrides template with Replace strategy
		},
		{
			name: "gateway overrides gatewayClass and template",
			gateway: &gwapiv1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "gateway-1",
				},
				Spec: gwapiv1.GatewaySpec{
					Infrastructure: &gwapiv1.GatewayInfrastructure{
						ParametersRef: &gwapiv1.LocalParametersReference{
							Group: gwapiv1.Group(egv1a1.GroupVersion.Group),
							Kind:  gwapiv1.Kind(egv1a1.KindEnvoyProxy),
							Name:  "gateway-proxy",
						},
					},
				},
			},
			gatewayClassProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.EnvoyProxyProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyDeployment: &egv1a1.KubernetesDeploymentSpec{
								Replicas: ptr.To(int32(3)),
							},
						},
					},
				},
			},
			gatewayProxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "gateway-proxy",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.EnvoyProxyProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyDeployment: &egv1a1.KubernetesDeploymentSpec{
								Replicas: ptr.To(int32(5)),
							},
						},
					},
				},
			},
			template: &egv1a1.EnvoyProxyTemplateSpec{
				Spec: &egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.EnvoyProxyProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyDeployment: &egv1a1.KubernetesDeploymentSpec{
								Replicas: ptr.To(int32(2)),
							},
						},
					},
				},
			},
			expectedReplicas: ptr.To(int32(5)), // Gateway has highest priority with Replace strategy
		},
		{
			name: "gatewayclass used with MergeGateways enabled",
			gateway: &gwapiv1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "gateway-1",
				},
				Spec: gwapiv1.GatewaySpec{
					Infrastructure: &gwapiv1.GatewayInfrastructure{
						ParametersRef: &gwapiv1.LocalParametersReference{
							Group: gwapiv1.Group(egv1a1.GroupVersion.Group),
							Kind:  gwapiv1.Kind(egv1a1.KindEnvoyProxy),
							Name:  "gateway-proxy",
						},
					},
				},
			},
			gatewayClassProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					MergeGateways: ptr.To(true),
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.EnvoyProxyProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyDeployment: &egv1a1.KubernetesDeploymentSpec{
								Replicas: ptr.To(int32(3)),
							},
						},
					},
				},
			},
			gatewayProxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "gateway-proxy",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.EnvoyProxyProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyDeployment: &egv1a1.KubernetesDeploymentSpec{
								Replicas: ptr.To(int32(5)),
							},
						},
					},
				},
			},
			expectedReplicas: ptr.To(int32(3)), // Should use GatewayClass (3), NOT Gateway (5)
		},
		{
			name: "gatewayclass used with MergeGateways enabled on template",
			gateway: &gwapiv1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "gateway-1",
				},
				Spec: gwapiv1.GatewaySpec{
					Infrastructure: &gwapiv1.GatewayInfrastructure{
						ParametersRef: &gwapiv1.LocalParametersReference{
							Group: gwapiv1.Group(egv1a1.GroupVersion.Group),
							Kind:  gwapiv1.Kind(egv1a1.KindEnvoyProxy),
							Name:  "gateway-proxy",
						},
					},
				},
			},
			template: &egv1a1.EnvoyProxyTemplateSpec{
				Spec: &egv1a1.EnvoyProxySpec{
					MergeGateways: ptr.To(true),
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.EnvoyProxyProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyDeployment: &egv1a1.KubernetesDeploymentSpec{
								Replicas: ptr.To(int32(2)),
							},
						},
					},
				},
			},
			gatewayClassProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.EnvoyProxyProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyDeployment: &egv1a1.KubernetesDeploymentSpec{
								Replicas: ptr.To(int32(3)),
							},
						},
					},
				},
			},
			gatewayProxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "gateway-proxy",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.EnvoyProxyProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyDeployment: &egv1a1.KubernetesDeploymentSpec{
								Replicas: ptr.To(int32(5)),
							},
						},
					},
				},
			},
			expectedReplicas: ptr.To(int32(3)), // Should use GatewayClass (3), NOT Gateway (5)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup resources
			resources := &resource.Resources{
				EnvoyProxyForGatewayClass: tt.gatewayClassProxy,
			}

			epMap := map[types.NamespacedName]*egv1a1.EnvoyProxy{}

			if tt.gatewayProxy != nil {
				// Setup envoyproxy map
				epMap[types.NamespacedName{
					Namespace: tt.gatewayProxy.Namespace,
					Name:      tt.gatewayProxy.Name,
				}] = tt.gatewayProxy
			}

			// Create gateway context
			gctx := &GatewayContext{
				Gateway: tt.gateway,
			}

			// Call attachEnvoyProxy
			gctx.attachEnvoyProxy(resources, epMap, tt.template)

			// Verify the envoyProxy field was set correctly
			if tt.expectedReplicas == nil {
				require.Nil(t, gctx.envoyProxy)
			} else {
				require.NotNil(t, gctx.envoyProxy)
				require.NotNil(t, gctx.envoyProxy.Spec.Provider)
				require.NotNil(t, gctx.envoyProxy.Spec.Provider.Kubernetes)
				require.NotNil(t, gctx.envoyProxy.Spec.Provider.Kubernetes.EnvoyDeployment)
				require.Equal(t, *tt.expectedReplicas, *gctx.envoyProxy.Spec.Provider.Kubernetes.EnvoyDeployment.Replicas)
			}
		})
	}
}
