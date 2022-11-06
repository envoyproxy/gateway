// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/log"
)

func TestGatewayHasMatchingController(t *testing.T) {
	match := &gwapiv1b1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "matched",
		},
		Spec: gwapiv1b1.GatewayClassSpec{
			ControllerName: v1alpha1.GatewayControllerName,
		},
	}

	nonMatch := &gwapiv1b1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "non-matched",
		},
		Spec: gwapiv1b1.GatewayClassSpec{
			ControllerName: "not.configured/controller-name",
		},
	}

	testCases := []struct {
		name   string
		obj    client.Object
		expect bool
	}{
		{
			name: "matching object type, gatewayclass, and controller name",
			obj: &gwapiv1b1.Gateway{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Gateway",
					APIVersion: fmt.Sprintf("%s/%s", gwapiv1b1.GroupName, gwapiv1b1.GroupVersion.Version),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				Spec: gwapiv1b1.GatewaySpec{
					GatewayClassName: gwapiv1b1.ObjectName(match.Name),
				},
			},
			expect: true,
		},
		{
			name: "matching object type but gatewayclass doesn't exist",
			obj: &gwapiv1b1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				Spec: gwapiv1b1.GatewaySpec{
					GatewayClassName: "non-existent-gc",
				},
			},
			expect: false,
		},
		{
			name: "matching object type and gatewayclass but not controller name",
			obj: &gwapiv1b1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				Spec: gwapiv1b1.GatewaySpec{
					GatewayClassName: gwapiv1b1.ObjectName(nonMatch.Name),
				},
			},
			expect: false,
		},
		{
			name: "gatewayclass name match but object type doesn't match",
			obj: &gwapiv1b1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: gwapiv1b1.GatewayClassSpec{
					ControllerName: gwapiv1b1.GatewayController(v1alpha1.GatewayControllerName),
				},
			},
			expect: false,
		},
		{
			name: "matching but very long name",
			obj: &gwapiv1b1.Gateway{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Gateway",
					APIVersion: fmt.Sprintf("%s/%s", gwapiv1b1.GroupName, gwapiv1b1.GroupVersion.Version),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "superdupermegalongnamethatisridiculouslylongandwaylongerthanitshouldeverbeinsideofkubernetes",
					Namespace: "test",
				},
				Spec: gwapiv1b1.GatewaySpec{
					GatewayClassName: gwapiv1b1.ObjectName(match.Name),
				},
			},
			expect: true,
		},
	}

	// Create the reconciler.
	logger, err := log.NewLogger()
	require.NoError(t, err)
	r := gatewayAPIReconciler{
		classController: v1alpha1.GatewayControllerName,
		log:             logger,
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			r.client = fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects(match, nonMatch, tc.obj).Build()
			actual := r.hasMatchingControllerForGateway(tc.obj)
			require.Equal(t, tc.expect, actual)
		})
	}
}

func TestIsAccepted(t *testing.T) {
	testCases := []struct {
		name   string
		gc     *gwapiv1b1.GatewayClass
		expect bool
	}{
		{
			name: "gatewayclass accepted condition",
			gc: &gwapiv1b1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: gwapiv1b1.GatewayClassSpec{
					ControllerName: gwapiv1b1.GatewayController(v1alpha1.GatewayControllerName),
				},
				Status: gwapiv1b1.GatewayClassStatus{
					Conditions: []metav1.Condition{
						{
							Type:   string(gwapiv1b1.GatewayClassConditionStatusAccepted),
							Status: metav1.ConditionTrue,
						},
					},
				},
			},
			expect: true,
		},
		{
			name: "gatewayclass not accepted condition",
			gc: &gwapiv1b1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: gwapiv1b1.GatewayClassSpec{
					ControllerName: gwapiv1b1.GatewayController(v1alpha1.GatewayControllerName),
				},
				Status: gwapiv1b1.GatewayClassStatus{
					Conditions: []metav1.Condition{
						{
							Type:   string(gwapiv1b1.GatewayClassConditionStatusAccepted),
							Status: metav1.ConditionFalse,
						},
					},
				},
			},
			expect: false,
		},
		{
			name: "no gatewayclass accepted condition type",
			gc: &gwapiv1b1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: gwapiv1b1.GatewayClassSpec{
					ControllerName: gwapiv1b1.GatewayController(v1alpha1.GatewayControllerName),
				},
				Status: gwapiv1b1.GatewayClassStatus{
					Conditions: []metav1.Condition{
						{
							Type:   "SomeOtherType",
							Status: metav1.ConditionTrue,
						},
					},
				},
			},
			expect: false,
		},
		{
			name:   "nil gatewayclass",
			expect: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			actual := isAccepted(tc.gc)
			require.Equal(t, tc.expect, actual)
		})
	}
}

func TestGatewaysOfClass(t *testing.T) {
	gc := &gwapiv1b1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
	}
	testCases := []struct {
		name   string
		gws    []gwapiv1b1.Gateway
		expect int
	}{
		{
			name: "no matching gateways",
			gws: []gwapiv1b1.Gateway{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
					},
					Spec: gwapiv1b1.GatewaySpec{
						GatewayClassName: gwapiv1b1.ObjectName("no-match"),
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
					},
					Spec: gwapiv1b1.GatewaySpec{
						GatewayClassName: gwapiv1b1.ObjectName("no-match2"),
					},
				},
			},
			expect: 0,
		},
		{
			name: "one of two matching gateways",
			gws: []gwapiv1b1.Gateway{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
					},
					Spec: gwapiv1b1.GatewaySpec{
						GatewayClassName: gwapiv1b1.ObjectName(gc.Name),
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test2",
						Namespace: "test",
					},
					Spec: gwapiv1b1.GatewaySpec{
						GatewayClassName: gwapiv1b1.ObjectName("no-match"),
					},
				},
			},
			expect: 1,
		},
		{
			name: "two of two matching gateways",
			gws: []gwapiv1b1.Gateway{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
					},
					Spec: gwapiv1b1.GatewaySpec{
						GatewayClassName: gwapiv1b1.ObjectName(gc.Name),
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test2",
						Namespace: "test",
					},
					Spec: gwapiv1b1.GatewaySpec{
						GatewayClassName: gwapiv1b1.ObjectName(gc.Name),
					},
				},
			},
			expect: 2,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			gwList := &gwapiv1b1.GatewayList{Items: tc.gws}
			actual := gatewaysOfClass(gc, gwList)
			require.Equal(t, tc.expect, len(actual))
		})
	}
}

func TestAddFinalizer(t *testing.T) {
	testCases := []struct {
		name   string
		gc     *gwapiv1b1.GatewayClass
		expect []string
	}{
		{
			name: "gatewayclass with no finalizers",
			gc: &gwapiv1b1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-gc",
				},
				Spec: gwapiv1b1.GatewayClassSpec{
					ControllerName: v1alpha1.GatewayControllerName,
				},
			},
			expect: []string{gatewayClassFinalizer},
		},
		{
			name: "gatewayclass with a different finalizer",
			gc: &gwapiv1b1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-gc",
					Finalizers: []string{"fooFinalizer"},
				},
				Spec: gwapiv1b1.GatewayClassSpec{
					ControllerName: v1alpha1.GatewayControllerName,
				},
			},
			expect: []string{"fooFinalizer", gatewayClassFinalizer},
		},
		{
			name: "gatewayclass with existing gatewayclass finalizer",
			gc: &gwapiv1b1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-gc",
					Finalizers: []string{gatewayClassFinalizer},
				},
				Spec: gwapiv1b1.GatewayClassSpec{
					ControllerName: v1alpha1.GatewayControllerName,
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

func TestRemoveFinalizer(t *testing.T) {
	testCases := []struct {
		name   string
		gc     *gwapiv1b1.GatewayClass
		expect []string
	}{
		{
			name: "gatewayclass with no finalizers",
			gc: &gwapiv1b1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-gc",
				},
				Spec: gwapiv1b1.GatewayClassSpec{
					ControllerName: v1alpha1.GatewayControllerName,
				},
			},
			expect: nil,
		},
		{
			name: "gatewayclass with a different finalizer",
			gc: &gwapiv1b1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-gc",
					Finalizers: []string{"fooFinalizer"},
				},
				Spec: gwapiv1b1.GatewayClassSpec{
					ControllerName: v1alpha1.GatewayControllerName,
				},
			},
			expect: []string{"fooFinalizer"},
		},
		{
			name: "gatewayclass with existing gatewayclass finalizer",
			gc: &gwapiv1b1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-gc",
					Finalizers: []string{gatewayClassFinalizer},
				},
				Spec: gwapiv1b1.GatewayClassSpec{
					ControllerName: v1alpha1.GatewayControllerName,
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

func TestSecretsAndRefGrantsForGateway(t *testing.T) {
	testCases := []struct {
		name      string
		gw        *gwapiv1b1.Gateway
		secrets   []corev1.Secret
		refGrants []gwapiv1a2.ReferenceGrant
	}{
		{
			name: "gateway with no https listeners",
			gw: &gwapiv1b1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-gw",
					Namespace: "test-ns",
				},
				Spec: gwapiv1b1.GatewaySpec{
					GatewayClassName: "test-gc",
					Listeners: []gwapiv1b1.Listener{
						{
							Name:     "http",
							Port:     gwapiv1b1.PortNumber(int32(80)),
							Protocol: gwapiv1b1.HTTPProtocolType,
						},
					},
				},
			},
		},
		{
			name: "gateway with one https listener and one same namespace secret",
			gw: &gwapiv1b1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-gw",
					Namespace: "test-ns",
				},
				Spec: gwapiv1b1.GatewaySpec{
					GatewayClassName: "test-gc",
					Listeners: []gwapiv1b1.Listener{
						{
							Name:     "tls",
							Port:     gwapiv1b1.PortNumber(int32(443)),
							Protocol: gwapiv1b1.HTTPSProtocolType,
							TLS: &gwapiv1b1.GatewayTLSConfig{
								Mode: gatewayapi.TLSModeTypePtr(gwapiv1b1.TLSModeTerminate),
								CertificateRefs: []gwapiv1b1.SecretObjectReference{
									{
										Name: gwapiv1b1.ObjectName("test-secret"),
									},
								},
							},
						},
					},
				},
			},
			secrets: []corev1.Secret{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Secret",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:            "test-secret",
						Namespace:       "test-ns",
						ResourceVersion: "1",
					},
				},
			},
		},
		{
			name: "gateway with one http and one https listener with one same namespace secret",
			gw: &gwapiv1b1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-gw",
					Namespace: "test-ns",
				},
				Spec: gwapiv1b1.GatewaySpec{
					GatewayClassName: "test-gc",
					Listeners: []gwapiv1b1.Listener{
						{
							Name:     "http",
							Port:     gwapiv1b1.PortNumber(int32(80)),
							Protocol: gwapiv1b1.HTTPProtocolType,
						},
						{
							Name:     "tls",
							Port:     gwapiv1b1.PortNumber(int32(443)),
							Protocol: gwapiv1b1.HTTPSProtocolType,
							TLS: &gwapiv1b1.GatewayTLSConfig{
								Mode: gatewayapi.TLSModeTypePtr(gwapiv1b1.TLSModeTerminate),
								CertificateRefs: []gwapiv1b1.SecretObjectReference{
									{
										Name: gwapiv1b1.ObjectName("test-secret"),
									},
								},
							},
						},
					},
				},
			},
			secrets: []corev1.Secret{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Secret",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:            "test-secret",
						Namespace:       "test-ns",
						ResourceVersion: "1",
					},
				},
			},
		},
		{
			name: "gateway with two https listeners each with two same namespace secrets",
			gw: &gwapiv1b1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-gw",
					Namespace: "test-ns",
				},
				Spec: gwapiv1b1.GatewaySpec{
					GatewayClassName: "test-gc",
					Listeners: []gwapiv1b1.Listener{
						{
							Name:     "tls1",
							Port:     gwapiv1b1.PortNumber(int32(443)),
							Protocol: gwapiv1b1.HTTPSProtocolType,
							TLS: &gwapiv1b1.GatewayTLSConfig{
								Mode: gatewayapi.TLSModeTypePtr(gwapiv1b1.TLSModeTerminate),
								CertificateRefs: []gwapiv1b1.SecretObjectReference{
									{
										Name: gwapiv1b1.ObjectName("test-secret1"),
									},
									{
										Name: gwapiv1b1.ObjectName("test-secret2"),
									},
								},
							},
						},
						{
							Name:     "tls2",
							Port:     gwapiv1b1.PortNumber(int32(443)),
							Protocol: gwapiv1b1.HTTPSProtocolType,
							TLS: &gwapiv1b1.GatewayTLSConfig{
								Mode: gatewayapi.TLSModeTypePtr(gwapiv1b1.TLSModeTerminate),
								CertificateRefs: []gwapiv1b1.SecretObjectReference{
									{
										Name: gwapiv1b1.ObjectName("test-secret3"),
									},
									{
										Name: gwapiv1b1.ObjectName("test-secret4"),
									},
								},
							},
						},
					},
				},
			},
			secrets: []corev1.Secret{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Secret",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:            "test-secret1",
						Namespace:       "test-ns",
						ResourceVersion: "1",
					},
				},
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Secret",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:            "test-secret2",
						Namespace:       "test-ns",
						ResourceVersion: "1",
					},
				},
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Secret",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:            "test-secret3",
						Namespace:       "test-ns",
						ResourceVersion: "1",
					},
				},
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Secret",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:            "test-secret4",
						Namespace:       "test-ns",
						ResourceVersion: "1",
					},
				},
			},
		},
		{
			name: "gateway with one https listener and two same namespace secret",
			gw: &gwapiv1b1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-gw",
					Namespace: "test-ns",
				},
				Spec: gwapiv1b1.GatewaySpec{
					GatewayClassName: "test-gc",
					Listeners: []gwapiv1b1.Listener{
						{
							Name:     "tls",
							Port:     gwapiv1b1.PortNumber(int32(443)),
							Protocol: gwapiv1b1.HTTPSProtocolType,
							TLS: &gwapiv1b1.GatewayTLSConfig{
								Mode: gatewayapi.TLSModeTypePtr(gwapiv1b1.TLSModeTerminate),
								CertificateRefs: []gwapiv1b1.SecretObjectReference{
									{
										Name: gwapiv1b1.ObjectName("test-secret1"),
									},
									{
										Name: gwapiv1b1.ObjectName("test-secret2"),
									},
								},
							},
						},
					},
				},
			},
			secrets: []corev1.Secret{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Secret",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:            "test-secret1",
						Namespace:       "test-ns",
						ResourceVersion: "1",
					},
				},
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Secret",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:            "test-secret2",
						Namespace:       "test-ns",
						ResourceVersion: "1",
					},
				},
			},
		},
		{
			name: "gateway with one https listener and one different namespace secret",
			gw: &gwapiv1b1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-gw",
					Namespace: "test-ns",
				},
				Spec: gwapiv1b1.GatewaySpec{
					GatewayClassName: "test-gc",
					Listeners: []gwapiv1b1.Listener{
						{
							Name:     "tls",
							Port:     gwapiv1b1.PortNumber(int32(443)),
							Protocol: gwapiv1b1.HTTPSProtocolType,
							TLS: &gwapiv1b1.GatewayTLSConfig{
								Mode: gatewayapi.TLSModeTypePtr(gwapiv1b1.TLSModeTerminate),
								CertificateRefs: []gwapiv1b1.SecretObjectReference{
									{
										Name:      gwapiv1b1.ObjectName("test-secret"),
										Namespace: gatewayapi.NamespacePtr("test-ns2"),
									},
								},
							},
						},
					},
				},
			},
			secrets: []corev1.Secret{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Secret",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:            "test-secret",
						Namespace:       "test-ns2",
						ResourceVersion: "1",
					},
				},
			},
			refGrants: []gwapiv1a2.ReferenceGrant{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ReferenceGrant",
						APIVersion: gwapiv1a2.GroupVersion.Version,
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-refgrant",
						Namespace: "test-ns2",
					},
					Spec: gwapiv1a2.ReferenceGrantSpec{
						From: []gwapiv1a2.ReferenceGrantFrom{
							{
								Group:     gwapiv1a2.GroupName,
								Kind:      gatewayapi.KindGateway,
								Namespace: gwapiv1a2.Namespace("test-ns"),
							},
						},
						To: []gwapiv1a2.ReferenceGrantTo{
							{
								Group: corev1.GroupName,
								Kind:  gatewayapi.KindSecret,
							},
						},
					},
				},
			},
		},
		{
			name: "gateway with one https listener and one different namespace secret with specific referencegrant",
			gw: &gwapiv1b1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-gw",
					Namespace: "test-ns",
				},
				Spec: gwapiv1b1.GatewaySpec{
					GatewayClassName: "test-gc",
					Listeners: []gwapiv1b1.Listener{
						{
							Name:     "tls",
							Port:     gwapiv1b1.PortNumber(int32(443)),
							Protocol: gwapiv1b1.HTTPSProtocolType,
							TLS: &gwapiv1b1.GatewayTLSConfig{
								Mode: gatewayapi.TLSModeTypePtr(gwapiv1b1.TLSModeTerminate),
								CertificateRefs: []gwapiv1b1.SecretObjectReference{
									{
										Name:      gwapiv1b1.ObjectName("test-secret"),
										Namespace: gatewayapi.NamespacePtr("test-ns2"),
									},
								},
							},
						},
					},
				},
			},
			secrets: []corev1.Secret{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Secret",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:            "test-secret",
						Namespace:       "test-ns2",
						ResourceVersion: "1",
					},
				},
			},
			refGrants: []gwapiv1a2.ReferenceGrant{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ReferenceGrant",
						APIVersion: gwapiv1a2.GroupVersion.Version,
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-refgrant",
						Namespace: "test-ns2",
					},
					Spec: gwapiv1a2.ReferenceGrantSpec{
						From: []gwapiv1a2.ReferenceGrantFrom{
							{
								Group:     gwapiv1a2.GroupName,
								Kind:      gatewayapi.KindGateway,
								Namespace: gwapiv1a2.Namespace("test-ns"),
							},
						},
						To: []gwapiv1a2.ReferenceGrantTo{
							{
								Group: corev1.GroupName,
								Kind:  gatewayapi.KindSecret,
								Name:  gatewayapi.ObjectNamePtr("test-secret"),
							},
						},
					},
				},
			},
		},
	}

	// Create the reconciler.
	r := new(gatewayAPIReconciler)
	ctx := context.Background()

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			var objs []client.Object
			for j := range tc.secrets {
				objs = append(objs, &tc.secrets[j])
			}
			for k := range tc.refGrants {
				objs = append(objs, &tc.refGrants[k])
			}
			r.client = fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects(objs...).Build()
			secrets, refGrants, err := r.secretsAndRefGrantsForGateway(ctx, tc.gw)
			require.NoError(t, err)
			require.Equal(t, tc.secrets, secrets)
			require.Equal(t, tc.refGrants, refGrants)
		})
	}
}
