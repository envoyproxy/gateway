// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"
	"slices"
	"strconv"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	filterchainv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/filter_chain/v3"
	luafilterv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/lua/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

func init() {
	registerHTTPFilter(&lua{})
}

type lua struct{}

var _ httpFilter = &lua{}

// patchHCM adds disabled envoy.filters.http.filter_chain placeholder filters to the HTTP
// Connection Manager: one for per-listener (per-connection) Lua and one for per-route Lua.
//
// Both placeholders are added together as soon as either scope has a Lua policy anywhere on
// this listener, even if the other scope currently has none. This keeps the HCM's filter set
// stable across that kind of policy churn too: e.g. adding a per-listener Lua policy later to
// a listener that already has per-route Lua only changes route/virtual host
// TypedPerFilterConfig (an RDS update), never the listener's filter list (which would require
// an LDS update and a connection drain).
//
// Envoy's LuaPerRoute API can only override one script for one filter instance, while EG's
// EnvoyExtensionPolicy API allows an ordered list of Lua filters per listener/route. The
// filter_chain filter wraps an ordered, named sub-chain of Lua filters that is supplied
// separately (per virtual host for listener-scoped Lua, per route for route-scoped Lua).
func (*lua) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	if mgr == nil {
		return errors.New("hcm is nil")
	}
	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	hasListenerLua := irListener.EnvoyExtensions != nil && len(irListener.EnvoyExtensions.Luas) > 0
	hasRouteLua := slices.ContainsFunc(irListener.Routes, routeContainsLua)
	if !hasListenerLua && !hasRouteLua {
		return nil
	}

	for _, filterName := range []string{luaListenerFCFilterName(), luaFCFilterName()} {
		if hcmContainsFilter(mgr, filterName) {
			continue
		}
		filter, err := buildHCMFilterChainFilter(filterName)
		if err != nil {
			return err
		}
		mgr.HttpFilters = append(mgr.HttpFilters, filter)
	}

	return nil
}

// luaFCFilterName returns the stable HCM-level filter name for the per-route Lua
// filter_chain placeholder.
func luaFCFilterName() string {
	return FilterChainFilterNamePrefixForEEP + "lua"
}

// luaListenerFCFilterName returns the stable HCM-level filter name for the per-listener
// (per-connection) Lua filter_chain placeholder.
func luaListenerFCFilterName() string {
	return FilterChainFilterNamePrefixForEEP + "lua.listener"
}

func buildHCMFilterChainFilter(filterName string) (*hcmv3.HttpFilter, error) {
	var (
		fcProto *filterchainv3.FilterChainConfig
		fcAny   *anypb.Any
		err     error
	)
	fcProto = &filterchainv3.FilterChainConfig{}

	if err = fcProto.ValidateAll(); err != nil {
		return nil, err
	}
	if fcAny, err = anypb.New(fcProto); err != nil {
		return nil, err
	}

	return &hcmv3.HttpFilter{
		Name:     filterName,
		Disabled: true,
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: fcAny,
		},
	}, nil
}

// luaFilterName returns the stable top-level filter name for the per-route Lua slot index.
// The index is the execution slot within the ordered EnvoyExtensionPolicy Lua
// list, so route 0th scripts always bind to the same listener-level filter.
func luaFilterName(idx int) string {
	return perRouteFilterName(egv1a1.EnvoyFilterLua, strconv.Itoa(idx))
}

// luaListenerFilterName returns the stable HCM-level filter name for a listener-level
// Lua slot. Using the envoy.filters.http.lua prefix (instead of the raw policy name)
// ensures sortHTTPFilters assigns it the correct order relative to route-level slots.
func luaListenerFilterName(idx int) string {
	return fmt.Sprintf("%s/listener/%d", egv1a1.EnvoyFilterLua, idx)
}

// routeContainsLua returns true if the route has any Lua extensions.
func routeContainsLua(irRoute *ir.HTTPRoute) bool {
	if irRoute == nil {
		return false
	}
	return irRoute.EnvoyExtensions != nil && len(irRoute.EnvoyExtensions.Luas) > 0
}

// patchResources patches the cluster resources for the http lua code source.
func (*lua) patchResources(_ *types.ResourceVersionTable, _ []*ir.HTTPRoute) error {
	return nil
}

// patchRoute patches the provided route with LuaPerRoute so the Lua filter runs with the route's script.
// Routes with no Lua entries fall back to the listener-level Lua inherited from the virtual host.
// Only routes with their own Lua entries disable the inherited listener-level Lua and install
// their own scripts in its place.
func (*lua) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute, irListener *ir.HTTPListener) error {
	if route == nil {
		return errors.New("xds route is nil")
	}
	if irRoute == nil {
		return errors.New("ir route is nil")
	}
	if irRoute.EnvoyExtensions == nil {
		return nil
	}

	// Disable the inherited listener-level Lua whenever EnvoyExtensions was set by a
	// more-specific route policy (FromGatewayPolicy is false/nil). The extension count
	// is intentionally not checked here: an empty result (e.g. fail-open invalid Wasm)
	// still represents a more-specific policy that owns this route and must suppress the
	// lower-scope Lua. When FromGatewayPolicy is true the route-level extensions come
	// from the same gateway/listener policy that also installed the listener Lua, so both coexist.
	disableListenerLevelFilter := !ptr.Deref(irRoute.EnvoyExtensions.FromGatewayPolicy, false)

	// Route has its own Lua entries — disable the inherited listener-level Lua and
	// install the route's scripts instead.
	if irListener != nil && irListener.EnvoyExtensions != nil && len(irListener.EnvoyExtensions.Luas) > 0 && disableListenerLevelFilter {
		if err := enableFilterOnRoute(route, luaListenerFCFilterName(), &routev3.FilterConfig{Disabled: true}); err != nil {
			return err
		}
	}

	if route.TypedPerFilterConfig == nil {
		route.TypedPerFilterConfig = make(map[string]*anypb.Any)
	}
	filterChainConfigPerRoute := &filterchainv3.FilterChainConfigPerRoute{
		FilterChain: &filterchainv3.FilterChain{},
	}
	for idx, ep := range irRoute.EnvoyExtensions.Luas {
		filterName := luaFilterName(idx)
		luaOnFCFilter := &luafilterv3.Lua{
			DefaultSourceCode: &corev3.DataSource{
				Specifier: &corev3.DataSource_InlineString{
					InlineString: *ep.Code,
				},
			},
		}

		// TODO: support filterContext in Lua filter make this simpler
		if ep.FilterContext != nil && ep.FilterContext.Raw != nil {
			luaPerRoute := &luafilterv3.LuaPerRoute{}
			filterCtx := &structpb.Struct{}
			if err := protojson.Unmarshal(ep.FilterContext.Raw, filterCtx); err != nil {
				return err
			}
			luaPerRoute.FilterContext = filterCtx
			luaPerRouteAny, err := anypb.New(luaPerRoute)
			if err != nil {
				return err
			}
			route.TypedPerFilterConfig[filterName] = luaPerRouteAny
		}
		luaOnFCFilterAny, err := anypb.New(luaOnFCFilter)
		if err != nil {
			return err
		}
		filterChainConfigPerRoute.FilterChain.Filters = append(filterChainConfigPerRoute.FilterChain.Filters,
			&corev3.TypedExtensionConfig{
				Name:        filterName,
				TypedConfig: luaOnFCFilterAny,
			},
		)
	}

	if len(filterChainConfigPerRoute.FilterChain.Filters) == 0 {
		return nil
	}
	fcAny, err := anypb.New(filterChainConfigPerRoute)
	if err != nil {
		return err
	}

	route.TypedPerFilterConfig[luaFCFilterName()] = fcAny
	return nil
}

// patchVirtualHost delivers listener-level Lua source at VirtualHost scope so that a listener's
// Lua policy does not bleed into virtual hosts belonging to a different listener that shares the
// same RouteConfiguration (cleartext listeners on the same port). Delivery via VirtualHost
// TypedPerFilterConfig still goes through RDS, so Lua script changes do not trigger listener drains.
func (*lua) patchVirtualHost(vh *routev3.VirtualHost, httpListener *ir.HTTPListener) error {
	if httpListener.EnvoyExtensions == nil || len(httpListener.EnvoyExtensions.Luas) == 0 {
		return nil
	}

	filterName := luaListenerFCFilterName()
	if vh.TypedPerFilterConfig != nil && vh.TypedPerFilterConfig[filterName] != nil {
		// Already delivered for this VirtualHost, e.g. because patchVirtualHost was called
		// again for a different IR listener sharing the same RouteConfiguration.
		return nil
	}

	if vh.TypedPerFilterConfig == nil {
		vh.TypedPerFilterConfig = map[string]*anypb.Any{}
	}

	filterChainConfigPerRoute := &filterchainv3.FilterChainConfigPerRoute{
		FilterChain: &filterchainv3.FilterChain{},
	}
	for i, ep := range httpListener.EnvoyExtensions.Luas {
		subFilterName := luaListenerFilterName(i)
		luaOnFCFilter := &luafilterv3.Lua{
			DefaultSourceCode: &corev3.DataSource{
				Specifier: &corev3.DataSource_InlineString{
					InlineString: *ep.Code,
				},
			},
		}

		// TODO: support filterContext in Lua filter make this simpler
		if ep.FilterContext != nil && ep.FilterContext.Raw != nil {
			luaPerRoute := &luafilterv3.LuaPerRoute{}
			filterCtx := &structpb.Struct{}
			if err := protojson.Unmarshal(ep.FilterContext.Raw, filterCtx); err != nil {
				return err
			}
			luaPerRoute.FilterContext = filterCtx
			luaPerRouteAny, err := anypb.New(luaPerRoute)
			if err != nil {
				return err
			}
			vh.TypedPerFilterConfig[subFilterName] = luaPerRouteAny
		}
		luaOnFCFilterAny, err := anypb.New(luaOnFCFilter)
		if err != nil {
			return err
		}
		filterChainConfigPerRoute.FilterChain.Filters = append(filterChainConfigPerRoute.FilterChain.Filters,
			&corev3.TypedExtensionConfig{
				Name:        subFilterName,
				TypedConfig: luaOnFCFilterAny,
			},
		)
	}

	fcAny, err := anypb.New(filterChainConfigPerRoute)
	if err != nil {
		return err
	}
	vh.TypedPerFilterConfig[filterName] = fcAny
	return nil
}
