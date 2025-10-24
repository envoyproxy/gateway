// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"reflect"
	"testing"
	"time"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func TestGetEffectiveRequestTimeout(t *testing.T) {
	// Case 1: Route with Gateway API timeout should take precedence
	gatewayTimeout := &metav1.Duration{Duration: 3 * time.Second}
	httpRoute := &ir.HTTPRoute{
		Timeout: gatewayTimeout,
		Traffic: &ir.TrafficFeatures{
			Timeout: &ir.Timeout{
				HTTP: &ir.HTTPTimeout{
					RequestTimeout: &metav1.Duration{Duration: 5 * time.Second},
				},
			},
		},
	}
	if got := getEffectiveRequestTimeout(httpRoute); got != gatewayTimeout {
		t.Errorf("expected Gateway API timeout, got %v", got)
	}

	// Case 2: Route with BackendTrafficPolicy RequestTimeout should be used
	policyTimeout := &metav1.Duration{Duration: 7 * time.Second}
	httpRoute2 := &ir.HTTPRoute{
		Traffic: &ir.TrafficFeatures{
			Timeout: &ir.Timeout{
				HTTP: &ir.HTTPTimeout{
					RequestTimeout: policyTimeout,
				},
			},
		},
	}
	if got := getEffectiveRequestTimeout(httpRoute2); got != policyTimeout {
		t.Errorf("expected BackendTrafficPolicy RequestTimeout, got %v", got)
	}

	// Case 3: Route with both timeouts but Gateway API should take precedence
	httpRoute3 := &ir.HTTPRoute{
		Timeout: &metav1.Duration{Duration: 2 * time.Second},
		Traffic: &ir.TrafficFeatures{
			Timeout: &ir.Timeout{
				HTTP: &ir.HTTPTimeout{
					RequestTimeout: &metav1.Duration{Duration: 8 * time.Second},
				},
			},
		},
	}
	if got := getEffectiveRequestTimeout(httpRoute3); got.Duration != 2*time.Second {
		t.Errorf("expected Gateway API timeout to take precedence, got %v", got)
	}

	// Case 4: Route with no timeouts should return nil
	httpRoute4 := &ir.HTTPRoute{IsHTTP2: true}
	if got := getEffectiveRequestTimeout(httpRoute4); got != nil {
		t.Errorf("expected nil for missing timeouts, got %v", got)
	}
}

func TestRouteTimeoutsIndependent(t *testing.T) {
	// Test that getEffectiveRequestTimeout ignores MaxStreamDuration
	// This ensures route timeout and MaxStreamDuration are independent
	reqTimeout := &metav1.Duration{Duration: 5 * time.Second}
	maxStreamDur := &metav1.Duration{Duration: 30 * time.Second}

	httpRoute := &ir.HTTPRoute{
		IsHTTP2: true,
		Timeout: reqTimeout,
		Traffic: &ir.TrafficFeatures{
			Timeout: &ir.Timeout{
				HTTP: &ir.HTTPTimeout{
					RequestTimeout:    &metav1.Duration{Duration: 10 * time.Second}, // Should be ignored due to Gateway API timeout
					MaxStreamDuration: maxStreamDur,                                 // Should be ignored by getEffectiveRequestTimeout
				},
			},
		},
	}

	// Test that getEffectiveRequestTimeout returns the Gateway API timeout
	// and ignores MaxStreamDuration and BackendTrafficPolicy RequestTimeout
	timeout := getEffectiveRequestTimeout(httpRoute)
	if timeout == nil {
		t.Errorf("expected timeout to be set")
	} else if timeout.Duration != reqTimeout.Duration {
		t.Errorf("expected Gateway API timeout %v, got %v", reqTimeout.Duration, timeout.Duration)
	}

	// Test route without Gateway API timeout should use BackendTrafficPolicy
	httpRoute2 := &ir.HTTPRoute{
		IsHTTP2: true,
		Traffic: &ir.TrafficFeatures{
			Timeout: &ir.Timeout{
				HTTP: &ir.HTTPTimeout{
					RequestTimeout:    &metav1.Duration{Duration: 15 * time.Second},
					MaxStreamDuration: maxStreamDur, // Should be ignored
				},
			},
		},
	}

	timeout2 := getEffectiveRequestTimeout(httpRoute2)
	if timeout2 == nil {
		t.Errorf("expected timeout to be set from BackendTrafficPolicy")
	} else if timeout2.Duration != 15*time.Second {
		t.Errorf("expected BackendTrafficPolicy timeout, got %v", timeout2.Duration)
	}

	// Test that MaxStreamDuration doesn't affect request timeout calculation
	httpRoute3 := &ir.HTTPRoute{
		IsHTTP2: true,
		Traffic: &ir.TrafficFeatures{
			Timeout: &ir.Timeout{
				HTTP: &ir.HTTPTimeout{
					MaxStreamDuration: maxStreamDur, // Only MaxStreamDuration set
				},
			},
		},
	}

	timeout3 := getEffectiveRequestTimeout(httpRoute3)
	if timeout3 != nil {
		t.Errorf("expected nil timeout when only MaxStreamDuration is set, got %v", timeout3)
	}
}
