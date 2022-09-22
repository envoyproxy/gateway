package kubernetes

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
)

func TestValidateParentRefs(t *testing.T) {
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
						ControllerName: gwapiv1b1.GatewayController(v1alpha1.GatewayControllerName),
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
						ControllerName: gwapiv1b1.GatewayController(v1alpha1.GatewayControllerName),
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
						ControllerName: gwapiv1b1.GatewayController(v1alpha1.GatewayControllerName),
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
	r := &httpRouteReconciler{classController: gwapiv1b1.GatewayController(v1alpha1.GatewayControllerName)}
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
			gws, err := r.validateParentRefs(ctx, tc.route)
			if tc.expected {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
			assert.Equal(t, tc.expect, gws)
		})
	}
}
