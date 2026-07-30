// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"reflect"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
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

func TestCTPClusterSettingsIndex(t *testing.T) {
	gateway1 := &GatewayContext{
		Gateway: &gwapiv1.Gateway{ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "gateway-1"}},
	}
	gateway2 := &GatewayContext{
		Gateway: &gwapiv1.Gateway{ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "gateway-2"}},
	}
	gateway3 := &GatewayContext{
		Gateway: &gwapiv1.Gateway{ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "gateway-3"}},
	}
	lsSection := &gwapiv1.ListenerSet{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "ls-section"},
		Spec:       gwapiv1.ListenerSetSpec{ParentRef: gwapiv1.ParentGatewayReference{Name: "gateway-1"}},
	}
	lsWide := &gwapiv1.ListenerSet{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "ls-wide"},
		Spec:       gwapiv1.ListenerSetSpec{ParentRef: gwapiv1.ParentGatewayReference{Name: "gateway-1"}},
	}
	sectionName := gwapiv1.SectionName("http-1")
	lsSectionName := gwapiv1.SectionName("ls-http")

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
		{
			ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "ctp-ls-listener"},
			Spec: egv1a1.ClientTrafficPolicySpec{
				PolicyTargetReferences: egv1a1.PolicyTargetReferences{
					TargetRefs: []gwapiv1.LocalPolicyTargetReferenceWithSectionName{
						{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
								Group: gwapiv1.GroupName,
								Kind:  resource.KindListenerSet,
								Name:  "ls-section",
							},
							SectionName: &lsSectionName,
						},
					},
				},
				HTTP1: &egv1a1.HTTP1Settings{},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "ctp-ls-wide"},
			Spec: egv1a1.ClientTrafficPolicySpec{
				PolicyTargetReferences: egv1a1.PolicyTargetReferences{
					TargetRefs: []gwapiv1.LocalPolicyTargetReferenceWithSectionName{
						{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
								Group: gwapiv1.GroupName,
								Kind:  resource.KindListenerSet,
								Name:  "ls-wide",
							},
						},
					},
				},
				HTTP1: &egv1a1.HTTP1Settings{},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "ctp-listener-oldest-accepted"},
			Spec: egv1a1.ClientTrafficPolicySpec{
				PolicyTargetReferences: egv1a1.PolicyTargetReferences{
					TargetRefs: []gwapiv1.LocalPolicyTargetReferenceWithSectionName{
						{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
								Group: gwapiv1.GroupName,
								Kind:  resource.KindGateway,
								Name:  "gateway-2",
							},
							SectionName: &sectionName,
						},
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "ctp-listener-younger-conflicting"},
			Spec: egv1a1.ClientTrafficPolicySpec{
				PolicyTargetReferences: egv1a1.PolicyTargetReferences{
					TargetRefs: []gwapiv1.LocalPolicyTargetReferenceWithSectionName{
						{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
								Group: gwapiv1.GroupName,
								Kind:  resource.KindGateway,
								Name:  "gateway-2",
							},
							SectionName: &sectionName,
						},
					},
				},
				HTTP1: &egv1a1.HTTP1Settings{},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "ctp-ls-wide-oldest-accepted"},
			Spec: egv1a1.ClientTrafficPolicySpec{
				PolicyTargetReferences: egv1a1.PolicyTargetReferences{
					TargetRefs: []gwapiv1.LocalPolicyTargetReferenceWithSectionName{
						{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
								Group: gwapiv1.GroupName,
								Kind:  resource.KindListenerSet,
								Name:  "ls-wide-oldest",
							},
						},
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "ctp-ls-wide-younger-conflicting"},
			Spec: egv1a1.ClientTrafficPolicySpec{
				PolicyTargetReferences: egv1a1.PolicyTargetReferences{
					TargetRefs: []gwapiv1.LocalPolicyTargetReferenceWithSectionName{
						{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
								Group: gwapiv1.GroupName,
								Kind:  resource.KindListenerSet,
								Name:  "ls-wide-oldest",
							},
						},
					},
				},
				HTTP1: &egv1a1.HTTP1Settings{},
			},
		},
	}

	lsWideOldest := &gwapiv1.ListenerSet{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "ls-wide-oldest"},
		Spec:       gwapiv1.ListenerSetSpec{ParentRef: gwapiv1.ParentGatewayReference{Name: "gateway-2"}},
	}

	gwDirectListener := func(name string) []*ListenerContext {
		return []*ListenerContext{{Listener: &gwapiv1.Listener{Name: gwapiv1.SectionName(name)}}}
	}
	lsListener := func(ls *gwapiv1.ListenerSet, name string) []*ListenerContext {
		return []*ListenerContext{{Listener: &gwapiv1.Listener{Name: gwapiv1.SectionName(name)}, listenerSet: ls}}
	}
	gwNN := func(name string) types.NamespacedName { return types.NamespacedName{Namespace: "default", Name: name} }

	tests := []struct {
		name      string
		gatewayNN types.NamespacedName
		listeners []*ListenerContext
		want      bool
	}{
		{"gateway-direct listener targeted by name", gwNN("gateway-1"), gwDirectListener("http-1"), true},
		{"gateway-direct listener, different name: not targeted", gwNN("gateway-1"), gwDirectListener("http-2"), false},
		{"different gateway sharing the same listener name: not targeted", gwNN("gateway-2"), gwDirectListener("http-1"), false},
		{"gateway-wide CTP is not tracked: uniform across the gateway, no divergence risk", gwNN("gateway-3"), gwDirectListener("any-listener"), false},
		{"gateway-wide CTP: still not tracked with an empty listener list", gwNN("gateway-3"), nil, false},
		{"ListenerSet-contributed listener targeted by name", gwNN("gateway-1"), lsListener(lsSection, "ls-http"), true},
		{"same ListenerSet, different listener name: not targeted", gwNN("gateway-1"), lsListener(lsSection, "http-2"), false},
		{"gateway-direct listener sharing a name with a targeted ListenerSet listener: not targeted", gwNN("gateway-1"), gwDirectListener("ls-http"), false},
		{"ListenerSet-wide CTP covers any of its own listeners", gwNN("gateway-1"), lsListener(lsWide, "any-listener"), true},
		{"a different ListenerSet under the same gateway must not inherit ls-wide's setting", gwNN("gateway-1"), lsListener(lsSection, "any-listener"), false},
		{"oldest accepted listener CTP with no HTTP1 blocks a younger conflicting one", gwNN("gateway-2"), gwDirectListener("http-1"), false},
		{"oldest accepted ListenerSet-wide CTP with no HTTP1 blocks a younger conflicting one", gwNN("gateway-2"), lsListener(lsWideOldest, "any-listener"), false},
	}

	idx := BuildCTPClusterSettingsIndex(ctps, []*GatewayContext{gateway1, gateway2, gateway3}, []*gwapiv1.ListenerSet{lsSection, lsWide, lsWideOldest}, nil, nil, true)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, idx.HasListenerLevelClusterSettings(tc.gatewayNN, tc.listeners))
		})
	}

	// mergeBackendsEnabled: false must produce an empty, non-nil index — no lookups should
	// ever return true.
	emptyIdx := BuildCTPClusterSettingsIndex(ctps, []*GatewayContext{gateway1}, []*gwapiv1.ListenerSet{lsSection, lsWide}, nil, nil, false)
	require.False(t, emptyIdx.HasListenerLevelClusterSettings(gwNN("gateway-1"), gwDirectListener("http-1")))
}

// TestCtpSpecHasClusterScopedFieldsExhaustive locks in today's field-by-field classification for
// ctpSpecHasClusterScopedFields, so a new field must be explicitly classified here too.
func TestCtpSpecHasClusterScopedFieldsExhaustive(t *testing.T) {
	expected := map[string]bool{
		"TCPKeepalive":        false,
		"EnableProxyProtocol": false,
		"ProxyProtocol":       false,
		"ClientIPDetection":   false,
		"TLS":                 false,
		"Path":                false,
		"Headers":             false,
		"Timeout":             false,
		"Connection":          false,
		"HTTP1":               true,
		"HTTP2":               false,
		"HTTP3":               false,
		"GRPC":                false,
		"HealthCheck":         false,
		"Scheme":              false,
	}

	actualFields := structFieldNames(reflect.TypeOf(egv1a1.ClientTrafficPolicySpec{}), map[string]bool{"PolicyTargetReferences": true})

	for _, name := range actualFields {
		want, ok := expected[name]
		if !ok {
			t.Fatalf("ClientTrafficPolicySpec field %q has no entry in this test's classification map - "+
				"decide whether it, like HTTP1, gets mirrored onto a merged backend Cluster's upstream "+
				"codec (must disqualify MergeBackends cluster deduplication, see "+
				"ctpSpecHasClusterScopedFields) and add it here", name)
		}
		t.Run(name, func(t *testing.T) {
			spec := structWithFieldSet[egv1a1.ClientTrafficPolicySpec](name)
			require.Equal(t, want, ctpSpecHasClusterScopedFields(spec),
				"ctpSpecHasClusterScopedFields's behavior for field %q doesn't match this test's classification map", name)
		})
	}

	for name := range expected {
		if !slices.Contains(actualFields, name) {
			t.Errorf("classification map has stale entry %q - field no longer exists on ClientTrafficPolicySpec", name)
		}
	}
}
