package xds_hooks

import (
	"fmt"

	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	"github.com/rs/zerolog/log"

	luav3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/lua/v3"

	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/exampleorg/envoygateway-extension/internal/ir"

	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"google.golang.org/protobuf/types/known/anypb"
)

// This file contains a few helper funcs for working with Envoy's xDS protos

// This func takes all of our resources from the cluster (GlobalLuaScripts) and creates
// http filters out of them
func buildHTTPLuaFilters(resources *ir.IR) ([]*hcm.HttpFilter, error) {
	ret := []*hcm.HttpFilter{}

	if resources == nil {
		log.Info().Msg("not injecting any lua http filters since there are no custom resources")
		return ret, nil
	}

	// Build an Envoy http lua filter for each of our GlobalLuaScript resources in the kubernetes cluster
	for name, luaScript := range resources.LoadLuaScripts() {
		if luaScript == nil {
			continue
		}

		// Append our org name and an identifier to the start of the filter name (namespace.name)
		// so it is easier to identify and to reduce chance of conflicting with another filter
		filterName := fmt.Sprintf("exampleorg.io.GlobalLuaFilter-%s", name)
		luaFilter, err := newLuaScriptFilter(filterName, luaScript.Lua)
		if err != nil {
			return ret, fmt.Errorf("unable to build http filter for lua script, err: %w", err)
		}

		log.Info().Msgf("injecting http lua filter for %s", filterName)
		ret = append(ret, luaFilter)
	}

	return ret, nil
}

// Builds the xDS proto for a new lua http filter
func newLuaScriptFilter(name string, luaScript string) (*hcm.HttpFilter, error) {
	luaFilter := &luav3.Lua{
		InlineCode: luaScript,
	}

	// Since Envoy protos make heavy use of anyconfig, we need to
	// first build our specific lua filter, then use anypb to make it generic so that we can add it to the list of the
	// HTTP connection manager's http filters (this is how we can mix a bunch of different filters into a single array inside
	// a strongly typed proto)
	anyLuaFilter, err := anypb.New(luaFilter)
	if err != nil {
		return nil, err
	}

	httpLuaFilter := &hcm.HttpFilter{
		Name: name,
		ConfigType: &hcm.HttpFilter_TypedConfig{
			TypedConfig: anyLuaFilter,
		},
	}

	return httpLuaFilter, nil
}

// Tries to find an HTTP connection manager
func findHCM(filterChain *listenerv3.FilterChain) (*hcm.HttpConnectionManager, int, error) {
	for filterIndex, filter := range filterChain.Filters {
		if filter.Name == wellknown.HTTPConnectionManager {
			hcm := new(hcm.HttpConnectionManager)
			if err := filter.GetTypedConfig().UnmarshalTo(hcm); err != nil {
				return nil, -1, err
			}
			return hcm, filterIndex, nil
		}
	}
	return nil, -1, fmt.Errorf("unable to find HTTPConnectionManager in FilterChain: %s", filterChain.Name)
}
