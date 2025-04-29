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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/logging"
)

func TestProcessExtensionPolicies(t *testing.T) {
	testCases := []struct {
		name                  string
		extensionPolicies     []*unstructured.Unstructured
		extensionPolicyGroups []schema.GroupVersionKind
		errorExpected         bool
	}{
		{
			name: "valid extension policy with targetRef",
			extensionPolicies: []*unstructured.Unstructured{
				{
					Object: map[string]any{
						"apiVersion": "gateway.example.io/v1alpha1",
						"kind":       "Foo",
						"metadata": map[string]any{
							"name":      "test",
							"namespace": "test",
						},
						"spec": map[string]any{
							"targetRef": map[string]any{
								"group": "gateway.networking.k8s.io",
								"kind":  "Gateway",
								"name":  "test",
							},
							"data": "some data",
						},
					},
				},
			},
			extensionPolicyGroups: []schema.GroupVersionKind{
				{
					Group:   "gateway.example.io",
					Version: "v1alpha1",
					Kind:    "Foo",
				},
			},
		},
		{
			name: "valid extension policy with targetRefs",
			extensionPolicies: []*unstructured.Unstructured{
				{
					Object: map[string]any{
						"apiVersion": "gateway.example.io/v1alpha1",
						"kind":       "Foo",
						"metadata": map[string]any{
							"name":      "test",
							"namespace": "test",
						},
						"spec": map[string]any{
							"targetRefs": []any{
								map[string]any{
									"group": "gateway.networking.k8s.io",
									"kind":  "Gateway",
									"name":  "test",
								},
							},
							"data": "some data",
						},
					},
				},
			},
			extensionPolicyGroups: []schema.GroupVersionKind{
				{
					Group:   "gateway.example.io",
					Version: "v1alpha1",
					Kind:    "Foo",
				},
			},
		},
		{
			name: "invalid extension policy - no target",
			extensionPolicies: []*unstructured.Unstructured{
				{
					Object: map[string]any{
						"apiVersion": "gateway.example.io/v1alpha1",
						"kind":       "Foo",
						"metadata": map[string]any{
							"name":      "test",
							"namespace": "test",
						},
						"spec": map[string]any{
							"data": "some data",
						},
					},
				},
			},
			extensionPolicyGroups: []schema.GroupVersionKind{
				{
					Group:   "gateway.example.io",
					Version: "v1alpha1",
					Kind:    "Foo",
				},
			},
			errorExpected: true,
		},
		{
			name: "invalid extension policy - no spec",
			extensionPolicies: []*unstructured.Unstructured{
				{
					Object: map[string]any{
						"apiVersion": "gateway.example.io/v1alpha1",
						"kind":       "Foo",
						"metadata": map[string]any{
							"name":      "test",
							"namespace": "test",
						},
					},
				},
			},
			extensionPolicyGroups: []schema.GroupVersionKind{
				{
					Group:   "gateway.example.io",
					Version: "v1alpha1",
					Kind:    "Foo",
				},
			},
			errorExpected: true,
		},
		{
			name: "multiple extension policy types with targetRefs",
			extensionPolicies: []*unstructured.Unstructured{
				{
					Object: map[string]any{
						"apiVersion": "gateway.example.io/v1alpha1",
						"kind":       "Bar",
						"metadata": map[string]any{
							"name":      "test",
							"namespace": "test",
						},
						"spec": map[string]any{
							"targetRefs": []any{
								map[string]any{
									"group": "gateway.networking.k8s.io",
									"kind":  "Gateway",
									"name":  "test",
								},
							},
							"data": "some data",
						},
					},
				},
				{
					Object: map[string]any{
						"apiVersion": "gateway.example.io/v1alpha1",
						"kind":       "Foo",
						"metadata": map[string]any{
							"name":      "test",
							"namespace": "test",
						},
						"spec": map[string]any{
							"targetRefs": []any{
								map[string]any{
									"group": "gateway.networking.k8s.io",
									"kind":  "Gateway",
									"name":  "test",
								},
							},
							"data": "some data",
						},
					},
				},
			},
			extensionPolicyGroups: []schema.GroupVersionKind{
				{
					Group:   "gateway.example.io",
					Version: "v1alpha1",
					Kind:    "Foo",
				},
				{
					Group:   "gateway.example.io",
					Version: "v1alpha1",
					Kind:    "Bar",
				},
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		// Run the test cases.
		t.Run(tc.name, func(t *testing.T) {
			// Add objects referenced by test cases.
			objs := []client.Object{}

			// Create the reconciler.
			logger := logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo)

			ctx := context.Background()

			r := &gatewayAPIReconciler{
				log:             logger,
				classController: "some-gateway-class",
			}

			for _, policyObject := range tc.extensionPolicies {
				objs = append(objs, policyObject)
			}
			if len(tc.extensionPolicyGroups) > 0 {
				r.extServerPolicies = append(r.extServerPolicies, tc.extensionPolicyGroups...)
			}
			r.client = fakeclient.NewClientBuilder().
				WithScheme(envoygateway.GetScheme()).
				WithObjects(objs...).
				Build()

			resourceTree := resource.NewResources()
			err := r.processExtensionServerPolicies(ctx, resourceTree)
			if !tc.errorExpected {
				require.NoError(t, err)
				// Ensure the resource tree has the extensions
				for _, policy := range tc.extensionPolicies {
					require.Contains(t, resourceTree.ExtensionServerPolicies, *policy)
				}
			} else {
				require.Error(t, err)
			}
		})
	}
}
