// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	cfgv1a1 "github.com/envoyproxy/gateway/api/config/v1alpha1"
	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/log"
	"github.com/envoyproxy/gateway/internal/provider/utils"
)

func TestProcessHTTPRoutes(t *testing.T) {
	// The gatewayclass configured for the reconciler and referenced by test cases.
	gcCtrlName := gwapiv1b1.GatewayController(cfgv1a1.GatewayControllerName)
	gc := &gwapiv1b1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: gwapiv1b1.GatewayClassSpec{
			ControllerName: gcCtrlName,
		},
	}

	// The gateway referenced by test cases.
	gw := &gwapiv1b1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "test",
		},
		Spec: gwapiv1b1.GatewaySpec{
			GatewayClassName: gwapiv1b1.ObjectName(gc.Name),
			Listeners: []gwapiv1b1.Listener{
				{
					Name:     "http",
					Protocol: gwapiv1b1.HTTPProtocolType,
					Port:     gwapiv1b1.PortNumber(int32(8080)),
				},
			},
		},
	}
	gwNsName := utils.NamespacedName(gw).String()

	testCases := []struct {
		name             string
		routes           []*gwapiv1b1.HTTPRoute
		authenFilters    []*egv1a1.AuthenticationFilter
		rateLimitFilters []*egv1a1.RateLimitFilter
		expected         bool
	}{
		{
			name: "valid httproute",
			routes: []*gwapiv1b1.HTTPRoute{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test",
						Name:      "test",
					},
					Spec: gwapiv1b1.HTTPRouteSpec{
						CommonRouteSpec: gwapiv1b1.CommonRouteSpec{
							ParentRefs: []gwapiv1b1.ParentReference{
								{
									Name: "test",
								},
							},
						},
						Rules: []gwapiv1b1.HTTPRouteRule{
							{
								Matches: []gwapiv1b1.HTTPRouteMatch{
									{
										Path: &gwapiv1b1.HTTPPathMatch{
											Type:  gatewayapi.PathMatchTypePtr(gwapiv1b1.PathMatchPathPrefix),
											Value: gatewayapi.StringPtr("/"),
										},
									},
								},
								BackendRefs: []gwapiv1b1.HTTPBackendRef{
									{
										BackendRef: gwapiv1b1.BackendRef{
											BackendObjectReference: gwapiv1b1.BackendObjectReference{
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
			name: "httproute with one authenticationfilter",
			routes: []*gwapiv1b1.HTTPRoute{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test",
						Name:      "test",
					},
					Spec: gwapiv1b1.HTTPRouteSpec{
						CommonRouteSpec: gwapiv1b1.CommonRouteSpec{
							ParentRefs: []gwapiv1b1.ParentReference{
								{
									Name: "test",
								},
							},
						},
						Rules: []gwapiv1b1.HTTPRouteRule{
							{
								Matches: []gwapiv1b1.HTTPRouteMatch{
									{
										Path: &gwapiv1b1.HTTPPathMatch{
											Type:  gatewayapi.PathMatchTypePtr(gwapiv1b1.PathMatchPathPrefix),
											Value: gatewayapi.StringPtr("/"),
										},
									},
								},
								Filters: []gwapiv1b1.HTTPRouteFilter{
									{
										Type: gwapiv1b1.HTTPRouteFilterExtensionRef,
										ExtensionRef: &gwapiv1b1.LocalObjectReference{
											Group: gwapiv1b1.Group(egv1a1.GroupVersion.Group),
											Kind:  gwapiv1b1.Kind(egv1a1.KindAuthenticationFilter),
											Name:  gwapiv1b1.ObjectName("test"),
										},
									},
								},
								BackendRefs: []gwapiv1b1.HTTPBackendRef{
									{
										BackendRef: gwapiv1b1.BackendRef{
											BackendObjectReference: gwapiv1b1.BackendObjectReference{
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
			authenFilters: []*egv1a1.AuthenticationFilter{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       egv1a1.KindAuthenticationFilter,
						APIVersion: egv1a1.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test",
						Name:      "test",
					},
					Spec: egv1a1.AuthenticationFilterSpec{
						Type: egv1a1.JwtAuthenticationFilterProviderType,
						JwtProviders: []egv1a1.JwtAuthenticationFilterProvider{
							{
								Name:      "test",
								Issuer:    "https://www.test.local",
								Audiences: []string{"test.local"},
								RemoteJWKS: egv1a1.RemoteJWKS{
									URI: "https://test.local/jwt/public-key/jwks.json",
								},
							},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "httproute with one rateLimitfilter",
			routes: []*gwapiv1b1.HTTPRoute{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test",
						Name:      "test",
					},
					Spec: gwapiv1b1.HTTPRouteSpec{
						CommonRouteSpec: gwapiv1b1.CommonRouteSpec{
							ParentRefs: []gwapiv1b1.ParentReference{
								{
									Name: "test",
								},
							},
						},
						Rules: []gwapiv1b1.HTTPRouteRule{
							{
								Matches: []gwapiv1b1.HTTPRouteMatch{
									{
										Path: &gwapiv1b1.HTTPPathMatch{
											Type:  gatewayapi.PathMatchTypePtr(gwapiv1b1.PathMatchPathPrefix),
											Value: gatewayapi.StringPtr("/"),
										},
									},
								},
								Filters: []gwapiv1b1.HTTPRouteFilter{
									{
										Type: gwapiv1b1.HTTPRouteFilterExtensionRef,
										ExtensionRef: &gwapiv1b1.LocalObjectReference{
											Group: gwapiv1b1.Group(egv1a1.GroupVersion.Group),
											Kind:  gwapiv1b1.Kind(egv1a1.KindRateLimitFilter),
											Name:  gwapiv1b1.ObjectName("test"),
										},
									},
								},
								BackendRefs: []gwapiv1b1.HTTPBackendRef{
									{
										BackendRef: gwapiv1b1.BackendRef{
											BackendObjectReference: gwapiv1b1.BackendObjectReference{
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
			rateLimitFilters: []*egv1a1.RateLimitFilter{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       egv1a1.KindRateLimitFilter,
						APIVersion: egv1a1.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test",
						Name:      "test",
					},
					Spec: egv1a1.RateLimitFilterSpec{
						Type: egv1a1.GlobalRateLimitType,
						Global: &egv1a1.GlobalRateLimit{
							Rules: []egv1a1.RateLimitRule{
								{
									ClientSelectors: []egv1a1.RateLimitSelectCondition{
										{
											Headers: []egv1a1.HeaderMatch{
												{
													Name:  "x-user-id",
													Value: gatewayapi.StringPtr("one"),
												},
											},
										},
									},
									Limit: egv1a1.RateLimitValue{
										Requests: 5,
										Unit:     "Second",
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
			name: "httproute with one authenticationfilter and ratelimitfilter",
			routes: []*gwapiv1b1.HTTPRoute{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test",
						Name:      "test",
					},
					Spec: gwapiv1b1.HTTPRouteSpec{
						CommonRouteSpec: gwapiv1b1.CommonRouteSpec{
							ParentRefs: []gwapiv1b1.ParentReference{
								{
									Name: "test",
								},
							},
						},
						Rules: []gwapiv1b1.HTTPRouteRule{
							{
								Matches: []gwapiv1b1.HTTPRouteMatch{
									{
										Path: &gwapiv1b1.HTTPPathMatch{
											Type:  gatewayapi.PathMatchTypePtr(gwapiv1b1.PathMatchPathPrefix),
											Value: gatewayapi.StringPtr("/"),
										},
									},
								},
								Filters: []gwapiv1b1.HTTPRouteFilter{
									{
										Type: gwapiv1b1.HTTPRouteFilterExtensionRef,
										ExtensionRef: &gwapiv1b1.LocalObjectReference{
											Group: gwapiv1b1.Group(egv1a1.GroupVersion.Group),
											Kind:  gwapiv1b1.Kind(egv1a1.KindAuthenticationFilter),
											Name:  gwapiv1b1.ObjectName("test"),
										},
									},
								},
								BackendRefs: []gwapiv1b1.HTTPBackendRef{
									{
										BackendRef: gwapiv1b1.BackendRef{
											BackendObjectReference: gwapiv1b1.BackendObjectReference{
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
			authenFilters: []*egv1a1.AuthenticationFilter{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       egv1a1.KindAuthenticationFilter,
						APIVersion: egv1a1.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test",
						Name:      "test",
					},
					Spec: egv1a1.AuthenticationFilterSpec{
						Type: egv1a1.JwtAuthenticationFilterProviderType,
						JwtProviders: []egv1a1.JwtAuthenticationFilterProvider{
							{
								Name:      "test",
								Issuer:    "https://www.test.local",
								Audiences: []string{"test.local"},
								RemoteJWKS: egv1a1.RemoteJWKS{
									URI: "https://test.local/jwt/public-key/jwks.json",
								},
							},
						},
					},
				},
			},
			rateLimitFilters: []*egv1a1.RateLimitFilter{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       egv1a1.KindRateLimitFilter,
						APIVersion: egv1a1.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test",
						Name:      "test",
					},
					Spec: egv1a1.RateLimitFilterSpec{
						Type: egv1a1.GlobalRateLimitType,
						Global: &egv1a1.GlobalRateLimit{
							Rules: []egv1a1.RateLimitRule{
								{
									ClientSelectors: []egv1a1.RateLimitSelectCondition{
										{
											Headers: []egv1a1.HeaderMatch{
												{
													Name:  "x-user-id",
													Value: gatewayapi.StringPtr("one"),
												},
											},
										},
									},
									Limit: egv1a1.RateLimitValue{
										Requests: 5,
										Unit:     "Second",
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

		// Add objects referenced by test cases.
		objs := []client.Object{gc, gw}

		// Create the reconciler.
		logger, err := log.NewLogger()
		require.NoError(t, err)
		r := &gatewayAPIReconciler{
			log:             logger,
			classController: gcCtrlName,
		}
		ctx := context.Background()

		// Run the test cases.
		t.Run(tc.name, func(t *testing.T) {
			// Add the test case objects to the reconciler client.
			for _, route := range tc.routes {
				objs = append(objs, route)
			}
			for _, filter := range tc.authenFilters {
				objs = append(objs, filter)
			}
			for _, filter := range tc.rateLimitFilters {
				objs = append(objs, filter)
			}
			r.client = fakeclient.NewClientBuilder().
				WithScheme(envoygateway.GetScheme()).
				WithObjects(objs...).
				WithIndex(&gwapiv1b1.HTTPRoute{}, gatewayHTTPRouteIndex, gatewayHTTPRouteIndexFunc).
				Build()

			// Process the test case httproutes.
			resourceTree := gatewayapi.NewResources()
			resourceMap := newResourceMapping()
			err := r.processHTTPRoutes(ctx, gwNsName, resourceMap, resourceTree)
			if tc.expected {
				require.NoError(t, err)
				// Ensure the resource tree and map are as expected.
				require.Equal(t, tc.routes, resourceTree.HTTPRoutes)
				if tc.authenFilters != nil {
					for i, filter := range tc.authenFilters {
						key := types.NamespacedName{
							// The AuthenticationFilter must be in the same namespace as the HTTPRoute.
							Namespace: tc.routes[i].Namespace,
							Name:      filter.Name,
						}
						require.Equal(t, filter, resourceMap.authenFilters[key])
					}
				}
				if tc.rateLimitFilters != nil {
					for i, filter := range tc.rateLimitFilters {
						key := types.NamespacedName{
							// The RateLimitFilter must be in the same namespace as the HTTPRoute.
							Namespace: tc.routes[i].Namespace,
							Name:      filter.Name,
						}
						require.Equal(t, filter, resourceMap.rateLimitFilters[key])
					}
				}
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestValidateHTTPRouteParentRefs(t *testing.T) {
	testCases := []struct {
		name     string
		route    *gwapiv1b1.HTTPRoute
		gateways []*gwapiv1b1.Gateway
		classes  []*gwapiv1b1.GatewayClass
		expect   []gwapiv1b1.Gateway
		expected bool
	}{
		{
			name: "valid parentRef",
			route: &gwapiv1b1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: gwapiv1b1.HTTPRouteSpec{
					CommonRouteSpec: gwapiv1b1.CommonRouteSpec{
						ParentRefs: []gwapiv1b1.ParentReference{
							{
								Group: gatewayapi.GroupPtr(gwapiv1b1.GroupName),
								Kind:  gatewayapi.KindPtr("Gateway"),
								Name:  "test",
							},
						},
					},
				},
			},
			gateways: []*gwapiv1b1.Gateway{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test",
						Name:      "test",
					},
					Spec: gwapiv1b1.GatewaySpec{
						GatewayClassName: "gc1",
					},
				},
			},
			classes: []*gwapiv1b1.GatewayClass{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "gc1",
					},
					Spec: gwapiv1b1.GatewayClassSpec{
						ControllerName: gwapiv1b1.GatewayController(cfgv1a1.GatewayControllerName),
					},
				},
			},
			expect: []gwapiv1b1.Gateway{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Gateway",
						APIVersion: gwapiv1b1.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace:       "test",
						Name:            "test",
						ResourceVersion: "999",
					},
					Spec: gwapiv1b1.GatewaySpec{
						GatewayClassName: "gc1",
					},
				},
			},
			expected: true,
		},
		{
			name: "invalid parentRef group",
			route: &gwapiv1b1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: gwapiv1b1.HTTPRouteSpec{
					CommonRouteSpec: gwapiv1b1.CommonRouteSpec{
						ParentRefs: []gwapiv1b1.ParentReference{
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
			route: &gwapiv1b1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: gwapiv1b1.HTTPRouteSpec{
					CommonRouteSpec: gwapiv1b1.CommonRouteSpec{
						ParentRefs: []gwapiv1b1.ParentReference{
							{
								Group: gatewayapi.GroupPtr(gwapiv1b1.GroupName),
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
			route: &gwapiv1b1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: gwapiv1b1.HTTPRouteSpec{
					CommonRouteSpec: gwapiv1b1.CommonRouteSpec{
						ParentRefs: []gwapiv1b1.ParentReference{
							{
								Group: gatewayapi.GroupPtr(gwapiv1b1.GroupName),
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
			route: &gwapiv1b1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: gwapiv1b1.HTTPRouteSpec{
					CommonRouteSpec: gwapiv1b1.CommonRouteSpec{
						ParentRefs: []gwapiv1b1.ParentReference{
							{
								Group: gatewayapi.GroupPtr(gwapiv1b1.GroupName),
								Kind:  gatewayapi.KindPtr("Gateway"),
								Name:  "test",
							},
							{
								Group: gatewayapi.GroupPtr(gwapiv1b1.GroupName),
								Kind:  gatewayapi.KindPtr("Gateway"),
								Name:  "test2",
							},
						},
					},
				},
			},
			gateways: []*gwapiv1b1.Gateway{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test",
						Name:      "test",
					},
					Spec: gwapiv1b1.GatewaySpec{
						GatewayClassName: "gc1",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test",
						Name:      "test2",
					},
					Spec: gwapiv1b1.GatewaySpec{
						GatewayClassName: "gc1",
					},
				},
			},
			classes: []*gwapiv1b1.GatewayClass{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "gc1",
					},
					Spec: gwapiv1b1.GatewayClassSpec{
						ControllerName: gwapiv1b1.GatewayController(cfgv1a1.GatewayControllerName),
					},
				},
			},
			expect: []gwapiv1b1.Gateway{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Gateway",
						APIVersion: gwapiv1b1.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace:       "test",
						Name:            "test",
						ResourceVersion: "999",
					},
					Spec: gwapiv1b1.GatewaySpec{
						GatewayClassName: "gc1",
					},
				},
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Gateway",
						APIVersion: gwapiv1b1.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace:       "test",
						Name:            "test2",
						ResourceVersion: "999",
					},
					Spec: gwapiv1b1.GatewaySpec{
						GatewayClassName: "gc1",
					},
				},
			},
			expected: true,
		},
		{
			name: "one of two parentRefs are managed",
			route: &gwapiv1b1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: gwapiv1b1.HTTPRouteSpec{
					CommonRouteSpec: gwapiv1b1.CommonRouteSpec{
						ParentRefs: []gwapiv1b1.ParentReference{
							{
								Group: gatewayapi.GroupPtr(gwapiv1b1.GroupName),
								Kind:  gatewayapi.KindPtr("Gateway"),
								Name:  "test",
							},
							{
								Group: gatewayapi.GroupPtr(gwapiv1b1.GroupName),
								Kind:  gatewayapi.KindPtr("Gateway"),
								Name:  "test2",
							},
						},
					},
				},
			},
			gateways: []*gwapiv1b1.Gateway{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test",
						Name:      "test",
					},
					Spec: gwapiv1b1.GatewaySpec{
						GatewayClassName: "gc1",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test",
						Name:      "test2",
					},
					Spec: gwapiv1b1.GatewaySpec{
						GatewayClassName: "gc2",
					},
				},
			},
			classes: []*gwapiv1b1.GatewayClass{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "gc1",
					},
					Spec: gwapiv1b1.GatewayClassSpec{
						ControllerName: gwapiv1b1.GatewayController(cfgv1a1.GatewayControllerName),
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "gc2",
					},
					Spec: gwapiv1b1.GatewayClassSpec{
						ControllerName: gwapiv1b1.GatewayController("unmanaged.controller"),
					},
				},
			},
			expect: []gwapiv1b1.Gateway{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Gateway",
						APIVersion: gwapiv1b1.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace:       "test",
						Name:            "test",
						ResourceVersion: "999",
					},
					Spec: gwapiv1b1.GatewaySpec{
						GatewayClassName: "gc1",
					},
				},
			},
			expected: true,
		},
		{
			name: "one of two valid parentRefs kind",
			route: &gwapiv1b1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: gwapiv1b1.HTTPRouteSpec{
					CommonRouteSpec: gwapiv1b1.CommonRouteSpec{
						ParentRefs: []gwapiv1b1.ParentReference{
							{
								Group: gatewayapi.GroupPtr(gwapiv1b1.GroupName),
								Kind:  gatewayapi.KindPtr("Gateway"),
								Name:  "test",
							},
							{
								Group: gatewayapi.GroupPtr(gwapiv1b1.GroupName),
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
	r := &gatewayAPIReconciler{classController: gwapiv1b1.GatewayController(cfgv1a1.GatewayControllerName)}
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
