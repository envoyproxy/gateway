// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"os"
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

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
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
					Port:     int32(8080),
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
		httpRouteFilters   []*egv1a1.HTTPRouteFilter
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
												Kind:  gatewayapi.KindPtr(resource.KindService),
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
			name: "httproute with extension filter multiple types same name",
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
											Kind:  gwapiv1.Kind("Bar"),
											Name:  gwapiv1.ObjectName("test"),
										},
									},
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
												Kind:  gatewayapi.KindPtr(resource.KindService),
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
						"kind":       "Bar",
						"metadata": map[string]interface{}{
							"name":      "test",
							"namespace": httpRouteNS,
						},
					},
				},
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
					Kind:    "Bar",
				},
				{
					Group:   "gateway.example.io",
					Version: "v1alpha1",
					Kind:    "Foo",
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
												Kind:  gatewayapi.KindPtr(resource.KindService),
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
												Kind:  gatewayapi.KindPtr(resource.KindService),
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
		{
			name: "multiple httproute with same extension filter",
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
											Value: ptr.To("/1"),
										},
									},
								},
								Filters: []gwapiv1.HTTPRouteFilter{
									{
										Type: gwapiv1.HTTPRouteFilterExtensionRef,
										ExtensionRef: &gwapiv1.LocalObjectReference{
											Group: gwapiv1.Group("gateway.example.io"),
											Kind:  gwapiv1.Kind("Bar"),
											Name:  gwapiv1.ObjectName("test"),
										},
									},
								},
								BackendRefs: []gwapiv1.HTTPBackendRef{
									{
										BackendRef: gwapiv1.BackendRef{
											BackendObjectReference: gwapiv1.BackendObjectReference{
												Group: gatewayapi.GroupPtr(corev1.GroupName),
												Kind:  gatewayapi.KindPtr(resource.KindService),
												Name:  "test",
											},
										},
									},
								},
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: httpRouteNS,
						Name:      "test-2",
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
											Value: ptr.To("/2"),
										},
									},
								},
								Filters: []gwapiv1.HTTPRouteFilter{
									{
										Type: gwapiv1.HTTPRouteFilterExtensionRef,
										ExtensionRef: &gwapiv1.LocalObjectReference{
											Group: gwapiv1.Group("gateway.example.io"),
											Kind:  gwapiv1.Kind("Bar"),
											Name:  gwapiv1.ObjectName("test"),
										},
									},
								},
								BackendRefs: []gwapiv1.HTTPBackendRef{
									{
										BackendRef: gwapiv1.BackendRef{
											BackendObjectReference: gwapiv1.BackendObjectReference{
												Group: gatewayapi.GroupPtr(corev1.GroupName),
												Kind:  gatewayapi.KindPtr(resource.KindService),
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
						"kind":       "Bar",
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
					Kind:    "Bar",
				},
				{
					Group:   "gateway.example.io",
					Version: "v1alpha1",
					Kind:    "Foo",
				},
			},
			expected: true,
		},
		{
			name: "multiple httproute with same extension filter: Envoy Gateway HTTPRouteFilter",
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
											Value: ptr.To("/1"),
										},
									},
								},
								Filters: []gwapiv1.HTTPRouteFilter{
									{
										Type: gwapiv1.HTTPRouteFilterExtensionRef,
										ExtensionRef: &gwapiv1.LocalObjectReference{
											Group: gwapiv1.Group(egv1a1.GroupName),
											Kind:  gwapiv1.Kind(egv1a1.KindHTTPRouteFilter),
											Name:  gwapiv1.ObjectName("test"),
										},
									},
								},
								BackendRefs: []gwapiv1.HTTPBackendRef{
									{
										BackendRef: gwapiv1.BackendRef{
											BackendObjectReference: gwapiv1.BackendObjectReference{
												Group: gatewayapi.GroupPtr(corev1.GroupName),
												Kind:  gatewayapi.KindPtr(resource.KindService),
												Name:  "test",
											},
										},
									},
								},
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: httpRouteNS,
						Name:      "test-2",
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
											Value: ptr.To("/2"),
										},
									},
								},
								Filters: []gwapiv1.HTTPRouteFilter{
									{
										Type: gwapiv1.HTTPRouteFilterExtensionRef,
										ExtensionRef: &gwapiv1.LocalObjectReference{
											Group: gwapiv1.Group(egv1a1.GroupName),
											Kind:  gwapiv1.Kind(egv1a1.KindHTTPRouteFilter),
											Name:  gwapiv1.ObjectName("test"),
										},
									},
								},
								BackendRefs: []gwapiv1.HTTPBackendRef{
									{
										BackendRef: gwapiv1.BackendRef{
											BackendObjectReference: gwapiv1.BackendObjectReference{
												Group: gatewayapi.GroupPtr(corev1.GroupName),
												Kind:  gatewayapi.KindPtr(resource.KindService),
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
			httpRouteFilters: []*egv1a1.HTTPRouteFilter{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       egv1a1.KindHTTPRouteFilter,
						APIVersion: egv1a1.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: httpRouteNS,
						Name:      "test",
					},
					Spec: egv1a1.HTTPRouteFilterSpec{
						URLRewrite: &egv1a1.HTTPURLRewriteFilter{
							Hostname: &egv1a1.HTTPHostnameModifier{
								Type: egv1a1.BackendHTTPHostnameModifier,
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
			logger := logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo)

			ctx := context.Background()

			r := &gatewayAPIReconciler{
				log:             logger,
				classController: gcCtrlName,
				hrfCRDExists:    true,
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
			for _, filter := range tc.httpRouteFilters {
				objs = append(objs, filter)
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
			resourceTree := resource.NewResources()
			resourceMap := newResourceMapping()
			err := r.processHTTPRoutes(ctx, gwNsName, resourceMap, resourceTree)
			if tc.expected {
				require.NoError(t, err)
				// Ensure the resource tree and map are as expected.
				require.Equal(t, tc.routes, resourceTree.HTTPRoutes)
				require.Len(t, resourceTree.ExtensionRefFilters, len(tc.extensionFilters))
				require.Len(t, resourceTree.HTTPRouteFilters, len(tc.httpRouteFilters))
				if tc.extensionFilters != nil {
					for _, filter := range tc.extensionFilters {
						key := utils.NamespacedNameWithGroupKind{
							NamespacedName: types.NamespacedName{
								Namespace: tc.routes[0].Namespace,
								Name:      filter.GetName(),
							},
							GroupKind: schema.GroupKind{
								Group: filter.GroupVersionKind().Group,
								Kind:  filter.GroupVersionKind().Kind,
							},
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
					Port:     int32(8080),
				},
			},
		},
	}
	gwNsName := utils.NamespacedName(gw).String()

	testCases := []struct {
		name               string
		routes             []*gwapiv1.GRPCRoute
		extensionAPIGroups []schema.GroupVersionKind
		expected           bool
	}{
		{
			name: "valid grpcroute",
			routes: []*gwapiv1.GRPCRoute{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test",
						Name:      "test",
					},
					Spec: gwapiv1.GRPCRouteSpec{
						CommonRouteSpec: gwapiv1.CommonRouteSpec{
							ParentRefs: []gwapiv1.ParentReference{
								{
									Name: "test",
								},
							},
						},
						Rules: []gwapiv1.GRPCRouteRule{
							{
								Matches: []gwapiv1.GRPCRouteMatch{
									{
										Method: &gwapiv1.GRPCMethodMatch{
											Method: ptr.To("Ping"),
										},
									},
								},
								BackendRefs: []gwapiv1.GRPCBackendRef{
									{
										BackendRef: gwapiv1.BackendRef{
											BackendObjectReference: gwapiv1.BackendObjectReference{
												Group: gatewayapi.GroupPtr(corev1.GroupName),
												Kind:  gatewayapi.KindPtr(resource.KindService),
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
			logger := logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo)

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
				WithIndex(&gwapiv1.GRPCRoute{}, gatewayGRPCRouteIndex, gatewayGRPCRouteIndexFunc).
				Build()

			// Process the test case httproutes.
			resourceTree := resource.NewResources()
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

func TestProcessHTTPRoutesWithCustomBackends(t *testing.T) {
	ctx := context.Background()

	// Create test custom backend resources
	s3Backend := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "storage.example.io/v1alpha1",
			"kind":       "S3Backend",
			"metadata": map[string]any{
				"name":      "s3-backend",
				"namespace": "default",
			},
			"spec": map[string]any{
				"bucket": "my-s3-bucket",
				"region": "us-west-2",
			},
		},
	}
	s3Backend.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "storage.example.io",
		Version: "v1alpha1",
		Kind:    "S3Backend",
	})

	lambdaBackend := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "compute.example.io/v1alpha1",
			"kind":       "LambdaBackend",
			"metadata": map[string]any{
				"name":      "lambda-backend",
				"namespace": "default",
			},
			"spec": map[string]any{
				"functionName": "my-function",
				"region":       "us-west-2",
			},
		},
	}
	lambdaBackend.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "compute.example.io",
		Version: "v1alpha1",
		Kind:    "LambdaBackend",
	})

	// Create test HTTPRoute with custom backend references
	httpRoute := &gwapiv1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-route",
			Namespace: "default",
		},
		Spec: gwapiv1.HTTPRouteSpec{
			CommonRouteSpec: gwapiv1.CommonRouteSpec{
				ParentRefs: []gwapiv1.ParentReference{
					{
						Name: "test-gateway",
					},
				},
			},
			Rules: []gwapiv1.HTTPRouteRule{
				{
					BackendRefs: []gwapiv1.HTTPBackendRef{
						{
							BackendRef: gwapiv1.BackendRef{
								BackendObjectReference: gwapiv1.BackendObjectReference{
									Group: ptr.To(gwapiv1.Group("storage.example.io")),
									Kind:  ptr.To(gwapiv1.Kind("S3Backend")),
									Name:  "s3-backend",
								},
							},
						},
						{
							BackendRef: gwapiv1.BackendRef{
								BackendObjectReference: gwapiv1.BackendObjectReference{
									Group: ptr.To(gwapiv1.Group("compute.example.io")),
									Kind:  ptr.To(gwapiv1.Kind("LambdaBackend")),
									Name:  "lambda-backend",
								},
							},
						},
					},
				},
			},
		},
	}

	// Create test Gateway
	gateway := &gwapiv1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-gateway",
			Namespace: "default",
		},
		Spec: gwapiv1.GatewaySpec{
			GatewayClassName: "test",
			Listeners: []gwapiv1.Listener{
				{
					Name:     "http",
					Port:     80,
					Protocol: gwapiv1.HTTPProtocolType,
				},
			},
		},
	}

	// Create test GatewayClass
	gatewayClass := &gwapiv1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: gwapiv1.GatewayClassSpec{
			ControllerName: gwapiv1.GatewayController(egv1a1.GatewayControllerName),
		},
	}

	testCases := []struct {
		name                     string
		extBackendGVKs           []schema.GroupVersionKind
		objects                  []client.Object
		expectedExtFiltersCount  int
		expectedBackendRefsCount int
	}{
		{
			name:                     "no custom backend GVKs configured",
			extBackendGVKs:           []schema.GroupVersionKind{},
			objects:                  []client.Object{httpRoute, gateway, gatewayClass},
			expectedExtFiltersCount:  0,
			expectedBackendRefsCount: 0, // Both backends will be rejected due to invalid group
		},
		{
			name: "custom backend GVKs configured with matching resources",
			extBackendGVKs: []schema.GroupVersionKind{
				{Group: "storage.example.io", Version: "v1alpha1", Kind: "S3Backend"},
				{Group: "compute.example.io", Version: "v1alpha1", Kind: "LambdaBackend"},
			},
			objects:                  []client.Object{httpRoute, gateway, gatewayClass, s3Backend, lambdaBackend},
			expectedExtFiltersCount:  2, // Both custom backends should be added to ExtensionRefFilters
			expectedBackendRefsCount: 2, // Both backends should be processed as backend refs
		},
		{
			name: "partial custom backend GVKs configured",
			extBackendGVKs: []schema.GroupVersionKind{
				{Group: "storage.example.io", Version: "v1alpha1", Kind: "S3Backend"},
			},
			objects:                  []client.Object{httpRoute, gateway, gatewayClass, s3Backend, lambdaBackend},
			expectedExtFiltersCount:  1, // Only S3Backend should be added to ExtensionRefFilters
			expectedBackendRefsCount: 1, // Only S3Backend should be processed, LambdaBackend will be rejected
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create fake client with test objects
			fakeClient := fakeclient.NewClientBuilder().
				WithScheme(envoygateway.GetScheme()).
				WithObjects(tc.objects...).
				WithIndex(&gwapiv1.HTTPRoute{}, gatewayHTTPRouteIndex, gatewayHTTPRouteIndexFunc).
				Build()

			// Create reconciler with test configuration
			r := &gatewayAPIReconciler{
				extBackendGVKs: tc.extBackendGVKs,
				log:            logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo),
				client:         fakeClient,
			}

			// Create resource mappings and tree
			resourceMap := newResourceMapping()
			resourceTree := resource.NewResources()
			resourceTree.GatewayClass = gatewayClass

			// Call the function under test
			err := r.processHTTPRoutes(ctx, "default/test-gateway", resourceMap, resourceTree)

			// Verify results
			require.NoError(t, err)
			require.Len(t, resourceMap.extensionRefFilters, tc.expectedExtFiltersCount)
			require.Len(t, resourceMap.allAssociatedBackendRefs, tc.expectedBackendRefsCount)

			// Verify that HTTPRoutes were processed
			require.Len(t, resourceTree.HTTPRoutes, 1)
			require.Equal(t, "test-route", resourceTree.HTTPRoutes[0].Name)
		})
	}
}
