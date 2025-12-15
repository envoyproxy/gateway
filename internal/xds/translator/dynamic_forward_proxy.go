// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"
	"sort"

	rbacconfigv3 "github.com/envoyproxy/go-control-plane/envoy/config/rbac/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	commondfpv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/common/dynamic_forward_proxy/v3"
	dfpv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/dynamic_forward_proxy/v3"
	rbacv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/rbac/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoymatcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"google.golang.org/protobuf/types/known/anypb"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

const (
	dfpLoopbackRBACFilterName = "envoy.filters.http.rbac.dfp_loopback_deny"
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

	// Add the RBAC filter once (no-op by default); per-route config will enable it only for DFP routes.
	if listenerHasDynamicResolverRoute(irListener) {
		loopbackRBAC, err := buildDFPLoopbackRBAC()
		if err != nil {
			return err
		}
		if !hcmContainsFilter(mgr, loopbackRBAC.Name) {
			mgr.HttpFilters = append([]*hcmv3.HttpFilter{loopbackRBAC}, mgr.HttpFilters...)
		}
	}

	if !listenerRequireDFP(irListener) {
		return nil
	}

	// Create a DFP filter for each unique DNS cache config needed by the listener's routes.
	// This is because DFP filter and Cluster must have consistent cache config.
	// Reference: https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/clusters/dynamic_forward_proxy/v3/cluster.proto
	cacheCfgs := dfpCacheConfigs(irListener.Routes)

	for _, cacheCfg := range cacheCfgs {
		cacheName := cacheCfg.GetName()
		filterName := dfpFilterName(cacheName)
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

	// Add per-route RBAC config to deny loopback addresses when DFP is used.
	if irRoute.IsDynamicResolverRoute() {
		hostFromLiteral := irRoute.URLRewrite != nil && irRoute.URLRewrite.Host != nil && irRoute.URLRewrite.Host.Name != nil
		hostFromHostHeaderOrCustomHeader := !hostFromLiteral
		// We don't enforce check for host rewrite from literal since it's static and known at config time.
		// The loopback check is mainly to prevent dynamic hostnames that may resolve to loopback addresses.
		if hostFromHostHeaderOrCustomHeader {
			rbacPerRouteCfg := buildDFPLoopbackRBACPerRoute(irRoute)
			rbacAny, err := anypb.New(rbacPerRouteCfg)
			if err != nil {
				return err
			}
			if route.TypedPerFilterConfig == nil {
				route.TypedPerFilterConfig = make(map[string]*anypb.Any)
			}
			route.TypedPerFilterConfig[dfpLoopbackRBACFilterName] = rbacAny
		}
	}

	if !routeRequireDFP(irRoute) {
		return nil
	}

	perRouteCfg := buildDFPPerRouteConfig(irRoute)
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
	filterName := dfpFilterName(dfpCacheName(determineIPFamily(irRoute.Destination.Settings), routeDNS(irRoute)))
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

// listenerHasDynamicResolverRoute checks if a given listener has any route with a dynamic resolver backend.
func listenerHasDynamicResolverRoute(listener *ir.HTTPListener) bool {
	for _, route := range listener.Routes {
		if route.IsDynamicResolverRoute() {
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

	if !route.IsDynamicResolverRoute() {
		return false
	}

	if route.URLRewrite != nil && route.URLRewrite.Host != nil &&
		(route.URLRewrite.Host.Header != nil || route.URLRewrite.Host.Name != nil) {
		return true
	}

	return false
}

// dfpCacheConfigs builds a sorted list of unique DFP DNS cache configs needed by the given routes.
func dfpCacheConfigs(routes []*ir.HTTPRoute) []*commondfpv3.DnsCacheConfig {
	cacheCfgs := make(map[string]*commondfpv3.DnsCacheConfig)

	for _, route := range routes {
		if !routeRequireDFP(route) {
			continue
		}

		dns := routeDNS(route)
		ipFamily := determineIPFamily(route.Destination.Settings)
		cacheName := dfpCacheName(ipFamily, dns)
		if _, existing := cacheCfgs[cacheName]; existing {
			continue
		}

		routeCfg := buildDFPDNSCacheConfig(cacheName, dns, computeDNSLookupFamily(ipFamily, dns))
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

func buildDFPPerRouteConfig(irRoute *ir.HTTPRoute) *dfpv3.PerRouteConfig {
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

func dfpFilterName(cacheName string) string {
	return fmt.Sprintf("%s.%s", string(egv1a1.EnvoyFilterDynamicForwardProxy), cacheName)
}

func buildDFPLoopbackRBAC() (*hcmv3.HttpFilter, error) {
	rbac := &rbacv3.RBAC{
		// No policies: acts as a placeholder; per-route config enables denial.
	}

	rbacAny, err := anypb.New(rbac)
	if err != nil {
		return nil, err
	}

	return &hcmv3.HttpFilter{
		Name: dfpLoopbackRBACFilterName,
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: rbacAny,
		},
	}, nil
}

func buildDFPLoopbackRBACPerRoute(irRoute *ir.HTTPRoute) *rbacv3.RBACPerRoute {
	loopbackPatterns := []string{
		`^127\.0\.0\.1(?::\d+)?$`,
		`^localhost(?::\d+)?$`,
		`^localhost\.localdomain(?::\d+)?$`,
		`^ip6-localhost(?::\d+)?$`,
		`^ip6-loopback(?::\d+)?$`,
		`^\[::1\](?::\d+)?$`,
		`^::1(?::\d+)?$`,
	}

	permissions := make([]*rbacconfigv3.Permission, 0, len(loopbackPatterns))

	hostFromCustomHeader := irRoute.URLRewrite != nil && irRoute.URLRewrite.Host != nil && irRoute.URLRewrite.Host.Header != nil
	if hostFromCustomHeader {
		for _, pattern := range loopbackPatterns {
			permissions = append(permissions, buildHeaderPermissionRegex(*irRoute.URLRewrite.Host.Header, pattern))
		}
	} else {
		for _, pattern := range loopbackPatterns {
			permissions = append(permissions, buildHeaderPermissionRegex(":authority", pattern))
		}
	}

	return &rbacv3.RBACPerRoute{
		Rbac: &rbacv3.RBAC{
			Rules: &rbacconfigv3.RBAC{
				Action: rbacconfigv3.RBAC_DENY,
				Policies: map[string]*rbacconfigv3.Policy{
					"deny-loopback-host": {
						Permissions: permissions,
						Principals: []*rbacconfigv3.Principal{
							{
								Identifier: &rbacconfigv3.Principal_Any{
									Any: true,
								},
							},
						},
					},
				},
			},
		},
	}
}

func buildHeaderPermissionRegex(name, pattern string) *rbacconfigv3.Permission {
	return &rbacconfigv3.Permission{
		Rule: &rbacconfigv3.Permission_Header{
			Header: &routev3.HeaderMatcher{
				Name: name,
				HeaderMatchSpecifier: &routev3.HeaderMatcher_SafeRegexMatch{
					SafeRegexMatch: &envoymatcherv3.RegexMatcher{
						Regex: pattern,
					},
				},
			},
		},
	}
}
