// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func TestCtpSpecHasClusterScopedFields(t *testing.T) {
	tests := []struct {
		name string
		spec *egv1a1.ClientTrafficPolicySpec
		want bool
	}{
		{name: "nil spec", spec: nil, want: false},
		{name: "empty spec", spec: &egv1a1.ClientTrafficPolicySpec{}, want: false},
		{name: "HTTP1 set", spec: &egv1a1.ClientTrafficPolicySpec{HTTP1: &egv1a1.HTTP1Settings{}}, want: true},
		{name: "HTTP2 set, no HTTP1", spec: &egv1a1.ClientTrafficPolicySpec{HTTP2: &egv1a1.HTTP2Settings{}}, want: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, ctpSpecHasClusterScopedFields(tc.spec))
		})
	}
}

func TestBuildCTPClusterSettingsIndex(t *testing.T) {
	gatewayWithHTTP1 := &GatewayContext{
		Gateway: &gwapiv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "gateway-1"},
		},
	}
	gatewayWithoutHTTP1 := &GatewayContext{
		Gateway: &gwapiv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "gateway-2"},
		},
	}
	gatewayWide := &GatewayContext{
		Gateway: &gwapiv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "gateway-3"},
		},
	}
	sectionName := gwapiv1.SectionName("http-1")

	ctps := []*egv1a1.ClientTrafficPolicy{
		{
			ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "ctp-listener"},
			Spec: egv1a1.ClientTrafficPolicySpec{
				PolicyTargetReferences: egv1a1.PolicyTargetReferences{
					TargetRefs: []gwapiv1.LocalPolicyTargetReferenceWithSectionName{
						{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
								Group: gwapiv1.GroupName,
								Kind:  resource.KindGateway,
								Name:  "gateway-1",
							},
							SectionName: &sectionName,
						},
					},
				},
				HTTP1: &egv1a1.HTTP1Settings{},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "ctp-gateway-wide"},
			Spec: egv1a1.ClientTrafficPolicySpec{
				PolicyTargetReferences: egv1a1.PolicyTargetReferences{
					TargetRefs: []gwapiv1.LocalPolicyTargetReferenceWithSectionName{
						{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
								Group: gwapiv1.GroupName,
								Kind:  resource.KindGateway,
								Name:  "gateway-3",
							},
						},
					},
				},
				HTTP1: &egv1a1.HTTP1Settings{},
			},
		},
	}

	idx := BuildCTPClusterSettingsIndex(ctps, []*GatewayContext{gatewayWithHTTP1, gatewayWithoutHTTP1, gatewayWide}, nil, nil, true)

	require.True(t, idx.HasListenerLevelClusterSettings(
		types.NamespacedName{Namespace: "default", Name: "gateway-1"}, &sectionName))
	require.False(t, idx.HasListenerLevelClusterSettings(
		types.NamespacedName{Namespace: "default", Name: "gateway-1"}, ptr.To(gwapiv1.SectionName("http-2"))))
	require.False(t, idx.HasListenerLevelClusterSettings(
		types.NamespacedName{Namespace: "default", Name: "gateway-2"}, &sectionName))

	// Gateway-wide CTP (no SectionName): any listener under that gateway inherits the
	// setting via the gatewayLevel fallback, and it's also reachable with listenerName == nil.
	require.True(t, idx.HasListenerLevelClusterSettings(
		types.NamespacedName{Namespace: "default", Name: "gateway-3"}, ptr.To(gwapiv1.SectionName("any-listener"))))
	require.True(t, idx.HasListenerLevelClusterSettings(
		types.NamespacedName{Namespace: "default", Name: "gateway-3"}, nil))
	// A different, untargeted gateway must not pick up the gateway-wide setting.
	require.False(t, idx.HasListenerLevelClusterSettings(
		types.NamespacedName{Namespace: "default", Name: "gateway-2"}, nil))

	// mergeBackendsEnabled: false must produce an empty, non-nil index — no lookups should
	// ever return true.
	emptyIdx := BuildCTPClusterSettingsIndex(ctps, []*GatewayContext{gatewayWithHTTP1}, nil, nil, false)
	require.False(t, emptyIdx.HasListenerLevelClusterSettings(
		types.NamespacedName{Namespace: "default", Name: "gateway-1"}, &sectionName))
}
