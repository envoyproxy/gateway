// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"

	luafilterv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/lua/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/envoyproxy/gateway/internal/ir"
)

// The slot count is the only Lua change that still rewrites the listener, so it is rounded
// up to keep the filter list still while policies are added and removed.
func TestLuaSlotCount(t *testing.T) {
	tests := []struct {
		name        string
		maxPerRoute int
		expected    int
	}{
		{name: "no lua needs no slots", maxPerRoute: 0, expected: 0},
		{name: "a single script still gets a full bucket", maxPerRoute: 1, expected: 10},
		{name: "a full bucket does not grow", maxPerRoute: 10, expected: 10},
		{name: "one past a bucket takes the next one", maxPerRoute: 11, expected: 20},
		{name: "two buckets and change", maxPerRoute: 21, expected: 30},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, luaSlotCount(tc.maxPerRoute))
		})
	}
}

// Routes that share a script share the Lua VM behind it: the script is stored once in the
// slot's sourceCodes map and both routes reference it by name.
func TestPatchHCMLuaSharesScriptsAcrossRoutes(t *testing.T) {
	shared := "function envoy_on_request(request_handle) end"
	second := "function envoy_on_response(response_handle) end"

	irListener := &ir.HTTPListener{
		Name: "envoy-gateway/gateway-1/http",
		Routes: []*ir.HTTPRoute{
			{
				Name: "route-1",
				EnvoyExtensions: &ir.EnvoyExtensionFeatures{
					Luas: []ir.Lua{{Name: "policy/lua/0", Code: &shared}},
				},
			},
			{
				Name: "route-2",
				EnvoyExtensions: &ir.EnvoyExtensionFeatures{
					Luas: []ir.Lua{
						{Name: "policy/lua/0", Code: &shared},
						{Name: "policy/lua/1", Code: &second},
					},
				},
			},
		},
	}

	mgr := &hcmv3.HttpConnectionManager{}
	require.NoError(t, (&lua{}).patchHCM(mgr, irListener))

	require.Len(t, mgr.HttpFilters, luaSlotCount(2))
	assert.Equal(t, "envoy.filters.http.lua/envoy-gateway/gateway-1/http/0", mgr.HttpFilters[0].Name)
	assert.True(t, mgr.HttpFilters[0].Disabled)

	slot0 := &luafilterv3.Lua{}
	require.NoError(t, mgr.HttpFilters[0].GetTypedConfig().UnmarshalTo(slot0))
	// The shared script appears once even though two routes use it.
	require.Len(t, slot0.SourceCodes, 1)
	assert.Equal(t, shared, slot0.SourceCodes["policy/lua/0"].GetInlineString())

	slot1 := &luafilterv3.Lua{}
	require.NoError(t, mgr.HttpFilters[1].GetTypedConfig().UnmarshalTo(slot1))
	assert.Equal(t, second, slot1.SourceCodes["policy/lua/1"].GetInlineString())

	// The slots beyond the deepest chain are there to keep the filter list stable, and
	// carry no script so that Envoy builds no VM for them.
	spare := &luafilterv3.Lua{}
	require.NoError(t, mgr.HttpFilters[2].GetTypedConfig().UnmarshalTo(spare))
	assert.Empty(t, spare.SourceCodes)
	assert.Nil(t, spare.DefaultSourceCode)
}
