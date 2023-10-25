// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/logging"
)

func TestAddGatewayClassFinalizer(t *testing.T) {
	testCases := []struct {
		name   string
		gc     *gwapiv1.GatewayClass
		expect []string
	}{
		{
			name: "gatewayclass with no finalizers",
			gc: &gwapiv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-gc",
				},
				Spec: gwapiv1.GatewayClassSpec{
					ControllerName: egv1a1.GatewayControllerName,
				},
			},
			expect: []string{gatewayClassFinalizer},
		},
		{
			name: "gatewayclass with a different finalizer",
			gc: &gwapiv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-gc",
					Finalizers: []string{"fooFinalizer"},
				},
				Spec: gwapiv1.GatewayClassSpec{
					ControllerName: egv1a1.GatewayControllerName,
				},
			},
			expect: []string{"fooFinalizer", gatewayClassFinalizer},
		},
		{
			name: "gatewayclass with existing gatewayclass finalizer",
			gc: &gwapiv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-gc",
					Finalizers: []string{gatewayClassFinalizer},
				},
				Spec: gwapiv1.GatewayClassSpec{
					ControllerName: egv1a1.GatewayControllerName,
				},
			},
			expect: []string{gatewayClassFinalizer},
		},
	}

	// Create the reconciler.
	r := new(gatewayAPIReconciler)
	ctx := context.Background()

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			r.client = fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects(tc.gc).Build()
			err := r.addFinalizer(ctx, tc.gc)
			require.NoError(t, err)
			key := types.NamespacedName{Name: tc.gc.Name}
			err = r.client.Get(ctx, key, tc.gc)
			require.NoError(t, err)
			require.Equal(t, tc.expect, tc.gc.Finalizers)
		})
	}
}

func TestRemoveGatewayClassFinalizer(t *testing.T) {
	testCases := []struct {
		name   string
		gc     *gwapiv1.GatewayClass
		expect []string
	}{
		{
			name: "gatewayclass with no finalizers",
			gc: &gwapiv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-gc",
				},
				Spec: gwapiv1.GatewayClassSpec{
					ControllerName: egv1a1.GatewayControllerName,
				},
			},
			expect: nil,
		},
		{
			name: "gatewayclass with a different finalizer",
			gc: &gwapiv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-gc",
					Finalizers: []string{"fooFinalizer"},
				},
				Spec: gwapiv1.GatewayClassSpec{
					ControllerName: egv1a1.GatewayControllerName,
				},
			},
			expect: []string{"fooFinalizer"},
		},
		{
			name: "gatewayclass with existing gatewayclass finalizer",
			gc: &gwapiv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-gc",
					Finalizers: []string{gatewayClassFinalizer},
				},
				Spec: gwapiv1.GatewayClassSpec{
					ControllerName: egv1a1.GatewayControllerName,
				},
			},
			expect: nil,
		},
	}

	// Create the reconciler.
	r := new(gatewayAPIReconciler)
	ctx := context.Background()

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			r.client = fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects(tc.gc).Build()
			err := r.removeFinalizer(ctx, tc.gc)
			require.NoError(t, err)
			key := types.NamespacedName{Name: tc.gc.Name}
			err = r.client.Get(ctx, key, tc.gc)
			require.NoError(t, err)
			require.Equal(t, tc.expect, tc.gc.Finalizers)
		})
	}
}

func TestHasManagedClass(t *testing.T) {
	gcCtrlName := gwapiv1.GatewayController(egv1a1.GatewayControllerName)

	testCases := []struct {
		name     string
		ep       client.Object
		classes  []*gwapiv1.GatewayClass
		expected bool
	}{
		{
			name: "no matching gatewayclasses",
			ep: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: config.DefaultNamespace,
					Name:      "test-envoyproxy",
				},
			},
			classes: []*gwapiv1.GatewayClass{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-gc",
					},
					Spec: gwapiv1.GatewayClassSpec{
						ControllerName: "SomeOtherController",
						ParametersRef: &gwapiv1.ParametersReference{
							Group:     gwapiv1.Group(egv1a1.GroupVersion.Group),
							Kind:      gwapiv1.Kind(egv1a1.KindEnvoyProxy),
							Name:      "test-envoyproxy",
							Namespace: gatewayapi.NamespacePtr(config.DefaultNamespace),
						},
					},
					Status: gwapiv1.GatewayClassStatus{
						Conditions: []metav1.Condition{
							{
								Type:   string(gwapiv1.GatewayClassConditionStatusAccepted),
								Status: metav1.ConditionTrue,
							},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "match one gatewayclass",
			ep: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: config.DefaultNamespace,
					Name:      "test-envoyproxy",
				},
			},
			classes: []*gwapiv1.GatewayClass{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-gc",
					},
					Spec: gwapiv1.GatewayClassSpec{
						ControllerName: gcCtrlName,
						ParametersRef: &gwapiv1.ParametersReference{
							Group:     gwapiv1.Group(egv1a1.GroupVersion.Group),
							Kind:      gwapiv1.Kind(egv1a1.KindEnvoyProxy),
							Name:      "test-envoyproxy",
							Namespace: gatewayapi.NamespacePtr(config.DefaultNamespace),
						},
					},
					Status: gwapiv1.GatewayClassStatus{
						Conditions: []metav1.Condition{
							{
								Type:   string(gwapiv1.GatewayClassConditionStatusAccepted),
								Status: metav1.ConditionTrue,
							},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "envoyproxy in different namespace as eg",
			ep: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "not-eg-ns",
					Name:      "test-envoyproxy",
				},
			},
			classes: []*gwapiv1.GatewayClass{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-gc",
					},
					Spec: gwapiv1.GatewayClassSpec{ControllerName: gcCtrlName},
				},
			},
			expected: false,
		},
		{
			name: "multiple gatewayclasses one with accepted status",
			ep: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: config.DefaultNamespace,
					Name:      "test-envoyproxy",
				},
			},
			classes: []*gwapiv1.GatewayClass{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-gc1",
					},
					Spec: gwapiv1.GatewayClassSpec{
						ControllerName: gcCtrlName,
						ParametersRef: &gwapiv1.ParametersReference{
							Group:     gwapiv1.Group(egv1a1.GroupVersion.Group),
							Kind:      gwapiv1.Kind(egv1a1.KindEnvoyProxy),
							Name:      "test-envoyproxy",
							Namespace: gatewayapi.NamespacePtr(config.DefaultNamespace),
						},
					},
					Status: gwapiv1.GatewayClassStatus{
						Conditions: []metav1.Condition{
							{
								Type:   string(gwapiv1.GatewayClassConditionStatusAccepted),
								Status: metav1.ConditionTrue,
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-gc2",
					},
					Spec: gwapiv1.GatewayClassSpec{
						ControllerName: gcCtrlName,
						ParametersRef: &gwapiv1.ParametersReference{
							Group:     gwapiv1.Group(egv1a1.GroupVersion.Group),
							Kind:      gwapiv1.Kind(egv1a1.KindEnvoyProxy),
							Name:      "test-envoyproxy",
							Namespace: gatewayapi.NamespacePtr(config.DefaultNamespace),
						},
					},
					Status: gwapiv1.GatewayClassStatus{
						Conditions: []metav1.Condition{
							{
								Type:   string(gwapiv1.GatewayClassConditionStatusAccepted),
								Status: metav1.ConditionFalse,
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

		// Create the reconciler.
		logger := logging.DefaultLogger(egv1a1.LogLevelInfo)
		r := &gatewayAPIReconciler{
			log:             logger,
			classController: gcCtrlName,
			namespace:       config.DefaultNamespace,
		}

		// Run the test cases.
		t.Run(tc.name, func(t *testing.T) {
			// Add the test case objects to the reconciler client.
			objs := []client.Object{tc.ep}
			for _, gc := range tc.classes {
				objs = append(objs, gc)
			}

			// Create the client.
			r.client = fakeclient.NewClientBuilder().
				WithScheme(envoygateway.GetScheme()).
				WithObjects(objs...).
				Build()

			// Process the test case gatewayclasses.
			results := r.hasManagedClass(tc.ep)
			require.Equal(t, tc.expected, results)
		})
	}
}

func TestProcessParamsRef(t *testing.T) {
	gcCtrlName := gwapiv1.GatewayController(egv1a1.GatewayControllerName)

	testCases := []struct {
		name     string
		gc       *gwapiv1.GatewayClass
		ep       *egv1a1.EnvoyProxy
		expected bool
	}{
		{
			name: "valid envoyproxy reference",
			gc: &gwapiv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: gwapiv1.GatewayClassSpec{
					ControllerName: gcCtrlName,
					ParametersRef: &gwapiv1.ParametersReference{
						Group:     gwapiv1.Group(egv1a1.GroupVersion.Group),
						Kind:      gwapiv1.Kind(egv1a1.KindEnvoyProxy),
						Name:      "test",
						Namespace: gatewayapi.NamespacePtr(config.DefaultNamespace),
					},
				},
			},
			ep: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: config.DefaultNamespace,
					Name:      "test",
				},
			},
			expected: true,
		},
		{
			name: "envoyproxy kind does not exist",
			gc: &gwapiv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: gwapiv1.GatewayClassSpec{
					ControllerName: gcCtrlName,
					ParametersRef: &gwapiv1.ParametersReference{
						Group:     gwapiv1.Group(egv1a1.GroupVersion.Group),
						Kind:      gwapiv1.Kind(egv1a1.KindEnvoyProxy),
						Name:      "test",
						Namespace: gatewayapi.NamespacePtr(config.DefaultNamespace),
					},
				},
			},
			expected: false,
		},
		{
			name: "referenced envoyproxy does not exist",
			gc: &gwapiv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: gwapiv1.GatewayClassSpec{
					ControllerName: gcCtrlName,
					ParametersRef: &gwapiv1.ParametersReference{
						Group:     gwapiv1.Group(egv1a1.GroupVersion.Group),
						Kind:      gwapiv1.Kind(egv1a1.KindEnvoyProxy),
						Name:      "non-exist",
						Namespace: gatewayapi.NamespacePtr(config.DefaultNamespace),
					},
				},
			},
			ep: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: config.DefaultNamespace,
					Name:      "test",
				},
			},
			expected: false,
		},
		{
			name: "invalid gatewayclass parameters ref",
			gc: &gwapiv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: gwapiv1.GatewayClassSpec{
					ControllerName: gcCtrlName,
					ParametersRef: &gwapiv1.ParametersReference{
						Group:     gwapiv1.Group("UnSupportedGroup"),
						Kind:      gwapiv1.Kind("UnSupportedKind"),
						Name:      "test",
						Namespace: gatewayapi.NamespacePtr(config.DefaultNamespace),
					},
				},
			},
			ep: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: config.DefaultNamespace,
					Name:      "test",
				},
			},
			expected: false,
		},
	}

	for i := range testCases {
		tc := testCases[i]

		// Create the reconciler.
		logger := logging.DefaultLogger(egv1a1.LogLevelInfo)

		r := &gatewayAPIReconciler{
			log:             logger,
			classController: gcCtrlName,
		}

		// Run the test cases.
		t.Run(tc.name, func(t *testing.T) {
			if tc.ep != nil {
				r.client = fakeclient.NewClientBuilder().
					WithScheme(envoygateway.GetScheme()).
					WithObjects(tc.ep).
					Build()
			} else {
				r.client = fakeclient.NewClientBuilder().
					Build()
			}

			// Process the test case gatewayclasses.
			resourceTree := gatewayapi.NewResources()
			err := r.processParamsRef(context.Background(), tc.gc, resourceTree)
			if tc.expected {
				require.NoError(t, err)
				// Ensure the resource tree and map are as expected.
				require.Equal(t, tc.ep, resourceTree.EnvoyProxy)
			} else {
				require.Error(t, err)
			}
		})
	}
}
