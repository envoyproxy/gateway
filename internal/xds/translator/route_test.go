// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"reflect"
	"testing"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"k8s.io/utils/ptr"

	"github.com/envoyproxy/gateway/internal/ir"
)

func TestBuildHashPolicy(t *testing.T) {
	tests := []struct {
		name      string
		httpRoute *ir.HTTPRoute
		want      []*routev3.RouteAction_HashPolicy
	}{
		{
			name:      "Nil HttpRoute",
			httpRoute: nil,
			want:      nil,
		},
		{
			name:      "Nil LoadBalancer in HttpRoute",
			httpRoute: &ir.HTTPRoute{},
			want:      nil,
		},
		{
			name: "Nil ConsistentHash in LoadBalancer",
			httpRoute: &ir.HTTPRoute{
				Traffic: &ir.TrafficFeatures{
					LoadBalancer: &ir.LoadBalancer{},
				},
			},
			want: nil,
		},
		{
			name: "ConsistentHash with nil SourceIP and Header",
			httpRoute: &ir.HTTPRoute{
				Traffic: &ir.TrafficFeatures{
					LoadBalancer: &ir.LoadBalancer{ConsistentHash: &ir.ConsistentHash{}},
				},
			},
			want: nil,
		},
		{
			name: "ConsistentHash with SourceIP set to false",
			httpRoute: &ir.HTTPRoute{
				Traffic: &ir.TrafficFeatures{
					LoadBalancer: &ir.LoadBalancer{ConsistentHash: &ir.ConsistentHash{SourceIP: ptr.To(false)}},
				},
			},
			want: nil,
		},
		{
			name: "ConsistentHash with SourceIP set to true",
			httpRoute: &ir.HTTPRoute{
				Traffic: &ir.TrafficFeatures{
					LoadBalancer: &ir.LoadBalancer{ConsistentHash: &ir.ConsistentHash{SourceIP: ptr.To(true)}},
				},
			},
			want: []*routev3.RouteAction_HashPolicy{
				{
					PolicySpecifier: &routev3.RouteAction_HashPolicy_ConnectionProperties_{
						ConnectionProperties: &routev3.RouteAction_HashPolicy_ConnectionProperties{
							SourceIp: true,
						},
					},
				},
			},
		},
		{
			name: "ConsistentHash with Header",
			httpRoute: &ir.HTTPRoute{
				Traffic: &ir.TrafficFeatures{
					LoadBalancer: &ir.LoadBalancer{ConsistentHash: &ir.ConsistentHash{Header: &ir.Header{Name: "name"}}},
				},
			},
			want: []*routev3.RouteAction_HashPolicy{
				{
					PolicySpecifier: &routev3.RouteAction_HashPolicy_Header_{
						Header: &routev3.RouteAction_HashPolicy_Header{
							HeaderName: "name",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildHashPolicy(tt.httpRoute)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildHashPolicy() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildUpgradeConfig(t *testing.T) {
	cases := []struct {
		name           string
		trafficFeature *ir.TrafficFeatures
		expected       []*routev3.RouteAction_UpgradeConfig
	}{
		{
			name:           "default",
			trafficFeature: nil,
			expected:       defaultUpgradeConfig,
		},
		{
			name: "empty",
			trafficFeature: &ir.TrafficFeatures{
				HTTPUpgrade: nil,
			},
			expected: defaultUpgradeConfig,
		},
		{
			name: "spdy",
			trafficFeature: &ir.TrafficFeatures{
				HTTPUpgrade: []string{"spdy/3.1"},
			},
			expected: []*routev3.RouteAction_UpgradeConfig{
				{
					UpgradeType: "spdy/3.1",
				},
			},
		},
		{
			name: "spdy-websocket",
			trafficFeature: &ir.TrafficFeatures{
				HTTPUpgrade: []string{"spdy/3.1", "websocket"},
			},
			expected: []*routev3.RouteAction_UpgradeConfig{
				{
					UpgradeType: "spdy/3.1",
				},
				{
					UpgradeType: "websocket",
				},
			},
		},
		{
			name: "websocket-spdy",
			trafficFeature: &ir.TrafficFeatures{
				HTTPUpgrade: []string{"websocket", "spdy/3.1"},
			},
			expected: []*routev3.RouteAction_UpgradeConfig{
				{
					UpgradeType: "websocket",
				},
				{
					UpgradeType: "spdy/3.1",
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := buildUpgradeConfig(tc.trafficFeature)
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("buildUpgradeConfig() got = %v, want %v", got, tc.expected)
			}
		})
	}
}

func TestBuildXdsAddedHeaders(t *testing.T) {
	tests := []struct {
		name         string
		headersToAdd []ir.AddHeader
		want         []*corev3.HeaderValueOption
	}{
		{
			name:         "No headers",
			headersToAdd: nil,
			want:         nil,
		},
		{
			name: "Single header, no value (should keep empty value)",
			headersToAdd: []ir.AddHeader{
				{Name: "X-Test-Empty", Value: []string{}, Append: false},
			},
			want: []*corev3.HeaderValueOption{
				{
					Header:         &corev3.HeaderValue{Key: "X-Test-Empty"},
					AppendAction:   corev3.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD,
					KeepEmptyValue: true,
				},
			},
		},
		{
			name: "Single header, one value, overwrite",
			headersToAdd: []ir.AddHeader{
				{Name: "X-Test", Value: []string{"foo"}, Append: false},
			},
			want: []*corev3.HeaderValueOption{
				{
					Header:         &corev3.HeaderValue{Key: "X-Test", Value: "foo"},
					AppendAction:   corev3.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD,
					KeepEmptyValue: false,
				},
			},
		},
		{
			name: "Single header, one value, append",
			headersToAdd: []ir.AddHeader{
				{Name: "X-Test", Value: []string{"foo"}, Append: true},
			},
			want: []*corev3.HeaderValueOption{
				{
					Header:         &corev3.HeaderValue{Key: "X-Test", Value: "foo"},
					AppendAction:   corev3.HeaderValueOption_APPEND_IF_EXISTS_OR_ADD,
					KeepEmptyValue: false,
				},
			},
		},
		{
			name: "Header with multiple values (should join with comma)",
			headersToAdd: []ir.AddHeader{
				{Name: "Cache-Control", Value: []string{"private,no-store"}, Append: false},
			},
			want: []*corev3.HeaderValueOption{
				{
					Header:         &corev3.HeaderValue{Key: "Cache-Control", Value: "private,no-store"},
					AppendAction:   corev3.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD,
					KeepEmptyValue: false,
				},
			},
		},
		{
			name: "Multiple headers",
			headersToAdd: []ir.AddHeader{
				{Name: "X-First", Value: []string{"foo"}, Append: false},
				{Name: "X-Second", Value: []string{"bar,baz"}, Append: true},
			},
			want: []*corev3.HeaderValueOption{
				{
					Header:         &corev3.HeaderValue{Key: "X-First", Value: "foo"},
					AppendAction:   corev3.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD,
					KeepEmptyValue: false,
				},
				{
					Header:         &corev3.HeaderValue{Key: "X-Second", Value: "bar,baz"},
					AppendAction:   corev3.HeaderValueOption_APPEND_IF_EXISTS_OR_ADD,
					KeepEmptyValue: false,
				},
			},
		},
		{
			name: "Header with explicit empty value string",
			headersToAdd: []ir.AddHeader{
				{Name: "X-Explicit-Empty", Value: []string{""}, Append: false},
			},
			want: []*corev3.HeaderValueOption{
				{
					Header:         &corev3.HeaderValue{Key: "X-Explicit-Empty", Value: ""},
					AppendAction:   corev3.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD,
					KeepEmptyValue: true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildXdsAddedHeaders(tt.headersToAdd)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildXdsAddedHeaders() = %v, want %v", got, tt.want)
			}
		})
	}
}
