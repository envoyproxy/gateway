// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"

	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	tls_inspectorv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/listener/tls_inspector/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/stretchr/testify/require"

	"github.com/envoyproxy/gateway/internal/ir"
)

func TestAddALPNMatch(t *testing.T) {
	tests := []struct {
		name               string
		alpnProtocols      []string
		serverNames        []string
		fingerprints       []ir.TLSFingerprintType
		expectMatch        []string
		expectTLSInspector bool
	}{
		{
			name:          "ALPN not configured",
			alpnProtocols: nil,
		},
		{
			name:          "ALPN disabled",
			alpnProtocols: []string{},
		},
		{
			name:          "h2 and http/1.1",
			alpnProtocols: []string{"h2", "http/1.1"},
		},
		{
			name:          "http/1.1 only",
			alpnProtocols: []string{"http/1.1"},
		},
		{
			name:          "h2 and http/1.0",
			alpnProtocols: []string{"h2", "http/1.0"},
		},
		{
			name:               "h2 only",
			alpnProtocols:      []string{"h2"},
			expectMatch:        []string{"h2"},
			expectTLSInspector: true,
		},
		{
			name:               "protocols other than HTTP/1.x",
			alpnProtocols:      []string{"h2", "some-other-protocol"},
			expectMatch:        []string{"h2", "some-other-protocol"},
			expectTLSInspector: true,
		},
		{
			name:               "h2 only keeps the server names match",
			alpnProtocols:      []string{"h2"},
			serverNames:        []string{"foo.com"},
			expectMatch:        []string{"h2"},
			expectTLSInspector: true,
		},
		{
			name:               "h2 only with fingerprints",
			alpnProtocols:      []string{"h2"},
			fingerprints:       []ir.TLSFingerprintType{ir.TLSFingerprintTypeJA3},
			expectMatch:        []string{"h2"},
			expectTLSInspector: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			xdsListener := &listenerv3.Listener{}
			filterChain := &listenerv3.FilterChain{}
			if len(tt.serverNames) > 0 {
				filterChain.FilterChainMatch = &listenerv3.FilterChainMatch{ServerNames: tt.serverNames}
			}

			err := addALPNMatch(xdsListener, filterChain, tt.alpnProtocols, tt.fingerprints)
			require.NoError(t, err)

			require.Equal(t, tt.expectMatch, filterChain.GetFilterChainMatch().GetApplicationProtocols())
			require.Equal(t, tt.serverNames, filterChain.GetFilterChainMatch().GetServerNames())

			if !tt.expectTLSInspector {
				require.Empty(t, xdsListener.ListenerFilters)
				return
			}
			require.Len(t, xdsListener.ListenerFilters, 1)
			require.Equal(t, wellknown.TlsInspector, xdsListener.ListenerFilters[0].Name)

			tlsInspector := &tls_inspectorv3.TlsInspector{}
			require.NoError(t, xdsListener.ListenerFilters[0].GetTypedConfig().UnmarshalTo(tlsInspector))
			require.Equal(t, len(tt.fingerprints) > 0, tlsInspector.GetEnableJa3Fingerprinting().GetValue())
		})
	}
}

func TestAddALPNMatchNilListener(t *testing.T) {
	filterChain := &listenerv3.FilterChain{}
	require.NoError(t, addALPNMatch(nil, filterChain, []string{"h2"}, nil))
	require.Nil(t, filterChain.FilterChainMatch)
}

func TestAddALPNMatchReusesTLSInspector(t *testing.T) {
	xdsListener := &listenerv3.Listener{}
	first := &listenerv3.FilterChain{}
	second := &listenerv3.FilterChain{}

	require.NoError(t, addALPNMatch(xdsListener, first, []string{"h2"}, nil))
	require.NoError(t, addALPNMatch(xdsListener, second, []string{"h2"}, nil))

	require.Len(t, xdsListener.ListenerFilters, 1)
	require.Equal(t, []string{"h2"}, first.GetFilterChainMatch().GetApplicationProtocols())
	require.Equal(t, []string{"h2"}, second.GetFilterChainMatch().GetApplicationProtocols())
}

func TestTLSInspectorKeepsFingerprintsOfEveryFilterChain(t *testing.T) {
	tests := []struct {
		name      string
		first     func(*listenerv3.Listener) error
		expectJA3 bool
	}{
		{
			name: "ALPN match first",
			first: func(xdsListener *listenerv3.Listener) error {
				return addALPNMatch(xdsListener, &listenerv3.FilterChain{}, []string{"h2"}, nil)
			},
		},
		{
			name: "server names match first",
			first: func(xdsListener *listenerv3.Listener) error {
				return addServerNamesMatch(xdsListener, &listenerv3.FilterChain{}, []string{"foo.com"}, nil)
			},
		},
		{
			name: "ALPN match with JA3 first",
			first: func(xdsListener *listenerv3.Listener) error {
				return addALPNMatch(xdsListener, &listenerv3.FilterChain{}, []string{"h2"}, []ir.TLSFingerprintType{ir.TLSFingerprintTypeJA3})
			},
			expectJA3: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			xdsListener := &listenerv3.Listener{}
			require.NoError(t, tt.first(xdsListener))

			// A later filter chain on the same listener asks for JA4.
			fingerprints := []ir.TLSFingerprintType{ir.TLSFingerprintTypeJA4}
			require.NoError(t, addServerNamesMatch(xdsListener, &listenerv3.FilterChain{}, []string{"bar.com"}, fingerprints))

			require.Len(t, xdsListener.ListenerFilters, 1)
			tlsInspector := &tls_inspectorv3.TlsInspector{}
			require.NoError(t, xdsListener.ListenerFilters[0].GetTypedConfig().UnmarshalTo(tlsInspector))
			require.True(t, tlsInspector.GetEnableJa4Fingerprinting().GetValue())
			require.Equal(t, tt.expectJA3, tlsInspector.GetEnableJa3Fingerprinting().GetValue())
		})
	}
}
