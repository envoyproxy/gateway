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
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"
	"k8s.io/utils/ptr"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

func TestLuaPatchHCM(t *testing.T) {
	luaFilter := &lua{}

	tests := []struct {
		name       string
		mgr        *hcmv3.HttpConnectionManager
		irListener *ir.HTTPListener
		wantErr    bool
		errMsg     string
		validate   func(t *testing.T, mgr *hcmv3.HttpConnectionManager)
	}{
		{
			name: "nil hcm",
			mgr:  nil,
			irListener: &ir.HTTPListener{
				Routes: []*ir.HTTPRoute{},
			},
			wantErr: true,
			errMsg:  "hcm is nil",
		},
		{
			name:       "nil ir listener",
			mgr:        &hcmv3.HttpConnectionManager{},
			irListener: nil,
			wantErr:    true,
			errMsg:     "ir listener is nil",
		},
		{
			name: "no lua filters",
			mgr:  &hcmv3.HttpConnectionManager{},
			irListener: &ir.HTTPListener{
				Routes: []*ir.HTTPRoute{
					{
						Name: "test-route",
					},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, mgr *hcmv3.HttpConnectionManager) {
				assert.Empty(t, mgr.HttpFilters)
			},
		},
		{
			name: "single lua filter",
			mgr:  &hcmv3.HttpConnectionManager{},
			irListener: &ir.HTTPListener{
				Routes: []*ir.HTTPRoute{
					{
						Name: "test-route",
						EnvoyExtensions: &ir.EnvoyExtensionFeatures{
							Luas: []ir.Lua{
								{
									Name: "test-lua",
									Code: ptr.To("function envoy_on_request(request_handle)\n  request_handle:headers():add(\"foo\", \"bar\")\nend"),
								},
							},
						},
					},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, mgr *hcmv3.HttpConnectionManager) {
				require.Len(t, mgr.HttpFilters, 1)
				assert.Equal(t, "envoy.filters.http.lua/test-lua", mgr.HttpFilters[0].Name)
				assert.True(t, mgr.HttpFilters[0].Disabled)
			},
		},
		{
			name: "multiple lua filters",
			mgr:  &hcmv3.HttpConnectionManager{},
			irListener: &ir.HTTPListener{
				Routes: []*ir.HTTPRoute{
					{
						Name: "test-route-1",
						EnvoyExtensions: &ir.EnvoyExtensionFeatures{
							Luas: []ir.Lua{
								{
									Name: "lua-1",
									Code: ptr.To("function envoy_on_request(request_handle)\n  request_handle:headers():add(\"x-lua-1\", \"true\")\nend"),
								},
								{
									Name: "lua-2",
									Code: ptr.To("function envoy_on_request(request_handle)\n  request_handle:headers():add(\"x-lua-2\", \"true\")\nend"),
								},
							},
						},
					},
					{
						Name: "test-route-2",
						EnvoyExtensions: &ir.EnvoyExtensionFeatures{
							Luas: []ir.Lua{
								{
									Name: "lua-3",
									Code: ptr.To("function envoy_on_response(response_handle)\n  response_handle:headers():add(\"x-lua-3\", \"true\")\nend"),
								},
							},
						},
					},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, mgr *hcmv3.HttpConnectionManager) {
				require.Len(t, mgr.HttpFilters, 3)
				assert.Equal(t, "envoy.filters.http.lua/lua-1", mgr.HttpFilters[0].Name)
				assert.Equal(t, "envoy.filters.http.lua/lua-2", mgr.HttpFilters[1].Name)
				assert.Equal(t, "envoy.filters.http.lua/lua-3", mgr.HttpFilters[2].Name)
			},
		},
		{
			name: "duplicate lua filter names - should not add duplicate",
			mgr:  &hcmv3.HttpConnectionManager{},
			irListener: &ir.HTTPListener{
				Routes: []*ir.HTTPRoute{
					{
						Name: "test-route-1",
						EnvoyExtensions: &ir.EnvoyExtensionFeatures{
							Luas: []ir.Lua{
								{
									Name: "shared-lua",
									Code: ptr.To("function envoy_on_request(request_handle)\nend"),
								},
							},
						},
					},
					{
						Name: "test-route-2",
						EnvoyExtensions: &ir.EnvoyExtensionFeatures{
							Luas: []ir.Lua{
								{
									Name: "shared-lua",
									Code: ptr.To("function envoy_on_request(request_handle)\nend"),
								},
							},
						},
					},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, mgr *hcmv3.HttpConnectionManager) {
				require.Len(t, mgr.HttpFilters, 1)
				assert.Equal(t, "envoy.filters.http.lua/shared-lua", mgr.HttpFilters[0].Name)
			},
		},
		{
			name: "mixed routes with and without lua",
			mgr:  &hcmv3.HttpConnectionManager{},
			irListener: &ir.HTTPListener{
				Routes: []*ir.HTTPRoute{
					{
						Name: "route-without-lua",
					},
					{
						Name: "route-with-lua",
						EnvoyExtensions: &ir.EnvoyExtensionFeatures{
							Luas: []ir.Lua{
								{
									Name: "test-lua",
									Code: ptr.To("function envoy_on_request(request_handle)\nend"),
								},
							},
						},
					},
					{
						Name: "another-route-without-lua",
					},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, mgr *hcmv3.HttpConnectionManager) {
				require.Len(t, mgr.HttpFilters, 1)
				assert.Equal(t, "envoy.filters.http.lua/test-lua", mgr.HttpFilters[0].Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := luaFilter.patchHCM(tt.mgr, tt.irListener)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, tt.mgr)
				}
			}
		})
	}
}

func TestBuildHCMLuaFilter(t *testing.T) {
	tests := []struct {
		name     string
		lua      ir.Lua
		wantErr  bool
		validate func(t *testing.T, filter *hcmv3.HttpFilter)
	}{
		{
			name: "valid lua filter",
			lua: ir.Lua{
				Name: "test-lua",
				Code: ptr.To("function envoy_on_request(request_handle)\n  request_handle:headers():add(\"foo\", \"bar\")\nend"),
			},
			wantErr: false,
			validate: func(t *testing.T, filter *hcmv3.HttpFilter) {
				assert.Equal(t, "envoy.filters.http.lua/test-lua", filter.Name)
				assert.True(t, filter.Disabled)

				// Verify the typed config
				require.NotNil(t, filter.GetTypedConfig())
				luaConfig := &luafilterv3.Lua{}
				err := filter.GetTypedConfig().UnmarshalTo(luaConfig)
				require.NoError(t, err)

				// Verify the inline string
				require.NotNil(t, luaConfig.DefaultSourceCode)
				inlineString := luaConfig.DefaultSourceCode.GetInlineString()
				assert.Equal(t, "function envoy_on_request(request_handle)\n  request_handle:headers():add(\"foo\", \"bar\")\nend", inlineString)
			},
		},
		{
			name: "lua with complex script",
			lua: ir.Lua{
				Name: "complex-lua",
				Code: ptr.To(`
function envoy_on_request(request_handle)
  local headers = request_handle:headers()
  local path = headers:get(":path")
  
  if path == "/admin" then
    request_handle:respond(
      {[":status"] = "403"},
      "Forbidden"
    )
  end
end

function envoy_on_response(response_handle)
  response_handle:headers():add("x-processed", "true")
end
`),
			},
			wantErr: false,
			validate: func(t *testing.T, filter *hcmv3.HttpFilter) {
				assert.Equal(t, "envoy.filters.http.lua/complex-lua", filter.Name)
				assert.True(t, filter.Disabled)

				luaConfig := &luafilterv3.Lua{}
				err := filter.GetTypedConfig().UnmarshalTo(luaConfig)
				require.NoError(t, err)

				inlineString := luaConfig.DefaultSourceCode.GetInlineString()
				assert.Contains(t, inlineString, "envoy_on_request")
				assert.Contains(t, inlineString, "envoy_on_response")
			},
		},
		{
			name: "empty lua code",
			lua: ir.Lua{
				Name: "empty-lua",
				Code: ptr.To(""),
			},
			wantErr: false,
			validate: func(t *testing.T, filter *hcmv3.HttpFilter) {
				luaConfig := &luafilterv3.Lua{}
				err := filter.GetTypedConfig().UnmarshalTo(luaConfig)
				require.NoError(t, err)
				assert.Equal(t, "", luaConfig.DefaultSourceCode.GetInlineString())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, err := buildHCMLuaFilter(tt.lua)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, filter)
				if tt.validate != nil {
					tt.validate(t, filter)
				}
			}
		})
	}
}

func TestLuaFilterName(t *testing.T) {
	tests := []struct {
		name string
		lua  ir.Lua
		want string
	}{
		{
			name: "simple name",
			lua: ir.Lua{
				Name: "test-lua",
			},
			want: "envoy.filters.http.lua/test-lua",
		},
		{
			name: "name with special characters",
			lua: ir.Lua{
				Name: "my-lua-filter-v1",
			},
			want: "envoy.filters.http.lua/my-lua-filter-v1",
		},
		{
			name: "empty name",
			lua: ir.Lua{
				Name: "",
			},
			want: "envoy.filters.http.lua/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := luaFilterName(tt.lua)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRouteContainsLua(t *testing.T) {
	tests := []struct {
		name    string
		irRoute *ir.HTTPRoute
		want    bool
	}{
		{
			name:    "nil route",
			irRoute: nil,
			want:    false,
		},
		{
			name: "route without envoy extensions",
			irRoute: &ir.HTTPRoute{
				Name: "test-route",
			},
			want: false,
		},
		{
			name: "route with empty envoy extensions",
			irRoute: &ir.HTTPRoute{
				Name:            "test-route",
				EnvoyExtensions: &ir.EnvoyExtensionFeatures{},
			},
			want: false,
		},
		{
			name: "route with empty lua list",
			irRoute: &ir.HTTPRoute{
				Name: "test-route",
				EnvoyExtensions: &ir.EnvoyExtensionFeatures{
					Luas: []ir.Lua{},
				},
			},
			want: false,
		},
		{
			name: "route with single lua",
			irRoute: &ir.HTTPRoute{
				Name: "test-route",
				EnvoyExtensions: &ir.EnvoyExtensionFeatures{
					Luas: []ir.Lua{
						{
							Name: "test-lua",
							Code: ptr.To("function envoy_on_request(request_handle)\nend"),
						},
					},
				},
			},
			want: true,
		},
		{
			name: "route with multiple luas",
			irRoute: &ir.HTTPRoute{
				Name: "test-route",
				EnvoyExtensions: &ir.EnvoyExtensionFeatures{
					Luas: []ir.Lua{
						{
							Name: "lua-1",
							Code: ptr.To("function envoy_on_request(request_handle)\nend"),
						},
						{
							Name: "lua-2",
							Code: ptr.To("function envoy_on_response(response_handle)\nend"),
						},
					},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := routeContainsLua(tt.irRoute)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLuaPatchResources(t *testing.T) {
	luaFilter := &lua{}

	// patchResources should always return nil and not modify resources
	rvt := &types.ResourceVersionTable{}
	routes := []*ir.HTTPRoute{
		{
			Name: "test-route",
			EnvoyExtensions: &ir.EnvoyExtensionFeatures{
				Luas: []ir.Lua{
					{
						Name: "test-lua",
						Code: ptr.To("function envoy_on_request(request_handle)\nend"),
					},
				},
			},
		},
	}

	err := luaFilter.patchResources(rvt, routes)
	assert.NoError(t, err)
}

func TestLuaPatchRoute(t *testing.T) {
	luaFilter := &lua{}

	tests := []struct {
		name       string
		route      *routev3.Route
		irRoute    *ir.HTTPRoute
		irListener *ir.HTTPListener
		wantErr    bool
		errMsg     string
		validate   func(t *testing.T, route *routev3.Route)
	}{
		{
			name:  "nil xds route",
			route: nil,
			irRoute: &ir.HTTPRoute{
				Name: "test-route",
			},
			irListener: &ir.HTTPListener{},
			wantErr:    true,
			errMsg:     "xds route is nil",
		},
		{
			name:       "nil ir route",
			route:      &routev3.Route{},
			irRoute:    nil,
			irListener: &ir.HTTPListener{},
			wantErr:    true,
			errMsg:     "ir route is nil",
		},
		{
			name:  "route without envoy extensions",
			route: &routev3.Route{},
			irRoute: &ir.HTTPRoute{
				Name: "test-route",
			},
			irListener: &ir.HTTPListener{},
			wantErr:    false,
			validate: func(t *testing.T, route *routev3.Route) {
				// No filters should be added
				assert.Nil(t, route.TypedPerFilterConfig)
			},
		},
		{
			name:  "route with empty lua list",
			route: &routev3.Route{},
			irRoute: &ir.HTTPRoute{
				Name: "test-route",
				EnvoyExtensions: &ir.EnvoyExtensionFeatures{
					Luas: []ir.Lua{},
				},
			},
			irListener: &ir.HTTPListener{},
			wantErr:    false,
			validate: func(t *testing.T, route *routev3.Route) {
				// No filters should be added for empty list
				if route.TypedPerFilterConfig != nil {
					assert.Empty(t, route.TypedPerFilterConfig)
				}
			},
		},
		{
			name:  "route with single lua filter",
			route: &routev3.Route{},
			irRoute: &ir.HTTPRoute{
				Name: "test-route",
				EnvoyExtensions: &ir.EnvoyExtensionFeatures{
					Luas: []ir.Lua{
						{
							Name: "test-lua",
							Code: ptr.To("function envoy_on_request(request_handle)\nend"),
						},
					},
				},
			},
			irListener: &ir.HTTPListener{},
			wantErr:    false,
			validate: func(t *testing.T, route *routev3.Route) {
				require.NotNil(t, route.TypedPerFilterConfig)
				filterName := "envoy.filters.http.lua/test-lua"
				assert.Contains(t, route.TypedPerFilterConfig, filterName)
			},
		},
		{
			name:  "route with multiple lua filters",
			route: &routev3.Route{},
			irRoute: &ir.HTTPRoute{
				Name: "test-route",
				EnvoyExtensions: &ir.EnvoyExtensionFeatures{
					Luas: []ir.Lua{
						{
							Name: "lua-1",
							Code: ptr.To("function envoy_on_request(request_handle)\nend"),
						},
						{
							Name: "lua-2",
							Code: ptr.To("function envoy_on_response(response_handle)\nend"),
						},
					},
				},
			},
			irListener: &ir.HTTPListener{},
			wantErr:    false,
			validate: func(t *testing.T, route *routev3.Route) {
				require.NotNil(t, route.TypedPerFilterConfig)
				assert.Contains(t, route.TypedPerFilterConfig, "envoy.filters.http.lua/lua-1")
				assert.Contains(t, route.TypedPerFilterConfig, "envoy.filters.http.lua/lua-2")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := luaFilter.patchRoute(tt.route, tt.irRoute, tt.irListener)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, tt.route)
				}
			}
		})
	}
}

func TestLuaFilterIntegration(t *testing.T) {
	// Test the complete flow: patchHCM -> patchRoute
	luaFilter := &lua{}

	luaCode := ptr.To(`
function envoy_on_request(request_handle)
  request_handle:headers():add("x-custom-header", "test-value")
end
`)

	irListener := &ir.HTTPListener{
		Routes: []*ir.HTTPRoute{
			{
				Name: "test-route",
				EnvoyExtensions: &ir.EnvoyExtensionFeatures{
					Luas: []ir.Lua{
						{
							Name: "integration-lua",
							Code: luaCode,
						},
					},
				},
			},
		},
	}

	// Step 1: Patch HCM
	mgr := &hcmv3.HttpConnectionManager{}
	err := luaFilter.patchHCM(mgr, irListener)
	require.NoError(t, err)
	require.Len(t, mgr.HttpFilters, 1)

	// Verify the filter is disabled in HCM
	assert.True(t, mgr.HttpFilters[0].Disabled)
	assert.Equal(t, "envoy.filters.http.lua/integration-lua", mgr.HttpFilters[0].Name)

	// Verify the Lua config
	luaConfig := &luafilterv3.Lua{}
	err = mgr.HttpFilters[0].GetTypedConfig().UnmarshalTo(luaConfig)
	require.NoError(t, err)
	assert.Equal(t, *luaCode, luaConfig.DefaultSourceCode.GetInlineString())

	// Step 2: Patch Route to enable the filter
	route := &routev3.Route{}
	err = luaFilter.patchRoute(route, irListener.Routes[0], irListener)
	require.NoError(t, err)

	// Verify the filter is enabled on the route
	require.NotNil(t, route.TypedPerFilterConfig)
	filterName := "envoy.filters.http.lua/integration-lua"
	assert.Contains(t, route.TypedPerFilterConfig, filterName)

	// Verify the per-route config
	filterConfig := route.TypedPerFilterConfig[filterName]
	require.NotNil(t, filterConfig)

	// The config should be an empty Any (filter is just enabled)
	routeFilterConfig := &routev3.FilterConfig{}
	err = filterConfig.UnmarshalTo(routeFilterConfig)
	require.NoError(t, err)
	assert.NotNil(t, routeFilterConfig.Config)
}

func TestLuaFilterValidation(t *testing.T) {
	// Test that buildHCMLuaFilter creates valid Envoy configuration
	lua := ir.Lua{
		Name: "validation-test",
		Code: ptr.To("function envoy_on_request(request_handle)\n  -- valid lua code\nend"),
	}

	filter, err := buildHCMLuaFilter(lua)
	require.NoError(t, err)
	require.NotNil(t, filter)

	// Verify the filter structure
	assert.Equal(t, "envoy.filters.http.lua/validation-test", filter.Name)
	assert.True(t, filter.Disabled)

	// Verify TypedConfig is valid
	require.NotNil(t, filter.GetTypedConfig())

	luaConfig := &luafilterv3.Lua{}
	err = filter.GetTypedConfig().UnmarshalTo(luaConfig)
	require.NoError(t, err)

	// Verify the Lua configuration is valid according to Envoy's validation
	err = luaConfig.ValidateAll()
	assert.NoError(t, err)

	// Verify the source code is set correctly
	require.NotNil(t, luaConfig.DefaultSourceCode)
	assert.IsType(t, &corev3.DataSource_InlineString{}, luaConfig.DefaultSourceCode.Specifier)
	assert.Equal(t, "function envoy_on_request(request_handle)\n  -- valid lua code\nend",
		luaConfig.DefaultSourceCode.GetInlineString())
}

func TestLuaFilterWithExistingFilters(t *testing.T) {
	// Test that lua filter doesn't interfere with existing filters
	luaFilter := &lua{}

	existingFilter := &hcmv3.HttpFilter{
		Name: "envoy.filters.http.router",
	}

	mgr := &hcmv3.HttpConnectionManager{
		HttpFilters: []*hcmv3.HttpFilter{existingFilter},
	}

	irListener := &ir.HTTPListener{
		Routes: []*ir.HTTPRoute{
			{
				Name: "test-route",
				EnvoyExtensions: &ir.EnvoyExtensionFeatures{
					Luas: []ir.Lua{
						{
							Name: "new-lua",
							Code: ptr.To("function envoy_on_request(request_handle)\nend"),
						},
					},
				},
			},
		},
	}

	err := luaFilter.patchHCM(mgr, irListener)
	require.NoError(t, err)

	// Should have both the existing filter and the new lua filter
	require.Len(t, mgr.HttpFilters, 2)
	assert.Equal(t, "envoy.filters.http.router", mgr.HttpFilters[0].Name)
	assert.Equal(t, "envoy.filters.http.lua/new-lua", mgr.HttpFilters[1].Name)
}

func TestLuaFilterEmptyConfig(t *testing.T) {
	// Test enableFilterOnRoute creates proper empty config
	route := &routev3.Route{}
	filterName := "envoy.filters.http.lua/test"

	err := enableFilterOnRoute(route, filterName, &routev3.FilterConfig{
		Config: &anypb.Any{},
	})
	require.NoError(t, err)

	require.NotNil(t, route.TypedPerFilterConfig)
	assert.Contains(t, route.TypedPerFilterConfig, filterName)
}
