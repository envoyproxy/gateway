// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
)

func TestNamespaceSelectorClient(t *testing.T) {
	// Create test namespaces
	nsMatching := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "matching-ns",
			Labels: map[string]string{
				"env": "production",
			},
		},
	}
	nsNonMatching := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "non-matching-ns",
			Labels: map[string]string{
				"env": "staging",
			},
		},
	}

	// Create test ClientTrafficPolicies in different namespaces
	ctpInMatchingNs := &egv1a1.ClientTrafficPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ctp-matching",
			Namespace: "matching-ns",
		},
	}
	ctpInNonMatchingNs := &egv1a1.ClientTrafficPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ctp-non-matching",
			Namespace: "non-matching-ns",
		},
	}

	// Create test Gateways in different namespaces
	gwInMatchingNs := &gwapiv1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gw-matching",
			Namespace: "matching-ns",
		},
		Spec: gwapiv1.GatewaySpec{
			GatewayClassName: "test-gc",
		},
	}
	gwInNonMatchingNs := &gwapiv1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gw-non-matching",
			Namespace: "non-matching-ns",
		},
		Spec: gwapiv1.GatewaySpec{
			GatewayClassName: "test-gc",
		},
	}

	// Get scheme with all required types
	scheme := envoygateway.GetScheme()

	testCases := []struct {
		name              string
		namespaceSelector *metav1.LabelSelector
		objects           []runtime.Object
		expectCTPCount    int
		expectGWCount     int
	}{
		{
			name:              "nil selector returns all resources",
			namespaceSelector: nil,
			objects: []runtime.Object{
				nsMatching, nsNonMatching,
				ctpInMatchingNs, ctpInNonMatchingNs,
				gwInMatchingNs, gwInNonMatchingNs,
			},
			expectCTPCount: 2,
			expectGWCount:  2,
		},
		{
			name: "selector filters resources by namespace labels",
			namespaceSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"env": "production",
				},
			},
			objects: []runtime.Object{
				nsMatching, nsNonMatching,
				ctpInMatchingNs, ctpInNonMatchingNs,
				gwInMatchingNs, gwInNonMatchingNs,
			},
			expectCTPCount: 1,
			expectGWCount:  1,
		},
		{
			name: "selector with no matching namespaces returns empty",
			namespaceSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"env": "development",
				},
			},
			objects: []runtime.Object{
				nsMatching, nsNonMatching,
				ctpInMatchingNs, ctpInNonMatchingNs,
				gwInMatchingNs, gwInNonMatchingNs,
			},
			expectCTPCount: 0,
			expectGWCount:  0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create fake client with objects
			fakeClient := fakeclient.NewClientBuilder().
				WithScheme(scheme).
				WithRuntimeObjects(tc.objects...).
				Build()

			// Wrap with namespace selector client
			wrappedClient := newNamespaceSelectorClient(fakeClient, tc.namespaceSelector)

			ctx := context.Background()

			// Test ClientTrafficPolicy list filtering
			ctpList := &egv1a1.ClientTrafficPolicyList{}
			err := wrappedClient.List(ctx, ctpList)
			require.NoError(t, err)
			require.Len(t, ctpList.Items, tc.expectCTPCount, "ClientTrafficPolicy count mismatch")

			// Test Gateway list filtering
			gwList := &gwapiv1.GatewayList{}
			err = wrappedClient.List(ctx, gwList)
			require.NoError(t, err)
			require.Len(t, gwList.Items, tc.expectGWCount, "Gateway count mismatch")
		})
	}
}

func TestNamespaceSelectorClientClusterScopedResources(t *testing.T) {
	// Create test GatewayClass (cluster-scoped)
	gc := &gwapiv1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-gc",
		},
		Spec: gwapiv1.GatewayClassSpec{
			ControllerName: "test-controller",
		},
	}

	scheme := envoygateway.GetScheme()

	// Create fake client
	fakeClient := fakeclient.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(gc).
		Build()

	// Wrap with namespace selector (should not affect cluster-scoped resources)
	namespaceSelector := &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"env": "production",
		},
	}
	wrappedClient := newNamespaceSelectorClient(fakeClient, namespaceSelector)

	ctx := context.Background()

	// Cluster-scoped resources should not be filtered
	gcList := &gwapiv1.GatewayClassList{}
	err := wrappedClient.List(ctx, gcList)
	require.NoError(t, err)
	require.Len(t, gcList.Items, 1, "GatewayClass should not be filtered")
}
