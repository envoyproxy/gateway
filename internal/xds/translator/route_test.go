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

func TestBuildXdsRoute_GrpcTimeoutHeaderMax(t *testing.T) {
	tests := []struct {
		name                     string
		httpRoute                *ir.HTTPRoute
		wantGrpcTimeoutHeaderMax *time.Duration
		wantMaxStreamDuration    *time.Duration
	}{
		{
			name: "gRPC route with GrpcTimeoutHeaderMax only",
			httpRoute: &ir.HTTPRoute{
				Name:      "test-grpc-route",
				IsHTTP2:   true,
				PathMatch: &ir.StringMatch{Exact: ptr.To("/test")},
				Destination: &ir.RouteDestination{
					Name: "test-dest",
					Settings: []*ir.DestinationSetting{
						{
							Weight: ptr.To[uint32](1),
						},
					},
				},
				Traffic: &ir.TrafficFeatures{
					Timeout: &ir.Timeout{
						HTTP: &ir.HTTPTimeout{
							GrpcTimeoutHeaderMax: &metav1.Duration{Duration: 30 * time.Second},
						},
					},
				},
			},
			wantGrpcTimeoutHeaderMax: ptr.To(30 * time.Second),
			wantMaxStreamDuration:    nil,
		},
		{
			name: "gRPC route with both GrpcTimeoutHeaderMax and StreamTimeout",
			httpRoute: &ir.HTTPRoute{
				Name:      "test-grpc-route",
				IsHTTP2:   true,
				PathMatch: &ir.StringMatch{Exact: ptr.To("/test")},
				Destination: &ir.RouteDestination{
					Name: "test-dest",
					Settings: []*ir.DestinationSetting{
						{
							Weight: ptr.To[uint32](1),
						},
					},
				},
				Traffic: &ir.TrafficFeatures{
					Timeout: &ir.Timeout{
						HTTP: &ir.HTTPTimeout{
							GrpcTimeoutHeaderMax: &metav1.Duration{Duration: 30 * time.Second},
							StreamTimeout:        &metav1.Duration{Duration: 60 * time.Second},
						},
					},
				},
			},
			wantGrpcTimeoutHeaderMax: ptr.To(30 * time.Second),
			wantMaxStreamDuration:    ptr.To(60 * time.Second),
		},
		{
			name: "gRPC route with StreamTimeout only",
			httpRoute: &ir.HTTPRoute{
				Name:      "test-grpc-route",
				IsHTTP2:   true,
				PathMatch: &ir.StringMatch{Exact: ptr.To("/test")},
				Destination: &ir.RouteDestination{
					Name: "test-dest",
					Settings: []*ir.DestinationSetting{
						{
							Weight: ptr.To[uint32](1),
						},
					},
				},
				Traffic: &ir.TrafficFeatures{
					Timeout: &ir.Timeout{
						HTTP: &ir.HTTPTimeout{
							StreamTimeout: &metav1.Duration{Duration: 60 * time.Second},
						},
					},
				},
			},
			wantGrpcTimeoutHeaderMax: nil,
			wantMaxStreamDuration:    ptr.To(60 * time.Second),
		},
		{
			name: "gRPC route with GrpcTimeoutHeaderMax set to 0 (unlimited)",
			httpRoute: &ir.HTTPRoute{
				Name:      "test-grpc-route",
				IsHTTP2:   true,
				PathMatch: &ir.StringMatch{Exact: ptr.To("/test")},
				Destination: &ir.RouteDestination{
					Name: "test-dest",
					Settings: []*ir.DestinationSetting{
						{
							Weight: ptr.To[uint32](1),
						},
					},
				},
				Traffic: &ir.TrafficFeatures{
					Timeout: &ir.Timeout{
						HTTP: &ir.HTTPTimeout{
							GrpcTimeoutHeaderMax: &metav1.Duration{Duration: 0},
						},
					},
				},
			},
			wantGrpcTimeoutHeaderMax: ptr.To(time.Duration(0)),
			wantMaxStreamDuration:    nil,
		},
		{
			name: "non-gRPC route should not have MaxStreamDuration",
			httpRoute: &ir.HTTPRoute{
				Name:      "test-http-route",
				IsHTTP2:   false,
				PathMatch: &ir.StringMatch{Exact: ptr.To("/test")},
				Destination: &ir.RouteDestination{
					Name: "test-dest",
					Settings: []*ir.DestinationSetting{
						{
							Weight: ptr.To[uint32](1),
						},
					},
				},
				Traffic: &ir.TrafficFeatures{
					Timeout: &ir.Timeout{
						HTTP: &ir.HTTPTimeout{
							GrpcTimeoutHeaderMax: &metav1.Duration{Duration: 30 * time.Second},
						},
					},
				},
			},
			wantGrpcTimeoutHeaderMax: nil,
			wantMaxStreamDuration:    nil,
		},
		{
			name: "gRPC route without timeout config should not have MaxStreamDuration",
			httpRoute: &ir.HTTPRoute{
				Name:      "test-grpc-route",
				IsHTTP2:   true,
				PathMatch: &ir.StringMatch{Exact: ptr.To("/test")},
				Destination: &ir.RouteDestination{
					Name: "test-dest",
					Settings: []*ir.DestinationSetting{
						{
							Weight: ptr.To[uint32](1),
						},
					},
				},
			},
			wantGrpcTimeoutHeaderMax: nil,
			wantMaxStreamDuration:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route, err := buildXdsRoute(tt.httpRoute, nil)
			if err != nil {
				t.Fatalf("buildXdsRoute() error = %v", err)
			}

			if route.GetRoute() == nil {
				t.Fatal("expected RouteAction, got nil")
			}

			got := route.GetRoute().MaxStreamDuration
			if tt.wantGrpcTimeoutHeaderMax == nil && tt.wantMaxStreamDuration == nil {
				if got != nil {
					t.Errorf("expected MaxStreamDuration to be nil, got %v", got)
				}
				return
			}

			if got == nil {
				t.Errorf("expected MaxStreamDuration to be set, got nil")
				return
			}

			// Check GrpcTimeoutHeaderMax
			if tt.wantGrpcTimeoutHeaderMax != nil {
				if got.GrpcTimeoutHeaderMax == nil {
					t.Errorf("expected GrpcTimeoutHeaderMax to be set, got nil")
				} else if got.GrpcTimeoutHeaderMax.AsDuration() != *tt.wantGrpcTimeoutHeaderMax {
					t.Errorf("GrpcTimeoutHeaderMax = %v, want %v",
						got.GrpcTimeoutHeaderMax.AsDuration(),
						*tt.wantGrpcTimeoutHeaderMax)
				}
			} else if got.GrpcTimeoutHeaderMax != nil {
				t.Errorf("expected GrpcTimeoutHeaderMax to be nil, got %v", got.GrpcTimeoutHeaderMax)
			}

			// Check MaxStreamDuration
			if tt.wantMaxStreamDuration != nil {
				if got.MaxStreamDuration == nil {
					t.Errorf("expected MaxStreamDuration to be set, got nil")
				} else if got.MaxStreamDuration.AsDuration() != *tt.wantMaxStreamDuration {
					t.Errorf("MaxStreamDuration = %v, want %v",
						got.MaxStreamDuration.AsDuration(),
						*tt.wantMaxStreamDuration)
				}
			} else if got.MaxStreamDuration != nil {
				t.Errorf("expected MaxStreamDuration to be nil, got %v", got.MaxStreamDuration)
			}
		})
	}
}
