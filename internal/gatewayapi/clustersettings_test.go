// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
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

func TestTranslateTrafficFeatures(t *testing.T) {
	tests := []struct {
		name    string
		policy  *egv1a1.ClusterSettings
		want    *ir.TrafficFeatures
		wantErr bool
	}{
		{
			name:   "nil policy",
			policy: nil,
			want:   nil,
		},
		{
			name:   "empty policy",
			policy: &egv1a1.ClusterSettings{},
			want:   nil,
		},
		{
			name: "policy with timeout",
			policy: &egv1a1.ClusterSettings{
				Timeout: &egv1a1.Timeout{
					TCP: &egv1a1.TCPTimeout{
						ConnectTimeout: ptr.To(gwapiv1.Duration("30s")),
					},
				},
			},
			want: &ir.TrafficFeatures{
				Timeout: &ir.Timeout{
					TCP: &ir.TCPTimeout{
						ConnectTimeout: ir.MetaV1DurationPtr(30 * time.Second),
					},
				},
			},
		},
		{
			name: "full policy",
			policy: &egv1a1.ClusterSettings{
				Timeout: &egv1a1.Timeout{
					TCP: &egv1a1.TCPTimeout{
						ConnectTimeout: ptr.To(gwapiv1.Duration("30s")),
					},
					HTTP: &egv1a1.HTTPTimeout{
						RequestTimeout: ptr.To(gwapiv1.Duration("60s")),
					},
				},
				Connection: &egv1a1.BackendConnection{
					BufferLimit: resource.NewQuantity(1024, resource.BinarySI),
				},
				LoadBalancer: &egv1a1.LoadBalancer{
					Type: egv1a1.RoundRobinLoadBalancerType,
				},
				ProxyProtocol: &egv1a1.ProxyProtocol{
					Version: egv1a1.ProxyProtocolVersionV1,
				},
			},
			want: &ir.TrafficFeatures{
				Timeout: &ir.Timeout{
					TCP: &ir.TCPTimeout{
						ConnectTimeout: ir.MetaV1DurationPtr(30 * time.Second),
					},
					HTTP: &ir.HTTPTimeout{
						RequestTimeout: ir.MetaV1DurationPtr(60 * time.Second),
					},
				},
				BackendConnection: &ir.BackendConnection{
					BufferLimitBytes: ptr.To(uint32(1024)),
				},
				LoadBalancer: &ir.LoadBalancer{
					RoundRobin: &ir.RoundRobin{},
				},
				ProxyProtocol: &ir.ProxyProtocol{
					Version: ir.ProxyProtocolVersionV1,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := translateTrafficFeatures(tt.policy)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestBuildClusterSettingsTimeout(t *testing.T) {
	tests := []struct {
		name    string
		policy  *egv1a1.ClusterSettings
		want    *ir.Timeout
		wantErr bool
	}{
		{
			name:   "nil timeout",
			policy: &egv1a1.ClusterSettings{},
			want:   nil,
		},
		{
			name: "valid TCP timeout",
			policy: &egv1a1.ClusterSettings{
				Timeout: &egv1a1.Timeout{
					TCP: &egv1a1.TCPTimeout{
						ConnectTimeout: ptr.To(gwapiv1.Duration("30s")),
					},
				},
			},
			want: &ir.Timeout{
				TCP: &ir.TCPTimeout{
					ConnectTimeout: ir.MetaV1DurationPtr(30 * time.Second),
				},
			},
		},
		{
			name: "valid HTTP timeout",
			policy: &egv1a1.ClusterSettings{
				Timeout: &egv1a1.Timeout{
					HTTP: &egv1a1.HTTPTimeout{
						ConnectionIdleTimeout: ptr.To(gwapiv1.Duration("300s")),
						MaxConnectionDuration: ptr.To(gwapiv1.Duration("900s")),
						RequestTimeout:        ptr.To(gwapiv1.Duration("15s")),
						MaxStreamDuration:     ptr.To(gwapiv1.Duration("60s")),
					},
				},
			},
			want: &ir.Timeout{
				HTTP: &ir.HTTPTimeout{
					ConnectionIdleTimeout: ir.MetaV1DurationPtr(300 * time.Second),
					MaxConnectionDuration: ir.MetaV1DurationPtr(900 * time.Second),
					RequestTimeout:        ir.MetaV1DurationPtr(15 * time.Second),
					MaxStreamDuration:     ptr.To(metav1.Duration{Duration: 60 * time.Second}),
				},
			},
		},
		{
			name: "invalid TCP timeout",
			policy: &egv1a1.ClusterSettings{
				Timeout: &egv1a1.Timeout{
					TCP: &egv1a1.TCPTimeout{
						ConnectTimeout: ptr.To(gwapiv1.Duration("invalid")),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid HTTP timeout",
			policy: &egv1a1.ClusterSettings{
				Timeout: &egv1a1.Timeout{
					HTTP: &egv1a1.HTTPTimeout{
						RequestTimeout: ptr.To(gwapiv1.Duration("invalid")),
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildClusterSettingsTimeout(tt.policy)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestBuildBackendConnection(t *testing.T) {
	tests := []struct {
		name    string
		policy  *egv1a1.ClusterSettings
		want    *ir.BackendConnection
		wantErr bool
	}{
		{
			name:   "nil connection",
			policy: &egv1a1.ClusterSettings{},
			want:   nil,
		},
		{
			name: "valid buffer limit",
			policy: &egv1a1.ClusterSettings{
				Connection: &egv1a1.BackendConnection{
					BufferLimit: resource.NewQuantity(1024, resource.BinarySI),
				},
			},
			want: &ir.BackendConnection{
				BufferLimitBytes: ptr.To(uint32(1024)),
			},
		},
		{
			name: "negative buffer limit",
			policy: &egv1a1.ClusterSettings{
				Connection: &egv1a1.BackendConnection{
					BufferLimit: resource.NewQuantity(-1, resource.BinarySI),
				},
			},
			wantErr: true,
		},
		{
			name: "overflow buffer limit",
			policy: &egv1a1.ClusterSettings{
				Connection: &egv1a1.BackendConnection{
					BufferLimit: resource.NewQuantity(int64(0xffffffff+1), resource.BinarySI),
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildBackendConnection(tt.policy)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestBuildTCPKeepAlive(t *testing.T) {
	tests := []struct {
		name    string
		policy  *egv1a1.ClusterSettings
		want    *ir.TCPKeepalive
		wantErr bool
	}{
		{
			name:   "nil tcp keepalive",
			policy: &egv1a1.ClusterSettings{},
			want:   nil,
		},
		{
			name: "valid keepalive settings",
			policy: &egv1a1.ClusterSettings{
				TCPKeepalive: &egv1a1.TCPKeepalive{
					Probes:   ptr.To(uint32(3)),
					IdleTime: ptr.To(gwapiv1.Duration("60s")),
					Interval: ptr.To(gwapiv1.Duration("10s")),
				},
			},
			want: &ir.TCPKeepalive{
				Probes:   ptr.To(uint32(3)),
				IdleTime: ptr.To(uint32(60)),
				Interval: ptr.To(uint32(10)),
			},
		},
		{
			name: "invalid idle time",
			policy: &egv1a1.ClusterSettings{
				TCPKeepalive: &egv1a1.TCPKeepalive{
					IdleTime: ptr.To(gwapiv1.Duration("invalid")),
				},
			},
			wantErr: true,
		},
		{
			name: "invalid interval",
			policy: &egv1a1.ClusterSettings{
				TCPKeepalive: &egv1a1.TCPKeepalive{
					Interval: ptr.To(gwapiv1.Duration("invalid")),
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildTCPKeepAlive(tt.policy)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestBuildCircuitBreaker(t *testing.T) {
	tests := []struct {
		name    string
		policy  *egv1a1.ClusterSettings
		want    *ir.CircuitBreaker
		wantErr bool
	}{
		{
			name:   "nil circuit breaker",
			policy: &egv1a1.ClusterSettings{},
			want:   nil,
		},
		{
			name: "valid circuit breaker settings",
			policy: &egv1a1.ClusterSettings{
				CircuitBreaker: &egv1a1.CircuitBreaker{
					MaxConnections:           ptr.To(int64(100)),
					MaxParallelRequests:      ptr.To(int64(200)),
					MaxPendingRequests:       ptr.To(int64(50)),
					MaxParallelRetries:       ptr.To(int64(10)),
					MaxRequestsPerConnection: ptr.To(int64(1000)),
					PerEndpoint: &egv1a1.PerEndpointCircuitBreakers{
						MaxConnections: ptr.To(int64(25)),
					},
				},
			},
			want: &ir.CircuitBreaker{
				MaxConnections:           ptr.To(uint32(100)),
				MaxParallelRequests:      ptr.To(uint32(200)),
				MaxPendingRequests:       ptr.To(uint32(50)),
				MaxParallelRetries:       ptr.To(uint32(10)),
				MaxRequestsPerConnection: ptr.To(uint32(1000)),
				PerEndpoint: &ir.PerEndpointCircuitBreakers{
					MaxConnections: ptr.To(uint32(25)),
				},
			},
		},
		{
			name: "invalid max connections",
			policy: &egv1a1.ClusterSettings{
				CircuitBreaker: &egv1a1.CircuitBreaker{
					MaxConnections: ptr.To(int64(-1)),
				},
			},
			wantErr: true,
		},
		{
			name: "invalid max parallel requests",
			policy: &egv1a1.ClusterSettings{
				CircuitBreaker: &egv1a1.CircuitBreaker{
					MaxParallelRequests: ptr.To(int64(0xffffffff + 1)),
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildCircuitBreaker(tt.policy)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestBuildLoadBalancer(t *testing.T) {
	tests := []struct {
		name    string
		policy  *egv1a1.ClusterSettings
		want    *ir.LoadBalancer
		wantErr bool
	}{
		{
			name:   "nil load balancer",
			policy: &egv1a1.ClusterSettings{},
			want:   nil,
		},
		{
			name: "round robin load balancer",
			policy: &egv1a1.ClusterSettings{
				LoadBalancer: &egv1a1.LoadBalancer{
					Type: egv1a1.RoundRobinLoadBalancerType,
				},
			},
			want: &ir.LoadBalancer{
				RoundRobin: &ir.RoundRobin{},
			},
		},
		{
			name: "least request load balancer with slow start",
			policy: &egv1a1.ClusterSettings{
				LoadBalancer: &egv1a1.LoadBalancer{
					Type: egv1a1.LeastRequestLoadBalancerType,
					SlowStart: &egv1a1.SlowStart{
						Window: ptr.To(gwapiv1.Duration("30s")),
					},
				},
			},
			want: &ir.LoadBalancer{
				LeastRequest: &ir.LeastRequest{
					SlowStart: &ir.SlowStart{
						Window: ir.MetaV1DurationPtr(30 * time.Second),
					},
				},
			},
		},
		{
			name: "random load balancer",
			policy: &egv1a1.ClusterSettings{
				LoadBalancer: &egv1a1.LoadBalancer{
					Type: egv1a1.RandomLoadBalancerType,
				},
			},
			want: &ir.LoadBalancer{
				Random: &ir.Random{},
			},
		},
		{
			name: "consistent hash load balancer - source IP",
			policy: &egv1a1.ClusterSettings{
				LoadBalancer: &egv1a1.LoadBalancer{
					Type: egv1a1.ConsistentHashLoadBalancerType,
					ConsistentHash: &egv1a1.ConsistentHash{
						Type: egv1a1.SourceIPConsistentHashType,
					},
				},
			},
			want: &ir.LoadBalancer{
				ConsistentHash: &ir.ConsistentHash{
					SourceIP: ptr.To(true),
				},
			},
		},
		{
			name: "consistent hash load balancer - header",
			policy: &egv1a1.ClusterSettings{
				LoadBalancer: &egv1a1.LoadBalancer{
					Type: egv1a1.ConsistentHashLoadBalancerType,
					ConsistentHash: &egv1a1.ConsistentHash{
						Type: egv1a1.HeaderConsistentHashType,
						Header: &egv1a1.Header{
							Name: "x-user-id",
						},
					},
				},
			},
			want: &ir.LoadBalancer{
				ConsistentHash: &ir.ConsistentHash{
					Header: &ir.Header{
						Name: "x-user-id",
					},
				},
			},
		},
		{
			name: "consistent hash load balancer - cookie",
			policy: &egv1a1.ClusterSettings{
				LoadBalancer: &egv1a1.LoadBalancer{
					Type: egv1a1.ConsistentHashLoadBalancerType,
					ConsistentHash: &egv1a1.ConsistentHash{
						Type: egv1a1.CookieConsistentHashType,
						Cookie: &egv1a1.Cookie{
							Name: "session-id",
							TTL:  ptr.To(gwapiv1.Duration("3600s")),
						},
					},
				},
			},
			want: &ir.LoadBalancer{
				ConsistentHash: &ir.ConsistentHash{
					Cookie: &egv1a1.Cookie{
						Name: "session-id",
						TTL:  ptr.To(gwapiv1.Duration("3600s")),
					},
				},
			},
		},
		{
			name: "consistent hash with valid table size",
			policy: &egv1a1.ClusterSettings{
				LoadBalancer: &egv1a1.LoadBalancer{
					Type: egv1a1.ConsistentHashLoadBalancerType,
					ConsistentHash: &egv1a1.ConsistentHash{
						Type:      egv1a1.SourceIPConsistentHashType,
						TableSize: ptr.To(uint64(101)), // 101 is prime
					},
				},
			},
			want: &ir.LoadBalancer{
				ConsistentHash: &ir.ConsistentHash{
					SourceIP:  ptr.To(true),
					TableSize: ptr.To(uint64(101)),
				},
			},
		},
		{
			name: "zone aware load balancer",
			policy: &egv1a1.ClusterSettings{
				LoadBalancer: &egv1a1.LoadBalancer{
					Type: egv1a1.RoundRobinLoadBalancerType,
					ZoneAware: &egv1a1.ZoneAware{
						PreferLocal: &egv1a1.PreferLocalZone{
							MinEndpointsThreshold: ptr.To(uint64(2)),
							Force: &egv1a1.ForceLocalZone{
								MinEndpointsInZoneThreshold: ptr.To(uint32(1)),
							},
						},
					},
				},
			},
			want: &ir.LoadBalancer{
				RoundRobin: &ir.RoundRobin{},
				PreferLocal: &ir.PreferLocalZone{
					MinEndpointsThreshold: ptr.To(uint64(2)),
					Force: &ir.ForceLocalZone{
						MinEndpointsInZoneThreshold: ptr.To(uint32(1)),
					},
				},
			},
		},
		{
			name: "with endpoint override",
			policy: &egv1a1.ClusterSettings{
				LoadBalancer: &egv1a1.LoadBalancer{
					Type: egv1a1.RoundRobinLoadBalancerType,
					EndpointOverride: &egv1a1.EndpointOverride{
						ExtractFrom: []egv1a1.EndpointOverrideExtractFrom{
							{
								Header: ptr.To("x-endpoint"),
							},
						},
					},
				},
			},
			want: &ir.LoadBalancer{
				RoundRobin: &ir.RoundRobin{},
				EndpointOverride: &ir.EndpointOverride{
					ExtractFrom: []ir.EndpointOverrideExtractFrom{
						{
							Header: ptr.To("x-endpoint"),
						},
					},
				},
			},
		},
		{
			name: "invalid slow start duration",
			policy: &egv1a1.ClusterSettings{
				LoadBalancer: &egv1a1.LoadBalancer{
					Type: egv1a1.LeastRequestLoadBalancerType,
					SlowStart: &egv1a1.SlowStart{
						Window: ptr.To(gwapiv1.Duration("invalid")),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid table size - too large",
			policy: &egv1a1.ClusterSettings{
				LoadBalancer: &egv1a1.LoadBalancer{
					Type: egv1a1.ConsistentHashLoadBalancerType,
					ConsistentHash: &egv1a1.ConsistentHash{
						TableSize: ptr.To(uint64(MaxConsistentHashTableSize + 1)),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid table size - not prime",
			policy: &egv1a1.ClusterSettings{
				LoadBalancer: &egv1a1.LoadBalancer{
					Type: egv1a1.ConsistentHashLoadBalancerType,
					ConsistentHash: &egv1a1.ConsistentHash{
						TableSize: ptr.To(uint64(100)), // 100 is not prime
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildLoadBalancer(tt.policy)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestBuildProxyProtocol(t *testing.T) {
	tests := []struct {
		name   string
		policy *egv1a1.ClusterSettings
		want   *ir.ProxyProtocol
	}{
		{
			name:   "nil proxy protocol",
			policy: &egv1a1.ClusterSettings{},
			want:   nil,
		},
		{
			name: "proxy protocol v1",
			policy: &egv1a1.ClusterSettings{
				ProxyProtocol: &egv1a1.ProxyProtocol{
					Version: egv1a1.ProxyProtocolVersionV1,
				},
			},
			want: &ir.ProxyProtocol{
				Version: ir.ProxyProtocolVersionV1,
			},
		},
		{
			name: "proxy protocol v2",
			policy: &egv1a1.ClusterSettings{
				ProxyProtocol: &egv1a1.ProxyProtocol{
					Version: egv1a1.ProxyProtocolVersionV2,
				},
			},
			want: &ir.ProxyProtocol{
				Version: ir.ProxyProtocolVersionV2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildProxyProtocol(tt.policy)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestBuildHealthCheck(t *testing.T) {
	tests := []struct {
		name   string
		policy *egv1a1.ClusterSettings
		want   *ir.HealthCheck
	}{
		{
			name:   "nil health check",
			policy: &egv1a1.ClusterSettings{},
			want:   nil,
		},
		{
			name: "health check with panic threshold",
			policy: &egv1a1.ClusterSettings{
				HealthCheck: &egv1a1.HealthCheck{
					PanicThreshold: ptr.To(uint32(50)),
				},
			},
			want: &ir.HealthCheck{
				PanicThreshold: ptr.To(uint32(50)),
			},
		},
		{
			name: "health check with passive check",
			policy: &egv1a1.ClusterSettings{
				HealthCheck: &egv1a1.HealthCheck{
					Passive: &egv1a1.PassiveHealthCheck{
						Interval:             ptr.To(gwapiv1.Duration("10s")),
						BaseEjectionTime:     ptr.To(gwapiv1.Duration("30s")),
						MaxEjectionPercent:   ptr.To(int32(10)),
						Consecutive5xxErrors: ptr.To(uint32(5)),
					},
				},
			},
			want: &ir.HealthCheck{
				Passive: &ir.OutlierDetection{
					Interval:             ir.MetaV1DurationPtr(10 * time.Second),
					BaseEjectionTime:     ir.MetaV1DurationPtr(30 * time.Second),
					MaxEjectionPercent:   ptr.To(int32(10)),
					Consecutive5xxErrors: ptr.To(uint32(5)),
				},
			},
		},
		{
			name: "health check with active HTTP check",
			policy: &egv1a1.ClusterSettings{
				HealthCheck: &egv1a1.HealthCheck{
					Active: &egv1a1.ActiveHealthCheck{
						Type:               egv1a1.ActiveHealthCheckerTypeHTTP,
						Timeout:            ptr.To(gwapiv1.Duration("5s")),
						Interval:           ptr.To(gwapiv1.Duration("10s")),
						UnhealthyThreshold: ptr.To(uint32(3)),
						HealthyThreshold:   ptr.To(uint32(2)),
						HTTP: &egv1a1.HTTPActiveHealthChecker{
							Path:             "/health",
							Method:           ptr.To("GET"),
							ExpectedStatuses: []egv1a1.HTTPStatus{200, 201},
						},
					},
				},
			},
			want: &ir.HealthCheck{
				Active: &ir.ActiveHealthCheck{
					Timeout:            ir.MetaV1DurationPtr(5 * time.Second),
					Interval:           ir.MetaV1DurationPtr(10 * time.Second),
					UnhealthyThreshold: ptr.To(uint32(3)),
					HealthyThreshold:   ptr.To(uint32(2)),
					HTTP: &ir.HTTPHealthChecker{
						Path:             "/health",
						Method:           ptr.To("GET"),
						ExpectedStatuses: []ir.HTTPStatus{200, 201},
					},
				},
			},
		},
		{
			name: "health check with active TCP check",
			policy: &egv1a1.ClusterSettings{
				HealthCheck: &egv1a1.HealthCheck{
					Active: &egv1a1.ActiveHealthCheck{
						Type:     egv1a1.ActiveHealthCheckerTypeTCP,
						Timeout:  ptr.To(gwapiv1.Duration("3s")),
						Interval: ptr.To(gwapiv1.Duration("15s")),
						TCP: &egv1a1.TCPActiveHealthChecker{
							Send: &egv1a1.ActiveHealthCheckPayload{
								Type: egv1a1.ActiveHealthCheckPayloadTypeText,
								Text: ptr.To("ping"),
							},
							Receive: &egv1a1.ActiveHealthCheckPayload{
								Type: egv1a1.ActiveHealthCheckPayloadTypeText,
								Text: ptr.To("pong"),
							},
						},
					},
				},
			},
			want: &ir.HealthCheck{
				Active: &ir.ActiveHealthCheck{
					Timeout:  ir.MetaV1DurationPtr(3 * time.Second),
					Interval: ir.MetaV1DurationPtr(15 * time.Second),
					TCP: &ir.TCPHealthChecker{
						Send: &ir.HealthCheckPayload{
							Text: ptr.To("ping"),
						},
						Receive: &ir.HealthCheckPayload{
							Text: ptr.To("pong"),
						},
					},
				},
			},
		},
		{
			name: "health check with active gRPC check",
			policy: &egv1a1.ClusterSettings{
				HealthCheck: &egv1a1.HealthCheck{
					Active: &egv1a1.ActiveHealthCheck{
						Type:     egv1a1.ActiveHealthCheckerTypeGRPC,
						Timeout:  ptr.To(gwapiv1.Duration("2s")),
						Interval: ptr.To(gwapiv1.Duration("5s")),
						GRPC: &egv1a1.GRPCActiveHealthChecker{
							Service: ptr.To("health.check.Service"),
						},
					},
				},
			},
			want: &ir.HealthCheck{
				Active: &ir.ActiveHealthCheck{
					Timeout:  ir.MetaV1DurationPtr(2 * time.Second),
					Interval: ir.MetaV1DurationPtr(5 * time.Second),
					GRPC: &ir.GRPCHealthChecker{
						Service: ptr.To("health.check.Service"),
					},
				},
			},
		},
		{
			name: "health check with binary payload",
			policy: &egv1a1.ClusterSettings{
				HealthCheck: &egv1a1.HealthCheck{
					Active: &egv1a1.ActiveHealthCheck{
						Type:     egv1a1.ActiveHealthCheckerTypeTCP,
						Timeout:  ptr.To(gwapiv1.Duration("1s")),
						Interval: ptr.To(gwapiv1.Duration("10s")),
						TCP: &egv1a1.TCPActiveHealthChecker{
							Send: &egv1a1.ActiveHealthCheckPayload{
								Type:   egv1a1.ActiveHealthCheckPayloadTypeBinary,
								Binary: []byte{0x01, 0x02, 0x03},
							},
							Receive: &egv1a1.ActiveHealthCheckPayload{
								Type:   egv1a1.ActiveHealthCheckPayloadTypeBinary,
								Binary: []byte{0x04, 0x05, 0x06},
							},
						},
					},
				},
			},
			want: &ir.HealthCheck{
				Active: &ir.ActiveHealthCheck{
					Timeout:  ir.MetaV1DurationPtr(1 * time.Second),
					Interval: ir.MetaV1DurationPtr(10 * time.Second),
					TCP: &ir.TCPHealthChecker{
						Send: &ir.HealthCheckPayload{
							Binary: []byte{0x01, 0x02, 0x03},
						},
						Receive: &ir.HealthCheckPayload{
							Binary: []byte{0x04, 0x05, 0x06},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildHealthCheck(tt.policy)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestTranslateDNS(t *testing.T) {
	tests := []struct {
		name   string
		policy *egv1a1.ClusterSettings
		want   *ir.DNS
	}{
		{
			name:   "nil DNS",
			policy: &egv1a1.ClusterSettings{},
			want:   nil,
		},
		{
			name: "DNS with basic settings",
			policy: &egv1a1.ClusterSettings{
				DNS: &egv1a1.DNS{
					LookupFamily:  ptr.To(egv1a1.IPv4DNSLookupFamily),
					RespectDNSTTL: ptr.To(true),
				},
			},
			want: &ir.DNS{
				LookupFamily:  ptr.To(egv1a1.IPv4DNSLookupFamily),
				RespectDNSTTL: ptr.To(true),
			},
		},
		{
			name: "DNS with refresh rate",
			policy: &egv1a1.ClusterSettings{
				DNS: &egv1a1.DNS{
					DNSRefreshRate: ptr.To(gwapiv1.Duration("60s")),
				},
			},
			want: &ir.DNS{
				DNSRefreshRate: ir.MetaV1DurationPtr(60 * time.Second),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := translateDNS(tt.policy)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestBuildRetry(t *testing.T) {
	tests := []struct {
		name    string
		policy  *egv1a1.Retry
		want    *ir.Retry
		wantErr bool
	}{
		{
			name:   "nil retry",
			policy: nil,
			want:   nil,
		},
		{
			name: "retry with num retries",
			policy: &egv1a1.Retry{
				NumRetries: ptr.To(int32(3)),
			},
			want: &ir.Retry{
				NumRetries: ptr.To(uint32(3)),
			},
		},
		{
			name: "retry with retry on triggers",
			policy: &egv1a1.Retry{
				RetryOn: &egv1a1.RetryOn{
					Triggers: []egv1a1.TriggerEnum{"5xx", "reset", "timeout"},
				},
			},
			want: &ir.Retry{
				RetryOn: &ir.RetryOn{
					Triggers: []ir.TriggerEnum{"5xx", "reset", "timeout"},
				},
			},
		},
		{
			name: "retry with retry on HTTP status codes",
			policy: &egv1a1.Retry{
				RetryOn: &egv1a1.RetryOn{
					HTTPStatusCodes: []egv1a1.HTTPStatus{500, 502, 503},
				},
			},
			want: &ir.Retry{
				RetryOn: &ir.RetryOn{
					HTTPStatusCodes: []ir.HTTPStatus{500, 502, 503},
				},
			},
		},
		{
			name: "retry with per-retry timeout",
			policy: &egv1a1.Retry{
				PerRetry: &egv1a1.PerRetryPolicy{
					Timeout: ptr.To(gwapiv1.Duration("5s")),
				},
			},
			want: &ir.Retry{
				PerRetry: &ir.PerRetryPolicy{
					Timeout: ir.MetaV1DurationPtr(5 * time.Second),
				},
			},
		},
		{
			name: "retry with backoff policy",
			policy: &egv1a1.Retry{
				PerRetry: &egv1a1.PerRetryPolicy{
					BackOff: &egv1a1.BackOffPolicy{
						BaseInterval: ptr.To(gwapiv1.Duration("1s")),
						MaxInterval:  ptr.To(gwapiv1.Duration("10s")),
					},
				},
			},
			want: &ir.Retry{
				PerRetry: &ir.PerRetryPolicy{
					BackOff: &ir.BackOffPolicy{
						BaseInterval: ir.MetaV1DurationPtr(1 * time.Second),
						MaxInterval:  ir.MetaV1DurationPtr(10 * time.Second),
					},
				},
			},
		},
		{
			name: "invalid per-retry timeout",
			policy: &egv1a1.Retry{
				PerRetry: &egv1a1.PerRetryPolicy{
					Timeout: ptr.To(gwapiv1.Duration("invalid")),
				},
			},
			wantErr: true,
		},
		{
			name: "zero base interval",
			policy: &egv1a1.Retry{
				PerRetry: &egv1a1.PerRetryPolicy{
					BackOff: &egv1a1.BackOffPolicy{
						BaseInterval: ptr.To(gwapiv1.Duration("0s")),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "zero max interval",
			policy: &egv1a1.Retry{
				PerRetry: &egv1a1.PerRetryPolicy{
					BackOff: &egv1a1.BackOffPolicy{
						MaxInterval: ptr.To(gwapiv1.Duration("0s")),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "max interval less than base interval",
			policy: &egv1a1.Retry{
				PerRetry: &egv1a1.PerRetryPolicy{
					BackOff: &egv1a1.BackOffPolicy{
						BaseInterval: ptr.To(gwapiv1.Duration("10s")),
						MaxInterval:  ptr.To(gwapiv1.Duration("5s")),
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildRetry(tt.policy)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestTranslateActiveHealthCheckPayload(t *testing.T) {
	tests := []struct {
		name   string
		policy *egv1a1.ActiveHealthCheckPayload
		want   *ir.HealthCheckPayload
	}{
		{
			name:   "nil payload",
			policy: nil,
			want:   nil,
		},
		{
			name: "text payload",
			policy: &egv1a1.ActiveHealthCheckPayload{
				Type: egv1a1.ActiveHealthCheckPayloadTypeText,
				Text: ptr.To("hello"),
			},
			want: &ir.HealthCheckPayload{
				Text: ptr.To("hello"),
			},
		},
		{
			name: "binary payload",
			policy: &egv1a1.ActiveHealthCheckPayload{
				Type:   egv1a1.ActiveHealthCheckPayloadTypeBinary,
				Binary: []byte{0x01, 0x02, 0x03, 0x04},
			},
			want: &ir.HealthCheckPayload{
				Binary: []byte{0x01, 0x02, 0x03, 0x04},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := translateActiveHealthCheckPayload(tt.policy)
			require.Equal(t, tt.want, got)
		})
	}
}
