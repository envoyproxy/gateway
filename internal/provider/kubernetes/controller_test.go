// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
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
		logger := logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo)

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
			resourceTree := resource.NewResources()
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

func TestProcessEnvoyExtensionPolicyObjectRefs(t *testing.T) {
	testCases := []struct {
		name                 string
		envoyExtensionPolicy *egv1a1.EnvoyExtensionPolicy
		backend              *egv1a1.Backend
		referenceGrant       *gwapiv1b1.ReferenceGrant
		shouldBeAdded        bool
	}{
		{
			name: "valid envoy extension policy with proper ref grant to backend",
			envoyExtensionPolicy: &egv1a1.EnvoyExtensionPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns-1",
					Name:      "test-policy",
				},
				Spec: egv1a1.EnvoyExtensionPolicySpec{
					ExtProc: []egv1a1.ExtProc{
						{
							BackendCluster: egv1a1.BackendCluster{
								BackendRefs: []egv1a1.BackendRef{
									{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Namespace: gatewayapi.NamespacePtr("ns-2"),
											Name:      "test-backend",
											Kind:      gatewayapi.KindPtr(resource.KindBackend),
											Group:     gatewayapi.GroupPtr(egv1a1.GroupName),
										},
									},
								},
							},
						},
					},
				},
			},
			backend: &egv1a1.Backend{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns-2",
					Name:      "test-backend",
				},
			},
			referenceGrant: &gwapiv1b1.ReferenceGrant{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns-2",
					Name:      "test-grant",
				},
				Spec: gwapiv1b1.ReferenceGrantSpec{
					From: []gwapiv1b1.ReferenceGrantFrom{
						{
							Namespace: gwapiv1.Namespace("ns-1"),
							Kind:      gwapiv1.Kind(resource.KindEnvoyExtensionPolicy),
							Group:     gwapiv1.Group(egv1a1.GroupName),
						},
					},
					To: []gwapiv1b1.ReferenceGrantTo{
						{
							Name:  gatewayapi.ObjectNamePtr("test-backend"),
							Kind:  gwapiv1.Kind(resource.KindBackend),
							Group: gwapiv1.Group(egv1a1.GroupName),
						},
					},
				},
			},
			shouldBeAdded: true,
		},
		{
			name: "valid envoy extension policy with wrong from kind in ref grant to backend",
			envoyExtensionPolicy: &egv1a1.EnvoyExtensionPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns-1",
					Name:      "test-policy",
				},
				Spec: egv1a1.EnvoyExtensionPolicySpec{
					ExtProc: []egv1a1.ExtProc{
						{
							BackendCluster: egv1a1.BackendCluster{
								BackendRefs: []egv1a1.BackendRef{
									{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Namespace: gatewayapi.NamespacePtr("ns-2"),
											Name:      "test-backend",
											Kind:      gatewayapi.KindPtr(resource.KindBackend),
											Group:     gatewayapi.GroupPtr(egv1a1.GroupName),
										},
									},
								},
							},
						},
					},
				},
			},
			backend: &egv1a1.Backend{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns-2",
					Name:      "test-backend",
				},
			},
			referenceGrant: &gwapiv1b1.ReferenceGrant{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns-2",
					Name:      "test-grant",
				},
				Spec: gwapiv1b1.ReferenceGrantSpec{
					From: []gwapiv1b1.ReferenceGrantFrom{
						{
							Namespace: gwapiv1.Namespace("ns-1"),
							Kind:      gwapiv1.Kind(resource.KindHTTPRoute),
							Group:     gwapiv1.Group(gwapiv1.GroupName),
						},
					},
					To: []gwapiv1b1.ReferenceGrantTo{
						{
							Name:  gatewayapi.ObjectNamePtr("test-backend"),
							Kind:  gwapiv1.Kind(resource.KindBackend),
							Group: gwapiv1.Group(egv1a1.GroupName),
						},
					},
				},
			},
			shouldBeAdded: false,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		// Run the test cases.
		t.Run(tc.name, func(t *testing.T) {
			// Add objects referenced by test cases.
			objs := []client.Object{tc.envoyExtensionPolicy, tc.backend, tc.referenceGrant}
			r := setupReferenceGrantReconciler(objs)

			ctx := context.Background()
			resourceTree := resource.NewResources()
			resourceMap := newResourceMapping()

			err := r.processEnvoyExtensionPolicies(ctx, resourceTree, resourceMap)
			require.NoError(t, err)
			if tc.shouldBeAdded {
				require.Contains(t, resourceTree.ReferenceGrants, tc.referenceGrant)
			} else {
				require.NotContains(t, resourceTree.ReferenceGrants, tc.referenceGrant)
			}
		})
	}
}

func TestProcessSecurityPolicyObjectRefs(t *testing.T) {
	testCases := []struct {
		name           string
		securityPolicy *egv1a1.SecurityPolicy
		backend        *egv1a1.Backend
		referenceGrant *gwapiv1b1.ReferenceGrant
		shouldBeAdded  bool
	}{
		{
			name: "valid security policy with remote jwks proper ref grant to backend",
			securityPolicy: &egv1a1.SecurityPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns-1",
					Name:      "test-policy",
				},
				Spec: egv1a1.SecurityPolicySpec{
					JWT: &egv1a1.JWT{
						Providers: []egv1a1.JWTProvider{
							{
								RemoteJWKS: &egv1a1.RemoteJWKS{
									BackendCluster: egv1a1.BackendCluster{
										BackendRefs: []egv1a1.BackendRef{
											{
												BackendObjectReference: gwapiv1.BackendObjectReference{
													Namespace: gatewayapi.NamespacePtr("ns-2"),
													Name:      "test-backend",
													Kind:      gatewayapi.KindPtr(resource.KindBackend),
													Group:     gatewayapi.GroupPtr(egv1a1.GroupName),
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			backend: &egv1a1.Backend{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns-2",
					Name:      "test-backend",
				},
			},
			referenceGrant: &gwapiv1b1.ReferenceGrant{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns-2",
					Name:      "test-grant",
				},
				Spec: gwapiv1b1.ReferenceGrantSpec{
					From: []gwapiv1b1.ReferenceGrantFrom{
						{
							Namespace: gwapiv1.Namespace("ns-1"),
							Kind:      gwapiv1.Kind(resource.KindSecurityPolicy),
							Group:     gwapiv1.Group(egv1a1.GroupName),
						},
					},
					To: []gwapiv1b1.ReferenceGrantTo{
						{
							Name:  gatewayapi.ObjectNamePtr("test-backend"),
							Kind:  gwapiv1.Kind(resource.KindBackend),
							Group: gwapiv1.Group(egv1a1.GroupName),
						},
					},
				},
			},
			shouldBeAdded: true,
		},
		{
			name: "valid security policy with remote jwks wrong namespace ref grant to backend",
			securityPolicy: &egv1a1.SecurityPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns-1",
					Name:      "test-policy",
				},
				Spec: egv1a1.SecurityPolicySpec{
					JWT: &egv1a1.JWT{
						Providers: []egv1a1.JWTProvider{
							{
								RemoteJWKS: &egv1a1.RemoteJWKS{
									BackendCluster: egv1a1.BackendCluster{
										BackendRefs: []egv1a1.BackendRef{
											{
												BackendObjectReference: gwapiv1.BackendObjectReference{
													Namespace: gatewayapi.NamespacePtr("ns-2"),
													Name:      "test-backend",
													Kind:      gatewayapi.KindPtr(resource.KindBackend),
													Group:     gatewayapi.GroupPtr(egv1a1.GroupName),
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			backend: &egv1a1.Backend{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns-2",
					Name:      "test-backend",
				},
			},
			referenceGrant: &gwapiv1b1.ReferenceGrant{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns-2",
					Name:      "test-grant",
				},
				Spec: gwapiv1b1.ReferenceGrantSpec{
					From: []gwapiv1b1.ReferenceGrantFrom{
						{
							Namespace: gwapiv1.Namespace("ns-invalid"),
							Kind:      gwapiv1.Kind(resource.KindSecurityPolicy),
							Group:     gwapiv1.Group(egv1a1.GroupName),
						},
					},
					To: []gwapiv1b1.ReferenceGrantTo{
						{
							Name:  gatewayapi.ObjectNamePtr("test-backend"),
							Kind:  gwapiv1.Kind(resource.KindBackend),
							Group: gwapiv1.Group(egv1a1.GroupName),
						},
					},
				},
			},
			shouldBeAdded: false,
		},
		{
			name: "valid security policy with extAuth grpc proper ref grant to backend",
			securityPolicy: &egv1a1.SecurityPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns-1",
					Name:      "test-policy",
				},
				Spec: egv1a1.SecurityPolicySpec{
					ExtAuth: &egv1a1.ExtAuth{
						GRPC: &egv1a1.GRPCExtAuthService{
							BackendCluster: egv1a1.BackendCluster{
								BackendRefs: []egv1a1.BackendRef{
									{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Namespace: gatewayapi.NamespacePtr("ns-2"),
											Name:      "test-backend",
											Kind:      gatewayapi.KindPtr(resource.KindBackend),
											Group:     gatewayapi.GroupPtr(egv1a1.GroupName),
										},
									},
								},
							},
						},
					},
				},
			},
			backend: &egv1a1.Backend{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns-2",
					Name:      "test-backend",
				},
			},
			referenceGrant: &gwapiv1b1.ReferenceGrant{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns-2",
					Name:      "test-grant",
				},
				Spec: gwapiv1b1.ReferenceGrantSpec{
					From: []gwapiv1b1.ReferenceGrantFrom{
						{
							Namespace: gwapiv1.Namespace("ns-1"),
							Kind:      gwapiv1.Kind(resource.KindSecurityPolicy),
							Group:     gwapiv1.Group(egv1a1.GroupName),
						},
					},
					To: []gwapiv1b1.ReferenceGrantTo{
						{
							Name:  gatewayapi.ObjectNamePtr("test-backend"),
							Kind:  gwapiv1.Kind(resource.KindBackend),
							Group: gwapiv1.Group(egv1a1.GroupName),
						},
					},
				},
			},
			shouldBeAdded: true,
		},
		{
			name: "valid security policy with extAuth grpc proper ref grant to backend (deprecated field)",
			securityPolicy: &egv1a1.SecurityPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns-1",
					Name:      "test-policy",
				},
				Spec: egv1a1.SecurityPolicySpec{
					ExtAuth: &egv1a1.ExtAuth{
						GRPC: &egv1a1.GRPCExtAuthService{
							BackendCluster: egv1a1.BackendCluster{
								BackendRef: &gwapiv1.BackendObjectReference{
									Namespace: gatewayapi.NamespacePtr("ns-2"),
									Name:      "test-backend",
									Kind:      gatewayapi.KindPtr(resource.KindBackend),
									Group:     gatewayapi.GroupPtr(egv1a1.GroupName),
								},
							},
						},
					},
				},
			},
			backend: &egv1a1.Backend{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns-2",
					Name:      "test-backend",
				},
			},
			referenceGrant: &gwapiv1b1.ReferenceGrant{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns-2",
					Name:      "test-grant",
				},
				Spec: gwapiv1b1.ReferenceGrantSpec{
					From: []gwapiv1b1.ReferenceGrantFrom{
						{
							Namespace: gwapiv1.Namespace("ns-1"),
							Kind:      gwapiv1.Kind(resource.KindSecurityPolicy),
							Group:     gwapiv1.Group(egv1a1.GroupName),
						},
					},
					To: []gwapiv1b1.ReferenceGrantTo{
						{
							Name:  gatewayapi.ObjectNamePtr("test-backend"),
							Kind:  gwapiv1.Kind(resource.KindBackend),
							Group: gwapiv1.Group(egv1a1.GroupName),
						},
					},
				},
			},
			shouldBeAdded: true,
		},
		{
			name: "valid security policy with extAuth grpc wrong namespace ref grant to backend",
			securityPolicy: &egv1a1.SecurityPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns-1",
					Name:      "test-policy",
				},
				Spec: egv1a1.SecurityPolicySpec{
					ExtAuth: &egv1a1.ExtAuth{
						GRPC: &egv1a1.GRPCExtAuthService{
							BackendCluster: egv1a1.BackendCluster{
								BackendRefs: []egv1a1.BackendRef{
									{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Namespace: gatewayapi.NamespacePtr("ns-2"),
											Name:      "test-backend",
											Kind:      gatewayapi.KindPtr(resource.KindBackend),
											Group:     gatewayapi.GroupPtr(egv1a1.GroupName),
										},
									},
								},
							},
						},
					},
				},
			},
			backend: &egv1a1.Backend{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns-2",
					Name:      "test-backend",
				},
			},
			referenceGrant: &gwapiv1b1.ReferenceGrant{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns-2",
					Name:      "test-grant",
				},
				Spec: gwapiv1b1.ReferenceGrantSpec{
					From: []gwapiv1b1.ReferenceGrantFrom{
						{
							Namespace: gwapiv1.Namespace("ns-invalid"),
							Kind:      gwapiv1.Kind(resource.KindSecurityPolicy),
							Group:     gwapiv1.Group(egv1a1.GroupName),
						},
					},
					To: []gwapiv1b1.ReferenceGrantTo{
						{
							Name:  gatewayapi.ObjectNamePtr("test-backend"),
							Kind:  gwapiv1.Kind(resource.KindBackend),
							Group: gwapiv1.Group(egv1a1.GroupName),
						},
					},
				},
			},
			shouldBeAdded: false,
		},
		{
			name: "valid security policy with extAuth http proper ref grant to backend",
			securityPolicy: &egv1a1.SecurityPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns-1",
					Name:      "test-policy",
				},
				Spec: egv1a1.SecurityPolicySpec{
					ExtAuth: &egv1a1.ExtAuth{
						HTTP: &egv1a1.HTTPExtAuthService{
							BackendCluster: egv1a1.BackendCluster{
								BackendRefs: []egv1a1.BackendRef{
									{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Namespace: gatewayapi.NamespacePtr("ns-2"),
											Name:      "test-backend",
											Kind:      gatewayapi.KindPtr(resource.KindBackend),
											Group:     gatewayapi.GroupPtr(egv1a1.GroupName),
										},
									},
								},
							},
						},
					},
				},
			},
			backend: &egv1a1.Backend{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns-2",
					Name:      "test-backend",
				},
			},
			referenceGrant: &gwapiv1b1.ReferenceGrant{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns-2",
					Name:      "test-grant",
				},
				Spec: gwapiv1b1.ReferenceGrantSpec{
					From: []gwapiv1b1.ReferenceGrantFrom{
						{
							Namespace: gwapiv1.Namespace("ns-1"),
							Kind:      gwapiv1.Kind(resource.KindSecurityPolicy),
							Group:     gwapiv1.Group(egv1a1.GroupName),
						},
					},
					To: []gwapiv1b1.ReferenceGrantTo{
						{
							Name:  gatewayapi.ObjectNamePtr("test-backend"),
							Kind:  gwapiv1.Kind(resource.KindBackend),
							Group: gwapiv1.Group(egv1a1.GroupName),
						},
					},
				},
			},
			shouldBeAdded: true,
		},
		{
			name: "valid security policy with extAuth http proper ref grant to backend (deprecated field)",
			securityPolicy: &egv1a1.SecurityPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns-1",
					Name:      "test-policy",
				},
				Spec: egv1a1.SecurityPolicySpec{
					ExtAuth: &egv1a1.ExtAuth{
						HTTP: &egv1a1.HTTPExtAuthService{
							BackendCluster: egv1a1.BackendCluster{
								BackendRef: &gwapiv1.BackendObjectReference{
									Namespace: gatewayapi.NamespacePtr("ns-2"),
									Name:      "test-backend",
									Kind:      gatewayapi.KindPtr(resource.KindBackend),
									Group:     gatewayapi.GroupPtr(egv1a1.GroupName),
								},
							},
						},
					},
				},
			},
			backend: &egv1a1.Backend{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns-2",
					Name:      "test-backend",
				},
			},
			referenceGrant: &gwapiv1b1.ReferenceGrant{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns-2",
					Name:      "test-grant",
				},
				Spec: gwapiv1b1.ReferenceGrantSpec{
					From: []gwapiv1b1.ReferenceGrantFrom{
						{
							Namespace: gwapiv1.Namespace("ns-1"),
							Kind:      gwapiv1.Kind(resource.KindSecurityPolicy),
							Group:     gwapiv1.Group(egv1a1.GroupName),
						},
					},
					To: []gwapiv1b1.ReferenceGrantTo{
						{
							Name:  gatewayapi.ObjectNamePtr("test-backend"),
							Kind:  gwapiv1.Kind(resource.KindBackend),
							Group: gwapiv1.Group(egv1a1.GroupName),
						},
					},
				},
			},
			shouldBeAdded: true,
		},
		{
			name: "valid security policy with extAuth http wrong namespace ref grant to backend",
			securityPolicy: &egv1a1.SecurityPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns-1",
					Name:      "test-policy",
				},
				Spec: egv1a1.SecurityPolicySpec{
					ExtAuth: &egv1a1.ExtAuth{
						HTTP: &egv1a1.HTTPExtAuthService{
							BackendCluster: egv1a1.BackendCluster{
								BackendRefs: []egv1a1.BackendRef{
									{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Namespace: gatewayapi.NamespacePtr("ns-2"),
											Name:      "test-backend",
											Kind:      gatewayapi.KindPtr(resource.KindBackend),
											Group:     gatewayapi.GroupPtr(egv1a1.GroupName),
										},
									},
								},
							},
						},
					},
				},
			},
			backend: &egv1a1.Backend{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns-2",
					Name:      "test-backend",
				},
			},
			referenceGrant: &gwapiv1b1.ReferenceGrant{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns-2",
					Name:      "test-grant",
				},
				Spec: gwapiv1b1.ReferenceGrantSpec{
					From: []gwapiv1b1.ReferenceGrantFrom{
						{
							Namespace: gwapiv1.Namespace("ns-invalid"),
							Kind:      gwapiv1.Kind(resource.KindSecurityPolicy),
							Group:     gwapiv1.Group(egv1a1.GroupName),
						},
					},
					To: []gwapiv1b1.ReferenceGrantTo{
						{
							Name:  gatewayapi.ObjectNamePtr("test-backend"),
							Kind:  gwapiv1.Kind(resource.KindBackend),
							Group: gwapiv1.Group(egv1a1.GroupName),
						},
					},
				},
			},
			shouldBeAdded: false,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		// Run the test cases.
		t.Run(tc.name, func(t *testing.T) {
			// Add objects referenced by test cases.
			objs := []client.Object{tc.securityPolicy, tc.backend, tc.referenceGrant}
			r := setupReferenceGrantReconciler(objs)

			ctx := context.Background()
			resourceTree := resource.NewResources()
			resourceMap := newResourceMapping()

			err := r.processSecurityPolicies(ctx, resourceTree, resourceMap)
			require.NoError(t, err)
			if tc.shouldBeAdded {
				require.Contains(t, resourceTree.ReferenceGrants, tc.referenceGrant)
			} else {
				require.NotContains(t, resourceTree.ReferenceGrants, tc.referenceGrant)
			}
		})
	}
}

func setupReferenceGrantReconciler(objs []client.Object) *gatewayAPIReconciler {
	logger := logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo)

	r := &gatewayAPIReconciler{
		log:             logger,
		classController: "some-gateway-class",
	}

	r.client = fakeclient.NewClientBuilder().
		WithScheme(envoygateway.GetScheme()).
		WithObjects(objs...).
		WithIndex(&gwapiv1b1.ReferenceGrant{}, targetRefGrantRouteIndex, getReferenceGrantIndexerFunc).
		Build()
	return r
}
