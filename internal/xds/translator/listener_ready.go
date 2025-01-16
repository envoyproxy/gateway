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
	"google.golang.org/protobuf/types/known/anypb"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/protocov"
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

	if ready.PrometheusCompressorType != nil {
		compressorCfg, err := filters.GenerateCompressorFilter(*ready.PrometheusCompressorType)
		if err != nil {
			return nil, err
		}

		hcmFilters = append(hcmFilters, compressorCfg)
	}

	routeCfg, err := buildReadyRouteConfig(ready.PrometheusEnabled, ready.PrometheusCompressorType != nil)
	if err != nil {
		return nil, err
	}

	router, err := filters.GenerateRouterFilter(false)
	if err != nil {
		return nil, err
	}
	hcmFilters = append(hcmFilters, router)

	hcm := &hcmv3.HttpConnectionManager{
		StatPrefix: "eg-ready-http",
		RouteSpecifier: &hcmv3.HttpConnectionManager_RouteConfig{
			RouteConfig: routeCfg,
		},
		HttpFilters: hcmFilters,
	}

	hcmAny, err := protocov.ToAnyWithValidation(hcm)
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
	}, nil
}

func buildReadyRouteConfig(promEnabled bool, compressorEnabled bool) (*routev3.RouteConfiguration, error) {
	if !promEnabled {
		return nil, nil
	}

	routeCfg := &routev3.RouteConfiguration{
		Name: "local_route",
	}

	if compressorEnabled {
		disabled, err := filters.GenerateCompressorPerRouteFilter(true)
		if err != nil {
			return nil, err
		}
		diabledAny, err := protocov.ToAnyWithValidation(disabled)
		if err != nil {
			return nil, err
		}
		// disable compression by default
		routeCfg.TypedPerFilterConfig = map[string]*anypb.Any{
			"envoy.filters.http.compression": diabledAny,
		}
	}

	vHosts := make([]*routev3.VirtualHost, 0, 1)
	if promEnabled {
		route := &routev3.Route{
			Match: &routev3.RouteMatch{
				PathSpecifier: &routev3.RouteMatch_Prefix{
					Prefix: "/stats/prometheus",
				},
			},
			Action: &routev3.Route_Route{
				Route: &routev3.RouteAction{
					ClusterSpecifier: &routev3.RouteAction_Cluster{
						Cluster: "prometheus_stats",
					},
				},
			},
		}

		if compressorEnabled {
			compressorCfg, err := filters.GenerateCompressorPerRouteFilter(false)
			if err != nil {
				return nil, err
			}

			anyCfg, err := protocov.ToAnyWithValidation(compressorCfg)
			if err != nil {
				return nil, err
			}

			// enable compression for this route(prefix: /stats/prometheus)
			route.TypedPerFilterConfig = map[string]*anypb.Any{
				"envoy.filters.http.compression": anyCfg,
			}
		}

		vhost := &routev3.VirtualHost{
			Name:    "prometheus_stats",
			Domains: []string{"*"},
			Routes:  []*routev3.Route{route},
		}

		vHosts = append(vHosts, vhost)
	}
	routeCfg.VirtualHosts = vHosts
	return routeCfg, nil
}
