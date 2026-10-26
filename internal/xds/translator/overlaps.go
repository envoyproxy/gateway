// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"net/http"
	"regexp"
	"strings"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"

	"github.com/envoyproxy/gateway/internal/ir"
	xdsfilters "github.com/envoyproxy/gateway/internal/xds/filters"
)

const (
	// allHostnames is the listener hostname recorded when the Gateway listener has no hostname.
	// A 421 route for it does not match :authority; the virtual host domains select the request.
	allHostnames = "*"
	// authorityPortPattern matches an optional ":443" suffix on :authority.
	authorityPortPattern = `(:\d+)?`
)

// detectMisdirectedRequests reports whether HTTP/2 is kept on a listener with overlapping TLS config
// and misdirected requests are answered with 421.
// Overlaps without recorded hostnames (e.g. opaque SDS certificates) keep the ALPN downgrade since no 421 routes can be generated for them.
func detectMisdirectedRequests(httpListener *ir.HTTPListener) bool {
	return httpListener != nil && httpListener.TLSOverlaps &&
		len(httpListener.TLSOverlapsHostnames) > 0 &&
		httpListener.OverlappingTLSHandling == ir.OverlappingTLSHandlingMisdirectedRequest
}

// addMisdirectedRequestRoutes returns 421 Misdirected Request for HTTP/2 requests whose authority matches the
// hostname of another listener with overlapping TLS config, since such requests were coalesced onto a connection
// established for this listener (see https://gateway-api.sigs.k8s.io/geps/gep-3567/).
//
// A virtual host whose domains already match an overlapping hostname gets a 421 route prepended, so HTTP/1.1
// requests keep falling through to its routes. A separate virtual host would win the domain match over a
// wildcard domain and turn those HTTP/1.1 requests into 404 Not Found. The overlapping hostnames that no virtual host
// matches are handled by a catch-all virtual host.
func (t *Translator) addMisdirectedRequestRoutes(vHosts []*routev3.VirtualHost, httpListener *ir.HTTPListener) []*routev3.VirtualHost {
	var unmatched []string
	for _, hostname := range httpListener.TLSOverlapsHostnames {
		var matched bool
		for _, vHost := range vHosts {
			if domainsMatchHostname(vHost.Domains, hostname) {
				vHost.Routes = append([]*routev3.Route{buildMisdirectedRequestRoute(hostname)}, vHost.Routes...)
				matched = true
			}
		}
		if !matched {
			unmatched = append(unmatched, hostname)
		}
	}

	if len(unmatched) == 0 {
		return vHosts
	}
	return append(vHosts, &routev3.VirtualHost{
		Name:    virtualHostName(httpListener, "catch_all_tls_overlapping", t.xdsNameSchemeV2()),
		Domains: unmatched,
		// The domains of this virtual host already select the misdirected requests.
		Routes: []*routev3.Route{buildMisdirectedRequestRoute(allHostnames)},
	})
}

// buildMisdirectedRequestRoute builds a route returning 421 for HTTP/2 requests whose authority matches the hostname.
func buildMisdirectedRequestRoute(hostname string) *routev3.Route {
	route := &routev3.Route{
		Match: &routev3.RouteMatch{
			PathSpecifier: &routev3.RouteMatch_Prefix{Prefix: "/"},
			FilterState: []*matcherv3.FilterStateMatcher{{
				Key: xdsfilters.DownstreamProtocolKey,
				Matcher: &matcherv3.FilterStateMatcher_StringMatch{
					StringMatch: &matcherv3.StringMatcher{
						MatchPattern: &matcherv3.StringMatcher_Exact{Exact: "HTTP/2"},
					},
				},
			}},
		},
		Action: &routev3.Route_DirectResponse{
			DirectResponse: &routev3.DirectResponseAction{Status: http.StatusMisdirectedRequest},
		},
	}
	if hostname == allHostnames {
		return route
	}

	// :authority is "foo.example.com" or "foo.example.com:443". QuoteMeta keeps "." literal.
	//   foo.example.com -> ^foo\.example\.com(:\d+)?$
	//   *.example.com   -> ^.+\.example\.com(:\d+)?$
	// The wildcard form requires at least one label before the suffix, so "example.com" does not match.
	pattern := "^" + regexp.QuoteMeta(hostname) + authorityPortPattern + "$"
	if suffix, ok := strings.CutPrefix(hostname, "*."); ok {
		pattern = `^.+\.` + regexp.QuoteMeta(suffix) + authorityPortPattern + "$"
	}
	route.Match.Headers = []*routev3.HeaderMatcher{{
		Name: AuthorityHeaderKey,
		HeaderMatchSpecifier: &routev3.HeaderMatcher_StringMatch{
			StringMatch: &matcherv3.StringMatcher{
				MatchPattern: &matcherv3.StringMatcher_SafeRegex{
					SafeRegex: &matcherv3.RegexMatcher{Regex: pattern},
				},
			},
		},
	}}
	return route
}

// domainsMatchHostname reports whether any of the virtual host domains matches the hostname, following the Envoy
// virtual host domain matching: "*" matches any hostname, and "*.example.com" matches any hostname ending with
// ".example.com", including a wildcard hostname such as "*.foo.example.com".
func domainsMatchHostname(domains []string, hostname string) bool {
	for _, domain := range domains {
		if domain == hostname {
			return true
		}
		if suffix, ok := strings.CutPrefix(domain, allHostnames); ok && len(hostname) > len(suffix) && strings.HasSuffix(hostname, suffix) {
			return true
		}
	}
	return false
}
