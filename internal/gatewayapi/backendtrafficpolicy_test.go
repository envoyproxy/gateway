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
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
								Type:  ptr.To(egv1a1.QueryParamMatchExact),
								Name:  "user",
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
						StringMatch: ir.StringMatch{
							Name:   "user",
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
								Type:  ptr.To(egv1a1.QueryParamMatchExact),
								Name:  "user",
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
						StringMatch: ir.StringMatch{
							Name:   "user",
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
								Type:  ptr.To(egv1a1.QueryParamMatchExact),
								Name:  "user",
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
						StringMatch: ir.StringMatch{
							Name:   "user",
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
		{
			name: "query parameters with RegularExpression",
			rule: egv1a1.RateLimitRule{
				ClientSelectors: []egv1a1.RateLimitSelectCondition{
					{
						QueryParams: []egv1a1.QueryParamMatch{
							{
								Type:  ptr.To(egv1a1.QueryParamMatchRegularExpression),
								Name:  "user",
								Value: ptr.To("alice.*"),
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
						StringMatch: ir.StringMatch{
							Name:      "user",
							SafeRegex: ptr.To("alice.*"),
							Invert:    nil,
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
			name: "query parameters with Distinct",
			rule: egv1a1.RateLimitRule{
				ClientSelectors: []egv1a1.RateLimitSelectCondition{
					{
						QueryParams: []egv1a1.QueryParamMatch{
							{
								Type: ptr.To(egv1a1.QueryParamMatchDistinct),
								Name: "user",
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
						StringMatch: ir.StringMatch{
							Name:     "user",
							Distinct: true,
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
			name: "query parameters with Invert",
			rule: egv1a1.RateLimitRule{
				ClientSelectors: []egv1a1.RateLimitSelectCondition{
					{
						QueryParams: []egv1a1.QueryParamMatch{
							{
								Type:   ptr.To(egv1a1.QueryParamMatchExact),
								Name:   "user",
								Value:  ptr.To("alice"),
								Invert: ptr.To(true),
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
						StringMatch: ir.StringMatch{
							Name:   "user",
							Exact:  ptr.To("alice"),
							Invert: ptr.To(true),
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
			name: "query parameters with invalid regex",
			rule: egv1a1.RateLimitRule{
				ClientSelectors: []egv1a1.RateLimitSelectCondition{
					{
						QueryParams: []egv1a1.QueryParamMatch{
							{
								Type:  ptr.To(egv1a1.QueryParamMatchRegularExpression),
								Name:  "user",
								Value: ptr.To("[invalid"),
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
			errorMsg:    "regex",
		},
		{
			name: "query parameters Distinct with Invert (should error)",
			rule: egv1a1.RateLimitRule{
				ClientSelectors: []egv1a1.RateLimitSelectCondition{
					{
						QueryParams: []egv1a1.QueryParamMatch{
							{
								Type:   ptr.To(egv1a1.QueryParamMatchDistinct),
								Name:   "user",
								Invert: ptr.To(true),
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
			errorMsg:    "Invert is not applicable for distinct query parameter match type",
		},
		{
			name: "query parameters with missing value for Exact",
			rule: egv1a1.RateLimitRule{
				ClientSelectors: []egv1a1.RateLimitSelectCondition{
					{
						QueryParams: []egv1a1.QueryParamMatch{
							{
								Type:  ptr.To(egv1a1.QueryParamMatchExact),
								Name:  "user",
								Value: ptr.To(""),
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
			errorMsg:    "value is required for Exact query parameter match",
		},
		{
			name: "query parameters with missing value for RegularExpression",
			rule: egv1a1.RateLimitRule{
				ClientSelectors: []egv1a1.RateLimitSelectCondition{
					{
						QueryParams: []egv1a1.QueryParamMatch{
							{
								Type:  ptr.To(egv1a1.QueryParamMatchRegularExpression),
								Name:  "user",
								Value: ptr.To(""),
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
			errorMsg:    "value is required for RegularExpression query parameter match",
		},
		{
			name: "query parameters with multiple query params",
			rule: egv1a1.RateLimitRule{
				ClientSelectors: []egv1a1.RateLimitSelectCondition{
					{
						QueryParams: []egv1a1.QueryParamMatch{
							{
								Type:  ptr.To(egv1a1.QueryParamMatchExact),
								Name:  "user",
								Value: ptr.To("alice"),
							},
							{
								Type:  ptr.To(egv1a1.QueryParamMatchExact),
								Name:  "role",
								Value: ptr.To("admin"),
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
						StringMatch: ir.StringMatch{
							Name:   "user",
							Exact:  ptr.To("alice"),
							Invert: nil,
						},
					},
					{
						StringMatch: ir.StringMatch{
							Name:   "role",
							Exact:  ptr.To("admin"),
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
			name: "query parameters with no client selectors",
			rule: egv1a1.RateLimitRule{
				ClientSelectors: nil,
				Limit: egv1a1.RateLimitValue{
					Requests: 10,
					Unit:     egv1a1.RateLimitUnitHour,
				},
			},
			expected: &ir.RateLimitRule{
				QueryParamMatches: nil,
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

func TestBTPRoutingTypeIndex(t *testing.T) {
	serviceRouting := egv1a1.ServiceRoutingType
	endpointRouting := egv1a1.EndpointRoutingType

	defaultHTTPRoute := &gwapiv1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "route-1",
		},
	}
	defaultGateway := &GatewayContext{
		Gateway: &gwapiv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "gateway-1",
			},
		},
	}

	routeNN := types.NamespacedName{Namespace: "default", Name: "route-1"}
	gatewayNN := types.NamespacedName{Namespace: "default", Name: "gateway-1"}

	tests := []struct {
		name          string
		btps          []*egv1a1.BackendTrafficPolicy
		routes        []client.Object
		gateways      []*GatewayContext
		routeKind     gwapiv1.Kind
		routeNN       types.NamespacedName
		gatewayNN     types.NamespacedName
		listenerName  *gwapiv1.SectionName
		routeRuleName *gwapiv1.SectionName
		expected      *egv1a1.RoutingType
	}{
		{
			name:      "no BTPs",
			btps:      nil,
			routes:    []client.Object{defaultHTTPRoute},
			gateways:  []*GatewayContext{defaultGateway},
			routeKind: "HTTPRoute",
			routeNN:   routeNN,
			gatewayNN: gatewayNN,
			expected:  nil,
		},
		{
			name: "BTP targeting route has priority over gateway",
			btps: []*egv1a1.BackendTrafficPolicy{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "btp-gateway",
					},
					Spec: egv1a1.BackendTrafficPolicySpec{
						PolicyTargetReferences: egv1a1.PolicyTargetReferences{
							TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
								LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
									Group: gwapiv1.Group("gateway.networking.k8s.io"),
									Kind:  gwapiv1.Kind("Gateway"),
									Name:  gwapiv1.ObjectName("gateway-1"),
								},
							},
						},
						RoutingType: &endpointRouting,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "btp-route",
					},
					Spec: egv1a1.BackendTrafficPolicySpec{
						PolicyTargetReferences: egv1a1.PolicyTargetReferences{
							TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
								LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
									Group: gwapiv1.Group("gateway.networking.k8s.io"),
									Kind:  gwapiv1.Kind("HTTPRoute"),
									Name:  gwapiv1.ObjectName("route-1"),
								},
							},
						},
						RoutingType: &serviceRouting,
					},
				},
			},
			routes:    []client.Object{defaultHTTPRoute},
			gateways:  []*GatewayContext{defaultGateway},
			routeKind: "HTTPRoute",
			routeNN:   routeNN,
			gatewayNN: gatewayNN,
			expected:  &serviceRouting,
		},
		{
			name: "BTP targeting gateway",
			btps: []*egv1a1.BackendTrafficPolicy{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "btp-gateway",
					},
					Spec: egv1a1.BackendTrafficPolicySpec{
						PolicyTargetReferences: egv1a1.PolicyTargetReferences{
							TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
								LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
									Group: gwapiv1.Group("gateway.networking.k8s.io"),
									Kind:  gwapiv1.Kind("Gateway"),
									Name:  gwapiv1.ObjectName("gateway-1"),
								},
							},
						},
						RoutingType: &serviceRouting,
					},
				},
			},
			routes:    []client.Object{defaultHTTPRoute},
			gateways:  []*GatewayContext{defaultGateway},
			routeKind: "HTTPRoute",
			routeNN:   routeNN,
			gatewayNN: gatewayNN,
			expected:  &serviceRouting,
		},
		{
			name: "BTP targeting listener (sectionName) has priority over gateway",
			btps: []*egv1a1.BackendTrafficPolicy{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "btp-gateway",
					},
					Spec: egv1a1.BackendTrafficPolicySpec{
						PolicyTargetReferences: egv1a1.PolicyTargetReferences{
							TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
								LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
									Group: gwapiv1.Group("gateway.networking.k8s.io"),
									Kind:  gwapiv1.Kind("Gateway"),
									Name:  gwapiv1.ObjectName("gateway-1"),
								},
							},
						},
						RoutingType: &endpointRouting,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "btp-listener",
					},
					Spec: egv1a1.BackendTrafficPolicySpec{
						PolicyTargetReferences: egv1a1.PolicyTargetReferences{
							TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
								LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
									Group: gwapiv1.Group("gateway.networking.k8s.io"),
									Kind:  gwapiv1.Kind("Gateway"),
									Name:  gwapiv1.ObjectName("gateway-1"),
								},
								SectionName: ptr.To(gwapiv1.SectionName("http")),
							},
						},
						RoutingType: &serviceRouting,
					},
				},
			},
			routes:       []client.Object{defaultHTTPRoute},
			gateways:     []*GatewayContext{defaultGateway},
			routeKind:    "HTTPRoute",
			routeNN:      routeNN,
			gatewayNN:    gatewayNN,
			listenerName: ptr.To(gwapiv1.SectionName("http")),
			expected:     &serviceRouting,
		},
		{
			name: "BTP with mismatched listener sectionName falls back to gateway",
			btps: []*egv1a1.BackendTrafficPolicy{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "btp-gateway",
					},
					Spec: egv1a1.BackendTrafficPolicySpec{
						PolicyTargetReferences: egv1a1.PolicyTargetReferences{
							TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
								LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
									Group: gwapiv1.Group("gateway.networking.k8s.io"),
									Kind:  gwapiv1.Kind("Gateway"),
									Name:  gwapiv1.ObjectName("gateway-1"),
								},
							},
						},
						RoutingType: &serviceRouting,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "btp-listener",
					},
					Spec: egv1a1.BackendTrafficPolicySpec{
						PolicyTargetReferences: egv1a1.PolicyTargetReferences{
							TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
								LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
									Group: gwapiv1.Group("gateway.networking.k8s.io"),
									Kind:  gwapiv1.Kind("Gateway"),
									Name:  gwapiv1.ObjectName("gateway-1"),
								},
								SectionName: ptr.To(gwapiv1.SectionName("https")),
							},
						},
						RoutingType: &endpointRouting,
					},
				},
			},
			routes:       []client.Object{defaultHTTPRoute},
			gateways:     []*GatewayContext{defaultGateway},
			routeKind:    "HTTPRoute",
			routeNN:      routeNN,
			gatewayNN:    gatewayNN,
			listenerName: ptr.To(gwapiv1.SectionName("http")),
			expected:     &serviceRouting,
		},
		{
			name: "BTP with nil RoutingType is skipped",
			btps: []*egv1a1.BackendTrafficPolicy{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "btp-route-no-routing",
					},
					Spec: egv1a1.BackendTrafficPolicySpec{
						PolicyTargetReferences: egv1a1.PolicyTargetReferences{
							TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
								LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
									Group: gwapiv1.Group("gateway.networking.k8s.io"),
									Kind:  gwapiv1.Kind("HTTPRoute"),
									Name:  gwapiv1.ObjectName("route-1"),
								},
							},
						},
						RoutingType: nil,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "btp-gateway",
					},
					Spec: egv1a1.BackendTrafficPolicySpec{
						PolicyTargetReferences: egv1a1.PolicyTargetReferences{
							TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
								LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
									Group: gwapiv1.Group("gateway.networking.k8s.io"),
									Kind:  gwapiv1.Kind("Gateway"),
									Name:  gwapiv1.ObjectName("gateway-1"),
								},
							},
						},
						RoutingType: &serviceRouting,
					},
				},
			},
			routes:    []client.Object{defaultHTTPRoute},
			gateways:  []*GatewayContext{defaultGateway},
			routeKind: "HTTPRoute",
			routeNN:   routeNN,
			gatewayNN: gatewayNN,
			expected:  &serviceRouting,
		},
		{
			name: "BTP in different namespace does not match",
			btps: []*egv1a1.BackendTrafficPolicy{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "other-namespace",
						Name:      "btp-route",
					},
					Spec: egv1a1.BackendTrafficPolicySpec{
						PolicyTargetReferences: egv1a1.PolicyTargetReferences{
							TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
								LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
									Group: gwapiv1.Group("gateway.networking.k8s.io"),
									Kind:  gwapiv1.Kind("HTTPRoute"),
									Name:  gwapiv1.ObjectName("route-1"),
								},
							},
						},
						RoutingType: &serviceRouting,
					},
				},
			},
			routes:    []client.Object{defaultHTTPRoute},
			gateways:  []*GatewayContext{defaultGateway},
			routeKind: "HTTPRoute",
			routeNN:   routeNN,
			gatewayNN: gatewayNN,
			expected:  nil,
		},
		{
			name: "BTP using targetRefs instead of targetRef",
			btps: []*egv1a1.BackendTrafficPolicy{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "btp-multiple-targets",
					},
					Spec: egv1a1.BackendTrafficPolicySpec{
						PolicyTargetReferences: egv1a1.PolicyTargetReferences{
							TargetRefs: []gwapiv1.LocalPolicyTargetReferenceWithSectionName{
								{
									LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
										Group: gwapiv1.Group("gateway.networking.k8s.io"),
										Kind:  gwapiv1.Kind("HTTPRoute"),
										Name:  gwapiv1.ObjectName("route-1"),
									},
								},
								{
									LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
										Group: gwapiv1.Group("gateway.networking.k8s.io"),
										Kind:  gwapiv1.Kind("HTTPRoute"),
										Name:  gwapiv1.ObjectName("route-2"),
									},
								},
							},
						},
						RoutingType: &serviceRouting,
					},
				},
			},
			routes:    []client.Object{defaultHTTPRoute},
			gateways:  []*GatewayContext{defaultGateway},
			routeKind: "HTTPRoute",
			routeNN:   routeNN,
			gatewayNN: gatewayNN,
			expected:  &serviceRouting,
		},
		{
			name: "full priority chain: route > listener > gateway",
			btps: []*egv1a1.BackendTrafficPolicy{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "btp-gateway",
					},
					Spec: egv1a1.BackendTrafficPolicySpec{
						PolicyTargetReferences: egv1a1.PolicyTargetReferences{
							TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
								LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
									Group: gwapiv1.Group("gateway.networking.k8s.io"),
									Kind:  gwapiv1.Kind("Gateway"),
									Name:  gwapiv1.ObjectName("gateway-1"),
								},
							},
						},
						RoutingType: &endpointRouting,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "btp-listener",
					},
					Spec: egv1a1.BackendTrafficPolicySpec{
						PolicyTargetReferences: egv1a1.PolicyTargetReferences{
							TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
								LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
									Group: gwapiv1.Group("gateway.networking.k8s.io"),
									Kind:  gwapiv1.Kind("Gateway"),
									Name:  gwapiv1.ObjectName("gateway-1"),
								},
								SectionName: ptr.To(gwapiv1.SectionName("http")),
							},
						},
						RoutingType: &endpointRouting,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "btp-route",
					},
					Spec: egv1a1.BackendTrafficPolicySpec{
						PolicyTargetReferences: egv1a1.PolicyTargetReferences{
							TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
								LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
									Group: gwapiv1.Group("gateway.networking.k8s.io"),
									Kind:  gwapiv1.Kind("HTTPRoute"),
									Name:  gwapiv1.ObjectName("route-1"),
								},
							},
						},
						RoutingType: &serviceRouting,
					},
				},
			},
			routes:       []client.Object{defaultHTTPRoute},
			gateways:     []*GatewayContext{defaultGateway},
			routeKind:    "HTTPRoute",
			routeNN:      routeNN,
			gatewayNN:    gatewayNN,
			listenerName: ptr.To(gwapiv1.SectionName("http")),
			expected:     &serviceRouting,
		},
		{
			name: "route-rule BTP has highest priority over route-level",
			btps: []*egv1a1.BackendTrafficPolicy{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "btp-route",
					},
					Spec: egv1a1.BackendTrafficPolicySpec{
						PolicyTargetReferences: egv1a1.PolicyTargetReferences{
							TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
								LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
									Group: gwapiv1.Group("gateway.networking.k8s.io"),
									Kind:  gwapiv1.Kind("HTTPRoute"),
									Name:  gwapiv1.ObjectName("route-1"),
								},
							},
						},
						RoutingType: &endpointRouting,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "btp-route-rule",
					},
					Spec: egv1a1.BackendTrafficPolicySpec{
						PolicyTargetReferences: egv1a1.PolicyTargetReferences{
							TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
								LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
									Group: gwapiv1.Group("gateway.networking.k8s.io"),
									Kind:  gwapiv1.Kind("HTTPRoute"),
									Name:  gwapiv1.ObjectName("route-1"),
								},
								SectionName: ptr.To(gwapiv1.SectionName("rule-0")),
							},
						},
						RoutingType: &serviceRouting,
					},
				},
			},
			routes:        []client.Object{defaultHTTPRoute},
			gateways:      []*GatewayContext{defaultGateway},
			routeKind:     "HTTPRoute",
			routeNN:       routeNN,
			gatewayNN:     gatewayNN,
			routeRuleName: ptr.To(gwapiv1.SectionName("rule-0")),
			expected:      &serviceRouting,
		},
		{
			name: "route-rule BTP with mismatched sectionName falls back to route",
			btps: []*egv1a1.BackendTrafficPolicy{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "btp-route-rule",
					},
					Spec: egv1a1.BackendTrafficPolicySpec{
						PolicyTargetReferences: egv1a1.PolicyTargetReferences{
							TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
								LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
									Group: gwapiv1.Group("gateway.networking.k8s.io"),
									Kind:  gwapiv1.Kind("HTTPRoute"),
									Name:  gwapiv1.ObjectName("route-1"),
								},
								SectionName: ptr.To(gwapiv1.SectionName("rule-1")),
							},
						},
						RoutingType: &endpointRouting,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "btp-route",
					},
					Spec: egv1a1.BackendTrafficPolicySpec{
						PolicyTargetReferences: egv1a1.PolicyTargetReferences{
							TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
								LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
									Group: gwapiv1.Group("gateway.networking.k8s.io"),
									Kind:  gwapiv1.Kind("HTTPRoute"),
									Name:  gwapiv1.ObjectName("route-1"),
								},
							},
						},
						RoutingType: &serviceRouting,
					},
				},
			},
			routes:        []client.Object{defaultHTTPRoute},
			gateways:      []*GatewayContext{defaultGateway},
			routeKind:     "HTTPRoute",
			routeNN:       routeNN,
			gatewayNN:     gatewayNN,
			routeRuleName: ptr.To(gwapiv1.SectionName("rule-0")),
			expected:      &serviceRouting,
		},
		{
			name: "route-rule BTP with nil routeRuleName does not match at rule level",
			btps: []*egv1a1.BackendTrafficPolicy{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "btp-route-rule",
					},
					Spec: egv1a1.BackendTrafficPolicySpec{
						PolicyTargetReferences: egv1a1.PolicyTargetReferences{
							TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
								LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
									Group: gwapiv1.Group("gateway.networking.k8s.io"),
									Kind:  gwapiv1.Kind("HTTPRoute"),
									Name:  gwapiv1.ObjectName("route-1"),
								},
								SectionName: ptr.To(gwapiv1.SectionName("rule-0")),
							},
						},
						RoutingType: &serviceRouting,
					},
				},
			},
			routes:        []client.Object{defaultHTTPRoute},
			gateways:      []*GatewayContext{defaultGateway},
			routeKind:     "HTTPRoute",
			routeNN:       routeNN,
			gatewayNN:     gatewayNN,
			routeRuleName: nil,
			expected:      nil,
		},
		{
			name: "BTP with targetSelector matching route labels",
			btps: []*egv1a1.BackendTrafficPolicy{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "btp-selector",
					},
					Spec: egv1a1.BackendTrafficPolicySpec{
						PolicyTargetReferences: egv1a1.PolicyTargetReferences{
							TargetSelectors: []egv1a1.TargetSelector{
								{
									Kind:        gwapiv1.Kind("HTTPRoute"),
									MatchLabels: map[string]string{"app": "web"},
								},
							},
						},
						RoutingType: &serviceRouting,
					},
				},
			},
			routes: []client.Object{
				&gwapiv1.HTTPRoute{
					TypeMeta: metav1.TypeMeta{Kind: "HTTPRoute", APIVersion: "gateway.networking.k8s.io/v1"},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "route-1",
						Labels:    map[string]string{"app": "web"},
					},
				},
			},
			gateways:  []*GatewayContext{defaultGateway},
			routeKind: "HTTPRoute",
			routeNN:   routeNN,
			gatewayNN: gatewayNN,
			expected:  &serviceRouting,
		},
		{
			name: "BTP with targetSelector matching gateway labels",
			btps: []*egv1a1.BackendTrafficPolicy{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "btp-selector",
					},
					Spec: egv1a1.BackendTrafficPolicySpec{
						PolicyTargetReferences: egv1a1.PolicyTargetReferences{
							TargetSelectors: []egv1a1.TargetSelector{
								{
									Kind:        gwapiv1.Kind("Gateway"),
									MatchLabels: map[string]string{"env": "prod"},
								},
							},
						},
						RoutingType: &serviceRouting,
					},
				},
			},
			routes: []client.Object{defaultHTTPRoute},
			gateways: []*GatewayContext{
				{
					Gateway: &gwapiv1.Gateway{
						TypeMeta: metav1.TypeMeta{Kind: "Gateway", APIVersion: "gateway.networking.k8s.io/v1"},
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "default",
							Name:      "gateway-1",
							Labels:    map[string]string{"env": "prod"},
						},
					},
				},
			},
			routeKind: "HTTPRoute",
			routeNN:   routeNN,
			gatewayNN: gatewayNN,
			expected:  &serviceRouting,
		},
		{
			name: "BTP with targetSelector not matching labels returns nil",
			btps: []*egv1a1.BackendTrafficPolicy{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "btp-selector",
					},
					Spec: egv1a1.BackendTrafficPolicySpec{
						PolicyTargetReferences: egv1a1.PolicyTargetReferences{
							TargetSelectors: []egv1a1.TargetSelector{
								{
									Kind:        gwapiv1.Kind("HTTPRoute"),
									MatchLabels: map[string]string{"app": "web"},
								},
							},
						},
						RoutingType: &serviceRouting,
					},
				},
			},
			routes: []client.Object{
				&gwapiv1.HTTPRoute{
					TypeMeta: metav1.TypeMeta{Kind: "HTTPRoute", APIVersion: "gateway.networking.k8s.io/v1"},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "route-1",
						Labels:    map[string]string{"app": "api"},
					},
				},
			},
			gateways:  []*GatewayContext{defaultGateway},
			routeKind: "HTTPRoute",
			routeNN:   routeNN,
			gatewayNN: gatewayNN,
			expected:  nil,
		},
		{
			name: "explicit route targetRef takes priority over targetSelector gateway",
			btps: []*egv1a1.BackendTrafficPolicy{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "btp-selector-gateway",
					},
					Spec: egv1a1.BackendTrafficPolicySpec{
						PolicyTargetReferences: egv1a1.PolicyTargetReferences{
							TargetSelectors: []egv1a1.TargetSelector{
								{
									Kind:        gwapiv1.Kind("Gateway"),
									MatchLabels: map[string]string{"env": "prod"},
								},
							},
						},
						RoutingType: &endpointRouting,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "btp-route",
					},
					Spec: egv1a1.BackendTrafficPolicySpec{
						PolicyTargetReferences: egv1a1.PolicyTargetReferences{
							TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
								LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
									Group: gwapiv1.Group("gateway.networking.k8s.io"),
									Kind:  gwapiv1.Kind("HTTPRoute"),
									Name:  gwapiv1.ObjectName("route-1"),
								},
							},
						},
						RoutingType: &serviceRouting,
					},
				},
			},
			routes: []client.Object{defaultHTTPRoute},
			gateways: []*GatewayContext{
				{
					Gateway: &gwapiv1.Gateway{
						TypeMeta: metav1.TypeMeta{Kind: "Gateway", APIVersion: "gateway.networking.k8s.io/v1"},
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "default",
							Name:      "gateway-1",
							Labels:    map[string]string{"env": "prod"},
						},
					},
				},
			},
			routeKind: "HTTPRoute",
			routeNN:   routeNN,
			gatewayNN: gatewayNN,
			expected:  &serviceRouting,
		},
		{
			name: "full priority chain: routeRule > route > listener > gateway",
			btps: []*egv1a1.BackendTrafficPolicy{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "btp-gateway",
					},
					Spec: egv1a1.BackendTrafficPolicySpec{
						PolicyTargetReferences: egv1a1.PolicyTargetReferences{
							TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
								LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
									Group: gwapiv1.Group("gateway.networking.k8s.io"),
									Kind:  gwapiv1.Kind("Gateway"),
									Name:  gwapiv1.ObjectName("gateway-1"),
								},
							},
						},
						RoutingType: &endpointRouting,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "btp-listener",
					},
					Spec: egv1a1.BackendTrafficPolicySpec{
						PolicyTargetReferences: egv1a1.PolicyTargetReferences{
							TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
								LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
									Group: gwapiv1.Group("gateway.networking.k8s.io"),
									Kind:  gwapiv1.Kind("Gateway"),
									Name:  gwapiv1.ObjectName("gateway-1"),
								},
								SectionName: ptr.To(gwapiv1.SectionName("http")),
							},
						},
						RoutingType: &endpointRouting,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "btp-route",
					},
					Spec: egv1a1.BackendTrafficPolicySpec{
						PolicyTargetReferences: egv1a1.PolicyTargetReferences{
							TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
								LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
									Group: gwapiv1.Group("gateway.networking.k8s.io"),
									Kind:  gwapiv1.Kind("HTTPRoute"),
									Name:  gwapiv1.ObjectName("route-1"),
								},
							},
						},
						RoutingType: &endpointRouting,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "btp-route-rule",
					},
					Spec: egv1a1.BackendTrafficPolicySpec{
						PolicyTargetReferences: egv1a1.PolicyTargetReferences{
							TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
								LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
									Group: gwapiv1.Group("gateway.networking.k8s.io"),
									Kind:  gwapiv1.Kind("HTTPRoute"),
									Name:  gwapiv1.ObjectName("route-1"),
								},
								SectionName: ptr.To(gwapiv1.SectionName("rule-0")),
							},
						},
						RoutingType: &serviceRouting,
					},
				},
			},
			routes:        []client.Object{defaultHTTPRoute},
			gateways:      []*GatewayContext{defaultGateway},
			routeKind:     "HTTPRoute",
			routeNN:       routeNN,
			gatewayNN:     gatewayNN,
			listenerName:  ptr.To(gwapiv1.SectionName("http")),
			routeRuleName: ptr.To(gwapiv1.SectionName("rule-0")),
			expected:      &serviceRouting,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			idx := BuildBTPRoutingTypeIndex(tt.btps, tt.routes, tt.gateways)
			got := idx.LookupBTPRoutingType(tt.routeKind, tt.routeNN, tt.gatewayNN, tt.listenerName, tt.routeRuleName)
			require.Equal(t, tt.expected, got)
		})
	}
}
