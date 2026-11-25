// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"regexp"
	"slices"
	"strings"
	"testing"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	rbacconfigv3 "github.com/envoyproxy/go-control-plane/envoy/config/rbac/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	rbacv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/rbac/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/stretchr/testify/require"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

func TestQUICWildcardClientValidation(t *testing.T) {
	for _, wildcard := range []string{"*.example.com", "*", "*.other.com"} {
		t.Run(wildcard, func(t *testing.T) {
			for _, required := range []bool{false, true} {
				mode := "optional"
				if required {
					mode = "required"
				}
				t.Run(mode, func(t *testing.T) {
					for _, order := range []string{"protected-first", "wildcard-first"} {
						t.Run(order, func(t *testing.T) {
							x := requireXdsIRFromInputTestData(t, "testdata/in/xds-ir/http3.yaml")
							public, protected, sibling := x.HTTP[0], x.HTTP[1], x.HTTP[2]
							public.Hostnames = []string{wildcard}
							publicRoute := protected.Routes[0].DeepCopy()
							publicRoute.Hostname = wildcard
							publicRoute.Security = &ir.SecurityFeatures{Authorization: &ir.Authorization{DefaultAction: egv1a1.AuthorizationActionAllow}}
							public.Routes = []*ir.HTTPRoute{publicRoute}
							protected.HTTP3 = nil
							protected.TLS.ClientValidationEnabled = true
							protected.TLS.RequireClientCertificate = required
							protected.TLS.CACertificate = &ir.TLSCACertificate{
								Name: "client-ca", Certificate: protected.TLS.Certificates[0].Certificate,
							}
							x.HTTP = []*ir.HTTPListener{protected, public, sibling}
							if order == "wildcard-first" {
								slices.Reverse(x.HTTP)
							}
							tr := &Translator{
								ControllerNamespace: "envoy-gateway-system",
								// Custom RBAC ordering must not move the protection filters.
								FilterOrder: []egv1a1.FilterPosition{{Name: egv1a1.EnvoyFilterRBAC, Before: new(egv1a1.EnvoyFilterCORS)}},
							}
							ctx, err := tr.Translate(t.Context(), x)
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
									require.Equal(t, quicClientValidationFilterName, mgr.HttpFilters[0].Name)
									require.Equal(t, quicClientValidationRouterFilterName, mgr.HttpFilters[len(mgr.HttpFilters)-2].Name)
									require.Equal(t, wellknown.Router, mgr.HttpFilters[len(mgr.HttpFilters)-1].Name)
									require.True(t, hcmContainsFilter(mgr, egv1a1.EnvoyFilterRBAC.String()))
									for _, guard := range []*hcmv3.HttpFilter{mgr.HttpFilters[0], mgr.HttpFilters[len(mgr.HttpFilters)-2]} {
										for _, authority := range []string{"foo.example.com", "foo.example.com:443", "foo.example.com:", "foo.example.com:garbage"} {
											require.True(t, quicGuardDenies(t, guard, "api.example.com", authority), authority)
										}
									}
									require.False(t, quicGuardDenies(t, mgr.HttpFilters[0], "api.example.com", "api.example.com"))
								}
								// The guard has no per-route config in the shared TCP/QUIC routes.
								cfg := findXdsRouteConfig(ctx, mgr.GetRds().GetRouteConfigName())
								require.NotNil(t, cfg)
								require.True(t, cfg.IgnorePortInHostMatching)
								for _, vHost := range cfg.VirtualHosts {
									for _, route := range vHost.Routes {
										require.NotContains(t, route.TypedPerFilterConfig, quicClientValidationFilterName)
										require.NotContains(t, route.TypedPerFilterConfig, quicClientValidationRouterFilterName)
										require.NotEmpty(t, route.ResponseHeadersToAdd)
									}
								}
							}
							for _, chain := range tcp.FilterChains {
								mgr, err := findHCMinFilterChain(chain)
								require.NoError(t, err)
								require.False(t, hcmContainsFilter(mgr, quicClientValidationFilterName))
								require.False(t, hcmContainsFilter(mgr, quicClientValidationRouterFilterName))
								if slices.Contains(chain.GetFilterChainMatch().GetServerNames(), "foo.example.com") {
									tls := &tlsv3.DownstreamTlsContext{}
									require.NoError(t, chain.TransportSocket.GetTypedConfig().UnmarshalTo(tls))
									require.Equal(t, required, tls.GetRequireClientCertificate().GetValue())
									require.NotNil(t, tls.CommonTlsContext.ValidationContextType)
								}
							}
						})
					}
				})
			}
		})
	}
}

func TestQUICWildcardClientValidationPrecedence(t *testing.T) {
	for _, broader := range [][]string{nil, {"*"}, {"*.example.com"}} {
		t.Run(strings.Join(broader, ""), func(t *testing.T) {
			for _, order := range []string{"protected-first", "public-first"} {
				t.Run(order, func(t *testing.T) {
					x := requireXdsIRFromInputTestData(t, "testdata/in/xds-ir/http3.yaml")
					public, protected, descendant := x.HTTP[0], x.HTTP[1], x.HTTP[2]
					public.Hostnames = []string{"*.public.example.com"}
					public.Routes = []*ir.HTTPRoute{protected.Routes[0].DeepCopy()}
					public.Routes[0].Hostname = public.Hostnames[0]
					protected.Hostnames = broader
					descendant.Hostnames = []string{"admin.public.example.com"}
					for _, listener := range []*ir.HTTPListener{protected, descendant} {
						listener.HTTP3 = nil
						listener.TLS.ClientValidationEnabled = true
						listener.TLS.RequireClientCertificate = true
						listener.TLS.CACertificate = &ir.TLSCACertificate{
							Name: "client-ca", Certificate: protected.TLS.Certificates[0].Certificate,
						}
					}
					x.HTTP = []*ir.HTTPListener{protected, descendant, public}
					if order == "public-first" {
						slices.Reverse(x.HTTP)
					}
					ctx, err := (&Translator{ControllerNamespace: "envoy-gateway-system"}).Translate(t.Context(), x)
					require.NoError(t, err)
					quic := findXdsListenerByHostPort(ctx, public.Address, public.Port, corev3.SocketAddress_UDP)
					require.NotNil(t, quic)
					require.Len(t, quic.FilterChains, 1)
					mgr, err := findHCMinFilterChain(quic.FilterChains[0])
					require.NoError(t, err)
					require.Equal(t, public.Hostnames, quic.FilterChains[0].GetFilterChainMatch().GetServerNames())
					for _, guard := range []*hcmv3.HttpFilter{mgr.HttpFilters[0], mgr.HttpFilters[len(mgr.HttpFilters)-2]} {
						require.False(t, quicGuardDenies(t, guard, "api.public.example.com", "api.public.example.com"))
						require.True(t, quicGuardDenies(t, guard, "api.public.example.com", "admin.public.example.com"))
						require.True(t, quicGuardDenies(t, guard, "admin.public.example.com", "api.public.example.com"))
						require.True(t, quicGuardDenies(t, guard, "api.public.example.com", "private.example.com"))
						require.True(t, quicGuardDenies(t, guard, "private.example.com", "api.public.example.com"))
					}
					tcp := findXdsListenerByHostPort(ctx, public.Address, public.Port, corev3.SocketAddress_TCP)
					require.NotNil(t, tcp)
					var publicChain *listenerv3.FilterChain
					for _, chain := range tcp.FilterChains {
						if slices.Contains(chain.GetFilterChainMatch().GetServerNames(), public.Hostnames[0]) {
							publicChain = chain
						}
					}
					require.NotNil(t, publicChain)
					tls := &tlsv3.DownstreamTlsContext{}
					require.NoError(t, publicChain.TransportSocket.GetTypedConfig().UnmarshalTo(tls))
					require.Nil(t, tls.CommonTlsContext.ValidationContextType)
				})
			}
		})
	}
}

func TestQUICClientValidationDenyRules(t *testing.T) {
	guard, err := buildQUICClientValidationFilter([]string{"foo.example.com", "*.secure.example.com"}, nil)
	require.NoError(t, err)
	for _, tt := range []struct {
		name, sni, authority string
		deny                 bool
	}{
		{name: "protected SNI and authority", sni: "foo.example.com", authority: "foo.example.com", deny: true},
		{name: "connection reuse", sni: "bar.example.com", authority: "foo.example.com", deny: true},
		{name: "protected SNI with public authority", sni: "foo.example.com", authority: "bar.example.com", deny: true},
		{name: "public request", sni: "bar.example.com", authority: "bar.example.com"},
		{name: "no SNI", authority: "foo.example.com", deny: true},
		{name: "case and port", sni: "bar.example.com", authority: "FOO.EXAMPLE.COM:443", deny: true},
		{name: "empty port", sni: "bar.example.com", authority: "foo.example.com:", deny: true},
		{name: "nonnumeric port", sni: "bar.example.com", authority: "foo.example.com:garbage", deny: true},
		{name: "case trailing dot and nonnumeric port", authority: "FOO.EXAMPLE.COM.:garbage", deny: true},
		{name: "wildcard empty port", authority: "api.secure.example.com:", deny: true},
		{name: "wildcard colon in routed hostname", authority: "x:admin.secure.example.com:443", deny: true},
		{name: "wildcard bracketed prefix", authority: "[x:admin].secure.example.com", deny: true},
		{name: "colon before wildcard suffix is stripped", authority: "x:admin.secure.example.com"},
		{name: "nested wildcard nonnumeric port", authority: "api.team.secure.example.com:garbage", deny: true},
		{name: "extra colon changes route hostname", authority: "foo.example.com:evil:443"},
		{name: "closing bracket prevents port stripping", authority: "foo.example.com:garbage]"},
		{name: "trailing dot and port", authority: "foo.example.com.:8443", deny: true},
		{name: "SNI case and trailing dot", sni: "FOO.EXAMPLE.COM.", authority: "bar.example.com", deny: true},
		{name: "nested wildcard", authority: "api.team.secure.example.com", deny: true},
		{name: "wildcard apex", authority: "secure.example.com"},
		{name: "suffix lookalike", authority: "foo.example.com.evil"},
		{name: "prefix lookalike", authority: "notfoo.example.com"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.deny, quicGuardDenies(t, guard, tt.sni, tt.authority))
		})
	}
}

func TestQUICClientValidationPublicExceptions(t *testing.T) {
	guard, err := buildQUICClientValidationFilter(
		[]string{"*.example.com", "*.secure.public.example.com"},
		[]string{"*.public.example.com", "status.example.com", "open.secure.public.example.com"},
	)
	require.NoError(t, err)
	for _, tt := range []struct {
		name, sni, authority string
		deny                 bool
	}{
		{name: "public wildcard", sni: "api.public.example.com", authority: "api.public.example.com"},
		{name: "public exact", sni: "status.example.com", authority: "status.example.com"},
		{name: "case and port", sni: "API.PUBLIC.EXAMPLE.COM.", authority: "STATUS.EXAMPLE.COM.:443"},
		{name: "public exact empty port", sni: "status.example.com", authority: "status.example.com:"},
		{name: "public exact nonnumeric port", sni: "status.example.com", authority: "STATUS.EXAMPLE.COM.:garbage"},
		{name: "public wildcard empty port", sni: "api.public.example.com", authority: "api.public.example.com:"},
		{name: "public wildcard nonnumeric port", sni: "api.public.example.com", authority: "api.public.example.com:garbage"},
		{name: "protected descendant empty port", sni: "api.public.example.com", authority: "admin.secure.public.example.com:", deny: true},
		{name: "protected descendant nonnumeric port", sni: "api.public.example.com", authority: "admin.secure.public.example.com:garbage", deny: true},
		{name: "protected SNI with public malformed authority", sni: "admin.secure.public.example.com", authority: "api.public.example.com:garbage", deny: true},
		{name: "public exception beneath protected wildcard", sni: "open.secure.public.example.com", authority: "open.secure.public.example.com"},
		{name: "protected descendant wildcard", sni: "admin.secure.public.example.com", authority: "admin.secure.public.example.com", deny: true},
		{name: "protected authority with public SNI", sni: "api.public.example.com", authority: "admin.secure.public.example.com", deny: true},
		{name: "protected SNI with public authority", sni: "admin.secure.public.example.com", authority: "api.public.example.com", deny: true},
		{name: "public wildcard apex remains protected", sni: "public.example.com", authority: "public.example.com", deny: true},
		{name: "public exact lookalike", sni: "notstatus.example.com", authority: "notstatus.example.com", deny: true},
		{name: "unrelated hostname", sni: "example.org", authority: "example.org"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.deny, quicGuardDenies(t, guard, tt.sni, tt.authority))
		})
	}
}

func TestQUICClientValidationPublicAuthorityPortExceptions(t *testing.T) {
	guard, err := buildQUICClientValidationFilter([]string{"*"}, []string{"status.example.com", "*.public.example.com"})
	require.NoError(t, err)
	for _, hostname := range []string{"status.example.com", "api.public.example.com"} {
		for _, suffix := range []string{"", ":443", ":", ":garbage", ":evil:443", ":garbage]"} {
			t.Run(hostname+suffix, func(t *testing.T) {
				// Extra colons and closing brackets must not exempt authorities
				// whose routed hostname differs from the public hostname.
				deny := suffix == ":evil:443" || suffix == ":garbage]"
				require.Equal(t, deny, quicGuardDenies(t, guard, hostname, hostname+suffix))
			})
		}
	}
}

// Match the authority hostname used by Envoy's getPortStart and wildcard
// virtual-host lookup, independently of the generated RBAC regexes.
func TestQUICClientValidationEnvoyAuthorityMatching(t *testing.T) {
	protectedGuard, err := buildQUICClientValidationFilter([]string{"*.public.example.com"}, nil)
	require.NoError(t, err)
	publicGuard, err := buildQUICClientValidationFilter([]string{"*"}, []string{"*.public.example.com"})
	require.NoError(t, err)
	for _, prefix := range []string{"", "admin", "x:admin", "[x:admin]", "x]:admin", "x:admin]", "x:admin]:evil", "[admin]", "x::admin"} {
		for _, port := range []string{"", ":443", ":", ":garbage", ":garbage]", ":evil:443", "]:443", ":a]:port"} {
			authority := prefix + ".public.example.com" + port
			t.Run(authority, func(t *testing.T) {
				host := authority
				if colon := strings.LastIndexByte(host, ':'); colon >= 0 && strings.LastIndexByte(host, ']') < colon {
					host = host[:colon]
				}
				matches := len(host) > len(".public.example.com") && strings.HasSuffix(host, ".public.example.com")
				require.Equal(t, matches, quicGuardDenies(t, protectedGuard, "unrelated.example.com", authority), "protected wildcard")
				require.Equal(t, !matches, quicGuardDenies(t, publicGuard, "api.public.example.com", authority), "public wildcard exemption")
			})
		}
	}
}

func TestQUICClientValidationLongHostnames(t *testing.T) {
	for _, letter := range []string{"a", "k", "s"} {
		// k and s also exercise Unicode case folding in Envoy's UTF-8 RE2 mode.
		suffix := strings.Join([]string{strings.Repeat(letter, 63), strings.Repeat(letter, 63), strings.Repeat(letter, 63), strings.Repeat(letter, 59)}, ".")
		wildcard := "*." + suffix
		hostname := "x." + suffix
		require.Len(t, wildcard, 253)
		for _, tt := range []struct {
			name              string
			protected, public []string
			deny              bool
		}{
			{name: "protected wildcard", protected: []string{wildcard}, deny: true},
			{name: "protected exact", protected: []string{hostname}, deny: true},
			{name: "public wildcard exemption", protected: []string{"*"}, public: []string{wildcard}},
		} {
			t.Run(letter+"/"+tt.name, func(t *testing.T) {
				guard, err := buildQUICClientValidationFilter(tt.protected, tt.public)
				require.NoError(t, err)
				require.Equal(t, tt.deny, quicGuardDenies(t, guard, hostname, strings.ToUpper(hostname)+".:443"))
				if tt.deny {
					require.True(t, quicGuardDenies(t, guard, "public.example.com", hostname))
					require.True(t, quicGuardDenies(t, guard, hostname, "public.example.com"))
				} else {
					require.True(t, quicGuardDenies(t, guard, "protected.example.com", hostname))
				}
				config := &rbacv3.RBAC{}
				require.NoError(t, guard.GetTypedConfig().UnmarshalTo(config))
				for _, policy := range config.Rules.Policies {
					for _, permission := range policy.Permissions {
						assertQUICRegexLimit(t, permission)
					}
				}
			})
		}
	}
}

func assertQUICRegexLimit(t *testing.T, permission *rbacconfigv3.Permission) {
	t.Helper()
	for _, rule := range permission.GetAndRules().GetRules() {
		assertQUICRegexLimit(t, rule)
	}
	for _, rule := range permission.GetOrRules().GetRules() {
		assertQUICRegexLimit(t, rule)
	}
	if rule := permission.GetNotRule(); rule != nil {
		assertQUICRegexLimit(t, rule)
	}
	var matcher *matcherv3.RegexMatcher
	if sni := permission.GetRequestedServerName(); sni != nil {
		matcher = sni.GetSafeRegex()
	} else if header := permission.GetHeader(); header != nil {
		matcher = header.GetSafeRegexMatch()
	}
	if matcher != nil {
		require.EqualValues(t, 2048, matcher.GetGoogleRe2().GetMaxProgramSize().GetValue())
	}
}

func TestQUICHostnameOverlap(t *testing.T) {
	for _, tt := range []struct {
		wildcard, protected string
		want                bool
	}{
		{wildcard: "*.example.com", protected: "foo.example.com", want: true},
		{wildcard: "*.example.com", protected: "example.com"},
		{wildcard: "*.example.com", protected: "foo.other.com"},
		{wildcard: "*.example.com", protected: "*.secure.example.com", want: true},
		{wildcard: "*.secure.example.com", protected: "*.example.com", want: true},
		{wildcard: "*.example.com", protected: "*.other.com"},
		{wildcard: "*.example.com", protected: "*", want: true},
	} {
		t.Run(tt.wildcard+"/"+tt.protected, func(t *testing.T) {
			require.Equal(t, tt.want, quicHostnamesOverlap(tt.wildcard, tt.protected))
		})
	}
}

func TestQUICClientValidationTransportScope(t *testing.T) {
	for _, tt := range []struct {
		name     string
		protocol corev3.SocketAddress_Protocol
		codec    hcmv3.HttpConnectionManager_CodecType
		port     uint32
		address  string
	}{
		{name: "TCP", protocol: corev3.SocketAddress_TCP, codec: hcmv3.HttpConnectionManager_HTTP3, address: "0.0.0.0", port: 443},
		{name: "UDP without HTTP3", protocol: corev3.SocketAddress_UDP, codec: hcmv3.HttpConnectionManager_AUTO, address: "0.0.0.0", port: 443},
		{name: "different port", protocol: corev3.SocketAddress_UDP, codec: hcmv3.HttpConnectionManager_HTTP3, address: "0.0.0.0", port: 8443},
		{name: "different address", protocol: corev3.SocketAddress_UDP, codec: hcmv3.HttpConnectionManager_HTTP3, address: "127.0.0.1", port: 443},
	} {
		t.Run(tt.name, func(t *testing.T) {
			mgrFilter, err := toNetworkFilter(wellknown.HTTPConnectionManager, &hcmv3.HttpConnectionManager{
				StatPrefix: "scope", CodecType: tt.codec,
				RouteSpecifier: &hcmv3.HttpConnectionManager_RouteConfig{RouteConfig: &routev3.RouteConfiguration{Name: "scope"}},
			})
			require.NoError(t, err)
			listener := &listenerv3.Listener{Address: &corev3.Address{Address: &corev3.Address_SocketAddress{
				SocketAddress: &corev3.SocketAddress{Protocol: tt.protocol, Address: tt.address, PortSpecifier: &corev3.SocketAddress_PortValue{PortValue: tt.port}},
			}}, FilterChains: []*listenerv3.FilterChain{{Filters: []*listenerv3.Filter{mgrFilter}}}}
			ctx := &types.ResourceVersionTable{}
			require.NoError(t, ctx.AddXdsResource(resourcev3.ListenerType, listener))
			require.NoError(t, patchQUICWildcardClientValidation(ctx, []*ir.HTTPListener{{
				CoreListenerDetails: ir.CoreListenerDetails{Address: "0.0.0.0", Port: 443},
				Hostnames:           []string{"foo.example.com"},
				TLS:                 &ir.TLSConfig{ClientValidationEnabled: true},
			}}, nil))
			mgr, err := findHCMinFilterChain(listener.FilterChains[0])
			require.NoError(t, err)
			require.False(t, hcmContainsFilter(mgr, quicClientValidationFilterName))
		})
	}
}

// Evaluate the emitted regex permissions, including the OR semantics of RBAC DENY.
func quicGuardDenies(t *testing.T, filter *hcmv3.HttpFilter, sni, authority string) bool {
	t.Helper()
	config := &rbacv3.RBAC{}
	require.NoError(t, filter.GetTypedConfig().UnmarshalTo(config))
	require.Equal(t, rbacconfigv3.RBAC_DENY, config.Rules.Action)
	for _, policy := range config.Rules.Policies {
		require.True(t, policy.Principals[0].GetAny())
		for _, permission := range policy.Permissions {
			if quicPermissionMatches(t, permission, sni, authority) {
				return true
			}
		}
	}
	return false
}

func quicPermissionMatches(t *testing.T, permission *rbacconfigv3.Permission, sni, authority string) bool {
	t.Helper()
	if rules := permission.GetAndRules(); rules != nil {
		for _, rule := range rules.Rules {
			if !quicPermissionMatches(t, rule, sni, authority) {
				return false
			}
		}
		return true
	}
	if rules := permission.GetOrRules(); rules != nil {
		for _, rule := range rules.Rules {
			if quicPermissionMatches(t, rule, sni, authority) {
				return true
			}
		}
		return false
	}
	if rule := permission.GetNotRule(); rule != nil {
		return !quicPermissionMatches(t, rule, sni, authority)
	}
	var pattern, value string
	if matcher := permission.GetRequestedServerName(); matcher != nil {
		pattern, value = matcher.GetSafeRegex().Regex, sni
	} else {
		header := permission.GetHeader()
		require.NotNil(t, header)
		require.Equal(t, AuthorityHeaderKey, header.Name)
		pattern, value = header.GetSafeRegexMatch().Regex, authority
	}
	matcher, err := regexp.Compile(pattern)
	require.NoError(t, err)
	return matcher.MatchString(value)
}

func TestQUICWildcardAuthorityClientValidation(t *testing.T) {
	for _, transport := range []string{"HTTPS", "TLS termination"} {
		t.Run(transport, func(t *testing.T) {
			x := requireXdsIRFromInputTestData(t, "testdata/in/xds-ir/http3.yaml")
			public, protected := x.HTTP[0], x.HTTP[1]
			public.Hostnames = []string{"*.example.com"}
			public.Routes = []*ir.HTTPRoute{protected.Routes[0].DeepCopy()}
			public.Routes[0].Hostname = "*.example.com"
			protected.Hostnames = []string{"*.secure.example.com"}
			protected.HTTP3 = nil
			protected.TLS.ClientValidationEnabled = true
			protected.TLS.RequireClientCertificate = true
			protected.TLS.CACertificate = &ir.TLSCACertificate{Name: "client-ca", Certificate: protected.TLS.Certificates[0].Certificate}
			x.HTTP = []*ir.HTTPListener{public, protected}
			if transport == "TLS termination" {
				x.HTTP = []*ir.HTTPListener{public}
				x.TCP = []*ir.TCPListener{quicTestTCPListener(protected, "protected-wildcard", protected.Hostnames, protected.TLS)}
			}
			ctx, err := (&Translator{ControllerNamespace: "envoy-gateway-system"}).Translate(t.Context(), x)
			require.NoError(t, err)
			quic := findXdsListenerByHostPort(ctx, public.Address, public.Port, corev3.SocketAddress_UDP)
			require.NotNil(t, quic)
			require.Len(t, quic.FilterChains, 1)
			mgr, err := findHCMinFilterChain(quic.FilterChains[0])
			require.NoError(t, err)
			cfg := findXdsRouteConfig(ctx, mgr.GetRds().GetRouteConfigName())
			require.True(t, cfg.IgnorePortInHostMatching)
			require.Equal(t, []string{"*.example.com"}, cfg.VirtualHosts[0].Domains)
			for _, authority := range []string{
				"admin.secure.example.com:443",
				"x:admin.secure.example.com:443",
				"foo:bar.secure.example.com:443",
				"foo:bar.secure.example.com:garbage",
				"[x:admin].secure.example.com",
			} {
				t.Run(authority, func(t *testing.T) {
					denied := make([]bool, 0, 2)
					for _, guard := range []*hcmv3.HttpFilter{mgr.HttpFilters[0], mgr.HttpFilters[len(mgr.HttpFilters)-2]} {
						denied = append(denied, quicGuardDenies(t, guard, "api.example.com", authority))
					}
					require.Equal(t, []bool{true, true}, denied)
				})
			}
		})
	}
}
