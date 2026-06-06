// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	rbacconfigv3 "github.com/envoyproxy/go-control-plane/envoy/config/rbac/v3"
	networkrbacv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/rbac/v3"
	snidfpv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/sni_dynamic_forward_proxy/v3"
	envoymatcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"

	"github.com/envoyproxy/gateway/internal/ir"
)

const (
	networkSNIDynamicForwardProxy = "envoy.filters.network.sni_dynamic_forward_proxy"

	// dfpSNIUpstreamPort is the upstream port used by the SNI dynamic forward proxy.
	// The SNI only carries a hostname, so a fixed port is required. 443 is the
	// conventional port for SNI-based (TLS) forwarding, matching the upstream Envoy example.
	dfpSNIUpstreamPort = 443

	// dfpLoopbackSNIRegex matches loopback hostnames/addresses presented as the SNI.
	// It mirrors the loopback protection applied to the HTTP dynamic forward proxy to
	// prevent the SNI dynamic forward proxy from being abused to reach loopback addresses.
	dfpLoopbackSNIRegex = `(?i)^(localhost|localhost\.localdomain|ip6-localhost|ip6-loopback|127\.0\.0\.1|::1|\[::1\])\.?$`
)

// buildSNIDynamicForwardProxyFilter builds the sni_dynamic_forward_proxy network filter for a
// dynamic resolver TCPRoute. It resolves the upstream host from the SNI extracted by the
// tls_inspector listener filter and shares the same DNS cache as the dynamic forward proxy cluster.
func buildSNIDynamicForwardProxyFilter(irRoute *ir.TCPRoute) (*listenerv3.Filter, error) {
	// The DNS cache must match the one used by the dynamic forward proxy cluster so that both
	// share the same resolved addresses. The cluster is built with a nil IP family for TCP routes
	// (see processTCPListenerXdsTranslation), so the same is used here to keep the cache name in sync.
	dnsLookupFamily := computeDNSLookupFamily(nil, irRoute.DNS)
	cacheName := dfpCacheName(nil, irRoute.DNS)
	dnsCacheConfig := buildDFPDNSCacheConfig(cacheName, irRoute.DNS, dnsLookupFamily)

	cfg := &snidfpv3.FilterConfig{
		DnsCacheConfig: dnsCacheConfig,
		PortSpecifier: &snidfpv3.FilterConfig_PortValue{
			PortValue: dfpSNIUpstreamPort,
		},
	}

	return toNetworkFilter(networkSNIDynamicForwardProxy, cfg)
}

// buildDFPLoopbackNetworkRBAC builds a network RBAC filter that denies connections whose SNI
// resolves to a loopback hostname/address. It is applied ahead of the SNI dynamic forward proxy
// filter to mitigate SSRF against loopback addresses, mirroring the HTTP dynamic forward proxy.
func buildDFPLoopbackNetworkRBAC(statPrefix string) (*listenerv3.Filter, error) {
	rbac := &networkrbacv3.RBAC{
		StatPrefix: statPrefix,
		Rules: &rbacconfigv3.RBAC{
			Action: rbacconfigv3.RBAC_DENY,
			Policies: map[string]*rbacconfigv3.Policy{
				"deny-loopback-sni": {
					Permissions: []*rbacconfigv3.Permission{
						{
							Rule: &rbacconfigv3.Permission_RequestedServerName{
								RequestedServerName: &envoymatcherv3.StringMatcher{
									MatchPattern: &envoymatcherv3.StringMatcher_SafeRegex{
										SafeRegex: &envoymatcherv3.RegexMatcher{
											Regex: dfpLoopbackSNIRegex,
										},
									},
								},
							},
						},
					},
					Principals: []*rbacconfigv3.Principal{
						{
							Identifier: &rbacconfigv3.Principal_Any{Any: true},
						},
					},
				},
			},
		},
	}

	return toNetworkFilter(wellknown.RoleBasedAccessControl, rbac)
}
