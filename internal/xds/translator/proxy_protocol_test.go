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

// extractHCM retrieves the HttpConnectionManager network filter from the built xDS listener's
// filter chains (or default filter chain), failing the test if it cannot be found or decoded.
func extractHCM(t *testing.T, xdsListener *listenerv3.Listener) *hcmv3.HttpConnectionManager {
	t.Helper()
	var filters []*listenerv3.Filter
	if len(xdsListener.FilterChains) > 0 {
		filters = xdsListener.FilterChains[0].Filters
	} else if xdsListener.DefaultFilterChain != nil {
		filters = xdsListener.DefaultFilterChain.Filters
	}
	for _, f := range filters {
		if f.Name == wellknown.HTTPConnectionManager {
			hcm := &hcmv3.HttpConnectionManager{}
			require.NoError(t, f.GetTypedConfig().UnmarshalTo(hcm))
			return hcm
		}
	}
	require.Fail(t, "HTTPConnectionManager network filter not found")
	return nil
}

func TestHCMForwardProtoConfig(t *testing.T) {
	// addHCMToXDSListener requires a non-nil Path on the IR HTTPListener.
	httpListener := func(proxyProtocol *ir.ProxyProtocolSettings) *ir.HTTPListener {
		return &ir.HTTPListener{
			CoreListenerDetails: ir.CoreListenerDetails{Name: "test-listener"},
			Path:                ir.PathSettings{},
			ProxyProtocol:       proxyProtocol,
		}
	}

	type testCase struct {
		name           string
		proxyProtocol  *ir.ProxyProtocolSettings
		wantConfigNil  bool
		wantHTTPSPorts []uint32
		wantHTTPPorts  []uint32
	}

	testCases := []testCase{
		{
			name: "forwardProtoConfig with both port lists",
			proxyProtocol: &ir.ProxyProtocolSettings{
				ForwardProtoConfig: &ir.ForwardProtoConfig{
					HTTPSDestinationPorts: []uint32{443, 8443},
					HTTPDestinationPorts:  []uint32{80, 8080},
				},
			},
			wantConfigNil:  false,
			wantHTTPSPorts: []uint32{443, 8443},
			wantHTTPPorts:  []uint32{80, 8080},
		},
		{
			name: "forwardProtoConfig with only https ports",
			proxyProtocol: &ir.ProxyProtocolSettings{
				ForwardProtoConfig: &ir.ForwardProtoConfig{
					HTTPSDestinationPorts: []uint32{443},
				},
			},
			wantConfigNil:  false,
			wantHTTPSPorts: []uint32{443},
			wantHTTPPorts:  nil,
		},
		{
			name: "proxyProtocol enabled without forwardProtoConfig",
			proxyProtocol: &ir.ProxyProtocolSettings{
				Optional: true,
			},
			wantConfigNil: true,
		},
		{
			name:          "proxyProtocol nil",
			proxyProtocol: nil,
			wantConfigNil: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tr := &Translator{}
			xdsListener := &listenerv3.Listener{Name: "test"}
			require.NoError(t, tr.addHCMToXDSListener(xdsListener, httpListener(tc.proxyProtocol), nil, nil, false, nil))

			hcm := extractHCM(t, xdsListener)
			if tc.wantConfigNil {
				assert.Nil(t, hcm.ForwardProtoConfig)
				return
			}
			require.NotNil(t, hcm.ForwardProtoConfig)
			assert.Equal(t, tc.wantHTTPSPorts, hcm.ForwardProtoConfig.HttpsDestinationPorts)
			assert.Equal(t, tc.wantHTTPPorts, hcm.ForwardProtoConfig.HttpDestinationPorts)
		})
	}
}
