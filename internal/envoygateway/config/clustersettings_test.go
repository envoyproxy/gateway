// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package config

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/validation/field"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func TestValidateClusterSettings(t *testing.T) {
	testCases := []struct {
		name        string
		cs          *egv1a1.ClusterSettings
		expectedErr string
	}{
		{
			name: "nil",
		},
		{
			name: "valid",
			cs: &egv1a1.ClusterSettings{
				CircuitBreaker: &egv1a1.CircuitBreaker{
					MaxConnections:     new(int64(2048)),
					MaxPendingRequests: new(int64(2048)),
				},
				HealthCheck: &egv1a1.HealthCheck{
					Active: &egv1a1.ActiveHealthCheck{
						Type: egv1a1.ActiveHealthCheckerTypeGRPC,
					},
					Passive: &egv1a1.PassiveHealthCheck{},
				},
				Timeout: &egv1a1.Timeout{
					TCP: &egv1a1.TCPTimeout{ConnectTimeout: new(gwapiv1.Duration("1s"))},
				},
			},
		},
		{
			name: "OpenAPI: value out of range",
			cs: &egv1a1.ClusterSettings{
				CircuitBreaker: &egv1a1.CircuitBreaker{
					MaxConnections: new(int64(-1)),
				},
			},
			expectedErr: "rateLimit.serviceSettings.circuitBreaker.maxConnections: Invalid value: -1",
		},
		{
			name: "OpenAPI: unsupported enum value",
			cs: &egv1a1.ClusterSettings{
				HealthCheck: &egv1a1.HealthCheck{
					Active: &egv1a1.ActiveHealthCheck{Type: "Foo"},
				},
			},
			expectedErr: "rateLimit.serviceSettings.healthCheck.active.type: Unsupported value: \"Foo\"",
		},
		{
			name: "OpenAPI: invalid duration format",
			cs: &egv1a1.ClusterSettings{
				Timeout: &egv1a1.Timeout{
					TCP: &egv1a1.TCPTimeout{ConnectTimeout: new(gwapiv1.Duration("1 second"))},
				},
			},
			expectedErr: "rateLimit.serviceSettings.timeout.tcp.connectTimeout",
		},
		{
			name: "CEL: tcp health checker without tcp settings",
			cs: &egv1a1.ClusterSettings{
				HealthCheck: &egv1a1.HealthCheck{
					Active: &egv1a1.ActiveHealthCheck{Type: egv1a1.ActiveHealthCheckerTypeTCP},
				},
			},
			expectedErr: "If Health Checker type is TCP, tcp field needs to be set.",
		},
		{
			name: "CEL: proxyProtocol with http3",
			cs: &egv1a1.ClusterSettings{
				ProxyProtocol: &egv1a1.ProxyProtocol{Version: egv1a1.ProxyProtocolVersionV2},
				HTTP3:         &egv1a1.BackendHTTP3Settings{},
			},
			expectedErr: "proxyProtocol cannot be used together with http3",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateClusterSettings(field.NewPath("rateLimit", "serviceSettings"), tc.cs)
			if tc.expectedErr == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, tc.expectedErr)
		})
	}
}

func TestDefaultClusterSettings(t *testing.T) {
	cs := &egv1a1.ClusterSettings{
		HealthCheck: &egv1a1.HealthCheck{
			Active: &egv1a1.ActiveHealthCheck{
				Type: egv1a1.ActiveHealthCheckerTypeGRPC,
			},
		},
	}

	got, err := DefaultClusterSettings(cs)
	require.NoError(t, err)

	// The input is left untouched.
	require.Nil(t, cs.HealthCheck.Active.Timeout)

	require.Equal(t, new(gwapiv1.Duration("1s")), got.HealthCheck.Active.Timeout)
	require.Equal(t, new(gwapiv1.Duration("3s")), got.HealthCheck.Active.Interval)
	require.Equal(t, new(uint32(3)), got.HealthCheck.Active.UnhealthyThreshold)
	require.Equal(t, new(uint32(1)), got.HealthCheck.Active.HealthyThreshold)
}

func TestValidateRateLimitServiceSettings(t *testing.T) {
	testCases := []struct {
		name        string
		cs          *egv1a1.ClusterSettings
		expectedErr string
	}{
		{
			name: "nil",
		},
		{
			name: "supported fields",
			cs: &egv1a1.ClusterSettings{
				HealthCheck: &egv1a1.HealthCheck{
					Active: &egv1a1.ActiveHealthCheck{Type: egv1a1.ActiveHealthCheckerTypeGRPC},
				},
				LoadBalancer: &egv1a1.LoadBalancer{Type: egv1a1.RoundRobinLoadBalancerType},
				Timeout: &egv1a1.Timeout{
					HTTP: &egv1a1.HTTPTimeout{MaxConnectionDuration: new(gwapiv1.Duration("30s"))},
				},
			},
		},
		{
			name:        "http1",
			cs:          &egv1a1.ClusterSettings{HTTP1: &egv1a1.BackendHTTP1Settings{}},
			expectedErr: "rateLimit.serviceSettings.http1: Forbidden",
		},
		{
			name:        "http3",
			cs:          &egv1a1.ClusterSettings{HTTP3: &egv1a1.BackendHTTP3Settings{}},
			expectedErr: "rateLimit.serviceSettings.http3: Forbidden",
		},
		{
			name:        "proxyProtocol",
			cs:          &egv1a1.ClusterSettings{ProxyProtocol: &egv1a1.ProxyProtocol{Version: egv1a1.ProxyProtocolVersionV2}},
			expectedErr: "rateLimit.serviceSettings.proxyProtocol: Forbidden",
		},
		{
			name: "HTTP active health check",
			cs: &egv1a1.ClusterSettings{
				HealthCheck: &egv1a1.HealthCheck{
					Active: &egv1a1.ActiveHealthCheck{Type: egv1a1.ActiveHealthCheckerTypeHTTP},
				},
			},
			expectedErr: "rateLimit.serviceSettings.healthCheck.active.type: Unsupported value: \"HTTP\"",
		},
		{
			name: "HTTP request timeout",
			cs: &egv1a1.ClusterSettings{
				Timeout: &egv1a1.Timeout{
					HTTP: &egv1a1.HTTPTimeout{RequestTimeout: new(gwapiv1.Duration("1s"))},
				},
			},
			expectedErr: "rateLimit.serviceSettings.timeout.http.requestTimeout: Forbidden",
		},
		{
			name:        "consistent hash load balancer",
			cs:          &egv1a1.ClusterSettings{LoadBalancer: &egv1a1.LoadBalancer{Type: egv1a1.ConsistentHashLoadBalancerType}},
			expectedErr: "rateLimit.serviceSettings.loadBalancer.type: Forbidden",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateRateLimitServiceSettings(field.NewPath("rateLimit", "serviceSettings"), tc.cs)
			if tc.expectedErr == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, tc.expectedErr)
		})
	}
}
