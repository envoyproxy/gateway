// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"

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
// One filter is added per distinct Lua name (across all routes). Each has empty default source code;
// the actual script is supplied per-route via LuaPerRoute.
func (*lua) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	if mgr == nil {
		return errors.New("hcm is nil")
	}
	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	var errs error
	for _, route := range irListener.Routes {
		if !routeContainsLua(route) {
			continue
		}
		for _, ep := range route.EnvoyExtensions.Luas {
			if hcmContainsFilter(mgr, luaFilterName(ep)) {
				continue
			}
			filter, err := buildHCMLuaFilter(ep)
			if err != nil {
				errs = errors.Join(errs, err)
				continue
			}
			mgr.HttpFilters = append(mgr.HttpFilters, filter)
		}
	}
	return errs
}

// buildHCMLuaFilter returns a disabled Lua filter for HCM with empty default source code.
// The actual Lua script for each route is provided via LuaPerRoute in the route's TypedPerFilterConfig.
func buildHCMLuaFilter(lua ir.Lua) (*hcmv3.HttpFilter, error) {
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
		Name:     luaFilterName(lua),
		Disabled: true,
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: luaAny,
		},
	}, nil
}

// luaFilterName returns the filter name for the given Lua (policy-based naming).
func luaFilterName(lua ir.Lua) string {
	return perRouteFilterName(egv1a1.EnvoyFilterLua, lua.Name)
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

	for _, ep := range irRoute.EnvoyExtensions.Luas {
		filterName := luaFilterName(ep)
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
