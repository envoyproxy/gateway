// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"reflect"
	"testing"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"

	"github.com/envoyproxy/gateway/internal/ir"
)

func Test_toNetworkFilter(t *testing.T) {
	tests := []struct {
		name    string
		proto   proto.Message
		wantErr error
	}{
		{
			name: "valid filter",
			proto: &hcmv3.HttpConnectionManager{
				StatPrefix: "stats",
				RouteSpecifier: &hcmv3.HttpConnectionManager_RouteConfig{
					RouteConfig: &routev3.RouteConfiguration{
						Name: "route",
					},
				},
			},
			wantErr: nil,
		},
		{
			name:    "invalid proto msg",
			proto:   &hcmv3.HttpConnectionManager{},
			wantErr: errors.New("invalid HttpConnectionManager.StatPrefix: value length must be at least 1 runes; invalid HttpConnectionManager.RouteSpecifier: value is required"),
		},
		{
			name:    "nil proto msg",
			proto:   nil,
			wantErr: errors.New("empty message received"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := toNetworkFilter("name", tt.proto)
			if tt.wantErr != nil {
				assert.Containsf(t, err.Error(), tt.wantErr.Error(), "toNetworkFilter(%v)", tt.proto)
			} else {
				assert.NoErrorf(t, err, "toNetworkFilter(%v)", tt.proto)
			}
		})
	}
}

func Test_buildTCPProxyHashPolicy(t *testing.T) {
	tests := []struct {
		name string
		lb   *ir.LoadBalancer
		want []*typev3.HashPolicy
	}{
		{
			name: "Nil LoadBalancer",
			lb:   nil,
			want: nil,
		},
		{
			name: "Nil ConsistentHash in LoadBalancer",
			lb:   &ir.LoadBalancer{},
			want: nil,
		},
		{
			name: "ConsistentHash without hash policy",
			lb:   &ir.LoadBalancer{ConsistentHash: &ir.ConsistentHash{}},
			want: nil,
		},
		{
			name: "ConsistentHash with SourceIP set to false",
			lb:   &ir.LoadBalancer{ConsistentHash: &ir.ConsistentHash{SourceIP: new(bool)}}, // *new(bool) defaults to false
			want: nil,
		},
		{
			name: "ConsistentHash with SourceIP set to true",
			lb:   &ir.LoadBalancer{ConsistentHash: &ir.ConsistentHash{SourceIP: func(b bool) *bool { return &b }(true)}},
			want: []*typev3.HashPolicy{{PolicySpecifier: &typev3.HashPolicy_SourceIp_{SourceIp: &typev3.HashPolicy_SourceIp{}}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildTCPProxyHashPolicy(tt.lb)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildTCPProxyHashPolicy() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_addXdsTLSInspectorFilter(t *testing.T) {
	tests := []struct {
		name        string
		xdsListener *listenerv3.Listener
		wantErr     bool
	}{
		{
			name:        "nil listener",
			xdsListener: nil,
			wantErr:     true,
		},
		{
			name: "valid listener",
			xdsListener: &listenerv3.Listener{
				Name:            "test-listener",
				ListenerFilters: []*listenerv3.ListenerFilter{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := addXdsTLSInspectorFilter(tt.xdsListener)
			if (err != nil) != tt.wantErr {
				t.Errorf("addXdsTLSInspectorFilter() error = %v, wantErr %v", err, tt.wantErr)
			}

			// If we didn't expect an error, verify the filter was added
			if !tt.wantErr && tt.xdsListener != nil {
				found := false
				for _, filter := range tt.xdsListener.ListenerFilters {
					if filter.Name == wellknown.TlsInspector {
						found = true
						break
					}
				}
				assert.True(t, found, "TLS inspector filter was not added to the listener")
			}
		})
	}
}

func Test_addServerNamesMatch(t *testing.T) {
	tests := []struct {
		name        string
		xdsListener *listenerv3.Listener
		filterChain *listenerv3.FilterChain
		hostnames   []string
		wantErr     bool
	}{
		{
			name:        "nil listener",
			xdsListener: nil,
			filterChain: &listenerv3.FilterChain{},
			hostnames:   []string{"example.com"},
			wantErr:     false, // Should not error with nil listener
		},
		{
			name: "UDP listener",
			xdsListener: &listenerv3.Listener{
				Name: "udp-listener",
				Address: &corev3.Address{
					Address: &corev3.Address_SocketAddress{
						SocketAddress: &corev3.SocketAddress{
							Protocol: corev3.SocketAddress_UDP,
						},
					},
				},
			},
			filterChain: &listenerv3.FilterChain{},
			hostnames:   []string{"example.com"},
			wantErr:     false, // Should not error with UDP listener
		},
		{
			name: "HTTP3 UDP listener",
			xdsListener: &listenerv3.Listener{
				Name: "http3-listener",
				Address: &corev3.Address{
					Address: &corev3.Address_SocketAddress{
						SocketAddress: &corev3.SocketAddress{
							Protocol: corev3.SocketAddress_UDP,
						},
					},
				},
				UdpListenerConfig: &listenerv3.UdpListenerConfig{
					QuicOptions: &listenerv3.QuicProtocolOptions{},
				},
			},
			filterChain: &listenerv3.FilterChain{},
			hostnames:   []string{"example.com"},
			wantErr:     false, // Should not error with HTTP3 UDP listener
		},
		{
			name: "TCP listener",
			xdsListener: &listenerv3.Listener{
				Name: "tcp-listener",
				Address: &corev3.Address{
					Address: &corev3.Address_SocketAddress{
						SocketAddress: &corev3.SocketAddress{
							Protocol: corev3.SocketAddress_TCP,
						},
					},
				},
				ListenerFilters: []*listenerv3.ListenerFilter{},
			},
			filterChain: &listenerv3.FilterChain{},
			hostnames:   []string{"example.com"},
			wantErr:     false,
		},
		{
			name: "wildcard hostname",
			xdsListener: &listenerv3.Listener{
				Name: "tcp-listener",
				Address: &corev3.Address{
					Address: &corev3.Address_SocketAddress{
						SocketAddress: &corev3.SocketAddress{
							Protocol: corev3.SocketAddress_TCP,
						},
					},
				},
				ListenerFilters: []*listenerv3.ListenerFilter{},
			},
			filterChain: &listenerv3.FilterChain{},
			hostnames:   []string{"*"},
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := addServerNamesMatch(tt.xdsListener, tt.filterChain, tt.hostnames)
			if (err != nil) != tt.wantErr {
				t.Errorf("addServerNamesMatch() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Check if filter chain match was set correctly for non-wildcard hostnames
			if len(tt.hostnames) > 0 && tt.hostnames[0] != "*" {
				// For UDP listeners, FilterChainMatch should not be set
				if tt.xdsListener != nil && tt.xdsListener.GetAddress() != nil &&
					tt.xdsListener.GetAddress().GetSocketAddress() != nil &&
					tt.xdsListener.GetAddress().GetSocketAddress().GetProtocol() == corev3.SocketAddress_UDP {
					assert.Nil(t, tt.filterChain.FilterChainMatch, "FilterChainMatch should not be set for UDP listeners")
				} else {
					assert.NotNil(t, tt.filterChain.FilterChainMatch, "FilterChainMatch should be set for non-wildcard hostnames on TCP listeners")
					assert.Equal(t, tt.hostnames, tt.filterChain.FilterChainMatch.ServerNames, "ServerNames should match hostnames")
				}
			}

			// Check if TLS inspector filter was added for TCP listeners with non-wildcard hostnames
			if tt.xdsListener != nil && len(tt.hostnames) > 0 && tt.hostnames[0] != "*" &&
				tt.xdsListener.GetAddress() != nil && tt.xdsListener.GetAddress().GetSocketAddress() != nil &&
				tt.xdsListener.GetAddress().GetSocketAddress().GetProtocol() == corev3.SocketAddress_TCP {
				found := false
				for _, filter := range tt.xdsListener.ListenerFilters {
					if filter.Name == wellknown.TlsInspector {
						found = true
						break
					}
				}
				assert.True(t, found, "TLS inspector filter should be added for TCP listeners with non-wildcard hostnames")
			}
		})
	}
}
