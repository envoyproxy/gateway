// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
)

func TestTranslateHeaderModifier(t *testing.T) {
	regexpType := egv1a1.StringMatchRegularExpression

	testCases := []struct {
		name    string
		in      *egv1a1.HTTPHeaderFilter
		want    []ir.HeaderMutation
		wantErr bool
	}{
		{
			name: "nil filter",
			in:   nil,
			want: nil,
		},
		{
			name: "mutations preserve order and map every action",
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
			name: "default write action is Append",
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
			name: "legacy fields keep historical order and dedup by name",
			in: &egv1a1.HTTPHeaderFilter{
				Add: []gwapiv1.HTTPHeader{
					{Name: "x-add", Value: "a"},
					{Name: "x-add", Value: "dup-ignored"},
				},
				Set:         []gwapiv1.HTTPHeader{{Name: "x-set", Value: "s"}},
				AddIfAbsent: []gwapiv1.HTTPHeader{{Name: "x-abs", Value: "d"}},
				Remove:      []string{"x-rm", "x-rm"},
				RemoveOnMatch: []egv1a1.StringMatch{
					{Type: &regexpType, Value: "^x-drop-.*"},
				},
			},
			want: []ir.HeaderMutation{
				{Write: &ir.HeaderWrite{Name: "x-add", Value: "a", Action: ir.HeaderWriteAppend}},
				{Write: &ir.HeaderWrite{Name: "x-set", Value: "s", Action: ir.HeaderWriteOverwrite}},
				{Write: &ir.HeaderWrite{Name: "x-abs", Value: "d", Action: ir.HeaderWriteAddIfAbsent}},
				{Remove: ptr.To("x-rm")},
				{RemoveOnMatch: irStringMatch("", egv1a1.StringMatch{Type: &regexpType, Value: "^x-drop-.*"})},
			},
		},
		{
			name: "mutations are applied before legacy fields",
			in: &egv1a1.HTTPHeaderFilter{
				Mutations: []egv1a1.HTTPHeaderMutation{
					{Write: &egv1a1.HTTPHeaderWrite{Header: gwapiv1.HTTPHeader{Name: "x-m", Value: "m"}, Action: egv1a1.HeaderWriteAppend}},
				},
				Add:    []gwapiv1.HTTPHeader{{Name: "x-add", Value: "a"}},
				Remove: []string{"x-rm"},
			},
			want: []ir.HeaderMutation{
				{Write: &ir.HeaderWrite{Name: "x-m", Value: "m", Action: ir.HeaderWriteAppend}},
				{Write: &ir.HeaderWrite{Name: "x-add", Value: "a", Action: ir.HeaderWriteAppend}},
				{Remove: ptr.To("x-rm")},
			},
		},
		{
			name: "legacy write is de-duplicated against an earlier explicit mutation",
			in: &egv1a1.HTTPHeaderFilter{
				Mutations: []egv1a1.HTTPHeaderMutation{
					{Write: &egv1a1.HTTPHeaderWrite{Header: gwapiv1.HTTPHeader{Name: "x-foo", Value: "from-mutation"}, Action: egv1a1.HeaderWriteAppend}},
				},
				Set: []gwapiv1.HTTPHeader{{Name: "X-Foo", Value: "from-set"}},
			},
			want: []ir.HeaderMutation{
				{Write: &ir.HeaderWrite{Name: "x-foo", Value: "from-mutation", Action: ir.HeaderWriteAppend}},
			},
		},
		{
			// An empty header value fails HeaderValueRegexp validation, so it is
			// rejected before reaching the mutation list.
			name: "empty header value is rejected",
			in: &egv1a1.HTTPHeaderFilter{
				Set: []gwapiv1.HTTPHeader{{Name: "x-empty", Value: ""}},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "explicit write keepEmptyValue is honored",
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
			name: "invalid mutation write name is skipped and reported",
			in: &egv1a1.HTTPHeaderFilter{
				Mutations: []egv1a1.HTTPHeaderMutation{
					{Write: &egv1a1.HTTPHeaderWrite{Header: gwapiv1.HTTPHeader{Name: "bad/name", Value: "v"}, Action: egv1a1.HeaderWriteAppend}},
					{Write: &egv1a1.HTTPHeaderWrite{Header: gwapiv1.HTTPHeader{Name: "x-ok", Value: "v"}, Action: egv1a1.HeaderWriteAppend}},
				},
			},
			want: []ir.HeaderMutation{
				{Write: &ir.HeaderWrite{Name: "x-ok", Value: "v", Action: ir.HeaderWriteAppend}},
			},
			wantErr: true,
		},
		{
			name:    "empty filter reports an error",
			in:      &egv1a1.HTTPHeaderFilter{},
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
