// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"
	"time"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	extauthv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ext_authz/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
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
					Destination: ir.RouteDestination{
						Name: "test-cluster",
						Settings: []*ir.DestinationSetting{{
							Weight:    new(uint32(1)),
							Endpoints: []*ir.DestinationEndpoint{{Host: "1.2.3.4", Port: 8080}},
						}},
					},
					Authority: "test-authority",
				},
			},
			expectedTimeout: durationpb.New(defaultExtServiceRequestTimeout),
		},
		{
			name: "GRPC custom timeout specified in milliseconds - should use custom timeout",
			extAuth: &ir.ExtAuth{
				Name: "test-extauth",
				GRPC: &ir.GRPCExtAuthService{
					Destination: ir.RouteDestination{
						Name: "test-cluster",
						Settings: []*ir.DestinationSetting{{
							Weight:    new(uint32(1)),
							Endpoints: []*ir.DestinationEndpoint{{Host: "1.2.3.4", Port: 8080}},
						}},
					},
					Authority: "test-authority",
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
					Destination: ir.RouteDestination{
						Name: "test-cluster",
						Settings: []*ir.DestinationSetting{{
							Weight:    new(uint32(1)),
							Endpoints: []*ir.DestinationEndpoint{{Host: "1.2.3.4", Port: 8080}},
						}},
					},
					Authority: "test-authority",
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
					Destination: ir.RouteDestination{
						Name: "test-cluster",
						Settings: []*ir.DestinationSetting{{
							Weight:    new(uint32(1)),
							Endpoints: []*ir.DestinationEndpoint{{Host: "1.2.3.4", Port: 8080}},
						}},
					},
					Authority: "test-authority",
					Path:      "/auth",
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
			service := httpService(tt.httpService, "test-cluster", tt.timeout)
			require.NotNil(t, service)
			assert.Equal(t, tt.expectedTimeout.Seconds, service.ServerUri.Timeout.Seconds)
			assert.Equal(t, tt.expectedTimeout.Nanos, service.ServerUri.Timeout.Nanos)
		})
	}
}

// extAuthForBackend builds an ExtAuth pointing at the given backend host/port,
// with a policy-derived name and context extensions that vary per policy but
// must not affect deduplication (they are applied as a per-route override).
func extAuthForBackend(policyName, host string, port uint32, ctxExtVal string) *ir.ExtAuth {
	authority := host
	return &ir.ExtAuth{
		Name: policyName,
		HTTP: &ir.HTTPExtAuthService{
			Destination: ir.RouteDestination{
				// Name is policy-derived today; the dedup must not depend on it.
				Name: policyName + "/extauth/0",
				Settings: []*ir.DestinationSetting{{
					Weight:    new(uint32(1)),
					Endpoints: []*ir.DestinationEndpoint{{Host: host, Port: port}},
				}},
			},
			Authority: authority,
			Path:      "/auth",
		},
		ContextExtensions: []*ir.ContextExtention{{Name: "k", Value: ir.PrivateBytes(ctxExtVal)}},
	}
}

// clusterRefOf returns the upstream cluster name referenced by the built
// ext_authz filter config for the given ExtAuth.
func clusterRefOf(t *testing.T, extAuth *ir.ExtAuth) string {
	t.Helper()
	cfg, err := extAuthConfig(extAuth)
	require.NoError(t, err)
	return cfg.Services.(*extauthv3.ExtAuthz_HttpService).HttpService.ServerUri.GetCluster()
}

// TestExtAuthDeduplicatesIdenticalClusters verifies that two SecurityPolicies
// pointing at the same ext auth backend collapse onto a single upstream cluster
// (named by a content hash), even though each policy keeps its own ext_authz
// filter. A policy with a different backend must not be collapsed.
func TestExtAuthDeduplicatesIdenticalClusters(t *testing.T) {
	f := &extAuth{}

	// Two distinct policies, identical backend, different context extensions.
	routeA := &ir.HTTPRoute{Name: "route-a", Security: &ir.SecurityFeatures{
		ExtAuth: extAuthForBackend("securitypolicy/ns/a", "auth.example.com", 443, "a"),
	}}
	routeB := &ir.HTTPRoute{Name: "route-b", Security: &ir.SecurityFeatures{
		ExtAuth: extAuthForBackend("securitypolicy/ns/b", "auth.example.com", 443, "b"),
	}}
	// A third policy pointing at a different backend must stay separate.
	routeC := &ir.HTTPRoute{Name: "route-c", Security: &ir.SecurityFeatures{
		ExtAuth: extAuthForBackend("securitypolicy/ns/c", "other.example.com", 443, "c"),
	}}

	// Filters stay per-policy (dedup is by policy-derived name): three policies ⇒ three filters.
	mgr := &hcmv3.HttpConnectionManager{}
	require.NoError(t, f.patchHCM(mgr, &ir.HTTPListener{Routes: []*ir.HTTPRoute{routeA, routeB, routeC}}))
	require.Len(t, mgr.HttpFilters, 3, "filters remain per-policy")

	// Clusters dedup by backend content: A and B share one; C is distinct ⇒ two clusters.
	tCtx := &types.ResourceVersionTable{XdsResources: types.XdsResources{}}
	require.NoError(t, f.patchResources(tCtx, []*ir.HTTPRoute{routeA, routeB, routeC}))
	require.Len(t, tCtx.XdsResources[resourcev3.ClusterType], 2, "identical backends should collapse to one cluster")

	// Both A and B filters reference the same shared cluster; C references a different one.
	refA := clusterRefOf(t, routeA.Security.ExtAuth)
	refB := clusterRefOf(t, routeB.Security.ExtAuth)
	refC := clusterRefOf(t, routeC.Security.ExtAuth)
	require.Equal(t, refA, refB, "identical backends must reference the same cluster")
	require.NotEqual(t, refA, refC, "distinct backends must reference distinct clusters")

	// Each policy keeps its own filter and its own context extensions.
	for _, tc := range []struct {
		route  *ir.HTTPRoute
		expect string
	}{{routeA, "a"}, {routeB, "b"}} {
		filterName := extAuthFilterName(tc.route.Security.ExtAuth)
		xdsRoute := &routev3.Route{}
		require.NoError(t, f.patchRoute(xdsRoute, tc.route, nil))
		require.Contains(t, xdsRoute.TypedPerFilterConfig, filterName)
		perRoute := &extauthv3.ExtAuthzPerRoute{}
		require.NoError(t, xdsRoute.TypedPerFilterConfig[filterName].UnmarshalTo(perRoute))
		assert.Equal(t, tc.expect, perRoute.GetCheckSettings().GetContextExtensions()["k"])
	}
}

// TestExtAuthClusterNameStableAcrossEndpointChurn verifies that the deduplicated
// cluster name is derived from the backend identity and settings, not the
// resolved endpoint membership. An EDS backend whose endpoints change (scale or
// rollout) must keep the same cluster name so Envoy applies an EDS update instead
// of recreating the cluster (which would reset stats and connection pools).
func TestExtAuthClusterNameStableAcrossEndpointChurn(t *testing.T) {
	base := extAuthForBackend("securitypolicy/ns/a", "auth.example.com", 443, "a")

	// Same backend/authority/settings, but different resolved endpoints.
	churned := extAuthForBackend("securitypolicy/ns/a", "auth.example.com", 443, "a")
	churned.HTTP.Destination.Settings[0].Endpoints = []*ir.DestinationEndpoint{
		{Host: "10.0.0.9", Port: 443},
		{Host: "10.0.0.10", Port: 443},
	}

	nameBase, err := extServiceClusterName(extAuthClusterPrefix, base.HTTP.Authority, &base.HTTP.Destination, base.Traffic)
	require.NoError(t, err)
	nameChurned, err := extServiceClusterName(extAuthClusterPrefix, churned.HTTP.Authority, &churned.HTTP.Destination, churned.Traffic)
	require.NoError(t, err)
	require.Equal(t, nameBase, nameChurned, "cluster name must not change when only the endpoints change")

	// A genuinely different setting (protocol) must still produce a different name.
	differentSettings := extAuthForBackend("securitypolicy/ns/a", "auth.example.com", 443, "a")
	differentSettings.HTTP.Destination.Settings[0].Protocol = ir.HTTP2
	nameDifferent, err := extServiceClusterName(extAuthClusterPrefix, differentSettings.HTTP.Authority, &differentSettings.HTTP.Destination, differentSettings.Traffic)
	require.NoError(t, err)
	require.NotEqual(t, nameBase, nameDifferent, "cluster name must change when cluster settings change")
}
