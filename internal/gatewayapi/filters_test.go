// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"reflect"
	"testing"

	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/envoyproxy/gateway/internal/ir"
)

func TestProcessResponseHeaderModifierFilter(t *testing.T) {
	tests := []struct {
		name           string
		filter         *gwapiv1.HTTPHeaderFilter
		expectedAdd    []ir.AddHeader
		expectedRemove []string
	}{
		{
			name: "Add headers",
			filter: &gwapiv1.HTTPHeaderFilter{
				Add: []gwapiv1.HTTPHeader{
					{Name: "X-Test-Add", Value: "foo,bar"},
				},
			},
			expectedAdd: []ir.AddHeader{
				{Name: "X-Test-Add", Append: true, Value: []string{"foo,bar"}},
			},
			expectedRemove: nil,
		},
		{
			name: "Set headers",
			filter: &gwapiv1.HTTPHeaderFilter{
				Set: []gwapiv1.HTTPHeader{
					{Name: "X-Test-Set", Value: "baz,qux"},
				},
			},
			expectedAdd: []ir.AddHeader{
				{Name: "X-Test-Set", Append: false, Value: []string{"baz,qux"}},
			},
			expectedRemove: nil,
		},
		{
			name: "Remove headers",
			filter: &gwapiv1.HTTPHeaderFilter{
				Remove: []string{"X-Test-Remove", "X-Test-Remove2"},
			},
			expectedAdd: nil,
			expectedRemove: []string{
				"X-Test-Remove", "X-Test-Remove2",
			},
		},
		{
			name: "Combined Add, Set, Remove",
			filter: &gwapiv1.HTTPHeaderFilter{
				Add: []gwapiv1.HTTPHeader{
					{Name: "X-Add", Value: "val1,val2"},
				},
				Set: []gwapiv1.HTTPHeader{
					{Name: "X-Set", Value: "val3"},
				},
				Remove: []string{"X-Remove"},
			},
			expectedAdd: []ir.AddHeader{
				{Name: "X-Add", Append: true, Value: []string{"val1,val2"}},
				{Name: "X-Set", Append: false, Value: []string{"val3"}},
			},
			expectedRemove: []string{"X-Remove"},
		},
		{
			name:           "Empty config",
			filter:         &gwapiv1.HTTPHeaderFilter{},
			expectedAdd:    nil,
			expectedRemove: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filterCtx := &HTTPFiltersContext{
				HTTPFilterIR: &HTTPFilterIR{},
			}

			translator := &Translator{}
			translator.processResponseHeaderModifierFilter(tt.filter, filterCtx)

			if !reflect.DeepEqual(filterCtx.AddResponseHeaders, tt.expectedAdd) {
				t.Errorf("unexpected AddResponseHeaders: got=%v, want=%v", filterCtx.AddResponseHeaders, tt.expectedAdd)
			}
			if !reflect.DeepEqual(filterCtx.RemoveResponseHeaders, tt.expectedRemove) {
				t.Errorf("unexpected RemoveResponseHeaders: got=%v, want=%v", filterCtx.RemoveResponseHeaders, tt.expectedRemove)
			}
		})
	}
}
