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
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/ir"
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

	gwDirectListener := func(name string) *ListenerContext {
		return &ListenerContext{Listener: &gwapiv1.Listener{Name: gwapiv1.SectionName(name)}}
	}
	lsListener := func(ls *gwapiv1.ListenerSet, name string) *ListenerContext {
		return &ListenerContext{Listener: &gwapiv1.Listener{Name: gwapiv1.SectionName(name)}, listenerSet: ls}
	}
	gwNN := func(name string) types.NamespacedName { return types.NamespacedName{Namespace: "default", Name: name} }

	tests := []struct {
		name      string
		gatewayNN types.NamespacedName
		listener  *ListenerContext
		want      bool
	}{
		{"gateway-direct listener targeted by name", gwNN("gateway-1"), gwDirectListener("http-1"), true},
		{"gateway-direct listener, different name: not targeted", gwNN("gateway-1"), gwDirectListener("http-2"), false},
		{"different gateway sharing the same listener name: not targeted", gwNN("gateway-2"), gwDirectListener("http-1"), false},
		{"gateway-wide CTP is not tracked: uniform across the gateway, no divergence risk", gwNN("gateway-3"), gwDirectListener("any-listener"), false},
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
			require.Equal(t, tc.want, idx.HasClusterSettingsBelowGateway(tc.gatewayNN, tc.listener))
		})
	}

	// mergeBackendsEnabled: false must produce an empty, non-nil index - no lookups should
	// ever return true.
	emptyIdx := BuildCTPClusterSettingsIndex(ctps, []*GatewayContext{gateway1}, []*gwapiv1.ListenerSet{lsSection, lsWide}, nil, nil, false)
	require.False(t, emptyIdx.HasClusterSettingsBelowGateway(gwNN("gateway-1"), gwDirectListener("http-1")))
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

// TestTranslateHeaderModifierMutations covers the ordered Mutations field, both
// on its own and combined with the legacy Set/Add/AddIfAbsent/Remove/RemoveOnMatch
// fields.
func TestTranslateHeaderModifierMutations(t *testing.T) {
	regexpType := egv1a1.StringMatchRegularExpression

	testCases := []struct {
		name    string
		in      *egv1a1.HTTPHeaderFilter
		want    []ir.HeaderMutation
		wantErr bool
	}{
		{
			name: "mutations preserve order and map every write action",
			in: &egv1a1.HTTPHeaderFilter{
				Mutations: []egv1a1.HTTPHeaderMutation{
					{Write: &egv1a1.HTTPHeaderWrite{Header: gwapiv1.HTTPHeader{Name: "x-a", Value: "1"}, Action: egv1a1.HeaderWriteOverwrite}},
					{Write: &egv1a1.HTTPHeaderWrite{Header: gwapiv1.HTTPHeader{Name: "x-a", Value: "2"}, Action: egv1a1.HeaderWriteAppend}},
					{Remove: ptr.To("x-b")},
					{RemoveOnMatch: &egv1a1.StringMatch{Type: &regexpType, Value: "^x-internal-.*"}},
					{Write: &egv1a1.HTTPHeaderWrite{Header: gwapiv1.HTTPHeader{Name: "x-c", Value: "3"}, Action: egv1a1.HeaderWriteAddIfAbsent}},
					{Write: &egv1a1.HTTPHeaderWrite{Header: gwapiv1.HTTPHeader{Name: "x-d", Value: "4"}, Action: egv1a1.HeaderWriteOverwriteIfExists}},
				},
			},
			want: []ir.HeaderMutation{
				{Write: &ir.HeaderWrite{Name: "x-a", Value: "1", Action: ir.HeaderWriteOverwrite}},
				{Write: &ir.HeaderWrite{Name: "x-a", Value: "2", Action: ir.HeaderWriteAppend}},
				{Remove: ptr.To("x-b")},
				{RemoveOnMatch: irStringMatch("", egv1a1.StringMatch{Type: &regexpType, Value: "^x-internal-.*"})},
				{Write: &ir.HeaderWrite{Name: "x-c", Value: "3", Action: ir.HeaderWriteAddIfAbsent}},
				{Write: &ir.HeaderWrite{Name: "x-d", Value: "4", Action: ir.HeaderWriteOverwriteIfExists}},
			},
		},
		{
			// The same name may be written more than once: unlike the legacy
			// fields, ordered mutations are emitted verbatim without de-duplication.
			name: "repeated writes of the same header are all kept",
			in: &egv1a1.HTTPHeaderFilter{
				Mutations: []egv1a1.HTTPHeaderMutation{
					{Write: &egv1a1.HTTPHeaderWrite{Header: gwapiv1.HTTPHeader{Name: "x-foo", Value: "1"}, Action: egv1a1.HeaderWriteOverwrite}},
					{Write: &egv1a1.HTTPHeaderWrite{Header: gwapiv1.HTTPHeader{Name: "X-Foo", Value: "2"}, Action: egv1a1.HeaderWriteAppend}},
				},
			},
			want: []ir.HeaderMutation{
				{Write: &ir.HeaderWrite{Name: "x-foo", Value: "1", Action: ir.HeaderWriteOverwrite}},
				{Write: &ir.HeaderWrite{Name: "X-Foo", Value: "2", Action: ir.HeaderWriteAppend}},
			},
		},
		{
			// Remove followed by a write of the same name is the case the flat
			// fields cannot express, since they always emit every write first.
			name: "remove then re-add the same header",
			in: &egv1a1.HTTPHeaderFilter{
				Mutations: []egv1a1.HTTPHeaderMutation{
					{Remove: ptr.To("x-recreate")},
					{Write: &egv1a1.HTTPHeaderWrite{Header: gwapiv1.HTTPHeader{Name: "x-recreate", Value: "fresh"}, Action: egv1a1.HeaderWriteAppend}},
				},
			},
			want: []ir.HeaderMutation{
				{Remove: ptr.To("x-recreate")},
				{Write: &ir.HeaderWrite{Name: "x-recreate", Value: "fresh", Action: ir.HeaderWriteAppend}},
			},
		},
		{
			name: "write action defaults to Append when unset",
			in: &egv1a1.HTTPHeaderFilter{
				Mutations: []egv1a1.HTTPHeaderMutation{
					{Write: &egv1a1.HTTPHeaderWrite{Header: gwapiv1.HTTPHeader{Name: "x-a", Value: "1"}}},
				},
			},
			want: []ir.HeaderMutation{
				{Write: &ir.HeaderWrite{Name: "x-a", Value: "1", Action: ir.HeaderWriteAppend}},
			},
		},
		{
			name: "explicit keepEmptyValue is honored",
			in: &egv1a1.HTTPHeaderFilter{
				Mutations: []egv1a1.HTTPHeaderMutation{
					{Write: &egv1a1.HTTPHeaderWrite{Header: gwapiv1.HTTPHeader{Name: "x-a", Value: "v"}, Action: egv1a1.HeaderWriteAppend, KeepEmptyValue: ptr.To(true)}},
					{Write: &egv1a1.HTTPHeaderWrite{Header: gwapiv1.HTTPHeader{Name: "x-b", Value: "w"}, Action: egv1a1.HeaderWriteAppend, KeepEmptyValue: ptr.To(false)}},
				},
			},
			want: []ir.HeaderMutation{
				{Write: &ir.HeaderWrite{Name: "x-a", Value: "v", Action: ir.HeaderWriteAppend, KeepEmptyValue: true}},
				{Write: &ir.HeaderWrite{Name: "x-b", Value: "w", Action: ir.HeaderWriteAppend, KeepEmptyValue: false}},
			},
		},
		{
			name: "every removeOnMatch matcher type is translated",
			in: &egv1a1.HTTPHeaderFilter{
				Mutations: []egv1a1.HTTPHeaderMutation{
					{RemoveOnMatch: &egv1a1.StringMatch{Value: "x-exact"}},
					{RemoveOnMatch: &egv1a1.StringMatch{Type: ptr.To(egv1a1.StringMatchPrefix), Value: "x-pre-"}},
					{RemoveOnMatch: &egv1a1.StringMatch{Type: ptr.To(egv1a1.StringMatchSuffix), Value: "-suf"}},
					{RemoveOnMatch: &egv1a1.StringMatch{Type: &regexpType, Value: "^x-re-.*"}},
				},
			},
			want: []ir.HeaderMutation{
				{RemoveOnMatch: irStringMatch("", egv1a1.StringMatch{Value: "x-exact"})},
				{RemoveOnMatch: irStringMatch("", egv1a1.StringMatch{Type: ptr.To(egv1a1.StringMatchPrefix), Value: "x-pre-"})},
				{RemoveOnMatch: irStringMatch("", egv1a1.StringMatch{Type: ptr.To(egv1a1.StringMatchSuffix), Value: "-suf"})},
				{RemoveOnMatch: irStringMatch("", egv1a1.StringMatch{Type: &regexpType, Value: "^x-re-.*"})},
			},
		},
		{
			name: "mutations are emitted before the legacy fields",
			in: &egv1a1.HTTPHeaderFilter{
				Mutations: []egv1a1.HTTPHeaderMutation{
					{Write: &egv1a1.HTTPHeaderWrite{Header: gwapiv1.HTTPHeader{Name: "x-m", Value: "m"}, Action: egv1a1.HeaderWriteAppend}},
					{Remove: ptr.To("x-m-rm")},
				},
				Add:           []gwapiv1.HTTPHeader{{Name: "x-add", Value: "a"}},
				Set:           []gwapiv1.HTTPHeader{{Name: "x-set", Value: "s"}},
				AddIfAbsent:   []gwapiv1.HTTPHeader{{Name: "x-abs", Value: "d"}},
				Remove:        []string{"x-rm"},
				RemoveOnMatch: []egv1a1.StringMatch{{Type: &regexpType, Value: "^x-drop-.*"}},
			},
			want: []ir.HeaderMutation{
				{Write: &ir.HeaderWrite{Name: "x-m", Value: "m", Action: ir.HeaderWriteAppend}},
				{Remove: ptr.To("x-m-rm")},
				{Write: &ir.HeaderWrite{Name: "x-add", Value: "a", Action: ir.HeaderWriteAppend}},
				{Write: &ir.HeaderWrite{Name: "x-set", Value: "s", Action: ir.HeaderWriteOverwrite}},
				{Write: &ir.HeaderWrite{Name: "x-abs", Value: "d", Action: ir.HeaderWriteAddIfAbsent}},
				{Remove: ptr.To("x-rm")},
				{RemoveOnMatch: irStringMatch("", egv1a1.StringMatch{Type: &regexpType, Value: "^x-drop-.*"})},
			},
		},
		{
			// The legacy de-duplication scans the whole mutation list, so a legacy
			// write whose name was already written by an explicit mutation is dropped.
			name: "legacy write is de-duplicated against an earlier explicit mutation",
			in: &egv1a1.HTTPHeaderFilter{
				Mutations: []egv1a1.HTTPHeaderMutation{
					{Write: &egv1a1.HTTPHeaderWrite{Header: gwapiv1.HTTPHeader{Name: "x-foo", Value: "from-mutation"}, Action: egv1a1.HeaderWriteAppend}},
				},
				Add: []gwapiv1.HTTPHeader{{Name: "X-Foo", Value: "from-add"}},
				Set: []gwapiv1.HTTPHeader{{Name: "x-foo", Value: "from-set"}},
			},
			want: []ir.HeaderMutation{
				{Write: &ir.HeaderWrite{Name: "x-foo", Value: "from-mutation", Action: ir.HeaderWriteAppend}},
			},
		},
		{
			name: "legacy remove is de-duplicated against an earlier explicit mutation",
			in: &egv1a1.HTTPHeaderFilter{
				Mutations: []egv1a1.HTTPHeaderMutation{
					{Remove: ptr.To("x-bar")},
				},
				Remove: []string{"X-Bar", "x-other"},
			},
			want: []ir.HeaderMutation{
				{Remove: ptr.To("x-bar")},
				{Remove: ptr.To("x-other")},
			},
		},
		{
			name: "invalid mutation entries are skipped and reported",
			in: &egv1a1.HTTPHeaderFilter{
				Mutations: []egv1a1.HTTPHeaderMutation{
					{Write: &egv1a1.HTTPHeaderWrite{Header: gwapiv1.HTTPHeader{Name: "bad/name", Value: "v"}, Action: egv1a1.HeaderWriteAppend}},
					{Write: &egv1a1.HTTPHeaderWrite{Header: gwapiv1.HTTPHeader{Name: "bad:name", Value: "v"}, Action: egv1a1.HeaderWriteAppend}},
					{Write: &egv1a1.HTTPHeaderWrite{Header: gwapiv1.HTTPHeader{Name: "", Value: "v"}, Action: egv1a1.HeaderWriteAppend}},
					{Write: &egv1a1.HTTPHeaderWrite{Header: gwapiv1.HTTPHeader{Name: "x-bad-value", Value: "  invalid"}, Action: egv1a1.HeaderWriteAppend}},
					{Remove: ptr.To("")},
					{RemoveOnMatch: &egv1a1.StringMatch{Value: ""}},
					{},
					{Write: &egv1a1.HTTPHeaderWrite{Header: gwapiv1.HTTPHeader{Name: "x-ok", Value: "v"}, Action: egv1a1.HeaderWriteAppend}},
				},
			},
			want: []ir.HeaderMutation{
				{Write: &ir.HeaderWrite{Name: "x-ok", Value: "v", Action: ir.HeaderWriteAppend}},
			},
			wantErr: true,
		},
		{
			name: "a filter whose only mutations are invalid reports an error and yields nothing",
			in: &egv1a1.HTTPHeaderFilter{
				Mutations: []egv1a1.HTTPHeaderMutation{
					{Remove: ptr.To("")},
				},
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := translateHeaderModifier(tc.in, "EarlyRequestHeaders")
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tc.want, got)
		})
	}
}
