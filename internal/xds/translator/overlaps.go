// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"sort"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"k8s.io/apimachinery/pkg/util/sets"

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
	return &routev3.Route{
		Match: &routev3.RouteMatch{
			PathSpecifier: &routev3.RouteMatch_Prefix{
				Prefix: "/",
			},
			Headers: []*routev3.HeaderMatcher{
				{
					Name: ":authority",
					HeaderMatchSpecifier: &routev3.HeaderMatcher_StringMatch{
						StringMatch: &matcherv3.StringMatcher{
							MatchPattern: &matcherv3.StringMatcher_Exact{
								Exact: hostname,
							},
						},
					},
				},
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
}

func (t *Translator) mayPatchVirtualHostsForOverlaps(xdsRouteCfg *routev3.RouteConfiguration, httpListener *ir.HTTPListener) {
	if !httpListener.TLSOverlaps || len(httpListener.TLSOverlapsHostnames) == 0 {
		return
	}
	// add a route that matches all hosts and returns 421 to handle TLS overlapping scenario
	domains := sets.NewString(httpListener.TLSOverlapsHostnames...)

	for _, vh := range xdsRouteCfg.VirtualHosts {
		// if vh.domains matched any of the overlaps hostnames, we add the special route with header :authority to return 421 when using h2.
		// Otherwise, envoy will return 404 instead of 200 when using http1.1(serverName: third-example.wildcard.org hostname: fourth-example.wildcard.org).
		for _, overlapsHostname := range httpListener.TLSOverlapsHostnames {
			if !domainsMatched(vh.Domains, overlapsHostname) {
				continue
			}
			// append return 421 route to the first of vh.Routes
			r := getReturn421RouteWithHost(overlapsHostname)
			vh.Routes = append([]*routev3.Route{r}, vh.Routes...)
			domains.Delete(overlapsHostname)
			break
		}
	}

	if len(domains) == 0 {
		return
	}
	out := domains.UnsortedList()
	// sort the domains and make sure * is always at the end of the list since Envoy will match the domains in order and * should be the last one to be matched
	sort.Slice(out, func(i, j int) bool {
		if out[i] == "*" {
			return false
		}
		if out[j] == "*" {
			return true
		}
		return out[i] < out[j]
	})
	xdsRouteCfg.VirtualHosts = append(xdsRouteCfg.VirtualHosts, &routev3.VirtualHost{
		Name:    virtualHostName(httpListener, "catch_all_tls_overlapping", t.xdsNameSchemeV2()),
		Domains: out,
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
// e.g. *.wildcard.com will match www.wildcard.com, but not match www.sub.wildcard.com; * will match any hostname.
func domainMatchHostname(vhDomain, overlapsHostname string) bool {
	if vhDomain == "*" {
		return true
	}
	if len(vhDomain) > 2 && vhDomain[:2] == "*." {
		domainSuffix := vhDomain[1:] // remove the leading '*'
		if len(overlapsHostname) > len(domainSuffix) && overlapsHostname[len(overlapsHostname)-len(domainSuffix):] == domainSuffix {
			return true
		}
	}
	if vhDomain == overlapsHostname {
		return true
	}
	return false
}
