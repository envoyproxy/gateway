// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/utils"
)

func TestProcessHTTPRoutes(t *testing.T) {
	const (
		defaultWait = time.Second * 10
		defaultTick = time.Millisecond * 20
	)

	// The gatewayclass configured for the reconciler and referenced by test cases.
	gcCtrlName := gwapiv1.GatewayController(egv1a1.GatewayControllerName)
	gc := &gwapiv1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: gwapiv1.GatewayClassSpec{
			ControllerName: gcCtrlName,
		},
	}

	// The gateway referenced by test cases.
	gw := &gwapiv1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "test",
		},
		Spec: gwapiv1.GatewaySpec{
			GatewayClassName: gwapiv1.ObjectName(gc.Name),
			Listeners: []gwapiv1.Listener{
				{
					Name:     "http",
					Protocol: gwapiv1.HTTPProtocolType,
					Port:     gwapiv1.PortNumber(int32(8080)),
				},
			},
		},
	}
	gwNsName := utils.NamespacedName(gw).String()

	invalidDuration := gwapiv1.Duration("invalid duration")

	httpRouteNS := "test"

	testCases := []struct {
		name               string
		routes             []*gwapiv1.HTTPRoute
		extensionFilters   []*unstructured.Unstructured
		extensionAPIGroups []schema.GroupVersionKind
		expected           bool
	}{
		{
			name: "valid httproute",
			routes: []*gwapiv1.HTTPRoute{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: httpRouteNS,
						Name:      "test",
					},
					Spec: gwapiv1.HTTPRouteSpec{
						CommonRouteSpec: gwapiv1.CommonRouteSpec{
							ParentRefs: []gwapiv1.ParentReference{
								{
									Name: "test",
								},
							},
						},
						Rules: []gwapiv1.HTTPRouteRule{
							{
								Matches: []gwapiv1.HTTPRouteMatch{
									{
										Path: &gwapiv1.HTTPPathMatch{
											Type:  ptr.To(gwapiv1.PathMatchPathPrefix),
											Value: ptr.To("/"),
										},
									},
								},
								BackendRefs: []gwapiv1.HTTPBackendRef{
									{
										BackendRef: gwapiv1.BackendRef{
											BackendObjectReference: gwapiv1.BackendObjectReference{
												Group: gatewayapi.GroupPtr(corev1.GroupName),
												Kind:  gatewayapi.KindPtr(gatewayapi.KindService),
												Name:  "test",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "httproute with one filter_from_extension",
			routes: []*gwapiv1.HTTPRoute{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: httpRouteNS,
						Name:      "test",
					},
					Spec: gwapiv1.HTTPRouteSpec{
						CommonRouteSpec: gwapiv1.CommonRouteSpec{
							ParentRefs: []gwapiv1.ParentReference{
								{
									Name: "test",
								},
							},
						},
						Rules: []gwapiv1.HTTPRouteRule{
							{
								Matches: []gwapiv1.HTTPRouteMatch{
									{
										Path: &gwapiv1.HTTPPathMatch{
											Type:  ptr.To(gwapiv1.PathMatchPathPrefix),
											Value: ptr.To("/"),
										},
									},
								},
								Filters: []gwapiv1.HTTPRouteFilter{
									{
										Type: gwapiv1.HTTPRouteFilterExtensionRef,
										ExtensionRef: &gwapiv1.LocalObjectReference{
											Group: gwapiv1.Group("gateway.example.io"),
											Kind:  gwapiv1.Kind("Foo"),
											Name:  gwapiv1.ObjectName("test"),
										},
									},
								},
								BackendRefs: []gwapiv1.HTTPBackendRef{
									{
										BackendRef: gwapiv1.BackendRef{
											BackendObjectReference: gwapiv1.BackendObjectReference{
												Group: gatewayapi.GroupPtr(corev1.GroupName),
												Kind:  gatewayapi.KindPtr(gatewayapi.KindService),
												Name:  "test",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			extensionFilters: []*unstructured.Unstructured{
				{
					Object: map[string]interface{}{
						"apiVersion": "gateway.example.io/v1alpha1",
						"kind":       "Foo",
						"metadata": map[string]interface{}{
							"name":      "test",
							"namespace": httpRouteNS,
						},
					},
				},
			},
			extensionAPIGroups: []schema.GroupVersionKind{
				{
					Group:   "gateway.example.io",
					Version: "v1alpha1",
					Kind:    "Foo",
				},
			},
			expected: true,
		},
		{
			name: "httproute with invalid timeout setting for HTTPRouteRule",
			routes: []*gwapiv1.HTTPRoute{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: httpRouteNS,
						Name:      "test",
					},
					Spec: gwapiv1.HTTPRouteSpec{
						CommonRouteSpec: gwapiv1.CommonRouteSpec{
							ParentRefs: []gwapiv1.ParentReference{
								{
									Name: "test",
								},
							},
						},
						Rules: []gwapiv1.HTTPRouteRule{
							{
								Matches: []gwapiv1.HTTPRouteMatch{
									{
										Path: &gwapiv1.HTTPPathMatch{
											Type:  ptr.To(gwapiv1.PathMatchPathPrefix),
											Value: ptr.To("/"),
										},
									},
								},
								BackendRefs: []gwapiv1.HTTPBackendRef{
									{
										BackendRef: gwapiv1.BackendRef{
											BackendObjectReference: gwapiv1.BackendObjectReference{
												Group: gatewayapi.GroupPtr(corev1.GroupName),
												Kind:  gatewayapi.KindPtr(gatewayapi.KindService),
												Name:  "test",
											},
										},
									},
								},
								Timeouts: &gwapiv1.HTTPRouteTimeouts{
									Request:        &invalidDuration,
									BackendRequest: &invalidDuration,
								},
							},
						},
					},
				},
			},
			expected: true,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		// Run the test cases.
		t.Run(tc.name, func(t *testing.T) {
			// Add objects referenced by test cases.
			objs := []client.Object{gc, gw}

			// Create the reconciler.
			logger := logging.DefaultLogger(egv1a1.LogLevelInfo)

			ctx := context.Background()

			r := &gatewayAPIReconciler{
				log:             logger,
				classController: gcCtrlName,
			}

			// Add the test case objects to the reconciler client.
			for _, route := range tc.routes {
				objs = append(objs, route)
			}
			for _, filter := range tc.extensionFilters {
				objs = append(objs, filter)
			}
			if len(tc.extensionAPIGroups) > 0 {
				r.extGVKs = append(r.extGVKs, tc.extensionAPIGroups...)
			}
			r.client = fakeclient.NewClientBuilder().
				WithScheme(envoygateway.GetScheme()).
				WithObjects(objs...).
				WithIndex(&gwapiv1.HTTPRoute{}, gatewayHTTPRouteIndex, gatewayHTTPRouteIndexFunc).
				Build()

			// Wait until all the httproutes have been initialized.
			require.Eventually(t, func() bool {
				httpRoutes := gwapiv1.HTTPRouteList{}
				if err := r.client.List(ctx, &httpRoutes, client.InNamespace(httpRouteNS)); err != nil {
					return false
				}
				return len(httpRoutes.Items) > 0
			}, defaultWait, defaultTick)

			// Process the test case httproutes.
			resourceTree := gatewayapi.NewResources()
			resourceMap := newResourceMapping()
			err := r.processHTTPRoutes(ctx, gwNsName, resourceMap, resourceTree)
			if tc.expected {
				require.NoError(t, err)
				// Ensure the resource tree and map are as expected.
				require.Equal(t, tc.routes, resourceTree.HTTPRoutes)
				if tc.extensionFilters != nil {
					for i, filter := range tc.extensionFilters {
						key := types.NamespacedName{
							Namespace: tc.routes[i].Namespace,
							Name:      filter.GetName(),
						}
						require.Equal(t, *filter, resourceMap.extensionRefFilters[key])
					}
				}
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestProcessGRPCRoutes(t *testing.T) {
	// The gatewayclass configured for the reconciler and referenced by test cases.
	gcCtrlName := gwapiv1.GatewayController(egv1a1.GatewayControllerName)
	gc := &gwapiv1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: gwapiv1.GatewayClassSpec{
			ControllerName: gcCtrlName,
		},
	}

	// The gateway referenced by test cases.
	gw := &gwapiv1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "test",
		},
		Spec: gwapiv1.GatewaySpec{
			GatewayClassName: gwapiv1.ObjectName(gc.Name),
			Listeners: []gwapiv1.Listener{
				{
					Name:     "http",
					Protocol: gwapiv1.HTTPProtocolType,
					Port:     gwapiv1.PortNumber(int32(8080)),
				},
			},
		},
	}
	gwNsName := utils.NamespacedName(gw).String()

	testCases := []struct {
		name               string
		routes             []*gwapiv1a2.GRPCRoute
		extensionAPIGroups []schema.GroupVersionKind
		expected           bool
	}{
		{
			name: "valid grpcroute",
			routes: []*gwapiv1a2.GRPCRoute{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test",
						Name:      "test",
					},
					Spec: gwapiv1a2.GRPCRouteSpec{
						CommonRouteSpec: gwapiv1.CommonRouteSpec{
							ParentRefs: []gwapiv1.ParentReference{
								{
									Name: "test",
								},
							},
						},
						Rules: []gwapiv1a2.GRPCRouteRule{
							{
								Matches: []gwapiv1a2.GRPCRouteMatch{
									{
										Method: &gwapiv1a2.GRPCMethodMatch{
											Method: ptr.To("Ping"),
										},
									},
								},
								BackendRefs: []gwapiv1a2.GRPCBackendRef{
									{
										BackendRef: gwapiv1.BackendRef{
											BackendObjectReference: gwapiv1.BackendObjectReference{
												Group: gatewayapi.GroupPtr(corev1.GroupName),
												Kind:  gatewayapi.KindPtr(gatewayapi.KindService),
												Name:  "test",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expected: true,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		// Run the test cases.
		t.Run(tc.name, func(t *testing.T) {
			// Add objects referenced by test cases.
			objs := []client.Object{gc, gw}

			// Create the reconciler.
			logger := logging.DefaultLogger(egv1a1.LogLevelInfo)

			ctx := context.Background()

			r := &gatewayAPIReconciler{
				log:             logger,
				classController: gcCtrlName,
			}

			// Add the test case objects to the reconciler client.
			for _, route := range tc.routes {
				objs = append(objs, route)
			}
			if len(tc.extensionAPIGroups) > 0 {
				r.extGVKs = append(r.extGVKs, tc.extensionAPIGroups...)
			}
			r.client = fakeclient.NewClientBuilder().
				WithScheme(envoygateway.GetScheme()).
				WithObjects(objs...).
				WithIndex(&gwapiv1a2.GRPCRoute{}, gatewayGRPCRouteIndex, gatewayGRPCRouteIndexFunc).
				Build()

			// Process the test case httproutes.
			resourceTree := gatewayapi.NewResources()
			resourceMap := newResourceMapping()
			err := r.processGRPCRoutes(ctx, gwNsName, resourceMap, resourceTree)
			if tc.expected {
				require.NoError(t, err)
				// Ensure the resource tree and map are as expected.
				require.Equal(t, tc.routes, resourceTree.GRPCRoutes)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestValidateHTTPRouteParentRefs(t *testing.T) {
	testCases := []struct {
		name     string
		route    *gwapiv1.HTTPRoute
		gateways []*gwapiv1.Gateway
		classes  []*gwapiv1.GatewayClass
		expect   []gwapiv1.Gateway
		expected bool
	}{
		{
			name: "valid parentRef",
			route: &gwapiv1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: gwapiv1.HTTPRouteSpec{
					CommonRouteSpec: gwapiv1.CommonRouteSpec{
						ParentRefs: []gwapiv1.ParentReference{
							{
								Group: gatewayapi.GroupPtr(gwapiv1.GroupName),
								Kind:  gatewayapi.KindPtr("Gateway"),
								Name:  "test",
							},
						},
					},
				},
			},
			gateways: []*gwapiv1.Gateway{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test",
						Name:      "test",
					},
					Spec: gwapiv1.GatewaySpec{
						GatewayClassName: "gc1",
					},
				},
			},
			classes: []*gwapiv1.GatewayClass{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "gc1",
					},
					Spec: gwapiv1.GatewayClassSpec{
						ControllerName: gwapiv1.GatewayController(egv1a1.GatewayControllerName),
					},
				},
			},
			expect: []gwapiv1.Gateway{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:       "test",
						Name:            "test",
						ResourceVersion: "999",
					},
					Spec: gwapiv1.GatewaySpec{
						GatewayClassName: "gc1",
					},
				},
			},
			expected: true,
		},
		{
			name: "invalid parentRef group",
			route: &gwapiv1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: gwapiv1.HTTPRouteSpec{
					CommonRouteSpec: gwapiv1.CommonRouteSpec{
						ParentRefs: []gwapiv1.ParentReference{
							{
								Group: gatewayapi.GroupPtr("unsupported.group"),
								Kind:  gatewayapi.KindPtr("Gateway"),
								Name:  "test",
							},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "invalid parentRef kind",
			route: &gwapiv1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: gwapiv1.HTTPRouteSpec{
					CommonRouteSpec: gwapiv1.CommonRouteSpec{
						ParentRefs: []gwapiv1.ParentReference{
							{
								Group: gatewayapi.GroupPtr(gwapiv1.GroupName),
								Kind:  gatewayapi.KindPtr("UnsupportedKind"),
								Name:  "test",
							},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "non-existent parentRef name",
			route: &gwapiv1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: gwapiv1.HTTPRouteSpec{
					CommonRouteSpec: gwapiv1.CommonRouteSpec{
						ParentRefs: []gwapiv1.ParentReference{
							{
								Group: gatewayapi.GroupPtr(gwapiv1.GroupName),
								Kind:  gatewayapi.KindPtr("Gateway"),
								Name:  "no-existent",
							},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "valid parentRefs",
			route: &gwapiv1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: gwapiv1.HTTPRouteSpec{
					CommonRouteSpec: gwapiv1.CommonRouteSpec{
						ParentRefs: []gwapiv1.ParentReference{
							{
								Group: gatewayapi.GroupPtr(gwapiv1.GroupName),
								Kind:  gatewayapi.KindPtr("Gateway"),
								Name:  "test",
							},
							{
								Group: gatewayapi.GroupPtr(gwapiv1.GroupName),
								Kind:  gatewayapi.KindPtr("Gateway"),
								Name:  "test2",
							},
						},
					},
				},
			},
			gateways: []*gwapiv1.Gateway{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test",
						Name:      "test",
					},
					Spec: gwapiv1.GatewaySpec{
						GatewayClassName: "gc1",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test",
						Name:      "test2",
					},
					Spec: gwapiv1.GatewaySpec{
						GatewayClassName: "gc1",
					},
				},
			},
			classes: []*gwapiv1.GatewayClass{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "gc1",
					},
					Spec: gwapiv1.GatewayClassSpec{
						ControllerName: gwapiv1.GatewayController(egv1a1.GatewayControllerName),
					},
				},
			},
			expect: []gwapiv1.Gateway{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:       "test",
						Name:            "test",
						ResourceVersion: "999",
					},
					Spec: gwapiv1.GatewaySpec{
						GatewayClassName: "gc1",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:       "test",
						Name:            "test2",
						ResourceVersion: "999",
					},
					Spec: gwapiv1.GatewaySpec{
						GatewayClassName: "gc1",
					},
				},
			},
			expected: true,
		},
		{
			name: "one of two parentRefs are managed",
			route: &gwapiv1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: gwapiv1.HTTPRouteSpec{
					CommonRouteSpec: gwapiv1.CommonRouteSpec{
						ParentRefs: []gwapiv1.ParentReference{
							{
								Group: gatewayapi.GroupPtr(gwapiv1.GroupName),
								Kind:  gatewayapi.KindPtr("Gateway"),
								Name:  "test",
							},
							{
								Group: gatewayapi.GroupPtr(gwapiv1.GroupName),
								Kind:  gatewayapi.KindPtr("Gateway"),
								Name:  "test2",
							},
						},
					},
				},
			},
			gateways: []*gwapiv1.Gateway{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test",
						Name:      "test",
					},
					Spec: gwapiv1.GatewaySpec{
						GatewayClassName: "gc1",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test",
						Name:      "test2",
					},
					Spec: gwapiv1.GatewaySpec{
						GatewayClassName: "gc2",
					},
				},
			},
			classes: []*gwapiv1.GatewayClass{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "gc1",
					},
					Spec: gwapiv1.GatewayClassSpec{
						ControllerName: gwapiv1.GatewayController(egv1a1.GatewayControllerName),
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "gc2",
					},
					Spec: gwapiv1.GatewayClassSpec{
						ControllerName: gwapiv1.GatewayController("unmanaged.controller"),
					},
				},
			},
			expect: []gwapiv1.Gateway{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:       "test",
						Name:            "test",
						ResourceVersion: "999",
					},
					Spec: gwapiv1.GatewaySpec{
						GatewayClassName: "gc1",
					},
				},
			},
			expected: true,
		},
		{
			name: "one of two valid parentRefs kind",
			route: &gwapiv1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: gwapiv1.HTTPRouteSpec{
					CommonRouteSpec: gwapiv1.CommonRouteSpec{
						ParentRefs: []gwapiv1.ParentReference{
							{
								Group: gatewayapi.GroupPtr(gwapiv1.GroupName),
								Kind:  gatewayapi.KindPtr("Gateway"),
								Name:  "test",
							},
							{
								Group: gatewayapi.GroupPtr(gwapiv1.GroupName),
								Kind:  gatewayapi.KindPtr("Unsupported"),
								Name:  "test2",
							},
						},
					},
				},
			},
			expected: false,
		},
	}

	// Create the reconciler.
	r := &gatewayAPIReconciler{classController: gwapiv1.GatewayController(egv1a1.GatewayControllerName)}
	ctx := context.Background()

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var objs []client.Object
			for i := range tc.classes {
				objs = append(objs, tc.classes[i])
			}
			for i := range tc.gateways {
				objs = append(objs, tc.gateways[i])
			}
			r.client = fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects(objs...).Build()
			gws, err := validateParentRefs(ctx, r.client, tc.route.Namespace, r.classController, tc.route.Spec.ParentRefs)
			if tc.expected {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
			assert.Equal(t, tc.expect, gws)
		})
	}
}
