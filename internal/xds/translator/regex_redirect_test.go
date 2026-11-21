// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"regexp"
	"strings"
	"testing"

	setfilterstatev3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/set_filter_state/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/stretchr/testify/require"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
)

func TestRegexRedirectMatch(t *testing.T) {
	for _, pathMatch := range []*ir.StringMatch{
		nil,
		{Prefix: new("/blogs")},
		{Exact: new("/blogs/123")},
		{SafeRegex: new(`^/blogs/.*$`)},
	} {
		for _, pattern := range []string{
			`^/blogs/([0-9]+)$`,
			`blogs`,
			`(?i)BLOGS`,
			`^/blogs/[0-9]+$|^/articles/[0-9]+$`,
			`123$\b`,
			`\?next=123`,
		} {
			t.Run(pattern, func(t *testing.T) {
				route := &ir.HTTPRoute{
					PathMatch:     pathMatch,
					HeaderMatches: []*ir.StringMatch{{Name: "x-client", Exact: new("test")}},
					QueryParamMatches: []*ir.StringMatch{{
						Name: "version", Exact: new("v1"),
					}},
					Redirect: &ir.Redirect{Path: &ir.ExtendedHTTPPathModifier{
						RegexMatchReplace: &ir.RegexMatchReplace{Pattern: pattern, Substitution: "/post"},
					}},
				}
				want := buildXdsRouteMatch(false, pathMatch, route.HeaderMatches, route.QueryParamMatches, nil)
				got, err := buildXdsRoute(route, &ir.HTTPListener{}, nil, nil, false)
				require.NoError(t, err)
				require.NoError(t, got.ValidateAll())
				require.Len(t, got.Match.FilterState, 1)
				guard := got.Match.FilterState[0]
				require.Equal(t, regexRedirectPathKey, guard.Key)
				matcher, err := regexp.Compile(`\A(?:` + guard.GetStringMatch().GetSafeRegex().Regex + `)\z`)
				require.NoError(t, err)
				rewrite, err := regexp.Compile(pattern)
				require.NoError(t, err)
				for _, request := range []string{
					"/blogs/123", "/blogs/abc", "/blogs/123/", "/BLOGS/123",
					"/blogs/123?version=v1", "/blogs/abc?next=/blogs/123",
					"/articles/456", "/other?next=123",
				} {
					path, _, _ := strings.Cut(request, "?")
					require.Equal(t, rewrite.MatchString(path), matcher.MatchString(path), request)
				}
				got.Match.FilterState = nil
				require.Equal(t, want, got.Match)
				require.Len(t, route.HeaderMatches, 1)
			})
		}
	}
}

func TestRegexRedirectPathFilter(t *testing.T) {
	for _, tc := range []struct {
		name  string
		route *ir.HTTPRoute
		want  bool
	}{
		{name: "forward", route: &ir.HTTPRoute{}},
		{name: "native redirect", route: &ir.HTTPRoute{Redirect: &ir.Redirect{}}},
		{
			name: "regex redirect",
			route: &ir.HTTPRoute{Redirect: &ir.Redirect{Path: &ir.ExtendedHTTPPathModifier{
				RegexMatchReplace: &ir.RegexMatchReplace{Pattern: "^/blogs/[0-9]+$", Substitution: "/post"},
			}}},
			want: true,
		},
		{
			name: "direct response takes precedence",
			route: &ir.HTTPRoute{
				DirectResponse: &ir.CustomResponse{StatusCode: new(uint32(500))},
				Redirect: &ir.Redirect{Path: &ir.ExtendedHTTPPathModifier{
					RegexMatchReplace: &ir.RegexMatchReplace{Pattern: "^/blogs/[0-9]+$", Substitution: "/post"},
				}},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			listener := &ir.HTTPListener{Routes: []*ir.HTTPRoute{tc.route}}
			mgr := &hcmv3.HttpConnectionManager{HttpFilters: []*hcmv3.HttpFilter{
				httpFilterForTest(egv1a1.EnvoyFilterCORS),
			}}
			translator := &Translator{FilterOrder: []egv1a1.FilterPosition{{
				Name: egv1a1.EnvoyFilterRouter, Before: new(egv1a1.EnvoyFilterCORS),
			}}}
			for range 2 {
				require.NoError(t, translator.patchHCMWithFilters(mgr, listener, nil))
				require.Equal(t, tc.want, hcmContainsFilter(mgr, regexRedirectPathFilter))
				if !tc.want {
					continue
				}
				require.Equal(t, regexRedirectPathFilter, mgr.HttpFilters[0].Name)
				require.Len(t, mgr.HttpFilters, 3)
				config := &setfilterstatev3.Config{}
				require.NoError(t, mgr.HttpFilters[0].GetTypedConfig().UnmarshalTo(config))
				require.True(t, config.ClearRouteCache)
				require.Len(t, config.OnRequestHeaders, 1)
				value := config.OnRequestHeaders[0]
				require.Equal(t, regexRedirectPathKey, value.GetObjectKey())
				require.Equal(t, "envoy.string", value.FactoryKey)
				require.Equal(t, "%PATH(NQ:PATH)%", value.GetFormatString().GetTextFormatSource().GetInlineString())
			}
			if tc.route.Redirect != nil {
				got, err := buildXdsRoute(tc.route, listener, nil, nil, false)
				require.NoError(t, err)
				if !tc.want {
					require.Empty(t, got.Match.FilterState)
				}
			}
		})
	}
}
