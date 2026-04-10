// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"sort"
	"strings"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/filters"
)

var return421Route = &routev3.Route{
	Match: &routev3.RouteMatch{
		PathSpecifier: &routev3.RouteMatch_Prefix{
			Prefix: "/",
		},
		FilterState: []*matcherv3.FilterStateMatcher{
			{
				Key: filters.DownstreamProtocolKey,
				Matcher: &matcherv3.FilterStateMatcher_StringMatch{
					StringMatch: &matcherv3.StringMatcher{
						MatchPattern: &matcherv3.StringMatcher_Exact{
							Exact: "HTTP/2",
						},
					},
				},
			},
		},
	},
	Action: &routev3.Route_DirectResponse{
		DirectResponse: &routev3.DirectResponseAction{
			Status: 421,
		},
	},
}

func getReturn421RouteWithHost(hostname string) *routev3.Route {
	route := &routev3.Route{
		Match: &routev3.RouteMatch{
			PathSpecifier: &routev3.RouteMatch_Prefix{
				Prefix: "/",
			},
			FilterState: []*matcherv3.FilterStateMatcher{
				{
					Key: filters.DownstreamProtocolKey,
					Matcher: &matcherv3.FilterStateMatcher_StringMatch{
						StringMatch: &matcherv3.StringMatcher{
							MatchPattern: &matcherv3.StringMatcher_Exact{
								Exact: "HTTP/2",
							},
						},
					},
				},
			},
		},
		Action: &routev3.Route_DirectResponse{
			DirectResponse: &routev3.DirectResponseAction{
				Status: 421,
			},
		},
	}

	// Handle wildcard hostnames appropriately
	// The :authority header can include a port (e.g., example.com:443), so we use regex
	// to match both with and without port.
	switch {
	case hostname == "*":
		// Wildcard matches all hostnames, no specific :authority check needed
		// The virtual host domain matching will handle it
		return route
	case len(hostname) > 2 && hostname[:2] == "*.":
		// Wildcard prefix like *.example.com - match one or more labels + suffix with optional port
		// e.g., *.example.com matches foo.example.com, foo.bar.example.com and with optional :port.
		suffix := hostname[2:] // Remove the *. prefix, e.g., "example.com"
		// Regex: ^.+\.escaped-suffix(:\d+)?$
		// .+ ensures at least one non-empty label prefix before the suffix.
		escapedSuffix := strings.ReplaceAll(suffix, ".", `\.`)
		pattern := `^.+\.` + escapedSuffix + `(:\d+)?$`
		route.Match.Headers = []*routev3.HeaderMatcher{
			{
				Name: ":authority",
				HeaderMatchSpecifier: &routev3.HeaderMatcher_StringMatch{
					StringMatch: &matcherv3.StringMatcher{
						MatchPattern: &matcherv3.StringMatcher_SafeRegex{
							SafeRegex: &matcherv3.RegexMatcher{
								Regex: pattern,
							},
						},
					},
				},
			},
		}
	default:
		// Exact hostname - match with optional port
		// e.g., example.com matches example.com or example.com:443
		escapedHostname := strings.ReplaceAll(hostname, ".", `\.`)
		pattern := `^` + escapedHostname + `(:\d+)?$`
		route.Match.Headers = []*routev3.HeaderMatcher{
			{
				Name: ":authority",
				HeaderMatchSpecifier: &routev3.HeaderMatcher_StringMatch{
					StringMatch: &matcherv3.StringMatcher{
						MatchPattern: &matcherv3.StringMatcher_SafeRegex{
							SafeRegex: &matcherv3.RegexMatcher{
								Regex: pattern,
							},
						},
					},
				},
			},
		}
	}

	return route
}

// patchVirtualHostForOverlaps patches a single virtual host for TLS overlaps and returns
// the remaining unmatched overlap hostnames.
func (t *Translator) patchVirtualHostForOverlaps(vh *routev3.VirtualHost, tlsOverlapsHostnames []string) []string {
	// if vh.domains matched any of the overlaps hostnames, we add the special route with header :authority to return 421 when using h2.
	// Otherwise, envoy will return 404 instead of 200 when using http1.1(serverName: third-example.wildcard.org hostname: fourth-example.wildcard.org).
	for i, overlapsHostname := range tlsOverlapsHostnames {
		if !domainsMatched(vh.Domains, overlapsHostname) {
			continue
		}
		// append return 421 route to the first of vh.Routes
		r := getReturn421RouteWithHost(overlapsHostname)
		vh.Routes = append([]*routev3.Route{r}, vh.Routes...)
		// Remove this hostname from the list by swapping with the last element and truncating
		tlsOverlapsHostnames[i] = tlsOverlapsHostnames[len(tlsOverlapsHostnames)-1]
		return tlsOverlapsHostnames[:len(tlsOverlapsHostnames)-1]
	}
	return tlsOverlapsHostnames
}

// addCatchAllForRemainingOverlaps adds a catch-all virtual host for any remaining TLS overlap hostnames
// that weren't matched by existing virtual hosts.
func (t *Translator) addCatchAllForRemainingOverlaps(xdsRouteCfg *routev3.RouteConfiguration, httpListener *ir.HTTPListener, remainingHostnames []string) {
	if len(remainingHostnames) == 0 {
		return
	}
	// Sort for stable XDS output. Envoy uses specificity-based domain matching
	// (exact > suffix wildcard > prefix wildcard > "*"), so order doesn't affect routing.
	sort.Strings(remainingHostnames)
	xdsRouteCfg.VirtualHosts = append(xdsRouteCfg.VirtualHosts, &routev3.VirtualHost{
		Name:    virtualHostName(httpListener, "catch_all_tls_overlapping", t.xdsNameSchemeV2()),
		Domains: remainingHostnames,
		Routes:  []*routev3.Route{return421Route},
	})
}

func domainsMatched(vhDomains []string, overlapsHostname string) bool {
	for _, domain := range vhDomains {
		if domainMatchHostname(domain, overlapsHostname) {
			return true
		}
	}
	return false
}

// domainMatchHostname checks if the hostname is matched the virtual host domain,
// it returns true if the hostname is matched by any of the overlaps hostnames, otherwise returns false.
// Per Gateway API spec, wildcards match only a single DNS label:
// - *.example.com matches both test.example.com, and foo.test.example.com, but not example.com
// - * matches any hostname
func domainMatchHostname(vhDomain, overlapsHostname string) bool {
	if vhDomain == "*" {
		return true
	}
	if len(vhDomain) > 2 && vhDomain[:2] == "*." {
		domainSuffix := vhDomain[1:] // e.g., ".example.com"
		// Check: hostname must have the suffix and at least one label prefix
		if strings.HasSuffix(overlapsHostname, domainSuffix) {
			prefix := overlapsHostname[:len(overlapsHostname)-len(domainSuffix)]
			// Wildcard matches any number of labels - prefix must be non-empty
			return prefix != ""
		}
		return false
	}
	if vhDomain == overlapsHostname {
		return true
	}
	return false
}
