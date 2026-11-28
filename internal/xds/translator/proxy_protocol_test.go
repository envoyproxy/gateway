// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"slices"
	"testing"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	proxyprotocolv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/listener/proxy_protocol/v3"
	tlsinspectorv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/listener/tls_inspector/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/envoyproxy/gateway/internal/ir"
)

func TestHTTP3ProxyProtocolFingerprints(t *testing.T) {
	for _, tc := range []struct {
		name         string
		fingerprints []ir.TLSFingerprintType
	}{
		{name: "JA3", fingerprints: []ir.TLSFingerprintType{ir.TLSFingerprintTypeJA3}},
		{name: "JA4", fingerprints: []ir.TLSFingerprintType{ir.TLSFingerprintTypeJA4}},
		{name: "JA3-and-JA4", fingerprints: []ir.TLSFingerprintType{ir.TLSFingerprintTypeJA3, ir.TLSFingerprintTypeJA4}},
	} {
		for _, hostname := range []string{"foo.example.com", "*"} {
			t.Run(tc.name+"/"+hostname, func(t *testing.T) {
				x := requireXdsIRFromInputTestData(t, "testdata/in/xds-ir/http3.yaml")
				listener := x.HTTP[1]
				listener.Hostnames = []string{hostname}
				listener.Routes[0].Hostname = hostname
				listener.ProxyProtocol = &ir.ProxyProtocolSettings{}
				listener.TLS.Fingerprints = tc.fingerprints
				x.HTTP = []*ir.HTTPListener{listener}
				ctx, err := (&Translator{ControllerNamespace: "envoy-gateway-system"}).Translate(t.Context(), x)
				require.NoError(t, err)
				tcp := findXdsListenerByHostPort(ctx, listener.Address, listener.Port, corev3.SocketAddress_TCP)
				quic := findXdsListenerByHostPort(ctx, listener.Address, listener.Port, corev3.SocketAddress_UDP)
				require.NotNil(t, tcp)
				require.NotNil(t, quic)
				// Both TCP listener filters remain configured with their requested settings.
				require.Len(t, tcp.ListenerFilters, 2)
				for _, filter := range tcp.ListenerFilters {
					switch filter.Name {
					case wellknown.ProxyProtocol:
						config := &proxyprotocolv3.ProxyProtocol{}
						require.NoError(t, filter.GetTypedConfig().UnmarshalTo(config))
						require.False(t, config.AllowRequestsWithoutProxyProtocol)
					case wellknown.TlsInspector:
						config := &tlsinspectorv3.TlsInspector{}
						require.NoError(t, filter.GetTypedConfig().UnmarshalTo(config))
						require.Equal(t, slices.Contains(tc.fingerprints, ir.TLSFingerprintTypeJA3), config.GetEnableJa3Fingerprinting().GetValue())
						require.Equal(t, slices.Contains(tc.fingerprints, ir.TLSFingerprintTypeJA4), config.GetEnableJa4Fingerprinting().GetValue())
					default:
						t.Fatalf("unexpected TCP listener filter: %s", filter.Name)
					}
				}
				// QUIC keeps its SNI match and HTTP/3 HCM without TCP listener filters.
				require.Empty(t, quic.ListenerFilters)
				require.Len(t, quic.FilterChains, 1)
				if hostname == "*" {
					require.Empty(t, quic.FilterChains[0].GetFilterChainMatch().GetServerNames())
				} else {
					require.Equal(t, listener.Hostnames, quic.FilterChains[0].GetFilterChainMatch().GetServerNames())
				}
				mgr := extractHCM(t, quic)
				require.Equal(t, hcmv3.HttpConnectionManager_HTTP3, mgr.CodecType)
				routes := findXdsRouteConfig(ctx, mgr.GetRds().GetRouteConfigName())
				require.NotNil(t, routes)
				require.NotEmpty(t, routes.VirtualHosts)
				for _, vhost := range routes.VirtualHosts {
					require.NotEmpty(t, vhost.Routes)
					for _, route := range vhost.Routes {
						require.True(t, slices.ContainsFunc(route.ResponseHeadersToAdd, func(header *corev3.HeaderValueOption) bool {
							return header.GetHeader().GetKey() == "alt-svc"
						}))
					}
				}
			})
		}
	}
}

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
			name: "UDP listener ignores required proxy protocol",
			listener: &listenerv3.Listener{
				Address: &corev3.Address{Address: &corev3.Address_SocketAddress{
					SocketAddress: &corev3.SocketAddress{Protocol: corev3.SocketAddress_UDP},
				}},
			},
			proxyProtocolSettings: &ir.ProxyProtocolSettings{Optional: false},
			expectedFilterCount:   0,
		},
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
				config := &proxyprotocolv3.ProxyProtocol{}
				require.NoError(t, tc.listener.ListenerFilters[0].GetTypedConfig().UnmarshalTo(config))
				assert.Equal(t, tc.proxyProtocolSettings.Optional, config.AllowRequestsWithoutProxyProtocol)
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
		http3Listener  bool
		wantConfigNil  bool
		wantHTTPSPorts []uint32
		wantHTTPPorts  []uint32
	}

	testCases := []testCase{
		{
			name: "HTTP/3 ignores forwardProtoConfig",
			proxyProtocol: &ir.ProxyProtocolSettings{
				ForwardProto: &ir.ForwardProtoConfig{
					HTTPSDestinationPorts: []uint32{443},
					HTTPDestinationPorts:  []uint32{80},
				},
			},
			http3Listener: true,
			wantConfigNil: true,
		},
		{
			name: "forwardProtoConfig with both port lists",
			proxyProtocol: &ir.ProxyProtocolSettings{
				ForwardProto: &ir.ForwardProtoConfig{
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
				ForwardProto: &ir.ForwardProtoConfig{
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
			require.NoError(t, tr.addHCMToXDSListener(xdsListener, httpListener(tc.proxyProtocol), nil, nil, tc.http3Listener, nil, nil))

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
