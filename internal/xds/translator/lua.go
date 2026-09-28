// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"strconv"

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
// is the one Lua change that rewrites, and so drains, the listener, so a listener stays
// at 10 slots whether it uses 1 or 10. A slot is an empty placeholder and builds no VM;
// listeners with no Lua at all get no slots.
const luaSlotBucket = 10

// luaSlotCount returns how many Lua filters to put in the HCM for a listener whose
// deepest route Lua chain is maxPerRoute.
func luaSlotCount(maxPerRoute int) int {
	if maxPerRoute == 0 {
		return 0
	}
	return ((maxPerRoute + luaSlotBucket - 1) / luaSlotBucket) * luaSlotBucket
}

// patchHCM appends one disabled, empty Lua filter per slot to the HTTP Connection Manager,
// a slot being the position a script occupies in a route's Lua chain. The scripts
// themselves travel with the routes, so the listener never changes when a policy does.
// Several IR listeners can share one HCM, so this may run more than once per manager.
func (*lua) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	if mgr == nil {
		return errors.New("hcm is nil")
	}
	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	maxPerRoute := 0
	for _, route := range irListener.Routes {
		if !routeContainsLua(route) {
			continue
		}
		maxPerRoute = max(maxPerRoute, len(route.EnvoyExtensions.Luas))
	}

	var errs error
	for slot := range luaSlotCount(maxPerRoute) {
		if hcmContainsFilter(mgr, luaFilterName(slot)) {
			continue
		}
		filter, err := buildHCMLuaFilter(slot)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}
		mgr.HttpFilters = append(mgr.HttpFilters, filter)
	}
	return errs
}

// buildHCMLuaFilter returns the placeholder Lua filter for a slot. It carries no script
// and is disabled, so it does nothing until a route configures it.
func buildHCMLuaFilter(slot int) (*hcmv3.HttpFilter, error) {
	luaAny, err := anypb.New(&luafilterv3.Lua{})
	if err != nil {
		return nil, err
	}
	return &hcmv3.HttpFilter{
		Name:     luaFilterName(slot),
		Disabled: true,
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: luaAny,
		},
	}, nil
}

// luaFilterName returns the name of the HCM Lua filter serving the given slot. The
// trailing index orders the filters within the HCM, see newOrderedHTTPFilter.
func luaFilterName(slot int) string {
	return perRouteFilterName(egv1a1.EnvoyFilterLua, strconv.Itoa(slot))
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

// patchRoute gives each of the route's Lua scripts to the slot filter at its position.
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

	for slot, ep := range irRoute.EnvoyExtensions.Luas {
		routeCfg, err := buildLuaRouteFilterConfig(ep)
		if err != nil {
			return err
		}
		if err := enableFilterOnRoute(route, luaFilterName(slot), routeCfg); err != nil {
			return err
		}
	}
	return nil
}

// buildLuaRouteFilterConfig carries the script on the route. The shared VM id is the
// script's own name, so every route running this script shares one set of Lua VMs
// instead of each building its own, while scripts of different policies stay apart.
func buildLuaRouteFilterConfig(lua ir.Lua) (proto.Message, error) {
	if lua.Code == nil {
		return nil, errors.New("lua code is nil")
	}
	perRoute := &luafilterv3.LuaPerRoute{
		Override: &luafilterv3.LuaPerRoute_SourceCode{
			SourceCode: &corev3.DataSource{
				Specifier: &corev3.DataSource_InlineString{InlineString: *lua.Code},
			},
		},
		SharedVmId: lua.Name,
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
