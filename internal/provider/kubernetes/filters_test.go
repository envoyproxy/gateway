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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/logging"
)

func TestGetExtensionRefFilters(t *testing.T) {
	ctx := context.Background()

	// Create test extension resources
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
				"namespace": "test-ns",
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

	// Create namespace with labels for testing namespace filtering
	testNamespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-ns",
			Labels: map[string]string{
				"env": "test",
			},
		},
	}

	defaultNamespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "default",
			Labels: map[string]string{
				"env": "prod",
			},
		},
	}

	testCases := []struct {
		name           string
		extGVKs        []schema.GroupVersionKind
		objects        []client.Object
		namespaceLabel *metav1.LabelSelector
		expectedCount  int
		expectedError  bool
	}{
		{
			name:          "no extension GVKs configured",
			extGVKs:       []schema.GroupVersionKind{},
			objects:       []client.Object{s3Backend, lambdaBackend},
			expectedCount: 0,
			expectedError: false,
		},
		{
			name: "single extension GVK with matching resources",
			extGVKs: []schema.GroupVersionKind{
				{Group: "storage.example.io", Version: "v1alpha1", Kind: "S3Backend"},
			},
			objects:       []client.Object{s3Backend, lambdaBackend, defaultNamespace, testNamespace},
			expectedCount: 1,
			expectedError: false,
		},
		{
			name: "multiple extension GVKs with matching resources",
			extGVKs: []schema.GroupVersionKind{
				{Group: "storage.example.io", Version: "v1alpha1", Kind: "S3Backend"},
				{Group: "compute.example.io", Version: "v1alpha1", Kind: "LambdaBackend"},
			},
			objects:       []client.Object{s3Backend, lambdaBackend, defaultNamespace, testNamespace},
			expectedCount: 2,
			expectedError: false,
		},
		{
			name: "namespace label filtering - include test namespace only",
			extGVKs: []schema.GroupVersionKind{
				{Group: "storage.example.io", Version: "v1alpha1", Kind: "S3Backend"},
				{Group: "compute.example.io", Version: "v1alpha1", Kind: "LambdaBackend"},
			},
			objects: []client.Object{s3Backend, lambdaBackend, defaultNamespace, testNamespace},
			namespaceLabel: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"env": "test",
				},
			},
			expectedCount: 1, // Only lambda-backend in test-ns should be included
			expectedError: false,
		},
		{
			name: "namespace label filtering - no matching namespaces",
			extGVKs: []schema.GroupVersionKind{
				{Group: "storage.example.io", Version: "v1alpha1", Kind: "S3Backend"},
				{Group: "compute.example.io", Version: "v1alpha1", Kind: "LambdaBackend"},
			},
			objects: []client.Object{s3Backend, lambdaBackend, defaultNamespace, testNamespace},
			namespaceLabel: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"env": "nonexistent",
				},
			},
			expectedCount: 0,
			expectedError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create fake client with test objects
			scheme := runtime.NewScheme()
			require.NoError(t, corev1.AddToScheme(scheme))

			fakeClient := fakeclient.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(tc.objects...).
				Build()

			// Create reconciler with test configuration
			r := &gatewayAPIReconciler{
				extGVKs:        tc.extGVKs,
				namespaceLabel: tc.namespaceLabel,
				log:            logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo),
				client:         fakeClient,
			}

			// Call the function under test
			result, err := r.getExtensionRefFilters(ctx)

			// Verify results
			if tc.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Len(t, result, tc.expectedCount)
			}
		})
	}
}

func TestGetExtensionBackendResources(t *testing.T) {
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
				"namespace": "test-ns",
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

	// Create namespace with labels for testing namespace filtering
	testNamespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-ns",
			Labels: map[string]string{
				"env": "test",
			},
		},
	}

	defaultNamespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "default",
			Labels: map[string]string{
				"env": "prod",
			},
		},
	}

	testCases := []struct {
		name           string
		extBackendGVKs []schema.GroupVersionKind
		objects        []client.Object
		namespaceLabel *metav1.LabelSelector
		expectedCount  int
		expectedError  bool
	}{
		{
			name:           "no extension backend GVKs configured",
			extBackendGVKs: []schema.GroupVersionKind{},
			objects:        []client.Object{s3Backend, lambdaBackend},
			expectedCount:  0,
			expectedError:  false,
		},
		{
			name: "single extension backend GVK with matching resources",
			extBackendGVKs: []schema.GroupVersionKind{
				{Group: "storage.example.io", Version: "v1alpha1", Kind: "S3Backend"},
			},
			objects:       []client.Object{s3Backend, lambdaBackend, defaultNamespace, testNamespace},
			expectedCount: 1,
			expectedError: false,
		},
		{
			name: "multiple extension backend GVKs with matching resources",
			extBackendGVKs: []schema.GroupVersionKind{
				{Group: "storage.example.io", Version: "v1alpha1", Kind: "S3Backend"},
				{Group: "compute.example.io", Version: "v1alpha1", Kind: "LambdaBackend"},
			},
			objects:       []client.Object{s3Backend, lambdaBackend, defaultNamespace, testNamespace},
			expectedCount: 2,
			expectedError: false,
		},
		{
			name: "namespace label filtering - include test namespace only",
			extBackendGVKs: []schema.GroupVersionKind{
				{Group: "storage.example.io", Version: "v1alpha1", Kind: "S3Backend"},
				{Group: "compute.example.io", Version: "v1alpha1", Kind: "LambdaBackend"},
			},
			objects: []client.Object{s3Backend, lambdaBackend, defaultNamespace, testNamespace},
			namespaceLabel: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"env": "test",
				},
			},
			expectedCount: 1, // Only lambda-backend in test-ns should be included
			expectedError: false,
		},
		{
			name: "namespace label filtering - no matching namespaces",
			extBackendGVKs: []schema.GroupVersionKind{
				{Group: "storage.example.io", Version: "v1alpha1", Kind: "S3Backend"},
				{Group: "compute.example.io", Version: "v1alpha1", Kind: "LambdaBackend"},
			},
			objects: []client.Object{s3Backend, lambdaBackend, defaultNamespace, testNamespace},
			namespaceLabel: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"env": "nonexistent",
				},
			},
			expectedCount: 0,
			expectedError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create fake client with test objects
			scheme := runtime.NewScheme()
			require.NoError(t, corev1.AddToScheme(scheme))

			fakeClient := fakeclient.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(tc.objects...).
				Build()

			// Create reconciler with test configuration
			r := &gatewayAPIReconciler{
				extBackendGVKs: tc.extBackendGVKs,
				namespaceLabel: tc.namespaceLabel,
				log:            logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo),
				client:         fakeClient,
			}

			// Call the function under test
			result, err := r.getExtensionBackendResources(ctx)

			// Verify results
			if tc.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Len(t, result, tc.expectedCount)
			}
		})
	}
}
