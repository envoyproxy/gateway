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

// patchHCM builds and appends a fixed number of disabled Lua filters to the HTTP Connection Manager.
// The number of filters is the maximum number of Lua filters on any single route (including gateway-level EEP).
// Filter names are envoy.filters.http.lua/0, envoy.filters.http.lua/1, ... so the HCM filter chain
// does not change when policies change; each route overrides the slots it needs via LuaPerRoute.
func (*lua) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	if mgr == nil {
		return errors.New("hcm is nil")
	}
	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	n := maxLuasInListener(irListener)
	if n == 0 {
		return nil
	}

	var errs error
	for i := 0; i < n; i++ {
		if hcmContainsFilter(mgr, luaFilterNameByIndex(i)) {
			continue
		}
		filter, err := buildHCMLuaFilter(i)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}
		mgr.HttpFilters = append(mgr.HttpFilters, filter)
	}
	return errs
}

// maxLuasInListener returns the maximum number of Lua filters on any single route in the listener.
func maxLuasInListener(irListener *ir.HTTPListener) int {
	var max int
	for _, route := range irListener.Routes {
		if route.EnvoyExtensions != nil && len(route.EnvoyExtensions.Luas) > max {
			max = len(route.EnvoyExtensions.Luas)
		}
	}
	return max
}

// buildHCMLuaFilter returns a disabled Lua filter for HCM with empty default source code.
// The actual Lua script for each route is provided via LuaPerRoute in the route's TypedPerFilterConfig.
func buildHCMLuaFilter(index int) (*hcmv3.HttpFilter, error) {
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
		Name:     luaFilterNameByIndex(index),
		Disabled: true,
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: luaAny,
		},
	}, nil
}

// luaFilterNameByIndex returns the filter name for the Lua slot at the given index.
// Used so the HCM filter chain is stable (envoy.filters.http.lua/0, /1, ...) regardless of policy names.
func luaFilterNameByIndex(index int) string {
	return perRouteFilterName(egv1a1.EnvoyFilterLua, strconv.Itoa(index))
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

	for i, ep := range irRoute.EnvoyExtensions.Luas {
		filterName := luaFilterNameByIndex(i)
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
