// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	commondfpv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/common/dynamic_forward_proxy/v3"
	dfpv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/dynamic_forward_proxy/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

func init() {
	registerHTTPFilter(&dynamicForwardProxy{})
}

type dynamicForwardProxy struct{}

var _ httpFilter = &dynamicForwardProxy{}

// patchHCM appends the Dynamic Forward Proxy filter to the HTTP Connection Manager if applicable.
func (*dynamicForwardProxy) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	if mgr == nil {
		return errors.New("hcm is nil")
	}

	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	if !listenerRequireDFP(irListener) {
		return nil
	}

	if hcmContainsFilter(mgr, string(egv1a1.EnvoyFilterDynamicForwardProxy)) {
		return nil
	}

	cacheCfg, err := dynamicForwardProxyCacheConfig(irListener.Routes)
	if err != nil {
		return err
	}

	filterCfg := &dfpv3.FilterConfig{
		ImplementationSpecifier: &dfpv3.FilterConfig_DnsCacheConfig{
			DnsCacheConfig: cacheCfg,
		},
	}

	filterAny, err := anypb.New(filterCfg)
	if err != nil {
		return err
	}

	mgr.HttpFilters = append(mgr.HttpFilters, &hcmv3.HttpFilter{
		Name: string(egv1a1.EnvoyFilterDynamicForwardProxy),
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: filterAny,
		},
	})

	return nil
}

// patchRoute adds per-route host rewrite configuration so DNS resolution can
// use rewritten hostnames when configured.
func (*dynamicForwardProxy) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute, _ *ir.HTTPListener) error {
	if route == nil {
		return errors.New("xds route is nil")
	}

	if irRoute == nil {
		return errors.New("ir route is nil")
	}

	if !routeRequireDFP(irRoute) {
		return nil
	}

	perRouteCfg := buildDynamicForwardProxyPerRouteConfig(irRoute)
	if perRouteCfg == nil {
		return nil
	}

	perRouteAny, err := anypb.New(perRouteCfg)
	if err != nil {
		return err
	}

	if route.TypedPerFilterConfig == nil {
		route.TypedPerFilterConfig = make(map[string]*anypb.Any)
	}
	route.TypedPerFilterConfig[string(egv1a1.EnvoyFilterDynamicForwardProxy)] = perRouteAny

	// Clear out any existing host rewrite specifier to avoid conflicts.
	routeAction := route.GetRoute()
	if routeAction != nil {
		routeAction.HostRewriteSpecifier = nil
		routeAction.AppendXForwardedHost = false
	}

	return nil
}

func (*dynamicForwardProxy) patchResources(_ *types.ResourceVersionTable, _ []*ir.HTTPRoute) error {
	return nil
}

// listenerRequireDFP checks if a given listener requires the dynamic forward proxy filter.
func listenerRequireDFP(listener *ir.HTTPListener) bool {
	for _, route := range listener.Routes {
		if routeRequireDFP(route) {
			return true
		}
	}
	return false
}

// routeRequireDFP check if a given route requires the dynamic forward proxy filter.
// A dynamic forward proxy is required when:
// * The route has a dynamic resolver backend.
// * The route needs DFP to rewrite the host header using a header value.
func routeRequireDFP(route *ir.HTTPRoute) bool {
	if route == nil || route.Destination == nil {
		return false
	}

	routeHasDynamicResolver := false
	for _, setting := range route.Destination.Settings {
		if setting != nil && setting.IsDynamicResolver {
			routeHasDynamicResolver = true
			break
		}
	}

	if !routeHasDynamicResolver {
		return false
	}

	if route.URLRewrite != nil && route.URLRewrite.Host != nil && route.URLRewrite.Host.Header != nil {
		return true
	}

	return false
}

func dynamicForwardProxyCacheConfig(routes []*ir.HTTPRoute) (*commondfpv3.DnsCacheConfig, error) {
	var cacheCfg *commondfpv3.DnsCacheConfig

	for _, route := range routes {
		if !routeRequireDFP(route) {
			continue
		}

		var dns *ir.DNS
		if route.Traffic != nil {
			dns = route.Traffic.DNS
		}

		ipFamily := determineIPFamily(route.Destination.Settings)
		routeCfg := buildDynamicForwardProxyDNSCacheConfig(dns, computeDNSLookupFamily(ipFamily, dns))
		if cacheCfg == nil {
			cacheCfg = routeCfg
			continue
		}

		// TODO: better check this at the Gateway API validation layer.
		// Ensure all routes with dynamic resolver backends under this listener have consistent DNS cache settings.
		if !proto.Equal(cacheCfg, routeCfg) {
			return nil, fmt.Errorf("dynamic forward proxy requires consistent DNS cache settings per listener: %v vs %v", cacheCfg, routeCfg)
		}
	}

	return cacheCfg, nil
}

func buildDynamicForwardProxyPerRouteConfig(irRoute *ir.HTTPRoute) *dfpv3.PerRouteConfig {
	if irRoute.URLRewrite == nil || irRoute.URLRewrite.Host == nil || irRoute.URLRewrite.Host.Header == nil {
		return nil
	}

	return &dfpv3.PerRouteConfig{
		HostRewriteSpecifier: &dfpv3.PerRouteConfig_HostRewriteHeader{
			HostRewriteHeader: *irRoute.URLRewrite.Host.Header,
		},
	}
}
