// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"fmt"
	"slices"
	"testing"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/stretchr/testify/require"

	"github.com/envoyproxy/gateway/internal/ir"
)

func TestQUICWildcardTCPClientValidation(t *testing.T) {
	for _, wildcard := range []string{"*.example.com", "*", "*.other.com"} {
		for _, required := range []bool{true, false} {
			for _, routeOnly := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/required=%t/route-only=%t", wildcard, required, routeOnly), func(t *testing.T) {
					x := requireXdsIRFromInputTestData(t, "testdata/in/xds-ir/http3.yaml")
					public, protected, sibling := x.HTTP[0], x.HTTP[1], x.HTTP[2]
					public.Hostnames = []string{wildcard}
					public.Routes = []*ir.HTTPRoute{protected.Routes[0].DeepCopy()}
					public.Routes[0].Hostname = wildcard
					tls := protected.TLS.DeepCopy()
					tls.ClientValidationEnabled = true
					tls.RequireClientCertificate = required
					tls.CACertificate = &ir.TLSCACertificate{Name: "client-ca", Certificate: tls.Certificates[0].Certificate}
					tcpListener := quicTestTCPListener(protected, "protected-tcp", []string{"foo.example.com"}, tls)
					if routeOnly {
						tcpListener.TLS = protected.TLS.DeepCopy()
						tcpListener.Hostnames = []string{"bar.example.com"}
						tcpListener.Routes[0].TLS.TLSInspectorConfig = &ir.TLSInspectorConfig{SNIs: []string{"foo.example.com"}}
					}
					x.HTTP = []*ir.HTTPListener{public, sibling}
					x.TCP = []*ir.TCPListener{tcpListener}
					ctx, err := (&Translator{ControllerNamespace: "envoy-gateway-system"}).Translate(t.Context(), x)
					require.NoError(t, err)
					quic := findXdsListenerByHostPort(ctx, public.Address, public.Port, corev3.SocketAddress_UDP)
					tcp := findXdsListenerByHostPort(ctx, public.Address, public.Port, corev3.SocketAddress_TCP)
					require.NotNil(t, quic)
					require.NotNil(t, tcp)
					require.Len(t, quic.FilterChains, 2)
					for _, chain := range quic.FilterChains {
						mgr, err := findHCMinFilterChain(chain)
						require.NoError(t, err)
						isWildcard := len(chain.GetFilterChainMatch().GetServerNames()) == 0 || slices.Contains(chain.GetFilterChainMatch().GetServerNames(), wildcard)
						wantGuard := isWildcard && wildcard != "*.other.com"
						require.Equal(t, wantGuard, hcmContainsFilter(mgr, quicClientValidationFilterName))
						require.Equal(t, wantGuard, hcmContainsFilter(mgr, quicClientValidationRouterFilterName))
						if wantGuard {
							for _, guard := range []*hcmv3.HttpFilter{mgr.HttpFilters[0], mgr.HttpFilters[len(mgr.HttpFilters)-2]} {
								require.True(t, quicGuardDenies(t, guard, "foo.example.com", "foo.example.com"))
								for _, authority := range []string{"foo.example.com", "foo.example.com:443", "foo.example.com:", "foo.example.com:garbage"} {
									require.True(t, quicGuardDenies(t, guard, "api.example.com", authority), authority)
								}
								require.True(t, quicGuardDenies(t, guard, "foo.example.com", "api.example.com"))
								require.False(t, quicGuardDenies(t, guard, "api.example.com", "api.example.com"))
							}
						}
					}
					foundProtected := false
					for _, chain := range tcp.FilterChains {
						if slices.Contains(chain.GetFilterChainMatch().GetServerNames(), "foo.example.com") {
							config := &tlsv3.DownstreamTlsContext{}
							require.NoError(t, chain.TransportSocket.GetTypedConfig().UnmarshalTo(config))
							require.Equal(t, required, config.GetRequireClientCertificate().GetValue())
							require.NotNil(t, config.CommonTlsContext.ValidationContextType)
							foundProtected = true
						}
					}
					require.True(t, foundProtected)
				})
			}
		}
	}
}

func TestQUICClientValidationTCPPrecedence(t *testing.T) {
	for _, publicHostname := range []string{"foo.example.com", "*.public.example.com"} {
		for _, publicFirst := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/public-first=%t", publicHostname, publicFirst), func(t *testing.T) {
				x := requireXdsIRFromInputTestData(t, "testdata/in/xds-ir/http3.yaml")
				public, base := x.HTTP[0], x.HTTP[1]
				public.Hostnames = []string{"*"}
				public.Routes = []*ir.HTTPRoute{base.Routes[0].DeepCopy()}
				public.Routes[0].Hostname = "*"
				protectedTLS := base.TLS.DeepCopy()
				protectedTLS.ClientValidationEnabled = true
				protectedTLS.RequireClientCertificate = true
				protectedTLS.CACertificate = &ir.TLSCACertificate{Name: "client-ca", Certificate: protectedTLS.Certificates[0].Certificate}
				x.HTTP = []*ir.HTTPListener{public}
				x.TCP = []*ir.TCPListener{
					quicTestTCPListener(base, "protected-wildcard", []string{"*.example.com"}, protectedTLS),
					quicTestTCPListener(base, "public-tcp", []string{publicHostname}, base.TLS.DeepCopy()),
				}
				requestHostname := publicHostname
				if publicHostname == "*.public.example.com" {
					requestHostname = "api.public.example.com"
					x.TCP = append(x.TCP, quicTestTCPListener(base, "protected-descendant", []string{"admin.public.example.com"}, protectedTLS.DeepCopy()))
				}
				if publicFirst {
					slices.Reverse(x.TCP)
				}
				ctx, err := (&Translator{ControllerNamespace: "envoy-gateway-system"}).Translate(t.Context(), x)
				require.NoError(t, err)
				quic := findXdsListenerByHostPort(ctx, public.Address, public.Port, corev3.SocketAddress_UDP)
				require.NotNil(t, quic)
				require.Len(t, quic.FilterChains, 1)
				mgr, err := findHCMinFilterChain(quic.FilterChains[0])
				require.NoError(t, err)
				for _, guard := range []*hcmv3.HttpFilter{mgr.HttpFilters[0], mgr.HttpFilters[len(mgr.HttpFilters)-2]} {
					require.False(t, quicGuardDenies(t, guard, requestHostname, requestHostname))
					require.True(t, quicGuardDenies(t, guard, requestHostname, "private.example.com"))
					require.True(t, quicGuardDenies(t, guard, "private.example.com", requestHostname))
					if publicHostname == "*.public.example.com" {
						require.True(t, quicGuardDenies(t, guard, requestHostname, "admin.public.example.com"))
						require.True(t, quicGuardDenies(t, guard, "admin.public.example.com", requestHostname))
					}
				}
			})
		}
	}
}

func TestQUICUnvalidatedTCPListener(t *testing.T) {
	for _, passthrough := range []bool{false, true} {
		t.Run(fmt.Sprintf("passthrough=%t", passthrough), func(t *testing.T) {
			x := requireXdsIRFromInputTestData(t, "testdata/in/xds-ir/http3.yaml")
			public, base := x.HTTP[0], x.HTTP[1]
			public.Hostnames = []string{"*.example.com"}
			tcp := quicTestTCPListener(base, "public-tcp", []string{"foo.example.com"}, base.TLS.DeepCopy())
			if passthrough {
				tcp.TLS = nil
				tcp.Routes[0].TLS = &ir.TLS{TLSInspectorConfig: &ir.TLSInspectorConfig{SNIs: tcp.Hostnames}}
			}
			x.HTTP = []*ir.HTTPListener{public}
			x.TCP = []*ir.TCPListener{tcp}
			ctx, err := (&Translator{ControllerNamespace: "envoy-gateway-system"}).Translate(t.Context(), x)
			require.NoError(t, err)
			quic := findXdsListenerByHostPort(ctx, public.Address, public.Port, corev3.SocketAddress_UDP)
			require.NotNil(t, quic)
			mgr, err := findHCMinFilterChain(quic.FilterChains[0])
			require.NoError(t, err)
			require.False(t, hcmContainsFilter(mgr, quicClientValidationFilterName))
		})
	}
}

func TestQUICClientValidationTCPPassthroughPrecedence(t *testing.T) {
	x := requireXdsIRFromInputTestData(t, "testdata/in/xds-ir/http3.yaml")
	public, base := x.HTTP[0], x.HTTP[1]
	public.Hostnames = []string{"*"}
	protectedTLS := base.TLS.DeepCopy()
	protectedTLS.ClientValidationEnabled = true
	protectedTLS.RequireClientCertificate = true
	protectedTLS.CACertificate = &ir.TLSCACertificate{Name: "client-ca", Certificate: protectedTLS.Certificates[0].Certificate}
	passthrough := quicTestTCPListener(base, "passthrough", []string{"foo.example.com"}, nil)
	passthrough.Routes[0].TLS = &ir.TLS{TLSInspectorConfig: &ir.TLSInspectorConfig{SNIs: passthrough.Hostnames}}
	x.HTTP = []*ir.HTTPListener{public}
	x.TCP = []*ir.TCPListener{
		quicTestTCPListener(base, "protected-wildcard", []string{"*.example.com"}, protectedTLS),
		passthrough,
	}
	ctx, err := (&Translator{ControllerNamespace: "envoy-gateway-system"}).Translate(t.Context(), x)
	require.NoError(t, err)
	quic := findXdsListenerByHostPort(ctx, public.Address, public.Port, corev3.SocketAddress_UDP)
	require.NotNil(t, quic)
	mgr, err := findHCMinFilterChain(quic.FilterChains[0])
	require.NoError(t, err)
	// Passthrough does not establish that this hostname's backend is public.
	require.True(t, quicGuardDenies(t, mgr.HttpFilters[0], "foo.example.com", "foo.example.com"))
}

func TestQUICClientValidationTCPSocketScope(t *testing.T) {
	for _, differentAddress := range []bool{false, true} {
		t.Run(fmt.Sprintf("different-address=%t", differentAddress), func(t *testing.T) {
			x := requireXdsIRFromInputTestData(t, "testdata/in/xds-ir/http3.yaml")
			public, base := x.HTTP[0], x.HTTP[1]
			public.Hostnames = []string{"*.example.com"}
			tls := base.TLS.DeepCopy()
			tls.ClientValidationEnabled = true
			tls.RequireClientCertificate = true
			tls.CACertificate = &ir.TLSCACertificate{Name: "client-ca", Certificate: tls.Certificates[0].Certificate}
			protected := quicTestTCPListener(base, "protected", []string{"foo.example.com"}, tls)
			if differentAddress {
				protected.Address = "127.0.0.1"
			} else {
				protected.Port++
			}
			x.HTTP = []*ir.HTTPListener{public}
			x.TCP = []*ir.TCPListener{protected}
			ctx, err := (&Translator{ControllerNamespace: "envoy-gateway-system"}).Translate(t.Context(), x)
			require.NoError(t, err)
			quic := findXdsListenerByHostPort(ctx, public.Address, public.Port, corev3.SocketAddress_UDP)
			require.NotNil(t, quic)
			mgr, err := findHCMinFilterChain(quic.FilterChains[0])
			require.NoError(t, err)
			require.False(t, hcmContainsFilter(mgr, quicClientValidationFilterName))
		})
	}
}

func quicTestTCPListener(base *ir.HTTPListener, name string, hostnames []string, tls *ir.TLSConfig) *ir.TCPListener {
	destination := base.Routes[0].Destination.DeepCopy()
	destination.Name = name + "-cluster"
	listener := &ir.TCPListener{
		CoreListenerDetails: base.CoreListenerDetails,
		Hostnames:           hostnames,
		TLS:                 tls,
		Routes:              []*ir.TCPRoute{{Name: name + "-route", TLS: &ir.TLS{Terminate: tls}, Destination: destination}},
	}
	listener.Name = name
	return listener
}
