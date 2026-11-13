// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	bodysizelimitv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/body_size_limit/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/stretchr/testify/require"

	"github.com/envoyproxy/gateway/internal/ir"
)

func TestBodySizeLimitPatchHCM(t *testing.T) {
	t.Run("nil hcm", func(t *testing.T) {
		require.ErrorContains(t, (&bodySizeLimit{}).patchHCM(nil, &ir.HTTPListener{}), "hcm is nil")
	})

	t.Run("nil listener", func(t *testing.T) {
		require.ErrorContains(t, (&bodySizeLimit{}).patchHCM(&hcmv3.HttpConnectionManager{}, nil), "ir listener is nil")
	})

	t.Run("one disabled filter per distinct limit", func(t *testing.T) {
		mgr := &hcmv3.HttpConnectionManager{}
		listener := &ir.HTTPListener{
			Routes: []*ir.HTTPRoute{
				{Name: "no-limit"},
				{Name: "four-mi", Traffic: &ir.TrafficFeatures{RequestBodyLimit: &ir.RequestBodyLimit{MaxSize: 4194304}}},
				{Name: "four-mi-again", Traffic: &ir.TrafficFeatures{RequestBodyLimit: &ir.RequestBodyLimit{MaxSize: 4194304}}},
				{Name: "eight-mi", Traffic: &ir.TrafficFeatures{RequestBodyLimit: &ir.RequestBodyLimit{MaxSize: 8388608}}},
			},
		}

		require.NoError(t, (&bodySizeLimit{}).patchHCM(mgr, listener))
		// Patching the same listener again must not duplicate the filters.
		require.NoError(t, (&bodySizeLimit{}).patchHCM(mgr, listener))

		require.Len(t, mgr.HttpFilters, 2)
		for i, want := range []struct {
			name     string
			maxBytes uint64
		}{
			{name: "envoy.filters.http.body_size_limit/4194304", maxBytes: 4194304},
			{name: "envoy.filters.http.body_size_limit/8388608", maxBytes: 8388608},
		} {
			filter := mgr.HttpFilters[i]
			require.Equal(t, want.name, filter.Name)
			require.True(t, filter.Disabled)

			cfg := &bodysizelimitv3.BodySizeLimit{}
			require.NoError(t, filter.GetTypedConfig().UnmarshalTo(cfg))
			require.Equal(t, want.maxBytes, cfg.GetMaxRequestBytes().GetValue())
		}
	})

	t.Run("zero limit is rejected", func(t *testing.T) {
		mgr := &hcmv3.HttpConnectionManager{}
		listener := &ir.HTTPListener{
			Routes: []*ir.HTTPRoute{
				{Name: "zero", Traffic: &ir.TrafficFeatures{RequestBodyLimit: &ir.RequestBodyLimit{MaxSize: 0}}},
			},
		}

		require.ErrorContains(t, (&bodySizeLimit{}).patchHCM(mgr, listener), "must be greater than 0")
		require.Empty(t, mgr.HttpFilters)
	})
}

func TestBodySizeLimitPatchRoute(t *testing.T) {
	t.Run("nil route", func(t *testing.T) {
		require.ErrorContains(t, (&bodySizeLimit{}).patchRoute(nil, &ir.HTTPRoute{}, nil), "xds route is nil")
	})

	t.Run("nil ir route", func(t *testing.T) {
		require.ErrorContains(t, (&bodySizeLimit{}).patchRoute(&routev3.Route{}, nil, nil), "ir route is nil")
	})

	t.Run("route without a limit is left untouched", func(t *testing.T) {
		route := &routev3.Route{}
		require.NoError(t, (&bodySizeLimit{}).patchRoute(route, &ir.HTTPRoute{Name: "no-limit"}, nil))
		require.Empty(t, route.TypedPerFilterConfig)
	})

	t.Run("route with a limit enables the matching filter", func(t *testing.T) {
		route := &routev3.Route{}
		irRoute := &ir.HTTPRoute{
			Name:    "four-mi",
			Traffic: &ir.TrafficFeatures{RequestBodyLimit: &ir.RequestBodyLimit{MaxSize: 4194304}},
		}

		require.NoError(t, (&bodySizeLimit{}).patchRoute(route, irRoute, nil))

		perFilter := route.TypedPerFilterConfig["envoy.filters.http.body_size_limit/4194304"]
		require.NotNil(t, perFilter)

		cfg := &routev3.FilterConfig{}
		require.NoError(t, perFilter.UnmarshalTo(cfg))
		require.False(t, cfg.Disabled)
	})
}
