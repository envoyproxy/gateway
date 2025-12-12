// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"
	"sort"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	commondfpv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/common/dynamic_forward_proxy/v3"
	dfpv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/dynamic_forward_proxy/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
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

	// Create a DFP filter for each unique DNS cache config needed by the listener's routes.
	// This is because DFP filter and Cluster must have consistent cache config.
	// Reference: https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/clusters/dynamic_forward_proxy/v3/cluster.proto
	cacheCfgs := dynamicForwardProxyCacheConfigs(irListener.Routes)

	for _, cacheCfg := range cacheCfgs {
		cacheName := cacheCfg.GetName()
		filterName := dynamicForwardProxyFilterName(cacheName)
		if hcmContainsFilter(mgr, filterName) {
			continue
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
			Name: filterName,
			ConfigType: &hcmv3.HttpFilter_TypedConfig{
				TypedConfig: filterAny,
			},
			Disabled: true,
		})
	}

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
	filterName := dynamicForwardProxyFilterName(dynamicForwardProxyCacheName(determineIPFamily(irRoute.Destination.Settings), routeDNS(irRoute)))
	route.TypedPerFilterConfig[filterName] = perRouteAny

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
// * The route needs DFP to rewrite the host header based on a header or literal name.
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

	if route.URLRewrite != nil && route.URLRewrite.Host != nil &&
		(route.URLRewrite.Host.Header != nil || route.URLRewrite.Host.Name != nil) {
		return true
	}

	return false
}

// dynamicForwardProxyCacheConfigs builds a sorted list of unique DFP DNS cache configs needed by the given routes.
func dynamicForwardProxyCacheConfigs(routes []*ir.HTTPRoute) []*commondfpv3.DnsCacheConfig {
	cacheCfgs := make(map[string]*commondfpv3.DnsCacheConfig)

	for _, route := range routes {
		if !routeRequireDFP(route) {
			continue
		}

		dns := routeDNS(route)
		ipFamily := determineIPFamily(route.Destination.Settings)
		cacheName := dynamicForwardProxyCacheName(ipFamily, dns)
		if _, existing := cacheCfgs[cacheName]; existing {
			continue
		}

		routeCfg := buildDynamicForwardProxyDNSCacheConfig(cacheName, dns, computeDNSLookupFamily(ipFamily, dns))
		cacheCfgs[cacheName] = routeCfg
	}

	if len(cacheCfgs) == 0 {
		return nil
	}

	cfgs := make([]*commondfpv3.DnsCacheConfig, 0, len(cacheCfgs))
	for _, cfg := range cacheCfgs {
		cfgs = append(cfgs, cfg)
	}

	// Sort the configs by name to ensure deterministic order for xDS generation.
	sort.Slice(cfgs, func(i, j int) bool {
		return cfgs[i].GetName() < cfgs[j].GetName()
	})

	return cfgs
}

func buildDynamicForwardProxyPerRouteConfig(irRoute *ir.HTTPRoute) *dfpv3.PerRouteConfig {
	switch {
	case irRoute.URLRewrite == nil || irRoute.URLRewrite.Host == nil:
		return nil
	case irRoute.URLRewrite.Host.Header != nil:
		return &dfpv3.PerRouteConfig{
			HostRewriteSpecifier: &dfpv3.PerRouteConfig_HostRewriteHeader{
				HostRewriteHeader: *irRoute.URLRewrite.Host.Header,
			},
		}
	case irRoute.URLRewrite.Host.Name != nil:
		return &dfpv3.PerRouteConfig{
			HostRewriteSpecifier: &dfpv3.PerRouteConfig_HostRewriteLiteral{
				HostRewriteLiteral: *irRoute.URLRewrite.Host.Name,
			},
		}
	default:
		return nil

	}
}

func routeDNS(route *ir.HTTPRoute) *ir.DNS {
	if route == nil || route.Traffic == nil {
		return nil
	}
	return route.Traffic.DNS
}

func dynamicForwardProxyFilterName(cacheName string) string {
	return fmt.Sprintf("%s.%s", string(egv1a1.EnvoyFilterDynamicForwardProxy), cacheName)
}
