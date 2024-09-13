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

func TestAddEnvoyProxyFinalizer(t *testing.T) {
	testCases := []struct {
		name   string
		ep     *egv1a1.EnvoyProxy
		expect []string
	}{
		{
			name: "envoyproxy with no finalizers",
			ep: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
			},
			expect: []string{envoyProxyFinalizer},
		},
		{
			name: "envoyproxy with a different finalizer",
			ep: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-ep",
					Finalizers: []string{"fooFinalizer"},
				},
			},
			expect: []string{"fooFinalizer", envoyProxyFinalizer},
		},
		{
			name: "envoyproxy with existing gatewayclass finalizer",
			ep: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-ep",
					Finalizers: []string{envoyProxyFinalizer},
				},
			},
			expect: []string{envoyProxyFinalizer},
		},
	}
	// Create the reconciler.
	r := new(gatewayAPIReconciler)
	ctx := context.Background()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r.client = fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects(tc.ep).Build()
			err := r.addFinalizer(ctx, tc.ep)
			require.NoError(t, err)
			key := types.NamespacedName{Name: tc.ep.Name}
			err = r.client.Get(ctx, key, tc.ep)
			require.NoError(t, err)
			require.Equal(t, tc.expect, tc.ep.Finalizers)
		})
	}
}

func TestRemoveEnvoyProxyFinalizer(t *testing.T) {
	testCases := []struct {
		name   string
		ep     *egv1a1.EnvoyProxy
		expect []string
	}{
		{
			name: "envoyproxy with no finalizers",
			ep: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
			},
			expect: nil,
		},
		{
			name: "envoyproxy with a different finalizer",
			ep: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-ep",
					Finalizers: []string{"fooFinalizer"},
				},
			},
			expect: []string{"fooFinalizer"},
		},
		{
			name: "envoyproxy with existing gatewayclass finalizer",
			ep: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-ep",
					Finalizers: []string{envoyProxyFinalizer},
				},
			},
			expect: nil,
		},
	}

	// Create the reconciler.
	r := new(gatewayAPIReconciler)
	ctx := context.Background()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r.client = fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects(tc.ep).Build()
			err := r.removeFinalizer(ctx, tc.ep)
			require.NoError(t, err)
			key := types.NamespacedName{Name: tc.ep.Name}
			err = r.client.Get(ctx, key, tc.ep)
			require.NoError(t, err)
			require.Equal(t, tc.expect, tc.ep.Finalizers)
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

func TestProcessGatewayClassParamsRef(t *testing.T) {
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
					Name:       "test",
					Finalizers: []string{gatewayClassFinalizer},
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
					Namespace:  config.DefaultNamespace,
					Name:       "test",
					Finalizers: []string{envoyProxyFinalizer},
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
			namespace:       config.DefaultNamespace,
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
			resourceMap := newResourceMapping()
			err := r.processGatewayClassParamsRef(context.Background(), tc.gc, resourceMap, resourceTree)
			if tc.expected {
				require.NoError(t, err)
				// Ensure the resource tree and map are as expected.
				require.Equal(t, tc.ep, resourceTree.EnvoyProxyForGatewayClass)
			} else {
				require.Error(t, err)
			}
		})
	}
}
