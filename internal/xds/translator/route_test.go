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

	"time"

	"github.com/envoyproxy/gateway/internal/ir"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func TestGetEffectiveTimeout(t *testing.T) {
	// Case 1: HTTP2 with StreamTimeout
	streamTimeout := &metav1.Duration{Duration: 10 * time.Second}
	httpRoute := &ir.HTTPRoute{
		IsHTTP2: true,
		Traffic: &ir.TrafficFeatures{
			Timeout: &ir.Timeout{
				HTTP: &ir.HTTPTimeout{
					StreamTimeout: streamTimeout,
				},
			},
		},
	}
	if got := getEffectiveTimeout(httpRoute); got != streamTimeout {
		t.Errorf("expected StreamTimeout, got %v", got)
	}

	// Case 2: HTTP2 without StreamTimeout (should fall back)
	reqTimeout := &metav1.Duration{Duration: 5 * time.Second}
	httpRoute2 := &ir.HTTPRoute{
		IsHTTP2: true,
		Traffic: &ir.TrafficFeatures{
			Timeout: &ir.Timeout{
				HTTP: &ir.HTTPTimeout{
					RequestTimeout: reqTimeout,
				},
			},
		},
	}
	if got := getEffectiveTimeout(httpRoute2); got != reqTimeout {
		t.Errorf("expected RequestTimeout fallback, got %v", got)
	}

	// Case 3: HTTP1 (IsHTTP2=false, should always use RequestTimeout)
	httpRoute3 := &ir.HTTPRoute{
		IsHTTP2: false,
		Traffic: &ir.TrafficFeatures{
			Timeout: &ir.Timeout{
				HTTP: &ir.HTTPTimeout{
					RequestTimeout: reqTimeout,
				},
			},
		},
	}
	if got := getEffectiveTimeout(httpRoute3); got != reqTimeout {
		t.Errorf("expected RequestTimeout for HTTP1, got %v", got)
	}

	// Case 4: HTTP2 with nil Traffic/Timeout/HTTP (should return nil)
	httpRoute4 := &ir.HTTPRoute{IsHTTP2: true}
	if got := getEffectiveTimeout(httpRoute4); got != nil {
		t.Errorf("expected nil for missing fields, got %v", got)
	}
}
