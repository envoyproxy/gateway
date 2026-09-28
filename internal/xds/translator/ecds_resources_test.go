// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	luafilterv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/lua/v3"
	resourceTypes "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/envoyproxy/gateway/internal/xds/types"
)

func luaExtensionConfig(t *testing.T, name string, keys ...string) resourceTypes.Resource {
	t.Helper()
	lua := &luafilterv3.Lua{SourceCodes: map[string]*corev3.DataSource{}}
	for _, key := range keys {
		lua.SourceCodes[key] = &corev3.DataSource{
			Specifier: &corev3.DataSource_InlineString{InlineString: "function envoy_on_request(h) end"},
		}
	}
	luaAny, err := anypb.New(lua)
	require.NoError(t, err)
	return &corev3.TypedExtensionConfig{Name: name, TypedConfig: luaAny}
}

func extensionConfigKeys(t *testing.T, phase types.XdsResources, name string) []string {
	t.Helper()
	for _, res := range phase[resourcev3.ExtensionConfigType] {
		ec := res.(*corev3.TypedExtensionConfig)
		if ec.Name != name {
			continue
		}
		lua := &luafilterv3.Lua{}
		require.NoError(t, ec.TypedConfig.UnmarshalTo(lua))
		keys := make([]string, 0, len(lua.SourceCodes))
		for key := range lua.SourceCodes {
			keys = append(keys, key)
		}
		return keys
	}
	return nil
}

// A key a route names must be present in ECDS before the route is delivered and stay
// present until the route is gone, whichever order the two types are sent in.
func TestOrderedXdsResources(t *testing.T) {
	const slot = "envoy.filters.http.lua/gateway-1/http/0"
	oldRoutes := []resourceTypes.Resource{&routev3.RouteConfiguration{Name: "old"}}
	newRoutes := []resourceTypes.Resource{&routev3.RouteConfiguration{Name: "new"}}

	snapshot := func(routes []resourceTypes.Resource, ecds ...resourceTypes.Resource) types.XdsResources {
		return types.XdsResources{
			resourcev3.RouteType:           routes,
			resourcev3.ExtensionConfigType: ecds,
		}
	}

	t.Run("first publish goes out in one step", func(t *testing.T) {
		next := snapshot(newRoutes, luaExtensionConfig(t, slot, "a"))
		phases, err := OrderedXdsResources(nil, next)
		require.NoError(t, err)
		require.Len(t, phases, 1)
		assert.Equal(t, next, phases[0])
	})

	t.Run("no ECDS on either side goes out in one step", func(t *testing.T) {
		prev := snapshot(oldRoutes)
		next := snapshot(newRoutes)
		phases, err := OrderedXdsResources(prev, next)
		require.NoError(t, err)
		require.Len(t, phases, 1)
		assert.Equal(t, next, phases[0])
	})

	t.Run("an edit keeps the keys and goes out in one step", func(t *testing.T) {
		prev := snapshot(oldRoutes, luaExtensionConfig(t, slot, "a"))
		next := snapshot(newRoutes, luaExtensionConfig(t, slot, "a"))
		phases, err := OrderedXdsResources(prev, next)
		require.NoError(t, err)
		require.Len(t, phases, 1)
		assert.Equal(t, next, phases[0])
	})

	t.Run("an added key lands before the routes that name it", func(t *testing.T) {
		prev := snapshot(oldRoutes, luaExtensionConfig(t, slot, "a"))
		next := snapshot(newRoutes, luaExtensionConfig(t, slot, "a", "b"))
		phases, err := OrderedXdsResources(prev, next)
		require.NoError(t, err)
		require.Len(t, phases, 2)

		// Old routes, new scripts.
		assert.Equal(t, oldRoutes, phases[0][resourcev3.RouteType])
		assert.ElementsMatch(t, []string{"a", "b"}, extensionConfigKeys(t, phases[0], slot))
		// Then the new routes, with nothing left to add.
		assert.Equal(t, next, phases[1])
	})

	t.Run("a removed key leaves after the routes that named it", func(t *testing.T) {
		prev := snapshot(oldRoutes, luaExtensionConfig(t, slot, "a", "b"))
		next := snapshot(newRoutes, luaExtensionConfig(t, slot, "a"))
		phases, err := OrderedXdsResources(prev, next)
		require.NoError(t, err)
		require.Len(t, phases, 2)

		// New routes, old scripts still present.
		assert.Equal(t, newRoutes, phases[0][resourcev3.RouteType])
		assert.ElementsMatch(t, []string{"a", "b"}, extensionConfigKeys(t, phases[0], slot))
		// Then the prune.
		assert.Equal(t, next, phases[1])
	})

	t.Run("a rename does both", func(t *testing.T) {
		prev := snapshot(oldRoutes, luaExtensionConfig(t, slot, "a"))
		next := snapshot(newRoutes, luaExtensionConfig(t, slot, "b"))
		phases, err := OrderedXdsResources(prev, next)
		require.NoError(t, err)
		require.Len(t, phases, 3)

		assert.Equal(t, oldRoutes, phases[0][resourcev3.RouteType])
		assert.ElementsMatch(t, []string{"a", "b"}, extensionConfigKeys(t, phases[0], slot))
		assert.Equal(t, newRoutes, phases[1][resourcev3.RouteType])
		assert.ElementsMatch(t, []string{"a", "b"}, extensionConfigKeys(t, phases[1], slot))
		assert.Equal(t, next, phases[2])
	})

	t.Run("a new slot counts as an added key", func(t *testing.T) {
		const second = "envoy.filters.http.lua/gateway-1/http/1"
		prev := snapshot(oldRoutes, luaExtensionConfig(t, slot, "a"))
		next := snapshot(newRoutes, luaExtensionConfig(t, slot, "a"), luaExtensionConfig(t, second, "b"))
		phases, err := OrderedXdsResources(prev, next)
		require.NoError(t, err)
		require.Len(t, phases, 2)
		assert.ElementsMatch(t, []string{"b"}, extensionConfigKeys(t, phases[0], second))
	})

	t.Run("the transition never touches the other types", func(t *testing.T) {
		prev := snapshot(oldRoutes, luaExtensionConfig(t, slot, "a"))
		next := snapshot(newRoutes, luaExtensionConfig(t, slot, "a", "b"))
		next[resourcev3.ClusterType] = []resourceTypes.Resource{}
		phases, err := OrderedXdsResources(prev, next)
		require.NoError(t, err)
		// The intermediate snapshot is prev plus the union, so it carries prev's set of types.
		_, hasClusters := phases[0][resourcev3.ClusterType]
		assert.False(t, hasClusters)
		assert.Equal(t, next[resourcev3.ClusterType], phases[1][resourcev3.ClusterType])
	})
}
