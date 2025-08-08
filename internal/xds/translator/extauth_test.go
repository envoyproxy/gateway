// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"
	"time"

	extauthv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ext_authz/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/envoyproxy/gateway/internal/ir"
)

func TestExtAuthConfigWithTimeout(t *testing.T) {
	tests := []struct {
		name            string
		extAuth         *ir.ExtAuth
		expectedTimeout int64
	}{
		{
			name: "no timeout specified - should use default",
			extAuth: &ir.ExtAuth{
				Name: "test-extauth",
				GRPC: &ir.GRPCExtAuthService{
					Destination: ir.RouteDestination{Name: "test-cluster"},
					Authority:   "test-authority",
				},
			},
			expectedTimeout: defaultExtServiceRequestTimeout,
		},
		{
			name: "custom timeout specified - should use custom timeout",
			extAuth: &ir.ExtAuth{
				Name: "test-extauth",
				GRPC: &ir.GRPCExtAuthService{
					Destination: ir.RouteDestination{Name: "test-cluster"},
					Authority:   "test-authority",
				},
				Timeout: &metav1.Duration{Duration: 500 * time.Millisecond},
			},
			expectedTimeout: 0, // 500ms = 0.5 seconds, rounds down to 0
		},
		{
			name: "custom timeout 2 seconds - should use custom timeout",
			extAuth: &ir.ExtAuth{
				Name: "test-extauth",
				GRPC: &ir.GRPCExtAuthService{
					Destination: ir.RouteDestination{Name: "test-cluster"},
					Authority:   "test-authority",
				},
				Timeout: &metav1.Duration{Duration: 2 * time.Second},
			},
			expectedTimeout: 2,
		},
		{
			name: "HTTP service with custom timeout",
			extAuth: &ir.ExtAuth{
				Name: "test-extauth",
				HTTP: &ir.HTTPExtAuthService{
					Destination: ir.RouteDestination{Name: "test-cluster"},
					Authority:   "test-authority",
					Path:        "/auth",
				},
				Timeout: &metav1.Duration{Duration: 1 * time.Second},
			},
			expectedTimeout: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := extAuthConfig(tt.extAuth)
			require.NotNil(t, config)

			if tt.extAuth.GRPC != nil {
				// Test gRPC service timeout
				grpcService := config.Services.(*extauthv3.ExtAuthz_GrpcService).GrpcService
				assert.Equal(t, tt.expectedTimeout, grpcService.Timeout.Seconds)
			} else if tt.extAuth.HTTP != nil {
				// Test HTTP service timeout
				httpService := config.Services.(*extauthv3.ExtAuthz_HttpService).HttpService
				assert.Equal(t, tt.expectedTimeout, httpService.ServerUri.Timeout.Seconds)
			}
		})
	}
}

func TestHttpServiceWithTimeout(t *testing.T) {
	tests := []struct {
		name            string
		httpService     *ir.HTTPExtAuthService
		timeoutSeconds  int64
		expectedTimeout int64
	}{
		{
			name: "test HTTP service with timeout",
			httpService: &ir.HTTPExtAuthService{
				Destination: ir.RouteDestination{Name: "test-cluster"},
				Authority:   "test-authority",
				Path:        "/auth",
			},
			timeoutSeconds:  5,
			expectedTimeout: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := httpService(tt.httpService, tt.timeoutSeconds)
			require.NotNil(t, service)
			assert.Equal(t, tt.expectedTimeout, service.ServerUri.Timeout.Seconds)
		})
	}
}
