// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"regexp"
	"slices"
	"strings"

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
	"google.golang.org/protobuf/types/known/wrapperspb"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/proto"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

const (
	quicClientValidationFilterName       = "envoy.filters.http.rbac.quic_client_validation"
	quicClientValidationRouterFilterName = quicClientValidationFilterName + "_before_router"
	// DNS names can contain 253 characters. Wildcard matching and Unicode
	// case folding can exceed Envoy's default RE2 limit of 100 instructions.
	// This bounded limit applies only to our generated hostname patterns.
	quicClientValidationRegexMaxProgramSize = 2048
)

// patchQUICWildcardClientValidation prevents an absent client-validated QUIC chain
// from falling through to an unvalidated wildcard chain on the same socket.
func patchQUICWildcardClientValidation(tCtx *types.ResourceVersionTable, httpListeners []*ir.HTTPListener, tcpListeners []*ir.TCPListener) error {
	protected := make(map[listenerKey]sets.Set[string])
	addProtected := func(details ir.CoreListenerDetails, hostnames []string, tls *ir.TLSConfig) {
		if tls == nil || (!tls.ClientValidationEnabled && tls.CACertificate == nil && !tls.RequireClientCertificate) {
			return
		}
		key := listenerKey{Address: details.Address, Port: details.Port}
		if protected[key] == nil {
			protected[key] = sets.New[string]()
		}
		if len(hostnames) == 0 {
			protected[key].Insert("*")
		}
		for _, hostname := range hostnames {
			protected[key].Insert(strings.ToLower(strings.TrimSuffix(hostname, ".")))
		}
	}
	for _, listener := range httpListeners {
		addProtected(listener.CoreListenerDetails, listener.Hostnames, listener.TLS)
	}
	for _, listener := range tcpListeners {
		addProtected(listener.CoreListenerDetails, listener.Hostnames, listener.TLS)
		for _, route := range listener.Routes {
			if route.TLS == nil || route.TLS.Terminate == nil {
				continue
			}
			hostnames := listener.Hostnames
			if inspector := route.TLS.TLSInspectorConfig; inspector != nil && len(inspector.SNIs) > 0 {
				hostnames = inspector.SNIs
			}
			addProtected(listener.CoreListenerDetails, hostnames, route.TLS.Terminate)
		}
	}

	var errs error
	for _, resource := range tCtx.XdsResources[resourcev3.ListenerType] {
		listener := resource.(*listenerv3.Listener)
		socket := listener.GetAddress().GetSocketAddress()
		if socket == nil || socket.Protocol != corev3.SocketAddress_UDP {
			continue
		}
		hostnames := sets.List(protected[listenerKey{Address: socket.Address, Port: socket.GetPortValue()}])
		if len(hostnames) == 0 {
			continue
		}
		// Use emitted chains for exemptions so unresolved certificates cannot
		// allow traffic through another wildcard chain.
		publicHostnames := sets.New[string]()
		for _, chain := range append(slices.Clone(listener.FilterChains), listener.DefaultFilterChain) {
			if chain == nil {
				continue
			}
			mgr, err := findHCMinFilterChain(chain)
			if err != nil || mgr.CodecType != hcmv3.HttpConnectionManager_HTTP3 {
				continue
			}
			serverNames := chain.GetFilterChainMatch().GetServerNames()
			if len(serverNames) == 0 {
				publicHostnames.Insert("*")
			}
			for _, hostname := range serverNames {
				publicHostnames.Insert(strings.ToLower(strings.TrimSuffix(hostname, ".")))
			}
		}
		// A public TLS termination listener can override a broader protected
		// hostname on TCP without having its own QUIC chain. Passthrough chains
		// cannot provide this exemption: their backend's client auth is unknown.
		if tcp := findXdsListenerByHostPort(tCtx, socket.Address, socket.GetPortValue(), corev3.SocketAddress_TCP); tcp != nil {
			for _, chain := range append(slices.Clone(tcp.FilterChains), tcp.DefaultFilterChain) {
				if chain == nil || chain.GetTransportSocket().GetName() != wellknown.TransportSocketTls {
					continue
				}
				tls := &tlsv3.DownstreamTlsContext{}
				if err := chain.TransportSocket.GetTypedConfig().UnmarshalTo(tls); err != nil {
					continue
				}
				if tls.CommonTlsContext == nil || tls.CommonTlsContext.ValidationContextType != nil || tls.GetRequireClientCertificate().GetValue() {
					continue
				}
				for _, hostname := range chain.GetFilterChainMatch().GetServerNames() {
					publicHostnames.Insert(strings.ToLower(strings.TrimSuffix(hostname, ".")))
				}
			}
		}
		exemptions := sets.List(publicHostnames)
		chains := listener.FilterChains[:0]
		for _, chain := range listener.FilterChains {
			if err := patchQUICWildcardFilterChain(chain, hostnames, exemptions); err != nil {
				// Omit a chain that could not be protected instead of allowing fallback.
				errs = errors.Join(errs, err)
				continue
			}
			chains = append(chains, chain)
		}
		listener.FilterChains = chains
		if listener.DefaultFilterChain != nil {
			if err := patchQUICWildcardFilterChain(listener.DefaultFilterChain, hostnames, exemptions); err != nil {
				errs = errors.Join(errs, err)
				listener.DefaultFilterChain = nil
			}
		}
	}
	return errs
}

func patchQUICWildcardFilterChain(chain *listenerv3.FilterChain, protected, public []string) error {
	var hostnames []string
	serverNames := chain.GetFilterChainMatch().GetServerNames()
	for _, hostname := range protected {
		if len(serverNames) == 0 {
			hostnames = append(hostnames, hostname)
			continue
		}
		for _, serverName := range serverNames {
			if strings.HasPrefix(serverName, "*.") && quicHostnamesOverlap(serverName, hostname) {
				hostnames = append(hostnames, hostname)
				break
			}
		}
	}
	if len(hostnames) == 0 {
		return nil
	}

	for _, filter := range chain.Filters {
		if filter.Name != wellknown.HTTPConnectionManager {
			continue
		}
		mgr := &hcmv3.HttpConnectionManager{}
		if err := filter.GetTypedConfig().UnmarshalTo(mgr); err != nil {
			return err
		}
		if mgr.CodecType != hcmv3.HttpConnectionManager_HTTP3 {
			return nil
		}
		guard, err := buildQUICClientValidationFilter(hostnames, public)
		if err != nil {
			return err
		}
		// Install after normal/custom filter sorting, before any terminal filters.
		filters := []*hcmv3.HttpFilter{guard}
		for _, httpFilter := range mgr.HttpFilters {
			if httpFilter.Name == quicClientValidationFilterName || httpFilter.Name == quicClientValidationRouterFilterName {
				continue
			}
			if httpFilter.Name == wellknown.Router {
				// Recheck after filters that can change :authority or clear the route cache.
				filters = append(filters, &hcmv3.HttpFilter{
					Name:       quicClientValidationRouterFilterName,
					ConfigType: &hcmv3.HttpFilter_TypedConfig{TypedConfig: guard.GetTypedConfig()},
				})
			}
			filters = append(filters, httpFilter)
		}
		mgr.HttpFilters = filters
		config, err := proto.ToAnyWithValidation(mgr)
		if err != nil {
			return err
		}
		filter.ConfigType = &listenerv3.Filter_TypedConfig{TypedConfig: config}
	}
	return nil
}

func quicHostnamesOverlap(wildcard, hostname string) bool {
	wildcard = strings.ToLower(strings.TrimSuffix(wildcard, "."))
	if hostname == "*" {
		return true
	}
	suffix := strings.TrimPrefix(wildcard, "*")
	if strings.HasPrefix(hostname, "*.") {
		otherSuffix := strings.TrimPrefix(hostname, "*")
		return strings.HasSuffix(suffix, otherSuffix) || strings.HasSuffix(otherSuffix, suffix)
	}
	return len(hostname) > len(suffix) && strings.HasSuffix(hostname, suffix)
}

// quicHostnameMoreSpecific follows SNI precedence: exact names, then the
// longest wildcard suffix, then the catch-all listener.
func quicHostnameMoreSpecific(hostname, broader string) bool {
	if broader == "*" {
		return hostname != "*"
	}
	if !strings.HasPrefix(broader, "*.") || hostname == "*" {
		return false
	}
	if strings.HasPrefix(hostname, "*.") {
		return hostname != broader && strings.HasSuffix(strings.TrimPrefix(hostname, "*"), strings.TrimPrefix(broader, "*"))
	}
	return quicHostnamesOverlap(broader, hostname)
}

func buildQUICClientValidationFilter(hostnames, public []string) (*hcmv3.HttpFilter, error) {
	permissions := make([]*rbacconfigv3.Permission, 0, 2*len(hostnames))
	for _, hostname := range hostnames {
		for _, authority := range []bool{false, true} {
			permission := quicHostnamePermission(hostname, authority)
			var exemptions []*rbacconfigv3.Permission
			for _, publicHostname := range public {
				if quicHostnameMoreSpecific(publicHostname, hostname) {
					exemptions = append(exemptions, quicHostnamePermission(publicHostname, authority))
				}
			}
			if len(exemptions) > 0 {
				// Exemptions apply separately to SNI and authority. A public
				// authority must never exempt a protected SNI, or vice versa.
				permission = &rbacconfigv3.Permission{Rule: &rbacconfigv3.Permission_AndRules{
					AndRules: &rbacconfigv3.Permission_Set{Rules: []*rbacconfigv3.Permission{
						permission,
						{Rule: &rbacconfigv3.Permission_NotRule{NotRule: &rbacconfigv3.Permission{Rule: &rbacconfigv3.Permission_OrRules{
							OrRules: &rbacconfigv3.Permission_Set{Rules: exemptions},
						}}}},
					}},
				}}
			}
			permissions = append(permissions, permission)
		}
	}
	config, err := proto.ToAnyWithValidation(&rbacv3.RBAC{Rules: &rbacconfigv3.RBAC{
		Action: rbacconfigv3.RBAC_DENY,
		Policies: map[string]*rbacconfigv3.Policy{"client-validation-required": {
			Permissions: permissions,
			Principals:  []*rbacconfigv3.Principal{{Identifier: &rbacconfigv3.Principal_Any{Any: true}}},
		}},
	}})
	if err != nil {
		return nil, err
	}
	return &hcmv3.HttpFilter{
		Name:       quicClientValidationFilterName,
		ConfigType: &hcmv3.HttpFilter_TypedConfig{TypedConfig: config},
	}, nil
}

func quicHostnamePermission(hostname string, authority bool) *rbacconfigv3.Permission {
	pattern := regexp.QuoteMeta(hostname)
	if hostname == "*" {
		pattern = ".*"
	} else if strings.HasPrefix(hostname, "*.") {
		pattern = "[^:]+" + regexp.QuoteMeta(strings.TrimPrefix(hostname, "*"))
	}
	if authority {
		// Envoy strips the last colon unless a closing bracket follows it.
		// Legacy validation accepts empty/nonnumeric ports and colons inside
		// the routed hostname. Wildcards must match that hostname, including
		// bracketed prefixes, without exempting a different routed hostname.
		headerPermission := func(authorityPattern string) *rbacconfigv3.Permission {
			return &rbacconfigv3.Permission{Rule: &rbacconfigv3.Permission_Header{
				Header: &routev3.HeaderMatcher{Name: AuthorityHeaderKey, HeaderMatchSpecifier: &routev3.HeaderMatcher_SafeRegexMatch{
					SafeRegexMatch: quicHostnameRegex("(?i)^" + authorityPattern + "$"),
				}},
			}}
		}
		if strings.HasPrefix(hostname, "*.") {
			suffix := regexp.QuoteMeta(strings.TrimPrefix(hostname, "*")) + `\.?`
			// Without a stripped port, the hostname either has no colon or
			// has a closing bracket after its last colon. With a stripped port,
			// earlier colons and brackets may remain in the hostname prefix.
			// Separate regexes keep long DNS names within the RE2 program limit.
			return &rbacconfigv3.Permission{Rule: &rbacconfigv3.Permission_OrRules{
				OrRules: &rbacconfigv3.Permission_Set{Rules: []*rbacconfigv3.Permission{
					headerPermission(`(?:[^:]+|.*:[^:]*\][^:]*)` + suffix),
					headerPermission(".+" + suffix + `:[^:\]]*`),
				}},
			}}
		}
		return headerPermission(pattern + `\.?(?::[^:\]]*)?`)
	}
	return &rbacconfigv3.Permission{Rule: &rbacconfigv3.Permission_RequestedServerName{
		RequestedServerName: &matcherv3.StringMatcher{MatchPattern: &matcherv3.StringMatcher_SafeRegex{
			SafeRegex: quicHostnameRegex("(?i)^" + pattern + "\\.?$"),
		}},
	}}
}

func quicHostnameRegex(pattern string) *matcherv3.RegexMatcher {
	return &matcherv3.RegexMatcher{
		Regex: pattern,
		EngineType: &matcherv3.RegexMatcher_GoogleRe2{GoogleRe2: &matcherv3.RegexMatcher_GoogleRE2{
			MaxProgramSize: wrapperspb.UInt32(quicClientValidationRegexMaxProgramSize),
		}},
	}
}
