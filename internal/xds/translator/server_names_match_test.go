// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/stretchr/testify/require"

	"github.com/envoyproxy/gateway/internal/ir"
)

func TestAddServerNamesMatch(t *testing.T) {
	tests := []struct {
		name               string
		xdsListener        *listenerv3.Listener
		hostnames          []string
		fingerprints       []ir.TLSFingerprintType
		expectFilterChain  bool
		expectTLSInspector bool
		expectServerNames  []string
	}{
		{
			name:               "nil listener",
			xdsListener:        nil,
			hostnames:          []string{"example.com"},
			fingerprints:       nil,
			expectFilterChain:  false,
			expectTLSInspector: false,
			expectServerNames:  nil,
		},
		{
			name: "UDP (QUIC) listener for HTTP3",
			xdsListener: &listenerv3.Listener{
				Address: &corev3.Address{
					Address: &corev3.Address_SocketAddress{
						SocketAddress: &corev3.SocketAddress{
							Protocol: corev3.SocketAddress_UDP,
							Address:  "0.0.0.0",
							PortSpecifier: &corev3.SocketAddress_PortValue{
								PortValue: 443,
							},
						},
					},
				},
			},
			hostnames:          []string{"example.com"},
			fingerprints:       []ir.TLSFingerprintType{},
			expectFilterChain:  true,
			expectTLSInspector: false,
			expectServerNames:  []string{"example.com"},
		},
		{
			name: "TCP listener with non-wildcard hostnames",
			xdsListener: &listenerv3.Listener{
				Address: &corev3.Address{
					Address: &corev3.Address_SocketAddress{
						SocketAddress: &corev3.SocketAddress{
							Protocol: corev3.SocketAddress_TCP,
							Address:  "0.0.0.0",
							PortSpecifier: &corev3.SocketAddress_PortValue{
								PortValue: 443,
							},
						},
					},
				},
			},
			hostnames:          []string{"example.com", "api.example.com"},
			fingerprints:       nil,
			expectFilterChain:  true,
			expectTLSInspector: true,
			expectServerNames:  []string{"example.com", "api.example.com"},
		},
		{
			name: "TCP listener with wildcard hostname",
			xdsListener: &listenerv3.Listener{
				Address: &corev3.Address{
					Address: &corev3.Address_SocketAddress{
						SocketAddress: &corev3.SocketAddress{
							Protocol: corev3.SocketAddress_TCP,
							Address:  "0.0.0.0",
							PortSpecifier: &corev3.SocketAddress_PortValue{
								PortValue: 443,
							},
						},
					},
				},
			},
			hostnames:          []string{"*"},
			fingerprints:       nil,
			expectFilterChain:  false,
			expectTLSInspector: false,
			expectServerNames:  nil,
		},
		{
			name: "TCP listener with wildcard hostname and fingerprint enabled",
			xdsListener: &listenerv3.Listener{
				Address: &corev3.Address{
					Address: &corev3.Address_SocketAddress{
						SocketAddress: &corev3.SocketAddress{
							Protocol: corev3.SocketAddress_TCP,
							Address:  "0.0.0.0",
							PortSpecifier: &corev3.SocketAddress_PortValue{
								PortValue: 443,
							},
						},
					},
				},
			},
			hostnames:          []string{"*"},
			fingerprints:       []ir.TLSFingerprintType{ir.TLSFingerprintTypeJA3, ir.TLSFingerprintTypeJA3},
			expectFilterChain:  false,
			expectTLSInspector: true,
			expectServerNames:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filterChain := &listenerv3.FilterChain{}

			err := addServerNamesMatch(tt.xdsListener, filterChain, tt.hostnames, tt.fingerprints)
			require.NoError(t, err)

			// Check if filter chain match was added
			if tt.expectFilterChain {
				require.NotNil(t, filterChain.FilterChainMatch)
				require.Equal(t, tt.expectServerNames, filterChain.FilterChainMatch.ServerNames)
			} else {
				require.Nil(t, filterChain.FilterChainMatch)
			}

			// Check if TLS inspector was added
			if tt.xdsListener != nil && tt.expectTLSInspector {
				hasTLSInspector := false
				for _, filter := range tt.xdsListener.ListenerFilters {
					if filter.Name == wellknown.TlsInspector {
						hasTLSInspector = true
						break
					}
				}
				require.True(t, hasTLSInspector, "TLS inspector filter should be added")
			} else if tt.xdsListener != nil {
				// For non-nil listeners that shouldn't have TLS inspector
				hasTLSInspector := false
				for _, filter := range tt.xdsListener.ListenerFilters {
					if filter.Name == wellknown.TlsInspector {
						hasTLSInspector = true
						break
					}
				}
				require.False(t, hasTLSInspector, "TLS inspector filter should not be added")
			}
		})
	}
}
