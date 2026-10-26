// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"net/http"
	"testing"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/envoyproxy/gateway/internal/ir"
	xdsfilters "github.com/envoyproxy/gateway/internal/xds/filters"
)

func TestDetectMisdirectedRequests(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		listener *ir.HTTPListener
		want     bool
	}{
		"nil listener": {},
		"no overlapping TLS": {
			listener: &ir.HTTPListener{
				OverlappingTLSHandling: ir.OverlappingTLSHandlingMisdirectedRequest,
			},
		},
		"unset handling": {
			listener: &ir.HTTPListener{
				TLSOverlaps: true,
			},
		},
		"downgrade handling": {
			listener: &ir.HTTPListener{
				TLSOverlaps:            true,
				OverlappingTLSHandling: ir.OverlappingTLSHandlingDowngradeToHTTP1,
			},
		},
		"misdirected request handling": {
			listener: &ir.HTTPListener{
				TLSOverlaps:            true,
				OverlappingTLSHandling: ir.OverlappingTLSHandlingMisdirectedRequest,
				TLSOverlapsHostnames:   []string{"bar.example.com"},
			},
			want: true,
		},
		"misdirected request handling without recorded hostnames (e.g. SDS)": {
			listener: &ir.HTTPListener{
				TLSOverlaps:            true,
				OverlappingTLSHandling: ir.OverlappingTLSHandlingMisdirectedRequest,
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			require.Empty(t, cmp.Diff(tc.want, detectMisdirectedRequests(tc.listener)))
		})
	}
}

func TestDomainsMatchHostname(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		domains  []string
		hostname string
		want     bool
	}{
		"exact":                              {domains: []string{"foo.example.com"}, hostname: "foo.example.com", want: true},
		"exact mismatch":                     {domains: []string{"foo.example.com"}, hostname: "bar.example.com"},
		"any":                                {domains: []string{"*"}, hostname: "foo.example.com", want: true},
		"wildcard single label":              {domains: []string{"*.example.com"}, hostname: "foo.example.com", want: true},
		"wildcard multiple labels":           {domains: []string{"*.example.com"}, hostname: "foo.bar.example.com", want: true},
		"wildcard matches narrower wildcard": {domains: []string{"*.example.com"}, hostname: "*.foo.example.com", want: true},
		"wildcard does not match apex":       {domains: []string{"*.example.com"}, hostname: "example.com"},
		"wildcard does not match broader wildcard": {domains: []string{"*.foo.example.com"}, hostname: "*.example.com"},
		"exact does not match wildcard":            {domains: []string{"foo.example.com"}, hostname: "*.example.com"},
		"any of the domains":                       {domains: []string{"bar.com", "*.example.com"}, hostname: "foo.example.com", want: true},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			require.Empty(t, cmp.Diff(tc.want, domainsMatchHostname(tc.domains, tc.hostname)))
		})
	}
}

func TestBuildMisdirectedRequestRouteAuthority(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		hostname string
		want     *routev3.Route
	}{
		"exact hostname includes an optional port": {
			hostname: "foo.example.com",
			want: &routev3.Route{
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
					Headers: []*routev3.HeaderMatcher{{
						Name: AuthorityHeaderKey,
						HeaderMatchSpecifier: &routev3.HeaderMatcher_StringMatch{
							StringMatch: &matcherv3.StringMatcher{
								MatchPattern: &matcherv3.StringMatcher_SafeRegex{
									SafeRegex: &matcherv3.RegexMatcher{Regex: `^foo\.example\.com(:\d+)?$`},
								},
							},
						},
					}},
				},
				Action: &routev3.Route_DirectResponse{
					DirectResponse: &routev3.DirectResponseAction{Status: http.StatusMisdirectedRequest},
				},
			},
		},
		"wildcard hostname matches one or more labels": {
			hostname: "*.example.com",
			want: &routev3.Route{
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
					Headers: []*routev3.HeaderMatcher{{
						Name: AuthorityHeaderKey,
						HeaderMatchSpecifier: &routev3.HeaderMatcher_StringMatch{
							StringMatch: &matcherv3.StringMatcher{
								MatchPattern: &matcherv3.StringMatcher_SafeRegex{
									SafeRegex: &matcherv3.RegexMatcher{Regex: `^.+\.example\.com(:\d+)?$`},
								},
							},
						},
					}},
				},
				Action: &routev3.Route_DirectResponse{
					DirectResponse: &routev3.DirectResponseAction{Status: http.StatusMisdirectedRequest},
				},
			},
		},
		"all hostnames omit the authority match": {
			hostname: "*",
			want: &routev3.Route{
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
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			require.Empty(t, cmp.Diff(tc.want, buildMisdirectedRequestRoute(tc.hostname), protocmp.Transform()))
		})
	}
}
