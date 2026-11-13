// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"
	"time"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/envoyproxy/gateway/internal/ir"
)

func TestBuildAdaptiveConcurrencyProtoDefaultsConcurrencyUpdateInterval(t *testing.T) {
	ac := buildAdaptiveConcurrencyProto(&ir.AdaptiveConcurrency{})
	params := ac.GetGradientControllerConfig().GetConcurrencyLimitParams()
	require.NotNil(t, params)
	require.Equal(t, defaultAdaptiveConcurrencyUpdateInterval, params.GetConcurrencyUpdateInterval().AsDuration())
	require.Nil(t, params.GetMinConcurrencyLimit())
	require.Nil(t, params.GetMaxConcurrencyLimit())
	require.NotNil(t, ac.GetGradientControllerConfig().GetMinRttCalcParams())
	require.Nil(t, ac.GetGradientControllerConfig().GetSampleAggregatePercentile())
}

func TestBuildAdaptiveConcurrencyProtoUsesConfiguredFields(t *testing.T) {
	ac := buildAdaptiveConcurrencyProto(&ir.AdaptiveConcurrency{
		SampleAggregatePercentile:      new(float64(90)),
		MinConcurrencyLimit:            new(uint32(10)),
		MaxConcurrencyLimit:            new(uint32(500)),
		ConcurrencyUpdateInterval:      &metav1.Duration{Duration: 200 * time.Millisecond},
		ConcurrencyLimitExceededStatus: new(int32(429)),
	})

	params := ac.GetGradientControllerConfig().GetConcurrencyLimitParams()
	require.NotNil(t, params)
	require.Equal(t, 200*time.Millisecond, params.GetConcurrencyUpdateInterval().AsDuration())
	require.Equal(t, uint32(10), params.GetMinConcurrencyLimit().GetValue())
	require.Equal(t, uint32(500), params.GetMaxConcurrencyLimit().GetValue())
	require.Equal(t, float64(90), ac.GetGradientControllerConfig().GetSampleAggregatePercentile().GetValue())
	require.Equal(t, int32(429), int32(ac.GetConcurrencyLimitExceededStatus().GetCode()))
}

func TestAdaptiveConcurrencyPatchHCM(t *testing.T) {
	filter := &adaptiveConcurrency{}

	t.Run("nil hcm", func(t *testing.T) {
		require.EqualError(t, filter.patchHCM(nil, &ir.HTTPListener{}), "hcm is nil")
	})
	t.Run("nil listener", func(t *testing.T) {
		require.EqualError(t, filter.patchHCM(&hcmv3.HttpConnectionManager{}, nil), "ir listener is nil")
	})
	t.Run("no adaptive concurrency", func(t *testing.T) {
		mgr := &hcmv3.HttpConnectionManager{}
		require.NoError(t, filter.patchHCM(mgr, &ir.HTTPListener{Routes: []*ir.HTTPRoute{{}}}))
		require.Empty(t, mgr.HttpFilters)
	})
	t.Run("adds disabled filter once", func(t *testing.T) {
		ac := &ir.AdaptiveConcurrency{Name: "backendtrafficpolicy/default/policy"}
		route := &ir.HTTPRoute{Traffic: &ir.TrafficFeatures{AdaptiveConcurrency: ac}}
		mgr := &hcmv3.HttpConnectionManager{}
		listener := &ir.HTTPListener{Routes: []*ir.HTTPRoute{route, route}}
		require.NoError(t, filter.patchHCM(mgr, listener))
		require.Len(t, mgr.HttpFilters, 1)
		require.True(t, mgr.HttpFilters[0].Disabled)
		require.Equal(t, "envoy.filters.http.adaptive_concurrency/backendtrafficpolicy/default/policy", mgr.HttpFilters[0].Name)
	})
}

func TestAdaptiveConcurrencyPatchRoute(t *testing.T) {
	filter := &adaptiveConcurrency{}

	t.Run("nil route", func(t *testing.T) {
		require.EqualError(t, filter.patchRoute(nil, &ir.HTTPRoute{}, nil), "xds route is nil")
	})
	t.Run("nil ir route", func(t *testing.T) {
		require.EqualError(t, filter.patchRoute(&routev3.Route{}, nil, nil), "ir route is nil")
	})
	t.Run("no adaptive concurrency", func(t *testing.T) {
		route := &routev3.Route{}
		require.NoError(t, filter.patchRoute(route, &ir.HTTPRoute{}, nil))
		require.Empty(t, route.TypedPerFilterConfig)
	})
	t.Run("enables filter", func(t *testing.T) {
		irRoute := &ir.HTTPRoute{
			Traffic: &ir.TrafficFeatures{
				AdaptiveConcurrency: &ir.AdaptiveConcurrency{Name: "backendtrafficpolicy/default/policy"},
			},
		}
		route := &routev3.Route{}
		require.NoError(t, filter.patchRoute(route, irRoute, nil))
		require.Contains(t, route.TypedPerFilterConfig, "envoy.filters.http.adaptive_concurrency/backendtrafficpolicy/default/policy")
		err := filter.patchRoute(route, irRoute, nil)
		require.ErrorContains(t, err, "route already contains filter config: envoy.filters.http.adaptive_concurrency/backendtrafficpolicy/default/policy")
	})
}

func TestRouteContainsAdaptiveConcurrency(t *testing.T) {
	require.False(t, routeContainsAdaptiveConcurrency(nil))
	require.False(t, routeContainsAdaptiveConcurrency(&ir.HTTPRoute{}))
	require.True(t, routeContainsAdaptiveConcurrency(&ir.HTTPRoute{
		Traffic: &ir.TrafficFeatures{AdaptiveConcurrency: &ir.AdaptiveConcurrency{}},
	}))
}

func TestAdaptiveConcurrencyPatchResources(t *testing.T) {
	require.NoError(t, (&adaptiveConcurrency{}).patchResources(nil, nil))
}
