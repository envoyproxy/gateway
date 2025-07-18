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
		enableProxyProtocol   bool
		proxyProtocolSettings *ir.ProxyProtocolSettings
		expectedFilterName    string
		expectedFilterCount   int
	}

	testCases := []testCase{
		{
			name:                  "enableProxyProtocol=true only",
			listener:              &listenerv3.Listener{},
			enableProxyProtocol:   true,
			proxyProtocolSettings: nil,
			expectedFilterName:    "envoy.filters.listener.proxy_protocol",
			expectedFilterCount:   1,
		},
		{
			name:                  "enableProxyProtocol=false only",
			listener:              &listenerv3.Listener{},
			enableProxyProtocol:   false,
			proxyProtocolSettings: nil,
			expectedFilterName:    "",
			expectedFilterCount:   0,
		},
		{
			name:                "proxyProtocolSettings configured (always enabled)",
			listener:            &listenerv3.Listener{},
			enableProxyProtocol: false,
			proxyProtocolSettings: &ir.ProxyProtocolSettings{
				AllowRequestsWithoutProxyProtocol: false,
			},
			expectedFilterName:  "envoy.filters.listener.proxy_protocol",
			expectedFilterCount: 1,
		},
		{
			name:                "proxyProtocolSettings with AllowRequestsWithoutProxyProtocol=false",
			listener:            &listenerv3.Listener{},
			enableProxyProtocol: false,
			proxyProtocolSettings: &ir.ProxyProtocolSettings{
				AllowRequestsWithoutProxyProtocol: false,
			},
			expectedFilterName:  "envoy.filters.listener.proxy_protocol",
			expectedFilterCount: 1,
		},
		{
			name:                "proxyProtocolSettings with AllowRequestsWithoutProxyProtocol=true",
			listener:            &listenerv3.Listener{},
			enableProxyProtocol: false,
			proxyProtocolSettings: &ir.ProxyProtocolSettings{
				AllowRequestsWithoutProxyProtocol: true,
			},
			expectedFilterName:  "envoy.filters.listener.proxy_protocol",
			expectedFilterCount: 1,
		},
		{
			name:                "precedence test: proxyProtocolSettings overrides enableProxyProtocol=false",
			listener:            &listenerv3.Listener{},
			enableProxyProtocol: false,
			proxyProtocolSettings: &ir.ProxyProtocolSettings{
				AllowRequestsWithoutProxyProtocol: false,
			},
			expectedFilterName:  "envoy.filters.listener.proxy_protocol",
			expectedFilterCount: 1,
		},
		{
			name:                "precedence test: proxyProtocolSettings overrides enableProxyProtocol=true",
			listener:            &listenerv3.Listener{},
			enableProxyProtocol: true,
			proxyProtocolSettings: &ir.ProxyProtocolSettings{
				AllowRequestsWithoutProxyProtocol: true,
			},
			expectedFilterName:  "envoy.filters.listener.proxy_protocol",
			expectedFilterCount: 1,
		},
		{
			name:                  "both disabled: enableProxyProtocol=false, proxyProtocolSettings=nil",
			listener:              &listenerv3.Listener{},
			enableProxyProtocol:   false,
			proxyProtocolSettings: nil,
			expectedFilterName:    "",
			expectedFilterCount:   0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			patchProxyProtocolFilter(tc.listener, tc.enableProxyProtocol, tc.proxyProtocolSettings)

			assert.Len(t, tc.listener.ListenerFilters, tc.expectedFilterCount)

			if tc.expectedFilterCount > 0 {
				assert.Equal(t, tc.expectedFilterName, tc.listener.ListenerFilters[0].Name)
			}
		})
	}
}
