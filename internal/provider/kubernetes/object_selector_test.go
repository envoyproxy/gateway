// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/provider/kubernetes/test"
)

// TestObjectSelectorEnabled verifies when watch.objectSelector is enabled: a
// non-nil selector on the Kubernetes watch configuration, orthogonal to the
// namespace watch mode.
func TestObjectSelectorEnabled(t *testing.T) {
	testCases := []struct {
		name   string
		watch  *egv1a1.KubernetesWatchMode
		expect bool
	}{
		{
			name:   "watch unset",
			watch:  nil,
			expect: false,
		},
		{
			name:   "objectSelector unset",
			watch:  &egv1a1.KubernetesWatchMode{},
			expect: false,
		},
		{
			name: "objectSelector set with the namespace list mode",
			watch: &egv1a1.KubernetesWatchMode{
				Type:           egv1a1.KubernetesWatchModeTypeNamespaces,
				Namespaces:     []string{"ns-1"},
				ObjectSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"gateway": "internal"}},
			},
			expect: true,
		},
		{
			name: "empty objectSelector is enabled and matches everything",
			watch: &egv1a1.KubernetesWatchMode{
				ObjectSelector: &metav1.LabelSelector{},
			},
			expect: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			eg := &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Provider: &egv1a1.EnvoyGatewayProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyGatewayKubernetesProvider{
							Watch: tc.watch,
						},
					},
				},
			}
			require.Equal(t, tc.expect, objectSelectorEnabled(eg))
		})
	}
}

// TestHasMatchingObjectLabels verifies the watch predicate used on the entry
// resource kinds: only objects carrying labels that match the configured
// selector are reconciled.
func TestHasMatchingObjectLabels(t *testing.T) {
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: map[string]string{"gateway": "internal"},
		MatchExpressions: []metav1.LabelSelectorRequirement{{
			Key:      "team",
			Operator: metav1.LabelSelectorOpIn,
			Values:   []string{"a", "b"},
		}},
	})
	require.NoError(t, err)

	testCases := []struct {
		name   string
		labels map[string]string
		expect bool
	}{
		{
			name:   "matches all requirements",
			labels: map[string]string{"gateway": "internal", "team": "a"},
			expect: true,
		},
		{
			name:   "missing matchLabels entry",
			labels: map[string]string{"team": "a"},
			expect: false,
		},
		{
			name:   "wrong matchLabels value",
			labels: map[string]string{"gateway": "external", "team": "a"},
			expect: false,
		},
		{
			name:   "matchExpression value not in set",
			labels: map[string]string{"gateway": "internal", "team": "c"},
			expect: false,
		},
		{
			name:   "unlabeled object",
			labels: nil,
			expect: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := gatewayAPIReconciler{
				objectSelector: selector,
			}
			route := test.GetHTTPRoute(
				types.NamespacedName{Namespace: "ns-1", Name: "httproute-test"},
				"scheduled-status-test", gwapiv1.BackendObjectReference{Name: "backend"}, "")
			route.Labels = tc.labels
			require.Equal(t, tc.expect, r.hasMatchingObjectLabels(route))
		})
	}
}

// TestObjectSelectorListOptions verifies that the ingestion List calls get the
// configured object selector applied, ANDed with any label selector already
// present on the options.
func TestObjectSelectorListOptions(t *testing.T) {
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: map[string]string{"gateway": "internal"},
	})
	require.NoError(t, err)
	matchingLabels := labels.Set{"gateway": "internal", "extra": "x"}
	nonMatchingLabels := labels.Set{"gateway": "external"}

	t.Run("no selector leaves the options unchanged", func(t *testing.T) {
		r := gatewayAPIReconciler{}
		opts := &client.ListOptions{Namespace: "ns-1"}
		got := r.objectSelectorListOptions(opts)
		require.Same(t, opts, got)
		require.Nil(t, got.LabelSelector)
	})

	t.Run("selector applied to options without a label selector", func(t *testing.T) {
		r := gatewayAPIReconciler{objectSelector: selector}
		opts := &client.ListOptions{Namespace: "ns-1"}
		got := r.objectSelectorListOptions(opts)
		require.Nil(t, opts.LabelSelector)
		require.NotNil(t, got.LabelSelector)
		require.True(t, got.LabelSelector.Matches(matchingLabels))
		require.False(t, got.LabelSelector.Matches(nonMatchingLabels))
	})

	t.Run("selector applied to nil options", func(t *testing.T) {
		r := gatewayAPIReconciler{objectSelector: selector}
		got := r.objectSelectorListOptions(nil)
		require.NotNil(t, got)
		require.NotNil(t, got.LabelSelector)
		require.True(t, got.LabelSelector.Matches(matchingLabels))
		require.False(t, got.LabelSelector.Matches(nonMatchingLabels))
	})

	t.Run("selector ANDs with an existing label selector", func(t *testing.T) {
		existing, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
			MatchLabels: map[string]string{"team": "a"},
		})
		require.NoError(t, err)
		before := existing.String()
		r := gatewayAPIReconciler{objectSelector: selector}
		got := r.objectSelectorListOptions(&client.ListOptions{LabelSelector: existing})
		require.True(t, got.LabelSelector.Matches(labels.Set{"gateway": "internal", "team": "a"}))
		require.False(t, got.LabelSelector.Matches(labels.Set{"gateway": "internal", "team": "b"}))
		require.False(t, got.LabelSelector.Matches(labels.Set{"gateway": "external", "team": "a"}))
		// The input selector is not mutated.
		require.Equal(t, before, existing.String())
	})
}

// TestLeaderElectionLeaseID verifies that the historic lease name is preserved
// for the default prefix and prefixed otherwise, so two co-located instances
// do not fight over the same lease.
func TestLeaderElectionLeaseID(t *testing.T) {
	testCases := []struct {
		prefix string
		expect string
	}{
		{prefix: "", expect: "5b9825d2.gateway.envoyproxy.io"},
		{prefix: "envoy-gateway", expect: "5b9825d2.gateway.envoyproxy.io"},
		{prefix: "tenant-a", expect: "tenant-a-5b9825d2.gateway.envoyproxy.io"},
	}
	for _, tc := range testCases {
		t.Run("prefix "+tc.prefix, func(t *testing.T) {
			require.Equal(t, tc.expect, leaderElectionLeaseID(tc.prefix))
		})
	}
}
