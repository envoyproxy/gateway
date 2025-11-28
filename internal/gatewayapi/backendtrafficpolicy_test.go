// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"math"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

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
			if got := makeIrStatusSet(tt.in); !slices.Equal(got, tt.want) {
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
			if got := makeIrTriggerSet(tt.in); !slices.Equal(got, tt.want) {
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

func TestBuildPassiveHealthCheck(t *testing.T) {
	cases := []struct {
		name     string
		policy   egv1a1.HealthCheck
		expected *ir.OutlierDetection
	}{
		{
			name: "nil passive health check",
			policy: egv1a1.HealthCheck{
				Passive: nil,
			},
			expected: nil,
		},
		{
			name: "basic passive health check",
			policy: egv1a1.HealthCheck{
				Passive: &egv1a1.PassiveHealthCheck{
					Interval:             ptr.To(gwapiv1.Duration("10s")),
					BaseEjectionTime:     ptr.To(gwapiv1.Duration("30s")),
					MaxEjectionPercent:   ptr.To[int32](10),
					Consecutive5xxErrors: ptr.To[uint32](5),
				},
			},
			expected: &ir.OutlierDetection{
				Interval:             ptr.To(metav1.Duration{Duration: 10 * time.Second}),
				BaseEjectionTime:     ptr.To(metav1.Duration{Duration: 30 * time.Second}),
				MaxEjectionPercent:   ptr.To[int32](10),
				Consecutive5xxErrors: ptr.To[uint32](5),
			},
		},
		{
			name: "passive health check with failure percentage threshold",
			policy: egv1a1.HealthCheck{
				Passive: &egv1a1.PassiveHealthCheck{
					Interval:                   ptr.To(gwapiv1.Duration("10s")),
					BaseEjectionTime:           ptr.To(gwapiv1.Duration("30s")),
					MaxEjectionPercent:         ptr.To[int32](10),
					Consecutive5xxErrors:       ptr.To[uint32](5),
					FailurePercentageThreshold: ptr.To[uint32](90),
				},
			},
			expected: &ir.OutlierDetection{
				Interval:                   ptr.To(metav1.Duration{Duration: 10 * time.Second}),
				BaseEjectionTime:           ptr.To(metav1.Duration{Duration: 30 * time.Second}),
				MaxEjectionPercent:         ptr.To[int32](10),
				Consecutive5xxErrors:       ptr.To[uint32](5),
				FailurePercentageThreshold: ptr.To[uint32](90),
			},
		},
		{
			name: "passive health check with all fields",
			policy: egv1a1.HealthCheck{
				Passive: &egv1a1.PassiveHealthCheck{
					SplitExternalLocalOriginErrors: ptr.To(true),
					Interval:                       ptr.To(gwapiv1.Duration("10s")),
					ConsecutiveLocalOriginFailures: ptr.To[uint32](3),
					ConsecutiveGatewayErrors:       ptr.To[uint32](2),
					Consecutive5xxErrors:           ptr.To[uint32](5),
					BaseEjectionTime:               ptr.To(gwapiv1.Duration("30s")),
					MaxEjectionPercent:             ptr.To[int32](10),
					FailurePercentageThreshold:     ptr.To[uint32](85),
				},
			},
			expected: &ir.OutlierDetection{
				SplitExternalLocalOriginErrors: ptr.To(true),
				Interval:                       ptr.To(metav1.Duration{Duration: 10 * time.Second}),
				ConsecutiveLocalOriginFailures: ptr.To[uint32](3),
				ConsecutiveGatewayErrors:       ptr.To[uint32](2),
				Consecutive5xxErrors:           ptr.To[uint32](5),
				BaseEjectionTime:               ptr.To(metav1.Duration{Duration: 30 * time.Second}),
				MaxEjectionPercent:             ptr.To[int32](10),
				FailurePercentageThreshold:     ptr.To[uint32](85),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := buildPassiveHealthCheck(tc.policy)
			require.Equal(t, tc.expected, got)
		})
	}
}

func TestBuildCompression(t *testing.T) {
	cases := []struct {
		name        string
		compression []*egv1a1.Compression
		compressor  []*egv1a1.Compression
		expected    []*ir.Compression
	}{
		{
			name:        "nil compression",
			compression: nil,
			compressor:  nil,
			expected:    nil,
		},
		{
			name: "compression without minContentLength",
			compression: []*egv1a1.Compression{
				{
					Type: egv1a1.GzipCompressorType,
					Gzip: &egv1a1.GzipCompressor{},
				},
			},
			expected: []*ir.Compression{
				{
					Type:             egv1a1.GzipCompressorType,
					ChooseFirst:      true,
					MinContentLength: nil,
				},
			},
		},
		{
			name: "compression with minContentLength",
			compression: []*egv1a1.Compression{
				{
					Type:             egv1a1.GzipCompressorType,
					Gzip:             &egv1a1.GzipCompressor{},
					MinContentLength: ptr.To(resource.MustParse("100")),
				},
			},
			expected: []*ir.Compression{
				{
					Type:             egv1a1.GzipCompressorType,
					ChooseFirst:      true,
					MinContentLength: ptr.To[uint32](100),
				},
			},
		},
		{
			name: "compressor with minContentLength",
			compressor: []*egv1a1.Compression{
				{
					Type:             egv1a1.BrotliCompressorType,
					Brotli:           &egv1a1.BrotliCompressor{},
					MinContentLength: ptr.To(resource.MustParse("200")),
				},
			},
			expected: []*ir.Compression{
				{
					Type:             egv1a1.BrotliCompressorType,
					ChooseFirst:      true,
					MinContentLength: ptr.To[uint32](200),
				},
			},
		},
		{
			name: "multiple compressors with different minContentLength",
			compressor: []*egv1a1.Compression{
				{
					Type:             egv1a1.BrotliCompressorType,
					Brotli:           &egv1a1.BrotliCompressor{},
					MinContentLength: ptr.To(resource.MustParse("50")),
				},
				{
					Type:             egv1a1.GzipCompressorType,
					Gzip:             &egv1a1.GzipCompressor{},
					MinContentLength: ptr.To(resource.MustParse("100")),
				},
			},
			expected: []*ir.Compression{
				{
					Type:             egv1a1.BrotliCompressorType,
					ChooseFirst:      true,
					MinContentLength: ptr.To[uint32](50),
				},
				{
					Type:             egv1a1.GzipCompressorType,
					ChooseFirst:      false,
					MinContentLength: ptr.To[uint32](100),
				},
			},
		},
		{
			name: "compressor takes priority over compression",
			compression: []*egv1a1.Compression{
				{
					Type:             egv1a1.GzipCompressorType,
					Gzip:             &egv1a1.GzipCompressor{},
					MinContentLength: ptr.To(resource.MustParse("100")),
				},
			},
			compressor: []*egv1a1.Compression{
				{
					Type:             egv1a1.BrotliCompressorType,
					Brotli:           &egv1a1.BrotliCompressor{},
					MinContentLength: ptr.To(resource.MustParse("200")),
				},
			},
			expected: []*ir.Compression{
				{
					Type:             egv1a1.BrotliCompressorType,
					ChooseFirst:      true,
					MinContentLength: ptr.To[uint32](200),
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := buildCompression(tc.compression, tc.compressor)
			require.Equal(t, tc.expected, got)
		})
	}
}

func TestBuildRateLimitRuleQueryParams(t *testing.T) {
	testCases := []struct {
		name        string
		rule        egv1a1.RateLimitRule
		expected    *ir.RateLimitRule
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid query parameters",
			rule: egv1a1.RateLimitRule{
				ClientSelectors: []egv1a1.RateLimitSelectCondition{
					{
						QueryParams: []egv1a1.QueryParamMatch{
							{
								Name:  "user",
								Type:  ptr.To(egv1a1.QueryParamMatchExact),
								Value: ptr.To("alice"),
							},
						},
					},
				},
				Limit: egv1a1.RateLimitValue{
					Requests: 10,
					Unit:     egv1a1.RateLimitUnitHour,
				},
			},
			expected: &ir.RateLimitRule{
				QueryParamMatches: []*ir.QueryParamMatch{
					{
						Name:          "user",
						DescriptorKey: "user", // Generated internally from name
						StringMatch: &ir.StringMatch{
							Exact:  ptr.To("alice"),
							Invert: nil,
						},
					},
				},
				Limit: ir.RateLimitValue{
					Requests: 10,
					Unit:     ir.RateLimitUnit(egv1a1.RateLimitUnitHour),
				},
				HeaderMatches: []*ir.StringMatch{},
				MethodMatches: []*ir.StringMatch{},
				Shared:        nil,
			},
			expectError: false,
		},
		{
			name: "query parameters with empty name",
			rule: egv1a1.RateLimitRule{
				ClientSelectors: []egv1a1.RateLimitSelectCondition{
					{
						QueryParams: []egv1a1.QueryParamMatch{
							{
								Name: "",
							},
						},
					},
				},
				Limit: egv1a1.RateLimitValue{
					Requests: 10,
					Unit:     egv1a1.RateLimitUnitHour,
				},
			},
			expected:    nil,
			expectError: true,
			errorMsg:    "name is required when QueryParamMatch is specified",
		},
		{
			name: "query parameters with headers",
			rule: egv1a1.RateLimitRule{
				ClientSelectors: []egv1a1.RateLimitSelectCondition{
					{
						Headers: []egv1a1.HeaderMatch{
							{
								Name:  "x-user-id",
								Type:  ptr.To(egv1a1.HeaderMatchExact),
								Value: ptr.To("alice"),
							},
						},
						QueryParams: []egv1a1.QueryParamMatch{
							{
								Name:  "user",
								Type:  ptr.To(egv1a1.QueryParamMatchExact),
								Value: ptr.To("alice"),
							},
						},
					},
				},
				Limit: egv1a1.RateLimitValue{
					Requests: 10,
					Unit:     egv1a1.RateLimitUnitHour,
				},
			},
			expected: &ir.RateLimitRule{
				QueryParamMatches: []*ir.QueryParamMatch{
					{
						Name:          "user",
						DescriptorKey: "user", // Generated internally from name
						StringMatch: &ir.StringMatch{
							Exact:  ptr.To("alice"),
							Invert: nil,
						},
					},
				},
				Limit: ir.RateLimitValue{
					Requests: 10,
					Unit:     ir.RateLimitUnit(egv1a1.RateLimitUnitHour),
				},
				HeaderMatches: []*ir.StringMatch{
					{
						Name:   "x-user-id",
						Exact:  ptr.To("alice"),
						Invert: nil,
					},
				},
				MethodMatches: []*ir.StringMatch{},
				Shared:        nil,
			},
			expectError: false,
		},
		{
			name: "query parameters with sourceCIDR",
			rule: egv1a1.RateLimitRule{
				ClientSelectors: []egv1a1.RateLimitSelectCondition{
					{
						SourceCIDR: &egv1a1.SourceMatch{
							Type:  ptr.To(egv1a1.SourceMatchDistinct),
							Value: "192.168.1.0/24",
						},
						QueryParams: []egv1a1.QueryParamMatch{
							{
								Name:  "user",
								Type:  ptr.To(egv1a1.QueryParamMatchExact),
								Value: ptr.To("alice"),
							},
						},
					},
				},
				Limit: egv1a1.RateLimitValue{
					Requests: 10,
					Unit:     egv1a1.RateLimitUnitHour,
				},
			},
			expected: &ir.RateLimitRule{
				QueryParamMatches: []*ir.QueryParamMatch{
					{
						Name:          "user",
						DescriptorKey: "user", // Generated internally from name
						StringMatch: &ir.StringMatch{
							Exact:  ptr.To("alice"),
							Invert: nil,
						},
					},
				},
				Limit: ir.RateLimitValue{
					Requests: 10,
					Unit:     ir.RateLimitUnit(egv1a1.RateLimitUnitHour),
				},
				HeaderMatches: []*ir.StringMatch{},
				MethodMatches: []*ir.StringMatch{},
				CIDRMatch: &ir.CIDRMatch{
					CIDR:     "192.168.1.0/24",
					IP:       "192.168.1.0",
					MaskLen:  24,
					Distinct: true,
				},
				Shared: nil,
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := buildRateLimitRule(tc.rule)
			if tc.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errorMsg)
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, got)
			}
		})
	}
}
