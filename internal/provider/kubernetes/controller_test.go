// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/proxy"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/provider/kubernetes/test"
	"github.com/envoyproxy/gateway/internal/utils"
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

func TestIsCustomBackendResource(t *testing.T) {
	testCases := []struct {
		name           string
		extBackendGVKs []schema.GroupVersionKind
		group          *gwapiv1.Group
		kind           string
		expected       bool
	}{
		{
			name:           "no extension backend GVKs configured",
			extBackendGVKs: []schema.GroupVersionKind{},
			group:          ptr.To(gwapiv1.Group("storage.example.io")),
			kind:           "S3Backend",
			expected:       false,
		},
		{
			name: "matching group and kind",
			extBackendGVKs: []schema.GroupVersionKind{
				{Group: "storage.example.io", Version: "v1alpha1", Kind: "S3Backend"},
				{Group: "compute.example.io", Version: "v1alpha1", Kind: "LambdaBackend"},
			},
			group:    ptr.To(gwapiv1.Group("storage.example.io")),
			kind:     "S3Backend",
			expected: true,
		},
		{
			name: "matching kind but different group",
			extBackendGVKs: []schema.GroupVersionKind{
				{Group: "storage.example.io", Version: "v1alpha1", Kind: "S3Backend"},
			},
			group:    ptr.To(gwapiv1.Group("compute.example.io")),
			kind:     "S3Backend",
			expected: false,
		},
		{
			name: "matching group but different kind",
			extBackendGVKs: []schema.GroupVersionKind{
				{Group: "storage.example.io", Version: "v1alpha1", Kind: "S3Backend"},
			},
			group:    ptr.To(gwapiv1.Group("storage.example.io")),
			kind:     "LambdaBackend",
			expected: false,
		},
		{
			name: "nil group with empty string group in GVK",
			extBackendGVKs: []schema.GroupVersionKind{
				{Group: "", Version: "v1", Kind: "Service"},
			},
			group:    nil,
			kind:     "Service",
			expected: true,
		},
		{
			name: "nil group with non-empty group in GVK",
			extBackendGVKs: []schema.GroupVersionKind{
				{Group: "storage.example.io", Version: "v1alpha1", Kind: "S3Backend"},
			},
			group:    nil,
			kind:     "S3Backend",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := &gatewayAPIReconciler{
				extBackendGVKs: tc.extBackendGVKs,
			}
			result := r.isCustomBackendResource(tc.group, tc.kind)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestProcessBackendRefsWithCustomBackends(t *testing.T) {
	ctx := context.Background()

	// Create test custom backend resources
	s3Backend := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "storage.example.io/v1alpha1",
			"kind":       "S3Backend",
			"metadata": map[string]interface{}{
				"name":      "s3-backend",
				"namespace": "default",
			},
			"spec": map[string]interface{}{
				"bucket": "my-s3-bucket",
				"region": "us-west-2",
			},
		},
	}

	lambdaBackend := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "compute.example.io/v1alpha1",
			"kind":       "LambdaBackend",
			"metadata": map[string]interface{}{
				"name":      "lambda-backend",
				"namespace": "default",
			},
			"spec": map[string]interface{}{
				"functionName": "my-function",
				"region":       "us-west-2",
			},
		},
	}

	testCases := []struct {
		name                    string
		extBackendGVKs          []schema.GroupVersionKind
		backendRefs             []gwapiv1.BackendObjectReference
		existingExtFilters      map[utils.NamespacedNameWithGroupKind]unstructured.Unstructured
		expectedExtFiltersCount int
		expectedNamespaces      []string
	}{
		{
			name: "process custom S3 backend",
			extBackendGVKs: []schema.GroupVersionKind{
				{Group: "storage.example.io", Version: "v1alpha1", Kind: "S3Backend"},
			},
			backendRefs: []gwapiv1.BackendObjectReference{
				{
					Group:     ptr.To(gwapiv1.Group("storage.example.io")),
					Kind:      ptr.To(gwapiv1.Kind("S3Backend")),
					Name:      "s3-backend",
					Namespace: ptr.To(gwapiv1.Namespace("default")),
				},
			},
			existingExtFilters: map[utils.NamespacedNameWithGroupKind]unstructured.Unstructured{
				{
					NamespacedName: types.NamespacedName{Namespace: "default", Name: "s3-backend"},
					GroupKind:      schema.GroupKind{Group: "storage.example.io", Kind: "S3Backend"},
				}: *s3Backend,
			},
			expectedExtFiltersCount: 1,
			expectedNamespaces:      []string{"default"},
		},
		{
			name: "process multiple custom backends",
			extBackendGVKs: []schema.GroupVersionKind{
				{Group: "storage.example.io", Version: "v1alpha1", Kind: "S3Backend"},
				{Group: "compute.example.io", Version: "v1alpha1", Kind: "LambdaBackend"},
			},
			backendRefs: []gwapiv1.BackendObjectReference{
				{
					Group:     ptr.To(gwapiv1.Group("storage.example.io")),
					Kind:      ptr.To(gwapiv1.Kind("S3Backend")),
					Name:      "s3-backend",
					Namespace: ptr.To(gwapiv1.Namespace("default")),
				},
				{
					Group:     ptr.To(gwapiv1.Group("compute.example.io")),
					Kind:      ptr.To(gwapiv1.Kind("LambdaBackend")),
					Name:      "lambda-backend",
					Namespace: ptr.To(gwapiv1.Namespace("default")),
				},
			},
			existingExtFilters: map[utils.NamespacedNameWithGroupKind]unstructured.Unstructured{
				{
					NamespacedName: types.NamespacedName{Namespace: "default", Name: "s3-backend"},
					GroupKind:      schema.GroupKind{Group: "storage.example.io", Kind: "S3Backend"},
				}: *s3Backend,
				{
					NamespacedName: types.NamespacedName{Namespace: "default", Name: "lambda-backend"},
					GroupKind:      schema.GroupKind{Group: "compute.example.io", Kind: "LambdaBackend"},
				}: *lambdaBackend,
			},
			expectedExtFiltersCount: 2,
			expectedNamespaces:      []string{"default"},
		},
		{
			name: "skip non-custom backends",
			extBackendGVKs: []schema.GroupVersionKind{
				{Group: "storage.example.io", Version: "v1alpha1", Kind: "S3Backend"},
			},
			backendRefs: []gwapiv1.BackendObjectReference{
				{
					// Standard Service backend - should be skipped
					Kind:      ptr.To(gwapiv1.Kind("Service")),
					Name:      "my-service",
					Namespace: ptr.To(gwapiv1.Namespace("default")),
				},
			},
			existingExtFilters:      map[utils.NamespacedNameWithGroupKind]unstructured.Unstructured{},
			expectedExtFiltersCount: 0,
			expectedNamespaces:      []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create fake client
			fakeClient := fakeclient.NewClientBuilder().Build()

			// Create reconciler with test configuration
			r := &gatewayAPIReconciler{
				extBackendGVKs: tc.extBackendGVKs,
				log:            logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo),
				client:         fakeClient,
			}

			// Create resource mappings
			resourceMappings := &resourceMappings{
				allAssociatedBackendRefs:                make(map[utils.NamespacedNameWithGroupKind]gwapiv1.BackendObjectReference),
				allAssociatedNamespaces:                 sets.New[string](),
				allAssociatedBackendRefExtensionFilters: sets.New[utils.NamespacedNameWithGroupKind](),
				extensionRefFilters:                     tc.existingExtFilters,
			}

			// Add backend refs to the mapping
			for _, backendRef := range tc.backendRefs {
				resourceMappings.insertBackendRef(backendRef)
			}

			// Create empty resource tree
			gwcResource := &resource.Resources{
				ExtensionRefFilters: []unstructured.Unstructured{},
			}

			// Call the function under test
			require.NoError(t, r.processBackendRefs(ctx, gwcResource, resourceMappings))
			// Compare the results
			require.Len(t, gwcResource.ExtensionRefFilters, tc.expectedExtFiltersCount)

			for _, expectedNS := range tc.expectedNamespaces {
				require.True(t, resourceMappings.allAssociatedNamespaces.Has(expectedNS))
			}
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
		name                 string
		gc                   *gwapiv1.GatewayClass
		ep                   *egv1a1.EnvoyProxy
		gatewayNamespaceMode bool
		expected             bool
		expectedError        string
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
		{
			name: "incompatible configuration: merged gateways with gateway namespace mode",
			gc: &gwapiv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-merged-gw",
				},
				Spec: gwapiv1.GatewayClassSpec{
					ControllerName: gcCtrlName,
					ParametersRef: &gwapiv1.ParametersReference{
						Group:     gwapiv1.Group(egv1a1.GroupVersion.Group),
						Kind:      gwapiv1.Kind(egv1a1.KindEnvoyProxy),
						Name:      "test-merge-gw",
						Namespace: gatewayapi.NamespacePtr(config.DefaultNamespace),
					},
				},
			},
			ep: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: config.DefaultNamespace,
					Name:      "test-merge-gw",
				},
				Spec: egv1a1.EnvoyProxySpec{
					MergeGateways: ptr.To(true),
				},
			},
			gatewayNamespaceMode: true,
			expected:             false,
			expectedError:        "using Merged Gateways with Gateway Namespace Mode is not supported",
		},
		{
			name: "valid merged gateways enabled configuration",
			gc: &gwapiv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-merge-gw",
				},
				Spec: gwapiv1.GatewayClassSpec{
					ControllerName: gcCtrlName,
					ParametersRef: &gwapiv1.ParametersReference{
						Group:     gwapiv1.Group(egv1a1.GroupVersion.Group),
						Kind:      gwapiv1.Kind(egv1a1.KindEnvoyProxy),
						Name:      "test-merge-gw",
						Namespace: gatewayapi.NamespacePtr(config.DefaultNamespace),
					},
				},
			},
			ep: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: config.DefaultNamespace,
					Name:      "test-merge-gw",
				},
				Spec: egv1a1.EnvoyProxySpec{
					MergeGateways: ptr.To(true),
				},
			},
			gatewayNamespaceMode: false,
			expected:             true,
		},
		{
			name: "valid gateway namespace mode enabled configuration",
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
			gatewayNamespaceMode: true,
			expected:             true,
		},
	}

	for i := range testCases {
		tc := testCases[i]

		// Create the reconciler.
		logger := logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo)

		r := &gatewayAPIReconciler{
			log:                  logger,
			classController:      gcCtrlName,
			namespace:            config.DefaultNamespace,
			gatewayNamespaceMode: tc.gatewayNamespaceMode,
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
				if tc.expectedError != "" {
					require.Contains(t, err.Error(), tc.expectedError)
				}
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

func TestProcessServiceClusterForGatewayClass(t *testing.T) {
	testCases := []struct {
		name            string
		gatewayClass    *gwapiv1.GatewayClass
		envoyProxy      *egv1a1.EnvoyProxy
		expectedSvcName string
	}{
		{
			name: "when merged gateways and no hardcoded svc name is used",
			gatewayClass: &gwapiv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "merged-gc",
				},
			},
			envoyProxy:      nil,
			expectedSvcName: proxy.ExpectedResourceHashedName("merged-gc"),
		},
		{
			name: "when merged gateways and a hardcoded svc name is used",
			gatewayClass: &gwapiv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "merged-gc",
				},
			},
			envoyProxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "merged-gc",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.EnvoyProxyProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyService: &egv1a1.KubernetesServiceSpec{
								Name: ptr.To("merged-gc-svc"),
							},
						},
					},
				},
			},
			expectedSvcName: "merged-gc-svc",
		},
	}

	for i := range testCases {
		tc := testCases[i]
		// Run the test cases.
		t.Run(tc.name, func(t *testing.T) {
			logger := logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo)
			resourceMap := newResourceMapping()

			r := newGatewayAPIReconciler(logger)
			r.namespace = "envoy-gateway-system"

			r.processServiceClusterForGatewayClass(tc.envoyProxy, tc.gatewayClass, resourceMap)

			expectedRef := gwapiv1.BackendObjectReference{
				Kind:      ptr.To(gwapiv1.Kind(resource.KindService)),
				Namespace: gatewayapi.NamespacePtr(r.namespace),
				Name:      gwapiv1.ObjectName(tc.expectedSvcName),
			}
			key := backendRefKey(&expectedRef)
			require.Contains(t, resourceMap.allAssociatedBackendRefs, key)
			require.Equal(t, expectedRef, resourceMap.allAssociatedBackendRefs[key])
		})
	}
}

func TestProcessServiceClusterForGateway(t *testing.T) {
	testCases := []struct {
		name                   string
		gateway                *gwapiv1.Gateway
		envoyProxy             *egv1a1.EnvoyProxy
		gatewayClassEnvoyProxy *egv1a1.EnvoyProxy
		gatewayNamespacedMode  bool
		expectedSvcName        string
		expectedSvcNamespace   string
	}{
		{
			name: "no gateway namespaced mode with no hardcoded service name",
			gateway: &gwapiv1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-gateway",
					Namespace: "app-namespace",
				},
			},
			envoyProxy:            nil,
			gatewayNamespacedMode: false,
			expectedSvcName:       "",
			expectedSvcNamespace:  "",
		},
		{
			name: "no gateway namespaced mode with hardcoded service name",
			gateway: &gwapiv1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-gateway",
					Namespace: "app-namespace",
				},
			},
			envoyProxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-gateway",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.EnvoyProxyProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyService: &egv1a1.KubernetesServiceSpec{
								Name: ptr.To("my-gateway-svc"),
							},
						},
					},
				},
			},
			gatewayNamespacedMode: false,
			expectedSvcName:       "my-gateway-svc",
			expectedSvcNamespace:  "",
		},
		{
			name: "gateway namespaced mode with no hardcoded service name",
			gateway: &gwapiv1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-gateway",
					Namespace: "app-namespace",
				},
			},
			envoyProxy:            nil,
			gatewayNamespacedMode: true,
			expectedSvcName:       "my-gateway",
			expectedSvcNamespace:  "app-namespace",
		},
		{
			name: "gateway namespaced mode with hardcoded service name",
			gateway: &gwapiv1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-gateway",
					Namespace: "app-namespace",
				},
			},
			envoyProxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-gateway",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.EnvoyProxyProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyService: &egv1a1.KubernetesServiceSpec{
								Name: ptr.To("my-gateway-svc"),
							},
						},
					},
				},
			},
			gatewayNamespacedMode: true,
			expectedSvcName:       "my-gateway-svc",
			expectedSvcNamespace:  "app-namespace",
		},
		{
			name: "no gateway namespaced mode with no hardcoded service name attached gatewayclass",
			gateway: &gwapiv1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-gateway",
					Namespace: "app-namespace",
				},
			},
			envoyProxy: nil,
			gatewayClassEnvoyProxy: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-gateway",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.EnvoyProxyProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyService: &egv1a1.KubernetesServiceSpec{
								Name: ptr.To("my-gateway-svc"),
							},
						},
					},
				},
			},
			gatewayNamespacedMode: false,
			expectedSvcName:       "my-gateway-svc",
			expectedSvcNamespace:  "",
		},
	}

	for i := range testCases {
		tc := testCases[i]
		// Run the test cases.
		t.Run(tc.name, func(t *testing.T) {
			logger := logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo)
			resourceMap := newResourceMapping()

			r := newGatewayAPIReconciler(logger)
			r.namespace = "envoy-gateway-system"
			r.gatewayNamespaceMode = tc.gatewayNamespacedMode

			if tc.expectedSvcNamespace == "" {
				tc.expectedSvcNamespace = r.namespace
			}

			if tc.expectedSvcName == "" {
				tc.expectedSvcName = proxy.ExpectedResourceHashedName(utils.NamespacedName(tc.gateway).String())
			}

			if tc.envoyProxy == nil && tc.gatewayClassEnvoyProxy != nil {
				tc.envoyProxy = tc.gatewayClassEnvoyProxy
			}

			r.processServiceClusterForGateway(tc.envoyProxy, tc.gateway, resourceMap)

			expectedRef := gwapiv1.BackendObjectReference{
				Kind:      ptr.To(gwapiv1.Kind(resource.KindService)),
				Namespace: gatewayapi.NamespacePtr(tc.expectedSvcNamespace),
				Name:      gwapiv1.ObjectName(tc.expectedSvcName),
			}
			key := backendRefKey(&expectedRef)
			require.Contains(t, resourceMap.allAssociatedBackendRefs, key)
			require.Equal(t, expectedRef, resourceMap.allAssociatedBackendRefs[key])
		})
	}
}

func newGatewayAPIReconciler(logger logging.Logger) *gatewayAPIReconciler {
	return &gatewayAPIReconciler{
		log:              logger,
		classController:  "some-gateway-class",
		backendCRDExists: true,
		envoyGateway: &egv1a1.EnvoyGateway{
			EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
				ExtensionAPIs: &egv1a1.ExtensionAPISettings{
					EnableBackend: true,
				},
			},
		},
	}
}

func TestProcessBackendRefDeduplicatesLogicalBackend(t *testing.T) {
	logger := logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo)
	r := newGatewayAPIReconciler(logger)
	resourceTree := resource.NewResources()
	resourceMap := newResourceMapping()

	backendRef := gwapiv1.BackendObjectReference{
		Namespace: gatewayapi.NamespacePtr("default"),
		Name:      "svc",
	}

	require.NoError(t, r.processBackendRef(t.Context(), resourceMap, resourceTree, resource.KindHTTPRoute, "default", "route-a", backendRef))
	require.NoError(t, r.processBackendRef(t.Context(), resourceMap, resourceTree, resource.KindHTTPRoute, "default", "route-b", backendRef))

	require.Len(t, resourceMap.allAssociatedBackendRefs, 1)
}

func TestProcessBackendRefs(t *testing.T) {
	ns := "default"
	ctb := test.GetClusterTrustBundle("fake-ctb")
	secret := test.GetSecret(types.NamespacedName{Namespace: ns, Name: "fake-secret"})
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      "fake-cm",
		},
		Data: map[string]string{
			"ca.crt": "fake-ca-cert",
		},
	}

	testCases := []struct {
		name                   string
		backend                *egv1a1.Backend
		ctpShouldBeAdded       bool
		secretShouldBeAdded    bool
		configmapShouldBeAdded bool
	}{
		{
			name: "DynamicResolver with ClusterTrustBundle",
			backend: &egv1a1.Backend{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: ns,
					Name:      "test-backend",
				},
				Spec: egv1a1.BackendSpec{
					Type: ptr.To(egv1a1.BackendTypeDynamicResolver),
					TLS: &egv1a1.BackendTLSSettings{
						CACertificateRefs: []gwapiv1.LocalObjectReference{
							{
								Kind: gwapiv1.Kind("ClusterTrustBundle"),
								Name: gwapiv1.ObjectName("fake-ctb"),
							},
						},
					},
				},
			},
			ctpShouldBeAdded: true,
		},
		{
			name: "DynamicResolver with ConfigMap",
			backend: &egv1a1.Backend{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: ns,
					Name:      "test-backend",
				},
				Spec: egv1a1.BackendSpec{
					Type: ptr.To(egv1a1.BackendTypeDynamicResolver),
					TLS: &egv1a1.BackendTLSSettings{
						CACertificateRefs: []gwapiv1.LocalObjectReference{
							{
								Kind: gwapiv1.Kind("ConfigMap"),
								Name: gwapiv1.ObjectName("fake-cm"),
							},
						},
					},
				},
			},
			configmapShouldBeAdded: true,
		},
		{
			name: "DynamicResolver with Secret",
			backend: &egv1a1.Backend{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: ns,
					Name:      "test-backend",
				},
				Spec: egv1a1.BackendSpec{
					Type: ptr.To(egv1a1.BackendTypeDynamicResolver),
					TLS: &egv1a1.BackendTLSSettings{
						CACertificateRefs: []gwapiv1.LocalObjectReference{
							{
								Kind: gwapiv1.Kind("Secret"),
								Name: gwapiv1.ObjectName("fake-secret"),
							},
						},
					},
				},
			},
			secretShouldBeAdded: true,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		// Run the test cases.
		t.Run(tc.name, func(t *testing.T) {
			// Add objects referenced by test cases.
			objs := []client.Object{tc.backend, ctb, secret, cm}
			logger := logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo)

			r := newGatewayAPIReconciler(logger)
			r.client = fakeclient.NewClientBuilder().
				WithScheme(envoygateway.GetScheme()).
				WithObjects(objs...).
				Build()

			resourceTree := resource.NewResources()
			resourceMap := newResourceMapping()
			backend := tc.backend
			resourceMap.insertBackendRef(gwapiv1.BackendObjectReference{
				Kind:      gatewayapi.KindPtr(resource.KindBackend),
				Namespace: gatewayapi.NamespacePtr(backend.Namespace),
				Name:      gwapiv1.ObjectName(backend.Name),
			})

			require.NoError(t, r.processBackendRefs(t.Context(), resourceTree, resourceMap))
			if tc.ctpShouldBeAdded {
				require.Contains(t, resourceTree.ClusterTrustBundles, ctb)
			} else {
				require.NotContains(t, resourceTree.ClusterTrustBundles, ctb)
			}

			if tc.secretShouldBeAdded {
				require.Contains(t, resourceTree.Secrets, secret)
			} else {
				require.NotContains(t, resourceTree.Secrets, secret)
			}

			if tc.configmapShouldBeAdded {
				require.Contains(t, resourceTree.ConfigMaps, cm)
			} else {
				require.NotContains(t, resourceTree.ConfigMaps, cm)
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

func TestIsTransientError(t *testing.T) {
	serverTimeoutErr := kerrors.NewServerTimeout(
		schema.GroupResource{Group: "core", Resource: "pods"}, "list", 10)
	timeoutErr := kerrors.NewTimeoutError("request timeout", 1)
	wrappedTooManyRequestsErr := fmt.Errorf("wrapping: %w", kerrors.NewTooManyRequests("too many requests", 1))
	serviceUnavailableErr := kerrors.NewServiceUnavailable("service unavailable")
	badRequestErr := kerrors.NewBadRequest("bad request")

	// new test errors for context
	canceledErr := context.Canceled
	deadlineExceededErr := context.DeadlineExceeded
	wrappedCanceledErr := fmt.Errorf("wrapped: %w", context.Canceled)
	wrappedDeadlineExceededErr := fmt.Errorf("wrapped: %w", context.DeadlineExceeded)

	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{"ServerTimeout", serverTimeoutErr, true},
		{"Timeout", timeoutErr, true},
		{"TooManyRequests", wrappedTooManyRequestsErr, true},
		{"ServiceUnavailable", serviceUnavailableErr, true},
		{"BadRequest", badRequestErr, false},
		{"NilError", nil, false},
		{"ContextCanceled", canceledErr, true},
		{"ContextDeadlineExceeded", deadlineExceededErr, true},
		{"WrappedContextCanceled", wrappedCanceledErr, true},
		{"WrappedContextDeadlineExceeded", wrappedDeadlineExceededErr, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := isTransientError(tc.err)
			require.Equal(t, tc.expected, actual)
		})
	}
}
