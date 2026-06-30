// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"
	"strconv"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
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

// patchHCM builds and appends disabled Lua filters to the HTTP Connection Manager.
// One top-level filter is added per Lua list position across all routes. Each has
// empty default source code; the actual script is supplied per-route via
// LuaPerRoute.
//
// We intentionally do not collapse everything into a single top-level Lua filter.
// Envoy's LuaPerRoute API can only override one script for one filter instance,
// while EG's EnvoyExtensionPolicy API allows an ordered list of Lua filters per
// route. A single HCM-level filter would lose that ordering unless EG started
// synthesizing one combined script per route, which would change user semantics
// and make independently-authored scripts interfere with each other.
//
// Using one stable filter per Lua slot index preserves the route-level ordering
// of multiple Lua entries while still avoiding policy-specific HCM filter names.
// This keeps the listener filter set stable across policy churn as long as the
// maximum number of Lua entries across routes does not change.
func (*lua) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	if mgr == nil {
		return errors.New("hcm is nil")
	}
	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	hasListenerLua := irListener.EnvoyExtensions != nil && len(irListener.EnvoyExtensions.Luas) > 0
	maxRouteLuaCount := 0
	for _, route := range irListener.Routes {
		if !routeContainsLua(route) {
			continue
		}
		maxRouteLuaCount = max(len(route.EnvoyExtensions.Luas), maxRouteLuaCount)
	}
	if maxRouteLuaCount == 0 && !hasListenerLua {
		// There's no Lua enabled, we didn't add place holder for per route lua filters
		return nil
	}

	var errs error
	// add place holder http filter for per listener lua filter
	if hasListenerLua {
		for i := range irListener.EnvoyExtensions.Luas {
			filterName := luaListenerFilterName(i)
			if hcmContainsFilter(mgr, filterName) {
				continue
			}
			luaFilter, err := buildHCMLuaFilter(filterName)
			if err != nil {
				errs = errors.Join(errs, err)
				continue
			}

			mgr.HttpFilters = append(mgr.HttpFilters, luaFilter)
		}
	}

	// add place holder filters for route Lua
	for idx := range maxRouteLuaCount {
		filterName := luaFilterName(idx)
		if hcmContainsFilter(mgr, filterName) {
			continue
		}
		filter, err := buildHCMLuaFilter(filterName)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}
		mgr.HttpFilters = append(mgr.HttpFilters, filter)
	}
	return errs
}

func buildHCMLuaFilter(filterName string) (*hcmv3.HttpFilter, error) {
	var (
		luaProto *luafilterv3.Lua
		luaAny   *anypb.Any
		err      error
	)
	luaProto = &luafilterv3.Lua{
		DefaultSourceCode: &corev3.DataSource{
			Specifier: &corev3.DataSource_InlineString{},
		},
	}
	if err = luaProto.ValidateAll(); err != nil {
		return nil, err
	}
	if luaAny, err = anypb.New(luaProto); err != nil {
		return nil, err
	}

	return &hcmv3.HttpFilter{
		Name: filterName,
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: luaAny,
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
	if irListener != nil && irListener.EnvoyExtensions != nil && disableListenerLevelFilter {
		for i := range irListener.EnvoyExtensions.Luas {
			luaPerRoute := &luafilterv3.LuaPerRoute{
				Override: &luafilterv3.LuaPerRoute_Disabled{
					Disabled: true,
				},
			}
			if err := enableFilterOnRoute(route, luaListenerFilterName(i), luaPerRoute); err != nil {
				return err
			}
		}
	}

	for idx, ep := range irRoute.EnvoyExtensions.Luas {
		filterName := luaFilterName(idx)
		luaPerRoute := &luafilterv3.LuaPerRoute{
			Override: &luafilterv3.LuaPerRoute_SourceCode{
				SourceCode: &corev3.DataSource{
					Specifier: &corev3.DataSource_InlineString{
						InlineString: *ep.Code,
					},
				},
			},
		}
		if ep.FilterContext != nil && ep.FilterContext.Raw != nil {
			filterCtx := &structpb.Struct{}
			if err := protojson.Unmarshal(ep.FilterContext.Raw, filterCtx); err != nil {
				return err
			}
			luaPerRoute.FilterContext = filterCtx
		}
		luaPerRouteAny, err := anypb.New(luaPerRoute)
		if err != nil {
			return err
		}
		if route.TypedPerFilterConfig == nil {
			route.TypedPerFilterConfig = make(map[string]*anypb.Any)
		}
		if _, exists := route.TypedPerFilterConfig[filterName]; exists {
			return fmt.Errorf("route already has Lua per-route config for %s", filterName)
		}
		route.TypedPerFilterConfig[filterName] = luaPerRouteAny
	}
	return nil
}

// patchVirtualHost delivers listener-level Lua source at VirtualHost scope so that a listener's
// Lua policy does not bleed into virtual hosts belonging to a different listener that shares the
// same RouteConfiguration (cleartext listeners on the same port). Delivery via VirtualHost
// TypedPerFilterConfig still goes through RDS, so Lua script changes do not trigger listener drains.
func (*lua) patchVirtualHost(vh *routev3.VirtualHost, httpListener *ir.HTTPListener) error {
	if httpListener.EnvoyExtensions == nil {
		return nil
	}

	if vh.TypedPerFilterConfig == nil {
		vh.TypedPerFilterConfig = map[string]*anypb.Any{}
	}

	var errs error
	for i, ep := range httpListener.EnvoyExtensions.Luas {
		filterName := luaListenerFilterName(i)
		if vh.TypedPerFilterConfig[filterName] != nil {
			continue
		}

		luaPerRoute := &luafilterv3.LuaPerRoute{
			Override: &luafilterv3.LuaPerRoute_SourceCode{
				SourceCode: &corev3.DataSource{
					Specifier: &corev3.DataSource_InlineString{
						InlineString: *ep.Code,
					},
				},
			},
		}
		if ep.FilterContext != nil && ep.FilterContext.Raw != nil {
			filterCtx := &structpb.Struct{}
			if err := protojson.Unmarshal(ep.FilterContext.Raw, filterCtx); err != nil {
				return err
			}
			luaPerRoute.FilterContext = filterCtx
		}
		luaPerRouteAny, err := anypb.New(luaPerRoute)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}

		vh.TypedPerFilterConfig[filterName] = luaPerRouteAny
	}

	return errs
}
