// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/stretchr/testify/require"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
)

func TestAlternateProtocolsCacheMixedBackendModes(t *testing.T) {
	tlsSetting := &ir.DestinationSetting{TLS: &ir.TLSUpstreamConfig{}}
	backendIndex := backendClusterIndex{
		"always": {Setting: tlsSetting, Traffic: &ir.ClusterTrafficFeatures{HTTP3: &ir.BackendHTTP3Settings{Mode: "Always"}}},
		"auto":   {Setting: tlsSetting, Traffic: &ir.ClusterTrafficFeatures{HTTP3: &ir.BackendHTTP3Settings{Mode: "Auto"}}},
		"plain":  {Setting: &ir.DestinationSetting{}},
	}
	for _, tc := range []struct {
		name             string
		backends         []string
		routeMode        string
		hasRouteSettings bool
		wantCache        bool
	}{
		{name: "Always then Auto", backends: []string{"always", "auto"}, wantCache: true},
		{name: "Auto then Always", backends: []string{"auto", "always"}, wantCache: true},
		{name: "only Always", backends: []string{"always"}},
		{name: "only Auto", backends: []string{"auto"}, wantCache: true},
		{name: "plain and unresolved backends", backends: []string{"plain", "missing"}},
		{name: "route Always and merged Auto", backends: []string{"auto"}, routeMode: "Always", hasRouteSettings: true, wantCache: true},
		{name: "route Auto and merged Always", backends: []string{"always"}, routeMode: "Auto", hasRouteSettings: true, wantCache: true},
		{name: "unused route Auto and merged Always", backends: []string{"always"}, routeMode: "Auto"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			route := &ir.HTTPRoute{Destination: &ir.RouteDestination{}}
			for _, backend := range tc.backends {
				route.Destination.BackendClusterRefs = append(route.Destination.BackendClusterRefs, &ir.BackendClusterRef{Name: backend})
			}
			if tc.hasRouteSettings {
				route.Destination.Settings = []*ir.DestinationSetting{tlsSetting}
			}
			if tc.routeMode != "" {
				route.Traffic = &ir.TrafficFeatures{ClusterTrafficFeatures: ir.ClusterTrafficFeatures{
					HTTP3: &ir.BackendHTTP3Settings{Mode: tc.routeMode},
				}}
			}

			tr := &Translator{backendIndex: backendIndex}
			listener := &ir.HTTPListener{Routes: []*ir.HTTPRoute{route}}
			hcm := &hcmv3.HttpConnectionManager{}
			require.NoError(t, tr.patchHCMWithAlternateProtocolsCache(hcm, listener))
			filterName := egv1a1.EnvoyFilterAlternateProtocolsCache.String()
			require.Equal(t, tc.wantCache, hcmContainsFilter(hcm, filterName))

			xdsRoute := &routev3.Route{}
			require.NoError(t, patchRouteWithAlternateProtocolsCache(xdsRoute, route, backendIndex))
			config, enabled := xdsRoute.TypedPerFilterConfig[filterName]
			require.Equal(t, tc.wantCache, enabled)
			if tc.wantCache {
				filterConfig := &routev3.FilterConfig{}
				require.NoError(t, config.UnmarshalTo(filterConfig))
				require.False(t, filterConfig.Disabled)
				require.True(t, hcm.HttpFilters[0].Disabled, "discovery must be enabled per route")
			}
		})
	}
}
