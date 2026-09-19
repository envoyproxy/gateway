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
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

func init() {
	registerHTTPFilter(&lua{})
}

type lua struct{}

var _ httpFilter = &lua{}

// luaSlotBucket rounds the number of Lua slots up to a multiple of itself. The slot count
// is the one Lua change that still rewrites, and so drains, the listener, so a listener
// stays at 10 slots whether it uses 1 or 10. An unused slot carries an empty Lua config
// and builds no VM; listeners with no Lua at all get no slots.
const luaSlotBucket = 10

// luaSlotCount returns how many Lua filters to put in the HCM for a listener whose
// deepest route Lua chain is maxPerRoute.
func luaSlotCount(maxPerRoute int) int {
	if maxPerRoute == 0 {
		return 0
	}
	return ((maxPerRoute + luaSlotBucket - 1) / luaSlotBucket) * luaSlotBucket
}

// patchHCM builds and appends the lua Filters to the HTTP Connection Manager, one per
// slot, a slot being the position a script occupies in a route's Lua chain. Scripts that
// can run in a slot are stored once in that filter's sourceCodes map and routes select
// one by name, so neither the filter chain nor the VM count grows with the route count.
// Lua filters are created in disabled mode.
func (*lua) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	if mgr == nil {
		return errors.New("hcm is nil")
	}
	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	maxPerRoute := 0
	sourceCodes := map[int]map[string]*corev3.DataSource{}
	for _, route := range irListener.Routes {
		if !routeContainsLua(route) {
			continue
		}
		if count := len(route.EnvoyExtensions.Luas); count > maxPerRoute {
			maxPerRoute = count
		}
		for slot, ep := range route.EnvoyExtensions.Luas {
			if ep.Code == nil {
				continue
			}
			if _, ok := sourceCodes[slot]; !ok {
				sourceCodes[slot] = map[string]*corev3.DataSource{}
			}
			sourceCodes[slot][ep.Name] = &corev3.DataSource{
				Specifier: &corev3.DataSource_InlineString{
					InlineString: *ep.Code,
				},
			}
		}
	}

	var errs error
	for slot := range luaSlotCount(maxPerRoute) {
		if hcmContainsFilter(mgr, luaFilterName(irListener, slot)) {
			continue
		}
		filter, err := buildHCMLuaFilter(irListener, slot, sourceCodes[slot])
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}
		mgr.HttpFilters = append(mgr.HttpFilters, filter)
	}

	return errs
}

// buildHCMLuaFilter returns a Lua filter for HCM holding every script that can run in
// the given slot, keyed by the name the routes reference it with.
func buildHCMLuaFilter(irListener *ir.HTTPListener, slot int, sourceCodes map[string]*corev3.DataSource) (*hcmv3.HttpFilter, error) {
	var (
		luaProto *luafilterv3.Lua
		luaAny   *anypb.Any
		err      error
	)
	luaProto = &luafilterv3.Lua{
		SourceCodes: sourceCodes,
	}
	if err = luaProto.ValidateAll(); err != nil {
		return nil, err
	}
	if luaAny, err = anypb.New(luaProto); err != nil {
		return nil, err
	}

	return &hcmv3.HttpFilter{
		Name:     luaFilterName(irListener, slot),
		Disabled: true,
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: luaAny,
		},
	}, nil
}

// luaFilterName returns the name of the HCM Lua filter serving the given slot. It carries
// the listener because a filter name doubles as its ECDS resource name, which has to be
// unique per IR listener, and the trailing index orders the filters within the HCM, see
// newOrderedHTTPFilter.
func luaFilterName(irListener *ir.HTTPListener, slot int) string {
	return perRouteFilterName(egv1a1.EnvoyFilterLua, fmt.Sprintf("%s/%d", irListener.Name, slot))
}

// routeContainsLua returns true if Luas exists for the provided route.
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

// patchRoute patches the provided route so Lua filters are enabled if applicable.
func (*lua) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute, irListener *ir.HTTPListener) error {
	if route == nil {
		return errors.New("xds route is nil")
	}
	if irRoute == nil {
		return errors.New("ir route is nil")
	}
	if irListener == nil {
		return errors.New("ir listener is nil")
	}
	if irRoute.EnvoyExtensions == nil {
		return nil
	}

	for slot, ep := range irRoute.EnvoyExtensions.Luas {
		routeCfg, err := buildLuaRouteFilterConfig(ep)
		if err != nil {
			return err
		}
		if err := enableFilterOnRoute(route, luaFilterName(irListener, slot), routeCfg); err != nil {
			return err
		}
	}
	return nil
}

// buildLuaRouteFilterConfig selects the script this route runs by name, which both
// enables the disabled HCM filter and points it at a VM the filter already owns.
func buildLuaRouteFilterConfig(lua ir.Lua) (proto.Message, error) {
	perRoute := &luafilterv3.LuaPerRoute{
		Override: &luafilterv3.LuaPerRoute_Name{Name: lua.Name},
	}

	if lua.FilterContext == nil || lua.FilterContext.Raw == nil {
		return perRoute, nil
	}

	filterCtx := &structpb.Struct{}
	if err := protojson.Unmarshal(lua.FilterContext.Raw, filterCtx); err != nil {
		return nil, err
	}
	perRoute.FilterContext = filterCtx

	return perRoute, nil
}
