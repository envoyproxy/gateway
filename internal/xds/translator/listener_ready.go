// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"fmt"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/proto"
	"github.com/envoyproxy/gateway/internal/xds/bootstrap"
	"github.com/envoyproxy/gateway/internal/xds/filters"
)

func buildReadyListener(ready *ir.ReadyListener) (*listenerv3.Listener, error) {
	ipv4Compact := false
	if ready.IPFamily == egv1a1.IPv6 || ready.IPFamily == egv1a1.DualStack {
		ipv4Compact = true
	}
	hcmFilters := make([]*hcmv3.HttpFilter, 0, 3)
	healthcheckFilter, err := filters.GenerateHealthCheckFilter(bootstrap.EnvoyReadinessPath)
	if err != nil {
		return nil, err
	}
	hcmFilters = append(hcmFilters, healthcheckFilter)

	router, err := filters.GenerateRouterFilter(false)
	if err != nil {
		return nil, err
	}
	hcmFilters = append(hcmFilters, router)

	hcm := &hcmv3.HttpConnectionManager{
		StatPrefix: "eg-ready-http",
		RouteSpecifier: &hcmv3.HttpConnectionManager_RouteConfig{
			RouteConfig: &routev3.RouteConfiguration{
				Name: "ready_route",
				VirtualHosts: []*routev3.VirtualHost{
					{
						Name:    "ready_route",
						Domains: []string{"*"},
						Routes: []*routev3.Route{
							{
								Match: &routev3.RouteMatch{
									PathSpecifier: &routev3.RouteMatch_Prefix{
										Prefix: "/", // match all
									},
								},
								Action: &routev3.Route_DirectResponse{
									DirectResponse: &routev3.DirectResponseAction{
										Status: 500, // you should not trigger this, healthcheck filter take care of it
									},
								},
							},
						},
					},
				},
			},
		},
		HttpFilters: hcmFilters,
	}

	hcmAny, err := proto.ToAnyWithValidation(hcm)
	if err != nil {
		return nil, err
	}

	return &listenerv3.Listener{
		Name: fmt.Sprintf("envoy-gateway-proxy-ready-%s-%d", ready.Address, ready.Port),
		Address: &corev3.Address{
			Address: &corev3.Address_SocketAddress{
				SocketAddress: &corev3.SocketAddress{
					Address: ready.Address,
					PortSpecifier: &corev3.SocketAddress_PortValue{
						PortValue: ready.Port,
					},
					Protocol:   corev3.SocketAddress_TCP,
					Ipv4Compat: ipv4Compact,
				},
			},
		},
		FilterChains: []*listenerv3.FilterChain{
			{
				Filters: []*listenerv3.Filter{
					{
						Name: wellknown.HTTPConnectionManager,
						ConfigType: &listenerv3.Filter_TypedConfig{
							TypedConfig: hcmAny,
						},
					},
				},
			},
		},
		BypassOverloadManager: true,
	}, nil
}
