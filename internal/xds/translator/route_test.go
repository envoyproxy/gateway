// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"reflect"
	"testing"

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
				HTTPUpgrade: []ir.HTTPUpgradeConfig{{Type: "spdy/3.1"}},
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
				HTTPUpgrade: []ir.HTTPUpgradeConfig{
					{Type: "spdy/3.1"},
					{Type: "websocket"},
				},
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
				HTTPUpgrade: []ir.HTTPUpgradeConfig{
					{Type: "websocket"},
					{Type: "spdy/3.1"},
				},
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
