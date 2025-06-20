// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"math"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
)

func TestInt64ToUint32(t *testing.T) {
	type testCase struct {
		Name    string
		In      int64
		Out     uint32
		Success bool
	}

	testCases := []testCase{
		{
			Name:    "valid",
			In:      1024,
			Out:     1024,
			Success: true,
		},
		{
			Name:    "invalid-underflow",
			In:      -1,
			Out:     0,
			Success: false,
		},
		{
			Name:    "invalid-overflow",
			In:      math.MaxUint32 + 1,
			Out:     0,
			Success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			out, success := int64ToUint32(tc.In)
			require.Equal(t, tc.Out, out)
			require.Equal(t, tc.Success, success)
		})
	}
}

func TestMakeIrStatusSet(t *testing.T) {
	tests := []struct {
		name string
		in   []egv1a1.HTTPStatus
		want []ir.HTTPStatus
	}{
		{
			name: "no duplicates",
			in:   []egv1a1.HTTPStatus{200, 404},
			want: []ir.HTTPStatus{200, 404},
		},
		{
			name: "with duplicates",
			in:   []egv1a1.HTTPStatus{200, 404, 200},
			want: []ir.HTTPStatus{200, 404},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makeIrStatusSet(tt.in); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("makeIrStatusSet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMakeIrTriggerSet(t *testing.T) {
	tests := []struct {
		name string
		in   []egv1a1.TriggerEnum
		want []ir.TriggerEnum
	}{
		{
			name: "no duplicates",
			in:   []egv1a1.TriggerEnum{"5xx", "reset"},
			want: []ir.TriggerEnum{"5xx", "reset"},
		},
		{
			name: "with duplicates",
			in:   []egv1a1.TriggerEnum{"5xx", "reset", "5xx"},
			want: []ir.TriggerEnum{"5xx", "reset"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makeIrTriggerSet(tt.in); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("makeIrTriggerSet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_translateRateLimitCost(t *testing.T) {
	for _, tc := range []struct {
		name string
		cost *egv1a1.RateLimitCostSpecifier
		exp  *ir.RateLimitCost
	}{
		{
			name: "number",
			cost: &egv1a1.RateLimitCostSpecifier{Number: ptr.To[uint64](1)},
			exp:  &ir.RateLimitCost{Number: ptr.To[uint64](1)},
		},
		{
			name: "metadata",
			cost: &egv1a1.RateLimitCostSpecifier{Metadata: &egv1a1.RateLimitCostMetadata{Namespace: "something.com", Key: "name"}},
			exp:  &ir.RateLimitCost{Format: ptr.To(`%DYNAMIC_METADATA(something.com:name)%`)},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			act := translateRateLimitCost(tc.cost)
			require.Equal(t, tc.exp, act)
		})
	}
}

func TestBuildHTTPProtocolUpgradeConfig(t *testing.T) {
	cases := []struct {
		name     string
		cfgs     []*egv1a1.ProtocolUpgradeConfig
		expected []ir.HTTPUpgradeConfig
	}{
		{
			name:     "empty",
			cfgs:     nil,
			expected: nil,
		},
		{
			name: "spdy",
			cfgs: []*egv1a1.ProtocolUpgradeConfig{
				{
					Type: "spdy/3.1",
				},
			},
			expected: []ir.HTTPUpgradeConfig{
				{Type: "spdy/3.1"},
			},
		},
		{
			name: "websockets-spdy",
			cfgs: []*egv1a1.ProtocolUpgradeConfig{
				{
					Type: "websockets",
				},
				{
					Type: "spdy/3.1",
				},
			},
			expected: []ir.HTTPUpgradeConfig{
				{Type: "websockets"},
				{Type: "spdy/3.1"},
			},
		},
		{
			name: "spdy-websockets",
			cfgs: []*egv1a1.ProtocolUpgradeConfig{
				{
					Type: "spdy/3.1",
				},
				{
					Type: "websockets",
				},
			},
			expected: []ir.HTTPUpgradeConfig{
				{Type: "spdy/3.1"},
				{Type: "websockets"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := buildHTTPProtocolUpgradeConfig(tc.cfgs)
			require.Equal(t, tc.expected, got)
		})
	}
}
