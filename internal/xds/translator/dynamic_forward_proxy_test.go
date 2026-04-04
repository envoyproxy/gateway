// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/stretchr/testify/require"

	"github.com/envoyproxy/gateway/internal/ir"
)

func TestDynamicForwardProxyPatchRoutePreservesHostRewriteMetadata(t *testing.T) {
	tests := []struct {
		name         string
		buildIRRoute func() *ir.HTTPRoute
		buildAction  func() *routev3.RouteAction
		assertSpec   func(t *testing.T, action *routev3.RouteAction)
	}{
		{
			name: "header rewrite",
			buildIRRoute: func() *ir.HTTPRoute {
				header := "x-dynamic-host"
				return &ir.HTTPRoute{
					URLRewrite: &ir.URLRewrite{
						Host: &ir.HTTPHostModifier{
							Header: &header,
						},
					},
					Destination: &ir.RouteDestination{
						Settings: []*ir.DestinationSetting{
							{IsDynamicResolver: true},
						},
					},
				}
			},
			buildAction: func() *routev3.RouteAction {
				return &routev3.RouteAction{
					HostRewriteSpecifier: &routev3.RouteAction_HostRewriteHeader{
						HostRewriteHeader: "x-dynamic-host",
					},
					AppendXForwardedHost: true,
				}
			},
			assertSpec: func(t *testing.T, action *routev3.RouteAction) {
				t.Helper()
				spec, ok := action.HostRewriteSpecifier.(*routev3.RouteAction_HostRewriteHeader)
				require.True(t, ok)
				require.Equal(t, "x-dynamic-host", spec.HostRewriteHeader)
			},
		},
		{
			name: "literal rewrite",
			buildIRRoute: func() *ir.HTTPRoute {
				host := "www.google.com"
				return &ir.HTTPRoute{
					URLRewrite: &ir.URLRewrite{
						Host: &ir.HTTPHostModifier{
							Name: &host,
						},
					},
					Destination: &ir.RouteDestination{
						Settings: []*ir.DestinationSetting{
							{IsDynamicResolver: true},
						},
					},
				}
			},
			buildAction: func() *routev3.RouteAction {
				return &routev3.RouteAction{
					HostRewriteSpecifier: &routev3.RouteAction_HostRewriteLiteral{
						HostRewriteLiteral: "www.google.com",
					},
					AppendXForwardedHost: true,
				}
			},
			assertSpec: func(t *testing.T, action *routev3.RouteAction) {
				t.Helper()
				spec, ok := action.HostRewriteSpecifier.(*routev3.RouteAction_HostRewriteLiteral)
				require.True(t, ok)
				require.Equal(t, "www.google.com", spec.HostRewriteLiteral)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			irRoute := tt.buildIRRoute()
			routeAction := tt.buildAction()
			route := &routev3.Route{
				Action: &routev3.Route_Route{
					Route: routeAction,
				},
			}

			err := (&dynamicForwardProxy{}).patchRoute(route, irRoute, nil)
			require.NoError(t, err)

			filterName := dfpFilterName(dfpCacheName(determineIPFamily(irRoute.Destination.Settings), routeDNS(irRoute)))
			require.Contains(t, route.TypedPerFilterConfig, filterName)
			require.True(t, route.GetRoute().AppendXForwardedHost)
			tt.assertSpec(t, route.GetRoute())
		})
	}
}
