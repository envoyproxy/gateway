// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	luafilterv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/lua/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/envoyproxy/gateway/internal/xds/types"
)

// The pass runs after the JSON patches and the extension hook, so it has to tell the
// filters Envoy Gateway generated from the ones somebody else added. Matching on the name
// would not do it: a patched-in filter can look exactly like one of ours.
func TestRecordECDSFilterNames(t *testing.T) {
	generated := "envoy.filters.http.lua/envoy-gateway/gateway-1/http/0"
	patchedIn := "envoy.filters.http.lua/my-filter/0"

	tr := &Translator{ecdsFilterNames: sets.New[string]()}
	tr.recordECDSFilterNames(&hcmv3.HttpConnectionManager{
		HttpFilters: []*hcmv3.HttpFilter{
			{Name: generated},
			{Name: "envoy.filters.http.wasm/envoy-gateway/gateway-1/http/0"},
			{Name: "envoy.filters.http.router"},
		},
	})

	assert.True(t, tr.ecdsFilterNames.Has(generated))
	// A filter type that is not eligible stays inline.
	assert.False(t, tr.ecdsFilterNames.Has("envoy.filters.http.wasm/envoy-gateway/gateway-1/http/0"))
	// And one added after Envoy Gateway built the manager was never recorded, even though
	// its name has the same shape as a generated one.
	assert.False(t, tr.ecdsFilterNames.Has(patchedIn))
}

// A filter an EnvoyPatchPolicy or an extension server added keeps its configuration inline,
// where its author put it, even when its name looks like one Envoy Gateway generates.
func TestExtractFilterChainToECDSLeavesForeignFilters(t *testing.T) {
	generated := "envoy.filters.http.lua/envoy-gateway/gateway-1/http/0"
	patchedIn := "envoy.filters.http.lua/my-filter/0"

	luaAny, err := anypb.New(&luafilterv3.Lua{})
	require.NoError(t, err)

	mgr := &hcmv3.HttpConnectionManager{
		StatPrefix: "http-10080",
		RouteSpecifier: &hcmv3.HttpConnectionManager_Rds{
			Rds: &hcmv3.Rds{RouteConfigName: "envoy-gateway/gateway-1/http"},
		},
		HttpFilters: []*hcmv3.HttpFilter{
			{Name: generated, ConfigType: &hcmv3.HttpFilter_TypedConfig{TypedConfig: luaAny}},
			{Name: patchedIn, ConfigType: &hcmv3.HttpFilter_TypedConfig{TypedConfig: luaAny}},
		},
	}
	mgrAny, err := anypb.New(mgr)
	require.NoError(t, err)

	filterChain := &listenerv3.FilterChain{
		Filters: []*listenerv3.Filter{{
			Name:       wellknown.HTTPConnectionManager,
			ConfigType: &listenerv3.Filter_TypedConfig{TypedConfig: mgrAny},
		}},
	}

	tr := &Translator{ecdsFilterNames: sets.New(generated)}
	tCtx := &types.ResourceVersionTable{}
	require.NoError(t, tr.extractFilterChainToECDS(tCtx, filterChain))

	extensionConfigs := tCtx.XdsResources[resourcev3.ExtensionConfigType]
	require.Len(t, extensionConfigs, 1)
	assert.Equal(t, generated, extensionConfigs[0].(*corev3.TypedExtensionConfig).Name)

	patched := &hcmv3.HttpConnectionManager{}
	require.NoError(t, filterChain.Filters[0].GetTypedConfig().UnmarshalTo(patched))
	assert.NotNil(t, patched.HttpFilters[0].GetConfigDiscovery(), "generated filter moves to ECDS")
	assert.NotNil(t, patched.HttpFilters[1].GetTypedConfig(), "foreign filter stays inline")
}

// A filter chain with no HTTP connection manager is skipped, but one whose manager cannot
// be read is a translation error rather than a chain quietly served without its filters.
func TestExtractFilterChainToECDSUnreadableHCM(t *testing.T) {
	tr := &Translator{ecdsFilterNames: sets.New[string]()}
	tCtx := &types.ResourceVersionTable{}

	noHCM := &listenerv3.FilterChain{
		Filters: []*listenerv3.Filter{{Name: "envoy.filters.network.tcp_proxy"}},
	}
	require.NoError(t, tr.extractFilterChainToECDS(tCtx, noHCM))

	unreadableHCM := &listenerv3.FilterChain{
		Filters: []*listenerv3.Filter{{
			Name: wellknown.HTTPConnectionManager,
			ConfigType: &listenerv3.Filter_TypedConfig{
				TypedConfig: &anypb.Any{
					TypeUrl: "type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager",
					Value:   []byte("not a protobuf"),
				},
			},
		}},
	}
	assert.Error(t, tr.extractFilterChainToECDS(tCtx, unreadableHCM))
}
