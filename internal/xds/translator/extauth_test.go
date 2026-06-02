// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"
	"time"

	extauthv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ext_authz/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/envoyproxy/gateway/internal/ir"
)

func TestExtAuthCheckSettingsWithTimeout(t *testing.T) {
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
			config, err := extAuthCheckSettings(tt.extAuth)
			require.NoError(t, err)
			require.NotNil(t, config)

			if tt.extAuth.GRPC != nil {
				// Test gRPC service timeout
				grpcService := config.ServiceOverride.(*extauthv3.CheckSettings_GrpcService).GrpcService
				assert.Equal(t, tt.expectedTimeout.Seconds, grpcService.Timeout.Seconds)
				assert.Equal(t, tt.expectedTimeout.Nanos, grpcService.Timeout.Nanos)
			} else if tt.extAuth.HTTP != nil {
				// Test HTTP service timeout
				httpService := config.ServiceOverride.(*extauthv3.CheckSettings_HttpService).HttpService
				assert.Equal(t, tt.expectedTimeout.Seconds, httpService.ServerUri.Timeout.Seconds)
				assert.Equal(t, tt.expectedTimeout.Nanos, httpService.ServerUri.Timeout.Nanos)
			}
		})
	}
}

func TestExtAuthFilterName(t *testing.T) {
	failOpenFalse := false
	failOpenTrue := true
	recomputeRoute := true
	includeRouteMetadata := true
	statusOnError := int32(503)

	base := &ir.ExtAuth{
		Name:             "policy-a",
		FailOpen:         &failOpenFalse,
		HeadersToExtAuth: []string{"authorization"},
		GRPC: &ir.GRPCExtAuthService{
			Destination: ir.RouteDestination{Name: "cluster-a"},
			Authority:   "authority-a",
		},
		ContextExtensions: []*ir.ContextExtention{
			{Name: "tenant", Value: []byte("a")},
		},
		BodyToExtAuth: &ir.BodyToExtAuth{MaxRequestBytes: 1024},
	}

	baseName, _, err := extAuthFilterName(base)
	require.NoError(t, err)

	sameBucket := &ir.ExtAuth{
		Name:             "policy-b",
		FailOpen:         &failOpenFalse,
		HeadersToExtAuth: []string{"authorization"},
		HTTP: &ir.HTTPExtAuthService{
			Destination: ir.RouteDestination{Name: "cluster-b"},
			Authority:   "authority-b",
			Path:        "/auth",
		},
		ContextExtensions: []*ir.ContextExtention{
			{Name: "tenant", Value: []byte("b")},
		},
		BodyToExtAuth: &ir.BodyToExtAuth{MaxRequestBytes: 2048},
	}
	sameBucketName, _, err := extAuthFilterName(sameBucket)
	require.NoError(t, err)
	require.Equal(t, baseName, sameBucketName)

	tests := []struct {
		name    string
		extAuth *ir.ExtAuth
	}{
		{
			name: "different failOpen",
			extAuth: &ir.ExtAuth{
				FailOpen:         &failOpenTrue,
				HeadersToExtAuth: []string{"authorization"},
			},
		},
		{
			name: "different headersToExtAuth",
			extAuth: &ir.ExtAuth{
				FailOpen:         &failOpenFalse,
				HeadersToExtAuth: []string{"authorization", "x-tenant"},
			},
		},
		{
			name: "different recomputeRoute",
			extAuth: &ir.ExtAuth{
				FailOpen:         &failOpenFalse,
				HeadersToExtAuth: []string{"authorization"},
				RecomputeRoute:   &recomputeRoute,
			},
		},
		{
			name: "different includeRouteMetadata",
			extAuth: &ir.ExtAuth{
				FailOpen:             &failOpenFalse,
				HeadersToExtAuth:     []string{"authorization"},
				IncludeRouteMetadata: &includeRouteMetadata,
			},
		},
		{
			name: "different statusOnError",
			extAuth: &ir.ExtAuth{
				FailOpen:         &failOpenFalse,
				HeadersToExtAuth: []string{"authorization"},
				StatusOnError:    &statusOnError,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filterName, _, err := extAuthFilterName(tt.extAuth)
			require.NoError(t, err)
			require.NotEqual(t, baseName, filterName)
		})
	}
}

func TestExtAuthPatchHCMDeterministicOrder(t *testing.T) {
	failOpenFalse := false
	failOpenTrue := true
	routeA := &ir.HTTPRoute{
		Security: &ir.SecurityFeatures{
			ExtAuth: &ir.ExtAuth{
				FailOpen: &failOpenTrue,
				GRPC: &ir.GRPCExtAuthService{
					Destination: ir.RouteDestination{Name: "cluster-a"},
				},
			},
		},
	}
	routeB := &ir.HTTPRoute{
		Security: &ir.SecurityFeatures{
			ExtAuth: &ir.ExtAuth{
				FailOpen: &failOpenFalse,
				GRPC: &ir.GRPCExtAuthService{
					Destination: ir.RouteDestination{Name: "cluster-b"},
				},
			},
		},
	}

	expectedNames := []string{}
	for _, route := range []*ir.HTTPRoute{routeA, routeB} {
		filterName, _, err := extAuthFilterName(route.Security.ExtAuth)
		require.NoError(t, err)
		expectedNames = append(expectedNames, filterName)
	}
	if expectedNames[0] > expectedNames[1] {
		expectedNames[0], expectedNames[1] = expectedNames[1], expectedNames[0]
	}

	mgr := &hcmv3.HttpConnectionManager{}
	err := (&extAuth{}).patchHCM(mgr, &ir.HTTPListener{Routes: []*ir.HTTPRoute{routeA, routeB}})
	require.NoError(t, err)
	require.Len(t, mgr.HttpFilters, 2)
	require.Equal(t, expectedNames, []string{mgr.HttpFilters[0].Name, mgr.HttpFilters[1].Name})

	mgr = &hcmv3.HttpConnectionManager{}
	err = (&extAuth{}).patchHCM(mgr, &ir.HTTPListener{Routes: []*ir.HTTPRoute{routeB, routeA}})
	require.NoError(t, err)
	require.Len(t, mgr.HttpFilters, 2)
	require.Equal(t, expectedNames, []string{mgr.HttpFilters[0].Name, mgr.HttpFilters[1].Name})
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
