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
	"google.golang.org/protobuf/types/known/anypb"

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

	maxLuaCount := 0
	for _, route := range irListener.Routes {
		if !routeContainsLua(route) {
			continue
		}
		if count := len(route.EnvoyExtensions.Luas); count > maxLuaCount {
			maxLuaCount = count
		}
	}

	var errs error
	for idx := range maxLuaCount {
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

// buildHCMLuaFilter returns a disabled Lua filter for HCM with empty default source code.
// The actual Lua script for each route is provided via LuaPerRoute in the route's TypedPerFilterConfig.
func buildHCMLuaFilter(filterName string) (*hcmv3.HttpFilter, error) {
	var (
		luaProto *luafilterv3.Lua
		luaAny   *anypb.Any
		err      error
	)
	luaProto = &luafilterv3.Lua{
		DefaultSourceCode: &corev3.DataSource{
			Specifier: &corev3.DataSource_InlineString{
				InlineString: "",
			},
		},
	}
	if err = luaProto.ValidateAll(); err != nil {
		return nil, err
	}
	if luaAny, err = anypb.New(luaProto); err != nil {
		return nil, err
	}

	return &hcmv3.HttpFilter{
		Name:     filterName,
		Disabled: true,
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: luaAny,
		},
	}, nil
}

// luaFilterName returns the stable top-level filter name for the Lua list index.
// The index is the execution slot within the ordered EnvoyExtensionPolicy Lua
// list, so route 0th scripts always bind to the same listener-level filter.
func luaFilterName(idx int) string {
	return perRouteFilterName(egv1a1.EnvoyFilterLua, strconv.Itoa(idx))
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
func (*lua) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute, _ *ir.HTTPListener) error {
	if route == nil {
		return errors.New("xds route is nil")
	}
	if irRoute == nil {
		return errors.New("ir route is nil")
	}
	if irRoute.EnvoyExtensions == nil {
		return nil
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
