// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"

	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/envoyproxy/gateway/internal/xds/types"
)

// The pass runs after the JSON patches and the extension hook, so it has to tell the
// filters Envoy Gateway generated from the ones somebody else added. Lifting a filter
// that Envoy Gateway did not name would move its configuration out from under its author,
// and two of them could claim the same ECDS resource name.
func TestECDSEligible(t *testing.T) {
	tests := []struct {
		name       string
		filterName string
		expected   bool
	}{
		{
			name:       "a generated lua filter",
			filterName: "envoy.filters.http.lua/envoy-gateway/gateway-1/http/0",
			expected:   true,
		},
		{
			name:       "a filter type that is not eligible",
			filterName: "envoy.filters.http.wasm/envoy-gateway/gateway-1/http/0",
			expected:   false,
		},
		{
			name:       "a bare filter added by a patch",
			filterName: "envoy.filters.http.lua",
			expected:   false,
		},
		{
			name:       "a patched filter carrying a name of its own",
			filterName: "envoy.filters.http.lua/my-own-filter",
			expected:   false,
		},
		{
			name:       "a patched filter whose last segment is not a slot",
			filterName: "envoy.filters.http.lua/envoy-gateway/gateway-1/http/mine",
			expected:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, ecdsEligible(&hcmv3.HttpFilter{Name: tc.filterName}))
		})
	}
}

// A filter chain with no HTTP connection manager is skipped, but one whose manager cannot
// be read is a translation error rather than a chain quietly served without its filters.
func TestExtractFilterChainToECDSUnreadableHCM(t *testing.T) {
	tCtx := &types.ResourceVersionTable{}

	noHCM := &listenerv3.FilterChain{
		Filters: []*listenerv3.Filter{{Name: "envoy.filters.network.tcp_proxy"}},
	}
	require.NoError(t, extractFilterChainToECDS(tCtx, noHCM))

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
	assert.Error(t, extractFilterChainToECDS(tCtx, unreadableHCM))
}
