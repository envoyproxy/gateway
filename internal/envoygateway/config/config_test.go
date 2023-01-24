// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package config

import (
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
)

func TestValidate(t *testing.T) {
	cfg, err := New()
	require.NoError(t, err)

	testCases := []struct {
		name   string
		cfg    *Server
		expect bool
	}{
		{
			name:   "default",
			cfg:    cfg,
			expect: true,
		},
		{
			name: "empty namespace",
			cfg: &Server{
				EnvoyGateway: &v1alpha1.EnvoyGateway{
					EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
						Gateway:  v1alpha1.DefaultGateway(),
						Provider: v1alpha1.DefaultProvider(),
					},
				},
				Namespace: "",
			},
			expect: false,
		},
		{
			name: "unspecified envoy gateway",
			cfg: &Server{
				Namespace: "test-ns",
				Logger:    logr.Logger{},
			},
			expect: false,
		},
		{
			name: "unspecified gateway",
			cfg: &Server{
				EnvoyGateway: &v1alpha1.EnvoyGateway{
					EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
						Provider: v1alpha1.DefaultProvider(),
					},
				},
				Namespace: "test-ns",
			},
			expect: false,
		},
		{
			name: "unspecified provider",
			cfg: &Server{
				EnvoyGateway: &v1alpha1.EnvoyGateway{
					EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
						Gateway: v1alpha1.DefaultGateway(),
					},
				},
				Namespace: "test-ns",
			},
			expect: false,
		},
		{
			name: "empty gateway controllerName",
			cfg: &Server{
				EnvoyGateway: &v1alpha1.EnvoyGateway{
					EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
						Gateway:  &v1alpha1.Gateway{ControllerName: ""},
						Provider: v1alpha1.DefaultProvider(),
					},
				},
				Namespace: "test-ns",
			},
			expect: false,
		},
		{
			name: "unsupported provider",
			cfg: &Server{
				EnvoyGateway: &v1alpha1.EnvoyGateway{
					EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
						Gateway:  v1alpha1.DefaultGateway(),
						Provider: &v1alpha1.Provider{Type: v1alpha1.ProviderTypeFile},
					},
				},
				Namespace: "test-ns",
			},
			expect: false,
		},
		{
			name: "empty ratelimit",
			cfg: &Server{
				EnvoyGateway: &v1alpha1.EnvoyGateway{
					EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
						Gateway:   v1alpha1.DefaultGateway(),
						Provider:  v1alpha1.DefaultProvider(),
						RateLimit: &v1alpha1.RateLimit{},
					},
				},
				Namespace: "test-ns",
			},
			expect: false,
		},
		{
			name: "empty ratelimit redis setting",
			cfg: &Server{
				EnvoyGateway: &v1alpha1.EnvoyGateway{
					EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
						Gateway:  v1alpha1.DefaultGateway(),
						Provider: v1alpha1.DefaultProvider(),
						RateLimit: &v1alpha1.RateLimit{
							Backend: v1alpha1.RateLimitDatabaseBackend{
								Type:  v1alpha1.RedisBackendType,
								Redis: &v1alpha1.RateLimitRedisSettings{},
							},
						},
					},
				},
				Namespace: "test-ns",
			},
			expect: false,
		},
		{
			name: "unknown ratelimit redis url format",
			cfg: &Server{
				EnvoyGateway: &v1alpha1.EnvoyGateway{
					EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
						Gateway:  v1alpha1.DefaultGateway(),
						Provider: v1alpha1.DefaultProvider(),
						RateLimit: &v1alpha1.RateLimit{
							Backend: v1alpha1.RateLimitDatabaseBackend{
								Type: v1alpha1.RedisBackendType,
								Redis: &v1alpha1.RateLimitRedisSettings{
									URL: ":foo",
								},
							},
						},
					},
				},
				Namespace: "test-ns",
			},
			expect: false,
		},
		{
			name: "happy ratelimit redis settings",
			cfg: &Server{
				EnvoyGateway: &v1alpha1.EnvoyGateway{
					EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
						Gateway:  v1alpha1.DefaultGateway(),
						Provider: v1alpha1.DefaultProvider(),
						RateLimit: &v1alpha1.RateLimit{
							Backend: v1alpha1.RateLimitDatabaseBackend{
								Type: v1alpha1.RedisBackendType,
								Redis: &v1alpha1.RateLimitRedisSettings{
									URL: "localhost:6376",
								},
							},
						},
					},
				},
				Namespace: "test-ns",
			},
			expect: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.Validate()
			if !tc.expect {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
