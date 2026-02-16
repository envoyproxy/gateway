// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"

	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	"github.com/stretchr/testify/assert"

	"github.com/envoyproxy/gateway/internal/ir"
)

func TestPatchProxyProtocolFilter(t *testing.T) {
	type testCase struct {
		name                  string
		listener              *listenerv3.Listener
		proxyProtocolSettings *ir.ProxyProtocolSettings
		expectedFilterName    string
		expectedFilterCount   int
	}

	testCases := []testCase{
		{
			name:                  "proxyProtocolSettings=nil (disabled)",
			listener:              &listenerv3.Listener{},
			proxyProtocolSettings: nil,
			expectedFilterName:    "",
			expectedFilterCount:   0,
		},
		{
			name:     "proxyProtocolSettings configured with Optional=false",
			listener: &listenerv3.Listener{},
			proxyProtocolSettings: &ir.ProxyProtocolSettings{
				Optional: false,
			},
			expectedFilterName:  "envoy.filters.listener.proxy_protocol",
			expectedFilterCount: 1,
		},
		{
			name:     "proxyProtocolSettings configured with Optional=true",
			listener: &listenerv3.Listener{},
			proxyProtocolSettings: &ir.ProxyProtocolSettings{
				Optional: true,
			},
			expectedFilterName:  "envoy.filters.listener.proxy_protocol",
			expectedFilterCount: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			patchProxyProtocolFilter(tc.listener, tc.proxyProtocolSettings)

			assert.Len(t, tc.listener.ListenerFilters, tc.expectedFilterCount)

			if tc.expectedFilterCount > 0 {
				assert.Equal(t, tc.expectedFilterName, tc.listener.ListenerFilters[0].Name)
			}
		})
	}
}
