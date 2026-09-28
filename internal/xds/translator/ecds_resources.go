// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"fmt"
	"sort"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	luafilterv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/lua/v3"
	resourceTypes "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/envoyproxy/gateway/internal/xds/types"
)

// OrderedXdsResources returns the resource sets to publish, in order, so that a route never
// references a Lua script before the ECDS resource that defines it arrives, and never
// references one after the ECDS resource that defines it has been removed. A push that
// changes no script key comes back as the single set it already is.
func OrderedXdsResources(prev, next types.XdsResources) ([]types.XdsResources, error) {
	if prev == nil {
		return []types.XdsResources{next}, nil
	}
	// Most gateways have no ECDS resources at all; spare them the diff.
	if len(prev[resourcev3.ExtensionConfigType]) == 0 && len(next[resourcev3.ExtensionConfigType]) == 0 {
		return []types.XdsResources{next}, nil
	}

	merged, added, removed, err := mergeExtensionConfigs(
		prev[resourcev3.ExtensionConfigType], next[resourcev3.ExtensionConfigType])
	if err != nil {
		return nil, err
	}

	var phases []types.XdsResources
	// If next added a script, publish it with the previous snapshot first, so that the
	// script arrives before any route that references it.
	if added {
		phases = append(phases, withExtensionConfigs(prev, merged))
	}
	// If next removed a script, keep it through the snapshot that drops the routes
	// referencing it, so that the script is removed only after no route references it.
	if removed {
		phases = append(phases, withExtensionConfigs(next, merged))
	}
	return append(phases, next), nil
}

// withExtensionConfigs returns a shallow copy of resources with the ECDS entry replaced.
func withExtensionConfigs(resources types.XdsResources, ecds []resourceTypes.Resource) types.XdsResources {
	out := make(types.XdsResources, len(resources)+1)
	for typ, res := range resources {
		out[typ] = res
	}
	out[resourcev3.ExtensionConfigType] = ecds
	return out
}

// mergeExtensionConfigs merges prev and next by resource name, keeping every Lua script
// either side carries, and reports whether next added or dropped a key.
func mergeExtensionConfigs(prev, next []resourceTypes.Resource) ([]resourceTypes.Resource, bool, bool, error) {
	byName := func(res []resourceTypes.Resource) map[string]*corev3.TypedExtensionConfig {
		out := make(map[string]*corev3.TypedExtensionConfig, len(res))
		for _, r := range res {
			if ec, ok := r.(*corev3.TypedExtensionConfig); ok {
				out[ec.Name] = ec
			}
		}
		return out
	}
	prevByName, nextByName := byName(prev), byName(next)

	names := make([]string, 0, len(prevByName)+len(nextByName))
	for name := range prevByName {
		names = append(names, name)
	}
	for name := range nextByName {
		if _, ok := prevByName[name]; !ok {
			names = append(names, name)
		}
	}
	sort.Strings(names)

	var (
		merged         []resourceTypes.Resource
		added, removed bool
	)
	for _, name := range names {
		p, n := prevByName[name], nextByName[name]
		switch {
		case p == nil:
			added = true
			merged = append(merged, n)
		case n == nil:
			removed = true
			merged = append(merged, p)
		default:
			m, a, r, err := mergeScripts(p, n)
			if err != nil {
				return nil, false, false, err
			}
			added, removed = added || a, removed || r
			merged = append(merged, m)
		}
	}
	return merged, added, removed, nil
}

// mergeScripts returns next with any Lua script that only prev had added back, and
// reports whether next added or removed a script. Only Lua is handled here, because it is
// the only filter whose routes refer to something inside the ECDS config. If another filter
// is served over ECDS and referenced by routes the same way, extend this function for its
// type. Any other config is returned as is, since routes never look inside it.
func mergeScripts(prev, next *corev3.TypedExtensionConfig) (*corev3.TypedExtensionConfig, bool, bool, error) {
	prevLua, nextLua := &luafilterv3.Lua{}, &luafilterv3.Lua{}
	if !prev.TypedConfig.MessageIs(prevLua) || !next.TypedConfig.MessageIs(nextLua) {
		return next, false, false, nil
	}
	if err := prev.TypedConfig.UnmarshalTo(prevLua); err != nil {
		return nil, false, false, fmt.Errorf("unmarshal previous %s: %w", prev.Name, err)
	}
	if err := next.TypedConfig.UnmarshalTo(nextLua); err != nil {
		return nil, false, false, fmt.Errorf("unmarshal next %s: %w", next.Name, err)
	}

	var added, removed bool
	for key := range nextLua.SourceCodes {
		if _, ok := prevLua.SourceCodes[key]; !ok {
			added = true
		}
	}
	for key, code := range prevLua.SourceCodes {
		if _, ok := nextLua.SourceCodes[key]; ok {
			continue
		}
		removed = true
		if nextLua.SourceCodes == nil {
			nextLua.SourceCodes = map[string]*corev3.DataSource{}
		}
		nextLua.SourceCodes[key] = code
	}
	if !removed {
		return next, added, false, nil
	}

	luaAny, err := anypb.New(nextLua)
	if err != nil {
		return nil, false, false, err
	}
	return &corev3.TypedExtensionConfig{Name: next.Name, TypedConfig: luaAny}, added, removed, nil
}
