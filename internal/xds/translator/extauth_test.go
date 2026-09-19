// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"
	"time"

	extauthv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ext_authz/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/envoyproxy/gateway/internal/ir"
)

func TestExtAuthConfigWithTimeout(t *testing.T) {
	tests := []struct {
		name            string
		extAuth         *ir.ExtAuth
		expectedTimeout *durationpb.Duration
	}{
		{
			name: "GRPC no timeout specified - should use default",
			extAuth: &ir.ExtAuth{
				Name: "test-extauth",
				GRPC: &ir.GRPCExtAuthService{
					Destination: ir.RouteDestination{Name: "test-cluster"},
					Authority:   "test-authority",
				},
			},
			expectedTimeout: durationpb.New(defaultExtServiceRequestTimeout),
		},
		{
			name: "GRPC custom timeout specified in milliseconds - should use custom timeout",
			extAuth: &ir.ExtAuth{
				Name: "test-extauth",
				GRPC: &ir.GRPCExtAuthService{
					Destination: ir.RouteDestination{Name: "test-cluster"},
					Authority:   "test-authority",
				},
				Timeout: &metav1.Duration{Duration: 500 * time.Millisecond},
			},
			expectedTimeout: durationpb.New(500 * time.Millisecond),
		},
		{
			name: "GRPC custom timeout specified in seconds - should use custom timeout",
			extAuth: &ir.ExtAuth{
				Name: "test-extauth",
				GRPC: &ir.GRPCExtAuthService{
					Destination: ir.RouteDestination{Name: "test-cluster"},
					Authority:   "test-authority",
				},
				Timeout: &metav1.Duration{Duration: 2 * time.Second},
			},
			expectedTimeout: durationpb.New(2 * time.Second),
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
			expectedTimeout: durationpb.New(1 * time.Second),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := extAuthConfig(tt.extAuth)
			require.NoError(t, err)
			require.NotNil(t, config)

			if tt.extAuth.GRPC != nil {
				// Test gRPC service timeout
				grpcService := config.Services.(*extauthv3.ExtAuthz_GrpcService).GrpcService
				assert.Equal(t, tt.expectedTimeout.Seconds, grpcService.Timeout.Seconds)
				assert.Equal(t, tt.expectedTimeout.Nanos, grpcService.Timeout.Nanos)
			} else if tt.extAuth.HTTP != nil {
				// Test HTTP service timeout
				httpService := config.Services.(*extauthv3.ExtAuthz_HttpService).HttpService
				assert.Equal(t, tt.expectedTimeout.Seconds, httpService.ServerUri.Timeout.Seconds)
				assert.Equal(t, tt.expectedTimeout.Nanos, httpService.ServerUri.Timeout.Nanos)
			}
		})
	}
}

func TestHttpServiceWithTimeout(t *testing.T) {
	tests := []struct {
		name            string
		httpService     *ir.HTTPExtAuthService
		timeout         *durationpb.Duration
		expectedTimeout *durationpb.Duration
	}{
		{
			name: "test HTTP service with timeout",
			httpService: &ir.HTTPExtAuthService{
				Destination: ir.RouteDestination{Name: "test-cluster"},
				Authority:   "test-authority",
				Path:        "/auth",
			},
			timeout:         durationpb.New(5 * time.Second),
			expectedTimeout: durationpb.New(5 * time.Second),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := httpService(tt.httpService, tt.timeout)
			require.NotNil(t, service)
			assert.Equal(t, tt.expectedTimeout.Seconds, service.ServerUri.Timeout.Seconds)
			assert.Equal(t, tt.expectedTimeout.Nanos, service.ServerUri.Timeout.Nanos)
		})
	}
}

func TestHTTPServiceAuthorizationResponseHeaders(t *testing.T) {
	headerMatcher := func(headers ...string) *matcherv3.ListStringMatcher {
		patterns := make([]*matcherv3.StringMatcher, 0, len(headers))
		for _, header := range headers {
			patterns = append(patterns, &matcherv3.StringMatcher{
				MatchPattern: &matcherv3.StringMatcher_Exact{
					Exact: header,
				},
				IgnoreCase: true,
			})
		}
		return &matcherv3.ListStringMatcher{Patterns: patterns}
	}

	tests := []struct {
		name                     string
		headersToBackend         []string
		headersToClientOnSuccess []string
		want                     *extauthv3.AuthorizationResponse
	}{
		{
			name: "no authorization response headers",
		},
		{
			name:             "headers to backend only",
			headersToBackend: []string{"x-current-user"},
			want: &extauthv3.AuthorizationResponse{
				AllowedUpstreamHeaders: headerMatcher("x-current-user"),
			},
		},
		{
			name:                     "headers to client on success only",
			headersToClientOnSuccess: []string{"Set-Cookie", "set-cookie"},
			want: &extauthv3.AuthorizationResponse{
				AllowedClientHeadersOnSuccess: headerMatcher(
					"Set-Cookie",
					"set-cookie",
				),
			},
		},
		{
			name:                     "headers to backend and client on success",
			headersToBackend:         []string{"x-current-user"},
			headersToClientOnSuccess: []string{"set-cookie"},
			want: &extauthv3.AuthorizationResponse{
				AllowedUpstreamHeaders:        headerMatcher("x-current-user"),
				AllowedClientHeadersOnSuccess: headerMatcher("set-cookie"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := httpService(
				&ir.HTTPExtAuthService{
					Destination:              ir.RouteDestination{Name: "test-cluster"},
					Authority:                "test-authority",
					Path:                     "/auth",
					HeadersToBackend:         tt.headersToBackend,
					HeadersToClientOnSuccess: tt.headersToClientOnSuccess,
				},
				durationpb.New(defaultExtServiceRequestTimeout),
			)

			require.NotNil(t, service)
			assert.Equal(t, tt.want, service.AuthorizationResponse)
		})
	}
}
